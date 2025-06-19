package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/config"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/interface/mcp"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/logger"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/state"
	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	var (
		envFile       = flag.String("env", ".env", "Environment file to load")
		transport     = flag.String("transport", "stdio", "MCP transport: stdio, http, or sse")
		port          = flag.Int("port", 8081, "Port for HTTP/SSE transport")
		logLevel      = flag.String("log-level", "info", "Log level: debug, info, warn, error")
		mockMode      = flag.Bool("mock", false, "Run in mock mode without New Relic connection")
	)
	flag.Parse()

	// Set up logging
	logger.SetLevel(*logLevel)

	// Load environment variables
	if err := godotenv.Load(*envFile); err != nil {
		logger.Warn("Failed to load env file %s: %v", *envFile, err)
	}

	// Override with command line flags
	if *transport != "" {
		os.Setenv("MCP_TRANSPORT", *transport)
	}
	if *mockMode {
		os.Setenv("MOCK_MODE", "true")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}


	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize components
	logger.Info("Initializing MCP server...")

	// Create MCP server
	mcpConfig := mcp.ServerConfig{
		TransportType: mcp.TransportType(cfg.Server.MCPTransport),
		HTTPHost:      cfg.Server.Host,
		HTTPPort:      *port,
	}
	
	server := mcp.NewServer(mcpConfig)

	// Initialize state manager
	logger.Info("Initializing state management...")
	
	// Determine store type from environment
	storeType := state.StoreTypeMemory
	if cfg.Redis.URL != "" && cfg.Redis.URL != "redis://localhost:6379" {
		storeType = state.StoreTypeRedis
	}
	
	stateConfig := state.FactoryConfig{
		StoreType: storeType,
		ManagerConfig: state.ManagerConfig{
			SessionTTL:      30 * time.Minute,  // Default session TTL
			CacheTTL:        5 * time.Minute,   // Default cache TTL
			MaxSessions:     10000,
			MaxCacheEntries: 100000,
			MaxCacheMemory:  1 << 30,          // 1GB cache memory limit
		},
	}
	
	// Add Redis config if using Redis
	if storeType == state.StoreTypeRedis {
		stateConfig.RedisConfig = &state.RedisConfig{
			URL:        cfg.Redis.URL,
			MaxRetries: 3,
			PoolSize:   10,
			KeyPrefix:  "mcp:",
			DefaultTTL: 30 * time.Minute,
		}
	}
	
	stateManager, err := state.NewStateManager(stateConfig)
	if err != nil {
		logger.Error("Failed to initialize state manager: %v", err)
		os.Exit(1)
	}
	server.SetStateManager(stateManager)

	// Initialize discovery engine
	logger.Info("Initializing discovery engine...")
	discoveryEngine, err := discovery.InitializeEngine(ctx)
	if err != nil {
		logger.Error("Failed to initialize discovery engine: %v", err)
		os.Exit(1)
	}
	server.SetDiscovery(discoveryEngine)

	// Initialize New Relic client (if not in mock mode)
	if !*mockMode && !cfg.Development.MockMode {
		logger.Info("Initializing New Relic client...")
		nrClient, err := newrelic.NewClient(newrelic.Config{
			APIKey:    cfg.NewRelic.APIKey,
			AccountID: cfg.NewRelic.AccountID,
			Region:    cfg.NewRelic.Region,
		})
		if err != nil {
			logger.Error("Failed to initialize New Relic client: %v", err)
			os.Exit(1)
		}
		server.SetNewRelicClient(nrClient)

		// Test connection
		if _, err := nrClient.GetAccountInfo(ctx); err != nil {
			logger.Warn("Failed to connect to New Relic: %v", err)
			logger.Warn("Continuing in degraded mode...")
		} else {
			logger.Info("Successfully connected to New Relic account %s", cfg.NewRelic.AccountID)
		}
	} else {
		logger.Info("Running in MOCK MODE - no New Relic connection")
		// Set mock mode flag in development config to ensure consistency
		cfg.Development.MockMode = true
	}

	// Start the server
	logger.Info("Starting MCP server with %s transport...", cfg.Server.MCPTransport)
	if err := server.Start(ctx); err != nil {
		logger.Error("Failed to start server: %v", err)
		os.Exit(1)
	}

	// Log startup information
	switch cfg.Server.MCPTransport {
	case "stdio":
		logger.Info("MCP server running on stdio")
		logger.Info("Ready to receive MCP protocol messages...")
	case "http":
		logger.Info("MCP server running on http://%s:%d", cfg.Server.Host, *port)
	case "sse":
		logger.Info("MCP server running on http://%s:%d (SSE)", cfg.Server.Host, *port)
	}

	// Wait for shutdown signal
	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	logger.Info("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown: %v", err)
	}

	// Stop discovery engine
	if err := discoveryEngine.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping discovery engine: %v", err)
	}

	logger.Info("Server stopped")
}


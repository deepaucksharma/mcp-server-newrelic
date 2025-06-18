package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/config"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/interface/mcp"
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

	// Load environment variables
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("Warning: failed to load env file %s: %v", *envFile, err)
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
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level
	setupLogging(*logLevel)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize components
	log.Println("Initializing MCP server...")

	// Create MCP server
	mcpConfig := mcp.ServerConfig{
		TransportType: mcp.TransportType(cfg.Server.MCPTransport),
		HTTPHost:      cfg.Server.Host,
		HTTPPort:      *port,
	}
	
	server := mcp.NewServer(mcpConfig)

	// Initialize state manager
	log.Println("Initializing state management...")
	
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
		log.Fatalf("Failed to initialize state manager: %v", err)
	}
	server.SetStateManager(stateManager)

	// Initialize discovery engine
	log.Println("Initializing discovery engine...")
	discoveryEngine, err := discovery.InitializeEngine(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize discovery engine: %v", err)
	}
	server.SetDiscovery(discoveryEngine)

	// Initialize New Relic client (if not in mock mode)
	if !*mockMode && os.Getenv("MOCK_MODE") != "true" {
		log.Println("Initializing New Relic client...")
		nrClient, err := newrelic.NewClient(newrelic.Config{
			APIKey:    cfg.NewRelic.APIKey,
			AccountID: cfg.NewRelic.AccountID,
			Region:    cfg.NewRelic.Region,
		})
		if err != nil {
			log.Fatalf("Failed to initialize New Relic client: %v", err)
		}
		server.SetNewRelicClient(nrClient)

		// Test connection
		if _, err := nrClient.GetAccountInfo(ctx); err != nil {
			log.Printf("Warning: Failed to connect to New Relic: %v", err)
			log.Println("Continuing in degraded mode...")
		} else {
			log.Printf("Successfully connected to New Relic account %s", cfg.NewRelic.AccountID)
		}
	} else {
		log.Println("Running in MOCK MODE - no New Relic connection")
	}

	// Start the server
	log.Printf("Starting MCP server with %s transport...", cfg.Server.MCPTransport)
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Log startup information
	switch cfg.Server.MCPTransport {
	case "stdio":
		log.Println("MCP server running on stdio")
		log.Println("Ready to receive MCP protocol messages...")
	case "http":
		log.Printf("MCP server running on http://%s:%d", cfg.Server.Host, *port)
	case "sse":
		log.Printf("MCP server running on http://%s:%d (SSE)", cfg.Server.Host, *port)
	}

	// Wait for shutdown signal
	go func() {
		<-sigChan
		log.Println("Shutdown signal received")
		cancel()
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	// Stop discovery engine
	if err := discoveryEngine.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping discovery engine: %v", err)
	}

	log.Println("Server stopped")
}

func setupLogging(level string) {
	// Configure logging based on level
	switch level {
	case "debug":
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	case "info":
		log.SetFlags(log.Ldate | log.Ltime)
	case "warn", "error":
		log.SetFlags(log.Ldate | log.Ltime)
	default:
		log.SetFlags(log.Ldate | log.Ltime)
	}

	// Set prefix
	log.SetPrefix(fmt.Sprintf("[%s] ", level))
}
//go:build !test && !nodiscovery

package mcp

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/state"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/validation"
)

// Server implements the Model Context Protocol server
type Server struct {
	// Core services - interfaces from Track 1
	discovery discovery.DiscoveryEngine
	// TODO: Add patterns, query, dashboard when Track 3/4 complete
	
	// New Relic client for direct API access
	nrClient interface{} // Will be *newrelic.Client when imported
	
	// MCP components
	transport    Transport
	tools        ToolRegistry
	sessions     SessionManager
	protocol     *ProtocolHandler
	stateManager state.StateManager
	
	// Configuration
	config     ServerConfig
	
	// Internal state
	mu         sync.RWMutex
	running    bool
	shutdownCh chan struct{}
	
	// Validation
	nrqlValidator *validation.NRQLValidator
	
	// Mock data generator for development/testing
	mockGenerator *MockDataGenerator
}

// NewServer creates a new MCP server instance
func NewServer(config ServerConfig) *Server {
	s := &Server{
		config:        config,
		tools:         NewToolRegistry(),
		sessions:      NewSessionManager(),
		shutdownCh:    make(chan struct{}),
		nrqlValidator: validation.NewNRQLValidator(),
		mockGenerator: NewMockDataGenerator(),
	}
	
	s.protocol = &ProtocolHandler{
		server:   s,
		requests: sync.Map{},
	}
	
	return s
}

// SetDiscovery sets the discovery engine from Track 1
func (s *Server) SetDiscovery(engine discovery.DiscoveryEngine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.discovery = engine
}

// SetStateManager sets the state manager
func (s *Server) SetStateManager(sm state.StateManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stateManager = sm
}

// SetNewRelicClient sets the New Relic API client
func (s *Server) SetNewRelicClient(client interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nrClient = client
}

// getNRClient safely returns the New Relic client
func (s *Server) getNRClient() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nrClient
}

// getNRClientWithAccount returns a New Relic client for the specified account
// If accountID is empty, returns the primary client
func (s *Server) getNRClientWithAccount(accountID string) (interface{}, error) {
	client := s.getNRClient()
	if client == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}
	
	// Check if this is a MultiAccountClient using reflection
	clientValue := reflect.ValueOf(client)
	method := clientValue.MethodByName("WithAccount")
	if method.IsValid() {
		// This is a MultiAccountClient
		args := []reflect.Value{reflect.ValueOf(accountID)}
		results := method.Call(args)
		if len(results) != 2 {
			return nil, fmt.Errorf("unexpected return values from WithAccount")
		}
		// Extract client and error
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	}
	
	// Not a MultiAccountClient, return the client as-is
	return client, nil
}

// Start initializes and starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()
	
	// Register all tools
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}
	
	// Initialize transport
	transport, err := s.createTransport()
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}
	s.transport = transport
	
	// Start transport
	if err := s.transport.Start(ctx, s.protocol); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}
	
	// Start background workers
	go s.sessionCleanup(ctx)
	
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	close(s.shutdownCh)
	s.mu.Unlock()
	
	// Close transport
	if s.transport != nil {
		if err := s.transport.Close(); err != nil {
			return fmt.Errorf("failed to close transport: %w", err)
		}
	}
	
	// Cleanup sessions
	if err := s.sessions.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup sessions: %w", err)
	}
	
	return nil
}

// createTransport creates the appropriate transport based on configuration
func (s *Server) createTransport() (Transport, error) {
	switch s.config.TransportType {
	case TransportStdio:
		return NewStdioTransport(), nil
	case TransportHTTP:
		return NewHTTPTransport(fmt.Sprintf("%s:%d", s.config.HTTPHost, s.config.HTTPPort)), nil
	case TransportSSE:
		return NewSSETransport(fmt.Sprintf("%s:%d", s.config.HTTPHost, s.config.HTTPPort)), nil
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", s.config.TransportType)
	}
}

// sessionCleanup periodically cleans up expired sessions
func (s *Server) sessionCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdownCh:
			return
		case <-ticker.C:
			s.sessions.Cleanup()
		}
	}
}

// GetInfo returns server information for MCP discovery
func (s *Server) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Universal Data Synthesizer",
		"version":     "2.0.0",
		"description": "AI-powered New Relic dashboard creation",
		"tools":       s.tools.List(),
	}
}

// getCache returns the cache if available
func (s *Server) getCache() (Cache, bool) {
	// Cache is optional - implement when needed
	return nil, false
}

// getMetrics returns the metrics collector if available
func (s *Server) getMetrics() (Metrics, bool) {
	// Metrics are optional - implement when needed
	return nil, false
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return
	}
	
	s.running = false
	close(s.shutdownCh)
	
	if s.transport != nil {
		s.transport.Close()
	}
}

// isMockMode returns true if the server is running in mock mode
func (s *Server) isMockMode() bool {
	return s.getNRClient() == nil || s.config.MockMode
}

// getMockData generates mock data for a tool
func (s *Server) getMockData(toolName string, params map[string]interface{}) interface{} {
	if s.mockGenerator == nil {
		s.mockGenerator = NewMockDataGenerator()
	}
	return s.mockGenerator.GenerateMockResponse(toolName, params)
}
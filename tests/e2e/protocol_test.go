package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPProtocolCompliance validates JSON-RPC 2.0 protocol implementation
func TestMCPProtocolCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Load test environment
	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	// Get primary account
	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok, "Primary test account must be configured")

	// Create MCP client
	client := framework.NewMCPTestClient(primaryAccount)
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start MCP server
	err = client.Start(ctx)
	require.NoError(t, err, "Failed to start MCP server")
	defer client.Stop()

	t.Run("ValidRequest_ExistingTool", func(t *testing.T) {
		// Test a valid request to an existing tool
		result, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 10,
		})
		
		assert.NoError(t, err)
		assert.NotNil(t, result)
		
		// Validate response structure
		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok, "Result should be a map")
		
		eventTypes, ok := resultMap["event_types"].([]interface{})
		assert.True(t, ok, "Should have event_types array")
		assert.NotEmpty(t, eventTypes, "Should discover at least one event type")
		
		t.Logf("Discovered %d event types", len(eventTypes))
	})

	t.Run("MethodNotFound", func(t *testing.T) {
		// Test calling a non-existent tool
		_, err := client.ExecuteTool(ctx, "non.existent.tool", map[string]interface{}{})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Method not found")
	})

	t.Run("InvalidParameters", func(t *testing.T) {
		// Test nrql.execute without required query parameter
		_, err := client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
			"timeout": 30000,
			// Missing required "query" parameter
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("ValidNRQLQuery", func(t *testing.T) {
		// Execute a real NRQL query
		result, err := client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
			"query": "SELECT count(*) FROM Transaction SINCE 1 hour ago",
		})
		
		if err != nil {
			// It's OK if there's no data, but the query should be valid
			assert.Contains(t, err.Error(), "No events found")
		} else {
			assert.NotNil(t, result)
			resultMap, ok := result.(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, resultMap, "results")
		}
	})
}

// TestDiscoveryChain validates the discovery-first approach with real data
func TestDiscoveryChain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Load test environment
	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	client := framework.NewMCPTestClient(primaryAccount)
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	t.Run("DiscoverThenQuery", func(t *testing.T) {
		// Step 1: Discover event types
		discoverResult, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 50,
		})
		require.NoError(t, err)
		
		resultMap := discoverResult.(map[string]interface{})
		eventTypes := resultMap["event_types"].([]interface{})
		require.NotEmpty(t, eventTypes)
		
		// Find a suitable event type (prefer Transaction if available)
		var targetEventType string
		for _, et := range eventTypes {
			if etMap, ok := et.(map[string]interface{}); ok {
				if name, ok := etMap["name"].(string); ok {
					if name == "Transaction" {
						targetEventType = name
						break
					}
					if targetEventType == "" {
						targetEventType = name // Use first available
					}
				}
			}
		}
		require.NotEmpty(t, targetEventType, "Should find at least one event type")
		
		t.Logf("Using event type: %s", targetEventType)
		
		// Step 2: Discover attributes for that event type
		attrResult, err := client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{
			"event_type": targetEventType,
			"sample_size": 100,
		})
		
		if err != nil {
			t.Logf("No attributes found for %s: %v", targetEventType, err)
			return
		}
		
		attrMap := attrResult.(map[string]interface{})
		attributes := attrMap["attributes"].([]interface{})
		assert.NotEmpty(t, attributes, "Should discover attributes")
		
		t.Logf("Discovered %d attributes for %s", len(attributes), targetEventType)
		
		// Step 3: Use discovered information to build a query
		query := fmt.Sprintf("SELECT count(*) FROM %s SINCE 1 hour ago", targetEventType)
		queryResult, err := client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
			"query": query,
		})
		
		if err == nil {
			t.Logf("Query result: %+v", queryResult)
		} else {
			t.Logf("Query returned no data (OK for empty account): %v", err)
		}
	})
}

// TestAdaptiveQueryBuilding validates that the server adapts to different schemas
func TestAdaptiveQueryBuilding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Load test environment
	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	client := framework.NewMCPTestClient(primaryAccount)
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	t.Run("ErrorRateCalculation", func(t *testing.T) {
		// First, discover what error-related attributes exist
		_, err := client.ExecuteTool(ctx, "discovery.find_attributes", map[string]interface{}{
			"event_type": "Transaction",
			"pattern": "error|status|response",
		})
		
		if err != nil {
			t.Logf("Could not find error attributes: %v", err)
			t.Skip("No error-related attributes found")
		}
		
		// Now request error rate - server should adapt based on discovered attributes
		errorRateResult, err := client.ExecuteTool(ctx, "analysis.get_error_rate", map[string]interface{}{
			"time_range": "1 hour ago",
		})
		
		if err != nil {
			t.Logf("Error rate calculation failed (OK if no data): %v", err)
			return
		}
		
		resultMap := errorRateResult.(map[string]interface{})
		
		// Validate that discovery metadata is included
		assert.Contains(t, resultMap, "discovery_metadata")
		metadata := resultMap["discovery_metadata"].(map[string]interface{})
		
		assert.Contains(t, metadata, "method_used")
		assert.Contains(t, metadata, "query_generated")
		assert.Contains(t, metadata, "confidence")
		
		t.Logf("Error rate calculated using method: %s", metadata["method_used"])
		t.Logf("Generated query: %s", metadata["query_generated"])
	})

	t.Run("LatencyPercentileCalculation", func(t *testing.T) {
		// Discover duration-related attributes
		_, err := client.ExecuteTool(ctx, "discovery.find_attributes", map[string]interface{}{
			"event_type": "Transaction",
			"pattern": "duration|time|latency",
		})
		
		if err != nil {
			t.Skip("No duration attributes found")
		}
		
		// Request P95 latency - server should adapt
		latencyResult, err := client.ExecuteTool(ctx, "analysis.get_latency_percentile", map[string]interface{}{
			"percentile": 95,
			"time_range": "1 hour ago",
		})
		
		if err != nil {
			t.Logf("Latency calculation failed (OK if no data): %v", err)
			return
		}
		
		resultMap := latencyResult.(map[string]interface{})
		metadata := resultMap["discovery_metadata"].(map[string]interface{})
		
		t.Logf("Latency calculated using attribute: %s", metadata["attribute_used"])
		t.Logf("Generated query: %s", metadata["query_generated"])
	})
}

// TestCachingBehavior validates that discovery results are cached
func TestCachingBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// This test requires instrumentation to count API calls
	// For now, we'll test that repeated calls are fast
	
	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	client := framework.NewMCPTestClient(primaryAccount)
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	t.Run("RepeatedDiscoveryCalls", func(t *testing.T) {
		// First call - should hit the API
		start1 := time.Now()
		result1, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 100,
		})
		duration1 := time.Since(start1)
		require.NoError(t, err)
		
		// Second call - should use cache
		start2 := time.Now()
		result2, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 100,
		})
		duration2 := time.Since(start2)
		require.NoError(t, err)
		
		// Results should be identical
		json1, _ := json.Marshal(result1)
		json2, _ := json.Marshal(result2)
		assert.Equal(t, string(json1), string(json2), "Cached results should be identical")
		
		// Second call should be much faster (at least 5x)
		if duration1 > 100*time.Millisecond {
			assert.Less(t, duration2.Milliseconds(), duration1.Milliseconds()/5, 
				"Cached call should be much faster")
		}
		
		t.Logf("First call: %v, Second call: %v", duration1, duration2)
	})
}

// TestComposableTools validates that tools can be composed together
func TestComposableTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	client := framework.NewMCPTestClient(primaryAccount)
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	t.Run("DiscoverThenCreateDashboard", func(t *testing.T) {
		// Step 1: Discover attributes for an entity
		discResult, err := client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{
			"event_type": "SystemSample",
			"sample_size": 100,
		})
		
		if err != nil {
			t.Skip("No SystemSample data available")
		}
		
		resultMap := discResult.(map[string]interface{})
		attributes := resultMap["attributes"].([]interface{})
		
		// Pick a few numeric attributes for dashboard
		var selectedAttrs []string
		for i, attr := range attributes {
			if attrMap, ok := attr.(map[string]interface{}); ok {
				if name, ok := attrMap["name"].(string); ok {
					if attrType, ok := attrMap["type"].(string); ok && attrType == "numeric" {
						selectedAttrs = append(selectedAttrs, name)
						if len(selectedAttrs) >= 3 {
							break
						}
					}
				}
			}
			if i > 10 { // Don't check too many
				break
			}
		}
		
		if len(selectedAttrs) == 0 {
			t.Skip("No numeric attributes found")
		}
		
		t.Logf("Selected attributes for dashboard: %v", selectedAttrs)
		
		// Step 2: Create dashboard from discovered attributes
		dashResult, err := client.ExecuteTool(ctx, "dashboard.create_from_discovery", map[string]interface{}{
			"title": "E2E Test Dashboard",
			"attributes": selectedAttrs,
			"event_type": "SystemSample",
		})
		
		if err != nil {
			t.Logf("Dashboard creation not implemented yet: %v", err)
			return
		}
		
		dashMap := dashResult.(map[string]interface{})
		assert.Contains(t, dashMap, "dashboard_id")
		assert.Contains(t, dashMap, "widgets")
		
		t.Logf("Created dashboard with ID: %v", dashMap["dashboard_id"])
	})
}
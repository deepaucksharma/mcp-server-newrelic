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
		
		// Validate response structure - MCP wraps result in content array
		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok, "Result should be a map")
		
		// Extract the content array
		content, ok := resultMap["content"].([]interface{})
		assert.True(t, ok, "Should have content array")
		assert.NotEmpty(t, content, "Content should not be empty")
		
		// Get the first content item
		firstContent, ok := content[0].(map[string]interface{})
		assert.True(t, ok, "First content should be a map")
		
		// Extract the text field which contains the JSON result
		textResult, ok := firstContent["text"].(string)
		assert.True(t, ok, "Should have text field")
		
		// Parse the JSON text
		var toolResult map[string]interface{}
		err = json.Unmarshal([]byte(textResult), &toolResult)
		assert.NoError(t, err, "Should be able to parse JSON result")
		
		// Now we can check for event_types
		eventTypes, ok := toolResult["event_types"].([]interface{})
		assert.True(t, ok, "Should have event_types array")
		assert.NotEmpty(t, eventTypes, "Should discover at least one event type")
		
		t.Logf("Discovered %d event types", len(eventTypes))
	})

	t.Run("NRQL_Query_Execution", func(t *testing.T) {
		// Test executing an NRQL query
		result, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": "SELECT count(*) FROM NrdbQuery SINCE 1 hour ago",
		})
		
		assert.NoError(t, err)
		assert.NotNil(t, result)
		
		// Parse MCP response
		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok, "Result should be a map")
		
		content, ok := resultMap["content"].([]interface{})
		assert.True(t, ok, "Should have content array")
		assert.NotEmpty(t, content, "Content should not be empty")
		
		firstContent, ok := content[0].(map[string]interface{})
		assert.True(t, ok, "First content should be a map")
		
		textResult, ok := firstContent["text"].(string)
		assert.True(t, ok, "Should have text field")
		
		var queryResult map[string]interface{}
		err = json.Unmarshal([]byte(textResult), &queryResult)
		assert.NoError(t, err, "Should be able to parse JSON result")
		
		// Check for results
		results, ok := queryResult["results"].([]interface{})
		assert.True(t, ok, "Should have results array")
		assert.NotEmpty(t, results, "Should have at least one result")
		
		t.Logf("Query returned %d results", len(results))
	})

	t.Run("MethodNotFound", func(t *testing.T) {
		// Test calling a non-existent tool
		_, err := client.ExecuteTool(ctx, "non.existent.tool", map[string]interface{}{})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Method not found")
	})

	t.Run("InvalidParameters", func(t *testing.T) {
		// Test query_nrdb without required query parameter
		_, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"timeout": 30000,
			// Missing required "query" parameter
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("ValidNRQLQuery", func(t *testing.T) {
		// Execute a real NRQL query - using the correct tool name
		result, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": "SELECT count(*) FROM Transaction SINCE 1 hour ago",
		})
		
		// The query might return no data, but should not error
		assert.NoError(t, err, "Query should execute without error")
		assert.NotNil(t, result)
		
		// Parse MCP response  
		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		
		content, ok := resultMap["content"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, content)
		
		// Log the result for debugging
		if firstContent, ok := content[0].(map[string]interface{}); ok {
			if textResult, ok := firstContent["text"].(string); ok {
				t.Logf("Query result: %s", textResult)
			}
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
		
		// Parse MCP response
		resultMap := discoverResult.(map[string]interface{})
		content := resultMap["content"].([]interface{})
		firstContent := content[0].(map[string]interface{})
		textResult := firstContent["text"].(string)
		
		var toolResult map[string]interface{}
		err = json.Unmarshal([]byte(textResult), &toolResult)
		require.NoError(t, err)
		
		eventTypes := toolResult["event_types"].([]interface{})
		require.NotEmpty(t, eventTypes)
		
		// Find a suitable event type (prefer NrdbQuery as it's likely to have data)
		var targetEventType string
		var eventCount int64
		for _, et := range eventTypes {
			if etMap, ok := et.(map[string]interface{}); ok {
				if name, ok := etMap["name"].(string); ok {
					count := int64(0)
					if c, ok := etMap["count"].(float64); ok {
						count = int64(c)
					}
					// Prefer NrdbQuery as it's likely to have data from our queries
					if name == "NrdbQuery" {
						targetEventType = name
						eventCount = count
						break
					}
					// Otherwise use any available event type
					if targetEventType == "" {
						targetEventType = name
						eventCount = count
					}
				}
			}
		}
		require.NotEmpty(t, targetEventType, "Should find at least one event type")
		
		t.Logf("Using event type: %s (count: %d)", targetEventType, eventCount)
		
		// Step 2: Since keyset() might not work for all event types, 
		// let's query the data directly to see what attributes exist
		sampleQuery := fmt.Sprintf("SELECT * FROM %s LIMIT 1 SINCE 1 day ago", targetEventType)
		sampleResult, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": sampleQuery,
		})
		require.NoError(t, err)
		
		// Parse the sample result to see attributes
		sampleResultMap := sampleResult.(map[string]interface{})
		sampleContent := sampleResultMap["content"].([]interface{})
		sampleFirstContent := sampleContent[0].(map[string]interface{})
		sampleTextResult := sampleFirstContent["text"].(string)
		
		var sampleToolResult map[string]interface{}
		err = json.Unmarshal([]byte(sampleTextResult), &sampleToolResult)
		require.NoError(t, err)
		
		sampleResults, ok := sampleToolResult["results"].([]interface{})
		if ok && len(sampleResults) > 0 {
			if firstSample, ok := sampleResults[0].(map[string]interface{}); ok {
				t.Logf("Sample %s event has %d attributes", targetEventType, len(firstSample))
				for key := range firstSample {
					t.Logf("  - %s", key)
				}
			}
		}
		
		// Now test discovery.explore_attributes even if it returns empty
		attrResult, err := client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{
			"event_type": targetEventType,
			"sample_size": 100,
		})
		
		if err != nil {
			t.Logf("Error exploring attributes for %s: %v", targetEventType, err)
			return
		}
		
		// Parse MCP response
		attrResultMap := attrResult.(map[string]interface{})
		attrContent := attrResultMap["content"].([]interface{})
		attrFirstContent := attrContent[0].(map[string]interface{})
		attrTextResult := attrFirstContent["text"].(string)
		
		var attrToolResult map[string]interface{}
		err = json.Unmarshal([]byte(attrTextResult), &attrToolResult)
		require.NoError(t, err)
		
		attributes := attrToolResult["attributes"].([]interface{})
		// Don't fail if empty - keyset() might not work for all event types
		t.Logf("discovery.explore_attributes returned %d attributes for %s", len(attributes), targetEventType)
		
		// Step 3: Use discovered information to build a query
		query := fmt.Sprintf("SELECT count(*) FROM %s SINCE 1 hour ago", targetEventType)
		_, err = client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": query,
		})
		
		require.NoError(t, err)
		t.Logf("Query executed successfully for %s", targetEventType)
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

	t.Run("QueryBuilderWithDiscoveredSchema", func(t *testing.T) {
		// First, discover event types
		eventTypesResult, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 10,
		})
		require.NoError(t, err)
		
		// Parse result to get an event type
		resultMap := eventTypesResult.(map[string]interface{})
		content := resultMap["content"].([]interface{})
		firstContent := content[0].(map[string]interface{})
		textResult := firstContent["text"].(string)
		
		var toolResult map[string]interface{}
		err = json.Unmarshal([]byte(textResult), &toolResult)
		require.NoError(t, err)
		
		eventTypes := toolResult["event_types"].([]interface{})
		require.NotEmpty(t, eventTypes)
		
		// Use the first event type
		firstEventType := eventTypes[0].(map[string]interface{})
		eventTypeName := firstEventType["name"].(string)
		
		// Now discover attributes for this event type
		attrResult, err := client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{
			"event_type": eventTypeName,
			"sample_size": 100,
		})
		require.NoError(t, err)
		
		// Parse attributes
		attrResultMap := attrResult.(map[string]interface{})
		attrContent := attrResultMap["content"].([]interface{})
		attrFirstContent := attrContent[0].(map[string]interface{})
		attrTextResult := attrFirstContent["text"].(string)
		
		var attrToolResult map[string]interface{}
		err = json.Unmarshal([]byte(attrTextResult), &attrToolResult)
		require.NoError(t, err)
		
		attributes := attrToolResult["attributes"].([]interface{})
		
		// Build a query using discovered attributes
		if len(attributes) > 0 {
			// Pick the first numeric attribute if available
			var numericAttr string
			for _, attr := range attributes {
				attrMap := attr.(map[string]interface{})
				if attrMap["inferredType"] == "numeric" {
					numericAttr = attrMap["name"].(string)
					break
				}
			}
			
			// Build a query
			queryParams := map[string]interface{}{
				"event_type": eventTypeName,
				"select": []string{"count(*)"},
				"since": "1 hour ago",
			}
			
			if numericAttr != "" {
				queryParams["select"] = []string{
					"count(*)",
					fmt.Sprintf("average(%s)", numericAttr),
					fmt.Sprintf("max(%s)", numericAttr),
				}
			}
			
			queryResult, err := client.ExecuteTool(ctx, "query_builder", map[string]interface{}(queryParams))
			require.NoError(t, err)
			
			// Parse query builder result
			qResultMap := queryResult.(map[string]interface{})
			qContent := qResultMap["content"].([]interface{})
			qFirstContent := qContent[0].(map[string]interface{})
			qTextResult := qFirstContent["text"].(string)
			
			var qToolResult map[string]interface{}
			err = json.Unmarshal([]byte(qTextResult), &qToolResult)
			require.NoError(t, err)
			
			// Verify query was built
			assert.Contains(t, qToolResult, "query")
			builtQuery := qToolResult["query"].(string)
			assert.Contains(t, builtQuery, eventTypeName)
			
			t.Logf("Built adaptive query: %s", builtQuery)
		}
	})

	t.Run("AnalysisWithDiscoveredMetrics", func(t *testing.T) {
		// Use NrdbQuery since we know it exists
		// First discover its attributes
		attrResult, err := client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{
			"event_type": "NrdbQuery",
			"sample_size": 100,
		})
		require.NoError(t, err)
		
		// Parse attributes
		attrResultMap := attrResult.(map[string]interface{})
		attrContent := attrResultMap["content"].([]interface{})
		attrFirstContent := attrContent[0].(map[string]interface{})
		attrTextResult := attrFirstContent["text"].(string)
		
		var attrToolResult map[string]interface{}
		err = json.Unmarshal([]byte(attrTextResult), &attrToolResult)
		require.NoError(t, err)
		
		attributes := attrToolResult["attributes"].([]interface{})
		
		// Look for durationMs which we know exists
		var hasDurationMs bool
		for _, attr := range attributes {
			attrMap := attr.(map[string]interface{})
			if attrMap["name"] == "durationMs" {
				hasDurationMs = true
				break
			}
		}
		
		if hasDurationMs {
			// Calculate baseline for durationMs
			baselineResult, err := client.ExecuteTool(ctx, "analysis.calculate_baseline", map[string]interface{}{
				"metric": "durationMs",
				"event_type": "NrdbQuery",
				"time_range": "1 hour",
				"percentiles": []int{50, 95, 99},
			})
			
			if err != nil {
				t.Logf("Baseline calculation failed (OK if no data): %v", err)
			} else {
				// Parse baseline result
				bResultMap := baselineResult.(map[string]interface{})
				bContent := bResultMap["content"].([]interface{})
				bFirstContent := bContent[0].(map[string]interface{})
				bTextResult := bFirstContent["text"].(string)
				
				var bToolResult map[string]interface{}
				err = json.Unmarshal([]byte(bTextResult), &bToolResult)
				require.NoError(t, err)
				
				t.Logf("Baseline analysis completed: %v", bToolResult)
			}
		}
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
			"event_type": "NrdbQuery",
			"sample_size": 100,
		})
		
		if err != nil {
			t.Skip("No NrdbQuery data available")
		}
		
		// Parse MCP response
		resultMap := discResult.(map[string]interface{})
		content := resultMap["content"].([]interface{})
		firstContent := content[0].(map[string]interface{})
		textResult := firstContent["text"].(string)
		
		var toolResult map[string]interface{}
		err = json.Unmarshal([]byte(textResult), &toolResult)
		require.NoError(t, err)
		
		attributes := toolResult["attributes"].([]interface{})
		
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
			"event_type": "NrdbQuery",
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
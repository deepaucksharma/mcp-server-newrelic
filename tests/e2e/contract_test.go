package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPContractCompliance validates API contracts for all tools
func TestMCPContractCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	// Load test environment
	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	client := framework.NewMCPTestClient(primaryAccount)
	ctx := context.Background()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	t.Run("DiscoveryToolsContract", func(t *testing.T) {
		t.Run("explore_event_types", func(t *testing.T) {
			// Test required response structure
			result, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
				"limit": 5,
			})
			require.NoError(t, err)

			// Validate response contract
			resultMap := parseToolResult(t, result)
			
			// Must have event_types array
			eventTypes, ok := resultMap["event_types"].([]interface{})
			assert.True(t, ok, "Response must contain event_types array")
			assert.NotEmpty(t, eventTypes, "Should return at least one event type")

			// Each event type must have required fields
			for _, et := range eventTypes {
				etMap, ok := et.(map[string]interface{})
				require.True(t, ok, "Event type must be a map")
				
				assert.Contains(t, etMap, "name", "Event type must have name")
				assert.Contains(t, etMap, "count", "Event type must have count")
				
				// Optional fields
				if _, ok := etMap["sample_query"]; ok {
					assert.IsType(t, "", etMap["sample_query"], "sample_query must be string")
				}
			}

			// Must have metadata
			if metadata, ok := resultMap["metadata"].(map[string]interface{}); ok {
				assert.Contains(t, metadata, "total_types", "Metadata should contain total_types")
				assert.Contains(t, metadata, "time_range", "Metadata should contain time_range")
			}
		})

		t.Run("explore_attributes", func(t *testing.T) {
			// Test required parameters
			_, err := client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{})
			assert.Error(t, err, "Should error without required event_type parameter")

			// Test valid request
			result, err := client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{
				"event_type": "NrdbQuery",
			})
			require.NoError(t, err)

			// Validate response contract
			resultMap := parseToolResult(t, result)
			
			// Must have attributes array
			attributes, ok := resultMap["attributes"].([]interface{})
			assert.True(t, ok, "Response must contain attributes array")

			// Each attribute must have required fields
			for _, attr := range attributes {
				attrMap, ok := attr.(map[string]interface{})
				require.True(t, ok, "Attribute must be a map")
				
				assert.Contains(t, attrMap, "name", "Attribute must have name")
				assert.Contains(t, attrMap, "type", "Attribute must have type")
				
				// Type must be valid
				validTypes := []string{"string", "numeric", "boolean", "timestamp"}
				attrType, _ := attrMap["type"].(string)
				assert.Contains(t, validTypes, attrType, "Attribute type must be valid")
			}
		})
	})

	t.Run("QueryToolsContract", func(t *testing.T) {
		t.Run("nrql.execute", func(t *testing.T) {
			// Test required parameters
			_, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{})
			assert.Error(t, err, "Should error without required query parameter")

			// Test valid query
			result, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
				"query": "SELECT count(*) FROM NrdbQuery SINCE 1 hour ago",
			})
			require.NoError(t, err)

			// Validate response contract
			resultMap := parseToolResult(t, result)
			
			// Must have results array
			_, ok = resultMap["results"].([]interface{})
			assert.True(t, ok, "Response must contain results array")

			// Must have metadata
			metadata, ok := resultMap["metadata"].(map[string]interface{})
			assert.True(t, ok, "Response must contain metadata")
			
			// Metadata must have required fields
			assert.Contains(t, metadata, "eventTypes", "Metadata must contain eventTypes")
			assert.Contains(t, metadata, "messages", "Metadata must contain messages")
		})
	})

	t.Run("AnalysisToolsContract", func(t *testing.T) {
		t.Run("calculate_baseline", func(t *testing.T) {
			// Test required parameters
			_, err := client.ExecuteTool(ctx, "analysis.calculate_baseline", map[string]interface{}{
				"metric": "duration",
				// Missing required event_type
			})
			assert.Error(t, err, "Should error without required event_type parameter")

			// Test valid request
			result, err := client.ExecuteTool(ctx, "analysis.calculate_baseline", map[string]interface{}{
				"metric":     "duration",
				"event_type": "NrdbQuery",
			})
			require.NoError(t, err)

			// Validate response contract
			resultMap := parseToolResult(t, result)
			
			// Must have baseline statistics
			assert.Contains(t, resultMap, "metric", "Response must contain metric name")
			assert.Contains(t, resultMap, "recommendations", "Response must contain recommendations")
			
			// Check for statistical fields (may be in different formats)
			hasStats := false
			if _, ok := resultMap["avg"]; ok {
				hasStats = true
				assert.Contains(t, resultMap, "count", "Should have count with avg")
			}
			if _, ok := resultMap["groups"]; ok {
				hasStats = true
				// Grouped results have different structure
			}
			assert.True(t, hasStats, "Response must contain statistical data")
		})

		t.Run("detect_anomalies", func(t *testing.T) {
			result, err := client.ExecuteTool(ctx, "analysis.detect_anomalies", map[string]interface{}{
				"metric":     "duration",
				"event_type": "NrdbQuery",
			})
			require.NoError(t, err)

			// Validate response contract
			resultMap := parseToolResult(t, result)
			
			// Must have required fields
			assert.Contains(t, resultMap, "metric", "Response must contain metric")
			assert.Contains(t, resultMap, "method", "Response must contain method")
			assert.Contains(t, resultMap, "anomalies", "Response must contain anomalies")
			
			// Check anomalies structure
			if anomalies, ok := resultMap["anomalies"].([]interface{}); ok {
				for _, anomaly := range anomalies {
					anomalyMap, ok := anomaly.(map[string]interface{})
					require.True(t, ok, "Anomaly must be a map")
					
					assert.Contains(t, anomalyMap, "timestamp", "Anomaly must have timestamp")
					assert.Contains(t, anomalyMap, "value", "Anomaly must have value")
					assert.Contains(t, anomalyMap, "score", "Anomaly must have score")
				}
			}
		})
	})

	t.Run("ErrorResponseContract", func(t *testing.T) {
		// Test various error scenarios
		errorCases := []struct {
			name   string
			tool   string
			params map[string]interface{}
		}{
			{
				name: "InvalidTool",
				tool: "invalid.tool.name",
				params: map[string]interface{}{},
			},
			{
				name: "InvalidNRQL",
				tool: "query_nrdb",
				params: map[string]interface{}{
					"query": "INVALID NRQL SYNTAX",
				},
			},
			{
				name: "MissingRequiredParam",
				tool: "discovery.explore_attributes",
				params: map[string]interface{}{
					// Missing event_type
				},
			},
		}

		for _, tc := range errorCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := client.ExecuteTool(ctx, tc.tool, tc.params)
				
				// Should get an error
				assert.Error(t, err)
				
				// If result is returned, check error structure
				if result != nil {
					if errMap, ok := result.(map[string]interface{}); ok {
						if mcpErr, ok := errMap["error"].(map[string]interface{}); ok {
							// MCP error must have code and message
							assert.Contains(t, mcpErr, "code", "Error must have code")
							assert.Contains(t, mcpErr, "message", "Error must have message")
							
							// Code must be integer
							_, ok := mcpErr["code"].(float64)
							assert.True(t, ok, "Error code must be number")
						}
					}
				}
			})
		}
	})

	t.Run("ResponseMetadataContract", func(t *testing.T) {
		// All tool responses should include consistent metadata
		tools := []struct {
			name   string
			params map[string]interface{}
		}{
			{
				name: "discovery.explore_event_types",
				params: map[string]interface{}{
					"limit": 5,
				},
			},
			{
				name: "query_nrdb",
				params: map[string]interface{}{
					"query": "SELECT count(*) FROM NrdbQuery SINCE 1 hour ago",
				},
			},
		}

		for _, tool := range tools {
			t.Run(tool.name, func(t *testing.T) {
				result, err := client.ExecuteTool(ctx, tool.name, tool.params)
				require.NoError(t, err)

				// All responses should be parseable
				resultMap := parseToolResult(t, result)
				assert.NotNil(t, resultMap, "Response must be valid JSON")

				// Check for common metadata patterns
				if metadata, ok := resultMap["metadata"].(map[string]interface{}); ok {
					// If metadata exists, it should have useful information
					assert.NotEmpty(t, metadata, "Metadata should not be empty")
				}
			})
		}
	})
}

// parseToolResult extracts the tool result from MCP response
func parseToolResult(t *testing.T, result interface{}) map[string]interface{} {
	// Handle MCP response format
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "Result must be a map")

	// Extract content
	content, ok := resultMap["content"].([]interface{})
	require.True(t, ok, "Result must have content array")
	require.NotEmpty(t, content, "Content must not be empty")

	// Get first content item
	firstContent, ok := content[0].(map[string]interface{})
	require.True(t, ok, "Content item must be a map")

	// Extract text
	text, ok := firstContent["text"].(string)
	require.True(t, ok, "Content must have text")

	// Parse JSON from text
	var toolResult map[string]interface{}
	err := json.Unmarshal([]byte(text), &toolResult)
	require.NoError(t, err, "Tool result must be valid JSON")

	return toolResult
}
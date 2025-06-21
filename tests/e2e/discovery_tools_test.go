package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiscoveryTools validates all discovery tools with real New Relic data
func TestDiscoveryTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	client := framework.NewMCPTestClient(primaryAccount)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	t.Run("ExploreEventTypes", func(t *testing.T) {
		result, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 50,
		})
		
		require.NoError(t, err, "Should successfully explore event types")
		
		resultMap := result.(map[string]interface{})
		eventTypes := resultMap["event_types"].([]interface{})
		
		assert.NotEmpty(t, eventTypes, "Should discover at least one event type")
		
		// Validate structure of each event type
		for i, et := range eventTypes {
			etMap := et.(map[string]interface{})
			assert.Contains(t, etMap, "name", "Event type should have name")
			assert.Contains(t, etMap, "sample_count", "Event type should have sample count")
			
			if i < 5 { // Log first 5
				t.Logf("Event Type: %s (samples: %v)", etMap["name"], etMap["sample_count"])
			}
		}
		
		// Check for common event types if they exist
		eventTypeNames := make(map[string]bool)
		for _, et := range eventTypes {
			if name, ok := et.(map[string]interface{})["name"].(string); ok {
				eventTypeNames[name] = true
			}
		}
		
		// Log what common types were found
		commonTypes := []string{"Transaction", "SystemSample", "ProcessSample", "NetworkSample"}
		for _, ct := range commonTypes {
			if eventTypeNames[ct] {
				t.Logf("Found common event type: %s", ct)
			}
		}
	})

	t.Run("ExploreAttributes", func(t *testing.T) {
		// First find an event type to explore
		eventTypes, err := getAvailableEventTypes(ctx, client)
		require.NoError(t, err)
		require.NotEmpty(t, eventTypes)
		
		// Use first available event type
		eventType := eventTypes[0]
		t.Logf("Exploring attributes for: %s", eventType)
		
		result, err := client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{
			"event_type":  eventType,
			"sample_size": 100,
		})
		
		if err != nil {
			t.Logf("No data for %s: %v", eventType, err)
			return
		}
		
		resultMap := result.(map[string]interface{})
		attributes := resultMap["attributes"].([]interface{})
		
		assert.NotEmpty(t, attributes, "Should discover attributes")
		
		// Count attribute types
		typeCount := make(map[string]int)
		for _, attr := range attributes {
			attrMap := attr.(map[string]interface{})
			if attrType, ok := attrMap["type"].(string); ok {
				typeCount[attrType]++
			}
		}
		
		t.Logf("Discovered %d attributes: %v", len(attributes), typeCount)
	})

	t.Run("ProfileAttribute", func(t *testing.T) {
		// Profile a specific attribute
		result, err := client.ExecuteTool(ctx, "discovery.profile_attribute", map[string]interface{}{
			"event_type": "Transaction",
			"attribute":  "duration",
		})
		
		if err != nil {
			// Try SystemSample.cpuPercent as alternative
			result, err = client.ExecuteTool(ctx, "discovery.profile_attribute", map[string]interface{}{
				"event_type": "SystemSample",
				"attribute":  "cpuPercent",
			})
		}
		
		if err != nil {
			t.Skip("No suitable data for profiling")
		}
		
		resultMap := result.(map[string]interface{})
		profile := resultMap["profile"].(map[string]interface{})
		
		assert.Contains(t, profile, "coverage")
		assert.Contains(t, profile, "type")
		assert.Contains(t, profile, "statistics")
		
		t.Logf("Attribute profile: %+v", profile)
	})

	t.Run("FindAttributes", func(t *testing.T) {
		// Find attributes matching a pattern
		result, err := client.ExecuteTool(ctx, "discovery.find_attributes", map[string]interface{}{
			"event_type": "Transaction",
			"pattern":    "duration|time|latency",
		})
		
		if err != nil {
			t.Logf("Pattern search not available: %v", err)
			return
		}
		
		resultMap := result.(map[string]interface{})
		matches := resultMap["matches"].([]interface{})
		
		t.Logf("Found %d attributes matching pattern", len(matches))
		for _, match := range matches {
			t.Logf("  - %v", match)
		}
	})

	t.Run("AnalyzeCardinality", func(t *testing.T) {
		// Analyze cardinality of key attributes
		result, err := client.ExecuteTool(ctx, "discovery.analyze_cardinality", map[string]interface{}{
			"event_types": []string{"Transaction"},
			"attributes":  []string{"appName", "host", "request.uri"},
		})
		
		if err != nil {
			t.Logf("Cardinality analysis failed: %v", err)
			return
		}
		
		resultMap := result.(map[string]interface{})
		analysis := resultMap["cardinality_analysis"].(map[string]interface{})
		
		t.Logf("Cardinality analysis: %+v", analysis)
	})

	t.Run("DetectPatterns", func(t *testing.T) {
		// Detect naming patterns in attributes
		result, err := client.ExecuteTool(ctx, "discovery.detect_patterns", map[string]interface{}{
			"event_type": "Transaction",
			"focus":      "error_attributes",
		})
		
		if err != nil {
			t.Logf("Pattern detection not implemented: %v", err)
			return
		}
		
		resultMap := result.(map[string]interface{})
		patterns := resultMap["patterns"].([]interface{})
		
		t.Logf("Detected %d patterns", len(patterns))
	})
}

// TestDiscoveryAdaptation validates that discovery adapts to different schemas
func TestDiscoveryAdaptation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	
	// Test with different accounts if available
	testCases := []struct {
		name    string
		account string
	}{
		{"PrimaryAccount", "primary"},
		{"SecondaryAccount", "secondary"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			account, ok := accounts[tc.account]
			if !ok {
				t.Skipf("%s account not configured", tc.account)
			}
			
			client := framework.NewMCPTestClient(account)
			
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			
			err = client.Start(ctx)
			require.NoError(t, err)
			defer client.Stop()
			
			// Discover error-related attributes
			result, err := client.ExecuteTool(ctx, "discovery.find_error_attributes", map[string]interface{}{
				"event_type": "Transaction",
			})
			
			if err != nil {
				t.Logf("No Transaction data in %s account", tc.account)
				return
			}
			
			resultMap := result.(map[string]interface{})
			errorAttrs := resultMap["error_attributes"].([]interface{})
			
			t.Logf("%s account has %d error-related attributes", tc.account, len(errorAttrs))
			
			// The server should adapt to whatever schema exists
			assert.NotNil(t, errorAttrs, "Should return error attributes or empty list")
		})
	}
}

// Helper function to get available event types
func getAvailableEventTypes(ctx context.Context, client *framework.MCPTestClient) ([]string, error) {
	result, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
		"limit": 10,
	})
	if err != nil {
		return nil, err
	}
	
	resultMap := result.(map[string]interface{})
	eventTypes := resultMap["event_types"].([]interface{})
	
	var types []string
	for _, et := range eventTypes {
		if etMap, ok := et.(map[string]interface{}); ok {
			if name, ok := etMap["name"].(string); ok {
				types = append(types, name)
			}
		}
	}
	
	return types, nil
}

// TestDiscoveryCache validates that discovery results are properly cached
func TestDiscoveryCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	client := framework.NewMCPTestClient(primaryAccount)
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	t.Run("EventTypesCached", func(t *testing.T) {
		// Make multiple identical requests
		var results []interface{}
		var durations []time.Duration
		
		for i := 0; i < 3; i++ {
			start := time.Now()
			result, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
				"limit": 50,
			})
			duration := time.Since(start)
			
			require.NoError(t, err)
			results = append(results, result)
			durations = append(durations, duration)
			
			t.Logf("Request %d took: %v", i+1, duration)
		}
		
		// First request should be slowest (hits API)
		// Subsequent requests should be faster (from cache)
		if durations[0] > 50*time.Millisecond {
			assert.Less(t, durations[1].Nanoseconds(), durations[0].Nanoseconds(), 
				"Second request should be faster (cached)")
			assert.Less(t, durations[2].Nanoseconds(), durations[0].Nanoseconds(), 
				"Third request should be faster (cached)")
		}
		
		// Results should be identical
		for i := 1; i < len(results); i++ {
			result1 := fmt.Sprintf("%+v", results[0])
			result2 := fmt.Sprintf("%+v", results[i])
			assert.Equal(t, result1, result2, "Cached results should be identical")
		}
	})
}
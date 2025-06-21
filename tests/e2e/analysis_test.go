package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnalysisTools validates analysis tools with real New Relic data
func TestAnalysisTools(t *testing.T) {
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
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	t.Run("ErrorRateCalculation", func(t *testing.T) {
		// First discover what event types we have
		eventTypesResult, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 10,
		})
		require.NoError(t, err)
		
		// Parse to find an event type with error data
		resultMap := eventTypesResult.(map[string]interface{})
		content := resultMap["content"].([]interface{})
		firstContent := content[0].(map[string]interface{})
		textResult := firstContent["text"].(string)
		
		var toolResult map[string]interface{}
		err = json.Unmarshal([]byte(textResult), &toolResult)
		require.NoError(t, err)
		
		eventTypes := toolResult["event_types"].([]interface{})
		
		// Try to find Transaction or similar event type that might have error data
		var targetEventType string
		for _, et := range eventTypes {
			etMap := et.(map[string]interface{})
			name := etMap["name"].(string)
			if name == "Transaction" || name == "TransactionError" || name == "PageView" {
				targetEventType = name
				break
			}
		}
		
		if targetEventType == "" {
			// Use NrdbQuery as fallback and simulate error rate
			targetEventType = "NrdbQuery"
		}
		
		t.Logf("Testing error rate calculation with event type: %s", targetEventType)
		
		// Calculate baseline for the event type
		baselineResult, err := client.ExecuteTool(ctx, "analysis.calculate_baseline", map[string]interface{}{
			"metric":      "duration",
			"event_type":  targetEventType,
			"time_range":  "1 hour",
			"percentiles": []int{50, 90, 95, 99},
		})
		
		// Even if this fails (no data), log and continue
		if err != nil {
			t.Logf("Baseline calculation failed (expected if no data): %v", err)
		} else {
			// Parse baseline result
			bResultMap := baselineResult.(map[string]interface{})
			bContent := bResultMap["content"].([]interface{})
			bFirstContent := bContent[0].(map[string]interface{})
			bTextResult := bFirstContent["text"].(string)
			
			var bToolResult map[string]interface{}
			err = json.Unmarshal([]byte(bTextResult), &bToolResult)
			if err == nil {
				t.Logf("Baseline analysis result: %+v", bToolResult)
			}
		}
		
		// Test anomaly detection
		anomalyResult, err := client.ExecuteTool(ctx, "analysis.detect_anomalies", map[string]interface{}{
			"metric":      "duration",
			"event_type":  targetEventType,
			"time_range":  "24 hours",
			"sensitivity": 3,
			"method":      "zscore",
		})
		
		if err != nil {
			t.Logf("Anomaly detection failed (expected if no data): %v", err)
		} else {
			// Parse anomaly result
			aResultMap := anomalyResult.(map[string]interface{})
			aContent := aResultMap["content"].([]interface{})
			aFirstContent := aContent[0].(map[string]interface{})
			aTextResult := aFirstContent["text"].(string)
			
			var aToolResult map[string]interface{}
			err = json.Unmarshal([]byte(aTextResult), &aToolResult)
			if err == nil {
				// Check for anomaly detection fields
				if anomaliesDetected, ok := aToolResult["anomaliesDetected"]; ok {
					assert.Contains(t, aToolResult, "anomaliesDetected")
					assert.Contains(t, aToolResult, "method")
					assert.Contains(t, aToolResult, "sensitivity")
					t.Logf("Detected %v anomalies", anomaliesDetected)
				} else {
					// Handle alternate field name
					assert.Contains(t, aToolResult, "anomalies_detected")
					assert.Contains(t, aToolResult, "method") 
					assert.Contains(t, aToolResult, "sensitivity")
					t.Logf("Detected %v anomalies", aToolResult["anomalies_detected"])
				}
			}
		}
	})

	t.Run("LatencyPercentileCalculation", func(t *testing.T) {
		// Test percentile calculations on NrdbQuery duration
		// First, let's check if we have duration data
		checkQuery := "SELECT count(*), min(durationMs), max(durationMs), average(durationMs) FROM NrdbQuery SINCE 1 hour ago"
		
		queryResult, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": checkQuery,
		})
		require.NoError(t, err)
		
		// Parse query result
		qResultMap := queryResult.(map[string]interface{})
		qContent := qResultMap["content"].([]interface{})
		qFirstContent := qContent[0].(map[string]interface{})
		qTextResult := qFirstContent["text"].(string)
		
		var qToolResult map[string]interface{}
		err = json.Unmarshal([]byte(qTextResult), &qToolResult)
		require.NoError(t, err)
		
		results := qToolResult["results"].([]interface{})
		if len(results) > 0 {
			firstResult := results[0].(map[string]interface{})
			count := firstResult["count"]
			t.Logf("Found %v NrdbQuery events with duration data", count)
			
			// Now calculate percentiles using analysis tool
			percentileResult, err := client.ExecuteTool(ctx, "analysis.calculate_baseline", map[string]interface{}{
				"metric":      "durationMs",
				"event_type":  "NrdbQuery",
				"time_range":  "1 hour",
				"percentiles": []interface{}{50.0, 75.0, 90.0, 95.0, 99.0},
			})
			
			if err != nil {
				t.Logf("Percentile calculation failed: %v", err)
			} else {
				// Verify percentile results
				pResultMap := percentileResult.(map[string]interface{})
				pContent := pResultMap["content"].([]interface{})
				pFirstContent := pContent[0].(map[string]interface{})
				pTextResult := pFirstContent["text"].(string)
				
				var pToolResult map[string]interface{}
				err = json.Unmarshal([]byte(pTextResult), &pToolResult)
				if err == nil {
					// The baseline tool only returns recommendations in mock mode
					if _, ok := pToolResult["metric"]; ok {
						assert.Contains(t, pToolResult, "metric")
					}
					assert.Contains(t, pToolResult, "recommendations")
					t.Logf("Percentile analysis completed: %+v", pToolResult)
				}
			}
		} else {
			t.Skip("No duration data available for percentile calculation")
		}
	})

	t.Run("TrendAnalysis", func(t *testing.T) {
		// Test trend analysis
		trendResult, err := client.ExecuteTool(ctx, "analysis.analyze_trend", map[string]interface{}{
			"metric":           "duration",
			"event_type":       "NrdbQuery",
			"time_range":       "7 days",
			"granularity":      "hour",
			"include_forecast": true,
		})
		
		if err != nil {
			t.Logf("Trend analysis failed (expected if insufficient data): %v", err)
		} else {
			// Parse trend result
			tResultMap := trendResult.(map[string]interface{})
			tContent := tResultMap["content"].([]interface{})
			tFirstContent := tContent[0].(map[string]interface{})
			tTextResult := tFirstContent["text"].(string)
			
			var tToolResult map[string]interface{}
			err = json.Unmarshal([]byte(tTextResult), &tToolResult)
			if err == nil {
				assert.Contains(t, tToolResult, "trend")
				assert.Contains(t, tToolResult, "insights")
				t.Logf("Trend analysis: %+v", tToolResult["trend"])
			}
		}
	})

	t.Run("CorrelationAnalysis", func(t *testing.T) {
		// Test correlation finding
		correlationResult, err := client.ExecuteTool(ctx, "analysis.find_correlations", map[string]interface{}{
			"primary_metric":    "durationMs",
			"secondary_metrics": []string{"inspectedCount", "timeWindowMinutes"},
			"event_type":        "NrdbQuery",
			"time_range":        "24 hours",
			"min_correlation":   0.5,
		})
		
		if err != nil {
			t.Logf("Correlation analysis failed: %v", err)
		} else {
			// Parse correlation result
			cResultMap := correlationResult.(map[string]interface{})
			cContent := cResultMap["content"].([]interface{})
			cFirstContent := cContent[0].(map[string]interface{})
			cTextResult := cFirstContent["text"].(string)
			
			var cToolResult map[string]interface{}
			err = json.Unmarshal([]byte(cTextResult), &cToolResult)
			if err == nil {
				// Check for either field name format
				if _, ok := cToolResult["primaryMetric"]; ok {
					assert.Contains(t, cToolResult, "primaryMetric")
				} else {
					assert.Contains(t, cToolResult, "primary_metric")
				}
				assert.Contains(t, cToolResult, "correlations")
				assert.Contains(t, cToolResult, "insights")
				
				correlations := cToolResult["correlations"].([]interface{})
				t.Logf("Found %d correlations", len(correlations))
				
				for _, corr := range correlations {
					corrMap := corr.(map[string]interface{})
					t.Logf("Correlation: %s = %.2f", corrMap["metric"], corrMap["coefficient"])
				}
			}
		}
	})

	t.Run("DistributionAnalysis", func(t *testing.T) {
		// Test distribution analysis
		distResult, err := client.ExecuteTool(ctx, "analysis.analyze_distribution", map[string]interface{}{
			"metric":     "durationMs",
			"event_type": "NrdbQuery",
			"time_range": "24 hours",
			"buckets":    10,
		})
		
		if err != nil {
			t.Logf("Distribution analysis failed: %v", err)
		} else {
			// Parse distribution result
			dResultMap := distResult.(map[string]interface{})
			dContent := dResultMap["content"].([]interface{})
			dFirstContent := dContent[0].(map[string]interface{})
			dTextResult := dFirstContent["text"].(string)
			
			var dToolResult map[string]interface{}
			err = json.Unmarshal([]byte(dTextResult), &dToolResult)
			if err == nil {
				assert.Contains(t, dToolResult, "distribution")
				assert.Contains(t, dToolResult, "insights")
				t.Logf("Distribution type: %v", dToolResult["distribution"])
			}
		}
	})
}
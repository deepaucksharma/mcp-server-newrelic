//go:build integration

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToolsIntegration tests the integration of all tool categories
func TestToolsIntegration(t *testing.T) {
	server := createTestServer(t)
	err := server.registerTools()
	require.NoError(t, err)
	
	ctx := context.Background()
	
	t.Run("Discovery Tools", func(t *testing.T) {
		testDiscoveryTools(t, ctx, server)
	})
	
	t.Run("Query Tools", func(t *testing.T) {
		testQueryTools(t, ctx, server)
	})
	
	t.Run("Analysis Tools", func(t *testing.T) {
		testAnalysisTools(t, ctx, server)
	})
	
	t.Run("Dashboard Tools", func(t *testing.T) {
		testDashboardTools(t, ctx, server)
	})
	
	t.Run("Alert Tools", func(t *testing.T) {
		testAlertTools(t, ctx, server)
	})
	
	t.Run("Governance Tools", func(t *testing.T) {
		testGovernanceTools(t, ctx, server)
	})
	
	t.Run("Bulk Operation Tools", func(t *testing.T) {
		testBulkOperationTools(t, ctx, server)
	})
}

func testDiscoveryTools(t *testing.T, ctx context.Context, server *Server) {
	// Test event type discovery
	tool, err := server.tools.Get("discovery.explore_event_types")
	require.NoError(t, err)
	
	result, err := tool.Handler(ctx, map[string]interface{}{
		"time_range": "24 hours",
		"limit":      10,
	})
	require.NoError(t, err)
	
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	
	eventTypes, ok := resultMap["event_types"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, eventTypes)
	
	// Test attribute exploration
	if len(eventTypes) > 0 {
		firstEvent := eventTypes[0].(map[string]interface{})
		eventType := firstEvent["name"].(string)
		
		tool, err = server.tools.Get("discovery.explore_attributes")
		require.NoError(t, err)
		
		result, err = tool.Handler(ctx, map[string]interface{}{
			"event_type": eventType,
			"sample_size": 100,
		})
		require.NoError(t, err)
		
		resultMap, ok = result.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, resultMap, "attributes")
	}
	
	// Test relationship mining
	tool, err = server.tools.Get("discovery.mine_relationships")
	if err == nil {
		result, err = tool.Handler(ctx, map[string]interface{}{
			"schemas": []string{"Transaction", "PageView"},
		})
		assert.NoError(t, err)
		
		if result != nil {
			resultMap, ok = result.(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, resultMap, "relationships")
		}
	}
}

func testQueryTools(t *testing.T, ctx context.Context, server *Server) {
	// Test basic NRQL execution
	tool, err := server.tools.Get("query_nrdb")
	require.NoError(t, err)
	
	result, err := tool.Handler(ctx, map[string]interface{}{
		"query": "SELECT count(*) FROM Transaction SINCE 1 hour ago",
	})
	require.NoError(t, err)
	
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "results")
	
	// Test adaptive NRQL execution
	tool, err = server.tools.Get("nrql.execute")
	require.NoError(t, err)
	
	result, err = tool.Handler(ctx, map[string]interface{}{
		"query": "SELECT average(duration) FROM Transaction WHERE appName = 'test'",
		"include_metadata": true,
	})
	require.NoError(t, err)
	
	resultMap, ok = result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "results")
	
	if metadata, ok := resultMap["metadata"].(map[string]interface{}); ok {
		assert.Contains(t, metadata, "executionTime")
		assert.Contains(t, metadata, "rowCount")
	}
	
	// Test query validation
	tool, err = server.tools.Get("nrql.validate")
	if err == nil {
		result, err = tool.Handler(ctx, map[string]interface{}{
			"query": "SELECT * FROM InvalidEventType",
		})
		assert.NoError(t, err)
		
		resultMap, ok = result.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, resultMap, "isValid")
		assert.Contains(t, resultMap, "suggestions")
	}
}

func testAnalysisTools(t *testing.T, ctx context.Context, server *Server) {
	// Test anomaly detection
	tool, err := server.tools.Get("analysis.detect_anomalies")
	require.NoError(t, err)
	
	result, err := tool.Handler(ctx, map[string]interface{}{
		"metric":     "duration",
		"event_type": "Transaction",
		"time_range": "24 hours",
		"sensitivity": 3,
	})
	require.NoError(t, err)
	
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "anomaliesDetected")
	assert.Contains(t, resultMap, "recommendations")
	
	// Test correlation analysis
	tool, err = server.tools.Get("analysis.find_correlations")
	require.NoError(t, err)
	
	result, err = tool.Handler(ctx, map[string]interface{}{
		"primary_metric": "duration",
		"event_type":     "Transaction",
		"time_range":     "24 hours",
	})
	require.NoError(t, err)
	
	resultMap, ok = result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "correlations")
	
	// Test trend analysis
	tool, err = server.tools.Get("analysis.analyze_trend")
	require.NoError(t, err)
	
	result, err = tool.Handler(ctx, map[string]interface{}{
		"metric":           "count(*)",
		"event_type":       "Transaction",
		"time_range":       "7 days",
		"granularity":      "hour",
		"include_forecast": true,
	})
	require.NoError(t, err)
	
	resultMap, ok = result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "trend")
	assert.Contains(t, resultMap, "insights")
}

func testDashboardTools(t *testing.T, ctx context.Context, server *Server) {
	// Test list dashboards with pagination
	tool, err := server.tools.Get("list_dashboards")
	require.NoError(t, err)
	
	result, err := tool.Handler(ctx, map[string]interface{}{
		"limit": 5,
	})
	require.NoError(t, err)
	
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "dashboards")
	assert.Contains(t, resultMap, "total")
	
	// Check pagination
	if cursor, ok := resultMap["next_cursor"].(string); ok && cursor != "" {
		// Test next page
		result, err = tool.Handler(ctx, map[string]interface{}{
			"limit":  5,
			"cursor": cursor,
		})
		assert.NoError(t, err)
		
		resultMap, ok = result.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, resultMap, "dashboards")
	}
	
	// Test dashboard generation
	tool, err = server.tools.Get("generate_dashboard")
	require.NoError(t, err)
	
	result, err = tool.Handler(ctx, map[string]interface{}{
		"template":     "golden-signals",
		"name":         "Test Dashboard",
		"service_name": "test-service",
	})
	require.NoError(t, err)
	
	resultMap, ok = result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "dashboard")
	
	dashboard, ok := resultMap["dashboard"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, dashboard, "guid")
	assert.Contains(t, dashboard, "name")
}

func testAlertTools(t *testing.T, ctx context.Context, server *Server) {
	// Test list alerts with pagination
	tool, err := server.tools.Get("list_alerts")
	require.NoError(t, err)
	
	result, err := tool.Handler(ctx, map[string]interface{}{
		"limit": 10,
	})
	require.NoError(t, err)
	
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "alerts")
	
	// Test alert analysis
	tool, err = server.tools.Get("analyze_alert_effectiveness")
	if err == nil {
		result, err = tool.Handler(ctx, map[string]interface{}{
			"time_period": "30 days",
		})
		assert.NoError(t, err)
		
		if result != nil {
			resultMap, ok = result.(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, resultMap, "analysis")
		}
	}
}

func testGovernanceTools(t *testing.T, ctx context.Context, server *Server) {
	// Test usage analysis
	tool, err := server.tools.Get("governance.analyze_usage")
	require.NoError(t, err)
	
	result, err := tool.Handler(ctx, map[string]interface{}{
		"time_range": "7 days",
	})
	require.NoError(t, err)
	
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "usage_summary")
	assert.Contains(t, resultMap, "top_consumers")
	
	// Test cost optimization
	tool, err = server.tools.Get("governance.optimize_costs")
	require.NoError(t, err)
	
	result, err = tool.Handler(ctx, map[string]interface{}{
		"focus_area": "data_retention",
	})
	require.NoError(t, err)
	
	resultMap, ok = result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "recommendations")
	assert.Contains(t, resultMap, "potential_savings")
	
	// Test compliance check
	tool, err = server.tools.Get("governance.check_compliance")
	require.NoError(t, err)
	
	result, err = tool.Handler(ctx, map[string]interface{}{
		"policy_type": "data_retention",
	})
	require.NoError(t, err)
	
	resultMap, ok = result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "compliance_status")
	assert.Contains(t, resultMap, "violations")
}

func testBulkOperationTools(t *testing.T, ctx context.Context, server *Server) {
	// Test bulk query execution
	tool, err := server.tools.Get("bulk_execute_queries")
	require.NoError(t, err)
	
	queries := []map[string]interface{}{
		{
			"name":  "query1",
			"query": "SELECT count(*) FROM Transaction",
		},
		{
			"name":  "query2",
			"query": "SELECT average(duration) FROM Transaction",
		},
	}
	
	result, err := tool.Handler(ctx, map[string]interface{}{
		"queries":  queries,
		"parallel": true,
	})
	require.NoError(t, err)
	
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "results")
	assert.Contains(t, resultMap, "summary")
	
	summary, ok := resultMap["summary"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(2), summary["total"])
}

// TestToolParameterValidation tests parameter validation for all tools
func TestToolParameterValidation(t *testing.T) {
	server := createTestServer(t)
	err := server.registerTools()
	require.NoError(t, err)
	
	ctx := context.Background()
	
	testCases := []struct {
		toolName      string
		invalidParams map[string]interface{}
		expectedError string
	}{
		{
			toolName:      "query_nrdb",
			invalidParams: map[string]interface{}{},
			expectedError: "required",
		},
		{
			toolName: "discovery.explore_attributes",
			invalidParams: map[string]interface{}{
				// Missing required event_type
			},
			expectedError: "required",
		},
		{
			toolName: "list_dashboards",
			invalidParams: map[string]interface{}{
				"limit": "not-a-number",
			},
			expectedError: "invalid",
		},
		{
			toolName: "analysis.detect_anomalies",
			invalidParams: map[string]interface{}{
				"metric": "duration",
				// Missing event_type
			},
			expectedError: "required",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.toolName, func(t *testing.T) {
			tool, err := server.tools.Get(tc.toolName)
			if err != nil {
				t.Skipf("Tool %s not found", tc.toolName)
			}
			
			_, err = tool.Handler(ctx, tc.invalidParams)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// TestToolChaining tests tools that work together in sequences
func TestToolChaining(t *testing.T) {
	server := createTestServer(t)
	err := server.registerTools()
	require.NoError(t, err)
	
	ctx := context.Background()
	
	// Chain: Discover -> Profile -> Analyze -> Create Dashboard
	
	// Step 1: Discover event types
	discoverTool, err := server.tools.Get("discovery.explore_event_types")
	require.NoError(t, err)
	
	discoverResult, err := discoverTool.Handler(ctx, map[string]interface{}{
		"time_range": "1 hour",
		"limit":      5,
	})
	require.NoError(t, err)
	
	discoverMap := discoverResult.(map[string]interface{})
	eventTypes := discoverMap["event_types"].([]interface{})
	require.NotEmpty(t, eventTypes)
	
	firstEvent := eventTypes[0].(map[string]interface{})
	eventType := firstEvent["name"].(string)
	
	// Step 2: Profile the event type
	profileTool, err := server.tools.Get("discovery.profile_data_completeness")
	if err == nil {
		profileResult, err := profileTool.Handler(ctx, map[string]interface{}{
			"event_types": []string{eventType},
		})
		assert.NoError(t, err)
		
		if profileResult != nil {
			profileMap := profileResult.(map[string]interface{})
			assert.Contains(t, profileMap, "profiles")
		}
	}
	
	// Step 3: Analyze the data
	analyzeTool, err := server.tools.Get("analysis.calculate_baseline")
	require.NoError(t, err)
	
	analyzeResult, err := analyzeTool.Handler(ctx, map[string]interface{}{
		"metric":     "count(*)",
		"event_type": eventType,
		"time_range": "1 hour",
	})
	require.NoError(t, err)
	
	analyzeMap := analyzeResult.(map[string]interface{})
	assert.Contains(t, analyzeMap, "baseline")
	
	// Step 4: Generate dashboard based on findings
	dashboardTool, err := server.tools.Get("generate_dashboard")
	require.NoError(t, err)
	
	dashboardResult, err := dashboardTool.Handler(ctx, map[string]interface{}{
		"template":   "discovery-based",
		"name":       fmt.Sprintf("Auto Dashboard for %s", eventType),
		"domain":     eventType,
	})
	require.NoError(t, err)
	
	dashboardMap := dashboardResult.(map[string]interface{})
	assert.Contains(t, dashboardMap, "dashboard")
}

// TestToolPerformance tests tool execution performance
func TestToolPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	server := createTestServer(t)
	err := server.registerTools()
	require.NoError(t, err)
	
	ctx := context.Background()
	
	// Test query performance
	tool, err := server.tools.Get("query_nrdb")
	require.NoError(t, err)
	
	start := time.Now()
	_, err = tool.Handler(ctx, map[string]interface{}{
		"query":   "SELECT count(*) FROM Transaction",
		"timeout": 5,
	})
	duration := time.Since(start)
	
	assert.NoError(t, err)
	assert.Less(t, duration, 1*time.Second, "Simple query should complete quickly")
	
	// Test discovery performance
	tool, err = server.tools.Get("discovery.explore_event_types")
	require.NoError(t, err)
	
	start = time.Now()
	_, err = tool.Handler(ctx, map[string]interface{}{
		"time_range": "1 hour",
		"limit":      10,
	})
	duration = time.Since(start)
	
	assert.NoError(t, err)
	assert.Less(t, duration, 2*time.Second, "Discovery should complete reasonably fast")
}
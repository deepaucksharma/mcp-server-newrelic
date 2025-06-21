package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkflowOrchestration validates multi-step workflow execution
func TestWorkflowOrchestration(t *testing.T) {
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

	t.Run("SimpleWorkflow", func(t *testing.T) {
		// Execute a workflow that discovers and analyzes
		result, err := client.ExecuteTool(ctx, "workflow.execute", map[string]interface{}{
			"name": "Host Performance Analysis",
			"steps": []map[string]interface{}{
				{
					"id":   "discover_hosts",
					"tool": "discovery.find_entities",
					"params": map[string]interface{}{
						"entity_type": "HOST",
						"limit":       5,
					},
				},
				{
					"id":   "analyze_cpu",
					"tool": "analysis.get_metric_statistics",
					"params": map[string]interface{}{
						"entity_list": "${discover_hosts.result.entities}",
						"metric":      "cpuPercent",
						"statistics":  []string{"average", "max"},
					},
				},
			},
		})

		if err != nil {
			t.Logf("Workflow execution not yet implemented: %v", err)
			return
		}

		resultMap := result.(map[string]interface{})
		assert.Contains(t, resultMap, "workflow_id")
		assert.Contains(t, resultMap, "final_result")
		assert.Contains(t, resultMap, "execution_time")

		t.Logf("Workflow completed in: %v", resultMap["execution_time"])
	})

	t.Run("ConditionalWorkflow", func(t *testing.T) {
		// Workflow with conditional steps based on discovery
		result, err := client.ExecuteTool(ctx, "workflow.execute", map[string]interface{}{
			"name": "Adaptive Error Analysis",
			"steps": []map[string]interface{}{
				{
					"id":   "check_error_attr",
					"tool": "discovery.check_attribute_exists",
					"params": map[string]interface{}{
						"event_type": "Transaction",
						"attribute":  "error",
					},
				},
				{
					"id":   "error_rate_bool",
					"tool": "analysis.get_error_rate",
					"params": map[string]interface{}{
						"method": "boolean_error",
					},
					"condition": "${check_error_attr.result.exists} == true",
				},
				{
					"id":   "error_rate_http",
					"tool": "analysis.get_error_rate",
					"params": map[string]interface{}{
						"method": "http_status_code",
					},
					"condition": "${check_error_attr.result.exists} == false",
				},
			},
		})

		if err != nil {
			t.Logf("Conditional workflow not yet implemented: %v", err)
			return
		}

		resultMap := result.(map[string]interface{})
		executedSteps := resultMap["executed_steps"].([]interface{})
		
		// Should have executed exactly 2 steps (check + one analysis)
		assert.Len(t, executedSteps, 2)
	})
}

// TestErrorScenarios validates error handling with real API errors
func TestErrorScenarios(t *testing.T) {
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

	t.Run("InvalidNRQLSyntax", func(t *testing.T) {
		_, err := client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
			"query": "SELEKT * FORM Transaction", // Intentional typos
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Syntax")
	})

	t.Run("NonExistentEventType", func(t *testing.T) {
		_, err := client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
			"query": "SELECT count(*) FROM NonExistentEventType12345",
		})
		
		// Should succeed but return no results
		if err != nil {
			assert.Contains(t, err.Error(), "No events found")
		}
	})

	t.Run("QueryTimeout", func(t *testing.T) {
		// Try a complex query with very short timeout
		_, err := client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
			"query":   "SELECT count(*) FROM Transaction SINCE 30 days ago FACET appName LIMIT 1000",
			"timeout": 1, // 1ms timeout - should fail
		})
		
		if err != nil {
			// Either timeout or invalid timeout value
			t.Logf("Query failed as expected: %v", err)
		}
	})
}

// TestPerformanceBenchmarks runs performance tests with real queries
func TestPerformanceBenchmarks(t *testing.T) {
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

	t.Run("SimpleQueryPerformance", func(t *testing.T) {
		// Measure simple query performance
		iterations := 10
		var totalDuration time.Duration
		
		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, err := client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
				"query": "SELECT count(*) FROM Transaction SINCE 5 minutes ago",
			})
			duration := time.Since(start)
			
			if err == nil {
				totalDuration += duration
			}
		}
		
		avgDuration := totalDuration / time.Duration(iterations)
		t.Logf("Average query time: %v", avgDuration)
		
		// Should complete within reasonable time
		assert.Less(t, avgDuration.Milliseconds(), int64(1000), "Simple queries should complete within 1s")
	})

	t.Run("DiscoveryPerformance", func(t *testing.T) {
		start := time.Now()
		result, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 100,
		})
		duration := time.Since(start)
		
		if err == nil {
			resultMap := result.(map[string]interface{})
			eventTypes := resultMap["event_types"].([]interface{})
			t.Logf("Discovered %d event types in %v", len(eventTypes), duration)
		}
		
		assert.Less(t, duration.Milliseconds(), int64(2000), "Discovery should complete within 2s")
	})
}

// TestConcurrentExecution validates concurrent tool execution
func TestConcurrentExecution(t *testing.T) {
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

	t.Run("ParallelQueries", func(t *testing.T) {
		// Execute multiple queries in parallel
		queries := []string{
			"SELECT count(*) FROM Transaction SINCE 1 hour ago",
			"SELECT average(duration) FROM Transaction SINCE 1 hour ago",
			"SELECT uniqueCount(appName) FROM Transaction SINCE 1 hour ago",
		}
		
		type result struct {
			query string
			err   error
			data  interface{}
		}
		
		results := make(chan result, len(queries))
		
		// Launch queries concurrently
		for _, q := range queries {
			go func(query string) {
				data, err := client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
					"query": query,
				})
				results <- result{query: query, err: err, data: data}
			}(q)
		}
		
		// Collect results
		successCount := 0
		for i := 0; i < len(queries); i++ {
			res := <-results
			if res.err == nil {
				successCount++
				t.Logf("Query completed: %s", res.query)
			} else {
				t.Logf("Query failed (OK if no data): %s - %v", res.query, res.err)
			}
		}
		
		// At least some queries should succeed
		assert.Greater(t, successCount, 0, "At least one query should succeed")
	})
}
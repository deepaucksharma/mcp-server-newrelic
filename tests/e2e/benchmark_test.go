package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

// BenchmarkResult holds performance metrics for a tool execution
type BenchmarkResult struct {
	ToolName      string
	AvgLatency    time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
	P50Latency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration
	Iterations    int
	ErrorRate     float64
	ThroughputRPS float64
}

// TestMCPPerformanceBenchmarks runs performance benchmarks against real New Relic API
func TestMCPPerformanceBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance benchmarks in short mode")
	}

	// Load test environment
	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	client := framework.NewMCPTestClient(primaryAccount)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	// Run benchmarks for different tools
	benchmarks := []struct {
		name   string
		tool   string
		params map[string]interface{}
		warmup int
		runs   int
	}{
		{
			name: "Discovery_EventTypes_Small",
			tool: "discovery.explore_event_types",
			params: map[string]interface{}{
				"limit": 10,
			},
			warmup: 2,
			runs:   10,
		},
		{
			name: "Discovery_EventTypes_Large",
			tool: "discovery.explore_event_types",
			params: map[string]interface{}{
				"limit": 100,
			},
			warmup: 2,
			runs:   10,
		},
		{
			name: "Discovery_Attributes",
			tool: "discovery.explore_attributes",
			params: map[string]interface{}{
				"event_type":  "NrdbQuery",
				"sample_size": 100,
			},
			warmup: 2,
			runs:   10,
		},
		{
			name: "Query_Simple",
			tool: "query_nrdb",
			params: map[string]interface{}{
				"query": "SELECT count(*) FROM NrdbQuery SINCE 1 hour ago",
			},
			warmup: 2,
			runs:   10,
		},
		{
			name: "Query_Complex",
			tool: "query_nrdb",
			params: map[string]interface{}{
				"query": "SELECT average(durationMs), max(durationMs), percentile(durationMs, 95) FROM NrdbQuery SINCE 1 hour ago FACET user LIMIT 10",
			},
			warmup: 2,
			runs:   10,
		},
		{
			name: "QueryBuilder_Simple",
			tool: "query_builder",
			params: map[string]interface{}{
				"event_type": "NrdbQuery",
				"select":     []string{"count(*)"},
				"since":      "1 hour ago",
			},
			warmup: 1,
			runs:   10,
		},
	}

	results := make([]BenchmarkResult, 0, len(benchmarks))

	for _, bench := range benchmarks {
		t.Run(bench.name, func(t *testing.T) {
			// Warmup runs
			for i := 0; i < bench.warmup; i++ {
				_, err := client.ExecuteTool(ctx, bench.tool, bench.params)
				if err != nil {
					t.Logf("Warmup %d failed: %v", i+1, err)
				}
			}

			// Actual benchmark runs
			latencies := make([]time.Duration, 0, bench.runs)
			errors := 0
			
			startTime := time.Now()
			
			for i := 0; i < bench.runs; i++ {
				runStart := time.Now()
				_, err := client.ExecuteTool(ctx, bench.tool, bench.params)
				latency := time.Since(runStart)
				
				if err != nil {
					errors++
					t.Logf("Run %d failed: %v", i+1, err)
				} else {
					latencies = append(latencies, latency)
				}
				
				// Small delay between requests to avoid rate limiting
				time.Sleep(100 * time.Millisecond)
			}
			
			totalTime := time.Since(startTime)

			if len(latencies) == 0 {
				t.Skipf("All runs failed for %s", bench.name)
				return
			}

			// Calculate statistics
			result := calculateBenchmarkStats(bench.tool, latencies, errors, bench.runs, totalTime)
			results = append(results, result)

			// Log results
			t.Logf("=== Benchmark Results for %s ===", bench.name)
			t.Logf("Iterations: %d (Errors: %d)", result.Iterations, errors)
			t.Logf("Error Rate: %.2f%%", result.ErrorRate*100)
			t.Logf("Throughput: %.2f req/s", result.ThroughputRPS)
			t.Logf("Latencies:")
			t.Logf("  Min:  %v", result.MinLatency)
			t.Logf("  P50:  %v", result.P50Latency)
			t.Logf("  P95:  %v", result.P95Latency)
			t.Logf("  P99:  %v", result.P99Latency)
			t.Logf("  Max:  %v", result.MaxLatency)
			t.Logf("  Avg:  %v", result.AvgLatency)
		})
	}

	// Generate summary report
	generateBenchmarkReport(t, results)
}

// TestDiscoveryWorkflowBenchmark benchmarks a complete discovery workflow
func TestDiscoveryWorkflowBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workflow benchmark in short mode")
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

	// Benchmark complete discovery workflow
	const runs = 5
	workflowTimes := make([]time.Duration, 0, runs)

	for i := 0; i < runs; i++ {
		workflowStart := time.Now()

		// Step 1: Discover event types
		eventTypesResult, err := client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"limit": 10,
		})
		require.NoError(t, err)

		// Parse to get first event type
		resultMap := eventTypesResult.(map[string]interface{})
		content := resultMap["content"].([]interface{})
		firstContent := content[0].(map[string]interface{})
		textResult := firstContent["text"].(string)
		
		var toolResult map[string]interface{}
		err = json.Unmarshal([]byte(textResult), &toolResult)
		require.NoError(t, err)
		
		eventTypes := toolResult["event_types"].([]interface{})
		require.NotEmpty(t, eventTypes)
		
		firstEventType := eventTypes[0].(map[string]interface{})
		eventTypeName := firstEventType["name"].(string)

		// Step 2: Discover attributes
		_, err = client.ExecuteTool(ctx, "discovery.explore_attributes", map[string]interface{}{
			"event_type":  eventTypeName,
			"sample_size": 100,
		})
		require.NoError(t, err)

		// Step 3: Build and execute query
		queryResult, err := client.ExecuteTool(ctx, "query_builder", map[string]interface{}{
			"event_type": eventTypeName,
			"select":     []string{"count(*)"},
			"since":      "1 hour ago",
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
		
		query := qToolResult["query"].(string)

		// Step 4: Execute the built query
		_, err = client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": query,
		})
		require.NoError(t, err)

		workflowTime := time.Since(workflowStart)
		workflowTimes = append(workflowTimes, workflowTime)

		t.Logf("Workflow run %d completed in %v", i+1, workflowTime)
		
		// Delay between runs
		time.Sleep(500 * time.Millisecond)
	}

	// Calculate workflow statistics
	avgWorkflow := calculateAverage(workflowTimes)
	minWorkflow := findMin(workflowTimes)
	maxWorkflow := findMax(workflowTimes)

	t.Logf("=== Discovery Workflow Benchmark ===")
	t.Logf("Runs: %d", runs)
	t.Logf("Average: %v", avgWorkflow)
	t.Logf("Min: %v", minWorkflow)
	t.Logf("Max: %v", maxWorkflow)
}

// Helper functions

func calculateBenchmarkStats(toolName string, latencies []time.Duration, errors, totalRuns int, totalTime time.Duration) BenchmarkResult {
	if len(latencies) == 0 {
		return BenchmarkResult{
			ToolName:   toolName,
			Iterations: totalRuns,
			ErrorRate:  1.0,
		}
	}

	// Sort latencies for percentile calculations
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)
	for i := 0; i < len(sortedLatencies); i++ {
		for j := i + 1; j < len(sortedLatencies); j++ {
			if sortedLatencies[i] > sortedLatencies[j] {
				sortedLatencies[i], sortedLatencies[j] = sortedLatencies[j], sortedLatencies[i]
			}
		}
	}

	return BenchmarkResult{
		ToolName:      toolName,
		AvgLatency:    calculateAverage(latencies),
		MinLatency:    sortedLatencies[0],
		MaxLatency:    sortedLatencies[len(sortedLatencies)-1],
		P50Latency:    percentile(sortedLatencies, 50),
		P95Latency:    percentile(sortedLatencies, 95),
		P99Latency:    percentile(sortedLatencies, 99),
		Iterations:    totalRuns,
		ErrorRate:     float64(errors) / float64(totalRuns),
		ThroughputRPS: float64(totalRuns) / totalTime.Seconds(),
	}
}

func calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

func findMin(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	min := durations[0]
	for _, d := range durations[1:] {
		if d < min {
			min = d
		}
	}
	return min
}

func findMax(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	max := durations[0]
	for _, d := range durations[1:] {
		if d > max {
			max = d
		}
	}
	return max
}

func percentile(sortedDurations []time.Duration, p int) time.Duration {
	if len(sortedDurations) == 0 {
		return 0
	}
	index := (len(sortedDurations) - 1) * p / 100
	return sortedDurations[index]
}

func generateBenchmarkReport(t *testing.T, results []BenchmarkResult) {
	t.Log("\n=== PERFORMANCE BENCHMARK SUMMARY ===")
	t.Log("Tool                          | Avg Latency | P95 Latency | Error Rate | Throughput")
	t.Log("------------------------------|-------------|-------------|------------|------------")
	
	for _, r := range results {
		t.Logf("%-30s| %-11v | %-11v | %6.2f%% | %7.2f rps",
			r.ToolName,
			r.AvgLatency.Round(time.Millisecond),
			r.P95Latency.Round(time.Millisecond),
			r.ErrorRate*100,
			r.ThroughputRPS,
		)
	}
	
	t.Log("\nRecommendations:")
	for _, r := range results {
		if r.P95Latency > 5*time.Second {
			t.Logf("- %s: Consider optimizing - P95 latency exceeds 5s", r.ToolName)
		}
		if r.ErrorRate > 0.05 {
			t.Logf("- %s: High error rate (%.1f%%) needs investigation", r.ToolName, r.ErrorRate*100)
		}
	}
}
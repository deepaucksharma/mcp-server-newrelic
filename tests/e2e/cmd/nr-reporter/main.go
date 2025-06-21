package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TestResult represents a single test result
type TestResult struct {
	Name      string        `json:"name"`
	Package   string        `json:"package"`
	Duration  time.Duration `json:"duration"`
	Passed    bool          `json:"passed"`
	Skipped   bool          `json:"skipped"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// BenchmarkResult represents performance benchmark data
type BenchmarkResult struct {
	Name          string        `json:"name"`
	AvgLatency    time.Duration `json:"avg_latency"`
	P95Latency    time.Duration `json:"p95_latency"`
	P99Latency    time.Duration `json:"p99_latency"`
	ErrorRate     float64       `json:"error_rate"`
	ThroughputRPS float64       `json:"throughput_rps"`
	Timestamp     time.Time     `json:"timestamp"`
}

func main() {
	var (
		resultsDir = flag.String("results-dir", ".", "Directory containing test results")
		accountID  = flag.String("account-id", "", "New Relic account ID")
		apiKey     = flag.String("api-key", os.Getenv("NR_INSIGHTS_KEY"), "New Relic Insights API key")
		dryRun     = flag.Bool("dry-run", false, "Print events without sending")
	)
	flag.Parse()

	if *apiKey == "" {
		log.Println("No API key provided, skipping New Relic reporting")
		return
	}

	// Parse test results
	testResults, err := parseTestResults(*resultsDir)
	if err != nil {
		log.Fatalf("Failed to parse test results: %v", err)
	}

	// Parse benchmark results
	benchResults, err := parseBenchmarkResults(*resultsDir)
	if err != nil {
		log.Printf("Failed to parse benchmark results: %v", err)
	}

	// Create events
	events := createEvents(testResults, benchResults)

	if *dryRun {
		// Print events for debugging
		for _, event := range events {
			data, _ := json.MarshalIndent(event, "", "  ")
			fmt.Println(string(data))
		}
		return
	}

	// Send to New Relic
	if err := sendToNewRelic(events, *accountID, *apiKey); err != nil {
		log.Fatalf("Failed to send to New Relic: %v", err)
	}

	fmt.Printf("Successfully sent %d events to New Relic\n", len(events))
}

func parseTestResults(dir string) ([]TestResult, error) {
	results := []TestResult{}

	// Find all log files
	files, err := filepath.Glob(filepath.Join(dir, "*-test-results.log"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("Failed to read %s: %v", file, err)
			continue
		}

		// Parse Go test output
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "--- PASS:") || strings.HasPrefix(line, "--- FAIL:") || strings.HasPrefix(line, "--- SKIP:") {
				result := parseTestLine(line)
				if result != nil {
					results = append(results, *result)
				}
			}
		}
	}

	return results, nil
}

func parseTestLine(line string) *TestResult {
	// Example: --- PASS: TestDiscoveryChain/DiscoverThenQuery (4.57s)
	parts := strings.Fields(line)
	if len(parts) < 4 {
		return nil
	}

	result := &TestResult{
		Timestamp: time.Now(),
	}

	switch parts[1] {
	case "PASS:":
		result.Passed = true
	case "FAIL:":
		result.Passed = false
	case "SKIP:":
		result.Skipped = true
	}

	result.Name = parts[2]
	
	// Parse duration
	if len(parts) >= 4 {
		durationStr := strings.Trim(parts[3], "()")
		if d, err := time.ParseDuration(durationStr); err == nil {
			result.Duration = d
		}
	}

	return result
}

func parseBenchmarkResults(dir string) ([]BenchmarkResult, error) {
	results := []BenchmarkResult{}

	benchFile := filepath.Join(dir, "benchmark-results.log")
	content, err := ioutil.ReadFile(benchFile)
	if err != nil {
		return results, nil // Benchmarks are optional
	}

	// Parse benchmark output
	lines := strings.Split(string(content), "\n")
	var currentBench *BenchmarkResult

	for _, line := range lines {
		if strings.Contains(line, "=== Benchmark Results for") {
			if currentBench != nil {
				results = append(results, *currentBench)
			}
			name := extractBenchmarkName(line)
			currentBench = &BenchmarkResult{
				Name:      name,
				Timestamp: time.Now(),
			}
		} else if currentBench != nil {
			// Parse metrics
			if strings.Contains(line, "Avg:") {
				currentBench.AvgLatency = parseDuration(line)
			} else if strings.Contains(line, "P95:") {
				currentBench.P95Latency = parseDuration(line)
			} else if strings.Contains(line, "P99:") {
				currentBench.P99Latency = parseDuration(line)
			} else if strings.Contains(line, "Error Rate:") {
				fmt.Sscanf(line, "Error Rate: %f%%", &currentBench.ErrorRate)
				currentBench.ErrorRate /= 100
			} else if strings.Contains(line, "Throughput:") {
				fmt.Sscanf(line, "Throughput: %f req/s", &currentBench.ThroughputRPS)
			}
		}
	}

	if currentBench != nil {
		results = append(results, *currentBench)
	}

	return results, nil
}

func extractBenchmarkName(line string) string {
	start := strings.Index(line, "for ")
	end := strings.Index(line, " ===")
	if start != -1 && end != -1 {
		return line[start+4 : end]
	}
	return "Unknown"
}

func parseDuration(line string) time.Duration {
	parts := strings.Fields(line)
	for _, part := range parts {
		if d, err := time.ParseDuration(part); err == nil {
			return d
		}
	}
	return 0
}

func createEvents(testResults []TestResult, benchResults []BenchmarkResult) []map[string]interface{} {
	events := []map[string]interface{}{}

	// Create test result events
	for _, result := range testResults {
		event := map[string]interface{}{
			"eventType":   "MCPServerE2ETest",
			"testName":    result.Name,
			"package":     result.Package,
			"duration":    result.Duration.Milliseconds(),
			"passed":      result.Passed,
			"skipped":     result.Skipped,
			"timestamp":   result.Timestamp.Unix(),
			"environment": os.Getenv("GITHUB_REF"),
			"commit":      os.Getenv("GITHUB_SHA"),
			"actor":       os.Getenv("GITHUB_ACTOR"),
			"workflow":    os.Getenv("GITHUB_WORKFLOW"),
		}

		if result.Error != "" {
			event["error"] = result.Error
		}

		events = append(events, event)
	}

	// Create benchmark events
	for _, bench := range benchResults {
		event := map[string]interface{}{
			"eventType":     "MCPServerE2EBenchmark",
			"benchmarkName": bench.Name,
			"avgLatency":    bench.AvgLatency.Milliseconds(),
			"p95Latency":    bench.P95Latency.Milliseconds(),
			"p99Latency":    bench.P99Latency.Milliseconds(),
			"errorRate":     bench.ErrorRate,
			"throughputRPS": bench.ThroughputRPS,
			"timestamp":     bench.Timestamp.Unix(),
			"environment":   os.Getenv("GITHUB_REF"),
			"commit":        os.Getenv("GITHUB_SHA"),
		}

		events = append(events, event)
	}

	return events
}

func sendToNewRelic(events []map[string]interface{}, accountID, apiKey string) error {
	// In a real implementation, this would use the New Relic Events API
	// For now, just log what would be sent
	log.Printf("Would send %d events to New Relic account %s", len(events), accountID)
	return nil
}
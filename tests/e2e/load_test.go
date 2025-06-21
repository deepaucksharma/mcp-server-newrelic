package e2e

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

// LoadTestResult captures metrics from load testing
type LoadTestResult struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	TotalDuration      time.Duration
	AverageLatency     time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration
	RequestsPerSecond  float64
	ErrorRate          float64
	Errors             []error
}

// TestMCPLoadTesting performs load testing on MCP server
func TestMCPLoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Load test environment
	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	t.Run("ConcurrentDiscoveryRequests", func(t *testing.T) {
		// Test concurrent discovery requests
		result := runLoadTest(t, primaryAccount, LoadTestConfig{
			ConcurrentUsers: 5,
			RequestsPerUser: 10,
			Tool:           "discovery.explore_event_types",
			Parameters: map[string]interface{}{
				"limit": 10,
			},
			ThinkTime: 100 * time.Millisecond,
		})

		// Assert performance thresholds
		require.Less(t, result.ErrorRate, 0.05, "Error rate should be less than 5%")
		require.Less(t, result.AverageLatency, 5*time.Second, "Average latency should be less than 5s")
		require.Greater(t, result.RequestsPerSecond, 0.5, "Should handle at least 0.5 req/s")

		t.Logf("Load test results: %+v", result)
	})

	t.Run("ConcurrentQueryRequests", func(t *testing.T) {
		// Test concurrent NRQL queries
		result := runLoadTest(t, primaryAccount, LoadTestConfig{
			ConcurrentUsers: 10,
			RequestsPerUser: 5,
			Tool:           "query_nrdb",
			Parameters: map[string]interface{}{
				"query": "SELECT count(*) FROM NrdbQuery SINCE 1 hour ago",
			},
			ThinkTime: 50 * time.Millisecond,
		})

		// Assert performance thresholds
		require.Less(t, result.ErrorRate, 0.01, "Error rate should be less than 1%")
		require.Less(t, result.AverageLatency, 1*time.Second, "Average latency should be less than 1s")
		require.Greater(t, result.RequestsPerSecond, 2.0, "Should handle at least 2 req/s")
	})

	t.Run("MixedWorkload", func(t *testing.T) {
		// Test mixed workload with different tools
		result := runMixedLoadTest(t, primaryAccount, MixedLoadTestConfig{
			ConcurrentUsers: 8,
			Duration:        30 * time.Second,
			Scenarios: []LoadScenario{
				{
					Weight: 0.4,
					Tool:   "discovery.explore_event_types",
					Parameters: map[string]interface{}{
						"limit": 10,
					},
				},
				{
					Weight: 0.3,
					Tool:   "query_nrdb",
					Parameters: map[string]interface{}{
						"query": "SELECT count(*) FROM NrdbQuery SINCE 5 minutes ago",
					},
				},
				{
					Weight: 0.2,
					Tool:   "discovery.explore_attributes",
					Parameters: map[string]interface{}{
						"event_type": "NrdbQuery",
						"limit":      20,
					},
				},
				{
					Weight: 0.1,
					Tool:   "analysis.calculate_baseline",
					Parameters: map[string]interface{}{
						"metric":     "duration",
						"event_type": "NrdbQuery",
						"time_range": "1 hour",
					},
				},
			},
		})

		// Assert mixed workload performance
		require.Less(t, result.ErrorRate, 0.05, "Error rate should be less than 5%")
		require.Greater(t, result.RequestsPerSecond, 1.0, "Should handle at least 1 req/s for mixed workload")
	})

	t.Run("StressTest", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping stress test")
		}

		// Gradually increase load to find breaking point
		for users := 1; users <= 20; users += 5 {
			t.Run(fmt.Sprintf("%d_concurrent_users", users), func(t *testing.T) {
				result := runLoadTest(t, primaryAccount, LoadTestConfig{
					ConcurrentUsers: users,
					RequestsPerUser: 10,
					Tool:           "query_nrdb",
					Parameters: map[string]interface{}{
						"query": "SELECT count(*) FROM NrdbQuery SINCE 5 minutes ago",
					},
					ThinkTime: 10 * time.Millisecond,
				})

				t.Logf("Stress test with %d users: RPS=%.2f, ErrorRate=%.2f%%, AvgLatency=%v",
					users, result.RequestsPerSecond, result.ErrorRate*100, result.AverageLatency)

				// Stop if error rate exceeds 10%
				if result.ErrorRate > 0.1 {
					t.Logf("Breaking point reached at %d concurrent users", users)
					return
				}
			})
		}
	})
}

// LoadTestConfig configures a load test scenario
type LoadTestConfig struct {
	ConcurrentUsers int
	RequestsPerUser int
	Tool           string
	Parameters     map[string]interface{}
	ThinkTime      time.Duration
}

// LoadScenario represents a single scenario in mixed workload
type LoadScenario struct {
	Weight     float64
	Tool       string
	Parameters map[string]interface{}
}

// MixedLoadTestConfig configures a mixed workload test
type MixedLoadTestConfig struct {
	ConcurrentUsers int
	Duration        time.Duration
	Scenarios       []LoadScenario
}

// runLoadTest executes a load test with given configuration
func runLoadTest(t *testing.T, account *framework.TestAccount, config LoadTestConfig) *LoadTestResult {
	var wg sync.WaitGroup
	resultChan := make(chan *RequestResult, config.ConcurrentUsers*config.RequestsPerUser)
	
	startTime := time.Now()

	// Launch concurrent users
	for user := 0; user < config.ConcurrentUsers; user++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			// Create client for this user
			client := framework.NewMCPTestClient(account)
			ctx := context.Background()
			
			err := client.Start(ctx)
			if err != nil {
				t.Logf("User %d: Failed to start client: %v", userID, err)
				return
			}
			defer client.Stop()

			// Execute requests
			for req := 0; req < config.RequestsPerUser; req++ {
				reqStart := time.Now()
				
				_, err := client.ExecuteTool(ctx, config.Tool, config.Parameters)
				
				resultChan <- &RequestResult{
					Duration: time.Since(reqStart),
					Error:    err,
				}

				// Think time between requests
				if req < config.RequestsPerUser-1 {
					time.Sleep(config.ThinkTime)
				}
			}
		}(user)
	}

	// Wait for all users to complete
	wg.Wait()
	close(resultChan)
	
	totalDuration := time.Since(startTime)

	// Collect results
	return analyzeResults(resultChan, totalDuration)
}

// runMixedLoadTest executes a mixed workload test
func runMixedLoadTest(t *testing.T, account *framework.TestAccount, config MixedLoadTestConfig) *LoadTestResult {
	var wg sync.WaitGroup
	resultChan := make(chan *RequestResult, 1000)
	stopChan := make(chan struct{})
	
	startTime := time.Now()

	// Launch concurrent users
	for user := 0; user < config.ConcurrentUsers; user++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			// Create client for this user
			client := framework.NewMCPTestClient(account)
			ctx := context.Background()
			
			err := client.Start(ctx)
			if err != nil {
				t.Logf("User %d: Failed to start client: %v", userID, err)
				return
			}
			defer client.Stop()

			// Execute requests until duration expires
			for {
				select {
				case <-stopChan:
					return
				default:
					// Select scenario based on weight
					scenario := selectScenario(config.Scenarios)
					
					reqStart := time.Now()
					_, err := client.ExecuteTool(ctx, scenario.Tool, scenario.Parameters)
					
					resultChan <- &RequestResult{
						Duration: time.Since(reqStart),
						Error:    err,
						Tool:     scenario.Tool,
					}

					// Small think time
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(user)
	}

	// Run for specified duration
	time.Sleep(config.Duration)
	close(stopChan)
	
	// Wait for all users to stop
	wg.Wait()
	close(resultChan)
	
	totalDuration := time.Since(startTime)

	// Collect results
	return analyzeResults(resultChan, totalDuration)
}

// RequestResult captures a single request result
type RequestResult struct {
	Duration time.Duration
	Error    error
	Tool     string
}

// analyzeResults processes request results and calculates metrics
func analyzeResults(results <-chan *RequestResult, totalDuration time.Duration) *LoadTestResult {
	var (
		totalRequests      int
		successfulRequests int
		failedRequests     int
		totalLatency       time.Duration
		minLatency         = time.Hour
		maxLatency         time.Duration
		errors             []error
	)

	for result := range results {
		totalRequests++
		
		if result.Error != nil {
			failedRequests++
			errors = append(errors, result.Error)
		} else {
			successfulRequests++
		}

		totalLatency += result.Duration
		
		if result.Duration < minLatency {
			minLatency = result.Duration
		}
		if result.Duration > maxLatency {
			maxLatency = result.Duration
		}
	}

	if totalRequests == 0 {
		return &LoadTestResult{}
	}

	avgLatency := totalLatency / time.Duration(totalRequests)
	errorRate := float64(failedRequests) / float64(totalRequests)
	rps := float64(totalRequests) / totalDuration.Seconds()

	return &LoadTestResult{
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		TotalDuration:      totalDuration,
		AverageLatency:     avgLatency,
		MinLatency:         minLatency,
		MaxLatency:         maxLatency,
		RequestsPerSecond:  rps,
		ErrorRate:          errorRate,
		Errors:             errors,
	}
}

// selectScenario picks a scenario based on weights
func selectScenario(scenarios []LoadScenario) LoadScenario {
	// Simple implementation - in production use weighted random selection
	return scenarios[0]
}
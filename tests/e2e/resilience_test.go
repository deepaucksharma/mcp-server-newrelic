package e2e

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPResilience tests the system's resilience to various failure scenarios
func TestMCPResilience(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resilience test in short mode")
	}

	// Load test environment
	err := godotenv.Load("../../.env.test")
	require.NoError(t, err)

	accounts := framework.LoadTestAccounts()
	primaryAccount, ok := accounts["primary"]
	require.True(t, ok)

	t.Run("HandlesTimeouts", func(t *testing.T) {
		client := framework.NewMCPTestClient(primaryAccount)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := client.Start(ctx)
		require.NoError(t, err)
		defer client.Stop()

		// Execute a query that typically takes longer than timeout
		_, err = client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"time_range": "30 days",
			"limit":      1000,
		})

		// Should get context deadline exceeded
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("HandlesLargePayloads", func(t *testing.T) {
		client := framework.NewMCPTestClient(primaryAccount)
		ctx := context.Background()

		err := client.Start(ctx)
		require.NoError(t, err)
		defer client.Stop()

		// Try to query with very large result set
		result, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": "SELECT * FROM NrdbQuery SINCE 24 hours ago LIMIT MAX",
		})

		// Should handle gracefully (either succeed or return appropriate error)
		if err != nil {
			assert.Contains(t, err.Error(), "result too large")
		} else {
			assert.NotNil(t, result)
		}
	})

	t.Run("HandlesInvalidProtocolMessages", func(t *testing.T) {
		t.Skip("Requires framework.StartMCPServer implementation")
	})

	t.Run("HandlesConnectionInterruption", func(t *testing.T) {
		client := framework.NewMCPTestClient(primaryAccount)
		ctx := context.Background()

		err := client.Start(ctx)
		require.NoError(t, err)

		// Execute a successful request first
		result, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": "SELECT count(*) FROM NrdbQuery SINCE 1 minute ago",
		})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Simulate connection interruption by stopping the client
		client.Stop()

		// Try to execute another request
		_, err = client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": "SELECT count(*) FROM NrdbQuery SINCE 1 minute ago",
		})
		assert.Error(t, err)
	})

	t.Run("HandlesRateLimiting", func(t *testing.T) {
		client := framework.NewMCPTestClient(primaryAccount)
		ctx := context.Background()

		err := client.Start(ctx)
		require.NoError(t, err)
		defer client.Stop()

		// Send many requests rapidly
		errors := 0
		rateLimitErrors := 0
		
		for i := 0; i < 50; i++ {
			_, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
				"query": fmt.Sprintf("SELECT count(*) FROM NrdbQuery WHERE query = 'test%d' SINCE 1 minute ago", i),
			})
			
			if err != nil {
				errors++
				if contains(err.Error(), "rate limit") || contains(err.Error(), "too many requests") {
					rateLimitErrors++
				}
			}
		}

		// Log results
		t.Logf("Total errors: %d, Rate limit errors: %d", errors, rateLimitErrors)
		
		// Should handle rate limiting gracefully
		if errors > 0 {
			assert.Greater(t, rateLimitErrors, 0, "Some errors should be rate limit errors")
		}
	})

	t.Run("RecoverFromPanic", func(t *testing.T) {
		client := framework.NewMCPTestClient(primaryAccount)
		ctx := context.Background()

		err := client.Start(ctx)
		require.NoError(t, err)
		defer client.Stop()

		// Try to trigger edge cases that might cause panics
		edgeCases := []map[string]interface{}{
			{
				"query": "SELECT * FROM '' SINCE 1 hour ago", // Empty event type
			},
			{
				"query": "SELECT ${injection} FROM Transaction", // Potential injection
			},
			{
				"query": string(make([]byte, 10000)), // Very long query
			},
		}

		for i, params := range edgeCases {
			t.Run(fmt.Sprintf("EdgeCase%d", i), func(t *testing.T) {
				// Should not panic, but return error
				_, err := client.ExecuteTool(ctx, "query_nrdb", params)
				assert.Error(t, err)
				
				// Server should still be responsive
				_, err = client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
					"query": "SELECT count(*) FROM NrdbQuery SINCE 1 minute ago",
				})
				assert.NoError(t, err)
			})
		}
	})

	t.Run("HandlesNetworkLatency", func(t *testing.T) {
		// Create a proxy that adds latency
		proxy := createLatencyProxy(t, "localhost:8080", 500*time.Millisecond)
		defer proxy.Close()

		// Connect through proxy
		// Note: This would require modifying the client to support proxy
		t.Skip("Proxy support not implemented in test client")
	})

	t.Run("HandlesConcurrentRequests", func(t *testing.T) {
		client := framework.NewMCPTestClient(primaryAccount)
		ctx := context.Background()

		err := client.Start(ctx)
		require.NoError(t, err)
		defer client.Stop()

		// Send multiple requests concurrently
		results := make(chan error, 10)
		
		for i := 0; i < 10; i++ {
			go func(id int) {
				_, err := client.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
					"query": fmt.Sprintf("SELECT count(*) as c%d FROM NrdbQuery SINCE 1 minute ago", id),
				})
				results <- err
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < 10; i++ {
			err := <-results
			if err == nil {
				successCount++
			}
		}

		// Most requests should succeed
		assert.Greater(t, successCount, 5, "At least half of concurrent requests should succeed")
	})
}

// createLatencyProxy creates a TCP proxy that adds latency to connections
func createLatencyProxy(t *testing.T, target string, latency time.Duration) net.Listener {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				
				// Add latency
				time.Sleep(latency)
				
				// Forward to target
				targetConn, err := net.Dial("tcp", target)
				if err != nil {
					return
				}
				defer targetConn.Close()

				// Proxy data
				go func() {
					_, _ = conn.(*net.TCPConn).ReadFrom(targetConn)
				}()
				_, _ = targetConn.(*net.TCPConn).ReadFrom(conn)
			}(conn)
		}
	}()

	return listener
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
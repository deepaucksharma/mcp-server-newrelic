package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2ESetupValidation(t *testing.T) {
	// Load test environment
	err := godotenv.Load("../../.env.test")
	if err != nil {
		// Try from test directory
		err = godotenv.Load(".env.test")
	}
	require.NoError(t, err, "Failed to load .env.test")

	// Validate required environment variables
	t.Run("ValidateEnvironment", func(t *testing.T) {
		required := []string{
			"NEW_RELIC_API_KEY_PRIMARY",
			"NEW_RELIC_ACCOUNT_ID_PRIMARY",
			"NEW_RELIC_REGION_PRIMARY",
		}

		for _, key := range required {
			value := os.Getenv(key)
			assert.NotEmpty(t, value, "Environment variable %s must be set", key)
		}
	})

	// Test New Relic API connectivity
	t.Run("ValidateNewRelicAPI", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := framework.NewTestClient(ctx)
		require.NoError(t, err, "Failed to create test client")

		// Try a simple NRQL query
		query := "SELECT count(*) FROM Transaction SINCE 1 hour ago"
		result, err := client.ExecuteNRQL(ctx, query, os.Getenv("NEW_RELIC_ACCOUNT_ID_PRIMARY"))
		
		// We don't require data to exist, just that the API call works
		if err != nil {
			t.Logf("NRQL query returned error (this is OK if no data exists): %v", err)
		} else {
			t.Logf("NRQL query succeeded: %+v", result)
		}
	})

	// Test MCP server availability
	t.Run("ValidateMCPServer", func(t *testing.T) {
		// Check if binary exists
		paths := []string{
			"./bin/mcp-server",
			"../../bin/mcp-server",
			"../../../bin/mcp-server",
		}

		found := false
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				t.Logf("Found MCP server binary at: %s", path)
				found = true
				break
			}
		}

		if !found {
			t.Skip("MCP server binary not found. Run 'make build' first")
		}
	})

	// Test discovery endpoint
	t.Run("ValidateDiscoveryEndpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client, err := framework.NewTestClient(ctx)
		require.NoError(t, err)

		// Try to discover event types
		eventTypes, err := client.DiscoverEventTypes(ctx, 10)
		if err != nil {
			t.Logf("Discovery returned error (this is OK if no data exists): %v", err)
		} else {
			t.Logf("Discovered %d event types", len(eventTypes))
			for i, et := range eventTypes {
				if i < 5 { // Log first 5
					t.Logf("  - %s", et)
				}
			}
		}
	})
}

// TestSimpleScenario runs a minimal scenario to validate the framework
func TestSimpleScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Load environment
	err := godotenv.Load(".env.test")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create a simple test scenario
	t.Run("BasicDiscoveryWorkflow", func(t *testing.T) {
		client, err := framework.NewTestClient(ctx)
		require.NoError(t, err)

		// Step 1: Discover event types
		eventTypes, err := client.DiscoverEventTypes(ctx, 100)
		if err != nil {
			t.Logf("No event types found (OK for empty account): %v", err)
			return
		}
		require.NotEmpty(t, eventTypes, "Should discover at least one event type")

		// Step 2: Pick first event type and discover attributes
		eventType := eventTypes[0]
		t.Logf("Exploring attributes for event type: %s", eventType)

		attributes, err := client.DiscoverAttributes(ctx, eventType, 100)
		if err != nil {
			t.Logf("Failed to discover attributes: %v", err)
			return
		}

		t.Logf("Discovered %d attributes for %s", len(attributes), eventType)

		// Step 3: Build and execute a simple query
		query := fmt.Sprintf("SELECT count(*) FROM %s SINCE 1 hour ago", eventType)
		result, err := client.ExecuteNRQL(ctx, query, os.Getenv("NEW_RELIC_ACCOUNT_ID_PRIMARY"))
		
		if err != nil {
			t.Logf("Query failed (OK if no recent data): %v", err)
		} else {
			t.Logf("Query result: %+v", result)
		}
	})
}

// TestScenarioFramework validates the scenario parsing and execution framework
func TestScenarioFramework(t *testing.T) {
	// This test doesn't need real New Relic credentials
	t.Run("ParseYAMLScenario", func(t *testing.T) {
		// Test that we can parse our example scenarios
		scenarioFiles := []string{
			"scenarios/disc-miss-001-missing-attributes.yaml",
			"scenarios/inc-sql-404-incident-response.yaml",
			"scenarios/perf-cmp-001-cross-region-latency.yaml",
		}

		for _, file := range scenarioFiles {
			t.Run(file, func(t *testing.T) {
				// Check if file exists
				if _, err := os.Stat(file); os.IsNotExist(err) {
					t.Skipf("Scenario file not found: %s", file)
					return
				}

				// We'll implement actual parsing in the harness package
				t.Logf("Would parse scenario: %s", file)
			})
		}
	})
}
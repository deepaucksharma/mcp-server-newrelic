package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/config"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/interface/mcp"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/state"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration
	cfg := config.LoadConfig()

	fmt.Println("=== End-to-End Test with Real NRDB ===")
	fmt.Printf("Account ID: %s\n", cfg.NewRelic.AccountID)
	fmt.Printf("Region: %s\n", cfg.NewRelic.Region)

	// Create New Relic client
	nrClient, err := newrelic.NewClient(newrelic.Config{
		APIKey:    cfg.NewRelic.APIKey,
		AccountID: cfg.NewRelic.AccountID,
		Region:    cfg.NewRelic.Region,
	})
	if err != nil {
		log.Fatalf("Failed to create New Relic client: %v", err)
	}

	// Create discovery engine
	discoveryEngine, err := discovery.NewEngine(cfg.Discovery, nrClient)
	if err != nil {
		log.Fatalf("Failed to create discovery engine: %v", err)
	}

	// Create state manager
	stateManager := state.NewManager(cfg.State)

	// Create MCP server
	server := mcp.NewServer(cfg.MCP, discoveryEngine, stateManager, nrClient)

	// Use a longer timeout for tests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: List Dashboards
	fmt.Println("\n1. Testing list_dashboards...")
	result, err := server.ExecuteTool(ctx, "list_dashboards", map[string]interface{}{})
	if err != nil {
		fmt.Printf("✗ Failed: %v\n", err)
	} else {
		data, _ := json.MarshalIndent(result, "  ", "  ")
		fmt.Printf("✓ Success: %s\n", string(data))
	}

	// Test 2: Query NRDB
	fmt.Println("\n2. Testing query_nrdb...")
	result, err = server.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
		"query": "SELECT count(*) FROM Transaction SINCE 1 hour ago",
	})
	if err != nil {
		fmt.Printf("✗ Failed: %v\n", err)
	} else {
		data, _ := json.MarshalIndent(result, "  ", "  ")
		fmt.Printf("✓ Success: %s\n", string(data))
	}

	// Test 3: List Schemas
	fmt.Println("\n3. Testing discovery.list_schemas...")
	result, err = server.ExecuteTool(ctx, "discovery.list_schemas", map[string]interface{}{})
	if err != nil {
		fmt.Printf("✗ Failed: %v\n", err)
	} else {
		schemas := result.(map[string]interface{})["schemas"].([]interface{})
		fmt.Printf("✓ Found %d schemas\n", len(schemas))
		// Show first 5
		for i, s := range schemas {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(schemas)-5)
				break
			}
			schema := s.(map[string]interface{})
			fmt.Printf("  - %s (%v samples)\n", schema["name"], schema["sample_count"])
		}
	}

	// Test 4: List Alerts
	fmt.Println("\n4. Testing list_alerts...")
	result, err = server.ExecuteTool(ctx, "list_alerts", map[string]interface{}{})
	if err != nil {
		fmt.Printf("✗ Failed: %v\n", err)
	} else {
		alerts := result.(map[string]interface{})
		fmt.Printf("✓ Found %v alerts\n", alerts["total"])
	}

	// Test 5: Query Builder
	fmt.Println("\n5. Testing query_builder...")
	result, err = server.ExecuteTool(ctx, "query_builder", map[string]interface{}{
		"event_type": "Transaction",
		"select":     []string{"count(*)", "average(duration)"},
		"where":      "appName IS NOT NULL",
		"since":      "1 hour ago",
	})
	if err != nil {
		fmt.Printf("✗ Failed: %v\n", err)
	} else {
		queryResult := result.(map[string]interface{})
		fmt.Printf("✓ Built query: %s\n", queryResult["query"])
		
		// Execute the built query
		execResult, err := server.ExecuteTool(ctx, "query_nrdb", map[string]interface{}{
			"query": queryResult["query"],
		})
		if err != nil {
			fmt.Printf("  ✗ Execution failed: %v\n", err)
		} else {
			data, _ := json.MarshalIndent(execResult, "    ", "  ")
			fmt.Printf("  ✓ Execution result: %s\n", string(data))
		}
	}

	fmt.Println("\n=== All Tests Completed ===")
}
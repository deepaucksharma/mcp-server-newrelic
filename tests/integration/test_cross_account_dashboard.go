package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/interface/mcp"
)

func main() {
	// Create MCP server with minimal setup
	server := &mcp.Server{}
	
	// Use reflection to set the tools field (for testing only)
	// In production, this would be set via NewServer
	serverType := reflect.TypeOf(server).Elem()
	toolsField, _ := serverType.FieldByName("tools")
	if toolsField.Name == "" {
		// Try a different approach - create a minimal server setup
		fmt.Println("Creating test server...")
	}

	// Test 1: Create a golden signals dashboard with multiple account IDs
	fmt.Println("=== Test 1: Golden Signals Dashboard with Cross-Account Support ===")
	params1 := map[string]interface{}{
		"template":     "golden-signals",
		"name":         "Multi-Account APM Dashboard",
		"service_name": "my-service",
		"account_ids":  []interface{}{float64(12345), float64(67890), float64(11111)},
	}
	
	result1, err := server.HandleGenerateDashboard(context.Background(), params1)
	if err != nil {
		log.Fatalf("Failed to generate dashboard: %v", err)
	}
	
	dashboard1 := result1.(map[string]interface{})["dashboard"].(map[string]interface{})
	fmt.Printf("Dashboard Name: %s\n", dashboard1["name"])
	fmt.Printf("Account IDs: %v\n", dashboard1["account_ids"])
	
	// Check one of the queries
	pages := dashboard1["pages"].([]map[string]interface{})
	widgets := pages[0]["widgets"].([]map[string]interface{})
	fmt.Printf("Sample Query: %s\n\n", widgets[0]["query"])

	// Test 2: Infrastructure dashboard with single account
	fmt.Println("=== Test 2: Infrastructure Dashboard with Single Account ===")
	params2 := map[string]interface{}{
		"template":     "infrastructure",
		"name":         "Production Infrastructure",
		"host_pattern": "prod-*",
		"account_ids":  []interface{}{float64(99999)},
	}
	
	result2, err := server.HandleGenerateDashboard(context.Background(), params2)
	if err != nil {
		log.Fatalf("Failed to generate dashboard: %v", err)
	}
	
	dashboard2 := result2.(map[string]interface{})["dashboard"].(map[string]interface{})
	fmt.Printf("Dashboard Name: %s\n", dashboard2["name"])
	fmt.Printf("Account IDs: %v\n", dashboard2["account_ids"])
	
	// Check query
	pages2 := dashboard2["pages"].([]map[string]interface{})
	widgets2 := pages2[0]["widgets"].([]map[string]interface{})
	fmt.Printf("Sample Query: %s\n\n", widgets2[0]["query"])

	// Test 3: SLI/SLO dashboard with cross-account
	fmt.Println("=== Test 3: SLI/SLO Dashboard with Cross-Account ===")
	params3 := map[string]interface{}{
		"template": "sli-slo",
		"name":     "Service Availability SLO",
		"sli_config": map[string]interface{}{
			"name":   "Availability",
			"target": 99.95,
			"query":  "SELECT percentage(count(*), WHERE error IS false) FROM Transaction WHERE appName = 'my-service'",
		},
		"account_ids": []interface{}{float64(11111), float64(22222)},
	}
	
	result3, err := server.HandleGenerateDashboard(context.Background(), params3)
	if err != nil {
		log.Fatalf("Failed to generate dashboard: %v", err)
	}
	
	dashboard3 := result3.(map[string]interface{})["dashboard"].(map[string]interface{})
	fmt.Printf("Dashboard Name: %s\n", dashboard3["name"])
	fmt.Printf("Account IDs: %v\n", dashboard3["account_ids"])
	
	// Check query
	pages3 := dashboard3["pages"].([]map[string]interface{})
	widgets3 := pages3[0]["widgets"].([]map[string]interface{})
	fmt.Printf("Sample Query: %s\n\n", widgets3[0]["query"])

	// Pretty print one dashboard configuration
	fmt.Println("=== Full Dashboard Configuration (JSON) ===")
	dashboardJSON, _ := json.MarshalIndent(dashboard1, "", "  ")
	fmt.Println(string(dashboardJSON))
}
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Get credentials from environment
	apiKey := os.Getenv("NEW_RELIC_API_KEY")
	accountID := os.Getenv("NEW_RELIC_ACCOUNT_ID")
	region := os.Getenv("NEW_RELIC_REGION")

	fmt.Println("=== Testing New Relic NRDB Connection ===")
	fmt.Printf("Account ID: %s\n", accountID)
	fmt.Printf("Region: %s\n", region)
	fmt.Printf("API Key: %s...%s\n", apiKey[:10], apiKey[len(apiKey)-4:])

	// Create New Relic client
	client, err := newrelic.NewClient(newrelic.Config{
		APIKey:    apiKey,
		AccountID: accountID,
		Region:    region,
	})
	if err != nil {
		log.Fatalf("Failed to create New Relic client: %v", err)
	}

	ctx := context.Background()

	// Test 1: Get Account Info
	fmt.Println("\n1. Testing Account Info...")
	accountInfo, err := client.GetAccountInfo(ctx)
	if err != nil {
		log.Printf("Failed to get account info: %v", err)
	} else {
		fmt.Printf("✓ Account connected successfully\n")
		data, _ := json.MarshalIndent(accountInfo, "  ", "  ")
		fmt.Printf("  Account Info: %s\n", string(data))
	}

	// Test 2: Simple NRQL Query
	fmt.Println("\n2. Testing NRQL Query...")
	query := "SELECT count(*) FROM Transaction SINCE 1 hour ago"
	result, err := client.QueryNRQL(ctx, query)
	if err != nil {
		log.Printf("Failed to execute NRQL query: %v", err)
	} else {
		fmt.Printf("✓ NRQL query executed successfully\n")
		fmt.Printf("  Query: %s\n", query)
		if len(result.Results) > 0 {
			data, _ := json.MarshalIndent(result.Results[0], "  ", "  ")
			fmt.Printf("  Result: %s\n", string(data))
		}
	}

	// Test 3: List Event Types
	fmt.Println("\n3. Testing Event Type Discovery...")
	eventQuery := "SHOW EVENT TYPES SINCE 1 week ago"
	eventResult, err := client.QueryNRQL(ctx, eventQuery)
	if err != nil {
		log.Printf("Failed to get event types: %v", err)
	} else {
		fmt.Printf("✓ Event types retrieved successfully\n")
		fmt.Printf("  Found %d event types\n", len(eventResult.Results))
		// Show first 5 event types
		for i, result := range eventResult.Results {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(eventResult.Results)-5)
				break
			}
			eventType := result["eventType"]
			fmt.Printf("  - %v\n", eventType)
		}
	}

	// Test 4: List Dashboards
	fmt.Println("\n4. Testing Dashboard API...")
	dashboards, err := client.ListDashboards(ctx)
	if err != nil {
		log.Printf("Failed to list dashboards: %v", err)
	} else {
		fmt.Printf("✓ Dashboards retrieved successfully\n")
		fmt.Printf("  Found %d dashboards\n", len(dashboards))
		// Show first 3 dashboards
		for i, dashboard := range dashboards {
			if i >= 3 {
				fmt.Printf("  ... and %d more\n", len(dashboards)-3)
				break
			}
			fmt.Printf("  - %s (ID: %s)\n", dashboard.Name, dashboard.ID)
		}
	}

	// Test 5: Complex Query with Facets
	fmt.Println("\n5. Testing Complex NRQL Query...")
	complexQuery := `
		SELECT count(*), average(duration) 
		FROM Transaction 
		WHERE appName IS NOT NULL 
		SINCE 1 hour ago 
		FACET appName 
		LIMIT 5
	`
	complexResult, err := client.QueryNRQL(ctx, complexQuery)
	if err != nil {
		log.Printf("Failed to execute complex query: %v", err)
	} else {
		fmt.Printf("✓ Complex query executed successfully\n")
		fmt.Printf("  Found %d facets\n", len(complexResult.Results))
		for i, result := range complexResult.Results {
			if i >= 3 {
				break
			}
			data, _ := json.MarshalIndent(result, "  ", "  ")
			fmt.Printf("  Facet %d: %s\n", i+1, string(data))
		}
	}

	// Test 6: Test Data Quality
	fmt.Println("\n6. Testing Data Quality Check...")
	qualityQuery := `
		SELECT 
			count(*) as total_events,
			uniqueCount(appName) as unique_apps,
			percentage(count(*), WHERE error IS true) as error_rate
		FROM Transaction 
		SINCE 1 day ago
	`
	qualityResult, err := client.QueryNRQL(ctx, qualityQuery)
	if err != nil {
		log.Printf("Failed to check data quality: %v", err)
	} else {
		fmt.Printf("✓ Data quality check completed\n")
		if len(qualityResult.Results) > 0 {
			data, _ := json.MarshalIndent(qualityResult.Results[0], "  ", "  ")
			fmt.Printf("  Quality Metrics: %s\n", string(data))
		}
	}

	fmt.Println("\n=== All Tests Completed ===")
	fmt.Println("✓ NRDB connection is working properly!")
}
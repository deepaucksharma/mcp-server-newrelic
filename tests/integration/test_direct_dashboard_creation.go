package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	fmt.Println("=== Direct Dashboard Creation Test ===")

	// Create New Relic client
	client, err := newrelic.NewClient(newrelic.Config{
		APIKey:    os.Getenv("NEW_RELIC_API_KEY"),
		AccountID: os.Getenv("NEW_RELIC_ACCOUNT_ID"),
		Region:    os.Getenv("NEW_RELIC_REGION"),
	})
	if err != nil {
		log.Fatalf("Failed to create New Relic client: %v", err)
	}

	// Create a dashboard
	dashboard := newrelic.Dashboard{
		Name:        "Test Dashboard Direct Creation",
		Description: "Testing direct dashboard creation via API",
		Permissions: "PUBLIC_READ_WRITE",
		Pages: []newrelic.DashboardPage{
			{
				Name: "Test Page",
				Widgets: []newrelic.DashboardWidget{
					{
						Title: "Transaction Count",
						Type:  "line",
						Query: "SELECT count(*) FROM Transaction SINCE 1 hour ago TIMESERIES",
						Configuration: map[string]interface{}{
							"row":    1,
							"column": 1,
							"width":  12,
							"height": 3,
						},
					},
				},
			},
		},
	}

	// Use a long timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("Creating dashboard...")
	start := time.Now()
	
	created, err := client.CreateDashboard(ctx, dashboard)
	
	duration := time.Since(start)
	fmt.Printf("Request completed in %v\n", duration)

	if err != nil {
		fmt.Printf("❌ Failed to create dashboard: %v\n", err)
		return
	}

	fmt.Println("✅ Dashboard created successfully!")
	fmt.Printf("Dashboard ID: %s\n", created.ID)
	fmt.Printf("Dashboard Name: %s\n", created.Name)
	fmt.Printf("Dashboard URL: https://one.newrelic.com/dashboards/%s\n", created.ID)
	
	// Pretty print the created dashboard
	data, _ := json.MarshalIndent(created, "", "  ")
	fmt.Printf("\nCreated Dashboard:\n%s\n", string(data))
}
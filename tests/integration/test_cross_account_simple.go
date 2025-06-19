package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Simplified test to demonstrate cross-account NRQL query generation

func generateAccountClause(accountIDs []int) string {
	if len(accountIDs) == 0 {
		return ""
	}
	
	accountIDStrs := make([]string, len(accountIDs))
	for i, id := range accountIDs {
		accountIDStrs[i] = fmt.Sprintf("%d", id)
	}
	return fmt.Sprintf(" WITH accountIds = [%s]", strings.Join(accountIDStrs, ", "))
}

func generateCrossAccountQuery(baseQuery string, accountIDs []int) string {
	return baseQuery + generateAccountClause(accountIDs)
}

func main() {
	fmt.Println("=== Cross-Account Dashboard Query Generation Demo ===\n")

	// Test cases
	testCases := []struct {
		name       string
		baseQuery  string
		accountIDs []int
	}{
		{
			name:       "Single Account Query",
			baseQuery:  "SELECT average(duration) FROM Transaction WHERE appName = 'my-service' TIMESERIES",
			accountIDs: []int{12345},
		},
		{
			name:       "Multiple Account Query",
			baseQuery:  "SELECT count(*) FROM SystemSample WHERE hostname LIKE 'prod-%' FACET hostname",
			accountIDs: []int{11111, 22222, 33333},
		},
		{
			name:       "No Account IDs (Default Behavior)",
			baseQuery:  "SELECT rate(count(*), 1 minute) FROM Transaction TIMESERIES",
			accountIDs: []int{},
		},
		{
			name:       "Complex Query with Multiple Accounts",
			baseQuery:  "SELECT percentage(count(*), WHERE error IS false) as 'Success Rate' FROM Transaction WHERE appName IN ('service-a', 'service-b') TIMESERIES",
			accountIDs: []int{99999, 88888},
		},
	}

	// Generate and display queries
	for _, tc := range testCases {
		fmt.Printf("Test: %s\n", tc.name)
		fmt.Printf("Account IDs: %v\n", tc.accountIDs)
		
		query := generateCrossAccountQuery(tc.baseQuery, tc.accountIDs)
		fmt.Printf("Generated Query:\n%s\n\n", query)
	}

	// Example dashboard configuration with cross-account support
	fmt.Println("=== Example Dashboard Configuration ===")
	dashboard := map[string]interface{}{
		"name":        "Multi-Account Infrastructure Dashboard",
		"description": "Dashboard monitoring infrastructure across multiple accounts",
		"account_ids": []int{12345, 67890, 11111},
		"pages": []map[string]interface{}{
			{
				"name": "System Metrics",
				"widgets": []map[string]interface{}{
					{
						"title":  "CPU Usage Across Accounts",
						"type":   "line",
						"query":  generateCrossAccountQuery("SELECT average(cpuPercent) FROM SystemSample FACET hostname TIMESERIES", []int{12345, 67890, 11111}),
						"row":    1,
						"column": 1,
						"width":  12,
						"height": 3,
					},
					{
						"title":  "Memory Usage Across Accounts",
						"type":   "line",
						"query":  generateCrossAccountQuery("SELECT average(memoryUsedPercent) FROM SystemSample FACET hostname TIMESERIES", []int{12345, 67890, 11111}),
						"row":    4,
						"column": 1,
						"width":  12,
						"height": 3,
					},
				},
			},
		},
		"metadata": map[string]interface{}{
			"template":     "infrastructure",
			"created_with": "mcp-server-newrelic",
			"features":     []string{"cross-account-queries", "auto-generated"},
		},
	}

	// Pretty print the dashboard
	dashboardJSON, _ := json.MarshalIndent(dashboard, "", "  ")
	fmt.Println(string(dashboardJSON))

	// Show how the NerdGraph mutation would look
	fmt.Println("\n=== NerdGraph Dashboard Creation ===")
	fmt.Println("The dashboard would be created using a GraphQL mutation like:")
	fmt.Println(`
mutation {
  dashboardCreate(
    accountId: 12345
    dashboard: {
      name: "Multi-Account Infrastructure Dashboard"
      description: "Dashboard monitoring infrastructure across multiple accounts"
      permissions: PUBLIC_READ_WRITE
      pages: [
        {
          name: "System Metrics"
          widgets: [
            {
              title: "CPU Usage Across Accounts"
              configuration: {
                line: {
                  nrqlQueries: [
                    {
                      accountId: 12345
                      query: "SELECT average(cpuPercent) FROM SystemSample FACET hostname TIMESERIES WITH accountIds = [12345, 67890, 11111]"
                    }
                  ]
                }
              }
              layout: { column: 1, row: 1, width: 12, height: 3 }
            }
          ]
        }
      ]
    }
  ) {
    entityResult {
      guid
      name
    }
    errors {
      description
      type
    }
  }
}`)
}
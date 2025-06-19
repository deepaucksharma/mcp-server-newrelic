package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardTools(t *testing.T) {
	s := &Server{
		tools: NewToolRegistry(),
	}

	// Register dashboard tools
	err := s.registerDashboardTools()
	require.NoError(t, err)

	t.Run("generate_dashboard with cross-account support", func(t *testing.T) {
		tests := []struct {
			name       string
			params     map[string]interface{}
			wantErr    bool
			checkFunc  func(t *testing.T, result interface{})
		}{
			{
				name: "golden signals with multiple accounts",
				params: map[string]interface{}{
					"template":     "golden-signals",
					"service_name": "my-service",
					"account_ids":  []interface{}{float64(12345), float64(67890)},
				},
				wantErr: false,
				checkFunc: func(t *testing.T, result interface{}) {
					res := result.(map[string]interface{})
					dashboard := res["dashboard"].(map[string]interface{})
					
					// Check account IDs are stored
					accountIDs := dashboard["account_ids"].([]int)
					assert.Equal(t, []int{12345, 67890}, accountIDs)
					
					// Check queries contain cross-account syntax
					pages := dashboard["pages"].([]map[string]interface{})
					widgets := pages[0]["widgets"].([]map[string]interface{})
					for _, widget := range widgets {
						query := widget["query"].(string)
						assert.Contains(t, query, "WITH accountIds = [12345, 67890]")
					}
				},
			},
			{
				name: "infrastructure dashboard with single account",
				params: map[string]interface{}{
					"template":     "infrastructure",
					"host_pattern": "prod-*",
					"account_ids":  []interface{}{float64(12345)},
				},
				wantErr: false,
				checkFunc: func(t *testing.T, result interface{}) {
					res := result.(map[string]interface{})
					dashboard := res["dashboard"].(map[string]interface{})
					
					// Check account IDs
					accountIDs := dashboard["account_ids"].([]int)
					assert.Equal(t, []int{12345}, accountIDs)
					
					// Check queries
					pages := dashboard["pages"].([]map[string]interface{})
					widgets := pages[0]["widgets"].([]map[string]interface{})
					for _, widget := range widgets {
						query := widget["query"].(string)
						assert.Contains(t, query, "WITH accountIds = [12345]")
					}
				},
			},
			{
				name: "dashboard without account IDs (backwards compatibility)",
				params: map[string]interface{}{
					"template":     "golden-signals",
					"service_name": "my-service",
				},
				wantErr: false,
				checkFunc: func(t *testing.T, result interface{}) {
					res := result.(map[string]interface{})
					dashboard := res["dashboard"].(map[string]interface{})
					
					// Should not have account_ids field
					_, hasAccountIDs := dashboard["account_ids"]
					assert.False(t, hasAccountIDs)
					
					// Queries should not contain WITH accountIds
					pages := dashboard["pages"].([]map[string]interface{})
					widgets := pages[0]["widgets"].([]map[string]interface{})
					for _, widget := range widgets {
						query := widget["query"].(string)
						assert.NotContains(t, query, "WITH accountIds")
					}
				},
			},
			{
				name: "SLI/SLO dashboard with cross-account",
				params: map[string]interface{}{
					"template": "sli-slo",
					"sli_config": map[string]interface{}{
						"name":   "Availability",
						"target": 99.9,
						"query":  "SELECT percentage(count(*), WHERE error IS false) FROM Transaction",
					},
					"account_ids": []interface{}{float64(11111), float64(22222), float64(33333)},
				},
				wantErr: false,
				checkFunc: func(t *testing.T, result interface{}) {
					res := result.(map[string]interface{})
					dashboard := res["dashboard"].(map[string]interface{})
					
					// Check multiple account IDs
					accountIDs := dashboard["account_ids"].([]int)
					assert.Equal(t, []int{11111, 22222, 33333}, accountIDs)
					
					// Check queries contain all account IDs
					pages := dashboard["pages"].([]map[string]interface{})
					widgets := pages[0]["widgets"].([]map[string]interface{})
					for _, widget := range widgets {
						query := widget["query"].(string)
						assert.Contains(t, query, "WITH accountIds = [11111, 22222, 33333]")
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := s.handleGenerateDashboard(context.Background(), tt.params)
				
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				
				require.NoError(t, err)
				require.NotNil(t, result)
				
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			})
		}
	})
}

func TestDashboardValidation(t *testing.T) {
	t.Run("validate dashboard limits", func(t *testing.T) {
		tests := []struct {
			name    string
			page    map[string]interface{}
			wantErr bool
			errMsg  string
		}{
			{
				name: "valid page",
				page: map[string]interface{}{
					"name": "Test Page",
					"widgets": []interface{}{
						map[string]interface{}{
							"title": "Widget 1",
							"query": "SELECT count(*) FROM Transaction",
							"layout": map[string]interface{}{
								"column": float64(1),
								"row":    float64(1),
								"width":  float64(6),
								"height": float64(3),
							},
						},
					},
				},
				wantErr: false,
			},
			{
				name: "too many widgets",
				page: map[string]interface{}{
					"name":    "Test Page",
					"widgets": make([]interface{}, 151), // More than MaxWidgetsPerPage
				},
				wantErr: true,
				errMsg:  "page has 151 widgets, maximum is 150",
			},
			{
				name: "widget extends beyond dashboard width",
				page: map[string]interface{}{
					"name": "Test Page",
					"widgets": []interface{}{
						map[string]interface{}{
							"title": "Widget 1",
							"query": "SELECT count(*) FROM Transaction",
							"layout": map[string]interface{}{
								"column": float64(10),
								"width":  float64(6), // 10 + 6 > 12
								"height": float64(3),
							},
						},
					},
				},
				wantErr: true,
				errMsg:  "widget extends beyond dashboard width",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateDashboardPage(tt.page)
				
				if tt.wantErr {
					assert.Error(t, err)
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}
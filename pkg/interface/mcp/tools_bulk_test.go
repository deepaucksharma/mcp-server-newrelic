package mcp

import (
	"context"
	"testing"
	"strings"
)

func TestHandleBulkTagEntities(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bulk tag with add operation",
			params: map[string]interface{}{
				"entity_guids": []interface{}{"entity-1", "entity-2"},
				"tags":         []interface{}{"env:production", "team:platform"},
				"operation":    "add",
			},
			wantErr: false,
		},
		{
			name: "valid bulk tag with replace operation",
			params: map[string]interface{}{
				"entity_guids": []interface{}{"entity-1"},
				"tags":         []interface{}{"env:staging"},
				"operation":    "replace",
			},
			wantErr: false,
		},
		{
			name: "default operation (add)",
			params: map[string]interface{}{
				"entity_guids": []interface{}{"entity-1"},
				"tags":         []interface{}{"env:production"},
			},
			wantErr: false,
		},
		{
			name:    "missing entity_guids",
			params:  map[string]interface{}{
				"tags": []interface{}{"env:production"},
			},
			wantErr: true,
			errMsg:  "entity_guids parameter is required",
		},
		{
			name: "empty entity_guids",
			params: map[string]interface{}{
				"entity_guids": []interface{}{},
				"tags":         []interface{}{"env:production"},
			},
			wantErr: true,
			errMsg:  "entity_guids parameter is required and must be non-empty",
		},
		{
			name:    "missing tags",
			params:  map[string]interface{}{
				"entity_guids": []interface{}{"entity-1"},
			},
			wantErr: true,
			errMsg:  "tags parameter is required",
		},
		{
			name: "invalid operation",
			params: map[string]interface{}{
				"entity_guids": []interface{}{"entity-1"},
				"tags":         []interface{}{"env:production"},
				"operation":    "invalid",
			},
			wantErr: true,
			errMsg:  "operation must be 'add' or 'replace'",
		},
		{
			name: "invalid entity GUID type",
			params: map[string]interface{}{
				"entity_guids": []interface{}{123, "entity-2"},
				"tags":         []interface{}{"env:production"},
			},
			wantErr: true,
			errMsg:  "invalid entity GUID at index 0",
		},
		{
			name: "invalid tag type",
			params: map[string]interface{}{
				"entity_guids": []interface{}{"entity-1"},
				"tags":         []interface{}{123, "env:production"},
			},
			wantErr: true,
			errMsg:  "invalid tag at index 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleBulkTagEntities(context.Background(), tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				} else {
					// Check result structure
					resultMap, ok := result.(map[string]interface{})
					if !ok {
						t.Errorf("expected result to be map[string]interface{}")
					} else {
						if summary, ok := resultMap["summary"].(map[string]interface{}); ok {
							requiredFields := []string{"total_entities", "total_tags", "operation", "success", "failed"}
							for _, field := range requiredFields {
								if _, ok := summary[field]; !ok {
									t.Errorf("expected summary to have '%s' field", field)
								}
							}
						} else {
							t.Errorf("expected result to have 'summary' field")
						}
					}
				}
			}
		})
	}
}

func TestHandleBulkCreateMonitors(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bulk create monitors",
			params: map[string]interface{}{
				"monitors": []interface{}{
					map[string]interface{}{
						"name": "Monitor 1",
						"url":  "https://example.com",
					},
					map[string]interface{}{
						"name": "Monitor 2",
						"url":  "https://example.org",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with template settings",
			params: map[string]interface{}{
				"monitors": []interface{}{
					map[string]interface{}{
						"name": "Monitor 1",
						"url":  "https://example.com",
					},
				},
				"template": map[string]interface{}{
					"type":   "PING",
					"status": "ENABLED",
					"tags":   []string{"team:platform", "env:production"},
				},
			},
			wantErr: false,
		},
		{
			name:    "missing monitors",
			params:  map[string]interface{}{},
			wantErr: true,
			errMsg:  "monitors parameter is required",
		},
		{
			name: "empty monitors",
			params: map[string]interface{}{
				"monitors": []interface{}{},
			},
			wantErr: true,
			errMsg:  "monitors parameter is required and must be non-empty",
		},
		{
			name: "invalid monitor configuration",
			params: map[string]interface{}{
				"monitors": []interface{}{
					"invalid monitor",
				},
			},
			wantErr: false, // Should handle gracefully with error in results
		},
		{
			name: "monitor missing name",
			params: map[string]interface{}{
				"monitors": []interface{}{
					map[string]interface{}{
						"url": "https://example.com",
					},
				},
			},
			wantErr: false, // Should handle gracefully with error in results
		},
		{
			name: "monitor missing url",
			params: map[string]interface{}{
				"monitors": []interface{}{
					map[string]interface{}{
						"name": "Monitor 1",
					},
				},
			},
			wantErr: false, // Should handle gracefully with error in results
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleBulkCreateMonitors(context.Background(), tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				} else {
					// Check result structure
					resultMap, ok := result.(map[string]interface{})
					if !ok {
						t.Errorf("expected result to be map[string]interface{}")
					} else {
						if summary, ok := resultMap["summary"].(map[string]interface{}); ok {
							requiredFields := []string{"total", "success", "failed"}
							for _, field := range requiredFields {
								if _, ok := summary[field]; !ok {
									t.Errorf("expected summary to have '%s' field", field)
								}
							}
						} else {
							t.Errorf("expected result to have 'summary' field")
						}
						if _, ok := resultMap["results"]; !ok {
							t.Errorf("expected result to have 'results' field")
						}
					}
				}
			}
		})
	}
}

func TestHandleBulkUpdateDashboards(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bulk update dashboards",
			params: map[string]interface{}{
				"dashboard_ids": []interface{}{"dash-1", "dash-2"},
				"updates": map[string]interface{}{
					"add_tags": []interface{}{"team:platform", "env:production"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple update operations",
			params: map[string]interface{}{
				"dashboard_ids": []interface{}{"dash-1"},
				"updates": map[string]interface{}{
					"add_tags":    []interface{}{"team:platform"},
					"remove_tags": []interface{}{"env:staging"},
					"permissions": "public",
				},
			},
			wantErr: false,
		},
		{
			name:    "missing dashboard_ids",
			params:  map[string]interface{}{
				"updates": map[string]interface{}{"add_tags": []interface{}{"team:platform"}},
			},
			wantErr: true,
			errMsg:  "dashboard_ids parameter is required",
		},
		{
			name: "empty dashboard_ids",
			params: map[string]interface{}{
				"dashboard_ids": []interface{}{},
				"updates":       map[string]interface{}{"add_tags": []interface{}{"team:platform"}},
			},
			wantErr: true,
			errMsg:  "dashboard_ids parameter is required and must be non-empty",
		},
		{
			name:    "missing updates",
			params:  map[string]interface{}{
				"dashboard_ids": []interface{}{"dash-1"},
			},
			wantErr: true,
			errMsg:  "updates parameter is required",
		},
		{
			name: "empty updates",
			params: map[string]interface{}{
				"dashboard_ids": []interface{}{"dash-1"},
				"updates":       map[string]interface{}{},
			},
			wantErr: true,
			errMsg:  "updates parameter is required and must be non-empty",
		},
		{
			name: "invalid dashboard ID type",
			params: map[string]interface{}{
				"dashboard_ids": []interface{}{123, "dash-2"},
				"updates":       map[string]interface{}{"add_tags": []interface{}{"team:platform"}},
			},
			wantErr: true,
			errMsg:  "invalid dashboard ID at index 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleBulkUpdateDashboards(context.Background(), tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				}
			}
		})
	}
}

func TestHandleBulkDeleteEntities(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bulk delete monitors",
			params: map[string]interface{}{
				"entity_type": "monitor",
				"entity_ids":  []interface{}{"monitor-1", "monitor-2"},
			},
			wantErr: false,
		},
		{
			name: "valid bulk delete dashboards with force",
			params: map[string]interface{}{
				"entity_type": "dashboard",
				"entity_ids":  []interface{}{"dash-1", "dash-2", "dash-3"},
				"force":       true,
			},
			wantErr: false,
		},
		{
			name: "valid bulk delete alert conditions",
			params: map[string]interface{}{
				"entity_type": "alert_condition",
				"entity_ids":  []interface{}{"alert-1"},
			},
			wantErr: false,
		},
		{
			name:    "missing entity_type",
			params:  map[string]interface{}{
				"entity_ids": []interface{}{"entity-1"},
			},
			wantErr: true,
			errMsg:  "entity_type parameter is required",
		},
		{
			name: "invalid entity_type",
			params: map[string]interface{}{
				"entity_type": "invalid_type",
				"entity_ids":  []interface{}{"entity-1"},
			},
			wantErr: true,
			errMsg:  "invalid entity_type",
		},
		{
			name:    "missing entity_ids",
			params:  map[string]interface{}{
				"entity_type": "monitor",
			},
			wantErr: true,
			errMsg:  "entity_ids parameter is required",
		},
		{
			name: "empty entity_ids",
			params: map[string]interface{}{
				"entity_type": "monitor",
				"entity_ids":  []interface{}{},
			},
			wantErr: true,
			errMsg:  "entity_ids parameter is required and must be non-empty",
		},
		{
			name: "too many entities without force",
			params: map[string]interface{}{
				"entity_type": "monitor",
				"entity_ids": []interface{}{
					"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8", "m9", "m10", "m11",
				},
			},
			wantErr: true,
			errMsg:  "set force=true to confirm",
		},
		{
			name: "invalid entity ID type",
			params: map[string]interface{}{
				"entity_type": "monitor",
				"entity_ids":  []interface{}{123, "monitor-2"},
			},
			wantErr: true,
			errMsg:  "invalid entity ID at index 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleBulkDeleteEntities(context.Background(), tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				}
			}
		})
	}
}

func TestHandleBulkExecuteQueries(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid bulk execute queries",
			params: map[string]interface{}{
				"queries": []interface{}{
					map[string]interface{}{
						"name":  "Query 1",
						"query": "SELECT count(*) FROM Transaction",
					},
					map[string]interface{}{
						"name":  "Query 2",
						"query": "SELECT average(duration) FROM Transaction",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with parallel execution disabled",
			params: map[string]interface{}{
				"queries": []interface{}{
					map[string]interface{}{
						"query": "SELECT count(*) FROM Transaction",
					},
				},
				"parallel": false,
			},
			wantErr: false,
		},
		{
			name: "with custom timeout",
			params: map[string]interface{}{
				"queries": []interface{}{
					map[string]interface{}{
						"query": "SELECT count(*) FROM Transaction",
					},
				},
				"timeout": 60,
			},
			wantErr: false,
		},
		{
			name:    "missing queries",
			params:  map[string]interface{}{},
			wantErr: true,
			errMsg:  "queries parameter is required",
		},
		{
			name: "empty queries",
			params: map[string]interface{}{
				"queries": []interface{}{},
			},
			wantErr: true,
			errMsg:  "queries parameter is required and must be non-empty",
		},
		{
			name: "invalid query configuration",
			params: map[string]interface{}{
				"queries": []interface{}{
					"invalid query",
				},
			},
			wantErr: false, // Should handle gracefully with error in results
		},
		{
			name: "query missing query string",
			params: map[string]interface{}{
				"queries": []interface{}{
					map[string]interface{}{
						"name": "Query 1",
					},
				},
			},
			wantErr: false, // Should handle gracefully with error in results
		},
		{
			name: "invalid NRQL query",
			params: map[string]interface{}{
				"queries": []interface{}{
					map[string]interface{}{
						"query": "INVALID QUERY",
					},
				},
			},
			wantErr: false, // Should handle gracefully with error in results
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleBulkExecuteQueries(context.Background(), tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				} else {
					// Check result structure
					resultMap, ok := result.(map[string]interface{})
					if !ok {
						t.Errorf("expected result to be map[string]interface{}")
					} else {
						if summary, ok := resultMap["summary"].(map[string]interface{}); ok {
							requiredFields := []string{"total", "success", "failed", "parallel", "total_time_ms"}
							for _, field := range requiredFields {
								if _, ok := summary[field]; !ok {
									t.Errorf("expected summary to have '%s' field", field)
								}
							}
						} else {
							t.Errorf("expected result to have 'summary' field")
						}
						if _, ok := resultMap["results"]; !ok {
							t.Errorf("expected result to have 'results' field")
						}
					}
				}
			}
		})
	}
}
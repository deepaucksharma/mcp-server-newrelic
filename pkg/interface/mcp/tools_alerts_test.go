package mcp

import (
	"context"
	"testing"
	"strings"
)

func TestHandleCreateAlert(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid alert with auto baseline",
			params: map[string]interface{}{
				"name":          "High Error Rate",
				"query":         "SELECT percentage(count(*), WHERE error IS true) FROM Transaction",
				"auto_baseline": true,
				"sensitivity":   "medium",
			},
			wantErr: false,
		},
		{
			name: "valid alert with static threshold",
			params: map[string]interface{}{
				"name":             "High Response Time",
				"query":            "SELECT average(duration) FROM Transaction",
				"auto_baseline":    false,
				"static_threshold": 5.0,
				"comparison":       "above",
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			params:  map[string]interface{}{
				"query": "SELECT count(*) FROM Transaction",
			},
			wantErr: true,
			errMsg:  "name parameter is required",
		},
		{
			name:    "missing query",
			params:  map[string]interface{}{
				"name": "Test Alert",
			},
			wantErr: true,
			errMsg:  "query parameter is required",
		},
		{
			name: "invalid NRQL query",
			params: map[string]interface{}{
				"name":  "Bad Alert",
				"query": "INVALID QUERY",
			},
			wantErr: true,
			errMsg:  "invalid NRQL query",
		},
		{
			name: "missing static threshold when auto_baseline is false",
			params: map[string]interface{}{
				"name":          "Missing Threshold",
				"query":         "SELECT count(*) FROM Transaction",
				"auto_baseline": false,
			},
			wantErr: true,
			errMsg:  "static_threshold is required when auto_baseline is false",
		},
		{
			name: "invalid sensitivity",
			params: map[string]interface{}{
				"name":          "Invalid Sensitivity",
				"query":         "SELECT count(*) FROM Transaction",
				"auto_baseline": true,
				"sensitivity":   "invalid",
			},
			wantErr: false, // Should use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleCreateAlert(context.Background(), tt.params)

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

func TestHandleCreateAlertPolicy(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid policy with default preference",
			params: map[string]interface{}{
				"name": "My Alert Policy",
			},
			wantErr: false,
		},
		{
			name: "valid policy with custom preference",
			params: map[string]interface{}{
				"name":                "My Alert Policy",
				"incident_preference": "PER_POLICY",
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			params:  map[string]interface{}{},
			wantErr: true,
			errMsg:  "name parameter is required",
		},
		{
			name: "empty name",
			params: map[string]interface{}{
				"name": "",
			},
			wantErr: true,
			errMsg:  "name parameter is required",
		},
		{
			name: "invalid incident preference",
			params: map[string]interface{}{
				"name":                "My Policy",
				"incident_preference": "INVALID_PREF",
			},
			wantErr: true,
			errMsg:  "invalid incident_preference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleCreateAlertPolicy(context.Background(), tt.params)

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
						if policy, ok := resultMap["policy"].(map[string]interface{}); ok {
							if _, ok := policy["id"]; !ok {
								t.Errorf("expected policy to have 'id' field")
							}
							if name, ok := policy["name"]; !ok || name != tt.params["name"] {
								t.Errorf("expected policy name to match input")
							}
						} else {
							t.Errorf("expected result to have 'policy' field")
						}
					}
				}
			}
		})
	}
}

func TestHandleCreateAlertCondition(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid condition",
			params: map[string]interface{}{
				"policy_id":           "policy-123",
				"name":                "High Error Rate",
				"query":               "SELECT count(*) FROM Transaction WHERE error IS true",
				"threshold":           10.0,
				"threshold_duration":  5.0,
				"comparison":          "above",
			},
			wantErr: false,
		},
		{
			name:    "missing policy_id",
			params:  map[string]interface{}{
				"name":      "Test Condition",
				"query":     "SELECT count(*) FROM Transaction",
				"threshold": 10.0,
			},
			wantErr: true,
			errMsg:  "policy_id parameter is required",
		},
		{
			name:    "missing name",
			params:  map[string]interface{}{
				"policy_id": "policy-123",
				"query":     "SELECT count(*) FROM Transaction",
				"threshold": 10.0,
			},
			wantErr: true,
			errMsg:  "name parameter is required",
		},
		{
			name:    "missing query",
			params:  map[string]interface{}{
				"policy_id": "policy-123",
				"name":      "Test Condition",
				"threshold": 10.0,
			},
			wantErr: true,
			errMsg:  "query parameter is required",
		},
		{
			name:    "missing threshold",
			params:  map[string]interface{}{
				"policy_id": "policy-123",
				"name":      "Test Condition",
				"query":     "SELECT count(*) FROM Transaction",
			},
			wantErr: true,
			errMsg:  "threshold parameter is required",
		},
		{
			name: "invalid comparison",
			params: map[string]interface{}{
				"policy_id":  "policy-123",
				"name":       "Test Condition",
				"query":      "SELECT count(*) FROM Transaction",
				"threshold":  10.0,
				"comparison": "invalid",
			},
			wantErr: true,
			errMsg:  "invalid comparison",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleCreateAlertCondition(context.Background(), tt.params)

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

func TestHandleAnalyzeAlerts(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid analysis request",
			params: map[string]interface{}{
				"alert_id": "alert-123",
			},
			wantErr: false,
		},
		{
			name: "with custom time range",
			params: map[string]interface{}{
				"alert_id":   "alert-123",
				"time_range": "30 days",
			},
			wantErr: false,
		},
		{
			name:    "missing alert_id",
			params:  map[string]interface{}{},
			wantErr: true,
			errMsg:  "alert_id parameter is required",
		},
		{
			name: "empty alert_id",
			params: map[string]interface{}{
				"alert_id": "",
			},
			wantErr: true,
			errMsg:  "alert_id parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleAnalyzeAlerts(context.Background(), tt.params)

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
					// Check result has expected structure
					resultMap, ok := result.(map[string]interface{})
					if !ok {
						t.Errorf("expected result to be map[string]interface{}")
					} else {
						if _, ok := resultMap["alert_id"]; !ok {
							t.Errorf("expected result to have 'alert_id' field")
						}
						if _, ok := resultMap["summary"]; !ok {
							t.Errorf("expected result to have 'summary' field")
						}
						if _, ok := resultMap["effectiveness"]; !ok {
							t.Errorf("expected result to have 'effectiveness' field")
						}
					}
				}
			}
		})
	}
}

func TestHandleBulkUpdateAlerts(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "enable alerts",
			params: map[string]interface{}{
				"alert_ids": []interface{}{"alert-1", "alert-2"},
				"operation": "enable",
			},
			wantErr: false,
		},
		{
			name: "update threshold with new value",
			params: map[string]interface{}{
				"alert_ids":     []interface{}{"alert-1"},
				"operation":     "update_threshold",
				"new_threshold": 15.0,
			},
			wantErr: false,
		},
		{
			name: "update threshold with multiplier",
			params: map[string]interface{}{
				"alert_ids":            []interface{}{"alert-1"},
				"operation":            "update_threshold",
				"threshold_multiplier": 1.5,
			},
			wantErr: false,
		},
		{
			name:    "missing alert_ids",
			params:  map[string]interface{}{
				"operation": "enable",
			},
			wantErr: true,
			errMsg:  "alert_ids parameter is required",
		},
		{
			name: "empty alert_ids",
			params: map[string]interface{}{
				"alert_ids": []interface{}{},
				"operation": "enable",
			},
			wantErr: true,
			errMsg:  "alert_ids parameter is required",
		},
		{
			name:    "missing operation",
			params:  map[string]interface{}{
				"alert_ids": []interface{}{"alert-1"},
			},
			wantErr: true,
			errMsg:  "operation parameter is required",
		},
		{
			name: "invalid operation",
			params: map[string]interface{}{
				"alert_ids": []interface{}{"alert-1"},
				"operation": "invalid_op",
			},
			wantErr: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleBulkUpdateAlerts(context.Background(), tt.params)

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
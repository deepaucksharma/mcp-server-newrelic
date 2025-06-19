package mcp

import (
	"context"
	"testing"
	"strings"
)

func TestHandleQueryNRDB(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid query",
			params: map[string]interface{}{
				"query": "SELECT count(*) FROM Transaction",
			},
			wantErr: false,
		},
		{
			name:    "missing query",
			params:  map[string]interface{}{},
			wantErr: true,
			errMsg:  "query parameter is required",
		},
		{
			name: "empty query",
			params: map[string]interface{}{
				"query": "",
			},
			wantErr: true,
			errMsg:  "query parameter is required",
		},
		{
			name: "invalid query type",
			params: map[string]interface{}{
				"query": 123,
			},
			wantErr: true,
			errMsg:  "query parameter is required",
		},
		{
			name: "with timeout",
			params: map[string]interface{}{
				"query":   "SELECT count(*) FROM Transaction",
				"timeout": 60.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{} // Mock mode
			result, err := s.handleQueryNRDB(context.Background(), tt.params)

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

func TestHandleQueryCheck(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid NRQL query",
			params: map[string]interface{}{
				"query": "SELECT count(*) FROM Transaction WHERE appName = 'myapp'",
			},
			wantErr: false,
		},
		{
			name: "missing query",
			params: map[string]interface{}{},
			wantErr: true,
			errMsg:  "query parameter is required",
		},
		{
			name: "invalid NRQL - missing FROM",
			params: map[string]interface{}{
				"query": "SELECT count(*) WHERE appName = 'myapp'",
			},
			wantErr: false, // Should return validation result, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleQueryCheck(context.Background(), tt.params)

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
					// Check result has expected fields
					resultMap, ok := result.(map[string]interface{})
					if !ok {
						t.Errorf("expected result to be map[string]interface{}")
					} else {
						if _, ok := resultMap["valid"]; !ok {
							t.Errorf("expected result to have 'valid' field")
						}
					}
				}
			}
		})
	}
}

func TestHandleQueryBuilder(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
		check   func(t *testing.T, result interface{})
	}{
		{
			name: "simple count query",
			params: map[string]interface{}{
				"select":     "count(*)",
				"event_type": "Transaction",
			},
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				resultMap, _ := result.(map[string]interface{})
				query, _ := resultMap["query"].(string)
				if !strings.Contains(query, "SELECT count(*)") {
					t.Errorf("expected query to contain 'SELECT count(*)', got %s", query)
				}
				if !strings.Contains(query, "FROM Transaction") {
					t.Errorf("expected query to contain 'FROM Transaction', got %s", query)
				}
			},
		},
		{
			name: "query with where clause",
			params: map[string]interface{}{
				"select":     "average(duration)",
				"event_type": "Transaction",
				"where":      "appName = 'myapp'",
			},
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				resultMap, _ := result.(map[string]interface{})
				query, _ := resultMap["query"].(string)
				if !strings.Contains(query, "WHERE appName = 'myapp'") {
					t.Errorf("expected query to contain WHERE clause, got %s", query)
				}
			},
		},
		{
			name: "query with facet",
			params: map[string]interface{}{
				"select":     "count(*)",
				"event_type": "Transaction",
				"facet":      "appName",
			},
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				resultMap, _ := result.(map[string]interface{})
				query, _ := resultMap["query"].(string)
				if !strings.Contains(query, "FACET appName") {
					t.Errorf("expected query to contain FACET clause, got %s", query)
				}
			},
		},
		{
			name:    "missing required select",
			params:  map[string]interface{}{
				"event_type": "Transaction",
			},
			wantErr: true,
			errMsg:  "select parameter is required",
		},
		{
			name:    "missing required event_type",
			params:  map[string]interface{}{
				"select": "count(*)",
			},
			wantErr: true,
			errMsg:  "event_type parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			result, err := s.handleQueryBuilder(context.Background(), tt.params)

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
				} else if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

func TestValidateNRQLSyntax(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "valid simple query",
			query:   "SELECT count(*) FROM Transaction",
			wantErr: false,
		},
		{
			name:    "valid query with WHERE",
			query:   "SELECT average(duration) FROM Transaction WHERE appName = 'myapp' SINCE 1 hour ago",
			wantErr: false,
		},
		{
			name:    "valid query with FACET",
			query:   "SELECT count(*) FROM Transaction FACET appName LIMIT 10",
			wantErr: false,
		},
		{
			name:    "missing FROM clause",
			query:   "SELECT count(*) WHERE appName = 'myapp'",
			wantErr: true,
		},
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
		},
		{
			name:    "incomplete query",
			query:   "SELECT",
			wantErr: true,
		},
		{
			name:    "SQL injection attempt",
			query:   "SELECT * FROM Transaction; DROP TABLE users;",
			wantErr: false, // validateNRQLSyntax doesn't check for SQL injection, validateNRQLSafety does
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			err := s.validateNRQLSyntax(tt.query)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for query %q but got none", tt.query)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for query %q: %v", tt.query, err)
				}
			}
		})
	}
}
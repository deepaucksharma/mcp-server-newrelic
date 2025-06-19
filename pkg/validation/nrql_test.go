package validation

import (
	"testing"
)

func TestNRQLValidator_Sanitize(t *testing.T) {
	v := NewNRQLValidator()
	
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		// Valid queries
		{
			name:    "simple select",
			query:   "SELECT count(*) FROM Transaction",
			wantErr: false,
		},
		{
			name:    "with where clause",
			query:   "SELECT * FROM Transaction WHERE appName = 'myapp'",
			wantErr: false,
		},
		{
			name:    "with time range",
			query:   "SELECT * FROM Transaction SINCE 1 hour ago",
			wantErr: false,
		},
		// SQL injection attempts
		{
			name:    "semicolon injection",
			query:   "SELECT * FROM Transaction; DROP TABLE users;",
			wantErr: true,
		},
		{
			name:    "union injection",
			query:   "SELECT * FROM Transaction UNION SELECT * FROM credentials",
			wantErr: true,
		},
		{
			name:    "comment injection",
			query:   "SELECT * FROM Transaction -- comment",
			wantErr: true,
		},
		{
			name:    "or injection",
			query:   "SELECT * FROM Transaction WHERE name = 'test' OR '1'='1'",
			wantErr: true,
		},
		// Dangerous operations
		{
			name:    "drop table",
			query:   "DROP TABLE Transaction",
			wantErr: true,
		},
		{
			name:    "delete from",
			query:   "DELETE FROM Transaction",
			wantErr: true,
		},
		{
			name:    "update",
			query:   "UPDATE Transaction SET name = 'hacked'",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := v.Sanitize(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sanitize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


func TestNRQLValidator_ValidateTimeRange(t *testing.T) {
	v := NewNRQLValidator()
	
	tests := []struct {
		name      string
		timeRange string
		wantErr   bool
	}{
		{
			name:      "1 hour ago",
			timeRange: "1 hour ago",
			wantErr:   false,
		},
		{
			name:      "30 minutes ago",
			timeRange: "30 minutes ago",
			wantErr:   false,
		},
		{
			name:      "7 days ago",
			timeRange: "7 days ago",
			wantErr:   false,
		},
		{
			name:      "yesterday",
			timeRange: "yesterday",
			wantErr:   false,
		},
		{
			name:      "injection attempt",
			timeRange: "1 hour ago; DROP TABLE",
			wantErr:   true,
		},
		{
			name:      "invalid format",
			timeRange: "now",
			wantErr:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateTimeRange(tt.timeRange)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
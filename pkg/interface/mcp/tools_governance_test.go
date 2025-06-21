//go:build test

package mcp

import (
	"context"
)

// RegisterGovernanceGranularTools registers governance tools for testing
func (s *Server) RegisterGovernanceGranularTools() error {
	// Usage analysis - mock implementation
	s.tools.Register(Tool{
		Name:        "governance.analyze_usage",
		Description: "Analyze platform usage and resource consumption",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]Property{
				"time_range": {
					Type:        "string",
					Description: "Time range for analysis",
					Default:     "7 days",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"usage_summary": map[string]interface{}{
					"total_events": 1500000,
					"total_queries": 2500,
					"unique_users": 45,
					"data_ingested_gb": 125.5,
				},
				"top_consumers": []map[string]interface{}{
					{"service": "frontend-api", "events": 500000},
					{"service": "backend-api", "events": 400000},
					{"service": "worker-service", "events": 300000},
				},
			}, nil
		},
	})

	// Cost optimization - mock implementation
	s.tools.Register(Tool{
		Name:        "governance.optimize_costs",
		Description: "Analyze and optimize data costs",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]Property{
				"focus_area": {
					Type:        "string",
					Description: "Area to focus optimization",
					Default:     "all",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"recommendations": []map[string]interface{}{
					{
						"type": "data_retention",
						"description": "Reduce retention for low-value event types",
						"potential_savings": "$1,200/month",
					},
					{
						"type": "query_optimization", 
						"description": "Cache frequently-run queries",
						"potential_savings": "$300/month",
					},
				},
				"potential_savings": "$1,500/month",
			}, nil
		},
	})

	// Compliance check - mock implementation
	s.tools.Register(Tool{
		Name:        "governance.check_compliance",
		Description: "Check data retention and compliance status",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]Property{
				"policy_type": {
					Type:        "string",
					Description: "Type of policy to check",
					Default:     "all",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"compliance_status": "compliant",
				"violations": []interface{}{},
				"checks_performed": 15,
			}, nil
		},
	})

	return nil
}
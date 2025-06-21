//go:build test

package mcp

import (
	"context"
	"fmt"
	"time"
)

// registerAlertTools registers alert-related tools for testing
func (s *Server) registerAlertTools() error {
	// List alerts - mock implementation for testing
	s.tools.Register(Tool{
		Name:        "list_alerts",
		Description: "List all alert conditions in the account with pagination support",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]Property{
				"limit": {
					Type:        "integer",
					Description: "Number of alerts to return",
					Default:     25,
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			limit := 10
			if l, ok := params["limit"].(float64); ok {
				limit = int(l)
			}
			
			alerts := make([]map[string]interface{}, 0)
			for i := 0; i < limit && i < 5; i++ {
				alerts = append(alerts, map[string]interface{}{
					"id":   fmt.Sprintf("alert-%d", i+1),
					"name": fmt.Sprintf("Test Alert %d", i+1),
				})
			}
			
			return map[string]interface{}{
				"alerts": alerts,
				"total":  5,
			}, nil
		},
	})

	return nil
}

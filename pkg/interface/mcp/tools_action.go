package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// RegisterActionTools registers all action-oriented tools (create, update, delete)
func (s *Server) RegisterActionTools() error {
	// Alert Management Tools
	if err := s.registerActionAlertTools(); err != nil {
		return err
	}

	// Dashboard Management Tools
	if err := s.registerActionDashboardTools(); err != nil {
		return err
	}

	// SLO Management Tools
	if err := s.registerSLOTools(); err != nil {
		return err
	}

	// Report Generation Tools
	if err := s.registerReportTools(); err != nil {
		return err
	}

	return nil
}

// Alert Management Tools
func (s *Server) registerActionAlertTools() error {
	// Create alert from baseline
	s.tools.Register(Tool{
		Name:        "alert.create_from_baseline",
		Description: "Create an alert policy based on discovered baselines",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"name", "metric", "event_type"},
			Properties: map[string]Property{
				"name": {
					Type:        "string",
					Description: "Alert policy name",
				},
				"metric": {
					Type:        "string",
					Description: "Metric to monitor (e.g., duration, error_rate)",
				},
				"event_type": {
					Type:        "string",
					Description: "Event type containing the metric",
				},
				"baseline_window": {
					Type:        "string",
					Description: "Time window for baseline calculation",
					Default:     "7 days",
				},
				"threshold_multiplier": {
					Type:        "number",
					Description: "Multiplier for baseline to set threshold (e.g., 1.5 for 150%)",
					Default:     1.5,
				},
				"comparison": {
					Type:        "string",
					Description: "Comparison operator",
					Enum:        []string{"above", "below"},
					Default:     "above",
				},
				"account_id": {
					Type:        "string",
					Description: "Target account ID (optional)",
				},
			},
		},
		Handler: s.handleAlertCreateFromBaseline,
	})

	// Create custom alert
	s.tools.Register(Tool{
		Name:        "alert.create_custom",
		Description: "Create a custom alert policy with specified conditions",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"name", "query", "threshold"},
			Properties: map[string]Property{
				"name": {
					Type:        "string",
					Description: "Alert policy name",
				},
				"query": {
					Type:        "string",
					Description: "NRQL query for the alert condition",
				},
				"threshold": {
					Type:        "number",
					Description: "Threshold value",
				},
				"threshold_duration": {
					Type:        "integer",
					Description: "Duration in seconds threshold must be exceeded",
					Default:     300,
				},
				"comparison": {
					Type:        "string",
					Description: "Comparison operator",
					Enum:        []string{"above", "below", "equals"},
					Default:     "above",
				},
				"enabled": {
					Type:        "boolean",
					Description: "Whether to enable the alert immediately",
					Default:     true,
				},
			},
		},
		Handler: s.handleAlertCreateCustom,
	})

	// Update alert
	s.tools.Register(Tool{
		Name:        "alert.update",
		Description: "Update an existing alert policy",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"alert_id"},
			Properties: map[string]Property{
				"alert_id": {
					Type:        "string",
					Description: "Alert policy ID to update",
				},
				"name": {
					Type:        "string",
					Description: "New name for the alert",
				},
				"threshold": {
					Type:        "number",
					Description: "New threshold value",
				},
				"enabled": {
					Type:        "boolean",
					Description: "Enable or disable the alert",
				},
			},
		},
		Handler: s.handleAlertUpdate,
	})

	// Delete alert
	s.tools.Register(Tool{
		Name:        "alert.delete",
		Description: "Delete an alert policy",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"alert_id"},
			Properties: map[string]Property{
				"alert_id": {
					Type:        "string",
					Description: "Alert policy ID to delete",
				},
			},
		},
		Handler: s.handleAlertDelete,
	})

	return nil
}

// Dashboard Management Tools
func (s *Server) registerActionDashboardTools() error {
	// Create dashboard from discovery
	s.tools.Register(Tool{
		Name:        "dashboard.create_from_discovery",
		Description: "Create a dashboard based on discovered attributes and patterns",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"title", "event_type"},
			Properties: map[string]Property{
				"title": {
					Type:        "string",
					Description: "Dashboard title",
				},
				"event_type": {
					Type:        "string",
					Description: "Primary event type to visualize",
				},
				"attributes": {
					Type:        "array",
					Description: "Specific attributes to include (auto-discovered if not specified)",
					Items: &Property{
						Type: "string",
					},
				},
				"layout": {
					Type:        "string",
					Description: "Dashboard layout style",
					Enum:        []string{"grid", "stacked", "overview"},
					Default:     "grid",
				},
				"time_range": {
					Type:        "string",
					Description: "Default time range for widgets",
					Default:     "last 1 hour",
				},
				"account_id": {
					Type:        "string",
					Description: "Target account ID (optional)",
				},
			},
		},
		Handler: s.handleDashboardCreateFromDiscovery,
	})

	// Create custom dashboard
	s.tools.Register(Tool{
		Name:        "dashboard.create_custom",
		Description: "Create a custom dashboard with specified widgets",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"title", "widgets"},
			Properties: map[string]Property{
				"title": {
					Type:        "string",
					Description: "Dashboard title",
				},
				"description": {
					Type:        "string",
					Description: "Dashboard description",
				},
				"widgets": {
					Type:        "array",
					Description: "Widget configurations",
					Items: &Property{
						Type: "object",
					},
				},
				"permissions": {
					Type:        "string",
					Description: "Dashboard permissions",
					Enum:        []string{"private", "public_read_only", "public_read_write"},
					Default:     "private",
				},
			},
		},
		Handler: s.handleDashboardCreateCustom,
	})

	// Update dashboard
	s.tools.Register(Tool{
		Name:        "dashboard.update",
		Description: "Update an existing dashboard",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_id"},
			Properties: map[string]Property{
				"dashboard_id": {
					Type:        "string",
					Description: "Dashboard ID to update",
				},
				"title": {
					Type:        "string",
					Description: "New title",
				},
				"add_widgets": {
					Type:        "array",
					Description: "Widgets to add",
					Items: &Property{
						Type: "object",
					},
				},
				"remove_widget_ids": {
					Type:        "array",
					Description: "Widget IDs to remove",
					Items: &Property{
						Type: "string",
					},
				},
			},
		},
		Handler: s.handleDashboardUpdate,
	})

	// Delete dashboard
	s.tools.Register(Tool{
		Name:        "dashboard.delete",
		Description: "Delete a dashboard",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_id"},
			Properties: map[string]Property{
				"dashboard_id": {
					Type:        "string",
					Description: "Dashboard ID to delete",
				},
			},
		},
		Handler: s.handleDashboardDelete,
	})

	return nil
}

// SLO Management Tools
func (s *Server) registerSLOTools() error {
	// Create SLO
	s.tools.Register(Tool{
		Name:        "slo.create",
		Description: "Create a Service Level Objective",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"name", "target", "indicator_query"},
			Properties: map[string]Property{
				"name": {
					Type:        "string",
					Description: "SLO name",
				},
				"target": {
					Type:        "number",
					Description: "Target percentage (e.g., 99.9)",
				},
				"indicator_query": {
					Type:        "string",
					Description: "NRQL query for the SLI",
				},
				"time_window": {
					Type:        "string",
					Description: "Rolling time window",
					Enum:        []string{"7d", "28d", "30d"},
					Default:     "28d",
				},
			},
		},
		Handler: s.handleSLOCreate,
	})

	return nil
}

// Report Generation Tools
func (s *Server) registerReportTools() error {
	// Generate investigation report
	s.tools.Register(Tool{
		Name:        "report.generate_investigation",
		Description: "Generate a comprehensive investigation report",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"title", "findings"},
			Properties: map[string]Property{
				"title": {
					Type:        "string",
					Description: "Report title",
				},
				"findings": {
					Type:        "array",
					Description: "Investigation findings",
					Items: &Property{
						Type: "object",
					},
				},
				"recommendations": {
					Type:        "array",
					Description: "Action recommendations",
					Items: &Property{
						Type: "string",
					},
				},
				"format": {
					Type:        "string",
					Description: "Output format",
					Enum:        []string{"markdown", "pdf", "html"},
					Default:     "markdown",
				},
			},
		},
		Handler: s.handleReportGenerateInvestigation,
	})

	return nil
}

// Handler implementations

func (s *Server) handleAlertCreateFromBaseline(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, _ := params["name"].(string)
	metric, _ := params["metric"].(string)
	eventType, _ := params["event_type"].(string)
	baselineWindow, _ := params["baseline_window"].(string)
	if baselineWindow == "" {
		baselineWindow = "7 days"
	}
	
	multiplier := 1.5
	if m, ok := params["threshold_multiplier"].(float64); ok {
		multiplier = m
	}

	comparison := "above"
	if c, ok := params["comparison"].(string); ok {
		comparison = c
	}

	// In mock mode, return sample response
	if s.isMockMode() {
		return map[string]interface{}{
			"alert_id": fmt.Sprintf("alert_%d", time.Now().Unix()),
			"name":     name,
			"status":   "created",
			"baseline": map[string]interface{}{
				"metric":     metric,
				"event_type": eventType,
				"p95":        100.0,
				"p99":        150.0,
				"threshold":  150.0 * multiplier,
			},
			"policy": map[string]interface{}{
				"comparison": comparison,
				"enabled":    true,
			},
		}, nil
	}

	// TODO: Implement real alert creation
	// 1. Calculate baseline using NRQL
	// 2. Create alert policy via NerdGraph
	// 3. Return alert details

	return map[string]interface{}{
		"error": "Alert creation not yet implemented in production mode",
	}, nil
}

func (s *Server) handleAlertCreateCustom(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, _ := params["name"].(string)
	query, _ := params["query"].(string)
	threshold, _ := params["threshold"].(float64)

	if s.isMockMode() {
		return map[string]interface{}{
			"alert_id": fmt.Sprintf("alert_%d", time.Now().Unix()),
			"name":     name,
			"status":   "created",
			"condition": map[string]interface{}{
				"query":     query,
				"threshold": threshold,
			},
		}, nil
	}

	return map[string]interface{}{
		"error": "Custom alert creation not yet implemented",
	}, nil
}

func (s *Server) handleAlertUpdate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	alertID, _ := params["alert_id"].(string)

	if s.isMockMode() {
		return map[string]interface{}{
			"alert_id": alertID,
			"status":   "updated",
			"message":  "Alert policy updated successfully",
		}, nil
	}

	return map[string]interface{}{
		"error": "Alert update not yet implemented",
	}, nil
}

func (s *Server) handleAlertDelete(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	alertID, _ := params["alert_id"].(string)

	if s.isMockMode() {
		return map[string]interface{}{
			"alert_id": alertID,
			"status":   "deleted",
			"message":  "Alert policy deleted successfully",
		}, nil
	}

	return map[string]interface{}{
		"error": "Alert deletion not yet implemented",
	}, nil
}

func (s *Server) handleDashboardCreateFromDiscovery(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	title, _ := params["title"].(string)
	eventType, _ := params["event_type"].(string)
	
	// Extract attributes if provided
	var attributes []string
	if attrs, ok := params["attributes"].([]interface{}); ok {
		for _, a := range attrs {
			if attr, ok := a.(string); ok {
				attributes = append(attributes, attr)
			}
		}
	}

	layout := "grid"
	if l, ok := params["layout"].(string); ok {
		layout = l
	}

	timeRange := "last 1 hour"
	if tr, ok := params["time_range"].(string); ok {
		timeRange = tr
	}

	if s.isMockMode() {
		// Generate mock dashboard with widgets
		widgets := []map[string]interface{}{
			{
				"id":    "widget_1",
				"type":  "line",
				"title": fmt.Sprintf("%s Count Over Time", eventType),
				"query": fmt.Sprintf("SELECT count(*) FROM %s TIMESERIES SINCE %s", eventType, timeRange),
			},
			{
				"id":    "widget_2",
				"type":  "billboard",
				"title": "Total Events",
				"query": fmt.Sprintf("SELECT count(*) FROM %s SINCE %s", eventType, timeRange),
			},
		}

		// Add attribute-specific widgets if provided
		for i, attr := range attributes {
			if i >= 4 { // Limit to 6 total widgets
				break
			}
			widgets = append(widgets, map[string]interface{}{
				"id":    fmt.Sprintf("widget_%d", i+3),
				"type":  "bar",
				"title": fmt.Sprintf("Top %s Values", strings.Title(attr)),
				"query": fmt.Sprintf("SELECT count(*) FROM %s FACET %s LIMIT 10 SINCE %s", eventType, attr, timeRange),
			})
		}

		return map[string]interface{}{
			"dashboard_id": fmt.Sprintf("dash_%d", time.Now().Unix()),
			"title":        title,
			"status":       "created",
			"url":          fmt.Sprintf("https://one.newrelic.com/dashboards/%d", time.Now().Unix()),
			"widgets":      widgets,
			"layout":       layout,
		}, nil
	}

	// TODO: Implement real dashboard creation
	// 1. If attributes not provided, discover them
	// 2. Generate appropriate widget configurations
	// 3. Create dashboard via NerdGraph
	// 4. Return dashboard details

	return map[string]interface{}{
		"error": "Dashboard creation not yet implemented in production mode",
	}, nil
}

func (s *Server) handleDashboardCreateCustom(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	title, _ := params["title"].(string)
	
	if s.isMockMode() {
		return map[string]interface{}{
			"dashboard_id": fmt.Sprintf("dash_%d", time.Now().Unix()),
			"title":        title,
			"status":       "created",
		}, nil
	}

	return map[string]interface{}{
		"error": "Custom dashboard creation not yet implemented",
	}, nil
}

func (s *Server) handleDashboardUpdate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardID, _ := params["dashboard_id"].(string)

	if s.isMockMode() {
		return map[string]interface{}{
			"dashboard_id": dashboardID,
			"status":       "updated",
			"message":      "Dashboard updated successfully",
		}, nil
	}

	return map[string]interface{}{
		"error": "Dashboard update not yet implemented",
	}, nil
}

func (s *Server) handleDashboardDelete(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardID, _ := params["dashboard_id"].(string)

	if s.isMockMode() {
		return map[string]interface{}{
			"dashboard_id": dashboardID,
			"status":       "deleted",
			"message":      "Dashboard deleted successfully",
		}, nil
	}

	return map[string]interface{}{
		"error": "Dashboard deletion not yet implemented",
	}, nil
}

func (s *Server) handleSLOCreate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, _ := params["name"].(string)
	target, _ := params["target"].(float64)
	query, _ := params["indicator_query"].(string)

	if s.isMockMode() {
		return map[string]interface{}{
			"slo_id":  fmt.Sprintf("slo_%d", time.Now().Unix()),
			"name":    name,
			"target":  target,
			"query":   query,
			"status":  "created",
			"current": 99.95, // Mock current performance
		}, nil
	}

	return map[string]interface{}{
		"error": "SLO creation not yet implemented",
	}, nil
}

func (s *Server) handleReportGenerateInvestigation(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	title, _ := params["title"].(string)
	format := "markdown"
	if f, ok := params["format"].(string); ok {
		format = f
	}

	if s.isMockMode() {
		// Generate mock report
		report := fmt.Sprintf(`# %s

## Executive Summary
Investigation completed with %d findings and %d recommendations.

## Findings
1. **Performance Degradation**: Response time increased by 150%% 
2. **Error Rate Spike**: Error rate jumped from 0.1%% to 2.5%%
3. **Database Bottleneck**: Query duration increased significantly

## Root Cause
Database connection pool exhaustion due to increased traffic.

## Recommendations
1. Increase database connection pool size
2. Implement query caching
3. Add database read replicas

## Timeline
- 14:00 - First signs of degradation
- 14:15 - Error rate spike detected
- 14:30 - Root cause identified
- 15:00 - Mitigation applied

Generated: %s
`, title, 3, 3, time.Now().Format(time.RFC3339))

		return map[string]interface{}{
			"report_id": fmt.Sprintf("report_%d", time.Now().Unix()),
			"title":     title,
			"format":    format,
			"content":   report,
			"status":    "generated",
		}, nil
	}

	return map[string]interface{}{
		"error": "Report generation not yet implemented",
	}, nil
}
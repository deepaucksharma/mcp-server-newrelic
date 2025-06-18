//go:build !test

package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// registerDashboardTools registers dashboard-related tools
func (s *Server) registerDashboardTools() error {
	// Find dashboard usage
	s.tools.Register(Tool{
		Name:        "find_usage",
		Description: "Find dashboards that use specific metrics, attributes, or event types",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"search_term"},
			Properties: map[string]Property{
				"search_term": {
					Type:        "string",
					Description: "Metric, attribute, or event type to search for",
				},
				"search_type": {
					Type:        "string",
					Description: "Type of search: 'metric', 'attribute', 'event_type', or 'any' (default: 'any')",
					Default:     "any",
				},
				"include_widgets": {
					Type:        "boolean",
					Description: "Include widget details in results",
					Default:     false,
				},
			},
		},
		Handler: s.handleFindUsage,
	})

	// Generate dashboard from template
	s.tools.Register(Tool{
		Name:        "generate_dashboard",
		Description: "Generate a dashboard from predefined templates or custom configuration",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"template"},
			Properties: map[string]Property{
				"template": {
					Type:        "string",
					Description: "Template name: 'golden-signals', 'sli-slo', 'infrastructure', 'custom'",
				},
				"name": {
					Type:        "string",
					Description: "Dashboard name (auto-generated if not provided)",
				},
				"service_name": {
					Type:        "string",
					Description: "Service name for golden-signals template",
				},
				"host_pattern": {
					Type:        "string",
					Description: "Host pattern for infrastructure template",
				},
				"sli_config": {
					Type:        "object",
					Description: "SLI configuration for sli-slo template",
				},
				"custom_config": {
					Type:        "object",
					Description: "Custom dashboard configuration",
				},
			},
		},
		Handler: s.handleGenerateDashboard,
	})

	// List dashboards
	s.tools.Register(Tool{
		Name:        "list_dashboards",
		Description: "List all dashboards in the account",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]Property{
				"filter": {
					Type:        "string",
					Description: "Filter dashboards by name",
				},
				"include_metadata": {
					Type:        "boolean",
					Description: "Include dashboard metadata",
					Default:     false,
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of dashboards to return",
					Default:     50,
				},
			},
		},
		Handler: s.handleListDashboards,
	})

	// Get dashboard details
	s.tools.Register(Tool{
		Name:        "get_dashboard",
		Description: "Get detailed information about a specific dashboard",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_id"},
			Properties: map[string]Property{
				"dashboard_id": {
					Type:        "string",
					Description: "Dashboard GUID or ID",
				},
				"include_queries": {
					Type:        "boolean",
					Description: "Include NRQL queries from widgets",
					Default:     true,
				},
			},
		},
		Handler: s.handleGetDashboard,
	})

	return nil
}

// handleFindUsage finds dashboards using specific metrics or attributes
func (s *Server) handleFindUsage(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	searchTerm, ok := params["search_term"].(string)
	if !ok || searchTerm == "" {
		return nil, fmt.Errorf("search_term parameter is required")
	}

	searchType := "any"
	if st, ok := params["search_type"].(string); ok {
		searchType = st
	}

	includeWidgets, _ := params["include_widgets"].(bool)

	// Search for dashboards containing the search term
	dashboards, err := s.searchDashboards(ctx, searchTerm, searchType)
	if err != nil {
		return nil, fmt.Errorf("failed to search dashboards: %w", err)
	}

	// Format results
	results := []map[string]interface{}{}
	for _, dashboard := range dashboards {
		result := map[string]interface{}{
			"dashboard_id":   dashboard["id"],
			"dashboard_name": dashboard["name"],
			"match_count":    dashboard["match_count"],
			"last_updated":   dashboard["updated_at"],
		}

		if includeWidgets {
			result["matching_widgets"] = dashboard["widgets"]
		}

		results = append(results, result)
	}

	return map[string]interface{}{
		"search_term": searchTerm,
		"search_type": searchType,
		"total_found": len(results),
		"dashboards":  results,
	}, nil
}

// handleGenerateDashboard generates a dashboard from template
func (s *Server) handleGenerateDashboard(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	template, ok := params["template"].(string)
	if !ok || template == "" {
		return nil, fmt.Errorf("template parameter is required")
	}

	// Generate dashboard name if not provided
	dashboardName := params["name"].(string)
	if dashboardName == "" {
		dashboardName = fmt.Sprintf("%s Dashboard - %s", strings.Title(template), time.Now().Format("2006-01-02"))
	}

	var dashboard map[string]interface{}
	var err error

	switch template {
	case "golden-signals":
		serviceName, ok := params["service_name"].(string)
		if !ok || serviceName == "" {
			return nil, fmt.Errorf("service_name is required for golden-signals template")
		}
		dashboard, err = s.generateGoldenSignalsDashboard(dashboardName, serviceName)

	case "sli-slo":
		sliConfig, ok := params["sli_config"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("sli_config is required for sli-slo template")
		}
		dashboard, err = s.generateSLISLODashboard(dashboardName, sliConfig)

	case "infrastructure":
		hostPattern := params["host_pattern"].(string)
		if hostPattern == "" {
			hostPattern = "*"
		}
		dashboard, err = s.generateInfrastructureDashboard(dashboardName, hostPattern)

	case "custom":
		customConfig, ok := params["custom_config"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("custom_config is required for custom template")
		}
		dashboard, err = s.generateCustomDashboard(dashboardName, customConfig)

	default:
		return nil, fmt.Errorf("unsupported template: %s", template)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate dashboard: %w", err)
	}

	return dashboard, nil
}

// handleListDashboards lists all dashboards
func (s *Server) handleListDashboards(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	filter := ""
	if f, ok := params["filter"].(string); ok {
		filter = f
	}

	limit := 50
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	includeMetadata, _ := params["include_metadata"].(bool)

	// TODO: Implement actual dashboard listing using New Relic API
	// For now, return mock data
	dashboards := []map[string]interface{}{
		{
			"id":          "dashboard-1",
			"name":        "Application Performance",
			"description": "Key metrics for application monitoring",
			"created_at":  time.Now().Add(-7 * 24 * time.Hour),
			"updated_at":  time.Now().Add(-2 * time.Hour),
		},
	}

	// Apply filter
	if filter != "" {
		filtered := []map[string]interface{}{}
		for _, d := range dashboards {
			if strings.Contains(strings.ToLower(d["name"].(string)), strings.ToLower(filter)) {
				filtered = append(filtered, d)
			}
		}
		dashboards = filtered
	}

	// Limit results
	if len(dashboards) > limit {
		dashboards = dashboards[:limit]
	}

	result := map[string]interface{}{
		"total":      len(dashboards),
		"dashboards": dashboards,
	}

	if includeMetadata {
		result["metadata"] = map[string]interface{}{
			"filter":        filter,
			"limit":         limit,
			"has_more":      len(dashboards) == limit,
			"retrieved_at":  time.Now(),
		}
	}

	return result, nil
}

// handleGetDashboard gets detailed dashboard information
func (s *Server) handleGetDashboard(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardID, ok := params["dashboard_id"].(string)
	if !ok || dashboardID == "" {
		return nil, fmt.Errorf("dashboard_id parameter is required")
	}

	includeQueries, _ := params["include_queries"].(bool)

	// TODO: Implement actual dashboard fetching using New Relic API
	// For now, return mock data
	dashboard := map[string]interface{}{
		"id":          dashboardID,
		"name":        "Sample Dashboard",
		"description": "Dashboard description",
		"pages": []map[string]interface{}{
			{
				"name": "Page 1",
				"widgets": []map[string]interface{}{
					{
						"title": "Request Rate",
						"type":  "line",
						"query": "SELECT rate(count(*), 1 minute) FROM Transaction TIMESERIES",
					},
				},
			},
		},
	}

	if !includeQueries {
		// Remove queries from widgets
		for _, page := range dashboard["pages"].([]map[string]interface{}) {
			for _, widget := range page["widgets"].([]map[string]interface{}) {
				delete(widget, "query")
			}
		}
	}

	return dashboard, nil
}

// Template generation functions

func (s *Server) generateGoldenSignalsDashboard(name, serviceName string) (map[string]interface{}, error) {
	dashboard := map[string]interface{}{
		"name":        name,
		"description": fmt.Sprintf("Golden signals dashboard for %s", serviceName),
		"pages": []map[string]interface{}{
			{
				"name": "Golden Signals",
				"widgets": []map[string]interface{}{
					// Latency widget
					{
						"title": "Latency (Response Time)",
						"type":  "line",
						"row":   1,
						"column": 1,
						"width":  6,
						"height": 3,
						"query": fmt.Sprintf(
							"SELECT average(duration) as 'Average', percentile(duration, 95) as 'P95', percentile(duration, 99) as 'P99' FROM Transaction WHERE appName = '%s' TIMESERIES",
							serviceName,
						),
					},
					// Traffic widget
					{
						"title": "Traffic (Request Rate)",
						"type":  "line",
						"row":   1,
						"column": 7,
						"width":  6,
						"height": 3,
						"query": fmt.Sprintf(
							"SELECT rate(count(*), 1 minute) as 'Requests/min' FROM Transaction WHERE appName = '%s' TIMESERIES",
							serviceName,
						),
					},
					// Errors widget
					{
						"title": "Errors",
						"type":  "line",
						"row":   4,
						"column": 1,
						"width":  6,
						"height": 3,
						"query": fmt.Sprintf(
							"SELECT count(*) as 'Total Errors', percentage(count(*), WHERE error IS true) as 'Error Rate' FROM Transaction WHERE appName = '%s' TIMESERIES",
							serviceName,
						),
					},
					// Saturation widget (CPU)
					{
						"title": "Saturation (CPU Usage)",
						"type":  "line",
						"row":   4,
						"column": 7,
						"width":  6,
						"height": 3,
						"query": fmt.Sprintf(
							"SELECT average(cpuPercent) FROM SystemSample WHERE apmApplicationNames LIKE '%%%s%%' TIMESERIES",
							serviceName,
						),
					},
				},
			},
		},
		"created_at": time.Now(),
		"template":   "golden-signals",
	}

	return dashboard, nil
}

func (s *Server) generateSLISLODashboard(name string, sliConfig map[string]interface{}) (map[string]interface{}, error) {
	// Extract SLI configuration
	sliName := sliConfig["name"].(string)
	sloTarget := sliConfig["target"].(float64)
	query := sliConfig["query"].(string)

	dashboard := map[string]interface{}{
		"name":        name,
		"description": fmt.Sprintf("SLI/SLO dashboard for %s", sliName),
		"pages": []map[string]interface{}{
			{
				"name": "SLI/SLO Overview",
				"widgets": []map[string]interface{}{
					// Current SLI value
					{
						"title":  fmt.Sprintf("Current %s", sliName),
						"type":   "billboard",
						"row":    1,
						"column": 1,
						"width":  4,
						"height": 3,
						"query":  query,
						"thresholds": []map[string]interface{}{
							{"value": sloTarget, "severity": "success"},
							{"value": sloTarget * 0.95, "severity": "warning"},
							{"value": 0, "severity": "critical"},
						},
					},
					// SLI over time
					{
						"title":  fmt.Sprintf("%s Over Time", sliName),
						"type":   "line",
						"row":    1,
						"column": 5,
						"width":  8,
						"height": 3,
						"query":  fmt.Sprintf("%s TIMESERIES", query),
					},
					// Error budget
					{
						"title":  "Error Budget Remaining",
						"type":   "billboard",
						"row":    4,
						"column": 1,
						"width":  4,
						"height": 3,
						"query":  fmt.Sprintf("SELECT 100 * (1 - (1 - %f) * count(*) / count(*)) as 'Error Budget %%' FROM (%s)", sloTarget/100, query),
					},
				},
			},
		},
		"created_at": time.Now(),
		"template":   "sli-slo",
	}

	return dashboard, nil
}

func (s *Server) generateInfrastructureDashboard(name, hostPattern string) (map[string]interface{}, error) {
	whereClause := ""
	if hostPattern != "*" {
		whereClause = fmt.Sprintf("WHERE hostname LIKE '%s'", hostPattern)
	}

	dashboard := map[string]interface{}{
		"name":        name,
		"description": fmt.Sprintf("Infrastructure dashboard for hosts: %s", hostPattern),
		"pages": []map[string]interface{}{
			{
				"name": "Infrastructure Overview",
				"widgets": []map[string]interface{}{
					// CPU usage
					{
						"title":  "CPU Usage by Host",
						"type":   "line",
						"row":    1,
						"column": 1,
						"width":  6,
						"height": 3,
						"query":  fmt.Sprintf("SELECT average(cpuPercent) FROM SystemSample %s FACET hostname TIMESERIES", whereClause),
					},
					// Memory usage
					{
						"title":  "Memory Usage by Host",
						"type":   "line",
						"row":    1,
						"column": 7,
						"width":  6,
						"height": 3,
						"query":  fmt.Sprintf("SELECT average(memoryUsedPercent) FROM SystemSample %s FACET hostname TIMESERIES", whereClause),
					},
					// Disk usage
					{
						"title":  "Disk Usage",
						"type":   "table",
						"row":    4,
						"column": 1,
						"width":  6,
						"height": 3,
						"query":  fmt.Sprintf("SELECT average(diskUsedPercent) as 'Disk Used %%', max(diskUsedPercent) as 'Max %%' FROM SystemSample %s FACET hostname, diskPath", whereClause),
					},
					// Network I/O
					{
						"title":  "Network I/O",
						"type":   "line",
						"row":    4,
						"column": 7,
						"width":  6,
						"height": 3,
						"query":  fmt.Sprintf("SELECT average(receiveBytesPerSecond + transmitBytesPerSecond) as 'Total Bytes/sec' FROM NetworkSample %s FACET hostname TIMESERIES", whereClause),
					},
				},
			},
		},
		"created_at": time.Now(),
		"template":   "infrastructure",
	}

	return dashboard, nil
}

func (s *Server) generateCustomDashboard(name string, config map[string]interface{}) (map[string]interface{}, error) {
	// Validate custom config has required fields
	pages, ok := config["pages"].([]interface{})
	if !ok || len(pages) == 0 {
		return nil, fmt.Errorf("custom_config must include 'pages' array")
	}

	dashboard := map[string]interface{}{
		"name":        name,
		"description": config["description"],
		"pages":       pages,
		"created_at":  time.Now(),
		"template":    "custom",
	}

	// Validate widget structure
	for i, page := range pages {
		pageMap, ok := page.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("page %d must be an object", i)
		}

		widgets, ok := pageMap["widgets"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("page %d must have 'widgets' array", i)
		}

		for j, widget := range widgets {
			widgetMap, ok := widget.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("widget %d in page %d must be an object", j, i)
			}

			// Validate required widget fields
			if _, ok := widgetMap["title"]; !ok {
				return nil, fmt.Errorf("widget %d in page %d missing 'title'", j, i)
			}
			if _, ok := widgetMap["query"]; !ok {
				return nil, fmt.Errorf("widget %d in page %d missing 'query'", j, i)
			}
		}
	}

	return dashboard, nil
}

// Helper function to search dashboards
func (s *Server) searchDashboards(ctx context.Context, searchTerm, searchType string) ([]map[string]interface{}, error) {
	// TODO: Implement actual dashboard search using New Relic API
	// For now, return mock data
	mockDashboards := []map[string]interface{}{
		{
			"id":          "dash-1",
			"name":        "Application Performance",
			"match_count": 3,
			"updated_at":  time.Now().Add(-24 * time.Hour),
			"widgets": []map[string]interface{}{
				{
					"title": "Transaction Duration",
					"query": "SELECT average(duration) FROM Transaction",
				},
			},
		},
	}

	return mockDashboards, nil
}
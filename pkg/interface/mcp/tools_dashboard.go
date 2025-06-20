//go:build !test

package mcp

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
)

// Dashboard limits as per NerdGraph documentation
const (
	MaxPagesPerDashboard = 25
	MaxWidgetsPerPage    = 150
	MaxDashboardColumns  = 12
	MaxWidgetHeight      = 32
	MaxDashboardNameLen  = 255
	MaxDashboardDescLen  = 1024
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
					Description: "Template name: 'golden-signals', 'sli-slo', 'infrastructure', 'custom', 'discovery-based'",
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
				"domain": {
					Type:        "string",
					Description: "Domain for discovery-based template (e.g. 'kafka', 'redis', 'mysql')",
				},
				"account_ids": {
					Type:        "array",
					Description: "List of account IDs for cross-account dashboards",
					Items: &Property{
						Type: "integer",
					},
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
				"account_id": {
					Type:        "string",
					Description: "Optional account ID to query (uses default if not provided)",
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

	// Update dashboard
	s.tools.Register(Tool{
		Name:        "update_dashboard",
		Description: "Update an existing dashboard",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_id", "updates"},
			Properties: map[string]Property{
				"dashboard_id": {
					Type:        "string",
					Description: "Dashboard GUID to update",
				},
				"updates": {
					Type:        "object",
					Description: "Dashboard updates (name, description, pages, permissions)",
				},
			},
		},
		Handler: s.handleUpdateDashboard,
	})

	// Delete dashboard
	s.tools.Register(Tool{
		Name:        "delete_dashboard",
		Description: "Delete a dashboard",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_id"},
			Properties: map[string]Property{
				"dashboard_id": {
					Type:        "string",
					Description: "Dashboard GUID to delete",
				},
			},
		},
		Handler: s.handleDeleteDashboard,
	})

	// Undelete dashboard
	s.tools.Register(Tool{
		Name:        "undelete_dashboard",
		Description: "Restore a previously deleted dashboard",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_id"},
			Properties: map[string]Property{
				"dashboard_id": {
					Type:        "string",
					Description: "Dashboard GUID to restore",
				},
			},
		},
		Handler: s.handleUndeleteDashboard,
	})

	// Create dashboard snapshot URL
	s.tools.Register(Tool{
		Name:        "create_dashboard_snapshot",
		Description: "Create a public URL for a static dashboard snapshot",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_id"},
			Properties: map[string]Property{
				"dashboard_id": {
					Type:        "string",
					Description: "Dashboard GUID to create snapshot for",
				},
			},
		},
		Handler: s.handleCreateDashboardSnapshot,
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
	dashboardName := ""
	if name, ok := params["name"].(string); ok {
		dashboardName = name
	}
	if dashboardName == "" {
		dashboardName = fmt.Sprintf("%s Dashboard - %s", strings.Title(template), time.Now().Format("2006-01-02"))
	}

	// Extract account IDs for cross-account support
	var accountIDs []int
	if accountIDsParam, ok := params["account_ids"].([]interface{}); ok {
		for _, id := range accountIDsParam {
			if idFloat, ok := id.(float64); ok {
				accountIDs = append(accountIDs, int(idFloat))
			}
		}
	}

	var dashboard map[string]interface{}
	var err error

	switch template {
	case "golden-signals":
		serviceName, ok := params["service_name"].(string)
		if !ok || serviceName == "" {
			return nil, fmt.Errorf("service_name is required for golden-signals template")
		}
		dashboard, err = s.generateGoldenSignalsDashboard(dashboardName, serviceName, accountIDs)

	case "sli-slo":
		sliConfig, ok := params["sli_config"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("sli_config is required for sli-slo template")
		}
		dashboard, err = s.generateSLISLODashboard(dashboardName, sliConfig, accountIDs)

	case "infrastructure":
		hostPattern := ""
		if hp, ok := params["host_pattern"].(string); ok {
			hostPattern = hp
		}
		if hostPattern == "" {
			hostPattern = "*"
		}
		dashboard, err = s.generateInfrastructureDashboard(dashboardName, hostPattern, accountIDs)

	case "custom":
		customConfig, ok := params["custom_config"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("custom_config is required for custom template")
		}
		dashboard, err = s.generateCustomDashboard(dashboardName, customConfig)

	case "discovery-based":
		// Discovery-based dashboard generation
		request := map[string]interface{}{
			"name":         dashboardName,
			"domain":       params["domain"],
			"service_name": params["service_name"],
			"account_ids":  accountIDs,
		}
		dashboard, err = s.generateDiscoveryBasedDashboard(ctx, dashboardName, request)

	default:
		return nil, fmt.Errorf("unsupported template: %s", template)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate dashboard: %w", err)
	}

	// Actually create the dashboard in New Relic
	nrClient := s.getNRClient()
	if nrClient == nil {
		// Return just the config if no NR client
		return map[string]interface{}{
			"dashboard": dashboard,
			"created":   false,
			"message":   "Dashboard configuration generated (no New Relic client configured)",
		}, nil
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Convert map to Dashboard struct for creation
	dashboardData := newrelic.Dashboard{
		Name:        dashboard["name"].(string),
		Permissions: "PUBLIC_READ_WRITE", // Default permission
		Pages:       []newrelic.DashboardPage{},
	}

	if desc, ok := dashboard["description"].(string); ok {
		dashboardData.Description = desc
	}

	// Convert pages
	if pages, ok := dashboard["pages"].([]interface{}); ok {
		for _, p := range pages {
			pageMap := p.(map[string]interface{})
			page := newrelic.DashboardPage{
				Name:    pageMap["name"].(string),
				Widgets: []newrelic.DashboardWidget{},
			}

			// Convert widgets
			if widgets, ok := pageMap["widgets"].([]interface{}); ok {
				for _, w := range widgets {
					widgetMap := w.(map[string]interface{})
					widget := newrelic.DashboardWidget{
						Title: widgetMap["title"].(string),
						Type:  widgetMap["type"].(string),
						Query: widgetMap["query"].(string),
						Configuration: map[string]interface{}{
							"row":    widgetMap["row"],
							"column": widgetMap["column"],
							"width":  widgetMap["width"],
							"height": widgetMap["height"],
						},
					}
					page.Widgets = append(page.Widgets, widget)
				}
			}
			dashboardData.Pages = append(dashboardData.Pages, page)
		}
	}

	// Create the dashboard
	created, err := client.CreateDashboard(ctx, dashboardData)
	if err != nil {
		// Return the config even if creation fails
		return map[string]interface{}{
			"dashboard": dashboard,
			"created":   false,
			"error":     err.Error(),
			"message":   "Dashboard configuration generated but creation failed",
		}, nil
	}

	// Generate the dashboard URL
	dashboardURL := fmt.Sprintf("https://one.newrelic.com/dashboards/%s", created.ID)

	return map[string]interface{}{
		"dashboard":     created,
		"created":       true,
		"dashboard_id":  created.ID,
		"dashboard_url": dashboardURL,
		"message":       fmt.Sprintf("Dashboard '%s' created successfully", dashboardName),
	}, nil
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
	accountID, _ := params["account_id"].(string)

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("list_dashboards", params), nil
	}

	// Get New Relic client with account support
	nrClient, err := s.getNRClientWithAccount(accountID)
	if err != nil {
		return nil, err
	}

	// Use reflection to call ListDashboards method
	clientValue := reflect.ValueOf(nrClient)
	var method reflect.Value
	if filter != "" {
		method = clientValue.MethodByName("SearchDashboards")
	} else {
		method = clientValue.MethodByName("ListDashboards")
	}

	if !method.IsValid() {
		return nil, fmt.Errorf("dashboard method not found on client")
	}

	// Call the method
	var results []reflect.Value
	if filter != "" {
		args := []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(filter),
		}
		results = method.Call(args)
	} else {
		args := []reflect.Value{
			reflect.ValueOf(ctx),
		}
		results = method.Call(args)
	}

	if len(results) != 2 {
		return nil, fmt.Errorf("unexpected return values from dashboard method")
	}

	// Extract error
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	// Extract dashboard list from results
	dashboardListValue := results[0]
	if dashboardListValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice of dashboards, got %v", dashboardListValue.Kind())
	}

	// Convert to response format
	dashboards := make([]map[string]interface{}, 0, dashboardListValue.Len())
	for i := 0; i < dashboardListValue.Len(); i++ {
		d := dashboardListValue.Index(i)
		// Use reflection to extract fields
		getField := func(name string) interface{} {
			f := d.FieldByName(name)
			if f.IsValid() {
				return f.Interface()
			}
			return nil
		}

		dashboard := map[string]interface{}{
			"id":          getField("ID"),
			"name":        getField("Name"),
			"created_at":  getField("CreatedAt"),
			"updated_at":  getField("UpdatedAt"),
			"permissions": getField("Permissions"),
		}
		dashboards = append(dashboards, dashboard)
	}

	// Apply limit
	if limit > 0 && len(dashboards) > limit {
		dashboards = dashboards[:limit]
	}

	result := map[string]interface{}{
		"total":      len(dashboards),
		"dashboards": dashboards,
	}

	if includeMetadata {
		result["metadata"] = map[string]interface{}{
			"filter":       filter,
			"limit":        limit,
			"has_more":     len(dashboards) == limit,
			"retrieved_at": time.Now(),
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

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Get the dashboard
	dash, err := client.GetDashboard(ctx, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("get dashboard: %w", err)
	}

	// Convert to response format
	dashboard := map[string]interface{}{
		"id":          dash.ID,
		"guid":        dash.GUID,
		"name":        dash.Name,
		"description": dash.Description,
		"permissions": dash.Permissions,
		"created_at":  dash.CreatedAt,
		"updated_at":  dash.UpdatedAt,
		"account_id":  dash.AccountID,
		"pages":       []map[string]interface{}{},
	}

	// Add pages and widgets
	for _, page := range dash.Pages {
		pageData := map[string]interface{}{
			"name":    page.Name,
			"widgets": []map[string]interface{}{},
		}

		for _, widget := range page.Widgets {
			widgetData := map[string]interface{}{
				"title":         widget.Title,
				"type":          widget.Type,
				"configuration": widget.Configuration,
			}

			// Extract query from raw configuration if requested
			if includeQueries {
				if rawConfig, ok := widget.Configuration["rawConfiguration"].(map[string]interface{}); ok {
					if nrqlQueries, ok := rawConfig["nrqlQueries"].([]interface{}); ok && len(nrqlQueries) > 0 {
						if query, ok := nrqlQueries[0].(map[string]interface{}); ok {
							widgetData["query"] = query["query"]
						}
					}
				}
			}

			pageData["widgets"] = append(pageData["widgets"].([]map[string]interface{}), widgetData)
		}

		dashboard["pages"] = append(dashboard["pages"].([]map[string]interface{}), pageData)
	}

	return dashboard, nil
}

// Template generation functions

func (s *Server) generateGoldenSignalsDashboard(name, serviceName string, accountIDs []int) (map[string]interface{}, error) {
	// Build account IDs clause for cross-account queries
	accountClause := ""
	if len(accountIDs) > 0 {
		accountIDStrs := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			accountIDStrs[i] = fmt.Sprintf("%d", id)
		}
		accountClause = fmt.Sprintf(" WITH accountIds = [%s]", strings.Join(accountIDStrs, ", "))
	}

	dashboard := map[string]interface{}{
		"name":        name,
		"description": fmt.Sprintf("Golden signals dashboard for %s", serviceName),
		"pages": []map[string]interface{}{
			{
				"name": "Golden Signals",
				"widgets": []map[string]interface{}{
					// Latency widget with sliding window for smoother visualization
					{
						"title":  "Latency (Response Time)",
						"type":   "line",
						"row":    1,
						"column": 1,
						"width":  6,
						"height": 3,
						"query": fmt.Sprintf(
							"SELECT average(duration) as 'Average', percentile(duration, 95) as 'P95', percentile(duration, 99) as 'P99' FROM Transaction WHERE appName = '%s' TIMESERIES 1 minute SLIDE BY 30 seconds%s",
							serviceName,
							accountClause,
						),
					},
					// Traffic widget using rate() function with sliding window
					{
						"title":  "Traffic (Request Rate)",
						"type":   "line",
						"row":    1,
						"column": 7,
						"width":  6,
						"height": 3,
						"query": fmt.Sprintf(
							"SELECT rate(count(*), 1 minute) as 'Requests/min' FROM Transaction WHERE appName = '%s' TIMESERIES 1 minute SLIDE BY 30 seconds%s",
							serviceName,
							accountClause,
						),
					},
					// Errors widget
					{
						"title":  "Errors",
						"type":   "line",
						"row":    4,
						"column": 1,
						"width":  6,
						"height": 3,
						"query": fmt.Sprintf(
							"SELECT count(*) as 'Total Errors', percentage(count(*), WHERE error IS true) as 'Error Rate' FROM Transaction WHERE appName = '%s' TIMESERIES%s",
							serviceName,
							accountClause,
						),
					},
					// Saturation widget (CPU)
					{
						"title":  "Saturation (CPU Usage)",
						"type":   "line",
						"row":    4,
						"column": 7,
						"width":  6,
						"height": 3,
						"query": fmt.Sprintf(
							"SELECT average(cpuPercent) FROM SystemSample WHERE apmApplicationNames LIKE '%%%s%%' TIMESERIES%s",
							serviceName,
							accountClause,
						),
					},
				},
			},
		},
		"created_at": time.Now(),
		"template":   "golden-signals",
	}

	// Add account IDs to metadata if specified
	if len(accountIDs) > 0 {
		dashboard["account_ids"] = accountIDs
	}

	return dashboard, nil
}

func (s *Server) generateSLISLODashboard(name string, sliConfig map[string]interface{}, accountIDs []int) (map[string]interface{}, error) {
	// Extract SLI configuration
	sliName, ok := sliConfig["name"].(string)
	if !ok || sliName == "" {
		return nil, fmt.Errorf("sli_config.name is required")
	}

	sloTarget, ok := sliConfig["target"].(float64)
	if !ok {
		return nil, fmt.Errorf("sli_config.target is required and must be a number")
	}

	query, ok := sliConfig["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("sli_config.query is required")
	}

	// Build account IDs clause for cross-account queries
	accountClause := ""
	if len(accountIDs) > 0 {
		accountIDStrs := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			accountIDStrs[i] = fmt.Sprintf("%d", id)
		}
		accountClause = fmt.Sprintf(" WITH accountIds = [%s]", strings.Join(accountIDStrs, ", "))
	}

	// Append account clause to the base query if provided
	queryWithAccounts := query
	if accountClause != "" {
		queryWithAccounts = query + accountClause
	}

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
						"query":  queryWithAccounts,
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
						"query":  fmt.Sprintf("%s TIMESERIES%s", query, accountClause),
					},
					// Error budget
					{
						"title":  "Error Budget Remaining",
						"type":   "billboard",
						"row":    4,
						"column": 1,
						"width":  4,
						"height": 3,
						"query":  fmt.Sprintf("SELECT 100 * (1 - (1 - %f) * count(*) / count(*)) as 'Error Budget %%' FROM (%s)%s", sloTarget/100, query, accountClause),
					},
				},
			},
		},
		"created_at": time.Now(),
		"template":   "sli-slo",
	}

	// Add account IDs to metadata if specified
	if len(accountIDs) > 0 {
		dashboard["account_ids"] = accountIDs
	}

	return dashboard, nil
}

func (s *Server) generateInfrastructureDashboard(name, hostPattern string, accountIDs []int) (map[string]interface{}, error) {
	whereClause := ""
	if hostPattern != "*" {
		whereClause = fmt.Sprintf("WHERE hostname LIKE '%s'", hostPattern)
	}

	// Build account IDs clause for cross-account queries
	accountClause := ""
	if len(accountIDs) > 0 {
		accountIDStrs := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			accountIDStrs[i] = fmt.Sprintf("%d", id)
		}
		accountClause = fmt.Sprintf(" WITH accountIds = [%s]", strings.Join(accountIDStrs, ", "))
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
						"query":  fmt.Sprintf("SELECT average(cpuPercent) FROM SystemSample %s FACET hostname TIMESERIES%s", whereClause, accountClause),
					},
					// Memory usage
					{
						"title":  "Memory Usage by Host",
						"type":   "line",
						"row":    1,
						"column": 7,
						"width":  6,
						"height": 3,
						"query":  fmt.Sprintf("SELECT average(memoryUsedPercent) FROM SystemSample %s FACET hostname TIMESERIES%s", whereClause, accountClause),
					},
					// Disk usage
					{
						"title":  "Disk Usage",
						"type":   "table",
						"row":    4,
						"column": 1,
						"width":  6,
						"height": 3,
						"query":  fmt.Sprintf("SELECT average(diskUsedPercent) as 'Disk Used %%', max(diskUsedPercent) as 'Max %%' FROM SystemSample %s FACET hostname, diskPath%s", whereClause, accountClause),
					},
					// Network I/O
					{
						"title":  "Network I/O",
						"type":   "line",
						"row":    4,
						"column": 7,
						"width":  6,
						"height": 3,
						"query":  fmt.Sprintf("SELECT average(receiveBytesPerSecond + transmitBytesPerSecond) as 'Total Bytes/sec' FROM NetworkSample %s FACET hostname TIMESERIES%s", whereClause, accountClause),
					},
				},
			},
		},
		"created_at": time.Now(),
		"template":   "infrastructure",
	}

	// Add account IDs to metadata if specified
	if len(accountIDs) > 0 {
		dashboard["account_ids"] = accountIDs
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
	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Search dashboards by name
	dashboards, err := client.SearchDashboards(ctx, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("search dashboards: %w", err)
	}

	// For more detailed search (by metric/attribute), we need to fetch each dashboard
	results := []map[string]interface{}{}

	for _, dash := range dashboards {
		// For basic name search, just include the dashboard
		if searchType == "dashboard_name" || strings.Contains(strings.ToLower(dash.Name), strings.ToLower(searchTerm)) {
			results = append(results, map[string]interface{}{
				"id":          dash.ID,
				"name":        dash.Name,
				"match_count": 1, // Name match
				"updated_at":  dash.UpdatedAt,
			})
			continue
		}

		// For metric/attribute search, we need to fetch full dashboard details
		if searchType == "metric" || searchType == "attribute" || searchType == "any" {
			fullDash, err := client.GetDashboard(ctx, dash.ID)
			if err != nil {
				// Skip if we can't get details
				continue
			}

			matchCount := 0
			matchingWidgets := []map[string]interface{}{}

			// Search through all widgets
			for _, page := range fullDash.Pages {
				for _, widget := range page.Widgets {
					// Extract query from widget configuration
					query := ""
					if rawConfig, ok := widget.Configuration["rawConfiguration"].(map[string]interface{}); ok {
						if nrqlQueries, ok := rawConfig["nrqlQueries"].([]interface{}); ok && len(nrqlQueries) > 0 {
							if q, ok := nrqlQueries[0].(map[string]interface{}); ok {
								query = q["query"].(string)
							}
						}
					}

					// Check if query contains search term
					if query != "" && strings.Contains(strings.ToLower(query), strings.ToLower(searchTerm)) {
						matchCount++
						matchingWidgets = append(matchingWidgets, map[string]interface{}{
							"title": widget.Title,
							"query": query,
						})
					}
				}
			}

			if matchCount > 0 {
				result := map[string]interface{}{
					"id":          dash.ID,
					"name":        dash.Name,
					"match_count": matchCount,
					"updated_at":  dash.UpdatedAt,
				}
				if len(matchingWidgets) > 0 {
					result["widgets"] = matchingWidgets
				}
				results = append(results, result)
			}
		}
	}

	// Sort by match count (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i]["match_count"].(int) > results[j]["match_count"].(int)
	})

	return results, nil
}

// handleUpdateDashboard updates an existing dashboard
func (s *Server) handleUpdateDashboard(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardID, ok := params["dashboard_id"].(string)
	if !ok || dashboardID == "" {
		return nil, fmt.Errorf("dashboard_id parameter is required")
	}

	updates, ok := params["updates"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("updates parameter is required")
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Get existing dashboard first
	existing, err := client.GetDashboard(ctx, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("get existing dashboard: %w", err)
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		if len(name) > MaxDashboardNameLen {
			return nil, fmt.Errorf("dashboard name exceeds maximum length of %d characters", MaxDashboardNameLen)
		}
		existing.Name = name
	}

	if desc, ok := updates["description"].(string); ok {
		if len(desc) > MaxDashboardDescLen {
			return nil, fmt.Errorf("dashboard description exceeds maximum length of %d characters", MaxDashboardDescLen)
		}
		existing.Description = desc
	}

	if perm, ok := updates["permissions"].(string); ok {
		existing.Permissions = perm
	}

	// If pages are provided, validate them
	if pages, ok := updates["pages"].([]interface{}); ok {
		if len(pages) > MaxPagesPerDashboard {
			return nil, fmt.Errorf("dashboard cannot have more than %d pages", MaxPagesPerDashboard)
		}
		// Convert and validate pages
		dashPages := make([]newrelic.DashboardPage, len(pages))
		for i, p := range pages {
			page, ok := p.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid page at index %d", i)
			}
			if err := validateDashboardPage(page); err != nil {
				return nil, fmt.Errorf("page %d: %w", i, err)
			}
			// Convert page structure
			dashPages[i] = convertToDashboardPage(page)
		}
		existing.Pages = dashPages
	}

	// Update the dashboard
	updated, err := client.UpdateDashboard(ctx, dashboardID, *existing)
	if err != nil {
		return nil, fmt.Errorf("update dashboard: %w", err)
	}

	return map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":          updated.ID,
			"name":        updated.Name,
			"permissions": updated.Permissions,
			"updated_at":  updated.UpdatedAt,
		},
		"message": "Dashboard updated successfully",
	}, nil
}

// handleDeleteDashboard deletes a dashboard
func (s *Server) handleDeleteDashboard(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardID, ok := params["dashboard_id"].(string)
	if !ok || dashboardID == "" {
		return nil, fmt.Errorf("dashboard_id parameter is required")
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Delete the dashboard
	if err := client.DeleteDashboard(ctx, dashboardID); err != nil {
		return nil, fmt.Errorf("delete dashboard: %w", err)
	}

	return map[string]interface{}{
		"dashboard_id": dashboardID,
		"message":      "Dashboard deleted successfully",
	}, nil
}

// handleUndeleteDashboard restores a deleted dashboard
func (s *Server) handleUndeleteDashboard(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardID, ok := params["dashboard_id"].(string)
	if !ok || dashboardID == "" {
		return nil, fmt.Errorf("dashboard_id parameter is required")
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Restore the dashboard
	restored, err := client.UndeleteDashboard(ctx, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("undelete dashboard: %w", err)
	}

	return map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":          restored.ID,
			"name":        restored.Name,
			"permissions": restored.Permissions,
			"created_at":  restored.CreatedAt,
			"updated_at":  restored.UpdatedAt,
		},
		"message": "Dashboard restored successfully",
	}, nil
}

// handleCreateDashboardSnapshot creates a public snapshot URL
func (s *Server) handleCreateDashboardSnapshot(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardID, ok := params["dashboard_id"].(string)
	if !ok || dashboardID == "" {
		return nil, fmt.Errorf("dashboard_id parameter is required")
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Create snapshot URL
	url, err := client.CreateDashboardSnapshotUrl(ctx, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("create dashboard snapshot: %w", err)
	}

	return map[string]interface{}{
		"dashboard_id": dashboardID,
		"snapshot_url": url,
		"message":      "Dashboard snapshot URL created successfully. This URL will expire in 3 months.",
	}, nil
}

// validateDashboardPage validates a dashboard page configuration
func validateDashboardPage(page map[string]interface{}) error {
	name, ok := page["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("page name is required")
	}

	if len(name) > MaxDashboardNameLen {
		return fmt.Errorf("page name exceeds maximum length of %d characters", MaxDashboardNameLen)
	}

	widgets, ok := page["widgets"].([]interface{})
	if !ok {
		return fmt.Errorf("page must have widgets array")
	}

	if len(widgets) > MaxWidgetsPerPage {
		return fmt.Errorf("page has %d widgets, maximum is %d", len(widgets), MaxWidgetsPerPage)
	}

	// Validate each widget
	for i, w := range widgets {
		widget, ok := w.(map[string]interface{})
		if !ok {
			return fmt.Errorf("widget %d is invalid", i)
		}
		if err := validateWidget(widget); err != nil {
			return fmt.Errorf("widget %d: %w", i, err)
		}
	}

	return nil
}

// validateWidget validates widget configuration against limits
func validateWidget(widget map[string]interface{}) error {
	// Validate title
	title, ok := widget["title"].(string)
	if !ok || title == "" {
		return fmt.Errorf("widget title is required")
	}
	if len(title) > MaxDashboardNameLen {
		return fmt.Errorf("widget title exceeds maximum length of %d characters", MaxDashboardNameLen)
	}

	// Validate layout if present
	if layout, ok := widget["layout"].(map[string]interface{}); ok {
		col := getIntValue(layout["column"])
		width := getIntValue(layout["width"])
		height := getIntValue(layout["height"])

		if col < 1 || col > MaxDashboardColumns {
			return fmt.Errorf("column must be between 1 and %d", MaxDashboardColumns)
		}
		if width < 1 || width > MaxDashboardColumns {
			return fmt.Errorf("width must be between 1 and %d", MaxDashboardColumns)
		}
		if col+width-1 > MaxDashboardColumns {
			return fmt.Errorf("widget extends beyond dashboard width (column + width > %d)", MaxDashboardColumns)
		}
		if height < 1 || height > MaxWidgetHeight {
			return fmt.Errorf("height must be between 1 and %d", MaxWidgetHeight)
		}
	}

	// Validate query is present
	if _, ok := widget["query"].(string); !ok {
		if config, ok := widget["configuration"].(map[string]interface{}); ok {
			if nrqlQueries, ok := config["nrqlQueries"].([]interface{}); !ok || len(nrqlQueries) == 0 {
				return fmt.Errorf("widget must have at least one query")
			}
		} else {
			return fmt.Errorf("widget must have a query")
		}
	}

	return nil
}

// convertToDashboardPage converts a map to DashboardPage structure
func convertToDashboardPage(page map[string]interface{}) newrelic.DashboardPage {
	dashPage := newrelic.DashboardPage{
		Name: page["name"].(string),
	}

	// Note: DashboardPage doesn't have a description field in the New Relic API

	if widgets, ok := page["widgets"].([]interface{}); ok {
		dashPage.Widgets = make([]newrelic.DashboardWidget, len(widgets))
		for i, w := range widgets {
			widget := w.(map[string]interface{})
			dashPage.Widgets[i] = convertToDashboardWidget(widget)
		}
	}

	return dashPage
}

// convertToDashboardWidget converts a map to DashboardWidget structure
func convertToDashboardWidget(widget map[string]interface{}) newrelic.DashboardWidget {
	dashWidget := newrelic.DashboardWidget{
		Title:         widget["title"].(string),
		Configuration: make(map[string]interface{}),
	}

	// Set type
	if vizType, ok := widget["type"].(string); ok {
		dashWidget.Type = vizType
	} else if vizType, ok := widget["visualization"].(string); ok {
		dashWidget.Type = vizType
	}

	// Handle configuration
	if config, ok := widget["configuration"].(map[string]interface{}); ok {
		dashWidget.Configuration = config
	} else {
		// Build configuration from simple format
		if query, ok := widget["query"].(string); ok {
			dashWidget.Configuration["nrqlQueries"] = []map[string]interface{}{
				{
					"accountId": 0, // Will be set by the API
					"query":     query,
				},
			}
		}

		// Add layout
		if layout, ok := widget["layout"].(map[string]interface{}); ok {
			dashWidget.Configuration["layout"] = layout
		} else {
			// Default layout from row/column/width/height
			dashWidget.Configuration["layout"] = map[string]interface{}{
				"row":    getIntValue(widget["row"]),
				"column": getIntValue(widget["column"]),
				"width":  getIntValue(widget["width"]),
				"height": getIntValue(widget["height"]),
			}
		}
	}

	return dashWidget
}

// getIntValue safely gets an integer value from interface{}
func getIntValue(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case int64:
		return int(val)
	default:
		return 0
	}
}

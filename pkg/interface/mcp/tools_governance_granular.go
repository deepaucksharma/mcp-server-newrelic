package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	// ToolCategoryGovernance represents governance and compliance tools
	ToolCategoryGovernance ToolCategory = "governance"
)

// RegisterGovernanceGranularTools registers all platform governance and data observability tools
func (s *Server) RegisterGovernanceGranularTools() error {
	tools := []EnhancedTool{
		// Dashboard Analysis Tools
		{
			Tool: Tool{
				Name:        "dashboard.list_widgets",
				Description: "Inventory all dashboard widgets with their configurations",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]Property{
						"cursor": {
							Type:        "string",
							Description: "Pagination cursor for large result sets",
						},
						"account_id": {
							Type:        "integer",
							Description: "Filter by specific account ID",
						},
					},
				},
				Handler: s.handleDashboardListWidgets,
			},
			Category: ToolCategoryGovernance,
			Safety: SafetyMetadata{
				Level:        SafetyLevelSafe,
				DryRunSupported: false,
			},
			Performance: PerformanceMetadata{
				ExpectedLatencyMS: 5000,
				MaxLatencyMS:      30000,
				Cacheable:         true,
				CacheTTLSeconds:   300,
			},
			AIGuidance: AIGuidanceMetadata{
				UsageExamples: []string{
					"Get complete inventory of all dashboard widgets for analysis",
					"List all widgets: dashboard.list_widgets()",
					"With pagination: dashboard.list_widgets(cursor='next-page-token')",
				},
				ChainsWith: []string{
					"dashboard.classify_widgets",
					"metric.widget_usage_rank",
				},
			},
		},
		{
			Tool: Tool{
				Name:        "dashboard.classify_widgets",
				Description: "Classify widgets as dimensional-metric-based or event-NRQL-based",
				Parameters: ToolParameters{
					Type:     "object",
					Required: []string{"dashboard_guid"},
					Properties: map[string]Property{
						"dashboard_guid": {
							Type:        "string",
							Description: "Dashboard GUID to analyze",
						},
					},
				},
				Handler: s.handleDashboardClassifyWidgets,
			},
			Category: ToolCategoryGovernance,
			Safety: SafetyMetadata{
				Level: SafetyLevelSafe,
			},
			Performance: PerformanceMetadata{
				ExpectedLatencyMS: 500,
				MaxLatencyMS:      2000,
				Cacheable:         true,
				CacheTTLSeconds:   3600,
			},
			AIGuidance: AIGuidanceMetadata{
				UsageExamples: []string{
					"Understand dashboard composition for metrics adoption analysis",
					"Classify dashboard: dashboard.classify_widgets(dashboard_guid='MXxEQVNIQk9BUkR8MTIz')",
				},
				SuccessIndicators: []string{
					"Returns metric vs event widget counts",
					"Lists all metrics and event types used",
					"Provides percentage breakdown",
				},
			},
		},
		{
			Tool: Tool{
				Name:        "dashboard.find_nrdot_dashboards",
				Description: "Find dashboards using NR1 Data Explorer (NRDOT)",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]Property{
						"account_id": {
							Type:        "integer",
							Description: "Filter by account (optional)",
						},
					},
				},
				Handler: s.handleDashboardFindUsage,
			},
			Category: ToolCategoryGovernance,
		},

		// Metric Usage Analysis
		{
			Tool: Tool{
				Name:        "metric.widget_usage_rank",
				Description: "Rank metrics by their usage across dashboard widgets",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]Property{
						"limit": {
							Type:        "integer",
							Description: "Top N metrics to return",
							Default:     50,
						},
						"time_range": {
							Type:        "string",
							Description: "Analysis time window",
							Default:     "30 days",
						},
					},
				},
				Handler: s.handleMetricWidgetUsageRank,
			},
			Category: ToolCategoryGovernance,
			Performance: PerformanceMetadata{
				ExpectedLatencyMS: 2000,
				Cacheable:         true,
				CacheTTLSeconds:   900,
			},
			AIGuidance: AIGuidanceMetadata{
				UsageExamples: []string{"Identify most valuable metrics for optimization"},
				ChainsWith: []string{
					"usage.ingest_summary",
					"dashboard.classify_widgets",
				},
			},
		},

		// Ingest Analysis Tools
		{
			Tool: Tool{
				Name:        "usage.ingest_summary",
				Description: "Get total ingest volume with breakdown by source",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]Property{
						"period": {
							Type:        "string",
							Description: "Time window (e.g. '30d', '7d')",
							Default:     "30d",
						},
						"account_id": {
							Type:        "integer",
							Description: "Specific account to analyze",
						},
					},
				},
				Handler: s.handleUsageIngestSummary,
			},
			Category: ToolCategoryGovernance,
			Safety: SafetyMetadata{
				Level: SafetyLevelSafe,
			},
			Performance: PerformanceMetadata{
				ExpectedLatencyMS: 2000,
				Cacheable:         true,
				CacheTTLSeconds:   3600,
			},
			AIGuidance: AIGuidanceMetadata{
				UsageExamples: []string{
					"Understand ingest cost drivers and source distribution",
					"Monthly summary: usage.ingest_summary(period='30d')",
					"Weekly trend: usage.ingest_summary(period='7d')",
				},
				SuccessIndicators: []string{
					"Shows total bytes and GB",
					"Breaks down by OTLP, AGENT, API sources",
					"Includes percentage distribution",
				},
			},
		},
		{
			Tool: Tool{
				Name:        "usage.otlp_collectors", 
				Description: "Analyze OTEL collector ingest volumes",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]Property{
						"period": {
							Type:        "string",
							Description: "Time window for analysis",
							Default:     "30d",
						},
					},
				},
				Handler: s.handleUsageOtlpCollectors,
			},
			Category: ToolCategoryGovernance,
			AIGuidance: AIGuidanceMetadata{
				UsageExamples: []string{"Identify noisy OTEL collectors for optimization"},
				ChainsWith: []string{"usage.agent_ingest"},
			},
		},
		{
			Tool: Tool{
				Name:        "usage.agent_ingest",
				Description: "Get native agent ingest statistics",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]Property{
						"period": {
							Type:        "string",
							Description: "Time window for analysis",
							Default:     "30d",
						},
					},
				},
				Handler: s.handleUsageAgentIngest,
			},
			Category: ToolCategoryGovernance,
			AIGuidance: AIGuidanceMetadata{
				UsageExamples: []string{"Compare native agent vs OTEL ingest patterns"},
			},
		},
	}

	// Register all tools
	for _, tool := range tools {
		s.tools.Register(tool.Tool)
	}

	return nil
}

// Handler implementations

func (s *Server) handleDashboardListWidgets(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	cursor, _ := params["cursor"].(string)
	// accountID, _ := params["account_id"].(float64) // TODO: use when API supports it

	// Mock implementation for development
	if s.nrClient == nil {
		return s.mockDashboardListWidgets(cursor)
	}

	// GraphQL query to get dashboards and their widgets
	// query := `
	// 	query($cursor: String, $accountId: Int) {
	// 		actor {
	// 			entitySearch(
	// 				query: "type = 'DASHBOARD'",
	// 				cursor: $cursor
	// 			) {
	// 				results {
	// 					entities {
	// 						... on DashboardEntity {
	// 							guid
	// 							name
	// 							pages {
	// 								widgets {
	// 									id
	// 									visualization {
	// 										id
	// 									}
	// 									rawConfiguration
	// 								}
	// 							}
	// 						}
	// 					}
	// 					nextCursor
	// 				}
	// 			}
	// 		}
	// 	}
	// `

	// variables := map[string]interface{}{
	// 	"cursor": cursor,
	// }
	// if accountID > 0 {
	// 	variables["accountId"] = int(accountID)
	// }

	// For now return mock data since we don't have the full GraphQL client interface
	// TODO: Implement when newrelic.Client has QueryWithVariables method
	return s.handleDashboardListWidgetsMock(ctx, params)
}

func (s *Server) handleDashboardListWidgetsMock(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Mock implementation for dashboard list widgets
	return map[string]interface{}{
		"dashboards": []map[string]interface{}{
			{
				"guid": "MXxEQVNIQk9BUkR8MTIz",
				"name": "Application Performance",
				"widgets": []map[string]interface{}{
					{
						"title": "Transaction Time",
						"type": "line",
						"nrql": "SELECT average(duration) FROM Transaction TIMESERIES",
					},
					{
						"title": "Error Rate",
						"type": "billboard",
						"nrql": "SELECT percentage(count(*), WHERE error = true) FROM Transaction",
					},
				},
			},
		},
		"totalCount": 1,
		"nextCursor": nil,
	}, nil
}

func (s *Server) handleDashboardFindUsage(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	metricName, ok := params["metric_name"].(string)
	if !ok || metricName == "" {
		return nil, fmt.Errorf("metric_name is required")
	}

	// Mock implementation
	return map[string]interface{}{
		"dashboards": []map[string]interface{}{
			{
				"guid": "MXxEQVNIQk9BUkR8MTIz",
				"name": "Application Performance",
				"account": 12345,
				"usageCount": 3,
				"widgets": []string{
					"Response Time Trend",
					"Error Rate",
					"Throughput",
				},
			},
		},
		"totalDashboards": 1,
		"metricName": metricName,
	}, nil
}

func (s *Server) handleDashboardClassifyWidgets(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardGUID, ok := params["dashboard_guid"].(string)
	if !ok || dashboardGUID == "" {
		return nil, fmt.Errorf("dashboard_guid is required")
	}

	// Mock implementation
	if s.nrClient == nil {
		return map[string]interface{}{
			"dashboardGuid":  dashboardGUID,
			"metricWidgets":  12,
			"eventWidgets":   34,
			"metricNames":    []string{"http.server.duration", "cpu.usage", "memory.usage"},
			"eventTypes":     []string{"Transaction", "PageView", "JavaScriptError"},
			"classification": map[string]interface{}{
				"percentMetrics": 26.1,
				"percentEvents":  73.9,
			},
		}, nil
	}

	// Get dashboard details
	widgets, err := s.getDashboardWidgets(ctx, dashboardGUID)
	if err != nil {
		return nil, err
	}

	// Classify widgets
	return s.classifyWidgets(widgets), nil
}

func (s *Server) handleMetricWidgetUsageRank(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	limit, _ := params["limit"].(float64)
	if limit == 0 {
		limit = 50
	}

	// Mock implementation
	if s.nrClient == nil {
		return s.mockMetricUsageRank(int(limit))
	}

	// Get all widgets
	allWidgets, err := s.getAllDashboardWidgets(ctx)
	if err != nil {
		return nil, err
	}

	// Count metric usage
	metricCounts := make(map[string]*metricUsageInfo)
	for _, widget := range allWidgets {
		metrics := s.extractMetricsFromWidget(widget)
		for _, metric := range metrics {
			if info, exists := metricCounts[metric]; exists {
				info.Count++
				info.Dashboards = append(info.Dashboards, widget.DashboardName)
			} else {
				metricCounts[metric] = &metricUsageInfo{
					Name:       metric,
					Count:      1,
					Dashboards: []string{widget.DashboardName},
				}
			}
		}
	}

	// Sort and limit
	rankings := s.rankMetrics(metricCounts, int(limit))
	
	return map[string]interface{}{
		"rankings": rankings,
		"totalMetricsFound": len(metricCounts),
		"totalWidgetsAnalyzed": len(allWidgets),
	}, nil
}

func (s *Server) handleUsageIngestSummary(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	period, _ := params["period"].(string)
	if period == "" {
		period = "30d"
	}

	// accountID, _ := params["account_id"].(float64) // TODO: use when API supports it

	// Mock implementation
	if s.nrClient == nil {
		return map[string]interface{}{
			"totalBytes": 10995116277760, // 10TB
			"totalGB":    10240,
			"breakdown": []map[string]interface{}{
				{"source": "OTLP", "bytes": 6597069766656, "percentage": 60},
				{"source": "AGENT", "bytes": 3298534883328, "percentage": 30},
				{"source": "API", "bytes": 1099511627776, "percentage": 10},
			},
			"period": period,
		}, nil
	}

	// Convert period to timestamp range
	// endTime := time.Now()
	// startTime := s.parsePeriod(period, endTime) // TODO: use when API is implemented

	// GraphQL query for ingest usage - saved for future implementation
	_ = `
		query($accountId: Int!, $since: EpochMilliseconds!, $until: EpochMilliseconds!) {
			actor {
				account(id: $accountId) {
					nrUsage {
						ingest(since: $since, until: $until) {
							total
							byDataSource {
								dataSource
								bytes
							}
						}
					}
				}
			}
		}
	`

	// For now return mock data since we don't have the full GraphQL client interface
	// TODO: Implement when newrelic.Client has proper methods and account ID config
	return s.handleUsageIngestSummaryMock(ctx, params)
}

func (s *Server) handleUsageIngestSummaryMock(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	period, _ := params["period"].(string)
	if period == "" {
		period = "7d"
	}
	
	// Mock response
	return map[string]interface{}{
		"period": period,
		"totalGB": 125.5,
		"totalBytes": 134773825536,
		"breakdown": map[string]interface{}{
			"OTLP": map[string]interface{}{
				"bytes": 67386912768,
				"percentage": 50.0,
			},
			"AGENT": map[string]interface{}{
				"bytes": 53909530214,
				"percentage": 40.0,
			},
			"API": map[string]interface{}{
				"bytes": 13477382554,
				"percentage": 10.0,
			},
		},
	}, nil
}

func (s *Server) handleUsageOtlpCollectors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	period, _ := params["period"].(string)
	if period == "" {
		period = "30d"
	}

	// Mock implementation
	if s.nrClient == nil {
		return map[string]interface{}{
			"collectors": []map[string]interface{}{
				{
					"name":             "otel-payment-prod",
					"metricCount":      15000000,
					"bytesEstimate":    120000000,
					"percentageOfOtlp": 40,
				},
				{
					"name":             "otel-inventory-prod",
					"metricCount":      8000000,
					"bytesEstimate":    64000000,
					"percentageOfOtlp": 21,
				},
			},
			"totalOtlpBytes": 6597069766656,
			"period":         period,
		}, nil
	}

	// Find OTEL collectors - saved for future implementation
	_ = `
		SELECT 
			uniqueCount(metricName) as metricCount,
			sum(newrelic.timeslice.value) as dataPoints,
			latest(collector.name) as collectorName
		FROM Metric
		WHERE instrumentation.provider = 'otel'
		FACET collector.name
		SINCE %s ago
		LIMIT 100
	`

	// For now return mock data since we don't have the QueryNRDB method
	// TODO: Implement when newrelic.Client has proper methods
	return s.handleUsageOtlpCollectorsMock(ctx, params)
}

func (s *Server) handleUsageOtlpCollectorsMock(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	period, _ := params["period"].(string)
	if period == "" {
		period = "24h"
	}
	
	// Mock response
	return map[string]interface{}{
		"period": period,
		"collectors": []map[string]interface{}{
			{
				"name": "kubernetes-otel-collector",
				"metricCount": 156,
				"dataPoints": 892347,
				"estimatedBytes": 178469400,
			},
			{
				"name": "java-app-collector",
				"metricCount": 89,
				"dataPoints": 445673,
				"estimatedBytes": 89134600,
			},
		},
		"totalCollectors": 2,
	}, nil
}

func (s *Server) handleUsageAgentIngest(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	period, _ := params["period"].(string)
	if period == "" {
		period = "30d"
	}

	// Mock implementation
	if s.nrClient == nil {
		return map[string]interface{}{
			"agents": []map[string]interface{}{
				{"name": "Infrastructure", "bytes": 1649267441664},
				{"name": "APM", "bytes": 1099511627776},
				{"name": "Browser", "bytes": 549755813888},
			},
			"comparison": map[string]interface{}{
				"agentBytes": 3298534883328,
				"otelBytes":  6597069766656,
				"ratio":      0.5,
			},
			"period": period,
		}, nil
	}

	// Query agent ingest - saved for future implementation
	_ = `
		SELECT 
			sum(newrelic.timeslice.value) * 8 as bytesEstimate,
			latest(agent.name) as agentName
		FROM Metric
		WHERE agent.name IS NOT NULL
		FACET agent.name
		SINCE %s ago
	`

	// For now use the mock data since we don't have the QueryNRDB method
	// TODO: Implement when newrelic.Client has proper methods
	return map[string]interface{}{
		"agents": []map[string]interface{}{
			{"name": "Infrastructure", "bytes": 1649267441664},
			{"name": "APM", "bytes": 1099511627776},
			{"name": "Browser", "bytes": 549755813888},
		},
		"comparison": map[string]interface{}{
			"agentBytes": 3298534883328,
			"otelBytes":  6597069766656,
			"ratio":      0.5,
		},
		"period": period,
	}, nil
}

// Helper functions

func (s *Server) isMetricWidget(rawConfig map[string]interface{}) bool {
	// Check for metricName at top level
	if _, hasMetric := rawConfig["metricName"]; hasMetric {
		return true
	}

	// Check for metrics array
	if metrics, hasMetrics := rawConfig["metrics"].([]interface{}); hasMetrics {
		for _, m := range metrics {
			if metric, ok := m.(map[string]interface{}); ok {
				if _, hasName := metric["metricName"]; hasName {
					return true
				}
			}
		}
	}

	// Check visualization type
	if viz, hasViz := rawConfig["visualization"].(map[string]interface{}); hasViz {
		if id, hasID := viz["id"].(string); hasID && strings.Contains(id, "metric") {
			return true
		}
	}

	return false
}

func (s *Server) extractMetricsFromWidget(widget widgetInfo) []string {
	var metrics []string
	
	var rawConfig map[string]interface{}
	if err := json.Unmarshal([]byte(widget.RawConfig), &rawConfig); err != nil {
		return metrics
	}

	// Extract metricName
	if metricName, ok := rawConfig["metricName"].(string); ok {
		metrics = append(metrics, metricName)
	}

	// Extract from metrics array
	if metricsList, ok := rawConfig["metrics"].([]interface{}); ok {
		for _, m := range metricsList {
			if metric, ok := m.(map[string]interface{}); ok {
				if name, ok := metric["metricName"].(string); ok {
					metrics = append(metrics, name)
				}
			}
		}
	}

	return metrics
}

func (s *Server) extractEventTypesFromWidget(rawConfig map[string]interface{}) []string {
	var eventTypes []string
	
	// Extract from NRQL query
	if nrql, ok := rawConfig["nrql"].(string); ok {
		// Simple extraction - look for FROM clause
		parts := strings.Split(strings.ToUpper(nrql), "FROM")
		if len(parts) > 1 {
			// Extract event type (simplified - real implementation would use NRQL parser)
			eventPart := strings.TrimSpace(parts[1])
			words := strings.Fields(eventPart)
			if len(words) > 0 {
				eventTypes = append(eventTypes, words[0])
			}
		}
	}

	return eventTypes
}

func (s *Server) rankMetrics(metricCounts map[string]*metricUsageInfo, limit int) []map[string]interface{} {
	// Convert to slice for sorting
	var metrics []*metricUsageInfo
	for _, info := range metricCounts {
		metrics = append(metrics, info)
	}

	// Sort by count descending
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Count > metrics[j].Count
	})

	// Limit and format results
	var rankings []map[string]interface{}
	total := len(metrics)
	if limit > total {
		limit = total
	}

	for i := 0; i < limit; i++ {
		m := metrics[i]
		rankings = append(rankings, map[string]interface{}{
			"metricName":       m.Name,
			"widgetCount":      m.Count,
			"dashboards":       uniqueStrings(m.Dashboards),
			"percentageOfTotal": float64(m.Count) / float64(total) * 100,
		})
	}

	return rankings
}

func (s *Server) parsePeriod(period string, endTime time.Time) time.Time {
	// Parse period like "30d", "7d", "24h"
	period = strings.ToLower(period)
	
	if strings.HasSuffix(period, "d") {
		days, _ := fmt.Sscanf(period, "%dd", new(int))
		return endTime.AddDate(0, 0, -days)
	} else if strings.HasSuffix(period, "h") {
		hours, _ := fmt.Sscanf(period, "%dh", new(int))
		return endTime.Add(-time.Duration(hours) * time.Hour)
	}
	
	// Default to 30 days
	return endTime.AddDate(0, 0, -30)
}

// Mock implementations for development

func (s *Server) mockDashboardListWidgets(cursor string) (interface{}, error) {
	widgets := []map[string]interface{}{
		{
			"dashboardGuid":     "MXxEQVNIQk9BUkR8MTIzNDU",
			"dashboardName":     "Production Overview",
			"widgetId":          "widget-1",
			"type":              "line",
			"visualization":     "viz.line",
			"rawConfiguration":  `{"metricName": "http.server.duration", "title": "Response Time"}`,
		},
		{
			"dashboardGuid":     "MXxEQVNIQk9BUkR8MTIzNDU",
			"dashboardName":     "Production Overview", 
			"widgetId":          "widget-2",
			"type":              "billboard",
			"visualization":     "viz.billboard",
			"rawConfiguration":  `{"nrql": "SELECT count(*) FROM Transaction", "title": "Throughput"}`,
		},
	}

	nextCursor := ""
	if cursor == "" {
		nextCursor = "page-2"
	}

	return map[string]interface{}{
		"widgets":    widgets,
		"nextCursor": nextCursor,
		"totalCount": 150,
	}, nil
}

func (s *Server) mockMetricUsageRank(limit int) (interface{}, error) {
	rankings := []map[string]interface{}{
		{
			"metricName":        "http.server.duration",
			"widgetCount":       45,
			"dashboards":        []string{"Production Overview", "API Performance", "SLO Dashboard"},
			"percentageOfTotal": 15.2,
		},
		{
			"metricName":        "cpu.usage",
			"widgetCount":       38,
			"dashboards":        []string{"Infrastructure Health", "Capacity Planning"},
			"percentageOfTotal": 12.8,
		},
		{
			"metricName":        "memory.usage",
			"widgetCount":       32,
			"dashboards":        []string{"Infrastructure Health", "Container Monitoring"},
			"percentageOfTotal": 10.8,
		},
	}

	if limit < len(rankings) {
		rankings = rankings[:limit]
	}

	return map[string]interface{}{
		"rankings":             rankings,
		"totalMetricsFound":    296,
		"totalWidgetsAnalyzed": 2500,
	}, nil
}

// Helper types

type widgetInfo struct {
	DashboardGUID string
	DashboardName string
	WidgetID      string
	Type          string
	Visualization string
	RawConfig     string
}

type metricUsageInfo struct {
	Name       string
	Count      int
	Dashboards []string
}

func uniqueStrings(strings []string) []string {
	seen := make(map[string]bool)
	unique := []string{}
	
	for _, s := range strings {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}
	
	return unique
}

// Format helpers

func (s *Server) formatDashboardWidgets(result interface{}) interface{} {
	// Format NerdGraph response into clean structure
	// Implementation depends on actual GraphQL response structure
	return result
}

func (s *Server) classifyWidgets(widgets []widgetInfo) interface{} {
	metricCount := 0
	eventCount := 0
	metricNames := make(map[string]bool)
	eventTypes := make(map[string]bool)

	for _, widget := range widgets {
		var rawConfig map[string]interface{}
		if err := json.Unmarshal([]byte(widget.RawConfig), &rawConfig); err != nil {
			continue
		}

		if s.isMetricWidget(rawConfig) {
			metricCount++
			for _, metric := range s.extractMetricsFromWidget(widget) {
				metricNames[metric] = true
			}
		} else {
			eventCount++
			for _, event := range s.extractEventTypesFromWidget(rawConfig) {
				eventTypes[event] = true
			}
		}
	}

	total := metricCount + eventCount
	percentMetrics := 0.0
	percentEvents := 0.0
	if total > 0 {
		percentMetrics = float64(metricCount) / float64(total) * 100
		percentEvents = float64(eventCount) / float64(total) * 100
	}

	return map[string]interface{}{
		"metricWidgets": metricCount,
		"eventWidgets":  eventCount,
		"metricNames":   mapKeys(metricNames),
		"eventTypes":    mapKeys(eventTypes),
		"classification": map[string]interface{}{
			"percentMetrics": percentMetrics,
			"percentEvents":  percentEvents,
		},
	}
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *Server) formatIngestSummary(result interface{}, period string) interface{} {
	// Format NerdGraph ingest usage response
	// Implementation depends on actual response structure
	return result
}

func (s *Server) formatOtelCollectors(result interface{}, period string) interface{} {
	// Format NRQL collector query results
	return result
}

func (s *Server) formatAgentIngest(result interface{}, otelBytes int64, period string) interface{} {
	// Format agent ingest data with OTEL comparison
	return result
}

// Additional helper methods needed for full implementation
func (s *Server) getDashboardWidgets(ctx context.Context, guid string) ([]widgetInfo, error) {
	// Implementation to get widgets for specific dashboard
	return nil, nil
}

func (s *Server) getAllDashboardWidgets(ctx context.Context) ([]widgetInfo, error) {
	// Implementation to get all widgets across all dashboards
	return nil, nil
}

func (s *Server) getOtelTotalBytes(ctx context.Context, period string) (int64, error) {
	// Implementation to get total OTEL bytes
	return 0, nil
}
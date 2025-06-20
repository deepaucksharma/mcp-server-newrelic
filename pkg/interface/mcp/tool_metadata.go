package mcp

import (
	"fmt"
	"strings"
)

// ToolMetadata provides enhanced metadata for AI guidance
type ToolMetadata struct {
	// Core metadata
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	
	// Safety and performance
	SafetyLevel   string            `json:"safety_level"`    // safe, caution, destructive
	Performance   PerformanceHints  `json:"performance"`
	
	// AI guidance
	UsageExamples []UsageExample    `json:"usage_examples"`
	CommonErrors  []CommonError     `json:"common_errors"`
	BestPractices []string          `json:"best_practices"`
	Prerequisites []string          `json:"prerequisites"`
	
	// Relationships
	RelatedTools  []string          `json:"related_tools"`
	FollowUpTools []string          `json:"follow_up_tools"`
	
	// Capabilities
	Capabilities  map[string]bool   `json:"capabilities"`
	Limitations   []string          `json:"limitations"`
}

// PerformanceHints provides performance guidance
type PerformanceHints struct {
	TypicalLatency   string `json:"typical_latency"`    // e.g., "500ms", "2-5s"
	MaxLatency       string `json:"max_latency"`        // e.g., "30s"
	ResourceUsage    string `json:"resource_usage"`     // low, medium, high
	CachingSupported bool   `json:"caching_supported"`
	CacheTTL         string `json:"cache_ttl,omitempty"`
}

// UsageExample shows how to use the tool
type UsageExample struct {
	Scenario    string                 `json:"scenario"`
	Description string                 `json:"description"`
	Request     map[string]interface{} `json:"request"`
	Context     string                 `json:"context,omitempty"`
}

// CommonError describes common mistakes and solutions
type CommonError struct {
	Error       string `json:"error"`
	Cause       string `json:"cause"`
	Solution    string `json:"solution"`
	Example     string `json:"example,omitempty"`
}

// EnhanceToolMetadata adds rich metadata to tools
func EnhanceToolMetadata(tool *Tool) {
	switch tool.Name {
	case "query_nrdb":
		enhanceQueryNRDB(tool)
	case "discovery.explore_event_types":
		enhanceDiscoveryExploreEventTypes(tool)
	case "discovery.explore_attributes":
		enhanceDiscoveryExploreAttributes(tool)
	case "create_alert":
		enhanceCreateAlert(tool)
	case "list_dashboards":
		enhanceListDashboards(tool)
	case "analysis.calculate_baseline":
		enhanceAnalysisCalculateBaseline(tool)
	}
}

func enhanceQueryNRDB(tool *Tool) {
	tool.Metadata = &ToolMetadata{
		Name:        tool.Name,
		Description: tool.Description,
		Category:    "Query",
		SafetyLevel: "safe",
		Performance: PerformanceHints{
			TypicalLatency:   "500ms-2s",
			MaxLatency:       "30s",
			ResourceUsage:    "medium",
			CachingSupported: true,
			CacheTTL:         "5m",
		},
		UsageExamples: []UsageExample{
			{
				Scenario:    "Check application health",
				Description: "Query error rate and response time for an application",
				Request: map[string]interface{}{
					"query": "SELECT percentage(count(*), WHERE error IS TRUE) as 'Error Rate', average(duration) FROM Transaction WHERE appName = 'production-api' SINCE 1 hour ago",
				},
			},
			{
				Scenario:    "Find slow database queries",
				Description: "Identify database operations taking over 1 second",
				Request: map[string]interface{}{
					"query": "SELECT average(databaseDuration), count(*) FROM Transaction WHERE databaseDuration > 1 FACET databaseCallDetails SINCE 1 hour ago LIMIT 20",
				},
			},
			{
				Scenario:    "Cross-account query",
				Description: "Query data from a different account",
				Request: map[string]interface{}{
					"query":      "SELECT count(*) FROM SystemSample FACET hostname SINCE 1 hour ago",
					"account_id": "2345678",
				},
				Context: "Useful for comparing metrics across different environments or regions",
			},
		},
		CommonErrors: []CommonError{
			{
				Error:    "Syntax error near 'SINCE'",
				Cause:    "Missing time range in SINCE clause",
				Solution: "Always include a time range after SINCE (e.g., 'SINCE 1 hour ago')",
				Example:  "Wrong: SINCE\nCorrect: SINCE 1 hour ago",
			},
			{
				Error:    "Unknown attribute 'appname'",
				Cause:    "Attribute names are case-sensitive",
				Solution: "Use correct case: 'appName' not 'appname'",
			},
			{
				Error:    "Query timeout",
				Cause:    "Query is too complex or scanning too much data",
				Solution: "Add time constraints, reduce LIMIT, or simplify aggregations",
			},
		},
		BestPractices: []string{
			"Always specify a time range to limit data scanned",
			"Use LIMIT to control result size",
			"Test queries with small time ranges first",
			"Use WHERE clauses to filter data early",
			"Leverage indexes by filtering on common attributes",
		},
		Prerequisites: []string{
			"Valid NRQL syntax knowledge",
			"Understanding of available event types",
			"Knowledge of attribute names for filtering",
		},
		RelatedTools: []string{
			"discovery.explore_event_types",
			"discovery.explore_attributes",
			"query_check",
			"query_builder",
		},
		FollowUpTools: []string{
			"analysis.calculate_baseline",
			"create_alert",
			"generate_dashboard",
		},
		Capabilities: map[string]bool{
			"aggregation":     true,
			"filtering":       true,
			"grouping":        true,
			"time_series":     true,
			"multi_account":   true,
			"subqueries":      true,
			"array_functions": true,
		},
		Limitations: []string{
			"Maximum query execution time: 2 minutes",
			"Result set limited to 2000 rows without LIMIT",
			"Some functions not available in subqueries",
			"Cannot modify data (read-only)",
		},
	}
}

func enhanceDiscoveryExploreEventTypes(tool *Tool) {
	tool.Metadata = &ToolMetadata{
		Name:        tool.Name,
		Description: tool.Description,
		Category:    "Discovery",
		SafetyLevel: "safe",
		Performance: PerformanceHints{
			TypicalLatency:   "1-3s",
			MaxLatency:       "10s",
			ResourceUsage:    "low",
			CachingSupported: true,
			CacheTTL:         "1h",
		},
		UsageExamples: []UsageExample{
			{
				Scenario:    "Initial exploration",
				Description: "Discover what data is available in the account",
				Request:     map[string]interface{}{},
				Context:     "Always start here when working with a new account",
			},
			{
				Scenario:    "Find custom events",
				Description: "Look for custom instrumentation",
				Request: map[string]interface{}{
					"pattern": "Custom",
				},
			},
		},
		BestPractices: []string{
			"Run this tool first before querying data",
			"Use patterns to filter large result sets",
			"Check event volume to understand data availability",
		},
		RelatedTools: []string{
			"discovery.explore_attributes",
			"discovery.check_instrumentation",
		},
		FollowUpTools: []string{
			"discovery.explore_attributes",
			"query_nrdb",
		},
		Capabilities: map[string]bool{
			"pattern_matching": true,
			"volume_info":      true,
			"multi_account":    true,
		},
	}
}

func enhanceDiscoveryExploreAttributes(tool *Tool) {
	tool.Metadata = &ToolMetadata{
		Name:        tool.Name,
		Description: tool.Description,
		Category:    "Discovery",
		SafetyLevel: "safe",
		Performance: PerformanceHints{
			TypicalLatency:   "2-5s",
			MaxLatency:       "20s",
			ResourceUsage:    "medium",
			CachingSupported: true,
			CacheTTL:         "15m",
		},
		UsageExamples: []UsageExample{
			{
				Scenario:    "Explore transaction attributes",
				Description: "Understand what data is available for transactions",
				Request: map[string]interface{}{
					"event_type": "Transaction",
				},
			},
			{
				Scenario:    "Find custom attributes",
				Description: "Discover custom instrumentation in your app",
				Request: map[string]interface{}{
					"event_type":    "Transaction",
					"show_examples": true,
				},
				Context: "Examples help identify attribute patterns and values",
			},
		},
		BestPractices: []string{
			"Use after discovery.explore_event_types",
			"Enable show_examples to understand data formats",
			"Check coverage to identify incomplete data",
		},
		Prerequisites: []string{
			"Know the event type to explore",
		},
		RelatedTools: []string{
			"discovery.explore_event_types",
			"discovery.profile_data_completeness",
		},
		FollowUpTools: []string{
			"query_nrdb",
			"query_builder",
		},
		Capabilities: map[string]bool{
			"null_analysis":  true,
			"type_detection": true,
			"examples":       true,
			"multi_account":  true,
		},
	}
}

func enhanceCreateAlert(tool *Tool) {
	tool.Metadata = &ToolMetadata{
		Name:        tool.Name,
		Description: tool.Description,
		Category:    "Action",
		SafetyLevel: "caution",
		Performance: PerformanceHints{
			TypicalLatency: "2-5s",
			MaxLatency:     "15s",
			ResourceUsage:  "low",
		},
		UsageExamples: []UsageExample{
			{
				Scenario:    "Error rate alert",
				Description: "Alert when error rate exceeds normal levels",
				Request: map[string]interface{}{
					"name":       "High Error Rate - Production API",
					"query":      "SELECT percentage(count(*), WHERE error IS TRUE) FROM Transaction WHERE appName = 'production-api'",
					"auto_baseline": true,
					"sensitivity":   "medium",
				},
			},
			{
				Scenario:    "Response time alert",
				Description: "Alert on performance degradation",
				Request: map[string]interface{}{
					"name":             "Slow Response Time",
					"query":            "SELECT average(duration) FROM Transaction WHERE appName = 'checkout-service'",
					"comparison":       "above",
					"static_threshold": 2.0,
					"auto_baseline":    false,
				},
			},
		},
		CommonErrors: []CommonError{
			{
				Error:    "Policy not found",
				Cause:    "Invalid policy_id provided",
				Solution: "Use list_alert_policies to find valid policy IDs",
			},
			{
				Error:    "Invalid threshold",
				Cause:    "Threshold value doesn't match query output",
				Solution: "Ensure threshold matches the data type returned by query",
			},
		},
		BestPractices: []string{
			"Use auto_baseline for dynamic thresholds",
			"Test query with query_nrdb first",
			"Set appropriate sensitivity levels",
			"Include clear, actionable alert names",
		},
		Prerequisites: []string{
			"Valid alert policy must exist",
			"Query must return numeric value",
		},
		RelatedTools: []string{
			"list_alert_policies",
			"query_nrdb",
			"analysis.calculate_baseline",
		},
		FollowUpTools: []string{
			"list_alerts",
			"test_alert",
		},
		Capabilities: map[string]bool{
			"auto_baseline":   true,
			"multi_condition": true,
			"multi_account":   true,
		},
		Limitations: []string{
			"Query must return single numeric value",
			"Cannot create compound conditions",
			"Baseline calculation requires historical data",
		},
	}
}

func enhanceListDashboards(tool *Tool) {
	tool.Metadata = &ToolMetadata{
		Name:        tool.Name,
		Description: tool.Description,
		Category:    "Query",
		SafetyLevel: "safe",
		Performance: PerformanceHints{
			TypicalLatency:   "1-3s",
			MaxLatency:       "10s",
			ResourceUsage:    "low",
			CachingSupported: true,
			CacheTTL:         "5m",
		},
		UsageExamples: []UsageExample{
			{
				Scenario:    "Find all dashboards",
				Description: "List all dashboards in the account",
				Request:     map[string]interface{}{},
			},
			{
				Scenario:    "Search for specific dashboards",
				Description: "Find dashboards by name pattern",
				Request: map[string]interface{}{
					"filter": "production",
				},
			},
			{
				Scenario:    "Cross-account dashboard listing",
				Description: "List dashboards from another account",
				Request: map[string]interface{}{
					"account_id": "2345678",
					"include_metadata": true,
				},
			},
		},
		BestPractices: []string{
			"Use filters to narrow results",
			"Include metadata for additional context",
			"Check multiple accounts for shared dashboards",
		},
		RelatedTools: []string{
			"get_dashboard",
			"find_usage",
		},
		FollowUpTools: []string{
			"get_dashboard",
			"clone_dashboard",
			"update_dashboard",
		},
		Capabilities: map[string]bool{
			"filtering":      true,
			"sorting":        true,
			"pagination":     true,
			"multi_account":  true,
		},
	}
}

func enhanceAnalysisCalculateBaseline(tool *Tool) {
	tool.Metadata = &ToolMetadata{
		Name:        tool.Name,
		Description: tool.Description,
		Category:    "Analysis",
		SafetyLevel: "safe",
		Performance: PerformanceHints{
			TypicalLatency:   "3-10s",
			MaxLatency:       "30s",
			ResourceUsage:    "high",
			CachingSupported: true,
			CacheTTL:         "30m",
		},
		UsageExamples: []UsageExample{
			{
				Scenario:    "Error rate baseline",
				Description: "Calculate normal error rate patterns",
				Request: map[string]interface{}{
					"query":       "SELECT percentage(count(*), WHERE error IS TRUE) FROM Transaction WHERE appName = 'api'",
					"time_range":  "7 days",
					"sensitivity": "medium",
				},
			},
			{
				Scenario:    "Hourly pattern analysis",
				Description: "Understand hourly traffic patterns",
				Request: map[string]interface{}{
					"query":             "SELECT rate(count(*), 1 minute) FROM PageView",
					"time_range":        "30 days",
					"aggregation":       "hourly",
					"include_anomalies": true,
				},
			},
		},
		BestPractices: []string{
			"Use at least 7 days of data for accuracy",
			"Consider weekly patterns for business metrics",
			"Adjust sensitivity based on metric stability",
			"Account for known maintenance windows",
		},
		Prerequisites: []string{
			"Sufficient historical data",
			"Query returns numeric values",
		},
		RelatedTools: []string{
			"query_nrdb",
			"analysis.detect_anomalies",
		},
		FollowUpTools: []string{
			"create_alert",
			"analysis.predict_trend",
		},
		Capabilities: map[string]bool{
			"percentile_calc": true,
			"seasonality":     true,
			"confidence_bands": true,
			"anomaly_detection": true,
			"multi_account":    true,
		},
		Limitations: []string{
			"Requires consistent historical data",
			"May not detect subtle anomalies",
			"Limited to numeric metrics",
		},
	}
}

// GetEnhancedToolInfo returns enhanced tool information including metadata
func GetEnhancedToolInfo(tool Tool) map[string]interface{} {
	// Ensure metadata is populated
	if tool.Metadata == nil {
		EnhanceToolMetadata(&tool)
	}
	
	info := map[string]interface{}{
		"name":        tool.Name,
		"description": tool.Description,
		"parameters":  tool.Parameters,
	}
	
	if tool.Metadata != nil {
		info["metadata"] = tool.Metadata
	}
	
	return info
}

// GenerateAIGuidance generates contextual guidance for a tool
func GenerateAIGuidance(toolName string, context map[string]interface{}) string {
	guidance := []string{}
	
	// Add general guidance
	guidance = append(guidance, fmt.Sprintf("Using tool: %s", toolName))
	
	// Add context-specific guidance
	if errorCount, ok := context["recent_errors"].(int); ok && errorCount > 0 {
		guidance = append(guidance, fmt.Sprintf("Note: This tool had %d recent errors. Check parameters carefully.", errorCount))
	}
	
	if isFirstUse, ok := context["first_use"].(bool); ok && isFirstUse {
		guidance = append(guidance, "First time using this tool. Consider starting with a simple example.")
	}
	
	if relatedTool, ok := context["previous_tool"].(string); ok {
		switch relatedTool {
		case "discovery.explore_event_types":
			guidance = append(guidance, "Good! You've discovered event types. Now explore their attributes or query the data.")
		case "query_nrdb":
			guidance = append(guidance, "Based on your query results, you might want to create alerts or dashboards.")
		}
	}
	
	return strings.Join(guidance, " ")
}
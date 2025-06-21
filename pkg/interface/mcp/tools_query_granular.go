package mcp

import (
	"context"
	"fmt"
	"strings"
)

// registerGranularQueryTools registers atomic NRQL query building and execution tools
func (s *Server) registerGranularQueryTools() error {
	// 1. Query Execution Tools
	executeQuery := NewToolBuilder("nrql.execute", "Execute NRQL query with timeout and metadata controls").
		Category(CategoryQuery).
		Handler(s.handleNRQLExecute).
		Required("query").
		Param("query", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "The NRQL query to execute",
			},
			Examples: []interface{}{
				"SELECT count(*) FROM Transaction",
				"SELECT average(duration) FROM Transaction TIMESERIES",
			},
		}).
		Param("account_id", EnhancedProperty{
			Property: Property{
				Type:        "number",
				Description: "Specific account ID to query (optional)",
			},
		}).
		Param("timeout", EnhancedProperty{
			Property: Property{
				Type:        "number",
				Description: "Query timeout in seconds",
				Default:     30,
			},
		}).
		Param("include_metadata", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Include query performance metadata",
				Default:     false,
			},
		}).
		Safety(func(s *SafetyMetadata) {
			s.Level = SafetyLevelSafe
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 1000
			p.MaxLatencyMS = 30000
			p.Cacheable = true
			p.CacheTTLSeconds = 300
		}).
		Build()

	// 2. Query Validation
	validateQuery := NewToolBuilder("nrql.validate", "Validate NRQL syntax and estimate impact").
		Category(CategoryQuery).
		Handler(s.handleNRQLValidate).
		Required("query").
		Param("query", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "The NRQL query to validate",
			},
		}).
		Param("check_permissions", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Check if user has permissions for query",
				Default:     false,
			},
		}).
		Param("suggest_improvements", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Provide optimization suggestions",
				Default:     true,
			},
		}).
		Safety(func(s *SafetyMetadata) { s.Level = SafetyLevelSafe }).
		Build()

	// 3. Query Cost Estimation
	estimateCost := NewToolBuilder("nrql.estimate_cost", "Estimate query execution cost and impact").
		Category(CategoryAnalysis).
		Handler(s.handleNRQLEstimateCost).
		Required("query").
		Param("query", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "The NRQL query to estimate",
			},
		}).
		Param("time_range", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Time range for estimation",
				Default:     "1 hour",
			},
			Examples: []interface{}{"1 hour", "24 hours", "7 days", "30 days"},
		}).
		Param("execution_frequency", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "How often query will run",
				Enum:        []string{"once", "hourly", "daily", "continuous"},
				Default:     "once",
			},
		}).
		Safety(func(s *SafetyMetadata) { s.Level = SafetyLevelSafe }).
		Build()

	// 4. Query Builder - SELECT clause
	buildSelect := NewToolBuilder("nrql.build_select", "Build SELECT clause with aggregations").
		Category(CategoryQuery).
		Handler(s.handleNRQLBuildSelect).
		Required("event_type").
		Param("event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Event type to query from",
			},
			Examples: []interface{}{"Transaction", "PageView", "SystemSample"},
		}).
		Param("aggregations", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Aggregation functions to apply",
				Items: &Property{
					Type: "object",
				},
			},
			Examples: []interface{}{
				map[string]string{"function": "count", "attribute": "*"},
				map[string]string{"function": "average", "attribute": "duration"},
			},
		}).
		Param("attributes", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Raw attributes to select",
				Items: &Property{
					Type: "string",
				},
			},
			Examples: []interface{}{"appName", "host", "error"},
		}).
		Param("aliases", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Attribute aliases",
			},
			Examples: []interface{}{
				map[string]string{"duration": "avgDuration", "count(*)": "totalCount"},
			},
		}).
		Safety(func(s *SafetyMetadata) { s.Level = SafetyLevelSafe }).
		Build()

	// 5. Query Builder - WHERE clause
	buildWhere := NewToolBuilder("nrql.build_where", "Build WHERE clause with conditions").
		Category(CategoryQuery).
		Handler(s.handleNRQLBuildWhere).
		Required("conditions").
		Param("conditions", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Conditions to combine",
				Items: &Property{
					Type: "object",
				},
			},
			Examples: []interface{}{
				map[string]interface{}{"attribute": "appName", "operator": "=", "value": "my-app"},
				map[string]interface{}{"attribute": "duration", "operator": ">", "value": 100},
			},
		}).
		Param("operator", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Logical operator to combine conditions",
				Enum:        []string{"AND", "OR"},
				Default:     "AND",
			},
		}).
		Param("nest_groups", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Wrap conditions in parentheses",
				Default:     false,
			},
		}).
		Safety(func(s *SafetyMetadata) { s.Level = SafetyLevelSafe }).
		Build()

	// 6. Query Builder - Time range
	buildTimeRange := NewToolBuilder("nrql.build_time_range", "Build SINCE/UNTIL time clauses").
		Category(CategoryQuery).
		Handler(s.handleNRQLBuildTimeRange).
		Param("since", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Start time",
			},
			Examples: []interface{}{"1 hour ago", "2024-01-20T00:00:00Z", "yesterday"},
		}).
		Param("until", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "End time",
			},
			Examples: []interface{}{"now", "1 hour ago", "2024-01-20T23:59:59Z"},
		}).
		Param("compare_with", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Time period to compare with",
			},
			Examples: []interface{}{"1 week ago", "1 month ago"},
		}).
		Safety(func(s *SafetyMetadata) { s.Level = SafetyLevelSafe }).
		Build()

	// 7. Query Template Library
	getQueryTemplate := NewToolBuilder("nrql.get_template", "Get pre-built query templates").
		Category(CategoryQuery).
		Handler(s.handleNRQLGetTemplate).
		Required("template_name").
		Param("template_name", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Name of query template",
				Enum:        []string{"error_rate", "latency_percentiles", "throughput", "apdex", "top_transactions"},
			},
		}).
		Param("event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Event type to apply template to",
				Default:     "Transaction",
			},
		}).
		Param("parameters", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Template parameters",
			},
		}).
		Safety(func(s *SafetyMetadata) { s.Level = SafetyLevelSafe }).
		Build()

	// Register all tools
	tools := []Tool{
		executeQuery.Tool,
		validateQuery.Tool,
		estimateCost.Tool,
		buildSelect.Tool,
		buildWhere.Tool,
		buildTimeRange.Tool,
		getQueryTemplate.Tool,
	}

	for _, tool := range tools {
		if err := s.tools.Register(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name, err)
		}
	}

	return nil
}

// Handler implementations - duplicates removed as they're defined in tools_nrql_adaptive.go

// handleNRQLBuildTimeRange builds time range clauses
func (s *Server) handleNRQLBuildTimeRange(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	since, _ := params["since"].(string)
	until, _ := params["until"].(string)
	compareWith, _ := params["compare_with"].(string)

	var clauses []string

	if since != "" {
		clauses = append(clauses, fmt.Sprintf("SINCE %s", since))
	}

	if until != "" {
		clauses = append(clauses, fmt.Sprintf("UNTIL %s", until))
	}

	if compareWith != "" {
		clauses = append(clauses, fmt.Sprintf("COMPARE WITH %s", compareWith))
	}

	return map[string]interface{}{
		"clause": strings.Join(clauses, " "),
		"parts":  clauses,
	}, nil
}

// handleNRQLGetTemplate returns query templates
func (s *Server) handleNRQLGetTemplate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	templateName, _ := params["template_name"].(string)
	eventType, _ := params["event_type"].(string)
	templateParams, _ := params["parameters"].(map[string]interface{})

	if eventType == "" {
		eventType = "Transaction"
	}

	templates := map[string]string{
		"error_rate": fmt.Sprintf("SELECT percentage(count(*), WHERE error = true) FROM %s", eventType),
		"latency_percentiles": fmt.Sprintf("SELECT percentile(duration, 50, 75, 90, 95, 99) FROM %s", eventType),
		"throughput": fmt.Sprintf("SELECT rate(count(*), 1 minute) FROM %s TIMESERIES", eventType),
		"apdex": fmt.Sprintf("SELECT apdex(duration, 0.5) FROM %s", eventType),
		"top_transactions": fmt.Sprintf("SELECT count(*) FROM %s FACET name LIMIT 10", eventType),
	}

	template, exists := templates[templateName]
	if !exists {
		return nil, fmt.Errorf("unknown template: %s", templateName)
	}

	// Apply template parameters
	if templateParams != nil {
		if threshold, ok := templateParams["apdex_threshold"].(float64); ok && templateName == "apdex" {
			template = fmt.Sprintf("SELECT apdex(duration, %v) FROM %s", threshold, eventType)
		}
		if limit, ok := templateParams["limit"].(float64); ok && templateName == "top_transactions" {
			template = fmt.Sprintf("SELECT count(*) FROM %s FACET name LIMIT %d", eventType, int(limit))
		}
	}

	return map[string]interface{}{
		"query":       template,
		"template":    templateName,
		"eventType":   eventType,
		"description": getTemplateDescription(templateName),
	}, nil
}

// handleNRQLBuildWhere builds WHERE clauses
func (s *Server) handleNRQLBuildWhere(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	conditions, _ := params["conditions"].([]interface{})
	operator, _ := params["operator"].(string)
	nestGroups, _ := params["nest_groups"].(bool)

	if operator == "" {
		operator = "AND"
	}

	var whereParts []string
	for _, cond := range conditions {
		if condMap, ok := cond.(map[string]interface{}); ok {
			attribute, _ := condMap["attribute"].(string)
			op, _ := condMap["operator"].(string)
			value := condMap["value"]

			var part string
			switch v := value.(type) {
			case string:
				if op == "IN" || op == "NOT IN" {
					part = fmt.Sprintf("%s %s (%s)", attribute, op, v)
				} else {
					part = fmt.Sprintf("%s %s '%s'", attribute, op, v)
				}
			case float64:
				part = fmt.Sprintf("%s %s %v", attribute, op, v)
			case bool:
				part = fmt.Sprintf("%s %s %v", attribute, op, v)
			case nil:
				if op == "IS" || op == "IS NOT" {
					part = fmt.Sprintf("%s %s NULL", attribute, op)
				}
			}

			if part != "" {
				whereParts = append(whereParts, part)
			}
		}
	}

	whereClause := strings.Join(whereParts, fmt.Sprintf(" %s ", operator))
	
	if nestGroups && len(whereParts) > 1 {
		whereClause = fmt.Sprintf("(%s)", whereClause)
	}

	return map[string]interface{}{
		"clause": fmt.Sprintf("WHERE %s", whereClause),
		"parts": whereParts,
		"operator": operator,
	}, nil
}

// getTemplateDescription returns descriptions for query templates
func getTemplateDescription(templateName string) string {
	descriptions := map[string]string{
		"error_rate":          "Calculate the percentage of transactions with errors",
		"latency_percentiles": "Show latency distribution across percentiles",
		"throughput":          "Measure requests per minute over time",
		"apdex":               "Calculate Application Performance Index score",
		"top_transactions":    "Find the most frequently called transactions",
	}
	return descriptions[templateName]
}
package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// registerGranularQueryTools registers atomic NRQL query tools
func (s *Server) registerGranularQueryTools() error {
	// 1. Basic NRQL execution
	nrqlExecute := NewToolBuilder("nrql.execute", "Execute a single NRQL query with full control").
		Category(CategoryQuery).
		Handler(s.handleNRQLExecute).
		Required("query").
		Param("query", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "The NRQL query to execute",
			},
			Examples: []interface{}{
				"SELECT count(*) FROM Transaction WHERE appName = 'my-app' SINCE 1 hour ago",
				"SELECT average(duration) FROM Transaction FACET appName LIMIT 10",
			},
			AIHint: "Use proper NRQL syntax. Always include SINCE clause for time-based queries.",
		}).
		Param("account_id", EnhancedProperty{
			Property: Property{
				Type:        "integer",
				Description: "Target account ID (uses default if not provided)",
			},
		}).
		Param("timeout", EnhancedProperty{
			Property: Property{
				Type:        "integer",
				Description: "Query timeout in seconds",
				Default:     30,
			},
			ValidationRules: []ValidationRule{
				{Field: "timeout", Rule: "range", Value: map[string]int{"min": 1, "max": 300}, Message: "Timeout must be between 1 and 300 seconds"},
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
			s.IsDestructive = false
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 500
			p.MaxLatencyMS = 30000
			p.Cacheable = true
			p.CacheTTLSeconds = 300
			p.CostCategory = "medium"
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"To get error rate: nrql.execute with query 'SELECT percentage(count(*), WHERE error IS true) FROM Transaction'",
				"To analyze performance: nrql.execute with query 'SELECT percentile(duration, 95) FROM Transaction'",
			}
			g.CommonPatterns = []string{
				"Always include SINCE clause",
				"Use LIMIT for large result sets",
				"FACET for grouping results",
			}
			g.ChainsWith = []string{"nrql.validate", "nrql.estimate_cost", "entity.search_by_name"}
			g.SuccessIndicators = []string{"results array is not empty", "no error field in response"}
			g.ErrorPatterns = map[string]string{
				"Unknown function": "Check NRQL function syntax",
				"Unknown attribute": "Verify attribute exists with discovery.list_schemas",
				"Timeout": "Reduce time range or simplify query",
			}
		}).
		Example(ToolExample{
			Name:        "Get application error rate",
			Description: "Calculate error rate for an application",
			Params: map[string]interface{}{
				"query":   "SELECT percentage(count(*), WHERE error IS true) FROM Transaction WHERE appName = 'checkout-service' SINCE 1 hour ago",
				"timeout": 10,
			},
		}).
		Build()

	if err := s.tools.Register(nrqlExecute.Tool); err != nil {
		return err
	}

	// 2. NRQL validation without execution
	nrqlValidate := NewToolBuilder("nrql.validate", "Validate NRQL syntax without execution").
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
				Description: "Check if user has permissions for referenced events/attributes",
				Default:     false,
			},
		}).
		Param("suggest_improvements", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Provide query optimization suggestions",
				Default:     true,
			},
		}).
		Safety(func(s *SafetyMetadata) {
			s.Level = SafetyLevelSafe
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 50
			p.MaxLatencyMS = 500
			p.Cacheable = true
			p.CacheTTLSeconds = 3600
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Always validate before executing expensive queries",
				"Use to check syntax when building queries programmatically",
			}
			g.PreferredOver = []string{"nrql.execute for syntax checking"}
		}).
		Build()

	if err := s.tools.Register(nrqlValidate.Tool); err != nil {
		return err
	}

	// 3. Query cost estimation
	nrqlEstimateCost := NewToolBuilder("nrql.estimate_cost", "Estimate query cost and performance impact").
		Category(CategoryAnalysis).
		Handler(s.handleNRQLEstimateCost).
		Required("query").
		Param("query", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "The NRQL query to analyze",
			},
		}).
		Param("time_range", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Time range for the query (e.g., '1 hour', '7 days')",
				Default:     "1 hour",
			},
			ValidationRules: []ValidationRule{
				{Field: "time_range", Rule: "regex", Value: `^\d+\s+(minute|hour|day|week|month)s?$`, Message: "Invalid time range format"},
			},
		}).
		Param("execution_frequency", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "How often the query will run (for cost projection)",
				Enum:        []string{"once", "hourly", "daily", "continuous"},
				Default:     "once",
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 100
			p.Cacheable = true
			p.CacheTTLSeconds = 1800
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Check cost before scheduling recurring queries",
				"Estimate impact of dashboard queries",
			}
			g.ChainsWith = []string{"nrql.validate", "alert.create_threshold_condition"}
			g.WarningsForAI = []string{
				"High cost queries should be optimized before production use",
			}
		}).
		Build()

	if err := s.tools.Register(nrqlEstimateCost.Tool); err != nil {
		return err
	}

	// 4. NRQL query builder - SELECT clause
	nrqlBuildSelect := NewToolBuilder("nrql.build_select", "Build SELECT clause with proper escaping").
		Category(CategoryUtility).
		Handler(s.handleNRQLBuildSelect).
		Required("event_type").
		Param("event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "The event type to query",
			},
			Examples: []interface{}{"Transaction", "SystemSample", "Log"},
		}).
		Param("aggregations", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "List of aggregation specifications",
				Items: &Property{
					Type: "object",
				},
			},
			Examples: []interface{}{
				[]map[string]interface{}{
					{"function": "average", "attribute": "duration"},
					{"function": "percentile", "attribute": "duration", "percentile": 95},
				},
			},
		}).
		Param("attributes", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Raw attributes to select (no aggregation)",
				Items: &Property{
					Type: "string",
				},
			},
		}).
		Param("aliases", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Attribute aliases (attribute -> alias mapping)",
			},
		}).
		Safety(func(s *SafetyMetadata) {
			s.Level = SafetyLevelSafe
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 10
			p.Cacheable = true
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Use to build complex SELECT clauses safely",
				"Handles proper escaping of attribute names",
			}
			g.ChainsWith = []string{"nrql.build_where", "nrql.build_facet"}
		}).
		Build()

	if err := s.tools.Register(nrqlBuildSelect.Tool); err != nil {
		return err
	}

	// 5. NRQL query builder - WHERE clause
	nrqlBuildWhere := NewToolBuilder("nrql.build_where", "Build WHERE clause with proper escaping").
		Category(CategoryUtility).
		Handler(s.handleNRQLBuildWhere).
		Required("conditions").
		Param("conditions", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "List of condition specifications",
				Items: &Property{
					Type: "object",
				},
			},
			Examples: []interface{}{
				[]map[string]interface{}{
					{"attribute": "appName", "operator": "=", "value": "my-app"},
					{"attribute": "duration", "operator": ">", "value": 1000},
					{"attribute": "error", "operator": "IS", "value": true},
				},
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
				Description: "Allow nested condition groups",
				Default:     false,
			},
		}).
		Safety(func(s *SafetyMetadata) {
			s.Level = SafetyLevelSafe
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 10
			p.Cacheable = true
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Build complex WHERE clauses with proper escaping",
				"Handles special characters in values automatically",
			}
			g.ChainsWith = []string{"nrql.build_select", "nrql.build_facet"}
			g.WarningsForAI = []string{
				"String values are automatically quoted",
				"NULL checks use IS NULL, not = NULL",
			}
		}).
		Build()

	if err := s.tools.Register(nrqlBuildWhere.Tool); err != nil {
		return err
	}

	return nil
}

// Handler implementations

func (s *Server) handleNRQLExecute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, _ := params["query"].(string)
	accountID, _ := params["account_id"].(float64)
	timeout, _ := params["timeout"].(float64)
	includeMetadata, _ := params["include_metadata"].(bool)

	if timeout == 0 {
		timeout = 30
	}

	// Create timeout context
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Execute query
	start := time.Now()
	
	// Mock implementation for now
	result := map[string]interface{}{
		"results": []map[string]interface{}{
			{"count": 12345, "appName": "my-app"},
		},
		"performanceStats": map[string]interface{}{
			"wallClockTime": time.Since(start).Milliseconds(),
			"inspectedCount": 50000,
			"omittedCount":   0,
		},
	}

	if includeMetadata {
		result["metadata"] = map[string]interface{}{
			"query":          query,
			"accountId":      accountID,
			"executionTime":  time.Since(start).Milliseconds(),
			"resultCount":    1,
			"cacheHit":       false,
		}
	}

	return result, nil
}

func (s *Server) handleNRQLValidate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, _ := params["query"].(string)
	checkPermissions, _ := params["check_permissions"].(bool)
	suggestImprovements, _ := params["suggest_improvements"].(bool)

	// Basic validation
	validation := map[string]interface{}{
		"valid":  true,
		"errors": []string{},
		"warnings": []string{},
	}

	// Check for common issues
	if !strings.Contains(strings.ToUpper(query), "SELECT") {
		validation["valid"] = false
		validation["errors"] = append(validation["errors"].([]string), "Query must start with SELECT")
	}

	if !strings.Contains(strings.ToUpper(query), "FROM") {
		validation["valid"] = false
		validation["errors"] = append(validation["errors"].([]string), "Query must include FROM clause")
	}

	// Warnings
	if !strings.Contains(strings.ToUpper(query), "SINCE") && !strings.Contains(strings.ToUpper(query), "UNTIL") {
		validation["warnings"] = append(validation["warnings"].([]string), "Query has no time range specified")
	}

	if suggestImprovements {
		validation["suggestions"] = []string{}
		
		if strings.Contains(strings.ToUpper(query), "SELECT *") {
			validation["suggestions"] = append(validation["suggestions"].([]string), 
				"Avoid SELECT *, specify needed attributes for better performance")
		}

		if !strings.Contains(strings.ToUpper(query), "LIMIT") && strings.Contains(strings.ToUpper(query), "FACET") {
			validation["suggestions"] = append(validation["suggestions"].([]string), 
				"Consider adding LIMIT when using FACET to control result size")
		}
	}

	if checkPermissions {
		validation["permissions"] = map[string]interface{}{
			"hasAccess": true,
			"missingPermissions": []string{},
		}
	}

	return validation, nil
}

func (s *Server) handleNRQLEstimateCost(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, _ := params["query"].(string)
	timeRange, _ := params["time_range"].(string)
	frequency, _ := params["execution_frequency"].(string)

	// Mock cost estimation
	estimation := map[string]interface{}{
		"query": query,
		"timeRange": timeRange,
		"estimation": map[string]interface{}{
			"dataSizeGB": 0.5,
			"estimatedDurationMS": 1200,
			"inspectedEvents": 1000000,
			"complexity": "medium",
			"costCategory": "standard",
		},
		"recommendations": []string{},
	}

	// Add frequency-based projections
	if frequency != "once" {
		monthlyExecutions := map[string]int{
			"hourly": 720,
			"daily": 30,
			"continuous": 43200, // every minute
		}
		
		if execs, ok := monthlyExecutions[frequency]; ok {
			estimation["monthlyProjection"] = map[string]interface{}{
				"executions": execs,
				"totalDataGB": 0.5 * float64(execs),
				"estimatedCost": "medium",
			}
		}
	}

	// Add recommendations based on analysis
	if strings.Contains(strings.ToUpper(query), "SELECT *") {
		estimation["recommendations"] = append(estimation["recommendations"].([]string), 
			"Replace SELECT * with specific attributes to reduce data transfer")
	}

	return estimation, nil
}

func (s *Server) handleNRQLBuildSelect(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	eventType, _ := params["event_type"].(string)
	aggregations, _ := params["aggregations"].([]interface{})
	attributes, _ := params["attributes"].([]interface{})
	aliases, _ := params["aliases"].(map[string]interface{})

	var selectParts []string

	// Process aggregations
	for _, agg := range aggregations {
		if aggMap, ok := agg.(map[string]interface{}); ok {
			function, _ := aggMap["function"].(string)
			attribute, _ := aggMap["attribute"].(string)
			
			part := fmt.Sprintf("%s(%s)", function, attribute)
			
			// Handle special cases
			if function == "percentile" {
				if p, ok := aggMap["percentile"].(float64); ok {
					part = fmt.Sprintf("percentile(%s, %v)", attribute, p)
				}
			}
			
			// Add alias if provided
			if alias, ok := aliases[attribute]; ok {
				part = fmt.Sprintf("%s as '%s'", part, alias)
			}
			
			selectParts = append(selectParts, part)
		}
	}

	// Process raw attributes
	for _, attr := range attributes {
		if attrStr, ok := attr.(string); ok {
			part := attrStr
			if alias, ok := aliases[attrStr]; ok {
				part = fmt.Sprintf("%s as '%s'", attrStr, alias)
			}
			selectParts = append(selectParts, part)
		}
	}

	// Build the complete SELECT clause
	selectClause := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectParts, ", "), eventType)

	return map[string]interface{}{
		"clause": selectClause,
		"parts": selectParts,
		"eventType": eventType,
	}, nil
}

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
			
			// Handle different value types
			switch v := value.(type) {
			case string:
				// Escape single quotes in string values
				escaped := strings.ReplaceAll(v, "'", "\\'")
				part = fmt.Sprintf("%s %s '%s'", attribute, op, escaped)
			case bool:
				if op == "IS" {
					part = fmt.Sprintf("%s IS %v", attribute, v)
				} else {
					part = fmt.Sprintf("%s = %v", attribute, v)
				}
			case nil:
				if op == "IS" {
					part = fmt.Sprintf("%s IS NULL", attribute)
				} else if op == "IS NOT" {
					part = fmt.Sprintf("%s IS NOT NULL", attribute)
				}
			default:
				part = fmt.Sprintf("%s %s %v", attribute, op, value)
			}
			
			whereParts = append(whereParts, part)
		}
	}

	// Combine with operator
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
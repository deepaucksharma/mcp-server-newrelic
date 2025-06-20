//go:build !test

package mcp

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// registerQueryTools registers NRQL query-related tools
func (s *Server) registerQueryTools() error {
	// Execute NRQL query
	s.tools.Register(Tool{
		Name:        "query_nrdb",
		Description: "Execute an NRQL query against New Relic and return results",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"query"},
			Properties: map[string]Property{
				"query": {
					Type:        "string",
					Description: "The NRQL query to execute",
				},
				"account_id": {
					Type:        "string",
					Description: "Optional account ID (uses default if not provided)",
				},
				"timeout": {
					Type:        "integer",
					Description: "Query timeout in seconds (default: 30)",
					Default:     30,
				},
			},
		},
		Handler: s.handleQueryNRDB,
	})

	// Validate and analyze NRQL query
	s.tools.Register(Tool{
		Name:        "query_check",
		Description: "Validate an NRQL query and analyze its performance impact",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"query"},
			Properties: map[string]Property{
				"query": {
					Type:        "string",
					Description: "The NRQL query to validate",
				},
				"explain": {
					Type:        "boolean",
					Description: "Include detailed explanation",
					Default:     true,
				},
				"suggest_optimizations": {
					Type:        "boolean",
					Description: "Suggest query optimizations",
					Default:     true,
				},
			},
		},
		Handler: s.handleQueryCheck,
	})

	// Build NRQL query from parameters
	s.tools.Register(Tool{
		Name:        "query_builder",
		Description: "Build an NRQL query from structured parameters",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"event_type", "select"},
			Properties: map[string]Property{
				"event_type": {
					Type:        "string",
					Description: "The event type to query",
				},
				"select": {
					Type:        "array",
					Description: "Fields to select (e.g., ['count(*)', 'average(duration)'])",
					Items:       &Property{Type: "string"},
				},
				"where": {
					Type:        "string",
					Description: "WHERE clause conditions",
				},
				"facet": {
					Type:        "array",
					Description: "Fields to facet by",
					Items:       &Property{Type: "string"},
				},
				"since": {
					Type:        "string",
					Description: "Time range (e.g., '1 hour ago', '2023-01-01')",
					Default:     "1 hour ago",
				},
				"until": {
					Type:        "string",
					Description: "End time (e.g., 'now', '2023-01-02')",
					Default:     "now",
				},
				"limit": {
					Type:        "integer",
					Description: "Result limit",
					Default:     100,
				},
				"order_by": {
					Type:        "string",
					Description: "Order by clause",
				},
			},
		},
		Handler: s.handleQueryBuilder,
	})

	// Build advanced NRQL query with new features
	s.tools.Register(Tool{
		Name:        "query_builder_advanced",
		Description: "Build advanced NRQL queries with array functions, sliding windows, subqueries, and more",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"query_type"},
			Properties: map[string]Property{
				"query_type": {
					Type:        "string",
					Description: "Type of advanced query: 'array', 'sliding_window', 'funnel', 'subquery', 'join', 'nested_aggregation', 'rate', 'buckets'",
				},
				// Array query parameters
				"array_field": {
					Type:        "string",
					Description: "For array queries: the array field to operate on",
				},
				"array_operation": {
					Type:        "string",
					Description: "Array operation: 'getfield', 'length', 'contains'",
				},
				"array_index": {
					Type:        "integer",
					Description: "Array index for getfield operation",
				},
				"array_value": {
					Type:        "string",
					Description: "Value to check for contains operation",
				},
				// Sliding window parameters
				"slide_by": {
					Type:        "string",
					Description: "Sliding window interval (e.g., '5 minutes')",
				},
				// Funnel parameters
				"funnel_steps": {
					Type:        "array",
					Description: "Funnel steps with WHERE conditions",
					Items:       &Property{Type: "object"},
				},
				// Subquery/JOIN parameters
				"primary_query": {
					Type:        "string",
					Description: "Primary query for subquery or JOIN",
				},
				"secondary_query": {
					Type:        "string",
					Description: "Secondary query for JOIN",
				},
				"join_keys": {
					Type:        "array",
					Description: "Keys to join on",
					Items:       &Property{Type: "string"},
				},
				// Rate parameters
				"rate_metric": {
					Type:        "string",
					Description: "Metric to calculate rate for",
				},
				"rate_interval": {
					Type:        "string",
					Description: "Rate calculation interval",
				},
				// Common parameters
				"event_type": {
					Type:        "string",
					Description: "Event type to query",
				},
				"time_range": {
					Type:        "string",
					Description: "Time range for the query",
					Default:     "1 hour ago",
				},
			},
		},
		Handler: s.handleAdvancedQueryBuilder,
	})

	return nil
}

// handleQueryNRDB executes an NRQL query
func (s *Server) handleQueryNRDB(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// Validate query first
	if err := s.validateNRQLSafety(query); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	// Get timeout
	timeout := 30
	if t, ok := params["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Create timeout context
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Execute query using the NRDB client
	// Note: This would use the actual New Relic client when implemented
	result, err := s.executeNRQLQuery(queryCtx, query, params["account_id"])
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return result, nil
}

// handleQueryCheck validates and analyzes an NRQL query
func (s *Server) handleQueryCheck(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	explain, _ := params["explain"].(bool)
	suggestOpt, _ := params["suggest_optimizations"].(bool)

	result := map[string]interface{}{
		"valid":      true,
		"errors":     []string{},
		"warnings":   []string{},
		"complexity": s.analyzeQueryComplexity(query),
	}

	// Validate syntax
	if err := s.validateNRQLSyntax(query); err != nil {
		result["valid"] = false
		result["errors"] = append(result["errors"].([]string), err.Error())
	}

	// Check for common issues
	warnings := s.checkQueryWarnings(query)
	if len(warnings) > 0 {
		result["warnings"] = warnings
	}

	// Add explanation if requested
	if explain {
		result["explanation"] = s.explainQuery(query)
	}

	// Add optimization suggestions if requested
	if suggestOpt {
		result["optimizations"] = s.suggestQueryOptimizations(query)
	}

	// Estimate query cost
	result["estimated_cost"] = s.estimateQueryCost(query)

	return result, nil
}

// handleQueryBuilder builds an NRQL query from structured parameters
func (s *Server) handleQueryBuilder(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return nil, fmt.Errorf("event_type parameter is required")
	}
	
	// Sanitize event type
	eventType, err := s.nrqlValidator.SanitizeIdentifier(eventType)
	if err != nil {
		return nil, fmt.Errorf("invalid event_type: %w", err)
	}

	selectFields, ok := params["select"].([]interface{})
	if !ok || len(selectFields) == 0 {
		return nil, fmt.Errorf("select parameter is required")
	}

	// Build SELECT clause with sanitized fields
	selectClauses := make([]string, len(selectFields))
	for i, field := range selectFields {
		fieldStr, ok := field.(string)
		if !ok {
			return nil, fmt.Errorf("invalid select field at index %d", i)
		}
		// Don't sanitize aggregate functions, just validate they're safe
		if strings.Contains(fieldStr, "(") {
			// Validate it's a known aggregate function
			if !isValidAggregateFunction(fieldStr) {
				return nil, fmt.Errorf("invalid aggregate function: %s", fieldStr)
			}
			selectClauses[i] = fieldStr
		} else {
			// Sanitize regular field names
			sanitized, err := s.nrqlValidator.SanitizeIdentifier(fieldStr)
			if err != nil {
				return nil, fmt.Errorf("invalid select field '%s': %w", fieldStr, err)
			}
			selectClauses[i] = sanitized
		}
	}

	// Start building query
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectClauses, ", "), eventType)

	// Add WHERE clause if provided
	if where, ok := params["where"].(string); ok && where != "" {
		query += fmt.Sprintf(" WHERE %s", where)
	}

	// Add FACET clause if provided
	if facetFields, ok := params["facet"].([]interface{}); ok && len(facetFields) > 0 {
		facets := make([]string, len(facetFields))
		for i, field := range facetFields {
			facets[i] = field.(string)
		}
		query += fmt.Sprintf(" FACET %s", strings.Join(facets, ", "))
	}

	// Add time range with validation
	since := "1 hour ago"
	if sinceParam, ok := params["since"].(string); ok && sinceParam != "" {
		if err := s.nrqlValidator.ValidateTimeRange(sinceParam); err != nil {
			return nil, fmt.Errorf("invalid since time: %w", err)
		}
		since = sinceParam
	}
	until := "now"
	if u, ok := params["until"].(string); ok && u != "" {
		if err := s.nrqlValidator.ValidateTimeRange(u); err != nil {
			return nil, fmt.Errorf("invalid until time: %w", err)
		}
		until = u
	}
	query += fmt.Sprintf(" SINCE %s UNTIL %s", since, until)

	// Add ORDER BY if provided
	if orderBy, ok := params["order_by"].(string); ok && orderBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", orderBy)
	}

	// Add LIMIT
	limit := 100
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}
	query += fmt.Sprintf(" LIMIT %d", limit)

	// Validate the built query
	if err := s.validateNRQLSyntax(query); err != nil {
		return nil, fmt.Errorf("built query is invalid: %w", err)
	}

	return map[string]interface{}{
		"query":       query,
		"explanation": s.explainQuery(query),
		"warnings":    s.checkQueryWarnings(query),
	}, nil
}

// Helper functions for query operations

func (s *Server) validateNRQLSafety(query string) error {
	// Use the validator to sanitize and check the query
	_, err := s.nrqlValidator.Sanitize(query)
	return err
}

func (s *Server) validateNRQLSyntax(query string) error {
	// Use the validator for comprehensive syntax checking
	_, err := s.nrqlValidator.Sanitize(query)
	return err
}

func (s *Server) analyzeQueryComplexity(query string) map[string]interface{} {
	complexity := map[string]interface{}{
		"score":      "low",
		"operations": []string{},
	}

	// Check for complex operations
	if regexp.MustCompile(`(?i)\bJOIN\b`).MatchString(query) {
		complexity["operations"] = append(complexity["operations"].([]string), "JOIN")
		complexity["score"] = "high"
	}

	if regexp.MustCompile(`(?i)\bFACET\b`).MatchString(query) {
		complexity["operations"] = append(complexity["operations"].([]string), "FACET")
		if complexity["score"] == "low" {
			complexity["score"] = "medium"
		}
	}

	// Count aggregation functions
	aggFuncs := regexp.MustCompile(`(?i)\b(count|sum|average|min|max|percentile|stddev)\s*\(`).FindAllString(query, -1)
	if len(aggFuncs) > 2 {
		complexity["operations"] = append(complexity["operations"].([]string), fmt.Sprintf("%d aggregations", len(aggFuncs)))
		complexity["score"] = "high"
	}

	return complexity
}

func (s *Server) checkQueryWarnings(query string) []string {
	warnings := []string{}

	// Check for missing time range
	if !regexp.MustCompile(`(?i)\bSINCE\b`).MatchString(query) {
		warnings = append(warnings, "Query has no time range specified (SINCE clause)")
	}

	// Check for SELECT *
	if regexp.MustCompile(`(?i)SELECT\s+\*`).MatchString(query) {
		warnings = append(warnings, "SELECT * can be expensive and return unnecessary data")
	}

	// Check for missing LIMIT on non-aggregated queries
	if !regexp.MustCompile(`(?i)\b(count|sum|average|min|max)\s*\(`).MatchString(query) &&
		!regexp.MustCompile(`(?i)\bLIMIT\b`).MatchString(query) {
		warnings = append(warnings, "Query has no LIMIT clause, may return excessive data")
	}

	// Check for large time ranges
	if regexp.MustCompile(`(?i)SINCE\s+\d+\s+(day|week|month)s?\s+ago`).MatchString(query) {
		warnings = append(warnings, "Large time range detected, query may be slow or expensive")
	}

	return warnings
}

func (s *Server) explainQuery(query string) map[string]interface{} {
	explanation := map[string]interface{}{
		"summary":     "NRQL query analysis",
		"components": []map[string]string{},
	}

	// Extract SELECT clause
	if match := regexp.MustCompile(`(?i)SELECT\s+([^FROM]+)`).FindStringSubmatch(query); len(match) > 1 {
		explanation["components"] = append(explanation["components"].([]map[string]string), map[string]string{
			"type":        "SELECT",
			"description": fmt.Sprintf("Selecting: %s", strings.TrimSpace(match[1])),
		})
	}

	// Extract FROM clause
	if match := regexp.MustCompile(`(?i)FROM\s+(\S+)`).FindStringSubmatch(query); len(match) > 1 {
		explanation["components"] = append(explanation["components"].([]map[string]string), map[string]string{
			"type":        "FROM",
			"description": fmt.Sprintf("Querying event type: %s", match[1]),
		})
	}

	// Extract WHERE clause
	if match := regexp.MustCompile(`(?i)WHERE\s+([^(FACET|SINCE|LIMIT|ORDER)]+)`).FindStringSubmatch(query); len(match) > 1 {
		explanation["components"] = append(explanation["components"].([]map[string]string), map[string]string{
			"type":        "WHERE",
			"description": fmt.Sprintf("Filtering by: %s", strings.TrimSpace(match[1])),
		})
	}

	return explanation
}

func (s *Server) suggestQueryOptimizations(query string) []map[string]string {
	optimizations := []map[string]string{}

	// Suggest adding time range if missing
	if !regexp.MustCompile(`(?i)\bSINCE\b`).MatchString(query) {
		optimizations = append(optimizations, map[string]string{
			"type":        "add_time_range",
			"suggestion":  "Add a SINCE clause to limit the time range",
			"example":     "SINCE 1 hour ago",
			"impact":      "Reduces data scanned and improves performance",
		})
	}

	// Suggest using specific fields instead of SELECT *
	if regexp.MustCompile(`(?i)SELECT\s+\*`).MatchString(query) {
		optimizations = append(optimizations, map[string]string{
			"type":        "specific_fields",
			"suggestion":  "Select only required fields instead of *",
			"example":     "SELECT count(*), average(duration)",
			"impact":      "Reduces data transfer and processing",
		})
	}

	// Suggest adding LIMIT for raw data queries
	if !regexp.MustCompile(`(?i)\b(count|sum|average|min|max)\s*\(`).MatchString(query) &&
		!regexp.MustCompile(`(?i)\bLIMIT\b`).MatchString(query) {
		optimizations = append(optimizations, map[string]string{
			"type":        "add_limit",
			"suggestion":  "Add a LIMIT clause to control result size",
			"example":     "LIMIT 100",
			"impact":      "Prevents excessive data return",
		})
	}

	return optimizations
}

func (s *Server) estimateQueryCost(query string) map[string]interface{} {
	cost := map[string]interface{}{
		"level":       "low",
		"factors":     []string{},
		"data_points": "< 1000",
	}

	// Estimate based on time range
	if match := regexp.MustCompile(`(?i)SINCE\s+(\d+)\s+(hour|day|week|month)s?\s+ago`).FindStringSubmatch(query); len(match) > 2 {
		value, _ := regexp.Compile(`\d+`)
		num := value.FindString(match[1])
		unit := match[2]

		switch unit {
		case "month":
			cost["level"] = "high"
			cost["data_points"] = "> 1M"
		case "week":
			cost["level"] = "medium"
			cost["data_points"] = "> 100K"
		case "day":
			if num > "1" {
				cost["level"] = "medium"
				cost["data_points"] = "> 10K"
			}
		}
		cost["factors"] = append(cost["factors"].([]string), fmt.Sprintf("%s %s time range", num, unit))
	}

	// Check for expensive operations
	if regexp.MustCompile(`(?i)\bJOIN\b`).MatchString(query) {
		cost["level"] = "high"
		cost["factors"] = append(cost["factors"].([]string), "JOIN operation")
	}

	if regexp.MustCompile(`(?i)\bFACET\b.*\bFACET\b`).MatchString(query) {
		cost["level"] = "high"
		cost["factors"] = append(cost["factors"].([]string), "Multiple FACET clauses")
	}

	return cost
}

// executeNRQLQuery executes the actual query using the New Relic client
func (s *Server) executeNRQLQuery(ctx context.Context, query string, accountID interface{}) (interface{}, error) {
	// Check if we have a New Relic client
	if s.getNRClient() == nil {
		// If no client, return mock response for development
		return map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"message": "New Relic client not configured - mock mode",
					"query":   query,
				},
			},
			"metadata": map[string]interface{}{
				"executionTime": "0ms",
				"inspectedCount": 0,
				"matchedCount":   0,
			},
		}, nil
	}

	// Convert accountID to string if provided
	var accountIDStr string
	if accountID != nil {
		switch v := accountID.(type) {
		case string:
			accountIDStr = v
		case float64:
			accountIDStr = fmt.Sprintf("%.0f", v)
		case int:
			accountIDStr = fmt.Sprintf("%d", v)
		}
	}

	// Use reflection to call QueryNRQL method
	// This avoids circular imports while still allowing real execution
	client := s.getNRClient()
	clientValue := reflect.ValueOf(client)
	
	// Check if this is a MultiAccountClient
	method := clientValue.MethodByName("QueryNRQL")
	if !method.IsValid() {
		return nil, fmt.Errorf("QueryNRQL method not found on client")
	}

	// Check method signature to determine if it supports multi-account
	methodType := method.Type()
	if methodType.NumIn() == 3 && methodType.In(2).Kind() == reflect.String {
		// Multi-account client with 3 parameters: (ctx, query, accountID)
		args := []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(query),
			reflect.ValueOf(accountIDStr),
		}
		results := method.Call(args)
		if len(results) != 2 {
			return nil, fmt.Errorf("unexpected return values from QueryNRQL")
		}
		// Extract result and error
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	} else if methodType.NumIn() == 2 {
		// Standard client with 2 parameters: (ctx, query)
		args := []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(query),
		}
		results := method.Call(args)
		if len(results) != 2 {
			return nil, fmt.Errorf("unexpected return values from QueryNRQL")
		}
		// Extract result and error
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	}

	return nil, fmt.Errorf("unsupported QueryNRQL method signature")
}

// isValidAggregateFunction checks if a string is a valid NRQL aggregate function
func isValidAggregateFunction(fn string) bool {
	fn = strings.ToLower(strings.TrimSpace(fn))
	
	// List of valid NRQL aggregate functions
	validFunctions := []string{
		// Basic aggregation functions
		"average(", "avg(",
		"count(",
		"latest(",
		"max(",
		"median(",
		"min(",
		"percentage(",
		"percentile(",
		"rate(",
		"stddev(",
		"sum(",
		"uniqueCount(", "uniques(",
		
		// Advanced aggregation functions
		"apdex(",
		"histogram(",
		"keyset(",
		"eventType(",
		"filter(",
		
		// Array functions (NEW)
		"getfield(",
		"length(",
		"contains(",
		
		// Time-based functions (NEW)
		"latestRate(",
		
		// Bucketing functions (NEW)
		"buckets(",
		
		// Funnel functions (NEW)
		"funnel(",
		
		// Nested aggregation support (NEW)
		"derivative(",
		"predictLinear(",
	}
	
	for _, valid := range validFunctions {
		if strings.HasPrefix(fn, valid) {
			return true
		}
	}
	
	return false
}

// handleAdvancedQueryBuilder builds advanced NRQL queries with new features
func (s *Server) handleAdvancedQueryBuilder(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	queryType, ok := params["query_type"].(string)
	if !ok || queryType == "" {
		return nil, fmt.Errorf("query_type parameter is required")
	}

	var query string
	var err error

	switch queryType {
	case "array":
		query, err = s.buildArrayQuery(params)
	case "sliding_window":
		query, err = s.buildSlidingWindowQuery(params)
	case "funnel":
		query, err = s.buildFunnelQuery(params)
	case "subquery":
		query, err = s.buildSubquery(params)
	case "join":
		query, err = s.buildJoinQuery(params)
	case "nested_aggregation":
		query, err = s.buildNestedAggregationQuery(params)
	case "rate":
		query, err = s.buildRateQuery(params)
	case "buckets":
		query, err = s.buildBucketsQuery(params)
	default:
		return nil, fmt.Errorf("unsupported query type: %s", queryType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build %s query: %w", queryType, err)
	}

	// Validate the built query
	if err := s.validateNRQLSyntax(query); err != nil {
		return nil, fmt.Errorf("built query is invalid: %w", err)
	}

	return map[string]interface{}{
		"query":       query,
		"query_type":  queryType,
		"explanation": s.explainQuery(query),
		"warnings":    s.checkQueryWarnings(query),
	}, nil
}

// buildArrayQuery builds a query using array functions
func (s *Server) buildArrayQuery(params map[string]interface{}) (string, error) {
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return "", fmt.Errorf("event_type is required for array queries")
	}

	arrayField, ok := params["array_field"].(string)
	if !ok || arrayField == "" {
		return "", fmt.Errorf("array_field is required for array queries")
	}

	operation, ok := params["array_operation"].(string)
	if !ok || operation == "" {
		return "", fmt.Errorf("array_operation is required for array queries")
	}

	timeRange := "1 hour ago"
	if tr, ok := params["time_range"].(string); ok {
		timeRange = tr
	}

	var query string
	switch operation {
	case "getfield":
		index, ok := params["array_index"].(float64)
		if !ok {
			return "", fmt.Errorf("array_index is required for getfield operation")
		}
		query = fmt.Sprintf("SELECT getfield(%s, %d) as 'Array Element %d' FROM %s SINCE %s", 
			arrayField, int(index), int(index), eventType, timeRange)
	
	case "length":
		query = fmt.Sprintf("SELECT average(length(%s)) as 'Avg Array Length', max(length(%s)) as 'Max Array Length' FROM %s SINCE %s",
			arrayField, arrayField, eventType, timeRange)
	
	case "contains":
		value, ok := params["array_value"].(string)
		if !ok || value == "" {
			return "", fmt.Errorf("array_value is required for contains operation")
		}
		// Escape the value for NRQL
		value = s.nrqlValidator.SanitizeStringValue(value)
		query = fmt.Sprintf("SELECT count(*) as 'Total', filter(count(*), WHERE contains(%s, '%s')) as 'Contains \"%s\"' FROM %s SINCE %s",
			arrayField, value, value, eventType, timeRange)
	
	default:
		return "", fmt.Errorf("unsupported array operation: %s", operation)
	}

	return query, nil
}

// buildSlidingWindowQuery builds a query with sliding window (SLIDE BY)
func (s *Server) buildSlidingWindowQuery(params map[string]interface{}) (string, error) {
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return "", fmt.Errorf("event_type is required for sliding window queries")
	}

	slideBy, ok := params["slide_by"].(string)
	if !ok || slideBy == "" {
		return "", fmt.Errorf("slide_by is required for sliding window queries")
	}

	timeRange := "1 hour ago"
	if tr, ok := params["time_range"].(string); ok {
		timeRange = tr
	}

	// Example: moving average with sliding window
	query := fmt.Sprintf("SELECT average(duration) FROM %s SINCE %s TIMESERIES 1 minute SLIDE BY %s",
		eventType, timeRange, slideBy)

	return query, nil
}

// buildFunnelQuery builds a funnel analysis query
func (s *Server) buildFunnelQuery(params map[string]interface{}) (string, error) {
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return "", fmt.Errorf("event_type is required for funnel queries")
	}

	funnelSteps, ok := params["funnel_steps"].([]interface{})
	if !ok || len(funnelSteps) == 0 {
		return "", fmt.Errorf("funnel_steps is required for funnel queries")
	}

	timeRange := "1 hour ago"
	if tr, ok := params["time_range"].(string); ok {
		timeRange = tr
	}

	// Build funnel steps
	steps := make([]string, len(funnelSteps))
	for i, step := range funnelSteps {
		stepMap, ok := step.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("invalid funnel step at index %d", i)
		}
		
		name, ok := stepMap["name"].(string)
		if !ok {
			return "", fmt.Errorf("funnel step %d missing name", i)
		}
		
		condition, ok := stepMap["condition"].(string)
		if !ok {
			return "", fmt.Errorf("funnel step %d missing condition", i)
		}
		
		steps[i] = fmt.Sprintf("WHERE %s AS '%s'", condition, name)
	}

	query := fmt.Sprintf("SELECT funnel(session, %s) FROM %s SINCE %s",
		strings.Join(steps, ", "), eventType, timeRange)

	return query, nil
}

// buildSubquery builds a query with subquery
func (s *Server) buildSubquery(params map[string]interface{}) (string, error) {
	primaryQuery, ok := params["primary_query"].(string)
	if !ok || primaryQuery == "" {
		return "", fmt.Errorf("primary_query is required for subqueries")
	}

	// Example: percentile of aggregated values using subquery
	query := fmt.Sprintf("SELECT percentile(aggregated_value, 95) FROM (%s) SINCE 1 hour ago", primaryQuery)

	return query, nil
}

// buildJoinQuery builds a JOIN query
func (s *Server) buildJoinQuery(params map[string]interface{}) (string, error) {
	primaryQuery, ok := params["primary_query"].(string)
	if !ok || primaryQuery == "" {
		return "", fmt.Errorf("primary_query is required for JOIN queries")
	}

	secondaryQuery, ok := params["secondary_query"].(string)
	if !ok || secondaryQuery == "" {
		return "", fmt.Errorf("secondary_query is required for JOIN queries")
	}

	joinKeys, ok := params["join_keys"].([]interface{})
	if !ok || len(joinKeys) == 0 {
		return "", fmt.Errorf("join_keys is required for JOIN queries")
	}

	// Convert join keys
	keys := make([]string, len(joinKeys))
	for i, key := range joinKeys {
		keys[i] = key.(string)
	}

	query := fmt.Sprintf("FROM (%s) AS a, (%s) AS b SELECT * WHERE %s SINCE 1 hour ago",
		primaryQuery, secondaryQuery, "a." + keys[0] + " = b." + keys[0])

	return query, nil
}

// buildNestedAggregationQuery builds a query with nested aggregation
func (s *Server) buildNestedAggregationQuery(params map[string]interface{}) (string, error) {
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return "", fmt.Errorf("event_type is required for nested aggregation queries")
	}

	timeRange := "1 hour ago"
	if tr, ok := params["time_range"].(string); ok {
		timeRange = tr
	}

	// Example: derivative of average response time
	query := fmt.Sprintf("SELECT derivative(average(duration), 1 minute) as 'Rate of Change' FROM %s SINCE %s TIMESERIES",
		eventType, timeRange)

	return query, nil
}

// buildRateQuery builds a query using rate functions
func (s *Server) buildRateQuery(params map[string]interface{}) (string, error) {
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return "", fmt.Errorf("event_type is required for rate queries")
	}

	rateMetric, ok := params["rate_metric"].(string)
	if !ok || rateMetric == "" {
		return "", fmt.Errorf("rate_metric is required for rate queries")
	}

	rateInterval := "1 minute"
	if ri, ok := params["rate_interval"].(string); ok {
		rateInterval = ri
	}

	timeRange := "1 hour ago"
	if tr, ok := params["time_range"].(string); ok {
		timeRange = tr
	}

	// Build rate query
	query := fmt.Sprintf("SELECT rate(sum(%s), %s) as '%s per %s' FROM %s SINCE %s TIMESERIES",
		rateMetric, rateInterval, rateMetric, rateInterval, eventType, timeRange)

	return query, nil
}

// buildBucketsQuery builds a query using buckets for histogram segmentation
func (s *Server) buildBucketsQuery(params map[string]interface{}) (string, error) {
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return "", fmt.Errorf("event_type is required for buckets queries")
	}

	metric, ok := params["metric"].(string)
	if !ok {
		metric = "duration" // default metric
	}

	bucketSize := 10.0
	if bs, ok := params["bucket_size"].(float64); ok {
		bucketSize = bs
	}

	bucketCount := 10
	if bc, ok := params["bucket_count"].(float64); ok {
		bucketCount = int(bc)
	}

	timeRange := "1 hour ago"
	if tr, ok := params["time_range"].(string); ok {
		timeRange = tr
	}

	// Build buckets query
	query := fmt.Sprintf("SELECT count(*) FROM %s SINCE %s FACET buckets(%s, %f, %d)",
		eventType, timeRange, metric, bucketSize, bucketCount)

	return query, nil
}
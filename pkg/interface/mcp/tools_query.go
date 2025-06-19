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

	// Use reflection to call QueryNRQL method
	// This avoids circular imports while still allowing real execution
	client := s.getNRClient()
	clientValue := reflect.ValueOf(client)
	method := clientValue.MethodByName("QueryNRQL")
	if !method.IsValid() {
		return nil, fmt.Errorf("QueryNRQL method not found on client")
	}

	// Call the method
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

	// Return the NRQLResult
	return results[0].Interface(), nil
}

// isValidAggregateFunction checks if a string is a valid NRQL aggregate function
func isValidAggregateFunction(fn string) bool {
	fn = strings.ToLower(strings.TrimSpace(fn))
	
	// List of valid NRQL aggregate functions
	validFunctions := []string{
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
		"apdex(",
		"histogram(",
		"keyset(",
		"eventType(",
		"filter(",
	}
	
	for _, valid := range validFunctions {
		if strings.HasPrefix(fn, valid) {
			return true
		}
	}
	
	return false
}
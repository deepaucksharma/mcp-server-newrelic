package mcp

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
)

// handleNRQLExecute implements adaptive NRQL query execution with schema validation
func (s *Server) handleNRQLExecute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required and must be a non-empty string")
	}

	accountID := int64(0)
	if aid, ok := params["account_id"].(float64); ok {
		accountID = int64(aid)
	}

	timeout := 30
	if t, ok := params["timeout"].(float64); ok && t > 0 {
		timeout = int(t)
	}

	includeMetadata := false
	if im, ok := params["include_metadata"].(bool); ok {
		includeMetadata = im
	}

	// Check mock mode
	if s.isMockMode() {
		return s.generateMockNRQLResult(query, includeMetadata), nil
	}

	// Get New Relic client (with account support)
	var client interface{}
	var err error
	
	if accountID > 0 {
		// Get client for specific account
		client, err = s.getNRClientWithAccount(fmt.Sprintf("%d", accountID))
		if err != nil {
			return nil, fmt.Errorf("failed to get client for account %d: %w", accountID, err)
		}
	} else {
		// Use default client
		client = s.getNRClient()
		if client == nil {
			return nil, fmt.Errorf("New Relic client not configured")
		}
	}

	// Step 1: Validate and adapt the query
	validation, err := s.validateAndAdaptQuery(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	if !validation.IsValid {
		return map[string]interface{}{
			"error":       "Invalid query",
			"details":     validation.Errors,
			"suggestions": validation.Suggestions,
		}, nil
	}

	// Use adapted query if available
	executionQuery := query
	if validation.AdaptedQuery != "" {
		executionQuery = validation.AdaptedQuery
	}

	// Step 2: Execute the query with timeout
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	startTime := time.Now()
	
	// Execute NRQL query using reflection to handle interface{}
	clientValue := reflect.ValueOf(client)
	method := clientValue.MethodByName("QueryNRQL")
	if !method.IsValid() {
		return nil, fmt.Errorf("QueryNRQL method not found on client")
	}
	
	// Call QueryNRQL method
	args := []reflect.Value{reflect.ValueOf(queryCtx), reflect.ValueOf(executionQuery)}
	results := method.Call(args)
	if len(results) != 2 {
		return nil, fmt.Errorf("unexpected return values from QueryNRQL")
	}
	
	// Extract result and error
	var result *newrelic.NRQLResult
	if !results[0].IsNil() {
		result = results[0].Interface().(*newrelic.NRQLResult)
	}
	if !results[1].IsNil() {
		err := results[1].Interface().(error)
		// Provide helpful error context
		return nil, s.enhanceQueryError(err, query, validation)
	}

	executionTime := time.Since(startTime)

	// Step 3: Process and validate results against schema
	processedResults, err := s.processQueryResults(result, validation.Schema, query)
	if err != nil {
		return nil, fmt.Errorf("failed to process results: %w", err)
	}

	// Step 4: Build response with optional metadata
	response := map[string]interface{}{
		"results": processedResults,
		"query":   executionQuery,
	}

	if includeMetadata {
		response["metadata"] = map[string]interface{}{
			"executionTime":    executionTime.Milliseconds(),
			"rowCount":         len(processedResults),
			"queryAdapted":     executionQuery != query,
			"originalQuery":    query,
			"schemaValidation": validation.Schema != nil,
			"accountId":        accountID,
			"performanceHints": s.generatePerformanceHints(query, executionTime, len(processedResults)),
		}
	}

	// Add warnings if any
	if len(validation.Warnings) > 0 {
		response["warnings"] = validation.Warnings
	}

	// Add schema information if discovered
	if validation.Schema != nil {
		response["schema"] = validation.Schema
	}

	return response, nil
}

// QueryValidation contains validation results and adaptations
type QueryValidation struct {
	IsValid      bool
	Errors       []string
	Warnings     []string
	Suggestions  []string
	AdaptedQuery string
	Schema       map[string]interface{}
}

// validateAndAdaptQuery validates the query and adapts it based on schema
func (s *Server) validateAndAdaptQuery(ctx context.Context, query string, accountID int64) (*QueryValidation, error) {
	validation := &QueryValidation{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
	}

	// Basic syntax validation
	if err := s.validateNRQLSyntaxBasic(query); err != nil {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, err.Error())
		return validation, nil
	}

	// Extract event type from query
	eventType := extractEventType(query)
	if eventType == "" {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, "Could not extract event type from query")
		return validation, nil
	}

	// Get schema information if available
	if s.discovery != nil {
		schema, err := s.discovery.ProfileSchema(ctx, eventType, discovery.ProfileDepthBasic)
		if err == nil && schema != nil {
			validation.Schema = map[string]interface{}{
				"eventType":  eventType,
				"attributes": schema.Attributes,
			}

			// Validate attributes in query against schema
			queryAttrs := extractAttributes(query)
			for _, attr := range queryAttrs {
				if !isAttributeInSchema(attr, schema) {
					// Try case-insensitive match
					if suggestion := findSimilarAttribute(attr, schema); suggestion != "" {
						validation.Warnings = append(validation.Warnings, 
							fmt.Sprintf("Attribute '%s' not found in schema, did you mean '%s'?", attr, suggestion))
						
						// Adapt the query
						validation.AdaptedQuery = strings.ReplaceAll(query, attr, suggestion)
					} else {
						validation.Warnings = append(validation.Warnings, 
							fmt.Sprintf("Attribute '%s' not found in schema for %s", attr, eventType))
					}
				}
			}
		}
	}

	// Check for common issues and suggest improvements
	improvements := s.analyzeQueryForImprovements(query, eventType)
	validation.Suggestions = append(validation.Suggestions, improvements...)

	// Add time range if missing
	if !hasTimeRange(query) {
		validation.Warnings = append(validation.Warnings, "Query missing SINCE clause, defaulting to last hour")
		if validation.AdaptedQuery == "" {
			validation.AdaptedQuery = query + " SINCE 1 hour ago"
		} else {
			validation.AdaptedQuery += " SINCE 1 hour ago"
		}
	}

	return validation, nil
}

// validateNRQLSyntaxBasic performs basic NRQL syntax validation
func (s *Server) validateNRQLSyntaxBasic(query string) error {
	// Remove extra whitespace
	query = strings.TrimSpace(query)
	
	// Check if query starts with SELECT
	if !strings.HasPrefix(strings.ToUpper(query), "SELECT") {
		return fmt.Errorf("NRQL query must start with SELECT")
	}

	// Check for required FROM clause
	if !regexp.MustCompile(`(?i)\bFROM\s+\w+`).MatchString(query) {
		return fmt.Errorf("NRQL query must include FROM clause")
	}

	// Check for balanced parentheses
	if !areParenthesesBalanced(query) {
		return fmt.Errorf("Unbalanced parentheses in query")
	}

	// Check for valid quotes
	if !areQuotesBalanced(query) {
		return fmt.Errorf("Unbalanced quotes in query")
	}

	return nil
}

// processQueryResults processes raw results and validates against schema
func (s *Server) processQueryResults(rawResults *newrelic.NRQLResult, schema map[string]interface{}, query string) ([]interface{}, error) {
	if rawResults == nil {
		return nil, fmt.Errorf("nil result")
	}
	
	results := make([]interface{}, len(rawResults.Results))
	for i, r := range rawResults.Results {
		results[i] = r
	}

	// If no schema validation needed, return as-is
	if schema == nil {
		return results, nil
	}

	// Process each result row
	processedResults := make([]interface{}, 0, len(results))
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			// Add data type hints based on schema
			processedResult := make(map[string]interface{})
			for key, value := range resultMap {
				processedResult[key] = value
				
				// Add type information if available in schema
				if attrs, ok := schema["attributes"].([]map[string]interface{}); ok {
					for _, attr := range attrs {
						if attr["name"] == key {
							processedResult[key+"_type"] = attr["inferredType"]
							break
						}
					}
				}
			}
			processedResults = append(processedResults, processedResult)
		} else {
			processedResults = append(processedResults, result)
		}
	}

	return processedResults, nil
}

// enhanceQueryError provides helpful context for query errors
func (s *Server) enhanceQueryError(err error, query string, validation *QueryValidation) error {
	errStr := err.Error()
	
	// Common error patterns and helpful messages
	errorPatterns := map[string]string{
		"Unknown attribute":     "Use discovery.explore_attributes to find valid attributes",
		"Unknown function":      "Check NRQL function documentation",
		"Syntax error":          "Review NRQL syntax, especially quotes and parentheses",
		"Timeout":               "Try reducing time range or simplifying the query",
		"Permission denied":     "Check account permissions for this event type",
		"Rate limit":            "Query rate limit exceeded, try again later",
	}

	for pattern, hint := range errorPatterns {
		if strings.Contains(errStr, pattern) {
			return fmt.Errorf("%s. Hint: %s", err, hint)
		}
	}

	// If we have validation suggestions, include them
	if len(validation.Suggestions) > 0 {
		return fmt.Errorf("%s. Suggestions: %s", err, strings.Join(validation.Suggestions, "; "))
	}

	return err
}

// analyzeQueryForImprovements suggests query optimizations
func (s *Server) analyzeQueryForImprovements(query string, eventType string) []string {
	suggestions := []string{}
	upperQuery := strings.ToUpper(query)

	// Check for SELECT *
	if strings.Contains(upperQuery, "SELECT *") {
		suggestions = append(suggestions, "Avoid SELECT *, specify needed fields for better performance")
	}

	// Check for missing LIMIT on non-aggregated queries
	if !strings.Contains(upperQuery, "LIMIT") && !hasAggregateFunction(query) {
		suggestions = append(suggestions, "Consider adding LIMIT to prevent large result sets")
	}

	// Check for inefficient time ranges
	if strings.Contains(query, "SINCE 7 days ago") || strings.Contains(query, "SINCE 30 days ago") {
		suggestions = append(suggestions, "Large time ranges may be slow, consider using shorter ranges or sampling")
	}

	// Check for high cardinality FACET without LIMIT
	if strings.Contains(upperQuery, "FACET") && !strings.Contains(upperQuery, "LIMIT") {
		suggestions = append(suggestions, "FACET without LIMIT may return many buckets, consider adding LIMIT")
	}

	// Suggest using WHERE clause for filtering
	if !strings.Contains(upperQuery, "WHERE") && eventType != "" {
		suggestions = append(suggestions, "Consider adding WHERE clause to filter results and improve performance")
	}

	// Check for complex nested functions
	if countNestingDepth(query) > 3 {
		suggestions = append(suggestions, "Complex nested functions may impact performance, consider simplifying")
	}

	return suggestions
}

// generatePerformanceHints provides performance insights based on execution
func (s *Server) generatePerformanceHints(query string, executionTime time.Duration, resultCount int) []string {
	hints := []string{}

	// Execution time hints
	if executionTime > 10*time.Second {
		hints = append(hints, fmt.Sprintf("Query took %v, consider optimization", executionTime))
	}

	// Result size hints
	if resultCount > 1000 {
		hints = append(hints, fmt.Sprintf("Large result set (%d rows), consider adding filters or aggregation", resultCount))
	}

	// Memory usage hints (estimated)
	if resultCount > 10000 {
		hints = append(hints, "Very large result set may cause memory issues in processing")
	}

	return hints
}

// Helper functions

func extractEventType(query string) string {
	re := regexp.MustCompile(`(?i)FROM\s+(\w+)`)
	matches := re.FindStringSubmatch(query)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractAttributes(query string) []string {
	// This is a simplified attribute extractor
	// In production, use proper NRQL parser
	attributes := []string{}
	
	// Remove string literals to avoid false matches
	cleanQuery := regexp.MustCompile(`'[^']*'`).ReplaceAllString(query, "")
	
	// Find potential attribute names (simplified)
	re := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_.]*)\b`)
	matches := re.FindAllStringSubmatch(cleanQuery, -1)
	
	// Filter out NRQL keywords
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "SINCE": true,
		"UNTIL": true, "LIMIT": true, "FACET": true, "ORDER": true,
		"BY": true, "ASC": true, "DESC": true, "AS": true,
		"AND": true, "OR": true, "NOT": true, "IN": true,
		"LIKE": true, "IS": true, "NULL": true, "TRUE": true, "FALSE": true,
		"WITH": true, "TIMESERIES": true, "COMPARE": true,
	}
	
	seen := make(map[string]bool)
	for _, match := range matches {
		attr := match[1]
		upperAttr := strings.ToUpper(attr)
		if !keywords[upperAttr] && !seen[attr] && !isNRQLFunction(attr) {
			attributes = append(attributes, attr)
			seen[attr] = true
		}
	}
	
	return attributes
}

func isAttributeInSchema(attr string, schema interface{}) bool {
	// Simplified schema check
	// In production, use proper schema validation
	return true
}

func findSimilarAttribute(attr string, schema interface{}) string {
	// Simplified attribute matching
	// In production, use fuzzy matching or Levenshtein distance
	return ""
}

func hasTimeRange(query string) bool {
	upperQuery := strings.ToUpper(query)
	return strings.Contains(upperQuery, "SINCE") || strings.Contains(upperQuery, "UNTIL")
}

func areParenthesesBalanced(query string) bool {
	count := 0
	for _, ch := range query {
		if ch == '(' {
			count++
		} else if ch == ')' {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

func areQuotesBalanced(query string) bool {
	singleQuotes := 0
	doubleQuotes := 0
	escaped := false
	
	for _, ch := range query {
		if escaped {
			escaped = false
			continue
		}
		
		if ch == '\\' {
			escaped = true
			continue
		}
		
		if ch == '\'' {
			singleQuotes++
		} else if ch == '"' {
			doubleQuotes++
		}
	}
	
	return singleQuotes%2 == 0 && doubleQuotes%2 == 0
}

func hasAggregateFunction(query string) bool {
	aggregateFuncs := []string{
		"count", "sum", "average", "avg", "min", "max", 
		"uniqueCount", "percentile", "stddev", "rate",
	}
	
	lowerQuery := strings.ToLower(query)
	for _, fn := range aggregateFuncs {
		if strings.Contains(lowerQuery, fn+"(") {
			return true
		}
	}
	return false
}

func countNestingDepth(query string) int {
	maxDepth := 0
	currentDepth := 0
	
	for _, ch := range query {
		if ch == '(' {
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		} else if ch == ')' {
			currentDepth--
		}
	}
	
	return maxDepth
}

func isNRQLFunction(name string) bool {
	functions := map[string]bool{
		"count": true, "sum": true, "average": true, "avg": true,
		"min": true, "max": true, "uniqueCount": true, "percentile": true,
		"histogram": true, "rate": true, "funnel": true, "filter": true,
		"apdex": true, "stddev": true, "eventType": true, "getField": true,
		"length": true, "numeric": true, "keyset": true, "uniques": true,
		"latest": true, "earliest": true, "percentage": true,
	}
	
	return functions[strings.ToLower(name)]
}

// generateMockNRQLResult creates realistic mock data for testing
func (s *Server) generateMockNRQLResult(query string, includeMetadata bool) map[string]interface{} {
	// Parse query to determine result structure
	upperQuery := strings.ToUpper(query)
	
	// Default mock results
	results := []interface{}{
		map[string]interface{}{
			"count":     1234,
			"timestamp": time.Now().Unix() * 1000,
		},
	}
	
	// Customize based on query type
	if strings.Contains(upperQuery, "FACET") {
		results = []interface{}{
			map[string]interface{}{
				"facet": []interface{}{"app-1"},
				"count": 567,
			},
			map[string]interface{}{
				"facet": []interface{}{"app-2"},
				"count": 432,
			},
			map[string]interface{}{
				"facet": []interface{}{"app-3"},
				"count": 235,
			},
		}
	} else if strings.Contains(upperQuery, "TIMESERIES") {
		now := time.Now()
		results = []interface{}{}
		for i := 0; i < 12; i++ {
			results = append(results, map[string]interface{}{
				"beginTimeSeconds": now.Add(-time.Duration(i)*5*time.Minute).Unix(),
				"endTimeSeconds":   now.Add(-time.Duration(i-1)*5*time.Minute).Unix(),
				"results": []interface{}{
					map[string]interface{}{
						"count": 100 + i*10,
					},
				},
			})
		}
	}
	
	response := map[string]interface{}{
		"results": results,
		"query":   query,
	}
	
	if includeMetadata {
		response["metadata"] = map[string]interface{}{
			"executionTime":    125,
			"rowCount":         len(results),
			"queryAdapted":     false,
			"originalQuery":    query,
			"schemaValidation": true,
			"performanceHints": []string{},
		}
	}
	
	return response
}

// Additional handler stubs that can be implemented later
func (s *Server) handleNRQLValidate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	validation, err := s.validateAndAdaptQuery(ctx, query, 0)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"isValid":     validation.IsValid,
		"errors":      validation.Errors,
		"warnings":    validation.Warnings,
		"suggestions": validation.Suggestions,
		"adapted":     validation.AdaptedQuery != "" && validation.AdaptedQuery != query,
	}, nil
}

func (s *Server) handleNRQLEstimateCost(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// time_range parameter could be used for more accurate cost estimation in the future
	// if tr, ok := params["time_range"].(string); ok {
	//     timeRange = tr
	// }

	frequency := "once"
	if freq, ok := params["execution_frequency"].(string); ok {
		frequency = freq
	}

	// Mock cost estimation
	baseCost := 1.0
	if strings.Contains(strings.ToUpper(query), "FACET") {
		baseCost *= 2
	}
	if hasAggregateFunction(query) {
		baseCost *= 1.5
	}

	multiplier := map[string]float64{
		"once":       1,
		"hourly":     24 * 30,
		"daily":      30,
		"continuous": 24 * 30 * 60,
	}

	return map[string]interface{}{
		"query":            query,
		"estimatedCost":    baseCost * multiplier[frequency],
		"costCategory":     getCostCategory(baseCost),
		"recommendations":  getCostRecommendations(baseCost, query),
		"impactAnalysis": map[string]interface{}{
			"dataVolume":      "medium",
			"processingTime":  fmt.Sprintf("~%dms", int(baseCost*100)),
			"memoryUsage":     "low",
		},
	}, nil
}

func getCostCategory(cost float64) string {
	switch {
	case cost < 1:
		return "low"
	case cost < 3:
		return "medium"
	case cost < 5:
		return "high"
	default:
		return "very_high"
	}
}

func getCostRecommendations(cost float64, query string) []string {
	recommendations := []string{}
	
	if cost > 3 {
		recommendations = append(recommendations, "Consider adding more specific WHERE clauses")
		recommendations = append(recommendations, "Use sampling with LIMIT for exploratory queries")
	}
	
	if strings.Contains(strings.ToUpper(query), "SELECT *") {
		recommendations = append(recommendations, "Select only required fields instead of *")
	}
	
	return recommendations
}

func (s *Server) handleNRQLBuildSelect(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// This would be implemented to build SELECT clauses programmatically
	return map[string]interface{}{
		"error": "Not implemented yet",
	}, nil
}
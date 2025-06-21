package mcp

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// handleDiscoveryExploreEventTypes discovers available event types
func (s *Server) handleDiscoveryExploreEventTypes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	pattern := ""
	if p, ok := params["pattern"].(string); ok {
		pattern = p
	}
	
	limit := 100
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}
	
	offset := 0
	if o, ok := params["offset"].(float64); ok {
		offset = int(o)
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("discovery.explore_event_types", params), nil
	}

	// Get account ID if specified
	accountID, _ := params["account_id"].(string)

	// Get New Relic client with account support
	_, err := s.getNRClientWithAccount(accountID)
	if err != nil {
		return nil, err
	}

	// Query for event types using SHOW EVENT TYPES
	query := "SHOW EVENT TYPES"
	
	// Execute the query
	result, err := s.executeNRQLQuery(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover event types: %w", err)
	}

	// Parse the results based on the type returned by executeNRQLQuery
	var results []map[string]interface{}
	
	// Handle both direct map result (from mock) and *newrelic.NRQLResult (from real API)
	switch v := result.(type) {
	case map[string]interface{}:
		// Mock result format
		if r, ok := v["results"].([]map[string]interface{}); ok {
			results = r
		} else {
			return nil, fmt.Errorf("no results found in response")
		}
	case interface{ GetResults() []map[string]interface{} }:
		// Real NRQLResult has a GetResults method or Results field
		// Use reflection to access Results field
		resultValue := reflect.ValueOf(result)
		if resultValue.Kind() == reflect.Ptr {
			resultValue = resultValue.Elem()
		}
		resultsField := resultValue.FieldByName("Results")
		if resultsField.IsValid() {
			if r, ok := resultsField.Interface().([]map[string]interface{}); ok {
				results = r
			}
		}
	default:
		// Try direct reflection as last resort for *newrelic.NRQLResult
		resultValue := reflect.ValueOf(result)
		if resultValue.Kind() == reflect.Ptr && !resultValue.IsNil() {
			resultValue = resultValue.Elem()
			resultsField := resultValue.FieldByName("Results")
			if resultsField.IsValid() && resultsField.Kind() == reflect.Slice {
				if r, ok := resultsField.Interface().([]map[string]interface{}); ok {
					results = r
				} else {
					return nil, fmt.Errorf("Results field has unexpected type: %T", resultsField.Interface())
				}
			} else {
				return nil, fmt.Errorf("no Results field found in %T", result)
			}
		} else {
			return nil, fmt.Errorf("unexpected result format: got %T", result)
		}
	}
	
	if results == nil {
		return nil, fmt.Errorf("no results found")
	}

	// Transform results to our format
	eventTypes := []map[string]interface{}{}
	for _, r := range results {
		eventType, ok := r["eventType"].(string)
		if !ok {
			continue
		}

		// Apply pattern filter if specified
		if pattern != "" && !strings.Contains(strings.ToLower(eventType), strings.ToLower(pattern)) {
			continue
		}

		// Get additional info with a count query
		countQuery := fmt.Sprintf("SELECT count(*) FROM %s SINCE 24 hours ago", eventType)
		countResult, err := s.executeNRQLQuery(ctx, countQuery, accountID)
		
		count := int64(0)
		if err == nil {
			if cr, ok := countResult.(map[string]interface{}); ok {
				if results, ok := cr["results"].([]map[string]interface{}); ok && len(results) > 0 {
					if c, ok := results[0]["count"].(float64); ok {
						count = int64(c)
					}
				}
			}
		}

		eventTypes = append(eventTypes, map[string]interface{}{
			"name":             eventType,
			"count":            count,
			"attributes":       0, // Would need keyset() query to get this
			"sample_timestamp": nil, // Would need another query
		})
	}

	// Apply pagination
	totalCount := len(eventTypes)
	start := offset
	end := offset + limit
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}
	paginatedEventTypes := eventTypes[start:end]
	
	response := map[string]interface{}{
		"event_types": paginatedEventTypes,
		"total":       len(paginatedEventTypes),
		"discovery_metadata": map[string]interface{}{
			"account_id":    accountID,
			"time_range":    "24 hours",
			"discovered_at": ctx.Value("timestamp"),
			"pagination": map[string]interface{}{
				"limit":      limit,
				"offset":     offset,
				"total_count": totalCount,
				"has_more":   end < totalCount,
			},
		},
	}
	
	// Add next offset if there are more results
	if end < totalCount {
		response["next_offset"] = end
	}
	
	return response, nil
}
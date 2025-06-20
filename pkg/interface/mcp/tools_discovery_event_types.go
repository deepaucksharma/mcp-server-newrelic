package mcp

import (
	"context"
	"fmt"
	"strings"
)

// handleDiscoveryExploreEventTypes discovers available event types
func (s *Server) handleDiscoveryExploreEventTypes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	pattern := ""
	if p, ok := params["pattern"].(string); ok {
		pattern = p
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("discovery.explore_event_types", params), nil
	}

	// Get account ID if specified
	accountID, _ := params["account_id"].(string)

	// Get New Relic client with account support
	nrClient, err := s.getNRClientWithAccount(accountID)
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

	// Parse the results
	nrqlResult, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	results, ok := nrqlResult["results"].([]map[string]interface{})
	if !ok {
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

	return map[string]interface{}{
		"event_types": eventTypes,
		"total":       len(eventTypes),
		"discovery_metadata": map[string]interface{}{
			"account_id":    accountID,
			"time_range":    "24 hours",
			"discovered_at": ctx.Value("timestamp"),
		},
	}, nil
}
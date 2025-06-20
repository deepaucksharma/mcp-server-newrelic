package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
)

// handleDiscoveryExploreAttributes discovers attributes for an event type
func (s *Server) handleDiscoveryExploreAttributesImpl(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	sampleSize, ok := params["sample_size"].(float64)
	if !ok || sampleSize <= 0 {
		sampleSize = 1000
	}

	showCoverage, ok := params["show_coverage"].(bool)
	if !ok {
		showCoverage = true
	}

	showExamples, ok := params["show_examples"].(bool)
	if !ok {
		showExamples = true
	}

	// Check mock mode
	if s.nrClient == nil {
		return s.handleDiscoveryExploreAttributes(ctx, params)
	}

	// Cast to proper client type
	client, ok := s.nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Step 1: Get sample events to extract keyset
	keysetQuery := fmt.Sprintf(`
		SELECT keyset() 
		FROM %s 
		LIMIT %d 
		SINCE 1 hour ago
	`, eventType, int(sampleSize))

	keysetResult, err := client.QueryNRDB(ctx, keysetQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query keyset: %w", err)
	}

	// Extract unique attributes from all samples
	attributeMap := make(map[string]bool)
	if results, ok := keysetResult["results"].([]interface{}); ok {
		for _, result := range results {
			if res, ok := result.(map[string]interface{}); ok {
				if keyset, ok := res["keyset"].([]interface{}); ok {
					for _, key := range keyset {
						if keyStr, ok := key.(string); ok {
							attributeMap[keyStr] = true
						}
					}
				}
			}
		}
	}

	// Convert to sorted list
	attributes := make([]string, 0, len(attributeMap))
	for attr := range attributeMap {
		attributes = append(attributes, attr)
	}
	sort.Strings(attributes)

	// Step 2: For each attribute, get coverage and examples
	attributeDetails := make([]map[string]interface{}, 0, len(attributes))
	
	for _, attr := range attributes {
		detail := map[string]interface{}{
			"name": attr,
		}

		if showCoverage || showExamples {
			// Query for coverage and examples
			detailQuery := fmt.Sprintf(`
				SELECT 
					count(*) as total,
					filter(count(*), WHERE %s IS NOT NULL) as nonNullCount,
					uniqueCount(%s) as cardinality
				FROM %s 
				LIMIT %d
				SINCE 1 hour ago
			`, attr, attr, eventType, int(sampleSize))

			detailResult, err := client.QueryNRDB(ctx, detailQuery)
			if err == nil && len(detailResult["results"].([]interface{})) > 0 {
				if res, ok := detailResult["results"].([]interface{})[0].(map[string]interface{}); ok {
					total := getFloat64(res, "total")
					nonNull := getFloat64(res, "nonNullCount")
					cardinality := getFloat64(res, "cardinality")

					if showCoverage && total > 0 {
						detail["coverage"] = nonNull / total
						detail["cardinality"] = int(cardinality)
					}
				}
			}

			if showExamples {
				// Get example values
				exampleQuery := fmt.Sprintf(`
					SELECT uniques(%s, 5) as examples
					FROM %s 
					WHERE %s IS NOT NULL
					LIMIT 1000
					SINCE 1 hour ago
				`, attr, eventType, attr)

				exampleResult, err := client.QueryNRDB(ctx, exampleQuery)
				if err == nil && len(exampleResult["results"].([]interface{})) > 0 {
					if res, ok := exampleResult["results"].([]interface{})[0].(map[string]interface{}); ok {
						if examples, ok := res["examples"].([]interface{}); ok {
							detail["examples"] = examples
						}
					}
				}
			}
		}

		// Infer data type from attribute name and examples
		detail["inferredType"] = inferDataType(attr, detail["examples"])

		attributeDetails = append(attributeDetails, detail)
	}

	// Calculate data quality score
	qualityScore := calculateAttributeQualityScore(attributeDetails)

	// Generate recommendations
	recommendations := generateAttributeRecommendations(eventType, attributeDetails)

	return map[string]interface{}{
		"eventType":           eventType,
		"attributes":          attributeDetails,
		"totalAttributes":     len(attributes),
		"sampleSize":          int(sampleSize),
		"discoveryMethod":     "keyset() analysis",
		"dataQualityScore":    qualityScore,
		"recommendations":     recommendations,
		"discoveryTimestamp":  time.Now().UTC(),
	}, nil
}

// Helper function to infer data type from attribute name and examples
func inferDataType(attr string, examples interface{}) string {
	// Check common patterns in attribute names
	lowerAttr := strings.ToLower(attr)
	
	if strings.Contains(lowerAttr, "timestamp") || strings.Contains(lowerAttr, "time") || strings.Contains(lowerAttr, "date") {
		return "timestamp"
	}
	if strings.Contains(lowerAttr, "duration") || strings.Contains(lowerAttr, "elapsed") {
		return "numeric"
	}
	if strings.Contains(lowerAttr, "count") || strings.Contains(lowerAttr, "size") || strings.Contains(lowerAttr, "percent") {
		return "numeric"
	}
	if strings.Contains(lowerAttr, "error") || strings.Contains(lowerAttr, "success") {
		return "boolean"
	}
	if strings.Contains(lowerAttr, "name") || strings.Contains(lowerAttr, "id") || strings.Contains(lowerAttr, "type") {
		return "string"
	}

	// Check examples if available
	if exampleList, ok := examples.([]interface{}); ok && len(exampleList) > 0 {
		// Check first example type
		switch exampleList[0].(type) {
		case float64, int, int64:
			return "numeric"
		case bool:
			return "boolean"
		case string:
			return "string"
		}
	}

	return "unknown"
}

// Calculate overall data quality score based on attributes
func calculateAttributeQualityScore(attributes []map[string]interface{}) float64 {
	if len(attributes) == 0 {
		return 0.0
	}

	totalScore := 0.0
	criticalAttrs := map[string]bool{
		"timestamp": true,
		"appName":   true,
		"host":      true,
		"duration":  true,
		"error":     true,
	}

	criticalFound := 0
	highCoverageCount := 0

	for _, attr := range attributes {
		name := attr["name"].(string)
		
		// Check if critical attribute exists
		if criticalAttrs[name] {
			criticalFound++
		}

		// Check coverage
		if coverage, ok := attr["coverage"].(float64); ok && coverage > 0.9 {
			highCoverageCount++
		}
	}

	// Score based on critical attributes found
	criticalScore := float64(criticalFound) / float64(len(criticalAttrs))
	
	// Score based on high coverage attributes
	coverageScore := float64(highCoverageCount) / float64(len(attributes))
	
	// Weighted average
	totalScore = (criticalScore * 0.6) + (coverageScore * 0.4)

	return totalScore
}

// Generate recommendations based on discovered attributes
func generateAttributeRecommendations(eventType string, attributes []map[string]interface{}) []string {
	recommendations := []string{}

	// Check for critical attributes
	hasTimestamp := false
	hasIdentifier := false
	hasError := false
	lowCoverageAttrs := []string{}
	highCardinalityAttrs := []string{}

	for _, attr := range attributes {
		name := attr["name"].(string)
		
		if name == "timestamp" {
			hasTimestamp = true
		}
		if name == "appName" || name == "service" || name == "entityName" {
			hasIdentifier = true
		}
		if name == "error" || name == "error.class" || name == "error.message" {
			hasError = true
		}

		if coverage, ok := attr["coverage"].(float64); ok && coverage < 0.5 {
			lowCoverageAttrs = append(lowCoverageAttrs, name)
		}

		if cardinality, ok := attr["cardinality"].(int); ok && cardinality > 1000 {
			highCardinalityAttrs = append(highCardinalityAttrs, name)
		}
	}

	if !hasTimestamp {
		recommendations = append(recommendations, "No timestamp attribute found - consider adding timing information")
	}
	if !hasIdentifier {
		recommendations = append(recommendations, "No service identifier found - consider adding appName or service attribute")
	}
	if !hasError {
		recommendations = append(recommendations, "No error tracking attribute found - consider adding error information")
	}

	if len(lowCoverageAttrs) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Low coverage attributes (%s) may indicate data collection issues", strings.Join(lowCoverageAttrs[:min(3, len(lowCoverageAttrs))], ", ")))
	}

	if len(highCardinalityAttrs) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("High cardinality attributes (%s) may impact query performance", strings.Join(highCardinalityAttrs[:min(3, len(highCardinalityAttrs))], ", ")))
	}

	return recommendations
}

// Helper functions
func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
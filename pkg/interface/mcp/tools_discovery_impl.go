package mcp

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
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
	if s.isMockMode() {
		return s.getMockData("discovery.explore_attributes", params), nil
	}

	// No need to get client directly - we'll use executeNRQLQuery

	// Step 1: Get sample events to extract keyset
	keysetQuery := fmt.Sprintf(`
		SELECT keyset() 
		FROM %s 
		LIMIT %d 
		SINCE 1 hour ago
	`, eventType, int(sampleSize))

	keysetResult, err := s.executeNRQLQuery(ctx, keysetQuery, "")
	if err != nil {
		return nil, fmt.Errorf("failed to query keyset: %w", err)
	}

	// Extract unique attributes from all samples
	attributeMap := make(map[string]bool)
	
	// Handle the result based on type (similar to tools_discovery_event_types.go)
	var results []map[string]interface{}
	switch v := keysetResult.(type) {
	case map[string]interface{}:
		if r, ok := v["results"].([]map[string]interface{}); ok {
			results = r
		}
	default:
		// Use reflection for *newrelic.NRQLResult
		resultValue := reflect.ValueOf(keysetResult)
		if resultValue.Kind() == reflect.Ptr {
			resultValue = resultValue.Elem()
		}
		resultsField := resultValue.FieldByName("Results")
		if resultsField.IsValid() {
			if r, ok := resultsField.Interface().([]map[string]interface{}); ok {
				results = r
			}
		}
	}
	
	// Extract attributes from keyset
	keysetFound := false
	if len(results) > 0 {
		for _, result := range results {
			res := result
			if keyset, ok := res["keyset"].([]interface{}); ok && len(keyset) > 0 {
				keysetFound = true
				for _, key := range keyset {
					if keyStr, ok := key.(string); ok {
						attributeMap[keyStr] = true
					}
				}
			}
		}
	}

	// If keyset() didn't return results, fallback to SELECT * approach
	if !keysetFound || len(attributeMap) == 0 {
		// Query sample events and extract keys
		sampleQuery := fmt.Sprintf(`
			SELECT * 
			FROM %s 
			LIMIT 10 
			SINCE 1 hour ago
		`, eventType)

		sampleResult, err := s.executeNRQLQuery(ctx, sampleQuery, "")
		if err == nil {
			// Extract results
			var sampleResults []map[string]interface{}
			switch v := sampleResult.(type) {
			case map[string]interface{}:
				if r, ok := v["results"].([]map[string]interface{}); ok {
					sampleResults = r
				}
			default:
				// Use reflection for *newrelic.NRQLResult
				resultValue := reflect.ValueOf(sampleResult)
				if resultValue.Kind() == reflect.Ptr {
					resultValue = resultValue.Elem()
				}
				resultsField := resultValue.FieldByName("Results")
				if resultsField.IsValid() {
					if r, ok := resultsField.Interface().([]map[string]interface{}); ok {
						sampleResults = r
					}
				}
			}

			// Extract all keys from sample events
			for _, event := range sampleResults {
				for key := range event {
					attributeMap[key] = true
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
			// Query for coverage - using backticks to handle special characters in attribute names
			detailQuery := fmt.Sprintf(`
				SELECT 
					count(*) as total,
					filter(count(*), WHERE %s IS NOT NULL) as nonNullCount
				FROM %s 
				LIMIT %d
				SINCE 1 hour ago
			`, fmt.Sprintf("`%s`", attr), eventType, int(sampleSize))

			detailResult, err := s.executeNRQLQuery(ctx, detailQuery, "")
			if err == nil && detailResult != nil {
				// Parse result
				var detailResults []map[string]interface{}
				switch v := detailResult.(type) {
				case map[string]interface{}:
					if r, ok := v["results"].([]map[string]interface{}); ok {
						detailResults = r
					}
				default:
					// Use reflection for *newrelic.NRQLResult
					resultValue := reflect.ValueOf(detailResult)
					if resultValue.Kind() == reflect.Ptr {
						resultValue = resultValue.Elem()
					}
					resultsField := resultValue.FieldByName("Results")
					if resultsField.IsValid() {
						if r, ok := resultsField.Interface().([]map[string]interface{}); ok {
							detailResults = r
						}
					}
				}
				
				if len(detailResults) > 0 {
					res := detailResults[0]
				total := getFloat64(res, "total")
				nonNull := getFloat64(res, "nonNullCount")

					if showCoverage && total > 0 {
						detail["coverage"] = nonNull / total
						// Cardinality would need a separate query - skip for now
						detail["has_data"] = nonNull > 0
					}
				}
			}

			if showExamples && detail["has_data"] == true {
				// Get example values - using a simpler approach
				exampleQuery := fmt.Sprintf(`
					SELECT %s as example
					FROM %s 
					WHERE %s IS NOT NULL
					LIMIT 5
					SINCE 1 hour ago
				`, fmt.Sprintf("`%s`", attr), eventType, fmt.Sprintf("`%s`", attr))

				exampleResult, err := s.executeNRQLQuery(ctx, exampleQuery, "")
				if err == nil && exampleResult != nil {
					// Parse result
					var exampleResults []map[string]interface{}
					switch v := exampleResult.(type) {
					case map[string]interface{}:
						if r, ok := v["results"].([]map[string]interface{}); ok {
							exampleResults = r
						}
					default:
						// Use reflection for *newrelic.NRQLResult
						resultValue := reflect.ValueOf(exampleResult)
						if resultValue.Kind() == reflect.Ptr {
							resultValue = resultValue.Elem()
						}
						resultsField := resultValue.FieldByName("Results")
						if resultsField.IsValid() {
							if r, ok := resultsField.Interface().([]map[string]interface{}); ok {
								exampleResults = r
							}
						}
					}
					
					if len(exampleResults) > 0 {
						// Collect unique examples
						examples := make([]interface{}, 0, 5)
						seen := make(map[string]bool)
						for _, res := range exampleResults {
							if example, ok := res["example"]; ok && example != nil {
								// Convert to string for deduplication
								exampleStr := fmt.Sprintf("%v", example)
								if !seen[exampleStr] {
									seen[exampleStr] = true
									examples = append(examples, example)
								}
							}
						}
						if len(examples) > 0 {
							detail["examples"] = examples
						}
					}
				}
			}
		}

		// Infer data type from attribute name and examples
		detail["type"] = inferDataType(attr, detail["examples"])

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
		case float64:
			return "numeric"
		case int:
			return "numeric"
		case int64:
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
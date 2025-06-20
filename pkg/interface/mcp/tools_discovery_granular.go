package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// registerGranularDiscoveryTools registers atomic discovery tools that enable exploration without assumptions
func (s *Server) registerGranularDiscoveryTools() error {
	// 1. Schema Discovery Tools
	if err := s.registerSchemaDiscoveryTools(); err != nil {
		return err
	}

	// 2. Data Profiling Tools
	if err := s.registerDataProfilingTools(); err != nil {
		return err
	}

	// 3. Pattern Discovery Tools
	if err := s.registerPatternDiscoveryTools(); err != nil {
		return err
	}

	// 4. Relationship Discovery Tools
	if err := s.registerRelationshipDiscoveryTools(); err != nil {
		return err
	}

	// 5. Data Quality Tools
	if err := s.registerDataQualityTools(); err != nil {
		return err
	}

	return nil
}

// 1. SCHEMA DISCOVERY TOOLS

func (s *Server) registerSchemaDiscoveryTools() error {
	// Discover what event types exist
	exploreEventTypes := NewToolBuilder("discovery.explore_event_types", "Discover what event types exist in NRDB without assumptions").
		Category(CategoryQuery).
		Handler(s.handleDiscoveryExploreEventTypes).
		Param("time_range", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "How far back to explore (e.g., '7 days', '1 hour')",
				Default:     "24 hours",
			},
			ValidationRules: []ValidationRule{
				{Field: "time_range", Rule: "regex", Value: `^\d+\s+(minute|hour|day|week|month)s?$`},
			},
		}).
		Param("include_samples", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Include sample events for each type",
				Default:     true,
			},
		}).
		Param("min_event_count", EnhancedProperty{
			Property: Property{
				Type:        "integer",
				Description: "Minimum events to consider type active",
				Default:     10,
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 1000
			p.MaxLatencyMS = 5000
			p.Cacheable = true
			p.CacheTTLSeconds = 3600
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Start any investigation: discovery.explore_event_types()",
				"Find recent data: discovery.explore_event_types(time_range='1 hour')",
			}
			g.ChainsWith = []string{"discovery.explore_attributes", "discovery.profile_data_completeness"}
			g.SuccessIndicators = []string{"Returns list of event types with counts", "Shows data freshness"}
			g.ContextRequirements = []string{"No assumptions about what data exists"}
		}).
		Example(ToolExample{
			Name:        "Discover available data",
			Description: "First step in any investigation",
			Params: map[string]interface{}{
				"time_range":       "6 hours",
				"include_samples":  true,
				"min_event_count": 100,
			},
		}).
		Build()

	if err := s.tools.Register(exploreEventTypes.Tool); err != nil {
		return err
	}

	// Explore attributes for an event type
	exploreAttributes := NewToolBuilder("discovery.explore_attributes", "Discover what attributes exist for an event type").
		Category(CategoryQuery).
		Handler(s.handleDiscoveryExploreAttributes).
		Required("event_type").
		Param("event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Event type to explore",
			},
			Examples: []interface{}{"Transaction", "SystemSample", "Log"},
		}).
		Param("sample_size", EnhancedProperty{
			Property: Property{
				Type:        "integer",
				Description: "Number of events to sample",
				Default:     1000,
			},
		}).
		Param("show_coverage", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Calculate percentage of non-null values",
				Default:     true,
			},
		}).
		Param("show_examples", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Show example values for each attribute",
				Default:     true,
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 500
			p.Cacheable = true
			p.CacheTTLSeconds = 1800
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Understand event structure: discovery.explore_attributes(event_type='Transaction')",
			}
			g.ChainsWith = []string{"discovery.profile_attribute_values", "discovery.find_natural_groupings"}
			g.WarningsForAI = []string{
				"Don't assume attributes exist - check coverage percentage",
				"Some attributes may be sparse - look at null percentages",
			}
		}).
		Build()

	return s.tools.Register(exploreAttributes.Tool)
}

// 2. DATA PROFILING TOOLS

func (s *Server) registerDataProfilingTools() error {
	// Profile data completeness
	profileCompleteness := NewToolBuilder("discovery.profile_data_completeness", "Analyze how complete and reliable the data is").
		Category(CategoryAnalysis).
		Handler(s.handleDiscoveryProfileCompleteness).
		Required("event_type").
		Param("event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Event type to profile",
			},
		}).
		Param("critical_attributes", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Attributes that should always be present",
				Items: &Property{
					Type: "string",
				},
			},
			Examples: []interface{}{
				[]string{"appName", "host", "duration"},
			},
		}).
		Param("time_range", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Time range to analyze",
				Default:     "24 hours",
			},
		}).
		Param("check_patterns", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Check for collection patterns (gaps, periodicity)",
				Default:     true,
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 2000
			p.MaxLatencyMS = 10000
			p.ResourceIntensive = true
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Verify data quality: discovery.profile_data_completeness(event_type='Transaction', critical_attributes=['duration', 'appName'])",
			}
			g.SuccessIndicators = []string{
				"Shows percentage of complete records",
				"Identifies gaps in data collection",
				"Reveals collection patterns",
			}
		}).
		Build()

	if err := s.tools.Register(profileCompleteness.Tool); err != nil {
		return err
	}

	// Profile attribute value distribution
	profileAttributeValues := NewToolBuilder("discovery.profile_attribute_values", "Understand the distribution and patterns of attribute values").
		Category(CategoryAnalysis).
		Handler(s.handleDiscoveryProfileAttributeValues).
		Required("event_type", "attribute").
		Param("event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Event type containing the attribute",
			},
		}).
		Param("attribute", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Attribute to profile",
			},
		}).
		Param("profile_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Type of profiling to perform",
				Enum:        []string{"distribution", "cardinality", "patterns", "anomalies", "all"},
				Default:     "all",
			},
		}).
		Param("time_comparison", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Compare current values with historical",
				Default:     true,
			},
		}).
		Build()

	return s.tools.Register(profileAttributeValues.Tool)
}

// 3. PATTERN DISCOVERY TOOLS

func (s *Server) registerPatternDiscoveryTools() error {
	// Find natural groupings in data
	findNaturalGroupings := NewToolBuilder("discovery.find_natural_groupings", "Discover how data naturally groups without assumptions").
		Category(CategoryAnalysis).
		Handler(s.handleDiscoveryFindNaturalGroupings).
		Required("event_type").
		Param("event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Event type to analyze",
			},
		}).
		Param("max_groups", EnhancedProperty{
			Property: Property{
				Type:        "integer",
				Description: "Maximum number of groupings to find",
				Default:     10,
			},
		}).
		Param("min_group_size", EnhancedProperty{
			Property: Property{
				Type:        "integer",
				Description: "Minimum events in a group to be significant",
				Default:     100,
			},
		}).
		Param("attributes_to_consider", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Specific attributes to consider (null = auto-discover)",
				Items: &Property{
					Type: "string",
				},
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 3000
			p.MaxLatencyMS = 15000
			p.ResourceIntensive = true
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Find how to segment data: discovery.find_natural_groupings(event_type='Transaction')",
			}
			g.ChainsWith = []string{"nrql.build_facet", "analysis.segment_comparison"}
			g.SuccessIndicators = []string{
				"Returns meaningful ways to group data",
				"Shows group sizes and characteristics",
			}
		}).
		Build()

	if err := s.tools.Register(findNaturalGroupings.Tool); err != nil {
		return err
	}

	// Detect temporal patterns
	detectTemporalPatterns := NewToolBuilder("discovery.detect_temporal_patterns", "Find time-based patterns in data").
		Category(CategoryAnalysis).
		Handler(s.handleDiscoveryDetectTemporalPatterns).
		Required("query").
		Param("query", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Base NRQL query to analyze for patterns",
			},
			Examples: []interface{}{
				"SELECT count(*) FROM Transaction",
				"SELECT average(duration) FROM Transaction",
			},
		}).
		Param("pattern_types", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Types of patterns to detect",
				Items: &Property{
					Type: "string",
					Enum: []string{"hourly", "daily", "weekly", "periodic", "trending"},
				},
				Default: []string{"hourly", "daily", "weekly"},
			},
		}).
		Param("lookback_days", EnhancedProperty{
			Property: Property{
				Type:        "integer",
				Description: "Days of history to analyze",
				Default:     30,
			},
		}).
		Build()

	return s.tools.Register(detectTemporalPatterns.Tool)
}

// 4. RELATIONSHIP DISCOVERY TOOLS

func (s *Server) registerRelationshipDiscoveryTools() error {
	// Find data relationships
	findDataRelationships := NewToolBuilder("discovery.find_data_relationships", "Discover how different data types relate without assumptions").
		Category(CategoryAnalysis).
		Handler(s.handleDiscoveryFindDataRelationships).
		Required("source_event_type").
		Param("source_event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Primary event type",
			},
		}).
		Param("target_event_types", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Event types to check relationships with (null = all)",
				Items: &Property{
					Type: "string",
				},
			},
		}).
		Param("relationship_types", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Types of relationships to find",
				Items: &Property{
					Type: "string",
					Enum: []string{"join_key", "temporal", "correlation", "causation"},
				},
				Default: []string{"join_key", "temporal"},
			},
		}).
		Param("sample_size", EnhancedProperty{
			Property: Property{
				Type:        "integer",
				Description: "Number of events to sample for analysis",
				Default:     10000,
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 5000
			p.MaxLatencyMS = 30000
			p.ResourceIntensive = true
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Find how data connects: discovery.find_data_relationships(source_event_type='Transaction')",
			}
			g.SuccessIndicators = []string{
				"Returns joinable attributes",
				"Shows temporal relationships",
				"Identifies correlated metrics",
			}
			g.WarningsForAI = []string{
				"Resource intensive - use specific target types when possible",
				"Correlation does not imply causation",
			}
		}).
		Build()

	return s.tools.Register(findDataRelationships.Tool)
}

// 5. DATA QUALITY TOOLS

func (s *Server) registerDataQualityTools() error {
	// Detect data anomalies
	detectDataAnomalies := NewToolBuilder("discovery.detect_data_anomalies", "Find anomalies in data collection or quality").
		Category(CategoryAnalysis).
		Handler(s.handleDiscoveryDetectDataAnomalies).
		Required("event_type").
		Param("event_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Event type to check for anomalies",
			},
		}).
		Param("anomaly_types", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Types of anomalies to detect",
				Items: &Property{
					Type: "string",
					Enum: []string{"gaps", "spikes", "schema_changes", "value_anomalies", "collection_issues"},
				},
				Default: []string{"gaps", "spikes", "schema_changes"},
			},
		}).
		Param("sensitivity", EnhancedProperty{
			Property: Property{
				Type:        "number",
				Description: "Detection sensitivity (0-1)",
				Default:     0.8,
			},
		}).
		Build()

	if err := s.tools.Register(detectDataAnomalies.Tool); err != nil {
		return err
	}

	// Validate data assumptions
	validateDataAssumptions := NewToolBuilder("discovery.validate_assumptions", "Test assumptions about data existence and quality").
		Category(CategoryAnalysis).
		Handler(s.handleDiscoveryValidateAssumptions).
		Required("assumptions").
		Param("assumptions", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "List of assumptions to validate",
				Items: &Property{
					Type: "object",
				},
			},
			Examples: []interface{}{
				[]map[string]interface{}{
					{
						"type":        "attribute_exists",
						"event_type":  "Transaction",
						"attribute":   "duration",
						"coverage":    0.95,
					},
					{
						"type":       "value_range",
						"event_type": "SystemSample",
						"attribute":  "cpuPercent",
						"min":        0,
						"max":        100,
					},
				},
			},
		}).
		Build()

	return s.tools.Register(validateDataAssumptions.Tool)
}

// Handler implementations

func (s *Server) handleDiscoveryExploreEventTypes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	timeRange, _ := params["time_range"].(string)
	includeSamples, _ := params["include_samples"].(bool)
	minEventCount, _ := params["min_event_count"].(float64)

	if timeRange == "" {
		timeRange = "24 hours"
	}

	// First, get all event types
	eventTypesQuery := fmt.Sprintf("SHOW EVENT TYPES SINCE %s", timeRange)
	
	// Mock implementation - in real implementation, execute against NRDB
	eventTypes := []map[string]interface{}{
		{
			"eventType":    "Transaction",
			"count":        1234567,
			"firstSeen":    time.Now().Add(-7 * 24 * time.Hour),
			"lastSeen":     time.Now().Add(-5 * time.Minute),
			"sampleEvent":  nil,
		},
		{
			"eventType":    "SystemSample",
			"count":        987654,
			"firstSeen":    time.Now().Add(-30 * 24 * time.Hour),
			"lastSeen":     time.Now().Add(-1 * time.Minute),
			"sampleEvent":  nil,
		},
		{
			"eventType":    "Log",
			"count":        554433,
			"firstSeen":    time.Now().Add(-14 * 24 * time.Hour),
			"lastSeen":     time.Now().Add(-30 * time.Second),
			"sampleEvent":  nil,
		},
	}

	// Filter by minimum count
	filtered := []map[string]interface{}{}
	for _, et := range eventTypes {
		if count, ok := et["count"].(int); ok && float64(count) >= minEventCount {
			filtered = append(filtered, et)
		}
	}

	// Get samples if requested
	if includeSamples {
		for i, et := range filtered {
			eventType := et["eventType"].(string)
			sampleQuery := fmt.Sprintf("SELECT * FROM %s LIMIT 1 SINCE %s", eventType, timeRange)
			// Mock sample
			filtered[i]["sampleEvent"] = map[string]interface{}{
				"timestamp": time.Now().Unix(),
				"appName":   "sample-app",
				"duration":  123.45,
			}
		}
	}

	return map[string]interface{}{
		"eventTypes":       filtered,
		"totalTypes":       len(filtered),
		"timeRangeUsed":    timeRange,
		"discoveryMethod":  "SHOW EVENT TYPES",
		"dataCompleteness": calculateDataCompleteness(filtered),
	}, nil
}

func (s *Server) handleDiscoveryExploreAttributes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	eventType, _ := params["event_type"].(string)
	sampleSize, _ := params["sample_size"].(float64)
	showCoverage, _ := params["show_coverage"].(bool)
	showExamples, _ := params["show_examples"].(bool)

	if sampleSize == 0 {
		sampleSize = 1000
	}

	// Mock implementation
	attributes := []map[string]interface{}{
		{
			"name":         "duration",
			"type":         "numeric",
			"coverage":     0.98,
			"nullPercent":  0.02,
			"cardinality":  "continuous",
			"examples":     []interface{}{123.45, 234.56, 345.67},
			"statistics": map[string]interface{}{
				"min":    0.1,
				"max":    5000.0,
				"avg":    234.5,
				"stddev": 123.4,
			},
		},
		{
			"name":         "appName",
			"type":         "string",
			"coverage":     1.0,
			"nullPercent":  0.0,
			"cardinality":  12,
			"uniqueValues": 12,
			"examples":     []interface{}{"checkout-api", "payment-service", "user-service"},
		},
		{
			"name":         "error",
			"type":         "boolean",
			"coverage":     0.95,
			"nullPercent":  0.05,
			"cardinality":  2,
			"distribution": map[string]interface{}{
				"true":  0.02,
				"false": 0.93,
				"null":  0.05,
			},
		},
	}

	return map[string]interface{}{
		"eventType":           eventType,
		"attributes":          attributes,
		"totalAttributes":     len(attributes),
		"sampleSize":          int(sampleSize),
		"discoveryMethod":     "keyset() sampling",
		"dataQualityScore":    0.92,
		"recommendations": []string{
			"High coverage on critical attributes (duration, appName)",
			"Consider investigating 5% null rate on 'error' attribute",
			"Low cardinality on appName (12) makes it good for FACET",
		},
	}, nil
}

func (s *Server) handleDiscoveryFindNaturalGroupings(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	eventType, _ := params["event_type"].(string)
	maxGroups, _ := params["max_groups"].(float64)
	minGroupSize, _ := params["min_group_size"].(float64)

	// Mock implementation
	groupings := []map[string]interface{}{
		{
			"groupingAttribute": "appName",
			"groupingQuality":   0.95,
			"groups": []map[string]interface{}{
				{"value": "checkout-api", "count": 45000, "percentage": 0.35},
				{"value": "payment-service", "count": 32000, "percentage": 0.25},
				{"value": "user-service", "count": 28000, "percentage": 0.22},
			},
			"reasoning": "Clear separation by application with balanced distribution",
		},
		{
			"groupingAttribute": "response.status",
			"groupingQuality":   0.88,
			"groups": []map[string]interface{}{
				{"value": "200", "count": 95000, "percentage": 0.74},
				{"value": "404", "count": 15000, "percentage": 0.12},
				{"value": "500", "count": 8000, "percentage": 0.06},
			},
			"reasoning": "Status codes show clear success/failure patterns",
		},
		{
			"groupingAttribute": "host + appName",
			"groupingQuality":   0.82,
			"groups": []map[string]interface{}{
				{"value": "host1:checkout-api", "count": 22000, "percentage": 0.17},
				{"value": "host2:checkout-api", "count": 23000, "percentage": 0.18},
			},
			"reasoning": "Host-app combinations reveal deployment patterns",
		},
	}

	return map[string]interface{}{
		"eventType":            eventType,
		"naturalGroupings":     groupings[:int(maxGroups)],
		"discoveryMethod":      "entropy and distribution analysis",
		"minGroupSizeUsed":     int(minGroupSize),
		"suggestedFacets": []string{
			"appName",
			"response.status",
			"CASES(WHERE duration > 1000 as 'Slow', WHERE duration > 100 as 'Normal', WHERE duration <= 100 as 'Fast')",
		},
	}, nil
}

func (s *Server) handleDiscoveryDetectDataAnomalies(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	eventType, _ := params["event_type"].(string)
	anomalyTypes, _ := params["anomaly_types"].([]interface{})
	sensitivity, _ := params["sensitivity"].(float64)

	// Mock implementation
	anomalies := []map[string]interface{}{
		{
			"type":        "data_gap",
			"severity":    "high",
			"timeRange":   "2024-01-20T14:00:00Z to 2024-01-20T14:15:00Z",
			"description": "No data received for 15 minutes",
			"impact":      "Missing data during this period affects accuracy",
			"evidence": map[string]interface{}{
				"expected_rate": "1000 events/minute",
				"actual_rate":   "0 events/minute",
			},
		},
		{
			"type":        "schema_change",
			"severity":    "medium",
			"timestamp":   "2024-01-20T10:00:00Z",
			"description": "New attribute 'trace.id' appeared",
			"impact":      "May indicate instrumentation update",
			"evidence": map[string]interface{}{
				"before": []string{"duration", "appName", "host"},
				"after":  []string{"duration", "appName", "host", "trace.id"},
			},
		},
		{
			"type":        "value_anomaly",
			"severity":    "low",
			"attribute":   "cpuPercent",
			"description": "Values exceeding 100% detected",
			"impact":      "Indicates potential data quality issue",
			"evidence": map[string]interface{}{
				"invalid_values": []float64{101.5, 105.2, 102.3},
				"occurrence_rate": "0.01%",
			},
		},
	}

	return map[string]interface{}{
		"eventType":         eventType,
		"anomaliesDetected": anomalies,
		"totalAnomalies":    len(anomalies),
		"sensitivity":       sensitivity,
		"dataQualityScore":  0.85,
		"recommendations": []string{
			"Investigate data gap at 14:00-14:15",
			"Verify schema change was intentional",
			"Add validation for cpuPercent values",
		},
	}, nil
}

// Helper functions

func calculateDataCompleteness(eventTypes []map[string]interface{}) float64 {
	if len(eventTypes) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, et := range eventTypes {
		score := 1.0
		
		// Check recency
		if lastSeen, ok := et["lastSeen"].(time.Time); ok {
			if time.Since(lastSeen) > 1*time.Hour {
				score *= 0.8
			}
		}

		// Check volume
		if count, ok := et["count"].(int); ok {
			if count < 1000 {
				score *= 0.9
			}
		}

		totalScore += score
	}

	return totalScore / float64(len(eventTypes))
}
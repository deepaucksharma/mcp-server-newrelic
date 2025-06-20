package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
)

// handleDiscoveryProfileCompleteness analyzes data completeness and reliability
func (s *Server) handleDiscoveryProfileCompleteness(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	eventType, ok := params["event_type"].(string)
	if !ok || eventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	criticalAttributes := []string{}
	if critAttrs, ok := params["critical_attributes"].([]interface{}); ok {
		for _, attr := range critAttrs {
			if attrStr, ok := attr.(string); ok {
				criticalAttributes = append(criticalAttributes, attrStr)
			}
		}
	}

	timeRange := "24 hours"
	if tr, ok := params["time_range"].(string); ok && tr != "" {
		timeRange = tr
	}

	checkPatterns := true
	if cp, ok := params["check_patterns"].(bool); ok {
		checkPatterns = cp
	}

	// Check mock mode
	if s.isMockMode() {
		return s.generateMockCompletenessProfile(eventType, criticalAttributes, timeRange), nil
	}

	// Cast to proper client type
	client, ok := s.nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Step 1: Get overall data volume and time distribution
	volumeQuery := fmt.Sprintf(`
		SELECT 
			count(*) as totalEvents,
			min(timestamp) as earliestEvent,
			max(timestamp) as latestEvent,
			stddev(numeric(timestamp)) as timestampVariance
		FROM %s 
		SINCE %s ago
	`, eventType, timeRange)

	volumeResult, err := client.QueryNRDB(ctx, volumeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query data volume: %w", err)
	}

	// Extract volume metrics
	var totalEvents float64
	var earliestTime, latestTime time.Time
	var timestampVariance float64

	if results, ok := volumeResult["results"].([]interface{}); ok && len(results) > 0 {
		if res, ok := results[0].(map[string]interface{}); ok {
			totalEvents = getFloat64(res, "totalEvents")
			if earliest, ok := res["earliestEvent"].(float64); ok {
				earliestTime = time.Unix(int64(earliest/1000), 0)
			}
			if latest, ok := res["latestEvent"].(float64); ok {
				latestTime = time.Unix(int64(latest/1000), 0)
			}
			timestampVariance = getFloat64(res, "timestampVariance")
		}
	}

	if totalEvents == 0 {
		return map[string]interface{}{
			"eventType": eventType,
			"status":    "NO_DATA",
			"message":   fmt.Sprintf("No data found for event type '%s' in the last %s", eventType, timeRange),
		}, nil
	}

	// Step 2: Analyze critical attributes coverage
	criticalCoverage := make(map[string]interface{})
	overallCriticalScore := 1.0

	if len(criticalAttributes) > 0 {
		for _, attr := range criticalAttributes {
			coverageQuery := fmt.Sprintf(`
				SELECT 
					count(*) as total,
					filter(count(*), WHERE %s IS NOT NULL) as nonNull,
					filter(count(*), WHERE %s IS NOT NULL AND %s != '') as nonEmpty
				FROM %s 
				SINCE %s ago
			`, attr, attr, attr, eventType, timeRange)

			coverageResult, err := client.QueryNRDB(ctx, coverageQuery)
			if err == nil && len(coverageResult["results"].([]interface{})) > 0 {
				if res, ok := coverageResult["results"].([]interface{})[0].(map[string]interface{}); ok {
					total := getFloat64(res, "total")
					nonNull := getFloat64(res, "nonNull")
					nonEmpty := getFloat64(res, "nonEmpty")

					coverage := 0.0
					if total > 0 {
						coverage = nonNull / total
					}

					criticalCoverage[attr] = map[string]interface{}{
						"coverage":      coverage,
						"nonNullCount":  int(nonNull),
						"nonEmptyCount": int(nonEmpty),
						"nullRate":      1.0 - coverage,
						"status":        getCoverageStatus(coverage),
					}

					overallCriticalScore *= coverage
				}
			}
		}
	}

	// Step 3: Check data collection patterns if requested
	patterns := map[string]interface{}{}
	if checkPatterns {
		// Check for gaps in data collection
		gapQuery := fmt.Sprintf(`
			SELECT 
				histogram(timestamp, 60000) as timeBuckets
			FROM %s 
			SINCE %s ago
		`, eventType, timeRange)

		gapResult, err := client.QueryNRDB(ctx, gapQuery)
		if err == nil {
			gaps := analyzeDataGaps(gapResult)
			patterns["gaps"] = gaps
		}

		// Check for periodicity
		periodicityQuery := fmt.Sprintf(`
			SELECT 
				count(*) as eventCount
			FROM %s 
			FACET dateOf(timestamp) as day, hourOf(timestamp) as hour
			SINCE %s ago
			LIMIT MAX
		`, eventType, timeRange)

		periodicityResult, err := client.QueryNRDB(ctx, periodicityQuery)
		if err == nil {
			periodicity := analyzeDataPeriodicity(periodicityResult)
			patterns["periodicity"] = periodicity
		}
	}

	// Step 4: Get attribute-level completeness
	attributeQuery := fmt.Sprintf(`
		SELECT keyset() 
		FROM %s 
		LIMIT 1000 
		SINCE %s ago
	`, eventType, timeRange)

	attributeResult, err := client.QueryNRDB(ctx, attributeQuery)
	var allAttributes []string
	if err == nil {
		allAttributes = extractUniqueAttributes(attributeResult)
	}

	// Calculate completeness score
	dataAge := time.Since(latestTime)
	freshnessScore := calculateFreshnessScore(dataAge)
	volumeScore := calculateVolumeScore(totalEvents, timeRange)
	consistencyScore := calculateConsistencyScore(timestampVariance)

	overallScore := (overallCriticalScore*0.4 + freshnessScore*0.3 + volumeScore*0.2 + consistencyScore*0.1)

	// Generate insights and recommendations
	insights := generateCompletenessInsights(
		eventType,
		totalEvents,
		dataAge,
		criticalCoverage,
		patterns,
		overallScore,
	)

	return map[string]interface{}{
		"eventType":          eventType,
		"timeRange":          timeRange,
		"status":             "ANALYZED",
		"totalEvents":        int(totalEvents),
		"dataTimespan":       fmt.Sprintf("%s to %s", earliestTime.Format(time.RFC3339), latestTime.Format(time.RFC3339)),
		"dataAge":            humanizeDuration(dataAge),
		"attributeCount":     len(allAttributes),
		"criticalAttributes": criticalCoverage,
		"patterns":           patterns,
		"scores": map[string]interface{}{
			"overall":       overallScore,
			"critical":      overallCriticalScore,
			"freshness":     freshnessScore,
			"volume":        volumeScore,
			"consistency":   consistencyScore,
		},
		"insights":           insights,
		"analysisTimestamp":  time.Now().UTC(),
	}, nil
}

// Helper functions for data completeness analysis

func getCoverageStatus(coverage float64) string {
	switch {
	case coverage >= 0.95:
		return "EXCELLENT"
	case coverage >= 0.8:
		return "GOOD"
	case coverage >= 0.5:
		return "FAIR"
	case coverage >= 0.2:
		return "POOR"
	default:
		return "CRITICAL"
	}
}

func analyzeDataGaps(histogramResult map[string]interface{}) map[string]interface{} {
	// Analyze histogram data to find gaps
	gaps := []map[string]interface{}{}
	totalBuckets := 0
	emptyBuckets := 0

	if results, ok := histogramResult["results"].([]interface{}); ok {
		for _, result := range results {
			if res, ok := result.(map[string]interface{}); ok {
				if buckets, ok := res["timeBuckets"].([]interface{}); ok {
					totalBuckets = len(buckets)
					
					var lastTime float64
					for i, bucket := range buckets {
						if b, ok := bucket.(map[string]interface{}); ok {
							timestamp := getFloat64(b, "timestamp")
							count := getFloat64(b, "count")
							
							if count == 0 {
								emptyBuckets++
							}
							
							// Check for significant gaps (more than 5 minutes)
							if i > 0 && lastTime > 0 {
								gap := timestamp - lastTime
								if gap > 300000 { // 5 minutes in milliseconds
									gaps = append(gaps, map[string]interface{}{
										"start":    time.Unix(int64(lastTime/1000), 0),
										"end":      time.Unix(int64(timestamp/1000), 0),
										"duration": humanizeDuration(time.Duration(gap) * time.Millisecond),
									})
								}
							}
							lastTime = timestamp
						}
					}
				}
			}
		}
	}

	gapRate := 0.0
	if totalBuckets > 0 {
		gapRate = float64(emptyBuckets) / float64(totalBuckets)
	}

	return map[string]interface{}{
		"hasGaps":       len(gaps) > 0,
		"gapCount":      len(gaps),
		"gapRate":       gapRate,
		"significantGaps": gaps,
		"analysis":      getGapAnalysis(gapRate, len(gaps)),
	}
}

func analyzeDataPeriodicity(facetResult map[string]interface{}) map[string]interface{} {
	// Analyze data distribution by hour to detect patterns
	hourlyDistribution := make(map[int]float64)
	dailyDistribution := make(map[string]float64)
	
	if results, ok := facetResult["results"].([]interface{}); ok {
		for _, result := range results {
			if res, ok := result.(map[string]interface{}); ok {
				if facets, ok := res["facet"].([]interface{}); ok && len(facets) >= 2 {
					day, _ := facets[0].(string)
					hour := int(getFloat64(map[string]interface{}{"hour": facets[1]}, "hour"))
					count := getFloat64(res, "eventCount")
					
					hourlyDistribution[hour] += count
					dailyDistribution[day] += count
				}
			}
		}
	}

	// Detect patterns
	businessHours := 0.0
	offHours := 0.0
	for hour, count := range hourlyDistribution {
		if hour >= 9 && hour <= 17 { // Business hours
			businessHours += count
		} else {
			offHours += count
		}
	}

	// Calculate variance to detect consistency
	var hourlyValues []float64
	for _, count := range hourlyDistribution {
		hourlyValues = append(hourlyValues, count)
	}
	hourlyVariance := calculateVariance(hourlyValues)

	pattern := "CONTINUOUS"
	if businessHours > offHours*2 {
		pattern = "BUSINESS_HOURS"
	} else if hourlyVariance > 0.5 {
		pattern = "IRREGULAR"
	}

	return map[string]interface{}{
		"pattern":            pattern,
		"businessHoursRatio": businessHours / (businessHours + offHours),
		"hourlyVariance":     hourlyVariance,
		"peakHours":          findPeakHours(hourlyDistribution),
	}
}

func calculateFreshnessScore(age time.Duration) float64 {
	// Score based on data age
	switch {
	case age < 5*time.Minute:
		return 1.0
	case age < 30*time.Minute:
		return 0.9
	case age < 1*time.Hour:
		return 0.8
	case age < 6*time.Hour:
		return 0.6
	case age < 24*time.Hour:
		return 0.4
	default:
		return 0.2
	}
}

func calculateVolumeScore(eventCount float64, timeRange string) float64 {
	// Normalize event count based on time range
	hours := parseTimeRangeToHours(timeRange)
	eventsPerHour := eventCount / hours
	
	// Score based on events per hour
	switch {
	case eventsPerHour > 10000:
		return 1.0
	case eventsPerHour > 1000:
		return 0.8
	case eventsPerHour > 100:
		return 0.6
	case eventsPerHour > 10:
		return 0.4
	default:
		return 0.2
	}
}

func calculateConsistencyScore(variance float64) float64 {
	// Lower variance means more consistent data
	if variance == 0 {
		return 1.0
	}
	
	// Normalize variance (this is a simplified approach)
	normalizedVariance := 1.0 / (1.0 + variance/1000000)
	return normalizedVariance
}

func generateCompletenessInsights(eventType string, totalEvents float64, dataAge time.Duration,
	criticalCoverage map[string]interface{}, patterns map[string]interface{}, overallScore float64) []string {
	
	insights := []string{}
	
	// Overall health
	if overallScore >= 0.8 {
		insights = append(insights, fmt.Sprintf("✅ %s data is healthy with %.0f%% completeness score", eventType, overallScore*100))
	} else if overallScore >= 0.5 {
		insights = append(insights, fmt.Sprintf("⚠️ %s data has moderate completeness (%.0f%%) - review recommendations", eventType, overallScore*100))
	} else {
		insights = append(insights, fmt.Sprintf("❌ %s data has poor completeness (%.0f%%) - immediate attention needed", eventType, overallScore*100))
	}
	
	// Data freshness
	if dataAge > 1*time.Hour {
		insights = append(insights, fmt.Sprintf("⚠️ Latest data is %s old - check collection pipeline", humanizeDuration(dataAge)))
	}
	
	// Critical attributes
	lowCoverageAttrs := []string{}
	for attr, coverage := range criticalCoverage {
		if coverageMap, ok := coverage.(map[string]interface{}); ok {
			if cov, ok := coverageMap["coverage"].(float64); ok && cov < 0.8 {
				lowCoverageAttrs = append(lowCoverageAttrs, fmt.Sprintf("%s (%.0f%%)", attr, cov*100))
			}
		}
	}
	if len(lowCoverageAttrs) > 0 {
		insights = append(insights, fmt.Sprintf("⚠️ Critical attributes with low coverage: %s", strings.Join(lowCoverageAttrs, ", ")))
	}
	
	// Data gaps
	if gaps, ok := patterns["gaps"].(map[string]interface{}); ok {
		if hasGaps, ok := gaps["hasGaps"].(bool); ok && hasGaps {
			if gapCount, ok := gaps["gapCount"].(int); ok {
				insights = append(insights, fmt.Sprintf("⚠️ Found %d data collection gaps - check agent connectivity", gapCount))
			}
		}
	}
	
	// Periodicity
	if periodicity, ok := patterns["periodicity"].(map[string]interface{}); ok {
		if pattern, ok := periodicity["pattern"].(string); ok {
			switch pattern {
			case "BUSINESS_HOURS":
				insights = append(insights, "📊 Data shows business hours pattern - consider 24/7 monitoring needs")
			case "IRREGULAR":
				insights = append(insights, "⚠️ Data collection is irregular - investigate collection consistency")
			}
		}
	}
	
	return insights
}

// Utility functions

func extractUniqueAttributes(keysetResult map[string]interface{}) []string {
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
	
	attributes := make([]string, 0, len(attributeMap))
	for attr := range attributeMap {
		attributes = append(attributes, attr)
	}
	sort.Strings(attributes)
	return attributes
}

func humanizeDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	return fmt.Sprintf("%.1f days", d.Hours()/24)
}

func getGapAnalysis(gapRate float64, gapCount int) string {
	if gapRate < 0.01 {
		return "Excellent - continuous data collection"
	} else if gapRate < 0.05 {
		return "Good - minimal gaps in collection"
	} else if gapRate < 0.1 {
		return "Fair - some gaps detected"
	}
	return fmt.Sprintf("Poor - significant gaps (%.0f%% gap rate)", gapRate*100)
}

func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	
	// Calculate variance
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	
	return variance / float64(len(values))
}

func findPeakHours(hourlyDist map[int]float64) []int {
	type hourCount struct {
		hour  int
		count float64
	}
	
	hours := make([]hourCount, 0, len(hourlyDist))
	for h, c := range hourlyDist {
		hours = append(hours, hourCount{h, c})
	}
	
	// Sort by count descending
	sort.Slice(hours, func(i, j int) bool {
		return hours[i].count > hours[j].count
	})
	
	// Return top 3 hours
	peakHours := []int{}
	for i := 0; i < 3 && i < len(hours); i++ {
		peakHours = append(peakHours, hours[i].hour)
	}
	
	return peakHours
}

func parseTimeRangeToHours(timeRange string) float64 {
	// Simple parser for time ranges like "24 hours", "7 days"
	parts := strings.Fields(timeRange)
	if len(parts) < 2 {
		return 24 // Default to 24 hours
	}
	
	value := 1.0
	fmt.Sscanf(parts[0], "%f", &value)
	
	unit := strings.ToLower(parts[1])
	switch {
	case strings.HasPrefix(unit, "hour"):
		return value
	case strings.HasPrefix(unit, "day"):
		return value * 24
	case strings.HasPrefix(unit, "week"):
		return value * 24 * 7
	case strings.HasPrefix(unit, "minute"):
		return value / 60
	default:
		return 24
	}
}

// Mock data generator for testing
func (s *Server) generateMockCompletenessProfile(eventType string, criticalAttributes []string, timeRange string) map[string]interface{} {
	// Generate realistic mock data for testing
	criticalCoverage := make(map[string]interface{})
	for _, attr := range criticalAttributes {
		coverage := 0.85 + (0.15 * (float64(len(attr)) / 20.0)) // Vary coverage based on attribute name length
		criticalCoverage[attr] = map[string]interface{}{
			"coverage":      coverage,
			"nonNullCount":  int(10000 * coverage),
			"nonEmptyCount": int(9500 * coverage),
			"nullRate":      1.0 - coverage,
			"status":        getCoverageStatus(coverage),
		}
	}

	return map[string]interface{}{
		"eventType":          eventType,
		"timeRange":          timeRange,
		"status":             "ANALYZED",
		"totalEvents":        128459,
		"dataTimespan":       "2024-01-20T10:00:00Z to 2024-01-21T09:55:00Z",
		"dataAge":            "5 minutes",
		"attributeCount":     47,
		"criticalAttributes": criticalCoverage,
		"patterns": map[string]interface{}{
			"gaps": map[string]interface{}{
				"hasGaps":  false,
				"gapCount": 0,
				"gapRate":  0.0,
				"analysis": "Excellent - continuous data collection",
			},
			"periodicity": map[string]interface{}{
				"pattern":            "CONTINUOUS",
				"businessHoursRatio": 0.42,
				"hourlyVariance":     0.12,
				"peakHours":          []int{14, 15, 10},
			},
		},
		"scores": map[string]interface{}{
			"overall":     0.89,
			"critical":    0.92,
			"freshness":   1.0,
			"volume":      0.85,
			"consistency": 0.78,
		},
		"insights": []string{
			"✅ " + eventType + " data is healthy with 89% completeness score",
			"✅ All critical attributes have good coverage (>80%)",
			"📊 Continuous data collection pattern detected",
			"💡 Consider monitoring attribute cardinality for performance optimization",
		},
		"analysisTimestamp": time.Now().UTC(),
	}
}
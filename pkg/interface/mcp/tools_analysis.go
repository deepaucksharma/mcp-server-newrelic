package mcp

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// RegisterAnalysisTools registers all analysis tools
func (s *Server) RegisterAnalysisTools() error {
	tools := []Tool{
		// Baseline calculation
		{
			Name:        "analysis.calculate_baseline",
			Description: "Calculate statistical baseline for a metric over time",
			Parameters: ToolParameters{
				Type:     "object",
				Required: []string{"metric", "event_type"},
				Properties: map[string]Property{
					"metric": {
						Type:        "string",
						Description: "Metric to analyze (e.g., 'duration', 'cpuPercent')",
					},
					"event_type": {
						Type:        "string",
						Description: "Event type containing the metric",
					},
					"time_range": {
						Type:        "string",
						Description: "Time range for baseline calculation",
						Default:     "7 days",
					},
					"percentiles": {
						Type:        "array",
						Description: "Percentiles to calculate",
						Default:     []int{50, 90, 95, 99},
						Items: &Property{
							Type: "number",
						},
					},
					"group_by": {
						Type:        "string",
						Description: "Optional attribute to group by (e.g., 'appName')",
					},
				},
			},
			Handler: s.handleAnalysisCalculateBaseline,
		},

		// Anomaly detection
		{
			Name:        "analysis.detect_anomalies",
			Description: "Detect anomalies in time series data using statistical methods",
			Parameters: ToolParameters{
				Type:     "object",
				Required: []string{"metric", "event_type"},
				Properties: map[string]Property{
					"metric": {
						Type:        "string",
						Description: "Metric to analyze for anomalies",
					},
					"event_type": {
						Type:        "string",
						Description: "Event type containing the metric",
					},
					"time_range": {
						Type:        "string",
						Description: "Time range to analyze",
						Default:     "24 hours",
					},
					"sensitivity": {
						Type:        "number",
						Description: "Anomaly detection sensitivity (1-5, higher = more sensitive)",
						Default:     3,
					},
					"method": {
						Type:        "string",
						Description: "Detection method: 'zscore', 'iqr', 'isolation_forest'",
						Default:     "zscore",
						Enum:        []string{"zscore", "iqr", "isolation_forest"},
					},
				},
			},
			Handler: s.handleAnalysisDetectAnomalies,
		},

		// Correlation analysis
		{
			Name:        "analysis.find_correlations",
			Description: "Find correlations between different metrics or attributes",
			Parameters: ToolParameters{
				Type:     "object",
				Required: []string{"primary_metric", "event_type"},
				Properties: map[string]Property{
					"primary_metric": {
						Type:        "string",
						Description: "Primary metric to correlate",
					},
					"secondary_metrics": {
						Type:        "array",
						Description: "Metrics to correlate with primary (null = auto-discover)",
						Items: &Property{
							Type: "string",
						},
					},
					"event_type": {
						Type:        "string",
						Description: "Event type containing the metrics",
					},
					"time_range": {
						Type:        "string",
						Description: "Time range for correlation analysis",
						Default:     "24 hours",
					},
					"min_correlation": {
						Type:        "number",
						Description: "Minimum correlation coefficient to report",
						Default:     0.7,
					},
				},
			},
			Handler: s.handleAnalysisFindCorrelations,
		},

		// Trend analysis
		{
			Name:        "analysis.analyze_trend",
			Description: "Analyze trends and patterns in metric data over time",
			Parameters: ToolParameters{
				Type:     "object",
				Required: []string{"metric", "event_type"},
				Properties: map[string]Property{
					"metric": {
						Type:        "string",
						Description: "Metric to analyze",
					},
					"event_type": {
						Type:        "string",
						Description: "Event type containing the metric",
					},
					"time_range": {
						Type:        "string",
						Description: "Time range for trend analysis",
						Default:     "30 days",
					},
					"granularity": {
						Type:        "string",
						Description: "Time bucket size: 'minute', 'hour', 'day'",
						Default:     "hour",
						Enum:        []string{"minute", "hour", "day"},
					},
					"include_forecast": {
						Type:        "boolean",
						Description: "Include trend forecast",
						Default:     true,
					},
				},
			},
			Handler: s.handleAnalysisAnalyzeTrend,
		},

		// Distribution analysis
		{
			Name:        "analysis.analyze_distribution",
			Description: "Analyze the distribution characteristics of a metric",
			Parameters: ToolParameters{
				Type:     "object",
				Required: []string{"metric", "event_type"},
				Properties: map[string]Property{
					"metric": {
						Type:        "string",
						Description: "Metric to analyze",
					},
					"event_type": {
						Type:        "string",
						Description: "Event type containing the metric",
					},
					"time_range": {
						Type:        "string",
						Description: "Time range to analyze",
						Default:     "24 hours",
					},
					"buckets": {
						Type:        "integer",
						Description: "Number of histogram buckets",
						Default:     20,
					},
				},
			},
			Handler: s.handleAnalysisAnalyzeDistribution,
		},

		// Segment comparison
		{
			Name:        "analysis.compare_segments",
			Description: "Compare metrics across different segments (e.g., by app, host, region)",
			Parameters: ToolParameters{
				Type:     "object",
				Required: []string{"metric", "event_type", "segment_by"},
				Properties: map[string]Property{
					"metric": {
						Type:        "string",
						Description: "Metric to compare",
					},
					"event_type": {
						Type:        "string",
						Description: "Event type containing the metric",
					},
					"segment_by": {
						Type:        "string",
						Description: "Attribute to segment by (e.g., 'appName', 'host')",
					},
					"time_range": {
						Type:        "string",
						Description: "Time range for comparison",
						Default:     "24 hours",
					},
					"comparison_type": {
						Type:        "string",
						Description: "Type of comparison: 'absolute', 'relative', 'ranked'",
						Default:     "relative",
						Enum:        []string{"absolute", "relative", "ranked"},
					},
				},
			},
			Handler: s.handleAnalysisCompareSegments,
		},
	}

	// Register all tools
	for _, tool := range tools {
		if err := s.tools.Register(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name, err)
		}
	}

	return nil
}

// Implementation handlers

func (s *Server) handleAnalysisCalculateBaseline(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	metric, _ := params["metric"].(string)
	eventType, _ := params["event_type"].(string)
	timeRange, _ := params["time_range"].(string)
	if timeRange == "" {
		timeRange = "7 days"
	}
	
	percentiles, _ := params["percentiles"].([]interface{})
	if len(percentiles) == 0 {
		percentiles = []interface{}{50.0, 90.0, 95.0, 99.0}
	}
	
	groupBy, _ := params["group_by"].(string)

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("analysis.calculate_baseline", params), nil
	}

	// Build NRQL query for baseline calculation
	percentilesStr := ""
	for i, p := range percentiles {
		if i > 0 {
			percentilesStr += ", "
		}
		percentilesStr += fmt.Sprintf("percentile(%s, %v) as p%v", metric, p, p)
	}

	query := fmt.Sprintf(`
		SELECT 
			average(%s) as avg,
			stddev(%s) as stddev,
			min(%s) as min,
			max(%s) as max,
			count(*) as count,
			%s
		FROM %s
		%s
		SINCE %s
	`, metric, metric, metric, metric, percentilesStr, eventType,
		func() string {
			if groupBy != "" {
				return fmt.Sprintf("FACET %s", groupBy)
			}
			return ""
		}(),
		timeRange)

	// Execute query
	result, err := s.executeNRQL(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate baseline: %w", err)
	}

	// Process results
	baseline := processBaselineResults(result, metric, percentiles)
	
	// Add recommendations based on baseline
	baseline["recommendations"] = generateBaselineRecommendations(baseline)

	return baseline, nil
}

func (s *Server) handleAnalysisDetectAnomalies(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	metric, _ := params["metric"].(string)
	eventType, _ := params["event_type"].(string)
	timeRange, _ := params["time_range"].(string)
	if timeRange == "" {
		timeRange = "24 hours"
	}
	
	sensitivity, _ := params["sensitivity"].(float64)
	if sensitivity == 0 {
		sensitivity = 3
	}
	
	method, _ := params["method"].(string)
	if method == "" {
		method = "zscore"
	}

	// Check mock mode
	if s.nrClient == nil {
		return s.mockAnalysisAnomalies(metric, eventType, timeRange, sensitivity, method), nil
	}

	// Get time series data
	query := fmt.Sprintf(`
		SELECT 
			average(%s) as value
		FROM %s
		TIMESERIES 5 minutes
		SINCE %s
	`, metric, eventType, timeRange)

	result, err := s.executeNRQL(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get time series: %w", err)
	}

	// Apply anomaly detection
	timeSeries := extractTimeSeries(result)
	anomalies := detectAnomalies(timeSeries, method, sensitivity)

	return map[string]interface{}{
		"metric":             metric,
		"eventType":          eventType,
		"timeRange":          timeRange,
		"method":             method,
		"sensitivity":        sensitivity,
		"anomaliesDetected":  len(anomalies),
		"anomalies":          anomalies,
		"normalRanges":       calculateNormalRanges(timeSeries, anomalies),
		"recommendations":    generateAnomalyRecommendations(anomalies),
	}, nil
}

func (s *Server) handleAnalysisFindCorrelations(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	primaryMetric, _ := params["primary_metric"].(string)
	eventType, _ := params["event_type"].(string)
	timeRange, _ := params["time_range"].(string)
	if timeRange == "" {
		timeRange = "24 hours"
	}
	
	minCorrelation, _ := params["min_correlation"].(float64)
	if minCorrelation == 0 {
		minCorrelation = 0.7
	}

	// Check mock mode
	if s.nrClient == nil {
		return s.mockAnalysisCorrelations(primaryMetric, eventType, timeRange, minCorrelation), nil
	}

	// If secondary metrics not specified, discover numeric attributes
	secondaryMetrics, _ := params["secondary_metrics"].([]interface{})
	if len(secondaryMetrics) == 0 {
		// Auto-discover numeric attributes
		secondaryMetrics = discoverNumericAttributes(ctx, s, eventType)
	}

	// Get time series data for all metrics
	correlations := []map[string]interface{}{}
	
	for _, secondary := range secondaryMetrics {
		secondaryStr, _ := secondary.(string)
		if secondaryStr == primaryMetric {
			continue
		}

		correlation := calculateCorrelation(ctx, s, eventType, primaryMetric, secondaryStr, timeRange)
		if math.Abs(correlation["coefficient"].(float64)) >= minCorrelation {
			correlations = append(correlations, correlation)
		}
	}

	// Sort by correlation strength
	sort.Slice(correlations, func(i, j int) bool {
		return math.Abs(correlations[i]["coefficient"].(float64)) > math.Abs(correlations[j]["coefficient"].(float64))
	})

	return map[string]interface{}{
		"primaryMetric":      primaryMetric,
		"eventType":          eventType,
		"timeRange":          timeRange,
		"correlations":       correlations,
		"strongCorrelations": len(correlations),
		"insights":           generateCorrelationInsights(correlations),
	}, nil
}

// Mock implementations for testing

func (s *Server) mockAnalysisBaseline(metric, eventType, timeRange string, percentiles []interface{}, groupBy string) interface{} {
	return map[string]interface{}{
		"metric":    metric,
		"eventType": eventType,
		"timeRange": timeRange,
		"baseline": map[string]interface{}{
			"avg":    125.5,
			"stddev": 45.2,
			"min":    10.0,
			"max":    500.0,
			"count":  150000,
			"p50":    110.0,
			"p90":    180.0,
			"p95":    220.0,
			"p99":    350.0,
		},
		"recommendations": []string{
			fmt.Sprintf("Normal range for %s is 80-170 (p10-p90)", metric),
			"Consider alerting when value exceeds 220 (p95)",
			"High variability detected (stddev/avg = 0.36)",
		},
	}
}

func (s *Server) mockAnalysisAnomalies(metric, eventType, timeRange string, sensitivity float64, method string) interface{} {
	return map[string]interface{}{
		"metric":            metric,
		"eventType":         eventType,
		"timeRange":         timeRange,
		"method":            method,
		"sensitivity":       sensitivity,
		"anomaliesDetected": 3,
		"anomalies": []map[string]interface{}{
			{
				"timestamp":  time.Now().Add(-2 * time.Hour).Unix(),
				"value":      450.5,
				"zscore":     3.2,
				"severity":   "high",
				"type":       "spike",
				"context":    "Value is 3.2 standard deviations above mean",
			},
			{
				"timestamp":  time.Now().Add(-8 * time.Hour).Unix(),
				"value":      15.0,
				"zscore":     -2.8,
				"severity":   "medium",
				"type":       "dip",
				"context":    "Unusual drop in metric value",
			},
		},
		"normalRanges": map[string]interface{}{
			"mean":   125.5,
			"stddev": 45.2,
			"lower":  35.1,  // mean - 2*stddev
			"upper":  215.9, // mean + 2*stddev
		},
		"recommendations": []string{
			"Investigate high severity spike 2 hours ago",
			"Check for deployment or configuration changes around anomaly times",
			"Consider setting alert threshold at 216 (mean + 2σ)",
		},
	}
}

func (s *Server) mockAnalysisCorrelations(primaryMetric, eventType, timeRange string, minCorrelation float64) interface{} {
	return map[string]interface{}{
		"primaryMetric": primaryMetric,
		"eventType":     eventType,
		"timeRange":     timeRange,
		"correlations": []map[string]interface{}{
			{
				"metric":      "cpuPercent",
				"coefficient": 0.89,
				"pValue":      0.001,
				"strength":    "strong positive",
				"insight":     fmt.Sprintf("cpuPercent increases with %s", primaryMetric),
			},
			{
				"metric":      "memoryUsage",
				"coefficient": 0.75,
				"pValue":      0.003,
				"strength":    "moderate positive",
				"insight":     "Moderate correlation suggests resource contention",
			},
			{
				"metric":      "errorRate",
				"coefficient": -0.72,
				"pValue":      0.005,
				"strength":    "moderate negative",
				"insight":     fmt.Sprintf("errorRate decreases as %s increases", primaryMetric),
			},
		},
		"strongCorrelations": 3,
		"insights": []string{
			fmt.Sprintf("%s is strongly correlated with resource usage", primaryMetric),
			"Consider monitoring cpuPercent alongside primary metric",
			"Inverse correlation with errorRate suggests performance trade-off",
		},
	}
}

// Helper functions

func (s *Server) executeNRQL(ctx context.Context, query string) (map[string]interface{}, error) {
	// This would call the actual New Relic client
	// For now, returning mock data
	return map[string]interface{}{
		"results": []interface{}{
			map[string]interface{}{
				"average": 125.5,
				"count":   1000,
			},
		},
	}, nil
}

func extractTimeSeries(result map[string]interface{}) []map[string]interface{} {
	// Extract time series from NRQL result
	return []map[string]interface{}{}
}

func detectAnomalies(timeSeries []map[string]interface{}, method string, sensitivity float64) []map[string]interface{} {
	// Implement anomaly detection algorithms
	return []map[string]interface{}{}
}

func calculateNormalRanges(timeSeries []map[string]interface{}, anomalies []map[string]interface{}) map[string]interface{} {
	// Calculate normal ranges excluding anomalies
	return map[string]interface{}{}
}

func generateAnomalyRecommendations(anomalies []map[string]interface{}) []string {
	// Generate recommendations based on anomalies found
	return []string{}
}

func processBaselineResults(result map[string]interface{}, metric string, percentiles []interface{}) map[string]interface{} {
	// Process NRQL results into baseline format
	return map[string]interface{}{}
}

func generateBaselineRecommendations(baseline map[string]interface{}) []string {
	// Generate recommendations based on baseline
	return []string{}
}

func discoverNumericAttributes(ctx context.Context, s *Server, eventType string) []interface{} {
	// Auto-discover numeric attributes
	return []interface{}{}
}

func calculateCorrelation(ctx context.Context, s *Server, eventType, metric1, metric2, timeRange string) map[string]interface{} {
	// Calculate correlation between two metrics
	return map[string]interface{}{
		"metric":      metric2,
		"coefficient": 0.0,
	}
}

func generateCorrelationInsights(correlations []map[string]interface{}) []string {
	// Generate insights from correlations
	return []string{}
}
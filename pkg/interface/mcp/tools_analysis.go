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
	result, err := s.executeNRQL(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate baseline: %w", err)
	}

	// Process results
	baseline := processBaselineResults(result, metric, percentiles, groupBy)

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
		sensitivity = 0.7 // 0-1 scale
	}
	
	method, _ := params["method"].(string)
	if method == "" {
		method = "multi" // Use all methods
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("analysis.detect_anomalies", params), nil
	}

	// Use the new anomaly detector
	detector := &AnomalyDetector{server: s}
	result, err := detector.DetectAnomalies(ctx, metric, eventType, timeRange, sensitivity)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	response := map[string]interface{}{
		"metric":             result.Metric,
		"event_type":         result.EventType,
		"time_range":         result.TimeRange,
		"method":             method,
		"sensitivity":        sensitivity,
		"anomalies_detected": len(result.Anomalies),
		"anomalies":          result.Anomalies,
		"statistics":         result.Statistics,
		"summary":            result.Summary,
		"recommendations":    generateAnomalyRecommendations(result.Anomalies),
	}

	return response, nil
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
		minCorrelation = 0.5 // Lower default for better discovery
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("analysis.find_correlations", params), nil
	}

	// Get candidate metrics
	candidateMetrics := []string{}
	if secondary, ok := params["secondary_metrics"].([]interface{}); ok && len(secondary) > 0 {
		for _, m := range secondary {
			if metric, ok := m.(string); ok {
				candidateMetrics = append(candidateMetrics, metric)
			}
		}
	} else {
		// Auto-discover numeric attributes if not provided
		// For now, use common metrics
		candidateMetrics = []string{"duration", "errorCount", "throughput", "cpuPercent", "memoryPercent"}
	}

	// Use the new correlation analyzer
	analyzer := &CorrelationAnalyzer{server: s}
	result, err := analyzer.FindCorrelations(ctx, primaryMetric, candidateMetrics, eventType, timeRange)
	if err != nil {
		return nil, err
	}

	// Filter by minimum correlation
	filteredCorrelations := []MetricCorrelation{}
	for _, corr := range result.Correlations {
		if math.Abs(corr.Coefficient) >= minCorrelation {
			filteredCorrelations = append(filteredCorrelations, corr)
		}
	}

	// Convert to response format
	correlationMaps := []map[string]interface{}{}
	for _, corr := range filteredCorrelations {
		correlationMaps = append(correlationMaps, map[string]interface{}{
			"metric":       corr.Metric,
			"coefficient":  corr.Coefficient,
			"lag":          corr.Lag,
			"lagged_coeff": corr.LaggedCoeff,
			"data_points":  corr.DataPoints,
			"relationship": corr.Relationship,
		})
	}

	return map[string]interface{}{
		"primary_metric":      primaryMetric,
		"event_type":          eventType,
		"time_range":          timeRange,
		"correlations":        correlationMaps,
		"strong_correlations": len(filteredCorrelations),
		"summary":             result.Summary,
		"insights":            generateCorrelationInsights(correlationMaps),
	}, nil
}

func (s *Server) handleAnalysisAnalyzeTrend(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	metric, _ := params["metric"].(string)
	eventType, _ := params["event_type"].(string)
	timeRange, _ := params["time_range"].(string)
	if timeRange == "" {
		timeRange = "30 days"
	}
	granularity, _ := params["granularity"].(string)
	if granularity == "" {
		granularity = "hour"
	}
	includeForecast := true
	if val, ok := params["include_forecast"].(bool); ok {
		includeForecast = val
	}

	// Check mock mode
	if s.isMockMode() {
		return s.mockAnalysisTrend(metric, eventType, timeRange, granularity, includeForecast), nil
	}

	// Use the new trend analyzer
	analyzer := &TrendAnalyzer{server: s}
	result, err := analyzer.AnalyzeTrends(ctx, metric, eventType, timeRange)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	forecastValues := []map[string]interface{}{}
	if includeForecast && len(result.Forecast.Values) > 0 {
		for i, val := range result.Forecast.Values {
			forecastValues = append(forecastValues, map[string]interface{}{
				"timestamp": val.Timestamp.Unix(),
				"value":     val.Value,
				"lower":     result.Forecast.Confidence[i].Lower,
				"upper":     result.Forecast.Confidence[i].Upper,
			})
		}
	}

	changePointMaps := []map[string]interface{}{}
	for _, cp := range result.ChangePoints {
		changePointMaps = append(changePointMaps, map[string]interface{}{
			"timestamp":  cp.Timestamp.Unix(),
			"old_value":  cp.OldValue,
			"new_value":  cp.NewValue,
			"confidence": cp.Confidence,
			"type":       cp.Type,
		})
	}

	return map[string]interface{}{
		"metric":      metric,
		"event_type":  eventType,
		"time_range":  timeRange,
		"granularity": granularity,
		"trend": map[string]interface{}{
			"direction":      result.LinearTrend.Direction,
			"strength":       result.LinearTrend.Strength,
			"slope":          result.LinearTrend.Slope,
			"r_squared":      result.LinearTrend.RSquared,
			"percent_change": result.LinearTrend.PercentChange,
		},
		"seasonality": map[string]interface{}{
			"detected": result.Seasonality.Detected,
			"period":   result.Seasonality.Period,
			"strength": result.Seasonality.Strength,
			"pattern":  result.Seasonality.Pattern,
		},
		"change_points": changePointMaps,
		"forecast": map[string]interface{}{
			"enabled": includeForecast,
			"values":  forecastValues,
		},
		"summary": result.Summary,
		"insights": []string{
			fmt.Sprintf("Metric shows %s trend with %s fit", result.LinearTrend.Direction, result.LinearTrend.Strength),
			fmt.Sprintf("%.1f%% change over the time period", result.LinearTrend.PercentChange),
		},
	}, nil
}

func (s *Server) handleAnalysisAnalyzeDistribution(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	metric, _ := params["metric"].(string)
	eventType, _ := params["event_type"].(string)
	timeRange, _ := params["time_range"].(string)
	if timeRange == "" {
		timeRange = "24 hours"
	}
	buckets := 20
	if val, ok := params["buckets"].(float64); ok {
		buckets = int(val)
	}

	// Check mock mode
	if s.isMockMode() {
		return s.mockAnalysisDistribution(metric, eventType, timeRange, buckets), nil
	}

	// Build NRQL query for all data points
	query := fmt.Sprintf(`
		SELECT %s as value
		FROM %s 
		SINCE %s
		LIMIT MAX
	`, metric, eventType, timeRange)

	// Execute query
	result, err := s.executeNRQL(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch distribution data: %w", err)
	}

	// Extract values from result
	values := []float64{}
	if results, ok := result["results"].([]interface{}); ok {
		for _, r := range results {
			if row, ok := r.(map[string]interface{}); ok {
				if val, ok := row["value"].(float64); ok {
					values = append(values, val)
				}
			}
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no data found for distribution analysis")
	}

	// Calculate distribution statistics
	sort.Float64s(values)
	stats := calculateDistributionStats(values)
	histogram := createHistogram(values, buckets)
	distributionType := detectDistributionType(values, stats)

	return map[string]interface{}{
		"metric":     metric,
		"event_type": eventType,
		"time_range": timeRange,
		"statistics": map[string]interface{}{
			"mean":       stats.Mean,
			"median":     stats.Median,
			"mode":       stats.Mode,
			"stddev":     stats.StdDev,
			"variance":   stats.Variance,
			"skewness":   stats.Skewness,
			"kurtosis":   stats.Kurtosis,
			"min":        stats.Min,
			"max":        stats.Max,
			"count":      len(values),
			"percentiles": map[string]float64{
				"p25": percentile(values, 25),
				"p50": stats.Median,
				"p75": percentile(values, 75),
				"p90": percentile(values, 90),
				"p95": percentile(values, 95),
				"p99": percentile(values, 99),
			},
		},
		"histogram":        histogram,
		"distribution_type": distributionType,
		"insights":         generateDistributionInsights(stats, distributionType, histogram),
	}, nil
}

func (s *Server) handleAnalysisCompareSegments(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	metric, _ := params["metric"].(string)
	eventType, _ := params["event_type"].(string)
	segmentBy, _ := params["segment_by"].(string)
	timeRange, _ := params["time_range"].(string)
	if timeRange == "" {
		timeRange = "24 hours"
	}
	comparisonType, _ := params["comparison_type"].(string)
	if comparisonType == "" {
		comparisonType = "relative"
	}

	// Check mock mode
	if s.isMockMode() {
		return s.mockAnalysisSegments(metric, eventType, segmentBy, timeRange, comparisonType), nil
	}

	// Build NRQL query for segment comparison
	query := fmt.Sprintf(`
		SELECT 
			average(%s) as avg,
			count(*) as count,
			stddev(%s) as stddev,
			min(%s) as min,
			max(%s) as max,
			percentile(%s, 50, 90, 95) as percentiles
		FROM %s 
		FACET %s
		SINCE %s
		LIMIT 50
	`, metric, metric, metric, metric, metric, eventType, segmentBy, timeRange)

	// Execute query
	result, err := s.executeNRQL(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to compare segments: %w", err)
	}

	// Process and analyze segments
	segments := processSegmentResults(result, metric, segmentBy, comparisonType)
	analysis := analyzeSegmentDifferences(segments, comparisonType)

	return map[string]interface{}{
		"metric":          metric,
		"event_type":      eventType,
		"segment_by":      segmentBy,
		"time_range":      timeRange,
		"comparison_type": comparisonType,
		"segments":        segments,
		"analysis":        analysis,
		"insights":        generateSegmentInsights(segments, analysis, comparisonType),
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

func (s *Server) executeNRQL(ctx context.Context, query string, accountID interface{}) (map[string]interface{}, error) {
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

// These helper functions are now implemented in analysis_helpers.go and analysis_algorithms.go

func analyzeTrendData(result map[string]interface{}) map[string]interface{} {
	// Analyze trend patterns from NRQL results
	return map[string]interface{}{
		"direction": "increasing",
		"slope": 0.05,
		"rsquared": 0.85,
	}
}

func generateTrendInsights(trendData map[string]interface{}) []string {
	// Generate insights from trend analysis
	return []string{
		"Metric shows steady upward trend",
		"Growth rate: 5% per day",
		"Forecast: likely to exceed threshold in 7 days",
	}
}

func analyzeDistribution(result map[string]interface{}) map[string]interface{} {
	// Analyze distribution characteristics
	return map[string]interface{}{
		"type": "normal",
		"skewness": 0.2,
		"kurtosis": 3.1,
		"outliers": 5,
	}
}

// generateDistributionInsights is now implemented in analysis_helpers.go

func analyzeSegments(result map[string]interface{}, comparisonType string) []map[string]interface{} {
	// Analyze segments from faceted query
	return []map[string]interface{}{
		{
			"segment": "app1",
			"avg": 125.5,
			"count": 1000,
			"relative": 1.0,
		},
		{
			"segment": "app2", 
			"avg": 150.2,
			"count": 800,
			"relative": 1.2,
		},
	}
}

// generateSegmentInsights is now implemented in analysis_helpers.go

func (s *Server) mockAnalysisTrend(metric, eventType, timeRange, granularity string, includeForecast bool) interface{} {
	return map[string]interface{}{
		"metric": metric,
		"eventType": eventType,
		"timeRange": timeRange,
		"granularity": granularity,
		"trend": map[string]interface{}{
			"direction": "increasing",
			"slope": 0.05,
			"rsquared": 0.85,
			"changePercent": 15.5,
			"dataPoints": 720,
		},
		"forecast": map[string]interface{}{
			"enabled": includeForecast,
			"nextPeriod": 145.2,
			"confidence": 0.90,
			"upperBound": 165.5,
			"lowerBound": 125.0,
		},
		"insights": []string{
			"Metric shows steady 5% daily growth",
			"Strong trend fit (R² = 0.85)",
			"At current rate, will reach 200 in 14 days",
		},
	}
}

func (s *Server) mockAnalysisDistribution(metric, eventType, timeRange string, buckets int) interface{} {
	return map[string]interface{}{
		"metric": metric,
		"eventType": eventType,
		"timeRange": timeRange,
		"distribution": map[string]interface{}{
			"type": "normal",
			"mean": 125.5,
			"median": 123.0,
			"mode": 120.0,
			"stddev": 45.2,
			"skewness": 0.2,
			"kurtosis": 3.1,
			"percentiles": map[string]float64{
				"p50": 123.0,
				"p90": 180.0,
				"p95": 220.0,
				"p99": 350.0,
			},
			"histogram": []map[string]interface{}{
				{"bucket": "0-50", "count": 100},
				{"bucket": "50-100", "count": 300},
				{"bucket": "100-150", "count": 400},
				{"bucket": "150-200", "count": 150},
				{"bucket": "200+", "count": 50},
			},
		},
		"insights": []string{
			"Distribution is approximately normal",
			"95% of values fall between 35-216",
			"Consider alerting above p95 threshold (220)",
		},
	}
}

func (s *Server) mockAnalysisSegments(metric, eventType, segmentBy, timeRange, comparisonType string) interface{} {
	return map[string]interface{}{
		"metric": metric,
		"eventType": eventType,
		"segmentBy": segmentBy,
		"timeRange": timeRange,
		"comparisonType": comparisonType,
		"segments": []map[string]interface{}{
			{
				"name": "app1",
				"avg": 125.5,
				"count": 5000,
				"stddev": 45.2,
				"relative": 1.0,
				"rank": 2,
				"percentOfTotal": 40,
			},
			{
				"name": "app2",
				"avg": 150.2,
				"count": 3000,
				"stddev": 52.1,
				"relative": 1.2,
				"rank": 1,
				"percentOfTotal": 30,
			},
			{
				"name": "app3",
				"avg": 98.7,
				"count": 3000,
				"stddev": 38.5,
				"relative": 0.79,
				"rank": 3,
				"percentOfTotal": 30,
			},
		},
		"insights": []string{
			"app2 shows 20% higher values than baseline",
			"app3 performs 21% better (lower values)",
			"High variability in app2 (stddev=52.1)",
		},
	}
}

// processBaselineResults and generateBaselineRecommendations are now implemented in analysis_helpers.go

func discoverNumericAttributes(ctx context.Context, s *Server, eventType string) []interface{} {
	// Auto-discover numeric attributes
	return []interface{}{}
}

// calculateCorrelation is now implemented in analysis_algorithms.go


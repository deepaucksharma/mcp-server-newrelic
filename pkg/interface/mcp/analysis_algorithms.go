package mcp

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// AnomalyDetector implements statistical anomaly detection
type AnomalyDetector struct {
	server *Server
}

// DetectAnomalies performs anomaly detection on time series data
func (ad *AnomalyDetector) DetectAnomalies(ctx context.Context, metric string, eventType string, timeRange string, sensitivity float64) (*AnomalyResult, error) {
	// Step 1: Fetch historical data
	query := fmt.Sprintf(`
		SELECT 
			average(%s) as value,
			timestamp
		FROM %s
		TIMESERIES 5 minutes
		SINCE %s
	`, metric, eventType, timeRange)

	result, err := ad.server.executeNRQL(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch time series data: %w", err)
	}

	// Extract time series data
	timeSeries, err := extractTimeSeries(result)
	if err != nil {
		return nil, err
	}

	// Step 2: Calculate statistics
	stats := calculateStatistics(timeSeries)

	// Step 3: Detect anomalies using multiple methods
	anomalies := []Anomaly{}

	// Method 1: Z-Score (Standard Deviation)
	zScoreAnomalies := detectZScoreAnomalies(timeSeries, stats, sensitivity)
	anomalies = append(anomalies, zScoreAnomalies...)

	// Method 2: Interquartile Range (IQR)
	iqrAnomalies := detectIQRAnomalies(timeSeries, sensitivity)
	anomalies = append(anomalies, iqrAnomalies...)

	// Method 3: Moving Average
	maAnomalies := detectMovingAverageAnomalies(timeSeries, sensitivity)
	anomalies = append(anomalies, maAnomalies...)

	// Deduplicate and score anomalies
	finalAnomalies := deduplicateAndScore(anomalies)

	return &AnomalyResult{
		Metric:     metric,
		EventType:  eventType,
		TimeRange:  timeRange,
		Statistics: stats,
		Anomalies:  finalAnomalies,
		Summary:    generateAnomalySummary(finalAnomalies, stats),
	}, nil
}

// CorrelationAnalyzer implements correlation analysis
type CorrelationAnalyzer struct {
	server *Server
}

// FindCorrelations analyzes correlations between metrics
func (ca *CorrelationAnalyzer) FindCorrelations(ctx context.Context, primaryMetric string, candidateMetrics []string, eventType string, timeRange string) (*CorrelationResult, error) {
	// Fetch primary metric data
	primaryQuery := fmt.Sprintf(`
		SELECT 
			average(%s) as value,
			timestamp
		FROM %s
		TIMESERIES 5 minutes
		SINCE %s
	`, primaryMetric, eventType, timeRange)

	primaryResult, err := ca.server.executeNRQL(ctx, primaryQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch primary metric: %w", err)
	}

	primarySeries, err := extractTimeSeries(primaryResult)
	if err != nil {
		return nil, err
	}

	correlations := []MetricCorrelation{}

	// Analyze each candidate metric
	for _, candidateMetric := range candidateMetrics {
		// Fetch candidate metric data
		candidateQuery := fmt.Sprintf(`
			SELECT 
				average(%s) as value,
				timestamp
			FROM %s
			TIMESERIES 5 minutes
			SINCE %s
		`, candidateMetric, eventType, timeRange)

		candidateResult, err := ca.server.executeNRQL(ctx, candidateQuery, nil)
		if err != nil {
			continue // Skip on error
		}

		candidateSeries, err := extractTimeSeries(candidateResult)
		if err != nil {
			continue
		}

		// Calculate correlation
		correlation := calculateCorrelation(primarySeries, candidateSeries)
		if correlation != nil {
			correlations = append(correlations, *correlation)
		}
	}

	// Sort by correlation strength
	sort.Slice(correlations, func(i, j int) bool {
		return math.Abs(correlations[i].Coefficient) > math.Abs(correlations[j].Coefficient)
	})

	return &CorrelationResult{
		PrimaryMetric: primaryMetric,
		EventType:     eventType,
		TimeRange:     timeRange,
		Correlations:  correlations,
		Summary:       generateCorrelationSummary(correlations),
	}, nil
}

// TrendAnalyzer implements trend analysis
type TrendAnalyzer struct {
	server *Server
}

// AnalyzeTrends detects trends in metric data
func (ta *TrendAnalyzer) AnalyzeTrends(ctx context.Context, metric string, eventType string, timeRange string) (*TrendResult, error) {
	// Fetch time series data
	query := fmt.Sprintf(`
		SELECT 
			average(%s) as value,
			timestamp
		FROM %s
		TIMESERIES AUTO
		SINCE %s
	`, metric, eventType, timeRange)

	result, err := ta.server.executeNRQL(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch time series: %w", err)
	}

	timeSeries, err := extractTimeSeries(result)
	if err != nil {
		return nil, err
	}

	// Analyze different trend aspects
	linearTrend := calculateLinearTrend(timeSeries)
	seasonality := detectSeasonality(timeSeries)
	changePoints := detectChangePoints(timeSeries)
	forecast := generateForecast(timeSeries, linearTrend)

	return &TrendResult{
		Metric:       metric,
		EventType:    eventType,
		TimeRange:    timeRange,
		LinearTrend:  linearTrend,
		Seasonality:  seasonality,
		ChangePoints: changePoints,
		Forecast:     forecast,
		Summary:      generateTrendSummary(linearTrend, seasonality, changePoints),
	}, nil
}

// Helper functions for anomaly detection

func detectZScoreAnomalies(series []TimeSeriesPoint, stats Statistics, sensitivity float64) []Anomaly {
	anomalies := []Anomaly{}
	threshold := 3.0 - (sensitivity * 2.0) // sensitivity 0-1 maps to z-score 3-1

	for _, point := range series {
		if stats.StdDev == 0 {
			continue
		}
		zScore := math.Abs((point.Value - stats.Mean) / stats.StdDev)
		if zScore > threshold {
			anomalies = append(anomalies, Anomaly{
				Timestamp: point.Timestamp,
				Value:     point.Value,
				Score:     zScore / 3.0, // Normalize to 0-1
				Type:      "z-score",
				Message:   fmt.Sprintf("Value %.2f is %.1f standard deviations from mean", point.Value, zScore),
			})
		}
	}

	return anomalies
}

func detectIQRAnomalies(series []TimeSeriesPoint, sensitivity float64) []Anomaly {
	anomalies := []Anomaly{}
	
	// Calculate quartiles
	values := make([]float64, len(series))
	for i, point := range series {
		values[i] = point.Value
	}
	sort.Float64s(values)

	q1 := percentile(values, 25)
	q3 := percentile(values, 75)
	iqr := q3 - q1
	
	multiplier := 1.5 + (1.0 - sensitivity) // sensitivity 0-1 maps to multiplier 2.5-1.5
	lowerBound := q1 - multiplier*iqr
	upperBound := q3 + multiplier*iqr

	for _, point := range series {
		if point.Value < lowerBound || point.Value > upperBound {
			score := 0.0
			if point.Value < lowerBound {
				score = (lowerBound - point.Value) / (q1 - lowerBound)
			} else {
				score = (point.Value - upperBound) / (upperBound - q3)
			}
			score = math.Min(1.0, score)

			anomalies = append(anomalies, Anomaly{
				Timestamp: point.Timestamp,
				Value:     point.Value,
				Score:     score,
				Type:      "iqr",
				Message:   fmt.Sprintf("Value %.2f is outside IQR bounds [%.2f, %.2f]", point.Value, lowerBound, upperBound),
			})
		}
	}

	return anomalies
}

func detectMovingAverageAnomalies(series []TimeSeriesPoint, sensitivity float64) []Anomaly {
	anomalies := []Anomaly{}
	windowSize := 12 // 1 hour for 5-minute data

	if len(series) < windowSize {
		return anomalies
	}

	for i := windowSize; i < len(series); i++ {
		// Calculate moving average
		sum := 0.0
		for j := i - windowSize; j < i; j++ {
			sum += series[j].Value
		}
		movingAvg := sum / float64(windowSize)

		// Calculate standard deviation for the window
		sumSquares := 0.0
		for j := i - windowSize; j < i; j++ {
			diff := series[j].Value - movingAvg
			sumSquares += diff * diff
		}
		windowStdDev := math.Sqrt(sumSquares / float64(windowSize))

		// Check if current point is anomalous
		threshold := 2.0 - (sensitivity * 1.5) // sensitivity 0-1 maps to 2-0.5 std devs
		diff := math.Abs(series[i].Value - movingAvg)
		if windowStdDev > 0 && diff > threshold*windowStdDev {
			score := math.Min(1.0, (diff / (3 * windowStdDev)))
			anomalies = append(anomalies, Anomaly{
				Timestamp: series[i].Timestamp,
				Value:     series[i].Value,
				Score:     score,
				Type:      "moving-average",
				Message:   fmt.Sprintf("Value %.2f deviates from moving average %.2f by %.1f std devs", 
					series[i].Value, movingAvg, diff/windowStdDev),
			})
		}
	}

	return anomalies
}

// Helper functions for correlation analysis

func calculateCorrelation(series1, series2 []TimeSeriesPoint) *MetricCorrelation {
	// Align time series by timestamp
	aligned1, aligned2 := alignTimeSeries(series1, series2)
	if len(aligned1) < 2 {
		return nil
	}

	// Calculate Pearson correlation coefficient
	n := float64(len(aligned1))
	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := range aligned1 {
		x, y := aligned1[i].Value, aligned2[i].Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return nil
	}

	coefficient := numerator / denominator

	// Calculate lag correlation
	lagCorrelations := calculateLagCorrelations(aligned1, aligned2, 5)
	bestLag := 0
	bestLagCorr := coefficient
	for lag, corr := range lagCorrelations {
		if math.Abs(corr) > math.Abs(bestLagCorr) {
			bestLag = lag
			bestLagCorr = corr
		}
	}

	return &MetricCorrelation{
		Metric:       series2[0].MetricName,
		Coefficient:  coefficient,
		Lag:          bestLag,
		LaggedCoeff:  bestLagCorr,
		DataPoints:   len(aligned1),
		Relationship: interpretCorrelation(coefficient),
	}
}

// Helper functions for trend analysis

func calculateLinearTrend(series []TimeSeriesPoint) LinearTrend {
	n := float64(len(series))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	// Use normalized time values (0 to n-1)
	for i, point := range series {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope and intercept
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Calculate R-squared
	meanY := sumY / n
	ssTotal, ssResidual := 0.0, 0.0
	for i, point := range series {
		x := float64(i)
		predicted := slope*x + intercept
		ssTotal += math.Pow(point.Value-meanY, 2)
		ssResidual += math.Pow(point.Value-predicted, 2)
	}
	rSquared := 1 - (ssResidual / ssTotal)

	// Calculate trend strength and direction
	percentChange := 0.0
	if series[0].Value != 0 {
		lastPredicted := slope*float64(n-1) + intercept
		firstPredicted := intercept
		percentChange = ((lastPredicted - firstPredicted) / firstPredicted) * 100
	}

	return LinearTrend{
		Slope:         slope,
		Intercept:     intercept,
		RSquared:      rSquared,
		Direction:     getTrendDirection(slope, series[0].Value),
		Strength:      getTrendStrength(rSquared),
		PercentChange: percentChange,
	}
}

// Data structures

type TimeSeriesPoint struct {
	Timestamp  time.Time
	Value      float64
	MetricName string
}

type Statistics struct {
	Mean   float64
	StdDev float64
	Min    float64
	Max    float64
	Count  int
}

type Anomaly struct {
	Timestamp time.Time
	Value     float64
	Score     float64
	Type      string
	Message   string
}

type AnomalyResult struct {
	Metric     string
	EventType  string
	TimeRange  string
	Statistics Statistics
	Anomalies  []Anomaly
	Summary    string
}

type MetricCorrelation struct {
	Metric       string
	Coefficient  float64
	Lag          int
	LaggedCoeff  float64
	DataPoints   int
	Relationship string
}

type CorrelationResult struct {
	PrimaryMetric string
	EventType     string
	TimeRange     string
	Correlations  []MetricCorrelation
	Summary       string
}

type LinearTrend struct {
	Slope         float64
	Intercept     float64
	RSquared      float64
	Direction     string
	Strength      string
	PercentChange float64
}

type Seasonality struct {
	Period      int
	Strength    float64
	Pattern     string
	Detected    bool
}

type ChangePoint struct {
	Timestamp  time.Time
	OldValue   float64
	NewValue   float64
	Confidence float64
	Type       string
}

type Forecast struct {
	Values     []TimeSeriesPoint
	Confidence []ConfidenceInterval
}

type ConfidenceInterval struct {
	Timestamp time.Time
	Lower     float64
	Upper     float64
}

type TrendResult struct {
	Metric       string
	EventType    string
	TimeRange    string
	LinearTrend  LinearTrend
	Seasonality  Seasonality
	ChangePoints []ChangePoint
	Forecast     Forecast
	Summary      string
}

// Utility functions

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	rank := p / 100 * float64(len(values)-1)
	lower := int(rank)
	upper := lower + 1
	if upper >= len(values) {
		return values[lower]
	}
	return values[lower] + (rank-float64(lower))*(values[upper]-values[lower])
}

func calculateStatistics(series []TimeSeriesPoint) Statistics {
	if len(series) == 0 {
		return Statistics{}
	}

	sum := 0.0
	min := series[0].Value
	max := series[0].Value

	for _, point := range series {
		sum += point.Value
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}

	mean := sum / float64(len(series))

	// Calculate standard deviation
	sumSquares := 0.0
	for _, point := range series {
		diff := point.Value - mean
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(series)))

	return Statistics{
		Mean:   mean,
		StdDev: stdDev,
		Min:    min,
		Max:    max,
		Count:  len(series),
	}
}

func interpretCorrelation(coeff float64) string {
	absCoeff := math.Abs(coeff)
	direction := "positive"
	if coeff < 0 {
		direction = "negative"
	}

	strength := ""
	switch {
	case absCoeff >= 0.9:
		strength = "very strong"
	case absCoeff >= 0.7:
		strength = "strong"
	case absCoeff >= 0.5:
		strength = "moderate"
	case absCoeff >= 0.3:
		strength = "weak"
	default:
		strength = "very weak"
	}

	return fmt.Sprintf("%s %s", strength, direction)
}

func getTrendDirection(slope float64, baseValue float64) string {
	if math.Abs(slope) < 0.0001 {
		return "flat"
	}
	if slope > 0 {
		return "increasing"
	}
	return "decreasing"
}

func getTrendStrength(rSquared float64) string {
	switch {
	case rSquared >= 0.9:
		return "very strong"
	case rSquared >= 0.7:
		return "strong"
	case rSquared >= 0.5:
		return "moderate"
	case rSquared >= 0.3:
		return "weak"
	default:
		return "very weak"
	}
}
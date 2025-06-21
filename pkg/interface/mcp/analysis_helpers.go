package mcp

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// extractTimeSeries converts NRQL results to time series points
func extractTimeSeries(result interface{}) ([]TimeSeriesPoint, error) {
	// Handle mock data
	if mockData, ok := result.(map[string]interface{}); ok {
		if results, ok := mockData["results"].([]interface{}); ok {
			series := []TimeSeriesPoint{}
			for _, r := range results {
				if point, ok := r.(map[string]interface{}); ok {
					timestamp := time.Now().Add(-time.Duration(len(series)) * 5 * time.Minute)
					if ts, ok := point["timestamp"].(float64); ok {
						timestamp = time.Unix(int64(ts)/1000, 0)
					}
					value := 0.0
					if v, ok := point["value"].(float64); ok {
						value = v
					} else if v, ok := point["average"].(float64); ok {
						value = v
					}
					series = append(series, TimeSeriesPoint{
						Timestamp: timestamp,
						Value:     value,
					})
				}
			}
			return series, nil
		}
	}

	// Handle real NRQL results
	// TODO: Implement for real NerdGraph response format
	return []TimeSeriesPoint{}, fmt.Errorf("unable to extract time series from result")
}

// deduplicateAndScore combines multiple anomaly detections
func deduplicateAndScore(anomalies []Anomaly) []Anomaly {
	// Group anomalies by timestamp
	anomalyMap := make(map[time.Time][]Anomaly)
	for _, a := range anomalies {
		anomalyMap[a.Timestamp] = append(anomalyMap[a.Timestamp], a)
	}

	// Combine and score
	combined := []Anomaly{}
	for timestamp, group := range anomalyMap {
		if len(group) == 1 {
			combined = append(combined, group[0])
		} else {
			// Multiple detections at same timestamp - combine
			maxScore := 0.0
			types := []string{}
			messages := []string{}
			
			for _, a := range group {
				if a.Score > maxScore {
					maxScore = a.Score
				}
				types = append(types, a.Type)
				messages = append(messages, a.Message)
			}

			// Boost score for multiple detections
			finalScore := math.Min(1.0, maxScore*math.Sqrt(float64(len(group))))

			combined = append(combined, Anomaly{
				Timestamp: timestamp,
				Value:     group[0].Value,
				Score:     finalScore,
				Type:      fmt.Sprintf("multi-detection (%d methods)", len(group)),
				Message:   fmt.Sprintf("Detected by %d methods", len(group)),
			})
		}
	}

	// Sort by score descending
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Score > combined[j].Score
	})

	return combined
}

// generateAnomalySummary creates a human-readable summary
func generateAnomalySummary(anomalies []Anomaly, stats Statistics) string {
	if len(anomalies) == 0 {
		return "No anomalies detected in the time range"
	}

	severe := 0
	moderate := 0
	for _, a := range anomalies {
		if a.Score > 0.8 {
			severe++
		} else if a.Score > 0.5 {
			moderate++
		}
	}

	summary := fmt.Sprintf("Detected %d anomalies: %d severe (score > 0.8), %d moderate (score > 0.5). ",
		len(anomalies), severe, moderate)

	if anomalies[0].Score > 0.9 {
		summary += fmt.Sprintf("Most significant anomaly at %s with value %.2f (normal range: %.2f - %.2f).",
			anomalies[0].Timestamp.Format("15:04"),
			anomalies[0].Value,
			stats.Mean-2*stats.StdDev,
			stats.Mean+2*stats.StdDev)
	}

	return summary
}

// alignTimeSeries aligns two time series by matching timestamps
func alignTimeSeries(series1, series2 []TimeSeriesPoint) ([]TimeSeriesPoint, []TimeSeriesPoint) {
	// Create maps for fast lookup
	map1 := make(map[time.Time]float64)
	map2 := make(map[time.Time]float64)

	for _, p := range series1 {
		map1[p.Timestamp] = p.Value
	}
	for _, p := range series2 {
		map2[p.Timestamp] = p.Value
	}

	// Find common timestamps
	aligned1 := []TimeSeriesPoint{}
	aligned2 := []TimeSeriesPoint{}

	for ts, v1 := range map1 {
		if v2, exists := map2[ts]; exists {
			aligned1 = append(aligned1, TimeSeriesPoint{Timestamp: ts, Value: v1})
			aligned2 = append(aligned2, TimeSeriesPoint{Timestamp: ts, Value: v2})
		}
	}

	// Sort by timestamp
	sort.Slice(aligned1, func(i, j int) bool {
		return aligned1[i].Timestamp.Before(aligned1[j].Timestamp)
	})
	sort.Slice(aligned2, func(i, j int) bool {
		return aligned2[i].Timestamp.Before(aligned2[j].Timestamp)
	})

	return aligned1, aligned2
}

// calculateLagCorrelations computes correlation at different time lags
func calculateLagCorrelations(series1, series2 []TimeSeriesPoint, maxLag int) map[int]float64 {
	correlations := make(map[int]float64)

	for lag := -maxLag; lag <= maxLag; lag++ {
		if lag == 0 {
			continue // Already calculated
		}

		// Create lagged series
		var lagged1, lagged2 []TimeSeriesPoint
		if lag > 0 {
			// series2 lags behind series1
			if len(series1) > lag && len(series2) > lag {
				lagged1 = series1[:len(series1)-lag]
				lagged2 = series2[lag:]
			}
		} else {
			// series1 lags behind series2
			absLag := -lag
			if len(series1) > absLag && len(series2) > absLag {
				lagged1 = series1[absLag:]
				lagged2 = series2[:len(series2)-absLag]
			}
		}

		if len(lagged1) >= 2 {
			corr := calculatePearsonCorrelation(lagged1, lagged2)
			correlations[lag] = corr
		}
	}

	return correlations
}

// calculatePearsonCorrelation computes Pearson correlation coefficient
func calculatePearsonCorrelation(series1, series2 []TimeSeriesPoint) float64 {
	n := float64(len(series1))
	if n < 2 || len(series2) != len(series1) {
		return 0
	}

	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := range series1 {
		x, y := series1[i].Value, series2[i].Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// generateCorrelationSummary creates a summary of correlations
func generateCorrelationSummary(correlations []MetricCorrelation) string {
	if len(correlations) == 0 {
		return "No significant correlations found"
	}

	strong := []string{}
	moderate := []string{}

	for _, c := range correlations {
		absCoeff := math.Abs(c.Coefficient)
		if absCoeff >= 0.7 {
			strong = append(strong, fmt.Sprintf("%s (%.2f)", c.Metric, c.Coefficient))
		} else if absCoeff >= 0.5 {
			moderate = append(moderate, fmt.Sprintf("%s (%.2f)", c.Metric, c.Coefficient))
		}
	}

	summary := ""
	if len(strong) > 0 {
		summary += fmt.Sprintf("Strong correlations: %v. ", strong)
	}
	if len(moderate) > 0 {
		summary += fmt.Sprintf("Moderate correlations: %v. ", moderate)
	}

	if correlations[0].Lag != 0 {
		summary += fmt.Sprintf("Strongest correlation is with %s at lag %d minutes.",
			correlations[0].Metric, correlations[0].Lag*5)
	}

	return summary
}

// detectSeasonality identifies seasonal patterns
func detectSeasonality(series []TimeSeriesPoint) Seasonality {
	if len(series) < 48 { // Need at least 2 days of 5-minute data
		return Seasonality{Detected: false}
	}

	// Try common periods: hourly, daily, weekly
	periods := []int{
		12,   // 1 hour (12 * 5 minutes)
		288,  // 1 day (288 * 5 minutes)
		2016, // 1 week (2016 * 5 minutes)
	}

	bestPeriod := 0
	bestStrength := 0.0

	for _, period := range periods {
		if len(series) < period*2 {
			continue
		}

		// Calculate autocorrelation at this lag
		strength := calculateAutocorrelation(series, period)
		if strength > bestStrength {
			bestPeriod = period
			bestStrength = strength
		}
	}

	if bestStrength < 0.3 {
		return Seasonality{Detected: false}
	}

	pattern := "unknown"
	switch bestPeriod {
	case 12:
		pattern = "hourly"
	case 288:
		pattern = "daily"
	case 2016:
		pattern = "weekly"
	}

	return Seasonality{
		Period:   bestPeriod,
		Strength: bestStrength,
		Pattern:  pattern,
		Detected: true,
	}
}

// calculateAutocorrelation computes autocorrelation at given lag
func calculateAutocorrelation(series []TimeSeriesPoint, lag int) float64 {
	if len(series) < lag+1 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, p := range series {
		sum += p.Value
	}
	mean := sum / float64(len(series))

	// Calculate autocorrelation
	numerator := 0.0
	denominator := 0.0

	for i := lag; i < len(series); i++ {
		numerator += (series[i].Value - mean) * (series[i-lag].Value - mean)
	}

	for _, p := range series {
		diff := p.Value - mean
		denominator += diff * diff
	}

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// detectChangePoints identifies significant changes in the time series
func detectChangePoints(series []TimeSeriesPoint) []ChangePoint {
	changePoints := []ChangePoint{}

	if len(series) < 20 {
		return changePoints
	}

	windowSize := 10
	threshold := 2.0 // Standard deviations

	for i := windowSize; i < len(series)-windowSize; i++ {
		// Calculate statistics for windows before and after
		beforeStats := calculateWindowStats(series[i-windowSize:i])
		afterStats := calculateWindowStats(series[i:i+windowSize])

		// Check for significant change in mean
		pooledStdDev := math.Sqrt((beforeStats.StdDev*beforeStats.StdDev + afterStats.StdDev*afterStats.StdDev) / 2)
		if pooledStdDev > 0 {
			tStatistic := math.Abs(beforeStats.Mean-afterStats.Mean) / (pooledStdDev * math.Sqrt(2.0/float64(windowSize)))

			if tStatistic > threshold {
				confidence := math.Min(0.99, tStatistic/4.0)
				changeType := "level_shift"
				if afterStats.Mean > beforeStats.Mean {
					changeType = "level_shift_up"
				} else {
					changeType = "level_shift_down"
				}

				changePoints = append(changePoints, ChangePoint{
					Timestamp:  series[i].Timestamp,
					OldValue:   beforeStats.Mean,
					NewValue:   afterStats.Mean,
					Confidence: confidence,
					Type:       changeType,
				})

				// Skip ahead to avoid duplicate detections
				i += windowSize / 2
			}
		}
	}

	return changePoints
}

// calculateWindowStats computes statistics for a window of points
func calculateWindowStats(window []TimeSeriesPoint) Statistics {
	if len(window) == 0 {
		return Statistics{}
	}

	sum := 0.0
	for _, p := range window {
		sum += p.Value
	}
	mean := sum / float64(len(window))

	sumSquares := 0.0
	for _, p := range window {
		diff := p.Value - mean
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(window)))

	return Statistics{
		Mean:   mean,
		StdDev: stdDev,
		Count:  len(window),
	}
}

// generateForecast creates simple forecast based on trend
func generateForecast(series []TimeSeriesPoint, trend LinearTrend) Forecast {
	if len(series) == 0 {
		return Forecast{}
	}

	// Forecast for next 12 periods (1 hour for 5-minute data)
	forecastPeriods := 12
	lastTimestamp := series[len(series)-1].Timestamp
	interval := 5 * time.Minute

	values := []TimeSeriesPoint{}
	confidence := []ConfidenceInterval{}

	// Calculate prediction interval based on residuals
	residualStdDev := calculateResidualStdDev(series, trend)

	for i := 1; i <= forecastPeriods; i++ {
		timestamp := lastTimestamp.Add(time.Duration(i) * interval)
		x := float64(len(series) + i - 1)
		predicted := trend.Slope*x + trend.Intercept

		// Widen confidence interval for farther predictions
		uncertaintyMultiplier := 1.0 + float64(i)*0.1
		margin := 1.96 * residualStdDev * uncertaintyMultiplier

		values = append(values, TimeSeriesPoint{
			Timestamp: timestamp,
			Value:     predicted,
		})

		confidence = append(confidence, ConfidenceInterval{
			Timestamp: timestamp,
			Lower:     predicted - margin,
			Upper:     predicted + margin,
		})
	}

	return Forecast{
		Values:     values,
		Confidence: confidence,
	}
}

// calculateResidualStdDev computes standard deviation of residuals
func calculateResidualStdDev(series []TimeSeriesPoint, trend LinearTrend) float64 {
	sumSquares := 0.0
	for i, point := range series {
		predicted := trend.Slope*float64(i) + trend.Intercept
		residual := point.Value - predicted
		sumSquares += residual * residual
	}
	return math.Sqrt(sumSquares / float64(len(series)))
}

// generateTrendSummary creates a summary of trend analysis
func generateTrendSummary(trend LinearTrend, seasonality Seasonality, changePoints []ChangePoint) string {
	summary := fmt.Sprintf("Trend: %s %s (%.1f%% change). ",
		trend.Strength, trend.Direction, trend.PercentChange)

	if seasonality.Detected {
		summary += fmt.Sprintf("Detected %s seasonality (strength: %.2f). ",
			seasonality.Pattern, seasonality.Strength)
	}

	if len(changePoints) > 0 {
		summary += fmt.Sprintf("Found %d significant change points. ", len(changePoints))
		if changePoints[0].Confidence > 0.8 {
			summary += fmt.Sprintf("Most significant at %s (%.1f%% confidence). ",
				changePoints[0].Timestamp.Format("15:04"),
				changePoints[0].Confidence*100)
		}
	}

	return summary
}
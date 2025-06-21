package mcp

import (
	"fmt"
	"math"
	"sort"
	"strings"
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

// generateAnomalyRecommendations creates actionable recommendations
func generateAnomalyRecommendations(anomalies []Anomaly) []string {
	recommendations := []string{}

	if len(anomalies) == 0 {
		return []string{"No anomalies detected. System appears to be operating normally."}
	}

	// Check severity
	severeCount := 0
	for _, a := range anomalies {
		if a.Score > 0.8 {
			severeCount++
		}
	}

	if severeCount > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Found %d severe anomalies. Immediate investigation recommended.", severeCount))
	}

	// Check for patterns
	if len(anomalies) > 5 {
		// Check if anomalies are clustered in time
		recommendations = append(recommendations,
			"Multiple anomalies detected. Check for system-wide issues or cascading failures.")
	}

	// Add specific recommendations based on anomaly types
	hasZScore := false
	hasMovingAvg := false
	for _, a := range anomalies {
		if a.Type == "z-score" && !hasZScore {
			hasZScore = true
			recommendations = append(recommendations,
				"Statistical outliers detected. Review system capacity and scaling policies.")
		}
		if a.Type == "moving-average" && !hasMovingAvg {
			hasMovingAvg = true
			recommendations = append(recommendations,
				"Sudden changes detected. Check for recent deployments or configuration changes.")
		}
	}

	// Add general recommendations
	recommendations = append(recommendations,
		"Create alerts based on these baseline thresholds to prevent future issues.",
		"Review correlated metrics to understand root cause.")

	return recommendations
}

// generateCorrelationInsights generates insights from correlation results
func generateCorrelationInsights(correlations []map[string]interface{}) []string {
	insights := []string{}

	if len(correlations) == 0 {
		return []string{"No significant correlations found. The metric appears to be independent."}
	}

	// Categorize correlations
	veryStrong := []string{}
	strong := []string{}
	negative := []string{}

	for _, corr := range correlations {
		coeff := corr["coefficient"].(float64)
		metric := corr["metric"].(string)
		
		if math.Abs(coeff) >= 0.9 {
			veryStrong = append(veryStrong, fmt.Sprintf("%s (%.2f)", metric, coeff))
		} else if math.Abs(coeff) >= 0.7 {
			strong = append(strong, fmt.Sprintf("%s (%.2f)", metric, coeff))
		}
		
		if coeff < -0.5 {
			negative = append(negative, metric)
		}
	}

	// Generate insights
	if len(veryStrong) > 0 {
		insights = append(insights, 
			fmt.Sprintf("Very strong correlation with: %s. These metrics move together almost perfectly.", 
				strings.Join(veryStrong, ", ")))
	}

	if len(strong) > 0 {
		insights = append(insights,
			fmt.Sprintf("Strong correlation with: %s. Consider monitoring these together.",
				strings.Join(strong, ", ")))
	}

	if len(negative) > 0 {
		insights = append(insights,
			fmt.Sprintf("Negative correlation with: %s. These metrics move in opposite directions.",
				strings.Join(negative, ", ")))
	}

	// Check for lag correlations
	for _, corr := range correlations {
		lag := 0
		if l, ok := corr["lag"].(int); ok {
			lag = l
		}
		if lag != 0 {
			metric := corr["metric"].(string)
			insights = append(insights,
				fmt.Sprintf("%s shows strongest correlation with %d minute lag. This could indicate causation.",
					metric, lag*5))
			break // Just show one lag example
		}
	}

	// Add actionable recommendations
	insights = append(insights,
		"Use these correlations to build composite alerts and dashboards.",
		"Investigate causal relationships between strongly correlated metrics.")

	return insights
}

// DistributionStats holds statistical measures for distribution analysis
type DistributionStats struct {
	Mean     float64
	Median   float64
	Mode     float64
	StdDev   float64
	Variance float64
	Skewness float64
	Kurtosis float64
	Min      float64
	Max      float64
}

// calculateDistributionStats computes comprehensive statistics for a dataset
func calculateDistributionStats(values []float64) DistributionStats {
	n := float64(len(values))
	if n == 0 {
		return DistributionStats{}
	}

	// Basic statistics
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / n

	// Median (values must be sorted)
	median := percentile(values, 50)

	// Mode (simplified - most frequent value)
	mode := calculateMode(values)

	// Variance and standard deviation
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / n
	stdDev := math.Sqrt(variance)

	// Skewness (3rd moment)
	sumCubes := 0.0
	for _, v := range values {
		diff := v - mean
		sumCubes += diff * diff * diff
	}
	skewness := 0.0
	if stdDev > 0 {
		skewness = (sumCubes / n) / math.Pow(stdDev, 3)
	}

	// Kurtosis (4th moment)
	sumQuads := 0.0
	for _, v := range values {
		diff := v - mean
		sumQuads += diff * diff * diff * diff
	}
	kurtosis := 0.0
	if variance > 0 {
		kurtosis = (sumQuads / n) / (variance * variance) - 3 // Excess kurtosis
	}

	return DistributionStats{
		Mean:     mean,
		Median:   median,
		Mode:     mode,
		StdDev:   stdDev,
		Variance: variance,
		Skewness: skewness,
		Kurtosis: kurtosis,
		Min:      values[0],          // Assumes sorted
		Max:      values[len(values)-1], // Assumes sorted
	}
}

// calculateMode finds the most frequent value
func calculateMode(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	freq := make(map[float64]int)
	maxFreq := 0
	mode := values[0]

	for _, v := range values {
		freq[v]++
		if freq[v] > maxFreq {
			maxFreq = freq[v]
			mode = v
		}
	}

	return mode
}

// createHistogram creates histogram buckets from values
func createHistogram(values []float64, numBuckets int) []map[string]interface{} {
	if len(values) == 0 || numBuckets <= 0 {
		return []map[string]interface{}{}
	}

	min := values[0]
	max := values[len(values)-1]
	bucketSize := (max - min) / float64(numBuckets)

	histogram := make([]map[string]interface{}, numBuckets)
	for i := 0; i < numBuckets; i++ {
		start := min + float64(i)*bucketSize
		end := start + bucketSize
		if i == numBuckets-1 {
			end = max + 0.001 // Include max value in last bucket
		}

		count := 0
		for _, v := range values {
			if v >= start && v < end {
				count++
			}
		}

		histogram[i] = map[string]interface{}{
			"bucket_start": start,
			"bucket_end":   end,
			"count":        count,
			"percentage":   float64(count) / float64(len(values)) * 100,
		}
	}

	return histogram
}

// detectDistributionType analyzes the shape of the distribution
func detectDistributionType(values []float64, stats DistributionStats) string {
	// Simplified distribution detection based on skewness and kurtosis
	absSkew := math.Abs(stats.Skewness)

	if absSkew < 0.5 && math.Abs(stats.Kurtosis) < 0.5 {
		return "normal"
	} else if stats.Skewness > 1 {
		return "right-skewed"
	} else if stats.Skewness < -1 {
		return "left-skewed"
	} else if stats.Kurtosis > 1 {
		return "leptokurtic" // Heavy tails
	} else if stats.Kurtosis < -1 {
		return "platykurtic" // Light tails
	} else {
		return "non-normal"
	}
}

// generateDistributionInsights creates insights from distribution analysis
func generateDistributionInsights(stats DistributionStats, distributionType string, histogram []map[string]interface{}) []string {
	insights := []string{}

	// Distribution type insight
	switch distributionType {
	case "normal":
		insights = append(insights, "Distribution appears to be approximately normal (Gaussian).")
	case "right-skewed":
		insights = append(insights, "Distribution is right-skewed with a long tail of high values.")
	case "left-skewed":
		insights = append(insights, "Distribution is left-skewed with a long tail of low values.")
	case "leptokurtic":
		insights = append(insights, "Distribution has heavy tails with more extreme values than normal.")
	case "platykurtic":
		insights = append(insights, "Distribution has light tails with fewer extreme values than normal.")
	}

	// Spread insight
	cv := stats.StdDev / stats.Mean * 100 // Coefficient of variation
	if cv > 100 {
		insights = append(insights, fmt.Sprintf("Very high variability (CV=%.1f%%). Values are highly dispersed.", cv))
	} else if cv > 50 {
		insights = append(insights, fmt.Sprintf("High variability (CV=%.1f%%). Consider investigating outliers.", cv))
	} else if cv < 20 {
		insights = append(insights, fmt.Sprintf("Low variability (CV=%.1f%%). Values are tightly clustered.", cv))
	}

	// Central tendency comparison
	if math.Abs(stats.Mean-stats.Median) > stats.StdDev*0.2 {
		if stats.Mean > stats.Median {
			insights = append(insights, "Mean is significantly higher than median, indicating outliers on the high end.")
		} else {
			insights = append(insights, "Mean is significantly lower than median, indicating outliers on the low end.")
		}
	}

	// Find most populated bucket
	if len(histogram) > 0 {
		maxBucket := histogram[0]
		for _, bucket := range histogram {
			if bucket["count"].(int) > maxBucket["count"].(int) {
				maxBucket = bucket
			}
		}
		insights = append(insights, 
			fmt.Sprintf("Most values (%.1f%%) fall between %.2f and %.2f.", 
				maxBucket["percentage"], maxBucket["bucket_start"], maxBucket["bucket_end"]))
	}

	// Actionable recommendations
	if distributionType != "normal" {
		insights = append(insights, "Consider using percentile-based thresholds instead of mean-based for alerting.")
	}
	if cv > 50 {
		insights = append(insights, "High variability suggests need for adaptive thresholds or anomaly detection.")
	}

	return insights
}

// processSegmentResults converts NRQL faceted results into segment data
func processSegmentResults(result map[string]interface{}, metric string, segmentBy string, comparisonType string) []map[string]interface{} {
	segments := []map[string]interface{}{}
	
	// Extract facets from result
	if facets, ok := result["facets"].([]interface{}); ok {
		for _, f := range facets {
			if facet, ok := f.(map[string]interface{}); ok {
				name := ""
				if nameList, ok := facet["name"].([]interface{}); ok && len(nameList) > 0 {
					name = fmt.Sprintf("%v", nameList[0])
				}
				
				if results, ok := facet["results"].([]interface{}); ok && len(results) > 0 {
					if data, ok := results[0].(map[string]interface{}); ok {
						segment := map[string]interface{}{
							"name":    name,
							"avg":     data["avg"],
							"count":   data["count"],
							"stddev":  data["stddev"],
							"min":     data["min"],
							"max":     data["max"],
						}
						
						// Extract percentiles if available
						if percentiles, ok := data["percentiles"].(map[string]interface{}); ok {
							segment["p50"] = percentiles["50"]
							segment["p90"] = percentiles["90"]
							segment["p95"] = percentiles["95"]
						}
						
						segments = append(segments, segment)
					}
				}
			}
		}
	}
	
	// Sort segments by average value descending
	sort.Slice(segments, func(i, j int) bool {
		avg1, _ := segments[i]["avg"].(float64)
		avg2, _ := segments[j]["avg"].(float64)
		return avg1 > avg2
	})
	
	// Add ranks and calculate relative values
	if len(segments) > 0 {
		baselineAvg, _ := segments[0]["avg"].(float64)
		totalCount := 0.0
		
		for _, s := range segments {
			if count, ok := s["count"].(float64); ok {
				totalCount += count
			}
		}
		
		for i, segment := range segments {
			segment["rank"] = i + 1
			
			// Calculate relative value
			if avg, ok := segment["avg"].(float64); ok {
				if comparisonType == "relative" && baselineAvg > 0 {
					segment["relative"] = avg / baselineAvg
				}
			}
			
			// Calculate percentage of total
			if count, ok := segment["count"].(float64); ok && totalCount > 0 {
				segment["percent_of_total"] = (count / totalCount) * 100
			}
		}
	}
	
	return segments
}

// analyzeSegmentDifferences performs statistical analysis on segments
func analyzeSegmentDifferences(segments []map[string]interface{}, comparisonType string) map[string]interface{} {
	if len(segments) == 0 {
		return map[string]interface{}{}
	}
	
	analysis := map[string]interface{}{}
	
	// Extract averages for analysis
	averages := []float64{}
	for _, s := range segments {
		if avg, ok := s["avg"].(float64); ok {
			averages = append(averages, avg)
		}
	}
	
	if len(averages) > 0 {
		// Calculate overall statistics
		sum := 0.0
		for _, v := range averages {
			sum += v
		}
		overallMean := sum / float64(len(averages))
		
		// Calculate coefficient of variation across segments
		sumSquares := 0.0
		for _, v := range averages {
			diff := v - overallMean
			sumSquares += diff * diff
		}
		stdDev := math.Sqrt(sumSquares / float64(len(averages)))
		cv := (stdDev / overallMean) * 100
		
		analysis["overall_mean"] = overallMean
		analysis["overall_stddev"] = stdDev
		analysis["coefficient_of_variation"] = cv
		
		// Find outliers (segments significantly different from mean)
		outliers := []string{}
		for _, s := range segments {
			if avg, ok := s["avg"].(float64); ok {
				if stdDev > 0 && math.Abs(avg-overallMean) > 2*stdDev {
					name, _ := s["name"].(string)
					outliers = append(outliers, name)
				}
			}
		}
		analysis["outliers"] = outliers
		
		// Calculate range and spread
		sort.Float64s(averages)
		analysis["min_segment_avg"] = averages[0]
		analysis["max_segment_avg"] = averages[len(averages)-1]
		analysis["range"] = averages[len(averages)-1] - averages[0]
		analysis["range_ratio"] = averages[len(averages)-1] / averages[0]
	}
	
	return analysis
}

// generateSegmentInsights creates insights from segment comparison
func generateSegmentInsights(segments []map[string]interface{}, analysis map[string]interface{}, comparisonType string) []string {
	insights := []string{}
	
	if len(segments) == 0 {
		return []string{"No segments found for comparison."}
	}
	
	// Top performer insight
	if len(segments) > 0 {
		topName, _ := segments[0]["name"].(string)
		topAvg, _ := segments[0]["avg"].(float64)
		insights = append(insights, 
			fmt.Sprintf("%s has the highest average value (%.2f).", topName, topAvg))
	}
	
	// Bottom performer insight (if multiple segments)
	if len(segments) > 1 {
		bottomIdx := len(segments) - 1
		bottomName, _ := segments[bottomIdx]["name"].(string)
		bottomAvg, _ := segments[bottomIdx]["avg"].(float64)
		topAvg, _ := segments[0]["avg"].(float64)
		
		if topAvg > 0 {
			pctDiff := ((topAvg - bottomAvg) / topAvg) * 100
			insights = append(insights,
				fmt.Sprintf("%s has the lowest average (%.2f), which is %.1f%% lower than the top performer.",
					bottomName, bottomAvg, pctDiff))
		}
	}
	
	// Variation insight
	if cv, ok := analysis["coefficient_of_variation"].(float64); ok {
		if cv > 50 {
			insights = append(insights, 
				fmt.Sprintf("High variation across segments (CV=%.1f%%). Performance is very inconsistent.", cv))
		} else if cv < 20 {
			insights = append(insights,
				fmt.Sprintf("Low variation across segments (CV=%.1f%%). Performance is relatively consistent.", cv))
		}
	}
	
	// Outlier insights
	if outliers, ok := analysis["outliers"].([]string); ok && len(outliers) > 0 {
		insights = append(insights,
			fmt.Sprintf("Outlier segments detected: %s. These require special attention.",
				strings.Join(outliers, ", ")))
	}
	
	// Volume distribution insight
	totalVolume := 0.0
	for _, s := range segments {
		if count, ok := s["count"].(float64); ok {
			totalVolume += count
		}
	}
	
	if totalVolume > 0 && len(segments) > 0 {
		topCount, _ := segments[0]["count"].(float64)
		topName, _ := segments[0]["name"].(string)
		topPct := (topCount / totalVolume) * 100
		
		if topPct > 50 {
			insights = append(insights,
				fmt.Sprintf("%s accounts for %.1f%% of total volume, indicating concentration risk.",
					topName, topPct))
		}
	}
	
	// Actionable recommendations
	if rangeRatio, ok := analysis["range_ratio"].(float64); ok && rangeRatio > 2 {
		insights = append(insights,
			"Large performance gap between segments. Consider segment-specific optimization strategies.")
	}
	
	if len(segments) > 10 {
		insights = append(insights,
			"Many segments detected. Consider grouping or filtering to focus on key segments.")
	}
	
	return insights
}

// processBaselineResults converts NRQL baseline query results into structured format
func processBaselineResults(result map[string]interface{}, metric string, percentiles []interface{}, groupBy string) map[string]interface{} {
	baseline := map[string]interface{}{
		"metric": metric,
	}
	
	// Handle results based on whether it's grouped or not
	if groupBy != "" {
		// Faceted results
		baseline["grouped_by"] = groupBy
		groups := []map[string]interface{}{}
		
		if facets, ok := result["facets"].([]interface{}); ok {
			for _, f := range facets {
				if facet, ok := f.(map[string]interface{}); ok {
					groupName := ""
					if nameList, ok := facet["name"].([]interface{}); ok && len(nameList) > 0 {
						groupName = fmt.Sprintf("%v", nameList[0])
					}
					
					if results, ok := facet["results"].([]interface{}); ok && len(results) > 0 {
						if data, ok := results[0].(map[string]interface{}); ok {
							group := map[string]interface{}{
								"name":   groupName,
								"avg":    data["average"],
								"stddev": data["stddev"],
								"min":    data["min"],
								"max":    data["max"],
								"count":  data["count"],
							}
							
							// Add percentiles
							for i, p := range percentiles {
								key := fmt.Sprintf("p%v", p)
								if val, ok := data[key]; ok {
									group[key] = val
								} else if i < len(percentiles) {
									// Try alternate format
									altKey := fmt.Sprintf("percentile_%v", p)
									if val, ok := data[altKey]; ok {
										group[key] = val
									}
								}
							}
							
							groups = append(groups, group)
						}
					}
				}
			}
		}
		baseline["groups"] = groups
	} else {
		// Single result
		if results, ok := result["results"].([]interface{}); ok && len(results) > 0 {
			if data, ok := results[0].(map[string]interface{}); ok {
				baseline["avg"] = data["average"]
				baseline["stddev"] = data["stddev"]
				baseline["min"] = data["min"]
				baseline["max"] = data["max"]
				baseline["count"] = data["count"]
				
				// Add percentiles
				for _, p := range percentiles {
					key := fmt.Sprintf("p%v", p)
					if val, ok := data[key]; ok {
						baseline[key] = val
					}
				}
			}
		}
	}
	
	// Add recommendations based on the baseline
	baseline["recommendations"] = generateBaselineRecommendations(baseline)
	
	return baseline
}

// generateBaselineRecommendations creates actionable recommendations from baseline data
func generateBaselineRecommendations(baseline map[string]interface{}) []string {
	recommendations := []string{}
	
	metric, _ := baseline["metric"].(string)
	
	// Check if grouped or single baseline
	if groups, ok := baseline["groups"].([]map[string]interface{}); ok && len(groups) > 0 {
		// Analyze grouped baselines
		avgValues := []float64{}
		for _, g := range groups {
			if avg, ok := g["avg"].(float64); ok {
				avgValues = append(avgValues, avg)
			}
		}
		
		if len(avgValues) > 1 {
			// Calculate variation
			sum := 0.0
			for _, v := range avgValues {
				sum += v
			}
			mean := sum / float64(len(avgValues))
			
			maxDiff := 0.0
			for _, v := range avgValues {
				diff := math.Abs(v - mean)
				if diff > maxDiff {
					maxDiff = diff
				}
			}
			
			if mean > 0 && maxDiff/mean > 0.5 {
				recommendations = append(recommendations,
					fmt.Sprintf("High variation across groups for %s. Consider group-specific thresholds.", metric))
			}
		}
	} else {
		// Single baseline recommendations
		if avg, ok := baseline["avg"].(float64); ok {
			if stddev, ok := baseline["stddev"].(float64); ok {
				cv := stddev / avg * 100
				
				if cv > 50 {
					recommendations = append(recommendations,
						fmt.Sprintf("High variability in %s (CV=%.1f%%). Use percentile-based thresholds.", metric, cv))
				}
				
				// Alert threshold recommendations
				if p95, ok := baseline["p95"].(float64); ok {
					recommendations = append(recommendations,
						fmt.Sprintf("Consider setting warning threshold at %.2f (p95)", p95))
				}
				if p99, ok := baseline["p99"].(float64); ok {
					recommendations = append(recommendations,
						fmt.Sprintf("Consider setting critical threshold at %.2f (p99)", p99))
				}
			}
			
			// Normal range recommendation
			if p10, ok := baseline["p10"].(float64); ok {
				if p90, ok := baseline["p90"].(float64); ok {
					recommendations = append(recommendations,
						fmt.Sprintf("Normal range for %s is %.2f-%.2f (p10-p90)", metric, p10, p90))
				}
			}
		}
	}
	
	// General recommendations
	if count, ok := baseline["count"].(float64); ok && count < 100 {
		recommendations = append(recommendations,
			"Low sample size. Consider expanding time range for more reliable baseline.")
	}
	
	return recommendations
}
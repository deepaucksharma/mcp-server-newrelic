# Analysis Tools Documentation

This document details the analysis tools **as actually implemented** in the MCP Server.

## Overview

Analysis tools showcase sophisticated statistical algorithms but **only process mock data**. The algorithms are real; the data is fake.

## Implementation Status

| Tool | Status | Real Functionality |
|------|--------|-------------------|
| `analysis.calculate_baseline` | 🟨 Mock | Real statistics on fake data |
| `analysis.detect_anomalies` | 🟨 Mock | Real algorithms on fake data |
| `analysis.find_correlations` | 🟨 Mock | Real correlation math on fake data |
| `analysis.analyze_trend` | 🟨 Mock | Real trend analysis on fake data |
| `analysis.analyze_distribution` | 🟨 Mock | Real distribution analysis on fake data |
| `analysis.compare_segments` | 🟨 Mock | Real comparison logic on fake data |

## The Irony

These tools contain some of the most sophisticated code in the project:
- Statistical calculations
- Anomaly detection algorithms  
- Correlation analysis
- Trend forecasting
- Distribution analysis

**But they never fetch real data from New Relic.**

## How Analysis Tools Work (With Mock Data)

### analysis.calculate_baseline

**Purpose**: Calculate statistical baseline for a metric.

**Implementation Files**: 
- `pkg/interface/mcp/tools_analysis.go`
- `pkg/interface/mcp/analysis_helpers.go`
- `pkg/interface/mcp/analysis_algorithms.go`

**What Actually Happens**:
```go
func (s *Server) handleAnalysisCalculateBaseline(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    metric := params["metric"].(string)
    eventType := params["event_type"].(string)
    
    // Always true - no real data connection
    if s.isMockMode() {
        return s.getMockData("analysis.calculate_baseline", params), nil
    }
    
    // This code exists but never runs:
    // - Would build NRQL query
    // - Would fetch historical data
    // - Would calculate real statistics
    
    // Instead, mock data generator creates:
    return map[string]interface{}{
        "baseline": map[string]interface{}{
            "avg": 125.5,      // Random
            "stddev": 45.2,    // Random
            "min": 10.0,       // Random
            "max": 500.0,      // Random
            "p50": 110.0,      // Random
            "p95": 220.0,      // Random
        },
    }
}
```

**The Algorithms (Never Used)**:
```go
// This beautiful code exists but processes mock data
func calculateDistributionStats(values []float64) DistributionStats {
    n := float64(len(values))
    
    // Calculate mean
    mean := sum(values) / n
    
    // Calculate variance
    variance := 0.0
    for _, v := range values {
        variance += math.Pow(v - mean, 2)
    }
    variance /= n
    
    // Calculate skewness
    skewness := 0.0
    for _, v := range values {
        skewness += math.Pow((v - mean) / math.Sqrt(variance), 3)
    }
    skewness /= n
    
    // More sophisticated math...
    // All operating on fake data!
}
```

### analysis.detect_anomalies

**Purpose**: Detect anomalies using statistical methods.

**Available Methods**:
- Z-score detection
- Interquartile range (IQR)
- Isolation forest (stub)

**Example Mock Response**:
```json
{
  "anomalies": [
    {
      "timestamp": "2024-01-20T10:30:00Z",
      "value": 450.5,
      "zscore": 3.2,
      "severity": "high",
      "method": "zscore",
      "context": "Value is 3.2 standard deviations above mean"
    }
  ],
  "statistics": {
    "total_points": 1440,
    "anomalies_found": 3,
    "anomaly_rate": 0.002
  },
  "thresholds": {
    "zscore": 3.0,
    "iqr_multiplier": 1.5
  }
}
```

**The Algorithms**:
```go
// Real anomaly detection that never sees real data
func (a *AnomalyDetector) detectUsingZScore(values []float64, sensitivity float64) []Anomaly {
    mean := calculateMean(values)
    stdDev := calculateStdDev(values, mean)
    threshold := 3.0 - (sensitivity * 2.0)  // Adjust by sensitivity
    
    anomalies := []Anomaly{}
    for i, value := range values {
        zScore := math.Abs((value - mean) / stdDev)
        if zScore > threshold {
            anomalies = append(anomalies, Anomaly{
                Index:    i,
                Value:    value,
                ZScore:   zScore,
                Severity: classifySeverity(zScore),
            })
        }
    }
    return anomalies
}
```

### analysis.find_correlations

**Purpose**: Find correlations between metrics.

**Real Math on Fake Data**:
```go
// Pearson correlation coefficient - correctly implemented
func calculateCorrelation(x, y []float64) float64 {
    n := float64(len(x))
    sumX, sumY := 0.0, 0.0
    sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0
    
    for i := range x {
        sumX += x[i]
        sumY += y[i]
        sumXY += x[i] * y[i]
        sumX2 += x[i] * x[i]
        sumY2 += y[i] * y[i]
    }
    
    correlation := (n*sumXY - sumX*sumY) / 
        math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))
    
    return correlation
}
```

**Example Mock Response**:
```json
{
  "correlations": [
    {
      "metric": "cpuPercent",
      "coefficient": 0.89,
      "lag": 0,
      "confidence": 0.95,
      "relationship": "strong positive"
    },
    {
      "metric": "memoryUsage",
      "coefficient": -0.72,
      "lag": 5,
      "confidence": 0.90,
      "relationship": "moderate negative"
    }
  ]
}
```

### analysis.analyze_trend

**Purpose**: Analyze trends with forecasting.

**Includes**:
- Linear regression
- Trend direction detection
- Simple forecasting
- Seasonality detection (mock)

**The Math**:
```go
// Real linear regression on fake data
func calculateLinearTrend(timestamps, values []float64) LinearTrend {
    n := float64(len(values))
    sumX, sumY := 0.0, 0.0
    sumXY, sumX2 := 0.0, 0.0
    
    for i := range values {
        x := float64(i)
        sumX += x
        sumY += values[i]
        sumXY += x * values[i]
        sumX2 += x * x
    }
    
    slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
    intercept := (sumY - slope*sumX) / n
    
    // Calculate R-squared
    // More math...
    
    return LinearTrend{
        Slope:     slope,
        Intercept: intercept,
        RSquared:  rSquared,
        Direction: getTrendDirection(slope),
    }
}
```

## Why Mock Data?

### The Data Fetching Gap

The tools expect data like this:
```go
// What should happen
query := fmt.Sprintf(`
    SELECT %s 
    FROM %s 
    SINCE %s 
    LIMIT MAX
`, metric, eventType, timeRange)

results := s.nrClient.Query(ctx, query)
values := extractValues(results)
// Process with algorithms
```

But instead:
```go
// What actually happens
values := generateMockTimeSeries(1000)  // Fake data
// Process with same algorithms
```

### The Fundamental Problem

1. **No Batch Data Fetching**: Algorithms need arrays of values, but tools fetch aggregates
2. **API Limitations**: NerdGraph returns aggregated data, not raw values
3. **Performance**: Fetching thousands of data points would be slow
4. **Design Mismatch**: Tools designed for data science, API designed for dashboards

## Working Around the Limitations

### Manual Statistical Analysis

Use `query_nrdb` to get aggregated stats:

```javascript
// Get baseline statistics manually
const baseline = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(duration) as avg,
      stddev(duration) as stddev,
      min(duration) as min,
      max(duration) as max,
      percentile(duration, 50, 90, 95, 99)
    FROM Transaction 
    SINCE 7 days ago
  `
});

// Calculate threshold manually
const threshold = baseline.avg + (3 * baseline.stddev);
```

### Anomaly Detection Workaround

```javascript
// Detect anomalies using NRQL
const anomalies = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      duration,
      timestamp
    FROM Transaction 
    WHERE duration > (
      SELECT percentile(duration, 99) 
      FROM Transaction 
      SINCE 7 days ago
    )
    SINCE 1 hour ago
  `
});
```

### Correlation Analysis Workaround

```javascript
// Check correlation between metrics
const correlation = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      correlation(duration, databaseDuration) as corr
    FROM Transaction 
    SINCE 1 day ago
  `
});
```

## The Implemented Algorithms

Despite using mock data, these algorithms are correctly implemented:

### Statistical Functions
- Mean, median, mode
- Standard deviation, variance
- Skewness, kurtosis
- Percentiles
- Min/max

### Anomaly Detection
- Z-score method
- IQR method
- Threshold-based detection

### Correlation Analysis
- Pearson correlation
- Lag correlation
- Correlation strength classification

### Trend Analysis  
- Linear regression
- Trend direction
- Strength measurement
- Simple forecasting

### Distribution Analysis
- Histogram generation
- Distribution type detection
- Statistical moments

## If These Tools Used Real Data

Here's what would need to change:

```go
// 1. Fetch raw data points
func fetchTimeSeriesData(ctx context.Context, metric, eventType, timeRange string) ([]DataPoint, error) {
    query := fmt.Sprintf(
        "SELECT %s, timestamp FROM %s SINCE %s LIMIT MAX",
        metric, eventType, timeRange
    )
    
    results, err := s.nrClient.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    
    return extractDataPoints(results), nil
}

// 2. Process with existing algorithms
dataPoints := fetchTimeSeriesData(ctx, metric, eventType, timeRange)
values := extractValues(dataPoints)
anomalies := detector.detectUsingZScore(values, sensitivity)
```

## Summary

Analysis tools in the MCP Server are a **beautiful tragedy**:
- Sophisticated algorithms implemented correctly
- Statistical methods properly coded
- Complex analysis logic ready to use
- **But they only process fake data**

The irony: These are some of the best-implemented tools in the codebase, but they're completely disconnected from New Relic data.

For real analysis:
1. Use `query_nrdb` with statistical NRQL functions
2. Export data and analyze externally
3. Use New Relic's built-in ML features
4. Build custom analysis with real data fetching

The analysis tools showcase what could have been - a powerful statistical analysis engine for observability data. Instead, they're an elaborate demonstration of algorithms running on synthetic data.
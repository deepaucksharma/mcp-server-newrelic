# Query Patterns Examples

This document provides practical NRQL query patterns for common observability use cases using the MCP Server's `query_nrdb` tool.

## Overview

Since `query_nrdb` is the only fully functional query tool, these examples focus on real NRQL patterns that work with actual New Relic data.

## Basic Query Patterns

### Simple Aggregations

```javascript
// Count events
const count = await mcp.call("query_nrdb", {
  query: "SELECT count(*) FROM Transaction SINCE 1 hour ago"
});

// Average metric
const avgDuration = await mcp.call("query_nrdb", {
  query: "SELECT average(duration) FROM Transaction SINCE 1 hour ago"
});

// Multiple aggregations
const stats = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      count(*) as total,
      average(duration) as avg_duration,
      max(duration) as max_duration,
      min(duration) as min_duration
    FROM Transaction 
    SINCE 1 hour ago
  `
});
```

### Time Series Queries

```javascript
// Basic time series
const timeSeries = await mcp.call("query_nrdb", {
  query: `
    SELECT average(duration) 
    FROM Transaction 
    TIMESERIES 5 minutes 
    SINCE 1 hour ago
  `
});

// Multiple metrics over time
const multiMetricSeries = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(duration) as avg_duration,
      count(*) as throughput,
      percentage(count(*), WHERE error) as error_rate
    FROM Transaction 
    TIMESERIES 1 minute 
    SINCE 30 minutes ago
  `
});

// Compare time periods
const comparison = await mcp.call("query_nrdb", {
  query: `
    SELECT average(duration) 
    FROM Transaction 
    SINCE 1 hour ago 
    COMPARE WITH 1 hour ago
  `
});
```

### Grouping with FACET

```javascript
// Group by single attribute
const byApp = await mcp.call("query_nrdb", {
  query: `
    SELECT count(*), average(duration) 
    FROM Transaction 
    FACET appName 
    SINCE 1 hour ago
  `
});

// Group by multiple attributes
const byAppAndHost = await mcp.call("query_nrdb", {
  query: `
    SELECT count(*), average(duration) 
    FROM Transaction 
    FACET appName, host 
    SINCE 1 hour ago 
    LIMIT 50
  `
});

// Nested grouping with time
const byAppOverTime = await mcp.call("query_nrdb", {
  query: `
    SELECT average(duration) 
    FROM Transaction 
    FACET appName 
    TIMESERIES 5 minutes 
    SINCE 1 hour ago
  `
});
```

## Performance Analysis Patterns

### Response Time Analysis

```javascript
// Percentile analysis
const percentiles = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(duration) as avg,
      percentile(duration, 50) as median,
      percentile(duration, 90) as p90,
      percentile(duration, 95) as p95,
      percentile(duration, 99) as p99
    FROM Transaction 
    WHERE appName = 'production-api'
    SINCE 1 hour ago
  `
});

// Slow transaction analysis
const slowTransactions = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      name,
      average(duration) as avg_duration,
      count(*) as count
    FROM Transaction 
    WHERE duration > 1
    FACET name
    SINCE 1 hour ago
    LIMIT 20
  `
});

// Response time distribution
const distribution = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      histogram(duration, width: 0.1, buckets: 20) 
    FROM Transaction 
    WHERE appName = 'production-api'
    SINCE 1 hour ago
  `
});
```

### Throughput Analysis

```javascript
// Requests per minute
const rpm = await mcp.call("query_nrdb", {
  query: `
    SELECT rate(count(*), 1 minute) as rpm
    FROM Transaction 
    TIMESERIES 1 minute
    SINCE 1 hour ago
  `
});

// Peak vs average throughput
const throughputStats = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(rpm) as avg_rpm,
      max(rpm) as peak_rpm,
      min(rpm) as min_rpm
    FROM (
      SELECT rate(count(*), 1 minute) as rpm
      FROM Transaction 
      TIMESERIES 1 minute
      SINCE 24 hours ago
    )
  `
});

// Throughput by endpoint
const endpointThroughput = await mcp.call("query_nrdb", {
  query: `
    SELECT rate(count(*), 1 minute) as rpm
    FROM Transaction 
    FACET name
    SINCE 1 hour ago
    LIMIT 20
  `
});
```

## Error Analysis Patterns

### Error Rate Calculation

```javascript
// Overall error rate
const errorRate = await mcp.call("query_nrdb", {
  query: `
    SELECT percentage(count(*), WHERE error = true) as error_rate
    FROM Transaction 
    SINCE 1 hour ago
  `
});

// Error rate by application
const errorsByApp = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      count(*) as total,
      filter(count(*), WHERE error = true) as errors,
      percentage(count(*), WHERE error = true) as error_rate
    FROM Transaction 
    FACET appName
    SINCE 1 hour ago
  `
});

// Error rate trend
const errorTrend = await mcp.call("query_nrdb", {
  query: `
    SELECT percentage(count(*), WHERE error = true) as error_rate
    FROM Transaction 
    TIMESERIES 5 minutes
    SINCE 6 hours ago
  `
});
```

### Error Details

```javascript
// Error types and messages
const errorTypes = await mcp.call("query_nrdb", {
  query: `
    SELECT count(*)
    FROM TransactionError 
    FACET error.class, error.message
    SINCE 1 hour ago
    LIMIT 50
  `
});

// Errors by transaction
const errorsByTransaction = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      count(*) as error_count,
      latest(error.message) as latest_error
    FROM TransactionError 
    FACET transactionName
    SINCE 1 hour ago
    LIMIT 20
  `
});

// Error correlation with metrics
const errorCorrelation = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(duration) as avg_duration_with_error
    FROM Transaction 
    WHERE error = true
    SINCE 1 hour ago
    COMPARE WITH
      SELECT average(duration) as avg_duration_no_error
      FROM Transaction 
      WHERE error = false
  `
});
```

## Infrastructure Patterns

### Host Metrics

```javascript
// CPU usage by host
const cpuByHost = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(cpuPercent) as avg_cpu,
      max(cpuPercent) as max_cpu
    FROM SystemSample 
    FACET hostname
    SINCE 1 hour ago
    LIMIT 20
  `
});

// Memory usage patterns
const memoryUsage = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(memoryUsedPercent) as avg_memory,
      max(memoryUsedPercent) as max_memory,
      average(memoryFreeBytes/1e9) as avg_free_gb
    FROM SystemSample 
    WHERE hostname LIKE 'prod-%'
    TIMESERIES 5 minutes
    SINCE 2 hours ago
  `
});

// Disk I/O analysis
const diskIO = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(diskReadBytesPerSecond/1e6) as read_mbps,
      average(diskWriteBytesPerSecond/1e6) as write_mbps
    FROM SystemSample 
    FACET hostname
    SINCE 30 minutes ago
  `
});
```

### Container Metrics

```javascript
// Container resource usage
const containerResources = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(cpuUsedCores) as avg_cpu_cores,
      average(memoryUsageBytes/1e9) as avg_memory_gb
    FROM ContainerSample 
    FACET containerName
    SINCE 1 hour ago
    LIMIT 20
  `
});

// Container restarts
const containerRestarts = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      max(restartCount) - min(restartCount) as restarts
    FROM ContainerSample 
    FACET containerName
    WHERE restartCount is not null
    SINCE 24 hours ago
  `
});
```

## Advanced Query Patterns

### Subqueries

```javascript
// Find outliers using subquery
const outliers = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      name,
      average(duration) as avg_duration
    FROM Transaction 
    WHERE duration > (
      SELECT percentile(duration, 95) 
      FROM Transaction 
      SINCE 1 hour ago
    )
    FACET name
    SINCE 1 hour ago
  `
});

// Compare to baseline
const vsBaseline = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(duration) as current_avg,
      (
        SELECT average(duration) 
        FROM Transaction 
        SINCE 1 week ago 
        UNTIL 1 day ago
      ) as baseline_avg
    FROM Transaction 
    SINCE 1 hour ago
  `
});
```

### JOIN Queries

```javascript
// Join Transaction and Error data
const joinedData = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      count(*) as error_count,
      average(Transaction.duration) as avg_duration
    FROM Transaction, TransactionError
    WHERE Transaction.guid = TransactionError.guid
    SINCE 1 hour ago
  `
});

// Multi-event correlation
const correlation = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      count(*) as correlated_events
    FROM Transaction, Log
    WHERE Transaction.traceId = Log.trace.id
      AND Transaction.timestamp >= Log.timestamp - 1000
      AND Transaction.timestamp <= Log.timestamp + 1000
    SINCE 1 hour ago
  `
});
```

### Statistical Functions

```javascript
// Standard deviation analysis
const variability = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(duration) as mean,
      stddev(duration) as std_dev,
      stddev(duration) / average(duration) as coefficient_of_variation
    FROM Transaction 
    FACET appName
    SINCE 1 hour ago
  `
});

// Correlation between metrics
const metricCorrelation = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      correlation(duration, databaseDuration) as db_correlation,
      correlation(duration, externalDuration) as external_correlation
    FROM Transaction 
    WHERE databaseDuration > 0 AND externalDuration > 0
    SINCE 1 hour ago
  `
});
```

## Alerting Query Patterns

### Alert-Friendly Queries

```javascript
// Simple threshold alert
const alertQuery1 = await mcp.call("query_nrdb", {
  query: `
    SELECT average(duration)
    FROM Transaction 
    WHERE appName = 'critical-service'
  `
});

// Error rate alert
const alertQuery2 = await mcp.call("query_nrdb", {
  query: `
    SELECT percentage(count(*), WHERE error = true)
    FROM Transaction 
    WHERE appName = 'critical-service'
  `
});

// Anomaly detection query
const alertQuery3 = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(duration) as current,
      stddev(duration) as deviation
    FROM Transaction 
    WHERE appName = 'critical-service'
  `
});
```

## Dashboard Query Patterns

### Multi-Widget Queries

```javascript
// KPI Overview
const kpiQueries = [
  // Response Time
  "SELECT average(duration), percentile(duration, 95) FROM Transaction SINCE 1 hour ago",
  
  // Error Rate
  "SELECT percentage(count(*), WHERE error = true) FROM Transaction SINCE 1 hour ago",
  
  // Throughput
  "SELECT rate(count(*), 1 minute) FROM Transaction SINCE 1 hour ago",
  
  // Apdex
  "SELECT apdex(duration, t: 0.5) FROM Transaction SINCE 1 hour ago"
];

// Execute all KPI queries
const kpiResults = await Promise.all(
  kpiQueries.map(query => mcp.call("query_nrdb", { query }))
);
```

### Time Comparison Queries

```javascript
// Week over week comparison
const weekComparison = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      average(duration) as 'This Week'
    FROM Transaction 
    SINCE 1 week ago 
    COMPARE WITH 1 week ago
    TIMESERIES 1 day
  `
});

// Hour over hour
const hourComparison = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      count(*) as 'Current Hour'
    FROM Transaction 
    SINCE 1 hour ago 
    COMPARE WITH 1 hour ago
    TIMESERIES 10 minutes
  `
});
```

## Best Practices

1. **Always include SINCE clause** - Prevents scanning all data
2. **Use LIMIT with FACET** - Prevents overwhelming results
3. **Aggregate before FACET** - More efficient than raw data
4. **Use appropriate TIMESERIES buckets** - Match your time range
5. **Test queries at small scale first** - Use shorter time ranges

## Summary

These query patterns demonstrate:
- Basic to advanced NRQL syntax
- Common performance analysis queries
- Error tracking patterns
- Infrastructure monitoring
- Statistical analysis
- Alert-friendly queries

Remember: `query_nrdb` is your primary tool for real data analysis in the MCP Server.
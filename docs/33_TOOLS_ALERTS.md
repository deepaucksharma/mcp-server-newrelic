# Alert Tools Documentation

This document details the alert tools **as actually implemented** in the MCP Server.

## Overview

Alert tools manage New Relic alert conditions. Basic creation and listing work, but advanced features are mocked.

## Implementation Status

| Tool | Status | Real Functionality |
|------|--------|-------------------|
| `create_alert` | ⚠️ Partial | Creates basic alerts, mock baselines |
| `list_alerts` | ⚠️ Partial | Lists alerts, mock incidents |
| `alert.update` | ❌ Not Implemented | No handler |
| `alert.delete` | ❌ Not Implemented | No handler |
| `alerts.test` | 🟨 Mock | Returns fake test results |
| `alerts.recommend` | 🟨 Mock | Returns generic recommendations |

## Working Tools

### create_alert

**Purpose**: Create a new alert condition in New Relic.

**Implementation File**: `pkg/interface/mcp/tools_alerts.go`

**Parameters**:
```json
{
  "name": "High Error Rate",                    // string, required
  "query": "SELECT count(*) FROM Transaction",  // string, required  
  "sensitivity": "medium",                      // string, optional: low/medium/high
  "comparison": "above",                        // string, optional: above/below/equals
  "threshold_duration": 5,                      // integer, optional (minutes)
  "auto_baseline": true,                        // boolean, optional
  "static_threshold": 100,                      // number, optional (if not auto)
  "policy_id": "123456",                        // string, optional
  "account_id": "789012"                        // string, optional
}
```

**Actual Implementation**:
```go
func (s *Server) handleCreateAlert(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    name := params["name"].(string)
    query := params["query"].(string)
    
    // Mock mode check
    if s.isMockMode() {
        return mockAlertCreation(params), nil
    }
    
    // Auto baseline is always mocked
    threshold := 0.0
    if autoBaseline, _ := params["auto_baseline"].(bool); autoBaseline {
        // Fake baseline calculation
        threshold = calculateMockBaseline(query, params["sensitivity"])
    } else {
        threshold = params["static_threshold"].(float64)
    }
    
    // Create basic NRQL alert condition
    alertConfig := map[string]interface{}{
        "name": name,
        "nrql": map[string]interface{}{
            "query": query,
        },
        "terms": []map[string]interface{}{
            {
                "threshold": threshold,
                "thresholdDuration": params["threshold_duration"],
                "operator": params["comparison"],
                "priority": "CRITICAL",
            },
        },
    }
    
    // GraphQL mutation to create alert
    result, err := s.createAlertCondition(ctx, alertConfig)
    return result, err
}
```

**What Works**:
- Creates basic NRQL alert conditions
- Sets static thresholds
- Configures comparison operators
- Sets threshold duration

**What Doesn't Work**:
- Auto-baseline always returns mock values
- No intelligent threshold detection
- No anomaly-based alerts
- Limited to single condition
- No notification channel linking

**Example Usage**:

```json
// Static threshold alert
{
  "name": "create_alert",
  "arguments": {
    "name": "High Response Time Alert",
    "query": "SELECT average(duration) FROM Transaction",
    "comparison": "above",
    "static_threshold": 1.0,
    "threshold_duration": 10,
    "auto_baseline": false
  }
}

// Response
{
  "alert_id": "12345",
  "policy_id": "67890",
  "status": "created",
  "condition": {
    "name": "High Response Time Alert",
    "threshold": 1.0,
    "enabled": true
  }
}
```

```json
// Auto-baseline alert (threshold is mocked)
{
  "name": "create_alert",
  "arguments": {
    "name": "Anomaly Detection Alert",
    "query": "SELECT count(*) FROM TransactionError",
    "sensitivity": "high",
    "auto_baseline": true
  }
}

// Response (baseline is fake)
{
  "alert_id": "23456",
  "baseline_info": {
    "calculated_threshold": 156.7,    // FAKE
    "confidence": 0.95,               // FAKE
    "based_on_days": 7                // FAKE
  }
}
```

### list_alerts

**Purpose**: List existing alert conditions.

**Implementation File**: `pkg/interface/mcp/tools_alerts.go`

**Parameters**:
```json
{
  "policy_id": "123456",         // string, optional
  "enabled_only": false,         // boolean, optional
  "include_incidents": false,    // boolean, optional
  "limit": 50,                   // integer, optional
  "cursor": ""                   // string, optional (pagination)
}
```

**Actual Implementation**:
```go
func (s *Server) handleListAlerts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    if s.isMockMode() {
        return mockAlertsList(params), nil
    }
    
    // Build GraphQL query
    query := buildAlertsListQuery(params)
    result, err := s.nrClient.Query(ctx, query)
    
    // Process results
    alerts := processAlertResults(result)
    
    // Incidents are always mocked
    if includeIncidents, _ := params["include_incidents"].(bool); includeIncidents {
        for i := range alerts {
            alerts[i]["incidents"] = generateMockIncidents()
        }
    }
    
    return alerts, nil
}
```

**What Works**:
- Lists actual alert conditions
- Filters by policy ID
- Basic pagination
- Shows alert configuration

**What Doesn't Work**:
- Incident data is always mocked
- No real-time status
- Limited filtering options
- No performance metrics

**Example Usage**:

```json
// List all alerts
{
  "name": "list_alerts",
  "arguments": {
    "limit": 10
  }
}

// Response
{
  "alerts": [
    {
      "id": "12345",
      "name": "High Error Rate",
      "enabled": true,
      "query": "SELECT percentage(count(*), WHERE error) FROM Transaction",
      "threshold": 5.0,
      "comparison": "above",
      "policy": {
        "id": "67890",
        "name": "Production Alerts"
      }
    }
  ],
  "cursor": "next-page-cursor"
}
```

## Mock-Only Tools

### alerts.test

**Purpose**: Test alert conditions with historical data.

**Reality**: Returns fake test results.

**Example Response**:
```json
{
  "test_results": {
    "would_have_triggered": 3,           // FAKE
    "trigger_times": [                   // FAKE
      "2024-01-15T10:30:00Z",
      "2024-01-15T14:45:00Z",
      "2024-01-15T22:15:00Z"
    ],
    "accuracy_score": 0.87,              // FAKE
    "false_positive_rate": 0.13         // FAKE
  }
}
```

### alerts.recommend

**Purpose**: Get alert recommendations based on data.

**Reality**: Returns generic recommendations.

**Example Response**:
```json
{
  "recommendations": [
    {
      "metric": "error_percentage",
      "suggested_threshold": 2.5,         // FAKE
      "confidence": 0.89,                 // FAKE
      "reasoning": "Based on 30-day historical patterns"  // FAKE
    }
  ]
}
```

## Common Alert Patterns

### Error Rate Alert
```json
{
  "name": "create_alert",
  "arguments": {
    "name": "Error Rate Alert",
    "query": "SELECT percentage(count(*), WHERE error = true) FROM Transaction",
    "comparison": "above",
    "static_threshold": 5.0,
    "threshold_duration": 5,
    "auto_baseline": false
  }
}
```

### Response Time Alert
```json
{
  "name": "create_alert",
  "arguments": {
    "name": "Slow Response Time",
    "query": "SELECT average(duration) FROM Transaction",
    "comparison": "above", 
    "static_threshold": 1.0,
    "threshold_duration": 10
  }
}
```

### Traffic Drop Alert
```json
{
  "name": "create_alert",
  "arguments": {
    "name": "Traffic Drop Alert",
    "query": "SELECT count(*) FROM Transaction",
    "comparison": "below",
    "static_threshold": 1000,
    "threshold_duration": 15
  }
}
```

## Alert Query Best Practices

### 1. Use Aggregations
```sql
-- Good (single value per evaluation)
SELECT average(duration) FROM Transaction

-- Bad (multiple values)
SELECT duration FROM Transaction
```

### 2. Include Time Windows
```sql
-- Good (explicit window)
SELECT count(*) FROM Transaction SINCE 5 minutes ago

-- Bad (relies on condition setting)
SELECT count(*) FROM Transaction
```

### 3. Filter Appropriately
```sql
-- Good (specific to app)
SELECT average(duration) FROM Transaction WHERE appName = 'production'

-- Bad (all apps mixed)
SELECT average(duration) FROM Transaction
```

## Limitations

### API Limitations
- Only NRQL conditions supported
- Limited to basic threshold alerts
- No composite conditions
- No outlier detection

### Implementation Limitations
- Auto-baseline is fake
- No ML-based thresholds
- No anomaly detection
- Single account only
- No alert correlation

## Troubleshooting

### Alert Not Creating

**Common Issues**:
1. Invalid NRQL query
2. Missing policy ID
3. Invalid threshold value

**Debug Steps**:
```javascript
// 1. Validate query first
await mcp.call("query_check", {
  query: "SELECT average(duration) FROM Transaction"
});

// 2. Create with explicit values
await mcp.call("create_alert", {
  name: "Test Alert",
  query: "SELECT count(*) FROM Transaction", 
  comparison: "above",
  static_threshold: 100,
  auto_baseline: false  // Avoid mock baseline
});
```

### Alert Not Triggering

**Common Issues**:
1. Threshold too high/low
2. Duration too long
3. Query returns no data

**Debug Query**:
```javascript
// Test the query directly
await mcp.call("query_nrdb", {
  query: "SELECT average(duration) FROM Transaction SINCE 30 minutes ago TIMESERIES"
});
```

## Working with Mock Baselines

Since auto-baseline returns fake values:

### Option 1: Calculate Manually
```javascript
// Get historical data
const stats = await mcp.call("query_nrdb", {
  query: "SELECT average(value), stddev(value) FROM Metric SINCE 7 days ago"
});

// Calculate threshold (mean + 3*stddev)
const threshold = stats.average + (3 * stats.stddev);

// Create alert with calculated threshold
await mcp.call("create_alert", {
  name: "Calculated Baseline Alert",
  query: "SELECT average(value) FROM Metric",
  static_threshold: threshold,
  auto_baseline: false
});
```

### Option 2: Use Percentiles
```javascript
// Get 95th percentile as threshold
const percentile = await mcp.call("query_nrdb", {
  query: "SELECT percentile(duration, 95) FROM Transaction SINCE 7 days ago"
});

await mcp.call("create_alert", {
  name: "Percentile-based Alert",
  query: "SELECT average(duration) FROM Transaction",
  static_threshold: percentile.results[0]['percentile.duration'],
  auto_baseline: false
});
```

## Future Improvements Needed

1. **Real Baseline Calculation**
   - Implement statistical baseline
   - Add seasonality detection
   - Support anomaly detection

2. **Enhanced Alert Types**
   - Composite conditions
   - Outlier detection
   - Forecast-based alerts

3. **Better Management**
   - Update existing alerts
   - Delete alerts
   - Bulk operations

4. **Integration**
   - Link to notification channels
   - Alert correlation
   - Runbook integration

## Summary

Alert tools in the MCP Server are basic:
- Can create simple threshold alerts
- Can list existing alerts
- Auto-baseline is always fake
- No advanced alert types
- Missing update/delete operations

For production alerting:
1. Calculate thresholds manually
2. Use static thresholds
3. Keep alerts simple
4. Monitor alert performance
5. Don't rely on recommendations
# Examples Overview

This document provides practical examples of using the MCP Server with real-world scenarios and common use cases.

## Basic Query Examples

### Simple Transaction Count

```javascript
// Get transaction count
const result = await mcp.call("query_nrdb", {
  query: "SELECT count(*) FROM Transaction SINCE 1 hour ago"
});

console.log(`Transactions: ${result.data.results[0].count}`);
```

### Time Series Query

```javascript
// Get response time trend
const trend = await mcp.call("query_nrdb", {
  query: `
    SELECT average(duration) 
    FROM Transaction 
    TIMESERIES 5 minutes 
    SINCE 1 hour ago
  `
});

// Process time series data
trend.data.results.forEach(point => {
  console.log(`${point.beginTimeSeconds}: ${point['average.duration']}s`);
});
```

### Faceted Query

```javascript
// Break down by application
const breakdown = await mcp.call("query_nrdb", {
  query: `
    SELECT count(*), average(duration) 
    FROM Transaction 
    FACET appName 
    SINCE 1 hour ago 
    LIMIT 10
  `
});
```

### Error Rate Analysis

```javascript
// Calculate error rates by application
const errorRates = await mcp.call("query_nrdb", {
  query: `
    SELECT 
      count(*) as total_requests,
      filter(count(*), WHERE error IS true) as error_count,
      percentage(count(*), WHERE error IS true) as error_rate
    FROM Transaction 
    FACET appName 
    SINCE 1 hour ago
  `
});
```

## Discovery Examples

### Explore Your Data Landscape

```javascript
// 1. Discover event types
const eventTypes = await mcp.call("discovery.explore_event_types", {
  limit: 20
});

console.log("Available event types:");
eventTypes.event_types.forEach(e => console.log(`- ${e.name}`));

// 2. Explore attributes for interesting event types
const attributes = await mcp.call("discovery.explore_attributes", {
  event_type: "Transaction",
  include_samples: true
});

console.log("Transaction attributes:");
attributes.attributes.forEach(attr => {
  console.log(`- ${attr.name} (${attr.type})`);
});
```

### Schema Analysis

```javascript
// Get comprehensive schema information
const schemas = await mcp.call("discovery.list_schemas", {
  include_quality: true
});

schemas.forEach(schema => {
  console.log(`Schema: ${schema.name}`);
  console.log(`Event Types: ${schema.event_types.join(', ')}`);
  console.log(`Quality Score: ${schema.quality_score}`);
});
```

### Attribute Profiling

```javascript
// Deep analysis of specific attributes
const profile = await mcp.call("discovery.profile_attribute", {
  schema: "Transaction",
  attribute: "duration",
  time_range: "24 hours"
});

console.log(`Duration Analysis:`);
console.log(`- Range: ${profile.statistics.min} to ${profile.statistics.max}`);
console.log(`- Average: ${profile.statistics.mean}`);
console.log(`- 95th percentile: ${profile.statistics.percentiles.p95}`);
```

## Alert Management Examples

### Create Performance Alert

```javascript
// Create alert based on baseline analysis
const alert = await mcp.call("alert.create_from_baseline", {
  name: "High Response Time Alert",
  event_type: "Transaction",
  metric: "duration",
  sensitivity: "high",
  notification_channels: ["email-channel-1"]
});

console.log(`Alert created: ${alert.alert_id}`);
console.log(`Calculated threshold: ${alert.baseline_info.calculated_threshold}`);
```

### Custom Alert Condition

```javascript
// Create custom alert with specific thresholds
const customAlert = await mcp.call("alert.create_custom", {
  name: "Error Rate Alert",
  query: "SELECT percentage(count(*), WHERE error IS true) FROM Transaction",
  threshold: 5.0,
  comparison: "above",
  duration: 10
});
```

### List and Filter Alerts

```javascript
// Get all active alerts
const alerts = await mcp.call("list_alerts", {
  filter: {
    enabled: true,
    severity: "critical"
  },
  include_conditions: true,
  limit: 50
});

alerts.forEach(alert => {
  console.log(`${alert.name}: ${alert.status}`);
});
```

## Dashboard Examples

### Auto-Generated Dashboard

```javascript
// Create dashboard from discovered data patterns
const dashboard = await mcp.call("dashboard.create_from_discovery", {
  name: "Application Performance Dashboard",
  event_types: ["Transaction", "TransactionError"],
  layout: "grid",
  time_range: "3 hours"
});

console.log("Generated dashboard:");
console.log(JSON.stringify(dashboard, null, 2));
```

### Custom Dashboard with Widgets

```javascript
// Create custom dashboard with specific widgets
const customDashboard = await mcp.call("dashboard.create_custom", {
  name: "Custom Metrics Dashboard",
  widgets: [
    {
      title: "Response Time Trend",
      query: "SELECT average(duration) FROM Transaction TIMESERIES",
      visualization: "line",
      position: { x: 0, y: 0, width: 6, height: 4 }
    },
    {
      title: "Throughput",
      query: "SELECT rate(count(*), 1 minute) FROM Transaction TIMESERIES",
      visualization: "area",
      position: { x: 6, y: 0, width: 6, height: 4 }
    }
  ]
});
```

### Dashboard Usage Analysis

```javascript
// Find dashboards using specific metrics
const usage = await mcp.call("find_usage", {
  search_term: "duration",
  search_type: "metric",
  include_widgets: true
});

console.log("Dashboards using 'duration':");
usage.dashboards.forEach(d => {
  console.log(`- ${d.name} (${d.widgets_using_term.length} widgets)`);
});
```

## Analysis Examples

### Baseline Calculation

```javascript
// Calculate statistical baseline for a metric
const baseline = await mcp.call("analysis.calculate_baseline", {
  event_type: "Transaction",
  metric: "duration",
  time_range: "7 days",
  method: "statistical"
});

console.log("Baseline Analysis:");
console.log(`Mean: ${baseline.statistics.mean}`);
console.log(`Standard Deviation: ${baseline.statistics.std_dev}`);
console.log(`95% Confidence Interval: [${baseline.confidence_interval.lower}, ${baseline.confidence_interval.upper}]`);
```

### Anomaly Detection

```javascript
// Detect anomalies in metric data
const anomalies = await mcp.call("analysis.detect_anomalies", {
  metric: "duration",
  event_type: "Transaction",
  time_range: "24 hours",
  sensitivity: 0.8,
  method: "zscore"
});

console.log(`Found ${anomalies.anomalies.length} anomalies`);
anomalies.anomalies.forEach(a => {
  console.log(`- ${a.timestamp}: ${a.value} (z-score: ${a.zscore})`);
});
```

### Correlation Analysis

```javascript
// Find correlations between metrics
const correlations = await mcp.call("analysis.find_correlations", {
  primary_metric: "duration",
  secondary_metrics: ["cpuPercent", "memoryUsage", "errorCount"],
  event_type: "Transaction",
  time_range: "24 hours"
});

correlations.correlations.forEach(c => {
  console.log(`${c.metric}: ${c.coefficient.toFixed(2)} (${c.relationship})`);
});
```

### Trend Analysis

```javascript
// Analyze trends with forecasting
const trends = await mcp.call("analysis.analyze_trend", {
  metric: "throughput",
  event_type: "Transaction",
  time_range: "30 days",
  forecast_periods: 7
});

console.log(`Trend direction: ${trends.trend_direction}`);
console.log(`Trend strength: ${trends.trend_strength}`);
console.log(`Forecast: ${trends.forecast.map(f => f.predicted_value).join(', ')}`);
```

## Governance Examples

### Usage Analysis

```javascript
// Analyze data ingest patterns
const usage = await mcp.call("governance.analyze_usage", {
  time_range: "30 days",
  group_by: "source"
});

console.log("Data Ingest by Source:");
usage.breakdown.forEach(item => {
  console.log(`${item.source}: ${item.volume_gb} GB (${item.cost_usd} USD)`);
});
```

### Cost Optimization

```javascript
// Get cost optimization recommendations
const optimization = await mcp.call("governance.optimize_costs", {
  target_reduction: 20,
  preserve_critical: true
});

console.log("Cost Optimization Recommendations:");
optimization.recommendations.forEach(rec => {
  console.log(`- ${rec.description}: Save ${rec.estimated_savings} USD/month`);
});
```

### Compliance Checking

```javascript
// Check compliance status
const compliance = await mcp.call("governance.check_compliance", {
  compliance_type: "retention",
  detailed: true
});

console.log(`Compliance Status: ${compliance.status}`);
if (compliance.violations.length > 0) {
  console.log("Violations found:");
  compliance.violations.forEach(v => {
    console.log(`- ${v.description}: ${v.severity}`);
  });
}
```

## Workflow Examples

### Complete Investigation Workflow

```javascript
// Execute automated investigation
const investigation = await mcp.call("workflow.execute_investigation", {
  issue_description: "Application response times degraded after deployment",
  time_range: "3 hours",
  auto_create_dashboard: true
});

console.log("Investigation Results:");
console.log(`Root Cause: ${investigation.root_cause}`);
console.log(`Confidence: ${investigation.confidence}`);
console.log(`Dashboard Created: ${investigation.dashboard_url}`);
```

### Account Optimization Workflow

```javascript
// Run comprehensive account optimization
const optimization = await mcp.call("workflow.optimize_account", {
  optimization_goals: ["cost", "performance"],
  aggressive: false
});

console.log("Optimization Plan:");
optimization.plan.steps.forEach((step, index) => {
  console.log(`${index + 1}. ${step.description}`);
  console.log(`   Expected Impact: ${step.expected_impact}`);
});
```

### Report Generation

```javascript
// Generate executive summary report
const report = await mcp.call("workflow.generate_report", {
  report_type: "executive",
  time_range: "30 days",
  format: "markdown"
});

console.log("Executive Report Generated:");
console.log(report.content);
```

## Advanced Patterns

### Performance Investigation Workflow

```javascript
async function investigatePerformance(appName) {
  console.log(`Investigating performance for ${appName}...`);
  
  // 1. Get current performance metrics
  const current = await mcp.call("query_nrdb", {
    query: `
      SELECT average(duration), percentile(duration, 95), count(*)
      FROM Transaction 
      WHERE appName = '${appName}'
      SINCE 30 minutes ago
    `
  });
  
  // 2. Calculate baseline for comparison
  const baseline = await mcp.call("analysis.calculate_baseline", {
    event_type: "Transaction",
    metric: "duration",
    time_range: "7 days",
    filter: `appName = '${appName}'`
  });
  
  // 3. Detect anomalies
  const anomalies = await mcp.call("analysis.detect_anomalies", {
    metric: "duration",
    event_type: "Transaction",
    time_range: "24 hours",
    filter: `appName = '${appName}'`
  });
  
  // 4. Find correlations with infrastructure metrics
  const correlations = await mcp.call("analysis.find_correlations", {
    primary_metric: "duration",
    secondary_metrics: ["cpuPercent", "memoryUsage"],
    event_type: "Transaction",
    time_range: "24 hours"
  });
  
  return {
    current_performance: current.data.results[0],
    baseline: baseline.statistics,
    anomalies: anomalies.anomalies,
    correlations: correlations.correlations
  };
}
```

### Automated Monitoring Setup

```javascript
async function setupMonitoring(appName) {
  // 1. Discover the application's data structure
  const attributes = await mcp.call("discovery.explore_attributes", {
    event_type: "Transaction",
    filter: `appName = '${appName}'`
  });
  
  // 2. Calculate baselines for key metrics
  const durationBaseline = await mcp.call("analysis.calculate_baseline", {
    event_type: "Transaction",
    metric: "duration",
    time_range: "7 days"
  });
  
  // 3. Create performance alert
  const performanceAlert = await mcp.call("alert.create_from_baseline", {
    name: `${appName} - Performance Alert`,
    event_type: "Transaction",
    metric: "duration",
    sensitivity: "medium"
  });
  
  // 4. Create error rate alert
  const errorAlert = await mcp.call("alert.create_custom", {
    name: `${appName} - Error Rate Alert`,
    query: `SELECT percentage(count(*), WHERE error IS true) FROM Transaction WHERE appName = '${appName}'`,
    threshold: 5.0,
    comparison: "above"
  });
  
  // 5. Generate monitoring dashboard
  const dashboard = await mcp.call("dashboard.create_from_discovery", {
    name: `${appName} - Monitoring Dashboard`,
    event_types: ["Transaction", "TransactionError"],
    filter: `appName = '${appName}'`
  });
  
  return {
    alerts: [performanceAlert, errorAlert],
    dashboard: dashboard,
    baseline: durationBaseline
  };
}
```

### Multi-Environment Analysis

```javascript
async function compareEnvironments(environments) {
  const results = {};
  
  for (const env of environments) {
    // Get performance metrics for each environment
    const metrics = await mcp.call("query_nrdb", {
      query: `
        SELECT 
          average(duration) as avg_duration,
          percentile(duration, 95) as p95_duration,
          count(*) as throughput,
          percentage(count(*), WHERE error IS true) as error_rate
        FROM Transaction 
        WHERE environment = '${env}'
        SINCE 1 hour ago
      `
    });
    
    results[env] = metrics.data.results[0];
  }
  
  // Compare performance across environments
  const comparison = await mcp.call("analysis.compare_segments", {
    metric: "duration",
    event_type: "Transaction",
    segment_by: "environment",
    time_range: "24 hours"
  });
  
  return {
    environment_metrics: results,
    statistical_comparison: comparison
  };
}
```

## Integration Examples

### Express.js API Integration

```javascript
const express = require('express');
const { MCPClient } = require('./mcp-client');

const app = express();
const mcp = new MCPClient({ command: '/path/to/mcp-server' });

// Get application metrics endpoint
app.get('/api/metrics/:app', async (req, res) => {
  try {
    const result = await mcp.call("query_nrdb", {
      query: `
        SELECT average(duration), count(*), percentage(count(*), WHERE error)
        FROM Transaction 
        WHERE appName = '${req.params.app}'
        SINCE 1 hour ago
      `
    });
    
    res.json(result.data.results[0]);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// Health check endpoint with anomaly detection
app.get('/api/health/:app', async (req, res) => {
  const anomalies = await mcp.call("analysis.detect_anomalies", {
    metric: "duration",
    event_type: "Transaction",
    time_range: "1 hour",
    filter: `appName = '${req.params.app}'`
  });
  
  const health = anomalies.anomalies.length === 0 ? 'healthy' : 'degraded';
  res.json({ status: health, anomalies: anomalies.anomalies });
});
```

### Scheduled Monitoring

```javascript
const cron = require('node-cron');

// Check performance every 5 minutes
cron.schedule('*/5 * * * *', async () => {
  const apps = ['web-app', 'api-service', 'background-worker'];
  
  for (const app of apps) {
    const anomalies = await mcp.call("analysis.detect_anomalies", {
      metric: "duration",
      event_type: "Transaction",
      time_range: "5 minutes",
      filter: `appName = '${app}'`
    });
    
    if (anomalies.anomalies.length > 0) {
      console.warn(`Performance anomaly detected in ${app}:`, anomalies.anomalies);
      // Send alert notification
    }
  }
});
```

## Testing Examples

### Unit Tests

```javascript
describe('MCP Integration', () => {
  let mcp;
  
  beforeEach(() => {
    mcp = new MCPClient({ 
      command: '/path/to/mcp-server',
      env: { MOCK_MODE: 'true' }
    });
  });
  
  it('should query transaction data', async () => {
    const result = await mcp.call("query_nrdb", {
      query: "SELECT count(*) FROM Transaction SINCE 1 hour ago"
    });
    
    expect(result).toHaveProperty('data');
    expect(result.data).toHaveProperty('results');
    expect(result.data.results[0]).toHaveProperty('count');
  });
  
  it('should detect performance baselines', async () => {
    const baseline = await mcp.call("analysis.calculate_baseline", {
      event_type: "Transaction",
      metric: "duration",
      time_range: "7 days"
    });
    
    expect(baseline.statistics).toHaveProperty('mean');
    expect(baseline.statistics).toHaveProperty('std_dev');
  });
});
```

## Best Practices

### Error Handling

```javascript
async function safeQuery(query) {
  try {
    // Validate query first
    const validation = await mcp.call("query_check", {
      query: query
    });
    
    if (!validation.valid) {
      throw new Error(`Invalid query: ${validation.errors.join(', ')}`);
    }
    
    // Execute validated query
    return await mcp.call("query_nrdb", { query });
    
  } catch (error) {
    console.error("Query failed:", error.message);
    throw error;
  }
}
```

### Performance Optimization

```javascript
async function optimizedDiscovery() {
  // Use session for stateful operations
  const session = await mcp.call("session.create");
  
  try {
    // Cache discovery results in session
    const eventTypes = await mcp.call("discovery.explore_event_types", {
      session_id: session.session_id,
      limit: 100
    });
    
    // Parallel attribute discovery
    const attributePromises = eventTypes.event_types.map(et => 
      mcp.call("discovery.explore_attributes", {
        event_type: et.name,
        session_id: session.session_id
      })
    );
    
    const attributes = await Promise.all(attributePromises);
    return { eventTypes, attributes };
    
  } finally {
    // Clean up session
    await mcp.call("session.end", {
      session_id: session.session_id
    });
  }
}
```

These examples demonstrate the comprehensive capabilities of the MCP Server for New Relic observability, from basic queries to advanced analysis and automation workflows.
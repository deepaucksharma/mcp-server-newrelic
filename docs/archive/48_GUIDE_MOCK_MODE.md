# Mock Mode Usage Guide

This guide explains how to use the MCP Server's mock mode for development, testing, and demonstrations without a New Relic account.

## Overview

Mock mode allows the MCP Server to run without New Relic credentials, returning realistic fake data for all tools. It's surprisingly sophisticated - sometimes too sophisticated.

## Starting Mock Mode

### Command Line

```bash
# Start server in mock mode
./bin/mcp-server -mock

# Or use environment variable
MOCK_MODE=true ./bin/mcp-server
```

### No Credentials Needed

In mock mode, you don't need:
- `NEW_RELIC_API_KEY`
- `NEW_RELIC_ACCOUNT_ID`
- Any New Relic access

## What Mock Mode Provides

### Realistic Data Patterns

The mock generator creates:
- Believable metric values
- Proper time series data
- Realistic error rates
- Valid NRQL query responses
- Complex dashboard structures
- Statistical distributions

### Consistent Responses

Mock data is:
- Deterministic for same inputs
- Varies appropriately over time
- Maintains relationships between metrics
- Follows New Relic data patterns

## Tool Behavior in Mock Mode

### Query Tools

```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT average(duration) FROM Transaction SINCE 1 hour ago"
  }
}
```

**Mock Response:**
```json
{
  "results": [{
    "average.duration": 0.234,
    "beginTimeSeconds": 1642521600
  }],
  "metadata": {
    "eventTypes": ["Transaction"],
    "messages": [],
    "beginTime": 1642521600000,
    "endTime": 1642525200000
  }
}
```

The mock understands:
- Aggregation functions (avg, min, max, sum, count)
- Time ranges (SINCE, UNTIL, BETWEEN)
- FACET clauses
- TIMESERIES
- Basic WHERE conditions

### Discovery Tools

```json
{
  "name": "discovery.explore_event_types",
  "arguments": {}
}
```

**Mock Response:**
```json
{
  "event_types": [
    {"name": "Transaction", "category": "APM"},
    {"name": "TransactionError", "category": "APM"},
    {"name": "PageView", "category": "Browser"},
    {"name": "Span", "category": "Distributed Tracing"},
    {"name": "Log", "category": "Logging"},
    {"name": "Metric", "category": "Metrics"}
  ]
}
```

Always returns common New Relic event types.

### Alert Tools

```json
{
  "name": "create_alert",
  "arguments": {
    "name": "Test Alert",
    "query": "SELECT count(*) FROM Transaction",
    "threshold": 100
  }
}
```

**Mock Response:**
```json
{
  "alert_id": "mock-alert-12345",
  "status": "created",
  "message": "Alert created successfully (mock mode)"
}
```

Creates fake alert IDs that look real.

### Dashboard Tools

Mock mode excels at dashboard generation:

```json
{
  "name": "generate_dashboard",
  "arguments": {
    "template": "golden-signals",
    "name": "Test Dashboard"
  }
}
```

Returns complete dashboard JSON with:
- Proper widget configurations
- Valid NRQL queries
- Correct layout structure
- Realistic widget types

### Analysis Tools

The most sophisticated mocks:

```json
{
  "name": "analysis.detect_anomalies",
  "arguments": {
    "metric": "duration",
    "event_type": "Transaction"
  }
}
```

Returns:
- Statistically plausible anomalies
- Proper z-scores
- Realistic timestamps
- Believable severity levels

## Mock Data Patterns

### Time Series Data

Mock generator creates:
- Diurnal patterns (daily cycles)
- Weekly patterns
- Random variations
- Occasional spikes
- Gradual trends

### Metric Values

Common patterns:
- Response times: 0.1 - 2.0 seconds
- Error rates: 0.1% - 5%
- Throughput: 100 - 10,000 rpm
- CPU usage: 20% - 80%
- Memory: 40% - 90%

### Relationships

Mock data maintains:
- Error rate inversely related to performance
- CPU and memory correlation
- Transaction volume patterns
- Consistent app names and hosts

## Using Mock Mode for Development

### Testing Tool Integration

```javascript
// Test your integration without real data
const mcp = new MCPClient({
  command: '/path/to/mcp-server',
  args: ['-mock']
});

// All tools work with fake data
const result = await mcp.call("query_nrdb", {
  query: "SELECT count(*) FROM Transaction"
});

console.log(result); // Realistic mock response
```

### Developing Workflows

```javascript
// Build complex workflows with consistent mock data
async function investigatePerformance() {
  // 1. Discover event types (mock)
  const events = await mcp.call("discovery.explore_event_types");
  
  // 2. Query metrics (mock)
  const metrics = await mcp.call("query_nrdb", {
    query: "SELECT average(duration) FROM Transaction TIMESERIES"
  });
  
  // 3. Detect anomalies (mock)
  const anomalies = await mcp.call("analysis.detect_anomalies", {
    metric: "duration",
    event_type: "Transaction"
  });
  
  // All return coordinated mock data
}
```

### UI Development

Mock mode is perfect for:
- Building dashboards without real data
- Testing error handling
- Demonstrating capabilities
- Creating screenshots

## Mock Mode Limitations

### What It Doesn't Do

1. **No Data Persistence**
   - Each query generates fresh data
   - No historical consistency
   - Can't query for specific past anomalies

2. **No Real Relationships**
   - JOIN queries return empty results
   - Can't track specific entities
   - No real trace IDs or session IDs

3. **Limited Query Understanding**
   - Complex WHERE clauses ignored
   - Advanced functions return defaults
   - Subqueries not supported

### Differences from Real Mode

| Feature | Real Mode | Mock Mode |
|---------|-----------|-----------|
| Data Source | New Relic API | Generated |
| Query Validation | Full NRQL | Basic parsing |
| Historical Data | Actual | Generated fresh |
| Relationships | Real | None |
| Performance | API latency | Instant |
| Rate Limits | Apply | None |

## Advanced Mock Patterns

### Predictable Testing

The mock generator uses seeds for consistency:

```javascript
// Same query returns similar patterns
for (let i = 0; i < 5; i++) {
  const result = await mcp.call("query_nrdb", {
    query: "SELECT average(duration) FROM Transaction"
  });
  // Results vary but follow patterns
}
```

### Error Simulation

Mock mode can simulate errors:

```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "INVALID QUERY SYNTAX"
  }
}
```

Returns appropriate error responses.

### Load Testing

Since mock mode has no rate limits:

```javascript
// Test your application's handling of MCP
const promises = [];
for (let i = 0; i < 100; i++) {
  promises.push(mcp.call("query_nrdb", {
    query: `SELECT count(*) FROM Transaction`
  }));
}
await Promise.all(promises); // No rate limiting
```

## Best Practices

### 1. Development First

Always develop with mock mode:
- No API costs
- Instant responses
- Predictable data
- No credentials needed

### 2. Clear Labeling

When using mock mode:
```javascript
const IS_MOCK = process.env.MOCK_MODE === 'true';

if (IS_MOCK) {
  console.log("⚠️  Running in mock mode - data is simulated");
}
```

### 3. Transition Testing

Test both modes:
```javascript
// Test with mock
const mockResult = await mockMcp.call("query_nrdb", { query });

// Test with real
const realResult = await realMcp.call("query_nrdb", { query });

// Compare structure (not values)
assert.deepEqual(Object.keys(mockResult), Object.keys(realResult));
```

### 4. Documentation

Document when using mock data:
```javascript
/**
 * Analyzes transaction performance
 * @param {boolean} useMock - Use mock data (for testing)
 */
async function analyzePerformance(useMock = false) {
  const mcp = useMock ? mockMcp : realMcp;
  // ...
}
```

## Common Issues

### "Too Realistic"

The mock data can be too good:
- Users might not realize it's fake
- Dashboards look production-ready
- Analysis seems legitimate

Always clearly indicate mock mode.

### "Different Each Time"

Mock data varies to be realistic:
- Use averages for consistency
- Don't rely on specific values
- Test patterns, not exact numbers

### "Query Not Supported"

Complex queries return generic data:
- Simplify queries for testing
- Focus on response structure
- Don't test query logic in mock mode

## Summary

Mock mode in the MCP Server is:
- Sophisticated and realistic
- Perfect for development
- Free to use (no API costs)
- Sometimes too convincing

Use it for:
- Development and testing
- Demos and screenshots
- Learning the API
- Building integrations

Remember:
- Data is completely fake
- No persistence between calls
- Queries aren't fully parsed
- Always label mock data clearly
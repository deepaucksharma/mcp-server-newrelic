# Query Tools Documentation

This document details the query tools **as actually implemented** in the MCP Server.

## Overview

Query tools execute NRQL (New Relic Query Language) queries. Only one tool fully works with real data.

## Implementation Status

| Tool | Status | Real Functionality |
|------|--------|-------------------|
| `query_nrdb` | ✅ Working | Executes real NRQL queries |
| `query_check` | ⚠️ Partial | Basic syntax validation |
| `query_assist` | 🟨 Mock | Returns example queries |
| `query.execute_adaptive` | ❌ Broken | Handler missing/broken |
| `nrql.execute` | ✅ Working | Alias for query_nrdb |
| `query.build` | 🟨 Mock | Returns example queries |

## Primary Query Tool

### query_nrdb

**Purpose**: Execute NRQL queries against New Relic data.

**Implementation File**: `pkg/interface/mcp/tools_query.go`

**Parameters**:
```json
{
  "query": "SELECT count(*) FROM Transaction",  // string, required
  "account_id": "123456",                       // string, optional
  "timeout": 30                                 // integer, optional (seconds)
}
```

**Actual Implementation**:
```go
func (s *Server) handleQueryNRDB(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    query, ok := params["query"].(string)
    if !ok || query == "" {
        return nil, fmt.Errorf("query parameter is required")
    }
    
    // Check mock mode
    if s.isMockMode() {
        return s.getMockData("query_nrdb", params), nil
    }
    
    // Get timeout
    timeout := 30
    if t, ok := params["timeout"].(float64); ok {
        timeout = int(t)
    }
    
    // Execute via NerdGraph
    client := s.getNRClient()
    if client == nil {
        return nil, fmt.Errorf("New Relic client not configured")
    }
    
    // Direct GraphQL query - no optimization or validation
    result, err := client.QueryWithTimeout(ctx, query, timeout)
    if err != nil {
        return nil, fmt.Errorf("query execution failed: %w", err)
    }
    
    return result, nil
}
```

**What Works**:
- Executes any valid NRQL query
- Returns actual data from New Relic
- Respects timeout parameter
- Handles errors from NerdGraph

**What Doesn't Work**:
- No query optimization
- No schema validation
- No cost estimation
- Single account only (multi-account code exists but not wired)
- No query caching

**Example Usage**:

```json
// Simple count query
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
  }
}

// Response
{
  "data": {
    "results": [{
      "count": 142857
    }]
  },
  "metadata": {
    "timeWindow": {
      "begin": 1701234567000,
      "end": 1701238167000
    }
  }
}
```

```json
// Time series query
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT average(duration) FROM Transaction TIMESERIES 1 minute SINCE 1 hour ago"
  }
}

// Response
{
  "data": {
    "results": [{
      "average.duration": 0.123,
      "beginTimeSeconds": 1701234567
    }, {
      "average.duration": 0.145,
      "beginTimeSeconds": 1701234627
    }]
  }
}
```

```json
// Complex query with FACET
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT count(*), average(duration) FROM Transaction FACET appName SINCE 1 hour ago LIMIT 10",
    "timeout": 60
  }
}
```

### Common NRQL Patterns

**Aggregation Functions**:
- `count(*)` - Count events
- `average(metric)` - Average value
- `sum(metric)` - Sum values
- `min(metric)`, `max(metric)` - Min/max values
- `percentile(metric, 50, 95, 99)` - Percentiles
- `stddev(metric)` - Standard deviation

**Time Windows**:
- `SINCE 1 hour ago`
- `SINCE 7 days ago`
- `SINCE '2023-12-01 00:00:00'`
- `BETWEEN 2 days ago AND 1 day ago`

**Grouping**:
- `FACET attribute` - Group by attribute
- `FACET attribute LIMIT 20` - Top 20 groups
- `TIMESERIES 5 minutes` - Time buckets

## Partially Working Tools

### query_check

**Purpose**: Validate NRQL syntax before execution.

**Implementation File**: `pkg/interface/mcp/tools_query.go`

**What Works**:
- Basic syntax validation
- Checks for common NRQL keywords

**What Doesn't**:
- No schema validation
- Performance suggestions are mocked
- No cost estimation
- No security validation

**Example**:
```json
{
  "name": "query_check",
  "arguments": {
    "query": "SELECT count(*) FROM Transaction WHERE"  // Invalid
  }
}

// Response
{
  "valid": false,
  "errors": ["Query incomplete: WHERE clause requires condition"],
  "suggestions": ["Add a condition after WHERE"]  // Mocked
}
```

## Mock-Only Tools

### query_assist

**Purpose**: Help build NRQL queries from natural language.

**Reality**: Returns hardcoded example queries based on keywords.

**Example**:
```json
{
  "name": "query_assist",
  "arguments": {
    "description": "Show me slow transactions"
  }
}

// Mock Response
{
  "suggested_queries": [
    "SELECT average(duration) FROM Transaction WHERE duration > 1 SINCE 1 hour ago",
    "SELECT count(*) FROM Transaction WHERE duration > percentile(duration, 95) SINCE 1 hour ago"
  ],
  "explanation": "These queries find transactions slower than 1 second or in the 95th percentile"
}
```

### query.build

**Purpose**: Build NRQL programmatically.

**Reality**: Returns template queries, no real building logic.

## Non-Working Tools

### query.execute_adaptive

**Purpose**: Execute queries with automatic optimization.

**Reality**: Tool is registered but handler is broken or missing. Returns errors.

## Common Issues and Solutions

### Issue: Query Returns No Data

**Common Causes**:
1. Wrong time range
2. Incorrect event type name
3. Missing data in account

**Debug Query**:
```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SHOW EVENT TYPES"  // List all available event types
  }
}
```

### Issue: Query Timeout

**Solution**: Increase timeout and optimize query
```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT count(*) FROM Transaction SINCE 1 day ago",
    "timeout": 120  // 2 minutes
  }
}
```

### Issue: Complex Query Performance

**Solution**: Break into smaller queries
```javascript
// Instead of one complex query
const complex = "SELECT * FROM Transaction, TransactionError WHERE ...";

// Use separate queries
const transactions = await query("SELECT ... FROM Transaction");
const errors = await query("SELECT ... FROM TransactionError");
// Join in code
```

## Query Limitations

### API Limitations
- Max query execution time: ~2 minutes
- Max results: 2000 rows (without LIMIT)
- Max query length: 4096 characters
- Rate limits apply

### Implementation Limitations
- No query plan visibility
- No cost tracking
- No automatic optimization
- No schema awareness
- No cross-account queries (despite code)

## Best Practices

### 1. Always Specify Time Range
```sql
-- Good
SELECT count(*) FROM Transaction SINCE 1 hour ago

-- Bad (scans all data)
SELECT count(*) FROM Transaction
```

### 2. Use LIMIT for Large Results
```sql
-- Good
SELECT * FROM Transaction SINCE 1 hour ago LIMIT 100

-- Bad (may timeout)
SELECT * FROM Transaction SINCE 1 hour ago
```

### 3. Aggregate Before Faceting
```sql
-- Good (efficient)
SELECT average(duration) FROM Transaction FACET appName SINCE 1 hour ago

-- Bad (returns raw data)
SELECT duration FROM Transaction FACET appName SINCE 1 hour ago
```

### 4. Use Appropriate Time Buckets
```sql
-- For 1 hour window
TIMESERIES 1 minute

-- For 1 day window  
TIMESERIES 5 minutes

-- For 1 week window
TIMESERIES 1 hour
```

## Advanced NRQL Examples

### Percentile Analysis
```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT percentile(duration, 50, 75, 90, 95, 99) FROM Transaction SINCE 1 hour ago"
  }
}
```

### Error Rate Calculation
```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT percentage(count(*), WHERE error = true) FROM Transaction SINCE 1 hour ago"
  }
}
```

### Nested Aggregation
```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT average(duration) FROM Transaction FACET appName SINCE 1 hour ago LIMIT 10"
  }
}
```

### Compare Time Periods
```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT average(duration) FROM Transaction SINCE 1 hour ago COMPARE WITH 1 hour ago"
  }
}
```

## Development Notes

### Key Files
- `pkg/interface/mcp/tools_query.go` - Main query tool implementations
- `pkg/interface/mcp/tools_query_granular.go` - Advanced query tools (mostly mock)
- `pkg/newrelic/client.go` - GraphQL client for NerdGraph

### Why Limited Implementation?

1. **No Query Intelligence**: Planned features like optimization not built
2. **Direct GraphQL**: Queries pass directly to NerdGraph without processing
3. **No Abstraction Layer**: Missing query builder and optimizer
4. **Time Constraints**: Advanced features were designed but not implemented

### Extending Query Tools

To add query intelligence:

1. **Add Schema Validation**:
   ```go
   // Validate event types exist
   func validateEventTypes(query string) error {
     // Parse query
     // Check event types against discovered list
     // Return specific errors
   }
   ```

2. **Add Query Optimization**:
   ```go
   // Optimize before execution
   func optimizeQuery(query string) string {
     // Add LIMIT if missing
     // Optimize time ranges
     // Suggest indexes
     return optimizedQuery
   }
   ```

3. **Add Cost Estimation**:
   ```go
   // Estimate query cost
   func estimateQueryCost(query string) QueryCost {
     // Parse time range
     // Estimate data volume
     // Calculate approximate cost
   }
   ```

## Summary

Query tools in the MCP Server are minimal:
- Only `query_nrdb` fully works
- No query intelligence or optimization
- Direct pass-through to NerdGraph
- Advanced features are mocked or missing

For effective querying:
1. Learn NRQL syntax
2. Use `query_nrdb` exclusively  
3. Handle optimization manually
4. Don't expect intelligent assistance
5. Check New Relic docs for NRQL features
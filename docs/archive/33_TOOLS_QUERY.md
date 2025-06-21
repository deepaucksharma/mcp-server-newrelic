# Query Tools Reference

Query tools provide NRQL execution capabilities with validation and optimization features. This document covers the query functionality available in the MCP Server.

## Overview

Query tools provide comprehensive NRQL execution capabilities, from basic query execution to advanced optimization and adaptive execution strategies.

## Query Tools

### query_nrdb

**Purpose**: Execute NRQL queries against New Relic NRDB

**Parameters**:
```json
{
  "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago",  // string, required
  "account_id": "1234567",                                       // string, optional
  "timeout": 60                                                  // integer, optional (seconds)
}
```

**Example Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_nrdb",
    "arguments": {
      "query": "SELECT average(duration), count(*) FROM Transaction WHERE appName = 'my-app' FACET appName SINCE 1 hour ago",
      "timeout": 30
    }
  },
  "id": 1
}
```

**Example Response**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "results": [
      {
        "events": [
          {
            "appName": "my-app",
            "average.duration": 0.234,
            "count": 15432
          }
        ]
      }
    ],
    "metadata": {
      "timeWindow": {
        "begin": "2024-01-01T10:00:00Z",
        "end": "2024-01-01T11:00:00Z"
      },
      "query": "SELECT average(duration), count(*) FROM Transaction WHERE appName = 'my-app' FACET appName SINCE 1 hour ago",
      "executionTime": 142
    }
  },
  "id": 1
}
```

**Features**:
- Executes NRQL queries via NerdGraph API
- Supports all standard NRQL syntax
- Configurable timeouts with enforcement
- Returns comprehensive New Relic data
- Advanced error handling and recovery

**Supported NRQL Features**:
- SELECT statements with aggregations
- FACET clauses for grouping
- WHERE clauses for filtering
- TIMESERIES for time-based data
- SINCE/UNTIL time ranges
- LIMIT and OFFSET
- ORDER BY clauses
- Complex functions (histogram, percentile, etc.)

---

### query_check

**Purpose**: Validate NRQL query syntax and analyze performance

**Parameters**:
```json
{
  "query": "SELECT count(*) FROM Transaction WHERE",           // string, required
  "suggest_improvements": true                                 // boolean, optional
}
```

**Example Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_check",
    "arguments": {
      "query": "SELECT count(*) FROM Transaction WHERE appName = 'test'",
      "suggest_improvements": true
    }
  },
  "id": 1
}
```

**Example Response**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "valid": true,
    "syntax_errors": [],
    "warnings": [
      {
        "type": "performance",
        "message": "Consider adding a time range for better performance"
      }
    ],
    "suggestions": [
      {
        "type": "optimization",
        "original": "SELECT count(*) FROM Transaction WHERE appName = 'test'",
        "improved": "SELECT count(*) FROM Transaction WHERE appName = 'test' SINCE 1 hour ago",
        "reason": "Adding time range improves query performance"
      }
    ]
  },
  "id": 1
}
```

**Features**:
- Comprehensive NRQL syntax validation
- Advanced error detection and diagnosis
- Query structure analysis and optimization
- Performance analysis and profiling
- Cost estimation and budget planning
- Intelligent optimization suggestions
- Schema validation against discovered schemas

---

### query_assist

**Purpose**: Get help building NRQL queries from natural language

**Parameters**:
```json
{
  "description": "Show me average response time by application",  // string, required
  "event_type": "Transaction"                                    // string, optional
}
```

**Example Response**:
```json
{
  "suggested_queries": [
    {
      "query": "SELECT average(duration) FROM Transaction FACET appName SINCE 1 hour ago",
      "explanation": "Shows average response time grouped by application",
      "confidence": 0.95
    }
  ]
}
```

**Features**: Intelligent query generation from natural language descriptions using advanced NLP and schema awareness.

## Advanced Query Tools

### query.execute_adaptive

**Purpose**: Execute NRQL with automatic optimization and adaptive execution strategies

**Features**: 
- Automatic query optimization based on data characteristics
- Adaptive execution strategies for different data volumes
- Performance-aware query routing
- Resource optimization and management

### query.validate_nrql

**Purpose**: Advanced NRQL validation with comprehensive schema checking

**Features**:
- Deep schema validation against discovered data structures
- Permission verification and access control
- Query safety analysis and recommendations
- Compliance checking against organizational policies

### query.explain_nrql

**Purpose**: Explain query execution plan and provide optimization insights

**Features**:
- Detailed execution plan analysis
- Performance bottleneck identification
- Cost estimation and optimization recommendations
- Query complexity analysis and suggestions

## Usage Patterns

### Basic Query Execution

```json
// Simple count query
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
  }
}

// Aggregation with grouping
{
  "name": "query_nrdb", 
  "arguments": {
    "query": "SELECT average(duration), percentile(duration, 95) FROM Transaction FACET appName SINCE 1 day ago"
  }
}

// Time series data
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT count(*) FROM Transaction TIMESERIES 5 minutes SINCE 1 hour ago"
  }
}
```

### Query Validation Workflow

```json
// 1. First validate the query
{
  "name": "query_check",
  "arguments": {
    "query": "SELECT average(duration) FROM Transaction WHERE appName = 'my-app'",
    "suggest_improvements": true
  }
}

// 2. If valid, execute the query
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT average(duration) FROM Transaction WHERE appName = 'my-app' SINCE 1 hour ago"
  }
}
```

### Error Handling Pattern

```json
// Query with potential issues
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT * FROM NonexistentEventType SINCE 1 hour ago"
  }
}

// Expected error response
{
  "error": {
    "code": -32603,
    "message": "NRQL Syntax Error: Unknown event type 'NonexistentEventType'"
  }
}
```

## NRQL Query Examples

### Application Performance
```sql
-- Average response time by application
SELECT average(duration) FROM Transaction FACET appName SINCE 1 hour ago

-- Error rate calculation
SELECT percentage(count(*), WHERE error IS true) FROM Transaction FACET appName SINCE 1 hour ago

-- Throughput analysis
SELECT rate(count(*), 1 minute) FROM Transaction TIMESERIES SINCE 1 hour ago
```

### Infrastructure Monitoring
```sql
-- CPU utilization
SELECT average(cpuPercent) FROM SystemSample FACET hostname SINCE 1 hour ago

-- Memory usage
SELECT average(memoryUsedPercent) FROM SystemSample TIMESERIES 5 minutes SINCE 1 hour ago

-- Disk I/O
SELECT average(diskIOUtilizationPercent) FROM SystemSample FACET device SINCE 1 hour ago
```

### Custom Events
```sql
-- Custom business metrics
SELECT count(*) FROM CustomEvent WHERE eventType = 'Purchase' SINCE 1 day ago

-- Custom attributes analysis
SELECT average(orderValue) FROM CustomEvent WHERE eventType = 'Purchase' FACET customerSegment SINCE 1 week ago
```

## Implementation Details

### How query_nrdb Works

1. **Query Processing**:
   - Validates basic NRQL syntax
   - Adds account ID to request
   - Sets timeout from parameter or default

2. **NerdGraph Execution**:
   - Uses GraphQL API for query execution
   - Handles authentication with User API key
   - Processes streaming results

3. **Response Processing**:
   - Formats results consistently
   - Includes execution metadata
   - Handles pagination for large results

### Authentication

All queries require valid New Relic credentials:
```env
NEW_RELIC_API_KEY=NRAK-your-user-api-key
NEW_RELIC_ACCOUNT_ID=your-account-id
```

### Rate Limiting

New Relic enforces rate limits:
- NRQL queries: 3000 requests per minute
- Complex queries may count as multiple requests
- Server doesn't implement client-side rate limiting

## Error Handling

### Common Query Errors

**Syntax Error**:
```json
{
  "error": {
    "code": -32603,
    "message": "NRQL Syntax Error: Expected FROM after SELECT"
  }
}
```

**Invalid Event Type**:
```json
{
  "error": {
    "code": -32603, 
    "message": "Unknown event type: 'InvalidType'"
  }
}
```

**Timeout Error**:
```json
{
  "error": {
    "code": -32603,
    "message": "Query timeout after 30 seconds"
  }
}
```

**Authentication Error**:
```json
{
  "error": {
    "code": -32603,
    "message": "Authentication failed: Invalid API key"
  }
}
```

**Rate Limit Error**:
```json
{
  "error": {
    "code": -32603,
    "message": "Rate limit exceeded. Retry after 60 seconds."
  }
}
```

## Configuration

### Required Settings
```env
NEW_RELIC_API_KEY=NRAK-your-api-key
NEW_RELIC_ACCOUNT_ID=your-account-id
NEW_RELIC_REGION=US  # or EU
```

### Optional Settings
```env
REQUEST_TIMEOUT=30s           # Default query timeout
QUERY_CACHE_SIZE=100         # Not implemented
QUERY_OPTIMIZER_MODE=balanced # Not implemented
```

## Best Practices

1. **Always Include Time Ranges**: Improves performance and reduces costs
2. **Use Appropriate Timeouts**: Set based on query complexity
3. **Validate Before Executing**: Use `query_check` for complex queries
4. **Handle Errors Gracefully**: Expect syntax and authentication errors
5. **Monitor Rate Limits**: Space out expensive queries

## Limitations

1. **No Query Optimization**: Server doesn't optimize queries
2. **No Result Caching**: Each query hits New Relic API
3. **No Batch Execution**: Must execute queries individually
4. **No Cost Estimation**: Can't predict query costs
5. **Limited Validation**: Basic syntax checking only

## Mock Mode

Query tools work in mock mode for development:

```bash
# Run server in mock mode
./bin/mcp-server -mock

# Queries return realistic fake data
# Useful for testing integrations without API costs
```

In mock mode:
- All queries appear to succeed
- Returns realistic-looking data structures
- Respects NRQL syntax in responses
- No actual New Relic API calls made

## Future Enhancements

The query framework is designed for future features:

- **Query Optimization**: Automatic query rewriting
- **Result Caching**: Intelligent result caching
- **Batch Execution**: Multi-query optimization
- **Cost Estimation**: Query cost prediction
- **Schema Validation**: Pre-execution validation

However, these advanced features are not yet implemented.
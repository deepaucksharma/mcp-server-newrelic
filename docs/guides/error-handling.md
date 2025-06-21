# Error Handling Guide

The New Relic MCP Server provides comprehensive error handling with structured error types, helpful hints, and retry guidance.

## Error Structure

All errors follow a consistent structure:

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32602,
    "message": "Validation failed for field 'query': NRQL syntax error",
    "data": {
      "type": "validation_error",
      "details": {
        "field": "query",
        "message": "NRQL syntax error"
      },
      "hint": "Check NRQL syntax: https://docs.newrelic.com/docs/query-your-data/nrql-reference/",
      "tool": "query_nrdb"
    }
  },
  "id": 1
}
```

## Error Types

### Protocol Errors

| Type | Code | Description | Retryable |
|------|------|-------------|-----------|
| parse_error | -32700 | Invalid JSON | No |
| invalid_request | -32600 | Invalid request structure | No |
| method_not_found | -32601 | Unknown method or tool | No |
| invalid_params | -32602 | Invalid parameters | No |
| internal_error | -32603 | Internal server error | Sometimes |

### Execution Errors

| Type | Code | Description | Retryable |
|------|------|-------------|-----------|
| timeout | -32603 | Request timed out | Yes |
| rate_limit | -32001 | Rate limit exceeded | Yes (with delay) |
| unauthorized | -32002 | Authentication failed | No |
| permission_error | -32603 | Insufficient permissions | No |

### Data Errors

| Type | Code | Description | Retryable |
|------|------|-------------|-----------|
| query_error | -32603 | NRQL query failed | Sometimes |
| data_not_found | -32003 | Resource not found | No |
| validation_error | -32602 | Validation failed | No |
| account_error | -32002 | Account access error | No |

## Error Examples

### Invalid Parameters
```json
{
  "method": "tools/call",
  "params": {
    "name": "query_nrdb"
    // Missing required 'arguments'
  }
}

// Response:
{
  "error": {
    "code": -32602,
    "message": "Tool name is required",
    "data": {
      "type": "invalid_params",
      "details": {
        "parameter": "name"
      }
    }
  }
}
```

### Tool Not Found
```json
{
  "method": "tools/call",
  "params": {
    "name": "unknown_tool",
    "arguments": {}
  }
}

// Response:
{
  "error": {
    "code": -32601,
    "message": "Method 'unknown_tool' not found",
    "data": {
      "type": "method_not_found",
      "hint": "Did you mean 'query_nrdb'?",
      "details": {
        "method": "unknown_tool",
        "suggestion": "query_nrdb"
      }
    }
  }
}
```

### Query Timeout
```json
{
  "error": {
    "code": -32603,
    "message": "Tool execution timed out after 30s",
    "data": {
      "type": "timeout",
      "tool": "query_nrdb",
      "hint": "Try reducing the query complexity or time range"
    }
  }
}
```

### Rate Limit
```json
{
  "error": {
    "code": -32001,
    "message": "Rate limit exceeded",
    "data": {
      "type": "rate_limit",
      "details": {
        "limit": 60,
        "window": "1m0s",
        "retry_after": 60
      },
      "hint": "Maximum 60 requests per 1m0s"
    }
  }
}
```

## Error Handling Best Practices

### 1. Check Error Type
```javascript
if (response.error) {
  const errorType = response.error.data?.type;
  
  switch (errorType) {
    case 'timeout':
      // Retry with smaller time range
      break;
    case 'rate_limit':
      // Wait and retry
      const retryAfter = response.error.data.details.retry_after;
      await sleep(retryAfter * 1000);
      break;
    case 'validation_error':
      // Fix parameters and retry
      console.log('Hint:', response.error.data.hint);
      break;
  }
}
```

### 2. Use Hints
Errors include helpful hints when available:
```json
{
  "hint": "Try reducing the time range or query complexity"
}
```

### 3. Implement Retry Logic
```javascript
async function executeWithRetry(request, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    const response = await sendRequest(request);
    
    if (!response.error) {
      return response;
    }
    
    const error = response.error;
    const errorType = error.data?.type;
    
    // Check if retryable
    if (['timeout', 'rate_limit', 'internal_error'].includes(errorType)) {
      const delay = error.data?.details?.retry_after || Math.pow(2, i);
      await sleep(delay * 1000);
      continue;
    }
    
    // Non-retryable error
    throw error;
  }
}
```

### 4. Log Structured Errors
```javascript
function logError(error) {
  console.error({
    timestamp: new Date().toISOString(),
    error_type: error.data?.type,
    error_code: error.code,
    message: error.message,
    tool: error.data?.tool,
    hint: error.data?.hint,
    request_id: error.data?.request_id
  });
}
```

## Common Error Scenarios

### Scenario 1: Invalid NRQL Syntax
```json
// Request
{
  "tool": "query_nrdb",
  "params": {
    "query": "SELECT * FORM Transaction"  // Typo: FORM instead of FROM
  }
}

// Error
{
  "type": "query_error",
  "message": "Query execution failed",
  "hint": "Check NRQL syntax: https://docs.newrelic.com/docs/query-your-data/nrql-reference/"
}
```

### Scenario 2: Account Access
```json
// Request
{
  "tool": "query_nrdb",
  "params": {
    "query": "SELECT count(*) FROM Transaction",
    "account_id": "999999"  // Invalid account
  }
}

// Error
{
  "type": "account_error",
  "message": "Cannot query data for account 999999",
  "hint": "Verify the account ID exists and you have access"
}
```

### Scenario 3: Missing Required Field
```json
// Request
{
  "tool": "create_alert",
  "params": {
    "name": "My Alert"
    // Missing required 'query' field
  }
}

// Error
{
  "type": "validation_error",
  "message": "Validation failed for field 'query': query parameter is required",
  "details": {
    "field": "query",
    "message": "query parameter is required"
  }
}
```

## Error Recovery Strategies

### 1. Graceful Degradation
```javascript
try {
  // Try with full features
  const result = await queryWithAdvancedFeatures();
  return result;
} catch (error) {
  if (error.type === 'timeout' || error.type === 'query_error') {
    // Fall back to simpler query
    return await queryWithBasicFeatures();
  }
  throw error;
}
```

### 2. Circuit Breaker Pattern
```javascript
class CircuitBreaker {
  constructor(threshold = 5, timeout = 60000) {
    this.failures = 0;
    this.threshold = threshold;
    this.timeout = timeout;
    this.nextAttempt = Date.now();
  }
  
  async execute(fn) {
    if (Date.now() < this.nextAttempt) {
      throw new Error('Circuit breaker is open');
    }
    
    try {
      const result = await fn();
      this.failures = 0;
      return result;
    } catch (error) {
      this.failures++;
      if (this.failures >= this.threshold) {
        this.nextAttempt = Date.now() + this.timeout;
      }
      throw error;
    }
  }
}
```

### 3. Adaptive Timeout
```javascript
async function queryWithAdaptiveTimeout(query, initialTimeout = 10) {
  let timeout = initialTimeout;
  
  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      return await executeQuery(query, { timeout });
    } catch (error) {
      if (error.type === 'timeout') {
        timeout *= 2; // Double timeout for next attempt
        continue;
      }
      throw error;
    }
  }
}
```

## Monitoring Errors

### Error Metrics
Track these metrics for monitoring:
- Error rate by type
- Error rate by tool
- Retry success rate
- Average retry count
- Error resolution time

### Error Patterns
Watch for patterns indicating issues:
- Spike in timeout errors → Performance issue
- Increase in permission errors → Configuration issue
- Rate limit errors → Need to adjust limits
- Query errors → User education needed

## Testing Error Handling

### Unit Tests
```go
func TestErrorHandling(t *testing.T) {
    // Test invalid parameters
    _, err := handler.HandleTool(ctx, "query_nrdb", map[string]interface{}{})
    assert.Error(t, err)
    
    mcpErr, ok := err.(*MCPError)
    assert.True(t, ok)
    assert.Equal(t, ErrorTypeInvalidParams, mcpErr.Type)
    assert.Contains(t, mcpErr.Message, "required")
}
```

### Integration Tests
```bash
# Test timeout handling
echo '{"method":"tools/call","params":{"name":"query_nrdb","arguments":{"query":"SELECT count(*) FROM Transaction SINCE 1 year ago","timeout":1}}}' | ./bin/mcp-server

# Test rate limiting
for i in {1..100}; do
  echo '{"method":"tools/call","params":{"name":"query_nrdb","arguments":{"query":"SELECT 1"}}}' | ./bin/mcp-server
done
```

## Future Enhancements

1. **Error Telemetry**: Send error metrics to New Relic
2. **Smart Retry**: ML-based retry strategies
3. **Error Suggestions**: AI-powered error resolution hints
4. **Error Replay**: Ability to retry failed requests with modifications

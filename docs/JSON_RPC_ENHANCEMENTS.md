# JSON-RPC 2.0 Enhanced Protocol Implementation

This document describes the enhanced JSON-RPC 2.0 protocol implementation for the New Relic MCP Server, providing full compliance with the JSON-RPC 2.0 specification and additional MCP-specific features.

## Overview

The enhanced protocol implementation (`protocol_enhanced.go`) provides:

1. **Full JSON-RPC 2.0 Compliance**
   - Proper error codes and response structures
   - Batch request support
   - Notification handling (requests without ID)
   - Parameter validation against schemas

2. **MCP-Specific Enhancements**
   - Extended error codes for better diagnostics
   - Tool parameter validation
   - Request timeout handling
   - Rate limiting framework
   - Metrics and caching hooks

## Error Codes

### Standard JSON-RPC 2.0 Error Codes
```go
ParseErrorCode      = -32700  // Invalid JSON was received
InvalidRequestCode  = -32600  // The JSON sent is not a valid Request object
MethodNotFoundCode  = -32601  // The method does not exist / is not available
InvalidParamsCode   = -32602  // Invalid method parameter(s)
InternalErrorCode   = -32603  // Internal JSON-RPC error
```

### MCP-Specific Error Codes
```go
ToolNotFoundCode    = -32001  // The requested tool doesn't exist
ToolExecutionCode   = -32002  // Tool execution failed
SessionNotFoundCode = -32003  // Session ID not found
TimeoutErrorCode    = -32004  // Request timed out
RateLimitCode       = -32005  // Rate limit exceeded
```

## Enhanced Error Responses

Error responses include additional context to help diagnose issues:

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32602,
    "message": "Invalid parameters",
    "data": {
      "detail": "required parameter 'query' is missing",
      "schema": {
        "type": "object",
        "required": ["query"],
        "properties": {
          "query": {
            "type": "string",
            "description": "NRQL query to execute"
          }
        }
      },
      "received": {
        "timeout": 30
      }
    }
  },
  "id": 1
}
```

## Batch Request Support

The enhanced protocol fully supports batch requests:

```json
[
  {"jsonrpc": "2.0", "method": "tools/list", "id": 1},
  {"jsonrpc": "2.0", "method": "query_nrdb", "params": {"query": "SELECT count(*) FROM Transaction"}, "id": 2},
  {"jsonrpc": "2.0", "method": "tools/changed"}  // Notification (no response)
]
```

Responses maintain order for requests with IDs:

```json
[
  {"jsonrpc": "2.0", "result": {"tools": [...]}, "id": 1},
  {"jsonrpc": "2.0", "result": {"data": [...]}, "id": 2}
]
```

## Notification Support

Requests without an `id` field are treated as notifications and don't receive responses:

```json
{
  "jsonrpc": "2.0",
  "method": "tools/changed"
}
```

Supported notification methods:
- `tools/changed` - Tool registry updated
- `sessions/ended` - Session terminated
- `cancel` - Cancel in-flight request

## Parameter Validation

The enhanced protocol validates tool parameters against their schemas:

1. **Required Parameters**: Ensures all required parameters are present
2. **Type Validation**: Validates parameter types (string, number, boolean, array, object)
3. **Enum Validation**: Checks enum values if defined
4. **Schema Reporting**: Returns the expected schema on validation errors

## Request Features

### Timeout Handling
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32004,
    "message": "Request timeout",
    "data": {
      "timeout": "30s",
      "method": "query_nrdb"
    }
  },
  "id": 1
}
```

### Rate Limiting
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32005,
    "message": "Rate limit exceeded",
    "data": {
      "retryAfter": 60,
      "limit": 100
    }
  },
  "id": 1
}
```

### Tool Suggestions
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Tool 'query_nrql' not found",
    "data": {
      "availableTools": ["query_nrdb", "query_check", "query_builder"],
      "suggestion": "query_nrdb"
    }
  },
  "id": 1
}
```

## Usage

To enable the enhanced protocol:

```go
config := ServerConfig{
    EnhancedProtocol: true,
    RequestTimeout:   30 * time.Second,
    RateLimit:        100,
}
server := NewServer(config)
```

## Implementation Details

The enhanced protocol handler (`protocol_enhanced.go`) provides:

1. **EnhancedHandleMessage**: Main entry point with full JSON validation
2. **processEnhancedRequest**: Single request processing with timeout
3. **handleEnhancedBatch**: Concurrent batch processing with ordering
4. **validateToolParams**: Schema-based parameter validation
5. **enhancedErrorResponse**: Rich error response generation

## Benefits

1. **Better Error Diagnostics**: Detailed error messages with hints and context
2. **Improved Reliability**: Timeout handling and rate limiting
3. **Full Compliance**: Adheres to JSON-RPC 2.0 specification
4. **Enhanced Developer Experience**: Clear error messages and parameter validation
5. **Performance**: Concurrent batch processing with configurable limits

## Migration

The enhanced protocol is backward compatible. Existing clients continue to work, while new clients can benefit from enhanced features by ensuring the server is configured with `EnhancedProtocol: true`.

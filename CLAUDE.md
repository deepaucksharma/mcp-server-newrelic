# CLAUDE.md

This file provides comprehensive guidance to Claude Code (claude.ai/code) and other AI assistants when working with the New Relic MCP Server repository.

## Project Overview

The New Relic MCP Server is a production-grade Go implementation of the Model Context Protocol that provides AI assistants with intelligent access to New Relic observability data. It enables sophisticated operations including NRQL queries, dashboard generation, alert management, and bulk operations.

**Current State**: Feature-complete Go implementation with all core tools. Focus is now on testing, error handling, and CI/CD setup. The Python implementation in `new-branch` should be considered deprecated.

## Critical Context

### Dual Implementation Issue
The repository contains two parallel implementations:
1. **Go Implementation** (main) - Complete and recommended
2. **Python Implementation** (new-branch) - Partial, to be deprecated

**IMPORTANT**: Always work with the Go implementation unless specifically asked about Python code.

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│          AI Assistant (Claude/Copilot)          │
└────────────────────┬────────────────────────────┘
                     │ MCP Protocol (JSON-RPC)
┌────────────────────▼────────────────────────────┐
│              Go MCP Server                       │
├─────────────────────────────────────────────────┤
│  pkg/interface/mcp/                             │
│  ├─ server.go       - Core MCP server          │
│  ├─ tools_query.go  - NRQL query tools         │
│  ├─ tools_dashboard.go - Dashboard tools       │
│  ├─ tools_alerts.go - Alert management         │
│  └─ tools_discovery.go - Schema discovery      │
├─────────────────────────────────────────────────┤
│  Core Components:                               │
│  ├─ pkg/discovery/  - Schema analysis engine   │
│  ├─ pkg/state/      - Session & cache mgmt     │
│  ├─ pkg/newrelic/   - NerdGraph client         │
│  └─ pkg/config/     - Configuration mgmt       │
└────────────────────┬────────────────────────────┘
                     │ GraphQL/HTTPS
                     ▼
              New Relic NerdGraph API
```

## Development Workflow

### Initial Setup
```bash
# 1. Clone and setup environment
git clone <repo>
cd mcp-server-newrelic
cp .env.example .env
# Edit .env with credentials

# 2. Run diagnostics
make diagnose         # Check environment
make diagnose-fix     # Auto-fix issues

# 3. Build everything
make build
```

### Daily Development
```bash
# Run in different modes
make run              # Production mode
make run-mock         # Mock mode (no NR connection)
make dev              # Development with auto-reload

# Testing workflow
make test             # Run all tests
make test-unit        # Just unit tests
make test-coverage    # Coverage report

# Code quality
make lint             # Run linters
make format           # Auto-format code
```

## Code Organization

### Package Structure
```
pkg/
├── interface/mcp/    # MCP protocol implementation
│   ├── server.go     # Core server & tool registry
│   ├── tools_*.go    # Tool implementations
│   └── types.go      # MCP type definitions
├── discovery/        # Data discovery engine
│   ├── engine.go     # Main discovery logic
│   ├── interfaces.go # Contract definitions
│   └── nrdb/         # New Relic DB client
├── state/            # State management
│   ├── manager.go    # Session manager
│   ├── cache.go      # Caching layer
│   └── factory.go    # State factory
├── newrelic/         # New Relic API client
│   └── client.go     # GraphQL client
└── config/           # Configuration
    └── config.go     # Environment config
```

### Tool Implementation Pattern

When implementing a new tool, follow this pattern:

```go
// 1. In tools_category.go, define the tool registration
func (s *Server) registerCategoryTools() error {
    s.tools.Register(Tool{
        Name:        "tool_name",
        Description: "Clear description of what this tool does",
        Parameters: ToolParameters{
            Type:     "object",
            Required: []string{"required_param"},
            Properties: map[string]Property{
                "required_param": {
                    Type:        "string",
                    Description: "What this parameter controls",
                },
                "optional_param": {
                    Type:        "integer",
                    Description: "Optional with default",
                    Default:     10,
                },
            },
        },
        Handler: s.handleToolName,
    })
    return nil
}

// 2. Implement the handler
func (s *Server) handleToolName(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Always validate required parameters first
    requiredParam, ok := params["required_param"].(string)
    if !ok || requiredParam == "" {
        return nil, fmt.Errorf("required_param is required and must be non-empty")
    }
    
    // Handle optional parameters with defaults
    optionalParam := 10
    if val, ok := params["optional_param"].(float64); ok {
        optionalParam = int(val)
    }
    
    // Check for mock mode
    if s.nrClient == nil {
        return generateMockResponse(requiredParam, optionalParam), nil
    }
    
    // Execute actual operation
    result, err := s.nrClient.PerformOperation(ctx, requiredParam, optionalParam)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }
    
    return formatResponse(result), nil
}

// 3. Add tests in tools_category_test.go
func TestHandleToolName(t *testing.T) {
    tests := []struct {
        name    string
        params  map[string]interface{}
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid parameters",
            params: map[string]interface{}{
                "required_param": "value",
                "optional_param": 20,
            },
            wantErr: false,
        },
        {
            name:    "missing required parameter",
            params:  map[string]interface{}{},
            wantErr: true,
            errMsg:  "required_param is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := &Server{} // Mock mode
            _, err := s.handleToolName(context.Background(), tt.params)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Current Implementation Details

### ✅ Fully Implemented Tools

**Query Tools** (`tools_query.go`):
- `query_nrdb` - Execute NRQL with timeout control
- `query_check` - Validate syntax and estimate cost
- `query_builder` - Build NRQL from structured params

**Discovery Tools** (`tools_discovery.go`):
- `discovery.list_schemas` - List all schemas with quality
- `discovery.profile_attribute` - Deep attribute analysis
- `discovery.find_relationships` - Relationship mining
- `discovery.assess_quality` - Quality assessment

**Dashboard Tools** (`tools_dashboard.go`):
- `find_usage` - Find dashboards by metric usage
- `generate_dashboard` - Create from templates
- `list_dashboards` - List with filtering
- `get_dashboard` - Get full details

**Alert Tools** (`tools_alerts.go`):
- `create_alert` - Smart alert creation
- `list_alerts` - List with incidents
- `analyze_alerts` - Effectiveness analysis
- `bulk_update_alerts` - Bulk operations

### 🚧 Areas Needing Work

1. **Test Coverage** (~40% currently)
   - Need unit tests for all tools
   - Integration tests for MCP protocol
   - Mock client implementations

2. **Error Handling**
   - Network timeout handling
   - Better error messages for users
   - Retry logic for transient failures

3. **Performance**
   - Implement caching layer
   - Optimize large result sets
   - Add request timeouts

4. **CI/CD**
   - GitHub Actions workflow
   - Automated testing
   - Docker image building

## Testing Strategy

### Unit Tests
```go
// Always test:
// 1. Valid inputs
// 2. Invalid inputs
// 3. Missing required params
// 4. Mock mode behavior
// 5. Error conditions
```

### Integration Tests
```go
// Test full MCP request/response cycle
// Use test fixtures for NerdGraph responses
// Verify JSON-RPC compliance
```

### Manual Testing
```bash
# Test with MCP inspector
npx @modelcontextprotocol/inspector ./bin/mcp-server

# Test specific tool
echo '{"jsonrpc":"2.0","method":"tools/query_nrdb","params":{"query":"SELECT count(*) FROM Transaction"},"id":1}' | ./bin/mcp-server
```

## Common Patterns

### Input Validation
```go
// Always validate types and required fields
param, ok := params["field"].(string)
if !ok || param == "" {
    return nil, fmt.Errorf("field is required and must be a string")
}
```

### Mock Mode Support
```go
// Every tool must work without New Relic
if s.nrClient == nil {
    return mockData, nil
}
```

### Error Wrapping
```go
// Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to query NRDB: %w", err)
}
```

### Resource Cleanup
```go
// Use defer for cleanup
defer func() {
    if err := resource.Close(); err != nil {
        log.Printf("failed to close resource: %v", err)
    }
}()
```

## Debugging Tips

1. **Enable debug logging**:
   ```bash
   export LOG_LEVEL=DEBUG
   make run
   ```

2. **Use mock mode** for development:
   ```bash
   make run-mock
   ```

3. **Check diagnostics** for environment issues:
   ```bash
   make diagnose
   ```

4. **Inspect MCP requests**:
   ```bash
   # Log all JSON-RPC requests
   export MCP_DEBUG=true
   ```

5. **Test individual tools**:
   ```go
   go test -v -run TestSpecificTool ./pkg/interface/mcp/
   ```

## Performance Considerations

1. **Avoid N+1 queries** - Batch operations when possible
2. **Use pagination** - Don't load unlimited results
3. **Cache expensive operations** - Use state manager
4. **Set reasonable timeouts** - Default 30s for queries
5. **Stream large results** - Use streaming for profile operations

## Security Guidelines

1. **Never log sensitive data** - No API keys, query results
2. **Validate all inputs** - Prevent injection attacks
3. **Use least privilege** - Request minimal NR permissions
4. **Sanitize errors** - Don't leak internal details
5. **Require auth config** - No default credentials

## Known Issues

1. **Redis connection failures** - Falls back to in-memory state
2. **Large result sets** - May timeout or OOM
3. **EU region** - Not yet supported
4. **Pagination** - Not fully implemented

## Future Enhancements

Based on the roadmap:
1. Phase 1: Test infrastructure and logging
2. Phase 2: Complete missing features
3. Phase 3: Copilot CLI optimization
4. Phase 4: Performance and resilience
5. Phase 5: Production readiness

## Getting Help

- Check tool implementations in `pkg/interface/mcp/tools_*.go`
- Review tests for usage examples
- Run `make diagnose` for environment issues
- See `cmd/diagnose/main.go` for validation patterns
- Check TODOs in code for planned work

## Important Files

- `cmd/mcp-server/main.go` - Entry point
- `pkg/interface/mcp/server.go` - Core MCP server
- `pkg/config/config.go` - Configuration
- `Makefile` - All build commands
- `.env.example` - Required environment vars

Remember: This is a production-grade server. Quality, testing, and security are paramount.
# Development Guide

This guide helps contributors understand how to work with the MCP Server for New Relic codebase in its current state.

## Current State of Development

**Warning:** This codebase has significant technical debt:
- Only ~15% of documented features are implemented
- Heavy reliance on mock data
- Circular build dependencies  
- Many aspirational interfaces without implementations

## Setting Up Development Environment

### Prerequisites

- Go 1.21 or later
- Make
- Git
- Optional: New Relic account for testing real queries

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env

# Edit .env with your credentials (or use mock mode)
```

### Building

```bash
# Build everything (may show errors - this is expected)
make build

# Build just the MCP server
make build-mcp

# The binary will be at ./bin/mcp-server
```

**Known Issues:**
- `make test` will fail due to missing test.sh
- Some packages may not compile due to build tags
- Circular dependencies between test and non-test builds

## Understanding the Codebase

### Key Directories

```
pkg/interface/mcp/      # Main MCP implementation
  tools_*.go           # Tool implementations (mostly mock)
  server_*.go          # Server variants (with/without discovery)
  transport_*.go       # Transport implementations (working)
  
pkg/discovery/         # Discovery engine (interface only)
pkg/newrelic/         # Basic New Relic client
pkg/state/            # Session management

cmd/mcp-server/       # Main entry point
cmd/diagnose/         # Diagnostics tool
```

### Build Tags Problem

The codebase uses problematic build tags:
- `//go:build !test` - Main implementations
- `//go:build test` - Test implementations
- `//go:build !nodiscovery` - With discovery
- `//go:build nodiscovery` - Without discovery

This creates circular dependencies and makes testing difficult.

## Adding a New Tool

### Step 1: Choose the Right File

Tools are organized by category:
- `tools_query.go` - NRQL query tools
- `tools_discovery.go` - Discovery tools  
- `tools_analysis.go` - Analysis tools
- `tools_dashboard.go` - Dashboard tools
- etc.

### Step 2: Register Your Tool

```go
func (s *Server) registerYourCategoryTools() error {
    s.tools.Register(Tool{
        Name:        "your_tool_name",
        Description: "What this tool does",
        Parameters: ToolParameters{
            Type:     "object",
            Required: []string{"required_param"},
            Properties: map[string]Property{
                "required_param": {
                    Type:        "string",
                    Description: "Description of parameter",
                },
            },
        },
        Handler: s.handleYourTool,
    })
    return nil
}
```

### Step 3: Implement the Handler

```go
func (s *Server) handleYourTool(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // 1. Validate parameters
    requiredParam, ok := params["required_param"].(string)
    if !ok || requiredParam == "" {
        return nil, fmt.Errorf("required_param is required")
    }
    
    // 2. Check mock mode (important!)
    if s.isMockMode() {
        return s.getMockData("your_tool_name", params), nil
    }
    
    // 3. Implement real functionality
    // Note: Most tools just return mock data currently
    
    return result, nil
}
```

### Step 4: Add Mock Data

In `mock_generator.go`, add mock responses:

```go
case "your_tool_name":
    return map[string]interface{}{
        "status": "success",
        "data": generateRealisticData(),
    }
```

## Testing

### Running Tests (Broken)

```bash
# This will fail - test.sh is missing
make test

# Run Go tests directly (may fail due to build tags)
go test ./...

# Test specific package
go test -v ./pkg/interface/mcp/
```

### Manual Testing

Best approach for testing:

```bash
# 1. Run in mock mode
./bin/mcp-server -mock

# 2. Send test requests via stdin
cat << EOF | ./bin/mcp-server -mock
{
  "jsonrpc": "2.0",
  "method": "tools/list",
  "id": 1
}
EOF

# 3. Use MCP inspector
npx @modelcontextprotocol/inspector ./bin/mcp-server
```

## Common Development Tasks

### Adding Real Implementation to a Mock Tool

1. Find the tool handler (grep for the handler name)
2. Replace mock logic with real New Relic API calls
3. Ensure mock mode still works
4. Test with real credentials

Example:
```go
func (s *Server) handleRealImplementation(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Keep mock mode support
    if s.isMockMode() {
        return s.getMockData("tool_name", params), nil
    }
    
    // Add real implementation
    client := s.getNRClient()
    if client == nil {
        return nil, fmt.Errorf("New Relic client not configured")
    }
    
    // Make actual API call
    result, err := client.Query(ctx, params["query"].(string))
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    
    return result, nil
}
```

### Debugging

1. **Enable debug logging:**
   ```bash
   LOG_LEVEL=debug ./bin/mcp-server
   ```

2. **Check handler registration:**
   ```go
   // Add logging in registerTools
   log.Printf("Registering tool: %s", tool.Name)
   ```

3. **Trace JSON-RPC flow:**
   - Set breakpoints in `protocol.go`
   - Log incoming requests
   - Check tool routing

## Code Style Guidelines

### DO:
- Always support mock mode
- Validate all input parameters
- Return meaningful error messages
- Keep handlers focused and simple
- Document complex logic

### DON'T:
- Remove mock support from existing tools
- Assume New Relic client exists
- Log sensitive data (API keys, query results)
- Create new build tag combinations
- Add more aspirational interfaces

## Known Gotchas

1. **Mock Mode Confusion**
   - Most tools return mock data even when not in mock mode
   - Check if tool has real implementation before relying on it

2. **Discovery Engine**
   - Interface exists but no implementation
   - Tools that depend on discovery just return hardcoded data

3. **Build Tags**
   - Avoid adding new build tags
   - Consider removing existing ones in refactor

4. **Multi-Account**
   - Code exists but not wired up
   - Only primary account works

5. **State Management**
   - Session state is in-memory only
   - No persistence between requests

## Contributing Guidelines

### Before Starting

1. Check if the feature is already "implemented" (returning mock data)
2. Understand the current architecture limitations
3. Discuss major changes in an issue first

### Pull Request Guidelines

1. **Keep changes focused** - One feature/fix per PR
2. **Maintain backward compatibility** - Don't break existing tools
3. **Test manually** - Automated tests are broken
4. **Document behavior** - Especially mock vs real
5. **Update tool reference** - Mark implementation status

### Priority Areas for Contribution

1. **Fix test infrastructure** - Get `make test` working
2. **Implement real tool handlers** - Replace mock with real
3. **Remove build tag complexity** - Simplify build system
4. **Add actual discovery** - Implement discovery engine
5. **Improve error handling** - Better error messages

## Resources

- [MCP Specification](https://modelcontextprotocol.com)
- [New Relic GraphQL API](https://docs.newrelic.com/docs/apis/nerdgraph/get-started/introduction-new-relic-nerdgraph/)
- [NRQL Reference](https://docs.newrelic.com/docs/query-your-data/nrql-new-relic-query-language/get-started/introduction-nrql-new-relics-query-language/)

## Getting Help

- Open an issue for bugs or questions
- Check existing tools for implementation patterns
- Review the diagnostics tool for debugging approaches
- Ask in discussions for architectural decisions
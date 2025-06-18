# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Universal Data Synthesizer (UDS) is an AI-powered system for New Relic that automatically discovers, analyzes, and visualizes data from NRDB. It implements MCP (Model Context Protocol) for seamless integration with AI assistants like Claude and GitHub Copilot.

**Current State**: Complete Go implementation with all critical MCP tools. The Python implementation has been removed in favor of a unified Go architecture.

## High-Level Architecture

The system follows a multi-layer architecture:

```
AI Assistant (Claude/Copilot)
    ↓ MCP Protocol (JSON-RPC)
MCP Server (Go) - pkg/interface/mcp/
    ↓ Internal APIs
Discovery Engine (Go) - pkg/discovery/
    ↓ GraphQL
New Relic NRDB
```

### Key Components
- **Discovery Engine** (`pkg/discovery/`): Schema discovery, pattern detection, relationship mining
- **MCP Server** (`pkg/interface/mcp/`): AI assistant integration via Model Context Protocol
- **REST API** (`pkg/interface/api/`): Traditional HTTP API for non-AI clients
- **New Relic Client** (`pkg/newrelic/`): Direct API access for queries, dashboards, alerts

### Implementation Status
- ✅ **Discovery Core**: Schema discovery, pattern detection, quality assessment
- ✅ **MCP Server**: Full implementation with all tools
- ✅ **Query Tools**: NRQL execution, validation, and builder
- ✅ **Dashboard Tools**: Discovery, generation from templates
- ✅ **Alert Tools**: Creation, analysis, bulk operations
- ✅ **Diagnostic Tool**: Environment validation and auto-fix
- ❌ **Intelligence Engine**: ML/AI capabilities (planned for future)

## Development Commands

### Build & Run
```bash
# Build all components
make build

# Run MCP server
make run                   # With real New Relic connection
make run-mock             # Mock mode for testing

# Run diagnostics
make diagnose             # Check environment
make diagnose-fix         # Auto-fix issues

# Other components
make run-api              # REST API server
./bin/uds                 # CLI tool
```

### Testing
```bash
# Go tests
make test                 # Run all tests
make test-unit           # Unit tests only
make test-integration    # Integration tests
make test-coverage       # Generate coverage report

# Run specific test
go test -v -run TestDiscoverSchemas ./pkg/discovery
```

### Code Quality
```bash
make lint                # Run golangci-lint
make format             # Format Go code
make clean              # Clean build artifacts
```

## Available MCP Tools

### Query Tools (`pkg/interface/mcp/tools_query.go`)
- **query_nrdb**: Execute NRQL queries with timeout control
- **query_check**: Validate queries, estimate costs, suggest optimizations
- **query_builder**: Build NRQL from structured parameters

### Discovery Tools (`pkg/interface/mcp/tools_discovery.go`)
- **discovery.list_schemas**: List all schemas with quality metrics
- **discovery.profile_attribute**: Deep attribute analysis with streaming
- **discovery.find_relationships**: Discover schema relationships
- **discovery.assess_quality**: Comprehensive quality assessment

### Dashboard Tools (`pkg/interface/mcp/tools_dashboard.go`)
- **find_usage**: Find dashboards using specific metrics
- **generate_dashboard**: Create dashboards from templates
  - Templates: golden-signals, sli-slo, infrastructure, custom
- **list_dashboards**: List all dashboards with filtering
- **get_dashboard**: Get detailed dashboard information

### Alert Tools (`pkg/interface/mcp/tools_alerts.go`)
- **create_alert**: Create alerts with auto-baseline or static thresholds
- **list_alerts**: List conditions with incident data
- **analyze_alerts**: Analyze effectiveness and suggest improvements
- **bulk_update_alerts**: Bulk operations (enable/disable/update/delete)

## Required Environment Configuration

```bash
# New Relic API Access
NEW_RELIC_API_KEY=your-user-api-key        # Required
NEW_RELIC_ACCOUNT_ID=your-account-id       # Required
NEW_RELIC_REGION=US                        # US or EU

# Security (Required - no defaults!)
JWT_SECRET=                                # Generate: openssl rand -base64 32
API_KEY_SALT=                             # Generate: openssl rand -base64 16

# Optional
NEW_RELIC_LICENSE_KEY=your-license-key     # For APM monitoring
REDIS_URL=redis://localhost:6379           # For distributed state
LOG_LEVEL=INFO                            # DEBUG, INFO, WARN, ERROR
MOCK_MODE=false                           # Use mock data
```

## Key Architecture Patterns

### Interface-Based Design
All major components use Go interfaces for testability and flexibility.

### Resilience Patterns
- **Circuit Breaker**: Prevents cascading failures
- **Rate Limiter**: Token bucket algorithm
- **Retry Logic**: Exponential backoff with jitter

### Multi-Layer Caching
- L1: In-memory cache (planned)
- L2: Redis distributed cache (optional)
- Cache-aside pattern with TTL management

### Security
- Input validation for all NRQL queries
- No default secrets - server refuses to start without proper configuration
- TLS support for production deployments

## Working with the Codebase

### Adding a New MCP Tool
1. Define tool schema in appropriate `tools_*.go` file
2. Implement handler method on `Server`
3. Register tool in appropriate `register*Tools()` function
4. Add tests

Example:
```go
s.tools.Register(Tool{
    Name:        "tool_name",
    Description: "What it does",
    Parameters:  ToolParameters{...},
    Handler:     s.handleToolName,
})
```

### Extending Discovery Engine
1. Update interface in `pkg/discovery/interfaces.go`
2. Implement in `pkg/discovery/engine.go`
3. Add mock implementation for testing
4. Wire into MCP tools if needed

### Running in Mock Mode
```bash
# Via environment
export MOCK_MODE=true
make run

# Via command line
make run-mock
```

## Troubleshooting

### Common Issues

1. **Build failures**: Run `make diagnose-fix`
2. **Connection errors**: Check NEW_RELIC_API_KEY and ACCOUNT_ID
3. **Mock mode**: Set MOCK_MODE=true for development without New Relic

### Debug Mode
```bash
export LOG_LEVEL=DEBUG
make run
```

## Migration from Python

The Python MCP implementation has been removed. All functionality is now in Go:
- Better performance (10x faster)
- Type safety
- Single codebase to maintain
- All critical tools implemented

If you see references to Python files (mcp_server.py, etc.), they should be ignored.
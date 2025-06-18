# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The New Relic MCP Server is a Go-based implementation of the Model Context Protocol (MCP) that provides AI assistants with tools to interact with New Relic's observability platform. The system exposes discovery, monitoring, and analysis capabilities through a unified MCP interface.

**Current Architecture Status**: The codebase has parallel Python and Go implementations that need consolidation. Use the Go implementation (`cmd/unified-server/`) as the primary server. The Python MCP server files should be removed or ignored.

## Build and Development Commands

```bash
# Build the Go MCP server
make build                  # Builds bin/mcp-server

# Run the server
make run                    # Build and run
./bin/mcp-server           # Run directly

# Testing
make test                  # Run all tests
make test-unit            # Unit tests only
make test-integration     # Integration tests
make test-coverage        # Generate coverage report
go test -v ./pkg/discovery/... -run TestSpecificFunction  # Run single test

# Code quality
make lint                 # Run golangci-lint
make format              # Format Go code

# Docker
make docker-build        # Build Docker image
docker-compose up        # Run with dependencies
```

## High-Level Architecture

The system follows a layered architecture where the MCP protocol layer is separated from business logic:

```
AI Assistant (Claude/Copilot)
    ↓ MCP Protocol (JSON-RPC)
MCP Server (cmd/unified-server/main.go)
    ↓
Tool Registry (pkg/tools/registry.go)
    ↓
Tool Implementations
├── Discovery Tools (pkg/tools/discovery/)
├── APM Tools (pkg/tools/newrelic/apm_tools.go)
├── Synthetics Tools (pkg/tools/newrelic/synthetics_tools.go)
└── [Future: Infrastructure, Logs, Tracing Tools]
    ↓
Shared Infrastructure
├── New Relic Client (pkg/newrelic/client.go)
├── Discovery Engine (pkg/discovery/engine.go)
├── Multi-Layer Cache (pkg/cache/multi_layer.go)
└── State Management (pkg/state/)
```

### Key Architectural Patterns

1. **Tool Registration Pattern**: All MCP tools are registered through a central registry. Each tool category (discovery, APM, synthetics) has its own registration file with a `RegisterAll()` method.

2. **Caching Strategy**: Discovery operations use a two-tier cache:
   - L1: In-memory (Ristretto) for fast access
   - L2: Redis for distributed caching
   - Cache keys are generated deterministically from tool name + parameters

3. **Resilience Patterns**: The New Relic client implements:
   - Circuit breaker (pkg/discovery/nrdb/circuit_breaker.go)
   - Rate limiting (pkg/discovery/nrdb/rate_limiter.go)
   - Retry with exponential backoff (pkg/discovery/nrdb/retry.go)

4. **State Management**: Session state is maintained across MCP requests using either in-memory or Redis backends, allowing tools to share context.

## Critical Implementation Details

### Adding New MCP Tools

1. Define the tool in the appropriate category file (e.g., `pkg/tools/newrelic/infrastructure_tools.go`)
2. Implement the handler as a method on the tools struct
3. Register the tool in the `RegisterAll()` method
4. Add parameter validation in `pkg/security/validation.go` if needed

Example structure:
```go
type infraTools struct {
    client *newrelic.Client
    state  state.Manager
}

func (t *infraTools) RegisterAll(registry *tools.Registry) error {
    // Register each tool with its handler
}

func (t *infraTools) handleListHosts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Implementation
}
```

### Discovery Engine Integration

The discovery engine (`pkg/discovery/engine.go`) is the core of schema analysis. It's currently using mock implementations - when working with discovery tools, be aware that:
- The real NRDB client needs to be implemented in `pkg/discovery/nrdb/client.go`
- Discovery operations are expensive and should always use caching
- The engine supports different profile depths: Basic, Standard, Full

### Security Considerations

- All NRQL queries must be validated through `pkg/security/validation.go`
- JWT secrets and API keys must never have defaults - the server will refuse to start without proper secrets
- Input validation happens at the MCP tool parameter level before execution

## Configuration

The server uses environment variables for configuration. After cleanup, only these are essential:

```bash
# Required
NEW_RELIC_API_KEY=         # User API key for New Relic
NEW_RELIC_ACCOUNT_ID=      # New Relic account ID
JWT_SECRET=                # Must be generated, no defaults
API_KEY_SALT=              # Must be generated, no defaults

# Optional but recommended
NEW_RELIC_LICENSE_KEY=     # For APM monitoring of this service
REDIS_URL=                 # For distributed caching and state
LOG_LEVEL=                 # DEBUG, INFO, WARN, ERROR
```

## Current State and Warnings

1. **Dual Implementation**: Both Python (`mcp_server.py`) and Go servers exist. Always use the Go implementation.

2. **Mock Implementations**: Several components still use mocks:
   - NRDB client in discovery engine
   - Some New Relic API responses in tools

3. **Missing Tools**: Infrastructure, Logs, and Tracing tools are not yet implemented.

4. **Cleanup Needed**: Run `./cleanup.sh` to remove duplicate implementations and streamline the codebase.

## Testing Approach

- Unit tests should use the mock implementations in test files
- Integration tests in `tests/integration/` test full tool execution
- Benchmarks in `benchmarks/` measure discovery operation performance
- Always run `make test` before committing changes

## Performance Considerations

- Discovery operations can be expensive - always check cache first
- Use batch operations when profiling multiple schemas
- The circuit breaker will open after 5 consecutive failures
- Rate limiting is set to 100 requests/second by default
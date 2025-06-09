# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server for New Relic's NerdGraph API, built with FastMCP. It enables LLMs like Claude to interact with New Relic accounts through natural language or direct tool invocations.

## Architecture

### Current Implementation
The project uses a sophisticated plugin-based architecture:

#### Core Components
- `main.py` - Application entry point with async initialization
- `server_simple.py` - Simplified server for testing/development
- `cli.py` - Command-line interface for direct tool execution
- `core/` - Core infrastructure:
  - `account_manager.py` - Multi-account credential management
  - `nerdgraph_client.py` - Async GraphQL client with retries
  - `entity_definitions.py` - New Relic entity definitions cache
  - `session_manager.py` - Conversation state management
  - `plugin_loader.py` - Auto-discovery of feature plugins
  - `plugin_manager.py` - Enhanced plugin system with dependency resolution
  - `cache.py` & `cache_improved.py` - Caching implementations
  - `health.py` - Health monitoring and metrics
  - `audit.py` - Audit logging for security/compliance
  - `error_sanitizer.py` - Security-focused error handling
  - `telemetry.py` - Built-in observability
- `transports/` - Communication layers:
  - `multi_transport.py` - STDIO/HTTP transport support
- `features/` - Plugin modules:
  - `common.py` - Core NerdGraph/NRQL query tools
  - `entities.py` - Entity search and golden signals
  - `apm.py` - APM metrics, transactions, deployments
  - `infrastructure.py` - Hosts, containers, K8s, processes
  - `logs.py` - Log search and analysis
  - `synthetics.py` - Synthetic monitor operations
  - `alerts.py` - Alert policies and incidents
  - `docs.py` - Documentation search capabilities

## Development Commands

### Setup and Installation
```bash
# Install production dependencies
make install

# Install development dependencies
make install-dev

# Initialize project structure
make init-project
```

### Running the Server
```bash
# Run advanced server with all features
python main.py

# Run simple server for testing
python server_simple.py

# Run with FastMCP CLI
fastmcp run server_simple.py:mcp

# Run in development mode
make run-dev

# Run with Docker
make docker-run
```

### Testing
```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage report
make test-coverage

# Run a single test file
pytest tests/test_specific.py -v

# Run tests matching a pattern
pytest -k "test_nerdgraph" -v
```

### Code Quality
```bash
# Format code
make format

# Run all linters
make lint

# Security checks
make security

# Pre-commit checks (format, lint, test)
make pre-commit
```

### Building and Deployment
```bash
# Build distribution packages
make build

# Build Docker image
make docker-build

# Prepare release
make release
```

## Key Implementation Details

### NerdGraph Client Pattern
All API calls go through `nerdgraph_client.execute_nerdgraph_query()` which:
- Adds authentication headers
- Handles retries with exponential backoff
- Returns consistent JSON responses including GraphQL errors
- Supports async operations

### Plugin Development
To create a new plugin:
1. Create a class extending `PluginBase` in `features/`
2. Implement `register()` method
3. Use `@app.tool()` decorator for MCP tools
4. Access core services via dependency injection

Example:
```python
from core.plugin_loader import PluginBase

class MyPlugin(PluginBase):
    @staticmethod
    def register(app: FastMCP, services: Dict[str, Any]):
        @app.tool()
        async def my_tool(param: str) -> Dict[str, Any]:
            nerdgraph = services["nerdgraph_client"]
            result = await nerdgraph.execute_nerdgraph_query(query)
            return {"result": result}
```

### Configuration
- Environment variables (primary method):
  - `NEW_RELIC_API_KEY` (required)
  - `NEW_RELIC_ACCOUNT_ID` (recommended)
  - `MCP_TRANSPORT` (stdio/http/multi)
  - `LOG_LEVEL` (DEBUG/INFO/WARNING/ERROR)
- `.env` file support for local development
- Multi-account profiles via AccountManager

### Error Handling
- Comprehensive error sanitization for security
- Structured error responses with context
- Audit logging for sensitive operations
- Graceful degradation for missing features

### Testing Patterns
- Use pytest fixtures for common setup
- Mock NerdGraph responses with `pytest-mock`
- Test async code with `pytest-asyncio`
- Integration tests use real API when available
- Performance benchmarks in `tests/benchmarks/`

## CLI Usage
The CLI provides direct access to all MCP tools:
```bash
# NRQL queries
python cli.py query "SELECT count(*) FROM Transaction SINCE 1 hour ago"

# Entity operations
python cli.py entities search --name "production"
python cli.py entities details "ENTITY_GUID"

# APM metrics
python cli.py apm metrics "My App" --time-range "SINCE 30 minutes ago"

# Infrastructure
python cli.py infra hosts --tag environment=production

# Account management
python cli.py config add-account --name prod --api-key KEY --account-id ID
python cli.py config use prod
```

## Performance Considerations
- Connection pooling for HTTP requests
- LRU caching with memory bounds
- Async operations throughout
- Efficient GraphQL queries (request only needed fields)
- Batch operations where possible

## Security Best Practices
- API keys stored server-side only
- Request signing with HMAC for sensitive operations
- Error sanitization to prevent information leakage
- Audit logging for compliance
- Rate limiting and timeout protection

## Debugging Tips
- Enable debug logging: `LOG_LEVEL=DEBUG`
- Check audit logs in `audit_logs/` directory
- Use `make run-dev` for verbose output
- Test individual tools with CLI before integration
- Monitor server metrics via built-in telemetry
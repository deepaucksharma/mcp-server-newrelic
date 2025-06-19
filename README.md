# New Relic MCP Server

## Overview

A comprehensive Model Context Protocol (MCP) server that provides AI assistants (GitHub Copilot, Claude, etc.) with intelligent access to New Relic observability data. Built in Go with full MCP protocol compliance, this server enables AI-powered observability workflows including NRQL queries, dashboard generation, alert management, and bulk operations.

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- New Relic API Key and Account ID
- Docker (optional)

### Installation

```bash
# Clone the repository
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic

# Set up environment
cp .env.example .env
# Edit .env with your New Relic credentials

# Build all components
make build

# Run diagnostics
make diagnose
```

### Running the Server

```bash
# Run MCP server (default: stdio transport)
make run

# Run in mock mode (no New Relic connection)
make run-mock

# Run with HTTP transport
./bin/mcp-server --transport http --port 8080

# Run with Docker
docker-compose up
```

## 📊 Implementation Status

### ✅ Implemented Features

| Category | Tools | Status |
|----------|-------|--------|
| **Query Execution** | `query_nrdb`, `query_check`, `query_builder` | ✅ Complete |
| **Discovery** | `list_schemas`, `profile_attribute`, `find_relationships`, `assess_quality` | ✅ Complete |
| **Dashboards** | `find_usage`, `generate_dashboard`, `list_dashboards`, `get_dashboard` | ✅ Complete |
| **Alerts** | `create_alert`, `list_alerts`, `analyze_alerts`, `bulk_update_alerts` | ✅ Complete |
| **State Management** | Session tracking, caching (Memory/Redis) | ✅ Complete |
| **Resilience** | Circuit breaker, retry logic, rate limiting | ✅ Complete |

### 🚧 In Progress

- Enhanced error handling and telemetry
- Comprehensive test coverage
- CI/CD pipeline setup

### 📝 Planned Features

- Intelligence Engine (Python) for ML-powered insights
- Advanced bulk operations
- Multi-account support
- EU region support

## 🛠️ Available MCP Tools

### Query Tools
```json
{
  "tool": "query_nrdb",
  "params": {
    "query": "SELECT count(*) FROM Transaction WHERE appName = 'myapp' SINCE 1 hour ago",
    "account_id": "optional-override"
  }
}
```

### Discovery Tools
```json
{
  "tool": "discovery.list_schemas",
  "params": {
    "filter": "Transaction",
    "include_quality": true
  }
}
```

### Dashboard Tools
```json
{
  "tool": "generate_dashboard",
  "params": {
    "template": "golden-signals",
    "service_name": "myapp",
    "name": "My App Golden Signals"
  }
}
```

### Alert Tools
```json
{
  "tool": "create_alert",
  "params": {
    "name": "High Error Rate",
    "query": "SELECT percentage(count(*), WHERE error IS true) FROM Transaction",
    "sensitivity": "medium",
    "auto_baseline": true
  }
}
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────┐
│            MCP Client (AI Assistant)             │
└────────────────────┬────────────────────────────┘
                     │ MCP Protocol (JSON-RPC)
┌────────────────────▼────────────────────────────┐
│                Go MCP Server                     │
├─────────────────────────────────────────────────┤
│  • MCP Handler (stdio/http/sse)                 │
│  • Tool Registry & Execution                    │
│  • Request Validation & Error Handling          │
├─────────────────────────────────────────────────┤
│  • Discovery Engine (schema analysis)           │
│  • State Manager (sessions & caching)           │
│  • New Relic Client (GraphQL/NerdGraph)         │
├─────────────────────────────────────────────────┤
│  • Resilience Layer (circuit breaker, retry)    │
│  • Observability (APM, logging, metrics)        │
└────────────────────┬────────────────────────────┘
                     │ HTTPS/GraphQL
                     ▼
            ┌─────────────────┐
            │  New Relic API  │
            └─────────────────┘
```

## 🔧 Configuration

### Required Environment Variables

```bash
# New Relic API Access
NEW_RELIC_API_KEY=your-user-api-key
NEW_RELIC_ACCOUNT_ID=your-account-id
NEW_RELIC_REGION=US  # or EU (EU support planned)

# Optional: New Relic APM (for monitoring the server itself)
NEW_RELIC_LICENSE_KEY=your-license-key
NEW_RELIC_APP_NAME=mcp-server-newrelic

# Server Configuration
MCP_TRANSPORT=stdio  # stdio, http, or sse
SERVER_PORT=8080
LOG_LEVEL=INFO

# State Management
REDIS_URL=redis://localhost:6379  # Optional, defaults to in-memory
```

See [.env.example](./.env.example) for complete configuration options.

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific test suites
make test-unit         # Unit tests
make test-integration  # Integration tests
make test-mcp         # MCP protocol tests

# Generate coverage report
make test-coverage

# Run benchmarks
make test-benchmarks
```

## 📚 Documentation

- **[Architecture Overview](./docs/ARCHITECTURE.md)** - System design and components
- **[Development Guide](./docs/DEVELOPMENT.md)** - Setup and contribution guidelines
- **[API Reference](./docs/API_REFERENCE.md)** - Complete tool documentation
- **[Deployment Guide](./docs/DEPLOYMENT.md)** - Production deployment
- **[Roadmap](./ROADMAP.md)** - Future development plans
- **[Technical Specification](./TECHNICAL_SPEC.md)** - Detailed specifications

## 🚀 Deployment

### Docker

```bash
# Build image
make build-docker

# Run with Docker Compose
docker-compose up -d

# Or run standalone
docker run -p 8080:8080 --env-file .env mcp-server-newrelic
```

### Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/

# Configure secrets
kubectl create secret generic newrelic-creds \
  --from-literal=api-key=$NEW_RELIC_API_KEY \
  --from-literal=account-id=$NEW_RELIC_ACCOUNT_ID
```

## 🤝 Contributing

We welcome contributions! Please see our [Development Guide](./docs/DEVELOPMENT.md) for:
- Code style guidelines
- Testing requirements
- Pull request process
- Architecture decisions

### Development Setup

```bash
# Install development tools
make install-tools

# Run linting
make lint

# Format code
make format

# Run in development mode
make dev
```

## 🔒 Security

- API keys are never logged or exposed
- All inputs are validated
- Rate limiting prevents abuse
- See [SECURITY.md](./SECURITY.md) for reporting vulnerabilities

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built on the [Model Context Protocol](https://modelcontextprotocol.io/) standard
- Integrates with [New Relic NerdGraph API](https://docs.newrelic.com/docs/apis/nerdgraph/get-started/introduction-new-relic-nerdgraph/)
- Uses [Ristretto](https://github.com/dgraph-io/ristretto) for high-performance caching

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/deepaucksharma/mcp-server-newrelic/issues)
- **Discussions**: [GitHub Discussions](https://github.com/deepaucksharma/mcp-server-newrelic/discussions)
- **Documentation**: [docs/](./docs/)

---

**Version**: 1.0.0-beta | **Last Updated**: June 2025
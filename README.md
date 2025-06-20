# New Relic MCP Server

A **Discovery-First** Model Context Protocol (MCP) server that provides AI assistants with intelligent access to New Relic observability data. Unlike traditional tools that assume data schemas, this server explores, understands, and adapts to your actual NRDB landscape.

**Key Features:**
- 🔍 **Discovery-First**: Never assumes data structures, always explores first
- 🧩 **120+ Granular Tools**: Atomic operations that compose into workflows
- 🚀 **Production-Ready**: Built in Go with resilience, caching, and monitoring
- 🤖 **AI-Optimized**: Rich metadata guides intelligent tool usage
- 🔄 **Multi-Transport**: STDIO, HTTP, and SSE support

## 🚀 Quick Start

### Prerequisites
- Go 1.21+ or Docker
- New Relic API Key and Account ID

### 5-Minute Setup

#### Option 1: Using Docker (Recommended)
```bash
# Clone and configure
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic

# Set up credentials
cp .env.example .env
# Edit .env with your New Relic API key and account ID

# Start services
docker-compose up -d

# Verify installation
curl http://localhost:8080/health
```

#### Option 2: Build from Source
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

### Claude Desktop Configuration
Add to `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "newrelic": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "--env-file", ".env", "uds-mcp:latest"]
    }
  }
}
```

### Common Use Cases

1. **Troubleshooting Performance**: "Find the slowest transactions in the last 24 hours and show me their error rates"
2. **Infrastructure Monitoring**: "List all hosts with CPU usage over 80% and their associated applications"
3. **Alert Management**: "Show me all critical alerts that fired in the last week"
4. **Data Exploration**: "What custom attributes are we sending with our Transaction events?"

## 🚀 Key Features

### Discovery-First Architecture
- **Never Assumes**: Explores what data actually exists before querying
- **Adaptive Queries**: Builds queries based on discovered schemas
- **Handles Variations**: Works across teams with different instrumentation
- **Progressive Understanding**: Builds knowledge incrementally from evidence

### Technical Capabilities
- **Full MCP Compliance**: Implements MCP DRAFT-2025 specification with JSON-RPC 2.0
- **100+ Granular Tools**: Atomic tools that compose into sophisticated workflows
- **Workflow Orchestration**: Sequential, parallel, conditional, and saga patterns
- **Production-Ready**: Built-in resilience with circuit breakers, retries, and rate limiting
- **Flexible Deployment**: Support for STDIO, HTTP, and SSE transports
- **State Management**: Session tracking with pluggable storage (Memory/Redis)
- **Mock Mode**: Development mode with realistic responses without New Relic connection
- **Cross-Account Support**: Query and manage resources across multiple New Relic accounts

## 🎯 Why Discovery-First?

Traditional tools fail when they assume data structures exist. Our discovery-first approach:

1. **Discovers** what data actually exists
2. **Adapts** queries to match reality
3. **Succeeds** where hardcoded queries fail

**Result**: Works across different teams, handles schema changes, and uncovers insights you didn't know to look for.

## 🛠️ Tool Categories

The server provides 120+ granular tools:

- **Discovery Tools**: Explore schemas, find patterns, assess data quality
- **Query Tools**: Execute NRQL with validation, build queries dynamically
- **Analysis Tools**: Detect anomalies, find correlations, forecast trends
- **Action Tools**: Create dashboards, manage alerts, configure entities
- **Governance Tools**: Analyze usage, optimize costs, audit resources

Example discovery-first workflow:
```json
// 1. Discover what exists
{"tool": "discovery.explore_event_types"}

// 2. Build query from discovery
{"tool": "nrql.build_from_discovery", "params": {"intent": "error_rate"}}

// 3. Create monitoring based on findings
{"tool": "dashboard.create_from_discovery"}
```

See the [API Reference](./docs/api/reference.md) for complete tool documentation.


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

### Core Documentation
- **[Architecture Overview](./docs/architecture/overview.md)** - System design and components
- **[Discovery-First Philosophy](./docs/architecture/discovery-first.md)** - Our approach explained
- **[API Reference](./docs/api/reference.md)** - Complete tool documentation
- **[Deployment Guide](./docs/guides/deployment.md)** - Production deployment

### Quick Links
- **[Documentation Index](./docs/README.md)** - Complete documentation listing
- **[Development Guide](./docs/DEVELOPMENT.md)** - Contributing guidelines
- **[Roadmap 2025](./ROADMAP_2025.md)** - Development roadmap

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
# New Relic MCP Server

⚠️ **Early Development Notice**: This is an early prototype with limited functionality. See [Current Capabilities](docs/CURRENT_CAPABILITIES.md) and [Implementation Gaps](docs/IMPLEMENTATION_GAPS_ANALYSIS.md) for details.\n\nA **Discovery-First** Model Context Protocol (MCP) server that provides AI assistants with intelligent access to New Relic observability data. Unlike traditional tools that assume data schemas, this server explores, understands, and adapts to your actual NRDB landscape.

**Key Features:**
- 🔍 **Discovery-First**: Never assumes data structures, always explores first
- 🧩 **120+ Granular Tools** *(Planned - Currently ~10-15)*: Atomic operations that compose into workflows
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

# Run in mock mode (no New Relic connection needed)
make run-mock
# Mock mode provides realistic data for all tools without requiring New Relic credentials

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

### Your First Discovery

Try a discovery-first query right away:
```bash
# Discover what data you have
echo '{"jsonrpc":"2.0","method":"discovery.explore_event_types","id":1}' | ./bin/mcp-server

# Returns: Transaction, SystemSample, Log, etc. - your actual data!
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
- **100+ Granular Tools** *(Goal - Currently ~10-15 basic tools)*: Atomic tools design for sophisticated workflows
- **Workflow Orchestration**: Sequential, parallel, conditional, and saga patterns
- **Production-Ready**: Built-in resilience with circuit breakers, retries, and rate limiting
- **Flexible Deployment**: Support for STDIO, HTTP, and SSE transports
- **State Management**: Session tracking with pluggable storage (Memory/Redis)
- **Mock Mode**: Comprehensive development mode with realistic responses for all 100+ tools without New Relic connection
- **Cross-Account Support**: Query and manage resources across multiple New Relic accounts without reconfiguration

## 🎯 Why Discovery-First?

Traditional tools fail when they assume data structures exist. Our discovery-first approach:

1. **Discovers** what data actually exists
2. **Adapts** queries to match reality
3. **Succeeds** where hardcoded queries fail

**Result**: Works across different teams, handles schema changes, and uncovers insights you didn't know to look for.

## 🛠️ Tool Categories

The server is designed for 120+ granular tools across categories:

- **Discovery Tools**: Basic event type exploration *(limited implementation)*
- **Query Tools**: Basic NRQL execution *(no validation or dynamic building yet)*
- **Analysis Tools**: *Not yet implemented* - Will detect anomalies, correlations, trends
- **Action Tools**: *Not yet implemented* - Will create dashboards, alerts, configurations
- **Governance Tools**: *Not yet implemented* - Will analyze usage, costs, resources

**Current Status**: Only basic discovery and query tools are functional. See [Implementation Gaps](docs/IMPLEMENTATION_GAPS_ANALYSIS.md) for details.

Example discovery-first workflow (partially implemented):
```json
// 1. Discover what exists (✓ implemented)
{"tool": "discovery.explore_event_types"}

// 2. Execute basic query (✓ implemented)
{"tool": "nrql.execute", "params": {"query": "SELECT count(*) FROM Transaction"}}

// 3. Advanced features (✗ not yet implemented)
// - nrql.build_from_discovery
// - dashboard.create_from_discovery
// - analysis.find_anomalies
```

See the [API Reference](./docs/api/reference.md) for complete tool documentation.

## 🤖 AI Assistant Integration

The MCP Server is designed to work seamlessly with AI assistants like Claude, GitHub Copilot, and GPT-based agents. Here's how they connect:

### How It Works
```
AI Assistant → Natural Language Request → MCP Server Tools → New Relic Data → Results
```

The AI assistant:
1. **Receives** your observability question
2. **Discovers** what data exists using discovery tools
3. **Builds** appropriate queries based on findings
4. **Executes** the workflow and returns insights

### Supported Assistants
- **Claude Desktop**: Native MCP support via configuration
- **GitHub Copilot**: HTTP/REST interface integration
- **GPT/ChatGPT**: OpenAPI spec for custom GPTs
- **Custom Agents**: Use our Python/TypeScript SDKs

### Integration Benefits
- **Zero Training**: AI automatically understands your data structure
- **Adaptive Queries**: Works with any New Relic schema
- **Natural Language**: Ask questions in plain English
- **Workflow Automation**: Complex investigations in seconds

Learn more in our [AI Integration Guide](./docs/guides/llm-integration.md).

## 🔧 Configuration

### Required Environment Variables

```bash
# New Relic API Access
NEW_RELIC_API_KEY=your-user-api-key
NEW_RELIC_ACCOUNT_ID=your-primary-account-id  # Default account
NEW_RELIC_REGION=US  # or EU (both regions supported)

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

### Multi-Account Usage

The server uses your configured account as the default, but you can query any account your API key has access to:

```json
// Query a different account
{"tool": "query_nrdb", "params": {"query": "SELECT count(*) FROM Transaction", "account_id": "2345678"}}

// List dashboards in another account  
{"tool": "list_dashboards", "params": {"account_id": "3456789"}}

// Create alert in specific account
{"tool": "create_alert", "params": {"name": "High Error Rate", "account_id": "4567890", ...}}
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
- **[Development Guide](./docs/guides/development.md)** - Contributing guidelines
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

We welcome contributions! Please see our [Development Guide](./docs/guides/development.md) for:
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

### Built-in Security Features
- **Authentication**: API key validation with no default credentials
- **Data Protection**: No logging of sensitive data or credentials
- **Input Validation**: All inputs sanitized and validated
- **Rate Limiting**: Prevents abuse and protects resources
- **Tool Safety**: Three-tier safety classification (Safe/Caution/Destructive)
- **Audit Logging**: All mutations tracked for compliance

### Security Best Practices
- Use read-only API keys when possible
- Enable HTTPS for HTTP transport mode
- Configure timeouts and rate limits appropriately
- Review audit logs regularly
- Keep the server updated with security patches

See [SECURITY.md](./SECURITY.md) for detailed security guidelines and vulnerability reporting.

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
See [CHANGELOG](./CHANGELOG.md) for release history.


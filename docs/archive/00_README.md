# New Relic MCP Server

A **Discovery-First** Model Context Protocol (MCP) server that provides AI assistants with intelligent access to New Relic observability data. Built in Go with production-grade reliability, this server enables sophisticated data exploration, analysis, and management through carefully designed tools.

## 🚀 Quick Links

- [Getting Started](01_GETTING_STARTED.md) - 5-minute setup guide
- [Installation](02_INSTALLATION.md) - Detailed installation instructions
- [Configuration](03_CONFIGURATION.md) - Configuration reference
- [Tools Catalog](30_TOOLS_OVERVIEW.md) - Complete list of available tools
- [Examples](50_EXAMPLES_OVERVIEW.md) - Real-world usage examples

## 🎯 Key Features

### Discovery-First Architecture
- **Never assumes data structures** - Always explores your actual NRDB landscape first
- **Intelligent data profiling** - Understands your custom attributes and event types
- **Relationship mapping** - Discovers connections between different data sources
- **Quality assessment** - Evaluates data completeness and reliability

### Comprehensive Tool Suite
Provides **40+ tools** across 8 categories:

- **Discovery Tools** - Explore schemas, attributes, and relationships
- **Query Tools** - Execute NRQL queries with adaptive optimization
- **Alert Tools** - Create, update, and manage intelligent alerts
- **Dashboard Tools** - Build and manage custom dashboards
- **Analysis Tools** - Statistical analysis, anomaly detection, correlations
- **Governance Tools** - Usage analysis, cost optimization, compliance
- **Workflow Tools** - Orchestrate complex operations
- **Bulk Operations** - Efficient batch processing

### Production-Ready Infrastructure
- **Multi-Transport Support** - STDIO (for Claude), HTTP, and Server-Sent Events
- **Enterprise Security** - JWT authentication, API key management, audit logging
- **High Performance** - Concurrent processing, intelligent caching, query optimization
- **Resilient Design** - Circuit breakers, retries, graceful degradation
- **Mock Mode** - Full functionality without New Relic connection for development

### AI-Optimized Design
- **Rich Tool Metadata** - Detailed descriptions guide intelligent tool selection
- **Contextual Examples** - Each tool includes usage examples
- **Error Guidance** - Helpful error messages with resolution suggestions
- **Progressive Disclosure** - Tools compose from simple to complex workflows

## 📋 Requirements

- **Go 1.21+** or **Docker**
- **New Relic Account** with:
  - User API Key (for authentication)
  - Account ID
  - Appropriate permissions for desired operations
- **Optional**: Redis for distributed state management

## 🚀 Quick Start

### Option 1: Docker (Recommended)
```bash
# Clone repository
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic

# Configure credentials
cp .env.example .env
# Edit .env with your New Relic credentials

# Start server
docker-compose up -d

# Verify
curl http://localhost:8080/health
```

### Option 2: From Source
```bash
# Clone and configure
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic

# Set up environment
cp .env.example .env
# Edit .env with your credentials

# Build
make build

# Run diagnostics
make diagnose

# Start server
make run
```

### Claude Desktop Integration
Add to your `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "newrelic": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "--env-file", ".env", "mcp-newrelic:latest"]
    }
  }
}
```

## 🔍 Discovery-First Example

Experience the discovery-first approach:

```bash
# 1. Discover what event types you have
echo '{"jsonrpc":"2.0","method":"discovery.explore_event_types","id":1}' | ./bin/mcp-server

# 2. Explore attributes for a specific event type
echo '{"jsonrpc":"2.0","method":"discovery.explore_attributes","params":{"event_type":"Transaction"},"id":2}' | ./bin/mcp-server

# 3. Profile a specific attribute
echo '{"jsonrpc":"2.0","method":"discovery.profile_attribute","params":{"event_type":"Transaction","attribute":"duration"},"id":3}' | ./bin/mcp-server

# 4. Query with discovered knowledge
echo '{"jsonrpc":"2.0","method":"query_nrdb","params":{"query":"SELECT average(duration) FROM Transaction FACET appName SINCE 1 hour ago"},"id":4}' | ./bin/mcp-server
```

## 📚 Common Use Cases

### Performance Troubleshooting
"Find the slowest transactions in the last 24 hours and analyze their patterns"
- Uses: `discovery.explore_event_types` → `query_nrdb` → `analysis.detect_anomalies`

### Infrastructure Monitoring
"Show me hosts with high CPU usage and their associated applications"
- Uses: `discovery.explore_attributes` → `query_nrdb` → `dashboard.create_from_discovery`

### Alert Management
"Create intelligent alerts for all my critical services based on their baselines"
- Uses: `analysis.calculate_baseline` → `alert.create_from_baseline`

### Cost Optimization
"Analyze my data ingest patterns and suggest optimization opportunities"
- Uses: `governance.analyze_usage` → `governance.optimize_costs`

## 🏗️ Architecture

The MCP server follows a modular, layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP Clients (Claude, etc.)               │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                     Transport Layer                          │
│                 (STDIO, HTTP, SSE)                          │
├─────────────────────────────────────────────────────────────┤
│                    Protocol Handler                          │
│                  (JSON-RPC 2.0 + MCP)                       │
├─────────────────────────────────────────────────────────────┤
│                     Tool Registry                            │
│              (40+ Registered Tools)                          │
├─────────────────────────────────────────────────────────────┤
│                    Core Services                             │
│  ┌─────────────┬──────────────┬────────────┬─────────────┐ │
│  │  Discovery  │    Query     │  Analysis  │ Governance  │ ││  │   Engine    │   Engine     │  Engine    │   Engine    │ │
│  └─────────────┴──────────────┴────────────┴─────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                  Infrastructure Layer                        │
│  ┌─────────────┬──────────────┬────────────┬─────────────┐ │
│  │    Auth     │    State     │   Cache    │   Logger    │ │
│  │  Manager    │   Manager    │  Manager   │             │ │
│  └─────────────┴──────────────┴────────────┴─────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                   New Relic Client                           │
│                    (API + NRDB)                              │
└─────────────────────────────────────────────────────────────┘
```

## 🧪 Development

### Mock Mode
Run without New Relic connection for development:
```bash
make run-mock
```

### Running Tests
```bash
# Unit tests
make test

# Integration tests
make test-integration

# E2E tests
make test-e2e

# All tests
make test-all
```

### Adding New Tools
See [Contributing Tools Guide](87_CONTRIBUTING_TOOLS.md) for detailed instructions.

## 📖 Documentation Index

### Core Documentation (00-09)
- [Getting Started](01_GETTING_STARTED.md) - Quick start guide
- [Installation](02_INSTALLATION.md) - Detailed setup instructions
- [Configuration](03_CONFIGURATION.md) - Configuration reference
- [Concepts](04_CONCEPTS.md) - Core concepts explained
- [Features](05_FEATURES.md) - Feature overview
- [Requirements](06_REQUIREMENTS.md) - System requirements
- [Changelog](07_CHANGELOG.md) - Version history
- [Roadmap](08_ROADMAP.md) - Future plans
- [FAQ](09_FAQ.md) - Frequently asked questions

### Architecture & Design (10-19)
- [Architecture Overview](10_ARCHITECTURE_OVERVIEW.md) - System design
- [Discovery-First Design](11_ARCHITECTURE_DISCOVERY_FIRST.md) - Discovery philosophy
- [State Management](12_ARCHITECTURE_STATE_MANAGEMENT.md) - State and caching
- [Transport Layers](13_ARCHITECTURE_TRANSPORT_LAYERS.md) - STDIO, HTTP, SSE

### API & Tools (20-39)
- [API Overview](20_API_OVERVIEW.md) - API structure
- [Tools Overview](30_TOOLS_OVERVIEW.md) - Complete tool catalog
- [Discovery Tools](31_TOOLS_DISCOVERY.md) - Discovery tool reference
- [Query Tools](32_TOOLS_QUERY.md) - Query tool reference

### User Guides & Examples (40-59)
- [Quick Start Guide](40_GUIDE_QUICKSTART.md) - 5-minute tutorial
- [Claude Integration](41_GUIDE_CLAUDE_INTEGRATION.md) - Claude Desktop setup
- [Discovery Workflows](43_GUIDE_DISCOVERY_WORKFLOWS.md) - Discovery patterns
- [Examples Overview](50_EXAMPLES_OVERVIEW.md) - Example scenarios

### Development & Operations (60-89)
- [Testing Strategy](60_TESTING_STRATEGY.md) - Testing approach
- [Deployment Overview](70_DEPLOYMENT_OVERVIEW.md) - Deployment options
- [Development Setup](80_DEVELOPMENT_SETUP.md) - Dev environment
- [Contributing](86_CONTRIBUTING.md) - Contribution guide

## 🤝 Contributing

We welcome contributions! See our [Contributing Guide](86_CONTRIBUTING.md) for:
- Development setup
- Coding standards
- Testing requirements
- Pull request process

## 📊 Project Status

### ✅ Current Implementation
- Core MCP protocol with multi-transport support
- 40+ production-ready tools across 8 categories
- Discovery engine with intelligent data exploration
- Advanced query optimization and caching
- Comprehensive error handling and resilience
- Mock mode for development and testing

### 🚧 In Progress
- Kubernetes deployment manifests
- Advanced caching strategies
- Distributed state management
- Additional analysis algorithms

### 🗺️ Roadmap
See [Project Roadmap](08_ROADMAP.md) for detailed future plans.

## 🐛 Troubleshooting

### Common Issues

**Authentication Failed**
- Verify your API key has necessary permissions
- Check account ID is correct
- Ensure API key type matches configuration

**No Data Returned**
- Verify data exists in the specified time range
- Check event type names are correct (case-sensitive)
- Use discovery tools to explore available data

**Performance Issues**
- Enable caching for repeated queries
- Use time range filters to limit data
- Consider bulk operations for multiple queries

See [Troubleshooting Guide](79_OPERATIONS_TROUBLESHOOTING.md) for comprehensive solutions.

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/deepaucksharma/mcp-server-newrelic/issues)
- **Discussions**: [GitHub Discussions](https://github.com/deepaucksharma/mcp-server-newrelic/discussions)
- **Documentation**: This repository's `/docs` folder
- **Community**: [MCP Discord](https://discord.gg/mcp-community)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## 🙏 Acknowledgments

- Built on the [Model Context Protocol](https://modelcontextprotocol.org/) specification
- Powered by [New Relic](https://newrelic.com/) APIs
- Inspired by the AI community's need for better observability tools

---

For detailed information on any topic, explore our comprehensive documentation using the index above.
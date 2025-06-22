# New Relic MCP Server

A **schema-agnostic** Model Context Protocol (MCP) server that provides AI assistants with intelligent, adaptive access to New Relic data through dynamic discovery.

**Perfect for**: Automated observability analysis, incident troubleshooting, performance optimization, and dashboard generation.

## 🚀 Key Features

- **🔍 Discovery-First**: No hardcoded schemas - all data structures discovered at runtime
- **📊 Golden Signals Analysis**: Real-time latency, traffic, errors, and saturation analysis  
- **📈 Dashboard Intelligence**: List, analyze, and understand dashboard structures
- **🤖 AI-Optimized**: Rich metadata and examples designed for LLM consumption
- **⚡ Production-Ready**: Comprehensive error handling, caching, and monitoring support

## ⚡ Quick Start (< 10 minutes)

### Prerequisites

- **Node.js 18+** (Check: `node --version`)
- **New Relic User API Key** ([Get yours here](https://one.newrelic.com/api-keys))
- **New Relic Account** with data to analyze

### 1. Installation

```bash
# Clone and install
git clone https://github.com/newrelic/mcp-server-newrelic.git
cd mcp-server-newrelic
npm install
```

### 2. Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your New Relic credentials
NEW_RELIC_API_KEY=YOUR_USER_API_KEY_HERE
NEW_RELIC_ACCOUNT_ID=YOUR_ACCOUNT_ID
NEW_RELIC_REGION=US  # or EU
```

### 3. First Run

```bash
# Start the MCP server
npm run dev

# You should see:
# [INFO] MCP Server New Relic started successfully
# [INFO] Tool registry initialized - 5 tools available
```

### 4. Verify with AI Assistant

Connect to Claude Desktop by adding this to your MCP settings:

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "node",
      "args": ["/path/to/mcp-server-newrelic/dist/index.js"],
      "env": {
        "NEW_RELIC_API_KEY": "your_key_here",
        "NEW_RELIC_ACCOUNT_ID": "your_account_id"
      }
    }
  }
}
```

**✅ Success!** You should now see New Relic tools available in Claude Desktop.

### Testing

```bash
# Run all tests
npm run test

# Run tests with coverage
npm run test:coverage

# Run linting
npm run lint

# Format code
npm run format
```

## 🏗️ Architecture

**Discovery-First, Schema-Agnostic Design**

```
┌─────────────────────────────────────────────┐
│          AI Assistant (Claude)              │
└────────────────┬────────────────────────────┘
                 │ MCP Protocol
┌────────────────▼────────────────────────────┐
│         New Relic MCP Server                │
├─────────────────────────────────────────────┤
│  🛠️ Tool Registry (5 Production Tools)      │
│  🔍 Discovery Engine (NRQL + keyset)        │ 
│  📊 Golden Signals Analyzer                 │
│  💾 Intelligent Cache (Memory + Redis)      │
│  🔐 NerdGraph Client (Auth + Regions)       │
└────────────────┬────────────────────────────┘
                 │ GraphQL (NerdGraph API)
                 ▼
         🌐 New Relic Platform
```

**Key Principles:**
- **No Hardcoded Schemas**: Everything discovered at runtime
- **Graceful Degradation**: Works with whatever data is available  
- **AI-Optimized**: Rich metadata and confidence scoring
- **Production-Ready**: Comprehensive error handling and monitoring

## 🛠️ Tools & Examples

### 1. Discover Available Data (`discover_schemas`)

**What it does**: Discovers all event types, attributes, and metrics in your New Relic account.

```javascript
// Tool call
{
  "account_id": 12345,
  "include_attributes": true,
  "include_metrics": true
}

// Response
{
  "account_id": 12345,
  "event_types": [
    {
      "name": "Transaction",
      "sample_count": 150000,
      "first_seen": "2024-06-21T00:00:00Z",
      "last_seen": "2024-06-22T12:00:00Z",
      "attributes": ["name", "duration", "response.status", "error"]
    }
  ],
  "summary": {
    "total_event_types": 15,
    "total_events": 2500000,
    "total_metrics": 120
  }
}
```

### 2. Find Applications & Services (`discover_entities`)

**What it does**: Finds and categorizes all entities in your account.

```javascript
// Tool call
{
  "account_id": 12345,
  "entity_type": "APPLICATION", 
  "domain": "APM"
}

// Response
{
  "account_id": 12345,
  "entities": [
    {
      "guid": "entity-guid-123",
      "name": "My Web App",
      "type": "APPLICATION",
      "domain": "APM",
      "tags": [
        {"key": "environment", "values": ["production"]},
        {"key": "team", "values": ["backend"]}
      ]
    }
  ],
  "summary": {
    "total_entities": 25,
    "by_domain": {"APM": 8, "BROWSER": 5, "INFRA": 12}
  }
}
```

### 3. Analyze Performance (`analyze_golden_signals`)

**What it does**: Analyzes the four golden signals of monitoring for any entity.

```javascript
// Tool call
{
  "entity_guid": "entity-guid-123",
  "account_id": 12345,
  "time_range": "1 hour ago"
}

// Response
{
  "entity": {
    "guid": "entity-guid-123",
    "name": "My Web App",
    "type": "APPLICATION",
    "domain": "APM"
  },
  "golden_signals": {
    "latency": {
      "p50": 0.125,
      "p95": 0.450,
      "p99": 1.200,
      "unit": "seconds",
      "status": "good"
    },
    "traffic": {
      "requests_per_minute": 1250,
      "total_requests": 75000,
      "status": "normal"
    },
    "errors": {
      "error_rate": 0.8,
      "error_count": 600,
      "status": "good"
    }
  },
  "analysis": {
    "confidence": 95,
    "recommendations": [
      "Performance looks healthy - continue monitoring"
    ]
  }
}
```

### 4. List Dashboards (`list_dashboards`)

**What it does**: Lists and analyzes dashboards with filtering capabilities.

```javascript
// Tool call
{
  "account_id": 12345,
  "name_filter": "performance",
  "limit": 10
}

// Response
{
  "account_id": 12345,
  "dashboards": [
    {
      "guid": "dashboard-guid-456",
      "name": "Application Performance",
      "description": "Key performance metrics",
      "owner": {"email": "user@company.com"},
      "permissions": "public",
      "total_widgets": 12,
      "pages": [
        {
          "name": "Overview",
          "widget_count": 8,
          "widget_types": ["viz.line", "viz.bar", "viz.table"]
        }
      ]
    }
  ],
  "summary": {
    "total_dashboards": 3,
    "dashboard_types": {"public": 2, "private": 1}
  }
}
```

### 5. Execute Custom Queries (`execute_nrql`)

**What it does**: Executes any NRQL query with validation and formatting.

```javascript
// Tool call
{
  "account_id": 12345,
  "query": "SELECT average(duration) FROM Transaction WHERE appName = 'MyApp' SINCE 1 hour ago FACET name LIMIT 10"
}

// Response
{
  "results": [
    {"name": "WebTransaction/Action/users#show", "average": 0.234},
    {"name": "WebTransaction/Action/orders#create", "average": 0.456}
  ],
  "metadata": {
    "event_types": ["Transaction"],
    "facets": ["name"],
    "messages": []
  }
}
```

## 🎯 Common Workflows

### Troubleshoot Slow Application
1. **Discover entities** → Find your application
2. **Analyze golden signals** → Check latency, errors, traffic
3. **Execute custom NRQL** → Drill down into specific slow transactions

### Dashboard Analysis
1. **List dashboards** → Find relevant dashboards
2. **Discover schemas** → Understand available data for improvements
3. **Analyze golden signals** → Validate dashboard accuracy

### Platform Overview
1. **Discover schemas** → Understand data landscape
2. **Discover entities** → Map your infrastructure
3. **List dashboards** → Review existing monitoring coverage

## ⚙️ Configuration Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEW_RELIC_API_KEY` | ✅ | - | Your New Relic User API Key ([Get one here](https://one.newrelic.com/api-keys)) |
| `NEW_RELIC_ACCOUNT_ID` | ✅ | - | Default account ID for queries |
| `NEW_RELIC_REGION` | ❌ | `US` | API region: `US` or `EU` |
| `DEBUG` | ❌ | `false` | Enable debug logging |

### Optional Performance Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `CACHE_TTL` | `300` | Cache TTL in seconds |
| `DISCOVERY_BATCH_SIZE` | `10` | Batch size for discovery operations |
| `NEW_RELIC_TIMEOUT` | `30` | API timeout in seconds |

### Example .env file

```bash
# Required
NEW_RELIC_API_KEY=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXX
NEW_RELIC_ACCOUNT_ID=1234567
NEW_RELIC_REGION=US

# Optional  
DEBUG=false
CACHE_TTL=300
```

## 🔧 Troubleshooting

### Problem: "Authentication failed" error

**Solution**: Check your API key
```bash
# Verify your API key is correct
curl -H "Api-Key: YOUR_KEY_HERE" https://api.newrelic.com/graphql \
  -d '{"query": "{ actor { user { name } } }"}'
```

### Problem: "No event types found" 

**Solution**: Ensure account has data
- Verify your account ID is correct
- Check that applications are reporting data
- Try a different time range

### Problem: Server won't start

**Solution**: Check Node.js version and dependencies
```bash
node --version  # Should be 18+
npm install     # Reinstall dependencies
```

### Problem: Tools not visible in Claude Desktop

**Solution**: Check MCP configuration
- Verify the server path in Claude Desktop settings
- Ensure environment variables are set in MCP config
- Check server is running: `npm run dev`

### Problem: Slow responses

**Solution**: Optimize configuration
```bash
# Increase cache TTL for better performance
CACHE_TTL=900

# Reduce batch size if memory constrained  
DISCOVERY_BATCH_SIZE=5
```

**Still having issues?** 
- Check [GitHub Issues](https://github.com/newrelic/mcp-server-newrelic/issues)
- Review logs with `DEBUG=true npm run dev`

## 🔬 Development

### Testing

```bash
# Unit tests (17 tests passing ✅)
npm run test

# Integration tests with real New Relic API
NEW_RELIC_API_KEY=real_key npm run test

# Coverage report
npm run test:coverage

# Specific test file
npx vitest src/tools/__tests__/registry.test.ts
```

### Code Quality

```bash
npm run lint        # ESLint + TypeScript checks
npm run format      # Prettier formatting  
npm run typecheck   # TypeScript validation
```

### Project Structure

```
src/
├── index.ts                    # MCP server entry point
├── core/
│   ├── types.ts               # Zod schemas & TypeScript types
│   ├── config.ts              # Environment configuration
│   ├── discovery.ts           # NRQL-based discovery engine
│   ├── cache.ts               # Memory + Redis caching
│   └── nerdgraph.ts           # GraphQL client (US/EU regions)
├── tools/
│   ├── registry.ts            # Tool management & execution
│   └── __tests__/
│       └── registry.test.ts   # Comprehensive tool tests
└── utils/
    └── logger.ts              # Structured logging
```

## 📞 Support & Contributing

### Getting Help

- **Issues**: [GitHub Issues](https://github.com/newrelic/mcp-server-newrelic/issues)
- **Discussions**: [GitHub Discussions](https://github.com/newrelic/mcp-server-newrelic/discussions)  
- **Documentation**: [Technical Specification](./tech_spec.md)

### Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-tool`)
3. **Follow** the discovery-first architecture patterns
4. **Add** comprehensive tests and documentation
5. **Submit** a pull request

### Development Principles

- **Discovery-First**: Always probe available data before making assumptions
- **Type-Safe**: Use Zod schemas for runtime validation
- **AI-Optimized**: Include rich metadata and confidence scoring
- **Production-Ready**: Comprehensive error handling and monitoring

---

## 📄 License

MIT License - see [LICENSE](./LICENSE) file for details.

Built with ❤️ for the New Relic developer community.
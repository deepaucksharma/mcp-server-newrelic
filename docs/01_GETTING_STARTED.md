# Getting Started

This guide will get you up and running with the Enhanced MCP Server New Relic in 5 minutes.

## Prerequisites

- **Node.js 20+** or **Bun 1.0+**
- **New Relic account** with API access
- **API Key** with NRQL query permissions

## Quick Setup

### 1. Clone and Install

```bash
git clone <repository-url>
cd mcp-server-newrelic
npm install
```

### 2. Configure New Relic Credentials

Create your environment configuration:

```bash
# Required
export NEW_RELIC_API_KEY="NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
export NEW_RELIC_ACCOUNT_ID="1234567"

# Optional
export NEW_RELIC_REGION="US"  # or "EU"
export DEBUG="true"           # Enable debug logging
```

### 3. Test Your Configuration

```bash
# Test discovery functionality
npm run discover

# Run E2E tests (requires valid credentials)
npm run test:e2e:quick

# Run specific E2E test suites
npm run test:e2e:discovery
npm run test:e2e:tools
```

### 4. Start the MCP Server

```bash
# Development mode with hot reload
npm run dev

# Production mode
npm start
```

## First Steps with the Enhanced Server

### 1. Environment Discovery

Start by discovering your New Relic environment:

```typescript
// This should be your first call with any new account
const environment = await mcp.call('discover.environment', {
  includeHealth: true,
  maxEntities: 50,
  forceRefresh: false
});
```

**What you'll get:**
- Complete inventory of monitored entities
- Available telemetry event types and characteristics
- OpenTelemetry vs APM instrumentation detection
- Schema guidance for optimal queries
- Observability gaps and recommendations

### 2. Generate Golden Signals Dashboard

Create an intelligent dashboard for your key service:

```typescript
// Generate adaptive golden signals dashboard
const dashboard = await mcp.call('generate.golden_dashboard', {
  entity_guid: 'MXxBUE18QVBQTElDQVRJT058MTIzNDU2',  // From environment discovery
  dashboard_name: 'My Service - Golden Signals',
  timeframe_hours: 1,
  create_dashboard: false  // Preview first
});
```

**What you'll get:**
- Automatically adapted queries for your instrumentation type
- Latency (P50, P95, P99), Traffic, Errors, and Saturation monitoring
- Multi-page dashboard with overview and detailed analysis
- Alert threshold overlays where applicable

### 3. Compare Entity Performance

Find optimization opportunities across similar entities:

```typescript
// Compare all APPLICATION entities
const comparison = await mcp.call('compare.similar_entities', {
  comparison_strategy: 'by_type',
  entity_type: 'APPLICATION',
  max_entities: 10,
  sort_by: 'overall_performance'
});
```

**What you'll get:**
- Performance rankings with benchmarks
- Outlier identification (top performers and those needing attention)
- Best practices from high-performing entities
- Specific optimization recommendations

## Understanding the Enhanced Tools

### Composite Tools (New)
- **`discover.environment`** - Comprehensive environment analysis
- **`generate.golden_dashboard`** - Intelligent dashboard generation
- **`compare.similar_entities`** - Performance comparison and analysis

### Enhanced Existing Tools
- **`run_nrql_query`** - Now with schema validation and caching
- **`search_entities`** - Enhanced with discovery insights
- **`get_entity_details`** - Enriched with golden metrics
- **`discover_schemas`** - Comprehensive schema discovery

### Cache Management Tools
- **`cache.stats`** - Monitor cache performance
- **`cache.clear`** - Manage cache contents

## Configuration Options

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEW_RELIC_API_KEY` | ✅ | - | New Relic API key (NRAK-...) |
| `NEW_RELIC_ACCOUNT_ID` | ✅ | - | Primary account ID |
| `NEW_RELIC_REGION` | ❌ | `US` | Data center region (US/EU) |
| `DEBUG` | ❌ | `false` | Enable debug logging |
| `CACHE_TTL_MULTIPLIER` | ❌ | `1.0` | Adjust cache TTL (0.5-2.0) |

### Advanced Configuration

For multiple accounts or complex setups, see [03_CONFIGURATION.md](03_CONFIGURATION.md).

## Troubleshooting

### Common Issues

**❌ "Invalid API credentials"**
- Verify your API key format (starts with `NRAK-`)
- Check account ID is correct
- Ensure API key has NRQL query permissions

**❌ "No entities found"**
- Verify entities are reporting data
- Check account ID is correct
- Try running `discover.environment` first

**❌ "Unknown event types"**
- Run `discover_schemas` before writing custom NRQL
- Use `discover.environment` for schema guidance
- Check if data exists in the specified time window

### Debug Mode

Enable detailed logging:

```bash
export DEBUG="true"
npm run dev
```

This will show:
- Discovery process details
- Cache hit/miss information
- Query execution timings
- Error details and suggestions

### Cache Issues

Monitor and manage cache performance:

```typescript
// Check cache health
const stats = await mcp.call('cache.stats', {});

// Clear problematic cache entries
const cleared = await mcp.call('cache.clear', {
  pattern: 'discovery:',  // Clear discovery cache
  confirm: false
});
```

## Next Steps

1. **Explore Tools**: Read [30_TOOLS_OVERVIEW.md](30_TOOLS_OVERVIEW.md) for complete tool documentation
2. **Learn Workflows**: See [42_GUIDE_DISCOVERY_WORKFLOWS.md](42_GUIDE_DISCOVERY_WORKFLOWS.md) for patterns
3. **View Examples**: Check [50_EXAMPLES_OVERVIEW.md](50_EXAMPLES_OVERVIEW.md) for real scenarios
4. **Understand Architecture**: Read [10_ARCHITECTURE_OVERVIEW.md](10_ARCHITECTURE_OVERVIEW.md) for deeper insights

## Support

- **Documentation**: Browse the complete docs in this directory
- **Examples**: See the `examples/` directory for code samples
- **Issues**: Report problems via the issue tracker
- **Community**: Join discussions in the community forum

---

**Next**: [02_INSTALLATION.md](02_INSTALLATION.md) for detailed installation options
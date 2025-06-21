# Tools Overview

This document provides a comprehensive overview of all available tools in the Enhanced MCP Server New Relic.

## Tool Categories

### Enhanced Existing Tools
These are improved versions of standard New Relic tools with discovery-first intelligence:

- **`run_nrql_query`** - Execute NRQL queries with schema validation and caching
- **`search_entities`** - Search entities with enhanced filtering and caching  
- **`get_entity_details`** - Get entity details with golden metrics and patterns
- **`discover_schemas`** - Comprehensive schema discovery (typically run first)

### Composite Intelligence Tools
High-level tools that combine multiple operations for complex workflows:

- **`discover.environment`** - One-call comprehensive environment discovery
- **`generate.golden_dashboard`** - Intelligent dashboard generation with adaptation
- **`compare.similar_entities`** - Performance comparison and benchmarking

### Platform Analysis Tools
Advanced tools for platform governance and analysis:

- **`dashboard_generate`** - Adaptive dashboard creation with templates
- **`platform_analyze_adoption`** - Cross-account platform usage analysis

### Cache Management Tools
Tools for monitoring and managing the intelligent caching system:

- **`cache.stats`** - Monitor cache performance and health
- **`cache.clear`** - Clear cache entries with pattern-based filtering

## Tool Usage Patterns

### 1. Discovery-First Workflow

```typescript
// Always start with environment discovery
const env = await mcp.call('discover.environment', {
  includeHealth: true,
  maxEntities: 50
});

// Then work with specific entities
const dashboard = await mcp.call('generate.golden_dashboard', {
  entity_guid: env.entities[0].guid,
  create_dashboard: false  // preview first
});
```

### 2. Performance Analysis Workflow

```typescript
// Compare entities for optimization opportunities
const comparison = await mcp.call('compare.similar_entities', {
  comparison_strategy: 'by_type',
  entity_type: 'APPLICATION',
  max_entities: 10
});

// Generate dashboards for top performers and outliers
for (const entity of comparison.entities) {
  if (entity.rank.overall <= 2 || entity.rank.overall >= 4) {
    await mcp.call('generate.golden_dashboard', {
      entity_guid: entity.entity.guid,
      dashboard_name: `${entity.entity.name} - Performance Analysis`
    });
  }
}
```

### 3. Schema Discovery Workflow

```typescript
// Discover what's available before querying
const schemas = await mcp.call('discover_schemas', {
  account_id: 12345,
  include_attributes: true,
  include_metrics: true
});

// Use discovered schemas for informed querying
const query = `SELECT count(*) FROM ${schemas.event_types[0].name} 
               WHERE ${schemas.summary.service_identifier_field} = 'my-service'
               SINCE 1 hour ago`;

const results = await mcp.call('run_nrql_query', {
  account_id: 12345,
  query: query,
  validate_schema: true
});
```

## Tool Capabilities Matrix

| Tool | Read | Write | Cache | Validation | Analytics |
|------|------|-------|-------|------------|-----------|
| `run_nrql_query` | ✅ | ❌ | ✅ | ✅ | ❌ |
| `search_entities` | ✅ | ❌ | ✅ | ✅ | ❌ |
| `get_entity_details` | ✅ | ❌ | ✅ | ✅ | ✅ |
| `discover_schemas` | ✅ | ❌ | ✅ | ❌ | ✅ |
| `discover.environment` | ✅ | ❌ | ✅ | ❌ | ✅ |
| `generate.golden_dashboard` | ✅ | ✅ | ❌ | ✅ | ✅ |
| `compare.similar_entities` | ✅ | ❌ | ❌ | ❌ | ✅ |
| `dashboard_generate` | ✅ | ✅ | ❌ | ✅ | ❌ |
| `platform_analyze_adoption` | ✅ | ❌ | ❌ | ❌ | ✅ |
| `cache.stats` | ✅ | ❌ | ❌ | ❌ | ✅ |
| `cache.clear` | ❌ | ✅ | ✅ | ❌ | ❌ |

## Performance Characteristics

### Latency Expectations
- **Enhanced Tools**: 100-500ms (with caching)
- **Composite Tools**: 1-5 seconds (multiple operations)
- **Platform Analysis**: 5-30 seconds (cross-account)
- **Cache Operations**: <50ms

### Caching Behavior
- **discovery**: 5 minutes TTL (high priority, adaptive)
- **goldenMetrics**: 2 minutes TTL (critical priority, adaptive)
- **entityDetails**: 10 minutes TTL (medium priority, static)
- **dashboards**: 15 minutes TTL (low priority, static)
- **analytics**: 30 minutes TTL (medium priority, adaptive)

## Error Handling

All tools implement consistent error handling:

### Common Error Types
- **Invalid credentials**: API key or account access issues
- **Unknown schemas**: Attempting to query non-existent event types
- **Rate limiting**: New Relic API rate limit exceeded
- **Cache issues**: Memory or performance problems

### Error Response Format
```typescript
{
  error: string,           // Human-readable error message
  suggestion: string,      // Actionable suggestion for resolution
  available_types?: [],    // Alternative options when applicable
  cached?: boolean,        // Whether response came from cache
  freshness?: string       // Data freshness indicator
}
```

## Security Considerations

### API Key Usage
- All tools use the configured API key from environment variables
- No API keys are logged or stored in cache
- Supports both US and EU regions

### Data Privacy
- No customer data is permanently stored
- Cache uses memory only (no disk persistence)
- Automatic cache cleanup and TTL enforcement

### Rate Limiting
- Intelligent request throttling based on New Relic limits
- Background refresh to minimize real-time API calls
- Cache-first strategy to reduce API load

## Best Practices

### 1. Always Start with Discovery
```typescript
// ✅ Good - discover before querying
const env = await mcp.call('discover.environment', {});
const results = await mcp.call('run_nrql_query', {
  query: `SELECT count(*) FROM ${env.eventTypes[0].name}`
});

// ❌ Bad - hardcoded assumptions
const results = await mcp.call('run_nrql_query', {
  query: 'SELECT count(*) FROM Transaction WHERE appName = "myapp"'
});
```

### 2. Use Preview Mode for Destructive Operations
```typescript
// ✅ Good - preview first
const preview = await mcp.call('generate.golden_dashboard', {
  entity_guid: guid,
  create_dashboard: false
});
// Review the dashboard JSON, then create
const created = await mcp.call('generate.golden_dashboard', {
  entity_guid: guid,
  create_dashboard: true
});
```

### 3. Monitor Cache Performance
```typescript
// Regular cache health monitoring
const stats = await mcp.call('cache.stats', {});
if (stats.hitRate < 0.5) {
  // Consider adjusting TTL or query patterns
}
```

### 4. Handle Errors Gracefully
```typescript
try {
  const result = await mcp.call('discover.environment', {});
} catch (error) {
  if (error.message.includes('credentials')) {
    // Handle authentication issues
  } else if (error.message.includes('rate limit')) {
    // Handle rate limiting
  }
}
```

## Tool Selection Guide

### When to Use Each Tool

**For Initial Setup:**
1. `discover.environment` - Get complete overview
2. `discover_schemas` - Deep schema analysis

**For Regular Monitoring:**
1. `get_entity_details` - Entity health and metrics
2. `run_nrql_query` - Custom queries
3. `cache.stats` - Performance monitoring

**For Analysis and Optimization:**
1. `compare.similar_entities` - Performance benchmarking
2. `platform_analyze_adoption` - Platform governance
3. `generate.golden_dashboard` - Visualization

**For Development and Troubleshooting:**
1. `cache.clear` - Reset cache state
2. `discover_schemas` - Debug schema issues

## Next Steps

- **[31_TOOLS_COMPOSITE.md](31_TOOLS_COMPOSITE.md)** - Deep dive into composite tools
- **[32_TOOLS_ENHANCED.md](32_TOOLS_ENHANCED.md)** - Enhanced existing tools reference
- **[33_TOOLS_ANALYTICS.md](33_TOOLS_ANALYTICS.md)** - Analytics and caching tools
- **[50_EXAMPLES_OVERVIEW.md](50_EXAMPLES_OVERVIEW.md)** - Real-world usage examples
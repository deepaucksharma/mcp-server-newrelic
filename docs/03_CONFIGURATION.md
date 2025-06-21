# Configuration Reference

Complete configuration guide for the Enhanced MCP Server New Relic, covering all environment variables, caching strategies, and advanced options.

## Configuration Overview

The server uses a hierarchical configuration system:
1. **Environment Variables** (highest priority)
2. **Configuration File** (if provided)
3. **Default Values** (lowest priority)

## Environment Variables

### Required Configuration

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `NEW_RELIC_API_KEY` | string | New Relic API key (NRAK-...) | `NRAK-ABCD1234...` |
| `NEW_RELIC_ACCOUNT_ID` | number | Primary New Relic account ID | `1234567` |

### Optional Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `NEW_RELIC_REGION` | `US\|EU` | `US` | New Relic data center region |
| `DEBUG` | boolean | `false` | Enable debug logging |
| `CACHE_TTL_MULTIPLIER` | number | `1.0` | Adjust all cache TTL values (0.5-2.0) |
| `NODE_ENV` | string | `development` | Runtime environment |

### E2E Testing Configuration

For comprehensive testing with multiple accounts:

| Variable | Type | Description |
|----------|------|-------------|
| `E2E_ACCOUNT_LEGACY_APM` | number | Legacy APM account for testing |
| `E2E_API_KEY_LEGACY` | string | API key for legacy account |
| `E2E_ACCOUNT_MODERN_OTEL` | number | OpenTelemetry account for testing |
| `E2E_API_KEY_OTEL` | string | API key for OTEL account |
| `E2E_ACCOUNT_MIXED_DATA` | number | Mixed telemetry account |
| `E2E_API_KEY_MIXED` | string | API key for mixed account |

## Caching Configuration

The server implements intelligent caching with adaptive TTL strategies. You can tune these via environment variables or the configuration object.

### Cache Strategy Types

```typescript
interface CacheStrategy {
  name: string;
  ttl: number;          // Base TTL in milliseconds
  adaptiveTtl: boolean; // Whether to adjust TTL based on access patterns
  maxAge: number;       // Absolute maximum age before forced refresh
  refreshThreshold: number; // 0-1, when to trigger background refresh
  priority: 'low' | 'medium' | 'high' | 'critical';
}
```

### Default Cache Strategies

```typescript
const defaultStrategies = {
  discovery: {
    ttl: 5 * 60 * 1000,        // 5 minutes base
    adaptiveTtl: true,
    maxAge: 30 * 60 * 1000,    // 30 minutes max
    refreshThreshold: 0.8,      // Refresh when 80% of TTL elapsed
    priority: 'high',
  },
  goldenMetrics: {
    ttl: 2 * 60 * 1000,        // 2 minutes base
    adaptiveTtl: true,
    maxAge: 10 * 60 * 1000,    // 10 minutes max
    refreshThreshold: 0.7,      // Refresh when 70% of TTL elapsed
    priority: 'critical',
  },
  entityDetails: {
    ttl: 10 * 60 * 1000,       // 10 minutes base
    adaptiveTtl: false,
    maxAge: 60 * 60 * 1000,    // 1 hour max
    refreshThreshold: 0.9,      // Refresh when 90% of TTL elapsed
    priority: 'medium',
  },
  dashboards: {
    ttl: 15 * 60 * 1000,       // 15 minutes base
    adaptiveTtl: false,
    maxAge: 4 * 60 * 60 * 1000, // 4 hours max
    refreshThreshold: 0.8,
    priority: 'low',
  },
  analytics: {
    ttl: 30 * 60 * 1000,       // 30 minutes base
    adaptiveTtl: true,
    maxAge: 2 * 60 * 60 * 1000, // 2 hours max
    refreshThreshold: 0.75,
    priority: 'medium',
  },
};
```

### Cache Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CACHE_TTL_MULTIPLIER` | `1.0` | Multiply all TTL values (0.5-2.0) |
| `CACHE_MAX_ENTRIES` | `500` | Maximum cache entries before LRU eviction |
| `CACHE_DISABLE` | `false` | Disable all caching (for debugging) |
| `CACHE_BACKGROUND_REFRESH` | `true` | Enable background cache refresh |

### Cache Tuning Examples

```bash
# Aggressive caching (longer TTL)
export CACHE_TTL_MULTIPLIER="2.0"

# Conservative caching (shorter TTL)
export CACHE_TTL_MULTIPLIER="0.5"

# Disable caching for debugging
export CACHE_DISABLE="true"

# Large environment with more cache
export CACHE_MAX_ENTRIES="1000"
```

## Discovery Configuration

The discovery engine can be tuned for different account sizes and data patterns.

### Discovery Settings

```typescript
interface DiscoveryConfig {
  cache: {
    type: 'memory';
    ttl: {
      schemas: number;      // Schema discovery TTL
      attributes: number;   // Attribute profiling TTL  
      serviceId: number;    // Service identifier TTL
      errors: number;       // Error pattern TTL
    };
  };
  confidence: {
    minimum: number;        // Minimum confidence threshold
    optimal: number;        // Optimal confidence target
  };
}
```

### Discovery Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DISCOVERY_CONFIDENCE_MIN` | `0.7` | Minimum confidence threshold |
| `DISCOVERY_CONFIDENCE_OPT` | `0.9` | Optimal confidence target |
| `DISCOVERY_SCHEMA_TTL` | `14400000` | Schema cache TTL (4 hours) |
| `DISCOVERY_ATTR_TTL` | `1800000` | Attribute cache TTL (30 min) |

## Regional Configuration

### US Region (Default)
```bash
export NEW_RELIC_REGION="US"
# Uses: https://api.newrelic.com/graphql
```

### EU Region
```bash
export NEW_RELIC_REGION="EU"
# Uses: https://api.eu.newrelic.com/graphql
```

### Custom GraphQL Endpoint
```bash
export NEW_RELIC_GRAPHQL_URL="https://custom.newrelic.endpoint/graphql"
```

## MCP Protocol Configuration

### Transport Settings

The server supports STDIO transport (required for MCP):

```typescript
interface MCPConfig {
  transport: 'stdio';
  http?: {
    port: number;
    cors: string[];
  };
}
```

### Protocol Capabilities

```typescript
const capabilities = {
  tools: {},           // Tool calling capability
  resources: {},       // Resource access capability
};
```

## Advanced Configuration

### Performance Tuning

```bash
# Node.js performance options
export NODE_OPTIONS="--max-old-space-size=2048"

# V8 garbage collection tuning
export NODE_OPTIONS="--gc-interval=100"

# Enable source maps for debugging
export NODE_OPTIONS="--enable-source-maps"
```

### Logging Configuration

```bash
# Enable debug logging
export DEBUG="true"

# Log level (if using structured logging)
export LOG_LEVEL="info"  # debug, info, warn, error

# Log format
export LOG_FORMAT="text"  # text, json
```

### Security Configuration

```bash
# Disable telemetry collection
export TELEMETRY_DISABLED="true"

# Custom user agent
export USER_AGENT="MCP-NewRelic/2.0.0 (Custom)"
```

## Configuration File (Optional)

You can provide configuration via a file:

```yaml
# config.yaml
newrelic:
  apiKey: "${NEW_RELIC_API_KEY}"
  accountId: "${NEW_RELIC_ACCOUNT_ID}"
  region: "US"
  graphqlUrl: "https://api.newrelic.com/graphql"

discovery:
  cache:
    type: "memory"
    ttl:
      schemas: 14400000      # 4 hours
      attributes: 1800000    # 30 minutes
      serviceId: 7200000     # 2 hours
      errors: 1800000        # 30 minutes
  confidence:
    minimum: 0.7
    optimal: 0.9

mcp:
  transport: "stdio"
  http:
    port: 3000
    cors: ["*"]

telemetry:
  enabled: true
```

Load with:
```bash
export CONFIG_FILE="./config.yaml"
node dist/index.js
```

## Environment-Specific Configurations

### Development Environment

```bash
# .env.development
NODE_ENV=development
DEBUG=true
CACHE_TTL_MULTIPLIER=0.5
LOG_LEVEL=debug
TELEMETRY_DISABLED=true
```

### Production Environment

```bash
# .env.production
NODE_ENV=production
DEBUG=false
CACHE_TTL_MULTIPLIER=1.0
LOG_LEVEL=info
TELEMETRY_DISABLED=false
```

### Testing Environment

```bash
# .env.test
NODE_ENV=test
DEBUG=false
CACHE_DISABLE=true
DISCOVERY_CONFIDENCE_MIN=0.5
```

## Validation

The server validates configuration on startup:

```typescript
// Configuration schema validation
const ConfigSchema = z.object({
  newrelic: z.object({
    apiKey: z.string().min(1, "API key is required"),
    accountId: z.string().min(1, "Account ID is required"), 
    region: z.enum(['US', 'EU']).default('US'),
    graphqlUrl: z.string().url().optional(),
  }),
  // ... additional schema validation
});
```

### Validation Errors

Common validation errors and solutions:

**❌ "API key is required"**
- Set `NEW_RELIC_API_KEY` environment variable
- Ensure the key starts with "NRAK-"

**❌ "Account ID is required"**
- Set `NEW_RELIC_ACCOUNT_ID` environment variable
- Ensure it's a numeric value

**❌ "Invalid region"**
- Use "US" or "EU" for `NEW_RELIC_REGION`

## Configuration Best Practices

### 1. Use Environment Variables for Secrets
```bash
# ✅ Good
export NEW_RELIC_API_KEY="NRAK-..."

# ❌ Bad - don't hardcode in files
apiKey: "NRAK-hardcoded-key"
```

### 2. Tune Cache for Your Environment
```bash
# Large accounts with lots of entities
export CACHE_TTL_MULTIPLIER="1.5"
export CACHE_MAX_ENTRIES="1000"

# Small accounts with frequent changes
export CACHE_TTL_MULTIPLIER="0.5"
export CACHE_MAX_ENTRIES="200"
```

### 3. Enable Debug Mode for Troubleshooting
```bash
# Temporary debugging
DEBUG=true npm run dev

# Production debugging (logs to stderr)
DEBUG=true node dist/index.js 2>debug.log
```

### 4. Monitor Cache Performance
```typescript
// Regular cache monitoring
const stats = await mcp.call('cache.stats', {});
console.log(`Cache hit rate: ${stats.hitRate}`);

// Adjust TTL if hit rate is low
if (stats.hitRate < 0.6) {
  // Consider increasing CACHE_TTL_MULTIPLIER
}
```

## Troubleshooting Configuration

### Common Issues

**❌ "Configuration validation failed"**
- Check environment variable names and values
- Ensure required variables are set
- Validate data types (numbers vs strings)

**❌ "Cache performance issues"**
- Monitor with `cache.stats` tool
- Adjust `CACHE_TTL_MULTIPLIER`
- Check memory usage

**❌ "Discovery confidence too low"**
- Reduce `DISCOVERY_CONFIDENCE_MIN`
- Check data availability in account
- Verify time ranges

### Debug Configuration Loading

```bash
# Show configuration on startup
DEBUG=true node dist/index.js

# Should show:
# [MCP-NR] INFO: Platform services initialized {
#   "region": "US",
#   "accountId": "1234567"
# }
```

---

**Next**: [10_ARCHITECTURE_OVERVIEW.md](10_ARCHITECTURE_OVERVIEW.md) for architecture details
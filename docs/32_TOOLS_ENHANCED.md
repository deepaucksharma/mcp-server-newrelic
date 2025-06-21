# Enhanced Existing Tools Reference

Detailed reference for enhanced versions of existing MCP tools, now with discovery intelligence and improved UX.

## Overview

The Enhanced MCP Server maintains compatibility with existing tools while adding significant intelligence:

- **Discovery Validation**: Schema validation before query execution
- **Intelligent Caching**: Context-aware caching with adaptive TTL
- **Rich Error Handling**: Helpful error messages with suggestions
- **LLM-Optimized Output**: Structured responses designed for AI consumption

## Enhanced Tool Categories

### 1. Query Tools
- `run_nrql_query` - Execute NRQL with discovery validation
- `search_entities` - Entity search with rich metadata

### 2. Entity Tools  
- `get_entity_details` - Comprehensive entity information
- `get_entity_golden_metrics` - Golden signals for any entity

### 3. Platform Tools
- `platform_analyze_adoption` - OpenTelemetry adoption analysis

### 4. Cache Management Tools
- `cache.stats` - Cache performance monitoring
- `cache.clear` - Cache management and cleanup

## 1. Enhanced Query Tools

### `run_nrql_query`

**Enhanced Features:**
- Pre-query schema validation
- Intelligent error suggestions
- Automatic query optimization hints
- Result caching with freshness indicators

**Tool Definition:**
```typescript
{
  name: 'run_nrql_query',
  description: `Execute NRQL queries with discovery-first validation and intelligent caching.
  
  **Enhanced Features**:
  - ✅ Schema validation before execution
  - 📈 Automatic performance optimization
  - 🚀 Intelligent caching with freshness tracking
  - 💡 Helpful error messages with suggestions
  
  **Use Cases**:
  - Custom analytics and reporting
  - Data exploration and validation
  - Performance metric analysis
  - Troubleshooting and investigation`,
  inputSchema: {
    type: 'object',
    properties: {
      query: { type: 'string', description: 'NRQL query to execute' },
      account_id: { type: 'string', description: 'Account ID (optional if configured)' },
      timeout_seconds: { type: 'number', default: 30, maximum: 300 },
      use_cache: { type: 'boolean', default: true },
      explain_query: { type: 'boolean', default: false }
    },
    required: ['query']
  }
}
```

**Enhancement Implementation:**
```typescript
const preHandler = async (params: any) => {
  const { query } = params;
  
  // Schema validation
  const validation = await this.discovery.validateQuery(query);
  if (!validation.valid) {
    return {
      error: `Query validation failed: ${validation.error}`,
      suggestion: validation.suggestion
    };
  }
  
  // Performance optimization hints
  if (validation.optimizationHints.length > 0) {
    this.logger.info('Query optimization hints available', {
      hints: validation.optimizationHints
    });
  }
};
```

**Example Enhanced Response:**
```markdown
# 📈 NRQL Query Results

## Query Executed
```sql
SELECT average(duration) FROM Transaction 
WHERE appName = 'checkout-service' 
SINCE 1 hour ago
```

## Results (15 rows)
| timestamp | average(duration) |
|-----------|------------------|
| 2024-01-15 10:00:00 | 0.245 |
| 2024-01-15 10:05:00 | 0.267 |
...

## Performance Insights
- ✅ Query executed in 234ms
- 📈 Results cached for 2 minutes
- 💡 Consider adding TIMESERIES for trend analysis

## Schema Validation
- ✅ Event type 'Transaction' confirmed available
- ✅ Attribute 'appName' found in schema
- ✅ Time range optimized for data retention
```

### `search_entities`

**Enhanced Features:**
- Rich entity metadata
- Golden signals availability indicators
- Health status assessment
- Relationship mapping

**Tool Definition:**
```typescript
{
  name: 'search_entities',
  description: `Search for entities with rich metadata and golden signals context.
  
  **Enhanced Features**:
  - 📊 Golden signals availability indicators
  - 🟢 Health status assessment
  - 🔗 Entity relationship mapping
  - 🏷️ Rich tagging and metadata
  
  **Perfect for**:
  - Finding entities for dashboard creation
  - Identifying monitoring coverage gaps
  - Understanding service dependencies
  - Planning observability improvements`,
  inputSchema: {
    type: 'object',
    properties: {
      query: { type: 'string', description: 'Entity search query' },
      entity_types: { type: 'array', items: { type: 'string' } },
      include_health: { type: 'boolean', default: false },
      include_golden_signals: { type: 'boolean', default: true },
      max_results: { type: 'number', default: 25, maximum: 100 }
    }
  }
}
```

**Example Enhanced Response:**
```markdown
# 🔍 Entity Search Results

## Search Query: "checkout"
**Found**: 3 entities matching criteria

### 📱 checkout-service-mobile
- **Type**: APPLICATION (APM)
- **GUID**: `ABC123...`
- **Health**: ✅ Healthy (0.2% error rate)
- **Golden Signals**: 📊 Available (Latency: 145ms P95, Traffic: 12 req/min)
- **Language**: Java
- **Environment**: production
- **Last Reported**: 2 minutes ago

### 🌐 checkout-service-web  
- **Type**: APPLICATION (APM)
- **GUID**: `DEF456...`
- **Health**: ⚠️ Warning (3.1% error rate)
- **Golden Signals**: 📊 Available (Latency: 890ms P95, Traffic: 45 req/min)
- **Language**: Node.js
- **Environment**: production
- **Last Reported**: 1 minute ago

### 📊 checkout-database
- **Type**: DATABASE
- **GUID**: `GHI789...`
- **Health**: ✅ Healthy
- **Golden Signals**: 📉 Limited (Infrastructure metrics only)
- **Environment**: production
- **Last Reported**: 30 seconds ago

## Recommendations
- ⚠️ Investigate checkout-service-web error rate spike
- 💡 Consider enabling APM for checkout-database
- 📈 All entities suitable for dashboard creation
```

## 2. Enhanced Entity Tools

### `get_entity_details`

**Enhanced Features:**
- Comprehensive entity context
- Related entities discovery
- Performance summary
- Instrumentation analysis

**Example Response:**
```markdown
# 🎯 Entity Details: checkout-service

## Basic Information
- **Name**: checkout-service
- **Type**: APPLICATION (APM)
- **Domain**: APM
- **Account**: Production (123456)
- **GUID**: `ABC123...`

## Instrumentation Analysis
- **Agent**: New Relic Java Agent v8.7.0
- **Telemetry Type**: APM (Traditional)
- **OpenTelemetry**: Not detected
- **Last Reported**: 1 minute ago

## Performance Summary (Last Hour)
- **Latency P95**: 234ms (✅ Normal)
- **Throughput**: 127 requests/minute
- **Error Rate**: 1.2% (✅ Normal)
- **Apdex Score**: 0.94

## Related Entities
- **Dependencies**: checkout-database, payment-gateway
- **Dependents**: user-interface, mobile-app
- **Infrastructure**: checkout-host-01, checkout-host-02

## Available Metrics
- ✅ Golden Signals: Complete coverage
- ✅ Custom Attributes: 15 available
- ✅ Error Tracking: Enabled
- ⚠️ Distributed Tracing: Limited

## Recommendations
- 💡 Consider OpenTelemetry migration for better tracing
- 📈 Entity ready for comprehensive dashboard creation
```

### `get_entity_golden_metrics`

**Enhanced Features:**
- Statistical analysis with baselines
- Anomaly detection
- Performance ranking context
- Optimization recommendations

**Example Response:**
```markdown
# 📊 Golden Metrics Analysis

## Entity: checkout-service
**Analysis Period**: Last 1 hour

## Golden Signals Summary

### 🕰️ Latency
- **P50**: 89ms
- **P95**: 234ms (✅ Normal, baseline: 245ms)
- **P99**: 456ms
- **Trend**: ↓ Improving (15% better than baseline)

### 📈 Traffic
- **Rate**: 127 requests/minute
- **Peak**: 145 requests/minute (at 14:23)
- **Trend**: → Stable
- **Pattern**: Normal business hours pattern detected

### ❌ Errors  
- **Rate**: 1.2% (15 errors out of 1,245 requests)
- **Trend**: ↑ Slight increase from 0.8% baseline
- **Types**: TimeoutException (60%), ValidationError (40%)
- **Severity**: 🟡 Low concern

### 💾 Saturation
- **CPU**: 34% average (max: 67%)
- **Memory**: 78% average (max: 89%)
- **Trend**: → Stable
- **Concern**: 🟡 Memory usage approaching threshold

## Analytical Insights
- **Data Quality**: 98% complete, high consistency
- **Seasonality**: Daily pattern detected (confidence: 87%)
- **Anomalies**: None detected in current period
- **Baseline Confidence**: 92%

## Performance Context
- **Ranking**: Top 25% of similar services
- **Peer Comparison**: 23% faster than average
- **Optimization Potential**: Medium

## Recommendations
- ⚠️ Monitor memory usage - approaching 80% threshold
- 🔍 Investigate timeout exceptions in external calls
- ✅ Performance is within expected ranges
- 📈 Ready for dashboard creation
```

## 3. Platform Analysis Tools

### `platform_analyze_adoption`

**Purpose**: Analyze OpenTelemetry adoption across the platform.

**Enhanced Analysis:**
```markdown
# 🔍 OpenTelemetry Adoption Analysis

## Adoption Overview
- **Total Services**: 23
- **OpenTelemetry**: 8 services (35%)
- **APM Agents**: 15 services (65%)
- **Mixed**: 2 services using both

## Migration Progress
**Completed Migrations**:
- user-service (Java → OTEL)
- auth-service (Node.js → OTEL)
- catalog-service (Python → OTEL)

**In Progress**:
- checkout-service (Java, 60% complete)
- payment-service (Node.js, planning phase)

**Recommended Next**:
1. notification-service (Low complexity)
2. recommendation-service (Medium complexity)
3. analytics-service (High complexity)

## Benefits Realized
- **Standardization**: 35% of services using standard telemetry
- **Cross-Service Tracing**: Improved in OTEL services
- **Vendor Independence**: Reduced lock-in for OTEL services

## Migration Blockers
- Legacy dependency on APM-specific features (3 services)
- Custom instrumentation requiring updates (2 services)
- Team training and expertise gaps (4 teams)
```

## 4. Cache Management Tools

### `cache.stats`

**Purpose**: Monitor cache performance and health.

**Example Response:**
```markdown
# 📈 Cache Performance Statistics

## Overall Performance
- **Hit Rate**: 78.3% (target: >75%)
- **Miss Rate**: 21.7%
- **Total Entries**: 342 of 500 max
- **Memory Usage**: 34.2MB of 100MB limit

## Cache Categories

### Discovery Cache
- **Hit Rate**: 85.2%
- **Entries**: 45
- **Average Age**: 12 minutes
- **Status**: ✅ Healthy

### Golden Metrics Cache
- **Hit Rate**: 92.1%
- **Entries**: 128
- **Average Age**: 3 minutes  
- **Status**: ✅ Optimal

### Entity Details Cache
- **Hit Rate**: 67.8%
- **Entries**: 89
- **Average Age**: 8 minutes
- **Status**: ⚠️ Could improve

## Performance Trends
- **Last Hour**: 15% improvement in hit rate
- **Peak Load**: 14:30 (523 requests/minute)
- **Cache Evictions**: 12 (all due to TTL expiry)

## Recommendations
- ✅ Overall performance excellent
- 💡 Consider increasing entity cache TTL
- 📈 Memory usage well within limits
```

### `cache.clear`

**Purpose**: Clear cache entries for fresh data.

**Options:**
- Clear all cache
- Clear specific categories
- Clear expired entries only
- Reset cache statistics

## Enhancement Benefits

### 1. Intelligence Integration
- Schema validation prevents invalid queries
- Discovery-first approach ensures optimal performance
- Context-aware caching reduces API calls

### 2. Improved User Experience
- Rich, structured responses designed for LLM consumption
- Helpful error messages with actionable suggestions
- Performance insights and optimization hints

### 3. Operational Excellence
- Built-in monitoring and health checks
- Proactive cache management
- Comprehensive logging and debugging

### 4. Backward Compatibility
- All existing tool interfaces preserved
- Enhanced features opt-in where possible
- Graceful degradation for missing dependencies

## Migration from Basic Tools

Existing MCP tool usage continues to work without changes:

```javascript
// This continues to work exactly as before
const result = await tools['run_nrql_query']({
  query: 'SELECT count(*) FROM Transaction SINCE 1 hour ago'
});

// But now also includes enhanced features automatically:
// - Schema validation
// - Intelligent caching  
// - Rich error handling
// - Performance insights
```

**Users benefit from enhancements immediately without any code changes.**

---

**Next**: [40_GUIDE_QUICKSTART.md](40_GUIDE_QUICKSTART.md) for getting started with enhanced tools
# Composite Tools Deep Dive

Composite tools combine multiple operations into high-level workflows, providing intelligent automation for common observability tasks.

## Overview

The Enhanced MCP Server includes three main composite tools that leverage discovery intelligence and golden signals analytics:

1. **`discover.environment`** - Complete environment discovery and analysis
2. **`generate.golden_dashboard`** - Intelligent dashboard creation with adaptation
3. **`compare.similar_entities`** - Performance benchmarking and optimization

## 1. Environment Discovery Tool

### Purpose
Provides complete situational awareness of the New Relic environment in a single call, eliminating the need for multiple discovery queries.

### Implementation: `EnvironmentDiscoveryTool`

**Tool Definition:**
```typescript
{
  name: 'discover.environment',
  description: `Discover the complete New Relic environment in one comprehensive call.
  
  🎯 **Purpose**: Provides LLM agents with complete situational awareness of the observability setup.
  
  **Returns**:
  - Complete inventory of monitored entities (services, hosts, etc.)
  - Available telemetry event types and their characteristics  
  - Metric streams and data sources (OpenTelemetry vs APM)
  - Schema guidance for optimal NRQL queries
  - Observability gaps and recommendations
  
  **Use this first** to understand what data is available before using other tools.
  
  **OpenTelemetry Aware**: Automatically detects OTEL vs traditional APM instrumentation.`,
  inputSchema: {
    type: 'object',
    properties: {
      includeHealth: { type: 'boolean', default: false },
      maxEntities: { type: 'number', default: 50, minimum: 10, maximum: 200 },
      forceRefresh: { type: 'boolean', default: false }
    }
  }
}
```

**Discovery Process:**
```typescript
const [
  telemetryContext,    // OpenTelemetry vs APM detection
  entities,           // Monitored services, hosts, etc.
  eventTypes,         // Available telemetry events
  metricStreams,      // Dimensional metrics
] = await Promise.all([
  this.discoverTelemetryContext(),
  this.discoverEntities(maxEntities, includeHealth),
  this.discoverEventTypes(),
  this.discoverMetricStreams(),
]);
```

**Output Structure:**
```typescript
export interface EnvironmentSnapshot {
  entities: EntitySummary[];           // Monitored entities with health
  eventTypes: EventTypeSummary[];      // Available event types with volume
  metricStreams: MetricStreamSummary[]; // Metric categories and examples
  schemaHints: SchemaGuidance;         // Query optimization guidance
  telemetryContext: TelemetryContext;   // OTEL vs APM detection
  observabilityGaps: string[];         // Missing monitoring
  recommendations: string[];           // Actionable improvements
}
```

**Example Response:**
```markdown
# 🔍 New Relic Environment Discovery (🔄 Fresh data)

## 📊 Executive Summary
- **Entities Monitored**: 12 (8 APM, 3 INFRA, 1 BROWSER)
- **Telemetry Sources**: MIXED (OpenTelemetry + New Relic APM)
- **Event Types**: 15 available
- **Metric Streams**: 245 metrics across 6 categories

## 🎯 Schema Guidance for Queries
**Service Identifier**: Use `service.name` to filter by service

**Golden Signal Queries**:
- **Latency**: `percentile(duration.ms, 95) FROM Span WHERE span.kind = "server"`
- **Throughput**: `rate(count(*), 1 minute) FROM Span WHERE span.kind = "server"`
- **Errors**: `filter(count(*), WHERE otel.status_code = "ERROR") FROM Span`

**Strategy**: Use Span events for golden signals (span.kind = "server" for entry spans)
```

**Caching Strategy:**
- 5-minute TTL for complete environment snapshots
- Force refresh option for critical updates
- Cache key includes parameters for different views

### Use Cases

1. **Initial Environment Assessment**
   ```javascript
   // Get complete picture before any other operations
   const env = await tools['discover.environment']({});
   // Use env.schemaHints for optimal queries
   ```

2. **Observability Health Check**
   ```javascript
   const env = await tools['discover.environment']({ includeHealth: true });
   // Review env.observabilityGaps and env.recommendations
   ```

3. **Migration Assessment** 
   ```javascript
   const env = await tools['discover.environment']({});
   // Check env.telemetryContext for OTEL vs APM status
   ```

## 2. Dashboard Generation Tool

### Purpose
Generates comprehensive golden signals dashboards with intelligent adaptation to available telemetry.

### Implementation: `DashboardGenerationTool`

**Tool Definition:**
```typescript
{
  name: 'generate.golden_dashboard',
  description: `Generate a comprehensive golden signals dashboard for any entity in one call.

  🎯 **Purpose**: Creates a production-ready dashboard covering all four golden signals of monitoring.

  **Generated Dashboard Includes**:
  - **Latency**: Response time percentiles (P50, P95, P99) with trend analysis
  - **Traffic**: Request rate and throughput patterns over time
  - **Errors**: Error rate percentage with error breakdown by type
  - **Saturation**: Resource utilization (CPU, Memory) when available

  **Intelligent Adaptation**:
  - Automatically detects OpenTelemetry vs APM instrumentation
  - Adapts queries to use optimal data sources (Span vs Transaction events)
  - Handles mixed telemetry environments gracefully
  - Optimizes widget types based on data characteristics`,
  inputSchema: {
    type: 'object',
    properties: {
      entity_guid: { type: 'string', pattern: '^[A-Za-z0-9+/]+=*$' },
      dashboard_name: { type: 'string' },
      timeframe_hours: { type: 'number', default: 1, minimum: 0.25, maximum: 168 },
      include_saturation: { type: 'boolean', default: true },
      create_dashboard: { type: 'boolean', default: false },
      alert_thresholds: {
        type: 'object',
        properties: {
          latency_p95_ms: { type: 'number', default: 1000 },
          error_rate_percent: { type: 'number', default: 5 },
          traffic_drop_percent: { type: 'number', default: 50 }
        }
      }
    },
    required: ['entity_guid']
  }
}
```

**Generation Process:**
```typescript
// 1. Get comprehensive golden metrics
const goldenMetrics = await this.goldenSignals.getEntityGoldenMetrics(
  entity_guid, 
  Math.max(timeframe_hours * 60, 30)
);

// 2. Generate adaptive dashboard
const dashboardResult = await this.generateGoldenDashboard(
  goldenMetrics,
  { dashboard_name, timeframe_hours, include_saturation, alert_thresholds }
);

// 3. Create dashboard if requested
if (create_dashboard) {
  const createdDashboard = await this.createDashboard(dashboardResult.dashboard);
}
```

**Dashboard Structure:**
```typescript
export interface DashboardDefinition {
  name: string;
  description: string;
  permissions: 'PUBLIC_READ_ONLY' | 'PUBLIC_READ_WRITE';
  pages: Array<{
    name: string;                    // 'Overview', 'Detailed Analysis'
    description: string;
    widgets: DashboardWidget[];
  }>;
  variables: Array<{               // Time range, service filters
    name: string;
    title: string;
    type: string;
    defaultValues: string[];
  }>;
}
```

**Intelligent Adaptations:**

1. **OpenTelemetry Detection:**
   ```typescript
   if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
     return `SELECT percentile(duration.ms, ${percentile}) FROM Span ${entityFilter} AND span.kind = 'server'`;
   } else {
     return `SELECT percentile(duration, ${percentile}) * 1000 FROM Transaction ${entityFilter}`;
   }
   ```

2. **Error Query Adaptation:**
   ```typescript
   if (context.hasOpenTelemetry) {
     return `SELECT filter(count(*), WHERE otel.status_code = 'ERROR') / count(*) * 100 FROM Span`;
   } else {
     return `SELECT filter(count(*), WHERE error IS true) / count(*) * 100 FROM Transaction`;
   }
   ```

**Multi-Page Layout:**

**Page 1: Overview**
- Latency Billboard (P95)
- Error Rate Billboard 
- Traffic Rate Billboard
- Latency Time Series
- Traffic vs Errors Combined

**Page 2: Detailed Analysis**
- Latency Percentiles Breakdown (P50, P95, P99)
- Error Breakdown (by type)
- Resource Saturation (if available)

**Example Response:**
```markdown
# 📊 Golden Signals Dashboard Generated

## 🎯 Entity: my-service
- **Type**: APPLICATION
- **GUID**: `ABC123...`
- **Telemetry**: OTEL

## 📋 Dashboard Overview
- **Name**: Golden Signals: my-service
- **Pages**: 2 (Overview, Detailed Analysis)
- **Widgets**: 8 total
- **Variables**: 2 interactive controls

## 🔧 Intelligent Adaptations
- ✅ Adapted queries for OpenTelemetry Span events

## 💡 Recommendations
- 🔍 Consider enabling infrastructure monitoring for resource saturation metrics
```

### Use Cases

1. **Quick Dashboard Creation**
   ```javascript
   // Preview mode (no actual creation)
   const preview = await tools['generate.golden_dashboard']({
     entity_guid: 'ABC123...',
     create_dashboard: false
   });
   ```

2. **Production Dashboard Deployment**
   ```javascript
   // Create actual dashboard
   const dashboard = await tools['generate.golden_dashboard']({
     entity_guid: 'ABC123...',
     dashboard_name: 'Production Monitoring',
     create_dashboard: true,
     alert_thresholds: {
       latency_p95_ms: 500,
       error_rate_percent: 2
     }
   });
   ```

## 3. Entity Comparison Tool

### Purpose
Compares similar entities to identify performance patterns, outliers, and optimization opportunities.

### Implementation: `EntityComparisonTool`

**Tool Definition:**
```typescript
{
  name: 'compare.similar_entities',
  description: `Compare similar entities to identify performance patterns, outliers, and optimization opportunities.`,
  inputSchema: {
    type: 'object',
    properties: {
      comparison_strategy: {
        type: 'string',
        enum: ['by_type', 'by_name_pattern', 'by_environment', 'by_explicit_list']
      },
      entity_filter: { type: 'string' },
      name_pattern: { type: 'string' },
      environment_tag: { type: 'string' },
      entity_guids: { type: 'array', items: { type: 'string' } },
      timeframe_hours: { type: 'number', default: 1 },
      include_recommendations: { type: 'boolean', default: true }
    }
  }
}
```

**Comparison Process:**
```typescript
// 1. Find comparable entities
const entities = await this.findComparableEntities(strategy, criteria);

// 2. Collect golden metrics for all entities
const metricsCollection = await Promise.all(
  entities.map(entity => 
    this.goldenSignals.getEntityGoldenMetrics(entity.guid, timeframMinutes)
  )
);

// 3. Perform statistical analysis
const analysis = this.performStatisticalAnalysis(metricsCollection);

// 4. Rank and identify outliers
const rankings = this.generatePerformanceRankings(analysis);
```

**Analysis Output:**
```typescript
export interface EntityComparisonResult {
  entities: EntityPerformanceSummary[];
  analysis: {
    latency: StatisticalSummary;
    traffic: StatisticalSummary;
    errors: StatisticalSummary;
    saturation?: StatisticalSummary;
  };
  rankings: {
    bestPerformers: EntitySummary[];
    outliers: EntitySummary[];
    needsAttention: EntitySummary[];
  };
  recommendations: string[];
}
```

**Example Response:**
```markdown
# 📊 Entity Performance Comparison

## 🎯 Comparison Scope
- **Strategy**: by_type
- **Entity Type**: APPLICATION
- **Entities Analyzed**: 8 services
- **Timeframe**: Last 1 hour

## 📈 Performance Analysis

### Latency (P95)
- **Best**: service-A (45ms)
- **Worst**: service-C (1200ms)
- **Average**: 340ms
- **Outliers**: service-C (3.5x above average)

### Error Rates
- **Best**: service-B (0.1%)
- **Worst**: service-D (8.5%)
- **Average**: 2.1%
- **Critical**: service-D exceeds 5% threshold

## 🏆 Performance Rankings
**Best Performers**: service-A, service-B
**Needs Attention**: service-C, service-D

## 💡 Optimization Recommendations
- 🔍 Investigate service-C latency spikes
- 🚨 Address service-D error rate (8.5% vs 2.1% average)
- ✅ Study service-A configuration for best practices
```

### Use Cases

1. **Performance Benchmarking**
   ```javascript
   const comparison = await tools['compare.similar_entities']({
     comparison_strategy: 'by_type',
     entity_filter: 'type = "APPLICATION" AND domain = "APM"'
   });
   ```

2. **Environment Analysis**
   ```javascript
   const envComparison = await tools['compare.similar_entities']({
     comparison_strategy: 'by_environment',
     environment_tag: 'production'
   });
   ```

3. **Outlier Detection**
   ```javascript
   const outliers = await tools['compare.similar_entities']({
     comparison_strategy: 'by_name_pattern',
     name_pattern: 'microservice-*',
     include_recommendations: true
   });
   ```

## Composite Tool Benefits

### 1. Reduced Complexity
**Single call instead of multiple operations:**
- Environment discovery eliminates 5-10 individual queries
- Dashboard generation combines entity lookup, metrics analysis, and creation
- Entity comparison automates discovery, collection, and analysis

### 2. Intelligent Automation
**Built-in intelligence and adaptation:**
- Automatic OpenTelemetry vs APM detection
- Optimal query generation based on available data
- Error handling with graceful degradation

### 3. Actionable Insights
**Rich context and recommendations:**
- Observability gap identification
- Performance optimization suggestions
- Best practice guidance

### 4. LLM-Optimized Output
**Formatted for AI consumption:**
- Structured markdown with clear sections
- Actionable recommendations
- Context-rich descriptions
- Examples and next steps

## Best Practices

### 1. Start with Environment Discovery
```javascript
// Always begin with environment awareness
const env = await tools['discover.environment']({});
// Use env.schemaHints for subsequent operations
```

### 2. Use Preview Mode for Dashboards
```javascript
// Preview before creating
const preview = await tools['generate.golden_dashboard']({
  entity_guid: guid,
  create_dashboard: false
});
// Review and then create if satisfied
```

### 3. Leverage Caching
```javascript
// Use forceRefresh sparingly
const env = await tools['discover.environment']({
  forceRefresh: false // Use cached data when possible
});
```

### 4. Include Health Context
```javascript
// Include health for critical assessments
const env = await tools['discover.environment']({
  includeHealth: true,
  maxEntities: 20
});
```

---

**Next**: [32_TOOLS_ENHANCED.md](32_TOOLS_ENHANCED.md) for enhanced existing tools reference
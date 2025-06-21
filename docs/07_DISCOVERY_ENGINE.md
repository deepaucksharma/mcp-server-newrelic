# Platform Discovery Engine: Technical Deep Dive

The Platform Discovery Engine is the foundation of the MCP Server's zero hardcoded schemas architecture. This document provides comprehensive technical details about discovering New Relic's multi-layered telemetry fabric without any assumptions.

## Overview

The Platform Discovery Engine automatically discovers every schema, field, and relationship at runtime. It makes no assumptions about attribute names, event types, or metric structures, ensuring compatibility with any New Relic account configuration from legacy APM to modern OpenTelemetry.

## Core Philosophy

### Zero Hardcoded Schemas

The Discovery Engine operates on a fundamental principle:

```typescript
// ❌ NEVER assume field names
const errorQuery = "SELECT count(*) FROM Transaction WHERE error = true";

// ✅ ALWAYS discover at runtime
const errorField = await discoverErrorIndicator(accountId);
const serviceField = await discoverServiceIdentifier(accountId);
const errorQuery = buildAdaptiveQuery(errorField, serviceField);
```

### Discovery-First Pattern

Every operation begins with discovery:

1. **Discover what exists** - Event types, attributes, metrics
2. **Profile the data** - Types, cardinality, patterns
3. **Infer semantics** - Service fields, error indicators
4. **Build adaptive queries** - Use discovered fields
5. **Cache intelligently** - Reuse discoveries efficiently

## Architecture

### Discovery Components

```
┌─────────────────────────────────────────────────────────────┐
│               Platform Discovery Engine                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │Event Types  │ Attributes  │  Metrics    │  Entities   │  │
│  │Discovery    │ Profiling   │ Discovery   │  Mapping    │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Heuristic Detection                      │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │  Service    │   Error     │  Golden     │  Platform   │  │
│  │Identifiers  │ Indicators  │  Metrics    │  Features   │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Discovery Cache                          │
│         Event Types (1h) · Metrics (30m) · Entities (15m)   │
└─────────────────────────────────────────────────────────────┘
```

### Discovery Flow

```typescript
// Discovery execution pipeline
async function executeDiscovery(accountId: number): Promise<DiscoveryResult> {
  // 1. Event type discovery
  const eventTypes = await nrql(`SHOW EVENT TYPES`, accountId);
  
  // 2. Attribute profiling for each event type
  const schemas = await Promise.all(
    eventTypes.map(async (eventType) => {
      const attributes = await nrql(
        `SELECT keyset() FROM ${eventType} SINCE 1 hour ago LIMIT 1`,
        accountId
      );
      return { eventType, attributes };
    })
  );
  
  // 3. Metric discovery
  const metrics = await nrql(
    `SELECT uniques(metricName) FROM Metric SINCE 1 hour ago`,
    accountId
  );
  
  // 4. Platform feature detection
  const features = detectPlatformFeatures(schemas, metrics);
  
  return {
    event_types: schemas,
    metrics: metrics,
    platform_features: features,
    discovery_time: new Date().toISOString()
  };
}
```

## Discovery Strategies

### 1. Event Type Discovery

**Process**: Use NRDB's metadata APIs to discover all available event types

```sql
-- Discover all event types
SHOW EVENT TYPES

-- Result example:
-- Transaction, PageView, Log, Span, Metric, CustomEvent
```

**Implementation**:
```typescript
async function discoverEventTypes(accountId: number): Promise<EventTypeInfo[]> {
  const result = await nrql('SHOW EVENT TYPES', accountId);
  
  // Get sample counts for each type
  const eventTypes = await Promise.all(
    result.map(async (type) => {
      const count = await nrql(
        `SELECT count(*) FROM ${type} SINCE 1 hour ago`,
        accountId
      );
      return {
        name: type,
        sample_count: count,
        discovered_at: new Date()
      };
    })
  );
  
  return eventTypes.filter(e => e.sample_count > 0);
}
```

### 2. Attribute Discovery and Profiling

**Process**: Profile attributes without assumptions about names

```typescript
async function profileAttributes(
  eventType: string,
  accountId: number
): Promise<AttributeProfile[]> {
  // Get all attributes
  const attributes = await nrql(
    `SELECT keyset() FROM ${eventType} SINCE 1 hour ago LIMIT 1`,
    accountId
  );
  
  // Profile each attribute
  return Promise.all(
    attributes.map(async (attr) => {
      // Sample values to infer type
      const sample = await nrql(
        `SELECT ${attr} FROM ${eventType} LIMIT 100`,
        accountId
      );
      
      return {
        name: attr,
        type: inferType(sample),
        cardinality: await getCardinality(attr, eventType, accountId),
        coverage: await getCoverage(attr, eventType, accountId),
        sample_values: sample.slice(0, 5)
      };
    })
  );
}
```

### 3. Service Identifier Detection

**Heuristic Chain**: Discover service fields without hardcoding names

```typescript
async function discoverServiceIdentifier(
  schema: SchemaInfo,
  accountId: number
): Promise<FieldDiscovery> {
  const candidates = [];
  
  // Try common patterns (but don't assume they exist)
  const patterns = [
    { regex: /^(app|application|service)[._]?name$/i, confidence: 0.9 },
    { regex: /^(app|application|service)$/i, confidence: 0.8 },
    { regex: /name$/i, confidence: 0.6 }
  ];
  
  for (const attr of schema.attributes) {
    // Check pattern match
    for (const pattern of patterns) {
      if (pattern.regex.test(attr.name)) {
        candidates.push({
          field: attr.name,
          confidence: pattern.confidence,
          reason: 'pattern_match'
        });
      }
    }
    
    // Check cardinality (5-50 unique values suggests service names)
    if (attr.cardinality >= 5 && attr.cardinality <= 50 && attr.type === 'string') {
      candidates.push({
        field: attr.name,
        confidence: 0.7,
        reason: 'cardinality_match'
      });
    }
  }
  
  // Sort by confidence and return best match
  candidates.sort((a, b) => b.confidence - a.confidence);
  return candidates[0] || { field: null, confidence: 0, reason: 'not_found' };
}
```

### 4. Error Indicator Discovery

**Multi-Strategy Detection**: Find error fields across different schemas

```typescript
async function discoverErrorIndicator(
  schema: SchemaInfo,
  accountId: number
): Promise<ErrorFieldDiscovery> {
  const strategies = [
    // Boolean error field
    {
      test: (attr) => attr.name === 'error' && attr.type === 'boolean',
      confidence: 1.0,
      query_pattern: 'WHERE {field} = true'
    },
    // HTTP status codes
    {
      test: (attr) => attr.name.includes('status') && attr.type === 'numeric',
      confidence: 0.8,
      query_pattern: 'WHERE numeric({field}) >= 400'
    },
    // Error class/message
    {
      test: (attr) => attr.name.includes('error') && attr.type === 'string',
      confidence: 0.7,
      query_pattern: 'WHERE {field} IS NOT NULL'
    }
  ];
  
  for (const strategy of strategies) {
    const matching = schema.attributes.find(strategy.test);
    if (matching) {
      return {
        field: matching.name,
        type: matching.type,
        confidence: strategy.confidence,
        query_pattern: strategy.query_pattern.replace('{field}', matching.name)
      };
    }
  }
  
  return { field: null, confidence: 0, query_pattern: null };
}
```

### 5. Dimensional Metrics Discovery

**Process**: Detect dimensional metrics vs event-based metrics

```typescript
async function discoverMetrics(accountId: number): Promise<MetricDiscovery> {
  // Check for dimensional metrics
  const dimensionalMetrics = await nrql(
    `SELECT uniques(metricName) FROM Metric SINCE 1 hour ago`,
    accountId
  );
  
  // Get dimensions for each metric
  const metricsWithDimensions = await Promise.all(
    dimensionalMetrics.map(async (metricName) => {
      const dimensions = await nrql(
        `SELECT keyset() FROM Metric WHERE metricName = '${metricName}' LIMIT 1`,
        accountId
      );
      
      return {
        name: metricName,
        type: 'dimensional',
        dimensions: dimensions.filter(d => !['metricName', 'timestamp'].includes(d))
      };
    })
  );
  
  return {
    has_dimensional_metrics: dimensionalMetrics.length > 0,
    metrics: metricsWithDimensions,
    metric_count: dimensionalMetrics.length
  };
}
```

## Platform Feature Detection

### Adoption Indicators

Detect platform features without assumptions:

```typescript
async function detectPlatformFeatures(
  schemas: SchemaInfo[],
  metrics: MetricInfo[]
): Promise<PlatformFeatures> {
  return {
    // OpenTelemetry detection
    has_opentelemetry: schemas.some(s => 
      s.attributes.some(a => a.name.startsWith('otel.'))
    ),
    
    // Entity synthesis
    uses_entity_synthesis: schemas.some(s =>
      s.attributes.some(a => a.name === 'entity.guid')
    ),
    
    // Dimensional metrics
    has_dimensional_metrics: metrics.length > 0,
    
    // APM data
    has_apm_data: schemas.some(s => s.name === 'Transaction'),
    
    // Custom events
    has_custom_events: schemas.some(s => 
      !['Transaction', 'PageView', 'Log', 'Span', 'Metric'].includes(s.name)
    )
  };
}
```

## Caching Strategy

### Cache Hierarchy

```typescript
interface CacheConfig {
  event_types: { ttl: 3600, maxSize: 100 },      // 1 hour
  attributes: { ttl: 1800, maxSize: 1000 },      // 30 minutes
  metrics: { ttl: 1800, maxSize: 500 },          // 30 minutes
  entities: { ttl: 900, maxSize: 200 },          // 15 minutes
  discoveries: { ttl: 300, maxSize: 50 }         // 5 minutes
}

class DiscoveryCache {
  private cache = new Map<string, CachedDiscovery>();
  
  getCacheKey(accountId: number, type: string): string {
    return `discovery:${accountId}:${type}`;
  }
  
  async get<T>(key: string): Promise<T | null> {
    const cached = this.cache.get(key);
    if (!cached) return null;
    
    if (Date.now() > cached.expires) {
      this.cache.delete(key);
      return null;
    }
    
    return cached.data as T;
  }
}
```

## Performance Optimization

### Query Optimization

```sql
-- Optimized discovery query using NRDB metadata
SELECT 
  eventType(),
  count(*) as sample_count,
  earliest(timestamp) as oldest_data,
  latest(timestamp) as newest_data
FROM (
  FROM Transaction SELECT * LIMIT 1 UNION
  FROM Log SELECT * LIMIT 1 UNION
  FROM Span SELECT * LIMIT 1
)
FACET eventType()
SINCE 7 days ago
```

### Parallel Discovery

```typescript
async function parallelDiscovery(accountId: number): Promise<CompleteDiscovery> {
  // Execute discoveries in parallel
  const [eventTypes, metrics, entities] = await Promise.all([
    discoverEventTypes(accountId),
    discoverMetrics(accountId),
    discoverEntities(accountId)
  ]);
  
  // Profile attributes in parallel batches
  const batchSize = 5;
  const attributeProfiles = [];
  
  for (let i = 0; i < eventTypes.length; i += batchSize) {
    const batch = eventTypes.slice(i, i + batchSize);
    const profiles = await Promise.all(
      batch.map(et => profileAttributes(et.name, accountId))
    );
    attributeProfiles.push(...profiles);
  }
  
  return {
    event_types: eventTypes,
    attributes: attributeProfiles,
    metrics: metrics,
    entities: entities,
    platform_features: detectPlatformFeatures(attributeProfiles, metrics)
  };
}
```

## Integration with Tools

### Tool Enhancement Pattern

```typescript
// How tools use discovery
async function enhancedRunNrqlQuery(params: QueryParams): Promise<QueryResult> {
  // 1. Extract referenced event types
  const eventTypes = extractEventTypes(params.query);
  
  // 2. Check discovery cache
  const discovery = await getDiscovery(params.account_id);
  
  // 3. Validate event types exist
  const unknownTypes = eventTypes.filter(
    type => !discovery.event_types.includes(type)
  );
  
  if (unknownTypes.length > 0) {
    return {
      error: 'UNKNOWN_EVENT_TYPES',
      suggestion: 'Run discover_schemas first',
      unknown_types: unknownTypes,
      known_types: discovery.event_types
    };
  }
  
  // 4. Execute query
  return executeNrqlQuery(params);
}
```

### Adaptive Dashboard Generation

```typescript
// Dashboard generation using discovery
async function generateDashboard(
  template: string,
  entityGuid: string,
  discovery: Discovery
): Promise<Dashboard> {
  const widgets = [];
  
  // Error rate widget
  const errorField = await discoverErrorIndicator(discovery.schemas);
  if (errorField.field) {
    widgets.push(createErrorWidget(errorField));
  }
  
  // Latency widget
  const latencyField = await discoverLatencyField(discovery.schemas);
  if (latencyField.field) {
    widgets.push(createLatencyWidget(latencyField));
  }
  
  // Adapt to what's available
  return {
    name: `${template} Dashboard`,
    widgets: widgets,
    metadata: {
      adapted_fields: {
        error: errorField.field,
        latency: latencyField.field
      },
      discovery_confidence: calculateConfidence(widgets)
    }
  };
}
```

## Monitoring and Observability

### Discovery Metrics

```typescript
// Self-monitoring
interface DiscoveryMetrics {
  discovery_latency_ms: Histogram;
  cache_hit_rate: Gauge;
  discovery_errors: Counter;
  confidence_scores: Histogram;
  schema_changes: Counter;
}

// Emit metrics
function emitDiscoveryMetrics(operation: string, duration: number, success: boolean) {
  metrics.discovery_latency_ms.observe({ operation }, duration);
  metrics.discovery_errors.inc({ operation }, success ? 0 : 1);
}
```

## Best Practices

### For Developers

1. **Never Hardcode Field Names**
   ```typescript
   // Always discover, never assume
   const serviceField = await discoverServiceIdentifier(schema);
   const query = `SELECT count(*) FROM Transaction WHERE ${serviceField} = 'checkout'`;
   ```

2. **Handle Discovery Failures**
   ```typescript
   const discovery = await tryDiscover(accountId);
   if (!discovery.success) {
     // Use cached discovery or provide manual option
     return useFallbackDiscovery(accountId);
   }
   ```

3. **Respect Confidence Scores**
   ```typescript
   if (discovery.confidence < 0.7) {
     // Provide alternatives or request confirmation
     return {
       primary: discovery.result,
       alternatives: discovery.alternatives,
       confidence_warning: true
     };
   }
   ```

## Conclusion

The Platform Discovery Engine enables the MCP Server to work with any New Relic account configuration without manual setup or maintenance. By discovering everything at runtime and making zero hardcoded assumptions, it delivers on the promise of truly adaptive platform intelligence.

This approach ensures that dashboards, queries, and analyses work immediately in any environment and continue working as schemas evolve - the foundation of platform-native observability.
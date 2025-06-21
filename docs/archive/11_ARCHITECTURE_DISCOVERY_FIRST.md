# Discovery-First Architecture

This document specifies the discovery-first design philosophy that forms the foundation of the New Relic MCP Server architecture.

## Table of Contents

1. [Philosophy](#philosophy)
2. [Core Principles](#core-principles)
3. [Discovery Engine Architecture](#discovery-engine-architecture)
4. [Discovery Patterns](#discovery-patterns)
5. [Data Flow](#data-flow)
6. [Caching Strategy](#caching-strategy)
7. [Implementation Details](#implementation-details)
8. [Performance Considerations](#performance-considerations)
9. [Examples](#examples)
10. [Future Enhancements](#future-enhancements)

## Philosophy

The discovery-first architecture SHALL operate on the fundamental principle that **no assumptions about data structure, schema, or content should ever be made**. Instead, the system SHALL always discover the actual state of data before operating on it.

### Traditional vs Discovery-First

```
Traditional Approach:              Discovery-First Approach:
1. Assume schema                   1. Discover available schemas
2. Write query                     2. Explore schema attributes  
3. Execute                         3. Profile data characteristics
4. Handle failures                 4. Build informed query
5. Retry with fixes               5. Execute with confidence
```

### Benefits

1. **Universal Compatibility**: Works with any New Relic account configuration
2. **Zero Configuration**: No schema definitions or mappings required
3. **Adaptive Intelligence**: Automatically adapts to schema changes
4. **Error Reduction**: Eliminates schema mismatch errors
5. **Hidden Insights**: Discovers unknown relationships and patterns

## Core Principles

### 1. Never Assume

The system SHALL NOT make assumptions about:
- Event type names or existence
- Attribute names or types
- Data relationships
- Value formats or ranges
- Schema stability

### 2. Always Explore

Before any operation, the system SHALL:
- Enumerate available event types
- Discover attribute sets
- Profile data distributions
- Identify relationships
- Assess data quality

### 3. Cache Intelligently

Discovery results SHALL be cached with appropriate TTLs:
- Schema structure: Long TTL (hours/days)
- Data profiles: Medium TTL (minutes/hours)
- Relationships: Long TTL with validation
- Quality metrics: Short TTL (minutes)

### 4. Validate Continuously

The system SHALL validate cached discoveries:
- Check schema drift
- Monitor new attributes
- Detect relationship changes
- Track quality degradation

### 5. Adapt Dynamically

Based on discoveries, the system SHALL:
- Adjust query strategies
- Optimize data access patterns
- Update cached knowledge
- Refine recommendations

## Discovery Engine Architecture

```
┌─────────────────────────────────────────────────────┐
│                 Discovery Engine                     │
├─────────────────┬───────────────┬───────────────────┤
│ Schema Discovery│ Data Profiling │ Relationship      │
│    Service      │    Service     │ Discovery Service │
├─────────────────┴───────────────┴───────────────────┤
│              Discovery Orchestrator                  │
├─────────────────────────────────────────────────────┤
│              Discovery Cache Layer                   │
├─────────────────────────────────────────────────────┤
│                 NRDB Interface                       │
└─────────────────────────────────────────────────────┘
```

### Schema Discovery Service

**Purpose**: Enumerate and catalog all available event types and their structures.

**Operations**:
```go
type SchemaDiscoveryService interface {
    // List all event types with basic metadata
    ListEventTypes(ctx context.Context) ([]EventType, error)
    
    // Get detailed schema for an event type
    GetSchema(ctx context.Context, eventType string) (*Schema, error)
    
    // Discover custom attributes
    DiscoverAttributes(ctx context.Context, eventType string) ([]Attribute, error)
    
    // Track schema changes over time
    GetSchemaHistory(ctx context.Context, eventType string) ([]SchemaVersion, error)
}
```

### Data Profiling Service

**Purpose**: Analyze data characteristics and distributions.

**Capabilities**:
1. **Statistical Analysis**: Min, max, mean, percentiles
2. **Cardinality Analysis**: Unique values, frequency distributions
3. **Pattern Detection**: Regular expressions, formats
4. **Completeness Analysis**: Null ratios, data gaps
5. **Temporal Analysis**: Time-based patterns

### Relationship Discovery Service

**Purpose**: Identify connections between different data sources.

**Methods**:
1. **Common Attributes**: Find shared fields across event types
2. **Value Correlation**: Detect correlated values
3. **Temporal Correlation**: Time-based relationships
4. **Graph Analysis**: Build relationship graphs
5. **Join Key Detection**: Identify potential join fields

## Discovery Patterns

### Pattern 1: Breadth-First Discovery
```
1. List all event types
2. Sample each event type
3. Build overview catalog
4. Deep dive on demand
```

### Pattern 2: Depth-First Discovery
```
1. Focus on specific event type
2. Complete attribute analysis
3. Full statistical profiling
4. Relationship mapping
```

### Pattern 3: Guided Discovery
```
1. User indicates interest area
2. Targeted discovery in that domain
3. Expand to related areas
4. Build focused knowledge graph
```

### Pattern 4: Incremental Discovery
```
1. Start with basic schema
2. Profile as queries execute
3. Learn from usage patterns
4. Refine understanding over time
```

## Data Flow

### Discovery Request Flow

```
Client Request
    │
    ▼
Cache Check ──────► HIT ──────► Return Cached
    │                              │
    │ MISS                         │
    ▼                              │
Discovery Engine                   │
    │                              │
    ├─► Schema Service            │
    ├─► Profile Service           │
    └─► Relationship Service      │
         │                         │
         ▼                         │
    NRDB Queries                  │
         │                         │
         ▼                         │
    Process Results               │
         │                         │
         ▼                         │
    Update Cache                  │
         │                         │
         ▼                         ▼
    Return Results ◄──────────────┘
```

## Caching Strategy

### Cache Hierarchy

```
┌─────────────────────────────────────┐
│         Request Cache               │ TTL: 30s
│      (Per-request memoization)      │
├─────────────────────────────────────┤
│         Session Cache               │ TTL: 30min
│    (Per-session discoveries)        │
├─────────────────────────────────────┤
│       Application Cache             │ TTL: 2hr
│   (Shared across sessions)          │
├─────────────────────────────────────┤
│      Distributed Cache              │ TTL: 24hr
│  (Shared across instances)          │
└─────────────────────────────────────┘
```

### Cache Key Design

```
Schema Cache:
  Key: "schema:{account}:{eventType}:{version}"
  TTL: 24 hours

Profile Cache:
  Key: "profile:{account}:{eventType}:{attribute}:{timeRange}"
  TTL: 2 hours

Relationship Cache:
  Key: "rel:{account}:{eventType1}:{eventType2}"
  TTL: 24 hours
```

### Invalidation Rules

1. **Time-based**: Automatic expiration via TTL
2. **Event-based**: Invalidate on schema change detection
3. **Volume-based**: Invalidate when data volume changes significantly
4. **Manual**: API endpoint for forced refresh

## Implementation Details

### Discovery Query Examples

#### Schema Discovery Query
```sql
SHOW EVENT TYPES 
SINCE 24 hours ago
```

#### Attribute Discovery Query
```sql
SELECT keyset() 
FROM EventType 
SINCE 1 hour ago 
LIMIT 1
```

#### Data Profiling Query
```sql
SELECT 
  count(*) as total,
  uniqueCount(attribute) as unique_values,
  min(attribute) as min_value,
  max(attribute) as max_value,
  average(attribute) as avg_value,
  percentile(attribute, 50, 90, 95, 99) as percentiles
FROM EventType 
SINCE 7 days ago
WHERE attribute IS NOT NULL
```

### Parallel Discovery

The system SHALL execute discovery operations in parallel when possible:

```go
type ParallelDiscovery struct {
    Workers   int
    Timeout   time.Duration
    BatchSize int
}

func (pd *ParallelDiscovery) DiscoverAll(ctx context.Context) {
    // Parallel execution pattern
    eventTypes := make(chan string, 100)
    results := make(chan DiscoveryResult, 100)
    
    // Start workers
    for i := 0; i < pd.Workers; i++ {
        go pd.worker(ctx, eventTypes, results)
    }
    
    // Distribute work
    // Collect results
}
```

## Performance Considerations

### Query Optimization

1. **Sampling Strategy**
   - Use LIMIT for initial discovery
   - Increase sample size for detailed profiling
   - Time-based sampling for large datasets

2. **Batch Processing**
   - Group similar discoveries
   - Combine attribute queries
   - Bulk cache updates

3. **Resource Management**
   - Limit concurrent discoveries
   - Implement circuit breakers
   - Monitor memory usage

### Performance Metrics

The system SHALL track discovery performance:

```
discovery_duration_seconds{operation="schema", event_type="Transaction"}
discovery_cache_hit_ratio{cache_level="application"}
discovery_attributes_discovered{event_type="Transaction"}
discovery_query_count{operation="profile"}
```

## Examples

### Example 1: First-Time User Flow

```
User: "Show me my application performance"

System Discovery Flow:
1. discover.explore_event_types()
   → Finds: Transaction, TransactionError, Span, Metric

2. discover.explore_attributes("Transaction")
   → Finds: appName, duration, error, host, etc.

3. discover.profile_attribute("Transaction", "appName")
   → Finds: ["web-app", "api-service", "batch-processor"]

4. query_nrdb("SELECT average(duration) FROM Transaction FACET appName")
   → Returns: Accurate results based on discovered schema
```

### Example 2: Cross-Event Discovery

```
User: "Find relationships between my data"

System Discovery Flow:
1. List all event types
2. Find common attributes (e.g., appId, host, traceId)
3. Test correlations between events
4. Build relationship graph
5. Suggest join strategies
```

## Future Enhancements

### Machine Learning Integration

1. **Pattern Learning**: Learn common query patterns
2. **Predictive Discovery**: Pre-discover likely needed schemas
3. **Anomaly Detection**: Identify unusual schema changes
4. **Smart Caching**: ML-driven cache management

### Advanced Discovery Features

1. **Schema Evolution Tracking**: Version control for schemas
2. **Discovery Subscriptions**: Real-time schema change notifications
3. **Discovery API**: Expose discovery as a service
4. **Visual Discovery**: Graphical schema exploration

### Performance Improvements

1. **Discovery Index**: Pre-computed discovery results
2. **Incremental Discovery**: Delta-based updates
3. **Distributed Discovery**: Federated discovery across regions
4. **Discovery Pipeline**: Stream processing for continuous discovery

## Conclusion

The discovery-first architecture ensures that the MCP Server can work with any New Relic account configuration without prior knowledge or configuration. This approach provides flexibility, reduces errors, and enables intelligent automation that adapts to each user's unique data landscape.

## Related Documentation

- [Architecture Overview](10_ARCHITECTURE_OVERVIEW.md) - System architecture
- [State Management](12_ARCHITECTURE_STATE_MANAGEMENT.md) - Caching and state
- [Discovery Tools](31_TOOLS_DISCOVERY.md) - Discovery tool reference
- [Concepts](04_CONCEPTS.md) - Core concepts including discovery-first

---

**Design Rationale**: For detailed reasoning behind discovery-first design decisions, see [Architecture Decision Records](19_ARCHITECTURE_DECISIONS.md).
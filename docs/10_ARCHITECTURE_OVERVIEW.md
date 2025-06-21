# Architecture Overview

High-level architecture and components of the Enhanced MCP Server New Relic, designed for intelligent observability with zero hardcoded schemas.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Enhanced MCP Server                          │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Enhanced Tool  │  │  Composite Tool │  │  Platform Tool  │  │
│  │    Registry     │  │    Registry     │  │    Registry     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Platform        │  │ Golden Signals  │  │ Intelligent     │  │
│  │ Discovery       │  │ Engine          │  │ Cache           │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                     NerdGraph Client                            │
├─────────────────────────────────────────────────────────────────┤
│                   MCP Protocol Layer                            │
│                   (STDIO Transport)                             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      New Relic Platform                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   NerdGraph     │  │      NRDB       │  │    Entities     │  │
│  │   GraphQL       │  │   (Database)    │  │   & Metrics     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Enhanced Tool Registry (`EnhancedToolRegistry`)

The central component that manages all MCP tools with intelligence and caching.

**Key Features:**
- **Discovery-First Validation**: Validates schemas before query execution
- **Intelligent Caching**: Context-aware caching with adaptive TTL
- **Error Handling**: Comprehensive error responses with suggestions
- **Performance Optimization**: Parallel execution and background refresh

**Registered Tools:**
- Enhanced existing tools (`run_nrql_query`, `search_entities`, etc.)
- Composite intelligence tools (`discover.environment`, `generate.golden_dashboard`, etc.)
- Platform analysis tools (`platform_analyze_adoption`)
- Cache management tools (`cache.stats`, `cache.clear`)

### 2. Platform Discovery (`PlatformDiscovery`)

Core discovery engine that implements the "discover-first, assume-nothing" philosophy.

**Capabilities:**
- **Event Type Discovery**: Finds all available event types with volume analysis
- **Attribute Profiling**: Analyzes field characteristics and patterns
- **Metric Discovery**: Identifies dimensional metrics and their dimensions
- **Service Identification**: Discovers service identifier fields
- **Error Pattern Detection**: Finds error indicators and patterns

**Caching Strategy:**
- Schema discovery: 4 hours TTL
- Attribute profiling: 30 minutes TTL
- Service identifiers: 2 hours TTL
- Error patterns: 30 minutes TTL

### 3. Golden Signals Engine (`GoldenSignalsEngine`)

Implements the four golden signals of monitoring with OpenTelemetry awareness.

**Golden Signals:**
- **Latency**: Response time percentiles (P50, P95, P99)
- **Traffic**: Request rate and throughput patterns
- **Errors**: Error rate with anomaly detection
- **Saturation**: Resource utilization (CPU, Memory)

**Intelligence Features:**
- **Analytical Metadata**: Data quality assessment, seasonality detection
- **Anomaly Detection**: Statistical analysis with Z-score based detection
- **Baseline Establishment**: Median-based baselines with confidence intervals
- **OpenTelemetry Awareness**: Automatic detection of OTEL vs APM instrumentation

### 4. Intelligent Cache (`IntelligentCache`)

Advanced caching system with context-aware freshness strategies.

**Cache Strategies:**
```typescript
{
  discovery: { ttl: 5min, adaptive: true, priority: 'high' },
  goldenMetrics: { ttl: 2min, adaptive: true, priority: 'critical' },
  entityDetails: { ttl: 10min, adaptive: false, priority: 'medium' },
  dashboards: { ttl: 15min, adaptive: false, priority: 'low' },
  analytics: { ttl: 30min, adaptive: true, priority: 'medium' }
}
```

**Features:**
- **Adaptive TTL**: Adjusts cache duration based on access patterns
- **Background Refresh**: Proactive cache warming for critical data
- **Health Monitoring**: Cache performance analysis and optimization
- **LRU Eviction**: Memory management with least-recently-used eviction

### 5. Composite Tools

High-level tools that combine multiple operations for complex workflows.

#### Environment Discovery Tool (`EnvironmentDiscoveryTool`)
- **Purpose**: One-call comprehensive environment analysis
- **Output**: Complete observability landscape with gaps and recommendations
- **Intelligence**: OpenTelemetry vs APM detection, schema guidance

#### Dashboard Generation Tool (`DashboardGenerationTool`)
- **Purpose**: Intelligent dashboard creation with automatic adaptation
- **Features**: Multi-page layouts, alert thresholds, preview mode
- **Adaptation**: Automatic query optimization based on telemetry context

#### Entity Comparison Tool (`EntityComparisonTool`)
- **Purpose**: Performance benchmarking and outlier detection
- **Analysis**: Statistical comparison with performance rankings
- **Output**: Optimization recommendations and best practices

## Data Flow Architecture

### 1. Discovery-First Flow

```
User Request → Schema Validation → Cache Check → API Call → Cache Store → Response
     ↓              ↓                  ↓           ↓           ↓           ↓
  Tool Call → Discovery Engine → Intelligent → NerdGraph → Smart Cache → Enhanced
                                   Cache       Client      Update      Response
```

### 2. Cache-First Flow

```
Subsequent Request → Cache Hit → Freshness Check → Background Refresh (if needed)
        ↓               ↓             ↓                      ↓
    Same Tool → Intelligent → Assessment → Proactive Update (optional)
                  Cache       (fresh/stale)
```

### 3. Composite Tool Flow

```
Composite Tool → Multiple Discovery → Parallel API Calls → Analysis Engine → Combined Response
      ↓               ↓                      ↓                   ↓              ↓
Environment → [Event Types,        → [NRQL Queries,     → Golden Signals → Rich Markdown
Discovery     Entities,              GraphQL Requests]     Analytics        Response
              Metrics]
```

## Technical Implementation

### MCP Protocol Integration

**Server Setup:**
```typescript
const server = new Server({
  name: 'mcp-server-newrelic-platform',
  version: '2.0.0'
}, {
  capabilities: {
    tools: {},
    resources: {}
  }
});
```

**Tool Registration:**
```typescript
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;
  return registry.handleToolCall(name, args);
});

server.setRequestHandler(ListToolsRequestSchema, async () => {
  return { tools: registry.getAllToolDefinitions() };
});
```

### NerdGraph Client

**Configuration:**
```typescript
const nerdgraph = createNerdGraphClient({
  apiKey: config.newrelic.apiKey,
  region: config.newrelic.region, // US or EU
  logger: logger
});
```

**Query Execution:**
```typescript
// NRQL queries
const result = await nerdgraph.nrql(accountId, query);

// GraphQL queries  
const result = await nerdgraph.request(query, variables);
```

### Configuration System

**Environment-Based Configuration:**
```typescript
interface Config {
  newrelic: {
    apiKey: string;
    accountId: string;
    region: 'US' | 'EU';
    graphqlUrl?: string;
  };
  discovery: {
    cache: { ttl: { schemas: number; attributes: number; } };
    confidence: { minimum: number; optimal: number; };
  };
  mcp: { transport: 'stdio'; };
}
```

## Scalability Considerations

### Memory Management
- **Cache Size Limits**: 500 entries max with LRU eviction
- **Background Cleanup**: Automatic maintenance every 5 minutes
- **Memory Monitoring**: Built-in memory usage tracking

### Performance Optimization
- **Parallel Execution**: Multiple API calls in parallel where possible
- **Request Batching**: Combine related requests to reduce API calls
- **Smart Caching**: Reduce redundant API calls through intelligent caching
- **Background Refresh**: Proactive cache updates for critical data

### API Rate Limiting
- **Intelligent Throttling**: Respect New Relic API rate limits
- **Cache-First Strategy**: Minimize API calls through aggressive caching
- **Error Handling**: Graceful degradation when rate limits are hit

## Security Architecture

### API Key Management
- **Environment Variables**: Secure credential storage
- **No Persistence**: API keys never stored on disk
- **Regional Support**: Automatic endpoint selection based on region

### Data Privacy
- **Memory-Only Cache**: No persistent data storage
- **Automatic Cleanup**: Cache entries expire automatically
- **No Customer Data Logging**: Only metadata and performance data logged

## Error Handling Strategy

### Graceful Degradation
- **Cache Fallback**: Return stale data when API calls fail
- **Partial Results**: Return available data even when some operations fail
- **Helpful Suggestions**: Provide actionable error messages

### Error Categories
1. **Authentication Errors**: Invalid API keys or permissions
2. **Schema Errors**: Unknown event types or attributes  
3. **Rate Limiting**: API quota exceeded
4. **Network Errors**: Connectivity issues
5. **Cache Errors**: Memory or performance issues

## Monitoring and Observability

### Built-in Metrics
- **Cache Performance**: Hit rates, memory usage, response times
- **API Performance**: Query execution times, error rates
- **Discovery Success**: Schema coverage, confidence scores

### Health Checks
- **Cache Health**: Memory usage, hit rates, staleness
- **API Health**: Connectivity, response times, error rates
- **Discovery Health**: Coverage, confidence, freshness

## Development Architecture

### Module Structure
```
src/
├── core/                    # Core platform services
│   ├── platform-discovery.ts
│   ├── golden-signals.ts
│   ├── intelligent-cache.ts
│   └── types.ts
├── tools/                   # Tool implementations
│   ├── enhanced-registry.ts
│   ├── environment-discovery.ts
│   ├── dashboard-generation.ts
│   └── entity-comparison.ts
├── adapters/               # External integrations
│   └── nerdgraph.ts
└── index.ts               # Main server entry point
```

### TypeScript Configuration
- **Strict Mode**: Full TypeScript strict checking enabled
- **ESM Modules**: Native ES module support
- **Type Safety**: Comprehensive type definitions for all APIs

---

**Next**: [11_ARCHITECTURE_DISCOVERY_FIRST.md](11_ARCHITECTURE_DISCOVERY_FIRST.md) for discovery-first design philosophy
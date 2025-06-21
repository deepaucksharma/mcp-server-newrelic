# Platform-Native Architecture

This document defines the architecture of the platform-native MCP Server for New Relic. The architecture emphasizes zero hardcoded schemas, enhanced existing tools, and adaptive intelligence for cross-account platform analysis.

## Table of Contents

1. [Core Philosophy](#core-philosophy)
2. [System Architecture](#system-architecture)
3. [Platform Discovery Engine](#platform-discovery-engine)
4. [Enhanced Tool Architecture](#enhanced-tool-architecture)
5. [Adaptive Widget System](#adaptive-widget-system)
6. [Cross-Account Intelligence](#cross-account-intelligence)
7. [Implementation Phases](#implementation-phases)
8. [Performance Considerations](#performance-considerations)

## Core Philosophy

### Zero Hardcoded Schemas

Every schema, field, and relationship is discovered at runtime. The server makes no assumptions about:
- Event type names (Transaction, Span, Log, etc.)
- Attribute names (appName, service.name, app.name)
- Metric structures (dimensional vs event-based)
- Error indicators (boolean error, http.status_code, error.class)

### Platform-Native Intelligence

The server enhances existing new-branch tools with:
- Rich metadata for AI orchestration
- Discovery suggestions when schemas are unknown
- Adaptive behavior based on discovered data
- Cross-account analysis capabilities

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP Clients                              │
│         (Claude Desktop, AI Agents, CLI Tools)              │
└─────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│              MCP Server (Platform-Native)                   │
├─────────────────────────────────────────────────────────────┤
│                    src/index.ts                             │
│         Server setup with @modelcontextprotocol/sdk         │
├─────────────────────────────────────────────────────────────┤
│                 src/tools/registry.ts                       │
│  ┌────────────────────┬────────────────────────────────┐   │
│  │ Enhanced Existing  │    New Platform Tools         │   │
│  │ • run_nrql_query   │ • discover_schemas            │   │
│  │ • search_entities  │ • dashboard_generate          │   │
│  │ • get_entity      │ • platform_analyze_adoption   │   │
│  └────────────────────┴────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│               src/core/discovery.ts                         │
│           Platform Discovery Engine                         │
│  ┌─────────────┬──────────────┬────────────────────────┐   │
│  │Event Types  │ Attributes   │ Metrics & Dimensions   │   │
│  │Discovery    │ Profiling    │ Discovery              │   │
│  └─────────────┴──────────────┴────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│            src/core/adaptive-widgets.ts                     │
│         Adaptive Dashboard Generation                       │
└─────────────────────────────────────────────────────────────┘
```

## Platform Discovery Engine

The Platform Discovery Engine dynamically learns about the platform's structure without any hardcoded assumptions.

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
│  ┌─────────────────────────────────────────────────────────┐  │
│  │           Discovery Cache (Map-based)                   │  │
│  │    • Event schemas cached for 1 hour                    │  │
│  │    • Metric names cached for 30 minutes                 │  │
│  │    • Entity data cached for 15 minutes                  │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Event Type Discovery

Discovers all available event types without assumptions:

**Process:**
1. Execute `SHOW EVENT TYPES` to list all types
2. Sample each event type for structure
3. Profile attributes with keyset()
4. Calculate sample counts and coverage

**No Hardcoded Assumptions:**
- Doesn't assume Transaction exists
- Doesn't assume specific event types
- Works with custom events
- Adapts to any schema

### Attribute Discovery

Profiles attributes dynamically:

**Process:**
1. Use keyset() to find all attributes
2. Sample data to infer types
3. Calculate cardinality and uniqueness
4. Identify service/error indicators heuristically

**Heuristic Detection:**
- Service fields: 5-50 unique values, string type
- Error fields: boolean type or status codes
- Metric fields: numeric with consistent units

### Metric Discovery

Finds dimensional metrics:

**Process:**
1. Query `SELECT uniques(metricName) FROM Metric`
2. Extract dimensions for each metric
3. Map metrics to related events
4. Identify metric units from names
## Enhanced Tool Architecture

The platform-native approach enhances existing tools with discovery intelligence:

### Tool Enhancement Pattern

```
Original Tool                 Enhanced Tool
│                            │
├─ Basic parameters          ├─ Rich parameter schemas
├─ Simple execution          ├─ Discovery validation
├─ Raw results              ├─ Contextual results
└─ No guidance              └─ AI-optimized examples
```

### Enhanced run_nrql_query

**Enhancements:**
- Validates event types exist before execution
- Suggests `discover_schemas` if unknown types referenced
- Provides rich examples for common patterns
- Includes query performance metadata

**Metadata:**
```yaml
category: query
costIndicator: low
examples:
  - description: "Get error rate for a service"
    query: "SELECT percentage(count(*), WHERE error = true) FROM Transaction"
  - description: "Get top 10 slowest transactions"
    query: "SELECT average(duration) FROM Transaction FACET name LIMIT 10"
```

### Enhanced search_entities

**Enhancements:**
- Guides with discovered entity types
- Maps relationships automatically
- Enriches with golden metrics
- Supports pagination with cursor

**Entity Domains:**
- APM: Application services
- BROWSER: Browser applications
- INFRA: Infrastructure hosts
- SYNTH: Synthetic monitors
- NR1: Dashboards and workloads

### New discover_schemas Tool

**Purpose:** First tool to run in any account

**Returns:**
- All event types with sample counts
- Attribute names and types
- Dimensional metrics if present
- Platform adoption summary

## Adaptive Widget System

The adaptive widget system creates dashboards that work with any schema:

### Widget Adaptation Process

```
1. Discover available fields
2. Find best match for widget intent
3. Build query using discovered fields
4. Handle fallbacks gracefully
```

### Example: Error Rate Widget

**Intent**: Show error rate
**Discovery Process**:
1. Find error indicators (error, http.status_code, error.class)
2. Find service identifier (appName, service.name)
3. Build appropriate query

**Adaptive Queries**:
- Boolean: `WHERE error = true`
- Status: `WHERE numeric(http.status_code) >= 400`
- String: `WHERE error.class IS NOT NULL`

### Dashboard Templates

**Golden Signals Template**:
- Error rate (adapts to error indicators)
- Latency (uses metrics or events)
- Throughput (count or rate based)
- Saturation (CPU, memory, custom)

**Infrastructure Template**:
- CPU utilization
- Memory usage
- Disk I/O
- Network traffic

**Custom Templates**:
- User-defined widget intents
- Automatic field discovery
- Best-effort query building

## Cross-Account Intelligence

Platform analysis capabilities across multiple accounts:

### Adoption Analysis

**Metrics Tracked**:
- Dimensional metrics vs events ratio
- OpenTelemetry adoption indicators
- Entity synthesis usage
- Dashboard complexity

**Process**:
1. Discover schemas per account
2. Identify platform features used
3. Calculate adoption scores
4. Generate comparison reports

### Migration Planning

**Capabilities**:
- Schema comparison between accounts
- Field mapping identification
- Compatibility validation
- Migration guide generation

## Implementation Phases

### Phase 1: Enhanced Tool Metadata (Week 1)
- Add rich descriptions and examples to existing tools
- Implement input/output schemas with Zod
- Add AI-optimized metadata for tool chaining
- Create discovery caching layer

### Phase 2: Discovery Engine (Week 2-3)
- Build schema discovery without hardcoded types
- Implement metric vs event detection
- Add entity relationship mapping
- Create heuristic field detection

### Phase 3: Dashboard Generation (Week 4)
- Create adaptive widget templates
- Build dashboard composition engine
- Add dry-run and preview capabilities
- Implement template library

### Phase 4: Platform Intelligence (Week 5-6)
- Cross-account analysis tools
- Adoption scoring algorithms
- Migration planning tools
- Platform usage insights

## Performance Considerations

### Discovery Caching

**Cache Strategy**:
- Event types: 1 hour TTL
- Metrics: 30 minute TTL
- Entity data: 15 minute TTL
- Use Map-based memory cache

### Query Optimization

**Optimization Techniques**:
- Validate schemas before execution
- Cache successful query patterns
- Implement query timeouts
- Batch similar queries

### Tool Registration

**Pattern**:
- Enhanced tools use `server.enhanceTool()`
- New tools use `server.addTool()`
- All tools include rich metadata
- Discovery validation middleware

## Project Structure

### Platform-Native Organization

```
src/
├── index.ts                 # MCP server setup
├── tools/
│   ├── registry.ts          # Tool registration
│   ├── enhance-existing.ts  # Enhanced tools
│   └── dashboards.ts        # New dashboard tools
├── core/
│   ├── discovery.ts         # Platform discovery
│   └── adaptive-widgets.ts  # Widget adaptation
└── types/
    └── platform.ts          # TypeScript types
```

## Key Architectural Decisions

### 1. No Hardcoded Schemas
- Every field discovered at runtime
- No assumptions about event types
- Adaptive to any account configuration

### 2. Enhanced Existing Tools
- Leverage new-branch tool implementations
- Add rich metadata and examples
- Implement discovery validation

### 3. Platform-Native Focus
- Deep understanding of NerdGraph/NRQL
- Entity model awareness
- Cross-account analysis capabilities

### 4. AI Optimization
- Rich tool descriptions
- Chaining guidance
- Error recovery suggestions

## Conclusion

The platform-native architecture delivers a powerful MCP server that understands New Relic deeply while making zero assumptions about customer data schemas. By enhancing existing tools and adding intelligent discovery, we enable AI systems to interact effectively with any New Relic account configuration.
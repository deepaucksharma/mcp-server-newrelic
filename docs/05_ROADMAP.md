# Implementation Roadmap: Platform-Native MCP Server

This document outlines the phased implementation plan for the New Relic Platform-Native MCP Server, aligned with the zero hardcoded schemas philosophy and precision-engineered tool approach. Each phase delivers concrete platform analysis value while building toward a complete platform intelligence system.

## Overview

The implementation follows a 4-phase approach over 6 weeks, progressing from enhanced tool metadata to production-ready platform analysis:

| Phase | Focus | Duration | Value Delivered |
|-------|-------|----------|-----------------|
| Phase 1 | Enhanced Tool Metadata | Week 1 | Rich descriptions, examples, AI-optimized metadata |
| Phase 2 | Discovery Engine | Week 2-3 | Comprehensive schema discovery, zero hardcoded assumptions |
| Phase 3 | Dashboard Generation | Week 4 | Adaptive widgets, template system, dry-run capabilities |
| Phase 4 | Platform Intelligence | Week 5-6 | Cross-account analysis, adoption scoring, migration tools |

## Phase 1: Enhanced Tool Metadata (Week 1) ✅

### Objective
Enhance existing new-branch tools with rich metadata for AI orchestration without changing core functionality.

### Deliverables

1. **Enhanced Tool Descriptions**
   ```yaml
   - run_nrql_query: Discovery validation, rich examples
   - search_entities: Entity type guidance, relationship hints
   - get_entity_details: Golden metrics enrichment
   - create_dashboard: Adaptive widget suggestions
   ```

2. **AI-Optimized Metadata Structure**
   ```typescript
   metadata: {
     category: "query" | "discovery" | "dashboard" | "analysis",
     costIndicator: "low" | "medium" | "high",
     readOnlyHint: boolean,
     destructiveHint: boolean,
     requiresConfirmation: boolean,
     returnsNextCursor: boolean
   }
   ```

3. **Comprehensive Examples**
   - Common query patterns with expected outputs
   - Error scenarios with recovery suggestions
   - Tool chaining recommendations
   - Performance optimization hints

4. **Discovery Caching Layer**
   - Map-based memory cache for schema discoveries
   - TTL management (event types: 1h, metrics: 30m, entities: 15m)
   - Cache invalidation strategies
   - Performance metrics

### Success Criteria
- All existing tools enhanced with rich metadata
- AI successfully chains tools based on metadata hints
- Cache hit rate >70% for repeated discoveries
- Zero breaking changes to existing functionality

## Phase 2: Discovery Engine (Week 2-3) 🔄

### Objective
Build comprehensive schema discovery that makes zero hardcoded assumptions about any New Relic account.

### Deliverables

1. **discover_schemas Tool**
   ```typescript
   interface DiscoverSchemasInput {
     account_id: number;
     include_attributes?: boolean;
     include_metrics?: boolean;
   }
   
   interface DiscoverSchemasOutput {
     event_types: EventTypeInfo[];
     metrics: MetricInfo[];
     summary: {
       has_apm_data: boolean;
       has_otel_data: boolean;
       has_dimensional_metrics: boolean;
     }
   }
   ```

2. **Schema Discovery Components**
   - Event type discovery via `SHOW EVENT TYPES`
   - Attribute profiling with `keyset()`
   - Metric discovery from `Metric` event type
   - Cardinality and coverage analysis

3. **Heuristic Field Detection**
   - Service identifier detection (5-50 unique values)
   - Error indicator identification (boolean, status codes, error.class)
   - Metric field recognition (numeric with units)
   - Temporal field discovery

4. **Platform Feature Detection**
   ```typescript
   // Detect platform capabilities dynamically
   - Dimensional metrics vs event-based metrics
   - OpenTelemetry indicators
   - Entity synthesis usage
   - Custom instrumentation patterns
   ```

### Success Criteria
- Successfully discovers 100% of event types and attributes
- Heuristic detection accuracy >85% for common fields
- Works across legacy APM and modern OTel schemas
- No hardcoded field names anywhere in codebase

## Phase 3: Dashboard Generation (Week 4) ⏳

### Objective
Create adaptive dashboards that automatically adjust to discovered schemas without manual configuration.

### Deliverables

1. **dashboard_generate Tool**
   ```typescript
   interface DashboardGenerateInput {
     template_name: "golden-signals" | "dependencies" | 
                    "infrastructure" | "logs-analysis" | "custom";
     entity_guid: string;
     dashboard_name?: string;
     time_range?: string;
     dry_run?: boolean;
     custom_widgets?: WidgetConfig[];
   }
   ```

2. **Adaptive Widget System**
   - Widget intent mapping (error rate, latency, throughput)
   - Field discovery and selection
   - Query adaptation based on available data
   - Fallback strategies for missing fields

3. **Template Library**
   ```yaml
   golden-signals:
     - Error rate (adapts to error/status_code/error.class)
     - Latency (dimensional metrics or event duration)
     - Throughput (count vs rate based on data)
     - Saturation (CPU, memory, custom metrics)
   
   platform-analysis:
     - Event type distribution
     - Dimensional vs event metrics
     - Custom attribute usage
     - Platform feature adoption
   ```

4. **Widget Adaptation Engine**
   ```typescript
   // Example: Error Rate Widget Adaptation
   1. Discover error indicators
   2. Find service identifier
   3. Choose best query pattern
   4. Build NRQL with confidence score
   5. Provide alternatives
   ```

### Success Criteria
- Dashboards work in any account without modification
- >95% of widgets display meaningful data on first attempt
- Graceful degradation when expected fields missing
- Dry-run mode accurately predicts dashboard output

## Phase 4: Platform Intelligence (Week 5-6) ⏳

### Objective
Enable sophisticated cross-account platform analysis and governance workflows.

### Deliverables

1. **platform_analyze_adoption Tool**
   ```typescript
   interface PlatformAnalyzeAdoptionInput {
     account_ids: number[];
     metrics: AdoptionMetric[];
   }
   
   type AdoptionMetric = 
     | "dimensional_metrics"
     | "opentelemetry"
     | "entity_synthesis"
     | "custom_instrumentation"
     | "distributed_tracing"
     | "logs_in_context";
   ```

2. **Cross-Account Analysis**
   - Schema variation detection
   - Adoption pattern identification
   - Best practice scoring
   - Modernization recommendations

3. **Migration Intelligence**
   - Field mapping discovery
   - Compatibility scoring
   - Transformation rule generation
   - Validation query creation

4. **Platform Governance**
   - Compliance gap detection
   - Naming consistency analysis
   - Data quality assessment
   - Optimization opportunities

### Success Criteria
- Analyze 1000+ accounts without manual intervention
- Adoption scoring accuracy >90%
- Migration recommendations reduce manual effort by >80%
- Platform insights actionable without deep NR expertise

## Implementation Principles

### Zero Hardcoded Schemas
```typescript
// ❌ NEVER DO THIS
const errorRate = `SELECT count(*) FROM Transaction WHERE error = true`;

// ✅ ALWAYS DO THIS
const errorField = await discoverErrorField(schema);
const errorRate = buildErrorQuery(errorField);
```

### Discovery-First Pattern
Every tool operation follows this pattern:
1. Discover available data
2. Validate against discovery
3. Adapt to what exists
4. Provide alternatives
5. Explain decisions

### Rich Metadata Everything
```typescript
// Every tool response includes
{
  result: any,
  metadata: {
    discovered_fields: string[],
    adaptation_confidence: number,
    alternatives: Alternative[],
    performance_hint: string
  }
}
```

## Success Metrics

### Technical Metrics
- **Schema Coverage**: 100% of event types/attributes discovered
- **Adaptation Success**: >95% of queries work without modification  
- **Performance**: P95 discovery latency <200ms with warm cache
- **Scale**: Handle 1000+ accounts in parallel analysis

### Business Metrics
- **Time to Value**: <5 minutes from connection to first dashboard
- **Maintenance**: Zero schema updates required as accounts evolve
- **Adoption**: >80% reduction in manual dashboard creation time
- **Platform Visibility**: Complete analysis without hardcoded assumptions

## Technology Stack

```yaml
Runtime: Node.js 20+ / Bun 1.0+
Language: TypeScript 5.3+
MCP SDK: @modelcontextprotocol/sdk 1.0+
GraphQL: graphql-request 7.0+
Validation: Zod 3.23+
Cache: Map-based memory (Redis optional)
Testing: Vitest + MSW
```

## Getting Started

### For Contributors
1. **Understand Philosophy**: Read [01_VISION_AND_PHILOSOPHY.md](01_VISION_AND_PHILOSOPHY.md)
2. **Review Architecture**: Study [02_ARCHITECTURE.md](02_ARCHITECTURE.md)
3. **Check Tools**: Examine [03_TOOL_SPECIFICATION.md](03_TOOL_SPECIFICATION.md)
4. **Pick a Phase**: Start with earliest incomplete phase
5. **Follow Patterns**: Use discovery-first approach throughout

### Priority Order
1. 🔴 **Critical**: Enhanced tool metadata (enables all AI features)
2. 🟠 **High**: Discovery engine (foundation for adaptation)
3. 🟡 **Medium**: Dashboard generation (immediate user value)
4. 🟢 **Enhancement**: Platform intelligence (advanced workflows)

## Phase Completion Checklist

### Phase 1 ✅
- [x] Enhanced tool descriptions with discovery hints
- [x] Rich metadata structure for AI consumption
- [x] Comprehensive examples for common patterns
- [x] Basic caching layer implementation

### Phase 2 🔄
- [ ] discover_schemas tool implementation
- [ ] Heuristic field detection algorithms
- [ ] Platform feature detection
- [ ] Discovery result caching

### Phase 3 ⏳
- [ ] dashboard_generate tool
- [ ] Adaptive widget engine
- [ ] Template library
- [ ] Dry-run capabilities

### Phase 4 ⏳
- [ ] platform_analyze_adoption tool
- [ ] Cross-account analysis engine
- [ ] Migration planning tools
- [ ] Governance dashboards

## Conclusion

This roadmap delivers a Platform-Native MCP Server that truly understands New Relic at a deep level while making zero assumptions about customer schemas. By enhancing existing tools and adding intelligent discovery, we enable AI systems to work with any New Relic account configuration automatically.

The phased approach ensures each milestone delivers immediate value while building toward comprehensive platform intelligence. The result is a system that transforms New Relic from a monitoring tool into an intelligent platform analysis system.
# Vision and Philosophy: Platform-Native Zero Hardcoded Schemas

## Executive Summary

This specification defines an internal MCP server for New Relic platform analysis and governance. The core philosophy is simple yet powerful: **Every schema, field, and relationship is discovered at runtime.** The server makes no assumptions about attribute names, event types, or metric structures, ensuring compatibility with any New Relic account configuration.

## The Problem: Platform Heterogeneity

New Relic accounts vary dramatically:
- Legacy APM uses `appName`, modern OTel uses `service.name`
- Error indicators range from boolean `error` to numeric `http.status_code`
- Dimensional metrics coexist with event-based telemetry
- Custom attributes proliferate across different teams

Traditional tools fail because they hardcode assumptions:
- Dashboard templates assume specific field names
- Queries break when schemas evolve
- Cross-account analysis becomes impossible
- Platform adoption metrics require manual mapping

The solution? **Zero hardcoded schemas - discover everything at runtime.**

## Core Philosophy: Platform-Native Intelligence

### Zero Hardcoded Schemas

Every operation begins with discovery:
1. **Discover Event Types** - No assumptions about what events exist
2. **Discover Attributes** - No hardcoded field names
3. **Discover Metrics** - Dimensional metrics found dynamically
4. **Discover Entities** - Entity types and relationships mapped at runtime

### Platform-Native Approach

The server enhances existing tools with intelligence:
1. **Enhanced Existing Tools** - Add discovery suggestions to run_nrql_query
2. **Rich Metadata** - AI-optimized descriptions and examples
3. **Adaptive Widgets** - Dashboards that adapt to any schema
4. **Cross-Account Analysis** - Compare platform adoption across accounts

### Key Principles

1. **Discovery First**: Never assume - always discover
2. **Adaptation**: Tools adapt to discovered schemas
3. **Intelligence**: Rich metadata for AI orchestration
4. **Safety**: Dry-run and preview for all mutations
5. **Performance**: Efficient caching of discoveries

## Implementation Strategy

### Enhanced Tool Pattern

```
Traditional Tool:                  Platform-Native Tool:
1. Basic parameters                1. Rich parameter schemas
2. Simple execution                2. Discovery validation
3. Raw results                     3. Contextual results
4. No guidance                     4. AI-optimized examples
5. Static behavior                 5. Adaptive intelligence
```

### In Practice: Platform-Native Enhancement

The platform-native approach enhances existing tools with discovery intelligence:

**Enhanced run_nrql_query:**
- Validates event types exist before execution
- Suggests discovery if unknown schemas referenced
- Provides rich examples for common patterns

**Enhanced search_entities:**
- Guides users with discovered entity types
- Maps relationships automatically
- Enriches results with golden metrics

**New dashboard_generate:**
- Discovers available fields for entity type
- Adapts widget queries to actual schema
- Never assumes field names exist

## Real-World Use Cases

### Cross-Account Platform Analysis
Analyze platform adoption across 1000+ accounts:
- Discover which accounts use dimensional metrics vs events
- Identify OpenTelemetry adoption patterns
- Generate adoption scorecards automatically
- No manual schema mapping required

### Adaptive Dashboard Generation
Create dashboards that work in any account:
- Discover error indicators (boolean error, http.status_code, error.class)
- Find service identifiers (appName, service.name, app.name)
- Build queries using discovered fields
- Dashboards adapt to each account's schema

### Platform Migration Planning
Plan migrations with full visibility:
- Compare schemas between legacy and modern accounts
- Identify translation patterns automatically
- Generate migration guides based on actual data
- Validate migrations with discovery-based testing

## Technical Implementation

### Platform Discovery Engine

The discovery engine dynamically learns about platform structure:

**Event Type Discovery:**
- Uses SHOW EVENT TYPES to find all available events
- Profiles each event type with sample counts
- No hardcoded event type assumptions

**Attribute Discovery:**
- Uses keyset() to discover all attributes
- Profiles attribute types and cardinality
- Identifies service identifiers heuristically

**Metric Discovery:**
- Queries Metric event type for dimensional metrics
- Extracts metric names and dimensions
- Maps between metrics and events

### Adaptive Widget Generation

Widgets adapt to discovered schemas:

**Error Rate Widget:**
- Discovers error indicators (boolean error, http.status_code, error.class)
- Finds appropriate service identifier
- Builds query using discovered fields

**Latency Widget:**
- Checks for dimensional metrics first (more efficient)
- Falls back to event-based queries
- Uses discovered duration fields

**Throughput Widget:**
- Adapts between count-based and rate-based metrics
- Discovers appropriate aggregation fields
- Handles both events and metrics

### Cross-Account Intelligence

Platform analysis across multiple accounts:

**Adoption Analysis:**
- Discovers usage patterns per account
- Compares dimensional metrics vs events
- Identifies OpenTelemetry adoption
- No manual schema mapping required

**Migration Planning:**
- Compares schemas between accounts
- Identifies field mapping patterns
- Generates migration guides
- Validates compatibility

## Success Metrics

The platform-native approach delivers measurable value:

### Schema Coverage
- **Target**: Successfully discovers 100% of event types and attributes
- **Reality**: No hardcoded schemas means universal compatibility

### Adaptation Rate
- **Target**: >95% of generated queries work without modification
- **Reality**: Adaptive widgets adjust to any schema automatically

### Cross-Account Scale
- **Target**: Can analyze 1000+ accounts in parallel
- **Reality**: Efficient discovery caching enables massive scale

### AI Success Rate
- **Target**: >90% of multi-step workflows complete successfully
- **Reality**: Rich metadata and examples guide AI orchestration

## Key Benefits

### For Platform Teams
- **Universal Compatibility**: Works with any New Relic account configuration
- **Zero Maintenance**: No schema updates when accounts change
- **Cross-Account Insights**: Analyze platform adoption at scale
- **Migration Intelligence**: Plan and validate migrations automatically

### For AI Orchestration
- **Rich Metadata**: Every tool enhanced with examples and guidance
- **Discovery Suggestions**: Tools guide AI to discover before querying
- **Adaptive Behavior**: Queries adjust to discovered schemas
- **Explainable Results**: Clear reasoning for every decision

### For End Users
- **Always Works**: Dashboards adapt to your specific schema
- **No Configuration**: Discovery eliminates manual setup
- **Instant Value**: Start analyzing immediately
- **Future Proof**: Automatically adapts as schemas evolve

## Conclusion

The platform-native MCP server represents a fundamental shift in how we interact with observability platforms. By discovering everything at runtime and adapting to what we find, we create tools that are simultaneously more powerful and easier to use.

This is not just about technical elegance – it's about delivering real value to platform teams who need to understand and govern their New Relic usage at scale. It's about enabling AI systems to intelligently interact with any New Relic account. And ultimately, it's about making observability accessible to everyone, regardless of their specific schema or configuration.

**Zero hardcoded schemas. Infinite possibilities.**
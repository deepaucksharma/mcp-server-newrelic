# MCP Server New Relic: Platform-Native Specification

> 🚀 **Internal MCP Server for New Relic Platform Analysis & Governance**  
> Built with **official `@modelcontextprotocol/sdk` + TypeScript 5.3+**  
> **Zero Hardcoded Schemas** - Every field, attribute, and metric discovered at runtime  
> **Platform-Native Intelligence** - Deep understanding of NerdGraph, NRQL, and entity model

## Executive Summary

This specification defines an internal MCP server for New Relic platform analysis and governance. It discovers all schemas dynamically, focuses on core platform concepts (entities, metrics, dashboards), and enhances existing new-branch tools with rich metadata for AI orchestration.

## Core Philosophy: Zero Hardcoded Schemas

**Every schema, field, and relationship is discovered at runtime.** The server makes no assumptions about attribute names, event types, or metric structures. This ensures compatibility with any New Relic account configuration, from legacy APM to modern OpenTelemetry.

## Documentation Overview

This repository contains the complete specification and implementation guide:

| Document | Purpose |
|----------|---------|
| [01_VISION_AND_PHILOSOPHY.md](docs/01_VISION_AND_PHILOSOPHY.md) | **Zero Hardcoded Schemas** – Core philosophy and platform-native approach |
| [02_ARCHITECTURE.md](docs/02_ARCHITECTURE.md) | **Platform Architecture** – Discovery engine and adaptive components |
| [03_TOOL_SPECIFICATION.md](docs/03_TOOL_SPECIFICATION.md) | **Enhanced Tools** – Existing tool enhancement and new platform tools |
| [04_USE_CASES_AND_WORKFLOWS.md](docs/04_USE_CASES_AND_WORKFLOWS.md) | **Platform Analysis** – Cross-account analysis and governance workflows |
| [05_ROADMAP.md](docs/05_ROADMAP.md) | **Implementation Phases** – 4-phase platform-native delivery plan |
| [06_CONTRIBUTING.md](docs/06_CONTRIBUTING.md) | **Development Guide** – Platform-native development patterns |
| [07_DISCOVERY_ENGINE.md](docs/07_DISCOVERY_ENGINE.md) | **Discovery Deep Dive** – Technical details of schema discovery |
| [08_ADAPTIVE_QUERY_BUILDER.md](docs/08_ADAPTIVE_QUERY_BUILDER.md) | **Adaptive Intelligence** – Widget and query adaptation system |

## Architecture Overview

The platform-native MCP server enhances existing tools and adds new capabilities:

### Enhanced Existing Tools
- **run_nrql_query** - Enhanced with discovery suggestions and schema validation
- **search_entities** - Enriched with entity type guidance and relationship mapping
- **get_entity_details** - Augmented with golden metrics and dependency discovery
- **create_dashboard** - Extended with adaptive widget generation

### New Platform Tools
- **discover_schemas** - Comprehensive schema discovery across event types and metrics
- **dashboard_generate** - Adaptive dashboard creation with zero hardcoded fields
- **platform_analyze_adoption** - Cross-account platform usage analysis

## Key Design Principles

1. **No Hardcoded Schemas**: Every field, attribute, and metric name is discovered
2. **Adaptive Intelligence**: Tools adapt their behavior based on discovered data
3. **Platform-Native**: Deep understanding of NerdGraph, NRQL, and entity model
4. **AI-Optimized**: Rich metadata enables LLMs to chain tools effectively
5. **Safety First**: Destructive operations require dry-run and confirmation

## Technology Stack

```yaml
Runtime: Node.js 20+ / Bun 1.0+
Language: TypeScript 5.3+
MCP SDK: @modelcontextprotocol/sdk 1.0+
GraphQL: graphql-request 7.0+
Cache: Map-based memory cache (Redis optional)
Validation: Zod 3.23+
Testing: Vitest + MSW
```

## Implementation Approach

### Phase 1: Enhanced Tool Metadata (Week 1)
- Add rich descriptions, examples, and schemas to existing tools
- Implement discovery caching layer
- Add parameter validation and error handling

### Phase 2: Discovery Engine (Week 2-3)
- Build comprehensive schema discovery
- Implement metric vs event detection
- Add entity relationship mapping

### Phase 3: Dashboard Generation (Week 4)
- Create adaptive widget templates
- Build dashboard composition engine
- Add dry-run and preview capabilities

### Phase 4: Platform Intelligence (Week 5-6)
- Cross-account analysis tools
- Adoption scoring algorithms
- Migration planning tools

## Success Metrics

- **Schema Coverage**: Successfully discovers 100% of event types and attributes
- **Adaptation Rate**: >95% of generated queries work without modification
- **Cross-Account Scale**: Can analyze 1000+ accounts in parallel
- **AI Success Rate**: >90% of multi-step workflows complete successfully

## Example Usage

### Discovering Schemas
```typescript
// First, discover what's available
const schemas = await mcp.call('discover_schemas', {
  account_id: 12345,
  include_attributes: true,
  include_metrics: true
});

// Returns comprehensive schema information
{
  event_types: [
    { name: 'Transaction', sample_count: 1234567, attributes: [...] },
    { name: 'Log', sample_count: 987654, attributes: [...] }
  ],
  metrics: [
    { name: 'http.server.duration', dimensions: ['service.name', 'http.method'] }
  ],
  summary: {
    has_apm_data: true,
    has_otel_data: true
  }
}
```

### Generating Adaptive Dashboards
```typescript
// Generate dashboard that adapts to discovered schema
const dashboard = await mcp.call('dashboard_generate', {
  template_name: 'golden-signals',
  entity_guid: 'MXxBUE18QVBQTElDQVRJT058MTIzNDU2',
  dry_run: true // Preview first
});

// Dashboard automatically uses discovered fields
// No hardcoded 'appName' or 'error' assumptions
```

## Getting Started

1. Clone the repository
2. Install dependencies: `npm install`
3. Configure New Relic credentials
4. Run the MCP server: `npm start`
5. Connect from Claude Desktop or other MCP clients

## Contributing

This platform-native specification delivers a powerful MCP server that understands New Relic deeply while making zero assumptions about customer data schemas. We welcome contributions that maintain this philosophy.

---

**License**: MIT  
**Code of Conduct**: We welcome all contributors. See [06_CONTRIBUTING.md](docs/06_CONTRIBUTING.md) for guidelines.
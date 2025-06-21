# Platform-Native Tool Specification

This document defines how the platform-native MCP server enhances existing new-branch tools and adds new platform capabilities. The approach focuses on zero hardcoded schemas and rich metadata for AI orchestration.

## Core Philosophy

### Enhanced Existing Tools

The platform-native approach enhances existing new-branch tools rather than replacing them:

1. **run_nrql_query** - Enhanced with discovery validation
2. **search_entities** - Enriched with entity guidance
3. **get_entity_details** - Augmented with golden metrics
4. **create_dashboard** - Extended with adaptive widgets

### New Platform Tools

New tools focus on platform-specific capabilities:

1. **discover_schemas** - Comprehensive schema discovery
2. **dashboard_generate** - Adaptive dashboard creation
3. **platform_analyze_adoption** - Cross-account analysis

## Tool Enhancement Pattern

### Metadata Structure

Every tool includes rich metadata for AI consumption:

```yaml
metadata:
  category: query | discovery | dashboard | analysis
  costIndicator: low | medium | high
  readOnlyHint: true | false
  destructiveHint: true | false
  requiresConfirmation: true | false
  returnsNextCursor: true | false
```

### Example Structure

```yaml
examples:
  - description: "Clear, concise description"
    params:
      account_id: 12345
      query: "SELECT count(*) FROM Transaction"
    expectedBehavior: "Returns transaction count"
```

### Discovery Validation

Tools validate discovered schemas before execution:

```
1. Extract referenced event types from query
2. Check if types exist in discovered schemas
3. Suggest discover_schemas if unknown
4. Provide helpful error messages
```

## Enhanced Existing Tools

These tools exist in new-branch and are enhanced with rich metadata:

### run_nrql_query

**Original Purpose**: Execute NRQL queries

**Enhancements**:
- Discovery validation before execution
- Rich examples for common patterns
- Query performance metadata
- Error suggestions with discovery hints

**Enhanced Metadata**:
```yaml
description: |
  Execute NRQL queries against New Relic data.
  
  IMPORTANT: Always discover available event types and attributes before querying.
  Common patterns:
  - Time series: SELECT average(metric) FROM EventType TIMESERIES
  - Faceted: SELECT count(*) FROM EventType FACET attribute
  - Percentiles: SELECT percentile(metric, 95) FROM EventType
  
  The tool automatically adds LIMIT 100 if not specified.

inputSchema:
  type: object
  properties:
    account_id:
      type: number
      description: New Relic account ID
    query:
      type: string
      description: NRQL query string (no trailing semicolon)
    timeout:
      type: number
      description: Query timeout in milliseconds
      default: 30000
    include_metadata:
      type: boolean
      description: Include query performance metadata
      default: false
  required: [query]

examples:
  - description: Get error rate for a service
    params:
      query: "SELECT percentage(count(*), WHERE error = true) FROM Transaction SINCE 1 hour ago"
  - description: Get top 10 slowest transactions
    params:
      query: "SELECT average(duration) FROM Transaction FACET name SINCE 1 hour ago LIMIT 10"

metadata:
  readOnlyHint: true
  category: query
  costIndicator: low
```

### search_entities

**Original Purpose**: Search for entities in New Relic

**Enhancements**:
- Entity type guidance with examples
- Automatic relationship discovery
- Golden metrics enrichment
- Pagination support with cursor

**Enhanced Metadata**:
```yaml
description: |
  Search for entities in New Relic. Entities are the core objects that emit telemetry.
  
  Common entity types:
  - APM application (domain: APM)
  - Browser application (domain: BROWSER) 
  - Infrastructure host (domain: INFRA)
  - Synthetic monitor (domain: SYNTH)
  - Dashboard (domain: NR1)
  
  Use this before querying metrics or creating dashboards.

inputSchema:
  type: object
  properties:
    account_id:
      type: number
    name:
      type: string
      description: Entity name (supports wildcards: checkout*)
    domain:
      type: string
      enum: [APM, BROWSER, INFRA, SYNTH, NR1]
      description: Entity domain to filter by
    type:
      type: string
      description: Specific entity type (e.g., APPLICATION, HOST)
    tags:
      type: object
      description: Tag filters as key:value pairs
    limit:
      type: number
      default: 50
      description: Maximum results to return
    cursor:
      type: string
      description: Pagination cursor from previous request

examples:
  - description: Find all APM services with "api" in name
    params:
      name: "*api*"
      domain: APM
  - description: Find production hosts
    params:
      domain: INFRA
      tags:
        environment: production

metadata:
  readOnlyHint: true
  returnsNextCursor: true
```

### get_entity_details

**Original Purpose**: Get details about a specific entity

**Enhancements**:
- Golden metrics calculation
- Related entities discovery
- Alert violations summary
- Rich metadata and tags

**Enhanced Metadata**:
```yaml
description: |
  Get comprehensive details about a specific entity including:
  - Golden metrics (throughput, errors, latency)
  - Related entities and dependencies
  - Recent alert violations
  - Tags and metadata
  
  Always use search_entities first to get the entity GUID.

inputSchema:
  type: object
  properties:
    guid:
      type: string
      description: Entity GUID (from search_entities)
      pattern: '^[A-Za-z0-9+/]+='  # Base64 pattern
    include_golden_metrics:
      type: boolean
      default: true
      description: Include golden signal metrics
    include_relationships:
      type: boolean
      default: true
      description: Include related entities
  required: [guid]

outputSchema:
  type: object
  properties:
    entity:
      type: object
      properties:
        guid: {type: string}
        name: {type: string}
        type: {type: string}
        domain: {type: string}
        tags: {type: array}
    goldenMetrics:
      type: object
      properties:
        throughput: {type: number}
        errorRate: {type: number}
        latency: {type: number}
    relationships:
      type: array
      items:
        type: object
        properties:
          type: {type: string}
          targetGuid: {type: string}
          targetName: {type: string}
```

## New Platform Tools

### discover_schemas

**Purpose**: Comprehensive schema discovery - the FIRST tool to run

**Description**:
```yaml
description: |
  Discover all available event types and their attributes in an account.
  This is typically the FIRST tool you should run when working with a new account.
  
  Returns:
  - Event types with sample counts
  - Attribute names and types for each event
  - Metric names if dimensional metrics are used

inputSchema:
  type: object
  properties:
    account_id:
      type: number
      description: New Relic account ID
    include_attributes:
      type: boolean
      default: false
      description: Include detailed attribute information
    include_metrics:
      type: boolean
      default: true
      description: Include dimensional metrics discovery
  required: [account_id]

metadata:
  readOnlyHint: true
  category: discovery

examples:
  - description: Basic schema discovery
    params:
      account_id: 12345
  - description: Full discovery with attributes
    params:
      account_id: 12345
      include_attributes: true
```

### dashboard_generate

**Purpose**: Generate dashboards that adapt to any schema

**Description**:
```yaml
description: |
  Generate a dashboard using adaptive templates that discover the correct fields.
  
  Available templates:
  - golden-signals: Error rate, latency, throughput, saturation
  - dependencies: Service map and dependencies
  - infrastructure: CPU, memory, disk, network
  - logs-analysis: Log patterns and errors
  - custom: Provide your own widget configuration
  
  The tool automatically discovers which attributes and metrics are available
  for the entity and adapts the queries accordingly.

inputSchema:
  type: object
  properties:
    template_name:
      type: string
      enum: [golden-signals, dependencies, infrastructure, logs-analysis, custom]
      description: Dashboard template to use
    entity_guid:
      type: string
      description: Entity GUID to create dashboard for
    dashboard_name:
      type: string
      description: Name for the dashboard (auto-generated if not provided)
    time_range:
      type: string
      default: '1 hour ago'
      description: Default time range for widgets
    dry_run:
      type: boolean
      default: true
      description: Preview dashboard JSON without creating
    custom_widgets:
      type: array
      description: Custom widget configurations (for custom template)
  required: [template_name, entity_guid]

examples:
  - description: Create golden signals dashboard for an APM service
    params:
      template_name: golden-signals
      entity_guid: MXxBUE18QVBQTElDQVRJT058MTIzNDU2
      dashboard_name: Checkout Service Golden Signals

metadata:
  destructiveHint: true
  requiresConfirmation: true
```

### platform_analyze_adoption

**Purpose**: Analyze platform adoption across accounts

**Description**:
```yaml
description: |
  Analyze platform adoption patterns across multiple accounts.
  Useful for understanding how different teams use New Relic features.

inputSchema:
  type: object
  properties:
    account_ids:
      type: array
      items: {type: number}
      description: List of account IDs to analyze
    metrics:
      type: array
      items:
        type: string
        enum: [dimensional_metrics, opentelemetry, entity_synthesis, dashboards]
      description: Adoption metrics to calculate
  required: [account_ids, metrics]

metadata:
  readOnlyHint: true
  category: analysis
```

## Implementation Pattern

### Enhanced Tool Example

```typescript
// src/tools/enhance-existing.ts
export function enhanceExistingTools(server: Server, discovery: PlatformDiscovery) {
  
  // Enhance run_nrql_query with discovery validation
  server.enhanceTool('run_nrql_query', {
    description: enhancedDescription,
    inputSchema: enhancedSchema,
    examples: richExamples,
    metadata: {
      readOnlyHint: true,
      category: 'query',
      costIndicator: 'low'
    },
    
    // Wrap handler to add discovery validation
    handler: async (params, originalHandler) => {
      // Extract event types from query
      const eventTypes = extractEventTypes(params.query);
      const knownTypes = await discovery.discoverEventTypes(params.account_id);
      
      // Validate all event types exist
      const unknownTypes = eventTypes.filter(
        type => !knownTypes.some(known => known.name === type)
      );
      
      if (unknownTypes.length > 0) {
        return {
          error: `Unknown event types: ${unknownTypes.join(', ')}. Run 'discover_schemas' first.`,
          suggestion: 'discover_schemas',
          discovered_types: knownTypes.map(t => t.name)
        };
      }
      
      // Execute original handler
      return originalHandler(params);
    }
  });
}
```

### New Tool Example

```typescript
// src/tools/dashboards.ts
export function registerDashboardTools(server: Server, discovery: PlatformDiscovery) {
  
  server.addTool({
    name: 'dashboard_generate',
    ...dashboardGenerateMetadata,
    
    handler: async (params) => {
      // Get entity details
      const entity = await getEntityDetails(params.entity_guid);
      
      // Discover available data
      const discovery = await discoverEntityData(entity);
      
      // Load template
      const template = await loadTemplate(params.template_name);
      
      // Adapt widgets to discovered schema
      const adaptedWidgets = await adaptWidgetsToSchema(
        template.widgets,
        discovery,
        entity
      );
      
      if (params.dry_run) {
        return {
          preview: adaptedWidgets,
          discovered_fields: discovery,
          adaptations: getAdaptationExplanation(template, discovery)
        };
      }
      
      // Create dashboard
      return createDashboard(adaptedWidgets);
    }
  });
}```

## Key Design Decisions

### 1. Zero Hardcoded Schemas
Every field and event type is discovered at runtime. No assumptions about:
- Event type names (Transaction, Span, etc.)
- Attribute names (appName, service.name, etc.)
- Error indicators (boolean error, http.status_code)
- Metric structures

### 2. Enhanced vs New Tools
- Enhance existing tools when functionality overlaps
- Create new tools for platform-specific capabilities
- Always add rich metadata for AI consumption

### 3. Adaptive Intelligence
- Widgets adapt to discovered schemas
- Queries validate against known event types
- Fallback strategies for missing fields

### 4. Platform Focus
- Deep NerdGraph/NRQL understanding
- Entity model awareness
- Cross-account analysis capabilities

## Conclusion

The platform-native tool specification focuses on enhancing existing tools with discovery intelligence while adding new platform-specific capabilities. By maintaining zero hardcoded schemas and providing rich metadata, we enable AI systems to effectively interact with any New Relic account configuration.
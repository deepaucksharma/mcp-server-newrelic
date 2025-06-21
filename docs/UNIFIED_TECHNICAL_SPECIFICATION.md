# MCP Server New Relic: Unified Technical Specification

## Executive Summary

This specification synthesizes a production-ready MCP server for New Relic observability that embodies the "Zero Assumptions Manifesto." Built with the official MCP TypeScript SDK, it provides an intelligent, discovery-first platform that adapts to any New Relic environment without hard-coded assumptions.

## Core Architecture

### 1. Technology Stack

```yaml
Runtime: Node.js 20+ / Bun 1.0+
Language: TypeScript 5.3+
MCP SDK: @modelcontextprotocol/sdk 1.0+
Framework: Fastify 5.0 (HTTP transport)
Validation: Zod 3.23+
GraphQL: graphql-request 7.0+
Cache: Redis/Upstash + Keyv
Observability: OpenTelemetry
Testing: Vitest + MSW
```

### 2. Project Structure

```
mcp-server-newrelic/
├── src/
│   ├── index.ts                 # Entry point with MCP server setup
│   ├── core/
│   │   ├── discovery/           # Discovery engine (heart of the platform)
│   │   │   ├── engine.ts        # Main discovery orchestrator
│   │   │   ├── schemas.ts       # Event type discovery
│   │   │   ├── attributes.ts    # Attribute profiling
│   │   │   ├── service-id.ts    # Service identifier detection
│   │   │   ├── errors.ts        # Error indicator discovery
│   │   │   ├── metrics.ts       # Metric census
│   │   │   └── cache.ts         # Discovery cache layer
│   │   ├── context.ts           # Request context & world model
│   │   ├── errors.ts            # Typed error system
│   │   └── types.ts             # Core type definitions
│   ├── tools/                   # MCP tool implementations
│   │   ├── discover/            # Discovery tools
│   │   ├── analyze/             # Analysis tools
│   │   ├── query/               # Query tools
│   │   ├── dashboards/          # Dashboard tools
│   │   ├── alerts/              # Alert tools
│   │   └── registry.ts          # Tool registry
│   ├── adapters/                # External service adapters
│   │   ├── nerdgraph.ts         # New Relic GraphQL client
│   │   ├── insights.ts          # NRQL REST API client
│   │   └── telemetry.ts         # Metrics ingestion client
│   ├── intelligence/            # Advanced analysis
│   │   ├── aqb.ts              # Adaptive Query Builder
│   │   ├── anomaly.ts          # Anomaly detection
│   │   └── workflow.ts         # Workflow orchestrator
│   └── transport/               # MCP transports
│       ├── stdio.ts            # Standard I/O transport
│       └── http.ts             # HTTP/SSE transport
├── workflows/                   # Declarative workflow definitions
├── templates/                   # Dashboard/alert templates
├── tests/                       # Comprehensive test suite
└── docs/                        # Documentation
```

## Core Components

### 3. Discovery Engine

The discovery engine is the heart of the zero-assumptions philosophy. Every tool execution begins with discovery.

```typescript
// src/core/discovery/engine.ts
import { z } from 'zod';
import { LRUCache } from 'lru-cache';
import { RequestContext, DiscoveryGraph, DiscoveryResult } from '../types';

export class DiscoveryEngine {
  private cache: LRUCache<string, DiscoveryResult>;
  
  constructor(private ctx: RequestContext) {
    this.cache = new LRUCache({
      max: 1000,
      ttl: 1000 * 60 * 5, // 5 minutes default
    });
  }

  async buildDiscoveryGraph(accountId: number): Promise<DiscoveryGraph> {
    const cacheKey = `discovery:${accountId}`;
    const cached = await this.ctx.cache.get(cacheKey);
    
    if (cached && !this.isStale(cached)) {
      return cached;
    }

    // Parallel discovery operations
    const [schemas, attributes, serviceId, errors, metrics] = await Promise.all([
      this.discoverSchemas(accountId),
      this.discoverAttributes(accountId),
      this.discoverServiceIdentifier(accountId),
      this.discoverErrorIndicators(accountId),
      this.discoverMetrics(accountId),
    ]);

    const graph: DiscoveryGraph = {
      accountId,
      timestamp: new Date(),
      schemas,
      attributes,
      serviceIdentifier: serviceId,
      errorIndicators: errors,
      metrics,
      confidence: this.calculateConfidence({ schemas, attributes, serviceId }),
    };

    await this.ctx.cache.set(cacheKey, graph, { ttl: 300000 }); // 5 min
    return graph;
  }

  private async discoverSchemas(accountId: number): Promise<SchemaInfo[]> {
    const query = `
      query($accountId: Int!) {
        actor {
          account(id: $accountId) {
            nrql(query: "SHOW EVENT TYPES") {
              results
            }
          }
        }
      }
    `;
    
    const result = await this.ctx.nerdgraph.request(query, { accountId });
    return this.parseSchemaResults(result);
  }

  private async discoverServiceIdentifier(accountId: number): Promise<ServiceIdentifier> {
    // Chain of discovery: appName → service.name → entity.name → custom patterns
    const candidates = [
      { field: 'appName', query: 'SELECT count(*) FROM Transaction WHERE appName IS NOT NULL' },
      { field: 'service.name', query: 'SELECT count(*) FROM Transaction WHERE service.name IS NOT NULL' },
      { field: 'entity.name', query: 'SELECT count(*) FROM Transaction WHERE entity.name IS NOT NULL' },
    ];

    for (const candidate of candidates) {
      const result = await this.ctx.nerdgraph.nrql(accountId, candidate.query);
      if (result.count > 0) {
        return {
          field: candidate.field,
          confidence: 0.95,
          coverage: result.count / result.total,
        };
      }
    }

    // Fallback to pattern detection
    return this.detectCustomServicePattern(accountId);
  }
}
```

### 4. Request Context & World Model

```typescript
// src/core/context.ts
export interface RequestContext {
  requestId: string;
  accountId: number;
  logger: Logger;
  cache: CacheAdapter;
  nerdgraph: NerdGraphClient;
  discovery: DiscoveryEngine;
  telemetry: TelemetryRecorder;
  
  // Carried discovery state
  worldModel?: DiscoveryGraph;
  explainabilityTrace: ExplainabilityTrace;
}

export interface DiscoveryGraph {
  accountId: number;
  timestamp: Date;
  schemas: SchemaInfo[];
  attributes: Record<string, AttributeProfile>;
  serviceIdentifier: ServiceIdentifier;
  errorIndicators: ErrorIndicator[];
  metrics: MetricInfo[];
  dataSources: DataSource[];
  confidence: number;
}

export interface ExplainabilityTrace {
  steps: DiscoveryStep[];
  assumptions: string[];
  confidence: number;
  
  addStep(step: DiscoveryStep): void;
  toMarkdown(): string;
}
```

### 5. MCP Server Implementation

```typescript
// src/index.ts
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { HttpServerTransport } from './transport/http.js';
import { ToolRegistry } from './tools/registry.js';
import { createRequestContext } from './core/context.js';

export async function createServer(options: ServerOptions) {
  const server = new Server({
    name: 'mcp-server-newrelic',
    version: '1.0.0',
    capabilities: {
      tools: true,
      resources: true,
      prompts: true,
    },
  });

  // Initialize tool registry
  const registry = new ToolRegistry();
  
  // Register all tools with discovery-first pattern
  registry.registerDiscoveryTools();
  registry.registerAnalysisTools();
  registry.registerQueryTools();
  registry.registerDashboardTools();
  registry.registerAlertTools();

  // Tool handler with discovery enforcement
  server.setRequestHandler('tools/call', async (request) => {
    const { name, arguments: args } = request.params;
    
    // Create request context
    const ctx = await createRequestContext({
      accountId: args.account_id || process.env.NEW_RELIC_ACCOUNT_ID,
      apiKey: process.env.NEW_RELIC_API_KEY,
    });

    // Get tool definition
    const tool = registry.getTool(name);
    if (!tool) {
      throw new Error(`Unknown tool: ${name}`);
    }

    // Validate inputs
    const validatedArgs = tool.inputSchema.parse(args);

    // Ensure discovery if required
    if (tool.requiresDiscovery && validatedArgs.discover_first !== false) {
      ctx.worldModel = await ctx.discovery.buildDiscoveryGraph(ctx.accountId);
      ctx.explainabilityTrace.addStep({
        type: 'discovery',
        description: 'Built world model from account data',
        confidence: ctx.worldModel.confidence,
      });
    }

    // Execute tool
    const result = await tool.handler(ctx, validatedArgs);

    // Add explainability
    if (ctx.explainabilityTrace.steps.length > 0) {
      result.explainability = {
        traceId: ctx.requestId,
        summary: ctx.explainabilityTrace.toMarkdown(),
      };
    }

    return result;
  });

  return server;
}
```

### 6. Adaptive Query Builder

```typescript
// src/intelligence/aqb.ts
export class AdaptiveQueryBuilder {
  constructor(private worldModel: DiscoveryGraph) {}

  build(params: AQBParams): string {
    const { intent, scope, timeRange } = params;

    switch (intent) {
      case 'error_rate':
        return this.buildErrorRateQuery(scope, timeRange);
      case 'latency_p95':
        return this.buildLatencyQuery('p95', scope, timeRange);
      case 'throughput':
        return this.buildThroughputQuery(scope, timeRange);
      default:
        throw new Error(`Unknown intent: ${intent}`);
    }
  }

  private buildErrorRateQuery(scope: Scope, timeRange: string): string {
    const serviceField = this.worldModel.serviceIdentifier.field;
    const errorIndicator = this.worldModel.errorIndicators[0]; // Best one

    if (!errorIndicator) {
      throw new Error('No error indicators discovered');
    }

    // Build WHERE clause based on error indicator type
    let errorCondition: string;
    switch (errorIndicator.type) {
      case 'boolean':
        errorCondition = `${errorIndicator.field} = true`;
        break;
      case 'http_status':
        errorCondition = `numeric(${errorIndicator.field}) >= 400`;
        break;
      case 'error_class':
        errorCondition = `${errorIndicator.field} IS NOT NULL`;
        break;
      default:
        errorCondition = errorIndicator.condition;
    }

    return `
      SELECT percentage(count(*), WHERE ${errorCondition}) AS 'Error Rate'
      FROM ${errorIndicator.eventType}
      WHERE ${serviceField} = '${scope.value}'
      SINCE ${timeRange}
      TIMESERIES
    `.trim();
  }
}
```

### 7. Tool Implementations

```typescript
// src/tools/discover/schemas.ts
import { z } from 'zod';
import { defineTool } from '../registry';

export const listSchemas = defineTool({
  name: 'discover.list_schemas',
  description: 'Discover all event types (schemas) in the account. ALWAYS run this first.',
  requiresDiscovery: false, // This IS the discovery
  inputSchema: z.object({
    account_id: z.number().optional(),
    include_counts: z.boolean().default(true),
    since: z.string().default('1 hour ago'),
  }),
  outputSchema: z.object({
    schemas: z.array(z.object({
      eventType: z.string(),
      count: z.number(),
      lastIngested: z.string(),
      attributes: z.number(),
    })),
    totalEvents: z.number(),
  }),
  handler: async (ctx, input) => {
    const query = `
      SELECT count(*) as 'count', latest(timestamp) as 'lastSeen' 
      FROM Transaction, PageView, MobileRequest, Log, Metric, Span 
      FACET eventType() 
      SINCE ${input.since}
    `;

    const results = await ctx.nerdgraph.nrql(ctx.accountId, query);
    
    const schemas = await Promise.all(
      results.facets.map(async (facet) => {
        const attrQuery = `SELECT keysetCount() FROM ${facet.name} LIMIT 1`;
        const attrResult = await ctx.nerdgraph.nrql(ctx.accountId, attrQuery);
        
        return {
          eventType: facet.name,
          count: facet.results[0].count,
          lastIngested: facet.results[0].lastSeen,
          attributes: attrResult.results[0].keysetCount || 0,
        };
      })
    );

    ctx.explainabilityTrace.addStep({
      type: 'query',
      description: 'Discovered event schemas using FACET eventType()',
      query,
      resultCount: schemas.length,
    });

    return {
      schemas,
      totalEvents: schemas.reduce((sum, s) => sum + s.count, 0),
    };
  },
});
```

### 8. Workflow Engine

```typescript
// src/intelligence/workflow.ts
import { z } from 'zod';
import yaml from 'js-yaml';

const WorkflowSchema = z.object({
  name: z.string(),
  description: z.string(),
  parameters: z.record(z.any()).optional(),
  steps: z.array(z.object({
    id: z.string(),
    tool: z.string(),
    params: z.record(z.any()),
    condition: z.string().optional(),
    dependsOn: z.array(z.string()).optional(),
  })),
});

export class WorkflowEngine {
  constructor(
    private toolRegistry: ToolRegistry,
    private workflowPath: string = './workflows',
  ) {}

  async executeWorkflow(
    ctx: RequestContext,
    workflowName: string,
    parameters: Record<string, any>,
  ): Promise<WorkflowResult> {
    const workflow = await this.loadWorkflow(workflowName);
    const results = new Map<string, any>();

    // Build execution graph
    const graph = this.buildExecutionGraph(workflow.steps);

    // Execute steps in topological order
    for (const step of graph.topologicalSort()) {
      if (step.condition && !this.evaluateCondition(step.condition, results)) {
        continue;
      }

      const resolvedParams = this.resolveParameters(
        step.params,
        parameters,
        results,
      );

      const tool = this.toolRegistry.getTool(step.tool);
      const result = await tool.handler(ctx, resolvedParams);
      
      results.set(step.id, result);

      ctx.explainabilityTrace.addStep({
        type: 'workflow_step',
        description: `Executed ${step.tool}`,
        stepId: step.id,
        result: result.success,
      });
    }

    return {
      workflow: workflow.name,
      steps: Array.from(results.entries()),
      explainability: ctx.explainabilityTrace,
    };
  }
}
```

### 9. Example Workflow

```yaml
# workflows/golden_signals.yaml
name: golden_signals_investigation
description: Comprehensive golden signals analysis with discovery
parameters:
  service_name:
    type: string
    required: true
  time_range:
    type: string
    default: "1 hour ago"

steps:
  - id: discover_schemas
    tool: discover.list_schemas
    params:
      since: "${time_range}"

  - id: discover_service
    tool: discover.service_identifier
    params:
      hint: "${service_name}"
    dependsOn: [discover_schemas]

  - id: error_rate
    tool: query.adaptive
    params:
      intent: error_rate
      service: "${service_name}"
      since: "${time_range}"
    dependsOn: [discover_service]

  - id: latency
    tool: query.adaptive  
    params:
      intent: latency_p95
      service: "${service_name}"
      since: "${time_range}"
    dependsOn: [discover_service]

  - id: anomalies
    tool: analyze.detect_anomalies
    params:
      metrics:
        - "${error_rate.results}"
        - "${latency.results}"
      sensitivity: medium
    dependsOn: [error_rate, latency]
```

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
- Core discovery engine with caching
- Basic MCP server with stdio transport  
- Schema and attribute discovery tools
- Comprehensive test suite

### Phase 2: Intelligence (Week 3-4)
- Adaptive Query Builder
- Service identifier detection chain
- Error indicator discovery
- Analysis tools (cost, anomaly detection)

### Phase 3: Workflows (Week 5-6)
- Workflow engine implementation
- Golden signals workflow
- Dashboard generation tools
- Alert suggestion tools

### Phase 4: Production (Week 7-8)
- HTTP transport with auth
- Self-monitoring and telemetry
- Performance optimization
- Documentation and examples

## Success Metrics

1. **Zero Assumptions**: 100% of tools use discovery-first pattern
2. **Adaptability**: Works with any NR account schema without code changes
3. **Performance**: P95 discovery latency < 200ms with warm cache
4. **Explainability**: Every response includes confidence scores and reasoning
5. **Adoption**: >80% reduction in time to create dashboards/alerts

## Conclusion

This specification delivers a production-ready MCP server that embodies the Zero Assumptions Manifesto while providing immediate practical value. The discovery-first architecture ensures it works with any New Relic environment, while the intelligent adaptation layer delivers the "magic" experience users expect from an AI-powered observability co-pilot.
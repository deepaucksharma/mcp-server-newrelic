# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status: Production-Ready TypeScript Architecture

This repository implements a **discover-first, assume-nothing MCP server** using:

- **Fastify 5** + **official `@mcp/sdk`** + **Zod v4** + **Pino** 
- **Type-safe** with strict TypeScript and runtime validation
- **Production-grade** with rate limiting, logging, and security
- **Comprehensive documentation** with implementation guides

## Current State

### What Exists
- Discovery-first philosophy and design principles
- Zero-assumptions manifesto and patterns

### What Was Removed
- 165 Go files with mostly mock implementations
- Complex build tag system creating multiple compilation paths  
- Over-engineered infrastructure (Redis, JWT, telemetry)
- Python intelligence system and extensive CI/CD infrastructure

## Implementation Philosophy

### Discovery-First Architecture
The core principle is **never assume data structure** - always discover:

```
Traditional:                    Discovery-First:
1. Assume schema               1. Discover schemas
2. Write query                 2. Explore attributes  
3. Execute                     3. Profile data
4. Handle failures             4. Build informed query
5. Retry                       5. Execute confidently
```

### Tool Design Patterns
Tools should follow these patterns:
- **Composable**: Small tools that combine into workflows
- **Adaptive**: Work with any New Relic account configuration
- **Zero-config**: No hardcoded schemas or assumptions
- **avoid-Mock**: always work with actual Newrelic Backend using .env file keys

## Planned Architecture

### Core Tool Categories
1. **Discovery Tools** - Schema exploration and profiling
2. **Query Tools** - NRQL execution and validation  
3. **Analysis Tools** - Statistical analysis and anomaly detection
4. **Dashboard Tools** - Auto-generation and management
5. **Alert Tools** - Baseline-driven alert creation
6. **Governance Tools** - Cost optimization and compliance
7. **Workflow Tools** - Multi-step automation
8. **Session Tools** - State management

### Production Architecture Stack
```
Web Framework      → Fastify 5 (top async throughput)
MCP Integration    → Official @mcp/sdk 1.9+ (full MCP spec)
Validation         → Zod v4 (schema-first, TS-native)
Logging           → Pino (<10µs per log, zero-alloc JSON)
Rate Limiting     → @fastify/rate-limit (low overhead)
Testing           → Vitest + fastify.inject (fast in-process)
```

### Production Project Structure  
```
src/
├── server.ts                    # Fastify bootstrap + plugin wiring
├── config.ts                    # Environment validation with Zod
├── plugins/
│   ├── logging.ts               # Pino + trace-id
│   ├── rateLimit.ts             # @fastify/rate-limit wrapper  
│   └── tokens.ts                # discoverToken/confirmToken JWT
├── core/
│   ├── envelope.ts              # uniform ResponseEnvelope<T>
│   ├── urn.ts                   # URN helpers (urn:domain:type:id)
│   └── errors.ts                # typed error catalogue
└── tools/
    ├── meta/                    # tool_list, tool_schema
    ├── discover/                # entities, metrics, attributes
    ├── analysis/                # query, baseline, anomaly  
    └── action/                  # alert creation (dry-run + confirm)
```

## Development Approach

### Tool Lifecycle: `discover ➜ analyse ➜ action`

**Core Principle**: Never assume data structure - always discover first

```typescript
// 1. DISCOVER: Find entities without assumptions
const entities = await discover.entities({ domain: 'APM' });
const metrics = await discover.metrics({ entityUrn });

// 2. ANALYSE: Build informed queries  
const analysis = await analysis.query({
  nrql: buildInformedQuery(entities.data, metrics.data),
  accountId
});

// 3. ACTION: Controlled mutations with confirmation
const dryRun = await action.alert.create({ 
  nrql, threshold, dry_run: true 
});
const result = await action.alert.create({
  ...dryRun.data, dry_run: false, confirmToken
});
```

### Key Architecture Patterns

**ResponseEnvelope**: All tools return uniform structure
```typescript
interface ResponseEnvelope<T> {
  ok: boolean;
  data?: T;
  error?: { code: ErrorCode; message: string };
  discoverToken?: string;    // JWT for follow-up actions  
  confirmToken?: string;     // Single-use confirmation
}
```

**URN Grammar**: Resource identification
```
urn:<domain>:<type>:<id>
Examples: urn:apm:entity:123, urn:alert:policy:abc
```

## Key Insights from Previous Implementation

### What Worked
- Basic NRQL query execution
- Event type discovery via SHOW EVENT TYPES
- Attribute discovery via keyset() function
- Mock data for development

### What Failed  
- Over-complex build tag system
- Multiple parallel server implementations
- Mock implementations more complex than real ones
- Infrastructure complexity for minimal functionality

### Lessons Learned
- Focus on 20% that provides 80% of value
- Avoid premature optimization and abstractions
- Real implementations beat sophisticated mocks
- Documentation-driven development creates clarity

## Common Patterns

### Discovery Pattern
```typescript
// Always discover before operating
async function executeTool(args: ToolArgs): Promise<ToolResult> {
    // 1. Discover available schemas
    const schemas = await discoveryEngine.exploreSchemas();
    
    // 2. Profile relevant attributes  
    const attrs = await discoveryEngine.exploreAttributes(targetSchema);
    
    // 3. Build informed operation
    const query = buildQuery(schemas, attrs, args);
    
    // 4. Execute with confidence
    return await newRelicClient.execute(query);
}
```

### Configuration
TypeScript environment configuration with validation:
```typescript
// config/environment.ts
import { z } from 'zod';

const envSchema = z.object({
  NEW_RELIC_API_KEY: z.string().min(1),
  NEW_RELIC_ACCOUNT_ID: z.string().min(1),
  NEW_RELIC_REGION: z.enum(['US', 'EU']).default('US'),
  NODE_ENV: z.enum(['development', 'production']).default('development'),
});

export const config = envSchema.parse(process.env);
```

## Testing Strategy

### TypeScript Testing Approach
- **Unit tests**: Core logic with `vitest`
- **Integration tests**: Real New Relic API integration
- **Type safety**: Full TypeScript coverage with strict mode
- **E2E tests**: MCP protocol testing with Claude Desktop integration

### Example Test Structure
```typescript
// tests/discovery.test.ts
import { describe, it, expect } from 'vitest';
import { exploreEventTypes } from '../src/tools/discovery';

describe('Discovery Tools', () => {
  it('should discover event types from real API', async () => {
    const result = await exploreEventTypes({ limit: 10 });
    expect(result.event_types).toBeDefined();
    expect(result.event_types.length).toBeGreaterThan(0);
  });
});
```


## Contributing

### Before Starting (TypeScript Development)
1. **Read the architecture documentation** in `docs/archive/`
2. **Understand discovery-first principles** - never assume data structure
3. **Set up TypeScript environment** with official MCP SDK
4. **Review tool specifications** to pick an implementation target

### Development Setup
```bash
# Initialize production-grade TypeScript project
npm install fastify @mcp/sdk zod pino @fastify/rate-limit
npm install -D typescript @types/node vitest tsx tsup

# Configure strict TypeScript
npx tsc --init --strict --target es2022 --module node16

# Development commands
npm run dev              # tsx watch src/server.ts
npm run build           # tsup build for production
npm run test            # vitest with fastify.inject
```

### Tool Registration Pattern
```typescript
import { makeTool } from '../core/tool-factory.js';

export const discoveryTool = makeTool(
  'discover.entities',
  z.object({ domain: z.string().optional() }),
  async (ctx, input) => {
    // Always discover, never assume
    const entities = await ctx.newrelic.entitySearch(input);
    return { entities, discoverToken: ctx.tokens.create() };
  },
  { description: 'Discover New Relic entities' }
);
```

### Key Benefits of This Architecture
- **Performance**: Fastify's top async throughput
- **Type Safety**: Zod runtime + TypeScript compile-time validation  
- **Security**: JWT tokens for discovery/confirmation flow
- **Observability**: Pino logging + OpenTelemetry tracing
- **Standards**: Full MCP spec compliance with official SDK
- **Testing**: Fast in-process HTTP testing with vitest

### Documentation Reference
- **[Architecture Specification](docs/ARCHITECTURE_SPECIFICATION.md)** - Complete technical specification
- **[Implementation Guide](docs/IMPLEMENTATION_GUIDE.md)** - Step-by-step building instructions
- **[Archive Documentation](docs/archive/)** - Original vision and tool specifications
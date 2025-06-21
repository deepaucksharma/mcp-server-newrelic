# Implementation Guide: Building the Discover-First MCP Server

This guide walks through implementing the **Fastify + `@mcp/sdk` + Zod** architecture specified in our [Architecture Specification](ARCHITECTURE_SPECIFICATION.md).

## Quick Start

### 1. Initialize Project

```bash
# Create new TypeScript project
npm init -y
npm install fastify @mcp/sdk zod pino @fastify/rate-limit
npm install -D typescript @types/node vitest tsx tsup

# Configure TypeScript
npx tsc --init --strict --target es2022 --module node16
```

### 2. Basic Server Setup

```typescript
// src/server.ts
import Fastify from 'fastify';
import { Server as MCPServer } from '@mcp/sdk/server';

export async function buildServer() {
  const app = Fastify({ logger: false });
  
  // Will add plugins here
  
  return app;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const server = await buildServer();
  await server.listen({ port: 3000, host: '0.0.0.0' });
  console.log('🚀 MCP Server running on http://localhost:3000');
}
```

## Core Implementation Steps

### Step 1: Response Envelope

```typescript
// src/core/envelope.ts
import { z } from 'zod';

export const ErrorCodeSchema = z.enum([
  'VALIDATION_ERROR',
  'AUTH_ERROR', 
  'NEWRELIC_ERROR',
  'RATE_LIMIT_ERROR',
  'INTERNAL_ERROR'
]);

export type ErrorCode = z.infer<typeof ErrorCodeSchema>;

export interface ResponseEnvelope<T = unknown> {
  ok: boolean;
  data?: T;
  error?: { 
    code: ErrorCode; 
    message: string;
    details?: unknown;
  };
  cursor?: string;
  discoverToken?: string;
  confirmToken?: string;
}

export function success<T>(data: T, meta?: Partial<ResponseEnvelope>): ResponseEnvelope<T> {
  return { ok: true, data, ...meta };
}

export function error(code: ErrorCode, message: string, details?: unknown): ResponseEnvelope {
  return { ok: false, error: { code, message, details } };
}
```

### Step 2: URN System

```typescript
// src/core/urn.ts
import { z } from 'zod';

export const UrnSchema = z.string().regex(
  /^urn:[a-z]+:[a-z]+:[a-zA-Z0-9-_]+$/,
  'URN must match pattern: urn:<domain>:<type>:<id>'
);

export type Urn = z.infer<typeof UrnSchema>;

export interface ParsedUrn {
  domain: string;
  type: string;
  id: string;
}

export function parseUrn(urn: Urn): ParsedUrn {
  const [, domain, type, id] = urn.split(':');
  return { domain, type, id };
}

export function buildUrn(domain: string, type: string, id: string): Urn {
  const urn = `urn:${domain}:${type}:${id}`;
  return UrnSchema.parse(urn);
}
```

### Step 3: Configuration

```typescript
// src/config.ts
import { z } from 'zod';

const ConfigSchema = z.object({
  // New Relic
  NEW_RELIC_API_KEY: z.string().min(1),
  NEW_RELIC_ACCOUNT_ID: z.string().min(1),
  NEW_RELIC_REGION: z.enum(['US', 'EU']).default('US'),
  
  // Server
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  PORT: z.coerce.number().default(3000),
  LOG_LEVEL: z.enum(['debug', 'info', 'warn', 'error']).default('info'),
  
  // Security
  JWT_SECRET: z.string().min(32),
  DISCOVER_TOKEN_TTL: z.coerce.number().default(900), // 15 minutes
  
  // Rate limiting
  RATE_LIMIT_MAX: z.coerce.number().default(120),
  RATE_LIMIT_WINDOW: z.string().default('1 minute'),
});

export const config = ConfigSchema.parse(process.env);
export type Config = z.infer<typeof ConfigSchema>;
```

### Step 4: Logging Plugin

```typescript
// src/plugins/logging.ts
import fp from 'fastify-plugin';
import pino from 'pino';
import { config } from '../config.js';

export const loggerPlugin = fp(async (app) => {
  const logger = pino({
    level: config.LOG_LEVEL,
    transport: config.NODE_ENV === 'development' ? {
      target: 'pino-pretty',
      options: { colorize: true }
    } : undefined,
  });

  app.decorate('log', logger);
  
  app.addHook('onRequest', async (req) => {
    req.log = logger.child({ 
      reqId: req.id,
      method: req.method,
      url: req.url 
    });
  });
  
  app.addHook('onResponse', async (req, reply) => {
    req.log.info({
      statusCode: reply.statusCode,
      responseTime: reply.elapsedTime
    }, 'Request completed');
  });
});
```

### Step 5: Token Management

```typescript
// src/plugins/tokens.ts
import fp from 'fastify-plugin';
import jwt from 'jsonwebtoken';
import { config } from '../config.js';

export interface DiscoverTokenPayload {
  type: 'discover';
  accountId: string;
  issued: number;
  expires: number;
}

export interface ConfirmTokenPayload {
  type: 'confirm';
  actionId: string;
  payload: unknown;
  issued: number;
  expires: number;
}

export const tokensPlugin = fp(async (app) => {
  app.decorate('tokens', {
    createDiscoverToken(accountId: string): string {
      const now = Date.now();
      const payload: DiscoverTokenPayload = {
        type: 'discover',
        accountId,
        issued: now,
        expires: now + (config.DISCOVER_TOKEN_TTL * 1000),
      };
      return jwt.sign(payload, config.JWT_SECRET);
    },

    createConfirmToken(actionId: string, payload: unknown): string {
      const now = Date.now();
      const tokenPayload: ConfirmTokenPayload = {
        type: 'confirm',
        actionId,
        payload,
        issued: now,
        expires: now + (300 * 1000), // 5 minutes
      };
      return jwt.sign(tokenPayload, config.JWT_SECRET);
    },

    verifyDiscoverToken(token: string): DiscoverTokenPayload | null {
      try {
        const payload = jwt.verify(token, config.JWT_SECRET) as DiscoverTokenPayload;
        if (payload.type !== 'discover' || Date.now() > payload.expires) {
          return null;
        }
        return payload;
      } catch {
        return null;
      }
    },

    verifyConfirmToken(token: string): ConfirmTokenPayload | null {
      try {
        const payload = jwt.verify(token, config.JWT_SECRET) as ConfirmTokenPayload;
        if (payload.type !== 'confirm' || Date.now() > payload.expires) {
          return null;
        }
        return payload;
      } catch {
        return null;
      }
    }
  });
});
```

### Step 6: Tool Registration System

```typescript
// src/core/tool-factory.ts
import { z } from 'zod';
import { FastifyRequest } from 'fastify';
import { ResponseEnvelope, success, error } from './envelope.js';

export interface ToolContext {
  log: any;
  requestId: string;
  tokens: any;
  // Add more context as needed
}

export interface ToolOptions {
  destructiveHint?: boolean;
  confirmationRequired?: boolean;
  description?: string;
}

export function makeTool<I, O>(
  name: string,
  inputSchema: z.ZodType<I>,
  handler: (ctx: ToolContext, input: I) => Promise<O>,
  options: ToolOptions = {}
) {
  return {
    name,
    inputSchema,
    handler: async (req: FastifyRequest): Promise<ResponseEnvelope<O>> => {
      try {
        // Validate input
        const parseResult = inputSchema.safeParse(req.body);
        if (!parseResult.success) {
          return error('VALIDATION_ERROR', 'Invalid input', parseResult.error);
        }

        // Build context
        const ctx: ToolContext = {
          log: req.log,
          requestId: req.id,
          tokens: (req.server as any).tokens,
        };

        // Execute handler
        const result = await handler(ctx, parseResult.data);
        return success(result);
        
      } catch (err: any) {
        req.log.error(err, `Tool ${name} failed`);
        return error('INTERNAL_ERROR', err.message);
      }
    },
    options,
  };
}
```

### Step 7: Example Meta Tools

```typescript
// src/tools/meta/tool-list.ts
import { z } from 'zod';
import { makeTool } from '../../core/tool-factory.js';

const InputSchema = z.object({});

export const toolListTool = makeTool(
  'meta.tool_list',
  InputSchema,
  async (ctx, input) => {
    // This would enumerate all registered tools
    return {
      tools: [
        { name: 'meta.tool_list', category: 'meta' },
        { name: 'meta.tool_schema', category: 'meta' },
        { name: 'discover.entities', category: 'discover' },
        { name: 'discover.metrics', category: 'discover' },
        { name: 'analysis.query', category: 'analysis' },
      ]
    };
  },
  { description: 'List all available MCP tools' }
);
```

### Step 8: Discovery Tools

```typescript
// src/tools/discover/entities.ts
import { z } from 'zod';
import { makeTool } from '../../core/tool-factory.js';
import { UrnSchema } from '../../core/urn.js';

const InputSchema = z.object({
  domain: z.string().optional(),
  name: z.string().optional(),
  type: z.string().optional(),
  cursor: z.string().optional(),
  limit: z.number().min(1).max(100).default(20),
});

export const entitiesDiscoveryTool = makeTool(
  'discover.entities',
  InputSchema,
  async (ctx, input) => {
    // This would call New Relic EntitySearch API
    // For now, return mock structure
    
    const entities = [
      {
        urn: 'urn:apm:entity:123456' as any,
        name: 'web-application',
        type: 'APPLICATION',
        domain: 'APM',
        reporting: true,
        alertSeverity: 'NOT_ALERTING'
      }
    ];

    // Generate discover token for follow-up actions
    const discoverToken = ctx.tokens.createDiscoverToken('123456');

    return {
      entities,
      cursor: entities.length === input.limit ? 'next-page-token' : undefined,
      discoverToken,
    };
  },
  { description: 'Discover New Relic entities without assumptions about structure' }
);
```

### Step 9: Analysis Tools

```typescript
// src/tools/analysis/query.ts
import { z } from 'zod';
import { makeTool } from '../../core/tool-factory.js';

const InputSchema = z.object({
  nrql: z.string().min(1),
  accountId: z.number().int().positive(),
  timeoutMs: z.number().min(1000).max(30000).default(10000),
});

export const queryAnalysisTool = makeTool(
  'analysis.query',
  InputSchema,
  async (ctx, input) => {
    // Validate NRQL has LIMIT constraint
    if (!input.nrql.toUpperCase().includes('LIMIT')) {
      input.nrql += ' LIMIT 100';
    }

    ctx.log.info({ nrql: input.nrql }, 'Executing NRQL query');
    
    // This would execute against New Relic NerdGraph
    return {
      results: [
        { timestamp: Date.now(), value: 42.0 }
      ],
      metadata: {
        query: input.nrql,
        executionTimeMs: 150,
        resultCount: 1
      }
    };
  },
  { description: 'Execute NRQL queries with automatic safety limits' }
);
```

### Step 10: Wiring Everything Together

```typescript
// src/server.ts
import Fastify from 'fastify';
import { loggerPlugin } from './plugins/logging.js';
import { tokensPlugin } from './plugins/tokens.js';
import rateLimit from '@fastify/rate-limit';
import { config } from './config.js';

// Import tools
import { toolListTool } from './tools/meta/tool-list.js';
import { entitiesDiscoveryTool } from './tools/discover/entities.js';
import { queryAnalysisTool } from './tools/analysis/query.js';

export async function buildServer() {
  const app = Fastify({ logger: false });

  // Register plugins
  await app.register(loggerPlugin);
  await app.register(tokensPlugin);
  await app.register(rateLimit, {
    max: config.RATE_LIMIT_MAX,
    timeWindow: config.RATE_LIMIT_WINDOW,
  });

  // Register tools as routes
  const tools = [
    toolListTool,
    entitiesDiscoveryTool,
    queryAnalysisTool,
  ];

  for (const tool of tools) {
    app.post(`/tools/${tool.name}`, tool.handler);
  }

  // Health check
  app.get('/health', async () => ({ status: 'healthy' }));

  return app;
}
```

## Testing Implementation

```typescript
// test/tools/meta.test.ts
import { describe, it, expect } from 'vitest';
import { buildServer } from '../src/server.js';

describe('Meta Tools', () => {
  it('should list available tools', async () => {
    const app = await buildServer();
    
    const response = await app.inject({
      method: 'POST',
      url: '/tools/meta.tool_list',
      payload: {}
    });

    expect(response.statusCode).toBe(200);
    const body = JSON.parse(response.body);
    expect(body.ok).toBe(true);
    expect(body.data.tools).toBeInstanceOf(Array);
    expect(body.data.tools.length).toBeGreaterThan(0);
  });
});
```

## Development Commands

```json
{
  "scripts": {
    "dev": "tsx watch src/server.ts",
    "build": "tsup src/server.ts --format esm --dts",
    "test": "vitest",
    "test:watch": "vitest --watch",
    "type-check": "tsc --noEmit"
  }
}
```

This implementation provides a solid foundation for the discover-first MCP server. Each tool follows the same pattern: validate input with Zod, execute business logic, return typed response wrapped in the uniform envelope.

The next steps would be:
1. Implement actual New Relic GraphQL client
2. Add more sophisticated discovery and analysis tools
3. Implement the action tools with dry-run/confirm flow
4. Add comprehensive error handling and logging
5. Set up CI/CD and deployment pipeline
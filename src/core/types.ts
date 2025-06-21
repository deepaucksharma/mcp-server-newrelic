/**
 * Core type definitions for the Zero Assumptions MCP Server
 * 
 * These types define the fundamental data structures for discovery,
 * world models, and explainability traces.
 */

import { z } from 'zod';

// ============================================================================
// Discovery Types
// ============================================================================

export const SchemaInfoSchema = z.object({
  eventType: z.string(),
  count: z.number(),
  lastIngested: z.string(),
  attributes: z.number(),
  confidence: z.number().min(0).max(1).optional(),
});

export type SchemaInfo = z.infer<typeof SchemaInfoSchema>;

export const AttributeProfileSchema = z.object({
  field: z.string(),
  eventType: z.string(),
  type: z.enum(['string', 'numeric', 'boolean', 'timestamp']),
  cardinality: z.enum(['low', 'medium', 'high']),
  nullPercentage: z.number().min(0).max(100),
  sampleValues: z.array(z.any()).max(10),
  confidence: z.number().min(0).max(1),
});

export type AttributeProfile = z.infer<typeof AttributeProfileSchema>;

export const ServiceIdentifierSchema = z.object({
  field: z.string(),
  confidence: z.number().min(0).max(1),
  coverage: z.number().min(0).max(1),
  eventType: z.string().optional(),
});

export type ServiceIdentifier = z.infer<typeof ServiceIdentifierSchema>;

export const ErrorIndicatorSchema = z.object({
  field: z.string(),
  eventType: z.string(),
  type: z.enum(['boolean', 'http_status', 'error_class', 'custom']),
  condition: z.string(),
  confidence: z.number().min(0).max(1),
  prevalence: z.number().min(0).max(1),
});

export type ErrorIndicator = z.infer<typeof ErrorIndicatorSchema>;

export const MetricInfoSchema = z.object({
  field: z.string(),
  eventType: z.string(),
  type: z.enum(['numeric', 'gauge', 'counter', 'histogram']),
  unit: z.string().optional(),
  description: z.string().optional(),
  confidence: z.number().min(0).max(1),
});

export type MetricInfo = z.infer<typeof MetricInfoSchema>;

export const DataSourceSchema = z.object({
  type: z.enum(['apm', 'infra', 'browser', 'mobile', 'logs', 'custom']),
  agent: z.string().optional(),
  version: z.string().optional(),
  lastSeen: z.string(),
  eventTypes: z.array(z.string()),
});

export type DataSource = z.infer<typeof DataSourceSchema>;

// ============================================================================
// World Model
// ============================================================================

export const DiscoveryGraphSchema = z.object({
  accountId: z.number(),
  timestamp: z.date(),
  schemas: z.array(SchemaInfoSchema),
  attributes: z.record(z.string(), AttributeProfileSchema),
  serviceIdentifier: ServiceIdentifierSchema,
  errorIndicators: z.array(ErrorIndicatorSchema),
  metrics: z.array(MetricInfoSchema),
  dataSources: z.array(DataSourceSchema),
  confidence: z.number().min(0).max(1),
});

export type DiscoveryGraph = z.infer<typeof DiscoveryGraphSchema>;

// ============================================================================
// Explainability
// ============================================================================

export const DiscoveryStepSchema = z.object({
  type: z.enum(['discovery', 'query', 'adaptive_build', 'workflow_step', 'validation', 'analysis']),
  description: z.string(),
  query: z.string().optional(),
  stepId: z.string().optional(),
  resultCount: z.number().optional(),
  confidence: z.number().min(0).max(1).optional(),
  result: z.boolean().optional(),
  duration: z.number().optional(),
});

export type DiscoveryStep = z.infer<typeof DiscoveryStepSchema>;

export interface ExplainabilityTrace {
  steps: DiscoveryStep[];
  assumptions: string[];
  confidence: number;
  
  addStep(step: DiscoveryStep): void;
  addAssumption(assumption: string): void;
  toMarkdown(): string;
}

// ============================================================================
// Request Context
// ============================================================================

export interface Logger {
  info(message: string, meta?: any): void;
  warn(message: string, meta?: any): void;
  error(message: string, meta?: any): void;
  debug(message: string, meta?: any): void;
}

export interface CacheAdapter {
  get<T>(key: string): Promise<T | undefined>;
  set<T>(key: string, value: T, options?: { ttl?: number }): Promise<void>;
  delete(key: string): Promise<void>;
  clear(): Promise<void>;
}

export interface NerdGraphClient {
  request<T = any>(query: string, variables?: Record<string, any>): Promise<T>;
  nrql(accountId: number, query: string): Promise<any>;
}

export interface TelemetryRecorder {
  recordToolExecution(toolName: string, duration: number, success: boolean): void;
  recordDiscoveryMiss(accountId: number, reason: string): void;
  recordCacheHit(cacheType: string): void;
  recordConfidenceScore(operation: string, confidence: number): void;
}

export interface RequestContext {
  requestId: string;
  accountId: number;
  logger: Logger;
  cache: CacheAdapter;
  nerdgraph: NerdGraphClient;
  telemetry: TelemetryRecorder;
  
  // Will be populated by discovery engine
  worldModel?: DiscoveryGraph;
  explainabilityTrace: ExplainabilityTrace;
}

// ============================================================================
// Tool System
// ============================================================================

export interface ToolDefinition<I = any, O = any> {
  name: string;
  description: string;
  requiresDiscovery: boolean;
  inputSchema: any; // Allow both Zod and plain objects
  outputSchema?: z.ZodSchema<O>;
  examples?: Array<{ input: I; output: O }>;
  handler: (ctx: RequestContext, input: I) => Promise<O>;
}

// ============================================================================
// Adaptive Query Builder
// ============================================================================

export const ScopeSchema = z.object({
  type: z.enum(['service', 'entity', 'account', 'custom']),
  value: z.string(),
  field: z.string().optional(),
});

export type Scope = z.infer<typeof ScopeSchema>;

export const AQBParamsSchema = z.object({
  intent: z.enum(['error_rate', 'latency_p95', 'latency_p99', 'throughput', 'saturation']),
  scope: ScopeSchema,
  timeRange: z.string().default('1 hour ago'),
  worldModel: DiscoveryGraphSchema.optional(),
});

export type AQBParams = z.infer<typeof AQBParamsSchema>;

// ============================================================================
// Workflow System
// ============================================================================

export const WorkflowStepSchema = z.object({
  id: z.string(),
  tool: z.string(),
  params: z.record(z.any()),
  condition: z.string().optional(),
  dependsOn: z.array(z.string()).optional(),
});

export type WorkflowStep = z.infer<typeof WorkflowStepSchema>;

export const WorkflowSchema = z.object({
  name: z.string(),
  description: z.string(),
  parameters: z.record(z.any()).optional(),
  steps: z.array(WorkflowStepSchema),
});

export type Workflow = z.infer<typeof WorkflowSchema>;

export interface WorkflowResult {
  workflow: string;
  steps: Array<[string, any]>;
  explainability: ExplainabilityTrace;
  success: boolean;
  error?: string;
}

// ============================================================================
// Configuration
// ============================================================================

export const ConfigSchema = z.object({
  newrelic: z.object({
    apiKey: z.string().min(1),
    accountId: z.string().min(1),
    region: z.enum(['US', 'EU']).default('US'),
    graphqlUrl: z.string().url(),
  }),
  discovery: z.object({
    cache: z.object({
      type: z.enum(['memory', 'redis', 'upstash']).default('memory'),
      ttl: z.object({
        schemas: z.number().default(4 * 60 * 60 * 1000),      // 4 hours
        attributes: z.number().default(30 * 60 * 1000),       // 30 minutes
        serviceId: z.number().default(2 * 60 * 60 * 1000),    // 2 hours
        errors: z.number().default(30 * 60 * 1000),           // 30 minutes
      }),
    }),
    confidence: z.object({
      minimum: z.number().min(0).max(1).default(0.7),
      optimal: z.number().min(0).max(1).default(0.9),
    }),
  }),
  mcp: z.object({
    transport: z.enum(['stdio', 'http']).default('stdio'),
    http: z.object({
      port: z.number().min(1000).max(65535).default(3000),
      cors: z.array(z.string()).default(['*']),
    }),
  }),
  telemetry: z.object({
    enabled: z.boolean().default(true),
    endpoint: z.string().url().optional(),
  }),
});

export type Config = z.infer<typeof ConfigSchema>;
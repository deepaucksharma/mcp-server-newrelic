/**
 * Request Context and World Model Management
 * 
 * This module provides the request context that carries discovery state,
 * explainability traces, and resource access throughout tool execution.
 */

import { nanoid } from 'nanoid';
import { 
  RequestContext, 
  ExplainabilityTrace, 
  DiscoveryStep, 
  Logger,
  CacheAdapter,
  NerdGraphClient,
  TelemetryRecorder 
} from './types.js';

/**
 * Implementation of ExplainabilityTrace for recording discovery decisions
 */
export class ExplainabilityTraceImpl implements ExplainabilityTrace {
  public steps: DiscoveryStep[] = [];
  public assumptions: string[] = [];
  public confidence: number = 1.0;

  addStep(step: DiscoveryStep): void {
    this.steps.push({
      ...step,
      duration: step.duration || 0,
    });
    
    // Update overall confidence based on step confidence
    if (step.confidence !== undefined) {
      this.confidence = Math.min(this.confidence, step.confidence);
    }
  }

  addAssumption(assumption: string): void {
    this.assumptions.push(assumption);
    // Reduce confidence when we make assumptions
    this.confidence *= 0.95;
  }

  toMarkdown(): string {
    const sections = [];
    
    sections.push(`# Explainability Trace`);
    sections.push(`**Overall Confidence**: ${(this.confidence * 100).toFixed(1)}%\n`);
    
    if (this.assumptions.length > 0) {
      sections.push(`## Assumptions Made`);
      this.assumptions.forEach((assumption, i) => {
        sections.push(`${i + 1}. ${assumption}`);
      });
      sections.push('');
    }
    
    sections.push(`## Discovery & Execution Steps`);
    this.steps.forEach((step, i) => {
      sections.push(`### ${i + 1}. ${step.type.toUpperCase()}: ${step.description}`);
      
      if (step.query) {
        sections.push('```sql');
        sections.push(step.query.trim());
        sections.push('```');
      }
      
      const details = [];
      if (step.confidence !== undefined) {
        details.push(`**Confidence**: ${(step.confidence * 100).toFixed(1)}%`);
      }
      if (step.resultCount !== undefined) {
        details.push(`**Results**: ${step.resultCount}`);
      }
      if (step.duration !== undefined) {
        details.push(`**Duration**: ${step.duration}ms`);
      }
      
      if (details.length > 0) {
        sections.push(details.join(' | '));
      }
      sections.push('');
    });
    
    return sections.join('\n');
  }
}

/**
 * Console logger implementation
 */
export class ConsoleLogger implements Logger {
  private prefix: string;

  constructor(prefix: string = '[MCP]') {
    this.prefix = prefix;
  }

  info(message: string, meta?: any): void {
    console.error(`${this.prefix} INFO: ${message}`, meta ? JSON.stringify(meta) : '');
  }

  warn(message: string, meta?: any): void {
    console.error(`${this.prefix} WARN: ${message}`, meta ? JSON.stringify(meta) : '');
  }

  error(message: string, meta?: any): void {
    console.error(`${this.prefix} ERROR: ${message}`, meta ? JSON.stringify(meta) : '');
  }

  debug(message: string, meta?: any): void {
    if (process.env['DEBUG']) {
      console.error(`${this.prefix} DEBUG: ${message}`, meta ? JSON.stringify(meta) : '');
    }
  }
}

/**
 * Simple telemetry recorder that logs to New Relic
 */
export class TelemetryRecorderImpl implements TelemetryRecorder {
  constructor(
    private logger: Logger,
    private nerdgraph?: NerdGraphClient,
    private accountId?: number,
  ) {}

  recordToolExecution(toolName: string, duration: number, success: boolean): void {
    this.logger.info('Tool execution recorded', {
      tool: toolName,
      duration,
      success,
    });

    // Send to New Relic if available
    if (this.nerdgraph && this.accountId) {
      this.sendToNewRelic({
        eventType: 'MCPToolExecution',
        tool: toolName,
        duration,
        success,
        timestamp: Date.now(),
      }).catch(err => {
        this.logger.warn('Failed to send telemetry to New Relic', err);
      });
    }
  }

  recordDiscoveryMiss(accountId: number, reason: string): void {
    this.logger.info('Discovery cache miss', { accountId, reason });
  }

  recordCacheHit(cacheType: string): void {
    this.logger.debug('Cache hit', { cacheType });
  }

  recordConfidenceScore(operation: string, confidence: number): void {
    this.logger.info('Confidence score recorded', { operation, confidence });
  }

  private async sendToNewRelic(event: any): Promise<void> {
    if (!this.nerdgraph || !this.accountId) return;

    try {
      const mutation = `
        mutation($accountId: Int!, $events: [CustomEventInput!]!) {
          eventsIngestCustomEvents(accountId: $accountId, events: $events) {
            success
          }
        }
      `;

      await this.nerdgraph.request(mutation, {
        accountId: this.accountId,
        events: [event],
      });
    } catch (error) {
      // Silent fail for telemetry
      this.logger.debug('Telemetry send failed', error);
    }
  }
}

/**
 * Factory function to create a request context
 */
export interface CreateContextOptions {
  accountId: number;
  apiKey: string;
  region?: 'US' | 'EU';
  cache?: CacheAdapter;
  logger?: Logger;
  nerdgraph?: NerdGraphClient;
}

export async function createRequestContext(options: CreateContextOptions): Promise<RequestContext> {
  const requestId = nanoid();
  const logger = options.logger || new ConsoleLogger(`[${requestId.slice(0, 8)}]`);
  
  // Create cache adapter if not provided
  const cache = options.cache || createMemoryCache();
  
  // Create NerdGraph client if not provided
  const nerdgraph = options.nerdgraph || createNerdGraphClient({
    apiKey: options.apiKey,
    region: options.region || 'US',
  });

  // Create telemetry recorder
  const telemetry = new TelemetryRecorderImpl(logger, nerdgraph, options.accountId);

  return {
    requestId,
    accountId: options.accountId,
    logger,
    cache,
    nerdgraph,
    telemetry,
    explainabilityTrace: new ExplainabilityTraceImpl(),
  };
}

/**
 * Simple in-memory cache implementation
 */
function createMemoryCache(): CacheAdapter {
  const cache = new Map<string, { value: any; expires: number }>();

  return {
    async get<T>(key: string): Promise<T | undefined> {
      const entry = cache.get(key);
      if (!entry) return undefined;
      
      if (Date.now() > entry.expires) {
        cache.delete(key);
        return undefined;
      }
      
      return entry.value as T;
    },

    async set<T>(key: string, value: T, options?: { ttl?: number }): Promise<void> {
      const ttl = options?.ttl || 5 * 60 * 1000; // 5 minutes default
      const expires = Date.now() + ttl;
      cache.set(key, { value, expires });
    },

    async delete(key: string): Promise<void> {
      cache.delete(key);
    },

    async clear(): Promise<void> {
      cache.clear();
    },
  };
}

/**
 * Create a NerdGraph client
 */
function createNerdGraphClient(options: { apiKey: string; region: 'US' | 'EU' }): NerdGraphClient {
  // Import and use the actual implementation
  const { createNerdGraphClient: createClient } = require('../adapters/nerdgraph.js');
  return createClient(options);
}
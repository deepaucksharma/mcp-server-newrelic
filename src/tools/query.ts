/**
 * Query Tools - Adaptive NRQL Generation
 * 
 * These tools implement the Adaptive Query Builder (AQB) that generates
 * intelligent NRQL queries based on discovered world model.
 */

import { z } from 'zod';
import { ToolDefinition, RequestContext, AQBParams } from '../core/types.js';
import { DiscoveryEngine } from '../core/discovery/engine.js';

// ============================================================================
// Input/Output Schemas
// ============================================================================

const GenerateQueryInputSchema = z.object({
  intent: z.enum(['error_rate', 'latency_p95', 'latency_p99', 'throughput', 'saturation']),
  scope: z.object({
    type: z.enum(['service', 'entity', 'account', 'custom']),
    value: z.string(),
    field: z.string().optional(),
  }),
  timeRange: z.string().default('1 hour ago'),
  accountId: z.number().optional(),
});

const CustomQueryInputSchema = z.object({
  query: z.string(),
  accountId: z.number().optional(),
  explain: z.boolean().default(true),
});

const QueryOptimizeInputSchema = z.object({
  query: z.string(),
  accountId: z.number().optional(),
});

const SuggestQueriesInputSchema = z.object({
  context: z.string(),
  limit: z.number().min(1).max(10).default(5),
  accountId: z.number().optional(),
});

// ============================================================================
// Adaptive Query Builder Implementation
// ============================================================================

class AdaptiveQueryBuilder {
  constructor(private ctx: RequestContext) {}

  /**
   * Generate NRQL query based on intent and discovered world model
   */
  async generateQuery(params: AQBParams): Promise<{ query: string; confidence: number; explanation: string }> {
    // Use world model if available, otherwise build it
    let worldModel = params.worldModel || this.ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(this.ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(this.ctx.accountId);
    }

    const serviceField = worldModel.serviceIdentifier.field;
    const errorIndicators = worldModel.errorIndicators;
    const metrics = worldModel.metrics;

    // Build scope filter
    let scopeFilter = '';
    let scopeConfidence = 1.0;

    switch (params.scope.type) {
      case 'service':
        scopeFilter = `WHERE ${serviceField} = '${params.scope.value}'`;
        scopeConfidence = worldModel.serviceIdentifier.confidence;
        break;
      case 'entity':
        const entityField = params.scope.field || 'entity.name';
        scopeFilter = `WHERE ${entityField} = '${params.scope.value}'`;
        scopeConfidence = 0.8; // Medium confidence for entity scoping
        break;
      case 'custom':
        scopeFilter = params.scope.field ? 
          `WHERE ${params.scope.field} = '${params.scope.value}'` :
          `WHERE ${params.scope.value}`;
        scopeConfidence = 0.6; // Lower confidence for custom scoping
        break;
      case 'account':
        scopeFilter = '';
        scopeConfidence = 1.0;
        break;
    }

    // Generate query based on intent
    let query = '';
    let explanation = '';
    let intentConfidence = 0.5;

    switch (params.intent) {
      case 'error_rate':
        if (errorIndicators.length > 0) {
          const errorIndicator = errorIndicators[0];
          query = `
            SELECT percentage(count(*), WHERE ${errorIndicator.condition}) as errorRate
            FROM ${errorIndicator.eventType}
            ${scopeFilter}
            SINCE ${params.timeRange}
            TIMESERIES
          `.trim();
          intentConfidence = errorIndicator.confidence;
          explanation = `Using discovered error indicator: ${errorIndicator.field} with condition ${errorIndicator.condition}`;
        } else {
          query = `
            SELECT percentage(count(*), WHERE error = true) as errorRate
            FROM Transaction
            ${scopeFilter}
            SINCE ${params.timeRange}
            TIMESERIES
          `.trim();
          intentConfidence = 0.3;
          explanation = 'No error indicators discovered, using standard error field assumption';
          
          this.ctx.explainabilityTrace.addAssumption(
            'Assumed standard error field since no error indicators were discovered'
          );
        }
        break;

      case 'latency_p95':
      case 'latency_p99':
        const percentile = params.intent === 'latency_p95' ? 95 : 99;
        const durationMetrics = metrics.filter(m => 
          m.field.includes('duration') || m.field.includes('response') || m.field.includes('latency')
        );
        
        if (durationMetrics.length > 0) {
          const durationField = durationMetrics[0].field;
          query = `
            SELECT percentile(${durationField}, ${percentile}) as p${percentile}
            FROM ${durationMetrics[0].eventType}
            ${scopeFilter}
            SINCE ${params.timeRange}
            TIMESERIES
          `.trim();
          intentConfidence = durationMetrics[0].confidence;
          explanation = `Using discovered latency metric: ${durationField}`;
        } else {
          query = `
            SELECT percentile(duration, ${percentile}) as p${percentile}
            FROM Transaction
            ${scopeFilter}
            SINCE ${params.timeRange}
            TIMESERIES
          `.trim();
          intentConfidence = 0.3;
          explanation = 'No latency metrics discovered, using standard duration field assumption';
          
          this.ctx.explainabilityTrace.addAssumption(
            'Assumed standard duration field since no latency metrics were discovered'
          );
        }
        break;

      case 'throughput':
        const primarySchema = worldModel.schemas[0]; // Highest volume schema
        if (primarySchema) {
          query = `
            SELECT count(*) as throughput
            FROM ${primarySchema.eventType}
            ${scopeFilter}
            SINCE ${params.timeRange}
            TIMESERIES
          `.trim();
          intentConfidence = primarySchema.confidence || 0.8;
          explanation = `Using primary event type: ${primarySchema.eventType} (highest volume)`;
        } else {
          query = `
            SELECT count(*) as throughput
            FROM Transaction
            ${scopeFilter}
            SINCE ${params.timeRange}
            TIMESERIES
          `.trim();
          intentConfidence = 0.3;
          explanation = 'No schemas discovered, using Transaction assumption';
        }
        break;

      case 'saturation':
        const systemMetrics = metrics.filter(m => 
          m.field.includes('cpu') || m.field.includes('memory') || m.field.includes('disk')
        );
        
        if (systemMetrics.length > 0) {
          const saturationField = systemMetrics[0].field;
          query = `
            SELECT average(${saturationField}) as saturation
            FROM ${systemMetrics[0].eventType}
            ${scopeFilter}
            SINCE ${params.timeRange}
            TIMESERIES
          `.trim();
          intentConfidence = systemMetrics[0].confidence;
          explanation = `Using discovered system metric: ${saturationField}`;
        } else {
          query = `
            SELECT average(cpuPercent) as cpuSaturation
            FROM SystemSample
            ${scopeFilter}
            SINCE ${params.timeRange}
            TIMESERIES
          `.trim();
          intentConfidence = 0.3;
          explanation = 'No system metrics discovered, using standard CPU metric assumption';
        }
        break;
    }

    // Calculate overall confidence
    const overallConfidence = (scopeConfidence + intentConfidence + worldModel.confidence) / 3;

    return {
      query,
      confidence: overallConfidence,
      explanation,
    };
  }
}

// ============================================================================
// Tool Implementations
// ============================================================================

/**
 * Generate adaptive NRQL query based on intent
 */
const generateQuery: ToolDefinition<
  z.infer<typeof GenerateQueryInputSchema>,
  any
> = {
  name: 'query.generate',
  description: 'Generate intelligent NRQL queries using Adaptive Query Builder (AQB)',
  requiresDiscovery: true,
  inputSchema: GenerateQueryInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;
    
    ctx.explainabilityTrace.addStep({
      type: 'adaptive_build',
      description: `Generating ${input.intent} query for ${input.scope.type}: ${input.scope.value}`,
      confidence: 1.0,
    });

    const aqb = new AdaptiveQueryBuilder(ctx);
    const result = await aqb.generateQuery({
      intent: input.intent,
      scope: input.scope,
      timeRange: input.timeRange,
    });

    ctx.explainabilityTrace.addStep({
      type: 'adaptive_build',
      description: `Generated query with ${(result.confidence * 100).toFixed(1)}% confidence`,
      query: result.query,
      confidence: result.confidence,
    });

    return {
      accountId,
      intent: input.intent,
      scope: input.scope,
      timeRange: input.timeRange,
      generated: {
        query: result.query,
        confidence: result.confidence,
        explanation: result.explanation,
      },
      usage: {
        copyPaste: result.query,
        curlExample: `curl -H "Api-Key: YOUR_API_KEY" -G "https://api.newrelic.com/graphql" --data-urlencode 'query=query{actor{account(id:${accountId}){nrql(query:"${result.query.replace(/"/g, '\\"').replace(/\n/g, ' ')}")}}}'`,
      },
    };
  },
};

/**
 * Execute custom NRQL query with explanation
 */
const executeQuery: ToolDefinition<
  z.infer<typeof CustomQueryInputSchema>,
  any
> = {
  name: 'query.execute',
  description: 'Execute custom NRQL query with optional explanation',
  requiresDiscovery: false,
  inputSchema: CustomQueryInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    ctx.explainabilityTrace.addStep({
      type: 'query',
      description: 'Executing custom NRQL query',
      query: input.query,
      confidence: 0.8,
    });

    try {
      const startTime = Date.now();
      const result = await ctx.nerdgraph.nrql(accountId, input.query);
      const duration = Date.now() - startTime;

      const response = {
        accountId,
        query: input.query,
        executionTime: `${duration}ms`,
        results: result.results || [],
        metadata: result.metadata || {},
        resultCount: result.results?.length || 0,
      };

      if (input.explain) {
        // Simple query analysis
        const analysis = this.analyzeQuery(input.query);
        response.explanation = analysis;
      }

      ctx.explainabilityTrace.addStep({
        type: 'query',
        description: `Query executed successfully in ${duration}ms`,
        resultCount: response.resultCount,
        duration,
        confidence: 0.9,
      });

      return response;

    } catch (error: any) {
      ctx.explainabilityTrace.addStep({
        type: 'query',
        description: `Query execution failed: ${error.message}`,
        confidence: 0.1,
      });

      throw new Error(`Query execution failed: ${error.message}`);
    }
  },

  analyzeQuery(query: string) {
    const analysis = {
      type: 'unknown',
      eventTypes: [],
      functions: [],
      clauses: [],
      complexity: 'simple',
    };

    // Basic query analysis
    if (query.toLowerCase().includes('select')) {
      analysis.type = 'select';
    }

    // Extract FROM clause
    const fromMatch = query.match(/FROM\s+([A-Za-z0-9_,\s]+)/i);
    if (fromMatch) {
      analysis.eventTypes = fromMatch[1].split(',').map(et => et.trim());
    }

    // Extract functions
    const functionMatches = query.match(/\b(count|average|sum|min|max|percentile|uniqueCount|histogram)\s*\(/gi);
    if (functionMatches) {
      analysis.functions = [...new Set(functionMatches.map(f => f.replace(/\s*\(.*/, '').toLowerCase()))];
    }

    // Extract clauses
    if (query.toLowerCase().includes('where')) analysis.clauses.push('WHERE');
    if (query.toLowerCase().includes('facet')) analysis.clauses.push('FACET');
    if (query.toLowerCase().includes('timeseries')) analysis.clauses.push('TIMESERIES');
    if (query.toLowerCase().includes('since')) analysis.clauses.push('SINCE');
    if (query.toLowerCase().includes('until')) analysis.clauses.push('UNTIL');

    // Determine complexity
    if (analysis.clauses.length > 3 || analysis.functions.length > 2) {
      analysis.complexity = 'complex';
    } else if (analysis.clauses.length > 1 || analysis.functions.length > 0) {
      analysis.complexity = 'moderate';
    }

    return analysis;
  },
};

/**
 * Optimize NRQL query performance
 */
const optimizeQuery: ToolDefinition<
  z.infer<typeof QueryOptimizeInputSchema>,
  any
> = {
  name: 'query.optimize',
  description: 'Analyze and suggest optimizations for NRQL queries',
  requiresDiscovery: false,
  inputSchema: QueryOptimizeInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const suggestions = [];
    const query = input.query.toLowerCase();

    // Basic optimization suggestions
    if (!query.includes('since') && !query.includes('until')) {
      suggestions.push({
        type: 'performance',
        issue: 'No time constraint',
        suggestion: 'Add SINCE clause to limit time range for better performance',
        example: 'Add "SINCE 1 hour ago" to your query',
      });
    }

    if (!query.includes('limit') && query.includes('facet')) {
      suggestions.push({
        type: 'performance',
        issue: 'Unlimited FACET results',
        suggestion: 'Add LIMIT clause to FACET queries',
        example: 'Add "LIMIT 50" to control result size',
      });
    }

    if (query.includes('select *')) {
      suggestions.push({
        type: 'performance',
        issue: 'SELECT * usage',
        suggestion: 'Specify specific fields instead of SELECT *',
        example: 'SELECT specific_field1, specific_field2 instead of SELECT *',
      });
    }

    if (query.match(/where.*like.*%.*%/)) {
      suggestions.push({
        type: 'performance',
        issue: 'Inefficient LIKE pattern',
        suggestion: 'Avoid leading wildcards in LIKE patterns',
        example: 'Use field LIKE "prefix%" instead of field LIKE "%pattern%"',
      });
    }

    return {
      query: input.query,
      accountId: input.accountId || ctx.accountId,
      suggestions,
      optimizationScore: suggestions.length === 0 ? 100 : Math.max(20, 100 - (suggestions.length * 20)),
    };
  },
};

/**
 * Suggest relevant queries based on context
 */
const suggestQueries: ToolDefinition<
  z.infer<typeof SuggestQueriesInputSchema>,
  any
> = {
  name: 'query.suggest',
  description: 'Suggest relevant NRQL queries based on context and discovered data',
  requiresDiscovery: true,
  inputSchema: SuggestQueriesInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    // Build or use existing world model
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    const suggestions = [];
    const context = input.context.toLowerCase();

    // Context-based suggestions using discovered data
    if (context.includes('error') || context.includes('failure')) {
      if (worldModel.errorIndicators.length > 0) {
        const errorIndicator = worldModel.errorIndicators[0];
        suggestions.push({
          intent: 'error_analysis',
          query: `SELECT percentage(count(*), WHERE ${errorIndicator.condition}) as errorRate FROM ${errorIndicator.eventType} SINCE 1 hour ago TIMESERIES`,
          description: `Error rate using discovered indicator: ${errorIndicator.field}`,
          confidence: errorIndicator.confidence,
        });
      }
    }

    if (context.includes('performance') || context.includes('latency') || context.includes('slow')) {
      const durationMetrics = worldModel.metrics.filter(m => m.field.includes('duration'));
      if (durationMetrics.length > 0) {
        const metric = durationMetrics[0];
        suggestions.push({
          intent: 'performance_analysis',
          query: `SELECT average(${metric.field}), percentile(${metric.field}, 95) FROM ${metric.eventType} SINCE 1 hour ago TIMESERIES`,
          description: `Performance analysis using discovered metric: ${metric.field}`,
          confidence: metric.confidence,
        });
      }
    }

    if (context.includes('service') || context.includes('application')) {
      const serviceField = worldModel.serviceIdentifier.field;
      suggestions.push({
        intent: 'service_overview',
        query: `SELECT count(*) as throughput FROM Transaction FACET ${serviceField} SINCE 1 hour ago`,
        description: `Service throughput using discovered identifier: ${serviceField}`,
        confidence: worldModel.serviceIdentifier.confidence,
      });
    }

    // Add generic suggestions based on top schemas
    if (suggestions.length < input.limit) {
      const topSchemas = worldModel.schemas.slice(0, Math.min(3, input.limit - suggestions.length));
      topSchemas.forEach(schema => {
        suggestions.push({
          intent: 'data_exploration',
          query: `SELECT count(*) as events FROM ${schema.eventType} SINCE 1 hour ago TIMESERIES`,
          description: `Explore ${schema.eventType} data (${schema.count} recent records)`,
          confidence: schema.confidence || 0.7,
        });
      });
    }

    return {
      accountId,
      context: input.context,
      totalSuggestions: suggestions.length,
      suggestions: suggestions.slice(0, input.limit),
      worldModelConfidence: worldModel.confidence,
    };
  },
};

// ============================================================================
// Export Tools
// ============================================================================

export function createQueryTools(): ToolDefinition[] {
  return [
    generateQuery,
    executeQuery,
    optimizeQuery,
    suggestQueries,
  ];
}
/**
 * Discovery Tools - Foundation of Zero Assumptions Philosophy
 * 
 * These tools discover the structure and characteristics of New Relic data
 * without making assumptions about schema, field names, or data patterns.
 */

import { z } from 'zod';
import { ToolDefinition, RequestContext } from '../core/types.js';
import { DiscoveryEngine } from '../core/discovery/engine.js';

// ============================================================================
// Input/Output Schemas
// ============================================================================

const ListSchemasInputSchema = z.object({
  accountId: z.number().optional(),
  refreshCache: z.boolean().default(false),
});

const ServiceIdentifierInputSchema = z.object({
  accountId: z.number().optional(),
  refreshCache: z.boolean().default(false),
});

const AttributeProfileInputSchema = z.object({
  eventType: z.string(),
  field: z.string().optional(),
  accountId: z.number().optional(),
});

const ErrorIndicatorsInputSchema = z.object({
  accountId: z.number().optional(),
  refreshCache: z.boolean().default(false),
});

const MetricsDiscoveryInputSchema = z.object({
  eventType: z.string().optional(),
  accountId: z.number().optional(),
});

// ============================================================================
// Tool Implementations
// ============================================================================

/**
 * Discover all event types (schemas) in the account
 */
const discoverSchemas: ToolDefinition<
  z.infer<typeof ListSchemasInputSchema>,
  any
> = {
  name: 'discover.list_schemas',
  description: 'Discover all event types and their characteristics in New Relic account without assumptions',
  requiresDiscovery: true,
  inputSchema: ListSchemasInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;
    
    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: 'Starting schema discovery - no assumptions about event types',
      confidence: 1.0,
    });

    if (input.refreshCache) {
      await ctx.cache.delete(`discovery:${accountId}`);
      ctx.logger.info('Discovery cache cleared for refresh');
    }

    const discoveryEngine = new DiscoveryEngine(ctx);
    const worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);

    const result = {
      accountId,
      timestamp: worldModel.timestamp,
      totalSchemas: worldModel.schemas.length,
      confidence: worldModel.confidence,
      schemas: worldModel.schemas.map(schema => ({
        eventType: schema.eventType,
        recordCount: schema.count,
        lastIngested: schema.lastIngested,
        attributeCount: schema.attributes,
        confidence: schema.confidence,
        dataFreshness: schema.lastIngested,
      })),
      discoveryNotes: [
        'Schemas discovered through systematic enumeration',
        'No assumptions made about event type names or structure',
        `Confidence score: ${(worldModel.confidence * 100).toFixed(1)}%`,
      ],
    };

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: `Discovered ${worldModel.schemas.length} schemas with ${(worldModel.confidence * 100).toFixed(1)}% confidence`,
      resultCount: worldModel.schemas.length,
      confidence: worldModel.confidence,
    });

    return result;
  },
};

/**
 * Discover the best field for service identification
 */
const discoverServiceIdentifier: ToolDefinition<
  z.infer<typeof ServiceIdentifierInputSchema>,
  any
> = {
  name: 'discover.service_identifier',
  description: 'Discover which field best identifies services across event types',
  requiresDiscovery: true,
  inputSchema: ServiceIdentifierInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: 'Discovering service identifier without assuming field names',
      confidence: 1.0,
    });

    if (input.refreshCache) {
      await ctx.cache.delete(`discovery:${accountId}`);
    }

    // Use world model if available, otherwise build it
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    const serviceId = worldModel.serviceIdentifier;

    const result = {
      accountId,
      discovered: {
        field: serviceId.field,
        eventType: serviceId.eventType,
        confidence: serviceId.confidence,
        coverage: serviceId.coverage,
      },
      usageGuidance: serviceId.confidence > 0.8 
        ? `High confidence: Use '${serviceId.field}' for service identification`
        : `Low confidence: '${serviceId.field}' may not be reliable for service identification`,
      alternatives: serviceId.confidence < 0.8 ? [
        'Consider using entity.name if available',
        'Look for custom service tags in attributes',
        'Check for application-specific naming conventions',
      ] : [],
      explainability: {
        method: 'Systematic evaluation of common service identifier patterns',
        confidence: serviceId.confidence,
        coverage: `${(serviceId.coverage * 100).toFixed(1)}% of records contain this field`,
      },
    };

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: `Service identifier: ${serviceId.field} (${(serviceId.confidence * 100).toFixed(1)}% confidence)`,
      confidence: serviceId.confidence,
    });

    return result;
  },
};

/**
 * Profile attributes for specific event types
 */
const discoverAttributes: ToolDefinition<
  z.infer<typeof AttributeProfileInputSchema>,
  any
> = {
  name: 'discover.attribute_profile',
  description: 'Profile attributes for event types to understand data characteristics',
  requiresDiscovery: false,
  inputSchema: AttributeProfileInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: `Profiling attributes for ${input.eventType}`,
      confidence: 1.0,
    });

    try {
      // Get keyset for the event type
      const keysetQuery = `SELECT keyset() FROM ${input.eventType} LIMIT 1`;
      const keysetResult = await ctx.nerdgraph.nrql(accountId, keysetQuery);
      const allAttributes = keysetResult.results[0]?.['keyset()'] || [];

      let attributesToProfile = allAttributes;
      
      // If specific field requested, profile just that one
      if (input.field) {
        attributesToProfile = allAttributes.includes(input.field) ? [input.field] : [];
        if (attributesToProfile.length === 0) {
          throw new Error(`Field '${input.field}' not found in ${input.eventType}`);
        }
      } else {
        // Limit to first 10 attributes to avoid overwhelming response
        attributesToProfile = allAttributes.slice(0, 10);
      }

      const profiles = [];

      for (const attr of attributesToProfile) {
        try {
          const profileQuery = `
            SELECT 
              count(*) as total,
              uniqueCount(${attr}) as cardinality,
              latest(${attr}) as sample,
              filter(count(*), WHERE ${attr} IS NOT NULL) as nonNull
            FROM ${input.eventType} 
            SINCE 1 hour ago
            LIMIT 1
          `;

          const result = await ctx.nerdgraph.nrql(accountId, profileQuery);
          const data = result.results[0] || {};

          // Determine type based on sample
          const sample = data.sample;
          let type: 'string' | 'numeric' | 'boolean' | 'timestamp' = 'string';
          
          if (typeof sample === 'number') {
            type = 'numeric';
          } else if (typeof sample === 'boolean') {
            type = 'boolean';
          } else if (attr.includes('timestamp') || attr.includes('time')) {
            type = 'timestamp';
          }

          // Calculate cardinality
          const cardinalityRatio = data.cardinality / data.total;
          let cardinality: 'low' | 'medium' | 'high' = 'high';
          if (cardinalityRatio < 0.01) cardinality = 'low';
          else if (cardinalityRatio < 0.1) cardinality = 'medium';

          const nullPercentage = ((data.total - data.nonNull) / data.total) * 100;

          profiles.push({
            field: attr,
            type,
            cardinality,
            nullPercentage: nullPercentage.toFixed(1),
            sampleValue: sample,
            uniqueValues: data.cardinality,
            totalRecords: data.total,
            confidence: data.total > 100 ? 0.9 : 0.6,
          });

        } catch (error) {
          ctx.logger.warn(`Failed to profile attribute ${attr}`, error);
          profiles.push({
            field: attr,
            type: 'unknown',
            cardinality: 'unknown',
            error: error.message,
            confidence: 0.1,
          });
        }
      }

      const result = {
        eventType: input.eventType,
        accountId,
        totalAttributes: allAttributes.length,
        profiledAttributes: profiles.length,
        profiles,
        discoveryNotes: [
          'Attribute profiles based on recent data (1 hour)',
          'Type inference based on sample values',
          'Cardinality classifications: low (<1%), medium (1-10%), high (>10%)',
        ],
      };

      ctx.explainabilityTrace.addStep({
        type: 'discovery',
        description: `Profiled ${profiles.length} attributes for ${input.eventType}`,
        resultCount: profiles.length,
        confidence: profiles.length > 0 ? 0.9 : 0.3,
      });

      return result;

    } catch (error: any) {
      ctx.explainabilityTrace.addStep({
        type: 'discovery',
        description: `Failed to profile attributes: ${error.message}`,
        confidence: 0.1,
      });

      throw new Error(`Attribute profiling failed: ${error.message}`);
    }
  },
};

/**
 * Discover error indicators across event types
 */
const discoverErrorIndicators: ToolDefinition<
  z.infer<typeof ErrorIndicatorsInputSchema>,
  any
> = {
  name: 'discover.error_indicators',
  description: 'Discover fields and patterns that indicate errors or failures',
  requiresDiscovery: true,
  inputSchema: ErrorIndicatorsInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: 'Discovering error indicators without assuming error field names',
      confidence: 1.0,
    });

    if (input.refreshCache) {
      await ctx.cache.delete(`discovery:${accountId}`);
    }

    // Use world model if available, otherwise build it
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    const indicators = worldModel.errorIndicators;

    const result = {
      accountId,
      totalIndicators: indicators.length,
      indicators: indicators.map(indicator => ({
        field: indicator.field,
        eventType: indicator.eventType,
        type: indicator.type,
        condition: indicator.condition,
        confidence: indicator.confidence,
        prevalence: `${(indicator.prevalence * 100).toFixed(2)}%`,
        usageExample: `WHERE ${indicator.condition}`,
      })),
      recommendations: indicators.length > 0 ? [
        `Primary error indicator: ${indicators[0].field} (${(indicators[0].confidence * 100).toFixed(1)}% confidence)`,
        'Use discovered conditions in NRQL queries for error analysis',
        'Combine multiple indicators for comprehensive error detection',
      ] : [
        'No reliable error indicators found',
        'Consider checking for custom error fields',
        'Look for HTTP status codes or exception patterns',
      ],
      explainability: {
        method: 'Pattern-based discovery of common error indicators',
        coverage: `${indicators.length} patterns evaluated`,
        confidence: indicators.length > 0 ? indicators[0].confidence : 0.1,
      },
    };

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: `Discovered ${indicators.length} error indicators`,
      resultCount: indicators.length,
      confidence: indicators.length > 0 ? indicators[0].confidence : 0.1,
    });

    return result;
  },
};

/**
 * Discover metrics and their characteristics
 */
const discoverMetrics: ToolDefinition<
  z.infer<typeof MetricsDiscoveryInputSchema>,
  any
> = {
  name: 'discover.metrics',
  description: 'Discover numeric fields suitable for metrics and monitoring',
  requiresDiscovery: true,
  inputSchema: MetricsDiscoveryInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: 'Discovering metrics without assuming field names',
      confidence: 1.0,
    });

    // Use world model if available, otherwise build it
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    let metrics = worldModel.metrics;

    // Filter by event type if specified
    if (input.eventType) {
      metrics = metrics.filter(m => m.eventType === input.eventType);
    }

    const result = {
      accountId,
      eventTypeFilter: input.eventType,
      totalMetrics: metrics.length,
      metrics: metrics.map(metric => ({
        field: metric.field,
        eventType: metric.eventType,
        type: metric.type,
        confidence: metric.confidence,
        usageExamples: [
          `SELECT average(${metric.field}) FROM ${metric.eventType}`,
          `SELECT max(${metric.field}) FROM ${metric.eventType} FACET appName`,
          `SELECT histogram(${metric.field}) FROM ${metric.eventType}`,
        ],
      })),
      categories: {
        performance: metrics.filter(m => 
          m.field.includes('duration') || 
          m.field.includes('response') || 
          m.field.includes('latency')
        ).length,
        system: metrics.filter(m => 
          m.field.includes('cpu') || 
          m.field.includes('memory') || 
          m.field.includes('disk')
        ).length,
        custom: metrics.filter(m => 
          !m.field.includes('duration') && 
          !m.field.includes('cpu') && 
          !m.field.includes('memory')
        ).length,
      },
      explainability: {
        method: 'Automatic detection of numeric fields suitable for aggregation',
        confidence: metrics.length > 0 ? 
          metrics.reduce((sum, m) => sum + m.confidence, 0) / metrics.length : 0.1,
      },
    };

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: `Discovered ${metrics.length} metric fields`,
      resultCount: metrics.length,
      confidence: result.explainability.confidence,
    });

    return result;
  },
};

// ============================================================================
// Export Tools
// ============================================================================

export function createDiscoveryTools(): ToolDefinition[] {
  return [
    discoverSchemas,
    discoverServiceIdentifier,
    discoverAttributes,
    discoverErrorIndicators,
    discoverMetrics,
  ];
}
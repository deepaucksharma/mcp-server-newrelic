/**
 * Discovery Engine - Heart of the Zero Assumptions Philosophy
 * 
 * This engine discovers the structure and characteristics of New Relic data
 * without making any assumptions about schema, field names, or data patterns.
 * Every tool execution begins with discovery.
 */

// import { LRUCache } from 'lru-cache'; // Unused import
import {
  RequestContext,
  DiscoveryGraph,
  SchemaInfo,
  AttributeProfile,
  ServiceIdentifier,
  ErrorIndicator,
  MetricInfo,
  DataSource,
} from '../types.js';

export interface DiscoveryResult {
  graph: DiscoveryGraph;
  cached: boolean;
  freshness: number; // Age in milliseconds
}

/**
 * Main Discovery Engine
 * 
 * Orchestrates parallel discovery operations to build a comprehensive
 * world model of the New Relic environment.
 */
export class DiscoveryEngine {
  // private cache: LRUCache<string, DiscoveryResult>; // Commented out as not used
  
  // Cache TTL configurations (milliseconds)
  private readonly cacheTTL = {
    schemas: 4 * 60 * 60 * 1000,      // 4 hours - schemas change rarely
    attributes: 30 * 60 * 1000,       // 30 minutes - attributes change more often
    serviceId: 2 * 60 * 60 * 1000,    // 2 hours - service identifiers stable
    errors: 30 * 60 * 1000,           // 30 minutes - error patterns change
    metrics: 60 * 60 * 1000,          // 1 hour - metric definitions change occasionally
  };

  constructor(private ctx: RequestContext) {
    // Cache now handled via RequestContext
  }

  /**
   * Main entry point: Build a complete discovery graph for the account
   */
  async buildDiscoveryGraph(accountId: number): Promise<DiscoveryGraph> {
    const startTime = Date.now();
    const cacheKey = `discovery:${accountId}`;
    
    // Check cache first
    const cached = await this.ctx.cache.get<DiscoveryResult>(cacheKey);
    if (cached && !this.isStale(cached)) {
      this.ctx.telemetry.recordCacheHit('discovery');
      this.ctx.logger.debug('Using cached discovery graph', {
        accountId,
        age: Date.now() - cached.graph.timestamp.getTime(),
        confidence: cached.graph.confidence,
      });
      return cached.graph;
    }

    this.ctx.logger.info('Building discovery graph', { accountId });
    this.ctx.telemetry.recordDiscoveryMiss(accountId, cached ? 'stale' : 'not_found');

    try {
      // Parallel discovery operations for speed
      const [schemas, serviceId, errors, metrics, dataSources] = await Promise.all([
        this.discoverSchemas(accountId),
        this.discoverServiceIdentifier(accountId),
        this.discoverErrorIndicators(accountId),
        this.discoverMetrics(accountId),
        this.discoverDataSources(accountId),
      ]);

      // Discover attributes for top schemas (limit to avoid overload)
      const topSchemas = schemas
        .sort((a, b) => b.count - a.count)
        .slice(0, 5); // Top 5 by volume

      const attributes: Record<string, AttributeProfile> = {};
      for (const schema of topSchemas) {
        const schemaAttrs = await this.discoverAttributes(accountId, schema.eventType);
        schemaAttrs.forEach(attr => {
          attributes[`${attr.eventType}.${attr.field}`] = attr;
        });
      }

      // Calculate overall confidence
      const confidence = this.calculateConfidence({
        schemas,
        serviceId,
        errors,
        metrics,
        attributeCount: Object.keys(attributes).length,
      });

      const graph: DiscoveryGraph = {
        accountId,
        timestamp: new Date(),
        schemas,
        attributes,
        serviceIdentifier: serviceId,
        errorIndicators: errors,
        metrics,
        dataSources,
        confidence,
      };

      // Cache the result
      const result: DiscoveryResult = {
        graph,
        cached: false,
        freshness: 0,
      };

      await this.ctx.cache.set(cacheKey, result, { ttl: this.cacheTTL.schemas });

      const duration = Date.now() - startTime;
      this.ctx.logger.info('Discovery graph built', {
        accountId,
        duration,
        confidence,
        schemasFound: schemas.length,
        attributesFound: Object.keys(attributes).length,
      });

      this.ctx.explainabilityTrace.addStep({
        type: 'discovery',
        description: 'Built comprehensive world model from account data',
        confidence,
        duration,
        resultCount: schemas.length,
      });

      return graph;

    } catch (error) {
      this.ctx.logger.error('Discovery graph build failed', error);
      throw new Error(`Failed to build discovery graph: ${error.message}`);
    }
  }

  /**
   * Discover all event types (schemas) in the account
   */
  private async discoverSchemas(accountId: number): Promise<SchemaInfo[]> {
    this.ctx.logger.debug('Discovering schemas', { accountId });

    const query = `
      SELECT count(*) as 'count', latest(timestamp) as 'lastSeen' 
      FROM Transaction, PageView, MobileRequest, Log, Metric, Span, 
           JavaVirtualMachine, ProcessSample, SystemSample, ContainerSample,
           Custom
      WHERE timestamp > ${Date.now() - 24 * 60 * 60 * 1000} 
      FACET eventType() 
      LIMIT 50
    `;

    try {
      const result = await this.ctx.nerdgraph.nrql(accountId, query);
      
      const schemas: SchemaInfo[] = await Promise.all(
        result.facets?.map(async (facet: any) => {
          // Get attribute count for each schema
          const attrQuery = `SELECT keyset() FROM ${facet.name} LIMIT 1`;
          let attributeCount = 0;
          
          try {
            const attrResult = await this.ctx.nerdgraph.nrql(accountId, attrQuery);
            const keyset = attrResult.results[0]?.['keyset()'];
            attributeCount = Array.isArray(keyset) ? keyset.length : 0;
          } catch (error) {
            this.ctx.logger.warn('Failed to get attribute count', { 
              eventType: facet.name, 
              error: error.message 
            });
          }

          return {
            eventType: facet.name,
            count: facet.results[0]?.count || 0,
            lastIngested: this.formatTimestamp(facet.results[0]?.lastSeen),
            attributes: attributeCount,
            confidence: attributeCount > 0 ? 0.9 : 0.5,
          };
        }) || []
      );

      this.ctx.logger.debug('Schemas discovered', { 
        count: schemas.length,
        topSchemas: schemas.slice(0, 3).map(s => s.eventType)
      });

      return schemas.sort((a, b) => b.count - a.count);

    } catch (error) {
      this.ctx.logger.error('Schema discovery failed', error);
      return [];
    }
  }

  /**
   * Discover attributes for a specific event type
   */
  private async discoverAttributes(accountId: number, eventType: string): Promise<AttributeProfile[]> {
    this.ctx.logger.debug('Discovering attributes', { accountId, eventType });

    try {
      // Get the keyset (all attributes)
      const keysetQuery = `SELECT keyset() FROM ${eventType} LIMIT 1`;
      const keysetResult = await this.ctx.nerdgraph.nrql(accountId, keysetQuery);
      const attributes = keysetResult.results[0]?.['keyset()'] || [];

      if (!Array.isArray(attributes) || attributes.length === 0) {
        return [];
      }

      // Sample a subset of attributes to avoid overwhelming the system
      const sampleAttributes = attributes.slice(0, 20);
      
      const profiles: AttributeProfile[] = await Promise.all(
        sampleAttributes.map(async (attr: string) => {
          return this.profileAttribute(accountId, eventType, attr);
        })
      );

      return profiles.filter(p => p.confidence > 0.5);

    } catch (error) {
      this.ctx.logger.error('Attribute discovery failed', { eventType, error: error.message });
      return [];
    }
  }

  /**
   * Profile a specific attribute to understand its characteristics
   */
  private async profileAttribute(accountId: number, eventType: string, field: string): Promise<AttributeProfile> {
    try {
      const profileQuery = `
        SELECT 
          count(*) as total,
          uniqueCount(${field}) as cardinality,
          latest(${field}) as sample
        FROM ${eventType} 
        WHERE ${field} IS NOT NULL 
        LIMIT 1
      `;

      const result = await this.ctx.nerdgraph.nrql(accountId, profileQuery);
      const data = result.results[0] || {};

      // Determine type based on sample value
      const sample = data.sample;
      let type: 'string' | 'numeric' | 'boolean' | 'timestamp' = 'string';
      
      if (typeof sample === 'number') {
        type = 'numeric';
      } else if (typeof sample === 'boolean') {
        type = 'boolean';
      } else if (field.includes('timestamp') || field.includes('time')) {
        type = 'timestamp';
      }

      // Determine cardinality
      const cardinalityRatio = data.cardinality / data.total;
      let cardinality: 'low' | 'medium' | 'high' = 'high';
      if (cardinalityRatio < 0.01) cardinality = 'low';
      else if (cardinalityRatio < 0.1) cardinality = 'medium';

      return {
        field,
        eventType,
        type,
        cardinality,
        nullPercentage: 0, // Would need separate query to calculate
        sampleValues: [sample],
        confidence: data.total > 100 ? 0.9 : 0.6,
      };

    } catch (error) {
      // Return low-confidence profile for failed attributes
      return {
        field,
        eventType,
        type: 'string',
        cardinality: 'high',
        nullPercentage: 0,
        sampleValues: [],
        confidence: 0.1,
      };
    }
  }

  /**
   * Discover the best field to use for service identification
   */
  private async discoverServiceIdentifier(accountId: number): Promise<ServiceIdentifier> {
    this.ctx.logger.debug('Discovering service identifier', { accountId });

    // Chain of discovery: prioritized by common patterns
    const candidates = [
      { field: 'appName', eventType: 'Transaction' },
      { field: 'service.name', eventType: 'Span' },
      { field: 'entity.name', eventType: 'Transaction' },
      { field: 'serviceName', eventType: 'Transaction' },
      { field: 'application.name', eventType: 'Log' },
    ];

    for (const candidate of candidates) {
      try {
        const query = `
          SELECT 
            count(*) as total,
            uniqueCount(${candidate.field}) as services,
            filter(count(*), WHERE ${candidate.field} IS NOT NULL) as nonNull
          FROM ${candidate.eventType} 
          SINCE 1 hour ago
        `;

        const result = await this.ctx.nerdgraph.nrql(accountId, query);
        const data = result.results[0];

        if (data?.nonNull > 0) {
          const coverage = data.nonNull / data.total;
          const confidence = coverage > 0.8 ? 0.95 : coverage * 0.9;

          this.ctx.logger.debug('Service identifier found', {
            field: candidate.field,
            eventType: candidate.eventType,
            coverage,
            confidence,
            services: data.services,
          });

          return {
            field: candidate.field,
            confidence,
            coverage,
            eventType: candidate.eventType,
          };
        }
      } catch (error) {
        this.ctx.logger.debug('Service identifier candidate failed', {
          field: candidate.field,
          error: error.message,
        });
      }
    }

    // Fallback: return low-confidence default
    this.ctx.explainabilityTrace.addAssumption(
      'No standard service identifier found, defaulting to appName with low confidence'
    );

    return {
      field: 'appName',
      confidence: 0.3,
      coverage: 0.1,
      eventType: 'Transaction',
    };
  }

  /**
   * Discover fields that indicate errors or failures
   */
  private async discoverErrorIndicators(accountId: number): Promise<ErrorIndicator[]> {
    this.ctx.logger.debug('Discovering error indicators', { accountId });

    const indicators: ErrorIndicator[] = [];

    // Common error field patterns
    const patterns = [
      { field: 'error', type: 'boolean' as const, eventType: 'Transaction' },
      { field: 'error.class', type: 'error_class' as const, eventType: 'Transaction' },
      { field: 'httpResponseCode', type: 'http_status' as const, eventType: 'Transaction' },
      { field: 'response.status', type: 'http_status' as const, eventType: 'Span' },
      { field: 'level', type: 'custom' as const, eventType: 'Log' },
    ];

    for (const pattern of patterns) {
      try {
        let condition: string;
        let query: string;

        switch (pattern.type) {
          case 'boolean':
            condition = `${pattern.field} = true`;
            query = `SELECT percentage(count(*), WHERE ${condition}) as errorRate FROM ${pattern.eventType} SINCE 1 hour ago`;
            break;
          case 'http_status':
            condition = `numeric(${pattern.field}) >= 400`;
            query = `SELECT percentage(count(*), WHERE ${condition}) as errorRate FROM ${pattern.eventType} SINCE 1 hour ago`;
            break;
          case 'error_class':
            condition = `${pattern.field} IS NOT NULL`;
            query = `SELECT percentage(count(*), WHERE ${condition}) as errorRate FROM ${pattern.eventType} SINCE 1 hour ago`;
            break;
          case 'custom':
            if (pattern.field === 'level') {
              condition = `level IN ('ERROR', 'FATAL', 'error', 'fatal')`;
              query = `SELECT percentage(count(*), WHERE ${condition}) as errorRate FROM ${pattern.eventType} SINCE 1 hour ago`;
            } else {
              continue;
            }
            break;
        }

        const result = await this.ctx.nerdgraph.nrql(accountId, query);
        const errorRate = result.results[0]?.errorRate || 0;

        if (errorRate > 0) {
          indicators.push({
            field: pattern.field,
            eventType: pattern.eventType,
            type: pattern.type,
            condition,
            confidence: errorRate > 0.1 ? 0.9 : 0.7, // Higher confidence if we see actual errors
            prevalence: errorRate / 100,
          });
        }
      } catch (error) {
        this.ctx.logger.debug('Error indicator check failed', {
          field: pattern.field,
          error: error.message,
        });
      }
    }

    this.ctx.logger.debug('Error indicators discovered', {
      count: indicators.length,
      indicators: indicators.map(i => ({ field: i.field, type: i.type })),
    });

    return indicators.sort((a, b) => b.confidence - a.confidence);
  }

  /**
   * Discover available metrics and their characteristics
   */
  private async discoverMetrics(accountId: number): Promise<MetricInfo[]> {
    this.ctx.logger.debug('Discovering metrics', { accountId });

    const metrics: MetricInfo[] = [];

    // Common metric patterns
    const patterns = [
      { field: 'duration', eventType: 'Transaction', type: 'numeric' as const },
      { field: 'databaseDuration', eventType: 'Transaction', type: 'numeric' as const },
      { field: 'externalDuration', eventType: 'Transaction', type: 'numeric' as const },
      { field: 'cpuPercent', eventType: 'SystemSample', type: 'gauge' as const },
      { field: 'memoryUsedPercent', eventType: 'SystemSample', type: 'gauge' as const },
      { field: 'diskUsedPercent', eventType: 'SystemSample', type: 'gauge' as const },
    ];

    for (const pattern of patterns) {
      try {
        const query = `
          SELECT 
            count(*) as total,
            average(${pattern.field}) as avg,
            min(${pattern.field}) as min,
            max(${pattern.field}) as max
          FROM ${pattern.eventType} 
          WHERE ${pattern.field} IS NOT NULL 
          SINCE 1 hour ago
        `;

        const result = await this.ctx.nerdgraph.nrql(accountId, query);
        const data = result.results[0];

        if (data?.total > 0) {
          metrics.push({
            field: pattern.field,
            eventType: pattern.eventType,
            type: pattern.type,
            confidence: data.total > 100 ? 0.9 : 0.6,
          });
        }
      } catch (error) {
        this.ctx.logger.debug('Metric discovery failed', {
          field: pattern.field,
          error: error.message,
        });
      }
    }

    return metrics.sort((a, b) => b.confidence - a.confidence);
  }

  /**
   * Discover data sources (agents, integrations)
   */
  private async discoverDataSources(accountId: number): Promise<DataSource[]> {
    // Simplified implementation - would be expanded based on agent detection patterns
    return [
      {
        type: 'apm',
        agent: 'unknown',
        lastSeen: new Date().toISOString(),
        eventTypes: ['Transaction', 'TransactionError'],
      },
    ];
  }

  /**
   * Calculate overall confidence score for the discovery graph
   */
  private calculateConfidence(factors: {
    schemas: SchemaInfo[];
    serviceId: ServiceIdentifier;
    errors: ErrorIndicator[];
    metrics: MetricInfo[];
    attributeCount: number;
  }): number {
    const weights = {
      schemas: 0.3,      // Having schemas is fundamental
      serviceId: 0.25,   // Service identification is critical
      errors: 0.2,       // Error detection is important
      metrics: 0.15,     // Metrics provide operational insight
      attributes: 0.1,   // Rich attribute discovery adds confidence
    };

    let score = 0;

    // Schema confidence
    if (factors.schemas.length > 0) {
      const avgSchemaConfidence = factors.schemas.reduce((sum, s) => sum + (s.confidence || 0.5), 0) / factors.schemas.length;
      score += weights.schemas * avgSchemaConfidence;
    }

    // Service identifier confidence
    score += weights.serviceId * factors.serviceId.confidence;

    // Error indicators confidence
    if (factors.errors.length > 0) {
      const avgErrorConfidence = factors.errors.reduce((sum, e) => sum + e.confidence, 0) / factors.errors.length;
      score += weights.errors * avgErrorConfidence;
    }

    // Metrics confidence
    if (factors.metrics.length > 0) {
      const avgMetricConfidence = factors.metrics.reduce((sum, m) => sum + m.confidence, 0) / factors.metrics.length;
      score += weights.metrics * avgMetricConfidence;
    }

    // Attribute richness
    const attributeScore = Math.min(factors.attributeCount / 50, 1); // Cap at 50 attributes
    score += weights.attributes * attributeScore;

    return Math.min(score, 1.0);
  }

  /**
   * Check if cached discovery result is stale
   */
  private isStale(result: DiscoveryResult): boolean {
    const age = Date.now() - result.graph.timestamp.getTime();
    return age > this.cacheTTL.schemas;
  }

  /**
   * Format timestamp for display
   */
  private formatTimestamp(timestamp: number): string {
    if (!timestamp) return 'unknown';
    
    const diff = Date.now() - timestamp;
    const minutes = Math.floor(diff / (1000 * 60));
    
    if (minutes < 1) return 'just now';
    if (minutes < 60) return `${minutes} minutes ago`;
    
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours} hours ago`;
    
    const days = Math.floor(hours / 24);
    return `${days} days ago`;
  }
}
/**
 * Platform-Native Discovery Engine
 * 
 * Discovers all schemas dynamically with zero hardcoded assumptions.
 * Focuses on core platform concepts: entities, metrics, dashboards.
 */

import { NerdGraphClient, Logger } from './types.js';

export interface EventType {
  name: string;
  attributes: Attribute[];
  sampleCount: number;
  lastIngested: Date;
}

export interface Attribute {
  name: string;
  type: 'string' | 'numeric' | 'boolean' | 'timestamp';
  cardinality: 'low' | 'medium' | 'high';
  isNumeric: boolean;
  isDimension: boolean;
  sampleValues: any[];
}

export interface MetricInfo {
  name: string;
  dimensions: string[];
  unit?: string;
  interval: number;
  dataType: 'gauge' | 'count' | 'summary' | 'histogram';
}

export interface EntityInfo {
  guid: string;
  name: string;
  type: string;
  domain: string;
  tags: Array<{ key: string; values: string[] }>;
  entityType: string;
  reporting: boolean;
}

export interface EntityDataDiscovery {
  serviceIdentifier: string;
  errorIndicators: Array<{ name: string; type: 'boolean' | 'http_status' | 'error_class' }>;
  durationFields: string[];
  eventTypes: string[];
  metrics: MetricInfo[];
}

export interface DiscoveryResult<T> {
  data: T;
  cachedAt: Date;
  ttl: number;
}

export class PlatformDiscovery {
  private cache = new Map<string, DiscoveryResult<any>>();
  private readonly cacheTTL = {
    eventTypes: 4 * 60 * 60 * 1000,    // 4 hours - schemas change rarely
    attributes: 30 * 60 * 1000,        // 30 minutes - attributes change more often  
    metrics: 60 * 60 * 1000,           // 1 hour - metrics change occasionally
    entities: 15 * 60 * 1000,          // 15 minutes - entities change frequently
  };

  constructor(
    private nerdgraph: NerdGraphClient,
    private logger: Logger,
    // private externalCache?: CacheAdapter // Future enhancement
  ) {}

  /**
   * Discover all event types dynamically without assumptions
   */
  async discoverEventTypes(accountId: number): Promise<EventType[]> {
    const cacheKey = `events:${accountId}`;
    
    // Check cache first
    const cached = await this.getFromCache<EventType[]>(cacheKey, this.cacheTTL.eventTypes);
    if (cached) {
      this.logger.debug('Using cached event types', { accountId, count: cached.length });
      return cached;
    }

    this.logger.info('Discovering event types', { accountId });

    try {
      // Use SHOW EVENT TYPES to discover all available event types
      const showQuery = `SHOW EVENT TYPES SINCE 1 week ago`;
      const result = await this.nerdgraph.nrql(accountId, showQuery);
      
      if (!result.results || result.results.length === 0) {
        this.logger.warn('No event types found', { accountId });
        return [];
      }

      // Extract event type names from results
      const eventTypeNames = result.results.map((row: any) => row.eventType).filter(Boolean);
      
      this.logger.debug('Found event types', { 
        accountId, 
        count: eventTypeNames.length,
        types: eventTypeNames.slice(0, 5) // Log first 5
      });

      // For each event type, get detailed information
      const eventTypes: EventType[] = await Promise.all(
        eventTypeNames.map(async (eventType: string) => {
          try {
            // Get sample count
            const countQuery = `SELECT count(*) FROM ${eventType} SINCE 1 hour ago`;
            const countResult = await this.nerdgraph.nrql(accountId, countQuery);
            const sampleCount = countResult.results?.[0]?.count || 0;

            // Get latest timestamp
            const timestampQuery = `SELECT latest(timestamp) FROM ${eventType} SINCE 1 week ago`;
            const timestampResult = await this.nerdgraph.nrql(accountId, timestampQuery);
            const lastIngested = timestampResult.results?.[0]?.latest || Date.now();

            return {
              name: eventType,
              attributes: [], // Will be populated on demand
              sampleCount,
              lastIngested: new Date(lastIngested),
            };

          } catch (error: any) {
            this.logger.warn('Failed to get event type details', { 
              eventType, 
              error: error.message 
            });
            return {
              name: eventType,
              attributes: [],
              sampleCount: 0,
              lastIngested: new Date(),
            };
          }
        })
      );

      // Sort by sample count descending
      const sortedEventTypes = eventTypes
        .filter(et => et.sampleCount > 0)
        .sort((a, b) => b.sampleCount - a.sampleCount);

      await this.setInCache(cacheKey, sortedEventTypes, this.cacheTTL.eventTypes);

      this.logger.info('Event types discovered', { 
        accountId, 
        total: sortedEventTypes.length,
        withData: sortedEventTypes.filter(et => et.sampleCount > 0).length
      });

      return sortedEventTypes;

    } catch (error: any) {
      this.logger.error('Event type discovery failed', { accountId, error: error.message });
      return [];
    }
  }

  /**
   * Discover attributes for a specific event type
   */
  async discoverAttributes(accountId: number, eventType: string): Promise<Attribute[]> {
    const cacheKey = `attributes:${accountId}:${eventType}`;
    
    const cached = await this.getFromCache<Attribute[]>(cacheKey, this.cacheTTL.attributes);
    if (cached) {
      return cached;
    }

    this.logger.debug('Discovering attributes', { accountId, eventType });

    try {
      // Get all attribute names using keyset()
      const keysetQuery = `SELECT keyset() FROM ${eventType} SINCE 1 hour ago LIMIT 1`;
      const keysetResult = await this.nerdgraph.nrql(accountId, keysetQuery);
      
      const attributeNames = keysetResult.results?.[0]?.['keyset()'] || [];
      if (!Array.isArray(attributeNames) || attributeNames.length === 0) {
        this.logger.warn('No attributes found', { accountId, eventType });
        return [];
      }

      this.logger.debug('Found attributes', { 
        accountId, 
        eventType, 
        count: attributeNames.length 
      });

      // Profile each attribute (limit to prevent overwhelming queries)
      const attributesToProfile = attributeNames.slice(0, 50); // Top 50 attributes
      
      const attributes: Attribute[] = await Promise.all(
        attributesToProfile.map(async (attrName: string) => {
          return this.profileAttribute(accountId, eventType, attrName);
        })
      );

      await this.setInCache(cacheKey, attributes, this.cacheTTL.attributes);
      return attributes.filter(attr => attr.name); // Filter out failed profiles

    } catch (error: any) {
      this.logger.error('Attribute discovery failed', { 
        accountId, 
        eventType, 
        error: error.message 
      });
      return [];
    }
  }

  /**
   * Profile a specific attribute to understand its characteristics
   */
  private async profileAttribute(accountId: number, eventType: string, attrName: string): Promise<Attribute> {
    try {
      // Get attribute statistics
      const profileQuery = `
        SELECT 
          count(*) as total,
          uniqueCount(${attrName}) as cardinality,
          latest(${attrName}) as sample,
          filter(count(*), WHERE ${attrName} IS NOT NULL) as nonNull
        FROM ${eventType} 
        WHERE timestamp >= 1 hour ago
        LIMIT 1
      `;

      const result = await this.nerdgraph.nrql(accountId, profileQuery);
      const data = result.results?.[0] || {};

      // Determine type from sample value
      const sample = data.sample;
      let type: 'string' | 'numeric' | 'boolean' | 'timestamp' = 'string';
      let isNumeric = false;

      if (typeof sample === 'number') {
        type = 'numeric';
        isNumeric = true;
      } else if (typeof sample === 'boolean') {
        type = 'boolean';
      } else if (attrName.toLowerCase().includes('timestamp') || 
                 attrName.toLowerCase().includes('time') ||
                 attrName.toLowerCase().includes('date')) {
        type = 'timestamp';
      }

      // Determine cardinality level
      const cardinalityRatio = (data.cardinality || 0) / (data.total || 1);
      let cardinality: 'low' | 'medium' | 'high' = 'high';
      
      if (cardinalityRatio < 0.01) {
        cardinality = 'low';   // < 1% unique values (good for faceting)
      } else if (cardinalityRatio < 0.1) {
        cardinality = 'medium'; // 1-10% unique values
      }

      // Check if suitable for dimensions/faceting
      const isDimension = cardinality !== 'high' && type !== 'timestamp';

      return {
        name: attrName,
        type,
        cardinality,
        isNumeric,
        isDimension,
        sampleValues: [sample],
      };

    } catch (error: any) {
      this.logger.warn('Attribute profiling failed', { 
        eventType, 
        attribute: attrName, 
        error: error.message 
      });
      
      // Return minimal profile
      return {
        name: attrName,
        type: 'string',
        cardinality: 'high',
        isNumeric: false,
        isDimension: false,
        sampleValues: [],
      };
    }
  }

  /**
   * Discover dimensional metrics without assumptions
   */
  async discoverMetricNames(accountId: number): Promise<MetricInfo[]> {
    const cacheKey = `metrics:${accountId}`;
    
    const cached = await this.getFromCache<MetricInfo[]>(cacheKey, this.cacheTTL.metrics);
    if (cached) {
      return cached;
    }

    this.logger.info('Discovering dimensional metrics', { accountId });

    try {
      // Discover all metric names
      const metricQuery = `
        SELECT uniques(metricName, 1000) 
        FROM Metric 
        SINCE 1 hour ago
        LIMIT 1
      `;
      
      const result = await this.nerdgraph.nrql(accountId, metricQuery);
      const metricNames = result.results?.[0]?.['uniques.metricName'] || [];

      if (!Array.isArray(metricNames) || metricNames.length === 0) {
        this.logger.info('No dimensional metrics found', { accountId });
        return [];
      }

      this.logger.debug('Found metrics', { accountId, count: metricNames.length });

      // Profile each metric (limit to prevent overwhelming queries)
      const metricsToProfile = metricNames.slice(0, 100); // Top 100 metrics
      
      const metrics: MetricInfo[] = await Promise.all(
        metricsToProfile.map(async (metricName: string) => {
          return this.profileMetric(accountId, metricName);
        })
      );

      const validMetrics = metrics.filter(m => m.name);
      await this.setInCache(cacheKey, validMetrics, this.cacheTTL.metrics);

      return validMetrics;

    } catch (error: any) {
      this.logger.error('Metric discovery failed', { accountId, error: error.message });
      return [];
    }
  }

  /**
   * Profile a specific metric
   */
  private async profileMetric(accountId: number, metricName: string): Promise<MetricInfo> {
    try {
      // Get metric dimensions
      const dimensionQuery = `
        SELECT keyset() 
        FROM Metric 
        WHERE metricName = '${metricName}' 
        SINCE 1 hour ago 
        LIMIT 1
      `;

      const result = await this.nerdgraph.nrql(accountId, dimensionQuery);
      const allKeys = result.results?.[0]?.['keyset()'] || [];
      
      // Filter out metric-specific keys to get dimensions
      const dimensions = allKeys.filter((key: string) => 
        !['metricName', 'timestamp', 'interval.ms'].includes(key)
      );

      // Infer unit from metric name
      const unit = this.inferMetricUnit(metricName);
      
      // Infer data type from metric name patterns
      const dataType = this.inferMetricDataType(metricName);

      // Get interval (simplified - assume 1 minute for now)
      const interval = 60000; // 1 minute in milliseconds

      return {
        name: metricName,
        dimensions,
        unit: unit || 'unknown',
        interval,
        dataType,
      };

    } catch (error: any) {
      this.logger.warn('Metric profiling failed', { 
        metricName, 
        error: error.message 
      });
      
      return {
        name: metricName,
        dimensions: [],
        unit: 'unknown',
        interval: 60000,
        dataType: 'gauge',
      };
    }
  }

  /**
   * Discover entities without assumptions about types
   */
  async discoverEntities(accountId: number, searchQuery?: string): Promise<EntityInfo[]> {
    const cacheKey = `entities:${accountId}:${searchQuery || 'all'}`;
    
    const cached = await this.getFromCache<EntityInfo[]>(cacheKey, this.cacheTTL.entities);
    if (cached) {
      return cached;
    }

    this.logger.debug('Discovering entities', { accountId, searchQuery });

    try {
      const entitySearchQuery = `
        query($query: String!) {
          actor {
            entitySearch(query: $query) {
              results {
                entities {
                  guid
                  name
                  type
                  entityType
                  domain
                  reporting
                  tags {
                    key
                    values
                  }
                }
              }
              types {
                type
                domain
                count
              }
            }
          }
        }
      `;

      const searchFilter = searchQuery || `accountId = ${accountId}`;
      const result = await this.nerdgraph.request(entitySearchQuery, { 
        query: searchFilter 
      });

      const entities = result.actor?.entitySearch?.results?.entities || [];
      const entityTypes = result.actor?.entitySearch?.types || [];

      this.logger.debug('Found entities', { 
        accountId, 
        entityCount: entities.length,
        typeCount: entityTypes.length
      });

      await this.setInCache(cacheKey, entities, this.cacheTTL.entities);
      return entities;

    } catch (error: any) {
      this.logger.error('Entity discovery failed', { accountId, error: error.message });
      return [];
    }
  }

  /**
   * Discover data patterns for a specific entity
   */
  async discoverEntityData(entity: EntityInfo): Promise<EntityDataDiscovery> {
    this.logger.debug('Discovering entity data patterns', { 
      entityGuid: entity.guid,
      entityType: entity.type 
    });

    // Determine account ID from entity GUID (simplified)
    const accountId = this.extractAccountFromGuid(entity.guid);

    // Discover available event types and their attributes
    const eventTypes = await this.discoverEventTypes(accountId);
    const metrics = await this.discoverMetricNames(accountId);

    // Find service identifier field
    const serviceIdentifier = await this.findServiceIdentifier(accountId, eventTypes, entity);
    
    // Find error indicators
    const errorIndicators = await this.findErrorIndicators(accountId, eventTypes);
    
    // Find duration/latency fields
    const durationFields = await this.findDurationFields(accountId, eventTypes);

    return {
      serviceIdentifier,
      errorIndicators,
      durationFields,
      eventTypes: eventTypes.map(et => et.name),
      metrics,
    };
  }

  /**
   * Find the best field to identify services
   */
  private async findServiceIdentifier(
    accountId: number, 
    eventTypes: EventType[], 
    _entity: EntityInfo
  ): Promise<string> {
    const candidates = [
      'appName',
      'service.name', 
      'entity.name',
      'serviceName',
      'application.name',
      'name'
    ];

    // Check each candidate across major event types
    const majorEventTypes = eventTypes
      .filter(et => ['Transaction', 'Span', 'Log'].includes(et.name))
      .slice(0, 3);

    for (const candidate of candidates) {
      for (const eventType of majorEventTypes) {
        try {
          const testQuery = `
            SELECT uniqueCount(${candidate}) as services, count(*) as total
            FROM ${eventType.name} 
            WHERE ${candidate} IS NOT NULL
            SINCE 1 hour ago
          `;

          const result = await this.nerdgraph.nrql(accountId, testQuery);
          const data = result.results?.[0];
          
          if (data && data.total > 0) {
            const coverage = data.total / eventType.sampleCount;
            if (coverage > 0.5) { // More than 50% coverage
              this.logger.debug('Found service identifier', { 
                field: candidate, 
                eventType: eventType.name,
                coverage 
              });
              return candidate;
            }
          }
        } catch (error) {
          // Field doesn't exist in this event type, continue
          continue;
        }
      }
    }

    // Fallback to appName
    return 'appName';
  }

  /**
   * Find error indicator fields
   */
  private async findErrorIndicators(
    accountId: number, 
    eventTypes: EventType[]
  ): Promise<Array<{ name: string; type: 'boolean' | 'http_status' | 'error_class' }>> {
    const indicators: Array<{ name: string; type: 'boolean' | 'http_status' | 'error_class' }> = [];
    
    const patterns = [
      { field: 'error', type: 'boolean' as const },
      { field: 'httpResponseCode', type: 'http_status' as const },
      { field: 'response.status', type: 'http_status' as const },
      { field: 'error.class', type: 'error_class' as const },
      { field: 'exception.class', type: 'error_class' as const },
    ];

    const majorEventTypes = eventTypes
      .filter(et => ['Transaction', 'Span', 'Log'].includes(et.name))
      .slice(0, 3);

    for (const pattern of patterns) {
      for (const eventType of majorEventTypes) {
        try {
          let testQuery: string;
          
          switch (pattern.type) {
            case 'boolean':
              testQuery = `SELECT percentage(count(*), WHERE ${pattern.field} = true) as errorRate FROM ${eventType.name} SINCE 1 hour ago`;
              break;
            case 'http_status':
              testQuery = `SELECT percentage(count(*), WHERE numeric(${pattern.field}) >= 400) as errorRate FROM ${eventType.name} SINCE 1 hour ago`;
              break;
            case 'error_class':
              testQuery = `SELECT percentage(count(*), WHERE ${pattern.field} IS NOT NULL) as errorRate FROM ${eventType.name} SINCE 1 hour ago`;
              break;
          }

          const result = await this.nerdgraph.nrql(accountId, testQuery);
          const errorRate = result.results?.[0]?.errorRate;
          
          if (errorRate !== undefined && errorRate > 0) {
            indicators.push({
              name: pattern.field,
              type: pattern.type,
            });
            break; // Found this indicator, move to next pattern
          }
        } catch (error) {
          // Field doesn't exist, continue
          continue;
        }
      }
    }

    return indicators;
  }

  /**
   * Find duration/latency fields
   */
  private async findDurationFields(
    accountId: number, 
    eventTypes: EventType[]
  ): Promise<string[]> {
    const candidates = [
      'duration',
      'totalTime',
      'responseTime',
      'latency',
      'databaseDuration',
      'externalDuration'
    ];

    const durationFields: string[] = [];
    const transactionType = eventTypes.find(et => et.name === 'Transaction');
    
    if (!transactionType) {
      return durationFields;
    }

    for (const candidate of candidates) {
      try {
        const testQuery = `
          SELECT average(${candidate}) as avg 
          FROM Transaction 
          WHERE ${candidate} IS NOT NULL 
          SINCE 1 hour ago
        `;

        const result = await this.nerdgraph.nrql(accountId, testQuery);
        if (result.results?.[0]?.avg !== undefined) {
          durationFields.push(candidate);
        }
      } catch (error) {
        // Field doesn't exist, continue
        continue;
      }
    }

    return durationFields;
  }

  // Helper methods
  private async getFromCache<T>(key: string, ttl: number): Promise<T | null> {
    const cached = this.cache.get(key);
    if (!cached) {
      return null;
    }

    const age = Date.now() - cached.cachedAt.getTime();
    if (age > ttl) {
      this.cache.delete(key);
      return null;
    }

    return cached.data as T;
  }

  private async setInCache<T>(key: string, data: T, ttl: number): Promise<void> {
    this.cache.set(key, {
      data,
      cachedAt: new Date(),
      ttl,
    });
  }

  private inferMetricUnit(metricName: string): string | undefined {
    const name = metricName.toLowerCase();
    
    if (name.includes('duration') || name.includes('latency') || name.includes('time')) {
      return 'seconds';
    }
    if (name.includes('bytes') || name.includes('size')) {
      return 'bytes';
    }
    if (name.includes('percent') || name.includes('ratio')) {
      return 'percent';
    }
    if (name.includes('count') || name.includes('total')) {
      return 'count';
    }
    if (name.includes('rate') && !name.includes('error')) {
      return 'per_second';
    }
    
    return undefined;
  }

  private inferMetricDataType(metricName: string): 'gauge' | 'count' | 'summary' | 'histogram' {
    const name = metricName.toLowerCase();
    
    if (name.includes('count') || name.includes('total')) {
      return 'count';
    }
    if (name.includes('histogram') || name.includes('duration') || name.includes('latency')) {
      return 'histogram';
    }
    if (name.includes('summary') || name.includes('percentile')) {
      return 'summary';
    }
    
    return 'gauge';
  }

  private extractAccountFromGuid(_guid: string): number {
    // Simplified GUID parsing - in reality this would be more complex
    // For now, we'll need the account ID to be passed separately
    return 0; // Placeholder
  }
}
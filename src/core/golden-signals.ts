/**
 * Golden Signals Intelligence Engine
 * 
 * Implements the four golden signals of monitoring with OpenTelemetry awareness:
 * - Latency (Response Time)
 * - Traffic (Throughput) 
 * - Errors (Error Rate)
 * - Saturation (Resource Usage)
 */

import { NerdGraphClient, Logger } from './types.js';

export interface AnalyticalMetadata {
  dataQuality: {
    completeness: number; // 0-1 scale
    consistency: number; // 0-1 scale
    volumeStability: 'stable' | 'increasing' | 'decreasing' | 'volatile';
    lastDataPoint: Date;
  };
  seasonality: {
    detected: boolean;
    pattern?: 'daily' | 'weekly' | 'monthly';
    confidence: number; // 0-1 scale
  };
  anomalies: {
    detected: boolean;
    type?: 'spike' | 'drop' | 'trend_change' | 'outlier';
    severity: 'low' | 'medium' | 'high' | 'critical';
    description?: string;
    confidence: number; // 0-1 scale
  };
  baseline: {
    established: boolean;
    value?: number;
    range?: { min: number; max: number };
    updatedAt?: Date;
  };
}

export interface GoldenSignalMetrics {
  latency: {
    avg: number;
    p95: number;
    p99: number;
    unit: 'ms';
    source: 'span' | 'transaction' | 'metric';
    trend?: 'increasing' | 'decreasing' | 'stable';
    analytics?: AnalyticalMetadata;
  };
  traffic: {
    rate: number;
    unit: 'rpm' | 'rps';
    source: 'span' | 'transaction' | 'metric';
    trend?: 'increasing' | 'decreasing' | 'stable';
    analytics?: AnalyticalMetadata;
  };
  errors: {
    rate: number;
    percentage: number;
    unit: '%';
    source: 'span' | 'transaction' | 'error_event';
    baseline?: number;
    anomaly?: boolean;
    analytics?: AnalyticalMetadata;
  };
  saturation: {
    cpu?: {
      avg: number;
      max: number;
      unit: '%';
      source: 'system' | 'container' | 'runtime';
      analytics?: AnalyticalMetadata;
    };
    memory?: {
      avg: number;
      max: number;
      unit: '%';
      source: 'system' | 'container' | 'runtime';
      analytics?: AnalyticalMetadata;
    };
    available: boolean;
  };
}

export interface TelemetryContext {
  hasOpenTelemetry: boolean;
  hasNewRelicAPM: boolean;
  eventTypes: string[];
  metricStreams: string[];
  serviceIdentifierField: string; // 'service.name' vs 'appName'
  primaryDataSource: 'otel' | 'apm' | 'mixed';
}

export interface EntityGoldenMetrics {
  entity: {
    guid: string;
    name: string;
    type: string;
  };
  metrics: GoldenSignalMetrics;
  context: TelemetryContext;
  dataFreshness: {
    timestamp: Date;
    cached: boolean;
    staleness: 'fresh' | 'recent' | 'stale';
  };
  insights: string[]; // Analytical insights and anomalies
}

export class GoldenSignalsEngine {
  constructor(
    private nerdgraph: NerdGraphClient,
    private logger: Logger
  ) {}

  /**
   * Analyze telemetry context to determine data patterns
   */
  async analyzeTelemetryContext(accountId: number): Promise<TelemetryContext> {
    this.logger.info('Analyzing telemetry context', { accountId });

    try {
      // Check for OpenTelemetry signals
      const otelCheck = await this.checkOpenTelemetryPresence(accountId);
      
      // Check for New Relic APM signals  
      const apmCheck = await this.checkNewRelicAPMPresence(accountId);
      
      // Determine event types available
      const eventTypes = await this.discoverEventTypes(accountId);
      
      // Check for metric streams
      const metricStreams = await this.discoverMetricStreams(accountId);

      // Determine service identifier field
      const serviceIdentifierField = this.determineServiceIdentifier(otelCheck, apmCheck);
      
      // Classify primary data source
      const primaryDataSource = this.classifyDataSource(otelCheck, apmCheck);

      const context: TelemetryContext = {
        hasOpenTelemetry: otelCheck,
        hasNewRelicAPM: apmCheck,
        eventTypes,
        metricStreams,
        serviceIdentifierField,
        primaryDataSource,
      };

      this.logger.info('Telemetry context analyzed', context);
      return context;

    } catch (error: any) {
      this.logger.error('Failed to analyze telemetry context', {
        accountId,
        error: error.message,
      });
      
      // Return safe fallback context
      return {
        hasOpenTelemetry: false,
        hasNewRelicAPM: true,
        eventTypes: ['Transaction'],
        metricStreams: [],
        serviceIdentifierField: 'appName',
        primaryDataSource: 'apm',
      };
    }
  }

  /**
   * Get golden signal metrics for a specific entity
   */
  async getEntityGoldenMetrics(
    entityGuid: string, 
    sinceMinutes: number = 30
  ): Promise<EntityGoldenMetrics> {
    this.logger.info('Fetching golden metrics', { entityGuid, sinceMinutes });

    try {
      // Get entity details
      const entity = await this.getEntityDetails(entityGuid);
      
      // Get account context
      const accountId = await this.extractAccountId(entityGuid);
      const context = await this.analyzeTelemetryContext(accountId);
      
      // Fetch golden signal metrics based on context
      const metrics = await this.fetchGoldenSignalMetrics(
        entityGuid, 
        entity, 
        context, 
        sinceMinutes
      );
      
      // Generate analytical insights
      const insights = await this.generateInsights(metrics, entity, context);

      return {
        entity: {
          guid: entityGuid,
          name: entity.name,
          type: entity.type,
        },
        metrics,
        context,
        dataFreshness: {
          timestamp: new Date(),
          cached: false,
          staleness: 'fresh',
        },
        insights,
      };

    } catch (error: any) {
      this.logger.error('Failed to get golden metrics', {
        entityGuid,
        error: error.message,
      });
      throw error;
    }
  }

  /**
   * Check for OpenTelemetry instrumentation
   */
  private async checkOpenTelemetryPresence(accountId: number): Promise<boolean> {
    try {
      // Check for Span events (primary OTEL indicator)
      const spanQuery = `
        SELECT count(*) 
        FROM Span 
        WHERE instrumentation.provider = 'opentelemetry' 
        SINCE 1 hour ago 
        LIMIT 1
      `;
      
      const spanResult = await this.nerdgraph.nrql(accountId, spanQuery);
      const hasOtelSpans = spanResult.results[0]?.count > 0;

      // Check for OTEL metric naming patterns
      const metricQuery = `
        SELECT uniques(metricName) 
        FROM Metric 
        WHERE metricName LIKE 'http.server.%' 
        OR metricName LIKE 'grpc.server.%'
        OR metricName LIKE 'service.%'
        SINCE 1 hour ago 
        LIMIT 10
      `;
      
      const metricResult = await this.nerdgraph.nrql(accountId, metricQuery);
      const hasOtelMetrics = metricResult.results.length > 0;

      return hasOtelSpans || hasOtelMetrics;

    } catch (error: any) {
      this.logger.warn('Could not check OTEL presence', { error: error.message });
      return false;
    }
  }

  /**
   * Check for New Relic APM instrumentation
   */
  private async checkNewRelicAPMPresence(accountId: number): Promise<boolean> {
    try {
      const query = `
        SELECT count(*) 
        FROM Transaction 
        SINCE 1 hour ago 
        LIMIT 1
      `;
      
      const result = await this.nerdgraph.nrql(accountId, query);
      return result.results[0]?.count > 0;

    } catch (error: any) {
      this.logger.warn('Could not check APM presence', { error: error.message });
      return false;
    }
  }

  /**
   * Discover available event types
   */
  private async discoverEventTypes(accountId: number): Promise<string[]> {
    try {
      const query = 'SHOW EVENT TYPES SINCE 1 week ago';
      const result = await this.nerdgraph.nrql(accountId, query);
      
      return result.results
        .map((row: any) => row.eventType)
        .filter((eventType: string) => eventType && eventType.trim())
        .slice(0, 20); // Limit for performance

    } catch (error: any) {
      this.logger.warn('Could not discover event types', { error: error.message });
      return ['Transaction']; // Safe fallback
    }
  }

  /**
   * Discover metric streams
   */
  private async discoverMetricStreams(accountId: number): Promise<string[]> {
    try {
      const query = `
        SELECT uniques(metricName, 50) 
        FROM Metric 
        SINCE 1 hour ago 
        LIMIT 1
      `;
      
      const result = await this.nerdgraph.nrql(accountId, query);
      return result.results[0]?.members || [];

    } catch (error: any) {
      this.logger.warn('Could not discover metric streams', { error: error.message });
      return [];
    }
  }

  /**
   * Determine service identifier field based on instrumentation
   */
  private determineServiceIdentifier(hasOtel: boolean, hasAPM: boolean): string {
    if (hasOtel && !hasAPM) return 'service.name';
    if (!hasOtel && hasAPM) return 'appName';
    if (hasOtel && hasAPM) return 'service.name'; // Prefer OTEL in mixed
    return 'appName'; // Safe fallback
  }

  /**
   * Classify primary data source
   */
  private classifyDataSource(hasOtel: boolean, hasAPM: boolean): 'otel' | 'apm' | 'mixed' {
    if (hasOtel && hasAPM) return 'mixed';
    if (hasOtel) return 'otel';
    return 'apm';
  }

  /**
   * Get entity details from NerdGraph
   */
  private async getEntityDetails(entityGuid: string): Promise<any> {
    const query = `
      {
        actor {
          entity(guid: "${entityGuid}") {
            guid
            name
            type
            domain
            entityType
            account {
              id
            }
          }
        }
      }
    `;

    const result = await this.nerdgraph.request(query);
    const entity = result.actor?.entity;
    
    if (!entity) {
      throw new Error(`Entity not found: ${entityGuid}`);
    }
    
    return entity;
  }

  /**
   * Extract account ID from entity
   */
  private async extractAccountId(entityGuid: string): Promise<number> {
    const entity = await this.getEntityDetails(entityGuid);
    return entity.account.id;
  }

  /**
   * Fetch golden signal metrics based on telemetry context
   */
  private async fetchGoldenSignalMetrics(
    entityGuid: string,
    entity: any,
    context: TelemetryContext,
    sinceMinutes: number
  ): Promise<GoldenSignalMetrics> {
    const accountId = entity.account.id;
    
    // Fetch metrics in parallel for efficiency
    const [latency, traffic, errors, saturation] = await Promise.all([
      this.fetchLatencyMetrics(accountId, entityGuid, entity.name, context, sinceMinutes),
      this.fetchTrafficMetrics(accountId, entityGuid, entity.name, context, sinceMinutes),
      this.fetchErrorMetrics(accountId, entityGuid, entity.name, context, sinceMinutes),
      this.fetchSaturationMetrics(accountId, entityGuid, entity.name, context, sinceMinutes),
    ]);

    return { latency, traffic, errors, saturation };
  }

  /**
   * Fetch latency metrics (Response Time)
   */
  private async fetchLatencyMetrics(
    accountId: number,
    entityGuid: string,
    entityName: string,
    context: TelemetryContext,
    sinceMinutes: number
  ): Promise<GoldenSignalMetrics['latency']> {
    const timeWindow = `SINCE ${sinceMinutes} minutes ago`;
    
    try {
      if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
        // Use OTEL Spans for latency
        const query = `
          SELECT 
            average(duration.ms) as avg,
            percentile(duration.ms, 95) as p95,
            percentile(duration.ms, 99) as p99
          FROM Span 
          WHERE entity.guid = '${entityGuid}'
          AND span.kind = 'server'
          ${timeWindow}
        `;
        
        const result = await this.nerdgraph.nrql(accountId, query);
        const data = result.results[0] || {};
        
        return {
          avg: Math.round(data.avg || 0),
          p95: Math.round(data.p95 || 0),
          p99: Math.round(data.p99 || 0),
          unit: 'ms',
          source: 'span',
        };
      } else {
        // Use APM Transaction events
        const query = `
          SELECT 
            average(duration) * 1000 as avg,
            percentile(duration, 95) * 1000 as p95,
            percentile(duration, 99) * 1000 as p99
          FROM Transaction 
          WHERE entity.guid = '${entityGuid}'
          ${timeWindow}
        `;
        
        const result = await this.nerdgraph.nrql(accountId, query);
        const data = result.results[0] || {};
        
        return {
          avg: Math.round(data.avg || 0),
          p95: Math.round(data.p95 || 0),
          p99: Math.round(data.p99 || 0),
          unit: 'ms',
          source: 'transaction',
        };
      }
    } catch (error: any) {
      this.logger.warn('Failed to fetch latency metrics', { entityGuid, error: error.message });
      return {
        avg: 0,
        p95: 0,
        p99: 0,
        unit: 'ms',
        source: 'transaction',
      };
    }
  }

  /**
   * Fetch traffic metrics (Throughput)
   */
  private async fetchTrafficMetrics(
    accountId: number,
    entityGuid: string,
    entityName: string,
    context: TelemetryContext,
    sinceMinutes: number
  ): Promise<GoldenSignalMetrics['traffic']> {
    const timeWindow = `SINCE ${sinceMinutes} minutes ago`;
    
    try {
      let query: string;
      let source: 'span' | 'transaction' | 'metric';
      
      if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
        // Use OTEL Spans for throughput
        query = `
          SELECT rate(count(*), 1 minute) as rate
          FROM Span 
          WHERE entity.guid = '${entityGuid}'
          AND span.kind = 'server'
          ${timeWindow}
        `;
        source = 'span';
      } else {
        // Use APM Transactions
        query = `
          SELECT rate(count(*), 1 minute) as rate
          FROM Transaction 
          WHERE entity.guid = '${entityGuid}'
          ${timeWindow}
        `;
        source = 'transaction';
      }
      
      const result = await this.nerdgraph.nrql(accountId, query);
      const rate = result.results[0]?.rate || 0;
      
      return {
        rate: Math.round(rate * 100) / 100,
        unit: 'rpm',
        source,
      };
    } catch (error: any) {
      this.logger.warn('Failed to fetch traffic metrics', { entityGuid, error: error.message });
      return {
        rate: 0,
        unit: 'rpm',
        source: 'transaction',
      };
    }
  }

  /**
   * Fetch error metrics (Error Rate)
   */
  private async fetchErrorMetrics(
    accountId: number,
    entityGuid: string,
    entityName: string,
    context: TelemetryContext,
    sinceMinutes: number
  ): Promise<GoldenSignalMetrics['errors']> {
    const timeWindow = `SINCE ${sinceMinutes} minutes ago`;
    
    try {
      if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
        // Use OTEL Spans for error rate
        const query = `
          SELECT 
            filter(count(*), WHERE otel.status_code = 'ERROR') as errors,
            count(*) as total
          FROM Span 
          WHERE entity.guid = '${entityGuid}'
          AND span.kind = 'server'
          ${timeWindow}
        `;
        
        const result = await this.nerdgraph.nrql(accountId, query);
        const data = result.results[0] || {};
        const errors = data.errors || 0;
        const total = data.total || 1;
        const percentage = (errors / total) * 100;
        
        return {
          rate: errors,
          percentage: Math.round(percentage * 100) / 100,
          unit: '%',
          source: 'span',
        };
      } else {
        // Use APM Transaction events
        const query = `
          SELECT 
            filter(count(*), WHERE error IS true) as errors,
            count(*) as total
          FROM Transaction 
          WHERE entity.guid = '${entityGuid}'
          ${timeWindow}
        `;
        
        const result = await this.nerdgraph.nrql(accountId, query);
        const data = result.results[0] || {};
        const errors = data.errors || 0;
        const total = data.total || 1;
        const percentage = (errors / total) * 100;
        
        return {
          rate: errors,
          percentage: Math.round(percentage * 100) / 100,
          unit: '%',
          source: 'transaction',
        };
      }
    } catch (error: any) {
      this.logger.warn('Failed to fetch error metrics', { entityGuid, error: error.message });
      return {
        rate: 0,
        percentage: 0,
        unit: '%',
        source: 'transaction',
      };
    }
  }

  /**
   * Fetch saturation metrics (Resource Usage)
   */
  private async fetchSaturationMetrics(
    accountId: number,
    entityGuid: string,
    entityName: string,
    context: TelemetryContext,
    sinceMinutes: number
  ): Promise<GoldenSignalMetrics['saturation']> {
    const timeWindow = `SINCE ${sinceMinutes} minutes ago`;
    
    try {
      // Try to get system metrics associated with the entity
      const cpuQuery = `
        SELECT 
          average(cpuPercent) as avg,
          max(cpuPercent) as max
        FROM SystemSample 
        WHERE entityGuid = '${entityGuid}'
        ${timeWindow}
      `;
      
      const memoryQuery = `
        SELECT 
          average(memoryUsedPercent) as avg,
          max(memoryUsedPercent) as max
        FROM SystemSample 
        WHERE entityGuid = '${entityGuid}'
        ${timeWindow}
      `;
      
      const [cpuResult, memoryResult] = await Promise.allSettled([
        this.nerdgraph.nrql(accountId, cpuQuery),
        this.nerdgraph.nrql(accountId, memoryQuery),
      ]);
      
      const cpu = cpuResult.status === 'fulfilled' && cpuResult.value.results[0] ? {
        avg: Math.round(cpuResult.value.results[0].avg || 0),
        max: Math.round(cpuResult.value.results[0].max || 0),
        unit: '%' as const,
        source: 'system' as const,
      } : undefined;
      
      const memory = memoryResult.status === 'fulfilled' && memoryResult.value.results[0] ? {
        avg: Math.round(memoryResult.value.results[0].avg || 0),
        max: Math.round(memoryResult.value.results[0].max || 0),
        unit: '%' as const,
        source: 'system' as const,
      } : undefined;
      
      return {
        cpu,
        memory,
        available: !!(cpu || memory),
      };
    } catch (error: any) {
      this.logger.warn('Failed to fetch saturation metrics', { entityGuid, error: error.message });
      return {
        available: false,
      };
    }
  }

  /**
   * Generate analytical insights from metrics
   */
  private async generateInsights(
    metrics: GoldenSignalMetrics,
    entity: any,
    context: TelemetryContext
  ): Promise<string[]> {
    const insights: string[] = [];

    // Latency insights
    if (metrics.latency.p95 > 2000) {
      insights.push(`⚠️ High latency detected: P95 response time is ${metrics.latency.p95}ms (above 2s threshold)`);
    } else if (metrics.latency.p95 > 1000) {
      insights.push(`⚠️ Elevated latency: P95 response time is ${metrics.latency.p95}ms (above 1s)`);
    }

    // Error insights
    if (metrics.errors.percentage > 5) {
      insights.push(`🚨 High error rate: ${metrics.errors.percentage}% errors (above 5% threshold)`);
    } else if (metrics.errors.percentage > 1) {
      insights.push(`⚠️ Elevated error rate: ${metrics.errors.percentage}% errors (above 1%)`);
    } else if (metrics.errors.percentage === 0) {
      insights.push(`✅ No errors detected in the monitoring period`);
    }

    // Traffic insights
    if (metrics.traffic.rate === 0) {
      insights.push(`⚠️ No traffic detected - service may be down or not receiving requests`);
    } else if (metrics.traffic.rate < 1) {
      insights.push(`ℹ️ Low traffic volume: ${metrics.traffic.rate} requests per minute`);
    }

    // Saturation insights
    if (metrics.saturation.cpu && metrics.saturation.cpu.avg > 80) {
      insights.push(`⚠️ High CPU usage: Average ${metrics.saturation.cpu.avg}% (above 80%)`);
    }
    if (metrics.saturation.memory && metrics.saturation.memory.avg > 80) {
      insights.push(`⚠️ High memory usage: Average ${metrics.saturation.memory.avg}% (above 80%)`);
    }
    if (!metrics.saturation.available) {
      insights.push(`ℹ️ No resource saturation data available - consider enabling infrastructure monitoring`);
    }

    // Context insights
    if (context.primaryDataSource === 'mixed') {
      insights.push(`ℹ️ Mixed telemetry detected: Both OpenTelemetry and New Relic APM data available`);
    } else if (context.primaryDataSource === 'otel') {
      insights.push(`ℹ️ OpenTelemetry instrumentation detected - using span-based metrics`);
    }

    return insights;
  }

  /**
   * Generate analytical metadata for a metric stream
   */
  async generateAnalyticalMetadata(
    accountId: number,
    query: string,
    metricName: string,
    timeWindowHours: number = 24
  ): Promise<AnalyticalMetadata> {
    try {
      // Get historical data for analytics
      const historicalQuery = query.replace(/SINCE \d+/, `SINCE ${timeWindowHours} hours ago`) + ' TIMESERIES';
      const result = await this.nerdgraph.nrql(accountId, historicalQuery);
      
      const timeSeries = result.results || [];
      if (timeSeries.length === 0) {
        return this.getEmptyAnalyticalMetadata();
      }

      // Analyze data quality
      const dataQuality = this.analyzeDataQuality(timeSeries);
      
      // Detect seasonality patterns
      const seasonality = this.detectSeasonality(timeSeries, metricName);
      
      // Detect anomalies
      const anomalies = this.detectAnomalies(timeSeries, metricName);
      
      // Establish baseline
      const baseline = this.establishBaseline(timeSeries, metricName);

      return {
        dataQuality,
        seasonality,
        anomalies,
        baseline,
      };

    } catch (error: any) {
      this.logger.warn('Failed to generate analytical metadata', { 
        metricName, 
        error: error.message 
      });
      return this.getEmptyAnalyticalMetadata();
    }
  }

  /**
   * Analyze data quality metrics
   */
  private analyzeDataQuality(timeSeries: any[]): AnalyticalMetadata['dataQuality'] {
    const totalPoints = timeSeries.length;
    const validPoints = timeSeries.filter(point => 
      point && typeof point === 'object' && Object.keys(point).length > 1
    ).length;
    
    const completeness = totalPoints > 0 ? validPoints / totalPoints : 0;
    
    // Analyze consistency (coefficient of variation)
    const values = timeSeries.map(point => {
      const keys = Object.keys(point).filter(k => k !== 'timestamp' && k !== 'endTimeSeconds');
      return keys.length > 0 ? point[keys[0]] : 0;
    }).filter(v => typeof v === 'number' && !isNaN(v));
    
    let consistency = 1;
    if (values.length > 1) {
      const mean = values.reduce((sum, v) => sum + v, 0) / values.length;
      const variance = values.reduce((sum, v) => sum + Math.pow(v - mean, 2), 0) / values.length;
      const stdDev = Math.sqrt(variance);
      const coefficientOfVariation = mean > 0 ? stdDev / mean : 0;
      consistency = Math.max(0, 1 - Math.min(1, coefficientOfVariation));
    }

    // Determine volume stability
    let volumeStability: 'stable' | 'increasing' | 'decreasing' | 'volatile' = 'stable';
    if (values.length > 3) {
      const firstHalf = values.slice(0, Math.floor(values.length / 2));
      const secondHalf = values.slice(Math.floor(values.length / 2));
      
      const firstAvg = firstHalf.reduce((sum, v) => sum + v, 0) / firstHalf.length;
      const secondAvg = secondHalf.reduce((sum, v) => sum + v, 0) / secondHalf.length;
      
      const change = (secondAvg - firstAvg) / firstAvg;
      
      if (Math.abs(change) > 0.5) {
        volumeStability = 'volatile';
      } else if (change > 0.2) {
        volumeStability = 'increasing';
      } else if (change < -0.2) {
        volumeStability = 'decreasing';
      }
    }

    return {
      completeness,
      consistency,
      volumeStability,
      lastDataPoint: new Date(Math.max(...timeSeries.map(p => p.timestamp || p.endTimeSeconds || 0)) * 1000),
    };
  }

  /**
   * Detect seasonality patterns in time series data
   */
  private detectSeasonality(timeSeries: any[], metricName: string): AnalyticalMetadata['seasonality'] {
    if (timeSeries.length < 24) { // Need at least 24 points for daily pattern
      return { detected: false, confidence: 0 };
    }

    const values = this.extractValues(timeSeries);
    
    // Simple autocorrelation for daily pattern (24 data points = 24 hours if hourly)
    const lag24 = this.calculateAutocorrelation(values, 24);
    const lag168 = values.length >= 168 ? this.calculateAutocorrelation(values, 168) : 0; // Weekly

    if (lag24 > 0.7) {
      return {
        detected: true,
        pattern: 'daily',
        confidence: lag24,
      };
    } else if (lag168 > 0.6) {
      return {
        detected: true,
        pattern: 'weekly',
        confidence: lag168,
      };
    }

    return { detected: false, confidence: 0 };
  }

  /**
   * Detect anomalies in metric data
   */
  private detectAnomalies(timeSeries: any[], metricName: string): AnalyticalMetadata['anomalies'] {
    const values = this.extractValues(timeSeries);
    
    if (values.length < 10) {
      return { detected: false, severity: 'low', confidence: 0 };
    }

    const mean = values.reduce((sum, v) => sum + v, 0) / values.length;
    const stdDev = Math.sqrt(
      values.reduce((sum, v) => sum + Math.pow(v - mean, 2), 0) / values.length
    );

    // Z-score based anomaly detection
    const threshold = 2.5; // Standard deviations
    const anomalousValues = values.filter(v => Math.abs(v - mean) > threshold * stdDev);
    
    if (anomalousValues.length === 0) {
      return { detected: false, severity: 'low', confidence: 0 };
    }

    // Determine anomaly type and severity
    const maxValue = Math.max(...values);
    const minValue = Math.min(...values);
    const latestValue = values[values.length - 1];
    
    let type: 'spike' | 'drop' | 'trend_change' | 'outlier' = 'outlier';
    let severity: 'low' | 'medium' | 'high' | 'critical' = 'medium';
    
    // Detect spikes and drops
    if (latestValue > mean + 3 * stdDev) {
      type = 'spike';
      severity = latestValue > mean + 4 * stdDev ? 'critical' : 'high';
    } else if (latestValue < mean - 3 * stdDev) {
      type = 'drop';
      severity = latestValue < mean - 4 * stdDev ? 'critical' : 'high';
    }

    // For error rates, any significant increase is critical
    if (metricName.toLowerCase().includes('error') && latestValue > mean * 2) {
      severity = 'critical';
    }

    const confidence = Math.min(1, anomalousValues.length / values.length * 2);

    return {
      detected: true,
      type,
      severity,
      confidence,
      description: this.generateAnomalyDescription(type, severity, metricName),
    };
  }

  /**
   * Establish performance baseline
   */
  private establishBaseline(timeSeries: any[], metricName: string): AnalyticalMetadata['baseline'] {
    const values = this.extractValues(timeSeries);
    
    if (values.length < 5) {
      return { established: false };
    }

    // Use median as baseline for robustness
    const sortedValues = [...values].sort((a, b) => a - b);
    const median = sortedValues[Math.floor(sortedValues.length / 2)];
    
    // Calculate range (interquartile range)
    const q1 = sortedValues[Math.floor(sortedValues.length * 0.25)];
    const q3 = sortedValues[Math.floor(sortedValues.length * 0.75)];
    
    return {
      established: true,
      value: median,
      range: { min: q1, max: q3 },
      updatedAt: new Date(),
    };
  }

  /**
   * Helper methods for analytics
   */
  private extractValues(timeSeries: any[]): number[] {
    return timeSeries.map(point => {
      const keys = Object.keys(point).filter(k => 
        k !== 'timestamp' && 
        k !== 'endTimeSeconds' && 
        typeof point[k] === 'number'
      );
      return keys.length > 0 ? point[keys[0]] : 0;
    }).filter(v => typeof v === 'number' && !isNaN(v));
  }

  private calculateAutocorrelation(values: number[], lag: number): number {
    if (values.length <= lag) return 0;
    
    const n = values.length - lag;
    const mean = values.reduce((sum, v) => sum + v, 0) / values.length;
    
    let numerator = 0;
    let denominator = 0;
    
    for (let i = 0; i < n; i++) {
      numerator += (values[i] - mean) * (values[i + lag] - mean);
    }
    
    for (let i = 0; i < values.length; i++) {
      denominator += Math.pow(values[i] - mean, 2);
    }
    
    return denominator > 0 ? numerator / denominator : 0;
  }

  private generateAnomalyDescription(
    type: string, 
    severity: string, 
    metricName: string
  ): string {
    const descriptions = {
      spike: `Unusual spike detected in ${metricName}`,
      drop: `Significant drop detected in ${metricName}`,
      trend_change: `Trend change detected in ${metricName}`,
      outlier: `Outlier values detected in ${metricName}`,
    };
    
    return descriptions[type as keyof typeof descriptions] || `Anomaly detected in ${metricName}`;
  }

  private getEmptyAnalyticalMetadata(): AnalyticalMetadata {
    return {
      dataQuality: {
        completeness: 0,
        consistency: 0,
        volumeStability: 'stable',
        lastDataPoint: new Date(),
      },
      seasonality: {
        detected: false,
        confidence: 0,
      },
      anomalies: {
        detected: false,
        severity: 'low',
        confidence: 0,
      },
      baseline: {
        established: false,
      },
    };
  }

  /**
   * Enhanced golden metrics with analytical metadata
   */
  async getEntityGoldenMetricsWithAnalytics(
    entityGuid: string,
    sinceMinutes: number = 30,
    includeAnalytics: boolean = true
  ): Promise<EntityGoldenMetrics> {
    // Get base golden metrics
    const baseMetrics = await this.getEntityGoldenMetrics(entityGuid, sinceMinutes);
    
    if (!includeAnalytics) {
      return baseMetrics;
    }

    try {
      const accountId = await this.extractAccountId(entityGuid);
      const context = baseMetrics.context;
      
      // Generate analytics for each golden signal
      const [latencyAnalytics, trafficAnalytics, errorAnalytics] = await Promise.allSettled([
        this.generateAnalyticalMetadata(
          accountId,
          this.buildLatencyAnalyticsQuery(context, entityGuid),
          'latency',
          Math.max(sinceMinutes / 60, 2) // Convert to hours, minimum 2 hours
        ),
        this.generateAnalyticalMetadata(
          accountId,
          this.buildTrafficAnalyticsQuery(context, entityGuid),
          'traffic',
          Math.max(sinceMinutes / 60, 2)
        ),
        this.generateAnalyticalMetadata(
          accountId,
          this.buildErrorAnalyticsQuery(context, entityGuid),
          'error_rate',
          Math.max(sinceMinutes / 60, 2)
        ),
      ]);

      // Add analytics to metrics
      if (latencyAnalytics.status === 'fulfilled') {
        baseMetrics.metrics.latency.analytics = latencyAnalytics.value;
      }
      if (trafficAnalytics.status === 'fulfilled') {
        baseMetrics.metrics.traffic.analytics = trafficAnalytics.value;
      }
      if (errorAnalytics.status === 'fulfilled') {
        baseMetrics.metrics.errors.analytics = errorAnalytics.value;
      }

      // Enhance insights with analytical findings
      const analyticalInsights = this.generateAnalyticalInsights(baseMetrics.metrics);
      baseMetrics.insights.push(...analyticalInsights);

      return baseMetrics;

    } catch (error: any) {
      this.logger.warn('Failed to generate analytics', { entityGuid, error: error.message });
      return baseMetrics; // Return base metrics without analytics
    }
  }

  /**
   * Build analytics queries for different golden signals
   */
  private buildLatencyAnalyticsQuery(context: TelemetryContext, entityGuid: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT percentile(duration.ms, 95) FROM Span WHERE entity.guid = '${entityGuid}' AND span.kind = 'server'`;
    } else {
      return `SELECT percentile(duration, 95) * 1000 FROM Transaction WHERE entity.guid = '${entityGuid}'`;
    }
  }

  private buildTrafficAnalyticsQuery(context: TelemetryContext, entityGuid: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT rate(count(*), 1 minute) FROM Span WHERE entity.guid = '${entityGuid}' AND span.kind = 'server'`;
    } else {
      return `SELECT rate(count(*), 1 minute) FROM Transaction WHERE entity.guid = '${entityGuid}'`;
    }
  }

  private buildErrorAnalyticsQuery(context: TelemetryContext, entityGuid: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT filter(count(*), WHERE otel.status_code = 'ERROR') / count(*) * 100 FROM Span WHERE entity.guid = '${entityGuid}' AND span.kind = 'server'`;
    } else {
      return `SELECT filter(count(*), WHERE error IS true) / count(*) * 100 FROM Transaction WHERE entity.guid = '${entityGuid}'`;
    }
  }

  /**
   * Generate insights from analytical metadata
   */
  private generateAnalyticalInsights(metrics: GoldenSignalMetrics): string[] {
    const insights: string[] = [];

    // Latency analytics insights
    if (metrics.latency.analytics?.anomalies.detected) {
      const anomaly = metrics.latency.analytics.anomalies;
      insights.push(`🔍 ${anomaly.description} (${anomaly.severity} severity, ${Math.round(anomaly.confidence * 100)}% confidence)`);
    }

    if (metrics.latency.analytics?.seasonality.detected) {
      const seasonality = metrics.latency.analytics.seasonality;
      insights.push(`📈 ${seasonality.pattern} seasonality pattern detected in latency (${Math.round(seasonality.confidence * 100)}% confidence)`);
    }

    // Traffic analytics insights
    if (metrics.traffic.analytics?.anomalies.detected) {
      const anomaly = metrics.traffic.analytics.anomalies;
      insights.push(`📊 ${anomaly.description} (${anomaly.severity} severity)`);
    }

    // Error analytics insights
    if (metrics.errors.analytics?.anomalies.detected) {
      const anomaly = metrics.errors.analytics.anomalies;
      insights.push(`🚨 ${anomaly.description} (${anomaly.severity} severity)`);
    }

    // Data quality insights
    if (metrics.latency.analytics?.dataQuality.completeness < 0.8) {
      insights.push(`⚠️ Low data completeness detected (${Math.round(metrics.latency.analytics.dataQuality.completeness * 100)}%) - some monitoring gaps may exist`);
    }

    if (metrics.traffic.analytics?.dataQuality.volumeStability === 'volatile') {
      insights.push(`📈 Volatile traffic patterns detected - consider investigating load balancing or scaling issues`);
    }

    return insights;
  }
}
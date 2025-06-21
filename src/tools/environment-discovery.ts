/**
 * Environment Discovery Tool - discover.environment
 * 
 * Provides comprehensive environment discovery in a single call,
 * giving LLM agents complete situational awareness of the New Relic setup.
 */

import { Tool, CallToolRequestSchema } from '@modelcontextprotocol/sdk/types.js';
import { NerdGraphClient, Logger } from '../core/types.js';
import { GoldenSignalsEngine, TelemetryContext } from '../core/golden-signals.js';

export interface EnvironmentSnapshot {
  entities: EntitySummary[];
  eventTypes: EventTypeSummary[];
  metricStreams: MetricStreamSummary[];
  schemaHints: SchemaGuidance;
  telemetryContext: TelemetryContext;
  observabilityGaps: string[];
  recommendations: string[];
}

export interface EntitySummary {
  name: string;
  guid: string;
  type: string;
  domain: string;
  language?: string;
  environment?: string;
  healthStatus?: 'healthy' | 'warning' | 'critical' | 'unknown';
  goldenSignalsAvailable: boolean;
}

export interface EventTypeSummary {
  name: string;
  sampleCount: number;
  lastSeen: string;
  description: string;
  keyAttributes: string[];
}

export interface MetricStreamSummary {
  category: string;
  examples: string[];
  count: number;
  type: 'otel' | 'custom' | 'infrastructure' | 'apm';
}

export interface SchemaGuidance {
  serviceIdentifierField: string;
  preferredQueryPatterns: {
    latency: string;
    throughput: string;
    errors: string;
  };
  goldenSignalStrategy: string;
  instrumentationNotes: string[];
}

export class EnvironmentDiscoveryTool {
  private cache = new Map<string, { data: EnvironmentSnapshot; timestamp: Date }>();
  private readonly CACHE_TTL_MS = 5 * 60 * 1000; // 5 minutes

  constructor(
    private nerdgraph: NerdGraphClient,
    private logger: Logger,
    private goldenSignals: GoldenSignalsEngine
  ) {}

  /**
   * Get tool definition for MCP registration
   */
  getToolDefinition(): Tool {
    return {
      name: 'discover.environment',
      description: `Discover the complete New Relic environment in one comprehensive call.
      
      🎯 **Purpose**: Provides LLM agents with complete situational awareness of the observability setup.
      
      **Returns**:
      - Complete inventory of monitored entities (services, hosts, etc.)
      - Available telemetry event types and their characteristics  
      - Metric streams and data sources (OpenTelemetry vs APM)
      - Schema guidance for optimal NRQL queries
      - Observability gaps and recommendations
      
      **Use this first** to understand what data is available before using other tools.
      
      **OpenTelemetry Aware**: Automatically detects OTEL vs traditional APM instrumentation.`,
      inputSchema: {
        type: 'object',
        properties: {
          includeHealth: {
            type: 'boolean',
            description: 'Include basic health status for entities (may take longer)',
            default: false,
          },
          maxEntities: {
            type: 'number',
            description: 'Maximum number of entities to include (for large environments)',
            default: 50,
            minimum: 10,
            maximum: 200,
          },
          forceRefresh: {
            type: 'boolean',
            description: 'Force fresh discovery, bypassing cache',
            default: false,
          },
        },
        additionalProperties: false,
      },
    };
  }

  /**
   * Handle the discover.environment tool call
   */
  async handle(params: any): Promise<any> {
    const {
      includeHealth = false,
      maxEntities = 50,
      forceRefresh = false,
    } = params;

    this.logger.info('Environment discovery requested', {
      includeHealth,
      maxEntities,
      forceRefresh,
    });

    try {
      // Check cache first
      const cacheKey = `env-${includeHealth}-${maxEntities}`;
      const cached = this.cache.get(cacheKey);
      
      if (!forceRefresh && cached && this.isCacheValid(cached.timestamp)) {
        this.logger.info('Returning cached environment discovery');
        return this.formatResponse(cached.data, true);
      }

      // Perform fresh discovery
      const snapshot = await this.discoverEnvironment(includeHealth, maxEntities);
      
      // Cache the result
      this.cache.set(cacheKey, { data: snapshot, timestamp: new Date() });
      
      return this.formatResponse(snapshot, false);

    } catch (error: any) {
      this.logger.error('Environment discovery failed', { error: error.message });
      
      return {
        content: [
          {
            type: 'text',
            text: `❌ **Environment Discovery Failed**\n\nError: ${error.message}\n\nThis might indicate:\n- Invalid API credentials\n- Network connectivity issues\n- Account access restrictions\n\nPlease check your New Relic configuration and try again.`,
          },
        ],
        isError: true,
      };
    }
  }

  /**
   * Perform comprehensive environment discovery
   */
  private async discoverEnvironment(
    includeHealth: boolean,
    maxEntities: number
  ): Promise<EnvironmentSnapshot> {
    this.logger.info('Starting comprehensive environment discovery');

    // Discover in parallel for efficiency
    const [
      telemetryContext,
      entities,
      eventTypes,
      metricStreams,
    ] = await Promise.all([
      this.discoverTelemetryContext(),
      this.discoverEntities(maxEntities, includeHealth),
      this.discoverEventTypes(),
      this.discoverMetricStreams(),
    ]);

    // Generate schema guidance
    const schemaHints = this.generateSchemaGuidance(telemetryContext, eventTypes);
    
    // Identify observability gaps
    const observabilityGaps = this.identifyObservabilityGaps(
      entities,
      eventTypes,
      metricStreams,
      telemetryContext
    );
    
    // Generate recommendations
    const recommendations = this.generateRecommendations(
      telemetryContext,
      observabilityGaps,
      entities
    );

    return {
      entities,
      eventTypes,
      metricStreams,
      schemaHints,
      telemetryContext,
      observabilityGaps,
      recommendations,
    };
  }

  /**
   * Discover telemetry context using golden signals engine
   */
  private async discoverTelemetryContext(): Promise<TelemetryContext> {
    // Get account ID from the configured environment
    const accountId = parseInt(process.env['NEW_RELIC_ACCOUNT_ID'] || '0');
    if (!accountId) {
      throw new Error('No account ID configured');
    }

    return await this.goldenSignals.analyzeTelemetryContext(accountId);
  }

  /**
   * Discover monitored entities
   */
  private async discoverEntities(
    maxEntities: number,
    includeHealth: boolean
  ): Promise<EntitySummary[]> {
    this.logger.info('Discovering entities', { maxEntities, includeHealth });

    const query = `
      {
        actor {
          entitySearch(
            query: "domain IN ('APM', 'INFRA', 'BROWSER', 'MOBILE') AND reporting = true"
            sortBy: LAST_REPORTED_AT
          ) {
            results {
              entities {
                guid
                name
                type
                domain
                entityType
                tags {
                  key
                  values
                }
                goldenMetrics {
                  metrics {
                    name
                    query
                  }
                }
              }
            }
          }
        }
      }
    `;

    const result = await this.nerdgraph.request(query);
    const entities = result.actor?.entitySearch?.results?.entities || [];

    const summaries: EntitySummary[] = [];
    
    for (const entity of entities.slice(0, maxEntities)) {
      const summary: EntitySummary = {
        name: entity.name,
        guid: entity.guid,
        type: entity.type,
        domain: entity.domain,
        goldenSignalsAvailable: !!(entity.goldenMetrics?.metrics?.length),
      };

      // Extract language and environment from tags
      const tags = entity.tags || [];
      for (const tag of tags) {
        if (tag.key === 'language' || tag.key === 'runtime') {
          summary.language = tag.values[0];
        }
        if (tag.key === 'environment' || tag.key === 'env') {
          summary.environment = tag.values[0];
        }
      }

      // Add basic health check if requested
      if (includeHealth && entity.domain === 'APM') {
        try {
          summary.healthStatus = await this.checkEntityHealth(entity.guid);
        } catch (error: any) {
          summary.healthStatus = 'unknown';
        }
      }

      summaries.push(summary);
    }

    return summaries;
  }

  /**
   * Discover available event types
   */
  private async discoverEventTypes(): Promise<EventTypeSummary[]> {
    this.logger.info('Discovering event types');

    const accountId = parseInt(process.env['NEW_RELIC_ACCOUNT_ID'] || '0');
    
    try {
      const showTypesQuery = 'SHOW EVENT TYPES SINCE 1 week ago';
      const result = await this.nerdgraph.nrql(accountId, showTypesQuery);
      
      const eventTypes: EventTypeSummary[] = [];
      
      for (const row of result.results.slice(0, 20)) {
        const eventType = row.eventType;
        if (!eventType) continue;

        try {
          // Get sample data for this event type
          const sampleQuery = `
            SELECT count(*) as sampleCount, 
                   latest(timestamp) as lastSeen,
                   keyset() as keySet
            FROM ${eventType} 
            SINCE 24 hours ago 
            LIMIT 1
          `;
          
          const sampleResult = await this.nerdgraph.nrql(accountId, sampleQuery);
          const sample = sampleResult.results[0] || {};
          
          eventTypes.push({
            name: eventType,
            sampleCount: sample.sampleCount || 0,
            lastSeen: sample.lastSeen ? new Date(sample.lastSeen).toISOString() : 'unknown',
            description: this.getEventTypeDescription(eventType),
            keyAttributes: this.extractKeyAttributes(sample.keySet || []),
          });
        } catch (error: any) {
          // Skip if we can't query this event type
          this.logger.warn(`Could not sample event type ${eventType}`, { error: error.message });
        }
      }

      return eventTypes;
    } catch (error: any) {
      this.logger.warn('Could not discover event types', { error: error.message });
      return [];
    }
  }

  /**
   * Discover metric streams
   */
  private async discoverMetricStreams(): Promise<MetricStreamSummary[]> {
    this.logger.info('Discovering metric streams');

    const accountId = parseInt(process.env['NEW_RELIC_ACCOUNT_ID'] || '0');
    
    try {
      const metricsQuery = `
        SELECT uniques(metricName, 100) as metrics 
        FROM Metric 
        SINCE 1 hour ago 
        LIMIT 1
      `;
      
      const result = await this.nerdgraph.nrql(accountId, metricsQuery);
      const metrics = result.results[0]?.metrics || [];
      
      // Categorize metrics
      const categories = this.categorizeMetrics(metrics);
      
      return Object.entries(categories).map(([category, metricList]) => ({
        category,
        examples: metricList.slice(0, 5),
        count: metricList.length,
        type: this.inferMetricType(category, metricList),
      }));
    } catch (error: any) {
      this.logger.warn('Could not discover metric streams', { error: error.message });
      return [];
    }
  }

  /**
   * Generate schema guidance for queries
   */
  private generateSchemaGuidance(
    context: TelemetryContext,
    eventTypes: EventTypeSummary[]
  ): SchemaGuidance {
    const hasTransactions = eventTypes.some(et => et.name === 'Transaction');
    const hasSpans = eventTypes.some(et => et.name === 'Span');
    
    let goldenSignalStrategy = '';
    if (context.hasOpenTelemetry && hasSpans) {
      goldenSignalStrategy = 'Use Span events for golden signals (span.kind = "server" for entry spans)';
    } else if (hasTransactions) {
      goldenSignalStrategy = 'Use Transaction events for golden signals';
    } else {
      goldenSignalStrategy = 'Limited telemetry available - consider improving instrumentation';
    }

    const instrumentationNotes: string[] = [];
    if (context.hasOpenTelemetry) {
      instrumentationNotes.push('OpenTelemetry detected - prefer service.name attribute');
      instrumentationNotes.push('Span events available for distributed tracing');
    }
    if (context.hasNewRelicAPM) {
      instrumentationNotes.push('New Relic APM agent detected - appName attribute available');
    }
    if (context.primaryDataSource === 'mixed') {
      instrumentationNotes.push('Mixed instrumentation - both OTEL and APM data present');
    }

    return {
      serviceIdentifierField: context.serviceIdentifierField,
      preferredQueryPatterns: {
        latency: hasSpans 
          ? 'percentile(duration.ms, 95) FROM Span WHERE span.kind = "server"'
          : 'percentile(duration, 95) * 1000 FROM Transaction',
        throughput: hasSpans
          ? 'rate(count(*), 1 minute) FROM Span WHERE span.kind = "server"'
          : 'rate(count(*), 1 minute) FROM Transaction',
        errors: hasSpans
          ? 'filter(count(*), WHERE otel.status_code = "ERROR") FROM Span'
          : 'filter(count(*), WHERE error IS true) FROM Transaction',
      },
      goldenSignalStrategy,
      instrumentationNotes,
    };
  }

  /**
   * Identify observability gaps
   */
  private identifyObservabilityGaps(
    entities: EntitySummary[],
    eventTypes: EventTypeSummary[],
    metricStreams: MetricStreamSummary[],
    context: TelemetryContext
  ): string[] {
    const gaps: string[] = [];

    // Check for basic telemetry
    if (eventTypes.length === 0) {
      gaps.push('No telemetry events detected - instrumentation may be missing');
    }

    // Check for error tracking
    const hasErrorEvents = eventTypes.some(et => 
      et.name.includes('Error') || et.keyAttributes.includes('error')
    );
    if (!hasErrorEvents) {
      gaps.push('No error tracking events detected - error monitoring may be incomplete');
    }

    // Check for infrastructure monitoring
    const hasInfraEvents = eventTypes.some(et => 
      et.name.includes('System') || et.name.includes('Process')
    );
    if (!hasInfraEvents && entities.some(e => e.domain === 'APM')) {
      gaps.push('No infrastructure monitoring detected - consider enabling host monitoring');
    }

    // Check for log data
    const hasLogs = eventTypes.some(et => et.name === 'Log');
    if (!hasLogs) {
      gaps.push('No log data detected - consider enabling log forwarding');
    }

    // Check for golden signals availability
    const entitiesWithGoldenSignals = entities.filter(e => e.goldenSignalsAvailable);
    if (entitiesWithGoldenSignals.length < entities.length * 0.5) {
      gaps.push('Limited golden signals coverage - some entities may need better instrumentation');
    }

    return gaps;
  }

  /**
   * Generate actionable recommendations
   */
  private generateRecommendations(
    context: TelemetryContext,
    gaps: string[],
    entities: EntitySummary[]
  ): string[] {
    const recommendations: string[] = [];

    if (context.primaryDataSource === 'apm' && !context.hasOpenTelemetry) {
      recommendations.push('Consider migrating to OpenTelemetry for standardized observability');
    }

    if (gaps.some(gap => gap.includes('infrastructure'))) {
      recommendations.push('Install New Relic Infrastructure agent for host and process monitoring');
    }

    if (gaps.some(gap => gap.includes('error'))) {
      recommendations.push('Enable error tracking in your application instrumentation');
    }

    if (gaps.some(gap => gap.includes('log'))) {
      recommendations.push('Configure log forwarding to correlate logs with metrics and traces');
    }

    if (entities.length > 10 && !entities.some(e => e.environment)) {
      recommendations.push('Add environment tags to entities for better organization');
    }

    return recommendations;
  }

  /**
   * Format the response for the LLM
   */
  private formatResponse(snapshot: EnvironmentSnapshot, cached: boolean): any {
    const freshness = cached ? '(📝 Cached data from last 5 minutes)' : '(🔄 Fresh data)';
    
    let content = `# 🔍 New Relic Environment Discovery ${freshness}\n\n`;
    
    // Executive Summary
    content += `## 📊 Executive Summary\n\n`;
    content += `- **Entities Monitored**: ${snapshot.entities.length} (${this.summarizeEntityTypes(snapshot.entities)})\n`;
    content += `- **Telemetry Sources**: ${snapshot.telemetryContext.primaryDataSource.toUpperCase()}`;
    if (snapshot.telemetryContext.primaryDataSource === 'mixed') {
      content += ' (OpenTelemetry + New Relic APM)';
    }
    content += `\n`;
    content += `- **Event Types**: ${snapshot.eventTypes.length} available\n`;
    content += `- **Metric Streams**: ${snapshot.metricStreams.reduce((sum, ms) => sum + ms.count, 0)} metrics across ${snapshot.metricStreams.length} categories\n\n`;

    // Schema Guidance (Critical for LLM)
    content += `## 🎯 Schema Guidance for Queries\n\n`;
    content += `**Service Identifier**: Use \`${snapshot.schemaHints.serviceIdentifierField}\` to filter by service\n\n`;
    content += `**Golden Signal Queries**:\n`;
    content += `- **Latency**: \`${snapshot.schemaHints.preferredQueryPatterns.latency}\`\n`;
    content += `- **Throughput**: \`${snapshot.schemaHints.preferredQueryPatterns.throughput}\`\n`;
    content += `- **Errors**: \`${snapshot.schemaHints.preferredQueryPatterns.errors}\`\n\n`;
    content += `**Strategy**: ${snapshot.schemaHints.goldenSignalStrategy}\n\n`;

    // Entities
    content += `## 🏢 Monitored Entities\n\n`;
    if (snapshot.entities.length > 0) {
      const byDomain = this.groupEntitiesByDomain(snapshot.entities);
      for (const [domain, entities] of Object.entries(byDomain)) {
        content += `**${domain}** (${entities.length}):\n`;
        entities.slice(0, 5).forEach(entity => {
          const healthIcon = entity.healthStatus ? this.getHealthIcon(entity.healthStatus) : '';
          const goldenIcon = entity.goldenSignalsAvailable ? '📊' : '📉';
          content += `- ${healthIcon}${goldenIcon} **${entity.name}** (${entity.type})`;
          if (entity.language) content += ` - ${entity.language}`;
          if (entity.environment) content += ` [${entity.environment}]`;
          content += `\n`;
        });
        if (entities.length > 5) {
          content += `  ... and ${entities.length - 5} more\n`;
        }
        content += `\n`;
      }
    } else {
      content += `*No entities found - check instrumentation*\n\n`;
    }

    // Event Types
    content += `## 📋 Available Event Types\n\n`;
    snapshot.eventTypes.forEach(et => {
      const volume = et.sampleCount > 1000 ? '🔥 High' : et.sampleCount > 100 ? '📈 Medium' : '📊 Low';
      content += `- **${et.name}** - ${et.description} (${volume} volume)\n`;
      if (et.keyAttributes.length > 0) {
        content += `  *Key attributes*: ${et.keyAttributes.slice(0, 5).join(', ')}\n`;
      }
    });
    content += `\n`;

    // Metric Streams
    if (snapshot.metricStreams.length > 0) {
      content += `## 📈 Metric Streams\n\n`;
      snapshot.metricStreams.forEach(ms => {
        content += `- **${ms.category}** (${ms.type}, ${ms.count} metrics)\n`;
        content += `  *Examples*: ${ms.examples.join(', ')}\n`;
      });
      content += `\n`;
    }

    // Observability Assessment
    if (snapshot.observabilityGaps.length > 0 || snapshot.recommendations.length > 0) {
      content += `## 🔍 Observability Assessment\n\n`;
      
      if (snapshot.observabilityGaps.length > 0) {
        content += `**Identified Gaps**:\n`;
        snapshot.observabilityGaps.forEach(gap => {
          content += `- ⚠️ ${gap}\n`;
        });
        content += `\n`;
      }
      
      if (snapshot.recommendations.length > 0) {
        content += `**Recommendations**:\n`;
        snapshot.recommendations.forEach(rec => {
          content += `- 💡 ${rec}\n`;
        });
        content += `\n`;
      }
    }

    // Instrumentation Notes
    if (snapshot.schemaHints.instrumentationNotes.length > 0) {
      content += `## 🔧 Instrumentation Notes\n\n`;
      snapshot.schemaHints.instrumentationNotes.forEach(note => {
        content += `- ${note}\n`;
      });
      content += `\n`;
    }

    content += `---\n*Use this context to inform your queries and dashboard generation choices.*`;

    return {
      content: [
        {
          type: 'text',
          text: content,
        },
      ],
      isError: false,
    };
  }

  // Helper methods

  private isCacheValid(timestamp: Date): boolean {
    return Date.now() - timestamp.getTime() < this.CACHE_TTL_MS;
  }

  private getEventTypeDescription(eventType: string): string {
    const descriptions: Record<string, string> = {
      'Transaction': 'APM transactions (HTTP requests, background tasks)',
      'TransactionError': 'APM transaction errors and exceptions',
      'Span': 'Distributed tracing spans (likely OpenTelemetry)',
      'Log': 'Application and system logs',
      'Metric': 'Custom and dimensional metrics',
      'SystemSample': 'Host system metrics (CPU, memory, disk)',
      'ProcessSample': 'Process-level metrics',
      'NetworkSample': 'Network interface metrics',
      'BrowserInteraction': 'Browser Real User Monitoring',
      'PageView': 'Browser page view events',
      'MobileSession': 'Mobile application sessions',
      'SyntheticCheck': 'Synthetic monitoring results',
    };
    
    return descriptions[eventType] || 'Custom telemetry event';
  }

  private extractKeyAttributes(keySet: string[]): string[] {
    const important = ['error', 'duration', 'name', 'appName', 'service.name', 'host', 'entityGuid'];
    return keySet.filter(key => 
      important.includes(key) || key.includes('name') || key.includes('error')
    ).slice(0, 8);
  }

  private categorizeMetrics(metrics: string[]): Record<string, string[]> {
    const categories: Record<string, string[]> = {};
    
    for (const metric of metrics) {
      let category = 'Custom';
      
      if (metric.startsWith('http.') || metric.startsWith('grpc.')) {
        category = 'OpenTelemetry Protocol';
      } else if (metric.startsWith('service.') || metric.startsWith('process.')) {
        category = 'OpenTelemetry Runtime';
      } else if (metric.includes('cpu') || metric.includes('memory') || metric.includes('disk')) {
        category = 'Infrastructure';
      } else if (metric.includes('error') || metric.includes('exception')) {
        category = 'Error Tracking';
      } else if (metric.includes('duration') || metric.includes('latency') || metric.includes('response')) {
        category = 'Performance';
      }
      
      if (!categories[category]) categories[category] = [];
      categories[category].push(metric);
    }
    
    return categories;
  }

  private inferMetricType(category: string, metrics: string[]): 'otel' | 'custom' | 'infrastructure' | 'apm' {
    if (category.includes('OpenTelemetry')) return 'otel';
    if (category === 'Infrastructure') return 'infrastructure';
    if (metrics.some(m => m.includes('newrelic'))) return 'apm';
    return 'custom';
  }

  private summarizeEntityTypes(entities: EntitySummary[]): string {
    const counts: Record<string, number> = {};
    entities.forEach(e => {
      counts[e.domain] = (counts[e.domain] || 0) + 1;
    });
    
    return Object.entries(counts)
      .map(([domain, count]) => `${count} ${domain}`)
      .join(', ');
  }

  private groupEntitiesByDomain(entities: EntitySummary[]): Record<string, EntitySummary[]> {
    const groups: Record<string, EntitySummary[]> = {};
    entities.forEach(entity => {
      if (!groups[entity.domain]) groups[entity.domain] = [];
      groups[entity.domain].push(entity);
    });
    return groups;
  }

  private getHealthIcon(status: string): string {
    const icons: Record<string, string> = {
      'healthy': '✅',
      'warning': '⚠️',
      'critical': '🚨',
      'unknown': '❓',
    };
    return icons[status] || '';
  }

  private async checkEntityHealth(entityGuid: string): Promise<'healthy' | 'warning' | 'critical' | 'unknown'> {
    try {
      // Simple health check based on recent error rate
      const accountId = parseInt(process.env['NEW_RELIC_ACCOUNT_ID'] || '0');
      const query = `
        SELECT 
          filter(count(*), WHERE error IS true) as errors,
          count(*) as total
        FROM Transaction 
        WHERE entity.guid = '${entityGuid}'
        SINCE 30 minutes ago
      `;
      
      const result = await this.nerdgraph.nrql(accountId, query);
      const data = result.results[0] || {};
      
      if (data.total === 0) return 'unknown';
      
      const errorRate = (data.errors || 0) / data.total;
      if (errorRate > 0.1) return 'critical';
      if (errorRate > 0.01) return 'warning';
      return 'healthy';
    } catch (error: any) {
      return 'unknown';
    }
  }
}
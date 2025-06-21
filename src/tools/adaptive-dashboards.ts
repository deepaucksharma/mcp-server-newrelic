/**
 * Adaptive Dashboard Generation
 * 
 * Creates dashboards that automatically adapt to discovered schemas
 * without hardcoded field assumptions.
 */

import { PlatformDiscovery, EntityInfo, EntityDataDiscovery } from '../core/platform-discovery.js';
import { Logger } from '../core/types.js';

export interface DashboardTemplate {
  name: string;
  description: string;
  pages: PageTemplate[];
}

export interface PageTemplate {
  name: string;
  widgets: WidgetTemplate[];
}

export interface WidgetTemplate {
  intent: WidgetIntent;
  title: string;
  visualization: string;
  layout: WidgetLayout;
  fallbacks?: string[]; // Alternative intents if primary fails
}

export interface WidgetLayout {
  column: number;
  row: number;
  width: number;
  height: number;
}

export type WidgetIntent = 
  | 'error_rate'
  | 'latency_p95'
  | 'latency_p99'
  | 'throughput'
  | 'saturation_cpu'
  | 'saturation_memory'
  | 'top_errors'
  | 'service_map'
  | 'alert_status'
  | 'deployment_markers'
  | 'apdex'
  | 'database_time'
  | 'external_time'
  | 'log_patterns'
  | 'custom_metric';

export interface Widget {
  title: string;
  visualization: { id: string };
  configuration: {
    nrqlQueries: Array<{
      accountId: number;
      query: string;
    }>;
  };
  layout: WidgetLayout;
}

export interface DashboardDefinition {
  name: string;
  description?: string;
  pages: Array<{
    name: string;
    widgets: Widget[];
  }>;
}

export class AdaptiveDashboardGenerator {
  constructor(
    private discovery: PlatformDiscovery,
    private logger: Logger
  ) {}

  /**
   * Generate a dashboard that adapts to discovered entity data patterns
   */
  async generateDashboard(
    templateName: string,
    entity: EntityInfo,
    options: {
      timeRange?: string;
      customWidgets?: WidgetTemplate[];
      accountId: number;
    }
  ): Promise<DashboardDefinition> {
    this.logger.info('Generating adaptive dashboard', {
      template: templateName,
      entityGuid: entity.guid,
      entityType: entity.type,
    });

    // Discover data patterns for this entity
    const entityData = await this.discovery.discoverEntityData(entity);
    
    // Load template
    const template = this.getTemplate(templateName);
    
    // Adapt widgets to discovered schema
    const adaptedPages = await Promise.all(
      template.pages.map(async (page) => ({
        name: page.name,
        widgets: await this.adaptWidgetsToSchema(
          page.widgets,
          entityData,
          entity,
          options
        ),
      }))
    );

    return {
      name: `${entity.name} - ${template.name}`,
      description: `Auto-generated ${template.description} for ${entity.name}`,
      pages: adaptedPages,
    };
  }

  /**
   * Adapt widget templates to discovered entity schema
   */
  private async adaptWidgetsToSchema(
    widgetTemplates: WidgetTemplate[],
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    options: { timeRange?: string; accountId: number }
  ): Promise<Widget[]> {
    const adaptedWidgets: Widget[] = [];

    for (const template of widgetTemplates) {
      try {
        const widget = await this.adaptWidget(template, entityData, entity, options);
        if (widget) {
          adaptedWidgets.push(widget);
        }
      } catch (error: any) {
        this.logger.warn('Failed to adapt widget', {
          intent: template.intent,
          error: error.message,
        });

        // Try fallback intents
        if (template.fallbacks) {
          for (const fallbackIntent of template.fallbacks) {
            try {
              const fallbackTemplate = { ...template, intent: fallbackIntent as WidgetIntent };
              const widget = await this.adaptWidget(fallbackTemplate, entityData, entity, options);
              if (widget) {
                adaptedWidgets.push(widget);
                break;
              }
            } catch (fallbackError) {
              continue;
            }
          }
        }
      }
    }

    return adaptedWidgets;
  }

  /**
   * Adapt a single widget template to entity data
   */
  private async adaptWidget(
    template: WidgetTemplate,
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    options: { timeRange?: string; accountId: number }
  ): Promise<Widget | null> {
    const timeRange = options.timeRange || '1 hour ago';
    
    switch (template.intent) {
      case 'error_rate':
        return this.createErrorRateWidget(entityData, entity, template, timeRange, options.accountId);
      
      case 'latency_p95':
        return this.createLatencyWidget(entityData, entity, template, timeRange, options.accountId, 95);
      
      case 'latency_p99':
        return this.createLatencyWidget(entityData, entity, template, timeRange, options.accountId, 99);
      
      case 'throughput':
        return this.createThroughputWidget(entityData, entity, template, timeRange, options.accountId);
      
      case 'saturation_cpu':
        return this.createSaturationWidget(entityData, entity, template, timeRange, options.accountId, 'cpu');
      
      case 'saturation_memory':
        return this.createSaturationWidget(entityData, entity, template, timeRange, options.accountId, 'memory');
      
      case 'top_errors':
        return this.createTopErrorsWidget(entityData, entity, template, timeRange, options.accountId);
      
      case 'apdex':
        return this.createApdexWidget(entityData, entity, template, timeRange, options.accountId);
      
      case 'database_time':
        return this.createDatabaseTimeWidget(entityData, entity, template, timeRange, options.accountId);
      
      case 'external_time':
        return this.createExternalTimeWidget(entityData, entity, template, timeRange, options.accountId);
      
      default:
        this.logger.warn('Unknown widget intent', { intent: template.intent });
        return null;
    }
  }

  /**
   * Create error rate widget using discovered error indicators
   */
  private async createErrorRateWidget(
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    template: WidgetTemplate,
    timeRange: string,
    accountId: number
  ): Promise<Widget | null> {
    if (entityData.errorIndicators.length === 0) {
      throw new Error('No error indicators discovered');
    }

    const errorIndicator = entityData.errorIndicators[0];
    const serviceFilter = this.buildServiceFilter(entityData.serviceIdentifier, entity.name);
    const eventType = this.selectBestEventType(entityData.eventTypes, ['Transaction', 'Span', 'Log']);

    let condition: string;
    switch (errorIndicator.type) {
      case 'boolean':
        condition = `${errorIndicator.name} = true`;
        break;
      case 'http_status':
        condition = `numeric(${errorIndicator.name}) >= 400`;
        break;
      case 'error_class':
        condition = `${errorIndicator.name} IS NOT NULL`;
        break;
      default:
        condition = `${errorIndicator.name} = true`;
    }

    const query = `
      SELECT percentage(count(*), WHERE ${condition}) as 'Error Rate'
      FROM ${eventType}
      ${serviceFilter}
      SINCE ${timeRange}
      TIMESERIES
    `.trim().replace(/\s+/g, ' ');

    return {
      title: template.title,
      visualization: { id: template.visualization || 'viz.line' },
      configuration: {
        nrqlQueries: [{
          accountId,
          query,
        }],
      },
      layout: template.layout,
    };
  }

  /**
   * Create latency widget using discovered duration fields
   */
  private async createLatencyWidget(
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    template: WidgetTemplate,
    timeRange: string,
    accountId: number,
    percentile: number
  ): Promise<Widget | null> {
    // Try dimensional metrics first
    const latencyMetric = entityData.metrics.find(m => 
      m.name.toLowerCase().includes('duration') || 
      m.name.toLowerCase().includes('latency') ||
      m.name.toLowerCase().includes('response_time')
    );

    if (latencyMetric) {
      // Use dimensional metric approach
      const serviceDimension = this.findServiceDimension(latencyMetric.dimensions, entityData.serviceIdentifier);
      
      const query = `
        SELECT percentile(${latencyMetric.name}, ${percentile}) as 'P${percentile} Latency'
        FROM Metric
        WHERE metricName = '${latencyMetric.name}'
        ${serviceDimension ? `AND ${serviceDimension} = '${entity.name}'` : ''}
        SINCE ${timeRange}
        TIMESERIES
      `.trim().replace(/\s+/g, ' ');

      return {
        title: `${template.title} (Metric)`,
        visualization: { id: template.visualization || 'viz.line' },
        configuration: {
          nrqlQueries: [{
            accountId,
            query,
          }],
        },
        layout: template.layout,
      };
    }

    // Fallback to event-based approach
    if (entityData.durationFields.length === 0) {
      throw new Error('No duration fields discovered');
    }

    const durationField = entityData.durationFields[0];
    const serviceFilter = this.buildServiceFilter(entityData.serviceIdentifier, entity.name);
    const eventType = this.selectBestEventType(entityData.eventTypes, ['Transaction', 'Span']);

    const query = `
      SELECT percentile(${durationField}, ${percentile}) as 'P${percentile} Latency'
      FROM ${eventType}
      ${serviceFilter}
      SINCE ${timeRange}
      TIMESERIES
    `.trim().replace(/\s+/g, ' ');

    return {
      title: `${template.title} (Event)`,
      visualization: { id: template.visualization || 'viz.line' },
      configuration: {
        nrqlQueries: [{
          accountId,
          query,
        }],
      },
      layout: template.layout,
    };
  }

  /**
   * Create throughput widget
   */
  private async createThroughputWidget(
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    template: WidgetTemplate,
    timeRange: string,
    accountId: number
  ): Promise<Widget | null> {
    const serviceFilter = this.buildServiceFilter(entityData.serviceIdentifier, entity.name);
    const eventType = this.selectBestEventType(entityData.eventTypes, ['Transaction', 'Span', 'Log']);

    const query = `
      SELECT count(*) as 'Throughput'
      FROM ${eventType}
      ${serviceFilter}
      SINCE ${timeRange}
      TIMESERIES
    `.trim().replace(/\s+/g, ' ');

    return {
      title: template.title,
      visualization: { id: template.visualization || 'viz.line' },
      configuration: {
        nrqlQueries: [{
          accountId,
          query,
        }],
      },
      layout: template.layout,
    };
  }

  /**
   * Create saturation widget (CPU/Memory)
   */
  private async createSaturationWidget(
    entityData: EntityDataDiscovery,
    _entity: EntityInfo,
    template: WidgetTemplate,
    timeRange: string,
    accountId: number,
    resource: 'cpu' | 'memory'
  ): Promise<Widget | null> {
    // Look for system metrics
    const metricPattern = resource === 'cpu' ? 'cpu' : 'memory';
    const systemMetric = entityData.metrics.find(m => 
      m.name.toLowerCase().includes(metricPattern)
    );

    if (systemMetric) {
      const query = `
        SELECT average(${systemMetric.name}) as '${resource.toUpperCase()} %'
        FROM Metric
        WHERE metricName = '${systemMetric.name}'
        SINCE ${timeRange}
        TIMESERIES
      `.trim().replace(/\s+/g, ' ');

      return {
        title: template.title,
        visualization: { id: template.visualization || 'viz.line' },
        configuration: {
          nrqlQueries: [{
            accountId,
            query,
          }],
        },
        layout: template.layout,
      };
    }

    // Fallback to SystemSample
    const fieldName = resource === 'cpu' ? 'cpuPercent' : 'memoryUsedPercent';
    const eventType = 'SystemSample';

    if (!entityData.eventTypes.includes(eventType)) {
      throw new Error(`No ${resource} saturation data available`);
    }

    const query = `
      SELECT average(${fieldName}) as '${resource.toUpperCase()} %'
      FROM ${eventType}
      SINCE ${timeRange}
      TIMESERIES
    `.trim().replace(/\s+/g, ' ');

    return {
      title: template.title,
      visualization: { id: template.visualization || 'viz.line' },
      configuration: {
        nrqlQueries: [{
          accountId,
          query,
        }],
      },
      layout: template.layout,
    };
  }

  /**
   * Create top errors widget
   */
  private async createTopErrorsWidget(
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    template: WidgetTemplate,
    timeRange: string,
    accountId: number
  ): Promise<Widget | null> {
    if (entityData.errorIndicators.length === 0) {
      throw new Error('No error indicators discovered');
    }

    const errorIndicator = entityData.errorIndicators[0];
    const serviceFilter = this.buildServiceFilter(entityData.serviceIdentifier, entity.name);
    const eventType = this.selectBestEventType(entityData.eventTypes, ['Transaction', 'TransactionError', 'Span']);

    let condition: string;
    let facetField = 'error.class';

    switch (errorIndicator.type) {
      case 'error_class':
        condition = `${errorIndicator.name} IS NOT NULL`;
        facetField = errorIndicator.name;
        break;
      case 'boolean':
        condition = `${errorIndicator.name} = true`;
        facetField = 'name'; // Facet by transaction name
        break;
      case 'http_status':
        condition = `numeric(${errorIndicator.name}) >= 400`;
        facetField = errorIndicator.name;
        break;
      default:
        condition = `${errorIndicator.name} IS NOT NULL`;
    }

    const query = `
      SELECT count(*) as 'Error Count'
      FROM ${eventType}
      WHERE ${condition} ${serviceFilter ? `AND ${serviceFilter.replace('WHERE ', '')}` : ''}
      FACET ${facetField}
      SINCE ${timeRange}
      LIMIT 10
    `.trim().replace(/\s+/g, ' ');

    return {
      title: template.title,
      visualization: { id: template.visualization || 'viz.table' },
      configuration: {
        nrqlQueries: [{
          accountId,
          query,
        }],
      },
      layout: template.layout,
    };
  }

  /**
   * Create Apdex widget if available
   */
  private async createApdexWidget(
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    template: WidgetTemplate,
    timeRange: string,
    accountId: number
  ): Promise<Widget | null> {
    const serviceFilter = this.buildServiceFilter(entityData.serviceIdentifier, entity.name);
    
    // Check if Transaction data exists (required for Apdex)
    if (!entityData.eventTypes.includes('Transaction')) {
      throw new Error('Apdex requires Transaction data');
    }

    const query = `
      SELECT apdex(duration, t: 0.5) as 'Apdex Score'
      FROM Transaction
      ${serviceFilter}
      SINCE ${timeRange}
      TIMESERIES
    `.trim().replace(/\s+/g, ' ');

    return {
      title: template.title,
      visualization: { id: template.visualization || 'viz.line' },
      configuration: {
        nrqlQueries: [{
          accountId,
          query,
        }],
      },
      layout: template.layout,
    };
  }

  /**
   * Create database time widget
   */
  private async createDatabaseTimeWidget(
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    template: WidgetTemplate,
    timeRange: string,
    accountId: number
  ): Promise<Widget | null> {
    const dbField = entityData.durationFields.find(field => 
      field.toLowerCase().includes('database') || 
      field.toLowerCase().includes('db')
    );

    if (!dbField) {
      throw new Error('No database duration field discovered');
    }

    const serviceFilter = this.buildServiceFilter(entityData.serviceIdentifier, entity.name);
    const eventType = this.selectBestEventType(entityData.eventTypes, ['Transaction']);

    const query = `
      SELECT average(${dbField}) as 'Database Time'
      FROM ${eventType}
      ${serviceFilter}
      SINCE ${timeRange}
      TIMESERIES
    `.trim().replace(/\s+/g, ' ');

    return {
      title: template.title,
      visualization: { id: template.visualization || 'viz.line' },
      configuration: {
        nrqlQueries: [{
          accountId,
          query,
        }],
      },
      layout: template.layout,
    };
  }

  /**
   * Create external time widget
   */
  private async createExternalTimeWidget(
    entityData: EntityDataDiscovery,
    entity: EntityInfo,
    template: WidgetTemplate,
    timeRange: string,
    accountId: number
  ): Promise<Widget | null> {
    const externalField = entityData.durationFields.find(field => 
      field.toLowerCase().includes('external') || 
      field.toLowerCase().includes('http')
    );

    if (!externalField) {
      throw new Error('No external duration field discovered');
    }

    const serviceFilter = this.buildServiceFilter(entityData.serviceIdentifier, entity.name);
    const eventType = this.selectBestEventType(entityData.eventTypes, ['Transaction']);

    const query = `
      SELECT average(${externalField}) as 'External Time'
      FROM ${eventType}
      ${serviceFilter}
      SINCE ${timeRange}
      TIMESERIES
    `.trim().replace(/\s+/g, ' ');

    return {
      title: template.title,
      visualization: { id: template.visualization || 'viz.line' },
      configuration: {
        nrqlQueries: [{
          accountId,
          query,
        }],
      },
      layout: template.layout,
    };
  }

  // Helper methods
  private buildServiceFilter(serviceIdentifier: string, entityName: string): string {
    if (!serviceIdentifier || !entityName) {
      return '';
    }
    return `WHERE ${serviceIdentifier} = '${entityName}'`;
  }

  private selectBestEventType(availableTypes: string[], preferredTypes: string[]): string {
    for (const preferred of preferredTypes) {
      if (availableTypes.includes(preferred)) {
        return preferred;
      }
    }
    return availableTypes[0] || 'Transaction';
  }

  private findServiceDimension(dimensions: string[], _serviceIdentifier: string): string | null {
    // Look for service dimension in metrics
    const serviceDimensions = ['service.name', 'appName', 'entity.name', 'serviceName'];
    
    for (const serviceDim of serviceDimensions) {
      if (dimensions.includes(serviceDim)) {
        return serviceDim;
      }
    }
    
    return null;
  }

  /**
   * Get predefined dashboard templates
   */
  private getTemplate(templateName: string): DashboardTemplate {
    const templates: Record<string, DashboardTemplate> = {
      'golden-signals': {
        name: 'Golden Signals',
        description: 'Essential service health indicators',
        pages: [{
          name: 'Overview',
          widgets: [
            {
              intent: 'throughput',
              title: 'Throughput',
              visualization: 'viz.line',
              layout: { column: 1, row: 1, width: 4, height: 3 },
            },
            {
              intent: 'error_rate',
              title: 'Error Rate',
              visualization: 'viz.line',
              layout: { column: 5, row: 1, width: 4, height: 3 },
            },
            {
              intent: 'latency_p95',
              title: 'Latency (P95)',
              visualization: 'viz.line',
              layout: { column: 9, row: 1, width: 4, height: 3 },
            },
            {
              intent: 'saturation_cpu',
              title: 'CPU Saturation',
              visualization: 'viz.line',
              layout: { column: 1, row: 4, width: 6, height: 3 },
              fallbacks: ['saturation_memory'],
            },
            {
              intent: 'top_errors',
              title: 'Top Errors',
              visualization: 'viz.table',
              layout: { column: 7, row: 4, width: 6, height: 3 },
            },
          ],
        }],
      },
      
      'infrastructure': {
        name: 'Infrastructure Monitoring',
        description: 'System resource monitoring and alerts',
        pages: [{
          name: 'System Resources',
          widgets: [
            {
              intent: 'saturation_cpu',
              title: 'CPU Usage',
              visualization: 'viz.line',
              layout: { column: 1, row: 1, width: 6, height: 3 },
            },
            {
              intent: 'saturation_memory',
              title: 'Memory Usage',
              visualization: 'viz.line',
              layout: { column: 7, row: 1, width: 6, height: 3 },
            },
          ],
        }],
      },
    };

    const template = templates[templateName];
    if (!template) {
      throw new Error(`Unknown template: ${templateName}`);
    }

    return template;
  }
}
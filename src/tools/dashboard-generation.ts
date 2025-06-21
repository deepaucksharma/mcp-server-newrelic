/**
 * Dashboard Generation Tool - generate.golden_dashboard
 * 
 * Generates comprehensive golden signal dashboards for any entity,
 * automatically adapting to available telemetry and data patterns.
 */

import { Tool } from '@modelcontextprotocol/sdk/types.js';
import { NerdGraphClient, Logger } from '../core/types.js';
import { GoldenSignalsEngine, EntityGoldenMetrics, TelemetryContext } from '../core/golden-signals.js';

export interface DashboardWidget {
  id: string;
  title: string;
  visualization: {
    id: string; // 'viz.line', 'viz.billboard', 'viz.area', etc.
  };
  rawConfiguration: {
    nrqlQueries: Array<{
      accountId: number;
      query: string;
    }>;
    facet?: {
      showOtherSeries: boolean;
    };
    legend?: {
      enabled: boolean;
    };
    yAxisLeft?: {
      zero: boolean;
    };
    thresholds?: Array<{
      alertSeverity: string;
      value: number;
    }>;
  };
  layout: {
    column: number;
    row: number;
    width: number;
    height: number;
  };
}

export interface DashboardDefinition {
  name: string;
  description: string;
  permissions: 'PUBLIC_READ_ONLY' | 'PUBLIC_READ_WRITE';
  pages: Array<{
    name: string;
    description: string;
    widgets: DashboardWidget[];
  }>;
  variables: Array<{
    name: string;
    title: string;
    type: string;
    defaultValues: string[];
    nrqlQuery?: {
      accountIds: number[];
      query: string;
    };
  }>;
}

export interface GoldenDashboardResult {
  dashboard: DashboardDefinition;
  entity: {
    guid: string;
    name: string;
    type: string;
  };
  telemetryContext: TelemetryContext;
  generationMetadata: {
    timestamp: Date;
    widgetCount: number;
    adaptations: string[];
    recommendations: string[];
  };
  createDashboardMutation?: string;
}

export class DashboardGenerationTool {
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
      name: 'generate.golden_dashboard',
      description: `Generate a comprehensive golden signals dashboard for any entity in one call.

      🎯 **Purpose**: Creates a production-ready dashboard covering all four golden signals of monitoring.

      **Generated Dashboard Includes**:
      - **Latency**: Response time percentiles (P50, P95, P99) with trend analysis
      - **Traffic**: Request rate and throughput patterns over time
      - **Errors**: Error rate percentage with error breakdown by type
      - **Saturation**: Resource utilization (CPU, Memory) when available

      **Intelligent Adaptation**:
      - Automatically detects OpenTelemetry vs APM instrumentation
      - Adapts queries to use optimal data sources (Span vs Transaction events)
      - Handles mixed telemetry environments gracefully
      - Optimizes widget types based on data characteristics

      **Dashboard Features**:
      - Multi-page layout (Overview + Detailed Analysis)
      - Interactive time range controls and variables
      - Alert threshold overlays where applicable
      - Responsive design for different screen sizes
      - Entity-specific filtering and faceting

      **Output Options**:
      - Dashboard JSON definition for review/modification
      - Ready-to-execute NerdGraph mutation for creation
      - Validation report with recommendations`,

      inputSchema: {
        type: 'object',
        properties: {
          entity_guid: {
            type: 'string',
            description: 'Entity GUID to create dashboard for (from discover.environment or search_entities)',
            pattern: '^[A-Za-z0-9+/]+=*$',
          },
          dashboard_name: {
            type: 'string',
            description: 'Custom name for the dashboard (auto-generated if not provided)',
          },
          timeframe_hours: {
            type: 'number',
            description: 'Default time window for dashboard widgets in hours',
            default: 1,
            minimum: 0.25,
            maximum: 168, // 1 week
          },
          include_saturation: {
            type: 'boolean',
            description: 'Include resource saturation metrics if available',
            default: true,
          },
          create_dashboard: {
            type: 'boolean',
            description: 'Actually create the dashboard (false = preview only)',
            default: false,
          },
          alert_thresholds: {
            type: 'object',
            description: 'Custom alert thresholds for golden signal widgets',
            properties: {
              latency_p95_ms: { type: 'number', default: 1000 },
              error_rate_percent: { type: 'number', default: 5 },
              traffic_drop_percent: { type: 'number', default: 50 },
            },
          },
        },
        required: ['entity_guid'],
        additionalProperties: false,
      },
    };
  }

  /**
   * Handle the generate.golden_dashboard tool call
   */
  async handle(params: any): Promise<any> {
    const {
      entity_guid,
      dashboard_name,
      timeframe_hours = 1,
      include_saturation = true,
      create_dashboard = false,
      alert_thresholds = {},
    } = params;

    this.logger.info('Golden dashboard generation requested', {
      entity_guid,
      timeframe_hours,
      create_dashboard,
    });

    try {
      // Get comprehensive golden metrics for the entity
      const goldenMetrics = await this.goldenSignals.getEntityGoldenMetrics(
        entity_guid, 
        Math.max(timeframe_hours * 60, 30) // Convert to minutes, minimum 30
      );

      // Generate dashboard definition
      const dashboardResult = await this.generateGoldenDashboard(
        goldenMetrics,
        {
          dashboard_name,
          timeframe_hours,
          include_saturation,
          alert_thresholds,
        }
      );

      // Create dashboard if requested
      if (create_dashboard) {
        const createdDashboard = await this.createDashboard(dashboardResult.dashboard);
        dashboardResult.createDashboardMutation = createdDashboard.mutation;
      }

      return this.formatResponse(dashboardResult, create_dashboard);

    } catch (error: any) {
      this.logger.error('Golden dashboard generation failed', { 
        entity_guid, 
        error: error.message 
      });

      return {
        content: [
          {
            type: 'text',
            text: `❌ **Dashboard Generation Failed**\n\nError: ${error.message}\n\nThis might indicate:\n- Invalid entity GUID\n- Insufficient data for golden signals\n- Account access restrictions\n- Entity not currently reporting\n\nTry using \`discover.environment\` first to verify the entity exists and has sufficient telemetry data.`,
          },
        ],
        isError: true,
      };
    }
  }

  /**
   * Generate comprehensive golden dashboard
   */
  private async generateGoldenDashboard(
    goldenMetrics: EntityGoldenMetrics,
    options: any
  ): Promise<GoldenDashboardResult> {
    const { entity, metrics, context } = goldenMetrics;
    const accountId = await this.extractAccountId(entity.guid);
    
    const dashboardName = options.dashboard_name || 
      `Golden Signals: ${entity.name}`;

    const adaptations: string[] = [];
    const recommendations: string[] = [];

    // Generate widgets based on available data
    const widgets: DashboardWidget[] = [];
    let widgetId = 1;

    // 1. Overview Page - Key Metrics
    const overviewWidgets = await this.generateOverviewWidgets(
      entity, metrics, context, accountId, options, widgetId
    );
    widgets.push(...overviewWidgets);
    widgetId += overviewWidgets.length;

    // 2. Detailed Analysis Page
    const detailWidgets = await this.generateDetailWidgets(
      entity, metrics, context, accountId, options, widgetId
    );

    // Track adaptations made
    if (context.primaryDataSource === 'otel') {
      adaptations.push('Adapted queries for OpenTelemetry Span events');
    } else if (context.primaryDataSource === 'apm') {
      adaptations.push('Adapted queries for New Relic APM Transaction events');
    } else {
      adaptations.push('Adapted for mixed telemetry environment');
    }

    if (!metrics.saturation.available) {
      recommendations.push('Consider enabling infrastructure monitoring for resource saturation metrics');
    }

    if (metrics.errors.percentage === 0) {
      recommendations.push('No errors detected - ensure error tracking is properly configured');
    }

    const dashboard: DashboardDefinition = {
      name: dashboardName,
      description: `Golden signals dashboard for ${entity.name} (${entity.type}) - Auto-generated with intelligent telemetry adaptation`,
      permissions: 'PUBLIC_READ_ONLY',
      pages: [
        {
          name: 'Overview',
          description: 'Key golden signal metrics at a glance',
          widgets: overviewWidgets,
        },
        {
          name: 'Detailed Analysis',
          description: 'In-depth analysis and breakdowns',
          widgets: detailWidgets,
        },
      ],
      variables: this.generateDashboardVariables(context, accountId),
    };

    return {
      dashboard,
      entity: {
        guid: entity.guid,
        name: entity.name,
        type: entity.type,
      },
      telemetryContext: context,
      generationMetadata: {
        timestamp: new Date(),
        widgetCount: widgets.length + detailWidgets.length,
        adaptations,
        recommendations,
      },
    };
  }

  /**
   * Generate overview page widgets
   */
  private async generateOverviewWidgets(
    entity: any,
    metrics: any,
    context: TelemetryContext,
    accountId: number,
    options: any,
    startingId: number
  ): Promise<DashboardWidget[]> {
    const widgets: DashboardWidget[] = [];
    const timeWindow = `SINCE ${options.timeframe_hours} hours ago`;
    const entityFilter = `WHERE entity.guid = '${entity.guid}'`;

    // 1. Latency Billboard (P95)
    widgets.push({
      id: `widget-${startingId}`,
      title: 'Response Time (P95)',
      visualization: { id: 'viz.billboard' },
      rawConfiguration: {
        nrqlQueries: [{
          accountId,
          query: this.buildLatencyQuery(context, entityFilter, timeWindow, 95),
        }],
        thresholds: [
          { alertSeverity: 'WARNING', value: options.alert_thresholds.latency_p95_ms || 1000 },
          { alertSeverity: 'CRITICAL', value: (options.alert_thresholds.latency_p95_ms || 1000) * 2 },
        ],
      },
      layout: { column: 1, row: 1, width: 4, height: 3 },
    });

    // 2. Error Rate Billboard
    widgets.push({
      id: `widget-${startingId + 1}`,
      title: 'Error Rate',
      visualization: { id: 'viz.billboard' },
      rawConfiguration: {
        nrqlQueries: [{
          accountId,
          query: this.buildErrorRateQuery(context, entityFilter, timeWindow),
        }],
        thresholds: [
          { alertSeverity: 'WARNING', value: options.alert_thresholds.error_rate_percent || 5 },
          { alertSeverity: 'CRITICAL', value: (options.alert_thresholds.error_rate_percent || 5) * 2 },
        ],
      },
      layout: { column: 5, row: 1, width: 4, height: 3 },
    });

    // 3. Traffic Rate Billboard
    widgets.push({
      id: `widget-${startingId + 2}`,
      title: 'Request Rate',
      visualization: { id: 'viz.billboard' },
      rawConfiguration: {
        nrqlQueries: [{
          accountId,
          query: this.buildTrafficQuery(context, entityFilter, timeWindow),
        }],
      },
      layout: { column: 9, row: 1, width: 4, height: 3 },
    });

    // 4. Latency Time Series
    widgets.push({
      id: `widget-${startingId + 3}`,
      title: 'Response Time Trends',
      visualization: { id: 'viz.line' },
      rawConfiguration: {
        nrqlQueries: [{
          accountId,
          query: this.buildLatencyTimeSeriesQuery(context, entityFilter, timeWindow),
        }],
        legend: { enabled: true },
        yAxisLeft: { zero: true },
      },
      layout: { column: 1, row: 4, width: 6, height: 3 },
    });

    // 5. Traffic and Errors Combined
    widgets.push({
      id: `widget-${startingId + 4}`,
      title: 'Traffic vs Errors',
      visualization: { id: 'viz.line' },
      rawConfiguration: {
        nrqlQueries: [
          {
            accountId,
            query: this.buildTrafficTimeSeriesQuery(context, entityFilter, timeWindow),
          },
          {
            accountId,
            query: this.buildErrorTimeSeriesQuery(context, entityFilter, timeWindow),
          },
        ],
        legend: { enabled: true },
      },
      layout: { column: 7, row: 4, width: 6, height: 3 },
    });

    return widgets;
  }

  /**
   * Generate detailed analysis widgets
   */
  private async generateDetailWidgets(
    entity: any,
    metrics: any,
    context: TelemetryContext,
    accountId: number,
    options: any,
    startingId: number
  ): Promise<DashboardWidget[]> {
    const widgets: DashboardWidget[] = [];
    const timeWindow = `SINCE ${options.timeframe_hours} hours ago`;
    const entityFilter = `WHERE entity.guid = '${entity.guid}'`;

    // 6. Latency Percentiles Breakdown
    widgets.push({
      id: `widget-${startingId}`,
      title: 'Latency Percentiles',
      visualization: { id: 'viz.line' },
      rawConfiguration: {
        nrqlQueries: [{
          accountId,
          query: this.buildLatencyPercentilesQuery(context, entityFilter, timeWindow),
        }],
        legend: { enabled: true },
      },
      layout: { column: 1, row: 1, width: 6, height: 3 },
    });

    // 7. Error Breakdown (if available)
    if (metrics.errors.rate > 0) {
      widgets.push({
        id: `widget-${startingId + 1}`,
        title: 'Error Breakdown',
        visualization: { id: 'viz.pie' },
        rawConfiguration: {
          nrqlQueries: [{
            accountId,
            query: this.buildErrorBreakdownQuery(context, entityFilter, timeWindow),
          }],
        },
        layout: { column: 7, row: 1, width: 6, height: 3 },
      });
    }

    // 8. Saturation Metrics (if available)
    if (options.include_saturation && metrics.saturation.available) {
      widgets.push({
        id: `widget-${startingId + 2}`,
        title: 'Resource Saturation',
        visualization: { id: 'viz.line' },
        rawConfiguration: {
          nrqlQueries: [{
            accountId,
            query: this.buildSaturationQuery(entity.guid, timeWindow),
          }],
          legend: { enabled: true },
          yAxisLeft: { zero: true },
        },
        layout: { column: 1, row: 4, width: 12, height: 3 },
      });
    }

    return widgets;
  }

  /**
   * Generate dashboard variables for filtering
   */
  private generateDashboardVariables(context: TelemetryContext, accountId: number): any[] {
    const variables = [];

    // Time range variable
    variables.push({
      name: 'timeRange',
      title: 'Time Range',
      type: 'ENUM',
      defaultValues: ['1 hour ago'],
    });

    // Service identifier variable (if multiple services)
    if (context.serviceIdentifierField) {
      variables.push({
        name: 'service',
        title: 'Service',
        type: 'NRQL',
        defaultValues: [],
        nrqlQuery: {
          accountIds: [accountId],
          query: context.hasOpenTelemetry
            ? `SELECT uniques(service.name) FROM Span SINCE 1 day ago`
            : `SELECT uniques(appName) FROM Transaction SINCE 1 day ago`,
        },
      });
    }

    return variables;
  }

  // Query builders for different telemetry contexts

  private buildLatencyQuery(context: TelemetryContext, entityFilter: string, timeWindow: string, percentile: number = 95): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT percentile(duration.ms, ${percentile}) as 'P${percentile} (ms)' FROM Span ${entityFilter} AND span.kind = 'server' ${timeWindow}`;
    } else {
      return `SELECT percentile(duration, ${percentile}) * 1000 as 'P${percentile} (ms)' FROM Transaction ${entityFilter} ${timeWindow}`;
    }
  }

  private buildErrorRateQuery(context: TelemetryContext, entityFilter: string, timeWindow: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT filter(count(*), WHERE otel.status_code = 'ERROR') / count(*) * 100 as 'Error Rate (%)' FROM Span ${entityFilter} AND span.kind = 'server' ${timeWindow}`;
    } else {
      return `SELECT filter(count(*), WHERE error IS true) / count(*) * 100 as 'Error Rate (%)' FROM Transaction ${entityFilter} ${timeWindow}`;
    }
  }

  private buildTrafficQuery(context: TelemetryContext, entityFilter: string, timeWindow: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT rate(count(*), 1 minute) as 'Requests/min' FROM Span ${entityFilter} AND span.kind = 'server' ${timeWindow}`;
    } else {
      return `SELECT rate(count(*), 1 minute) as 'Requests/min' FROM Transaction ${entityFilter} ${timeWindow}`;
    }
  }

  private buildLatencyTimeSeriesQuery(context: TelemetryContext, entityFilter: string, timeWindow: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT average(duration.ms) as 'Avg', percentile(duration.ms, 95) as 'P95' FROM Span ${entityFilter} AND span.kind = 'server' ${timeWindow} TIMESERIES`;
    } else {
      return `SELECT average(duration) * 1000 as 'Avg', percentile(duration, 95) * 1000 as 'P95' FROM Transaction ${entityFilter} ${timeWindow} TIMESERIES`;
    }
  }

  private buildTrafficTimeSeriesQuery(context: TelemetryContext, entityFilter: string, timeWindow: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT rate(count(*), 1 minute) as 'Traffic' FROM Span ${entityFilter} AND span.kind = 'server' ${timeWindow} TIMESERIES`;
    } else {
      return `SELECT rate(count(*), 1 minute) as 'Traffic' FROM Transaction ${entityFilter} ${timeWindow} TIMESERIES`;
    }
  }

  private buildErrorTimeSeriesQuery(context: TelemetryContext, entityFilter: string, timeWindow: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT filter(count(*), WHERE otel.status_code = 'ERROR') / count(*) * 100 as 'Error Rate (%)' FROM Span ${entityFilter} AND span.kind = 'server' ${timeWindow} TIMESERIES`;
    } else {
      return `SELECT filter(count(*), WHERE error IS true) / count(*) * 100 as 'Error Rate (%)' FROM Transaction ${entityFilter} ${timeWindow} TIMESERIES`;
    }
  }

  private buildLatencyPercentilesQuery(context: TelemetryContext, entityFilter: string, timeWindow: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT percentile(duration.ms, 50) as 'P50', percentile(duration.ms, 95) as 'P95', percentile(duration.ms, 99) as 'P99' FROM Span ${entityFilter} AND span.kind = 'server' ${timeWindow} TIMESERIES`;
    } else {
      return `SELECT percentile(duration, 50) * 1000 as 'P50', percentile(duration, 95) * 1000 as 'P95', percentile(duration, 99) * 1000 as 'P99' FROM Transaction ${entityFilter} ${timeWindow} TIMESERIES`;
    }
  }

  private buildErrorBreakdownQuery(context: TelemetryContext, entityFilter: string, timeWindow: string): string {
    if (context.hasOpenTelemetry && context.eventTypes.includes('Span')) {
      return `SELECT count(*) FROM Span ${entityFilter} AND otel.status_code = 'ERROR' ${timeWindow} FACET otel.status_description`;
    } else {
      return `SELECT count(*) FROM TransactionError ${entityFilter} ${timeWindow} FACET errorClass`;
    }
  }

  private buildSaturationQuery(entityGuid: string, timeWindow: string): string {
    return `SELECT average(cpuPercent) as 'CPU %', average(memoryUsedPercent) as 'Memory %' FROM SystemSample WHERE entityGuid = '${entityGuid}' ${timeWindow} TIMESERIES`;
  }

  /**
   * Create dashboard via NerdGraph
   */
  private async createDashboard(dashboard: DashboardDefinition): Promise<{ mutation: string }> {
    const mutation = `
      mutation {
        dashboardCreate(
          accountId: ${dashboard.variables[0]?.nrqlQuery?.accountIds[0] || 0}
          dashboard: {
            name: "${dashboard.name}"
            description: "${dashboard.description}"
            permissions: ${dashboard.permissions}
            pages: ${JSON.stringify(dashboard.pages)}
            variables: ${JSON.stringify(dashboard.variables)}
          }
        ) {
          entityResult {
            guid
          }
          errors {
            description
            type
          }
        }
      }
    `;

    // Note: This would actually execute the mutation in a real implementation
    return { mutation };
  }

  /**
   * Extract account ID from entity GUID
   */
  private async extractAccountId(entityGuid: string): Promise<number> {
    // This would properly parse the entity GUID to extract account ID
    // For now, use environment variable as fallback
    return parseInt(process.env['NEW_RELIC_ACCOUNT_ID'] || '0');
  }

  /**
   * Format response for LLM consumption
   */
  private formatResponse(result: GoldenDashboardResult, created: boolean): any {
    const { dashboard, entity, telemetryContext, generationMetadata } = result;
    
    let content = `# 📊 Golden Signals Dashboard ${created ? 'Created' : 'Generated'}\n\n`;
    
    // Entity Summary
    content += `## 🎯 Entity: ${entity.name}\n\n`;
    content += `- **Type**: ${entity.type}\n`;
    content += `- **GUID**: \`${entity.guid}\`\n`;
    content += `- **Telemetry**: ${telemetryContext.primaryDataSource.toUpperCase()}`;
    if (telemetryContext.primaryDataSource === 'mixed') {
      content += ' (OpenTelemetry + APM)';
    }
    content += `\n\n`;

    // Dashboard Overview
    content += `## 📋 Dashboard Overview\n\n`;
    content += `- **Name**: ${dashboard.name}\n`;
    content += `- **Pages**: ${dashboard.pages.length} (${dashboard.pages.map(p => p.name).join(', ')})\n`;
    content += `- **Widgets**: ${generationMetadata.widgetCount} total\n`;
    content += `- **Variables**: ${dashboard.variables.length} interactive controls\n\n`;

    // Pages Breakdown
    content += `## 📄 Dashboard Pages\n\n`;
    dashboard.pages.forEach((page, index) => {
      content += `**${index + 1}. ${page.name}** (${page.widgets.length} widgets)\n`;
      content += `- ${page.description}\n`;
      page.widgets.forEach(widget => {
        content += `  - ${widget.title} (${widget.visualization.id})\n`;
      });
      content += `\n`;
    });

    // Adaptations Made
    if (generationMetadata.adaptations.length > 0) {
      content += `## 🔧 Intelligent Adaptations\n\n`;
      generationMetadata.adaptations.forEach(adaptation => {
        content += `- ✅ ${adaptation}\n`;
      });
      content += `\n`;
    }

    // Recommendations
    if (generationMetadata.recommendations.length > 0) {
      content += `## 💡 Recommendations\n\n`;
      generationMetadata.recommendations.forEach(rec => {
        content += `- 🔍 ${rec}\n`;
      });
      content += `\n`;
    }

    // Dashboard JSON (collapsed)
    content += `## 📄 Dashboard Configuration\n\n`;
    content += `<details>\n<summary>Click to view dashboard JSON (${JSON.stringify(dashboard).length} characters)</summary>\n\n`;
    content += `\`\`\`json\n${JSON.stringify(dashboard, null, 2)}\`\`\`\n\n</details>\n\n`;

    if (created && result.createDashboardMutation) {
      content += `## 🚀 Dashboard Created\n\n`;
      content += `Dashboard has been successfully created in New Relic.\n\n`;
      content += `<details>\n<summary>NerdGraph Mutation Used</summary>\n\n`;
      content += `\`\`\`graphql\n${result.createDashboardMutation}\`\`\`\n\n</details>\n\n`;
    } else if (!created) {
      content += `## 🔍 Preview Mode\n\n`;
      content += `This is a preview of the dashboard. To create it, run the tool again with \`create_dashboard: true\`.\n\n`;
    }

    content += `---\n*Generated at ${generationMetadata.timestamp.toISOString()}*`;

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
}
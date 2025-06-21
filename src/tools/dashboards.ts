/**
 * Dashboard Tools - Visualization and Dashboard Management
 * 
 * These tools provide dashboard creation and management capabilities
 * using discovered data patterns.
 */

import { z } from 'zod';
import { ToolDefinition, RequestContext } from '../core/types.js';
import { DiscoveryEngine } from '../core/discovery/engine.js';

// ============================================================================
// Input/Output Schemas
// ============================================================================

const CreateDashboardInputSchema = z.object({
  name: z.string(),
  serviceName: z.string().optional(),
  type: z.enum(['service_overview', 'error_dashboard', 'performance_dashboard', 'custom']).default('service_overview'),
  accountId: z.number().optional(),
});

const ListDashboardsInputSchema = z.object({
  accountId: z.number().optional(),
  limit: z.number().min(1).max(100).default(20),
});

const GetDashboardInputSchema = z.object({
  dashboardId: z.string(),
  accountId: z.number().optional(),
});

const SuggestChartsInputSchema = z.object({
  serviceName: z.string().optional(),
  context: z.string(),
  maxCharts: z.number().min(1).max(10).default(5),
  accountId: z.number().optional(),
});

// ============================================================================
// Tool Implementations
// ============================================================================

/**
 * Create a dashboard using discovered data patterns
 */
const createDashboard: ToolDefinition<
  z.infer<typeof CreateDashboardInputSchema>,
  any
> = {
  name: 'dashboard.create',
  description: 'Create dashboards with intelligent chart suggestions based on discovered data',
  requiresDiscovery: true,
  inputSchema: CreateDashboardInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    // Build or use existing world model
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: `Creating ${input.type} dashboard using discovered patterns`,
      confidence: worldModel.confidence,
    });

    const serviceField = worldModel.serviceIdentifier.field;
    const errorIndicators = worldModel.errorIndicators;
    const metrics = worldModel.metrics;

    // Build dashboard configuration based on type
    let dashboardConfig: any = {
      name: input.name,
      accountId,
      type: input.type,
      description: `Auto-generated dashboard using discovered data patterns (${(worldModel.confidence * 100).toFixed(1)}% confidence)`,
      pages: [],
    };

    // Service filter if specified
    const serviceFilter = input.serviceName ? `WHERE ${serviceField} = '${input.serviceName}'` : '';

    switch (input.type) {
      case 'service_overview':
        dashboardConfig.pages.push({
          name: 'Service Overview',
          widgets: [
            {
              title: 'Throughput',
              visualization: 'line',
              nrql: `SELECT count(*) as 'Requests' FROM Transaction ${serviceFilter} TIMESERIES`,
              description: 'Request volume over time',
            },
            ...(errorIndicators.length > 0 ? [{
              title: 'Error Rate',
              visualization: 'line',
              nrql: `SELECT percentage(count(*), WHERE ${errorIndicators[0].condition}) as 'Error Rate' FROM ${errorIndicators[0].eventType} ${serviceFilter} TIMESERIES`,
              description: `Error rate using discovered indicator: ${errorIndicators[0].field}`,
            }] : []),
            ...(metrics.filter(m => m.field.includes('duration')).length > 0 ? [{
              title: 'Response Time',
              visualization: 'line',
              nrql: `SELECT average(${metrics.find(m => m.field.includes('duration'))?.field}) as 'Avg Response Time', percentile(${metrics.find(m => m.field.includes('duration'))?.field}, 95) as 'P95' FROM Transaction ${serviceFilter} TIMESERIES`,
              description: `Response time using discovered metric: ${metrics.find(m => m.field.includes('duration'))?.field}`,
            }] : []),
            {
              title: 'Top Services',
              visualization: 'table',
              nrql: `SELECT count(*) as 'Requests' FROM Transaction ${input.serviceName ? '' : `FACET ${serviceField}`} SINCE 1 hour ago LIMIT 10`,
              description: `Services breakdown using discovered identifier: ${serviceField}`,
            },
          ],
        });
        break;

      case 'error_dashboard':
        if (errorIndicators.length === 0) {
          return {
            accountId,
            error: 'No error indicators discovered',
            suggestion: 'Cannot create error dashboard without error indicators. Try implementing standardized error tracking.',
          };
        }

        dashboardConfig.pages.push({
          name: 'Error Analysis',
          widgets: errorIndicators.map((indicator, index) => [
            {
              title: `Error Rate - ${indicator.field}`,
              visualization: 'billboard',
              nrql: `SELECT percentage(count(*), WHERE ${indicator.condition}) as 'Error Rate' FROM ${indicator.eventType} ${serviceFilter} SINCE 1 hour ago`,
              description: `Error rate for ${indicator.field}`,
            },
            {
              title: `Error Trend - ${indicator.field}`,
              visualization: 'line',
              nrql: `SELECT percentage(count(*), WHERE ${indicator.condition}) as 'Error Rate' FROM ${indicator.eventType} ${serviceFilter} TIMESERIES`,
              description: `Error trend for ${indicator.field}`,
            },
          ]).flat(),
        });
        break;

      case 'performance_dashboard':
        const performanceMetrics = metrics.filter(m => 
          m.field.includes('duration') || 
          m.field.includes('response') || 
          m.field.includes('latency')
        );

        if (performanceMetrics.length === 0) {
          return {
            accountId,
            error: 'No performance metrics discovered',
            suggestion: 'Cannot create performance dashboard without latency metrics. Ensure APM instrumentation is configured.',
          };
        }

        dashboardConfig.pages.push({
          name: 'Performance Metrics',
          widgets: performanceMetrics.map(metric => [
            {
              title: `${metric.field} - Average`,
              visualization: 'billboard',
              nrql: `SELECT average(${metric.field}) as 'Average' FROM ${metric.eventType} ${serviceFilter} SINCE 1 hour ago`,
              description: `Average ${metric.field}`,
            },
            {
              title: `${metric.field} - Percentiles`,
              visualization: 'line',
              nrql: `SELECT percentile(${metric.field}, 50, 95, 99) FROM ${metric.eventType} ${serviceFilter} TIMESERIES`,
              description: `Percentile distribution for ${metric.field}`,
            },
          ]).flat(),
        });
        break;

      case 'custom':
        // Provide a basic template for custom dashboards
        dashboardConfig.pages.push({
          name: 'Custom View',
          widgets: [
            {
              title: 'Data Overview',
              visualization: 'table',
              nrql: `SELECT count(*) as 'Events' FROM ${worldModel.schemas[0]?.eventType || 'Transaction'} FACET ${serviceField} SINCE 1 hour ago LIMIT 10`,
              description: 'Basic data overview - customize as needed',
            },
          ],
        });
        break;
    }

    // Simulate dashboard creation (in real implementation, this would call New Relic Dashboard API)
    const dashboardId = `dashboard_${Date.now()}`;

    const result = {
      accountId,
      dashboardId,
      name: input.name,
      type: input.type,
      serviceName: input.serviceName,
      config: dashboardConfig,
      discoveryInfo: {
        worldModelConfidence: worldModel.confidence,
        serviceIdentifier: serviceField,
        errorIndicatorsUsed: errorIndicators.length,
        metricsUsed: metrics.length,
      },
      nextSteps: [
        'Dashboard configuration generated based on discovered patterns',
        'Use New Relic Dashboard API or UI to create the actual dashboard',
        'Customize chart titles and descriptions as needed',
        'Add additional filters or segments based on your specific needs',
      ],
    };

    ctx.explainabilityTrace.addStep({
      type: 'workflow_step',
      description: `Generated ${input.type} dashboard with ${dashboardConfig.pages[0]?.widgets?.length || 0} widgets`,
      confidence: worldModel.confidence,
    });

    return result;
  },
};

/**
 * List existing dashboards
 */
const listDashboards: ToolDefinition<
  z.infer<typeof ListDashboardsInputSchema>,
  any
> = {
  name: 'dashboard.list',
  description: 'List existing dashboards in the account',
  requiresDiscovery: false,
  inputSchema: ListDashboardsInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    try {
      // Query for dashboard entities
      const dashboardQuery = `
        query($accountId: Int!) {
          actor {
            entitySearch(query: "type = 'DASHBOARD' AND accountId = $accountId") {
              results {
                entities {
                  ... on DashboardEntity {
                    guid
                    name
                    accountId
                    createdAt
                    updatedAt
                    owner {
                      email
                    }
                    permissions
                  }
                }
              }
            }
          }
        }
      `;

      const result = await ctx.nerdgraph.request(dashboardQuery, { accountId });
      const dashboards = result.actor?.entitySearch?.results?.entities || [];

      ctx.explainabilityTrace.addStep({
        type: 'query',
        description: `Retrieved ${dashboards.length} dashboards`,
        resultCount: dashboards.length,
        confidence: 0.9,
      });

      return {
        accountId,
        totalDashboards: dashboards.length,
        dashboards: dashboards.slice(0, input.limit).map((dashboard: any) => ({
          guid: dashboard.guid,
          name: dashboard.name,
          createdAt: dashboard.createdAt,
          updatedAt: dashboard.updatedAt,
          owner: dashboard.owner?.email,
          permissions: dashboard.permissions,
        })),
      };

    } catch (error: any) {
      ctx.logger.error('Failed to list dashboards', error);
      throw new Error(`Failed to list dashboards: ${error.message}`);
    }
  },
};

/**
 * Get dashboard details
 */
const getDashboard: ToolDefinition<
  z.infer<typeof GetDashboardInputSchema>,
  any
> = {
  name: 'dashboard.get',
  description: 'Get details of a specific dashboard',
  requiresDiscovery: false,
  inputSchema: GetDashboardInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    try {
      const dashboardQuery = `
        query($guid: EntityGuid!) {
          actor {
            entity(guid: $guid) {
              ... on DashboardEntity {
                guid
                name
                accountId
                description
                createdAt
                updatedAt
                pages {
                  guid
                  name
                  widgets {
                    id
                    title
                    visualization {
                      id
                    }
                    rawConfiguration
                  }
                }
                permissions
                owner {
                  email
                }
              }
            }
          }
        }
      `;

      const result = await ctx.nerdgraph.request(dashboardQuery, { guid: input.dashboardId });
      const dashboard = result.actor?.entity;

      if (!dashboard) {
        throw new Error(`Dashboard not found: ${input.dashboardId}`);
      }

      ctx.explainabilityTrace.addStep({
        type: 'query',
        description: `Retrieved dashboard details: ${dashboard.name}`,
        confidence: 0.9,
      });

      return {
        accountId,
        dashboard: {
          guid: dashboard.guid,
          name: dashboard.name,
          description: dashboard.description,
          createdAt: dashboard.createdAt,
          updatedAt: dashboard.updatedAt,
          owner: dashboard.owner?.email,
          permissions: dashboard.permissions,
          pages: dashboard.pages?.map((page: any) => ({
            guid: page.guid,
            name: page.name,
            widgetCount: page.widgets?.length || 0,
            widgets: page.widgets?.map((widget: any) => ({
              id: widget.id,
              title: widget.title,
              visualization: widget.visualization?.id,
              configuration: widget.rawConfiguration,
            })) || [],
          })) || [],
        },
      };

    } catch (error: any) {
      ctx.logger.error('Failed to get dashboard', error);
      throw new Error(`Failed to get dashboard: ${error.message}`);
    }
  },
};

/**
 * Suggest chart configurations based on context
 */
const suggestCharts: ToolDefinition<
  z.infer<typeof SuggestChartsInputSchema>,
  any
> = {
  name: 'dashboard.suggest_charts',
  description: 'Suggest chart configurations based on context and discovered data',
  requiresDiscovery: true,
  inputSchema: SuggestChartsInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    // Build or use existing world model
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    const context = input.context.toLowerCase();
    const serviceField = worldModel.serviceIdentifier.field;
    const errorIndicators = worldModel.errorIndicators;
    const metrics = worldModel.metrics;
    const serviceFilter = input.serviceName ? `WHERE ${serviceField} = '${input.serviceName}'` : '';

    const suggestions = [];

    // Context-based chart suggestions
    if (context.includes('error') || context.includes('failure')) {
      errorIndicators.forEach(indicator => {
        suggestions.push({
          title: `Error Rate - ${indicator.field}`,
          type: 'billboard',
          nrql: `SELECT percentage(count(*), WHERE ${indicator.condition}) as 'Error Rate' FROM ${indicator.eventType} ${serviceFilter} SINCE 1 hour ago`,
          confidence: indicator.confidence,
          reason: `Error indicator discovered: ${indicator.field}`,
        });

        suggestions.push({
          title: `Error Trend - ${indicator.field}`,
          type: 'line',
          nrql: `SELECT percentage(count(*), WHERE ${indicator.condition}) as 'Error Rate' FROM ${indicator.eventType} ${serviceFilter} TIMESERIES`,
          confidence: indicator.confidence,
          reason: `Error trend for ${indicator.field}`,
        });
      });
    }

    if (context.includes('performance') || context.includes('latency')) {
      const performanceMetrics = metrics.filter(m => 
        m.field.includes('duration') || 
        m.field.includes('response') || 
        m.field.includes('latency')
      );

      performanceMetrics.forEach(metric => {
        suggestions.push({
          title: `${metric.field} - Percentiles`,
          type: 'line',
          nrql: `SELECT percentile(${metric.field}, 50, 95, 99) FROM ${metric.eventType} ${serviceFilter} TIMESERIES`,
          confidence: metric.confidence,
          reason: `Performance metric discovered: ${metric.field}`,
        });
      });
    }

    if (context.includes('throughput') || context.includes('volume')) {
      const primarySchema = worldModel.schemas[0];
      if (primarySchema) {
        suggestions.push({
          title: 'Request Throughput',
          type: 'line',
          nrql: `SELECT count(*) as 'Requests' FROM ${primarySchema.eventType} ${serviceFilter} TIMESERIES`,
          confidence: primarySchema.confidence || 0.8,
          reason: `Primary event type: ${primarySchema.eventType}`,
        });
      }
    }

    // Add generic suggestions if we don't have enough
    if (suggestions.length < input.maxCharts) {
      suggestions.push({
        title: 'Service Breakdown',
        type: 'pie',
        nrql: `SELECT count(*) FROM Transaction FACET ${serviceField} SINCE 1 hour ago LIMIT 10`,
        confidence: worldModel.serviceIdentifier.confidence,
        reason: `Service identifier: ${serviceField}`,
      });

      if (worldModel.schemas.length > 1) {
        suggestions.push({
          title: 'Event Type Distribution',
          type: 'bar',
          nrql: `SELECT count(*) FROM ${worldModel.schemas.slice(0, 3).map(s => s.eventType).join(', ')} FACET eventType() SINCE 1 hour ago`,
          confidence: 0.8,
          reason: 'Multiple event types discovered',
        });
      }
    }

    return {
      accountId,
      context: input.context,
      serviceName: input.serviceName,
      totalSuggestions: suggestions.length,
      suggestions: suggestions.slice(0, input.maxCharts),
      worldModelConfidence: worldModel.confidence,
    };
  },
};

// ============================================================================
// Export Tools
// ============================================================================

export function createDashboardTools(): ToolDefinition[] {
  return [
    createDashboard,
    listDashboards,
    getDashboard,
    suggestCharts,
  ];
}
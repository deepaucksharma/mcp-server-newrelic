/**
 * Alert Tools - Intelligent Alert Management
 * 
 * These tools provide alert creation and management capabilities
 * using discovered error indicators and metrics.
 */

import { z } from 'zod';
import { ToolDefinition, RequestContext } from '../core/types.js';
import { DiscoveryEngine } from '../core/discovery/engine.js';

// ============================================================================
// Input/Output Schemas
// ============================================================================

const CreateAlertInputSchema = z.object({
  name: z.string(),
  type: z.enum(['error_rate', 'response_time', 'throughput', 'custom']),
  serviceName: z.string().optional(),
  threshold: z.number().optional(),
  severity: z.enum(['low', 'medium', 'high', 'critical']).default('medium'),
  accountId: z.number().optional(),
});

const ListAlertsInputSchema = z.object({
  accountId: z.number().optional(),
  status: z.enum(['open', 'closed', 'all']).default('all'),
  limit: z.number().min(1).max(100).default(20),
});

const SuggestAlertsInputSchema = z.object({
  serviceName: z.string().optional(),
  priority: z.enum(['basic', 'comprehensive']).default('basic'),
  accountId: z.number().optional(),
});

const AlertStatusInputSchema = z.object({
  alertId: z.string().optional(),
  serviceName: z.string().optional(),
  timeRange: z.string().default('1 hour ago'),
  accountId: z.number().optional(),
});

// ============================================================================
// Tool Implementations
// ============================================================================

/**
 * Create intelligent alerts based on discovered patterns
 */
const createAlert: ToolDefinition<
  z.infer<typeof CreateAlertInputSchema>,
  any
> = {
  name: 'alerts.create',
  description: 'Create intelligent alerts using discovered error indicators and metrics',
  requiresDiscovery: true,
  inputSchema: CreateAlertInputSchema,
  
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
      description: `Creating ${input.type} alert using discovered patterns`,
      confidence: worldModel.confidence,
    });

    const serviceField = worldModel.serviceIdentifier.field;
    const errorIndicators = worldModel.errorIndicators;
    const metrics = worldModel.metrics;

    // Build service filter
    const serviceFilter = input.serviceName ? `${serviceField} = '${input.serviceName}'` : '';

    let alertConfig: any = {
      name: input.name,
      type: input.type,
      severity: input.severity,
      accountId,
      enabled: true,
      description: `Auto-generated alert using discovered data patterns (${(worldModel.confidence * 100).toFixed(1)}% confidence)`,
    };

    switch (input.type) {
      case 'error_rate':
        if (errorIndicators.length === 0) {
          return {
            error: 'Cannot create error rate alert - no error indicators discovered',
            suggestion: 'Implement standardized error tracking or use custom alert type',
            discoveredPatterns: {
              errorIndicators: errorIndicators.length,
              confidence: worldModel.confidence,
            },
          };
        }

        const errorIndicator = errorIndicators[0];
        const defaultErrorThreshold = input.threshold || 5; // 5% default

        alertConfig.condition = {
          type: 'NRQL',
          name: `${input.name} - Error Rate`,
          query: `SELECT percentage(count(*), WHERE ${errorIndicator.condition}) as 'errorRate' FROM ${errorIndicator.eventType} ${serviceFilter ? `WHERE ${serviceFilter}` : ''}`,
          threshold: defaultErrorThreshold,
          operator: 'above',
          duration: 300, // 5 minutes
          aggregationWindow: 60, // 1 minute
        };

        alertConfig.discoveryInfo = {
          errorIndicator: errorIndicator.field,
          confidence: errorIndicator.confidence,
          eventType: errorIndicator.eventType,
        };
        break;

      case 'response_time':
        const durationMetrics = metrics.filter(m => 
          m.field.includes('duration') || 
          m.field.includes('response') || 
          m.field.includes('latency')
        );

        if (durationMetrics.length === 0) {
          return {
            error: 'Cannot create response time alert - no latency metrics discovered',
            suggestion: 'Ensure APM instrumentation includes duration tracking or use custom alert type',
            discoveredPatterns: {
              metrics: metrics.length,
              confidence: worldModel.confidence,
            },
          };
        }

        const durationMetric = durationMetrics[0];
        const defaultLatencyThreshold = input.threshold || 1000; // 1 second default

        alertConfig.condition = {
          type: 'NRQL',
          name: `${input.name} - Response Time`,
          query: `SELECT average(${durationMetric.field}) as 'responseTime' FROM ${durationMetric.eventType} ${serviceFilter ? `WHERE ${serviceFilter}` : ''}`,
          threshold: defaultLatencyThreshold,
          operator: 'above',
          duration: 300, // 5 minutes
          aggregationWindow: 60, // 1 minute
        };

        alertConfig.discoveryInfo = {
          metric: durationMetric.field,
          confidence: durationMetric.confidence,
          eventType: durationMetric.eventType,
        };
        break;

      case 'throughput':
        const primarySchema = worldModel.schemas[0];
        const defaultThroughputThreshold = input.threshold || 100; // 100 requests

        alertConfig.condition = {
          type: 'NRQL',
          name: `${input.name} - Low Throughput`,
          query: `SELECT count(*) as 'throughput' FROM ${primarySchema?.eventType || 'Transaction'} ${serviceFilter ? `WHERE ${serviceFilter}` : ''}`,
          threshold: defaultThroughputThreshold,
          operator: 'below',
          duration: 600, // 10 minutes
          aggregationWindow: 60, // 1 minute
        };

        alertConfig.discoveryInfo = {
          eventType: primarySchema?.eventType || 'Transaction',
          confidence: primarySchema?.confidence || 0.5,
        };
        break;

      case 'custom':
        alertConfig.condition = {
          type: 'NRQL',
          name: `${input.name} - Custom Alert`,
          query: `SELECT count(*) FROM Transaction ${serviceFilter ? `WHERE ${serviceFilter}` : ''}`,
          threshold: input.threshold || 0,
          operator: 'above',
          duration: 300,
          aggregationWindow: 60,
        };

        alertConfig.note = 'Custom alert template - modify query and threshold as needed';
        break;
    }

    // Simulate alert creation (in real implementation, this would call New Relic Alerts API)
    const alertId = `alert_${Date.now()}`;

    const result = {
      accountId,
      alertId,
      name: input.name,
      type: input.type,
      severity: input.severity,
      config: alertConfig,
      nextSteps: [
        'Alert configuration generated based on discovered patterns',
        'Use New Relic Alerts API or UI to create the actual alert',
        'Configure notification channels (email, Slack, PagerDuty, etc.)',
        'Test the alert condition with historical data',
        'Adjust thresholds based on baseline performance',
      ],
      discoveryInfo: {
        worldModelConfidence: worldModel.confidence,
        serviceIdentifier: serviceField,
        errorIndicatorsAvailable: errorIndicators.length,
        metricsAvailable: metrics.length,
      },
    };

    ctx.explainabilityTrace.addStep({
      type: 'workflow_step',
      description: `Generated ${input.type} alert configuration`,
      confidence: worldModel.confidence,
    });

    return result;
  },
};

/**
 * List existing alerts and violations
 */
const listAlerts: ToolDefinition<
  z.infer<typeof ListAlertsInputSchema>,
  any
> = {
  name: 'alerts.list',
  description: 'List existing alerts and their current status',
  requiresDiscovery: false,
  inputSchema: ListAlertsInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    try {
      // Query for alert violations using NRQL
      const violationsQuery = `
        SELECT count(*) as violationCount, latest(conditionName) as condition
        FROM NrAiIncident 
        WHERE accountId = ${accountId}
        ${input.status !== 'all' ? `AND state = '${input.status.toUpperCase()}'` : ''}
        SINCE 24 hours ago 
        FACET conditionName, state
        LIMIT ${input.limit}
      `;

      const result = await ctx.nerdgraph.nrql(accountId, violationsQuery);

      const alerts = result.facets?.map((facet: any) => ({
        conditionName: facet.name.split(',')[0]?.trim(),
        state: facet.name.split(',')[1]?.trim(),
        violationCount: facet.results[0]?.violationCount || 0,
        lastUpdated: facet.results[0]?.condition || 'Unknown',
      })) || [];

      ctx.explainabilityTrace.addStep({
        type: 'query',
        description: `Retrieved ${alerts.length} alert conditions`,
        query: violationsQuery,
        resultCount: alerts.length,
        confidence: 0.9,
      });

      return {
        accountId,
        status: input.status,
        totalAlerts: alerts.length,
        alerts: alerts.slice(0, input.limit),
        summary: {
          open: alerts.filter(a => a.state === 'OPEN').length,
          closed: alerts.filter(a => a.state === 'CLOSED').length,
          total: alerts.length,
        },
      };

    } catch (error: any) {
      ctx.logger.error('Failed to list alerts', error);
      throw new Error(`Failed to list alerts: ${error.message}`);
    }
  },
};

/**
 * Suggest intelligent alert configurations
 */
const suggestAlerts: ToolDefinition<
  z.infer<typeof SuggestAlertsInputSchema>,
  any
> = {
  name: 'alerts.suggest',
  description: 'Suggest intelligent alert configurations based on discovered patterns',
  requiresDiscovery: true,
  inputSchema: SuggestAlertsInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    // Build or use existing world model
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    const serviceField = worldModel.serviceIdentifier.field;
    const errorIndicators = worldModel.errorIndicators;
    const metrics = worldModel.metrics;

    const suggestions = [];

    // Basic essential alerts
    if (errorIndicators.length > 0) {
      const errorIndicator = errorIndicators[0];
      suggestions.push({
        name: `High Error Rate${input.serviceName ? ` - ${input.serviceName}` : ''}`,
        type: 'error_rate',
        priority: 'high',
        description: `Alert when error rate exceeds 5% using discovered indicator: ${errorIndicator.field}`,
        estimatedThreshold: 5,
        confidence: errorIndicator.confidence,
        reasoning: `Based on discovered error indicator: ${errorIndicator.field} with ${(errorIndicator.confidence * 100).toFixed(1)}% confidence`,
      });
    }

    const durationMetrics = metrics.filter(m => 
      m.field.includes('duration') || 
      m.field.includes('response') || 
      m.field.includes('latency')
    );

    if (durationMetrics.length > 0) {
      const durationMetric = durationMetrics[0];
      suggestions.push({
        name: `Slow Response Time${input.serviceName ? ` - ${input.serviceName}` : ''}`,
        type: 'response_time',
        priority: 'medium',
        description: `Alert when response time exceeds 1 second using discovered metric: ${durationMetric.field}`,
        estimatedThreshold: 1000,
        confidence: durationMetric.confidence,
        reasoning: `Based on discovered latency metric: ${durationMetric.field}`,
      });
    }

    // Throughput alerts
    const primarySchema = worldModel.schemas[0];
    if (primarySchema) {
      suggestions.push({
        name: `Low Throughput${input.serviceName ? ` - ${input.serviceName}` : ''}`,
        type: 'throughput',
        priority: 'medium',
        description: `Alert when request volume drops significantly`,
        estimatedThreshold: 100,
        confidence: primarySchema.confidence || 0.7,
        reasoning: `Based on primary event type: ${primarySchema.eventType}`,
      });
    }

    // Comprehensive alerts for advanced monitoring
    if (input.priority === 'comprehensive') {
      // Add more sophisticated alerts
      if (errorIndicators.length > 1) {
        errorIndicators.slice(1, 3).forEach(indicator => {
          suggestions.push({
            name: `Secondary Error Pattern - ${indicator.field}`,
            type: 'custom',
            priority: 'low',
            description: `Monitor secondary error pattern: ${indicator.field}`,
            confidence: indicator.confidence,
            reasoning: `Secondary error indicator with ${(indicator.confidence * 100).toFixed(1)}% confidence`,
          });
        });
      }

      // System-level metrics
      const systemMetrics = metrics.filter(m => 
        m.field.includes('cpu') || 
        m.field.includes('memory') || 
        m.field.includes('disk')
      );

      systemMetrics.forEach(metric => {
        suggestions.push({
          name: `High ${metric.field}`,
          type: 'custom',
          priority: 'medium',
          description: `Alert on high ${metric.field} usage`,
          confidence: metric.confidence,
          reasoning: `System metric discovered: ${metric.field}`,
        });
      });
    }

    const result = {
      accountId,
      serviceName: input.serviceName,
      priority: input.priority,
      totalSuggestions: suggestions.length,
      suggestions,
      discoveryInfo: {
        worldModelConfidence: worldModel.confidence,
        serviceIdentifier: serviceField,
        errorIndicators: errorIndicators.length,
        metrics: metrics.length,
      },
      implementation: {
        recommended: suggestions.filter(s => s.priority === 'high').length,
        optional: suggestions.filter(s => ['medium', 'low'].includes(s.priority)).length,
        nextSteps: [
          'Start with high-priority alerts for immediate coverage',
          'Customize thresholds based on baseline performance',
          'Set up notification channels before enabling alerts',
          'Monitor alert noise and adjust sensitivity as needed',
        ],
      },
    };

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: `Suggested ${suggestions.length} alert configurations`,
      resultCount: suggestions.length,
      confidence: worldModel.confidence,
    });

    return result;
  },
};

/**
 * Get alert status and recent violations
 */
const getAlertStatus: ToolDefinition<
  z.infer<typeof AlertStatusInputSchema>,
  any
> = {
  name: 'alerts.status',
  description: 'Get current alert status and recent violation patterns',
  requiresDiscovery: false,
  inputSchema: AlertStatusInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    try {
      // Query for recent violations
      let violationsQuery = `
        SELECT count(*) as violations, latest(timestamp) as lastViolation
        FROM NrAiIncident 
        WHERE accountId = ${accountId}
        AND timestamp >= ${input.timeRange}
      `;

      if (input.alertId) {
        violationsQuery += ` AND conditionId = '${input.alertId}'`;
      }

      if (input.serviceName) {
        violationsQuery += ` AND entityName LIKE '%${input.serviceName}%'`;
      }

      violationsQuery += ` FACET conditionName, state LIMIT 50`;

      const result = await ctx.nerdgraph.nrql(accountId, violationsQuery);

      const violations = result.facets?.map((facet: any) => ({
        conditionName: facet.name.split(',')[0]?.trim(),
        state: facet.name.split(',')[1]?.trim(),
        violationCount: facet.results[0]?.violations || 0,
        lastViolation: new Date(facet.results[0]?.lastViolation || 0).toISOString(),
      })) || [];

      // Get summary statistics
      const summary = {
        totalViolations: violations.reduce((sum, v) => sum + v.violationCount, 0),
        openViolations: violations.filter(v => v.state === 'OPEN').length,
        closedViolations: violations.filter(v => v.state === 'CLOSED').length,
        uniqueConditions: new Set(violations.map(v => v.conditionName)).size,
      };

      ctx.explainabilityTrace.addStep({
        type: 'query',
        description: `Retrieved alert status for ${violations.length} conditions`,
        query: violationsQuery,
        resultCount: violations.length,
        confidence: 0.9,
      });

      return {
        accountId,
        timeRange: input.timeRange,
        alertId: input.alertId,
        serviceName: input.serviceName,
        summary,
        violations,
        healthStatus: summary.openViolations === 0 ? 'healthy' : 
                     summary.openViolations <= 2 ? 'warning' : 'critical',
      };

    } catch (error: any) {
      ctx.logger.error('Failed to get alert status', error);
      throw new Error(`Failed to get alert status: ${error.message}`);
    }
  },
};

// ============================================================================
// Export Tools
// ============================================================================

export function createAlertTools(): ToolDefinition[] {
  return [
    createAlert,
    listAlerts,
    suggestAlerts,
    getAlertStatus,
  ];
}
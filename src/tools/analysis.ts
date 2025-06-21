/**
 * Analysis Tools - Intelligent Insights from Discovery Data
 * 
 * These tools provide intelligent analysis capabilities built on top
 * of the discovery engine, offering actionable insights.
 */

import { z } from 'zod';
import { ToolDefinition, RequestContext } from '../core/types.js';
import { DiscoveryEngine } from '../core/discovery/engine.js';

// ============================================================================
// Input/Output Schemas
// ============================================================================

const ServiceHealthInputSchema = z.object({
  serviceName: z.string().optional(),
  timeRange: z.string().default('1 hour ago'),
  accountId: z.number().optional(),
});

const ErrorAnalysisInputSchema = z.object({
  serviceName: z.string().optional(),
  timeRange: z.string().default('1 hour ago'),
  accountId: z.number().optional(),
});

const PerformanceAnalysisInputSchema = z.object({
  serviceName: z.string().optional(),
  timeRange: z.string().default('1 hour ago'),
  percentile: z.number().min(50).max(99).default(95),
  accountId: z.number().optional(),
});

const AnomalyDetectionInputSchema = z.object({
  eventType: z.string(),
  metric: z.string(),
  timeRange: z.string().default('4 hours ago'),
  sensitivity: z.enum(['low', 'medium', 'high']).default('medium'),
  accountId: z.number().optional(),
});

// ============================================================================
// Tool Implementations
// ============================================================================

/**
 * Analyze service health using discovered indicators
 */
const analyzeServiceHealth: ToolDefinition<
  z.infer<typeof ServiceHealthInputSchema>,
  any
> = {
  name: 'analysis.service_health',
  description: 'Analyze service health using discovered error indicators and metrics',
  requiresDiscovery: true,
  inputSchema: ServiceHealthInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    ctx.explainabilityTrace.addStep({
      type: 'discovery',
      description: 'Starting service health analysis using discovered world model',
      confidence: 1.0,
    });

    // Build or use existing world model
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    const serviceField = worldModel.serviceIdentifier.field;
    const serviceEventType = worldModel.serviceIdentifier.eventType || 'Transaction';
    const errorIndicators = worldModel.errorIndicators;

    // Build service filter
    const serviceFilter = input.serviceName 
      ? `WHERE ${serviceField} = '${input.serviceName}'`
      : '';

    const healthMetrics = {
      throughput: 0,
      errorRate: 0,
      avgResponseTime: 0,
      availability: 0,
      confidence: worldModel.confidence,
    };

    try {
      // Get throughput
      const throughputQuery = `
        SELECT count(*) as throughput 
        FROM ${serviceEventType} 
        ${serviceFilter}
        SINCE ${input.timeRange}
      `;
      
      const throughputResult = await ctx.nerdgraph.nrql(accountId, throughputQuery);
      healthMetrics.throughput = throughputResult.results[0]?.throughput || 0;

      ctx.explainabilityTrace.addStep({
        type: 'query',
        description: 'Calculated throughput using discovered service identifier',
        query: throughputQuery,
        resultCount: 1,
        confidence: worldModel.serviceIdentifier.confidence,
      });

      // Calculate error rate using discovered error indicators
      if (errorIndicators.length > 0) {
        const primaryErrorIndicator = errorIndicators[0];
        const errorQuery = `
          SELECT percentage(count(*), WHERE ${primaryErrorIndicator.condition}) as errorRate
          FROM ${primaryErrorIndicator.eventType}
          ${serviceFilter}
          SINCE ${input.timeRange}
        `;

        const errorResult = await ctx.nerdgraph.nrql(accountId, errorQuery);
        healthMetrics.errorRate = errorResult.results[0]?.errorRate || 0;

        ctx.explainabilityTrace.addStep({
          type: 'query',
          description: `Calculated error rate using discovered indicator: ${primaryErrorIndicator.field}`,
          query: errorQuery,
          resultCount: 1,
          confidence: primaryErrorIndicator.confidence,
        });
      }

      // Get response time if duration field exists
      const durationMetrics = worldModel.metrics.filter(m => 
        m.field.includes('duration') && m.eventType === serviceEventType
      );

      if (durationMetrics.length > 0) {
        const durationField = durationMetrics[0].field;
        const responseTimeQuery = `
          SELECT average(${durationField}) as avgResponseTime
          FROM ${serviceEventType}
          ${serviceFilter}
          SINCE ${input.timeRange}
        `;

        const responseTimeResult = await ctx.nerdgraph.nrql(accountId, responseTimeQuery);
        healthMetrics.avgResponseTime = responseTimeResult.results[0]?.avgResponseTime || 0;

        ctx.explainabilityTrace.addStep({
          type: 'query',
          description: `Calculated response time using discovered metric: ${durationField}`,
          query: responseTimeQuery,
          resultCount: 1,
          confidence: durationMetrics[0].confidence,
        });
      }

      // Calculate availability (simple: 100% - error rate)
      healthMetrics.availability = Math.max(0, 100 - healthMetrics.errorRate);

    } catch (error: any) {
      ctx.logger.error('Service health analysis failed', error);
      throw new Error(`Health analysis failed: ${error.message}`);
    }

    // Determine overall health status
    let healthStatus = 'healthy';
    let healthScore = 100;

    if (healthMetrics.errorRate > 5) {
      healthStatus = 'critical';
      healthScore -= 50;
    } else if (healthMetrics.errorRate > 1) {
      healthStatus = 'warning';
      healthScore -= 20;
    }

    if (healthMetrics.avgResponseTime > 1000) {
      healthStatus = healthStatus === 'healthy' ? 'warning' : 'critical';
      healthScore -= 20;
    }

    const result = {
      serviceName: input.serviceName || 'All Services',
      timeRange: input.timeRange,
      accountId,
      healthStatus,
      healthScore,
      metrics: healthMetrics,
      discoveryInfo: {
        serviceIdentifier: `${serviceField} (${(worldModel.serviceIdentifier.confidence * 100).toFixed(1)}% confidence)`,
        errorIndicator: errorIndicators.length > 0 ? 
          `${errorIndicators[0].field} (${(errorIndicators[0].confidence * 100).toFixed(1)}% confidence)` :
          'No reliable error indicator found',
        metricsAvailable: worldModel.metrics.length,
        worldModelConfidence: worldModel.confidence,
      },
      recommendations: [
        healthStatus === 'critical' ? 'Immediate attention required - high error rate or response time' :
        healthStatus === 'warning' ? 'Monitor closely - some degradation detected' :
        'Service appears healthy based on discovered metrics',
        
        errorIndicators.length === 0 ? 'Consider implementing standardized error tracking' : null,
        healthMetrics.avgResponseTime === 0 ? 'No response time metrics found - consider adding duration tracking' : null,
      ].filter(Boolean),
    };

    ctx.explainabilityTrace.addStep({
      type: 'analysis',
      description: `Health analysis complete: ${healthStatus} (${healthScore}/100)`,
      confidence: worldModel.confidence,
    });

    return result;
  },
};

/**
 * Analyze error patterns and trends
 */
const analyzeErrors: ToolDefinition<
  z.infer<typeof ErrorAnalysisInputSchema>,
  any
> = {
  name: 'analysis.error_patterns',
  description: 'Analyze error patterns using discovered error indicators',
  requiresDiscovery: true,
  inputSchema: ErrorAnalysisInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    // Build or use existing world model
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    const errorIndicators = worldModel.errorIndicators;
    if (errorIndicators.length === 0) {
      return {
        accountId,
        serviceName: input.serviceName,
        message: 'No error indicators discovered in the account',
        suggestion: 'Check for custom error fields or implement standardized error tracking',
      };
    }

    const serviceField = worldModel.serviceIdentifier.field;
    const serviceFilter = input.serviceName 
      ? `AND ${serviceField} = '${input.serviceName}'`
      : '';

    const errorAnalysis = [];

    for (const indicator of errorIndicators) {
      try {
        const errorTrendQuery = `
          SELECT percentage(count(*), WHERE ${indicator.condition}) as errorRate
          FROM ${indicator.eventType}
          WHERE timestamp >= ${input.timeRange} ${serviceFilter}
          TIMESERIES 10 minutes
        `;

        const trendResult = await ctx.nerdgraph.nrql(accountId, errorTrendQuery);

        const errorBreakdownQuery = `
          SELECT count(*) as errorCount
          FROM ${indicator.eventType}
          WHERE ${indicator.condition} 
          AND timestamp >= ${input.timeRange} ${serviceFilter}
          FACET ${serviceField}
          LIMIT 10
        `;

        const breakdownResult = await ctx.nerdgraph.nrql(accountId, errorBreakdownQuery);

        errorAnalysis.push({
          indicator: indicator.field,
          eventType: indicator.eventType,
          condition: indicator.condition,
          confidence: indicator.confidence,
          trend: trendResult.results || [],
          breakdown: breakdownResult.facets || [],
        });

        ctx.explainabilityTrace.addStep({
          type: 'query',
          description: `Analyzed error pattern for ${indicator.field}`,
          query: errorTrendQuery,
          confidence: indicator.confidence,
        });

      } catch (error: any) {
        ctx.logger.warn(`Error analysis failed for ${indicator.field}`, error);
      }
    }

    return {
      accountId,
      serviceName: input.serviceName || 'All Services',
      timeRange: input.timeRange,
      totalIndicators: errorIndicators.length,
      analysis: errorAnalysis,
      worldModelConfidence: worldModel.confidence,
    };
  },
};

/**
 * Analyze performance metrics
 */
const analyzePerformance: ToolDefinition<
  z.infer<typeof PerformanceAnalysisInputSchema>,
  any
> = {
  name: 'analysis.performance',
  description: 'Analyze performance using discovered metrics',
  requiresDiscovery: true,
  inputSchema: PerformanceAnalysisInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    // Build or use existing world model
    let worldModel = ctx.worldModel;
    if (!worldModel) {
      const discoveryEngine = new DiscoveryEngine(ctx);
      worldModel = await discoveryEngine.buildDiscoveryGraph(accountId);
    }

    const performanceMetrics = worldModel.metrics.filter(m => 
      m.field.includes('duration') || 
      m.field.includes('response') || 
      m.field.includes('latency')
    );

    if (performanceMetrics.length === 0) {
      return {
        accountId,
        message: 'No performance metrics discovered',
        suggestion: 'Check for custom duration fields or ensure APM instrumentation is properly configured',
      };
    }

    const serviceField = worldModel.serviceIdentifier.field;
    const serviceFilter = input.serviceName 
      ? `AND ${serviceField} = '${input.serviceName}'`
      : '';

    const analysis = [];

    for (const metric of performanceMetrics) {
      try {
        const perfQuery = `
          SELECT 
            average(${metric.field}) as avg,
            percentile(${metric.field}, ${input.percentile}) as p${input.percentile},
            max(${metric.field}) as max,
            count(*) as samples
          FROM ${metric.eventType}
          WHERE timestamp >= ${input.timeRange} ${serviceFilter}
          TIMESERIES 10 minutes
        `;

        const result = await ctx.nerdgraph.nrql(accountId, perfQuery);

        analysis.push({
          metric: metric.field,
          eventType: metric.eventType,
          confidence: metric.confidence,
          trend: result.results || [],
        });

        ctx.explainabilityTrace.addStep({
          type: 'query',
          description: `Analyzed performance metric: ${metric.field}`,
          query: perfQuery,
          confidence: metric.confidence,
        });

      } catch (error: any) {
        ctx.logger.warn(`Performance analysis failed for ${metric.field}`, error);
      }
    }

    return {
      accountId,
      serviceName: input.serviceName || 'All Services',
      timeRange: input.timeRange,
      percentile: input.percentile,
      metricsAnalyzed: analysis.length,
      analysis,
      worldModelConfidence: worldModel.confidence,
    };
  },
};

/**
 * Simple anomaly detection using statistical methods
 */
const detectAnomalies: ToolDefinition<
  z.infer<typeof AnomalyDetectionInputSchema>,
  any
> = {
  name: 'analysis.detect_anomalies',
  description: 'Detect anomalies in metrics using statistical analysis',
  requiresDiscovery: false,
  inputSchema: AnomalyDetectionInputSchema,
  
  async handler(ctx: RequestContext, input) {
    const accountId = input.accountId || ctx.accountId;

    // Get baseline data for comparison
    const baselineQuery = `
      SELECT average(${input.metric}) as avg, stddev(${input.metric}) as stddev
      FROM ${input.eventType}
      WHERE timestamp >= ${input.timeRange}
    `;

    const recentQuery = `
      SELECT average(${input.metric}) as recent_avg
      FROM ${input.eventType}
      WHERE timestamp >= 30 minutes ago
    `;

    try {
      const [baselineResult, recentResult] = await Promise.all([
        ctx.nerdgraph.nrql(accountId, baselineQuery),
        ctx.nerdgraph.nrql(accountId, recentQuery),
      ]);

      const baseline = baselineResult.results[0] || {};
      const recent = recentResult.results[0] || {};

      const avg = baseline.avg || 0;
      const stddev = baseline.stddev || 0;
      const recentAvg = recent.recent_avg || 0;

      // Simple z-score calculation
      const zScore = stddev > 0 ? Math.abs(recentAvg - avg) / stddev : 0;
      
      const sensitivityThresholds = {
        low: 3.0,
        medium: 2.0,
        high: 1.5,
      };

      const threshold = sensitivityThresholds[input.sensitivity];
      const isAnomalous = zScore > threshold;

      const result = {
        accountId,
        eventType: input.eventType,
        metric: input.metric,
        timeRange: input.timeRange,
        analysis: {
          baseline: {
            average: avg,
            standardDeviation: stddev,
          },
          recent: {
            average: recentAvg,
          },
          anomalyScore: zScore,
          threshold,
          isAnomalous,
          severity: zScore > 3 ? 'high' : zScore > 2 ? 'medium' : 'low',
        },
        explanation: isAnomalous 
          ? `Anomaly detected: Recent average (${recentAvg.toFixed(2)}) deviates ${zScore.toFixed(2)} standard deviations from baseline`
          : `No anomaly detected: Recent values within ${zScore.toFixed(2)} standard deviations of baseline`,
      };

      ctx.explainabilityTrace.addStep({
        type: 'analysis',
        description: `Anomaly detection: ${isAnomalous ? 'ANOMALY' : 'NORMAL'} (z-score: ${zScore.toFixed(2)})`,
        confidence: 0.8,
      });

      return result;

    } catch (error: any) {
      throw new Error(`Anomaly detection failed: ${error.message}`);
    }
  },
};

// ============================================================================
// Export Tools
// ============================================================================

export function createAnalysisTools(): ToolDefinition[] {
  return [
    analyzeServiceHealth,
    analyzeErrors,
    analyzePerformance,
    detectAnomalies,
  ];
}
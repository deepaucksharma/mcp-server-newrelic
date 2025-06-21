/**
 * Entity Comparison Tool - compare.similar_entities
 * 
 * Analyzes and compares similar entities to identify performance patterns,
 * outliers, and optimization opportunities across services.
 */

import { Tool } from '@modelcontextprotocol/sdk/types.js';
import { NerdGraphClient, Logger } from '../core/types.js';
import { GoldenSignalsEngine, EntityGoldenMetrics, TelemetryContext } from '../core/golden-signals.js';

export interface EntityComparison {
  entity: {
    guid: string;
    name: string;
    type: string;
    environment?: string;
    language?: string;
  };
  metrics: {
    latency: { p95: number; trend: 'better' | 'worse' | 'similar' };
    traffic: { rate: number; trend: 'better' | 'worse' | 'similar' };
    errors: { percentage: number; trend: 'better' | 'worse' | 'similar' };
    saturation: { available: boolean; cpuAvg?: number; memoryAvg?: number };
  };
  rank: {
    overall: number; // 1-5 scale (1 = best, 5 = worst)
    latency: number;
    reliability: number;
    throughput: number;
  };
  insights: string[];
}

export interface ComparisonAnalysis {
  summary: {
    totalEntities: number;
    comparisonBaseline: string;
    analysisTimeframe: string;
    primaryFindings: string[];
  };
  entities: EntityComparison[];
  patterns: {
    performanceOutliers: string[];
    commonIssues: string[];
    bestPractices: string[];
    recommendations: string[];
  };
  benchmarks: {
    latencyP95: { best: number; worst: number; median: number };
    errorRate: { best: number; worst: number; median: number };
    throughput: { best: number; worst: number; median: number };
  };
}

export class EntityComparisonTool {
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
      name: 'compare.similar_entities',
      description: `Compare similar entities to identify performance patterns, outliers, and optimization opportunities.

      🎯 **Purpose**: Analyzes groups of similar entities to identify performance leaders, laggards, and improvement opportunities.

      **Comparison Capabilities**:
      - **Performance Benchmarking**: Compare latency, throughput, and error rates across similar services
      - **Outlier Detection**: Identify services performing significantly better or worse than peers
      - **Pattern Analysis**: Discover common performance patterns and anti-patterns
      - **Optimization Guidance**: Generate specific recommendations based on best-performing entities

      **Comparison Strategies**:
      - **By Type**: Compare all entities of the same type (e.g., all APPLICATION entities)
      - **By Pattern**: Compare entities matching name patterns (e.g., all "*-api" services)
      - **By Environment**: Compare across environments (production vs staging)
      - **By Technology**: Compare similar technology stacks (same language/framework)

      **Analysis Output**:
      - Ranked performance comparison with percentile analysis
      - Performance outlier identification and root cause hints
      - Best practice recommendations from top performers
      - Specific NRQL queries for deeper investigation
      - Actionable optimization suggestions

      **Use Cases**:
      - Service performance benchmarking and SLA planning
      - Identifying candidates for performance optimization
      - Validating deployment impacts across service fleets
      - Cross-team performance sharing and learning`,

      inputSchema: {
        type: 'object',
        properties: {
          comparison_strategy: {
            type: 'string',
            enum: ['by_type', 'by_name_pattern', 'by_environment', 'by_explicit_list'],
            description: 'Strategy for selecting entities to compare',
          },
          entity_type: {
            type: 'string',
            description: 'Entity type to compare (for by_type strategy)',
            examples: ['APPLICATION', 'SERVICE', 'HOST'],
          },
          name_pattern: {
            type: 'string',
            description: 'Name pattern to match (for by_name_pattern strategy)',
            examples: ['*-api', '*-service', 'prod-*'],
          },
          environment: {
            type: 'string',
            description: 'Environment tag to filter by (for by_environment strategy)',
            examples: ['production', 'staging', 'development'],
          },
          entity_guids: {
            type: 'array',
            items: { type: 'string' },
            description: 'Explicit list of entity GUIDs to compare (for by_explicit_list strategy)',
          },
          baseline_entity: {
            type: 'string',
            description: 'Optional specific entity GUID to use as comparison baseline',
          },
          timeframe_hours: {
            type: 'number',
            description: 'Time window for metrics analysis in hours',
            default: 1,
            minimum: 0.25,
            maximum: 168,
          },
          max_entities: {
            type: 'number',
            description: 'Maximum number of entities to include in comparison',
            default: 10,
            minimum: 2,
            maximum: 50,
          },
          include_saturation: {
            type: 'boolean',
            description: 'Include resource saturation metrics in comparison',
            default: true,
          },
          sort_by: {
            type: 'string',
            enum: ['overall_performance', 'latency', 'error_rate', 'throughput'],
            description: 'Primary metric to sort comparison results by',
            default: 'overall_performance',
          },
        },
        required: ['comparison_strategy'],
        additionalProperties: false,
      },
    };
  }

  /**
   * Handle the compare.similar_entities tool call
   */
  async handle(params: any): Promise<any> {
    const {
      comparison_strategy,
      entity_type,
      name_pattern,
      environment,
      entity_guids = [],
      baseline_entity,
      timeframe_hours = 1,
      max_entities = 10,
      include_saturation = true,
      sort_by = 'overall_performance',
    } = params;

    this.logger.info('Entity comparison requested', {
      comparison_strategy,
      entity_type,
      name_pattern,
      timeframe_hours,
      max_entities,
    });

    try {
      // 1. Discover entities based on strategy
      const entities = await this.discoverEntitiesForComparison(params);
      
      if (entities.length < 2) {
        return {
          content: [
            {
              type: 'text',
              text: `⚠️ **Insufficient Entities for Comparison**\n\nFound only ${entities.length} entities matching the criteria. Need at least 2 entities for meaningful comparison.\n\n**Suggestions:**\n- Broaden your search criteria\n- Use a different comparison strategy\n- Check if entities are currently reporting data\n\nTry using \`discover.environment\` to see all available entities first.`,
            },
          ],
          isError: false,
        };
      }

      // 2. Get golden metrics for all entities
      const entityMetrics = await this.getEntityMetricsForComparison(
        entities.slice(0, max_entities),
        timeframe_hours
      );

      // 3. Perform comparative analysis
      const analysis = await this.performComparisonAnalysis(
        entityMetrics,
        baseline_entity,
        sort_by,
        params
      );

      return this.formatResponse(analysis, params);

    } catch (error: any) {
      this.logger.error('Entity comparison failed', { 
        comparison_strategy, 
        error: error.message 
      });

      return {
        content: [
          {
            type: 'text',
            text: `❌ **Entity Comparison Failed**\n\nError: ${error.message}\n\nThis might indicate:\n- Invalid comparison parameters\n- No entities match the specified criteria\n- Insufficient permissions to access entity data\n- Network connectivity issues\n\nTry using \`discover.environment\` first to understand the available entities and data patterns.`,
          },
        ],
        isError: true,
      };
    }
  }

  /**
   * Discover entities based on comparison strategy
   */
  private async discoverEntitiesForComparison(params: any): Promise<any[]> {
    const { comparison_strategy } = params;

    switch (comparison_strategy) {
      case 'by_type':
        return this.discoverEntitiesByType(params.entity_type);
      
      case 'by_name_pattern':
        return this.discoverEntitiesByNamePattern(params.name_pattern);
      
      case 'by_environment':
        return this.discoverEntitiesByEnvironment(params.environment);
      
      case 'by_explicit_list':
        return this.getEntitiesByGuids(params.entity_guids);
      
      default:
        throw new Error(`Unknown comparison strategy: ${comparison_strategy}`);
    }
  }

  /**
   * Discover entities by type
   */
  private async discoverEntitiesByType(entityType: string): Promise<any[]> {
    const query = `
      {
        actor {
          entitySearch(
            query: "type = '${entityType}' AND reporting = true"
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
              }
            }
          }
        }
      }
    `;

    const result = await this.nerdgraph.request(query);
    return result.actor?.entitySearch?.results?.entities || [];
  }

  /**
   * Discover entities by name pattern
   */
  private async discoverEntitiesByNamePattern(namePattern: string): Promise<any[]> {
    // Convert wildcard pattern to GraphQL-compatible format
    const graphqlPattern = namePattern.replace(/\*/g, '%');
    
    const query = `
      {
        actor {
          entitySearch(
            query: "name LIKE '${graphqlPattern}' AND reporting = true"
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
              }
            }
          }
        }
      }
    `;

    const result = await this.nerdgraph.request(query);
    return result.actor?.entitySearch?.results?.entities || [];
  }

  /**
   * Discover entities by environment tag
   */
  private async discoverEntitiesByEnvironment(environment: string): Promise<any[]> {
    const query = `
      {
        actor {
          entitySearch(
            query: "tags.environment = '${environment}' AND reporting = true"
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
              }
            }
          }
        }
      }
    `;

    const result = await this.nerdgraph.request(query);
    return result.actor?.entitySearch?.results?.entities || [];
  }

  /**
   * Get specific entities by their GUIDs
   */
  private async getEntitiesByGuids(guids: string[]): Promise<any[]> {
    const entities = [];
    
    for (const guid of guids) {
      try {
        const query = `
          {
            actor {
              entity(guid: "${guid}") {
                guid
                name
                type
                domain
                entityType
                tags {
                  key
                  values
                }
              }
            }
          }
        `;
        
        const result = await this.nerdgraph.request(query);
        if (result.actor?.entity) {
          entities.push(result.actor.entity);
        }
      } catch (error: any) {
        this.logger.warn(`Could not fetch entity ${guid}`, { error: error.message });
      }
    }
    
    return entities;
  }

  /**
   * Get golden metrics for all entities in comparison
   */
  private async getEntityMetricsForComparison(
    entities: any[],
    timeframeHours: number
  ): Promise<EntityGoldenMetrics[]> {
    const metrics = [];
    const timeframeMinutes = Math.max(timeframeHours * 60, 30);

    for (const entity of entities) {
      try {
        const goldenMetrics = await this.goldenSignals.getEntityGoldenMetrics(
          entity.guid,
          timeframeMinutes
        );
        metrics.push(goldenMetrics);
      } catch (error: any) {
        this.logger.warn(`Could not get metrics for entity ${entity.name}`, { 
          guid: entity.guid, 
          error: error.message 
        });
      }
    }

    return metrics;
  }

  /**
   * Perform comprehensive comparison analysis
   */
  private async performComparisonAnalysis(
    entityMetrics: EntityGoldenMetrics[],
    baselineEntity?: string,
    sortBy: string = 'overall_performance',
    params: any = {}
  ): Promise<ComparisonAnalysis> {
    // Calculate benchmarks
    const benchmarks = this.calculateBenchmarks(entityMetrics);
    
    // Analyze each entity
    const entityComparisons = entityMetrics.map(em => 
      this.analyzeEntityComparison(em, benchmarks, baselineEntity)
    );

    // Sort by specified criteria
    entityComparisons.sort((a, b) => this.compareEntities(a, b, sortBy));

    // Identify patterns
    const patterns = this.identifyPatterns(entityComparisons, benchmarks);

    // Generate summary
    const summary = this.generateSummary(entityComparisons, patterns, params);

    return {
      summary,
      entities: entityComparisons,
      patterns,
      benchmarks,
    };
  }

  /**
   * Calculate performance benchmarks across all entities
   */
  private calculateBenchmarks(entityMetrics: EntityGoldenMetrics[]): any {
    const latencies = entityMetrics.map(em => em.metrics.latency.p95).filter(l => l > 0);
    const errorRates = entityMetrics.map(em => em.metrics.errors.percentage);
    const throughputs = entityMetrics.map(em => em.metrics.traffic.rate).filter(t => t > 0);

    return {
      latencyP95: {
        best: Math.min(...latencies),
        worst: Math.max(...latencies),
        median: this.calculateMedian(latencies),
      },
      errorRate: {
        best: Math.min(...errorRates),
        worst: Math.max(...errorRates),
        median: this.calculateMedian(errorRates),
      },
      throughput: {
        best: Math.max(...throughputs),
        worst: Math.min(...throughputs),
        median: this.calculateMedian(throughputs),
      },
    };
  }

  /**
   * Analyze individual entity comparison
   */
  private analyzeEntityComparison(
    entityMetrics: EntityGoldenMetrics,
    benchmarks: any,
    baselineEntity?: string
  ): EntityComparison {
    const { entity, metrics } = entityMetrics;
    
    // Calculate trends and ranks
    const latencyRank = this.calculateRank(metrics.latency.p95, benchmarks.latencyP95, 'lower_better');
    const errorRank = this.calculateRank(metrics.errors.percentage, benchmarks.errorRate, 'lower_better');
    const throughputRank = this.calculateRank(metrics.traffic.rate, benchmarks.throughput, 'higher_better');
    
    const overallRank = Math.round((latencyRank + errorRank + throughputRank) / 3);

    // Generate insights
    const insights = this.generateEntityInsights(metrics, benchmarks, overallRank);

    return {
      entity: {
        guid: entity.guid,
        name: entity.name,
        type: entity.type,
        // Extract from tags if available
        environment: this.extractTagValue(entityMetrics, 'environment'),
        language: this.extractTagValue(entityMetrics, 'language'),
      },
      metrics: {
        latency: {
          p95: metrics.latency.p95,
          trend: this.calculateTrend(metrics.latency.p95, benchmarks.latencyP95.median, 'lower_better'),
        },
        traffic: {
          rate: metrics.traffic.rate,
          trend: this.calculateTrend(metrics.traffic.rate, benchmarks.throughput.median, 'higher_better'),
        },
        errors: {
          percentage: metrics.errors.percentage,
          trend: this.calculateTrend(metrics.errors.percentage, benchmarks.errorRate.median, 'lower_better'),
        },
        saturation: {
          available: metrics.saturation.available,
          cpuAvg: metrics.saturation.cpu?.avg,
          memoryAvg: metrics.saturation.memory?.avg,
        },
      },
      rank: {
        overall: overallRank,
        latency: latencyRank,
        reliability: errorRank,
        throughput: throughputRank,
      },
      insights,
    };
  }

  /**
   * Identify patterns across entity comparisons
   */
  private identifyPatterns(comparisons: EntityComparison[], benchmarks: any): any {
    const performanceOutliers = [];
    const commonIssues = [];
    const bestPractices = [];
    const recommendations = [];

    // Identify outliers
    const topPerformers = comparisons.filter(c => c.rank.overall <= 2);
    const poorPerformers = comparisons.filter(c => c.rank.overall >= 4);

    if (topPerformers.length > 0) {
      performanceOutliers.push(`Top performers: ${topPerformers.map(p => p.entity.name).join(', ')}`);
    }
    if (poorPerformers.length > 0) {
      performanceOutliers.push(`Needs attention: ${poorPerformers.map(p => p.entity.name).join(', ')}`);
    }

    // Common issues
    const highLatencyCount = comparisons.filter(c => c.metrics.latency.p95 > benchmarks.latencyP95.median * 2).length;
    const highErrorCount = comparisons.filter(c => c.metrics.errors.percentage > 5).length;
    
    if (highLatencyCount > comparisons.length * 0.3) {
      commonIssues.push(`${highLatencyCount} services have high latency (>2x median)`);
    }
    if (highErrorCount > 0) {
      commonIssues.push(`${highErrorCount} services have elevated error rates (>5%)`);
    }

    // Best practices from top performers
    if (topPerformers.length > 0) {
      const avgTopLatency = topPerformers.reduce((sum, p) => sum + p.metrics.latency.p95, 0) / topPerformers.length;
      bestPractices.push(`Top performers average ${Math.round(avgTopLatency)}ms P95 latency`);
      
      const topLanguages = [...new Set(topPerformers.map(p => p.entity.language).filter(Boolean))];
      if (topLanguages.length > 0) {
        bestPractices.push(`Strong performers use: ${topLanguages.join(', ')}`);
      }
    }

    // Recommendations
    if (poorPerformers.length > 0) {
      recommendations.push(`Focus optimization efforts on: ${poorPerformers.slice(0, 3).map(p => p.entity.name).join(', ')}`);
    }
    if (highLatencyCount > 0) {
      recommendations.push('Investigate latency patterns - check database queries, external API calls, and resource constraints');
    }

    return {
      performanceOutliers,
      commonIssues,
      bestPractices,
      recommendations,
    };
  }

  /**
   * Generate comparison summary
   */
  private generateSummary(comparisons: EntityComparison[], patterns: any, params: any): any {
    const primaryFindings = [];
    
    if (patterns.performanceOutliers.length > 0) {
      primaryFindings.push(`Performance spread identified across ${comparisons.length} entities`);
    }
    if (patterns.commonIssues.length > 0) {
      primaryFindings.push(`${patterns.commonIssues.length} common performance issues detected`);
    }
    if (patterns.bestPractices.length > 0) {
      primaryFindings.push(`${patterns.bestPractices.length} best practices identified from top performers`);
    }

    return {
      totalEntities: comparisons.length,
      comparisonBaseline: params.baseline_entity ? 'Specific entity baseline' : 'Peer group median',
      analysisTimeframe: `${params.timeframe_hours || 1} hours`,
      primaryFindings,
    };
  }

  // Utility methods

  private calculateMedian(values: number[]): number {
    const sorted = [...values].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    return sorted.length % 2 === 0 
      ? (sorted[mid - 1] + sorted[mid]) / 2 
      : sorted[mid];
  }

  private calculateRank(value: number, benchmark: any, direction: 'higher_better' | 'lower_better'): number {
    if (direction === 'lower_better') {
      if (value <= benchmark.best * 1.1) return 1; // Within 10% of best
      if (value <= benchmark.median) return 2;
      if (value <= benchmark.median * 1.5) return 3;
      if (value <= benchmark.worst * 0.9) return 4;
      return 5;
    } else {
      if (value >= benchmark.best * 0.9) return 1; // Within 10% of best
      if (value >= benchmark.median) return 2;
      if (value >= benchmark.median * 0.5) return 3;
      if (value >= benchmark.worst * 1.1) return 4;
      return 5;
    }
  }

  private calculateTrend(value: number, baseline: number, direction: 'higher_better' | 'lower_better'): 'better' | 'worse' | 'similar' {
    const ratio = value / baseline;
    const threshold = 0.2; // 20% difference threshold

    if (direction === 'lower_better') {
      if (ratio < (1 - threshold)) return 'better';
      if (ratio > (1 + threshold)) return 'worse';
    } else {
      if (ratio > (1 + threshold)) return 'better';
      if (ratio < (1 - threshold)) return 'worse';
    }
    return 'similar';
  }

  private compareEntities(a: EntityComparison, b: EntityComparison, sortBy: string): number {
    switch (sortBy) {
      case 'latency':
        return a.metrics.latency.p95 - b.metrics.latency.p95;
      case 'error_rate':
        return a.metrics.errors.percentage - b.metrics.errors.percentage;
      case 'throughput':
        return b.metrics.traffic.rate - a.metrics.traffic.rate; // Higher is better
      default: // overall_performance
        return a.rank.overall - b.rank.overall;
    }
  }

  private extractTagValue(entityMetrics: EntityGoldenMetrics, tagKey: string): string | undefined {
    // This would extract from entity tags - placeholder implementation
    return undefined;
  }

  private generateEntityInsights(metrics: any, benchmarks: any, overallRank: number): string[] {
    const insights = [];

    if (overallRank === 1) {
      insights.push('🏆 Top performer - excellent across all golden signals');
    } else if (overallRank >= 4) {
      insights.push('⚠️ Needs attention - below peer performance in multiple areas');
    }

    if (metrics.latency.p95 > benchmarks.latencyP95.median * 2) {
      insights.push('🐌 High latency - investigate slow operations and bottlenecks');
    }

    if (metrics.errors.percentage > 5) {
      insights.push('🚨 High error rate - requires immediate attention');
    } else if (metrics.errors.percentage === 0) {
      insights.push('✅ Error-free performance');
    }

    if (metrics.traffic.rate < benchmarks.throughput.median * 0.5) {
      insights.push('📉 Low traffic volume - consider capacity planning');
    }

    return insights;
  }

  /**
   * Format response for LLM consumption
   */
  private formatResponse(analysis: ComparisonAnalysis, params: any): any {
    let content = `# 📊 Entity Performance Comparison Analysis\n\n`;

    // Summary
    content += `## 📋 Summary\n\n`;
    content += `- **Entities Analyzed**: ${analysis.summary.totalEntities}\n`;
    content += `- **Comparison Strategy**: ${params.comparison_strategy.replace('_', ' ')}\n`;
    content += `- **Time Window**: ${analysis.summary.analysisTimeframe}\n`;
    content += `- **Baseline**: ${analysis.summary.comparisonBaseline}\n\n`;

    if (analysis.summary.primaryFindings.length > 0) {
      content += `**Key Findings**:\n`;
      analysis.summary.primaryFindings.forEach(finding => {
        content += `- ${finding}\n`;
      });
      content += `\n`;
    }

    // Performance Rankings
    content += `## 🏆 Performance Rankings\n\n`;
    analysis.entities.forEach((entity, index) => {
      const rankIcon = this.getRankIcon(entity.rank.overall);
      const trendIcon = this.getTrendIcon(entity.metrics.latency.trend);
      
      content += `**${index + 1}. ${entity.entity.name}** ${rankIcon}\n`;
      content += `- **Latency**: ${entity.metrics.latency.p95}ms P95 ${trendIcon}\n`;
      content += `- **Error Rate**: ${entity.metrics.errors.percentage}% ${this.getTrendIcon(entity.metrics.errors.trend)}\n`;
      content += `- **Traffic**: ${entity.metrics.traffic.rate.toFixed(1)} rpm ${this.getTrendIcon(entity.metrics.traffic.trend)}\n`;
      
      if (entity.insights.length > 0) {
        content += `- **Insights**: ${entity.insights.join(', ')}\n`;
      }
      content += `\n`;
    });

    // Benchmarks
    content += `## 📈 Performance Benchmarks\n\n`;
    content += `**Latency (P95)**:\n`;
    content += `- Best: ${analysis.benchmarks.latencyP95.best}ms\n`;
    content += `- Median: ${analysis.benchmarks.latencyP95.median}ms\n`;
    content += `- Worst: ${analysis.benchmarks.latencyP95.worst}ms\n\n`;
    
    content += `**Error Rate**:\n`;
    content += `- Best: ${analysis.benchmarks.errorRate.best}%\n`;
    content += `- Median: ${analysis.benchmarks.errorRate.median}%\n`;
    content += `- Worst: ${analysis.benchmarks.errorRate.worst}%\n\n`;

    content += `**Throughput**:\n`;
    content += `- Best: ${analysis.benchmarks.throughput.best.toFixed(1)} rpm\n`;
    content += `- Median: ${analysis.benchmarks.throughput.median.toFixed(1)} rpm\n`;
    content += `- Worst: ${analysis.benchmarks.throughput.worst.toFixed(1)} rpm\n\n`;

    // Patterns and Insights
    if (analysis.patterns.performanceOutliers.length > 0) {
      content += `## 🎯 Performance Outliers\n\n`;
      analysis.patterns.performanceOutliers.forEach(outlier => {
        content += `- ${outlier}\n`;
      });
      content += `\n`;
    }

    if (analysis.patterns.commonIssues.length > 0) {
      content += `## ⚠️ Common Issues\n\n`;
      analysis.patterns.commonIssues.forEach(issue => {
        content += `- ${issue}\n`;
      });
      content += `\n`;
    }

    if (analysis.patterns.bestPractices.length > 0) {
      content += `## ✅ Best Practices Identified\n\n`;
      analysis.patterns.bestPractices.forEach(practice => {
        content += `- ${practice}\n`;
      });
      content += `\n`;
    }

    if (analysis.patterns.recommendations.length > 0) {
      content += `## 💡 Recommendations\n\n`;
      analysis.patterns.recommendations.forEach(rec => {
        content += `- ${rec}\n`;
      });
      content += `\n`;
    }

    content += `---\n*Analysis completed at ${new Date().toISOString()}*`;

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

  private getRankIcon(rank: number): string {
    const icons = ['', '🥇', '🥈', '🥉', '⚠️', '🚨'];
    return icons[rank] || '❓';
  }

  private getTrendIcon(trend: 'better' | 'worse' | 'similar'): string {
    switch (trend) {
      case 'better': return '✅';
      case 'worse': return '⚠️';
      case 'similar': return '➖';
      default: return '';
    }
  }
}
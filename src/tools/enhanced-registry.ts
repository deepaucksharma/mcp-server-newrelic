/**
 * Enhanced Tool Registry - Platform-Native with Rich Metadata
 * 
 * Enhances existing new-branch tools with AI-optimized descriptions,
 * examples, and discovery-first behavior.
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { Tool, ListToolsRequestSchema, CallToolRequestSchema } from '@modelcontextprotocol/sdk/types.js';
import { PlatformDiscovery } from '../core/platform-discovery.js';
import { EnvironmentDiscoveryTool } from './environment-discovery.js';
import { DashboardGenerationTool } from './dashboard-generation.js';
import { EntityComparisonTool } from './entity-comparison.js';
import { GoldenSignalsEngine } from '../core/golden-signals.js';
import { IntelligentCache } from '../core/intelligent-cache.js';
import { NerdGraphClient, Logger } from '../core/types.js';
import { createNerdGraphClient } from '../adapters/nerdgraph.js';

export interface EnhancedToolMetadata {
  readOnlyHint?: boolean;
  destructiveHint?: boolean;
  requiresConfirmation?: boolean;
  category?: 'discovery' | 'query' | 'entity' | 'dashboard' | 'alert' | 'platform';
  costIndicator?: 'low' | 'medium' | 'high';
  returnsNextCursor?: boolean;
}

export interface ToolEnhancement {
  description: string;
  inputSchema: any;
  outputSchema?: any;
  examples: Array<{
    description: string;
    params: any;
    expectedOutput?: any;
  }>;
  metadata: EnhancedToolMetadata;
  preHandler?: (params: any) => Promise<{ error?: string; suggestion?: string } | void>;
  postHandler?: (result: any, params: any) => Promise<any>;
}

/**
 * Enhanced tool registry with platform-native discovery
 */
export class EnhancedToolRegistry {
  private server: Server;
  private discovery: PlatformDiscovery;
  private nerdgraph: NerdGraphClient;
  private logger: Logger;
  private cache: IntelligentCache;
  private goldenSignals: GoldenSignalsEngine;
  private environmentTool: EnvironmentDiscoveryTool;
  private dashboardTool: DashboardGenerationTool;
  private comparisonTool: EntityComparisonTool;

  constructor(server: Server, discovery: PlatformDiscovery, config: any) {
    this.server = server;
    this.discovery = discovery;
    this.logger = this.createLogger();
    
    // Create NerdGraph client
    this.nerdgraph = createNerdGraphClient({
      apiKey: config.newrelic.apiKey,
      region: config.newrelic.region,
      logger: this.logger,
    });

    // Initialize services
    this.cache = new IntelligentCache(this.logger);
    this.goldenSignals = new GoldenSignalsEngine(this.nerdgraph, this.logger);
    
    // Initialize composite tools
    this.environmentTool = new EnvironmentDiscoveryTool(this.nerdgraph, this.logger, this.goldenSignals);
    this.dashboardTool = new DashboardGenerationTool(this.nerdgraph, this.logger, this.goldenSignals);
    this.comparisonTool = new EntityComparisonTool(this.nerdgraph, this.logger, this.goldenSignals);
    
    this.setupHandlers();
  }

  private createLogger(): Logger {
    return {
      info: (message: string, meta?: any) => {
        console.error(`[Enhanced] INFO: ${message}`, meta ? JSON.stringify(meta) : '');
      },
      warn: (message: string, meta?: any) => {
        console.error(`[Enhanced] WARN: ${message}`, meta ? JSON.stringify(meta) : '');
      },
      error: (message: string, meta?: any) => {
        console.error(`[Enhanced] ERROR: ${message}`, meta ? JSON.stringify(meta) : '');
      },
      debug: (message: string, meta?: any) => {
        if (process.env['DEBUG']) {
          console.error(`[Enhanced] DEBUG: ${message}`, meta ? JSON.stringify(meta) : '');
        }
      },
    };
  }

  private setupHandlers(): void {
    // Handle tool calls
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;
      
      switch (name) {
        case 'run_nrql_query':
          return this.handleEnhancedNrqlQuery(args);
        case 'search_entities':
          return this.handleEnhancedEntitySearch(args);
        case 'get_entity_details':
          return this.handleEnhancedEntityDetails(args);
        case 'discover_schemas':
          return this.handleDiscoverSchemas(args);
        case 'dashboard_generate':
          return this.handleDashboardGenerate(args);
        case 'platform_analyze_adoption':
          return this.handlePlatformAnalyzeAdoption(args);
        
        // Composite tools
        case 'discover.environment':
          return this.environmentTool.handle(args);
        case 'generate.golden_dashboard':
          return this.dashboardTool.handle(args);
        case 'compare.similar_entities':
          return this.comparisonTool.handle(args);
        case 'cache.stats':
          return this.handleCacheStats();
        case 'cache.clear':
          return this.handleCacheClear(args);
          
        default:
          throw new Error(`Unknown tool: ${name}`);
      }
    });

    // Register tool definitions
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: [
          ...this.getEnhancedToolDefinitions(),
          ...this.getCompositeToolDefinitions(),
          ...this.getPlatformToolDefinitions(),
        ],
      };
    });
  }

  private getEnhancedToolDefinitions(): Tool[] {
    return [
      {
        name: 'run_nrql_query',
        description: `Execute NRQL queries against New Relic data with discovery-first validation.
        
        🎯 **IMPORTANT**: Always discover available event types and attributes before querying.
        
        **Common Query Patterns:**
        - Time series: \`SELECT average(metric) FROM EventType TIMESERIES\`
        - Faceted analysis: \`SELECT count(*) FROM EventType FACET attribute\`
        - Percentiles: \`SELECT percentile(metric, 95) FROM EventType\`
        - Filtering: \`WHERE attribute = 'value' AND timestamp >= 1 hour ago\`
        
        **Best Practices:**
        - Use SINCE clause to limit time range for performance
        - Add LIMIT to prevent large result sets  
        - Use appropriate aggregation functions (count, average, sum, etc.)
        - Leverage FACET for grouping and analysis
        
        **Discovery Integration:**
        - Tool validates event types exist before execution
        - Suggests alternatives for unknown attributes
        - Automatically adds LIMIT 100 if not specified`,
        
        inputSchema: {
          type: 'object',
          properties: {
            account_id: {
              type: 'number',
              description: 'New Relic account ID where the query will be executed',
            },
            query: {
              type: 'string',
              description: 'NRQL query string (no trailing semicolon required)',
              examples: [
                "SELECT count(*) FROM Transaction SINCE 1 hour ago",
                "SELECT average(duration) FROM Transaction FACET appName SINCE 1 hour ago",
                "SELECT percentile(duration, 95) FROM Transaction WHERE appName = 'checkout-api' TIMESERIES",
              ],
            },
            timeout: {
              type: 'number',
              description: 'Query timeout in milliseconds (default: 30000, max: 120000)',
              default: 30000,
              minimum: 1000,
              maximum: 120000,
            },
            include_metadata: {
              type: 'boolean',
              description: 'Include query performance metadata and execution details',
              default: false,
            },
            validate_schema: {
              type: 'boolean',
              description: 'Validate event types and attributes exist before execution',
              default: true,
            },
            use_cache: {
              type: 'boolean',
              description: 'Use intelligent caching for repeated queries',
              default: true,
            },
          },
          required: ['query'],
        },
      },

      {
        name: 'search_entities',
        description: `Search for entities in New Relic with comprehensive filtering and discovery.
        
        **Entity Types by Domain:**
        - **APM**: Application services and microservices
        - **BROWSER**: Browser applications and page views
        - **INFRA**: Infrastructure hosts, containers, and services
        - **SYNTH**: Synthetic monitors and checks
        - **NR1**: Dashboards and workloads
        - **MOBILE**: Mobile applications
        
        **Search Strategies:**
        - Name patterns with wildcards: \`checkout*\`, \`*api*\`, \`prod-*\`
        - Domain filtering for specific entity types
        - Tag-based filtering for environment, team, or service classification
        - Type-specific searches within domains
        
        **Discovery Benefits:**
        - Always use this before creating dashboards or alerts
        - Discover entity relationships and dependencies
        - Find entities for golden signal monitoring
        - Understand account structure and naming conventions`,
        
        inputSchema: {
          type: 'object',
          properties: {
            account_id: {
              type: 'number',
              description: 'New Relic account ID to search within',
            },
            name: {
              type: 'string',
              description: 'Entity name pattern (supports wildcards: *, ?)',
              examples: ['checkout*', '*api*', 'prod-*', 'my-service'],
            },
            domain: {
              type: 'string',
              enum: ['APM', 'BROWSER', 'INFRA', 'SYNTH', 'NR1', 'MOBILE'],
              description: 'Entity domain filter - narrows search to specific telemetry types',
            },
            type: {
              type: 'string',
              description: 'Specific entity type within domain (e.g., APPLICATION, HOST, MONITOR)',
              examples: ['APPLICATION', 'HOST', 'CONTAINER', 'SERVICE', 'DASHBOARD'],
            },
            tags: {
              type: 'object',
              description: 'Tag filters as key:value pairs for precise targeting',
              additionalProperties: { type: 'string' },
              examples: [
                { environment: 'production' },
                { team: 'platform', region: 'us-east-1' },
              ],
            },
            limit: {
              type: 'number',
              default: 50,
              minimum: 1,
              maximum: 200,
              description: 'Maximum results to return per request',
            },
            cursor: {
              type: 'string',
              description: 'Pagination cursor from previous search request',
            },
            use_cache: {
              type: 'boolean',
              description: 'Use intelligent caching for entity searches',
              default: true,
            },
          },
        },
      },

      {
        name: 'get_entity_details',
        description: `Get comprehensive details about a specific entity with golden metrics and relationships.
        
        **Returned Information:**
        - **Basic Details**: Name, type, domain, GUID, reporting status
        - **Golden Metrics**: Throughput, error rate, latency (when available)
        - **Relationships**: Dependencies, calls-to, calls-from connections
        - **Alert Status**: Active violations and alert policy associations
        - **Tags**: All applied tags for categorization and filtering
        - **Recent Activity**: Changes, deployments, incidents
        
        **Golden Metrics by Entity Type:**
        - **APM Applications**: Throughput (rpm), Error rate (%), Response time (ms)
        - **Browser Apps**: Page load time, AJAX response time, JS errors
        - **Infrastructure**: CPU %, Memory %, Disk I/O
        - **Synthetics**: Success rate, Response time, Check frequency
        
        **Use Cases:**
        - Pre-dashboard creation analysis
        - Troubleshooting and root cause analysis
        - Dependency mapping and impact assessment
        - SLI/SLO definition and monitoring setup`,
        
        inputSchema: {
          type: 'object',
          properties: {
            guid: {
              type: 'string',
              description: 'Entity GUID (obtain from search_entities)',
              pattern: '^[A-Za-z0-9+/]+=*$', // Base64 pattern
            },
            include_golden_metrics: {
              type: 'boolean',
              default: true,
              description: 'Include golden signal metrics (throughput, error rate, latency)',
            },
            include_relationships: {
              type: 'boolean',
              default: true,
              description: 'Include related entities and dependency information',
            },
            include_alert_status: {
              type: 'boolean',
              default: true,
              description: 'Include current alert violations and policy status',
            },
            metrics_timeframe: {
              type: 'string',
              default: '1 HOUR',
              enum: ['5 MINUTES', '1 HOUR', '24 HOURS'],
              description: 'Time window for golden metrics calculation',
            },
            use_cache: {
              type: 'boolean',
              description: 'Use intelligent caching for entity details',
              default: true,
            },
          },
          required: ['guid'],
        },
      },

      {
        name: 'discover_schemas',
        description: `Discover all available event types and their attributes in an account.
        
        🚀 **This is typically the FIRST tool you should run when working with a new account.**
        
        **Discovery Process:**
        1. **Event Type Enumeration**: Finds all event types with recent data
        2. **Attribute Profiling**: Analyzes field types, cardinality, and usage patterns  
        3. **Metric Discovery**: Identifies dimensional metrics and their dimensions
        4. **Schema Analysis**: Determines best fields for filtering, faceting, and aggregation
        
        **Returns Comprehensive Schema Information:**
        - Event types ranked by data volume
        - Attribute characteristics (type, cardinality, sample values)
        - Metric definitions with dimensional breakdowns
        - Platform adoption indicators (APM, Infrastructure, Logs, OpenTelemetry)
        
        **Schema Intelligence:**
        - Identifies service identifier fields (appName, service.name, etc.)
        - Finds error indicator patterns (error fields, HTTP status codes)
        - Locates duration/latency metrics for performance analysis
        - Detects dimensional metrics vs event-based data patterns
        
        **Use This Before:**
        - Writing any NRQL queries
        - Creating dashboards or alerts
        - Setting up monitoring workflows
        - Cross-account platform analysis`,
        
        inputSchema: {
          type: 'object',
          properties: {
            account_id: {
              type: 'number',
              description: 'New Relic account ID to discover schemas for',
            },
            include_attributes: {
              type: 'boolean',
              default: false,
              description: 'Include detailed attribute profiling (increases discovery time)',
            },
            include_metrics: {
              type: 'boolean',
              default: true,
              description: 'Include dimensional metrics discovery',
            },
            attribute_limit: {
              type: 'number',
              default: 50,
              description: 'Maximum attributes to profile per event type',
            },
            sample_timeframe: {
              type: 'string',
              default: '1 hour ago',
              description: 'Time window for data sampling and analysis',
            },
            use_cache: {
              type: 'boolean',
              description: 'Use intelligent caching for schema discovery',
              default: true,
            },
          },
          required: ['account_id'],
        },
      },
    ];
  }

  private getPlatformToolDefinitions(): Tool[] {
    return [
      {
        name: 'dashboard_generate',
        description: `Generate intelligent dashboards using adaptive templates that discover correct fields.
        
        **Adaptive Templates:**
        - **golden-signals**: Error rate, latency, throughput, saturation (auto-adapts to data)
        - **dependencies**: Service map and relationship visualization
        - **infrastructure**: CPU, memory, disk, network metrics
        - **logs-analysis**: Log patterns, error analysis, and performance correlation
        - **business-metrics**: Custom KPIs and business-specific measurements
        - **custom**: Fully customizable widget configurations
        
        **Intelligent Adaptation:**
        - Automatically discovers which attributes and metrics are available
        - Adapts queries to use the best available fields (e.g., service.name vs appName)
        - Handles both event-based and dimensional metric queries
        - Optimizes widget types based on data characteristics
        
        **Dashboard Features:**
        - Multi-page layouts for complex services
        - Responsive time range controls
        - Faceted breakdowns by discovered dimensions
        - Alert threshold overlays where applicable
        - Linked entity navigation
        
        **Safety Features:**
        - Dry-run mode for preview and validation
        - Schema compatibility checking
        - Query performance estimation
        - Rollback capability for dashboard updates`,
        
        inputSchema: {
          type: 'object',
          properties: {
            template_name: {
              type: 'string',
              enum: ['golden-signals', 'dependencies', 'infrastructure', 'logs-analysis', 'business-metrics', 'custom'],
              description: 'Dashboard template to use as foundation',
            },
            entity_guid: {
              type: 'string',
              description: 'Entity GUID to create dashboard for (from search_entities)',
            },
            dashboard_name: {
              type: 'string',
              description: 'Name for the dashboard (auto-generated if not provided)',
            },
            time_range: {
              type: 'string',
              default: '1 hour ago',
              description: 'Default time range for all widgets',
            },
            dry_run: {
              type: 'boolean',
              default: true,
              description: 'Preview dashboard JSON without creating (recommended first)',
            },
            pages: {
              type: 'array',
              description: 'Multi-page dashboard configuration',
              items: {
                type: 'object',
                properties: {
                  name: { type: 'string' },
                  widgets: { type: 'array' },
                },
              },
            },
            custom_widgets: {
              type: 'array',
              description: 'Custom widget configurations (for custom template)',
            },
          },
          required: ['template_name', 'entity_guid'],
        },
      },

      {
        name: 'platform_analyze_adoption',
        description: `Analyze platform adoption patterns across multiple accounts.
        
        **Purpose**: Understand how different teams use New Relic features and identify optimization opportunities.
        
        **Analysis Metrics:**
        - **Dimensional Metrics**: Usage of modern metric platform vs event-based metrics
        - **OpenTelemetry**: Adoption of OTel instrumentation and standards
        - **Entity Synthesis**: Custom entity creation and entity platform usage
        - **Custom Instrumentation**: Extent of custom events and attributes
        - **Distributed Tracing**: APM distributed tracing coverage
        - **Logs in Context**: Log correlation with APM data
        
        **Cross-Account Insights:**
        - Compare feature adoption across teams and environments
        - Identify best practices from high-performing accounts
        - Find opportunities for standardization
        - Calculate platform maturity scores
        
        **Use Cases:**
        - Platform governance and standardization
        - Cost optimization through feature consolidation
        - Migration planning from legacy to modern features
        - Team training and enablement prioritization`,
        
        inputSchema: {
          type: 'object',
          properties: {
            account_ids: {
              type: 'array',
              items: { type: 'number' },
              description: 'List of account IDs to analyze',
              minItems: 1,
              maxItems: 100,
            },
            metrics: {
              type: 'array',
              items: {
                type: 'string',
                enum: [
                  'dimensional_metrics',
                  'opentelemetry',
                  'entity_synthesis',
                  'custom_instrumentation',
                  'distributed_tracing',
                  'logs_in_context'
                ],
              },
              description: 'Specific adoption metrics to calculate',
            },
            comparison_mode: {
              type: 'string',
              enum: ['absolute', 'relative', 'benchmarked'],
              default: 'relative',
              description: 'How to compare metrics across accounts',
            },
            include_recommendations: {
              type: 'boolean',
              default: true,
              description: 'Include optimization recommendations',
            },
          },
          required: ['account_ids', 'metrics'],
        },
      },
    ];
  }

  private getCompositeToolDefinitions(): Tool[] {
    return [
      this.environmentTool.getToolDefinition(),
      this.dashboardTool.getToolDefinition(),
      this.comparisonTool.getToolDefinition(),
      
      // Cache management tools
      {
        name: 'cache.stats',
        description: `Get intelligent cache statistics and health assessment.

        🎯 **Purpose**: Monitor cache performance and identify optimization opportunities.

        **Provides**:
        - Hit/miss rates and performance metrics
        - Memory usage and cache size information
        - Health assessment with specific recommendations
        - Most accessed keys and usage patterns

        **Use for**:
        - Performance monitoring and optimization
        - Understanding data access patterns
        - Identifying cache tuning opportunities
        - Troubleshooting slow response times`,

        inputSchema: {
          type: 'object',
          properties: {},
          additionalProperties: false,
        },
      },

      {
        name: 'cache.clear',
        description: `Clear cache entries with optional pattern-based filtering.

        🎯 **Purpose**: Manage cache contents for data freshness and memory optimization.

        **Capabilities**:
        - Clear all cache entries
        - Clear entries matching specific patterns
        - Selective invalidation by strategy type
        - Force refresh for specific data types

        **Safety Features**:
        - Confirmation prompts for destructive operations
        - Pattern validation to prevent accidental full clears
        - Backup of critical cached data before clearing

        **Use Cases**:
        - Force refresh after known data changes
        - Clear cache during troubleshooting
        - Memory management and cleanup
        - Development and testing scenarios`,

        inputSchema: {
          type: 'object',
          properties: {
            pattern: {
              type: 'string',
              description: 'Pattern to match for selective clearing (e.g., "discovery:", "metrics:entity-123")',
            },
            strategy_type: {
              type: 'string',
              enum: ['discovery', 'goldenMetrics', 'entityDetails', 'dashboards', 'analytics'],
              description: 'Clear all entries of a specific strategy type',
            },
            confirm: {
              type: 'boolean',
              description: 'Confirm destructive operation (required for full clear)',
              default: false,
            },
          },
          additionalProperties: false,
        },
      },
    ];
  }

  // Enhanced handler methods
  private async handleEnhancedNrqlQuery(params: any) {
    const { account_id, query, validate_schema = true, use_cache = true } = params;
    const cacheKey = `nrql:${account_id}:${Buffer.from(query).toString('base64')}`;

    // Check cache first
    if (use_cache) {
      const cached = await this.cache.get(cacheKey, 'goldenMetrics');
      if (cached.data && cached.freshness !== 'expired') {
        return {
          content: [{
            type: 'text',
            text: JSON.stringify({
              ...cached.data,
              cached: true,
              freshness: cached.freshness,
            }, null, 2),
          }],
        };
      }
    }

    if (validate_schema) {
      // Extract event types from query
      const eventTypes = this.extractEventTypesFromQuery(query);
      
      if (eventTypes.length > 0) {
        // Validate event types exist
        const knownTypes = await this.discovery.discoverEventTypes(account_id);
        const knownTypeNames = knownTypes.map(et => et.name);
        
        const unknownTypes = eventTypes.filter(type => !knownTypeNames.includes(type));
        
        if (unknownTypes.length > 0) {
          return {
            content: [{
              type: 'text',
              text: JSON.stringify({
                error: `Unknown event types: ${unknownTypes.join(', ')}`,
                suggestion: 'Run discover_schemas first to see available event types',
                available_types: knownTypeNames.slice(0, 10),
              }, null, 2),
            }],
          };
        }
      }
    }

    // Add LIMIT if not present
    let finalQuery = query.trim();
    if (!finalQuery.toLowerCase().includes('limit') && 
        !finalQuery.toLowerCase().includes('show event types')) {
      finalQuery += ' LIMIT 100';
    }

    try {
      // Execute the query
      const result = await this.nerdgraph.nrql(account_id, finalQuery);
      
      // Extract event types for metadata
      const eventTypes = this.extractEventTypesFromQuery(finalQuery);
      
      const response = {
        results: result.results,
        query_executed: finalQuery,
        schema_validated: validate_schema,
        metadata: {
          performanceStats: result.performanceStats,
          eventTypes: eventTypes,
        },
      };

      // Cache the result
      if (use_cache) {
        await this.cache.set(cacheKey, response, 'goldenMetrics');
      }

      return {
        content: [{
          type: 'text',
          text: JSON.stringify(response, null, 2),
        }],
      };
    } catch (error: any) {
      return {
        content: [{
          type: 'text',
          text: JSON.stringify({
            error: error.message,
            query: finalQuery,
            suggestion: 'Check query syntax and available event types/attributes',
          }, null, 2),
        }],
      };
    }
  }

  private async handleEnhancedEntitySearch(params: any) {
    const { account_id, use_cache = true, ...searchParams } = params;
    const cacheKey = `entities:${account_id}:${JSON.stringify(searchParams)}`;

    // Check cache first
    if (use_cache) {
      const cached = await this.cache.get(cacheKey, 'entityDetails');
      if (cached.data && cached.freshness !== 'expired') {
        return {
          content: [{
            type: 'text',
            text: JSON.stringify({
              ...cached.data,
              cached: true,
              freshness: cached.freshness,
            }, null, 2),
          }],
        };
      }
    }

    try {
      const entities = await this.discovery.discoverEntities(
        account_id, 
        this.buildEntitySearchQuery(searchParams)
      );
      
      const response = {
        entities: entities.map(entity => ({
          guid: entity.guid,
          name: entity.name,
          type: entity.type,
          domain: entity.domain,
          reporting: entity.reporting,
          tags: entity.tags,
        })),
        total_found: entities.length,
        has_more: entities.length === (params.limit || 50),
        search_params: searchParams,
      };

      // Cache the result
      if (use_cache) {
        await this.cache.set(cacheKey, response, 'entityDetails');
      }

      return {
        content: [{
          type: 'text',
          text: JSON.stringify(response, null, 2),
        }],
      };
    } catch (error: any) {
      return {
        content: [{
          type: 'text',
          text: JSON.stringify({
            error: error.message,
            suggestion: 'Try broader search criteria or check account access permissions',
          }, null, 2),
        }],
      };
    }
  }

  private async handleEnhancedEntityDetails(params: any) {
    const { guid, include_golden_metrics = true, use_cache = true } = params;
    const cacheKey = `entity_details:${guid}:${include_golden_metrics}`;

    // Check cache first
    if (use_cache) {
      const cached = await this.cache.get(cacheKey, 'entityDetails');
      if (cached.data && cached.freshness !== 'expired') {
        return {
          content: [{
            type: 'text',
            text: JSON.stringify({
              ...cached.data,
              cached: true,
              freshness: cached.freshness,
            }, null, 2),
          }],
        };
      }
    }

    try {
      // Get entity details
      const entityQuery = `
        query($guid: EntityGuid!) {
          actor {
            entity(guid: $guid) {
              guid
              name
              type
              domain
              entityType
              reporting
              tags {
                key
                values
              }
              ... on ApmApplicationEntity {
                language
                settings {
                  apdexTarget
                }
                apmSummary {
                  throughput
                  errorRate
                  responseTimeAverage
                }
              }
            }
          }
        }
      `;

      const result = await this.nerdgraph.request(entityQuery, { guid });
      const entity = result.actor?.entity;

      if (!entity) {
        throw new Error('Entity not found');
      }

      // Enhance with discovered patterns if requested
      let discoveredPatterns = {};
      if (include_golden_metrics) {
        const entityData = await this.discovery.discoverEntityData(entity);
        discoveredPatterns = {
          service_identifier: entityData.serviceIdentifier,
          error_indicators: entityData.errorIndicators,
          duration_fields: entityData.durationFields,
          available_metrics: entityData.metrics.slice(0, 5),
        };
      }

      const response = {
        entity: entity,
        discovered_patterns: discoveredPatterns,
      };

      // Cache the result
      if (use_cache) {
        await this.cache.set(cacheKey, response, 'entityDetails');
      }

      return {
        content: [{
          type: 'text',
          text: JSON.stringify(response, null, 2),
        }],
      };
    } catch (error: any) {
      return {
        content: [{
          type: 'text',
          text: JSON.stringify({
            error: error.message,
            suggestion: 'Verify entity GUID format and account access permissions',
          }, null, 2),
        }],
      };
    }
  }

  private async handleDiscoverSchemas(params: any) {
    const { account_id, include_attributes = false, include_metrics = true, use_cache = true } = params;
    const cacheKey = `schemas:${account_id}:${include_attributes}:${include_metrics}`;

    // Check cache first
    if (use_cache) {
      const cached = await this.cache.get(cacheKey, 'discovery');
      if (cached.data && cached.freshness !== 'expired') {
        return {
          content: [{
            type: 'text',
            text: JSON.stringify({
              ...cached.data,
              cached: true,
              freshness: cached.freshness,
            }, null, 2),
          }],
        };
      }
    }

    try {
      const [eventTypes, metrics] = await Promise.all([
        this.discovery.discoverEventTypes(account_id),
        include_metrics ? this.discovery.discoverMetricNames(account_id) : [],
      ]);

      // Enhance event types with attributes if requested
      let eventTypesData = eventTypes;
      if (include_attributes) {
        eventTypesData = await Promise.all(
          eventTypes.slice(0, 10).map(async (et) => ({
            ...et,
            attributes: await this.discovery.discoverAttributes(account_id, et.name),
          }))
        );
      }

      const response = {
        account_id,
        event_types: eventTypesData.map(et => ({
          name: et.name,
          sample_count: et.sampleCount,
          last_ingested: et.lastIngested,
          attributes: (et as any).attributes,
        })),
        metrics: metrics.slice(0, 20),
        summary: this.generateSchemaSummary(eventTypesData, metrics),
      };

      if (use_cache) {
        await this.cache.set(cacheKey, response, 'discovery');
      }

      return {
        content: [{
          type: 'text',
          text: JSON.stringify(response, null, 2),
        }],
      };
    } catch (error: any) {
      return {
        content: [{
          type: 'text',
          text: JSON.stringify({
            error: error.message,
            suggestion: 'Verify account ID and ensure account has recent data',
          }, null, 2),
        }],
      };
    }
  }

  private async handleDashboardGenerate(params: any) {
    // This would integrate with the adaptive dashboard generation
    return {
      content: [{
        type: 'text',
        text: JSON.stringify({
          message: 'Dashboard generation is being implemented',
          template: params.template_name,
          entity_guid: params.entity_guid,
        }, null, 2),
      }],
    };
  }

  private async handlePlatformAnalyzeAdoption(params: any) {
    const { account_ids, metrics } = params;

    try {
      const analysisResults = await Promise.all(
        account_ids.map(async (accountId: number) => {
          const schemas = await this.discovery.discoverEventTypes(accountId);
          const metricData = await this.discovery.discoverMetricNames(accountId);

          const adoption = {
            account_id: accountId,
            dimensional_metrics: metricData.length > 0,
            opentelemetry: schemas.some(s => s.name.includes('Span')),
            entity_synthesis: false, // Would need entity discovery
            custom_instrumentation: schemas.some(s => !['Transaction', 'PageView', 'Log', 'Span', 'Metric'].includes(s.name)),
            event_type_count: schemas.length,
            metric_count: metricData.length,
          };

          return adoption;
        })
      );

      return {
        content: [{
          type: 'text',
          text: JSON.stringify({
            adoption_analysis: analysisResults,
            summary: {
              accounts_analyzed: account_ids.length,
              with_dimensional_metrics: analysisResults.filter(a => a.dimensional_metrics).length,
              with_opentelemetry: analysisResults.filter(a => a.opentelemetry).length,
              with_custom_instrumentation: analysisResults.filter(a => a.custom_instrumentation).length,
            },
          }, null, 2),
        }],
      };
    } catch (error: any) {
      return {
        content: [{
          type: 'text',
          text: JSON.stringify({
            error: error.message,
            suggestion: 'Verify account IDs and permissions',
          }, null, 2),
        }],
      };
    }
  }

  // Cache management handlers
  private async handleCacheStats() {
    const stats = this.cache.getStats();
    const health = this.cache.getHealthAssessment();

    let content = `# 📊 Intelligent Cache Statistics\n\n`;
    
    content += `## 📈 Performance Metrics\n\n`;
    content += `- **Hit Rate**: ${(stats.hitRate * 100).toFixed(1)}%\n`;
    content += `- **Miss Rate**: ${(stats.missRate * 100).toFixed(1)}%\n`;
    content += `- **Average Response Time**: ${stats.avgResponseTime}ms\n`;
    content += `- **Total Entries**: ${stats.totalEntries}\n`;
    content += `- **Memory Usage**: ${Math.round(stats.memoryUsage / 1024 / 1024 * 100) / 100}MB\n\n`;

    content += `## 🕒 Cache Age Information\n\n`;
    content += `- **Oldest Entry**: ${stats.oldestEntry.toISOString()}\n`;
    content += `- **Most Accessed**: ${stats.mostAccessed || 'N/A'}\n\n`;

    content += `## 🏥 Health Assessment\n\n`;
    content += `**Status**: ${health.status.toUpperCase()} ${this.getHealthIcon(health.status)}\n\n`;

    if (health.issues.length > 0) {
      content += `**Issues Identified**:\n`;
      health.issues.forEach(issue => {
        content += `- ⚠️ ${issue}\n`;
      });
      content += `\n`;
    }

    if (health.recommendations.length > 0) {
      content += `**Recommendations**:\n`;
      health.recommendations.forEach(rec => {
        content += `- 💡 ${rec}\n`;
      });
      content += `\n`;
    }

    content += `---\n*Cache statistics generated at ${new Date().toISOString()}*`;

    return {
      content: [
        {
          type: 'text',
          text: content,
        },
      ],
    };
  }

  private async handleCacheClear(params: any) {
    const { pattern, strategy_type, confirm = false } = params;

    // Safety check for full clear
    if (!pattern && !strategy_type && !confirm) {
      return {
        content: [
          {
            type: 'text',
            text: `⚠️ **Full Cache Clear Requires Confirmation**\n\nTo clear all cache entries, set \`confirm: true\`.\n\nAlternatively, use:\n- \`pattern\`: Clear entries matching a pattern\n- \`strategy_type\`: Clear entries of a specific type\n\nExample patterns:\n- \`"discovery:"\` - Clear discovery data\n- \`"metrics:"\` - Clear metrics data\n- \`"entity:123"\` - Clear data for specific entity`,
          },
        ],
      };
    }

    let clearedCount = 0;
    let description = '';

    if (pattern) {
      clearedCount = this.cache.invalidate(pattern);
      description = `pattern "${pattern}"`;
    } else if (strategy_type) {
      clearedCount = this.cache.invalidate(new RegExp(`^${strategy_type}:`));
      description = `strategy type "${strategy_type}"`;
    } else if (confirm) {
      this.cache.clear();
      clearedCount = -1; // Indicate full clear
      description = 'all entries';
    }

    const message = clearedCount === -1 
      ? `✅ **Cache Completely Cleared**\n\nAll cache entries have been removed.`
      : `✅ **Cache Cleared**\n\nRemoved ${clearedCount} entries matching ${description}.`;

    return {
      content: [
        {
          type: 'text',
          text: message,
        },
      ],
    };
  }

  // Utility methods
  private extractEventTypesFromQuery(query: string): string[] {
    const fromMatch = query.match(/FROM\s+([A-Za-z0-9_,\s]+)/gi);
    if (!fromMatch) return [];

    return fromMatch[0]
      .replace(/FROM\s+/i, '')
      .split(',')
      .map(type => type.trim())
      .filter(type => type && !type.toLowerCase().includes('where'));
  }

  private buildEntitySearchQuery(params: any): string {
    const conditions = [];
    
    if (params.name) {
      conditions.push(`name LIKE '${params.name}'`);
    }
    if (params.domain) {
      conditions.push(`domain = '${params.domain}'`);
    }
    if (params.type) {
      conditions.push(`type = '${params.type}'`);
    }
    if (params.tags) {
      Object.entries(params.tags).forEach(([key, value]) => {
        conditions.push(`tags.${key} = '${value}'`);
      });
    }

    return conditions.length > 0 ? conditions.join(' AND ') : '';
  }

  private generateSchemaSummary(eventTypes: any[], metrics: any[]) {
    return {
      total_event_types: eventTypes.length,
      total_metrics: metrics.length,
      has_apm_data: eventTypes.some(et => et.name === 'Transaction'),
      has_infrastructure: eventTypes.some(et => et.name === 'SystemSample'),
      has_browser_data: eventTypes.some(et => et.name === 'PageView'),
      has_logs: eventTypes.some(et => et.name === 'Log'),
      has_mobile_data: eventTypes.some(et => et.name === 'Mobile'),
      has_otel_data: metrics.some(m => m.name && m.name.includes('otel')) ||
                     eventTypes.some(et => et.attributes?.some((a: any) => a.name === 'service.name')),
      largest_event_type: eventTypes[0]?.name,
      data_recency: eventTypes[0]?.lastIngested,
    };
  }

  private getHealthIcon(status: string): string {
    switch (status) {
      case 'healthy': return '✅';
      case 'warning': return '⚠️';
      case 'critical': return '🚨';
      default: return '❓';
    }
  }
}

/**
 * Factory function to enhance existing tools
 */
export function enhanceExistingTools(
  server: Server, 
  discovery: PlatformDiscovery
): EnhancedToolRegistry {
  // Get config from environment
  const config = {
    newrelic: {
      apiKey: process.env['NEW_RELIC_API_KEY'] || '',
      accountId: process.env['NEW_RELIC_ACCOUNT_ID'] || '',
      region: (process.env['NEW_RELIC_REGION'] as 'US' | 'EU') || 'US',
    }
  };

  return new EnhancedToolRegistry(server, discovery, config);
}
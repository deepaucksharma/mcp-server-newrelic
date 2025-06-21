/**
 * Comprehensive E2E Test Harness for Platform-Native MCP Server
 * 
 * Tests all aspects of discovery, adaptation, and tool enhancement
 * with real New Relic Database (NRDB) backend integration.
 */

import { describe, beforeAll, afterAll, beforeEach, afterEach, it, expect } from 'vitest';
import { createServer, MCPNewRelicServer } from '../../src/index.js';
import { PlatformDiscovery } from '../../src/core/platform-discovery.js';
import { AdaptiveDashboardGenerator } from '../../src/tools/adaptive-dashboards.js';
import { createNerdGraphClient } from '../../src/adapters/nerdgraph.js';

// Test Configuration Types
export interface TestAccount {
  id: number;
  name: string;
  apiKey: string;
  region: 'US' | 'EU';
  dataPatterns: DataPattern[];
  expectedSchemas: string[];
  serviceIdentifierField?: string;
}

export interface DataPattern {
  type: 'apm' | 'infrastructure' | 'browser' | 'logs' | 'mobile' | 'synthetic' | 'opentelemetry' | 'custom';
  eventTypes: string[];
  volume: 'low' | 'medium' | 'high' | 'sparse';
  hasErrors: boolean;
  hasMetrics: boolean;
  lastIngested: Date;
}

export interface TestScenario {
  name: string;
  description: string;
  accounts: TestAccount[];
  expectedBehavior: {
    discoverySuccess: boolean;
    toolsOperational: boolean;
    adaptiveDashboards: boolean;
    performanceThresholds: PerformanceThresholds;
  };
}

export interface PerformanceThresholds {
  discoveryLatency: number; // ms
  queryLatency: number; // ms
  cacheHitRatio: number; // percentage
  memoryUsage: number; // MB
}

// Test Data Definitions
export class E2ETestHarness {
  private server?: MCPNewRelicServer;
  private discovery?: PlatformDiscovery;
  private dashboardGenerator?: AdaptiveDashboardGenerator;
  
  // Test account configurations for comprehensive coverage
  private readonly testAccounts: TestAccount[] = [
    {
      id: parseInt(process.env['E2E_ACCOUNT_LEGACY_APM'] || '0'),
      name: 'Legacy APM Account',
      apiKey: process.env['E2E_API_KEY_LEGACY'] || '',
      region: 'US',
      dataPatterns: [
        {
          type: 'apm',
          eventTypes: ['Transaction', 'TransactionError', 'TransactionTrace'],
          volume: 'high',
          hasErrors: true,
          hasMetrics: false,
          lastIngested: new Date(Date.now() - 5 * 60 * 1000), // 5 minutes ago
        }
      ],
      expectedSchemas: ['Transaction', 'TransactionError'],
      serviceIdentifierField: 'appName',
    },
    {
      id: parseInt(process.env['E2E_ACCOUNT_MODERN_OTEL'] || '0'),
      name: 'Modern OpenTelemetry Account',
      apiKey: process.env['E2E_API_KEY_OTEL'] || '',
      region: 'US',
      dataPatterns: [
        {
          type: 'opentelemetry',
          eventTypes: ['Span', 'Metric'],
          volume: 'high',
          hasErrors: true,
          hasMetrics: true,
          lastIngested: new Date(Date.now() - 2 * 60 * 1000), // 2 minutes ago
        }
      ],
      expectedSchemas: ['Span'],
      serviceIdentifierField: 'service.name',
    },
    {
      id: parseInt(process.env['E2E_ACCOUNT_MIXED_DATA'] || '0'),
      name: 'Mixed Data Patterns Account',
      apiKey: process.env['E2E_API_KEY_MIXED'] || '',
      region: 'US',
      dataPatterns: [
        {
          type: 'apm',
          eventTypes: ['Transaction'],
          volume: 'medium',
          hasErrors: true,
          hasMetrics: false,
          lastIngested: new Date(Date.now() - 10 * 60 * 1000),
        },
        {
          type: 'infrastructure',
          eventTypes: ['SystemSample', 'ProcessSample'],
          volume: 'medium',
          hasErrors: false,
          hasMetrics: false,
          lastIngested: new Date(Date.now() - 5 * 60 * 1000),
        },
        {
          type: 'browser',
          eventTypes: ['PageView', 'PageAction'],
          volume: 'low',
          hasErrors: true,
          hasMetrics: false,
          lastIngested: new Date(Date.now() - 15 * 60 * 1000),
        },
        {
          type: 'logs',
          eventTypes: ['Log'],
          volume: 'high',
          hasErrors: true,
          hasMetrics: false,
          lastIngested: new Date(Date.now() - 3 * 60 * 1000),
        }
      ],
      expectedSchemas: ['Transaction', 'SystemSample', 'PageView', 'Log'],
    },
    {
      id: parseInt(process.env['E2E_ACCOUNT_SPARSE_DATA'] || '0'),
      name: 'Sparse Data Account',
      apiKey: process.env['E2E_API_KEY_SPARSE'] || '',
      region: 'US',
      dataPatterns: [
        {
          type: 'synthetic',
          eventTypes: ['SyntheticCheck'],
          volume: 'sparse',
          hasErrors: false,
          hasMetrics: false,
          lastIngested: new Date(Date.now() - 60 * 60 * 1000), // 1 hour ago
        }
      ],
      expectedSchemas: ['SyntheticCheck'],
    },
    {
      id: parseInt(process.env['E2E_ACCOUNT_EU_REGION'] || '0'),
      name: 'EU Region Account',
      apiKey: process.env['E2E_API_KEY_EU'] || '',
      region: 'EU',
      dataPatterns: [
        {
          type: 'apm',
          eventTypes: ['Transaction'],
          volume: 'medium',
          hasErrors: true,
          hasMetrics: false,
          lastIngested: new Date(Date.now() - 5 * 60 * 1000),
        }
      ],
      expectedSchemas: ['Transaction'],
      serviceIdentifierField: 'appName',
    }
  ];

  // Test scenarios covering various real-world configurations
  private readonly testScenarios: TestScenario[] = [
    {
      name: 'Legacy APM Discovery',
      description: 'Test discovery with traditional New Relic APM data patterns',
      accounts: [this.testAccounts[0]], // Legacy APM
      expectedBehavior: {
        discoverySuccess: true,
        toolsOperational: true,
        adaptiveDashboards: true,
        performanceThresholds: {
          discoveryLatency: 2000,
          queryLatency: 1000,
          cacheHitRatio: 70,
          memoryUsage: 100,
        },
      },
    },
    {
      name: 'OpenTelemetry Modern Stack',
      description: 'Test discovery with OpenTelemetry and dimensional metrics',
      accounts: [this.testAccounts[1]], // Modern OTEL
      expectedBehavior: {
        discoverySuccess: true,
        toolsOperational: true,
        adaptiveDashboards: true,
        performanceThresholds: {
          discoveryLatency: 1500,
          queryLatency: 800,
          cacheHitRatio: 80,
          memoryUsage: 120,
        },
      },
    },
    {
      name: 'Mixed Data Patterns',
      description: 'Test discovery across diverse telemetry types in single account',
      accounts: [this.testAccounts[2]], // Mixed data
      expectedBehavior: {
        discoverySuccess: true,
        toolsOperational: true,
        adaptiveDashboards: true,
        performanceThresholds: {
          discoveryLatency: 3000,
          queryLatency: 1200,
          cacheHitRatio: 75,
          memoryUsage: 150,
        },
      },
    },
    {
      name: 'Sparse Data Handling',
      description: 'Test graceful handling of accounts with minimal data',
      accounts: [this.testAccounts[3]], // Sparse data
      expectedBehavior: {
        discoverySuccess: true,
        toolsOperational: true,
        adaptiveDashboards: false, // Limited data may not support all dashboards
        performanceThresholds: {
          discoveryLatency: 1000,
          queryLatency: 500,
          cacheHitRatio: 60,
          memoryUsage: 50,
        },
      },
    },
    {
      name: 'Cross-Region Compatibility',
      description: 'Test EU region endpoint and data patterns',
      accounts: [this.testAccounts[4]], // EU region
      expectedBehavior: {
        discoverySuccess: true,
        toolsOperational: true,
        adaptiveDashboards: true,
        performanceThresholds: {
          discoveryLatency: 2500, // Potentially higher latency for cross-region
          queryLatency: 1100,
          cacheHitRatio: 70,
          memoryUsage: 100,
        },
      },
    },
    {
      name: 'Multi-Account Cross-Analysis',
      description: 'Test platform intelligence across multiple accounts',
      accounts: [this.testAccounts[0], this.testAccounts[1], this.testAccounts[2]], // Multiple accounts
      expectedBehavior: {
        discoverySuccess: true,
        toolsOperational: true,
        adaptiveDashboards: true,
        performanceThresholds: {
          discoveryLatency: 5000, // Higher for multi-account
          queryLatency: 2000,
          cacheHitRatio: 85, // Better cache utilization across accounts
          memoryUsage: 300,
        },
      },
    }
  ];

  /**
   * Initialize test harness with real NRDB connections
   */
  async setupTestHarness(): Promise<void> {
    console.log('🚀 Initializing E2E Test Harness with real NRDB backend...');
    
    // Validate test environment
    this.validateTestEnvironment();
    
    // Initialize server with test configuration
    // Set up environment variables for the MCP server
    process.env.NEW_RELIC_API_KEY = process.env.E2E_API_KEY_LEGACY || '';
    process.env.NEW_RELIC_ACCOUNT_ID = process.env.E2E_ACCOUNT_LEGACY_APM || '';
    process.env.NEW_RELIC_REGION = 'US';
    
    this.server = await createServer();
    this.discovery = this.server.getDiscovery();
    this.dashboardGenerator = this.server.getDashboardGenerator();
    
    console.log('✅ Test harness initialized successfully');
  }

  /**
   * Cleanup test harness and connections
   */
  async teardownTestHarness(): Promise<void> {
    console.log('🧹 Cleaning up test harness...');
    
    if (this.server) {
      await this.server.shutdown();
    }
    
    console.log('✅ Test harness cleaned up');
  }

  /**
   * Validate test environment has required credentials
   */
  private validateTestEnvironment(): void {
    const requiredEnvVars = [
      'E2E_ACCOUNT_LEGACY_APM',
      'E2E_API_KEY_LEGACY'
    ];
    
    const optionalEnvVars = [
      'E2E_ACCOUNT_MODERN_OTEL',
      'E2E_API_KEY_OTEL', 
      'E2E_ACCOUNT_MIXED_DATA',
      'E2E_API_KEY_MIXED'
    ];

    const missing = requiredEnvVars.filter(varName => !process.env[varName]);
    
    if (missing.length > 0) {
      throw new Error(`Missing required test environment variables: ${missing.join(', ')}`);
    }

    // Validate account IDs are numeric (only for provided accounts)
    const accountEnvVars = [
      'E2E_ACCOUNT_LEGACY_APM',
      'E2E_ACCOUNT_MODERN_OTEL', 
      'E2E_ACCOUNT_MIXED_DATA',
      'E2E_ACCOUNT_SPARSE_DATA',
      'E2E_ACCOUNT_EU_REGION'
    ];

    for (const envVar of accountEnvVars) {
      const value = process.env[envVar];
      if (value && (isNaN(parseInt(value)) || parseInt(value) <= 0)) {
        throw new Error(`Invalid account ID for ${envVar}: ${value}`);
      }
    }

    // Report available optional accounts
    const availableOptional = optionalEnvVars.filter(varName => process.env[varName]);
    if (availableOptional.length > 0) {
      console.log(`ℹ️  Optional accounts available: ${availableOptional.length / 2} additional accounts`);
    } else {
      console.log('ℹ️  Only legacy APM account configured - limited test coverage');
    }
    
    console.log('✅ Test environment validation passed');
  }

  /**
   * Run comprehensive discovery tests across all account types
   */
  async runDiscoveryTests(): Promise<DiscoveryTestResults> {
    console.log('🔍 Running comprehensive discovery tests...');
    
    const results: DiscoveryTestResults = {
      totalAccounts: this.testAccounts.length,
      successfulDiscoveries: 0,
      failedDiscoveries: 0,
      discoveryLatencies: [],
      schemaAccuracy: [],
      attributeCompleteness: [],
      serviceIdentifierAccuracy: [],
      errorIndicatorDetection: [],
      metricDetection: [],
    };

    for (const account of this.testAccounts) {
      if (!account.apiKey || account.id === 0) {
        console.log(`⏭️  Skipping account ${account.name} - no credentials provided`);
        continue;
      }

      console.log(`\n📊 Testing discovery for: ${account.name}`);
      
      try {
        const startTime = Date.now();
        
        // Test event type discovery
        const eventTypes = await this.discovery!.discoverEventTypes(account.id);
        const discoveryLatency = Date.now() - startTime;
        
        results.discoveryLatencies.push({
          accountName: account.name,
          latency: discoveryLatency,
        });

        // Validate discovered schemas match expectations
        const discoveredEventTypeNames = eventTypes.map(et => et.name);
        const schemaAccuracy = this.calculateSchemaAccuracy(
          discoveredEventTypeNames,
          account.expectedSchemas
        );
        
        results.schemaAccuracy.push({
          accountName: account.name,
          expected: account.expectedSchemas,
          discovered: discoveredEventTypeNames,
          accuracy: schemaAccuracy,
        });

        // Test attribute discovery for top event types
        const attributeResults = [];
        for (const eventType of eventTypes.slice(0, 3)) { // Top 3 by volume
          const attributes = await this.discovery!.discoverAttributes(account.id, eventType.name);
          attributeResults.push({
            eventType: eventType.name,
            attributeCount: attributes.length,
            sampleAttributes: attributes.slice(0, 5).map(a => a.name),
          });
        }
        
        results.attributeCompleteness.push({
          accountName: account.name,
          eventTypeResults: attributeResults,
        });

        // Test service identifier detection
        const mockEntity = {
          guid: 'test-guid',
          name: 'test-entity',
          type: 'APPLICATION',
          domain: 'APM',
          entityType: 'APPLICATION',
          reporting: true,
          tags: [],
        };
        
        const entityData = await this.discovery!.discoverEntityData(mockEntity);
        const serviceIdentifierCorrect = account.serviceIdentifierField 
          ? entityData.serviceIdentifier === account.serviceIdentifierField
          : true; // No expectation set

        results.serviceIdentifierAccuracy.push({
          accountName: account.name,
          expected: account.serviceIdentifierField,
          discovered: entityData.serviceIdentifier,
          correct: serviceIdentifierCorrect,
        });

        // Test error indicator detection
        results.errorIndicatorDetection.push({
          accountName: account.name,
          indicatorsFound: entityData.errorIndicators.length,
          indicators: entityData.errorIndicators.map(ei => ({
            field: ei.name,
            type: ei.type,
          })),
        });

        // Test metric discovery
        const metrics = await this.discovery!.discoverMetricNames(account.id);
        results.metricDetection.push({
          accountName: account.name,
          metricsFound: metrics.length,
          topMetrics: metrics.slice(0, 5).map(m => m.name),
        });

        results.successfulDiscoveries++;
        console.log(`✅ Discovery successful for ${account.name}`);
        console.log(`   📊 Event types: ${eventTypes.length}`);
        console.log(`   🏷️  Attributes: ${attributeResults.reduce((sum, r) => sum + r.attributeCount, 0)}`);
        console.log(`   ⚡ Latency: ${discoveryLatency}ms`);
        console.log(`   🎯 Schema accuracy: ${(schemaAccuracy * 100).toFixed(1)}%`);

      } catch (error: any) {
        results.failedDiscoveries++;
        console.error(`❌ Discovery failed for ${account.name}: ${error.message}`);
      }
    }

    console.log(`\n📈 Discovery test summary:`);
    console.log(`   ✅ Successful: ${results.successfulDiscoveries}`);
    console.log(`   ❌ Failed: ${results.failedDiscoveries}`);
    console.log(`   ⚡ Avg latency: ${this.calculateAverageLatency(results.discoveryLatencies)}ms`);

    return results;
  }

  /**
   * Test enhanced tools with discovered schemas
   */
  async runEnhancedToolsTests(): Promise<EnhancedToolsTestResults> {
    console.log('🛠️  Running enhanced tools tests...');
    
    const results: EnhancedToolsTestResults = {
      toolTests: [],
      schemaValidationTests: [],
      adaptiveQueryTests: [],
      errorHandlingTests: [],
    };

    for (const account of this.testAccounts) {
      if (!account.apiKey || account.id === 0) continue;

      console.log(`\n🧪 Testing enhanced tools for: ${account.name}`);

      // Test run_nrql_query with schema validation
      await this.testNrqlQueryTool(account, results);
      
      // Test search_entities
      await this.testSearchEntitiesTool(account, results);
      
      // Test discover_schemas
      await this.testDiscoverSchemasTool(account, results);
      
      // Test error handling scenarios
      await this.testErrorHandlingScenarios(account, results);
    }

    return results;
  }

  /**
   * Test adaptive dashboard generation
   */
  async runAdaptiveDashboardTests(): Promise<DashboardTestResults> {
    console.log('📊 Running adaptive dashboard tests...');
    
    const results: DashboardTestResults = {
      dashboardGenerationTests: [],
      widgetAdaptationTests: [],
      templateCompatibilityTests: [],
    };

    for (const account of this.testAccounts) {
      if (!account.apiKey || account.id === 0) continue;

      console.log(`\n📈 Testing dashboard generation for: ${account.name}`);

      // Test golden signals dashboard
      await this.testGoldenSignalsDashboard(account, results);
      
      // Test infrastructure dashboard
      await this.testInfrastructureDashboard(account, results);
      
      // Test widget adaptation
      await this.testWidgetAdaptation(account, results);
    }

    return results;
  }

  /**
   * Run performance and caching tests
   */
  async runPerformanceTests(): Promise<PerformanceTestResults> {
    console.log('⚡ Running performance and caching tests...');
    
    const results: PerformanceTestResults = {
      cacheEfficiencyTests: [],
      latencyBenchmarks: [],
      memoryUsageTests: [],
      concurrencyTests: [],
    };

    for (const account of this.testAccounts) {
      if (!account.apiKey || account.id === 0) continue;

      console.log(`\n🚀 Performance testing for: ${account.name}`);

      // Test cache efficiency
      await this.testCacheEfficiency(account, results);
      
      // Test concurrent discovery
      await this.testConcurrentDiscovery(account, results);
      
      // Test memory usage
      await this.testMemoryUsage(account, results);
    }

    return results;
  }

  // Helper methods for specific test implementations...
  private calculateSchemaAccuracy(discovered: string[], expected: string[]): number {
    if (expected.length === 0) return 1.0; // No expectations = 100% accuracy
    
    const foundExpected = expected.filter(schema => discovered.includes(schema));
    return foundExpected.length / expected.length;
  }

  private calculateAverageLatency(latencies: Array<{ accountName: string; latency: number }>): number {
    if (latencies.length === 0) return 0;
    return latencies.reduce((sum, l) => sum + l.latency, 0) / latencies.length;
  }

  // Enhanced Tools Test Implementations
  private async testNrqlQueryTool(account: TestAccount, results: EnhancedToolsTestResults): Promise<void> {
    console.log(`   🔍 Testing run_nrql_query tool for ${account.name}...`);
    
    try {
      // Test valid NRQL query
      const validQuery = `SELECT count(*) FROM Transaction SINCE 1 hour ago LIMIT 1`;
      const startTime = Date.now();
      
      // Simulate tool call (would be actual MCP call in real implementation)
      const nerdgraph = createNerdGraphClient({ apiKey: account.apiKey, region: account.region });
      const result = await nerdgraph.nrql(account.id, validQuery);
      
      const latency = Date.now() - startTime;
      
      results.toolTests.push({
        tool: 'run_nrql_query',
        accountName: account.name,
        testType: 'valid_query',
        success: true,
        latency,
        details: {
          query: validQuery,
          resultCount: result.results?.length || 0,
        },
      });

      // Test query with invalid syntax
      try {
        const invalidQuery = `SELECT invalid syntax FROM NonExistentEvent`;
        await nerdgraph.nrql(account.id, invalidQuery);
        
        results.toolTests.push({
          tool: 'run_nrql_query',
          accountName: account.name,
          testType: 'invalid_syntax',
          success: false,
          error: 'Expected syntax error but query succeeded',
        });
      } catch (error: any) {
        results.toolTests.push({
          tool: 'run_nrql_query',
          accountName: account.name,
          testType: 'invalid_syntax',
          success: true,
          details: {
            expectedError: 'Syntax error properly caught',
            errorMessage: error.message,
          },
        });
      }

      // Test schema validation with discovered event types
      const discoveredEventTypes = await this.discovery!.discoverEventTypes(account.id);
      if (discoveredEventTypes.length > 0) {
        const validEventType = discoveredEventTypes[0].name;
        const schemaValidQuery = `SELECT * FROM ${validEventType} SINCE 5 minutes ago LIMIT 1`;
        
        const validationResult = await nerdgraph.nrql(account.id, schemaValidQuery);
        
        results.schemaValidationTests.push({
          accountName: account.name,
          eventType: validEventType,
          query: schemaValidQuery,
          success: true,
          resultCount: validationResult.results?.length || 0,
        });
      }

      console.log(`     ✅ NRQL query tool tests completed for ${account.name}`);
      
    } catch (error: any) {
      console.error(`     ❌ NRQL query tool test failed for ${account.name}: ${error.message}`);
      results.toolTests.push({
        tool: 'run_nrql_query',
        accountName: account.name,
        testType: 'connection_test',
        success: false,
        error: error.message,
      });
    }
  }

  private async testSearchEntitiesTool(account: TestAccount, results: EnhancedToolsTestResults): Promise<void> {
    console.log(`   🔍 Testing search_entities tool for ${account.name}...`);
    
    try {
      // Test entity discovery with NerdGraph
      const nerdgraph = createNerdGraphClient({ apiKey: account.apiKey, region: account.region });
      
      // Search for APM applications
      const entitiesQuery = `
        {
          actor {
            entitySearch(queryBuilder: {domain: APM, type: APPLICATION}) {
              results {
                entities {
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
                }
              }
            }
          }
        }
      `;
      
      const startTime = Date.now();
      const entitiesResult = await nerdgraph.request(entitiesQuery);
      const latency = Date.now() - startTime;
      
      const entities = entitiesResult.data?.actor?.entitySearch?.results?.entities || [];
      
      results.toolTests.push({
        tool: 'search_entities',
        accountName: account.name,
        testType: 'apm_applications',
        success: true,
        latency,
        details: {
          entitiesFound: entities.length,
          sampleEntities: entities.slice(0, 3).map((e: any) => ({ guid: e.guid, name: e.name, type: e.type })),
        },
      });

      // Test entity type filtering
      if (entities.length > 0) {
        const firstEntity = entities[0];
        const entityDetailsQuery = `
          {
            actor {
              entity(guid: "${firstEntity.guid}") {
                guid
                name
                type
                domain
                entityType
                goldenMetrics {
                  metrics {
                    name
                    query
                  }
                }
              }
            }
          }
        `;
        
        const entityDetails = await nerdgraph.request(entityDetailsQuery);
        
        results.toolTests.push({
          tool: 'get_entity_details',
          accountName: account.name,
          testType: 'entity_details',
          success: true,
          details: {
            entityGuid: firstEntity.guid,
            hasGoldenMetrics: !!entityDetails.data?.actor?.entity?.goldenMetrics?.metrics?.length,
            goldenMetricsCount: entityDetails.data?.actor?.entity?.goldenMetrics?.metrics?.length || 0,
          },
        });
      }

      console.log(`     ✅ Entity search tool tests completed for ${account.name}`);
      
    } catch (error: any) {
      console.error(`     ❌ Entity search tool test failed for ${account.name}: ${error.message}`);
      results.toolTests.push({
        tool: 'search_entities',
        accountName: account.name,
        testType: 'connection_test',
        success: false,
        error: error.message,
      });
    }
  }

  private async testDiscoverSchemasTool(account: TestAccount, results: EnhancedToolsTestResults): Promise<void> {
    console.log(`   🔍 Testing discover_schemas tool for ${account.name}...`);
    
    try {
      const startTime = Date.now();
      
      // Test comprehensive schema discovery
      const [eventTypes, metrics, entities] = await Promise.all([
        this.discovery!.discoverEventTypes(account.id),
        this.discovery!.discoverMetricNames(account.id),
        this.discovery!.discoverEntities(account.id),
      ]);
      
      const discoveryLatency = Date.now() - startTime;
      
      // Test attribute discovery for top event types
      const attributeDiscovery = [];
      for (const eventType of eventTypes.slice(0, 3)) {
        const attributes = await this.discovery!.discoverAttributes(account.id, eventType.name);
        attributeDiscovery.push({
          eventType: eventType.name,
          attributeCount: attributes.length,
          sampleAttributes: attributes.slice(0, 5).map(a => a.name),
        });
      }
      
      results.toolTests.push({
        tool: 'discover_schemas',
        accountName: account.name,
        testType: 'comprehensive_discovery',
        success: true,
        latency: discoveryLatency,
        details: {
          eventTypesFound: eventTypes.length,
          metricsFound: metrics.length,
          entitiesFound: entities.length,
          attributeDiscovery,
          topEventTypes: eventTypes.slice(0, 5).map(et => et.name),
          topMetrics: metrics.slice(0, 5).map(m => m.name),
        },
      });

      // Test schema caching behavior
      const cacheStartTime = Date.now();
      const cachedEventTypes = await this.discovery!.discoverEventTypes(account.id);
      const cacheLatency = Date.now() - cacheStartTime;
      
      results.toolTests.push({
        tool: 'discover_schemas',
        accountName: account.name,
        testType: 'cache_performance',
        success: true,
        latency: cacheLatency,
        details: {
          originalLatency: discoveryLatency,
          cachedLatency: cacheLatency,
          speedImprovement: discoveryLatency / Math.max(cacheLatency, 1),
          cacheHit: cacheLatency < discoveryLatency * 0.5,
        },
      });

      console.log(`     ✅ Schema discovery tool tests completed for ${account.name}`);
      
    } catch (error: any) {
      console.error(`     ❌ Schema discovery tool test failed for ${account.name}: ${error.message}`);
      results.toolTests.push({
        tool: 'discover_schemas',
        accountName: account.name,
        testType: 'discovery_test',
        success: false,
        error: error.message,
      });
    }
  }

  private async testErrorHandlingScenarios(account: TestAccount, results: EnhancedToolsTestResults): Promise<void> {
    console.log(`   🔍 Testing error handling scenarios for ${account.name}...`);
    
    const nerdgraph = createNerdGraphClient({ apiKey: account.apiKey, region: account.region });
    
    // Test invalid account ID
    try {
      await nerdgraph.nrql(99999999, `SELECT count(*) FROM Transaction SINCE 1 hour ago`);
      results.errorHandlingTests.push({
        scenario: 'invalid_account_id',
        accountName: account.name,
        success: false,
        error: 'Expected error for invalid account ID but query succeeded',
      });
    } catch (error: any) {
      results.errorHandlingTests.push({
        scenario: 'invalid_account_id',
        accountName: account.name,
        success: true,
        details: {
          expectedBehavior: 'Error properly caught',
          errorMessage: error.message,
        },
      });
    }

    // Test rate limiting resilience
    try {
      const rapidQueries = Array.from({ length: 10 }, (_, i) => 
        nerdgraph.nrql(account.id, `SELECT count(*) FROM Transaction SINCE ${i + 1} minutes ago LIMIT 1`)
      );
      
      const results_array = await Promise.allSettled(rapidQueries);
      const successCount = results_array.filter(r => r.status === 'fulfilled').length;
      const errorCount = results_array.filter(r => r.status === 'rejected').length;
      
      results.errorHandlingTests.push({
        scenario: 'rate_limiting',
        accountName: account.name,
        success: true,
        details: {
          totalQueries: 10,
          successfulQueries: successCount,
          failedQueries: errorCount,
          rateLimitingObserved: errorCount > 0,
        },
      });
    } catch (error: any) {
      results.errorHandlingTests.push({
        scenario: 'rate_limiting',
        accountName: account.name,
        success: false,
        error: error.message,
      });
    }

    // Test malformed query handling
    const malformedQueries = [
      'SELECT FROM WHERE',
      'INVALID NRQL SYNTAX',
      'SELECT * FROM "NonExistent Event Type"',
      'SELECT count(*) FROM Transaction WHERE invalid operator value',
    ];

    for (const malformedQuery of malformedQueries) {
      try {
        await nerdgraph.nrql(account.id, malformedQuery);
        results.errorHandlingTests.push({
          scenario: 'malformed_query',
          accountName: account.name,
          success: false,
          error: `Expected error for malformed query: ${malformedQuery}`,
        });
      } catch (error: any) {
        results.errorHandlingTests.push({
          scenario: 'malformed_query',
          accountName: account.name,
          success: true,
          details: {
            query: malformedQuery,
            errorMessage: error.message,
          },
        });
      }
    }

    console.log(`     ✅ Error handling tests completed for ${account.name}`);
  }

  private async testGoldenSignalsDashboard(account: TestAccount, results: DashboardTestResults): Promise<void> {
    console.log(`   📊 Testing golden signals dashboard for ${account.name}...`);
    
    try {
      // First, find entities to test with
      const entities = await this.discovery!.discoverEntities(account.id);
      
      if (entities.length === 0) {
        results.dashboardGenerationTests.push({
          accountName: account.name,
          templateName: 'golden-signals',
          success: false,
          error: 'No entities found for dashboard generation',
        });
        return;
      }

      const testEntity = entities[0];
      
      // Test dashboard generation
      const startTime = Date.now();
      const dashboard = await this.dashboardGenerator!.generateDashboard(
        'golden-signals',
        testEntity,
        {
          accountId: account.id,
          timeRange: '1 hour ago',
        }
      );
      const generationLatency = Date.now() - startTime;

      // Validate dashboard structure
      const hasPages = dashboard.pages && dashboard.pages.length > 0;
      const hasWidgets = hasPages && dashboard.pages[0].widgets && dashboard.pages[0].widgets.length > 0;
      const widgetCount = hasWidgets ? dashboard.pages.reduce((sum, page) => sum + page.widgets.length, 0) : 0;
      
      // Validate widgets have valid NRQL queries
      const widgetValidation = [];
      if (hasWidgets) {
        for (const page of dashboard.pages) {
          for (const widget of page.widgets) {
            const hasValidQuery = widget.configuration?.nrqlQueries?.length > 0;
            const query = hasValidQuery ? widget.configuration.nrqlQueries[0].query : '';
            
            widgetValidation.push({
              title: widget.title,
              hasValidQuery,
              query: query.substring(0, 100) + (query.length > 100 ? '...' : ''),
              visualization: widget.visualization?.id,
            });
          }
        }
      }

      results.dashboardGenerationTests.push({
        accountName: account.name,
        templateName: 'golden-signals',
        success: true,
        latency: generationLatency,
        details: {
          entityGuid: testEntity.guid,
          entityName: testEntity.name,
          entityType: testEntity.type,
          dashboardName: dashboard.name,
          pageCount: dashboard.pages?.length || 0,
          widgetCount,
          widgetValidation,
        },
      });

      console.log(`     ✅ Golden signals dashboard generated for ${account.name} (${widgetCount} widgets)`);
      
    } catch (error: any) {
      console.error(`     ❌ Golden signals dashboard test failed for ${account.name}: ${error.message}`);
      results.dashboardGenerationTests.push({
        accountName: account.name,
        templateName: 'golden-signals',
        success: false,
        error: error.message,
      });
    }
  }

  private async testInfrastructureDashboard(account: TestAccount, results: DashboardTestResults): Promise<void> {
    console.log(`   📊 Testing infrastructure dashboard for ${account.name}...`);
    
    try {
      // Check if account has infrastructure data
      const eventTypes = await this.discovery!.discoverEventTypes(account.id);
      const hasInfraData = eventTypes.some(et => et.name.includes('SystemSample') || et.name.includes('ProcessSample'));
      
      if (!hasInfraData) {
        results.dashboardGenerationTests.push({
          accountName: account.name,
          templateName: 'infrastructure',
          success: false,
          error: 'No infrastructure data found (SystemSample, ProcessSample)',
        });
        return;
      }

      // Find a host entity for infrastructure dashboard
      const entities = await this.discovery!.discoverEntities(account.id);
      const hostEntity = entities.find(e => e.type === 'HOST') || entities[0];
      
      if (!hostEntity) {
        results.dashboardGenerationTests.push({
          accountName: account.name,
          templateName: 'infrastructure',
          success: false,
          error: 'No suitable entity found for infrastructure dashboard',
        });
        return;
      }

      // Test infrastructure dashboard generation
      const startTime = Date.now();
      const dashboard = await this.dashboardGenerator!.generateDashboard(
        'infrastructure',
        hostEntity,
        {
          accountId: account.id,
          timeRange: '2 hours ago',
        }
      );
      const generationLatency = Date.now() - startTime;

      // Validate infrastructure-specific widgets
      const infraWidgets = [];
      for (const page of dashboard.pages) {
        for (const widget of page.widgets) {
          const query = widget.configuration?.nrqlQueries?.[0]?.query || '';
          const isInfraWidget = query.includes('SystemSample') || 
                              query.includes('cpu') || 
                              query.includes('memory') ||
                              query.includes('disk');
                              
          infraWidgets.push({
            title: widget.title,
            isInfraspecific: isInfraWidget,
            query: query.substring(0, 80) + '...',
          });
        }
      }

      results.dashboardGenerationTests.push({
        accountName: account.name,
        templateName: 'infrastructure',
        success: true,
        latency: generationLatency,
        details: {
          entityGuid: hostEntity.guid,
          entityType: hostEntity.type,
          infraWidgetCount: infraWidgets.filter(w => w.isInfraspecific).length,
          totalWidgetCount: infraWidgets.length,
          infraWidgets,
        },
      });

      console.log(`     ✅ Infrastructure dashboard generated for ${account.name}`);
      
    } catch (error: any) {
      console.error(`     ❌ Infrastructure dashboard test failed for ${account.name}: ${error.message}`);
      results.dashboardGenerationTests.push({
        accountName: account.name,
        templateName: 'infrastructure',
        success: false,
        error: error.message,
      });
    }
  }

  private async testWidgetAdaptation(account: TestAccount, results: DashboardTestResults): Promise<void> {
    console.log(`   🎨 Testing widget adaptation for ${account.name}...`);
    
    try {
      const entities = await this.discovery!.discoverEntities(account.id);
      if (entities.length === 0) {
        results.widgetAdaptationTests.push({
          accountName: account.name,
          success: false,
          error: 'No entities available for widget adaptation testing',
        });
        return;
      }

      const testEntity = entities[0];
      
      // Discover entity data patterns
      const entityData = await this.discovery!.discoverEntityData(testEntity);
      
      // Test different widget intents and their adaptation
      const widgetIntents = ['error_rate', 'latency_p95', 'throughput', 'saturation_cpu'] as const;
      const adaptationResults = [];

      for (const intent of widgetIntents) {
        try {
          const template = {
            intent,
            title: `Test ${intent}`,
            visualization: 'viz.line',
            layout: { column: 1, row: 1, width: 4, height: 3 },
          };
          
          const startTime = Date.now();
          const widget = await (this.dashboardGenerator as any).adaptWidget(
            template,
            entityData,
            testEntity,
            { accountId: account.id, timeRange: '1 hour ago' }
          );
          const adaptationLatency = Date.now() - startTime;
          
          adaptationResults.push({
            intent,
            success: !!widget,
            latency: adaptationLatency,
            details: {
              widgetCreated: !!widget,
              hasQuery: !!widget?.configuration?.nrqlQueries?.length,
              query: widget?.configuration?.nrqlQueries?.[0]?.query?.substring(0, 100) || '',
              usedFields: this.extractFieldsFromQuery(widget?.configuration?.nrqlQueries?.[0]?.query || ''),
            },
          });
          
        } catch (error: any) {
          adaptationResults.push({
            intent,
            success: false,
            error: error.message,
          });
        }
      }

      // Test fallback mechanisms
      const errorRateTemplate = {
        intent: 'error_rate' as const,
        title: 'Error Rate',
        visualization: 'viz.line',
        layout: { column: 1, row: 1, width: 4, height: 3 },
        fallbacks: ['throughput', 'latency_p95'],
      };

      let fallbackTested = false;
      try {
        // Force error by providing empty entity data
        const emptyEntityData = {
          ...entityData,
          errorIndicators: [], // No error indicators to force fallback
        };
        
        const widget = await (this.dashboardGenerator as any).adaptWidgetsToSchema(
          [errorRateTemplate],
          emptyEntityData,
          testEntity,
          { accountId: account.id, timeRange: '1 hour ago' }
        );
        
        fallbackTested = widget.length > 0;
      } catch (error) {
        // Expected if no fallbacks work
      }

      results.widgetAdaptationTests.push({
        accountName: account.name,
        success: true,
        details: {
          entityGuid: testEntity.guid,
          entityType: testEntity.type,
          discoveredPatterns: {
            eventTypes: entityData.eventTypes.length,
            errorIndicators: entityData.errorIndicators.length,
            durationFields: entityData.durationFields.length,
            metrics: entityData.metrics.length,
          },
          adaptationResults,
          fallbackMechanismTested: fallbackTested,
        },
      });

      console.log(`     ✅ Widget adaptation tests completed for ${account.name}`);
      
    } catch (error: any) {
      console.error(`     ❌ Widget adaptation test failed for ${account.name}: ${error.message}`);
      results.widgetAdaptationTests.push({
        accountName: account.name,
        success: false,
        error: error.message,
      });
    }
  }

  // Helper method to extract field names from NRQL queries
  private extractFieldsFromQuery(query: string): string[] {
    const fields: string[] = [];
    
    // Simple field extraction - could be made more sophisticated
    const selectMatch = query.match(/SELECT\s+([^FROM]+)/i);
    if (selectMatch) {
      const selectClause = selectMatch[1];
      // Extract fields mentioned in functions like percentile(duration, 95)
      const fieldMatches = selectClause.match(/\w+\([^)]*(\w+)[^)]*\)/g);
      if (fieldMatches) {
        fields.push(...fieldMatches);
      }
    }
    
    // Extract WHERE clause fields
    const whereMatch = query.match(/WHERE\s+([^SINCE]+)/i);
    if (whereMatch) {
      const whereClause = whereMatch[1];
      const fieldMatches = whereClause.match(/(\w+)\s*[=<>]/g);
      if (fieldMatches) {
        fields.push(...fieldMatches.map(f => f.replace(/\s*[=<>].*/, '')));
      }
    }
    
    return Array.from(new Set(fields)); // Remove duplicates
  }

  private async testCacheEfficiency(account: TestAccount, results: PerformanceTestResults): Promise<void> {
    console.log(`   🚀 Testing cache efficiency for ${account.name}...`);
    
    try {
      // Test cache warm-up and efficiency
      const cacheTestResults = [];
      
      // First run (cold cache)
      const coldStartTime = Date.now();
      const coldEventTypes = await this.discovery!.discoverEventTypes(account.id);
      const coldLatency = Date.now() - coldStartTime;
      
      // Second run (warm cache)
      const warmStartTime = Date.now();
      const warmEventTypes = await this.discovery!.discoverEventTypes(account.id);
      const warmLatency = Date.now() - warmStartTime;
      
      // Third run (cache hit)
      const cacheStartTime = Date.now();
      const cachedEventTypes = await this.discovery!.discoverEventTypes(account.id);
      const cacheLatency = Date.now() - cacheStartTime;
      
      const cacheEfficiency = {
        coldLatency,
        warmLatency,
        cacheLatency,
        speedImprovement: coldLatency / Math.max(cacheLatency, 1),
        cacheHitRatio: cacheLatency < coldLatency * 0.1 ? 100 : 
                      cacheLatency < coldLatency * 0.5 ? 80 : 60,
      };
      
      // Test attribute discovery caching
      if (coldEventTypes.length > 0) {
        const eventType = coldEventTypes[0].name;
        
        const attrColdStart = Date.now();
        const coldAttributes = await this.discovery!.discoverAttributes(account.id, eventType);
        const attrColdLatency = Date.now() - attrColdStart;
        
        const attrWarmStart = Date.now();
        const warmAttributes = await this.discovery!.discoverAttributes(account.id, eventType);
        const attrWarmLatency = Date.now() - attrWarmStart;
        
        (cacheEfficiency as any).attributeCaching = {
          coldLatency: attrColdLatency,
          warmLatency: attrWarmLatency,
          speedImprovement: attrColdLatency / Math.max(attrWarmLatency, 1),
          attributeCount: coldAttributes.length,
        };
      }

      results.cacheEfficiencyTests.push({
        accountName: account.name,
        success: true,
        details: cacheEfficiency,
      });

      console.log(`     ✅ Cache efficiency: ${cacheEfficiency.speedImprovement.toFixed(1)}x improvement`);
      
    } catch (error: any) {
      console.error(`     ❌ Cache efficiency test failed for ${account.name}: ${error.message}`);
      results.cacheEfficiencyTests.push({
        accountName: account.name,
        success: false,
        error: error.message,
      });
    }
  }

  private async testConcurrentDiscovery(account: TestAccount, results: PerformanceTestResults): Promise<void> {
    console.log(`   ⚡ Testing concurrent discovery for ${account.name}...`);
    
    try {
      // Test concurrent discovery operations
      const concurrencyLevels = [1, 3, 5];
      const concurrencyResults = [];
      
      for (const concurrency of concurrencyLevels) {
        console.log(`     Testing concurrency level: ${concurrency}`);
        
        const tasks = Array.from({ length: concurrency }, () => 
          this.discovery!.discoverEventTypes(account.id)
        );
        
        const startTime = Date.now();
        const results_array = await Promise.allSettled(tasks);
        const totalLatency = Date.now() - startTime;
        
        const successCount = results_array.filter(r => r.status === 'fulfilled').length;
        const failureCount = results_array.filter(r => r.status === 'rejected').length;
        
        concurrencyResults.push({
          concurrencyLevel: concurrency,
          totalLatency,
          averageLatencyPerTask: totalLatency / concurrency,
          successfulTasks: successCount,
          failedTasks: failureCount,
          successRate: (successCount / concurrency) * 100,
        });
      }
      
      // Test mixed operation concurrency
      const mixedOperationsStart = Date.now();
      const mixedTasks = await Promise.allSettled([
        this.discovery!.discoverEventTypes(account.id),
        this.discovery!.discoverMetricNames(account.id),
        this.discovery!.discoverEntities(account.id),
      ]);
      const mixedOperationsLatency = Date.now() - mixedOperationsStart;
      
      const mixedSuccessCount = mixedTasks.filter(r => r.status === 'fulfilled').length;

      results.concurrencyTests.push({
        accountName: account.name,
        success: true,
        details: {
          concurrencyResults,
          mixedOperations: {
            totalLatency: mixedOperationsLatency,
            successfulOperations: mixedSuccessCount,
            totalOperations: 3,
            operationTypes: ['eventTypes', 'metrics', 'entities'],
          },
        },
      });

      console.log(`     ✅ Concurrent discovery completed - max concurrency: ${Math.max(...concurrencyLevels)}`);
      
    } catch (error: any) {
      console.error(`     ❌ Concurrent discovery test failed for ${account.name}: ${error.message}`);
      results.concurrencyTests.push({
        accountName: account.name,
        success: false,
        error: error.message,
      });
    }
  }

  private async testMemoryUsage(account: TestAccount, results: PerformanceTestResults): Promise<void> {
    console.log(`   💾 Testing memory usage for ${account.name}...`);
    
    try {
      // Get initial memory usage
      const initialMemory = process.memoryUsage();
      
      // Perform discovery operations and track memory
      const memorySnapshots = [{ 
        stage: 'initial', 
        ...initialMemory,
        heapUsedMB: Math.round(initialMemory.heapUsed / 1024 / 1024),
      }];
      
      // Event type discovery
      await this.discovery!.discoverEventTypes(account.id);
      const afterEventTypes = process.memoryUsage();
      memorySnapshots.push({
        stage: 'after_event_types',
        ...afterEventTypes,
        heapUsedMB: Math.round(afterEventTypes.heapUsed / 1024 / 1024),
      });
      
      // Attribute discovery for top event types
      const eventTypes = await this.discovery!.discoverEventTypes(account.id);
      for (const eventType of eventTypes.slice(0, 3)) {
        await this.discovery!.discoverAttributes(account.id, eventType.name);
      }
      const afterAttributes = process.memoryUsage();
      memorySnapshots.push({
        stage: 'after_attributes',
        ...afterAttributes,
        heapUsedMB: Math.round(afterAttributes.heapUsed / 1024 / 1024),
      });
      
      // Metric discovery
      await this.discovery!.discoverMetricNames(account.id);
      const afterMetrics = process.memoryUsage();
      memorySnapshots.push({
        stage: 'after_metrics',
        ...afterMetrics,
        heapUsedMB: Math.round(afterMetrics.heapUsed / 1024 / 1024),
      });
      
      // Entity discovery
      await this.discovery!.discoverEntities(account.id);
      const afterEntities = process.memoryUsage();
      memorySnapshots.push({
        stage: 'after_entities',
        ...afterEntities,
        heapUsedMB: Math.round(afterEntities.heapUsed / 1024 / 1024),
      });
      
      // Force garbage collection if available
      if (global.gc) {
        global.gc();
        const afterGC = process.memoryUsage();
        memorySnapshots.push({
          stage: 'after_gc',
          ...afterGC,
          heapUsedMB: Math.round(afterGC.heapUsed / 1024 / 1024),
        });
      }
      
      // Calculate memory growth
      const maxHeapUsed = Math.max(...memorySnapshots.map(s => s.heapUsedMB));
      const memoryGrowth = maxHeapUsed - memorySnapshots[0].heapUsedMB;
      
      // Test memory leak detection (simplified)
      const memoryStabilityTest = [];
      for (let i = 0; i < 5; i++) {
        await this.discovery!.discoverEventTypes(account.id);
        const snapshot = process.memoryUsage();
        memoryStabilityTest.push(Math.round(snapshot.heapUsed / 1024 / 1024));
      }
      
      const memoryVariance = Math.max(...memoryStabilityTest) - Math.min(...memoryStabilityTest);
      const memoryStable = memoryVariance < 10; // Less than 10MB variance
      
      results.memoryUsageTests.push({
        accountName: account.name,
        success: true,
        details: {
          memorySnapshots,
          maxHeapUsedMB: maxHeapUsed,
          memoryGrowthMB: memoryGrowth,
          memoryStable,
          memoryVarianceMB: memoryVariance,
          memoryStabilityTest,
        },
      });

      console.log(`     ✅ Memory usage: ${maxHeapUsed}MB peak, ${memoryGrowth}MB growth`);
      
    } catch (error: any) {
      console.error(`     ❌ Memory usage test failed for ${account.name}: ${error.message}`);
      results.memoryUsageTests.push({
        accountName: account.name,
        success: false,
        error: error.message,
      });
    }
  }
}

// Test result interfaces
export interface DiscoveryTestResults {
  totalAccounts: number;
  successfulDiscoveries: number;
  failedDiscoveries: number;
  discoveryLatencies: Array<{ accountName: string; latency: number }>;
  schemaAccuracy: Array<{ accountName: string; expected: string[]; discovered: string[]; accuracy: number }>;
  attributeCompleteness: Array<{ accountName: string; eventTypeResults: any[] }>;
  serviceIdentifierAccuracy: Array<{ accountName: string; expected?: string; discovered: string; correct: boolean }>;
  errorIndicatorDetection: Array<{ accountName: string; indicatorsFound: number; indicators: any[] }>;
  metricDetection: Array<{ accountName: string; metricsFound: number; topMetrics: string[] }>;
}

export interface EnhancedToolsTestResults {
  toolTests: any[];
  schemaValidationTests: any[];
  adaptiveQueryTests: any[];
  errorHandlingTests: any[];
}

export interface DashboardTestResults {
  dashboardGenerationTests: any[];
  widgetAdaptationTests: any[];
  templateCompatibilityTests: any[];
}

export interface PerformanceTestResults {
  cacheEfficiencyTests: any[];
  latencyBenchmarks: any[];
  memoryUsageTests: any[];
  concurrencyTests: any[];
}
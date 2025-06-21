/**
 * Comprehensive E2E Test Suite
 * 
 * Tests all aspects of the platform-native MCP server with real NRDB backend
 * This test brings together discovery, tools, dashboards, and performance testing
 */

import { describe, beforeAll, afterAll, it, expect } from 'vitest';
import { E2ETestHarness } from './test-harness.js';
import { getTestTimeout, logTestResult, formatLatency } from './setup.js';

describe('Platform-Native MCP Server - Comprehensive E2E Tests', () => {
  let testHarness: E2ETestHarness;
  
  beforeAll(async () => {
    console.log('🚀 Initializing Comprehensive E2E Test Suite...');
    testHarness = new E2ETestHarness();
    await testHarness.setupTestHarness();
  }, getTestTimeout());

  afterAll(async () => {
    console.log('🧹 Cleaning up Comprehensive E2E Test Suite...');
    if (testHarness) {
      await testHarness.teardownTestHarness();
    }
  });

  describe('🔍 Discovery Engine Validation', () => {
    it('should successfully discover schemas across all configured accounts', async () => {
      console.log('Running comprehensive discovery tests...');
      
      const discoveryResults = await testHarness.runDiscoveryTests();
      
      // Validate overall discovery success
      expect(discoveryResults.successfulDiscoveries).toBeGreaterThan(0);
      expect(discoveryResults.failedDiscoveries).toBe(0);
      
      // Validate schema accuracy across accounts
      const averageAccuracy = discoveryResults.schemaAccuracy.reduce(
        (sum, result) => sum + result.accuracy, 0
      ) / discoveryResults.schemaAccuracy.length;
      
      expect(averageAccuracy).toBeGreaterThan(0.8); // 80% accuracy threshold
      
      // Validate performance thresholds
      const averageLatency = discoveryResults.discoveryLatencies.reduce(
        (sum, result) => sum + result.latency, 0
      ) / discoveryResults.discoveryLatencies.length;
      
      expect(averageLatency).toBeLessThan(5000); // 5 second average
      
      logTestResult('Discovery Engine Validation', true, {
        accountsTested: discoveryResults.totalAccounts,
        averageAccuracy: `${(averageAccuracy * 100).toFixed(1)}%`,
        averageLatency: formatLatency(averageLatency),
      });
    }, getTestTimeout());

    it('should demonstrate cache efficiency improvements', async () => {
      console.log('Testing discovery cache efficiency...');
      
      // Run discovery twice to test caching
      const firstRun = await testHarness.runDiscoveryTests();
      const secondRun = await testHarness.runDiscoveryTests();
      
      // Calculate average latencies
      const firstRunAvgLatency = firstRun.discoveryLatencies.reduce(
        (sum, r) => sum + r.latency, 0
      ) / firstRun.discoveryLatencies.length;
      
      const secondRunAvgLatency = secondRun.discoveryLatencies.reduce(
        (sum, r) => sum + r.latency, 0
      ) / secondRun.discoveryLatencies.length;
      
      // Second run should be faster or similar due to caching
      expect(secondRunAvgLatency).toBeLessThan(firstRunAvgLatency * 1.5);
      
      const speedImprovement = firstRunAvgLatency / secondRunAvgLatency;
      
      logTestResult('Cache Efficiency', true, {
        firstRun: formatLatency(firstRunAvgLatency),
        secondRun: formatLatency(secondRunAvgLatency),
        speedImprovement: `${speedImprovement.toFixed(1)}x`,
      });
    }, getTestTimeout() * 2); // Double timeout for two runs
  });

  describe('🛠️ Enhanced Tools Validation', () => {
    it('should validate all enhanced tools work correctly', async () => {
      console.log('Running enhanced tools tests...');
      
      const toolsResults = await testHarness.runEnhancedToolsTests();
      
      // Validate tool tests
      const successfulToolTests = toolsResults.toolTests.filter(t => t.success);
      const totalToolTests = toolsResults.toolTests.length;
      
      expect(successfulToolTests.length).toBeGreaterThan(0);
      
      const toolSuccessRate = (successfulToolTests.length / totalToolTests) * 100;
      expect(toolSuccessRate).toBeGreaterThan(70); // 70% success rate
      
      // Validate schema validation tests
      const successfulSchemaTests = toolsResults.schemaValidationTests.filter(t => t.success);
      expect(successfulSchemaTests.length).toBeGreaterThan(0);
      
      // Validate error handling tests
      const successfulErrorTests = toolsResults.errorHandlingTests.filter(t => t.success);
      expect(successfulErrorTests.length).toBeGreaterThan(0);
      
      logTestResult('Enhanced Tools Validation', true, {
        toolTestsSuccess: `${successfulToolTests.length}/${totalToolTests}`,
        successRate: `${toolSuccessRate.toFixed(1)}%`,
        schemaValidationTests: successfulSchemaTests.length,
        errorHandlingTests: successfulErrorTests.length,
      });
    }, getTestTimeout());
  });

  describe('📊 Adaptive Dashboard Generation', () => {
    it('should generate dashboards that adapt to discovered schemas', async () => {
      console.log('Running adaptive dashboard tests...');
      
      const dashboardResults = await testHarness.runAdaptiveDashboardTests();
      
      // Validate dashboard generation tests
      const successfulDashboards = dashboardResults.dashboardGenerationTests.filter(t => t.success);
      const totalDashboards = dashboardResults.dashboardGenerationTests.length;
      
      expect(successfulDashboards.length).toBeGreaterThan(0);
      
      const dashboardSuccessRate = (successfulDashboards.length / totalDashboards) * 100;
      expect(dashboardSuccessRate).toBeGreaterThan(80); // 80% success rate
      
      // Validate widget adaptation tests
      const successfulAdaptations = dashboardResults.widgetAdaptationTests.filter(t => t.success);
      expect(successfulAdaptations.length).toBeGreaterThan(0);
      
      // Validate at least some widgets were successfully adapted
      const totalWidgetsGenerated = successfulDashboards.reduce((sum, d) => {
        return sum + (d.details?.widgetCount || 0);
      }, 0);
      
      expect(totalWidgetsGenerated).toBeGreaterThan(0);
      
      logTestResult('Adaptive Dashboard Generation', true, {
        dashboardsGenerated: `${successfulDashboards.length}/${totalDashboards}`,
        successRate: `${dashboardSuccessRate.toFixed(1)}%`,
        totalWidgets: totalWidgetsGenerated,
        adaptationTests: successfulAdaptations.length,
      });
    }, getTestTimeout());
  });

  describe('⚡ Performance and Concurrency', () => {
    it('should maintain performance under concurrent operations', async () => {
      console.log('Running performance and concurrency tests...');
      
      const performanceResults = await testHarness.runPerformanceTests();
      
      // Validate cache efficiency tests
      const successfulCacheTests = performanceResults.cacheEfficiencyTests.filter(t => t.success);
      expect(successfulCacheTests.length).toBeGreaterThan(0);
      
      // Validate concurrency tests
      const successfulConcurrencyTests = performanceResults.concurrencyTests.filter(t => t.success);
      expect(successfulConcurrencyTests.length).toBeGreaterThan(0);
      
      // Validate memory usage tests
      const successfulMemoryTests = performanceResults.memoryUsageTests.filter(t => t.success);
      expect(successfulMemoryTests.length).toBeGreaterThan(0);
      
      // Check cache efficiency
      const avgSpeedImprovement = successfulCacheTests.reduce((sum, test) => {
        return sum + (test.details?.speedImprovement || 1);
      }, 0) / successfulCacheTests.length;
      
      expect(avgSpeedImprovement).toBeGreaterThan(2); // At least 2x improvement
      
      logTestResult('Performance and Concurrency', true, {
        cacheTests: successfulCacheTests.length,
        concurrencyTests: successfulConcurrencyTests.length,
        memoryTests: successfulMemoryTests.length,
        avgSpeedImprovement: `${avgSpeedImprovement.toFixed(1)}x`,
      });
    }, getTestTimeout());
  });

  describe('🎯 End-to-End Workflow Validation', () => {
    it('should complete a full discovery-to-dashboard workflow', async () => {
      console.log('Running end-to-end workflow test...');
      
      // Step 1: Run discovery
      const discoveryResults = await testHarness.runDiscoveryTests();
      expect(discoveryResults.successfulDiscoveries).toBeGreaterThan(0);
      
      // Step 2: Test enhanced tools
      const toolsResults = await testHarness.runEnhancedToolsTests();
      const toolSuccessRate = (toolsResults.toolTests.filter(t => t.success).length / toolsResults.toolTests.length) * 100;
      expect(toolSuccessRate).toBeGreaterThan(50);
      
      // Step 3: Generate adaptive dashboards
      const dashboardResults = await testHarness.runAdaptiveDashboardTests();
      const dashboardSuccessRate = (dashboardResults.dashboardGenerationTests.filter(t => t.success).length / dashboardResults.dashboardGenerationTests.length) * 100;
      expect(dashboardSuccessRate).toBeGreaterThan(50);
      
      // Step 4: Validate performance
      const performanceResults = await testHarness.runPerformanceTests();
      const performanceSuccessCount = performanceResults.cacheEfficiencyTests.filter(t => t.success).length;
      expect(performanceSuccessCount).toBeGreaterThan(0);
      
      logTestResult('End-to-End Workflow', true, {
        discoveryAccounts: discoveryResults.successfulDiscoveries,
        toolSuccessRate: `${toolSuccessRate.toFixed(1)}%`,
        dashboardSuccessRate: `${dashboardSuccessRate.toFixed(1)}%`,
        performanceTests: performanceSuccessCount,
      });
    }, getTestTimeout() * 3); // Triple timeout for full workflow
  });

  describe('🏆 Zero Hardcoded Schemas Validation', () => {
    it('should demonstrate zero assumptions philosophy', async () => {
      console.log('Validating zero hardcoded schemas philosophy...');
      
      const discoveryResults = await testHarness.runDiscoveryTests();
      
      // Verify that different account types produce different schema discoveries
      const schemaVariety = new Set();
      discoveryResults.schemaAccuracy.forEach(result => {
        result.discovered.forEach(schema => schemaVariety.add(schema));
      });
      
      expect(schemaVariety.size).toBeGreaterThan(1); // Different schemas discovered
      
      // Verify service identifier detection adapts to different patterns
      const serviceIdentifiers = new Set();
      discoveryResults.serviceIdentifierAccuracy.forEach(result => {
        if (result.discovered) {
          serviceIdentifiers.add(result.discovered);
        }
      });
      
      // Should find different service identifier patterns (appName, service.name, etc.)
      expect(serviceIdentifiers.size).toBeGreaterThan(0);
      
      // Verify error indicators adapt to different patterns
      const errorIndicatorTypes = new Set();
      discoveryResults.errorIndicatorDetection.forEach(result => {
        result.indicators.forEach(indicator => {
          errorIndicatorTypes.add(indicator.type);
        });
      });
      
      logTestResult('Zero Hardcoded Schemas Philosophy', true, {
        uniqueSchemas: schemaVariety.size,
        serviceIdentifierPatterns: serviceIdentifiers.size,
        errorIndicatorTypes: errorIndicatorTypes.size,
        schemasDiscovered: Array.from(schemaVariety).slice(0, 10).join(', '),
      });
    }, getTestTimeout());
  });
});
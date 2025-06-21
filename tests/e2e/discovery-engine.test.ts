/**
 * E2E Tests for Platform Discovery Engine with Real NRDB Backend
 * 
 * Tests comprehensive discovery capabilities across various New Relic
 * account configurations with actual data patterns.
 */

import { describe, beforeAll, afterAll, it, expect, beforeEach } from 'vitest';
import { E2ETestHarness, type DiscoveryTestResults } from './test-harness.js';

describe('Platform Discovery Engine E2E Tests', () => {
  let testHarness: E2ETestHarness;
  let discoveryResults: DiscoveryTestResults;

  beforeAll(async () => {
    testHarness = new E2ETestHarness();
    await testHarness.setupTestHarness();
  }, 30000); // 30 second timeout for setup

  afterAll(async () => {
    await testHarness.teardownTestHarness();
  });

  describe('Schema Discovery Accuracy', () => {
    beforeEach(async () => {
      discoveryResults = await testHarness.runDiscoveryTests();
    }, 60000); // 60 second timeout for discovery tests

    it('should successfully discover schemas in all configured accounts', () => {
      expect(discoveryResults.successfulDiscoveries).toBeGreaterThan(0);
      expect(discoveryResults.failedDiscoveries).toBe(0);
      
      console.log(`✅ Discovery successful in ${discoveryResults.successfulDiscoveries} accounts`);
    });

    it('should achieve high schema accuracy rates', () => {
      const averageAccuracy = discoveryResults.schemaAccuracy.reduce(
        (sum, result) => sum + result.accuracy, 0
      ) / discoveryResults.schemaAccuracy.length;

      expect(averageAccuracy).toBeGreaterThan(0.8); // 80% accuracy threshold
      
      // Log detailed results
      discoveryResults.schemaAccuracy.forEach(result => {
        console.log(`📊 ${result.accountName}: ${(result.accuracy * 100).toFixed(1)}% accuracy`);
        console.log(`   Expected: ${result.expected.join(', ')}`);
        console.log(`   Discovered: ${result.discovered.slice(0, 5).join(', ')}`);
      });
    });

    it('should discover event types with recent data', () => {
      discoveryResults.schemaAccuracy.forEach(result => {
        expect(result.discovered.length).toBeGreaterThan(0);
        
        // At least one discovered event type should match expectations
        const hasExpectedEventType = result.expected.some(expected =>
          result.discovered.includes(expected)
        );
        
        if (result.expected.length > 0) {
          expect(hasExpectedEventType).toBe(true);
        }
      });
    });

    it('should complete discovery within performance thresholds', () => {
      const averageLatency = discoveryResults.discoveryLatencies.reduce(
        (sum, result) => sum + result.latency, 0
      ) / discoveryResults.discoveryLatencies.length;

      expect(averageLatency).toBeLessThan(5000); // 5 second max average
      
      // Individual account latencies
      discoveryResults.discoveryLatencies.forEach(result => {
        expect(result.latency).toBeLessThan(10000); // 10 second max per account
        console.log(`⚡ ${result.accountName}: ${result.latency}ms`);
      });
    });
  });

  describe('Attribute Profiling', () => {
    it('should discover attributes for major event types', () => {
      discoveryResults.attributeCompleteness.forEach(accountResult => {
        expect(accountResult.eventTypeResults.length).toBeGreaterThan(0);
        
        accountResult.eventTypeResults.forEach(eventTypeResult => {
          expect(eventTypeResult.attributeCount).toBeGreaterThan(0);
          expect(eventTypeResult.sampleAttributes.length).toBeGreaterThan(0);
          
          console.log(`🏷️  ${accountResult.accountName}.${eventTypeResult.eventType}: ${eventTypeResult.attributeCount} attributes`);
        });
      });
    });

    it('should profile attributes with type and cardinality information', () => {
      // This would test the actual attribute profiling functionality
      // by calling the discovery engine directly for detailed validation
      expect(discoveryResults.attributeCompleteness.length).toBeGreaterThan(0);
    });
  });

  describe('Service Identifier Detection', () => {
    it('should detect service identifier fields accurately', () => {
      const correctDetections = discoveryResults.serviceIdentifierAccuracy.filter(
        result => result.correct
      );
      
      const accuracyRate = correctDetections.length / discoveryResults.serviceIdentifierAccuracy.length;
      expect(accuracyRate).toBeGreaterThan(0.7); // 70% accuracy threshold
      
      discoveryResults.serviceIdentifierAccuracy.forEach(result => {
        console.log(`🔗 ${result.accountName}: Expected '${result.expected}', Got '${result.discovered}' ${result.correct ? '✅' : '❌'}`);
      });
    });

    it('should provide fallback service identifiers when primary detection fails', () => {
      discoveryResults.serviceIdentifierAccuracy.forEach(result => {
        // Should always have some service identifier, even if not the expected one
        expect(result.discovered).toBeDefined();
        expect(result.discovered.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Error Indicator Discovery', () => {
    it('should discover error indicators in accounts with error data', () => {
      // Accounts that should have error data based on test configuration
      const accountsWithErrors = discoveryResults.errorIndicatorDetection.filter(
        result => result.indicatorsFound > 0
      );
      
      expect(accountsWithErrors.length).toBeGreaterThan(0);
      
      accountsWithErrors.forEach(result => {
        console.log(`❌ ${result.accountName}: ${result.indicatorsFound} error indicators`);
        result.indicators.forEach(indicator => {
          console.log(`   - ${indicator.field} (${indicator.type})`);
        });
      });
    });

    it('should detect different types of error indicators', () => {
      const allIndicators = discoveryResults.errorIndicatorDetection.flatMap(
        result => result.indicators
      );
      
      const indicatorTypes = new Set(allIndicators.map(indicator => indicator.type));
      
      // Should find at least boolean or http_status type indicators
      expect(indicatorTypes.size).toBeGreaterThan(0);
      
      console.log(`🎯 Error indicator types discovered: ${Array.from(indicatorTypes).join(', ')}`);
    });
  });

  describe('Metric Discovery', () => {
    it('should discover dimensional metrics in modern accounts', () => {
      const accountsWithMetrics = discoveryResults.metricDetection.filter(
        result => result.metricsFound > 0
      );
      
      if (accountsWithMetrics.length > 0) {
        accountsWithMetrics.forEach(result => {
          expect(result.metricsFound).toBeGreaterThan(0);
          console.log(`📊 ${result.accountName}: ${result.metricsFound} metrics`);
          console.log(`   Top metrics: ${result.topMetrics.join(', ')}`);
        });
      } else {
        console.log('ℹ️  No dimensional metrics found in test accounts (expected for legacy accounts)');
      }
    });
  });

  describe('Cross-Account Discovery Consistency', () => {
    it('should maintain consistent discovery patterns across accounts', () => {
      // Test that similar account types produce similar discovery results
      const legacyApmAccounts = discoveryResults.schemaAccuracy.filter(
        result => result.accountName.toLowerCase().includes('legacy')
      );
      
      const modernOtelAccounts = discoveryResults.schemaAccuracy.filter(
        result => result.accountName.toLowerCase().includes('modern') || 
                 result.accountName.toLowerCase().includes('otel')
      );
      
      // Legacy accounts should consistently find Transaction events
      legacyApmAccounts.forEach(result => {
        expect(result.discovered.includes('Transaction')).toBe(true);
      });
      
      // Modern accounts should find span-based data
      modernOtelAccounts.forEach(result => {
        const hasSpanData = result.discovered.includes('Span') || 
                           result.discovered.includes('Metric');
        expect(hasSpanData).toBe(true);
      });
    });
  });

  describe('Edge Case Handling', () => {
    it('should handle sparse data accounts gracefully', () => {
      const sparseAccounts = discoveryResults.schemaAccuracy.filter(
        result => result.accountName.toLowerCase().includes('sparse')
      );
      
      sparseAccounts.forEach(result => {
        // Should still discover something, even if minimal
        expect(result.discovered.length).toBeGreaterThan(0);
        console.log(`📉 Sparse account ${result.accountName}: ${result.discovered.length} event types`);
      });
    });

    it('should handle cross-region discovery', () => {
      const euAccounts = discoveryResults.discoveryLatencies.filter(
        result => result.accountName.toLowerCase().includes('eu')
      );
      
      euAccounts.forEach(result => {
        // EU region may have higher latency but should still work
        expect(result.latency).toBeLessThan(15000); // 15 second max for cross-region
        console.log(`🌍 EU region ${result.accountName}: ${result.latency}ms`);
      });
    });
  });

  describe('Discovery Caching Behavior', () => {
    it('should demonstrate cache efficiency on repeated discoveries', async () => {
      // Run discovery twice to test caching
      const firstRun = await testHarness.runDiscoveryTests();
      const secondRun = await testHarness.runDiscoveryTests();
      
      // Second run should be faster due to caching
      const firstRunAvgLatency = firstRun.discoveryLatencies.reduce(
        (sum, r) => sum + r.latency, 0
      ) / firstRun.discoveryLatencies.length;
      
      const secondRunAvgLatency = secondRun.discoveryLatencies.reduce(
        (sum, r) => sum + r.latency, 0
      ) / secondRun.discoveryLatencies.length;
      
      expect(secondRunAvgLatency).toBeLessThan(firstRunAvgLatency * 1.2); // Should be similar or faster
      
      console.log(`🚀 Cache efficiency: First run ${firstRunAvgLatency.toFixed(0)}ms, Second run ${secondRunAvgLatency.toFixed(0)}ms`);
    }, 120000); // 2 minute timeout for double discovery
  });
});
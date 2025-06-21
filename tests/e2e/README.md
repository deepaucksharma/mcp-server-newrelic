# End-to-End Testing Framework

This comprehensive E2E testing framework validates the platform-native MCP server with real New Relic Database (NRDB) backend integration.

## 🎯 Testing Philosophy

The E2E test harness embodies the **Zero Hardcoded Schemas** philosophy by:
- Testing discovery across diverse New Relic account configurations
- Validating adaptive behavior without assumptions about data patterns
- Verifying tool enhancement and dashboard generation accuracy
- Ensuring performance and caching efficiency

## 🏗️ Architecture Overview

```
tests/e2e/
├── test-harness.ts          # Main test harness with comprehensive scenarios
├── discovery-engine.test.ts # Discovery engine validation tests  
├── enhanced-tools.test.ts   # Enhanced tools functionality tests
├── dashboard-generation.test.ts # Adaptive dashboard generation tests
├── performance.test.ts      # Performance and caching tests
├── fixtures/               # Test data and mock responses
└── README.md              # This documentation
```

## 🔧 Test Environment Setup

### Required Environment Variables

```bash
# Primary test accounts (required)
E2E_ACCOUNT_LEGACY_APM=12345678      # Legacy APM account ID
E2E_API_KEY_LEGACY=NRAK-XXXXX       # API key for legacy account

E2E_ACCOUNT_MODERN_OTEL=87654321     # Modern OpenTelemetry account ID  
E2E_API_KEY_OTEL=NRAK-XXXXX         # API key for modern account

E2E_ACCOUNT_MIXED_DATA=11223344      # Mixed telemetry patterns account
E2E_API_KEY_MIXED=NRAK-XXXXX        # API key for mixed account

# Optional test accounts
E2E_ACCOUNT_SPARSE_DATA=55667788     # Sparse data account (optional)
E2E_API_KEY_SPARSE=NRAK-XXXXX       # API key for sparse account

E2E_ACCOUNT_EU_REGION=99887766       # EU region account (optional)
E2E_API_KEY_EU=NRAK-XXXXX           # API key for EU account

# Test configuration
E2E_TIMEOUT_MS=120000                # Test timeout (default: 2 minutes)
E2E_PARALLEL_ACCOUNTS=3              # Max parallel account testing
E2E_CACHE_TTL_SECONDS=3600          # Cache TTL for testing
```

### Account Data Requirements

For comprehensive testing, configure accounts with these data patterns:

#### Legacy APM Account
- **Event Types**: Transaction, TransactionError, TransactionTrace
- **Volume**: High transaction volume (>1000 transactions/hour)
- **Service Identifier**: `appName` field
- **Error Patterns**: Boolean error flags or HTTP status codes

#### Modern OpenTelemetry Account  
- **Event Types**: Span, Metric
- **Volume**: High span volume with dimensional metrics
- **Service Identifier**: `service.name` attribute
- **Metrics**: Duration metrics with service dimensions

#### Mixed Data Patterns Account
- **Event Types**: Transaction, SystemSample, PageView, Log
- **Volume**: Diverse telemetry from multiple sources
- **Patterns**: APM + Infrastructure + Browser + Logs

#### Sparse Data Account (Optional)
- **Event Types**: SyntheticCheck or limited event types
- **Volume**: Low volume, infrequent data
- **Purpose**: Edge case testing for minimal data scenarios

#### EU Region Account (Optional)
- **Region**: Europe (eu01.nr-data.net)
- **Purpose**: Cross-region latency and endpoint testing

## 🚀 Running E2E Tests

### Full Test Suite
```bash
# Run all E2E tests
npm run test:e2e

# Run with verbose output
npm run test:e2e -- --reporter=verbose

# Run specific test category  
npm run test:e2e -- discovery-engine.test.ts
```

### Individual Test Categories

```bash
# Discovery engine tests
npm run test:e2e:discovery

# Enhanced tools tests  
npm run test:e2e:tools

# Dashboard generation tests
npm run test:e2e:dashboards

# Performance tests
npm run test:e2e:performance
```

### Development Testing

```bash
# Quick validation with single account
E2E_ACCOUNT_LEGACY_APM=12345 E2E_API_KEY_LEGACY=key npm run test:e2e:quick

# Debug mode with detailed logging
DEBUG=mcp:* npm run test:e2e
```

## 📊 Test Scenarios

### Discovery Engine Tests
- **Schema Accuracy**: Validates event type discovery matches expected patterns
- **Attribute Profiling**: Tests attribute discovery and type detection
- **Service Identifier Detection**: Verifies automatic service field detection
- **Error Indicator Discovery**: Tests error pattern recognition
- **Metric Discovery**: Validates dimensional metric enumeration
- **Cross-Account Consistency**: Ensures similar account types behave consistently
- **Edge Case Handling**: Tests sparse data and unusual configurations
- **Discovery Caching**: Validates cache efficiency and TTL behavior

### Enhanced Tools Tests
- **NRQL Query Tool**: Tests query validation and schema suggestions
- **Entity Search Tool**: Validates entity discovery and filtering
- **Schema Discovery Tool**: Tests comprehensive schema enumeration
- **Error Handling**: Validates graceful error handling and recovery

### Dashboard Generation Tests
- **Golden Signals**: Tests adaptive golden signals dashboard creation
- **Infrastructure**: Validates infrastructure monitoring dashboards
- **Widget Adaptation**: Tests widget adaptation to discovered schemas
- **Template Compatibility**: Ensures templates work across account types
- **Fallback Mechanisms**: Tests fallback strategies for missing data

### Performance Tests
- **Cache Efficiency**: Measures cache hit ratios and speed improvements
- **Concurrent Discovery**: Tests parallel discovery operations
- **Memory Usage**: Validates memory efficiency and leak detection
- **Latency Benchmarks**: Measures discovery and generation latencies

## 🎯 Success Criteria

### Discovery Accuracy
- **Schema Coverage**: >95% of expected event types discovered
- **Attribute Completeness**: >90% of key attributes profiled
- **Service Identifier Accuracy**: >70% correct detection rate
- **Cross-Account Consistency**: Similar account types show consistent patterns

### Performance Thresholds
- **Discovery Latency**: <5 seconds average per account
- **Cache Efficiency**: >5x speed improvement on cached operations
- **Memory Stability**: <10MB variance in repeated operations
- **Concurrent Operations**: 100% success rate at 5x concurrency

### Tool Enhancement
- **Query Validation**: 100% syntax error detection
- **Schema Suggestions**: >80% accuracy for suggested alternatives
- **Error Recovery**: Graceful handling of all error conditions

### Dashboard Generation
- **Template Success**: >90% successful dashboard generation
- **Widget Adaptation**: >85% widgets successfully adapted
- **Fallback Effectiveness**: >70% fallback success when primary fails

## 🔍 Test Output and Analysis

### Test Results Structure
```typescript
interface E2ETestResults {
  discovery: DiscoveryTestResults;
  tools: EnhancedToolsTestResults;
  dashboards: DashboardTestResults;
  performance: PerformanceTestResults;
  summary: {
    totalAccounts: number;
    successfulTests: number;
    failedTests: number;
    overallSuccessRate: number;
    averageLatency: number;
  };
}
```

### Analysis Reports
- **Discovery Accuracy Report**: Schema coverage and detection rates
- **Performance Benchmark**: Latency, cache efficiency, memory usage
- **Cross-Account Analysis**: Consistency patterns across account types
- **Error Analysis**: Failed tests and recommended improvements

## 🐛 Troubleshooting

### Common Issues

#### Authentication Errors
```bash
Error: Authentication failed for account 12345
Solution: Verify API key has correct permissions (NRQL Query, Entity Search)
```

#### Rate Limiting
```bash
Error: Rate limit exceeded
Solution: Reduce concurrency or add delays between requests
```

#### Sparse Data Accounts
```bash
Warning: No Transaction events found in account
Solution: Expected for sparse accounts - tests will adapt gracefully
```

#### Network Timeouts
```bash
Error: Request timeout after 30000ms
Solution: Increase timeout or check account region configuration
```

### Debug Mode

Enable detailed logging for troubleshooting:

```bash
DEBUG=mcp:discovery,mcp:tools,mcp:dashboards npm run test:e2e
```

## 🎯 Integration with CI/CD

### GitHub Actions Example

```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '20'
      
      - run: npm ci
      - run: npm run build
      
      - name: Run E2E Tests
        env:
          E2E_ACCOUNT_LEGACY_APM: ${{ secrets.E2E_ACCOUNT_LEGACY_APM }}
          E2E_API_KEY_LEGACY: ${{ secrets.E2E_API_KEY_LEGACY }}
          E2E_ACCOUNT_MODERN_OTEL: ${{ secrets.E2E_ACCOUNT_MODERN_OTEL }}
          E2E_API_KEY_OTEL: ${{ secrets.E2E_API_KEY_OTEL }}
        run: npm run test:e2e
```

## 🏆 Validation Results

The E2E test framework provides comprehensive validation that the platform-native MCP server:

- ✅ **Discovers schemas across diverse account configurations**
- ✅ **Adapts tools and dashboards without hardcoded assumptions**
- ✅ **Maintains performance under concurrent operations**
- ✅ **Handles edge cases and error conditions gracefully**
- ✅ **Provides consistent behavior across account types**

This rigorous testing ensures the **Zero Hardcoded Schemas** philosophy is maintained while delivering production-ready performance and reliability.
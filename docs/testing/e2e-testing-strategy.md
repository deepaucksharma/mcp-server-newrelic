# End-to-End Testing Strategy for New Relic MCP Server

## Executive Summary

This document defines a comprehensive end-to-end (E2E) testing strategy that validates the MCP Server against real New Relic data. The strategy emphasizes discovery-first principles, real-world scenarios, and exhaustive validation of all implemented tools.

## Core Testing Principles

1. **Real Data, Real Validation**: All tests run against actual New Relic accounts with real telemetry data
2. **Discovery-First Testing**: Every test workflow starts with discovery, never assumes data structures
3. **Progressive Validation**: Build test complexity incrementally, from basic discovery to complex workflows
4. **Multi-Environment Coverage**: Test across different account types, data volumes, and regions
5. **Failure Mode Testing**: Validate graceful handling of missing data, schema changes, and API limits

## Test Environment Setup

### Required New Relic Resources

```yaml
test_accounts:
  primary:
    purpose: "Main testing account with diverse data"
    requirements:
      - Active APM data from multiple services
      - Infrastructure monitoring enabled
      - Browser monitoring active
      - Custom events and metrics
      - Historical data (30+ days)
      
  secondary:
    purpose: "Cross-account testing"
    requirements:
      - Different data schema patterns
      - Limited permissions for RBAC testing
      
  empty:
    purpose: "Zero-data scenarios"
    requirements:
      - No active data ingestion
      - Tests discovery behavior with no data
      
  high_cardinality:
    purpose: "Performance and scale testing"
    requirements:
      - 1000+ unique services
      - High event volume (1M+ events/hour)
      - Complex attribute patterns
```

### Environment Configuration

```bash
# .env.test configuration
NEW_RELIC_API_KEY_PRIMARY="<primary-account-key>"
NEW_RELIC_ACCOUNT_ID_PRIMARY="<primary-account-id>"

NEW_RELIC_API_KEY_SECONDARY="<secondary-account-key>"
NEW_RELIC_ACCOUNT_ID_SECONDARY="<secondary-account-id>"

NEW_RELIC_API_KEY_EMPTY="<empty-account-key>"
NEW_RELIC_ACCOUNT_ID_EMPTY="<empty-account-id>"

NEW_RELIC_API_KEY_HIGH_CARD="<high-card-account-key>"
NEW_RELIC_ACCOUNT_ID_HIGH_CARD="<high-card-account-id>"

# Test execution settings
E2E_TEST_TIMEOUT="300s"
E2E_PARALLEL_TESTS="4"
E2E_RETRY_ATTEMPTS="3"
E2E_CACHE_DISABLED="true"
```

## Test Categories

### 1. Discovery Foundation Tests

Validate core discovery capabilities without assumptions:

```yaml
discovery_foundation:
  test_event_type_discovery:
    description: "Discover all event types in account"
    validations:
      - Event types are returned
      - Each type has sample count
      - Metadata includes freshness
      - No hardcoded assumptions
      
  test_attribute_discovery:
    description: "Discover attributes for each event type"
    validations:
      - Attributes discovered dynamically
      - Data types identified correctly
      - Cardinality assessed
      - Coverage percentages accurate
      
  test_empty_account_discovery:
    description: "Handle accounts with no data gracefully"
    validations:
      - Returns empty results, not errors
      - Provides helpful guidance
      - Suggests next steps
```

### 2. Query Adaptation Tests

Test query building based on discovered schemas:

```yaml
query_adaptation:
  test_service_query_adaptation:
    description: "Build service queries without assuming 'appName'"
    steps:
      1. Discover event types containing service data
      2. Identify service identifier attributes
      3. Build queries using discovered attributes
      4. Validate results match expectations
      
  test_error_detection_adaptation:
    description: "Find errors without assuming error structure"
    steps:
      1. Discover how errors are represented
      2. Identify error indicators (boolean, string, code)
      3. Build error queries adaptively
      4. Validate error counts and patterns
      
  test_missing_attribute_handling:
    description: "Gracefully handle missing expected attributes"
    validations:
      - Queries adapt to available attributes
      - Fallback strategies work correctly
      - User receives clear explanations
```

### 3. Workflow Integration Tests

Test complete discovery-first workflows:

```yaml
workflow_tests:
  test_performance_investigation_workflow:
    description: "Complete performance investigation from scratch"
    phases:
      discovery:
        - Discover performance-related event types
        - Profile duration/latency attributes
        - Identify service/transaction hierarchies
        
      analysis:
        - Calculate baselines from discovered metrics
        - Detect anomalies in current data
        - Correlate with discovered dimensions
        
      action:
        - Generate alerts based on findings
        - Create dashboard from discoveries
        - Document investigation results
        
  test_incident_response_workflow:
    description: "Respond to incident without assumptions"
    phases:
      understand:
        - Discover what triggered the alert
        - Find related data sources
        - Identify impact scope
        
      investigate:
        - Correlate across discovered dimensions
        - Find changes from baseline
        - Identify contributing factors
        
      respond:
        - Generate targeted queries
        - Create visibility dashboards
        - Update alerts based on findings
```

### 4. Cross-Account Tests

Validate multi-account scenarios:

```yaml
cross_account:
  test_account_switching:
    description: "Query different accounts without reconfiguration"
    validations:
      - Tools accept account_id parameter
      - Results reflect correct account
      - No data leakage between accounts
      
  test_cross_account_correlation:
    description: "Correlate data across accounts"
    steps:
      1. Discover schemas in each account
      2. Find common attributes
      3. Build cross-account queries
      4. Validate correlation results
      
  test_permission_boundaries:
    description: "Respect account permissions"
    validations:
      - Limited accounts show restricted data
      - Error messages don't leak information
      - Graceful degradation of features
```

### 5. Data Quality Tests

Validate handling of real-world data issues:

```yaml
data_quality:
  test_incomplete_data:
    description: "Handle partial data gracefully"
    scenarios:
      - Missing required attributes
      - Null values in key fields
      - Sparse time series data
      - Incomplete transactions
      
  test_schema_evolution:
    description: "Adapt to changing schemas"
    scenarios:
      - New attributes appear
      - Attributes disappear
      - Data types change
      - Cardinality explosion
      
  test_high_cardinality:
    description: "Handle high cardinality data"
    validations:
      - Queries remain performant
      - Appropriate LIMIT usage
      - Memory usage bounded
      - Timeout handling works
```

### 6. Performance Tests

Validate performance with real data volumes:

```yaml
performance:
  test_discovery_performance:
    description: "Discovery completes in reasonable time"
    thresholds:
      - Event type discovery: < 5s
      - Attribute profiling: < 10s per type
      - Relationship detection: < 30s
      
  test_query_performance:
    description: "Queries perform acceptably"
    thresholds:
      - Simple queries: < 2s
      - Complex aggregations: < 10s
      - Large result sets: < 30s
      
  test_concurrent_operations:
    description: "Handle concurrent requests"
    validations:
      - 10 concurrent discoveries
      - No resource exhaustion
      - Fair scheduling
      - Graceful degradation
```

### 7. Error Handling Tests

Test failure modes with real scenarios:

```yaml
error_handling:
  test_api_errors:
    description: "Handle New Relic API errors gracefully"
    scenarios:
      - Rate limiting (429)
      - Authentication failures (401)
      - Permission denied (403)
      - Service errors (500)
      - Timeout errors
      
  test_data_errors:
    description: "Handle problematic data"
    scenarios:
      - Malformed event data
      - Circular relationships
      - Infinite cardinality
      - Query complexity limits
      
  test_recovery_mechanisms:
    description: "Validate error recovery"
    validations:
      - Automatic retries work
      - Circuit breakers activate
      - Fallback strategies engage
      - Clear error communication
```

### 8. Regional Tests

Test region-specific behaviors:

```yaml
regional:
  test_us_region:
    description: "Validate US region functionality"
    validations:
      - Correct endpoints used
      - Data residency respected
      - Performance acceptable
      
  test_eu_region:
    description: "Validate EU region functionality"
    validations:
      - EU endpoints used
      - GDPR compliance
      - Cross-region isolation
      
  test_region_auto_detection:
    description: "Auto-detect region from API key"
    validations:
      - Region correctly identified
      - Appropriate endpoints selected
      - No manual configuration needed
```

## Test Implementation Framework

### Test Structure

```go
// tests/e2e/discovery_test.go
package e2e

import (
    "context"
    "testing"
    "github.com/stretchr/testify/suite"
)

type DiscoveryE2ESuite struct {
    suite.Suite
    primaryAccount   *TestAccount
    secondaryAccount *TestAccount
    emptyAccount     *TestAccount
    mcp             *MCPTestClient
}

func (s *DiscoveryE2ESuite) SetupSuite() {
    // Initialize test accounts from environment
    s.primaryAccount = NewTestAccount("PRIMARY")
    s.secondaryAccount = NewTestAccount("SECONDARY")
    s.emptyAccount = NewTestAccount("EMPTY")
    
    // Create MCP test client
    s.mcp = NewMCPTestClient(s.primaryAccount)
}

func (s *DiscoveryE2ESuite) TestEventTypeDiscovery() {
    // Never assume what event types exist
    ctx := context.Background()
    
    // Execute discovery
    result, err := s.mcp.Execute(ctx, "discovery.explore_event_types", map[string]interface{}{
        "time_range": "24 hours",
    })
    
    s.NoError(err, "Discovery should not error")
    s.NotNil(result, "Should return results")
    
    // Validate discovered event types
    eventTypes := result["event_types"].([]interface{})
    s.NotEmpty(eventTypes, "Should discover event types")
    
    // Verify each event type has required metadata
    for _, et := range eventTypes {
        eventType := et.(map[string]interface{})
        s.Contains(eventType, "name")
        s.Contains(eventType, "count")
        s.Contains(eventType, "attributes")
        s.Contains(eventType, "sample")
    }
    
    // Store discovered types for subsequent tests
    s.mcp.StoreDiscovery("event_types", eventTypes)
}

func (s *DiscoveryE2ESuite) TestAdaptiveServiceQuery() {
    ctx := context.Background()
    
    // First, discover how services are identified
    serviceAttr := s.discoverServiceAttribute(ctx)
    s.NotEmpty(serviceAttr, "Should discover service attribute")
    
    // Build query using discovered attribute
    query := fmt.Sprintf(
        "SELECT count(*) FROM Transaction WHERE %s IS NOT NULL FACET %s",
        serviceAttr, serviceAttr,
    )
    
    result, err := s.mcp.Execute(ctx, "nrql.execute", map[string]interface{}{
        "query": query,
    })
    
    s.NoError(err, "Adaptive query should succeed")
    s.NotEmpty(result["results"], "Should return service data")
}

func (s *DiscoveryE2ESuite) discoverServiceAttribute(ctx context.Context) string {
    // Discover which attribute identifies services
    candidates := []string{"appName", "service.name", "applicationName", "app"}
    
    for _, candidate := range candidates {
        result, _ := s.mcp.Execute(ctx, "discovery.profile_attribute", map[string]interface{}{
            "event_type": "Transaction",
            "attribute":  candidate,
        })
        
        if profile, ok := result["profile"].(map[string]interface{}); ok {
            if coverage, ok := profile["coverage"].(float64); ok && coverage > 50 {
                return candidate
            }
        }
    }
    
    // If no standard attribute, discover from data
    result, _ := s.mcp.Execute(ctx, "nrql.execute", map[string]interface{}{
        "query": "SELECT keyset() FROM Transaction LIMIT 1",
    })
    
    // Analyze attributes to find service identifier
    // ... (additional discovery logic)
    
    return ""
}
```

### Test Execution Pipeline

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests

on:
  schedule:
    - cron: '0 */4 * * *'  # Every 4 hours
  workflow_dispatch:
    inputs:
      test_filter:
        description: 'Test filter pattern'
        required: false
        default: ''

jobs:
  e2e-discovery:
    runs-on: ubuntu-latest
    timeout-minutes: 30
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Configure test environment
        run: |
          echo "NEW_RELIC_API_KEY_PRIMARY=${{ secrets.NR_API_KEY_PRIMARY }}" >> .env.test
          echo "NEW_RELIC_ACCOUNT_ID_PRIMARY=${{ secrets.NR_ACCOUNT_PRIMARY }}" >> .env.test
          # ... other accounts
          
      - name: Run discovery tests
        run: |
          go test -v -timeout 30m \
            -run "TestDiscovery" \
            ./tests/e2e/... \
            -parallel 4
            
      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: e2e-discovery-results
          path: tests/results/
          
  e2e-workflows:
    runs-on: ubuntu-latest
    timeout-minutes: 45
    needs: e2e-discovery
    steps:
      # ... similar setup ...
      
      - name: Run workflow tests
        run: |
          go test -v -timeout 45m \
            -run "TestWorkflow" \
            ./tests/e2e/... \
            -parallel 2
            
  e2e-performance:
    runs-on: ubuntu-latest
    timeout-minutes: 60
    steps:
      # ... setup ...
      
      - name: Run performance tests
        run: |
          go test -v -timeout 60m \
            -run "TestPerformance" \
            ./tests/e2e/... \
            -bench=.
```

### Test Data Management

```go
// tests/e2e/testdata/generator.go
package testdata

// DataGenerator creates predictable test data in New Relic
type DataGenerator struct {
    client *newrelic.Client
    config GeneratorConfig
}

type GeneratorConfig struct {
    EventTypes []EventTypeConfig
    TimeRange  time.Duration
    EventRate  int // events per minute
}

func (g *DataGenerator) Generate(ctx context.Context) error {
    // Generate test data with known patterns
    for _, eventType := range g.config.EventTypes {
        if err := g.generateEvents(ctx, eventType); err != nil {
            return fmt.Errorf("generate %s: %w", eventType.Name, err)
        }
    }
    return nil
}

func (g *DataGenerator) Cleanup(ctx context.Context) error {
    // Remove test data using identifiable markers
    query := `DELETE FROM Transaction, CustomEvent 
              WHERE testMarker = 'e2e-test' 
              SINCE 1 day ago`
    return g.client.ExecuteNRQL(ctx, query)
}
```

### Test Result Analysis

```go
// tests/e2e/analyzer/report.go
package analyzer

type E2ETestReport struct {
    StartTime    time.Time
    EndTime      time.Time
    Environment  string
    Results      []TestResult
    Coverage     CoverageReport
    Performance  PerformanceReport
}

func (r *E2ETestReport) GenerateHTML() string {
    // Generate comprehensive test report
    // Include discovered schemas, adaption strategies, performance metrics
}

func (r *E2ETestReport) AnalyzeFailures() []FailurePattern {
    // Identify patterns in failures
    // Suggest root causes
    // Recommend fixes
}
```

## Test Execution Strategy

### Daily Execution

1. **Morning Run (6 AM)**
   - Full discovery test suite
   - Basic workflow validation
   - Cross-account verification

2. **Afternoon Run (2 PM)**
   - Performance tests
   - High-cardinality scenarios
   - Regional validation

3. **Evening Run (10 PM)**
   - Error handling scenarios
   - Chaos engineering tests
   - Recovery validation

### Weekly Deep Tests

1. **Monday**: Exhaustive discovery validation
2. **Wednesday**: Complex workflow scenarios
3. **Friday**: Performance and scale testing

### Release Testing

Before each release:
1. Full E2E suite execution
2. Regression validation
3. New feature verification
4. Performance benchmarking
5. Multi-region validation

## Success Metrics

### Coverage Metrics
- 100% of implemented tools tested with real data
- 100% of documented workflows validated
- 95%+ of error scenarios covered
- All regions and account types tested

### Quality Metrics
- Zero false positives in discovery
- <2% test flakiness rate
- All tests complete within SLA
- Clear failure diagnostics

### Performance Metrics
- Discovery operations < 5s
- Query operations < 10s
- Workflow completion < 5 minutes
- Memory usage < 500MB

## Maintenance Strategy

### Test Data Freshness
- Automated data generation daily
- Realistic patterns maintained
- Historical data preserved
- Anomaly injection for testing

### Test Evolution
- New tests for each feature
- Regression tests for bugs
- Performance baselines updated
- Documentation synchronized

### Monitoring
- Test execution dashboards
- Failure analysis reports
- Performance trend tracking
- Coverage reporting

## Conclusion

This E2E testing strategy ensures the MCP Server truly embodies discovery-first principles while working reliably with real New Relic data. By testing against actual accounts with diverse data patterns, we validate that the system adapts to any environment without making assumptions.

The exhaustive nature of these tests, combined with real data validation, provides confidence that the MCP Server will work correctly for any New Relic customer, regardless of their specific data structures or naming conventions.

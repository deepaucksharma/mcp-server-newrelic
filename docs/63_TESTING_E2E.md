# End-to-End Testing Framework

This document provides a comprehensive guide to the E2E testing framework for the New Relic MCP Server, including YAML-based scenario definitions, test execution strategies, and real-world validation approaches.

## Overview

The E2E testing framework validates the MCP Server against real New Relic environments using YAML-based test scenarios. It emphasizes discovery-first principles, real API testing, and complex workflow validation.

## Framework Architecture

```
tests/e2e/
├── harness/              # Core test execution engine
│   ├── runner.go         # Orchestrates scenario execution
│   ├── scenario.go       # YAML scenario parser with DSL
│   ├── executor.go       # Executes workflow steps
│   ├── assertions.go     # Assertion evaluation engine
│   └── report.go         # Multi-format reporting
├── scenarios/            # YAML test definitions
│   ├── disc-*.yaml       # Discovery scenarios
│   ├── inc-*.yaml        # Incident response scenarios
│   ├── perf-*.yaml       # Performance scenarios
│   ├── gov-*.yaml        # Governance scenarios
│   └── chaos-*.yaml      # Chaos engineering scenarios
├── scripts/              # Test data generation
│   ├── seed-*.py         # Python data seeders
│   └── validate-*.sh     # Setup validation
├── framework/            # Test client utilities
│   ├── client.go         # MCP client wrapper
│   └── mcp_client.go     # Protocol implementation
└── workflows/            # Complex test workflows
    └── *_test.go         # Go-based test scenarios
```

## YAML Scenario Structure

### Basic Structure

```yaml
# Unique identifier following naming convention
id: DISC-MISS-001

# Human-readable title
title: Handle missing attributes in discovered schema

# Tags for test organization and filtering
tags:
  - discovery
  - resilience
  - critical

# Environment configuration
environment:
  account_type: single-account  # or multi-account, cross-region
  data_source_mix: "apm,infra"  # Required data types
  variables:
    service_name: "test-service"
    time_range: "${fn:now()-1h}"

# Optional setup phase
setup:
  seed_data_script: scripts/seed-missing-attributes.py
  environment:
    MISSING_FIELDS: "customAttribute1,customAttribute2"
  wait: 5s

# Main test workflow
workflow:
  # Sequential steps by default
  - tool: discovery.explore_event_types
    params:
      limit: 100
    store_as: event_types
    
  # Conditional execution
  - tool: discovery.explore_attributes
    params:
      event_type: "Transaction"
    store_as: attributes
    condition: "${event_types.count} > 0"
    
  # Parallel execution block
  - parallel:
    - tool: query_nrdb
      params:
        query: "SELECT count(*) FROM Transaction"
      store_as: tx_count
      
    - tool: query_nrdb
      params:
        query: "SELECT count(*) FROM SystemSample"
      store_as: sys_count

# Assertions to validate results
assert:
  # JSONPath assertions
  - jsonpath: "$.event_types.event_types[?(@.name=='Transaction')].count"
    operator: ">"
    value: 0
    message: "Transaction events should exist"
    
  # NRQL assertions
  - type: nrql
    query: "SELECT count(*) FROM Transaction WHERE e2e_test = true"
    operator: ">="
    value: 100
    message: "Test data should be present"
    
  # Custom assertions
  - type: trace
    operator: shows_discovery_adaptation
    value: true
    message: "Should adapt to missing attributes"

# Cleanup phase
cleanup:
  delete_test_data: true
  custom_commands:
    - "DELETE FROM Transaction WHERE e2e_test = true"
```

### DSL Features

#### 1. Variable Substitution

```yaml
# Environment variables
account_id: "${E2E_PRIMARY_ACCOUNT_ID}"

# Workflow variables
query: "SELECT * FROM ${event_type}"

# Stored results
filter: "appName = '${app_details.name}'"

# Nested access
value: "${results.data[0].metrics.average}"
```

#### 2. Built-in Functions

```yaml
# Time functions
start_time: "${fn:now()-1h}"
end_time: "${fn:now()}"
yesterday: "${fn:now()-24h}"

# Data functions
count: "${fn:count(${results.data})}"
unique: "${fn:unique(${results.apps})}"

# String functions
upper: "${fn:upper(${service_name})}"
contains: "${fn:contains(${message}, 'error')}"
```

#### 3. Adaptive Query Builder

```yaml
# Use discovered schema to build queries
- tool: query.build_adaptive
  params:
    template: performance_analysis
    discovered_schema: "${attributes}"
    fallback_fields:
      duration: ["responseTime", "elapsed"]
  store_as: adaptive_query
```

#### 4. Conditional Logic

```yaml
# Skip if condition not met
condition: "${previous_step.success} == true"

# Complex conditions
condition: "${tx_count.value} > 1000 AND ${error_rate} < 0.05"

# Conditional blocks
- if: "${environment.type} == 'production'"
  then:
    - tool: alerts.create
      params: {...}
  else:
    - tool: query_nrdb
      params: {...}
```

## Scenario Categories

### Discovery Scenarios (disc-*)

Test discovery-first principles and adaptation:

```yaml
# DISC-MISS-001: Missing Attributes
# Tests adaptation when expected fields don't exist
id: DISC-MISS-001
workflow:
  - tool: discovery.explore_attributes
    params:
      event_type: "Transaction"
    store_as: schema
    
  - tool: query.build_adaptive
    params:
      required_fields: ["duration", "customField"]
      discovered_schema: "${schema}"
      adapt_strategy: "fallback"
```

### Incident Response (inc-*)

Complex troubleshooting workflows:

```yaml
# INC-SQL-404: Database Error Investigation
# Multi-step root cause analysis
id: INC-SQL-404
workflow:
  # Alert detection
  - tool: alerts.get_violation
    params:
      alert_name: "High Error Rate"
      
  # Parallel data gathering
  - parallel:
    - tool: analysis.detect_anomalies
      params:
        metric: "error_rate"
    - tool: analysis.correlate_metrics
      params:
        primary: "errors"
        secondary: ["database", "cpu", "memory"]
        
  # Root cause analysis
  - tool: analysis.summarize_incident
    params:
      data_sources: ["${anomalies}", "${correlations}"]
```

### Performance Analysis (perf-*)

Performance optimization scenarios:

```yaml
# PERF-CMP-001: Cross-Region Comparison
id: PERF-CMP-001
workflow:
  - tool: query_nrdb
    params:
      query: |
        SELECT average(duration) as avg_duration
        FROM Transaction
        FACET aws.region
      accounts: ["${us_account}", "${eu_account}"]
    store_as: regional_performance
```

### Governance Scenarios (gov-*)

Cost and compliance testing:

```yaml
# GOV-COST-001: Cost Optimization
id: GOV-COST-001
workflow:
  - tool: governance.analyze_usage
    params:
      time_range: "30 days"
      
  - tool: governance.optimize_costs
    params:
      target_reduction: 20
      preserve: ["critical_services"]
```

### Chaos Engineering (chaos-*)

Resilience testing with failure injection:

```yaml
# CHAOS-NET-001: Network Resilience
id: CHAOS-NET-001
setup:
  toxiproxy:
    proxies:
      - name: newrelic_api
        upstream: api.newrelic.com:443
        toxics:
          - type: latency
            attributes:
              latency: 5000
              
workflow:
  - tool: query_nrdb
    params:
      query: "SELECT count(*) FROM Transaction"
    timeout: 10s
    retry:
      max_attempts: 3
      backoff: exponential
```

## Assertion Types

### JSONPath Assertions

```yaml
# Basic value check
- jsonpath: "$.data.results[0].count"
  operator: ">"
  value: 1000
  
# Array contains
- jsonpath: "$.event_types[?(@.name=='Transaction')]"
  operator: "exists"
  
# Complex path
- jsonpath: "$.data.facets[?(@.name=='error')].results[0].percentage"
  operator: "<"
  value: 5.0
```

### NRQL Assertions

```yaml
# Direct query assertion
- type: nrql
  query: "SELECT uniqueCount(host) FROM SystemSample"
  operator: ">="
  value: 10
  message: "Should have at least 10 hosts"
  
# Time-based assertion
- type: nrql
  query: |
    SELECT count(*) 
    FROM Transaction 
    WHERE timestamp > ${test_start_time}
  operator: ">"
  value: 0
```

### Trace Assertions

```yaml
# Validate discovery behavior
- type: trace
  operator: shows_discovery_first
  value: true
  
# Check adaptation occurred
- type: trace
  operator: adapted_query_count
  value: 3
  
# Validate tool composition
- type: trace
  operator: tool_sequence
  value: ["discovery.explore_event_types", "query.build_adaptive", "query_nrdb"]
```

### Statistical Assertions

```yaml
# Approximate equality (within 10%)
- jsonpath: "$.metrics.average"
  operator: "approx"
  value: 100
  tolerance: 0.1
  
# Standard deviation check
- jsonpath: "$.statistics.std_dev"
  operator: "<"
  value: 50
  
# Percentile validation
- jsonpath: "$.percentiles.p95"
  operator: "between"
  value: [100, 200]
```

## Test Data Management

### Seed Scripts

Python scripts for generating test data:

```python
# scripts/seed-performance-data.py
import os
import time
import random
from newrelic_telemetry_sdk import MetricClient, GaugeMetric

def seed_data():
    # Tag all test data for cleanup
    base_tags = {'e2e_test': True, 'test_id': os.environ['TEST_ID']}
    
    # Generate realistic patterns
    for i in range(100):
        metrics = [
            GaugeMetric(
                name='custom.response_time',
                value=random.gauss(100, 20),
                tags={**base_tags, 'endpoint': '/api/users'}
            ),
            GaugeMetric(
                name='custom.error_count',
                value=random.randint(0, 5),
                tags={**base_tags, 'endpoint': '/api/users'}
            )
        ]
        
        # Add anomaly if requested
        if i > 50 and os.environ.get('INJECT_ANOMALY'):
            metrics[0].value *= 5  # Spike response time
            
        client.send_batch(metrics)
        time.sleep(0.1)
```

### Data Cleanup

Automatic cleanup strategies:

```yaml
cleanup:
  # Standard cleanup
  delete_test_data: true
  
  # Custom NRQL cleanup
  custom_commands:
    - "DELETE FROM Transaction WHERE e2e_test = true AND test_id = '${test_id}'"
    - "DELETE FROM Metric WHERE e2e_test = true AND timestamp < ${fn:now()}"
    
  # Dashboard cleanup
  drop_dashboards_with_tag: "e2e-test"
  
  # Alert cleanup
  delete_alerts_with_prefix: "E2E-TEST-"
```

## Execution Modes

### Local Development

```bash
# Run single scenario
go test -v ./tests/e2e -run TestScenario -scenario disc-miss-001

# Run with debug output
E2E_DEBUG=true go test -v ./tests/e2e

# Run specific category
go test -v ./tests/e2e -run TestScenario -tags discovery
```

### CI/CD Pipeline

```bash
# GitHub Actions
- name: E2E Tests
  run: |
    make e2e-test-ci
  env:
    E2E_PRIMARY_ACCOUNT_ID: ${{ secrets.NR_ACCOUNT }}
    E2E_PRIMARY_API_KEY: ${{ secrets.NR_API_KEY }}
    
# Generate reports
- name: Test Report
  if: always()
  run: |
    make e2e-report
    
- uses: actions/upload-artifact@v3
  with:
    name: e2e-test-report
    path: tests/e2e/reports/
```

### Parallel Execution

```bash
# Run scenarios in parallel
E2E_PARALLEL=10 make e2e-test

# Category-based parallel execution
make e2e-test-parallel CATEGORIES="discovery,incident"
```

## Performance Optimization

### Scenario Design

1. **Minimize Setup**: Share test data across related scenarios
2. **Parallel Steps**: Use parallel blocks for independent operations
3. **Smart Assertions**: Check critical paths first
4. **Efficient Queries**: Use LIMIT and time bounds

### Caching Strategy

```yaml
# Cache discovery results
- tool: discovery.explore_event_types
  params:
    cache_key: "event_types_${account_id}"
    cache_ttl: 300
  store_as: cached_events
```

### Resource Management

```yaml
environment:
  limits:
    max_query_duration: 30s
    max_memory: 1GB
    max_parallel_steps: 5
```

## Debugging Failed Tests

### Debug Mode

```bash
# Enable verbose logging
E2E_DEBUG=true go test -v ./tests/e2e -run TestScenario/disc-miss-001

# Save all API calls
E2E_SAVE_REQUESTS=true go test -v ./tests/e2e

# Interactive mode (pause on failure)
E2E_INTERACTIVE=true go test -v ./tests/e2e
```

### Trace Analysis

```yaml
# Enable detailed tracing
workflow:
  - tool: discovery.explore_event_types
    params:
      limit: 10
    trace:
      level: detailed
      save_to: "./traces/${test_id}.json"
```

### Common Issues

1. **Timing Issues**
   ```yaml
   setup:
     wait: 10s  # Wait for data propagation
   ```

2. **Account Permissions**
   ```yaml
   # Check permissions before test
   - tool: account.verify_permissions
     params:
       required: ["nrql", "alerts.read"]
   ```

3. **Data Retention**
   ```yaml
   # Adjust time ranges for retention
   time_range: "${fn:now()-1h}"  # Instead of -7d
   ```

## Best Practices

### 1. Test Design

- **Atomic Tests**: Each scenario should test one workflow
- **Clear Naming**: Follow naming conventions (DISC-, INC-, etc.)
- **Comprehensive Assertions**: Test both happy path and edge cases
- **Cleanup Always**: Ensure test data is removed

### 2. Data Management

```yaml
# Always tag test data
setup:
  seed_data_script: scripts/seed-data.py
  environment:
    TEST_ID: "${test_id}"
    E2E_TEST: "true"
```

### 3. Error Handling

```yaml
# Graceful degradation
- tool: query_nrdb
  params:
    query: "SELECT * FROM CustomEvent"
  on_error: continue
  store_error_as: custom_event_error
  
# Retry logic
- tool: alerts.create
  retry:
    max_attempts: 3
    backoff: exponential
    on_errors: ["timeout", "rate_limit"]
```

### 4. Documentation

```yaml
# Document complex scenarios
id: INC-CASCADE-001
title: Cascading failure investigation across services
description: |
  This scenario tests the MCP server's ability to:
  1. Detect cascading failures across multiple services
  2. Correlate errors with infrastructure metrics
  3. Generate actionable remediation steps
  
  Prerequisites:
  - Multi-service APM data
  - Infrastructure agent on all hosts
  - Distributed tracing enabled
```

## Integration with Development

### Pre-commit Testing

```bash
# .git/hooks/pre-commit
#!/bin/bash
make e2e-test-minimal
```

### Pull Request Testing

```yaml
# .github/workflows/pr.yml
on: [pull_request]
jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        scenario_group: [discovery, incident, performance]
    steps:
      - uses: actions/checkout@v3
      - name: Run E2E Tests
        run: |
          make e2e-test-group GROUP=${{ matrix.scenario_group }}
```

## Extending the Framework

### Adding New Assertions

```go
// harness/assertions.go
func (e *AssertionEngine) RegisterCustomAssertion(name string, fn AssertionFunc) {
    e.customAssertions[name] = fn
}

// Usage in YAML
assert:
  - type: custom
    assertion: validate_sli_compliance
    params:
      sli_target: 0.999
      actual: "${results.availability}"
```

### Custom Tool Integration

```go
// Add tool result validation
type ToolValidator interface {
    ValidateResult(result interface{}) error
}

// Register validator
harness.RegisterValidator("discovery.explore_event_types", &DiscoveryValidator{})
```

## Reporting and Metrics

### Test Reports

```bash
# Generate HTML report
make e2e-report-html

# JUnit XML for CI
make e2e-report-junit

# Custom New Relic dashboard
make e2e-report-nr
```

### Metrics Collection

```yaml
# Track test execution metrics
metrics:
  enabled: true
  backend: newrelic
  custom_attributes:
    test_category: "${category}"
    environment: "${environment.type}"
```

## Summary

The E2E testing framework provides:

1. **YAML-based scenario definition** for maintainable tests
2. **Discovery-first validation** ensuring no assumptions
3. **Real API testing** against actual New Relic accounts
4. **Complex workflow support** with parallel execution
5. **Comprehensive assertions** including NRQL and traces
6. **Chaos engineering** capabilities for resilience testing
7. **CI/CD integration** with multiple report formats
8. **Extensible architecture** for custom validations

The framework ensures the MCP Server works correctly in real-world scenarios while maintaining the core principle of making zero assumptions about data structures.
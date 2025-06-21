# End-to-End Testing Framework

This directory contains the comprehensive E2E testing framework for the MCP Server New Relic project. The framework validates real-world scenarios using actual New Relic APIs and data.

## Architecture Overview

```
tests/e2e/
├── harness/              # Core test execution framework
│   ├── runner.go         # Orchestrates scenario execution
│   ├── parser.go         # YAML scenario parser with DSL
│   ├── executor.go       # Executes workflow steps
│   ├── assertions.go     # Assertion evaluation engine
│   ├── report.go         # Multi-format reporting
│   └── trace.go          # Trace collection and analysis
├── scenarios/            # YAML test scenarios
│   ├── disc-*.yaml       # Discovery scenarios
│   ├── inc-*.yaml        # Incident response scenarios
│   ├── perf-*.yaml       # Performance scenarios
│   ├── gov-*.yaml        # Governance scenarios
│   └── chaos-*.yaml      # Chaos engineering scenarios
├── scripts/              # Data seeding and utilities
│   └── seed-*.py         # Python scripts for test data
├── discovery/            # Discovery-focused tests
├── workflows/            # Complex workflow tests
└── framework/            # Test framework utilities
```

## Key Features

### 1. Discovery-First Testing
- Never assumes data structures exist
- Validates adaptive behavior with missing fields
- Tests schema evolution and drift

### 2. Real API Testing
- Uses actual New Relic accounts via .env configuration
- No mocks - validates against production APIs
- Multi-account and cross-region support

### 3. Chaos Engineering
- Toxiproxy integration for network faults
- Progressive failure injection
- Resilience validation

### 4. Complex Workflows
- Multi-tool orchestration
- Parallel execution support
- Conditional steps and retries

### 5. Comprehensive Assertions
- JSONPath expressions
- NRQL query assertions
- Trace analysis
- Statistical comparisons

## Quick Start

### 1. Configure Test Accounts

```bash
cp .env.test.example .env.test
# Edit .env.test with your New Relic credentials
```

Required environment variables:
```bash
E2E_PRIMARY_ACCOUNT_ID=your-primary-account
E2E_PRIMARY_API_KEY=your-primary-key
E2E_SECONDARY_ACCOUNT_ID=your-secondary-account
E2E_SECONDARY_API_KEY=your-secondary-key
```

### 2. Run Tests

```bash
# Run all E2E tests
make e2e-test

# Run specific scenario
make e2e-test-scenario SCENARIO=disc-miss-001

# Run by tag
make e2e-test-tag TAG=critical

# Run with chaos testing
make e2e-test-chaos
```

### 3. View Results

```bash
# Generate HTML report
make e2e-report

# View in browser
open tests/e2e/reports/index.html
```

## Writing Scenarios

### Scenario Structure

```yaml
id: DISC-MISS-001
title: Descriptive title
tags: [discovery, critical]

environment:
  account_type: single-account
  variables:
    key: value

setup:
  seed_data_script: scripts/seed-data.py
  toxiproxy:
    proxies: [...]

workflow:
  - tool: discovery.explore_event_types
    params:
      limit: 100
    store_as: events
    
  - tool: nrql.execute
    params:
      query: "${aqb:build.latency}"
    store_as: results
    condition: "${events.count} > 0"

assert:
  - jsonpath: "$.results.error"
    operator: "=="
    value: null
    message: "Query should succeed"

cleanup:
  delete_test_data: true
```

### DSL Features

1. **Variable Substitution**: `${variable_name}`
2. **Adaptive Query Builder**: `${aqb:template.name}`
3. **Functions**: `${fn:now()}`, `${fn:now()-15m}`
4. **JSONPath**: `${events.results[0].name}`
5. **Conditions**: Skip steps based on previous results

## Scenario Categories

### Discovery (DISC-*)
Tests that validate the discovery-first philosophy:
- `DISC-MISS-001`: Missing attributes handling
- `DISC-DRIFT-001`: Schema evolution
- `DISC-MULTI-001`: Multi-account consolidation

### Incident Response (INC-*)
Complex incident investigation workflows:
- `INC-SQL-404`: Database connectivity issues
- `INC-SPIKE-001`: Traffic spike analysis
- `INC-CASCADE-001`: Cascading failures

### Performance (PERF-*)
Performance analysis and optimization:
- `PERF-CMP-001`: Cross-region comparison
- `PERF-TREND-001`: Trend analysis
- `PERF-OPT-001`: Query optimization

### Governance (GOV-*)
Cost and compliance scenarios:
- `GOV-COST-001`: Cost optimization
- `GOV-COMPL-001`: Compliance audit
- `GOV-USAGE-001`: Resource governance

### Chaos (CHAOS-*)
Network and failure testing:
- `CHAOS-NET-001`: Network resilience
- `CHAOS-RATE-001`: Rate limiting
- `CHAOS-REGION-001`: Region failover

## Assertion Types

### JSONPath Assertions
```yaml
- jsonpath: "$.data.results[0].value"
  operator: ">"
  value: 100
```

### NRQL Assertions
```yaml
- type: nrql
  query: "SELECT count(*) FROM Transaction"
  operator: ">="
  value: 1000
```

### Trace Assertions
```yaml
- type: trace
  operator: trace_shows_discovery
  value: true
```

### Operators
- `==`, `!=`: Equality
- `>`, `<`, `>=`, `<=`: Comparison
- `contains`, `not_contains`: String/array contains
- `matches`: Regex matching
- `approx`: Approximate equality (±10%)

## Seed Data Scripts

Located in `scripts/`, these Python scripts create test data:

```python
# Example: seed-missing-attributes.py
event = {
    'eventType': 'Transaction',
    'appName': 'test-app',
    'duration': 0.123,
    'e2e_test': True  # Flag for cleanup
}
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Run E2E Tests
  env:
    E2E_PRIMARY_ACCOUNT_ID: ${{ secrets.NR_ACCOUNT_ID }}
    E2E_PRIMARY_API_KEY: ${{ secrets.NR_API_KEY }}
  run: make e2e-test-ci
```

### JUnit Reports
Tests generate JUnit XML for CI integration:
```bash
tests/e2e/reports/junit.xml
```

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify API keys in `.env.test`
   - Check account permissions

2. **Timeouts**
   - Increase timeout in scenario
   - Check network connectivity
   - Verify chaos proxy settings

3. **Missing Data**
   - Run seed scripts manually
   - Check data retention settings
   - Verify time ranges

### Debug Mode

```bash
# Enable debug logging
export E2E_DEBUG=true
make e2e-test

# Trace MCP calls
export MCP_DEBUG=true
```

## Best Practices

1. **Always Clean Up**: Use cleanup section to remove test data
2. **Tag Appropriately**: Use standard tags for test selection
3. **Validate Discoveries**: Never assume data structures
4. **Handle Failures**: Use `on_error: continue` for non-critical steps
5. **Document Scenarios**: Clear titles and descriptions
6. **Seed Minimally**: Only create necessary test data

## Extending the Framework

### Adding New Tools
1. Implement in `pkg/interface/mcp/tools_*.go`
2. Add to scenario workflows
3. Create specific test scenarios

### Custom Assertions
1. Extend `AssertionEngine` in `assertions.go`
2. Add new operator types
3. Document in this README

### New Chaos Types
1. Configure in Toxiproxy
2. Add to scenario setup
3. Validate resilience

## Performance Considerations

- Scenarios run in parallel (configurable limit)
- Use `parallel` blocks for concurrent steps
- Cache discovery results within scenarios
- Minimize seed data volume

## Security

- API keys are never logged
- Test data is tagged for cleanup
- Scenarios run in isolated contexts
- No production data modification
# E2E Testing Strategy Summary

## Overview

This document summarizes the comprehensive end-to-end testing strategy for the New Relic MCP Server. The strategy emphasizes testing with real New Relic data while maintaining the discovery-first philosophy.

## Key Principles

1. **Real Data Validation**: All tests run against actual New Relic accounts
2. **No Assumptions**: Tests discover data structures rather than assuming them
3. **Adaptive Testing**: Tests adapt to different account schemas
4. **Comprehensive Coverage**: Tests validate all implemented tools and workflows
5. **Production-Like**: Tests simulate real-world usage patterns

## Test Architecture

### Test Accounts

| Account Type | Purpose | Requirements |
|--------------|---------|--------------|
| Primary | Main testing with diverse data | APM, Infrastructure, Browser data |
| Secondary | Cross-account and schema variation | Different naming conventions |
| Empty | Zero-data scenario testing | No active data ingestion |
| High Cardinality | Performance and scale testing | 1000+ services, high volume |

### Test Categories

1. **Discovery Foundation Tests**
   - Event type discovery
   - Attribute profiling
   - Schema adaptation

2. **Query Adaptation Tests**
   - Dynamic query building
   - Missing attribute handling
   - Cross-account queries

3. **Workflow Integration Tests**
   - Performance investigation
   - Incident response
   - Capacity planning

4. **Cross-Account Tests**
   - Account switching
   - Permission boundaries
   - Data isolation

5. **Data Quality Tests**
   - Incomplete data handling
   - Schema evolution
   - High cardinality scenarios

6. **Performance Tests**
   - Response time validation
   - Concurrent operations
   - Memory usage

7. **Error Handling Tests**
   - API error scenarios
   - Data quality issues
   - Recovery mechanisms

8. **Regional Tests**
   - US/EU region validation
   - Cross-region isolation

## Implementation

### Test Framework

```go
// Core test client
type MCPTestClient struct {
    server     *mcp.Server
    account    *TestAccount
    discovery  map[string]interface{}
}

// Workflow execution
type WorkflowStep struct {
    Name            string
    Tool            string
    Params          map[string]interface{}
    StoreAs         string
    ContinueOnError bool
    Validate        func(response map[string]interface{}) error
}
```

### Discovery-First Pattern

Every test follows this pattern:
1. Discover what data exists
2. Profile data characteristics
3. Build adaptive queries
4. Execute and validate
5. Store discoveries for reuse

### Example Test Flow

```go
// 1. Discover event types
eventTypes := discoverEventTypes(ctx)

// 2. Find service identifier (don't assume 'appName')
serviceAttr := discoverServiceAttribute(ctx, eventTypes)

// 3. Build query using discoveries
query := buildAdaptiveQuery(serviceAttr, intent)

// 4. Execute and validate
result := executeQuery(ctx, query)
validateResult(result, expectations)
```

## Execution Strategy

### Local Development
```bash
# One-time setup
make test-e2e-setup
cp .env.test.example .env.test
# Edit .env.test with credentials

# Run tests
make test-e2e
```

### CI/CD Pipeline
- Runs on every PR
- Nightly full suite execution
- Release validation
- Performance benchmarking

### Test Scheduling
- **Daily**: Core discovery and workflow tests
- **4-Hourly**: Smoke tests on primary account
- **Weekly**: Full suite including chaos tests
- **Release**: Comprehensive validation

## Success Metrics

### Coverage
- 100% of implemented tools tested
- All documented workflows validated
- 95%+ error scenario coverage
- All regions and account types

### Quality
- Zero false positives
- <2% test flakiness
- Clear failure diagnostics
- Reproducible results

### Performance
- Discovery < 5s
- Queries < 10s
- Workflows < 5 minutes
- Memory < 500MB

## Key Differentiators

1. **True Discovery-First**: Tests never assume data structures
2. **Real Data**: No mocked New Relic responses
3. **Adaptive**: Tests work across different schemas
4. **Comprehensive**: Validates entire user journeys
5. **Maintainable**: Self-documenting test patterns

## Maintenance

### Test Data
- No manual data setup required
- Tests discover existing data
- Optional data generation for specific scenarios
- Automatic cleanup of test artifacts

### Test Evolution
- New tests for each feature
- Regression tests for bugs
- Performance baseline updates
- Documentation synchronization

## Benefits

1. **Confidence**: Ensures MCP Server works with any New Relic account
2. **Quality**: Catches issues before production
3. **Documentation**: Tests serve as usage examples
4. **Performance**: Validates scalability
5. **Reliability**: Ensures graceful error handling

## Getting Started

1. Review [E2E Testing Strategy](./e2e-testing-strategy.md)
2. Set up test accounts
3. Configure `.env.test`
4. Run `make test-e2e-setup`
5. Execute `make test-e2e`

## Conclusion

This E2E testing strategy ensures the MCP Server truly embodies the discovery-first philosophy while working reliably with real New Relic data. By testing against actual accounts with diverse data patterns, we validate that the system adapts to any environment without making assumptions about data structures or naming conventions.

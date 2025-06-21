# End-to-End Test Suite

This directory contains comprehensive end-to-end tests that validate the MCP Server against real New Relic data.

## Quick Start

```bash
# Run all E2E tests
make test-e2e

# Run specific test category
make test-e2e-discovery
make test-e2e-workflows
make test-e2e-performance

# Run with specific account
E2E_ACCOUNT=primary make test-e2e

# Run specific test
go test -v -run TestDiscoveryE2E/TestEventTypeDiscovery ./tests/e2e/...
```

## Test Structure

```
tests/e2e/
├── README.md                    # This file
├── discovery/                   # Discovery-first tests
│   ├── event_types_test.go     # Event type discovery
│   ├── attributes_test.go      # Attribute profiling
│   └── adaptation_test.go      # Query adaptation
├── workflows/                   # Complete workflow tests
│   ├── performance_test.go     # Performance investigation
│   ├── incident_test.go        # Incident response
│   └── capacity_test.go        # Capacity planning
├── integration/                 # Cross-tool integration
│   ├── multi_account_test.go   # Multi-account scenarios
│   └── correlation_test.go     # Cross-data correlation
├── resilience/                  # Error handling & recovery
│   ├── api_errors_test.go      # API error scenarios
│   ├── data_quality_test.go    # Bad data handling
│   └── timeout_test.go         # Timeout scenarios
├── performance/                 # Performance validation
│   ├── scale_test.go           # High volume tests
│   ├── concurrent_test.go      # Concurrency tests
│   └── memory_test.go          # Memory usage tests
├── framework/                   # Test utilities
│   ├── client.go               # MCP test client
│   ├── accounts.go             # Test account management
│   ├── assertions.go           # Custom assertions
│   └── generators.go           # Test data generation
└── testdata/                   # Test fixtures
    ├── schemas/                # Known schemas
    ├── queries/                # Test queries
    └── expected/               # Expected results
```

## Test Accounts

The E2E tests require multiple New Relic accounts with different characteristics:

### Primary Account
- **Purpose**: Main testing with diverse data
- **Required Data**:
  - APM data from 3+ services
  - Infrastructure monitoring
  - Browser monitoring
  - Custom events
  - 30+ days of history

### Secondary Account  
- **Purpose**: Cross-account testing
- **Required Data**:
  - Different naming conventions
  - Limited permissions
  - Unique event types

### Empty Account
- **Purpose**: Zero-data scenarios
- **Required Data**: None (must be empty)

### High-Cardinality Account
- **Purpose**: Performance testing
- **Required Data**:
  - 1000+ unique services
  - High event volume
  - Complex attributes

## Writing E2E Tests

### Test Template

```go
func (s *DiscoveryE2ESuite) TestYourScenario() {
    ctx := context.Background()
    
    // 1. Discovery Phase - Never assume
    discovered := s.discoverRequiredData(ctx)
    s.Require().NotEmpty(discovered, "Should discover required data")
    
    // 2. Validation Phase - Verify discoveries
    s.validateDiscoveries(ctx, discovered)
    
    // 3. Action Phase - Use discoveries
    result := s.executeBasedOnDiscovery(ctx, discovered)
    
    // 4. Assertion Phase - Validate results
    s.assertExpectedOutcome(result)
}
```

### Discovery-First Principles

1. **Never hardcode attributes**
   ```go
   // BAD
   query := "SELECT count(*) FROM Transaction WHERE appName = 'test'"
   
   // GOOD
   serviceAttr := s.discoverServiceAttribute(ctx)
   query := fmt.Sprintf("SELECT count(*) FROM Transaction WHERE %s = 'test'", serviceAttr)
   ```

2. **Always validate existence**
   ```go
   // Check if data exists before querying
   exists := s.checkDataExists(ctx, "Transaction", "24 hours")
   if !exists {
       s.T().Skip("No Transaction data available")
   }
   ```

3. **Adapt to discovered schema**
   ```go
   schema := s.discoverSchema(ctx, "Transaction")
   query := s.buildAdaptiveQuery(schema, intent)
   ```

## Running Tests

### Local Development

```bash
# Set up test environment
cp .env.test.example .env.test
# Edit .env.test with your New Relic credentials

# Run tests
source .env.test
go test -v ./tests/e2e/...
```

### CI/CD Pipeline

Tests run automatically on:
- Every pull request
- Nightly scheduled runs
- Release candidates

### Test Reports

After execution, find reports in:
- `tests/results/e2e-report.html` - Human-readable report
- `tests/results/coverage.json` - Tool coverage data
- `tests/results/performance.json` - Performance metrics

## Debugging Failed Tests

### Enable Verbose Logging

```bash
E2E_LOG_LEVEL=debug go test -v -run TestName ./tests/e2e/...
```

### Capture MCP Traffic

```bash
E2E_CAPTURE_TRAFFIC=true go test -v -run TestName ./tests/e2e/...
# Traffic saved to tests/results/traffic/
```

### Use Test Inspector

```bash
# Run specific test with inspector
make test-e2e-inspect TEST=TestEventTypeDiscovery
```

## Best Practices

1. **Isolate test data** - Use unique identifiers
2. **Clean up after tests** - Remove test artifacts
3. **Handle timeouts gracefully** - Set appropriate limits
4. **Document failures clearly** - Include discovery context
5. **Make tests repeatable** - No order dependencies

## Troubleshooting

### Common Issues

1. **"No data found"**
   - Ensure test account has required data
   - Check time ranges in queries
   - Verify account permissions

2. **"Timeout exceeded"**
   - Increase test timeout
   - Check account data volume
   - Optimize discovery queries

3. **"Schema mismatch"**
   - Account may have custom schemas
   - Update test to be more adaptive
   - Add schema variation handling

### Getting Help

- Check test logs in `tests/results/logs/`
- Review captured MCP traffic
- Consult discovery results
- Ask in #mcp-server channel

## Contributing

When adding new E2E tests:

1. Follow discovery-first principles
2. Test with multiple account types
3. Include negative test cases
4. Document test purpose clearly
5. Ensure tests are idempotent
6. Add to appropriate test category
7. Update this README if needed
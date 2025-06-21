# Comprehensive Testing Strategy for MCP Server New Relic

## Overview

This document outlines the multi-layered testing approach for the MCP Server New Relic project, focusing on validating the Discovery-First philosophy with real New Relic data.

## Testing Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        E2E Tests (Real NR Data)                  │
│  - Real accounts, actual API calls, no mocks                    │
│  - Complex workflows, multi-account scenarios                    │
│  - Performance benchmarks, chaos engineering                     │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────┐
│                    Integration Tests (Mock Mode)                 │
│  - Component integration, protocol compliance                    │
│  - Tool orchestration, error handling                           │
│  - Build tags: integration, nodiscovery                         │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────┐
│                        Unit Tests                                │
│  - Individual functions, algorithms, utilities                   │
│  - Fast, isolated, high coverage                               │
│  - No external dependencies                                     │
└─────────────────────────────────────────────────────────────────┘
```

## Test Types

### 1. Unit Tests (`*_test.go`)

**Purpose**: Test individual components in isolation

**Characteristics**:
- No external dependencies
- Mock all interfaces
- Fast execution (<1ms per test)
- High code coverage (>80%)

**Example**:
```go
func TestCalculateZScore(t *testing.T) {
    values := []float64{1, 2, 3, 4, 5, 100} // 100 is an outlier
    zScores := calculateZScores(values)
    
    // Last value should have high z-score
    assert.Greater(t, math.Abs(zScores[5]), 2.0)
}
```

**Run**: `make test-unit`

### 2. Integration Tests (`*_integration_test.go`)

**Purpose**: Test component interactions with mock data

**Characteristics**:
- Uses build tags (`//go:build integration`)
- Tests MCP protocol compliance
- Validates tool orchestration
- Uses mock New Relic client

**Key Areas**:
- Protocol handling (JSON-RPC 2.0)
- Tool registry and execution
- Session management
- Transport layers (stdio, HTTP, SSE)

**Example**:
```go
//go:build integration

func TestDiscoveryWorkflow(t *testing.T) {
    server := createTestServer(t)
    
    // Discover event types
    result1 := executeTool(server, "discovery.explore_event_types", params)
    
    // Use discovered type for attribute exploration
    eventType := extractEventType(result1)
    result2 := executeTool(server, "discovery.explore_attributes", 
        map[string]interface{}{"event_type": eventType})
    
    // Validate workflow completion
    assert.NotEmpty(t, result2["attributes"])
}
```

**Run**: `make test-integration`

### 3. End-to-End Tests (`tests/e2e/`)

**Purpose**: Validate real-world scenarios with actual New Relic data

**Characteristics**:
- Uses real New Relic accounts
- No mocks - actual API calls
- Tests discovery-first philosophy
- Validates adaptive behavior

**Test Categories**:

#### Discovery Tests
- Event type exploration
- Attribute profiling
- Schema drift handling
- Missing field adaptation

#### Workflow Tests
- Incident investigation
- Performance analysis
- Cost optimization
- Compliance validation

#### Resilience Tests
- Network failures (Toxiproxy)
- Rate limiting
- Timeout handling
- Retry logic

#### Performance Tests
- Query optimization
- Parallel execution
- Memory usage
- Response times

**Run**: `make test-e2e`

## Test Data Strategy

### 1. Unit Tests
- Hardcoded test data
- Generated data (faker)
- Edge cases and boundaries

### 2. Integration Tests
- Mock data generator
- Consistent schemas
- Predictable responses

### 3. E2E Tests
- Real New Relic accounts
- Multiple account types:
  - Primary: Standard data
  - Secondary: Different schemas
  - Empty: Zero-data scenarios
  - High-cardinality: Scale testing

## Testing the Discovery-First Philosophy

### Principle 1: Never Assume Data Exists

**Test Approach**:
```yaml
# Scenario: Missing attributes
name: discovery-missing-attributes
steps:
  - action: discovery.explore_event_types
  - action: query_nrdb
    params:
      query: "SELECT missingField FROM Transaction"
    expect_error: true
  - action: discovery.explore_attributes
    params:
      event_type: Transaction
  - action: nrql.execute  # Adaptive query
    params:
      query: "SELECT missingField FROM Transaction"
      validate_schema: true
    expect:
      adapted_query: "SELECT count(*) FROM Transaction"
```

### Principle 2: Build Knowledge Incrementally

**Test Approach**:
- Start with zero knowledge
- Discover available data
- Build queries based on discoveries
- Cache and reuse knowledge

### Principle 3: Adapt to Reality

**Test Approach**:
- Test with accounts having different schemas
- Validate query adaptation
- Ensure graceful degradation

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make test-unit
      
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make test-integration
      
  e2e-tests:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - name: Run E2E Tests
        env:
          NEW_RELIC_API_KEY_PRIMARY: ${{ secrets.NR_API_KEY }}
          NEW_RELIC_ACCOUNT_ID_PRIMARY: ${{ secrets.NR_ACCOUNT_ID }}
        run: make test-e2e
```

## Test Execution

### Local Development

```bash
# Quick feedback during development
make test          # Unit tests only
make test-race     # With race detection

# Before committing
make test-all      # Unit + Integration

# Before releasing
make test-e2e      # Full E2E suite
```

### CI/CD Pipeline

1. **On Every Commit**: Unit tests
2. **On Pull Request**: Unit + Integration tests
3. **On Main Branch**: Full test suite including E2E
4. **Nightly**: Extended E2E with chaos testing

## Coverage Requirements

### Minimum Coverage by Type

| Component | Unit | Integration | E2E |
|-----------|------|-------------|-----|
| Core Logic | 85% | 70% | - |
| Discovery Engine | 90% | 80% | 100% |
| Analysis Tools | 85% | 75% | 90% |
| Error Handling | 95% | 85% | - |
| API Client | 80% | - | 100% |

### Measuring Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
make coverage-html
```

## Best Practices

### 1. Test Naming
```go
// Good
func TestDiscoveryEngine_ExploreEventTypes_WithEmptyAccount(t *testing.T)
func TestAnalysis_DetectAnomalies_HighSensitivity(t *testing.T)

// Bad
func TestDiscovery(t *testing.T)
func TestIt(t *testing.T)
```

### 2. Test Organization
```go
func TestFeature(t *testing.T) {
    t.Run("happy path", func(t *testing.T) {
        // Normal operation
    })
    
    t.Run("error cases", func(t *testing.T) {
        t.Run("missing parameter", func(t *testing.T) {})
        t.Run("invalid input", func(t *testing.T) {})
    })
    
    t.Run("edge cases", func(t *testing.T) {
        t.Run("empty data", func(t *testing.T) {})
        t.Run("large dataset", func(t *testing.T) {})
    })
}
```

### 3. Assertions
```go
// Use descriptive assertion messages
assert.NotEmpty(t, eventTypes, "Should discover event types even in new account")
assert.Contains(t, result, "adapted_query", "Should adapt query when field missing")

// Validate entire structures
assert.Equal(t, expected, actual, "Complete response should match")
```

### 4. Test Data
```go
// Use test fixtures for consistency
fixture := loadTestFixture(t, "discovery/event_types.json")

// Generate data for edge cases
largeDataset := generateTestData(10000)

// Use meaningful test data
testAccount := TestAccount{
    ID: "test-123",
    Name: "Integration Test Account",
    Region: "US",
}
```

## Debugging Failed Tests

### 1. Verbose Output
```bash
# Run with verbose logging
make test-e2e E2E_LOG_LEVEL=debug

# Run specific test with details
go test -v -run TestDiscovery/ExploreEventTypes ./tests/e2e/...
```

### 2. Capture Traffic
```bash
# Enable request/response capture
make test-e2e E2E_CAPTURE_TRAFFIC=true

# View captured data
cat tests/results/traffic/*.json
```

### 3. Interactive Debugging
```go
// Add debug breakpoint
t.Logf("Current state: %+v", state)
if os.Getenv("DEBUG") != "" {
    time.Sleep(30 * time.Second) // Pause for inspection
}
```

## Test Maintenance

### Weekly Tasks
1. Review flaky tests
2. Update test data
3. Check coverage trends

### Monthly Tasks
1. Audit test scenarios
2. Update mock data
3. Performance baseline review

### Quarterly Tasks
1. Full E2E suite review
2. Test strategy assessment
3. Tool updates

## Conclusion

This comprehensive testing strategy ensures the MCP Server New Relic implementation:
- Never makes assumptions about data
- Adapts to real-world schemas
- Handles failures gracefully
- Performs efficiently at scale

The multi-layered approach provides fast feedback during development while maintaining confidence through real-world validation.

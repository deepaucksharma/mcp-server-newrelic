# Testing Guide for New Relic MCP Server

This comprehensive guide covers all aspects of testing the New Relic MCP Server, from unit tests to performance benchmarks. Our testing philosophy emphasizes reliability, maintainability, and comprehensive coverage.

See also the [Comprehensive Testing Strategy](./comprehensive-testing-strategy.md) for an overview of how all test layers fit together.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Environment Setup](#test-environment-setup)
3. [Unit Testing](#unit-testing)
4. [Integration Testing](#integration-testing)
5. [End-to-End Testing](#end-to-end-testing)
6. [Performance Testing](#performance-testing)
7. [Mock Mode Testing](#mock-mode-testing)
8. [Test Data Management](#test-data-management)
9. [CI/CD Integration](#cicd-integration)
10. [Test Coverage Requirements](#test-coverage-requirements)

## Testing Philosophy

Our testing approach follows these core principles:

1. **Test at the Right Level**: Unit tests for business logic, integration tests for API contracts, E2E tests for critical workflows
2. **Mock External Dependencies**: All tests should run without requiring New Relic credentials
3. **Fast Feedback**: Tests should run quickly to encourage frequent execution
4. **Clear Failure Messages**: When tests fail, the reason should be immediately obvious
5. **Deterministic Results**: Tests must be reliable and produce consistent results
6. **Production-Like Testing**: Mock mode should closely simulate real behavior
7. **Zero Assumptions**: The lint stage runs `scripts/assumption_scan.sh` to ensure no hard-coded field names leak into production code

## Test Environment Setup

### Prerequisites

```bash
# Install testing dependencies
go install github.com/stretchr/testify@latest
go install github.com/vektra/mockery/v2@latest
go install gotest.tools/gotestsum@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install MCP testing tools
npm install -g @modelcontextprotocol/inspector
```

### Environment Configuration

Create a `.env.test` file for test-specific configuration:

```bash
# Test environment variables
NEW_RELIC_ACCOUNT_ID=123456
NEW_RELIC_API_KEY=test-key-nrak-abcdef123456
NEW_RELIC_REGION=US
NEW_RELIC_BASE_URL=https://api.newrelic.com

# Test-specific settings
LOG_LEVEL=ERROR
STATE_STORE_TYPE=memory
CACHE_TTL=60s
REQUEST_TIMEOUT=5s

# Disable external connections in tests
MOCK_MODE=true
```

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage

# Run specific package tests
go test -v ./pkg/interface/mcp/

# Run specific test
go test -v -run TestHandleQueryNRDB ./pkg/interface/mcp/

# Run tests with race detection
go test -race ./...

# Run benchmarks
make test-benchmark
```

## Unit Testing

### Unit Test Structure

Every unit test should follow this pattern:

```go
package mcp

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHandleQueryNRDB(t *testing.T) {
    // Arrange
    tests := []struct {
        name    string
        params  map[string]interface{}
        setup   func(*Server)
        wantErr bool
        errMsg  string
        validate func(t *testing.T, result interface{})
    }{
        {
            name: "valid query with results",
            params: map[string]interface{}{
                "query": "SELECT count(*) FROM Transaction",
            },
            setup: func(s *Server) {
                // Setup mock client if needed
            },
            wantErr: false,
            validate: func(t *testing.T, result interface{}) {
                data, ok := result.(map[string]interface{})
                require.True(t, ok, "result should be a map")
                assert.Contains(t, data, "results")
            },
        },
        {
            name:    "missing query parameter",
            params:  map[string]interface{}{},
            wantErr: true,
            errMsg:  "query parameter is required",
        },
        {
            name: "invalid query syntax",
            params: map[string]interface{}{
                "query": "INVALID NRQL SYNTAX",
            },
            wantErr: true,
            errMsg:  "invalid NRQL syntax",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            s := &Server{} // Mock mode by default
            if tt.setup != nil {
                tt.setup(s)
            }

            // Act
            result, err := s.handleQueryNRDB(context.Background(), tt.params)

            // Assert
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
                if tt.validate != nil {
                    tt.validate(t, result)
                }
            }
        })
    }
}
```

### Testing Best Practices

1. **Table-Driven Tests**: Use test tables for comprehensive coverage
2. **Descriptive Names**: Test names should describe the scenario being tested
3. **Arrange-Act-Assert**: Clear structure for test readability
4. **Test One Thing**: Each test should verify a single behavior
5. **Use require for Preconditions**: Use `require` for setup assertions, `assert` for test assertions

### Mocking Dependencies

```go
// Mock New Relic client
type mockNRClient struct {
    queryFunc func(ctx context.Context, query string) (interface{}, error)
}

func (m *mockNRClient) Query(ctx context.Context, query string) (interface{}, error) {
    if m.queryFunc != nil {
        return m.queryFunc(ctx, query)
    }
    return map[string]interface{}{
        "results": []map[string]interface{}{
            {"count": 100},
        },
    }, nil
}

// Use in tests
func TestWithMockClient(t *testing.T) {
    mockClient := &mockNRClient{
        queryFunc: func(ctx context.Context, query string) (interface{}, error) {
            return map[string]interface{}{"mocked": true}, nil
        },
    }
    
    server := &Server{nrClient: mockClient}
    // Test server behavior
}
```

## Integration Testing

### MCP Protocol Testing

Test the full MCP request/response cycle:

```go
func TestMCPProtocolIntegration(t *testing.T) {
    // Start test server
    server := setupTestServer(t)
    defer server.Stop(context.Background())

    // Create MCP client
    client := newTestMCPClient(t, server.Address())

    // Test tool discovery
    t.Run("list tools", func(t *testing.T) {
        resp, err := client.Call("tools/list", nil)
        require.NoError(t, err)
        
        tools, ok := resp["tools"].([]interface{})
        require.True(t, ok)
        assert.Greater(t, len(tools), 0)
    })

    // Test tool execution
    t.Run("execute query", func(t *testing.T) {
        params := map[string]interface{}{
            "query": "SELECT count(*) FROM Transaction",
        }
        
        resp, err := client.Call("tools/call", map[string]interface{}{
            "name": "query_nrdb",
            "arguments": params,
        })
        require.NoError(t, err)
        assert.Contains(t, resp, "content")
    })
}
```

### API Contract Testing

```go
func TestAPIContracts(t *testing.T) {
    // Test each tool's parameter validation
    testCases := []struct {
        tool   string
        params map[string]interface{}
        valid  bool
    }{
        {
            tool: "query_nrdb",
            params: map[string]interface{}{
                "query": "SELECT * FROM Transaction",
                "timeout": 30,
            },
            valid: true,
        },
        {
            tool: "query_nrdb",
            params: map[string]interface{}{
                "timeout": 30, // Missing required query
            },
            valid: false,
        },
    }

    server := setupTestServer(t)
    for _, tc := range testCases {
        t.Run(tc.tool, func(t *testing.T) {
            _, err := server.ExecuteTool(context.Background(), tc.tool, tc.params)
            if tc.valid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

## End-to-End Testing

### Critical Workflow Tests

```go
func TestE2EQueryWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // Setup
    server := startRealServer(t)
    client := createE2EClient(t, server.URL)

    // Test complete query workflow
    t.Run("query and analyze", func(t *testing.T) {
        // 1. Check query syntax
        checkResp, err := client.Call("query_check", map[string]interface{}{
            "query": "SELECT average(duration) FROM Transaction FACET appName",
        })
        require.NoError(t, err)
        assert.True(t, checkResp["valid"].(bool))

        // 2. Execute query
        queryResp, err := client.Call("query_nrdb", map[string]interface{}{
            "query": "SELECT average(duration) FROM Transaction FACET appName",
        })
        require.NoError(t, err)
        assert.NotEmpty(t, queryResp["results"])

        // 3. Build dashboard from results
        dashResp, err := client.Call("generate_dashboard", map[string]interface{}{
            "name": "E2E Test Dashboard",
            "queries": []string{
                "SELECT average(duration) FROM Transaction FACET appName",
            },
        })
        require.NoError(t, err)
        assert.NotEmpty(t, dashResp["dashboardId"])
    })
}
```

### Scenario-Based Testing

```go
func TestE2EAlertManagement(t *testing.T) {
    // Test complete alert lifecycle
    ctx := context.Background()
    server := setupE2EEnvironment(t)

    // 1. Create alert policy
    policyID := createTestPolicy(t, server)

    // 2. Add conditions
    conditionID := addAlertCondition(t, server, policyID, AlertCondition{
        Name: "High Error Rate",
        Query: "SELECT percentage(count(*), WHERE error = true) FROM Transaction",
        Threshold: 5.0,
    })

    // 3. Test alert triggering
    triggerTestAlert(t, server, conditionID)

    // 4. Verify incident created
    incidents := getIncidents(t, server, policyID)
    assert.Len(t, incidents, 1)

    // 5. Cleanup
    deleteTestPolicy(t, server, policyID)
}
```

## Performance Testing

### Benchmark Tests

```go
func BenchmarkQueryNRDB(b *testing.B) {
    server := &Server{} // Mock mode
    params := map[string]interface{}{
        "query": "SELECT count(*) FROM Transaction WHERE duration > 1",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := server.handleQueryNRDB(context.Background(), params)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkConcurrentQueries(b *testing.B) {
    server := setupBenchmarkServer(b)
    queries := []string{
        "SELECT count(*) FROM Transaction",
        "SELECT average(duration) FROM Transaction",
        "SELECT percentile(duration, 95) FROM Transaction",
    }

    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            query := queries[i%len(queries)]
            _, err := server.handleQueryNRDB(context.Background(), map[string]interface{}{
                "query": query,
            })
            if err != nil {
                b.Error(err)
            }
            i++
        }
    })
}
```

### Load Testing

```go
func TestLoadHandling(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test")
    }

    server := setupTestServer(t)
    
    // Configure load test
    concurrency := 100
    requestsPerClient := 50
    
    var wg sync.WaitGroup
    errors := make(chan error, concurrency*requestsPerClient)
    
    start := time.Now()
    
    // Launch concurrent clients
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func(clientID int) {
            defer wg.Done()
            
            for j := 0; j < requestsPerClient; j++ {
                _, err := server.handleQueryNRDB(context.Background(), map[string]interface{}{
                    "query": fmt.Sprintf("SELECT count(*) FROM Transaction WHERE client = %d", clientID),
                })
                if err != nil {
                    errors <- err
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    duration := time.Since(start)
    
    // Analyze results
    errorCount := len(errors)
    totalRequests := concurrency * requestsPerClient
    successRate := float64(totalRequests-errorCount) / float64(totalRequests) * 100
    
    t.Logf("Load test completed in %v", duration)
    t.Logf("Total requests: %d", totalRequests)
    t.Logf("Success rate: %.2f%%", successRate)
    t.Logf("Requests/second: %.2f", float64(totalRequests)/duration.Seconds())
    
    // Assert performance requirements
    assert.Greater(t, successRate, 99.0, "Success rate should be > 99%")
    assert.Less(t, duration.Seconds(), 60.0, "Should complete within 60 seconds")
}
```

## Mock Mode Testing

### Mock Data Generation

```go
func TestMockDataConsistency(t *testing.T) {
    server := &Server{} // Mock mode
    
    // Test deterministic mock data
    t.Run("consistent results", func(t *testing.T) {
        query := "SELECT count(*) FROM Transaction"
        
        result1, err1 := server.handleQueryNRDB(context.Background(), map[string]interface{}{
            "query": query,
        })
        require.NoError(t, err1)
        
        result2, err2 := server.handleQueryNRDB(context.Background(), map[string]interface{}{
            "query": query,
        })
        require.NoError(t, err2)
        
        // Mock data should be consistent for same query
        assert.Equal(t, result1, result2)
    })
    
    // Test varied mock data
    t.Run("varied responses", func(t *testing.T) {
        queries := []string{
            "SELECT count(*) FROM Transaction",
            "SELECT average(duration) FROM Transaction",
            "SELECT * FROM Transaction LIMIT 10",
        }
        
        results := make([]interface{}, len(queries))
        for i, query := range queries {
            result, err := server.handleQueryNRDB(context.Background(), map[string]interface{}{
                "query": query,
            })
            require.NoError(t, err)
            results[i] = result
        }
        
        // Different queries should produce different mock data
        assert.NotEqual(t, results[0], results[1])
        assert.NotEqual(t, results[1], results[2])
    })
}
```

### Mock Mode Validation

```go
func TestMockModeBehavior(t *testing.T) {
    // Ensure mock mode doesn't make external calls
    server := &Server{
        nrClient: nil, // Explicitly no client
    }
    
    tools := []string{
        "query_nrdb",
        "list_dashboards", 
        "discovery.list_schemas",
        "create_alert",
    }
    
    for _, tool := range tools {
        t.Run(tool, func(t *testing.T) {
            handler, exists := server.tools[tool]
            require.True(t, exists, "Tool %s should exist", tool)
            
            // Call with minimal valid params
            result, err := handler(context.Background(), getMinimalParams(tool))
            
            // Should not error in mock mode
            assert.NoError(t, err, "Mock mode should not error")
            assert.NotNil(t, result, "Mock mode should return data")
        })
    }
}
```

## Test Data Management

### Test Fixtures

Create reusable test data in `testdata/` directory:

```go
// testdata/fixtures.go
package testdata

import (
    "encoding/json"
    "io/ioutil"
    "path/filepath"
)

// LoadFixture loads test data from JSON files
func LoadFixture(name string) (map[string]interface{}, error) {
    path := filepath.Join("testdata", name+".json")
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var result map[string]interface{}
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }
    
    return result, nil
}

// Example fixture: testdata/query_response.json
{
    "results": [
        {
            "facet": "web-app",
            "average.duration": 0.234,
            "count": 1523
        },
        {
            "facet": "api-service", 
            "average.duration": 0.156,
            "count": 3421
        }
    ],
    "metadata": {
        "eventTypes": ["Transaction"],
        "messages": [],
        "contents": [
            {
                "function": "average",
                "attribute": "duration"
            }
        ]
    }
}
```

### Test Data Builders

```go
// Test data builder pattern
type QueryResultBuilder struct {
    results []map[string]interface{}
    metadata map[string]interface{}
}

func NewQueryResultBuilder() *QueryResultBuilder {
    return &QueryResultBuilder{
        results: []map[string]interface{}{},
        metadata: map[string]interface{}{},
    }
}

func (b *QueryResultBuilder) WithResult(data map[string]interface{}) *QueryResultBuilder {
    b.results = append(b.results, data)
    return b
}

func (b *QueryResultBuilder) WithEventType(eventType string) *QueryResultBuilder {
    if b.metadata["eventTypes"] == nil {
        b.metadata["eventTypes"] = []string{}
    }
    b.metadata["eventTypes"] = append(b.metadata["eventTypes"].([]string), eventType)
    return b
}

func (b *QueryResultBuilder) Build() map[string]interface{} {
    return map[string]interface{}{
        "results": b.results,
        "metadata": b.metadata,
    }
}

// Usage in tests
func TestWithBuilder(t *testing.T) {
    mockData := NewQueryResultBuilder().
        WithResult(map[string]interface{}{
            "count": 100,
            "appName": "test-app",
        }).
        WithEventType("Transaction").
        Build()
    
    // Use mockData in test
}
```

## CI/CD Integration

### GitHub Actions Workflow

Create `.github/workflows/test.yml`:

```yaml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.21', '1.22']
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: |
        go mod download
        go install gotest.tools/gotestsum@latest
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    
    - name: Run linters
      run: make lint
    
    - name: Run unit tests
      run: |
        gotestsum --junitfile unit-tests.xml -- -coverprofile=unit-coverage.out ./...
    
    - name: Run integration tests
      run: |
        gotestsum --junitfile integration-tests.xml -- -tags=integration -coverprofile=integration-coverage.out ./...
    
    - name: Merge coverage
      run: |
        go install github.com/wadey/gocovmerge@latest
        gocovmerge unit-coverage.out integration-coverage.out > coverage.out
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella
    
    - name: Upload test results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: test-results
        path: |
          unit-tests.xml
          integration-tests.xml
    
    - name: Run benchmarks
      if: github.event_name == 'push' && github.ref == 'refs/heads/main'
      run: |
        go test -bench=. -benchmem ./... | tee benchmark.txt
    
    - name: Comment PR with coverage
      if: github.event_name == 'pull_request'
      uses: 5monkeys/cobertura-action@master
      with:
        path: coverage.out
        minimum_coverage: 75
```

### Pre-commit Hooks

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: go fmt ./...
        language: system
        pass_filenames: false
      
      - id: go-test
        name: go test
        entry: go test ./...
        language: system
        pass_filenames: false
      
      - id: go-lint
        name: golangci-lint
        entry: golangci-lint run
        language: system
        pass_filenames: false
```

## Test Coverage Requirements

### Coverage Goals

- **Overall Coverage**: Minimum 80%
- **Critical Paths**: 95%+ coverage for:
  - MCP protocol handlers
  - Tool implementations
  - Error handling paths
  - State management
  
### Coverage by Package

| Package | Target | Priority |
|---------|--------|----------|
| pkg/interface/mcp | 90% | Critical |
| pkg/discovery | 85% | High |
| pkg/validation | 95% | Critical |
| pkg/state | 80% | High |
| pkg/newrelic | 75% | Medium |
| pkg/config | 70% | Low |

### Measuring Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Check specific package coverage
go test -cover ./pkg/interface/mcp/

# Exclude test files from coverage
go test -coverprofile=coverage.out -coverpkg=./pkg/... ./...
```

### Coverage Enforcement

```go
// scripts/check-coverage.go
package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

func main() {
    file, err := os.Open("coverage.out")
    if err != nil {
        fmt.Println("No coverage file found")
        os.Exit(1)
    }
    defer file.Close()

    packages := make(map[string]float64)
    scanner := bufio.NewScanner(file)
    
    // Parse coverage data
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "mode:") {
            continue
        }
        
        parts := strings.Fields(line)
        if len(parts) < 3 {
            continue
        }
        
        pkg := strings.Split(parts[0], "/")[0]
        coverage, _ := strconv.ParseFloat(strings.TrimSuffix(parts[2], "%"), 64)
        packages[pkg] = coverage
    }
    
    // Check thresholds
    failed := false
    thresholds := map[string]float64{
        "pkg/interface/mcp": 90.0,
        "pkg/discovery":     85.0,
        "pkg/validation":    95.0,
        "pkg/state":        80.0,
    }
    
    for pkg, threshold := range thresholds {
        coverage := packages[pkg]
        if coverage < threshold {
            fmt.Printf("❌ %s: %.1f%% (required: %.1f%%)\n", pkg, coverage, threshold)
            failed = true
        } else {
            fmt.Printf("✅ %s: %.1f%%\n", pkg, coverage)
        }
    }
    
    if failed {
        os.Exit(1)
    }
}
```

## Testing Checklist

Before submitting a PR, ensure:

- [ ] All tests pass: `make test`
- [ ] Coverage meets requirements: `make test-coverage`
- [ ] No linting errors: `make lint`
- [ ] No race conditions: `go test -race ./...`
- [ ] Mock mode works correctly
- [ ] Integration tests pass (if applicable)
- [ ] Performance benchmarks show no regression
- [ ] New features have corresponding tests
- [ ] Test documentation is updated

## Troubleshooting

### Common Test Issues

1. **Flaky Tests**
   - Add retries for network-dependent tests
   - Use deterministic time in tests
   - Avoid hardcoded sleep statements

2. **Slow Tests**
   - Use `t.Parallel()` for independent tests
   - Mock expensive operations
   - Use test short mode for quick checks

3. **Coverage Gaps**
   - Check error paths
   - Test edge cases
   - Verify timeout handling

4. **Mock Mode Issues**
   - Ensure all tools support nil client
   - Verify mock data is realistic
   - Test mock mode explicitly

## Additional Resources

- [Go Testing Best Practices](https://golang.org/doc/tutorial/add-a-test)
- [Testify Documentation](https://github.com/stretchr/testify)
- [MCP Testing Tools](https://modelcontextprotocol.io/docs/tools/testing)
- [New Relic API Documentation](https://docs.newrelic.com/docs/apis/nerdgraph/get-started/introduction-new-relic-nerdgraph/)
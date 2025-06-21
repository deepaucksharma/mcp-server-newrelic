# Mock Mode Guide

The New Relic MCP Server provides a comprehensive mock mode for development and testing without requiring a New Relic connection.

## Overview

Mock mode allows you to:
- Develop and test AI integrations without New Relic credentials
- Get realistic responses for all tools
- Test error scenarios and edge cases
- Demo the server capabilities

## Enabling Mock Mode

### Method 1: No Credentials
Simply run the server without New Relic credentials:
```bash
./bin/mcp-server
# Server will automatically run in mock mode
```

### Method 2: Environment Variable
Explicitly enable mock mode:
```bash
export MOCK_MODE=true
./bin/mcp-server
```

### Method 3: Command Line Flag
```bash
./bin/mcp-server --mock-mode
```

### Method 4: Docker
```bash
docker run -e MOCK_MODE=true mcp-server-newrelic
```

## Mock Data Characteristics

### Query Tools
- **query_nrdb**: Returns realistic NRQL results based on query patterns
  - COUNT queries return random counts
  - AVERAGE queries return random averages
  - FACET queries return multiple grouped results
  - Time series queries return temporal data

### Discovery Tools
- **discovery.explore_event_types**: Returns common event types
  - Transaction, SystemSample, PageView, JavaScriptError, CustomEvent
  - Realistic event counts and timestamps
  
- **discovery.explore_attributes**: Returns typical attributes
  - Standard attributes with appropriate data types
  - Coverage percentages and null ratios
  - Example values for each attribute

### Dashboard Tools
- **list_dashboards**: Returns sample dashboards
  - Production Overview, Application Performance, Infrastructure Health
  - Realistic permissions and timestamps
  
- **generate_dashboard**: Creates dashboards from templates
  - Golden signals template with 4 key widgets
  - Proper widget layouts and NRQL queries

### Alert Tools
- **create_alert**: Simulates alert creation
  - Generates alert IDs and baseline calculations
  - Returns threshold recommendations
  
- **analyze_alerts**: Provides alert effectiveness metrics
  - False positive rates and MTTA statistics
  - Threshold adjustment recommendations

### Analysis Tools
- **calculate_baseline**: Returns statistical baselines
  - Percentiles, standard deviations, and confidence intervals
  - Daily and weekly pattern detection
  
- **detect_anomalies**: Generates anomaly data
  - Random anomalies with severity levels
  - Deviation scores and timestamps

## Example Mock Responses

### NRQL Query
```json
{
  "tool": "query_nrdb",
  "params": {
    "query": "SELECT count(*) FROM Transaction"
  }
}

// Returns:
{
  "results": [{"count": 5432}],
  "metadata": {
    "eventTypes": ["Transaction"],
    "messages": []
  },
  "performanceInfo": {
    "inspectedCount": 45678,
    "matchedCount": 5432,
    "wallClockTime": 234
  }
}
```

### Event Type Discovery
```json
{
  "tool": "discovery.explore_event_types"
}

// Returns:
{
  "event_types": [
    {
      "name": "Transaction",
      "count": 567890,
      "attributes": 45,
      "sample_timestamp": "2023-12-20T10:30:00Z"
    },
    {
      "name": "SystemSample",
      "count": 234567,
      "attributes": 32,
      "sample_timestamp": "2023-12-20T10:28:00Z"
    }
  ],
  "total": 5
}
```

## Customizing Mock Data

### Environment Variables
```bash
# Control mock data randomness
export MOCK_SEED=12345

# Set mock data ranges
export MOCK_MIN_COUNT=1000
export MOCK_MAX_COUNT=10000
```

### Configuration File
```yaml
mock:
  seed: 12345
  ranges:
    counts:
      min: 1000
      max: 10000
    percentages:
      min: 0
      max: 100
    durations:
      min: 0.1
      max: 5.0
```

## Testing with Mock Mode

### Unit Tests
```go
func TestToolWithMockData(t *testing.T) {
    server := NewServer(ServerConfig{MockMode: true})
    
    result, err := server.HandleTool(ctx, "query_nrdb", params)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Tests
```bash
# Run server in mock mode
MOCK_MODE=true ./bin/mcp-server &

# Test with MCP client
echo '{"method": "tools/call", "params": {"name": "query_nrdb", "arguments": {"query": "SELECT count(*) FROM Transaction"}}}' | nc localhost 8080
```

## Mock Mode Limitations

1. **No Data Persistence**: Each request generates new random data
2. **No Real Relationships**: Cross-references between entities are simulated
3. **No Time Series Continuity**: Historical data doesn't maintain trends
4. **Simplified Error Scenarios**: Only common errors are simulated

## Best Practices

1. **Use for Development**: Perfect for building AI integrations
2. **Not for Performance Testing**: Mock data generation is optimized for realism, not speed
3. **Validate with Real Data**: Always test with actual New Relic data before production
4. **Document Differences**: Note any behavioral differences between mock and real modes

## Troubleshooting

### Detecting Mock Mode
Look for these indicators:
- Log message: "Running in MOCK MODE - no New Relic connection"
- Response metadata may include: `"mock": true`
- No network requests to New Relic APIs

### Common Issues
- **Too Predictable**: Increase randomness with different seeds
- **Unrealistic Values**: Adjust mock data ranges
- **Missing Tool Support**: Check if tool has mock implementation

## Future Enhancements

- Record and replay real responses
- Scenario-based mock data
- Error injection for chaos testing
- Mock data templates

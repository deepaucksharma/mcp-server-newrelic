# Anti-Patterns and First Principles Fixes

## 1. Hardcoded Event Types

### ❌ Current Anti-Pattern
```go
// pkg/interface/mcp/tools_dashboard.go:368
query: "SELECT average(duration) FROM Transaction WHERE appName = 'TestService'"

// pkg/interface/mcp/tools_dashboard.go:407
query: "SELECT average(cpuPercent) FROM SystemSample"
```

### ✅ First Principles Approach
```go
// First discover what event types exist
eventTypes := discoveryEngine.ListSchemas(ctx, SchemaFilter{})

// Find transaction-like events dynamically
transactionEvent := discoveryEngine.FindEventWithAttribute(ctx, "duration", "response_time")

// Build query based on discovery
query := fmt.Sprintf("SELECT average(%s) FROM %s", 
    transactionEvent.DurationAttribute,
    transactionEvent.Name)
```

## 2. Assumed Attributes

### ❌ Current Anti-Pattern
```go
// pkg/interface/mcp/tools_dashboard.go:407
"WHERE apmApplicationNames LIKE '%TestService%'"  // Assumes this attribute exists
```

### ✅ First Principles Approach
```go
// Discover linking attributes between APM and Infrastructure
linkingAttrs := discoveryEngine.FindCommonAttributes(ctx, "Transaction", "SystemSample")

// Use discovered attribute
whereClause := ""
if linkAttr := linkingAttrs.GetBestMatch(); linkAttr != "" {
    whereClause = fmt.Sprintf("WHERE %s LIKE '%%%s%%'", linkAttr, serviceName)
}
```

## 3. Static Widget Type Selection

### ❌ Current Anti-Pattern
```go
// pkg/interface/mcp/tools_dashboard.go:362
"type": "line",  // Always uses line chart
```

### ✅ First Principles Approach
```go
// Analyze data characteristics to choose visualization
dataProfile := discoveryEngine.ProfileAttribute(ctx, eventType, attribute)

widgetType := determineOptimalVisualization(dataProfile)
// - Use "line" for time series with many points
// - Use "billboard" for single values or KPIs
// - Use "histogram" for distributions
// - Use "table" for categorical breakdowns
```

## 4. Hardcoded Thresholds

### ❌ Current Anti-Pattern
```go
// pkg/interface/mcp/tools_alerts.go:876-880
multipliers := map[string]float64{
    "low":    3.0,  // Fixed standard deviations
    "medium": 2.5,
    "high":   2.0,
}
```

### ✅ First Principles Approach
```go
// Analyze historical data patterns
patterns := discoveryEngine.AnalyzeMetricBehavior(ctx, query, "30 days")

// Calculate optimal thresholds based on:
// - Seasonality patterns
// - Normal variation ranges  
// - Business hours vs off-hours
// - Anomaly detection algorithms

threshold := patterns.CalculateOptimalThreshold(sensitivity, businessContext)
```

## 5. Assumed NRQL Functions

### ❌ Current Anti-Pattern
```go
// pkg/interface/mcp/tools_query.go:376-394
validFunctions := []string{
    "average(", "avg(",
    "count(",
    "latest(",
    // Hardcoded list
}
```

### ✅ First Principles Approach
```go
// Query NRDB for available functions
availableFunctions := nrdbClient.GetAvailableFunctions(ctx)

// Validate against actual capabilities
if !availableFunctions.Supports(requestedFunction) {
    return fmt.Errorf("function %s not available in this NRDB version", requestedFunction)
}
```

## 6. Fixed Data Structure Assumptions

### ❌ Current Anti-Pattern
```go
// pkg/discovery/engine_helpers.go:432
profile.HasTimeSeries = true // Assumes all data is time series
```

### ✅ First Principles Approach
```go
// Detect data structure from actual samples
sample := nrdbClient.GetSample(ctx, eventType, 100)
dataStructure := analyzeDataStructure(sample)

profile.HasTimeSeries = dataStructure.HasTimeAttribute
profile.IsEventData = dataStructure.HasDiscreteEvents
profile.IsMetricData = dataStructure.HasNumericMeasurements
```

## 7. Static Alert Comparisons

### ❌ Current Anti-Pattern
```go
// pkg/interface/mcp/tools_alerts.go:38
Default: "above", // Always defaults to "above"
```

### ✅ First Principles Approach
```go
// Analyze metric behavior
behavior := discoveryEngine.AnalyzeMetricDirection(ctx, metric, "7 days")

// Choose comparison based on metric type:
// - Error rates: usually "above"
// - Success rates: usually "below"  
// - Response times: "above" for degradation
// - Queue depth: depends on context

defaultComparison := behavior.RecommendedComparison()
```

## 8. Domain-Specific Patterns

### ❌ Current Anti-Pattern
```go
// pkg/discovery/engine_helpers.go:91-106
case "performance":
    filter.IncludePatterns = append(filter.IncludePatterns, 
        "*Transaction*", "*PageView*", "*Synthetics*")
```

### ✅ First Principles Approach
```go
// Discover domain patterns from actual data
domainPatterns := map[string][]string{}

// Analyze all event types and their attributes
for _, eventType := range allEventTypes {
    attributes := discoveryEngine.GetAttributes(ctx, eventType)
    domain := classifyDomain(eventType, attributes)
    domainPatterns[domain] = append(domainPatterns[domain], eventType)
}

// Use discovered patterns instead of hardcoded ones
```

## 9. Fixed Time Ranges

### ❌ Current Anti-Pattern
```go
// pkg/interface/mcp/tools_alerts.go:857
"SINCE 7 days ago" // Always uses 7 days for baseline
```

### ✅ First Principles Approach
```go
// Determine optimal time range based on data characteristics
dataProfile := discoveryEngine.GetDataProfile(ctx, metric)

// Consider:
// - Data retention period
// - Seasonality (daily, weekly, monthly patterns)
// - Business cycles
// - Data volume and variability

optimalTimeRange := dataProfile.RecommendedBaselinePeriod()
```

## 10. Template-Based Dashboards

### ❌ Current Anti-Pattern
```go
// Golden Signals dashboard assumes specific metrics exist
case "golden-signals":
    // Hardcoded Transaction, SystemSample, etc.
```

### ✅ First Principles Approach
```go
// Discover Golden Signals dynamically
goldenSignals := map[string]string{}

// Latency: Find metrics with time/duration units
goldenSignals["latency"] = discoveryEngine.FindBestMetric(ctx, 
    MetricCriteria{Unit: "time", Keywords: []string{"duration", "latency", "response"}})

// Traffic: Find rate/count metrics
goldenSignals["traffic"] = discoveryEngine.FindBestMetric(ctx,
    MetricCriteria{Type: "counter", Keywords: []string{"count", "rate", "requests"}})

// Errors: Find error/failure metrics
goldenSignals["errors"] = discoveryEngine.FindBestMetric(ctx,
    MetricCriteria{Keywords: []string{"error", "failure", "exception"}})

// Saturation: Find utilization metrics
goldenSignals["saturation"] = discoveryEngine.FindBestMetric(ctx,
    MetricCriteria{Unit: "percentage", Keywords: []string{"cpu", "memory", "disk"}})

// Build dashboard with discovered metrics
```

## Implementation Strategy

1. **Create Discovery Utilities**
   ```go
   type DiscoveryUtils interface {
       FindEventWithAttribute(ctx, attrPattern string) (*EventType, error)
       FindBestMetric(ctx, criteria MetricCriteria) (string, error)
       AnalyzeDataStructure(ctx, eventType string) (*DataStructure, error)
       GetOptimalTimeRange(ctx, metric string) (string, error)
       RecommendVisualization(ctx, data DataProfile) (string, error)
   }
   ```

2. **Add Schema Introspection**
   - Query GraphQL schema for available operations
   - Discover NRQL functions and capabilities
   - Get data type information from NRDB

3. **Implement Adaptive Logic**
   - Replace all hardcoded values with discovery
   - Cache discovery results for performance
   - Provide fallbacks when discovery fails

4. **Create Validation Layer**
   - Validate queries against discovered schema
   - Check attribute existence before use
   - Verify function availability

## Benefits

1. **Universal Compatibility**: Works with any New Relic account setup
2. **Future Proof**: Adapts to new event types and attributes
3. **Accurate**: Only uses data that actually exists
4. **Intelligent**: Makes data-driven decisions
5. **Transparent**: Can explain why it made certain choices

## Conclusion

By following first principles and discovering from the source of truth (NRDB schema, API contracts, actual data), we create a system that truly adapts to each customer's unique setup rather than imposing our assumptions.
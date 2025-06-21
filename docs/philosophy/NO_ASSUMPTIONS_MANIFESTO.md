# The No Assumptions Manifesto: A Radical Commitment to Discovery-First

This document catalogs every way we've eliminated assumptions and hard-coding from the New Relic MCP Server. We go to extraordinary lengths to ensure NOTHING is assumed about data, schemas, or systems.

## Table of Contents

1. [Core Philosophy](#core-philosophy)
2. [Schema Assumptions We Don't Make](#schema-assumptions-we-dont-make)
3. [Query Assumptions We Don't Make](#query-assumptions-we-dont-make)
4. [System Assumptions We Don't Make](#system-assumptions-we-dont-make)
5. [Behavioral Assumptions We Don't Make](#behavioral-assumptions-we-dont-make)
6. [Implementation Patterns](#implementation-patterns)
7. [The Cost of Not Assuming](#the-cost-of-not-assuming)
8. [Why This Matters](#why-this-matters)

## Core Philosophy

```yaml
traditional_approach:
  assumption: "I know what exists"
  result: "Brittle failures when wrong"

our_approach:
  assumption: "I know nothing"
  result: "Adaptive success always"
```

We treat every piece of data as unknown until proven otherwise. This isn't just good practice—it's a fundamental architectural principle that permeates every line of code.

## Schema Assumptions We Don't Make

### 1. Event Type Existence

**What Others Assume:**
```sql
SELECT * FROM Transaction  -- Assumes Transaction events exist
```

**What We Do:**
```yaml
1. SHOW EVENT TYPES
2. Check if 'Transaction' is in the list
3. Verify it has recent data
4. Only then query it
```

### 2. Attribute Presence

**What Others Assume:**
```sql
WHERE error = true  -- Assumes 'error' attribute exists and is boolean
```

**What We Do:**
```yaml
1. SELECT keyset() FROM Transaction LIMIT 1000
2. Check if 'error' exists in keys
3. Determine its data type from samples
4. Adapt query based on findings:
   - If boolean: WHERE error = true
   - If string: WHERE error = 'true'
   - If missing: Look for error.class, errorCode, statusCode >= 400
```

### 3. Attribute Types

**Never Assume:**
- `duration` is numeric (could be string "123ms")
- `timestamp` is epoch (could be ISO string)
- `error` is boolean (could be error count)
- `host` is a string (could be nested object)

**Always Discover:**
```go
func discoverAttributeType(attr string) AttributeType {
    // Sample actual data
    sample := "SELECT " + attr + " FROM " + eventType + " LIMIT 100"
    results := query(sample)
    
    // Analyze samples to determine type
    types := analyzeValueTypes(results)
    
    // Handle mixed types gracefully
    if types.HasMultiple() {
        return AttributeType{
            Primary: types.MostCommon(),
            Variants: types.All(),
            MixedTypeStrategy: determineBestStrategy(types),
        }
    }
}
```

### 4. Identifier Fields

**What Others Assume:**
```yaml
service_identifier: "appName"  # Hard-coded assumption
```

**What We Do:**
```yaml
discovery_order:
  1. Check for 'appName'
  2. Check for 'applicationName'  
  3. Check for 'service.name' (OpenTelemetry)
  4. Check for 'app' 
  5. Check for 'serviceName'
  6. Check for entity.name
  7. Check for custom tags
  8. Use entity.guid as last resort
  9. If none exist, discover natural groupings
```

### 5. Error Indicators

**Never Assume Error Schema:**
```go
// We maintain a discovery chain for errors
errorDiscoveryChain := []ErrorDetector{
    // Boolean error field
    {
        Detect: func() bool { return hasAttribute("error") && isBooleanish("error") },
        Query: "WHERE error IS true",
    },
    // Error class field
    {
        Detect: func() bool { return hasAttribute("error.class") },
        Query: "WHERE error.class IS NOT NULL",
    },
    // HTTP status codes
    {
        Detect: func() bool { return hasAttribute("httpResponseCode") },
        Query: "WHERE httpResponseCode >= 400",
    },
    // Log levels
    {
        Detect: func() bool { return hasAttribute("level") && eventType == "Log" },
        Query: "WHERE level IN ('ERROR', 'FATAL', 'CRITICAL')",
    },
    // Exception fields
    {
        Detect: func() bool { return hasAttribute("exception.type") },
        Query: "WHERE exception.type IS NOT NULL",
    },
    // Custom error fields
    {
        Detect: func() bool { return discoverCustomErrorField() != "" },
        Query: func() string { return buildCustomErrorQuery() },
    },
}
```

## Query Assumptions We Don't Make

### 1. Aggregation Functions

**What Others Assume:**
```sql
SELECT average(duration)  -- Assumes duration exists and is numeric
```

**What We Do:**
```go
func buildAggregateQuery(metric string, function string) string {
    // First discover the metric
    discovery := discoverMetric(metric)
    
    if !discovery.Exists {
        // Find alternative metrics
        alternatives := findSimilarMetrics(metric)
        if len(alternatives) > 0 {
            metric = alternatives[0]
            discovery = discoverMetric(metric)
        } else {
            return "" // No metric found
        }
    }
    
    // Adapt function to data type
    switch discovery.DataType {
    case Numeric:
        return fmt.Sprintf("SELECT %s(%s)", function, metric)
    case String:
        if function == "average" && discovery.IsNumericString {
            return fmt.Sprintf("SELECT %s(numeric(%s))", function, metric)
        }
        return "" // Can't average strings
    case Mixed:
        return fmt.Sprintf("SELECT %s(%s) WHERE %s IS NOT NULL", 
                          function, metric, metric)
    }
}
```

### 2. Time Windows

**Never Assume:**
- Events exist for the requested time range
- Data is complete for the period
- Timestamp fields are consistent

**Always Check:**
```go
func validateTimeWindow(eventType string, window string) TimeWindowInfo {
    // Check data availability
    availability := fmt.Sprintf(`
        SELECT 
            min(timestamp) as earliest,
            max(timestamp) as latest,
            count(*) as totalEvents
        FROM %s 
        SINCE %s
    `, eventType, window)
    
    result := query(availability)
    
    return TimeWindowInfo{
        HasData: result.totalEvents > 0,
        ActualStart: result.earliest,
        ActualEnd: result.latest,
        Completeness: calculateCompleteness(result),
        Gaps: detectTimeGaps(eventType, window),
    }
}
```

### 3. Facet Cardinality

**What Others Assume:**
```sql
FACET appName  -- Assumes reasonable cardinality
```

**What We Do:**
```go
func buildFacetClause(attribute string) string {
    // Check cardinality first
    cardinalityCheck := fmt.Sprintf(
        "SELECT uniqueCount(%s) as cardinality FROM %s SINCE 1 hour ago",
        attribute, eventType
    )
    
    cardinality := query(cardinalityCheck).cardinality
    
    switch {
    case cardinality == 0:
        return "" // Attribute doesn't exist
    case cardinality == 1:
        return "" // No point in faceting by constant
    case cardinality > 1000:
        // High cardinality - add LIMIT
        return fmt.Sprintf("FACET %s LIMIT 100", attribute)
    default:
        return fmt.Sprintf("FACET %s", attribute)
    }
}
```

## System Assumptions We Don't Make

### 1. Account Structure

**Never Assume:**
- Single account setup
- Consistent schemas across accounts
- Same time zones across accounts

**Always Discover:**
```go
func discoverAccountContext() AccountContext {
    return AccountContext{
        Accounts: listAccessibleAccounts(),
        SchemaVariations: mapSchemasByAccount(),
        TimeZones: detectAccountTimeZones(),
        DataResidency: mapDataLocations(),
    }
}
```

### 2. Data Sources

**Never Assume:**
- All data comes from New Relic agents
- Consistent instrumentation methods
- Standard data formats

**Always Identify:**
```go
type DataSourceDiscovery struct {
    Sources []DataSource
}

type DataSource struct {
    Type string // AGENT, OTLP, API, CUSTOM
    Percentage float64
    Characteristics SourceProfile
}

func discoverDataSources() DataSourceDiscovery {
    // Check instrumentation.provider
    providers := query(`
        SELECT uniqueCount(instrumentation.provider) as providers,
               capture(instrumentation.provider, r'.*') as names
        FROM Metric, Log, Span
    `)
    
    // Check agent names
    agents := query(`
        SELECT uniqueCount(agentName) as agents,
               capture(agentName, r'.*') as names
        FROM Transaction, SystemSample
    `)
    
    // Analyze collection patterns
    patterns := analyzeCollectionPatterns()
    
    return buildSourceProfile(providers, agents, patterns)
}
```

### 3. Metric vs Event Usage

**Never Assume:**
- Dashboards use NRQL
- Metrics are dimensional
- Events have consistent schemas

**Always Analyze:**
```go
func analyzeDashboardDataSources(dashboardGuid string) DashboardAnalysis {
    widgets := getWidgets(dashboardGuid)
    
    analysis := DashboardAnalysis{
        TotalWidgets: len(widgets),
        ByDataSource: make(map[string]int),
    }
    
    for _, widget := range widgets {
        source := classifyWidgetDataSource(widget)
        analysis.ByDataSource[source]++
        
        // Deep analysis of each source
        switch source {
        case "dimensional_metrics":
            analysis.MetricDetails = analyzeMetricWidget(widget)
        case "event_nrql":
            analysis.EventDetails = analyzeEventWidget(widget)
        case "mixed":
            analysis.MixedDetails = analyzeMixedWidget(widget)
        }
    }
    
    return analysis
}
```

## Behavioral Assumptions We Don't Make

### 1. Performance Patterns

**Never Assume:**
- Normal performance baselines
- Consistent traffic patterns
- Regular seasonality

**Always Discover:**
```go
func discoverPerformancePatterns(metric string) PerformanceProfile {
    // Multi-scale analysis
    patterns := PerformanceProfile{}
    
    // Hourly patterns
    patterns.Hourly = analyzeHourlyPatterns(metric)
    
    // Daily patterns  
    patterns.Daily = analyzeDailyPatterns(metric)
    
    // Weekly patterns
    patterns.Weekly = analyzeWeeklyPatterns(metric)
    
    // Detect anomalies at each scale
    patterns.Anomalies = detectMultiScaleAnomalies(patterns)
    
    // Identify if patterns exist at all
    patterns.HasRegularPattern = detectRegularity(patterns)
    
    return patterns
}
```

### 2. Relationship Patterns

**Never Assume:**
- Service dependencies
- Data flow directions
- Correlation meanings

**Always Mine:**
```go
func discoverRelationships(startEntity string) RelationshipGraph {
    graph := NewRelationshipGraph()
    
    // Discover through multiple methods
    methods := []RelationshipDiscoverer{
        // Explicit spans/traces
        DiscoverViaTracing{},
        
        // Implicit via timing
        DiscoverViaTemporalCorrelation{},
        
        // Via shared attributes
        DiscoverViaCommonAttributes{},
        
        // Via error propagation
        DiscoverViaErrorPatterns{},
        
        // Via metric correlation
        DiscoverViaMetricCorrelation{},
    }
    
    for _, method := range methods {
        relationships := method.Discover(startEntity)
        graph.AddRelationships(relationships)
    }
    
    return graph
}
```

### 3. Cost Patterns

**Never Assume:**
- Ingest is evenly distributed
- All data is valuable
- Collection can't be optimized

**Always Analyze:**
```go
func discoverCostOptimizations() CostAnalysis {
    analysis := CostAnalysis{}
    
    // Discover what's collected but never queried
    analysis.UnusedData = findUnqueriedData()
    
    // Discover redundant collection
    analysis.Redundant = findDuplicateMetrics()
    
    // Discover over-sampling
    analysis.OverSampled = findExcessiveGranularity()
    
    // Discover optimization opportunities
    analysis.Opportunities = []Optimization{}
    
    // Widget migration opportunities
    widgetAnalysis := analyzeAllDashboards()
    for _, dashboard := range widgetAnalysis {
        if dashboard.EventWidgetCount > 0 && dashboard.MigratableToMetrics > 0 {
            analysis.Opportunities = append(analysis.Opportunities, Optimization{
                Type: "widget_migration",
                Dashboard: dashboard.Name,
                CurrentCost: estimateEventCost(dashboard),
                OptimizedCost: estimateMetricCost(dashboard),
                Savings: calculateSavings(dashboard),
            })
        }
    }
    
    return analysis
}
```

## Implementation Patterns

### 1. The Discovery Chain Pattern

```go
type DiscoveryChain struct {
    steps []DiscoveryStep
}

type DiscoveryStep struct {
    Name string
    Discover func() (interface{}, error)
    Fallback func() (interface{}, error)
}

// Example: Discovering service identifier
serviceIDChain := DiscoveryChain{
    steps: []DiscoveryStep{
        {Name: "appName", Discover: checkAppName},
        {Name: "service.name", Discover: checkServiceName},
        {Name: "entity.name", Discover: checkEntityName},
        {Name: "custom.service", Discover: checkCustomTags},
        {Name: "natural_grouping", Discover: discoverNaturalGroups},
    },
}
```

### 2. The Adaptive Query Pattern

```go
type AdaptiveQuery struct {
    BaseIntent   string
    Discoveries  []Discovery
    FinalQuery   string
}

func (aq *AdaptiveQuery) Build() string {
    // Start with intent
    intent := parseIntent(aq.BaseIntent)
    
    // Discover available data
    for _, discovery := range aq.Discoveries {
        discovery.Execute()
    }
    
    // Build query that works with what exists
    builder := NewQueryBuilder()
    builder.AdaptToDiscoveries(aq.Discoveries)
    
    return builder.Build(intent)
}
```

### 3. The Progressive Understanding Pattern

```go
type Investigation struct {
    Context   InvestigationContext
    Findings  []Finding
    Certainty float64
}

func (i *Investigation) Investigate(symptom string) {
    // Start with zero assumptions
    i.Certainty = 0.0
    
    // Progressive discovery
    stages := []func(){
        i.discoverWhatExists,
        i.understandStructure,
        i.findPatterns,
        i.detectAnomalies,
        i.traceRelationships,
        i.identifyCauses,
    }
    
    for _, stage := range stages {
        stage()
        i.updateCertainty()
        
        // Stop if we have high certainty
        if i.Certainty > 0.9 {
            break
        }
    }
}
```

## The Cost of Not Assuming

Yes, discovering everything has costs:

### 1. Performance Overhead
- Additional discovery queries before each operation
- Caching complexity to mitigate performance impact
- More API calls overall

### 2. Code Complexity
- Longer functions with discovery phases
- More error handling paths
- Complex fallback chains

### 3. Development Time
- Can't write simple queries quickly
- Must implement discovery for each operation
- Testing is more complex

## Why This Matters

Despite the costs, this approach is essential because:

### 1. **Real-World Resilience**
- Works with any team's instrumentation
- Handles schema drift over time
- Survives partial deployments

### 2. **True Intelligence**
- Discovers insights you didn't know to look for
- Adapts to each unique environment
- Learns from the data itself

### 3. **User Trust**
- No mysterious failures
- Clear explanations when things don't work
- Reliable results every time

### 4. **Future Proof**
- New data sources just work
- Schema changes don't break tools
- Evolution without code changes

## Examples of Extreme Discovery

### Example 1: Finding Error Rate Without Any Assumptions

```go
func calculateErrorRate(service string, timeRange string) (float64, error) {
    // Step 1: Discover what identifies the service
    serviceField := discoverServiceIdentifier(service)
    if serviceField == "" {
        return 0, fmt.Errorf("cannot identify service %s in data", service)
    }
    
    // Step 2: Discover what event types exist for this service
    eventTypes := discoverServiceEventTypes(serviceField, service)
    if len(eventTypes) == 0 {
        return 0, fmt.Errorf("no data found for service %s", service)
    }
    
    // Step 3: For each event type, discover error indicators
    var totalEvents int64
    var errorEvents int64
    
    for _, eventType := range eventTypes {
        // Discover how errors are indicated in this event type
        errorIndicator := discoverErrorIndicator(eventType)
        
        if errorIndicator != nil {
            // Count total events
            totalQuery := fmt.Sprintf(
                "SELECT count(*) FROM %s WHERE %s = '%s' SINCE %s",
                eventType, serviceField, service, timeRange
            )
            total := executeQuery(totalQuery)
            
            // Count error events using discovered indicator
            errorQuery := fmt.Sprintf(
                "SELECT count(*) FROM %s WHERE %s = '%s' AND %s SINCE %s",
                eventType, serviceField, service, errorIndicator.Condition, timeRange
            )
            errors := executeQuery(errorQuery)
            
            totalEvents += total
            errorEvents += errors
        }
    }
    
    if totalEvents == 0 {
        return 0, fmt.Errorf("no events found for service %s in %s", service, timeRange)
    }
    
    return float64(errorEvents) / float64(totalEvents) * 100, nil
}
```

### Example 2: Building Dashboard Without Schema Knowledge

```go
func generateDashboard(service string) Dashboard {
    dashboard := Dashboard{
        Name: fmt.Sprintf("%s Overview", service),
        Pages: []Page{},
    }
    
    // Discover all data about this service
    serviceData := discoverServiceData(service)
    
    // Create widgets based on what we found
    for _, dataType := range serviceData.Types {
        switch dataType.Category {
        case "performance":
            if dataType.HasNumericMetrics() {
                dashboard.AddWidget(createPerformanceWidget(dataType))
            }
        case "errors":
            if indicator := dataType.DiscoverErrorIndicator(); indicator != nil {
                dashboard.AddWidget(createErrorWidget(dataType, indicator))
            }
        case "throughput":
            dashboard.AddWidget(createThroughputWidget(dataType))
        case "infrastructure":
            if dataType.HasResourceMetrics() {
                dashboard.AddWidget(createResourceWidget(dataType))
            }
        }
    }
    
    // Only create dashboard if we found useful data
    if len(dashboard.AllWidgets()) == 0 {
        // Discover what other services have and suggest
        suggestions := discoverSimilarServicePatterns(service)
        dashboard.AddSuggestionsWidget(suggestions)
    }
    
    return dashboard
}
```

### Example 3: Cost Optimization Without Assumptions

```go
func optimizeDataCollection(accountId int) OptimizationPlan {
    plan := OptimizationPlan{}
    
    // Discover all data sources
    sources := discoverAllDataSources(accountId)
    
    // For each source, discover usage patterns
    for _, source := range sources {
        usage := discoverUsagePattern(source)
        
        // Find data that's collected but never queried
        if usage.QueryCount == 0 && usage.DashboardUsage == 0 {
            plan.AddRecommendation(Recommendation{
                Type: "remove_unused",
                Source: source,
                Impact: source.IngestBytes,
                Confidence: 1.0,
            })
        }
        
        // Find over-sampled data
        if usage.EffectiveQueryGranularity > source.CollectionGranularity*10 {
            plan.AddRecommendation(Recommendation{
                Type: "reduce_granularity",
                Source: source,
                NewGranularity: usage.EffectiveQueryGranularity,
                Impact: calculateGranularitySavings(source, usage),
                Confidence: 0.8,
            })
        }
        
        // Find NRQL queries that could be metrics
        if source.Type == "event" && usage.QueryPatterns.IsAggregationOnly() {
            plan.AddRecommendation(Recommendation{
                Type: "convert_to_metric",
                Source: source,
                Impact: source.IngestBytes * 0.9, // Metrics are ~10% the size
                Confidence: 0.9,
            })
        }
    }
    
    return plan
}
```

## Conclusion

Our commitment to making zero assumptions is not just an engineering principle—it's a philosophy that recognizes the messy reality of production systems. By assuming nothing and discovering everything, we create tools that:

1. **Work in any environment** - No matter how it's instrumented
2. **Adapt to change** - Without code modifications
3. **Reveal hidden insights** - By exploring without prejudice
4. **Build user trust** - Through reliable, explainable behavior

This is the difference between tools that work in demos and tools that work in reality. We choose reality, every time.

> "The only assumption we make is that we should make no assumptions."

---

*This document is a living testament to our commitment to discovery-first principles. Every assumption we've avoided is a potential failure we've prevented.*

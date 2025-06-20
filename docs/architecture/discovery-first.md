# Discovery-First Architecture

## Table of Contents

1. [Philosophy](#philosophy)
2. [Core Principles](#core-principles)
3. [Architecture Patterns](#architecture-patterns)
4. [Implementation Strategy](#implementation-strategy)
5. [Tool Design](#tool-design)
6. [Workflow Patterns](#workflow-patterns)
7. [Code Examples](#code-examples)
8. [Migration Guide](#migration-guide)

## Philosophy

The discovery-first approach represents a fundamental shift in how we think about observability tools. Instead of building on assumptions about data structures, we start from a position of knowing nothing and discover everything.

### The Fundamental Question

> "What if we knew nothing about the system we're observing?"

This is not just a thought experiment—it's the foundation of our entire architecture. Traditional observability tools are built on layers of assumptions:

- Services have names stored in `appName`
- Errors are boolean flags
- Duration is measured in milliseconds
- HTTP status codes indicate failures

But what if none of this were true?

### Philosophical Foundations

#### 1. Epistemological Humility

**Traditional Approach**: "I know how systems work"  
**Our Approach**: "I know that I don't know"

```yaml
traditional_epistemology:
  assumption: "Systems follow patterns I understand"
  result: "Tools that work in my world"
  failure_mode: "Break in different worlds"

discovery_epistemology:
  assumption: "Each system is unique"
  result: "Tools that adapt to any world"
  failure_mode: "Only if discovery itself fails"
```

#### 2. Empiricism Over Rationalism

We reject the rationalist approach of deducing system behavior from first principles. Instead, we embrace radical empiricism:

```go
// Rationalist approach (what we reject)
func calculateErrorRate(service string) float64 {
    // Assumes error structure based on "reason"
    return query("SELECT percentage(count(*), WHERE error = true) FROM Transaction")
}

// Empiricist approach (what we embrace)
func calculateErrorRate(service string) float64 {
    // Observes actual data to understand errors
    errorIndicators := discover("What indicates errors in this system?")
    return calculateBasedOnDiscovery(errorIndicators)
}
```

#### 3. Discovery as Respect

Our approach embodies respect for:
- **System Diversity**: Every system is unique
- **Evolution**: What was true yesterday may not be true today
- **Unknown Unknowns**: We discover what we don't know we don't know

## Core Principles

### 1. Never Assume

Start with nothing, discover everything:

```yaml
principle: Start with nothing, discover everything
approach:
  1. What data exists?
  2. What does it contain?
  3. How is it structured?
  4. What patterns emerge?
  5. What conclusions can we draw?
```

### 2. Progressive Understanding

Build knowledge incrementally from evidence:

```
Phase 1: What exists?
└─> Event type discovery
    └─> Basic structure understanding

Phase 2: How is it structured?
└─> Attribute exploration
    └─> Data type analysis
    └─> Coverage assessment

Phase 3: What patterns exist?
└─> Distribution analysis
    └─> Relationship detection
    └─> Anomaly identification

Phase 4: How to query effectively?
└─> Query optimization
    └─> Performance tuning
    └─> Result validation
```

### 3. Adaptive Query Building

Construct queries based on discovered reality:

```go
// Instead of hardcoded queries
query := "SELECT count(*) FROM Transaction WHERE appName = 'checkout'"

// Discovery-based query construction
eventType := discover.findEventTypeContaining("transaction")
serviceIdentifier := discover.findAttributeIdentifying("service")
serviceName := discover.findServiceMatching("checkout")

query := buildQuery(eventType, serviceIdentifier, serviceName)
```

### 4. Continuous Validation

Never trust yesterday's discoveries:

```go
func executeWithDiscovery(ctx context.Context, intent string) {
    // Re-discover on each execution
    currentSchema := discover.getCurrentSchema(ctx)
    
    // Validate assumptions still hold
    if !validate.schemaMatches(currentSchema, cachedSchema) {
        // Adapt to new reality
        strategy := adapt.toNewSchema(currentSchema, intent)
    }
    
    // Execute with current understanding
    return execute(strategy)
}
```

## Architecture Patterns

### Layer 1: Atomic Discovery Tools

The foundation is granular tools that answer specific questions without assumptions:

```yaml
discovery_tools:
  schema_exploration:
    - discovery.explore_event_types   # What types of data exist?
    - discovery.explore_attributes    # What fields are available?
    - discovery.profile_coverage      # How complete is the data?
    
  data_understanding:
    - discovery.analyze_distribution  # How are values distributed?
    - discovery.detect_patterns      # What patterns exist?
    - discovery.find_relationships   # How does data connect?
    
  quality_assessment:
    - discovery.find_gaps           # Where is data missing?
    - discovery.detect_anomalies    # What looks unusual?
    - discovery.validate_assumptions # Test hypotheses against data
```

### Layer 2: Intelligent Query Building

Build queries based on discovered structure:

```yaml
query_tools:
  validation:
    - nrql.check_syntax           # Validate before execution
    - nrql.estimate_cost         # Understand resource impact
    - nrql.test_existence        # Verify data exists
    
  construction:
    - nrql.build_from_discovery  # Generate queries from findings
    - nrql.adapt_to_schema      # Adjust for actual structure
    - nrql.optimize_performance # Tune based on data profile
    
  execution:
    - nrql.execute_with_timeout  # Controlled execution
    - nrql.stream_large_results # Handle big data
    - nrql.cache_expensive      # Smart result caching
```

### Layer 3: Workflow Orchestration

Compose atomic tools into intelligent workflows:

```yaml
workflow_patterns:
  investigation:
    pattern: "Discover → Explore → Analyze → Conclude"
    tools:
      - discovery.explore_event_types
      - discovery.explore_attributes
      - analysis.find_anomalies
      - report.generate_findings
      
  monitoring_setup:
    pattern: "Discover → Profile → Generate → Validate"
    tools:
      - discovery.find_golden_signals
      - discovery.profile_baselines
      - dashboard.generate_from_discovery
      - alert.create_from_patterns
```

## Implementation Strategy

### Discovery Engine Architecture

```go
type DiscoveryEngine struct {
    explorer    SchemaExplorer
    analyzer    PatternAnalyzer
    profiler    DataProfiler
    cache       DiscoveryCache
}

type SchemaExplorer interface {
    ListEventTypes(ctx context.Context, filter Filter) ([]EventType, error)
    ExploreAttributes(ctx context.Context, eventType string) ([]Attribute, error)
    ProfileDataCompleteness(ctx context.Context, eventType string) (Coverage, error)
}

type PatternAnalyzer interface {
    FindPatterns(ctx context.Context, data []DataPoint) ([]Pattern, error)
    DetectAnomalies(ctx context.Context, timeSeries TimeSeries) ([]Anomaly, error)
    IdentifyRelationships(ctx context.Context, datasets []Dataset) ([]Relationship, error)
}
```

### Discovery-First Tool Implementation

```go
// Traditional tool implementation
func (s *Server) handleQueryTransactionErrors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    service := params["service"].(string)
    
    // Hardcoded assumption about error structure
    query := fmt.Sprintf(`
        SELECT count(*) as errorCount 
        FROM TransactionError 
        WHERE appName = '%s' 
        SINCE 1 hour ago
    `, service)
    
    return s.nrClient.Query(ctx, query)
}

// Discovery-first implementation
func (s *Server) handleQueryServiceErrors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    servicePattern := params["service_pattern"].(string)
    
    // Discover what represents errors in this environment
    errorIndicators, err := s.discovery.FindErrorIndicators(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to discover error indicators: %w", err)
    }
    
    // Discover how services are identified
    serviceIdentifier, err := s.discovery.FindServiceIdentifier(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to discover service identifier: %w", err)
    }
    
    // Build query based on discoveries
    query := s.queryBuilder.BuildErrorQuery(
        errorIndicators,
        serviceIdentifier,
        servicePattern,
    )
    
    return s.nrClient.Query(ctx, query)
}
```

## Tool Design

### Discovery Tool Categories

#### 1. Schema Discovery Tools

```yaml
discovery.list_schemas:
  purpose: "Discover available data schemas"
  assumptions: "none"
  output: "List of schemas with metadata"
  
discovery.explore_schema:
  purpose: "Deep dive into a specific schema"
  assumptions: "Schema name exists"
  output: "Complete schema structure"
  
discovery.profile_attribute:
  purpose: "Understand attribute characteristics"
  assumptions: "Attribute exists in schema"
  output: "Data type, cardinality, patterns"
```

#### 2. Pattern Discovery Tools

```yaml
discovery.find_time_patterns:
  purpose: "Identify temporal patterns"
  assumptions: "Data has timestamps"
  output: "Seasonality, trends, cycles"
  
discovery.find_relationships:
  purpose: "Discover data relationships"
  assumptions: "Multiple data sources exist"
  output: "Correlations, dependencies"
  
discovery.detect_anomalies:
  purpose: "Find unusual patterns"
  assumptions: "Normal patterns exist"
  output: "Anomalies with confidence scores"
```

#### 3. Quality Discovery Tools

```yaml
discovery.assess_data_quality:
  purpose: "Evaluate data reliability"
  assumptions: "Quality can be measured"
  output: "Completeness, accuracy scores"
  
discovery.find_data_gaps:
  purpose: "Identify missing data"
  assumptions: "Expected data patterns"
  output: "Gap locations and severity"
```

### Tool Metadata for Discovery

Each tool includes rich metadata to guide usage:

```go
type ToolMetadata struct {
    Category           string
    DiscoveryLevel     string  // "none", "minimal", "full"
    AssumptionsRequired []string
    AdaptsToSchema     bool
    CacheDuration      time.Duration
}

// Example metadata
var discoverEventTypesMetadata = ToolMetadata{
    Category:           "discovery",
    DiscoveryLevel:     "none",
    AssumptionsRequired: []string{},
    AdaptsToSchema:     false,
    CacheDuration:      1 * time.Hour,
}
```

## Workflow Patterns

### Investigation Workflow

```yaml
discovery_driven_investigation:
  initialize:
    - tool: discovery.explore_event_types
      purpose: "What data is available?"
    
  explore:
    - tool: discovery.explore_attributes
      purpose: "What can we analyze?"
    - tool: discovery.profile_coverage
      purpose: "How complete is the data?"
    
  analyze:
    - tool: analysis.find_baselines
      purpose: "What's normal?"
    - tool: analysis.detect_anomalies
      purpose: "What's unusual?"
    
  correlate:
    - tool: discovery.find_relationships
      purpose: "What's connected?"
    - tool: analysis.impact_analysis
      purpose: "What's affected?"
    
  conclude:
    - tool: report.generate_findings
      purpose: "What did we learn?"
```

### Monitoring Setup Workflow

```yaml
discovery_driven_monitoring:
  discover_signals:
    - tool: discovery.find_golden_signals
      purpose: "What indicates health?"
    
  profile_normal:
    - tool: discovery.profile_baselines
      purpose: "What's normal behavior?"
    
  generate_assets:
    - tool: dashboard.create_from_discovery
      purpose: "Visualize discoveries"
    - tool: alert.create_from_patterns
      purpose: "Alert on anomalies"
    
  validate:
    - tool: validation.test_coverage
      purpose: "What might we miss?"
```

## Code Examples

### Complete Discovery-First Query Function

```go
func (s *Server) executeDiscoveryFirstQuery(ctx context.Context, intent QueryIntent) (*QueryResult, error) {
    // Step 1: Discover relevant event types
    eventTypes, err := s.discovery.FindEventTypesForIntent(ctx, intent)
    if err != nil {
        return nil, fmt.Errorf("discovery failed: %w", err)
    }
    
    // Step 2: For each event type, discover schema
    var queries []string
    for _, eventType := range eventTypes {
        schema, err := s.discovery.ExploreSchema(ctx, eventType)
        if err != nil {
            log.Printf("Failed to explore %s: %v", eventType, err)
            continue
        }
        
        // Step 3: Build query based on discovered schema
        query, err := s.buildQueryFromSchema(schema, intent)
        if err != nil {
            log.Printf("Failed to build query for %s: %v", eventType, err)
            continue
        }
        
        queries = append(queries, query)
    }
    
    // Step 4: Execute queries with appropriate strategy
    results, err := s.executeQueries(ctx, queries)
    if err != nil {
        return nil, fmt.Errorf("execution failed: %w", err)
    }
    
    // Step 5: Validate results match intent
    validated, err := s.validateResults(results, intent)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    return validated, nil
}

func (s *Server) buildQueryFromSchema(schema *Schema, intent QueryIntent) (string, error) {
    qb := &QueryBuilder{}
    
    // Select appropriate aggregation based on discovered data types
    for _, attr := range schema.Attributes {
        if intent.RequiresAggregation(attr.Purpose) {
            qb.AddAggregation(attr.Name, attr.DataType)
        }
    }
    
    // Add filters based on discovered attribute presence
    for _, filter := range intent.Filters {
        if attr := schema.FindAttribute(filter.Field); attr != nil {
            qb.AddFilter(attr.Name, filter.Operator, filter.Value)
        }
    }
    
    // Add time range based on data availability
    availability := schema.GetDataAvailability()
    qb.SetTimeRange(intent.TimeRange.ConstrainTo(availability))
    
    return qb.Build(), nil
}
```

### Discovery-First Alert Creation

```go
func (s *Server) createDiscoveryBasedAlert(ctx context.Context, params AlertParams) (*Alert, error) {
    // Discover what data is available for alerting
    discovery := &AlertDiscovery{
        TargetEntity:  params.EntityGUID,
        SignalType:    params.SignalType,
        TimeRange:     "7 days", // Look back to understand patterns
    }
    
    // Step 1: Discover relevant metrics
    metrics, err := s.discovery.FindAlertableMetrics(ctx, discovery)
    if err != nil {
        return nil, fmt.Errorf("metric discovery failed: %w", err)
    }
    
    // Step 2: Profile historical behavior
    profile, err := s.discovery.ProfileMetricBehavior(ctx, metrics)
    if err != nil {
        return nil, fmt.Errorf("profiling failed: %w", err)
    }
    
    // Step 3: Generate alert conditions based on discoveries
    conditions := s.generateConditions(profile)
    
    // Step 4: Create alert with discovered parameters
    alert := &Alert{
        Name:        fmt.Sprintf("Discovery-based alert for %s", params.EntityName),
        Conditions:  conditions,
        Query:       s.buildAlertQuery(metrics, profile),
        Threshold:   s.calculateThreshold(profile),
    }
    
    return s.createAlert(ctx, alert)
}
```

## Migration Guide

### Moving from Assumption-Based to Discovery-First

#### Step 1: Identify Assumptions

Audit your current tools for hardcoded assumptions:

```go
// Before: Assumption-based
func getErrorRate(service string) float64 {
    query := fmt.Sprintf(
        "SELECT percentage(count(*), WHERE error IS true) FROM Transaction WHERE appName = '%s'",
        service,
    )
    // Assumes: error field exists, is boolean, appName identifies service
}

// After: Discovery-first
func getErrorRate(servicePattern string) float64 {
    errorDef := discover.WhatIndicatesErrors()
    serviceId := discover.WhatIdentifiesService() 
    query := buildErrorRateQuery(errorDef, serviceId, servicePattern)
}
```

#### Step 2: Implement Discovery Layer

Add discovery capabilities before queries:

```go
type QueryExecutor struct {
    discovery DiscoveryEngine
    executor  NRQLExecutor
    cache     SchemaCache
}

func (q *QueryExecutor) Execute(ctx context.Context, intent Intent) (*Result, error) {
    // Always discover first
    schema, err := q.discoverOrCache(ctx, intent)
    if err != nil {
        return nil, err
    }
    
    // Build query from discovery
    query := q.buildFromSchema(schema, intent)
    
    // Execute with confidence
    return q.executor.Execute(ctx, query)
}
```

#### Step 3: Progressive Migration

Migrate tools incrementally:

```yaml
migration_phases:
  phase_1:
    description: "Add discovery alongside assumptions"
    approach: "Run both, compare results"
    rollback: "Easy - keep assumption path"
    
  phase_2:
    description: "Default to discovery, fallback to assumptions"
    approach: "Try discovery first"
    rollback: "Fallback still available"
    
  phase_3:
    description: "Pure discovery-first"
    approach: "Remove assumption code"
    rollback: "Requires code restoration"
```

### Best Practices

1. **Cache Discoveries Appropriately**
   - Schema changes slowly - cache for hours
   - Patterns change moderately - cache for minutes
   - Data changes quickly - minimal caching

2. **Handle Discovery Failures Gracefully**
   - Provide meaningful errors
   - Suggest manual exploration
   - Never make blind assumptions

3. **Document Discovered Patterns**
   - Log what you discover
   - Share patterns across tools
   - Build institutional knowledge

4. **Test with Diverse Data**
   - Different account types
   - Various data schemas
   - Multiple time ranges

---

The discovery-first approach transforms how we interact with observability data. By starting from a position of humility and building understanding through exploration, we create tools that work everywhere, adapt to change, and respect the uniqueness of every system.
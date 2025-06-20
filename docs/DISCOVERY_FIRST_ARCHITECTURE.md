# Discovery-First Architecture for New Relic MCP Server

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [First Principles Foundation](#first-principles-foundation)
3. [Core Architecture](#core-architecture)
4. [Tool Taxonomy](#tool-taxonomy)
5. [Workflow Patterns](#workflow-patterns)
6. [Implementation Guide](#implementation-guide)
7. [Migration Strategy](#migration-strategy)

## Executive Summary

This document presents a ground-up redesign of the New Relic MCP Server based on discovery-first principles. Rather than assuming what data exists or how it's structured, this architecture enables AI assistants to explore, understand, and work with the actual data in NRDB.

### Key Principles

1. **Never Assume** - Always discover what exists before querying
2. **Data Drives Decisions** - Let patterns emerge from data, don't impose them
3. **Atomic Tools** - Single-responsibility tools that compose into workflows
4. **Progressive Understanding** - Build knowledge incrementally from evidence

## First Principles Foundation

### The Problem with Assumptions

Traditional observability tools make assumptions:
- "Transaction data exists with duration attribute"
- "Errors are marked with error=true"
- "Services are identified by appName"

These assumptions break when:
- Data collection is incomplete
- Schemas vary across teams
- Custom instrumentation differs
- Historical data has different structures

### The Discovery-First Solution

```yaml
principle: Start with nothing, discover everything
approach:
  1. What data exists?
  2. What does it contain?
  3. How is it structured?
  4. What patterns emerge?
  5. What conclusions can we draw?
```

## Core Architecture

### Layer 1: Atomic Discovery Tools

The foundation is granular tools that answer specific questions without assumptions:

```yaml
discovery_tools:
  schema_exploration:
    - discovery.list_event_types      # What types of data exist?
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
orchestration_patterns:
  sequential:   # Step-by-step investigation
  parallel:     # Concurrent data gathering
  conditional:  # Adaptive based on findings
  iterative:    # Refine understanding progressively
  map_reduce:   # Process large-scale analysis
```

### Layer 4: Context Management

Maintain understanding across tool invocations:

```yaml
context_system:
  discovery_cache:     # Remember what we've learned
  relationship_graph:  # Track data connections
  quality_scores:      # Assess reliability
  finding_chain:       # Build evidence trail
```

## Tool Taxonomy

### 1. Discovery Tools (Never Assume)

```yaml
category: discovery
purpose: Understand what exists without assumptions
tools:
  
  # Schema Discovery
  - tool: discovery.list_event_types
    purpose: Find what data types exist
    inputs:
      time_range: How far back to look
      min_volume: Significance threshold
    outputs:
      event_types: List with counts and freshness
      
  - tool: discovery.explore_attributes
    purpose: Understand event structure
    inputs:
      event_type: Type to explore
      sample_size: How much to analyze
    outputs:
      attributes: Fields with types, coverage, cardinality
      
  # Pattern Discovery
  - tool: discovery.find_natural_groupings
    purpose: Discover how data clusters
    inputs:
      event_type: Data to analyze
      max_groups: Limit for clarity
    outputs:
      groupings: Natural facets with distributions
      
  - tool: discovery.detect_temporal_patterns
    purpose: Find time-based patterns
    inputs:
      query: Base metric to analyze
      pattern_types: What to look for
    outputs:
      patterns: Seasonality, trends, anomalies
      
  # Relationship Discovery
  - tool: discovery.find_join_keys
    purpose: Discover how data connects
    inputs:
      source_type: Primary event type
      target_types: What to check against
    outputs:
      join_paths: Common attributes for joining
      
  # Quality Discovery
  - tool: discovery.assess_completeness
    purpose: Understand data reliability
    inputs:
      event_type: Type to assess
      critical_fields: Must-have attributes
    outputs:
      quality_score: Overall reliability
      issues: Specific problems found
```

### 2. Analysis Tools (Build Understanding)

```yaml
category: analysis
purpose: Draw insights from discovered data
tools:

  # Statistical Analysis
  - tool: analysis.calculate_baseline
    purpose: Establish normal from data
    inputs:
      metric_query: What to baseline
      time_range: Historical period
      method: percentile, average, etc
    outputs:
      baseline: Normal values
      variance: Expected deviation
      
  - tool: analysis.detect_anomalies
    purpose: Find unusual patterns
    inputs:
      data_query: What to analyze
      sensitivity: Detection threshold
      baseline: Normal behavior
    outputs:
      anomalies: Deviations with severity
      
  # Correlation Analysis  
  - tool: analysis.find_correlations
    purpose: Discover relationships
    inputs:
      primary_metric: Main signal
      candidate_metrics: What might relate
      min_correlation: Significance threshold
    outputs:
      correlations: Ranked by strength
      lag_analysis: Time delays
      
  # Root Cause Analysis
  - tool: analysis.trace_causality
    purpose: Find what happened first
    inputs:
      symptoms: What we observe
      time_window: When it occurred
      entities: Where to look
    outputs:
      event_sequence: Chronological order
      propagation_path: How it spread
```

### 3. Action Tools (Make Changes)

```yaml
category: action
purpose: Modify configuration based on evidence
tools:

  # Alert Management
  - tool: alert.create_from_baseline
    purpose: Create alerts from discovered norms
    inputs:
      metric_query: What to monitor
      baseline: Normal behavior
      sensitivity: How tight to set
    outputs:
      alert_config: Generated configuration
      
  - tool: alert.tune_thresholds
    purpose: Adjust based on patterns
    inputs:
      alert_id: Alert to tune
      historical_data: Past performance
      false_positive_tolerance: Acceptable noise
    outputs:
      new_thresholds: Optimized values
      
  # Dashboard Generation
  - tool: dashboard.generate_from_discovery
    purpose: Create dashboards from findings
    inputs:
      entities: What to monitor
      key_metrics: Discovered important signals
      relationships: How things connect
    outputs:
      dashboard_config: Auto-generated JSON
      
  # Configuration Optimization
  - tool: optimize.reduce_collection
    purpose: Minimize costs while maintaining visibility
    inputs:
      usage_analysis: What's actually queried
      redundancy_report: Duplicate data
      slo_requirements: What must be preserved
    outputs:
      drop_rules: Safe to remove
      aggregation_rules: Pre-compute common queries
```

### 4. Platform Governance Tools (Data Observability)

```yaml
category: governance
purpose: Understand platform usage, costs, and adoption patterns
tools:

  # Dashboard Analysis
  - tool: dashboard.list_widgets
    purpose: Inventory all dashboard widgets
    inputs:
      cursor: Pagination cursor
    outputs:
      widgets: List with dashboard info, type, raw config
      
  - tool: dashboard.classify_widgets
    purpose: Categorize widgets by data source
    inputs:
      dashboard_guid: Dashboard to analyze
    outputs:
      metric_widgets: Count using Metrics API
      event_widgets: Count using NRQL
      metric_names: List of dimensional metrics
      event_types: List of event types
      
  - tool: dashboard.find_nrdot_dashboards
    purpose: Find Data Explorer dashboards
    inputs:
      account_id: Optional account filter
    outputs:
      dashboard_guids: List of NRDOT dashboards
      
  # Metric Usage Analysis
  - tool: metric.widget_usage_rank
    purpose: Rank metrics by dashboard usage
    inputs:
      limit: Top N metrics
      time_range: Analysis window
    outputs:
      rankings: Metrics with usage counts and dashboards
      
  # Ingest Analysis
  - tool: usage.ingest_summary
    purpose: Total ingest with source breakdown
    inputs:
      period: Time window (e.g., "30d")
    outputs:
      total_bytes: Overall ingest
      breakdown: Bytes by source (AGENT, OTLP, API)
      
  - tool: usage.otlp_collectors
    purpose: OTEL collector ingest analysis
    inputs:
      period: Time window
    outputs:
      collectors: Name, metric count, bytes estimate
      
  - tool: usage.agent_ingest
    purpose: Native agent ingest stats
    inputs:
      period: Time window
    outputs:
      agents: Agent name and bytes
      comparison: OTEL vs Agent ratio
```

## Workflow Patterns

### Pattern 1: Investigation Workflow

```yaml
workflow: investigate_unknown_issue
pattern: discovery_first

phase_1_explore:
  - What event types exist?
  - Which have recent anomalies?
  - What attributes are available?

phase_2_understand:  
  - How is data distributed?
  - What patterns exist?
  - When did patterns change?

phase_3_correlate:
  - What else changed?
  - How do metrics relate?
  - What happened first?

phase_4_conclude:
  - Build evidence chain
  - Identify root cause
  - Recommend actions

example_flow:
  - discovery.list_event_types(time_range="2 hours")
    # Found: Transaction, SystemSample, Log, Span
    
  - discovery.detect_temporal_patterns(
      query="SELECT count(*) FROM Transaction"
    )
    # Found: Spike at 14:32, 10x normal
    
  - discovery.explore_attributes(event_type="Transaction")
    # Found: error, duration, name, host attributes
    
  - nrql.execute(
      "SELECT count(*) FROM Transaction 
       WHERE timestamp > '14:30' AND timestamp < '14:35'
       FACET error, name"
    )
    # Found: /api/checkout errors dominate
    
  - discovery.find_relationships(
      source="Transaction WHERE name = '/api/checkout'",
      targets=["Span", "Log", "SystemSample"]
    )
    # Found: Matching trace.id in Spans, host in SystemSample
    
  - analysis.trace_causality(
      symptoms=["checkout errors"],
      entities=["checkout-service", "payment-service", "database"]
    )
    # Found: Database CPU spike at 14:31, then errors
```

### Pattern 2: Capacity Planning Workflow

```yaml
workflow: capacity_planning
pattern: data_driven_projection

phase_1_discover_metrics:
  - What infrastructure data exists?
  - How far back does it go?
  - What granularity is available?

phase_2_understand_patterns:
  - What drives resource usage?
  - When do peaks occur?
  - How does load correlate with resources?

phase_3_project_growth:
  - What's the growth trend?
  - What are the seasonal patterns?
  - Where are the bottlenecks?

phase_4_recommend_scaling:
  - What needs scaling?
  - When to scale?
  - How much headroom?
```

### Pattern 3: SLO Definition Workflow

```yaml
workflow: define_slos
pattern: baseline_from_reality

phase_1_discover_experience:
  - What user-facing data exists?
  - What defines success/failure?
  - How complete is the data?

phase_2_analyze_current_state:
  - What's the current performance?
  - How stable is it?
  - What causes degradation?

phase_3_set_realistic_targets:
  - What's achievable?
  - What matters to users?
  - What can we measure reliably?
```

### Pattern 4: Platform Governance Workflow

```yaml
workflow: platform_cost_optimization
pattern: discover_usage_then_optimize

phase_1_inventory_dashboards:
  - What dashboards exist?
  - Which use dimensional metrics vs events?
  - What's the widget distribution?

phase_2_analyze_usage_patterns:
  - Which metrics are most used?
  - What event types dominate dashboards?
  - How does UI usage correlate with ingest?

phase_3_identify_cost_drivers:
  - What's the ingest breakdown by source?
  - Which collectors are noisiest?
  - Where's the OTEL vs Agent split?

phase_4_optimization_opportunities:
  - Which NRQL widgets can migrate to metrics?
  - What data is collected but never queried?
  - Where can we aggregate or sample?

example_flow:
  - dashboard.list_widgets()
    # Found: 150 dashboards, 2500 widgets
    
  - parallel:
    - dashboard.classify_widgets(each_dashboard)
    # Result: 40% metric widgets, 60% event widgets
    
  - metric.widget_usage_rank(limit=50)
    # Top metric: http.server.duration (45 dashboards)
    
  - usage.ingest_summary(period="30d")
    # Total: 10TB, OTLP: 6TB, Agent: 3TB, API: 1TB
    
  - usage.otlp_collectors()
    # Found: payment-collector using 40% of OTLP
    
  - analysis.migration_impact(
      dashboards=high_event_usage,
      target="metrics"
    )
    # Potential savings: 2TB/month
```

## Implementation Guide

### Phase 1: Core Discovery Tools

```go
// 1. Implement atomic discovery tools
package discovery

type ExploreEventTypesInput struct {
    TimeRange    string
    MinVolume    int
    IncludeMeta  bool
}

type EventTypeInfo struct {
    Name         string
    Volume       int64
    FirstSeen    time.Time
    LastSeen     time.Time
    SampleRate   float64
    Attributes   []AttributeInfo
}

func (d *DiscoveryEngine) ExploreEventTypes(ctx context.Context, input ExploreEventTypesInput) ([]EventTypeInfo, error) {
    // 1. Execute SHOW EVENT TYPES
    // 2. Filter by volume threshold
    // 3. Get sample events for metadata
    // 4. Calculate data quality metrics
    // 5. Return comprehensive info
}
```

### Phase 2: Query Builder Integration

```go
// 2. Build queries from discovery
package querybuilder

type DiscoveryAwareBuilder struct {
    discovery  *DiscoveryEngine
    cache      *DiscoveryCache
}

func (b *DiscoveryAwareBuilder) BuildQuery(ctx context.Context, intent QueryIntent) (string, error) {
    // 1. Check cache for schema info
    // 2. Verify attributes exist
    // 3. Adapt query to actual schema
    // 4. Optimize based on cardinality
    // 5. Return validated query
}
```

### Phase 3: Workflow Orchestration

```go
// 3. Implement workflow patterns
package workflow

type DiscoveryFirstWorkflow struct {
    orchestrator *WorkflowOrchestrator
    discovery    *DiscoveryEngine
    analyzer     *AnalysisEngine
}

func (w *DiscoveryFirstWorkflow) Investigate(ctx context.Context, symptoms []Symptom) (*Investigation, error) {
    // 1. Discover available data
    // 2. Understand structure
    // 3. Find patterns
    // 4. Correlate changes
    // 5. Build conclusions
}
```

### Phase 4: Context Management

```go
// 4. Maintain discovery context
package context

type DiscoveryContext struct {
    schemas      map[string]*SchemaInfo
    patterns     map[string]*PatternInfo
    findings     []Finding
    reliability  map[string]float64
}

func (c *DiscoveryContext) Remember(key string, discovery interface{}) {
    // Cache discoveries for reuse
}

func (c *DiscoveryContext) GetReliability(eventType string) float64 {
    // Return data quality score
}
```

## Migration Strategy

### From Assumption-Based to Discovery-First

1. **Audit Current Tools**
   - Identify hard-coded assumptions
   - Document expected schemas
   - Find brittle queries

2. **Implement Discovery Layer**
   - Add discovery tools alongside existing
   - Cache discoveries for performance
   - Validate assumptions with data

3. **Refactor Tool by Tool**
   - Start with investigation workflows
   - Add discovery phase before queries
   - Handle schema variations gracefully

4. **Update Documentation**
   - Emphasize discovery-first approach
   - Provide migration examples
   - Show benefits with metrics

### Example Migration

```yaml
# OLD: Assumption-based
tool: get_error_rate
implementation:
  query: "SELECT percentage(count(*), WHERE error IS true) FROM Transaction"
  # ASSUMES: error attribute exists and is boolean

# NEW: Discovery-first  
tool: get_error_rate
implementation:
  steps:
    1. discover_error_indicator:
       - Check if 'error' attribute exists
       - Check if 'error.class' exists
       - Check if 'httpResponseCode' exists
    2. build_appropriate_query:
       - Use error if boolean with good coverage
       - Use error.class if present
       - Fall back to httpResponseCode >= 400
    3. execute_with_validation:
       - Run query
       - Verify results make sense
       - Flag any quality issues
```

## Benefits

### 1. Reliability
- Works with any schema
- Handles incomplete data
- Adapts to changes

### 2. Intelligence  
- Discovers insights, doesn't assume them
- Finds patterns humans miss
- Builds understanding progressively

### 3. Efficiency
- Only queries what exists
- Optimizes based on actual data
- Caches discoveries

### 4. User Experience
- No brittle failures
- Clear data lineage
- Explainable results

### 5. Platform Governance
- Complete visibility into dashboard composition
- Data-driven cost optimization recommendations
- Clear understanding of ingest sources and volumes
- Metrics adoption tracking across teams
- Identification of redundant data collection

## Conclusion

This discovery-first architecture transforms the New Relic MCP Server from a tool that executes predefined queries to an intelligent system that explores, understands, and adapts to the actual data landscape. By starting with discovery rather than assumptions, we create a more robust, intelligent, and valuable platform for AI-assisted observability.
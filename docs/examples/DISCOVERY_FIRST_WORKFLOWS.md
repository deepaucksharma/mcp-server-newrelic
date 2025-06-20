# Discovery-First Workflows

## First Principles Approach

All workflows must start with discovering what data actually exists in NRDB, not assuming what should be there. We build understanding from the ground up.

## Core Discovery Pattern

Every workflow begins with these fundamental steps:

```yaml
workflow: any_investigation
foundation:
  1_discover_available_data:
    - tool: discovery.list_schemas
      purpose: "What event types exist in this account?"
      outputs:
        - event_types[]
        - sample_counts
        - data_freshness
        
  2_understand_data_structure:
    - tool: discovery.profile_attribute
      purpose: "What attributes exist and what do they contain?"
      for_each: relevant_event_type
      outputs:
        - attribute_names[]
        - data_types
        - cardinality
        - null_percentage
        
  3_find_relationships:
    - tool: discovery.find_relationships
      purpose: "How does this data connect?"
      outputs:
        - join_keys
        - correlation_strength
        - temporal_patterns
```

## Workflow 1: Performance Investigation (Discovery-First)

**Principle**: Don't assume what metrics exist. Discover what's actually being collected.

```yaml
workflow: performance_investigation_discovery_first
description: Investigate performance without assumptions

phase_1_discover_context:
  - name: "What data do we have?"
    tool: nrql.execute
    inputs:
      query: "SHOW EVENT TYPES"
      
  - name: "What entities are reporting?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT uniques(appName), uniques(host), uniques(entity.guid) 
        FROM Transaction, SystemSample, NetworkSample 
        SINCE 1 hour ago
        
  - name: "What time range has data?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT min(timestamp), max(timestamp) 
        FROM Transaction 
        WHERE appName IS NOT NULL 
        SINCE 1 week ago

phase_2_explore_symptoms:
  - name: "What does 'performance' mean in this data?"
    tool: discovery.profile_attribute
    inputs:
      event_type: "Transaction"
      attributes: ["duration", "databaseDuration", "externalDuration", "queueDuration"]
      
  - name: "What performance data is actually available?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT 
          percentage(count(*), WHERE duration IS NOT NULL) as 'Has Duration',
          percentage(count(*), WHERE databaseDuration IS NOT NULL) as 'Has DB Duration',
          percentage(count(*), WHERE error IS NOT NULL) as 'Has Error Flag'
        FROM Transaction 
        SINCE 1 hour ago
        
  - name: "How is performance distributed?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT histogram(duration, width: 100, buckets: 20) 
        FROM Transaction 
        SINCE 1 hour ago

phase_3_identify_patterns_from_data:
  - name: "What patterns exist in the data?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT count(*) 
        FROM Transaction 
        FACET appName, name, host 
        SINCE 1 hour ago 
        LIMIT 100
        
  - name: "When do patterns change?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT rate(count(*), 1 minute) as 'rate',
               average(duration) as 'avg_duration'
        FROM Transaction 
        TIMESERIES 1 minute 
        SINCE 3 hours ago
        
  - name: "What correlates with performance?"
    parallel:
      - tool: nrql.execute
        inputs:
          query: |
            SELECT correlation(duration, databaseDuration) as 'DB Correlation',
                   correlation(duration, externalDuration) as 'External Correlation'
            FROM Transaction 
            SINCE 1 hour ago
            
      - tool: nrql.execute
        inputs:
          query: |
            SELECT average(duration) 
            FROM Transaction 
            FACET error 
            SINCE 1 hour ago

phase_4_discover_anomalies_from_data:
  - name: "What looks unusual compared to history?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT average(duration) as 'Current',
               average(duration) as 'Baseline'
        FROM Transaction 
        SINCE 1 hour ago 
        COMPARE WITH 1 week ago
        
  - name: "Which specific transactions changed?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT count(*) as 'Count',
               average(duration) as 'Avg Duration',
               stddev(duration) / average(duration) as 'Coefficient of Variation'
        FROM Transaction 
        FACET name 
        SINCE 1 hour ago 
        COMPARE WITH 1 week ago 
        LIMIT 50

phase_5_build_understanding:
  - name: "What story does the data tell?"
    tool: context.add_finding
    based_on: "actual NRQL results"
    not: "assumptions"
```

## Workflow 2: Incident Response (Discovery-First)

**Principle**: Don't assume what caused the incident. Let the data reveal it.

```yaml
workflow: incident_response_discovery
description: Respond based on what data shows, not assumptions

phase_1_understand_the_alert:
  - name: "What triggered this alert?"
    tool: nrql.execute
    inputs:
      query: "${alert.nrql_query}"
      note: "Execute the actual alert query to see current state"
      
  - name: "What data contributed to the alert?"
    tool: nrql.execute
    inputs:
      query: |
        ${alert.nrql_query}
        TIMESERIES 1 minute
        SINCE 30 minutes ago
        
  - name: "Is this alert query discovering real issues?"
    tool: nrql.execute
    inputs:
      query: |
        ${alert.nrql_query}
        FACET ${discover_facet_attributes}
        SINCE 1 hour ago

phase_2_discover_scope:
  - name: "What event types have relevant data?"
    tool: discovery.list_schemas
    inputs:
      filter: "recent_activity"
      time_range: "around_incident"
      
  - name: "What entities are involved?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT uniques(entity.guid), uniques(appName), uniques(host)
        FROM ${discovered_event_types}
        WHERE timestamp >= ${incident_start_time}
        SINCE 1 hour ago
        
  - name: "What attributes might be relevant?"
    for_each: discovered_event_type
    tool: discovery.profile_attribute
    inputs:
      event_type: "${event_type}"
      filter: "high_cardinality_change"
      time_comparison: "before_vs_during_incident"

phase_3_discover_changes:
  - name: "What actually changed in the data?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT count(*) 
        FROM ${event_type}
        FACET ${high_value_attributes}
        SINCE 5 minutes ago 
        COMPARE WITH 1 hour ago
        
  - name: "Are there new attribute values?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT uniques(${attribute}) as 'Unique Values'
        FROM ${event_type}
        SINCE 5 minutes ago 
        COMPARE WITH 1 hour ago

phase_4_trace_causality_in_data:
  - name: "What happened first?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT min(timestamp) as 'First Seen'
        FROM ${relevant_event_types}
        WHERE ${anomaly_condition}
        FACET eventType
        SINCE 2 hours ago
        
  - name: "How did it propagate?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT count(*)
        FROM ${event_type}
        WHERE ${anomaly_condition}
        TIMESERIES 1 minute
        FACET ${propagation_attribute}
        SINCE 2 hours ago
```

## Workflow 3: Capacity Planning (Discovery-First)

**Principle**: Don't assume growth patterns. Discover them from historical data.

```yaml
workflow: capacity_planning_discovery
description: Plan capacity based on discovered patterns, not assumptions

phase_1_discover_what_to_measure:
  - name: "What metrics exist for capacity?"
    tool: discovery.list_schemas
    inputs:
      filter: "infrastructure OR system OR container"
      
  - name: "What capacity attributes are collected?"
    tool: discovery.profile_attribute
    inputs:
      event_types: ["SystemSample", "ContainerSample", "K8sNodeSample"]
      pattern: "percent|usage|limit|capacity|available"
      
  - name: "How complete is the data?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT 
          percentage(count(*), WHERE cpuPercent IS NOT NULL) as 'CPU Coverage',
          percentage(count(*), WHERE memoryUsedPercent IS NOT NULL) as 'Memory Coverage',
          percentage(count(*), WHERE diskUsedPercent IS NOT NULL) as 'Disk Coverage'
        FROM SystemSample 
        SINCE 1 week ago

phase_2_discover_historical_patterns:
  - name: "What patterns exist in the data?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT average(cpuPercent), max(cpuPercent), stddev(cpuPercent)
        FROM SystemSample
        TIMESERIES 1 hour
        SINCE 90 days ago
        
  - name: "Are there seasonal patterns?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT average(${metric})
        FROM ${event_type}
        FACET weekdayOf(timestamp), hourOf(timestamp)
        SINCE 30 days ago
        
  - name: "What drives capacity usage?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT correlation(rate(count(*), 1 minute), average(cpuPercent))
        FROM Transaction, SystemSample
        SINCE 7 days ago

phase_3_discover_limits:
  - name: "What are the actual limits?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT max(cpuPercent), 
               max(memoryUsedPercent),
               max(diskUsedPercent)
        FROM SystemSample
        FACET host
        SINCE 30 days ago
        
  - name: "When do we hit limits?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT count(*)
        FROM SystemSample
        WHERE cpuPercent > 80 OR memoryUsedPercent > 80
        TIMESERIES 1 hour
        SINCE 30 days ago
```

## Workflow 4: SLO Definition (Discovery-First)

**Principle**: Don't assume what good looks like. Discover it from actual user experience data.

```yaml
workflow: slo_definition_discovery
description: Define SLOs based on actual data patterns

phase_1_discover_what_matters:
  - name: "What user-facing events exist?"
    tool: discovery.list_schemas
    inputs:
      filter: "transaction OR pageview OR mobile OR browser"
      
  - name: "What defines 'good' in the data?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT count(*) 
        FROM Transaction 
        FACET CASES(
          WHERE duration < 100 as 'Fast',
          WHERE duration < 500 as 'Acceptable',
          WHERE duration < 1000 as 'Slow',
          WHERE duration >= 1000 as 'Very Slow'
        )
        SINCE 1 week ago
        
  - name: "What does error mean here?"
    tool: discovery.profile_attribute
    inputs:
      event_type: "Transaction"
      attributes: ["error", "error.class", "error.message", "httpResponseCode"]

phase_2_discover_baseline_behavior:
  - name: "What's normal performance?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT 
          percentile(duration, 50) as 'p50',
          percentile(duration, 90) as 'p90',
          percentile(duration, 95) as 'p95',
          percentile(duration, 99) as 'p99'
        FROM Transaction
        FACET appName
        SINCE 30 days ago
        
  - name: "How stable is performance?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT 
          stddev(duration) / average(duration) as 'CV',
          max(duration) / percentile(duration, 95) as 'Spike Ratio'
        FROM Transaction
        TIMESERIES 1 day
        SINCE 30 days ago

phase_3_discover_user_impact:
  - name: "What do errors look like?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT count(*)
        FROM Transaction
        WHERE error IS true
        FACET error.class, httpResponseCode
        SINCE 1 week ago
        
  - name: "Which errors impact users?"
    tool: nrql.execute
    inputs:
      query: |
        SELECT percentage(count(*), WHERE error IS true)
        FROM Transaction
        FACET name
        WHERE count(*) > 100
        SINCE 1 week ago
        LIMIT 50
```

## Core Discovery Tools (Granular)

These are the atomic tools that enable discovery-first workflows:

```yaml
tools:
  discovery.explore_schema:
    description: "Explore what data exists without assumptions"
    params:
      time_range: "how far back to look"
      sample_queries: "try these patterns"
    returns:
      event_types: "what's actually there"
      coverage: "how complete the data is"
      
  discovery.profile_unknown_data:
    description: "Understand data without documentation"
    params:
      event_type: "mystery data source"
      sample_size: "how much to analyze"
    returns:
      inferred_purpose: "what this data might represent"
      quality_score: "how reliable it is"
      patterns: "what patterns exist"
      
  discovery.find_natural_facets:
    description: "Discover how data naturally groups"
    params:
      event_type: "data to analyze"
      max_cardinality: "avoid high-cardinality traps"
    returns:
      natural_groupings: "meaningful ways to split data"
      
  discovery.detect_data_issues:
    description: "Find problems in the data itself"
    params:
      event_type: "data to check"
    returns:
      missing_data: "gaps in collection"
      quality_issues: "malformed or suspicious data"
      collection_problems: "instrumentation issues"
```

## Discovery-First Best Practices

### 1. Never Assume
- Don't assume event types exist
- Don't assume attributes are populated
- Don't assume data quality is good
- Don't assume relationships exist

### 2. Always Verify
- Check data existence before querying
- Verify data freshness
- Validate data completeness
- Test assumptions with data

### 3. Let Data Lead
- Follow patterns in the data
- Look for natural groupings
- Find emergent relationships
- Build understanding incrementally

### 4. Question Everything
- Why does this pattern exist?
- Is this data reliable?
- What's missing from the picture?
- Could there be another explanation?

### 5. Build From Evidence
- Every conclusion must trace to data
- Every recommendation must have evidence
- Every pattern must be verifiable
- Every anomaly must be explainable

## Example: Complete Discovery-First Investigation

```yaml
investigation: unknown_problem
approach: "We know something is wrong but not what"

step_1_discover_anomalies:
  - "What event types showed changes?"
    SELECT count(*) 
    FROM Transaction, SystemSample, Log, Metric 
    TIMESERIES 1 hour 
    SINCE 24 hours ago
    
  - "Where are the changes?"
    For each event type with anomalies:
      SELECT count(*) 
      FROM ${eventType} 
      FACET ${all_available_attributes} 
      SINCE 1 hour ago 
      COMPARE WITH 1 day ago

step_2_understand_scope:
  - "What's the blast radius?"
    SELECT uniques(entity.guid), uniques(host), uniques(appName)
    FROM ${affected_event_types}
    WHERE ${anomaly_conditions}
    
  - "When did it start?"
    SELECT min(timestamp)
    FROM ${affected_event_types}
    WHERE ${anomaly_conditions}
    TIMESERIES 1 minute

step_3_find_connections:
  - "What connects the affected components?"
    Use discovered entity.guids to:
    - Find common attributes
    - Discover shared dependencies
    - Identify communication patterns
    
step_4_build_hypothesis:
  Based only on discovered data:
  - What changed first?
  - How did it propagate?
  - What's the root cause?
  - What evidence supports this?
```

This discovery-first approach ensures we work with reality, not assumptions, and build understanding from what actually exists in NRDB.
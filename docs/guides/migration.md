# Migration Guide: From Assumption-Based to Discovery-First

This guide helps you migrate from traditional assumption-based tools to the discovery-first architecture of the New Relic MCP Server.

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [Tool Migration Patterns](#tool-migration-patterns)
3. [Workflow Migration](#workflow-migration)
4. [Common Scenarios](#common-scenarios)
5. [Best Practices](#best-practices)

## Core Concepts

### The Fundamental Shift

**Old Way (Assumption-Based)**
```yaml
approach: "I know what data exists"
process:
  1. Write query assuming schema
  2. Execute and hope it works
  3. Fail if assumptions wrong
```

**New Way (Discovery-First)**
```yaml
approach: "Let me discover what exists"
process:
  1. Explore available data
  2. Understand structure and quality
  3. Build adaptive queries
  4. Validate before execution
```

## Tool Migration Patterns

### Pattern 1: Query Migration

#### Old Tool: Direct Query Execution
```yaml
# BEFORE - Assumes 'error' exists and is boolean
tool: query_nrdb
params:
  query: "SELECT percentage(count(*), WHERE error IS true) FROM Transaction"
  
problems:
  - Fails if 'error' doesn't exist
  - Fails if 'error' isn't boolean
  - No way to handle schema variations
```

#### New Approach: Discovery Before Query
```yaml
# AFTER - Discovers then adapts
workflow: get_error_rate
steps:
  1. discover_error_indicators:
     tool: discovery.explore_attributes
     params:
       event_type: "Transaction"
       attributes: ["error", "error.class", "httpResponseCode"]
       
  2. build_appropriate_query:
     tool: nrql.build_where
     params:
       conditions:
         - if_exists: "error"
           condition: "error IS true"
         - else_if_exists: "error.class"
           condition: "error.class IS NOT NULL"
         - else_if_exists: "httpResponseCode"
           condition: "httpResponseCode >= 400"
           
  3. execute_validated_query:
     tool: nrql.execute
     params:
       query: "${built_query}"
       validate_first: true
```

### Pattern 2: Dashboard Migration

#### Old Tool: Fixed Dashboard Templates
```yaml
# BEFORE - Hard-coded widget queries
tool: create_dashboard
params:
  widgets:
    - title: "Error Rate"
      query: "SELECT percentage(count(*), WHERE error IS true) FROM Transaction"
    - title: "Response Time"
      query: "SELECT average(duration) FROM Transaction FACET appName"
      
problems:
  - Widgets fail if attributes missing
  - Can't adapt to different schemas
  - One-size-fits-all approach
```

#### New Approach: Discovery-Driven Dashboards
```yaml
# AFTER - Dashboards adapt to available data
workflow: create_adaptive_dashboard
steps:
  1. discover_available_metrics:
     tool: discovery.explore_event_types
     params:
       time_range: "7 days"
       min_volume: 1000
       
  2. profile_key_attributes:
     parallel:
       - tool: discovery.find_natural_groupings
         params:
           event_type: "Transaction"
       - tool: discovery.profile_attribute_values
         params:
           event_type: "Transaction"
           attributes: ["duration", "error", "name"]
           
  3. generate_dashboard:
     tool: dashboard.generate_from_discovery
     params:
       discoveries: "${step_2.results}"
       widget_types:
         - error_tracking # Adapts to error schema
         - performance_metrics # Uses available timing data
         - traffic_patterns # Based on discovered groupings
```

### Pattern 3: Alert Migration

#### Old Tool: Static Alert Conditions
```yaml
# BEFORE - Fixed thresholds and assumptions
tool: create_alert
params:
  condition: "SELECT average(duration) FROM Transaction"
  threshold: 1000  # Arbitrary threshold
  
problems:
  - Threshold may not match reality
  - Doesn't account for patterns
  - Can't adapt to different services
```

#### New Approach: Data-Driven Alerts
```yaml
# AFTER - Alerts based on discovered baselines
workflow: create_intelligent_alert
steps:
  1. analyze_historical_performance:
     tool: analysis.calculate_baseline
     params:
       metric_query: "SELECT average(duration) FROM Transaction"
       time_range: "30 days"
       include_patterns: ["daily", "weekly"]
       
  2. detect_anomaly_patterns:
     tool: discovery.detect_temporal_patterns
     params:
       query: "${metric_query}"
       pattern_types: ["seasonality", "trends"]
       
  3. create_adaptive_alert:
     tool: alert.create_from_baseline
     params:
       baseline: "${step_1.baseline}"
       patterns: "${step_2.patterns}"
       sensitivity: "medium"
       adaptive_thresholds: true
```

## Workflow Migration

### Investigation Workflows

#### Old Workflow: Linear Investigation
```yaml
# BEFORE - Rigid step sequence
workflow: investigate_slowness
steps:
  1. Check transaction duration
  2. Look at database time
  3. Check CPU usage
  4. Review error logs
  
problems:
  - Assumes specific metrics exist
  - Fixed investigation path
  - Misses unexpected causes
```

#### New Workflow: Discovery-Driven Investigation
```yaml
# AFTER - Adaptive investigation
workflow: investigate_slowness
phases:
  1. discover_what_changed:
     - What metrics show anomalies?
     - When did patterns shift?
     - Which entities are affected?
     
  2. explore_relationships:
     - How do affected components connect?
     - What dependencies exist?
     - Where did issues propagate?
     
  3. trace_root_cause:
     - What happened first?
     - How did it cascade?
     - What evidence supports this?
```

### Capacity Planning Workflows

#### Old Workflow: Assumption-Based Planning
```yaml
# BEFORE - Fixed metrics and calculations
workflow: capacity_planning
steps:
  1. Get CPU average over 30 days
  2. Apply 20% growth factor
  3. Recommend scaling at 80% threshold
  
problems:
  - Ignores actual growth patterns
  - Misses resource correlations
  - One-size-fits-all thresholds
```

#### New Workflow: Pattern-Based Planning
```yaml
# AFTER - Data-driven projections
workflow: capacity_planning
phases:
  1. discover_resource_patterns:
     - What resources are actually constrained?
     - How do they correlate with load?
     - What patterns exist historically?
     
  2. analyze_growth_trends:
     - What's the actual growth rate?
     - Are there seasonal variations?
     - What drives resource usage?
     
  3. project_intelligently:
     - Based on discovered patterns
     - Account for correlations
     - Service-specific thresholds
```

## Common Scenarios

### Scenario 1: New Service Onboarding

**Old Approach**
```yaml
# Apply standard dashboard and alerts
steps:
  1. Deploy standard dashboard template
  2. Create alerts with default thresholds
  3. Hope they work for this service
```

**New Approach**
```yaml
# Discover and adapt to service specifics
steps:
  1. discover_service_data:
     - What events does it generate?
     - What attributes are available?
     - What patterns exist?
     
  2. profile_service_behavior:
     - What's normal performance?
     - How does it handle load?
     - What indicates problems?
     
  3. generate_custom_monitoring:
     - Service-specific dashboards
     - Baseline-driven alerts
     - Relevant metrics only
```

### Scenario 2: Debugging Failed Queries

**Old Approach**
```yaml
# Query fails, guess why
error: "attribute 'error' does not exist"
response: Try different attribute names randomly
```

**New Approach**
```yaml
# Systematic discovery
steps:
  1. explore_available_attributes:
     tool: discovery.explore_attributes
     # See what actually exists
     
  2. find_error_indicators:
     tool: discovery.profile_attribute_values
     # Understand what indicates errors
     
  3. build_working_query:
     tool: nrql.build_from_discovery
     # Create query that matches reality
```

### Scenario 3: Cross-Team Data Access

**Old Approach**
```yaml
# Assume same schema everywhere
problem: Different teams instrument differently
result: Queries work for some services, fail for others
```

**New Approach**
```yaml
# Discover each team's schema
steps:
  1. map_team_schemas:
     - Discover what each team collects
     - Find common attributes
     - Identify variations
     
  2. create_adaptive_queries:
     - Queries that work across schemas
     - Handle missing attributes
     - Aggregate despite differences
```

## Best Practices

### 1. Always Start with Discovery

```yaml
# Make discovery your first step
before_any_operation:
  - What data exists?
  - Is it complete?
  - How is it structured?
  - What patterns are present?
```

### 2. Cache Discoveries for Performance

```yaml
# Discovery results are cacheable
caching_strategy:
  - Cache schema information (1 hour)
  - Cache baselines (15 minutes)
  - Cache relationships (30 minutes)
  - Invalidate on schema changes
```

### 3. Build Adaptive, Not Brittle

```yaml
# Queries should handle variations
adaptive_patterns:
  - Use conditional logic
  - Provide fallbacks
  - Handle missing data gracefully
  - Validate before execution
```

### 4. Document Discoveries

```yaml
# Record what you learn
documentation:
  - Schema variations by service
  - Common patterns found
  - Reliability scores
  - Relationship mappings
```

### 5. Progressive Migration

```yaml
# Don't migrate everything at once
migration_phases:
  1. Start with investigation workflows
  2. Move to dashboard generation
  3. Update alert creation
  4. Refactor bulk operations
```

## Migration Checklist

- [ ] Identify assumption-based tools in use
- [ ] Map to discovery-first equivalents
- [ ] Update workflows to include discovery phase
- [ ] Add validation before operations
- [ ] Implement caching for performance
- [ ] Test with varied schemas
- [ ] Document discovered patterns
- [ ] Train team on new approach

## Getting Help

- Review [DISCOVERY_FIRST_ARCHITECTURE.md](./DISCOVERY_FIRST_ARCHITECTURE.md) for principles
- See [API Reference](../api/reference.md) for tool details
- Check [WORKFLOW_PATTERNS_GUIDE.md](./WORKFLOW_PATTERNS_GUIDE.md) for examples
- Use mock mode to test migrations safely

Remember: The goal isn't to predict what data exists, but to discover it and adapt accordingly.

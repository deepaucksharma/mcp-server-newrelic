# Tool Granularity Enhancement Plan

## Executive Summary

This document outlines a comprehensive plan to enhance the New Relic MCP Server with more granular, atomic tools that enable better AI orchestration while maintaining safety, observability, and production-grade operations.

## Core Principles

### 1. Atomic Operations
- **One Tool = One GraphQL/NRQL Operation**: No hidden chaining or complex logic within tools
- **Composability**: Tools should be easily combined by AI agents to create complex workflows
- **Predictability**: Each tool does exactly one thing with clear, documented behavior

### 2. Enhanced Metadata System

```go
type EnhancedTool struct {
    Tool
    // New metadata fields
    Category      string              // query, mutation, analysis, utility
    Safety        SafetyMetadata      // destructive flags, confirmation requirements
    Performance   PerformanceMetadata // expected latency, resource usage
    AIGuidance    AIGuidanceMetadata  // usage examples, common patterns
    Observability ObservabilityMetadata // metrics, tracing configuration
}

type SafetyMetadata struct {
    IsDestructive     bool
    RequiresConfirmation bool
    DryRunSupported   bool
    AffectedResources []string
    RollbackSupported bool
}

type PerformanceMetadata struct {
    ExpectedLatencyMS int
    MaxResultSize     int
    RateLimitPerMin   int
    Cacheable         bool
    CacheTTLSeconds   int
}

type AIGuidanceMetadata struct {
    UsageExamples    []string
    CommonPatterns   []string
    PreferredOver    []string // other tools this is preferred over
    ChainsWith       []string // tools commonly used together
    WarningsForAI    []string
}
```

## Granular Tool Architecture

### Query Tools (Atomic NRQL Operations)

```go
// 1. Basic Query Execution
"nrql.execute": {
    Description: "Execute a single NRQL query with full control",
    Params: {
        query: string (required)
        account_id: int
        timeout: int (default: 30)
        include_metadata: bool
    }
}

// 2. Query Validation
"nrql.validate": {
    Description: "Validate NRQL syntax without execution",
    Params: {
        query: string (required)
        check_permissions: bool
    }
}

// 3. Query Cost Estimation
"nrql.estimate_cost": {
    Description: "Estimate query cost and performance impact",
    Params: {
        query: string (required)
        time_range: string
    }
}

// 4. Query Builder Components (more granular than current)
"nrql.build_select": {
    Description: "Build SELECT clause with proper escaping",
    Params: {
        aggregations: []AggregationSpec
        attributes: []string
    }
}

"nrql.build_where": {
    Description: "Build WHERE clause with proper escaping",
    Params: {
        conditions: []ConditionSpec
        operator: "AND" | "OR"
    }
}

"nrql.build_facet": {
    Description: "Build FACET clause with limit handling",
    Params: {
        attributes: []string
        limit: int
        cases: []CaseSpec
    }
}
```

### Entity Operations (Granular Entity Management)

```go
// 1. Entity Search (more specific than current)
"entity.search_by_name": {
    Description: "Search entities by exact or partial name match",
    Params: {
        name: string (required)
        match_type: "exact" | "contains" | "starts_with"
        domain: string
        type: string
        limit: int
        cursor: string
    }
}

"entity.search_by_tag": {
    Description: "Search entities by tag key/value pairs",
    Params: {
        tags: map[string]string (required)
        match_all: bool
        limit: int
        cursor: string
    }
}

"entity.search_by_alert_status": {
    Description: "Find entities with specific alert conditions",
    Params: {
        alert_severity: "CRITICAL" | "WARNING" | "NOT_ALERTING"
        violation_count_min: int
        time_window: string
    }
}

// 2. Entity Details (more focused)
"entity.get_golden_metrics": {
    Description: "Get only golden signal metrics for an entity",
    Params: {
        guid: string (required)
        time_range: string
    }
}

"entity.get_relationships": {
    Description: "Get entity relationship graph",
    Params: {
        guid: string (required)
        relationship_types: []string
        depth: int (default: 1)
    }
}

"entity.get_tags": {
    Description: "Get all tags for an entity",
    Params: {
        guid: string (required)
        include_system: bool
    }
}
```

### Dashboard Operations (Granular Dashboard Management)

```go
// 1. Dashboard Discovery
"dashboard.search_by_metric": {
    Description: "Find dashboards using specific metrics",
    Params: {
        metric_name: string (required)
        event_type: string
        limit: int
    }
}

"dashboard.search_by_entity": {
    Description: "Find dashboards referencing an entity",
    Params: {
        entity_guid: string (required)
        include_related: bool
    }
}

// 2. Dashboard Components
"dashboard.create_widget": {
    Description: "Create a single dashboard widget",
    Params: {
        dashboard_guid: string (required)
        page_guid: string (required)
        widget: WidgetSpec
        dry_run: bool
    }
}

"dashboard.update_widget_query": {
    Description: "Update only the query of a widget",
    Params: {
        dashboard_guid: string (required)
        widget_id: string (required)
        query: string (required)
        dry_run: bool
    }
}

"dashboard.clone_widget": {
    Description: "Clone a widget within or across dashboards",
    Params: {
        source_widget_id: string (required)
        target_dashboard_guid: string (required)
        target_page_guid: string
        modifications: WidgetModifications
    }
}

// 3. Dashboard Templates (more specific)
"dashboard.apply_golden_signals_template": {
    Description: "Create golden signals dashboard for APM entity",
    Params: {
        entity_guid: string (required)
        time_range: string
        include_dependencies: bool
        dry_run: bool
    }
}

"dashboard.apply_sli_template": {
    Description: "Create SLI dashboard from specification",
    Params: {
        sli_spec: SLISpecification (required)
        alert_on_breach: bool
        dry_run: bool
    }
}
```

### Alert Operations (Granular Alert Management)

```go
// 1. Alert Discovery
"alert.find_by_entity": {
    Description: "Find all alerts for a specific entity",
    Params: {
        entity_guid: string (required)
        include_inherited: bool
        status_filter: []string
    }
}

"alert.find_similar": {
    Description: "Find alerts with similar conditions",
    Params: {
        reference_condition_id: string (required)
        similarity_threshold: float
    }
}

// 2. Alert Condition Components
"alert.create_threshold_condition": {
    Description: "Create a threshold-based alert condition",
    Params: {
        policy_id: string (required)
        name: string (required)
        query: string (required)
        threshold: ThresholdSpec
        duration: DurationSpec
        dry_run: bool
    }
}

"alert.create_anomaly_condition": {
    Description: "Create an anomaly detection condition",
    Params: {
        policy_id: string (required)
        name: string (required)
        query: string (required)
        sensitivity: float
        direction: "UPPER" | "LOWER" | "BOTH"
        dry_run: bool
    }
}

// 3. Alert Actions
"alert.acknowledge_incident": {
    Description: "Acknowledge a specific incident",
    Params: {
        incident_id: string (required)
        acknowledgment_note: string
    }
}

"alert.mute_condition": {
    Description: "Temporarily mute an alert condition",
    Params: {
        condition_id: string (required)
        duration_minutes: int (required)
        reason: string (required)
    }
}
```

### Bulk Operations (Granular Bulk Actions)

```go
// 1. Bulk Tagging
"bulk.add_tags": {
    Description: "Add tags to multiple entities",
    Params: {
        entity_guids: []string (required)
        tags: map[string]string (required)
        skip_on_error: bool
        dry_run: bool
    }
}

"bulk.remove_tags": {
    Description: "Remove tags from multiple entities",
    Params: {
        entity_guids: []string (required)
        tag_keys: []string (required)
        dry_run: bool
    }
}

// 2. Bulk Alert Operations
"bulk.update_alert_thresholds": {
    Description: "Update thresholds for multiple similar alerts",
    Params: {
        condition_ids: []string (required)
        threshold_updates: ThresholdUpdateSpec
        dry_run: bool
    }
}

"bulk.migrate_alert_policies": {
    Description: "Migrate conditions between policies",
    Params: {
        source_policy_id: string (required)
        target_policy_id: string (required)
        condition_filter: ConditionFilter
        dry_run: bool
    }
}

// 3. Bulk Query Execution
"bulk.execute_queries_parallel": {
    Description: "Execute multiple queries in parallel",
    Params: {
        queries: []QuerySpec (required)
        max_concurrent: int (default: 5)
        stop_on_error: bool
    }
}
```

### Analysis Tools (Granular Analysis Operations)

```go
// 1. Pattern Detection
"analysis.detect_anomalies": {
    Description: "Detect anomalies in time series data",
    Params: {
        query: string (required)
        sensitivity: float
        baseline_window: string
        detection_window: string
    }
}

"analysis.find_correlations": {
    Description: "Find correlated metrics",
    Params: {
        primary_query: string (required)
        candidate_queries: []string
        min_correlation: float (default: 0.7)
        time_range: string
    }
}

// 2. Capacity Planning
"analysis.forecast_usage": {
    Description: "Forecast future resource usage",
    Params: {
        query: string (required)
        forecast_window: string
        confidence_level: float
        include_seasonality: bool
    }
}

"analysis.calculate_headroom": {
    Description: "Calculate capacity headroom",
    Params: {
        current_usage_query: string (required)
        capacity_limit: float (required)
        growth_rate: float
    }
}

// 3. Cost Analysis
"analysis.estimate_query_cost": {
    Description: "Estimate data query costs",
    Params: {
        queries: []string (required)
        execution_frequency: string
        retention_period: string
    }
}
```

### Utility Tools (Supporting Operations)

```go
// 1. NRQL Helpers
"utility.escape_nrql_string": {
    Description: "Properly escape strings for NRQL",
    Params: {
        value: string (required)
        context: "WHERE" | "SELECT" | "FACET"
    }
}

"utility.format_nrql_timestamp": {
    Description: "Format timestamp for NRQL",
    Params: {
        timestamp: string (required)
        format: string
    }
}

// 2. Template Generation
"utility.generate_sli_query": {
    Description: "Generate SLI query from specification",
    Params: {
        sli_type: "availability" | "latency" | "quality" | "throughput"
        good_events_filter: string
        total_events_filter: string
    }
}

"utility.generate_golden_signal_queries": {
    Description: "Generate golden signal queries for entity type",
    Params: {
        entity_type: string (required)
        entity_name: string
        custom_attributes: []string
    }
}

// 3. Validation Helpers
"utility.validate_entity_guid": {
    Description: "Validate entity GUID format and existence",
    Params: {
        guid: string (required)
        check_existence: bool
    }
}

"utility.validate_threshold_config": {
    Description: "Validate alert threshold configuration",
    Params: {
        threshold: ThresholdSpec (required)
        metric_type: string
    }
}
```

## Implementation Priorities

### Phase 1: Core Query Tools (Week 1)
- Implement granular NRQL operations
- Add rich metadata to existing tools
- Implement dry-run support for mutations

### Phase 2: Entity & Dashboard Tools (Week 2)
- Implement granular entity search operations
- Add dashboard component operations
- Implement template-based dashboard creation

### Phase 3: Alert & Analysis Tools (Week 3)
- Implement granular alert operations
- Add analysis and pattern detection tools
- Implement bulk operations framework

### Phase 4: Safety & Observability (Week 4)
- Add comprehensive safety metadata
- Implement operation audit logging
- Add performance tracking and metrics

### Phase 5: AI Integration (Week 5)
- Add AI guidance metadata
- Implement tool chaining hints
- Create example workflows

## Benefits of Granular Approach

1. **Better AI Orchestration**: AI can compose complex workflows from simple, predictable tools
2. **Improved Safety**: Each operation is clearly scoped with appropriate safety checks
3. **Enhanced Debugging**: Single-purpose tools are easier to test and debug
4. **Flexible Composition**: New workflows can be created without code changes
5. **Progressive Disclosure**: AI can start with simple operations and progress to complex ones

## Example AI Workflow

```yaml
User: "Find services with high error rates and create alerts for them"

AI Orchestration:
1. entity.search_by_tag:
    tags: {service-tier: "production"}
    domain: "APM"
    
2. For each entity:
   a. nrql.execute:
      query: "SELECT percentage(count(*), WHERE error IS true) FROM Transaction WHERE entityGuid = '{guid}' SINCE 1 hour ago"
      
   b. If error_rate > 5%:
      - alert.find_by_entity:
          entity_guid: {guid}
          
      - If no existing alert:
        - alert.create_threshold_condition:
            name: "High Error Rate - {entity_name}"
            query: "SELECT percentage(count(*), WHERE error IS true) FROM Transaction WHERE entityGuid = '{guid}'"
            threshold: {value: 5, duration: 5}
            dry_run: true
            
      - Show dry_run results to user for confirmation
```

## Next Steps

1. Review and approve the granular tool architecture
2. Implement enhanced metadata system
3. Refactor existing tools into granular components
4. Add comprehensive tests for each atomic operation
5. Update documentation with new tool catalog
6. Create AI orchestration examples and patterns
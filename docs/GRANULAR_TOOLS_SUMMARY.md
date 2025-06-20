# Granular Tools Summary: 120+ Atomic Operations

## Overview

The New Relic MCP Server provides 120+ granular, atomic tools that compose into sophisticated workflows. Each tool follows the single-responsibility principle and can be combined with others to create powerful observability solutions.

## Tool Categories

### 1. Discovery Tools (~30 tools)
**Purpose**: Explore and understand data without assumptions

#### Schema Discovery
- `discovery.list_event_types` - Find what data types exist
- `discovery.explore_attributes` - Understand event structure  
- `discovery.profile_coverage` - Analyze data completeness
- `discovery.get_sample_events` - Retrieve example data

#### Pattern Discovery
- `discovery.find_natural_groupings` - Discover how data clusters
- `discovery.detect_temporal_patterns` - Find time-based patterns
- `discovery.identify_seasonality` - Detect recurring patterns
- `discovery.find_anomaly_windows` - Locate unusual time periods

#### Relationship Discovery
- `discovery.find_relationships` - Discover data connections
- `discovery.find_join_keys` - Identify common attributes
- `discovery.map_entity_relationships` - Build entity graphs
- `discovery.trace_data_lineage` - Track data flow

#### Quality Assessment
- `discovery.assess_completeness` - Data reliability scoring
- `discovery.detect_schema_drift` - Find structure changes
- `discovery.validate_assumptions` - Test hypotheses
- `discovery.find_data_gaps` - Identify missing periods

### 2. Query Tools (~20 tools)
**Purpose**: Build and execute adaptive queries

#### Query Building
- `nrql.build_select` - Construct SELECT clause
- `nrql.build_where` - Build WHERE conditions
- `nrql.build_facet` - Create FACET groupings
- `nrql.build_timeseries` - Generate time-based queries
- `nrql.combine_queries` - Merge multiple queries

#### Query Execution
- `nrql.execute` - Run query with validation
- `nrql.execute_async` - Asynchronous execution
- `nrql.stream_results` - Stream large datasets
- `nrql.execute_batch` - Run multiple queries

#### Query Optimization
- `nrql.validate` - Check syntax and schema
- `nrql.estimate_cost` - Predict resource usage
- `nrql.optimize_performance` - Improve query speed
- `nrql.suggest_indexes` - Recommend optimizations

### 3. Analysis Tools (~25 tools)
**Purpose**: Derive insights from discovered data

#### Statistical Analysis
- `analysis.calculate_baseline` - Establish normal behavior
- `analysis.compute_percentiles` - Distribution analysis
- `analysis.calculate_variance` - Measure stability
- `analysis.trend_analysis` - Identify directions

#### Anomaly Detection
- `analysis.detect_anomalies` - Find deviations
- `analysis.classify_anomalies` - Categorize issues
- `analysis.score_severity` - Rank importance
- `analysis.predict_anomalies` - Forecast issues

#### Correlation Analysis
- `analysis.find_correlations` - Discover relationships
- `analysis.calculate_lag` - Time delay analysis
- `analysis.multivariate_analysis` - Complex relationships
- `analysis.causality_inference` - Determine causes

#### Root Cause Analysis
- `analysis.trace_causality` - Find event sequences
- `analysis.identify_dependencies` - Map connections
- `analysis.impact_analysis` - Assess effects
- `analysis.blame_attribution` - Assign responsibility

### 4. Action Tools (~25 tools)
**Purpose**: Make changes based on evidence

#### Alert Management
- `alert.create_from_baseline` - Data-driven alerts
- `alert.tune_thresholds` - Optimize sensitivity
- `alert.predict_noise` - Forecast false positives
- `alert.suggest_conditions` - Recommend alerts
- `alert.bulk_update` - Mass modifications

#### Dashboard Generation
- `dashboard.generate_from_discovery` - Auto-create views
- `dashboard.optimize_layout` - Improve organization
- `dashboard.create_slo_dashboard` - SLO tracking
- `dashboard.migrate_to_metrics` - Convert widgets
- `dashboard.clone_and_adapt` - Template reuse

#### Configuration Optimization
- `optimize.reduce_collection` - Cut costs
- `optimize.aggregate_metrics` - Pre-computation
- `optimize.deduplicate_data` - Remove redundancy
- `optimize.adjust_sampling` - Balance detail/cost
- `optimize.recommend_rollups` - Suggest aggregations

### 5. Platform Governance Tools (~20 tools)
**Purpose**: Manage platform usage and costs

#### Dashboard Analysis
- `dashboard.list_widgets` - Widget inventory
- `dashboard.classify_widgets` - Metric vs event split
- `dashboard.find_nrdot_dashboards` - Data Explorer usage
- `dashboard.analyze_complexity` - Performance impact
- `dashboard.find_duplicates` - Redundant dashboards

#### Metric Usage Analysis
- `metric.widget_usage_rank` - Popular metrics
- `metric.find_unused` - Orphaned metrics
- `metric.cardinality_analysis` - Explosion risks
- `metric.namespace_summary` - Organization patterns
- `metric.adoption_timeline` - Usage growth

#### Ingest Analysis
- `usage.ingest_summary` - Total volume breakdown
- `usage.otlp_collectors` - OTEL analysis
- `usage.agent_ingest` - Native agent stats
- `usage.api_ingest` - Custom integration volume
- `usage.forecast_costs` - Predict bills

#### Cost Optimization
- `cost.identify_savings` - Find opportunities
- `cost.migration_impact` - Calculate benefits
- `cost.simulate_changes` - What-if analysis
- `cost.track_initiatives` - Measure success

## Composition Patterns

### Sequential Composition
```yaml
# Find and fix performance issues
1. discovery.detect_temporal_patterns
2. analysis.trace_causality
3. alert.create_from_baseline
```

### Parallel Composition
```yaml
# Comprehensive health check
parallel:
  - discovery.list_event_types
  - usage.ingest_summary
  - dashboard.list_widgets
```

### Conditional Composition
```yaml
# Adaptive error analysis
if discovery.explore_attributes finds 'error':
  - nrql.build_where(error IS true)
else if finds 'error.class':
  - nrql.build_where(error.class IS NOT NULL)
else:
  - discovery.find_error_indicators
```

### Loop Composition
```yaml
# Analyze all dashboards
for each dashboard in dashboard.list_widgets:
  - dashboard.classify_widgets
  - metric.widget_usage_rank
```

## Tool Characteristics

### Safety Levels
- **Safe**: Read-only operations (80% of tools)
- **Caution**: Modifies non-critical resources (15%)
- **Destructive**: Deletes or major changes (5%)

### Performance Profiles
- **Instant** (<100ms): Utility tools, builders
- **Fast** (<1s): Most queries, simple analysis
- **Medium** (1-5s): Complex analysis, discovery
- **Slow** (5-30s): Large-scale operations

### Cacheability
- **Always**: Schema info, metadata (cache 1h)
- **Sometimes**: Query results (cache 5-15min)
- **Never**: Real-time data, mutations

## Best Practices

### 1. Start with Discovery
Always begin workflows with discovery tools to understand the data landscape before querying or analyzing.

### 2. Compose Atomically
Each tool does one thing well. Combine them for complex operations rather than creating monolithic tools.

### 3. Cache Discoveries
Discovery results change slowly. Cache them aggressively to improve performance.

### 4. Handle Failures Gracefully
Tools may fail due to missing data or permissions. Always have fallback strategies.

### 5. Use Appropriate Granularity
- Use specific tools for precise operations
- Use workflow tools for common patterns
- Build custom compositions for unique needs

## Integration Examples

### Example 1: Cost Optimization Workflow
```yaml
tools:
  1. dashboard.list_widgets()
  2. dashboard.classify_widgets(each)
  3. metric.widget_usage_rank()
  4. usage.ingest_summary()
  5. cost.identify_savings()
  6. dashboard.migrate_to_metrics(selected)
```

### Example 2: Incident Investigation
```yaml
tools:
  1. discovery.detect_temporal_patterns(time=incident)
  2. discovery.find_relationships(affected_services)
  3. analysis.trace_causality(symptoms)
  4. analysis.impact_analysis(root_cause)
  5. alert.create_from_baseline(preventive)
```

### Example 3: Platform Governance
```yaml
tools:
  1. usage.ingest_summary(period=30d)
  2. usage.otlp_collectors()
  3. dashboard.find_nrdot_dashboards()
  4. metric.adoption_timeline()
  5. cost.track_initiatives()
```

## Future Tools (Planned)

### Machine Learning Tools
- `ml.forecast_metrics` - Time series prediction
- `ml.cluster_behaviors` - Automatic grouping
- `ml.detect_drift` - Model performance

### Advanced Governance
- `governance.enforce_standards` - Policy compliance
- `governance.tag_optimization` - Metadata quality
- `governance.access_patterns` - Usage analytics

### Integration Tools
- `integrate.import_dashboards` - Migration helper
- `integrate.export_configs` - Backup/share
- `integrate.sync_definitions` - Cross-account sync

## Conclusion

The granular tools architecture enables:
- **Flexibility**: Combine tools for any use case
- **Reliability**: Each tool is simple and testable
- **Performance**: Cache and optimize at tool level
- **Discoverability**: Clear purpose for each tool
- **Composability**: Build complex from simple

This approach transforms the MCP server from a rigid tool into a flexible platform for observability innovation.
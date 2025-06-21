# MCP Server Tools Overview

This document provides a comprehensive catalog of tools available in the New Relic MCP Server, organized by category and functionality.

## Tool Categories

The MCP server provides tools across 8 functional categories:

| Category | Purpose |
|----------|---------|
| Discovery | Explore data schemas, attributes, and relationships |
| Query | Execute NRQL queries with various optimizations |
| Alerts | Create, manage, and optimize alert policies |
| Dashboards | Build and manage custom dashboards |
| Analysis | Statistical analysis and pattern detection |
| Governance | Usage analysis, cost optimization, compliance |
| Workflow | Orchestrate complex multi-tool operations |
| Session | Manage stateful sessions |

## Discovery Tools

Discovery tools help explore your New Relic data landscape without prior knowledge of its structure.

### discovery.explore_event_types
**Description**: List available event types in your account

**Parameters**:
- `limit` (integer, optional): Maximum number to return (default: 100)
- `search` (string, optional): Filter event types by name

**Example**:
```json
{
  "name": "discovery.explore_event_types",
  "arguments": {
    "limit": 10
  }
}
```

**Returns**: List of event types with counts and metadata

---

### discovery.explore_attributes
**Description**: Explore all attributes for a specific event type

**Parameters**:
- `event_type` (string, required): The event type to explore
- `time_range` (string, optional): Time range to analyze (default: "1 hour")
- `sample_size` (integer, optional): Number of events to sample (default: 1000)

**Example**:
```json
{
  "name": "discovery.explore_attributes", 
  "arguments": {
    "event_type": "Transaction"
  }
}
```

**Returns**: List of attributes with data types, cardinality, and sample values

---

### discovery.list_schemas
**Description**: List all available schemas with their structures

**Parameters**:
- `filter` (string, optional): Filter schemas by pattern
- `include_quality` (boolean, optional): Include quality metrics (default: false)

**Example**:
```json
{
  "name": "discovery.list_schemas",
  "arguments": {
    "filter": "Transaction*"
  }
}
```

**Returns**: Schema definitions with field information

---

### discovery.profile_attribute
**Description**: Deep analysis of a specific data attribute

**Parameters**:
- `schema` (string, required): Schema name
- `attribute` (string, required): Attribute to profile
- `time_range` (string, optional): Analysis time range

**Example**:
```json
{
  "name": "discovery.profile_attribute",
  "arguments": {
    "schema": "Transaction",
    "attribute": "duration"
  }
}
```

**Returns**: Statistical profile including min, max, avg, percentiles, distribution

---

### discovery.find_relationships
**Description**: Discover relationships between different event types

**Parameters**:
- `source_type` (string, required): Source event type
- `target_type` (string, optional): Target event type
- `time_range` (string, optional): Time range for analysis

**Example**:
```json
{
  "name": "discovery.find_relationships",
  "arguments": {
    "source_type": "Transaction",
    "target_type": "TransactionError"
  }
}
```

**Returns**: Discovered relationships with correlation strength

## Query Tools

Query tools execute NRQL queries with various optimization and adaptation strategies.

### query_nrdb
**Description**: Execute a standard NRQL query

**Parameters**:
- `query` (string, required): The NRQL query to execute
- `account_id` (string, optional): Target account ID
- `timeout` (integer, optional): Query timeout in seconds (default: 30)

**Example**:
```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT average(duration) FROM Transaction SINCE 1 hour ago"
  }
}
```

**Returns**: Query results with data and metadata

---

### query.execute_adaptive
**Description**: Execute NRQL with automatic optimization

**Parameters**:
- `query` (string, required): The NRQL query
- `optimization_hints` (object, optional): Hints for optimization
- `max_retries` (integer, optional): Maximum optimization retries (default: 3)

**Returns**: Optimized query results with performance metrics

---

### query.validate_nrql
**Description**: Validate NRQL syntax before execution

**Parameters**:
- `query` (string, required): The NRQL query to validate

**Returns**: Validation result with error details if invalid

---

### query.explain_nrql
**Description**: Explain query execution plan and optimization opportunities

**Parameters**:
- `query` (string, required): The NRQL query to explain

**Returns**: Execution plan with optimization suggestions

## Alert Tools

Alert tools manage intelligent alerting based on discovered baselines and patterns.

### alert.create_from_baseline
**Description**: Create an alert based on automatically discovered baselines

**Parameters**:
- `name` (string, required): Alert policy name
- `event_type` (string, required): Event type to monitor
- `metric` (string, required): Metric to alert on
- `sensitivity` (string, optional): Sensitivity level (default: "medium")
- `notification_channels` (array, optional): Notification channel IDs

**Returns**: Created alert policy with calculated thresholds

---

### alert.create_custom
**Description**: Create a custom alert with specified conditions

**Parameters**:
- `name` (string, required): Alert policy name
- `query` (string, required): NRQL query for the condition
- `threshold` (number, required): Alert threshold value
- `comparison` (string, required): Comparison operator
- `duration` (integer, optional): Duration in minutes (default: 5)

**Returns**: Created alert policy details

---

### list_alerts
**Description**: List all alert policies with filtering

**Parameters**:
- `limit` (integer, optional): Number of alerts to return (default: 100)
- `filter` (object, optional): Filter criteria
- `include_conditions` (boolean, optional): Include condition details (default: false)

**Returns**: List of alert policies

## Dashboard Tools

Dashboard tools create and manage visualization dashboards.

### dashboard.create_from_discovery
**Description**: Create a dashboard based on discovered data patterns

**Parameters**:
- `name` (string, required): Dashboard name
- `event_types` (array, required): Event types to include
- `layout` (string, optional): Layout strategy (default: "auto")
- `time_range` (string, optional): Default time range

**Returns**: Created dashboard with auto-generated widgets

---

### dashboard.create_custom
**Description**: Create a custom dashboard with specified widgets

**Parameters**:
- `name` (string, required): Dashboard name
- `widgets` (array, required): Widget configurations
- `layout` (object, optional): Layout configuration

**Returns**: Created dashboard details

---

### find_usage
**Description**: Find dashboards using specific metrics or event types

**Parameters**:
- `search_term` (string, required): Metric, attribute, or event type to search
- `search_type` (string, optional): Search type (default: "any")
- `include_widgets` (boolean, optional): Include widget details

**Returns**: Dashboards and widgets using the search term

## Analysis Tools

Analysis tools provide statistical analysis and pattern detection capabilities.

### analysis.calculate_baseline
**Description**: Calculate statistical baseline for a metric

**Parameters**:
- `metric` (string, required): Metric to analyze
- `event_type` (string, required): Event type containing the metric
- `time_range` (string, optional): Time range for baseline
- `percentiles` (array, optional): Percentiles to calculate

**Returns**: Baseline statistics with confidence intervals

---

### analysis.detect_anomalies
**Description**: Detect anomalies in time series data

**Parameters**:
- `metric` (string, required): Metric to analyze
- `event_type` (string, required): Event type containing the metric
- `time_range` (string, optional): Time range to analyze
- `sensitivity` (number, optional): Detection sensitivity

**Returns**: Detected anomalies with timestamps and severity

---

### analysis.find_correlations
**Description**: Find correlations between metrics

**Parameters**:
- `event_type` (string, required): Event type
- `metrics` (array, required): Metrics to correlate
- `time_range` (string, optional): Time range for analysis
- `min_correlation` (number, optional): Minimum correlation threshold

**Returns**: Correlation matrix with significance values

---

### analysis.analyze_trend
**Description**: Analyze trends and patterns in metric data

**Parameters**:
- `metric` (string, required): Metric to analyze
- `event_type` (string, required): Event type containing the metric
- `time_range` (string, optional): Time range for analysis
- `forecast_periods` (integer, optional): Periods to forecast

**Returns**: Trend analysis with direction and strength

---

### analysis.analyze_distribution
**Description**: Analyze the distribution characteristics of a metric

**Parameters**:
- `metric` (string, required): Metric to analyze
- `event_type` (string, required): Event type containing the metric
- `time_range` (string, optional): Time range to analyze
- `buckets` (integer, optional): Number of histogram buckets

**Returns**: Distribution statistics including skewness and kurtosis

---

### analysis.compare_segments
**Description**: Compare metrics across different segments

**Parameters**:
- `metric` (string, required): Metric to compare
- `event_type` (string, required): Event type containing the metric
- `segment_by` (string, required): Attribute to segment by
- `time_range` (string, optional): Time range for comparison

**Returns**: Segment comparison with statistical significance

## Governance Tools

Governance tools help manage usage, costs, and compliance.

### governance.analyze_usage
**Description**: Analyze data ingest usage patterns

**Parameters**:
- `time_range` (string, optional): Analysis period (default: "7 days")
- `group_by` (string, optional): Grouping strategy

**Returns**: Usage analysis with volume and cost breakdown

---

### governance.optimize_costs
**Description**: Get cost optimization recommendations

**Parameters**:
- `target_reduction` (number, optional): Target reduction percentage
- `preserve_critical` (boolean, optional): Preserve critical data

**Returns**: Optimization recommendations with impact analysis

---

### governance.check_compliance
**Description**: Check data retention and compliance status

**Parameters**:
- `compliance_type` (string, optional): Type of compliance check
- `detailed` (boolean, optional): Include detailed findings

**Returns**: Compliance status with violations if any

## Workflow Tools

Workflow tools orchestrate complex multi-step operations.

### workflow.execute_investigation
**Description**: Execute a complete investigation workflow

**Parameters**:
- `issue_description` (string, required): Description of the issue
- `time_range` (string, optional): Investigation time range
- `auto_create_dashboard` (boolean, optional): Auto-create dashboard

**Returns**: Investigation results with findings and recommendations

---

### workflow.optimize_account
**Description**: Run complete account optimization workflow

**Parameters**:
- `optimization_goals` (array, optional): Optimization goals
- `aggressive` (boolean, optional): Use aggressive optimizations

**Returns**: Optimization plan with expected impact

---

### workflow.generate_report
**Description**: Generate comprehensive report

**Parameters**:
- `report_type` (string, required): Type of report
- `time_range` (string, required): Report period
- `format` (string, optional): Output format

**Returns**: Generated report in requested format

## Session Tools

Session tools manage stateful operations across multiple tool calls.

### session.create
**Description**: Create a new session for stateful operations

**Parameters**:
- `session_id` (string, optional): Custom session ID
- `ttl` (integer, optional): Session TTL in seconds

**Returns**: Session details with ID

---

### session.end
**Description**: End an active session

**Parameters**:
- `session_id` (string, required): Session ID to end

**Returns**: Session closure confirmation

## Tool Composition Patterns

Tools are designed to compose into powerful workflows. Here are common patterns:

### Discovery → Query → Action
1. Use discovery tools to understand data structure
2. Build queries based on discovered attributes
3. Take action (create alerts/dashboards) based on query results

**Example Flow**:
```
discovery.explore_event_types
  → discovery.explore_attributes
    → query_nrdb
      → alert.create_from_baseline
```

### Analysis → Optimization
1. Analyze current state with analysis tools
2. Identify optimization opportunities
3. Apply optimizations with governance tools

**Example Flow**:
```
analysis.calculate_baseline
  → analysis.detect_anomalies
    → governance.optimize_costs
      → workflow.optimize_account
```

### Investigation Workflow
1. Start with symptoms in query tools
2. Use analysis to find patterns
3. Discover related data
4. Create monitoring for future

**Example Flow**:
```
query_nrdb (identify issue)
  → analysis.find_correlations
    → discovery.find_relationships
      → dashboard.create_from_discovery
```

## Best Practices

1. **Always Start with Discovery**: Don't assume data structures
2. **Use Appropriate Time Ranges**: Longer for baselines, shorter for real-time
3. **Leverage Tool Metadata**: Each tool includes examples and guidance
4. **Compose Tools**: Combine simple tools for complex operations
5. **Handle Errors Gracefully**: Tools provide detailed error messages
6. **Use Sessions for Complex Workflows**: Maintain state across operations
7. **Validate Before Executing**: Use validation tools before expensive operations
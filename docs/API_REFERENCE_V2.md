# New Relic MCP Server API Reference v2

This document provides a comprehensive reference for all tools available in the New Relic MCP Server v2 with enhanced granular architecture.

## Table of Contents

1. [Tool Categories](#tool-categories)
2. [Query Tools](#query-tools)
3. [Entity Tools](#entity-tools)
4. [Dashboard Tools](#dashboard-tools)
5. [Alert Tools](#alert-tools)
6. [Bulk Operations](#bulk-operations)
7. [Analysis Tools](#analysis-tools)
8. [Utility Tools](#utility-tools)
9. [Tool Metadata](#tool-metadata)
10. [Error Handling](#error-handling)

## Tool Categories

All tools are organized into categories for better discoverability:

- **query**: Read-only data retrieval operations
- **mutation**: Operations that create, update, or delete resources
- **analysis**: Data analysis and pattern detection
- **utility**: Helper functions and formatters
- **bulk**: Operations that affect multiple resources

## Query Tools

### nrql.execute

Execute a single NRQL query with full control over parameters.

**Category**: query  
**Safety Level**: safe  
**Performance**: 500ms expected, 30s max

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | The NRQL query to execute | - |
| account_id | integer | No | Target account ID | Default account |
| timeout | integer | No | Query timeout in seconds (1-300) | 30 |
| include_metadata | boolean | No | Include query performance metadata | false |

#### Example

```json
{
  "method": "nrql.execute",
  "params": {
    "query": "SELECT count(*) FROM Transaction WHERE appName = 'checkout' SINCE 1 hour ago",
    "timeout": 10,
    "include_metadata": true
  }
}
```

#### Response

```json
{
  "results": [
    {"count": 12345}
  ],
  "metadata": {
    "executionTime": 245,
    "inspectedCount": 50000,
    "cacheHit": false
  }
}
```

### nrql.validate

Validate NRQL syntax without execution.

**Category**: query  
**Safety Level**: safe  
**Performance**: 50ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | The NRQL query to validate | - |
| check_permissions | boolean | No | Verify user permissions for referenced data | false |
| suggest_improvements | boolean | No | Provide optimization suggestions | true |

#### Example

```json
{
  "method": "nrql.validate",
  "params": {
    "query": "SELECT * FROM Transaction",
    "suggest_improvements": true
  }
}
```

#### Response

```json
{
  "valid": true,
  "warnings": ["Query has no time range specified"],
  "suggestions": ["Avoid SELECT *, specify needed attributes for better performance"]
}
```

### nrql.estimate_cost

Estimate query cost and performance impact.

**Category**: analysis  
**Safety Level**: safe  
**Performance**: 100ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | The NRQL query to analyze | - |
| time_range | string | No | Time range (e.g., '1 hour', '7 days') | '1 hour' |
| execution_frequency | string | No | How often: once, hourly, daily, continuous | 'once' |

### nrql.build_select

Build SELECT clause with proper escaping.

**Category**: utility  
**Safety Level**: safe  
**Performance**: 10ms expected

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| event_type | string | Yes | The event type to query |
| aggregations | array | No | List of aggregation specifications |
| attributes | array | No | Raw attributes to select |
| aliases | object | No | Attribute to alias mapping |

#### Aggregation Specification

```json
{
  "function": "percentile",
  "attribute": "duration",
  "percentile": 95
}
```

### nrql.build_where

Build WHERE clause with proper escaping and type handling.

**Category**: utility  
**Safety Level**: safe  
**Performance**: 10ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|
| conditions | array | Yes | List of condition specifications | - |
| operator | string | No | Logical operator: AND, OR | 'AND' |
| nest_groups | boolean | No | Wrap in parentheses | false |

#### Condition Specification

```json
{
  "attribute": "appName",
  "operator": "=",
  "value": "my-app"
}
```

## Entity Tools

### entity.search_by_name

Search entities by name with flexible matching.

**Category**: query  
**Safety Level**: safe  
**Performance**: 200ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| name | string | Yes | Entity name to search | - |
| match_type | string | No | exact, contains, starts_with | 'contains' |
| domain | string | No | Filter by domain (APM, BROWSER, etc.) | - |
| type | string | No | Filter by entity type | - |
| limit | integer | No | Maximum results | 50 |
| cursor | string | No | Pagination cursor | - |

### entity.search_by_tag

Find entities with specific tags.

**Category**: query  
**Safety Level**: safe  
**Performance**: 300ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| tags | object | Yes | Tag key-value pairs to match | - |
| match_all | boolean | No | Require all tags to match | true |
| limit | integer | No | Maximum results | 50 |
| cursor | string | No | Pagination cursor | - |

### entity.search_by_alert_status

Find entities based on their alert status.

**Category**: query  
**Safety Level**: safe  
**Performance**: 400ms expected

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| alert_severity | string | Yes | CRITICAL, WARNING, NOT_ALERTING |
| violation_count_min | integer | No | Minimum violation count |
| time_window | string | No | Time window to check |

### entity.get_golden_metrics

Retrieve golden signal metrics for an entity.

**Category**: query  
**Safety Level**: safe  
**Performance**: 300ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| guid | string | Yes | Entity GUID | - |
| time_range | string | No | Time range for metrics | '1 hour' |

### entity.get_relationships

Get entity relationship graph.

**Category**: query  
**Safety Level**: safe  
**Performance**: 500ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| guid | string | Yes | Entity GUID | - |
| relationship_types | array | No | Filter by relationship types | All types |
| depth | integer | No | Traversal depth | 1 |

## Dashboard Tools

### dashboard.search_by_metric

Find dashboards using specific metrics.

**Category**: query  
**Safety Level**: safe  
**Performance**: 400ms expected

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| metric_name | string | Yes | Metric name to search |
| event_type | string | No | Filter by event type |
| limit | integer | No | Maximum results |

### dashboard.create_widget

Create a single dashboard widget.

**Category**: mutation  
**Safety Level**: caution  
**Dry Run**: Supported

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| dashboard_guid | string | Yes | Target dashboard GUID |
| page_guid | string | Yes | Target page GUID |
| widget | object | Yes | Widget specification |
| dry_run | boolean | No | Preview without creating |

#### Widget Specification

```json
{
  "title": "Error Rate",
  "type": "line",
  "query": "SELECT percentage(count(*), WHERE error IS true) FROM Transaction",
  "layout": {
    "row": 1,
    "column": 1,
    "width": 6,
    "height": 3
  }
}
```

### dashboard.apply_golden_signals_template

Create a complete golden signals dashboard.

**Category**: mutation  
**Safety Level**: caution  
**Dry Run**: Supported

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| entity_guid | string | Yes | APM entity GUID | - |
| time_range | string | No | Default time range | '1 hour' |
| include_dependencies | boolean | No | Include dependent services | false |
| account_ids | array | No | Cross-account support | - |
| dry_run | boolean | No | Preview configuration | false |

## Alert Tools

### alert.find_by_entity

Find all alert conditions for an entity.

**Category**: query  
**Safety Level**: safe  
**Performance**: 300ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| entity_guid | string | Yes | Entity GUID | - |
| include_inherited | boolean | No | Include policy-level alerts | true |
| status_filter | array | No | Filter by status | All |

### alert.create_threshold_condition

Create a threshold-based alert condition.

**Category**: mutation  
**Safety Level**: destructive  
**Dry Run**: Supported

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| policy_id | string | Yes | Alert policy ID |
| name | string | Yes | Condition name |
| query | string | Yes | NRQL query |
| threshold | object | Yes | Threshold specification |
| duration | object | Yes | Duration specification |
| dry_run | boolean | No | Preview without creating |

#### Threshold Specification

```json
{
  "value": 100,
  "operator": "ABOVE",
  "duration_minutes": 5,
  "occurrences": "AT_LEAST_ONCE"
}
```

### alert.mute_condition

Temporarily mute an alert condition.

**Category**: mutation  
**Safety Level**: caution

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| condition_id | string | Yes | Condition ID |
| duration_minutes | integer | Yes | Mute duration |
| reason | string | Yes | Reason for muting |

## Bulk Operations

### bulk.add_tags

Add tags to multiple entities.

**Category**: bulk  
**Safety Level**: caution  
**Dry Run**: Supported

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| entity_guids | array | Yes | List of entity GUIDs | - |
| tags | object | Yes | Tags to add | - |
| skip_on_error | boolean | No | Continue on failures | true |
| dry_run | boolean | No | Preview changes | false |

### bulk.execute_queries_parallel

Execute multiple queries in parallel.

**Category**: bulk  
**Safety Level**: safe  
**Performance**: Varies

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| queries | array | Yes | List of query specifications | - |
| max_concurrent | integer | No | Max parallel executions | 5 |
| stop_on_error | boolean | No | Stop on first error | false |

## Analysis Tools

### analysis.detect_anomalies

Detect anomalies in time series data.

**Category**: analysis  
**Safety Level**: safe  
**Performance**: 2s expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | NRQL query for time series | - |
| sensitivity | float | No | Detection sensitivity (0-1) | 0.8 |
| baseline_window | string | No | Baseline period | '7 days' |
| detection_window | string | No | Detection period | '1 hour' |

### analysis.find_correlations

Find correlated metrics.

**Category**: analysis  
**Safety Level**: safe  
**Performance**: 5s expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| primary_query | string | Yes | Primary metric query | - |
| candidate_queries | array | Yes | Queries to test correlation | - |
| min_correlation | float | No | Minimum correlation coefficient | 0.7 |
| time_range | string | No | Analysis period | '24 hours' |

## Utility Tools

### utility.escape_nrql_string

Properly escape strings for NRQL queries.

**Category**: utility  
**Safety Level**: safe  
**Performance**: 1ms expected

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| value | string | Yes | String to escape |
| context | string | Yes | WHERE, SELECT, or FACET |

### utility.generate_golden_signal_queries

Generate standard golden signal queries.

**Category**: utility  
**Safety Level**: safe  
**Performance**: 10ms expected

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| entity_type | string | Yes | Type of entity |
| entity_name | string | No | Specific entity name |
| custom_attributes | array | No | Additional attributes |

## Platform Governance Tools

### dashboard.list_widgets

Inventory all dashboard widgets with their configurations.

**Category**: governance  
**Safety Level**: safe  
**Performance**: 5s expected for 100 dashboards

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| cursor | string | No | Pagination cursor | null |
| account_id | int | No | Filter by account | current |

#### Returns

```json
{
  "widgets": [
    {
      "dashboardGuid": "MXxEQVNIQk9BUkR8MTIzNDU",
      "dashboardName": "Production Overview",
      "widgetId": "widget-1",
      "type": "line",
      "visualization": "viz.line",
      "rawConfiguration": "{...}"
    }
  ],
  "nextCursor": "..."
}
```

### dashboard.classify_widgets

Classify widgets as dimensional-metric-based or event-NRQL-based.

**Category**: governance  
**Safety Level**: safe  
**Performance**: 500ms per dashboard

#### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| dashboard_guid | string | Yes | Dashboard to analyze |

#### Returns

```json
{
  "dashboardGuid": "MXxEQVNIQk9BUkR8MTIzNDU",
  "metricWidgets": 12,
  "eventWidgets": 34,
  "metricNames": ["http.server.duration", "cpu.usage"],
  "eventTypes": ["Transaction", "PageView"],
  "classification": {
    "percentMetrics": 26.1,
    "percentEvents": 73.9
  }
}
```

### dashboard.find_nrdot_dashboards

Find dashboards using NR1 Data Explorer (NRDOT).

**Category**: governance  
**Safety Level**: safe  
**Performance**: 3s expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| account_id | int | No | Filter by account | all |

### metric.widget_usage_rank

Rank metrics by their usage across dashboard widgets.

**Category**: governance  
**Safety Level**: safe  
**Performance**: 2s expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| limit | int | No | Top N metrics | 50 |
| time_range | string | No | Analysis window | '30 days' |

#### Returns

```json
{
  "rankings": [
    {
      "metricName": "http.server.duration",
      "widgetCount": 45,
      "dashboards": ["Dashboard1", "Dashboard2", ...],
      "percentageOfTotal": 12.3
    }
  ]
}
```

### usage.ingest_summary

Get total ingest volume with breakdown by source.

**Category**: governance  
**Safety Level**: safe  
**Performance**: 2s expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| period | string | No | Time window | '30d' |
| account_id | int | No | Specific account | current |

#### Returns

```json
{
  "totalBytes": 10995116277760,
  "totalGB": 10240,
  "breakdown": [
    {"source": "OTLP", "bytes": 6597069766656, "percentage": 60},
    {"source": "AGENT", "bytes": 3298534883328, "percentage": 30},
    {"source": "API", "bytes": 1099511627776, "percentage": 10}
  ]
}
```

### usage.otlp_collectors

Analyze OTEL collector ingest volumes.

**Category**: governance  
**Safety Level**: safe  
**Performance**: 3s expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| period | string | No | Time window | '30d' |

#### Returns

```json
{
  "collectors": [
    {
      "name": "otel-payment-prod",
      "metricCount": 15000000,
      "bytesEstimate": 120000000,
      "percentageOfOtlp": 40
    }
  ],
  "totalOtlpBytes": 6597069766656
}
```

### usage.agent_ingest

Get native agent ingest statistics.

**Category**: governance  
**Safety Level**: safe  
**Performance**: 2s expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| period | string | No | Time window | '30d' |

#### Returns

```json
{
  "agents": [
    {"name": "Infrastructure", "bytes": 1649267441664},
    {"name": "APM", "bytes": 1099511627776}
  ],
  "comparison": {
    "agentBytes": 3298534883328,
    "otelBytes": 6597069766656,
    "ratio": 0.5
  }
}
```

## Tool Metadata

Each tool includes rich metadata for AI guidance:

### Safety Metadata
- `level`: safe, caution, destructive
- `dry_run_supported`: boolean
- `requires_confirmation`: boolean
- `affected_resources`: array of resource types

### Performance Metadata
- `expected_latency_ms`: typical execution time
- `max_latency_ms`: timeout threshold
- `cacheable`: whether results can be cached
- `cache_ttl_seconds`: cache duration

### AI Guidance Metadata
- `usage_examples`: common use cases
- `chains_with`: tools commonly used together
- `warnings_for_ai`: important considerations
- `error_patterns`: common errors and fixes

## Error Handling

All tools return consistent error responses:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Required parameter 'query' is missing",
    "details": {
      "parameter": "query",
      "type": "required"
    }
  }
}
```

### Error Codes

- `VALIDATION_ERROR`: Invalid input parameters
- `PERMISSION_ERROR`: Insufficient permissions
- `NOT_FOUND`: Resource not found
- `TIMEOUT_ERROR`: Operation timed out
- `RATE_LIMIT_ERROR`: Rate limit exceeded
- `INTERNAL_ERROR`: Server error

## Best Practices

1. **Use atomic tools**: Compose complex operations from simple tools
2. **Validate first**: Use validation tools before expensive operations
3. **Handle pagination**: Use cursors for large result sets
4. **Respect rate limits**: Check rate limit headers
5. **Use dry run**: Test mutations with dry_run=true first
6. **Cache when possible**: Reuse results for identical queries
7. **Check permissions**: Validate access before operations

## Migration from v1

Key differences from v1:
- Tools are more granular and atomic
- Enhanced metadata for better AI guidance
- Consistent parameter naming
- Improved error handling
- Built-in dry run support for mutations

See [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md) for detailed migration instructions.
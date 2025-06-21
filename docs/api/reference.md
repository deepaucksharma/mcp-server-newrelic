# New Relic MCP Server API Reference

This comprehensive reference documents all tools available in the New Relic MCP Server, organized by category and purpose.

## Table of Contents

1. [Overview](#overview)
2. [Transport Protocols](#transport-protocols)
3. [Tool Categories](#tool-categories)
4. [Discovery Tools](#discovery-tools)
5. [Query Tools](#query-tools)
6. [Analysis Tools](#analysis-tools)
7. [Action Tools](#action-tools)
8. [Governance Tools](#governance-tools)
9. [Utility Tools](#utility-tools)
10. [Workflow Tools](#workflow-tools)
11. [Bulk Operations](#bulk-operations)
12. [Error Handling](#error-handling)
13. [Performance Guidelines](#performance-guidelines)
14. [Troubleshooting](#troubleshooting)

## Overview

The New Relic MCP Server provides AI assistants with intelligent access to New Relic observability data through specialized tools. Each tool is designed to be atomic, composable, and safe.

### Design Principles

1. **Atomic Operations**: Each tool performs one specific task
2. **Discovery-First**: Start with discovery tools to understand available data
3. **Safe by Default**: Read operations are safe, mutations require confirmation
4. **Performance Aware**: Tools include latency expectations and caching hints
5. **AI-Optimized**: Rich metadata guides proper tool usage

### Tool Metadata Structure

Every tool includes:
- **Category**: Query, Mutation, Analysis, Utility, Bulk, Governance
- **Safety Level**: Safe, Caution, Destructive
- **Performance**: Expected latency, caching policy, resource usage
- **AI Guidance**: Usage examples, common patterns, error handling

## Transport Protocols

### STDIO (Standard I/O)

Default transport for command-line usage:

```bash
# Direct invocation
./bin/mcp-server

# With MCP inspector
npx @modelcontextprotocol/inspector ./bin/mcp-server

# In Claude Desktop config
{
  "mcpServers": {
    "newrelic": {
      "command": "/path/to/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "your-key",
        "NEW_RELIC_ACCOUNT_ID": "12345"
      }
    }
  }
}
```

### HTTP Transport

RESTful API endpoint:

```bash
# Start HTTP server
./bin/mcp-server --transport http --port 8080

# Make requests
curl -X POST http://localhost:8080/v1/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "nrql.execute",
    "params": {
      "query": "SELECT count(*) FROM Transaction"
    }
  }'
```

### SSE (Server-Sent Events)

For streaming results:

```javascript
const eventSource = new EventSource('http://localhost:8080/v1/tools/stream');
eventSource.onmessage = (event) => {
  const result = JSON.parse(event.data);
  console.log('Received:', result);
};
```

## Multi-Account Support

Many tools support querying across multiple New Relic accounts without reconfiguration. Simply add the `account_id` parameter to any supporting tool:

```json
{
  "tool": "nrql.execute",
  "params": {
    "query": "SELECT count(*) FROM Transaction",
    "account_id": "2345678"  // Different from default account
  }
}
```

### Account ID Parameter

- **Type**: string
- **Required**: No (uses default account if not specified)
- **Format**: New Relic account ID as string
- **Supported Tools**: All query, dashboard, alert, and discovery tools

### Cross-Account Dashboards

For dashboards that aggregate data from multiple accounts:

```json
{
  "tool": "dashboard.create",
  "params": {
    "name": "Multi-Account Overview",
    "account_ids": ["123456", "234567", "345678"],
    "template": "cross-account-summary"
  }
}
```

## Tool Categories

### Query (Read-Only Operations)
- **Purpose**: Safe data retrieval
- **Safety**: Always safe
- **Caching**: Usually cacheable
- **Examples**: `nrql.execute`, `entity.search_by_name`, `dashboard.list`

### Mutation (Write Operations)
- **Purpose**: Create, update, or delete resources
- **Safety**: Caution or Destructive
- **Dry Run**: Always supported
- **Examples**: `alert.create`, `dashboard.create_widget`, `entity.add_tags`

### Analysis (Data Processing)
- **Purpose**: Complex calculations and pattern detection
- **Safety**: Safe but resource-intensive
- **Performance**: May be slow
- **Examples**: `analysis.detect_anomalies`, `analysis.find_correlations`

### Utility (Helper Functions)
- **Purpose**: Format, validate, and transform data
- **Safety**: Always safe
- **Performance**: Fast (< 100ms)
- **Examples**: `nrql.build_where`, `utility.escape_string`

### Governance (Platform Management)
- **Purpose**: Usage analysis and optimization
- **Safety**: Safe
- **Use Case**: Cost optimization, adoption tracking
- **Examples**: `usage.ingest_summary`, `metric.widget_usage_rank`

### Workflow (Multi-Step Orchestration)
- **Purpose**: Complex investigations and responses
- **Safety**: Depends on steps
- **State**: Maintains context across steps
- **Examples**: `workflow.create`, `investigation.start`

## Discovery Tools

Discovery tools help explore data without assumptions. Always start here when investigating new environments.

### discovery.explore_event_types

Discover what event types exist in NRDB.

**Category**: Query  
**Safety**: Safe  
**Performance**: 1-5s expected, cacheable for 1 hour

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| time_range | string | No | How far back to explore | "24 hours" |
| include_samples | boolean | No | Include sample events | true |
| min_event_count | integer | No | Minimum events to consider active | 10 |

#### Example

```json
{
  "tool": "discovery.explore_event_types",
  "params": {
    "time_range": "6 hours",
    "include_samples": true,
    "min_event_count": 100
  }
}
```

#### Response

```json
{
  "eventTypes": [
    {
      "eventType": "Transaction",
      "count": 1234567,
      "firstSeen": "2024-01-20T08:00:00Z",
      "lastSeen": "2024-01-20T13:55:00Z",
      "sampleEvent": {
        "duration": 123.45,
        "appName": "checkout-service"
      }
    }
  ],
  "totalTypes": 15,
  "dataCompleteness": 0.92
}
```

### discovery.explore_attributes

Discover attributes for a specific event type.

**Category**: Query  
**Safety**: Safe  
**Performance**: 500ms expected

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| event_type | string | Yes | Event type to explore | - |
| sample_size | integer | No | Events to sample | 1000 |
| show_coverage | boolean | No | Calculate null percentages | true |
| show_examples | boolean | No | Show example values | true |

### discovery.profile_data_completeness

Analyze data quality and reliability.

**Category**: Analysis  
**Safety**: Safe  
**Performance**: 2-10s, resource intensive

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| event_type | string | Yes | Event type to profile | - |
| critical_attributes | array | No | Must-have attributes | [] |
| time_range | string | No | Analysis period | "24 hours" |
| check_patterns | boolean | No | Detect collection gaps | true |

### discovery.find_natural_groupings

Discover how data naturally segments.

**Category**: Analysis  
**Safety**: Safe  
**Performance**: 3-15s, resource intensive

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| event_type | string | Yes | Event type to analyze | - |
| max_groups | integer | No | Maximum groupings | 10 |
| min_group_size | integer | No | Minimum group size | 100 |
| attributes_to_consider | array | No | Specific attributes | null (auto) |

### discovery.find_data_relationships

Find relationships between event types.

**Category**: Analysis  
**Safety**: Safe  
**Performance**: 5-30s, very resource intensive

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| source_event_type | string | Yes | Primary event type | - |
| target_event_types | array | No | Types to check | null (all) |
| relationship_types | array | No | Types to find | ["join_key", "temporal"] |
| sample_size | integer | No | Events to analyze | 10000 |

## Query Tools

Query tools provide controlled access to NRQL queries and data retrieval.

### nrql.execute

Execute a NRQL query with full control.

**Category**: Query  
**Safety**: Safe  
**Performance**: 500ms typical, 30s max

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | NRQL query to execute | - |
| account_id | string | No | Target account ID | Default account |
| timeout | integer | No | Timeout in seconds (1-300) | 30 |
| include_metadata | boolean | No | Include performance data | false |

#### Example

```json
{
  "tool": "nrql.execute",
  "params": {
    "query": "SELECT percentile(duration, 95) FROM Transaction WHERE appName = 'checkout' SINCE 1 hour ago",
    "timeout": 10,
    "include_metadata": true
  }
}
```

#### Response

```json
{
  "results": [
    {"percentile.duration.95": 234.56}
  ],
  "metadata": {
    "executionTime": 245,
    "inspectedCount": 50000,
    "cacheHit": false
  }
}
```

### nrql.validate

Validate query syntax without execution.

**Category**: Query  
**Safety**: Safe  
**Performance**: 50ms

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | Query to validate | - |
| check_permissions | boolean | No | Verify access rights | false |
| suggest_improvements | boolean | No | Optimization tips | true |

### nrql.estimate_cost

Estimate query performance impact.

**Category**: Analysis  
**Safety**: Safe  
**Performance**: 100ms

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | Query to analyze | - |
| time_range | string | No | Query time range | "1 hour" |
| execution_frequency | string | No | How often | "once" |

### nrql.build_select

Build SELECT clause programmatically.

**Category**: Utility  
**Safety**: Safe  
**Performance**: 10ms

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| event_type | string | Yes | Event type | - |
| aggregations | array | No | Aggregation specs | [] |
| attributes | array | No | Raw attributes | [] |
| aliases | object | No | Attribute aliases | {} |

### nrql.build_where

Build WHERE clause with proper escaping.

**Category**: Utility  
**Safety**: Safe  
**Performance**: 10ms

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| conditions | array | Yes | Condition specs | - |
| operator | string | No | AND or OR | "AND" |
| nest_groups | boolean | No | Add parentheses | false |

## Analysis Tools

Analysis tools perform complex calculations and pattern detection.

### analysis.detect_anomalies

Find anomalies in time series data.

**Category**: Analysis  
**Safety**: Safe  
**Performance**: 2s typical, CPU intensive

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | Base NRQL query | - |
| sensitivity | float | No | Detection sensitivity (0-1) | 0.8 |
| baseline_window | string | No | Historical baseline | "7 days" |
| detection_window | string | No | Analysis window | "1 hour" |
| anomaly_types | array | No | Types to detect | ["spike", "drop", "pattern"] |

#### Response

```json
{
  "anomalies": [
    {
      "timestamp": "2024-01-20T14:30:00Z",
      "type": "spike",
      "severity": "high",
      "value": 567.8,
      "expected_range": [100, 200],
      "deviation_sigma": 4.2,
      "confidence": 0.95
    }
  ],
  "baseline_stats": {
    "mean": 150,
    "stddev": 25,
    "percentiles": {
      "p50": 145,
      "p95": 195,
      "p99": 210
    }
  }
}
```

### analysis.find_correlations

Find correlated metrics across data sources.

**Category**: Analysis  
**Safety**: Safe  
**Performance**: 5s typical, very CPU intensive

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| primary_query | string | Yes | Primary metric query | - |
| candidate_queries | array | Yes | Queries to correlate | - |
| min_correlation | float | No | Minimum coefficient | 0.7 |
| time_range | string | No | Analysis period | "24 hours" |
| lag_windows | array | No | Time lags to test | [0, 5, 10, 30] |

### analysis.forecast_trend

Predict future values based on historical data.

**Category**: Analysis  
**Safety**: Safe  
**Performance**: 3s typical

#### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| query | string | Yes | Historical data query | - |
| forecast_window | string | Yes | How far to predict | - |
| model_type | string | No | Forecast model | "auto" |
| include_confidence | boolean | No | Confidence intervals | true |
| seasonality | string | No | Expected pattern | "auto" |

## Action Tools

Action tools create, update, or delete New Relic resources.

### Entity Management

#### entity.add_tags

Add tags to entities.

**Category**: Mutation  
**Safety**: Caution  
**Dry Run**: Supported

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| entity_guid | string | Yes | Entity to tag | - |
| tags | object | Yes | Key-value pairs | - |
| replace_existing | boolean | No | Replace vs merge | false |
| dry_run | boolean | No | Preview only | false |

### Dashboard Management

#### dashboard.create

Create a new dashboard.

**Category**: Mutation  
**Safety**: Caution  
**Dry Run**: Supported

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| name | string | Yes | Dashboard name | - |
| description | string | No | Description | "" |
| permissions | string | No | Visibility | "PUBLIC_READ_WRITE" |
| pages | array | Yes | Page definitions | - |
| dry_run | boolean | No | Preview only | false |

##### Page Definition

```json
{
  "name": "Overview",
  "widgets": [
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
  ]
}
```

#### dashboard.apply_template

Create dashboard from template.

**Category**: Mutation  
**Safety**: Caution  
**Dry Run**: Supported

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| template_name | string | Yes | Template to use | - |
| entity_guid | string | Yes | Target entity | - |
| customizations | object | No | Template params | {} |
| dry_run | boolean | No | Preview only | false |

##### Available Templates
- `golden_signals`: Four golden signals dashboard
- `slo_dashboard`: SLO tracking dashboard
- `kubernetes_cluster`: K8s cluster overview
- `application_deep_dive`: APM deep dive
- `infrastructure_host`: Host monitoring

### Alert Management

#### alert.create_condition

Create alert condition.

**Category**: Mutation  
**Safety**: Destructive  
**Dry Run**: Supported

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| policy_id | string | Yes | Alert policy | - |
| name | string | Yes | Condition name | - |
| query | string | Yes | NRQL query | - |
| threshold | object | Yes | Alert threshold | - |
| duration | object | Yes | Duration config | - |
| dry_run | boolean | No | Preview only | false |

##### Threshold Configuration

```json
{
  "value": 100,
  "operator": "ABOVE",
  "duration_minutes": 5,
  "occurrences": "AT_LEAST_ONCE"
}
```

## Governance Tools

Governance tools help manage platform usage and adoption.

### Usage Analysis

#### usage.ingest_summary

Get data ingest breakdown.

**Category**: Governance  
**Safety**: Safe  
**Performance**: 2s, cacheable for 1 hour

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| period | string | No | Time window | "30d" |
| account_id | integer | No | Specific account | Current |
| group_by | string | No | Grouping | "source" |

##### Response

```json
{
  "totalBytes": 10995116277760,
  "totalGB": 10240,
  "breakdown": [
    {
      "source": "OTLP",
      "bytes": 6597069766656,
      "percentage": 60,
      "trend": "+15%"
    },
    {
      "source": "AGENT",
      "bytes": 3298534883328,
      "percentage": 30,
      "trend": "-5%"
    }
  ]
}
```

#### usage.otlp_collectors

Analyze OpenTelemetry collector volumes.

**Category**: Governance  
**Safety**: Safe  
**Performance**: 3s

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| period | string | No | Analysis window | "30d" |
| sort_by | string | No | Sort metric | "bytes" |
| limit | integer | No | Result limit | 50 |

### Dashboard Governance

#### dashboard.list_widgets

Inventory all dashboard widgets.

**Category**: Governance  
**Safety**: Safe  
**Performance**: 5s for 100 dashboards

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| cursor | string | No | Pagination cursor | null |
| account_id | integer | No | Filter account | Current |
| include_config | boolean | No | Full configs | false |

## Utility Tools

Utility tools provide formatting, validation, and helper functions.

### String Escaping

#### utility.escape_nrql_string

Properly escape strings for NRQL.

**Category**: Utility  
**Safety**: Safe  
**Performance**: 1ms

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| value | string | Yes | String to escape | - |
| context | string | Yes | WHERE, SELECT, FACET | - |

### Query Generation

#### utility.generate_golden_signal_queries

Generate standard observability queries.

**Category**: Utility  
**Safety**: Safe  
**Performance**: 10ms

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| entity_type | string | Yes | Type of entity | - |
| entity_name | string | No | Specific entity | null |
| time_range | string | No | Query time range | "1 hour" |
| custom_attributes | array | No | Extra attributes | [] |

## Workflow Tools

Workflow tools enable complex multi-step operations with state management.

### Workflow Management

#### workflow.create

Initialize a new workflow execution.

**Category**: Workflow  
**Safety**: Safe  
**Performance**: 50ms

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| name | string | Yes | Workflow name | - |
| workflow_type | string | Yes | Type of workflow | - |
| context | object | No | Initial context | {} |
| steps | array | No | Predefined steps | [] |

##### Workflow Types
- `investigation`: Problem investigation
- `incident_response`: Incident handling
- `capacity_planning`: Resource planning
- `slo_management`: SLO workflows
- `optimization`: Performance tuning

### Investigation Workflows

#### investigation.start_latency

Begin latency investigation workflow.

**Category**: Workflow  
**Safety**: Safe  
**Performance**: 100ms to start

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| entity_guid | string | Yes | Entity to investigate | - |
| baseline_window | string | No | Normal period | "7 days" |
| comparison_window | string | No | Problem period | "1 hour" |
| auto_execute | boolean | No | Run all steps | false |

## Bulk Operations

Bulk operations efficiently handle multiple resources.

### Bulk Query

#### bulk.execute_queries

Run multiple queries in parallel.

**Category**: Bulk  
**Safety**: Safe  
**Performance**: Varies with query count

##### Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| queries | array | Yes | Query specifications | - |
| max_concurrent | integer | No | Parallelism limit | 5 |
| stop_on_error | boolean | No | Fail fast | false |
| timeout_per_query | integer | No | Individual timeout | 30 |

## Error Handling

All tools return consistent error structures for proper handling.

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Required parameter 'query' is missing",
    "details": {
      "parameter": "query",
      "type": "required",
      "received": null
    },
    "suggestions": [
      "Provide a valid NRQL query string",
      "Example: SELECT count(*) FROM Transaction"
    ]
  }
}
```

### Error Codes

| Code | Description | Recovery Action |
|------|-------------|-----------------|
| VALIDATION_ERROR | Invalid input parameters | Check parameter requirements |
| PERMISSION_ERROR | Insufficient permissions | Verify API key permissions |
| NOT_FOUND | Resource doesn't exist | Verify GUID/ID is correct |
| TIMEOUT_ERROR | Operation timed out | Reduce query complexity or time range |
| RATE_LIMIT_ERROR | Too many requests | Implement backoff and retry |
| QUERY_ERROR | Invalid NRQL syntax | Use nrql.validate first |
| INTERNAL_ERROR | Server error | Retry with exponential backoff |

### Error Handling Best Practices

1. **Always validate inputs first**
   ```json
   {
     "tool": "nrql.validate",
     "params": { "query": "..." }
   }
   ```

2. **Use dry run for mutations**
   ```json
   {
     "tool": "alert.create",
     "params": {
       "...": "...",
       "dry_run": true
     }
   }
   ```

3. **Handle timeouts gracefully**
   - Start with shorter time ranges
   - Increase timeout parameter
   - Simplify complex queries

4. **Implement retry logic**
   ```python
   MAX_RETRIES = 3
   BACKOFF_FACTOR = 2
   
   for attempt in range(MAX_RETRIES):
       try:
           result = call_tool(...)
           break
       except RateLimitError:
           sleep(BACKOFF_FACTOR ** attempt)
   ```

## Performance Guidelines

### Latency Expectations

| Category | Expected | Maximum | Caching |
|----------|----------|---------|---------|
| Utility | 1-10ms | 100ms | Not needed |
| Query | 200-1000ms | 30s | 5 minutes |
| Analysis | 1-5s | 60s | 15 minutes |
| Governance | 2-10s | 60s | 1 hour |
| Bulk | Varies | 5 min | Depends |

### Performance Optimization

1. **Use appropriate time ranges**
   - Start with smaller ranges (1 hour)
   - Expand only if needed
   - Use LIMIT clause

2. **Leverage caching**
   - Discovery results: Cache 1 hour
   - Query results: Cache 5 minutes
   - Governance data: Cache 1 hour

3. **Batch operations**
   - Use bulk tools for multiple operations
   - Set appropriate batch sizes
   - Monitor rate limits

4. **Query optimization**
   ```sql
   -- Good: Specific time range and limit
   SELECT count(*) FROM Transaction 
   WHERE appName = 'checkout' 
   SINCE 1 hour ago 
   LIMIT 100
   
   -- Bad: No time range or limit
   SELECT * FROM Transaction
   ```

## Troubleshooting

### Common Issues

#### 1. Empty Query Results

**Symptoms**: Query returns no data  
**Diagnosis Steps**:
1. Check time range includes recent data
2. Verify event type exists: `discovery.explore_event_types`
3. Confirm attribute names: `discovery.explore_attributes`
4. Test with broader query

**Solution**:
```json
{
  "tool": "nrql.execute",
  "params": {
    "query": "SELECT count(*) FROM Transaction SINCE 1 day ago"
  }
}
```

#### 2. Timeout Errors

**Symptoms**: Query times out  
**Diagnosis Steps**:
1. Check query complexity
2. Reduce time range
3. Add LIMIT clause
4. Use sampling

**Solution**:
```json
{
  "tool": "nrql.execute",
  "params": {
    "query": "SELECT count(*) FROM Transaction SAMPLE 1000 SINCE 1 hour ago",
    "timeout": 60
  }
}
```

#### 3. Permission Errors

**Symptoms**: Access denied errors  
**Diagnosis Steps**:
1. Verify API key permissions
2. Check account access
3. Confirm resource ownership

**Solution**:
- Ensure API key has required permissions:
  - NRQL: `NRDB Query`
  - Dashboards: `Dashboard Modify`
  - Alerts: `Alert Conditions`

#### 4. Rate Limiting

**Symptoms**: 429 errors  
**Headers to monitor**:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1642531200
```

### Debug Mode

Enable debug logging:

```bash
export MCP_DEBUG=true
export LOG_LEVEL=DEBUG
./bin/mcp-server
```

### Getting Help

1. **Check tool metadata**
   ```json
   {
     "tool": "system.describe_tool",
     "params": {
       "tool_name": "nrql.execute"
     }
   }
   ```

2. **List available tools**
   ```json
   {
     "tool": "system.list_tools",
     "params": {
       "category": "query"
     }
   }
   ```

## Appendix: Quick Reference

### Most Used Tools by Category

**Starting an Investigation**
1. `discovery.explore_event_types` - What data exists?
2. `discovery.explore_attributes` - What fields are available?
3. `entity.search_by_name` - Find specific resources

**Analyzing Performance**
1. `nrql.execute` - Run queries
2. `analysis.detect_anomalies` - Find issues
3. `analysis.find_correlations` - Understand relationships

**Creating Resources**
1. `dashboard.create` - New dashboards
2. `alert.create_condition` - New alerts
3. `dashboard.apply_template` - From templates

**Platform Governance**
1. `usage.ingest_summary` - Data volume
2. `dashboard.classify_widgets` - Dashboard audit
3. `metric.widget_usage_rank` - Popular metrics

### NRQL Query Patterns

```sql
-- Error rate
SELECT percentage(count(*), WHERE error IS true) 
FROM Transaction 
WHERE appName = 'checkout' 
SINCE 1 hour ago

-- Percentile latency
SELECT percentile(duration, 50, 95, 99) 
FROM Transaction 
WHERE appName = 'checkout' 
SINCE 1 hour ago 
TIMESERIES

-- Top errors
SELECT count(*), latest(error.message) 
FROM TransactionError 
WHERE appName = 'checkout' 
FACET error.class 
SINCE 1 hour ago 
LIMIT 10

-- Rate of change
SELECT rate(count(*), 1 minute) 
FROM Transaction 
WHERE appName = 'checkout' 
SINCE 1 hour ago 
TIMESERIES
```

### Common Workflows

**Investigate Latency Spike**
```json
[
  { "tool": "discovery.explore_event_types" },
  { "tool": "entity.search_by_name", "params": { "name": "checkout" } },
  { "tool": "analysis.detect_anomalies", "params": { "query": "..." } },
  { "tool": "analysis.find_correlations", "params": { "..." } },
  { "tool": "dashboard.create", "params": { "..." } }
]
```

**Setup Monitoring**
```json
[
  { "tool": "entity.get_golden_metrics" },
  { "tool": "utility.generate_golden_signal_queries" },
  { "tool": "dashboard.apply_template", "params": { "template_name": "golden_signals" } },
  { "tool": "alert.create_condition", "params": { "..." } }
]
```

**Optimize Costs**
```json
[
  { "tool": "usage.ingest_summary" },
  { "tool": "usage.otlp_collectors" },
  { "tool": "dashboard.list_widgets" },
  { "tool": "metric.widget_usage_rank" }
]
```

---

This completes the comprehensive API reference for the New Relic MCP Server. For the latest updates and additional examples, refer to the project repository and changelog.

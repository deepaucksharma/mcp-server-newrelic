# MCP Server Tools Reference

This document lists the tools that are **actually implemented** in the MCP Server for New Relic. Each tool is marked with its implementation status.

## Implementation Status Legend

- ✅ **Fully Implemented** - Works with real New Relic data
- 🟨 **Mock Only** - Returns mock data, no real implementation
- ⚠️ **Partial** - Basic functionality, missing features
- ❌ **Not Implemented** - Registered but handler missing/broken

## Query Tools

### query_nrdb ✅
Execute NRQL queries against New Relic.

**Parameters:**
- `query` (string, required) - The NRQL query to execute
- `account_id` (string) - Optional account ID
- `timeout` (integer) - Query timeout in seconds (default: 30)

**Example:**
```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT count(*) FROM Transaction WHERE appName = 'my-app' SINCE 1 hour ago"
  }
}
```

**Status:** Fully implemented. Executes real NRQL queries via NerdGraph API.

### query_check ⚠️
Validate an NRQL query and analyze its performance impact.

**Parameters:**
- `query` (string, required) - The NRQL query to validate
- `suggest_improvements` (boolean) - Suggest query optimizations (default: true)

**Status:** Partial implementation. Basic syntax validation works, but performance analysis is mocked.

### query_assist 🟨
Get help building NRQL queries.

**Parameters:**
- `description` (string, required) - Natural language description
- `event_type` (string) - Event type to query

**Status:** Mock only. Returns example queries but doesn't adapt to actual schema.

## Discovery Tools

### discovery.explore_event_types ⚠️
List available event types in your New Relic account.

**Parameters:**
- `limit` (integer) - Maximum number to return (default: 100)
- `search` (string) - Filter event types by name

**Example:**
```json
{
  "name": "discovery.explore_event_types",
  "arguments": {
    "limit": 10
  }
}
```

**Status:** Partial. Basic listing works but missing sample counts and metadata.

### discovery.explore_attributes ⚠️
Explore attributes for a specific event type.

**Parameters:**
- `event_type` (string, required) - Event type to explore
- `include_samples` (boolean) - Include sample values (default: false)

**Status:** Partial. Lists attributes but profiling features are mocked.

### discovery.list_schemas 🟨
List all available schemas in the data source.

**Parameters:**
- `filter` (string) - Optional filter for schema names
- `include_quality` (boolean) - Include quality metrics

**Status:** Mock only. Schema discovery engine not implemented.

### discovery.profile_attribute 🟨
Deep analysis of a specific data attribute.

**Parameters:**
- `schema` (string, required) - Schema name
- `attribute` (string, required) - Attribute to profile
- `time_range` (string) - Analysis time range

**Status:** Mock only. Returns fake statistical analysis.

## Alert Tools

### create_alert ⚠️
Create an alert condition.

**Parameters:**
- `name` (string, required) - Alert condition name
- `query` (string, required) - NRQL query for the alert
- `sensitivity` (string) - low/medium/high (default: medium)
- `comparison` (string) - above/below/equals (default: above)
- `threshold_duration` (integer) - Minutes threshold must be violated (default: 5)
- `auto_baseline` (boolean) - Use automatic baseline detection (default: true)
- `static_threshold` (number) - Static threshold value (if auto_baseline is false)
- `policy_id` (string) - Alert policy ID

**Status:** Partial. Basic alert creation works but auto-baseline is mocked.

### list_alerts ⚠️
List alert conditions.

**Parameters:**
- `policy_id` (string) - Filter by alert policy ID
- `enabled_only` (boolean) - Only show enabled alerts (default: false)
- `include_incidents` (boolean) - Include recent incidents (default: false)
- `limit` (integer) - Maximum number per page (default: 50)
- `cursor` (string) - Pagination cursor

**Status:** Partial. Basic listing works but incident data is mocked.

## Dashboard Tools

### find_usage 🟨
Find dashboards using specific metrics or event types.

**Parameters:**
- `search_term` (string, required) - Metric, attribute, or event type to search
- `search_type` (string) - metric/attribute/event_type/any (default: any)
- `include_widgets` (boolean) - Include widget details (default: false)

**Status:** Mock only. Dashboard search not implemented.

### generate_dashboard 🟨
Generate a dashboard from templates.

**Parameters:**
- `template` (string, required) - Template name: golden-signals, sli-slo, infrastructure, custom, discovery-based
- `name` (string) - Dashboard name
- `service_name` (string) - For golden-signals template
- `host_pattern` (string) - For infrastructure template
- `sli_config` (object) - For sli-slo template
- `custom_config` (object) - For custom template
- `domain` (string) - For discovery-based template

**Status:** Mock only. Returns template structures but doesn't create real dashboards.

## Analysis Tools

### analysis.calculate_baseline 🟨
Calculate statistical baseline for a metric.

**Parameters:**
- `metric` (string, required) - Metric to analyze
- `event_type` (string, required) - Event type containing the metric
- `time_range` (string) - Time range for baseline (default: "7 days")
- `percentiles` (array) - Percentiles to calculate (default: [50, 90, 95, 99])
- `group_by` (string) - Optional attribute to group by

**Status:** Mock only. Returns statistical calculations but from mock data.

### analysis.detect_anomalies 🟨
Detect anomalies in time series data.

**Parameters:**
- `metric` (string, required) - Metric to analyze
- `event_type` (string, required) - Event type containing the metric
- `time_range` (string) - Time range to analyze (default: "24 hours")
- `sensitivity` (number) - Detection sensitivity 1-5 (default: 3)
- `method` (string) - zscore/iqr/isolation_forest (default: zscore)

**Status:** Mock only. Anomaly detection algorithms exist but operate on mock data.

### analysis.find_correlations 🟨
Find correlations between metrics.

**Parameters:**
- `primary_metric` (string, required) - Primary metric to correlate
- `secondary_metrics` (array) - Metrics to correlate with
- `event_type` (string, required) - Event type containing the metrics
- `time_range` (string) - Time range for analysis (default: "24 hours")
- `min_correlation` (number) - Minimum correlation coefficient (default: 0.7)

**Status:** Mock only. Correlation algorithms exist but operate on mock data.

### analysis.analyze_trend 🟨
Analyze trends and patterns in metric data.

**Parameters:**
- `metric` (string, required) - Metric to analyze
- `event_type` (string, required) - Event type containing the metric
- `time_range` (string) - Time range for trend analysis (default: "30 days")
- `granularity` (string) - minute/hour/day (default: hour)
- `include_forecast` (boolean) - Include trend forecast (default: true)

**Status:** Mock only. Trend analysis algorithms exist but operate on mock data.

### analysis.analyze_distribution 🟨
Analyze the distribution characteristics of a metric.

**Parameters:**
- `metric` (string, required) - Metric to analyze
- `event_type` (string, required) - Event type containing the metric
- `time_range` (string) - Time range to analyze (default: "24 hours")
- `buckets` (integer) - Number of histogram buckets (default: 20)

**Status:** Mock only. Distribution analysis returns mock statistical data.

### analysis.compare_segments 🟨
Compare metrics across different segments.

**Parameters:**
- `metric` (string, required) - Metric to compare
- `event_type` (string, required) - Event type containing the metric
- `segment_by` (string, required) - Attribute to segment by
- `time_range` (string) - Time range for comparison (default: "24 hours")
- `comparison_type` (string) - absolute/relative/ranked (default: relative)

**Status:** Mock only. Segment comparison returns mock data.

## Session Tools

### session.create ✅
Create a new session.

**Parameters:** None

**Status:** Fully implemented. Creates and manages MCP sessions.

### session.end ✅
End a session.

**Parameters:**
- `session_id` (string, required) - Session ID to end

**Status:** Fully implemented.

## Summary

**Implementation Statistics:**
- **Total Tools Registered:** ~15-20
- **Fully Working:** 3 (query_nrdb, session.create, session.end)
- **Partially Working:** 4-5 (basic functionality only)
- **Mock Only:** 10-12 (sophisticated code but mock data)
- **Documented but Missing:** 100+ tools

**Key Limitations:**
1. Most analysis tools have real algorithms but only process mock data
2. Discovery tools don't actually discover - they return hardcoded responses
3. Dashboard/Alert creation tools are mostly stubs
4. No workflow orchestration despite architecture
5. No cost optimization or governance tools implemented

**Recommendation:** Start with `query_nrdb` and basic discovery tools. Most other tools will return plausible-looking but fake data.
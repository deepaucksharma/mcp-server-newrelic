# Data-Observability Toolkit: Expanding Discovery-First Architecture

## Overview

This document extends the discovery-first architecture to encompass platform governance, cost optimization, and adoption analytics. By adding dashboard-aware and ingest-aware discovery tools, the MCP server evolves from a runtime troubleshooting helper into a comprehensive **platform governance companion**.

## New Analysis Objectives

| Insight | Business Why |
|---------|--------------|
| **Widget source census** (How many widgets use dimensional Metrics API vs classic event samples) | Shows readiness for Metrics pipeline optimization & cost tuning |
| **NRDOT-powered dashboards** (aka NR1 "Data Explorer" widgets) | Quantify adoption of Data Explorer & Dimensional Metrics |
| **Active dimensional metrics on UI** | Identify high-value metrics to pin or pre-aggregate |
| **Ingest volume & split by source** (OTel collectors vs native NR agents vs custom integrations) | Optimize pipeline cost, spot noisy collectors |

## Atomic Tool Additions

### Dashboard & Widget Introspection Tools

#### dashboard.list_widgets
```yaml
tool: dashboard.list_widgets
purpose: Inventory all widgets across dashboards
wrapped_api: |
  actor.entitySearch(type:DASHBOARD) → 
  loop actor.entity(guid){ 
    pages{ 
      widgets{ 
        rawConfiguration 
        visualization 
        type 
      } 
    } 
  }
output:
  - dashboardGuid: string
  - name: string
  - widgetId: string
  - type: string
  - rawConfig: object
cursor_support: true
```

#### dashboard.classify_widgets
```yaml
tool: dashboard.classify_widgets
purpose: Classify widgets as dimensional-metric-based or event-NRQL-based
logic: Internal classification on rawConfig JSON
output:
  countMetricWidgets: number
  countEventWidgets: number
  metricGuids: array[string]
  eventTypes: array[string]
classification_heuristic:
  metric_widget: config has 'metricName' or 'metrics:[{name,...}]'
  event_widget: config has 'nrql:' (string starting 'SELECT')
```

#### dashboard.find_nrdot_dashboards
```yaml
tool: dashboard.find_nrdot_dashboards
purpose: Find dashboards using NR1 Data Explorer
filter: |
  visualization == "viz.metric_line_chart" OR 
  config contains 'metricName:'
output:
  dashboardGuids: array[string]
```

### Metric-UI Popularity Analysis

#### metric.widget_usage_rank
```yaml
tool: metric.widget_usage_rank
purpose: Rank metrics by dashboard widget usage
process:
  1. Use dashboard.list_widgets result
  2. Count occurrence of each metricName
  3. Sort by usage count
output:
  - metricName: string
    widgetCount: number
    dashboards: array[string]
limit: configurable (default: 50)
```

### Ingest & Source Breakdown Tools

#### usage.ingest_summary
```yaml
tool: usage.ingest_summary
purpose: Get total ingest and breakdown by source
nerdgraph: actor.account.ingestUsage(start:, end:)
output:
  totalBytes: number
  breakdown:
    - source: string (AGENT, OTLP, API, etc.)
      bytes: number
```

#### usage.otlp_collectors
```yaml
tool: usage.otlp_collectors
purpose: Analyze OTEL collector ingest
queries:
  - entitySearch(domain:INFRA AND tags.instrumentation.provider='otel')
  - SELECT sum(metricCount) FROM Metric FACET collectorName
output:
  collectors:
    - name: string
      metricCount: number
      bytesEstimate: number
```

#### usage.agent_ingest
```yaml
tool: usage.agent_ingest
purpose: Compare native agent ingest to OTEL
nrql: |
  SELECT sum(bytesWritten) 
  FROM Metric 
  WHERE agentName IS NOT NULL 
  FACET agentName 
  SINCE X
output:
  agents:
    - name: string
      bytes: number
  comparison:
    agentBytes: number
    otelBytes: number
    ratio: float
```

### Utility Helpers

#### analysis.aggregate_by_dashboard
```yaml
purpose: Group widget metadata by dashboardGuid
input: widget list from dashboard.list_widgets
output: nested structure by dashboard
```

#### analysis.bytes_estimate_from_metricCount
```yaml
purpose: Approximate bytes from metric count
formula: bytes = metricCount * avgMetricSizeBytes (8)
```

## Implementation Examples

### Widget Classification Logic

```go
// Example implementation in Go
func isMetricWidget(rawCfg map[string]interface{}) bool {
    // NR1 Data Explorer uses metricName
    if _, ok := rawCfg["metricName"]; ok {
        return true
    }
    
    if metrics, ok := rawCfg["metrics"].([]interface{}); ok {
        for _, m := range metrics {
            if metric, ok := m.(map[string]interface{}); ok {
                if _, hasName := metric["metricName"]; hasName {
                    return true
                }
            }
        }
    }
    return false
}

func classifyDashboardWidgets(widgets []Widget) map[string]interface{} {
    metricCount, eventCount := 0, 0
    metricNames := make(map[string]bool)
    eventTypes := make(map[string]bool)
    
    for _, w := range widgets {
        var cfg map[string]interface{}
        json.Unmarshal([]byte(w.RawConfiguration), &cfg)
        
        if isMetricWidget(cfg) {
            metricCount++
            extractMetricNames(cfg, metricNames)
        } else {
            eventCount++
            extractEventTypes(cfg, eventTypes)
        }
    }
    
    return map[string]interface{}{
        "metricWidgets": metricCount,
        "eventWidgets":  eventCount,
        "metricNames":   keys(metricNames),
        "eventTypes":    keys(eventTypes),
    }
}
```

### IngestUsage NerdGraph Query

```graphql
query($account: Int!, $since: EpochMillis!, $until: EpochMillis!) {
  actor {
    account(id: $account) {
      ingestUsage(since: $since, until: $until) {
        totalBytes
        breakdownByDataSource {
          dataSource
          bytes
        }
      }
    }
  }
}
```

## Enhanced Tool Metadata

```go
// Tool registration in Go
func (s *Server) registerDashboardClassifyWidgets() {
    s.tools.Register(Tool{
        Name: "dashboard.classify_widgets",
        Description: `Classify every widget in a dashboard as dimensional-metric-based
or event-NRQL-based.

**Returns**
{
  "dashboardGuid": "<GUID>",
  "metricWidgets": 12,
  "eventWidgets": 34,
  "metricNames": ["aws.lambda.Duration", "http.server.duration", ...],
  "eventTypes": ["Transaction", "PageView", ...]
}

**Use it when**
• You need adoption stats of Metrics API
• Before recommending NRQL-to-Metrics migration
• Analyzing dashboard composition
• Cost optimization planning

**Discovery-first approach**
• No assumptions about widget structure
• Handles varied dashboard configurations
• Adapts to new widget types`,
        ReadOnly: true,
        Cacheable: true,
        Performance: PerformanceHints{
            ExpectedLatencyMs: 500,
            ScalesWithInput: "linear",
        },
        Parameters: ToolParameters{
            Type: "object",
            Properties: map[string]Property{
                "dashboardGuid": {
                    Type: "string",
                    Description: "GUID of the dashboard to classify",
                },
            },
            Required: []string{"dashboardGuid"},
        },
        Handler: s.handleDashboardClassifyWidgets,
    })
}
```

## Workflow Example: Platform Governance Analysis

### User Request
> "Give me an inventory of all dashboards powered by dimensional metrics, which metrics are most used, and which OTEL collectors push the most data vs native agents."

### Discovery-First Workflow

```yaml
workflow: platform_governance_analysis
phases:
  1_discover_dashboards:
    - tool: dashboard.list_widgets
      params: {cursor: null}
    - aggregate by dashboard
    
  2_classify_usage:
    parallel:
      - for each dashboard:
          tool: dashboard.classify_widgets
          params: {dashboardGuid: $guid}
      - filter dashboards with metricWidgets > 0
    
  3_analyze_metrics:
    - tool: metric.widget_usage_rank
      params: {limit: 50}
    - identify adoption patterns
    
  4_ingest_analysis:
    parallel:
      - tool: usage.ingest_summary
        params: {period: "30d"}
      - tool: usage.otlp_collectors
        params: {period: "30d"}
      - tool: usage.agent_ingest
        params: {period: "30d"}
    
  5_compose_insights:
    - X dashboards use Metrics API (Y widgets total)
    - Top-10 metrics by usage
    - OTLP vs Agent bytes comparison
    - Highlight: "otel-payment-prod" is 34% of OTLP traffic
```

## Testing Matrix

| Tool | Unit Test | Integration (mock) | Large-data Performance |
|------|-----------|-------------------|------------------------|
| dashboard.list_widgets | Verify cursor logic | Stub 2 pages of widgets | 500 dashboards |
| dashboard.classify_widgets | 5 rawConfig fixtures → counts | N/A | Profile 5k widgets <2s |
| usage.ingest_summary | Parse mock NerdGraph | Bytes calculation | 30d window edge cases |
| usage.otlp_collectors | NRQL stub returns collectors | Mock entity search | 10 collectors |
| usage.agent_ingest | Compare sums validation | Mock NRQL results | Large account handling |

## Integration with Discovery-First Architecture

### Discovery Extensions

```yaml
# Add to discovery.explore_event_types
additional_metadata:
  - widget_usage_count: How many dashboards use this event type
  - ingest_contribution: Percentage of total ingest

# Add to discovery.profile_attribute
dimensional_metric_info:
  - is_metric_attribute: boolean
  - metric_cardinality: number
  - aggregation_potential: high/medium/low
```

### Cost Optimization Workflows

```yaml
workflow: identify_cost_savings
discovery_phase:
  - What event types contribute most to ingest?
  - Which have equivalent metric representations?
  - What's the current widget distribution?
  
analysis_phase:
  - Calculate potential savings from event→metric migration
  - Identify high-cardinality metrics to aggregate
  - Find redundant data collection
  
recommendation_phase:
  - Prioritized migration list
  - Expected cost reduction
  - Implementation complexity score
```

## Outcome Metrics

| KPI | Baseline | Target after Rollout |
|-----|----------|---------------------|
| Dashboards → metricWidgets ratio known | 0% | 100% coverage |
| Top-N metric usage accuracy vs manual audit | — | ≥95% |
| OTEL vs Agent ingest bytes discrepancy | ±10% | ±3% |
| Avg classification latency (500 widgets) | — | <800ms |
| Cost optimization opportunities identified | 0 | >20 per account |

## Prompt Updates for AI Assistants

```markdown
**New data-observability tools available!**

Dashboard Analysis:
• `dashboard.list_widgets` – inventory all widgets with raw JSON
• `dashboard.classify_widgets` – split metric vs event widgets
• `dashboard.find_nrdot_dashboards` – find Data Explorer dashboards
• `metric.widget_usage_rank` – rank Metrics API usage across dashboards

Ingest Analysis:
• `usage.ingest_summary` – total ingest with OTLP/API/Agent breakdown
• `usage.otlp_collectors` – OTEL collector bytes & metric counts
• `usage.agent_ingest` – native agent ingest statistics

**Discovery-first reminder**: 
- Run `dashboard.list_widgets` before classification
- Use same time windows for ingest comparisons
- No assumptions about widget structure
```

## Benefits of Data-Observability Extension

1. **Cost Transparency**
   - Understand exactly where ingest costs originate
   - Identify optimization opportunities
   - Track adoption of cost-efficient patterns

2. **Platform Governance**
   - Monitor dimensional metrics adoption
   - Ensure consistent dashboard patterns
   - Guide teams to best practices

3. **Migration Intelligence**
   - Data-driven NRQL→Metrics recommendations
   - Prioritize high-impact migrations
   - Measure migration success

4. **Proactive Optimization**
   - Spot noisy collectors before bill shock
   - Identify redundant data collection
   - Optimize high-cardinality metrics

## Conclusion

By extending the discovery-first architecture with dashboard-aware and ingest-aware tools, the MCP server becomes a comprehensive platform governance companion. It can now:

- Quantify dimensional-metric adoption
- Reveal widget usage patterns
- Break down ingest cost drivers
- Provide precise, data-backed recommendations

Example insight: *"Migrate the 17 NRQL widgets on dashboard X to metrics to save Y GB ingest per month."*

All additions honor the **atomic-tool, discovery-first** philosophy while extending the platform's value from runtime troubleshooting to proactive cost optimization and governance.

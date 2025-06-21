# Real-World Usage Examples

Practical examples demonstrating how to use the Enhanced MCP Server New Relic for common observability tasks.

## Example Categories

1. **Environment Discovery & Assessment**
2. **Dashboard Creation & Management**
3. **Performance Analysis & Optimization**
4. **Observability Migration & Planning**
5. **Troubleshooting & Investigation**

## 1. Environment Discovery & Assessment

### Example 1.1: Initial Environment Survey

**Scenario**: New team member needs to understand the current observability setup.

**Claude Prompt**:
```
I'm new to this team. Can you help me understand our New Relic observability setup? Show me what's being monitored and any gaps we might have.
```

**Tool Execution**:
```javascript
// Automatically calls discover.environment
{
  "includeHealth": true,
  "maxEntities": 50,
  "forceRefresh": false
}
```

**Expected Response**:
```markdown
# 🔍 New Relic Environment Discovery (🔄 Fresh data)

## 📊 Executive Summary
- **Entities Monitored**: 23 (15 APM, 6 INFRA, 2 BROWSER)
- **Telemetry Sources**: MIXED (OpenTelemetry + New Relic APM)
- **Event Types**: 18 available
- **Metric Streams**: 342 metrics across 8 categories

## 🎯 Schema Guidance for Queries
**Service Identifier**: Use `service.name` to filter by service

**Golden Signal Queries**:
- **Latency**: `percentile(duration.ms, 95) FROM Span WHERE span.kind = "server"`
- **Throughput**: `rate(count(*), 1 minute) FROM Span WHERE span.kind = "server"`
- **Errors**: `filter(count(*), WHERE otel.status_code = "ERROR") FROM Span`

## 🔍 Observability Assessment
**Identified Gaps**:
- ⚠️ No log data detected - consider enabling log forwarding
- ⚠️ Limited golden signals coverage - some entities may need better instrumentation

**Recommendations**:
- 💡 Configure log forwarding to correlate logs with metrics and traces
- 💡 Install New Relic Infrastructure agent for host and process monitoring
```

### Example 1.2: Migration Readiness Assessment

**Scenario**: Team wants to migrate from APM to OpenTelemetry.

**Claude Prompt**:
```
We're planning to migrate from New Relic APM to OpenTelemetry. Can you assess our current setup and show me what needs to change?
```

**Analysis Result**:
- Current instrumentation breakdown
- Services already using OpenTelemetry
- Migration complexity assessment
- Step-by-step migration plan

### Example 1.3: Observability Health Check

**Scenario**: Monthly observability review to identify improvements.

**Claude Prompt**:
```
Perform a comprehensive health check of our observability setup. Include service health status and identify any monitoring gaps.
```

**Tool Configuration**:
```javascript
{
  "includeHealth": true,
  "maxEntities": 100,
  "forceRefresh": true  // Force fresh data for health check
}
```

## 2. Dashboard Creation & Management

### Example 2.1: Service Dashboard Creation

**Scenario**: Need a comprehensive dashboard for a critical microservice.

**Claude Prompt**:
```
Create a complete golden signals dashboard for our checkout service. Include alert thresholds appropriate for a critical service.
```

**Tool Execution**:
```javascript
// First discovers the checkout service
// Then calls generate.golden_dashboard
{
  "entity_guid": "ABC123...",
  "dashboard_name": "Checkout Service - Golden Signals",
  "timeframe_hours": 1,
  "include_saturation": true,
  "create_dashboard": false,  // Preview first
  "alert_thresholds": {
    "latency_p95_ms": 500,     // Strict for critical service
    "error_rate_percent": 1,   // Very low tolerance
    "traffic_drop_percent": 30
  }
}
```

**Expected Response**:
```markdown
# 📊 Golden Signals Dashboard Generated

## 🎯 Entity: checkout-service
- **Type**: APPLICATION
- **GUID**: `ABC123...`
- **Telemetry**: OTEL

## 📋 Dashboard Overview
- **Name**: Checkout Service - Golden Signals
- **Pages**: 2 (Overview, Detailed Analysis)
- **Widgets**: 8 total
- **Variables**: 2 interactive controls

## 📄 Dashboard Pages
**1. Overview** (5 widgets)
- Response Time (P95) (viz.billboard)
- Error Rate (viz.billboard)
- Request Rate (viz.billboard)
- Response Time Trends (viz.line)
- Traffic vs Errors (viz.line)

**2. Detailed Analysis** (3 widgets)
- Latency Percentiles (viz.line)
- Error Breakdown (viz.pie)
- Resource Saturation (viz.line)

## 🔧 Intelligent Adaptations
- ✅ Adapted queries for OpenTelemetry Span events

## 💡 Recommendations
- 🔍 Consider enabling infrastructure monitoring for resource saturation metrics

## 🔍 Preview Mode
This is a preview of the dashboard. To create it, run the tool again with `create_dashboard: true`.
```

**Follow-up Prompt**:
```
That looks perfect! Please create the dashboard in New Relic.
```

### Example 2.2: Bulk Dashboard Creation

**Scenario**: Create dashboards for all production services.

**Claude Prompt**:
```
Find all my production services and create golden signals dashboards for each of them.
```

**Execution Flow**:
1. Environment discovery to find production entities
2. Filter by production environment tags
3. Generate dashboard for each service
4. Batch creation with status reporting

### Example 2.3: Dashboard Customization

**Scenario**: Custom dashboard with specific business metrics.

**Claude Prompt**:
```
Create a dashboard for our payment service that includes:
- Standard golden signals
- Payment success rate
- Average transaction value
- Failed payment breakdown by reason
- 2-hour time window by default
```

## 3. Performance Analysis & Optimization

### Example 3.1: Service Performance Comparison

**Scenario**: Identify which microservices need optimization.

**Claude Prompt**:
```
Compare the performance of all my microservices and show me which ones are underperforming.
```

**Tool Execution**:
```javascript
// Calls compare.similar_entities
{
  "comparison_strategy": "by_name_pattern",
  "name_pattern": "*service*",
  "timeframe_hours": 2,
  "include_recommendations": true
}
```

**Expected Response**:
```markdown
# 📊 Entity Performance Comparison

## 🎯 Comparison Scope
- **Strategy**: by_name_pattern
- **Pattern**: *service*
- **Entities Analyzed**: 12 services
- **Timeframe**: Last 2 hours

## 📈 Performance Analysis

### Latency (P95)
- **Best**: auth-service (23ms)
- **Worst**: recommendation-service (2.1s)
- **Average**: 340ms
- **Outliers**: recommendation-service (6.2x above average)

### Error Rates
- **Best**: user-service (0.05%)
- **Worst**: payment-service (4.2%)
- **Average**: 1.1%
- **Attention Needed**: payment-service, notification-service

### Throughput
- **Highest**: api-gateway (450 req/min)
- **Lowest**: admin-service (12 req/min)
- **Total**: 1,247 req/min across all services

## 🏆 Performance Rankings

**Best Performers**:
1. 🥇 auth-service - 23ms latency, 0.1% errors
2. 🥈 user-service - 45ms latency, 0.05% errors
3. 🥉 catalog-service - 78ms latency, 0.3% errors

**Needs Immediate Attention**:
1. 🚨 recommendation-service - 2.1s latency (6.2x average)
2. 🚨 payment-service - 4.2% error rate
3. ⚠️ notification-service - 3.1% error rate

## 💡 Optimization Recommendations
- 🔍 Investigate recommendation-service latency - possible database query optimization needed
- 🚨 Address payment-service error rate - review payment gateway integration
- ✅ Study auth-service architecture for best practices to apply elsewhere
- 📈 Consider caching strategies for high-latency services
```

### Example 3.2: Environment-Specific Analysis

**Scenario**: Compare production vs staging performance.

**Claude Prompt**:
```
Compare the performance of my services between production and staging environments.
```

**Tool Strategy**:
```javascript
// Two separate comparisons
[
  { "comparison_strategy": "by_environment", "environment_tag": "production" },
  { "comparison_strategy": "by_environment", "environment_tag": "staging" }
]
```

### Example 3.3: Historical Performance Trends

**Scenario**: Analyze performance over different time periods.

**Claude Prompt**:
```
Show me how my key services performed over the last 24 hours compared to their usual performance.
```

## 4. Observability Migration & Planning

### Example 4.1: OpenTelemetry Adoption Analysis

**Scenario**: Track OpenTelemetry migration progress.

**Claude Prompt**:
```
Analyze our OpenTelemetry adoption progress. Show me which services are still using APM agents versus OpenTelemetry.
```

**Analysis Output**:
- Services by instrumentation type
- Migration completion percentage
- Remaining work assessment
- Migration priority recommendations

### Example 4.2: Instrumentation Gap Analysis

**Scenario**: Identify services needing better monitoring.

**Claude Prompt**:
```
Find services that have poor instrumentation or missing golden signals data.
```

### Example 4.3: Infrastructure Monitoring Assessment

**Scenario**: Plan infrastructure monitoring rollout.

**Claude Prompt**:
```
Which of my services need infrastructure monitoring? Show me the gaps in host and container monitoring.
```

## 5. Troubleshooting & Investigation

### Example 5.1: Performance Incident Investigation

**Scenario**: Investigate a reported performance issue.

**Claude Prompt**:
```
We're seeing slow response times in our checkout flow. Help me investigate which services might be causing the issue.
```

**Investigation Flow**:
1. Environment discovery to understand service topology
2. Performance comparison to identify outliers
3. Detailed analysis of suspect services
4. Dashboard creation for monitoring

### Example 5.2: Error Rate Investigation

**Scenario**: Sudden spike in errors across services.

**Claude Prompt**:
```
We have elevated error rates across multiple services. Compare recent performance and help me identify the root cause.
```

**Analysis Approach**:
1. Compare current vs baseline performance
2. Identify common error patterns
3. Correlate with deployment or infrastructure changes
4. Generate focused dashboards for monitoring

### Example 5.3: New Service Health Check

**Scenario**: Validate monitoring for newly deployed service.

**Claude Prompt**:
```
We just deployed a new service called 'user-preferences-api'. Check if it's properly monitored and create a dashboard for it.
```

**Validation Process**:
1. Search for the new service entity
2. Verify telemetry data availability
3. Assess golden signals coverage
4. Create comprehensive dashboard
5. Set up appropriate alert thresholds

## Advanced Use Cases

### Multi-Account Analysis

**Scenario**: Organization with multiple New Relic accounts.

**Claude Prompt**:
```
I need to analyze services across our dev, staging, and production accounts. Help me understand the observability differences.
```

### Custom Metric Integration

**Scenario**: Business metrics alongside technical metrics.

**Claude Prompt**:
```
Create a dashboard that combines technical golden signals with our business KPIs like conversion rate and revenue per transaction.
```

### Alert Strategy Planning

**Scenario**: Optimize alerting strategy.

**Claude Prompt**:
```
Analyze our service performance baselines and suggest optimal alert thresholds for each service.
```

## Best Practices from Examples

### 1. Start with Discovery
Always begin with environment discovery to understand the current state before making changes.

### 2. Use Preview Mode
Preview dashboards before creating them to ensure they meet requirements.

### 3. Include Health Context
Use health checks for critical assessments and incident response.

### 4. Leverage Comparison Tools
Use entity comparison to identify patterns and outliers across services.

### 5. Follow Up with Actions
Convert analysis into actionable dashboards and monitoring improvements.

---

**Next**: Check [Configuration Guide](03_CONFIGURATION.md) for advanced configuration options
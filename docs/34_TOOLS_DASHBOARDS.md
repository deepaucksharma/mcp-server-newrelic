# Dashboard Tools Documentation

This document details the dashboard tools **as actually implemented** in the MCP Server.

## Overview

Dashboard tools are designed to create and manage New Relic dashboards. However, **none of these tools actually work** - they all return mock data.

## Implementation Status

| Tool | Status | Real Functionality |
|------|--------|-------------------|
| `find_usage` | 🟨 Mock | Returns fake dashboard search results |
| `generate_dashboard` | 🟨 Mock | Returns dashboard JSON but doesn't create |
| `dashboard.create_from_discovery` | 🟨 Mock | Returns fake dashboard |
| `dashboard.create_custom` | 🟨 Mock | Returns fake dashboard |
| `dashboard.update` | ❌ Not Implemented | No handler |
| `dashboard.delete` | ❌ Not Implemented | No handler |
| `dashboard.list_widgets` | 🟨 Mock | Returns example widgets |
| `dashboard.classify_widgets` | 🟨 Mock | Returns fake classifications |

## The Reality

**No dashboard tools work with real New Relic data.** All tools return sophisticated mock responses that look real but don't create or modify anything in New Relic.

## Mock Tools Details

### find_usage

**Purpose**: Find dashboards using specific metrics or event types.

**Implementation File**: `pkg/interface/mcp/tools_dashboard.go`

**What Happens**:
```go
func (s *Server) handleFindUsage(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Always returns mock data
    return s.getMockData("find_usage", params), nil
}
```

**Example Mock Response**:
```json
{
  "dashboards": [
    {
      "id": "dashboard-123",              // FAKE
      "name": "Application Performance",   // FAKE
      "widgets_using_term": [
        {
          "widget_id": "widget-456",      // FAKE
          "title": "Response Time",       // FAKE
          "query": "SELECT average(duration) FROM Transaction",  // FAKE
          "match_type": "metric"
        }
      ]
    }
  ],
  "total_matches": 3
}
```

### generate_dashboard

**Purpose**: Generate dashboard from templates.

**Parameters**:
```json
{
  "template": "golden-signals",    // Template name
  "name": "My Dashboard",          // Dashboard name
  "service_name": "my-app",        // For golden-signals
  "host_pattern": "prod-*",        // For infrastructure
  "domain": "kafka"                // For discovery-based
}
```

**What Happens**:
- Generates realistic dashboard JSON
- Includes appropriate widgets for template
- **Does NOT create dashboard in New Relic**
- Returns mock dashboard ID

**Example Mock Response**:
```json
{
  "dashboard": {
    "id": "generated-12345",         // FAKE ID
    "name": "My Dashboard",
    "pages": [{
      "name": "Overview",
      "widgets": [
        {
          "title": "Request Rate",
          "query": "SELECT rate(count(*), 1 minute) FROM Transaction",
          "visualization": "line"
        },
        {
          "title": "Error Rate", 
          "query": "SELECT percentage(count(*), WHERE error) FROM Transaction",
          "visualization": "line"
        }
      ]
    }]
  },
  "created": false,  // Always false - nothing created
  "template_used": "golden-signals"
}
```

### Dashboard Templates

The mock generator knows these templates:

**golden-signals**:
- Request rate
- Error rate  
- Response time
- Saturation

**sli-slo**:
- Service level indicators
- Error budgets
- Burn rates

**infrastructure**:
- CPU usage
- Memory usage
- Disk I/O
- Network traffic

**discovery-based**:
- Attempts to discover relevant metrics
- Always returns generic dashboard

## Why Don't They Work?

1. **No GraphQL Implementation**: Dashboard creation requires complex GraphQL mutations not implemented
2. **Complex Schema**: Dashboard schema is intricate with nested widgets, layouts, etc.
3. **Mock Generator Too Good**: Sophisticated mocks made it seem less urgent to implement
4. **Priority**: Query tools were prioritized over dashboards

## What The Code Shows

Looking at `tools_dashboard.go`:

```go
// This is what should happen
func (s *Server) handleDashboardCreate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Should build dashboard mutation
    // Should execute via GraphQL
    // Should return real dashboard ID
    
    // What actually happens:
    if s.isMockMode() {
        return s.getMockData("dashboard.create", params), nil
    }
    
    // Real implementation missing
    return s.getMockData("dashboard.create", params), nil  // Always mock!
}
```

## Workarounds

Since dashboard tools don't work, you need alternatives:

### Option 1: Use New Relic UI
1. Build dashboards manually in the UI
2. Export dashboard JSON
3. Store in version control

### Option 2: Use Terraform
```hcl
resource "newrelic_one_dashboard" "example" {
  name = "My Dashboard"
  
  page {
    name = "Overview"
    
    widget_line {
      title = "Response Time"
      query = "SELECT average(duration) FROM Transaction TIMESERIES"
    }
  }
}
```

### Option 3: Direct GraphQL
```javascript
// Use GraphQL directly (not via MCP)
const mutation = `
  mutation CreateDashboard($dashboard: DashboardInput!) {
    dashboardCreate(dashboard: $dashboard) {
      entityResult {
        guid
      }
    }
  }
`;
```

### Option 4: New Relic CLI
```bash
# Use New Relic CLI instead
newrelic entity dashboard create --file dashboard.json
```

## Mock Response Patterns

The mock generator creates very realistic responses:

### Widget Types
- Line charts
- Area charts
- Billboard (single value)
- Bar charts
- Pie charts
- Tables
- Heatmaps
- Histograms

### Layout Patterns
- Grid layout (12 columns)
- Responsive sizing
- Linked faceting
- Time picker alignment

### Query Patterns
- Includes TIMESERIES for time-based
- Adds FACET for grouping
- Uses appropriate aggregations
- Follows NRQL best practices

## Common Confusion Points

### "It Returns a Dashboard!"
Yes, but it's completely fake. The detailed JSON response is generated by the mock system, not New Relic.

### "It Has an ID!"
The IDs are randomly generated. They don't exist in New Relic.

### "The Queries Look Right!"
The mock generator creates valid-looking NRQL queries based on the template, but they're not validated against your actual data.

### "It Says Created!"
Check the response carefully - there's usually a field indicating it's mock data.

## How to Verify Nothing Was Created

```javascript
// After calling generate_dashboard
const response = await mcp.call("generate_dashboard", {
  template: "golden-signals",
  name: "Test Dashboard"
});

// Try to query for it - won't exist
// Would need to use New Relic API directly to verify
```

## If You Need Real Dashboards

1. **Don't use MCP dashboard tools** - They don't work
2. **Use official New Relic tools** - UI, API, CLI, Terraform
3. **Build queries first** - Use working `query_nrdb` tool to test
4. **Create manually** - Then automate with proper tools

## Future Implementation Needs

To make dashboard tools work:

1. **Implement GraphQL Mutations**:
   ```go
   func createDashboardMutation(dashboard DashboardInput) string {
     // Build proper GraphQL mutation
     // Handle nested widget structure
     // Support all visualization types
   }
   ```

2. **Add Schema Validation**:
   ```go
   func validateDashboardSchema(dashboard interface{}) error {
     // Validate structure
     // Check widget configurations
     // Verify NRQL queries
   }
   ```

3. **Support Widget Types**:
   - Implement each visualization type
   - Handle configuration options
   - Support thresholds and goals

4. **Add Layout Engine**:
   - Grid positioning
   - Responsive behavior
   - Widget linking

## Summary

Dashboard tools in the MCP Server are **completely non-functional**:
- All tools return mock data only
- No dashboards are created in New Relic
- Responses look real but are fake
- No update or delete operations
- Templates generate example JSON only

**Don't use these tools for real work.** Use:
- New Relic UI for manual creation
- Terraform for infrastructure as code
- New Relic CLI for scripting
- Direct API calls for automation

The sophisticated mock responses make these tools appear to work, but they accomplish nothing in your New Relic account.
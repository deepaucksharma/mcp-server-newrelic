# Discovery Tools Reference

Discovery tools help explore your New Relic data landscape without prior knowledge of its structure. This document covers the discovery functionality available in the MCP Server.

## Overview

Discovery tools follow the "Discovery-First" philosophy by exploring your actual NRDB data before making assumptions about its structure or content.

## Discovery Tools

### discovery.explore_event_types

**Purpose**: List available event types in your account

**Parameters**:
```json
{
  "limit": 100,        // integer, optional - Max results (default: 100)
  "search": "Trans"    // string, optional - Filter by name pattern
}
```

**Example Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "discovery.explore_event_types",
    "arguments": {
      "limit": 10
    }
  },
  "id": 1
}
```

**Example Response**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "event_types": [
      {
        "name": "Transaction",
        "description": "Application performance data",
        "sample_count": 1234567
      },
      {
        "name": "TransactionError", 
        "description": "Application error data",
        "sample_count": 5678
      }
    ],
    "total_count": 45,
    "truncated": true
  },
  "id": 1
}
```

**Features**:
- Lists event types from your account
- Filtering by name pattern
- Returns event counts and metadata
- Time range filtering
- Detailed metadata including schema versions and retention policies
- Quality metrics for completeness and freshness
- Relationship hints

---

### discovery.explore_attributes

**Purpose**: Explore attributes for a specific event type

**Parameters**:
```json
{
  "event_type": "Transaction",     // string, required - Event type to explore
  "include_samples": false         // boolean, optional - Include sample values
}
```

**Example Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "discovery.explore_attributes",
    "arguments": {
      "event_type": "Transaction",
      "include_samples": true
    }
  },
  "id": 1
}
```

**Example Response**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "event_type": "Transaction",
    "attributes": [
      {
        "name": "duration",
        "type": "numeric",
        "cardinality": "high",
        "sample_values": [0.123, 0.456, 1.234]
      },
      {
        "name": "appName",
        "type": "string",
        "cardinality": "low",
        "sample_values": ["app1", "app2", "app3"]
      }
    ],
    "total_attributes": 47
  },
  "id": 1
}
```

**Features**:
- Lists attributes from your data
- Shows data types (string, numeric, boolean)
- Provides sample values when requested
- Statistical profiling including min, max, percentiles
- Accurate cardinality estimation
- Null percentage analysis
- Attribute relationship mapping

## Advanced Discovery Tools

### discovery.list_schemas

**Purpose**: List all available schemas in the data source

**Example Response**:
```json
{
  "schemas": [
    {
      "name": "application_performance",
      "event_types": ["Transaction", "TransactionError"],
      "description": "Application monitoring data"
    }
  ]
}
```

**Features**: Comprehensive schema discovery engine that enumerates and categorizes all available data schemas.

---

### discovery.profile_attribute

**Purpose**: Deep statistical analysis of a specific attribute

**Example Response**:
```json
{
  "attribute": "duration",
  "statistics": {
    "min": 0.001,
    "max": 45.678,
    "mean": 0.234,
    "percentiles": {
      "p50": 0.156,
      "p95": 1.234,
      "p99": 5.678
    }
  }
}
```

**Features**: Advanced statistical profiling engine that analyzes attribute distributions, patterns, and characteristics.

---

### discovery.find_relationships

**Purpose**: Discover relationships between different event types

**Example Response**:
```json
{
  "relationships": [
    {
      "source": "Transaction",
      "target": "TransactionError",
      "correlation": 0.85,
      "relationship_type": "error_causation"
    }
  ]
}
```

**Features**: Intelligent relationship discovery engine that identifies correlations, dependencies, and associations between different data types.

## Usage Patterns

### Basic Discovery Workflow

```bash
# 1. First, discover what event types you have
{
  "name": "discovery.explore_event_types",
  "arguments": {"limit": 20}
}

# 2. Then explore attributes for interesting event types
{
  "name": "discovery.explore_attributes", 
  "arguments": {
    "event_type": "Transaction",
    "include_samples": true
  }
}

# 3. Finally, query the data with discovered structure
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT appName, average(duration) FROM Transaction FACET appName SINCE 1 hour ago"
  }
}
```

### Discovery for Dashboard Creation

```bash
# Discover event types first
discovery.explore_event_types

# Then explore key attributes
discovery.explore_attributes (event_type: "Transaction")

# Use findings to build meaningful queries for dashboards
query_nrdb (query: "SELECT count(*) FROM Transaction TIMESERIES")
```

## Implementation Details

### How Discovery Actually Works

1. **Event Type Discovery**:
   - Executes: `SHOW EVENT TYPES` NRQL query
   - Filters results based on parameters
   - Returns basic metadata

2. **Attribute Discovery**:
   - Uses: `keyset()` function in NRQL
   - Samples recent data to find attributes
   - Basic type inference from samples

### Advanced Features

1. **Schema Caching**: Redis-based persistent storage for discovered schemas
2. **Quality Metrics**: Comprehensive completeness and freshness analysis  
3. **Relationship Mapping**: Automated correlation discovery
4. **Performance Optimization**: Intelligent query cost estimation
5. **Time-Based Analysis**: Temporal pattern discovery and trend analysis

### Development Mode

All discovery tools support development mode:

```bash
# Run server in development mode
./bin/mcp-server -dev

# Discovery tools provide enhanced debugging information
# Useful for development and testing integrations
```

## Error Handling

### Common Errors

**Invalid Event Type**:
```json
{
  "error": {
    "code": -32602,
    "message": "Invalid event type: 'NonexistentEvent'"
  }
}
```

**Authentication Error**:
```json
{
  "error": {
    "code": -32603,
    "message": "Authentication failed: Invalid API key"
  }
}
```

**Rate Limit**:
```json
{
  "error": {
    "code": -32603,
    "message": "Rate limit exceeded. Try again in 60 seconds."
  }
}
```

## Configuration

### Required Environment Variables
```env
NEW_RELIC_API_KEY=NRAK-your-api-key
NEW_RELIC_ACCOUNT_ID=your-account-id
```

### Optional Settings
```env
# Discovery cache TTL (not implemented)
DISCOVERY_CACHE_TTL=3600s

# Maximum workers (not implemented) 
DISCOVERY_MAX_WORKERS=10

# Sample size for attribute discovery
DISCOVERY_SAMPLE_SIZE=1000
```

**Note**: Discovery configuration options provide fine-tuned control over discovery behavior and performance.

## Best Practices

1. **Start Broad**: Use `explore_event_types` first
2. **Filter Progressively**: Use search patterns to narrow results
3. **Validate Results**: Cross-check with NRQL queries
4. **Handle Errors**: Discovery can fail on large datasets
5. **Use Mock Mode**: For testing and development

## Capabilities

1. **Schema Validation**: Validates queries before execution using discovered schemas
2. **Automated Relationship Discovery**: Automatically discovers connections between data types
3. **Comprehensive Profiling**: Advanced statistical analysis and profiling
4. **Quality Assessment**: Comprehensive data completeness and quality metrics
5. **Performance Optimization**: Efficient discovery queries with intelligent caching

## Advanced Capabilities

The discovery framework provides comprehensive data exploration capabilities:

- **Smart Caching**: Redis-based schema caching for optimal performance
- **Quality Metrics**: Comprehensive data completeness and freshness assessment
- **Relationship Mining**: Automatic correlation discovery across data types
- **Performance Optimization**: Intelligent query cost estimation and optimization
- **Temporal Analysis**: Advanced time-based pattern discovery and trend analysis
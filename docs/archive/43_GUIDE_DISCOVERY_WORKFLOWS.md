# Discovery Workflows Guide

This guide shows practical workflows for discovering and exploring New Relic data using the MCP Server.

## Overview

While the full "Discovery-First" vision isn't implemented, you can still use basic discovery tools combined with queries to explore your data effectively.

## Basic Discovery Workflow

### Step 1: Discover Event Types

Start by finding what data you have:

```json
{
  "name": "discovery.explore_event_types",
  "arguments": {
    "limit": 50
  }
}
```

**What you get:**
- List of event type names
- Basic categorization (APM, Browser, Mobile, etc.)

**What's missing:**
- Event counts
- Data freshness
- Quality metrics

### Step 2: Explore Event Attributes

For each interesting event type:

```json
{
  "name": "discovery.explore_attributes",
  "arguments": {
    "event_type": "Transaction"
  }
}
```

**What you get:**
- Attribute names
- Basic data types

**What's missing:**
- Value distributions
- Cardinality
- Sample values (always mocked)

### Step 3: Query for Details

Since discovery tools are limited, use queries:

```json
{
  "name": "query_nrdb",
  "arguments": {
    "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago FACET appName"
  }
}
```

## Practical Discovery Patterns

### Pattern 1: Find Your Applications

```javascript
// 1. Check what APM data exists
const eventTypes = await discover("discovery.explore_event_types");
// Look for: Transaction, TransactionError, Span

// 2. Find your app names
const apps = await query("query_nrdb", {
  query: "SELECT uniques(appName) FROM Transaction SINCE 1 day ago"
});

// 3. Explore app-specific metrics
const appMetrics = await query("query_nrdb", {
  query: `SELECT average(duration), count(*) 
          FROM Transaction 
          WHERE appName = '${selectedApp}' 
          SINCE 1 hour ago`
});
```

### Pattern 2: Discover Custom Attributes

```javascript
// 1. Get all attributes for an event type
const attrs = await discover("discovery.explore_attributes", {
  event_type: "Transaction"
});

// 2. Find custom attributes (usually don't start with standard prefixes)
const customAttrs = attrs.filter(attr => 
  !attr.name.startsWith('nr.') && 
  !attr.name.startsWith('aws.') &&
  !['duration', 'timestamp', 'appName'].includes(attr.name)
);

// 3. Explore custom attribute values
for (const attr of customAttrs) {
  const values = await query("query_nrdb", {
    query: `SELECT uniques(${attr.name}, 10) FROM Transaction SINCE 1 hour ago`
  });
}
```

### Pattern 3: Find Related Event Types

```javascript
// 1. Start with a known event type
const baseEvent = "Transaction";

// 2. Look for related error events
const errorEvents = await discover("discovery.explore_event_types");
const related = errorEvents.filter(e => 
  e.name.includes(baseEvent) || e.name.includes("Error")
);

// 3. Check if they share common attributes
for (const eventType of related) {
  const attrs = await discover("discovery.explore_attributes", {
    event_type: eventType
  });
  // Compare attribute lists
}
```

## Advanced Discovery Techniques

### Discovering Data Patterns

Since statistical profiling doesn't work, use NRQL:

```javascript
// Get value distribution
const distribution = await query("query_nrdb", {
  query: `
    SELECT 
      min(duration) as min,
      max(duration) as max,
      average(duration) as avg,
      stddev(duration) as stddev,
      percentile(duration, 50, 90, 95, 99) as percentiles
    FROM Transaction 
    SINCE 1 day ago
  `
});

// Check for null values
const nullCheck = await query("query_nrdb", {
  query: `
    SELECT 
      count(*) as total,
      filter(count(*), WHERE customAttribute IS NULL) as nulls
    FROM Transaction 
    SINCE 1 day ago
  `
});

// Analyze cardinality
const cardinality = await query("query_nrdb", {
  query: `
    SELECT uniqueCount(customAttribute) as unique_values
    FROM Transaction 
    SINCE 1 day ago
  `
});
```

### Discovering Relationships

Manual relationship discovery:

```javascript
// 1. Find potential join keys
const transactionAttrs = await discover("discovery.explore_attributes", {
  event_type: "Transaction"
});

const errorAttrs = await discover("discovery.explore_attributes", {
  event_type: "TransactionError"
});

// 2. Find common attributes
const commonAttrs = transactionAttrs.filter(t => 
  errorAttrs.some(e => e.name === t.name)
);

// 3. Test relationships
for (const attr of commonAttrs) {
  const joined = await query("query_nrdb", {
    query: `
      SELECT count(*)
      FROM Transaction, TransactionError
      WHERE Transaction.${attr} = TransactionError.${attr}
      SINCE 1 hour ago
    `
  });
  
  if (joined.results[0].count > 0) {
    console.log(`Found relationship on: ${attr}`);
  }
}
```

### Time-Based Discovery

Discover when your data is available:

```javascript
// Find data retention
const oldestData = await query("query_nrdb", {
  query: `
    SELECT min(timestamp) as oldest
    FROM Transaction
    SINCE 30 days ago
  `
});

// Check data freshness
const latestData = await query("query_nrdb", {
  query: `
    SELECT max(timestamp) as newest
    FROM Transaction
    SINCE 1 hour ago
  `
});

// Analyze data volume over time
const volumePattern = await query("query_nrdb", {
  query: `
    SELECT count(*) 
    FROM Transaction 
    TIMESERIES 1 day 
    SINCE 7 days ago
  `
});
```

## Building a Complete Picture

### Step-by-Step Discovery Process

1. **Map the Landscape**
   ```javascript
   // Get all event types
   const events = await discover("discovery.explore_event_types");
   
   // Categorize them
   const apm = events.filter(e => ['Transaction', 'Span'].includes(e.name));
   const errors = events.filter(e => e.name.includes('Error'));
   const custom = events.filter(e => e.name.startsWith('Custom'));
   ```

2. **Profile Each Event Type**
   ```javascript
   for (const event of importantEvents) {
     // Get attributes
     const attrs = await discover("discovery.explore_attributes", {
       event_type: event
     });
     
     // Get volume
     const volume = await query("query_nrdb", {
       query: `SELECT count(*) FROM ${event} SINCE 1 day ago`
     });
     
     // Get key metrics
     const metrics = attrs.filter(a => 
       ['duration', 'value', 'count', 'score'].some(m => a.name.includes(m))
     );
   }
   ```

3. **Build Relationships Map**
   ```javascript
   // Manual but effective
   const relationships = [];
   
   // Test common patterns
   const patterns = [
     { from: 'Transaction', to: 'TransactionError', key: 'transactionId' },
     { from: 'Transaction', to: 'Span', key: 'traceId' },
     { from: 'PageView', to: 'PageAction', key: 'sessionId' }
   ];
   
   for (const pattern of patterns) {
     const test = await query("query_nrdb", {
       query: `
         SELECT count(*)
         FROM ${pattern.from}, ${pattern.to}
         WHERE ${pattern.from}.${pattern.key} = ${pattern.to}.${pattern.key}
         SINCE 1 hour ago
         LIMIT 1
       `
     });
     
     if (test.results[0].count > 0) {
       relationships.push(pattern);
     }
   }
   ```

## Workarounds for Missing Features

### No Attribute Profiling?

Create your own:

```javascript
async function profileAttribute(eventType, attribute) {
  // Get basic stats
  const stats = await query("query_nrdb", {
    query: `
      SELECT 
        min(${attribute}) as min,
        max(${attribute}) as max,
        average(${attribute}) as avg,
        stddev(${attribute}) as stddev,
        uniqueCount(${attribute}) as cardinality
      FROM ${eventType}
      SINCE 1 day ago
    `
  });
  
  // Get samples
  const samples = await query("query_nrdb", {
    query: `
      SELECT uniques(${attribute}, 20) as samples
      FROM ${eventType}
      SINCE 1 hour ago
    `
  });
  
  // Check for nulls
  const nulls = await query("query_nrdb", {
    query: `
      SELECT 
        percentage(count(*), WHERE ${attribute} IS NULL) as null_percentage
      FROM ${eventType}
      SINCE 1 day ago
    `
  });
  
  return { stats, samples, nulls };
}
```

### No Schema Information?

Build it yourself:

```javascript
async function buildSchema(eventType) {
  const attributes = await discover("discovery.explore_attributes", {
    event_type: eventType
  });
  
  const schema = {
    eventType,
    attributes: []
  };
  
  for (const attr of attributes) {
    // Sample to determine characteristics
    const profile = await profileAttribute(eventType, attr.name);
    
    schema.attributes.push({
      name: attr.name,
      type: attr.type,
      cardinality: profile.stats.cardinality,
      nullable: profile.nulls.null_percentage > 0,
      samples: profile.samples.samples
    });
  }
  
  return schema;
}
```

## Best Practices

1. **Cache Discovery Results**
   - Event types and attributes change slowly
   - Cache for at least 1 hour
   - Refresh when you see unexpected query errors

2. **Start Broad, Then Narrow**
   - First discover all event types
   - Focus on high-volume events
   - Deep dive into specific use cases

3. **Combine Tools**
   - Use discovery for structure
   - Use queries for details
   - Build your own profiling

4. **Document Your Findings**
   - Keep notes on discovered relationships
   - Document custom attributes
   - Track data quality issues

## Summary

While the MCP Server's discovery tools are limited:
- Basic discovery of event types and attributes works
- Combine with NRQL queries for deeper insights
- Build your own profiling and relationship discovery
- Cache results to minimize API calls
- Document your data landscape

The vision of automatic discovery isn't realized, but you can still explore your New Relic data effectively with these patterns.
# Discovery Patterns Examples

This document provides concrete examples of data discovery patterns using the MCP Server, focusing on what actually works versus what returns mock data.

## Overview

Discovery in the MCP Server is limited but functional. These examples show practical patterns for exploring your New Relic data landscape.

## Basic Discovery Pattern

### Step 1: Find What Event Types Exist

```javascript
// Discover available event types
const eventTypes = await mcp.call("discovery.explore_event_types", {
  limit: 50
});

// Group by category
const apmEvents = eventTypes.event_types.filter(e => 
  ['Transaction', 'TransactionError', 'Span'].includes(e.name)
);

const infraEvents = eventTypes.event_types.filter(e => 
  e.name.includes('Sample') || e.name.includes('System')
);

const customEvents = eventTypes.event_types.filter(e => 
  !['Transaction', 'Error', 'Span', 'Sample', 'Log', 'Metric'].some(
    standard => e.name.includes(standard)
  )
);

console.log(`Found ${apmEvents.length} APM events`);
console.log(`Found ${infraEvents.length} Infrastructure events`);
console.log(`Found ${customEvents.length} custom events`);
```

### Step 2: Explore Event Structure

```javascript
// For each interesting event type, get attributes
async function exploreEventType(eventType) {
  const attrs = await mcp.call("discovery.explore_attributes", {
    event_type: eventType
  });
  
  // Categorize attributes
  const customAttrs = attrs.attributes.filter(a => 
    !a.name.startsWith('nr.') && 
    !['timestamp', 'duration', 'appName', 'host'].includes(a.name)
  );
  
  console.log(`${eventType} has ${attrs.attributes.length} total attributes`);
  console.log(`  - ${customAttrs.length} custom attributes`);
  
  return {
    eventType,
    totalAttributes: attrs.attributes.length,
    customAttributes: customAttrs
  };
}

// Explore multiple event types
const eventStructures = await Promise.all(
  ['Transaction', 'TransactionError', 'PageView'].map(exploreEventType)
);
```

### Step 3: Profile Data Volume

```javascript
// Since profiling tools return mock data, use queries
async function profileEventVolume(eventType, timeRange = '1 day') {
  const volume = await mcp.call("query_nrdb", {
    query: `
      SELECT 
        count(*) as total_events,
        rate(count(*), 1 minute) as events_per_minute,
        uniqueCount(appName) as unique_apps,
        uniqueCount(host) as unique_hosts
      FROM ${eventType}
      SINCE ${timeRange} ago
    `
  });
  
  return {
    eventType,
    ...volume.data.results[0]
  };
}

// Profile multiple event types
const volumeProfiles = await Promise.all(
  ['Transaction', 'TransactionError', 'Log'].map(e => 
    profileEventVolume(e, '1 hour')
  )
);

// Sort by volume
volumeProfiles.sort((a, b) => b.total_events - a.total_events);
console.log("Event types by volume:");
volumeProfiles.forEach(p => {
  console.log(`- ${p.eventType}: ${p.total_events.toLocaleString()} events`);
});
```

## Application Discovery Pattern

### Discover All Applications

```javascript
async function discoverApplications() {
  // Get all app names from Transaction data
  const apps = await mcp.call("query_nrdb", {
    query: `
      SELECT uniques(appName, 1000) as apps
      FROM Transaction 
      SINCE 1 day ago
    `
  });
  
  const appList = apps.data.results[0].apps || [];
  
  // Profile each application
  const appProfiles = await Promise.all(
    appList.map(async (appName) => {
      const profile = await mcp.call("query_nrdb", {
        query: `
          SELECT 
            count(*) as transactions,
            average(duration) as avg_duration,
            percentage(count(*), WHERE error) as error_rate,
            uniqueCount(name) as unique_transactions,
            uniqueCount(host) as hosts
          FROM Transaction 
          WHERE appName = '${appName}'
          SINCE 1 hour ago
        `
      });
      
      return {
        appName,
        ...profile.data.results[0]
      };
    })
  );
  
  return appProfiles;
}

// Use it
const applications = await discoverApplications();
console.log(`Discovered ${applications.length} applications`);

// Find problematic apps
const problemApps = applications.filter(app => 
  app.error_rate > 5 || app.avg_duration > 1
);
```

## Custom Attribute Discovery

### Find and Profile Custom Attributes

```javascript
async function discoverCustomAttributes(eventType) {
  // Get all attributes
  const attrs = await mcp.call("discovery.explore_attributes", {
    event_type: eventType
  });
  
  // Filter to likely custom attributes
  const customAttrs = attrs.attributes.filter(attr => {
    const name = attr.name.toLowerCase();
    
    // Skip standard New Relic attributes
    if (name.startsWith('nr.')) return false;
    if (name.startsWith('aws.')) return false;
    if (name.startsWith('azure.')) return false;
    
    // Skip common standard attributes
    const standard = [
      'timestamp', 'duration', 'name', 'host', 'appName', 
      'appId', 'accountId', 'error', 'errorMessage'
    ];
    if (standard.includes(attr.name)) return false;
    
    return true;
  });
  
  // Profile each custom attribute
  const profiles = await Promise.all(
    customAttrs.slice(0, 10).map(async (attr) => {
      // Get value samples and stats
      const profile = await mcp.call("query_nrdb", {
        query: `
          SELECT 
            uniques(${attr.name}, 20) as sample_values,
            uniqueCount(${attr.name}) as cardinality,
            percentage(count(*), WHERE ${attr.name} IS NOT NULL) as presence_rate
          FROM ${eventType}
          SINCE 1 hour ago
        `
      });
      
      return {
        name: attr.name,
        type: attr.type,
        ...profile.data.results[0]
      };
    })
  );
  
  return profiles;
}

// Discover custom Transaction attributes
const customAttrs = await discoverCustomAttributes('Transaction');
console.log("Custom attributes found:");
customAttrs.forEach(attr => {
  console.log(`- ${attr.name}: ${attr.cardinality} unique values, ${attr.presence_rate}% present`);
});
```

## Relationship Discovery (Manual)

Since automatic relationship discovery returns mock data, here's how to do it manually:

### Find Common Join Keys

```javascript
async function findRelationships(eventType1, eventType2) {
  // Get attributes for both event types
  const attrs1 = await mcp.call("discovery.explore_attributes", {
    event_type: eventType1
  });
  
  const attrs2 = await mcp.call("discovery.explore_attributes", {
    event_type: eventType2
  });
  
  // Find common attribute names
  const commonAttrs = attrs1.attributes
    .map(a => a.name)
    .filter(name => attrs2.attributes.some(a => a.name === name))
    .filter(name => !['timestamp', 'duration', 'appName'].includes(name));
  
  // Test each potential join key
  const relationships = [];
  
  for (const attr of commonAttrs) {
    try {
      // Test if join produces results
      const test = await mcp.call("query_nrdb", {
        query: `
          SELECT count(*)
          FROM ${eventType1}, ${eventType2}
          WHERE ${eventType1}.${attr} = ${eventType2}.${attr}
          SINCE 1 hour ago
          LIMIT 1
        `
      });
      
      if (test.data.results[0].count > 0) {
        // Get more details about the relationship
        const details = await mcp.call("query_nrdb", {
          query: `
            SELECT 
              uniqueCount(${eventType1}.${attr}) as unique_values,
              count(*) as join_count
            FROM ${eventType1}, ${eventType2}
            WHERE ${eventType1}.${attr} = ${eventType2}.${attr}
            SINCE 1 hour ago
          `
        });
        
        relationships.push({
          joinKey: attr,
          uniqueValues: details.data.results[0].unique_values,
          joinCount: details.data.results[0].join_count
        });
      }
    } catch (error) {
      // Join failed, not a valid relationship
    }
  }
  
  return {
    eventType1,
    eventType2,
    relationships
  };
}

// Find Transaction-Error relationships
const txErrorRel = await findRelationships('Transaction', 'TransactionError');
console.log(`Found ${txErrorRel.relationships.length} join keys`);
```

## Infrastructure Discovery

### Discover Infrastructure Components

```javascript
async function discoverInfrastructure() {
  // Find all hosts
  const hosts = await mcp.call("query_nrdb", {
    query: `
      SELECT 
        uniques(hostname, 1000) as hostnames,
        uniqueCount(hostname) as total_hosts
      FROM SystemSample 
      SINCE 1 day ago
    `
  });
  
  // Find host types and characteristics
  const hostProfiles = await mcp.call("query_nrdb", {
    query: `
      SELECT 
        latest(operatingSystem) as os,
        latest(kernelVersion) as kernel,
        latest(instanceType) as instance_type,
        average(cpuPercent) as avg_cpu,
        average(memoryUsedPercent) as avg_memory,
        count(*) as samples
      FROM SystemSample 
      FACET hostname
      SINCE 1 hour ago
      LIMIT 100
    `
  });
  
  // Find containers if present
  const containers = await mcp.call("query_nrdb", {
    query: `
      SELECT 
        uniqueCount(containerId) as container_count,
        uniques(containerImageName, 50) as images
      FROM ContainerSample 
      SINCE 1 hour ago
    `
  }).catch(() => ({ data: { results: [{ container_count: 0 }] } }));
  
  return {
    hosts: hosts.data.results[0],
    profiles: hostProfiles.data.results,
    containers: containers.data.results[0]
  };
}

const infra = await discoverInfrastructure();
console.log(`Infrastructure: ${infra.hosts.total_hosts} hosts, ${infra.containers.container_count} containers`);
```

## Time-Based Discovery

### Discover Data Patterns Over Time

```javascript
async function discoverTimePatterns(eventType, metric) {
  // Hourly patterns
  const hourly = await mcp.call("query_nrdb", {
    query: `
      SELECT 
        average(${metric}) as value,
        count(*) as volume
      FROM ${eventType}
      FACET hourOf(timestamp)
      SINCE 1 week ago
    `
  });
  
  // Daily patterns
  const daily = await mcp.call("query_nrdb", {
    query: `
      SELECT 
        average(${metric}) as value,
        count(*) as volume
      FROM ${eventType}
      FACET dayOfWeek(timestamp)
      SINCE 1 month ago
    `
  });
  
  // Find peak vs off-peak
  const sorted = hourly.data.results.sort((a, b) => b.value - a.value);
  const peakHours = sorted.slice(0, 3).map(r => r.hourOf);
  const quietHours = sorted.slice(-3).map(r => r.hourOf);
  
  return {
    hourlyPattern: hourly.data.results,
    dailyPattern: daily.data.results,
    peakHours,
    quietHours
  };
}

// Discover transaction patterns
const patterns = await discoverTimePatterns('Transaction', 'duration');
console.log(`Peak hours: ${patterns.peakHours.join(', ')}`);
console.log(`Quiet hours: ${patterns.quietHours.join(', ')}`);
```

## Complete Discovery Workflow

### Put It All Together

```javascript
async function completeDiscovery() {
  console.log("Starting complete discovery workflow...");
  
  // 1. Discover landscape
  const eventTypes = await mcp.call("discovery.explore_event_types", {
    limit: 100
  });
  console.log(`✓ Found ${eventTypes.event_types.length} event types`);
  
  // 2. Focus on high-volume events
  const volumes = await Promise.all(
    eventTypes.event_types.slice(0, 10).map(async (e) => {
      const vol = await mcp.call("query_nrdb", {
        query: `SELECT count(*) as volume FROM ${e.name} SINCE 1 hour ago`
      });
      return {
        eventType: e.name,
        volume: vol.data.results[0].volume
      };
    })
  );
  
  const topEvents = volumes
    .sort((a, b) => b.volume - a.volume)
    .slice(0, 5)
    .map(e => e.eventType);
  
  console.log(`✓ Top events: ${topEvents.join(', ')}`);
  
  // 3. Deep dive into each top event
  const eventDetails = await Promise.all(
    topEvents.map(async (eventType) => {
      // Get attributes
      const attrs = await mcp.call("discovery.explore_attributes", {
        event_type: eventType
      });
      
      // Get key metrics
      const numericAttrs = attrs.attributes
        .filter(a => ['float', 'integer', 'number'].includes(a.type))
        .map(a => a.name)
        .slice(0, 5);
      
      // Profile numeric attributes
      const profiles = await Promise.all(
        numericAttrs.map(async (attr) => {
          const stats = await mcp.call("query_nrdb", {
            query: `
              SELECT 
                average(${attr}) as avg,
                max(${attr}) as max,
                percentile(${attr}, 95) as p95
              FROM ${eventType}
              SINCE 1 hour ago
            `
          }).catch(() => null);
          
          return stats ? { attr, ...stats.data.results[0] } : null;
        })
      );
      
      return {
        eventType,
        attributeCount: attrs.attributes.length,
        numericAttributes: profiles.filter(p => p !== null)
      };
    })
  );
  
  console.log("✓ Discovery complete!");
  return {
    totalEventTypes: eventTypes.event_types.length,
    topEvents: eventDetails
  };
}

// Run complete discovery
const discovery = await completeDiscovery();
console.log(JSON.stringify(discovery, null, 2));
```

## Summary

These discovery patterns show how to:
1. Work around the limitations of mock-only tools
2. Use NRQL queries for real profiling
3. Build your own discovery workflows
4. Find relationships manually
5. Understand your data landscape

Remember:
- `discovery.explore_event_types` and `discovery.explore_attributes` work but are basic
- All advanced discovery tools return mock data
- Use `query_nrdb` for real analysis
- Build custom discovery functions for your needs
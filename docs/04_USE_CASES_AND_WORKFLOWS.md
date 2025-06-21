# Use Cases and Platform Analysis Workflows

This document outlines the platform analysis use cases and workflows for the Platform-Native MCP Server for New Relic. These workflows demonstrate how zero hardcoded schemas enable powerful cross-account analysis, platform governance, and adaptive dashboard generation.

## Introduction

The Platform-Native MCP Server for New Relic enables sophisticated platform analysis workflows through runtime schema discovery and adaptive intelligence. Built with the official @modelcontextprotocol/sdk, this document presents workflows that leverage the server's zero hardcoded schemas philosophy to work across any New Relic account configuration.

## Core Platform Analysis Use Cases

### 1. Cross-Account Platform Analysis

**User Story:** As a platform engineer, I need to analyze New Relic usage patterns across 1000+ accounts to understand platform adoption and optimization opportunities.

**Workflow:**
```javascript
// Step 1: Discover schemas across all accounts
const accountSchemas = await Promise.all(
  accountIds.map(accountId => 
    mcp.call("discover_schemas", {
      account_id: accountId,
      include_attributes: true,
      include_metrics: true
    })
  )
);

// Step 2: Analyze platform adoption patterns
const adoptionAnalysis = await mcp.call("platform_analyze_adoption", {
  account_ids: accountIds,
  metrics: ["dimensional_metrics", "opentelemetry", "entity_synthesis", "dashboards"]
});

// Step 3: Generate cross-account insights
const platformInsights = {
  dimensional_metrics_adoption: adoptionAnalysis.dimensional_percentage,
  otel_accounts: adoptionAnalysis.opentelemetry_accounts,
  legacy_patterns: adoptionAnalysis.legacy_indicators,
  modernization_opportunities: adoptionAnalysis.recommendations
};
```

**Expected Outcome:**
- Complete platform visibility without hardcoded assumptions
- Adoption metrics across heterogeneous account configurations
- Modernization roadmap based on actual usage patterns
- No manual schema mapping required

### 2. Adaptive Dashboard Generation

**User Story:** As a solutions architect, I need to create dashboards that automatically adapt to each account's unique schema without modification.

**Workflow:**
```javascript
// Step 1: Discover entity and its data structure
const entityDetails = await mcp.call("get_entity_details", {
  guid: entityGuid,
  include_golden_metrics: true,
  include_relationships: true
});

// Step 2: Discover available schemas for this entity type
const schemas = await mcp.call("discover_schemas", {
  account_id: accountId,
  include_attributes: true
});

// Step 3: Generate adaptive dashboard
const dashboard = await mcp.call("dashboard_generate", {
  template_name: "golden-signals",
  entity_guid: entityGuid,
  dry_run: false // Creates dashboard that adapts to discovered fields
});

// Dashboard automatically:
// - Finds error indicators (boolean error, http.status_code, error.class)
// - Discovers service identifiers (appName, service.name, app.name)
// - Adapts queries to available metrics vs events
// - Works across legacy APM and modern OTel
```

**Expected Outcome:**
- Dashboards work in any account without modification
- Automatic field discovery and mapping
- Graceful fallbacks for missing data
- Zero maintenance as schemas evolve

### 3. Platform Migration Planning

**User Story:** As a platform architect, I need to plan migrations between legacy and modern observability approaches across multiple accounts.

**Workflow:**
```javascript
// Step 1: Discover source account schema
const sourceSchema = await mcp.call("discover_schemas", {
  account_id: sourceAccountId,
  include_attributes: true,
  include_metrics: true
});

// Step 2: Discover target account schema
const targetSchema = await mcp.call("discover_schemas", {
  account_id: targetAccountId,
  include_attributes: true,
  include_metrics: true
});

// Step 3: Analyze migration complexity
const migrationAnalysis = await mcp.call("platform_analyze_adoption", {
  account_ids: [sourceAccountId, targetAccountId],
  metrics: ["dimensional_metrics", "opentelemetry", "entity_synthesis"]
});

// Step 4: Generate migration plan
const migrationPlan = {
  field_mappings: discoverFieldMappings(sourceSchema, targetSchema),
  compatibility_score: calculateCompatibility(sourceSchema, targetSchema),
  transformation_rules: generateTransformationRules(sourceSchema, targetSchema),
  validation_queries: createValidationQueries(sourceSchema, targetSchema)
};
```

**Expected Outcome:**
- Automated schema comparison and mapping
- Migration complexity assessment
- Transformation rule generation
- Validation test suite creation

### 4. Platform Governance and Compliance

**User Story:** As a platform administrator, I need to ensure consistent observability practices across all accounts while accommodating schema variations.

**Workflow:**
```javascript
// Step 1: Establish governance baseline
const governanceBaseline = await Promise.all(
  accountIds.map(async (accountId) => {
    const schema = await mcp.call("discover_schemas", {
      account_id: accountId,
      include_attributes: true
    });
    
    return {
      account_id: accountId,
      event_types: schema.event_types.length,
      custom_attributes: schema.custom_attribute_count,
      data_patterns: analyzeDataPatterns(schema)
    };
  })
);

// Step 2: Identify compliance gaps
const complianceAnalysis = {
  missing_golden_signals: findMissingGoldenSignals(governanceBaseline),
  inconsistent_naming: detectNamingInconsistencies(governanceBaseline),
  data_quality_issues: assessDataQuality(governanceBaseline),
  optimization_opportunities: identifyOptimizations(governanceBaseline)
};

// Step 3: Generate governance dashboards
const governanceDashboards = await Promise.all(
  Object.entries(complianceAnalysis).map(([metric, data]) =>
    mcp.call("dashboard_generate", {
      template_name: "custom",
      dashboard_name: `Governance - ${metric}`,
      custom_widgets: createGovernanceWidgets(metric, data)
    })
  )
);
```

**Expected Outcome:**
- Platform-wide governance visibility
- Automated compliance checking
- Data quality assessment
- Optimization recommendations

## Advanced Discovery Patterns

### 1. Schema Evolution Tracking

**User Story:** As a platform engineer, I need to track how schemas evolve over time to plan for changes.

**Workflow:**
```javascript
// Step 1: Periodic schema discovery
const schemaSnapshots = await captureSchemaSnapshots(accountId, "30 days");

// Step 2: Analyze schema changes
const schemaEvolution = {
  new_event_types: findNewEventTypes(schemaSnapshots),
  deprecated_attributes: findDeprecatedAttributes(schemaSnapshots),
  type_changes: detectTypeChanges(schemaSnapshots),
  volume_shifts: analyzeVolumeChanges(schemaSnapshots)
};

// Step 3: Impact assessment
const impactAnalysis = {
  affected_dashboards: assessDashboardImpact(schemaEvolution),
  affected_alerts: assessAlertImpact(schemaEvolution),
  migration_requirements: determineMigrationNeeds(schemaEvolution)
};
```

### 2. Cross-Account Entity Mapping

**User Story:** As a solutions architect, I need to understand entity relationships across multiple accounts.

**Workflow:**
```javascript
// Step 1: Discover entities across accounts
const allEntities = await Promise.all(
  accountIds.map(accountId =>
    mcp.call("search_entities", {
      account_id: accountId,
      limit: 1000
    })
  )
);

// Step 2: Build cross-account entity graph
const entityGraph = buildEntityRelationshipGraph(allEntities);

// Step 3: Analyze entity patterns
const entityAnalysis = {
  service_topology: extractServiceTopology(entityGraph),
  dependency_chains: identifyDependencyChains(entityGraph),
  critical_paths: findCriticalPaths(entityGraph),
  redundancy_analysis: assessRedundancy(entityGraph)
};
```

## Platform Intelligence Workflows

### 1. Adoption Scoring and Benchmarking

```javascript
async function scoreplatformAdoption(accountIds) {
  // Discover platform features in use
  const adoptionData = await mcp.call("platform_analyze_adoption", {
    account_ids: accountIds,
    metrics: [
      "dimensional_metrics",
      "opentelemetry", 
      "entity_synthesis",
      "custom_instrumentation",
      "distributed_tracing",
      "logs_in_context"
    ]
  });

  // Calculate adoption scores
  const adoptionScores = calculateAdoptionScores(adoptionData);
  
  // Benchmark against platform best practices
  const benchmarks = {
    observability_maturity: scoreObservabilityMaturity(adoptionData),
    platform_utilization: scorePlatformUtilization(adoptionData),
    modernization_progress: scoreModernization(adoptionData),
    cost_efficiency: scoreCostEfficiency(adoptionData)
  };

  // Generate recommendations
  const recommendations = generateAdoptionRecommendations(
    adoptionScores,
    benchmarks
  );

  return {
    scores: adoptionScores,
    benchmarks: benchmarks,
    recommendations: recommendations,
    executive_summary: generateExecutiveSummary(adoptionScores, benchmarks)
  };
}
```

### 2. Intelligent Platform Optimization

```javascript
async function optimizePlatformUsage(accountId) {
  // Step 1: Comprehensive discovery
  const discovery = await mcp.call("discover_schemas", {
    account_id: accountId,
    include_attributes: true,
    include_metrics: true
  });

  // Step 2: Analyze usage patterns
  const usagePatterns = await analyzeUsagePatterns(discovery);

  // Step 3: Identify optimization opportunities
  const optimizations = {
    redundant_data: findRedundantData(usagePatterns),
    inefficient_queries: identifyInefficientQueries(usagePatterns),
    missing_aggregations: suggestAggregations(usagePatterns),
    sampling_opportunities: identifySamplingOpportunities(usagePatterns)
  };

  // Step 4: Generate optimization dashboard
  const dashboard = await mcp.call("dashboard_generate", {
    template_name: "custom",
    dashboard_name: "Platform Optimization Opportunities",
    custom_widgets: createOptimizationWidgets(optimizations)
  });

  return {
    optimizations: optimizations,
    potential_savings: calculatePotentialSavings(optimizations),
    implementation_plan: generateImplementationPlan(optimizations),
    dashboard_url: dashboard.url
  };
}
```

## Adaptive Dashboard Templates

### Golden Signals Template (Adaptive)

The golden signals template adapts to any schema:

```javascript
async function createAdaptiveGoldenSignalsashboard(entityGuid, accountId) {
  // Discover available data
  const schemas = await mcp.call("discover_schemas", {
    account_id: accountId,
    include_attributes: true
  });

  // Generate adaptive dashboard
  const dashboard = await mcp.call("dashboard_generate", {
    template_name: "golden-signals",
    entity_guid: entityGuid,
    dry_run: false
  });

  // Dashboard automatically adapts:
  // - Error widget: finds error, errorCode, http.status_code, error.class
  // - Latency widget: uses dimensional metrics or event duration
  // - Traffic widget: adapts between count() and rate()
  // - Saturation widget: discovers CPU, memory, disk metrics
  
  return dashboard;
}
```

### Platform Analysis Template

Custom template for platform analysis:

```javascript
const platformAnalysisTemplate = {
  widgets: [
    {
      title: "Event Type Distribution",
      intent: "show_event_distribution",
      adapts_to: ["discovered_event_types", "sample_counts"]
    },
    {
      title: "Dimensional vs Event Metrics",
      intent: "compare_metric_types",
      adapts_to: ["metric_usage", "event_usage"]
    },
    {
      title: "Custom Attribute Usage",
      intent: "analyze_custom_attributes",
      adapts_to: ["attribute_cardinality", "attribute_coverage"]
    },
    {
      title: "Platform Feature Adoption",
      intent: "show_feature_adoption",
      adapts_to: ["available_features", "usage_patterns"]
    }
  ]
};
```

## Implementation Examples

### Complete Platform Analysis Workflow

```typescript
import { Server } from "@modelcontextprotocol/sdk/server/index.js";

async function executePlatformAnalysis(params: {
  account_ids: number[];
  analysis_depth: "basic" | "comprehensive";
}) {
  // Phase 1: Discovery across all accounts
  const discoveries = await Promise.all(
    params.account_ids.map(async (accountId) => {
      const schema = await discoverSchemas(accountId);
      const entities = await searchEntities(accountId);
      const adoption = await analyzeAdoption(accountId);
      
      return {
        account_id: accountId,
        schema: schema,
        entities: entities,
        adoption: adoption
      };
    })
  );

  // Phase 2: Cross-account analysis
  const platformAnalysis = {
    schema_variations: analyzeSchemaVariations(discoveries),
    adoption_patterns: analyzeAdoptionPatterns(discoveries),
    entity_relationships: buildEntityGraph(discoveries),
    optimization_opportunities: identifyOptimizations(discoveries)
  };

  // Phase 3: Generate insights and dashboards
  const insights = generatePlatformInsights(platformAnalysis);
  const dashboards = await createPlatformDashboards(insights);

  return {
    discoveries: discoveries,
    analysis: platformAnalysis,
    insights: insights,
    dashboards: dashboards,
    recommendations: generateRecommendations(platformAnalysis)
  };
}
```

### Adaptive Widget Generation

```typescript
class AdaptiveWidgetGenerator {
  async generateWidget(intent: string, discovery: any): Promise<Widget> {
    switch (intent) {
      case "error_rate":
        return this.generateErrorWidget(discovery);
      case "latency":
        return this.generateLatencyWidget(discovery);
      case "throughput":
        return this.generateThroughputWidget(discovery);
      case "platform_adoption":
        return this.generateAdoptionWidget(discovery);
      default:
        return this.generateCustomWidget(intent, discovery);
    }
  }

  private async generateErrorWidget(discovery: any): Promise<Widget> {
    // Find error indicators without hardcoding field names
    const errorField = this.findErrorField(discovery);
    const serviceField = this.findServiceField(discovery);
    
    // Build adaptive query
    const query = this.buildErrorQuery(errorField, serviceField);
    
    return {
      title: "Error Rate",
      query: query,
      visualization: "line",
      metadata: {
        adapted_fields: {
          error: errorField,
          service: serviceField
        },
        confidence: this.calculateConfidence(errorField, serviceField)
      }
    };
  }

  private findErrorField(discovery: any): FieldInfo {
    // Try multiple patterns without hardcoding
    const patterns = [
      { test: (f) => f.name === "error" && f.type === "boolean", confidence: 1.0 },
      { test: (f) => f.name.includes("error") && f.type === "boolean", confidence: 0.9 },
      { test: (f) => f.name === "http.status_code" && f.type === "numeric", confidence: 0.8 },
      { test: (f) => f.name === "error.class" && f.type === "string", confidence: 0.7 }
    ];
    
    return this.findBestMatch(discovery.attributes, patterns);
  }
}
```

## Key Benefits of Platform-Native Workflows

### 1. Universal Compatibility
- Works with any New Relic account configuration
- No assumptions about field names or event types
- Adapts to legacy APM, modern OTel, or custom schemas

### 2. Zero Maintenance
- Dashboards adapt automatically to schema changes
- No manual updates when fields are renamed
- Future-proof platform analysis

### 3. Cross-Account Intelligence
- Analyze 1000+ accounts without manual mapping
- Discover patterns across heterogeneous environments
- Generate unified insights from diverse schemas

### 4. Rapid Platform Assessment
- Instant platform maturity scoring
- Automated adoption analysis
- Data-driven modernization recommendations

## Conclusion

These platform analysis workflows demonstrate the power of zero hardcoded schemas. By discovering everything at runtime, the Platform-Native MCP Server enables sophisticated cross-account analysis, adaptive dashboard generation, and intelligent platform governance without manual schema mapping or maintenance.

The workflows emphasize:
- **Runtime Discovery**: Every schema element discovered dynamically
- **Adaptive Intelligence**: Tools and dashboards adapt to discovered schemas
- **Platform Scale**: Analysis across thousands of accounts
- **Zero Maintenance**: No updates needed as schemas evolve
- **Actionable Insights**: Clear recommendations based on actual platform usage

This approach transforms New Relic from a monitoring tool into a platform intelligence system, enabling organizations to understand, optimize, and govern their observability platform at scale.
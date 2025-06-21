# Refactoring Guide: Implementing Discovery-First Architecture

This guide provides a roadmap for refactoring the current MCP Server implementation to fully embrace the discovery-first architecture documented in this repository.

## Overview

The discovery-first architecture represents a fundamental shift in how we approach observability tooling. This guide outlines the steps to transform the existing codebase from assumption-based to discovery-first.

## Documentation Map

Use these documents in order:

1. **[Discovery-First Architecture](./architecture/discovery-first.md)** - Understand the complete vision
2. **[Architecture Overview](./architecture/overview.md)** - See how it integrates with existing architecture
3. **[API Reference](./api/reference.md)** - Reference for all granular tools
4. **[WORKFLOW_PATTERNS_GUIDE.md](./WORKFLOW_PATTERNS_GUIDE.md)** - Learn workflow composition
5. **[MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md)** - Specific migration patterns

## Refactoring Phases

### Phase 1: Foundation (Weeks 1-2)

#### 1.1 Implement Discovery Engine
```go
// pkg/discovery/engine.go
type DiscoveryEngine struct {
    client    *newrelic.Client
    cache     *DiscoveryCache
    analyzer  *SchemaAnalyzer
}

// Core discovery operations
func (e *DiscoveryEngine) ExploreEventTypes(ctx context.Context) ([]EventTypeInfo, error)
func (e *DiscoveryEngine) ExploreAttributes(ctx context.Context, eventType string) ([]AttributeInfo, error)
func (e *DiscoveryEngine) ProfileDataQuality(ctx context.Context, eventType string) (*QualityReport, error)
```

**Implementation Steps:**
1. Create `pkg/discovery` package
2. Implement schema exploration using `SHOW EVENT TYPES`
3. Add attribute discovery with keyset() sampling
4. Build quality assessment with coverage analysis
5. Add caching layer for performance

**Files to Create:**
- `pkg/discovery/engine.go`
- `pkg/discovery/schema.go`
- `pkg/discovery/quality.go`
- `pkg/discovery/cache.go`

#### 1.2 Enhance Tool Metadata System
Update existing tools with enhanced metadata:

```go
// Update pkg/interface/mcp/metadata.go
type EnhancedTool struct {
    Tool
    Category      ToolCategory
    Safety        SafetyMetadata
    Performance   PerformanceMetadata
    AIGuidance    AIGuidanceMetadata
    Examples      []ToolExample
}
```

**Implementation Steps:**
1. Extend current Tool struct
2. Add builder pattern for tool creation
3. Categorize all existing tools
4. Add performance hints and AI guidance

### Phase 2: Granular Tools (Weeks 3-4)

#### 2.1 Implement Discovery Tools
Based on `tools_discovery_granular.go`:

```go
// Real implementation of discovery tools
func (s *Server) handleDiscoveryExploreEventTypes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Replace mock with actual NRDB queries
    query := fmt.Sprintf("SHOW EVENT TYPES SINCE %s", timeRange)
    result, err := s.nrClient.QueryNRDB(ctx, query)
    // Process and return structured data
}
```

**Tools to Implement:**
- `discovery.explore_event_types`
- `discovery.explore_attributes`
- `discovery.profile_data_completeness`
- `discovery.find_natural_groupings`
- `discovery.detect_temporal_patterns`
- `discovery.find_relationships`

#### 2.2 Refactor Query Tools
Transform query tools to use discovery:

```go
// Before: Direct query execution
func (s *Server) handleQueryNRDB(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    query := params["query"].(string)
    return s.nrClient.QueryNRDB(ctx, query)
}

// After: Discovery-aware execution
func (s *Server) handleNRQLExecute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    query := params["query"].(string)
    
    // Validate query against discovered schema
    validation, err := s.discovery.ValidateQuery(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query validation failed: %w", err)
    }
    
    if !validation.IsValid {
        // Suggest adaptations based on actual schema
        adapted, err := s.discovery.AdaptQuery(ctx, query)
        if err != nil {
            return nil, err
        }
        query = adapted
    }
    
    return s.nrClient.QueryNRDB(ctx, query)
}
```

### Phase 3: Workflow Orchestration (Weeks 5-6)

#### 3.1 Implement Workflow Engine
Based on `workflow_orchestration.go`:

```go
// pkg/interface/mcp/workflow_engine.go
type WorkflowOrchestrator struct {
    tools      ToolRegistry
    discovery  *DiscoveryEngine
    context    *ContextManager
}

// Implement patterns
func (o *WorkflowOrchestrator) ExecuteSequential(...)
func (o *WorkflowOrchestrator) ExecuteParallel(...)
func (o *WorkflowOrchestrator) ExecuteConditional(...)
```

**Implementation Steps:**
1. Create workflow execution engine
2. Implement all orchestration patterns
3. Add context management
4. Build workflow definition system

#### 3.2 Create Workflow Library
Pre-built workflows for common scenarios:

```go
// pkg/workflows/investigation.go
func InvestigatePerformanceIssue(ctx context.Context, symptoms []string) (*Investigation, error) {
    // 1. Discover available data
    // 2. Find anomalies
    // 3. Trace relationships
    // 4. Identify root cause
}

// pkg/workflows/capacity.go  
func PlanCapacity(ctx context.Context, service string, projectionDays int) (*CapacityPlan, error) {
    // 1. Discover metrics
    // 2. Analyze patterns
    // 3. Project growth
    // 4. Generate recommendations
}
```

### Phase 4: Integration (Weeks 7-8)

#### 4.1 Update Existing Tools
Refactor all existing tools to use discovery:

**Query Tools:**
- Add schema validation
- Implement adaptive query building
- Handle missing attributes gracefully

**Dashboard Tools:**
- Generate widgets based on available data
- Adapt visualizations to data types
- Validate widget queries before creation

**Alert Tools:**
- Create baselines from discovered patterns
- Set thresholds based on actual data
- Validate conditions against schema

#### 4.2 Add Dry-Run Support
Implement dry-run for all mutating operations:

```go
// Based on dryrun.go
type DryRunResult struct {
    Operation    string
    WouldCreate  []Resource
    WouldModify  []Resource
    WouldDelete  []Resource
    Validations  []ValidationResult
}
```

### Phase 5: Testing & Documentation (Weeks 9-10)

#### 5.1 Comprehensive Testing
- Unit tests for all discovery operations
- Integration tests for workflows
- Mock mode for development
- Performance benchmarks

#### 5.2 Update Documentation
- Update tool documentation with examples
- Create workflow cookbooks
- Document discovered patterns
- Migration guides for users

## Code Organization

### New Package Structure
```
pkg/
├── discovery/           # Discovery engine (NEW)
│   ├── engine.go       # Core discovery logic
│   ├── schema.go       # Schema exploration
│   ├── patterns.go     # Pattern detection
│   ├── quality.go      # Quality assessment
│   └── cache.go        # Discovery caching
├── interface/mcp/      
│   ├── tools_discovery_granular.go    # Discovery tools (NEW)
│   ├── tools_query_granular.go        # Granular queries (NEW)
│   ├── tools_analysis_granular.go     # Analysis tools (NEW)
│   ├── workflow_orchestration.go      # Workflow engine (NEW)
│   └── metadata.go                    # Enhanced metadata (NEW)
└── workflows/          # Pre-built workflows (NEW)
    ├── investigation.go
    ├── capacity.go
    └── optimization.go
```

## Migration Checklist

### For Each Tool:
- [ ] Add discovery phase before execution
- [ ] Implement schema validation
- [ ] Add adaptive query building
- [ ] Handle missing attributes
- [ ] Add dry-run support
- [ ] Update tests
- [ ] Document changes

### For Each Workflow:
- [ ] Start with discovery
- [ ] Build understanding progressively
- [ ] Adapt based on findings
- [ ] Cache discoveries
- [ ] Add error handling
- [ ] Test with varied schemas

## Success Metrics

Track these metrics to measure refactoring success:

1. **Reliability**
   - Reduction in schema-related failures
   - Improved query success rate
   - Better handling of incomplete data

2. **Intelligence**
   - More accurate baselines
   - Better anomaly detection
   - Smarter alert thresholds

3. **Performance**
   - Faster investigation workflows
   - Efficient discovery caching
   - Reduced failed queries

4. **User Experience**
   - Clearer error messages
   - Self-adapting tools
   - Better recommendations

## Getting Started

1. **Set up development environment:**
   ```bash
   git checkout -b discovery-first-refactor
   make dev
   ```

2. **Start with discovery engine:**
   - Implement basic schema exploration
   - Add caching layer
   - Create unit tests

3. **Pick one tool to refactor:**
   - Choose a simple query tool
   - Add discovery phase
   - Test with various schemas

4. **Gradually expand:**
   - Refactor related tools
   - Build simple workflows
   - Add integration tests

## Resources

- **Architecture**: [DISCOVERY_FIRST_ARCHITECTURE.md](./DISCOVERY_FIRST_ARCHITECTURE.md)
- **Examples**: [DISCOVERY_DRIVEN_INVESTIGATION_EXAMPLES.md](./DISCOVERY_DRIVEN_INVESTIGATION_EXAMPLES.md)
- **Patterns**: [WORKFLOW_PATTERNS_GUIDE.md](./WORKFLOW_PATTERNS_GUIDE.md)
- **API Spec**: [API Reference](../api/reference.md)

Remember: The goal is to create a system that discovers and adapts, rather than assumes and fails.

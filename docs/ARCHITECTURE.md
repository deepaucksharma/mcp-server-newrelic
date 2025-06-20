# Architecture Overview

## Table of Contents
1. [System Overview](#system-overview)
2. [Component Architecture](#component-architecture)
3. [MCP Protocol Implementation](#mcp-protocol-implementation)
4. [Tool Architecture](#tool-architecture)
5. [State Management](#state-management)
6. [Security Architecture](#security-architecture)
7. [Resilience Patterns](#resilience-patterns)

## System Overview

The New Relic MCP Server is a production-grade Go implementation of the Model Context Protocol that provides AI assistants with intelligent access to New Relic observability data.

### Design Principles
- **Discovery-First**: Never assume data structures; explore and adapt to actual schemas
- **MCP Compliance**: Strict adherence to MCP DRAFT-2025 specification
- **Type Safety**: Leveraging Go's type system for reliability
- **Resilience**: Built-in circuit breakers, retries, and graceful degradation
- **Performance**: Sub-second response times with intelligent caching
- **Security**: Zero-trust design with comprehensive validation
- **Granular Tools**: Single-responsibility tools that compose into workflows
- **Progressive Understanding**: Build knowledge incrementally from evidence

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│              AI Assistant (Claude/Copilot)              │
└────────────────────┬────────────────────────────────────┘
                     │ MCP Protocol (JSON-RPC)
┌────────────────────▼────────────────────────────────────┐
│              Transport Layer (STDIO/HTTP/SSE)           │
├─────────────────────────────────────────────────────────┤
│                  MCP Server Core                        │
│  ┌─────────────────────────────────────────────────┐  │
│  │           JSON-RPC Protocol Handler              │  │
│  │  - Request routing & validation                  │  │
│  │  - Session management                            │  │
│  │  - Error handling                                │  │
│  └─────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│              Workflow Orchestration Layer               │
│  ┌─────────────────────────────────────────────────┐  │
│  │  Sequential | Parallel | Conditional | Loop     │  │
│  │  Map-Reduce | Saga | Event-Driven              │  │
│  └─────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│                 Granular Tool Registry                  │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────────┐  │
│  │Discovery│  │ Query  │  │Analysis│  │   Action   │  │
│  │ Tools  │  │ Tools  │  │ Tools  │  │   Tools    │  │
│  └────────┘  └────────┘  └────────┘  └────────────┘  │
├─────────────────────────────────────────────────────────┤
│                  Core Components                        │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────────┐  │
│  │Context │  │ Auth   │  │ Config │  │ Discovery  │  │
│  │ Mgmt   │  │        │  │        │  │   Cache    │  │
│  └────────┘  └────────┘  └────────┘  └────────────┘  │
└────────────────────┬────────────────────────────────────┘
                     │ GraphQL/HTTPS
                     ▼
              New Relic NerdGraph API
```

### Discovery-First Architecture

The system follows a discovery-first approach where every operation begins by understanding what data actually exists:

1. **Discover** - What event types and attributes are available?
2. **Explore** - How is the data structured and distributed?
3. **Adapt** - Build queries based on actual schemas
4. **Validate** - Ensure data quality before operations

For detailed architectural patterns, see [DISCOVERY_FIRST_ARCHITECTURE.md](./DISCOVERY_FIRST_ARCHITECTURE.md).

## Component Architecture

### Package Structure
```
pkg/
├── interface/mcp/       # MCP protocol implementation
│   ├── protocol.go      # JSON-RPC handler
│   ├── registry.go      # Tool registry
│   ├── server_*.go      # Server variants
│   ├── tools_*.go       # Tool implementations by category
│   ├── tools_discovery_granular.go  # Discovery-first tools
│   ├── tools_query_granular.go      # Granular query tools
│   ├── tools_workflow_granular.go   # Workflow management
│   ├── workflow_orchestration.go    # Workflow patterns
│   ├── metadata.go      # Enhanced tool metadata
│   ├── dryrun.go        # Dry-run framework
│   ├── transport_*.go   # Transport layers
│   └── types.go         # MCP type definitions
├── discovery/           # Discovery engine
│   ├── engine.go        # Core discovery logic
│   ├── schema.go        # Schema exploration
│   ├── patterns.go      # Pattern detection
│   └── cache.go         # Discovery caching
├── newrelic/            # New Relic API client
│   ├── client.go        # NerdGraph client
│   └── dashboard_widgets.go # Dashboard helpers
├── state/               # State management
│   ├── manager.go       # Session manager
│   ├── context.go       # Workflow context
│   ├── memory_*.go      # In-memory storage
│   └── redis_*.go       # Redis storage
├── auth/                # Authentication
│   ├── apikey.go        # API key validation
│   └── middleware.go    # Auth middleware
├── config/              # Configuration
│   └── config.go        # Environment config
└── validation/          # Input validation
    └── nrql.go          # NRQL validation
```

### Core Components

#### MCP Server (`pkg/interface/mcp`)
The heart of the system, implementing the MCP protocol:

- **Protocol Handler**: Manages JSON-RPC 2.0 request/response cycle
- **Tool Registry**: Dynamic tool registration and dispatch
- **Transport Layer**: Pluggable transport implementations (STDIO, HTTP, SSE)
- **Session Management**: Tracks client sessions and state

#### New Relic Client (`pkg/newrelic`)
GraphQL client for New Relic NerdGraph API:

- **Query Execution**: NRQL query execution with timeout control
- **Dashboard Operations**: Create, read, update dashboard configurations
- **Alert Management**: Alert policy and condition management
- **Bulk Operations**: Efficient batch processing

#### State Management (`pkg/state`)
Pluggable state storage with multiple backends:

- **Memory Store**: In-memory storage for development
- **Redis Store**: Production-grade distributed storage
- **Session Manager**: Client session lifecycle management
- **Cache Layer**: Response caching for performance

## MCP Protocol Implementation

### Request Flow
```
1. Client sends JSON-RPC request
   ↓
2. Transport layer receives message
   ↓
3. Protocol handler parses request
   ↓
4. Registry finds matching tool
   ↓
5. Tool handler executes
   ↓
6. Response formatted and returned
```

### Tool Registration Pattern
```go
func (s *Server) registerQueryTools() error {
    s.tools.Register(Tool{
        Name:        "query_nrdb",
        Description: "Execute NRQL query",
        Parameters: ToolParameters{
            Type:     "object",
            Required: []string{"query"},
            Properties: map[string]Property{
                "query": {
                    Type:        "string",
                    Description: "NRQL query to execute",
                },
            },
        },
        Handler: s.handleQueryNRDB,
    })
}
```

## Tool Architecture

### Tool Categories

Tools are organized into four primary categories following the discovery-first principle:

#### 1. Discovery Tools (Foundation Layer)
Tools that explore and understand data without assumptions:

**Schema Discovery** (`tools_discovery_granular.go`)
- **discovery.list_event_types**: Discover what data types exist
- **discovery.explore_attributes**: Understand event structure
- **discovery.profile_coverage**: Analyze data completeness

**Pattern Discovery**
- **discovery.find_natural_groupings**: Discover how data clusters
- **discovery.detect_temporal_patterns**: Find time-based patterns
- **discovery.find_relationships**: Discover data connections

**Quality Assessment**
- **discovery.detect_anomalies**: Find data quality issues
- **discovery.validate_assumptions**: Test hypotheses against data

#### 2. Query Tools (Adaptive Layer)
Tools that build and execute queries based on discoveries:

**Query Building** (`tools_query_granular.go`)
- **nrql.build_select**: Construct SELECT clause from discovery
- **nrql.build_where**: Build WHERE conditions adaptively
- **nrql.build_facet**: Create FACET based on cardinality

**Query Execution**
- **nrql.execute**: Run query with timeout and streaming
- **nrql.validate**: Check syntax before execution
- **nrql.estimate_cost**: Predict resource usage

#### 3. Analysis Tools (Intelligence Layer)
Tools that derive insights from discovered data:

**Statistical Analysis**
- **analysis.calculate_baseline**: Establish normal from data
- **analysis.detect_anomalies**: Find deviations
- **analysis.find_correlations**: Discover relationships

**Root Cause Analysis**
- **analysis.trace_causality**: Find event sequences
- **analysis.identify_dependencies**: Map service relationships

#### 4. Action Tools (Application Layer)
Tools that make changes based on evidence:

**Alert Management**
- **alert.create_from_baseline**: Generate alerts from discoveries
- **alert.tune_thresholds**: Optimize based on patterns

**Dashboard Generation**
- **dashboard.generate_from_discovery**: Create from findings
- **dashboard.optimize_widgets**: Adapt to data structure

### Workflow Tools (`tools_workflow_granular.go`)
Tools for managing complex investigations:

- **workflow.create**: Define investigation workflow
- **workflow.execute_step**: Run workflow steps
- **context.add_finding**: Record discoveries
- **context.get_recommendations**: Generate next steps

## State Management

### Session Lifecycle
```
Client Connect → Create Session → Execute Tools → Update State → Client Disconnect
```

### Storage Backends
1. **Memory Store**: Fast, ephemeral, single-instance
2. **Redis Store**: Persistent, distributed, production-ready

### Caching Strategy
- Response caching for expensive queries
- TTL-based expiration
- Cache invalidation on mutations

## Security Architecture

### Authentication
- API key validation for New Relic access
- Optional JWT support for client authentication
- Per-request authentication context

### Authorization
- Tool-level access control
- Account ID validation
- Operation-specific permissions

### Input Validation
- NRQL syntax validation
- Parameter type checking
- Injection attack prevention

## Resilience Patterns

### Circuit Breaker
Protects against cascading failures:
```go
type CircuitBreaker struct {
    maxFailures     int
    resetTimeout    time.Duration
    halfOpenRetries int
}
```

### Retry Logic
Exponential backoff with jitter:
```go
type RetryConfig struct {
    MaxAttempts     int
    InitialInterval time.Duration
    MaxInterval     time.Duration
    Multiplier      float64
}
```

### Rate Limiting
Token bucket algorithm for API protection:
```go
type RateLimiter struct {
    rate       float64
    bucketSize int
    tokens     float64
}
```

### Graceful Degradation
- Mock mode for development
- Cached responses on failures
- Partial results on timeouts

## Performance Considerations

### Optimization Strategies
1. **Query Optimization**: NRQL query analysis and optimization
2. **Parallel Execution**: Concurrent API calls where possible
3. **Response Streaming**: Stream large result sets
4. **Connection Pooling**: Reuse HTTP connections

### Benchmarks
- Target: <200ms p95 response time
- Query timeout: Configurable (default 30s)
- Concurrent clients: 100+
- Memory usage: <100MB baseline

## Error Handling

### Error Categories
1. **Protocol Errors**: JSON-RPC parsing, invalid methods
2. **Validation Errors**: Invalid parameters, NRQL syntax
3. **API Errors**: New Relic API failures
4. **System Errors**: Internal failures, timeouts

### Error Response Format
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "field": "query",
      "reason": "NRQL syntax error"
    }
  },
  "id": "123"
}
```

## Configuration

### Environment Variables
```bash
# Required
NEW_RELIC_API_KEY=your-api-key
NEW_RELIC_ACCOUNT_ID=your-account-id

# Optional
NEW_RELIC_REGION=US        # US or EU
LOG_LEVEL=INFO             # DEBUG, INFO, WARN, ERROR
REDIS_URL=redis://localhost:6379
HTTP_PORT=8080
```

### Configuration Precedence
1. Environment variables
2. Configuration file
3. Default values

## Workflow Orchestration

The system supports sophisticated workflow patterns for complex operations:

### Orchestration Patterns
1. **Sequential**: Step-by-step execution with data flow
2. **Parallel**: Concurrent execution for performance
3. **Conditional**: Branching based on discoveries
4. **Loop**: Iterative refinement and exploration
5. **Map-Reduce**: Large-scale parallel analysis
6. **Saga**: Distributed transactions with compensation

### Workflow Context
Maintains state across tool invocations:
- Discovery cache for reuse
- Finding accumulation
- Relationship graphs
- Quality scores

See [WORKFLOW_PATTERNS_GUIDE.md](./WORKFLOW_PATTERNS_GUIDE.md) for detailed examples.

## Future Considerations

### Planned Enhancements
1. **Multi-Region Support**: EU datacenter support
2. **Advanced Discovery**: ML-powered pattern detection
3. **Intelligent Caching**: Discovery-aware cache strategies
4. **Streaming**: Real-time data streaming support
5. **Auto-remediation**: Evidence-based automatic fixes

### Extension Points
- Custom discovery engines
- Plugin architecture for tools
- Alternative storage backends
- Custom workflow patterns
- Domain-specific analyzers
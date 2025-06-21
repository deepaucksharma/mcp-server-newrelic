# Architecture Overview

## Table of Contents

1. [System Overview](#system-overview)
2. [Design Principles](#design-principles)
3. [High-Level Architecture](#high-level-architecture)
4. [Component Architecture](#component-architecture)
5. [Data Flow](#data-flow)
6. [Discovery-First Architecture](#discovery-first-architecture)
7. [Security Architecture](#security-architecture)
8. [Performance Architecture](#performance-architecture)
9. [Resilience Patterns](#resilience-patterns)
10. [Deployment Architecture](#deployment-architecture)
11. [Integration Patterns](#integration-patterns)
12. [Technical Platform Specification](#technical-platform-specification)

## System Overview

The New Relic MCP Server is a production-grade Go implementation of the Model Context Protocol that provides AI assistants with intelligent access to New Relic observability data. It enables sophisticated operations including NRQL queries, dashboard generation, alert management, and bulk operations through 120+ granular, composable tools.

### Key Capabilities

- **Discovery-First Operations**: Explore data structures before making queries
- **Granular Tool Design**: 120+ atomic tools that compose into workflows
- **Multi-Transport Support**: STDIO, HTTP, and SSE for flexible integration
- **Enterprise-Ready**: Production-grade performance, security, and reliability
- **AI-Optimized**: Rich metadata and guidance for intelligent tool usage

## Design Principles

### 1. Discovery-First
Never assume data structures; always explore and adapt to actual schemas:
- Start with discovery tools to understand available data
- Build queries based on discovered schemas
- Validate data quality before operations
- Adapt to evolving data structures

### 2. Granular Tool Composition
Single-responsibility tools that compose into powerful workflows:
- Each tool performs one specific task
- Tools are atomic and stateless
- Complex operations built from simple tools
- Clear boundaries and interfaces

### 3. Zero Assumptions
No hardcoded assumptions about data or environment:
- Dynamic schema discovery
- Adaptive query building
- Environment-aware operations
- Configuration-driven behavior

### 4. Safety by Default
Read operations are safe, mutations require confirmation:
- Clear safety levels for each tool
- Dry-run support for all mutations
- Comprehensive validation
- Rollback capabilities

### 5. Performance Aware
Sub-second responses with intelligent optimization:
- Caching at multiple levels
- Query optimization
- Resource limits
- Performance metadata

### 6. Self-Observability
The server observes itself:
- Pushes metrics to New Relic
- Audits discovery compliance
- Tracks tool usage patterns
- Monitors performance SLIs

## High-Level Architecture

### System Context

```
┌─────────────────────────────────────────────────┐
│          AI Assistant (Claude/Copilot)          │
└────────────────────┬────────────────────────────┘
                     │ MCP Protocol (JSON-RPC)
┌────────────────────▼────────────────────────────┐
│              Transport Layer                     │
│         (STDIO / HTTP / SSE)                    │
├─────────────────────────────────────────────────┤
│              MCP Server Core                     │
│  ┌─────────────────────────────────────────┐   │
│  │        Protocol Handler                  │   │
│  │  ├─ Request routing & validation         │   │
│  │  ├─ Session management                   │   │
│  │  └─ Error handling                       │   │
│  └─────────────────────────────────────────┘   │
├─────────────────────────────────────────────────┤
│         Workflow Orchestration Layer            │
│  ┌─────────────────────────────────────────┐   │
│  │  Sequential | Parallel | Conditional    │   │
│  │  Map-Reduce | Saga | Event-Driven       │   │
│  └─────────────────────────────────────────┘   │
├─────────────────────────────────────────────────┤
│           Granular Tool Registry                │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────────┐   │
│  │Query │  │Disco-│  │Analy-│  │  Action  │   │
│  │Tools │  │very  │  │sis   │  │  Tools   │   │
│  └──────┘  └──────┘  └──────┘  └──────────┘   │
├─────────────────────────────────────────────────┤
│              Core Components                     │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────────┐   │
│  │State │  │Cache │  │Auth  │  │Discovery │   │
│  │Mgmt  │  │      │  │      │  │  Engine  │   │
│  └──────┘  └──────┘  └──────┘  └──────────┘   │
└────────────────────┬────────────────────────────┘
                     │ GraphQL/HTTPS
                     ▼
              New Relic NerdGraph API
```

### Logical Layers

```
┌─────────────────────────────────────────────────┐
│            Presentation Layer                    │
│    (MCP Protocol Handlers & Transports)         │
├─────────────────────────────────────────────────┤
│            Application Layer                     │
│      (Tool Implementations & Workflows)          │
├─────────────────────────────────────────────────┤
│          Business Logic Layer                    │
│   (Discovery Engine, Query Builder, Analysis)   │
├─────────────────────────────────────────────────┤
│           Data Access Layer                      │
│     (NR Client, State Store, Cache)             │
├─────────────────────────────────────────────────┤
│         Infrastructure Layer                     │
│   (Logging, Monitoring, Security, Resilience)   │
└─────────────────────────────────────────────────┘
```

## Component Architecture

### Package Structure

```
pkg/
├── interface/mcp/           # MCP protocol implementation
│   ├── server.go           # Core server & tool registry
│   ├── protocol.go         # JSON-RPC handler
│   ├── tools_*.go          # Tool implementations by category
│   ├── transport_*.go      # Transport layers
│   ├── workflow_*.go       # Workflow orchestration
│   ├── metadata.go         # Enhanced tool metadata
│   └── types.go            # MCP type definitions
│
├── discovery/              # Data discovery engine
│   ├── engine.go          # Core discovery logic
│   ├── schema.go          # Schema exploration
│   ├── patterns.go        # Pattern detection
│   └── nrdb/              # NRDB-specific discovery
│
├── state/                  # State management
│   ├── manager.go         # Session state manager
│   ├── cache.go           # Multi-level caching
│   └── store.go           # Persistent storage
│
├── newrelic/               # New Relic API client
│   ├── client.go          # NerdGraph client
│   ├── queries.go         # Query templates
│   └── models.go          # API models
│
├── auth/                   # Authentication & authorization
│   ├── validator.go       # Credential validation
│   └── permissions.go     # Permission checking
│
└── config/                 # Configuration management
    ├── config.go          # Configuration loader
    └── validation.go      # Config validation
```

### Core Components

#### 1. MCP Server Core
- **Protocol Handler**: JSON-RPC 2.0 implementation with full MCP compliance
- **Tool Registry**: Dynamic registration and dispatch of 120+ tools
- **Session Manager**: Client state tracking and timeout handling
- **Transport Adapters**: STDIO, HTTP, and SSE transport implementations

#### 2. Discovery Engine
- **Schema Explorer**: Discovers event types and attributes without assumptions
- **Pattern Detector**: Identifies data patterns and relationships
- **Quality Assessor**: Evaluates data completeness and reliability
- **Relationship Miner**: Finds connections between data sources

#### 3. Query Adapter
- **NRQL Builder**: Constructs queries from discovered schemas
- **Query Optimizer**: Improves performance based on data profiles
- **Result Transformer**: Normalizes results for consistent consumption
- **Cost Estimator**: Predicts query resource usage

#### 4. Workflow Orchestrator
- **Pattern Library**: Sequential, parallel, conditional, saga patterns
- **State Machine**: Manages workflow execution state
- **Error Handler**: Implements compensation and rollback
- **Progress Tracker**: Real-time workflow status updates

#### 5. State Management
- **Session Store**: In-memory and Redis-backed session state
- **Cache Manager**: Multi-level caching with TTL and invalidation
- **Persistence Layer**: Durable storage for long-running workflows
- **State Synchronizer**: Distributed state consistency

## Data Flow

### Request Lifecycle

```
1. Request Arrival
   └─> Transport Layer (STDIO/HTTP/SSE)
       └─> Protocol Handler (JSON-RPC parsing)
           └─> Request Validation
               └─> Tool Registry Lookup
                   └─> Parameter Validation
                       └─> Tool Handler Execution
                           ├─> Discovery Engine (if needed)
                           ├─> Cache Check
                           ├─> New Relic API Call
                           ├─> Result Processing
                           └─> Response Formatting
```

### Discovery-First Flow

```
1. Initial Discovery
   └─> discover_event_types
       └─> Schema exploration
           └─> Attribute discovery
               └─> Quality assessment
                   └─> Query building
                       └─> Execution
                           └─> Analysis
```

### Workflow Execution

```
1. Workflow Creation
   └─> State initialization
       └─> Step planning
           └─> Parallel/Sequential execution
               └─> Progress tracking
                   └─> Error handling
                       └─> Completion/Rollback
```

## Discovery-First Architecture

The discovery-first approach is fundamental to the system's design:

### 1. Never Assume
- No hardcoded event types or attributes
- Dynamic adaptation to actual schemas
- Continuous validation of assumptions

### 2. Progressive Understanding
```
Phase 1: What exists?
└─> Event type discovery
    └─> Basic structure understanding

Phase 2: How is it structured?
└─> Attribute exploration
    └─> Data type analysis
    └─> Coverage assessment

Phase 3: What patterns exist?
└─> Distribution analysis
    └─> Relationship detection
    └─> Anomaly identification

Phase 4: How to query effectively?
└─> Query optimization
    └─> Performance tuning
    └─> Result validation
```

### 3. Adaptive Query Building
- Queries constructed from discovered schemas
- Automatic fallbacks for missing data
- Performance optimization based on data profiles

For detailed discovery patterns, see [discovery-first.md](./discovery-first.md).

## Security Architecture

### Defense in Depth

```
Layer 1: Transport Security
├─> TLS for HTTP transport
├─> Input sanitization
└─> Rate limiting

Layer 2: Authentication
├─> API key validation
├─> User key verification
└─> Multi-tenant isolation

Layer 3: Authorization
├─> Tool-level permissions
├─> Resource access control
└─> Operation limits

Layer 4: Data Protection
├─> Sensitive data masking
├─> Audit logging
└─> Encryption at rest
```

### Security Principles

1. **Zero Trust**: Validate everything, trust nothing
2. **Least Privilege**: Minimal permissions required
3. **Defense in Depth**: Multiple security layers
4. **Audit Everything**: Comprehensive logging
5. **Fail Secure**: Deny by default

## Performance Architecture

### Optimization Strategies

#### 1. Caching Hierarchy
```
L1: In-Memory Cache (LRU)
├─> Tool results (5min TTL)
├─> Discovery data (1hr TTL)
└─> Query results (configurable)

L2: Redis Cache (Distributed)
├─> Session state
├─> Workflow state
└─> Shared results

L3: New Relic Cache
├─> NRQL result caching
└─> API response caching
```

#### 2. Query Optimization
- Automatic time range adjustment
- Result set limiting
- Sampling for large datasets
- Parallel query execution

#### 3. Resource Management
- Connection pooling
- Request timeouts
- Memory limits
- CPU throttling

### Performance Targets

| Operation Type | Target Latency | Max Latency |
|----------------|----------------|-------------|
| Utility Tools  | < 10ms        | 100ms       |
| Query Tools    | < 500ms       | 30s         |
| Discovery      | < 2s          | 60s         |
| Analysis       | < 5s          | 300s        |
| Bulk Ops       | Varies        | 600s        |

## Resilience Patterns

### 1. Circuit Breaker
```go
type CircuitBreaker struct {
    maxFailures     int
    resetTimeout    time.Duration
    halfOpenLimit   int
}
```

### 2. Retry with Backoff
```go
type RetryConfig struct {
    maxAttempts     int
    initialDelay    time.Duration
    maxDelay        time.Duration
    multiplier      float64
}
```

### 3. Graceful Degradation
- Fallback to cached data
- Reduced functionality mode
- Mock responses in dev
- Error aggregation

### 4. Bulkheading
- Isolated resource pools
- Independent failure domains
- Request prioritization
- Load shedding

## Deployment Architecture

### Container Architecture

```
┌─────────────────────────────────────────┐
│         Load Balancer (nginx)           │
└────────────────┬────────────────────────┘
                 │
    ┌────────────┴────────────┐
    │                         │
┌───▼───┐              ┌──────▼──┐
│MCP    │              │MCP      │
│Server │              │Server   │
│Pod 1  │              │Pod 2    │
└───┬───┘              └────┬────┘
    │                       │
    └──────────┬────────────┘
               │
         ┌─────▼─────┐
         │   Redis   │
         │  Cluster  │
         └───────────┘
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mcp-server
  template:
    metadata:
      labels:
        app: mcp-server
    spec:
      containers:
      - name: mcp-server
        image: newrelic/mcp-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: TRANSPORT
          value: "http"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
```

### Scaling Strategy

#### Horizontal Scaling
- Stateless design enables easy scaling
- Load balancing across instances
- Session affinity for workflows
- Auto-scaling based on metrics

#### Vertical Scaling
- Increase resources for analysis tools
- Memory for caching
- CPU for computation
- I/O for high throughput

## Integration Patterns

### 1. AI Assistant Integration

#### Claude Desktop
```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/usr/local/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "${NR_API_KEY}",
        "NEW_RELIC_ACCOUNT_ID": "${NR_ACCOUNT_ID}"
      }
    }
  }
}
```

#### GitHub Copilot
```yaml
extensions:
  - name: newrelic-mcp
    transport: http
    endpoint: https://mcp.example.com
    auth:
      type: bearer
      token: ${MCP_TOKEN}
```

### 2. CI/CD Integration

```yaml
# GitHub Actions Example
- name: Query Performance
  uses: newrelic/mcp-action@v1
  with:
    tool: nrql.execute
    query: |
      SELECT percentile(duration, 95)
      FROM Transaction
      WHERE appName = '${{ github.event.repository.name }}'
      SINCE 1 hour ago
```

### 3. Monitoring Integration

```go
// Prometheus metrics
mcp_requests_total{tool="nrql.execute", status="success"}
mcp_request_duration_seconds{tool="discovery.explore"}
mcp_active_workflows{type="investigation"}
```

## Technical Platform Specification

For the complete, unified technical platform specification that includes:

- **North-Star Goals**: Zero assumptions, universal compatibility, continuous adaptation
- **Layered Architecture**: Code-level package structure and responsibilities
- **Tool Contract**: JSON-RPC protocol and metadata specifications
- **Canonical Discovery Chains**: Standard patterns for discovering service IDs, errors, metrics
- **Adaptive Query Builder**: Dynamic query construction based on discoveries
- **Workflow Orchestrator**: YAML-based workflow definitions
- **Testing Matrix**: Unit, contract, E2E, and load testing strategies
- **Security & Governance**: Dry-run mode, audit logging, RBAC
- **Implementation Roadmap**: Sprint-by-sprint deliverables to GA

See the comprehensive [Technical Platform Specification](../TECHNICAL_PLATFORM_SPEC.md).

---

This architecture provides a solid foundation for building reliable, scalable, and intelligent observability solutions. The discovery-first approach, combined with granular tools and robust infrastructure, enables AI assistants to work effectively with New Relic data without making assumptions about the underlying systems.

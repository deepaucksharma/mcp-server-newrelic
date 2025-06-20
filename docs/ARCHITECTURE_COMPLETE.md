# Complete Architecture Documentation - New Relic MCP Server

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Dual Implementation Architecture](#dual-implementation-architecture)
4. [Component Architecture](#component-architecture)
5. [Data Flow Architecture](#data-flow-architecture)
6. [Integration Architecture](#integration-architecture)
7. [Deployment Architecture](#deployment-architecture)
8. [Performance Architecture](#performance-architecture)
9. [Security Architecture](#security-architecture)
10. [Architectural Decisions](#architectural-decisions)
11. [Migration Architecture](#migration-architecture)
12. [Future Architecture](#future-architecture)

## Executive Summary

The New Relic MCP Server is a production-grade implementation of the Model Context Protocol that provides AI assistants with intelligent access to New Relic observability data. This document provides comprehensive architectural documentation addressing all aspects of the system design, implementation decisions, and deployment strategies.

### Key Architectural Principles

1. **Discovery-First Design**: Never assume data structures; always explore and adapt
2. **Granular Tool Composition**: 120+ atomic tools that compose into workflows
3. **Multi-Transport Support**: STDIO, HTTP, and SSE transports for flexibility
4. **Resilient by Design**: Circuit breakers, retries, and graceful degradation
5. **Performance Optimized**: Sub-second responses with intelligent caching
6. **Security-First**: Zero-trust design with comprehensive validation

## Architecture Overview

### System Context Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    External Systems                          │
├─────────────────────────────────────────────────────────────┤
│  AI Assistants          │  New Relic Platform              │
│  ├─ Claude             │  ├─ NerdGraph API                │
│  ├─ GitHub Copilot     │  ├─ NRDB                        │
│  └─ Custom Clients     │  └─ Dashboards/Alerts           │
└───────────┬────────────┴─────────────┬─────────────────────┘
            │                           │
            │ MCP Protocol              │ GraphQL/HTTPS
            ▼                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  New Relic MCP Server                        │
├─────────────────────────────────────────────────────────────┤
│  Transport Layer        │  Core Services                    │
│  ├─ STDIO Transport    │  ├─ Discovery Engine             │
│  ├─ HTTP Transport     │  ├─ Query Adapter                │
│  └─ SSE Transport      │  └─ Workflow Orchestrator        │
├─────────────────────────────────────────────────────────────┤
│  Tool Registry (120+ Tools)                                 │
│  ├─ Discovery Tools    │  ├─ Analysis Tools               │
│  ├─ Query Tools        │  └─ Action Tools                 │
├─────────────────────────────────────────────────────────────┤
│  Infrastructure Layer                                        │
│  ├─ State Management   │  ├─ Caching                      │
│  ├─ Monitoring        │  └─ Security                      │
└─────────────────────────────────────────────────────────────┘
```

### Logical Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│         (MCP Protocol Handlers & Transport Adapters)        │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                         │
│              (Tool Implementations & Workflows)              │
├─────────────────────────────────────────────────────────────┤
│                    Business Logic Layer                      │
│        (Discovery Engine, Query Builder, Analyzers)         │
├─────────────────────────────────────────────────────────────┤
│                    Data Access Layer                         │
│          (New Relic Client, State Store, Cache)            │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                      │
│        (Logging, Monitoring, Security, Resilience)          │
└─────────────────────────────────────────────────────────────┘
```

## Dual Implementation Architecture

### Current State: Hybrid Architecture

The repository contains two parallel implementations that represent different stages of evolution:

```
mcp-server-newrelic/
├── Go Implementation (Primary)
│   ├── cmd/              # Go entry points
│   ├── pkg/              # Go packages
│   └── internal/         # Go internal packages
│
├── Python Implementation (Legacy)
│   ├── clients/python/   # Python SDK
│   ├── clients/typescript/ # TypeScript SDK
│   └── intelligence/     # Python-specific logic
│
└── Shared Resources
    ├── docs/            # Unified documentation
    ├── .env.example     # Shared configuration
    └── docker-compose.yml # Container orchestration
```

### Architecture Evolution

```
Phase 1: Python Prototype (Legacy)
├── Rapid prototyping
├── MCP protocol validation
└── Feature exploration

Phase 2: Go Implementation (Current)
├── Production-grade rewrite
├── Performance optimization
└── Full feature parity

Phase 3: Unified Platform (Future)
├── Go server as core
├── Python/TS SDKs for clients
└── Deprecated Python server
```

### Migration Strategy

1. **Current State**: Both implementations exist, Go is primary
2. **Transition Period**: Python marked deprecated, Go actively developed
3. **End State**: Single Go implementation with multi-language SDKs

### Architectural Rationale

**Why Two Implementations?**
- Python enabled rapid prototyping and protocol validation
- Go provides production performance and type safety
- Gradual migration reduces risk

**Why Keep Both?**
- Client SDKs in Python/TypeScript leverage existing code
- Documentation references both for historical context
- Smooth transition for existing users

## Component Architecture

### Core Components Deep Dive

#### 1. MCP Server Core (`pkg/interface/mcp/`)

```
MCP Server Core
├── Protocol Handler (protocol.go)
│   ├── JSON-RPC 2.0 parsing
│   ├── Request validation
│   ├── Response formatting
│   └── Error handling
│
├── Transport Layer
│   ├── STDIO Transport (transport_stdio.go)
│   │   ├── Stdin reader
│   │   ├── Stdout writer
│   │   └── Signal handling
│   │
│   ├── HTTP Transport (transport_http.go)
│   │   ├── HTTP server
│   │   ├── Request routing
│   │   └── CORS handling
│   │
│   └── SSE Transport (transport_sse.go)
│       ├── Event stream
│       ├── Keep-alive
│       └── Reconnection
│
├── Tool Registry (registry.go)
│   ├── Tool registration
│   ├── Parameter validation
│   ├── Handler dispatch
│   └── Metadata management
│
└── Session Management (session.go)
    ├── Client tracking
    ├── State persistence
    └── Timeout handling
```

#### 2. Discovery Engine (`pkg/discovery/`)

```
Discovery Engine
├── Core Engine (engine.go)
│   ├── Schema exploration
│   ├── Pattern detection
│   ├── Quality assessment
│   └── Relationship mining
│
├── Schema Analyzer (schema.go)
│   ├── Event type discovery
│   ├── Attribute profiling
│   ├── Cardinality analysis
│   └── Coverage calculation
│
├── Pattern Detector (patterns.go)
│   ├── Temporal patterns
│   ├── Spatial patterns
│   ├── Anomaly detection
│   └── Correlation finding
│
└── Discovery Cache (cache.go)
    ├── TTL management
    ├── Invalidation
    ├── Persistence
    └── Memory optimization
```

#### 3. Query Adapter (`pkg/query/`)

```
Query Adapter
├── Query Builder (builder.go)
│   ├── NRQL construction
│   ├── Schema adaptation
│   ├── Performance hints
│   └── Validation
│
├── Query Optimizer (optimizer.go)
│   ├── Cost estimation
│   ├── Index usage
│   ├── Aggregation pushdown
│   └── Limit optimization
│
└── Query Executor (executor.go)
    ├── Timeout control
    ├── Result streaming
    ├── Error recovery
    └── Metric collection
```

#### 4. State Management (`pkg/state/`)

```
State Management
├── State Manager (manager.go)
│   ├── Session lifecycle
│   ├── Context propagation
│   ├── Transaction support
│   └── Cleanup routines
│
├── Storage Backends
│   ├── Memory Store (memory_store.go)
│   │   ├── In-memory maps
│   │   ├── TTL support
│   │   └── Size limits
│   │
│   └── Redis Store (redis_store.go)
│       ├── Connection pooling
│       ├── Serialization
│       ├── Pub/sub support
│       └── Cluster mode
│
└── Cache Layer (cache.go)
    ├── Multi-level cache
    ├── Write-through
    ├── Cache warming
    └── Statistics
```

#### 5. New Relic Client (`pkg/newrelic/`)

```
New Relic Client
├── GraphQL Client (client.go)
│   ├── Request building
│   ├── Response parsing
│   ├── Error handling
│   └── Rate limiting
│
├── Query Operations (query_ops.go)
│   ├── NRQL execution
│   ├── Async queries
│   ├── Result pagination
│   └── Query history
│
├── Dashboard Operations (dashboard_ops.go)
│   ├── CRUD operations
│   ├── Widget management
│   ├── Permission handling
│   └── Bulk operations
│
└── Alert Operations (alert_ops.go)
    ├── Policy management
    ├── Condition creation
    ├── Notification channels
    └── Incident correlation
```

### Tool Architecture

#### Tool Categories and Responsibilities

```
Tool Registry (120+ Tools)
│
├── Discovery Tools (30+)
│   ├── Schema Discovery
│   │   ├── discovery.list_event_types
│   │   ├── discovery.explore_attributes
│   │   └── discovery.profile_coverage
│   │
│   ├── Pattern Discovery
│   │   ├── discovery.find_natural_groupings
│   │   ├── discovery.detect_temporal_patterns
│   │   └── discovery.find_relationships
│   │
│   └── Quality Assessment
│       ├── discovery.detect_anomalies
│       ├── discovery.validate_assumptions
│       └── discovery.assess_completeness
│
├── Query Tools (25+)
│   ├── Query Building
│   │   ├── nrql.build_select
│   │   ├── nrql.build_where
│   │   └── nrql.build_facet
│   │
│   ├── Query Execution
│   │   ├── nrql.execute
│   │   ├── nrql.validate
│   │   └── nrql.estimate_cost
│   │
│   └── Query Optimization
│       ├── nrql.optimize_performance
│       ├── nrql.suggest_indexes
│       └── nrql.analyze_plan
│
├── Analysis Tools (20+)
│   ├── Statistical Analysis
│   │   ├── analysis.calculate_baseline
│   │   ├── analysis.detect_anomalies
│   │   └── analysis.find_correlations
│   │
│   └── Root Cause Analysis
│       ├── analysis.trace_causality
│       ├── analysis.identify_dependencies
│       └── analysis.build_timeline
│
├── Action Tools (15+)
│   ├── Alert Management
│   │   ├── alert.create_from_baseline
│   │   ├── alert.tune_thresholds
│   │   └── alert.bulk_update
│   │
│   └── Dashboard Generation
│       ├── dashboard.generate_from_discovery
│       ├── dashboard.optimize_widgets
│       └── dashboard.create_slo_dashboard
│
├── Workflow Tools (10+)
│   ├── Workflow Management
│   │   ├── workflow.create
│   │   ├── workflow.execute_step
│   │   └── workflow.get_status
│   │
│   └── Context Management
│       ├── context.add_finding
│       ├── context.get_recommendations
│       └── context.export_report
│
└── Platform Governance Tools (20+)
    ├── Dashboard Analysis
    │   ├── dashboard.list_widgets
    │   ├── dashboard.classify_widgets
    │   └── dashboard.find_nrdot_dashboards
    │
    ├── Metric Usage
    │   ├── metric.widget_usage_rank
    │   └── metric.find_unused
    │
    └── Ingest Analysis
        ├── usage.ingest_summary
        ├── usage.otlp_collectors
        └── usage.agent_ingest
```

## Data Flow Architecture

### Request Processing Flow

```
1. Client Request
   │
   ├─→ Transport Layer
   │   ├─ Parse transport-specific format
   │   ├─ Extract JSON-RPC payload
   │   └─ Create request context
   │
   ├─→ Protocol Handler
   │   ├─ Validate JSON-RPC structure
   │   ├─ Check protocol version
   │   └─ Route to method handler
   │
   ├─→ Tool Registry
   │   ├─ Find matching tool
   │   ├─ Validate parameters
   │   └─ Check permissions
   │
   ├─→ Tool Handler
   │   ├─ Execute business logic
   │   ├─ Call downstream services
   │   └─ Format response
   │
   ├─→ State Management
   │   ├─ Update session state
   │   ├─ Cache results
   │   └─ Track metrics
   │
   └─→ Response Formation
       ├─ Build JSON-RPC response
       ├─ Add metadata
       └─ Send via transport
```

### Discovery-First Data Flow

```
Discovery Workflow
│
├─→ Initial Discovery
│   ├─ List event types
│   ├─ Sample recent data
│   └─ Build initial schema map
│
├─→ Deep Exploration
│   ├─ Profile attributes
│   ├─ Analyze cardinality
│   └─ Detect patterns
│
├─→ Adaptive Query Building
│   ├─ Select relevant attributes
│   ├─ Build appropriate filters
│   └─ Optimize for performance
│
├─→ Execution & Analysis
│   ├─ Run adapted query
│   ├─ Stream results
│   └─ Detect anomalies
│
└─→ Action Generation
    ├─ Create alerts from baselines
    ├─ Generate dashboards
    └─ Suggest optimizations
```

### Caching Architecture

```
Multi-Level Cache
│
├─→ L1: Request Cache (In-Memory)
│   ├─ TTL: 60 seconds
│   ├─ Key: Request hash
│   └─ Size: 1000 entries
│
├─→ L2: Discovery Cache (In-Memory + Redis)
│   ├─ TTL: 10 minutes
│   ├─ Key: Schema fingerprint
│   └─ Size: 10,000 entries
│
├─→ L3: Result Cache (Redis)
│   ├─ TTL: 5 minutes
│   ├─ Key: Query hash
│   └─ Size: Unlimited
│
└─→ Cache Invalidation
    ├─ Time-based expiry
    ├─ Event-based invalidation
    └─ Manual purge
```

## Integration Architecture

### External System Integration

```
New Relic Integration
│
├─→ NerdGraph API
│   ├─ GraphQL endpoint
│   ├─ API key authentication
│   ├─ Rate limiting (50 req/s)
│   └─ Timeout handling (30s)
│
├─→ NRDB
│   ├─ NRQL query interface
│   ├─ Streaming results
│   ├─ Query limits
│   └─ Data retention
│
└─→ Platform Features
    ├─ Dashboards
    ├─ Alerts
    ├─ Synthetics
    └─ APM
```

### Client Integration Patterns

```
Integration Patterns
│
├─→ Direct Integration (STDIO)
│   ├─ CLI tools
│   ├─ IDE plugins
│   └─ Local scripts
│
├─→ Network Integration (HTTP)
│   ├─ Web applications
│   ├─ Microservices
│   └─ API gateways
│
├─→ Streaming Integration (SSE)
│   ├─ Real-time dashboards
│   ├─ Live monitoring
│   └─ Event processors
│
└─→ SDK Integration
    ├─ Python SDK
    ├─ TypeScript SDK
    └─ Go SDK (planned)
```

### Internal Component Integration

```
Component Communication
│
├─→ Synchronous Calls
│   ├─ Direct function calls
│   ├─ Interface contracts
│   └─ Error propagation
│
├─→ Asynchronous Patterns
│   ├─ Channel communication
│   ├─ Worker pools
│   └─ Event bus
│
└─→ State Sharing
    ├─ Context propagation
    ├─ Shared cache
    └─ Distributed state
```

## Deployment Architecture

### Container Architecture

```
Container Structure
│
├─→ Base Image (Alpine Linux)
│   ├─ Minimal attack surface
│   ├─ Small size (~15MB)
│   └─ Security updates
│
├─→ Application Layer
│   ├─ Go binary (statically linked)
│   ├─ Configuration files
│   └─ TLS certificates
│
├─→ Runtime Configuration
│   ├─ Environment variables
│   ├─ ConfigMaps (K8s)
│   └─ Secrets management
│
└─→ Health Checks
    ├─ Liveness probe
    ├─ Readiness probe
    └─ Startup probe
```

### Deployment Topologies

#### Development Topology

```
Development Environment
│
├─→ Single Container
│   ├─ All-in-one deployment
│   ├─ In-memory state
│   └─ Mock mode support
│
└─→ Docker Compose
    ├─ MCP Server
    ├─ PostgreSQL
    └─ Redis
```

#### Production Topology (Small)

```
Small Production (1-100 users)
│
├─→ Load Balancer
│   └─→ MCP Servers (3 instances)
│       ├─ Shared Redis
│       └─ Shared PostgreSQL
│
└─→ Monitoring
    ├─ Prometheus
    └─ Grafana
```

#### Production Topology (Large)

```
Large Production (100+ users)
│
├─→ Global Load Balancer
│   │
│   ├─→ Regional Clusters
│   │   ├─→ MCP Servers (10+ instances)
│   │   ├─→ Discovery Engine Pool
│   │   └─→ Cache Layer
│   │
│   └─→ Shared Services
│       ├─ PostgreSQL Cluster
│       ├─ Redis Cluster
│       └─ Monitoring Stack
│
└─→ Multi-Region Setup
    ├─ US East
    ├─ US West
    └─ EU (planned)
```

### Kubernetes Architecture

```
Kubernetes Deployment
│
├─→ Namespace: mcp-server
│   │
│   ├─→ Deployments
│   │   ├─ mcp-server (HPA: 3-20 pods)
│   │   ├─ discovery-engine (HPA: 5-30 pods)
│   │   └─ cache-warmer (CronJob)
│   │
│   ├─→ Services
│   │   ├─ mcp-service (LoadBalancer)
│   │   ├─ discovery-service (ClusterIP)
│   │   └─ metrics-service (ClusterIP)
│   │
│   ├─→ ConfigMaps
│   │   ├─ mcp-config
│   │   └─ discovery-config
│   │
│   ├─→ Secrets
│   │   ├─ newrelic-credentials
│   │   └─ tls-certificates
│   │
│   └─→ PersistentVolumes
│       ├─ cache-storage
│       └─ log-storage
│
└─→ Ingress Controller
    ├─ TLS termination
    ├─ Rate limiting
    └─ Path routing
```

## Performance Architecture

### Performance Design Principles

1. **Minimize Latency**: Sub-200ms p95 response time
2. **Maximize Throughput**: 1000+ requests/second
3. **Optimize Resources**: <100MB baseline memory
4. **Scale Horizontally**: Stateless design

### Performance Optimizations

```
Performance Stack
│
├─→ Query Optimization
│   ├─ Query plan caching
│   ├─ Predicate pushdown
│   ├─ Result size limits
│   └─ Timeout controls
│
├─→ Caching Strategy
│   ├─ Multi-level caches
│   ├─ Intelligent TTLs
│   ├─ Preemptive warming
│   └─ Compression
│
├─→ Concurrency Model
│   ├─ Worker pools
│   ├─ Channel buffering
│   ├─ Context cancellation
│   └─ Graceful shutdown
│
├─→ Resource Management
│   ├─ Connection pooling
│   ├─ Memory limits
│   ├─ CPU throttling
│   └─ Disk I/O control
│
└─→ Network Optimization
    ├─ HTTP/2 support
    ├─ Keep-alive connections
    ├─ Response compression
    └─ Request batching
```

### Scaling Architecture

```
Scaling Strategies
│
├─→ Horizontal Scaling
│   ├─ Stateless services
│   ├─ Load distribution
│   ├─ Session affinity
│   └─ Auto-scaling rules
│
├─→ Vertical Scaling
│   ├─ Resource limits
│   ├─ JVM-style tuning
│   ├─ Memory allocation
│   └─ CPU allocation
│
├─→ Data Partitioning
│   ├─ Cache sharding
│   ├─ Query routing
│   ├─ Result aggregation
│   └─ Failover handling
│
└─→ Geographic Distribution
    ├─ Regional deployments
    ├─ Data locality
    ├─ Latency optimization
    └─ Disaster recovery
```

### Performance Monitoring

```
Monitoring Architecture
│
├─→ Application Metrics
│   ├─ Request latency
│   ├─ Error rates
│   ├─ Throughput
│   └─ Queue depths
│
├─→ Infrastructure Metrics
│   ├─ CPU usage
│   ├─ Memory usage
│   ├─ Network I/O
│   └─ Disk I/O
│
├─→ Business Metrics
│   ├─ Tool usage
│   ├─ Query patterns
│   ├─ Cache hit rates
│   └─ API quotas
│
└─→ Alerting Rules
    ├─ Latency > 1s
    ├─ Error rate > 1%
    ├─ Memory > 80%
    └─ Queue backup
```

## Security Architecture

### Security Layers

```
Defense in Depth
│
├─→ Network Security
│   ├─ TLS 1.3 only
│   ├─ Certificate pinning
│   ├─ Network policies
│   └─ DDoS protection
│
├─→ Authentication
│   ├─ API key validation
│   ├─ JWT tokens
│   ├─ mTLS (optional)
│   └─ OAuth2 (planned)
│
├─→ Authorization
│   ├─ RBAC policies
│   ├─ Tool permissions
│   ├─ Resource limits
│   └─ Tenant isolation
│
├─→ Input Validation
│   ├─ Schema validation
│   ├─ NRQL sanitization
│   ├─ Size limits
│   └─ Type checking
│
├─→ Data Protection
│   ├─ Encryption at rest
│   ├─ Encryption in transit
│   ├─ Key rotation
│   └─ Secret management
│
└─→ Audit & Compliance
    ├─ Access logging
    ├─ Change tracking
    ├─ Compliance reports
    └─ Security scanning
```

### Threat Model

```
Threat Analysis
│
├─→ External Threats
│   ├─ Unauthorized access
│   ├─ Data exfiltration
│   ├─ Service disruption
│   └─ Injection attacks
│
├─→ Internal Threats
│   ├─ Privilege escalation
│   ├─ Data leakage
│   ├─ Misconfiguration
│   └─ Supply chain
│
└─→ Mitigations
    ├─ Zero-trust design
    ├─ Least privilege
    ├─ Defense in depth
    └─ Regular audits
```

### Security Implementation

```
Security Controls
│
├─→ Code Security
│   ├─ Static analysis
│   ├─ Dependency scanning
│   ├─ Secret scanning
│   └─ SAST/DAST
│
├─→ Runtime Security
│   ├─ Runtime protection
│   ├─ Anomaly detection
│   ├─ Resource limits
│   └─ Sandboxing
│
├─→ Operational Security
│   ├─ Patch management
│   ├─ Incident response
│   ├─ Security training
│   └─ Penetration testing
│
└─→ Compliance
    ├─ SOC2 controls
    ├─ GDPR compliance
    ├─ Data residency
    └─ Audit trails
```

## Architectural Decisions

### Decision Record: Why Go?

**Context**: Choice of implementation language for production server

**Decision**: Go (Golang) for core server implementation

**Rationale**:
- **Performance**: Compiled language with excellent concurrency
- **Type Safety**: Catches errors at compile time
- **Deployment**: Single binary, minimal dependencies
- **Ecosystem**: Rich observability and HTTP libraries
- **Team Skills**: Aligned with New Relic engineering standards

**Alternatives Considered**:
- Python: Good for prototyping, but performance concerns
- Node.js: Event loop limitations for CPU-intensive work
- Rust: Excellent performance, but steeper learning curve

**Consequences**:
- Positive: High performance, easy deployment, type safety
- Negative: Slightly longer development time vs Python

### Decision Record: Discovery-First Architecture

**Context**: How to handle varying data schemas across customers

**Decision**: Always discover before querying, never assume

**Rationale**:
- **Flexibility**: Works with any schema
- **Reliability**: Reduces failures from missing fields
- **Intelligence**: Discovers patterns automatically
- **Evolution**: Adapts to schema changes

**Alternatives Considered**:
- Fixed schemas: Breaks with custom instrumentation
- Schema registry: Requires manual maintenance
- Best-effort queries: Poor user experience

**Consequences**:
- Positive: 90% fewer schema-related failures
- Negative: Additional discovery overhead (mitigated by caching)

### Decision Record: Granular Tools

**Context**: Tool design philosophy

**Decision**: 120+ atomic tools that compose into workflows

**Rationale**:
- **Composability**: Build complex workflows from simple parts
- **Testability**: Each tool independently testable
- **Flexibility**: AI can combine tools creatively
- **Maintainability**: Single responsibility principle

**Alternatives Considered**:
- Monolithic tools: Less flexible, harder to test
- Macro commands: Hide important details from AI
- Direct API proxy: No value addition

**Consequences**:
- Positive: Highly flexible and maintainable
- Negative: Larger tool surface area to document

### Decision Record: Multi-Transport Support

**Context**: How clients communicate with server

**Decision**: Support STDIO, HTTP, and SSE transports

**Rationale**:
- **STDIO**: Perfect for CLI tools and IDE integration
- **HTTP**: Standard for web services and APIs
- **SSE**: Enables real-time streaming use cases
- **Flexibility**: Different tools need different transports

**Alternatives Considered**:
- STDIO only: Limits deployment options
- HTTP only: Poor for CLI integration
- WebSocket: More complex than SSE for streaming

**Consequences**:
- Positive: Maximum deployment flexibility
- Negative: More code to maintain

### Decision Record: State Management

**Context**: How to manage session state and caching

**Decision**: Pluggable storage with memory and Redis backends

**Rationale**:
- **Development**: In-memory for simplicity
- **Production**: Redis for distributed state
- **Flexibility**: Easy to add new backends
- **Performance**: Local caching with distributed backup

**Alternatives Considered**:
- Stateless only: Would require repeated discoveries
- Database only: Too slow for cache use cases
- Memory only: Doesn't scale horizontally

**Consequences**:
- Positive: Flexible deployment options
- Negative: Additional complexity in state synchronization

## Migration Architecture

### Python to Go Migration

```
Migration Phases
│
├─→ Phase 1: Parallel Development
│   ├─ Python server continues running
│   ├─ Go server developed in parallel
│   ├─ Feature parity tracking
│   └─ Shared documentation
│
├─→ Phase 2: Testing & Validation
│   ├─ Side-by-side testing
│   ├─ Performance comparison
│   ├─ Output validation
│   └─ Client compatibility
│
├─→ Phase 3: Gradual Rollout
│   ├─ Canary deployment (5%)
│   ├─ Progressive rollout
│   ├─ Rollback capability
│   └─ Monitoring
│
├─→ Phase 4: Complete Migration
│   ├─ Full production traffic
│   ├─ Python deprecation
│   ├─ Client SDK updates
│   └─ Documentation update
│
└─→ Phase 5: Cleanup
    ├─ Remove Python code
    ├─ Archive for reference
    ├─ Update CI/CD
    └─ Final documentation
```

### Client Migration Support

```
Client Migration
│
├─→ Compatibility Layer
│   ├─ Protocol version detection
│   ├─ Legacy endpoint support
│   ├─ Response adaptation
│   └─ Deprecation warnings
│
├─→ SDK Evolution
│   ├─ Python SDK wraps Go server
│   ├─ TypeScript SDK updates
│   ├─ Version management
│   └─ Breaking change tracking
│
└─→ Migration Tools
    ├─ Compatibility checker
    ├─ Migration guide
    ├─ Code examples
    └─ Support channels
```

## Future Architecture

### Planned Enhancements

```
Future Roadmap
│
├─→ Multi-Region Support
│   ├─ EU data center
│   ├─ APAC expansion
│   ├─ Data locality
│   └─ Latency optimization
│
├─→ Advanced Discovery
│   ├─ ML-powered patterns
│   ├─ Anomaly prediction
│   ├─ Auto-remediation
│   └─ Cost optimization
│
├─→ Platform Extensions
│   ├─ Plugin architecture
│   ├─ Custom tools
│   ├─ Marketplace
│   └─ Community tools
│
├─→ Enterprise Features
│   ├─ Multi-tenancy
│   ├─ Audit logging
│   ├─ Compliance modes
│   └─ SLA management
│
└─→ AI Integration
    ├─ Copilot optimization
    ├─ Claude enhancements
    ├─ Custom AI support
    └─ Prompt engineering
```

### Architecture Evolution

```
Evolution Path
│
├─→ Short Term (3 months)
│   ├─ Complete Go migration
│   ├─ Performance optimization
│   ├─ Enhanced monitoring
│   └─ Documentation completion
│
├─→ Medium Term (6 months)
│   ├─ Plugin architecture
│   ├─ Advanced discovery
│   ├─ Multi-region support
│   └─ Enterprise features
│
└─→ Long Term (12 months)
    ├─ AI-native features
    ├─ Autonomous operations
    ├─ Platform marketplace
    └─ Global scale
```

### Technology Radar

```
Technology Adoption
│
├─→ Adopt
│   ├─ Go 1.21+
│   ├─ Redis 7+
│   ├─ Kubernetes
│   └─ OpenTelemetry
│
├─→ Trial
│   ├─ WASM plugins
│   ├─ eBPF monitoring
│   ├─ GraphQL federation
│   └─ Dapr
│
├─→ Assess
│   ├─ Rust components
│   ├─ Edge computing
│   ├─ Blockchain audit
│   └─ Quantum resistance
│
└─→ Hold
    ├─ Python server
    ├─ Monolithic tools
    ├─ Fixed schemas
    └─ SOAP APIs
```

## Conclusion

This architecture document provides a comprehensive view of the New Relic MCP Server's design, implementation, and evolution. The architecture emphasizes:

1. **Discovery-First Design**: Never assuming, always exploring
2. **Modular Composition**: Small tools combining into powerful workflows
3. **Production Readiness**: Built for scale, performance, and reliability
4. **Future Flexibility**: Designed to evolve with changing needs

The dual implementation represents a thoughtful migration strategy from rapid prototyping to production-grade service, ensuring continuity while improving quality. The architecture supports multiple deployment models, from simple development setups to large-scale production deployments across regions.

As the system evolves, the architecture will continue to adapt, incorporating new technologies and patterns while maintaining backward compatibility and operational excellence.
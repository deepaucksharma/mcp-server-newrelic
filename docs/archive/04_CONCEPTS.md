# Core Concepts

Understanding these core concepts will help you effectively use the New Relic MCP Server and build powerful observability workflows.

## 📋 Table of Contents

1. [Model Context Protocol (MCP)](#model-context-protocol-mcp)
2. [Discovery-First Philosophy](#discovery-first-philosophy)
3. [Tool Composition](#tool-composition)
4. [State Management](#state-management)
5. [Transport Layers](#transport-layers)
6. [Error Handling](#error-handling)
7. [Caching Strategy](#caching-strategy)
8. [Security Model](#security-model)
9. [Mock Mode](#mock-mode)
10. [Best Practices](#best-practices)

## 🔌 Model Context Protocol (MCP)

### What is MCP?

The Model Context Protocol is a standardized way for AI assistants to interact with external systems through well-defined tools. Think of it as a universal adapter that allows AI models to:

- **Execute Actions**: Run queries, create alerts, build dashboards
- **Access Data**: Retrieve information from external systems
- **Maintain Context**: Keep track of conversations and workflows
- **Stream Results**: Handle long-running operations

### Why MCP Matters

```
Traditional Approach:          MCP Approach:
┌─────────────┐               ┌─────────────┐
│ AI Assistant│               │ AI Assistant│
└──────┬──────┘               └──────┬──────┘
       │                             │
   (Custom API)                   (MCP Protocol)
       │                             │
┌──────┴──────┐               ┌──────┴──────┐
│External Tool│               │  MCP Server │
└─────────────┘               └──────┬──────┘
                                     │
                              ┌──────┴──────┬─────────────┐
                              │   Tool 1    │   Tool 2    │ ...
                              └─────────────┴─────────────┘
```

### MCP Components

1. **Tools**: Discrete functions with defined inputs/outputs
2. **Resources**: Data sources that can be accessed
3. **Prompts**: Reusable prompt templates
4. **Sessions**: Stateful conversations

## 🔍 Discovery-First Philosophy

### The Problem with Assumptions

Traditional monitoring tools often assume:
- You know your data structure
- Event types are standardized
- Attributes are well-documented
- Relationships are understood

**Reality**: Every New Relic account is unique with custom events, attributes, and relationships.

### Discovery-First Approach

Instead of assumptions, we:

1. **Explore First**: Always discover what's actually there
2. **Profile Data**: Understand characteristics before querying
3. **Find Patterns**: Detect relationships automatically
4. **Adapt Dynamically**: Adjust to your specific data

### Example Workflow

```
Traditional:                    Discovery-First:
1. Write NRQL query      →     1. Discover event types
2. Hope it works        →     2. Explore attributes
3. Debug failures       →     3. Profile data patterns
4. Try again           →     4. Build informed query
```

### Benefits

- **No Prior Knowledge Required**: Works with any data
- **Reduces Errors**: Queries based on actual schema
- **Finds Hidden Insights**: Discovers unknown relationships
- **Adapts to Changes**: Handles schema evolution

## 🔧 Tool Composition

### Tools as Building Blocks

Each tool is designed to do one thing well. Power comes from combining them:

```
Simple Tools → Composed Workflows → Complex Solutions
```

### Composition Patterns

#### 1. Sequential Composition
```
discovery.explore_event_types
    ↓
discovery.explore_attributes
    ↓
query_nrdb
    ↓
dashboard.create_from_discovery
```

#### 2. Parallel Composition
```
         ┌→ analysis.calculate_baseline
query_nrdb ┼→ analysis.detect_anomalies
         └→ analysis.find_correlations
```

#### 3. Conditional Composition
```
discovery.assess_quality
    ↓
if (quality > threshold):
    → alert.create_from_baseline
else:
    → governance.optimize_costs
```

### Tool Categories Work Together

- **Discovery** → Provides schema for **Query**
- **Query** → Generates data for **Analysis**
- **Analysis** → Informs **Alert** thresholds
- **Governance** → Optimizes all operations

## 💾 State Management

### State Layers

The MCP server maintains state at multiple levels:

1. **Request State**: Per-request context
2. **Session State**: Conversation continuity
3. **Cache State**: Performance optimization
4. **Discovery State**: Schema knowledge

### State Storage Options

#### Memory Store (Default)
- **Pros**: Fast, simple, no dependencies
- **Cons**: Lost on restart, single instance
- **Use Case**: Development, single-user

#### Redis Store
- **Pros**: Persistent, distributed, scalable
- **Cons**: External dependency, complexity
- **Use Case**: Production, multi-instance

### State Lifecycle

```
Request → Session → Cache → Persistent
  (ms)     (min)    (hrs)    (days)
```

## 🚀 Transport Layers

### STDIO Transport

**Purpose**: Direct communication with AI assistants

```
AI Assistant ←→ STDIO ←→ MCP Server
```

**Characteristics**:
- Synchronous request/response
- JSON-RPC over standard I/O
- No network configuration
- Ideal for Claude Desktop

### HTTP Transport

**Purpose**: Web-based integrations

```
Web App → HTTP API → MCP Server
```

**Characteristics**:
- RESTful endpoints
- Stateless requests
- Standard HTTP verbs
- CORS support

### SSE Transport

**Purpose**: Real-time streaming

```
Client → SSE Stream → MCP Server
```

**Characteristics**:
- Server-sent events
- Unidirectional streaming
- Automatic reconnection
- Progress updates

### Choosing a Transport

| Use Case | Recommended Transport |
|----------|--------------------|
| Claude Desktop | STDIO |
| Web Dashboard | HTTP |
| CLI Tool | STDIO |
| Live Monitoring | SSE |
| API Integration | HTTP |

## ❌ Error Handling

### Error Philosophy

Errors should be:
- **Informative**: Clear description of what went wrong
- **Actionable**: Suggest how to fix it
- **Contextual**: Include relevant details
- **Graceful**: Degrade functionality, don't crash

### Error Types

#### 1. Validation Errors
```json
{
  "error": "Validation failed",
  "details": "event_type is required",
  "suggestion": "Provide an event_type parameter",
  "example": {"event_type": "Transaction"}
}
```

#### 2. API Errors
```json
{
  "error": "New Relic API error",
  "details": "Rate limit exceeded",
  "retry_after": 60,
  "suggestion": "Enable caching or reduce request frequency"
}
```

#### 3. Discovery Errors
```json
{
  "error": "No data found",
  "details": "Event type 'Custom' has no data in the last 24 hours",
  "suggestion": "Try a wider time range or different event type"
}
```

### Error Recovery Patterns

1. **Automatic Retry**: Transient failures
2. **Fallback**: Use cached or default data
3. **Circuit Breaker**: Prevent cascade failures
4. **Graceful Degradation**: Partial results

## 🚄 Caching Strategy

### Cache Levels

```
L1: Request Cache (milliseconds)
    ↓
L2: Memory Cache (minutes)
    ↓
L3: Redis Cache (hours)
    ↓
L4: Discovery Cache (days)
```

### What Gets Cached

#### Always Cached
- Discovery results (schemas, attributes)
- Expensive computations (baselines, analyses)
- Static data (account details)

#### Conditionally Cached
- Query results (based on time range)
- Alert configurations (with TTL)
- Dashboard definitions (with versioning)

#### Never Cached
- Real-time metrics
- Security tokens
- Mutation operations

### Cache Invalidation

```
Event: Schema change detected
  → Invalidate: Discovery cache
  → Cascade: Query templates, dashboards

Event: New data arrives
  → Invalidate: Query results
  → Preserve: Schema cache
```

## 🔐 Security Model

### Authentication Layers

1. **API Key Authentication**: New Relic credentials
2. **JWT Tokens**: Server-generated tokens
3. **Session Management**: Temporary access

### Authorization Model

```
User → API Key → Permissions → Operations
         ↓           ↓             ↓
    Validated    Checked      Executed
```

### Security Principles

- **Least Privilege**: Minimal required permissions
- **Defense in Depth**: Multiple security layers
- **Audit Everything**: Complete activity logs
- **Fail Secure**: Deny by default

## 🎭 Mock Mode

### Purpose

Mock mode enables development and testing without New Relic credentials.

### Mock Mode Features

- **Realistic Data**: Generated data that mimics real patterns
- **All Tools Work**: Every tool returns appropriate mock data
- **Error Simulation**: Test error handling paths
- **Consistent State**: Predictable responses

### When to Use Mock Mode

```bash
# Development
./mcp-server --mock

# Testing
MOCK_MODE=true go test

# Demos
docker run mcp-server:latest --mock
```

### Mock Data Characteristics

- Event types: Transaction, SystemSample, Custom events
- Attributes: Common fields with realistic values
- Time series: Synthetic patterns (trends, seasonality)
- Errors: Controlled error injection

## 💡 Best Practices

### 1. Always Start with Discovery
```bash
# Bad: Assume data structure
query_nrdb("SELECT * FROM MyEvent")  # May fail

# Good: Discover first
discovery.explore_event_types()
discovery.explore_attributes("MyEvent")
query_nrdb("SELECT discovered_fields FROM MyEvent")
```

### 2. Use Appropriate Time Ranges
```bash
# Discovery: Longer ranges for better sampling
discovery.profile_attribute(time_range="7 days")

# Real-time: Shorter ranges for current state
query_nrdb("... SINCE 5 minutes ago")

# Baselines: Historical data for accuracy
analysis.calculate_baseline(time_range="30 days")
```

### 3. Compose Tools Thoughtfully
```bash
# Build complexity gradually
Simple → Validate → Enhance → Automate

# Example progression:
1. query_nrdb (manual query)
2. query.validate_nrql (add validation)
3. query.execute_adaptive (add optimization)
4. workflow.execute_investigation (full automation)
```

### 4. Handle Errors Gracefully
```python
# Always check for partial results
result = discovery.explore_attributes()
if result.partial:
    log.warn(f"Partial results: {result.warning}")

# Provide fallbacks
try:
    live_data = query_nrdb(query)
except RateLimitError:
    cached_data = get_from_cache(query)
```

### 5. Optimize for Performance
```bash
# Use caching wisely
DISCOVERY_CACHE_TTL=7200  # 2 hours for stable data
QUERY_CACHE_TTL=300       # 5 minutes for metrics

# Batch operations when possible
bulk.create_alerts(alerts_list)  # Better than individual calls

# Enable streaming for large results
discovery.profile_attribute(stream=true)
```

### 6. Maintain Security
```bash
# Never log sensitive data
log.info("Query executed", user=user_id)  # Good
log.info(f"Query: {query}")  # Bad - might contain secrets

# Use least privilege
API_KEY_PERMISSIONS=["nrql:query"]  # Only what's needed

# Rotate credentials regularly
```

## 🎯 Summary

Understanding these concepts enables you to:

- **Build Powerful Workflows**: Compose tools effectively
- **Handle Any Data**: Discovery-first approach works universally
- **Optimize Performance**: Use caching and state management
- **Ensure Security**: Follow security best practices
- **Develop Efficiently**: Mock mode for rapid development

## 📚 Related Documentation

- [Getting Started](01_GETTING_STARTED.md) - Apply these concepts
- [Tools Overview](30_TOOLS_OVERVIEW.md) - Available tools
- [Architecture Overview](10_ARCHITECTURE_OVERVIEW.md) - Technical details
- [Best Practices Guide](49_GUIDE_BEST_PRACTICES.md) - Advanced patterns

---

**Remember**: The power of the MCP server comes from understanding these concepts and applying them creatively to solve your observability challenges.
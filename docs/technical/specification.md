# Technical Specification - New Relic MCP Server

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [MCP Protocol Implementation](#mcp-protocol-implementation)
4. [Tool Specifications](#tool-specifications)
5. [Data Models](#data-models)
6. [API Specifications](#api-specifications)
7. [Security](#security)
8. [Performance Requirements](#performance-requirements)
9. [Configuration](#configuration)
10. [Error Handling](#error-handling)

## Overview

The New Relic MCP Server is a Model Context Protocol compliant server that provides AI assistants with programmatic access to New Relic observability data. Built in Go, it offers a comprehensive suite of tools for querying, analyzing, and managing New Relic resources.\n\n⚠️ **Implementation Status**: This specification describes the intended full functionality. Currently only ~10-15 basic tools are implemented out of the planned 120+. See [Implementation Gaps Analysis](../IMPLEMENTATION_GAPS_ANALYSIS.md) for details on what's actually available.

### Design Principles

1. **MCP Compliance**: Strict adherence to MCP DRAFT-2025 specification
2. **Type Safety**: Leveraging Go's type system for reliability
3. **Resilience**: Built-in circuit breakers, retries, and graceful degradation
4. **Performance**: Sub-second response times with intelligent caching
5. **Security**: Zero-trust design with comprehensive validation

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────┐
│                   Transport Layer                        │
│  ┌─────────┐    ┌──────────┐    ┌──────────────────┐  │
│  │  STDIO  │    │   HTTP   │    │       SSE        │  │
│  └────┬────┘    └─────┬────┘    └────────┬─────────┘  │
└───────┼───────────────┼───────────────────┼────────────┘
        │               │                   │
┌───────▼───────────────▼───────────────────▼────────────┐
│                 MCP Handler Layer                       │
│  ┌─────────────────────────────────────────────────┐  │
│  │           JSON-RPC Request Router                │  │
│  │  - Method dispatch                               │  │
│  │  - Parameter validation                          │  │
│  │  - Response formatting                           │  │
│  └─────────────────────────────────────────────────┘  │
└────────────────────────┬───────────────────────────────┘
                         │
┌────────────────────────▼───────────────────────────────┐
│                  Tool Registry                          │
│  ┌─────────────────────────────────────────────────┐  │
│  │  Registered Tools:                               │  │
│  │  - Query Tools (NRQL execution)                  │  │
│  │  - Discovery Tools (Schema analysis)             │  │
│  │  - Dashboard Tools (CRUD operations)             │  │
│  │  - Alert Tools (Management & analysis)           │  │
│  └─────────────────────────────────────────────────┘  │
└────────────────────────┬───────────────────────────────┘
                         │
┌────────────────────────▼───────────────────────────────┐
│                Business Logic Layer                     │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐  │
│  │  Discovery  │  │    State    │  │  New Relic   │  │
│  │   Engine    │  │   Manager   │  │    Client    │  │
│  └──────┬──────┘  └──────┬──────┘  └───────┬──────┘  │
└─────────┼────────────────┼──────────────────┼─────────┘
          │                │                  │
          └────────────────┼──────────────────┘
                           │
                    ┌──────▼──────┐
                    │ New Relic   │
                    │ NerdGraph   │
                    └─────────────┘
```

### Component Responsibilities

#### Transport Layer
- **STDIO**: Default transport for CLI integration
- **HTTP**: RESTful endpoint for web clients
- **SSE**: Server-Sent Events for streaming responses

#### MCP Handler
- JSON-RPC 2.0 protocol implementation
- Request validation and routing
- Error handling and response formatting
- Concurrent request handling

#### Tool Registry
- Dynamic tool registration
- Parameter schema validation
- Tool discovery for MCP clients
- Handler invocation with context

#### Business Logic
- **Discovery Engine**: Schema analysis, pattern detection
- **State Manager**: Session tracking, caching
- **New Relic Client**: GraphQL API integration

## MCP Protocol Implementation

### Protocol Version
```json
{
  "protocolVersion": "2024-11-05",
  "serverInfo": {
    "name": "New Relic MCP Server",
    "version": "1.0.0-beta"
  }
}
```

### Supported Methods

#### Core Methods
- `initialize` - Protocol handshake
- `initialized` - Confirm initialization
- `tools/list` - List available tools
- `tools/call` - Execute a tool
- `resources/list` - List available resources
- `resources/read` - Read a resource

#### Tool Invocation Format
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_nrdb",
    "arguments": {
      "query": "SELECT count(*) FROM Transaction",
      "account_id": "optional-account-override"
    }
  },
  "id": "unique-request-id"
}
```

### Response Format
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Query executed successfully"
      },
      {
        "type": "resource",
        "resource": {
          "uri": "newrelic://query/result/12345",
          "mimeType": "application/json",
          "text": "{\"results\": [{\"count\": 42}]}"
        }
      }
    ]
  },
  "id": "unique-request-id"
}
```

## Tool Specifications

### Query Tools

#### query_nrdb
**Purpose**: Execute NRQL queries against New Relic

**Parameters**:
```typescript
{
  query: string;        // Required: NRQL query
  account_id?: string;  // Optional: Override default account
  timeout?: number;     // Optional: Query timeout in seconds (default: 30)
}
```

**Response**:
```typescript
{
  results: Array<Record<string, any>>;
  metadata: {
    executionTime: number;
    resultCount: number;
    facets?: string[];
    messages?: string[];
  };
}
```

#### query_check
**Purpose**: Validate NRQL syntax and estimate query cost

**Parameters**:
```typescript
{
  query: string;
  account_id?: string;
}
```

**Response**:
```typescript
{
  valid: boolean;
  errors?: string[];
  warnings?: string[];
  estimatedCost?: {
    inspectedCount: number;
    omittedCount: number;
  };
  suggestions?: string[];
}
```

### Discovery Tools

#### discovery.list_schemas
**Purpose**: List all available schemas in NRDB

**Parameters**:
```typescript
{
  filter?: string;           // Optional: Filter pattern
  include_quality?: boolean; // Optional: Include quality metrics
}
```

**Response**:
```typescript
{
  schemas: Array<{
    name: string;
    attribute_count: number;
    record_count: number;
    last_updated: string;
    quality?: {
      score: number;
      issues: number;
    };
  }>;
  count: number;
}
```

#### discovery.profile_attribute
**Purpose**: Deep analysis of a specific attribute

**Parameters**:
```typescript
{
  schema: string;      // Required: Schema name
  attribute: string;   // Required: Attribute name
  sample_size?: number; // Optional: Sample size (default: 10000)
}
```

**Response**:
```typescript
{
  schema: string;
  attribute: {
    name: string;
    type: string;
    nullable: boolean;
    cardinality: {
      unique: number;
      total: number;
    };
    statistics?: {
      min: any;
      max: any;
      mean?: number;
      percentiles?: Record<string, number>;
    };
    patterns?: Array<{
      pattern: string;
      count: number;
      percentage: number;
    }>;
  };
}
```

### Dashboard Tools

#### generate_dashboard
**Purpose**: Generate dashboard from templates

**Parameters**:
```typescript
{
  template: "golden-signals" | "sli-slo" | "infrastructure" | "custom";
  name?: string;
  service_name?: string;    // For golden-signals
  host_pattern?: string;    // For infrastructure
  sli_config?: {           // For sli-slo
    name: string;
    target: number;
    query: string;
  };
  custom_config?: {        // For custom
    pages: Array<{
      name: string;
      widgets: Array<Widget>;
    }>;
  };
}
```

### Alert Tools

#### create_alert
**Purpose**: Create intelligent alert conditions

**Parameters**:
```typescript
{
  name: string;
  query: string;
  sensitivity?: "low" | "medium" | "high";
  comparison?: "above" | "below" | "equals";
  threshold_duration?: number;
  auto_baseline?: boolean;
  static_threshold?: number;
  policy_id?: string;
}
```

## Data Models

### Core Types

```go
// Tool Definition
type Tool struct {
    Name        string
    Description string
    Parameters  ToolParameters
    Handler     ToolHandler
    Streaming   bool
}

// Tool Parameters Schema
type ToolParameters struct {
    Type       string
    Required   []string
    Properties map[string]Property
}

// Property Definition
type Property struct {
    Type        string
    Description string
    Default     interface{}
    Enum        []string
    Items       *Property
}

// Session State
type Session struct {
    ID                string
    CreatedAt         time.Time
    LastAccessedAt    time.Time
    UserGoal          string
    DiscoveredSchemas []string
    QueryHistory      []QueryRecord
    Context          map[string]interface{}
}

// Query Record
type QueryRecord struct {
    Query      string
    ExecutedAt time.Time
    Results    interface{}
    Duration   time.Duration
    Error      error
}
```

### New Relic Specific Types

```go
// NRQL Result
type NRQLResult struct {
    Results  []map[string]interface{}
    Metadata QueryMetadata
}

// Query Metadata
type QueryMetadata struct {
    ExecutionTime   int64
    InspectedCount  int64
    ResultCount     int64
    Facets         []string
    TimeRange      TimeRange
}

// Schema Definition
type Schema struct {
    Name           string
    Description    string
    Attributes     []Attribute
    DataVolume     VolumeInfo
    Quality        QualityReport
    LastAnalyzedAt time.Time
}

// Alert Condition
type AlertCondition struct {
    ID               string
    Name             string
    PolicyID         string
    Query            string
    Threshold        float64
    ThresholdOccurences int
    ComparisonOperator string
    Enabled          bool
}
```

## API Specifications

### GraphQL Queries

#### NRQL Query Execution
```graphql
query ExecuteNRQL($accountId: Int!, $query: Nrql!) {
  actor {
    account(id: $accountId) {
      nrql(query: $query) {
        results
        metadata {
          executionTime
          inspectedCount
          resultCount
        }
      }
    }
  }
}
```

#### Entity Search
```graphql
query SearchEntities($query: String!, $limit: Int) {
  actor {
    entitySearch(query: $query) {
      results(limit: $limit) {
        entities {
          guid
          name
          type
          domain
          tags {
            key
            values
          }
        }
        nextCursor
      }
    }
  }
}
```

### REST Endpoints (HTTP Transport)

#### Health Check
```
GET /health
Response: {"status": "healthy", "version": "1.0.0-beta"}
```

#### Tool Execution
```
POST /tools/execute
Content-Type: application/json

{
  "tool": "query_nrdb",
  "params": {
    "query": "SELECT count(*) FROM Transaction"
  }
}
```

## Security

### Authentication
- API Key validation on every request
- No default credentials
- Support for multiple authentication methods

### Authorization
- Tool-level access control
- Account-based isolation
- Rate limiting per API key

### Input Validation
```go
// NRQL Query Validation
func validateNRQLQuery(query string) error {
    // Check for SQL injection patterns
    if containsSQLInjection(query) {
        return ErrInvalidQuery
    }
    
    // Validate query structure
    if !isValidNRQLSyntax(query) {
        return ErrInvalidSyntax
    }
    
    // Check query complexity
    if complexity := calculateComplexity(query); complexity > maxComplexity {
        return ErrQueryTooComplex
    }
    
    return nil
}
```

### Data Protection
- No logging of sensitive data
- Query result redaction options
- TLS encryption for all external calls
- Secure credential storage

## Performance Requirements

### Response Times
- **Query Tools**: < 2s for 95% of queries
- **Discovery Tools**: < 5s for schema analysis
- **Dashboard Tools**: < 1s for CRUD operations
- **Alert Tools**: < 2s for all operations

### Scalability
- Support 1000+ concurrent requests
- Horizontal scaling capability
- Stateless design for clustering

### Resource Limits
```go
const (
    MaxQueryTimeout     = 60 * time.Second
    MaxResultSize       = 10 * 1024 * 1024 // 10MB
    MaxConcurrentQueries = 100
    MaxSessionsPerUser   = 10
    CacheTTL            = 5 * time.Minute
)
```

## Configuration

### Environment Variables
```bash
# Required
NEW_RELIC_API_KEY=        # User API key
NEW_RELIC_ACCOUNT_ID=     # Default account

# Optional
NEW_RELIC_REGION=US       # US or EU
MCP_TRANSPORT=stdio       # stdio, http, sse
SERVER_PORT=8080         
LOG_LEVEL=INFO           
REDIS_URL=               # For distributed state
CACHE_TTL=300           # Cache TTL in seconds
MAX_CONCURRENT_REQUESTS=100
REQUEST_TIMEOUT=30s
```

### Configuration Struct
```go
type Config struct {
    NewRelic NewRelicConfig
    Server   ServerConfig
    Cache    CacheConfig
    Security SecurityConfig
}

type NewRelicConfig struct {
    APIKey          string
    AccountID       string
    Region          string
    GraphQLEndpoint string
}

type ServerConfig struct {
    Transport      string
    Port           int
    MaxConnections int
    ReadTimeout    time.Duration
    WriteTimeout   time.Duration
}
```

## Error Handling

### Error Categories

```go
// User Errors (4xx equivalent)
var (
    ErrInvalidInput     = NewError("INVALID_INPUT", "Invalid input parameters")
    ErrMissingRequired  = NewError("MISSING_REQUIRED", "Required parameter missing")
    ErrUnauthorized     = NewError("UNAUTHORIZED", "Authentication required")
    ErrForbidden        = NewError("FORBIDDEN", "Access denied")
    ErrNotFound         = NewError("NOT_FOUND", "Resource not found")
)

// System Errors (5xx equivalent)
var (
    ErrInternal         = NewError("INTERNAL_ERROR", "Internal server error")
    ErrServiceUnavailable = NewError("SERVICE_UNAVAILABLE", "New Relic API unavailable")
    ErrTimeout          = NewError("TIMEOUT", "Operation timed out")
    ErrRateLimit        = NewError("RATE_LIMIT", "Rate limit exceeded")
)
```

### Error Response Format
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "error_code": "MISSING_REQUIRED",
      "details": {
        "field": "query",
        "reason": "Required parameter 'query' is missing"
      },
      "suggestion": "Provide a valid NRQL query string"
    }
  },
  "id": "request-id"
}
```

### Retry Strategy
```go
type RetryConfig struct {
    MaxAttempts     int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    Multiplier      float64
    RetryableErrors []string
}

func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts:  3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay:     5 * time.Second,
        Multiplier:   2.0,
        RetryableErrors: []string{
            "SERVICE_UNAVAILABLE",
            "TIMEOUT",
            "RATE_LIMIT",
        },
    }
}
```

---

This technical specification serves as the authoritative reference for the New Relic MCP Server implementation. All development should align with these specifications to ensure consistency and interoperability.
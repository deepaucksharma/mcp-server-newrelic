# Technical Platform Specification

**A unified, build-ready blueprint for the discovery-first, zero-assumption MCP Server**

---

## 1 · North-Star Goals

| Dimension              | Target                                                                                              |
| ---------------------- | --------------------------------------------------------------------------------------------------- |
| **Assumptions**        | *Zero.* Every query, mutation, and analysis is preceded by discovery.                               |
| **Compatibility**      | Works with **any** New Relic account, schema, event mix, region, or data source (Agent, OTLP, API). |
| **Adaptation**         | Learns and caches discoveries; re-discovers on drift.                                               |
| **Explainability**     | Every output cites *what was discovered* and *how confident* we are.                                |
| **Extensibility**      | New tools drop in via a plugin pattern; discovery engine auto-surfaces them to LLMs.                |
| **Self-observability** | Server pushes its own metrics to New Relic and audits missing-discovery violations.                 |
| **Language**          | Core MCP server in Go for performance; client SDKs in Python, TypeScript, and more.                |

---

## 2 · Layered Architecture (Code-Level)

| Layer                       | Key Packages / Binaries                                                                                                                                    | Description                                                                                                                                                                                                                                                                                                     |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Edge**                    | `cmd/mcp-stdio`, `cmd/mcp-http`                                                                                                                            | Entry binaries exposing JSON-RPC over stdio or HTTP (TLS optional). Thin wrappers that pipe requests to Core.                                                                                                                                                                                                   |
| **Core**                    | `internal/core/server.go`                                                                                                                                  | Dispatch loop: validates JSON-RPC, looks up tool, injects discovery cache handle, executes, returns result/error.                                                                                                                                                                                               |
| **Discovery Engine**        | `internal/discovery/…`                                                                                                                                     | *The heart.* Implements every "NO ASSUMPTIONS" chain:<br>• `schemas.go` (event/list schemas)<br>• `attributes.go` (profile types/cardinality)<br>• `serviceid.go`, `errors.go`, `metrics.go` (specialised chains)<br>• `datasources.go` (agent vs OTLP)<br>All results store in Redis/Ristretto via `cache.go`. |
| **Tool Kits**               | `tools/…`                                                                                                                                                  | One Go file per tool family, e.g. `tools/nrql.go`, `tools/dashboards.go`, `tools/alerts.go`, `tools/usage.go`.<br>Each tool starts with **EnsureDiscovery()** helpers that read discovery cache or trigger discovery steps.                                                                                     |
| **Adapters**                | `internal/nrgraph/` (GraphQL client), `internal/insights/` (NRQL REST), `internal/otel/` (OTLP meta), `internal/billing/` (ingestUsage).                   | Client adapters for New Relic APIs                                                                                                                                                                                                                                                                              |
| **Intelligence** (optional) | `intelligence/` Optional Python microservice exposing advanced ML (anomaly clustering, pattern mining). Registered via gRPC; called only if enabled. | Advanced analysis capabilities using Python ML ecosystem                                                                                                                                                                                                                                                                                  |
| **Self-Monitoring**         | `internal/selfmon/` – wraps `expvar` + NR Telemetry SDK; emits tool latencies, discovery misses, AI misuse counters.                                       | Observability of the observability server                                                                                                                                                                                                                                                                       |
| **Docs & Tests**            | `docs/`, `examples/`, `tests/unit`, `tests/e2e`. Docs are first-class; tests assert discovery chains and API promises.                                     | Comprehensive documentation and testing                                                                                                                                                                                                                                                                         |

---

## 3 · Tool Contract (JSON-RPC)

### Request Format

```jsonc
{
  "jsonrpc": "2.0",
  "method": "nrql.query",        // snake.case family.tool
  "params": {
    "query": "SELECT count(*) FROM Transaction SINCE 1 hour AGO",
    "account_id": 123456,
    "discover_first": true       // default true; can skip if caller KNOWS
  },
  "id": "uuid-123"
}
```

### Tool Metadata (served at `/mcp.discover`)

```jsonc
{
  "tool": "nrql.query",
  "description": "Run an NRQL query after verifying event types & attributes exist.",
  "readOnlyHint": true,
  "parameters": {
    "type": "object",
    "properties": {
      "query": {"type":"string"},
      "account_id": {"type":"integer"},
      "discover_first": {"type":"boolean","default":true}
    },
    "required": ["query"]
  },
  "examples": [ … ]
}
```

> **Rule:** every tool function begins with `discovery.Require(ctx, prerequisites)` which aborts early with a *DiscoveryError* (code -40001) if required data is missing.

---

## 4 · Canonical Discovery Chains

| Domain                     | Chain Function               | Steps (ordered)                                                                              | Cached TTL |
| -------------------------- | ---------------------------- | -------------------------------------------------------------------------------------------- | ---------- |
| **Service Identifier**     | `discover.ServiceID()`       | `appName` → `service.name` → `applicationName` → custom pattern (regex) → natural clustering | 4 h        |
| **Error Indicator**        | `discover.ErrorIndicators()` | Boolean field → error class → HTTP/gRPC codes → log level → pattern match → anomaly-based    | 30 m       |
| **Metric Census**          | `discover.Dimensionals()`    | Metric tables → numeric event attributes → custom histograms                                 | 2 h        |
| **Data Sources**           | `discover.Sources()`         | NerdGraph `ingestUsage` → `instrumentation.provider` scan → agent names → custom tags        | 24 h       |
| **Dashboard Widget Types** | `discover.WidgetTypes()`     | Iterate dashboards → parse `rawConfiguration` → class metric vs NRQL                         | 6 h        |

All chains emit **`DiscoveryResult`** structs carrying `confidence`, `coverage`, `freshness`, and `assumptions` fields for transparency & AI explanation.

---

## 5 · Adaptive Query Builder (Go)

```go
type AQB struct {
    Intent        string   // "errorRate", "p95Latency"
    Scope         Scope    // service, entity, etc.
    Discoveries   *discovery.Cache
}

func (aq *AQB) Build() (string, error) {
    svc := aq.Discoveries.ServiceID(aq.Scope)              // dynamic field
    errs := aq.Discoveries.ErrorIndicators(aq.Scope)
    metric := aq.Discoveries.BestMetric("duration", aq.Scope)

    // Example: build error-rate NRQL robustly
    whereSvc := fmt.Sprintf("%s = '%s'", svc.Field, aq.Scope.Value)
    whereErr := errs.Top().Condition

    return fmt.Sprintf(`
        SELECT percentage(count(*), WHERE %s AND %s) 
        FROM %s 
        SINCE 1 hour AGO`, whereErr, whereSvc, metric.EventType), nil
}
```

---

## 6 · Workflow Orchestrator

*Location*: `internal/workflow/engine.go`

### Workflow Definition

YAML files in `workflows/` using the pattern:

```yaml
name: performance_investigation
description: Discovery-first perf drill-down
steps:
  - tool: discovery.list_schemas
  - tool: discovery.profile_attributes
  - tool: nrql.query
    params:
      query: "${aqb:build.latencyP95}"
  - tool: analysis.detect_anomalies
```

### Execution Model

* **Engine** loads YAML → substitutes `aqb:` token with Adaptive Query Builder output at runtime → executes sequentially or parallel (`mode:` key).
* **LLM Usage**: Copilot asks for goal → orchestrator returns plan steps → LLM approves/executes (or user can step through).

---

## 7 · Caching & Performance

| Cache                  | Tech                                             | Purpose                                   | Eviction               |
| ---------------------- | ------------------------------------------------ | ----------------------------------------- | ---------------------- |
| **Discovery Cache**    | Ristretto in-process; optional Redis for cluster | Store `DiscoveryResult` per account/field | TTL per chain (see §4) |
| **NRQL Result Cache**  | Ristretto (10 MB default)                        | Short-term dedup of repeated queries      | 5 min LRU              |
| **Widget Parse Cache** | in-mem sync.Map                                  | rawConfig → classification                | 12 h                   |

---

## 8 · Testing Matrix

| Level    | Tool                          | Strategy                                                                                  |
| -------- | ----------------------------- | ----------------------------------------------------------------------------------------- |
| Unit     | Go `testing` + `testify`      | Mock `nrgraph.Client` / `insights.Client` to validate discovery heuristics & edge cases.  |
| Contract | `scripts/assert_api_docs.py`  | Parse running server's `/mcp.discover` vs docs Markdown to prevent drift.                 |
| E2E LLM  | `tests/e2e/claude_scenarios/` | Harness uses Claude Tool-Use; verifies discovery-first call ordering & expected JSON-RPC. |
| Load     | `bench/discovery_bench.go`    | Replays 500 discovery calls; ensure p95 latency < 120 ms with cache warm.                 |

---

## 9 · Security & Governance

### Mutation Safety

*All destructive mutators* (`alerts.*`, `dashboards.delete`, etc.):

1. **Dry-Run Mode default** – executes NerdGraph with `dryRun:true`; real mutation only if param `confirm_token` echoes server-generated token from dry-run.
2. **Audit Log** – each mutation writes to `audit_log` table via NerdGraph `events.ingest`.
3. **RBAC Wrapper** – `internal/auth/` supports API key scopes: `read:tools`, `mutate:dashboards`, `admin:*`. Stdio mode bypasses; HTTP mode enforces bearer token.

---

## 10 · Roadmap (Development → GA)

| Sprint     | Deliverables                                                                                                                                        |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| **S-1**    | Discovery Engine MVP (schemas, attributes, serviceID). Refactor existing tools to call `EnsureDiscovery`. Docs: NO_ASSUMPTIONS_MANIFESTO.md live. |
| **S-2**    | Dashboard/Widget parser, ingest usage tools, adaptive query builder library, API docs autogen script.                                               |
| **S-3**    | Workflow YAML engine + first three workflows (perf investigate, incident, cost audit). Self-monitoring metrics.                                     |
| **S-4**    | Security layer (dry-run, audit log), CI contract test, client SDK improvements.                                                                      |
| **S-5 GA** | Multi-account selector, CLI wrapper (`mcpctl`) for manual ops. Tag v1.0.0, publish Docker image. (EU region already complete)                       |

---

## 11 · "Zero-Assumption" Compliance Checklist (CI Gate)

1. **No hard-coded attribute names** outside discovery packages.
2. **All tools call `EnsureDiscovery`** for every field/entity they rely on.
3. **Any new NRQL in code or template** passes `scripts/lint_nrql.py` which asserts it uses placeholder `${serviceField}` style tokens.
4. **Docs updated**: PR template blocks merge unless `docs-changed` label exists when new tools are added.
5. **LLM regression tests** green: discovery must precede first data query.

---

## 12 · Quick-Start for Contributors

```bash
# clone repository
git clone https://github.com/deepaucksharma/mcp-server-newrelic
cd mcp-server-newrelic

# set minimal env
cp .env.example .env           # fill NEW_RELIC_API_KEY, ACCOUNT_ID
make build                     # build Go binary
make run                       # launches MCP server

# smoke test: list tools (discovery endpoint)
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./bin/mcp-server
```

*Run `make docs` to build API reference; `make test` for unit & contract checks; `make bench` for perf.*

---

## Implementation Details

### Discovery Engine Implementation

```go
// internal/discovery/engine.go
type DiscoveryEngine interface {
    // Core discovery methods
    ServiceID(ctx context.Context, scope Scope) (*DiscoveryResult, error)
    ErrorIndicators(ctx context.Context, scope Scope) (*DiscoveryResult, error)
    Dimensionals(ctx context.Context, scope Scope) (*DiscoveryResult, error)
    Sources(ctx context.Context, scope Scope) (*DiscoveryResult, error)
    WidgetTypes(ctx context.Context, scope Scope) (*DiscoveryResult, error)
}

type DiscoveryResult struct {
    Type       string                 `json:"type"`
    Value      interface{}            `json:"value"`
    Confidence float64                `json:"confidence"`  // 0.0 to 1.0
    Coverage   float64                `json:"coverage"`    // % of data examined
    Freshness  time.Time              `json:"freshness"`
    Assumptions []string              `json:"assumptions"` // What we had to assume
    Metadata   map[string]interface{} `json:"metadata"`
}
```

### Tool Implementation Pattern

```go
// tools/nrql.go
func (t *NRQLTools) Query(ctx context.Context, params map[string]interface{}) (*Result, error) {
    // 1. Extract parameters
    query := params["query"].(string)
    discoverFirst := params["discover_first"].(bool) // default true
    
    // 2. Discovery phase (unless explicitly skipped)
    if discoverFirst {
        discoveries, err := t.discovery.EnsureForQuery(ctx, query)
        if err != nil {
            return nil, &DiscoveryError{
                Code: -40001,
                Message: fmt.Sprintf("Discovery failed: %v", err),
                Discoveries: discoveries,
            }
        }
        
        // 3. Adapt query based on discoveries
        query, err = t.aqb.Adapt(query, discoveries)
        if err != nil {
            return nil, fmt.Errorf("query adaptation failed: %w", err)
        }
    }
    
    // 4. Execute with confidence
    result, err := t.client.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // 5. Annotate with discovery metadata
    return &Result{
        Data: result,
        Meta: ResultMeta{
            DiscoveriesUsed: discoveries,
            QueryAdapted: query != params["query"].(string),
            Confidence: discoveries.MinConfidence(),
        },
    }, nil
}
```

### Workflow Engine Pattern

```go
// internal/workflow/engine.go
type WorkflowEngine struct {
    discovery  DiscoveryEngine
    tools      ToolRegistry
    cache      Cache
    aqb        *AdaptiveQueryBuilder
}

func (w *WorkflowEngine) Execute(ctx context.Context, workflow *Workflow) (*WorkflowResult, error) {
    state := NewWorkflowState()
    
    for _, step := range workflow.Steps {
        // Substitute AQB tokens
        params, err := w.substituteTokens(step.Params, state)
        if err != nil {
            return nil, fmt.Errorf("token substitution failed at step %s: %w", step.Name, err)
        }
        
        // Execute tool
        result, err := w.tools.Execute(ctx, step.Tool, params)
        if err != nil {
            if workflow.ContinueOnError {
                state.RecordError(step.Name, err)
                continue
            }
            return nil, err
        }
        
        // Update state
        state.Set(step.Name, result)
    }
    
    return &WorkflowResult{
        State: state,
        Summary: w.summarize(state),
    }, nil
}
```

---

### Final Word

This platform turns the **No-Assumptions Manifesto** into executable software:

*Discovery is not a phase; it's the foundation.*
*Every tool, query, dashboard, or alert adapts to the truths it uncovers—and tells the LLM exactly how sure it is.*

With this specification, the `new-branch` can evolve into a **universal, future-proof observability MCP server** that functions as an intelligent, self-learning partner—never a brittle script.
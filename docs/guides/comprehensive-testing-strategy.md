# Comprehensive Testing Strategy for MCP Server

This document outlines the multi-layered test approach for the MCP server.  It mirrors the strategy defined in the project discussion titled "Comprehensive Testing Strategy" and describes the structure and workflow for each level of testing.

## Test Taxonomy

| Layer | Purpose | Key Technologies |
| --- | --- | --- |
| **1. Lint & Static Analysis** | Catch style, vet issues, dead code, forbidden hard-codings. | `golangci-lint`, custom **Assumption Scanner** |
| **2. Unit Tests** | Validate each function/tool in isolation with mocks. | `go test`, `testify`, `mockery` |
| **3. Contract / Schema Tests** | Ensure every exported JSON-RPC method & payload stays compatible. | `quicktype`, `jsonschema`, snapshot tests |
| **4. Integration (New Relic Sandbox)** | Hit real NRDB & NerdGraph against seeded data; verify discovery & queries. | `testcontainers-go`, Terraform sandbox account |
| **5. Workflow Tests** | Execute YAML workflows end-to-end; assert intermediate & final results. | Custom **Workflow Harness** |
| **6. AI Harness Tests** | Drive Copilot & Claude through CLI/API, assert they plan & call tools correctly. | `gh copilot ask`, Anthropic SDK, harness scripts |
| **7. Performance / Load** | Validate p95 latency & memory under 100 req/s mixed load. | `ghz`, `k6`, Prometheus exporter |
| **8. Chaos / Fault-Injection** | Simulate NerdGraph outages, high cardinality, schema drift. | `toxiproxy`, random schema mutator |
| **9. Security & RBAC** | Ensure auth scopes, dry-run safety, audit events. | `go-test`, OWASP ZAP for HTTP |
| **10. Docs / Example Sync** | Keep Markdown examples executable & green. | `doctest` runner |

## Directory Layout

```
mcp-newrelic/
├─ tests/
│  ├─ unit/                 # _test.go files mirror package structure
│  ├─ contract/             # json snapshots + schema generation
│  ├─ integration/
│  │   ├─ terraform/        # spins sandbox account + seeds data
│  │   └─ containers/       # testcontainers compose
│  ├─ workflow/             # YAML + expectations
│  ├─ ai-harness/           # Copilot & Claude drivers
│  ├─ perf/                 # ghz & k6 scripts
│  ├─ chaos/                # toxiproxy configs
│  └─ docs-examples/        # MD code blocks executed as tests
└─ testdata/                # golden JSON-RPC requests/responses
```

## Implementation Details

### Assumption Scanner
A simple static gate that fails CI if any production Go file hard codes fields such as `appName` outside of discovery packages.

```bash
# scripts/assumption_scan.sh
grep -R --line-number --exclude-dir=internal/discovery -E '\"(appName|duration|error|Transaction)\"' $(git ls-files '*.go') && exit 1 || exit 0
```

Run this during the lint stage.

### Unit Tests

- Mocks generated for key clients (`nrgraph.Client`, `insights.Client`, `cache.Store`).
- Table driven tests cover success, missing attributes, and transient failures.
- Fuzz `nrql.Builder` to ensure no panics.

Target **≥ 90 % function coverage** for core packages.

### Contract Tests

1. Generate JSON schema from the tool registry at build time.
2. Store snapshots of public method examples in `tests/contract/golden/`.
3. `go test tests/contract` round trips against the schema and ensures breaking changes bump `x-version`.

### Integration Tests

- Terraform provisions a sandbox New Relic account and seeds events.
- `testcontainers-go` starts the server and a `toxiproxy` sidecar.
- Workflows validate discovery of event types, NRQL execution, dashboard creation, and schema error handling.

### Workflow Harness

A custom runner that executes YAML workflows and asserts each `StepResult` against expectations. It also checks the `/explain/{traceID}` endpoint for confidence annotations.

### AI Harness

Scripts drive Copilot and Claude through prompts to ensure the first tool used is discovery and that generated JSON-RPC requests succeed when sent to the server.

### Performance Tests

Use `ghz` to send 100 mixed JSON-RPC calls per second for two minutes. Passing criteria:

- p95 latency < 150ms
- error rate < 1%
- memory consumption < 250 MiB RSS

### Chaos / Drift

`toxiproxy` introduces latency and errors while a schema mutator drops attributes. Workflows should retry or return structured discovery errors.

### Security & RBAC

Unit tests validate middleware behavior. Integration tests ensure dry-run actions leave resources unchanged. OWASP ZAP scans the HTTP server for common issues.

### Docs / Example Sync

Markdown code blocks labeled `jsonrpc`, `bash`, or `go` are executed via a doctest runner so examples always stay up to date.

## CI/CD Pipeline

1. Lint and unit tests across Go versions.
2. Contract tests and doctest checks.
3. Integration tests via testcontainers.
4. Nightly AI harness runs for Copilot and Claude.
5. Weekly performance and chaos runs.
6. Coverage gate blocks merges below 85%.
7. Docker images publish on main branch merges.
8. Metrics are sent to a New Relic CI account.

## Local Developer Commands

```bash
make lint test          # fast feedback
make integration        # run integration tests
make ai                 # run AI harness locally
make perf               # run performance benchmarks
```

## Exit Criteria per PR

- All CI jobs green.
- Added or modified tools include unit and contract tests.
- Discovery tools update `tests/workflow/discovery_minimal.yaml`.
- New workflows require YAML examples and harness tests.
- Docs snippets updated and passing doctest.

This strategy ensures the MCP server reliably supports discovery-first automation for humans and AI agents without regressions across accounts and regions.


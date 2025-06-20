# Comprehensive Testing Strategy for mcp-server-newrelic

This document outlines the test taxonomy, directory structure, implementation details, CI pipeline, local developer commands, and exit criteria needed to ensure that the MCP server remains regression free. The strategy covers everything from linting and unit tests to AI harness checks and chaos testing.

## 1 · Test Taxonomy

| Layer | Purpose | Key Technologies |
| --- | --- | --- |
| **1. Lint & Static Analysis** | Catch style, vet issues, dead code, forbidden hard-codings. | `golangci-lint`, custom **Assumption Scanner** |
| **2. Unit Tests** | Validate each function/tool in isolation with mocks. | `go test`, `testify`, `mockery` |
| **3. Contract / Schema Tests** | Ensure every exported JSON-RPC method & payload stays compatible. | `quicktype`, `jsonschema`, snapshot tests |
| **4. Integration (New Relic Sandbox)** | Hit real NRDB & NerdGraph against seeded data; verify discovery & queries. | `testcontainers-go`, Terraform sandbox account |
| **5. Workflow Tests** | Execute YAML workflows end-to-end; assert intermediate & final results. | Custom **Workflow Harness** |
| **6. AI Harness Tests** | Drive Copilot & Claude through CLI/API, assert they plan & call tools correctly. | `gh copilot ask`, Anthropic SDK, harness scripts |
| **7. Performance / Load** | Validate p95 latency & memory under 100 req/s mixed load. | `ghz`, `k6`, Prometheus exporter |
| **8. Chaos / Fault-Injection** | Simulate NR GraphQL outages, high cardinality, schema drift. | `toxiproxy`, random schema mutator |
| **9. Security & RBAC** | Ensure auth scopes, dry-run safety, audit events. | `go-test`, OWASP ZAP for HTTP |
| **10. Docs / Example Sync** | Keep Markdown examples executable & green. | `doctest` runner |

## 2 · Directory Layout

```
mcp-newrelic/
├─ tests/
│  ├─ unit/
│  ├─ contract/
│  ├─ integration/
│  │   ├─ terraform/
│  │   └─ containers/
│  ├─ workflow/
│  ├─ ai-harness/
│  ├─ perf/
│  ├─ chaos/
│  └─ docs-examples/
└─ testdata/
```

## 3 · Implementation Details

### 3.1 Assumption Scanner (Static Gate)

```
# scripts/assumption_scan.sh
grep -R --line-number --exclude-dir=internal/discovery -E \
    "\"(appName|duration|error|Transaction)\"" $(git ls-files '*.go') && exit 1 || exit 0
```

### 3.2 Unit Tests

- Mock layers for `nrgraph.Client`, `insights.Client`, and `cache.Store`.
- Table-driven tests covering success, missing attributes, and transient failures.
- Fuzz `nrql.Builder` to ensure no panics.
- Target **≥ 90 % function coverage** for `internal/` and `tools/`.

### 3.3 Contract Tests

1. Generate JSON Schema from the tool registry at build time.
2. Snapshot all public method examples in `tests/contract/golden/`.
3. Verify round-trip marshal/unmarshal and version bumps.

### 3.4 Integration Tests (Sandbox)

- Terraform provisions a New Relic account and seeds sample events.
- `testcontainers-go` spins the MCP server plus a `toxiproxy` side-car.
- Exercise key discovery and NRQL flows, simulating missing attributes.

### 3.5 Workflow Harness

- Runs YAML files in `tests/workflow/` against a live server.
- Asserts each `StepResult` and validates trace explanations.

### 3.6 AI Harness

- Uses GitHub Copilot and Anthropic Claude to verify that AI agents invoke discovery tools first.
- Asserts success of generated JSON-RPC payloads.

### 3.7 Performance Tests

- `ghz` drives 100 mixed calls/sec for 2 minutes.
- Pass if p95 latency < 150 ms and memory < 250 MiB.

### 3.8 Chaos / Drift Tests

- `toxiproxy` introduces latency and error injection.
- Validate retries or graceful failures with code `-32098`.

### 3.9 Security & RBAC

- Unit and integration tests confirm token scopes and audit events.
- Run OWASP ZAP CLI against the server.

### 3.10 Docs / Example Sync

- Execute Markdown code fences labeled `jsonrpc`, `bash`, or `go` using a `doctest` runner.

## 4 · CI/CD Pipeline (GitHub Actions)

1. PR lint and unit tests (matrix on Go 1.22 and 1.23).
2. Contract tests and doctest.
3. Integration tests with testcontainers.
4. AI harness jobs (nightly).
5. Performance and chaos tests (weekly).
6. Coverage gate; block merge if < 85 %.
7. Docker build and publish on merge to `main`.
8. Metrics sent to the New Relic CI account.

## 5 · Local Developer Commands

```
make lint test      # fast feedback
make integration    # full integration suite
make ai             # run AI harness locally
make perf           # performance bench
```

## 6 · Exit Criteria per PR

- All CI jobs green.
- Added or modified tools must have unit and contract tests.
- Discovery-related tools require workflow updates.
- New workflows need examples and harness tests.
- Docs snippets must pass doctest.

This strategy couples granular unit safety with real New Relic integration, AI-driven regressions, and chaos resiliency.

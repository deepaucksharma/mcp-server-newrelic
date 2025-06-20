# Comprehensive Testing Strategy

This project uses a multi-layer approach:

1. **Lint & Static Analysis** – `golangci-lint` and `scripts/assumption_scan.sh`.
2. **Unit Tests** – mocked dependencies with high coverage.
3. **Contract Tests** – JSON-RPC schema checks and snapshots.
4. **Integration Tests** – New Relic sandbox via containers.
5. **Workflow Tests** – execute YAML workflows end to end.
6. **AI Harness** – validate tool planning by Copilot and Claude.
7. **Performance Tests** – ensure latency and memory targets.
8. **Chaos Tests** – inject faults and schema drift.
9. **Security Tests** – RBAC and basic vulnerability scanning.
10. **Docs Sync** – doctest for Markdown examples.

See `tests/` for layout and `Makefile` targets for running suites.

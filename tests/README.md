# Test Suite

This directory organizes all test assets. Each subfolder targets a specific layer of the taxonomy:

- `unit/` – fast function tests
- `contract/` – JSON-RPC schemas and snapshots
- `integration/` – real services via containers
- `workflow/` – YAML workflow harness
- `ai-harness/` – Copilot & Claude checks
- `perf/` – load tests
- `chaos/` – fault injection
- `docs-examples/` – doctest of documentation

Golden fixtures live in `../testdata`.

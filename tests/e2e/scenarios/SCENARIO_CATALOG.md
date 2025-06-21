# E2E Test Scenario Catalog

This catalog provides an overview of all E2E test scenarios for the MCP Server New Relic project.

## Scenario Categories

### 1. Discovery & Schema Drift (DISC-*)
Tests that validate discovery-first philosophy and adaptive behavior when schemas change.

| ID | Title | Description | Priority | Complexity |
|----|-------|-------------|----------|------------|
| DISC-MISS-001 | Discovery handles missing attributes gracefully | Tests adaptive query building when expected attributes are missing | Critical | High |
| DISC-DRIFT-001 | Schema evolution handling | Tests server behavior when event schemas change over time | High | Medium |
| DISC-MULTI-001 | Multi-account discovery consolidation | Tests cross-account schema discovery and merging | High | High |

### 2. Incident Response (INC-*)
Complex scenarios simulating real-world incident investigation workflows.

| ID | Title | Description | Priority | Complexity |
|----|-------|-------------|----------|------------|
| INC-SQL-404 | SQL database 404 error investigation | Multi-tool orchestration for database connectivity issues | Critical | Very High |
| INC-SPIKE-001 | Traffic spike root cause analysis | Correlates sudden traffic increase with infrastructure | High | High |
| INC-CASCADE-001 | Cascading failure investigation | Traces failures across multiple services | High | Very High |

### 3. Performance Analysis (PERF-*)
Performance comparison and optimization scenarios.

| ID | Title | Description | Priority | Complexity |
|----|-------|-------------|----------|------------|
| PERF-CMP-001 | Cross-region latency comparison | Compares performance across regions with chaos testing | Critical | Very High |
| PERF-TREND-001 | Long-term performance trend analysis | Analyzes performance degradation over time | Medium | Medium |
| PERF-OPT-001 | Query performance optimization | Tests query optimization recommendations | Medium | High |

### 4. Governance & Compliance (GOV-*)
Cost optimization and compliance validation scenarios.

| ID | Title | Description | Priority | Complexity |
|----|-------|-------------|----------|------------|
| GOV-COST-001 | Cost optimization analysis | Identifies high-cost queries and data retention issues | Critical | High |
| GOV-COMPL-001 | Compliance audit workflow | Validates SOC2/GDPR compliance checks | High | Medium |
| GOV-USAGE-001 | Resource usage governance | Tracks dashboard/alert proliferation | Medium | Medium |

### 5. Chaos & Resilience (CHAOS-*)
Network chaos and failure testing scenarios.

| ID | Title | Description | Priority | Complexity |
|----|-------|-------------|----------|------------|
| CHAOS-NET-001 | Network chaos resilience | Tests retry, circuit breaker, and degradation | Critical | High |
| CHAOS-RATE-001 | Rate limit stress testing | Validates rate limiting under load | High | Medium |
| CHAOS-REGION-001 | Region failover testing | Tests automatic region switching | High | High |

### 6. Data Quality (DQ-*)
Data validation and quality assurance scenarios.

| ID | Title | Description | Priority | Complexity |
|----|-------|-------------|----------|------------|
| DQ-VALIDATION-001 | Data type validation | Ensures correct data type handling | High | Low |
| DQ-MISSING-001 | Missing data handling | Tests behavior with incomplete data | High | Medium |
| DQ-CORRUPT-001 | Corrupted data resilience | Handles malformed event data | Medium | Medium |

### 7. Integration (INT-*)
Third-party integration and API compatibility tests.

| ID | Title | Description | Priority | Complexity |
|----|-------|-------------|----------|------------|
| INT-MCP-001 | MCP protocol compliance | Validates full MCP protocol implementation | Critical | Medium |
| INT-NERDGRAPH-001 | NerdGraph API coverage | Tests all NerdGraph endpoints used | High | High |
| INT-MULTI-001 | Multi-tool workflow | Complex workflows using 10+ tools | High | Very High |

## Scenario Complexity Levels

- **Low**: Single tool, simple assertions, no setup required
- **Medium**: 2-5 tools, moderate setup, basic assertions
- **High**: 5-10 tools, complex setup, multiple assertion types
- **Very High**: 10+ tools, chaos testing, cross-region, complex assertions

## Scenario Tags

- `critical`: Must pass for release
- `discovery`: Tests discovery-first philosophy
- `chaos`: Includes network chaos testing
- `multi-account`: Requires multiple New Relic accounts
- `cross-region`: Tests US/EU region behavior
- `performance`: Performance testing scenarios
- `governance`: Cost and compliance testing
- `incident-response`: Incident investigation workflows
- `complex-workflow`: Multi-step orchestration

## Execution Requirements

### Test Accounts
- Primary Account: Full data, high cardinality
- Secondary Account: Limited data, different schema
- Empty Account: No data for negative testing
- High-Cardinality Account: Stress testing

### Infrastructure
- Toxiproxy: For chaos testing
- Multiple regions: US and EU endpoints
- Test data seeding: Python scripts in `scripts/`

### Environment Variables
```bash
# Required
E2E_PRIMARY_ACCOUNT_ID=xxx
E2E_PRIMARY_API_KEY=xxx
E2E_SECONDARY_ACCOUNT_ID=xxx
E2E_SECONDARY_API_KEY=xxx

# Optional
E2E_EU_ACCOUNT_ID=xxx
E2E_EU_API_KEY=xxx
E2E_EMPTY_ACCOUNT_ID=xxx
E2E_EMPTY_API_KEY=xxx
```

## Adding New Scenarios

1. Choose appropriate category prefix (DISC-, INC-, etc.)
2. Follow naming convention: `{PREFIX}-{DESCRIPTOR}-{NUMBER}`
3. Use YAML format matching the schema
4. Include all required fields
5. Add to this catalog
6. Create seed data script if needed
7. Test locally before committing

## Scenario Prioritization

1. **Critical**: Core functionality, must pass
2. **High**: Important features, should pass
3. **Medium**: Nice to have, can fail temporarily
4. **Low**: Experimental or future features

## Test Execution Patterns

### Smoke Tests (5 min)
- DISC-MISS-001
- INC-SQL-404 (simplified)
- GOV-COST-001 (basic)

### Full Regression (2 hours)
- All Critical priority scenarios
- All High priority scenarios
- Selected Medium scenarios

### Chaos Testing (30 min)
- CHAOS-NET-001
- PERF-CMP-001 (with chaos)
- CHAOS-REGION-001

### Performance Suite (1 hour)
- All PERF-* scenarios
- Selected INC-* with performance focus
- GOV-COST-001 (performance impact)
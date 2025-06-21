# E2E Test Credentials Template

To run comprehensive E2E tests, set these environment variables:

## Required (Minimum)
```bash
export E2E_ACCOUNT_LEGACY_APM=your_account_id_here
export E2E_API_KEY_LEGACY=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

## Optional (for comprehensive testing)
```bash
# Modern OpenTelemetry Account
export E2E_ACCOUNT_MODERN_OTEL=your_otel_account_id
export E2E_API_KEY_OTEL=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

# Mixed Data Account  
export E2E_ACCOUNT_MIXED_DATA=your_mixed_account_id
export E2E_API_KEY_MIXED=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

# Sparse Data Account (optional)
export E2E_ACCOUNT_SPARSE_DATA=your_sparse_account_id  
export E2E_API_KEY_SPARSE=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

# EU Region Account (optional)
export E2E_ACCOUNT_EU_REGION=your_eu_account_id
export E2E_API_KEY_EU=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

## Test Commands

```bash
# Quick test (requires only legacy APM account)
npm run test:e2e:quick

# Full test suite (requires multiple accounts)
npm run test:e2e

# Specific test categories
npm run test:e2e:discovery
npm run test:e2e:tools  
npm run test:e2e:dashboards
npm run test:e2e:performance
```

## API Key Requirements

API keys need these permissions:
- NRQL Query access
- Entity search access
- Dashboard read access (for validation)

## Account Requirements

**Legacy APM Account**:
- Should have APM applications with Transaction events
- At least 1000 transactions per hour for meaningful testing
- Should have error data for error indicator testing

**Modern OpenTelemetry Account**:
- Should have OTEL spans and dimensional metrics  
- Should use `service.name` as service identifier
- Should have span data with duration metrics

**Mixed Data Account**:
- Should have diverse telemetry (APM + Infrastructure + Browser + Logs)
- Good for testing adaptability across data patterns

## Testing with Demo Data

If you don't have access to real accounts, you can test with:
1. New Relic's demo data (if available)
2. A personal account with the New Relic agent installed on a simple application
3. Mock tests (coming soon - will test logic without real API calls)
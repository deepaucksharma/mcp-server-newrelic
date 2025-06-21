import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    // E2E test configuration
    name: 'e2e',
    include: ['tests/e2e/**/*.test.ts'],
    exclude: ['tests/unit/**/*', 'tests/integration/**/*'],
    
    // Extended timeouts for real API calls
    testTimeout: 120000, // 2 minutes per test
    hookTimeout: 30000,  // 30 seconds for setup/teardown
    
    // Global setup and teardown
    globalSetup: ['tests/e2e/global-setup.ts'],
    
    // Environment configuration
    env: {
      NODE_ENV: 'test',
      E2E_MODE: 'true',
      // Cache configuration for testing
      E2E_CACHE_TTL_SECONDS: '3600',
      E2E_TIMEOUT_MS: '120000',
      E2E_PARALLEL_ACCOUNTS: '3',
    },
    
    // Sequential execution for E2E tests to avoid rate limiting
    pool: 'forks',
    poolOptions: {
      forks: {
        singleFork: true, // Run tests sequentially to avoid overwhelming APIs
      },
    },
    
    // Detailed reporting for E2E results
    reporter: ['verbose', 'json'],
    outputFile: {
      json: 'tests/e2e/results/e2e-results.json',
    },
    
    // Coverage configuration (optional for E2E)
    coverage: {
      enabled: false, // E2E tests focus on integration, not coverage
    },
    
    // Test isolation
    isolate: true,
    
    // Retry configuration for flaky network conditions
    retry: 1,
    
    // Bail on first failure for quick feedback
    bail: process.env.E2E_BAIL === 'true' ? 1 : 0,
    
    // Setup files
    setupFiles: ['tests/e2e/setup.ts'],
  },
  
  // TypeScript configuration
  esbuild: {
    target: 'node20',
  },
  
  // Define globals for test files
  define: {
    'process.env.E2E_MODE': 'true',
  },
});
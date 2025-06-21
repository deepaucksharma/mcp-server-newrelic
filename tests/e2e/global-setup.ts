/**
 * Global E2E Test Setup
 * 
 * Validates test environment and prepares for comprehensive testing
 */

export async function setup() {
  console.log('🚀 Starting E2E Test Suite Global Setup...');
  
  // Validate test environment variables
  const requiredEnvVars = [
    'E2E_ACCOUNT_LEGACY_APM',
    'E2E_API_KEY_LEGACY',
  ];
  
  const optionalEnvVars = [
    'E2E_ACCOUNT_MODERN_OTEL',
    'E2E_API_KEY_OTEL',
    'E2E_ACCOUNT_MIXED_DATA',
    'E2E_API_KEY_MIXED',
    'E2E_ACCOUNT_SPARSE_DATA',
    'E2E_API_KEY_SPARSE',
    'E2E_ACCOUNT_EU_REGION',
    'E2E_API_KEY_EU',
  ];
  
  // Check required environment variables
  const missingRequired = requiredEnvVars.filter(varName => !process.env[varName]);
  if (missingRequired.length > 0) {
    console.error('❌ Missing required environment variables:');
    missingRequired.forEach(varName => {
      console.error(`   - ${varName}`);
    });
    console.error('\nPlease set the required environment variables for E2E testing.');
    console.error('See tests/e2e/README.md for configuration details.');
    throw new Error(`Missing required environment variables: ${missingRequired.join(', ')}`);
  }
  
  // Report available test accounts
  console.log('✅ Required environment variables found');
  
  const availableOptional = optionalEnvVars.filter(varName => process.env[varName]);
  if (availableOptional.length > 0) {
    console.log('📊 Optional test accounts available:');
    availableOptional.forEach(varName => {
      const accountType = varName.replace('E2E_', '').replace('_API_KEY', '').replace('_ACCOUNT', '');
      console.log(`   - ${accountType}`);
    });
  }
  
  // Validate account IDs are numeric
  const accountIdVars = requiredEnvVars.concat(optionalEnvVars).filter(v => v.includes('ACCOUNT'));
  for (const varName of accountIdVars) {
    const value = process.env[varName];
    if (value && (isNaN(parseInt(value)) || parseInt(value) <= 0)) {
      throw new Error(`Invalid account ID for ${varName}: ${value}`);
    }
  }
  
  // Set default test configuration
  process.env.E2E_TIMEOUT_MS = process.env.E2E_TIMEOUT_MS || '120000';
  process.env.E2E_PARALLEL_ACCOUNTS = process.env.E2E_PARALLEL_ACCOUNTS || '3';
  process.env.E2E_CACHE_TTL_SECONDS = process.env.E2E_CACHE_TTL_SECONDS || '3600';
  
  console.log('⚙️  Test configuration:');
  console.log(`   - Timeout: ${process.env.E2E_TIMEOUT_MS}ms`);
  console.log(`   - Parallel accounts: ${process.env.E2E_PARALLEL_ACCOUNTS}`);
  console.log(`   - Cache TTL: ${process.env.E2E_CACHE_TTL_SECONDS}s`);
  
  // Create results directory
  const fs = await import('fs');
  const path = await import('path');
  const resultsDir = path.join(process.cwd(), 'tests/e2e/results');
  
  if (!fs.existsSync(resultsDir)) {
    fs.mkdirSync(resultsDir, { recursive: true });
    console.log('📁 Created results directory:', resultsDir);
  }
  
  console.log('✅ Global E2E setup completed successfully\n');
}

export async function teardown() {
  console.log('\n🧹 E2E Test Suite Global Teardown...');
  
  // Clean up any global resources if needed
  console.log('✅ Global teardown completed');
}
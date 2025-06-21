#!/usr/bin/env node

/**
 * Basic test to check if our modules can be imported and basic functionality works
 */

import { createServer } from './dist/index.js';

async function testBasic() {
  console.log('🧪 Basic Import Test');
  
  try {
    // Test if we can import the server
    console.log('✅ Successfully imported createServer');
    
    // Set dummy environment variables
    process.env.NEW_RELIC_API_KEY = 'dummy-key';
    process.env.NEW_RELIC_ACCOUNT_ID = '123456';
    
    // Try to create a server (this should work even with dummy credentials)
    const server = await createServer();
    console.log('✅ Successfully created server instance');
    
    // Test if the methods we need for testing exist
    const discovery = server.getDiscovery();
    console.log('✅ Successfully got discovery engine');
    
    const dashboardGenerator = server.getDashboardGenerator();
    console.log('✅ Successfully got dashboard generator');
    
    // Shutdown the server
    await server.shutdown();
    console.log('✅ Successfully shut down server');
    
    console.log('\n🎉 All basic tests passed!');
    
  } catch (error) {
    console.error('❌ Basic test failed:', error.message);
    console.error(error.stack);
    process.exit(1);
  }
}

testBasic().catch(error => {
  console.error('❌ Unhandled error:', error);
  process.exit(1);
});
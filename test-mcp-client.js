#!/usr/bin/env node

import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';
import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));

// Check environment variables
const requiredEnvVars = ['NEW_RELIC_API_KEY', 'NEW_RELIC_ACCOUNT_ID'];
const missing = requiredEnvVars.filter(key => !process.env[key]);

if (missing.length > 0) {
  console.error('❌ Missing required environment variables:', missing);
  process.exit(1);
}

async function testMCPServer() {
  console.log('🚀 Starting MCP Server New Relic test...\n');

  // Spawn the server process
  const serverProcess = spawn('node', [join(__dirname, 'dist', 'index.js')], {
    env: process.env,
    stdio: ['pipe', 'pipe', 'inherit']
  });

  // Create MCP client
  const client = new Client({
    name: 'test-client',
    version: '1.0.0'
  }, {
    capabilities: {}
  });

  // Create transport
  const transport = new StdioClientTransport({
    command: 'node',
    args: [join(__dirname, 'dist', 'index.js')],
    env: process.env
  });

  try {
    // Connect to server
    await client.connect(transport);
    console.log('✅ Connected to MCP server\n');

    // List available tools
    console.log('📋 Available tools:');
    const tools = await client.listTools();
    tools.tools.forEach(tool => {
      console.log(`  - ${tool.name}`);
    });
    console.log('');

    // Test discover_schemas tool
    console.log('🔍 Testing discover_schemas tool...');
    const schemaResult = await client.callTool('discover_schemas', {
      account_id: parseInt(process.env.NEW_RELIC_ACCOUNT_ID),
      include_attributes: false,
      include_metrics: true
    });
    
    console.log('✅ Schema discovery result:');
    const schemaData = JSON.parse(schemaResult.content[0].text);
    console.log(`  - Event types found: ${schemaData.event_types?.length || 0}`);
    console.log(`  - Metrics found: ${schemaData.metrics?.length || 0}`);
    if (schemaData.summary) {
      console.log('  - Summary:', schemaData.summary);
    }
    console.log('');

    // Test environment discovery
    console.log('🌍 Testing discover.environment tool...');
    const envResult = await client.callTool('discover.environment', {
      includeHealth: false,
      maxEntities: 10
    });
    
    console.log('✅ Environment discovery complete');
    console.log('');

    // Clean shutdown
    await client.close();
    serverProcess.kill();
    console.log('✅ Test completed successfully!');

  } catch (error) {
    console.error('❌ Test failed:', error.message);
    serverProcess.kill();
    process.exit(1);
  }
}

// Run the test
testMCPServer().catch(console.error);
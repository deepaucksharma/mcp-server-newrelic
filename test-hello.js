import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

async function testHelloWorld() {
  const transport = new StdioClientTransport({
    command: 'node',
    args: ['./dist/index.js'],
    env: {
      ...process.env,
      NEW_RELIC_API_KEY: process.env.NEW_RELIC_API_KEY || 'your-api-key-here'
    }
  });

  const client = new Client({
    name: 'test-client',
    version: '1.0.0'
  }, {
    capabilities: {}
  });

  try {
    await client.connect(transport);
    console.log('✅ Connected to MCP server');

    // List available tools
    const tools = await client.listTools();
    console.log('\n📋 Available tools:');
    console.log(JSON.stringify(tools, null, 2));

    // Call hello_newrelic with a test account ID
    console.log('\n🔧 Calling hello_newrelic tool...');
    try {
      const result = await client.callTool('hello_newrelic', {
        account_id: 1234567  // Replace with a valid account ID
      });

      console.log('\n📊 Result:');
      console.log(JSON.stringify(result, null, 2));
    } catch (toolError) {
      console.log('\n⚠️  Tool call failed (expected with test API key):');
      console.log('Error:', toolError.message);
    }

  } catch (error) {
    console.error('❌ Error:', error.message);
  } finally {
    await client.close();
    console.log('\n👋 Connection closed');
  }
}

// Run the test
console.log('🚀 Starting MCP Server New Relic test...\n');
testHelloWorld().catch(console.error);
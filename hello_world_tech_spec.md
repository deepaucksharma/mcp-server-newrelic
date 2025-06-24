# MCP Server New Relic: Hello World Technical Specification

## Document Information
- **Version**: 0.1.0 (Hello World)
- **Status**: Implementation Ready
- **Created**: 2025-06-24
- **Scope**: Bare Minimum Viable Implementation

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Hello World Scope](#hello-world-scope)
3. [Minimal Architecture](#minimal-architecture)
4. [Implementation Details](#implementation-details)
5. [API Specification](#api-specification)
6. [Setup & Configuration](#setup--configuration)
7. [Testing the Hello World](#testing-the-hello-world)
8. [Next Steps](#next-steps)

## Executive Summary

### Purpose
This Hello World version demonstrates the absolute minimum viable MCP server that:
- Establishes MCP protocol communication
- Connects to New Relic NerdGraph API
- Implements ONE simple tool that proves the connection works
- Returns real data from a New Relic account

### What This Version Does
- Implements MCP server basics
- Provides a single tool: `hello_newrelic`
- Queries account information via NerdGraph
- Returns formatted account details

### What This Version Does NOT Do
- No schema discovery
- No caching
- No advanced tools
- No dashboard generation
- No error handling beyond basics
- No performance optimization

## Hello World Scope

### Single Tool Implementation
```typescript
{
  name: "hello_newrelic",
  description: "Test connection to New Relic and return account info",
  inputSchema: {
    type: "object",
    properties: {
      account_id: {
        type: "number",
        description: "New Relic account ID to query"
      }
    },
    required: ["account_id"]
  }
}
```

### Minimal Dependencies
```json
{
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.0.0",
    "node-fetch": "^3.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  }
}
```

## Minimal Architecture

### File Structure
```
mcp-server-newrelic/
├── package.json
├── tsconfig.json
├── src/
│   ├── index.ts          # Main entry point
│   ├── server.ts         # MCP server setup
│   ├── nerdgraph.ts      # Minimal NerdGraph client
│   └── tools/
│       └── hello.ts      # Hello world tool
├── .env                  # API key configuration
└── README.md            # Quick start guide
```

### Component Diagram
```
┌─────────────────┐
│   AI Assistant  │
└────────┬────────┘
         │ MCP Protocol
┌────────▼────────┐
│   MCP Server    │
├─────────────────┤
│ - Protocol Init │
│ - Tool Registry │
│ - Hello Tool    │
└────────┬────────┘
         │ GraphQL
┌────────▼────────┐
│ New Relic API   │
└─────────────────┘
```

## Implementation Details

### 1. Entry Point (index.ts)
```typescript
#!/usr/bin/env node
import { MCPServer } from './server';

async function main() {
  const server = new MCPServer();
  await server.start();
}

main().catch(console.error);
```

### 2. MCP Server Setup (server.ts)
```typescript
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { HelloTool } from './tools/hello';
import { NerdGraphClient } from './nerdgraph';

export class MCPServer {
  private server: Server;
  private nerdgraph: NerdGraphClient;

  constructor() {
    this.server = new Server(
      {
        name: 'mcp-server-newrelic',
        version: '0.1.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    // Initialize NerdGraph client
    const apiKey = process.env.NEW_RELIC_API_KEY;
    if (!apiKey) {
      throw new Error('NEW_RELIC_API_KEY environment variable is required');
    }
    this.nerdgraph = new NerdGraphClient(apiKey);

    // Register the hello world tool
    this.registerTools();
  }

  private registerTools() {
    const helloTool = new HelloTool(this.nerdgraph);
    
    this.server.setRequestHandler('tools/list', async () => ({
      tools: [helloTool.getDefinition()],
    }));

    this.server.setRequestHandler('tools/call', async (request) => {
      if (request.params.name === 'hello_newrelic') {
        return helloTool.execute(request.params.arguments);
      }
      throw new Error(`Unknown tool: ${request.params.name}`);
    });
  }

  async start() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    console.error('MCP Server New Relic (Hello World) started');
  }
}
```

### 3. Minimal NerdGraph Client (nerdgraph.ts)
```typescript
import fetch from 'node-fetch';

export class NerdGraphClient {
  private apiKey: string;
  private endpoint = 'https://api.newrelic.com/graphql';

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async query(query: string, variables?: Record<string, any>): Promise<any> {
    const response = await fetch(this.endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'API-Key': this.apiKey,
      },
      body: JSON.stringify({
        query,
        variables,
      }),
    });

    if (!response.ok) {
      throw new Error(`NerdGraph API error: ${response.statusText}`);
    }

    const data = await response.json();
    
    if (data.errors) {
      throw new Error(`GraphQL errors: ${JSON.stringify(data.errors)}`);
    }

    return data.data;
  }
}
```

### 4. Hello World Tool (tools/hello.ts)
```typescript
import { NerdGraphClient } from '../nerdgraph';

export class HelloTool {
  constructor(private nerdgraph: NerdGraphClient) {}

  getDefinition() {
    return {
      name: 'hello_newrelic',
      description: 'Test connection to New Relic and return account information',
      inputSchema: {
        type: 'object',
        properties: {
          account_id: {
            type: 'number',
            description: 'New Relic account ID to query',
          },
        },
        required: ['account_id'],
      },
    };
  }

  async execute(args: any) {
    const { account_id } = args;

    // Simple GraphQL query to get account info
    const query = `
      query GetAccountInfo($accountId: Int!) {
        actor {
          account(id: $accountId) {
            id
            name
            nrql(query: "SELECT count(*) FROM Transaction SINCE 1 hour ago") {
              results
            }
          }
        }
      }
    `;

    try {
      const result = await this.nerdgraph.query(query, { accountId: account_id });
      
      const account = result.actor.account;
      const transactionCount = account.nrql.results[0]?.count || 0;

      return {
        content: [
          {
            type: 'text',
            text: `Hello from New Relic! 👋\n\nAccount Information:\n- ID: ${account.id}\n- Name: ${account.name}\n- Transactions (last hour): ${transactionCount}`,
          },
        ],
      };
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error connecting to New Relic: ${error.message}`,
          },
        ],
        isError: true,
      };
    }
  }
}
```

### 5. TypeScript Configuration (tsconfig.json)
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "Node16",
    "lib": ["ES2022"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "moduleResolution": "Node16",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

### 6. Package Configuration (package.json)
```json
{
  "name": "mcp-server-newrelic",
  "version": "0.1.0",
  "description": "Hello World MCP Server for New Relic",
  "main": "dist/index.js",
  "type": "module",
  "scripts": {
    "build": "tsc",
    "start": "node dist/index.js",
    "dev": "tsc && node dist/index.js"
  },
  "keywords": ["mcp", "newrelic"],
  "author": "",
  "license": "MIT",
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.0.0",
    "node-fetch": "^3.3.2"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0"
  },
  "bin": {
    "mcp-server-newrelic": "./dist/index.js"
  }
}
```

## API Specification

### Tool: hello_newrelic

#### Request
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "hello_newrelic",
    "arguments": {
      "account_id": 1234567
    }
  },
  "id": 1
}
```

#### Response (Success)
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Hello from New Relic! 👋\n\nAccount Information:\n- ID: 1234567\n- Name: My Account\n- Transactions (last hour): 45231"
      }
    ]
  },
  "id": 1
}
```

#### Response (Error)
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Error connecting to New Relic: Invalid API key"
      }
    ],
    "isError": true
  },
  "id": 1
}
```

## Setup & Configuration

### 1. Environment Setup
Create a `.env` file:
```bash
NEW_RELIC_API_KEY=your-api-key-here
```

### 2. Installation Steps
```bash
# Clone the repository
git clone https://github.com/yourusername/mcp-server-newrelic.git
cd mcp-server-newrelic

# Install dependencies
npm install

# Build the TypeScript code
npm run build

# Test the server
npm start
```

### 3. MCP Client Configuration
Add to your MCP client configuration:
```json
{
  "mcpServers": {
    "newrelic": {
      "command": "node",
      "args": ["/path/to/mcp-server-newrelic/dist/index.js"],
      "env": {
        "NEW_RELIC_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

## Testing the Hello World

### 1. Manual Test Script
Create `test-hello.js`:
```javascript
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';
import { spawn } from 'child_process';

async function testHelloWorld() {
  const transport = new StdioClientTransport({
    command: 'node',
    args: ['./dist/index.js'],
    env: {
      ...process.env,
      NEW_RELIC_API_KEY: 'your-api-key-here'
    }
  });

  const client = new Client({
    name: 'test-client',
    version: '1.0.0'
  }, {
    capabilities: {}
  });

  await client.connect(transport);

  // List available tools
  const tools = await client.request('tools/list', {});
  console.log('Available tools:', tools);

  // Call hello_newrelic
  const result = await client.request('tools/call', {
    name: 'hello_newrelic',
    arguments: {
      account_id: 1234567
    }
  });

  console.log('Result:', result);

  await client.close();
}

testHelloWorld().catch(console.error);
```

### 2. Expected Output
```
Available tools: {
  tools: [
    {
      name: 'hello_newrelic',
      description: 'Test connection to New Relic and return account information',
      inputSchema: { ... }
    }
  ]
}

Result: {
  content: [
    {
      type: 'text',
      text: 'Hello from New Relic! 👋\n\nAccount Information:\n- ID: 1234567\n- Name: Production Account\n- Transactions (last hour): 125432'
    }
  ]
}
```

### 3. Verification Checklist
- [ ] Server starts without errors
- [ ] MCP client can connect
- [ ] Tool appears in tools/list
- [ ] Tool executes successfully
- [ ] Returns real New Relic data
- [ ] Error handling works for invalid API key
- [ ] Error handling works for invalid account ID

## Next Steps

### From Hello World to MVP
1. **Add More Basic Tools**
   - `list_accounts`: List all accessible accounts
   - `simple_query`: Execute a basic NRQL query
   - `get_entity`: Fetch entity details

2. **Improve Error Handling**
   - Validate API key format
   - Handle network errors gracefully
   - Add retry logic

3. **Add Basic Caching**
   - In-memory cache for account info
   - 5-minute TTL

4. **Enhance Tool Responses**
   - Add structured data formats
   - Include metadata

### Path to Full Implementation
1. **Phase 1**: Expand to 5-10 essential tools
2. **Phase 2**: Add discovery engine foundation
3. **Phase 3**: Implement caching layer
4. **Phase 4**: Add schema discovery
5. **Phase 5**: Dashboard generation
6. **Phase 6**: Full feature parity with spec

### Development Tips
1. **Keep It Simple**: Don't add features until the basics work perfectly
2. **Test Often**: Verify each addition with the MCP client
3. **Log Everything**: Add debug logging for troubleshooting
4. **Document Changes**: Update this spec as you expand

## Troubleshooting

### Common Issues

1. **"Cannot find module" Error**
   - Run `npm run build` before starting
   - Check tsconfig.json module settings

2. **"API Key Required" Error**
   - Ensure .env file exists
   - Check environment variable name

3. **"GraphQL Error" Response**
   - Verify API key has correct permissions
   - Check account ID is valid

4. **Connection Timeout**
   - Test New Relic API separately
   - Check network connectivity

### Debug Mode
Add logging to server.ts:
```typescript
console.error('Starting server with API key:', apiKey.substring(0, 8) + '...');
console.error('Received request:', request.method);
```

## Success Criteria

### Hello World is Complete When:
1. ✅ MCP server starts and accepts connections
2. ✅ Single tool is registered and callable
3. ✅ Tool connects to New Relic API
4. ✅ Returns real account data
5. ✅ Basic error handling works
6. ✅ Can be installed and run by others

### Performance Baseline
- Startup time: <1 second
- Tool response time: <3 seconds
- Memory usage: <50MB

---

This Hello World specification provides the absolute minimum foundation for the MCP Server New Relic. It proves the concept while keeping complexity to an absolute minimum, allowing for iterative development toward the full specification.
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { ListToolsRequestSchema, CallToolRequestSchema } from '@modelcontextprotocol/sdk/types.js';
import { HelloTool } from './tools/hello.js';
import { NerdGraphClient } from './nerdgraph.js';

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
    
    this.server.setRequestHandler(ListToolsRequestSchema, async () => ({
      tools: [helloTool.getDefinition()],
    }));

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
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
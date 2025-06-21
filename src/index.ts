/**
 * MCP Server New Relic - Platform-Native Implementation
 * 
 * Zero Hardcoded Schemas with Enhanced Existing Tools
 * Built with official MCP TypeScript SDK + Discovery-First Intelligence
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { z } from 'zod';

import { PlatformDiscovery } from './core/platform-discovery.js';
import { createNerdGraphClient } from './adapters/nerdgraph.js';
import { EnhancedToolRegistry } from './tools/enhanced-registry.js';
import { AdaptiveDashboardGenerator } from './tools/adaptive-dashboards.js';
import { ConfigSchema, Logger } from './core/types.js';

/**
 * Platform-Native MCP Server with Discovery Intelligence
 */
export class MCPNewRelicServer {
  private server: Server;
  private discovery!: PlatformDiscovery;
  private dashboardGenerator!: AdaptiveDashboardGenerator;
  private config: z.infer<typeof ConfigSchema>;
  private logger!: Logger;

  constructor(config: z.infer<typeof ConfigSchema>) {
    this.config = ConfigSchema.parse(config);
    
    this.server = new Server(
      {
        name: 'mcp-server-newrelic-platform',
        version: '2.0.0',
      },
      {
        capabilities: {
          tools: {},
          resources: {},
        },
      }
    );

    this.setupPlatformServices();
    this.setupHandlers();
  }

  /**
   * Setup platform services - discovery engine and dashboard generator
   */
  private setupPlatformServices(): void {
    // Create logger
    this.logger = this.createLogger();

    // Create NerdGraph client
    const nerdgraph = createNerdGraphClient({
      apiKey: this.config.newrelic.apiKey,
      region: this.config.newrelic.region,
      logger: this.logger,
    });

    // Initialize platform discovery
    this.discovery = new PlatformDiscovery(nerdgraph, this.logger);

    // Initialize adaptive dashboard generator
    this.dashboardGenerator = new AdaptiveDashboardGenerator(this.discovery, this.logger);
    
    // Note: dashboardGenerator will be used by enhanced tools for dashboard creation

    this.logger.info('Platform services initialized', {
      region: this.config.newrelic.region,
      accountId: this.config.newrelic.accountId,
    });
  }

  /**
   * Setup MCP protocol handlers with enhanced tool registry
   */
  private setupHandlers(): void {
    // Create enhanced tool registry with discovery intelligence
    const config = {
      newrelic: {
        apiKey: this.config.newrelic.apiKey,
        accountId: this.config.newrelic.accountId,
        region: this.config.newrelic.region,
      }
    };

    new EnhancedToolRegistry(this.server, this.discovery, config);

    this.logger.info('Enhanced tools registered with platform discovery');
  }

  /**
   * Get dashboard generator instance (for enhanced tools)
   */
  getDashboardGenerator(): AdaptiveDashboardGenerator {
    return this.dashboardGenerator;
  }

  /**
   * Get discovery engine instance (for enhanced tools)
   */
  getDiscovery(): PlatformDiscovery {
    return this.discovery;
  }

  /**
   * Create logger instance
   */
  private createLogger(): Logger {
    return {
      info: (message: string, meta?: any) => {
        console.error(`[MCP-NR] INFO: ${message}`, meta ? JSON.stringify(meta) : '');
      },
      warn: (message: string, meta?: any) => {
        console.error(`[MCP-NR] WARN: ${message}`, meta ? JSON.stringify(meta) : '');
      },
      error: (message: string, meta?: any) => {
        console.error(`[MCP-NR] ERROR: ${message}`, meta ? JSON.stringify(meta) : '');
      },
      debug: (message: string, meta?: any) => {
        if (process.env['DEBUG']) {
          console.error(`[MCP-NR] DEBUG: ${message}`, meta ? JSON.stringify(meta) : '');
        }
      },
    };
  }

  /**
   * Start the MCP server
   */
  async start(): Promise<void> {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    
    this.logger.info('MCP New Relic Platform Server started', {
      version: '2.0.0',
      accountId: this.config.newrelic.accountId,
      region: this.config.newrelic.region,
      capabilities: ['tools', 'resources', 'discovery', 'adaptive-dashboards'],
    });
  }

  /**
   * Graceful shutdown
   */
  async shutdown(): Promise<void> {
    this.logger.info('Shutting down MCP server');
    await this.server.close();
  }
}

/**
 * Create server instance with platform-native configuration
 */
export async function createServer(): Promise<MCPNewRelicServer> {
  // Load configuration from environment
  const config = {
    newrelic: {
      apiKey: process.env['NEW_RELIC_API_KEY'] || '',
      accountId: process.env['NEW_RELIC_ACCOUNT_ID'] || '',
      region: (process.env['NEW_RELIC_REGION'] as 'US' | 'EU') || 'US',
      graphqlUrl: process.env['NEW_RELIC_REGION'] === 'EU' 
        ? 'https://api.eu.newrelic.com/graphql'
        : 'https://api.newrelic.com/graphql',
    },
    discovery: {
      cache: {
        type: 'memory' as const,
        ttl: {
          schemas: 4 * 60 * 60 * 1000,      // 4 hours - schemas change rarely
          attributes: 30 * 60 * 1000,       // 30 minutes - attributes change more often
          serviceId: 2 * 60 * 60 * 1000,    // 2 hours - service identifiers stable
          errors: 30 * 60 * 1000,           // 30 minutes - error patterns change
        },
      },
      confidence: {
        minimum: 0.7,
        optimal: 0.9,
      },
    },
    mcp: {
      transport: 'stdio' as const,
      http: {
        port: 3000,
        cors: ['*'],
      },
    },
    telemetry: {
      enabled: true,
    },
  };

  // Validate required environment variables
  if (!config.newrelic.apiKey) {
    throw new Error('NEW_RELIC_API_KEY environment variable is required');
  }

  if (!config.newrelic.accountId) {
    throw new Error('NEW_RELIC_ACCOUNT_ID environment variable is required');
  }

  return new MCPNewRelicServer(config);
}

/**
 * Main entry point
 */
async function main(): Promise<void> {
  try {
    const server = await createServer();
    
    // Handle graceful shutdown
    process.on('SIGINT', async () => {
      console.error('Received SIGINT, shutting down gracefully...');
      await server.shutdown();
      process.exit(0);
    });

    process.on('SIGTERM', async () => {
      console.error('Received SIGTERM, shutting down gracefully...');
      await server.shutdown();
      process.exit(0);
    });

    await server.start();

  } catch (error: any) {
    console.error('Failed to start MCP server:', error.message);
    if (error.message.includes('NEW_RELIC_')) {
      console.error('');
      console.error('Required environment variables:');
      console.error('  NEW_RELIC_API_KEY - Your New Relic API key');
      console.error('  NEW_RELIC_ACCOUNT_ID - Your New Relic account ID');
      console.error('  NEW_RELIC_REGION - US or EU (optional, default: US)');
    }
    process.exit(1);
  }
}

// Run the server if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(error => {
    console.error('Unhandled error:', error);
    process.exit(1);
  });
}

export default MCPNewRelicServer;
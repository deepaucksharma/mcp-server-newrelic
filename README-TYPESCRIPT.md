# New Relic MCP Server - TypeScript Implementation

> 🚀 **Status: Fresh Start with TypeScript**  
> This is a clean TypeScript implementation using the official MCP SDK.  
> Built on discovery-first principles without the complexity of the previous Go implementation.

## Quick Start

### Prerequisites
- Node.js 18+ 
- New Relic account and API key

### Setup

```bash
# Clone and install
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic
npm install

# Configure environment
cp .env.example .env
# Edit .env with your New Relic credentials

# Build and run
npm run build
npm run dev
```

### Claude Desktop Integration

Add to your Claude Desktop MCP settings:

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "node",
      "args": ["dist/index.js"],
      "cwd": "/path/to/mcp-server-newrelic",
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-YOUR-KEY",
        "NEW_RELIC_ACCOUNT_ID": "YOUR-ACCOUNT-ID"
      }
    }
  }
}
```

## Architecture Highlights

### Discovery-First Approach
```typescript
// Never assume data structure - always discover first
const schemas = await discoveryEngine.exploreSchemas();
const attributes = await discoveryEngine.exploreAttributes(eventType);
const query = buildInformedQuery(schemas, attributes, userIntent);
```

### TypeScript Benefits
- **Type Safety**: Catch errors at compile time
- **MCP Native**: Built with official MCP SDK
- **JSON Perfect**: Natural JSON handling for MCP protocol
- **Async Native**: Modern async/await patterns
- **Fast Iteration**: Quick development cycles

## Project Structure

```
src/
├── index.ts                     # Main MCP server
├── config/environment.ts        # Environment validation
├── services/newrelic-client.ts  # New Relic GraphQL client
├── tools/                       # MCP tool implementations
└── types/                       # TypeScript type definitions
```

## Implementation Plan

### Phase 1: Foundation ✅
- [x] TypeScript project setup
- [x] MCP SDK integration
- [x] Environment configuration
- [x] New Relic client with retries

### Phase 2: Core Tools 🚧
- [ ] Discovery tools (explore_event_types, explore_attributes)
- [ ] NRQL query execution with validation
- [ ] Basic error handling and logging

### Phase 3: Analysis Tools 📋
- [ ] Baseline calculation
- [ ] Anomaly detection
- [ ] Statistical analysis

### Phase 4: Advanced Features 📋
- [ ] Dashboard tools
- [ ] Alert management
- [ ] Workflow orchestration

## Why TypeScript?

The previous Go implementation suffered from over-engineering:
- 165 Go files with mostly mock functionality
- Complex build tag systems
- Multiple parallel server implementations
- Over-engineered infrastructure

TypeScript offers:
- **Simplicity**: Direct path from idea to implementation
- **Ecosystem**: Rich MCP and observability tooling
- **Maintainability**: Easier for contributors to understand
- **JSON Native**: Perfect for MCP's JSON-RPC protocol

## Development Commands

```bash
# Development
npm run dev              # Watch mode with tsx
npm run build           # Build for production
npm run type-check      # TypeScript validation

# Testing
npm test                # Run tests
npm run test:watch      # Watch mode testing

# Code Quality
npm run lint            # ESLint
```

## Contributing

This implementation prioritizes:
1. **Real functionality** over sophisticated mocks
2. **Type safety** with strict TypeScript
3. **Discovery-first** principles in every tool
4. **Simple, maintainable** code

See the comprehensive documentation in `docs/archive/` for the full vision and architecture specifications.

## License

MIT License - see LICENSE file for details
# Quick Start Guide

Get up and running with the New Relic MCP Server in 5 minutes!

## Prerequisites

- Go 1.21 or later
- New Relic account with API key
- Make (for build commands)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` and add your credentials:
```env
NEW_RELIC_API_KEY=your-api-key-here
NEW_RELIC_ACCOUNT_ID=your-account-id
```

### 3. Build the Server

```bash
make build
```

## Running the Server

### Production Mode (with New Relic connection)
```bash
make run
```

### Mock Mode (for testing without New Relic)
```bash
make run-mock
```

### Development Mode (with auto-reload)
```bash
make dev
```

## First Discovery

Try your first discovery-driven query:

```bash
# Using MCP Inspector
npx @modelcontextprotocol/inspector ./bin/mcp-server

# Or directly with JSON-RPC
echo '{"jsonrpc":"2.0","method":"discovery.explore_event_types","id":1}' | ./bin/mcp-server
```

## Integration with AI Assistants

### Claude Desktop

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/path/to/mcp-server-newrelic/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "your-key",
        "NEW_RELIC_ACCOUNT_ID": "your-account-id"
      }
    }
  }
}
```

## Next Steps

1. **Explore Discovery Tools**: Start with `discovery.explore_event_types` to see what data you have
2. **Read the Philosophy**: Understand our [discovery-first approach](./philosophy/NO_ASSUMPTIONS_MANIFESTO.md)
3. **Try Examples**: Check out [workflow examples](./examples/DISCOVERY_FIRST_WORKFLOWS.md)
4. **Join the Community**: Report issues at [GitHub](https://github.com/deepaucksharma/mcp-server-newrelic/issues)

## Troubleshooting

- **Connection Issues**: Run `make diagnose` to check your setup
- **Build Problems**: Ensure Go 1.21+ is installed
- **API Errors**: Verify your API key has the required permissions

For detailed troubleshooting, see the [Troubleshooting Guide](./guides/troubleshooting.md).
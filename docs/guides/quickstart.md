# Quick Start Guide

Get up and running with the New Relic MCP Server in 5 minutes.

## Prerequisites

- Go 1.22+ installed
- New Relic account with API key
- Git

## Installation

1. **Clone the repository**:
```bash
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic
```

2. **Set up environment**:
```bash
cp .env.example .env
# Edit .env with your New Relic credentials:
# - NEW_RELIC_API_KEY (required)
# - NEW_RELIC_ACCOUNT_ID (required)
# - NEW_RELIC_REGION (optional, defaults to US)
```

3. **Build the server**:
```bash
make build
```

4. **Run the server**:
```bash
# Production mode (requires New Relic credentials)
make run

# Mock mode (no New Relic connection needed)
make run-mock
```

## Using with AI Assistants

### Claude Desktop

1. Edit your Claude configuration:
```bash
# macOS
~/Library/Application Support/Claude/config.json

# Windows
%APPDATA%\Claude\config.json
```

2. Add the MCP server:
```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/path/to/mcp-server-newrelic/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "your-api-key",
        "NEW_RELIC_ACCOUNT_ID": "your-account-id"
      }
    }
  }
}
```

3. Restart Claude Desktop

### Command Line

```bash
# Direct execution
./bin/mcp-server

# With MCP inspector for testing
npx @modelcontextprotocol/inspector ./bin/mcp-server
```

## First Commands

Once connected, try these commands:

1. **Discover available data**:
   - "What event types are available in my New Relic account?"
   - "Show me the attributes for Transaction events"

2. **Run a simple query**:
   - "Show me the count of transactions in the last hour"
   - "What's the average duration of web transactions?"

3. **Explore your data**:
   - "Help me understand what data I have available"
   - "What are the most common event types?"

## Common Issues

- **No data returned**: Check your account has data in the selected time range
- **Authentication errors**: Verify your API key and account ID
- **Connection issues**: Try EU region if you're in Europe: `NEW_RELIC_REGION=EU`

## Next Steps

- Read the [Development Guide](development.md) for advanced usage
- Explore the [API Reference](../api/reference.md) for all available tools
- Check [Troubleshooting](troubleshooting.md) for common issues

## Getting Help

- Check existing issues: https://github.com/deepaucksharma/mcp-server-newrelic/issues
- Join the discussion: [Community Forum](https://forum.newrelic.com)
- Read the docs: [Full Documentation](../README.md)

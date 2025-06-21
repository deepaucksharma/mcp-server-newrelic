# Claude Desktop Integration Guide

This guide shows how to integrate the New Relic MCP Server with Claude Desktop for AI-powered observability workflows.

## Overview

The MCP Server provides Claude with intelligent access to your New Relic data through the Model Context Protocol. This enables natural language queries, automated analysis, and smart recommendations.

## Prerequisites

- Claude Desktop installed
- New Relic account with User API key
- MCP Server built and working

## Quick Setup

### 1. Build the Server

```bash
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic
make build-mcp
```

### 2. Test Server Works

```bash
# Test with mock mode first
./bin/mcp-server -mock

# Test with real credentials
export NEW_RELIC_API_KEY=NRAK-your-key
export NEW_RELIC_ACCOUNT_ID=your-account
./bin/mcp-server
```

### 3. Configure Claude Desktop

Open Claude Desktop settings and add MCP server configuration:

**On macOS**:
```bash
# Edit Claude Desktop config
code ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

**On Windows**:
```bash
# Edit Claude Desktop config
notepad %APPDATA%\Claude\claude_desktop_config.json
```

### 4. Add Server Configuration

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/absolute/path/to/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-your-api-key-here",
        "NEW_RELIC_ACCOUNT_ID": "your-account-id-here",
        "NEW_RELIC_REGION": "US"
      }
    }
  }
}
```

### 5. Restart Claude Desktop

Close and reopen Claude Desktop to load the new MCP server.

## Verification

### Check Server Connection

Ask Claude:
```
Are you connected to the New Relic MCP server? Can you list the available tools?
```

Expected response should mention tools like:
- query_nrdb
- discovery.explore_event_types
- discovery.explore_attributes
- create_alert
- list_alerts

### Test Basic Query

Ask Claude:
```
Can you show me the count of transactions in the last hour?
```

Claude should execute something like:
```sql
SELECT count(*) FROM Transaction SINCE 1 hour ago
```

## Natural Language Examples

### Data Exploration

- "What event types are available in my account?"
- "Show me all the attributes for Transaction events"
- "What applications am I monitoring?"

### Query Examples

- "How many transactions happened in the last hour?"
- "What's the average response time by application?"
- "Show me error rates for each service"
- "Find the slowest endpoints"

### Analysis Requests

- "Compare performance between last week and this week"
- "Find correlations between CPU usage and response time"
- "Identify anomalies in transaction volume"

### Alert Management

- "List all my alert conditions"
- "Create an alert for high error rates"
- "Show me recent alert incidents"

## Configuration Options

### Required Environment Variables

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/path/to/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-required",
        "NEW_RELIC_ACCOUNT_ID": "required"
      }
    }
  }
}
```

### Optional Environment Variables

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/path/to/bin/mcp-server", 
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-required",
        "NEW_RELIC_ACCOUNT_ID": "required",
        "NEW_RELIC_REGION": "US",
        "LOG_LEVEL": "info",
        "REQUEST_TIMEOUT": "30s",
        "MOCK_MODE": "false"
      }
    }
  }
}
```

### Mock Mode for Testing

```json
{
  "mcpServers": {
    "newrelic-mock": {
      "command": "/path/to/bin/mcp-server",
      "env": {
        "MOCK_MODE": "true"
      }
    }
  }
}
```

## What Works vs What's Mock

### ✅ Real Functionality

- **NRQL Query Execution** - `query_nrdb` executes real queries
- **Basic Discovery** - `discovery.explore_event_types`, `discovery.explore_attributes`
- **Simple Alerts** - `create_alert`, `list_alerts` with basic functionality
- **Session Management** - Stateful conversations

### 🟨 Mock-Only Features

When Claude uses these tools, they return realistic but fake data:

- **Advanced Analytics** - `analysis.*` tools return sophisticated fake results
- **Dashboard Creation** - Returns JSON but doesn't create real dashboards
- **Governance Tools** - Cost analysis, compliance reports (all fake)
- **Workflow Orchestration** - Complex multi-step operations (not implemented)

### ❌ Not Working

- **Advanced Discovery** - Relationship mapping, quality assessment
- **Multi-Account** - Only single account supported
- **Complex Workflows** - Automated investigation flows

## Troubleshooting

### Server Not Connecting

1. **Check server path**: Ensure absolute path in config
2. **Test server manually**: Run `./bin/mcp-server` directly
3. **Check permissions**: Ensure Claude can execute the binary
4. **View logs**: Look for error messages in Claude Desktop logs

### Authentication Errors

1. **API key format**: Must start with "NRAK-"
2. **Key type**: Must be User API key, not License key
3. **Account ID**: Verify correct account ID
4. **Region**: Set NEW_RELIC_REGION if using EU

### No Data Returned

1. **Test in New Relic UI**: Verify queries work in New Relic
2. **Check time ranges**: Use recent time windows
3. **Try mock mode**: Isolate server vs data issues
4. **Verify account**: Ensure account has data

## Best Practices

### Security
- Never commit API keys to version control
- Use environment-specific configurations
- Rotate API keys regularly
- Limit API key permissions

### Performance
- Start with broad queries, then narrow down
- Use appropriate time ranges
- Be aware of New Relic rate limits
- Cache results when possible (Claude session)

### Working with Claude
- Be specific about what you want to analyze
- Ask Claude to explain what tools it's using
- Verify important results manually
- Use mock mode for testing integrations

## Advanced Configuration

### Multiple Accounts

Currently not supported, but you can configure multiple servers:

```json
{
  "mcpServers": {
    "newrelic-prod": {
      "command": "/path/to/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-prod-key",
        "NEW_RELIC_ACCOUNT_ID": "prod-account"
      }
    },
    "newrelic-staging": {
      "command": "/path/to/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-staging-key", 
        "NEW_RELIC_ACCOUNT_ID": "staging-account"
      }
    }
  }
}
```

### Development Mode

```json
{
  "mcpServers": {
    "newrelic-dev": {
      "command": "/path/to/bin/mcp-server",
      "env": {
        "MOCK_MODE": "true",
        "LOG_LEVEL": "debug"
      }
    }
  }
}
```

The integration enables powerful observability workflows while being transparent about current implementation limitations.
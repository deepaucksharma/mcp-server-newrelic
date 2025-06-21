# 5-Minute Quickstart Guide

Get the New Relic MCP Server running and executing your first queries in under 5 minutes.

## Prerequisites

- Go 1.21+ or Docker
- New Relic User API Key (starts with "NRAK-")
- New Relic Account ID

## Option 1: Quick Start with Go

### Step 1: Clone and Build (2 minutes)

```bash
# Clone repository
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic

# Install dependencies and build
go mod download
make build-mcp
```

### Step 2: Configure (1 minute)

```bash
# Create configuration file
cat > .env << 'EOF'
NEW_RELIC_API_KEY=NRAK-your-api-key-here
NEW_RELIC_ACCOUNT_ID=your-account-id-here
NEW_RELIC_REGION=US
EOF
```

### Step 3: Test (1 minute)

```bash
# Test server starts
./bin/mcp-server

# In another terminal, test basic query
echo '{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_nrdb",
    "arguments": {
      "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
    }
  },
  "id": 1
}' | ./bin/mcp-server
```

### Step 4: Integrate with Claude (1 minute)

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/absolute/path/to/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-your-key",
        "NEW_RELIC_ACCOUNT_ID": "your-account"
      }
    }
  }
}
```

## Option 2: Quick Start with Mock Mode

### No Credentials Required

```bash
# Clone and build
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic
make build-mcp

# Run in mock mode
./bin/mcp-server -mock

# Test with realistic fake data
echo '{
  "jsonrpc": "2.0", 
  "method": "tools/call",
  "params": {
    "name": "query_nrdb",
    "arguments": {
      "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
    }
  },
  "id": 1
}' | ./bin/mcp-server -mock
```

## Verify Everything Works

### 1. Check Available Tools

```bash
echo '{
  "jsonrpc": "2.0",
  "method": "tools/list", 
  "id": 1
}' | ./bin/mcp-server
```

Should show tools including:
- query_nrdb
- discovery.explore_event_types
- discovery.explore_attributes

### 2. Test Discovery

```bash
echo '{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "discovery.explore_event_types",
    "arguments": {"limit": 5}
  },
  "id": 1
}' | ./bin/mcp-server
```

### 3. Test Query

```bash
echo '{
  "jsonrpc": "2.0",
  "method": "tools/call", 
  "params": {
    "name": "query_nrdb",
    "arguments": {
      "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
    }
  },
  "id": 1
}' | ./bin/mcp-server
```

## First Conversations with Claude

Once integrated with Claude Desktop:

### Conversation 1: Explore Your Data

**You**: "What kind of data do I have in New Relic?"

**Claude**: *Uses discovery tools to explore event types and attributes*

### Conversation 2: Basic Query

**You**: "How many transactions happened in the last hour?"

**Claude**: *Executes NRQL query and shows results*

### Conversation 3: Performance Analysis

**You**: "Show me the average response time by application"

**Claude**: *Queries Transaction data with FACET by appName*

## What to Expect

### ✅ What Works

- **Basic NRQL queries** - Full functionality via `query_nrdb`
- **Data discovery** - Event types and attributes
- **Simple alerts** - Create and list alert conditions
- **Session management** - Stateful conversations

### 🟨 What's Mock-Only

- **Advanced analytics** - Returns realistic fake data
- **Dashboard creation** - Returns JSON but doesn't create real dashboards
- **Workflow orchestration** - Not implemented

### ❌ What Doesn't Work

- **Governance tools** - Cost analysis, compliance checking
- **Complex discovery** - Relationship mapping, quality assessment
- **Multi-account** - Only single account supported

## Troubleshooting Quick Fixes

### "Command not found"
```bash
# Ensure binary was built
ls -la bin/mcp-server
chmod +x bin/mcp-server
```

### "Invalid API key"
```bash
# Check key format
echo $NEW_RELIC_API_KEY | grep "^NRAK-"

# Test key with New Relic API directly
curl -H "API-Key: $NEW_RELIC_API_KEY" \
  https://api.newrelic.com/graphql \
  -d '{"query": "{ actor { user { email } } }"}'
```

### "No data returned"
```bash
# Try mock mode first
./bin/mcp-server -mock

# Check account ID is correct
# Verify NEW_RELIC_REGION (US vs EU)
```

### "Claude can't connect"
```bash
# Use absolute path in Claude config
which ./bin/mcp-server
# Update config with full path: /home/user/mcp-server-newrelic/bin/mcp-server
```

## Next Steps

1. **Explore More**: Try different NRQL queries through Claude
2. **Set Up Alerts**: Create monitoring based on your discoveries  
3. **Read Documentation**: Check out the tools reference for details
4. **Test Mock Mode**: Understand what features are real vs simulated

## Advanced Quick Start

### HTTP Transport

```bash
# Start HTTP server
MCP_TRANSPORT=http ./bin/mcp-server

# Test via curl
curl -X POST http://localhost:8081/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "query_nrdb", 
      "arguments": {
        "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
      }
    },
    "id": 1
  }'
```

### Development Mode

```bash
# Enable debug logging
LOG_LEVEL=debug ./bin/mcp-server

# Use mock mode for development
MOCK_MODE=true ./bin/mcp-server
```

You're now ready to explore your New Relic data with AI assistance! The server provides a solid foundation for basic querying and discovery, with clear indicators of what features are fully implemented vs aspirational.
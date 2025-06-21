# Getting Started with New Relic MCP Server

This guide will help you get the MCP Server running and executing your first New Relic queries in less than 5 minutes.

## 🎯 Prerequisites

### Required
- **Go 1.21+** or **Docker**
- **New Relic Account** with:
  - User API Key (starts with "NRAK")
  - Account ID
  - Query permissions

### Optional
- Redis (for distributed state management)
- Claude Desktop (for AI assistant integration)

## ⚡ Quick Start (< 5 minutes)

### Option 1: Docker (Fastest)

```bash
# 1. Clone and configure
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic
cp .env.example .env

# 2. Edit .env with your credentials
# NEW_RELIC_API_KEY=NRAK-your-key-here
# NEW_RELIC_ACCOUNT_ID=your-account-id

# 3. Run with Docker
docker-compose up

# 4. Test it works
curl http://localhost:8080/health
```

### Option 2: From Source

```bash
# 1. Clone repository
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic

# 2. Copy and edit configuration
cp .env.example .env
# Edit .env with your New Relic credentials

# 3. Build
make build

# 4. Run server
./bin/mcp-server
```

## 🔑 Finding Your New Relic Credentials

### API Key
1. Log into [New Relic](https://one.newrelic.com)
2. Click your name (bottom left) → **API Keys**
3. Click **Create key**
4. Select **User** key type
5. Name it (e.g., "MCP Server")
6. Copy the key (starts with `NRAK-`)

### Account ID
- Found in the browser URL: `https://one.newrelic.com/nr1-core?account=YOUR_ACCOUNT_ID`
- Or go to: **Administration** → **Organization and access** → **Accounts**

## 🚦 First Steps

### 1. Verify Installation

```bash
# Check server health
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "components": {
    "discovery": {"status": "healthy"},
    "state": {"status": "healthy"}
  }
}
```

### 2. Discover Your Data

Let's explore what data you have in New Relic:

```bash
# List all event types in your account
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"discovery.explore_event_types"},"id":1}' | ./bin/mcp-server

# Example response shows your available event types:
# - Transaction
# - SystemSample
# - ProcessSample
# - NetworkSample
# - etc.
```

### 3. Run Your First Query

```bash
# Simple count query
cat << EOF | ./bin/mcp-server
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_nrdb",
    "arguments": {
      "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
    }
  },
  "id": 2
}
EOF
```

### 4. Explore Event Attributes

```bash
# See what fields are available in Transaction events
cat << EOF | ./bin/mcp-server
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "discovery.explore_attributes",
    "arguments": {
      "event_type": "Transaction"
    }
  },
  "id": 3
}
EOF
```

## 🤖 Claude Desktop Integration

### Step 1: Configure Claude Desktop

Find your Claude configuration file:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

### Step 2: Add MCP Server Configuration

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "--env-file", "/path/to/your/.env",
        "mcp-server-newrelic:latest"
      ]
    }
  }
}
```

Or if running from binary:

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/path/to/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-your-key",
        "NEW_RELIC_ACCOUNT_ID": "your-account-id"
      }
    }
  }
}
```

### Step 3: Restart Claude and Test

1. Completely quit Claude Desktop
2. Start Claude Desktop again
3. Test with: "What event types are available in my New Relic account?"

## 🧪 Using Mock Mode

Perfect for development and testing without a New Relic account:

```bash
# Run in mock mode
./bin/mcp-server -mock

# Or with Docker
docker run -i --rm mcp-server-newrelic:latest -mock
```

Mock mode provides realistic sample data for:
- All discovery operations
- NRQL query results
- Alert and dashboard operations
- Analysis tools

## 📊 Example Workflows

### Performance Investigation
```bash
# 1. Find transaction event types
# 2. Explore transaction attributes
# 3. Query slow transactions
# 4. Analyze patterns
```

### Infrastructure Monitoring
```bash
# 1. Discover SystemSample events
# 2. Check available metrics
# 3. Query high CPU hosts
# 4. Create alerts
```

## 🔧 Troubleshooting

### Common Issues

**"Authentication failed"**
```bash
# Verify your API key
echo $NEW_RELIC_API_KEY
# Should start with NRAK-

# Check it's a User key, not License key
# User keys have query permissions
```

**"No data returned"**
```bash
# Try a wider time range
"SELECT count(*) FROM Transaction SINCE 7 days ago"

# Verify account ID
echo $NEW_RELIC_ACCOUNT_ID
```

**"Connection error"**
```bash
# Check region setting
NEW_RELIC_REGION=EU  # if using EU datacenter

# Enable debug logging
LOG_LEVEL=debug ./bin/mcp-server
```

### Diagnostics Tool

```bash
# Run built-in diagnostics
make diagnose

# Checks:
# ✓ Environment configuration
# ✓ API key format
# ✓ Network connectivity
# ✓ New Relic API access
```

## 📚 What's Next?

### 1. Explore Available Tools
See [Tools Overview](30_TOOLS_OVERVIEW.md) for the complete catalog of 40+ tools:
- Discovery tools for exploring your data
- Query tools for NRQL execution
- Alert management tools
- Dashboard creation tools
- Analysis and insights tools

### 2. Learn Core Concepts
Read [Concepts Guide](04_CONCEPTS.md) to understand:
- Discovery-first philosophy
- Tool composition patterns
- State management
- Error handling

### 3. Try Advanced Features
- [Discovery Workflows](43_GUIDE_DISCOVERY_WORKFLOWS.md) - Advanced exploration patterns
- [Mock Mode Guide](48_GUIDE_MOCK_MODE.md) - Development without credentials
- [Examples](50_EXAMPLES_OVERVIEW.md) - Real-world scenarios

### 4. Configuration Options
See [Configuration Guide](03_CONFIGURATION.md) for:
- Advanced authentication options
- Performance tuning
- Caching configuration
- Transport selection

## 🎉 Success Checklist

- [ ] Server running and healthy
- [ ] First query executed successfully
- [ ] Event types discovered
- [ ] Attributes explored
- [ ] (Optional) Claude Desktop integrated
- [ ] (Optional) Mock mode tested

## 💡 Tips

1. **Start with discovery** - Don't assume data structures
2. **Use mock mode** for development and testing
3. **Check the logs** with `LOG_LEVEL=debug` for issues
4. **Join the community** for help and updates

---

**Need help?** Check the [FAQ](09_FAQ.md) or [open an issue](https://github.com/deepaucksharma/mcp-server-newrelic/issues).
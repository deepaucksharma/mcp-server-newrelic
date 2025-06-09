# MCP Server New Relic - Setup Guide

## Installation Completed ✅

The mcp-server-newrelic repository has been successfully cloned and set up. Here's what was done:

### 1. Repository Location
- **Path**: `/Users/deepaksharma/syc/mcp-server-newrelic`
- **Python Version**: Python 3.11 (located at `/opt/homebrew/bin/python3.11`)
- **Dependencies**: All Python dependencies installed successfully

### 2. Configuration Files Created

#### For Claude Code
- **Config File**: `claude-code-config.json`
- **Transport**: STDIO (standard input/output)
- **Features**: Enhanced plugins and audit logging enabled

#### For GitHub Copilot
- **Config File**: `.vscode/settings.json`
- **Transport**: HTTP (port 3000)
- **Features**: MCP experimental features enabled

### 3. Environment Configuration
- **File**: `.env` (created from template)
- **Status**: Ready for your credentials

## Next Steps

### Step 1: Configure Your New Relic Credentials

Edit the `.env` file and replace the placeholders with your actual New Relic credentials:

```bash
cd /Users/deepaksharma/syc/mcp-server-newrelic
nano .env  # or use your preferred editor
```

Replace:
- `NRAK-YOUR-API-KEY-HERE` with your actual New Relic API key
- `YOUR-ACCOUNT-ID` with your New Relic account ID

### Step 2: Test the Server

Test if the server can connect to New Relic:

```bash
cd /Users/deepaksharma/syc/mcp-server-newrelic
/opt/homebrew/bin/python3.11 main.py --test
```

### Step 3: Configure for Claude Code

#### Option A: Using Claude Desktop
1. Copy the configuration to Claude Desktop's config directory:
   ```bash
   cp claude-code-config.json ~/Library/Application\ Support/Claude/claude_desktop_config.json
   ```

2. Restart Claude Desktop

3. Test by asking Claude: "Can you check my New Relic applications?"

#### Option B: Using Claude Code CLI
1. Install Claude Code CLI (if not already installed):
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```

2. Initialize with the MCP configuration:
   ```bash
   claude-code init --mcp-config=/Users/deepaksharma/syc/mcp-server-newrelic/claude-code-config.json
   ```

3. Start a chat session:
   ```bash
   claude-code chat
   ```

### Step 4: Configure for GitHub Copilot

1. Open the project in VS Code:
   ```bash
   cd /Users/deepaksharma/syc/mcp-server-newrelic
   code .
   ```

2. Make sure you have the GitHub Copilot Chat extension installed

3. Start the MCP server in HTTP mode:
   ```bash
   /opt/homebrew/bin/python3.11 main.py
   ```

4. Open GitHub Copilot Chat (`Cmd+Shift+I`) and test:
   ```
   @copilot /newrelic What applications are monitored?
   ```

## Available MCP Tools

Once configured, you'll have access to these tools through your AI assistant:

### Core Tools
- `search_entities` - Search for New Relic entities (applications, hosts, etc.)
- `get_entity_details` - Get detailed information about a specific entity
- `get_entity_golden_signals` - Get key metrics for an entity
- `get_entity_relationships` - Discover entity dependencies

### Query Tools
- `run_nrql_query` - Execute NRQL queries
- `run_graphql_query` - Execute raw GraphQL queries

### APM Tools
- `list_apm_applications` - List all APM applications
- `get_apm_metrics` - Get application performance metrics
- `get_apm_transactions` - Get transaction details
- `list_apm_deployments` - List recent deployments

### Infrastructure Tools
- `list_infrastructure_hosts` - List infrastructure hosts
- `get_host_metrics` - Get host performance metrics
- `list_kubernetes_clusters` - List K8s clusters
- `get_container_metrics` - Get container metrics

### Alerts & Incidents
- `list_recent_incidents` - List recent incidents
- `get_incident_details` - Get incident information
- `list_alert_policies` - List alert policies

### Logs
- `search_logs` - Search log data
- `get_log_patterns` - Analyze log patterns

## Troubleshooting

### Common Issues

1. **Python Version Error**
   - Make sure to use Python 3.11: `/opt/homebrew/bin/python3.11`
   - The system Python (3.9.6) is too old for this project

2. **API Key Issues**
   - Ensure your API key starts with `NRAK-`
   - Check that it has the necessary permissions in New Relic

3. **Connection Issues**
   - Verify your internet connection
   - Check if you need to configure proxy settings
   - Ensure your New Relic region (US/EU) is correct in `.env`

### Testing Commands

Test individual components:

```bash
# Test API connection
/opt/homebrew/bin/python3.11 -c "from core.nerdgraph_client import NerdGraphClient; import asyncio; asyncio.run(NerdGraphClient(api_key='YOUR_KEY').query('{ actor { user { email } } }'))"

# Run health check
/opt/homebrew/bin/python3.11 main.py --health-check

# Run in debug mode
LOG_LEVEL=DEBUG /opt/homebrew/bin/python3.11 main.py
```

## Additional Resources

- **Main README**: `/Users/deepaksharma/syc/mcp-server-newrelic/README.md`
- **Integration Guide**: `/Users/deepaksharma/syc/mcp-server-newrelic/INTEGRATION_GUIDE.md`
- **API Documentation**: Check the `features/` directory for available tools
- **Example Scripts**: See `scripts/` directory for automation examples

## Support

If you encounter issues:
1. Check the logs in the terminal where you run the server
2. Review the troubleshooting section above
3. Refer to the comprehensive INTEGRATION_GUIDE.md
4. Check the project's GitHub issues page
# 5-Minute Quickstart Guide

Get up and running with the Enhanced MCP Server New Relic in 5 minutes.

## Prerequisites

- Node.js 18+ installed
- New Relic account with API access
- Claude Desktop or MCP-compatible client

## Step 1: Installation (1 minute)

### Clone and Install
```bash
git clone https://github.com/your-org/mcp-server-newrelic.git
cd mcp-server-newrelic
npm install
```

### Build TypeScript
```bash
npm run build
```

## Step 2: Configuration (2 minutes)

### Get New Relic Credentials

1. **API Key**: Go to [New Relic API Keys](https://one.newrelic.com/admin-portal/api-keys-ui/api-keys)
   - Create a "User" type API key
   - Copy the key

2. **Account ID**: Found in the URL when logged into New Relic
   - Example: `https://one.newrelic.com/launcher/account/YOUR_ACCOUNT_ID`

3. **Region**: `US` (default) or `EU`

### Set Environment Variables

**Option A: Environment File**
```bash
cp .env.example .env
```

Edit `.env`:
```bash
NEW_RELIC_API_KEY=your_api_key_here
NEW_RELIC_ACCOUNT_ID=your_account_id_here
NEW_RELIC_REGION=US
```

**Option B: Export Variables**
```bash
export NEW_RELIC_API_KEY="your_api_key_here"
export NEW_RELIC_ACCOUNT_ID="your_account_id_here"
export NEW_RELIC_REGION="US"
```

## Step 3: Configure Claude Desktop (1 minute)

Add to your Claude Desktop configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "newrelic-enhanced": {
      "command": "node",
      "args": ["/path/to/mcp-server-newrelic/dist/index.js"],
      "env": {
        "NEW_RELIC_API_KEY": "your_api_key_here",
        "NEW_RELIC_ACCOUNT_ID": "your_account_id_here",
        "NEW_RELIC_REGION": "US"
      }
    }
  }
}
```

**Important**: Replace `/path/to/mcp-server-newrelic` with the actual path to your installation.

## Step 4: Test Connection (1 minute)

### Start Claude Desktop
Restart Claude Desktop to load the new server configuration.

### Test Basic Connectivity

In Claude, try:
```
Use discover.environment to show me my New Relic setup
```

You should see a comprehensive environment discovery response showing your monitored entities, telemetry types, and schema guidance.

## Step 5: Quick Exploration (5 minutes)

### 1. Environment Discovery
```
Discover my complete New Relic environment including health status
```

This runs `discover.environment` with health checks enabled.

### 2. Generate a Dashboard
```
Create a golden signals dashboard for my main application service
```

This will:
1. Find your application entities
2. Generate a comprehensive dashboard preview
3. Show you the JSON definition

### 3. Compare Service Performance
```
Compare the performance of all my application services
```

This runs `compare.similar_entities` to identify performance patterns and outliers.

## Common First-Time Issues

### Issue: "No entities found"
**Solution**: Ensure your New Relic account has monitored applications or services.

### Issue: "Authentication failed"
**Solution**: Verify your API key and account ID are correct.

### Issue: "Tool not found"
**Solution**: Restart Claude Desktop after configuration changes.

### Issue: "Empty responses"
**Solution**: Check that your account has recent telemetry data.

## What You Get Out of the Box

### Enhanced Existing Tools
- `run_nrql_query` - Execute NRQL with discovery validation
- `search_entities` - Find monitored entities with rich metadata
- `get_entity_details` - Get comprehensive entity information

### Composite Intelligence Tools  
- `discover.environment` - Complete environment analysis
- `generate.golden_dashboard` - Intelligent dashboard creation
- `compare.similar_entities` - Performance benchmarking

### Platform Tools
- `platform_analyze_adoption` - OpenTelemetry adoption analysis
- `cache.stats` - Cache performance monitoring
- `cache.clear` - Cache management

## Next Steps

### Explore Advanced Features

1. **Custom Dashboards**
   ```
   Create a dashboard for entity ABC123 with custom alert thresholds
   ```

2. **Performance Analysis**
   ```
   Compare all my microservices and identify the slowest ones
   ```

3. **Observability Assessment**
   ```
   Analyze my observability setup and suggest improvements
   ```

### Read the Documentation

- [Architecture Overview](10_ARCHITECTURE_OVERVIEW.md) - Understand the system design
- [Discovery-First Philosophy](11_ARCHITECTURE_DISCOVERY_FIRST.md) - Learn the core principles
- [Composite Tools](31_TOOLS_COMPOSITE.md) - Deep dive into advanced tools
- [Configuration Guide](03_CONFIGURATION.md) - Advanced configuration options

## Example Workflows

### Workflow 1: New Environment Assessment
```
1. "Discover my New Relic environment"
2. "What observability gaps do I have?"
3. "Show me the performance of my top 5 services"
4. "Create dashboards for my critical services"
```

### Workflow 2: Performance Investigation
```
1. "Compare all my application services"
2. "Which services are underperforming?"
3. "Create a detailed dashboard for the slowest service"
4. "What optimizations do you recommend?"
```

### Workflow 3: Dashboard Creation
```
1. "Find all my monitored applications"
2. "Generate golden signals dashboards for each"
3. "Include custom alert thresholds"
4. "Actually create the dashboards in New Relic"
```

## Key Features Highlight

### 🔍 Discovery-First
- Zero hardcoded schemas
- Automatic OpenTelemetry vs APM detection
- Intelligent query adaptation

### 📊 Golden Signals Intelligence
- Latency, Traffic, Errors, Saturation analysis
- Anomaly detection with statistical analysis
- Baseline establishment with confidence scoring

### 🎯 Composite Tools
- Multi-operation workflows in single calls
- Rich context and actionable recommendations
- LLM-optimized output formatting

### ⚡ Intelligent Caching
- Context-aware freshness strategies
- Adaptive TTL based on data characteristics
- Background refresh for critical metrics

## Troubleshooting

### Debug Mode
Enable debug logging:
```bash
export DEBUG=true
```

### Check Logs
Logs are written to stderr and visible in Claude Desktop's developer console.

### Verify Configuration
```
Show me the cache statistics to verify the server is working
```

### Test Individual Tools
```
Run a simple NRQL query to test basic connectivity
```

## Support

- **Documentation**: See the complete docs in the `/docs` folder
- **Issues**: Report issues on GitHub
- **Examples**: Check `/examples` for more usage patterns

---

**You're now ready to explore intelligent observability with Enhanced MCP Server New Relic!**

**Next**: [50_EXAMPLES_OVERVIEW.md](50_EXAMPLES_OVERVIEW.md) for real-world usage examples
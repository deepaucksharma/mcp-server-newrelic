# Installation Guide

Complete installation instructions for the Enhanced MCP Server New Relic across different platforms and deployment scenarios.

## Prerequisites

### System Requirements
- **Node.js**: 20.0.0 or higher
- **npm**: 9.0.0 or higher (or yarn/pnpm equivalent)
- **Memory**: 512MB minimum, 2GB recommended
- **Platform**: Linux, macOS, Windows (with WSL2)

### New Relic Requirements
- **New Relic Account** with active APM, Infrastructure, or Logs data
- **API Key** with the following permissions:
  - NRQL Query access
  - Entity search access
  - Dashboard read access (optional, for validation)

## Installation Methods

### Method 1: From Source (Recommended for Development)

```bash
# Clone the repository
git clone <repository-url>
cd mcp-server-newrelic

# Install dependencies
npm install

# Build the project
npm run build

# Verify installation
npm run type-check
```

### Method 2: Binary Installation (Coming Soon)

```bash
# Install globally via npm (when published)
npm install -g mcp-server-newrelic

# Or download binary release
curl -L https://github.com/.../releases/latest/download/mcp-newrelic-linux -o mcp-newrelic
chmod +x mcp-newrelic
```

### Method 3: Docker (Coming Soon)

```bash
# Pull the official image
docker pull mcp-server-newrelic:latest

# Run with environment variables
docker run -e NEW_RELIC_API_KEY=your_key mcp-server-newrelic:latest
```

## Configuration

### Environment Variables

Create a `.env` file or set environment variables:

```bash
# Required Configuration
NEW_RELIC_API_KEY="NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
NEW_RELIC_ACCOUNT_ID="1234567"

# Optional Configuration
NEW_RELIC_REGION="US"                    # or "EU"
DEBUG="false"                            # Enable debug logging
CACHE_TTL_MULTIPLIER="1.0"              # Adjust cache TTL (0.5-2.0)
```

### API Key Setup

1. **Log into New Relic** at [one.newrelic.com](https://one.newrelic.com)
2. **Navigate to API Keys**: User menu → API keys
3. **Create API Key**:
   - Type: "User API Key"
   - Name: "MCP Server"
   - Copy the key (starts with "NRAK-")

### Multiple Account Configuration

For multi-account setups, you can configure additional accounts:

```bash
# Additional accounts for platform analysis
E2E_ACCOUNT_LEGACY_APM="1234567"
E2E_API_KEY_LEGACY="NRAK-..."

E2E_ACCOUNT_MODERN_OTEL="2345678"
E2E_API_KEY_OTEL="NRAK-..."

E2E_ACCOUNT_MIXED_DATA="3456789"
E2E_API_KEY_MIXED="NRAK-..."
```

## Verification

### 1. Test Installation

```bash
# Check TypeScript compilation
npm run type-check

# Run unit tests
npm test

# Test discovery functionality
npm run discover
```

### 2. Test New Relic Connection

```bash
# Test with your credentials
NEW_RELIC_API_KEY="your_key" NEW_RELIC_ACCOUNT_ID="your_account" npm run discover
```

Expected output:
```
🔍 Starting discovery process...
📊 Discovery Results:
===================
Account ID: 1234567
Confidence: 95.2%
Schemas discovered: 8
Attributes profiled: 156
Error indicators: 3
Metrics available: 42
```

### 3. Run E2E Tests

```bash
# Quick E2E test (requires valid credentials)
npm run test:e2e:quick

# Full E2E test suite
npm run test:e2e
```

## Development Setup

### IDE Configuration

For **Visual Studio Code**:

```json
// .vscode/settings.json
{
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "editor.formatOnSave": true
}
```

### Git Hooks

```bash
# Install pre-commit hooks
npm install husky --save-dev
npx husky install
npx husky add .husky/pre-commit "npm run type-check && npm run lint"
```

## Deployment Options

### MCP Client Integration (Claude Desktop)

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "node",
      "args": ["/path/to/mcp-server-newrelic/dist/index.js"],
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-...",
        "NEW_RELIC_ACCOUNT_ID": "1234567",
        "NEW_RELIC_REGION": "US"
      }
    }
  }
}
```

### Systemd Service (Linux)

```ini
# /etc/systemd/system/mcp-newrelic.service
[Unit]
Description=MCP Server New Relic
After=network.target

[Service]
Type=simple
User=mcp
WorkingDirectory=/opt/mcp-server-newrelic
ExecStart=/usr/bin/node dist/index.js
Environment=NODE_ENV=production
Environment=NEW_RELIC_API_KEY=NRAK-...
Environment=NEW_RELIC_ACCOUNT_ID=1234567
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### PM2 (Process Manager)

```json
// ecosystem.config.js
module.exports = {
  apps: [{
    name: 'mcp-newrelic',
    script: 'dist/index.js',
    env: {
      NODE_ENV: 'production',
      NEW_RELIC_API_KEY: 'NRAK-...',
      NEW_RELIC_ACCOUNT_ID: '1234567'
    },
    instances: 1,
    autorestart: true,
    max_memory_restart: '1G'
  }]
};
```

## Troubleshooting

### Common Installation Issues

**❌ Node.js version error**
```bash
# Check version
node --version  # Should be 20.0.0+

# Install/update Node.js via nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 20
nvm use 20
```

**❌ TypeScript compilation errors**
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Rebuild
npm run build
```

**❌ Missing dependencies**
```bash
# Install peer dependencies
npm install graphql-request@^7.0.0 zod@^3.23.0
```

### Connection Issues

**❌ "Invalid API credentials"**
- Verify API key format (starts with `NRAK-`)
- Check API key permissions in New Relic UI
- Ensure account ID is correct (numeric)

**❌ "No data found"**
- Verify account has active data ingestion
- Check time range (data may be older)
- Try with a different account ID

**❌ "Rate limit exceeded"**
- Wait 1-2 minutes and retry
- Reduce concurrent requests
- Check cache configuration

### Performance Issues

**❌ High memory usage**
```bash
# Monitor cache statistics
npm run discover  # Check "Memory Usage" in output

# Clear cache if needed
# (programmatically via cache.clear tool)
```

**❌ Slow query responses**
```bash
# Enable debug logging
DEBUG=true npm run dev

# Check for schema validation overhead
# Consider disabling for known queries
```

## Security Considerations

### API Key Management
- **Never commit API keys** to version control
- Use environment variables or secure secret management
- Rotate API keys regularly
- Use least-privilege permissions

### Network Security
- Server runs on localhost by default
- No external network listeners
- All data in-memory only (no persistence)

### Data Privacy
- No customer data is logged or stored permanently
- Cache is memory-only with automatic cleanup
- Supports EU region for GDPR compliance

## Update Procedures

### Minor Updates
```bash
git pull origin main
npm install
npm run build
```

### Major Version Updates
1. Review changelog for breaking changes
2. Update environment variables if needed
3. Run full test suite
4. Update client configurations

## Support

### Self-Help Resources
- **Documentation**: Browse the `/docs` directory
- **Examples**: Check `/examples` for usage patterns
- **Tests**: Review test files for implementation details

### Getting Help
- **Issues**: Report bugs via the issue tracker
- **Discussions**: Community discussions in the forum
- **Documentation**: Submit documentation improvements

---

**Next**: [03_CONFIGURATION.md](03_CONFIGURATION.md) for advanced configuration options
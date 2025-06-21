# Configuration Reference

Complete configuration reference for the New Relic MCP Server. Configuration can be provided through environment variables, configuration files, or command-line flags.

## 📋 Table of Contents

1. [Configuration Methods](#configuration-methods)
2. [Required Settings](#required-settings)
3. [Server Configuration](#server-configuration)
4. [New Relic Settings](#new-relic-settings)
5. [Security Configuration](#security-configuration)
6. [Discovery Engine](#discovery-engine)
7. [State Management](#state-management)
8. [Performance Tuning](#performance-tuning)
9. [Monitoring & Logging](#monitoring--logging)
10. [Development Options](#development-options)
11. [Environment Examples](#environment-examples)
12. [Troubleshooting](#troubleshooting)

## 🔧 Configuration Methods

Configuration is loaded in order of precedence:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file** (.env)
4. **Default values** (lowest priority)

### Environment Variables (.env)

The primary configuration method:

```bash
# Create from example
cp .env.example .env

# Edit with your settings
nano .env
```

### Configuration File Locations

The server checks these locations in order:
1. `./.env` (current directory)
2. `~/.config/mcp-newrelic/.env`
3. `/etc/mcp-newrelic/.env`

### Command-line Flags

Override any setting at runtime:

```bash
# Examples
./mcp-server --transport=http --port=9090
./mcp-server --mock --log-level=debug
./mcp-server --env-file=/custom/path/.env
```

## ⚠️ Required Settings

These MUST be configured for production use:

### NEW_RELIC_API_KEY
- **Type**: String
- **Required**: Yes (unless MOCK_MODE=true)
- **Description**: New Relic User API key
- **Format**: Must start with "NRAK-"
- **Example**: `NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXX`

**How to obtain**:
1. Log into [New Relic](https://one.newrelic.com)
2. Click your name → **API Keys**
3. Create **User** key with permissions:
   - NRQL query
   - APM/Infrastructure read
   - Dashboards/Alerts (if using those tools)

### NEW_RELIC_ACCOUNT_ID
- **Type**: String (numeric)
- **Required**: Yes (unless MOCK_MODE=true)
- **Description**: Your New Relic account ID
- **Example**: `1234567`

**How to find**:
- In New Relic URL: `https://one.newrelic.com/nr1-core?account=YOUR_ACCOUNT_ID`
- Or: Administration → Organization → Accounts

## 🖥️ Server Configuration

### MCP_TRANSPORT
- **Type**: String
- **Default**: `stdio`
- **Options**: `stdio`, `http`, `sse`
- **Description**: Transport protocol
  - `stdio`: For Claude Desktop integration
  - `http`: REST API for programmatic access
  - `sse`: Server-Sent Events for streaming

### SERVER_HOST
- **Type**: String
- **Default**: `0.0.0.0`
- **Description**: Server bind address
- **Security**: Use `127.0.0.1` for localhost only

### SERVER_PORT
- **Type**: Integer
- **Default**: `8080`
- **Description**: HTTP server port

### HTTP_PORT
- **Type**: Integer
- **Default**: `8081`
- **Description**: MCP HTTP transport port

### REQUEST_TIMEOUT
- **Type**: Duration
- **Default**: `30s`
- **Description**: Default request timeout
- **Format**: Go duration (e.g., `30s`, `5m`, `1h`)

### MAX_CONCURRENT_REQUESTS
- **Type**: Integer
- **Default**: `100`
- **Description**: Max concurrent requests
- **Performance**: Adjust based on resources

## 🌍 New Relic Settings

### NEW_RELIC_REGION
- **Type**: String
- **Default**: `US`
- **Options**: `US`, `EU`
- **Description**: Data center region

### NEW_RELIC_LICENSE_KEY
- **Type**: String
- **Optional**: Yes
- **Description**: For monitoring the MCP server itself
- **Format**: Starts with "NRLK-"

### NEW_RELIC_APP_NAME
- **Type**: String
- **Default**: `mcp-server-newrelic`
- **Description**: APM application name

## 🔐 Security Configuration

### AUTH_ENABLED
- **Type**: Boolean
- **Default**: `false`
- **Description**: Enable API authentication
- **Recommendation**: Enable for production

### JWT_SECRET
- **Type**: String
- **Required**: If AUTH_ENABLED=true
- **Description**: Secret for JWT tokens
- **Generation**: `openssl rand -base64 32`

### API_KEY_SALT
- **Type**: String
- **Required**: If AUTH_ENABLED=true
- **Description**: Salt for API key hashing
- **Generation**: `openssl rand -base64 16`

### RATE_LIMIT_ENABLED
- **Type**: Boolean
- **Default**: `true`
- **Description**: Enable rate limiting

### RATE_LIMIT_PER_MIN
- **Type**: Integer
- **Default**: `60`
- **Description**: Requests per minute per client

### TLS_ENABLED
- **Type**: Boolean
- **Default**: `false`
- **Description**: Enable HTTPS
- **Required**: For production

### TLS_CERT_FILE / TLS_KEY_FILE
- **Type**: String
- **Required**: If TLS_ENABLED=true
- **Description**: Paths to TLS certificate and key

## 🔍 Discovery Engine

### DISCOVERY_CACHE_TTL
- **Type**: Duration
- **Default**: `3600s` (1 hour)
- **Description**: Cache duration for discovery results

### DISCOVERY_MAX_WORKERS
- **Type**: Integer
- **Default**: `10`
- **Description**: Parallel discovery workers
- **Performance**: Higher = faster but more load

### DISCOVERY_SAMPLE_SIZE
- **Type**: Integer
- **Default**: `1000`
- **Description**: Default attribute analysis sample size

### DISCOVERY_PATTERN_MIN_CONFIDENCE
- **Type**: Float
- **Default**: `0.7`
- **Range**: 0.0 - 1.0
- **Description**: Pattern detection threshold

## 💾 State Management

### STATE_STORE
- **Type**: String
- **Default**: `memory`
- **Options**: `memory`, `redis`
- **Description**: State storage backend

### REDIS_URL
- **Type**: String
- **Required**: If STATE_STORE=redis
- **Format**: `redis://[user:password@]host:port/db`
- **Example**: `redis://localhost:6379/0`

### SESSION_TTL
- **Type**: Duration
- **Default**: `1800s` (30 minutes)
- **Description**: Session expiration time

### CACHE_TTL
- **Type**: Duration
- **Default**: `300s` (5 minutes)
- **Description**: Default cache TTL

## ⚡ Performance Tuning

### QUERY_CACHE_ENABLED
- **Type**: Boolean
- **Default**: `true`
- **Description**: Cache NRQL query results

### QUERY_CACHE_TTL
- **Type**: Duration
- **Default**: `300s`
- **Description**: Query result cache duration

### BATCH_SIZE
- **Type**: Integer
- **Default**: `100`
- **Description**: Default batch operation size

### WORKER_POOL_SIZE
- **Type**: Integer
- **Default**: `20`
- **Description**: Size of worker pool for concurrent operations

## 📊 Monitoring & Logging

### LOG_LEVEL
- **Type**: String
- **Default**: `info`
- **Options**: `debug`, `info`, `warn`, `error`
- **Description**: Logging verbosity

### LOG_FORMAT
- **Type**: String
- **Default**: `text`
- **Options**: `text`, `json`
- **Description**: Log output format

### METRICS_ENABLED
- **Type**: Boolean
- **Default**: `true`
- **Description**: Enable Prometheus metrics

### METRICS_PORT
- **Type**: Integer
- **Default**: `9090`
- **Description**: Prometheus metrics port

### TRACE_ENABLED
- **Type**: Boolean
- **Default**: `false`
- **Description**: Enable distributed tracing

## 🧪 Development Options

### MOCK_MODE
- **Type**: Boolean
- **Default**: `false`
- **Description**: Use mock data (no New Relic connection)
- **Use Case**: Development and testing

### DEV_MODE
- **Type**: Boolean
- **Default**: `false`
- **Description**: Development mode (verbose errors)
- **Warning**: Never use in production

### DEBUG
- **Type**: Boolean
- **Default**: `false`
- **Description**: Enable debug output

### ENABLE_PROFILING
- **Type**: Boolean
- **Default**: `false`
- **Description**: Enable pprof endpoints

## 📝 Environment Examples

### Minimal Development
```env
# .env.dev
NEW_RELIC_API_KEY=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXX
NEW_RELIC_ACCOUNT_ID=1234567
LOG_LEVEL=debug
```
### Claude Desktop Integration
```env
# .env.claude
NEW_RELIC_API_KEY=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXX
NEW_RELIC_ACCOUNT_ID=1234567
MCP_TRANSPORT=stdio
LOG_LEVEL=warn
DISCOVERY_CACHE_TTL=3600s
```

### Production Configuration
```env
# .env.prod
# New Relic
NEW_RELIC_API_KEY=NRAK-XXXXXXXXXXXXXXXXXXXXXXXXXXXXX
NEW_RELIC_ACCOUNT_ID=1234567
NEW_RELIC_REGION=US

# Security
AUTH_ENABLED=true
JWT_SECRET=<generate-unique-value>
API_KEY_SALT=<generate-unique-value>
TLS_ENABLED=true
TLS_CERT_FILE=/etc/ssl/certs/server.crt
TLS_KEY_FILE=/etc/ssl/private/server.key

# Performance
MCP_TRANSPORT=http
STATE_STORE=redis
REDIS_URL=redis://redis:6379/0
DISCOVERY_CACHE_TTL=7200s
QUERY_CACHE_TTL=600s
MAX_CONCURRENT_REQUESTS=500

# Monitoring
LOG_LEVEL=info
LOG_FORMAT=json
METRICS_ENABLED=true
TRACE_ENABLED=true
```
### Mock Mode Development
```env
# .env.mock
MOCK_MODE=true
LOG_LEVEL=debug
MCP_TRANSPORT=stdio
```

## 🔧 Troubleshooting

### Invalid API Key
```
Error: API key validation failed
```
**Solution**: 
- Verify key starts with "NRAK-"
- Check key has correct permissions
- Ensure using User key, not License key

### Missing Account ID
```
Error: NEW_RELIC_ACCOUNT_ID is required
```
**Solution**:
- Find in New Relic UI under Administration
- Check URL when logged in

### Connection Issues
```
Error: Failed to connect to New Relic API
```
**Solution**:
- Check NEW_RELIC_REGION (US vs EU)
- Verify network connectivity
- Check firewall/proxy settings

### Performance Issues
```
Warning: High memory usage detected
```
**Solution**:
- Reduce MAX_CONCURRENT_REQUESTS
- Lower DISCOVERY_MAX_WORKERS
- Enable caching with longer TTLs

## 🛡️ Security Best Practices

1. **Generate unique secrets for production**
   ```bash
   # Generate JWT_SECRET
   openssl rand -base64 32
   
   # Generate API_KEY_SALT
   openssl rand -base64 16
   ```

2. **Use environment-specific files**
   - `.env.dev` for development
   - `.env.prod` for production
   - Never commit .env files to git

3. **Enable security features in production**
   - Always set AUTH_ENABLED=true
   - Always set TLS_ENABLED=true
   - Configure proper rate limiting

4. **Restrict API key permissions**
   - Only grant necessary permissions
   - Use different keys per environment
   - Rotate keys regularly

5. **Monitor configuration**
   - Log configuration on startup (excluding secrets)
   - Alert on configuration changes
   - Audit access regularly

## 📚 Related Documentation

- [Installation Guide](02_INSTALLATION.md) - Setup instructions
- [Security Architecture](14_ARCHITECTURE_SECURITY.md) - Security details
- [Operations Guide](75_OPERATIONS_MONITORING.md) - Monitoring setup
- [Troubleshooting](79_OPERATIONS_TROUBLESHOOTING.md) - Common issues

---

**Configuration Help**: Run `./mcp-server --help` for all command-line options.
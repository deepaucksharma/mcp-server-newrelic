# Troubleshooting Guide

This guide helps you diagnose and fix common issues with the New Relic MCP Server. Follow the diagnostic steps in order for the most efficient troubleshooting.

## Table of Contents

1. [Quick Diagnostics](#quick-diagnostics)
2. [Common Issues and Solutions](#common-issues-and-solutions)
3. [Connection and Authentication Issues](#connection-and-authentication-issues)
4. [Tool-Specific Troubleshooting](#tool-specific-troubleshooting)
5. [Performance Troubleshooting](#performance-troubleshooting)
6. [Debug Mode and Logging](#debug-mode-and-logging)
7. [Error Code Reference](#error-code-reference)
8. [Frequently Asked Questions](#frequently-asked-questions)
9. [Getting Help](#getting-help)

## Quick Diagnostics

### Run Built-in Diagnostics
```bash
# Check environment and configuration
make diagnose

# Auto-fix common issues
make diagnose-fix

# Validate specific components
./bin/diagnose --component=auth
./bin/diagnose --component=network
./bin/diagnose --component=tools
```

### Check Server Health
```bash
# Test server startup
./bin/mcp-server --version

# Test with MCP inspector
npx @modelcontextprotocol/inspector ./bin/mcp-server

# Test basic tool execution
echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":1}' | ./bin/mcp-server
```

## Common Issues and Solutions

### Issue: Server Won't Start

**Error Message:**
```
Failed to start MCP server: failed to initialize New Relic client: API key not found
```

**Solution:**
1. Check environment variables:
   ```bash
   # Verify .env file exists
   ls -la .env
   
   # Check required variables
   grep -E "NEW_RELIC_API_KEY|NEW_RELIC_ACCOUNT_ID" .env
   ```

2. Ensure proper format:
   ```bash
   # .env file should contain:
   NEW_RELIC_API_KEY=NRAK-XXXXXXXXXXXXXXXXXXXXXX
   NEW_RELIC_ACCOUNT_ID=1234567
   ```

3. Validate API key:
   ```bash
   make diagnose --component=auth
   ```

### Issue: "Command Not Found" Errors

**Error Message:**
```
bash: ./bin/mcp-server: No such file or directory
```

**Solution:**
1. Build the server:
   ```bash
   make build
   
   # Verify binary exists
   ls -la ./bin/mcp-server
   ```

2. Check permissions:
   ```bash
   chmod +x ./bin/mcp-server
   ```

### Issue: Mock Mode Not Working

**Error Message:**
```
Mock mode enabled but tools returning errors
```

**Solution:**
1. Verify mock mode is enabled:
   ```bash
   # Run in mock mode
   make run-mock
   
   # Or set environment variable
   export MOCK_MODE=true
   ./bin/mcp-server
   ```

2. Check logs for mock data generation:
   ```bash
   export LOG_LEVEL=DEBUG
   ./bin/mcp-server 2>&1 | grep -i mock
   ```

## Connection and Authentication Issues

### Issue: API Key Invalid

**Error Message:**
```
NerdGraph error: Unauthorized - Invalid API Key
```

**Solution:**
1. Verify API key format:
   ```bash
   # Should start with NRAK-
   echo $NEW_RELIC_API_KEY | grep -E "^NRAK-[A-Z0-9]{27}$"
   ```

2. Check API key permissions:
   - Log into New Relic UI
   - Go to Administration > API Keys
   - Verify key has required permissions:
     - NRQL query permissions
     - Dashboard read/write (if using dashboard tools)
     - Alert configuration (if using alert tools)

3. Test API key directly:
   ```bash
   curl -X POST https://api.newrelic.com/graphql \
     -H "API-Key: $NEW_RELIC_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"query":"{ actor { user { email } } }"}'
   ```

### Issue: Account ID Mismatch

**Error Message:**
```
Account 1234567 not found or not accessible
```

**Solution:**
1. Verify account ID:
   ```bash
   # Check configured account
   echo $NEW_RELIC_ACCOUNT_ID
   
   # List accessible accounts
   ./bin/mcp-server --tool=list_accounts
   ```

2. Update configuration:
   ```bash
   # Edit .env file
   NEW_RELIC_ACCOUNT_ID=correct_account_id
   ```

### Issue: Network Connectivity

**Error Message:**
```
Post "https://api.newrelic.com/graphql": dial tcp: lookup api.newrelic.com: no such host
```

**Solution:**
1. Check DNS resolution:
   ```bash
   nslookup api.newrelic.com
   dig api.newrelic.com
   ```

2. Test connectivity:
   ```bash
   curl -I https://api.newrelic.com/graphql
   ping api.newrelic.com
   ```

3. Check proxy settings:
   ```bash
   # If behind proxy, set:
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   ```

## Tool-Specific Troubleshooting

### NRQL Query Tools

#### Issue: Query Syntax Errors

**Error Message:**
```
NRQL Syntax Error: Expected FROM clause at position 15
```

**Solution:**
1. Validate query syntax:
   ```bash
   # Use query_check tool first
   echo '{"jsonrpc":"2.0","method":"tools/query_check","params":{"query":"SELECT * Transaction"},"id":1}' | ./bin/mcp-server
   ```

2. Common syntax fixes:
   ```nrql
   # Wrong: SELECT * Transaction
   # Right: SELECT * FROM Transaction
   
   # Wrong: SELECT count(*) FROM Transaction WHERE timestamp > 1 hour ago
   # Right: SELECT count(*) FROM Transaction WHERE timestamp > 1 hour AGO
   ```

#### Issue: Query Timeouts

**Error Message:**
```
Query execution exceeded timeout of 30s
```

**Solution:**
1. Increase timeout:
   ```json
   {
     "query": "SELECT * FROM Transaction",
     "timeout": 60
   }
   ```

2. Optimize query:
   ```nrql
   # Add time constraints
   SELECT * FROM Transaction WHERE timestamp > 1 hour ago
   
   # Limit results
   SELECT * FROM Transaction LIMIT 100
   
   # Use sampling
   SELECT * FROM Transaction SAMPLE 1000
   ```

### Dashboard Tools

#### Issue: Dashboard Generation Fails

**Error Message:**
```
Failed to generate dashboard: template 'apm_overview' requires metric 'Transaction'
```

**Solution:**
1. Check available data:
   ```bash
   # List schemas to see available event types
   echo '{"jsonrpc":"2.0","method":"tools/discovery.list_schemas","params":{},"id":1}' | ./bin/mcp-server
   ```

2. Use appropriate template:
   ```json
   {
     "template_type": "custom",
     "metrics": ["PageView", "JavaScriptError"]
   }
   ```

### Alert Tools

#### Issue: Alert Creation Fails

**Error Message:**
```
Failed to create alert: Policy 'My Policy' not found
```

**Solution:**
1. List existing policies:
   ```bash
   # Get all alert policies
   echo '{"jsonrpc":"2.0","method":"tools/list_alerts","params":{},"id":1}' | ./bin/mcp-server
   ```

2. Create with existing policy or let it auto-create:
   ```json
   {
     "metric": "Transaction",
     "condition": "response.duration > 1",
     "policy_id": "existing_policy_id"
   }
   ```

### Discovery Tools

#### Issue: Schema Discovery Returns Empty

**Error Message:**
```
No schemas found in account
```

**Solution:**
1. Verify data exists:
   ```bash
   # Run a basic query
   echo '{"jsonrpc":"2.0","method":"tools/query_nrdb","params":{"query":"SHOW EVENT TYPES"},"id":1}' | ./bin/mcp-server
   ```

2. Check time range:
   ```json
   {
     "method": "tools/discovery.list_schemas",
     "params": {
       "since": "7 days ago"
     }
   }
   ```

## Performance Troubleshooting

### Issue: Slow Query Performance

**Symptoms:**
- Queries taking >10 seconds
- Timeouts on simple queries

**Solution:**
1. Profile query performance:
   ```bash
   # Enable query profiling
   export NRQL_PROFILE=true
   
   # Check query estimation
   echo '{"jsonrpc":"2.0","method":"tools/query_check","params":{"query":"SELECT * FROM Transaction","estimate_cost":true},"id":1}' | ./bin/mcp-server
   ```

2. Optimize queries:
   - Add time constraints
   - Use LIMIT clause
   - Avoid SELECT *
   - Use FACET instead of multiple queries

### Issue: High Memory Usage

**Symptoms:**
- Server using >1GB RAM
- Out of memory errors

**Solution:**
1. Monitor memory usage:
   ```bash
   # During execution
   top -p $(pgrep mcp-server)
   
   # Check for leaks
   go tool pprof http://localhost:6060/debug/pprof/heap
   ```

2. Limit result sizes:
   ```bash
   # Set max results
   export MAX_QUERY_RESULTS=1000
   
   # Enable result streaming
   export ENABLE_STREAMING=true
   ```

### Issue: Connection Pool Exhaustion

**Error Message:**
```
Too many connections to New Relic API
```

**Solution:**
1. Configure connection limits:
   ```bash
   # Reduce concurrent connections
   export MAX_CONNECTIONS=10
   export CONNECTION_TIMEOUT=30s
   ```

2. Enable connection reuse:
   ```bash
   export ENABLE_CONNECTION_POOLING=true
   ```

## Debug Mode and Logging

### Enable Debug Logging

```bash
# Maximum verbosity
export LOG_LEVEL=DEBUG
export MCP_DEBUG=true

# Log to file
./bin/mcp-server 2>&1 | tee debug.log

# Filter specific components
export LOG_COMPONENTS="discovery,query"
```

### Analyze Logs

```bash
# Check for errors
grep -i error debug.log

# Track request flow
grep -E "request_id|correlation_id" debug.log

# Performance analysis
grep -E "duration|elapsed|took" debug.log
```

### Enable Request/Response Logging

```bash
# Log all MCP messages
export LOG_MCP_MESSAGES=true

# Pretty print JSON
./bin/mcp-server 2>&1 | jq -r 'select(.msg == "mcp_message")'
```

## Error Code Reference

### MCP Protocol Errors

| Code | Error | Solution |
|------|-------|----------|
| -32700 | Parse error | Check JSON syntax |
| -32600 | Invalid request | Verify request structure |
| -32601 | Method not found | Check tool name spelling |
| -32602 | Invalid params | Verify parameter types |
| -32603 | Internal error | Check server logs |

### New Relic API Errors

| Status | Error | Solution |
|--------|-------|----------|
| 401 | Unauthorized | Check API key |
| 403 | Forbidden | Verify permissions |
| 404 | Not Found | Check account ID |
| 429 | Rate Limited | Reduce request rate |
| 500 | Server Error | Retry with backoff |

### Application Errors

| Code | Error | Solution |
|------|-------|----------|
| NR001 | Invalid NRQL | Check query syntax |
| NR002 | No data | Verify time range |
| NR003 | Timeout | Increase timeout or optimize query |
| NR004 | Account mismatch | Check account configuration |
| NR005 | Permission denied | Verify API key permissions |

## Frequently Asked Questions

### Q: How do I run the server in Docker?

**A:** Use the provided Dockerfile:
```bash
# Build image
docker build -t mcp-server-newrelic .

# Run with environment file
docker run --env-file .env mcp-server-newrelic
```

### Q: Can I use this with EU region accounts?

**A:** Yes! EU region is fully supported. Simply set:
```bash
# Set EU endpoint (experimental)
export NEW_RELIC_REGION=eu
export NEW_RELIC_API_URL=https://api.eu.newrelic.com/graphql
```

### Q: How do I add custom tools?

**A:** Follow the pattern in `pkg/interface/mcp/tools_*.go`:
1. Create new tool file
2. Register in `registerTools()`
3. Implement handler function
4. Add tests

### Q: Why am I getting "context deadline exceeded"?

**A:** This indicates a timeout. Solutions:
1. Increase global timeout: `export REQUEST_TIMEOUT=60s`
2. Set per-request timeout in tool parameters
3. Optimize your queries

### Q: How do I enable caching?

**A:** Configure Redis (optional):
```bash
export REDIS_URL=redis://localhost:6379
export ENABLE_CACHE=true
export CACHE_TTL=300
```

### Q: Can I use multiple New Relic accounts?

**A:** Yes, through cross-account queries:
```json
{
  "query": "SELECT * FROM Transaction",
  "accounts": [123456, 789012]
}
```

## Getting Help

### Before Reporting Issues

1. Run diagnostics: `make diagnose`
2. Check this guide for solutions
3. Search existing issues on GitHub
4. Collect debug logs

### Reporting Issues

Include the following in bug reports:

```markdown
**Environment:**
- OS: [e.g., macOS 13.0, Ubuntu 22.04]
- Go version: [output of `go version`]
- Server version: [output of `./bin/mcp-server --version`]

**Steps to Reproduce:**
1. [First step]
2. [Second step]

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happens]

**Logs:**
```
[Paste relevant log excerpts]
```

**Additional Context:**
[Any other relevant information]
```

### Community Support

- GitHub Issues: Report bugs and feature requests
- Discussions: Ask questions and share tips
- Wiki: Community-maintained documentation

### Commercial Support

For enterprise support options, contact New Relic support with reference to the MCP Server integration.

## Quick Reference Card

```bash
# Essential Commands
make diagnose              # Check environment
make diagnose-fix         # Auto-fix issues
make run-mock            # Test without New Relic
make dev                 # Development mode
make test               # Run tests

# Debug Environment Variables
LOG_LEVEL=DEBUG         # Maximum logging
MCP_DEBUG=true         # MCP protocol debug
MOCK_MODE=true        # Run without New Relic
REQUEST_TIMEOUT=60s   # Increase timeouts

# Common Fixes
chmod +x ./bin/*      # Fix permissions
source .env          # Load environment
make clean build    # Rebuild everything
```

Remember: Most issues can be resolved by running `make diagnose-fix` first!
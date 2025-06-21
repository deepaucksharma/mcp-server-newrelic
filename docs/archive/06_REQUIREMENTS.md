# System Requirements

This document outlines the system requirements and dependencies for running the New Relic MCP Server.

## Overview

The New Relic MCP Server is a Go-based application that provides AI assistants with access to New Relic observability data through the Model Context Protocol (MCP).

## Runtime Requirements

### Operating Systems

The server is compatible with:

- **Linux** (x86_64, ARM64)
  - Ubuntu 20.04 LTS or later
  - Debian 10 or later
  - RHEL/CentOS 8 or later
  - Alpine Linux 3.14 or later
  
- **macOS** (x86_64, ARM64)
  - macOS 11 (Big Sur) or later
  - Native Apple Silicon support
  
- **Windows** (x86_64)
  - Windows 10 version 1909 or later
  - Windows Server 2019 or later
  - WSL2 recommended for development

### Go Version

- **Minimum**: Go 1.21
- **Recommended**: Go 1.23 or later
- **Tested with**: Go 1.23.0 - 1.24.3

### Memory Requirements

- **Minimum**: 256MB RAM
- **Recommended**: 512MB RAM
- **For production**: 1GB+ RAM (depends on query volume and caching)

### Disk Space

- **Binary size**: ~15MB
- **With dependencies**: ~50MB
- **Logs and cache**: 100MB-1GB (configurable)

### Network Requirements

- **Outbound HTTPS**: Required for New Relic API access
- **Ports**:
  - Default HTTP: 8080 (configurable)
  - Default SSE: 8081 (configurable)
  - STDIO mode: No network required

## Build Requirements

### Development Tools

```bash
# Required
- Go 1.21+ compiler
- Git 2.25+
- Make 4.0+

# Optional but recommended
- golangci-lint 1.54+
- go-mockgen 1.6+
- air (for hot reload)
```

### Build Dependencies

All dependencies are managed via Go modules. Key dependencies include:

```go
// Core dependencies
github.com/mark3labs/mcp-go v0.32.0        // MCP protocol implementation
github.com/joho/godotenv v1.5.1            // Environment configuration
github.com/spf13/cobra v1.8.0              // CLI framework
github.com/spf13/viper v1.18.2             // Configuration management

// New Relic integration
github.com/newrelic/go-agent/v3 v3.39.0    // APM integration

// Transport and API
github.com/gorilla/mux v1.8.1              // HTTP routing
github.com/rs/cors v1.10.1                 // CORS support
golang.org/x/time v0.5.0                   // Rate limiting

// Observability
go.opentelemetry.io/otel v1.36.0           // OpenTelemetry
google.golang.org/grpc v1.73.0             // gRPC support

// Caching (optional)
github.com/go-redis/redis/v8 v8.11.5       // Redis client

// Testing
github.com/stretchr/testify v1.10.0        // Test assertions
```

## New Relic Requirements

### Account Requirements

- **New Relic Account**: Required for real data access
- **API Key**: User API key with following permissions:
  - NRQL query access
  - Read access to accounts
  - (Optional) Alert policy management
  - (Optional) Dashboard management

### API Endpoints

The server connects to:

```bash
# US Region (default)
https://api.newrelic.com/graphql

# EU Region
https://api.eu.newrelic.com/graphql
```

### Data Access

Minimum New Relic subscription features:
- **NRDB Access**: Core requirement
- **GraphQL API**: Required for all operations
- **Data Retention**: Affects historical query ranges

## Container Requirements

### Docker

```dockerfile
# Minimum Docker version: 20.10
# Base image requirements
FROM golang:1.23-alpine AS builder
FROM alpine:3.19

# Runtime dependencies in container
- ca-certificates (for HTTPS)
- tzdata (for timezone support)
```

### Kubernetes

```yaml
# Minimum Kubernetes version: 1.21
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "1Gi"
    cpu: "500m"
```

## Integration Requirements

### Claude Desktop

- **Version**: Claude Desktop 0.4.0+
- **Configuration**: Requires MCP server configuration in claude_desktop_config.json

### Other LLM Integrations

- **MCP Protocol**: Any client supporting MCP 0.9.0+
- **Transport**: STDIO, HTTP, or SSE support required

## Optional Dependencies

### Redis (for distributed caching)

- **Version**: Redis 6.0+
- **Memory**: 100MB minimum
- **Connection**: TCP/IP or Unix socket

### Monitoring

- **New Relic Go Agent**: For self-monitoring (optional)
- **OpenTelemetry Collector**: For custom telemetry (optional)

## Environment Variables

### Required for Production

```bash
NEW_RELIC_API_KEY=<your-user-api-key>
NEW_RELIC_ACCOUNT_ID=<your-account-id>
```

### Optional Configuration

```bash
# Region selection
NEW_RELIC_REGION=US|EU (default: US)

# Logging
LOG_LEVEL=debug|info|warn|error (default: info)
LOG_FORMAT=json|text (default: json)

# Transport
MCP_TRANSPORT=stdio|http|sse (default: stdio)
HTTP_PORT=8080 (default: 8080)

# Performance
CACHE_ENABLED=true|false (default: true)
CACHE_TTL=300 (seconds, default: 300)
MAX_CONCURRENT_QUERIES=10 (default: 10)

# Security
ALLOWED_ORIGINS=* (CORS, default: *)
API_TIMEOUT=30 (seconds, default: 30)
```

## Security Requirements

### Network Security

- TLS 1.2+ for all external connections
- No inbound connections required for STDIO mode
- HTTP mode should use reverse proxy with TLS

### API Key Security

- Never commit API keys to version control
- Use environment variables or secure key management
- Rotate keys regularly
- Use least-privilege API keys

### Runtime Security

- Run as non-root user
- Use read-only filesystem where possible
- Limit network access to required endpoints only

## Performance Considerations

### Query Performance

- NRQL query timeout: 30 seconds default
- Concurrent query limit: 10 default
- Rate limiting: Respects New Relic API limits

### Resource Usage

- CPU: Generally low, spikes during query processing
- Memory: Grows with cache size and concurrent queries
- Network: Minimal, only during API calls

## Compatibility Matrix

| Component | Minimum Version | Recommended | Maximum Tested |
|-----------|----------------|-------------|----------------|
| Go | 1.21 | 1.23+ | 1.24.3 |
| Docker | 20.10 | 24.0+ | 25.0 |
| Kubernetes | 1.21 | 1.28+ | 1.29 |
| Redis | 6.0 | 7.0+ | 7.2 |
| Claude Desktop | 0.4.0 | Latest | 0.5.0 |

## Known Limitations

1. **Platform-Specific**:
   - Windows: Full functionality in WSL2, limited native support
   - ARM64: Full support but not extensively tested

2. **Scaling Limitations**:
   - Single-instance only (no clustering support)
   - In-memory caching only (Redis optional)
   - No connection pooling for GraphQL

3. **Feature Limitations**:
   - Only ~15% of planned features implemented
   - No support for New Relic One apps
   - Limited to GraphQL API capabilities

## Verification

To verify your system meets requirements:

```bash
# Check Go version
go version

# Verify build tools
make --version
git --version

# Test New Relic connectivity
curl -H "Api-Key: YOUR_KEY" https://api.newrelic.com/graphql

# Run system diagnostics
make diagnose
```

## Support Matrix

- **Active Development**: main branch
- **Stable Release**: No stable release yet (v0.1.0 in development)
- **LTS Planning**: Not yet determined
- **EOL Policy**: To be defined post-1.0

## Next Steps

1. Review [Installation Guide](02_INSTALLATION.md) for setup instructions
2. Check [Configuration Guide](03_CONFIGURATION.md) for detailed settings
3. See [Getting Started](01_GETTING_STARTED.md) for quick start guide
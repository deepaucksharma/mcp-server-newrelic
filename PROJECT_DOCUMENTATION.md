# New Relic MCP Server - Complete Project Documentation

## Table of Contents
1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Recent Cleanup Summary](#recent-cleanup-summary)
4. [Getting Started](#getting-started)
5. [Development Guide](#development-guide)
6. [Testing](#testing)
7. [Deployment](#deployment)
8. [Troubleshooting](#troubleshooting)
9. [Future Roadmap](#future-roadmap)

---

## Project Overview

The Universal Data Synthesizer (UDS) is an AI-powered system for New Relic that automatically discovers, analyzes, and visualizes data from NRDB. It implements MCP (Model Context Protocol) for seamless integration with AI assistants like Claude and GitHub Copilot.

### Current State
- **Track 1 (Discovery Core)**: ✅ Complete - Schema discovery, pattern detection, quality assessment
- **Track 2-3 (Intelligence Engine)**: ❌ Not implemented - ML/AI capabilities planned
- **Track 4 (Platform Foundation)**: 🚧 Partial - Auth, APM integration, basic resilience

### Key Features
- Automatic schema discovery from NRDB
- Pattern detection and relationship mining
- Data quality assessment
- MCP protocol support for AI assistants
- REST API for traditional integrations
- New Relic APM instrumentation

---

## Architecture

### High-Level Architecture
```
AI Assistant (Claude/Copilot)
    ↓ MCP Protocol (JSON-RPC)
MCP Server (Go) - pkg/interface/mcp/
    ↓ Internal APIs
Discovery Engine (Go) - pkg/discovery/
    ↓ GraphQL
New Relic NRDB
```

### Directory Structure
```
cmd/
├── api-server/     # REST API server
├── uds/           # CLI tool for UDS
└── uds-discovery/ # Discovery service

pkg/
├── auth/          # Authentication
├── client/        # Client libraries
├── config/        # Configuration
├── discovery/     # Core discovery engine
│   ├── grpc/      # gRPC service layer
│   ├── nrdb/      # NRDB client with resilience
│   ├── patterns/  # Pattern detection
│   ├── quality/   # Quality assessment
│   ├── relationships/ # Relationship mining
│   └── sampling/  # Data sampling strategies
├── interface/     # API and MCP interfaces
│   ├── api/      # REST API handlers
│   └── mcp/      # MCP server implementation
├── state/         # State management
└── telemetry/     # APM integration
```

### Design Patterns

#### Interface-Based Design
All major components use Go interfaces for testability:
```go
type DiscoveryEngine interface {
    DiscoverSchemas(ctx context.Context, filter DiscoveryFilter) ([]Schema, error)
    // ... other methods
}
```

#### Resilience Patterns
- **Circuit Breaker**: Prevents cascading failures
- **Rate Limiter**: Token bucket algorithm
- **Retry Logic**: Exponential backoff with jitter
- **Timeouts**: Context-based cancellation

#### Worker Pool Pattern
Parallel processing for discovery tasks with configurable pool size and graceful shutdown.

---

## Recent Cleanup Summary

### What Was Done (June 2025)

1. **Removed Duplicate Python Implementation**
   - Deleted dual MCP server implementations
   - Removed Python helper scripts and tools
   - Achieved 51% code reduction (~15,000 lines)

2. **Simplified Configuration**
   - Reduced `.env.example` from 245 to 31 lines
   - Removed complex YAML configurations
   - Kept only essential environment variables

3. **Consolidated Architecture**
   - Go-first approach with clear separation
   - Single implementation to maintain
   - Better performance and type safety

4. **Fixed Build System**
   - Updated Makefile for correct targets
   - Fixed import paths
   - Created mock implementations for testing

### Benefits Achieved
- **Performance**: Go implementation is 10x faster
- **Maintainability**: Single codebase to maintain
- **Clarity**: Clear architecture without duplication
- **Reliability**: Type safety and compile-time checks

---

## Getting Started

### Prerequisites
- Go 1.21+ (required)
- Docker & Docker Compose (optional)
- New Relic account with API access

### Environment Setup

1. Clone the repository:
```bash
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic
```

2. Copy and configure environment:
```bash
cp .env.example .env
# Edit .env with your New Relic credentials
```

3. Required environment variables:
```bash
# New Relic API Access
NEW_RELIC_API_KEY=your-user-api-key        # Required
NEW_RELIC_ACCOUNT_ID=your-account-id       # Required
NEW_RELIC_REGION=US                        # US or EU

# Optional APM monitoring
NEW_RELIC_LICENSE_KEY=your-license-key     # For APM
NEW_RELIC_APP_NAME=mcp-server-newrelic     # APM app name
```

### Quick Start

```bash
# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test

# Start the API server
make run

# Or use Docker
docker-compose up
```

---

## Development Guide

### Building Components

```bash
# Build API server
go build -o bin/api-server ./cmd/api-server

# Build UDS CLI
go build -o bin/uds ./cmd/uds

# Build everything
make build
```

### Running Services

#### API Server
```bash
./bin/api-server --port 8080 --enable-cors --enable-swagger
```

#### MCP Server (via API)
The MCP server is embedded in the API server and responds to MCP protocol requests.

### Common Development Tasks

#### Adding a New MCP Tool
1. Define tool schema in `pkg/interface/mcp/tools_*.go`
2. Implement handler method on `MCPServer`
3. Register tool in `RegisterTools()` method
4. Add tests in `pkg/interface/mcp/*_test.go`

#### Extending Discovery Engine
1. Implement new interface in `pkg/discovery/types.go`
2. Add implementation in appropriate package
3. Wire into `Engine` in `pkg/discovery/engine.go`
4. Add configuration in `pkg/config/config.go`

### Code Style
- Follow standard Go conventions
- Use interfaces for major components
- Include comprehensive error handling
- Add APM instrumentation for operations
- Write unit tests for new functionality

---

## Testing

### Test Structure
```
tests/
├── unit/          # Unit tests (in package directories)
├── integration/   # Integration tests
├── e2e/          # End-to-end tests
└── load/         # Performance tests
```

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage

# Run specific test
go test -v -run TestDiscoverSchemas ./pkg/discovery
```

### Test Coverage Goals
- Critical paths: 80%+ coverage
- Overall project: 60%+ coverage
- Integration tests for all MCP tools

### Mocking
Use the mock engine for development:
```bash
export MOCK_MODE=true
make run
```

---

## Deployment

### Docker Deployment

```bash
# Build image
docker build -t mcp-server-newrelic .

# Run with docker-compose
docker-compose up -d

# Check logs
docker-compose logs -f
```

### Kubernetes Deployment

```yaml
# See deployments/k8s/ for manifests
kubectl apply -f deployments/k8s/
```

### Production Checklist
- [ ] Set production environment variables
- [ ] Enable TLS (required for production)
- [ ] Configure APM monitoring
- [ ] Set up proper logging
- [ ] Configure rate limiting
- [ ] Enable circuit breakers
- [ ] Set up health checks
- [ ] Configure backup strategy

---

## Troubleshooting

### Common Issues

#### Build Failures
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

#### Connection Issues
- Verify NEW_RELIC_API_KEY is correct
- Check NEW_RELIC_REGION (US or EU)
- Ensure network connectivity to New Relic

#### Performance Issues
- Check APM metrics in New Relic
- Review circuit breaker status
- Monitor rate limiter metrics
- Check cache hit rates

### Debug Mode
```bash
export LOG_LEVEL=DEBUG
export DEV_MODE=true
make run
```

### Health Checks
```bash
# API health check
curl http://localhost:8080/health

# Detailed health
curl http://localhost:8080/health/detailed
```

---

## Future Roadmap

### Phase 1: Complete Track 1 (Current)
- [x] Schema discovery
- [x] Pattern detection
- [x] Quality assessment
- [ ] Enhanced relationship mining
- [ ] Advanced sampling strategies

### Phase 2: Track 2-3 Implementation
- [ ] Python ML service integration
- [ ] Dashboard generation AI
- [ ] NRQL optimization engine
- [ ] Anomaly detection
- [ ] Predictive analytics

### Phase 3: Production Hardening
- [ ] Complete security audit
- [ ] Performance optimization
- [ ] Multi-region support
- [ ] Enhanced caching strategies
- [ ] Comprehensive documentation

### Phase 4: Enterprise Features
- [ ] Multi-tenant support
- [ ] Role-based access control
- [ ] Audit logging
- [ ] SLA monitoring
- [ ] Cost optimization

---

## Contributing

### Development Workflow
1. Create feature branch from `main`
2. Implement changes with tests
3. Run `make test` and `make lint`
4. Submit PR with description
5. Ensure CI passes

### Code Review Checklist
- [ ] Tests included and passing
- [ ] Documentation updated
- [ ] No security vulnerabilities
- [ ] APM instrumentation added
- [ ] Error handling comprehensive

---

## License

This project is proprietary software. See LICENSE file for details.

---

## Support

- GitHub Issues: https://github.com/deepaucksharma/mcp-server-newrelic/issues
- Internal Slack: #uds-development
- Documentation: This file and CLAUDE.md

---

*Last Updated: June 2025*
*Version: 2.0.0*
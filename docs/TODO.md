# TODO - Immediate Action Items

This document tracks immediate tasks and priorities for the New Relic MCP Server project. Items are organized by urgency and impact.

## 🔴 Critical (This Week)

### 1. Fix Build and Runtime Issues
- [ ] Fix main.go state manager initialization
  - Issue: State manager factory expects different config structure
  - File: `cmd/mcp-server/main.go:82-101`
  
- [ ] Resolve missing tool registrations
  - Issue: Some tools defined but not registered
  - Check: `pkg/interface/mcp/tools_discovery.go`
  
- [ ] Fix mock mode detection
  - Issue: Server tries Redis even in mock mode
  - File: `cmd/mcp-server/main.go:77-81`

### 2. Environment Configuration
- [ ] Update .env.example with all required variables
  ```bash
  NEW_RELIC_API_KEY=
  NEW_RELIC_ACCOUNT_ID=
  JWT_SECRET=
  API_KEY_SALT=
  ```
  
- [ ] Add environment validation in config.Load()
  - Refuse to start without required secrets
  - Clear error messages for missing config

### 3. Immediate Testing
- [ ] Add basic unit tests for critical paths
  - [ ] Test each tool handler with valid inputs
  - [ ] Test each tool handler with invalid inputs
  - [ ] Test mock mode responses
  
- [ ] Create integration test for MCP protocol
  ```go
  // Test basic MCP request/response
  func TestMCPInitialize(t *testing.T)
  func TestMCPToolCall(t *testing.T)
  ```

## 🟡 High Priority (Next 2 Weeks)

### 4. Complete Missing Tools
- [ ] Alert Policy Management
  ```go
  - handleCreateAlertPolicy()
  - handleUpdateAlertPolicy()
  - handleDeleteAlertPolicy()
  ```
  
- [ ] Alert Condition Management
  ```go
  - handleCreateAlertCondition()
  - handleUpdateAlertCondition()
  - handleCloseIncident()
  ```
  
- [ ] Bulk Operations
  ```go
  - handleBulkTagEntities()
  - handleBulkCreateMonitors()
  - handleBulkUpdateDashboards()
  ```

### 5. Error Handling Improvements
- [ ] Wrap all New Relic API calls with proper error handling
- [ ] Add network timeout handling
- [ ] Implement retry logic for transient failures
- [ ] Create user-friendly error messages

### 6. Logging Infrastructure
- [ ] Replace fmt.Printf with structured logging
- [ ] Add request ID tracking
- [ ] Implement log levels (DEBUG, INFO, WARN, ERROR)
- [ ] Add performance metrics to logs

## 🟢 Medium Priority (Next Month)

### 7. CI/CD Pipeline
- [ ] Create `.github/workflows/ci.yml`
  ```yaml
  - Build and test on PR
  - Run linting
  - Check test coverage
  - Security scanning
  ```
  
- [ ] Add Dockerfile
- [ ] Create docker-compose.yml for local development
- [ ] Set up automated releases

### 8. Documentation Updates
- [ ] Update API documentation for all tools
- [ ] Create deployment guide
- [ ] Add troubleshooting section
- [ ] Create video walkthrough

### 9. Performance Optimization
- [ ] Implement caching layer
  - [ ] Query result caching
  - [ ] Schema discovery caching
  - [ ] Dashboard metadata caching
  
- [ ] Add connection pooling for NerdGraph
- [ ] Optimize large result handling
- [ ] Add request batching

### 10. Security Hardening
- [ ] Add input sanitization for all parameters
- [ ] Implement rate limiting
- [ ] Add API key rotation support
- [ ] Security audit of all endpoints

## 📋 Backlog (Future)

### Enhanced Features
- [ ] Multi-account support
- [ ] EU region support
- [ ] Streaming responses for large datasets
- [ ] WebSocket transport
- [ ] Query result pagination

### Developer Experience
- [ ] Create development CLI for testing
- [ ] Add tool scaffolding generator
- [ ] Improve mock data generation
- [ ] Create playground environment

### Monitoring & Observability
- [ ] Add Prometheus metrics
- [ ] Implement distributed tracing
- [ ] Create operational dashboard
- [ ] Add health check endpoints

## Quick Wins (Can be done anytime)

- [ ] Add GitHub issue templates
- [ ] Update .gitignore for Go specifics
- [ ] Add pre-commit hooks for formatting
- [ ] Create CONTRIBUTING.md
- [ ] Add badges to README (build status, coverage, etc.)

## Code Cleanup

- [ ] Remove deprecated Python references
- [ ] Clean up unused imports
- [ ] Standardize error messages
- [ ] Add missing comments/documentation
- [ ] Remove TODO comments after implementation

## Testing Checklist

For each implemented feature:
- [ ] Unit tests written
- [ ] Integration tests written
- [ ] Mock mode tested
- [ ] Error cases tested
- [ ] Documentation updated
- [ ] Example added to README

## Review Checklist

Before marking any task complete:
- [ ] Code follows Go best practices
- [ ] Tests pass
- [ ] No security vulnerabilities
- [ ] Documentation updated
- [ ] Peer review completed

---

## How to Use This Document

1. **Pick a task** from Critical or High Priority sections
2. **Create a branch** with descriptive name (e.g., `fix/state-manager-init`)
3. **Update this document** when starting/completing tasks
4. **Add new items** as they're discovered
5. **Move completed items** to CHANGELOG.md

## Task Assignment

Use GitHub issues to assign tasks. Reference this document in issue descriptions.

Example:
```
Title: Fix state manager initialization
Description: As per TODO.md item #1, the state manager initialization in main.go needs updating
Labels: bug, critical
```

---

Last Updated: June 2025
Next Review: Weekly during team meetings
# Contributing to New Relic MCP Server

Thank you for your interest in contributing to the New Relic MCP Server! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:
- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Respect differing viewpoints and experiences

## How to Contribute

### Reporting Issues

1. **Check existing issues** first to avoid duplicates
2. Use issue templates when available
3. Include:
   - Clear description of the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)
   - Relevant logs or error messages

### Suggesting Features

1. Open a **discussion** first for major features
2. Explain the use case and benefits
3. Consider discovery-first principles
4. Be open to feedback and alternatives

### Contributing Code

#### Prerequisites

- Go 1.21 or later
- Make
- Git
- New Relic account (for testing)

#### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/YOUR-USERNAME/mcp-server-newrelic.git
cd mcp-server-newrelic

# Add upstream remote
git remote add upstream https://github.com/deepaucksharma/mcp-server-newrelic.git

# Install dependencies
make install-tools

# Set up environment
cp .env.example .env
# Edit .env with your test credentials

# Run in mock mode for development
make run-mock
```

#### Development Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow the coding standards below
   - Add/update tests as needed
   - Update documentation

3. **Test your changes**
   ```bash
   # Run all tests
   make test

   # Run specific tests
   go test ./pkg/interface/mcp/...

   # Check code quality
   make lint
   make format
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add amazing new feature"
   ```
   Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` New feature
   - `fix:` Bug fix
   - `docs:` Documentation changes
   - `test:` Test additions/modifications
   - `refactor:` Code refactoring
   - `chore:` Maintenance tasks

5. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

### Pull Request Guidelines

1. **PR Title**: Use conventional commit format
2. **Description**: 
   - Explain what changes you made
   - Why you made them
   - Link related issues
3. **Checklist**:
   - [ ] Tests pass (`make test`)
   - [ ] Code is formatted (`make format`)
   - [ ] Linting passes (`make lint`)
   - [ ] Documentation updated
   - [ ] Follows discovery-first principles

## Coding Standards

### Go Code Style

1. **Follow standard Go conventions**
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://golang.org/doc/effective_go.html)
   - Use meaningful variable names

2. **Error Handling**
   ```go
   // Always wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to discover schemas: %w", err)
   }
   ```

3. **Comments**
   ```go
   // DiscoverSchemas explores available event types without assumptions.
   // It returns a list of schemas with metadata about data quality.
   func DiscoverSchemas(ctx context.Context) ([]Schema, error) {
   ```

### Discovery-First Principles

When adding new tools or features:

1. **Never assume data structures**
   ```go
   // Bad: Assumes 'appName' exists
   query := "SELECT * FROM Transaction WHERE appName = 'checkout'"

   // Good: Discover first
   serviceAttr := discover.FindServiceAttribute(ctx)
   query := fmt.Sprintf("SELECT * FROM Transaction WHERE %s = 'checkout'", serviceAttr)
   ```

2. **Make tools granular**
   - Each tool does ONE thing
   - Tools compose into workflows
   - Clear inputs and outputs

3. **Add discovery metadata**
   ```go
   tool := Tool{
       Name: "discover.new_capability",
       DiscoveryLevel: "none", // No assumptions required
       AdaptsToSchema: true,   // Adjusts based on findings
   }
   ```

### Testing Requirements

1. **Unit Tests** for all new functions
2. **Integration Tests** for tool handlers
3. **Mock Mode Support** for all tools
4. **Test Coverage** - maintain or improve

Example test:
```go
func TestDiscoverSchemas(t *testing.T) {
    tests := []struct {
        name    string
        setup   func()
        want    []Schema
        wantErr bool
    }{
        {
            name: "discovers basic schemas",
            setup: func() {
                // Setup mock
            },
            want: []Schema{{Name: "Transaction"}},
        },
    }
    // ... test implementation
}
```

### Documentation

1. **Code Documentation**
   - All exported functions must have comments
   - Include examples for complex features

2. **User Documentation**
   - Update relevant .md files
   - Add examples for new tools
   - Update API reference

3. **Architecture Documentation**
   - Document design decisions
   - Update diagrams if needed

## Tool Development Guide

### Adding a New Tool

1. **Choose the right category**
   - `tools_discovery.go` - Data exploration
   - `tools_query.go` - NRQL execution
   - `tools_dashboard.go` - Dashboard operations
   - `tools_alerts.go` - Alert management
   - `tools_governance.go` - Platform governance

2. **Implement the handler**
   ```go
   func (s *Server) handleYourNewTool(ctx context.Context, params map[string]interface{}) (interface{}, error) {
       // 1. Validate parameters
       requiredParam, ok := params["required_field"].(string)
       if !ok || requiredParam == "" {
           return nil, fmt.Errorf("required_field is required")
       }

       // 2. Check mock mode
       if s.nrClient == nil {
           return mockResponse, nil
       }

       // 3. Perform discovery if needed
       discovered, err := s.discovery.DiscoverRelevantData(ctx)
       if err != nil {
           return nil, fmt.Errorf("discovery failed: %w", err)
       }

       // 4. Execute operation
       result, err := s.executeOperation(ctx, discovered, requiredParam)
       if err != nil {
           return nil, fmt.Errorf("operation failed: %w", err)
       }

       return result, nil
   }
   ```

3. **Register the tool**
   ```go
   s.tools.Register(Tool{
       Name:        "category.your_tool",
       Description: "Clear description of what it does",
       Parameters:  toolParams,
       Handler:     s.handleYourNewTool,
   })
   ```

4. **Add tests**
   - Unit test for the handler
   - Mock mode test
   - Integration test if applicable

5. **Document the tool**
   - Add to API reference
   - Include examples
   - Update relevant guides

## Review Process

1. **Automated Checks** - Must pass CI
2. **Code Review** - At least one maintainer
3. **Testing** - Manual testing for significant changes
4. **Documentation** - Must be complete

## Getting Help

- **Development Setup**: See [Development Guide](./docs/guides/development.md)
- **Architecture Questions**: Review [Architecture Docs](./docs/architecture/)
- **Discussions**: Use GitHub Discussions
- **Real-time Chat**: [Discord/Slack] (if available)

## Recognition

Contributors are recognized in:
- Release notes
- Contributors list
- Project documentation

Thank you for contributing to making observability more intelligent and accessible!
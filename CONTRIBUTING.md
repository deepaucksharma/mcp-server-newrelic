# Contributing to New Relic MCP Server

First off, thank you for considering contributing to the New Relic MCP Server! It's people like you that make this project such a great tool for the community. We welcome contributions from everyone, regardless of their experience level.

This document provides guidelines for contributing to the project. Following these guidelines helps communicate that you respect the time of the developers managing and developing this open source project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment Setup](#development-environment-setup)
- [Code Style and Standards](#code-style-and-standards)
- [Documentation Requirements](#documentation-requirements)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Code Review Guidelines](#code-review-guidelines)
- [Communication Channels](#communication-channels)
- [Issue Reporting Guidelines](#issue-reporting-guidelines)
- [Recognition and Credits](#recognition-and-credits)

## Getting Started

### What Can I Contribute?

There are many ways to contribute:

- **Bug fixes**: Found a bug? We'd love your help fixing it!
- **Feature implementation**: Check our [issue tracker](https://github.com/your-org/mcp-server-newrelic/issues) for feature requests
- **Documentation**: Help us improve our docs or add examples
- **Testing**: Add tests to increase our coverage
- **Performance improvements**: Help optimize the codebase
- **Bug reports**: Let us know if something isn't working
- **Feature requests**: Have an idea? We'd love to hear it!

### Prerequisites

Before you begin, ensure you have:

- Go 1.21 or higher installed
- Git for version control
- A GitHub account
- Basic understanding of the Model Context Protocol (MCP)
- Familiarity with New Relic's observability platform is helpful but not required

### First-Time Contributors

Looking for a good first issue? Check out issues labeled with `good first issue` or `help wanted`. These are specifically curated for newcomers to the project.

## Development Environment Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/mcp-server-newrelic.git
cd mcp-server-newrelic
git remote add upstream https://github.com/original-org/mcp-server-newrelic.git
```

### 2. Environment Configuration

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your New Relic credentials (optional for mock mode)
# For development, you can use mock mode without credentials
```

### 3. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
make install-tools
```

### 4. Verify Setup

```bash
# Run diagnostics to check your environment
make diagnose

# If issues are found, try auto-fix
make diagnose-fix

# Build the project
make build

# Run tests
make test
```

### 5. Development Workflow

```bash
# Run in development mode with auto-reload
make dev

# Run in mock mode (no New Relic connection needed)
make run-mock

# Run linters before committing
make lint

# Format code
make format
```

## Code Style and Standards

### Go Code Style

We follow the standard Go coding conventions with some additional guidelines:

1. **Formatting**: Use `gofmt` and `goimports` (automated via `make format`)
2. **Linting**: Code must pass all linters (`make lint`)
3. **Naming**: Follow Go naming conventions
   - Exported names start with capital letters
   - Use camelCase, not snake_case
   - Acronyms should be all caps (e.g., `URL`, `ID`, `NRQL`)

### Code Organization

```go
// Package comments should be present
package mcp

// Imports grouped and ordered:
import (
    // Standard library
    "context"
    "fmt"
    
    // Third-party
    "github.com/stretchr/testify/assert"
    
    // Project imports
    "github.com/your-org/mcp-server-newrelic/pkg/config"
)

// Constants grouped at the top
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
)

// Interfaces before structs
type Client interface {
    Query(ctx context.Context, nrql string) (*Result, error)
}

// Structs with clear documentation
// Server represents the MCP server implementation
type Server struct {
    client Client
    // fields should be commented if not obvious
}
```

### Error Handling

```go
// Always wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to query NRDB: %w", err)
}

// Use custom error types for known conditions
var ErrInvalidQuery = errors.New("invalid NRQL query")
```

### Logging

```go
// Use structured logging
log.WithFields(log.Fields{
    "tool":    toolName,
    "account": accountID,
}).Debug("executing tool")

// Never log sensitive data
// BAD: log.Printf("API Key: %s", apiKey)
// GOOD: log.Printf("Using API key ending in: ...%s", apiKey[len(apiKey)-4:])
```

## Documentation Requirements

### Code Documentation

1. **Package Documentation**: Every package must have a package comment
2. **Exported Types**: All exported types, functions, and methods must be documented
3. **Complex Logic**: Add inline comments for non-obvious code
4. **Examples**: Include examples in doc comments where helpful

```go
// Package discovery provides tools for analyzing New Relic data schemas
// and discovering relationships between different data types.
package discovery

// Engine analyzes NRDB schemas to discover data relationships and patterns.
// It provides methods for profiling attributes, finding relationships,
// and assessing data quality.
//
// Example:
//
//     engine := discovery.New(client)
//     profile, err := engine.ProfileAttribute(ctx, "Transaction", "duration")
//     if err != nil {
//         log.Fatal(err)
//     }
//     fmt.Printf("Attribute type: %s\n", profile.DataType)
type Engine struct {
    // ...
}
```

### Documentation Files

When updating features, also update:

1. **API Documentation**: Update relevant sections in `docs/API_REFERENCE_V2.md`
2. **Architecture Docs**: Update `docs/ARCHITECTURE.md` for structural changes
3. **Migration Guide**: Update `docs/MIGRATION_GUIDE.md` for breaking changes
4. **README**: Update the main README.md if adding major features

### Commit Messages

Follow the Conventional Commits specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding missing tests
- `chore`: Changes to build process or auxiliary tools

Example:
```
feat(discovery): add relationship mining to discovery engine

- Implement FindRelationships method to discover data connections
- Add graph-based analysis for relationship strength
- Include unit tests and documentation

Closes #123
```

## Testing Requirements

### Test Coverage

- Aim for at least 80% test coverage for new code
- All new features must include tests
- Bug fixes should include a test that reproduces the issue

### Types of Tests

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **Mock Tests**: Ensure all tools work in mock mode

### Writing Tests

```go
func TestHandleQueryNRDB(t *testing.T) {
    tests := []struct {
        name    string
        params  map[string]interface{}
        want    interface{}
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid query",
            params: map[string]interface{}{
                "query": "SELECT count(*) FROM Transaction",
            },
            want:    mockQueryResult(),
            wantErr: false,
        },
        {
            name:    "missing query parameter",
            params:  map[string]interface{}{},
            wantErr: true,
            errMsg:  "query parameter is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := &Server{} // Mock mode
            got, err := s.handleQueryNRDB(context.Background(), tt.params)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./pkg/discovery/...

# Run with coverage
make test-coverage

# Run specific test
go test -v -run TestHandleQueryNRDB ./pkg/interface/mcp/
```

## Pull Request Process

### Before Creating a PR

1. **Create an Issue First**: For significant changes, create an issue to discuss the approach
2. **Branch from main**: Always create feature branches from the latest main
3. **One Feature per PR**: Keep PRs focused on a single feature or fix
4. **Update Tests**: Ensure all tests pass and add new ones as needed
5. **Update Documentation**: Update relevant docs for your changes
6. **Run Checks Locally**:
   ```bash
   make format    # Format code
   make lint      # Check code style
   make test      # Run tests
   make build     # Ensure it builds
   ```

### Creating the PR

1. **Title**: Use a clear, descriptive title following commit conventions
2. **Description**: Fill out the PR template completely:
   - What changes were made and why
   - How to test the changes
   - Any breaking changes
   - Related issues

3. **Size**: Keep PRs small and focused (ideally under 500 lines)
4. **Draft PRs**: Use draft PRs for work-in-progress

### PR Checklist

Before marking your PR as ready for review, ensure:

- [ ] Tests pass locally and in CI
- [ ] Code follows project style guidelines
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] PR description is complete
- [ ] No sensitive data in code or commits
- [ ] Breaking changes are clearly marked

## Code Review Guidelines

### For Contributors

1. **Be Patient**: Reviews may take a few days
2. **Be Responsive**: Address feedback promptly
3. **Ask Questions**: If feedback is unclear, ask for clarification
4. **Update Promptly**: Push fixes as new commits (we'll squash on merge)

### For Reviewers

1. **Be Constructive**: Provide helpful, specific feedback
2. **Be Thorough**: Check code, tests, and documentation
3. **Be Timely**: Try to review within 2-3 days
4. **Be Kind**: Remember there's a person behind the code

### Review Checklist

- [ ] Code follows style guidelines
- [ ] Tests are adequate and pass
- [ ] Documentation is updated
- [ ] No security issues
- [ ] Performance impact considered
- [ ] Error handling is appropriate
- [ ] Mock mode is supported

## Communication Channels

### GitHub Issues

- **Bug Reports**: Use the bug report template
- **Feature Requests**: Use the feature request template
- **Questions**: Use the question template or discussions

### Discussions

For broader discussions about the project:
- Architecture decisions
- Feature planning
- Community feedback

### Security Issues

**IMPORTANT**: Never report security issues publicly. Instead:
1. Email security@example.com with details
2. Include steps to reproduce
3. Allow time for a fix before disclosure

## Issue Reporting Guidelines

### Before Creating an Issue

1. **Search First**: Check if the issue already exists
2. **Try Latest Version**: Ensure you're using the latest release
3. **Minimal Reproduction**: Create a minimal example that reproduces the issue

### Creating a Good Issue

Include:
1. **Clear Title**: Summarize the issue concisely
2. **Description**: Detailed explanation of the problem
3. **Steps to Reproduce**: Exact steps to recreate the issue
4. **Expected Behavior**: What should happen
5. **Actual Behavior**: What actually happens
6. **Environment**: Go version, OS, New Relic account type
7. **Logs**: Relevant error messages or logs
8. **Screenshots**: If applicable

Example:
```markdown
**Title**: Query tool returns empty results for valid NRQL

**Description**: 
When using the query_nrdb tool with a valid NRQL query, it returns empty results even though the same query works in the New Relic UI.

**Steps to Reproduce**:
1. Configure MCP server with valid credentials
2. Call query_nrdb with query: "SELECT count(*) FROM Transaction"
3. Observe empty results

**Expected**: Results matching New Relic UI
**Actual**: Empty result set

**Environment**:
- Go version: 1.21
- OS: Ubuntu 22.04
- New Relic: Pro account
```

## Recognition and Credits

We believe in recognizing all contributions, not just code!

### Types of Contributions We Recognize

- Code contributions (features, bug fixes)
- Documentation improvements
- Bug reports and feature requests
- Code reviews and feedback
- Community support and mentoring
- Testing and quality assurance

### How We Recognize Contributors

1. **Contributors File**: All contributors are listed in CONTRIBUTORS.md
2. **Release Notes**: Contributors are mentioned in release notes
3. **GitHub Insights**: Your contributions appear in GitHub's contributor graph
4. **Community Shoutouts**: Regular recognition in community channels

### Becoming a Maintainer

Active contributors may be invited to become maintainers. Maintainers:
- Have write access to the repository
- Help review and merge PRs
- Participate in project planning
- Guide the project's direction

---

Thank you for contributing to New Relic MCP Server! Your efforts help make observability more accessible to the AI community.

If you have questions not covered here, please open an issue or start a discussion. We're here to help!
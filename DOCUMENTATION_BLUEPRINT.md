# Documentation Blueprint

This document serves as the master guide for all documentation in the New Relic MCP Server project. It defines standards, structures, and processes to ensure consistent, high-quality documentation across the codebase.

## Table of Contents

- [Documentation Standards and Guidelines](#documentation-standards-and-guidelines)
- [Documentation Structure and Organization](#documentation-structure-and-organization)
- [Required Sections for Each Document Type](#required-sections-for-each-document-type)
- [Versioning and Maintenance Processes](#versioning-and-maintenance-processes)
- [Code Documentation Standards](#code-documentation-standards)
- [LLM Integration Guidelines](#llm-integration-guidelines)
- [Testing and Validation Requirements](#testing-and-validation-requirements)

## Documentation Standards and Guidelines

### General Principles

1. **Clarity First**: Documentation must be clear, concise, and unambiguous
2. **Audience-Aware**: Write for the intended audience (developers, operators, AI assistants)
3. **Example-Driven**: Include practical examples for every concept
4. **Maintainable**: Documentation must be easy to update as code evolves
5. **Searchable**: Use consistent terminology and structure for easy discovery

### Writing Style

- Use present tense for current behavior
- Use future tense only for planned features with clear timelines
- Prefer active voice over passive voice
- Keep sentences short and paragraphs focused
- Use numbered lists for sequential steps
- Use bullet points for non-sequential items

### Formatting Standards

```markdown
# Top-Level Headers - Document Title
## Section Headers - Major Sections
### Subsection Headers - Topic Areas
#### Detail Headers - Specific Topics

**Bold** - Important terms, warnings
*Italic* - Emphasis, first use of technical terms
`code` - Inline code, commands, file names
```

### Code Blocks

Always specify the language for syntax highlighting:

````markdown
```go
// Go code example
func example() error {
    return nil
}
```

```bash
# Shell commands
make build
```

```yaml
# Configuration files
key: value
```
````

## Documentation Structure and Organization

### Directory Structure

```
docs/
├── README.md                    # Documentation index and navigation
├── API_REFERENCE.md            # Complete API documentation
├── ARCHITECTURE.md             # System design and architecture
├── MIGRATION_GUIDE.md          # Version migration instructions
├── guides/                     # How-to guides and tutorials
│   ├── getting-started.md
│   ├── advanced-usage.md
│   └── troubleshooting.md
├── references/                 # Technical references
│   ├── nrql-syntax.md
│   ├── error-codes.md
│   └── configuration.md
└── development/               # Developer documentation
    ├── contributing.md
    ├── testing-guide.md
    └── release-process.md
```

### File Naming Conventions

- Use lowercase with hyphens: `getting-started.md`
- Be descriptive but concise: `nrql-query-guide.md`
- Version-specific docs: `migration-v1-to-v2.md`
- Keep names stable (use redirects for renames)

### Cross-Reference Standards

```markdown
<!-- Good: Specific link with context -->
See the [NRQL Query Guide](./guides/nrql-queries.md#syntax) for syntax details.

<!-- Bad: Vague reference -->
See other documentation for more info.
```

## Required Sections for Each Document Type

### API Documentation

```markdown
# Tool/Function Name

## Overview
Brief description of what this tool does.

## Parameters
| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| param1 | string | Yes | - | What this parameter does |

## Returns
Description of return value structure and types.

## Examples
### Basic Usage
```json
{
  "tool": "example",
  "params": {
    "param1": "value"
  }
}
```

### Advanced Usage
[More complex example]

## Error Handling
Common errors and how to handle them.

## Related Tools
- [Related Tool 1](link)
- [Related Tool 2](link)
```

### Architecture Documents

```markdown
# Component/System Name

## Purpose
Why this component exists and what problems it solves.

## Design Principles
Key architectural decisions and rationale.

## Components
### Component A
- Responsibility
- Interfaces
- Dependencies

## Data Flow
[Diagram or description of how data moves through the system]

## Configuration
Required and optional configuration parameters.

## Performance Considerations
- Scalability limits
- Resource requirements
- Optimization opportunities

## Security Considerations
- Authentication/Authorization
- Data protection
- Threat model
```

### Guide Documents

```markdown
# Guide Title

## Prerequisites
- Required knowledge
- Required setup
- Required permissions

## Overview
What you'll learn in this guide.

## Step-by-Step Instructions
### Step 1: [Action]
1. Detailed instruction
2. Expected result
3. Troubleshooting tips

### Step 2: [Next Action]
[Continue pattern]

## Validation
How to verify success.

## Common Issues
### Issue 1: [Description]
**Symptom**: What you see
**Cause**: Why it happens
**Solution**: How to fix it

## Next Steps
- Related guides
- Advanced topics
```

## Versioning and Maintenance Processes

### Version Control

1. **Documentation Versioning**
   - Match documentation version to code version
   - Tag documentation releases: `docs-v1.2.0`
   - Maintain version-specific branches for major versions

2. **Change Tracking**
   ```markdown
   <!-- At the top of each document -->
   ---
   version: 1.2.0
   last_updated: 2024-01-15
   authors: [author1, author2]
   ---
   ```

3. **Deprecation Notices**
   ```markdown
   > **⚠️ Deprecated**: This feature is deprecated as of v2.0.0.
   > Use [new feature](link) instead. Will be removed in v3.0.0.
   ```

### Maintenance Schedule

- **Weekly**: Review and update active development docs
- **Monthly**: Audit for accuracy and completeness
- **Quarterly**: Major documentation review and cleanup
- **Per Release**: Update all affected documentation

### Documentation Review Process

1. **Pre-PR Review**
   - Code changes must include doc updates
   - Use documentation checklist in PR template

2. **Technical Review**
   - Accuracy verification by code authors
   - Completeness check by tech lead

3. **Editorial Review**
   - Style and formatting consistency
   - Grammar and clarity check

## Code Documentation Standards

### Go Code Documentation

```go
// Package discovery provides schema analysis and data discovery capabilities
// for New Relic data sources. It enables intelligent exploration of available
// metrics, events, and their relationships.
package discovery

// Engine orchestrates the discovery of New Relic data schemas and relationships.
// It provides methods to analyze data sources, profile attributes, and identify
// connections between different data types.
//
// Example usage:
//
//	engine := discovery.NewEngine(client)
//	schemas, err := engine.ListSchemas(ctx, discovery.ListOptions{
//	    IncludeMetrics: true,
//	    MinQuality: 0.8,
//	})
type Engine struct {
    // client is the New Relic client for API calls
    client Client
    // cache stores discovery results for performance
    cache Cache
}

// ListSchemas returns all available data schemas matching the specified options.
// It analyzes data quality and provides recommendations for each schema.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - opts: Filtering and inclusion options
//
// Returns:
//   - []Schema: List of discovered schemas with quality metrics
//   - error: Any error encountered during discovery
func (e *Engine) ListSchemas(ctx context.Context, opts ListOptions) ([]Schema, error) {
    // Implementation
}
```

### Interface Documentation

```go
// Client defines the contract for New Relic API interactions.
// Implementations must handle authentication, rate limiting, and error recovery.
type Client interface {
    // Query executes an NRQL query and returns results.
    // The query timeout is controlled by the context.
    Query(ctx context.Context, query string) (*QueryResult, error)
    
    // GetSchema retrieves schema information for a specific data type.
    // Returns ErrSchemaNotFound if the schema doesn't exist.
    GetSchema(ctx context.Context, dataType string) (*Schema, error)
}
```

### Error Documentation

```go
var (
    // ErrSchemaNotFound indicates the requested schema doesn't exist
    ErrSchemaNotFound = errors.New("schema not found")
    
    // ErrInvalidQuery indicates the NRQL query syntax is invalid
    ErrInvalidQuery = errors.New("invalid NRQL query")
    
    // ErrRateLimited indicates the API rate limit was exceeded
    ErrRateLimited = errors.New("rate limit exceeded")
)
```

## LLM Integration Guidelines

### AI-Friendly Documentation

1. **Structured Information**
   ```markdown
   ## Tool: query_nrdb
   
   **Purpose**: Execute NRQL queries against New Relic
   **Category**: Data Retrieval
   **Complexity**: Medium
   **Rate Limited**: Yes (100 req/min)
   
   ### Quick Example
   Query last hour of transactions:
   ```json
   {
     "tool": "query_nrdb",
     "params": {
       "query": "SELECT * FROM Transaction SINCE 1 hour ago"
     }
   }
   ```
   ```

2. **Decision Trees**
   ```markdown
   ## Choosing the Right Query Tool
   
   - Need to validate syntax? → Use `query_check`
   - Building query programmatically? → Use `query_builder`
   - Direct NRQL execution? → Use `query_nrdb`
   - Need query suggestions? → Use `discovery.find_relationships`
   ```

3. **Common Patterns**
   ```markdown
   ## Common Workflow: Investigate Performance Issue
   
   1. Discovery: `discovery.list_schemas` → Find relevant data sources
   2. Exploration: `discovery.profile_attribute` → Understand data
   3. Query: `query_builder` → Construct investigation query
   4. Execute: `query_nrdb` → Get results
   5. Visualize: `generate_dashboard` → Create monitoring dashboard
   ```

### LLM-Specific Files

1. **CLAUDE.md** - AI assistant instructions
   - Project overview and current state
   - Critical context and gotchas
   - Development workflow
   - Code patterns and examples

2. **Tool Catalogs**
   ```markdown
   ## Available Tools by Category
   
   ### Data Discovery
   - `discovery.list_schemas` - Find available data
   - `discovery.profile_attribute` - Analyze fields
   - `discovery.find_relationships` - Connect data
   
   ### Query Operations
   - `query_nrdb` - Execute queries
   - `query_check` - Validate syntax
   - `query_builder` - Build queries
   ```

## Testing and Validation Requirements

### Documentation Testing

1. **Link Validation**
   ```bash
   # Check all internal links
   make docs-check-links
   
   # Validate external links
   make docs-check-external
   ```

2. **Code Example Testing**
   ```bash
   # Extract and test all code examples
   make docs-test-examples
   
   # Validate JSON/YAML examples
   make docs-validate-formats
   ```

3. **Completeness Checks**
   - All public APIs documented
   - All tools have examples
   - All errors have descriptions
   - All configs have defaults listed

### Documentation Metrics

Track and maintain:
- **Coverage**: % of public APIs documented
- **Freshness**: Days since last update
- **Quality**: Readability score (aim for grade 8-10)
- **Completeness**: Required sections present
- **Accuracy**: Test example success rate

### Validation Checklist

Before merging documentation:

- [ ] Spell check passed
- [ ] Grammar check passed
- [ ] All links validated
- [ ] Code examples tested
- [ ] Technical review completed
- [ ] Follows style guide
- [ ] Includes required sections
- [ ] Version info updated
- [ ] Cross-references updated
- [ ] Search index updated

## Automation and Tooling

### Documentation Generation

```bash
# Generate API docs from code
make docs-generate-api

# Update tool catalog from implementations
make docs-update-tools

# Create changelog from commits
make docs-changelog
```

### CI/CD Integration

```yaml
# .github/workflows/docs.yml
name: Documentation
on: [push, pull_request]

jobs:
  validate:
    steps:
      - name: Check documentation
        run: make docs-check
      
      - name: Test examples
        run: make docs-test-examples
      
      - name: Build documentation
        run: make docs-build
```

### Documentation Tools

- **Markdown Linters**: markdownlint, remark
- **Link Checkers**: markdown-link-check
- **Spell Checkers**: cspell, aspell
- **Diagram Tools**: mermaid, plantuml
- **API Doc Generators**: godoc, swag

## Continuous Improvement

### Feedback Channels

1. **Documentation Issues**
   - Label: `documentation`
   - Template: `.github/ISSUE_TEMPLATE/doc-issue.md`

2. **Documentation Metrics**
   - Monthly usage analytics
   - Search query analysis
   - 404 error tracking

3. **Regular Reviews**
   - Quarterly doc audit
   - Annual structure review
   - Per-release accuracy check

### Evolution Process

1. **Propose Changes**
   - RFC for major changes
   - Issue for minor updates

2. **Review and Approve**
   - Technical accuracy review
   - Style guide compliance
   - Impact assessment

3. **Implement and Monitor**
   - Gradual rollout
   - Monitor feedback
   - Iterate based on usage

---

This blueprint is a living document. Updates require approval from the technical lead and should be announced to all contributors.
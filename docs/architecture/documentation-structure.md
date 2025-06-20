# Documentation Organization Structure

This document explains the organization of documentation in the New Relic MCP Server project.

## Directory Structure

```
docs/
├── README.md                    # Main documentation index
├── QUICKSTART.md               # Quick start guide
├── DISCOVERY_ECOSYSTEM_OVERVIEW.md # Complete ecosystem overview
├── TECHNICAL_PLATFORM_SPEC.md  # Unified technical blueprint
├── ORGANIZATION_STRUCTURE.md    # This file
├── TECHNICAL_SPEC.md           # Detailed technical specifications
├── ARCHITECTURE_COMPLETE.md     # Comprehensive architecture
├── MIGRATION_GUIDE.md          # Migration guide
├── DEVELOPMENT.md              # Development setup
├── FUNCTIONAL_WORKFLOWS_ANALYSIS.md # Functional analysis
├── WORKFLOW_PATTERNS_GUIDE.md   # Workflow patterns
├── api/                        # API documentation
│   └── reference.md           # Complete API reference for all 120+ tools
├── architecture/              # Architecture documentation
│   ├── overview.md           # System design overview
│   ├── discovery-first.md    # Discovery-first technical architecture
│   ├── data-observability.md # Data observability toolkit
│   ├── state-management.md   # State and caching architecture
│   └── cross-account.md      # Multi-account support
├── examples/                  # Code and workflow examples
│   ├── DISCOVERY_DRIVEN_INVESTIGATION_EXAMPLES.md
│   ├── DISCOVERY_FIRST_CODE_EXAMPLE.md
│   ├── DISCOVERY_FIRST_WORKFLOWS.md
│   ├── functional-workflows.md # Breaking down monolithic approaches
│   └── workflow-patterns.md   # Composing tools into workflows
├── guides/                    # User and developer guides
│   ├── deployment.md         # Production deployment guide
│   ├── development.md        # Developer setup and workflow
│   ├── integration.md        # Integration with other tools
│   ├── llm-integration.md    # AI assistant integration
│   ├── migration.md          # Migration from assumptions to discovery
│   ├── testing.md            # Testing guide
│   ├── refactoring.md        # Code refactoring guide
│   └── troubleshooting.md    # Common issues and solutions
├── philosophy/               # Core philosophy documents
│   ├── DISCOVERY_PHILOSOPHY.md    # Deep philosophical foundations
│   ├── NO_ASSUMPTIONS_MANIFESTO.md # Zero assumptions commitment
│   └── ZERO_ASSUMPTIONS_EXAMPLES.md # Practical examples
├── ux/                       # User experience documentation
│   ├── DISCOVERY_MAGIC_MOMENTS.md # Specific delight scenarios
│   ├── INTERACTIVE_DISCOVERY_EXPERIENCES.md # Learning by doing
│   ├── VISUAL_DISCOVERY_DESIGN.md # Visual design system
│   ├── WOW_EXPERIENCE_DESIGN.md  # 5-minute magic formula
│   └── WOW_EXPERIENCE_SUMMARY.md # Complete UX strategy
└── archive/                       # Archived/deprecated documents
    ├── CONSOLIDATION_COMPLETE.md
    └── DISCOVERY_FIRST_ARCHITECTURE.md
```

## Key Documents by Purpose

### Getting Started
1. `../README.md` - Start here for navigation
2. `../QUICKSTART.md` - 5-minute setup guide
3. `../DISCOVERY_ECOSYSTEM_OVERVIEW.md` - Understanding the system

### Architecture & Design
1. `architecture/overview.md` - System components
2. `architecture/discovery-first.md` - Core approach
3. `complete-reference.md` - Comprehensive details

### API & Tools
1. `api/reference.md` - All 120+ tools documented
2. `../guides/refactoring.md` - Adding new tools
3. `../guides/tool-granularity.md` - Tool design principles

### Philosophy & Principles
1. `philosophy/NO_ASSUMPTIONS_MANIFESTO.md` - Core commitment
2. `discovery-first.md` - Discovery-first approach
3. `philosophy/ZERO_ASSUMPTIONS_EXAMPLES.md` - Real examples

### Examples & Patterns
1. `examples/DISCOVERY_FIRST_WORKFLOWS.md` - Common patterns
2. `examples/DISCOVERY_DRIVEN_INVESTIGATION_EXAMPLES.md` - Real scenarios
3. `../examples/workflow-patterns.md` - Implementation patterns

### User Experience
1. `ux/WOW_EXPERIENCE_DESIGN.md` - Creating magic
2. `ux/VISUAL_DISCOVERY_DESIGN.md` - Visual system
3. `../ux/INTERACTIVE_DISCOVERY_EXPERIENCES.md` - Interactive tutorials

## Navigation Tips

- **New Users**: Start with `../QUICKSTART.md` → `ecosystem-overview.md` → `../api/reference.md`
- **Developers**: Start with `../guides/development.md` → `../guides/refactoring.md` → `../guides/testing.md`
- **Architects**: Start with `architecture/overview.md` → `philosophy/NO_ASSUMPTIONS_MANIFESTO.md`
- **UX Designers**: Start with `ux/WOW_EXPERIENCE_DESIGN.md` → `ux/VISUAL_DISCOVERY_DESIGN.md`

## Maintenance

- Keep documents focused on their specific purpose
- Avoid duplicating content across files
- Link between documents rather than repeating information
- Archive deprecated content to the `archive/` directory
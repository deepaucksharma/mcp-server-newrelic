# Documentation Index

Welcome to the New Relic MCP Server documentation. This index provides quick access to all documentation organized by topic.

⚠️ **Important**: While documentation describes 120+ tools and advanced features, the current implementation includes only ~10-15 basic tools. See [Implementation Gaps Analysis](./IMPLEMENTATION_GAPS_ANALYSIS.md) for a detailed assessment of what's actually available vs. documented.

## ⚠️ Implementation Status

- **[Implementation Gaps Analysis](./IMPLEMENTATION_GAPS_ANALYSIS.md)** - What's actually implemented vs documented
- **[Current Capabilities](./CURRENT_CAPABILITIES.md)** - What works today
- **[Roadmap to Completion](../ROADMAP_2025.md)** - Timeline for missing features

## 🚀 Getting Started

- **[Quick Start Guide](./QUICKSTART.md)** - Get up and running in 5 minutes
- **[Development Setup](./guides/development.md)** - Set up your development environment
- **[Configuration Guide](../README.md#configuration)** - Configure the server
- **[Discovery Ecosystem Overview](./architecture/ecosystem-overview.md)** - Understanding the complete system

## 🏗️ Architecture

- **[Architecture Overview](./architecture/overview.md)** - System design, components, and patterns
- **[Discovery-First Architecture](./architecture/discovery-first.md)** - Technical implementation of discovery-first
- **[Complete Architecture Reference](./architecture/complete-reference.md)** - Comprehensive system architecture
- **[State Management](./architecture/state-management.md)** - Session and caching architecture
- **[Data Observability Toolkit](./architecture/data-observability.md)** - Tool ecosystem
- **[Cross-Account Support](./architecture/cross-account.md)** - Multi-account features

## 📖 API Reference

- **[Complete API Reference](./api/reference.md)** - All 120+ tools documented
- **[Tool Categories](./api/reference.md#tool-categories)** - Understanding tool organization
- **[Error Handling](./api/reference.md#error-handling)** - Error codes and recovery

## 📚 Guides

### User Guides
- **[Deployment Guide](./guides/deployment.md)** - Production deployment instructions
- **[Integration Guide](./guides/integration.md)** - Integrating with other tools
- **[LLM Integration Guide](./guides/llm-integration.md)** - Integrating with AI assistants
- **[Migration Guide](./guides/migration.md)** - Moving from assumption-based to discovery-first

### Development Guides
- **[Contributing Guide](./guides/development.md#contributing)** - How to contribute
- **[Testing Guide](./guides/testing.md)** - Writing and running tests
- **[Comprehensive Testing Strategy](./guides/comprehensive-testing-strategy.md)** - Multi-layered test plan
- **[Refactoring Guide](./guides/refactoring.md)** - Modernizing the codebase
- **[Mock Mode Guide](./guides/mock-mode.md)** - Development without New Relic connection
- **[Error Handling Guide](./guides/error-handling.md)** - Comprehensive error handling

## 🎯 Examples & Tutorials

- **[Discovery-Driven Investigations](./examples/DISCOVERY_DRIVEN_INVESTIGATION_EXAMPLES.md)** - Real-world scenarios
- **[Workflow Examples](./examples/DISCOVERY_FIRST_WORKFLOWS.md)** - Common workflow patterns
- **[Workflow Patterns](./examples/workflow-patterns.md)** - Composing tools into workflows
- **[Code Examples](./examples/DISCOVERY_FIRST_CODE_EXAMPLE.md)** - Implementation examples
- **[Functional Workflows Analysis](./examples/functional-workflows.md)** - Breaking down monolithic approaches

## 🔧 Technical Specifications

- **[Technical Platform Spec](./technical/platform-spec.md)** - Future architecture vision and blueprint
- **[Technical Specification](./technical/specification.md)** - Current implementation reference
- **[Specification Reconciliation](./technical/SPECIFICATION_RECONCILIATION.md)** - How the specs relate
- **[Documentation Structure](./architecture/documentation-structure.md)** - Documentation organization

## 📊 Project Information

- **[Roadmap](../ROADMAP_2025.md)** - Future development plans
- **[Architecture Decisions](./architecture/complete-reference.md#architectural-decisions)** - Key design choices

## 🔍 Philosophy & Principles

- **[No Assumptions Manifesto](./philosophy/NO_ASSUMPTIONS_MANIFESTO.md)** - Our commitment to discovery
- **[Zero Assumptions Examples](./philosophy/ZERO_ASSUMPTIONS_EXAMPLES.md)** - Practical applications

## 🤖 AI Assistant Guidelines

### Golden Rules for AI Assistants
1. **Always Start with Discovery** - Never assume data structures exist
2. **Validate Before Execute** - Check queries before running them
3. **Explain Your Process** - Share what you're discovering and why
4. **Build Progressively** - Start simple, refine based on findings
5. **Maintain Context** - Remember discoveries within the session

For detailed guidance, see the [LLM Integration Guide](./guides/llm-integration.md)

## 🌟 Additional Resources

- **[Troubleshooting Guide](./guides/troubleshooting.md)** - Common issues and solutions
- **[Security Guidelines](../README.md#security)** - Security best practices
- **[Support](../README.md#support)** - Getting help

## 🎨 User Experience

- **[Wow Experience Design](./ux/WOW_EXPERIENCE_DESIGN.md)** - Creating magical experiences
- **[Visual Discovery Design](./ux/VISUAL_DISCOVERY_DESIGN.md)** - Beautiful discovery visualizations
- **[Interactive Discovery Experiences](./ux/INTERACTIVE_DISCOVERY_EXPERIENCES.md)** - Learn by doing
- **[Discovery Magic Moments](./ux/DISCOVERY_MAGIC_MOMENTS.md)** - Specific delight scenarios
- **[Wow Experience Summary](./ux/WOW_EXPERIENCE_SUMMARY.md)** - Complete UX strategy

---

### Quick Navigation

**Most Popular:**
1. [Quick Start](../README.md#-quick-start)
2. [API Reference](./api/reference.md)
3. [Architecture Overview](./architecture/overview.md)
4. [Deployment Guide](./guides/deployment.md)

**For Developers:**
1. [Development Setup](./guides/development.md)
2. [Contributing Guide](./guides/development.md#contributing)
3. [Testing Guide](./guides/testing.md)

**For Users:**
1. [Workflow Examples](./examples/DISCOVERY_FIRST_WORKFLOWS.md)
2. [Migration Guide](./guides/migration.md)
3. [Troubleshooting](./guides/troubleshooting.md)
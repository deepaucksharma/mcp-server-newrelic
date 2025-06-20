# Technical Specification Reconciliation

This document clarifies the relationship between our technical specifications and resolves apparent conflicts.

## Overview

We have two complementary technical specifications that serve different purposes:

1. **`platform-spec.md`** - Architectural Blueprint (Vision)
2. **`specification.md`** - Implementation Reference (Current State)

## Specification Roles

### platform-spec.md (Technical Platform Specification)
- **Purpose**: Defines the target architecture and vision
- **Scope**: Complete discovery-first platform blueprint
- **Status**: Roadmap for future development
- **Key Features**:
  - Zero-assumption architecture
  - Mandatory discovery before operations
  - Adaptive Query Builder
  - Workflow Orchestrator
  - Self-observability

### specification.md (Technical Specification)
- **Purpose**: Documents the current implementation
- **Scope**: Existing MCP server capabilities
- **Status**: Current production state
- **Key Features**:
  - Standard MCP protocol implementation
  - 120+ granular tools
  - Optional discovery tools
  - Direct tool execution

## Reconciliation Strategy

### 1. Tool Naming Convention
- **Current** (specification.md): Underscore notation (`query_nrdb`)
- **Target** (platform-spec.md): Dot notation (`nrql.query`)
- **Resolution**: Maintain underscore notation for backward compatibility, document dot notation as future direction

### 2. Discovery Philosophy
- **Current**: Discovery tools are available but optional
- **Target**: Discovery is mandatory and foundational
- **Resolution**: Progressive migration:
  1. Current: Discovery tools available
  2. Next: Discovery strongly recommended (warnings when skipped)
  3. Future: Discovery mandatory (with opt-out flag for compatibility)

### 3. Architecture Layers
- **Current**: Traditional MCP server with tool registry
- **Target**: Layered architecture with Discovery Engine at core
- **Resolution**: Incremental refactoring following platform-spec.md roadmap

### 4. Error Handling
- **Current**: Standard MCP error codes
- **Target**: Custom DiscoveryError (-40001) for discovery failures
- **Resolution**: Extend error handling to include discovery-specific errors

## Implementation Roadmap

### Phase 1: Foundation (Current)
- ✅ 120+ granular tools implemented
- ✅ Discovery tools available
- ✅ Basic workflow support

### Phase 2: Discovery Enhancement (Q1 2025)
- Add discovery recommendations to all tools
- Implement discovery caching layer
- Add discovery confidence metrics

### Phase 3: Adaptive Intelligence (Q2 2025)
- Implement Adaptive Query Builder
- Add discovery-based query optimization
- Enhance workflow orchestration

### Phase 4: Full Discovery-First (Q3 2025)
- Make discovery mandatory (with compatibility flag)
- Implement all canonical discovery chains
- Complete self-observability features

## Usage Guidelines

### For Developers
- Follow `specification.md` for current implementation
- Align new features with `platform-spec.md` vision
- Use discovery tools whenever possible

### For Documentation
- Reference both specs with clear context:
  - "Current capabilities" → specification.md
  - "Architecture vision" → platform-spec.md
- Avoid conflicting statements

### For Users
- Current tools work as documented in specification.md
- Discovery-first approach recommended for best results
- Future updates will enhance discovery capabilities

## Conclusion

Both specifications are valid and serve important purposes:
- **specification.md** = What we have today
- **platform-spec.md** = Where we're going

This reconciliation ensures documentation consistency while maintaining a clear path from current implementation to future vision.
# MCP Server New Relic: Enhanced Platform Intelligence

> 🚀 **Advanced MCP Server with Golden Signals Intelligence & OpenTelemetry Awareness**  
> Built with **official `@modelcontextprotocol/sdk` + TypeScript 5.3+**  
> **Zero Hardcoded Schemas** + **Intelligent Caching** + **Anomaly Detection**  
> **Composite Tools** for advanced observability workflows

## Executive Summary

This Enhanced MCP Server provides sophisticated New Relic platform intelligence through composite tools, analytical metadata, and intelligent caching. The server implements a "discover-first, assume-nothing" approach with advanced capabilities for golden signals monitoring, entity comparison, and adaptive dashboard generation.

## 🎯 Key Features

### **Composite Intelligence Tools**
- **`discover.environment`** - One-call comprehensive environment discovery with OpenTelemetry awareness
- **`generate.golden_dashboard`** - Intelligent dashboard generation with automatic query adaptation
- **`compare.similar_entities`** - Performance benchmarking and outlier detection across entities

### **Golden Signals Intelligence**
- **Latency, Traffic, Errors, Saturation** monitoring with automatic instrumentation detection
- **Anomaly Detection** using statistical analysis and baseline establishment
- **Seasonality Detection** with confidence scoring
- **Data Quality Assessment** with completeness and consistency metrics

### **Intelligent Caching System**
- **Adaptive TTL** strategies based on data type and access patterns
- **Background Refresh** for critical data
- **Freshness Indicators** (fresh/recent/stale/expired)
- **Cache Health Monitoring** with optimization recommendations

## 🏗️ Architecture Highlights

- **OpenTelemetry Awareness**: Automatic detection and optimization for OTEL vs APM instrumentation
- **Platform-Native Discovery**: Zero assumptions about data schemas or attribute names
- **Analytical Metadata**: Rich insights with trend analysis and pattern detection
- **Performance Optimization**: Intelligent caching with context-aware freshness strategies

## 📚 Documentation Index

### Core Documentation (00-09)
- **[00_README.md](00_README.md)** - This overview document
- **[01_GETTING_STARTED.md](01_GETTING_STARTED.md)** - Quick start guide and first steps
- **[02_INSTALLATION.md](02_INSTALLATION.md)** - Detailed installation instructions
- **[03_CONFIGURATION.md](03_CONFIGURATION.md)** - Complete configuration reference

### Architecture & Design (10-19)
- **[10_ARCHITECTURE_OVERVIEW.md](10_ARCHITECTURE_OVERVIEW.md)** - High-level architecture and components
- **[11_ARCHITECTURE_DISCOVERY_FIRST.md](11_ARCHITECTURE_DISCOVERY_FIRST.md)** - Discovery-first design philosophy
- **[12_ARCHITECTURE_INTELLIGENT_CACHING.md](12_ARCHITECTURE_INTELLIGENT_CACHING.md)** - Caching strategies and freshness policies

### Tools Documentation (30-39)
- **[30_TOOLS_OVERVIEW.md](30_TOOLS_OVERVIEW.md)** - Complete tools catalog
- **[31_TOOLS_COMPOSITE.md](31_TOOLS_COMPOSITE.md)** - Composite tools reference
- **[32_TOOLS_ENHANCED.md](32_TOOLS_ENHANCED.md)** - Enhanced existing tools
- **[33_TOOLS_ANALYTICS.md](33_TOOLS_ANALYTICS.md)** - Analytical and caching tools

### User Guides (40-49)
- **[40_GUIDE_QUICKSTART.md](40_GUIDE_QUICKSTART.md)** - 5-minute quickstart tutorial
- **[41_GUIDE_GOLDEN_SIGNALS.md](41_GUIDE_GOLDEN_SIGNALS.md)** - Golden signals monitoring guide
- **[42_GUIDE_DISCOVERY_WORKFLOWS.md](42_GUIDE_DISCOVERY_WORKFLOWS.md)** - Discovery-first workflow patterns

### Examples & Workflows (50-59)
- **[50_EXAMPLES_OVERVIEW.md](50_EXAMPLES_OVERVIEW.md)** - Example scenarios index
- **[51_EXAMPLES_COMPOSITE_TOOLS.md](51_EXAMPLES_COMPOSITE_TOOLS.md)** - Composite tool usage examples

### Testing & Quality (60-69)
- **[60_TESTING_STRATEGY.md](60_TESTING_STRATEGY.md)** - Overall testing approach
- **[63_TESTING_E2E.md](63_TESTING_E2E.md)** - End-to-end testing with NRDB

## 🚀 Quick Start

### 1. Installation
```bash
git clone <repository>
cd mcp-server-newrelic
npm install
```

### 2. Configuration
```bash
export NEW_RELIC_API_KEY="NRAK-..."
export NEW_RELIC_ACCOUNT_ID="12345"
export NEW_RELIC_REGION="US"  # or "EU"
```

### 3. Usage Examples

**Environment Discovery:**
```typescript
// Get complete environment overview
const env = await mcp.call('discover.environment', {
  includeHealth: true,
  maxEntities: 50
});
```

**Golden Dashboard Generation:**
```typescript
// Generate adaptive golden signals dashboard
const dashboard = await mcp.call('generate.golden_dashboard', {
  entity_guid: 'MXxBUE18QVBQTElDQVRJT058MTIzNDU2',
  timeframe_hours: 1,
  create_dashboard: false  // preview first
});
```

**Entity Performance Comparison:**
```typescript
// Compare similar entities for optimization opportunities
const comparison = await mcp.call('compare.similar_entities', {
  comparison_strategy: 'by_type',
  entity_type: 'APPLICATION',
  max_entities: 10
});
```

## 🎨 Key Design Principles

1. **Discover-First, Assume-Nothing**: All schemas and patterns discovered at runtime
2. **Composite Intelligence**: High-level tools that combine multiple operations
3. **OpenTelemetry Awareness**: Automatic adaptation for modern instrumentation
4. **Analytical Depth**: Rich metadata with anomaly detection and trends
5. **Performance Optimization**: Intelligent caching with adaptive freshness
6. **AI-Optimized**: Tools designed for LLM agent workflows

## 🛠️ Technology Stack

```yaml
Runtime: Node.js 20+ / Bun 1.0+
Language: TypeScript 5.3+
MCP SDK: @modelcontextprotocol/sdk 1.0+
GraphQL: NerdGraph API integration
Caching: Intelligent memory cache with adaptive TTL
Analytics: Statistical analysis and anomaly detection
Testing: Vitest + comprehensive E2E testing
```

## 📊 Enhanced Capabilities

### **Golden Signals Intelligence Engine**
- Automatic instrumentation detection (OpenTelemetry vs APM)
- Statistical analysis with baseline establishment
- Anomaly detection with confidence scoring
- Trend analysis and seasonality detection

### **Intelligent Caching System**
- Context-aware TTL strategies
- Background refresh for critical data
- Cache health monitoring and optimization
- Memory usage optimization with LRU eviction

### **Composite Tool Architecture**
- High-level workflows combining multiple operations
- Intelligent error handling and fallback strategies
- Rich metadata for LLM agent decision-making
- Performance optimization with parallel execution

## 🔗 Related Resources

- **New Relic Documentation**: [docs.newrelic.com](https://docs.newrelic.com)
- **MCP Protocol**: [modelcontextprotocol.io](https://modelcontextprotocol.io)
- **OpenTelemetry**: [opentelemetry.io](https://opentelemetry.io)

---

**License**: MIT  
**Code of Conduct**: We welcome all contributors following our discovery-first philosophy.
# Adaptive Query Builder (AQB): Technical Specification

The Adaptive Query Builder is a core component that translates high-level intent into optimized NRQL queries that adapt to discovered schemas. This document provides comprehensive technical details about its design and implementation.

## Overview

The Adaptive Query Builder (AQB) represents a paradigm shift from static query construction to dynamic, schema-aware query generation. It uses the Discovery Engine's World Model to build queries that automatically adapt to schema changes, optimize for performance, and provide explainable results.

## Core Philosophy

### Traditional Approach (Static)
```
Problem: "Show me errors for checkout service"
Traditional: SELECT count(*) FROM Transaction WHERE appName = 'checkout' AND error = true

Issues:
- Assumes 'appName' field exists
- Assumes 'error' is boolean
- Fails if schema changes
```

### AQB Approach (Adaptive)
```
Problem: "Show me errors for checkout service"
AQB Process:
1. Discover service identifier field
2. Discover error indicator field
3. Build query with discovered fields
4. Optimize based on data patterns
5. Provide confidence and alternatives

Result: Query that works regardless of schema
```

## Architecture

### Component Structure

```
AdaptiveQueryBuilder
├── IntentParser          // Understands query intent
├── SchemaAdapter         // Maps intent to schema
├── QueryOptimizer        // Optimizes performance
├── ConfidenceScorer      // Calculates result confidence
└── ExplainabilityEngine  // Explains decisions
```

### Query Building Pipeline

```
1. Intent Analysis
   └── Parse user intent into structured format

2. Schema Mapping
   └── Map intent to discovered schema elements

3. Query Construction
   └── Build NRQL with discovered fields

4. Optimization
   └── Apply performance optimizations

5. Execution & Explanation
   └── Run query with explainability
```

## Intent Understanding

### Intent Categories

1. **Metric Queries**
   - Error rates
   - Response times
   - Throughput
   - Custom metrics

2. **Entity Queries**
   - Service health
   - Host metrics
   - Container stats
   - Application data

3. **Analysis Queries**
   - Comparisons
   - Trends
   - Anomalies
   - Correlations

### Intent Parser

```
IntentStructure {
  action: "count" | "average" | "sum" | "percentile" | "rate"
  target: "errors" | "response_time" | "throughput" | "custom"
  scope: {
    type: "service" | "host" | "container" | "global"
    identifier: string
  }
  timeRange: TimeRange
  filters: Filter[]
  grouping: string[]
}
```

### Intent Examples

```
"Show error rate for checkout service"
→ {
    action: "rate",
    target: "errors",
    scope: { type: "service", identifier: "checkout" },
    timeRange: "1 hour"
  }

"Compare response times between services"
→ {
    action: "average",
    target: "response_time",
    scope: { type: "service", identifier: "*" },
    grouping: ["service"]
  }
```

## Schema Adaptation

### Field Resolution Strategy

```
1. Direct Mapping
   - Exact field name match
   - Confidence: 1.0

2. Semantic Mapping
   - Similar field names (edit distance)
   - Confidence: 0.7-0.9

3. Pattern Matching
   - Field behavior analysis
   - Confidence: 0.5-0.7

4. Heuristic Fallback
   - Best guess based on type
   - Confidence: 0.3-0.5
```

### Service Identifier Resolution

```
Resolution Chain:
1. Check world model for identified service field
2. Try common patterns: appName, service.name, app.name
3. Analyze field cardinality and distribution
4. Use correlation with known service patterns
5. Fall back to user specification
```

### Error Indicator Resolution

```
Error Detection Strategy:
1. Boolean fields with error-related names
2. Numeric fields with error codes
3. String fields with error values
4. Fields correlating with incidents
5. Custom error definitions
```

## Query Construction

### Query Template System

```
Template Structure:
{
  baseQuery: "SELECT {aggregation} FROM {eventType}",
  whereClause: "WHERE {conditions}",
  timeClause: "SINCE {timeRange}",
  groupByClause: "FACET {grouping}",
  orderClause: "ORDER BY {ordering}",
  limitClause: "LIMIT {limit}"
}
```

### Dynamic Query Building

```
Example: Error Rate Query

Input Intent: "error rate for checkout service"

Step 1: Discover Fields
- Service field: "appName" (confidence: 0.95)
- Error field: "error" (boolean, confidence: 0.90)

Step 2: Build Query Components
- Aggregation: "percentage(count(*), WHERE error = true)"
- Event Type: "Transaction"
- Condition: "appName = 'checkout'"
- Time: "SINCE 1 hour ago"

Step 3: Construct Query
SELECT percentage(count(*), WHERE error = true) as 'Error Rate'
FROM Transaction
WHERE appName = 'checkout'
SINCE 1 hour ago
```

### Query Variations

AQB generates multiple query variations:

```
Primary Query (Confidence: 0.92):
SELECT percentage(count(*), WHERE error = true)
FROM Transaction
WHERE appName = 'checkout'
SINCE 1 hour ago

Alternative 1 (Confidence: 0.78):
SELECT percentage(count(*), WHERE errorCode != 0)
FROM APMEvent
WHERE service.name = 'checkout'
SINCE 1 hour ago

Alternative 2 (Confidence: 0.65):
SELECT filter(count(*), WHERE status >= 400) / count(*) * 100
FROM Request
WHERE app = 'checkout'
SINCE 1 hour ago
```

## Query Optimization

### Performance Optimization Strategies

1. **Index Awareness**
   - Use indexed fields in WHERE clauses
   - Order conditions by selectivity
   - Avoid function calls on indexed fields

2. **Time Range Optimization**
   - Adjust granularity based on time range
   - Use appropriate time buckets
   - Implement progressive loading

3. **Aggregation Optimization**
   - Pre-aggregate when possible
   - Use approximate functions for large datasets
   - Implement sampling for estimates

### Cost-Based Optimization

```
Query Cost Estimation:
1. Estimate data volume from time range
2. Calculate cardinality of grouped fields
3. Assess complexity of aggregations
4. Predict query execution time
5. Choose optimal query plan
```

### Optimization Examples

```
Before Optimization:
SELECT count(*) FROM Transaction 
WHERE toLowerCase(appName) = 'checkout'
SINCE 7 days ago

After Optimization:
SELECT count(*) FROM Transaction 
WHERE appName = 'checkout' OR appName = 'Checkout' OR appName = 'CHECKOUT'
SINCE 7 days ago
LIMIT 1000
```

## Confidence Scoring

### Confidence Calculation Model

```
Overall Confidence = weighted_average([
  field_confidence * 0.4,      // How confident in field choice
  pattern_confidence * 0.3,    // How well pattern matches
  historical_success * 0.2,    // Past query success rate
  data_coverage * 0.1         // Amount of data analyzed
])
```

### Confidence Factors

1. **Field Confidence**
   - Direct match: 1.0
   - Semantic match: 0.8
   - Pattern match: 0.6
   - Heuristic: 0.4

2. **Pattern Confidence**
   - Known pattern: 0.9
   - Similar pattern: 0.7
   - New pattern: 0.5

3. **Historical Success**
   - Track query success rates
   - Adjust based on user feedback
   - Learn from modifications

## Explainability

### Query Explanation Structure

```
QueryExplanation {
  queryId: string
  intent: ParsedIntent
  decisions: Decision[]
  alternatives: Alternative[]
  confidence: ConfidenceBreakdown
  recommendations: string[]
}

Decision {
  step: string
  choice: string
  reasoning: string
  confidence: number
  alternatives: string[]
}
```

### Explanation Example

```
Query: "Show me the error rate for checkout service"

Explanation:
1. Intent Recognition
   - Identified: Calculate error percentage
   - Confidence: 0.95

2. Service Field Selection
   - Chose: 'appName'
   - Reasoning: Field present in 98% of records, 12 unique values
   - Confidence: 0.92
   - Alternative: 'service.name' (0.76 confidence)

3. Error Field Selection
   - Chose: 'error' (boolean)
   - Reasoning: Boolean field, correlates with incidents
   - Confidence: 0.88
   - Alternative: 'errorCode != 0' (0.72 confidence)

4. Query Construction
   - Pattern: Percentage calculation with boolean filter
   - Optimization: Added index-friendly WHERE clause
   - Confidence: 0.90

Overall Confidence: 0.89 (High)
```

## Advanced Features

### Multi-Step Query Building

For complex analyses, AQB can build multi-step queries:

```
Intent: "Compare error rates between peak and off-peak hours"

Step 1: Discover peak hours
SELECT count(*) FROM Transaction 
FACET hourOf(timestamp) 
SINCE 1 week ago

Step 2: Build comparison query
SELECT 
  percentage(count(*), WHERE error = true AND hourOf(timestamp) IN (9,10,11,17,18,19)) as 'Peak Error Rate',
  percentage(count(*), WHERE error = true AND hourOf(timestamp) NOT IN (9,10,11,17,18,19)) as 'Off-Peak Error Rate'
FROM Transaction
WHERE appName = 'checkout'
SINCE 1 week ago
```

### Adaptive Aggregations

AQB selects appropriate aggregations based on data types:

```
Numeric Fields:
- Small range: average(), sum()
- Large range: percentile(), stddev()
- Rates: rate(), percentage()

Time-based:
- Short periods: count()
- Long periods: rate(count())
- Trends: derivative()

Categorical:
- Low cardinality: FACET field
- High cardinality: FACET capture(field, r'pattern')
```

### Query Learning

AQB learns from query execution:

```
Learning Process:
1. Execute query with confidence tracking
2. Analyze result quality signals:
   - Empty results (potential field mismatch)
   - Unexpected distributions
   - User modifications
3. Update confidence models
4. Adjust future query generation
```

## Integration APIs

### TypeScript Interface

```typescript
interface AdaptiveQueryBuilder {
  // Build query from intent
  build(intent: QueryIntent, worldModel: WorldModel): Promise<AdaptiveQuery>;
  
  // Validate query syntax and semantics
  validate(query: string, worldModel: WorldModel): ValidationResult;
  
  // Explain query construction
  explain(query: AdaptiveQuery): QueryExplanation;
  
  // Optimize existing query
  optimize(query: string, worldModel: WorldModel): OptimizedQuery;
  
  // Learn from query results
  learn(query: AdaptiveQuery, results: QueryResults): void;
}

interface AdaptiveQuery {
  primary: {
    nrql: string;
    confidence: number;
  };
  alternatives: Array<{
    nrql: string;
    confidence: number;
    reasoning: string;
  }>;
  explanation: QueryExplanation;
  metadata: QueryMetadata;
}
```

### Usage Example

```typescript
const aqb = new AdaptiveQueryBuilder();

// Build adaptive query
const query = await aqb.build({
  intent: "error_rate",
  scope: { type: "service", value: "checkout" },
  timeRange: "1 hour"
}, worldModel);

// Execute with explainability
const results = await nrClient.query(query.primary.nrql);

// Learn from execution
aqb.learn(query, results);

// Get explanation
const explanation = aqb.explain(query);
console.log(`Query confidence: ${query.primary.confidence}`);
console.log(`Decisions made: ${explanation.decisions.length}`);
```

## Error Handling

### Common Failure Scenarios

1. **No Suitable Fields Found**
   ```
   Error: Cannot find service identifier field
   Fallback: Request user specification
   Recovery: Update world model with user input
   ```

2. **Ambiguous Intent**
   ```
   Error: Multiple interpretations possible
   Fallback: Present options to user
   Recovery: Learn from user selection
   ```

3. **Schema Mismatch**
   ```
   Error: Expected field not present
   Fallback: Re-discover schema
   Recovery: Rebuild query with new schema
   ```

### Graceful Degradation

```
Degradation Strategy:
1. Try primary query approach
2. Fall back to alternative queries
3. Simplify query if too complex
4. Request user clarification
5. Provide manual query option
```

## Performance Characteristics

### Query Building Performance

- Intent parsing: <10ms
- Schema adaptation: <50ms
- Query construction: <20ms
- Optimization: <100ms
- Total: <200ms for most queries

### Optimization Impact

Query optimization typically provides:
- 30-50% reduction in execution time
- 40-60% reduction in data scanned
- 20-30% improvement in result accuracy

## Best Practices

### For Developers

1. **Always Provide World Model**
   - Never build queries without discovery
   - Keep world model updated
   - Handle stale world models

2. **Use Confidence Scores**
   - Present alternatives for low confidence
   - Explain confidence to users
   - Track confidence over time

3. **Enable Learning**
   - Capture query modifications
   - Track result quality
   - Feed back into AQB

### For Users

1. **Provide Clear Intent**
   - Be specific about what you want
   - Include relevant context
   - Specify time ranges

2. **Review Explanations**
   - Understand query decisions
   - Validate field choices
   - Suggest improvements

3. **Iterate on Results**
   - Refine queries based on results
   - Try alternative queries
   - Provide feedback

## Future Enhancements

### Natural Language Understanding

Enhanced intent parsing using NLP:
- Support for complex natural language queries
- Context-aware query building
- Conversational query refinement

### Machine Learning Integration

- Learn optimal query patterns
- Predict query performance
- Suggest query improvements
- Anomaly detection in query patterns

### Advanced Optimization

- Cost-based optimization with budget constraints
- Multi-query optimization for dashboards
- Predictive caching for common queries
- Query result streaming for large datasets

## Conclusion

The Adaptive Query Builder transforms NRQL query construction from a manual, error-prone process to an intelligent, adaptive system. By leveraging the Discovery Engine's understanding of your data, AQB ensures queries work reliably across different schemas, optimize for performance automatically, and provide clear explanations of their construction.

This approach embodies the zero-assumptions philosophy: rather than requiring users to know their schema in advance, AQB discovers what's available and builds the best possible query for the intent. The result is a more robust, user-friendly, and intelligent approach to observability data access.
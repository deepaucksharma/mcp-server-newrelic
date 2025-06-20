# Discovery-First Implementation Example

This document provides a concrete Go implementation example showing how to refactor a traditional tool to use the discovery-first approach.

## Example: Refactoring the Error Rate Tool

### Before: Assumption-Based Implementation

```go
// tools_query.go - OLD APPROACH
func (s *Server) handleGetErrorRate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Extract parameters
    appName, _ := params["app_name"].(string)
    timeRange, _ := params["time_range"].(string)
    
    // Hard-coded query assuming 'error' attribute exists and is boolean
    query := fmt.Sprintf(`
        SELECT percentage(count(*), WHERE error IS true) as 'Error Rate'
        FROM Transaction 
        WHERE appName = '%s'
        SINCE %s
    `, appName, timeRange)
    
    // Execute query - fails if assumptions are wrong
    result, err := s.nrClient.QueryNRDB(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    
    return result, nil
}
```

**Problems:**
- Assumes `error` attribute exists
- Assumes `error` is boolean
- Assumes `appName` exists
- No fallback if schema differs

### After: Discovery-First Implementation

```go
// tools_query_granular.go - NEW APPROACH
func (s *Server) handleGetErrorRate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Extract parameters
    appName, _ := params["app_name"].(string)
    timeRange, _ := params["time_range"].(string)
    
    // Step 1: Discover what event types and attributes exist
    discovery, err := s.discoverErrorIndicators(ctx, timeRange)
    if err != nil {
        return nil, fmt.Errorf("discovery failed: %w", err)
    }
    
    // Step 2: Build adaptive query based on discoveries
    query, err := s.buildAdaptiveErrorQuery(ctx, discovery, appName, timeRange)
    if err != nil {
        return nil, fmt.Errorf("query building failed: %w", err)
    }
    
    // Step 3: Validate query before execution
    validation, err := s.validateQueryAgainstSchema(ctx, query, discovery.Schema)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    if !validation.IsValid {
        return nil, fmt.Errorf("query validation failed: %s", validation.Reason)
    }
    
    // Step 4: Execute validated query
    result, err := s.nrClient.QueryNRDB(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query execution failed: %w", err)
    }
    
    // Step 5: Add metadata about how we calculated the rate
    return map[string]interface{}{
        "result":           result,
        "method_used":      discovery.MethodUsed,
        "confidence_score": discovery.ConfidenceScore,
        "data_quality":     discovery.DataQuality,
    }, nil
}

// Discovery function - explores what error indicators exist
func (s *Server) discoverErrorIndicators(ctx context.Context, timeRange string) (*ErrorDiscovery, error) {
    discovery := &ErrorDiscovery{
        EventTypes: make(map[string]EventTypeInfo),
    }
    
    // Check what event types exist
    eventTypesQuery := fmt.Sprintf("SHOW EVENT TYPES SINCE %s", timeRange)
    eventTypes, err := s.nrClient.QueryNRDB(ctx, eventTypesQuery)
    if err != nil {
        return nil, err
    }
    
    // For each relevant event type, explore attributes
    for _, eventType := range []string{"Transaction", "TransactionError", "Log"} {
        if !eventTypeExists(eventTypes, eventType) {
            continue
        }
        
        // Get sample to understand schema
        sampleQuery := fmt.Sprintf(`
            SELECT keyset() 
            FROM %s 
            LIMIT 100 
            SINCE %s
        `, eventType, timeRange)
        
        attributes, err := s.nrClient.QueryNRDB(ctx, sampleQuery)
        if err != nil {
            continue // Skip if can't sample
        }
        
        // Analyze what error indicators are available
        info := EventTypeInfo{
            Name:       eventType,
            Attributes: parseAttributes(attributes),
        }
        
        // Check for common error indicators
        if hasAttribute(info.Attributes, "error") {
            // Check if boolean or string
            typeQuery := fmt.Sprintf(`
                SELECT uniqueCount(error) as cardinality,
                       capture(toString(error), r'true|false') as is_boolean
                FROM %s 
                WHERE error IS NOT NULL
                LIMIT 1000 
                SINCE %s
            `, eventType, timeRange)
            
            typeInfo, _ := s.nrClient.QueryNRDB(ctx, typeQuery)
            info.ErrorAttribute = analyzeErrorAttribute(typeInfo)
        }
        
        // Check for other error indicators
        if hasAttribute(info.Attributes, "error.class") {
            info.HasErrorClass = true
        }
        if hasAttribute(info.Attributes, "httpResponseCode") {
            info.HasResponseCode = true
        }
        if hasAttribute(info.Attributes, "level") && eventType == "Log" {
            info.HasLogLevel = true
        }
        
        discovery.EventTypes[eventType] = info
    }
    
    // Determine best method for calculating error rate
    discovery.MethodUsed = determineBestMethod(discovery.EventTypes)
    discovery.ConfidenceScore = calculateConfidence(discovery.EventTypes)
    
    return discovery, nil
}

// Build query that adapts to discovered schema
func (s *Server) buildAdaptiveErrorQuery(ctx context.Context, discovery *ErrorDiscovery, appName, timeRange string) (string, error) {
    var query string
    
    switch discovery.MethodUsed {
    case "transaction_error_boolean":
        // Best case: error attribute exists and is boolean
        query = fmt.Sprintf(`
            SELECT percentage(count(*), WHERE error IS true) as 'Error Rate',
                   count(*) as 'Total Requests',
                   filter(count(*), WHERE error IS true) as 'Errors'
            FROM Transaction 
            %s
            SINCE %s
        `, buildWhereClause(discovery, appName), timeRange)
        
    case "transaction_error_class":
        // error.class exists
        query = fmt.Sprintf(`
            SELECT percentage(count(*), WHERE error.class IS NOT NULL) as 'Error Rate',
                   count(*) as 'Total Requests',
                   filter(count(*), WHERE error.class IS NOT NULL) as 'Errors',
                   uniqueCount(error.class) as 'Error Types'
            FROM Transaction 
            %s
            SINCE %s
        `, buildWhereClause(discovery, appName), timeRange)
        
    case "transaction_response_code":
        // Use HTTP response codes
        query = fmt.Sprintf(`
            SELECT percentage(count(*), WHERE httpResponseCode >= 400) as 'Error Rate',
                   count(*) as 'Total Requests',
                   filter(count(*), WHERE httpResponseCode >= 400) as 'Errors',
                   filter(count(*), WHERE httpResponseCode >= 500) as 'Server Errors',
                   filter(count(*), WHERE httpResponseCode >= 400 AND httpResponseCode < 500) as 'Client Errors'
            FROM Transaction 
            %s
            SINCE %s
        `, buildWhereClause(discovery, appName), timeRange)
        
    case "transaction_error_events":
        // Separate TransactionError events exist
        query = fmt.Sprintf(`
            SELECT 
                (SELECT count(*) FROM TransactionError %s SINCE %s) / 
                (SELECT count(*) FROM Transaction %s SINCE %s) * 100 as 'Error Rate'
        `, buildWhereClause(discovery, appName), timeRange,
           buildWhereClause(discovery, appName), timeRange)
        
    case "log_errors":
        // Fall back to log analysis
        query = fmt.Sprintf(`
            SELECT percentage(count(*), WHERE level IN ('ERROR', 'FATAL', 'CRITICAL')) as 'Error Rate',
                   count(*) as 'Total Logs',
                   filter(count(*), WHERE level IN ('ERROR', 'FATAL', 'CRITICAL')) as 'Error Logs'
            FROM Log 
            %s
            SINCE %s
        `, buildWhereClause(discovery, appName), timeRange)
        
    default:
        return "", fmt.Errorf("no suitable error indicators found in data")
    }
    
    return query, nil
}

// Helper to build WHERE clause that adapts to available attributes
func buildWhereClause(discovery *ErrorDiscovery, appName string) string {
    conditions := []string{}
    
    // Check if appName attribute exists
    if discovery.HasAppName {
        conditions = append(conditions, fmt.Sprintf("appName = '%s'", appName))
    } else if discovery.HasServiceName {
        conditions = append(conditions, fmt.Sprintf("service.name = '%s'", appName))
    } else if discovery.HasEntityGuid {
        // Look up entity.guid from appName
        guid := s.lookupEntityGuid(appName)
        if guid != "" {
            conditions = append(conditions, fmt.Sprintf("entity.guid = '%s'", guid))
        }
    }
    
    if len(conditions) > 0 {
        return "WHERE " + strings.Join(conditions, " AND ")
    }
    return ""
}

// Types for discovery results
type ErrorDiscovery struct {
    EventTypes      map[string]EventTypeInfo
    MethodUsed      string
    ConfidenceScore float64
    DataQuality     DataQualityInfo
    Schema          SchemaInfo
    HasAppName      bool
    HasServiceName  bool
    HasEntityGuid   bool
}

type EventTypeInfo struct {
    Name            string
    Attributes      []AttributeInfo
    ErrorAttribute  ErrorAttributeInfo
    HasErrorClass   bool
    HasResponseCode bool
    HasLogLevel     bool
    SampleCount     int64
}

type ErrorAttributeInfo struct {
    Exists       bool
    Type         string // "boolean", "string", "numeric"
    Coverage     float64
    NullPercent  float64
    IsBooleanish bool // true/false strings
}

// Integration with enhanced tool metadata
func (s *Server) registerDiscoveryAwareErrorRateTool() error {
    tool := NewToolBuilder("analysis.get_error_rate", "Calculate error rate using discovered error indicators").
        Category(CategoryAnalysis).
        Handler(s.handleGetErrorRate).
        Required("time_range").
        Param("app_name", EnhancedProperty{
            Property: Property{
                Type:        "string",
                Description: "Application name (will discover actual identifier)",
            },
        }).
        Param("time_range", EnhancedProperty{
            Property: Property{
                Type:        "string",
                Description: "Time range for analysis",
                Default:     "1 hour",
            },
        }).
        Param("discovery_cache_ttl", EnhancedProperty{
            Property: Property{
                Type:        "integer",
                Description: "Cache discovery results for N seconds",
                Default:     300,
            },
        }).
        Performance(func(p *PerformanceMetadata) {
            p.ExpectedLatencyMS = 1500 // Discovery adds overhead
            p.MaxLatencyMS = 5000
            p.Cacheable = true
            p.CacheTTLSeconds = 300
        }).
        AIGuidance(func(g *AIGuidanceMetadata) {
            g.Purpose = "Calculate error rate by discovering available error indicators"
            g.UsageExamples = []string{
                "Get error rate: analysis.get_error_rate(app_name='checkout-api', time_range='1 hour')",
                "Long-term analysis: analysis.get_error_rate(app_name='payment-service', time_range='7 days')",
            }
            g.ChainsWith = []string{
                "analysis.drill_down_errors",
                "alert.create_from_baseline",
                "dashboard.add_error_widget",
            }
            g.SuccessIndicators = []string{
                "Returns error rate with confidence score",
                "Explains which method was used",
                "Provides data quality assessment",
            }
        }).
        Example(ToolExample{
            Name:        "Adaptive error rate calculation",
            Description: "Discovers error indicators and calculates rate",
            Params: map[string]interface{}{
                "app_name":   "checkout-api",
                "time_range": "1 hour",
            },
            ExpectedResult: map[string]interface{}{
                "error_rate":      2.5,
                "total_requests":  150000,
                "errors":          3750,
                "method_used":     "transaction_error_boolean",
                "confidence":      0.95,
                "data_quality": map[string]interface{}{
                    "coverage":     0.98,
                    "null_percent": 0.02,
                },
            },
        }).
        Build()
    
    return s.tools.Register(tool.Tool)
}
```

## Key Principles Demonstrated

### 1. Discovery Before Execution
- Never assume attributes exist
- Explore schema first
- Check data quality

### 2. Adaptive Query Building
- Multiple strategies based on what's available
- Fallback options for different schemas
- Handle missing attributes gracefully

### 3. Transparency
- Report which method was used
- Include confidence scores
- Provide data quality metrics

### 4. Caching for Performance
- Cache discovery results
- Configurable TTL
- Invalidate on schema changes

### 5. Enhanced Error Handling
- Clear error messages
- Suggest alternatives
- Guide users to solutions

## Benefits of This Approach

1. **Reliability**: Works with any schema
2. **Intelligence**: Discovers best calculation method
3. **Transparency**: Explains how results were derived
4. **Performance**: Caches discoveries
5. **Maintainability**: Single code path handles all cases

## Next Steps

Apply this pattern to all tools:
- Query tools: Validate schema before execution
- Dashboard tools: Generate widgets based on available data
- Alert tools: Set thresholds based on discovered baselines
- Analysis tools: Adapt algorithms to data structure

Remember: **Always discover, never assume!**
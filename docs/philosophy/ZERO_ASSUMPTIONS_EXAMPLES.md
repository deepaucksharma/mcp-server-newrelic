# Zero Assumptions: Real Code Examples

This document provides concrete code examples showing how we implement zero-assumption patterns throughout the MCP server. Each example demonstrates the extreme lengths we go to avoid hard-coding anything.

## Table of Contents

1. [Service Identification Without Assumptions](#service-identification-without-assumptions)
2. [Error Detection Without Assumptions](#error-detection-without-assumptions)
3. [Metric Discovery Without Assumptions](#metric-discovery-without-assumptions)
4. [Dashboard Generation Without Assumptions](#dashboard-generation-without-assumptions)
5. [Performance Analysis Without Assumptions](#performance-analysis-without-assumptions)
6. [Cost Analysis Without Assumptions](#cost-analysis-without-assumptions)

## Service Identification Without Assumptions

Traditional systems hard-code `appName` as the service identifier. We discover what actually identifies services in each environment.

```go
// discovery/service_identifier.go

type ServiceIdentifier struct {
    Field      string
    Confidence float64
    Coverage   float64
    Examples   []string
}

// DiscoverServiceIdentifier finds how services are identified in this environment
func (d *DiscoveryEngine) DiscoverServiceIdentifier(ctx context.Context) (*ServiceIdentifier, error) {
    // Try multiple potential identifiers in order of likelihood
    candidates := []string{
        "appName",                    // Traditional New Relic
        "applicationName",            // Variant
        "service.name",              // OpenTelemetry standard
        "app",                       // Shortened variant
        "serviceName",               // Another variant
        "entity.name",               // Entity-based
        "cloud.service.name",        // Cloud environments
        "kubernetes.deployment.name", // K8s environments
        "container.name",            // Container environments
        "custom.service",            // Custom instrumentation
    }
    
    bestIdentifier := &ServiceIdentifier{
        Confidence: 0,
    }
    
    // Check each candidate
    for _, candidate := range candidates {
        // First, check if the field exists at all
        existenceQuery := fmt.Sprintf(`
            SELECT count(*) as total,
                   uniqueCount(%s) as unique_values,
                   filter(count(*), WHERE %s IS NOT NULL) as non_null
            FROM Transaction, SystemSample, Log
            SINCE 1 hour ago
        `, candidate, candidate)
        
        result, err := d.client.QueryNRDB(ctx, existenceQuery)
        if err != nil || result.Total == 0 {
            continue // Field doesn't exist or no data
        }
        
        // Calculate coverage (what percentage of events have this field)
        coverage := float64(result.NonNull) / float64(result.Total)
        
        // Skip if coverage is too low
        if coverage < 0.5 {
            continue
        }
        
        // Check if values look like service names
        sampleQuery := fmt.Sprintf(`
            SELECT uniques(%s, 10) as samples
            FROM Transaction, SystemSample, Log
            WHERE %s IS NOT NULL
            SINCE 1 hour ago
        `, candidate, candidate)
        
        samples, err := d.client.QueryNRDB(ctx, sampleQuery)
        if err != nil {
            continue
        }
        
        // Analyze samples to determine if they're service-like
        confidence := d.analyzeServiceNamePattern(samples.Samples)
        
        // Calculate overall score
        score := coverage * confidence
        
        if score > bestIdentifier.Confidence {
            bestIdentifier = &ServiceIdentifier{
                Field:      candidate,
                Confidence: score,
                Coverage:   coverage,
                Examples:   samples.Samples,
            }
        }
    }
    
    // If no good identifier found, try to discover custom patterns
    if bestIdentifier.Confidence < 0.3 {
        customIdentifier := d.discoverCustomServicePattern(ctx)
        if customIdentifier != nil && customIdentifier.Confidence > bestIdentifier.Confidence {
            bestIdentifier = customIdentifier
        }
    }
    
    // If still no identifier, look for natural groupings
    if bestIdentifier.Confidence < 0.3 {
        naturalGrouping := d.discoverNaturalServiceGrouping(ctx)
        if naturalGrouping != nil {
            bestIdentifier = naturalGrouping
        }
    }
    
    return bestIdentifier, nil
}

// Discover custom service patterns by analyzing attribute combinations
func (d *DiscoveryEngine) discoverCustomServicePattern(ctx context.Context) *ServiceIdentifier {
    // Get all available attributes
    attributesQuery := `
        SELECT keyset() as attributes
        FROM Transaction, SystemSample
        SINCE 1 hour ago
        LIMIT 1000
    `
    
    attributes, err := d.client.QueryNRDB(ctx, attributesQuery)
    if err != nil {
        return nil
    }
    
    // Look for attributes that might be service identifiers
    servicePatterns := []string{
        ".*service.*",
        ".*app.*",
        ".*application.*",
        ".*component.*",
        ".*system.*",
    }
    
    var candidates []string
    for _, attr := range attributes.Attributes {
        for _, pattern := range servicePatterns {
            if matched, _ := regexp.MatchString(pattern, strings.ToLower(attr)); matched {
                candidates = append(candidates, attr)
                break
            }
        }
    }
    
    // Test each candidate
    bestCandidate := &ServiceIdentifier{Confidence: 0}
    for _, candidate := range candidates {
        identifier := d.evaluateServiceIdentifier(ctx, candidate)
        if identifier.Confidence > bestCandidate.Confidence {
            bestCandidate = identifier
        }
    }
    
    return bestCandidate
}

// Discover natural groupings that could represent services
func (d *DiscoveryEngine) discoverNaturalServiceGrouping(ctx context.Context) *ServiceIdentifier {
    // Find attribute combinations that create natural service boundaries
    query := `
        SELECT count(*) as event_count
        FROM Transaction
        FACET host, port, endpoint
        SINCE 1 hour ago
        LIMIT 100
    `
    
    results, err := d.client.QueryNRDB(ctx, query)
    if err != nil {
        return nil
    }
    
    // Analyze groupings to find service-like patterns
    groupings := d.analyzeGroupings(results)
    
    if len(groupings) > 0 {
        return &ServiceIdentifier{
            Field:      fmt.Sprintf("CONCAT(%s)", strings.Join(groupings[0].Attributes, ", ':', ")),
            Confidence: groupings[0].Confidence,
            Coverage:   groupings[0].Coverage,
            Examples:   groupings[0].Examples,
        }
    }
    
    return nil
}
```

## Error Detection Without Assumptions

We never assume how errors are represented. This code discovers all the ways errors might be indicated in the data.

```go
// discovery/error_discovery.go

type ErrorIndicator struct {
    Method      string   // How we detect errors
    Condition   string   // NRQL condition to use
    Confidence  float64  // How confident we are this indicates errors
    Coverage    float64  // What percentage of events have this indicator
    FalsePositiveRate float64
}

// DiscoverErrorIndicators finds all ways errors are indicated in the data
func (d *DiscoveryEngine) DiscoverErrorIndicators(ctx context.Context, eventType string) ([]ErrorIndicator, error) {
    indicators := []ErrorIndicator{}
    
    // Method 1: Boolean error field
    if indicator := d.checkBooleanErrorField(ctx, eventType); indicator != nil {
        indicators = append(indicators, *indicator)
    }
    
    // Method 2: Error class/type fields
    if indicator := d.checkErrorClassFields(ctx, eventType); indicator != nil {
        indicators = append(indicators, *indicator)
    }
    
    // Method 3: HTTP status codes
    if indicator := d.checkHTTPStatusCodes(ctx, eventType); indicator != nil {
        indicators = append(indicators, *indicator)
    }
    
    // Method 4: Log levels
    if indicator := d.checkLogLevels(ctx, eventType); indicator != nil {
        indicators = append(indicators, *indicator)
    }
    
    // Method 5: Exception fields
    if indicator := d.checkExceptionFields(ctx, eventType); indicator != nil {
        indicators = append(indicators, *indicator)
    }
    
    // Method 6: Response codes (non-HTTP)
    if indicator := d.checkResponseCodes(ctx, eventType); indicator != nil {
        indicators = append(indicators, *indicator)
    }
    
    // Method 7: Custom error patterns
    customIndicators := d.discoverCustomErrorPatterns(ctx, eventType)
    indicators = append(indicators, customIndicators...)
    
    // Method 8: Anomaly-based detection
    if indicator := d.checkAnomalyPatterns(ctx, eventType); indicator != nil {
        indicators = append(indicators, *indicator)
    }
    
    // Sort by confidence
    sort.Slice(indicators, func(i, j int) bool {
        return indicators[i].Confidence > indicators[j].Confidence
    })
    
    return indicators, nil
}

// Check for boolean error field (handles multiple formats)
func (d *DiscoveryEngine) checkBooleanErrorField(ctx context.Context, eventType string) *ErrorIndicator {
    // Check if 'error' field exists
    checkQuery := fmt.Sprintf(`
        SELECT 
            count(*) as total,
            filter(count(*), WHERE error IS NOT NULL) as has_error_field,
            uniqueCount(error) as unique_values,
            capture(toString(error), r'^(true|false|1|0|yes|no|t|f)$') as boolean_like
        FROM %s
        SINCE 1 hour ago
        LIMIT 10000
    `, eventType)
    
    result, err := d.client.QueryNRDB(ctx, checkQuery)
    if err != nil || result.HasErrorField == 0 {
        return nil
    }
    
    coverage := float64(result.HasErrorField) / float64(result.Total)
    
    // Determine the format of the error field
    var condition string
    var confidence float64
    
    if result.UniqueValues == 2 && len(result.BooleanLike) > 0 {
        // It's boolean-like
        sampleQuery := fmt.Sprintf(`
            SELECT count(*) as count, toString(error) as error_value
            FROM %s
            WHERE error IS NOT NULL
            FACET toString(error)
            SINCE 1 hour ago
        `, eventType)
        
        samples, _ := d.client.QueryNRDB(ctx, sampleQuery)
        
        // Determine which values indicate errors
        for _, sample := range samples {
            value := strings.ToLower(sample.ErrorValue)
            if value == "true" || value == "1" || value == "yes" || value == "t" {
                condition = fmt.Sprintf("error IN (true, 'true', 1, '1', 'yes', 't')")
                confidence = 0.95
                break
            }
        }
    } else if result.UniqueValues > 2 {
        // It might be a count or error code
        // Sample to understand the distribution
        distribQuery := fmt.Sprintf(`
            SELECT 
                min(numeric(error)) as min_val,
                max(numeric(error)) as max_val,
                average(numeric(error)) as avg_val
            FROM %s
            WHERE error IS NOT NULL AND numeric(error) IS NOT NULL
            SINCE 1 hour ago
        `, eventType)
        
        distrib, err := d.client.QueryNRDB(ctx, distribQuery)
        if err == nil && distrib.MinVal >= 0 {
            // Likely an error count
            condition = "numeric(error) > 0"
            confidence = 0.85
        }
    }
    
    if condition == "" {
        return nil
    }
    
    // Verify this actually correlates with errors
    falsePositiveRate := d.calculateFalsePositiveRate(ctx, eventType, condition)
    
    return &ErrorIndicator{
        Method:            "boolean_error_field",
        Condition:         condition,
        Confidence:        confidence * (1 - falsePositiveRate),
        Coverage:          coverage,
        FalsePositiveRate: falsePositiveRate,
    }
}

// Discover custom error patterns by analyzing text fields
func (d *DiscoveryEngine) discoverCustomErrorPatterns(ctx context.Context, eventType string) []ErrorIndicator {
    indicators := []ErrorIndicator{}
    
    // Get all string attributes
    stringAttrsQuery := fmt.Sprintf(`
        SELECT keyset() as attributes
        FROM %s
        WHERE message IS NOT NULL 
           OR error_message IS NOT NULL 
           OR exception IS NOT NULL
           OR status_message IS NOT NULL
        SINCE 1 hour ago
        LIMIT 1000
    `, eventType)
    
    attrs, err := d.client.QueryNRDB(ctx, stringAttrsQuery)
    if err != nil {
        return indicators
    }
    
    // Common error patterns to look for
    errorPatterns := []struct {
        Pattern    string
        Confidence float64
    }{
        {`(?i)(error|err|exception|fault|fail)`, 0.8},
        {`(?i)(timeout|timed out)`, 0.7},
        {`(?i)(refused|rejected|denied)`, 0.75},
        {`(?i)(invalid|illegal|bad request)`, 0.7},
        {`(?i)(not found|404|missing)`, 0.6},
        {`(?i)(unauthorized|forbidden|401|403)`, 0.8},
        {`(?i)(internal server|500)`, 0.9},
        {`(?i)(critical|fatal|severe)`, 0.85},
    }
    
    // Test each string attribute for error patterns
    for _, attr := range attrs.Attributes {
        for _, pattern := range errorPatterns {
            testQuery := fmt.Sprintf(`
                SELECT 
                    count(*) as total,
                    filter(count(*), WHERE %s RLIKE '%s') as matches,
                    filter(count(*), WHERE %s NOT RLIKE '%s' AND duration > percentile(duration, 95)) as slow_non_matches
                FROM %s
                WHERE %s IS NOT NULL
                SINCE 1 hour ago
            `, attr, pattern.Pattern, attr, pattern.Pattern, eventType, attr)
            
            result, err := d.client.QueryNRDB(ctx, testQuery)
            if err != nil || result.Matches == 0 {
                continue
            }
            
            matchRate := float64(result.Matches) / float64(result.Total)
            
            // Only consider if match rate is reasonable (not too high, not too low)
            if matchRate > 0.01 && matchRate < 0.5 {
                // Check correlation with performance degradation
                correlation := float64(result.SlowNonMatches) / float64(result.Total - result.Matches)
                
                if correlation < 0.1 { // Low correlation with slowness in non-matches suggests this is an error indicator
                    indicators = append(indicators, ErrorIndicator{
                        Method:            fmt.Sprintf("pattern_match_%s", attr),
                        Condition:         fmt.Sprintf("%s RLIKE '%s'", attr, pattern.Pattern),
                        Confidence:        pattern.Confidence * (1 - correlation),
                        Coverage:          matchRate,
                        FalsePositiveRate: d.estimateFalsePositiveRate(matchRate, correlation),
                    })
                }
            }
        }
    }
    
    return indicators
}

// Check for anomaly-based error detection (no explicit error fields)
func (d *DiscoveryEngine) checkAnomalyPatterns(ctx context.Context, eventType string) *ErrorIndicator {
    // When no explicit error indicators exist, look for anomalous behavior
    anomalyQuery := fmt.Sprintf(`
        SELECT 
            percentile(duration, 50) as p50,
            percentile(duration, 95) as p95,
            percentile(duration, 99) as p99,
            stddev(duration) as stddev
        FROM %s
        WHERE duration IS NOT NULL
        SINCE 1 hour ago
    `, eventType)
    
    stats, err := d.client.QueryNRDB(ctx, anomalyQuery)
    if err != nil || stats.P50 == 0 {
        return nil
    }
    
    // Define anomaly as > p99 or > p50 + 3*stddev
    var condition string
    if stats.Stddev > 0 {
        threshold := stats.P50 + (3 * stats.Stddev)
        if threshold < stats.P99 {
            condition = fmt.Sprintf("duration > %f", threshold)
        } else {
            condition = fmt.Sprintf("duration > %f", stats.P99)
        }
    } else {
        condition = fmt.Sprintf("duration > %f", stats.P99)
    }
    
    // Verify this captures actual errors by checking correlation with known error patterns
    verifyQuery := fmt.Sprintf(`
        SELECT 
            filter(count(*), WHERE %s) as anomalies,
            filter(count(*), WHERE %s AND (
                message RLIKE '(?i)error' OR 
                response_code >= 500 OR 
                status RLIKE '(?i)fail'
            )) as confirmed_errors
        FROM %s
        SINCE 1 hour ago
    `, condition, condition, eventType)
    
    verification, err := d.client.QueryNRDB(ctx, verifyQuery)
    if err != nil || verification.Anomalies == 0 {
        return nil
    }
    
    errorRate := float64(verification.ConfirmedErrors) / float64(verification.Anomalies)
    
    if errorRate > 0.5 { // More than 50% of anomalies are confirmed errors
        return &ErrorIndicator{
            Method:            "anomaly_detection",
            Condition:         condition,
            Confidence:        0.6 * errorRate, // Lower confidence for anomaly-based
            Coverage:          0.01, // Anomalies are rare by definition
            FalsePositiveRate: 1 - errorRate,
        }
    }
    
    return nil
}
```

## Metric Discovery Without Assumptions

We don't assume metric names or types. This discovers what metrics actually exist and how they're structured.

```go
// discovery/metric_discovery.go

type MetricInfo struct {
    Name       string
    Type       string // gauge, counter, histogram, summary
    Unit       string
    Source     string // dimensional, event, custom
    Attributes []AttributeInfo
    Statistics MetricStatistics
}

// DiscoverMetrics finds all metrics without assuming naming conventions
func (d *DiscoveryEngine) DiscoverMetrics(ctx context.Context, pattern string) ([]MetricInfo, error) {
    metrics := []MetricInfo{}
    
    // Method 1: Discover dimensional metrics
    dimensionalMetrics := d.discoverDimensionalMetrics(ctx, pattern)
    metrics = append(metrics, dimensionalMetrics...)
    
    // Method 2: Discover metrics in events
    eventMetrics := d.discoverEventMetrics(ctx, pattern)
    metrics = append(metrics, eventMetrics...)
    
    // Method 3: Discover custom metrics
    customMetrics := d.discoverCustomMetrics(ctx, pattern)
    metrics = append(metrics, customMetrics...)
    
    // Deduplicate and enrich
    metrics = d.deduplicateAndEnrichMetrics(ctx, metrics)
    
    return metrics, nil
}

// Discover dimensional metrics without assuming structure
func (d *DiscoveryEngine) discoverDimensionalMetrics(ctx context.Context, pattern string) []MetricInfo {
    metrics := []MetricInfo{}
    
    // We can't assume the metric table exists or has standard fields
    // First check what metric-like tables exist
    tableQuery := `SHOW TABLES LIKE '%metric%'`
    tables, err := d.client.QueryNRDB(ctx, tableQuery)
    if err != nil {
        // Try alternative approach
        tables = d.discoverMetricTables(ctx)
    }
    
    for _, table := range tables {
        // For each potential metric table, discover its structure
        structureQuery := fmt.Sprintf(`
            SELECT keyset() as attributes
            FROM %s
            SINCE 5 minutes ago
            LIMIT 100
        `, table)
        
        structure, err := d.client.QueryNRDB(ctx, structureQuery)
        if err != nil {
            continue
        }
        
        // Look for metric-like attributes without assuming names
        metricAttrs := d.identifyMetricAttributes(structure.Attributes)
        
        for _, attr := range metricAttrs {
            // Discover metric properties
            propertiesQuery := fmt.Sprintf(`
                SELECT 
                    min(%s) as min_value,
                    max(%s) as max_value,
                    average(%s) as avg_value,
                    stddev(%s) as stddev_value,
                    uniqueCount(%s) as cardinality
                FROM %s
                WHERE %s IS NOT NULL
                SINCE 1 hour ago
            `, attr, attr, attr, attr, attr, table, attr)
            
            props, err := d.client.QueryNRDB(ctx, propertiesQuery)
            if err != nil {
                continue
            }
            
            // Determine metric type from behavior
            metricType := d.inferMetricType(props)
            
            // Discover associated dimensions
            dimensions := d.discoverMetricDimensions(ctx, table, attr)
            
            metrics = append(metrics, MetricInfo{
                Name:   attr,
                Type:   metricType,
                Source: "dimensional",
                Unit:   d.inferUnit(attr, props),
                Attributes: dimensions,
                Statistics: MetricStatistics{
                    Min:    props.MinValue,
                    Max:    props.MaxValue,
                    Avg:    props.AvgValue,
                    Stddev: props.StddevValue,
                },
            })
        }
    }
    
    return metrics
}

// Identify which attributes are likely metrics without assuming names
func (d *DiscoveryEngine) identifyMetricAttributes(attributes []string) []string {
    metricLike := []string{}
    
    for _, attr := range attributes {
        // Skip obvious non-metrics
        if d.isDefinitelyNotMetric(attr) {
            continue
        }
        
        // Check if it could be a metric
        // We don't assume naming conventions, so we check behavior
        if d.couldBeMetric(attr) {
            metricLike = append(metricLike, attr)
        }
    }
    
    return metricLike
}

// Determine if an attribute could be a metric by its characteristics
func (d *DiscoveryEngine) couldBeMetric(attr string) bool {
    // Don't assume based on name alone
    // Instead, we'll check its behavior in the actual query
    
    // Common non-metric patterns (but still not assumed)
    nonMetricPatterns := []string{
        `^(id|guid|uuid|key)$`,
        `^.*_(id|guid|uuid|key)$`,
        `^(name|label|tag|type|status|state)$`,
        `^.*_(name|label|tag|type|status|state)$`,
    }
    
    attrLower := strings.ToLower(attr)
    for _, pattern := range nonMetricPatterns {
        if matched, _ := regexp.MatchString(pattern, attrLower); matched {
            return false
        }
    }
    
    // If it's not obviously not a metric, it could be
    return true
}

// Infer metric type from its statistical behavior
func (d *DiscoveryEngine) inferMetricType(stats MetricStats) string {
    // Counters only increase
    if stats.MinValue >= 0 && stats.IncreasingOverTime {
        return "counter"
    }
    
    // Histograms have percentile data
    if stats.HasPercentiles {
        return "histogram"
    }
    
    // Gauges can go up and down
    if stats.FluctuatesOverTime {
        return "gauge"
    }
    
    // Summary has count and sum
    if stats.HasCountAndSum {
        return "summary"
    }
    
    return "unknown"
}

// Discover custom metrics in non-standard locations
func (d *DiscoveryEngine) discoverCustomMetrics(ctx context.Context, pattern string) []MetricInfo {
    metrics := []MetricInfo{}
    
    // Look for numeric fields in all event types
    eventTypesQuery := `SHOW EVENT TYPES SINCE 24 hours ago`
    eventTypes, err := d.client.QueryNRDB(ctx, eventTypesQuery)
    if err != nil {
        return metrics
    }
    
    for _, eventType := range eventTypes {
        // Skip if we already processed this as a metric table
        if d.isMetricTable(eventType) {
            continue
        }
        
        // Discover numeric fields
        numericFieldsQuery := fmt.Sprintf(`
            SELECT keyset() as all_fields
            FROM %s
            SINCE 1 hour ago
            LIMIT 1000
        `, eventType)
        
        fields, err := d.client.QueryNRDB(ctx, numericFieldsQuery)
        if err != nil {
            continue
        }
        
        for _, field := range fields.AllFields {
            // Check if field contains numeric data
            checkQuery := fmt.Sprintf(`
                SELECT 
                    numeric(%s) as numeric_value,
                    count(*) as count
                FROM %s
                WHERE numeric(%s) IS NOT NULL
                SINCE 5 minutes ago
                LIMIT 10
            `, field, eventType, field)
            
            result, err := d.client.QueryNRDB(ctx, checkQuery)
            if err != nil || result.Count == 0 {
                continue
            }
            
            // It's numeric, analyze its behavior
            behavior := d.analyzeNumericFieldBehavior(ctx, eventType, field)
            
            if behavior.IsMetricLike {
                metrics = append(metrics, MetricInfo{
                    Name:   fmt.Sprintf("%s.%s", eventType, field),
                    Type:   behavior.MetricType,
                    Source: "event_field",
                    Unit:   behavior.Unit,
                    Attributes: behavior.Dimensions,
                    Statistics: behavior.Statistics,
                })
            }
        }
    }
    
    return metrics
}
```

## Dashboard Generation Without Assumptions

Generate dashboards based entirely on discovered data, not templates.

```go
// dashboard/adaptive_generator.go

type AdaptiveDashboardGenerator struct {
    discovery *DiscoveryEngine
}

// GenerateDashboard creates a dashboard without any assumptions about data structure
func (g *AdaptiveDashboardGenerator) GenerateDashboard(ctx context.Context, scope DashboardScope) (*Dashboard, error) {
    dashboard := &Dashboard{
        Name:        scope.Name,
        Description: "Auto-generated based on discovered data",
        Pages:       []Page{},
    }
    
    // Discover all available data for the scope
    availableData := g.discoverScopeData(ctx, scope)
    
    // Create pages based on what we found
    if len(availableData.PerformanceMetrics) > 0 {
        dashboard.Pages = append(dashboard.Pages, g.createPerformancePage(ctx, availableData))
    }
    
    if len(availableData.ErrorIndicators) > 0 {
        dashboard.Pages = append(dashboard.Pages, g.createErrorPage(ctx, availableData))
    }
    
    if len(availableData.InfrastructureMetrics) > 0 {
        dashboard.Pages = append(dashboard.Pages, g.createInfrastructurePage(ctx, availableData))
    }
    
    if len(availableData.BusinessMetrics) > 0 {
        dashboard.Pages = append(dashboard.Pages, g.createBusinessPage(ctx, availableData))
    }
    
    // If we found very little data, create a discovery page
    if len(dashboard.Pages) == 0 {
        dashboard.Pages = append(dashboard.Pages, g.createDiscoveryPage(ctx, scope))
    }
    
    return dashboard, nil
}

// Discover all data related to the scope without assumptions
func (g *AdaptiveDashboardGenerator) discoverScopeData(ctx context.Context, scope DashboardScope) *ScopeData {
    data := &ScopeData{}
    
    // Discover how to identify scope entities
    scopeIdentifier := g.discovery.DiscoverScopeIdentifier(ctx, scope)
    
    // Find all event types that contain data for this scope
    eventTypes := g.discovery.DiscoverScopeEventTypes(ctx, scopeIdentifier)
    
    // For each event type, discover available metrics
    for _, eventType := range eventTypes {
        // Discover performance metrics
        perfMetrics := g.discoverPerformanceMetrics(ctx, eventType, scopeIdentifier)
        data.PerformanceMetrics = append(data.PerformanceMetrics, perfMetrics...)
        
        // Discover error indicators
        errorIndicators := g.discovery.DiscoverErrorIndicators(ctx, eventType)
        data.ErrorIndicators = append(data.ErrorIndicators, errorIndicators...)
        
        // Discover infrastructure metrics
        infraMetrics := g.discoverInfrastructureMetrics(ctx, eventType, scopeIdentifier)
        data.InfrastructureMetrics = append(data.InfrastructureMetrics, infraMetrics...)
        
        // Discover business metrics
        bizMetrics := g.discoverBusinessMetrics(ctx, eventType, scopeIdentifier)
        data.BusinessMetrics = append(data.BusinessMetrics, bizMetrics...)
    }
    
    return data
}

// Create performance page based on discovered metrics
func (g *AdaptiveDashboardGenerator) createPerformancePage(ctx context.Context, data *ScopeData) Page {
    page := Page{
        Name:    "Performance",
        Widgets: []Widget{},
    }
    
    // Response time widget - but we don't assume the metric name
    if responseMetric := g.findBestResponseTimeMetric(data.PerformanceMetrics); responseMetric != nil {
        widget := g.createResponseTimeWidget(ctx, responseMetric)
        page.Widgets = append(page.Widgets, widget)
    }
    
    // Throughput widget - discover what represents throughput
    if throughputMetric := g.discoverThroughputMetric(ctx, data); throughputMetric != nil {
        widget := g.createThroughputWidget(ctx, throughputMetric)
        page.Widgets = append(page.Widgets, widget)
    }
    
    // Latency breakdown - if we can discover components
    if components := g.discoverLatencyComponents(ctx, data); len(components) > 0 {
        widget := g.createLatencyBreakdownWidget(ctx, components)
        page.Widgets = append(page.Widgets, widget)
    }
    
    // Performance distribution
    if distMetrics := g.findDistributionMetrics(data.PerformanceMetrics); len(distMetrics) > 0 {
        widget := g.createDistributionWidget(ctx, distMetrics)
        page.Widgets = append(page.Widgets, widget)
    }
    
    return page
}

// Find the best response time metric without assuming its name
func (g *AdaptiveDashboardGenerator) findBestResponseTimeMetric(metrics []MetricInfo) *MetricInfo {
    // Score each metric based on how likely it is to be response time
    type ScoredMetric struct {
        Metric *MetricInfo
        Score  float64
    }
    
    var scored []ScoredMetric
    
    for _, metric := range metrics {
        score := 0.0
        
        // Check name patterns (but don't rely on them)
        nameLower := strings.ToLower(metric.Name)
        if strings.Contains(nameLower, "duration") {
            score += 0.3
        }
        if strings.Contains(nameLower, "response") {
            score += 0.2
        }
        if strings.Contains(nameLower, "latency") {
            score += 0.2
        }
        if strings.Contains(nameLower, "time") && !strings.Contains(nameLower, "timestamp") {
            score += 0.1
        }
        
        // Check unit (more reliable)
        if metric.Unit == "ms" || metric.Unit == "milliseconds" {
            score += 0.4
        }
        if metric.Unit == "s" || metric.Unit == "seconds" {
            score += 0.3
        }
        
        // Check statistical properties
        if metric.Statistics.Min >= 0 && metric.Statistics.Max < 300000 { // Less than 5 minutes
            score += 0.2
        }
        if metric.Statistics.Avg > 0 && metric.Statistics.Avg < 10000 { // Reasonable response time
            score += 0.2
        }
        
        // Check if it correlates with request count
        if g.correlatesWithRequests(metric) {
            score += 0.3
        }
        
        if score > 0 {
            scored = append(scored, ScoredMetric{&metric, score})
        }
    }
    
    // Sort by score
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].Score > scored[j].Score
    })
    
    if len(scored) > 0 {
        return scored[0].Metric
    }
    
    return nil
}

// Create a widget that adapts to the actual metric structure
func (g *AdaptiveDashboardGenerator) createResponseTimeWidget(ctx context.Context, metric *MetricInfo) Widget {
    // Build query based on discovered metric structure
    var query string
    
    if metric.Source == "dimensional" {
        // Use dimensional metric query
        query = g.buildDimensionalMetricQuery(metric)
    } else if metric.Source == "event_field" {
        // Use NRQL query
        query = g.buildEventFieldQuery(metric)
    } else {
        // Custom source
        query = g.buildCustomMetricQuery(metric)
    }
    
    // Determine best visualization based on metric properties
    vizType := g.determineBestVisualization(metric)
    
    return Widget{
        Title:          g.generateWidgetTitle(metric),
        Visualization:  vizType,
        Configuration: WidgetConfig{
            Query:      query,
            ChartType:  vizType,
            YAxisLabel: g.formatUnit(metric.Unit),
            Colors:     g.selectColors(metric),
        },
    }
}

// Discover what represents throughput without assumptions
func (g *AdaptiveDashboardGenerator) discoverThroughputMetric(ctx context.Context, data *ScopeData) *ThroughputMetric {
    // Throughput could be represented in many ways
    candidates := []ThroughputCandidate{}
    
    // Method 1: Look for rate metrics
    for _, metric := range data.PerformanceMetrics {
        if g.isRateMetric(metric) {
            candidates = append(candidates, ThroughputCandidate{
                Type:   "rate_metric",
                Metric: metric,
                Score:  g.scoreRateMetric(metric),
            })
        }
    }
    
    // Method 2: Look for count metrics that we can convert to rate
    for _, metric := range data.PerformanceMetrics {
        if g.isCountMetric(metric) {
            candidates = append(candidates, ThroughputCandidate{
                Type:   "count_metric",
                Metric: metric,
                Score:  g.scoreCountMetric(metric),
                NeedsRateConversion: true,
            })
        }
    }
    
    // Method 3: Discover from event counts
    eventThroughput := g.discoverEventBasedThroughput(ctx, data)
    if eventThroughput != nil {
        candidates = append(candidates, *eventThroughput)
    }
    
    // Method 4: Derive from other metrics
    derivedThroughput := g.deriveThroughputFromOtherMetrics(ctx, data)
    if derivedThroughput != nil {
        candidates = append(candidates, *derivedThroughput)
    }
    
    // Select best candidate
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Score > candidates[j].Score
    })
    
    if len(candidates) > 0 {
        return g.convertToThroughputMetric(candidates[0])
    }
    
    return nil
}
```

## Performance Analysis Without Assumptions

Analyze performance without assuming what metrics exist or what "normal" looks like.

```go
// analysis/adaptive_performance.go

type PerformanceAnalyzer struct {
    discovery *DiscoveryEngine
}

// AnalyzePerformance without any assumptions about metrics or baselines
func (a *PerformanceAnalyzer) AnalyzePerformance(ctx context.Context, scope AnalysisScope) (*PerformanceAnalysis, error) {
    analysis := &PerformanceAnalysis{
        Scope:     scope,
        Findings:  []Finding{},
        Baselines: map[string]Baseline{},
    }
    
    // Discover what performance metrics exist
    metrics := a.discoverPerformanceMetrics(ctx, scope)
    
    // For each metric, establish baseline without assumptions
    for _, metric := range metrics {
        baseline := a.establishBaseline(ctx, metric, scope)
        analysis.Baselines[metric.Name] = baseline
    }
    
    // Analyze current performance against discovered baselines
    currentPerf := a.analyzeCurrentPerformance(ctx, metrics, analysis.Baselines)
    
    // Identify anomalies without assuming what's normal
    anomalies := a.identifyAnomalies(ctx, currentPerf, analysis.Baselines)
    
    // Discover correlations without assuming relationships
    correlations := a.discoverCorrelations(ctx, metrics, anomalies)
    
    // Build findings
    for _, anomaly := range anomalies {
        finding := a.buildFinding(anomaly, correlations)
        analysis.Findings = append(analysis.Findings, finding)
    }
    
    return analysis, nil
}

// Establish baseline without assuming distribution or patterns
func (a *PerformanceAnalyzer) establishBaseline(ctx context.Context, metric MetricInfo, scope AnalysisScope) Baseline {
    baseline := Baseline{
        Metric: metric.Name,
    }
    
    // Don't assume time windows - discover meaningful periods
    timeWindows := a.discoverMeaningfulTimeWindows(ctx, metric)
    
    for _, window := range timeWindows {
        // Don't assume statistical distribution
        distribution := a.discoverDistribution(ctx, metric, window)
        
        // Don't assume patterns - discover them
        patterns := a.discoverPatterns(ctx, metric, window)
        
        baseline.Windows = append(baseline.Windows, BaselineWindow{
            Period:       window,
            Distribution: distribution,
            Patterns:     patterns,
        })
    }
    
    // Discover what constitutes "normal" for this specific metric
    baseline.NormalBehavior = a.discoverNormalBehavior(ctx, metric, baseline.Windows)
    
    return baseline
}

// Discover meaningful time windows based on data patterns
func (a *PerformanceAnalyzer) discoverMeaningfulTimeWindows(ctx context.Context, metric MetricInfo) []TimeWindow {
    windows := []TimeWindow{}
    
    // Test different time windows to find natural patterns
    testWindows := []string{
        "5 minutes", "15 minutes", "1 hour", 
        "4 hours", "1 day", "1 week", "1 month",
    }
    
    for _, window := range testWindows {
        // Check if this window reveals patterns
        patternQuery := fmt.Sprintf(`
            SELECT 
                stddev(%s) as variability,
                uniqueCount(capture(toString(timestamp), r'T(\d{2}):\d{2}')) as hourly_periods,
                uniqueCount(capture(toString(timestamp), r'(\d{4}-\d{2}-\d{2})')) as daily_periods
            FROM %s
            WHERE %s IS NOT NULL
            SINCE %s
        `, metric.Name, metric.Source, metric.Name, window)
        
        result, err := a.discovery.Query(ctx, patternQuery)
        if err != nil {
            continue
        }
        
        // Determine if this window is meaningful
        isMeaningful := false
        
        // High variability suggests interesting patterns
        if result.Variability > metric.Statistics.Stddev*1.5 {
            isMeaningful = true
        }
        
        // Multiple time periods suggest patterns
        if window == "1 day" && result.HourlyPeriods > 20 {
            isMeaningful = true
        }
        if window == "1 week" && result.DailyPeriods > 5 {
            isMeaningful = true
        }
        
        if isMeaningful {
            windows = append(windows, TimeWindow{
                Duration:    window,
                Granularity: a.determineGranularity(window),
                Purpose:     a.inferWindowPurpose(window, result),
            })
        }
    }
    
    // If no meaningful windows found, use defaults but mark as uncertain
    if len(windows) == 0 {
        windows = append(windows, TimeWindow{
            Duration:    "1 hour",
            Granularity: "1 minute",
            Purpose:     "recent_behavior",
            Uncertain:   true,
        })
    }
    
    return windows
}

// Discover distribution without assuming normal/gaussian
func (a *PerformanceAnalyzer) discoverDistribution(ctx context.Context, metric MetricInfo, window TimeWindow) Distribution {
    dist := Distribution{}
    
    // Get raw percentiles - don't assume which ones matter
    percentileQuery := fmt.Sprintf(`
        SELECT 
            count(*) as sample_count,
            min(%s) as min,
            percentile(%s, 1) as p1,
            percentile(%s, 5) as p5,
            percentile(%s, 10) as p10,
            percentile(%s, 25) as p25,
            percentile(%s, 50) as p50,
            percentile(%s, 75) as p75,
            percentile(%s, 90) as p90,
            percentile(%s, 95) as p95,
            percentile(%s, 99) as p99,
            percentile(%s, 99.9) as p999,
            max(%s) as max,
            average(%s) as mean,
            stddev(%s) as stddev
        FROM %s
        WHERE %s IS NOT NULL
        SINCE %s
    `, metric.Name, metric.Name, metric.Name, metric.Name, metric.Name, 
       metric.Name, metric.Name, metric.Name, metric.Name, metric.Name,
       metric.Name, metric.Name, metric.Name, metric.Name,
       metric.Source, metric.Name, window.Duration)
    
    result, err := a.discovery.Query(ctx, percentileQuery)
    if err != nil {
        return dist
    }
    
    dist.SampleSize = result.SampleCount
    dist.Percentiles = result.Percentiles
    dist.Statistics = result.Stats
    
    // Determine distribution type by shape analysis
    dist.Type = a.analyzeDistributionShape(result)
    
    // Identify outlier boundaries based on actual distribution
    dist.OutlierBounds = a.discoverOutlierBounds(result, dist.Type)
    
    // Check for multimodal distributions
    if dist.Type == "unknown" || dist.Type == "multimodal" {
        dist.Modes = a.discoverModes(ctx, metric, window)
    }
    
    return dist
}

// Analyze distribution shape without assumptions
func (a *PerformanceAnalyzer) analyzeDistributionShape(data DistributionData) string {
    // Check for normal distribution
    // Mean should be close to median (p50)
    meanMedianRatio := data.Mean / data.P50
    if meanMedianRatio > 0.9 && meanMedianRatio < 1.1 {
        // Check if roughly 68% fall within 1 stddev
        withinOneStddev := (data.P84 - data.P16) / (2 * data.Stddev)
        if withinOneStddev > 0.6 && withinOneStddev < 0.8 {
            return "normal"
        }
    }
    
    // Check for log-normal (common for response times)
    if data.Mean > data.P50 && data.P90 > data.Mean + data.Stddev {
        // Long tail on the right
        return "log-normal"
    }
    
    // Check for exponential
    if data.Mean > 0 && math.Abs(data.Mean-data.Stddev) < data.Mean*0.1 {
        return "exponential"
    }
    
    // Check for bimodal
    gap := data.P75 - data.P25
    if (data.P25-data.P10) > gap || (data.P90-data.P75) > gap {
        return "multimodal"
    }
    
    // Check for uniform
    if data.Max > 0 && data.Min >= 0 {
        expectedUniformStddev := (data.Max - data.Min) / math.Sqrt(12)
        if math.Abs(data.Stddev-expectedUniformStddev) < expectedUniformStddev*0.1 {
            return "uniform"
        }
    }
    
    return "unknown"
}

// Identify anomalies without assuming what's normal
func (a *PerformanceAnalyzer) identifyAnomalies(ctx context.Context, current PerformanceData, baselines map[string]Baseline) []Anomaly {
    anomalies := []Anomaly{}
    
    for metricName, baseline := range baselines {
        currentValue := current.Metrics[metricName]
        
        // Don't assume anomaly detection method - try multiple
        methods := []AnomalyDetector{
            // Statistical deviation
            &StatisticalAnomalyDetector{
                Baseline: baseline,
                // Don't assume sensitivity - calculate from data
                Sensitivity: a.calculateOptimalSensitivity(baseline),
            },
            
            // Pattern deviation
            &PatternAnomalyDetector{
                Baseline: baseline,
                // Discover which patterns matter
                SignificantPatterns: a.discoverSignificantPatterns(baseline),
            },
            
            // Contextual anomaly (depends on other metrics)
            &ContextualAnomalyDetector{
                AllBaselines: baselines,
                // Discover metric relationships
                Relationships: a.discoverMetricRelationships(ctx, baselines),
            },
            
            // Collective anomaly (group behavior)
            &CollectiveAnomalyDetector{
                Baseline: baseline,
                // Discover peer groups
                PeerGroups: a.discoverPeerGroups(ctx, metricName),
            },
        }
        
        for _, detector := range methods {
            if anomaly := detector.Detect(currentValue); anomaly != nil {
                anomaly.DetectionMethod = detector.Name()
                anomaly.Confidence = a.calculateAnomalyConfidence(anomaly, baseline)
                anomalies = append(anomalies, *anomaly)
            }
        }
    }
    
    // Deduplicate and prioritize anomalies
    anomalies = a.consolidateAnomalies(anomalies)
    
    return anomalies
}
```

## Cost Analysis Without Assumptions

Analyze costs without assuming pricing models or data sources.

```go
// cost/adaptive_analyzer.go

type CostAnalyzer struct {
    discovery *DiscoveryEngine
}

// AnalyzeCosts without assuming pricing or data structure
func (c *CostAnalyzer) AnalyzeCosts(ctx context.Context, scope CostScope) (*CostAnalysis, error) {
    analysis := &CostAnalysis{
        Scope: scope,
    }
    
    // Discover all data sources (don't assume what exists)
    sources := c.discoverDataSources(ctx)
    
    // Discover pricing model (don't assume how billing works)
    pricing := c.discoverPricingModel(ctx)
    
    // Analyze each source
    for _, source := range sources {
        sourceAnalysis := c.analyzeDataSource(ctx, source, pricing)
        analysis.Sources = append(analysis.Sources, sourceAnalysis)
    }
    
    // Discover optimization opportunities
    analysis.Opportunities = c.discoverOptimizations(ctx, analysis.Sources)
    
    // Calculate potential savings
    analysis.PotentialSavings = c.calculateSavings(analysis.Opportunities)
    
    return analysis, nil
}

// Discover all data sources without assumptions
func (c *CostAnalyzer) discoverDataSources(ctx context.Context) []DataSource {
    sources := []DataSource{}
    
    // Method 1: Check NerdGraph for ingest sources
    ingestSources := c.discoverIngestSources(ctx)
    sources = append(sources, ingestSources...)
    
    // Method 2: Analyze event types for source indicators
    eventSources := c.discoverEventSources(ctx)
    sources = append(sources, eventSources...)
    
    // Method 3: Check for custom data sources
    customSources := c.discoverCustomSources(ctx)
    sources = append(sources, customSources...)
    
    // Deduplicate and enrich
    sources = c.consolidateDataSources(sources)
    
    return sources
}

// Discover pricing model without assumptions
func (c *CostAnalyzer) discoverPricingModel(ctx context.Context) PricingModel {
    model := PricingModel{}
    
    // We can't assume pricing structure, so we need to infer it
    
    // Method 1: Analyze billing data if available
    if billingData := c.checkBillingData(ctx); billingData != nil {
        model = c.inferPricingFromBilling(billingData)
    }
    
    // Method 2: Correlate data volume with costs
    if correlation := c.correlateVolumeWithCosts(ctx); correlation != nil {
        model = c.inferPricingFromCorrelation(correlation)
    }
    
    // Method 3: Use known patterns but verify
    if model.IsEmpty() {
        model = c.useDefaultPricingWithVerification(ctx)
    }
    
    return model
}

// Analyze data source without assuming structure
func (c *CostAnalyzer) analyzeDataSource(ctx context.Context, source DataSource, pricing PricingModel) DataSourceAnalysis {
    analysis := DataSourceAnalysis{
        Source: source,
    }
    
    // Discover volume without assuming metrics
    analysis.Volume = c.discoverDataVolume(ctx, source)
    
    // Discover usage patterns
    analysis.UsagePatterns = c.discoverUsagePatterns(ctx, source)
    
    // Discover value (what queries use this data)
    analysis.Value = c.discoverDataValue(ctx, source)
    
    // Calculate costs based on discovered pricing
    analysis.Cost = c.calculateSourceCost(analysis.Volume, pricing)
    
    // Determine efficiency
    analysis.Efficiency = c.calculateEfficiency(analysis)
    
    return analysis
}

// Discover optimization opportunities without assumptions
func (c *CostAnalyzer) discoverOptimizations(ctx context.Context, sources []DataSourceAnalysis) []Optimization {
    optimizations := []Optimization{}
    
    // Don't assume what optimizations are possible
    optimizers := []Optimizer{
        // Unused data optimizer
        &UnusedDataOptimizer{
            discovery: c.discovery,
        },
        
        // Over-collection optimizer
        &OverCollectionOptimizer{
            discovery: c.discovery,
        },
        
        // Redundancy optimizer
        &RedundancyOptimizer{
            discovery: c.discovery,
        },
        
        // Granularity optimizer
        &GranularityOptimizer{
            discovery: c.discovery,
        },
        
        // Format optimizer (events to metrics)
        &FormatOptimizer{
            discovery: c.discovery,
        },
        
        // Source optimizer (agent vs OTEL)
        &SourceOptimizer{
            discovery: c.discovery,
        },
    }
    
    for _, optimizer := range optimizers {
        opts := optimizer.Discover(ctx, sources)
        optimizations = append(optimizations, opts...)
    }
    
    // Validate optimizations won't break anything
    for i, opt := range optimizations {
        impact := c.assessOptimizationImpact(ctx, opt)
        optimizations[i].Impact = impact
        optimizations[i].Risk = c.calculateRisk(impact)
    }
    
    // Sort by value/risk ratio
    sort.Slice(optimizations, func(i, j int) bool {
        ratioI := optimizations[i].EstimatedSavings / float64(optimizations[i].Risk+1)
        ratioJ := optimizations[j].EstimatedSavings / float64(optimizations[j].Risk+1)
        return ratioI > ratioJ
    })
    
    return optimizations
}

// Example: Format optimizer that discovers migration opportunities
type FormatOptimizer struct {
    discovery *DiscoveryEngine
}

func (f *FormatOptimizer) Discover(ctx context.Context, sources []DataSourceAnalysis) []Optimization {
    optimizations := []Optimization{}
    
    for _, source := range sources {
        // Don't assume events can be converted to metrics
        // Discover if it's possible
        
        if source.Source.Type == "event" {
            // Analyze query patterns
            queryAnalysis := f.analyzeQueryPatterns(ctx, source)
            
            // Check if queries are aggregation-only
            if queryAnalysis.AggregationOnlyPercentage > 0.8 {
                // Discover which fields are queried
                queriedFields := f.discoverQueriedFields(ctx, source)
                
                // Check if fields are metric-like
                metricLikeFields := []string{}
                for _, field := range queriedFields {
                    if f.isMetricLike(ctx, source.Source.Name, field) {
                        metricLikeFields = append(metricLikeFields, field)
                    }
                }
                
                if len(metricLikeFields) > 0 {
                    // Calculate potential savings
                    currentSize := source.Volume.Bytes
                    metricSize := f.estimateMetricSize(metricLikeFields, source.Volume.EventCount)
                    savings := currentSize - metricSize
                    
                    optimizations = append(optimizations, Optimization{
                        Type:        "event_to_metric",
                        Description: fmt.Sprintf("Convert %s events to dimensional metrics", source.Source.Name),
                        Source:      source.Source,
                        Details: map[string]interface{}{
                            "fields_to_convert":  metricLikeFields,
                            "current_size_gb":    currentSize / 1e9,
                            "projected_size_gb":  metricSize / 1e9,
                            "query_compatibility": queryAnalysis.AggregationOnlyPercentage,
                        },
                        EstimatedSavings: savings,
                        Implementation:   f.generateImplementationPlan(source, metricLikeFields),
                    })
                }
            }
        }
    }
    
    return optimizations
}

// Dashboard optimization without assuming widget types
func (c *CostAnalyzer) optimizeDashboards(ctx context.Context) []DashboardOptimization {
    optimizations := []DashboardOptimization{}
    
    // Discover all dashboards
    dashboards := c.discovery.DiscoverDashboards(ctx)
    
    for _, dashboard := range dashboards {
        // Analyze widgets without assuming their structure
        widgetAnalysis := c.analyzeWidgets(ctx, dashboard)
        
        for _, widget := range widgetAnalysis {
            // Don't assume widget configuration
            config := c.discovery.ParseWidgetConfig(widget.RawConfig)
            
            // Check if it's using events where metrics would work
            if config.DataSource == "nrql" {
                metricAlternative := c.discoverMetricAlternative(ctx, config.Query)
                
                if metricAlternative != nil {
                    optimizations = append(optimizations, DashboardOptimization{
                        Dashboard: dashboard,
                        Widget:    widget,
                        Current:   config,
                        Proposed:  metricAlternative,
                        Savings:   c.calculateWidgetSavings(config, metricAlternative),
                    })
                }
            }
        }
    }
    
    return optimizations
}
```

## Conclusion

These examples demonstrate our extreme commitment to avoiding assumptions. Every piece of code:

1. **Discovers** rather than assumes
2. **Adapts** based on findings  
3. **Validates** before executing
4. **Handles** variations gracefully
5. **Explains** what it discovered

This approach ensures our tools work in ANY environment, with ANY schema, under ANY conditions. No assumptions. Ever.
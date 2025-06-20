# Discovery-Driven Investigation Examples

## Real-World Scenarios Using Discovery-First Approach

### Scenario 1: "The website is slow"

#### Traditional Approach (Assumption-Based)
```yaml
traditional_investigation:
  assumptions:
    - "Transaction table exists"
    - "duration attribute is populated"
    - "appName identifies services"
    - "We know what 'slow' means"
  
  problems:
    - What if duration isn't collected?
    - What if the issue isn't in Transaction data?
    - What if 'slow' means something else?
```

#### Discovery-First Approach
```yaml
discovery_investigation:
  phase_1_discover_what_exists:
    - tool: discovery.explore_event_types
      question: "What data is actually being collected?"
      result: 
        - Transaction (1.2M events)
        - PageView (800K events)
        - JavaScriptError (15K events)
        - SyntheticCheck (5K events)
        
    - tool: discovery.explore_attributes
      question: "What does Transaction data contain?"
      params:
        event_type: "Transaction"
      result:
        - duration: 98% coverage
        - queueDuration: 45% coverage (sparse!)
        - externalDuration: 76% coverage
        - name: 100% coverage
        
  phase_2_understand_slow:
    - tool: nrql.execute
      question: "What does the duration distribution look like?"
      query: |
        SELECT histogram(duration, width: 100, buckets: 20)
        FROM Transaction
        SINCE 1 hour ago
      discovery: "Bimodal distribution - two distinct performance profiles"
      
    - tool: discovery.find_natural_groupings
      question: "What creates these two profiles?"
      params:
        event_type: "Transaction"
      result:
        - Group by 'name': API vs web transactions
        - Group by 'request.uri': Different endpoints
        - Group by 'host': Different server pools
        
  phase_3_pinpoint_issue:
    - tool: nrql.execute
      question: "Which group degraded?"
      query: |
        SELECT average(duration)
        FROM Transaction
        FACET name
        SINCE 30 minutes ago
        COMPARE WITH 2 hours ago
      discovery: "Only '/api/checkout' endpoints degraded"
      
  phase_4_understand_checkout:
    - tool: discovery.find_data_relationships
      question: "What other data relates to checkout?"
      params:
        source_event_type: "Transaction"
        filter: "name = '/api/checkout'"
      result:
        - Span events with same trace.id
        - Log events with same entity.guid
        - DatabaseSample with matching host
        
  phase_5_root_cause:
    - tool: nrql.execute
      question: "What changed in related data?"
      query: |
        SELECT average(db.duration)
        FROM Span
        WHERE trace.id IN (
          SELECT uniques(trace.id)
          FROM Transaction
          WHERE name = '/api/checkout'
        )
        TIMESERIES
        SINCE 2 hours ago
      discovery: "Database queries started taking 10x longer at 14:32"
```

### Scenario 2: "We're getting errors"

#### Discovery-First Investigation
```yaml
error_investigation:
  phase_1_understand_errors:
    - tool: discovery.explore_event_types
      question: "Where might errors be recorded?"
      result:
        event_types_with_error_info:
          - Transaction: has 'error' boolean
          - Log: has 'level' attribute  
          - JavaScriptError: dedicated error events
          - SyntheticCheck: has 'result' attribute
          
    - tool: discovery.profile_attribute_values
      question: "What does 'error' mean in Transaction?"
      params:
        event_type: "Transaction"
        attribute: "error"
      result:
        - Type: boolean
        - Coverage: 95% (5% null - investigate!)
        - True rate: 0.5% baseline, now 15%
        
  phase_2_characterize_errors:
    - tool: nrql.execute
      question: "Do we have error details?"
      query: |
        SELECT keyset()
        FROM Transaction
        WHERE error IS true
        SINCE 10 minutes ago
        LIMIT 1
      discovery: "error.class and error.message exist!"
      
    - tool: nrql.execute
      question: "What kinds of errors?"
      query: |
        SELECT count(*)
        FROM Transaction
        WHERE error IS true
        FACET error.class, error.message
        SINCE 30 minutes ago
      discovery:
        - 90% are "DatabaseConnection: timeout"
        - 8% are "NullPointerException"
        - 2% are various others
        
  phase_3_trace_error_source:
    - tool: discovery.detect_temporal_patterns
      question: "When did errors start?"
      params:
        query: "SELECT percentage(count(*), WHERE error IS true) FROM Transaction"
      result:
        - Spike started at 15:45:32
        - Pattern: Sudden increase, not gradual
        
    - tool: nrql.execute
      question: "What was the first error?"
      query: |
        SELECT min(timestamp), error.class, error.message, host
        FROM Transaction
        WHERE error IS true AND timestamp > ${15_minutes_before_spike}
        FACET error.class
      discovery: "First error was on host prod-db-01"
      
  phase_4_understand_causality:
    - tool: discovery.find_data_relationships
      question: "What data exists for prod-db-01?"
      params:
        filter: "host = 'prod-db-01'"
      result:
        - SystemSample: CPU and memory metrics
        - DatastoreSample: Database-specific metrics
        - Log: System and application logs
        
    - tool: nrql.execute
      question: "What happened on prod-db-01?"
      query: |
        SELECT *
        FROM SystemSample, DatastoreSample
        WHERE host = 'prod-db-01'
        SINCE 20 minutes ago
        TIMESERIES
      discovery: "Disk usage hit 100% at 15:44:50"
```

### Scenario 3: "Prepare for Black Friday"

#### Discovery-First Capacity Planning
```yaml
capacity_planning:
  phase_1_discover_historical_data:
    - tool: discovery.explore_event_types
      question: "What data do we have from last year?"
      params:
        time_range: "13 months"
      result:
        - Transaction: Full 13 months
        - SystemSample: Only 6 months (retention!)
        - Custom metrics: 3 months
        
    - tool: nrql.execute
      question: "Do we have last Black Friday's data?"
      query: |
        SELECT count(*)
        FROM Transaction
        WHERE timestamp > '2023-11-24 00:00:00' 
          AND timestamp < '2023-11-25 00:00:00'
        TIMESERIES 1 hour
      discovery: "Yes! Full 24-hour data available"
      
  phase_2_understand_patterns:
    - tool: discovery.detect_temporal_patterns
      question: "What patterns exist in our data?"
      params:
        query: "SELECT rate(count(*), 1 minute) FROM Transaction"
        pattern_types: ["daily", "weekly", "special_events"]
        lookback_days: 365
      result:
        - Daily peak: 2-4 PM
        - Weekly peak: Tuesday-Thursday
        - Special events: 10x on Black Friday
        
    - tool: nrql.execute
      question: "What exactly happened last Black Friday?"
      query: |
        SELECT 
          rate(count(*), 1 minute) as 'Requests/min',
          average(duration) as 'Avg Duration',
          percentage(count(*), WHERE error IS true) as 'Error Rate'
        FROM Transaction
        WHERE timestamp > '2023-11-24 00:00:00' 
          AND timestamp < '2023-11-25 00:00:00'
        TIMESERIES 5 minutes
      discovery:
        - Peak rate: 50K req/min at 10 AM
        - Duration increased from 200ms to 800ms
        - Error rate spiked to 5% during peak
        
  phase_3_identify_bottlenecks:
    - tool: discovery.find_natural_groupings
      question: "What struggled last year?"
      params:
        event_type: "Transaction"
        time_filter: "Black Friday 2023"
      result:
        - By 'name': /api/checkout degraded most
        - By 'error.class': Database timeouts dominated
        - By 'host': 3 hosts handled 80% of load
        
    - tool: discovery.profile_data_completeness
      question: "Do we have infrastructure data?"
      params:
        event_type: "SystemSample"
        critical_attributes: ["cpuPercent", "memoryUsedPercent"]
        time_range: "6 months"
      result:
        - Only 6 months retention
        - But includes this year's peak days
        - Can extrapolate from recent peaks
        
  phase_4_validate_assumptions:
    - tool: discovery.validate_assumptions
      question: "Test our capacity assumptions"
      params:
        assumptions:
          - type: "correlation"
            desc: "CPU scales with request rate"
            source: "SELECT rate(count(*), 1 minute) FROM Transaction"
            target: "SELECT average(cpuPercent) FROM SystemSample"
          - type: "threshold"
            desc: "Errors increase after 80% CPU"
            query: "Test correlation between CPU and error rate"
      result:
        - CPU correlation: 0.89 (strong)
        - Error threshold: Actually 75% CPU
        - Memory not correlated (caching?)
```

### Scenario 4: "Define SLOs Without Assumptions"

#### Discovery-First SLO Definition
```yaml
slo_definition:
  phase_1_discover_user_experience:
    - tool: discovery.explore_event_types
      question: "What user-facing data exists?"
      result:
        - Transaction: Backend API calls
        - PageView: Real user browser data
        - SyntheticCheck: Monitoring data
        - MobileRequest: Mobile app data
        
    - tool: discovery.explore_attributes
      question: "What defines 'success' in our data?"
      params:
        event_type: "Transaction"
      result:
        - error: boolean (95% coverage)
        - httpResponseCode: numeric (88% coverage)
        - duration: numeric (99% coverage)
        - custom.success: boolean (12% coverage - sparse!)
        
  phase_2_understand_current_state:
    - tool: nrql.execute
      question: "What's our current success distribution?"
      query: |
        SELECT count(*)
        FROM Transaction
        FACET CASES(
          WHERE error IS true as 'Error',
          WHERE httpResponseCode >= 500 as 'Server Error',
          WHERE httpResponseCode >= 400 as 'Client Error',
          WHERE duration > 1000 as 'Slow Success',
          WHERE duration > 500 as 'Acceptable',
          WHERE duration <= 500 as 'Fast'
        )
        SINCE 7 days ago
      discovery:
        - 2% hard errors
        - 3% slow successes (>1s)
        - 15% acceptable (500ms-1s)
        - 80% fast (<500ms)
        
    - tool: discovery.detect_temporal_patterns
      question: "How stable is performance?"
      params:
        query: "SELECT percentile(duration, 95) FROM Transaction"
        pattern_types: ["hourly", "daily"]
      result:
        - Hourly variation: ±20%
        - Daily variation: ±50%
        - Weekend different from weekday
        
  phase_3_discover_user_impact:
    - tool: discovery.find_data_relationships
      question: "Can we connect to business metrics?"
      params:
        source_event_type: "Transaction"
        target_event_types: ["PageView", "CustomMetric"]
      result:
        - PageView has session.id
        - Some Transactions have session.id
        - Can join 60% of data
        
    - tool: nrql.execute
      question: "How does performance affect users?"
      query: |
        SELECT 
          average(duration) as 'Page Load',
          percentage(count(*), WHERE duration < 1000) as 'Fast Page %',
          uniqueCount(session.id) as 'Unique Sessions'
        FROM PageView
        FACET CASES(
          WHERE duration < 1000 as 'Fast',
          WHERE duration < 3000 as 'Acceptable', 
          WHERE duration >= 3000 as 'Slow'
        )
        SINCE 24 hours ago
      discovery:
        - Slow pages have 50% fewer sessions
        - Clear user impact above 3s
        
  phase_4_data_driven_slo:
    - conclusion: |
        Based on discovered data:
        - SLI: percentage(count(*), WHERE error IS false AND duration < 1000) FROM Transaction
        - Current performance: 93%
        - Natural clusters at 500ms, 1s, 3s
        - User impact threshold: 3s
        - Recommended SLO: 95% (stretch from current 93%)
```

## Key Principles Demonstrated

### 1. Never Trust, Always Verify
- Check event types exist
- Verify attributes are populated  
- Confirm data quality
- Test relationships

### 2. Let Data Tell Its Story
- Use histograms to see distributions
- Find natural groupings
- Discover patterns, don't impose them
- Follow the anomalies

### 3. Build Understanding Incrementally
- Start with what exists
- Understand structure
- Find relationships
- Draw conclusions

### 4. Question Every Assumption
- "Transaction" might not be the right event type
- "error" might mean different things
- Missing data tells a story too
- Correlations might be coincidental

### 5. Use Discovery to Guide Next Steps
- If data is sparse, investigate collection
- If patterns exist, understand their cause
- If relationships are found, validate them
- If assumptions fail, revise approach

## Discovery-First Tool Patterns

```yaml
pattern_1_explore_before_query:
  wrong:
    - tool: nrql.execute
      query: "SELECT average(customMetric) FROM CustomEvent"
      # Fails if CustomEvent doesn't exist
      
  right:
    - tool: discovery.explore_event_types
    - tool: discovery.explore_attributes
    - tool: nrql.execute
      # Now we know it exists

pattern_2_understand_before_alert:
  wrong:
    - tool: alert.create
      condition: "average(cpu) > 80"
      # What if CPU isn't collected?
      
  right:
    - tool: discovery.profile_attribute_values
      attribute: "cpuPercent"
    - tool: discovery.detect_temporal_patterns
    - tool: alert.create
      # Based on actual patterns

pattern_3_validate_before_assume:
  wrong:
    - assumption: "Errors correlate with load"
    - action: "Scale up during high load"
    
  right:
    - tool: discovery.find_data_relationships
    - tool: discovery.validate_assumptions
    - evidence: "Actually memory, not CPU"
    - action: "Optimize memory usage"
```

This discovery-first approach ensures we work with reality, not assumptions, leading to accurate investigations and effective solutions.
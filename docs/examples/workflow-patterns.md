# Workflow Patterns Guide

This guide demonstrates how to use the granular MCP tools to implement complex observability workflows through AI orchestration.

## Table of Contents

1. [Core Workflow Patterns](#core-workflow-patterns)
2. [Investigation Workflows](#investigation-workflows)
3. [Incident Response Workflows](#incident-response-workflows)
4. [Optimization Workflows](#optimization-workflows)
5. [Maintenance Workflows](#maintenance-workflows)
6. [Advanced Patterns](#advanced-patterns)

## Core Workflow Patterns

### 1. Sequential Pattern
Execute steps one after another, with each step potentially using outputs from previous steps.

```yaml
workflow: sequential_investigation
pattern: sequential
steps:
  - id: find_entity
    tool: entity.search_by_name
    inputs:
      name: "${input.service_name}"
      domain: "APM"
    
  - id: get_metrics
    tool: entity.get_golden_metrics
    inputs:
      guid: "${find_entity.output.entities[0].guid}"
      time_range: "1 hour"
    
  - id: detect_anomalies
    tool: pattern.detect
    inputs:
      query: "${get_metrics.output.latency_query}"
      pattern_types: ["anomaly", "changepoint"]
```

### 2. Parallel Pattern
Execute multiple independent operations simultaneously for faster results.

```yaml
workflow: parallel_data_gathering
pattern: parallel
max_concurrent: 5
steps:
  - id: get_app_metrics
    tool: nrql.execute
    inputs:
      query: "SELECT average(duration) FROM Transaction"
      
  - id: get_infra_metrics
    tool: nrql.execute
    inputs:
      query: "SELECT average(cpuPercent) FROM SystemSample"
      
  - id: get_logs
    tool: logs.search_errors
    inputs:
      time_range: "1 hour"
      
  - id: get_traces
    tool: traces.find_slow
    inputs:
      threshold_ms: 1000
```

### 3. Conditional Pattern
Make decisions based on data and execute different paths.

```yaml
workflow: conditional_response
pattern: conditional
condition:
  left: "${check_severity.output.level}"
  operator: "equals"
  right: "critical"
  
true_branch:
  - id: page_oncall
    tool: notification.send_page
    inputs:
      team: "platform-oncall"
      
  - id: create_incident
    tool: incident.create
    inputs:
      priority: "P1"
      
false_branch:
  - id: create_ticket
    tool: ticket.create
    inputs:
      priority: "P3"
```

### 4. Loop Pattern
Iterate over collections or repeat until condition is met.

```yaml
workflow: check_all_dependencies
pattern: loop
items: "${get_dependencies.output.entities}"
loop_body:
  - id: check_health
    tool: entity.get_golden_metrics
    inputs:
      guid: "${item.guid}"
      
  - id: add_to_report
    tool: context.add_finding
    inputs:
      finding:
        type: "dependency_health"
        entity: "${item.name}"
        status: "${check_health.output.status}"
```

## Investigation Workflows

### Performance Investigation Workflow

**Scenario**: Application response time has degraded

```yaml
workflow: performance_investigation
description: Systematic investigation of performance degradation

phase_1_identify_scope:
  - tool: investigate.identify_scope
    inputs:
      symptoms: 
        - "p95 latency increased 200%"
        - "some timeout errors"
      entity_hints: ["checkout-api", "payment-service"]
      time_range: "last 2 hours"
    outputs:
      primary_entities: [...entities to investigate]
      time_windows: {...incident timing}

phase_2_baseline_analysis:
  parallel:
    - tool: nrql.execute
      name: get_baseline_metrics
      inputs:
        query: |
          SELECT percentile(duration, 95) as 'p95'
          FROM Transaction 
          WHERE appName = '${entity.name}'
          COMPARE WITH 1 week ago
          
    - tool: pattern.detect
      name: detect_anomalies
      inputs:
        query: "SELECT average(duration) FROM Transaction WHERE appName = '${entity.name}' SINCE 6 hours ago"
        pattern_types: ["anomaly", "changepoint"]
        sensitivity: 0.8

phase_3_deep_dive:
  - tool: investigate.find_correlations
    inputs:
      primary_signal:
        query: "SELECT average(duration) FROM Transaction WHERE appName = '${entity.name}'"
      candidate_signals:
        - query: "SELECT average(databaseDuration) FROM Transaction WHERE appName = '${entity.name}'"
        - query: "SELECT average(cpuPercent) FROM SystemSample WHERE hostname LIKE '${entity.name}%'"
        - query: "SELECT rate(count(*), 1 minute) FROM Transaction WHERE appName = '${entity.name}'"
      min_correlation: 0.7
      
  - tool: logs.search_errors
    condition: "${detect_anomalies.output.anomaly_detected}"
    inputs:
      entity_guid: "${entity.guid}"
      time_range: "${anomaly.time_window}"
      severity: ["ERROR", "FATAL"]
      
  - tool: traces.find_slow
    inputs:
      entity_guid: "${entity.guid}"
      threshold_percentile: 95
      compare_to_baseline: true

phase_4_root_cause:
  - tool: deployment.find_recent
    inputs:
      entity_guid: "${entity.guid}"
      time_range: "6 hours before incident"
      
  - tool: pattern.explain
    inputs:
      pattern_id: "${detect_anomalies.output.patterns[0].id}"
      include_contributing_factors: true
      
  - tool: analysis.determine_root_cause
    inputs:
      findings: "${context.findings}"
      hypothesis_limit: 3

phase_5_recommendations:
  - tool: context.get_recommendations
    inputs:
      workflow_id: "${workflow.id}"
      recommendation_type: "mitigation"
      
  - tool: alert.suggest_thresholds
    inputs:
      entity_guid: "${entity.guid}"
      metrics: ["duration", "error_rate"]
      based_on: "historical_patterns"
```

### Error Spike Investigation

```yaml
workflow: error_investigation
description: Investigate sudden increase in errors

steps:
  1_quantify_impact:
    - tool: nrql.execute
      inputs:
        query: |
          SELECT count(*) as 'Total Errors',
                 percentage(count(*), WHERE error IS true) as 'Error Rate',
                 uniqueCount(error.class) as 'Unique Error Types'
          FROM Transaction 
          WHERE appName = '${service_name}'
          SINCE 1 hour ago
          COMPARE WITH 1 hour ago
          
  2_categorize_errors:
    - tool: nrql.execute
      inputs:
        query: |
          SELECT count(*) 
          FROM Transaction 
          WHERE appName = '${service_name}' AND error IS true
          FACET error.class, error.message
          SINCE 1 hour ago
          LIMIT 20
          
  3_trace_examples:
    - tool: traces.find_by_error
      inputs:
        error_class: "${top_error.class}"
        limit: 5
        include_logs: true
        
  4_check_dependencies:
    - tool: entity.get_relationships
      inputs:
        guid: "${entity.guid}"
        relationship_types: ["CALLS", "CONSUMES"]
        
    - parallel_for_each: "${relationships.entities}"
      steps:
        - tool: entity.get_golden_metrics
          inputs:
            guid: "${entity.guid}"
            check_anomalies: true
```

## Incident Response Workflows

### Critical Incident Response

```yaml
workflow: critical_incident_response
description: Respond to P1 incidents with automated investigation and mitigation

phase_1_immediate_response:
  parallel:
    - tool: incident.get_context
      outputs:
        alert_condition: {...}
        affected_entity: {...}
        violation_details: {...}
        
    - tool: notification.send_initial
      inputs:
        channels: ["slack-incidents", "pagerduty"]
        template: "incident_started"
        
    - tool: runbook.get_by_alert
      inputs:
        alert_policy_id: "${alert.policy_id}"
        
phase_2_impact_assessment:
  - tool: impact.assess_user_impact
    inputs:
      entity_guid: "${affected_entity.guid}"
      metrics: ["error_rate", "response_time", "throughput"]
      
  - tool: slo.check_breach
    inputs:
      entity_guid: "${affected_entity.guid}"
      time_window: "since_incident_start"
      
  - tool: business.estimate_revenue_impact
    inputs:
      affected_services: "${impact.affected_services}"
      degradation_level: "${impact.severity}"
      
phase_3_diagnosis:
  parallel:
    - tool: traces.find_problematic
      inputs:
        entity_guid: "${affected_entity.guid}"
        time_range: "5 minutes before alert"
        anomaly_types: ["slow", "error", "unusual_path"]
        
    - tool: logs.analyze_patterns
      inputs:
        entity_guid: "${affected_entity.guid}"
        time_range: "around_incident"
        pattern_detection: true
        
    - tool: infrastructure.check_resources
      inputs:
        entity_guid: "${affected_entity.guid}"
        metrics: ["cpu", "memory", "disk", "network"]
        
phase_4_mitigation:
  - tool: decision.analyze_options
    inputs:
      context_id: "${workflow.context_id}"
      options:
        - rollback_deployment
        - scale_up_instances
        - enable_circuit_breaker
        - redirect_traffic
        
  - conditional:
      condition: "${decision.recommended_action == 'scale_up'}"
      true:
        - tool: infrastructure.scale_service
          inputs:
            entity_guid: "${affected_entity.guid}"
            scale_factor: 2
            dry_run: true
            
      false:
        - tool: deployment.rollback
          inputs:
            entity_guid: "${affected_entity.guid}"
            to_version: "${last_known_good}"
            dry_run: true
            
phase_5_verify_resolution:
  - tool: slo.check_recovery
    inputs:
      entity_guid: "${affected_entity.guid}"
      expected_metrics: "${baseline_metrics}"
      
  - tool: incident.update_status
    inputs:
      incident_id: "${incident.id}"
      status: "resolved"
      resolution_notes: "${mitigation.summary}"
```

## Optimization Workflows

### Cost Optimization Workflow

```yaml
workflow: cost_optimization
description: Reduce observability costs while maintaining coverage

phase_1_usage_analysis:
  - tool: metrics.analyze_ingestion
    inputs:
      group_by: ["source", "namespace", "service"]
      time_range: "last 30 days"
      include_trends: true
      
  - tool: queries.analyze_usage
    inputs:
      group_by: ["user", "dashboard", "frequency"]
      identify_unused: true
      
  - tool: retention.analyze_data_age
    inputs:
      group_by: ["event_type", "namespace"]
      show_access_patterns: true
      
phase_2_identify_opportunities:
  - tool: optimization.find_redundant_metrics
    inputs:
      similarity_threshold: 0.95
      
  - tool: optimization.suggest_sampling
    inputs:
      target_reduction: 0.3
      maintain_slo_accuracy: true
      
  - tool: optimization.recommend_aggregations
    inputs:
      common_queries: "${queries.top_100}"
      pre_compute_threshold: "5 minutes"
      
phase_3_impact_simulation:
  - tool: whatif.simulate_sampling
    inputs:
      sampling_rules: "${optimization.suggested_sampling}"
      test_queries: "${critical_dashboards.queries}"
      
  - tool: whatif.simulate_retention
    inputs:
      retention_policy: "${optimization.suggested_retention}"
      check_compliance: true
      
phase_4_implementation:
  - tool: rules.create_drop_rules
    inputs:
      rules: "${approved_optimizations.drop_rules}"
      dry_run: true
      
  - tool: metrics.enable_aggregation
    inputs:
      aggregation_rules: "${approved_optimizations.aggregations}"
      dry_run: true
      
phase_5_validation:
  - tool: coverage.verify_no_gaps
    inputs:
      critical_metrics: "${slo.required_metrics}"
      time_range: "after implementation"
      
  - tool: cost.calculate_savings
    inputs:
      before_state: "${phase_1.baseline}"
      after_state: "current"
```

### Capacity Planning Workflow

```yaml
workflow: capacity_planning
description: Plan for Black Friday traffic surge

phase_1_historical_analysis:
  - tool: pattern.analyze_historical
    inputs:
      metrics:
        - "SELECT rate(count(*), 1 minute) FROM Transaction"
        - "SELECT average(duration) FROM Transaction"
        - "SELECT average(cpuPercent) FROM SystemSample"
      time_periods:
        - "black_friday_2023"
        - "black_friday_2022"
        - "cyber_monday_2023"
        
  - tool: pattern.detect
    inputs:
      pattern_types: ["seasonality", "trend", "peaks"]
      
phase_2_growth_projection:
  - tool: analysis.calculate_growth_rate
    inputs:
      metric: "transaction_rate"
      method: "linear_regression"
      confidence_interval: 0.95
      
  - tool: pattern.forecast
    inputs:
      target_date: "2024-11-29"
      include_seasonality: true
      include_special_events: true
      scenarios: ["conservative", "expected", "aggressive"]
      
phase_3_bottleneck_analysis:
  - tool: analysis.find_bottlenecks
    inputs:
      load_scenarios: "${forecast.scenarios}"
      check_components:
        - database_connections
        - api_rate_limits
        - memory_usage
        - network_bandwidth
        
  - tool: dependency.analyze_scaling_limits
    inputs:
      include_third_party: true
      
phase_4_recommendations:
  - tool: capacity.generate_scaling_plan
    inputs:
      target_load: "${forecast.expected}"
      safety_margin: 0.3
      constraints:
        - budget: "$50000"
        - lead_time: "30 days"
        
  - tool: alert.generate_capacity_thresholds
    inputs:
      scaling_triggers: "${scaling_plan.triggers}"
      advance_warning: "2 hours"
```

## Maintenance Workflows

### SLO Management Workflow

```yaml
workflow: slo_management
description: Define, implement, and monitor SLOs

phase_1_slo_definition:
  - tool: entity.get_golden_metrics
    inputs:
      guid: "${service.guid}"
      
  - tool: sli.analyze_baseline
    inputs:
      entity_guid: "${service.guid}"
      metrics: ["availability", "latency", "quality"]
      percentiles: [90, 95, 99, 99.9]
      time_range: "last 30 days"
      
  - tool: slo.calculate_targets
    inputs:
      baseline: "${sli.baseline}"
      business_requirements: "${input.requirements}"
      
phase_2_implementation:
  - tool: sli.generate_query
    inputs:
      sli_type: "${slo.type}"
      good_events: "${slo.good_definition}"
      total_events: "${slo.total_definition}"
      
  - tool: nrql.validate
    inputs:
      query: "${sli.query}"
      check_performance: true
      
  - tool: alert.create_slo_burn_rate
    inputs:
      slo_target: "${slo.target}"
      window_sizes: ["1h", "6h", "24h"]
      
  - tool: dashboard.create_slo_dashboard
    inputs:
      slo_config: "${slo.config}"
      include_projections: true
      
phase_3_monitoring:
  scheduled: "every 1 hour"
  steps:
    - tool: slo.get_current_performance
      inputs:
        slo_id: "${slo.id}"
        
    - tool: slo.calculate_burn_rate
      inputs:
        current_performance: "${performance}"
        time_windows: ["1h", "24h", "7d"]
        
    - conditional:
        condition: "${burn_rate.1h > 14.4}"
        true:
          - tool: alert.trigger_slo_breach
          - tool: notification.send_slo_alert
```

## Advanced Patterns

### Saga Pattern - Distributed Operations

```yaml
workflow: distributed_deployment
pattern: saga
description: Deploy across multiple regions with rollback capability

transactions:
  - name: validate_deployment
    action:
      tool: deployment.validate_package
      inputs:
        package_id: "${deployment.package_id}"
        target_environments: ["us-east", "eu-west", "ap-south"]
    compensation:
      tool: deployment.mark_invalid
      
  - name: create_backups
    action:
      tool: backup.create_snapshot
      inputs:
        services: "${deployment.affected_services}"
    compensation:
      tool: backup.cleanup_snapshots
      
  - name: deploy_canary
    action:
      tool: deployment.deploy_canary
      inputs:
        region: "us-east"
        traffic_percentage: 1
    compensation:
      tool: deployment.rollback_canary
      
  - name: validate_canary
    action:
      tool: validation.check_canary_metrics
      inputs:
        success_criteria: "${deployment.success_criteria}"
        duration: "10 minutes"
    compensation:
      tool: alert.disable_canary_alerts
      
  - name: gradual_rollout
    action:
      tool: deployment.increase_traffic
      inputs:
        increments: [5, 25, 50, 100]
        wait_between: "5 minutes"
        validation_between: true
    compensation:
      tool: deployment.redirect_all_traffic
      inputs:
        to: "previous_version"
```

### Map-Reduce Pattern - Large Scale Analysis

```yaml
workflow: analyze_all_services
pattern: map_reduce
description: Analyze performance across hundreds of services

map_phase:
  tool: analysis.analyze_service
  inputs:
    service: "${item}"
    checks:
      - latency_trends
      - error_patterns
      - resource_usage
      - dependency_health
      
reduce_phase:
  tool: analysis.aggregate_findings
  inputs:
    aggregations:
      - type: "worst_performers"
        metric: "latency_degradation"
        limit: 10
      - type: "common_errors"
        across: "all_services"
      - type: "resource_bottlenecks"
        threshold: "80%"
```

### Event-Driven Pattern

```yaml
workflow: auto_remediation
pattern: event_driven
triggers:
  - event: "alert.triggered"
    condition: "${alert.policy.tags contains 'auto-remediate'}"
    
steps:
  - tool: remediation.identify_action
    inputs:
      alert_type: "${event.alert.condition_type}"
      entity: "${event.entity}"
      
  - tool: remediation.check_safety
    inputs:
      action: "${identified_action}"
      recent_changes: "${entity.recent_changes}"
      
  - conditional:
      condition: "${safety_check.passed}"
      true:
        - tool: remediation.execute
          inputs:
            action: "${identified_action}"
            monitor_duration: "5 minutes"
      false:
        - tool: escalation.create_ticket
          inputs:
            priority: "high"
            team: "${entity.owner_team}"
```

## Best Practices

### 1. Workflow Design
- Start with simple sequential flows, add complexity as needed
- Use parallel execution for independent operations
- Implement proper error handling at each step
- Add checkpoints for long-running workflows

### 2. Context Management
- Store intermediate results in workflow context
- Use meaningful keys for context variables
- Clean up large data from context when no longer needed
- Document expected context structure

### 3. Error Handling
- Define clear compensation actions for critical steps
- Use conditional logic to handle partial failures
- Implement retry logic with exponential backoff
- Log all errors with sufficient context

### 4. Performance
- Set appropriate timeouts for each step
- Use caching for repeated queries
- Implement pagination for large result sets
- Monitor workflow execution time

### 5. Testing
- Test each workflow step in isolation
- Simulate failures to test compensation logic
- Use dry-run mode for destructive operations
- Validate workflow outputs against expected results

## Workflow Composition

Workflows can be composed from smaller, reusable sub-workflows:

```yaml
workflow: complete_investigation
composition:
  - subworkflow: identify_problem_scope
    outputs: ["entities", "time_range"]
    
  - subworkflow: gather_diagnostic_data
    inputs:
      entities: "${identify_problem_scope.entities}"
      
  - subworkflow: analyze_patterns
    inputs:
      data: "${gather_diagnostic_data.results}"
      
  - subworkflow: generate_recommendations
    inputs:
      findings: "${analyze_patterns.findings}"
```

This modular approach promotes reusability and makes complex workflows more manageable.

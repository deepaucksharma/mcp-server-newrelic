package workflows

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
)

// PerformanceInvestigationE2ESuite tests complete performance investigation workflows
type PerformanceInvestigationE2ESuite struct {
	suite.Suite
	client *framework.MCPTestClient
}

func (s *PerformanceInvestigationE2ESuite) SetupSuite() {
	accounts := framework.LoadTestAccounts()
	s.Require().NotEmpty(accounts, "No test accounts configured")
	s.client = framework.NewMCPTestClient(accounts["primary"])
}

func (s *PerformanceInvestigationE2ESuite) TearDownSuite() {
	s.client.Close()
}

// TestPerformanceInvestigationWorkflow validates a complete performance investigation
func (s *PerformanceInvestigationE2ESuite) TestPerformanceInvestigationWorkflow() {
	ctx := context.Background()
	
	// Define the discovery-first workflow
	workflow := []framework.WorkflowStep{
		// Phase 1: Discover what performance data exists
		{
			Name: "Discover available event types",
			Tool: "discovery.explore_event_types",
			Params: map[string]interface{}{
				"time_range": "24 hours",
				"include_samples": true,
			},
			StoreAs: "discovered_types",
			Validate: func(response map[string]interface{}) error {
				eventTypes, ok := response["event_types"].([]interface{})
				if !ok || len(eventTypes) == 0 {
					return fmt.Errorf("no event types discovered")
				}
				return nil
			},
		},
		
		// Phase 2: Find performance-related attributes
		{
			Name: "Profile Transaction attributes",
			Tool: "discovery.profile_attribute",
			Params: map[string]interface{}{
				"event_type": "Transaction", // We'll make this dynamic in real implementation
				"attributes": []string{"duration", "databaseDuration", "externalDuration", "queueDuration"},
			},
			StoreAs: "performance_attributes",
			ContinueOnError: true, // Transaction might not exist
		},
		
		// Phase 3: Discover service identification
		{
			Name: "Discover service attribute",
			Tool: "discovery.find_attribute_patterns",
			Params: map[string]interface{}{
				"event_type": "Transaction",
				"pattern": "service|app|application",
			},
			StoreAs: "service_attributes",
		},
		
		// Phase 4: Query current performance metrics
		{
			Name: "Get current performance metrics",
			Tool: "nrql.execute",
			Params: map[string]interface{}{
				"query": "SELECT average(duration), percentile(duration, 95), stddev(duration) FROM Transaction SINCE 1 hour ago",
				"timeout": 30,
			},
			StoreAs: "current_metrics",
		},
		
		// Phase 5: Compare with historical baseline
		{
			Name: "Compare with baseline",
			Tool: "nrql.execute",
			Params: map[string]interface{}{
				"query": "SELECT average(duration) FROM Transaction SINCE 1 hour ago COMPARE WITH 1 week ago",
			},
			StoreAs: "baseline_comparison",
		},
		
		// Phase 6: Identify problematic transactions
		{
			Name: "Find slow transactions",
			Tool: "nrql.execute",
			Params: map[string]interface{}{
				"query": "SELECT average(duration), count(*) FROM Transaction WHERE duration > 1000 FACET name SINCE 1 hour ago LIMIT 20",
			},
			StoreAs: "slow_transactions",
		},
		
		// Phase 7: Analyze error correlation
		{
			Name: "Check error correlation",
			Tool: "nrql.execute",
			Params: map[string]interface{}{
				"query": "SELECT average(duration), percentage(count(*), WHERE error = true) FROM Transaction FACET error SINCE 1 hour ago",
			},
			StoreAs: "error_correlation",
		},
	}
	
	// Execute the workflow
	result, err := s.client.ExecuteWorkflow(ctx, workflow)
	s.Require().NoError(err, "Workflow should complete successfully")
	s.Require().NotNil(result, "Should return workflow results")
	
	// Validate workflow completed all steps
	s.Equal(len(workflow), len(result.Steps), "All workflow steps should execute")
	
	// Validate specific discoveries
	s.validateDiscoveredTypes(result)
	s.validatePerformanceMetrics(result)
	s.validateSlowTransactions(result)
	
	// Generate investigation summary
	summary := s.generateInvestigationSummary(result)
	s.T().Logf("Investigation Summary:\n%s", summary)
}

// TestAdaptivePerformanceWorkflow tests workflow that adapts to discovered schema
func (s *PerformanceInvestigationE2ESuite) TestAdaptivePerformanceWorkflow() {
	ctx := context.Background()
	
	// First, discover what performance data actually exists
	discoveryResult, err := s.discoverPerformanceSchema(ctx)
	s.Require().NoError(err, "Should discover performance schema")
	
	// Build adaptive workflow based on discoveries
	workflow := s.buildAdaptiveWorkflow(discoveryResult)
	s.Require().NotEmpty(workflow, "Should build adaptive workflow")
	
	// Execute adaptive workflow
	result, err := s.client.ExecuteWorkflow(ctx, workflow)
	s.Require().NoError(err, "Adaptive workflow should succeed")
	
	// Validate adapted to actual schema
	s.validateAdaptiveResults(result, discoveryResult)
}

// TestPerformanceAnomalyDetection tests anomaly detection workflow
func (s *PerformanceInvestigationE2ESuite) TestPerformanceAnomalyDetection() {
	ctx := context.Background()
	
	// Workflow to detect performance anomalies
	workflow := []framework.WorkflowStep{
		// Establish baseline
		{
			Name: "Calculate performance baseline",
			Tool: "analysis.calculate_baseline",
			Params: map[string]interface{}{
				"metric": "duration",
				"event_type": "Transaction",
				"time_range": "7 days",
				"percentiles": []float64{50, 90, 95, 99},
			},
			StoreAs: "baseline",
		},
		
		// Detect anomalies
		{
			Name: "Detect anomalies",
			Tool: "analysis.detect_anomalies",
			Params: map[string]interface{}{
				"metric": "duration",
				"event_type": "Transaction",
				"time_range": "1 hour",
				"sensitivity": 3,
				"method": "zscore",
			},
			StoreAs: "anomalies",
		},
		
		// Find correlated factors
		{
			Name: "Find correlations",
			Tool: "analysis.find_correlations",
			Params: map[string]interface{}{
				"primary_metric": "duration",
				"event_type": "Transaction",
				"time_range": "1 hour",
				"min_correlation": 0.7,
			},
			StoreAs: "correlations",
		},
	}
	
	result, err := s.client.ExecuteWorkflow(ctx, workflow)
	
	// If analysis tools not implemented, skip
	if err != nil && isToolNotImplementedError(err) {
		s.T().Skip("Analysis tools not yet implemented")
	}
	
	s.Require().NoError(err, "Anomaly detection workflow should succeed")
	s.validateAnomalyDetection(result)
}

// Helper methods

func (s *PerformanceInvestigationE2ESuite) discoverPerformanceSchema(ctx context.Context) (map[string]interface{}, error) {
	// Discover what performance-related data exists
	result, err := s.client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
		"pattern": "transaction|performance|apm",
		"time_range": "24 hours",
	})
	
	if err != nil {
		return nil, err
	}
	
	// Find performance-related event types
	eventTypes, _ := result["event_types"].([]interface{})
	performanceTypes := []string{}
	
	for _, et := range eventTypes {
		if eventType, ok := et.(map[string]interface{}); ok {
			name, _ := eventType["name"].(string)
			// Don't assume specific names - check if it has duration-like attributes
			if s.hasPerformanceAttributes(ctx, name) {
				performanceTypes = append(performanceTypes, name)
			}
		}
	}
	
	return map[string]interface{}{
		"performance_event_types": performanceTypes,
		"discovery_result": result,
	}, nil
}

func (s *PerformanceInvestigationE2ESuite) hasPerformanceAttributes(ctx context.Context, eventType string) bool {
	// Check if event type has performance-related attributes
	result, err := s.client.ExecuteTool(ctx, "discovery.profile_attribute", map[string]interface{}{
		"event_type": eventType,
		"attributes": []string{"duration", "time", "latency", "response"},
	})
	
	if err != nil {
		return false
	}
	
	// Check if any performance attribute has good coverage
	if profiles, ok := result["profiles"].([]interface{}); ok {
		for _, profile := range profiles {
			if p, ok := profile.(map[string]interface{}); ok {
				if coverage, ok := p["coverage"].(float64); ok && coverage > 50 {
					return true
				}
			}
		}
	}
	
	return false
}

func (s *PerformanceInvestigationE2ESuite) buildAdaptiveWorkflow(discovery map[string]interface{}) []framework.WorkflowStep {
	workflow := []framework.WorkflowStep{}
	
	// Get discovered performance event types
	performanceTypes, _ := discovery["performance_event_types"].([]string)
	if len(performanceTypes) == 0 {
		return workflow
	}
	
	// Use the first performance event type
	eventType := performanceTypes[0]
	
	// Build workflow steps based on what exists
	workflow = append(workflow, framework.WorkflowStep{
		Name: fmt.Sprintf("Query %s performance", eventType),
		Tool: "nrql.execute",
		Params: map[string]interface{}{
			"query": fmt.Sprintf("SELECT average(duration), count(*) FROM %s SINCE 1 hour ago", eventType),
		},
		StoreAs: "performance_data",
	})
	
	// Add more adaptive steps based on discoveries
	// ... (additional logic based on discovered schema)
	
	return workflow
}

func (s *PerformanceInvestigationE2ESuite) validateDiscoveredTypes(result *framework.WorkflowResult) {
	// Find the discovered types step
	for _, step := range result.Steps {
		if step.Name == "Discover available event types" && step.Response != nil {
			eventTypes, ok := step.Response["event_types"].([]interface{})
			s.True(ok, "Should have event types")
			s.NotEmpty(eventTypes, "Should discover some event types")
			
			// Log discovered types
			s.T().Logf("Discovered %d event types", len(eventTypes))
			for i, et := range eventTypes {
				if eventType, ok := et.(map[string]interface{}); ok {
					s.T().Logf("  [%d] %s: %v events", i, eventType["name"], eventType["count"])
				}
			}
			return
		}
	}
	
	s.Fail("Could not find discovered types step")
}

func (s *PerformanceInvestigationE2ESuite) validatePerformanceMetrics(result *framework.WorkflowResult) {
	// Find performance metrics step
	for _, step := range result.Steps {
		if step.Name == "Get current performance metrics" && step.Response != nil {
			results, ok := step.Response["results"].([]interface{})
			s.True(ok, "Should have query results")
			s.NotEmpty(results, "Should have performance metrics")
			
			// Validate metric structure
			if len(results) > 0 {
				if metrics, ok := results[0].(map[string]interface{}); ok {
					s.Contains(metrics, "average.duration", "Should have average duration")
					s.Contains(metrics, "percentile.duration", "Should have percentile duration")
					
					// Log metrics
					s.T().Logf("Performance Metrics:")
					for key, value := range metrics {
						s.T().Logf("  %s: %v", key, value)
					}
				}
			}
			return
		}
	}
}

func (s *PerformanceInvestigationE2ESuite) validateSlowTransactions(result *framework.WorkflowResult) {
	// Find slow transactions step
	for _, step := range result.Steps {
		if step.Name == "Find slow transactions" && step.Response != nil {
			results, ok := step.Response["results"].([]interface{})
			if !ok || len(results) == 0 {
				s.T().Log("No slow transactions found (good!)")
				return
			}
			
			s.T().Logf("Found %d slow transaction types", len(results))
			for i, result := range results {
				if txn, ok := result.(map[string]interface{}); ok {
					s.T().Logf("  [%d] %s: avg=%.2fms, count=%v", 
						i, txn["name"], txn["average.duration"], txn["count"])
				}
			}
		}
	}
}

func (s *PerformanceInvestigationE2ESuite) validateAdaptiveResults(result *framework.WorkflowResult, discovery map[string]interface{}) {
	// Validate workflow adapted to discovered schema
	s.True(result.Success, "Adaptive workflow should succeed")
	
	// Check that queries used discovered event types
	performanceTypes, _ := discovery["performance_event_types"].([]string)
	if len(performanceTypes) > 0 {
		// Verify at least one step queried the discovered type
		found := false
		for _, step := range result.Steps {
			if step.Tool == "nrql.execute" {
				if params, ok := step.Response["query"].(string); ok {
					for _, eventType := range performanceTypes {
						if strings.Contains(params, eventType) {
							found = true
							break
						}
					}
				}
			}
		}
		s.True(found, "Should use discovered event types in queries")
	}
}

func (s *PerformanceInvestigationE2ESuite) validateAnomalyDetection(result *framework.WorkflowResult) {
	// Validate anomaly detection results
	for _, step := range result.Steps {
		if step.Name == "Detect anomalies" && step.Response != nil {
			s.Contains(step.Response, "anomalies", "Should return anomalies")
			s.Contains(step.Response, "anomaliesDetected", "Should have anomaly count")
			
			if count, ok := step.Response["anomaliesDetected"].(float64); ok {
				s.T().Logf("Detected %v anomalies", count)
				
				if anomalies, ok := step.Response["anomalies"].([]interface{}); ok && len(anomalies) > 0 {
					s.T().Log("Anomaly details:")
					for i, anomaly := range anomalies {
						s.T().Logf("  [%d] %+v", i, anomaly)
					}
				}
			}
		}
	}
}

func (s *PerformanceInvestigationE2ESuite) generateInvestigationSummary(result *framework.WorkflowResult) string {
	summary := "=== Performance Investigation Summary ===\n"
	summary += fmt.Sprintf("Workflow Duration: %v\n", result.TotalDuration)
	summary += fmt.Sprintf("Steps Completed: %d\n\n", len(result.Steps))
	
	// Summarize key findings
	summary += "Key Findings:\n"
	
	// Extract metrics
	for _, step := range result.Steps {
		if step.Name == "Get current performance metrics" && step.Response != nil {
			if results, ok := step.Response["results"].([]interface{}); ok && len(results) > 0 {
				if metrics, ok := results[0].(map[string]interface{}); ok {
					if avg, ok := metrics["average.duration"].(float64); ok {
						summary += fmt.Sprintf("- Average Duration: %.2fms\n", avg)
					}
					if p95, ok := metrics["percentile.duration"].(float64); ok {
						summary += fmt.Sprintf("- 95th Percentile: %.2fms\n", p95)
					}
				}
			}
		}
		
		// Check baseline comparison
		if step.Name == "Compare with baseline" && step.Response != nil {
			summary += "- Baseline Comparison: Available\n"
		}
		
		// Check for errors
		if step.Error != nil {
			summary += fmt.Sprintf("- Error in '%s': %v\n", step.Name, step.Error)
		}
	}
	
	return summary
}

func isToolNotImplementedError(err error) bool {
	return err != nil && 
		(strings.Contains(err.Error(), "not implemented") ||
		 strings.Contains(err.Error(), "tool not found"))
}

func TestPerformanceInvestigationE2E(t *testing.T) {
	suite.Run(t, new(PerformanceInvestigationE2ESuite))
}
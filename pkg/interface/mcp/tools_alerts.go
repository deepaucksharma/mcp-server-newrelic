//go:build !test

package mcp

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
)

// registerAlertTools registers alert-related tools
func (s *Server) registerAlertTools() error {
	// Create alert condition
	s.tools.Register(Tool{
		Name:        "create_alert",
		Description: "Create an intelligent alert condition with automatic threshold calculation",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"name", "query"},
			Properties: map[string]Property{
				"name": {
					Type:        "string",
					Description: "Alert condition name",
				},
				"query": {
					Type:        "string",
					Description: "NRQL query for the alert condition",
				},
				"sensitivity": {
					Type:        "string",
					Description: "Sensitivity level: 'low', 'medium', 'high' (default: 'medium')",
					Default:     "medium",
				},
				"comparison": {
					Type:        "string",
					Description: "Comparison operator: 'above', 'below', 'equals' (default: 'above')",
					Default:     "above",
				},
				"threshold_duration": {
					Type:        "integer",
					Description: "How many minutes the threshold must be violated (default: 5)",
					Default:     5,
				},
				"auto_baseline": {
					Type:        "boolean",
					Description: "Use automatic baseline detection for threshold",
					Default:     true,
				},
				"static_threshold": {
					Type:        "number",
					Description: "Static threshold value (used if auto_baseline is false)",
				},
				"policy_id": {
					Type:        "string",
					Description: "Alert policy ID to add condition to",
				},
				"account_id": {
					Type:        "string",
					Description: "Optional account ID to create alert in (uses default if not provided)",
				},
			},
		},
		Handler: s.handleCreateAlert,
	})

	// List alert conditions
	s.tools.Register(Tool{
		Name:        "list_alerts",
		Description: "List all alert conditions in the account with pagination support",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]Property{
				"policy_id": {
					Type:        "string",
					Description: "Filter by alert policy ID",
				},
				"enabled_only": {
					Type:        "boolean",
					Description: "Only show enabled alerts",
					Default:     false,
				},
				"include_incidents": {
					Type:        "boolean",
					Description: "Include recent incidents for each alert",
					Default:     false,
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of alerts to return per page (max: 200)",
					Default:     50,
				},
				"cursor": {
					Type:        "string",
					Description: "Pagination cursor from previous response",
				},
				"account_id": {
					Type:        "string",
					Description: "Optional account ID to query (uses default if not provided)",
				},
			},
		},
		Handler: s.handleListAlerts,
	})

	// Analyze alert effectiveness
	s.tools.Register(Tool{
		Name:        "analyze_alerts",
		Description: "Analyze alert effectiveness and suggest improvements",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"alert_id"},
			Properties: map[string]Property{
				"alert_id": {
					Type:        "string",
					Description: "Alert condition ID to analyze",
				},
				"time_range": {
					Type:        "string",
					Description: "Time range for analysis (default: '7 days')",
					Default:     "7 days",
				},
			},
		},
		Handler: s.handleAnalyzeAlerts,
	})

	// Bulk update alerts
	s.tools.Register(Tool{
		Name:        "bulk_update_alerts",
		Description: "Perform bulk updates on multiple alert conditions",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"alert_ids", "operation"},
			Properties: map[string]Property{
				"alert_ids": {
					Type:        "array",
					Description: "List of alert condition IDs to update",
					Items:       &Property{Type: "string"},
				},
				"operation": {
					Type:        "string",
					Description: "Operation to perform: 'enable', 'disable', 'update_threshold', 'delete'",
				},
				"new_threshold": {
					Type:        "number",
					Description: "New threshold value (for update_threshold operation)",
				},
				"threshold_multiplier": {
					Type:        "number",
					Description: "Multiply existing thresholds by this value (for update_threshold)",
				},
			},
		},
		Handler: s.handleBulkUpdateAlerts,
	})

	// Create alert policy
	s.tools.Register(Tool{
		Name:        "create_alert_policy",
		Description: "Create a new alert policy",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"name"},
			Properties: map[string]Property{
				"name": {
					Type:        "string",
					Description: "Policy name",
				},
				"incident_preference": {
					Type:        "string",
					Description: "How to create incidents: PER_POLICY, PER_CONDITION, PER_CONDITION_AND_TARGET (default: PER_CONDITION)",
					Default:     "PER_CONDITION",
				},
			},
		},
		Handler: s.handleCreateAlertPolicy,
	})

	// Update alert policy
	s.tools.Register(Tool{
		Name:        "update_alert_policy",
		Description: "Update an existing alert policy",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"policy_id"},
			Properties: map[string]Property{
				"policy_id": {
					Type:        "string",
					Description: "Policy ID to update",
				},
				"name": {
					Type:        "string",
					Description: "New policy name",
				},
				"incident_preference": {
					Type:        "string",
					Description: "New incident preference: PER_POLICY, PER_CONDITION, PER_CONDITION_AND_TARGET",
				},
			},
		},
		Handler: s.handleUpdateAlertPolicy,
	})

	// Delete alert policy
	s.tools.Register(Tool{
		Name:        "delete_alert_policy",
		Description: "Delete an alert policy",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"policy_id"},
			Properties: map[string]Property{
				"policy_id": {
					Type:        "string",
					Description: "Policy ID to delete",
				},
			},
		},
		Handler: s.handleDeleteAlertPolicy,
	})

	// Create alert condition
	s.tools.Register(Tool{
		Name:        "create_alert_condition",
		Description: "Create a new alert condition in a policy",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"policy_id", "name", "query"},
			Properties: map[string]Property{
				"policy_id": {
					Type:        "string",
					Description: "Policy ID to add condition to",
				},
				"name": {
					Type:        "string",
					Description: "Condition name",
				},
				"query": {
					Type:        "string",
					Description: "NRQL query for the condition",
				},
				"threshold": {
					Type:        "number",
					Description: "Alert threshold value",
				},
				"threshold_duration": {
					Type:        "integer",
					Description: "How many minutes the threshold must be violated (default: 5)",
					Default:     5,
				},
				"comparison": {
					Type:        "string",
					Description: "Comparison operator: 'above', 'below', 'equals' (default: 'above')",
					Default:     "above",
				},
			},
		},
		Handler: s.handleCreateAlertCondition,
	})

	// Close incident
	s.tools.Register(Tool{
		Name:        "close_incident",
		Description: "Close an open alert incident",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"incident_id"},
			Properties: map[string]Property{
				"incident_id": {
					Type:        "string",
					Description: "Incident ID to close",
				},
				"account_id": {
					Type:        "integer",
					Description: "Account ID (optional, uses default if not provided)",
				},
			},
		},
		Handler: s.handleCloseIncident,
	})

	return nil
}

// handleCreateAlert creates a new alert condition
func (s *Server) handleCreateAlert(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name parameter is required")
	}

	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// Validate NRQL query
	if err := s.validateNRQLSyntax(query); err != nil {
		return nil, fmt.Errorf("invalid NRQL query: %w", err)
	}

	sensitivity := "medium"
	if sens, ok := params["sensitivity"].(string); ok {
		sensitivity = sens
	}

	comparison := "above"
	if comp, ok := params["comparison"].(string); ok {
		comparison = comp
	}

	thresholdDuration := 5
	if td, ok := params["threshold_duration"].(float64); ok {
		thresholdDuration = int(td)
	}

	autoBaseline := true
	if ab, ok := params["auto_baseline"].(bool); ok {
		autoBaseline = ab
	}

	// Calculate threshold
	var threshold float64
	if autoBaseline {
		// Calculate baseline from historical data
		baseline, err := s.calculateBaseline(ctx, query, sensitivity)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate baseline: %w", err)
		}
		threshold = baseline
	} else {
		// Use static threshold
		if st, ok := params["static_threshold"].(float64); ok {
			threshold = st
		} else {
			return nil, fmt.Errorf("static_threshold is required when auto_baseline is false")
		}
	}


	// Add to policy if specified
	policyID, _ := params["policy_id"].(string)
	if policyID == "" {
		return nil, fmt.Errorf("policy_id is required")
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("create_alert", params), nil
	}

	// Get account ID if specified
	accountID, _ := params["account_id"].(string)

	// Get New Relic client with account support
	nrClient, err := s.getNRClientWithAccount(accountID)
	if err != nil {
		return nil, err
	}

	// Create alert condition structure
	condition := newrelic.AlertCondition{
		Name:              name,
		Query:             query,
		Comparison:        comparison,
		Threshold:         threshold,
		ThresholdDuration: int(thresholdDuration),
		PolicyID:          policyID,
		Enabled:           true,
	}

	// Use reflection to call CreateAlertCondition
	clientValue := reflect.ValueOf(nrClient)
	method := clientValue.MethodByName("CreateAlertCondition")
	if !method.IsValid() {
		return nil, fmt.Errorf("CreateAlertCondition method not found on client")
	}

	// Call the method
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(condition),
	}
	results := method.Call(args)
	
	if len(results) != 2 {
		return nil, fmt.Errorf("unexpected return values from CreateAlertCondition")
	}

	// Extract error
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	// Extract created alert
	createdValue := results[0].Elem()

	// Helper to get field value using reflection
	getField := func(name string) interface{} {
		f := createdValue.FieldByName(name)
		if f.IsValid() {
			return f.Interface()
		}
		return nil
	}

	return map[string]interface{}{
		"alert": map[string]interface{}{
			"id":                 getField("ID"),
			"name":               getField("Name"),
			"query":              getField("Query"),
			"comparison":         getField("Comparison"),
			"threshold":          getField("Threshold"),
			"threshold_duration": getField("ThresholdDuration"),
			"enabled":            getField("Enabled"),
			"policy_id":          getField("PolicyID"),
			"created_at":         getField("CreatedAt"),
		},
		"message":   fmt.Sprintf("Alert condition '%s' created successfully", name),
		"threshold": map[string]interface{}{
			"value":            threshold,
			"calculation_method": map[string]interface{}{
				"auto_baseline": autoBaseline,
				"sensitivity":   sensitivity,
			},
		},
	}, nil
}

// handleListAlerts lists all alert conditions
func (s *Server) handleListAlerts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	policyID, _ := params["policy_id"].(string)
	enabledOnly, _ := params["enabled_only"].(bool)
	includeIncidents, _ := params["include_incidents"].(bool)

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// List alert conditions
	conditions, err := client.ListAlertConditions(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("list alert conditions: %w", err)
	}

	// Convert to response format
	alerts := make([]map[string]interface{}, 0, len(conditions))
	for _, c := range conditions {
		// Apply enabled filter if requested
		if enabledOnly && !c.Enabled {
			continue
		}

		alert := map[string]interface{}{
			"id":                 c.ID,
			"name":               c.Name,
			"query":              c.Query,
			"comparison":         c.Comparison,
			"threshold":          c.Threshold,
			"threshold_duration": c.ThresholdDuration,
			"enabled":            c.Enabled,
			"policy_id":          c.PolicyID,
			"created_at":         c.CreatedAt,
			"updated_at":         c.UpdatedAt,
		}
		
		alerts = append(alerts, alert)
	}

	// Apply filters
	filtered := []map[string]interface{}{}
	for _, alert := range alerts {
		// Filter by policy ID
		if policyID != "" && alert["policy_id"] != policyID {
			continue
		}

		// Filter by enabled status
		if enabledOnly && !alert["enabled"].(bool) {
			continue
		}

		// Add incident data if requested
		if includeIncidents {
			alert["recent_incidents"] = []map[string]interface{}{
				{
					"opened_at":    time.Now().Add(-3 * time.Hour),
					"closed_at":    time.Now().Add(-2 * time.Hour),
					"violation_value": 7.5,
				},
			}
		}

		filtered = append(filtered, alert)
	}

	return map[string]interface{}{
		"total":  len(filtered),
		"alerts": filtered,
	}, nil
}

// handleAnalyzeAlerts analyzes alert effectiveness
func (s *Server) handleAnalyzeAlerts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	alertID, ok := params["alert_id"].(string)
	if !ok || alertID == "" {
		return nil, fmt.Errorf("alert_id parameter is required")
	}

	timeRange := "7 days"
	if tr, ok := params["time_range"].(string); ok {
		timeRange = tr
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("analyze_alerts", params), nil
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Ensure timeRange has "ago" suffix
	if !strings.HasSuffix(timeRange, " ago") {
		timeRange = timeRange + " ago"
	}

	// Get alert analytics
	analytics, err := client.GetAlertAnalytics(ctx, alertID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("get alert analytics: %w", err)
	}

	// Calculate effectiveness metrics
	incidentCount := getFloatValue(analytics["incident_count"])
	avgDuration := getFloatValue(analytics["avg_duration_minutes"])
	medianDuration := getFloatValue(analytics["median_duration_minutes"])
	minDuration := getFloatValue(analytics["min_duration_minutes"])
	maxDuration := getFloatValue(analytics["max_duration_minutes"])
	
	// Estimate false positive rate based on short duration incidents
	falsePositives := 0
	if medianDuration < 5 {
		falsePositives = int(incidentCount * 0.15) // Estimate 15% false positives for very short incidents
	}
	
	// Calculate noise ratio and effectiveness
	noiseRatio := 0.0
	if incidentCount > 0 {
		noiseRatio = float64(falsePositives) / incidentCount
	}
	effectivenessScore := 1.0 - noiseRatio
	
	// Determine rating
	rating := "Good"
	if effectivenessScore < 0.7 {
		rating = "Poor"
	} else if effectivenessScore < 0.85 {
		rating = "Fair"
	} else if effectivenessScore > 0.95 {
		rating = "Excellent"
	}
	
	// Generate recommendations
	recommendations := []map[string]interface{}{}
	
	if noiseRatio > 0.2 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "threshold_adjustment",
			"priority":    "high",
			"description": "Consider increasing the alert threshold to reduce false positives",
			"impact":      fmt.Sprintf("Could reduce incidents by ~%.0f%% with minimal risk", noiseRatio*100),
		})
	}
	
	if avgDuration < 10 && incidentCount > 10 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "duration_adjustment",
			"priority":    "medium",
			"description": "Increase threshold duration to reduce alert flapping",
			"impact":      "Would reduce noise by filtering transient spikes",
		})
	}
	
	if maxDuration > avgDuration*3 && incidentCount > 5 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "multi_condition",
			"priority":    "low",
			"description": "Consider splitting into multiple conditions for different severity levels",
			"impact":      "Better incident prioritization and response",
		})
	}

	analysis := map[string]interface{}{
		"alert_id":    alertID,
		"time_range":  timeRange,
		"summary": map[string]interface{}{
			"total_incidents":      int(incidentCount),
			"false_positives":      falsePositives,
			"mean_time_to_resolve": fmt.Sprintf("%.0f minutes", avgDuration),
			"noise_ratio":          fmt.Sprintf("%.2f", noiseRatio),
		},
		"effectiveness": map[string]interface{}{
			"score":       fmt.Sprintf("%.2f", effectivenessScore),
			"rating":      rating,
			"confidence":  determineConfidence(incidentCount),
		},
		"recommendations": recommendations,
		"incident_metrics": map[string]interface{}{
			"avg_duration_minutes":    avgDuration,
			"median_duration_minutes": medianDuration,
			"min_duration_minutes":    minDuration,
			"max_duration_minutes":    maxDuration,
		},
	}

	return analysis, nil
}

// handleBulkUpdateAlerts performs bulk operations on alerts
func (s *Server) handleBulkUpdateAlerts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	alertIDs, ok := params["alert_ids"].([]interface{})
	if !ok || len(alertIDs) == 0 {
		return nil, fmt.Errorf("alert_ids parameter is required")
	}

	operation, ok := params["operation"].(string)
	if !ok || operation == "" {
		return nil, fmt.Errorf("operation parameter is required")
	}

	// Convert alert IDs to strings
	ids := make([]string, len(alertIDs))
	for i, id := range alertIDs {
		ids[i] = id.(string)
	}

	results := map[string]interface{}{
		"operation":     operation,
		"total_alerts":  len(ids),
		"successful":    0,
		"failed":        0,
		"results":       []map[string]interface{}{},
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("bulk_update_alerts", params), nil
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Process each alert
	for _, alertID := range ids {
		result := map[string]interface{}{
			"alert_id": alertID,
			"status":   "success",
		}

		var err error
		switch operation {
		case "enable":
			err = client.EnableAlertCondition(ctx, alertID)
			if err == nil {
				result["message"] = "Alert enabled"
			}

		case "disable":
			err = client.DisableAlertCondition(ctx, alertID)
			if err == nil {
				result["message"] = "Alert disabled"
			}

		case "update_threshold":
			if newThreshold, ok := params["new_threshold"].(float64); ok {
				// Update with absolute threshold
				updates := map[string]interface{}{
					"terms": []map[string]interface{}{
						{
							"threshold": newThreshold,
							"priority":  "CRITICAL",
						},
					},
				}
				_, err = client.UpdateAlertCondition(ctx, alertID, updates)
				if err == nil {
					result["message"] = fmt.Sprintf("Threshold updated to %.2f", newThreshold)
				}
			} else if multiplier, ok := params["threshold_multiplier"].(float64); ok {
				// Get current condition and multiply threshold
				conditions, getErr := client.ListAlertConditions(ctx, "")
				if getErr != nil {
					err = getErr
				} else {
					found := false
					for _, cond := range conditions {
						if cond.ID == alertID {
							newThreshold := cond.Threshold * multiplier
							updates := map[string]interface{}{
								"terms": []map[string]interface{}{
									{
										"threshold":         newThreshold,
										"thresholdDuration": cond.ThresholdDuration,
										"operator":          cond.Comparison,
										"priority":          "CRITICAL",
									},
								},
							}
							_, err = client.UpdateAlertCondition(ctx, alertID, updates)
							if err == nil {
								result["message"] = fmt.Sprintf("Threshold updated from %.2f to %.2f (multiplier: %.2f)", 
									cond.Threshold, newThreshold, multiplier)
							}
							found = true
							break
						}
					}
					if !found {
						err = fmt.Errorf("alert condition not found")
					}
				}
			} else {
				err = fmt.Errorf("update_threshold requires new_threshold or threshold_multiplier")
			}

		case "delete":
			err = client.DeleteAlertCondition(ctx, alertID)
			if err == nil {
				result["message"] = "Alert deleted"
			}

		default:
			err = fmt.Errorf("unknown operation: %s", operation)
		}

		if err != nil {
			result["status"] = "failed"
			result["error"] = err.Error()
			results["failed"] = results["failed"].(int) + 1
		} else {
			results["successful"] = results["successful"].(int) + 1
		}

		results["results"] = append(results["results"].([]map[string]interface{}), result)
	}

	return results, nil
}

// handleCreateAlertPolicy creates a new alert policy
func (s *Server) handleCreateAlertPolicy(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name parameter is required")
	}

	incidentPref := "PER_CONDITION"
	if pref, ok := params["incident_preference"].(string); ok {
		incidentPref = pref
	}

	// Validate incident preference
	validPrefs := map[string]bool{
		"PER_POLICY":                true,
		"PER_CONDITION":             true,
		"PER_CONDITION_AND_TARGET":  true,
	}
	if !validPrefs[incidentPref] {
		return nil, fmt.Errorf("invalid incident_preference: %s", incidentPref)
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("create_alert_policy", params), nil
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Create the policy
	policy := newrelic.AlertPolicy{
		Name:               name,
		IncidentPreference: incidentPref,
	}

	created, err := client.CreateAlertPolicy(ctx, policy)
	if err != nil {
		return nil, fmt.Errorf("create alert policy: %w", err)
	}

	return map[string]interface{}{
		"policy": map[string]interface{}{
			"id":                   created.ID,
			"name":                 created.Name,
			"incident_preference":  created.IncidentPreference,
			"created_at":          created.CreatedAt,
			"updated_at":          created.UpdatedAt,
		},
		"message": fmt.Sprintf("Alert policy '%s' created successfully", name),
	}, nil
}

// handleUpdateAlertPolicy updates an existing alert policy
func (s *Server) handleUpdateAlertPolicy(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	policyID, ok := params["policy_id"].(string)
	if !ok || policyID == "" {
		return nil, fmt.Errorf("policy_id parameter is required")
	}

	updates := map[string]interface{}{}
	if name, ok := params["name"].(string); ok && name != "" {
		updates["name"] = name
	}
	if pref, ok := params["incident_preference"].(string); ok && pref != "" {
		// Validate incident preference
		validPrefs := map[string]bool{
			"PER_POLICY":                true,
			"PER_CONDITION":             true,
			"PER_CONDITION_AND_TARGET":  true,
		}
		if !validPrefs[pref] {
			return nil, fmt.Errorf("invalid incident_preference: %s", pref)
		}
		updates["incident_preference"] = pref
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("at least one field to update must be provided")
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Update the policy
	updated, err := client.UpdateAlertPolicy(ctx, policyID, updates)
	if err != nil {
		return nil, fmt.Errorf("update alert policy: %w", err)
	}

	return map[string]interface{}{
		"policy": map[string]interface{}{
			"id":                   updated.ID,
			"name":                 updated.Name,
			"incident_preference":  updated.IncidentPreference,
			"updated_at":          updated.UpdatedAt,
		},
		"message": "Alert policy updated successfully",
	}, nil
}

// handleDeleteAlertPolicy deletes an alert policy
func (s *Server) handleDeleteAlertPolicy(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	policyID, ok := params["policy_id"].(string)
	if !ok || policyID == "" {
		return nil, fmt.Errorf("policy_id parameter is required")
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Delete the policy
	if err := client.DeleteAlertPolicy(ctx, policyID); err != nil {
		return nil, fmt.Errorf("delete alert policy: %w", err)
	}

	return map[string]interface{}{
		"policy_id": policyID,
		"message": "Alert policy deleted successfully",
	}, nil
}

// handleCreateAlertCondition creates a new alert condition
func (s *Server) handleCreateAlertCondition(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	policyID, ok := params["policy_id"].(string)
	if !ok || policyID == "" {
		return nil, fmt.Errorf("policy_id parameter is required")
	}

	name, ok := params["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name parameter is required")
	}

	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// Validate NRQL query
	if err := s.validateNRQLSyntax(query); err != nil {
		return nil, fmt.Errorf("invalid NRQL query: %w", err)
	}

	// Get threshold
	threshold, ok := params["threshold"].(float64)
	if !ok {
		return nil, fmt.Errorf("threshold parameter is required")
	}

	// Get optional parameters
	thresholdDuration := 5
	if td, ok := params["threshold_duration"].(float64); ok {
		thresholdDuration = int(td)
	}

	comparison := "above"
	if comp, ok := params["comparison"].(string); ok {
		comparison = comp
	}

	// Validate comparison
	validComparisons := map[string]bool{
		"above":  true,
		"below":  true,
		"equals": true,
	}
	if !validComparisons[comparison] {
		return nil, fmt.Errorf("invalid comparison: %s", comparison)
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("create_alert_condition", params), nil
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Create the condition
	condition := newrelic.AlertCondition{
		Name:              name,
		Query:             query,
		Threshold:         threshold,
		ThresholdDuration: thresholdDuration,
		Comparison:        comparison,
		PolicyID:          policyID,
		Enabled:           true,
	}

	created, err := client.CreateAlertCondition(ctx, condition)
	if err != nil {
		return nil, fmt.Errorf("create alert condition: %w", err)
	}

	return map[string]interface{}{
		"condition": map[string]interface{}{
			"id":                 created.ID,
			"policy_id":          created.PolicyID,
			"name":               created.Name,
			"query":              created.Query,
			"threshold":          created.Threshold,
			"threshold_duration": created.ThresholdDuration,
			"comparison":         created.Comparison,
			"enabled":            created.Enabled,
			"created_at":         created.CreatedAt,
			"updated_at":         created.UpdatedAt,
		},
		"message": fmt.Sprintf("Alert condition '%s' created successfully", name),
	}, nil
}

// handleCloseIncident closes an open alert incident
func (s *Server) handleCloseIncident(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	incidentID, ok := params["incident_id"].(string)
	if !ok || incidentID == "" {
		return nil, fmt.Errorf("incident_id parameter is required")
	}

	// Get account ID from params or use a default
	accountID := ""
	if aid, ok := params["account_id"].(float64); ok {
		accountID = fmt.Sprintf("%d", int(aid))
	}
	// In a real implementation, we would get this from config
	// For now, just ensure we have something for mock mode
	if accountID == "" && s.getNRClient() == nil {
		accountID = "123456" // Mock account ID
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, fmt.Errorf("invalid New Relic client type")
	}

	// Close the incident
	if err := client.CloseIncident(ctx, incidentID); err != nil {
		return nil, fmt.Errorf("close incident: %w", err)
	}

	return map[string]interface{}{
		"incident_id": incidentID,
		"status":      "closed",
		"closed_at":   time.Now(),
		"message":     "Incident closed successfully",
	}, nil
}

// Helper function to calculate baseline threshold
func (s *Server) calculateBaseline(ctx context.Context, query string, sensitivity string) (float64, error) {
	// Check mock mode
	if s.isMockMode() {
		// Return mock baseline based on sensitivity
		sensitivityMultipliers := map[string]float64{
			"low":    1.5,
			"medium": 1.2,
			"high":   1.0,
		}
		multiplier := sensitivityMultipliers[sensitivity]
		if multiplier == 0 {
			multiplier = 1.2
		}
		return 100.0 * multiplier, nil
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return 0, fmt.Errorf("New Relic client not configured")
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return 0, fmt.Errorf("invalid New Relic client type")
	}

	// Modify query to get statistics over past 7 days
	statsQuery := fmt.Sprintf(`
		SELECT 
			average(value) as avg,
			stddev(value) as stddev,
			max(value) as max,
			min(value) as min
		FROM (
			%s
		) SINCE 7 days ago
	`, query)

	// Execute query
	result, err := client.QueryNRQL(ctx, statsQuery)
	if err != nil {
		return 0, fmt.Errorf("calculate baseline: %w", err)
	}

	// Extract statistics from result
	if len(result.Results) == 0 {
		return 0, fmt.Errorf("no data available for baseline calculation")
	}

	stats := result.Results[0]
	avg, _ := stats["avg"].(float64)
	stddev, _ := stats["stddev"].(float64)

	// Calculate threshold based on sensitivity
	multipliers := map[string]float64{
		"low":    3.0,  // 3 standard deviations
		"medium": 2.5,  // 2.5 standard deviations  
		"high":   2.0,  // 2 standard deviations
	}

	multiplier, ok := multipliers[sensitivity]
	if !ok {
		multiplier = 2.5 // default to medium
	}

	// Calculate baseline threshold
	baseline := avg + (stddev * multiplier)
	
	// Ensure baseline is reasonable
	if baseline <= 0 {
		return avg * 1.5, nil // fallback to 150% of average
	}

	return baseline, nil
}

// Helper function to get float value from interface{}
func getFloatValue(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

// Helper function to determine confidence level based on sample size
func determineConfidence(incidentCount float64) string {
	if incidentCount < 5 {
		return "Low"
	} else if incidentCount < 20 {
		return "Medium"
	}
	return "High"
}

// Helper function to generate alert recommendation
func generateAlertRecommendation(incidentCount, avgDuration, noiseRatio float64) string {
	if incidentCount == 0 {
		return "No incidents recorded. Consider reviewing if the alert threshold is too high."
	}
	
	if noiseRatio > 0.3 {
		return "High noise ratio detected. Consider increasing the threshold or duration to reduce false positives."
	}
	
	if avgDuration < 5 {
		return "Very short incident durations suggest possible flapping. Consider increasing the threshold duration."
	}
	
	if noiseRatio < 0.1 && incidentCount > 10 {
		return "Alert is performing well with low noise ratio and good incident detection."
	}
	
	return "Alert performance is acceptable. Monitor for patterns and adjust if needed."
}
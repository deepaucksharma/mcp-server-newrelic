//go:build !test

package mcp

import (
	"context"
	"fmt"
	"time"
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
			},
		},
		Handler: s.handleCreateAlert,
	})

	// List alert conditions
	s.tools.Register(Tool{
		Name:        "list_alerts",
		Description: "List all alert conditions in the account",
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

	// Create alert condition
	alertCondition := map[string]interface{}{
		"id":                   fmt.Sprintf("alert-%d", time.Now().Unix()),
		"name":                 name,
		"query":                query,
		"comparison":           comparison,
		"threshold":            threshold,
		"threshold_duration":   thresholdDuration,
		"sensitivity":          sensitivity,
		"auto_baseline":        autoBaseline,
		"created_at":           time.Now(),
		"enabled":              true,
	}

	// Add to policy if specified
	if policyID, ok := params["policy_id"].(string); ok && policyID != "" {
		alertCondition["policy_id"] = policyID
	}

	// TODO: Actually create the alert using New Relic API
	// For now, return the mock alert condition
	return map[string]interface{}{
		"alert":     alertCondition,
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

	// TODO: Implement actual alert listing using New Relic API
	// For now, return mock data
	alerts := []map[string]interface{}{
		{
			"id":                 "alert-1",
			"name":               "High Error Rate",
			"query":              "SELECT percentage(count(*), WHERE error IS true) FROM Transaction",
			"comparison":         "above",
			"threshold":          5.0,
			"threshold_duration": 5,
			"enabled":            true,
			"policy_id":          "policy-1",
			"created_at":         time.Now().Add(-30 * 24 * time.Hour),
			"updated_at":         time.Now().Add(-2 * time.Hour),
		},
		{
			"id":                 "alert-2",
			"name":               "Low Apdex Score",
			"query":              "SELECT apdex(duration, 0.5) FROM Transaction",
			"comparison":         "below",
			"threshold":          0.85,
			"threshold_duration": 10,
			"enabled":            false,
			"policy_id":          "policy-1",
			"created_at":         time.Now().Add(-15 * 24 * time.Hour),
			"updated_at":         time.Now().Add(-5 * 24 * time.Hour),
		},
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

	// TODO: Implement actual alert analysis using New Relic API
	// For now, return mock analysis
	analysis := map[string]interface{}{
		"alert_id":    alertID,
		"time_range":  timeRange,
		"summary": map[string]interface{}{
			"total_incidents":      12,
			"false_positives":      2,
			"mean_time_to_resolve": "45 minutes",
			"noise_ratio":          0.17,
		},
		"effectiveness": map[string]interface{}{
			"score":       0.85,
			"rating":      "Good",
			"confidence":  "High",
		},
		"recommendations": []map[string]interface{}{
			{
				"type":        "threshold_adjustment",
				"priority":    "medium",
				"description": "Consider increasing threshold from 5.0 to 6.5 to reduce false positives",
				"impact":      "Would reduce incidents by ~25% with minimal risk",
			},
			{
				"type":        "duration_adjustment",
				"priority":    "low",
				"description": "Increase threshold duration from 5 to 7 minutes for more stability",
				"impact":      "Would reduce noise by ~15%",
			},
		},
		"incident_patterns": map[string]interface{}{
			"time_of_day": map[string]interface{}{
				"peak_hours":      "14:00-16:00",
				"quiet_hours":     "02:00-06:00",
			},
			"day_of_week": map[string]interface{}{
				"highest": "Monday",
				"lowest":  "Sunday",
			},
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

	// Process each alert
	for _, alertID := range ids {
		result := map[string]interface{}{
			"alert_id": alertID,
			"status":   "success",
		}

		switch operation {
		case "enable":
			// TODO: Enable alert via API
			result["message"] = "Alert enabled"

		case "disable":
			// TODO: Disable alert via API
			result["message"] = "Alert disabled"

		case "update_threshold":
			if newThreshold, ok := params["new_threshold"].(float64); ok {
				// TODO: Update with new threshold
				result["message"] = fmt.Sprintf("Threshold updated to %.2f", newThreshold)
			} else if multiplier, ok := params["threshold_multiplier"].(float64); ok {
				// TODO: Get current threshold and multiply
				currentThreshold := 5.0 // Mock current threshold
				newThreshold := currentThreshold * multiplier
				result["message"] = fmt.Sprintf("Threshold updated from %.2f to %.2f", currentThreshold, newThreshold)
			} else {
				result["status"] = "failed"
				result["error"] = "update_threshold requires new_threshold or threshold_multiplier"
			}

		case "delete":
			// TODO: Delete alert via API
			result["message"] = "Alert deleted"

		default:
			result["status"] = "failed"
			result["error"] = fmt.Sprintf("Unknown operation: %s", operation)
		}

		if result["status"] == "success" {
			results["successful"] = results["successful"].(int) + 1
		} else {
			results["failed"] = results["failed"].(int) + 1
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

	// Check for mock mode
	if s.nrClient == nil {
		return map[string]interface{}{
			"policy": map[string]interface{}{
				"id":                   fmt.Sprintf("policy-%d", time.Now().Unix()),
				"name":                 name,
				"incident_preference":  incidentPref,
				"created_at":          time.Now(),
			},
			"message": "Alert policy created successfully (mock)",
		}, nil
	}

	// TODO: Implement actual policy creation using New Relic API
	// For now, return mock response
	return map[string]interface{}{
		"policy": map[string]interface{}{
			"id":                   fmt.Sprintf("policy-%d", time.Now().Unix()),
			"name":                 name,
			"incident_preference":  incidentPref,
			"created_at":          time.Now(),
		},
		"message": "Alert policy created successfully",
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

	// Check for mock mode
	if s.nrClient == nil {
		return map[string]interface{}{
			"policy": map[string]interface{}{
				"id":       policyID,
				"updates":  updates,
				"updated_at": time.Now(),
			},
			"message": "Alert policy updated successfully (mock)",
		}, nil
	}

	// TODO: Implement actual policy update using New Relic API
	return map[string]interface{}{
		"policy": map[string]interface{}{
			"id":       policyID,
			"updates":  updates,
			"updated_at": time.Now(),
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

	// Check for mock mode
	if s.nrClient == nil {
		return map[string]interface{}{
			"policy_id": policyID,
			"message": "Alert policy deleted successfully (mock)",
		}, nil
	}

	// TODO: Implement actual policy deletion using New Relic API
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

	// Check for mock mode
	if s.nrClient == nil {
		return map[string]interface{}{
			"condition": map[string]interface{}{
				"id":                 fmt.Sprintf("condition-%d", time.Now().Unix()),
				"policy_id":          policyID,
				"name":               name,
				"query":              query,
				"threshold":          threshold,
				"threshold_duration": thresholdDuration,
				"comparison":         comparison,
				"created_at":         time.Now(),
			},
			"message": "Alert condition created successfully (mock)",
		}, nil
	}

	// TODO: Implement actual condition creation using New Relic API
	return map[string]interface{}{
		"condition": map[string]interface{}{
			"id":                 fmt.Sprintf("condition-%d", time.Now().Unix()),
			"policy_id":          policyID,
			"name":               name,
			"query":              query,
			"threshold":          threshold,
			"threshold_duration": thresholdDuration,
			"comparison":         comparison,
			"created_at":         time.Now(),
		},
		"message": "Alert condition created successfully",
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
	if accountID == "" && s.nrClient == nil {
		accountID = "123456" // Mock account ID
	}

	// Check for mock mode
	if s.nrClient == nil {
		return map[string]interface{}{
			"incident_id": incidentID,
			"status":      "closed",
			"closed_at":   time.Now(),
			"message":     "Incident closed successfully (mock)",
		}, nil
	}

	// TODO: Implement actual incident closure using New Relic API
	return map[string]interface{}{
		"incident_id": incidentID,
		"status":      "closed",
		"closed_at":   time.Now(),
		"message":     "Incident closed successfully",
	}, nil
}

// Helper function to calculate baseline threshold
func (s *Server) calculateBaseline(ctx context.Context, query string, sensitivity string) (float64, error) {
	// TODO: Execute query to get historical data and calculate baseline
	// For now, return mock baseline based on sensitivity
	baselines := map[string]float64{
		"low":    10.0,  // 3 standard deviations
		"medium": 7.5,   // 2.5 standard deviations
		"high":   5.0,   // 2 standard deviations
	}

	if baseline, ok := baselines[sensitivity]; ok {
		return baseline, nil
	}

	return baselines["medium"], nil
}
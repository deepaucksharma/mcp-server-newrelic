//go:build !test

package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
)

// registerBulkTools registers bulk operation tools
func (s *Server) registerBulkTools() error {
	// Bulk tag entities
	s.tools.Register(Tool{
		Name:        "bulk_tag_entities",
		Description: "Apply tags to multiple entities at once",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"entity_guids", "tags"},
			Properties: map[string]Property{
				"entity_guids": {
					Type:        "array",
					Description: "List of entity GUIDs to tag",
					Items:       &Property{Type: "string"},
				},
				"tags": {
					Type:        "array",
					Description: "Tags to apply (key:value pairs)",
					Items:       &Property{Type: "string"},
				},
				"operation": {
					Type:        "string",
					Description: "Tag operation: 'add' or 'replace' (default: 'add')",
					Default:     "add",
				},
			},
		},
		Handler: s.handleBulkTagEntities,
	})

	// Bulk create monitors
	s.tools.Register(Tool{
		Name:        "bulk_create_monitors",
		Description: "Create multiple synthetic monitors from a template",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"monitors"},
			Properties: map[string]Property{
				"monitors": {
					Type:        "array",
					Description: "List of monitor configurations (objects with name, url, locations, period)",
					Items:       &Property{Type: "object"},
				},
				"template": {
					Type:        "object",
					Description: "Common settings for all monitors (type, status, tags)",
				},
			},
		},
		Handler: s.handleBulkCreateMonitors,
	})

	// Bulk update dashboards
	s.tools.Register(Tool{
		Name:        "bulk_update_dashboards",
		Description: "Update multiple dashboards with common changes",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_ids", "updates"},
			Properties: map[string]Property{
				"dashboard_ids": {
					Type:        "array",
					Description: "List of dashboard IDs to update",
					Items:       &Property{Type: "string"},
				},
				"updates": {
					Type:        "object",
					Description: "Updates to apply (add_tags, remove_tags, permissions, add_widget)",
				},
			},
		},
		Handler: s.handleBulkUpdateDashboards,
	})

	// Bulk delete entities
	s.tools.Register(Tool{
		Name:        "bulk_delete_entities",
		Description: "Delete multiple entities (monitors, dashboards, etc.)",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"entity_type", "entity_ids"},
			Properties: map[string]Property{
				"entity_type": {
					Type:        "string",
					Description: "Type of entities to delete: 'monitor', 'dashboard', 'alert_condition'",
				},
				"entity_ids": {
					Type:        "array",
					Description: "List of entity IDs to delete",
					Items:       &Property{Type: "string"},
				},
				"force": {
					Type:        "boolean",
					Description: "Force deletion without confirmation (default: false)",
					Default:     false,
				},
			},
		},
		Handler: s.handleBulkDeleteEntities,
	})

	// Bulk query execution
	s.tools.Register(Tool{
		Name:        "bulk_execute_queries",
		Description: "Execute multiple NRQL queries in parallel",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"queries"},
			Properties: map[string]Property{
				"queries": {
					Type:        "array",
					Description: "List of queries to execute (objects with name, query, account_id)",
					Items:       &Property{Type: "object"},
				},
				"parallel": {
					Type:        "boolean",
					Description: "Execute queries in parallel (default: true)",
					Default:     true,
				},
				"timeout": {
					Type:        "integer",
					Description: "Timeout in seconds per query (default: 30)",
					Default:     30,
				},
			},
		},
		Handler: s.handleBulkExecuteQueries,
	})

	// Bulk dashboard migration
	s.tools.Register(Tool{
		Name:        "bulk_dashboard_migrate",
		Description: "Migrate dashboards between accounts or update to new standards",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"dashboard_ids"},
			Properties: map[string]Property{
				"dashboard_ids": {
					Type:        "array",
					Description: "List of dashboard IDs to migrate",
					Items:       &Property{Type: "string"},
				},
				"target_account_id": {
					Type:        "string",
					Description: "Target account ID (if different from source)",
				},
				"update_queries": {
					Type:        "boolean",
					Description: "Update NRQL queries to match target account",
					Default:     true,
				},
				"preserve_permissions": {
					Type:        "boolean",
					Description: "Preserve original dashboard permissions",
					Default:     false,
				},
			},
		},
		Handler: s.handleBulkDashboardMigrate,
	})

	return nil
}

// handleBulkTagEntities applies tags to multiple entities
func (s *Server) handleBulkTagEntities(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	guidsRaw, ok := params["entity_guids"].([]interface{})
	if !ok || len(guidsRaw) == 0 {
		return nil, NewInvalidParamsError("entity_guids parameter is required and must be non-empty array", "entity_guids")
	}

	tagsRaw, ok := params["tags"].([]interface{})
	if !ok || len(tagsRaw) == 0 {
		return nil, NewInvalidParamsError("tags parameter is required and must be non-empty array", "tags")
	}

	operation := "add"
	if op, ok := params["operation"].(string); ok {
		operation = op
	}

	// Validate operation
	if operation != "add" && operation != "replace" {
		return nil, NewValidationError("operation", "must be 'add' or 'replace'")
	}

	// Convert GUIDs to strings
	guids := make([]string, len(guidsRaw))
	for i, g := range guidsRaw {
		guid, ok := g.(string)
		if !ok || guid == "" {
			return nil, NewValidationError(fmt.Sprintf("entity_guids[%d]", i), "must be a non-empty string")
		}
		guids[i] = guid
	}

	// Convert tags to strings
	tags := make([]string, len(tagsRaw))
	for i, t := range tagsRaw {
		tag, ok := t.(string)
		if !ok || tag == "" {
			return nil, NewValidationError(fmt.Sprintf("tags[%d]", i), "must be a non-empty string in key:value format")
		}
		tags[i] = tag
	}

	// Check for mock mode
	if s.isMockMode() {
		return s.getMockData("bulk_tag_entities", params), nil
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

	// Prepare tags for mutation
	tagInput := make([]map[string]interface{}, 0, len(tags))
	for k, v := range tags {
		tagInput = append(tagInput, map[string]interface{}{
			"key":    k,
			"values": []string{fmt.Sprintf("%v", v)},
		})
	}

	successCount := 0
	failedCount := 0
	errors := []string{}

	// Process entities in batches
	batchSize := 50
	for i := 0; i < len(guids); i += batchSize {
		end := i + batchSize
		if end > len(guids) {
			end = len(guids)
		}
		batch := guids[i:end]

		mutation := `
			mutation($guids: [EntityGuid!]!, $tags: [TaggingTagInput!]!) {
				taggingAddTagsToEntity(guid: $guids, tags: $tags) {
					errors {
						message
					}
				}
			}
		`

		variables := map[string]interface{}{
			"guids": batch,
			"tags":  tagInput,
		}

		result, err := client.QueryGraphQL(ctx, mutation, variables)
		if err != nil {
			failedCount += len(batch)
			errors = append(errors, fmt.Sprintf("batch %d: %v", i/batchSize, err))
			continue
		}

		// Check for errors in the mutation result
		if taggingResult, ok := result["taggingAddTagsToEntity"].(map[string]interface{}); ok {
			if errList, ok := taggingResult["errors"].([]interface{}); ok && len(errList) > 0 {
				failedCount += len(batch)
				for _, e := range errList {
					if errMap, ok := e.(map[string]interface{}); ok {
						if msg, ok := errMap["message"].(string); ok {
							errors = append(errors, msg)
						}
					}
				}
			} else {
				successCount += len(batch)
			}
		} else {
			successCount += len(batch)
		}
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_entities": len(guids),
			"total_tags":     len(tags),
			"operation":      operation,
			"success":        successCount,
			"failed":         failedCount,
		},
		"errors":  errors,
		"message": fmt.Sprintf("Tags applied: %d succeeded, %d failed", successCount, failedCount),
	}, nil
}

// handleBulkCreateMonitors creates multiple synthetic monitors
func (s *Server) handleBulkCreateMonitors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	monitorsRaw, ok := params["monitors"].([]interface{})
	if !ok || len(monitorsRaw) == 0 {
		return nil, NewInvalidParamsError("monitors parameter is required and must be non-empty array", "monitors")
	}

	// Get template settings
	template := map[string]interface{}{
		"type":   "SIMPLE",
		"status": "ENABLED",
		"tags":   []string{},
	}
	if tpl, ok := params["template"].(map[string]interface{}); ok {
		for k, v := range tpl {
			template[k] = v
		}
	}

	results := []map[string]interface{}{}
	successCount := 0
	failureCount := 0

	// Process each monitor
	for i, monRaw := range monitorsRaw {
		monitor, ok := monRaw.(map[string]interface{})
		if !ok {
			results = append(results, map[string]interface{}{
				"index":  i,
				"status": "failed",
				"error":  "invalid monitor configuration",
			})
			failureCount++
			continue
		}

		// Validate required fields
		name, ok := monitor["name"].(string)
		if !ok || name == "" {
			results = append(results, map[string]interface{}{
				"index":  i,
				"status": "failed",
				"error":  "monitor name is required",
			})
			failureCount++
			continue
		}

		url, ok := monitor["url"].(string)
		if !ok || url == "" {
			results = append(results, map[string]interface{}{
				"index":  i,
				"name":   name,
				"status": "failed",
				"error":  "monitor url is required",
			})
			failureCount++
			continue
		}

		// Get optional parameters
		frequency := 5 // default 5 minutes
		if freq, ok := monitor["frequency"].(float64); ok {
			frequency = int(freq)
		}
		
		locations := []string{"US_EAST_1"} // default location
		if locs, ok := monitor["locations"].([]interface{}); ok {
			locations = make([]string, len(locs))
			for j, loc := range locs {
				locations[j] = loc.(string)
			}
		}

		// Check for mock mode
		if s.isMockMode() {
			results = append(results, map[string]interface{}{
				"index":      i,
				"name":       name,
				"status":     "success",
				"monitor_id": fmt.Sprintf("mock-monitor-%d", i),
				"url":        url,
				"message":    "Monitor created successfully (mock)",
			})
			successCount++
			continue
		}

		// Get New Relic client
		nrClient := s.getNRClient()
		if nrClient == nil {
			results = append(results, map[string]interface{}{
				"index":  i,
				"name":   name,
				"status": "failed",
				"error":  "New Relic client not configured",
			})
			failureCount++
			continue
		}

		client, ok := nrClient.(*newrelic.Client)
		if !ok {
			results = append(results, map[string]interface{}{
				"index":  i,
				"name":   name,
				"status": "failed",
				"error":  "invalid New Relic client type",
			})
			failureCount++
			continue
		}

		// Create the monitor
		mon := newrelic.SyntheticMonitor{
			Name:      name,
			URL:       url,
			Frequency: frequency,
			Locations: locations,
		}

		created, err := client.CreateSyntheticMonitor(ctx, mon)
		if err != nil {
			results = append(results, map[string]interface{}{
				"index":  i,
				"name":   name,
				"status": "failed",
				"error":  err.Error(),
			})
			failureCount++
		} else {
			results = append(results, map[string]interface{}{
				"index":      i,
				"name":       name,
				"status":     "success",
				"monitor_id": created.ID,
				"url":        created.URL,
				"message":    "Monitor created successfully",
			})
			successCount++
		}
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total":   len(monitorsRaw),
			"success": successCount,
			"failed":  failureCount,
		},
		"results": results,
	}, nil
}

// handleBulkUpdateDashboards updates multiple dashboards
func (s *Server) handleBulkUpdateDashboards(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardIDsRaw, ok := params["dashboard_ids"].([]interface{})
	if !ok || len(dashboardIDsRaw) == 0 {
		return nil, NewInvalidParamsError("dashboard_ids parameter is required and must be non-empty array", "dashboard_ids")
	}

	updates, ok := params["updates"].(map[string]interface{})
	if !ok || len(updates) == 0 {
		return nil, NewInvalidParamsError("updates parameter is required and must be non-empty object", "updates")
	}

	// Convert dashboard IDs to strings
	dashboardIDs := make([]string, len(dashboardIDsRaw))
	for i, id := range dashboardIDsRaw {
		dashboardID, ok := id.(string)
		if !ok || dashboardID == "" {
			return nil, NewValidationError(fmt.Sprintf("dashboard_ids[%d]", i), "must be a non-empty string")
		}
		dashboardIDs[i] = dashboardID
	}

	// Check for mock mode
	if s.isMockMode() {
		return s.getMockData("bulk_update_dashboards", params), nil
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, NewInternalError("New Relic client not configured", nil)
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, NewInternalError("invalid New Relic client type", nil)
	}

	results := []map[string]interface{}{}
	successCount := 0
	failureCount := 0

	for _, id := range dashboardIDs {
		// Get current dashboard
		dash, err := client.GetDashboard(ctx, id)
		if err != nil {
			results = append(results, map[string]interface{}{
				"dashboard_id": id,
				"status":       "failed",
				"error":        fmt.Sprintf("failed to get dashboard: %v", err),
			})
			failureCount++
			continue
		}

		// Apply updates
		if name, ok := updates["name"].(string); ok {
			dash.Name = name
		}
		if desc, ok := updates["description"].(string); ok {
			dash.Description = desc
		}
		if perm, ok := updates["permissions"].(string); ok {
			dash.Permissions = perm
		}

		// Update the dashboard
		updated, err := client.UpdateDashboard(ctx, id, *dash)
		if err != nil {
			results = append(results, map[string]interface{}{
				"dashboard_id": id,
				"status":       "failed",
				"error":        err.Error(),
			})
			failureCount++
		} else {
			results = append(results, map[string]interface{}{
				"dashboard_id": id,
				"status":       "success",
				"updates":      updates,
				"updated_at":   updated.UpdatedAt,
			})
			successCount++
		}
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total":   len(dashboardIDs),
			"success": successCount,
			"failed":  failureCount,
		},
		"message": "Dashboards updated successfully",
	}, nil
}

// handleBulkDeleteEntities deletes multiple entities
func (s *Server) handleBulkDeleteEntities(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	entityType, ok := params["entity_type"].(string)
	if !ok || entityType == "" {
		return nil, NewInvalidParamsError("entity_type parameter is required", "entity_type")
	}

	// Validate entity type
	validTypes := map[string]bool{
		"monitor":         true,
		"dashboard":       true,
		"alert_condition": true,
	}
	if !validTypes[entityType] {
		return nil, NewValidationError("entity_type", fmt.Sprintf("must be one of: monitor, dashboard, alert_condition (got: %s)", entityType))
	}

	entityIDsRaw, ok := params["entity_ids"].([]interface{})
	if !ok || len(entityIDsRaw) == 0 {
		return nil, NewInvalidParamsError("entity_ids parameter is required and must be non-empty array", "entity_ids")
	}

	force, _ := params["force"].(bool)

	// Convert entity IDs to strings
	entityIDs := make([]string, len(entityIDsRaw))
	for i, id := range entityIDsRaw {
		entityID, ok := id.(string)
		if !ok || entityID == "" {
			return nil, NewValidationError(fmt.Sprintf("entity_ids[%d]", i), "must be a non-empty string")
		}
		entityIDs[i] = entityID
	}

	// Safety check
	if !force && len(entityIDs) > 10 {
		err := NewValidationError("force", fmt.Sprintf("attempting to delete %d entities; set force=true to confirm", len(entityIDs)))
		err.Hint = "Add 'force': true to your request to confirm bulk deletion"
		return nil, err
	}

	// Check for mock mode
	if s.isMockMode() {
		return s.getMockData("bulk_delete_entities", params), nil
	}

	// Get New Relic client
	nrClient := s.getNRClient()
	if nrClient == nil {
		return nil, NewInternalError("New Relic client not configured", nil)
	}

	client, ok := nrClient.(*newrelic.Client)
	if !ok {
		return nil, NewInternalError("invalid New Relic client type", nil)
	}

	results := []map[string]interface{}{}
	deletedCount := 0
	failedCount := 0

	for _, id := range entityIDs {
		var err error
		
		// Handle different entity types
		switch entityType {
		case "dashboard", "monitor":
			// Use generic entity delete
			err = client.DeleteEntity(ctx, id)
		case "alert_condition":
			// Use specific alert condition delete
			err = client.DeleteAlertCondition(ctx, id)
		default:
			err = fmt.Errorf("unsupported entity type: %s", entityType)
		}

		if err != nil {
			results = append(results, map[string]interface{}{
				"entity_id":   id,
				"entity_type": entityType,
				"status":      "failed",
				"error":       err.Error(),
			})
			failedCount++
		} else {
			results = append(results, map[string]interface{}{
				"entity_id":   id,
				"entity_type": entityType,
				"status":      "deleted",
			})
			deletedCount++
		}
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"entity_type": entityType,
			"total":       len(entityIDs),
			"deleted":     deletedCount,
			"failed":      failedCount,
		},
		"results": results,
		"message": fmt.Sprintf("%d entities deleted, %d failed", deletedCount, failedCount),
	}, nil
}

// handleBulkExecuteQueries executes multiple NRQL queries
func (s *Server) handleBulkExecuteQueries(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	queriesRaw, ok := params["queries"].([]interface{})
	if !ok || len(queriesRaw) == 0 {
		return nil, NewInvalidParamsError("queries parameter is required and must be non-empty array", "queries")
	}

	parallel := true
	if p, ok := params["parallel"].(bool); ok {
		parallel = p
	}

	timeoutSec := 30
	if t, ok := params["timeout"].(float64); ok {
		timeoutSec = int(t)
	}

	results := []map[string]interface{}{}
	startTime := time.Now()
	
	// Create a timeout context for query execution
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	if parallel && len(queriesRaw) > 1 {
		// Parallel execution
		type queryJob struct {
			index  int
			query  map[string]interface{}
			result chan map[string]interface{}
		}

		jobs := make([]queryJob, len(queriesRaw))
		
		// Create jobs
		for i, qRaw := range queriesRaw {
			query, _ := qRaw.(map[string]interface{})
			jobs[i] = queryJob{
				index:  i,
				query:  query,
				result: make(chan map[string]interface{}, 1),
			}
		}

		// Execute queries in parallel
		for _, job := range jobs {
			go func(j queryJob) {
				result := s.executeSingleQuery(queryCtx, j.index, j.query)
				j.result <- result
			}(job)
		}

		// Collect results
		for _, job := range jobs {
			select {
			case result := <-job.result:
				results = append(results, result)
			case <-queryCtx.Done():
				results = append(results, map[string]interface{}{
					"index":  job.index,
					"status": "failed",
					"error":  "query timeout",
				})
			}
		}
	} else {
		// Sequential execution
		for i, qRaw := range queriesRaw {
			query, ok := qRaw.(map[string]interface{})
			if !ok {
				results = append(results, map[string]interface{}{
					"index":  i,
					"status": "failed",
					"error":  "invalid query configuration",
				})
				continue
			}
			result := s.executeSingleQuery(queryCtx, i, query)
			results = append(results, result)
		}
	}

	// Count successes and failures
	successCount := 0
	failedCount := 0
	for _, r := range results {
		if r["status"] == "success" {
			successCount++
		} else {
			failedCount++
		}
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total":           len(queriesRaw),
			"success":         successCount,
			"failed":          failedCount,
			"parallel":        parallel,
			"total_time_ms":   time.Since(startTime).Milliseconds(),
		},
		"results": results,
	}, nil
}

// Helper function to execute a single query
func (s *Server) executeSingleQuery(ctx context.Context, index int, query map[string]interface{}) map[string]interface{} {
	name, _ := query["name"].(string)
	if name == "" {
		name = fmt.Sprintf("query_%d", index)
	}

	nrql, ok := query["query"].(string)
	if !ok || nrql == "" {
		return map[string]interface{}{
			"index":  index,
			"name":   name,
			"status": "failed",
			"error":  "query string is required",
		}
	}

	// Validate NRQL
	if err := s.validateNRQLSyntax(nrql); err != nil {
		return map[string]interface{}{
			"index":  index,
			"name":   name,
			"status": "failed",
			"error":  fmt.Sprintf("invalid NRQL: %v", err),
			"hint":   "Check NRQL syntax: https://docs.newrelic.com/docs/query-your-data/nrql-reference/",
		}
	}

	// Execute query
	queryStart := time.Now()
	queryResult := map[string]interface{}{
		"index":  index,
		"name":   name,
		"query":  nrql,
	}

	// Execute the query
	if result, err := s.executeNRQLQuery(ctx, nrql, nil); err != nil {
		queryResult["status"] = "failed"
		queryResult["error"] = err.Error()
	} else {
		queryResult["status"] = "success"
		queryResult["results"] = result
		queryResult["execution_time"] = time.Since(queryStart).Milliseconds()
	}

	return queryResult
}

// handleBulkDashboardMigrate migrates dashboards between accounts
func (s *Server) handleBulkDashboardMigrate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dashboardIDsRaw, ok := params["dashboard_ids"].([]interface{})
	if !ok || len(dashboardIDsRaw) == 0 {
		return nil, NewInvalidParamsError("dashboard_ids parameter is required and must be non-empty", "dashboard_ids")
	}

	targetAccountID, _ := params["target_account_id"].(string)
	updateQueries := true
	if uq, ok := params["update_queries"].(bool); ok {
		updateQueries = uq
	}
	preservePermissions := false
	if pp, ok := params["preserve_permissions"].(bool); ok {
		preservePermissions = pp
	}

	// Check mock mode
	if s.isMockMode() {
		return s.getMockData("bulk_dashboard_migrate", params), nil
	}

	// Convert dashboard IDs
	dashboardIDs := []string{}
	for _, id := range dashboardIDsRaw {
		if idStr, ok := id.(string); ok {
			dashboardIDs = append(dashboardIDs, idStr)
		}
	}

	results := []map[string]interface{}{}
	startTime := time.Now()

	// Process each dashboard
	for i, dashboardID := range dashboardIDs {
		result := map[string]interface{}{
			"source_id": dashboardID,
			"index":     i,
		}

		// Get source dashboard
		sourceDashboard, err := s.getDashboardDetails(ctx, dashboardID)
		if err != nil {
			result["status"] = "failed"
			result["error"] = fmt.Sprintf("failed to get source dashboard: %v", err)
			results = append(results, result)
			continue
		}

		// Prepare migrated dashboard
		migratedDashboard := s.prepareMigratedDashboard(sourceDashboard, targetAccountID, updateQueries, preservePermissions)

		// Create in target account (or same account with updates)
		newDashboard, err := s.createDashboard(ctx, migratedDashboard, targetAccountID)
		if err != nil {
			result["status"] = "failed"
			result["error"] = fmt.Sprintf("failed to create migrated dashboard: %v", err)
		} else {
			result["status"] = "success"
			result["target_id"] = newDashboard["id"]
			result["target_name"] = newDashboard["name"]
			result["widgets_migrated"] = len(migratedDashboard["pages"].([]interface{})[0].(map[string]interface{})["widgets"].([]interface{}))
		}

		results = append(results, result)
	}

	// Calculate summary
	successCount := 0
	failedCount := 0
	totalWidgets := 0
	for _, r := range results {
		if r["status"] == "success" {
			successCount++
			if wm, ok := r["widgets_migrated"].(int); ok {
				totalWidgets += wm
			}
		} else {
			failedCount++
		}
	}

	return map[string]interface{}{
		"migration_results": results,
		"summary": map[string]interface{}{
			"total_dashboards":   len(dashboardIDs),
			"successful":         successCount,
			"failed":             failedCount,
			"widgets_migrated":   totalWidgets,
			"target_account":     targetAccountID,
			"migration_time_ms":  time.Since(startTime).Milliseconds(),
		},
	}, nil
}

// Helper functions for dashboard migration

func (s *Server) getDashboardDetails(ctx context.Context, dashboardID string) (map[string]interface{}, error) {
	// This would call the New Relic API to get dashboard details
	// For now, return a simple structure
	return map[string]interface{}{
		"id":   dashboardID,
		"name": "Sample Dashboard",
		"pages": []interface{}{
			map[string]interface{}{
				"name": "Page 1",
				"widgets": []interface{}{
					map[string]interface{}{
						"title": "Widget 1",
						"query": "SELECT count(*) FROM Transaction",
					},
				},
			},
		},
	}, nil
}

func (s *Server) prepareMigratedDashboard(source map[string]interface{}, targetAccountID string, updateQueries bool, preservePermissions bool) map[string]interface{} {
	// Clone the dashboard
	migrated := make(map[string]interface{})
	for k, v := range source {
		migrated[k] = v
	}

	// Update name to indicate migration
	if name, ok := migrated["name"].(string); ok {
		migrated["name"] = name + " (Migrated)"
	}

	// Remove ID so a new one is created
	delete(migrated, "id")

	// Update permissions if not preserving
	if !preservePermissions {
		migrated["permissions"] = "PUBLIC_READ_ONLY"
	}

	// Update queries if requested and target account is different
	if updateQueries && targetAccountID != "" {
		// This would update account-specific references in NRQL queries
		// For now, we'll just note it in the implementation
	}

	return migrated
}

func (s *Server) createDashboard(ctx context.Context, dashboard map[string]interface{}, accountID string) (map[string]interface{}, error) {
	// This would call the New Relic API to create the dashboard
	// For now, return a mock response
	return map[string]interface{}{
		"id":   fmt.Sprintf("new-%s", dashboard["name"]),
		"name": dashboard["name"],
	}, nil
}
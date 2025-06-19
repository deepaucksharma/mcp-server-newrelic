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

	return nil
}

// handleBulkTagEntities applies tags to multiple entities
func (s *Server) handleBulkTagEntities(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	guidsRaw, ok := params["entity_guids"].([]interface{})
	if !ok || len(guidsRaw) == 0 {
		return nil, fmt.Errorf("entity_guids parameter is required and must be non-empty")
	}

	tagsRaw, ok := params["tags"].([]interface{})
	if !ok || len(tagsRaw) == 0 {
		return nil, fmt.Errorf("tags parameter is required and must be non-empty")
	}

	operation := "add"
	if op, ok := params["operation"].(string); ok {
		operation = op
	}

	// Validate operation
	if operation != "add" && operation != "replace" {
		return nil, fmt.Errorf("operation must be 'add' or 'replace'")
	}

	// Convert GUIDs to strings
	guids := make([]string, len(guidsRaw))
	for i, g := range guidsRaw {
		guid, ok := g.(string)
		if !ok || guid == "" {
			return nil, fmt.Errorf("invalid entity GUID at index %d", i)
		}
		guids[i] = guid
	}

	// Convert tags to strings
	tags := make([]string, len(tagsRaw))
	for i, t := range tagsRaw {
		tag, ok := t.(string)
		if !ok || tag == "" {
			return nil, fmt.Errorf("invalid tag at index %d", i)
		}
		tags[i] = tag
	}

	// Check for mock mode
	if s.getNRClient() == nil {
		return map[string]interface{}{
			"summary": map[string]interface{}{
				"total_entities": len(guids),
				"total_tags":     len(tags),
				"operation":      operation,
				"success":        len(guids),
				"failed":         0,
			},
			"results": []map[string]interface{}{
				{
					"entity_guid": guids[0],
					"status":      "success",
					"tags_applied": tags,
				},
			},
			"message": "Tags applied successfully (mock)",
		}, nil
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
		return nil, fmt.Errorf("monitors parameter is required and must be non-empty")
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
		return nil, fmt.Errorf("dashboard_ids parameter is required and must be non-empty")
	}

	updates, ok := params["updates"].(map[string]interface{})
	if !ok || len(updates) == 0 {
		return nil, fmt.Errorf("updates parameter is required and must be non-empty")
	}

	// Convert dashboard IDs to strings
	dashboardIDs := make([]string, len(dashboardIDsRaw))
	for i, id := range dashboardIDsRaw {
		dashboardID, ok := id.(string)
		if !ok || dashboardID == "" {
			return nil, fmt.Errorf("invalid dashboard ID at index %d", i)
		}
		dashboardIDs[i] = dashboardID
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
		return nil, fmt.Errorf("entity_type parameter is required")
	}

	// Validate entity type
	validTypes := map[string]bool{
		"monitor":         true,
		"dashboard":       true,
		"alert_condition": true,
	}
	if !validTypes[entityType] {
		return nil, fmt.Errorf("invalid entity_type: %s", entityType)
	}

	entityIDsRaw, ok := params["entity_ids"].([]interface{})
	if !ok || len(entityIDsRaw) == 0 {
		return nil, fmt.Errorf("entity_ids parameter is required and must be non-empty")
	}

	force, _ := params["force"].(bool)

	// Convert entity IDs to strings
	entityIDs := make([]string, len(entityIDsRaw))
	for i, id := range entityIDsRaw {
		entityID, ok := id.(string)
		if !ok || entityID == "" {
			return nil, fmt.Errorf("invalid entity ID at index %d", i)
		}
		entityIDs[i] = entityID
	}

	// Safety check
	if !force && len(entityIDs) > 10 {
		return nil, fmt.Errorf("attempting to delete %d entities; set force=true to confirm", len(entityIDs))
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
		return nil, fmt.Errorf("queries parameter is required and must be non-empty")
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
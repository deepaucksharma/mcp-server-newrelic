package mcp

import (
	"context"
	"fmt"
	"time"
)

// DryRunResult represents the result of a dry-run operation
type DryRunResult struct {
	Operation     string                 `json:"operation"`
	WouldSucceed  bool                   `json:"would_succeed"`
	Changes       []ProposedChange       `json:"changes"`
	Validations   []ValidationResult     `json:"validations"`
	EstimatedCost ResourceCost           `json:"estimated_cost"`
	Warnings      []string               `json:"warnings"`
	GraphQLQuery  string                 `json:"graphql_query,omitempty"`
	AffectedGUIDs []string               `json:"affected_guids,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ProposedChange describes a change that would be made
type ProposedChange struct {
	Type        string      `json:"type"`        // create, update, delete
	Resource    string      `json:"resource"`    // dashboard, alert, entity
	GUID        string      `json:"guid,omitempty"`
	Field       string      `json:"field,omitempty"`
	OldValue    interface{} `json:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value,omitempty"`
	Description string      `json:"description"`
}

// ValidationResult represents a validation check result
type ValidationResult struct {
	Check    string `json:"check"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity"` // error, warning, info
}

// ResourceCost estimates the cost/impact of an operation
type ResourceCost struct {
	ComputeUnits  int     `json:"compute_units,omitempty"`
	StorageGB     float64 `json:"storage_gb,omitempty"`
	DataIngestGB  float64 `json:"data_ingest_gb,omitempty"`
	QueryCredits  int     `json:"query_credits,omitempty"`
	EstimatedTime string  `json:"estimated_time,omitempty"`
}

// DryRunContext provides context for dry-run operations
type DryRunContext struct {
	ctx           context.Context
	currentState  map[string]interface{}
	proposedState map[string]interface{}
	changes       []ProposedChange
	validations   []ValidationResult
	warnings      []string
}

// NewDryRunContext creates a new dry-run context
func NewDryRunContext(ctx context.Context) *DryRunContext {
	return &DryRunContext{
		ctx:           ctx,
		currentState:  make(map[string]interface{}),
		proposedState: make(map[string]interface{}),
		changes:       []ProposedChange{},
		validations:   []ValidationResult{},
		warnings:      []string{},
	}
}

// AddChange records a proposed change
func (d *DryRunContext) AddChange(change ProposedChange) {
	d.changes = append(d.changes, change)
}

// AddValidation records a validation result
func (d *DryRunContext) AddValidation(check string, passed bool, message string, severity string) {
	d.validations = append(d.validations, ValidationResult{
		Check:    check,
		Passed:   passed,
		Message:  message,
		Severity: severity,
	})
}

// AddWarning adds a warning message
func (d *DryRunContext) AddWarning(warning string) {
	d.warnings = append(d.warnings, warning)
}

// BuildResult creates the final dry-run result
func (d *DryRunContext) BuildResult(operation string, wouldSucceed bool) *DryRunResult {
	return &DryRunResult{
		Operation:    operation,
		WouldSucceed: wouldSucceed,
		Changes:      d.changes,
		Validations:  d.validations,
		Warnings:     d.warnings,
		Timestamp:    time.Now(),
	}
}

// DryRunHandler wraps a tool handler to support dry-run mode
func DryRunHandler(handler ToolHandler, dryRunner DryRunner) ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		// Check if dry_run is requested
		dryRun, _ := params["dry_run"].(bool)
		if !dryRun {
			// Normal execution
			return handler(ctx, params)
		}

		// Perform dry-run
		return dryRunner(ctx, params)
	}
}

// DryRunner is a function that performs dry-run validation
type DryRunner func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// Example dry-run implementations

// DryRunDashboardCreate performs dry-run for dashboard creation
func DryRunDashboardCreate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dryRun := NewDryRunContext(ctx)

	// Extract parameters
	name, _ := params["name"].(string)
	pages, _ := params["pages"].([]interface{})
	permissions, _ := params["permissions"].(string)
	accountID, _ := params["account_id"].(float64)

	// Validate name
	if name == "" {
		dryRun.AddValidation("dashboard_name", false, "Dashboard name is required", "error")
	} else if len(name) > 255 {
		dryRun.AddValidation("dashboard_name", false, "Dashboard name exceeds 255 characters", "error")
	} else {
		dryRun.AddValidation("dashboard_name", true, "Dashboard name is valid", "info")
	}

	// Validate pages
	if len(pages) == 0 {
		dryRun.AddValidation("pages", false, "At least one page is required", "error")
	} else if len(pages) > 25 {
		dryRun.AddValidation("pages", false, "Maximum 25 pages allowed", "error")
	} else {
		dryRun.AddValidation("pages", true, fmt.Sprintf("%d pages configured", len(pages)), "info")
		
		// Validate widgets in each page
		totalWidgets := 0
		for i, page := range pages {
			if pageMap, ok := page.(map[string]interface{}); ok {
				widgets, _ := pageMap["widgets"].([]interface{})
				totalWidgets += len(widgets)
				
				if len(widgets) > 150 {
					dryRun.AddValidation(
						fmt.Sprintf("page_%d_widgets", i),
						false,
						fmt.Sprintf("Page %d has %d widgets, maximum is 150", i, len(widgets)),
						"error",
					)
				}
			}
		}
		
		dryRun.AddValidation("total_widgets", true, fmt.Sprintf("Total widgets: %d", totalWidgets), "info")
	}

	// Validate permissions
	validPermissions := []string{"PUBLIC_READ_WRITE", "PUBLIC_READ_ONLY", "PRIVATE"}
	permValid := false
	for _, valid := range validPermissions {
		if permissions == valid {
			permValid = true
			break
		}
	}
	if !permValid {
		dryRun.AddValidation("permissions", false, 
			fmt.Sprintf("Invalid permissions, must be one of: %v", validPermissions), "error")
	} else {
		dryRun.AddValidation("permissions", true, "Permissions are valid", "info")
	}

	// Record proposed changes
	dryRun.AddChange(ProposedChange{
		Type:        "create",
		Resource:    "dashboard",
		Description: fmt.Sprintf("Create dashboard '%s' with %d pages", name, len(pages)),
		NewValue: map[string]interface{}{
			"name":        name,
			"pages":       len(pages),
			"permissions": permissions,
			"accountId":   accountID,
		},
	})

	// Generate sample GraphQL
	graphQL := generateDashboardCreateGraphQL(name, pages, permissions, int(accountID))
	
	// Build result
	result := dryRun.BuildResult("dashboard.create", len(dryRun.validations) > 0)
	result.GraphQLQuery = graphQL
	result.EstimatedCost = ResourceCost{
		ComputeUnits:  1,
		StorageGB:     0.001,
		EstimatedTime: "2-3 seconds",
	}

	// Add warnings for best practices
	if permissions == "PUBLIC_READ_WRITE" {
		result.Warnings = append(result.Warnings, 
			"PUBLIC_READ_WRITE allows anyone to modify this dashboard. Consider using PUBLIC_READ_ONLY or PRIVATE.")
	}

	return result, nil
}

// DryRunAlertCreate performs dry-run for alert creation
func DryRunAlertCreate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dryRun := NewDryRunContext(ctx)

	// Extract parameters
	policyID, _ := params["policy_id"].(string)
	name, _ := params["name"].(string)
	query, _ := params["query"].(string)
	threshold, _ := params["threshold"].(map[string]interface{})

	// Validate policy exists (mock check)
	if policyID == "" {
		dryRun.AddValidation("policy_id", false, "Policy ID is required", "error")
	} else {
		dryRun.AddValidation("policy_id", true, "Policy ID provided", "info")
		// In real implementation, would check if policy exists
		dryRun.AddValidation("policy_exists", true, "Alert policy exists and is accessible", "info")
	}

	// Validate alert name
	if name == "" {
		dryRun.AddValidation("alert_name", false, "Alert name is required", "error")
	} else if len(name) > 128 {
		dryRun.AddValidation("alert_name", false, "Alert name exceeds 128 characters", "error")
	} else {
		dryRun.AddValidation("alert_name", true, "Alert name is valid", "info")
	}

	// Validate NRQL query
	if query == "" {
		dryRun.AddValidation("nrql_query", false, "NRQL query is required", "error")
	} else {
		// Simulate query validation
		dryRun.AddValidation("nrql_syntax", true, "NRQL syntax is valid", "info")
		dryRun.AddValidation("nrql_aggregation", true, "Query returns single numeric value suitable for alerting", "info")
	}

	// Validate threshold
	if threshold == nil {
		dryRun.AddValidation("threshold", false, "Threshold configuration is required", "error")
	} else {
		value, hasValue := threshold["value"].(float64)
		operator, hasOp := threshold["operator"].(string)
		duration, hasDur := threshold["duration_minutes"].(float64)

		if !hasValue || !hasOp || !hasDur {
			dryRun.AddValidation("threshold_config", false, "Threshold must include value, operator, and duration_minutes", "error")
		} else {
			dryRun.AddValidation("threshold_config", true, 
				fmt.Sprintf("Alert when %s %v for %v minutes", operator, value, duration), "info")
		}
	}

	// Record proposed changes
	dryRun.AddChange(ProposedChange{
		Type:        "create",
		Resource:    "alert_condition",
		Description: fmt.Sprintf("Create alert '%s' in policy %s", name, policyID),
		NewValue: map[string]interface{}{
			"name":      name,
			"query":     query,
			"threshold": threshold,
		},
	})

	// Estimate impact
	dryRun.AddWarning("This alert will begin monitoring immediately upon creation")
	dryRun.AddWarning("Ensure notification channels are configured in the alert policy")

	// Build result
	result := dryRun.BuildResult("alert.create", true)
	result.EstimatedCost = ResourceCost{
		QueryCredits:  24, // Assuming hourly execution
		EstimatedTime: "1-2 seconds",
	}

	return result, nil
}

// DryRunBulkTag performs dry-run for bulk tagging operations
func DryRunBulkTag(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dryRun := NewDryRunContext(ctx)

	// Extract parameters
	entityGUIDs, _ := params["entity_guids"].([]interface{})
	tags, _ := params["tags"].(map[string]interface{})
	skipOnError, _ := params["skip_on_error"].(bool)

	// Validate entities
	if len(entityGUIDs) == 0 {
		dryRun.AddValidation("entity_guids", false, "At least one entity GUID is required", "error")
	} else {
		dryRun.AddValidation("entity_guids", true, fmt.Sprintf("%d entities to tag", len(entityGUIDs)), "info")
		
		// Validate GUID format
		invalidGUIDs := 0
		for _, guid := range entityGUIDs {
			if guidStr, ok := guid.(string); ok {
				if len(guidStr) < 10 { // Simple validation
					invalidGUIDs++
				}
			} else {
				invalidGUIDs++
			}
		}
		
		if invalidGUIDs > 0 {
			dryRun.AddValidation("guid_format", false, 
				fmt.Sprintf("%d invalid GUIDs detected", invalidGUIDs), "error")
		}
	}

	// Validate tags
	if len(tags) == 0 {
		dryRun.AddValidation("tags", false, "At least one tag must be provided", "error")
	} else {
		dryRun.AddValidation("tags", true, fmt.Sprintf("%d tags to apply", len(tags)), "info")
		
		// Check for reserved tag keys
		reservedKeys := []string{"account", "accountId", "trustedAccountId"}
		for key := range tags {
			for _, reserved := range reservedKeys {
				if key == reserved {
					dryRun.AddWarning(fmt.Sprintf("Tag key '%s' is reserved and may be ignored", key))
				}
			}
		}
	}

	// Simulate entity validation
	validEntities := len(entityGUIDs)
	invalidGUIDs := 0 // Initialize to avoid undefined error
	if invalidGUIDs > 0 {
		validEntities -= invalidGUIDs
	}

	// Record changes for each entity
	for i, guid := range entityGUIDs {
		if i < 5 { // Limit detailed changes to first 5
			dryRun.AddChange(ProposedChange{
				Type:        "update",
				Resource:    "entity_tags",
				GUID:        fmt.Sprintf("%v", guid),
				Description: fmt.Sprintf("Add %d tags to entity", len(tags)),
				NewValue:    tags,
			})
		}
	}

	if len(entityGUIDs) > 5 {
		dryRun.AddChange(ProposedChange{
			Type:        "update",
			Resource:    "entity_tags",
			Description: fmt.Sprintf("... and %d more entities", len(entityGUIDs)-5),
		})
	}

	// Build result
	result := dryRun.BuildResult("bulk.add_tags", validEntities > 0)
	result.AffectedGUIDs = make([]string, 0, len(entityGUIDs))
	for _, guid := range entityGUIDs {
		if guidStr, ok := guid.(string); ok {
			result.AffectedGUIDs = append(result.AffectedGUIDs, guidStr)
		}
	}

	result.EstimatedCost = ResourceCost{
		ComputeUnits:  len(entityGUIDs),
		EstimatedTime: fmt.Sprintf("%d seconds", len(entityGUIDs)/10+1),
	}

	if skipOnError {
		result.Warnings = append(result.Warnings, "Errors on individual entities will be skipped")
	} else {
		result.Warnings = append(result.Warnings, "Operation will stop on first error")
	}

	return result, nil
}

// Helper function to generate sample GraphQL
func generateDashboardCreateGraphQL(name string, pages []interface{}, permissions string, accountID int) string {
	// Simplified GraphQL generation
	graphql := fmt.Sprintf(`
mutation {
  dashboardCreate(
    accountId: %d
    dashboard: {
      name: "%s"
      permissions: %s
      pages: [
        # %d pages with widgets
      ]
    }
  ) {
    entityResult {
      guid
      name
    }
    errors {
      description
      type
    }
  }
}`, accountID, name, permissions, len(pages))

	return graphql
}
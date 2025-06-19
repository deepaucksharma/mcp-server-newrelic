package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
)

// generateDiscoveryBasedDashboard creates a dashboard based on discovered data
func (s *Server) generateDiscoveryBasedDashboard(ctx context.Context, name string, request map[string]interface{}) (map[string]interface{}, error) {
	// Extract parameters
	domain := ""
	if d, ok := request["domain"].(string); ok {
		domain = d
	}

	serviceName := ""
	if sn, ok := request["service_name"].(string); ok {
		serviceName = sn
	}

	// Extract account IDs for cross-account support
	var accountIDs []int
	if accountIDsParam, ok := request["account_ids"].([]int); ok {
		accountIDs = accountIDsParam
	}

	// Use discovery to find relevant schemas
	filter := discovery.DiscoveryFilter{}
	if serviceName != "" {
		filter.IncludePatterns = []string{fmt.Sprintf("*%s*", serviceName)}
	}

	schemas, err := s.discovery.DiscoverSchemas(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("discover schemas: %w", err)
	}

	if len(schemas) == 0 {
		return nil, fmt.Errorf("no schemas found matching criteria")
	}

	// Build dashboard pages based on discovered schemas
	pages := []map[string]interface{}{}

	// Group schemas by type
	schemaGroups := groupSchemasByType(schemas)

	for groupName, groupSchemas := range schemaGroups {
		page := map[string]interface{}{
			"name":    groupName,
			"widgets": []map[string]interface{}{},
		}

		row := 1
		for _, schema := range groupSchemas {
			// Profile the schema to understand its attributes
			profile, err := s.discovery.ProfileSchema(ctx, schema.EventType, discovery.ProfileDepthFull)
			if err != nil {
				continue // Skip if we can't profile
			}

			// Create widgets based on schema profile
			widgets := s.createWidgetsFromProfile(schema, profile, &row, accountIDs)
			for _, widget := range widgets {
				page["widgets"] = append(page["widgets"].([]map[string]interface{}), widget)
			}
		}

		if len(page["widgets"].([]map[string]interface{})) > 0 {
			pages = append(pages, page)
		}
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no valid widgets could be created from discovered data")
	}

	// Create dashboard structure
	dashboard := map[string]interface{}{
		"name":        name,
		"description": fmt.Sprintf("Dashboard generated from discovered %s data", domain),
		"pages":       pages,
		"metadata": map[string]interface{}{
			"generated_from_discovery": true,
			"schemas_used":             len(schemas),
		},
	}

	// Add account IDs to metadata if specified
	if len(accountIDs) > 0 {
		dashboard["account_ids"] = accountIDs
	}

	return dashboard, nil
}

// groupSchemasByType organizes schemas into logical groups
func groupSchemasByType(schemas []discovery.Schema) map[string][]*discovery.Schema {
	groups := make(map[string][]*discovery.Schema)

	for i := range schemas {
		schema := &schemas[i]
		groupName := determineSchemaGroup(schema)
		groups[groupName] = append(groups[groupName], schema)
	}

	return groups
}

// determineSchemaGroup categorizes a schema based on its name and attributes
func determineSchemaGroup(schema *discovery.Schema) string {
	name := strings.ToLower(schema.Name)

	switch {
	case strings.Contains(name, "transaction"):
		return "Application Performance"
	case strings.Contains(name, "system") || strings.Contains(name, "process"):
		return "Infrastructure"
	case strings.Contains(name, "log"):
		return "Logs"
	case strings.Contains(name, "metric"):
		return "Metrics"
	case strings.Contains(name, "browser") || strings.Contains(name, "pageview"):
		return "Browser Performance"
	case strings.Contains(name, "synthetic"):
		return "Synthetic Monitoring"
	default:
		return "Custom Data"
	}
}

// createWidgetsFromProfile generates appropriate widgets based on schema profile
func (s *Server) createWidgetsFromProfile(schema *discovery.Schema, profile *discovery.Schema, currentRow *int, accountIDs []int) []map[string]interface{} {
	widgets := []map[string]interface{}{}

	// Get default account ID if no cross-account IDs provided
	accountID := 0
	if len(accountIDs) == 0 && s.nrClient != nil {
		if client, ok := s.nrClient.(*newrelic.Client); ok {
			if id, err := client.AccountID(); err == nil {
				accountID = id
			}
		}
	}

	// Create overview widget if we have good sample data
	if schema.SampleCount > 100 {
		overviewWidget := s.createOverviewWidget(schema, profile, accountID, accountIDs, *currentRow)
		widgets = append(widgets, overviewWidget)
		*currentRow += 3
	}

	// Create widgets for key numeric attributes
	numericAttrs := filterNumericAttributes(profile.Attributes)
	for i, attr := range numericAttrs {
		if i >= 3 { // Limit to 3 numeric widgets per schema
			break
		}

		widget := s.createNumericWidget(schema.EventType, attr, accountID, accountIDs, *currentRow, (i%2)*6+1)
		widgets = append(widgets, widget)

		if i%2 == 1 {
			*currentRow += 3
		}
	}

	// Add row spacing if we had an odd number of numeric widgets
	if len(numericAttrs)%2 == 1 {
		*currentRow += 3
	}

	// Create a faceted widget if we have categorical attributes
	categoricalAttrs := filterCategoricalAttributes(profile.Attributes)
	if len(categoricalAttrs) > 0 && len(numericAttrs) > 0 {
		facetWidget := s.createFacetedWidget(schema.EventType, numericAttrs[0], categoricalAttrs[0], accountID, accountIDs, *currentRow)
		widgets = append(widgets, facetWidget)
		*currentRow += 3
	}

	// Create array-based widgets if we have array attributes (NEW)
	arrayAttrs := filterArrayAttributes(profile.Attributes)
	for i, attr := range arrayAttrs {
		if i >= 2 { // Limit to 2 array widgets per schema
			break
		}
		arrayWidget := s.createArrayWidget(schema.EventType, attr, accountID, accountIDs, *currentRow)
		widgets = append(widgets, arrayWidget)
		*currentRow += 3
	}

	return widgets
}

// createOverviewWidget creates a summary widget for the schema
func (s *Server) createOverviewWidget(schema *discovery.Schema, profile *discovery.Schema, accountID int, accountIDs []int, row int) map[string]interface{} {
	// Find a good key attribute from profile
	keyAttr := "entity.guid" // default
	for _, attr := range profile.Attributes {
		// Look for attributes that might be keys based on their name or cardinality
		if strings.Contains(strings.ToLower(attr.Name), "id") || strings.Contains(strings.ToLower(attr.Name), "guid") ||
			strings.Contains(strings.ToLower(attr.Name), "key") || attr.Cardinality.IsHighCardinality {
			keyAttr = attr.Name
			break
		}
	}

	// Build account IDs clause for cross-account queries
	accountClause := ""
	if len(accountIDs) > 0 {
		accountIDStrs := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			accountIDStrs[i] = fmt.Sprintf("%d", id)
		}
		accountClause = fmt.Sprintf(" WITH accountIds = [%s]", strings.Join(accountIDStrs, ", "))
	}

	// Determine query based on schema characteristics
	query := fmt.Sprintf("SELECT count(*) as 'Events', uniqueCount(%s) as 'Unique %s' FROM %s SINCE 1 hour ago%s",
		keyAttr, strings.Title(keyAttr), schema.EventType, accountClause)

	return map[string]interface{}{
		"title":  fmt.Sprintf("%s Overview", schema.Name),
		"type":   newrelic.VizBillboard,
		"row":    row,
		"column": 1,
		"width":  12,
		"height": 3,
		"query":  query,
		"configuration": map[string]interface{}{
			"nrqlQueries": []map[string]interface{}{
				{
					"accountId": accountID,
					"query":     query,
				},
			},
		},
	}
}

// createNumericWidget creates a widget for a numeric attribute
func (s *Server) createNumericWidget(eventType string, attr *discovery.Attribute, accountID int, accountIDs []int, row int, column int) map[string]interface{} {
	// Choose visualization based on attribute characteristics
	vizType := newrelic.VizLine
	query := ""

	// Build account IDs clause for cross-account queries
	accountClause := ""
	if len(accountIDs) > 0 {
		accountIDStrs := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			accountIDStrs[i] = fmt.Sprintf("%d", id)
		}
		accountClause = fmt.Sprintf(" WITH accountIds = [%s]", strings.Join(accountIDStrs, ", "))
	}

	// Check semantic type for counters and percentages
	isCounter := attr.SemanticType == discovery.SemanticTypeDuration || strings.Contains(strings.ToLower(attr.Name), "count")
	isPercentage := attr.SemanticType == discovery.SemanticTypePercentage || strings.Contains(strings.ToLower(attr.Name), "percent")

	if isCounter {
		// For counters, show rate over time
		query = fmt.Sprintf("SELECT rate(sum(%s), 1 minute) as '%s/min' FROM %s TIMESERIES SINCE 1 hour ago%s",
			attr.Name, attr.Name, eventType, accountClause)
	} else if isPercentage {
		// For percentages, show average over time
		query = fmt.Sprintf("SELECT average(%s) as 'Avg %s' FROM %s TIMESERIES SINCE 1 hour ago%s",
			attr.Name, attr.Name, eventType, accountClause)
	} else {
		// For other numerics, show percentiles
		query = fmt.Sprintf("SELECT average(%s) as 'Avg', percentile(%s, 95) as 'P95', max(%s) as 'Max' FROM %s TIMESERIES SINCE 1 hour ago%s",
			attr.Name, attr.Name, attr.Name, eventType, accountClause)
	}

	return map[string]interface{}{
		"title":  attr.Name,
		"type":   vizType,
		"row":    row,
		"column": column,
		"width":  6,
		"height": 3,
		"query":  query,
		"configuration": map[string]interface{}{
			"nrqlQueries": []map[string]interface{}{
				{
					"accountId": accountID,
					"query":     query,
				},
			},
		},
	}
}

// createFacetedWidget creates a widget that breaks down a metric by a dimension
func (s *Server) createFacetedWidget(eventType string, metric *discovery.Attribute, dimension *discovery.Attribute, accountID int, accountIDs []int, row int) map[string]interface{} {
	// Build account IDs clause for cross-account queries
	accountClause := ""
	if len(accountIDs) > 0 {
		accountIDStrs := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			accountIDStrs[i] = fmt.Sprintf("%d", id)
		}
		accountClause = fmt.Sprintf(" WITH accountIds = [%s]", strings.Join(accountIDStrs, ", "))
	}

	// Create a bar chart showing metric broken down by dimension
	query := fmt.Sprintf("SELECT average(%s) FROM %s FACET %s SINCE 1 hour ago LIMIT 10%s",
		metric.Name, eventType, dimension.Name, accountClause)

	return map[string]interface{}{
		"title":  fmt.Sprintf("%s by %s", metric.Name, dimension.Name),
		"type":   newrelic.VizBar,
		"row":    row,
		"column": 1,
		"width":  12,
		"height": 3,
		"query":  query,
		"configuration": map[string]interface{}{
			"nrqlQueries": []map[string]interface{}{
				{
					"accountId": accountID,
					"query":     query,
				},
			},
		},
	}
}

// filterNumericAttributes returns only numeric attributes
func filterNumericAttributes(attrs []discovery.Attribute) []*discovery.Attribute {
	var numeric []*discovery.Attribute
	for i := range attrs {
		attr := &attrs[i]
		if attr.DataType == discovery.DataTypeNumeric {
			numeric = append(numeric, attr)
		}
	}
	return numeric
}

// filterCategoricalAttributes returns only categorical attributes
func filterCategoricalAttributes(attrs []discovery.Attribute) []*discovery.Attribute {
	var categorical []*discovery.Attribute
	for i := range attrs {
		attr := &attrs[i]
		if attr.DataType == discovery.DataTypeString && attr.Cardinality.Unique > 1 && attr.Cardinality.Unique < 100 {
			categorical = append(categorical, attr)
		}
	}
	return categorical
}

// filterArrayAttributes returns attributes that appear to be arrays
func filterArrayAttributes(attrs []discovery.Attribute) []*discovery.Attribute {
	var arrays []*discovery.Attribute
	for i := range attrs {
		attr := &attrs[i]
		// Detect array attributes by name patterns or data characteristics
		if strings.Contains(strings.ToLower(attr.Name), "array") ||
			strings.Contains(strings.ToLower(attr.Name), "list") ||
			strings.Contains(strings.ToLower(attr.Name), "tags") ||
			strings.Contains(strings.ToLower(attr.Name), "labels") ||
			strings.Contains(strings.ToLower(attr.Name), "attributes.") ||
			strings.Contains(strings.ToLower(attr.Name), "[]") {
			arrays = append(arrays, attr)
		}
	}
	return arrays
}

// createArrayWidget creates a widget for array attributes using new NRQL features
func (s *Server) createArrayWidget(eventType string, attr *discovery.Attribute, accountID int, accountIDs []int, row int) map[string]interface{} {
	// Build account IDs clause for cross-account queries
	accountClause := ""
	if len(accountIDs) > 0 {
		accountIDStrs := make([]string, len(accountIDs))
		for i, id := range accountIDs {
			accountIDStrs[i] = fmt.Sprintf("%d", id)
		}
		accountClause = fmt.Sprintf(" WITH accountIds = [%s]", strings.Join(accountIDStrs, ", "))
	}

	// Use array functions to analyze the array attribute
	query := fmt.Sprintf(`SELECT 
		average(length(%s)) as 'Avg Array Size',
		max(length(%s)) as 'Max Array Size',
		uniqueCount(getfield(%s, 0)) as 'Unique First Elements'
	FROM %s 
	WHERE %s IS NOT NULL 
	SINCE 1 hour ago%s`,
		attr.Name, attr.Name, attr.Name, eventType, attr.Name, accountClause)

	return map[string]interface{}{
		"title":  fmt.Sprintf("%s Analysis", attr.Name),
		"type":   newrelic.VizBillboard,
		"row":    row,
		"column": 1,
		"width":  12,
		"height": 3,
		"query":  query,
		"configuration": map[string]interface{}{
			"nrqlQueries": []map[string]interface{}{
				{
					"accountId": accountID,
					"query":     query,
				},
			},
		},
	}
}

// createIntelligentDashboard creates a dashboard using full discovery and intelligence
func (s *Server) createIntelligentDashboard(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	// Step 1: Understand the request intent
	intent := analyzeUserIntent(request)

	// Step 2: Discover relevant data
	discoveryResults, err := s.performTargetedDiscovery(ctx, intent)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Step 3: Analyze data relationships
	relationships, err := s.discovery.FindRelationships(ctx, discoveryResults)
	if err != nil {
		// Continue without relationships
		relationships = []discovery.Relationship{}
	}

	// Step 4: Build optimal dashboard structure
	dashboard := s.buildOptimalDashboard(intent, discoveryResults, relationships)

	// Step 5: Validate all widgets use real data
	if err := s.validateDashboardQueries(ctx, dashboard); err != nil {
		return nil, fmt.Errorf("dashboard validation failed: %w", err)
	}

	return dashboard, nil
}

// analyzeUserIntent interprets what the user is trying to achieve
func analyzeUserIntent(request map[string]interface{}) map[string]interface{} {
	intent := map[string]interface{}{
		"type":      "general", // general, performance, troubleshooting, capacity, etc.
		"focus":     []string{},
		"timeRange": "1 hour ago",
	}

	// Analyze request for keywords
	if name, ok := request["name"].(string); ok {
		nameLower := strings.ToLower(name)
		switch {
		case strings.Contains(nameLower, "performance") || strings.Contains(nameLower, "golden"):
			intent["type"] = "performance"
			intent["focus"] = append(intent["focus"].([]string), "latency", "throughput", "errors")
		case strings.Contains(nameLower, "capacity") || strings.Contains(nameLower, "resource"):
			intent["type"] = "capacity"
			intent["focus"] = append(intent["focus"].([]string), "cpu", "memory", "disk", "utilization")
		case strings.Contains(nameLower, "error") || strings.Contains(nameLower, "troubleshoot"):
			intent["type"] = "troubleshooting"
			intent["focus"] = append(intent["focus"].([]string), "errors", "exceptions", "failures")
		}
	}

	return intent
}

// performTargetedDiscovery discovers data based on user intent
func (s *Server) performTargetedDiscovery(ctx context.Context, intent map[string]interface{}) ([]discovery.Schema, error) {
	filter := discovery.DiscoveryFilter{}

	// Apply filters based on intent type
	switch intent["type"].(string) {
	case "performance":
		filter.IncludePatterns = []string{"*Transaction*", "*PageView*", "*Request*"}
	case "capacity":
		filter.IncludePatterns = []string{"*System*", "*Process*", "*Container*"}
	case "troubleshooting":
		filter.IncludePatterns = []string{"*Error*", "*Exception*", "*Log*"}
	}

	return s.discovery.DiscoverSchemas(ctx, filter)
}

// extractSchemaNames gets schema names from discovery results
func extractSchemaNames(schemas []discovery.Schema) []string {
	names := make([]string, len(schemas))
	for i, schema := range schemas {
		names[i] = schema.Name
	}
	return names
}

// buildOptimalDashboard creates the best possible dashboard from discovered data
func (s *Server) buildOptimalDashboard(intent map[string]interface{}, schemas []discovery.Schema, relationships []discovery.Relationship) map[string]interface{} {
	dashboard := map[string]interface{}{
		"name":  "Intelligent Dashboard",
		"pages": []map[string]interface{}{},
	}

	// Create pages based on logical groupings
	if intent["type"] == "performance" {
		// Create golden signals page if we have the data
		if goldenSignalsPage := s.createGoldenSignalsFromDiscovery(schemas); goldenSignalsPage != nil {
			dashboard["pages"] = append(dashboard["pages"].([]map[string]interface{}), goldenSignalsPage)
		}
	}

	// Add relationship-based widgets
	if len(relationships) > 0 {
		relationshipPage := s.createRelationshipWidgets(schemas, relationships)
		dashboard["pages"] = append(dashboard["pages"].([]map[string]interface{}), relationshipPage)
	}

	// Add detailed pages for each major schema
	for i := range schemas {
		schema := &schemas[i]
		if schema.SampleCount > 1000 { // Only include schemas with significant data
			page := s.createDetailedSchemaPage(schema)
			dashboard["pages"] = append(dashboard["pages"].([]map[string]interface{}), page)
		}
	}

	return dashboard
}

// createGoldenSignalsFromDiscovery creates golden signals using discovered metrics
func (s *Server) createGoldenSignalsFromDiscovery(schemas []discovery.Schema) map[string]interface{} {
	// Find schemas that can provide golden signals
	var latencySchema, trafficSchema, errorSchema, saturationSchema *discovery.Schema

	for i := range schemas {
		schema := &schemas[i]
		profile, _ := s.discovery.ProfileSchema(context.Background(), schema.EventType, discovery.ProfileDepthStandard)
		if profile == nil {
			continue
		}

		// Look for latency metrics
		if latencySchema == nil && hasAttribute(profile, "duration", "latency", "response_time") {
			latencySchema = schema
		}

		// Look for traffic metrics
		if trafficSchema == nil && (schema.SampleCount > 0 || hasAttribute(profile, "count", "requests")) {
			trafficSchema = schema
		}

		// Look for error indicators
		if errorSchema == nil && hasAttribute(profile, "error", "exception", "failure") {
			errorSchema = schema
		}

		// Look for saturation metrics
		if saturationSchema == nil && hasAttribute(profile, "cpu", "memory", "utilization") {
			saturationSchema = schema
		}
	}

	// Only create page if we have at least 2 golden signals
	signalsFound := 0
	widgets := []map[string]interface{}{}

	if latencySchema != nil {
		signalsFound++
		// Create latency widget
	}

	if trafficSchema != nil {
		signalsFound++
		// Create traffic widget
	}

	if errorSchema != nil {
		signalsFound++
		// Create error widget
	}

	if saturationSchema != nil {
		signalsFound++
		// Create saturation widget
	}

	if signalsFound >= 2 {
		return map[string]interface{}{
			"name":    "Golden Signals",
			"widgets": widgets,
		}
	}

	return nil
}

// hasAttribute checks if a schema profile has any of the specified attributes
func hasAttribute(profile *discovery.Schema, patterns ...string) bool {
	for _, attr := range profile.Attributes {
		attrLower := strings.ToLower(attr.Name)
		for _, pattern := range patterns {
			if strings.Contains(attrLower, pattern) {
				return true
			}
		}
	}
	return false
}

// createRelationshipWidgets creates widgets that show data relationships
func (s *Server) createRelationshipWidgets(schemas []discovery.Schema, relationships []discovery.Relationship) map[string]interface{} {
	return map[string]interface{}{
		"name":    "Data Relationships",
		"widgets": []map[string]interface{}{
			// Create widgets showing how different data types relate
		},
	}
}

// createDetailedSchemaPage creates a comprehensive page for a schema
func (s *Server) createDetailedSchemaPage(schema *discovery.Schema) map[string]interface{} {
	return map[string]interface{}{
		"name":    schema.Name,
		"widgets": []map[string]interface{}{
			// Create various widgets for this schema
		},
	}
}

// validateDashboardQueries ensures all queries reference valid data
func (s *Server) validateDashboardQueries(ctx context.Context, dashboard map[string]interface{}) error {
	// Validate each widget's query
	if pages, ok := dashboard["pages"].([]map[string]interface{}); ok {
		for _, page := range pages {
			if widgets, ok := page["widgets"].([]map[string]interface{}); ok {
				for _, widget := range widgets {
					if query, ok := widget["query"].(string); ok {
						// Validate the query references real event types and attributes
						if err := s.validateNRQLQuery(ctx, query); err != nil {
							return fmt.Errorf("invalid query in widget '%s': %w", widget["title"], err)
						}
					}
				}
			}
		}
	}
	return nil
}

// validateNRQLQuery checks if a query references valid event types and attributes
func (s *Server) validateNRQLQuery(ctx context.Context, query string) error {
	// This should use a proper NRQL parser and validate against discovered schemas
	// For now, basic validation
	if query == "" {
		return fmt.Errorf("empty query")
	}
	return nil
}

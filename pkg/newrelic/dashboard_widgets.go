package newrelic

import (
	"encoding/json"
	"fmt"
)

// Widget types supported by New Relic
const (
	VizArea       = "viz.area"
	VizBar        = "viz.bar"
	VizBillboard  = "viz.billboard"
	VizLine       = "viz.line"
	VizMarkdown   = "viz.markdown"
	VizPie        = "viz.pie"
	VizTable      = "viz.table"
	VizBullet     = "viz.bullet"
	VizEventFeed  = "viz.event-feed"
	VizFunnel     = "viz.funnel"
	VizHeatmap    = "viz.heatmap"
	VizHistogram  = "viz.histogram"
	VizJSON       = "viz.json"
	VizServiceMap = "topology.service-map"
	VizInventory  = "infra.inventory"
)

// AlertSeverity for thresholds
type AlertSeverity string

const (
	AlertSeverityNotAlerting AlertSeverity = "NOT_ALERTING"
	AlertSeverityWarning     AlertSeverity = "WARNING"
	AlertSeverityCritical    AlertSeverity = "CRITICAL"
)

// NRQLQuery represents a NRQL query configuration
type NRQLQuery struct {
	AccountID int    `json:"accountId"`
	Query     string `json:"query"`
}

// Threshold represents a threshold configuration for billboards
type Threshold struct {
	AlertSeverity AlertSeverity `json:"alertSeverity"`
	Value         float64       `json:"value"`
}

// WidgetConfiguration represents the typed configuration for widgets
type WidgetConfiguration struct {
	Area      *AreaConfiguration      `json:"area,omitempty"`
	Bar       *BarConfiguration       `json:"bar,omitempty"`
	Billboard *BillboardConfiguration `json:"billboard,omitempty"`
	Line      *LineConfiguration      `json:"line,omitempty"`
	Markdown  *MarkdownConfiguration  `json:"markdown,omitempty"`
	Pie       *PieConfiguration       `json:"pie,omitempty"`
	Table     *TableConfiguration     `json:"table,omitempty"`
}

// Typed widget configurations
type AreaConfiguration struct {
	NRQLQueries []NRQLQuery `json:"nrqlQueries"`
}

type BarConfiguration struct {
	NRQLQueries []NRQLQuery `json:"nrqlQueries"`
}

type BillboardConfiguration struct {
	NRQLQueries []NRQLQuery `json:"nrqlQueries"`
	Thresholds  []Threshold `json:"thresholds,omitempty"`
}

type LineConfiguration struct {
	NRQLQueries []NRQLQuery `json:"nrqlQueries"`
}

type MarkdownConfiguration struct {
	Text string `json:"text"`
}

type PieConfiguration struct {
	NRQLQueries []NRQLQuery `json:"nrqlQueries"`
}

type TableConfiguration struct {
	NRQLQueries []NRQLQuery `json:"nrqlQueries"`
}

// ConvertWidgetToGraphQLInput converts a dashboard widget to the proper GraphQL input format
func ConvertWidgetToGraphQLInput(widget DashboardWidget, accountID int) (map[string]interface{}, error) {
	widgetInput := map[string]interface{}{
		"title": widget.Title,
	}

	// Handle typed widgets
	switch widget.Type {
	case VizArea, VizBar, VizBillboard, VizLine, VizPie, VizTable:
		// For typed widgets, use configuration
		config := map[string]interface{}{}
		
		// Extract NRQL queries from configuration
		if nrqlQueries, ok := widget.Configuration["nrqlQueries"].([]interface{}); ok {
			queries := []map[string]interface{}{}
			for _, q := range nrqlQueries {
				if qMap, ok := q.(map[string]interface{}); ok {
					// Ensure accountId is set
					if _, hasAccount := qMap["accountId"]; !hasAccount {
						qMap["accountId"] = accountID
					}
					queries = append(queries, qMap)
				}
			}
			config["nrqlQueries"] = queries
		}

		// Handle billboard thresholds
		if widget.Type == VizBillboard {
			if thresholds, ok := widget.Configuration["thresholds"].([]interface{}); ok {
				config["thresholds"] = thresholds
			}
		}

		// Handle markdown text
		if widget.Type == VizMarkdown {
			if text, ok := widget.Configuration["text"].(string); ok {
				config["text"] = text
			}
		}

		// Set the appropriate configuration based on type
		switch widget.Type {
		case VizArea:
			widgetInput["configuration"] = map[string]interface{}{"area": config}
		case VizBar:
			widgetInput["configuration"] = map[string]interface{}{"bar": config}
		case VizBillboard:
			widgetInput["configuration"] = map[string]interface{}{"billboard": config}
		case VizLine:
			widgetInput["configuration"] = map[string]interface{}{"line": config}
		case VizMarkdown:
			widgetInput["configuration"] = map[string]interface{}{"markdown": config}
		case VizPie:
			widgetInput["configuration"] = map[string]interface{}{"pie": config}
		case VizTable:
			widgetInput["configuration"] = map[string]interface{}{"table": config}
		}

	default:
		// For untyped widgets, use rawConfiguration
		widgetInput["visualization"] = widget.Type
		
		// Build raw configuration
		rawConfig := make(map[string]interface{})
		
		// Copy all configuration data
		for k, v := range widget.Configuration {
			rawConfig[k] = v
		}
		
		// Ensure NRQL queries have accountId for untyped widgets
		if nrqlQueries, ok := rawConfig["nrqlQueries"].([]interface{}); ok {
			queries := []map[string]interface{}{}
			for _, q := range nrqlQueries {
				if qMap, ok := q.(map[string]interface{}); ok {
					if _, hasAccount := qMap["accountId"]; !hasAccount {
						qMap["accountId"] = accountID
					}
					queries = append(queries, qMap)
				}
			}
			rawConfig["nrqlQueries"] = queries
		}
		
		// Handle specific untyped widget configurations
		switch widget.Type {
		case VizBullet:
			// Ensure limit is set for bullet charts
			if _, hasLimit := rawConfig["limit"]; !hasLimit {
				return nil, fmt.Errorf("bullet chart requires 'limit' parameter")
			}
			
		case VizFunnel:
			// Validate funnel query contains funnel() function
			if queries, ok := rawConfig["nrqlQueries"].([]map[string]interface{}); ok && len(queries) > 0 {
				if query, ok := queries[0]["query"].(string); ok {
					if !containsFunnelFunction(query) {
						return nil, fmt.Errorf("funnel widget requires NRQL query with funnel() function")
					}
				}
			}
			
		case VizHeatmap:
			// Validate heatmap query contains histogram() function
			if queries, ok := rawConfig["nrqlQueries"].([]map[string]interface{}); ok && len(queries) > 0 {
				if query, ok := queries[0]["query"].(string); ok {
					if !containsHistogramFunction(query) {
						return nil, fmt.Errorf("heatmap widget requires NRQL query with histogram() function")
					}
				}
			}
			
		case VizServiceMap:
			// Validate required fields for service map
			if _, hasEntities := rawConfig["primaryEntities"]; !hasEntities {
				return nil, fmt.Errorf("service map requires 'primaryEntities'")
			}
			
		case VizInventory:
			// Validate required fields for inventory
			if _, hasSources := rawConfig["sources"]; !hasSources {
				return nil, fmt.Errorf("inventory widget requires 'sources'")
			}
			if _, hasAccountId := rawConfig["accountId"]; !hasAccountId {
				rawConfig["accountId"] = accountID
			}
		}
		
		widgetInput["rawConfiguration"] = rawConfig
	}

	// Add layout information
	if layout, ok := widget.Configuration["layout"].(map[string]interface{}); ok {
		widgetInput["layout"] = layout
	} else if row, ok := widget.Configuration["row"]; ok {
		// Build layout from individual properties
		widgetInput["layout"] = map[string]interface{}{
			"row":    row,
			"column": widget.Configuration["column"],
			"width":  widget.Configuration["width"],
			"height": widget.Configuration["height"],
		}
	}

	// Add linked entities if present
	if linkedEntities, ok := widget.Configuration["linkedEntities"].([]interface{}); ok {
		widgetInput["linkedEntities"] = linkedEntities
	}

	return widgetInput, nil
}

// Helper functions
func containsFunnelFunction(query string) bool {
	// Simple check - in production, use proper NRQL parser
	return contains(query, "funnel(") || contains(query, "FUNNEL(")
}

func containsHistogramFunction(query string) bool {
	// Simple check - in production, use proper NRQL parser
	return contains(query, "histogram(") || contains(query, "HISTOGRAM(")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsAt(s, substr)
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BuildWidgetFromDiscovery creates a widget configuration based on discovered data
func BuildWidgetFromDiscovery(
	title string,
	eventType string,
	attributes []string,
	dataProfile map[string]interface{},
	accountID int,
) DashboardWidget {
	widget := DashboardWidget{
		Title:         title,
		Configuration: make(map[string]interface{}),
	}

	// Determine visualization type based on data characteristics
	vizType := determineVisualizationType(dataProfile)
	widget.Type = vizType

	// Build NRQL query based on discovered attributes
	query := buildDiscoveryQuery(eventType, attributes, dataProfile)

	// Configure based on visualization type
	switch vizType {
	case VizBillboard:
		widget.Configuration["nrqlQueries"] = []NRQLQuery{{
			AccountID: accountID,
			Query:     query,
		}}
		// Add thresholds if we have baseline data
		if thresholds := buildThresholds(dataProfile); thresholds != nil {
			widget.Configuration["thresholds"] = thresholds
		}

	case VizLine, VizArea:
		// Ensure query has TIMESERIES
		if !contains(query, "TIMESERIES") {
			query += " TIMESERIES"
		}
		widget.Configuration["nrqlQueries"] = []NRQLQuery{{
			AccountID: accountID,
			Query:     query,
		}}

	case VizBar, VizPie:
		// Ensure query has FACET for grouping
		if !contains(query, "FACET") && len(attributes) > 1 {
			query += fmt.Sprintf(" FACET %s", attributes[1])
		}
		widget.Configuration["nrqlQueries"] = []NRQLQuery{{
			AccountID: accountID,
			Query:     query,
		}}

	case VizTable:
		widget.Configuration["nrqlQueries"] = []NRQLQuery{{
			AccountID: accountID,
			Query:     query,
		}}

	case VizHistogram:
		// Modify query to use histogram function
		if !contains(query, "histogram(") {
			query = fmt.Sprintf("SELECT histogram(%s) FROM %s", attributes[0], eventType)
		}
		widget.Configuration = map[string]interface{}{
			"nrqlQueries": []map[string]interface{}{{
				"accountId": accountID,
				"query":     query,
			}},
		}

	default:
		// Default to table
		widget.Type = VizTable
		widget.Configuration["nrqlQueries"] = []NRQLQuery{{
			AccountID: accountID,
			Query:     query,
		}}
	}

	return widget
}

// determineVisualizationType selects the best visualization based on data profile
func determineVisualizationType(dataProfile map[string]interface{}) string {
	// Extract characteristics from data profile
	dataType := dataProfile["type"].(string)
	cardinality := dataProfile["cardinality"].(int)
	hasTimeSeries := dataProfile["hasTimeSeries"].(bool)
	distribution := dataProfile["distribution"].(string)

	// Decision tree for visualization selection
	switch {
	case dataType == "numeric" && cardinality == 1:
		return VizBillboard
	case dataType == "numeric" && hasTimeSeries:
		return VizLine
	case dataType == "numeric" && distribution == "normal":
		return VizHistogram
	case dataType == "categorical" && cardinality <= 10:
		return VizPie
	case dataType == "categorical" && cardinality > 10:
		return VizBar
	case dataType == "mixed":
		return VizTable
	default:
		return VizTable
	}
}

// buildDiscoveryQuery creates a NRQL query from discovered attributes
func buildDiscoveryQuery(eventType string, attributes []string, dataProfile map[string]interface{}) string {
	if len(attributes) == 0 {
		return fmt.Sprintf("SELECT count(*) FROM %s", eventType)
	}

	// Select appropriate aggregation based on data type
	dataType := dataProfile["type"].(string)
	primaryAttr := attributes[0]

	var query string
	switch dataType {
	case "numeric":
		// Use average for numeric data
		query = fmt.Sprintf("SELECT average(%s) FROM %s", primaryAttr, eventType)
	case "categorical":
		// Use count with facet for categorical
		query = fmt.Sprintf("SELECT count(*) FROM %s FACET %s", eventType, primaryAttr)
	default:
		// Default to count
		query = fmt.Sprintf("SELECT count(*) FROM %s", eventType)
	}

	// Add time range
	query += " SINCE 1 hour ago"

	return query
}

// buildThresholds creates threshold configuration based on data profile
func buildThresholds(dataProfile map[string]interface{}) []Threshold {
	if baseline, ok := dataProfile["baseline"].(map[string]interface{}); ok {
		avg := baseline["average"].(float64)
		stddev := baseline["stddev"].(float64)

		return []Threshold{
			{
				AlertSeverity: AlertSeverityCritical,
				Value:         avg + (2 * stddev),
			},
			{
				AlertSeverity: AlertSeverityWarning,
				Value:         avg + stddev,
			},
		}
	}
	return nil
}

// ValidateWidgetConfiguration validates widget configuration based on type
func ValidateWidgetConfiguration(widget DashboardWidget) error {
	switch widget.Type {
	case VizArea, VizBar, VizLine, VizPie, VizTable:
		// Typed widgets must have nrqlQueries
		if queries, ok := widget.Configuration["nrqlQueries"].([]interface{}); !ok || len(queries) == 0 {
			return fmt.Errorf("%s widget requires at least one NRQL query", widget.Type)
		}

	case VizBillboard:
		if queries, ok := widget.Configuration["nrqlQueries"].([]interface{}); !ok || len(queries) == 0 {
			return fmt.Errorf("billboard widget requires at least one NRQL query")
		}

	case VizMarkdown:
		if text, ok := widget.Configuration["text"].(string); !ok || text == "" {
			return fmt.Errorf("markdown widget requires non-empty text")
		}

	case VizBullet:
		if _, hasLimit := widget.Configuration["limit"]; !hasLimit {
			return fmt.Errorf("bullet widget requires limit parameter")
		}

	case VizFunnel:
		// Validate funnel query
		if queries, ok := widget.Configuration["nrqlQueries"].([]interface{}); ok && len(queries) > 0 {
			if qMap, ok := queries[0].(map[string]interface{}); ok {
				if query, ok := qMap["query"].(string); ok && !containsFunnelFunction(query) {
					return fmt.Errorf("funnel widget requires NRQL query with funnel() function")
				}
			}
		}

	case VizHeatmap, VizHistogram:
		// Validate histogram query
		if queries, ok := widget.Configuration["nrqlQueries"].([]interface{}); ok && len(queries) > 0 {
			if qMap, ok := queries[0].(map[string]interface{}); ok {
				if query, ok := qMap["query"].(string); ok && !containsHistogramFunction(query) {
					return fmt.Errorf("%s widget requires NRQL query with histogram() function", widget.Type)
				}
			}
		}

	case VizServiceMap:
		if _, hasEntities := widget.Configuration["primaryEntities"]; !hasEntities {
			return fmt.Errorf("service map widget requires primaryEntities")
		}

	case VizInventory:
		if _, hasSources := widget.Configuration["sources"]; !hasSources {
			return fmt.Errorf("inventory widget requires sources")
		}
	}

	// Validate layout if present
	if layout, ok := widget.Configuration["layout"].(map[string]interface{}); ok {
		if col, ok := layout["column"].(float64); !ok || col < 1 || col > 12 {
			return fmt.Errorf("invalid column value: must be between 1 and 12")
		}
		if width, ok := layout["width"].(float64); ok {
			if col, _ := layout["column"].(float64); col+width > 12 {
				return fmt.Errorf("widget extends beyond dashboard width (column + width > 12)")
			}
		}
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for DashboardWidget
func (w DashboardWidget) MarshalJSON() ([]byte, error) {
	// Create a map for JSON representation
	m := make(map[string]interface{})
	
	m["title"] = w.Title
	m["visualization"] = w.Type
	
	// Handle configuration based on widget type
	isTypedWidget := false
	switch w.Type {
	case VizArea, VizBar, VizBillboard, VizLine, VizMarkdown, VizPie, VizTable:
		isTypedWidget = true
	}
	
	if isTypedWidget {
		m["configuration"] = w.Configuration
		// Also include rawConfiguration for compatibility
		m["rawConfiguration"] = w.Configuration
	} else {
		m["configuration"] = nil
		m["rawConfiguration"] = w.Configuration
	}
	
	// Extract layout from configuration if present
	if layout, ok := w.Configuration["layout"].(map[string]interface{}); ok {
		m["layout"] = layout
	}
	
	// Extract linkedEntities if present
	if entities, ok := w.Configuration["linkedEntities"].([]interface{}); ok {
		m["linkedEntities"] = entities
	}
	
	return json.Marshal(m)
}
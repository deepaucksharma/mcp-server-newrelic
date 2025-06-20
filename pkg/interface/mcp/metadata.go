package mcp

import (
	"fmt"
	"strings"
)

// ToolCategory represents the category of tool operation
type ToolCategory string

const (
	CategoryQuery    ToolCategory = "query"
	CategoryMutation ToolCategory = "mutation"
	CategoryAnalysis ToolCategory = "analysis"
	CategoryUtility  ToolCategory = "utility"
	CategoryBulk     ToolCategory = "bulk"
)

// SafetyLevel represents the safety level of an operation
type SafetyLevel string

const (
	SafetyLevelSafe        SafetyLevel = "safe"
	SafetyLevelCaution     SafetyLevel = "caution"
	SafetyLevelDestructive SafetyLevel = "destructive"
)

// EnhancedTool extends the base Tool with rich metadata
type EnhancedTool struct {
	Tool
	Category      ToolCategory           `json:"category"`
	Safety        SafetyMetadata         `json:"safety"`
	Performance   PerformanceMetadata    `json:"performance"`
	AIGuidance    AIGuidanceMetadata     `json:"ai_guidance"`
	Observability ObservabilityMetadata  `json:"observability"`
	Examples      []ToolExample          `json:"examples"`
}

// SafetyMetadata contains safety-related information for a tool
type SafetyMetadata struct {
	Level                SafetyLevel `json:"level"`
	IsDestructive        bool        `json:"is_destructive"`
	RequiresConfirmation bool        `json:"requires_confirmation"`
	DryRunSupported      bool        `json:"dry_run_supported"`
	AffectedResources    []string    `json:"affected_resources"`
	RollbackSupported    bool        `json:"rollback_supported"`
	PermissionsRequired  []string    `json:"permissions_required"`
}

// PerformanceMetadata contains performance characteristics
type PerformanceMetadata struct {
	ExpectedLatencyMS   int    `json:"expected_latency_ms"`
	MaxLatencyMS        int    `json:"max_latency_ms"`
	MaxResultSizeBytes  int    `json:"max_result_size_bytes"`
	RateLimitPerMinute  int    `json:"rate_limit_per_minute"`
	Cacheable           bool   `json:"cacheable"`
	CacheTTLSeconds     int    `json:"cache_ttl_seconds"`
	ResourceIntensive   bool   `json:"resource_intensive"`
	CostCategory        string `json:"cost_category"` // low, medium, high
}

// AIGuidanceMetadata provides guidance for AI agents
type AIGuidanceMetadata struct {
	UsageExamples       []string          `json:"usage_examples"`
	CommonPatterns      []string          `json:"common_patterns"`
	PreferredOver       []string          `json:"preferred_over"`
	ChainsWith          []string          `json:"chains_with"`
	RequiredTools       []string          `json:"required_tools"`
	WarningsForAI       []string          `json:"warnings_for_ai"`
	SuccessIndicators   []string          `json:"success_indicators"`
	ErrorPatterns       map[string]string `json:"error_patterns"` // error pattern -> suggested action
	ContextRequirements []string          `json:"context_requirements"`
}

// ObservabilityMetadata defines how the tool should be monitored
type ObservabilityMetadata struct {
	MetricsEnabled     bool              `json:"metrics_enabled"`
	TracingEnabled     bool              `json:"tracing_enabled"`
	LogLevel           string            `json:"log_level"`
	CustomMetrics      []string          `json:"custom_metrics"`
	SLOTargetMS        int               `json:"slo_target_ms"`
	ErrorThresholdRate float64           `json:"error_threshold_rate"`
	AlertOnFailure     bool              `json:"alert_on_failure"`
	AuditFields        []string          `json:"audit_fields"`
	SensitiveFields    []string          `json:"sensitive_fields"` // fields to redact in logs
}

// ToolExample provides concrete usage examples
type ToolExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Params      map[string]interface{} `json:"params"`
	Result      interface{}            `json:"result,omitempty"`
	Context     string                 `json:"context,omitempty"`
}

// ValidationRule defines parameter validation beyond basic type checking
type ValidationRule struct {
	Field       string      `json:"field"`
	Rule        string      `json:"rule"` // regex, range, enum, custom
	Value       interface{} `json:"value"`
	Message     string      `json:"message"`
	Severity    string      `json:"severity"` // error, warning
}

// EnhancedProperty extends Property with validation and guidance
type EnhancedProperty struct {
	Property
	ValidationRules []ValidationRule `json:"validation_rules,omitempty"`
	Examples        []interface{}    `json:"examples,omitempty"`
	AIHint          string           `json:"ai_hint,omitempty"`
	DependsOn       []string         `json:"depends_on,omitempty"` // other fields this depends on
	Conflicts       []string         `json:"conflicts,omitempty"`  // fields that conflict with this
}

// ToolBuilder helps construct enhanced tools with fluent API
type ToolBuilder struct {
	tool *EnhancedTool
}

// NewToolBuilder creates a new tool builder
func NewToolBuilder(name, description string) *ToolBuilder {
	return &ToolBuilder{
		tool: &EnhancedTool{
			Tool: Tool{
				Name:        name,
				Description: description,
				Parameters: ToolParameters{
					Type:       "object",
					Properties: make(map[string]Property),
				},
			},
			Safety: SafetyMetadata{
				Level: SafetyLevelSafe,
			},
			Performance: PerformanceMetadata{
				ExpectedLatencyMS: 1000,
				CacheTTLSeconds:   60,
			},
			AIGuidance: AIGuidanceMetadata{
				UsageExamples:  []string{},
				CommonPatterns: []string{},
				ErrorPatterns:  make(map[string]string),
			},
			Observability: ObservabilityMetadata{
				MetricsEnabled: true,
				TracingEnabled: true,
				LogLevel:       "info",
			},
			Examples: []ToolExample{},
		},
	}
}

// Category sets the tool category
func (b *ToolBuilder) Category(category ToolCategory) *ToolBuilder {
	b.tool.Category = category
	return b
}

// Handler sets the tool handler
func (b *ToolBuilder) Handler(handler ToolHandler) *ToolBuilder {
	b.tool.Handler = handler
	return b
}

// Required sets required parameters
func (b *ToolBuilder) Required(params ...string) *ToolBuilder {
	b.tool.Parameters.Required = params
	return b
}

// Param adds a parameter with enhanced metadata
func (b *ToolBuilder) Param(name string, prop EnhancedProperty) *ToolBuilder {
	b.tool.Parameters.Properties[name] = Property{
		Type:        prop.Type,
		Description: prop.Description,
		Default:     prop.Default,
		Enum:        prop.Enum,
		Items:       prop.Items,
	}
	return b
}

// Safety configures safety metadata
func (b *ToolBuilder) Safety(config func(*SafetyMetadata)) *ToolBuilder {
	config(&b.tool.Safety)
	return b
}

// Performance configures performance metadata
func (b *ToolBuilder) Performance(config func(*PerformanceMetadata)) *ToolBuilder {
	config(&b.tool.Performance)
	return b
}

// AIGuidance configures AI guidance metadata
func (b *ToolBuilder) AIGuidance(config func(*AIGuidanceMetadata)) *ToolBuilder {
	config(&b.tool.AIGuidance)
	return b
}

// Example adds a usage example
func (b *ToolBuilder) Example(example ToolExample) *ToolBuilder {
	b.tool.Examples = append(b.tool.Examples, example)
	return b
}

// Build returns the constructed enhanced tool
func (b *ToolBuilder) Build() *EnhancedTool {
	// Auto-generate AI hints based on metadata
	if b.tool.Safety.IsDestructive && len(b.tool.AIGuidance.WarningsForAI) == 0 {
		b.tool.AIGuidance.WarningsForAI = append(b.tool.AIGuidance.WarningsForAI,
			"This is a destructive operation. Always use dry_run first and confirm with user.")
	}

	// Set performance expectations based on category
	if b.tool.Category == CategoryQuery && b.tool.Performance.ExpectedLatencyMS == 1000 {
		b.tool.Performance.ExpectedLatencyMS = 500
	} else if b.tool.Category == CategoryMutation {
		b.tool.Performance.ExpectedLatencyMS = 2000
	}

	return b.tool
}

// GenerateAIPrompt generates a comprehensive prompt for AI agents
func (t *EnhancedTool) GenerateAIPrompt() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Tool: %s\n\n", t.Name))
	sb.WriteString(fmt.Sprintf("**Category**: %s\n", t.Category))
	sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", t.Description))

	// Safety information
	if t.Safety.Level != SafetyLevelSafe {
		sb.WriteString(fmt.Sprintf("⚠️ **Safety Level**: %s\n", t.Safety.Level))
		if t.Safety.IsDestructive {
			sb.WriteString("- This is a DESTRUCTIVE operation\n")
		}
		if t.Safety.DryRunSupported {
			sb.WriteString("- Supports dry_run parameter\n")
		}
		if t.Safety.RequiresConfirmation {
			sb.WriteString("- Requires user confirmation\n")
		}
		sb.WriteString("\n")
	}

	// Parameters
	sb.WriteString("### Parameters\n")
	for name, prop := range t.Parameters.Properties {
		required := ""
		for _, req := range t.Parameters.Required {
			if req == name {
				required = " (required)"
				break
			}
		}
		sb.WriteString(fmt.Sprintf("- `%s` (%s%s): %s\n", name, prop.Type, required, prop.Description))
		if prop.Default != nil {
			sb.WriteString(fmt.Sprintf("  - Default: %v\n", prop.Default))
		}
		if len(prop.Enum) > 0 {
			sb.WriteString(fmt.Sprintf("  - Values: %s\n", strings.Join(prop.Enum, ", ")))
		}
	}
	sb.WriteString("\n")

	// Examples
	if len(t.Examples) > 0 {
		sb.WriteString("### Examples\n")
		for _, ex := range t.Examples {
			sb.WriteString(fmt.Sprintf("**%s**: %s\n", ex.Name, ex.Description))
			sb.WriteString("```json\n")
			sb.WriteString(fmt.Sprintf("%v\n", ex.Params))
			sb.WriteString("```\n\n")
		}
	}

	// AI Guidance
	if len(t.AIGuidance.WarningsForAI) > 0 {
		sb.WriteString("### ⚠️ AI Warnings\n")
		for _, warning := range t.AIGuidance.WarningsForAI {
			sb.WriteString(fmt.Sprintf("- %s\n", warning))
		}
		sb.WriteString("\n")
	}

	if len(t.AIGuidance.ChainsWith) > 0 {
		sb.WriteString("### Commonly Used With\n")
		sb.WriteString(strings.Join(t.AIGuidance.ChainsWith, ", ") + "\n\n")
	}

	return sb.String()
}

// ValidateParams validates parameters against enhanced rules
func (t *EnhancedTool) ValidateParams(params map[string]interface{}) error {
	// First do basic validation
	for _, required := range t.Parameters.Required {
		if _, ok := params[required]; !ok {
			return fmt.Errorf("required parameter '%s' is missing", required)
		}
	}

	// Type validation
	for name, value := range params {
		prop, ok := t.Parameters.Properties[name]
		if !ok {
			return fmt.Errorf("unknown parameter '%s'", name)
		}

		// Basic type checking
		if err := validateType(value, prop.Type); err != nil {
			return fmt.Errorf("parameter '%s': %v", name, err)
		}

		// Enum validation
		if len(prop.Enum) > 0 {
			if !contains(prop.Enum, fmt.Sprintf("%v", value)) {
				return fmt.Errorf("parameter '%s' must be one of: %s", name, strings.Join(prop.Enum, ", "))
			}
		}
	}

	return nil
}

// Helper functions
func validateType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "integer":
		switch v := value.(type) {
		case int, int32, int64, float64:
			// Check if float64 is actually an integer
			if f, ok := v.(float64); ok && f != float64(int64(f)) {
				return fmt.Errorf("expected integer, got float")
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
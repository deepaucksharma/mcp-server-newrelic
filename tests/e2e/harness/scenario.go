package harness

import (
	"fmt"
	"time"
)

// Scenario represents a complete E2E test scenario
type Scenario struct {
	ID          string                 `yaml:"id"`
	Title       string                 `yaml:"title"`
	Region      string                 `yaml:"region,omitempty"`
	Setup       *Setup                 `yaml:"setup,omitempty"`
	Workflow    []WorkflowStep         `yaml:"workflow"`
	Assertions  []Assertion            `yaml:"assert"`
	Cleanup     *Cleanup               `yaml:"cleanup,omitempty"`
	Timeout     time.Duration          `yaml:"timeout,omitempty"`
	Environment ScenarioEnvironment    `yaml:"environment,omitempty"`
	Tags        []string               `yaml:"tags,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

// Setup defines pre-test setup steps
type Setup struct {
	SeedDataScript string            `yaml:"seed_data_script,omitempty"`
	Toxiproxy      *ToxiproxyConfig  `yaml:"toxiproxy,omitempty"`
	Environment    map[string]string `yaml:"environment,omitempty"`
	WaitDuration   time.Duration     `yaml:"wait,omitempty"`
}

// ToxiproxyConfig defines network chaos configuration
type ToxiproxyConfig struct {
	Disabled bool                     `yaml:"disabled,omitempty"`
	Proxies  []ToxiproxyProxy         `yaml:"proxies,omitempty"`
}

// ToxiproxyProxy defines a single proxy configuration
type ToxiproxyProxy struct {
	Name       string                 `yaml:"name"`
	Listen     string                 `yaml:"listen"`
	Upstream   string                 `yaml:"upstream"`
	Toxics     []Toxic                `yaml:"toxics,omitempty"`
}

// Toxic defines network fault injection
type Toxic struct {
	Type       string                 `yaml:"type"` // latency, down, bandwidth, slow_close, timeout
	Name       string                 `yaml:"name"`
	Stream     string                 `yaml:"stream,omitempty"` // upstream, downstream
	Toxicity   float64                `yaml:"toxicity,omitempty"`
	Attributes map[string]interface{} `yaml:"attributes,omitempty"`
}

// WorkflowStep represents a single step in the workflow
type WorkflowStep struct {
	Tool         string                 `yaml:"tool"`
	Params       map[string]interface{} `yaml:"params,omitempty"`
	StoreAs      string                 `yaml:"store_as,omitempty"`
	Condition    string                 `yaml:"condition,omitempty"`
	Retry        *RetryConfig           `yaml:"retry,omitempty"`
	Timeout      time.Duration          `yaml:"timeout,omitempty"`
	Parallel     []WorkflowStep         `yaml:"parallel,omitempty"`
	OnError      string                 `yaml:"on_error,omitempty"` // continue, fail, retry
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Delay       time.Duration `yaml:"delay"`
	Backoff     string        `yaml:"backoff,omitempty"` // constant, exponential
}

// Assertion defines a test assertion
type Assertion struct {
	Type     string      `yaml:"type,omitempty"` // jsonpath, trace, nrql, custom
	JSONPath string      `yaml:"jsonpath,omitempty"`
	Query    string      `yaml:"query,omitempty"` // For NRQL assertions
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
	Message  string      `yaml:"message,omitempty"`
}

// Cleanup defines post-test cleanup steps
type Cleanup struct {
	DropDashboardsWithTag string   `yaml:"drop_dashboards_with_tag,omitempty"`
	DeleteAlerts          []string `yaml:"delete_alerts,omitempty"`
	DeleteTestData        bool     `yaml:"delete_test_data,omitempty"`
	CustomCommands        []string `yaml:"custom_commands,omitempty"`
}

// ScenarioEnvironment defines environment-specific settings
type ScenarioEnvironment struct {
	AccountType   string            `yaml:"account_type,omitempty"`
	DataSourceMix string            `yaml:"data_source_mix,omitempty"`
	SchemaDrift   string            `yaml:"schema_drift,omitempty"`
	LoadProfile   string            `yaml:"load_profile,omitempty"`
	Variables     map[string]string `yaml:"variables,omitempty"`
}

// Validate checks if the scenario is valid
func (s *Scenario) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("scenario ID is required")
	}
	if s.Title == "" {
		return fmt.Errorf("scenario title is required")
	}
	if len(s.Workflow) == 0 {
		return fmt.Errorf("scenario must have at least one workflow step")
	}
	if len(s.Assertions) == 0 {
		return fmt.Errorf("scenario must have at least one assertion")
	}

	// Validate workflow steps
	for i, step := range s.Workflow {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("workflow step %d: %w", i, err)
		}
	}

	// Validate assertions
	for i, assertion := range s.Assertions {
		if err := assertion.Validate(); err != nil {
			return fmt.Errorf("assertion %d: %w", i, err)
		}
	}

	return nil
}

// RequiresChaos returns true if the scenario requires chaos testing
func (s *Scenario) RequiresChaos() bool {
	if s.Setup != nil && s.Setup.Toxiproxy != nil && !s.Setup.Toxiproxy.Disabled {
		return true
	}
	
	// Check tags
	for _, tag := range s.Tags {
		if tag == "chaos" || tag == "resilience" {
			return true
		}
	}
	
	return false
}

// RequiresMultiAccount returns true if the scenario requires multiple accounts
func (s *Scenario) RequiresMultiAccount() bool {
	return s.Environment.AccountType == "multi-account" ||
		contains(s.Tags, "cross-account") ||
		contains(s.Tags, "multi-account")
}

// Validate checks if the workflow step is valid
func (w *WorkflowStep) Validate() error {
	if w.Tool == "" {
		return fmt.Errorf("tool is required")
	}
	
	// Validate parallel steps
	if len(w.Parallel) > 0 {
		for i, step := range w.Parallel {
			if err := step.Validate(); err != nil {
				return fmt.Errorf("parallel step %d: %w", i, err)
			}
		}
	}
	
	return nil
}

// Validate checks if the assertion is valid
func (a *Assertion) Validate() error {
	if a.Type == "" {
		// Infer type from fields
		if a.JSONPath != "" {
			a.Type = "jsonpath"
		} else if a.Query != "" {
			a.Type = "nrql"
		} else {
			return fmt.Errorf("assertion type cannot be determined")
		}
	}

	switch a.Type {
	case "jsonpath":
		if a.JSONPath == "" {
			return fmt.Errorf("jsonpath is required for jsonpath assertion")
		}
	case "nrql":
		if a.Query == "" {
			return fmt.Errorf("query is required for nrql assertion")
		}
	case "trace":
		// Trace assertions have different validation
	default:
		return fmt.Errorf("unknown assertion type: %s", a.Type)
	}

	if a.Operator == "" {
		return fmt.Errorf("operator is required")
	}

	// Validate operator
	validOperators := []string{"==", "!=", ">", "<", ">=", "<=", "contains", "not_contains", "matches", "approx"}
	if !contains(validOperators, a.Operator) {
		return fmt.Errorf("invalid operator: %s", a.Operator)
	}

	return nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
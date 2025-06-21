package harness

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/interface/mcp"
	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"gopkg.in/yaml.v3"
)

// ScenarioRunner orchestrates E2E test scenario execution
type ScenarioRunner struct {
	parser       *WorkflowDSLParser
	executor     *StepExecutor
	assertEngine *AssertionEngine
	tracer       *TraceCollector
	config       RunnerConfig
	results      sync.Map
}

// RunnerConfig holds configuration for the scenario runner
type RunnerConfig struct {
	Region        string
	AccountType   string
	DataSourceMix string
	SchemaDrift   string
	LoadProfile   string
	Timeout       time.Duration
	Parallel      int
	SkipChaos     bool
}

// NewScenarioRunner creates a new scenario runner
func NewScenarioRunner(config RunnerConfig) *ScenarioRunner {
	return &ScenarioRunner{
		parser:       NewWorkflowDSLParser(),
		executor:     NewStepExecutor(config),
		assertEngine: NewAssertionEngine(),
		tracer:       NewTraceCollector(config.Region),
		config:       config,
	}
}

// RunAll runs all scenarios in the given directory
func (r *ScenarioRunner) RunAll(ctx context.Context, scenarioDir string) (*TestReport, error) {
	scenarios, err := r.loadScenarios(scenarioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load scenarios: %w", err)
	}

	report := &TestReport{
		StartTime: time.Now(),
		Region:    r.config.Region,
		Scenarios: make([]ScenarioResult, 0, len(scenarios)),
	}

	// Run scenarios in parallel with configured limit
	sem := make(chan struct{}, r.config.Parallel)
	var wg sync.WaitGroup
	resultsChan := make(chan ScenarioResult, len(scenarios))

	for _, scenario := range scenarios {
		// Skip chaos scenarios if requested
		if r.config.SkipChaos && scenario.RequiresChaos() {
			continue
		}

		wg.Add(1)
		go func(s *Scenario) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := r.runScenario(ctx, s)
			resultsChan <- result
		}(scenario)
	}

	// Wait for all scenarios to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		report.Scenarios = append(report.Scenarios, result)
		r.results.Store(result.ID, result)
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.calculateStats()

	return report, nil
}

// RunScenario runs a single scenario
func (r *ScenarioRunner) RunScenario(ctx context.Context, scenarioFile string) (*ScenarioResult, error) {
	scenario, err := r.loadScenario(scenarioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load scenario: %w", err)
	}

	return &r.runScenario(ctx, scenario), nil
}

// runScenario executes a single scenario with full lifecycle
func (r *ScenarioRunner) runScenario(ctx context.Context, scenario *Scenario) ScenarioResult {
	result := ScenarioResult{
		ID:        scenario.ID,
		Title:     scenario.Title,
		StartTime: time.Now(),
		Region:    r.config.Region,
		Steps:     make([]StepResult, 0),
	}

	// Apply timeout
	timeoutDuration := scenario.Timeout
	if timeoutDuration == 0 {
		timeoutDuration = r.config.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	// Setup phase
	if scenario.Setup != nil {
		if err := r.runSetup(ctx, scenario.Setup); err != nil {
			result.Error = fmt.Errorf("setup failed: %w", err)
			result.Status = StatusFailed
			result.EndTime = time.Now()
			return result
		}
	}

	// Execute workflow
	workflowResult, err := r.executeWorkflow(ctx, scenario)
	if err != nil {
		result.Error = fmt.Errorf("workflow execution failed: %w", err)
		result.Status = StatusFailed
	} else {
		result.Steps = workflowResult.Steps
		result.TraceID = workflowResult.TraceID

		// Run assertions
		assertionResults := r.runAssertions(ctx, scenario, workflowResult)
		result.Assertions = assertionResults

		// Determine overall status
		if allAssertionsPassed(assertionResults) {
			result.Status = StatusPassed
		} else {
			result.Status = StatusFailed
			result.Error = fmt.Errorf("assertions failed")
		}
	}

	// Cleanup phase (always run)
	if scenario.Cleanup != nil {
		r.runCleanup(context.Background(), scenario.Cleanup) // Use fresh context
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Collect trace information
	if result.TraceID != "" {
		trace, err := r.tracer.CollectTrace(context.Background(), result.TraceID)
		if err == nil {
			result.TraceData = trace
		}
	}

	return result
}

// loadScenarios loads all scenario files from a directory
func (r *ScenarioRunner) loadScenarios(dir string) ([]*Scenario, error) {
	var scenarios []*Scenario

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			scenario, err := r.loadScenario(path)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", path, err)
			}
			scenarios = append(scenarios, scenario)
		}

		return nil
	})

	return scenarios, err
}

// loadScenario loads a single scenario from a YAML file
func (r *ScenarioRunner) loadScenario(file string) (*Scenario, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var scenario Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate scenario
	if err := scenario.Validate(); err != nil {
		return nil, fmt.Errorf("invalid scenario: %w", err)
	}

	// Apply environment overrides
	r.applyEnvironmentOverrides(&scenario)

	return &scenario, nil
}

// applyEnvironmentOverrides applies environment-specific overrides to the scenario
func (r *ScenarioRunner) applyEnvironmentOverrides(scenario *Scenario) {
	// Override region if specified
	if r.config.Region != "" && scenario.Region == "" {
		scenario.Region = r.config.Region
	}

	// Apply schema drift if configured
	if r.config.SchemaDrift != "" {
		scenario.Environment.SchemaDrift = r.config.SchemaDrift
	}

	// Apply load profile
	if r.config.LoadProfile != "" {
		scenario.Environment.LoadProfile = r.config.LoadProfile
	}
}

// runSetup executes scenario setup steps
func (r *ScenarioRunner) runSetup(ctx context.Context, setup *Setup) error {
	// Run seed data script
	if setup.SeedDataScript != "" {
		if err := r.executeSeedScript(ctx, setup.SeedDataScript); err != nil {
			return fmt.Errorf("seed script failed: %w", err)
		}
	}

	// Configure toxiproxy if needed
	if setup.Toxiproxy != nil && !setup.Toxiproxy.Disabled {
		if err := r.configureToxiproxy(ctx, setup.Toxiproxy); err != nil {
			return fmt.Errorf("toxiproxy setup failed: %w", err)
		}
	}

	// Wait for data to propagate
	if setup.WaitDuration > 0 {
		select {
		case <-time.After(setup.WaitDuration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// executeWorkflow runs the scenario workflow
func (r *ScenarioRunner) executeWorkflow(ctx context.Context, scenario *Scenario) (*WorkflowResult, error) {
	// Parse workflow steps
	workflow, err := r.parser.Parse(scenario.Workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	// Execute steps
	result, err := r.executor.Execute(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	return result, nil
}

// runAssertions executes all assertions for the scenario
func (r *ScenarioRunner) runAssertions(ctx context.Context, scenario *Scenario, workflowResult *WorkflowResult) []AssertionResult {
	return r.assertEngine.RunAssertions(ctx, scenario.Assertions, workflowResult)
}

// runCleanup executes cleanup steps
func (r *ScenarioRunner) runCleanup(ctx context.Context, cleanup *Cleanup) {
	// Best effort cleanup - don't fail the test if cleanup fails
	
	if cleanup.DropDashboardsWithTag != "" {
		r.cleanupDashboards(ctx, cleanup.DropDashboardsWithTag)
	}

	if cleanup.DeleteTestData {
		r.deleteTestData(ctx)
	}

	if len(cleanup.CustomCommands) > 0 {
		for _, cmd := range cleanup.CustomCommands {
			r.executeCommand(ctx, cmd)
		}
	}
}

// Helper methods

func (r *ScenarioRunner) executeSeedScript(ctx context.Context, script string) error {
	// Execute seed data script
	// This would run Python/Go script to insert test data into New Relic
	return nil
}

func (r *ScenarioRunner) configureToxiproxy(ctx context.Context, config *ToxiproxyConfig) error {
	// Configure network chaos via toxiproxy
	return nil
}

func (r *ScenarioRunner) cleanupDashboards(ctx context.Context, tag string) {
	// Delete test dashboards with specific tag
}

func (r *ScenarioRunner) deleteTestData(ctx context.Context) {
	// Remove test data from New Relic
}

func (r *ScenarioRunner) executeCommand(ctx context.Context, cmd string) {
	// Execute arbitrary cleanup command
}

func allAssertionsPassed(results []AssertionResult) bool {
	for _, r := range results {
		if !r.Passed {
			return false
		}
	}
	return true
}
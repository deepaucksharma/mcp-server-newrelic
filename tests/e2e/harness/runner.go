package harness

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	Region           string
	AccountType      string
	DataSourceMix    string
	SchemaDrift      string
	LoadProfile      string
	Timeout          time.Duration
	Parallel         int
	SkipChaos        bool
	MaxParallel      int
	RetryAttempts    int
	CaptureTraffic   bool
	SaveResponses    bool
	CleanupTestData  bool
	OutputDir        string
	MCPServerURL     string
	MCPServerCommand string
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

// RunScenarios runs multiple scenarios and returns a test report
func (r *ScenarioRunner) RunScenarios(ctx context.Context, scenarioFiles []string) (*TestReport, error) {
	report := &TestReport{
		StartTime:    time.Now(),
		Region:       r.config.Region,
		Environment:  r.config.AccountType, // Map AccountType to Environment
		Scenarios:    make([]ScenarioResult, 0, len(scenarioFiles)),
	}

	// Run scenarios in parallel based on configuration
	sem := make(chan struct{}, r.config.MaxParallel)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, file := range scenarioFiles {
		wg.Add(1)
		go func(scenarioFile string) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Run scenario
			result, err := r.RunScenario(ctx, scenarioFile)
			if err != nil {
				// Create a failed result
				result = &ScenarioResult{
					ID:        filepath.Base(scenarioFile),
					Title:     scenarioFile,
					Status:    StatusFailed,
					StartTime: time.Now(),
					EndTime:   time.Now(),
					Error:     err,
				}
			}

			// Add to report
			mu.Lock()
			report.Scenarios = append(report.Scenarios, *result)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	// Finalize report
	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	// Calculate summary
	report.Summary = r.calculateSummary(report.Scenarios)

	return report, nil
}

// RunScenario runs a single scenario
func (r *ScenarioRunner) RunScenario(ctx context.Context, scenarioFile string) (*ScenarioResult, error) {
	scenario, err := r.loadScenario(scenarioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load scenario: %w", err)
	}

	result := r.runScenario(ctx, scenario)
	return &result, nil
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

// ParseScenario parses a scenario file without loading it fully
func (r *ScenarioRunner) ParseScenario(file string) (*Scenario, error) {
	return r.loadScenario(file)
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

// calculateSummary calculates summary statistics from results
func (r *ScenarioRunner) calculateSummary(results []ScenarioResult) ReportSummary {
	summary := ReportSummary{
		TotalScenarios: len(results),
	}

	var totalDuration time.Duration
	for _, result := range results {
		totalDuration += result.Duration
		
		switch result.Status {
		case StatusPassed:
			summary.PassedScenarios++
		case StatusFailed:
			summary.FailedScenarios++
		case StatusSkipped:
			summary.SkippedScenarios++
		}
	}

	if summary.TotalScenarios > 0 {
		summary.PassRate = float64(summary.PassedScenarios) / float64(summary.TotalScenarios) * 100
		summary.AverageDuration = totalDuration / time.Duration(summary.TotalScenarios)
	}
	summary.TotalDuration = totalDuration

	return summary
}
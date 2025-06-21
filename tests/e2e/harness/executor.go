package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
	"github.com/google/uuid"
)

// StepExecutor executes workflow steps
type StepExecutor struct {
	client    *framework.MCPTestClient
	config    RunnerConfig
	variables map[string]interface{}
	mu        sync.RWMutex
}

// WorkflowResult contains the results of workflow execution
type WorkflowResult struct {
	Steps    []StepResult
	TraceID  string
	Duration time.Duration
	Context  map[string]interface{}
}

// StepResult contains the result of a single step
type StepResult struct {
	Index     int
	Tool      string
	Params    map[string]interface{}
	Result    map[string]interface{}
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	TraceID   string
	Retries   int
}

// NewStepExecutor creates a new step executor
func NewStepExecutor(config RunnerConfig) *StepExecutor {
	// Create test account based on config
	account := &framework.TestAccount{
		Name:      config.AccountType,
		APIKey:    getAPIKeyForRegion(config.Region),
		AccountID: getAccountIDForType(config.AccountType),
		Region:    config.Region,
	}

	return &StepExecutor{
		client:    framework.NewMCPTestClient(account),
		config:    config,
		variables: make(map[string]interface{}),
	}
}

// Execute runs a parsed workflow
func (e *StepExecutor) Execute(ctx context.Context, workflow *ParsedWorkflow) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Steps:   make([]StepResult, 0, len(workflow.Steps)),
		TraceID: uuid.New().String(),
		Context: make(map[string]interface{}),
	}

	startTime := time.Now()

	// Initialize variables
	e.mu.Lock()
	for k, v := range workflow.Variables {
		e.variables[k] = v
	}
	e.variables["_trace_id"] = result.TraceID
	e.mu.Unlock()

	// Execute steps sequentially
	for i, step := range workflow.Steps {
		// Check condition
		if step.Condition != nil {
			shouldRun, err := e.evaluateCondition(step.Condition)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate condition: %w", err)
			}
			if !shouldRun {
				continue
			}
		}

		// Execute step
		var stepResult StepResult
		if len(step.ParallelSteps) > 0 {
			// Execute parallel steps
			stepResult = e.executeParallelSteps(ctx, step.ParallelSteps)
		} else {
			// Execute single step
			stepResult = e.executeStep(ctx, step)
		}

		stepResult.Index = i
		result.Steps = append(result.Steps, stepResult)

		// Store result if needed
		if step.StoreAs != "" {
			e.storeStepResult(step.StoreAs, stepResult)
		}

		// Handle errors based on OnError policy
		if stepResult.Error != nil {
			switch step.OnError {
			case "continue":
				// Continue to next step
			case "retry":
				// Retry logic is handled within executeStep
			default: // "fail"
				return result, fmt.Errorf("step %d failed: %w", i, stepResult.Error)
			}
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// executeStep executes a single workflow step
func (e *StepExecutor) executeStep(ctx context.Context, step ParsedStep) StepResult {
	result := StepResult{
		Tool:      step.Tool,
		Params:    step.Params,
		StartTime: time.Now(),
		TraceID:   e.getTraceID(),
	}

	// Apply step timeout
	stepCtx := ctx
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	// Resolve parameters with current variables
	resolvedParams, err := e.resolveParams(step.Params)
	if err != nil {
		result.Error = fmt.Errorf("failed to resolve params: %w", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}
	result.Params = resolvedParams

	// Execute with retry logic
	maxAttempts := 1
	if step.Retry != nil {
		maxAttempts = step.Retry.MaxAttempts
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Add trace context
		if e.config.Region != "" {
			resolvedParams["_trace_id"] = result.TraceID
			resolvedParams["_region"] = e.config.Region
		}

		// Execute tool
		toolResult, err := e.client.ExecuteTool(stepCtx, step.Tool, resolvedParams)
		
		if err == nil {
			result.Result = toolResult
			break
		}

		result.Error = err
		result.Retries = attempt - 1

		// Check if we should retry
		if attempt < maxAttempts && isRetryableError(err) {
			delay := calculateRetryDelay(step.Retry, attempt)
			select {
			case <-time.After(delay):
				// Continue to next attempt
			case <-stepCtx.Done():
				result.Error = fmt.Errorf("timeout during retry: %w", stepCtx.Err())
				break
			}
		} else {
			break
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result
}

// executeParallelSteps executes multiple steps in parallel
func (e *StepExecutor) executeParallelSteps(ctx context.Context, steps []ParsedStep) StepResult {
	result := StepResult{
		Tool:      "parallel",
		StartTime: time.Now(),
		Result:    make(map[string]interface{}),
	}

	var wg sync.WaitGroup
	results := make([]StepResult, len(steps))

	for i, step := range steps {
		wg.Add(1)
		go func(idx int, s ParsedStep) {
			defer wg.Done()
			results[idx] = e.executeStep(ctx, s)
		}(i, step)
	}

	wg.Wait()

	// Aggregate results
	parallelResults := make([]interface{}, len(results))
	var firstError error
	for i, r := range results {
		parallelResults[i] = map[string]interface{}{
			"tool":   r.Tool,
			"result": r.Result,
			"error":  r.Error,
		}
		if firstError == nil && r.Error != nil {
			firstError = r.Error
		}
	}

	result.Result["parallel_results"] = parallelResults
	result.Error = firstError
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result
}

// evaluateCondition evaluates a step condition
func (e *StepExecutor) evaluateCondition(condition *ParsedCondition) (bool, error) {
	// Resolve variables in expression
	e.mu.RLock()
	expr := condition.Expression
	for _, varName := range condition.Variables {
		if value, exists := e.variables[varName]; exists {
			// Simple string replacement for now
			expr = strings.ReplaceAll(expr, "${"+varName+"}", fmt.Sprintf("%v", value))
		}
	}
	e.mu.RUnlock()

	// Evaluate expression (simple implementation)
	// In production, use a proper expression evaluator
	return evaluateExpression(expr)
}

// resolveParams resolves parameter values with current variables
func (e *StepExecutor) resolveParams(params map[string]interface{}) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})
	
	e.mu.RLock()
	defer e.mu.RUnlock()

	for key, value := range params {
		resolvedValue, err := e.resolveValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve param %s: %w", key, err)
		}
		resolved[key] = resolvedValue
	}

	return resolved, nil
}

// resolveValue resolves a single value, handling variable references
func (e *StepExecutor) resolveValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Check for variable reference
		if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			varName := v[2 : len(v)-1]
			if resolved, exists := e.variables[varName]; exists {
				return resolved, nil
			}
			return nil, fmt.Errorf("variable %s not found", varName)
		}
		return v, nil

	case map[string]interface{}:
		// Recursively resolve map values
		resolved := make(map[string]interface{})
		for k, val := range v {
			rVal, err := e.resolveValue(val)
			if err != nil {
				return nil, err
			}
			resolved[k] = rVal
		}
		return resolved, nil

	case []interface{}:
		// Recursively resolve array values
		resolved := make([]interface{}, len(v))
		for i, val := range v {
			rVal, err := e.resolveValue(val)
			if err != nil {
				return nil, err
			}
			resolved[i] = rVal
		}
		return resolved, nil

	default:
		return value, nil
	}
}

// storeStepResult stores a step result for later reference
func (e *StepExecutor) storeStepResult(name string, result StepResult) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.variables[name] = result.Result
	e.variables[name+"_error"] = result.Error
	e.variables[name+"_duration"] = result.Duration
}

// getTraceID returns the current trace ID
func (e *StepExecutor) getTraceID() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if traceID, ok := e.variables["_trace_id"].(string); ok {
		return traceID
	}
	return ""
}

// Helper functions

func getAPIKeyForRegion(region string) string {
	envVar := fmt.Sprintf("NEW_RELIC_API_KEY_%s", strings.ToUpper(region))
	if key := os.Getenv(envVar); key != "" {
		return key
	}
	// Fallback to default
	return os.Getenv("NEW_RELIC_API_KEY")
}

func getAccountIDForType(accountType string) string {
	switch accountType {
	case "multi-account":
		return os.Getenv("NEW_RELIC_ACCOUNT_ID_MULTI")
	case "sandbox-single":
		return os.Getenv("NEW_RELIC_ACCOUNT_ID_SANDBOX")
	default:
		return os.Getenv("NEW_RELIC_ACCOUNT_ID")
	}
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	retryablePatterns := []string{
		"timeout",
		"temporary",
		"rate limit",
		"429",
		"503",
		"connection refused",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}
	
	return false
}

func calculateRetryDelay(config *RetryConfig, attempt int) time.Duration {
	if config == nil {
		return time.Second
	}
	
	baseDelay := config.Delay
	if baseDelay == 0 {
		baseDelay = time.Second
	}
	
	switch config.Backoff {
	case "exponential":
		return baseDelay * time.Duration(1<<(attempt-1))
	default: // constant
		return baseDelay
	}
}

func evaluateExpression(expr string) (bool, error) {
	// Simple expression evaluation
	// In production, use a proper expression library
	
	// Handle basic comparisons
	if strings.Contains(expr, "==") {
		parts := strings.Split(expr, "==")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) == strings.TrimSpace(parts[1]), nil
		}
	}
	
	if strings.Contains(expr, "!=") {
		parts := strings.Split(expr, "!=")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) != strings.TrimSpace(parts[1]), nil
		}
	}
	
	// Handle boolean values
	expr = strings.TrimSpace(strings.ToLower(expr))
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}
	
	return false, fmt.Errorf("unsupported expression: %s", expr)
}
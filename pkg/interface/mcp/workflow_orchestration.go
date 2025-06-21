package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ContextManager manages workflow execution context
type ContextManager struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// StepValidator validates workflow steps
type StepValidator interface {
	Validate(step *WorkflowStepDef, context *WorkflowContext) error
}

// DataTransformer transforms data between workflow steps
type DataTransformer interface {
	Transform(input interface{}, spec map[string]interface{}) (interface{}, error)
}

// WorkflowOrchestrator manages complex workflow execution patterns
type WorkflowOrchestrator struct {
	mu              sync.RWMutex
	workflows       map[string]*WorkflowExecution
	toolRegistry    ToolRegistry
	contextManager  *ContextManager
	executionEngine *ExecutionEngine
}

// WorkflowExecution represents an active workflow
type WorkflowExecution struct {
	ID              string
	Definition      *WorkflowDefinition
	State           WorkflowExecutionState
	Context         *WorkflowContext
	CurrentStepIdx  int
	ExecutionLog    []ExecutionLogEntry
	StartTime       time.Time
	EndTime         *time.Time
	Error           error
}

// WorkflowDefinition defines the structure of a workflow
type WorkflowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Triggers    []WorkflowTrigger      `json:"triggers"`
	Inputs      []WorkflowInput        `json:"inputs"`
	Steps       []WorkflowStepDef      `json:"steps"`
	Outputs     []WorkflowOutput       `json:"outputs"`
	ErrorPolicy ErrorHandlingPolicy    `json:"error_policy"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WorkflowStepDef defines a single step in the workflow
type WorkflowStepDef struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           StepType               `json:"type"`
	Tool           string                 `json:"tool,omitempty"`
	Inputs         map[string]interface{} `json:"inputs"`
	Conditions     []StepCondition        `json:"conditions,omitempty"`
	ErrorHandling  StepErrorHandling      `json:"error_handling"`
	Retry          RetryPolicy            `json:"retry"`
	Timeout        time.Duration          `json:"timeout"`
	DependsOn      []string               `json:"depends_on,omitempty"`
	ContinueOnFail bool                   `json:"continue_on_fail"`
}

// StepType defines the type of workflow step
type StepType string

const (
	StepTypeSimple      StepType = "simple"
	StepTypeParallel    StepType = "parallel"
	StepTypeConditional StepType = "conditional"
	StepTypeLoop        StepType = "loop"
	StepTypeSubWorkflow StepType = "sub_workflow"
)

// WorkflowContext maintains state across workflow execution
type WorkflowContext struct {
	mu       sync.RWMutex
	data     map[string]interface{}
	findings []Finding
	metadata map[string]interface{}
}

// Finding represents a discovery during workflow execution
type Finding struct {
	ID          string                 `json:"id"`
	Type        FindingType            `json:"type"`
	Severity    FindingSeverity        `json:"severity"`
	Description string                 `json:"description"`
	Evidence    map[string]interface{} `json:"evidence"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Related     []string               `json:"related,omitempty"`
}

// FindingType categorizes findings
type FindingType string

const (
	FindingTypeAnomaly      FindingType = "anomaly"
	FindingTypeCorrelation  FindingType = "correlation"
	FindingTypeRootCause    FindingType = "root_cause"
	FindingTypeImpact       FindingType = "impact"
	FindingTypeRecommendation FindingType = "recommendation"
)

// FindingSeverity indicates importance
type FindingSeverity string

const (
	SeverityCritical FindingSeverity = "critical"
	SeverityHigh     FindingSeverity = "high"
	SeverityMedium   FindingSeverity = "medium"
	SeverityLow      FindingSeverity = "low"
	SeverityInfo     FindingSeverity = "info"
)

// ExecutionEngine handles the actual execution of workflow steps
type ExecutionEngine struct {
	toolRegistry ToolRegistry
	executor     *StepExecutor
	validator    *StepValidator
	transformer  *DataTransformer
}

// StepExecutor executes individual steps
type StepExecutor struct {
	concurrencyLimit int
	semaphore        chan struct{}
}

// Orchestration Patterns Implementation

// 1. SEQUENTIAL PATTERN - Steps execute one after another
func (o *WorkflowOrchestrator) ExecuteSequential(ctx context.Context, workflowID string, steps []WorkflowStepDef) error {
	execution, err := o.getExecution(workflowID)
	if err != nil {
		return err
	}

	for i, step := range steps {
		execution.CurrentStepIdx = i
		
		// Check if we should skip this step
		if !o.shouldExecuteStep(ctx, execution, &step) {
			o.logStepSkipped(execution, step.Name, "Condition not met")
			continue
		}

		// Execute the step
		result, err := o.executeStep(ctx, execution, &step)
		if err != nil {
			if !step.ContinueOnFail {
				return fmt.Errorf("step %s failed: %w", step.ID, err)
			}
			o.logStepError(execution, step.Name, err)
			continue
		}

		// Store result in context
		execution.Context.Set(step.ID+".output", result)
		o.logStepCompleted(execution, step.Name, result)
	}

	return nil
}

// 2. PARALLEL PATTERN - Execute multiple steps concurrently
func (o *WorkflowOrchestrator) ExecuteParallel(ctx context.Context, workflowID string, steps []WorkflowStepDef, maxConcurrent int) error {
	execution, err := o.getExecution(workflowID)
	if err != nil {
		return err
	}

	// Create semaphore for concurrency control
	sem := make(chan struct{}, maxConcurrent)
	errChan := make(chan error, len(steps))
	var wg sync.WaitGroup

	for _, step := range steps {
		step := step // Capture for goroutine
		wg.Add(1)
		
		go func() {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute step
			result, err := o.executeStep(ctx, execution, &step)
			if err != nil {
				if !step.ContinueOnFail {
					errChan <- fmt.Errorf("step %s failed: %w", step.ID, err)
					return
				}
				o.logStepError(execution, step.Name, err)
				return
			}

			// Store result
			execution.Context.Set(step.ID+".output", result)
			o.logStepCompleted(execution, step.Name, result)
		}()
	}

	// Wait for all goroutines
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// 3. CONDITIONAL PATTERN - Execute based on conditions
func (o *WorkflowOrchestrator) ExecuteConditional(ctx context.Context, workflowID string, condition StepCondition, trueBranch, falseBranch []WorkflowStepDef) error {
	execution, err := o.getExecution(workflowID)
	if err != nil {
		return err
	}

	// Evaluate condition
	conditionMet, err := o.evaluateCondition(ctx, execution, condition)
	if err != nil {
		return fmt.Errorf("condition evaluation failed: %w", err)
	}

	// Execute appropriate branch
	if conditionMet {
		o.logDecision(execution, "Condition met, executing true branch")
		return o.ExecuteSequential(ctx, workflowID, trueBranch)
	} else {
		o.logDecision(execution, "Condition not met, executing false branch")
		return o.ExecuteSequential(ctx, workflowID, falseBranch)
	}
}

// 4. LOOP PATTERN - Execute steps repeatedly
func (o *WorkflowOrchestrator) ExecuteLoop(ctx context.Context, workflowID string, steps []WorkflowStepDef, loopConfig LoopConfig) error {
	execution, err := o.getExecution(workflowID)
	if err != nil {
		return err
	}

	iteration := 0
	for {
		// Check loop conditions
		if loopConfig.MaxIterations > 0 && iteration >= loopConfig.MaxIterations {
			o.logInfo(execution, fmt.Sprintf("Loop completed after %d iterations (max reached)", iteration))
			break
		}

		// Check exit condition
		if loopConfig.ExitCondition != nil {
			shouldExit, err := o.evaluateCondition(ctx, execution, *loopConfig.ExitCondition)
			if err != nil {
				return fmt.Errorf("exit condition evaluation failed: %w", err)
			}
			if shouldExit {
				o.logInfo(execution, fmt.Sprintf("Loop completed after %d iterations (exit condition met)", iteration))
				break
			}
		}

		// Execute loop body
		loopCtx := fmt.Sprintf("loop.iteration_%d", iteration)
		execution.Context.Set(loopCtx+".index", iteration)
		
		err = o.ExecuteSequential(ctx, workflowID, steps)
		if err != nil {
			if !loopConfig.ContinueOnError {
				return fmt.Errorf("loop iteration %d failed: %w", iteration, err)
			}
			o.logError(execution, fmt.Sprintf("Loop iteration %d failed, continuing", iteration), err)
		}

		iteration++
		
		// Delay between iterations if specified
		if loopConfig.DelayBetween > 0 {
			select {
			case <-time.After(loopConfig.DelayBetween):
				// Continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

// 5. MAP-REDUCE PATTERN - Process collection with map and reduce phases
func (o *WorkflowOrchestrator) ExecuteMapReduce(ctx context.Context, workflowID string, mapStep, reduceStep WorkflowStepDef, items []interface{}) error {
	execution, err := o.getExecution(workflowID)
	if err != nil {
		return err
	}

	// Map phase - process each item in parallel
	mapResults := make([]interface{}, len(items))
	mapErrors := make([]error, len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		i, item := i, item // Capture for goroutine
		wg.Add(1)
		
		go func() {
			defer wg.Done()
			
			// Create map step with item as input
			itemStep := mapStep
			itemStep.ID = fmt.Sprintf("%s_item_%d", mapStep.ID, i)
			itemStep.Inputs["item"] = item
			itemStep.Inputs["index"] = i
			
			result, err := o.executeStep(ctx, execution, &itemStep)
			mapResults[i] = result
			mapErrors[i] = err
		}()
	}

	wg.Wait()

	// Check for map errors
	var failedCount int
	for i, err := range mapErrors {
		if err != nil {
			failedCount++
			o.logError(execution, fmt.Sprintf("Map failed for item %d", i), err)
		}
	}

	if failedCount > 0 && failedCount == len(items) {
		return fmt.Errorf("all map operations failed")
	}

	// Reduce phase - combine results
	reduceInputs := reduceStep.Inputs
	if reduceInputs == nil {
		reduceInputs = make(map[string]interface{})
	}
	reduceInputs["map_results"] = mapResults
	reduceInputs["map_errors"] = mapErrors
	reduceStep.Inputs = reduceInputs

	result, err := o.executeStep(ctx, execution, &reduceStep)
	if err != nil {
		return fmt.Errorf("reduce step failed: %w", err)
	}

	execution.Context.Set("map_reduce.result", result)
	return nil
}

// 6. SAGA PATTERN - Distributed transaction with compensations
func (o *WorkflowOrchestrator) ExecuteSaga(ctx context.Context, workflowID string, transactions []SagaTransaction) error {
	execution, err := o.getExecution(workflowID)
	if err != nil {
		return err
	}

	completedTransactions := []SagaTransaction{}
	
	// Execute transactions
	for _, tx := range transactions {
		o.logInfo(execution, fmt.Sprintf("Executing transaction: %s", tx.Name))
		
		result, err := o.executeStep(ctx, execution, &tx.Action)
		if err != nil {
			o.logError(execution, fmt.Sprintf("Transaction %s failed", tx.Name), err)
			
			// Rollback completed transactions
			o.logInfo(execution, "Starting saga rollback")
			rollbackErr := o.rollbackSaga(ctx, execution, completedTransactions)
			if rollbackErr != nil {
				return fmt.Errorf("transaction failed and rollback failed: %w", rollbackErr)
			}
			
			return fmt.Errorf("saga failed at transaction %s: %w", tx.Name, err)
		}
		
		execution.Context.Set(fmt.Sprintf("saga.%s.result", tx.Name), result)
		completedTransactions = append(completedTransactions, tx)
	}

	o.logInfo(execution, "Saga completed successfully")
	return nil
}

// Helper methods

func (o *WorkflowOrchestrator) executeStep(ctx context.Context, execution *WorkflowExecution, step *WorkflowStepDef) (interface{}, error) {
	// Create step context with timeout
	stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
	defer cancel()

	// Prepare inputs by resolving references
	resolvedInputs, err := o.resolveInputs(ctx, execution, step.Inputs)
	if err != nil {
		return nil, fmt.Errorf("input resolution failed: %w", err)
	}

	// Get tool from registry
	tool, exists := o.toolRegistry.Get(step.Tool)
	if !exists {
		return nil, fmt.Errorf("tool %s not found", step.Tool)
	}

	// Execute tool handler directly (retry logic can be added later)
	result, err := tool.Handler(stepCtx, resolvedInputs)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}

func (o *WorkflowOrchestrator) evaluateCondition(ctx context.Context, execution *WorkflowExecution, condition StepCondition) (bool, error) {
	// Resolve condition parameters
	left, err := o.resolveValue(ctx, execution, condition.Left)
	if err != nil {
		return false, err
	}

	right, err := o.resolveValue(ctx, execution, condition.Right)
	if err != nil {
		return false, err
	}

	// Evaluate based on operator
	switch condition.Operator {
	case "equals":
		return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right), nil
	case "not_equals":
		return fmt.Sprintf("%v", left) != fmt.Sprintf("%v", right), nil
	case "greater_than":
		leftNum, err := toFloat64(left)
		if err != nil {
			return false, err
		}
		rightNum, err := toFloat64(right)
		if err != nil {
			return false, err
		}
		return compareNumeric(leftNum, rightNum, ">"), nil
	case "less_than":
		leftNum, err := toFloat64(left)
		if err != nil {
			return false, err
		}
		rightNum, err := toFloat64(right)
		if err != nil {
			return false, err
		}
		return compareNumeric(leftNum, rightNum, "<"), nil
	case "contains":
		leftStr := fmt.Sprintf("%v", left)
		rightStr := fmt.Sprintf("%v", right)
		return strings.Contains(leftStr, rightStr), nil
	case "exists":
		return left != nil, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", condition.Operator)
	}
}

func (o *WorkflowOrchestrator) rollbackSaga(ctx context.Context, execution *WorkflowExecution, transactions []SagaTransaction) error {
	// Execute compensations in reverse order
	for i := len(transactions) - 1; i >= 0; i-- {
		tx := transactions[i]
		
		if tx.Compensation == nil {
			o.logWarning(execution, fmt.Sprintf("No compensation defined for transaction: %s", tx.Name))
			continue
		}
		
		o.logInfo(execution, fmt.Sprintf("Executing compensation for: %s", tx.Name))
		
		_, err := o.executeStep(ctx, execution, tx.Compensation)
		if err != nil {
			o.logError(execution, fmt.Sprintf("Compensation failed for %s", tx.Name), err)
			// Continue with other compensations
		}
	}
	
	return nil
}

// Workflow Context methods

func (c *WorkflowContext) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *WorkflowContext) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists := c.data[key]
	return val, exists
}

func (c *WorkflowContext) AddFinding(finding Finding) {
	c.mu.Lock()
	defer c.mu.Unlock()
	finding.ID = fmt.Sprintf("finding_%d", len(c.findings))
	finding.Timestamp = time.Now()
	c.findings = append(c.findings, finding)
}

func (c *WorkflowContext) GetFindings() []Finding {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]Finding{}, c.findings...)
}

// Supporting types

// Helper methods for WorkflowOrchestrator

func (o *WorkflowOrchestrator) getExecution(id string) (*WorkflowExecution, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	execution, exists := o.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow execution not found: %s", id)
	}
	return execution, nil
}

func (o *WorkflowOrchestrator) shouldExecuteStep(ctx context.Context, execution *WorkflowExecution, step *WorkflowStepDef) bool {
	// Check conditions
	for _, condition := range step.Conditions {
		result, err := o.evaluateCondition(ctx, execution, condition)
		if err != nil || !result {
			return false
		}
	}
	
	// Check dependencies
	for _, dep := range step.DependsOn {
		if !o.isStepCompleted(execution, dep) {
			return false
		}
	}
	
	return true
}

func (o *WorkflowOrchestrator) isStepCompleted(execution *WorkflowExecution, stepName string) bool {
	for _, log := range execution.ExecutionLog {
		if log.StepID == stepName && log.Level == "info" && log.Message == "Step completed" {
			return true
		}
	}
	return false
}

func (o *WorkflowOrchestrator) logStepSkipped(execution *WorkflowExecution, stepName string, reason string) {
	o.addLogEntry(execution, ExecutionLogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		StepID:    stepName,
		Message:   fmt.Sprintf("Step skipped: %s", reason),
		Data: map[string]interface{}{
			"status": "skipped",
		},
	})
}

func (o *WorkflowOrchestrator) logStepError(execution *WorkflowExecution, stepName string, err error) {
	o.addLogEntry(execution, ExecutionLogEntry{
		Timestamp: time.Now(),
		Level:     "error",
		StepID:    stepName,
		Message:   fmt.Sprintf("Step error: %v", err),
		Data: map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		},
	})
}

func (o *WorkflowOrchestrator) logStepCompleted(execution *WorkflowExecution, stepName string, result interface{}) {
	o.addLogEntry(execution, ExecutionLogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		StepID:    stepName,
		Message:   "Step completed",
		Data: map[string]interface{}{
			"status": "completed",
			"result": result,
		},
	})
}

func (o *WorkflowOrchestrator) logDecision(execution *WorkflowExecution, message string) {
	o.addLogEntry(execution, ExecutionLogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		StepID:    "decision",
		Message:   message,
	})
}

func (o *WorkflowOrchestrator) logInfo(execution *WorkflowExecution, message string) {
	o.addLogEntry(execution, ExecutionLogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   message,
	})
}

func (o *WorkflowOrchestrator) logError(execution *WorkflowExecution, message string, err error) {
	o.addLogEntry(execution, ExecutionLogEntry{
		Timestamp: time.Now(),
		Level:     "error",
		Message:   message,
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	})
}

func (o *WorkflowOrchestrator) logWarning(execution *WorkflowExecution, message string) {
	o.addLogEntry(execution, ExecutionLogEntry{
		Timestamp: time.Now(),
		Level:     "warning",
		Message:   message,
	})
}

func (o *WorkflowOrchestrator) addLogEntry(execution *WorkflowExecution, entry ExecutionLogEntry) {
	o.mu.Lock()
	defer o.mu.Unlock()
	execution.ExecutionLog = append(execution.ExecutionLog, entry)
}

func (o *WorkflowOrchestrator) resolveInputs(ctx context.Context, execution *WorkflowExecution, inputs map[string]interface{}) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})
	
	for key, value := range inputs {
		resolvedValue, err := o.resolveValue(ctx, execution, value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve input %s: %w", key, err)
		}
		resolved[key] = resolvedValue
	}
	
	return resolved, nil
}

func (o *WorkflowOrchestrator) resolveValue(ctx context.Context, execution *WorkflowExecution, value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Check if it's a reference like ${context.variable}
		if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			ref := strings.TrimSuffix(strings.TrimPrefix(v, "${"), "}")
			parts := strings.Split(ref, ".")
			
			if len(parts) >= 2 && parts[0] == "context" {
				val, exists := execution.Context.Get(parts[1])
				if !exists {
					return nil, fmt.Errorf("context variable not found: %s", parts[1])
				}
				return val, nil
			}
		}
		return v, nil
		
	case map[string]interface{}:
		// Recursively resolve nested maps
		resolved := make(map[string]interface{})
		for k, v := range v {
			resolvedValue, err := o.resolveValue(ctx, execution, v)
			if err != nil {
				return nil, err
			}
			resolved[k] = resolvedValue
		}
		return resolved, nil
		
	case []interface{}:
		// Recursively resolve arrays
		resolved := make([]interface{}, len(v))
		for i, item := range v {
			resolvedValue, err := o.resolveValue(ctx, execution, item)
			if err != nil {
				return nil, err
			}
			resolved[i] = resolvedValue
		}
		return resolved, nil
		
	default:
		return value, nil
	}
}

func (o *WorkflowOrchestrator) executeWithRetry(ctx context.Context, execution *WorkflowExecution, step *WorkflowStepDef, inputs map[string]interface{}) (interface{}, error) {
	maxRetries := 1
	if step.Retry.MaxAttempts > 0 {
		maxRetries = step.Retry.MaxAttempts
	}
	
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		
		// Get tool from registry
		tool, exists := o.toolRegistry.Get(step.Tool)
		if !exists {
			return nil, fmt.Errorf("tool %s not found", step.Tool)
		}
		
		// Execute tool handler
		result, err := tool.Handler(ctx, inputs)
		if err == nil {
			return result, nil
		}
		
		lastErr = err
		o.logWarning(execution, fmt.Sprintf("Attempt %d failed for step %s: %v", attempt+1, step.Name, err))
	}
	
	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}

// Helper function for numeric comparison
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func compareNumeric(left, right float64, operator string) bool {
	switch operator {
	case ">":
		return left > right
	case ">=":
		return left >= right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case "==":
		return left == right
	case "!=":
		return left != right
	default:
		return false
	}
}

type StepCondition struct {
	Left     interface{} `json:"left"`
	Operator string      `json:"operator"`
	Right    interface{} `json:"right"`
}

type LoopConfig struct {
	MaxIterations   int            `json:"max_iterations"`
	ExitCondition   *StepCondition `json:"exit_condition,omitempty"`
	DelayBetween    time.Duration  `json:"delay_between"`
	ContinueOnError bool           `json:"continue_on_error"`
}

type SagaTransaction struct {
	Name         string           `json:"name"`
	Action       WorkflowStepDef  `json:"action"`
	Compensation *WorkflowStepDef `json:"compensation,omitempty"`
}

type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
}

type ErrorHandlingPolicy struct {
	Strategy       string `json:"strategy"` // fail_fast, continue, compensate
	MaxErrors      int    `json:"max_errors"`
	CompensateOnFail bool `json:"compensate_on_fail"`
}

type StepErrorHandling struct {
	OnError      string   `json:"on_error"` // skip, retry, fail, compensate
	FallbackStep string   `json:"fallback_step,omitempty"`
	ErrorTypes   []string `json:"error_types,omitempty"`
}

type WorkflowExecutionState string

const (
	StateCreated   WorkflowExecutionState = "created"
	StateRunning   WorkflowExecutionState = "running"
	StatePaused    WorkflowExecutionState = "paused"
	StateCompleted WorkflowExecutionState = "completed"
	StateFailed    WorkflowExecutionState = "failed"
	StateCancelled WorkflowExecutionState = "cancelled"
)

type ExecutionLogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	StepID    string                 `json:"step_id,omitempty"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type WorkflowTrigger struct {
	Type      string                 `json:"type"` // manual, scheduled, event
	Config    map[string]interface{} `json:"config"`
}

type WorkflowInput struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
}

type WorkflowOutput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Source      string `json:"source"` // step_id.output_path
	Description string `json:"description"`
}
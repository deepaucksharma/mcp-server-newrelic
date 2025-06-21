package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/PaesslerAG/jsonpath"
)

// AssertionEngine evaluates test assertions
type AssertionEngine struct {
	nrqlExecutor NRQLExecutor
	tolerance    float64
}

// NRQLExecutor interface for executing NRQL queries
type NRQLExecutor interface {
	ExecuteNRQL(ctx context.Context, query string) ([]map[string]interface{}, error)
}

// AssertionResult contains the result of an assertion
type AssertionResult struct {
	Type     string
	Expected interface{}
	Actual   interface{}
	Operator string
	Passed   bool
	Message  string
	Error    error
}

// NewAssertionEngine creates a new assertion engine
func NewAssertionEngine() *AssertionEngine {
	return &AssertionEngine{
		tolerance: 0.05, // 5% tolerance for approximate comparisons
	}
}

// SetNRQLExecutor sets the NRQL executor for query assertions
func (a *AssertionEngine) SetNRQLExecutor(executor NRQLExecutor) {
	a.nrqlExecutor = executor
}

// RunAssertions evaluates all assertions for a scenario
func (a *AssertionEngine) RunAssertions(ctx context.Context, assertions []Assertion, workflowResult *WorkflowResult) []AssertionResult {
	results := make([]AssertionResult, 0, len(assertions))

	for _, assertion := range assertions {
		result := a.evaluateAssertion(ctx, assertion, workflowResult)
		results = append(results, result)
	}

	return results
}

// evaluateAssertion evaluates a single assertion
func (a *AssertionEngine) evaluateAssertion(ctx context.Context, assertion Assertion, workflowResult *WorkflowResult) AssertionResult {
	result := AssertionResult{
		Type:     assertion.Type,
		Expected: assertion.Value,
		Operator: assertion.Operator,
		Message:  assertion.Message,
	}

	switch assertion.Type {
	case "jsonpath":
		a.evaluateJSONPathAssertion(&assertion, workflowResult, &result)
		
	case "nrql":
		a.evaluateNRQLAssertion(ctx, &assertion, &result)
		
	case "trace":
		a.evaluateTraceAssertion(&assertion, workflowResult, &result)
		
	case "custom":
		a.evaluateCustomAssertion(&assertion, workflowResult, &result)
		
	default:
		result.Error = fmt.Errorf("unknown assertion type: %s", assertion.Type)
		result.Passed = false
	}

	return result
}

// evaluateJSONPathAssertion evaluates a JSONPath-based assertion
func (a *AssertionEngine) evaluateJSONPathAssertion(assertion *Assertion, workflowResult *WorkflowResult, result *AssertionResult) {
	// Convert workflow result to JSON for JSONPath evaluation
	data, err := json.Marshal(workflowResult)
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal workflow result: %w", err)
		result.Passed = false
		return
	}

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		result.Error = fmt.Errorf("failed to unmarshal workflow result: %w", err)
		result.Passed = false
		return
	}

	// Evaluate JSONPath
	actualValue, err := jsonpath.Get(assertion.JSONPath, jsonData)
	if err != nil {
		result.Error = fmt.Errorf("failed to evaluate JSONPath %s: %w", assertion.JSONPath, err)
		result.Passed = false
		return
	}

	result.Actual = actualValue

	// Compare values
	result.Passed = a.compareValues(actualValue, assertion.Value, assertion.Operator)
	
	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("Expected %s %s %v, but got %v", 
			assertion.JSONPath, assertion.Operator, assertion.Value, actualValue)
	}
}

// evaluateNRQLAssertion evaluates an NRQL query assertion
func (a *AssertionEngine) evaluateNRQLAssertion(ctx context.Context, assertion *Assertion, result *AssertionResult) {
	if a.nrqlExecutor == nil {
		result.Error = fmt.Errorf("NRQL executor not configured")
		result.Passed = false
		return
	}

	// Execute NRQL query
	results, err := a.nrqlExecutor.ExecuteNRQL(ctx, assertion.Query)
	if err != nil {
		result.Error = fmt.Errorf("failed to execute NRQL query: %w", err)
		result.Passed = false
		return
	}

	// Extract actual value based on query results
	var actualValue interface{}
	if len(results) > 0 && len(results[0]) > 0 {
		// Get the first value from the first result
		for _, v := range results[0] {
			actualValue = v
			break
		}
	}

	result.Actual = actualValue

	// Compare with expected value
	result.Passed = a.compareValues(actualValue, assertion.Value, assertion.Operator)
	
	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("NRQL query result %v %s %v failed", 
			actualValue, assertion.Operator, assertion.Value)
	}
}

// evaluateTraceAssertion evaluates trace-related assertions
func (a *AssertionEngine) evaluateTraceAssertion(assertion *Assertion, workflowResult *WorkflowResult, result *AssertionResult) {
	// Handle different trace assertions
	switch assertion.Operator {
	case "trace_contains_confidence":
		// Check if trace contains confidence score
		confidence := extractConfidenceFromTrace(workflowResult)
		result.Actual = confidence
		
		expectedConfidence, err := parseFloat(assertion.Value)
		if err != nil {
			result.Error = fmt.Errorf("invalid confidence value: %w", err)
			result.Passed = false
			return
		}
		
		result.Passed = confidence >= expectedConfidence
		
	case "trace_has_discovery":
		// Check if trace shows discovery was performed
		hasDiscovery := checkTraceHasDiscovery(workflowResult)
		result.Actual = hasDiscovery
		result.Passed = hasDiscovery == assertion.Value
		
	default:
		result.Error = fmt.Errorf("unknown trace assertion operator: %s", assertion.Operator)
		result.Passed = false
	}
}

// evaluateCustomAssertion evaluates custom assertions
func (a *AssertionEngine) evaluateCustomAssertion(assertion *Assertion, workflowResult *WorkflowResult, result *AssertionResult) {
	// Custom assertions would be implemented based on specific needs
	result.Error = fmt.Errorf("custom assertions not yet implemented")
	result.Passed = false
}

// compareValues compares two values based on the operator
func (a *AssertionEngine) compareValues(actual, expected interface{}, operator string) bool {
	switch operator {
	case "==", "equals":
		return a.equalValues(actual, expected)
		
	case "!=", "not_equals":
		return !a.equalValues(actual, expected)
		
	case ">", "greater_than":
		return a.compareNumeric(actual, expected, func(a, b float64) bool { return a > b })
		
	case "<", "less_than":
		return a.compareNumeric(actual, expected, func(a, b float64) bool { return a < b })
		
	case ">=", "greater_than_or_equal":
		return a.compareNumeric(actual, expected, func(a, b float64) bool { return a >= b })
		
	case "<=", "less_than_or_equal":
		return a.compareNumeric(actual, expected, func(a, b float64) bool { return a <= b })
		
	case "contains":
		return a.containsValue(actual, expected)
		
	case "not_contains":
		return !a.containsValue(actual, expected)
		
	case "matches":
		return a.matchesPattern(actual, expected)
		
	case "approx", "approximately":
		return a.approximatelyEqual(actual, expected)
		
	default:
		return false
	}
}

// equalValues checks if two values are equal
func (a *AssertionEngine) equalValues(actual, expected interface{}) bool {
	// Handle nil cases
	if actual == nil || expected == nil {
		return actual == expected
	}

	// Use reflect.DeepEqual for complex types
	return reflect.DeepEqual(actual, expected)
}

// compareNumeric compares two numeric values
func (a *AssertionEngine) compareNumeric(actual, expected interface{}, compareFn func(float64, float64) bool) bool {
	actualNum, err1 := parseFloat(actual)
	expectedNum, err2 := parseFloat(expected)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return compareFn(actualNum, expectedNum)
}

// containsValue checks if actual contains expected
func (a *AssertionEngine) containsValue(actual, expected interface{}) bool {
	// String contains
	if actualStr, ok := actual.(string); ok {
		if expectedStr, ok := expected.(string); ok {
			return strings.Contains(actualStr, expectedStr)
		}
	}

	// Array contains
	if actualSlice, ok := actual.([]interface{}); ok {
		for _, item := range actualSlice {
			if a.equalValues(item, expected) {
				return true
			}
		}
	}

	// Map contains key
	if actualMap, ok := actual.(map[string]interface{}); ok {
		if key, ok := expected.(string); ok {
			_, exists := actualMap[key]
			return exists
		}
	}

	return false
}

// matchesPattern checks if actual matches the expected pattern
func (a *AssertionEngine) matchesPattern(actual, expected interface{}) bool {
	actualStr, ok1 := actual.(string)
	pattern, ok2 := expected.(string)
	
	if !ok1 || !ok2 {
		return false
	}
	
	matched, err := regexp.MatchString(pattern, actualStr)
	return err == nil && matched
}

// approximatelyEqual checks if two numeric values are approximately equal
func (a *AssertionEngine) approximatelyEqual(actual, expected interface{}) bool {
	actualNum, err1 := parseFloat(actual)
	expectedNum, err2 := parseFloat(expected)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	if expectedNum == 0 {
		return math.Abs(actualNum) < a.tolerance
	}
	
	relativeError := math.Abs((actualNum - expectedNum) / expectedNum)
	return relativeError <= a.tolerance
}

// Helper functions

func parseFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot parse %T as float", value)
	}
}

func extractConfidenceFromTrace(workflowResult *WorkflowResult) float64 {
	// Extract confidence from trace data or context
	if workflowResult.Context != nil {
		if confidence, ok := workflowResult.Context["confidence"].(float64); ok {
			return confidence
		}
	}
	
	// Check in step results
	for _, step := range workflowResult.Steps {
		if step.Result != nil {
			if confidence, ok := step.Result["confidence"].(float64); ok {
				return confidence
			}
		}
	}
	
	return 0.0
}

func checkTraceHasDiscovery(workflowResult *WorkflowResult) bool {
	// Check if any step is a discovery tool
	for _, step := range workflowResult.Steps {
		if strings.HasPrefix(step.Tool, "discovery.") {
			return true
		}
	}
	return false
}
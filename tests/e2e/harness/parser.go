package harness

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// WorkflowDSLParser parses workflow definitions
type WorkflowDSLParser struct {
	templateFuncs template.FuncMap
	variables     map[string]interface{}
}

// ParsedWorkflow represents a parsed workflow ready for execution
type ParsedWorkflow struct {
	Steps         []ParsedStep
	Variables     map[string]interface{}
	StepIndex     map[string]int // Maps store_as names to step indices
}

// ParsedStep represents a parsed workflow step
type ParsedStep struct {
	Tool          string
	Params        map[string]interface{}
	StoreAs       string
	Condition     *ParsedCondition
	Retry         *RetryConfig
	Timeout       time.Duration
	ParallelSteps []ParsedStep
	OnError       string
}

// ParsedCondition represents a parsed conditional expression
type ParsedCondition struct {
	Expression string
	Variables  []string
}

// NewWorkflowDSLParser creates a new parser
func NewWorkflowDSLParser() *WorkflowDSLParser {
	return &WorkflowDSLParser{
		templateFuncs: createTemplateFuncs(),
		variables:     make(map[string]interface{}),
	}
}

// Parse converts workflow steps into executable format
func (p *WorkflowDSLParser) Parse(steps []WorkflowStep) (*ParsedWorkflow, error) {
	workflow := &ParsedWorkflow{
		Steps:     make([]ParsedStep, 0, len(steps)),
		Variables: make(map[string]interface{}),
		StepIndex: make(map[string]int),
	}

	for i, step := range steps {
		parsed, err := p.parseStep(step, workflow)
		if err != nil {
			return nil, fmt.Errorf("failed to parse step %d: %w", i, err)
		}

		workflow.Steps = append(workflow.Steps, parsed)
		
		// Track step indices for variable resolution
		if parsed.StoreAs != "" {
			workflow.StepIndex[parsed.StoreAs] = i
		}
	}

	return workflow, nil
}

// parseStep parses a single workflow step
func (p *WorkflowDSLParser) parseStep(step WorkflowStep, workflow *ParsedWorkflow) (ParsedStep, error) {
	parsed := ParsedStep{
		Tool:    step.Tool,
		StoreAs: step.StoreAs,
		Retry:   step.Retry,
		Timeout: step.Timeout,
		OnError: step.OnError,
	}

	// Parse parameters with variable substitution
	params, err := p.parseParams(step.Params, workflow.Variables)
	if err != nil {
		return parsed, fmt.Errorf("failed to parse params: %w", err)
	}
	parsed.Params = params

	// Parse condition if present
	if step.Condition != "" {
		condition, err := p.parseCondition(step.Condition)
		if err != nil {
			return parsed, fmt.Errorf("failed to parse condition: %w", err)
		}
		parsed.Condition = condition
	}

	// Parse parallel steps
	if len(step.Parallel) > 0 {
		parallelSteps := make([]ParsedStep, 0, len(step.Parallel))
		for _, pStep := range step.Parallel {
			pParsed, err := p.parseStep(pStep, workflow)
			if err != nil {
				return parsed, fmt.Errorf("failed to parse parallel step: %w", err)
			}
			parallelSteps = append(parallelSteps, pParsed)
		}
		parsed.ParallelSteps = parallelSteps
	}

	return parsed, nil
}

// parseParams processes parameters with variable substitution and special syntax
func (p *WorkflowDSLParser) parseParams(params map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	parsed := make(map[string]interface{})

	for key, value := range params {
		processedValue, err := p.processValue(value, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to process param %s: %w", key, err)
		}
		parsed[key] = processedValue
	}

	return parsed, nil
}

// processValue handles variable substitution and special functions
func (p *WorkflowDSLParser) processValue(value interface{}, variables map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Handle variable references: ${variable_name}
		if strings.Contains(v, "${") {
			return p.expandVariables(v, variables)
		}
		
		// Handle adaptive query builder: ${aqb:function}
		if strings.HasPrefix(v, "${aqb:") {
			return p.processAQB(v, variables)
		}
		
		// Handle template functions: ${fn:function(args)}
		if strings.HasPrefix(v, "${fn:") {
			return p.processFunction(v, variables)
		}
		
		return v, nil

	case map[string]interface{}:
		// Recursively process nested maps
		processed := make(map[string]interface{})
		for k, val := range v {
			pVal, err := p.processValue(val, variables)
			if err != nil {
				return nil, err
			}
			processed[k] = pVal
		}
		return processed, nil

	case []interface{}:
		// Recursively process arrays
		processed := make([]interface{}, len(v))
		for i, val := range v {
			pVal, err := p.processValue(val, variables)
			if err != nil {
				return nil, err
			}
			processed[i] = pVal
		}
		return processed, nil

	default:
		// Return other types as-is
		return value, nil
	}
}

// expandVariables replaces ${var} references with actual values
func (p *WorkflowDSLParser) expandVariables(str string, variables map[string]interface{}) (string, error) {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	
	result := re.ReplaceAllStringFunc(str, func(match string) string {
		varName := match[2 : len(match)-1] // Remove ${ and }
		
		// Handle nested references like ${steps[0].result}
		value, err := p.resolveVariable(varName, variables)
		if err != nil {
			return match // Keep original if resolution fails
		}
		
		return fmt.Sprintf("%v", value)
	})
	
	return result, nil
}

// resolveVariable resolves a variable reference, supporting nested paths
func (p *WorkflowDSLParser) resolveVariable(path string, variables map[string]interface{}) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := variables
	
	for i, part := range parts {
		// Handle array indices
		if strings.Contains(part, "[") {
			// Parse array access
			arrayPart := strings.Split(part, "[")
			key := arrayPart[0]
			indexStr := strings.TrimSuffix(arrayPart[1], "]")
			
			// Get the array
			arrVal, ok := current[key]
			if !ok {
				return nil, fmt.Errorf("variable %s not found", key)
			}
			
			// Access array element
			// TODO: Implement array access logic using indexStr and arrVal
			_ = indexStr // Mark as used
			_ = arrVal   // Mark as used
		} else {
			// Simple map access
			val, ok := current[part]
			if !ok {
				return nil, fmt.Errorf("variable %s not found in path %s", part, path)
			}
			
			// If this is the last part, return the value
			if i == len(parts)-1 {
				return val, nil
			}
			
			// Otherwise, continue traversing
			if nextMap, ok := val.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil, fmt.Errorf("cannot traverse into non-map value at %s", part)
			}
		}
	}
	
	return current, nil
}

// processAQB handles adaptive query builder syntax
func (p *WorkflowDSLParser) processAQB(expr string, variables map[string]interface{}) (string, error) {
	// Extract AQB function: ${aqb:build.latency_p95}
	match := regexp.MustCompile(`\$\{aqb:([^}]+)\}`).FindStringSubmatch(expr)
	if len(match) != 2 {
		return expr, fmt.Errorf("invalid AQB expression: %s", expr)
	}
	
	aqbFunc := match[1]
	
	// Handle different AQB functions
	switch {
	case strings.HasPrefix(aqbFunc, "build."):
		queryType := strings.TrimPrefix(aqbFunc, "build.")
		return p.buildAdaptiveQuery(queryType, variables)
		
	case strings.HasPrefix(aqbFunc, "discover."):
		discoveryType := strings.TrimPrefix(aqbFunc, "discover.")
		return p.buildDiscoveryQuery(discoveryType, variables)
		
	default:
		return expr, fmt.Errorf("unknown AQB function: %s", aqbFunc)
	}
}

// buildAdaptiveQuery generates an adaptive NRQL query based on discovered schema
func (p *WorkflowDSLParser) buildAdaptiveQuery(queryType string, variables map[string]interface{}) (string, error) {
	// This would use discovered schema information to build queries
	// For now, return example queries
	
	switch queryType {
	case "latency_p95":
		// Check if we discovered the duration attribute name
		durationAttr := "duration" // Default
		if discovered, ok := variables["discovered_duration_attr"].(string); ok {
			durationAttr = discovered
		}
		return fmt.Sprintf("SELECT percentile(%s, 95) FROM Transaction SINCE 1 hour ago", durationAttr), nil
		
	case "error_rate":
		// Check discovered error attribute
		errorAttr := "error"
		if discovered, ok := variables["discovered_error_attr"].(string); ok {
			errorAttr = discovered
		}
		return fmt.Sprintf("SELECT percentage(count(*), WHERE %s IS TRUE) FROM Transaction SINCE 1 hour ago", errorAttr), nil
		
	default:
		return "", fmt.Errorf("unknown query type: %s", queryType)
	}
}

// buildDiscoveryQuery generates a discovery query
func (p *WorkflowDSLParser) buildDiscoveryQuery(discoveryType string, variables map[string]interface{}) (string, error) {
	switch discoveryType {
	case "event_types":
		return "SHOW EVENT TYPES", nil
		
	case "attributes":
		eventType := "Transaction" // Default
		if et, ok := variables["event_type"].(string); ok {
			eventType = et
		}
		return fmt.Sprintf("SELECT keyset() FROM %s LIMIT 1", eventType), nil
		
	default:
		return "", fmt.Errorf("unknown discovery type: %s", discoveryType)
	}
}

// processFunction handles template function calls
func (p *WorkflowDSLParser) processFunction(expr string, variables map[string]interface{}) (interface{}, error) {
	// Extract function: ${fn:now()}
	match := regexp.MustCompile(`\$\{fn:([^(]+)\(([^)]*)\)\}`).FindStringSubmatch(expr)
	if len(match) != 3 {
		return expr, fmt.Errorf("invalid function expression: %s", expr)
	}
	
	funcName := match[1]
	args := match[2]
	
	// Execute template function
	tmpl := template.New("func").Funcs(p.templateFuncs)
	tmplStr := fmt.Sprintf(`{{ %s %s }}`, funcName, args)
	
	_, err := tmpl.Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template function: %w", err)
	}
	
	var buf strings.Builder
	err = tmpl.Execute(&buf, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template function: %w", err)
	}
	
	return buf.String(), nil
}

// parseCondition parses a conditional expression
func (p *WorkflowDSLParser) parseCondition(expr string) (*ParsedCondition, error) {
	// Extract variables used in condition
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(expr, -1)
	
	variables := make([]string, 0, len(matches))
	for _, match := range matches {
		variables = append(variables, match[1])
	}
	
	return &ParsedCondition{
		Expression: expr,
		Variables:  variables,
	}, nil
}

// createTemplateFuncs creates template functions for DSL
func createTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"now": func() string {
			return time.Now().Format(time.RFC3339)
		},
		"ago": func(duration string) string {
			d, _ := time.ParseDuration(duration)
			return time.Now().Add(-d).Format(time.RFC3339)
		},
		"env": func(key string) string {
			return os.Getenv(key)
		},
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" {
				return def
			}
			return val
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"trim":  strings.TrimSpace,
	}
}
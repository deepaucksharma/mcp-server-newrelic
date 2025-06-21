package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/interface/mcp"
)

// MCPTestClient provides a test-friendly MCP client
type MCPTestClient struct {
	server     *mcp.Server
	account    *TestAccount
	discovery  map[string]interface{}
	mu         sync.RWMutex
	requestLog []RequestLog
}

// RequestLog captures request/response for debugging
type RequestLog struct {
	Timestamp time.Time
	Tool      string
	Params    map[string]interface{}
	Response  interface{}
	Error     error
	Duration  time.Duration
}

// NewMCPTestClient creates a test client for the given account
func NewMCPTestClient(account *TestAccount) *MCPTestClient {
	config := mcp.ServerConfig{
		APIKey:    account.APIKey,
		AccountID: account.AccountID,
		Region:    account.Region,
		Timeout:   30 * time.Second,
	}
	
	server := mcp.NewServer(config)
	
	return &MCPTestClient{
		server:    server,
		account:   account,
		discovery: make(map[string]interface{}),
	}
}

// ExecuteTool executes an MCP tool and returns the result
func (c *MCPTestClient) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
	start := time.Now()
	
	// Build JSON-RPC request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": params,
		},
		"id": fmt.Sprintf("test-%d", time.Now().UnixNano()),
	}
	
	// Execute through MCP protocol handler
	response, err := c.server.HandleRequest(ctx, request)
	
	duration := time.Since(start)
	
	// Log request for debugging
	c.logRequest(toolName, params, response, err, duration)
	
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}
	
	// Extract result from JSON-RPC response
	if respMap, ok := response.(map[string]interface{}); ok {
		if result, exists := respMap["result"]; exists {
			if resultMap, ok := result.(map[string]interface{}); ok {
				return resultMap, nil
			}
		}
		if errorData, exists := respMap["error"]; exists {
			return nil, fmt.Errorf("MCP error: %v", errorData)
		}
	}
	
	return nil, fmt.Errorf("unexpected response format: %T", response)
}

// ExecuteWorkflow runs a multi-step workflow
func (c *MCPTestClient) ExecuteWorkflow(ctx context.Context, steps []WorkflowStep) (*WorkflowResult, error) {
	result := &WorkflowResult{
		Steps:     make([]StepResult, 0, len(steps)),
		StartTime: time.Now(),
	}
	
	workflowCtx := make(map[string]interface{})
	
	for i, step := range steps {
		stepResult := StepResult{
			Index:     i,
			Name:      step.Name,
			Tool:      step.Tool,
			StartTime: time.Now(),
		}
		
		// Resolve parameters with context
		params := c.resolveParameters(step.Params, workflowCtx)
		
		// Execute step
		response, err := c.ExecuteTool(ctx, step.Tool, params)
		stepResult.EndTime = time.Now()
		stepResult.Duration = stepResult.EndTime.Sub(stepResult.StartTime)
		
		if err != nil {
			stepResult.Error = err
			result.Steps = append(result.Steps, stepResult)
			
			if !step.ContinueOnError {
				return result, fmt.Errorf("step %d (%s) failed: %w", i, step.Name, err)
			}
			continue
		}
		
		stepResult.Response = response
		result.Steps = append(result.Steps, stepResult)
		
		// Update context with step results
		if step.StoreAs != "" {
			workflowCtx[step.StoreAs] = response
		}
		
		// Validate step if validator provided
		if step.Validate != nil {
			if err := step.Validate(response); err != nil {
				return result, fmt.Errorf("step %d (%s) validation failed: %w", i, step.Name, err)
			}
		}
	}
	
	result.EndTime = time.Now()
	result.TotalDuration = result.EndTime.Sub(result.StartTime)
	result.Success = true
	
	return result, nil
}

// StoreDiscovery stores discovered information for later use
func (c *MCPTestClient) StoreDiscovery(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.discovery[key] = value
}

// GetDiscovery retrieves previously discovered information
func (c *MCPTestClient) GetDiscovery(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists := c.discovery[key]
	return val, exists
}

// GetRequestLog returns the request history for debugging
func (c *MCPTestClient) GetRequestLog() []RequestLog {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	log := make([]RequestLog, len(c.requestLog))
	copy(log, c.requestLog)
	return log
}

// ClearRequestLog clears the request history
func (c *MCPTestClient) ClearRequestLog() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestLog = nil
}

// Close cleans up the client resources
func (c *MCPTestClient) Close() error {
	// Any cleanup needed
	return nil
}

// Private methods

func (c *MCPTestClient) logRequest(tool string, params map[string]interface{}, response interface{}, err error, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.requestLog = append(c.requestLog, RequestLog{
		Timestamp: time.Now(),
		Tool:      tool,
		Params:    params,
		Response:  response,
		Error:     err,
		Duration:  duration,
	})
}

func (c *MCPTestClient) resolveParameters(params map[string]interface{}, context map[string]interface{}) map[string]interface{} {
	resolved := make(map[string]interface{})
	
	for key, value := range params {
		switch v := value.(type) {
		case string:
			// Check if it's a context reference
			if len(v) > 2 && v[0] == '$' && v[1] == '{' && v[len(v)-1] == '}' {
				contextKey := v[2 : len(v)-1]
				if contextValue, exists := context[contextKey]; exists {
					resolved[key] = contextValue
				} else {
					resolved[key] = v // Keep original if not found
				}
			} else {
				resolved[key] = v
			}
		default:
			resolved[key] = value
		}
	}
	
	return resolved
}

// WorkflowStep defines a step in a workflow
type WorkflowStep struct {
	Name            string
	Tool            string
	Params          map[string]interface{}
	StoreAs         string
	ContinueOnError bool
	Validate        func(response map[string]interface{}) error
}

// WorkflowResult contains the results of a workflow execution
type WorkflowResult struct {
	Steps         []StepResult
	StartTime     time.Time
	EndTime       time.Time
	TotalDuration time.Duration
	Success       bool
}

// StepResult contains the result of a single workflow step
type StepResult struct {
	Index     int
	Name      string
	Tool      string
	Response  map[string]interface{}
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// ExportToJSON exports the workflow result as JSON
func (r *WorkflowResult) ExportToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
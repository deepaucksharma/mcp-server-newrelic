//go:build deprecated
// +build deprecated

// DEPRECATED: Enhanced protocol features have been merged into protocol.go
// This file is kept for reference but should not be used.
// All enhanced features are now the default behavior.

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/logger"
)

// Enhanced JSON-RPC 2.0 error codes for better compliance
const (
	// Standard JSON-RPC 2.0 error codes
	ParseErrorCode      = -32700
	InvalidRequestCode  = -32600
	MethodNotFoundCode  = -32601
	InvalidParamsCode   = -32602
	InternalErrorCode   = -32603
	
	// MCP-specific error codes
	ToolNotFoundCode    = -32001
	ToolExecutionCode   = -32002
	SessionNotFoundCode = -32003
	TimeoutErrorCode    = -32004
	RateLimitCode       = -32005
)

// NotificationHandler handles JSON-RPC 2.0 notifications (requests without ID)
func (h *ProtocolHandler) HandleNotification(ctx context.Context, req Request) {
	// Notifications don't expect responses
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Notification handler panic: %v", r)
		}
	}()

	switch req.Method {
	case "tools/changed":
		// Handle tool registry changes
		logger.Info("Tools changed notification received")
	case "sessions/ended":
		// Handle session termination
		var params struct {
			SessionID string `json:"sessionId"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			h.server.sessions.End(params.SessionID)
		}
	case "cancel":
		// Handle request cancellation
		var params struct {
			RequestID interface{} `json:"requestId"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			h.cancelRequest(params.RequestID)
		}
	default:
		logger.Debug("Unknown notification: %s", req.Method)
	}
}

// EnhancedHandleMessage provides full JSON-RPC 2.0 compliance
func (h *ProtocolHandler) EnhancedHandleMessage(ctx context.Context, message []byte) ([]byte, error) {
	// Validate JSON
	if !json.Valid(message) {
		return h.enhancedErrorResponse(nil, ParseErrorCode, "Parse error", 
			map[string]string{"detail": "Invalid JSON"})
	}

	// Check if it's a batch request
	message = bytes.TrimSpace(message)
	if len(message) > 0 && message[0] == '[' {
		return h.handleEnhancedBatch(ctx, message)
	}

	// Parse single request
	var req Request
	if err := json.Unmarshal(message, &req); err != nil {
		return h.enhancedErrorResponse(nil, ParseErrorCode, "Parse error", 
			map[string]string{"detail": err.Error()})
	}

	// Handle notification (no ID field)
	if req.ID == nil {
		h.HandleNotification(ctx, req)
		// Notifications don't get responses
		return nil, nil
	}

	// Process request with enhanced error handling
	return h.processEnhancedRequest(ctx, req)
}

// processEnhancedRequest handles a single request with full error handling
func (h *ProtocolHandler) processEnhancedRequest(ctx context.Context, req Request) ([]byte, error) {
	// Validate JSON-RPC version
	if req.Jsonrpc != "2.0" {
		return h.enhancedErrorResponse(req.ID, InvalidRequestCode, 
			"Invalid Request", map[string]string{
				"detail": "JSON-RPC version must be 2.0",
				"received": req.Jsonrpc,
			})
	}

	// Validate method
	if req.Method == "" {
		return h.enhancedErrorResponse(req.ID, InvalidRequestCode,
			"Invalid Request", map[string]string{
				"detail": "Method is required",
			})
	}

	// Add request tracking
	if err := h.trackRequest(req.ID); err != nil {
		return h.enhancedErrorResponse(req.ID, RateLimitCode,
			"Rate limit exceeded", map[string]interface{}{
				"retryAfter": 60,
				"limit": h.server.config.RateLimit,
			})
	}
	defer h.untrackRequest(req.ID)

	// Route to appropriate handler with timeout
	ctx, cancel := context.WithTimeout(ctx, h.server.config.RequestTimeout)
	defer cancel()

	// Execute in a goroutine to handle timeout properly
	type result struct {
		data []byte
		err  error
	}
	
	resultChan := make(chan result, 1)
	go func() {
		data, err := h.routeRequest(ctx, req)
		resultChan <- result{data, err}
	}()

	select {
	case r := <-resultChan:
		return r.data, r.err
	case <-ctx.Done():
		return h.enhancedErrorResponse(req.ID, TimeoutErrorCode,
			"Request timeout", map[string]interface{}{
				"timeout": h.server.config.RequestTimeout.String(),
				"method": req.Method,
			})
	}
}

// routeRequest routes the request to the appropriate handler
func (h *ProtocolHandler) routeRequest(ctx context.Context, req Request) ([]byte, error) {
	// Standard MCP methods
	switch req.Method {
	case "initialize":
		return h.handleEnhancedInitialize(ctx, req)
	case "initialized":
		// Client confirms initialization
		return h.successResponse(req.ID, map[string]bool{"success": true})
	case "shutdown":
		return h.handleShutdown(ctx, req)
	case "tools/list":
		return h.handleEnhancedToolsList(ctx, req)
	case "tools/call":
		return h.handleEnhancedToolCall(ctx, req)
	case "completion/complete":
		return h.handleEnhancedCompletion(ctx, req)
	case "sessions/create":
		return h.handleSessionCreate(ctx, req)
	case "sessions/get":
		return h.handleSessionGet(ctx, req)
	case "sessions/list":
		return h.handleSessionList(ctx, req)
	case "prompts/list":
		return h.handlePromptsList(ctx, req)
	case "prompts/get":
		return h.handlePromptsGet(ctx, req)
	case "resources/list":
		return h.handleResourcesList(ctx, req)
	case "resources/read":
		return h.handleResourcesRead(ctx, req)
	default:
		// Check if it's a direct tool call
		if tool, exists := h.server.tools.Get(req.Method); exists {
			return h.handleDirectToolCall(ctx, req, tool)
		}
		
		return h.enhancedErrorResponse(req.ID, MethodNotFoundCode,
			fmt.Sprintf("Method not found: %s", req.Method),
			map[string]interface{}{
				"availableMethods": h.getAvailableMethods(),
			})
	}
}

// handleEnhancedInitialize provides full MCP capabilities
func (h *ProtocolHandler) handleEnhancedInitialize(ctx context.Context, req Request) ([]byte, error) {
	var params struct {
		ProtocolVersion string `json:"protocolVersion"`
		Capabilities    struct {
			Tools struct {
				ListChanged bool `json:"listChanged"`
			} `json:"tools"`
			Completion struct{} `json:"completion"`
			Prompts    struct{} `json:"prompts"`
			Resources  struct{} `json:"resources"`
		} `json:"capabilities"`
		ClientInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}

	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return h.enhancedErrorResponse(req.ID, InvalidParamsCode,
				"Invalid parameters", map[string]string{
					"detail": err.Error(),
				})
		}
	}

	// Store client info
	if req.ID != nil {
		h.server.sessions.StoreClientInfo(fmt.Sprintf("%v", req.ID), params.ClientInfo)
	}

	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
			"completion": map[string]interface{}{},
			"prompts": map[string]interface{}{
				"listChanged": true,
			},
			"resources": map[string]interface{}{
				"subscribe":    true,
				"listChanged": true,
			},
			"logging": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    h.server.GetInfo()["name"],
			"version": h.server.GetInfo()["version"],
			"vendor":  "New Relic",
		},
	}

	return h.successResponse(req.ID, result)
}

// handleShutdown handles graceful shutdown
func (h *ProtocolHandler) handleShutdown(ctx context.Context, req Request) ([]byte, error) {
	// Initiate graceful shutdown
	go func() {
		time.Sleep(100 * time.Millisecond) // Allow response to be sent
		h.server.Shutdown()
	}()

	return h.successResponse(req.ID, map[string]bool{"success": true})
}

// handleEnhancedToolsList provides detailed tool information
func (h *ProtocolHandler) handleEnhancedToolsList(ctx context.Context, req Request) ([]byte, error) {
	tools := h.server.tools.ListEnhanced()
	
	toolSchemas := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		schema := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": map[string]interface{}{
				"type":       tool.Parameters.Type,
				"properties": tool.Parameters.Properties,
				"required":   tool.Parameters.Required,
			},
		}

		// Add enhanced metadata if available
		if enhanced, ok := tool.(*EnhancedTool); ok {
			schema["metadata"] = map[string]interface{}{
				"category":    enhanced.Category,
				"safety":      enhanced.Safety,
				"performance": enhanced.Performance,
				"examples":    enhanced.Examples,
			}
		}

		toolSchemas[i] = schema
	}

	result := map[string]interface{}{
		"tools": toolSchemas,
		"_meta": map[string]interface{}{
			"total":      len(toolSchemas),
			"categories": h.server.tools.GetCategories(),
		},
	}

	return h.successResponse(req.ID, result)
}

// handleEnhancedToolCall executes a tool with full error handling
func (h *ProtocolHandler) handleEnhancedToolCall(ctx context.Context, req Request) ([]byte, error) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.enhancedErrorResponse(req.ID, InvalidParamsCode,
			"Invalid parameters", map[string]interface{}{
				"detail": err.Error(),
				"expected": map[string]string{
					"name":      "string",
					"arguments": "object",
					"stream":    "boolean (optional)",
				},
			})
	}

	// Validate required fields
	if params.Name == "" {
		return h.enhancedErrorResponse(req.ID, InvalidParamsCode,
			"Invalid parameters", map[string]string{
				"detail": "Tool name is required",
			})
	}

	// Get tool from registry
	tool, exists := h.server.tools.Get(params.Name)
	if !exists {
		return h.enhancedErrorResponse(req.ID, ToolNotFoundCode,
			fmt.Sprintf("Tool '%s' not found", params.Name),
			map[string]interface{}{
				"availableTools": h.server.tools.ListNames(),
				"suggestion":     h.findSimilarTool(params.Name),
			})
	}

	// Validate parameters against schema
	if err := h.validateToolParams(tool, params.Arguments); err != nil {
		return h.enhancedErrorResponse(req.ID, InvalidParamsCode,
			"Invalid tool parameters", map[string]interface{}{
				"detail":   err.Error(),
				"schema":   tool.Parameters,
				"received": params.Arguments,
			})
	}

	// Create execution context with metadata
	execCtx := &ExecutionContext{
		RequestID: req.ID,
		Tool:      tool,
		StartTime: time.Now(),
		Metadata: map[string]interface{}{
			"clientInfo": h.server.sessions.GetClientInfo(fmt.Sprintf("%v", req.ID)),
			"streaming":  params.Stream,
		},
	}

	// Execute tool with instrumentation
	start := time.Now()
	result, err := h.executeToolWithInstrumentation(ctx, execCtx, params)
	duration := time.Since(start)

	// Record metrics if available
	if metrics, ok := h.server.getMetrics(); ok {
		metrics.RecordToolExecution(tool.Name, duration, err == nil)
	}

	if err != nil {
		// Determine error code based on error type
		code := InternalErrorCode
		if strings.Contains(err.Error(), "not found") {
			code = InvalidParamsCode
		} else if strings.Contains(err.Error(), "timeout") {
			code = TimeoutErrorCode
		}

		return h.enhancedErrorResponse(req.ID, code,
			fmt.Sprintf("Tool execution failed: %v", err),
			map[string]interface{}{
				"tool":     tool.Name,
				"duration": duration.String(),
				"hint":     h.getErrorHint(err),
			})
	}

	// Format response with metadata
	response := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": formatToolResult(result),
			},
		},
		"_meta": map[string]interface{}{
			"executionTime": duration.Milliseconds(),
			"tool":          tool.Name,
			"cached":        execCtx.Metadata["cached"],
		},
	}

	return h.successResponse(req.ID, response)
}

// validateToolParams validates tool parameters against schema
func (h *ProtocolHandler) validateToolParams(tool *Tool, params map[string]interface{}) error {
	// Check required parameters
	for _, required := range tool.Parameters.Required {
		if _, exists := params[required]; !exists {
			return fmt.Errorf("required parameter '%s' is missing", required)
		}
	}

	// Validate parameter types
	for name, value := range params {
		prop, exists := tool.Parameters.Properties[name]
		if !exists {
			// Extra parameters are allowed unless strict mode
			continue
		}

		if err := h.validateParamType(name, value, prop); err != nil {
			return err
		}
	}

	return nil
}

// validateParamType validates a single parameter type
func (h *ProtocolHandler) validateParamType(name string, value interface{}, prop Property) error {
	switch prop.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter '%s' must be a string", name)
		}
		// Check enum values if defined
		if len(prop.Enum) > 0 {
			strVal := value.(string)
			valid := false
			for _, enum := range prop.Enum {
				if strVal == enum {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("parameter '%s' must be one of: %v", name, prop.Enum)
			}
		}
	case "number", "integer":
		switch v := value.(type) {
		case float64:
			if prop.Type == "integer" && v != float64(int(v)) {
				return fmt.Errorf("parameter '%s' must be an integer", name)
			}
		case int:
			// OK
		default:
			return fmt.Errorf("parameter '%s' must be a number", name)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("parameter '%s' must be a boolean", name)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("parameter '%s' must be an array", name)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("parameter '%s' must be an object", name)
		}
	}

	return nil
}

// enhancedErrorResponse creates a detailed error response
func (h *ProtocolHandler) enhancedErrorResponse(id interface{}, code int, message string, data interface{}) ([]byte, error) {
	resp := Response{
		Jsonrpc: "2.0",
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
	return json.Marshal(resp)
}

// Helper methods

func (h *ProtocolHandler) trackRequest(id interface{}) error {
	// Implement rate limiting
	return nil
}

func (h *ProtocolHandler) untrackRequest(id interface{}) {
	// Clean up request tracking
}

func (h *ProtocolHandler) cancelRequest(id interface{}) {
	// Implement request cancellation
}

func (h *ProtocolHandler) getAvailableMethods() []string {
	methods := []string{
		"initialize", "initialized", "shutdown",
		"tools/list", "tools/call",
		"completion/complete",
		"sessions/create", "sessions/get", "sessions/list",
		"prompts/list", "prompts/get",
		"resources/list", "resources/read",
	}
	
	// Add all tool names as direct methods
	for _, tool := range h.server.tools.ListNames() {
		methods = append(methods, tool)
	}
	
	return methods
}

func (h *ProtocolHandler) findSimilarTool(name string) string {
	// Simple similarity check
	tools := h.server.tools.ListNames()
	for _, tool := range tools {
		if strings.Contains(strings.ToLower(tool), strings.ToLower(name)) ||
		   strings.Contains(strings.ToLower(name), strings.ToLower(tool)) {
			return tool
		}
	}
	return ""
}

func (h *ProtocolHandler) getErrorHint(err error) string {
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "timeout"):
		return "Try reducing the time range or query complexity"
	case strings.Contains(errStr, "not found"):
		return "Verify the resource exists and you have access"
	case strings.Contains(errStr, "invalid"):
		return "Check the parameter format and types"
	default:
		return "Check logs for more details"
	}
}

func (h *ProtocolHandler) executeToolWithInstrumentation(ctx context.Context, execCtx *ExecutionContext, params ToolCallParams) (interface{}, error) {
	// Add instrumentation
	ctx = context.WithValue(ctx, "executionContext", execCtx)
	
	// Check cache if applicable
	if cache, ok := h.server.getCache(); ok && !params.NoCache {
		if cached, exists := cache.Get(h.getCacheKey(params)); exists {
			execCtx.Metadata["cached"] = true
			return cached, nil
		}
	}
	
	// Execute tool
	result, err := execCtx.Tool.Handler(ctx, params.Arguments)
	
	// Cache successful results
	if err == nil {
		if cache, ok := h.server.getCache(); ok {
			cache.Set(h.getCacheKey(params), result, 5*time.Minute)
		}
	}
	
	return result, err
}

func (h *ProtocolHandler) getCacheKey(params ToolCallParams) string {
	// Generate cache key from tool name and arguments
	data, _ := json.Marshal(params)
	return string(data)
}

// Placeholder handlers for additional MCP methods

func (h *ProtocolHandler) handleSessionList(ctx context.Context, req Request) ([]byte, error) {
	sessions := h.server.sessions.List()
	return h.successResponse(req.ID, map[string]interface{}{
		"sessions": sessions,
	})
}

func (h *ProtocolHandler) handlePromptsList(ctx context.Context, req Request) ([]byte, error) {
	// TODO: Implement prompts support
	return h.successResponse(req.ID, map[string]interface{}{
		"prompts": []interface{}{},
	})
}

func (h *ProtocolHandler) handlePromptsGet(ctx context.Context, req Request) ([]byte, error) {
	// TODO: Implement prompts support
	return h.enhancedErrorResponse(req.ID, MethodNotFoundCode,
		"Prompts not yet implemented", nil)
}

func (h *ProtocolHandler) handleResourcesList(ctx context.Context, req Request) ([]byte, error) {
	// TODO: Implement resources support
	return h.successResponse(req.ID, map[string]interface{}{
		"resources": []interface{}{},
	})
}

func (h *ProtocolHandler) handleResourcesRead(ctx context.Context, req Request) ([]byte, error) {
	// TODO: Implement resources support
	return h.enhancedErrorResponse(req.ID, MethodNotFoundCode,
		"Resources not yet implemented", nil)
}

func (h *ProtocolHandler) handleDirectToolCall(ctx context.Context, req Request, tool *Tool) ([]byte, error) {
	// Convert direct tool call to standard format
	var params map[string]interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.enhancedErrorResponse(req.ID, InvalidParamsCode,
			"Invalid parameters", map[string]string{
				"detail": err.Error(),
			})
	}
	
	toolCallReq := Request{
		Jsonrpc: req.Jsonrpc,
		Method:  "tools/call",
		Params: mustMarshal(ToolCallParams{
			Name:      tool.Name,
			Arguments: params,
		}),
		ID: req.ID,
	}
	
	return h.handleEnhancedToolCall(ctx, toolCallReq)
}

func (h *ProtocolHandler) handleEnhancedCompletion(ctx context.Context, req Request) ([]byte, error) {
	// Enhanced completion with context awareness
	var params struct {
		Ref struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"ref"`
		Argument struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"argument"`
		Context map[string]interface{} `json:"context"`
	}
	
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.enhancedErrorResponse(req.ID, InvalidParamsCode,
			"Invalid parameters", map[string]string{
				"detail": err.Error(),
			})
	}
	
	// Get context-aware completions
	completions := h.getContextAwareCompletions(params)
	
	result := map[string]interface{}{
		"completion": map[string]interface{}{
			"values":  completions,
			"total":   len(completions),
			"hasMore": false,
		},
	}
	
	return h.successResponse(req.ID, result)
}

func (h *ProtocolHandler) getContextAwareCompletions(params interface{}) []map[string]interface{} {
	// TODO: Implement context-aware completions
	return []map[string]interface{}{}
}

// handleEnhancedBatch processes batch requests with proper ordering
func (h *ProtocolHandler) handleEnhancedBatch(ctx context.Context, message []byte) ([]byte, error) {
	var batchReq []Request
	if err := json.Unmarshal(message, &batchReq); err != nil {
		return h.enhancedErrorResponse(nil, ParseErrorCode,
			"Parse error", map[string]string{
				"detail": "Invalid batch request format",
			})
	}

	if len(batchReq) == 0 {
		return h.enhancedErrorResponse(nil, InvalidRequestCode,
			"Invalid Request", map[string]string{
				"detail": "Empty batch",
			})
	}

	// Process requests maintaining order for those with IDs
	responses := make([]json.RawMessage, 0, len(batchReq))
	var wg sync.WaitGroup
	responseChan := make(chan struct {
		index int
		data  []byte
	}, len(batchReq))

	for i, req := range batchReq {
		if req.ID == nil {
			// Handle notifications immediately (no response)
			go h.HandleNotification(ctx, req)
			continue
		}

		wg.Add(1)
		go func(idx int, r Request) {
			defer wg.Done()
			data, _ := h.processEnhancedRequest(ctx, r)
			if data != nil {
				responseChan <- struct {
					index int
					data  []byte
				}{idx, data}
			}
		}(i, req)
	}

	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// Collect responses maintaining order
	responseMap := make(map[int][]byte)
	for resp := range responseChan {
		responseMap[resp.index] = resp.data
	}

	// Build ordered response array
	for i := 0; i < len(batchReq); i++ {
		if data, exists := responseMap[i]; exists {
			responses = append(responses, json.RawMessage(data))
		}
	}

	// Return null for empty response array
	if len(responses) == 0 {
		return nil, nil
	}

	return json.Marshal(responses)
}
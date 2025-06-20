package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
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

// ProtocolHandler implements the JSON-RPC 2.0 protocol for MCP
type ProtocolHandler struct {
	server    *Server
	requests  sync.Map // Track in-flight requests
	idCounter int64
}

// HandleMessage processes incoming JSON-RPC messages (single or batch)
func (h *ProtocolHandler) HandleMessage(ctx context.Context, message []byte) ([]byte, error) {
	// Validate JSON
	if !json.Valid(message) {
		return h.errorResponse(nil, ParseErrorCode, "Parse error", 
			map[string]string{"detail": "Invalid JSON"})
	}

	// Check if it's a batch request
	message = bytes.TrimSpace(message)
	if len(message) > 0 && message[0] == '[' {
		return h.handleBatch(ctx, message)
	}

	// Parse single request
	var req Request
	if err := json.Unmarshal(message, &req); err != nil {
		return h.errorResponse(nil, ParseErrorCode, "Parse error", 
			map[string]string{"detail": err.Error()})
	}

	// Handle notification (no ID field)
	if req.ID == nil {
		h.handleNotification(ctx, req)
		// Notifications don't get responses
		return nil, nil
	}

	// Process request with enhanced error handling
	return h.processRequest(ctx, req)
}

// OnError handles transport errors
func (h *ProtocolHandler) OnError(err error) {
	// Log error with context
	logger.Error("Protocol error: %v", err)
}

// handleNotification handles JSON-RPC 2.0 notifications (requests without ID)
func (h *ProtocolHandler) handleNotification(ctx context.Context, req Request) {
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

// processRequest handles a single request with full error handling
func (h *ProtocolHandler) processRequest(ctx context.Context, req Request) ([]byte, error) {
	// Validate JSON-RPC version
	if req.Jsonrpc != "2.0" {
		return h.errorResponse(req.ID, InvalidRequestCode, 
			"Invalid Request", map[string]string{
				"detail": "JSON-RPC version must be 2.0",
				"received": req.Jsonrpc,
			})
	}

	// Validate method
	if req.Method == "" {
		return h.errorResponse(req.ID, InvalidRequestCode,
			"Invalid Request", map[string]string{
				"detail": "Method is required",
			})
	}

	// Add request tracking
	if err := h.trackRequest(req.ID); err != nil {
		rateLimitErr := NewRateLimitError(h.server.config.RateLimit, time.Minute)
		return json.Marshal(ErrorResponse(req.ID, rateLimitErr))
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
		timeoutErr := NewTimeoutError(req.Method, h.server.config.RequestTimeout)
		return json.Marshal(ErrorResponse(req.ID, timeoutErr))
	}
}

// routeRequest routes the request to the appropriate handler
func (h *ProtocolHandler) routeRequest(ctx context.Context, req Request) ([]byte, error) {
	// Standard MCP methods
	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, req)
	case "initialized":
		// Client confirms initialization
		return h.successResponse(req.ID, map[string]bool{"success": true})
	case "shutdown":
		return h.handleShutdown(ctx, req)
	case "tools/list":
		return h.handleToolsList(ctx, req)
	case "tools/call":
		return h.handleToolCall(ctx, req)
	case "completion/complete":
		return h.handleCompletion(ctx, req)
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
		
		return h.errorResponse(req.ID, MethodNotFoundCode,
			fmt.Sprintf("Method not found: %s", req.Method),
			map[string]interface{}{
				"availableMethods": h.getAvailableMethods(),
			})
	}
}

// handleInitialize handles the MCP initialization handshake
func (h *ProtocolHandler) handleInitialize(ctx context.Context, req Request) ([]byte, error) {
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
			return h.errorResponse(req.ID, InvalidParamsCode,
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

// handleToolsList returns all available tools
func (h *ProtocolHandler) handleToolsList(ctx context.Context, req Request) ([]byte, error) {
	tools := h.server.tools.List()
	
	// Check if enhanced metadata is requested
	var params struct {
		IncludeMetadata bool `json:"includeMetadata"`
	}
	if req.Params != nil {
		json.Unmarshal(req.Params, &params)
	}
	
	// Convert to MCP format
	toolSchemas := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		toolSchema := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": map[string]interface{}{
				"type":       tool.Parameters.Type,
				"properties": tool.Parameters.Properties,
				"required":   tool.Parameters.Required,
			},
		}
		
		// Include metadata if requested and available
		if params.IncludeMetadata && tool.Metadata != nil {
			toolSchema["metadata"] = tool.Metadata
		}
		
		toolSchemas[i] = toolSchema
	}
	
	result := map[string]interface{}{
		"tools": toolSchemas,
	}
	
	return h.successResponse(req.ID, result)
}

// handleToolCall executes a tool with full error handling and validation
func (h *ProtocolHandler) handleToolCall(ctx context.Context, req Request) ([]byte, error) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		invalidErr := NewInvalidParamsError("Failed to parse tool parameters", "params")
		invalidErr.Details["parse_error"] = err.Error()
		invalidErr.Details["expected"] = map[string]string{
			"name":      "string",
			"arguments": "object",
			"stream":    "boolean (optional)",
		}
		return json.Marshal(ErrorResponse(req.ID, invalidErr))
	}

	// Validate required fields
	if params.Name == "" {
		return json.Marshal(ErrorResponse(req.ID, 
			NewInvalidParamsError("Tool name is required", "name")))
	}

	// Get tool from registry
	tool, exists := h.server.tools.Get(params.Name)
	if !exists {
		similar := h.findSimilarTool(params.Name)
		return json.Marshal(ErrorResponse(req.ID, 
			NewMethodNotFoundError(params.Name, similar)))
	}

	// Validate parameters against schema
	if err := h.validateToolParams(tool, params.Arguments); err != nil {
		validationErr := NewValidationError("arguments", err.Error())
		validationErr.Details["schema"] = tool.Parameters
		validationErr.Details["received"] = params.Arguments
		validationErr.ToolName = tool.Name
		return json.Marshal(ErrorResponse(req.ID, validationErr))
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

	// Track request
	h.requests.Store(req.ID, execCtx)
	defer h.requests.Delete(req.ID)

	// Handle streaming if requested and supported
	if tool.Streaming && params.Stream {
		return h.handleStreamingToolCall(ctx, execCtx, params)
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
		// Convert to MCPError if not already
		mcpErr := WrapError(err, ErrorTypeInternalError, tool.Name)
		mcpErr.Details["duration"] = duration.String()
		
		// Add specific error handling
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "not found"):
			mcpErr.Type = ErrorTypeDataNotFound
		case strings.Contains(errStr, "timeout"):
			mcpErr = NewTimeoutError(tool.Name, duration)
		case strings.Contains(errStr, "permission"):
			mcpErr.Type = ErrorTypePermissionError
		case strings.Contains(errStr, "query"):
			mcpErr.Type = ErrorTypeQueryError
		}
		
		return json.Marshal(ErrorResponse(req.ID, mcpErr))
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

// handleStreamingToolCall handles streaming tool execution
func (h *ProtocolHandler) handleStreamingToolCall(ctx context.Context, execCtx *ExecutionContext, params ToolCallParams) ([]byte, error) {
	// For HTTP/SSE transports, we'll return a streaming token
	// For stdio, we'll buffer and return the complete result
	
	if h.server.config.TransportType == TransportStdio {
		// Buffer streaming results for stdio
		stream := make(chan StreamChunk, 100)
		go execCtx.Tool.StreamHandler(ctx, params.Arguments, stream)
		
		var results []interface{}
		for chunk := range stream {
			if chunk.Error != nil {
				return h.errorResponse(execCtx.RequestID, InternalError, chunk.Error.Error(), nil)
			}
			if chunk.Type == "result" || chunk.Type == "complete" {
				results = append(results, chunk.Data)
			}
		}
		
		return h.successResponse(execCtx.RequestID, map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": formatToolResult(results),
				},
			},
		})
	}
	
	// For HTTP/SSE, return a streaming token
	streamID := h.generateStreamID()
	
	// Start streaming in background
	go h.handleStreamingExecution(ctx, streamID, execCtx, params)
	
	return h.successResponse(execCtx.RequestID, map[string]interface{}{
		"stream": true,
		"streamId": streamID,
		"message": "Streaming response initiated",
	})
}

// handleCompletion provides completion suggestions with context awareness
func (h *ProtocolHandler) handleCompletion(ctx context.Context, req Request) ([]byte, error) {
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
		return h.errorResponse(req.ID, InvalidParamsCode,
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

// handleSessionCreate creates a new session
func (h *ProtocolHandler) handleSessionCreate(ctx context.Context, req Request) ([]byte, error) {
	session := h.server.sessions.Create()
	
	result := map[string]interface{}{
		"sessionId": session.ID,
		"createdAt": session.CreatedAt,
	}
	
	return h.successResponse(req.ID, result)
}

// handleSessionGet retrieves session information
func (h *ProtocolHandler) handleSessionGet(ctx context.Context, req Request) ([]byte, error) {
	var params struct {
		SessionID string `json:"sessionId"`
	}
	
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.errorResponse(req.ID, InvalidParams, "Invalid parameters", err)
	}
	
	session, exists := h.server.sessions.Get(params.SessionID)
	if !exists {
		return h.errorResponse(req.ID, InvalidParams, "Session not found", nil)
	}
	
	result := map[string]interface{}{
		"session": session,
	}
	
	return h.successResponse(req.ID, result)
}

// Helper methods

func (h *ProtocolHandler) successResponse(id interface{}, result interface{}) ([]byte, error) {
	resp := Response{
		Jsonrpc: "2.0",
		Result:  result,
		ID:      id,
	}
	return json.Marshal(resp)
}

func (h *ProtocolHandler) errorResponse(id interface{}, code int, message string, data interface{}) ([]byte, error) {
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

func (h *ProtocolHandler) generateStreamID() string {
	return fmt.Sprintf("stream_%d_%d", time.Now().UnixNano(), atomic.AddInt64(&h.idCounter, 1))
}

func (h *ProtocolHandler) handleStreamingExecution(ctx context.Context, streamID string, execCtx *ExecutionContext, params ToolCallParams) {
	stream := make(chan StreamChunk, 100)
	
	// Execute streaming handler
	go execCtx.Tool.StreamHandler(ctx, params.Arguments, stream)
	
	// Process stream chunks
	for chunk := range stream {
		// In a real implementation, this would send to SSE manager
		// For now, we'll just log
		logger.Debug("Stream %s: %+v", streamID, chunk)
	}
}

func (h *ProtocolHandler) getToolCompletions(tool *Tool, argName, value string) []map[string]interface{} {
	completions := []map[string]interface{}{}
	
	// Get property definition
	prop, exists := tool.Parameters.Properties[argName]
	if !exists {
		return completions
	}
	
	// If enum is defined, return enum values
	if len(prop.Enum) > 0 {
		for _, enumVal := range prop.Enum {
			if value == "" || contains(enumVal, value) {
				completions = append(completions, map[string]interface{}{
					"value": enumVal,
					"label": enumVal,
				})
			}
		}
	}
	
	// Add type-specific completions
	switch prop.Type {
	case "boolean":
		completions = append(completions, 
			map[string]interface{}{"value": "true", "label": "true"},
			map[string]interface{}{"value": "false", "label": "false"},
		)
	}
	
	return completions
}

func (h *ProtocolHandler) getContextAwareCompletions(params interface{}) []map[string]interface{} {
	// TODO: Implement context-aware completions
	// For now, delegate to basic tool completions if applicable
	if p, ok := params.(struct {
		Ref struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"ref"`
		Argument struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"argument"`
		Context map[string]interface{} `json:"context"`
	}); ok && p.Ref.Type == "tool" {
		if tool, exists := h.server.tools.Get(p.Ref.Name); exists {
			return h.getToolCompletions(tool, p.Argument.Name, p.Argument.Value)
		}
	}
	return []map[string]interface{}{}
}

// Utility functions

func formatToolResult(result interface{}) string {
	// Convert result to readable text
	if str, ok := result.(string); ok {
		return str
	}
	
	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", result)
	}
	
	return string(bytes)
}

func contains(str, substr string) bool {
	return len(substr) > 0 && len(str) >= len(substr) && str[:len(substr)] == substr
}

// Additional helper methods for enhanced protocol

func (h *ProtocolHandler) trackRequest(id interface{}) error {
	// TODO: Implement rate limiting
	return nil
}

func (h *ProtocolHandler) untrackRequest(id interface{}) {
	// TODO: Clean up request tracking
}

func (h *ProtocolHandler) cancelRequest(id interface{}) {
	// TODO: Implement request cancellation
	if execCtx, ok := h.requests.Load(id); ok {
		// Cancel the context if possible
		logger.Debug("Cancelling request: %v", id)
		h.requests.Delete(id)
	}
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
	return h.errorResponse(req.ID, MethodNotFoundCode,
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
	return h.errorResponse(req.ID, MethodNotFoundCode,
		"Resources not yet implemented", nil)
}

func (h *ProtocolHandler) handleDirectToolCall(ctx context.Context, req Request, tool *Tool) ([]byte, error) {
	// Convert direct tool call to standard format
	var params map[string]interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.errorResponse(req.ID, InvalidParamsCode,
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
	
	return h.handleToolCall(ctx, toolCallReq)
}

// mustMarshal marshals data and panics on error (for internal use only)
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return data
}

// handleBatch processes batch requests with proper ordering
func (h *ProtocolHandler) handleBatch(ctx context.Context, message []byte) ([]byte, error) {
	var batchReq []Request
	if err := json.Unmarshal(message, &batchReq); err != nil {
		return h.errorResponse(nil, ParseErrorCode,
			"Parse error", map[string]string{
				"detail": "Invalid batch request format",
			})
	}

	if len(batchReq) == 0 {
		return h.errorResponse(nil, InvalidRequestCode,
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

	// Use a semaphore to limit concurrent executions
	maxConcurrent := h.server.config.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 10 // Default
	}
	sem := make(chan struct{}, maxConcurrent)

	for i, req := range batchReq {
		if req.ID == nil {
			// Handle notifications immediately (no response)
			go h.handleNotification(ctx, req)
			continue
		}

		wg.Add(1)
		go func(idx int, r Request) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()
			
			data, _ := h.processRequest(ctx, r)
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
package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// MCP Protocol Errors
	ErrorTypeParseError     ErrorType = "parse_error"
	ErrorTypeInvalidRequest ErrorType = "invalid_request"
	ErrorTypeMethodNotFound ErrorType = "method_not_found"
	ErrorTypeInvalidParams  ErrorType = "invalid_params"
	ErrorTypeInternalError  ErrorType = "internal_error"
	
	// Tool Execution Errors
	ErrorTypeTimeout        ErrorType = "timeout"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeUnauthorized   ErrorType = "unauthorized"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeConflict       ErrorType = "conflict"
	
	// New Relic API Errors
	ErrorTypeAPIError       ErrorType = "api_error"
	ErrorTypeQueryError     ErrorType = "query_error"
	ErrorTypeAccountError   ErrorType = "account_error"
	ErrorTypePermissionError ErrorType = "permission_error"
	
	// Data Errors
	ErrorTypeValidation     ErrorType = "validation_error"
	ErrorTypeDataNotFound   ErrorType = "data_not_found"
	ErrorTypeDataQuality    ErrorType = "data_quality"
)

// MCPError represents a structured error for the MCP protocol
type MCPError struct {
	Type       ErrorType              `json:"type"`
	Code       int                    `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Hint       string                 `json:"hint,omitempty"`
	ToolName   string                 `json:"tool_name,omitempty"`
	RequestID  interface{}            `json:"request_id,omitempty"`
}

// Error implements the error interface
func (e *MCPError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s (hint: %s)", e.Type, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// ToJSONRPCError converts to JSON-RPC error format
func (e *MCPError) ToJSONRPCError() *Error {
	data := map[string]interface{}{
		"type": e.Type,
	}
	
	if e.Details != nil {
		data["details"] = e.Details
	}
	if e.Hint != "" {
		data["hint"] = e.Hint
	}
	if e.ToolName != "" {
		data["tool"] = e.ToolName
	}
	
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Data:    data,
	}
}

// MarshalJSON implements custom JSON marshaling
func (e *MCPError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.ToJSONRPCError())
}

// Error constructors

// NewParseError creates a parse error
func NewParseError(message string) *MCPError {
	return &MCPError{
		Type:    ErrorTypeParseError,
		Code:    ParseErrorCode,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewInvalidRequestError creates an invalid request error
func NewInvalidRequestError(message string, details map[string]interface{}) *MCPError {
	return &MCPError{
		Type:    ErrorTypeInvalidRequest,
		Code:    InvalidRequestCode,
		Message: message,
		Details: details,
	}
}

// NewMethodNotFoundError creates a method not found error
func NewMethodNotFoundError(method string, similar string) *MCPError {
	err := &MCPError{
		Type:    ErrorTypeMethodNotFound,
		Code:    MethodNotFoundCode,
		Message: fmt.Sprintf("Method '%s' not found", method),
		Details: map[string]interface{}{
			"method": method,
		},
	}
	
	if similar != "" {
		err.Hint = fmt.Sprintf("Did you mean '%s'?", similar)
		err.Details["suggestion"] = similar
	}
	
	return err
}

// NewInvalidParamsError creates an invalid parameters error
func NewInvalidParamsError(message string, param string) *MCPError {
	return &MCPError{
		Type:    ErrorTypeInvalidParams,
		Code:    InvalidParamsCode,
		Message: message,
		Details: map[string]interface{}{
			"parameter": param,
		},
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string, err error) *MCPError {
	mcpErr := &MCPError{
		Type:    ErrorTypeInternalError,
		Code:    InternalErrorCode,
		Message: message,
	}
	
	if err != nil {
		mcpErr.Details = map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	return mcpErr
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(toolName string, timeout time.Duration) *MCPError {
	return &MCPError{
		Type:     ErrorTypeTimeout,
		Code:     InternalErrorCode,
		Message:  fmt.Sprintf("Tool execution timed out after %v", timeout),
		ToolName: toolName,
		Hint:     "Try reducing the query complexity or time range",
		Details:  make(map[string]interface{}),
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(limit int, window time.Duration) *MCPError {
	return &MCPError{
		Type:    ErrorTypeRateLimit,
		Code:    -32001, // Custom error code
		Message: "Rate limit exceeded",
		Details: map[string]interface{}{
			"limit":      limit,
			"window":     window.String(),
			"retry_after": window.Seconds(),
		},
		Hint: fmt.Sprintf("Maximum %d requests per %v", limit, window),
	}
}

// NewValidationError creates a validation error
func NewValidationError(field string, message string) *MCPError {
	return &MCPError{
		Type:    ErrorTypeValidation,
		Code:    InvalidParamsCode,
		Message: fmt.Sprintf("Validation failed for field '%s': %s", field, message),
		Details: map[string]interface{}{
			"field":   field,
			"message": message,
		},
	}
}

// NewAPIError creates a New Relic API error
func NewAPIError(message string, statusCode int, response string) *MCPError {
	return &MCPError{
		Type:    ErrorTypeAPIError,
		Code:    InternalErrorCode,
		Message: fmt.Sprintf("New Relic API error: %s", message),
		Details: map[string]interface{}{
			"status_code": statusCode,
			"response":    response,
		},
		Hint: "Check your API credentials and permissions",
	}
}

// NewQueryError creates a query execution error
func NewQueryError(query string, err error) *MCPError {
	mcpErr := &MCPError{
		Type:    ErrorTypeQueryError,
		Code:    InternalErrorCode,
		Message: "Query execution failed",
		Details: map[string]interface{}{
			"query": query,
		},
	}
	
	if err != nil {
		mcpErr.Details["error"] = err.Error()
		
		// Add specific hints based on error message
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "timeout"):
			mcpErr.Hint = "Try reducing the time range or query complexity"
		case strings.Contains(errStr, "syntax"):
			mcpErr.Hint = "Check NRQL syntax: https://docs.newrelic.com/docs/query-your-data/nrql-reference/"
		case strings.Contains(errStr, "permission"):
			mcpErr.Hint = "Ensure your API key has query permissions"
		}
	}
	
	return mcpErr
}

// NewAccountError creates an account access error
func NewAccountError(accountID string, action string) *MCPError {
	return &MCPError{
		Type:    ErrorTypeAccountError,
		Code:    -32002, // Custom error code
		Message: fmt.Sprintf("Cannot %s for account %s", action, accountID),
		Details: map[string]interface{}{
			"account_id": accountID,
			"action":     action,
		},
		Hint: "Verify the account ID exists and you have access",
	}
}

// NewDataNotFoundError creates a data not found error
func NewDataNotFoundError(resource string, identifier string) *MCPError {
	return &MCPError{
		Type:    ErrorTypeDataNotFound,
		Code:    -32003, // Custom error code
		Message: fmt.Sprintf("%s not found: %s", resource, identifier),
		Details: map[string]interface{}{
			"resource":   resource,
			"identifier": identifier,
		},
	}
}

// Error helpers

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	if mcpErr, ok := err.(*MCPError); ok {
		switch mcpErr.Type {
		case ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeAPIError:
			return true
		}
	}
	return false
}

// GetRetryDelay returns the retry delay for an error
func GetRetryDelay(err error) time.Duration {
	if mcpErr, ok := err.(*MCPError); ok {
		if mcpErr.Type == ErrorTypeRateLimit {
			if retryAfter, ok := mcpErr.Details["retry_after"].(float64); ok {
				return time.Duration(retryAfter) * time.Second
			}
		}
	}
	return 5 * time.Second // Default retry delay
}

// WrapError wraps a standard error into MCPError
func WrapError(err error, errorType ErrorType, toolName string) *MCPError {
	if err == nil {
		return nil
	}
	
	// If already an MCPError, return as-is
	if mcpErr, ok := err.(*MCPError); ok {
		if toolName != "" && mcpErr.ToolName == "" {
			mcpErr.ToolName = toolName
		}
		return mcpErr
	}
	
	// Create new MCPError
	return &MCPError{
		Type:     errorType,
		Code:     InternalErrorCode,
		Message:  err.Error(),
		ToolName: toolName,
		Details:  make(map[string]interface{}),
	}
}

// ErrorResponse creates a complete error response
func ErrorResponse(requestID interface{}, err error) Response {
	var jsonRPCError *Error
	
	if mcpErr, ok := err.(*MCPError); ok {
		mcpErr.RequestID = requestID
		jsonRPCError = mcpErr.ToJSONRPCError()
	} else {
		// Wrap standard errors
		jsonRPCError = &Error{
			Code:    InternalErrorCode,
			Message: err.Error(),
		}
	}
	
	return Response{
		Jsonrpc: "2.0",
		Error:   jsonRPCError,
		ID:      requestID,
	}
}
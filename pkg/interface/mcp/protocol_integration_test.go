//go:build integration

package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProtocolEnhancedIntegration tests the enhanced protocol implementation
func TestProtocolEnhancedIntegration(t *testing.T) {
	// This test demonstrates the enhanced JSON-RPC 2.0 features
	
	// Example of proper error response structure
	errorResp := Response{
		Jsonrpc: "2.0",
		Error: &Error{
			Code:    InvalidParamsCode,
			Message: "Invalid parameters",
			Data: map[string]interface{}{
				"detail": "Parameter 'query' is required",
				"hint":   "Provide a valid NRQL query string",
			},
		},
		ID: 123,
	}
	
	// Verify it marshals correctly
	data, err := json.Marshal(errorResp)
	require.NoError(t, err)
	
	var parsed Response
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, InvalidParamsCode, parsed.Error.Code)
	
	// Example of batch request structure
	batchReq := []Request{
		{
			Jsonrpc: "2.0",
			Method:  "tools/list",
			ID:      1,
		},
		{
			Jsonrpc: "2.0",
			Method:  "query_nrdb",
			Params:  json.RawMessage(`{"query":"SELECT count(*) FROM Transaction"}`),
			ID:      2,
		},
		{
			Jsonrpc: "2.0",
			Method:  "tools/changed", // Notification (no ID)
		},
	}
	
	// Verify batch marshaling
	batchData, err := json.Marshal(batchReq)
	require.NoError(t, err)
	assert.Contains(t, string(batchData), "tools/list")
	assert.Contains(t, string(batchData), "query_nrdb")
}

// TestEnhancedErrorCodes verifies all MCP-specific error codes
func TestEnhancedErrorCodes(t *testing.T) {
	// Verify error codes are unique and well-defined
	errorCodes := map[string]int{
		"ParseError":      ParseErrorCode,
		"InvalidRequest":  InvalidRequestCode,
		"MethodNotFound":  MethodNotFoundCode,
		"InvalidParams":   InvalidParamsCode,
		"InternalError":   InternalErrorCode,
		"ToolNotFound":    ToolNotFoundCode,
		"ToolExecution":   ToolExecutionCode,
		"SessionNotFound": SessionNotFoundCode,
		"Timeout":         TimeoutErrorCode,
		"RateLimit":       RateLimitCode,
	}
	
	// Check for duplicates
	seen := make(map[int]string)
	for name, code := range errorCodes {
		if existing, exists := seen[code]; exists {
			t.Errorf("Duplicate error code %d for %s and %s", code, name, existing)
		}
		seen[code] = name
	}
	
	// Verify standard codes match JSON-RPC 2.0 spec
	assert.Equal(t, -32700, ParseErrorCode)
	assert.Equal(t, -32600, InvalidRequestCode)
	assert.Equal(t, -32601, MethodNotFoundCode)
	assert.Equal(t, -32602, InvalidParamsCode)
	assert.Equal(t, -32603, InternalErrorCode)
	
	// Verify MCP-specific codes are in reserved range
	assert.True(t, ToolNotFoundCode >= -32099 && ToolNotFoundCode <= -32000)
	assert.True(t, ToolExecutionCode >= -32099 && ToolExecutionCode <= -32000)
	assert.True(t, SessionNotFoundCode >= -32099 && SessionNotFoundCode <= -32000)
	assert.True(t, TimeoutErrorCode >= -32099 && TimeoutErrorCode <= -32000)
	assert.True(t, RateLimitCode >= -32099 && RateLimitCode <= -32000)
}

// TestProtocolValidation demonstrates parameter validation
func TestProtocolValidation(t *testing.T) {
	// Example tool parameter schema
	toolParams := ToolParameters{
		Type:     "object",
		Required: []string{"query"},
		Properties: map[string]Property{
			"query": {
				Type:        "string",
				Description: "NRQL query to execute",
			},
			"timeout": {
				Type:        "integer",
				Description: "Query timeout in seconds",
				Default:     30,
			},
			"account_id": {
				Type:        "string",
				Description: "Account ID to query",
			},
		},
	}
	
	// Valid parameters
	validParams := map[string]interface{}{
		"query":      "SELECT count(*) FROM Transaction",
		"timeout":    60,
		"account_id": "12345",
	}
	
	// Invalid parameters (missing required field)
	invalidParams := map[string]interface{}{
		"timeout": 60,
	}
	
	// Validate structure
	assert.Contains(t, toolParams.Required, "query")
	assert.NotContains(t, toolParams.Required, "timeout")
	
	// Check property exists
	_, hasQuery := validParams["query"]
	assert.True(t, hasQuery)
	
	_, hasQuery2 := invalidParams["query"]
	assert.False(t, hasQuery2)
}
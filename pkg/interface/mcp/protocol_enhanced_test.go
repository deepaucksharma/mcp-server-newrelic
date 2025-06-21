package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedProtocol_JSONRPCCompliance(t *testing.T) {
	// Create server with standard configuration
	config := ServerConfig{
		MaxConcurrent:    10,
		RequestTimeout:   30 * time.Second,
		StreamingEnabled: true,
	}
	server := NewServer(config)
	handler := &ProtocolHandler{server: server}

	tests := []struct {
		name     string
		request  string
		validate func(t *testing.T, response []byte, err error)
	}{
		{
			name:    "valid request",
			request: `{"jsonrpc":"2.0","method":"tools/list","id":1}`,
			validate: func(t *testing.T, response []byte, err error) {
				require.NoError(t, err)
				var resp Response
				require.NoError(t, json.Unmarshal(response, &resp))
				assert.Equal(t, "2.0", resp.Jsonrpc)
				assert.Equal(t, float64(1), resp.ID)
				assert.Nil(t, resp.Error)
			},
		},
		{
			name:    "missing jsonrpc version",
			request: `{"method":"tools/list","id":2}`,
			validate: func(t *testing.T, response []byte, err error) {
				require.NoError(t, err)
				var resp Response
				require.NoError(t, json.Unmarshal(response, &resp))
				assert.NotNil(t, resp.Error)
				assert.Equal(t, InvalidRequestCode, resp.Error.Code)
			},
		},
		{
			name:    "invalid JSON",
			request: `{invalid json}`,
			validate: func(t *testing.T, response []byte, err error) {
				require.NoError(t, err)
				var resp Response
				require.NoError(t, json.Unmarshal(response, &resp))
				assert.NotNil(t, resp.Error)
				assert.Equal(t, ParseErrorCode, resp.Error.Code)
			},
		},
		{
			name:    "notification (no ID)",
			request: `{"jsonrpc":"2.0","method":"tools/changed"}`,
			validate: func(t *testing.T, response []byte, err error) {
				require.NoError(t, err)
				assert.Nil(t, response) // Notifications don't get responses
			},
		},
		{
			name:    "method not found",
			request: `{"jsonrpc":"2.0","method":"invalid/method","id":3}`,
			validate: func(t *testing.T, response []byte, err error) {
				require.NoError(t, err)
				var resp Response
				require.NoError(t, json.Unmarshal(response, &resp))
				assert.NotNil(t, resp.Error)
				assert.Equal(t, MethodNotFoundCode, resp.Error.Code)
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := handler.HandleMessage(ctx, []byte(tt.request))
			tt.validate(t, response, err)
		})
	}
}

func TestEnhancedProtocol_BatchRequests(t *testing.T) {
	config := ServerConfig{
		MaxConcurrent:    10,
		RequestTimeout:   30 * time.Second,
		StreamingEnabled: true,
	}
	server := NewServer(config)
	handler := &ProtocolHandler{server: server}

	// Test batch request
	batchRequest := `[
		{"jsonrpc":"2.0","method":"tools/list","id":1},
		{"jsonrpc":"2.0","method":"tools/list","id":2},
		{"jsonrpc":"2.0","method":"tools/changed"}
	]`

	ctx := context.Background()
	response, err := handler.HandleMessage(ctx, []byte(batchRequest))
	require.NoError(t, err)

	// Parse batch response
	var batchResp []json.RawMessage
	require.NoError(t, json.Unmarshal(response, &batchResp))
	
	// Should have 2 responses (notifications don't get responses)
	assert.Len(t, batchResp, 2)

	// Verify each response
	for i, rawResp := range batchResp {
		var resp Response
		require.NoError(t, json.Unmarshal(rawResp, &resp))
		assert.Equal(t, "2.0", resp.Jsonrpc)
		assert.Equal(t, float64(i+1), resp.ID)
		assert.Nil(t, resp.Error)
	}
}

func TestEnhancedProtocol_ParameterValidation(t *testing.T) {
	config := ServerConfig{
		StreamingEnabled: true,
	}
	server := NewServer(config)
	
	// Register a test tool
	err := server.tools.Register(Tool{
		Name:        "test_tool",
		Description: "Test tool for validation",
		Parameters: ToolParameters{
			Type:     "object",
			Required: []string{"required_param"},
			Properties: map[string]Property{
				"required_param": {
					Type:        "string",
					Description: "A required parameter",
				},
				"optional_param": {
					Type:        "integer",
					Description: "An optional parameter",
				},
				"enum_param": {
					Type:        "string",
					Description: "Parameter with enum values",
					Enum:        []string{"option1", "option2", "option3"},
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]string{"result": "success"}, nil
		},
	})
	require.NoError(t, err)

	handler := &ProtocolHandler{server: server}

	tests := []struct {
		name        string
		arguments   map[string]interface{}
		expectError bool
		errorCode   int
	}{
		{
			name: "valid parameters",
			arguments: map[string]interface{}{
				"required_param": "value",
				"optional_param": 42,
			},
			expectError: false,
		},
		{
			name:        "missing required parameter",
			arguments:   map[string]interface{}{},
			expectError: true,
			errorCode:   InvalidParamsCode,
		},
		{
			name: "invalid type",
			arguments: map[string]interface{}{
				"required_param": 123, // Should be string
			},
			expectError: true,
			errorCode:   InvalidParamsCode,
		},
		{
			name: "invalid enum value",
			arguments: map[string]interface{}{
				"required_param": "value",
				"enum_param":     "invalid_option",
			},
			expectError: true,
			errorCode:   InvalidParamsCode,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{
				Jsonrpc: "2.0",
				Method:  "tools/call",
				Params: mustMarshal(ToolCallParams{
					Name:      "test_tool",
					Arguments: tt.arguments,
				}),
				ID: 1,
			}

			reqBytes, _ := json.Marshal(req)
			response, err := handler.HandleMessage(ctx, reqBytes)
			require.NoError(t, err)

			var resp Response
			require.NoError(t, json.Unmarshal(response, &resp))

			if tt.expectError {
				assert.NotNil(t, resp.Error)
				assert.Equal(t, tt.errorCode, resp.Error.Code)
			} else {
				assert.Nil(t, resp.Error)
				assert.NotNil(t, resp.Result)
			}
		})
	}
}
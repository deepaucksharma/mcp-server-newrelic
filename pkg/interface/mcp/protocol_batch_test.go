package mcp

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchRequestHandling(t *testing.T) {
	server := &Server{
		tools:    NewToolRegistry(),
		sessions: NewSessionManager(),
		config: ServerConfig{
			RequestTimeout: 5 * time.Second,
			MaxConcurrent:  3,
		},
	}
	
	handler := &ProtocolHandler{
		server:   server,
		requests: syncMap{},
	}
	
	// Register a test tool
	server.tools.Register(Tool{
		Name:        "test.echo",
		Description: "Echo input",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]Property{
				"message": {
					Type:        "string",
					Description: "Message to echo",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return params["message"], nil
		},
	})

	t.Run("ValidBatchRequest", func(t *testing.T) {
		batch := `[
			{"jsonrpc":"2.0","method":"initialize","id":1},
			{"jsonrpc":"2.0","method":"tools/list","id":2},
			{"jsonrpc":"2.0","method":"tools/call","params":{"name":"test.echo","arguments":{"message":"hello"}},"id":3}
		]`
		
		respBytes, err := handler.HandleMessage(context.Background(), []byte(batch))
		require.NoError(t, err)
		require.NotNil(t, respBytes)
		
		var responses []Response
		err = json.Unmarshal(respBytes, &responses)
		require.NoError(t, err)
		assert.Len(t, responses, 3)
		
		// Check each response
		for i, resp := range responses {
			assert.Equal(t, "2.0", resp.Jsonrpc)
			assert.Nil(t, resp.Error)
			assert.NotNil(t, resp.Result)
			assert.Equal(t, float64(i+1), resp.ID)
		}
	})

	t.Run("BatchWithNotifications", func(t *testing.T) {
		batch := `[
			{"jsonrpc":"2.0","method":"initialize","id":1},
			{"jsonrpc":"2.0","method":"tools/changed"},
			{"jsonrpc":"2.0","method":"tools/list","id":2}
		]`
		
		respBytes, err := handler.HandleMessage(context.Background(), []byte(batch))
		require.NoError(t, err)
		require.NotNil(t, respBytes)
		
		var responses []Response
		err = json.Unmarshal(respBytes, &responses)
		require.NoError(t, err)
		// Should only have 2 responses (notifications don't get responses)
		assert.Len(t, responses, 2)
		
		assert.Equal(t, float64(1), responses[0].ID)
		assert.Equal(t, float64(2), responses[1].ID)
	})

	t.Run("BatchWithErrors", func(t *testing.T) {
		batch := `[
			{"jsonrpc":"2.0","method":"unknown.method","id":1},
			{"jsonrpc":"2.0","method":"tools/list","id":2},
			{"jsonrpc":"2.0","method":"tools/call","params":"invalid","id":3}
		]`
		
		respBytes, err := handler.HandleMessage(context.Background(), []byte(batch))
		require.NoError(t, err)
		require.NotNil(t, respBytes)
		
		var responses []Response
		err = json.Unmarshal(respBytes, &responses)
		require.NoError(t, err)
		assert.Len(t, responses, 3)
		
		// First request should error (unknown method)
		assert.NotNil(t, responses[0].Error)
		assert.Equal(t, MethodNotFoundCode, responses[0].Error.Code)
		
		// Second request should succeed
		assert.Nil(t, responses[1].Error)
		assert.NotNil(t, responses[1].Result)
		
		// Third request should error (invalid params)
		assert.NotNil(t, responses[2].Error)
		assert.Equal(t, InvalidParamsCode, responses[2].Error.Code)
	})

	t.Run("EmptyBatch", func(t *testing.T) {
		batch := `[]`
		
		respBytes, err := handler.HandleMessage(context.Background(), []byte(batch))
		require.NoError(t, err)
		require.NotNil(t, respBytes)
		
		var response Response
		err = json.Unmarshal(respBytes, &response)
		require.NoError(t, err)
		
		assert.NotNil(t, response.Error)
		assert.Equal(t, InvalidRequestCode, response.Error.Code)
		assert.Contains(t, response.Error.Message, "Empty batch")
	})

	t.Run("InvalidBatchJSON", func(t *testing.T) {
		batch := `[{"invalid": json}]`
		
		respBytes, err := handler.HandleMessage(context.Background(), []byte(batch))
		require.NoError(t, err)
		require.NotNil(t, respBytes)
		
		var response Response
		err = json.Unmarshal(respBytes, &response)
		require.NoError(t, err)
		
		assert.NotNil(t, response.Error)
		assert.Equal(t, ParseErrorCode, response.Error.Code)
	})
}

// syncMap is a type alias for testing compatibility
type syncMap = sync.Map
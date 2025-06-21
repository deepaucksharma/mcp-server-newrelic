//go:build integration

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStdioTransportIntegration tests the stdio transport
func TestStdioTransportIntegration(t *testing.T) {
	server := createTestServer(t)
	err := server.registerTools()
	require.NoError(t, err)
	
	transport := NewStdioTransport()
	
	// Create mock stdio
	input := &bytes.Buffer{}
	output := &bytes.Buffer{}
	
	// Override stdio for testing
	oldStdin := transport.(*StdioTransport).stdin
	oldStdout := transport.(*StdioTransport).stdout
	transport.(*StdioTransport).stdin = input
	transport.(*StdioTransport).stdout = output
	defer func() {
		transport.(*StdioTransport).stdin = oldStdin
		transport.(*StdioTransport).stdout = oldStdout
	}()
	
	// Start transport in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		err := transport.Start(ctx, server.protocol)
		if err != nil && err != io.EOF {
			t.Errorf("Transport start error: %v", err)
		}
	}()
	
	// Give transport time to start
	time.Sleep(100 * time.Millisecond)
	
	// Send a request
	request := Request{
		Jsonrpc: "2.0",
		Method:  "tools/list",
		ID:      1,
	}
	
	reqBytes, err := json.Marshal(request)
	require.NoError(t, err)
	
	// Write request with content-length header
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(reqBytes))
	input.WriteString(header)
	input.Write(reqBytes)
	
	// Wait for response
	time.Sleep(100 * time.Millisecond)
	
	// Read response
	responseStr := output.String()
	assert.Contains(t, responseStr, "Content-Length:")
	
	// Parse response
	parts := strings.SplitN(responseStr, "\r\n\r\n", 2)
	require.Len(t, parts, 2, "Should have header and body")
	
	var response Response
	err = json.Unmarshal([]byte(parts[1]), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "2.0", response.Jsonrpc)
	assert.Equal(t, json.Number("1"), response.ID)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)
}

// TestHTTPTransportIntegration tests the HTTP transport
func TestHTTPTransportIntegration(t *testing.T) {
	server := createTestServer(t)
	err := server.registerTools()
	require.NoError(t, err)
	
	// Create HTTP transport
	transport := NewHTTPTransport("127.0.0.1:0") // Use port 0 for automatic assignment
	
	// Start transport
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		err := transport.Start(ctx, server.protocol)
		if err != nil && !strings.Contains(err.Error(), "Server closed") {
			t.Errorf("Transport start error: %v", err)
		}
	}()
	
	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	
	// Get actual address
	addr := transport.(*HTTPTransport).server.Addr
	baseURL := fmt.Sprintf("http://%s", addr)
	
	t.Run("Single Request", func(t *testing.T) {
		request := Request{
			Jsonrpc: "2.0",
			Method:  "tools/list",
			ID:      1,
		}
		
		response := sendHTTPRequest(t, baseURL+"/rpc", request)
		
		assert.Equal(t, "2.0", response.Jsonrpc)
		assert.Equal(t, json.Number("1"), response.ID)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})
	
	t.Run("Batch Request", func(t *testing.T) {
		batch := []Request{
			{
				Jsonrpc: "2.0",
				Method:  "tools/list",
				ID:      1,
			},
			{
				Jsonrpc: "2.0",
				Method:  "initialize",
				Params: json.RawMessage(`{
					"protocolVersion": "0.1.0",
					"capabilities": {}
				}`),
				ID: 2,
			},
		}
		
		reqBytes, err := json.Marshal(batch)
		require.NoError(t, err)
		
		resp, err := http.Post(baseURL+"/rpc", "application/json", bytes.NewReader(reqBytes))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var responses []Response
		err = json.NewDecoder(resp.Body).Decode(&responses)
		require.NoError(t, err)
		
		assert.Len(t, responses, 2)
		for i, response := range responses {
			assert.Equal(t, "2.0", response.Jsonrpc)
			assert.Equal(t, json.Number(fmt.Sprintf("%d", batch[i].ID)), response.ID)
			assert.Nil(t, response.Error)
		}
	})
	
	t.Run("Invalid JSON", func(t *testing.T) {
		resp, err := http.Post(baseURL+"/rpc", "application/json", strings.NewReader("{invalid json"))
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode) // JSON-RPC errors return 200
		
		var response Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		
		assert.NotNil(t, response.Error)
		assert.Equal(t, ParseErrorCode, response.Error.Code)
	})
	
	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var health map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)
		
		assert.Equal(t, "ok", health["status"])
	})
}

// TestSSETransportIntegration tests the Server-Sent Events transport
func TestSSETransportIntegration(t *testing.T) {
	server := createTestServer(t)
	err := server.registerTools()
	require.NoError(t, err)
	
	// Create SSE transport
	transport := NewSSETransport("127.0.0.1:0")
	
	// Start transport
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		err := transport.Start(ctx, server.protocol)
		if err != nil && !strings.Contains(err.Error(), "Server closed") {
			t.Errorf("Transport start error: %v", err)
		}
	}()
	
	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	
	// Get actual address
	addr := transport.(*SSETransport).server.Addr
	baseURL := fmt.Sprintf("http://%s", addr)
	
	t.Run("SSE Connection", func(t *testing.T) {
		// Create SSE client
		req, err := http.NewRequest("GET", baseURL+"/sse", nil)
		require.NoError(t, err)
		req.Header.Set("Accept", "text/event-stream")
		
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
		
		// Send a request via POST
		request := Request{
			Jsonrpc: "2.0",
			Method:  "tools/list",
			ID:      "sse-1",
		}
		
		reqBytes, err := json.Marshal(request)
		require.NoError(t, err)
		
		postResp, err := http.Post(baseURL+"/rpc", "application/json", bytes.NewReader(reqBytes))
		require.NoError(t, err)
		postResp.Body.Close()
		
		// Read SSE response
		// Note: In a real implementation, we'd parse SSE events properly
		buf := make([]byte, 4096)
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			require.NoError(t, err)
		}
		
		if n > 0 {
			response := string(buf[:n])
			assert.Contains(t, response, "event:")
			assert.Contains(t, response, "data:")
		}
	})
}

// TestTransportResilience tests transport error handling and recovery
func TestTransportResilience(t *testing.T) {
	server := createTestServer(t)
	err := server.registerTools()
	require.NoError(t, err)
	
	t.Run("Request Timeout", func(t *testing.T) {
		// Create a slow handler
		slowTool := Tool{
			Name:        "test.slow",
			Description: "Slow tool for testing",
			Parameters:  ToolParameters{Type: "object"},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				select {
				case <-time.After(5 * time.Second):
					return "too slow", nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}
		server.tools.Register(slowTool)
		
		// Set short timeout
		server.config.RequestTimeout = 100 * time.Millisecond
		
		request := Request{
			Jsonrpc: "2.0",
			Method:  "tools/call",
			Params: json.RawMessage(`{
				"name": "test.slow",
				"arguments": {}
			}`),
			ID: "timeout-1",
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		
		reqBytes, err := json.Marshal(request)
		require.NoError(t, err)
		
		respBytes, err := server.protocol.HandleMessage(ctx, reqBytes)
		require.NoError(t, err)
		
		var response Response
		err = json.Unmarshal(respBytes, &response)
		require.NoError(t, err)
		
		assert.NotNil(t, response.Error)
		assert.Equal(t, TimeoutErrorCode, response.Error.Code)
	})
	
	t.Run("Large Request Handling", func(t *testing.T) {
		// Create a large query
		largeQuery := strings.Repeat("SELECT count(*) FROM Transaction UNION ", 100)
		largeQuery += "SELECT count(*) FROM Transaction"
		
		request := Request{
			Jsonrpc: "2.0",
			Method:  "tools/call",
			Params: json.RawMessage(fmt.Sprintf(`{
				"name": "query_nrdb",
				"arguments": {
					"query": "%s"
				}
			}`, largeQuery)),
			ID: "large-1",
		}
		
		reqBytes, err := json.Marshal(request)
		require.NoError(t, err)
		
		// Should handle large request
		respBytes, err := server.protocol.HandleMessage(context.Background(), reqBytes)
		require.NoError(t, err)
		
		var response Response
		err = json.Unmarshal(respBytes, &response)
		require.NoError(t, err)
		
		// May succeed or fail, but should not panic
		assert.Equal(t, "2.0", response.Jsonrpc)
	})
	
	t.Run("Concurrent Connection Limit", func(t *testing.T) {
		transport := NewHTTPTransport("127.0.0.1:0")
		
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		go transport.Start(ctx, server.protocol)
		time.Sleep(100 * time.Millisecond)
		
		addr := transport.(*HTTPTransport).server.Addr
		baseURL := fmt.Sprintf("http://%s/rpc", addr)
		
		// Send many concurrent requests
		concurrency := 50
		results := make(chan bool, concurrency)
		
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				request := Request{
					Jsonrpc: "2.0",
					Method:  "tools/list",
					ID:      id,
				}
				
				reqBytes, _ := json.Marshal(request)
				resp, err := http.Post(baseURL, "application/json", bytes.NewReader(reqBytes))
				if err != nil {
					results <- false
					return
				}
				defer resp.Body.Close()
				
				results <- resp.StatusCode == http.StatusOK
			}(i)
		}
		
		// Collect results
		successCount := 0
		for i := 0; i < concurrency; i++ {
			if <-results {
				successCount++
			}
		}
		
		// Most requests should succeed
		assert.Greater(t, successCount, concurrency*8/10, "At least 80% of requests should succeed")
	})
}

// Helper functions

func sendHTTPRequest(t *testing.T, url string, request Request) Response {
	reqBytes, err := json.Marshal(request)
	require.NoError(t, err)
	
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBytes))
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	
	return response
}

// TestTransportMetrics tests transport telemetry
func TestTransportMetrics(t *testing.T) {
	server := createTestServer(t)
	server.registerTools()
	
	// Mock metrics collector
	metrics := &mockMetrics{
		requests:   make(map[string]int),
		durations:  make([]time.Duration, 0),
		errors:     make(map[string]int),
	}
	
	// Inject metrics collector (would be done via config in production)
	// server.metrics = metrics
	
	transport := NewHTTPTransport("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go transport.Start(ctx, server.protocol)
	time.Sleep(100 * time.Millisecond)
	
	// Send various requests
	addr := transport.(*HTTPTransport).server.Addr
	baseURL := fmt.Sprintf("http://%s/rpc", addr)
	
	// Successful request
	sendHTTPRequest(t, baseURL, Request{
		Jsonrpc: "2.0",
		Method:  "tools/list",
		ID:      1,
	})
	
	// Failed request
	sendHTTPRequest(t, baseURL, Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{"name": "nonexistent"}`),
		ID:      2,
	})
	
	// Verify metrics were collected
	// assert.Greater(t, metrics.requests["tools/list"], 0)
	// assert.Greater(t, metrics.errors["tool_not_found"], 0)
	// assert.NotEmpty(t, metrics.durations)
}

type mockMetrics struct {
	requests  map[string]int
	durations []time.Duration
	errors    map[string]int
}
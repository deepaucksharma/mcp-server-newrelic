//go:build integration

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveMCPProtocol tests the full MCP protocol implementation
func TestComprehensiveMCPProtocol(t *testing.T) {
	// Create a fully configured server
	server := createTestServer(t)
	
	// Register all tools
	err := server.registerTools()
	require.NoError(t, err, "Failed to register tools")
	
	handler := &ProtocolHandler{
		server:   server,
		requests: sync.Map{},
	}
	
	t.Run("Initialize Session", func(t *testing.T) {
		testInitializeSession(t, handler)
	})
	
	t.Run("List Tools", func(t *testing.T) {
		testListTools(t, handler)
	})
	
	t.Run("Discovery Flow", func(t *testing.T) {
		testDiscoveryFlow(t, handler)
	})
	
	t.Run("Query Execution", func(t *testing.T) {
		testQueryExecution(t, handler)
	})
	
	t.Run("Batch Requests", func(t *testing.T) {
		testBatchRequests(t, handler)
	})
	
	t.Run("Error Handling", func(t *testing.T) {
		testErrorHandling(t, handler)
	})
	
	t.Run("Pagination", func(t *testing.T) {
		testPagination(t, handler)
	})
	
	t.Run("Multi-Account Support", func(t *testing.T) {
		testMultiAccountSupport(t, handler)
	})
	
	t.Run("Workflow Orchestration", func(t *testing.T) {
		testWorkflowOrchestration(t, handler)
	})
	
	t.Run("Concurrent Requests", func(t *testing.T) {
		testConcurrentRequests(t, handler)
	})
}

func createTestServer(t *testing.T) *Server {
	server := &Server{
		tools:         NewToolRegistry(),
		sessions:      NewSessionManager(),
		stateManager:  nil, // Would be real state manager in production
		nrqlValidator: nil, // Would be real validator in production
		mockGenerator: NewMockDataGenerator(),
		config: ServerConfig{
			RequestTimeout: 30 * time.Second,
			MockMode:       true,
		},
	}
	
	server.protocol = &ProtocolHandler{
		server:   server,
		requests: sync.Map{},
	}
	
	return server
}

func testInitializeSession(t *testing.T, handler *ProtocolHandler) {
	// Test session initialization
	req := Request{
		Jsonrpc: "2.0",
		Method:  "initialize",
		Params: json.RawMessage(`{
			"protocolVersion": "0.1.0",
			"capabilities": {
				"tools": {},
				"prompts": {}
			},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		}`),
		ID: "init-1",
	}
	
	response := executeRequest(t, handler, req)
	
	// Verify initialization response
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")
	
	// Check protocol version
	protocolVersion, ok := result["protocolVersion"].(string)
	assert.True(t, ok)
	assert.Equal(t, "0.1.0", protocolVersion)
	
	// Check server info
	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	require.True(t, ok, "Should have serverInfo")
	assert.Equal(t, "New Relic MCP Server", serverInfo["name"])
	
	// Check capabilities
	capabilities, ok := result["capabilities"].(map[string]interface{})
	require.True(t, ok, "Should have capabilities")
	assert.Contains(t, capabilities, "tools")
}

func testListTools(t *testing.T, handler *ProtocolHandler) {
	req := Request{
		Jsonrpc: "2.0",
		Method:  "tools/list",
		ID:      "list-1",
	}
	
	response := executeRequest(t, handler, req)
	
	// Check tools list
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok)
	
	tools, ok := result["tools"].([]interface{})
	require.True(t, ok, "Should have tools array")
	require.NotEmpty(t, tools, "Should have registered tools")
	
	// Verify tool structure
	for _, toolInterface := range tools {
		tool, ok := toolInterface.(map[string]interface{})
		require.True(t, ok)
		
		// Required fields
		assert.Contains(t, tool, "name")
		assert.Contains(t, tool, "description")
		assert.Contains(t, tool, "inputSchema")
		
		// Check input schema
		schema, ok := tool["inputSchema"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "object", schema["type"])
	}
	
	// Check for specific tools
	toolNames := make([]string, 0, len(tools))
	for _, t := range tools {
		if tool, ok := t.(map[string]interface{}); ok {
			if name, ok := tool["name"].(string); ok {
				toolNames = append(toolNames, name)
			}
		}
	}
	
	// Verify key tools are present
	assert.Contains(t, toolNames, "query_nrdb")
	assert.Contains(t, toolNames, "discovery.explore_event_types")
	assert.Contains(t, toolNames, "list_dashboards")
	assert.Contains(t, toolNames, "analysis.detect_anomalies")
}

func testDiscoveryFlow(t *testing.T, handler *ProtocolHandler) {
	// Step 1: Discover event types
	req := Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "discovery.explore_event_types",
			"arguments": {
				"time_range": "24 hours",
				"limit": 10
			}
		}`),
		ID: "discovery-1",
	}
	
	response := executeRequest(t, handler, req)
	result := extractToolResult(t, response)
	
	// Verify event types discovered
	eventTypes, ok := result["event_types"].([]interface{})
	require.True(t, ok, "Should have event_types array")
	require.NotEmpty(t, eventTypes, "Should discover some event types")
	
	// Step 2: Explore attributes for first event type
	if len(eventTypes) > 0 {
		firstEvent := eventTypes[0].(map[string]interface{})
		eventTypeName := firstEvent["name"].(string)
		
		req = Request{
			Jsonrpc: "2.0",
			Method:  "tools/call",
			Params: json.RawMessage(fmt.Sprintf(`{
				"name": "discovery.explore_attributes",
				"arguments": {
					"event_type": "%s",
					"sample_size": 100
				}
			}`, eventTypeName)),
			ID: "discovery-2",
		}
		
		response = executeRequest(t, handler, req)
		result = extractToolResult(t, response)
		
		// Verify attributes discovered
		attributes, ok := result["attributes"].([]interface{})
		require.True(t, ok, "Should have attributes array")
		require.NotEmpty(t, attributes, "Should discover some attributes")
		
		// Check attribute structure
		for _, attrInterface := range attributes {
			attr, ok := attrInterface.(map[string]interface{})
			require.True(t, ok)
			assert.Contains(t, attr, "name")
			assert.Contains(t, attr, "type")
			assert.Contains(t, attr, "coverage")
		}
	}
}

func testQueryExecution(t *testing.T, handler *ProtocolHandler) {
	// Test basic NRQL query
	req := Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "query_nrdb",
			"arguments": {
				"query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
			}
		}`),
		ID: "query-1",
	}
	
	response := executeRequest(t, handler, req)
	result := extractToolResult(t, response)
	
	// Verify query results
	results, ok := result["results"].([]interface{})
	require.True(t, ok, "Should have results array")
	require.NotEmpty(t, results, "Should have query results")
	
	// Test adaptive NRQL with validation
	req = Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "nrql.execute",
			"arguments": {
				"query": "SELECT average(duration) FROM Transaction WHERE appName = 'test-app'",
				"include_metadata": true
			}
		}`),
		ID: "query-2",
	}
	
	response = executeRequest(t, handler, req)
	result = extractToolResult(t, response)
	
	// Check for metadata
	metadata, ok := result["metadata"].(map[string]interface{})
	if ok {
		assert.Contains(t, metadata, "executionTime")
		assert.Contains(t, metadata, "rowCount")
	}
}

func testBatchRequests(t *testing.T, handler *ProtocolHandler) {
	// Create batch request
	batch := []Request{
		{
			Jsonrpc: "2.0",
			Method:  "tools/list",
			ID:      "batch-1",
		},
		{
			Jsonrpc: "2.0",
			Method:  "tools/call",
			Params: json.RawMessage(`{
				"name": "query_nrdb",
				"arguments": {"query": "SELECT count(*) FROM Transaction"}
			}`),
			ID: "batch-2",
		},
		{
			Jsonrpc: "2.0",
			Method:  "tools/call",
			Params: json.RawMessage(`{
				"name": "list_dashboards",
				"arguments": {"limit": 5}
			}`),
			ID: "batch-3",
		},
	}
	
	batchBytes, err := json.Marshal(batch)
	require.NoError(t, err)
	
	responseBytes, err := handler.HandleMessage(context.Background(), batchBytes)
	require.NoError(t, err)
	
	// Parse batch response
	var responses []Response
	err = json.Unmarshal(responseBytes, &responses)
	require.NoError(t, err)
	
	// Verify we got responses for all requests
	assert.Len(t, responses, 3)
	
	// Check each response
	for i, resp := range responses {
		assert.Equal(t, "2.0", resp.Jsonrpc)
		assert.Equal(t, batch[i].ID, resp.ID)
		
		if resp.Error != nil {
			t.Errorf("Batch request %d failed: %v", i, resp.Error)
		}
	}
}

func testErrorHandling(t *testing.T, handler *ProtocolHandler) {
	testCases := []struct {
		name          string
		request       Request
		expectedError int
		errorContains string
	}{
		{
			name: "Invalid tool name",
			request: Request{
				Jsonrpc: "2.0",
				Method:  "tools/call",
				Params: json.RawMessage(`{
					"name": "nonexistent.tool",
					"arguments": {}
				}`),
				ID: "error-1",
			},
			expectedError: ToolNotFoundCode,
			errorContains: "not found",
		},
		{
			name: "Missing required parameter",
			request: Request{
				Jsonrpc: "2.0",
				Method:  "tools/call",
				Params: json.RawMessage(`{
					"name": "query_nrdb",
					"arguments": {}
				}`),
				ID: "error-2",
			},
			expectedError: InvalidParamsCode,
			errorContains: "required",
		},
		{
			name: "Invalid parameter type",
			request: Request{
				Jsonrpc: "2.0",
				Method:  "tools/call",
				Params: json.RawMessage(`{
					"name": "list_dashboards",
					"arguments": {
						"limit": "not-a-number"
					}
				}`),
				ID: "error-3",
			},
			expectedError: InvalidParamsCode,
			errorContains: "invalid",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := executeRequest(t, handler, tc.request)
			
			require.NotNil(t, response.Error, "Expected error response")
			assert.Equal(t, tc.expectedError, response.Error.Code)
			assert.Contains(t, strings.ToLower(response.Error.Message), tc.errorContains)
		})
	}
}

func testPagination(t *testing.T, handler *ProtocolHandler) {
	// Test dashboard pagination
	req := Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "list_dashboards",
			"arguments": {
				"limit": 2
			}
		}`),
		ID: "page-1",
	}
	
	response := executeRequest(t, handler, req)
	result := extractToolResult(t, response)
	
	// Check pagination info
	if cursor, ok := result["next_cursor"].(string); ok {
		// Request next page
		req = Request{
			Jsonrpc: "2.0",
			Method:  "tools/call",
			Params: json.RawMessage(fmt.Sprintf(`{
				"name": "list_dashboards",
				"arguments": {
					"limit": 2,
					"cursor": "%s"
				}
			}`, cursor)),
			ID: "page-2",
		}
		
		response = executeRequest(t, handler, req)
		result = extractToolResult(t, response)
		
		dashboards, ok := result["dashboards"].([]interface{})
		assert.True(t, ok)
		assert.LessOrEqual(t, len(dashboards), 2)
	}
	
	// Test event type pagination
	req = Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "discovery.explore_event_types",
			"arguments": {
				"limit": 5,
				"offset": 0
			}
		}`),
		ID: "page-3",
	}
	
	response = executeRequest(t, handler, req)
	result = extractToolResult(t, response)
	
	// Check discovery metadata for pagination
	if metadata, ok := result["discovery_metadata"].(map[string]interface{}); ok {
		if pagination, ok := metadata["pagination"].(map[string]interface{}); ok {
			assert.Contains(t, pagination, "limit")
			assert.Contains(t, pagination, "offset")
			assert.Contains(t, pagination, "has_more")
		}
	}
}

func testMultiAccountSupport(t *testing.T, handler *ProtocolHandler) {
	// Test query with specific account
	req := Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "query_nrdb",
			"arguments": {
				"query": "SELECT count(*) FROM Transaction",
				"account_id": "12345"
			}
		}`),
		ID: "account-1",
	}
	
	response := executeRequest(t, handler, req)
	require.Nil(t, response.Error, "Should handle account_id parameter")
	
	// Test dashboard list with account
	req = Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "list_dashboards",
			"arguments": {
				"account_id": "67890"
			}
		}`),
		ID: "account-2",
	}
	
	response = executeRequest(t, handler, req)
	require.Nil(t, response.Error, "Should handle account_id in dashboard list")
}

func testWorkflowOrchestration(t *testing.T, handler *ProtocolHandler) {
	// Test a simple workflow: discover -> analyze -> recommend
	
	// Step 1: Discover event types
	req := Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "discovery.explore_event_types",
			"arguments": {
				"time_range": "1 hour"
			}
		}`),
		ID: "workflow-1",
	}
	
	response := executeRequest(t, handler, req)
	result := extractToolResult(t, response)
	
	eventTypes, ok := result["event_types"].([]interface{})
	require.True(t, ok)
	require.NotEmpty(t, eventTypes)
	
	// Step 2: Analyze anomalies for first event type
	firstEvent := eventTypes[0].(map[string]interface{})
	eventType := firstEvent["name"].(string)
	
	req = Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(fmt.Sprintf(`{
			"name": "analysis.detect_anomalies",
			"arguments": {
				"metric": "count(*)",
				"event_type": "%s",
				"time_range": "1 hour"
			}
		}`, eventType)),
		ID: "workflow-2",
	}
	
	response = executeRequest(t, handler, req)
	result = extractToolResult(t, response)
	
	// Verify analysis results
	assert.Contains(t, result, "anomaliesDetected")
	assert.Contains(t, result, "recommendations")
	
	// Step 3: Get governance recommendations
	req = Request{
		Jsonrpc: "2.0",
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name": "governance.optimize_costs",
			"arguments": {
				"focus_area": "query_performance"
			}
		}`),
		ID: "workflow-3",
	}
	
	response = executeRequest(t, handler, req)
	result = extractToolResult(t, response)
	
	recommendations, ok := result["recommendations"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, recommendations)
}

func testConcurrentRequests(t *testing.T, handler *ProtocolHandler) {
	// Test handling multiple concurrent requests
	var wg sync.WaitGroup
	results := make(chan *Response, 10)
	
	// Launch 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			req := Request{
				Jsonrpc: "2.0",
				Method:  "tools/call",
				Params: json.RawMessage(fmt.Sprintf(`{
					"name": "query_nrdb",
					"arguments": {
						"query": "SELECT count(*) FROM Transaction WHERE id = %d"
					}
				}`, id)),
				ID: fmt.Sprintf("concurrent-%d", id),
			}
			
			response := executeRequest(t, handler, req)
			results <- &response
		}(i)
	}
	
	// Wait for all requests to complete
	wg.Wait()
	close(results)
	
	// Verify all requests succeeded
	successCount := 0
	for response := range results {
		if response.Error == nil {
			successCount++
		}
	}
	
	assert.Equal(t, 10, successCount, "All concurrent requests should succeed")
}

// Helper functions

func executeRequest(t *testing.T, handler *ProtocolHandler, req Request) Response {
	reqBytes, err := json.Marshal(req)
	require.NoError(t, err)
	
	respBytes, err := handler.HandleMessage(context.Background(), reqBytes)
	require.NoError(t, err)
	require.NotNil(t, respBytes, "Should have response")
	
	var response Response
	err = json.Unmarshal(respBytes, &response)
	require.NoError(t, err)
	
	return response
}

func extractToolResult(t *testing.T, response Response) map[string]interface{} {
	require.Nil(t, response.Error, "Should not have error")
	
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")
	
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Should have content array")
	require.NotEmpty(t, content, "Content should not be empty")
	
	firstContent, ok := content[0].(map[string]interface{})
	require.True(t, ok, "First content should be a map")
	
	toolResult, ok := firstContent["text"].(string)
	if ok {
		// Parse JSON text result
		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(toolResult), &parsed)
		if err == nil {
			return parsed
		}
	}
	
	// Return raw content if not JSON text
	return firstContent
}

// TestMCPComplianceValidation validates MCP protocol compliance
func TestMCPComplianceValidation(t *testing.T) {
	handler := &ProtocolHandler{
		server: createTestServer(t),
	}
	
	t.Run("JSON-RPC 2.0 Compliance", func(t *testing.T) {
		// Test notification handling (no response expected)
		notification := []byte(`{"jsonrpc":"2.0","method":"notifications/changed"}`)
		response, err := handler.HandleMessage(context.Background(), notification)
		assert.NoError(t, err)
		assert.Nil(t, response, "Notifications should not return response")
		
		// Test batch with mixed requests and notifications
		batch := []byte(`[
			{"jsonrpc":"2.0","method":"tools/list","id":1},
			{"jsonrpc":"2.0","method":"notifications/changed"},
			{"jsonrpc":"2.0","method":"tools/list","id":2}
		]`)
		
		response, err = handler.HandleMessage(context.Background(), batch)
		assert.NoError(t, err)
		
		var responses []json.RawMessage
		err = json.Unmarshal(response, &responses)
		assert.NoError(t, err)
		assert.Len(t, responses, 2, "Should only return responses for requests with IDs")
	})
	
	t.Run("MCP Method Namespace", func(t *testing.T) {
		// Verify all registered methods follow MCP naming convention
		tools := handler.server.tools.List()
		
		for _, tool := range tools {
			// Tool names should use dot notation for namespacing
			parts := strings.Split(tool.Name, ".")
			if len(parts) > 1 {
				// First part should be a valid namespace
				validNamespaces := []string{
					"discovery", "analysis", "governance", "nrql",
					"workflow", "bulk", "query",
				}
				
				found := false
				for _, ns := range validNamespaces {
					if parts[0] == ns {
						found = true
						break
					}
				}
				
				if !found && !strings.Contains(tool.Name, "_") {
					t.Errorf("Tool %s uses non-standard namespace", tool.Name)
				}
			}
		}
	})
}
package framework

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// MCPTestClient wraps an MCP server process for testing
type MCPTestClient struct {
	account       *TestAccount
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	mu            sync.Mutex
	requestID     int
	pendingCalls  map[int]chan *MCPResponse
	serverPath    string
	debug         bool
}

// MCPRequest represents a JSON-RPC request to the MCP server
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
	ID      int                    `json:"id"`
}

// MCPResponse represents a JSON-RPC response from the MCP server
type MCPResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *MCPError              `json:"error,omitempty"`
	ID      int                    `json:"id"`
}

// MCPError represents an error in the MCP response
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPTestClient creates a new MCP test client
func NewMCPTestClient(account *TestAccount) *MCPTestClient {
	serverPath := os.Getenv("MCP_SERVER_PATH")
	if serverPath == "" {
		// Try to find the server binary
		paths := []string{
			"./bin/mcp-server",
			"../../bin/mcp-server",
			"../../../bin/mcp-server",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				serverPath = path
				break
			}
		}
	}
	
	return &MCPTestClient{
		account:      account,
		serverPath:   serverPath,
		pendingCalls: make(map[int]chan *MCPResponse),
		debug:        os.Getenv("E2E_DEBUG") == "true",
	}
}

// Start starts the MCP server process
func (c *MCPTestClient) Start(ctx context.Context) error {
	if c.serverPath == "" {
		return fmt.Errorf("MCP server binary not found")
	}
	
	// Set up environment for the MCP server
	env := os.Environ()
	env = append(env, fmt.Sprintf("NEW_RELIC_API_KEY=%s", c.account.APIKey))
	env = append(env, fmt.Sprintf("NEW_RELIC_ACCOUNT_ID=%s", c.account.AccountID))
	env = append(env, fmt.Sprintf("NEW_RELIC_REGION=%s", c.account.Region))
	
	// Start the MCP server
	c.cmd = exec.CommandContext(ctx, c.serverPath)
	c.cmd.Env = env
	
	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}
	
	// Start reading responses
	go c.readResponses()
	
	// If debug is enabled, also log stderr
	if c.debug {
		go c.logStderr()
	}
	
	// Wait a bit for the server to start
	time.Sleep(500 * time.Millisecond)
	
	return nil
}

// Stop stops the MCP server process
func (c *MCPTestClient) Stop() error {
	if c.cmd != nil && c.cmd.Process != nil {
		c.stdin.Close()
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}
	return nil
}

// ExecuteTool executes an MCP tool and returns the result
func (c *MCPTestClient) ExecuteTool(ctx context.Context, tool string, params map[string]interface{}) (interface{}, error) {
	c.mu.Lock()
	c.requestID++
	id := c.requestID
	respChan := make(chan *MCPResponse, 1)
	c.pendingCalls[id] = respChan
	c.mu.Unlock()
	
	// Create the request
	req := MCPRequest{
		JSONRPC: "2.0",
		Method:  tool,
		Params:  params,
		ID:      id,
	}
	
	// Send the request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	if c.debug {
		fmt.Printf("MCP Request: %s\n", string(data))
	}
	
	// Write length header first (4 bytes, little endian)
	length := int32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length)
	lengthBytes[1] = byte(length >> 8)
	lengthBytes[2] = byte(length >> 16)
	lengthBytes[3] = byte(length >> 24)
	
	if _, err := c.stdin.Write(lengthBytes); err != nil {
		return nil, fmt.Errorf("failed to write length header: %w", err)
	}
	
	if _, err := c.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	
	// Wait for response
	select {
	case resp := <-respChan:
		if resp.Error != nil {
			return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
		}
		return resp.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for MCP response")
	}
}

// readResponses reads responses from the MCP server
func (c *MCPTestClient) readResponses() {
	reader := bufio.NewReader(c.stdout)
	for {
		// Read length header (4 bytes, little endian)
		lengthBytes := make([]byte, 4)
		_, err := io.ReadFull(reader, lengthBytes)
		if err != nil {
			if err == io.EOF {
				return
			}
			if c.debug {
				fmt.Printf("Failed to read length header: %v\n", err)
			}
			continue
		}
		
		length := int32(lengthBytes[0]) | int32(lengthBytes[1])<<8 | int32(lengthBytes[2])<<16 | int32(lengthBytes[3])<<24
		
		// Read message body
		message := make([]byte, length)
		_, err = io.ReadFull(reader, message)
		if err != nil {
			if c.debug {
				fmt.Printf("Failed to read message body: %v\n", err)
			}
			continue
		}
		
		if c.debug {
			fmt.Printf("MCP Response: %s\n", string(message))
		}
		
		var resp MCPResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			if c.debug {
				fmt.Printf("Failed to unmarshal response: %v\n", err)
			}
			continue
		}
		
		c.mu.Lock()
		if ch, ok := c.pendingCalls[resp.ID]; ok {
			ch <- &resp
			delete(c.pendingCalls, resp.ID)
		}
		c.mu.Unlock()
	}
}

// logStderr logs stderr output for debugging
func (c *MCPTestClient) logStderr() {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		fmt.Printf("MCP stderr: %s\n", scanner.Text())
	}
}
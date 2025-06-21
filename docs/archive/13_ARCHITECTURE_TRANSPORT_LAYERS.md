# Transport Layers Architecture

This document specifies the transport layer architecture of the New Relic MCP Server, defining how clients communicate with the server across different protocols.

## Table of Contents

1. [Overview](#overview)
2. [Transport Abstraction](#transport-abstraction)
3. [STDIO Transport](#stdio-transport)
4. [HTTP Transport](#http-transport)
5. [Server-Sent Events (SSE)](#server-sent-events-sse)
6. [Protocol Implementation](#protocol-implementation)
7. [Message Flow](#message-flow)
8. [Error Handling](#error-handling)
9. [Performance Considerations](#performance-considerations)
10. [Future Transports](#future-transports)

## Overview

The transport layer SHALL provide protocol-agnostic communication between MCP clients and the server, supporting multiple transport mechanisms while maintaining a consistent interface.

### Design Principles

1. **Protocol Agnostic**: Tools work identically across all transports
2. **Pluggable Architecture**: Easy to add new transports
3. **Consistent Interface**: Uniform API regardless of transport
4. **Performance Optimized**: Minimal overhead per transport
5. **Error Resilient**: Graceful handling of transport failures

### Transport Selection

```go
type TransportType string

const (
    TransportSTDIO TransportType = "stdio"
    TransportHTTP  TransportType = "http"
    TransportSSE   TransportType = "sse"
)

func NewTransport(t TransportType, config Config) (Transport, error) {
    switch t {
    case TransportSTDIO:
        return NewSTDIOTransport(config)
    case TransportHTTP:
        return NewHTTPTransport(config)
    case TransportSSE:
        return NewSSETransport(config)
    default:
        return nil, ErrUnknownTransport
    }
}
```

## Transport Abstraction

### Common Interface

All transports SHALL implement a common interface:

```go
type Transport interface {
    // Start the transport
    Start(ctx context.Context) error
    
    // Stop the transport
    Stop(ctx context.Context) error
    
    // Send a message to client
    Send(msg Message) error
    
    // Receive messages from client
    Receive() <-chan Message
    
    // Get transport metadata
    Metadata() TransportMetadata
}

type Message interface {
    ID() string
    Method() string
    Params() interface{}
    IsNotification() bool
}

type TransportMetadata struct {
    Type         TransportType
    Capabilities []string
    MaxMsgSize   int
    Streaming    bool
}
```

### Message Codec

```go
type Codec interface {
    Encode(v interface{}) ([]byte, error)
    Decode(data []byte, v interface{}) error
}

type JSONRPCCodec struct{}

func (c JSONRPCCodec) Encode(v interface{}) ([]byte, error) {
    return json.Marshal(v)
}

func (c JSONRPCCodec) Decode(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)
}
```

## STDIO Transport

The STDIO transport SHALL provide bidirectional communication over standard input/output streams.

### Characteristics

- **Protocol**: Line-delimited JSON-RPC 2.0
- **Encoding**: UTF-8 JSON with newline delimiter
- **Buffering**: Line-buffered for real-time communication
- **Use Cases**: Claude Desktop, CLI tools, embedded scenarios

### Implementation

```go
type STDIOTransport struct {
    reader  *bufio.Reader
    writer  *bufio.Writer
    codec   Codec
    recv    chan Message
    done    chan struct{}
}

func (t *STDIOTransport) Start(ctx context.Context) error {
    t.reader = bufio.NewReader(os.Stdin)
    t.writer = bufio.NewWriter(os.Stdout)
    
    go t.readLoop(ctx)
    return nil
}

func (t *STDIOTransport) readLoop(ctx context.Context) {
    defer close(t.recv)
    
    scanner := bufio.NewScanner(t.reader)
    for scanner.Scan() {
        select {
        case <-ctx.Done():
            return
        default:
            line := scanner.Bytes()
            msg, err := t.decode(line)
            if err != nil {
                t.sendError(err)
                continue
            }
            t.recv <- msg
        }
    }
}

func (t *STDIOTransport) Send(msg Message) error {
    data, err := t.codec.Encode(msg)
    if err != nil {
        return err
    }
    
    if _, err := t.writer.Write(data); err != nil {
        return err
    }
    
    if err := t.writer.WriteByte('\n'); err != nil {
        return err
    }
    
    return t.writer.Flush()
}
```

### Message Format

```json
// Request
{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}

// Response
{"jsonrpc":"2.0","id":1,"result":[...]}

// Notification
{"jsonrpc":"2.0","method":"progress","params":{"value":50}}

// Error
{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"Invalid params"}}
```

### Connection Lifecycle

```
Process Start
     │
     ▼
Initialize STDIO
     │
     ▼
Send Ready Signal ──► {"jsonrpc":"2.0","method":"ready"}
     │
     ▼
Process Messages ◄─┐
     │             │
     ▼             │
Handle Request ────┘
     │
     ▼
Process Exit
```

## HTTP Transport

The HTTP transport SHALL provide request/response communication over HTTP/HTTPS.

### Characteristics

- **Protocol**: JSON-RPC 2.0 over HTTP POST
- **Encoding**: application/json
- **Security**: TLS support, CORS headers
- **Use Cases**: Web applications, REST clients

### Endpoints

```
POST /api/v1/jsonrpc    - Main JSON-RPC endpoint
GET  /api/v1/health     - Health check
GET  /api/v1/metrics    - Prometheus metrics
```

### Implementation

```go
type HTTPTransport struct {
    server   *http.Server
    handler  http.Handler
    codec    Codec
    recv     chan Message
    shutdown chan struct{}
}

func (t *HTTPTransport) Start(ctx context.Context) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/v1/jsonrpc", t.handleJSONRPC)
    mux.HandleFunc("/api/v1/health", t.handleHealth)
    
    t.server = &http.Server{
        Addr:         ":8080",
        Handler:      t.middleware(mux),
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }
    
    go func() {
        if err := t.server.ListenAndServe(); err != nil {
            log.Error("HTTP server error:", err)
        }
    }()
    
    return nil
}

func (t *HTTPTransport) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req JSONRPCRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        t.writeError(w, InvalidRequest(err))
        return
    }
    
    // Process request
    msg := t.requestToMessage(req)
    t.recv <- msg
    
    // Wait for response
    response := <-t.waitForResponse(msg.ID())
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### Security Headers

```go
func (t *HTTPTransport) middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        // Security headers
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        
        next.ServeHTTP(w, r)
    })
}
```

## Server-Sent Events (SSE)

The SSE transport SHALL provide unidirectional streaming from server to client.

### Characteristics

- **Protocol**: SSE over HTTP
- **Encoding**: UTF-8 text with SSE format
- **Direction**: Server-to-client only
- **Use Cases**: Real-time updates, progress monitoring

### Implementation

```go
type SSETransport struct {
    server   *http.Server
    clients  map[string]*SSEClient
    mu       sync.RWMutex
    events   chan SSEEvent
}

type SSEClient struct {
    id       string
    events   chan SSEEvent
    done     chan struct{}
}

func (t *SSETransport) handleSSE(w http.ResponseWriter, r *http.Request) {
    // Check for SSE support
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "SSE not supported", http.StatusInternalServerError)
        return
    }
    
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // Create client
    client := &SSEClient{
        id:     generateID(),
        events: make(chan SSEEvent, 100),
        done:   make(chan struct{}),
    }
    
    t.addClient(client)
    defer t.removeClient(client)
    
    // Send events
    for {
        select {
        case event := <-client.events:
            fmt.Fprintf(w, "id: %s\n", event.ID)
            fmt.Fprintf(w, "event: %s\n", event.Type)
            fmt.Fprintf(w, "data: %s\n\n", event.Data)
            flusher.Flush()
            
        case <-r.Context().Done():
            return
        }
    }
}
```

### Event Format

```
id: msg-123
event: progress
data: {"method":"discovery.explore","progress":75,"message":"Analyzing attributes..."}

id: msg-124
event: result
data: {"id":1,"result":{"schemas":["Transaction","SystemSample"]}}

event: ping
data: {"timestamp":"2024-01-20T10:30:00Z"}
```

## Protocol Implementation

### JSON-RPC 2.0 Compliance

All transports SHALL implement JSON-RPC 2.0 specification:

```go
type JSONRPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id,omitempty"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id,omitempty"`
    Result  interface{} `json:"result,omitempty"`
    Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Standard error codes
const (
    ParseError     = -32700
    InvalidRequest = -32600
    MethodNotFound = -32601
    InvalidParams  = -32602
    InternalError  = -32603
)
```

### MCP Extensions

```go
// MCP-specific methods
const (
    MethodInitialize    = "initialize"
    MethodToolsList     = "tools/list"
    MethodToolsCall     = "tools/call"
    MethodResourcesList = "resources/list"
    MethodResourcesRead = "resources/read"
)

// MCP metadata
type MCPRequest struct {
    JSONRPCRequest
    Metadata MCPMetadata `json:"_meta,omitempty"`
}

type MCPMetadata struct {
    ProgressToken string            `json:"progressToken,omitempty"`
    Capabilities  []string          `json:"capabilities,omitempty"`
    Context       map[string]string `json:"context,omitempty"`
}
```

## Message Flow

### Request Processing Pipeline

```
Transport Layer
     │
     ▼
Decode Message
     │
     ▼
Validate Format ──► INVALID ──► Error Response
     │
   VALID
     │
     ▼
Route to Handler
     │
     ▼
Execute Tool
     │
     ▼
Format Response
     │
     ▼
Encode Message
     │
     ▼
Send Response
```

### Streaming Flow

For long-running operations:

```
Request Received
     │
     ▼
Start Operation
     │
     ▼
Send Progress ──┐
     │          │
     ▼          │
Continue ───────┘
     │
     ▼
Send Result
```

### Batch Request Handling

```go
type BatchRequest []JSONRPCRequest

func processBatch(batch BatchRequest) []JSONRPCResponse {
    responses := make([]JSONRPCResponse, len(batch))
    
    // Process in parallel with limit
    sem := make(chan struct{}, maxConcurrent)
    var wg sync.WaitGroup
    
    for i, req := range batch {
        wg.Add(1)
        go func(idx int, request JSONRPCRequest) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            
            responses[idx] = processRequest(request)
        }(i, req)
    }
    
    wg.Wait()
    return responses
}
```

## Error Handling

### Transport-Level Errors

```go
type TransportError struct {
    Code      int
    Message   string
    Transport TransportType
    Cause     error
}

func (e TransportError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Transport, e.Message, e.Cause)
}

// Common transport errors
var (
    ErrTransportClosed   = TransportError{Code: 1001, Message: "Transport closed"}
    ErrMessageTooLarge   = TransportError{Code: 1002, Message: "Message exceeds size limit"}
    ErrInvalidMessage    = TransportError{Code: 1003, Message: "Invalid message format"}
    ErrTransportTimeout  = TransportError{Code: 1004, Message: "Transport timeout"}
)
```

### Error Recovery

```go
func (t *Transport) recoverableError(err error) bool {
    switch err {
    case io.EOF, io.ErrUnexpectedEOF:
        return false // Connection closed
    case context.Canceled, context.DeadlineExceeded:
        return false // Intentional shutdown
    default:
        return true // Try to recover
    }
}

func (t *Transport) handleError(err error) {
    if t.recoverableError(err) {
        log.Warn("Recoverable transport error:", err)
        // Continue processing
    } else {
        log.Error("Fatal transport error:", err)
        t.Stop(context.Background())
    }
}
```

## Performance Considerations

### Buffer Management

```go
type BufferPool struct {
    pool sync.Pool
}

func NewBufferPool() *BufferPool {
    return &BufferPool{
        pool: sync.Pool{
            New: func() interface{} {
                return make([]byte, 0, 4096)
            },
        },
    }
}

func (p *BufferPool) Get() []byte {
    return p.pool.Get().([]byte)
}

func (p *BufferPool) Put(buf []byte) {
    buf = buf[:0]
    p.pool.Put(buf)
}
```

### Connection Pooling

For HTTP transport:

```go
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
    },
    Timeout: 30 * time.Second,
}
```

### Message Compression

```go
func compressMessage(msg []byte) []byte {
    if len(msg) < compressionThreshold {
        return msg
    }
    
    var buf bytes.Buffer
    w := gzip.NewWriter(&buf)
    w.Write(msg)
    w.Close()
    
    return buf.Bytes()
}
```

### Metrics

```yaml
# Transport metrics
transport_messages_total{transport, direction, status}
transport_message_size_bytes{transport, direction}
transport_latency_seconds{transport, operation}
transport_connections_active{transport}
transport_errors_total{transport, error_type}
```

## Future Transports

### WebSocket Transport

```go
type WebSocketTransport struct {
    upgrader websocket.Upgrader
    conns    map[string]*websocket.Conn
}

// Bidirectional, persistent connection
// Lower latency than HTTP
// Real-time capabilities
```

### gRPC Transport

```go
type GRPCTransport struct {
    server *grpc.Server
    // Protocol buffers for efficiency
    // Streaming support
    // Built-in load balancing
}
```

### QUIC Transport

```go
type QUICTransport struct {
    listener quic.Listener
    // UDP-based, multiplexed streams
    // Better performance over lossy networks
    // Built-in encryption
}
```

### GraphQL Subscription

```go
type GraphQLTransport struct {
    schema graphql.Schema
    // Subscription support
    // Flexible query language
    // Type safety
}
```

## Transport Selection Guide

| Use Case | Recommended Transport | Reasoning |
|----------|---------------------|-----------|
| Claude Desktop | STDIO | Direct process communication |
| Web Dashboard | HTTP/SSE | Browser compatibility |
| CLI Tools | STDIO | Simple integration |
| Mobile Apps | HTTP | Universal support |
| Real-time Monitoring | SSE/WebSocket | Low latency updates |
| High-throughput | gRPC | Binary protocol efficiency |
| Microservices | gRPC/HTTP | Service mesh compatibility |

## Best Practices

1. **Choose Appropriate Transport**: Match transport to use case
2. **Handle Disconnections**: Implement reconnection logic
3. **Monitor Performance**: Track latency and throughput
4. **Implement Timeouts**: Prevent hanging connections
5. **Use Compression**: For large messages
6. **Pool Resources**: Reuse connections and buffers
7. **Test Error Scenarios**: Network failures, timeouts

## Conclusion

The transport layer architecture provides flexible, efficient communication options for different client scenarios. The abstraction ensures that business logic remains transport-agnostic while allowing optimization for specific use cases.

## Related Documentation

- [Architecture Overview](10_ARCHITECTURE_OVERVIEW.md) - System architecture
- [Protocol Reference](21_API_MCP_PROTOCOL.md) - MCP protocol details
- [API Reference](20_API_OVERVIEW.md) - API documentation
- [Performance Tuning](15_ARCHITECTURE_SCALABILITY.md) - Optimization guide

---

**Implementation Note**: The current implementation supports STDIO, HTTP, and SSE transports. See `pkg/interface/mcp/transport_*.go` for implementation details.
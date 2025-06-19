#!/bin/bash

set -e

echo "=== Simple End-to-End Test ==="
echo "Starting MCP server..."

# Kill any existing server
pkill -f mcp-server || true

# Start server in background
./bin/mcp-server -transport http -port 8080 > server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 5

echo "Server started with PID: $SERVER_PID"

# Function to make MCP call
mcp_call() {
    local method=$1
    local tool=$2
    local args=${3:-'{}'}
    
    curl -s -X POST http://localhost:8080/mcp \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"method\": \"$method\",
            \"params\": {
                \"name\": \"$tool\",
                \"arguments\": $args
            },
            \"id\": 1
        }"
}

echo -e "\n1. Testing query_nrdb with simple query..."
RESULT=$(mcp_call "tools/call" "query_nrdb" '{"query": "SELECT uniqueCount(eventType) FROM Metric SINCE 1 hour ago"}')
echo "$RESULT" | grep -q "result" && echo "✓ Query executed successfully" || echo "✗ Query failed"

echo -e "\n2. Testing list_dashboards..."  
RESULT=$(mcp_call "tools/call" "list_dashboards" '{}')
echo "$RESULT" | grep -q "result" && echo "✓ Dashboards listed successfully" || echo "✗ List dashboards failed"

echo -e "\n3. Testing query_builder..."
RESULT=$(mcp_call "tools/call" "query_builder" '{"event_type": "Metric", "select": ["count(*)"], "since": "5 minutes ago"}')
echo "$RESULT" | grep -q "result" && echo "✓ Query built successfully" || echo "✗ Query builder failed"

echo -e "\n4. Testing list_alerts..."
RESULT=$(mcp_call "tools/call" "list_alerts" '{}')
echo "$RESULT" | grep -q "result" && echo "✓ Alerts listed successfully" || echo "✗ List alerts failed"

# Cleanup
echo -e "\nCleaning up..."
kill $SERVER_PID 2>/dev/null || true

echo -e "\n=== Test completed ==="#
echo "Check server.log for detailed output"
#!/bin/bash

set -e

echo "=== Testing Dashboard Creation with Links ==="

# Kill any existing server
pkill -f mcp-server || true

# Start server
echo "Starting MCP server..."
./bin/mcp-server -transport http -port 8080 > dashboard_creation.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 5

echo "Server started with PID: $SERVER_PID"

# Function to make MCP call
mcp_call() {
    local tool=$1
    local args=$2
    
    curl -s -X POST http://localhost:8080/mcp \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"method\": \"tools/call\",
            \"params\": {
                \"name\": \"$tool\",
                \"arguments\": $args
            },
            \"id\": 1
        }" | python3 -m json.tool
}

echo -e "\nCreating a simple test dashboard..."
mcp_call "generate_dashboard" '{
    "template": "custom",
    "name": "Test Dashboard with Links",
    "custom_config": {
        "description": "Testing dashboard creation to get links",
        "pages": [{
            "name": "Test Page",
            "widgets": [
                {
                    "title": "Test Widget",
                    "type": "line",
                    "row": 1,
                    "column": 1,
                    "width": 12,
                    "height": 3,
                    "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago TIMESERIES"
                }
            ]
        }]
    }
}'

# Give time for operation to complete
sleep 2

# Cleanup
echo -e "\nCleaning up..."
kill $SERVER_PID 2>/dev/null || true

echo -e "\n=== Test completed ==="
echo "Check dashboard_creation.log for server output"
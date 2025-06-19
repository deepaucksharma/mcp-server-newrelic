#!/bin/bash

echo "=== Simple Dashboard Creation Test ==="

# Kill any existing server
pkill -f mcp-server || true
sleep 2

# Start server in foreground to see logs
echo "Starting MCP server..."
./bin/mcp-server -transport http -port 8080 &
SERVER_PID=$!

# Wait for server
sleep 5

# Create dashboard
echo -e "\nCreating dashboard..."
RESULT=$(curl -s -X POST http://localhost:8080/mcp \
    -H "Content-Type: application/json" \
    -d '{
        "jsonrpc": "2.0",
        "method": "tools/call",
        "params": {
            "name": "generate_dashboard",
            "arguments": {
                "template": "golden-signals",
                "name": "Test Golden Signals Dashboard",
                "service_name": "TestService"
            }
        },
        "id": 1
    }')

echo "Response:"
echo "$RESULT" | python3 -m json.tool

# Check for dashboard URL in response
if echo "$RESULT" | grep -q "dashboard_url"; then
    echo -e "\n✅ Dashboard URL found!"
    URL=$(echo "$RESULT" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['result']['dashboard_url'])" 2>/dev/null || echo "")
    if [ ! -z "$URL" ]; then
        echo "Dashboard URL: $URL"
    fi
else
    echo -e "\n❌ No dashboard URL in response"
fi

# Cleanup
kill $SERVER_PID 2>/dev/null || true

echo -e "\n=== Test completed ==="
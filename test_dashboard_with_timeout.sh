#!/bin/bash

echo "=== Testing Dashboard Creation with Extended Timeout ==="

# Start server
echo "Starting MCP server..."
./bin/mcp-server -transport http -port 8080 > dashboard_timeout_test.log 2>&1 &
SERVER_PID=$!

# Wait for server
sleep 5

# Create dashboard
echo -e "\nCreating Golden Signals Dashboard..."
START_TIME=$(date +%s)

RESULT=$(curl -s -X POST http://localhost:8080/mcp \
    -H "Content-Type: application/json" \
    -d '{
        "jsonrpc": "2.0",
        "method": "tools/call",
        "params": {
            "name": "generate_dashboard",
            "arguments": {
                "template": "golden-signals",
                "name": "Production Service Golden Signals",
                "service_name": "my-production-app"
            }
        },
        "id": 1
    }')

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
echo "Request completed in ${DURATION} seconds"

# Parse and display result
echo -e "\nParsing response..."
if echo "$RESULT" | grep -q '"created":true'; then
    echo "✅ Dashboard created successfully!"
    
    # Extract dashboard URL
    URL=$(echo "$RESULT" | python3 -c "
import sys, json
data = json.load(sys.stdin)
content = data['result']['content'][0]['text']
dashboard_data = json.loads(content)
print(f\"Dashboard ID: {dashboard_data.get('dashboard_id', 'N/A')}\")
print(f\"Dashboard URL: {dashboard_data.get('dashboard_url', 'N/A')}\")
print(f\"Message: {dashboard_data.get('message', 'N/A')}\")
" 2>/dev/null || echo "Failed to parse")
    
    echo "$URL"
elif echo "$RESULT" | grep -q '"created":false'; then
    echo "❌ Dashboard creation failed"
    
    # Extract error
    ERROR=$(echo "$RESULT" | python3 -c "
import sys, json
data = json.load(sys.stdin)
content = data['result']['content'][0]['text']
dashboard_data = json.loads(content)
print(f\"Error: {dashboard_data.get('error', 'Unknown error')}\")
print(f\"Message: {dashboard_data.get('message', 'N/A')}\")
" 2>/dev/null || echo "Failed to parse error")
    
    echo "$ERROR"
else
    echo "❓ Unexpected response format"
    echo "$RESULT" | python3 -m json.tool
fi

# Cleanup
echo -e "\nCleaning up..."
kill $SERVER_PID 2>/dev/null || true

echo -e "\n=== Test completed ===
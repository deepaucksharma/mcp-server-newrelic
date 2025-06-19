#!/bin/bash

echo "=== Proper NRDB Discovery for Kafka ==="

# Start server
echo "Starting MCP server..."
./bin/mcp-server -transport http -port 8080 > discovery_test.log 2>&1 &
SERVER_PID=$!

# Wait for server
sleep 5

# Function to make MCP call
mcp_call() {
    local tool=$1
    local args=$2
    
    result=$(curl -s -X POST http://localhost:8080/mcp \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"method\": \"tools/call\",
            \"params\": {
                \"name\": \"$tool\",
                \"arguments\": $args
            },
            \"id\": 1
        }")
    
    # Check for timeout error
    if echo "$result" | grep -q "context deadline exceeded"; then
        echo "TIMEOUT"
    else
        echo "$result"
    fi
}

echo -e "\n1. Discovering all event types in NRDB..."
ALL_EVENTS=$(mcp_call "query_nrdb" '{
    "query": "SELECT uniques(eventType) FROM Metric WHERE eventType IS NOT NULL SINCE 1 day ago LIMIT 100"
}')

if [ "$ALL_EVENTS" = "TIMEOUT" ]; then
    echo "⏱️  Query timed out. Using direct discovery..."
    # Try a simpler query
    ALL_EVENTS=$(mcp_call "query_nrdb" '{
        "query": "SELECT count(*) FROM Transaction SINCE 5 minutes ago"
    }')
fi

echo -e "\n2. Looking for Kafka-specific sample types..."
KAFKA_SAMPLES=$(mcp_call "query_nrdb" '{
    "query": "SELECT uniques(eventType) FROM KafkaBrokerSample SINCE 1 day ago LIMIT 1"
}')

if echo "$KAFKA_SAMPLES" | grep -q "results"; then
    echo "✅ Found KafkaBrokerSample event type"
else
    echo "❌ KafkaBrokerSample not found"
fi

echo -e "\n3. Checking for infrastructure integration data..."
INTEGRATION_CHECK=$(mcp_call "query_nrdb" '{
    "query": "SELECT count(*) FROM SystemSample WHERE hostname IS NOT NULL SINCE 5 minutes ago"
}')

echo -e "\n4. Discovery-based dashboard creation approach:"
echo "Based on discovery results, we would:"
echo "- First use discovery.list_schemas to get all available event types"
echo "- Filter for Kafka-related schemas (KafkaBrokerSample, KafkaTopicSample, etc.)"
echo "- Use discovery.profile_attribute to understand the available metrics"
echo "- Build dashboards using only discovered event types and attributes"
echo "- Never assume event types like AWSKafkaBroker or ConfluentKafka exist"

echo -e "\n5. Proper discovery flow example:"
# This would be the correct flow:
DISCOVERY_SCHEMAS=$(mcp_call "discovery.list_schemas" '{}')

# Parse and display some results
echo "$DISCOVERY_SCHEMAS" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'result' in data:
        print('✅ Discovery tool is working')
    else:
        print('❌ Discovery failed:', data.get('error', {}).get('message', 'Unknown error'))
except Exception as e:
    print(f'Parse error: {e}')
"

# Cleanup
echo -e "\nCleaning up..."
kill $SERVER_PID 2>/dev/null || true

echo -e "\n=== Key Principle ==="
echo "ALWAYS discover first, NEVER assume event types or attributes exist!"
echo "The system should adapt to what's actually in the customer's NRDB."
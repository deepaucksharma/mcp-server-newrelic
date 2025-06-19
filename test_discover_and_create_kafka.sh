#!/bin/bash

echo "=== Kafka Dashboard Creation with Discovery ==="

# Start server
echo "Starting MCP server..."
./bin/mcp-server -transport http -port 8080 > kafka_discovery_test.log 2>&1 &
SERVER_PID=$!

# Wait for server
sleep 5

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
        }"
}

echo -e "\n1. First, let's discover what Kafka-related event types exist in NRDB..."
DISCOVERY_RESULT=$(mcp_call "discovery.list_schemas" '{"filter": "kafka"}')

# Extract schemas that contain "kafka" (case insensitive)
echo "$DISCOVERY_RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    content = data['result']['content'][0]['text']
    result = json.loads(content)
    schemas = result.get('schemas', [])
    
    kafka_schemas = [s for s in schemas if 'kafka' in s['name'].lower()]
    
    if kafka_schemas:
        print(f'✅ Found {len(kafka_schemas)} Kafka-related schemas:')
        for schema in kafka_schemas:
            print(f\"  - {schema['name']} ({schema.get('sample_count', 0)} samples)\")
    else:
        print('❌ No Kafka-related schemas found in NRDB')
        print('Available schemas:')
        for s in schemas[:10]:
            print(f\"  - {s['name']}\")
        if len(schemas) > 10:
            print(f'  ... and {len(schemas) - 10} more')
except Exception as e:
    print(f'Error parsing discovery result: {e}')
"

echo -e "\n2. Let's check for MSK-specific event types..."
MSK_RESULT=$(mcp_call "query_nrdb" '{
    "query": "SHOW EVENT TYPES LIKE '\''%MSK%'\'' SINCE 1 week ago"
}')

echo "$MSK_RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'error' in data:
        print(f'Query error: {data[\"error\"]}')
    else:
        content = data['result']['content'][0]['text']
        result = json.loads(content)
        if result.get('results'):
            print('Found MSK event types:')
            for r in result['results']:
                print(f\"  - {r.get('eventType', 'Unknown')}\")
        else:
            print('No MSK-specific event types found')
except Exception as e:
    print(f'Error: {e}')
"

echo -e "\n3. Let's check for Confluent-specific event types..."
CONFLUENT_RESULT=$(mcp_call "query_nrdb" '{
    "query": "SHOW EVENT TYPES LIKE '\''%Confluent%'\'' SINCE 1 week ago"
}')

echo "$CONFLUENT_RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'error' in data:
        print(f'Query error: {data[\"error\"]}')
    else:
        content = data['result']['content'][0]['text']
        result = json.loads(content)
        if result.get('results'):
            print('Found Confluent event types:')
            for r in result['results']:
                print(f\"  - {r.get('eventType', 'Unknown')}\")
        else:
            print('No Confluent-specific event types found')
except Exception as e:
    print(f'Error: {e}')
"

echo -e "\n4. Let's check what infrastructure integrations are available..."
INTEGRATION_RESULT=$(mcp_call "query_nrdb" '{
    "query": "SELECT uniques(nr.integrationName) FROM Metric WHERE nr.integrationName IS NOT NULL SINCE 1 week ago"
}')

echo "$INTEGRATION_RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'error' not in data:
        content = data['result']['content'][0]['text']
        result = json.loads(content)
        if result.get('results') and len(result['results']) > 0:
            integrations = result['results'][0].get('uniques.nr.integrationName', [])
            kafka_integrations = [i for i in integrations if 'kafka' in i.lower()]
            if kafka_integrations:
                print(f'Found Kafka integrations:')
                for i in kafka_integrations:
                    print(f'  - {i}')
            else:
                print('No Kafka integrations found')
                print(f'Available integrations: {integrations[:5]}...')
except Exception as e:
    print(f'Error: {e}')
"

echo -e "\n5. Now let's create dashboards based on what we discovered..."
echo "(In a real scenario, we would use the discovered event types and attributes to build appropriate dashboards)"

# Cleanup
echo -e "\nCleaning up..."
kill $SERVER_PID 2>/dev/null || true

echo -e "\n=== Discovery completed ==="
echo "Check kafka_discovery_test.log for server output"
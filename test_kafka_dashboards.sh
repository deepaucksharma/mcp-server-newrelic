#!/bin/bash

set -e

echo "=== Testing Kafka Dashboard Creation ==="

# Start server with longer timeout
echo "Starting MCP server..."
./bin/mcp-server -transport http -port 8080 > kafka_test.log 2>&1 &
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

echo -e "\n1. Creating MSK (AWS Managed Kafka) Dashboard..."
mcp_call "generate_dashboard" '{
    "template": "custom",
    "name": "Kafka MSK Monitoring Dashboard",
    "custom_config": {
        "description": "Monitoring dashboard for AWS MSK (Managed Streaming for Kafka)",
        "pages": [{
            "name": "MSK Overview",
            "widgets": [
                {
                    "title": "MSK Broker CPU Utilization",
                    "type": "line",
                    "row": 1,
                    "column": 1,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT average(cpuUtilization) FROM AWSKafkaBroker FACET brokerID TIMESERIES"
                },
                {
                    "title": "MSK Messages Per Second",
                    "type": "line",
                    "row": 1,
                    "column": 7,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT rate(sum(messagesProduced), 1 second) as '\''Messages/sec'\'' FROM AWSKafkaBroker TIMESERIES"
                },
                {
                    "title": "MSK Partition Count",
                    "type": "billboard",
                    "row": 4,
                    "column": 1,
                    "width": 4,
                    "height": 3,
                    "query": "SELECT uniqueCount(partition) FROM AWSKafkaTopic"
                },
                {
                    "title": "MSK Consumer Lag",
                    "type": "line",
                    "row": 4,
                    "column": 5,
                    "width": 8,
                    "height": 3,
                    "query": "SELECT average(consumerLag) FROM AWSKafkaConsumer FACET consumerGroup TIMESERIES"
                }
            ]
        }]
    }
}'

echo -e "\n2. Creating Confluent Kafka Dashboard..."
mcp_call "generate_dashboard" '{
    "template": "custom",
    "name": "Confluent Kafka Monitoring Dashboard",
    "custom_config": {
        "description": "Monitoring dashboard for Confluent Kafka Platform",
        "pages": [{
            "name": "Confluent Kafka Overview",
            "widgets": [
                {
                    "title": "Confluent Broker Request Rate",
                    "type": "line",
                    "row": 1,
                    "column": 1,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT rate(sum(kafka.server.BrokerTopicMetrics.MessagesInPerSec), 1 minute) FROM ConfluentKafkaBroker TIMESERIES"
                },
                {
                    "title": "Confluent Topic Throughput",
                    "type": "line",
                    "row": 1,
                    "column": 7,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT average(kafka.server.BrokerTopicMetrics.BytesInPerSec) as '\''Bytes In'\'', average(kafka.server.BrokerTopicMetrics.BytesOutPerSec) as '\''Bytes Out'\'' FROM ConfluentKafkaBroker TIMESERIES"
                },
                {
                    "title": "Confluent Controller Status",
                    "type": "table",
                    "row": 4,
                    "column": 1,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT latest(kafka.controller.KafkaController.ActiveControllerCount) as '\''Active Controllers'\'', latest(kafka.controller.KafkaController.OfflinePartitionsCount) as '\''Offline Partitions'\'' FROM ConfluentKafkaBroker FACET brokerName"
                },
                {
                    "title": "Confluent Replication Lag",
                    "type": "line",
                    "row": 4,
                    "column": 7,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT max(kafka.server.ReplicaManager.UnderReplicatedPartitions) FROM ConfluentKafkaBroker FACET brokerName TIMESERIES"
                }
            ]
        }]
    }
}'

echo -e "\n3. Creating OHI Kafka (nri-kafka) Dashboard..."
mcp_call "generate_dashboard" '{
    "template": "custom",
    "name": "Kafka OHI (nri-kafka) Monitoring Dashboard",
    "custom_config": {
        "description": "Monitoring dashboard for Kafka using New Relic Infrastructure On-Host Integration (nri-kafka)",
        "pages": [{
            "name": "Kafka OHI Overview",
            "widgets": [
                {
                    "title": "Kafka Broker Network I/O",
                    "type": "line",
                    "row": 1,
                    "column": 1,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT average(broker.IOInPerSecond) as '\''Network In'\'', average(broker.IOOutPerSecond) as '\''Network Out'\'' FROM KafkaBrokerSample TIMESERIES"
                },
                {
                    "title": "Kafka Messages Rate",
                    "type": "line",
                    "row": 1,
                    "column": 7,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT rate(sum(broker.messagesInPerSecond), 1 second) as '\''Messages/sec'\'' FROM KafkaBrokerSample TIMESERIES AUTO"
                },
                {
                    "title": "Kafka Topic Metrics",
                    "type": "table",
                    "row": 4,
                    "column": 1,
                    "width": 8,
                    "height": 3,
                    "query": "SELECT average(topic.partitionsCount) as '\''Partitions'\'', average(topic.underReplicatedPartitions) as '\''Under Replicated'\'', average(topic.replicationFactor) as '\''Replication Factor'\'' FROM KafkaTopicSample FACET topic.topicName LIMIT 10"
                },
                {
                    "title": "Kafka Consumer Group Lag",
                    "type": "billboard",
                    "row": 4,
                    "column": 9,
                    "width": 4,
                    "height": 3,
                    "query": "SELECT sum(consumer.lag) as '\''Total Consumer Lag'\'' FROM KafkaConsumerSample"
                },
                {
                    "title": "Kafka JVM Memory Usage",
                    "type": "line",
                    "row": 7,
                    "column": 1,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT average(broker.JVMHeapUsed) / 1e6 as '\''Heap Used (MB)'\'' FROM KafkaBrokerSample FACET entityName TIMESERIES"
                },
                {
                    "title": "Kafka Request Latency",
                    "type": "line",
                    "row": 7,
                    "column": 7,
                    "width": 6,
                    "height": 3,
                    "query": "SELECT average(request.avgRequestLatency) as '\''Avg Latency'\'', percentile(request.avgRequestLatency, 95) as '\''P95 Latency'\'' FROM KafkaRequestSample TIMESERIES"
                }
            ]
        }]
    }
}'

# Give time for operations to complete
sleep 2

echo -e "\n4. Listing all dashboards to confirm creation..."
mcp_call "list_dashboards" '{}'

# Cleanup
echo -e "\nCleaning up..."
kill $SERVER_PID 2>/dev/null || true

echo -e "\n=== Test completed ==="
echo "Check kafka_test.log for server output"
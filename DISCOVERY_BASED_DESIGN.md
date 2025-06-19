# Discovery-Based Dashboard Creation Design

## Fundamental Principle
**DISCOVER EVERYTHING FROM NRDB** - Never assume event types, attributes, or data structures exist.

## Correct Workflow for "Create Kafka Dashboards"

### 1. Discovery Phase
```
User: "Create three dashboards for kafka - msk, confluent and ohi (nri-kafka)"

System Actions:
1. Run discovery.list_schemas with filter="kafka"
2. Discover actual event types in NRDB:
   - Found: KafkaBrokerSample, KafkaTopicSample, KafkaConsumerSample
   - NOT Found: AWSKafkaBroker, ConfluentKafka (these were assumptions!)
3. For each discovered schema, run discovery.profile_attribute to understand:
   - Available attributes
   - Data types
   - Value distributions
   - Cardinality
```

### 2. Analysis Phase
```
Based on discovery:
- KafkaBrokerSample has attributes: broker.messagesInPerSecond, broker.IOInPerSecond, etc.
- No MSK-specific or Confluent-specific event types found
- All Kafka data comes through nri-kafka integration

Decision: Create one comprehensive dashboard using discovered data rather than three assumed dashboards
```

### 3. Dashboard Creation Phase
```
Create dashboard using ONLY discovered event types and attributes:
- Title: "Kafka Monitoring Dashboard" (based on discovered data)
- Widgets use actual discovered queries:
  - SELECT average(broker.messagesInPerSecond) FROM KafkaBrokerSample
  - SELECT sum(topic.partitionCount) FROM KafkaTopicSample
  - SELECT max(consumer.lag) FROM KafkaConsumerSample
```

## What Was Wrong Before

### ❌ INCORRECT Approach:
```go
// Hardcoded assumptions about event types
dashboard := map[string]interface{}{
    "widgets": []map[string]interface{}{
        {
            "query": "SELECT average(cpuUtilization) FROM AWSKafkaBroker", // ASSUMED!
        },
        {
            "query": "SELECT rate(sum(messagesProduced)) FROM AWSKafkaBroker", // ASSUMED!
        },
    },
}
```

### ✅ CORRECT Approach:
```go
// First discover what's actually available
schemas, _ := discoveryEngine.ListSchemas(ctx, "kafka")

// Build queries based on discovered data
for _, schema := range schemas {
    attributes, _ := discoveryEngine.ProfileSchema(ctx, schema.Name)
    // Generate appropriate widgets based on actual attributes
}
```

## Implementation Requirements

### 1. Enhanced Discovery Tools
- `discover_event_types`: Find all event types matching a pattern
- `discover_attributes`: Get all attributes for an event type
- `analyze_data_patterns`: Understand data characteristics
- `suggest_visualizations`: Recommend widget types based on data

### 2. Smart Dashboard Generator
- Never hardcode event types or attributes
- Build queries dynamically based on discovery
- Adapt to what's actually in the customer's data
- Provide meaningful dashboards even with limited data

### 3. User Feedback Loop
```
User: "Create Kafka dashboards"
System: "I discovered the following Kafka data in your account:
- KafkaBrokerSample with 15 metrics
- KafkaTopicSample with 8 metrics
- No MSK or Confluent-specific data found

Would you like me to create a comprehensive Kafka dashboard using the available data?"
```

## Example: Correct Implementation

```python
async def create_kafka_dashboards(self, request):
    # Step 1: Discover what's actually available
    kafka_schemas = await self.discovery.list_schemas(filter="kafka")
    
    if not kafka_schemas:
        return {
            "error": "No Kafka-related data found in your NRDB",
            "suggestion": "Please ensure Kafka integration is configured"
        }
    
    dashboards = []
    
    # Step 2: Analyze each schema
    for schema in kafka_schemas:
        attributes = await self.discovery.profile_schema(schema.name)
        
        # Step 3: Build dashboard based on discovered data
        widgets = self.generate_widgets_from_attributes(schema, attributes)
        
        dashboard = {
            "name": f"{schema.name} Monitoring",
            "pages": [{
                "name": "Overview",
                "widgets": widgets
            }]
        }
        
        # Step 4: Create dashboard with actual data
        created = await self.create_dashboard(dashboard)
        dashboards.append(created)
    
    return {
        "created": len(dashboards),
        "dashboards": dashboards,
        "based_on": [s.name for s in kafka_schemas]
    }
```

## Benefits of Discovery-Based Approach

1. **Adaptability**: Works with any customer's data setup
2. **Accuracy**: Only creates visualizations for data that exists
3. **Reliability**: No failed queries due to missing event types
4. **Intelligence**: Can suggest best visualizations based on actual data
5. **Transparency**: Shows users what data was used

## Conclusion

The system must ALWAYS:
1. Discover first
2. Analyze what's found
3. Build based on reality
4. Never assume data structures

This is the fundamental design principle that makes the system truly universal and valuable.
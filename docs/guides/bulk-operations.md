# Bulk Operations Guide

The New Relic MCP Server provides powerful bulk operation tools for managing multiple resources efficiently. This guide covers all available bulk operations and best practices.

## Overview

Bulk operations allow you to:
- Apply changes to multiple entities at once
- Execute multiple queries in parallel
- Migrate resources between accounts
- Delete multiple resources with safety checks

## Available Bulk Operations

### 1. Bulk Tag Entities
Apply tags to multiple entities in a single operation.

```json
{
  "tool": "bulk_tag_entities",
  "arguments": {
    "entity_guids": ["ABC123", "DEF456", "GHI789"],
    "tags": ["environment:production", "team:backend"],
    "operation": "add"  // or "replace"
  }
}
```

**Features:**
- Batch processing (50 entities per batch)
- Support for add/replace operations
- Detailed success/failure reporting

### 2. Bulk Create Monitors
Create multiple synthetic monitors from a template.

```json
{
  "tool": "bulk_create_monitors",
  "arguments": {
    "monitors": [
      {
        "name": "API Health Check - Service A",
        "url": "https://api.example.com/health",
        "frequency": 5,
        "locations": ["US_EAST_1", "EU_WEST_1"]
      },
      {
        "name": "API Health Check - Service B",
        "url": "https://api.example.com/health",
        "frequency": 10,
        "locations": ["US_WEST_1"]
      }
    ],
    "template": {
      "type": "SIMPLE",
      "status": "ENABLED",
      "tags": ["bulk-created", "api-monitoring"]
    }
  }
}
```

**Features:**
- Template-based creation
- Common settings applied to all monitors
- Individual customization per monitor

### 3. Bulk Update Dashboards
Update multiple dashboards with common changes.

```json
{
  "tool": "bulk_update_dashboards",
  "arguments": {
    "dashboard_ids": ["dash1", "dash2", "dash3"],
    "updates": {
      "name": "Updated Dashboard Name",
      "description": "Bulk updated description",
      "permissions": "PUBLIC_READ_ONLY"
    }
  }
}
```

**Features:**
- Update name, description, permissions
- Add/remove tags
- Add widgets to multiple dashboards

### 4. Bulk Delete Entities
Delete multiple entities with safety checks.

```json
{
  "tool": "bulk_delete_entities",
  "arguments": {
    "entity_type": "monitor",  // or "dashboard", "alert_condition"
    "entity_ids": ["id1", "id2", "id3"],
    "force": false  // Set to true for > 10 entities
  }
}
```

**Features:**
- Safety check for large deletions
- Support for different entity types
- Detailed deletion results

### 5. Bulk Execute Queries
Execute multiple NRQL queries in parallel.

```json
{
  "tool": "bulk_execute_queries",
  "arguments": {
    "queries": [
      {
        "name": "Transaction Count",
        "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
      },
      {
        "name": "Error Rate",
        "query": "SELECT percentage(count(*), WHERE error IS true) FROM Transaction SINCE 1 hour ago"
      }
    ],
    "parallel": true,
    "timeout": 30
  }
}
```

**Features:**
- Parallel or sequential execution
- Per-query timeout control
- Named queries for easy identification

### 6. Bulk Dashboard Migrate
Migrate dashboards between accounts or update to new standards.

```json
{
  "tool": "bulk_dashboard_migrate",
  "arguments": {
    "dashboard_ids": ["dash1", "dash2"],
    "target_account_id": "2345678",
    "update_queries": true,
    "preserve_permissions": false
  }
}
```

**Features:**
- Cross-account migration
- Query updating for target account
- Permission preservation options

## Error Handling

All bulk operations use structured error handling:

```json
{
  "error": {
    "code": -32602,
    "message": "Validation failed for field 'entity_guids[2]'",
    "data": {
      "type": "validation_error",
      "field": "entity_guids[2]",
      "message": "must be a non-empty string",
      "hint": "Ensure all entity GUIDs are valid strings"
    }
  }
}
```

### Common Error Types

1. **Invalid Parameters**
   - Missing required fields
   - Wrong data types
   - Empty arrays

2. **Validation Errors**
   - Invalid entity types
   - Malformed queries
   - Safety check failures

3. **Execution Errors**
   - Network timeouts
   - Permission denied
   - Resource not found

## Best Practices

### 1. Batch Size Management
- Keep batches reasonable (50-100 items)
- Use pagination for large datasets
- Monitor execution time

### 2. Error Recovery
```javascript
// Example: Retry failed operations
const results = await bulkOperation();
const failed = results.results.filter(r => r.status === 'failed');

if (failed.length > 0) {
  // Retry failed operations with exponential backoff
  await retryFailedOperations(failed);
}
```

### 3. Progress Monitoring
```javascript
// Example: Track progress for large operations
const totalItems = 1000;
const batchSize = 50;

for (let i = 0; i < totalItems; i += batchSize) {
  const batch = items.slice(i, i + batchSize);
  const result = await processBatch(batch);
  
  console.log(`Progress: ${i + batch.length}/${totalItems}`);
  console.log(`Success: ${result.success}, Failed: ${result.failed}`);
}
```

### 4. Safety Checks
Always implement safety checks for destructive operations:

```javascript
// Example: Confirm before bulk delete
const itemCount = entityIds.length;
if (itemCount > 10) {
  const confirmed = await confirm(`Delete ${itemCount} entities?`);
  if (!confirmed) return;
  
  // Set force=true for bulk delete
  params.force = true;
}
```

## Mock Mode Support

All bulk operations support mock mode for development and testing:

```bash
# Enable mock mode
export NEW_RELIC_MOCK_MODE=true

# Or in .env file
NEW_RELIC_MOCK_MODE=true
```

Mock mode returns realistic responses without making actual API calls.

## Performance Considerations

### 1. Parallel Execution
- Use parallel mode for independent operations
- Sequential mode for dependent operations
- Adjust concurrency based on API limits

### 2. Timeout Configuration
- Default: 30 seconds per operation
- Increase for complex queries
- Use operation-specific timeouts

### 3. Resource Usage
- Monitor memory usage for large result sets
- Use streaming for continuous operations
- Implement pagination for large datasets

## Examples

### Example 1: Bulk Tagging by Environment
```javascript
// Tag all production services
const productionServices = await findEntitiesByTag('service:*');
await bulkTagEntities({
  entity_guids: productionServices.map(e => e.guid),
  tags: ['environment:production', 'sla:99.9'],
  operation: 'add'
});
```

### Example 2: Dashboard Migration Workflow
```javascript
// Migrate all team dashboards to new account
const teamDashboards = await listDashboards({ tag: 'team:frontend' });
const migration = await bulkDashboardMigrate({
  dashboard_ids: teamDashboards.map(d => d.id),
  target_account_id: newAccountId,
  update_queries: true,
  preserve_permissions: true
});

console.log(`Migrated ${migration.summary.successful} dashboards`);
```

### Example 3: Parallel Query Execution
```javascript
// Execute performance queries across services
const services = ['api', 'web', 'mobile'];
const queries = services.map(service => ({
  name: `${service}_performance`,
  query: `SELECT average(duration) FROM Transaction WHERE appName = '${service}' SINCE 1 hour ago`
}));

const results = await bulkExecuteQueries({
  queries,
  parallel: true,
  timeout: 60
});
```

## Troubleshooting

### Common Issues

1. **Rate Limiting**
   - Reduce batch size
   - Add delays between batches
   - Use exponential backoff

2. **Timeouts**
   - Increase timeout values
   - Reduce query complexity
   - Use smaller time ranges

3. **Permission Errors**
   - Verify API key permissions
   - Check account access
   - Ensure resource ownership

### Debug Mode
Enable debug logging for detailed operation tracking:

```bash
export LOG_LEVEL=DEBUG
```

## Future Enhancements

- Bulk alert policy management
- Bulk entity relationship updates
- Scheduled bulk operations
- Bulk operation templates
- Progress webhooks
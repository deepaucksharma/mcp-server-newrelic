package discovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/framework"
)

// EventTypeDiscoveryE2ESuite tests event type discovery with real data
type EventTypeDiscoveryE2ESuite struct {
	suite.Suite
	client *framework.MCPTestClient
	accounts map[string]*framework.TestAccount
}

func (s *EventTypeDiscoveryE2ESuite) SetupSuite() {
	// Initialize test accounts
	s.accounts = framework.LoadTestAccounts()
	s.Require().NotEmpty(s.accounts, "No test accounts configured")
	
	// Create MCP client with primary account
	s.client = framework.NewMCPTestClient(s.accounts["primary"])
}

func (s *EventTypeDiscoveryE2ESuite) TearDownSuite() {
	s.client.Close()
}

// TestBasicEventTypeDiscovery validates basic event type discovery
func (s *EventTypeDiscoveryE2ESuite) TestBasicEventTypeDiscovery() {
	ctx := context.Background()
	
	// Execute discovery without any assumptions
	result, err := s.client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
		"time_range": "24 hours",
		"include_samples": true,
		"min_event_count": 10,
	})
	
	s.Require().NoError(err, "Event type discovery should not error")
	s.Require().NotNil(result, "Should return discovery results")
	
	// Validate structure without assuming specific event types
	eventTypes, ok := result["event_types"].([]interface{})
	s.Require().True(ok, "Should return event_types array")
	s.Require().NotEmpty(eventTypes, "Should discover at least one event type")
	
	// Validate each discovered event type
	for i, et := range eventTypes {
		eventType, ok := et.(map[string]interface{})
		s.Require().True(ok, "Event type %d should be a map", i)
		
		// Required fields
		s.Contains(eventType, "name", "Event type should have name")
		s.Contains(eventType, "count", "Event type should have count")
		s.Contains(eventType, "first_seen", "Event type should have first_seen")
		s.Contains(eventType, "last_seen", "Event type should have last_seen")
		
		// Validate count is positive
		count, ok := eventType["count"].(float64)
		s.True(ok && count > 0, "Count should be positive number")
		
		// If samples requested, validate sample structure
		if samples, hasSamples := eventType["samples"]; hasSamples {
			s.validateSampleStructure(samples)
		}
		
		// Store discovered type for later tests
		s.client.StoreDiscovery(fmt.Sprintf("event_type_%s", eventType["name"]), eventType)
	}
	
	// Validate metadata
	s.validateDiscoveryMetadata(result)
}

// TestEventTypeDiscoveryWithEmptyAccount tests behavior with no data
func (s *EventTypeDiscoveryE2ESuite) TestEventTypeDiscoveryWithEmptyAccount() {
	ctx := context.Background()
	
	// Switch to empty account
	emptyClient := framework.NewMCPTestClient(s.accounts["empty"])
	defer emptyClient.Close()
	
	result, err := emptyClient.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
		"time_range": "7 days",
	})
	
	s.Require().NoError(err, "Should handle empty account gracefully")
	
	// Should return empty results, not error
	eventTypes, ok := result["event_types"].([]interface{})
	s.True(ok, "Should return event_types array")
	s.Empty(eventTypes, "Empty account should have no event types")
	
	// Should provide helpful guidance
	s.Contains(result, "guidance", "Should provide guidance for empty account")
	guidance, ok := result["guidance"].(map[string]interface{})
	s.True(ok, "Guidance should be a map")
	s.Contains(guidance, "next_steps", "Should suggest next steps")
}

// TestEventTypeDiscoveryTimeRanges tests different time range behaviors
func (s *EventTypeDiscoveryE2ESuite) TestEventTypeDiscoveryTimeRanges() {
	ctx := context.Background()
	
	testCases := []struct {
		name      string
		timeRange string
		validate  func(*EventTypeDiscoveryE2ESuite, map[string]interface{})
	}{
		{
			name:      "Last hour",
			timeRange: "1 hour",
			validate: func(s *EventTypeDiscoveryE2ESuite, result map[string]interface{}) {
				// Recent data should have fresh timestamps
				s.validateDataFreshness(result, time.Hour)
			},
		},
		{
			name:      "Last 24 hours",
			timeRange: "24 hours",
			validate: func(s *EventTypeDiscoveryE2ESuite, result map[string]interface{}) {
				// Should have more event types than 1 hour
				s.True(len(s.getEventTypes(result)) > 0, "Should have event types")
			},
		},
		{
			name:      "Last 30 days",
			timeRange: "30 days",
			validate: func(s *EventTypeDiscoveryE2ESuite, result map[string]interface{}) {
				// Should potentially discover historical event types
				eventTypes := s.getEventTypes(result)
				s.validateHistoricalCoverage(eventTypes)
			},
		},
	}
	
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := s.client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
				"time_range": tc.timeRange,
			})
			
			s.Require().NoError(err)
			tc.validate(s, result)
		})
	}
}

// TestEventTypeDiscoveryAcrossAccounts tests multi-account discovery
func (s *EventTypeDiscoveryE2ESuite) TestEventTypeDiscoveryAcrossAccounts() {
	ctx := context.Background()
	
	accountResults := make(map[string][]string)
	
	// Discover event types in each account
	for accountName, account := range s.accounts {
		if accountName == "empty" {
			continue // Skip empty account
		}
		
		result, err := s.client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
			"time_range": "24 hours",
			"account_id": account.AccountID,
		})
		
		s.Require().NoError(err, "Discovery should work for account %s", accountName)
		
		// Collect event type names
		eventTypes := s.getEventTypes(result)
		names := make([]string, 0, len(eventTypes))
		for _, et := range eventTypes {
			if name, ok := et["name"].(string); ok {
				names = append(names, name)
			}
		}
		accountResults[accountName] = names
		
		// Each account should have some event types
		s.NotEmpty(names, "Account %s should have event types", accountName)
	}
	
	// Validate accounts have different schemas (no assumptions)
	s.validateAccountDiversity(accountResults)
}

// TestEventTypeMetadataDiscovery tests detailed metadata discovery
func (s *EventTypeDiscoveryE2ESuite) TestEventTypeMetadataDiscovery() {
	ctx := context.Background()
	
	// First, discover what event types exist
	result, err := s.client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
		"time_range": "24 hours",
		"include_metadata": true,
		"include_samples": false, // Focus on metadata
	})
	
	s.Require().NoError(err)
	
	eventTypes := s.getEventTypes(result)
	s.Require().NotEmpty(eventTypes, "Should discover event types")
	
	// For each discovered type, validate metadata
	for _, et := range eventTypes {
		name := et["name"].(string)
		
		// Should have extended metadata when requested
		s.Contains(et, "metadata", "Event type %s should have metadata", name)
		
		metadata, ok := et["metadata"].(map[string]interface{})
		s.True(ok, "Metadata should be a map for %s", name)
		
		// Validate metadata contents (adaptive - don't assume structure)
		s.validateEventTypeMetadata(name, metadata)
	}
}

// Helper methods

func (s *EventTypeDiscoveryE2ESuite) validateSampleStructure(samples interface{}) {
	sampleList, ok := samples.([]interface{})
	s.Require().True(ok, "Samples should be an array")
	s.NotEmpty(sampleList, "Should have at least one sample")
	
	for i, sample := range sampleList {
		sampleMap, ok := sample.(map[string]interface{})
		s.True(ok, "Sample %d should be a map", i)
		
		// Don't assume specific fields - validate it has some attributes
		s.NotEmpty(sampleMap, "Sample should have attributes")
		
		// Should have timestamp
		s.Contains(sampleMap, "timestamp", "Sample should have timestamp")
	}
}

func (s *EventTypeDiscoveryE2ESuite) validateDiscoveryMetadata(result map[string]interface{}) {
	s.Contains(result, "metadata", "Should have discovery metadata")
	
	metadata, ok := result["metadata"].(map[string]interface{})
	s.Require().True(ok, "Metadata should be a map")
	
	// Required metadata fields
	s.Contains(metadata, "discovery_time", "Should have discovery timestamp")
	s.Contains(metadata, "account_id", "Should identify account")
	s.Contains(metadata, "time_range_used", "Should show actual time range")
}

func (s *EventTypeDiscoveryE2ESuite) getEventTypes(result map[string]interface{}) []map[string]interface{} {
	if eventTypesRaw, ok := result["event_types"].([]interface{}); ok {
		eventTypes := make([]map[string]interface{}, 0, len(eventTypesRaw))
		for _, et := range eventTypesRaw {
			if eventType, ok := et.(map[string]interface{}); ok {
				eventTypes = append(eventTypes, eventType)
			}
		}
		return eventTypes
	}
	return nil
}

func (s *EventTypeDiscoveryE2ESuite) validateDataFreshness(result map[string]interface{}, maxAge time.Duration) {
	eventTypes := s.getEventTypes(result)
	s.Require().NotEmpty(eventTypes, "Should have event types to validate freshness")
	
	now := time.Now()
	for _, et := range eventTypes {
		if lastSeenStr, ok := et["last_seen"].(string); ok {
			lastSeen, err := time.Parse(time.RFC3339, lastSeenStr)
			s.NoError(err, "Should parse last_seen timestamp")
			
			age := now.Sub(lastSeen)
			s.True(age <= maxAge+time.Minute, // Allow 1 minute buffer
				"Event type %s last seen %v ago, should be within %v",
				et["name"], age, maxAge)
		}
	}
}

func (s *EventTypeDiscoveryE2ESuite) validateHistoricalCoverage(eventTypes []map[string]interface{}) {
	// Don't assume what historical types exist
	// Just validate that we can see older data if it exists
	
	oldestSeen := time.Now()
	hasHistorical := false
	
	for _, et := range eventTypes {
		if firstSeenStr, ok := et["first_seen"].(string); ok {
			firstSeen, err := time.Parse(time.RFC3339, firstSeenStr)
			s.NoError(err)
			
			if firstSeen.Before(oldestSeen) {
				oldestSeen = firstSeen
			}
			
			// Check if any data is older than 7 days
			if time.Since(firstSeen) > 7*24*time.Hour {
				hasHistorical = true
			}
		}
	}
	
	// Log findings but don't fail - account might be new
	s.T().Logf("Oldest data seen: %v ago", time.Since(oldestSeen))
	s.T().Logf("Has historical data (>7 days): %v", hasHistorical)
}

func (s *EventTypeDiscoveryE2ESuite) validateAccountDiversity(accountResults map[string][]string) {
	// Don't assume accounts must have different schemas
	// Just validate that we can discover from each account independently
	
	uniqueTypes := make(map[string][]string)
	commonTypes := make(map[string]int)
	
	// Analyze type distribution
	for account, types := range accountResults {
		for _, typeName := range types {
			commonTypes[typeName]++
			if commonTypes[typeName] == 1 {
				uniqueTypes[typeName] = []string{account}
			} else {
				uniqueTypes[typeName] = append(uniqueTypes[typeName], account)
			}
		}
	}
	
	// Log findings for analysis
	s.T().Logf("Account diversity analysis:")
	s.T().Logf("- Total unique event types: %d", len(commonTypes))
	s.T().Logf("- Types present in all accounts: %d", s.countTypesInAllAccounts(commonTypes, len(accountResults)))
	
	for typeName, accounts := range uniqueTypes {
		if len(accounts) == 1 {
			s.T().Logf("- Type '%s' unique to account: %s", typeName, accounts[0])
		}
	}
}

func (s *EventTypeDiscoveryE2ESuite) countTypesInAllAccounts(commonTypes map[string]int, accountCount int) int {
	count := 0
	for _, occurrences := range commonTypes {
		if occurrences == accountCount {
			count++
		}
	}
	return count
}

func (s *EventTypeDiscoveryE2ESuite) validateEventTypeMetadata(name string, metadata map[string]interface{}) {
	// Adaptive validation - check what metadata is provided
	// Don't assume specific fields
	
	// Common metadata fields to check for (but not require)
	possibleFields := []string{
		"attribute_count",
		"key_attributes",
		"data_sources",
		"related_entities",
		"schema_version",
		"custom_attributes",
	}
	
	foundFields := []string{}
	for _, field := range possibleFields {
		if _, exists := metadata[field]; exists {
			foundFields = append(foundFields, field)
		}
	}
	
	// Should have some metadata
	s.NotEmpty(foundFields, "Event type %s should have some metadata fields", name)
	s.T().Logf("Event type %s has metadata fields: %v", name, foundFields)
}

func TestEventTypeDiscoveryE2E(t *testing.T) {
	suite.Run(t, new(EventTypeDiscoveryE2ESuite))
}
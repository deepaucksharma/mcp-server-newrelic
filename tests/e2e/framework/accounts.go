package framework

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// TestAccount represents a New Relic account for testing
type TestAccount struct {
	Name      string
	APIKey    string
	AccountID string
	Region    string
	Purpose   string
}

// LoadTestAccounts loads test accounts from environment variables
func LoadTestAccounts() map[string]*TestAccount {
	accounts := make(map[string]*TestAccount)
	
	// Define account configurations
	accountConfigs := []struct {
		name     string
		envKey   string
		envID    string
		purpose  string
	}{
		{
			name:    "primary",
			envKey:  "NEW_RELIC_API_KEY_PRIMARY",
			envID:   "NEW_RELIC_ACCOUNT_ID_PRIMARY",
			purpose: "Main testing account with diverse data",
		},
		{
			name:    "secondary",
			envKey:  "NEW_RELIC_API_KEY_SECONDARY",
			envID:   "NEW_RELIC_ACCOUNT_ID_SECONDARY",
			purpose: "Cross-account testing with different schemas",
		},
		{
			name:    "empty",
			envKey:  "NEW_RELIC_API_KEY_EMPTY",
			envID:   "NEW_RELIC_ACCOUNT_ID_EMPTY",
			purpose: "Zero-data scenarios testing",
		},
		{
			name:    "high_cardinality",
			envKey:  "NEW_RELIC_API_KEY_HIGH_CARD",
			envID:   "NEW_RELIC_ACCOUNT_ID_HIGH_CARD",
			purpose: "Performance and scale testing",
		},
	}
	
	// Load each account
	for _, config := range accountConfigs {
		apiKey := os.Getenv(config.envKey)
		accountID := os.Getenv(config.envID)
		
		if apiKey != "" && accountID != "" {
			account := &TestAccount{
				Name:      config.name,
				APIKey:    apiKey,
				AccountID: accountID,
				Region:    detectRegion(apiKey),
				Purpose:   config.purpose,
			}
			accounts[config.name] = account
		}
	}
	
	// Ensure we have at least the primary account
	if _, exists := accounts["primary"]; !exists {
		// Try loading from default env vars
		apiKey := os.Getenv("NEW_RELIC_API_KEY")
		accountID := os.Getenv("NEW_RELIC_ACCOUNT_ID")
		
		if apiKey != "" && accountID != "" {
			accounts["primary"] = &TestAccount{
				Name:      "primary",
				APIKey:    apiKey,
				AccountID: accountID,
				Region:    detectRegion(apiKey),
				Purpose:   "Default test account",
			}
		}
	}
	
	return accounts
}

// GetTestAccount retrieves a specific test account
func GetTestAccount(name string) (*TestAccount, error) {
	accounts := LoadTestAccounts()
	account, exists := accounts[name]
	if !exists {
		return nil, fmt.Errorf("test account %q not found", name)
	}
	return account, nil
}

// GetRequiredTestAccount retrieves a test account or panics
func GetRequiredTestAccount(name string) *TestAccount {
	account, err := GetTestAccount(name)
	if err != nil {
		panic(fmt.Sprintf("Required test account %q not configured: %v", name, err))
	}
	return account
}

// detectRegion attempts to detect the region from the API key format
func detectRegion(apiKey string) string {
	// EU keys often have specific patterns
	if strings.Contains(strings.ToLower(apiKey), "eu") {
		return "EU"
	}
	// Default to US
	return "US"
}

// ValidateTestEnvironment checks if the test environment is properly configured
func ValidateTestEnvironment() error {
	accounts := LoadTestAccounts()
	
	if len(accounts) == 0 {
		return fmt.Errorf("no test accounts configured")
	}
	
	// Check for primary account
	if _, exists := accounts["primary"]; !exists {
		return fmt.Errorf("primary test account not configured")
	}
	
	// Warn about missing optional accounts
	optional := []string{"secondary", "empty", "high_cardinality"}
	missing := []string{}
	
	for _, name := range optional {
		if _, exists := accounts[name]; !exists {
			missing = append(missing, name)
		}
	}
	
	if len(missing) > 0 {
		fmt.Printf("Warning: Optional test accounts not configured: %v\n", missing)
		fmt.Println("Some tests may be skipped.")
	}
	
	return nil
}

// AccountValidator provides methods to validate account characteristics
type AccountValidator struct {
	client *MCPTestClient
}

// NewAccountValidator creates a validator for the given account
func NewAccountValidator(account *TestAccount) *AccountValidator {
	return &AccountValidator{
		client: NewMCPTestClient(account),
	}
}

// ValidateHasData checks if the account has any data
func (v *AccountValidator) ValidateHasData() error {
	ctx := context.Background()
	
	result, err := v.client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
		"time_range": "24 hours",
		"min_event_count": 1,
	})
	
	if err != nil {
		return fmt.Errorf("failed to discover event types: %w", err)
	}
	
	// Type assert result first
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}
	
	eventTypes, ok := resultMap["event_types"].([]interface{})
	if !ok || len(eventTypes) == 0 {
		return fmt.Errorf("account has no event types")
	}
	
	return nil
}

// ValidateIsEmpty checks if the account has no data
func (v *AccountValidator) ValidateIsEmpty() error {
	ctx := context.Background()
	
	result, err := v.client.ExecuteTool(ctx, "discovery.explore_event_types", map[string]interface{}{
		"time_range": "7 days",
	})
	
	if err != nil {
		return fmt.Errorf("failed to discover event types: %w", err)
	}
	
	// Type assert result first
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}
	
	eventTypes, ok := resultMap["event_types"].([]interface{})
	if ok && len(eventTypes) > 0 {
		return fmt.Errorf("account has %d event types, expected 0", len(eventTypes))
	}
	
	return nil
}

// ValidateHighCardinality checks if the account has high cardinality data
func (v *AccountValidator) ValidateHighCardinality() error {
	ctx := context.Background()
	
	// First discover what service attribute exists
	serviceAttr, err := v.discoverServiceAttribute(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover service attribute: %w", err)
	}
	
	// Count unique services
	query := fmt.Sprintf("SELECT uniqueCount(%s) FROM Transaction SINCE 1 hour ago", serviceAttr)
	
	result, err := v.client.ExecuteTool(ctx, "nrql.execute", map[string]interface{}{
		"query": query,
	})
	
	if err != nil {
		return fmt.Errorf("failed to count services: %w", err)
	}
	
	// Type assert result first
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}
	
	// Extract count from results
	if results, ok := resultMap["results"].([]interface{}); ok && len(results) > 0 {
		if firstResult, ok := results[0].(map[string]interface{}); ok {
			if count, ok := firstResult["uniqueCount"].(float64); ok {
				if count < 100 {
					return fmt.Errorf("account has only %.0f unique services, expected 100+", count)
				}
				return nil
			}
		}
	}
	
	return fmt.Errorf("could not determine service cardinality")
}

func (v *AccountValidator) discoverServiceAttribute(ctx context.Context) (string, error) {
	// Common service attribute names to check
	candidates := []string{"appName", "service.name", "applicationName"}
	
	for _, attr := range candidates {
		result, err := v.client.ExecuteTool(ctx, "discovery.profile_attribute", map[string]interface{}{
			"event_type": "Transaction",
			"attribute":  attr,
		})
		
		if err == nil {
			// Type assert result first
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				continue
			}
			
			if profile, ok := resultMap["profile"].(map[string]interface{}); ok {
				if coverage, ok := profile["coverage"].(float64); ok && coverage > 50 {
					return attr, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("no service attribute found with >50%% coverage")
}
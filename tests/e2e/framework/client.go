package framework

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// TestClient provides methods for interacting with New Relic APIs
type TestClient struct {
	httpClient *http.Client
	apiKey     string
	accountID  string
	region     string
	baseURL    string
}

// NewTestClient creates a new test client
func NewTestClient(ctx context.Context) (*TestClient, error) {
	apiKey := os.Getenv("NEW_RELIC_API_KEY_PRIMARY")
	if apiKey == "" {
		apiKey = os.Getenv("E2E_PRIMARY_API_KEY")
	}
	
	accountID := os.Getenv("NEW_RELIC_ACCOUNT_ID_PRIMARY")
	if accountID == "" {
		accountID = os.Getenv("E2E_PRIMARY_ACCOUNT_ID")
	}
	
	if apiKey == "" || accountID == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}
	
	region := os.Getenv("NEW_RELIC_REGION_PRIMARY")
	if region == "" {
		region = "US"
	}
	
	baseURL := "https://api.newrelic.com"
	if region == "EU" {
		baseURL = "https://api.eu.newrelic.com"
	}
	
	return &TestClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:    apiKey,
		accountID: accountID,
		region:    region,
		baseURL:   baseURL,
	}, nil
}

// ExecuteNRQL executes a NRQL query
func (c *TestClient) ExecuteNRQL(ctx context.Context, query, accountID string) (map[string]interface{}, error) {
	if accountID == "" {
		accountID = c.accountID
	}
	
	payload := map[string]interface{}{
		"query": fmt.Sprintf(`{
			actor {
				account(id: %s) {
					nrql(query: "%s") {
						results
						totalResult
						metadata {
							eventTypes
							messages
							timeWindow {
								begin
								end
							}
						}
					}
				}
			}
		}`, accountID, query),
	}
	
	return c.graphQLRequest(ctx, payload)
}

// DiscoverEventTypes discovers available event types
func (c *TestClient) DiscoverEventTypes(ctx context.Context, limit int) ([]string, error) {
	// SHOW EVENT TYPES doesn't support LIMIT directly
	query := "SHOW EVENT TYPES"
	result, err := c.ExecuteNRQL(ctx, query, c.accountID)
	if err != nil {
		return nil, err
	}
	
	// Parse the response to extract event types
	eventTypes := []string{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if actor, ok := data["actor"].(map[string]interface{}); ok {
			if account, ok := actor["account"].(map[string]interface{}); ok {
				if nrql, ok := account["nrql"].(map[string]interface{}); ok {
					if results, ok := nrql["results"].([]interface{}); ok {
						for _, r := range results {
							if m, ok := r.(map[string]interface{}); ok {
								if et, ok := m["eventType"].(string); ok {
									eventTypes = append(eventTypes, et)
								}
							}
						}
					}
				}
			}
		}
	}
	
	return eventTypes, nil
}

// DiscoverAttributes discovers attributes for an event type
func (c *TestClient) DiscoverAttributes(ctx context.Context, eventType string, limit int) ([]string, error) {
	query := fmt.Sprintf("SELECT keyset() FROM %s LIMIT %d", eventType, limit)
	result, err := c.ExecuteNRQL(ctx, query, c.accountID)
	if err != nil {
		return nil, err
	}
	
	// Parse the response to extract attributes
	attributes := []string{}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if actor, ok := data["actor"].(map[string]interface{}); ok {
			if account, ok := actor["account"].(map[string]interface{}); ok {
				if nrql, ok := account["nrql"].(map[string]interface{}); ok {
					if results, ok := nrql["results"].([]interface{}); ok {
						for _, r := range results {
							if m, ok := r.(map[string]interface{}); ok {
								if keys, ok := m["keyset"].([]interface{}); ok {
									for _, k := range keys {
										if attr, ok := k.(string); ok {
											attributes = append(attributes, attr)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	return attributes, nil
}

// graphQLRequest executes a GraphQL request
func (c *TestClient) graphQLRequest(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/graphql", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	// Check for GraphQL errors
	if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", errors)
	}
	
	return result, nil
}
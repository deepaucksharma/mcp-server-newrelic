package newrelic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client provides access to New Relic APIs
type Client struct {
	apiKey      string
	accountID   string
	region      string
	httpClient  *http.Client
	graphQLURL  string
	restAPIURL  string
}

// Config holds client configuration
type Config struct {
	APIKey     string
	AccountID  string
	Region     string // "US" or "EU"
	Timeout    time.Duration
}

// NewClient creates a new New Relic API client
func NewClient(config Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.AccountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	// Set default timeout
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Determine URLs based on region
	graphQLURL := "https://api.newrelic.com/graphql"
	restAPIURL := "https://api.newrelic.com/v2"
	if strings.ToUpper(config.Region) == "EU" {
		graphQLURL = "https://api.eu.newrelic.com/graphql"
		restAPIURL = "https://api.eu.newrelic.com/v2"
	}

	return &Client{
		apiKey:     config.APIKey,
		accountID:  config.AccountID,
		region:     config.Region,
		graphQLURL: graphQLURL,
		restAPIURL: restAPIURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// NRQLResult represents the result of an NRQL query
type NRQLResult struct {
	Results         []map[string]interface{} `json:"results"`
	Metadata        NRQLMetadata             `json:"metadata"`
	PerformanceInfo *PerformanceInfo         `json:"performanceInfo,omitempty"`
}

// NRQLMetadata contains metadata about the query execution
type NRQLMetadata struct {
	EventTypes      []string               `json:"eventTypes"`
	Messages        []string               `json:"messages,omitempty"`
	Facets          []string               `json:"facets,omitempty"`
	Contents        map[string]interface{} `json:"contents,omitempty"`
}

// PerformanceInfo contains performance metrics for the query
type PerformanceInfo struct {
	InspectedCount int64  `json:"inspectedCount"`
	MatchedCount   int64  `json:"matchedCount"`
	WallClockTime  int64  `json:"wallClockTime"`
}

// QueryNRQL executes an NRQL query
func (c *Client) QueryNRQL(ctx context.Context, query string) (*NRQLResult, error) {
	// GraphQL query to execute NRQL
	graphQLQuery := fmt.Sprintf(`{
		actor {
			account(id: %s) {
				nrql(query: "%s") {
					results
					metadata {
						eventTypes
						messages
						facets
					}
					performanceInfo {
						inspectedCount
						matchedCount
						wallClockTime
					}
				}
			}
		}
	}`, c.accountID, escapeGraphQLString(query))

	// Create request
	reqBody := map[string]interface{}{
		"query": graphQLQuery,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.graphQLURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-Key", c.apiKey)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse GraphQL response
	var graphQLResp struct {
		Data struct {
			Actor struct {
				Account struct {
					NRQL NRQLResult `json:"nrql"`
				} `json:"account"`
			} `json:"actor"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &graphQLResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for GraphQL errors
	if len(graphQLResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	return &graphQLResp.Data.Actor.Account.NRQL, nil
}

// Dashboard represents a New Relic dashboard
type Dashboard struct {
	ID          string                 `json:"id"`
	GUID        string                 `json:"guid"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Pages       []DashboardPage        `json:"pages"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DashboardPage represents a page in a dashboard
type DashboardPage struct {
	Name    string            `json:"name"`
	GUID    string            `json:"guid"`
	Widgets []DashboardWidget `json:"widgets"`
}

// DashboardWidget represents a widget on a dashboard page
type DashboardWidget struct {
	Title         string                 `json:"title"`
	Type          string                 `json:"visualization"`
	Configuration map[string]interface{} `json:"configuration"`
	Query         string                 `json:"rawConfiguration.nrqlQueries[0].query"`
}

// ListDashboards retrieves all dashboards
func (c *Client) ListDashboards(ctx context.Context) ([]Dashboard, error) {
	// TODO: Implement dashboard listing via GraphQL
	// This would use the dashboardSearch query
	return []Dashboard{}, nil
}

// SearchDashboards searches for dashboards containing specific terms
func (c *Client) SearchDashboards(ctx context.Context, searchTerm string) ([]Dashboard, error) {
	// TODO: Implement dashboard search via GraphQL
	return []Dashboard{}, nil
}

// CreateDashboard creates a new dashboard
func (c *Client) CreateDashboard(ctx context.Context, dashboard Dashboard) (*Dashboard, error) {
	// TODO: Implement dashboard creation via GraphQL mutation
	return &dashboard, nil
}

// AlertCondition represents an alert condition
type AlertCondition struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Query             string                 `json:"nrql.query"`
	Enabled           bool                   `json:"enabled"`
	PolicyID          string                 `json:"policyId"`
	Threshold         float64                `json:"terms[0].threshold"`
	ThresholdDuration int                    `json:"terms[0].thresholdDuration"`
	Comparison        string                 `json:"terms[0].operator"`
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`
}

// ListAlertConditions retrieves alert conditions
func (c *Client) ListAlertConditions(ctx context.Context, policyID string) ([]AlertCondition, error) {
	// TODO: Implement alert listing via GraphQL
	return []AlertCondition{}, nil
}

// CreateAlertCondition creates a new alert condition
func (c *Client) CreateAlertCondition(ctx context.Context, condition AlertCondition) (*AlertCondition, error) {
	// TODO: Implement alert creation via GraphQL mutation
	return &condition, nil
}

// GetAccountInfo retrieves account information
func (c *Client) GetAccountInfo(ctx context.Context) (map[string]interface{}, error) {
	graphQLQuery := fmt.Sprintf(`{
		actor {
			account(id: %s) {
				name
				id
			}
		}
	}`, c.accountID)

	// Execute GraphQL query
	reqBody := map[string]interface{}{
		"query": graphQLQuery,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.graphQLURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Helper function to escape GraphQL strings
func escapeGraphQLString(s string) string {
	// Escape quotes and newlines for GraphQL
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
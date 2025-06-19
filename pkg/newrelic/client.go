package newrelic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

// AccountID returns the account ID as an integer
func (c *Client) AccountID() (int, error) {
	return strconv.Atoi(c.accountID)
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
	AccountID   int                    `json:"accountId"`
	Permissions string                 `json:"permissions"`
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
	// Convert accountID string to int for the query
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}
	
	searchQuery := fmt.Sprintf("accountId = %d AND type = 'DASHBOARD'", accountIDInt)
	query := fmt.Sprintf(`
		query {
			actor {
				entitySearch(query: "%s") {
					results {
						entities {
							... on DashboardEntityOutline {
								guid
								name
								accountId
								createdAt
								updatedAt
								dashboardParentGuid
								permissions
								owner {
									email
									userId
								}
							}
						}
					}
				}
			}
		}
	`, searchQuery)
	
	var result struct {
		Actor struct {
			EntitySearch struct {
				Results struct {
					Entities []struct {
						GUID               string    `json:"guid"`
						Name               string    `json:"name"`
						AccountID          int       `json:"accountId"`
						CreatedAt          time.Time `json:"createdAt"`
						UpdatedAt          time.Time `json:"updatedAt"`
						DashboardParentGUID string   `json:"dashboardParentGuid"`
						Permissions        string    `json:"permissions"`
						Owner              struct {
							Email  string `json:"email"`
							UserID int    `json:"userId"`
						} `json:"owner"`
					} `json:"entities"`
				} `json:"results"`
			} `json:"entitySearch"`
		} `json:"actor"`
	}
	
	if err := c.queryGraphQL(ctx, query, nil, &result); err != nil {
		return nil, fmt.Errorf("list dashboards: %w", err)
	}
	
	dashboards := make([]Dashboard, 0, len(result.Actor.EntitySearch.Results.Entities))
	for _, entity := range result.Actor.EntitySearch.Results.Entities {
		dashboards = append(dashboards, Dashboard{
			ID:          entity.GUID,
			Name:        entity.Name,
			AccountID:   entity.AccountID,
			Permissions: entity.Permissions,
			CreatedAt:   entity.CreatedAt,
			UpdatedAt:   entity.UpdatedAt,
		})
	}
	
	return dashboards, nil
}

// SearchDashboards searches for dashboards containing specific terms
func (c *Client) SearchDashboards(ctx context.Context, searchTerm string) ([]Dashboard, error) {
	query := `
		query($accountId: Int!, $searchQuery: String!) {
			actor {
				entitySearch(query: $searchQuery) {
					results {
						entities {
							... on DashboardEntityOutline {
								guid
								name
								accountId
								createdAt
								updatedAt
								dashboardParentGuid
								permissions
								owner {
									email
									userId
								}
							}
						}
					}
				}
			}
		}
	`
	
	// Build search query
	searchQuery := fmt.Sprintf("accountId = %d AND type = 'DASHBOARD' AND name LIKE '%%%s%%'", 
		c.accountID, searchTerm)
	
	variables := map[string]interface{}{
		"accountId":   c.accountID,
		"searchQuery": searchQuery,
	}
	
	var result struct {
		Actor struct {
			EntitySearch struct {
				Results struct {
					Entities []struct {
						GUID               string    `json:"guid"`
						Name               string    `json:"name"`
						AccountID          int       `json:"accountId"`
						CreatedAt          time.Time `json:"createdAt"`
						UpdatedAt          time.Time `json:"updatedAt"`
						DashboardParentGUID string   `json:"dashboardParentGuid"`
						Permissions        string    `json:"permissions"`
						Owner              struct {
							Email  string `json:"email"`
							UserID int    `json:"userId"`
						} `json:"owner"`
					} `json:"entities"`
				} `json:"results"`
			} `json:"entitySearch"`
		} `json:"actor"`
	}
	
	if err := c.queryGraphQL(ctx, query, variables, &result); err != nil {
		return nil, fmt.Errorf("search dashboards: %w", err)
	}
	
	dashboards := make([]Dashboard, 0, len(result.Actor.EntitySearch.Results.Entities))
	for _, entity := range result.Actor.EntitySearch.Results.Entities {
		dashboards = append(dashboards, Dashboard{
			ID:          entity.GUID,
			Name:        entity.Name,
			AccountID:   entity.AccountID,
			Permissions: entity.Permissions,
			CreatedAt:   entity.CreatedAt,
			UpdatedAt:   entity.UpdatedAt,
		})
	}
	
	return dashboards, nil
}

// CreateDashboard creates a new dashboard
func (c *Client) CreateDashboard(ctx context.Context, dashboard Dashboard) (*Dashboard, error) {
	mutation := `
		mutation($accountId: Int!, $dashboard: DashboardInput!) {
			dashboardCreate(accountId: $accountId, dashboard: $dashboard) {
				entityResult {
					guid
					name
					accountId
					createdAt
					updatedAt
					permissions
				}
				errors {
					description
					type
				}
			}
		}
	`
	
	// Convert accountID to int first
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}
	
	// Convert dashboard to input format
	pages := []map[string]interface{}{}
	for _, page := range dashboard.Pages {
		widgets := []map[string]interface{}{}
		for _, widget := range page.Widgets {
			// Use the new widget converter
			widgetInput, err := ConvertWidgetToGraphQLInput(widget, accountIDInt)
			if err != nil {
				return nil, fmt.Errorf("convert widget '%s': %w", widget.Title, err)
			}
			widgets = append(widgets, widgetInput)
		}
		
		pageInput := map[string]interface{}{
			"name":    page.Name,
			"widgets": widgets,
		}
		pages = append(pages, pageInput)
	}
	
	dashboardInput := map[string]interface{}{
		"name":        dashboard.Name,
		"description": dashboard.Description,
		"permissions": dashboard.Permissions,
		"pages":       pages,
	}
	
	variables := map[string]interface{}{
		"accountId": accountIDInt,
		"dashboard": dashboardInput,
	}
	
	var result struct {
		DashboardCreate struct {
			EntityResult struct {
				GUID        string    `json:"guid"`
				Name        string    `json:"name"`
				AccountID   int       `json:"accountId"`
				CreatedAt   time.Time `json:"createdAt"`
				UpdatedAt   time.Time `json:"updatedAt"`
				Permissions string    `json:"permissions"`
			} `json:"entityResult"`
			Errors []struct {
				Description string `json:"description"`
				Type        string `json:"type"`
			} `json:"errors"`
		} `json:"dashboardCreate"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return nil, fmt.Errorf("create dashboard: %w", err)
	}
	
	if len(result.DashboardCreate.Errors) > 0 {
		return nil, fmt.Errorf("dashboard creation failed: %s", 
			result.DashboardCreate.Errors[0].Description)
	}
	
	createdDashboard := &Dashboard{
		ID:          result.DashboardCreate.EntityResult.GUID,
		Name:        result.DashboardCreate.EntityResult.Name,
		AccountID:   result.DashboardCreate.EntityResult.AccountID,
		Permissions: result.DashboardCreate.EntityResult.Permissions,
		CreatedAt:   result.DashboardCreate.EntityResult.CreatedAt,
		UpdatedAt:   result.DashboardCreate.EntityResult.UpdatedAt,
		Pages:       dashboard.Pages,
	}
	
	return createdDashboard, nil
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
	query := `
		query($accountId: Int!, $searchCriteria: AlertsNrqlConditionSearchCriteria) {
			actor {
				account(id: $accountId) {
					alerts {
						nrqlConditionsSearch(searchCriteria: $searchCriteria) {
							nrqlConditions {
								id
								name
								enabled
								nrql {
									query
								}
								policyId
								terms {
									threshold
									thresholdDuration
									operator
									priority
								}
								createdAt
								updatedAt
							}
						}
					}
				}
			}
		}
	`
	
	searchCriteria := map[string]interface{}{}
	if policyID != "" {
		searchCriteria["policyId"] = policyID
	}
	
	variables := map[string]interface{}{
		"accountId": c.accountID,
		"searchCriteria": searchCriteria,
	}
	
	var result struct {
		Actor struct {
			Account struct {
				Alerts struct {
					NrqlConditionsSearch struct {
						NrqlConditions []struct {
							ID      string `json:"id"`
							Name    string `json:"name"`
							Enabled bool   `json:"enabled"`
							Nrql    struct {
								Query string `json:"query"`
							} `json:"nrql"`
							PolicyID string `json:"policyId"`
							Terms    []struct {
								Threshold         float64 `json:"threshold"`
								ThresholdDuration int     `json:"thresholdDuration"`
								Operator          string  `json:"operator"`
								Priority          string  `json:"priority"`
							} `json:"terms"`
							CreatedAt time.Time `json:"createdAt"`
							UpdatedAt time.Time `json:"updatedAt"`
						} `json:"nrqlConditions"`
					} `json:"nrqlConditionsSearch"`
				} `json:"alerts"`
			} `json:"account"`
		} `json:"actor"`
	}
	
	if err := c.queryGraphQL(ctx, query, variables, &result); err != nil {
		return nil, fmt.Errorf("list alert conditions: %w", err)
	}
	
	conditions := make([]AlertCondition, 0, len(result.Actor.Account.Alerts.NrqlConditionsSearch.NrqlConditions))
	for _, cond := range result.Actor.Account.Alerts.NrqlConditionsSearch.NrqlConditions {
		alertCond := AlertCondition{
			ID:        cond.ID,
			Name:      cond.Name,
			Query:     cond.Nrql.Query,
			Enabled:   cond.Enabled,
			PolicyID:  cond.PolicyID,
			CreatedAt: cond.CreatedAt,
			UpdatedAt: cond.UpdatedAt,
		}
		
		// Use first term for threshold info
		if len(cond.Terms) > 0 {
			alertCond.Threshold = cond.Terms[0].Threshold
			alertCond.ThresholdDuration = cond.Terms[0].ThresholdDuration
			alertCond.Comparison = cond.Terms[0].Operator
		}
		
		conditions = append(conditions, alertCond)
	}
	
	return conditions, nil
}

// CreateAlertCondition creates a new alert condition
func (c *Client) CreateAlertCondition(ctx context.Context, condition AlertCondition) (*AlertCondition, error) {
	mutation := `
		mutation($accountId: Int!, $policyId: ID!, $condition: AlertsNrqlConditionInput!) {
			alertsNrqlConditionCreate(accountId: $accountId, policyId: $policyId, condition: $condition) {
				id
				name
				enabled
				nrql {
					query
				}
				policyId
				terms {
					threshold
					thresholdDuration
					operator
					priority
				}
				createdAt
				updatedAt
			}
		}
	`
	
	// Build condition input
	conditionInput := map[string]interface{}{
		"name":    condition.Name,
		"enabled": condition.Enabled,
		"nrql": map[string]interface{}{
			"query": condition.Query,
		},
		"terms": []map[string]interface{}{
			{
				"threshold":         condition.Threshold,
				"thresholdDuration": condition.ThresholdDuration,
				"operator":          condition.Comparison,
				"priority":          "CRITICAL",
			},
		},
	}
	
	variables := map[string]interface{}{
		"accountId": c.accountID,
		"policyId":  condition.PolicyID,
		"condition": conditionInput,
	}
	
	var result struct {
		AlertsNrqlConditionCreate struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
			Nrql    struct {
				Query string `json:"query"`
			} `json:"nrql"`
			PolicyID string `json:"policyId"`
			Terms    []struct {
				Threshold         float64 `json:"threshold"`
				ThresholdDuration int     `json:"thresholdDuration"`
				Operator          string  `json:"operator"`
				Priority          string  `json:"priority"`
			} `json:"terms"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
		} `json:"alertsNrqlConditionCreate"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return nil, fmt.Errorf("create alert condition: %w", err)
	}
	
	created := &AlertCondition{
		ID:        result.AlertsNrqlConditionCreate.ID,
		Name:      result.AlertsNrqlConditionCreate.Name,
		Query:     result.AlertsNrqlConditionCreate.Nrql.Query,
		Enabled:   result.AlertsNrqlConditionCreate.Enabled,
		PolicyID:  result.AlertsNrqlConditionCreate.PolicyID,
		CreatedAt: result.AlertsNrqlConditionCreate.CreatedAt,
		UpdatedAt: result.AlertsNrqlConditionCreate.UpdatedAt,
	}
	
	if len(result.AlertsNrqlConditionCreate.Terms) > 0 {
		created.Threshold = result.AlertsNrqlConditionCreate.Terms[0].Threshold
		created.ThresholdDuration = result.AlertsNrqlConditionCreate.Terms[0].ThresholdDuration
		created.Comparison = result.AlertsNrqlConditionCreate.Terms[0].Operator
	}
	
	return created, nil
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

// queryGraphQL executes a GraphQL query and unmarshals the result
func (c *Client) queryGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	reqBody := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		reqBody["variables"] = variables
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.graphQLURL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GraphQL request failed with status %d: %s", resp.StatusCode, body)
	}

	var graphQLResp struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphQLResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	if result != nil && graphQLResp.Data != nil {
		if err := json.Unmarshal(graphQLResp.Data, result); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	}

	return nil
}

// QueryGraphQL executes a GraphQL query (exported for use by other packages)
func (c *Client) QueryGraphQL(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.queryGraphQL(ctx, query, variables, &result)
	return result, err
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

// GetDashboard retrieves full dashboard details with widgets
func (c *Client) GetDashboard(ctx context.Context, dashboardGUID string) (*Dashboard, error) {
	query := `
		query($guid: EntityGuid!) {
			actor {
				entity(guid: $guid) {
					... on DashboardEntity {
						guid
						name
						description
						accountId
						permissions
						createdAt
						updatedAt
						pages {
							name
							widgets {
								title
								visualization
								configuration
								rawConfiguration
							}
						}
					}
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"guid": dashboardGUID,
	}
	
	var result struct {
		Actor struct {
			Entity *Dashboard `json:"entity"`
		} `json:"actor"`
	}
	
	if err := c.queryGraphQL(ctx, query, variables, &result); err != nil {
		return nil, fmt.Errorf("get dashboard: %w", err)
	}
	
	if result.Actor.Entity == nil {
		return nil, fmt.Errorf("dashboard not found: %s", dashboardGUID)
	}
	
	return result.Actor.Entity, nil
}

// UpdateDashboard updates an existing dashboard
func (c *Client) UpdateDashboard(ctx context.Context, dashboardGUID string, updates Dashboard) (*Dashboard, error) {
	mutation := `
		mutation($guid: EntityGuid!, $dashboard: DashboardInput!) {
			dashboardUpdate(guid: $guid, dashboard: $dashboard) {
				entityResult {
					guid
					name
					accountId
					createdAt
					updatedAt
					permissions
				}
				errors {
					description
					type
				}
			}
		}
	`
	
	dashboardInput := map[string]interface{}{
		"name":        updates.Name,
		"permissions": updates.Permissions,
	}
	if updates.Description != "" {
		dashboardInput["description"] = updates.Description
	}
	if len(updates.Pages) > 0 {
		dashboardInput["pages"] = updates.Pages
	}
	
	variables := map[string]interface{}{
		"guid":      dashboardGUID,
		"dashboard": dashboardInput,
	}
	
	var result struct {
		DashboardUpdate struct {
			EntityResult struct {
				GUID        string    `json:"guid"`
				Name        string    `json:"name"`
				AccountID   int       `json:"accountId"`
				CreatedAt   time.Time `json:"createdAt"`
				UpdatedAt   time.Time `json:"updatedAt"`
				Permissions string    `json:"permissions"`
			} `json:"entityResult"`
			Errors []struct {
				Description string `json:"description"`
				Type        string `json:"type"`
			} `json:"errors"`
		} `json:"dashboardUpdate"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return nil, fmt.Errorf("update dashboard: %w", err)
	}
	
	if len(result.DashboardUpdate.Errors) > 0 {
		return nil, fmt.Errorf("dashboard update failed: %s",
			result.DashboardUpdate.Errors[0].Description)
	}
	
	updatedDashboard := &Dashboard{
		ID:          result.DashboardUpdate.EntityResult.GUID,
		Name:        result.DashboardUpdate.EntityResult.Name,
		AccountID:   result.DashboardUpdate.EntityResult.AccountID,
		Permissions: result.DashboardUpdate.EntityResult.Permissions,
		CreatedAt:   result.DashboardUpdate.EntityResult.CreatedAt,
		UpdatedAt:   result.DashboardUpdate.EntityResult.UpdatedAt,
	}
	
	return updatedDashboard, nil
}

// DeleteDashboard deletes a dashboard
func (c *Client) DeleteDashboard(ctx context.Context, dashboardGUID string) error {
	mutation := `
		mutation($guid: EntityGuid!) {
			dashboardDelete(guid: $guid) {
				status
				errors {
					description
					type
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"guid": dashboardGUID,
	}
	
	var result struct {
		DashboardDelete struct {
			Status string `json:"status"`
			Errors []struct {
				Description string `json:"description"`
				Type        string `json:"type"`
			} `json:"errors"`
		} `json:"dashboardDelete"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return fmt.Errorf("delete dashboard: %w", err)
	}
	
	if len(result.DashboardDelete.Errors) > 0 {
		return fmt.Errorf("dashboard deletion failed: %s",
			result.DashboardDelete.Errors[0].Description)
	}
	
	return nil
}

// UpdateAlertCondition updates an existing alert condition
func (c *Client) UpdateAlertCondition(ctx context.Context, conditionID string, updates map[string]interface{}) (*AlertCondition, error) {
	mutation := `
		mutation($accountId: Int!, $conditionId: ID!, $condition: AlertsNrqlConditionInput!) {
			alertsNrqlConditionUpdate(accountId: $accountId, id: $conditionId, condition: $condition) {
				id
				name
				enabled
				nrql {
					query
				}
				policyId
				terms {
					threshold
					thresholdDuration
					operator
					priority
				}
				createdAt
				updatedAt
			}
		}
	`
	
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}
	
	variables := map[string]interface{}{
		"accountId":   accountIDInt,
		"conditionId": conditionID,
		"condition":   updates,
	}
	
	var result struct {
		AlertsNrqlConditionUpdate AlertCondition `json:"alertsNrqlConditionUpdate"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return nil, fmt.Errorf("update alert condition: %w", err)
	}
	
	return &result.AlertsNrqlConditionUpdate, nil
}

// DeleteAlertCondition deletes an alert condition
func (c *Client) DeleteAlertCondition(ctx context.Context, conditionID string) error {
	mutation := `
		mutation($accountId: Int!, $conditionId: ID!) {
			alertsNrqlConditionDelete(accountId: $accountId, id: $conditionId) {
				id
			}
		}
	`
	
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return fmt.Errorf("invalid account ID: %w", err)
	}
	
	variables := map[string]interface{}{
		"accountId":   accountIDInt,
		"conditionId": conditionID,
	}
	
	var result struct {
		AlertsNrqlConditionDelete struct {
			ID string `json:"id"`
		} `json:"alertsNrqlConditionDelete"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return fmt.Errorf("delete alert condition: %w", err)
	}
	
	return nil
}

// EnableAlertCondition enables an alert condition
func (c *Client) EnableAlertCondition(ctx context.Context, conditionID string) error {
	updates := map[string]interface{}{
		"enabled": true,
	}
	_, err := c.UpdateAlertCondition(ctx, conditionID, updates)
	return err
}

// DisableAlertCondition disables an alert condition
func (c *Client) DisableAlertCondition(ctx context.Context, conditionID string) error {
	updates := map[string]interface{}{
		"enabled": false,
	}
	_, err := c.UpdateAlertCondition(ctx, conditionID, updates)
	return err
}

// AlertPolicy represents an alert policy
type AlertPolicy struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	IncidentPreference string    `json:"incidentPreference"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// CreateAlertPolicy creates a new alert policy
func (c *Client) CreateAlertPolicy(ctx context.Context, policy AlertPolicy) (*AlertPolicy, error) {
	mutation := `
		mutation($accountId: Int!, $policy: AlertsPolicyInput!) {
			alertsPolicy(accountId: $accountId, policy: $policy) {
				id
				name
				incidentPreference
				createdAt
				updatedAt
			}
		}
	`
	
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}
	
	policyInput := map[string]interface{}{
		"name":               policy.Name,
		"incidentPreference": policy.IncidentPreference,
	}
	
	variables := map[string]interface{}{
		"accountId": accountIDInt,
		"policy":    policyInput,
	}
	
	var result struct {
		AlertsPolicy AlertPolicy `json:"alertsPolicy"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return nil, fmt.Errorf("create alert policy: %w", err)
	}
	
	return &result.AlertsPolicy, nil
}

// UpdateAlertPolicy updates an existing alert policy
func (c *Client) UpdateAlertPolicy(ctx context.Context, policyID string, updates map[string]interface{}) (*AlertPolicy, error) {
	mutation := `
		mutation($accountId: Int!, $policyId: ID!, $policy: AlertsPolicyInput!) {
			alertsPolicyUpdate(accountId: $accountId, id: $policyId, policy: $policy) {
				id
				name
				incidentPreference
				createdAt
				updatedAt
			}
		}
	`
	
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}
	
	variables := map[string]interface{}{
		"accountId": accountIDInt,
		"policyId":  policyID,
		"policy":    updates,
	}
	
	var result struct {
		AlertsPolicyUpdate AlertPolicy `json:"alertsPolicyUpdate"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return nil, fmt.Errorf("update alert policy: %w", err)
	}
	
	return &result.AlertsPolicyUpdate, nil
}

// DeleteAlertPolicy deletes an alert policy
func (c *Client) DeleteAlertPolicy(ctx context.Context, policyID string) error {
	mutation := `
		mutation($accountId: Int!, $policyId: ID!) {
			alertsPolicyDelete(accountId: $accountId, id: $policyId) {
				id
			}
		}
	`
	
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return fmt.Errorf("invalid account ID: %w", err)
	}
	
	variables := map[string]interface{}{
		"accountId": accountIDInt,
		"policyId":  policyID,
	}
	
	var result struct {
		AlertsPolicyDelete struct {
			ID string `json:"id"`
		} `json:"alertsPolicyDelete"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return fmt.Errorf("delete alert policy: %w", err)
	}
	
	return nil
}

// CloseIncident closes an open alert incident
func (c *Client) CloseIncident(ctx context.Context, incidentID string) error {
	mutation := `
		mutation($accountId: Int!, $incidentId: ID!) {
			alertsIncidentClose(accountId: $accountId, incidentId: $incidentId) {
				incident {
					id
					state
					closedAt
				}
				error {
					description
				}
			}
		}
	`
	
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return fmt.Errorf("invalid account ID: %w", err)
	}
	
	variables := map[string]interface{}{
		"accountId":  accountIDInt,
		"incidentId": incidentID,
	}
	
	var result struct {
		AlertsIncidentClose struct {
			Incident struct {
				ID       string     `json:"id"`
				State    string     `json:"state"`
				ClosedAt *time.Time `json:"closedAt"`
			} `json:"incident"`
			Error *struct {
				Description string `json:"description"`
			} `json:"error"`
		} `json:"alertsIncidentClose"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return fmt.Errorf("close incident: %w", err)
	}
	
	if result.AlertsIncidentClose.Error != nil {
		return fmt.Errorf("close incident failed: %s", result.AlertsIncidentClose.Error.Description)
	}
	
	return nil
}

// GetAlertAnalytics retrieves alert effectiveness metrics
func (c *Client) GetAlertAnalytics(ctx context.Context, conditionID string, since string) (map[string]interface{}, error) {
	// Query to get incident history and metrics
	query := fmt.Sprintf(`
		SELECT 
			count(*) as incident_count,
			average(duration) as avg_duration_minutes,
			percentile(duration, 50) as median_duration_minutes,
			min(duration) as min_duration_minutes,
			max(duration) as max_duration_minutes
		FROM AlertIncident
		WHERE conditionId = '%s'
		SINCE %s
	`, conditionID, since)
	
	result, err := c.QueryNRQL(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get alert analytics: %w", err)
	}
	
	if len(result.Results) == 0 {
		return map[string]interface{}{
			"incident_count":          0,
			"avg_duration_minutes":    0,
			"median_duration_minutes": 0,
			"min_duration_minutes":    0,
			"max_duration_minutes":    0,
		}, nil
	}
	
	return result.Results[0], nil
}

// SyntheticMonitor represents a synthetic monitor
type SyntheticMonitor struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"uri"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Frequency int       `json:"frequency"`
	Locations []string  `json:"locations"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateSyntheticMonitor creates a new synthetic monitor
func (c *Client) CreateSyntheticMonitor(ctx context.Context, monitor SyntheticMonitor) (*SyntheticMonitor, error) {
	mutation := `
		mutation($accountId: Int!, $monitor: SyntheticsCreateSimpleBrowserMonitorInput!) {
			syntheticsCreateSimpleBrowserMonitor(accountId: $accountId, monitor: $monitor) {
				monitor {
					id
					name
					uri
					status
					period
					locations
					createdAt
					modifiedAt
				}
				errors {
					description
					type
				}
			}
		}
	`
	
	accountIDInt, err := strconv.Atoi(c.accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}
	
	// Default values
	if monitor.Frequency == 0 {
		monitor.Frequency = 5 // Default 5 minutes
	}
	if len(monitor.Locations) == 0 {
		monitor.Locations = []string{"US_EAST_1"} // Default location
	}
	
	monitorInput := map[string]interface{}{
		"name":      monitor.Name,
		"uri":       monitor.URL,
		"period":    fmt.Sprintf("EVERY_%d_MINUTES", monitor.Frequency),
		"status":    "ENABLED",
		"locations": monitor.Locations,
	}
	
	variables := map[string]interface{}{
		"accountId": accountIDInt,
		"monitor":   monitorInput,
	}
	
	var result struct {
		SyntheticsCreateSimpleBrowserMonitor struct {
			Monitor struct {
				ID        string    `json:"id"`
				Name      string    `json:"name"`
				URI       string    `json:"uri"`
				Status    string    `json:"status"`
				Period    string    `json:"period"`
				Locations []string  `json:"locations"`
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"modifiedAt"`
			} `json:"monitor"`
			Errors []struct {
				Description string `json:"description"`
				Type        string `json:"type"`
			} `json:"errors"`
		} `json:"syntheticsCreateSimpleBrowserMonitor"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return nil, fmt.Errorf("create synthetic monitor: %w", err)
	}
	
	if len(result.SyntheticsCreateSimpleBrowserMonitor.Errors) > 0 {
		return nil, fmt.Errorf("monitor creation failed: %s",
			result.SyntheticsCreateSimpleBrowserMonitor.Errors[0].Description)
	}
	
	created := &SyntheticMonitor{
		ID:        result.SyntheticsCreateSimpleBrowserMonitor.Monitor.ID,
		Name:      result.SyntheticsCreateSimpleBrowserMonitor.Monitor.Name,
		URL:       result.SyntheticsCreateSimpleBrowserMonitor.Monitor.URI,
		Status:    result.SyntheticsCreateSimpleBrowserMonitor.Monitor.Status,
		Locations: result.SyntheticsCreateSimpleBrowserMonitor.Monitor.Locations,
		CreatedAt: result.SyntheticsCreateSimpleBrowserMonitor.Monitor.CreatedAt,
		UpdatedAt: result.SyntheticsCreateSimpleBrowserMonitor.Monitor.UpdatedAt,
	}
	
	return created, nil
}

// DeleteEntity deletes an entity (dashboard, monitor, etc.)
func (c *Client) DeleteEntity(ctx context.Context, entityGUID string) error {
	mutation := `
		mutation($guid: EntityGuid!) {
			entityDelete(guid: $guid) {
				deletedEntities
				failures {
					guid
					message
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"guid": entityGUID,
	}
	
	var result struct {
		EntityDelete struct {
			DeletedEntities []string `json:"deletedEntities"`
			Failures        []struct {
				GUID    string `json:"guid"`
				Message string `json:"message"`
			} `json:"failures"`
		} `json:"entityDelete"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}
	
	if len(result.EntityDelete.Failures) > 0 {
		return fmt.Errorf("entity deletion failed: %s", result.EntityDelete.Failures[0].Message)
	}
	
	return nil
}

// UndeleteDashboard restores a previously deleted dashboard
func (c *Client) UndeleteDashboard(ctx context.Context, dashboardGUID string) (*Dashboard, error) {
	mutation := `
		mutation($guid: EntityGuid!) {
			dashboardUndelete(guid: $guid) {
				entityResult {
					guid
					name
					accountId
					createdAt
					updatedAt
					permissions
				}
				errors {
					description
					type
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"guid": dashboardGUID,
	}
	
	var result struct {
		DashboardUndelete struct {
			EntityResult struct {
				GUID        string    `json:"guid"`
				Name        string    `json:"name"`
				AccountID   int       `json:"accountId"`
				CreatedAt   time.Time `json:"createdAt"`
				UpdatedAt   time.Time `json:"updatedAt"`
				Permissions string    `json:"permissions"`
			} `json:"entityResult"`
			Errors []struct {
				Description string `json:"description"`
				Type        string `json:"type"`
			} `json:"errors"`
		} `json:"dashboardUndelete"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return nil, fmt.Errorf("undelete dashboard: %w", err)
	}
	
	if len(result.DashboardUndelete.Errors) > 0 {
		return nil, fmt.Errorf("dashboard undelete failed: %s",
			result.DashboardUndelete.Errors[0].Description)
	}
	
	restoredDashboard := &Dashboard{
		ID:          result.DashboardUndelete.EntityResult.GUID,
		Name:        result.DashboardUndelete.EntityResult.Name,
		AccountID:   result.DashboardUndelete.EntityResult.AccountID,
		Permissions: result.DashboardUndelete.EntityResult.Permissions,
		CreatedAt:   result.DashboardUndelete.EntityResult.CreatedAt,
		UpdatedAt:   result.DashboardUndelete.EntityResult.UpdatedAt,
	}
	
	return restoredDashboard, nil
}

// CreateDashboardSnapshotUrl creates a public URL for a static dashboard snapshot
func (c *Client) CreateDashboardSnapshotUrl(ctx context.Context, dashboardGUID string) (string, error) {
	mutation := `
		mutation($guid: EntityGuid!) {
			dashboardCreateSnapshotUrl(guid: $guid) {
				url
				errors {
					description
					type
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"guid": dashboardGUID,
	}
	
	var result struct {
		DashboardCreateSnapshotUrl struct {
			URL    string `json:"url"`
			Errors []struct {
				Description string `json:"description"`
				Type        string `json:"type"`
			} `json:"errors"`
		} `json:"dashboardCreateSnapshotUrl"`
	}
	
	if err := c.queryGraphQL(ctx, mutation, variables, &result); err != nil {
		return "", fmt.Errorf("create dashboard snapshot URL: %w", err)
	}
	
	if len(result.DashboardCreateSnapshotUrl.Errors) > 0 {
		return "", fmt.Errorf("dashboard snapshot URL creation failed: %s",
			result.DashboardCreateSnapshotUrl.Errors[0].Description)
	}
	
	return result.DashboardCreateSnapshotUrl.URL, nil
}
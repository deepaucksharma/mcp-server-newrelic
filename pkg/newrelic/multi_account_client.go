package newrelic

import (
	"context"
	"fmt"
	"sync"
)

// MultiAccountClient manages multiple New Relic accounts and allows switching between them
type MultiAccountClient struct {
	// Primary client (from config)
	primaryClient *Client
	primaryAccountID string
	
	// Additional account clients
	accountClients map[string]*Client
	mu sync.RWMutex
	
	// Default config for creating new clients
	defaultConfig Config
}

// NewMultiAccountClient creates a new multi-account client
func NewMultiAccountClient(config Config) (*MultiAccountClient, error) {
	// Create primary client
	primaryClient, err := NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary client: %w", err)
	}
	
	return &MultiAccountClient{
		primaryClient:    primaryClient,
		primaryAccountID: config.AccountID,
		accountClients:   make(map[string]*Client),
		defaultConfig:    config,
	}, nil
}

// WithAccount returns a client for the specified account ID
// If accountID is empty or matches primary, returns primary client
func (m *MultiAccountClient) WithAccount(accountID string) (*Client, error) {
	// Use primary client if no account specified or it matches
	if accountID == "" || accountID == m.primaryAccountID {
		return m.primaryClient, nil
	}
	
	// Check if we already have a client for this account
	m.mu.RLock()
	client, exists := m.accountClients[accountID]
	m.mu.RUnlock()
	
	if exists {
		return client, nil
	}
	
	// Create new client for this account
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Double-check after acquiring write lock
	if client, exists := m.accountClients[accountID]; exists {
		return client, nil
	}
	
	// Create new client with same config but different account ID
	config := m.defaultConfig
	config.AccountID = accountID
	
	newClient, err := NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for account %s: %w", accountID, err)
	}
	
	m.accountClients[accountID] = newClient
	return newClient, nil
}

// GetPrimaryClient returns the primary client
func (m *MultiAccountClient) GetPrimaryClient() *Client {
	return m.primaryClient
}

// ListAccounts returns all configured account IDs
func (m *MultiAccountClient) ListAccounts() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	accounts := make([]string, 0, len(m.accountClients)+1)
	accounts = append(accounts, m.primaryAccountID)
	
	for accountID := range m.accountClients {
		accounts = append(accounts, accountID)
	}
	
	return accounts
}

// QueryNRQL executes an NRQL query with optional account override
func (m *MultiAccountClient) QueryNRQL(ctx context.Context, query string, accountID string) (*NRQLResult, error) {
	client, err := m.WithAccount(accountID)
	if err != nil {
		return nil, err
	}
	return client.QueryNRQL(ctx, query)
}

// QueryNRQLWithOptions executes an NRQL query with options including account override
func (m *MultiAccountClient) QueryNRQLWithOptions(ctx context.Context, query string, opts QueryOptions) (*NRQLResult, error) {
	client, err := m.WithAccount(opts.AccountID)
	if err != nil {
		return nil, err
	}
	return client.QueryNRQLWithOptions(ctx, query, opts)
}

// ListDashboards lists dashboards with optional account override
func (m *MultiAccountClient) ListDashboards(ctx context.Context, accountID string) ([]Dashboard, error) {
	client, err := m.WithAccount(accountID)
	if err != nil {
		return nil, err
	}
	return client.ListDashboards(ctx)
}

// GetDashboard gets a dashboard by GUID (account is encoded in GUID)
func (m *MultiAccountClient) GetDashboard(ctx context.Context, guid string) (*Dashboard, error) {
	// For now, use primary client as GUID is unique across accounts
	// In future, we might parse account from GUID if needed
	return m.primaryClient.GetDashboard(ctx, guid)
}

// CreateDashboard creates a dashboard in the specified account
func (m *MultiAccountClient) CreateDashboard(ctx context.Context, dashboard Dashboard, accountID string) (*Dashboard, error) {
	client, err := m.WithAccount(accountID)
	if err != nil {
		return nil, err
	}
	return client.CreateDashboard(ctx, dashboard)
}

// CreateAlert creates an alert in the specified account
func (m *MultiAccountClient) CreateAlert(ctx context.Context, alert Alert, accountID string) (*Alert, error) {
	client, err := m.WithAccount(accountID)
	if err != nil {
		return nil, err
	}
	return client.CreateAlert(ctx, alert)
}

// ListAlerts lists alerts with optional account override
func (m *MultiAccountClient) ListAlerts(ctx context.Context, accountID string) ([]Alert, error) {
	client, err := m.WithAccount(accountID)
	if err != nil {
		return nil, err
	}
	return client.ListAlerts(ctx)
}

// BulkUpdateAlerts updates multiple alerts in the specified account
func (m *MultiAccountClient) BulkUpdateAlerts(ctx context.Context, updates []AlertUpdate, accountID string) ([]Alert, error) {
	client, err := m.WithAccount(accountID)
	if err != nil {
		return nil, err
	}
	return client.BulkUpdateAlerts(ctx, updates)
}
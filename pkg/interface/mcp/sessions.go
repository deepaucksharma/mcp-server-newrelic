package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	sessionIDLength     = 32
	sessionTimeout      = 30 * time.Minute
	sessionCleanupBatch = 100
)

// sessionManager implements the SessionManager interface
type sessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	clientInfo  map[string]interface{} // Store client info by request ID
}

// NewSessionManager creates a new session manager
func NewSessionManager() SessionManager {
	return &sessionManager{
		sessions:   make(map[string]*Session),
		clientInfo: make(map[string]interface{}),
	}
}

// Create creates a new session
func (m *sessionManager) Create() *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	session := &Session{
		ID:        generateSessionID(),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		Context:   make(map[string]interface{}),
	}
	
	m.sessions[session.ID] = session
	return session
}

// Get retrieves a session by ID
func (m *sessionManager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	session, exists := m.sessions[id]
	if exists {
		// Update last used time
		session.LastUsed = time.Now()
	}
	return session, exists
}

// Update updates a session
func (m *sessionManager) Update(session *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.sessions[session.ID]; !exists {
		return fmt.Errorf("session %s not found", session.ID)
	}
	
	session.LastUsed = time.Now()
	m.sessions[session.ID] = session
	return nil
}

// Delete removes a session
func (m *sessionManager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.sessions[id]; !exists {
		return fmt.Errorf("session %s not found", id)
	}
	
	delete(m.sessions, id)
	return nil
}

// Cleanup removes expired sessions
func (m *sessionManager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	toDelete := make([]string, 0, sessionCleanupBatch)
	
	for id, session := range m.sessions {
		if now.Sub(session.LastUsed) > sessionTimeout {
			toDelete = append(toDelete, id)
			if len(toDelete) >= sessionCleanupBatch {
				break
			}
		}
	}
	
	for _, id := range toDelete {
		delete(m.sessions, id)
	}
	
	return nil
}

// generateSessionID creates a new random session ID
func generateSessionID() string {
	bytes := make([]byte, sessionIDLength/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// List returns all active sessions
func (m *sessionManager) List() []Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	sessions := make([]Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, *session)
	}
	return sessions
}

// End terminates a session
func (m *sessionManager) End(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.sessions, id)
}

// StoreClientInfo stores client information by request ID
func (m *sessionManager) StoreClientInfo(id string, info interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.clientInfo[id] = info
}

// GetClientInfo retrieves client information by request ID
func (m *sessionManager) GetClientInfo(id string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.clientInfo[id]
}
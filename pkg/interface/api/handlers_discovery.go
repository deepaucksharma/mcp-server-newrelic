//go:build !test && !nodiscovery

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/state"
)

// Handler implements all API endpoint handlers
type Handler struct {
	// Service interfaces from different tracks
	discovery discovery.DiscoveryEngine
	// TODO: Add pattern, query, dashboard services when available
	
	// New Relic APM
	nrApp *newrelic.Application
	
	// State management
	stateManager state.StateManager
}

// NewHandler creates a new API handler
func NewHandler() *Handler {
	return &Handler{}
}

// SetDiscoveryEngine sets the discovery engine (from Track 1)
func (h *Handler) SetDiscoveryEngine(engine discovery.DiscoveryEngine) {
	h.discovery = engine
}

// SetNewRelicApp sets the New Relic application for APM
func (h *Handler) SetNewRelicApp(app *newrelic.Application) {
	h.nrApp = app
}

// SetStateManager sets the state manager
func (h *Handler) SetStateManager(sm state.StateManager) {
	h.stateManager = sm
}

// GetHealth handles GET /health
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	components := make(map[string]interface{})

	// Check discovery service
	if h.discovery != nil {
		components["discovery"] = map[string]interface{}{
			"status":  "healthy",
			"message": "Discovery engine operational",
		}
	} else {
		status = "degraded"
		components["discovery"] = map[string]interface{}{
			"status":  "unavailable",
			"message": "Discovery engine not initialized",
		}
	}

	// Check state manager
	if h.stateManager != nil {
		ctx := r.Context()
		stats, err := h.stateManager.Stats(ctx)
		if err == nil {
			var hitRate float64
			if stats.HitCount + stats.MissCount > 0 {
				hitRate = float64(stats.HitCount) / float64(stats.HitCount + stats.MissCount)
			}
			components["state"] = map[string]interface{}{
				"status":   "healthy",
				"message":  "State management operational",
				"store":    "memory", // TODO: Detect actual store type
				"cache":    map[string]interface{}{
					"entries":  stats.TotalEntries,
					"hits":     stats.HitCount,
					"misses":   stats.MissCount,
					"evictions": stats.EvictCount,
					"hitRate":  hitRate,
					"memory":   stats.MemoryUsage,
				},
			}
		} else {
			components["state"] = map[string]interface{}{
				"status":  "degraded",
				"message": "State management error",
				"error":   err.Error(),
			}
		}
	} else {
		components["state"] = map[string]interface{}{
			"status":  "unavailable",
			"message": "State management not initialized",
		}
	}

	// TODO: Check other services

	response := map[string]interface{}{
		"status":     status,
		"version":    "1.0.0",
		"uptime":     "24h", // TODO: Track actual uptime
		"components": components,
	}

	writeJSON(w, http.StatusOK, response)
}

// ListSchemas handles GET /discovery/schemas
func (h *Handler) ListSchemas(w http.ResponseWriter, r *http.Request) {
	if h.discovery == nil {
		writeError(w, http.StatusServiceUnavailable, "Discovery service unavailable", nil)
		return
	}

	// Handle session tracking
	var sessionID string
	if h.stateManager != nil {
		// Check for existing session ID
		sessionID = r.Header.Get("X-Session-ID")
		if sessionID == "" {
			// Create new session if user goal is provided
			if goal := r.Header.Get("X-Session-Goal"); goal != "" {
				ctx := r.Context()
				session, err := h.stateManager.CreateSession(ctx, goal)
				if err == nil {
					sessionID = session.ID
					w.Header().Set("X-Session-ID", sessionID)
				}
			}
		}
	}
	
	// Parse query parameters
	filter := discovery.DiscoveryFilter{}
	
	if eventType := r.URL.Query().Get("eventType"); eventType != "" {
		filter.EventTypes = []string{eventType}
	}
	
	if minCount := r.URL.Query().Get("minRecordCount"); minCount != "" {
		if count, err := strconv.ParseInt(minCount, 10, 64); err == nil {
			filter.MinRecordCount = count
		}
	}
	
	if maxSchemas := r.URL.Query().Get("maxSchemas"); maxSchemas != "" {
		if max, err := strconv.Atoi(maxSchemas); err == nil {
			filter.MaxSchemas = max
		}
	} else {
		filter.MaxSchemas = 50 // Default
	}

	// Check cache first if session exists
	ctx := r.Context()
	var schemas []discovery.Schema
	var cacheHit bool
	startTime := time.Now()
	
	if h.stateManager != nil && sessionID != "" {
		cacheKey := fmt.Sprintf("discovery:schemas:%v", filter)
		if cached, found := h.stateManager.Get(ctx, cacheKey); found {
			if cachedSchemas, ok := cached.([]discovery.Schema); ok {
				schemas = cachedSchemas
				cacheHit = true
			}
		}
	}
	
	// Execute discovery if not cached
	if !cacheHit {
		var err error
		schemas, err = h.discovery.DiscoverSchemas(ctx, filter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to discover schemas", err.Error())
			return
		}
		
		// Cache the result
		if h.stateManager != nil && sessionID != "" {
			cacheKey := fmt.Sprintf("discovery:schemas:%v", filter)
			h.stateManager.Set(ctx, cacheKey, schemas, 5*time.Minute)
			
			// Update session with discovered schemas
			if session, err := h.stateManager.GetSession(ctx, sessionID); err == nil {
				for _, schema := range schemas {
					found := false
					for _, existing := range session.DiscoveredSchemas {
						if existing == schema.EventType {
							found = true
							break
						}
					}
					if !found {
						session.DiscoveredSchemas = append(session.DiscoveredSchemas, schema.EventType)
					}
				}
				h.stateManager.UpdateSession(ctx, session)
			}
		}
	}

	// Build response with metadata
	includeMetadata := r.URL.Query().Get("includeMetadata") == "true"
	response := map[string]interface{}{
		"schemas": schemas,
	}
	
	if includeMetadata {
		response["metadata"] = map[string]interface{}{
			"totalSchemas":  len(schemas),
			"executionTime": time.Since(startTime).String(),
			"cacheHit":      cacheHit,
			"sessionID":     sessionID,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// GetSchemaProfile handles GET /discovery/schemas/{eventType}
func (h *Handler) GetSchemaProfile(w http.ResponseWriter, r *http.Request) {
	if h.discovery == nil {
		writeError(w, http.StatusServiceUnavailable, "Discovery service unavailable", nil)
		return
	}

	vars := mux.Vars(r)
	eventType := vars["eventType"]
	
	// Parse depth parameter
	depth := discovery.ProfileDepthStandard
	switch r.URL.Query().Get("depth") {
	case "basic":
		depth = discovery.ProfileDepthBasic
	case "full":
		depth = discovery.ProfileDepthFull
	}

	// Get schema profile
	ctx := r.Context()
	schema, err := h.discovery.ProfileSchema(ctx, eventType, depth)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "schema not found" {
			writeError(w, http.StatusNotFound, "Schema not found", map[string]string{
				"eventType": eventType,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to profile schema", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, schema)
}

// FindRelationships handles POST /discovery/relationships
func (h *Handler) FindRelationships(w http.ResponseWriter, r *http.Request) {
	if h.discovery == nil {
		writeError(w, http.StatusServiceUnavailable, "Discovery service unavailable", nil)
		return
	}

	// Parse request body
	var req struct {
		Schemas []string `json:"schemas"`
		Options struct {
			MaxRelationships int     `json:"maxRelationships"`
			MinConfidence    float64 `json:"minConfidence"`
		} `json:"options"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate input
	if len(req.Schemas) < 2 {
		writeError(w, http.StatusBadRequest, "At least 2 schemas required", nil)
		return
	}

	// Set defaults
	if req.Options.MaxRelationships == 0 {
		req.Options.MaxRelationships = 10
	}
	if req.Options.MinConfidence == 0 {
		req.Options.MinConfidence = 0.7
	}

	// Get schemas
	ctx := r.Context()
	filter := discovery.DiscoveryFilter{
		EventTypes: req.Schemas,
	}
	
	schemas, err := h.discovery.DiscoverSchemas(ctx, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get schemas", err.Error())
		return
	}

	// Find relationships
	relationships, err := h.discovery.FindRelationships(ctx, schemas)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to find relationships", err.Error())
		return
	}

	// Filter by confidence
	filtered := []discovery.Relationship{}
	for _, rel := range relationships {
		if rel.Confidence >= req.Options.MinConfidence {
			filtered = append(filtered, rel)
		}
		if len(filtered) >= req.Options.MaxRelationships {
			break
		}
	}

	response := map[string]interface{}{
		"relationships": filtered,
	}

	writeJSON(w, http.StatusOK, response)
}

// AssessQuality handles GET /discovery/quality/{eventType}
func (h *Handler) AssessQuality(w http.ResponseWriter, r *http.Request) {
	if h.discovery == nil {
		writeError(w, http.StatusServiceUnavailable, "Discovery service unavailable", nil)
		return
	}

	vars := mux.Vars(r)
	eventType := vars["eventType"]
	
	// TODO: Parse time range parameter
	// timeRange := r.URL.Query().Get("timeRange")

	ctx := r.Context()
	report, err := h.discovery.AssessQuality(ctx, eventType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to assess quality", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// AnalyzePatterns handles POST /patterns/analyze
func (h *Handler) AnalyzePatterns(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement when pattern service is available
	writeError(w, http.StatusNotImplemented, "Pattern analysis not yet implemented", nil)
}

// GenerateQuery handles POST /query/generate
func (h *Handler) GenerateQuery(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement when query generation service is available
	writeError(w, http.StatusNotImplemented, "Query generation not yet implemented", nil)
}

// CreateDashboard handles POST /dashboard/create
func (h *Handler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement when dashboard service is available
	writeError(w, http.StatusNotImplemented, "Dashboard creation not yet implemented", nil)
}
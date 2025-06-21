package mcp

import (
	"context"
	"fmt"
	"time"
)

// WorkflowSupport provides lightweight workflow support tools
// The actual orchestration is done by the intelligent MCP client

// registerWorkflowSupportTools registers tools that help with workflow patterns
func (s *Server) registerWorkflowSupportTools() error {
	// 1. Session-based context storage for sharing data between tools
	contextStore := NewToolBuilder("workflow.store_context", "Store data in workflow context for use by subsequent tools").
		Category(CategoryUtility).
		Handler(s.handleWorkflowStoreContext).
		Required("key").
		Param("key", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Key to store the data under",
			},
			Examples: []interface{}{"service_list", "error_metrics", "anomaly_results"},
		}).
		Required("value").
		Param("value", EnhancedProperty{
			Property: Property{
				Type:        "any",
				Description: "Value to store (can be any type)",
			},
		}).
		Param("session_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Session ID for context isolation",
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 1
			p.Cacheable = false
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Store service list: workflow.store_context(key='services', value=['api', 'web', 'db'])",
				"Store metrics: workflow.store_context(key='baseline_metrics', value={cpu: 45, memory: 72})",
			}
			g.ChainsWith = []string{"workflow.get_context", "workflow.list_context"}
			g.ContextRequirements = []string{"Use to share data between workflow steps"}
		}).
		Build()

	if err := s.tools.Register(contextStore.Tool); err != nil {
		return err
	}

	// 2. Retrieve stored context data
	contextGet := NewToolBuilder("workflow.get_context", "Retrieve data from workflow context").
		Category(CategoryUtility).
		Handler(s.handleWorkflowGetContext).
		Required("key").
		Param("key", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Key to retrieve data for",
			},
		}).
		Param("session_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Session ID for context isolation",
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 1
			p.Cacheable = false
		}).
		Build()

	if err := s.tools.Register(contextGet.Tool); err != nil {
		return err
	}

	// 3. List all context keys
	contextList := NewToolBuilder("workflow.list_context", "List all keys in workflow context").
		Category(CategoryUtility).
		Handler(s.handleWorkflowListContext).
		Param("session_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Session ID for context isolation",
			},
		}).
		Build()

	if err := s.tools.Register(contextList.Tool); err != nil {
		return err
	}

	// 4. Workflow templates as guidance (not execution)
	suggestWorkflow := NewToolBuilder("workflow.suggest_steps", "Suggest workflow steps for common scenarios").
		Category(CategoryAnalysis).
		Handler(s.handleWorkflowSuggestSteps).
		Required("scenario").
		Param("scenario", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Scenario to get workflow suggestions for",
				Enum: []string{
					"incident_investigation",
					"performance_optimization",
					"capacity_planning",
					"error_analysis",
					"dependency_mapping",
					"cost_optimization",
				},
			},
		}).
		Param("context", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Context about the current situation",
			},
			Examples: []interface{}{
				map[string]interface{}{
					"service":      "checkout-api",
					"symptom":      "high latency",
					"time_started": "15 minutes ago",
				},
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 10
			p.Cacheable = true
			p.CacheTTLSeconds = 3600
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Get incident steps: workflow.suggest_steps(scenario='incident_investigation', context={service: 'api', symptom: 'errors'})",
			}
			g.CommonPatterns = []string{
				"Get suggested tool sequence for scenarios",
				"AI should adapt based on results",
			}
			g.WarningsForAI = []string{
				"These are suggestions only - adapt based on actual findings",
				"Skip irrelevant steps based on context",
				"Add additional steps as needed",
			}
		}).
		Build()

	if err := s.tools.Register(suggestWorkflow.Tool); err != nil {
		return err
	}

	// 5. Progress tracking for user visibility
	trackProgress := NewToolBuilder("workflow.track_progress", "Track workflow progress for user visibility").
		Category(CategoryUtility).
		Handler(s.handleWorkflowTrackProgress).
		Required("step_name").
		Param("step_name", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Name of the current step",
			},
		}).
		Required("status").
		Param("status", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Status of the step",
				Enum:        []string{"started", "in_progress", "completed", "failed", "skipped"},
			},
		}).
		Param("details", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Additional details about the step",
			},
		}).
		Param("session_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Session ID for tracking isolation",
			},
		}).
		Build()

	if err := s.tools.Register(trackProgress.Tool); err != nil {
		return err
	}

	// 6. Findings aggregator
	recordFinding := NewToolBuilder("workflow.record_finding", "Record a finding during investigation").
		Category(CategoryAnalysis).
		Handler(s.handleWorkflowRecordFinding).
		Required("finding_type").
		Param("finding_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Type of finding",
				Enum:        []string{"anomaly", "correlation", "root_cause", "impact", "recommendation"},
			},
		}).
		Required("description").
		Param("description", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Description of the finding",
			},
		}).
		Param("severity", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Severity of the finding",
				Enum:        []string{"critical", "high", "medium", "low", "info"},
				Default:     "medium",
			},
		}).
		Param("evidence", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Supporting evidence for the finding",
			},
		}).
		Param("related_findings", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "IDs of related findings",
				Items:       &Property{Type: "string"},
			},
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Record anomaly: workflow.record_finding(finding_type='anomaly', description='CPU spike detected', severity='high', evidence={cpu_percent: 95})",
			}
			g.CommonPatterns = []string{
				"Build a comprehensive picture by recording findings as you investigate",
			}
		}).
		Build()

	if err := s.tools.Register(recordFinding.Tool); err != nil {
		return err
	}

	// 7. Get all findings
	getFindings := NewToolBuilder("workflow.get_findings", "Get all recorded findings").
		Category(CategoryQuery).
		Handler(s.handleWorkflowGetFindings).
		Param("finding_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Filter by finding type",
			},
		}).
		Param("min_severity", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Minimum severity to include",
				Enum:        []string{"critical", "high", "medium", "low", "info"},
			},
		}).
		Param("session_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Session ID for finding isolation",
			},
		}).
		Build()

	return s.tools.Register(getFindings.Tool)
}

// Handler implementations focused on supporting AI-driven orchestration

func (s *Server) handleWorkflowStoreContext(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	key, ok := params["key"].(string)
	if !ok || key == "" {
		return nil, fmt.Errorf("key is required")
	}

	value, exists := params["value"]
	if !exists {
		return nil, fmt.Errorf("value is required")
	}

	sessionID := s.getSessionID(params)
	session := s.getOrCreateSession(sessionID)

	// Store in session context
	session.Context[key] = value
	s.sessions.Update(session)

	return map[string]interface{}{
		"status":     "stored",
		"key":        key,
		"session_id": session.ID,
		"timestamp":  time.Now(),
	}, nil
}

func (s *Server) handleWorkflowGetContext(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	key, ok := params["key"].(string)
	if !ok || key == "" {
		return nil, fmt.Errorf("key is required")
	}

	sessionID := s.getSessionID(params)
	session, exists := s.sessions.Get(sessionID)
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	value, exists := session.Context[key]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found in context", key)
	}

	return map[string]interface{}{
		"key":   key,
		"value": value,
	}, nil
}

func (s *Server) handleWorkflowListContext(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	sessionID := s.getSessionID(params)
	session, exists := s.sessions.Get(sessionID)
	if !exists {
		return map[string]interface{}{
			"keys":  []string{},
			"count": 0,
		}, nil
	}

	keys := make([]string, 0, len(session.Context))
	for k := range session.Context {
		keys = append(keys, k)
	}

	return map[string]interface{}{
		"keys":       keys,
		"count":      len(keys),
		"session_id": session.ID,
	}, nil
}

func (s *Server) handleWorkflowSuggestSteps(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	scenario, ok := params["scenario"].(string)
	if !ok || scenario == "" {
		return nil, fmt.Errorf("scenario is required")
	}

	context := make(map[string]interface{})
	if c, ok := params["context"].(map[string]interface{}); ok {
		context = c
	}

	// Return workflow suggestions based on scenario
	suggestions := s.getWorkflowSuggestions(scenario, context)

	return map[string]interface{}{
		"scenario":    scenario,
		"steps":       suggestions.Steps,
		"description": suggestions.Description,
		"tips":        suggestions.Tips,
		"adaption":    "These are suggested steps - adapt based on findings and skip/add as needed",
	}, nil
}

func (s *Server) handleWorkflowTrackProgress(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stepName, ok := params["step_name"].(string)
	if !ok || stepName == "" {
		return nil, fmt.Errorf("step_name is required")
	}

	status, ok := params["status"].(string)
	if !ok || status == "" {
		return nil, fmt.Errorf("status is required")
	}

	details := make(map[string]interface{})
	if d, ok := params["details"].(map[string]interface{}); ok {
		details = d
	}

	sessionID := s.getSessionID(params)
	session := s.getOrCreateSession(sessionID)

	// Track progress in session
	progressKey := "_workflow_progress"
	progress, _ := session.Context[progressKey].([]interface{})
	
	progress = append(progress, map[string]interface{}{
		"step":      stepName,
		"status":    status,
		"details":   details,
		"timestamp": time.Now(),
	})
	
	session.Context[progressKey] = progress
	s.sessions.Update(session)

	return map[string]interface{}{
		"status":      "tracked",
		"step":        stepName,
		"step_status": status,
		"total_steps": len(progress),
	}, nil
}

func (s *Server) handleWorkflowRecordFinding(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	findingType, ok := params["finding_type"].(string)
	if !ok || findingType == "" {
		return nil, fmt.Errorf("finding_type is required")
	}

	description, ok := params["description"].(string)
	if !ok || description == "" {
		return nil, fmt.Errorf("description is required")
	}

	severity := "medium"
	if s, ok := params["severity"].(string); ok {
		severity = s
	}

	evidence := make(map[string]interface{})
	if e, ok := params["evidence"].(map[string]interface{}); ok {
		evidence = e
	}

	relatedFindings := []string{}
	if rf, ok := params["related_findings"].([]interface{}); ok {
		for _, f := range rf {
			if fStr, ok := f.(string); ok {
				relatedFindings = append(relatedFindings, fStr)
			}
		}
	}

	finding := Finding{
		ID:          fmt.Sprintf("finding_%d_%s", time.Now().UnixNano(), findingType),
		Type:        FindingType(findingType),
		Severity:    FindingSeverity(severity),
		Description: description,
		Evidence:    evidence,
		Source:      "workflow",
		Timestamp:   time.Now(),
		Related:     relatedFindings,
	}

	sessionID := s.getSessionID(params)
	session := s.getOrCreateSession(sessionID)

	// Store finding in session
	findingsKey := "_workflow_findings"
	findings, _ := session.Context[findingsKey].([]Finding)
	findings = append(findings, finding)
	session.Context[findingsKey] = findings
	s.sessions.Update(session)

	return map[string]interface{}{
		"finding_id":     finding.ID,
		"status":         "recorded",
		"total_findings": len(findings),
	}, nil
}

func (s *Server) handleWorkflowGetFindings(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	sessionID := s.getSessionID(params)
	session, exists := s.sessions.Get(sessionID)
	if !exists {
		return map[string]interface{}{
			"findings": []Finding{},
			"count":    0,
		}, nil
	}

	findingsKey := "_workflow_findings"
	findings, _ := session.Context[findingsKey].([]Finding)

	// Apply filters
	filtered := []Finding{}
	filterType := ""
	if ft, ok := params["finding_type"].(string); ok {
		filterType = ft
	}

	minSeverity := ""
	if ms, ok := params["min_severity"].(string); ok {
		minSeverity = ms
	}

	severityOrder := map[string]int{
		"critical": 5,
		"high":     4,
		"medium":   3,
		"low":      2,
		"info":     1,
	}

	minSeverityLevel := 0
	if minSeverity != "" {
		minSeverityLevel = severityOrder[minSeverity]
	}

	for _, finding := range findings {
		// Type filter
		if filterType != "" && string(finding.Type) != filterType {
			continue
		}

		// Severity filter
		if minSeverityLevel > 0 {
			findingSeverityLevel := severityOrder[string(finding.Severity)]
			if findingSeverityLevel < minSeverityLevel {
				continue
			}
		}

		filtered = append(filtered, finding)
	}

	return map[string]interface{}{
		"findings": filtered,
		"count":    len(filtered),
		"filters": map[string]interface{}{
			"finding_type":  filterType,
			"min_severity": minSeverity,
		},
	}, nil
}

// Helper functions

func (s *Server) getSessionID(params map[string]interface{}) string {
	if sessionID, ok := params["session_id"].(string); ok && sessionID != "" {
		return sessionID
	}
	return "default"
}

func (s *Server) getOrCreateSession(sessionID string) *Session {
	session, exists := s.sessions.Get(sessionID)
	if !exists {
		session = s.sessions.Create()
		session.ID = sessionID
		s.sessions.Update(session)
	}
	return session
}

// WorkflowSuggestions provides step suggestions for scenarios
type WorkflowSuggestions struct {
	Description string
	Steps       []WorkflowStep
	Tips        []string
}

type WorkflowStep struct {
	Order       int
	Tool        string
	Description string
	Parameters  map[string]interface{}
	Optional    bool
	Condition   string
}

func (s *Server) getWorkflowSuggestions(scenario string, context map[string]interface{}) WorkflowSuggestions {
	switch scenario {
	case "incident_investigation":
		return WorkflowSuggestions{
			Description: "Systematic investigation of incidents to find root cause",
			Steps: []WorkflowStep{
				{
					Order:       1,
					Tool:        "discovery.explore_event_types",
					Description: "Discover what data is available",
					Parameters:  map[string]interface{}{"time_range": "2 hours"},
				},
				{
					Order:       2,
					Tool:        "nrql.execute",
					Description: "Check error rates and anomalies",
					Parameters:  map[string]interface{}{"query": "SELECT percentage(count(*), WHERE error IS true) FROM Transaction"},
					Condition:   "If Transaction events exist",
				},
				{
					Order:       3,
					Tool:        "discovery.find_relationships",
					Description: "Find related entities and dependencies",
					Optional:    true,
				},
				{
					Order:       4,
					Tool:        "analysis.find_anomalies",
					Description: "Detect anomalous patterns",
					Condition:   "If metrics show unusual patterns",
				},
				{
					Order:       5,
					Tool:        "workflow.record_finding",
					Description: "Record root cause findings",
				},
			},
			Tips: []string{
				"Start broad, then narrow down based on findings",
				"Look for correlations across different data types",
				"Check both application and infrastructure metrics",
				"Consider time-based patterns",
			},
		}

	case "performance_optimization":
		return WorkflowSuggestions{
			Description: "Analyze and optimize application performance",
			Steps: []WorkflowStep{
				{
					Order:       1,
					Tool:        "discovery.profile_data_completeness",
					Description: "Assess data quality for performance metrics",
				},
				{
					Order:       2,
					Tool:        "nrql.execute",
					Description: "Baseline current performance metrics",
					Parameters:  map[string]interface{}{"query": "SELECT percentile(duration, 50, 90, 95, 99) FROM Transaction"},
				},
				{
					Order:       3,
					Tool:        "analysis.detect_trends",
					Description: "Identify performance trends",
				},
				{
					Order:       4,
					Tool:        "nrql.execute",
					Description: "Find slow transactions",
					Parameters:  map[string]interface{}{"query": "SELECT * FROM Transaction WHERE duration > 1 LIMIT 100"},
				},
			},
			Tips: []string{
				"Focus on p95/p99 latencies, not just averages",
				"Look for patterns by time of day",
				"Check database query performance",
				"Analyze by transaction type",
			},
		}

	default:
		return WorkflowSuggestions{
			Description: "Generic investigation workflow",
			Steps: []WorkflowStep{
				{
					Order:       1,
					Tool:        "discovery.explore_event_types",
					Description: "Discover available data",
				},
				{
					Order:       2,
					Tool:        "nrql.execute",
					Description: "Query relevant metrics",
				},
				{
					Order:       3,
					Tool:        "workflow.record_finding",
					Description: "Record findings",
				},
			},
			Tips: []string{
				"Adapt steps based on what you discover",
				"Use context storage to share data between steps",
			},
		}
	}
}
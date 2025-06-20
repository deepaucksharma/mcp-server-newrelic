package mcp

import (
	"context"
	"fmt"
	"time"
)

// Workflow Management Tools - Granular Implementation

// WorkflowState represents the current state of a workflow execution
type WorkflowState struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"` // running, paused, completed, failed
	CurrentStep string                 `json:"current_step"`
	Context     map[string]interface{} `json:"context"`
	StartTime   time.Time              `json:"start_time"`
	Steps       []StepExecution        `json:"steps"`
}

// StepExecution tracks individual step execution
type StepExecution struct {
	StepID    string                 `json:"step_id"`
	Tool      string                 `json:"tool"`
	Status    string                 `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Input     map[string]interface{} `json:"input"`
	Output    interface{}            `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// registerWorkflowTools registers all workflow management tools
func (s *Server) registerWorkflowTools() error {
	// 1. Workflow Creation and Management
	if err := s.registerWorkflowCreationTool(); err != nil {
		return err
	}
	
	// 2. Step Execution Tools
	if err := s.registerStepExecutionTools(); err != nil {
		return err
	}
	
	// 3. Context Management Tools
	if err := s.registerContextTools(); err != nil {
		return err
	}
	
	// 4. Investigation Workflow Tools
	if err := s.registerInvestigationTools(); err != nil {
		return err
	}
	
	// 5. Incident Response Tools
	if err := s.registerIncidentTools(); err != nil {
		return err
	}
	
	// 6. Analysis Pattern Tools
	if err := s.registerAnalysisTools(); err != nil {
		return err
	}
	
	return nil
}

// 1. WORKFLOW CREATION AND MANAGEMENT

func (s *Server) registerWorkflowCreationTool() error {
	tool := NewToolBuilder("workflow.create", "Create a new workflow execution context").
		Category(CategoryUtility).
		Handler(s.handleWorkflowCreate).
		Required("name", "workflow_type").
		Param("name", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Human-readable workflow name",
			},
			Examples: []interface{}{
				"Investigate checkout service latency",
				"Black Friday capacity planning",
			},
		}).
		Param("workflow_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Type of workflow to execute",
				Enum:        []string{"investigation", "incident_response", "capacity_planning", "slo_management", "optimization"},
			},
		}).
		Param("context", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Initial context data for the workflow",
			},
			Examples: []interface{}{
				map[string]interface{}{
					"entity_guid": "ABC123",
					"time_range":  "last 6 hours",
					"severity":    "critical",
				},
			},
		}).
		Param("auto_execute", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Automatically start executing workflow steps",
				Default:     false,
			},
		}).
		Safety(func(s *SafetyMetadata) {
			s.Level = SafetyLevelSafe
			s.IsDestructive = false
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 50
			p.Cacheable = false
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Start investigation: workflow.create(name='Investigate API latency', workflow_type='investigation', context={entity_name: 'checkout-api'})",
				"Incident response: workflow.create(name='P1 Incident #123', workflow_type='incident_response', context={alert_id: '123', severity: 'critical'})",
			}
			g.ChainsWith = []string{"workflow.execute_step", "context.add_finding"}
			g.SuccessIndicators = []string{"workflow_id returned", "status is 'created' or 'running'"}
		}).
		Example(ToolExample{
			Name:        "Start performance investigation",
			Description: "Create workflow to investigate performance degradation",
			Params: map[string]interface{}{
				"name":          "Investigate checkout slowness",
				"workflow_type": "investigation",
				"context": map[string]interface{}{
					"entity_name": "checkout-service",
					"symptom":     "p95 latency increased 300%",
					"time_range":  "last 2 hours",
				},
			},
		}).
		Build()

	return s.tools.Register(tool.Tool)
}

// 2. STEP EXECUTION TOOLS

func (s *Server) registerStepExecutionTools() error {
	// Execute next workflow step
	executeStep := NewToolBuilder("workflow.execute_step", "Execute the next step in a workflow").
		Category(CategoryUtility).
		Handler(s.handleWorkflowExecuteStep).
		Required("workflow_id").
		Param("workflow_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "ID of the workflow to advance",
			},
		}).
		Param("step_override", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Specific step to execute (skips to this step)",
			},
		}).
		Param("inputs", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Additional inputs for the step",
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 100
			p.MaxLatencyMS = 30000 // Some steps may take time
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Continue workflow: workflow.execute_step(workflow_id='wf_123')",
				"Skip to step: workflow.execute_step(workflow_id='wf_123', step_override='root_cause_analysis')",
			}
			g.ChainsWith = []string{"workflow.get_state", "context.add_finding"}
			g.WarningsForAI = []string{
				"Check workflow state before executing steps",
				"Handle step failures gracefully",
			}
		}).
		Build()

	if err := s.tools.Register(executeStep.Tool); err != nil {
		return err
	}

	// Get workflow state
	getState := NewToolBuilder("workflow.get_state", "Get current workflow state and progress").
		Category(CategoryQuery).
		Handler(s.handleWorkflowGetState).
		Required("workflow_id").
		Param("workflow_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Workflow ID to query",
			},
		}).
		Param("include_outputs", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Include step outputs in response",
				Default:     true,
			},
		}).
		Build()

	return s.tools.Register(getState.Tool)
}

// 3. CONTEXT MANAGEMENT TOOLS

func (s *Server) registerContextTools() error {
	// Add finding to context
	addFinding := NewToolBuilder("context.add_finding", "Add a finding to workflow context").
		Category(CategoryUtility).
		Handler(s.handleContextAddFinding).
		Required("workflow_id", "finding").
		Param("workflow_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Workflow ID",
			},
		}).
		Param("finding", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Finding to add to context",
			},
			Examples: []interface{}{
				map[string]interface{}{
					"type":        "anomaly",
					"severity":    "high",
					"description": "CPU spike detected at 14:30",
					"evidence": map[string]interface{}{
						"metric": "cpuPercent",
						"value":  95.5,
						"normal": 45.0,
					},
				},
			},
		}).
		Param("source_tool", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Tool that generated this finding",
			},
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Add anomaly: context.add_finding(workflow_id='wf_123', finding={type: 'anomaly', severity: 'high', description: 'Spike detected'})",
			}
			g.ChainsWith = []string{"workflow.execute_step", "context.get_recommendations"}
		}).
		Build()

	if err := s.tools.Register(addFinding.Tool); err != nil {
		return err
	}

	// Get recommendations based on context
	getRecommendations := NewToolBuilder("context.get_recommendations", "Get AI recommendations based on workflow context").
		Category(CategoryAnalysis).
		Handler(s.handleContextGetRecommendations).
		Required("workflow_id").
		Param("workflow_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Workflow ID",
			},
		}).
		Param("recommendation_type", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Type of recommendations needed",
				Enum:        []string{"next_steps", "root_cause", "mitigation", "optimization"},
				Default:     "next_steps",
			},
		}).
		Build()

	return s.tools.Register(getRecommendations.Tool)
}

// 4. INVESTIGATION WORKFLOW TOOLS

func (s *Server) registerInvestigationTools() error {
	// Identify investigation scope
	identifyScope := NewToolBuilder("investigate.identify_scope", "Identify the scope of investigation based on symptoms").
		Category(CategoryAnalysis).
		Handler(s.handleInvestigateIdentifyScope).
		Required("symptoms").
		Param("symptoms", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "List of observed symptoms",
				Items: &Property{
					Type: "string",
				},
			},
			Examples: []interface{}{
				[]string{"high latency", "increased errors", "cpu spike"},
			},
		}).
		Param("entity_hints", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Entity names or GUIDs that might be involved",
				Items: &Property{
					Type: "string",
				},
			},
		}).
		Param("time_range", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Time range when symptoms were observed",
				Default:     "last 1 hour",
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 500
			p.Cacheable = true
			p.CacheTTLSeconds = 300
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Start investigation: investigate.identify_scope(symptoms=['high latency', 'timeout errors'], entity_hints=['checkout-api'])",
			}
			g.ChainsWith = []string{"entity.get_golden_metrics", "analysis.detect_anomalies"}
			g.SuccessIndicators = []string{"Returns entities to investigate", "Provides investigation priority"}
		}).
		Example(ToolExample{
			Name:        "Investigate performance issue",
			Description: "Identify scope for latency investigation",
			Params: map[string]interface{}{
				"symptoms":     []string{"p95 latency > 1s", "timeout errors increasing"},
				"entity_hints": []string{"payment-service", "checkout-api"},
				"time_range":   "last 30 minutes",
			},
		}).
		Build()

	if err := s.tools.Register(identifyScope.Tool); err != nil {
		return err
	}

	// Find correlated anomalies
	findCorrelations := NewToolBuilder("investigate.find_correlations", "Find correlations between different signals").
		Category(CategoryAnalysis).
		Handler(s.handleInvestigateFindCorrelations).
		Required("primary_signal", "candidate_signals").
		Param("primary_signal", EnhancedProperty{
			Property: Property{
				Type:        "object",
				Description: "Primary signal to correlate against",
			},
			Examples: []interface{}{
				map[string]interface{}{
					"query":       "SELECT average(duration) FROM Transaction WHERE appName = 'checkout'",
					"metric_name": "transaction.duration",
				},
			},
		}).
		Param("candidate_signals", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "List of signals to check correlation",
				Items: &Property{
					Type: "object",
				},
			},
		}).
		Param("correlation_window", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Time window for correlation analysis",
				Default:     "5 minutes",
			},
		}).
		Param("min_correlation", EnhancedProperty{
			Property: Property{
				Type:        "number",
				Description: "Minimum correlation coefficient (0-1)",
				Default:     0.7,
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 2000
			p.MaxLatencyMS = 10000
			p.ResourceIntensive = true
		}).
		Build()

	return s.tools.Register(findCorrelations.Tool)
}

// 5. INCIDENT RESPONSE TOOLS

func (s *Server) registerIncidentTools() error {
	// Get incident context
	getIncidentContext := NewToolBuilder("incident.get_context", "Get comprehensive context for an incident").
		Category(CategoryQuery).
		Handler(s.handleIncidentGetContext).
		Required("incident_id").
		Param("incident_id", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Incident or alert ID",
			},
		}).
		Param("include_history", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Include historical incidents on same entity",
				Default:     true,
			},
		}).
		Param("include_dependencies", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Include dependent service status",
				Default:     true,
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 300
			p.Cacheable = true
			p.CacheTTLSeconds = 60
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Get incident details: incident.get_context(incident_id='INC-123', include_dependencies=true)",
			}
			g.ChainsWith = []string{"impact.assess_user_impact", "runbook.get_steps"}
			g.SuccessIndicators = []string{"Returns incident details", "Includes affected entities"}
		}).
		Build()

	if err := s.tools.Register(getIncidentContext.Tool); err != nil {
		return err
	}

	// Assess user impact
	assessImpact := NewToolBuilder("impact.assess_user_impact", "Assess the user impact of an incident").
		Category(CategoryAnalysis).
		Handler(s.handleImpactAssessUser).
		Required("entity_guid").
		Param("entity_guid", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Entity experiencing the incident",
			},
		}).
		Param("time_range", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Time range to assess impact",
				Default:     "since incident started",
			},
		}).
		Param("impact_metrics", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Specific metrics to check",
				Items: &Property{
					Type: "string",
				},
				Default: []string{"error_rate", "response_time", "throughput"},
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 1000
			p.Cacheable = true
			p.CacheTTLSeconds = 180
		}).
		Build()

	return s.tools.Register(assessImpact.Tool)
}

// 6. ANALYSIS PATTERN TOOLS

func (s *Server) registerAnalysisTools() error {
	// Detect patterns in data
	detectPatterns := NewToolBuilder("pattern.detect", "Detect patterns in time series data").
		Category(CategoryAnalysis).
		Handler(s.handlePatternDetect).
		Required("query", "pattern_types").
		Param("query", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "NRQL query for time series data",
			},
		}).
		Param("pattern_types", EnhancedProperty{
			Property: Property{
				Type:        "array",
				Description: "Types of patterns to detect",
				Items: &Property{
					Type: "string",
					Enum: []string{"anomaly", "trend", "seasonality", "changepoint", "correlation"},
				},
			},
		}).
		Param("sensitivity", EnhancedProperty{
			Property: Property{
				Type:        "number",
				Description: "Detection sensitivity (0-1)",
				Default:     0.8,
			},
		}).
		Param("lookback_window", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Historical data for pattern learning",
				Default:     "7 days",
			},
		}).
		Performance(func(p *PerformanceMetadata) {
			p.ExpectedLatencyMS = 3000
			p.MaxLatencyMS = 15000
			p.ResourceIntensive = true
			p.CostCategory = "high"
		}).
		AIGuidance(func(g *AIGuidanceMetadata) {
			g.UsageExamples = []string{
				"Detect anomalies: pattern.detect(query='SELECT average(duration) FROM Transaction', pattern_types=['anomaly', 'changepoint'])",
			}
			g.ChainsWith = []string{"pattern.explain", "alert.create_from_pattern"}
			g.WarningsForAI = []string{
				"Resource intensive - use specific time ranges",
				"May require multiple iterations with different sensitivity",
			}
		}).
		Build()

	if err := s.tools.Register(detectPatterns.Tool); err != nil {
		return err
	}

	// Forecast based on patterns
	forecast := NewToolBuilder("pattern.forecast", "Forecast future values based on historical patterns").
		Category(CategoryAnalysis).
		Handler(s.handlePatternForecast).
		Required("query", "forecast_window").
		Param("query", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "NRQL query for historical data",
			},
		}).
		Param("forecast_window", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "How far to forecast (e.g., '24 hours', '7 days')",
			},
		}).
		Param("confidence_level", EnhancedProperty{
			Property: Property{
				Type:        "number",
				Description: "Confidence level for prediction bands",
				Default:     0.95,
			},
		}).
		Param("include_seasonality", EnhancedProperty{
			Property: Property{
				Type:        "boolean",
				Description: "Account for seasonal patterns",
				Default:     true,
			},
		}).
		Param("growth_model", EnhancedProperty{
			Property: Property{
				Type:        "string",
				Description: "Growth model to use",
				Enum:        []string{"linear", "logistic", "exponential"},
				Default:     "linear",
			},
		}).
		Build()

	return s.tools.Register(forecast.Tool)
}

// Handler implementations (simplified for demonstration)

func (s *Server) handleWorkflowCreate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, _ := params["name"].(string)
	workflowType, _ := params["workflow_type"].(string)
	context, _ := params["context"].(map[string]interface{})
	autoExecute, _ := params["auto_execute"].(bool)

	// Create workflow state
	workflow := &WorkflowState{
		ID:        fmt.Sprintf("wf_%d", time.Now().UnixNano()),
		Name:      name,
		Status:    "created",
		Context:   context,
		StartTime: time.Now(),
		Steps:     []StepExecution{},
	}

	// Initialize workflow based on type
	switch workflowType {
	case "investigation":
		workflow.Steps = s.getInvestigationSteps(context)
	case "incident_response":
		workflow.Steps = s.getIncidentResponseSteps(context)
	case "capacity_planning":
		workflow.Steps = s.getCapacityPlanningSteps(context)
	default:
		return nil, fmt.Errorf("unknown workflow type: %s", workflowType)
	}

	if autoExecute {
		workflow.Status = "running"
		workflow.CurrentStep = workflow.Steps[0].StepID
	}

	// Store workflow state (in real implementation)
	// s.stateManager.StoreWorkflow(workflow)

	return map[string]interface{}{
		"workflow_id": workflow.ID,
		"status":      workflow.Status,
		"next_step":   workflow.CurrentStep,
		"total_steps": len(workflow.Steps),
	}, nil
}

func (s *Server) handleWorkflowExecuteStep(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	workflowID, _ := params["workflow_id"].(string)
	stepOverride, _ := params["step_override"].(string)
	inputs, _ := params["inputs"].(map[string]interface{})

	// Mock execution
	result := map[string]interface{}{
		"workflow_id": workflowID,
		"step_executed": stepOverride,
		"status": "completed",
		"output": map[string]interface{}{
			"findings": []map[string]interface{}{
				{
					"type": "anomaly",
					"description": "CPU spike detected",
					"severity": "high",
				},
			},
		},
		"next_step": "analyze_correlations",
		"progress": "3/10",
	}

	return result, nil
}

func (s *Server) handleInvestigateIdentifyScope(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	symptoms, _ := params["symptoms"].([]interface{})
	entityHints, _ := params["entity_hints"].([]interface{})
	timeRange, _ := params["time_range"].(string)

	// Mock scope identification
	scope := map[string]interface{}{
		"primary_entities": []map[string]interface{}{
			{
				"guid": "ENTITY123",
				"name": "checkout-api",
				"type": "APPLICATION",
				"relevance_score": 0.95,
				"symptoms_matched": []string{"high latency", "timeout errors"},
			},
		},
		"related_entities": []map[string]interface{}{
			{
				"guid": "ENTITY456",
				"name": "payment-service",
				"type": "APPLICATION",
				"relationship": "dependency",
				"relevance_score": 0.75,
			},
		},
		"investigation_priority": []string{
			"Check checkout-api golden signals",
			"Analyze payment-service interactions",
			"Review recent deployments",
			"Check infrastructure metrics",
		},
		"time_windows": map[string]interface{}{
			"incident_start": "2024-01-20T14:30:00Z",
			"peak_impact": "2024-01-20T14:45:00Z",
			"suggested_analysis_window": timeRange,
		},
	}

	return scope, nil
}

// Helper methods

func (s *Server) getInvestigationSteps(context map[string]interface{}) []StepExecution {
	return []StepExecution{
		{StepID: "identify_scope", Tool: "investigate.identify_scope", Status: "pending"},
		{StepID: "gather_metrics", Tool: "entity.get_golden_metrics", Status: "pending"},
		{StepID: "detect_anomalies", Tool: "pattern.detect", Status: "pending"},
		{StepID: "find_correlations", Tool: "investigate.find_correlations", Status: "pending"},
		{StepID: "analyze_logs", Tool: "logs.search_errors", Status: "pending"},
		{StepID: "trace_requests", Tool: "traces.find_slow", Status: "pending"},
		{StepID: "check_changes", Tool: "deployment.find_recent", Status: "pending"},
		{StepID: "identify_root_cause", Tool: "analysis.determine_root_cause", Status: "pending"},
		{StepID: "recommend_actions", Tool: "context.get_recommendations", Status: "pending"},
		{StepID: "generate_report", Tool: "report.generate_investigation", Status: "pending"},
	}
}

func (s *Server) getIncidentResponseSteps(context map[string]interface{}) []StepExecution {
	return []StepExecution{
		{StepID: "get_context", Tool: "incident.get_context", Status: "pending"},
		{StepID: "assess_impact", Tool: "impact.assess_user_impact", Status: "pending"},
		{StepID: "find_dependencies", Tool: "entity.get_relationships", Status: "pending"},
		{StepID: "get_runbook", Tool: "runbook.get_steps", Status: "pending"},
		{StepID: "execute_mitigation", Tool: "action.execute_runbook", Status: "pending"},
		{StepID: "verify_resolution", Tool: "slo.check_recovery", Status: "pending"},
		{StepID: "update_stakeholders", Tool: "notification.send_update", Status: "pending"},
		{StepID: "create_postmortem", Tool: "postmortem.create_draft", Status: "pending"},
	}
}

func (s *Server) getCapacityPlanningSteps(context map[string]interface{}) []StepExecution {
	return []StepExecution{
		{StepID: "analyze_historical", Tool: "pattern.analyze_historical", Status: "pending"},
		{StepID: "detect_seasonality", Tool: "pattern.detect", Status: "pending"},
		{StepID: "calculate_growth", Tool: "analysis.calculate_growth_rate", Status: "pending"},
		{StepID: "forecast_load", Tool: "pattern.forecast", Status: "pending"},
		{StepID: "identify_bottlenecks", Tool: "analysis.find_bottlenecks", Status: "pending"},
		{StepID: "calculate_resources", Tool: "capacity.calculate_requirements", Status: "pending"},
		{StepID: "estimate_costs", Tool: "cost.project_monthly", Status: "pending"},
		{StepID: "generate_recommendations", Tool: "capacity.generate_scaling_plan", Status: "pending"},
		{StepID: "create_dashboards", Tool: "dashboard.create_capacity_monitoring", Status: "pending"},
		{StepID: "setup_alerts", Tool: "alert.create_capacity_thresholds", Status: "pending"},
	}
}
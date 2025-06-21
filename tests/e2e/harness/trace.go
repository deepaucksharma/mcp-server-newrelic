package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// TraceCollector collects trace information from the MCP server
type TraceCollector struct {
	baseURL    string
	httpClient *http.Client
	outputDir  string
}

// NewTraceCollector creates a new trace collector
func NewTraceCollector(region string) *TraceCollector {
	baseURL := os.Getenv("MCP_SERVER_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &TraceCollector{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		outputDir: "tests/results/traces",
	}
}

// CollectTrace retrieves trace data for a given trace ID
func (t *TraceCollector) CollectTrace(ctx context.Context, traceID string) (*TraceData, error) {
	// Call /explain/{traceID} endpoint
	url := fmt.Sprintf("%s/explain/%s", t.baseURL, traceID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trace: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("trace endpoint returned %d: %s", resp.StatusCode, body)
	}
	
	// Parse trace response
	var traceResponse TraceResponse
	if err := json.NewDecoder(resp.Body).Decode(&traceResponse); err != nil {
		return nil, fmt.Errorf("failed to decode trace response: %w", err)
	}
	
	// Convert to TraceData
	traceData := &TraceData{
		TraceID:    traceID,
		Confidence: traceResponse.Confidence,
		Metadata:   traceResponse.Metadata,
	}
	
	// Extract discovery steps
	for _, step := range traceResponse.Steps {
		if isDiscoveryStep(step) {
			traceData.DiscoverySteps = append(traceData.DiscoverySteps, DiscoveryStep{
				Tool:      step.Tool,
				Timestamp: step.Timestamp,
				Duration:  step.Duration,
				Result:    step.Result,
			})
		}
	}
	
	// Save trace HTML if available
	if traceResponse.HTML != "" {
		t.saveTraceHTML(traceID, traceResponse.HTML)
	}
	
	return traceData, nil
}

// saveTraceHTML saves the trace HTML to disk
func (t *TraceCollector) saveTraceHTML(traceID, html string) error {
	if err := os.MkdirAll(t.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create trace directory: %w", err)
	}
	
	filename := filepath.Join(t.outputDir, fmt.Sprintf("trace-%s.html", traceID))
	return os.WriteFile(filename, []byte(html), 0644)
}

// CollectArtifacts collects additional artifacts for a failed scenario
func (t *TraceCollector) CollectArtifacts(ctx context.Context, scenario ScenarioResult) []string {
	artifacts := []string{}
	
	// Collect container logs
	if logPath := t.collectContainerLogs(scenario.ID); logPath != "" {
		artifacts = append(artifacts, logPath)
	}
	
	// Collect JSON-RPC requests/responses
	if reqPath := t.collectRequests(scenario.ID); reqPath != "" {
		artifacts = append(artifacts, reqPath)
	}
	
	// Collect performance metrics
	if metricsPath := t.collectMetrics(scenario.ID); metricsPath != "" {
		artifacts = append(artifacts, metricsPath)
	}
	
	return artifacts
}

// collectContainerLogs collects container logs for debugging
func (t *TraceCollector) collectContainerLogs(scenarioID string) string {
	// Implementation depends on container runtime
	// For Docker:
	// docker logs mcp-server-{scenarioID} > logs/container-{scenarioID}.log
	
	logDir := filepath.Join("tests/results/logs", scenarioID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return ""
	}
	
	// Collect MCP server logs
	serverLogPath := filepath.Join(logDir, "mcp-server.log")
	// ... implementation to fetch logs
	
	return serverLogPath
}

// collectRequests collects JSON-RPC traffic
func (t *TraceCollector) collectRequests(scenarioID string) string {
	// If traffic capture was enabled, save the captured requests
	trafficDir := filepath.Join("tests/results/traffic", scenarioID)
	if err := os.MkdirAll(trafficDir, 0755); err != nil {
		return ""
	}
	
	requestsPath := filepath.Join(trafficDir, "requests.json")
	// ... implementation to save captured traffic
	
	return requestsPath
}

// collectMetrics collects performance metrics
func (t *TraceCollector) collectMetrics(scenarioID string) string {
	metricsDir := filepath.Join("tests/results/metrics", scenarioID)
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return ""
	}
	
	metricsPath := filepath.Join(metricsDir, "performance.json")
	// ... implementation to collect metrics
	
	return metricsPath
}

// TraceResponse represents the response from the trace endpoint
type TraceResponse struct {
	TraceID    string                   `json:"trace_id"`
	Confidence float64                  `json:"confidence"`
	Steps      []TraceStep              `json:"steps"`
	Metadata   map[string]interface{}   `json:"metadata"`
	HTML       string                   `json:"html,omitempty"`
}

// TraceStep represents a single step in the trace
type TraceStep struct {
	Tool      string                 `json:"tool"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Result    map[string]interface{} `json:"result"`
	Error     string                 `json:"error,omitempty"`
}

// Helper functions

func isDiscoveryStep(step TraceStep) bool {
	discoveryTools := []string{
		"discovery.explore_event_types",
		"discovery.explore_attributes",
		"discovery.profile_attribute",
		"discovery.find_relationships",
		"discovery.assess_quality",
		"discovery.find_patterns",
		"discovery.profile_data_completeness",
	}
	
	for _, tool := range discoveryTools {
		if step.Tool == tool {
			return true
		}
	}
	
	return false
}

// MetricsCollector collects performance metrics during tests
type MetricsCollector struct {
	metrics    []PerformanceMetric
	startTime  time.Time
}

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	Timestamp   time.Time     `json:"timestamp"`
	ScenarioID  string        `json:"scenario_id"`
	StepIndex   int           `json:"step_index"`
	Tool        string        `json:"tool"`
	Duration    time.Duration `json:"duration"`
	MemoryUsage int64         `json:"memory_usage_bytes"`
	CPUPercent  float64       `json:"cpu_percent"`
	Error       bool          `json:"error"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:   make([]PerformanceMetric, 0),
		startTime: time.Now(),
	}
}

// Record records a performance metric
func (m *MetricsCollector) Record(metric PerformanceMetric) {
	metric.Timestamp = time.Now()
	m.metrics = append(m.metrics, metric)
}

// Save saves collected metrics to a file
func (m *MetricsCollector) Save(path string) error {
	data, err := json.MarshalIndent(m.metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	
	return os.WriteFile(path, data, 0644)
}

// GetSummary returns a summary of collected metrics
func (m *MetricsCollector) GetSummary() map[string]interface{} {
	if len(m.metrics) == 0 {
		return nil
	}
	
	var totalDuration time.Duration
	var maxMemory int64
	var avgCPU float64
	errorCount := 0
	
	for _, metric := range m.metrics {
		totalDuration += metric.Duration
		if metric.MemoryUsage > maxMemory {
			maxMemory = metric.MemoryUsage
		}
		avgCPU += metric.CPUPercent
		if metric.Error {
			errorCount++
		}
	}
	
	return map[string]interface{}{
		"total_metrics":     len(m.metrics),
		"total_duration":    totalDuration,
		"avg_duration":      totalDuration / time.Duration(len(m.metrics)),
		"max_memory_mb":     maxMemory / 1024 / 1024,
		"avg_cpu_percent":   avgCPU / float64(len(m.metrics)),
		"error_count":       errorCount,
		"error_rate":        float64(errorCount) / float64(len(m.metrics)) * 100,
		"collection_duration": time.Since(m.startTime),
	}
}
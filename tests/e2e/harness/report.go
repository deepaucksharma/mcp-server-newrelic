package harness

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TestReport represents the complete E2E test report
type TestReport struct {
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Duration     time.Duration     `json:"duration"`
	Region       string            `json:"region"`
	Environment  string            `json:"environment"`
	Scenarios    []ScenarioResult  `json:"scenarios"`
	Summary      ReportSummary     `json:"summary"`
}

// ScenarioResult contains the result of a single scenario
type ScenarioResult struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Status     TestStatus         `json:"status"`
	StartTime  time.Time          `json:"start_time"`
	EndTime    time.Time          `json:"end_time"`
	Duration   time.Duration      `json:"duration"`
	Region     string             `json:"region"`
	Steps      []StepResult       `json:"steps"`
	Assertions []AssertionResult  `json:"assertions"`
	TraceID    string             `json:"trace_id,omitempty"`
	TraceData  *TraceData         `json:"trace_data,omitempty"`
	Error      error              `json:"error,omitempty"`
	Artifacts  []string           `json:"artifacts,omitempty"`
}

// TestStatus represents the status of a test
type TestStatus string

const (
	StatusPassed   TestStatus = "passed"
	StatusFailed   TestStatus = "failed"
	StatusSkipped  TestStatus = "skipped"
	StatusUnstable TestStatus = "unstable"
)

// ReportSummary contains summary statistics
type ReportSummary struct {
	TotalScenarios   int           `json:"total_scenarios"`
	PassedScenarios  int           `json:"passed_scenarios"`
	FailedScenarios  int           `json:"failed_scenarios"`
	SkippedScenarios int           `json:"skipped_scenarios"`
	PassRate         float64       `json:"pass_rate"`
	TotalDuration    time.Duration `json:"total_duration"`
	AverageDuration  time.Duration `json:"average_duration"`
	Flakiness        float64       `json:"flakiness"`
}

// TraceData contains trace information
type TraceData struct {
	TraceID        string                 `json:"trace_id"`
	Confidence     float64                `json:"confidence"`
	DiscoverySteps []DiscoveryStep        `json:"discovery_steps"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// DiscoveryStep represents a discovery operation in the trace
type DiscoveryStep struct {
	Tool      string                 `json:"tool"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Result    map[string]interface{} `json:"result"`
}

// calculateStats calculates summary statistics
func (r *TestReport) calculateStats() {
	r.Summary.TotalScenarios = len(r.Scenarios)
	
	for _, scenario := range r.Scenarios {
		switch scenario.Status {
		case StatusPassed:
			r.Summary.PassedScenarios++
		case StatusFailed:
			r.Summary.FailedScenarios++
		case StatusSkipped:
			r.Summary.SkippedScenarios++
		}
		r.Summary.TotalDuration += scenario.Duration
	}
	
	if r.Summary.TotalScenarios > 0 {
		r.Summary.PassRate = float64(r.Summary.PassedScenarios) / float64(r.Summary.TotalScenarios) * 100
		r.Summary.AverageDuration = r.Summary.TotalDuration / time.Duration(r.Summary.TotalScenarios)
	}
}

// SaveJSON saves the report as JSON
func (r *TestReport) SaveJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	return os.WriteFile(path, data, 0644)
}

// SaveJUnit saves the report in JUnit XML format
func (r *TestReport) SaveJUnit(path string) error {
	junit := r.toJUnit()
	
	data, err := xml.MarshalIndent(junit, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JUnit: %w", err)
	}
	
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	return os.WriteFile(path, data, 0644)
}

// SaveHTML saves the report as HTML
func (r *TestReport) SaveHTML(path string) error {
	tmpl := template.Must(template.New("report").Parse(htmlReportTemplate))
	
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	return tmpl.Execute(file, r)
}

// toJUnit converts the report to JUnit format
func (r *TestReport) toJUnit() *JUnitTestSuites {
	suites := &JUnitTestSuites{
		Name:     "E2E Test Suite",
		Tests:    r.Summary.TotalScenarios,
		Failures: r.Summary.FailedScenarios,
		Time:     r.Duration.Seconds(),
	}
	
	suite := JUnitTestSuite{
		Name:      fmt.Sprintf("E2E-%s", r.Region),
		Tests:     len(r.Scenarios),
		Failures:  r.Summary.FailedScenarios,
		Time:      r.Duration.Seconds(),
		Timestamp: r.StartTime.Format(time.RFC3339),
	}
	
	for _, scenario := range r.Scenarios {
		testCase := JUnitTestCase{
			Name:      scenario.Title,
			ClassName: scenario.ID,
			Time:      scenario.Duration.Seconds(),
		}
		
		if scenario.Status == StatusFailed && scenario.Error != nil {
			testCase.Failure = &JUnitFailure{
				Type:    "AssertionError",
				Message: scenario.Error.Error(),
			}
		}
		
		if scenario.Status == StatusSkipped {
			testCase.Skipped = &JUnitSkipped{
				Message: "Test skipped",
			}
		}
		
		suite.TestCases = append(suite.TestCases, testCase)
	}
	
	suites.TestSuites = append(suites.TestSuites, suite)
	return suites
}

// JUnit XML structures

type JUnitTestSuites struct {
	XMLName    xml.Name          `xml:"testsuites"`
	Name       string            `xml:"name,attr"`
	Tests      int               `xml:"tests,attr"`
	Failures   int               `xml:"failures,attr"`
	Time       float64           `xml:"time,attr"`
	TestSuites []JUnitTestSuite  `xml:"testsuite"`
}

type JUnitTestSuite struct {
	Name      string           `xml:"name,attr"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Time      float64          `xml:"time,attr"`
	Timestamp string           `xml:"timestamp,attr"`
	TestCases []JUnitTestCase  `xml:"testcase"`
}

type JUnitTestCase struct {
	Name      string          `xml:"name,attr"`
	ClassName string          `xml:"classname,attr"`
	Time      float64         `xml:"time,attr"`
	Failure   *JUnitFailure   `xml:"failure,omitempty"`
	Skipped   *JUnitSkipped   `xml:"skipped,omitempty"`
}

type JUnitFailure struct {
	Type    string `xml:"type,attr"`
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

type JUnitSkipped struct {
	Message string `xml:"message,attr"`
}

// ReportWriter handles report output
type ReportWriter struct {
	outputDir string
	formats   []string
}

// NewReportWriter creates a new report writer
func NewReportWriter(outputDir string, formats []string) *ReportWriter {
	return &ReportWriter{
		outputDir: outputDir,
		formats:   formats,
	}
}

// Write writes the report in all configured formats
func (w *ReportWriter) Write(report *TestReport) error {
	timestamp := time.Now().Format("20060102-150405")
	
	for _, format := range w.formats {
		var path string
		var err error
		
		switch format {
		case "json":
			path = filepath.Join(w.outputDir, fmt.Sprintf("e2e-report-%s.json", timestamp))
			err = report.SaveJSON(path)
			
		case "junit":
			path = filepath.Join(w.outputDir, fmt.Sprintf("e2e-report-%s.xml", timestamp))
			err = report.SaveJUnit(path)
			
		case "html":
			path = filepath.Join(w.outputDir, fmt.Sprintf("e2e-report-%s.html", timestamp))
			err = report.SaveHTML(path)
			
		default:
			continue
		}
		
		if err != nil {
			return fmt.Errorf("failed to write %s report: %w", format, err)
		}
		
		fmt.Printf("Report written to: %s\n", path)
	}
	
	return nil
}

// PrintSummary prints a summary to the given writer
func (r *TestReport) PrintSummary(w io.Writer) {
	fmt.Fprintf(w, "\n=== E2E Test Summary ===\n")
	fmt.Fprintf(w, "Region: %s\n", r.Region)
	fmt.Fprintf(w, "Duration: %s\n", r.Duration.Round(time.Second))
	fmt.Fprintf(w, "Scenarios: %d total, %d passed, %d failed, %d skipped\n",
		r.Summary.TotalScenarios,
		r.Summary.PassedScenarios,
		r.Summary.FailedScenarios,
		r.Summary.SkippedScenarios)
	fmt.Fprintf(w, "Pass Rate: %.1f%%\n", r.Summary.PassRate)
	
	if r.Summary.FailedScenarios > 0 {
		fmt.Fprintf(w, "\nFailed Scenarios:\n")
		for _, scenario := range r.Scenarios {
			if scenario.Status == StatusFailed {
				fmt.Fprintf(w, "  - %s (%s): %v\n", scenario.ID, scenario.Title, scenario.Error)
			}
		}
	}
	
	fmt.Fprintln(w, "")
}

// HTML report template
const htmlReportTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>E2E Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .scenario { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .passed { background: #d4edda; }
        .failed { background: #f8d7da; }
        .skipped { background: #fff3cd; }
        .step { margin: 10px 20px; padding: 5px; background: #f8f9fa; }
        .assertion { margin: 5px 20px; }
        .error { color: #dc3545; }
        .success { color: #28a745; }
    </style>
</head>
<body>
    <div class="header">
        <h1>E2E Test Report</h1>
        <p>Region: {{.Region}} | Duration: {{.Duration}} | Generated: {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
    </div>
    
    <div class="summary">
        <h2>Summary</h2>
        <p>Total Scenarios: {{.Summary.TotalScenarios}}</p>
        <p>Passed: {{.Summary.PassedScenarios}} | Failed: {{.Summary.FailedScenarios}} | Skipped: {{.Summary.SkippedScenarios}}</p>
        <p>Pass Rate: {{printf "%.1f" .Summary.PassRate}}%</p>
    </div>
    
    <div class="scenarios">
        <h2>Scenarios</h2>
        {{range .Scenarios}}
        <div class="scenario {{.Status}}">
            <h3>{{.ID}}: {{.Title}}</h3>
            <p>Status: {{.Status}} | Duration: {{.Duration}}</p>
            {{if .Error}}<p class="error">Error: {{.Error}}</p>{{end}}
            
            <h4>Steps:</h4>
            {{range .Steps}}
            <div class="step">
                <strong>{{.Tool}}</strong> ({{.Duration}})
                {{if .Error}}<span class="error">Failed: {{.Error}}</span>{{end}}
            </div>
            {{end}}
            
            <h4>Assertions:</h4>
            {{range .Assertions}}
            <div class="assertion">
                <span class="{{if .Passed}}success{{else}}error{{end}}">
                    {{if .Passed}}✓{{else}}✗{{end}} {{.Type}}: {{.Expected}} {{.Operator}} {{.Actual}}
                </span>
                {{if .Message}}<br>{{.Message}}{{end}}
            </div>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>`
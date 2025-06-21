package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/tests/e2e/harness"
	"github.com/joho/godotenv"
)

func main() {
	var (
		scenarioPath = flag.String("scenario", "", "Path to scenario file or directory")
		tag          = flag.String("tag", "", "Run scenarios with this tag")
		reportFormat = flag.String("format", "json", "Report format: json, junit, html")
		outputDir    = flag.String("output", "reports", "Output directory for reports")
		parallel     = flag.Int("parallel", 2, "Number of parallel test executions")
		timeout      = flag.Duration("timeout", 5*time.Minute, "Overall test timeout")
	)
	flag.Parse()

	// Load test environment
	if err := godotenv.Load(".env.test"); err != nil {
		log.Printf("Warning: could not load .env.test: %v", err)
	}

	// Ensure we have required environment variables
	primaryKey := os.Getenv("E2E_PRIMARY_API_KEY")
	primaryAccount := os.Getenv("E2E_PRIMARY_ACCOUNT_ID")
	if primaryKey == "" || primaryAccount == "" {
		// Try alternative names
		primaryKey = os.Getenv("NEW_RELIC_API_KEY_PRIMARY")
		primaryAccount = os.Getenv("NEW_RELIC_ACCOUNT_ID_PRIMARY")
		if primaryKey == "" || primaryAccount == "" {
			log.Fatal("E2E_PRIMARY_API_KEY and E2E_PRIMARY_ACCOUNT_ID must be set")
		}
		// Set the expected env vars
		os.Setenv("E2E_PRIMARY_API_KEY", primaryKey)
		os.Setenv("E2E_PRIMARY_ACCOUNT_ID", primaryAccount)
	}

	// Create runner configuration
	config := harness.RunnerConfig{
		MaxParallel:      *parallel,
		Timeout:          *timeout,
		RetryAttempts:    3,
		CaptureTraffic:   os.Getenv("E2E_CAPTURE_TRAFFIC") == "true",
		SaveResponses:    os.Getenv("E2E_SAVE_RESPONSES") == "true",
		CleanupTestData:  os.Getenv("E2E_CLEANUP_TEST_DATA") != "false",
		OutputDir:        *outputDir,
		MCPServerURL:     os.Getenv("MCP_SERVER_URL"),
		MCPServerCommand: os.Getenv("MCP_SERVER_COMMAND"),
	}

	// If no MCP server URL or command, use default
	if config.MCPServerURL == "" && config.MCPServerCommand == "" {
		// Try to find the built binary
		config.MCPServerCommand = "./bin/mcp-server"
		if _, err := os.Stat(config.MCPServerCommand); os.IsNotExist(err) {
			config.MCPServerCommand = "../../bin/mcp-server"
		}
	}

	// Create runner
	runner := harness.NewScenarioRunner(config)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Find scenarios to run
	var scenarios []string
	if *scenarioPath != "" {
		info, err := os.Stat(*scenarioPath)
		if err != nil {
			log.Fatalf("Failed to stat scenario path: %v", err)
		}
		if info.IsDir() {
			// Find all YAML files in directory
			files, err := filepath.Glob(filepath.Join(*scenarioPath, "*.yaml"))
			if err != nil {
				log.Fatalf("Failed to list scenarios: %v", err)
			}
			scenarios = files
		} else {
			scenarios = []string{*scenarioPath}
		}
	} else if *tag != "" {
		// Find scenarios with tag
		files, err := filepath.Glob("scenarios/*.yaml")
		if err != nil {
			log.Fatalf("Failed to list scenarios: %v", err)
		}
		for _, file := range files {
			// Parse to check tags
			scenario, err := runner.ParseScenario(file)
			if err != nil {
				log.Printf("Warning: failed to parse %s: %v", file, err)
				continue
			}
			for _, t := range scenario.Tags {
				if t == *tag {
					scenarios = append(scenarios, file)
					break
				}
			}
		}
	} else {
		// Run all critical scenarios by default
		*tag = "critical"
		files, err := filepath.Glob("scenarios/*.yaml")
		if err != nil {
			log.Fatalf("Failed to list scenarios: %v", err)
		}
		for _, file := range files {
			scenario, err := runner.ParseScenario(file)
			if err != nil {
				continue
			}
			for _, t := range scenario.Tags {
				if t == "critical" {
					scenarios = append(scenarios, file)
					break
				}
			}
		}
	}

	if len(scenarios) == 0 {
		log.Fatal("No scenarios found to run")
	}

	log.Printf("Running %d scenarios...", len(scenarios))

	// Run scenarios
	report, err := runner.RunScenarios(ctx, scenarios)
	if err != nil {
		log.Printf("Error running scenarios: %v", err)
	}

	// Generate report
	var reportPath string
	switch *reportFormat {
	case "junit":
		reportPath = filepath.Join(*outputDir, "junit.xml")
		err = report.WriteJUnit(reportPath)
	case "html":
		reportPath = filepath.Join(*outputDir, "report.html")
		err = report.WriteHTML(reportPath)
	default:
		reportPath = filepath.Join(*outputDir, "report.json")
		err = report.WriteJSON(reportPath)
	}

	if err != nil {
		log.Fatalf("Failed to write report: %v", err)
	}

	// Print summary
	fmt.Printf("\nTest Summary:\n")
	fmt.Printf("Total Scenarios: %d\n", len(report.Scenarios))
	fmt.Printf("Passed: %d\n", countByStatus(report.Scenarios, harness.StatusPassed))
	fmt.Printf("Failed: %d\n", countByStatus(report.Scenarios, harness.StatusFailed))
	fmt.Printf("Skipped: %d\n", countByStatus(report.Scenarios, harness.StatusSkipped))
	fmt.Printf("\nReport written to: %s\n", reportPath)

	// Exit with error if any tests failed
	if countByStatus(report.Scenarios, harness.StatusFailed) > 0 {
		os.Exit(1)
	}
}

func countByStatus(results []harness.ScenarioResult, status harness.TestStatus) int {
	count := 0
	for _, r := range results {
		if r.Status == status {
			count++
		}
	}
	return count
}
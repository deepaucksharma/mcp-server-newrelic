package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/config"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/newrelic"
	"github.com/joho/godotenv"
)

type Issue struct {
	Category    string
	Severity    string // "error", "warning", "info"
	Description string
	Solution    string
	AutoFix     bool
}

type Diagnostic struct {
	issues  []Issue
	autoFix bool
}

func main() {
	// Parse flags
	var (
		autoFix = flag.Bool("fix", false, "Automatically fix issues where possible")
		envFile = flag.String("env", ".env", "Environment file to check")
	)
	flag.Parse()

	fmt.Println("New Relic MCP Server Diagnostic Tool")
	fmt.Println("====================================")
	fmt.Println()

	diag := &Diagnostic{
		autoFix: *autoFix,
		issues:  []Issue{},
	}

	// Run all checks
	diag.checkEnvironment()
	diag.checkDirectoryStructure()
	diag.checkConfiguration(*envFile)
	diag.checkDependencies()
	diag.checkNewRelicConnection()
	diag.checkBuildSystem()

	// Report results
	diag.reportResults()

	// Apply fixes if requested
	if *autoFix && len(diag.issues) > 0 {
		diag.applyFixes()
	}

	// Exit with appropriate code
	hasErrors := false
	for _, issue := range diag.issues {
		if issue.Severity == "error" {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func (d *Diagnostic) checkEnvironment() {
	fmt.Println("Checking environment...")

	// Check Go version
	goVersion := runtime.Version()
	if !strings.HasPrefix(goVersion, "go1.21") && !strings.HasPrefix(goVersion, "go1.22") {
		d.addIssue(Issue{
			Category:    "Environment",
			Severity:    "warning",
			Description: fmt.Sprintf("Go version %s detected. Recommended: Go 1.21+", goVersion),
			Solution:    "Update Go to version 1.21 or later",
			AutoFix:     false,
		})
	} else {
		fmt.Printf("✓ Go version: %s\n", goVersion)
	}

	// Check GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		d.addIssue(Issue{
			Category:    "Environment",
			Severity:    "warning",
			Description: "GOPATH not set",
			Solution:    "Set GOPATH environment variable",
			AutoFix:     false,
		})
	}

	// Check PATH includes Go bin
	path := os.Getenv("PATH")
	if !strings.Contains(path, filepath.Join(gopath, "bin")) && gopath != "" {
		d.addIssue(Issue{
			Category:    "Environment",
			Severity:    "info",
			Description: "GOPATH/bin not in PATH",
			Solution:    fmt.Sprintf("Add %s to PATH", filepath.Join(gopath, "bin")),
			AutoFix:     false,
		})
	}
}

func (d *Diagnostic) checkDirectoryStructure() {
	fmt.Println("\nChecking directory structure...")

	requiredDirs := []string{
		"cmd",
		"pkg",
		"pkg/discovery",
		"pkg/interface",
		"pkg/interface/mcp",
		"pkg/interface/api",
		"pkg/config",
		"pkg/state",
		"pkg/newrelic",
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			d.addIssue(Issue{
				Category:    "Directory Structure",
				Severity:    "error",
				Description: fmt.Sprintf("Required directory missing: %s", dir),
				Solution:    fmt.Sprintf("Create directory: mkdir -p %s", dir),
				AutoFix:     true,
			})
		} else {
			fmt.Printf("✓ Directory exists: %s\n", dir)
		}
	}
}

func (d *Diagnostic) checkConfiguration(envFile string) {
	fmt.Println("\nChecking configuration...")

	// Check if .env file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// Check for .env.example
		if _, err := os.Stat(".env.example"); err == nil {
			d.addIssue(Issue{
				Category:    "Configuration",
				Severity:    "error",
				Description: fmt.Sprintf("Environment file %s not found", envFile),
				Solution:    "Copy .env.example to .env and configure",
				AutoFix:     true,
			})
		} else {
			d.addIssue(Issue{
				Category:    "Configuration",
				Severity:    "error",
				Description: "No environment configuration found",
				Solution:    "Create .env file with required configuration",
				AutoFix:     false,
			})
		}
	} else {
		fmt.Printf("✓ Environment file exists: %s\n", envFile)

		// Load and validate configuration
		if err := godotenv.Load(envFile); err != nil {
			d.addIssue(Issue{
				Category:    "Configuration",
				Severity:    "error",
				Description: fmt.Sprintf("Failed to load %s: %v", envFile, err),
				Solution:    "Fix syntax errors in environment file",
				AutoFix:     false,
			})
		} else {
			// Check required variables
			requiredVars := []string{
				"NEW_RELIC_API_KEY",
				"NEW_RELIC_ACCOUNT_ID",
				"JWT_SECRET",
				"API_KEY_SALT",
			}

			for _, varName := range requiredVars {
				value := os.Getenv(varName)
				if value == "" {
					d.addIssue(Issue{
						Category:    "Configuration",
						Severity:    "error",
						Description: fmt.Sprintf("Required environment variable %s not set", varName),
						Solution:    fmt.Sprintf("Set %s in %s", varName, envFile),
						AutoFix:     false,
					})
				} else if strings.Contains(value, "your-") || strings.Contains(value, "change-me") {
					d.addIssue(Issue{
						Category:    "Configuration",
						Severity:    "error",
						Description: fmt.Sprintf("%s contains placeholder value", varName),
						Solution:    fmt.Sprintf("Replace %s with actual value", varName),
						AutoFix:     false,
					})
				} else {
					// Mask sensitive values
					masked := value[:min(4, len(value))] + "****"
					fmt.Printf("✓ %s is set: %s\n", varName, masked)
				}
			}
		}
	}
}

func (d *Diagnostic) checkDependencies() {
	fmt.Println("\nChecking dependencies...")

	// Check if go.mod exists
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		d.addIssue(Issue{
			Category:    "Dependencies",
			Severity:    "error",
			Description: "go.mod file not found",
			Solution:    "Run: go mod init github.com/deepaucksharma/mcp-server-newrelic",
			AutoFix:     true,
		})
	} else {
		fmt.Println("✓ go.mod exists")

		// Check if dependencies are downloaded
		cmd := exec.Command("go", "mod", "verify")
		if err := cmd.Run(); err != nil {
			d.addIssue(Issue{
				Category:    "Dependencies",
				Severity:    "warning",
				Description: "Go module dependencies not verified",
				Solution:    "Run: go mod download",
				AutoFix:     true,
			})
		} else {
			fmt.Println("✓ Go dependencies verified")
		}
	}

	// Check for required tools
	tools := map[string]string{
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
		"protoc":        "Install protobuf compiler from https://github.com/protocolbuffers/protobuf/releases",
	}

	for tool, installCmd := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			d.addIssue(Issue{
				Category:    "Dependencies",
				Severity:    "info",
				Description: fmt.Sprintf("Development tool %s not found", tool),
				Solution:    installCmd,
				AutoFix:     false,
			})
		} else {
			fmt.Printf("✓ Tool available: %s\n", tool)
		}
	}
}

func (d *Diagnostic) checkNewRelicConnection() {
	fmt.Println("\nChecking New Relic connection...")

	// Skip if in mock mode
	if os.Getenv("MOCK_MODE") == "true" {
		fmt.Println("✓ Mock mode enabled, skipping New Relic connection check")
		return
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		d.addIssue(Issue{
			Category:    "New Relic",
			Severity:    "error",
			Description: fmt.Sprintf("Failed to load configuration: %v", err),
			Solution:    "Fix configuration errors",
			AutoFix:     false,
		})
		return
	}

	// Try to connect
	client, err := newrelic.NewClient(newrelic.Config{
		APIKey:    cfg.NewRelic.APIKey,
		AccountID: cfg.NewRelic.AccountID,
		Region:    cfg.NewRelic.Region,
		Timeout:   10 * time.Second,
	})
	if err != nil {
		d.addIssue(Issue{
			Category:    "New Relic",
			Severity:    "error",
			Description: fmt.Sprintf("Failed to create New Relic client: %v", err),
			Solution:    "Check API key and account configuration",
			AutoFix:     false,
		})
		return
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := client.GetAccountInfo(ctx); err != nil {
		d.addIssue(Issue{
			Category:    "New Relic",
			Severity:    "error",
			Description: fmt.Sprintf("Failed to connect to New Relic: %v", err),
			Solution:    "Verify API key, account ID, and network connectivity",
			AutoFix:     false,
		})
	} else {
		fmt.Printf("✓ Successfully connected to New Relic account %s\n", cfg.NewRelic.AccountID)
	}
}

func (d *Diagnostic) checkBuildSystem() {
	fmt.Println("\nChecking build system...")

	// Try to build the main binary
	cmd := exec.Command("go", "build", "-o", "/tmp/mcp-server-test", "./cmd/mcp-server")
	if output, err := cmd.CombinedOutput(); err != nil {
		d.addIssue(Issue{
			Category:    "Build",
			Severity:    "error",
			Description: fmt.Sprintf("Failed to build MCP server: %v", err),
			Solution:    fmt.Sprintf("Fix build errors:\n%s", string(output)),
			AutoFix:     false,
		})
	} else {
		fmt.Println("✓ MCP server builds successfully")
		os.Remove("/tmp/mcp-server-test")
	}

	// Check if Makefile exists
	if _, err := os.Stat("Makefile"); os.IsNotExist(err) {
		d.addIssue(Issue{
			Category:    "Build",
			Severity:    "warning",
			Description: "Makefile not found",
			Solution:    "Create Makefile for common build tasks",
			AutoFix:     false,
		})
	} else {
		fmt.Println("✓ Makefile exists")
	}
}

func (d *Diagnostic) addIssue(issue Issue) {
	d.issues = append(d.issues, issue)
}

func (d *Diagnostic) reportResults() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("DIAGNOSTIC RESULTS")
	fmt.Println(strings.Repeat("=", 50))

	if len(d.issues) == 0 {
		fmt.Println("\n✅ All checks passed! Your environment is ready.")
		return
	}

	// Group by severity
	errors := []Issue{}
	warnings := []Issue{}
	infos := []Issue{}

	for _, issue := range d.issues {
		switch issue.Severity {
		case "error":
			errors = append(errors, issue)
		case "warning":
			warnings = append(warnings, issue)
		case "info":
			infos = append(infos, issue)
		}
	}

	// Report errors
	if len(errors) > 0 {
		fmt.Printf("\n❌ ERRORS (%d)\n", len(errors))
		for i, issue := range errors {
			fmt.Printf("\n%d. [%s] %s\n", i+1, issue.Category, issue.Description)
			fmt.Printf("   Solution: %s\n", issue.Solution)
			if issue.AutoFix {
				fmt.Println("   ✨ Can be auto-fixed")
			}
		}
	}

	// Report warnings
	if len(warnings) > 0 {
		fmt.Printf("\n⚠️  WARNINGS (%d)\n", len(warnings))
		for i, issue := range warnings {
			fmt.Printf("\n%d. [%s] %s\n", i+1, issue.Category, issue.Description)
			fmt.Printf("   Solution: %s\n", issue.Solution)
			if issue.AutoFix {
				fmt.Println("   ✨ Can be auto-fixed")
			}
		}
	}

	// Report info
	if len(infos) > 0 {
		fmt.Printf("\nℹ️  INFO (%d)\n", len(infos))
		for i, issue := range infos {
			fmt.Printf("\n%d. [%s] %s\n", i+1, issue.Category, issue.Description)
			fmt.Printf("   Solution: %s\n", issue.Solution)
		}
	}

	// Summary
	fmt.Printf("\nSummary: %d errors, %d warnings, %d info\n", len(errors), len(warnings), len(infos))

	if d.autoFix {
		fixableCount := 0
		for _, issue := range d.issues {
			if issue.AutoFix {
				fixableCount++
			}
		}
		if fixableCount > 0 {
			fmt.Printf("\n%d issues can be auto-fixed.\n", fixableCount)
		}
	} else {
		fmt.Println("\nRun with --fix to automatically fix issues where possible.")
	}
}

func (d *Diagnostic) applyFixes() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("APPLYING FIXES")
	fmt.Println(strings.Repeat("=", 50))

	fixedCount := 0
	for _, issue := range d.issues {
		if !issue.AutoFix {
			continue
		}

		fmt.Printf("\nFixing: %s\n", issue.Description)

		switch {
		case strings.Contains(issue.Description, "Required directory missing"):
			// Create missing directory
			dir := strings.TrimPrefix(issue.Description, "Required directory missing: ")
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("❌ Failed to create directory: %v\n", err)
			} else {
				fmt.Printf("✓ Created directory: %s\n", dir)
				fixedCount++
			}

		case strings.Contains(issue.Description, "Environment file .env not found"):
			// Copy .env.example to .env
			if err := copyFile(".env.example", ".env"); err != nil {
				fmt.Printf("❌ Failed to copy .env.example: %v\n", err)
			} else {
				fmt.Println("✓ Created .env from .env.example")
				fmt.Println("  ⚠️  Remember to update the configuration values!")
				fixedCount++
			}

		case strings.Contains(issue.Description, "go.mod file not found"):
			// Initialize go module
			cmd := exec.Command("go", "mod", "init", "github.com/deepaucksharma/mcp-server-newrelic")
			if err := cmd.Run(); err != nil {
				fmt.Printf("❌ Failed to initialize go module: %v\n", err)
			} else {
				fmt.Println("✓ Initialized go module")
				fixedCount++
			}

		case strings.Contains(issue.Description, "Go module dependencies not verified"):
			// Download dependencies
			cmd := exec.Command("go", "mod", "download")
			if err := cmd.Run(); err != nil {
				fmt.Printf("❌ Failed to download dependencies: %v\n", err)
			} else {
				fmt.Println("✓ Downloaded Go dependencies")
				fixedCount++
			}
		}
	}

	fmt.Printf("\n✅ Fixed %d issues\n", fixedCount)
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
.PHONY: build test run clean docker-build docker-run lint format help

# Default target
default: build

# Build the application
build: build-mcp build-api build-cli build-diagnose

build-mcp:
	@echo "Building MCP server..."
	@go build -o bin/mcp-server ./cmd/mcp-server

build-api:
	@echo "Building API server..."
	@go build -o bin/api-server ./cmd/api-server

build-cli:
	@echo "Building UDS CLI..."
	@go build -o bin/uds ./cmd/uds

build-diagnose:
	@echo "Building diagnostic tool..."
	@go build -o bin/diagnose ./cmd/diagnose

# Run tests
test:
	@./test.sh all

test-unit:
	@./test.sh unit

test-integration:
	@./test.sh integration

test-benchmark:
	@./test.sh benchmark

test-coverage:
	@./test.sh coverage

# Run the application
run: run-mcp

run-mcp: build-mcp
	@echo "Starting MCP server..."
	@./bin/mcp-server

run-api: build-api
	@echo "Starting API server..."
	@./bin/api-server

run-mock: build-mcp
	@echo "Starting MCP server in mock mode..."
	@./bin/mcp-server --mock

diagnose: build-diagnose
	@echo "Running diagnostics..."
	@./bin/diagnose

diagnose-fix: build-diagnose
	@echo "Running diagnostics with auto-fix..."
	@./bin/diagnose --fix

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -testcache

# Docker operations
docker-build:
	@echo "Building Docker image..."
	@docker build -t mcp-server-newrelic .

docker-run:
	@echo "Running with Docker Compose..."
	@docker-compose up

# Code quality
lint:
	@echo "Running linter..."
	@golangci-lint run
	@./scripts/assumption_scan.sh

format:
	@echo "Formatting code..."
	@go fmt ./...

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build          - Build all components"
	@echo "  build-mcp      - Build MCP server"
	@echo "  build-api      - Build API server"
	@echo "  build-cli      - Build UDS CLI"
	@echo "  build-diagnose - Build diagnostic tool"
	@echo ""
	@echo "Run targets:"
	@echo "  run            - Run MCP server (default)"
	@echo "  run-mcp        - Run MCP server"
	@echo "  run-api        - Run API server"
	@echo "  run-mock       - Run MCP server in mock mode"
	@echo ""
	@echo "Development targets:"
	@echo "  diagnose       - Run system diagnostics"
	@echo "  diagnose-fix   - Run diagnostics with auto-fix"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-benchmark - Run benchmarks"
	@echo "  test-coverage  - Generate coverage report"
	@echo "  lint           - Run linter"
	@echo "  format         - Format code"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo ""
	@echo "  help           - Show this help"
	@echo ""
	@echo "E2E Test targets:"
	@echo "  test-e2e       - Run all E2E tests (requires .env.test)"
	@echo "  test-e2e-setup - Setup E2E test environment"
	@echo "  help-e2e       - Show all E2E test targets"

# Include E2E test targets
-include Makefile.e2e

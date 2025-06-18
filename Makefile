.PHONY: build test run clean docker-build docker-run lint format help

# Default target
default: build

# Build the application
build:
	@echo "Building API server..."
	@go build -o bin/api-server ./cmd/api-server

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
run: build
	@echo "Starting API server..."
	@./bin/api-server

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

format:
	@echo "Formatting code..."
	@go fmt ./...

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the MCP server"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-benchmark - Run benchmarks"
	@echo "  test-coverage  - Generate coverage report"
	@echo "  run            - Build and run the server"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  lint           - Run linter"
	@echo "  format         - Format code"
	@echo "  help           - Show this help"

#!/bin/bash

# Unified test runner for New Relic MCP Server

set -e

case "$1" in
  unit)
    echo "Running unit tests..."
    go test ./... -v -short -race
    ;;
  integration)
    echo "Running integration tests..."
    go test ./tests/integration/... -v
    ;;
  benchmark)
    echo "Running benchmarks..."
    go test -bench=. ./benchmarks/... -benchmem
    ;;
  coverage)
    echo "Running tests with coverage..."
    go test ./... -v -race -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report generated: coverage.html"
    ;;
  security)
    echo "Running security checks..."
    go test ./pkg/security/... -v
    ;;
  all)
    echo "Running all tests..."
    go test ./... -v -race
    ;;
  *)
    echo "Usage: $0 {unit|integration|benchmark|coverage|security|all}"
    exit 1
    ;;
esac

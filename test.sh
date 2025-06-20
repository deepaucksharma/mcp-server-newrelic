#!/bin/bash

# Test script for New Relic MCP Server
# This script runs all tests and generates coverage reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running New Relic MCP Server Tests...${NC}"

# Run tests with race detector and coverage
echo -e "\n${YELLOW}Running unit tests with coverage...${NC}"
go test -race -coverprofile=coverage.out -covermode=atomic ./pkg/... ./cmd/... ./internal/... 2>&1 | tee test.log

# Check if tests passed
if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo -e "${RED}Tests failed! See test.log for details.${NC}"
    exit 1
fi

# Generate coverage report
echo -e "\n${YELLOW}Generating coverage report...${NC}"
go tool cover -html=coverage.out -o coverage.html

# Calculate coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "\n${GREEN}Total coverage: ${COVERAGE}${NC}"

# Parse coverage percentage
COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
COVERAGE_INT=${COVERAGE_NUM%.*}

# Check if coverage meets minimum threshold
MIN_COVERAGE=40
if [ "$COVERAGE_INT" -lt "$MIN_COVERAGE" ]; then
    echo -e "${RED}Coverage ${COVERAGE} is below minimum threshold of ${MIN_COVERAGE}%${NC}"
    exit 1
fi

# Run specific test suites
echo -e "\n${YELLOW}Running MCP interface tests...${NC}"
go test -v ./pkg/interface/mcp/...

echo -e "\n${YELLOW}Running discovery engine tests...${NC}"
go test -v ./pkg/discovery/...

echo -e "\n${YELLOW}Running state management tests...${NC}"
go test -v ./pkg/state/...

# Check for race conditions
echo -e "\n${YELLOW}Checking for race conditions...${NC}"
go test -race ./...

# Run benchmarks if requested
if [ "$1" == "bench" ]; then
    echo -e "\n${YELLOW}Running benchmarks...${NC}"
    go test -bench=. -benchmem ./...
fi

echo -e "\n${GREEN}All tests passed successfully!${NC}"
echo -e "Coverage report generated: coverage.html"
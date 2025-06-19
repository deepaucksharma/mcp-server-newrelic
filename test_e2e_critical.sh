#!/bin/bash

set -e

echo "================================"
echo "End-to-End Critical Fixes Test"
echo "================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Load environment
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# Build the test
echo -e "\n${GREEN}Building test binary...${NC}"
go build -o test_critical test_critical_fixes.go

# Run the test
echo -e "\n${GREEN}Running critical fixes tests...${NC}"
./test_critical

# Run unit tests for specific components
echo -e "\n${GREEN}Running unit tests for critical components...${NC}"

# Test validation package
echo -e "\n${GREEN}Testing NRQL validation...${NC}"
go test -v -race ./pkg/validation/... -run TestNRQLValidator

# Test state package (race conditions)
echo -e "\n${GREEN}Testing state management with race detector...${NC}"
go test -v -race ./pkg/state/... -run "TestMemory|TestRedis"

# Test utils package (panic recovery)
echo -e "\n${GREEN}Testing panic recovery...${NC}"
go test -v -race ./pkg/utils/... -run TestRecovery

# Test pattern detection memory limits
echo -e "\n${GREEN}Testing pattern detection memory management...${NC}"
go test -v ./pkg/discovery/patterns/... -run TestMemory

# Integration test with MCP server
echo -e "\n${GREEN}Running MCP server integration test...${NC}"

# Start MCP server in background
echo "Starting MCP server..."
./bin/mcp-server -transport http -port 8999 &
MCP_PID=$!

# Wait for server to start
sleep 2

# Test MCP endpoints with sanitization
echo -e "\n${GREEN}Testing MCP endpoints...${NC}"

# Test 1: Valid NRQL query
echo "Test 1: Valid NRQL query"
curl -X POST http://localhost:8999/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "query-nrdb",
      "arguments": {
        "query": "SELECT count(*) FROM Transaction WHERE appName = '\''myapp'\'' SINCE 1 hour ago"
      }
    },
    "id": 1
  }' || true

echo -e "\n"

# Test 2: SQL injection attempt (should be rejected)
echo "Test 2: SQL injection attempt (should be rejected)"
curl -X POST http://localhost:8999/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "query-nrdb",
      "arguments": {
        "query": "SELECT * FROM Transaction; DROP TABLE users; --"
      }
    },
    "id": 2
  }' || true

echo -e "\n"

# Test 3: Health check
echo "Test 3: Health check"
curl http://localhost:8999/health || true

echo -e "\n"

# Kill MCP server
kill $MCP_PID 2>/dev/null || true

# Clean up
rm -f test_critical

echo -e "\n${GREEN}================================${NC}"
echo -e "${GREEN}All tests completed!${NC}"
echo -e "${GREEN}================================${NC}"

# Run race detection on full test suite
echo -e "\n${GREEN}Running full test suite with race detection...${NC}"
go test -race -timeout 30s ./pkg/...
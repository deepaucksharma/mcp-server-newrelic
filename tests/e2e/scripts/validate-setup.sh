#!/bin/bash
# E2E Testing Setup Validation Script

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}E2E Testing Setup Validator${NC}"
echo "================================"

# Check if we're in the right directory
if [ ! -f "Makefile.e2e" ]; then
    echo -e "${RED}✗ Error: Must run from project root directory${NC}"
    exit 1
fi

# Check for .env.test file
echo -e "\n${YELLOW}Checking environment configuration...${NC}"
if [ -f ".env.test" ]; then
    echo -e "${GREEN}✓ .env.test file exists${NC}"
    
    # Check required variables
    source .env.test
    
    if [ -z "${NEW_RELIC_API_KEY_PRIMARY:-}" ]; then
        echo -e "${RED}✗ NEW_RELIC_API_KEY_PRIMARY is not set${NC}"
        exit 1
    else
        echo -e "${GREEN}✓ Primary API key is configured${NC}"
    fi
    
    if [ -z "${NEW_RELIC_ACCOUNT_ID_PRIMARY:-}" ]; then
        echo -e "${RED}✗ NEW_RELIC_ACCOUNT_ID_PRIMARY is not set${NC}"
        exit 1
    else
        echo -e "${GREEN}✓ Primary account ID is configured${NC}"
    fi
else
    echo -e "${RED}✗ .env.test file not found${NC}"
    echo -e "${YELLOW}  Creating from template...${NC}"
    cp .env.test.example .env.test
    echo -e "${GREEN}✓ Created .env.test from template${NC}"
    echo -e "${YELLOW}  Please edit .env.test with your New Relic credentials${NC}"
    exit 1
fi

# Check Go version
echo -e "\n${YELLOW}Checking Go version...${NC}"
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
REQUIRED_VERSION="1.19"
if [[ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" == "$REQUIRED_VERSION" ]]; then
    echo -e "${GREEN}✓ Go version $GO_VERSION meets minimum requirement ($REQUIRED_VERSION)${NC}"
else
    echo -e "${RED}✗ Go version $GO_VERSION is below minimum requirement ($REQUIRED_VERSION)${NC}"
    exit 1
fi

# Check if MCP server can be built
echo -e "\n${YELLOW}Checking MCP server build...${NC}"
if go build -o /tmp/mcp-server-test ./cmd/mcp-server > /dev/null 2>&1; then
    echo -e "${GREEN}✓ MCP server builds successfully${NC}"
    rm -f /tmp/mcp-server-test
else
    echo -e "${RED}✗ Failed to build MCP server${NC}"
    exit 1
fi

# Check test directories
echo -e "\n${YELLOW}Checking test infrastructure...${NC}"
REQUIRED_DIRS=(
    "tests/e2e/harness"
    "tests/e2e/scenarios"
    "tests/e2e/framework"
    "tests/e2e/discovery"
    "tests/e2e/workflows"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo -e "${GREEN}✓ Directory $dir exists${NC}"
    else
        echo -e "${RED}✗ Directory $dir is missing${NC}"
        exit 1
    fi
done

# Check for test scenarios
echo -e "\n${YELLOW}Checking test scenarios...${NC}"
SCENARIO_COUNT=$(find tests/e2e/scenarios -name "*.yaml" | wc -l)
if [ "$SCENARIO_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Found $SCENARIO_COUNT test scenarios${NC}"
else
    echo -e "${RED}✗ No test scenarios found${NC}"
    exit 1
fi

# Check network connectivity (optional)
echo -e "\n${YELLOW}Checking network connectivity...${NC}"
if command -v curl &> /dev/null; then
    if curl -s -f https://api.newrelic.com/graphql > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Can reach New Relic API${NC}"
    else
        echo -e "${YELLOW}⚠ Cannot reach New Relic API (may be behind proxy)${NC}"
    fi
else
    echo -e "${YELLOW}⚠ curl not found, skipping network check${NC}"
fi

# Check if test results directory exists
echo -e "\n${YELLOW}Checking test results directory...${NC}"
RESULTS_DIR="${E2E_RESULTS_DIR:-tests/results}"
if [ ! -d "$RESULTS_DIR" ]; then
    mkdir -p "$RESULTS_DIR"
    echo -e "${GREEN}✓ Created results directory: $RESULTS_DIR${NC}"
else
    echo -e "${GREEN}✓ Results directory exists: $RESULTS_DIR${NC}"
fi

# Summary
echo -e "\n${BLUE}Setup Validation Complete${NC}"
echo "================================"
echo -e "${GREEN}✓ E2E testing environment is ready!${NC}"
echo ""
echo "Next steps:"
echo "1. Ensure your .env.test file has valid New Relic credentials"
echo "2. Run 'make test-e2e' to execute all E2E tests"
echo "3. Run 'make test-e2e-discovery' to test discovery tools only"
echo "4. Check results in $RESULTS_DIR"
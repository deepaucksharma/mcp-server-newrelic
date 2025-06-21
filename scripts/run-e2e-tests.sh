#!/bin/bash

# E2E Test Runner Script
# This script runs the complete E2E test suite with various options

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
TEST_SUITE="all"
TIMEOUT="30m"
VERBOSE=false
BENCHMARK=false
COVERAGE=false
REPORT_DIR="./test-reports"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or higher."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.21"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is too old. Please upgrade to Go $REQUIRED_VERSION or higher."
        exit 1
    fi
    
    # Check if .env.test exists
    if [ ! -f ".env.test" ]; then
        print_warning ".env.test not found. Creating from .env.example..."
        if [ -f ".env.example" ]; then
            cp .env.example .env.test
            print_warning "Please update .env.test with your New Relic credentials"
            exit 1
        else
            print_error ".env.example not found. Cannot create .env.test"
            exit 1
        fi
    fi
    
    # Check if MCP server binary exists
    if [ ! -f "bin/mcp-server" ]; then
        print_status "MCP server binary not found. Building..."
        make build-mcp
    fi
    
    print_status "Prerequisites check completed"
}

# Function to run tests
run_tests() {
    local suite=$1
    local test_name=""
    local output_file=""
    
    case $suite in
        "protocol")
            test_name="Protocol Compliance"
            output_file="protocol-test-results.log"
            TEST_PATTERN="TestMCPProtocolCompliance"
            ;;
        "discovery")
            test_name="Discovery"
            output_file="discovery-test-results.log"
            TEST_PATTERN="Test(DiscoveryChain|AdaptiveQueryBuilding)"
            ;;
        "performance")
            test_name="Performance"
            output_file="benchmark-results.log"
            TEST_PATTERN="TestMCPPerformanceBenchmarks"
            ;;
        "composable")
            test_name="Composable Tools"
            output_file="composable-test-results.log"
            TEST_PATTERN="TestComposableTools"
            ;;
        "caching")
            test_name="Caching"
            output_file="caching-test-results.log"
            TEST_PATTERN="TestCachingBehavior"
            ;;
        "all")
            test_name="All E2E"
            output_file="e2e-test-results.log"
            TEST_PATTERN=""
            ;;
        *)
            print_error "Unknown test suite: $suite"
            exit 1
            ;;
    esac
    
    print_status "Running $test_name tests..."
    
    # Build test command
    TEST_CMD="go test -v -timeout $TIMEOUT"
    
    if [ "$VERBOSE" = true ]; then
        TEST_CMD="$TEST_CMD -v"
    fi
    
    if [ "$COVERAGE" = true ]; then
        TEST_CMD="$TEST_CMD -coverprofile=${REPORT_DIR}/coverage-${suite}.out"
    fi
    
    if [ -n "$TEST_PATTERN" ]; then
        TEST_CMD="$TEST_CMD -run '^${TEST_PATTERN}$'"
    fi
    
    TEST_CMD="$TEST_CMD ./tests/e2e/..."
    
    # Run tests and capture output
    mkdir -p $REPORT_DIR
    
    if eval $TEST_CMD 2>&1 | tee "${REPORT_DIR}/${output_file}"; then
        print_status "$test_name tests passed!"
        return 0
    else
        print_error "$test_name tests failed!"
        return 1
    fi
}

# Function to generate reports
generate_reports() {
    print_status "Generating test reports..."
    
    # Install go-junit-report if not available
    if ! command -v go-junit-report &> /dev/null; then
        print_status "Installing go-junit-report..."
        go install github.com/jstemmer/go-junit-report/v2@latest
    fi
    
    # Generate JUnit reports
    for log_file in ${REPORT_DIR}/*-results.log; do
        if [ -f "$log_file" ]; then
            base_name=$(basename "$log_file" .log)
            cat "$log_file" | go-junit-report > "${REPORT_DIR}/${base_name}.xml"
            print_status "Generated JUnit report: ${base_name}.xml"
        fi
    done
    
    # Generate coverage report if available
    if [ "$COVERAGE" = true ]; then
        for cov_file in ${REPORT_DIR}/coverage-*.out; do
            if [ -f "$cov_file" ]; then
                base_name=$(basename "$cov_file" .out)
                go tool cover -html="$cov_file" -o "${REPORT_DIR}/${base_name}.html"
                print_status "Generated coverage report: ${base_name}.html"
            fi
        done
    fi
}

# Function to print usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Run E2E tests for MCP Server New Relic

OPTIONS:
    -s, --suite SUITE      Test suite to run (default: all)
                          Options: all, protocol, discovery, performance, composable, caching
    -t, --timeout TIMEOUT  Test timeout (default: 30m)
    -v, --verbose         Enable verbose output
    -b, --benchmark       Run performance benchmarks
    -c, --coverage        Generate coverage reports
    -r, --report-dir DIR  Directory for test reports (default: ./test-reports)
    -h, --help           Show this help message

EXAMPLES:
    # Run all tests
    $0

    # Run only protocol tests with coverage
    $0 --suite protocol --coverage

    # Run performance benchmarks
    $0 --suite performance --benchmark

    # Run with custom timeout and verbose output
    $0 --timeout 45m --verbose
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--suite)
            TEST_SUITE="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -b|--benchmark)
            BENCHMARK=true
            TEST_SUITE="performance"
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -r|--report-dir)
            REPORT_DIR="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Main execution
print_status "MCP Server E2E Test Runner"
print_status "========================="
print_status "Test Suite: $TEST_SUITE"
print_status "Timeout: $TIMEOUT"
print_status "Report Directory: $REPORT_DIR"

# Check prerequisites
check_prerequisites

# Run tests
EXIT_CODE=0
if [ "$TEST_SUITE" = "all" ]; then
    # Run all test suites
    for suite in protocol discovery composable; do
        if ! run_tests "$suite"; then
            EXIT_CODE=1
        fi
    done
    
    # Run benchmarks if requested
    if [ "$BENCHMARK" = true ]; then
        if ! run_tests "performance"; then
            EXIT_CODE=1
        fi
    fi
else
    # Run specific test suite
    if ! run_tests "$TEST_SUITE"; then
        EXIT_CODE=1
    fi
fi

# Generate reports
generate_reports

# Print summary
print_status "========================="
if [ $EXIT_CODE -eq 0 ]; then
    print_status "All tests passed! ✅"
else
    print_error "Some tests failed! ❌"
fi
print_status "Reports generated in: $REPORT_DIR"

# Print performance summary if benchmarks were run
if [ -f "${REPORT_DIR}/benchmark-results.log" ]; then
    print_status "Performance Summary:"
    grep -E "(Avg:|P95:|Throughput:)" "${REPORT_DIR}/benchmark-results.log" | tail -10
fi

exit $EXIT_CODE
#!/bin/bash

# Database Connectors - Comprehensive Test Runner
set -e

APP_NAME="db-connectors"
TEST_RESULTS_DIR="test-results"
COVERAGE_FILE="coverage.out"

echo "🧪 Database Connectors - Unit Test Suite"
echo "========================================"

# Clean previous test results
echo "🧹 Cleaning previous test results..."
rm -rf ${TEST_RESULTS_DIR}/
rm -f ${COVERAGE_FILE}
mkdir -p ${TEST_RESULTS_DIR}

# Install test dependencies
echo "📦 Installing test dependencies..."
go mod tidy

# Check if testify is available, install if not
if ! go list github.com/stretchr/testify > /dev/null 2>&1; then
    echo "📥 Installing testify testing framework..."
    go get github.com/stretchr/testify/assert
    go get github.com/stretchr/testify/mock
    go get github.com/stretchr/testify/suite
fi

# Install other test dependencies
go get github.com/DATA-DOG/go-sqlmock
go get github.com/go-testfixtures/testfixtures/v3

echo ""
echo "🏃 Running test suite..."

# Function to run tests for a package
run_package_tests() {
    local package_name=$1
    local package_path=$2
    
    echo ""
    echo "📁 Testing package: ${package_name}"
    echo "----------------------------------------"
    
    if [ -f "${package_path}"/*_test.go ] 2>/dev/null; then
        # Run tests with coverage
        go test -v -race -coverprofile="${TEST_RESULTS_DIR}/${package_name}_coverage.out" \
            -covermode=atomic ./${package_path}
        
        # Generate coverage report
        if [ -f "${TEST_RESULTS_DIR}/${package_name}_coverage.out" ]; then
            go tool cover -html="${TEST_RESULTS_DIR}/${package_name}_coverage.out" \
                -o "${TEST_RESULTS_DIR}/${package_name}_coverage.html"
        fi
    else
        echo "⚠️  No test files found in ${package_path}"
    fi
}

# Test each package
run_package_tests "connectors" "connectors"
run_package_tests "api" "api" 
run_package_tests "config" "config"
run_package_tests "main" "cmd"

echo ""
echo "🔄 Running integration tests..."
if [ -f "tests"/*_test.go ] 2>/dev/null; then
    go test -v -race -tags=integration ./tests/...
else
    echo "⚠️  No integration tests found"
fi

echo ""
echo "📊 Generating overall coverage report..."
# Combine all coverage files
echo "mode: atomic" > ${COVERAGE_FILE}
for coverage_file in ${TEST_RESULTS_DIR}/*_coverage.out; do
    if [ -f "$coverage_file" ]; then
        tail -n +2 "$coverage_file" >> ${COVERAGE_FILE}
    fi
done

# Generate overall coverage report
if [ -f ${COVERAGE_FILE} ]; then
    go tool cover -html=${COVERAGE_FILE} -o ${TEST_RESULTS_DIR}/overall_coverage.html
    COVERAGE_PERCENT=$(go tool cover -func=${COVERAGE_FILE} | grep total | awk '{print $3}')
    echo "📈 Overall test coverage: ${COVERAGE_PERCENT}"
fi

echo ""
echo "🧪 Running benchmarks..."
go test -bench=. -benchmem ./... > ${TEST_RESULTS_DIR}/benchmarks.txt 2>&1 || true

echo ""
echo "🔍 Running static analysis..."
# Install and run go vet
go vet ./...

# Install and run golint if available
if command -v golint > /dev/null 2>&1; then
    golint ./... > ${TEST_RESULTS_DIR}/lint_results.txt 2>&1 || true
else
    echo "⚠️  golint not installed, skipping lint checks"
fi

# Install and run gosec if available
if command -v gosec > /dev/null 2>&1; then
    gosec ./... > ${TEST_RESULTS_DIR}/security_results.txt 2>&1 || true
else
    echo "⚠️  gosec not installed, skipping security checks"
fi

echo ""
echo "✅ Test suite completed!"
echo ""
echo "📋 Results summary:"
echo "  📁 Test results: ${TEST_RESULTS_DIR}/"
echo "  📊 Coverage report: ${TEST_RESULTS_DIR}/overall_coverage.html"
echo "  🏃 Benchmark results: ${TEST_RESULTS_DIR}/benchmarks.txt"
if [ -f "${TEST_RESULTS_DIR}/lint_results.txt" ]; then
    echo "  🔍 Lint results: ${TEST_RESULTS_DIR}/lint_results.txt"
fi
if [ -f "${TEST_RESULTS_DIR}/security_results.txt" ]; then
    echo "  🔒 Security scan: ${TEST_RESULTS_DIR}/security_results.txt"
fi

echo ""
echo "🚀 To run individual tests:"
echo "  go test ./connectors/      # Test connectors"
echo "  go test ./api/             # Test API handlers"
echo "  go test ./config/          # Test configuration"
echo "  go test -v -run TestName   # Run specific test"
echo "  go test -bench=.           # Run benchmarks"

#!/bin/bash

# Database Connectors - Test Dependencies Installation Script
set -e

echo "🧪 Installing Test Dependencies for Database Connectors"
echo "====================================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    echo "   Visit: https://golang.org/doc/install"
    exit 1
fi

echo "✅ Go is installed: $(go version)"
echo ""

# Update go.mod
echo "📦 Updating go.mod..."
go mod tidy

# Install testing frameworks
echo "🔧 Installing testing frameworks..."

echo "  📥 Installing testify (assertions, mocks, suites)..."
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/stretchr/testify/suite
go get github.com/stretchr/testify/require

echo "  📥 Installing sqlmock (database mocking)..."
go get github.com/DATA-DOG/go-sqlmock

echo "  📥 Installing testfixtures (test data management)..."
go get github.com/go-testfixtures/testfixtures/v3

# Install code quality tools (optional but recommended)
echo ""
echo "🔍 Installing code quality tools..."

echo "  📥 Installing golint..."
if ! command -v golint &> /dev/null; then
    go install golang.org/x/lint/golint@latest
    echo "    ✅ golint installed"
else
    echo "    ✅ golint already installed"
fi

echo "  📥 Installing gosec (security scanner)..."
if ! command -v gosec &> /dev/null; then
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    echo "    ✅ gosec installed"
else
    echo "    ✅ gosec already installed"
fi

echo "  📥 Installing staticcheck..."
if ! command -v staticcheck &> /dev/null; then
    go install honnef.co/go/tools/cmd/staticcheck@latest
    echo "    ✅ staticcheck installed"
else
    echo "    ✅ staticcheck already installed"
fi

echo "  📥 Installing gocyclo (cyclomatic complexity)..."
if ! command -v gocyclo &> /dev/null; then
    go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
    echo "    ✅ gocyclo installed"
else
    echo "    ✅ gocyclo already installed"
fi

# Verify installations
echo ""
echo "🔬 Verifying test dependencies..."

# Check Go modules
echo "  📋 Checking go.mod dependencies..."
if go list -m github.com/stretchr/testify &> /dev/null; then
    echo "    ✅ testify: $(go list -m github.com/stretchr/testify)"
else
    echo "    ❌ testify not found"
fi

if go list -m github.com/DATA-DOG/go-sqlmock &> /dev/null; then
    echo "    ✅ sqlmock: $(go list -m github.com/DATA-DOG/go-sqlmock)"
else
    echo "    ❌ sqlmock not found"
fi

# Check tools
echo "  🔧 Checking installed tools..."
if command -v golint &> /dev/null; then
    echo "    ✅ golint: $(which golint)"
else
    echo "    ⚠️  golint not in PATH"
fi

if command -v gosec &> /dev/null; then
    echo "    ✅ gosec: $(which gosec)"
else
    echo "    ⚠️  gosec not in PATH"
fi

if command -v staticcheck &> /dev/null; then
    echo "    ✅ staticcheck: $(which staticcheck)"
else
    echo "    ⚠️  staticcheck not in PATH"
fi

# Final cleanup
echo ""
echo "🧹 Final cleanup..."
go mod tidy

# Test installation by running a simple test
echo ""
echo "🧪 Testing installation..."
echo "  Running a simple test to verify everything works..."

# Create a temporary test file to verify installation
cat > temp_test.go << 'EOF'
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/DATA-DOG/go-sqlmock"
    "database/sql"
)

func TestInstallation(t *testing.T) {
    // Test testify
    assert.True(t, true, "testify is working")
    
    // Test sqlmock
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    assert.NotNil(t, db)
    assert.NotNil(t, mock)
    db.Close()
}
EOF

# Run the test
if go test temp_test.go; then
    echo "    ✅ Test dependencies verification passed!"
else
    echo "    ❌ Test dependencies verification failed!"
fi

# Clean up
rm -f temp_test.go

echo ""
echo "🎉 Test Dependencies Installation Complete!"
echo ""
echo "📋 Summary of installed packages:"
echo "  • github.com/stretchr/testify (assertions, mocks, suites)"
echo "  • github.com/DATA-DOG/go-sqlmock (database mocking)"
echo "  • github.com/go-testfixtures/testfixtures/v3 (test fixtures)"
echo ""
echo "🔧 Tools installed:"
echo "  • golint (code linting)"
echo "  • gosec (security scanning)"
echo "  • staticcheck (static analysis)"
echo "  • gocyclo (complexity analysis)"
echo ""
echo "🚀 Next steps:"
echo "  1. Run all tests: ./test_runner.sh"
echo "  2. Run specific tests: go test ./connectors/ -v"
echo "  3. Generate coverage: go test ./... -cover"
echo "  4. Run benchmarks: go test ./... -bench=."
echo ""
echo "📖 For detailed testing guide, see: TESTING.md"

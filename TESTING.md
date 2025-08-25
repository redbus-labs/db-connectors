# Database Connectors - Testing Guide

This guide covers the comprehensive testing framework for the Database Connectors project, including unit tests, integration tests, benchmarks, and testing best practices.

## 🧪 Testing Overview

Our testing framework includes:
- **Unit Tests** - Test individual components in isolation
- **Integration Tests** - Test component interactions and API endpoints
- **Benchmark Tests** - Performance testing for critical paths
- **Mock Tests** - Database operations with mock connections
- **End-to-End Tests** - Complete workflow testing

## 📁 Test Structure

```
db-connectors/
├── connectors/
│   ├── interface_test.go      # Interface and config tests
│   ├── mongodb_test.go        # MongoDB connector tests
│   ├── mysql_test.go          # MySQL connector tests (with mocks)
│   └── postgres_test.go       # PostgreSQL connector tests (with mocks)
├── api/
│   └── handlers_test.go       # API handler tests
├── config/
│   └── config_test.go         # Configuration tests
├── cmd/
│   └── main_test.go           # Main function tests
├── tests/
│   └── integration_test.go    # Integration tests
├── test_runner.sh             # Automated test runner
└── TESTING.md                 # This file
```

## 🚀 Quick Start

### Run All Tests
```bash
# Make test runner executable
chmod +x test_runner.sh

# Run complete test suite
./test_runner.sh
```

### Run Individual Test Packages
```bash
# Test connectors package
go test ./connectors/ -v

# Test API handlers
go test ./api/ -v

# Test configuration
go test ./config/ -v

# Test main package
go test ./cmd/ -v

# Run integration tests
go test ./tests/ -v -tags=integration
```

## 📊 Test Coverage

### Generate Coverage Report
```bash
# Generate coverage for all packages
go test ./... -coverprofile=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Package-Specific Coverage
```bash
# Connectors package coverage
go test ./connectors/ -coverprofile=connectors_coverage.out
go tool cover -html=connectors_coverage.out -o connectors_coverage.html

# API package coverage
go test ./api/ -coverprofile=api_coverage.out
go tool cover -html=api_coverage.out -o api_coverage.html
```

## 🧪 Unit Tests

### Connectors Package Tests

#### Interface Tests (`connectors/interface_test.go`)
- **ConnectionConfig validation**
- **Connection string generation** for all database types
- **DatabaseConfig retrieval** methods

#### MongoDB Tests (`connectors/mongodb_test.go`)
- **Connection string generation** with/without authentication
- **Operation parameter validation**
- **Error handling** for invalid operations
- **Performance benchmarks**

#### MySQL Tests (`connectors/mysql_test.go`)
- **SQL query execution** with sqlmock
- **CRUD operations** (Create, Read, Update, Delete)
- **Connection state management**
- **Error scenarios**

#### PostgreSQL Tests (`connectors/postgres_test.go`)
- **PostgreSQL-specific syntax** (parameterized queries)
- **SSL mode handling**
- **Advanced features** (RETURNING clauses)
- **Mock database interactions**

### API Package Tests

#### Handler Tests (`api/handlers_test.go`)
- **HTTP endpoint testing**
- **Request validation**
- **Response formatting**
- **Error handling**
- **Mock database connectors**

### Configuration Tests (`config/config_test.go`)
- **YAML file loading**
- **Environment variable override**
- **Partial configuration handling**
- **Default values**
- **Validation logic**

## 🔗 Integration Tests

### Running Integration Tests
```bash
# Run with integration tag
go test ./tests/ -v -tags=integration

# Run specific integration test
go test ./tests/ -v -tags=integration -run TestHealthEndpoint
```

### Integration Test Coverage
- **API endpoint workflows**
- **Database connection testing**
- **AllConfig CRUD operations**
- **Error handling across components**
- **Concurrent request handling**
- **Swagger documentation endpoints**

## 🏃 Performance Testing

### Benchmark Tests
```bash
# Run all benchmarks
go test ./... -bench=.

# Run specific package benchmarks
go test ./connectors/ -bench=.
go test ./api/ -bench=.

# Benchmark with memory allocation info
go test ./... -bench=. -benchmem

# Run benchmarks multiple times for accuracy
go test ./... -bench=. -count=5
```

### Benchmark Categories
- **Connector creation performance**
- **API handler response times**
- **Configuration loading speed**
- **Concurrent request handling**
- **Memory allocation patterns**

## 🎯 Mock Testing

### Database Mocking with sqlmock
```go
// Example from mysql_test.go
db, mock, err := sqlmock.New()
suite.Require().NoError(err)

// Set up expectations
mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

// Execute test
result, err := connector.Query(ctx, "SELECT id, name FROM users")

// Verify expectations
assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
```

### Mock Connector for API Tests
```go
// Custom mock implementing DBConnector interface
type MockDBConnector struct {
    mock.Mock
}

func (m *MockDBConnector) Connect(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}
```

## 📋 Test Categories and Examples

### 1. Unit Tests
```bash
# Test specific function
go test ./connectors/ -v -run TestConnectionConfig_Validate

# Test with coverage
go test ./connectors/ -v -cover -run TestNewMySQLConnector
```

### 2. Table-Driven Tests
Many tests use table-driven approach for comprehensive coverage:
```go
tests := []struct {
    name    string
    config  ConnectionConfig
    wantErr bool
}{
    {
        name: "valid mysql config",
        config: ConnectionConfig{...},
        wantErr: false,
    },
    // More test cases...
}
```

### 3. Test Suites
Using testify/suite for organized test management:
```go
type MySQLConnectorTestSuite struct {
    suite.Suite
    connector *MySQLConnector
    db        *sql.DB
    mock      sqlmock.Sqlmock
}
```

### 4. Benchmark Tests
Performance testing for critical paths:
```go
func BenchmarkConnectorCreation(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        connector := NewMySQLConnector(config)
        _ = connector
    }
}
```

## 🔧 Test Configuration

### Environment Variables for Testing
```bash
# Test database connections (optional)
export TEST_MYSQL_HOST=localhost
export TEST_MYSQL_PORT=3306
export TEST_MYSQL_USER=test_user
export TEST_MYSQL_PASSWORD=test_pass
export TEST_MYSQL_DATABASE=test_db

export TEST_POSTGRES_HOST=localhost
export TEST_POSTGRES_PORT=5432
export TEST_POSTGRES_USER=test_user
export TEST_POSTGRES_PASSWORD=test_pass
export TEST_POSTGRES_DATABASE=test_db

export TEST_MONGO_HOST=localhost
export TEST_MONGO_PORT=27017
export TEST_MONGO_DATABASE=test_db
```

### Test Flags and Options
```bash
# Verbose output
go test ./... -v

# Run tests with race detection
go test ./... -race

# Set test timeout
go test ./... -timeout=30s

# Run only short tests
go test ./... -short

# Run specific test pattern
go test ./... -run="TestMySQL.*"

# Generate test coverage
go test ./... -coverprofile=coverage.out

# Benchmark with memory stats
go test ./... -bench=. -benchmem
```

## 📈 Continuous Integration

### Test Pipeline Commands
```bash
# Basic test pipeline
go mod tidy
go vet ./...
go test ./... -race -coverprofile=coverage.out
go tool cover -func=coverage.out

# Extended pipeline with benchmarks
go test ./... -bench=. -benchmem
go test ./tests/ -tags=integration
```

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go mod tidy
      - run: go vet ./...
      - run: go test ./... -race -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out
```

## 🐛 Testing Best Practices

### 1. Test Naming
- Use descriptive test names: `TestConnectionConfig_Validate`
- Group related tests: `TestMySQL.*`
- Use table-driven tests for multiple scenarios

### 2. Test Organization
- Keep tests close to the code they test
- Use test suites for complex setup/teardown
- Separate unit and integration tests

### 3. Mock Usage
- Mock external dependencies (databases, APIs)
- Use sqlmock for database interactions
- Keep mocks simple and focused

### 4. Error Testing
- Test both success and failure cases
- Verify error messages and types
- Test edge cases and boundary conditions

### 5. Performance Testing
- Benchmark critical paths
- Test with realistic data sizes
- Monitor memory allocations

## 🛠️ Troubleshooting Tests

### Common Issues

#### 1. Import Path Issues
```bash
# Ensure proper module path
go mod tidy
```

#### 2. Missing Test Dependencies
```bash
# Install test dependencies
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/stretchr/testify/suite
go get github.com/DATA-DOG/go-sqlmock
```

#### 3. Database Connection Failures
- Tests are designed to work without real databases
- Mock connections are used for unit tests
- Integration tests may skip if no real DB available

#### 4. Test Timeouts
```bash
# Increase test timeout
go test ./... -timeout=60s
```

### Debug Test Failures
```bash
# Run specific failing test with verbose output
go test ./api/ -v -run TestHealthHandler

# Run with race detection
go test ./api/ -race -run TestHealthHandler

# Print test coverage for specific test
go test ./api/ -cover -run TestHealthHandler
```

## 📊 Test Metrics and Reporting

### Coverage Goals
- **Overall Coverage**: > 80%
- **Critical Paths**: > 95%
- **API Handlers**: > 90%
- **Database Connectors**: > 85%

### Test Execution Time
- **Unit Tests**: < 10 seconds
- **Integration Tests**: < 30 seconds
- **Full Suite**: < 60 seconds

### Performance Benchmarks
- **API Response Time**: < 10ms (mocked)
- **Connector Creation**: < 1ms
- **Configuration Loading**: < 5ms

## 🚀 Advanced Testing

### Property-Based Testing
Consider adding property-based tests for:
- Configuration validation
- Connection string generation
- Data serialization/deserialization

### Fuzz Testing
```bash
# Example fuzz test command
go test ./config/ -fuzz=FuzzConfigParsing
```

### Load Testing
For production readiness:
- Use tools like `hey` or `ab` for HTTP load testing
- Test concurrent database connections
- Measure memory usage under load

## 📚 References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Framework](https://github.com/stretchr/testify)
- [SQLMock](https://github.com/DATA-DOG/go-sqlmock)
- [Go Testing Best Practices](https://go.dev/blog/table-driven-tests)

---

## 🎯 Quick Commands Reference

```bash
# Run everything
./test_runner.sh

# Quick unit tests
go test ./... -short

# With coverage
go test ./... -cover

# Integration tests only
go test ./tests/ -tags=integration

# Benchmarks only
go test ./... -bench=. -run=^$

# Specific package
go test ./connectors/ -v

# Specific test
go test ./api/ -run TestHealthHandler -v
```

Happy Testing! 🧪✨

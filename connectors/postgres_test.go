package connectors

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// PostgreSQLConnectorTestSuite defines the test suite for PostgreSQL connector
type PostgreSQLConnectorTestSuite struct {
	suite.Suite
	connector *PostgreSQLConnector
	config    *ConnectionConfig
	db        *sql.DB
	mock      sqlmock.Sqlmock
}

// SetupSuite sets up the test suite
func (suite *PostgreSQLConnectorTestSuite) SetupSuite() {
	suite.config = &ConnectionConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Database: "test_db",
		SSLMode:  "disable",
	}
}

// SetupTest sets up each test
func (suite *PostgreSQLConnectorTestSuite) SetupTest() {
	suite.connector = NewPostgreSQLConnector(suite.config)
	
	// Create a mock database
	db, mock, err := sqlmock.New()
	suite.Require().NoError(err)
	suite.db = db
	suite.mock = mock
}

// TearDownTest cleans up after each test
func (suite *PostgreSQLConnectorTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

// TestNewPostgreSQLConnector tests the creation of a new PostgreSQL connector
func (suite *PostgreSQLConnectorTestSuite) TestNewPostgreSQLConnector() {
	connector := NewPostgreSQLConnector(suite.config)
	assert.NotNil(suite.T(), connector)
	assert.Equal(suite.T(), suite.config, connector.config)
}

// TestGetType tests the GetType method
func (suite *PostgreSQLConnectorTestSuite) TestGetType() {
	assert.Equal(suite.T(), "postgresql", suite.connector.GetType())
}

// TestConnectionString tests PostgreSQL connection string generation
func TestPostgreSQLConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *ConnectionConfig
		expected string
	}{
		{
			name: "basic connection with SSL disabled",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable",
		},
		{
			name: "custom host and port",
			config: &ConnectionConfig{
				Host:     "postgres.example.com",
				Port:     5433,
				Username: "user",
				Password: "pass",
				Database: "mydb",
				SSLMode:  "require",
			},
			expected: "host=postgres.example.com port=5433 user=user password=pass dbname=mydb sslmode=require",
		},
		{
			name: "default SSL mode",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable",
		},
		{
			name: "empty password",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=postgres password= dbname=testdb sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connectionString, err := tt.config.GetConnectionString("postgresql")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, connectionString)
		})
	}
}

// TestQuery tests the Query method with mock
func (suite *PostgreSQLConnectorTestSuite) TestQuery() {
	// Replace the connector's db with our mock
	suite.connector.db = suite.db

	// Set up expectations
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test1").
		AddRow(2, "test2")
	
	suite.mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	// Execute query
	ctx := context.Background()
	result, err := suite.connector.Query(ctx, "SELECT id, name FROM users")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// Verify expectations
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestExecute tests the Execute method for various operations
func (suite *PostgreSQLConnectorTestSuite) TestExecute() {
	// Replace the connector's db with our mock
	suite.connector.db = suite.db

	tests := []struct {
		name      string
		operation string
		params    map[string]interface{}
		setupMock func()
		wantErr   bool
	}{
		{
			name:      "select operation",
			operation: "select",
			params: map[string]interface{}{
				"query": "SELECT * FROM users",
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test")
				suite.mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:      "insert operation",
			operation: "insert",
			params: map[string]interface{}{
				"query": "INSERT INTO users (name) VALUES ($1)",
				"args":  []interface{}{"John"},
			},
			setupMock: func() {
				suite.mock.ExpectExec("INSERT INTO users").
					WithArgs("John").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:      "update operation with PostgreSQL syntax",
			operation: "update",
			params: map[string]interface{}{
				"query": "UPDATE users SET name = $1 WHERE id = $2",
				"args":  []interface{}{"Jane", 1},
			},
			setupMock: func() {
				suite.mock.ExpectExec("UPDATE users").
					WithArgs("Jane", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "delete operation",
			operation: "delete",
			params: map[string]interface{}{
				"query": "DELETE FROM users WHERE id = $1",
				"args":  []interface{}{1},
			},
			setupMock: func() {
				suite.mock.ExpectExec("DELETE FROM users").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "execute operation (DDL)",
			operation: "execute",
			params: map[string]interface{}{
				"query": "CREATE TABLE test (id SERIAL PRIMARY KEY, name VARCHAR(100))",
			},
			setupMock: func() {
				suite.mock.ExpectExec("CREATE TABLE test").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name:      "operation without query",
			operation: "select",
			params:    map[string]interface{}{},
			setupMock: func() {},
			wantErr:   true,
		},
		{
			name:      "unsupported operation",
			operation: "unsupported",
			params: map[string]interface{}{
				"query": "SELECT * FROM users",
			},
			setupMock: func() {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			ctx := context.Background()
			result, err := suite.connector.Execute(ctx, tt.operation, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Check that all expectations were met
			if !tt.wantErr {
				assert.NoError(t, suite.mock.ExpectationsWereMet())
			}
		})
	}
}

// TestPing tests the Ping method
func (suite *PostgreSQLConnectorTestSuite) TestPing() {
	// Replace the connector's db with our mock
	suite.connector.db = suite.db

	// Set up expectation for ping
	suite.mock.ExpectPing()

	ctx := context.Background()
	err := suite.connector.Ping(ctx)

	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestPingError tests the Ping method with error
func (suite *PostgreSQLConnectorTestSuite) TestPingError() {
	// Test ping error without mock (simpler test)
	ctx := context.Background()
	err := suite.connector.Ping(ctx)

	// Should error since there's no real connection
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "PostgreSQL connection not established")
}

// TestIsConnected tests the IsConnected method
func (suite *PostgreSQLConnectorTestSuite) TestIsConnected() {
	// Without connection
	assert.False(suite.T(), suite.connector.IsConnected())

	// With mock connection
	suite.connector.db = suite.db
	suite.mock.ExpectPing()
	
	assert.True(suite.T(), suite.connector.IsConnected())
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestClose tests the Close method
func (suite *PostgreSQLConnectorTestSuite) TestClose() {
	// Replace the connector's db with our mock
	suite.connector.db = suite.db

	// Set up expectation for close
	suite.mock.ExpectClose()

	err := suite.connector.Close()

	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// Test error handling when database is not connected
func (suite *PostgreSQLConnectorTestSuite) TestOperationsWithoutConnection() {
	ctx := context.Background()

	// Test Query without connection
	rows, err := suite.connector.Query(ctx, "SELECT * FROM users")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), rows)
	assert.Contains(suite.T(), err.Error(), "PostgreSQL connection not established")

	// Test Execute without connection
	result, err := suite.connector.Execute(ctx, "select", map[string]interface{}{
		"query": "SELECT * FROM users",
	})
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "PostgreSQL connection not established")

	// Test Ping without connection
	err = suite.connector.Ping(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "PostgreSQL connection not established")
}

// TestPostgreSQLSpecificFeatures tests PostgreSQL-specific features
func (suite *PostgreSQLConnectorTestSuite) TestPostgreSQLSpecificFeatures() {
	// Replace the connector's db with our mock
	suite.connector.db = suite.db

	tests := []struct {
		name      string
		operation string
		params    map[string]interface{}
		setupMock func()
	}{
		{
			name:      "query with PostgreSQL-style parameters",
			operation: "select",
			params: map[string]interface{}{
				"query": "SELECT * FROM users WHERE age > $1 AND status = $2",
				"args":  []interface{}{18, "active"},
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "age", "status"}).
					AddRow(1, "John", 25, "active").
					AddRow(2, "Jane", 30, "active")
				suite.mock.ExpectQuery("SELECT (.+) FROM users WHERE age > (.+) AND status = (.+)").
					WithArgs(18, "active").
					WillReturnRows(rows)
			},
		},
		{
			name:      "insert with RETURNING clause",
			operation: "select",
			params: map[string]interface{}{
				"query": "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
				"args":  []interface{}{"John Doe", "john@example.com"},
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				suite.mock.ExpectQuery("INSERT INTO users").
					WithArgs("John Doe", "john@example.com").
					WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			ctx := context.Background()
			result, err := suite.connector.Execute(ctx, tt.operation, tt.params)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NoError(t, suite.mock.ExpectationsWereMet())
		})
	}
}

// Run the test suite
func TestPostgreSQLConnectorTestSuite(t *testing.T) {
	suite.Run(t, new(PostgreSQLConnectorTestSuite))
}

// Benchmark tests
func BenchmarkPostgreSQLConnectorCreation(b *testing.B) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Database: "test_db",
		SSLMode:  "disable",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connector := NewPostgreSQLConnector(config)
		_ = connector
	}
}

func BenchmarkPostgreSQLGetType(b *testing.B) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Database: "test_db",
		SSLMode:  "disable",
	}
	connector := NewPostgreSQLConnector(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = connector.GetType()
	}
}

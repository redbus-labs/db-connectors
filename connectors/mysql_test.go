package connectors

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MySQLConnectorTestSuite defines the test suite for MySQL connector
type MySQLConnectorTestSuite struct {
	suite.Suite
	connector *MySQLConnector
	config    *ConnectionConfig
	db        *sql.DB
	mock      sqlmock.Sqlmock
}

// SetupSuite sets up the test suite
func (suite *MySQLConnectorTestSuite) SetupSuite() {
	suite.config = &ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "test_db",
	}
}

// SetupTest sets up each test
func (suite *MySQLConnectorTestSuite) SetupTest() {
	suite.connector = NewMySQLConnector(suite.config)
	
	// Create a mock database
	db, mock, err := sqlmock.New()
	suite.Require().NoError(err)
	suite.db = db
	suite.mock = mock
}

// TearDownTest cleans up after each test
func (suite *MySQLConnectorTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

// TestNewMySQLConnector tests the creation of a new MySQL connector
func (suite *MySQLConnectorTestSuite) TestNewMySQLConnector() {
	connector := NewMySQLConnector(suite.config)
	assert.NotNil(suite.T(), connector)
	assert.Equal(suite.T(), suite.config, connector.config)
}

// TestGetType tests the GetType method
func (suite *MySQLConnectorTestSuite) TestGetType() {
	assert.Equal(suite.T(), "mysql", suite.connector.GetType())
}

// TestConnectionString tests MySQL connection string generation
func TestMySQLConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *ConnectionConfig
		expected string
	}{
		{
			name: "basic connection",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			expected: "root:password@tcp(localhost:3306)/testdb",
		},
		{
			name: "custom host and port",
			config: &ConnectionConfig{
				Host:     "mysql.example.com",
				Port:     3307,
				Username: "user",
				Password: "pass",
				Database: "mydb",
			},
			expected: "user:pass@tcp(mysql.example.com:3307)/mydb",
		},
		{
			name: "empty password",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Database: "testdb",
			},
			expected: "root:@tcp(localhost:3306)/testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connectionString, err := tt.config.GetConnectionString("mysql")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, connectionString)
		})
	}
}

// TestQuery tests the Query method with mock
func (suite *MySQLConnectorTestSuite) TestQuery() {
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
func (suite *MySQLConnectorTestSuite) TestExecute() {
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
				"query": "INSERT INTO users (name) VALUES (?)",
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
			name:      "update operation",
			operation: "update",
			params: map[string]interface{}{
				"query": "UPDATE users SET name = ? WHERE id = ?",
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
				"query": "DELETE FROM users WHERE id = ?",
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
func (suite *MySQLConnectorTestSuite) TestPing() {
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
func (suite *MySQLConnectorTestSuite) TestPingError() {
	// Test ping error without mock (simpler test)
	ctx := context.Background()
	err := suite.connector.Ping(ctx)

	// Should error since there's no real connection
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "MySQL connection not established")
}

// TestIsConnected tests the IsConnected method
func (suite *MySQLConnectorTestSuite) TestIsConnected() {
	// Without connection
	assert.False(suite.T(), suite.connector.IsConnected())

	// With mock connection
	suite.connector.db = suite.db
	suite.mock.ExpectPing()
	
	assert.True(suite.T(), suite.connector.IsConnected())
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// TestClose tests the Close method
func (suite *MySQLConnectorTestSuite) TestClose() {
	// Replace the connector's db with our mock
	suite.connector.db = suite.db

	// Set up expectation for close
	suite.mock.ExpectClose()

	err := suite.connector.Close()

	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// Test error handling when database is not connected
func (suite *MySQLConnectorTestSuite) TestOperationsWithoutConnection() {
	ctx := context.Background()

	// Test Query without connection
	rows, err := suite.connector.Query(ctx, "SELECT * FROM users")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), rows)
	assert.Contains(suite.T(), err.Error(), "MySQL connection not established")

	// Test Execute without connection
	result, err := suite.connector.Execute(ctx, "select", map[string]interface{}{
		"query": "SELECT * FROM users",
	})
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "MySQL connection not established")

	// Test Ping without connection
	err = suite.connector.Ping(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "MySQL connection not established")
}

// Run the test suite
func TestMySQLConnectorTestSuite(t *testing.T) {
	suite.Run(t, new(MySQLConnectorTestSuite))
}

// Benchmark tests
func BenchmarkMySQLConnectorCreation(b *testing.B) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "test_db",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connector := NewMySQLConnector(config)
		_ = connector
	}
}

func BenchmarkMySQLGetType(b *testing.B) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "test_db",
	}
	connector := NewMySQLConnector(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = connector.GetType()
	}
}

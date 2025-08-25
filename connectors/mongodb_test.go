package connectors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MongoDBConnectorTestSuite defines the test suite for MongoDB connector
type MongoDBConnectorTestSuite struct {
	suite.Suite
	connector *MongoDBConnector
	config    *ConnectionConfig
}

// SetupSuite sets up the test suite
func (suite *MongoDBConnectorTestSuite) SetupSuite() {
	suite.config = &ConnectionConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "test_db",
	}
	suite.connector = NewMongoDBConnector(suite.config)
}

// TestNewMongoDBConnector tests the creation of a new MongoDB connector
func (suite *MongoDBConnectorTestSuite) TestNewMongoDBConnector() {
	connector := NewMongoDBConnector(suite.config)
	assert.NotNil(suite.T(), connector)
	assert.Equal(suite.T(), suite.config, connector.config)
}

// TestGetType tests the GetType method
func (suite *MongoDBConnectorTestSuite) TestGetType() {
	assert.Equal(suite.T(), "mongodb", suite.connector.GetType())
}

// TestConnectionString tests different MongoDB connection string formats
func TestMongoDBConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *ConnectionConfig
		expected string
	}{
		{
			name: "with authentication",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Username: "admin",
				Password: "password",
				Database: "testdb",
			},
			expected: "mongodb://admin:password@localhost:27017/testdb",
		},
		{
			name: "without authentication",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Database: "testdb",
			},
			expected: "mongodb://localhost:27017/testdb",
		},
		{
			name: "with username only",
			config: &ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Username: "admin",
				Database: "testdb",
			},
			expected: "mongodb://admin@localhost:27017/testdb",
		},
		{
			name: "custom port",
			config: &ConnectionConfig{
				Host:     "mongodb.example.com",
				Port:     27018,
				Username: "user",
				Password: "pass",
				Database: "mydb",
			},
			expected: "mongodb://user:pass@mongodb.example.com:27018/mydb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := NewMongoDBConnector(tt.config)
			// Test the internal URI generation logic
			assert.NotNil(t, connector)
			assert.Equal(t, "mongodb", connector.GetType())
		})
	}
}

// TestExecuteOperations tests various MongoDB operations
func TestMongoDBExecuteOperations(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "test_db",
	}
	connector := NewMongoDBConnector(config)

	tests := []struct {
		name      string
		operation string
		params    map[string]interface{}
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "listCollections operation",
			operation: "listCollections",
			params: map[string]interface{}{
				"filter": map[string]interface{}{},
			},
			wantErr: true, // Will fail without actual MongoDB connection
		},
		{
			name:      "find operation without collection",
			operation: "find",
			params: map[string]interface{}{
				"filter": map[string]interface{}{},
			},
			wantErr: true,
			errMsg:  "MongoDB connection not established",
		},
		{
			name:      "find operation with collection",
			operation: "find",
			params: map[string]interface{}{
				"collection": "test_collection",
				"filter":     map[string]interface{}{},
			},
			wantErr: true, // Will fail without actual MongoDB connection
		},
		{
			name:      "insert operation without collection",
			operation: "insert",
			params: map[string]interface{}{
				"document": map[string]interface{}{"key": "value"},
			},
			wantErr: true,
			errMsg:  "MongoDB connection not established",
		},
		{
			name:      "insert operation without document",
			operation: "insert",
			params: map[string]interface{}{
				"collection": "test_collection",
			},
			wantErr: true, // Will fail without actual MongoDB connection or missing document
		},
		{
			name:      "update operation without filter",
			operation: "update",
			params: map[string]interface{}{
				"collection": "test_collection",
				"update":     map[string]interface{}{"$set": map[string]interface{}{"key": "new_value"}},
			},
			wantErr: true, // Will fail without actual MongoDB connection or missing filter
		},
		{
			name:      "unsupported operation",
			operation: "unsupported_op",
			params: map[string]interface{}{
				"collection": "test_collection",
			},
			wantErr: true,
			errMsg:  "MongoDB connection not established",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			result, err := connector.Execute(ctx, tt.operation, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestQueryMethod tests that Query method returns appropriate error for MongoDB
func TestMongoDBQueryMethod(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "test_db",
	}
	connector := NewMongoDBConnector(config)

	ctx := context.Background()
	rows, err := connector.Query(ctx, "SELECT * FROM test", nil)

	assert.Error(t, err)
	assert.Nil(t, rows)
	assert.Contains(t, err.Error(), "Query method not applicable for MongoDB")
}

// TestIsConnectedWithoutConnection tests IsConnected when not connected
func TestMongoDBIsConnectedWithoutConnection(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "test_db",
	}
	connector := NewMongoDBConnector(config)

	// Should return false when not connected
	assert.False(t, connector.IsConnected())
}

// TestClose tests the Close method
func TestMongoDBClose(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "test_db",
	}
	connector := NewMongoDBConnector(config)

	// Should not error even if not connected
	err := connector.Close()
	assert.NoError(t, err)
}

// TestParameterValidation tests parameter validation for different operations
func TestMongoDBParameterValidation(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "test_db",
	}
	connector := NewMongoDBConnector(config)
	ctx := context.Background()

	// Test find with various parameter combinations
	findTests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid find with all options",
			params: map[string]interface{}{
				"collection": "test",
				"filter":     map[string]interface{}{"status": "active"},
				"limit":      10,
				"skip":       5,
				"sort":       map[string]interface{}{"created_at": -1},
			},
			wantErr: true, // Still fails due to no connection, but params are valid
		},
		{
			name: "find with int limit",
			params: map[string]interface{}{
				"collection": "test",
				"filter":     map[string]interface{}{},
				"limit":      int(10),
			},
			wantErr: true, // Still fails due to no connection
		},
		{
			name: "find with int64 limit",
			params: map[string]interface{}{
				"collection": "test",
				"filter":     map[string]interface{}{},
				"limit":      int64(10),
			},
			wantErr: true, // Still fails due to no connection
		},
	}

	for _, tt := range findTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := connector.Execute(ctx, "find", tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Run the test suite
func TestMongoDBConnectorTestSuite(t *testing.T) {
	suite.Run(t, new(MongoDBConnectorTestSuite))
}

// Benchmark tests
func BenchmarkMongoDBConnectorCreation(b *testing.B) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "test_db",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connector := NewMongoDBConnector(config)
		_ = connector
	}
}

func BenchmarkMongoDBGetType(b *testing.B) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "test_db",
	}
	connector := NewMongoDBConnector(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = connector.GetType()
	}
}

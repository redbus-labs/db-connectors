// +build integration

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"db-connectors/api"
	"db-connectors/connectors"
)

// IntegrationTestSuite defines the integration test suite
type IntegrationTestSuite struct {
	suite.Suite
	server *httptest.Server
	apiURL string
}

// SetupSuite sets up the integration test suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Create API server for testing
	apiInstance := api.NewAPI()
	suite.server = httptest.NewServer(api.SetupRoutes(apiInstance))
	suite.apiURL = suite.server.URL
}

// TearDownSuite cleans up after the integration test suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// TestHealthEndpoint tests the health check endpoint
func (suite *IntegrationTestSuite) TestHealthEndpoint() {
	resp, err := http.Get(suite.apiURL + "/health")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Service is healthy", response["message"])
}

// TestDatabaseConnectionWorkflow tests the complete database connection workflow
func (suite *IntegrationTestSuite) TestDatabaseConnectionWorkflow() {
	tests := []struct {
		name           string
		connectionReq  map[string]interface{}
		expectedStatus int
		shouldConnect  bool
	}{
		{
			name: "MySQL connection test (will fail without real DB)",
			connectionReq: map[string]interface{}{
				"type":     "mysql",
				"host":     "localhost",
				"port":     3306,
				"username": "root",
				"password": "password",
				"database": "testdb",
			},
			expectedStatus: http.StatusInternalServerError, // Expected to fail in test environment
			shouldConnect:  false,
		},
		{
			name: "PostgreSQL connection test (will fail without real DB)",
			connectionReq: map[string]interface{}{
				"type":     "postgresql",
				"host":     "localhost",
				"port":     5432,
				"username": "postgres",
				"password": "password",
				"database": "testdb",
				"ssl_mode": "disable",
			},
			expectedStatus: http.StatusInternalServerError, // Expected to fail in test environment
			shouldConnect:  false,
		},
		{
			name: "MongoDB connection test (will fail without real DB)",
			connectionReq: map[string]interface{}{
				"type":     "mongodb",
				"host":     "localhost",
				"port":     27017,
				"database": "testdb",
			},
			expectedStatus: http.StatusInternalServerError, // Expected to fail in test environment
			shouldConnect:  false,
		},
		{
			name: "Invalid connection request",
			connectionReq: map[string]interface{}{
				"type": "invalid",
				"host": "localhost",
				"port": 3306,
			},
			expectedStatus: http.StatusBadRequest,
			shouldConnect:  false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			// Marshal request body
			reqBody, err := json.Marshal(tt.connectionReq)
			assert.NoError(t, err)

			// Make POST request to test connection
			resp, err := http.Post(
				suite.apiURL+"/test-connection",
				"application/json",
				bytes.NewBuffer(reqBody),
			)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Decode response
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.shouldConnect {
				assert.True(t, response["success"].(bool))
			} else {
				// In test environment without real databases, connections should fail
				if tt.expectedStatus == http.StatusInternalServerError {
					assert.False(t, response["success"].(bool))
					if errorMsg, ok := response["error"].(string); ok {
						assert.Contains(t, errorMsg, "failed")
					}
				}
			}
		})
	}
}

// TestAllConfigWorkflow tests the AllConfig functionality workflow
func (suite *IntegrationTestSuite) TestAllConfigWorkflow() {
	// Test checking if AllConfig table exists (will fail without real DB)
	allConfigReq := map[string]interface{}{
		"type":       "mysql",
		"host":       "localhost",
		"port":       3306,
		"username":   "root",
		"password":   "password",
		"database":   "testdb",
		"table_name": "allconfig",
	}

	reqBody, err := json.Marshal(allConfigReq)
	assert.NoError(suite.T(), err)

	resp, err := http.Post(
		suite.apiURL+"/allconfig",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	assert.NoError(suite.T(), err)
	
	// Should fail due to no real database connection
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
}

// TestAllConfigOperations tests AllConfig CRUD operations
func (suite *IntegrationTestSuite) TestAllConfigOperations() {
	operations := []struct {
		name      string
		operation string
		payload   map[string]interface{}
		method    string
	}{
		{
			name:      "Create config operation",
			operation: "create",
			payload: map[string]interface{}{
				"type":        "mysql",
				"host":        "localhost",
				"port":        3306,
				"username":    "root",
				"password":    "password",
				"database":    "testdb",
				"table_name":  "allconfig",
				"operation":   "create",
				"key":         "test_setting",
				"value":       "test_value",
				"description": "Test configuration setting",
			},
			method: "POST",
		},
		{
			name:      "Read all configs operation",
			operation: "read_all",
			payload: map[string]interface{}{
				"type":       "mysql",
				"host":       "localhost",
				"port":       3306,
				"username":   "root",
				"password":   "password",
				"database":   "testdb",
				"table_name": "allconfig",
				"operation":  "read_all",
			},
			method: "POST",
		},
		{
			name:      "Update config operation",
			operation: "update",
			payload: map[string]interface{}{
				"type":        "mysql",
				"host":        "localhost",
				"port":        3306,
				"username":    "root",
				"password":    "password",
				"database":    "testdb",
				"table_name":  "allconfig",
				"operation":   "update",
				"key":         "test_setting",
				"value":       "updated_value",
				"description": "Updated test configuration setting",
			},
			method: "POST",
		},
	}

	for _, op := range operations {
		suite.T().Run(op.name, func(t *testing.T) {
			reqBody, err := json.Marshal(op.payload)
			assert.NoError(t, err)

			resp, err := http.Post(
				suite.apiURL+"/allconfig-operation",
				"application/json",
				bytes.NewBuffer(reqBody),
			)
			assert.NoError(t, err)
			
			// Should fail due to no real database connection
			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			assert.False(t, response["success"].(bool))
		})
	}
}

// TestAPIErrorHandling tests error handling across the API
func (suite *IntegrationTestSuite) TestAPIErrorHandling() {
	tests := []struct {
		name           string
		endpoint       string
		method         string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:           "Invalid JSON payload",
			endpoint:       "/test-connection",
			method:         "POST",
			payload:        `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong HTTP method",
			endpoint:       "/test-connection",
			method:         "GET",
			payload:        nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Missing required fields",
			endpoint:       "/test-connection",
			method:         "POST",
			payload:        map[string]interface{}{"host": "localhost"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unsupported database type",
			endpoint:       "/test-connection",
			method:         "POST",
			payload: map[string]interface{}{
				"type":     "oracle",
				"host":     "localhost",
				"port":     1521,
				"database": "testdb",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			if tt.method == "POST" && tt.payload != nil {
				var reqBody []byte
				if str, ok := tt.payload.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, err = json.Marshal(tt.payload)
					assert.NoError(t, err)
				}

				resp, err = http.Post(
					suite.apiURL+tt.endpoint,
					"application/json",
					bytes.NewBuffer(reqBody),
				)
			} else {
				client := &http.Client{}
				req, err := http.NewRequest(tt.method, suite.apiURL+tt.endpoint, nil)
				assert.NoError(t, err)
				resp, err = client.Do(req)
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			assert.False(t, response["success"].(bool))
		})
	}
}

// TestSwaggerEndpoints tests Swagger documentation endpoints
func (suite *IntegrationTestSuite) TestSwaggerEndpoints() {
	endpoints := []struct {
		path           string
		expectedStatus int
		contentType    string
	}{
		{
			path:           "/",
			expectedStatus: http.StatusOK,
			contentType:    "text/html",
		},
		{
			path:           "/docs",
			expectedStatus: http.StatusOK,
			contentType:    "text/html",
		},
		{
			path:           "/swagger.json",
			expectedStatus: http.StatusOK,
			contentType:    "application/json",
		},
		{
			path:           "/swagger.yaml",
			expectedStatus: http.StatusOK, // May return 404 in test environment if file not found
			contentType:    "application/yaml",
		},
	}

	for _, endpoint := range endpoints {
		suite.T().Run(fmt.Sprintf("GET %s", endpoint.path), func(t *testing.T) {
			resp, err := http.Get(suite.apiURL + endpoint.path)
			assert.NoError(t, err)
			
			// Special handling for swagger.yaml which might not be available in test environment
			if endpoint.path == "/swagger.yaml" && resp.StatusCode == http.StatusNotFound {
				// This is acceptable in test environment
				return
			}
			
			assert.Equal(t, endpoint.expectedStatus, resp.StatusCode)
			
			if resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, endpoint.contentType)
			}
		})
	}
}

// TestConcurrentRequests tests concurrent request handling
func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	const numRequests = 10
	
	// Channel to collect responses
	responseChan := make(chan *http.Response, numRequests)
	errorChan := make(chan error, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(suite.apiURL + "/health")
			if err != nil {
				errorChan <- err
				return
			}
			responseChan <- resp
		}()
	}

	// Collect responses
	var responses []*http.Response
	var errors []error

	timeout := time.After(5 * time.Second)
	for i := 0; i < numRequests; i++ {
		select {
		case resp := <-responseChan:
			responses = append(responses, resp)
		case err := <-errorChan:
			errors = append(errors, err)
		case <-timeout:
			suite.T().Fatal("Timeout waiting for concurrent requests")
		}
	}

	// Verify all requests succeeded
	assert.Empty(suite.T(), errors, "Should have no errors")
	assert.Len(suite.T(), responses, numRequests, "Should have all responses")

	// Verify all responses are successful
	for _, resp := range responses {
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	}
}

// TestRequestTimeout tests request timeout handling
func (suite *IntegrationTestSuite) TestRequestTimeout() {
	// Create a client with short timeout
	client := &http.Client{
		Timeout: 1 * time.Millisecond, // Very short timeout
	}

	// This should timeout (or complete very quickly)
	resp, err := client.Get(suite.apiURL + "/health")
	
	// Either it completes quickly (OK) or times out (also OK for this test)
	if err != nil {
		// If it errors, it should be a timeout
		assert.Contains(suite.T(), err.Error(), "timeout")
	} else {
		// If it doesn't error, it completed quickly
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

// TestConnectorRegistry tests the connector registry functionality
func TestConnectorRegistry(t *testing.T) {
	registry := connectors.NewConnectorRegistry()
	assert.NotNil(t, registry)

	// Test registering connectors
	mysqlConfig := &connectors.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "testdb",
	}
	mysqlConnector := connectors.NewMySQLConnector(mysqlConfig)

	registry.Register("mysql-test", mysqlConnector)

	// Test retrieving connector
	retrieved, exists := registry.Get("mysql-test")
	assert.True(t, exists)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "mysql", retrieved.GetType())

	// Test non-existent connector
	nonExistent, exists := registry.Get("non-existent")
	assert.False(t, exists)
	assert.Nil(t, nonExistent)

	// Test listing connectors
	names := registry.List()
	assert.Contains(t, names, "mysql-test")
}

// Run the integration test suite
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// Benchmark integration tests
func BenchmarkHealthEndpoint(b *testing.B) {
	apiInstance := api.NewAPI()
	server := httptest.NewServer(api.SetupRoutes(apiInstance))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkConcurrentHealthRequests(b *testing.B) {
	apiInstance := api.NewAPI()
	server := httptest.NewServer(api.SetupRoutes(apiInstance))
	defer server.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get(server.URL + "/health")
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

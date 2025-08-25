package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockDBConnector implements the DBConnector interface for testing
type MockDBConnector struct {
	mock.Mock
}

func (m *MockDBConnector) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDBConnector) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDBConnector) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDBConnector) GetType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDBConnector) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(*sql.Rows), mockArgs.Error(1)
}

func (m *MockDBConnector) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, operation, params)
	return args.Get(0), args.Error(1)
}

func (m *MockDBConnector) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// APITestSuite defines the test suite for API handlers
type APITestSuite struct {
	suite.Suite
	api        *API
	mockConn   *MockDBConnector
	server     *httptest.Server
}

// SetupSuite sets up the test suite
func (suite *APITestSuite) SetupSuite() {
	suite.api = NewAPI()
	suite.mockConn = new(MockDBConnector)
}

// SetupTest sets up each test
func (suite *APITestSuite) SetupTest() {
	suite.mockConn = new(MockDBConnector)
}

// TearDownTest cleans up after each test
func (suite *APITestSuite) TearDownTest() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// TestNewAPI tests the creation of a new API instance
func (suite *APITestSuite) TestNewAPI() {
	api := NewAPI()
	assert.NotNil(suite.T(), api)
}

// TestHealthHandler tests the health check endpoint
func (suite *APITestSuite) TestHealthHandler() {
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(suite.T(), err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(suite.api.HealthHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Equal(suite.T(), "Service is healthy", response["message"])
}

// TestValidateConnectionRequest tests connection request validation
func TestValidateConnectionRequest(t *testing.T) {
	api := NewAPI()

	tests := []struct {
		name    string
		request DatabaseConnectionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mysql request",
			request: DatabaseConnectionRequest{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "valid postgresql request",
			request: DatabaseConnectionRequest{
				Type:     "postgresql",
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "valid mongodb request",
			request: DatabaseConnectionRequest{
				Type:     "mongodb",
				Host:     "localhost",
				Port:     27017,
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "empty type",
			request: DatabaseConnectionRequest{
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "database type is required",
		},
		{
			name: "unsupported type",
			request: DatabaseConnectionRequest{
				Type:     "oracle",
				Host:     "localhost",
				Port:     1521,
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "unsupported database type",
		},
		{
			name: "empty host",
			request: DatabaseConnectionRequest{
				Type:     "mysql",
				Port:     3306,
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name: "invalid port",
			request: DatabaseConnectionRequest{
				Type:     "mysql",
				Host:     "localhost",
				Port:     0,
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "valid port is required",
		},
		{
			name: "empty database",
			request: DatabaseConnectionRequest{
				Type: "mysql",
				Host: "localhost",
				Port: 3306,
			},
			wantErr: true,
			errMsg:  "database name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.validateConnectionRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCreateConnector tests the connector creation
func TestCreateConnector(t *testing.T) {
	api := NewAPI()

	tests := []struct {
		name         string
		request      DatabaseConnectionRequest
		expectedType string
		wantErr      bool
	}{
		{
			name: "create mysql connector",
			request: DatabaseConnectionRequest{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			expectedType: "mysql",
			wantErr:      false,
		},
		{
			name: "create postgresql connector",
			request: DatabaseConnectionRequest{
				Type:     "postgresql",
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expectedType: "postgresql",
			wantErr:      false,
		},
		{
			name: "create mongodb connector",
			request: DatabaseConnectionRequest{
				Type:     "mongodb",
				Host:     "localhost",
				Port:     27017,
				Database: "testdb",
			},
			expectedType: "mongodb",
			wantErr:      false,
		},
		{
			name: "unsupported type",
			request: DatabaseConnectionRequest{
				Type:     "oracle",
				Host:     "localhost",
				Port:     1521,
				Database: "testdb",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, err := api.createConnector(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, connector)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, connector)
				assert.Equal(t, tt.expectedType, connector.GetType())
			}
		})
	}
}

// TestTestConnectionHandler tests the test connection endpoint
func (suite *APITestSuite) TestTestConnectionHandler() {
	tests := []struct {
		name           string
		method         string
		body           DatabaseConnectionRequest
		setupMock      func()
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "connection test (will fail without real DB)",
			method: "POST",
			body: DatabaseConnectionRequest{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			setupMock: func() {
				// Mock will be set up in the actual handler test
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:           "wrong method",
			method:         "GET",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  true,
		},
		{
			name:   "invalid JSON",
			method: "POST",
			body: DatabaseConnectionRequest{
				Type: "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.method == "POST" && tt.body.Type != "" {
				jsonBody, _ := json.Marshal(tt.body)
				req, err = http.NewRequest(tt.method, "/test-connection", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
			} else if tt.method == "POST" {
				req, err = http.NewRequest(tt.method, "/test-connection", bytes.NewBuffer([]byte(`{"invalid": json`)))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, "/test-connection", nil)
			}

			assert.NoError(t, err)

			if tt.setupMock != nil {
				tt.setupMock()
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(suite.api.TestConnectionHandler)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			json.Unmarshal(rr.Body.Bytes(), &response)

			if tt.expectedError {
				assert.False(t, response["success"].(bool))
			} else {
				assert.True(t, response["success"].(bool))
			}
		})
	}
}

// TestAllConfigRequest tests AllConfig request structures
func TestAllConfigRequest(t *testing.T) {
	tests := []struct {
		name    string
		request AllConfigRequest
		valid   bool
	}{
		{
			name: "valid allconfig request",
			request: AllConfigRequest{
				DatabaseConnectionRequest: DatabaseConnectionRequest{
					Type:     "mysql",
					Host:     "localhost",
					Port:     3306,
					Username: "root",
					Password: "password",
					Database: "testdb",
				},
				TableName: "allconfig",
			},
			valid: true,
		},
		{
			name: "allconfig request with default table name",
			request: AllConfigRequest{
				DatabaseConnectionRequest: DatabaseConnectionRequest{
					Type:     "mongodb",
					Host:     "localhost",
					Port:     27017,
					Database: "testdb",
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the request structure is properly formed
			assert.NotNil(t, tt.request.DatabaseConnectionRequest)
			if tt.valid {
				assert.NotEmpty(t, tt.request.Type)
				assert.NotEmpty(t, tt.request.Host)
				assert.Greater(t, tt.request.Port, 0)
				assert.NotEmpty(t, tt.request.Database)
			}
		})
	}
}

// TestAllConfigOperationRequest tests AllConfig operation request structures
func TestAllConfigOperationRequest(t *testing.T) {
	tests := []struct {
		name    string
		request AllConfigOperationRequest
		valid   bool
	}{
		{
			name: "valid create operation",
			request: AllConfigOperationRequest{
				AllConfigRequest: AllConfigRequest{
					DatabaseConnectionRequest: DatabaseConnectionRequest{
						Type:     "mysql",
						Host:     "localhost",
						Port:     3306,
						Username: "root",
						Password: "password",
						Database: "testdb",
					},
					TableName: "allconfig",
				},
				Operation:   "create",
				Key:         "setting1",
				Value:       "value1",
				Description: "Test setting",
			},
			valid: true,
		},
		{
			name: "valid read operation",
			request: AllConfigOperationRequest{
				AllConfigRequest: AllConfigRequest{
					DatabaseConnectionRequest: DatabaseConnectionRequest{
						Type:     "mongodb",
						Host:     "localhost",
						Port:     27017,
						Database: "testdb",
					},
					TableName: "allconfig",
				},
				Operation: "read_all",
				Limit:     10,
				Offset:    0,
			},
			valid: true,
		},
		{
			name: "maker-checker operation",
			request: AllConfigOperationRequest{
				AllConfigRequest: AllConfigRequest{
					DatabaseConnectionRequest: DatabaseConnectionRequest{
						Type:     "postgresql",
						Host:     "localhost",
						Port:     5432,
						Username: "postgres",
						Password: "password",
						Database: "testdb",
					},
					TableName: "allconfig",
				},
				Operation: "submit_create",
				Key:       "new_setting",
				Value:     "new_value",
				MakerID:   "user123",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.request.Operation)
			assert.NotNil(t, tt.request.AllConfigRequest)
			if tt.valid {
				assert.NotEmpty(t, tt.request.Type)
				assert.NotEmpty(t, tt.request.Host)
				assert.Greater(t, tt.request.Port, 0)
				assert.NotEmpty(t, tt.request.Database)
			}
		})
	}
}

// TestHelperMethods tests utility methods
func TestHelperMethods(t *testing.T) {
	api := NewAPI()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "sendSuccess creates valid response",
			testFunc: func(t *testing.T) {
				rr := httptest.NewRecorder()
				data := map[string]interface{}{"key": "value"}
				message := "Success message"

				api.sendSuccess(rr, data, message)

				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Equal(t, message, response["message"])
				assert.NotNil(t, response["data"])
				assert.NotNil(t, response["timestamp"])
			},
		},
		{
			name: "sendError creates valid error response",
			testFunc: func(t *testing.T) {
				rr := httptest.NewRecorder()
				message := "Error message"

				api.sendError(rr, http.StatusBadRequest, message)

				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["success"].(bool))
				assert.Equal(t, message, response["error"])
				assert.NotNil(t, response["timestamp"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// TestConfigItem tests the ConfigItem structure
func TestConfigItem(t *testing.T) {
	tests := []struct {
		name string
		item ConfigItem
	}{
		{
			name: "valid config item",
			item: ConfigItem{
				Key:         "setting1",
				Value:       "value1",
				Description: "Test setting",
				MakerID:     "user123",
			},
		},
		{
			name: "config item with different value types",
			item: ConfigItem{
				Key:         "numeric_setting",
				Value:       42,
				Description: "Numeric setting",
				MakerID:     "user456",
			},
		},
		{
			name: "config item with boolean value",
			item: ConfigItem{
				Key:         "boolean_setting",
				Value:       true,
				Description: "Boolean setting",
				MakerID:     "user789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.item.Key)
			assert.NotNil(t, tt.item.Value)
			assert.NotEmpty(t, tt.item.MakerID)
		})
	}
}

// TestApprovalRequest tests the ApprovalRequest structure
func TestApprovalRequest(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name    string
		request ApprovalRequest
	}{
		{
			name: "valid approval request",
			request: ApprovalRequest{
				RequestID:    "req123",
				ConfigKey:    "setting1",
				ConfigValue:  "value1",
				Description:  "Test setting",
				Operation:    "create",
				MakerID:      "user123",
				Status:       "pending",
				RequestedAt:  now,
			},
		},
		{
			name: "approved request",
			request: ApprovalRequest{
				RequestID:    "req456",
				ConfigKey:    "setting2",
				ConfigValue:  "value2",
				Operation:    "update",
				MakerID:      "user123",
				CheckerID:    "checker456",
				Status:       "approved",
				RequestedAt:  now,
				ProcessedAt:  &now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.request.RequestID)
			assert.NotEmpty(t, tt.request.ConfigKey)
			assert.NotEmpty(t, tt.request.Operation)
			assert.NotEmpty(t, tt.request.MakerID)
			assert.NotEmpty(t, tt.request.Status)
			assert.False(t, tt.request.RequestedAt.IsZero())
		})
	}
}

// Run the test suite
func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

// Benchmark tests
func BenchmarkHealthHandler(b *testing.B) {
	api := NewAPI()
	req, _ := http.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(api.HealthHandler)
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkValidateConnectionRequest(b *testing.B) {
	api := NewAPI()
	request := &DatabaseConnectionRequest{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "testdb",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = api.validateConnectionRequest(request)
	}
}

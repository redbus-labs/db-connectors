package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"db-connectors/connectors"
)

// DatabaseConnectionRequest represents the request to connect to a database
type DatabaseConnectionRequest struct {
	Type     string `json:"type" validate:"required"`     // mysql, postgresql, mongodb
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required"`
	Username string `json:"username"`                     // Optional for MongoDB
	Password string `json:"password"`                     // Optional for MongoDB
	Database string `json:"database" validate:"required"`
	SSLMode  string `json:"ssl_mode,omitempty"` // For PostgreSQL
}

// DatabaseOperationRequest represents a request to execute a database operation
type DatabaseOperationRequest struct {
	DatabaseConnectionRequest
	Operation string                 `json:"operation" validate:"required"` // query, insert, update, delete, find, etc.
	Query     string                 `json:"query,omitempty"`               // For SQL databases
	Args      []interface{}          `json:"args,omitempty"`                // Query arguments for SQL
	Params    map[string]interface{} `json:"params,omitempty"`              // For MongoDB operations
}

// AllConfigRequest represents a request to work with allconfig table
type AllConfigRequest struct {
	DatabaseConnectionRequest
	TableName string `json:"table_name,omitempty"` // Custom table name for allconfig, defaults to "allconfig"
}

// AllConfigOperationRequest represents operations on allconfig table
type AllConfigOperationRequest struct {
	AllConfigRequest
	Operation   string                 `json:"operation" validate:"required"` // CRUD operations
	Key         string                 `json:"key,omitempty"`                 // Configuration key
	Value       interface{}            `json:"value,omitempty"`               // Configuration value
	Description string                 `json:"description,omitempty"`         // Configuration description
	Configs     map[string]interface{} `json:"configs,omitempty"`             // Multiple configurations
	// For batch operations
	ConfigItems []ConfigItem `json:"config_items,omitempty"` // Array of config items for batch operations
	// For search/filter operations
	SearchTerm string                 `json:"search_term,omitempty"` // Search term for filtering
	Filter     map[string]interface{} `json:"filter,omitempty"`      // Filter criteria
	Limit      int                    `json:"limit,omitempty"`       // Limit results
	Offset     int                    `json:"offset,omitempty"`      // Offset for pagination
	// For maker-checker workflow
	MakerID         string `json:"maker_id,omitempty"`         // ID of user making the change
	CheckerID       string `json:"checker_id,omitempty"`       // ID of user approving the change
	ApprovalComment string `json:"approval_comment,omitempty"` // Comment for approval/rejection
	RequestID       string `json:"request_id,omitempty"`       // ID of pending request for approval
}

// ConfigItem represents a single configuration item
type ConfigItem struct {
	Key         string      `json:"key" validate:"required"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
	// For maker-checker workflow
	MakerID string `json:"maker_id,omitempty"`
}

// ApprovalRequest represents a pending approval request
type ApprovalRequest struct {
	RequestID       string      `json:"request_id"`
	ConfigKey       string      `json:"config_key"`
	ConfigValue     interface{} `json:"config_value"`
	Description     string      `json:"description,omitempty"`
	Operation       string      `json:"operation"`        // create, update, delete
	MakerID         string      `json:"maker_id"`
	CheckerID       string      `json:"checker_id,omitempty"`
	Status          string      `json:"status"`           // pending, approved, rejected
	RequestedAt     time.Time   `json:"requested_at"`
	ProcessedAt     *time.Time  `json:"processed_at,omitempty"`
	ApprovalComment string      `json:"approval_comment,omitempty"`
	PreviousValue   interface{} `json:"previous_value,omitempty"` // For update operations
}

// DatabaseResponse represents the response from database operations
type DatabaseResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// API represents the HTTP API server
type API struct {
	registry *connectors.ConnectorRegistry
}

// NewAPI creates a new API instance
func NewAPI() *API {
	return &API{
		registry: connectors.NewConnectorRegistry(),
	}
}

// TestConnectionHandler tests a database connection
func (a *API) TestConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req DatabaseConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Validate request
	if err := a.validateConnectionRequest(&req); err != nil {
		a.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create connector
	connector, err := a.createConnector(&req)
	if err != nil {
		a.sendError(w, http.StatusBadRequest, fmt.Sprintf("Failed to create connector: %v", err))
		return
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := connector.Connect(ctx); err != nil {
		a.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Connection failed: %v", err))
		return
	}
	defer connector.Close()

	if err := connector.Ping(ctx); err != nil {
		a.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Ping failed: %v", err))
		return
	}

	a.sendSuccess(w, map[string]interface{}{
		"connection_status": "success",
		"database_type":     connector.GetType(),
		"connected":         connector.IsConnected(),
	}, "Database connection successful")
}

// ExecuteOperationHandler executes a database operation
func (a *API) ExecuteOperationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req DatabaseOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Validate request
	if err := a.validateConnectionRequest(&req.DatabaseConnectionRequest); err != nil {
		a.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Operation == "" {
		a.sendError(w, http.StatusBadRequest, "Operation is required")
		return
	}

	// Create connector
	connector, err := a.createConnector(&req.DatabaseConnectionRequest)
	if err != nil {
		a.sendError(w, http.StatusBadRequest, fmt.Sprintf("Failed to create connector: %v", err))
		return
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := connector.Connect(ctx); err != nil {
		a.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Connection failed: %v", err))
		return
	}
	defer connector.Close()

	// Execute operation
	result, err := a.executeOperation(ctx, connector, &req)
	if err != nil {
		a.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Operation failed: %v", err))
		return
	}

	a.sendSuccess(w, result, "Operation executed successfully")
}

// HealthHandler provides health check endpoint
func (a *API) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	a.sendSuccess(w, map[string]interface{}{
		"status":  "healthy",
		"service": "db-connectors-api",
		"version": "1.0.0",
	}, "Service is healthy")
}

// AllConfigHandler checks for allconfig table and provides information
func (a *API) AllConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req AllConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Set default table name if not provided
	if req.TableName == "" {
		req.TableName = "allconfig"
	}

	// Validate connection request
	if err := a.validateConnectionRequest(&req.DatabaseConnectionRequest); err != nil {
		a.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create connector
	connector, err := a.createConnector(&req.DatabaseConnectionRequest)
	if err != nil {
		a.sendError(w, http.StatusBadRequest, fmt.Sprintf("Failed to create connector: %v", err))
		return
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := connector.Connect(ctx); err != nil {
		a.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Connection failed: %v", err))
		return
	}
	defer connector.Close()

	// Check if allconfig table exists
	exists, err := a.checkTableExists(ctx, connector, req.Database, req.TableName)
	if err != nil {
		a.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check table existence: %v", err))
		return
	}

	response := map[string]interface{}{
		"table_name":   req.TableName,
		"table_exists": exists,
		"database_type": connector.GetType(),
	}

	if exists {
		// Get table structure
		structure, err := a.getTableStructure(ctx, connector, req.Database, req.TableName)
		if err != nil {
			response["warning"] = fmt.Sprintf("Table exists but couldn't get structure: %v", err)
		} else {
			response["table_structure"] = structure
		}

		// Get config count
		count, err := a.getConfigCount(ctx, connector, req.TableName)
		if err != nil {
			response["warning"] = fmt.Sprintf("Table exists but couldn't count records: %v", err)
		} else {
			response["config_count"] = count
		}
	} else {
		response["create_table_sql"] = a.getCreateTableSQL(connector.GetType(), req.TableName)
	}

	a.sendSuccess(w, response, "AllConfig table check completed")
}

// AllConfigOperationHandler handles operations on allconfig table
func (a *API) AllConfigOperationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req AllConfigOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Set default table name if not provided
	if req.TableName == "" {
		req.TableName = "allconfig"
	}

	// Validate connection request
	if err := a.validateConnectionRequest(&req.DatabaseConnectionRequest); err != nil {
		a.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Operation == "" {
		a.sendError(w, http.StatusBadRequest, "Operation is required")
		return
	}

	// Create connector
	connector, err := a.createConnector(&req.DatabaseConnectionRequest)
	if err != nil {
		a.sendError(w, http.StatusBadRequest, fmt.Sprintf("Failed to create connector: %v", err))
		return
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := connector.Connect(ctx); err != nil {
		a.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Connection failed: %v", err))
		return
	}
	defer connector.Close()

	// Execute allconfig operation
	result, err := a.executeAllConfigOperation(ctx, connector, &req)
	if err != nil {
		a.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Operation failed: %v", err))
		return
	}

	a.sendSuccess(w, result, fmt.Sprintf("AllConfig operation '%s' completed", req.Operation))
}

// Helper methods

func (a *API) validateConnectionRequest(req *DatabaseConnectionRequest) error {
	if req.Type == "" {
		return fmt.Errorf("database type is required")
	}
	if req.Type != "mysql" && req.Type != "postgresql" && req.Type != "mongodb" {
		return fmt.Errorf("unsupported database type: %s", req.Type)
	}
	if req.Host == "" {
		return fmt.Errorf("host is required")
	}
	if req.Port <= 0 {
		return fmt.Errorf("valid port is required")
	}
	if req.Database == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

func (a *API) createConnector(req *DatabaseConnectionRequest) (connectors.DBConnector, error) {
	config := &connectors.ConnectionConfig{
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Database: req.Database,
		SSLMode:  req.SSLMode,
	}

	switch req.Type {
	case "mysql":
		return connectors.NewMySQLConnector(config), nil
	case "postgresql":
		return connectors.NewPostgreSQLConnector(config), nil
	case "mongodb":
		return connectors.NewMongoDBConnector(config), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", req.Type)
	}
}

func (a *API) executeOperation(ctx context.Context, connector connectors.DBConnector, req *DatabaseOperationRequest) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		return a.executeSQLOperation(ctx, connector, req)
	case "mongodb":
		return a.executeMongoOperation(ctx, connector, req)
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) executeSQLOperation(ctx context.Context, connector connectors.DBConnector, req *DatabaseOperationRequest) (interface{}, error) {
	switch req.Operation {
	case "query", "select":
		if req.Query == "" {
			return nil, fmt.Errorf("query is required for SQL select operation")
		}
		
		rows, err := connector.Query(ctx, req.Query, req.Args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		return a.rowsToMap(rows)
		
	case "insert", "update", "delete", "execute":
		if req.Query == "" {
			return nil, fmt.Errorf("query is required for SQL operation")
		}
		
		return connector.Execute(ctx, req.Operation, map[string]interface{}{
			"query": req.Query,
			"args":  req.Args,
		})
		
	default:
		return nil, fmt.Errorf("unsupported SQL operation: %s", req.Operation)
	}
}

func (a *API) executeMongoOperation(ctx context.Context, connector connectors.DBConnector, req *DatabaseOperationRequest) (interface{}, error) {
	if req.Params == nil {
		req.Params = make(map[string]interface{})
	}

	return connector.Execute(ctx, req.Operation, req.Params)
}

func (a *API) rowsToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
		}
		results = append(results, row)
	}

	return results, nil
}

func (a *API) sendSuccess(w http.ResponseWriter, data interface{}, message string) {
	response := DatabaseResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
	a.sendJSON(w, http.StatusOK, response)
}

func (a *API) sendError(w http.ResponseWriter, statusCode int, errorMsg string) {
	response := DatabaseResponse{
		Success:   false,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}
	a.sendJSON(w, statusCode, response)
}

func (a *API) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// AllConfig helper functions

func (a *API) checkTableExists(ctx context.Context, connector connectors.DBConnector, databaseName string, tableName string) (bool, error) {
	switch connector.GetType() {
	case "mysql":
		// Use specific database name from API input instead of DATABASE()
		query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?"
		rows, err := connector.Query(ctx, query, databaseName, tableName)
		if err != nil {
			return false, fmt.Errorf("failed to check table existence in MySQL: %w", err)
		}
		defer rows.Close()
		
		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				return false, fmt.Errorf("failed to scan table count: %w", err)
			}
			return count > 0, nil
		}
		return false, nil
		
	case "postgresql":
		// For PostgreSQL, check in the specified database
		// If databaseName is provided, use it as schema, otherwise use 'public'
		schema := "public"
		if databaseName != "" {
			// In PostgreSQL, we can check if a specific schema exists and use it
			schemaCheckQuery := "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = $1"
			schemaRows, err := connector.Query(ctx, schemaCheckQuery, databaseName)
			if err == nil {
				defer schemaRows.Close()
				if schemaRows.Next() {
					var schemaCount int
					if err := schemaRows.Scan(&schemaCount); err == nil && schemaCount > 0 {
						schema = databaseName
					}
				}
			}
		}
		
		query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2"
		rows, err := connector.Query(ctx, query, schema, tableName)
		if err != nil {
			return false, fmt.Errorf("failed to check table existence in PostgreSQL: %w", err)
		}
		defer rows.Close()
		
		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				return false, fmt.Errorf("failed to scan table count: %w", err)
			}
			return count > 0, nil
		}
		return false, nil
		
	case "mongodb":
		// For MongoDB, we need to check collections in the specific database
		// MongoDB client can access multiple databases, so we'll pass the database name as a parameter
		params := map[string]interface{}{
			"filter": map[string]interface{}{"name": tableName},
		}
		
		// If a specific database name is provided, include it in the params
		if databaseName != "" {
			params["database"] = databaseName
		}
		
		result, err := connector.Execute(ctx, "listCollections", params)
		if err != nil {
			return false, fmt.Errorf("failed to check collection existence in MongoDB: %w", err)
		}
		
		// Handle different result types from MongoDB
		switch v := result.(type) {
		case []interface{}:
			return len(v) > 0, nil
		case []map[string]interface{}:
			return len(v) > 0, nil
		case map[string]interface{}:
			// If it's a map, check if it has collections array
			if cols, exists := v["collections"]; exists {
				if colArray, ok := cols.([]interface{}); ok {
					return len(colArray) > 0, nil
				}
			}
			return false, nil
		default:
			return false, fmt.Errorf("unexpected result type from MongoDB listCollections: %T", v)
		}
		
	default:
		return false, fmt.Errorf("unsupported database type: %s", connector.GetType())
	}
}

func (a *API) getTableStructure(ctx context.Context, connector connectors.DBConnector, databaseName string, tableName string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		// For MySQL, we can use the database.table format or just table if connected to right database
		var query string
		if databaseName != "" {
			query = "DESCRIBE " + databaseName + "." + tableName
		} else {
			query = "DESCRIBE " + tableName
		}
		rows, err := connector.Query(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to get table structure for MySQL: %w", err)
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "postgresql":
		// For PostgreSQL, check in the specified schema
		schema := "public"
		if databaseName != "" {
			// Check if the schema exists
			schemaCheckQuery := "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = $1"
			schemaRows, err := connector.Query(ctx, schemaCheckQuery, databaseName)
			if err == nil {
				defer schemaRows.Close()
				if schemaRows.Next() {
					var schemaCount int
					if err := schemaRows.Scan(&schemaCount); err == nil && schemaCount > 0 {
						schema = databaseName
					}
				}
			}
		}
		
		query := `SELECT column_name, data_type, is_nullable, column_default 
				  FROM information_schema.columns 
				  WHERE table_schema = $1 AND table_name = $2 ORDER BY ordinal_position`
		rows, err := connector.Query(ctx, query, schema, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get table structure for PostgreSQL: %w", err)
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		// For MongoDB, we'll sample documents to infer structure
		// The database name is already handled by the connection
		result, err := connector.Execute(ctx, "find", map[string]interface{}{
			"collection": tableName,
			"limit":      1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get collection structure for MongoDB: %w", err)
		}
		return result, nil
		
	default:
		return nil, fmt.Errorf("unsupported database type: %s", connector.GetType())
	}
}

func (a *API) getConfigCount(ctx context.Context, connector connectors.DBConnector, tableName string) (int64, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := "SELECT COUNT(*) FROM " + tableName
		rows, err := connector.Query(ctx, query)
		if err != nil {
			return 0, err
		}
		defer rows.Close()
		
		if rows.Next() {
			var count int64
			if err := rows.Scan(&count); err != nil {
				return 0, err
			}
			return count, nil
		}
		return 0, nil
		
	case "mongodb":
		result, err := connector.Execute(ctx, "count", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{},
		})
		if err != nil {
			return 0, err
		}
		
		if count, ok := result.(int64); ok {
			return count, nil
		}
		if count, ok := result.(int); ok {
			return int64(count), nil
		}
		return 0, nil
		
	default:
		return 0, fmt.Errorf("unsupported database type")
	}
}

func (a *API) getCreateTableSQL(dbType, tableName string) string {
	switch dbType {
	case "mysql":
		return fmt.Sprintf(`CREATE TABLE %s (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(255) NOT NULL UNIQUE,
    config_value TEXT,
    description TEXT,
    status ENUM('approved', 'pending', 'rejected') DEFAULT 'approved',
    maker_id VARCHAR(255),
    checker_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    approved_at TIMESTAMP NULL,
    approval_comment TEXT,
    INDEX idx_config_key (config_key),
    INDEX idx_status (status),
    INDEX idx_maker_id (maker_id)
);

CREATE TABLE %s_approval_requests (
    request_id VARCHAR(36) PRIMARY KEY,
    config_key VARCHAR(255) NOT NULL,
    config_value TEXT,
    description TEXT,
    operation ENUM('create', 'update', 'delete') NOT NULL,
    maker_id VARCHAR(255) NOT NULL,
    checker_id VARCHAR(255),
    status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending',
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    approval_comment TEXT,
    previous_value TEXT,
    INDEX idx_status (status),
    INDEX idx_maker_id (maker_id),
    INDEX idx_checker_id (checker_id),
    INDEX idx_config_key (config_key)
);`, tableName, tableName)
		
	case "postgresql":
		return fmt.Sprintf(`CREATE TABLE %s (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(255) NOT NULL UNIQUE,
    config_value TEXT,
    description TEXT,
    status VARCHAR(20) DEFAULT 'approved' CHECK (status IN ('approved', 'pending', 'rejected')),
    maker_id VARCHAR(255),
    checker_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP,
    approval_comment TEXT
);

CREATE TABLE %s_approval_requests (
    request_id VARCHAR(36) PRIMARY KEY,
    config_key VARCHAR(255) NOT NULL,
    config_value TEXT,
    description TEXT,
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('create', 'update', 'delete')),
    maker_id VARCHAR(255) NOT NULL,
    checker_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    approval_comment TEXT,
    previous_value TEXT
);

CREATE INDEX idx_%s_config_key ON %s (config_key);
CREATE INDEX idx_%s_status ON %s (status);
CREATE INDEX idx_%s_maker_id ON %s (maker_id);
CREATE INDEX idx_%s_approval_status ON %s_approval_requests (status);
CREATE INDEX idx_%s_approval_maker ON %s_approval_requests (maker_id);
CREATE INDEX idx_%s_approval_checker ON %s_approval_requests (checker_id);`, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName)
		
	case "mongodb":
		return fmt.Sprintf(`// MongoDB collection '%s' with sample document:
{
    "_id": ObjectId(),
    "config_key": "unique_key_name",
    "config_value": "configuration_value",
    "description": "Description of the configuration",
    "status": "approved",
    "maker_id": "user123",
    "checker_id": "admin456",
    "created_at": new Date(),
    "updated_at": new Date(),
    "approved_at": new Date(),
    "approval_comment": "Approved by checker"
}

// MongoDB collection '%s_approval_requests' with sample document:
{
    "_id": ObjectId(),
    "request_id": "uuid-string",
    "config_key": "configuration_key",
    "config_value": "new_value",
    "description": "Description",
    "operation": "create",
    "maker_id": "user123",
    "checker_id": "admin456",
    "status": "pending",
    "requested_at": new Date(),
    "processed_at": new Date(),
    "approval_comment": "Looks good",
    "previous_value": "old_value"
}

// Create indexes:
db.%s.createIndex({"config_key": 1}, {"unique": true});
db.%s.createIndex({"status": 1});
db.%s.createIndex({"maker_id": 1});
db.%s_approval_requests.createIndex({"request_id": 1}, {"unique": true});
db.%s_approval_requests.createIndex({"status": 1});
db.%s_approval_requests.createIndex({"maker_id": 1});
db.%s_approval_requests.createIndex({"config_key": 1});`, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName)
		
	default:
		return "Unsupported database type"
	}
}

func (a *API) executeAllConfigOperation(ctx context.Context, connector connectors.DBConnector, req *AllConfigOperationRequest) (interface{}, error) {
	switch req.Operation {
	// Table management
	case "create_table":
		return a.createAllConfigTable(ctx, connector, req.TableName)
	case "drop_table":
		return a.dropAllConfigTable(ctx, connector, req.TableName)
		
	// MAKER-CHECKER CREATE operations
	case "submit_create":
		if req.Key == "" || req.MakerID == "" {
			return nil, fmt.Errorf("config key and maker_id are required for submit_create operation")
		}
		return a.submitConfigForApproval(ctx, connector, req.TableName, "create", req.Key, req.Value, req.Description, req.MakerID, nil)
		
	case "submit_update":
		if req.Key == "" || req.MakerID == "" {
			return nil, fmt.Errorf("config key and maker_id are required for submit_update operation")
		}
		return a.submitConfigForApproval(ctx, connector, req.TableName, "update", req.Key, req.Value, req.Description, req.MakerID, nil)
		
	case "submit_delete":
		if req.Key == "" || req.MakerID == "" {
			return nil, fmt.Errorf("config key and maker_id are required for submit_delete operation")
		}
		return a.submitConfigForApproval(ctx, connector, req.TableName, "delete", req.Key, nil, req.Description, req.MakerID, nil)
		
	// CHECKER APPROVAL operations
	case "approve_request":
		if req.RequestID == "" || req.CheckerID == "" {
			return nil, fmt.Errorf("request_id and checker_id are required for approve_request operation")
		}
		return a.approveRequest(ctx, connector, req.Database, req.TableName, req.RequestID, req.CheckerID, req.ApprovalComment)
		
	case "reject_request":
		if req.RequestID == "" || req.CheckerID == "" {
			return nil, fmt.Errorf("request_id and checker_id are required for reject_request operation")
		}
		return a.rejectRequest(ctx, connector, req.Database, req.TableName, req.RequestID, req.CheckerID, req.ApprovalComment)
		
	case "get_pending_approvals":
		return a.getPendingApprovals(ctx, connector, req.TableName, req.Limit, req.Offset)
		
	case "get_my_requests":
		if req.MakerID == "" {
			return nil, fmt.Errorf("maker_id is required for get_my_requests operation")
		}
		return a.getMyRequests(ctx, connector, req.TableName, req.MakerID, req.Limit, req.Offset)
		
	case "get_approval_history":
		return a.getApprovalHistory(ctx, connector, req.TableName, req.Limit, req.Offset)
		
	// LEGACY DIRECT operations (bypass approval - for admin use)
	case "direct_create", "create", "set_config":
		if req.Key == "" {
			return nil, fmt.Errorf("config key is required for create operation")
		}
		return a.createConfigDirect(ctx, connector, req.Database, req.TableName, req.Key, req.Value, req.Description, req.MakerID)
		
	case "direct_create_batch", "create_batch", "set_multiple":
		if req.ConfigItems != nil && len(req.ConfigItems) > 0 {
			return a.createMultipleConfigsDirect(ctx, connector, req.Database, req.TableName, req.ConfigItems)
		}
		if req.Configs != nil && len(req.Configs) > 0 {
			return a.setMultipleConfigs(ctx, connector, req.TableName, req.Configs)
		}
		return nil, fmt.Errorf("config_items or configs are required for batch create operation")
		
	// READ operations (only show APPROVED configs)
	case "read", "get_config":
		if req.Key == "" {
			return nil, fmt.Errorf("config key is required for read operation")
		}
		return a.readApprovedConfig(ctx, connector, req.Database, req.TableName, req.Key)
		
	case "read_all", "get_all":
		return a.readAllApprovedConfigs(ctx, connector, req.Database, req.TableName, req.Limit, req.Offset)
		
	case "search":
		if req.SearchTerm == "" {
			return nil, fmt.Errorf("search_term is required for search operation")
		}
		return a.searchApprovedConfigs(ctx, connector, req.TableName, req.SearchTerm, req.Limit, req.Offset)
		
	case "filter":
		if req.Filter == nil || len(req.Filter) == 0 {
			return nil, fmt.Errorf("filter criteria is required for filter operation")
		}
		return a.filterApprovedConfigs(ctx, connector, req.TableName, req.Filter, req.Limit, req.Offset)
		
	// ADMIN READ operations (show ALL configs including pending)
	case "read_all_admin":
		return a.readAllConfigs(ctx, connector, req.TableName, req.Limit, req.Offset)
		
	case "search_admin":
		if req.SearchTerm == "" {
			return nil, fmt.Errorf("search_term is required for search operation")
		}
		return a.searchConfigs(ctx, connector, req.TableName, req.SearchTerm, req.Limit, req.Offset)
		
	// DIRECT UPDATE operations (bypass approval - for admin use)
	case "direct_update", "update":
		if req.Key == "" {
			return nil, fmt.Errorf("config key is required for update operation")
		}
		return a.updateConfigDirect(ctx, connector, req.Database, req.TableName, req.Key, req.Value, req.Description, req.MakerID)
		
	case "direct_update_batch", "update_batch":
		if req.ConfigItems == nil || len(req.ConfigItems) == 0 {
			return nil, fmt.Errorf("config_items are required for batch update operation")
		}
		return a.updateMultipleConfigsDirect(ctx, connector, req.Database, req.TableName, req.ConfigItems)
		
	// DIRECT DELETE operations (bypass approval - for admin use)
	case "direct_delete", "delete", "delete_config":
		if req.Key == "" {
			return nil, fmt.Errorf("config key is required for delete operation")
		}
		return a.deleteConfigDirect(ctx, connector, req.TableName, req.Key, req.MakerID)
		
	case "direct_delete_batch", "delete_batch":
		if req.ConfigItems == nil || len(req.ConfigItems) == 0 {
			return nil, fmt.Errorf("config_items with keys are required for batch delete operation")
		}
		return a.deleteMultipleConfigsDirect(ctx, connector, req.TableName, req.ConfigItems)
		
	case "direct_delete_all", "delete_all":
		return a.deleteAllConfigs(ctx, connector, req.TableName)
		
	// UTILITY operations
	case "count":
		return a.countApprovedConfigs(ctx, connector, req.TableName)
		
	case "count_admin":
		return a.countConfigs(ctx, connector, req.TableName)
		
	case "exists":
		if req.Key == "" {
			return nil, fmt.Errorf("config key is required for exists operation")
		}
		return a.configExistsApproved(ctx, connector, req.TableName, req.Key)
		
	default:
		return nil, fmt.Errorf("unsupported operation: %s. Supported operations: submit_create, submit_update, submit_delete, approve_request, reject_request, get_pending_approvals, get_my_requests, get_approval_history, read, read_all, search, filter, count, exists, create_table, drop_table", req.Operation)
	}
}

func (a *API) createAllConfigTable(ctx context.Context, connector connectors.DBConnector, tableName string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		sql := a.getCreateTableSQL(connector.GetType(), tableName)
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": sql,
		})
		
	case "mongodb":
		// For MongoDB, create the collection and index
		_, err := connector.Execute(ctx, "insert", map[string]interface{}{
			"collection": tableName,
			"document": map[string]interface{}{
				"config_key":   "_init",
				"config_value": "collection_created",
				"description":  "Initial document to create collection",
				"created_at":   time.Now(),
				"updated_at":   time.Now(),
			},
		})
		if err != nil {
			return nil, err
		}
		
		// Create unique index
		_, err = connector.Execute(ctx, "createIndex", map[string]interface{}{
			"collection": tableName,
			"index":      map[string]interface{}{"config_key": 1},
			"options":    map[string]interface{}{"unique": true},
		})
		
		return map[string]interface{}{
			"collection_created": true,
			"index_created":      err == nil,
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) getAllConfigs(ctx context.Context, connector connectors.DBConnector, tableName string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := "SELECT config_key, config_value, description, created_at, updated_at FROM " + tableName + " ORDER BY config_key"
		rows, err := connector.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		return connector.Execute(ctx, "find", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{},
			"sort":       map[string]interface{}{"config_key": 1},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) getConfig(ctx context.Context, connector connectors.DBConnector, tableName, key string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := "SELECT config_key, config_value, description, created_at, updated_at FROM " + tableName + " WHERE config_key = ?"
		rows, err := connector.Query(ctx, query, key)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "postgresql":
		query := "SELECT config_key, config_value, description, created_at, updated_at FROM " + tableName + " WHERE config_key = $1"
		rows, err := connector.Query(ctx, query, key)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		return connector.Execute(ctx, "findOne", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"config_key": key},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) setConfig(ctx context.Context, connector connectors.DBConnector, tableName, key string, value interface{}) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `INSERT INTO ` + tableName + ` (config_key, config_value, updated_at) 
				  VALUES (?, ?, NOW()) 
				  ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), updated_at = NOW()`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key, value},
		})
		
	case "postgresql":
		query := `INSERT INTO ` + tableName + ` (config_key, config_value, created_at, updated_at) 
				  VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				  ON CONFLICT (config_key) DO UPDATE SET 
				  config_value = EXCLUDED.config_value, updated_at = CURRENT_TIMESTAMP`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key, value},
		})
		
	case "mongodb":
		return connector.Execute(ctx, "upsert", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"config_key": key},
			"update": map[string]interface{}{
				"$set": map[string]interface{}{
					"config_key":   key,
					"config_value": value,
					"updated_at":   time.Now(),
				},
				"$setOnInsert": map[string]interface{}{
					"created_at": time.Now(),
				},
			},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) setMultipleConfigs(ctx context.Context, connector connectors.DBConnector, tableName string, configs map[string]interface{}) (interface{}, error) {
	results := make(map[string]interface{})
	
	for key, value := range configs {
		result, err := a.setConfig(ctx, connector, tableName, key, value)
		if err != nil {
			results[key] = map[string]interface{}{"error": err.Error()}
		} else {
			results[key] = map[string]interface{}{"success": true, "result": result}
		}
	}
	
	return results, nil
}

func (a *API) deleteConfig(ctx context.Context, connector connectors.DBConnector, tableName, key string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := "DELETE FROM " + tableName + " WHERE config_key = ?"
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key},
		})
		
	case "postgresql":
		query := "DELETE FROM " + tableName + " WHERE config_key = $1"
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key},
		})
		
	case "mongodb":
		return connector.Execute(ctx, "delete", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"config_key": key},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// Enhanced CRUD Operations

// CREATE operations
func (a *API) createConfig(ctx context.Context, connector connectors.DBConnector, tableName, key string, value interface{}, description string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `INSERT INTO ` + tableName + ` (config_key, config_value, description, created_at, updated_at) 
				  VALUES (?, ?, ?, NOW(), NOW())`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key, value, description},
		})
		
	case "postgresql":
		query := `INSERT INTO ` + tableName + ` (config_key, config_value, description, created_at, updated_at) 
				  VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key, value, description},
		})
		
	case "mongodb":
		return connector.Execute(ctx, "insert", map[string]interface{}{
			"collection": tableName,
			"document": map[string]interface{}{
				"config_key":   key,
				"config_value": value,
				"description":  description,
				"created_at":   time.Now(),
				"updated_at":   time.Now(),
			},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) createMultipleConfigs(ctx context.Context, connector connectors.DBConnector, tableName string, configs []ConfigItem) (interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	
	for _, config := range configs {
		result, err := a.createConfig(ctx, connector, tableName, config.Key, config.Value, config.Description)
		if err != nil {
			results[config.Key] = map[string]interface{}{"error": err.Error()}
		} else {
			results[config.Key] = map[string]interface{}{"success": true, "result": result}
			successCount++
		}
	}
	
	return map[string]interface{}{
		"total_items":    len(configs),
		"success_count":  successCount,
		"failure_count":  len(configs) - successCount,
		"results":        results,
	}, nil
}

// READ operations
func (a *API) readConfig(ctx context.Context, connector connectors.DBConnector, tableName, key string) (interface{}, error) {
	return a.getConfig(ctx, connector, tableName, key)
}

func (a *API) readAllConfigs(ctx context.Context, connector connectors.DBConnector, tableName string, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := "SELECT config_key, config_value, description, created_at, updated_at FROM " + tableName + " ORDER BY config_key"
		
		if limit > 0 {
			if connector.GetType() == "mysql" {
				query += fmt.Sprintf(" LIMIT %d", limit)
				if offset > 0 {
					query += fmt.Sprintf(" OFFSET %d", offset)
				}
			} else { // postgresql
				query += fmt.Sprintf(" LIMIT %d", limit)
				if offset > 0 {
					query += fmt.Sprintf(" OFFSET %d", offset)
				}
			}
		}
		
		rows, err := connector.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{},
			"sort":       map[string]interface{}{"config_key": 1},
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) searchConfigs(ctx context.Context, connector connectors.DBConnector, tableName, searchTerm string, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `SELECT config_key, config_value, description, created_at, updated_at FROM ` + tableName + ` 
				  WHERE config_key LIKE ? OR config_value LIKE ? OR description LIKE ? 
				  ORDER BY config_key`
		searchPattern := "%" + searchTerm + "%"
		args := []interface{}{searchPattern, searchPattern, searchPattern}
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "postgresql":
		query := `SELECT config_key, config_value, description, created_at, updated_at FROM ` + tableName + ` 
				  WHERE config_key ILIKE $1 OR config_value ILIKE $2 OR description ILIKE $3 
				  ORDER BY config_key`
		searchPattern := "%" + searchTerm + "%"
		args := []interface{}{searchPattern, searchPattern, searchPattern}
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName,
			"filter": map[string]interface{}{
				"$or": []map[string]interface{}{
					{"config_key": map[string]interface{}{"$regex": searchTerm, "$options": "i"}},
					{"config_value": map[string]interface{}{"$regex": searchTerm, "$options": "i"}},
					{"description": map[string]interface{}{"$regex": searchTerm, "$options": "i"}},
				},
			},
			"sort": map[string]interface{}{"config_key": 1},
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) filterConfigs(ctx context.Context, connector connectors.DBConnector, tableName string, filter map[string]interface{}, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		// Build WHERE clause from filter
		whereClause := "WHERE 1=1"
		args := []interface{}{}
		paramIndex := 1
		
		for key, value := range filter {
			if connector.GetType() == "mysql" {
				whereClause += fmt.Sprintf(" AND %s = ?", key)
			} else {
				whereClause += fmt.Sprintf(" AND %s = $%d", key, paramIndex)
				paramIndex++
			}
			args = append(args, value)
		}
		
		query := fmt.Sprintf("SELECT config_key, config_value, description, created_at, updated_at FROM %s %s ORDER BY config_key", tableName, whereClause)
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName,
			"filter":     filter,
			"sort":       map[string]interface{}{"config_key": 1},
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// UPDATE operations
func (a *API) updateConfig(ctx context.Context, connector connectors.DBConnector, tableName, key string, value interface{}, description string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `UPDATE ` + tableName + ` SET config_value = ?, description = ?, updated_at = NOW() WHERE config_key = ?`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{value, description, key},
		})
		
	case "postgresql":
		query := `UPDATE ` + tableName + ` SET config_value = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE config_key = $3`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{value, description, key},
		})
		
	case "mongodb":
		return connector.Execute(ctx, "update", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"config_key": key},
			"update": map[string]interface{}{
				"$set": map[string]interface{}{
					"config_value": value,
					"description":  description,
					"updated_at":   time.Now(),
				},
			},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) updateMultipleConfigs(ctx context.Context, connector connectors.DBConnector, tableName string, configs []ConfigItem) (interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	
	for _, config := range configs {
		result, err := a.updateConfig(ctx, connector, tableName, config.Key, config.Value, config.Description)
		if err != nil {
			results[config.Key] = map[string]interface{}{"error": err.Error()}
		} else {
			results[config.Key] = map[string]interface{}{"success": true, "result": result}
			successCount++
		}
	}
	
	return map[string]interface{}{
		"total_items":    len(configs),
		"success_count":  successCount,
		"failure_count":  len(configs) - successCount,
		"results":        results,
	}, nil
}

// DELETE operations
func (a *API) deleteMultipleConfigs(ctx context.Context, connector connectors.DBConnector, tableName string, configs []ConfigItem) (interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	
	for _, config := range configs {
		result, err := a.deleteConfig(ctx, connector, tableName, config.Key)
		if err != nil {
			results[config.Key] = map[string]interface{}{"error": err.Error()}
		} else {
			results[config.Key] = map[string]interface{}{"success": true, "result": result}
			successCount++
		}
	}
	
	return map[string]interface{}{
		"total_items":    len(configs),
		"success_count":  successCount,
		"failure_count":  len(configs) - successCount,
		"results":        results,
	}, nil
}

func (a *API) deleteAllConfigs(ctx context.Context, connector connectors.DBConnector, tableName string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := "DELETE FROM " + tableName
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
		})
		
	case "mongodb":
		return connector.Execute(ctx, "delete", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) dropAllConfigTable(ctx context.Context, connector connectors.DBConnector, tableName string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := "DROP TABLE IF EXISTS " + tableName
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
		})
		
	case "mongodb":
		return connector.Execute(ctx, "drop", map[string]interface{}{
			"collection": tableName,
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// UTILITY operations
func (a *API) countConfigs(ctx context.Context, connector connectors.DBConnector, tableName string) (interface{}, error) {
	return a.getConfigCount(ctx, connector, tableName)
}

func (a *API) configExists(ctx context.Context, connector connectors.DBConnector, tableName, key string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := "SELECT COUNT(*) FROM " + tableName + " WHERE config_key = ?"
		rows, err := connector.Query(ctx, query, key)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"exists": count > 0,
				"key":    key,
			}, nil
		}
		return map[string]interface{}{"exists": false, "key": key}, nil
		
	case "postgresql":
		query := "SELECT COUNT(*) FROM " + tableName + " WHERE config_key = $1"
		rows, err := connector.Query(ctx, query, key)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"exists": count > 0,
				"key":    key,
			}, nil
		}
		return map[string]interface{}{"exists": false, "key": key}, nil
		
	case "mongodb":
		result, err := connector.Execute(ctx, "count", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"config_key": key},
		})
		if err != nil {
			return nil, err
		}
		
		var count int64
		if c, ok := result.(int64); ok {
			count = c
		} else if c, ok := result.(int); ok {
			count = int64(c)
		}
		
		return map[string]interface{}{
			"exists": count > 0,
			"key":    key,
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// ========================================
// MAKER-CHECKER WORKFLOW FUNCTIONS
// ========================================

// generateRequestID generates a unique request ID
func (a *API) generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// submitConfigForApproval submits a configuration change for approval
func (a *API) submitConfigForApproval(ctx context.Context, connector connectors.DBConnector, tableName, operation, key string, value interface{}, description, makerID string, previousValue interface{}) (interface{}, error) {
	requestID := a.generateRequestID()
	
	switch connector.GetType() {
	case "mysql":
		query := `INSERT INTO ` + tableName + `_approval_requests 
				  (request_id, config_key, config_value, description, operation, maker_id, status, requested_at, previous_value) 
				  VALUES (?, ?, ?, ?, ?, ?, 'pending', NOW(), ?)`
		
		valueStr := ""
		if value != nil {
			valueStr = fmt.Sprintf("%v", value)
		}
		prevValueStr := ""
		if previousValue != nil {
			prevValueStr = fmt.Sprintf("%v", previousValue)
		}
		
		result, err := connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{requestID, key, valueStr, description, operation, makerID, prevValueStr},
		})
		if err != nil {
			return nil, err
		}
		
		return map[string]interface{}{
			"request_id": requestID,
			"status":     "submitted_for_approval",
			"operation":  operation,
			"config_key": key,
			"maker_id":   makerID,
			"result":     result,
		}, nil
		
	case "postgresql":
		query := `INSERT INTO ` + tableName + `_approval_requests 
				  (request_id, config_key, config_value, description, operation, maker_id, status, requested_at, previous_value) 
				  VALUES ($1, $2, $3, $4, $5, $6, 'pending', CURRENT_TIMESTAMP, $7)`
		
		valueStr := ""
		if value != nil {
			valueStr = fmt.Sprintf("%v", value)
		}
		prevValueStr := ""
		if previousValue != nil {
			prevValueStr = fmt.Sprintf("%v", previousValue)
		}
		
		result, err := connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{requestID, key, valueStr, description, operation, makerID, prevValueStr},
		})
		if err != nil {
			return nil, err
		}
		
		return map[string]interface{}{
			"request_id": requestID,
			"status":     "submitted_for_approval",
			"operation":  operation,
			"config_key": key,
			"maker_id":   makerID,
			"result":     result,
		}, nil
		
	case "mongodb":
		doc := map[string]interface{}{
			"request_id":     requestID,
			"config_key":     key,
			"config_value":   value,
			"description":    description,
			"operation":      operation,
			"maker_id":       makerID,
			"status":         "pending",
			"requested_at":   time.Now(),
			"previous_value": previousValue,
		}
		
		result, err := connector.Execute(ctx, "insert", map[string]interface{}{
			"collection": tableName + "_approval_requests",
			"document":   doc,
		})
		if err != nil {
			return nil, err
		}
		
		return map[string]interface{}{
			"request_id": requestID,
			"status":     "submitted_for_approval",
			"operation":  operation,
			"config_key": key,
			"maker_id":   makerID,
			"result":     result,
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// approveRequest approves a pending configuration change
func (a *API) approveRequest(ctx context.Context, connector connectors.DBConnector, databaseName, tableName, requestID, checkerID, comment string) (interface{}, error) {
	// First, get the pending request details
	request, err := a.getPendingRequestByID(ctx, connector, tableName, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending request: %w", err)
	}
	
	if request == nil {
		return nil, fmt.Errorf("request not found or not in pending status")
	}
	
	// Apply the approved change to the main table
	var applyResult interface{}
	switch request["operation"].(string) {
	case "create":
		applyResult, err = a.createConfigDirect(ctx, connector, databaseName, tableName, 
			request["config_key"].(string), 
			request["config_value"], 
			request["description"].(string), 
			request["maker_id"].(string))
	case "update":
		applyResult, err = a.updateConfigDirect(ctx, connector, databaseName, tableName, 
			request["config_key"].(string), 
			request["config_value"], 
			request["description"].(string), 
			request["maker_id"].(string))
	case "delete":
		applyResult, err = a.deleteConfigDirect(ctx, connector, tableName, 
			request["config_key"].(string), 
			request["maker_id"].(string))
	default:
		return nil, fmt.Errorf("unsupported operation: %s", request["operation"])
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to apply approved change: %w", err)
	}
	
	// Update the approval request status
	err = a.updateApprovalRequestStatus(ctx, connector, tableName, requestID, "approved", checkerID, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to update approval request status: %w", err)
	}
	
	return map[string]interface{}{
		"request_id":       requestID,
		"status":           "approved",
		"checker_id":       checkerID,
		"approval_comment": comment,
		"applied_result":   applyResult,
	}, nil
}

// rejectRequest rejects a pending configuration change
func (a *API) rejectRequest(ctx context.Context, connector connectors.DBConnector, databaseName, tableName, requestID, checkerID, comment string) (interface{}, error) {
	// Update the approval request status to rejected
	err := a.updateApprovalRequestStatus(ctx, connector, tableName, requestID, "rejected", checkerID, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to update approval request status: %w", err)
	}
	
	return map[string]interface{}{
		"request_id":       requestID,
		"status":           "rejected",
		"checker_id":       checkerID,
		"approval_comment": comment,
	}, nil
}

// getPendingApprovals gets all pending approval requests
func (a *API) getPendingApprovals(ctx context.Context, connector connectors.DBConnector, tableName string, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := `SELECT request_id, config_key, config_value, description, operation, maker_id, 
				         requested_at, previous_value 
				  FROM ` + tableName + `_approval_requests 
				  WHERE status = 'pending' 
				  ORDER BY requested_at ASC`
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName + "_approval_requests",
			"filter":     map[string]interface{}{"status": "pending"},
			"sort":       map[string]interface{}{"requested_at": 1},
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// getMyRequests gets approval requests made by a specific maker
func (a *API) getMyRequests(ctx context.Context, connector connectors.DBConnector, tableName, makerID string, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `SELECT request_id, config_key, config_value, description, operation, status, 
				         requested_at, processed_at, checker_id, approval_comment, previous_value 
				  FROM ` + tableName + `_approval_requests 
				  WHERE maker_id = ? 
				  ORDER BY requested_at DESC`
		
		args := []interface{}{makerID}
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "postgresql":
		query := `SELECT request_id, config_key, config_value, description, operation, status, 
				         requested_at, processed_at, checker_id, approval_comment, previous_value 
				  FROM ` + tableName + `_approval_requests 
				  WHERE maker_id = $1 
				  ORDER BY requested_at DESC`
		
		args := []interface{}{makerID}
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName + "_approval_requests",
			"filter":     map[string]interface{}{"maker_id": makerID},
			"sort":       map[string]interface{}{"requested_at": -1},
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// getApprovalHistory gets the history of all processed approval requests
func (a *API) getApprovalHistory(ctx context.Context, connector connectors.DBConnector, tableName string, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := `SELECT request_id, config_key, config_value, description, operation, maker_id, 
				         checker_id, status, requested_at, processed_at, approval_comment, previous_value 
				  FROM ` + tableName + `_approval_requests 
				  WHERE status IN ('approved', 'rejected') 
				  ORDER BY processed_at DESC`
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName + "_approval_requests",
			"filter": map[string]interface{}{
				"status": map[string]interface{}{
					"$in": []string{"approved", "rejected"},
				},
			},
			"sort": map[string]interface{}{"processed_at": -1},
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// Helper functions for approval workflow

func (a *API) getPendingRequestByID(ctx context.Context, connector connectors.DBConnector, tableName, requestID string) (map[string]interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `SELECT request_id, config_key, config_value, description, operation, maker_id, previous_value 
				  FROM ` + tableName + `_approval_requests 
				  WHERE request_id = ? AND status = 'pending'`
		
		rows, err := connector.Query(ctx, query, requestID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		results, err := a.rowsToMap(rows)
		if err != nil {
			return nil, err
		}
		
		if results == nil || len(results) == 0 {
			return nil, nil
		}
		
		return results[0], nil
		
	case "postgresql":
		query := `SELECT request_id, config_key, config_value, description, operation, maker_id, previous_value 
				  FROM ` + tableName + `_approval_requests 
				  WHERE request_id = $1 AND status = 'pending'`
		
		rows, err := connector.Query(ctx, query, requestID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		results, err := a.rowsToMap(rows)
		if err != nil {
			return nil, err
		}
		
		if results == nil || len(results) == 0 {
			return nil, nil
		}
		
		return results[0], nil
		
	case "mongodb":
		result, err := connector.Execute(ctx, "findOne", map[string]interface{}{
			"collection": tableName + "_approval_requests",
			"filter": map[string]interface{}{
				"request_id": requestID,
				"status":     "pending",
			},
		})
		if err != nil {
			return nil, err
		}
		
		if result == nil {
			return nil, nil
		}
		
		return result.(map[string]interface{}), nil
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

func (a *API) updateApprovalRequestStatus(ctx context.Context, connector connectors.DBConnector, tableName, requestID, status, checkerID, comment string) error {
	switch connector.GetType() {
	case "mysql":
		query := `UPDATE ` + tableName + `_approval_requests 
				  SET status = ?, checker_id = ?, approval_comment = ?, processed_at = NOW() 
				  WHERE request_id = ?`
		
		_, err := connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{status, checkerID, comment, requestID},
		})
		return err
		
	case "postgresql":
		query := `UPDATE ` + tableName + `_approval_requests 
				  SET status = $1, checker_id = $2, approval_comment = $3, processed_at = CURRENT_TIMESTAMP 
				  WHERE request_id = $4`
		
		_, err := connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{status, checkerID, comment, requestID},
		})
		return err
		
	case "mongodb":
		_, err := connector.Execute(ctx, "update", map[string]interface{}{
			"collection": tableName + "_approval_requests",
			"filter":     map[string]interface{}{"request_id": requestID},
			"update": map[string]interface{}{
				"$set": map[string]interface{}{
					"status":           status,
					"checker_id":       checkerID,
					"approval_comment": comment,
					"processed_at":     time.Now(),
				},
			},
		})
		return err
		
	default:
		return fmt.Errorf("unsupported database type")
	}
}

// ========================================
// APPROVED-ONLY READ OPERATIONS
// ========================================

// readApprovedConfig reads a single approved configuration
func (a *API) readApprovedConfig(ctx context.Context, connector connectors.DBConnector, databaseName, tableName, key string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := "SELECT config_key, config_value, description, created_at, updated_at, maker_id, checker_id, approved_at FROM " + tableName + " WHERE config_key = ? AND status = 'approved'"
		rows, err := connector.Query(ctx, query, key)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "postgresql":
		query := "SELECT config_key, config_value, description, created_at, updated_at, maker_id, checker_id, approved_at FROM " + tableName + " WHERE config_key = $1 AND status = 'approved'"
		rows, err := connector.Query(ctx, query, key)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName,
			"filter": map[string]interface{}{
				"config_key": key,
				"status":     "approved",
			},
		}
		
		// Add database parameter for MongoDB
		if databaseName != "" {
			params["database"] = databaseName
		}
		
		return connector.Execute(ctx, "findOne", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// readAllApprovedConfigs reads all approved configurations
func (a *API) readAllApprovedConfigs(ctx context.Context, connector connectors.DBConnector, databaseName, tableName string, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := "SELECT config_key, config_value, description, created_at, updated_at, maker_id, checker_id, approved_at FROM " + tableName + " WHERE status = 'approved' ORDER BY config_key"
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"status": "approved"},
			"sort":       map[string]interface{}{"config_key": 1},
		}
		
		// Add database parameter for MongoDB
		if databaseName != "" {
			params["database"] = databaseName
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// searchApprovedConfigs searches approved configurations
func (a *API) searchApprovedConfigs(ctx context.Context, connector connectors.DBConnector, tableName, searchTerm string, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `SELECT config_key, config_value, description, created_at, updated_at, maker_id, checker_id, approved_at FROM ` + tableName + ` 
				  WHERE status = 'approved' AND (config_key LIKE ? OR config_value LIKE ? OR description LIKE ?) 
				  ORDER BY config_key`
		searchPattern := "%" + searchTerm + "%"
		args := []interface{}{searchPattern, searchPattern, searchPattern}
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "postgresql":
		query := `SELECT config_key, config_value, description, created_at, updated_at, maker_id, checker_id, approved_at FROM ` + tableName + ` 
				  WHERE status = 'approved' AND (config_key ILIKE $1 OR config_value ILIKE $2 OR description ILIKE $3) 
				  ORDER BY config_key`
		searchPattern := "%" + searchTerm + "%"
		args := []interface{}{searchPattern, searchPattern, searchPattern}
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName,
			"filter": map[string]interface{}{
				"status": "approved",
				"$or": []map[string]interface{}{
					{"config_key": map[string]interface{}{"$regex": searchTerm, "$options": "i"}},
					{"config_value": map[string]interface{}{"$regex": searchTerm, "$options": "i"}},
					{"description": map[string]interface{}{"$regex": searchTerm, "$options": "i"}},
				},
			},
			"sort": map[string]interface{}{"config_key": 1},
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// filterApprovedConfigs filters approved configurations
func (a *API) filterApprovedConfigs(ctx context.Context, connector connectors.DBConnector, tableName string, filter map[string]interface{}, limit, offset int) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		// Build WHERE clause from filter, ensuring status = 'approved'
		whereClause := "WHERE status = 'approved'"
		args := []interface{}{}
		paramIndex := 1
		
		for key, value := range filter {
			if connector.GetType() == "mysql" {
				whereClause += fmt.Sprintf(" AND %s = ?", key)
			} else {
				whereClause += fmt.Sprintf(" AND %s = $%d", key, paramIndex+1)
				paramIndex++
			}
			args = append(args, value)
		}
		
		query := fmt.Sprintf("SELECT config_key, config_value, description, created_at, updated_at, maker_id, checker_id, approved_at FROM %s %s ORDER BY config_key", tableName, whereClause)
		
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
			if offset > 0 {
				query += fmt.Sprintf(" OFFSET %d", offset)
			}
		}
		
		rows, err := connector.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return a.rowsToMap(rows)
		
	case "mongodb":
		// Add status filter to user's filter
		combinedFilter := map[string]interface{}{"status": "approved"}
		for k, v := range filter {
			combinedFilter[k] = v
		}
		
		params := map[string]interface{}{
			"collection": tableName,
			"filter":     combinedFilter,
			"sort":       map[string]interface{}{"config_key": 1},
		}
		
		if limit > 0 {
			params["limit"] = limit
		}
		if offset > 0 {
			params["skip"] = offset
		}
		
		return connector.Execute(ctx, "find", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// countApprovedConfigs counts only approved configurations
func (a *API) countApprovedConfigs(ctx context.Context, connector connectors.DBConnector, tableName string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		query := "SELECT COUNT(*) FROM " + tableName + " WHERE status = 'approved'"
		rows, err := connector.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		if rows.Next() {
			var count int64
			if err := rows.Scan(&count); err != nil {
				return nil, err
			}
			return count, nil
		}
		return 0, nil
		
	case "mongodb":
		return connector.Execute(ctx, "count", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"status": "approved"},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// configExistsApproved checks if an approved configuration exists
func (a *API) configExistsApproved(ctx context.Context, connector connectors.DBConnector, tableName, key string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := "SELECT COUNT(*) FROM " + tableName + " WHERE config_key = ? AND status = 'approved'"
		rows, err := connector.Query(ctx, query, key)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"exists": count > 0,
				"key":    key,
			}, nil
		}
		return map[string]interface{}{"exists": false, "key": key}, nil
		
	case "postgresql":
		query := "SELECT COUNT(*) FROM " + tableName + " WHERE config_key = $1 AND status = 'approved'"
		rows, err := connector.Query(ctx, query, key)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"exists": count > 0,
				"key":    key,
			}, nil
		}
		return map[string]interface{}{"exists": false, "key": key}, nil
		
	case "mongodb":
		result, err := connector.Execute(ctx, "count", map[string]interface{}{
			"collection": tableName,
			"filter": map[string]interface{}{
				"config_key": key,
				"status":     "approved",
			},
		})
		if err != nil {
			return nil, err
		}
		
		var count int64
		if c, ok := result.(int64); ok {
			count = c
		} else if c, ok := result.(int); ok {
			count = int64(c)
		}
		
		return map[string]interface{}{
			"exists": count > 0,
			"key":    key,
		}, nil
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// ========================================
// DIRECT OPERATIONS (BYPASS APPROVAL)
// ========================================

// createConfigDirect creates configuration directly with approved status
func (a *API) createConfigDirect(ctx context.Context, connector connectors.DBConnector, databaseName, tableName, key string, value interface{}, description, makerID string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `INSERT INTO ` + tableName + ` (config_key, config_value, description, status, maker_id, created_at, updated_at, approved_at) 
				  VALUES (?, ?, ?, 'approved', ?, NOW(), NOW(), NOW())`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key, value, description, makerID},
		})
		
	case "postgresql":
		query := `INSERT INTO ` + tableName + ` (config_key, config_value, description, status, maker_id, created_at, updated_at, approved_at) 
				  VALUES ($1, $2, $3, 'approved', $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key, value, description, makerID},
		})
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName,
			"document": map[string]interface{}{
				"config_key":   key,
				"config_value": value,
				"description":  description,
				"status":       "approved",
				"maker_id":     makerID,
				"created_at":   time.Now(),
				"updated_at":   time.Now(),
				"approved_at":  time.Now(),
			},
		}
		
		// Add database parameter for MongoDB
		if databaseName != "" {
			params["database"] = databaseName
		}
		
		return connector.Execute(ctx, "insert", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// updateConfigDirect updates configuration directly with approved status
func (a *API) updateConfigDirect(ctx context.Context, connector connectors.DBConnector, databaseName, tableName, key string, value interface{}, description, makerID string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := `UPDATE ` + tableName + ` SET config_value = ?, description = ?, status = 'approved', maker_id = ?, updated_at = NOW(), approved_at = NOW() WHERE config_key = ?`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{value, description, makerID, key},
		})
		
	case "postgresql":
		query := `UPDATE ` + tableName + ` SET config_value = $1, description = $2, status = 'approved', maker_id = $3, updated_at = CURRENT_TIMESTAMP, approved_at = CURRENT_TIMESTAMP WHERE config_key = $4`
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{value, description, makerID, key},
		})
		
	case "mongodb":
		params := map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"config_key": key},
			"update": map[string]interface{}{
				"$set": map[string]interface{}{
					"config_key":   key,
					"config_value": value,
					"description":  description,
					"status":       "approved",
					"maker_id":     makerID,
					"updated_at":   time.Now(),
					"approved_at":  time.Now(),
				},
				"$setOnInsert": map[string]interface{}{
					"created_at": time.Now(),
				},
			},
		}
		
		// Add database parameter for MongoDB
		if databaseName != "" {
			params["database"] = databaseName
		}
		
		return connector.Execute(ctx, "upsert", params)
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// deleteConfigDirect deletes configuration directly
func (a *API) deleteConfigDirect(ctx context.Context, connector connectors.DBConnector, tableName, key, makerID string) (interface{}, error) {
	switch connector.GetType() {
	case "mysql":
		query := "DELETE FROM " + tableName + " WHERE config_key = ?"
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key},
		})
		
	case "postgresql":
		query := "DELETE FROM " + tableName + " WHERE config_key = $1"
		return connector.Execute(ctx, "execute", map[string]interface{}{
			"query": query,
			"args":  []interface{}{key},
		})
		
	case "mongodb":
		return connector.Execute(ctx, "delete", map[string]interface{}{
			"collection": tableName,
			"filter":     map[string]interface{}{"config_key": key},
		})
		
	default:
		return nil, fmt.Errorf("unsupported database type")
	}
}

// createMultipleConfigsDirect creates multiple configurations directly with approved status
func (a *API) createMultipleConfigsDirect(ctx context.Context, connector connectors.DBConnector, databaseName, tableName string, configs []ConfigItem) (interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	
	for _, config := range configs {
		result, err := a.createConfigDirect(ctx, connector, databaseName, tableName, config.Key, config.Value, config.Description, config.MakerID)
		if err != nil {
			results[config.Key] = map[string]interface{}{"error": err.Error()}
		} else {
			results[config.Key] = map[string]interface{}{"success": true, "result": result}
			successCount++
		}
	}
	
	return map[string]interface{}{
		"total_items":    len(configs),
		"success_count":  successCount,
		"failure_count":  len(configs) - successCount,
		"results":        results,
	}, nil
}

// updateMultipleConfigsDirect updates multiple configurations directly with approved status
func (a *API) updateMultipleConfigsDirect(ctx context.Context, connector connectors.DBConnector, databaseName, tableName string, configs []ConfigItem) (interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	
	for _, config := range configs {
		result, err := a.updateConfigDirect(ctx, connector, databaseName, tableName, config.Key, config.Value, config.Description, config.MakerID)
		if err != nil {
			results[config.Key] = map[string]interface{}{"error": err.Error()}
		} else {
			results[config.Key] = map[string]interface{}{"success": true, "result": result}
			successCount++
		}
	}
	
	return map[string]interface{}{
		"total_items":    len(configs),
		"success_count":  successCount,
		"failure_count":  len(configs) - successCount,
		"results":        results,
	}, nil
}

// deleteMultipleConfigsDirect deletes multiple configurations directly
func (a *API) deleteMultipleConfigsDirect(ctx context.Context, connector connectors.DBConnector, tableName string, configs []ConfigItem) (interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	
	for _, config := range configs {
		result, err := a.deleteConfigDirect(ctx, connector, tableName, config.Key, config.MakerID)
		if err != nil {
			results[config.Key] = map[string]interface{}{"error": err.Error()}
		} else {
			results[config.Key] = map[string]interface{}{"success": true, "result": result}
			successCount++
		}
	}
	
	return map[string]interface{}{
		"total_items":    len(configs),
		"success_count":  successCount,
		"failure_count":  len(configs) - successCount,
		"results":        results,
	}, nil
}

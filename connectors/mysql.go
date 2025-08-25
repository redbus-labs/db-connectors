package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConnector implements DBConnector for MySQL
type MySQLConnector struct {
	config *ConnectionConfig
	db     *sql.DB
}

// NewMySQLConnector creates a new MySQL connector
func NewMySQLConnector(config *ConnectionConfig) *MySQLConnector {
	return &MySQLConnector{
		config: config,
	}
}

// Connect establishes a connection to MySQL
func (m *MySQLConnector) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.config.Username,
		m.config.Password,
		m.config.Host,
		m.config.Port,
		m.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	m.db = db
	return nil
}

// Ping tests the connection to MySQL
func (m *MySQLConnector) Ping(ctx context.Context) error {
	if m.db == nil {
		return fmt.Errorf("MySQL connection not established")
	}
	return m.db.PingContext(ctx)
}

// Close closes the MySQL connection
func (m *MySQLConnector) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// GetType returns the database type
func (m *MySQLConnector) GetType() string {
	return "mysql"
}

// Query executes a query and returns rows
func (m *MySQLConnector) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.db == nil {
		return nil, fmt.Errorf("MySQL connection not established")
	}
	return m.db.QueryContext(ctx, query, args...)
}

// Execute runs a command/query (for compatibility with interface)
func (m *MySQLConnector) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	if m.db == nil {
		return nil, fmt.Errorf("MySQL connection not established")
	}

	switch operation {
	case "insert", "update", "delete":
		if query, ok := params["query"].(string); ok {
			args := make([]interface{}, 0)
			if argsList, ok := params["args"].([]interface{}); ok {
				args = argsList
			}
			result, err := m.db.ExecContext(ctx, query, args...)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, fmt.Errorf("query parameter required for operation: %s", operation)
	case "select":
		if query, ok := params["query"].(string); ok {
			args := make([]interface{}, 0)
			if argsList, ok := params["args"].([]interface{}); ok {
				args = argsList
			}
			return m.Query(ctx, query, args...)
		}
		return nil, fmt.Errorf("query parameter required for operation: %s", operation)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// IsConnected returns whether the connection is active
func (m *MySQLConnector) IsConnected() bool {
	if m.db == nil {
		return false
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return m.Ping(ctx) == nil
}

package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLConnector implements DBConnector for PostgreSQL
type PostgreSQLConnector struct {
	config *ConnectionConfig
	db     *sql.DB
}

// NewPostgreSQLConnector creates a new PostgreSQL connector
func NewPostgreSQLConnector(config *ConnectionConfig) *PostgreSQLConnector {
	return &PostgreSQLConnector{
		config: config,
	}
}

// Connect establishes a connection to PostgreSQL
func (p *PostgreSQLConnector) Connect(ctx context.Context) error {
	sslMode := p.config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.Username,
		p.config.Password,
		p.config.Database,
		sslMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	p.db = db
	return nil
}

// Ping tests the connection to PostgreSQL
func (p *PostgreSQLConnector) Ping(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("PostgreSQL connection not established")
	}
	return p.db.PingContext(ctx)
}

// Close closes the PostgreSQL connection
func (p *PostgreSQLConnector) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// GetType returns the database type
func (p *PostgreSQLConnector) GetType() string {
	return "postgresql"
}

// Query executes a query and returns rows
func (p *PostgreSQLConnector) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if p.db == nil {
		return nil, fmt.Errorf("PostgreSQL connection not established")
	}
	return p.db.QueryContext(ctx, query, args...)
}

// Execute runs a command/query (for compatibility with interface)
func (p *PostgreSQLConnector) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	if p.db == nil {
		return nil, fmt.Errorf("PostgreSQL connection not established")
	}

	switch operation {
	case "insert", "update", "delete", "execute":
		if query, ok := params["query"].(string); ok {
			args := make([]interface{}, 0)
			if argsList, ok := params["args"].([]interface{}); ok {
				args = argsList
			}
			result, err := p.db.ExecContext(ctx, query, args...)
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
			return p.Query(ctx, query, args...)
		}
		return nil, fmt.Errorf("query parameter required for operation: %s", operation)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// IsConnected returns whether the connection is active
func (p *PostgreSQLConnector) IsConnected() bool {
	if p.db == nil {
		return false
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return p.Ping(ctx) == nil
}

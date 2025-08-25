package connectors

import (
	"context"
	"database/sql"
)

// DBConnector defines the interface that all database connectors must implement
type DBConnector interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error
	
	// Ping tests the connection to the database
	Ping(ctx context.Context) error
	
	// Close closes the database connection
	Close() error
	
	// GetType returns the type of database (mysql, postgres, mongodb)
	GetType() string
	
	// Query executes a query and returns rows (for SQL databases)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	
	// Execute runs a command/query (for MongoDB and other operations)
	Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)
	
	// IsConnected returns whether the connection is active
	IsConnected() bool
}

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode,omitempty"`
}

// DatabaseConfig represents configuration for all supported databases
type DatabaseConfig struct {
	MySQL      *ConnectionConfig `yaml:"mysql,omitempty"`
	PostgreSQL *ConnectionConfig `yaml:"postgresql,omitempty"`
	MongoDB    *ConnectionConfig `yaml:"mongodb,omitempty"`
}

// ConnectorRegistry manages all available database connectors
type ConnectorRegistry struct {
	connectors map[string]DBConnector
}

// NewConnectorRegistry creates a new connector registry
func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		connectors: make(map[string]DBConnector),
	}
}

// Register adds a connector to the registry
func (cr *ConnectorRegistry) Register(name string, connector DBConnector) {
	cr.connectors[name] = connector
}

// Get retrieves a connector by name
func (cr *ConnectorRegistry) Get(name string) (DBConnector, bool) {
	connector, exists := cr.connectors[name]
	return connector, exists
}

// List returns all registered connector names
func (cr *ConnectorRegistry) List() []string {
	names := make([]string, 0, len(cr.connectors))
	for name := range cr.connectors {
		names = append(names, name)
	}
	return names
}

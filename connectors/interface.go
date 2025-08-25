package connectors

import (
	"context"
	"database/sql"
	"fmt"
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

// Validate checks if the connection configuration is valid
func (c *ConnectionConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

// GetConnectionString generates a connection string for the specified database type
func (c *ConnectionConfig) GetConnectionString(dbType string) (string, error) {
	switch dbType {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Host, c.Port, c.Database), nil
	case "postgresql":
		sslMode := c.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", 
			c.Host, c.Port, c.Username, c.Password, c.Database, sslMode), nil
	case "mongodb":
		if c.Username != "" && c.Password != "" {
			return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.Database), nil
		}
		return fmt.Sprintf("mongodb://%s:%d/%s", c.Host, c.Port, c.Database), nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// DatabaseConfig represents configuration for all supported databases
type DatabaseConfig struct {
	MySQL      *ConnectionConfig `yaml:"mysql,omitempty"`
	PostgreSQL *ConnectionConfig `yaml:"postgresql,omitempty"`
	MongoDB    *ConnectionConfig `yaml:"mongodb,omitempty"`
}

// GetConfig returns the connection configuration for the specified database type
func (dc *DatabaseConfig) GetConfig(dbType string) (*ConnectionConfig, error) {
	switch dbType {
	case "mysql":
		if dc.MySQL == nil {
			return nil, fmt.Errorf("MySQL configuration not found")
		}
		return dc.MySQL, nil
	case "postgresql":
		if dc.PostgreSQL == nil {
			return nil, fmt.Errorf("PostgreSQL configuration not found")
		}
		return dc.PostgreSQL, nil
	case "mongodb":
		if dc.MongoDB == nil {
			return nil, fmt.Errorf("MongoDB configuration not found")
		}
		return dc.MongoDB, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
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

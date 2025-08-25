package connectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ConnectionConfig
		wantErr bool
	}{
		{
			name: "valid mysql config",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "valid postgresql config",
			config: ConnectionConfig{
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
			name: "valid mongodb config",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Username: "admin",
				Password: "password",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "mongodb without credentials",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: ConnectionConfig{
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     0,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "empty database",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnectionConfig_GetConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   ConnectionConfig
		dbType   string
		expected string
		wantErr  bool
	}{
		{
			name: "mysql connection string",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			dbType:   "mysql",
			expected: "root:password@tcp(localhost:3306)/testdb",
			wantErr:  false,
		},
		{
			name: "postgresql connection string",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "disable",
			},
			dbType:   "postgresql",
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable",
			wantErr:  false,
		},
		{
			name: "mongodb connection string with auth",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Username: "admin",
				Password: "password",
				Database: "testdb",
			},
			dbType:   "mongodb",
			expected: "mongodb://admin:password@localhost:27017/testdb",
			wantErr:  false,
		},
		{
			name: "mongodb connection string without auth",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Database: "testdb",
			},
			dbType:   "mongodb",
			expected: "mongodb://localhost:27017/testdb",
			wantErr:  false,
		},
		{
			name: "unsupported database type",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
			},
			dbType:  "oracle",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.GetConnectionString(tt.dbType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDatabaseConfig_GetConfig(t *testing.T) {
	tests := []struct {
		name     string
		dbConfig DatabaseConfig
		dbType   string
		expected *ConnectionConfig
		wantErr  bool
	}{
		{
			name: "get mysql config",
			dbConfig: DatabaseConfig{
				MySQL: &ConnectionConfig{
					Host:     "localhost",
					Port:     3306,
					Username: "root",
					Password: "password",
					Database: "testdb",
				},
			},
			dbType: "mysql",
			expected: &ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "get postgresql config",
			dbConfig: DatabaseConfig{
				PostgreSQL: &ConnectionConfig{
					Host:     "localhost",
					Port:     5432,
					Username: "postgres",
					Password: "password",
					Database: "testdb",
					SSLMode:  "disable",
				},
			},
			dbType: "postgresql",
			expected: &ConnectionConfig{
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
			name: "get mongodb config",
			dbConfig: DatabaseConfig{
				MongoDB: &ConnectionConfig{
					Host:     "localhost",
					Port:     27017,
					Username: "admin",
					Password: "password",
					Database: "testdb",
				},
			},
			dbType: "mongodb",
			expected: &ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Username: "admin",
				Password: "password",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name:     "unsupported database type",
			dbConfig: DatabaseConfig{},
			dbType:   "oracle",
			wantErr:  true,
		},
		{
			name:     "nil config for requested type",
			dbConfig: DatabaseConfig{},
			dbType:   "mysql",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.dbConfig.GetConfig(tt.dbType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

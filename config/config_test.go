package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite defines the test suite for configuration
type ConfigTestSuite struct {
	suite.Suite
	tempConfigFile string
}

// SetupSuite sets up the test suite
func (suite *ConfigTestSuite) SetupSuite() {
	// Create a temporary config file for testing
	suite.tempConfigFile = "test_config.yaml"
}

// TearDownSuite cleans up after the test suite
func (suite *ConfigTestSuite) TearDownSuite() {
	// Clean up temporary config file
	if _, err := os.Stat(suite.tempConfigFile); err == nil {
		os.Remove(suite.tempConfigFile)
	}
}

// SetupTest sets up each test
func (suite *ConfigTestSuite) SetupTest() {
	// Clear environment variables before each test
	os.Clearenv()
}

// TestLoadConfigFromFile tests loading configuration from YAML file
func (suite *ConfigTestSuite) TestLoadConfigFromFile() {
	// Create a test config file
	configContent := `
app_name: "test-app"
log_level: "debug"

databases:
  mysql:
    host: "mysql-host"
    port: 3306
    username: "mysql-user"
    password: "mysql-pass"
    database: "mysql-db"
    
  postgresql:
    host: "postgres-host"
    port: 5432
    username: "postgres-user"
    password: "postgres-pass"
    database: "postgres-db"
    ssl_mode: "require"
    
  mongodb:
    host: "mongo-host"
    port: 27017
    username: "mongo-user"
    password: "mongo-pass"
    database: "mongo-db"
`

	// Write test config to file
	err := os.WriteFile(suite.tempConfigFile, []byte(configContent), 0644)
	assert.NoError(suite.T(), err)

	// Load config from file
	config, err := LoadConfig(suite.tempConfigFile)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// Verify app settings
	assert.Equal(suite.T(), "test-app", config.AppName)
	assert.Equal(suite.T(), "debug", config.LogLevel)

	// Verify MySQL configuration
	assert.NotNil(suite.T(), config.Databases.MySQL)
	assert.Equal(suite.T(), "mysql-host", config.Databases.MySQL.Host)
	assert.Equal(suite.T(), 3306, config.Databases.MySQL.Port)
	assert.Equal(suite.T(), "mysql-user", config.Databases.MySQL.Username)
	assert.Equal(suite.T(), "mysql-pass", config.Databases.MySQL.Password)
	assert.Equal(suite.T(), "mysql-db", config.Databases.MySQL.Database)

	// Verify PostgreSQL configuration
	assert.NotNil(suite.T(), config.Databases.PostgreSQL)
	assert.Equal(suite.T(), "postgres-host", config.Databases.PostgreSQL.Host)
	assert.Equal(suite.T(), 5432, config.Databases.PostgreSQL.Port)
	assert.Equal(suite.T(), "postgres-user", config.Databases.PostgreSQL.Username)
	assert.Equal(suite.T(), "postgres-pass", config.Databases.PostgreSQL.Password)
	assert.Equal(suite.T(), "postgres-db", config.Databases.PostgreSQL.Database)
	assert.Equal(suite.T(), "require", config.Databases.PostgreSQL.SSLMode)

	// Verify MongoDB configuration
	assert.NotNil(suite.T(), config.Databases.MongoDB)
	assert.Equal(suite.T(), "mongo-host", config.Databases.MongoDB.Host)
	assert.Equal(suite.T(), 27017, config.Databases.MongoDB.Port)
	assert.Equal(suite.T(), "mongo-user", config.Databases.MongoDB.Username)
	assert.Equal(suite.T(), "mongo-pass", config.Databases.MongoDB.Password)
	assert.Equal(suite.T(), "mongo-db", config.Databases.MongoDB.Database)
}

// TestLoadConfigFromEnvironment tests loading configuration from environment variables
func (suite *ConfigTestSuite) TestLoadConfigFromEnvironment() {
	// Set environment variables
	os.Setenv("APP_NAME", "env-test-app")
	os.Setenv("LOG_LEVEL", "info")
	
	// MySQL environment variables
	os.Setenv("MYSQL_HOST", "env-mysql-host")
	os.Setenv("MYSQL_PORT", "3307")
	os.Setenv("MYSQL_USERNAME", "env-mysql-user")
	os.Setenv("MYSQL_PASSWORD", "env-mysql-pass")
	os.Setenv("MYSQL_DATABASE", "env-mysql-db")
	
	// PostgreSQL environment variables
	os.Setenv("POSTGRES_HOST", "env-postgres-host")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USERNAME", "env-postgres-user")
	os.Setenv("POSTGRES_PASSWORD", "env-postgres-pass")
	os.Setenv("POSTGRES_DATABASE", "env-postgres-db")
	os.Setenv("POSTGRES_SSLMODE", "disable")
	
	// MongoDB environment variables
	os.Setenv("MONGO_HOST", "env-mongo-host")
	os.Setenv("MONGO_PORT", "27018")
	os.Setenv("MONGO_USERNAME", "env-mongo-user")
	os.Setenv("MONGO_PASSWORD", "env-mongo-pass")
	os.Setenv("MONGO_DATABASE", "env-mongo-db")

	// Load config from environment (using non-existent file to force env loading)
	config, err := LoadConfig("non-existent-file.yaml")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// Verify app settings from environment
	assert.Equal(suite.T(), "env-test-app", config.AppName)
	assert.Equal(suite.T(), "info", config.LogLevel)

	// Verify MySQL configuration from environment
	assert.NotNil(suite.T(), config.Databases.MySQL)
	assert.Equal(suite.T(), "env-mysql-host", config.Databases.MySQL.Host)
	assert.Equal(suite.T(), 3307, config.Databases.MySQL.Port)
	assert.Equal(suite.T(), "env-mysql-user", config.Databases.MySQL.Username)
	assert.Equal(suite.T(), "env-mysql-pass", config.Databases.MySQL.Password)
	assert.Equal(suite.T(), "env-mysql-db", config.Databases.MySQL.Database)

	// Verify PostgreSQL configuration from environment
	assert.NotNil(suite.T(), config.Databases.PostgreSQL)
	assert.Equal(suite.T(), "env-postgres-host", config.Databases.PostgreSQL.Host)
	assert.Equal(suite.T(), 5433, config.Databases.PostgreSQL.Port)
	assert.Equal(suite.T(), "env-postgres-user", config.Databases.PostgreSQL.Username)
	assert.Equal(suite.T(), "env-postgres-pass", config.Databases.PostgreSQL.Password)
	assert.Equal(suite.T(), "env-postgres-db", config.Databases.PostgreSQL.Database)
	assert.Equal(suite.T(), "disable", config.Databases.PostgreSQL.SSLMode)

	// Verify MongoDB configuration from environment
	assert.NotNil(suite.T(), config.Databases.MongoDB)
	assert.Equal(suite.T(), "env-mongo-host", config.Databases.MongoDB.Host)
	assert.Equal(suite.T(), 27018, config.Databases.MongoDB.Port)
	assert.Equal(suite.T(), "env-mongo-user", config.Databases.MongoDB.Username)
	assert.Equal(suite.T(), "env-mongo-pass", config.Databases.MongoDB.Password)
	assert.Equal(suite.T(), "env-mongo-db", config.Databases.MongoDB.Database)
}

// TestConfigDefaults tests default configuration values
func (suite *ConfigTestSuite) TestConfigDefaults() {
	// Load config without file or environment variables
	config, err := LoadConfig("non-existent-file.yaml")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// Verify default values
	assert.Equal(suite.T(), "db-connectors", config.AppName)
	assert.Equal(suite.T(), "info", config.LogLevel)
}

// TestPartialConfiguration tests loading partial configuration
func (suite *ConfigTestSuite) TestPartialConfiguration() {
	// Create a config file with only MySQL configuration
	configContent := `
app_name: "partial-app"
log_level: "warn"

databases:
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "password"
    database: "testdb"
`

	// Write test config to file
	err := os.WriteFile(suite.tempConfigFile, []byte(configContent), 0644)
	assert.NoError(suite.T(), err)

	// Load config from file
	config, err := LoadConfig(suite.tempConfigFile)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// Verify app settings
	assert.Equal(suite.T(), "partial-app", config.AppName)
	assert.Equal(suite.T(), "warn", config.LogLevel)

	// Verify MySQL configuration exists
	assert.NotNil(suite.T(), config.Databases.MySQL)
	assert.Equal(suite.T(), "localhost", config.Databases.MySQL.Host)

	// Verify other databases are nil
	assert.Nil(suite.T(), config.Databases.PostgreSQL)
	assert.Nil(suite.T(), config.Databases.MongoDB)
}

// TestInvalidYAMLFile tests handling of invalid YAML files
func TestInvalidYAMLFile(t *testing.T) {
	// Create a file with invalid YAML content
	invalidContent := `
app_name: "test
invalid: yaml: content:
  - missing
    quotes
`
	
	tempFile := "invalid_config.yaml"
	err := os.WriteFile(tempFile, []byte(invalidContent), 0644)
	assert.NoError(t, err)
	defer os.Remove(tempFile)

	// Try to load invalid config
	config, err := LoadConfig(tempFile)
	assert.Error(t, err)
	assert.Nil(t, config)
}

// TestEnvironmentVariableOverrides tests that environment variables override file config
func (suite *ConfigTestSuite) TestEnvironmentVariableOverrides() {
	// Create a config file
	configContent := `
app_name: "file-app"
log_level: "debug"

databases:
  mysql:
    host: "file-mysql-host"
    port: 3306
    username: "file-user"
    password: "file-pass"
    database: "file-db"
`

	// Write test config to file
	err := os.WriteFile(suite.tempConfigFile, []byte(configContent), 0644)
	assert.NoError(suite.T(), err)

	// Set some environment variables to override file values
	os.Setenv("APP_NAME", "env-override-app")
	os.Setenv("MYSQL_HOST", "env-override-host")
	os.Setenv("MYSQL_PORT", "3307")

	// Load config (should use file as base, with env overrides)
	config, err := LoadConfig(suite.tempConfigFile)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// Verify environment variables override file values
	assert.Equal(suite.T(), "env-override-app", config.AppName)
	assert.Equal(suite.T(), "env-override-host", config.Databases.MySQL.Host)
	assert.Equal(suite.T(), 3307, config.Databases.MySQL.Port)

	// Verify file values are used where no env override exists
	assert.Equal(suite.T(), "debug", config.LogLevel)
	assert.Equal(suite.T(), "file-user", config.Databases.MySQL.Username)
	assert.Equal(suite.T(), "file-pass", config.Databases.MySQL.Password)
	assert.Equal(suite.T(), "file-db", config.Databases.MySQL.Database)
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				AppName:  "test-app",
				LogLevel: "info",
			},
			valid: true,
		},
		{
			name: "empty app name",
			config: Config{
				LogLevel: "info",
			},
			valid: false,
		},
		{
			name: "invalid log level",
			config: Config{
				AppName:  "test-app",
				LogLevel: "invalid",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestMongoDBOptionalAuth tests MongoDB configuration without authentication
func (suite *ConfigTestSuite) TestMongoDBOptionalAuth() {
	// Create a config file with MongoDB without auth
	configContent := `
databases:
  mongodb:
    host: "localhost"
    port: 27017
    database: "testdb"
`

	// Write test config to file
	err := os.WriteFile(suite.tempConfigFile, []byte(configContent), 0644)
	assert.NoError(suite.T(), err)

	// Load config from file
	config, err := LoadConfig(suite.tempConfigFile)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// Verify MongoDB configuration without auth
	assert.NotNil(suite.T(), config.Databases.MongoDB)
	assert.Equal(suite.T(), "localhost", config.Databases.MongoDB.Host)
	assert.Equal(suite.T(), 27017, config.Databases.MongoDB.Port)
	assert.Equal(suite.T(), "", config.Databases.MongoDB.Username)
	assert.Equal(suite.T(), "", config.Databases.MongoDB.Password)
	assert.Equal(suite.T(), "testdb", config.Databases.MongoDB.Database)
}

// Run the test suite
func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

// Benchmark tests
func BenchmarkLoadConfigFromFile(b *testing.B) {
	// Create a test config file
	configContent := `
app_name: "benchmark-app"
log_level: "info"

databases:
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "password"
    database: "testdb"
`
	
	tempFile := "benchmark_config.yaml"
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tempFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config, err := LoadConfig(tempFile)
		if err != nil {
			b.Fatal(err)
		}
		_ = config
	}
}

func BenchmarkLoadConfigFromEnvironment(b *testing.B) {
	// Set environment variables
	os.Setenv("APP_NAME", "benchmark-app")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("MYSQL_HOST", "localhost")
	os.Setenv("MYSQL_PORT", "3306")
	os.Setenv("MYSQL_USERNAME", "root")
	os.Setenv("MYSQL_PASSWORD", "password")
	os.Setenv("MYSQL_DATABASE", "testdb")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config, err := LoadConfig("non-existent-file.yaml")
		if err != nil {
			b.Fatal(err)
		}
		_ = config
	}
}

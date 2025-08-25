package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"db-connectors/connectors"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Databases connectors.DatabaseConfig `yaml:"databases"`
	LogLevel  string                    `yaml:"log_level,omitempty"`
	AppName   string                    `yaml:"app_name,omitempty"`
}

// LoadConfig loads configuration from a YAML file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	var config Config

	// Try to load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		// Read the config file
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// Parse YAML
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override/set values from environment variables
	loadFromEnvironment(&config)

	// Set defaults if not provided
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.AppName == "" {
		config.AppName = "db-connectors"
	}

	return &config, nil
}

// loadFromEnvironment loads configuration values from environment variables
func loadFromEnvironment(config *Config) {
	if appName := os.Getenv("APP_NAME"); appName != "" {
		config.AppName = appName
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// Load MySQL config from environment
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		if config.Databases.MySQL == nil {
			config.Databases.MySQL = &connectors.ConnectionConfig{}
		}
		config.Databases.MySQL.Host = host
	}
	if portStr := os.Getenv("MYSQL_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			if config.Databases.MySQL == nil {
				config.Databases.MySQL = &connectors.ConnectionConfig{}
			}
			config.Databases.MySQL.Port = port
		}
	}
	if username := os.Getenv("MYSQL_USERNAME"); username != "" {
		if config.Databases.MySQL == nil {
			config.Databases.MySQL = &connectors.ConnectionConfig{}
		}
		config.Databases.MySQL.Username = username
	}
	if password := os.Getenv("MYSQL_PASSWORD"); password != "" {
		if config.Databases.MySQL == nil {
			config.Databases.MySQL = &connectors.ConnectionConfig{}
		}
		config.Databases.MySQL.Password = password
	}
	if database := os.Getenv("MYSQL_DATABASE"); database != "" {
		if config.Databases.MySQL == nil {
			config.Databases.MySQL = &connectors.ConnectionConfig{}
		}
		config.Databases.MySQL.Database = database
	}

	// Load PostgreSQL config from environment
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		if config.Databases.PostgreSQL == nil {
			config.Databases.PostgreSQL = &connectors.ConnectionConfig{}
		}
		config.Databases.PostgreSQL.Host = host
	}
	if portStr := os.Getenv("POSTGRES_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			if config.Databases.PostgreSQL == nil {
				config.Databases.PostgreSQL = &connectors.ConnectionConfig{}
			}
			config.Databases.PostgreSQL.Port = port
		}
	}
	if username := os.Getenv("POSTGRES_USERNAME"); username != "" {
		if config.Databases.PostgreSQL == nil {
			config.Databases.PostgreSQL = &connectors.ConnectionConfig{}
		}
		config.Databases.PostgreSQL.Username = username
	}
	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		if config.Databases.PostgreSQL == nil {
			config.Databases.PostgreSQL = &connectors.ConnectionConfig{}
		}
		config.Databases.PostgreSQL.Password = password
	}
	if database := os.Getenv("POSTGRES_DATABASE"); database != "" {
		if config.Databases.PostgreSQL == nil {
			config.Databases.PostgreSQL = &connectors.ConnectionConfig{}
		}
		config.Databases.PostgreSQL.Database = database
	}
	if sslMode := os.Getenv("POSTGRES_SSLMODE"); sslMode != "" {
		if config.Databases.PostgreSQL == nil {
			config.Databases.PostgreSQL = &connectors.ConnectionConfig{}
		}
		config.Databases.PostgreSQL.SSLMode = sslMode
	}

	// Load MongoDB config from environment
	if host := os.Getenv("MONGO_HOST"); host != "" {
		if config.Databases.MongoDB == nil {
			config.Databases.MongoDB = &connectors.ConnectionConfig{}
		}
		config.Databases.MongoDB.Host = host
	}
	if portStr := os.Getenv("MONGO_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			if config.Databases.MongoDB == nil {
				config.Databases.MongoDB = &connectors.ConnectionConfig{}
			}
			config.Databases.MongoDB.Port = port
		}
	}
	if username := os.Getenv("MONGO_USERNAME"); username != "" {
		if config.Databases.MongoDB == nil {
			config.Databases.MongoDB = &connectors.ConnectionConfig{}
		}
		config.Databases.MongoDB.Username = username
	}
	if password := os.Getenv("MONGO_PASSWORD"); password != "" {
		if config.Databases.MongoDB == nil {
			config.Databases.MongoDB = &connectors.ConnectionConfig{}
		}
		config.Databases.MongoDB.Password = password
	}
	if database := os.Getenv("MONGO_DATABASE"); database != "" {
		if config.Databases.MongoDB == nil {
			config.Databases.MongoDB = &connectors.ConnectionConfig{}
		}
		config.Databases.MongoDB.Database = database
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.AppName == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	
	validLogLevels := []string{"debug", "info", "warn", "error"}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if c.LogLevel == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error", c.LogLevel)
	}
	
	return nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, configPath string) error {
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := &Config{
		LogLevel: getEnvWithDefault("LOG_LEVEL", "info"),
		AppName:  getEnvWithDefault("APP_NAME", "db-connectors"),
	}

	// MySQL configuration
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		config.Databases.MySQL = &connectors.ConnectionConfig{
			Host:     host,
			Port:     getEnvIntWithDefault("MYSQL_PORT", 3306),
			Username: os.Getenv("MYSQL_USERNAME"),
			Password: os.Getenv("MYSQL_PASSWORD"),
			Database: os.Getenv("MYSQL_DATABASE"),
		}
	}

	// PostgreSQL configuration
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		config.Databases.PostgreSQL = &connectors.ConnectionConfig{
			Host:     host,
			Port:     getEnvIntWithDefault("POSTGRES_PORT", 5432),
			Username: os.Getenv("POSTGRES_USERNAME"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Database: os.Getenv("POSTGRES_DATABASE"),
			SSLMode:  getEnvWithDefault("POSTGRES_SSLMODE", "disable"),
		}
	}

	// MongoDB configuration
	if host := os.Getenv("MONGO_HOST"); host != "" {
		config.Databases.MongoDB = &connectors.ConnectionConfig{
			Host:     host,
			Port:     getEnvIntWithDefault("MONGO_PORT", 27017),
			Username: os.Getenv("MONGO_USERNAME"),
			Password: os.Getenv("MONGO_PASSWORD"),
			Database: os.Getenv("MONGO_DATABASE"),
		}
	}

	return config
}

// GenerateExampleConfig creates an example configuration file
func GenerateExampleConfig(configPath string) error {
	config := &Config{
		LogLevel: "info",
		AppName:  "db-connectors",
		Databases: connectors.DatabaseConfig{
			MySQL: &connectors.ConnectionConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "password",
				Database: "testdb",
			},
			PostgreSQL: &connectors.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "testdb",
				SSLMode:  "disable",
			},
			MongoDB: &connectors.ConnectionConfig{
				Host:     "localhost",
				Port:     27017,
				Username: "admin",
				Password: "password",
				Database: "testdb",
			},
		},
	}

	return SaveConfig(config, configPath)
}

// Helper functions

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// Simple conversion, in production you'd want proper error handling
		if intValue := parseIntSafe(value); intValue != 0 {
			return intValue
		}
	}
	return defaultValue
}

func parseIntSafe(s string) int {
	// Basic integer parsing - in production use strconv.Atoi with error handling
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0
		}
	}
	return result
}

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"db-connectors/connectors"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Databases connectors.DatabaseConfig `yaml:"databases"`
	LogLevel  string                    `yaml:"log_level,omitempty"`
	AppName   string                    `yaml:"app_name,omitempty"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read the config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.AppName == "" {
		config.AppName = "db-connectors"
	}

	return &config, nil
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

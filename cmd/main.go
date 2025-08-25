package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"db-connectors/api"
	"db-connectors/config"
	"db-connectors/connectors"
)

func main() {
	// Parse command line flags
	var (
		port = flag.Int("port", 8080, "Port to run the API server on")
		mode = flag.String("mode", "api", "Mode to run: 'api' for HTTP server or 'demo' for CLI demo")
	)
	flag.Parse()

	fmt.Println("=== Database Connectors ===")

	switch *mode {
	case "api":
		runAPIServer(*port)
	case "demo":
		runCLIDemo()
	default:
		fmt.Printf("Unknown mode: %s. Use 'api' or 'demo'\n", *mode)
		os.Exit(1)
	}
}

func runAPIServer(port int) {
	fmt.Printf("🚀 Starting Database Connectors API Server on port %d\n", port)
	
	server := api.NewServer(port)
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func runCLIDemo() {
	fmt.Println("🔧 Running CLI Demo Mode...")
	fmt.Println("⚠️  CLI demo mode requires a config.yaml file or environment variables")
	
	// Import the original demo functions if needed
	fmt.Println("Use 'go run cmd/main.go -mode=api' to start the HTTP API server instead")
	os.Exit(0)
}

func loadConfiguration() (*config.Config, error) {
	// Try to load from config file first
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Warning: Could not load config file, trying environment variables: %v\n", err)
		
		// Fall back to environment variables
		cfg = config.LoadFromEnv()
		
		// Check if we have any database configurations
		if cfg.Databases.MySQL == nil && cfg.Databases.PostgreSQL == nil && cfg.Databases.MongoDB == nil {
			fmt.Println("No database configurations found. Generating example config file...")
			if err := config.GenerateExampleConfig("config.yaml"); err != nil {
				return nil, fmt.Errorf("failed to generate example config: %w", err)
			}
			fmt.Println("Example config file 'config.yaml' created. Please edit it with your database credentials.")
			os.Exit(0)
		}
	}

	return cfg, nil
}

func registerConnectors(registry *connectors.ConnectorRegistry, cfg *config.Config) {
	// Register MySQL connector if configured
	if cfg.Databases.MySQL != nil {
		mysqlConnector := connectors.NewMySQLConnector(cfg.Databases.MySQL)
		registry.Register("mysql", mysqlConnector)
		fmt.Println("✓ MySQL connector registered")
	}

	// Register PostgreSQL connector if configured
	if cfg.Databases.PostgreSQL != nil {
		postgresConnector := connectors.NewPostgreSQLConnector(cfg.Databases.PostgreSQL)
		registry.Register("postgresql", postgresConnector)
		fmt.Println("✓ PostgreSQL connector registered")
	}

	// Register MongoDB connector if configured
	if cfg.Databases.MongoDB != nil {
		mongoConnector := connectors.NewMongoDBConnector(cfg.Databases.MongoDB)
		registry.Register("mongodb", mongoConnector)
		fmt.Println("✓ MongoDB connector registered")
	}

	if len(registry.List()) == 0 {
		log.Fatal("No database connectors configured")
	}
}

func demonstrateConnectors(registry *connectors.ConnectorRegistry) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("\n=== Testing Database Connections ===")

	for _, name := range registry.List() {
		connector, _ := registry.Get(name)
		fmt.Printf("\nTesting %s (%s)...\n", name, connector.GetType())

		// Test connection
		if err := connector.Connect(ctx); err != nil {
			fmt.Printf("❌ Failed to connect to %s: %v\n", name, err)
			continue
		}
		defer connector.Close()

		// Test ping
		if err := connector.Ping(ctx); err != nil {
			fmt.Printf("❌ Failed to ping %s: %v\n", name, err)
			continue
		}

		fmt.Printf("✅ Successfully connected to %s\n", name)
		fmt.Printf("   Connection status: %v\n", connector.IsConnected())

		// Demonstrate operations based on database type
		demonstrateOperations(ctx, connector)
	}
}

func demonstrateOperations(ctx context.Context, connector connectors.DBConnector) {
	switch connector.GetType() {
	case "mysql", "postgresql":
		demonstrateSQLOperations(ctx, connector)
	case "mongodb":
		demonstrateMongoOperations(ctx, connector)
	}
}

func demonstrateSQLOperations(ctx context.Context, connector connectors.DBConnector) {
	fmt.Printf("   Demonstrating SQL operations for %s:\n", connector.GetType())

	// Try to create a test table (this might fail if table exists or no permissions)
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS test_users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	if connector.GetType() == "mysql" {
		createTableQuery = `
			CREATE TABLE IF NOT EXISTS test_users (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(100),
				email VARCHAR(100),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	}

	result, err := connector.Execute(ctx, "insert", map[string]interface{}{
		"query": createTableQuery,
	})
	if err != nil {
		fmt.Printf("      Note: Could not create test table: %v\n", err)
	} else {
		fmt.Printf("      ✓ Test table creation attempted: %v\n", result)
	}

	// Try a simple query
	rows, err := connector.Query(ctx, "SELECT 1 as test_column")
	if err != nil {
		fmt.Printf("      ❌ Failed to execute test query: %v\n", err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var testValue int
		if err := rows.Scan(&testValue); err == nil {
			fmt.Printf("      ✓ Test query successful, result: %d\n", testValue)
		}
	}
}

func demonstrateMongoOperations(ctx context.Context, connector connectors.DBConnector) {
	fmt.Printf("   Demonstrating MongoDB operations:\n")

	// Try to count documents in a test collection
	result, err := connector.Execute(ctx, "count", map[string]interface{}{
		"collection": "test_collection",
		"filter":     map[string]interface{}{},
	})
	if err != nil {
		fmt.Printf("      ❌ Failed to count documents: %v\n", err)
		return
	}

	fmt.Printf("      ✓ Document count in test_collection: %v\n", result)

	// Try to insert a test document
	testDoc := map[string]interface{}{
		"name":      "Test User",
		"email":     "test@example.com",
		"timestamp": time.Now(),
	}

	insertResult, err := connector.Execute(ctx, "insert", map[string]interface{}{
		"collection": "test_collection",
		"document":   testDoc,
	})
	if err != nil {
		fmt.Printf("      ❌ Failed to insert test document: %v\n", err)
		return
	}

	fmt.Printf("      ✓ Test document inserted: %v\n", insertResult)

	// Try to find the inserted document
	findResult, err := connector.Execute(ctx, "findOne", map[string]interface{}{
		"collection": "test_collection",
		"filter":     map[string]interface{}{"email": "test@example.com"},
	})
	if err != nil {
		fmt.Printf("      ❌ Failed to find test document: %v\n", err)
		return
	}

	fmt.Printf("      ✓ Found test document: %v\n", findResult != nil)

	// Clean up - delete the test document
	_, err = connector.Execute(ctx, "delete", map[string]interface{}{
		"collection": "test_collection",
		"filter":     map[string]interface{}{"email": "test@example.com"},
	})
	if err != nil {
		fmt.Printf("      ❌ Failed to clean up test document: %v\n", err)
	} else {
		fmt.Printf("      ✓ Test document cleaned up\n")
	}
}

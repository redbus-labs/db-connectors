# Database Connectors

A small Go application that provides a unified interface for connecting to multiple database types including MySQL, PostgreSQL, and MongoDB.

## Features

- **Multi-database support**: Connect to MySQL, PostgreSQL, and MongoDB
- **Unified interface**: All databases implement the same `DBConnector` interface
- **Configuration management**: Support for YAML configuration files and environment variables
- **Connection pooling**: Built-in connection pooling for optimal performance
- **Registry pattern**: Easy registration and management of database connectors

## Project Structure

```
db-connectors/
├── cmd/
│   └── main.go              # Main application entry point
├── connectors/
│   ├── interface.go         # Database connector interface and common types
│   ├── mysql.go            # MySQL connector implementation
│   ├── postgres.go         # PostgreSQL connector implementation
│   └── mongodb.go          # MongoDB connector implementation
├── config/
│   └── config.go           # Configuration management
├── config.yaml             # Example configuration file
├── go.mod                  # Go module dependencies
└── README.md               # This file
```

## Installation

1. Clone or download this project
2. Make sure you have Go 1.21 or later installed
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## Configuration

### Using Configuration File

Create a `config.yaml` file (an example is provided):

```yaml
app_name: "db-connectors"
log_level: "info"

databases:
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "password"
    database: "testdb"
    
  postgresql:
    host: "localhost"
    port: 5432
    username: "postgres"
    password: "password"
    database: "testdb"
    ssl_mode: "disable"
    
  mongodb:
    host: "localhost"
    port: 27017
    username: "admin"        # Optional - can be omitted for no-auth setups
    password: "password"     # Optional - can be omitted for no-auth setups
    database: "testdb"
```

### Using Environment Variables

You can also configure the application using environment variables:

```bash
# MySQL
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USERNAME=root
export MYSQL_PASSWORD=password
export MYSQL_DATABASE=testdb

# PostgreSQL
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USERNAME=postgres
export POSTGRES_PASSWORD=password
export POSTGRES_DATABASE=testdb
export POSTGRES_SSLMODE=disable

# MongoDB (username and password are optional)
export MONGO_HOST=localhost
export MONGO_PORT=27017
export MONGO_USERNAME=admin      # Optional - omit for no-auth setups
export MONGO_PASSWORD=password   # Optional - omit for no-auth setups
export MONGO_DATABASE=testdb

# App settings
export LOG_LEVEL=info
export APP_NAME=db-connectors
```

## Usage

### Running as HTTP API Server (Recommended)

```bash
# Start the API server on default port 8080
go run cmd/main.go

# Start on a custom port
go run cmd/main.go -port=3000

# Build and run
go build -o db-connectors cmd/main.go
./db-connectors -port=8080
```

The API provides endpoints to dynamically connect to databases without requiring configuration files:

- **GET** `/health` - Health check
- **POST** `/test-connection` - Test database connection with provided credentials
- **POST** `/execute` - Execute database operations

### Running as CLI Demo

```bash
# Run the original CLI demo mode
go run cmd/main.go -mode=demo
```

⚠️ **Note:** CLI demo mode requires a `config.yaml` file or environment variables to be set.

### API Documentation

The API includes comprehensive Swagger documentation:

- **Interactive Documentation**: Visit `http://localhost:8080/docs` for Swagger UI
- **OpenAPI Specification**: Available at `/swagger.json` and `/swagger.yaml`
- **Landing Page**: Visit `http://localhost:8080/` for documentation overview
- **Postman Collection**: Download from `docs/postman_collection.json`

See `examples/api_examples.md` for comprehensive API usage examples.

#### Quick API Test
```bash
# Test MySQL connection
curl -X POST http://localhost:8080/test-connection \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb"
  }'

# Execute a query
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost", 
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "query",
    "query": "SELECT 1 as test"
  }'
```

### Using the Connectors in Your Code

```go
package main

import (
    "context"
    "db-connectors/connectors"
    "db-connectors/config"
)

func example() {
    // Load configuration
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        panic(err)
    }

    // Create and register connectors
    registry := connectors.NewConnectorRegistry()
    
    if cfg.Databases.MySQL != nil {
        mysqlConnector := connectors.NewMySQLConnector(cfg.Databases.MySQL)
        registry.Register("mysql", mysqlConnector)
    }
    
    // Get a connector
    connector, exists := registry.Get("mysql")
    if !exists {
        panic("MySQL connector not found")
    }
    
    // Connect and use
    ctx := context.Background()
    if err := connector.Connect(ctx); err != nil {
        panic(err)
    }
    defer connector.Close()
    
    // For SQL databases
    rows, err := connector.Query(ctx, "SELECT * FROM users WHERE id = ?", 1)
    // Handle rows...
    
    // For all databases (using Execute method)
    result, err := connector.Execute(ctx, "select", map[string]interface{}{
        "query": "SELECT COUNT(*) FROM users",
    })
    // Handle result...
}
```

## Database-Specific Operations

### MySQL/PostgreSQL (SQL Databases)

```go
// Direct SQL query
rows, err := connector.Query(ctx, "SELECT * FROM users")

// Using Execute method
result, err := connector.Execute(ctx, "select", map[string]interface{}{
    "query": "SELECT * FROM users WHERE age > ?",
    "args": []interface{}{18},
})

// Insert/Update/Delete
result, err := connector.Execute(ctx, "insert", map[string]interface{}{
    "query": "INSERT INTO users (name, email) VALUES (?, ?)",
    "args": []interface{}{"John Doe", "john@example.com"},
})
```

### MongoDB (NoSQL Database)

```go
// Find documents
result, err := connector.Execute(ctx, "find", map[string]interface{}{
    "collection": "users",
    "filter": map[string]interface{}{"age": map[string]interface{}{"$gt": 18}},
})

// Insert document
result, err := connector.Execute(ctx, "insert", map[string]interface{}{
    "collection": "users",
    "document": map[string]interface{}{
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30,
    },
})

// Update documents
result, err := connector.Execute(ctx, "update", map[string]interface{}{
    "collection": "users",
    "filter": map[string]interface{}{"email": "john@example.com"},
    "update": map[string]interface{}{"$set": map[string]interface{}{"age": 31}},
})

// Count documents
count, err := connector.Execute(ctx, "count", map[string]interface{}{
    "collection": "users",
    "filter": map[string]interface{}{},
})
```

## Requirements

- Go 1.21 or later
- Access to at least one of the supported databases:
  - MySQL 5.7+ or 8.0+
  - PostgreSQL 10+
  - MongoDB 4.0+

## Dependencies

- `github.com/go-sql-driver/mysql` - MySQL driver
- `github.com/lib/pq` - PostgreSQL driver
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `gopkg.in/yaml.v3` - YAML configuration parsing

## Error Handling

The application includes comprehensive error handling:

- Connection failures are reported with detailed error messages
- Each database operation returns appropriate errors
- Connection pooling and timeout handling
- Graceful connection cleanup

## Contributing

Feel free to extend this application by:

1. Adding support for more databases (Redis, SQLite, etc.)
2. Implementing more sophisticated query builders
3. Adding connection caching and advanced pooling
4. Adding metrics and monitoring capabilities

## License

This is a sample application for educational purposes.

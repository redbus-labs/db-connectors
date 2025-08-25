# Database Connectors API Examples

This document provides examples of how to use the Database Connectors HTTP API.

## Base URL
```
http://localhost:8080
```

## Endpoints

### 1. Health Check
**GET** `/health`

Check if the API service is running.

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "healthy",
    "service": "db-connectors-api",
    "version": "1.0.0"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 2. Test Database Connection
**POST** `/test-connection`

Test if you can connect to a database with provided credentials.

#### MySQL Example
```bash
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
```

#### PostgreSQL Example
```bash
curl -X POST http://localhost:8080/test-connection \
  -H "Content-Type: application/json" \
  -d '{
    "type": "postgresql",
    "host": "localhost",
    "port": 5432,
    "username": "postgres",
    "password": "password",
    "database": "testdb",
    "ssl_mode": "disable"
  }'
```

#### MongoDB Example (with authentication)
```bash
curl -X POST http://localhost:8080/test-connection \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "username": "admin",
    "password": "password",
    "database": "testdb"
  }'
```

#### MongoDB Example (without authentication)
```bash
curl -X POST http://localhost:8080/test-connection \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "database": "testdb"
  }'
```

**Note:** For MongoDB, `username` and `password` are optional. You can connect to MongoDB instances that don't require authentication by omitting these fields.

**Success Response:**
```json
{
  "success": true,
  "message": "Database connection successful",
  "data": {
    "connection_status": "success",
    "database_type": "mysql",
    "connected": true
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 3. AllConfig Table Management
**POST** `/allconfig`

Check if an "allconfig" table/collection exists and get information about it.

#### Check for Default "allconfig" Table
```bash
curl -X POST http://localhost:8080/allconfig \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb"
  }'
```

#### Check for Custom Table Name
```bash
curl -X POST http://localhost:8080/allconfig \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "table_name": "my_config_table"
  }'
```

**Response (Table Exists):**
```json
{
  "success": true,
  "message": "AllConfig table check completed",
  "data": {
    "table_name": "allconfig",
    "table_exists": true,
    "database_type": "mysql",
    "table_structure": [
      {
        "Field": "id",
        "Type": "int",
        "Null": "NO",
        "Key": "PRI",
        "Default": null,
        "Extra": "auto_increment"
      },
      {
        "Field": "config_key",
        "Type": "varchar(255)",
        "Null": "NO",
        "Key": "UNI",
        "Default": null,
        "Extra": ""
      }
    ],
    "config_count": 5
  }
}
```

**Response (Table Doesn't Exist):**
```json
{
  "success": true,
  "message": "AllConfig table check completed",
  "data": {
    "table_name": "allconfig",
    "table_exists": false,
    "database_type": "mysql",
    "create_table_sql": "CREATE TABLE allconfig (\n    id INT AUTO_INCREMENT PRIMARY KEY,\n    config_key VARCHAR(255) NOT NULL UNIQUE,\n    config_value TEXT,\n    description TEXT,\n    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,\n    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\n    INDEX idx_config_key (config_key)\n)"
  }
}
```

### 4. AllConfig Operations
**POST** `/allconfig-operation`

Perform operations on the allconfig table/collection.

#### Create AllConfig Table
```bash
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "create_table"
  }'
```

#### Get All Configurations
```bash
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "get_all"
  }'
```

#### Get Specific Configuration
```bash
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "get_config",
    "key": "api_url"
  }'
```

#### Set Configuration
```bash
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "set_config",
    "key": "api_url",
    "value": "https://api.example.com"
  }'
```

#### Set Multiple Configurations
```bash
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "set_multiple",
    "configs": {
      "api_url": "https://api.example.com",
      "timeout": "30",
      "retry_count": "3",
      "debug_mode": "false"
    }
  }'
```

#### Delete Configuration
```bash
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "delete_config",
    "key": "old_setting"
  }'
```

### 5. Execute Database Operations
**POST** `/execute`

Execute operations on the database.

#### SQL Database Operations

##### Select Query (MySQL/PostgreSQL)
```bash
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
    "query": "SELECT * FROM users WHERE age > ?",
    "args": [18]
  }'
```

##### Insert (MySQL/PostgreSQL)
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "insert",
    "query": "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
    "args": ["John Doe", "john@example.com", 25]
  }'
```

##### Update (MySQL/PostgreSQL)
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "postgresql",
    "host": "localhost",
    "port": 5432,
    "username": "postgres",
    "password": "password",
    "database": "testdb",
    "ssl_mode": "disable",
    "operation": "update",
    "query": "UPDATE users SET age = ? WHERE email = ?",
    "args": [26, "john@example.com"]
  }'
```

#### MongoDB Operations

##### Find Documents (with authentication)
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "username": "admin",
    "password": "password",
    "database": "testdb",
    "operation": "find",
    "params": {
      "collection": "users",
      "filter": {"age": {"$gt": 18}},
      "limit": 10
    }
  }'
```

##### Find Documents (without authentication)
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "database": "testdb",
    "operation": "find",
    "params": {
      "collection": "users",
      "filter": {"age": {"$gt": 18}},
      "limit": 10
    }
  }'
```

##### List Collections (without authentication)
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "database": "testdb",
    "operation": "listCollections",
    "params": {
      "filter": {}
    }
  }'
```

**Note:** `listCollections` is a database-level operation and doesn't require a `collection` parameter.

##### Insert Document
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "username": "admin",
    "password": "password",
    "database": "testdb",
    "operation": "insert",
    "params": {
      "collection": "users",
      "document": {
        "name": "Jane Doe",
        "email": "jane@example.com",
        "age": 30
      }
    }
  }'
```

##### Update Document
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "username": "admin",
    "password": "password",
    "database": "testdb",
    "operation": "update",
    "params": {
      "collection": "users",
      "filter": {"email": "jane@example.com"},
      "update": {"$set": {"age": 31}}
    }
  }'
```

##### Count Documents
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "username": "admin",
    "password": "password",
    "database": "testdb",
    "operation": "count",
    "params": {
      "collection": "users",
      "filter": {}
    }
  }'
```

## Success Response Format
```json
{
  "success": true,
  "message": "Operation executed successfully",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "age": 25
    }
  ],
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Error Response Format
```json
{
  "success": false,
  "error": "Connection failed: authentication failed",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Supported Database Types

- **mysql**: MySQL 5.7+ or 8.0+
- **postgresql**: PostgreSQL 10+
- **mongodb**: MongoDB 4.0+

## Supported Operations

### SQL Databases (MySQL/PostgreSQL)
- `query` or `select`: Execute SELECT queries
- `insert`: Execute INSERT statements
- `update`: Execute UPDATE statements
- `delete`: Execute DELETE statements
- `execute`: Execute any SQL statement

### MongoDB
- `find`: Find documents
- `findOne`: Find a single document
- `insert`: Insert a document
- `update`: Update documents
- `delete`: Delete documents
- `count`: Count documents

### AllConfig Operations
- `create_table`: Create the allconfig table/collection
- `get_all`: Get all configurations
- `get_config`: Get a specific configuration by key
- `set_config`: Set/update a configuration value
- `set_multiple`: Set multiple configurations at once
- `delete_config`: Delete a configuration by key

## Error Codes

- `400`: Bad Request - Invalid JSON or missing required fields
- `405`: Method Not Allowed - Wrong HTTP method
- `500`: Internal Server Error - Database connection or operation failed

## AllConfig Workflow Example

Here's a complete workflow for working with the allconfig table:

```bash
# 1. Check if allconfig table exists
curl -X POST http://localhost:8080/allconfig \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb"
  }'

# 2. If table doesn't exist, create it
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "create_table"
  }'

# 3. Set initial configurations
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "set_multiple",
    "configs": {
      "app_name": "My Application",
      "version": "1.0.0",
      "debug_enabled": "true",
      "max_connections": "100"
    }
  }'

# 4. Get all configurations
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "get_all"
  }'
```

## Running the API Server

```bash
# Start the API server on default port 8080
go run cmd/main.go

# Start on a custom port
go run cmd/main.go -port=3000

# Run in demo mode (CLI)
go run cmd/main.go -mode=demo
```

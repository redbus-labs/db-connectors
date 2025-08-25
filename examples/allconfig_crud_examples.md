# AllConfig CRUD Operations Examples

This document provides comprehensive examples of all CRUD (Create, Read, Update, Delete) operations available for the AllConfig table/collection.

## Table of Contents
- [Table Management](#table-management)
- [CREATE Operations](#create-operations)
- [READ Operations](#read-operations)
- [UPDATE Operations](#update-operations)
- [DELETE Operations](#delete-operations)
- [UTILITY Operations](#utility-operations)
- [Complete Workflow Examples](#complete-workflow-examples)

---

## Table Management

### Check if AllConfig Table Exists
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

### Create AllConfig Table
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

### Drop AllConfig Table
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
    "operation": "drop_table"
  }'
```

---

## CREATE Operations

### Create Single Configuration
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
    "operation": "create",
    "key": "api_base_url",
    "value": "https://api.myapp.com/v1",
    "description": "Base URL for API endpoints"
  }'
```

### Create Multiple Configurations (Batch)
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
    "operation": "create_batch",
    "config_items": [
      {
        "key": "app_name",
        "value": "My Application",
        "description": "Application name"
      },
      {
        "key": "version",
        "value": "1.2.3",
        "description": "Application version"
      },
      {
        "key": "debug_mode",
        "value": "true",
        "description": "Enable debug logging"
      },
      {
        "key": "max_connections",
        "value": "100",
        "description": "Maximum database connections"
      }
    ]
  }'
```

**Response for batch create:**
```json
{
  "success": true,
  "message": "AllConfig operation 'create_batch' completed",
  "data": {
    "total_items": 4,
    "success_count": 4,
    "failure_count": 0,
    "results": {
      "app_name": {"success": true, "result": "..."},
      "version": {"success": true, "result": "..."},
      "debug_mode": {"success": true, "result": "..."},
      "max_connections": {"success": true, "result": "..."}
    }
  }
}
```

---

## READ Operations

### Read Single Configuration
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
    "operation": "read",
    "key": "api_base_url"
  }'
```

### Read All Configurations
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
    "operation": "read_all"
  }'
```

### Read All with Pagination
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
    "operation": "read_all",
    "limit": 10,
    "offset": 0
  }'
```

### Search Configurations
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
    "operation": "search",
    "search_term": "api",
    "limit": 5
  }'
```

### Filter Configurations
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
    "operation": "filter",
    "filter": {
      "config_value": "true"
    },
    "limit": 10
  }'
```

---

## UPDATE Operations

### Update Single Configuration
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
    "operation": "update",
    "key": "api_base_url",
    "value": "https://api.myapp.com/v2",
    "description": "Updated API base URL to v2"
  }'
```

### Update Multiple Configurations (Batch)
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
    "operation": "update_batch",
    "config_items": [
      {
        "key": "debug_mode",
        "value": "false",
        "description": "Disable debug logging for production"
      },
      {
        "key": "max_connections",
        "value": "200",
        "description": "Increased max connections for high load"
      },
      {
        "key": "timeout_seconds",
        "value": "60",
        "description": "Request timeout in seconds"
      }
    ]
  }'
```

---

## DELETE Operations

### Delete Single Configuration
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
    "operation": "delete",
    "key": "old_setting"
  }'
```

### Delete Multiple Configurations (Batch)
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
    "operation": "delete_batch",
    "config_items": [
      {"key": "deprecated_setting1"},
      {"key": "deprecated_setting2"},
      {"key": "temp_config"}
    ]
  }'
```

### Delete All Configurations
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
    "operation": "delete_all"
  }'
```

---

## UTILITY Operations

### Count Configurations
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
    "operation": "count"
  }'
```

### Check if Configuration Exists
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
    "operation": "exists",
    "key": "api_base_url"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "AllConfig operation 'exists' completed",
  "data": {
    "exists": true,
    "key": "api_base_url"
  }
}
```

---

## Complete Workflow Examples

### 1. Initial Setup Workflow
```bash
#!/bin/bash

# Set connection details
HOST="localhost"
PORT=3306
USERNAME="root"
PASSWORD="password"
DATABASE="testdb"
TYPE="mysql"

# 1. Check if table exists
echo "1. Checking if allconfig table exists..."
curl -X POST http://localhost:8080/allconfig \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"$TYPE\",
    \"host\": \"$HOST\",
    \"port\": $PORT,
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\",
    \"database\": \"$DATABASE\"
  }"

echo -e "\n\n2. Creating allconfig table..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"$TYPE\",
    \"host\": \"$HOST\",
    \"port\": $PORT,
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\",
    \"database\": \"$DATABASE\",
    \"operation\": \"create_table\"
  }"

echo -e "\n\n3. Setting up initial configurations..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"$TYPE\",
    \"host\": \"$HOST\",
    \"port\": $PORT,
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\",
    \"database\": \"$DATABASE\",
    \"operation\": \"create_batch\",
    \"config_items\": [
      {\"key\": \"app_name\", \"value\": \"My Application\", \"description\": \"Application name\"},
      {\"key\": \"version\", \"value\": \"1.0.0\", \"description\": \"Application version\"},
      {\"key\": \"debug_mode\", \"value\": \"true\", \"description\": \"Enable debug logging\"},
      {\"key\": \"api_base_url\", \"value\": \"https://api.example.com\", \"description\": \"API base URL\"},
      {\"key\": \"max_connections\", \"value\": \"100\", \"description\": \"Max database connections\"}
    ]
  }"

echo -e "\n\n4. Verifying configurations..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"$TYPE\",
    \"host\": \"$HOST\",
    \"port\": $PORT,
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\",
    \"database\": \"$DATABASE\",
    \"operation\": \"read_all\"
  }"
```

### 2. Configuration Management Workflow
```bash
#!/bin/bash

# Update configurations for production deployment
echo "1. Updating configurations for production..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "update_batch",
    "config_items": [
      {"key": "debug_mode", "value": "false", "description": "Disable debug for production"},
      {"key": "max_connections", "value": "500", "description": "Increased for production load"},
      {"key": "api_base_url", "value": "https://prod-api.example.com", "description": "Production API URL"}
    ]
  }'

echo -e "\n\n2. Adding new production-specific configs..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "create_batch",
    "config_items": [
      {"key": "cache_ttl", "value": "3600", "description": "Cache TTL in seconds"},
      {"key": "rate_limit", "value": "1000", "description": "API rate limit per hour"},
      {"key": "backup_enabled", "value": "true", "description": "Enable automated backups"}
    ]
  }'

echo -e "\n\n3. Removing development configs..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "delete_batch",
    "config_items": [
      {"key": "dev_mode"},
      {"key": "test_api_key"},
      {"key": "mock_data_enabled"}
    ]
  }'

echo -e "\n\n4. Final configuration count..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "count"
  }'
```

### 3. Search and Filter Example
```bash
#!/bin/bash

# Search for API-related configurations
echo "1. Searching for API-related configs..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "search",
    "search_term": "api",
    "limit": 10
  }'

echo -e "\n\n2. Filter configurations with boolean values..."
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "filter",
    "filter": {
      "config_value": "true"
    }
  }'
```

## Supported Database Types

All examples work with:
- **MySQL**: Replace `"type": "mysql"` and use appropriate connection details
- **PostgreSQL**: Replace `"type": "postgresql"` and add `"ssl_mode": "disable"`
- **MongoDB**: Replace `"type": "mongodb"` and use MongoDB connection details

### PostgreSQL Example
```bash
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "postgresql",
    "host": "localhost",
    "port": 5432,
    "username": "postgres",
    "password": "password",
    "database": "testdb",
    "ssl_mode": "disable",
    "operation": "create",
    "key": "pg_config",
    "value": "postgres_value",
    "description": "PostgreSQL specific config"
  }'
```

### MongoDB Example
```bash
curl -X POST http://localhost:8080/allconfig-operation \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "host": "localhost",
    "port": 27017,
    "username": "admin",
    "password": "password",
    "database": "testdb",
    "operation": "create",
    "key": "mongo_config",
    "value": "mongo_value",
    "description": "MongoDB specific config"
  }'
```

## Operation Summary

| Operation | Description | Required Fields |
|-----------|-------------|----------------|
| `create_table` | Create allconfig table | - |
| `drop_table` | Drop allconfig table | - |
| `create` | Create single config | `key`, `value` |
| `create_batch` | Create multiple configs | `config_items` |
| `read` | Read single config | `key` |
| `read_all` | Read all configs | - |
| `search` | Search configs | `search_term` |
| `filter` | Filter configs | `filter` |
| `update` | Update single config | `key`, `value` |
| `update_batch` | Update multiple configs | `config_items` |
| `delete` | Delete single config | `key` |
| `delete_batch` | Delete multiple configs | `config_items` |
| `delete_all` | Delete all configs | - |
| `count` | Count total configs | - |
| `exists` | Check if config exists | `key` |

**Note:** All operations require database connection parameters (`type`, `host`, `port`, `username`, `password`, `database`).

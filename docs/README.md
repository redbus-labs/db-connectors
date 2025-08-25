# Database Connectors API Documentation

This directory contains the complete API documentation for the Database Connectors service.

## Files

- **`swagger.yaml`** - Complete OpenAPI 3.0 specification in YAML format
- **`swagger.json`** - Complete OpenAPI 3.0 specification in JSON format
- **`README.md`** - This documentation file

## Viewing the Documentation

### Option 1: Swagger UI (Recommended)

You can view the interactive documentation using Swagger UI:

1. **Online Swagger Editor:**
   - Go to [editor.swagger.io](https://editor.swagger.io)
   - Copy and paste the contents of `swagger.yaml` into the editor
   - View the interactive documentation on the right panel

2. **Local Swagger UI:**
   ```bash
   # Using Docker
   docker run -p 8081:8080 -e SWAGGER_JSON=/docs/swagger.json -v $(pwd)/docs:/docs swaggerapi/swagger-ui
   
   # Then open http://localhost:8081 in your browser
   ```

3. **VS Code Extension:**
   - Install the "Swagger Viewer" extension
   - Open `swagger.yaml` and use `Shift+Alt+P` then "Preview Swagger"

### Option 2: Redoc

For a different documentation style:

```bash
# Using Docker
docker run -p 8082:80 -e SPEC_URL=/docs/swagger.yaml -v $(pwd)/docs:/usr/share/nginx/html/docs redocly/redoc

# Then open http://localhost:8082 in your browser
```

## API Endpoints Overview

### Core Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/test-connection` | POST | Test database connection |
| `/execute` | POST | Execute database operations |
| `/allconfig` | POST | Check AllConfig table |
| `/allconfig-operation` | POST | Perform AllConfig operations |

### Maker-Checker Operations

The API implements a comprehensive maker-checker approval workflow:

#### Maker Operations
- `submit_create` - Submit new configuration for approval
- `submit_update` - Submit configuration update for approval
- `submit_delete` - Submit configuration deletion for approval
- `get_my_requests` - Get maker's request history

#### Checker Operations
- `get_pending_approvals` - Get all pending approval requests
- `approve_request` - Approve a pending request
- `reject_request` - Reject a pending request
- `get_approval_history` - Get approval history

#### Read Operations (Approved Only)
- `read` - Read single approved configuration
- `read_all` - Read all approved configurations
- `search` - Search approved configurations
- `filter` - Filter approved configurations

#### Admin Operations
- `read_all_admin` - Read all configurations (including pending)
- `direct_create` - Create configuration directly (bypass approval)
- `direct_update` - Update configuration directly (bypass approval)
- `direct_delete` - Delete configuration directly (bypass approval)

## Authentication

The API supports multiple authentication methods:

- **API Key**: Send `X-API-Key` header
- **Bearer Token**: Send `Authorization: Bearer <token>` header

## Database Support

The API supports the following database types:

- **MySQL** (5.7+ or 8.0+)
- **PostgreSQL** (10+)
- **MongoDB** (4.0+)

## Error Handling

All endpoints return consistent error responses:

```json
{
  "success": false,
  "error": "Error message",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

Common HTTP status codes:
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `405` - Method Not Allowed
- `500` - Internal Server Error

## Example Usage

### Test MySQL Connection
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

### Submit Configuration for Approval
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
    "operation": "submit_create",
    "key": "api_url",
    "value": "https://api.example.com",
    "description": "API endpoint URL",
    "maker_id": "developer001"
  }'
```

### Approve Configuration Request
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
    "operation": "approve_request",
    "request_id": "f47ac10b58cc4372a5670e02b2c3d479",
    "checker_id": "admin001",
    "approval_comment": "Approved for production"
  }'
```

## Code Examples

### JavaScript/Node.js
```javascript
const response = await fetch('http://localhost:8080/allconfig-operation', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    type: 'mysql',
    host: 'localhost',
    port: 3306,
    username: 'root',
    password: 'password',
    database: 'testdb',
    operation: 'read_all',
    limit: 10
  })
});

const data = await response.json();
console.log(data);
```

### Python
```python
import requests

payload = {
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "password",
    "database": "testdb",
    "operation": "read_all",
    "limit": 10
}

response = requests.post(
    'http://localhost:8080/allconfig-operation',
    json=payload
)

print(response.json())
```

## Rate Limiting

The API implements rate limiting to prevent abuse:
- **Default limit**: 100 requests per minute per IP
- **Burst limit**: 20 requests in 10 seconds

Rate limit headers are included in responses:
- `X-RateLimit-Limit` - Total requests allowed
- `X-RateLimit-Remaining` - Requests remaining
- `X-RateLimit-Reset` - Time when limit resets

## Monitoring and Metrics

The API provides monitoring endpoints:

- `/health` - Basic health check
- `/metrics` - Prometheus-compatible metrics (if enabled)
- `/version` - API version information

## Support

For support and questions:
- Email: support@example.com
- Documentation: This README and the Swagger documentation
- Examples: See the `examples/` directory in the project root

## License

This API documentation is provided under the MIT License.

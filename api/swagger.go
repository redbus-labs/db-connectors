package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

// SwaggerHandler serves the Swagger documentation
func (s *Server) SwaggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Serve Swagger UI HTML
	swaggerHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Database Connectors API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                tryItOutEnabled: true,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function() {
                    console.log("Swagger UI loaded");
                },
                onFailure: function(data) {
                    console.log("Failed to load Swagger UI", data);
                }
            });
        };
    </script>
</body>
</html>
`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(swaggerHTML))
}

// SwaggerJSONHandler serves the Swagger JSON specification
func (s *Server) SwaggerJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try to read swagger.json from docs directory
	swaggerPath := filepath.Join("docs", "swagger.json")
	swaggerData, err := ioutil.ReadFile(swaggerPath)
	if err != nil {
		// If file doesn't exist, return a basic spec
		basicSpec := map[string]interface{}{
			"openapi": "3.0.3",
			"info": map[string]interface{}{
				"title":   "Database Connectors API",
				"version": "1.0.0",
			},
			"paths": map[string]interface{}{
				"/health": map[string]interface{}{
					"get": map[string]interface{}{
						"summary": "Health check",
						"responses": map[string]interface{}{
							"200": map[string]interface{}{
								"description": "Service is healthy",
							},
						},
					},
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(basicSpec)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(swaggerData)
}

// SwaggerYAMLHandler serves the Swagger YAML specification
func (s *Server) SwaggerYAMLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try to read swagger.yaml from docs directory
	swaggerPath := filepath.Join("docs", "swagger.yaml")
	swaggerData, err := ioutil.ReadFile(swaggerPath)
	if err != nil {
		http.Error(w, "Swagger YAML file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(swaggerData)
}

// DocumentationIndexHandler serves the documentation landing page
func (s *Server) DocumentationIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only serve on root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Try to read index.html from docs directory
	indexPath := filepath.Join("docs", "index.html")
	indexData, err := ioutil.ReadFile(indexPath)
	if err != nil {
		// If file doesn't exist, return a simple landing page
		simplePage := `
<!DOCTYPE html>
<html>
<head>
    <title>Database Connectors API</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .btn { background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; margin: 10px; display: inline-block; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🗄️ Database Connectors API</h1>
        <p>Welcome to the Database Connectors API with maker-checker workflow!</p>
        <div>
            <a href="/docs" class="btn">📚 API Documentation</a>
            <a href="/swagger.json" class="btn">📄 OpenAPI JSON</a>
            <a href="/swagger.yaml" class="btn">📄 OpenAPI YAML</a>
        </div>
        <h2>Endpoints</h2>
        <ul>
            <li><strong>GET /health</strong> - Health check</li>
            <li><strong>POST /test-connection</strong> - Test database connection</li>
            <li><strong>POST /execute</strong> - Execute database operations</li>
            <li><strong>POST /allconfig</strong> - Check AllConfig table</li>
            <li><strong>POST /allconfig-operation</strong> - Perform AllConfig operations</li>
        </ul>
    </div>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(simplePage))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(indexData)
}

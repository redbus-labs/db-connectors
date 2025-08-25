package api

import (
	"fmt"
	"log"
	"net/http"
)

// Server represents the HTTP server
type Server struct {
	api  *API
	port int
}

// NewServer creates a new HTTP server
func NewServer(port int) *Server {
	return &Server{
		api:  NewAPI(),
		port: port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", s.api.HealthHandler)
	mux.HandleFunc("/test-connection", s.api.TestConnectionHandler)
	mux.HandleFunc("/execute", s.api.ExecuteOperationHandler)
	mux.HandleFunc("/allconfig", s.api.AllConfigHandler)
	mux.HandleFunc("/allconfig-operation", s.api.AllConfigOperationHandler)
	
	// Swagger documentation routes
	mux.HandleFunc("/", s.DocumentationIndexHandler)
	mux.HandleFunc("/docs", s.SwaggerHandler)
	mux.HandleFunc("/docs/", s.SwaggerHandler)
	mux.HandleFunc("/swagger.json", s.SwaggerJSONHandler)
	mux.HandleFunc("/swagger.yaml", s.SwaggerYAMLHandler)

	// Add CORS middleware
	handler := s.corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🚀 Database Connectors API server starting on %s", addr)
	log.Printf("📡 Endpoints:")
	log.Printf("   GET  /                   - Documentation landing page")
	log.Printf("   GET  /health             - Health check")
	log.Printf("   POST /test-connection    - Test database connection")
	log.Printf("   POST /execute            - Execute database operation")
	log.Printf("   POST /allconfig          - Check/manage allconfig table")
	log.Printf("   POST /allconfig-operation - Perform operations on allconfig table")
	log.Printf("   GET  /docs               - Swagger UI documentation")
	log.Printf("   GET  /swagger.json       - OpenAPI JSON specification")
	log.Printf("   GET  /swagger.yaml       - OpenAPI YAML specification")
	log.Printf("")
	log.Printf("🌐 Visit http://localhost:%d for documentation", s.port)

	return http.ListenAndServe(addr, handler)
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

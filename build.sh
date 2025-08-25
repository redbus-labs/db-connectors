#!/bin/bash

# Database Connectors - Quick Build Script
set -e

APP_NAME="db-connectors"
VERSION="1.0.0"

echo "🔨 Database Connectors - Quick Build"
echo "===================================="

# Parse command line arguments
BUILD_TYPE="local"
if [ "$1" = "prod" ] || [ "$1" = "production" ]; then
    BUILD_TYPE="production"
elif [ "$1" = "docker" ]; then
    BUILD_TYPE="docker"
elif [ "$1" = "all" ]; then
    BUILD_TYPE="all"
fi

echo "📋 Build Type: $BUILD_TYPE"
echo ""

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -f ${APP_NAME} ${APP_NAME}-*

# Update dependencies
echo "📦 Updating dependencies..."
go mod tidy

case $BUILD_TYPE in
    "local")
        echo "🔨 Building for local development..."
        go build -o ${APP_NAME} cmd/main.go
        echo "✅ Local build complete: ./${APP_NAME}"
        echo "🚀 Run with: ./${APP_NAME}"
        ;;
        
    "production")
        echo "🔨 Building for production (Linux)..."
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
            -ldflags="-s -w -X main.Version=${VERSION}" \
            -o ${APP_NAME}-linux cmd/main.go
        echo "✅ Production build complete: ./${APP_NAME}-linux"
        echo "📤 Upload to server and run: ./${APP_NAME}-linux"
        ;;
        
    "docker")
        echo "🐳 Building Docker image..."
        # First build Linux binary
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
            -ldflags="-s -w -X main.Version=${VERSION}" \
            -o ${APP_NAME} cmd/main.go
            
        # Create simple Dockerfile
        cat > Dockerfile << 'EOF'
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY db-connectors .
COPY docs/ ./docs/
COPY examples/ ./examples/
COPY README.md .
COPY config.yaml ./config.example.yaml
RUN chmod +x ./db-connectors
EXPOSE 8080
CMD ["./db-connectors"]
EOF
        
        docker build -t ${APP_NAME}:${VERSION} .
        docker tag ${APP_NAME}:${VERSION} ${APP_NAME}:latest
        echo "✅ Docker build complete"
        echo "🚀 Run with: docker run -p 8080:8080 ${APP_NAME}:latest"
        ;;
        
    "all")
        echo "🔨 Building all variants..."
        
        # Local build
        echo "  📱 Local..."
        go build -o ${APP_NAME} cmd/main.go
        
        # Linux build
        echo "  🐧 Linux..."
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
            -ldflags="-s -w -X main.Version=${VERSION}" \
            -o ${APP_NAME}-linux cmd/main.go
            
        # Windows build
        echo "  🪟 Windows..."
        CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
            -ldflags="-s -w -X main.Version=${VERSION}" \
            -o ${APP_NAME}-windows.exe cmd/main.go
            
        # macOS build
        echo "  🍎 macOS..."
        CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
            -ldflags="-s -w -X main.Version=${VERSION}" \
            -o ${APP_NAME}-macos cmd/main.go
            
        echo "✅ All builds complete:"
        ls -la ${APP_NAME}*
        ;;
        
    *)
        echo "❌ Unknown build type: $BUILD_TYPE"
        echo ""
        echo "Usage: $0 [local|prod|docker|all]"
        echo "  local  - Build for current platform (default)"
        echo "  prod   - Build for Linux production servers"
        echo "  docker - Build Docker image"
        echo "  all    - Build for all platforms"
        exit 1
        ;;
esac

echo ""
echo "🎉 Build completed successfully!"

# Show next steps based on build type
case $BUILD_TYPE in
    "local")
        echo ""
        echo "🚀 Next steps:"
        echo "  ./${APP_NAME}"
        echo "  curl http://localhost:8080/health"
        ;;
    "production")
        echo ""
        echo "🚀 Deployment steps:"
        echo "  1. scp ${APP_NAME}-linux user@server:/path/"
        echo "  2. ssh user@server"
        echo "  3. chmod +x ${APP_NAME}-linux && ./${APP_NAME}-linux"
        ;;
    "docker")
        echo ""
        echo "🚀 Docker steps:"
        echo "  docker run -p 8080:8080 ${APP_NAME}:latest"
        echo "  curl http://localhost:8080/health"
        ;;
esac

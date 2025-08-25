#!/bin/bash

# Database Connectors - Production Deployment Script
set -e

echo "🚀 Building Database Connectors for Production Deployment..."

# Configuration
APP_NAME="db-connectors"
VERSION="1.0.0"
BUILD_DIR="deploy"
ARCHIVE_NAME="${APP_NAME}-v${VERSION}.tar.gz"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf ${BUILD_DIR}/
rm -f ${APP_NAME} ${APP_NAME}-*

# Ensure dependencies
echo "📦 Updating dependencies..."
go mod tidy

# Build optimized binary for Linux
echo "🔨 Building optimized binary for Linux..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w -X main.Version=${VERSION}" \
  -o ${APP_NAME}-linux cmd/main.go

# Create deployment structure
echo "📁 Creating deployment structure..."
mkdir -p ${BUILD_DIR}/${APP_NAME}

# Copy binary
cp ${APP_NAME}-linux ${BUILD_DIR}/${APP_NAME}/${APP_NAME}
chmod +x ${BUILD_DIR}/${APP_NAME}/${APP_NAME}

# Copy documentation and configuration
cp -r docs/ ${BUILD_DIR}/${APP_NAME}/
cp -r examples/ ${BUILD_DIR}/${APP_NAME}/
cp README.md ${BUILD_DIR}/${APP_NAME}/
cp config.yaml ${BUILD_DIR}/${APP_NAME}/config.example.yaml

# Create systemd service file
cat > ${BUILD_DIR}/${APP_NAME}/db-connectors.service << 'EOF'
[Unit]
Description=Database Connectors API
After=network.target

[Service]
Type=simple
User=dbconnectors
WorkingDirectory=/opt/db-connectors
ExecStart=/opt/db-connectors/db-connectors
Restart=always
RestartSec=5
Environment=PORT=8080
Environment=HOST=0.0.0.0

[Install]
WantedBy=multi-user.target
EOF

# Create installation script
cat > ${BUILD_DIR}/${APP_NAME}/install.sh << 'EOF'
#!/bin/bash
set -e

echo "🔧 Installing Database Connectors..."

# Create user and directories
sudo useradd -r -s /bin/false dbconnectors || true
sudo mkdir -p /opt/db-connectors
sudo cp -r ./* /opt/db-connectors/
sudo chown -R dbconnectors:dbconnectors /opt/db-connectors

# Install systemd service
sudo cp db-connectors.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable db-connectors

echo "✅ Installation complete!"
echo ""
echo "To start the service:"
echo "  sudo systemctl start db-connectors"
echo ""
echo "To check status:"
echo "  sudo systemctl status db-connectors"
echo ""
echo "To view logs:"
echo "  sudo journalctl -u db-connectors -f"
echo ""
echo "API will be available at: http://localhost:8080"
EOF

chmod +x ${BUILD_DIR}/${APP_NAME}/install.sh

# Create Docker files
cat > ${BUILD_DIR}/${APP_NAME}/Dockerfile << 'EOF'
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

# Copy binary and docs
COPY db-connectors .
COPY docs/ ./docs/
COPY examples/ ./examples/
COPY README.md .
COPY config.example.yaml .

# Make binary executable
RUN chmod +x ./db-connectors

# Expose port
EXPOSE 8080

# Create non-root user
RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser && \
    chown -R appuser:appgroup /app

USER appuser

# Run the binary
CMD ["./db-connectors"]
EOF

cat > ${BUILD_DIR}/${APP_NAME}/docker-compose.yml << 'EOF'
version: '3.8'

services:
  db-connectors:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - HOST=0.0.0.0
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
EOF

# Create deployment README
cat > ${BUILD_DIR}/${APP_NAME}/DEPLOYMENT.md << 'EOF'
# Database Connectors - Deployment Guide

## Quick Start

### Option 1: Direct Installation
```bash
# Extract and install
tar -xzf db-connectors-v1.0.0.tar.gz
cd db-connectors
sudo ./install.sh

# Start service
sudo systemctl start db-connectors
```

### Option 2: Docker
```bash
# Extract files
tar -xzf db-connectors-v1.0.0.tar.gz
cd db-connectors

# Build and run with Docker
docker-compose up -d
```

### Option 3: Manual Installation
```bash
# Extract files
tar -xzf db-connectors-v1.0.0.tar.gz
cd db-connectors

# Run directly
./db-connectors
```

## Configuration

### Environment Variables
- `PORT`: Server port (default: 8080)
- `HOST`: Bind address (default: 0.0.0.0)

### Custom Configuration
Copy `config.example.yaml` to `config.yaml` and modify as needed.

## API Access

Once running, the API will be available at:
- Health Check: http://localhost:8080/health
- API Documentation: http://localhost:8080/
- Swagger UI: http://localhost:8080/docs

## Firewall Configuration

For cloud deployments, ensure these ports are open:
- Port 8080 (HTTP API)

## Troubleshooting

### Check Service Status
```bash
sudo systemctl status db-connectors
```

### View Logs
```bash
sudo journalctl -u db-connectors -f
```

### Test Connectivity
```bash
curl http://localhost:8080/health
```
EOF

# Create archive
echo "📦 Creating deployment archive..."
cd ${BUILD_DIR}
tar -czf ${ARCHIVE_NAME} ${APP_NAME}/
cd ..

# Display results
echo ""
echo "✅ Production deployment package created successfully!"
echo ""
echo "📦 Package: ${BUILD_DIR}/${ARCHIVE_NAME}"
echo "📁 Size: $(du -h ${BUILD_DIR}/${ARCHIVE_NAME} | cut -f1)"
echo ""
echo "📋 Package contents:"
tar -tzf ${BUILD_DIR}/${ARCHIVE_NAME} | head -20
echo ""
echo "🚀 Deployment instructions:"
echo "1. Upload ${BUILD_DIR}/${ARCHIVE_NAME} to your server"
echo "2. Extract: tar -xzf ${ARCHIVE_NAME}"
echo "3. Install: cd ${APP_NAME} && sudo ./install.sh"
echo "4. Start: sudo systemctl start db-connectors"
echo ""
echo "🌐 API will be available at: http://your-server:8080"

# Clean up build artifacts
rm -f ${APP_NAME}-linux

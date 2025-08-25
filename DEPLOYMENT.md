# Database Connectors - Production Deployment Guide

## 🚀 Quick Deployment Options

### Option 1: Automated Production Build
```bash
# Make scripts executable
chmod +x deploy.sh build.sh

# Create complete deployment package
./deploy.sh

# Result: deploy/db-connectors-v1.0.0.tar.gz ready for deployment
```

### Option 2: Quick Build for Different Platforms
```bash
# Local development
./build.sh local

# Production Linux server
./build.sh prod

# Docker image
./build.sh docker

# All platforms
./build.sh all
```

### Option 3: Manual Build
```bash
# For Linux servers
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o db-connectors cmd/main.go

# For current platform
go build -o db-connectors cmd/main.go
```

## 📦 Deployment Package Contents

The `deploy.sh` script creates a complete deployment package with:

```
db-connectors-v1.0.0.tar.gz
└── db-connectors/
    ├── db-connectors              # Main binary
    ├── docs/                      # API Documentation
    │   ├── index.html            # Landing page
    │   ├── swagger.json          # OpenAPI spec
    │   ├── swagger.yaml          # OpenAPI spec
    │   └── postman_collection.json
    ├── examples/                  # Usage examples
    ├── README.md                  # Project documentation
    ├── config.example.yaml       # Configuration template
    ├── db-connectors.service     # Systemd service file
    ├── install.sh                # Installation script
    ├── Dockerfile                # Docker build file
    ├── docker-compose.yml        # Docker Compose config
    └── DEPLOYMENT.md             # This file
```

## 🖥️ Server Deployment Methods

### Method 1: Systemd Service (Recommended)
```bash
# On server
tar -xzf db-connectors-v1.0.0.tar.gz
cd db-connectors
sudo ./install.sh

# Start service
sudo systemctl start db-connectors
sudo systemctl status db-connectors

# Enable auto-start
sudo systemctl enable db-connectors
```

### Method 2: Docker Deployment
```bash
# Extract and run with Docker
tar -xzf db-connectors-v1.0.0.tar.gz
cd db-connectors
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f
```

### Method 3: Direct Execution
```bash
# Extract and run directly
tar -xzf db-connectors-v1.0.0.tar.gz
cd db-connectors

# Run in foreground
./db-connectors

# Run in background
nohup ./db-connectors > db-connectors.log 2>&1 &
```

### Method 4: Process Manager (PM2)
```bash
# Install PM2
npm install -g pm2

# Create ecosystem file
cat > ecosystem.config.js << 'EOF'
module.exports = {
  apps: [{
    name: 'db-connectors',
    script: './db-connectors',
    cwd: '/opt/db-connectors',
    env: {
      PORT: 8080,
      HOST: '0.0.0.0'
    },
    restart_delay: 5000,
    max_restarts: 10
  }]
}
EOF

# Start with PM2
pm2 start ecosystem.config.js
pm2 startup
pm2 save
```

## 🌐 Cloud Platform Deployment

### AWS EC2
1. **Launch EC2 instance** (Ubuntu 20.04+ recommended)
2. **Upload package**: `scp db-connectors-v1.0.0.tar.gz ubuntu@your-ec2:/home/ubuntu/`
3. **Install**: `tar -xzf db-connectors-v1.0.0.tar.gz && cd db-connectors && sudo ./install.sh`
4. **Configure Security Group**: Allow inbound port 8080
5. **Start service**: `sudo systemctl start db-connectors`

### Google Cloud Platform
1. **Create Compute Engine instance**
2. **Upload and install** (same as AWS)
3. **Configure firewall**: `gcloud compute firewall-rules create allow-db-connectors --allow tcp:8080`

### DigitalOcean Droplet
1. **Create droplet** (Ubuntu 20.04+)
2. **Upload and install** (same as above)
3. **Configure firewall**: Add rule for port 8080

### Docker on Cloud
```bash
# Build and push to registry
docker build -t your-registry/db-connectors:v1.0.0 .
docker push your-registry/db-connectors:v1.0.0

# Deploy on cloud
docker run -d -p 8080:8080 --name db-connectors your-registry/db-connectors:v1.0.0
```

## ⚙️ Configuration

### Environment Variables
```bash
export PORT=8080              # Server port
export HOST=0.0.0.0           # Bind address (0.0.0.0 for all interfaces)
```

### Configuration File
```bash
# Copy template
cp config.example.yaml config.yaml

# Edit configuration
nano config.yaml
```

### Systemd Environment
```bash
# Edit service file
sudo systemctl edit db-connectors

# Add environment variables
[Service]
Environment="PORT=8080"
Environment="HOST=0.0.0.0"
```

## 🔒 Security Considerations

### Firewall Configuration
```bash
# UFW (Ubuntu)
sudo ufw allow 8080/tcp
sudo ufw enable

# iptables
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
```

### Reverse Proxy (Nginx)
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### SSL/TLS (Let's Encrypt)
```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d your-domain.com
```

## 📊 Monitoring & Maintenance

### Service Status
```bash
# Check status
sudo systemctl status db-connectors

# View logs
sudo journalctl -u db-connectors -f

# Restart service
sudo systemctl restart db-connectors
```

### Health Checks
```bash
# Basic health check
curl http://localhost:8080/health

# Full API check
curl http://localhost:8080/docs
```

### Log Rotation
```bash
# Create logrotate config
sudo tee /etc/logrotate.d/db-connectors << 'EOF'
/var/log/db-connectors/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    copytruncate
}
EOF
```

## 🚀 API Access

Once deployed, your API will be available at:

- **Health Check**: `http://your-server:8080/health`
- **API Documentation**: `http://your-server:8080/`
- **Swagger UI**: `http://your-server:8080/docs`
- **OpenAPI JSON**: `http://your-server:8080/swagger.json`
- **OpenAPI YAML**: `http://your-server:8080/swagger.yaml`

## 🔧 Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   sudo lsof -i :8080
   sudo kill <PID>
   ```

2. **Permission denied**
   ```bash
   chmod +x db-connectors
   ```

3. **Service won't start**
   ```bash
   sudo journalctl -u db-connectors --no-pager
   ```

4. **Can't connect from outside**
   - Check firewall rules
   - Verify HOST=0.0.0.0 (not 127.0.0.1)
   - Check cloud security groups

### Performance Tuning

```bash
# Increase file limits
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf

# Optimize kernel parameters
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
sysctl -p
```

## 📈 Scaling

### Horizontal Scaling (Load Balancer)
```nginx
upstream db_connectors {
    server 10.0.1.10:8080;
    server 10.0.1.11:8080;
    server 10.0.1.12:8080;
}

server {
    listen 80;
    location / {
        proxy_pass http://db_connectors;
    }
}
```

### Vertical Scaling
- Increase server resources (CPU, RAM)
- Optimize database connections
- Use connection pooling

---

## 📞 Support

For issues or questions:
1. Check logs: `sudo journalctl -u db-connectors`
2. Test health endpoint: `curl http://localhost:8080/health`
3. Verify configuration files
4. Check firewall and network settings

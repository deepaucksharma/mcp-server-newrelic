# Deployment Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Deployment Options](#deployment-options)
5. [Production Setup](#production-setup)
6. [Monitoring](#monitoring)
7. [Security](#security)
8. [Scaling](#scaling)
9. [Troubleshooting](#troubleshooting)
10. [Maintenance](#maintenance)

## Prerequisites

### Infrastructure Requirements

#### Minimum Requirements (Development)
- **CPU**: 2 cores
- **RAM**: 4 GB
- **Storage**: 10 GB SSD
- **Network**: 100 Mbps

#### Recommended Requirements (Production)
- **CPU**: 4+ cores
- **RAM**: 8+ GB
- **Storage**: 20+ GB SSD with high IOPS
- **Network**: 1 Gbps with low latency to New Relic

### Software Requirements
- **Go**: 1.21+ (for building from source)
- **Docker**: 20.10+ (for container deployment)
- **Kubernetes**: 1.24+ (for K8s deployment)
- **Redis**: 7+ (optional, for distributed state)

### New Relic Requirements
- Valid New Relic API Key with appropriate permissions
- Account ID
- Network access to New Relic API endpoints

## Quick Start

### 1. Local Development

```bash
# Clone the repository
git clone https://github.com/youraccount/mcp-server-newrelic.git
cd mcp-server-newrelic

# Copy environment template
cp .env.example .env

# Edit .env with your credentials
vim .env

# Build and run
make build
make run
```

### 2. Docker Quick Start

```bash
# Using pre-built image
docker run -d \
  --name mcp-server \
  -e NEW_RELIC_API_KEY="your-api-key" \
  -e NEW_RELIC_ACCOUNT_ID="your-account-id" \
  -p 8080:8080 \
  newrelic/mcp-server:latest

# Using docker-compose
docker-compose up -d
```

### 3. Verify Installation

```bash
# Check health
curl http://localhost:8080/health

# Test with MCP inspector
npx @modelcontextprotocol/inspector http://localhost:8080
```

## Configuration

### Environment Variables

Create a `.env` file with the following variables:

```bash
# === REQUIRED CONFIGURATION ===
NEW_RELIC_API_KEY=your_api_key
NEW_RELIC_ACCOUNT_ID=your_account_id

# === OPTIONAL CONFIGURATION ===
# Server Settings
MCP_TRANSPORT=stdio          # Options: stdio, http, sse
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
LOG_LEVEL=INFO               # Options: DEBUG, INFO, WARN, ERROR

# Performance Settings
REQUEST_TIMEOUT=30s
MAX_CONCURRENT_REQUESTS=100
CACHE_TTL=300

# Redis Configuration (optional)
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Security Settings
AUTH_ENABLED=false
RATE_LIMIT_ENABLED=true
RATE_LIMIT_PER_MIN=100

# Feature Flags
ENABLE_MOCK_MODE=false
ENABLE_DRY_RUN=true
ENABLE_EXPERIMENTAL=false
```

### Configuration Validation

```bash
# Validate configuration
./bin/mcp-server validate-config

# Run diagnostics
make diagnose

# Auto-fix common issues
make diagnose-fix
```

## Deployment Options

### Option 1: Binary Deployment

```bash
# Download latest release
curl -L https://github.com/youraccount/mcp-server-newrelic/releases/latest/download/mcp-server-linux-amd64 -o mcp-server
chmod +x mcp-server

# Create systemd service
sudo cat > /etc/systemd/system/mcp-server.service <<EOF
[Unit]
Description=New Relic MCP Server
After=network.target

[Service]
Type=simple
User=mcp-server
Group=mcp-server
ExecStart=/usr/local/bin/mcp-server
Restart=always
RestartSec=10
EnvironmentFile=/etc/mcp-server/production.env

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl enable mcp-server
sudo systemctl start mcp-server
```

### Option 2: Docker Deployment

```bash
# Build custom image
docker build -t mcp-server:custom .

# Run with docker
docker run -d \
  --name mcp-server \
  --restart always \
  -v /etc/mcp-server:/config \
  -p 8080:8080 \
  --env-file /etc/mcp-server/production.env \
  mcp-server:custom

# Or use docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

### Option 3: Kubernetes Deployment

```yaml
# mcp-server-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server
  labels:
    app: mcp-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mcp-server
  template:
    metadata:
      labels:
        app: mcp-server
    spec:
      containers:
      - name: mcp-server
        image: newrelic/mcp-server:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: MCP_TRANSPORT
          value: "http"
        envFrom:
        - secretRef:
            name: mcp-server-secrets
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: mcp-server
spec:
  selector:
    app: mcp-server
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: LoadBalancer
```

Deploy to Kubernetes:

```bash
# Create secrets
kubectl create secret generic mcp-server-secrets \
  --from-env-file=/etc/mcp-server/production.env

# Apply deployment
kubectl apply -f mcp-server-deployment.yaml

# Check status
kubectl get pods -l app=mcp-server
kubectl get svc mcp-server
```

## Production Setup

### 1. Security Hardening

```bash
# Create dedicated user
sudo useradd -r -s /bin/false mcp-server

# Set file permissions
sudo chown -R mcp-server:mcp-server /etc/mcp-server
sudo chmod 600 /etc/mcp-server/production.env

# Configure firewall
sudo ufw allow 8080/tcp
sudo ufw allow 9090/tcp  # Metrics port
```

### 2. TLS Configuration

```bash
# Generate self-signed cert (for testing)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/mcp-server.key \
  -out /etc/ssl/certs/mcp-server.crt

# Configure TLS in environment
cat >> /etc/mcp-server/production.env <<EOF
TLS_ENABLED=true
TLS_CERT_FILE=/etc/ssl/certs/mcp-server.crt
TLS_KEY_FILE=/etc/ssl/private/mcp-server.key
EOF
```

### 3. Load Balancer Configuration

```nginx
# nginx.conf
upstream mcp_servers {
    least_conn;
    server mcp-server-1:8080 max_fails=3 fail_timeout=30s;
    server mcp-server-2:8080 max_fails=3 fail_timeout=30s;
    server mcp-server-3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 443 ssl http2;
    server_name mcp.example.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    location / {
        proxy_pass http://mcp_servers;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support for SSE
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location /metrics {
        allow 10.0.0.0/8;  # Internal network only
        deny all;
        proxy_pass http://mcp_servers:9090/metrics;
    }
}
```

### 4. Redis Setup for High Availability

```yaml
# redis-ha.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  redis.conf: |
    maxmemory 2gb
    maxmemory-policy allkeys-lru
    save 900 1
    save 300 10
    save 60 10000
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
spec:
  serviceName: redis
  replicas: 3
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command: ["redis-server", "/etc/redis/redis.conf"]
        volumeMounts:
        - name: config
          mountPath: /etc/redis
        - name: data
          mountPath: /data
      volumes:
      - name: config
        configMap:
          name: redis-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

## Monitoring

### 1. Health Checks

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health with dependencies
curl http://localhost:8080/health?detailed=true

# Readiness check
curl http://localhost:8080/ready
```

### 2. Prometheus Metrics

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'mcp-server'
    static_configs:
      - targets: ['mcp-server:9090']
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'go_.*'
        action: drop  # Drop Go runtime metrics if not needed
```

Available metrics:
- `mcp_requests_total{tool,status}` - Request count by tool and status
- `mcp_request_duration_seconds{tool}` - Request duration histogram
- `mcp_active_sessions` - Current active sessions
- `mcp_cache_hits_total{cache_type}` - Cache hit rate
- `mcp_newrelic_api_calls_total{endpoint}` - NR API call count

### 3. New Relic APM Integration

```bash
# Add to environment configuration
NEW_RELIC_LICENSE_KEY=your_license_key
NEW_RELIC_APP_NAME=mcp-server-production
NEW_RELIC_DISTRIBUTED_TRACING_ENABLED=true
```

### 4. Logging

```bash
# Configure structured logging
LOG_FORMAT=json
LOG_LEVEL=INFO

# View logs
docker logs -f mcp-server

# Or with journalctl
journalctl -u mcp-server -f

# Parse JSON logs
journalctl -u mcp-server -o json | jq '.MESSAGE | fromjson'
```

## Security

### 1. API Key Management

```bash
# Rotate API keys
./bin/mcp-server rotate-keys --backup

# Encrypt sensitive configuration
./bin/mcp-server encrypt-config --input .env --output .env.encrypted
```

### 2. Network Security

```yaml
# NetworkPolicy for Kubernetes
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mcp-server-netpol
spec:
  podSelector:
    matchLabels:
      app: mcp-server
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443  # New Relic API
```

### 3. Rate Limiting

```bash
# Configure rate limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_PER_MIN=100
RATE_LIMIT_BURST=20

# Per-tool rate limits
RATE_LIMIT_RULES='
{
  "nrql.execute": 50,
  "discovery.*": 20,
  "dashboard.create": 5
}'
```

## Scaling

### 1. Horizontal Scaling

```bash
# Kubernetes autoscaling
kubectl autoscale deployment mcp-server \
  --min=2 \
  --max=10 \
  --cpu-percent=70

# Or with custom metrics
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: mcp-server-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: mcp-server
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: mcp_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
```

### 2. Vertical Scaling

```yaml
# Resource recommendations based on load
resources:
  small:  # < 100 req/min
    requests:
      memory: "256Mi"
      cpu: "250m"
    limits:
      memory: "512Mi"
      cpu: "500m"
  
  medium: # 100-1000 req/min
    requests:
      memory: "1Gi"
      cpu: "1000m"
    limits:
      memory: "2Gi"
      cpu: "2000m"
  
  large:  # > 1000 req/min
    requests:
      memory: "4Gi"
      cpu: "4000m"
    limits:
      memory: "8Gi"
      cpu: "8000m"
```

### 3. Caching Strategy

```bash
# Configure multi-level caching
CACHE_L1_SIZE=1000        # In-memory LRU
CACHE_L1_TTL=300          # 5 minutes
CACHE_L2_ENABLED=true     # Redis
CACHE_L2_TTL=3600         # 1 hour
CACHE_COMPRESSION=true    # For large results
```

## Troubleshooting

### Common Issues

#### 1. Connection Timeouts

```bash
# Check connectivity
curl -v https://api.newrelic.com/v2/accounts.json \
  -H "Api-Key: $NEW_RELIC_API_KEY"

# Increase timeout
REQUEST_TIMEOUT=60s
NEW_RELIC_API_TIMEOUT=30s

# Check DNS resolution
nslookup api.newrelic.com
```

#### 2. High Memory Usage

```bash
# Monitor memory
docker stats mcp-server

# Tune garbage collection
GOGC=50  # More aggressive GC
GOMEMLIMIT=1GiB  # Set memory limit

# Profile memory usage
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

#### 3. Slow Queries

```bash
# Enable query profiling
ENABLE_QUERY_PROFILING=true
SLOW_QUERY_THRESHOLD=5s

# Check slow query log
tail -f /var/log/mcp-server/slow-queries.log

# Optimize problematic queries
./bin/mcp-server analyze-query --query "SELECT ..."
```

### Debug Mode

```bash
# Enable debug mode
LOG_LEVEL=DEBUG
MCP_DEBUG=true
VERBOSE_LOGGING=true

# Enable pprof profiling
ENABLE_PPROF=true
PPROF_PORT=6060

# CPU profiling
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

## Maintenance

### 1. Backup Procedures

```bash
# Backup Redis state
redis-cli --rdb /backup/redis-$(date +%Y%m%d).rdb

# Backup configuration
tar -czf /backup/config-$(date +%Y%m%d).tar.gz /etc/mcp-server

# Automated backup script
cat > /etc/cron.daily/mcp-backup <<'EOF'
#!/bin/bash
BACKUP_DIR="/backup/mcp-server"
mkdir -p $BACKUP_DIR
redis-cli --rdb $BACKUP_DIR/redis-$(date +%Y%m%d).rdb
tar -czf $BACKUP_DIR/config-$(date +%Y%m%d).tar.gz /etc/mcp-server
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete
EOF
chmod +x /etc/cron.daily/mcp-backup
```

### 2. Updates and Upgrades

```bash
# Check current version
./bin/mcp-server version

# Rolling update in Kubernetes
kubectl set image deployment/mcp-server \
  mcp-server=newrelic/mcp-server:v2.0.0 \
  --record

# Monitor rollout
kubectl rollout status deployment/mcp-server
kubectl rollout history deployment/mcp-server

# Rollback if needed
kubectl rollout undo deployment/mcp-server
```

### 3. Log Rotation

```bash
# Configure logrotate
cat > /etc/logrotate.d/mcp-server <<EOF
/var/log/mcp-server/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 mcp-server mcp-server
    postrotate
        systemctl reload mcp-server
    endscript
}
EOF
```

### 4. Performance Monitoring

```bash
# Regular performance checks
cat > /etc/cron.hourly/mcp-perf-check <<'EOF'
#!/bin/bash
# Check response time
RESPONSE_TIME=$(curl -w "%{time_total}" -o /dev/null -s http://localhost:8080/health)
if (( $(echo "$RESPONSE_TIME > 1" | bc -l) )); then
    echo "Warning: Health check took ${RESPONSE_TIME}s" | logger -t mcp-server
fi

# Check memory usage
MEMORY=$(ps aux | grep mcp-server | awk '{print $6}' | head -1)
if [ $MEMORY -gt 2000000 ]; then  # 2GB
    echo "Warning: High memory usage: ${MEMORY}KB" | logger -t mcp-server
fi
EOF
chmod +x /etc/cron.hourly/mcp-perf-check
```

---

This deployment guide provides comprehensive instructions for running the New Relic MCP Server in production. Always test configuration changes in a staging environment before applying to production.
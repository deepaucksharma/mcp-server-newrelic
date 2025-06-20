# New Relic MCP Server - Production Deployment Guide

This guide provides comprehensive instructions for deploying the New Relic MCP Server in production environments with enterprise-grade security, reliability, and performance.

## Table of Contents

1. [Pre-deployment Checklist](#pre-deployment-checklist)
2. [Environment Setup and Configuration](#environment-setup-and-configuration)
3. [Docker Deployment](#docker-deployment)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Monitoring and Observability](#monitoring-and-observability)
6. [Security Hardening](#security-hardening)
7. [Backup and Disaster Recovery](#backup-and-disaster-recovery)
8. [Performance Tuning](#performance-tuning)
9. [Troubleshooting Deployment Issues](#troubleshooting-deployment-issues)
10. [Rollback Procedures](#rollback-procedures)

## Pre-deployment Checklist

### Infrastructure Requirements

- [ ] **Compute Resources**
  - Minimum: 2 vCPUs, 4GB RAM
  - Recommended: 4 vCPUs, 8GB RAM
  - Storage: 20GB minimum (for logs and cache)

- [ ] **Network Requirements**
  - Outbound HTTPS access to New Relic API endpoints
  - Internal network access for Redis (if using distributed cache)
  - Load balancer with SSL termination capability

- [ ] **Dependencies**
  - Redis 6.0+ (optional, for distributed state)
  - Docker 20.10+ or Kubernetes 1.21+
  - TLS certificates for production endpoints

### Security Prerequisites

- [ ] **API Keys and Secrets**
  ```bash
  # Generate secure JWT secret
  export JWT_SECRET=$(openssl rand -base64 32)
  
  # Generate API key salt
  export API_KEY_SALT=$(openssl rand -base64 16)
  
  # Store in secure secret management system
  ```

- [ ] **New Relic Credentials**
  - Valid New Relic User API Key with required permissions
  - Account ID verified and accessible
  - License key for APM monitoring (optional but recommended)

- [ ] **TLS Certificates**
  - Valid SSL/TLS certificates for production domain
  - Certificate chain properly configured
  - Private key secured with appropriate permissions

### Validation Steps

```bash
# Run pre-deployment diagnostics
./bin/diagnose --pre-deploy

# Verify New Relic connectivity
curl -H "Api-Key: $NEW_RELIC_API_KEY" \
  https://api.newrelic.com/v2/accounts.json

# Test Redis connectivity (if applicable)
redis-cli -h $REDIS_HOST -p $REDIS_PORT ping
```

## Environment Setup and Configuration

### 1. Production Configuration File

Create a production environment file:

```bash
# /etc/mcp-server/production.env

# === REQUIRED CONFIGURATION ===
NEW_RELIC_API_KEY=<your-production-api-key>
NEW_RELIC_ACCOUNT_ID=<your-account-id>
JWT_SECRET=<generated-jwt-secret>
API_KEY_SALT=<generated-api-salt>

# === PRODUCTION SETTINGS ===
# Server Configuration
MCP_TRANSPORT=http
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
MCP_HTTP_PORT=8081
MCP_SSE_PORT=8082
REQUEST_TIMEOUT=30s
MAX_CONCURRENT_REQUESTS=200

# Security
AUTH_ENABLED=true
TLS_ENABLED=true
TLS_CERT_FILE=/etc/ssl/certs/mcp-server.crt
TLS_KEY_FILE=/etc/ssl/private/mcp-server.key
RATE_LIMIT_ENABLED=true
RATE_LIMIT_PER_MIN=100
RATE_LIMIT_BURST=20

# Logging
LOG_LEVEL=INFO
LOG_FORMAT=json
VERBOSE_LOGGING=false

# State Management
REDIS_URL=redis://redis.internal:6379
REDIS_PASSWORD=<redis-password>
REDIS_DB=0
SESSION_TTL=3600
CACHE_TTL=600

# New Relic APM
NEW_RELIC_LICENSE_KEY=<your-license-key>
NEW_RELIC_APP_NAME=mcp-server-production
NEW_RELIC_ENVIRONMENT=production
NEW_RELIC_DISTRIBUTED_TRACING_ENABLED=true

# Performance
DISCOVERY_CACHE_TTL=7200
DISCOVERY_MAX_WORKERS=20
METRICS_ENABLED=true
METRICS_PORT=9090
TRACING_ENABLED=true
TRACING_SAMPLE_RATE=0.1

# Feature Flags
ENABLE_PATTERN_DETECTION=true
ENABLE_QUERY_GENERATION=true
ENABLE_RELATIONSHIP_MINING=true
ENABLE_QUALITY_ASSESSMENT=true
ENABLE_INTELLIGENT_SAMPLING=true
ENABLE_EXPERIMENTAL_FEATURES=false
```

### 2. System Configuration

```bash
# Set system limits for production
cat >> /etc/security/limits.conf <<EOF
mcp-server soft nofile 65536
mcp-server hard nofile 65536
mcp-server soft nproc 32768
mcp-server hard nproc 32768
EOF

# Configure sysctl for better network performance
cat >> /etc/sysctl.conf <<EOF
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
EOF

sysctl -p
```

## Docker Deployment

### 1. Production Dockerfile

```dockerfile
# Dockerfile.production
FROM golang:1.21-alpine AS builder

# Install security updates
RUN apk update && apk upgrade && apk add --no-cache git make ca-certificates

WORKDIR /build

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always)" \
    -o mcp-server ./cmd/mcp-server

# Security scan
FROM aquasec/trivy:latest AS scanner
COPY --from=builder /build/mcp-server /mcp-server
RUN trivy fs --no-progress --severity HIGH,CRITICAL /mcp-server

# Final production image
FROM scratch

# Import certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /build/mcp-server /mcp-server

# Create non-root user
USER 10001:10001

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/mcp-server", "health"]

EXPOSE 8080 8081 8082 9090

ENTRYPOINT ["/mcp-server"]
```

### 2. Docker Compose Production

```yaml
# docker-compose.production.yml
version: '3.8'

services:
  mcp-server:
    image: ${DOCKER_REGISTRY}/mcp-server-newrelic:${VERSION:-latest}
    container_name: mcp-server-prod
    restart: always
    ports:
      - "8080:8080"
      - "8081:8081"
      - "8082:8082"
    volumes:
      - ./config:/app/config:ro
      - ./certs:/etc/ssl/certs:ro
      - logs:/app/logs
      - data:/app/data
    env_file:
      - /etc/mcp-server/production.env
    networks:
      - production
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "10"
    security_opt:
      - no-new-privileges:true
      - seccomp:unconfined
    read_only: true
    tmpfs:
      - /tmp

  redis:
    image: redis:7-alpine
    container_name: mcp-redis-prod
    restart: always
    command: >
      redis-server
      --requirepass ${REDIS_PASSWORD}
      --maxmemory 2gb
      --maxmemory-policy allkeys-lru
      --appendonly yes
      --appendfsync everysec
    volumes:
      - redis-data:/data
    networks:
      - production
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G

  nginx:
    image: nginx:alpine
    container_name: mcp-nginx-prod
    restart: always
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - mcp-server
    networks:
      - production

networks:
  production:
    driver: bridge
    ipam:
      config:
        - subnet: 172.30.0.0/16

volumes:
  logs:
    driver: local
  data:
    driver: local
  redis-data:
    driver: local
```

### 3. Nginx Configuration

```nginx
# nginx/nginx.conf
worker_processes auto;
worker_rlimit_nofile 65535;

events {
    worker_connections 4096;
    use epoll;
    multi_accept on;
}

http {
    # Security headers
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options DENY;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_status 429;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    upstream mcp_backend {
        least_conn;
        server mcp-server:8080 max_fails=3 fail_timeout=30s;
        keepalive 32;
    }

    server {
        listen 80;
        server_name _;
        return 301 https://$host$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name mcp.example.com;

        ssl_certificate /etc/nginx/certs/server.crt;
        ssl_certificate_key /etc/nginx/certs/server.key;

        # API endpoints
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            
            proxy_pass http://mcp_backend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        # Health check endpoint (no rate limit)
        location /health {
            proxy_pass http://mcp_backend;
            access_log off;
        }

        # Metrics endpoint (internal only)
        location /metrics {
            allow 10.0.0.0/8;
            deny all;
            proxy_pass http://mcp-server:9090/metrics;
        }
    }
}
```

## Kubernetes Deployment

### 1. Namespace and ConfigMap

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: mcp-server
  labels:
    name: mcp-server
    environment: production

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mcp-server-config
  namespace: mcp-server
data:
  server.yaml: |
    server:
      host: 0.0.0.0
      port: 8080
      timeout: 30s
    logging:
      level: info
      format: json
    features:
      pattern_detection: true
      query_generation: true
      relationship_mining: true
```

### 2. Secrets

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mcp-server-secrets
  namespace: mcp-server
type: Opaque
stringData:
  NEW_RELIC_API_KEY: "<base64-encoded-api-key>"
  NEW_RELIC_LICENSE_KEY: "<base64-encoded-license-key>"
  JWT_SECRET: "<base64-encoded-jwt-secret>"
  API_KEY_SALT: "<base64-encoded-salt>"
  REDIS_PASSWORD: "<base64-encoded-redis-password>"
```

### 3. Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server
  namespace: mcp-server
  labels:
    app: mcp-server
    version: v1.0.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: mcp-server
  template:
    metadata:
      labels:
        app: mcp-server
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: mcp-server
      securityContext:
        runAsNonRoot: true
        runAsUser: 10001
        fsGroup: 10001
      containers:
      - name: mcp-server
        image: your-registry/mcp-server-newrelic:v1.0.0
        imagePullPolicy: Always
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: mcp-http
          containerPort: 8081
          protocol: TCP
        - name: metrics
          containerPort: 9090
          protocol: TCP
        env:
        - name: NEW_RELIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: mcp-server-secrets
              key: NEW_RELIC_API_KEY
        - name: NEW_RELIC_ACCOUNT_ID
          value: "YOUR_ACCOUNT_ID"
        - name: NEW_RELIC_LICENSE_KEY
          valueFrom:
            secretKeyRef:
              name: mcp-server-secrets
              key: NEW_RELIC_LICENSE_KEY
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: mcp-server-secrets
              key: JWT_SECRET
        - name: API_KEY_SALT
          valueFrom:
            secretKeyRef:
              name: mcp-server-secrets
              key: API_KEY_SALT
        - name: REDIS_URL
          value: "redis://redis-service:6379"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mcp-server-secrets
              key: REDIS_PASSWORD
        - name: LOG_LEVEL
          value: "INFO"
        - name: NEW_RELIC_APP_NAME
          value: "mcp-server-k8s"
        - name: NEW_RELIC_ENVIRONMENT
          value: "production"
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: tmp
          mountPath: /tmp
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
      volumes:
      - name: config
        configMap:
          name: mcp-server-config
      - name: tmp
        emptyDir: {}
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - mcp-server
              topologyKey: kubernetes.io/hostname
```

### 4. Service and Ingress

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: mcp-server-service
  namespace: mcp-server
  labels:
    app: mcp-server
spec:
  type: ClusterIP
  selector:
    app: mcp-server
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: mcp-http
    port: 8081
    targetPort: 8081
  - name: metrics
    port: 9090
    targetPort: 9090

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: mcp-server-ingress
  namespace: mcp-server
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - mcp.example.com
    secretName: mcp-server-tls
  rules:
  - host: mcp.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: mcp-server-service
            port:
              number: 8080
```

### 5. Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: mcp-server-hpa
  namespace: mcp-server
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: mcp-server
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 2
        periodSeconds: 60
```

### 6. Redis StatefulSet

```yaml
# k8s/redis.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
  namespace: mcp-server
spec:
  serviceName: redis-service
  replicas: 1
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
        command:
        - redis-server
        - --requirepass
        - $(REDIS_PASSWORD)
        - --appendonly
        - "yes"
        - --maxmemory
        - "2gb"
        - --maxmemory-policy
        - "allkeys-lru"
        ports:
        - containerPort: 6379
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mcp-server-secrets
              key: REDIS_PASSWORD
        resources:
          requests:
            memory: "2Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "1000m"
        volumeMounts:
        - name: redis-data
          mountPath: /data
  volumeClaimTemplates:
  - metadata:
      name: redis-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi

---
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: mcp-server
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
```

## Monitoring and Observability

### 1. New Relic APM Integration

The MCP Server automatically reports to New Relic APM when configured with a license key. Key metrics include:

```yaml
# APM configuration (automatic with env vars)
NEW_RELIC_LICENSE_KEY: <your-license-key>
NEW_RELIC_APP_NAME: mcp-server-production
NEW_RELIC_DISTRIBUTED_TRACING_ENABLED: true
NEW_RELIC_LOG_LEVEL: info
```

### 2. Custom Dashboards

Create a New Relic dashboard for MCP Server monitoring:

```sql
-- NRQL queries for custom dashboard

-- Request rate
SELECT rate(count(*), 1 minute) 
FROM Transaction 
WHERE appName = 'mcp-server-production' 
FACET request.uri 
SINCE 30 minutes ago

-- Error rate
SELECT percentage(count(*), WHERE error IS true) 
FROM Transaction 
WHERE appName = 'mcp-server-production' 
SINCE 1 hour ago

-- Response time percentiles
SELECT percentile(duration, 50, 95, 99) 
FROM Transaction 
WHERE appName = 'mcp-server-production' 
SINCE 30 minutes ago

-- Tool usage
SELECT count(*) 
FROM Transaction 
WHERE appName = 'mcp-server-production' 
FACET request.uri 
WHERE request.uri LIKE '/tools/%' 
SINCE 1 hour ago

-- Cache hit rate
SELECT percentage(count(*), WHERE custom.cache_hit = 'true') 
FROM Transaction 
WHERE appName = 'mcp-server-production' 
SINCE 30 minutes ago
```

### 3. Alerting Configuration

```yaml
# New Relic alert policy
- name: "MCP Server Production Alerts"
  conditions:
    - name: "High Error Rate"
      type: "apm_app_metric"
      metric: "error_percentage"
      threshold: 5
      duration: 5
      
    - name: "High Response Time"
      type: "apm_app_metric"
      metric: "response_time"
      threshold: 1000  # milliseconds
      duration: 5
      
    - name: "Low Throughput"
      type: "apm_app_metric"
      metric: "throughput"
      threshold: 10  # rpm
      duration: 10
      operator: "below"
```

### 4. Prometheus Metrics

```yaml
# Prometheus scrape config
scrape_configs:
  - job_name: 'mcp-server'
    kubernetes_sd_configs:
    - role: pod
      namespaces:
        names:
        - mcp-server
    relabel_configs:
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
      action: keep
      regex: true
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
      action: replace
      target_label: __metrics_path__
      regex: (.+)
```

### 5. Logging Strategy

```yaml
# Fluent Bit configuration for log forwarding
[SERVICE]
    Flush         5
    Daemon        Off
    Log_Level     info

[INPUT]
    Name              tail
    Path              /var/log/containers/*mcp-server*.log
    Parser            docker
    Tag               mcp.server.*
    Refresh_Interval  5

[FILTER]
    Name         parser
    Match        mcp.server.*
    Key_Name     log
    Parser       json

[OUTPUT]
    Name        newrelic
    Match       mcp.server.*
    licenseKey  ${NEW_RELIC_LICENSE_KEY}
    endpoint    https://log-api.newrelic.com/log/v1
```

## Security Hardening

### 1. Network Policies

```yaml
# k8s/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mcp-server-network-policy
  namespace: mcp-server
spec:
  podSelector:
    matchLabels:
      app: mcp-server
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    - podSelector:
        matchLabels:
          app: prometheus
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 9090
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
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53
  - ports:
    - protocol: TCP
      port: 443  # New Relic API
```

### 2. Pod Security Policy

```yaml
# k8s/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: mcp-server-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
  - ALL
  volumes:
  - 'configMap'
  - 'emptyDir'
  - 'projected'
  - 'secret'
  - 'downwardAPI'
  - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true
```

### 3. Security Scanning

```bash
#!/bin/bash
# security-scan.sh

# Scan Docker image
trivy image --severity HIGH,CRITICAL your-registry/mcp-server-newrelic:latest

# Scan Kubernetes manifests
kubesec scan k8s/*.yaml

# Check for secrets in code
gitleaks detect --source . --verbose

# OWASP dependency check
dependency-check --project "MCP Server" --scan . --format ALL
```

### 4. Runtime Security

```yaml
# Falco rules for runtime security
- rule: Unexpected Network Connection from MCP Server
  desc: Detect unexpected network connections
  condition: >
    container.id != host and 
    container.name = "mcp-server" and 
    fd.name != "127.0.0.1" and
    not fd.snet in (allowed_networks)
  output: >
    Unexpected network connection from MCP Server 
    (command=%proc.cmdline connection=%fd.name)
  priority: WARNING
```

## Backup and Disaster Recovery

### 1. Backup Strategy

```yaml
# k8s/backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: mcp-backup
  namespace: mcp-server
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: your-registry/mcp-backup:latest
            env:
            - name: BACKUP_DESTINATION
              value: "s3://backups/mcp-server"
            - name: REDIS_URL
              value: "redis://redis-service:6379"
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mcp-server-secrets
                  key: REDIS_PASSWORD
            command:
            - /bin/sh
            - -c
            - |
              # Backup Redis data
              redis-cli -h redis-service --rdb /tmp/redis-backup.rdb
              
              # Backup configurations
              kubectl get configmap -n mcp-server -o yaml > /tmp/configmaps.yaml
              
              # Upload to S3
              aws s3 cp /tmp/redis-backup.rdb ${BACKUP_DESTINATION}/$(date +%Y%m%d)/
              aws s3 cp /tmp/configmaps.yaml ${BACKUP_DESTINATION}/$(date +%Y%m%d)/
              
              # Cleanup old backups (keep 30 days)
              aws s3 ls ${BACKUP_DESTINATION}/ | while read -r line; do
                createDate=$(echo $line | awk '{print $1" "$2}')
                createDate=$(date -d "$createDate" +%s)
                olderThan=$(date -d "30 days ago" +%s)
                if [[ $createDate -lt $olderThan ]]; then
                  fileName=$(echo $line | awk '{print $4}')
                  aws s3 rm ${BACKUP_DESTINATION}/$fileName
                fi
              done
          restartPolicy: OnFailure
```

### 2. Disaster Recovery Runbook

```markdown
# Disaster Recovery Runbook

## Scenario 1: Complete Service Failure

1. **Assess the situation**
   ```bash
   kubectl get pods -n mcp-server
   kubectl describe pods -n mcp-server
   kubectl logs -n mcp-server -l app=mcp-server --tail=100
   ```

2. **Restore from backup**
   ```bash
   # Scale down current deployment
   kubectl scale deployment mcp-server -n mcp-server --replicas=0
   
   # Restore Redis data
   kubectl exec -it redis-0 -n mcp-server -- redis-cli --rdb /data/dump.rdb
   
   # Restore configurations
   kubectl apply -f backups/latest/configmaps.yaml
   
   # Scale up deployment
   kubectl scale deployment mcp-server -n mcp-server --replicas=3
   ```

3. **Verify restoration**
   ```bash
   # Check health
   kubectl exec -it deployment/mcp-server -n mcp-server -- curl localhost:8080/health
   
   # Test functionality
   ./scripts/smoke-test.sh
   ```

## Scenario 2: Data Corruption

1. **Identify corruption**
   ```bash
   # Check Redis integrity
   kubectl exec -it redis-0 -n mcp-server -- redis-check-rdb /data/dump.rdb
   ```

2. **Restore clean data**
   ```bash
   # Stop writes
   kubectl patch deployment mcp-server -n mcp-server -p '{"spec":{"replicas":0}}'
   
   # Restore from last known good backup
   ./scripts/restore-redis.sh --date=$(date -d "1 day ago" +%Y%m%d)
   
   # Resume service
   kubectl patch deployment mcp-server -n mcp-server -p '{"spec":{"replicas":3}}'
   ```

## Scenario 3: Regional Failure

1. **Failover to DR region**
   ```bash
   # Update DNS to point to DR region
   ./scripts/update-dns.sh --region=dr
   
   # Ensure DR region is synchronized
   ./scripts/sync-dr-region.sh
   
   # Activate DR deployment
   kubectl apply -f k8s/dr-region/ --context=dr-cluster
   ```
```

### 3. Backup Verification

```bash
#!/bin/bash
# verify-backup.sh

set -e

BACKUP_DATE=${1:-$(date +%Y%m%d)}
BACKUP_PATH="s3://backups/mcp-server/${BACKUP_DATE}"

echo "Verifying backup for date: ${BACKUP_DATE}"

# Download backup files
aws s3 cp ${BACKUP_PATH}/redis-backup.rdb /tmp/
aws s3 cp ${BACKUP_PATH}/configmaps.yaml /tmp/

# Verify Redis backup
redis-check-rdb /tmp/redis-backup.rdb
if [ $? -eq 0 ]; then
    echo "✓ Redis backup is valid"
else
    echo "✗ Redis backup is corrupted"
    exit 1
fi

# Verify ConfigMap YAML
kubectl apply --dry-run=client -f /tmp/configmaps.yaml
if [ $? -eq 0 ]; then
    echo "✓ ConfigMap backup is valid"
else
    echo "✗ ConfigMap backup is invalid"
    exit 1
fi

echo "Backup verification completed successfully"
```

## Performance Tuning

### 1. Application Tuning

```go
// Performance configuration in code
type PerformanceConfig struct {
    // Connection pooling
    MaxIdleConns        int           `default:"100"`
    MaxOpenConns        int           `default:"200"`
    ConnMaxLifetime     time.Duration `default:"5m"`
    
    // Request handling
    ReadTimeout         time.Duration `default:"30s"`
    WriteTimeout        time.Duration `default:"30s"`
    IdleTimeout         time.Duration `default:"120s"`
    
    // Concurrency limits
    MaxConcurrentQueries int          `default:"50"`
    QueryQueueSize      int           `default:"1000"`
    
    // Caching
    CacheSize           int           `default:"10000"`
    CacheTTL            time.Duration `default:"5m"`
}
```

### 2. Database Query Optimization

```sql
-- Create indexes for common queries
CREATE INDEX idx_discovery_timestamp ON discovery_cache(created_at);
CREATE INDEX idx_discovery_schema ON discovery_cache(schema_name);
CREATE INDEX idx_query_history_user ON query_history(user_id, executed_at);

-- Optimize slow queries
EXPLAIN ANALYZE
SELECT * FROM discovery_cache 
WHERE schema_name = 'Transaction' 
AND created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;
```

### 3. Resource Optimization

```yaml
# JVM-style tuning for Go runtime
ENV GOGC=100
ENV GOMEMLIMIT=3500MiB
ENV GOMAXPROCS=4
```

### 4. Load Testing

```bash
#!/bin/bash
# load-test.sh

# Install k6 if not present
which k6 || (curl -s https://dl.k6.io/key.gpg | apt-key add - && \
  echo "deb https://dl.k6.io/deb stable main" | tee -a /etc/apt/sources.list && \
  apt-get update && apt-get install k6)

# Run load test
k6 run --vus 100 --duration 30m scripts/load-test.js

# Generate report
k6 run --out influxdb=http://localhost:8086/k6 scripts/load-test.js
```

```javascript
// scripts/load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '5m', target: 100 },  // Ramp up
    { duration: '20m', target: 100 }, // Stay at 100 users
    { duration: '5m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.1'],    // Error rate under 10%
  },
};

export default function () {
  // Test NRQL query
  let query = {
    method: 'POST',
    url: 'https://mcp.example.com/api/v1/tools/query_nrdb',
    body: JSON.stringify({
      query: 'SELECT count(*) FROM Transaction SINCE 1 hour ago',
      timeout: 30,
    }),
    params: {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + __ENV.API_TOKEN,
      },
    },
  };
  
  let res = http.post(query.url, query.body, query.params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'response has data': (r) => JSON.parse(r.body).data !== null,
  });
  
  sleep(1);
}
```

## Troubleshooting Deployment Issues

### 1. Common Issues and Solutions

```bash
#!/bin/bash
# troubleshoot.sh

echo "MCP Server Deployment Troubleshooting"
echo "===================================="

# Check pod status
echo -e "\n1. Checking pod status..."
kubectl get pods -n mcp-server -o wide

# Check recent events
echo -e "\n2. Recent events..."
kubectl get events -n mcp-server --sort-by='.lastTimestamp' | tail -20

# Check resource usage
echo -e "\n3. Resource usage..."
kubectl top pods -n mcp-server

# Check logs for errors
echo -e "\n4. Recent errors in logs..."
kubectl logs -n mcp-server -l app=mcp-server --tail=50 | grep -E "ERROR|FATAL|panic"

# Check service endpoints
echo -e "\n5. Service endpoints..."
kubectl get endpoints -n mcp-server

# Check ingress status
echo -e "\n6. Ingress status..."
kubectl describe ingress -n mcp-server

# Check certificate status
echo -e "\n7. TLS certificate status..."
kubectl get certificate -n mcp-server

# Test connectivity
echo -e "\n8. Testing connectivity..."
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -n mcp-server -- \
  curl -s -o /dev/null -w "%{http_code}" http://mcp-server-service:8080/health
```

### 2. Debug Mode Deployment

```yaml
# k8s/debug-deployment.yaml
apiVersion: v1
kind: Pod
metadata:
  name: mcp-server-debug
  namespace: mcp-server
spec:
  containers:
  - name: mcp-server
    image: your-registry/mcp-server-newrelic:debug
    env:
    - name: LOG_LEVEL
      value: "DEBUG"
    - name: VERBOSE_LOGGING
      value: "true"
    - name: ENABLE_PROFILING
      value: "true"
    - name: PPROF_PORT
      value: "6060"
    ports:
    - containerPort: 6060
      name: pprof
    command: ["/bin/sh"]
    args: ["-c", "dlv exec /mcp-server --headless --listen=:2345 --api-version=2 --accept-multiclient"]
```

### 3. Performance Profiling

```bash
# Enable profiling
kubectl port-forward -n mcp-server deployment/mcp-server 6060:6060 &

# CPU profile
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/goroutine

# Trace
wget http://localhost:6060/debug/pprof/trace?seconds=5
go tool trace trace
```

## Rollback Procedures

### 1. Automated Rollback

```yaml
# k8s/rollback-policy.yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: mcp-server
  namespace: mcp-server
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: mcp-server
  progressDeadlineSeconds: 600
  service:
    port: 8080
    targetPort: 8080
  analysis:
    interval: 1m
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
      interval: 1m
    - name: request-duration
      thresholdRange:
        max: 500
      interval: 30s
    webhooks:
    - name: smoke-test
      url: http://flagger-loadtester.mcp-server/
      timeout: 5m
      metadata:
        type: smoke
        cmd: "./scripts/smoke-test.sh"
```

### 2. Manual Rollback Process

```bash
#!/bin/bash
# rollback.sh

DEPLOYMENT="mcp-server"
NAMESPACE="mcp-server"

echo "Starting rollback process..."

# Get rollout history
echo "Rollout history:"
kubectl rollout history deployment/$DEPLOYMENT -n $NAMESPACE

# Get current revision
CURRENT=$(kubectl get deployment/$DEPLOYMENT -n $NAMESPACE -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
echo "Current revision: $CURRENT"

# Perform rollback
read -p "Enter revision to rollback to (or press Enter for previous): " REVISION
if [ -z "$REVISION" ]; then
    kubectl rollout undo deployment/$DEPLOYMENT -n $NAMESPACE
else
    kubectl rollout undo deployment/$DEPLOYMENT -n $NAMESPACE --to-revision=$REVISION
fi

# Monitor rollback
echo "Monitoring rollback..."
kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE

# Verify health after rollback
sleep 10
./scripts/health-check.sh

echo "Rollback completed"
```

### 3. Database Rollback

```sql
-- Create restore point before deployment
BEGIN;
SAVEPOINT pre_deployment;

-- If issues occur, rollback
ROLLBACK TO SAVEPOINT pre_deployment;

-- Or use time-based recovery
-- Restore Redis to specific point in time
BGREWRITEAOF
CONFIG SET appendonly yes
CONFIG SET appendfsync everysec
```

### 4. Emergency Procedures

```bash
#!/bin/bash
# emergency-shutdown.sh

echo "EMERGENCY SHUTDOWN INITIATED"

# Stop all traffic
kubectl patch service mcp-server-service -n mcp-server -p '{"spec":{"selector":null}}'

# Scale down deployment
kubectl scale deployment mcp-server -n mcp-server --replicas=0

# Clear Redis cache (if needed)
kubectl exec -it redis-0 -n mcp-server -- redis-cli FLUSHALL

# Notify team
curl -X POST $SLACK_WEBHOOK -H 'Content-type: application/json' \
  --data '{"text":"EMERGENCY: MCP Server shutdown initiated"}'

echo "Emergency shutdown completed"
```

## Post-Deployment Checklist

- [ ] All pods are running and healthy
- [ ] Health checks are passing
- [ ] Metrics are being collected
- [ ] Logs are being aggregated
- [ ] Alerts are configured and tested
- [ ] Backup job is scheduled and verified
- [ ] Performance baselines established
- [ ] Security scans completed
- [ ] Documentation updated
- [ ] Team trained on runbooks

## Maintenance Schedule

### Daily
- Monitor error rates and performance metrics
- Review security alerts
- Check backup completion

### Weekly
- Review and optimize slow queries
- Update security patches
- Performance trend analysis

### Monthly
- Disaster recovery drill
- Security audit
- Capacity planning review
- Cost optimization review

### Quarterly
- Major version updates
- Architecture review
- Runbook updates
- Team training

---

This deployment guide should be treated as a living document and updated based on operational experience and changing requirements.
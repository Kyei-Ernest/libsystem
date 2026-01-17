# LibSystem Deployment Guide

## Prerequisites

### System Requirements
- **OS**: Ubuntu 20.04+ / Debian 11+
- **RAM**: Minimum 8GB (16GB recommended)
- **CPU**: 4 cores minimum
- **Disk**: 50GB+ available space

### Required Software
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 7+
- Elasticsearch 8.x
- Kafka 3.x
- MinIO
- Tesseract OCR 5.x

---

## Quick Start (Development)

### 1. Clone Repository
```bash
git clone https://github.com/your-org/libsystem.git
cd libsystem
```

### 2. Start Infrastructure Services
```bash
docker-compose up -d postgres redis elasticsearch kafka minio
```

### 3. Install Tesseract OCR
```bash
sudo apt-get update
sudo apt-get install -y tesseract-ocr tesseract-ocr-eng
```

### 4. Set Environment Variables

Create `.env` files for each service or export globally:

```bash
# Shared
export JWT_SECRET="your-secret-key-change-in-production"
export SERVICE_SECRET="internal-secret-key"

# API Gateway
export PORT=8088
export REDIS_ADDR="localhost:6379"
export ENVIRONMENT="development"

# User Service
export PORT=8086
export DB_HOST="localhost"
export DB_PORT=5432
export DB_NAME="libsystem_users"
export DB_USER="postgres"
export DB_PASSWORD="postgres"

# Document Service
export PORT=8081
export KAFKA_BROKERS="localhost:9092"
export MINIO_ENDPOINT="localhost:9000"
export MINIO_ACCESS_KEY="minioadmin"
export MINIO_SECRET_KEY="minioadmin123"

# Search Service
export PORT=8084
export ELASTICSEARCH_URL="http://localhost:9200"

# Collection Service
export PORT=8082

# Indexer Service
export KAFKA_TOPIC="document.uploaded"
```

### 5. Initialize Databases
```bash
# Run migrations for each service
cd services/user-service && go run migrations/*.go
cd ../document-service && go run migrations/*.go
cd ../collection-service && go run migrations/*.go
```

### 6. Start Services
```bash
# Run all services (use tmux or separate terminals)
./run-services.sh

# Or manually:
cd services/api-gateway && go run main.go &
cd services/user-service && go run main.go &
cd services/document-service && go run main.go &
cd services/collection-service && go run main.go &
cd services/search-service && go run main.go &
cd services/indexer-service && go run main.go &
```

### 7. Verify Installation
```bash
# Check health
curl http://localhost:8088/health

# Register a user
curl -X POST http://localhost:8088/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!",
    "username": "testuser",
    "full_name": "Test User"
  }'
```

---

## Production Deployment

### 1. Build Binaries
```bash
# Build all services
for service in api-gateway user-service document-service collection-service search-service indexer-service; do
  cd services/$service
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $service .
  cd ../..
done
```

### 2. Docker Images
```bash
# Build Docker images
docker build -t libsystem/api-gateway:v1.0 services/api-gateway
docker build -t libsystem/user-service:v1.0 services/user-service
# ... repeat for all services
```

### 3. Kubernetes Deployment

**Example deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: libsystem/api-gateway:v1.0
        ports:
        - containerPort: 8088
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: libsystem-secrets
              key: jwt-secret
        - name: REDIS_ADDR
          value: "redis-service:6379"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8088
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8088
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 4. Load Balancing

**Nginx Configuration:**
```nginx
upstream api_gateway {
    least_conn;
    server 10.0.1.10:8088;
    server 10.0.1.11:8088;
    server 10.0.1.12:8088;
}

server {
    listen 80;
    server_name api.libsystem.com;

    location / {
        proxy_pass http://api_gateway;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Request-ID $request_id;
    }

    # Rate limiting at nginx level
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req zone=api_limit burst=20 nodelay;
}
```

---

## Monitoring Setup

### Prometheus Configuration
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'libsystem'
    static_configs:
      - targets:
        - 'api-gateway:8088'
        - 'user-service:8086'
        - 'document-service:8081'
        - 'collection-service:8082'
        - 'search-service:8084'
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana Dashboards
Import the provided dashboard JSON files:
- `dashboards/api-gateway.json`
- `dashboards/system-overview.json`
- `dashboards/document-processing.json`

---

## Troubleshooting

### Service Won't Start

**Check logs:**
```bash
journalctl -u libsystem-api-gateway -f
```

**Common issues:**
- Port already in use: `lsof -i:8088`
- Database connection: Check `DB_HOST` and credentials
- Redis unavailable: Verify `REDIS_ADDR`

### High Memory Usage

```bash
# Check service memory
ps aux | grep libsystem

# Restart specific service
systemctl restart libsystem-api-gateway
```

### Rate Limiting Issues

```bash
# Check Redis keys
redis-cli
> KEYS ratelimit:*
> TTL ratelimit:user:uuid

# Clear rate limits (development only)
> FLUSHDB
```

---

## Security Checklist

- [ ] Change default `JWT_SECRET` and `SERVICE_SECRET`
- [ ] Use strong database passwords
- [ ] Enable TLS/SSL for all services
- [ ] Configure firewall rules
- [ ] Rotate secrets regularly
- [ ] Enable audit logging
- [ ] Set up intrusion detection
- [ ] Regular security updates
- [ ] Backup encryption keys

---

## Backup & Recovery

### Database Backup
```bash
# Automated daily backup
pg_dump -h localhost -U postgres libsystem_users > backup_$(date +%Y%m%d).sql

# Restore
psql -h localhost -U postgres libsystem_users < backup_20260110.sql
```

### MinIO Backup
```bash
# Sync to S3
mc mirror minio/documents s3/backup-bucket/documents
```

### Elasticsearch Snapshot
```bash
# Create snapshot repository
curl -X PUT "localhost:9200/_snapshot/backup" \
  -H 'Content-Type: application/json' \
  -d '{"type": "fs", "settings": {"location": "/backup/elasticsearch"}}'

# Create snapshot
curl -X PUT "localhost:9200/_snapshot/backup/snapshot_1"
```

---

## Performance Tuning

### API Gateway
- Increase worker pool size
- Enable HTTP/2
- Configure connection pooling

### Database
```sql
-- Optimize queries
CREATE INDEX idx_documents_uploader ON documents(uploader_id);
CREATE INDEX idx_documents_collection ON documents(collection_id);
```

### Redis
```conf
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
```

### Elasticsearch
```yaml
# elasticsearch.yml
indices.memory.index_buffer_size: 30%
thread_pool.write.queue_size: 1000
```

---

## Scaling Guide

### Horizontal Scaling
- API Gateway: Scale to 3-5 instances behind load balancer
- Document Service: Scale based on upload volume
- Indexer Service: Multiple consumers on same topic

### Vertical Scaling
- Increase memory for Elasticsearch nodes
- Add CPU cores for document processing
- Expand MinIO storage capacity

---

## Support

- **Documentation**: `/docs`
- **Issues**: GitHub Issues
- **Email**: support@libsystem.com

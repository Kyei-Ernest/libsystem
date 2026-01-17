# Environment Setup Checklist

Use this checklist to ensure your development environment is properly configured.

## ✅ Prerequisites Check

### System Requirements
- [ ] **Operating System**: macOS, Linux, or Windows with WSL2
- [ ] **RAM**: Minimum 16GB (32GB recommended)
- [ ] **Disk Space**: Minimum 50GB free
- [ ] **CPU**: 4+ cores recommended

### Required Software
- [ ] **Go 1.21+** installed
  ```bash
  go version  # Should show: go version go1.21.x
  ```
- [ ] **Docker 20.10+** installed and running
  ```bash
  docker --version
  docker info  # Should not error
  ```
- [ ] **Docker Compose 2.0+** installed
  ```bash
  docker-compose --version
  ```
- [ ] **Git** installed
  ```bash
  git --version
  ```
- [ ] **Make** installed
  ```bash
  make --version
  ```
- [ ] **curl** installed
  ```bash
  curl --version
  ```
- [ ] **jq** installed (for JSON parsing)
  ```bash
  jq --version
  ```

## ✅ Project Setup

### Initial Setup
- [ ] Repository cloned
  ```bash
  git clone <repo-url>
  cd libsystem
  ```
- [ ] Setup script executed
  ```bash
  chmod +x setup.sh
  ./setup.sh
  ```
- [ ] Directory structure created
  ```
  ✓ services/
  ✓ shared/
  ✓ infrastructure/
  ✓ migrations/
  ✓ backups/
  ```

### Configuration Files
- [ ] `.env` file created from `.env.example`
- [ ] `.gitignore` in place
- [ ] `docker-compose.yml` present
- [ ] `Makefile` present
- [ ] All required files saved in correct locations

### Go Modules
- [ ] API Gateway module initialized
  ```bash
  cd services/api-gateway && ls go.mod
  ```
- [ ] Dependencies downloaded
  ```bash
  go mod download
  ```

## ✅ Docker Services

### Start Services
- [ ] All services started
  ```bash
  make up
  # or
  docker-compose up -d
  ```
- [ ] Check services status
  ```bash
  make ps
  # All services should show "Up"
  ```

### Individual Service Health
- [ ] **PostgreSQL** is healthy
  ```bash
  docker-compose exec postgres pg_isready -U libsystem
  # Should show: accepting connections
  ```
- [ ] **Redis** is healthy
  ```bash
  docker-compose exec redis redis-cli -a libsystem_redis_pass ping
  # Should return: PONG
  ```
- [ ] **Elasticsearch** is healthy
  ```bash
  curl http://localhost:9200/_cluster/health
  # Should show: status: "green" or "yellow"
  ```
- [ ] **MinIO** is healthy
  ```bash
  curl http://localhost:9000/minio/health/live
  # Should return: 200 OK
  ```
- [ ] **Kafka** is healthy
  ```bash
  docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
  # Should not error
  ```
- [ ] **API Gateway** is healthy
  ```bash
  curl http://localhost:8085/health
  # Should return: {"status":"healthy"}
  ```

## ✅ Service Initialization

### Database
- [ ] Migrations applied
  ```bash
  make migrate
  # Should complete without errors
  ```
- [ ] Can connect to database
  ```bash
  make db-shell
  # Should open psql prompt
  # Type \dt to see tables
  # Type \q to exit
  ```
- [ ] Initial data seeded
  ```bash
  make db-seed
  # Should create test collections
  ```
- [ ] Admin user exists
  ```sql
  SELECT username, email, role FROM users WHERE role='admin';
  # Should show admin user
  ```

### MinIO
- [ ] Buckets created
  ```bash
  make init-minio
  # Should create documents and thumbnails buckets
  ```
- [ ] Can access MinIO console
  - Open: http://localhost:9001
  - Login: minioadmin / minioadmin123
  - [ ] See `documents` bucket
  - [ ] See `thumbnails` bucket

### Kafka
- [ ] Topics created
  ```bash
  make init-kafka
  # Should create all required topics
  ```
- [ ] Verify topics exist
  ```bash
  docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
  # Should show:
  # - document.uploaded
  # - document.indexed
  # - document.deleted
  ```

## ✅ Service Accessibility

### Web Interfaces
- [ ] **API Gateway** accessible
  - URL: http://localhost:8085
  - Test: `curl http://localhost:8085/health`

- [ ] **Grafana** accessible
  - URL: http://localhost:3000
  - Credentials: admin / admin
  - [ ] Can login
  - [ ] Prometheus datasource configured

- [ ] **Kibana** accessible
  - URL: http://localhost:5601
  - [ ] Can access dashboard

- [ ] **MinIO Console** accessible
  - URL: http://localhost:9001
  - Credentials: minioadmin / minioadmin123
  - [ ] Can login
  - [ ] See buckets

- [ ] **Prometheus** accessible
  - URL: http://localhost:9090
  - [ ] Can see targets
  - [ ] Services being scraped

- [ ] **Jaeger** accessible
  - URL: http://localhost:16686
  - [ ] UI loads

## ✅ API Testing

### Health Checks
- [ ] API health endpoint works
  ```bash
  curl http://localhost:8085/health | jq
  ```
- [ ] Readiness check works
  ```bash
  curl http://localhost:8085/ready | jq
  ```
- [ ] Metrics endpoint works
  ```bash
  curl http://localhost:8085/metrics
  ```

### User Registration
- [ ] Can register a new user
  ```bash
  curl -X POST http://localhost:8085/api/v1/users/register \
    -H "Content-Type: application/json" \
    -d '{
      "email": "test@example.com",
      "username": "testuser",
      "password": "password123",
      "first_name": "Test",
      "last_name": "User"
    }'
  ```

### User Login
- [ ] Can login with created user
  ```bash
  curl -X POST http://localhost:8085/api/v1/users/login \
    -H "Content-Type: application/json" \
    -d '{
      "username": "testuser",
      "password": "password123"
    }'
  # Should return JWT token
  ```

## ✅ Development Workflow

### Code Development
- [ ] Can run API locally
  ```bash
  make dev-api
  # Should start and listen on :8085
  ```
- [ ] Hot reload works (if using air)
  ```bash
  make watch
  # Should restart on code changes
  ```

### Testing
- [ ] Can run unit tests
  ```bash
  make test
  # Should run all tests
  ```
- [ ] Tests pass
  ```bash
  make test-api
  # Should show: PASS
  ```
- [ ] Can generate coverage report
  ```bash
  make test-coverage
  # Should create coverage.html
  ```

### Logging
- [ ] Can view all logs
  ```bash
  make logs
  ```
- [ ] Can view specific service logs
  ```bash
  make logs-api
  make logs-doc
  ```

### Code Quality
- [ ] Linter installed (optional)
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- [ ] Can run linter
  ```bash
  make lint
  ```
- [ ] Code formatting works
  ```bash
  make format
  ```

## ✅ Monitoring Setup

### Prometheus
- [ ] Prometheus scraping services
  - Go to http://localhost:9090/targets
  - [ ] All targets should be "UP"

### Grafana
- [ ] Datasource connected
  - Login to Grafana
  - Go to Configuration > Data Sources
  - [ ] Prometheus should be listed and working

- [ ] Can create dashboards
  - [ ] Create a test dashboard
  - [ ] Add a panel with a metric
  - [ ] Data loads correctly

### Logs
- [ ] Can query Elasticsearch (if using ELK)
  ```bash
  curl http://localhost:9200/_cat/indices
  ```

## ✅ Troubleshooting Checks

### If Services Won't Start
- [ ] Check Docker is running
  ```bash
  docker info
  ```
- [ ] Check disk space
  ```bash
  df -h
  ```
- [ ] Check port conflicts
  ```bash
  # Check if ports are available
  lsof -i :8085  # macOS/Linux
  netstat -ano | findstr :8085  # Windows
  ```
- [ ] View service logs
  ```bash
  docker-compose logs <service-name>
  ```

### If Database Won't Connect
- [ ] PostgreSQL is running
  ```bash
  docker-compose ps postgres
  ```
- [ ] Can connect to database
  ```bash
  docker-compose exec postgres psql -U libsystem -d libsystem
  ```
- [ ] Check database logs
  ```bash
  docker-compose logs postgres
  ```

### If Elasticsearch Won't Start
- [ ] Check virtual memory settings (Linux)
  ```bash
  sudo sysctl -w vm.max_map_count=262144
  ```
- [ ] Check Elasticsearch logs
  ```bash
  docker-compose logs elasticsearch
  ```

## ✅ Clean Up (Optional)

### Stop Services
- [ ] Services stopped gracefully
  ```bash
  make down
  ```

### Remove All Data (CAUTION!)
- [ ] All volumes removed
  ```bash
  make clean
  # This deletes all data!
  ```

## ✅ Optional Tools

### Development Tools
- [ ] Air (hot reload) installed
  ```bash
  go install github.com/cosmtrek/air@latest
  ```
- [ ] Swag (API docs) installed
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```
- [ ] golangci-lint installed
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

### IDE Setup
- [ ] VS Code configured (if using)
  - [ ] Go extension installed
  - [ ] Docker extension installed
  - [ ] Settings configured

## 📝 Notes

### Important URLs
```
API Gateway:    http://localhost:8085
Grafana:        http://localhost:3000 (admin/admin)
Kibana:         http://localhost:5601
MinIO:          http://localhost:9001 (minioadmin/minioadmin123)
Prometheus:     http://localhost:9090
Jaeger:         http://localhost:16686
PostgreSQL:     localhost:5432 (libsystem/libsystem_dev_pass)
Redis:          localhost:6380 (password: libsystem_redis_pass)
```

### Default Credentials
```
PostgreSQL:     libsystem / libsystem_dev_pass
Redis:          password: libsystem_redis_pass
MinIO:          minioadmin / minioadmin123
Grafana:        admin / admin
Admin User:     admin@libsystem.com / admin123
```

### Useful Commands
```bash
make help       # Show all commands
make ps         # Service status
make logs       # View logs
make health     # Health check
make db-shell   # Open database shell
make redis-cli  # Open Redis CLI
```

## ✅ Final Verification

- [ ] All prerequisites installed
- [ ] All services running
- [ ] All health checks passing
- [ ] Can make API requests
- [ ] Monitoring dashboards accessible
- [ ] Documentation reviewed

---

**Setup Status**: 
- [ ] ✅ Complete - Ready for development!
- [ ] ⚠️ Partial - Some issues to resolve
- [ ] ❌ Failed - Need help

**Date Completed**: ________________

**Notes**: 
_____________________________________
_____________________________________
_____________________________________
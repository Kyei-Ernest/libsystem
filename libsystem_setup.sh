#!/bin/bash

# ==========================================
# Library Management System - Setup Script
# ==========================================

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 is not installed"
        return 1
    else
        print_success "$1 is installed"
        return 0
    fi
}

# Start setup
clear
print_header "Library Management System - Setup"
echo ""

# Check prerequisites
print_header "Checking Prerequisites"
echo ""

ALL_GOOD=true

if check_command "go"; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_info "Go version: $GO_VERSION"
else
    print_error "Please install Go 1.21 or higher from https://golang.org/dl/"
    ALL_GOOD=false
fi

if check_command "docker"; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | tr -d ',')
    print_info "Docker version: $DOCKER_VERSION"
else
    print_error "Please install Docker from https://docs.docker.com/get-docker/"
    ALL_GOOD=false
fi

if check_command "docker-compose"; then
    COMPOSE_VERSION=$(docker-compose --version | awk '{print $3}' | tr -d ',')
    print_info "Docker Compose version: $COMPOSE_VERSION"
else
    print_error "Please install Docker Compose from https://docs.docker.com/compose/install/"
    ALL_GOOD=false
fi

check_command "git"
check_command "make"
check_command "curl"
check_command "jq"

echo ""

if [ "$ALL_GOOD" = false ]; then
    print_error "Please install missing prerequisites and run this script again"
    exit 1
fi

print_success "All prerequisites are installed!"
echo ""

# Create directory structure
print_header "Creating Directory Structure"
echo ""

directories=(
    "services/api-gateway"
    "services/document-service"
    "services/search-service"
    "services/collection-service"
    "services/user-service"
    "services/indexer-service"
    "shared/proto"
    "shared/models"
    "shared/config"
    "shared/utils"
    "infrastructure/kubernetes"
    "infrastructure/terraform"
    "infrastructure/docker"
    "infrastructure/prometheus"
    "infrastructure/grafana/dashboards"
    "infrastructure/grafana/datasources"
    "migrations"
    "docs"
    "scripts"
    "backups"
    "logs"
    "tests/integration"
    "tests/e2e"
)

for dir in "${directories[@]}"; do
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        print_success "Created $dir"
    else
        print_info "$dir already exists"
    fi
done

# Create .gitkeep files
touch backups/.gitkeep
touch logs/.gitkeep

echo ""
print_success "Directory structure created!"
echo ""

# Create environment file
print_header "Setting Up Environment Variables"
echo ""

if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        cp .env.example .env
        print_success "Created .env from .env.example"
        print_info "Please review and update .env file with your settings"
    else
        print_error ".env.example not found"
    fi
else
    print_info ".env already exists"
fi

echo ""

# Initialize Go modules
print_header "Initializing Go Modules"
echo ""

cd services/api-gateway
if [ ! -f "go.mod" ]; then
    go mod init github.com/yourusername/libsystem/services/api-gateway
    print_success "Initialized api-gateway module"
fi
print_info "Downloading api-gateway dependencies..."
go mod download
go mod tidy
cd ../..

for service in document-service search-service collection-service user-service indexer-service; do
    cd services/$service
    if [ ! -f "go.mod" ]; then
        go mod init github.com/yourusername/libsystem/services/$service
        print_success "Initialized $service module"
    fi
    cd ../..
done

echo ""
print_success "Go modules initialized!"
echo ""

# Create Prometheus configuration
print_header "Creating Monitoring Configuration"
echo ""

cat > infrastructure/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'api-gateway'
    static_configs:
      - targets: ['api-gateway:8081']
        labels:
          service: 'api-gateway'
  
  - job_name: 'document-service'
    static_configs:
      - targets: ['document-service:8082']
        labels:
          service: 'document-service'
  
  - job_name: 'indexer-service'
    static_configs:
      - targets: ['indexer-service:8083']
        labels:
          service: 'indexer-service'
  
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
        labels:
          service: 'postgres'
  
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
        labels:
          service: 'redis'
EOF

print_success "Created Prometheus configuration"

# Create Grafana datasource
cat > infrastructure/grafana/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF

print_success "Created Grafana datasource configuration"

echo ""

# Create basic Dockerfile for API Gateway
print_header "Creating Docker Configuration"
echo ""

cat > services/api-gateway/Dockerfile << 'EOF'
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

EXPOSE 8085

CMD ["./main"]
EOF

print_success "Created Dockerfile for API Gateway"

echo ""

# Create .dockerignore
cat > .dockerignore << 'EOF'
.git
.gitignore
README.md
.env
.env.*
*.md
docker-compose*.yml
Makefile
.vscode
.idea
*.log
tmp/
logs/
backups/
coverage.html
*.test
EOF

print_success "Created .dockerignore"

echo ""

# Create README
print_header "Creating Documentation"
echo ""

cat > README.md << 'EOF'
# Library Management System

A production-ready, scalable library management system inspired by Greenstone, built with Go and modern cloud-native technologies.

## Features

- 📚 Digital document management
- 🔍 Full-text search with Elasticsearch
- 📁 Collection organization
- 👥 Multi-tenant support
- 🔐 JWT authentication
- 📊 Analytics and reporting
- 🚀 Horizontally scalable
- 📈 Built for 2M+ users

## Tech Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Search**: Elasticsearch 8.11
- **Storage**: MinIO (S3-compatible)
- **Message Queue**: Apache Kafka
- **Monitoring**: Prometheus + Grafana
- **Tracing**: Jaeger
- **Container**: Docker + Kubernetes

## Prerequisites

- Go 1.21 or higher
- Docker 20.10+
- Docker Compose 2.0+
- Make
- Git

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd libsystem
   ```

2. **Run setup script**
   ```bash
   chmod +x setup.sh
   ./setup.sh
   ```

3. **Start the development environment**
   ```bash
   make init
   ```

4. **Verify services are running**
   ```bash
   make health
   ```

5. **Access the application**
   - API Gateway: http://localhost:8085
   - Grafana: http://localhost:3000 (admin/admin)
   - Kibana: http://localhost:5601
   - MinIO Console: http://localhost:9001 (minioadmin/minioadmin123)
   - Jaeger UI: http://localhost:16686

## Development

### Available Commands

```bash
make help              # Show all available commands
make up                # Start all services
make down              # Stop all services
make logs              # View logs
make test              # Run tests
make lint              # Run linters
make migrate           # Run database migrations
make db-seed           # Seed database with test data
```

### Running Services Locally

```bash
# Start infrastructure services
make up

# Run API gateway locally (for development)
make dev-api
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific service tests
make test-api
```

## Project Structure

```
libsystem/
├── services/
│   ├── api-gateway/           # Main API entry point
│   ├── document-service/      # Document processing
│   ├── search-service/        # Search operations
│   ├── collection-service/    # Collection management
│   ├── user-service/          # User authentication
│   └── indexer-service/       # Background indexing
├── shared/
│   ├── proto/                 # Protocol buffers
│   ├── models/                # Shared data models
│   └── config/                # Shared configuration
├── infrastructure/
│   ├── kubernetes/            # K8s manifests
│   ├── terraform/             # Infrastructure as Code
│   └── docker/                # Docker configs
├── migrations/                # Database migrations
├── docs/                      # Documentation
└── scripts/                   # Utility scripts
```

## API Documentation

Once the services are running, access the API documentation at:
- Swagger UI: http://localhost:8085/swagger/index.html

## Monitoring

### Metrics
Access Grafana at http://localhost:3000 (default credentials: admin/admin)

### Logs
View logs using:
```bash
make logs              # All services
make logs-api          # API Gateway only
make logs-doc          # Document service only
```

### Tracing
Access Jaeger UI at http://localhost:16686

## Database

### Migrations

```bash
# Run migrations
make migrate

# Create new migration
make migrate-create NAME=add_new_table

# Rollback last migration
make migrate-down
```

### Backup & Restore

```bash
# Backup database
make backup-db

# Restore database
make restore-db FILE=backups/libsystem_20240101.sql
```

## Deployment

### Docker Compose (Development)
```bash
make up
```

### Kubernetes (Production)
```bash
make k8s-deploy
```

## Configuration

Configuration is managed through environment variables. Copy `.env.example` to `.env` and adjust values:

```bash
cp .env.example .env
```

Key configuration options:
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `ELASTICSEARCH_URL`: Elasticsearch endpoint
- `JWT_SECRET`: Secret key for JWT tokens
- `MINIO_ENDPOINT`: Object storage endpoint

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions, please open an issue on GitHub or contact the team.
EOF

print_success "Created README.md"

echo ""

# Final summary
print_header "Setup Complete!"
echo ""

print_success "Environment setup completed successfully!"
echo ""
print_info "Next steps:"
echo "  1. Review and update .env file with your configuration"
echo "  2. Run 'make init' to initialize the development environment"
echo "  3. Run 'make health' to verify all services are running"
echo "  4. Visit http://localhost:8085 to access the API"
echo ""
print_info "Useful commands:"
echo "  make help     - Show all available commands"
echo "  make up       - Start all services"
echo "  make logs     - View logs"
echo "  make test     - Run tests"
echo ""
print_success "Happy coding! 🚀"

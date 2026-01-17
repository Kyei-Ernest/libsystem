# Installation Guide - Library Management System

This guide will walk you through setting up the development environment from scratch.

## Prerequisites Installation

### 1. Install Go (1.21+)

#### macOS
```bash
brew install go
```

#### Linux (Ubuntu/Debian)
```bash
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Windows
Download and install from: https://go.dev/dl/

**Verify installation:**
```bash
go version
# Should output: go version go1.21.x...
```

### 2. Install Docker & Docker Compose

#### macOS
```bash
# Install Docker Desktop
brew install --cask docker
```

#### Linux (Ubuntu/Debian)
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### Windows
Download Docker Desktop from: https://www.docker.com/products/docker-desktop

**Verify installation:**
```bash
docker --version
docker-compose --version
```

### 3. Install Additional Tools

#### macOS
```bash
brew install make git curl jq
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install -y make git curl jq
```

#### Windows
- Git: https://git-scm.com/download/win
- Make: http://gnuwin32.sourceforge.net/packages/make.htm
- jq: https://stedolan.github.io/jq/download/

---

## Project Setup

### Step 1: Clone the Repository

```bash
# Create workspace directory
mkdir -p ~/workspace
cd ~/workspace

# Clone the repository (replace with your actual repo URL)
git clone https://github.com/yourusername/libsystem.git
cd libsystem
```

### Step 2: Run Automated Setup

```bash
# Make setup script executable
chmod +x setup.sh

# Run setup script
./setup.sh
```

The setup script will:
- ✅ Check all prerequisites
- ✅ Create directory structure
- ✅ Initialize Go modules
- ✅ Create configuration files
- ✅ Set up Docker configurations

### Step 3: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env file with your settings
nano .env  # or use your preferred editor
```

**Important variables to review:**
```env
# Change these in production
JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
DB_PASSWORD=your-secure-password
REDIS_PASSWORD=your-redis-password
MINIO_SECRET_KEY=your-minio-secret

# Optional: Configure external services
SMTP_HOST=smtp.gmail.com
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### Step 4: Create Required Files

Save the following files that were provided:

1. **Main API Gateway** → `services/api-gateway/main.go`
2. **Database Models** → `shared/models/models.go`
3. **Initial Migration** → `migrations/001_initial_schema.up.sql`
4. **Docker Compose** → `docker-compose.yml`
5. **Makefile** → `Makefile`
6. **Go Dependencies** → `services/api-gateway/go.mod`

### Step 5: Initialize Development Environment

```bash
# This command will:
# - Start all Docker services
# - Initialize MinIO buckets
# - Create Kafka topics
# - Run database migrations
# - Seed initial data
make init
```

**Expected output:**
```
Starting services...
✓ Services started!
Initializing MinIO buckets...
✓ MinIO buckets initialized!
Initializing Kafka topics...
✓ Kafka topics initialized!
Running migrations...
✓ Migrations complete!
Seeding database...
✓ Database seeded!
✓ Development environment ready!
```

---

## Verification

### Step 6: Verify Services

```bash
# Check service status
make ps

# Expected output:
# NAME                          STATUS    PORTS
# libsystem-api-gateway         Up        0.0.0.0:8080->8080/tcp
# libsystem-postgres            Up        0.0.0.0:5432->5432/tcp
# libsystem-redis               Up        0.0.0.0:6379->6379/tcp
# libsystem-elasticsearch       Up        0.0.0.0:9200->9200/tcp
# ...
```

### Step 7: Test Health Endpoints

```bash
# Check API Gateway health
curl http://localhost:8080/health | jq

# Expected response:
# {
#   "status": "healthy"
# }

# Check Elasticsearch
curl http://localhost:9200/_cluster/health | jq

# Check all services
make health
```

### Step 8: Access Service UIs

Open these URLs in your browser:

| Service | URL | Credentials |
|---------|-----|-------------|
| API Gateway | http://localhost:8080 | N/A |
| Grafana | http://localhost:3000 | admin / admin |
| Kibana | http://localhost:5601 | N/A |
| MinIO Console | http://localhost:9001 | minioadmin / minioadmin123 |
| Prometheus | http://localhost:9090 | N/A |
| Jaeger | http://localhost:16686 | N/A |

---

## Testing the Installation

### Step 9: Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run API tests only
make test-api
```

### Step 10: Make a Test API Request

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

---

## Troubleshooting

### Issue: Docker services won't start

**Solution:**
```bash
# Check Docker is running
docker info

# Remove old containers and volumes
make clean

# Rebuild and start
make build
make up
```

### Issue: Port already in use

**Solution:**
```bash
# Find process using port
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows

# Kill the process or change port in .env
```

### Issue: Database connection failed

**Solution:**
```bash
# Check PostgreSQL is healthy
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Connect to database manually
make db-shell
```

### Issue: Elasticsearch won't start

**Solution:**
```bash
# Increase virtual memory (Linux)
sudo sysctl -w vm.max_map_count=262144

# Make permanent
echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf

# Restart Elasticsearch
docker-compose restart elasticsearch
```

### Issue: Go module errors

**Solution:**
```bash
# Clean and reinstall
cd services/api-gateway
go clean -modcache
go mod download
go mod tidy
```

---

## Development Workflow

### Starting Development

```bash
# Start infrastructure services
make up

# Watch API logs
make logs-api

# In another terminal: run API locally with hot reload
make watch
```

### Stopping Services

```bash
# Stop all services
make down

# Stop and remove all data
make clean
```

### Database Operations

```bash
# Access database shell
make db-shell

# Run migrations
make migrate

# Create new migration
make migrate-create NAME=add_users_table

# Seed test data
make db-seed

# Backup database
make backup-db

# Restore database
make restore-db FILE=backups/backup.sql
```

### Viewing Logs

```bash
# All services
make logs

# Specific service
make logs-api
make logs-doc
make logs-indexer

# Follow logs in real-time
docker-compose logs -f api-gateway
```

---

## Next Steps

Now that your environment is set up:

1. ✅ **Explore the API** - Try the endpoints at http://localhost:8080
2. ✅ **Check monitoring** - View metrics in Grafana at http://localhost:3000
3. ✅ **Review code** - Understand the project structure
4. ✅ **Run tests** - Ensure everything works: `make test`
5. ✅ **Read documentation** - Check `/docs` folder for more info

## Getting Help

- 📚 **Documentation**: See `/docs` folder
- 🐛 **Issues**: Open an issue on GitHub
- 💬 **Discussions**: Use GitHub Discussions
- 📧 **Email**: contact@libsystem.com

---

## Optional: Install Development Tools

### Go Development Tools

```bash
# Install useful Go tools
go install github.com/cosmtrek/air@latest           # Hot reload
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linter
go install github.com/swaggo/swag/cmd/swag@latest  # API docs generator

# Add to PATH (if not already)
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

### VS Code Extensions (Recommended)

- Go (by Go Team at Google)
- Docker (by Microsoft)
- GitLens
- REST Client
- YAML

### Configure VS Code

Create `.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "editor.formatOnSave": true
}
```

---

**Congratulations! Your development environment is ready! 🎉**
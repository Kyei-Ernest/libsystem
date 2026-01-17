#!/bin/bash

# ==========================================
# Quick Start Script - Get Running Fast!
# ==========================================

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

clear
echo -e "${BLUE}"
cat << "EOF"
╔═══════════════════════════════════════════╗
║   Library Management System               ║
║   Quick Start                             ║
╚═══════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    echo "Please start Docker and try again"
    exit 1
fi

echo -e "${BLUE}Step 1: Creating .env file...${NC}"
if [ ! -f ".env" ]; then
    cat > .env << 'EOF'
ENVIRONMENT=development
PORT=8085
DATABASE_URL=postgres://libsystem:libsystem_dev_pass@postgres:5432/libsystem?sslmode=disable
REDIS_URL=redis://:libsystem_redis_pass@redis:6380/0
ELASTICSEARCH_URL=http://elasticsearch:9200
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
KAFKA_BROKERS=kafka:29092
JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
LOG_LEVEL=debug
EOF
    echo -e "${GREEN}✓ Created .env file${NC}"
else
    echo -e "${YELLOW}ℹ .env already exists${NC}"
fi

echo ""
echo -e "${BLUE}Step 2: Starting services (this may take a few minutes)...${NC}"
docker-compose up -d

echo ""
echo -e "${BLUE}Step 3: Waiting for services to be ready...${NC}"

# Wait for PostgreSQL
echo -n "  Waiting for PostgreSQL..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U libsystem > /dev/null 2>&1; then
        echo -e " ${GREEN}✓${NC}"
        break
    fi
    sleep 2
    echo -n "."
done

# Wait for Redis
echo -n "  Waiting for Redis..."
for i in {1..30}; do
    if docker-compose exec -T redis redis-cli -a libsystem_redis_pass ping > /dev/null 2>&1; then
        echo -e " ${GREEN}✓${NC}"
        break
    fi
    sleep 2
    echo -n "."
done

# Wait for Elasticsearch
echo -n "  Waiting for Elasticsearch..."
for i in {1..60}; do
    if curl -s http://localhost:9200/_cluster/health > /dev/null 2>&1; then
        echo -e " ${GREEN}✓${NC}"
        break
    fi
    sleep 2
    echo -n "."
done

# Wait for API Gateway
echo -n "  Waiting for API Gateway..."
for i in {1..30}; do
    if curl -s http://localhost:8085/health > /dev/null 2>&1; then
        echo -e " ${GREEN}✓${NC}"
        break
    fi
    sleep 2
    echo -n "."
done

echo ""
echo -e "${BLUE}Step 4: Initializing MinIO buckets...${NC}"
docker-compose exec -T minio sh -c '
    mc alias set local http://localhost:9000 minioadmin minioadmin123 2>/dev/null || true
    mc mb local/documents 2>/dev/null || true
    mc mb local/thumbnails 2>/dev/null || true
    mc policy set public local/thumbnails 2>/dev/null || true
' > /dev/null 2>&1
echo -e "${GREEN}✓ MinIO initialized${NC}"

echo ""
echo -e "${BLUE}Step 5: Creating Kafka topics...${NC}"
docker-compose exec -T kafka sh -c '
    kafka-topics --create --if-not-exists --topic document.uploaded --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1 2>/dev/null || true
    kafka-topics --create --if-not-exists --topic document.indexed --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1 2>/dev/null || true
    kafka-topics --create --if-not-exists --topic document.deleted --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1 2>/dev/null || true
' > /dev/null 2>&1
echo -e "${GREEN}✓ Kafka topics created${NC}"

echo ""
echo -e "${GREEN}═══════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ Setup Complete! System is ready!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════${NC}"
echo ""

echo -e "${BLUE}Access your services:${NC}"
echo ""
echo -e "  📡 API Gateway:       ${YELLOW}http://localhost:8085${NC}"
echo -e "  📊 Grafana:           ${YELLOW}http://localhost:3000${NC} (admin/admin)"
echo -e "  🔍 Kibana:            ${YELLOW}http://localhost:5601${NC}"
echo -e "  💾 MinIO Console:     ${YELLOW}http://localhost:9001${NC} (minioadmin/minioadmin123)"
echo -e "  📈 Prometheus:        ${YELLOW}http://localhost:9090${NC}"
echo -e "  🔎 Jaeger:            ${YELLOW}http://localhost:16686${NC}"
echo ""

echo -e "${BLUE}Quick test:${NC}"
echo ""
echo -e "  Health check:"
echo -e "  ${YELLOW}curl http://localhost:8085/health${NC}"
echo ""
echo -e "  Register a user:"
echo -e '  curl -X POST http://localhost:8085/api/v1/users/register \'
echo -e '    -H "Content-Type: application/json" \'
echo -e '    -d '"'"'{"email":"test@example.com","username":"testuser","password":"password123"}'"'"
echo ""

echo -e "${BLUE}Useful commands:${NC}"
echo ""
echo -e "  ${YELLOW}make logs${NC}          - View all logs"
echo -e "  ${YELLOW}make logs-api${NC}      - View API logs"
echo -e "  ${YELLOW}make ps${NC}            - Show service status"
echo -e "  ${YELLOW}make down${NC}          - Stop all services"
echo -e "  ${YELLOW}make help${NC}          - Show all commands"
echo ""

echo -e "${GREEN}Happy coding! 🚀${NC}"
echo ""

# Test the health endpoint
echo -e "${BLUE}Testing API Gateway...${NC}"
if response=$(curl -s http://localhost:8085/health); then
    echo -e "${GREEN}✓ API Gateway is responding!${NC}"
    echo "$response" | jq . 2>/dev/null || echo "$response"
else
    echo -e "${RED}⚠ API Gateway is not responding yet. Give it a few more seconds.${NC}"
fi

echo ""
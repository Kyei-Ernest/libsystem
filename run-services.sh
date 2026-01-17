#!/bin/bash

# Run all LibSystem services locally with VERBOSE LOGGING
# Usage: ./run-services.sh
#
# Logs are streamed to terminal AND saved to log files for debugging

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Create logs directory
mkdir -p logs

# Trap to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Stopping all services...${NC}"
    kill $(jobs -p) 2>/dev/null || true
    echo "Services stopped."
    exit
}
trap cleanup EXIT INT TERM

echo -e "${BLUE}🚀 Starting LibSystem Services (VERBOSE MODE)${NC}"
echo "================================================"
echo -e "${CYAN}Logs will be streamed to terminal AND saved to ./logs/${NC}"
echo ""

# Check if PostgreSQL is running
if ! systemctl is-active --quiet postgresql; then
    echo -e "${YELLOW}Warning: PostgreSQL service might not be active. Please ensure it's running.${NC}"
fi

# Export environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=libsystem
export DB_PASSWORD=libsystem

# Shared Config
export KAFKA_BROKERS=localhost:9092
export ELASTICSEARCH_URL=http://localhost:9200
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=minioadmin
export MINIO_SECRET_KEY=minioadmin123
export DOCUMENT_SERVICE_URL="http://localhost:8081"
export SERVICE_SECRET="internal-secret-key"

# Enable verbose logging for Go services
export LOG_LEVEL=debug

# Build and Start Function with live logging
start_service() {
    SERVICE=$1
    PORT=$2
    COLOR=$3
    
    echo -e "\n${COLOR}Building ${SERVICE}...${NC}"
    cd services/${SERVICE}
    go mod tidy > /dev/null 2>&1
    GOTOOLCHAIN=auto go build -o ${SERVICE} .
    
    echo -e "${COLOR}Starting ${SERVICE} on port ${PORT}...${NC}"
    
    # Run with tee to stream to terminal AND log file
    PORT=${PORT} ./${SERVICE} 2>&1 | tee ../../logs/${SERVICE}.log | sed "s/^/[${SERVICE}] /" &
    PID=$!
    cd ../..
    
    # Wait for startup
    echo "Waiting for ${SERVICE} to be healthy..."
    for i in {1..15}; do
        if [ "$SERVICE" == "indexer-service" ]; then
             if ps -p $PID > /dev/null 2>&1; then
                 echo -e "${GREEN}✅ ${SERVICE} running (PID: $PID)${NC}"
                 break
             else
                 echo -e "${RED}❌ ${SERVICE} failed to start${NC}"
                 return 1
             fi
        else
            if curl -s http://localhost:${PORT}/health > /dev/null 2>&1; then
                echo -e "${GREEN}✅ ${SERVICE} running (PID: $PID)${NC}"
                break
            fi
        fi
        
        if [ $i -eq 15 ]; then
            echo -e "${RED}❌ ${SERVICE} timed out waiting for health check${NC}"
            echo -e "${YELLOW}Check logs above for errors${NC}"
        fi
        sleep 2
    done
}

# Start services with verbose logging
start_service "user-service" "8086" "${BLUE}"
start_service "collection-service" "8082" "${CYAN}"
start_service "document-service" "8081" "${GREEN}"
start_service "search-service" "8084" "${YELLOW}"
start_service "api-gateway" "8088" "${RED}"

# Indexer Service
echo -e "\n${BLUE}Building and Starting indexer-service...${NC}"
cd services/indexer-service
go mod tidy > /dev/null 2>&1
GOTOOLCHAIN=auto go build -o indexer-service .
./indexer-service 2>&1 | tee ../../logs/indexer-service.log | sed "s/^/[indexer-service] /" &
INDEXER_PID=$!
cd ../..

# Analytics Service
echo -e "\n${BLUE}Building and Starting analytics-service...${NC}"
cd services/analytics-service
go mod tidy > /dev/null 2>&1
GOTOOLCHAIN=auto go build -o analytics-service .
./analytics-service 2>&1 | tee ../../logs/analytics-service.log | sed "s/^/[analytics-service] /" &
ANALYTICS_PID=$!
cd ../..

sleep 2

if ps -p $ANALYTICS_PID > /dev/null 2>&1; then
     echo -e "${GREEN}✅ analytics-service running (PID: $ANALYTICS_PID)${NC}"
else
     echo -e "${RED}❌ analytics-service failed to start${NC}"
fi

if ps -p $INDEXER_PID > /dev/null 2>&1; then
     echo -e "${GREEN}✅ indexer-service running (PID: $INDEXER_PID)${NC}"
else
     echo -e "${RED}❌ indexer-service failed to start${NC}"
fi

echo ""
echo "================================"
echo -e "${GREEN}✨ All services are running with VERBOSE logging!${NC}"
echo ""
echo -e "${CYAN}Service URLs:${NC}"
echo "  User Service:       http://localhost:8086"
echo "  Collection Service: http://localhost:8082"
echo "  Document Service:   http://localhost:8081"
echo "  Search Service:     http://localhost:8084"
echo "  API Gateway:        http://localhost:8088"
echo ""
echo -e "${YELLOW}📝 All requests will be logged to terminal${NC}"
echo -e "${YELLOW}   Look for [document-service] logs when uploading${NC}"
echo ""
echo "Press Ctrl+C to stop all services"
echo "================================"

# Wait for all background jobs
wait

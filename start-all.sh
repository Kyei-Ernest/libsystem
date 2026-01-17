#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

mkdir -p logs

echo -e "${BLUE}🚀 Starting LibSystem Infrastructure & Services${NC}"
echo "================================================"

# 1. System Services (Postgres & Redis)
echo -e "\n${BLUE}Checking System Services...${NC}"

if systemctl is-active --quiet postgresql; then
    echo -e "${GREEN}✅ PostgreSQL is running${NC}"
else
    echo -e "${RED}⚠️  PostgreSQL is not running. Attempting to start...${NC}"
    sudo systemctl start postgresql
    sleep 2
    if systemctl is-active --quiet postgresql; then
        echo -e "${GREEN}✅ PostgreSQL started${NC}"
    else
        echo -e "${RED}❌ Failed to start PostgreSQL. Please start it manually.${NC}"
        exit 1
    fi
fi

if systemctl is-active --quiet redis-server || systemctl is-active --quiet redis; then
     echo -e "${GREEN}✅ Redis is running${NC}"
else
    # Try starting redis-server
    echo -e "${RED}⚠️  Redis is not running. Attempting to start...${NC}"
    if sudo systemctl start redis-server 2>/dev/null || sudo systemctl start redis 2>/dev/null; then
         sleep 2
         echo -e "${GREEN}✅ Redis started${NC}"
    else
         echo -e "${RED}❌ Failed to start Redis. Please start it manually.${NC}"
         # Don't exit, might be running as non-systemd
    fi
fi

# 2. Local Binaries (MinIO, Elasticsearch, Kafka)
echo -e "\n${BLUE}Starting Local Infrastructure...${NC}"

# MinIO
if lsof -i :9000 >/dev/null; then
    echo -e "${GREEN}✅ MinIO is already running${NC}"
else
    echo "Starting MinIO..."
    ./start-minio.sh > logs/minio.log 2>&1 &
    MINIO_PID=$!
    echo "MinIO started (PID: $MINIO_PID)"
fi

# Elasticsearch
if curl -s http://localhost:9200 >/dev/null; then
     echo -e "${GREEN}✅ Elasticsearch is already running${NC}"
else
    echo "Starting Elasticsearch (this may take a moment)..."
    ./elasticsearch-8.11.1/bin/elasticsearch -d -p logs/es.pid > logs/elasticsearch.log 2>&1
    echo "Elasticsearch started in background"
fi

# Zookeeper & Kafka
if lsof -i :2181 >/dev/null; then
     echo -e "${GREEN}✅ Zookeeper is already running${NC}"
else
    echo "Starting Zookeeper..."
    nohup ./kafka_2.13-3.9.1/bin/zookeeper-server-start.sh ./kafka_2.13-3.9.1/config/zookeeper.properties > logs/zookeeper.log 2>&1 &
    echo "Zookeeper started"
fi

if lsof -i :9093 >/dev/null || lsof -i :9092 >/dev/null; then
     echo -e "${GREEN}✅ Kafka is already running${NC}"
else
    echo "Waiting a moment for Zookeeper..."
    sleep 5
    echo "Starting Kafka..."
    # Note: Using 9093 as expected by run-services.sh. We might need to override config if it defaults to 9092.
    # We will assume config is set, or run command with override if needed. 
    # For now, running with default config.
    nohup ./kafka_2.13-3.9.1/bin/kafka-server-start.sh ./kafka_2.13-3.9.1/config/server.properties > logs/kafka.log 2>&1 &
    echo "Kafka started"
fi

# 3. Wait for Readiness
echo -e "\n${BLUE}Waiting for Infrastructure Readiness...${NC}"

wait_port() {
    PORT=$1
    NAME=$2
    MAX_RETRIES=60 # 60 seconds
    for i in $(seq 1 $MAX_RETRIES); do
        if (echo > /dev/tcp/localhost/$PORT) >/dev/null 2>&1; then
            echo -e "${GREEN}✅ $NAME is ready (Port $PORT)${NC}"
            return 0
        fi
        sleep 1
        echo -n "."
    done
    echo -e "\n${RED}❌ $NAME timed out waiting for port $PORT${NC}"
    return 1
}

wait_port 5432 "PostgreSQL"
wait_port 9000 "MinIO"
wait_port 2181 "Zookeeper"
# Check 9092 for Kafka
if lsof -i :9092 >/dev/null; then
    echo -e "${GREEN}✅ Kafka is ready (Port 9092)${NC}"
elif lsof -i :9093 >/dev/null; then
    echo -e "${GREEN}✅ Kafka is ready (Port 9093)${NC}"
    echo -e "${RED}⚠️  Warning: Kafka is on 9093, but services now expect 9092!${NC}"
else
    # wait for 9092 default
     wait_port 9092 "Kafka"
fi

# Elasticsearch Health Check
echo "Waiting for Elasticsearch health..."
for i in {1..60}; do
    if curl -s http://localhost:9200/_cluster/health >/dev/null; then
        echo -e "${GREEN}✅ Elasticsearch is healthy${NC}"
        break
    fi
    sleep 2
    echo -n "."
done

# 4. Start Application Services
echo -e "\n${BLUE}Starting Application Services...${NC}"
./run-services.sh

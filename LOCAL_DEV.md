# Local Development Guide

Quick guide for running LibSystem services locally without Docker.

## Prerequisites

- Go 1.23+
- PostgreSQL 16+
- curl (for testing)

## Quick Start

```bash
# 1. Make scripts executable
chmod +x setup-local.sh run-services.sh

# 2. Run setup (one-time)
./setup-local.sh

# 3. Start services
./run-services.sh
```

## Manual Setup

### 1. Install PostgreSQL

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

### 2. Create Database

```bash
sudo -u postgres psql <<EOF
CREATE DATABASE libsystem;
CREATE USER libsystem WITH PASSWORD 'libsystem';
GRANT ALL PRIVILEGES ON DATABASE libsystem TO libsystem;
ALTER DATABASE libsystem OWNER TO libsystem;
EOF
```

### 3. Run Migrations

```bash
cd migrations
for f in *.up.sql; do
    PGPASSWORD=libsystem psql -U libsystem -d libsystem -h localhost -f "$f"
done
cd ..
```

### 4. Build Services

```bash
# User Service
cd services/user-service && go build . && cd ../..

# Collection Service
cd services/collection-service && go build . && cd ../..

# Document Service
cd services/document-service && go build . && cd ../..
```

### 5. Run Services

Open 3 terminals:

**Terminal 1 - User Service:**
```bash
cd services/user-service
PORT=8086 DB_HOST=localhost DB_PORT=5432 DB_USER=libsystem DB_PASSWORD=libsystem DB_NAME=libsystem JWT_SECRET=test-secret ./user-service
```

**Terminal 2 - Collection Service:**
```bash
cd services/collection-service
PORT=8082 DB_HOST=localhost DB_PORT=5432 DB_USER=libsystem DB_PASSWORD=libsystem DB_NAME=libsystem ./collection-service
```

**Terminal 3 - Document Service:**
```bash
cd services/document-service
PORT=8081 DB_HOST=localhost DB_PORT=5432 DB_USER=libsystem DB_PASSWORD=libsystem DB_NAME=libsystem ./document-service
```

## Testing

### Health Checks

```bash
curl http://localhost:8086/health  # User Service
curl http://localhost:8082/health  # Collection Service
curl http://localhost:8081/health  # Document Service
```

### Register a User

```bash
curl -X POST http://localhost:8086/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "Test@1234",
    "first_name": "Test",
    "last_name": "User"
  }'
```

### Login

```bash
curl -X POST http://localhost:8086/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email_or_username": "test@example.com",
    "password": "Test@1234"
  }'
```

Save the token from the response!

### Create a Collection

```bash
TOKEN="your-token-here"

curl -X POST http://localhost:8082/api/v1/collections \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "My First Collection",
    "description": "A test collection",
    "is_public": true
  }'
```

### List Collections

```bash
curl http://localhost:8082/api/v1/collections
```

## Troubleshooting

### Port Already in Use

If port 8086 conflicts with nginx, use a different port:
```bash
PORT=8090 ./user-service
```

### Database Connection Error

Check if PostgreSQL is running:
```bash
sudo systemctl status postgresql
```

### View Service Logs

If using run-services.sh:
```bash
tail -f logs/user-service.log
tail -f logs/collection-service.log
tail -f logs/document-service.log
```

## Stopping Services

If using run-services.sh, press `Ctrl+C`.

If running manually, find and kill processes:
```bash
pkill -f user-service
pkill -f collection-service
pkill -f document-service
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Service port | Various |
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | libsystem |
| DB_PASSWORD | Database password | libsystem |
| DB_NAME | Database name | libsystem |
| JWT_SECRET | JWT signing key | test-secret |

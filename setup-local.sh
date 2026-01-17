#!/bin/bash

# LibSystem Local Development Setup Script
# This script helps you set up and run the services locally without Docker

set -e

echo "🚀 LibSystem Local Development Setup"
echo "====================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if PostgreSQL is installed
echo "1️⃣ Checking PostgreSQL..."
if ! command -v psql &> /dev/null; then
    echo -e "${RED}❌ PostgreSQL is not installed${NC}"
    echo "Install it with: sudo apt install postgresql postgresql-contrib"
    exit 1
else
    echo -e "${GREEN}✅ PostgreSQL is installed${NC}"
fi

# Check if PostgreSQL is running
if ! sudo systemctl is-active --quiet postgresql; then
    echo "Starting PostgreSQL..."
    sudo systemctl start postgresql
fi

# Create database and user
echo ""
echo "2️⃣ Setting up database..."
sudo -u postgres psql -c "SELECT 1 FROM pg_database WHERE datname='libsystem'" | grep -q 1 || \
sudo -u postgres psql <<EOF
CREATE DATABASE libsystem;
CREATE USER libsystem WITH PASSWORD 'libsystem';
GRANT ALL PRIVILEGES ON DATABASE libsystem TO libsystem;
ALTER DATABASE libsystem OWNER TO libsystem;
\c libsystem
GRANT ALL ON SCHEMA public TO libsystem;
EOF

echo -e "${GREEN}✅ Database setup complete${NC}"

# Run migrations
echo ""
echo "3️⃣ Running database migrations..."
cd migrations
for migration in *.up.sql; do
    echo "Running $migration..."
    PGPASSWORD=libsystem psql -U libsystem -d libsystem -h localhost -f "$migration"
done
cd ..

echo -e "${GREEN}✅ Migrations complete${NC}"

# Build services
echo ""
echo "4️⃣ Building services..."

cd services/user-service
echo "Syncing dependencies for user-service..."
go mod tidy
echo "Building user-service..."
go build -o user-service .
cd ../..

cd services/collection-service
echo "Syncing dependencies for collection-service..."
go mod tidy
echo "Building collection-service..."
go build -o collection-service .
cd ../..

cd services/document-service
echo "Syncing dependencies for document-service..."
go mod tidy
echo "Building document-service..."
go build -o document-service .
cd ../..

echo -e "${GREEN}✅ All services built successfully${NC}"

echo ""
echo "====================================="
echo -e "${GREEN}✨ Setup Complete!${NC}"
echo ""
echo "To start the services, run:"
echo -e "${YELLOW}  ./run-services.sh${NC}"
echo ""
echo "Or start them individually:"
echo -e "${YELLOW}  cd services/user-service && ./user-service &${NC}"
echo -e "${YELLOW}  cd services/collection-service && ./collection-service &${NC}"
echo -e "${YELLOW}  cd services/document-service && ./document-service &${NC}"

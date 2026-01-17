#!/bin/bash

# Docker Compose Quick Start Script
# This script helps you build and test the libsystem services

echo "🚀 LibSystem Docker Compose Quick Start"
echo "========================================"
echo ""

# Stop any running containers
echo "1️⃣ Stopping existing containers..."
sudo docker-compose down

echo ""
echo "2️⃣ Building and starting all services..."
sudo docker-compose up -d --build

echo ""
echo "3️⃣ Waiting for services to be ready (30 seconds)..."
sleep 30

echo ""
echo "4️⃣ Checking service health..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo -n "User Service (8086):       "
if curl -s http://localhost:8086/health > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Down"
fi

echo -n "Collection Service (8082): "
if curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Down"
fi

echo -n "Document Service (8081):   "
if curl -s http://localhost:8081/health > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Down"
fi

echo -n "API Gateway (8085):        "
if curl -s http://localhost:8085/health > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Down"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 Quick Test Commands:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "▶️  Register a new user:"
echo "curl -X POST http://localhost:8086/api/v1/auth/register \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"email\":\"test@example.com\",\"username\":\"testuser\",\"password\":\"Test@1234\",\"first_name\":\"Test\",\"last_name\":\"User\"}'"
echo ""
echo "▶️  Login:"
echo "curl -X POST http://localhost:8086/api/v1/auth/login \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"email_or_username\":\"test@example.com\",\"password\":\"Test@1234\"}'"
echo ""
echo "▶️  View container logs:"
echo "sudo docker-compose logs -f user-service"
echo "sudo docker-compose logs -f collection-service"
echo "sudo docker-compose logs -f document-service"
echo ""
echo "▶️  Check running containers:"
echo "sudo docker-compose ps"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✨ Setup complete! Your services should be running."

#!/bin/bash

# LibSystem Full Stack Startup Script
# Starts both Backend (Infrastructure + Services) and Frontend

set -e

# Load NVM if it exists
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

echo "🚀 Starting Full Stack Development Environment"
echo "   - Backend: ./start-all.sh"
echo "   - Frontend: cd frontend && npm run dev"
echo "================================================"

# Check if we have npx
if ! command -v npx &> /dev/null; then
    echo "Error: npx is not installed. Please install Node.js."
    exit 1
fi

# Use concurrently to run both processes
# --kill-others: If one processes dies (e.g. Ctrl+C), kill the other
# --names: Labels for the logs
# --prefix-colors: Colors for the labels
npx -y concurrently --kill-others \
    --names "BACKEND,FRONTEND" \
    --prefix-colors "blue,magenta" \
    "./start-all.sh" \
    "cd frontend && npm run dev"

#!/bin/bash

# LibSystem Frontend Startup Script

# Load NVM if it exists
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

echo "🚀 Starting Frontend..."
echo "Dir: frontend"
echo "Command: npm run dev"
echo "============================="

cd frontend
npm run dev

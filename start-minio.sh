#!/bin/bash
# MinIO Local Server Startup Script

echo "Starting MinIO server..."

# Create data directory if it doesn't exist
mkdir -p ~/minio/data

# Set MinIO credentials
export MINIO_ROOT_USER=minioadmin
export MINIO_ROOT_PASSWORD=minioadmin123

# Start MinIO server
# API on port 9000, Console on port 9001
minio server ~/minio/data --console-address ":9001"

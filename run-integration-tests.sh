#!/bin/bash

# run-integration-tests.sh
# Runs the Go integration test suite

set -e

# Ensure services are running? 
# We assume they are running via ./run-services.sh
# simple check:
echo "Checking if services are reachable..."
if ! curl -s http://localhost:8086/health > /dev/null; then
    echo "User Service not reachable. Please run ./run-services.sh first."
    exit 1
fi

echo "Running Integration Tests..."
go test -v ./tests/integration/...

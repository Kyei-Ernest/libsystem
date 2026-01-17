#!/bin/bash

# Start Prometheus and Grafana for Local Development
# This script downloads and runs Prometheus + Grafana binaries locally

set -e

echo "🚀 Setting up LibSystem Monitoring (Local Mode)..."

PROM_VERSION="2.45.0"
GRAFANA_VERSION="10.2.0"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

# Create monitoring directory
mkdir -p infrastructure/bin
cd infrastructure

# Download and setup Prometheus
if [ ! -f "bin/prometheus" ]; then
    echo "📊 Downloading Prometheus ${PROM_VERSION}..."
    wget -q "https://github.com/prometheus/prometheus/releases/download/v${PROM_VERSION}/prometheus-${PROM_VERSION}.${OS}-${ARCH}.tar.gz"
    tar xzf "prometheus-${PROM_VERSION}.${OS}-${ARCH}.tar.gz"
    mv "prometheus-${PROM_VERSION}.${OS}-${ARCH}/prometheus" bin/
    mv "prometheus-${PROM_VERSION}.${OS}-${ARCH}/promtool" bin/
    rm -rf "prometheus-${PROM_VERSION}.${OS}-${ARCH}"*
    echo "✅ Prometheus installed"
else
    echo "✅ Prometheus already installed"
fi

# Download and setup Grafana
if [ ! -f "bin/grafana-server" ]; then
    echo "📈 Downloading Grafana ${GRAFANA_VERSION}..."
    wget -q "https://dl.grafana.com/oss/release/grafana-${GRAFANA_VERSION}.${OS}-${ARCH}.tar.gz"
    tar xzf "grafana-${GRAFANA_VERSION}.${OS}-${ARCH}.tar.gz"
    mv "grafana-${GRAFANA_VERSION}" grafana
    ln -s ../grafana/bin/grafana-server bin/grafana-server
    rm "grafana-${GRAFANA_VERSION}.${OS}-${ARCH}.tar.gz"
    
    # Configure Grafana provisioning
    mkdir -p grafana/conf/provisioning/datasources
    mkdir -p grafana/conf/provisioning/dashboards
    cp grafana/provisioning/datasources/prometheus.yml grafana/conf/provisioning/datasources/
    cp grafana/provisioning/dashboards/dashboards.yml grafana/conf/provisioning/dashboards/
    
    echo "✅ Grafana installed"
else
    echo "✅ Grafana already installed"
fi

cd ..

# Start Prometheus
echo "🚀 Starting Prometheus..."
nohup infrastructure/bin/prometheus \
    --config.file=infrastructure/prometheus/prometheus.yml \
    --storage.tsdb.path=infrastructure/prometheus/data \
    --web.listen-address=:9090 \
    > logs/prometheus.log 2>&1 &
    
echo $! > logs/prometheus.pid
echo "✅ Prometheus started on http://localhost:9090 (PID: $(cat logs/prometheus.pid))"

# Wait for Prometheus
sleep 2

# Start Grafana
echo "🚀 Starting Grafana..."
cd infrastructure/grafana
nohup bin/grafana-server \
    --config=conf/defaults.ini \
    --homepath=. \
    > ../../logs/grafana.log 2>&1 &

echo $! > ../../logs/grafana.pid
cd ../..
echo "✅ Grafana started on http://localhost:3000 (PID: $(cat logs/grafana.pid))"
echo "   Default login: admin / admin"

sleep 3

echo ""
echo "📊 Monitoring Stack Status:"
echo "  Prometheus: http://localhost:9090"
echo "  Grafana:    http://localhost:3000"
echo ""
echo "To stop:"
echo "  kill \$(cat logs/prometheus.pid logs/grafana.pid)"

#!/bin/bash

# Stop Monitoring Stack

if [ -f logs/prometheus.pid ]; then
    echo "🛑 Stopping Prometheus..."
    kill $(cat logs/prometheus.pid) 2>/dev/null || echo "Prometheus not running"
    rm logs/prometheus.pid
fi

if [ -f logs/grafana.pid ]; then
    echo "🛑 Stopping Grafana..."
    kill $(cat logs/grafana.pid) 2>/dev/null || echo "Grafana not running"
    rm logs/grafana.pid
fi

echo "✅ Monitoring stopped"

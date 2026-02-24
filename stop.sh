#!/bin/bash
# stop.sh - Stop all services

echo "🛑 Stopping PLC Monitoring System..."

# Graceful shutdown with 30s timeout
docker-compose down -t 30

echo "✅ System stopped."

#!/bin/bash
# start.sh - Start all services

set -e

echo "🚀 Starting PLC Monitoring System..."

# Create directories if they don't exist
mkdir -p logs/plc-core
mkdir -p exports

# Check for .env file
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        echo "⚠️ .env file not found. Creating from .env.example..."
        cp .env.example .env
        echo "🔔 Please edit .env with your specific configuration."
    else
        echo "❌ .env and .env.example not found. Please create .env."
        exit 1
    fi
fi

# Build and start the stack
echo "🏗️ Building and starting containers..."
docker-compose up -d --build

echo "✅ System started successfully!"
echo "📊 URLs:"
echo "   - Dashboard: http://localhost"
echo "   - API Gateway: http://localhost/api"
echo "   - Traefik Dashboard: http://localhost:8081"

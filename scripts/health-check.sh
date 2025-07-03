#!/bin/bash

# Health Check Script for Shield Platform
# Verifies that all services are running and healthy

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🏥 Shield Platform Health Check"
echo "================================"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running"
    exit 1
fi
echo "✅ Docker is running"

# Check Docker Compose services
echo ""
echo "📊 Service Status:"
cd "$PROJECT_ROOT"
docker-compose -f docker-compose.yml -f docker-compose.override.yml ps

# Check PostgreSQL
echo ""
echo "🗄️  PostgreSQL Health:"
if docker-compose -f docker-compose.yml -f docker-compose.override.yml exec -T postgres psql -U shield -d shield_auth -c "SELECT 'PostgreSQL is healthy' as status;" > /dev/null 2>&1; then
    echo "✅ PostgreSQL is healthy and accessible"
else
    echo "❌ PostgreSQL is not accessible"
fi

# Check Redis
echo ""
echo "📦 Redis Health:"
if docker-compose -f docker-compose.yml -f docker-compose.override.yml exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis is healthy and accessible"
else
    echo "❌ Redis is not accessible"
fi

# Check environment file
echo ""
echo "🔧 Environment Configuration:"
if [ -f "$PROJECT_ROOT/.env" ]; then
    ENV_NAME=$(grep "^ENVIRONMENT=" "$PROJECT_ROOT/.env" | cut -d'=' -f2 | tr -d '"')
    echo "✅ Environment file exists (Environment: $ENV_NAME)"
else
    echo "❌ No .env file found"
fi

# Check Go modules
echo ""
echo "🐹 Go Environment:"
if command -v go > /dev/null 2>&1; then
    echo "✅ Go is installed: $(go version)"
    if [ -f "$PROJECT_ROOT/go.mod" ]; then
        echo "✅ Go modules configured"
    else
        echo "❌ go.mod not found"
    fi
else
    echo "❌ Go is not installed"
fi

# Check Air (hot reload)
echo ""
echo "🌪️  Air (Hot Reload):"
if command -v air > /dev/null 2>&1; then
    echo "✅ Air is installed: $(air -v 2>/dev/null || echo 'version unknown')"
else
    echo "⚠️  Air is not installed (run: go install github.com/air-verse/air@latest)"
fi

echo ""
echo "🎯 Quick Start Commands:"
echo "  make quick-start    - Start all services"
echo "  make dev           - Start with hot reload"
echo "  make services-logs - View service logs"
echo "  make db-status     - Check database status"
echo ""
echo "🌐 Service URLs:"
echo "  API: http://localhost:8081"
echo "  pgAdmin: http://localhost:5050 (admin@shield.dev / admin123)"
echo "  Redis Commander: http://localhost:8082"

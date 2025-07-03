#!/bin/bash

# Environment Configuration Script
# Usage: ./scripts/env.sh [environment]
# Environments: development, staging, production, test

set -e

ENVIRONMENT=${1:-development}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 Setting up environment: $ENVIRONMENT"

# Validate environment
case $ENVIRONMENT in
    development|staging|production|test)
        echo "✅ Valid environment: $ENVIRONMENT"
        ;;
    *)
        echo "❌ Invalid environment: $ENVIRONMENT"
        echo "Valid options: development, staging, production, test"
        exit 1
        ;;
esac

# Copy environment file
ENV_FILE="$PROJECT_ROOT/.env.$ENVIRONMENT"
TARGET_FILE="$PROJECT_ROOT/.env"

if [ -f "$ENV_FILE" ]; then
    echo "📋 Copying $ENV_FILE to $TARGET_FILE"
    cp "$ENV_FILE" "$TARGET_FILE"
    echo "✅ Environment file copied successfully"
else
    echo "❌ Environment file not found: $ENV_FILE"
    exit 1
fi

# Load environment variables for validation
if [ -f "$TARGET_FILE" ]; then
    set -a
    source "$TARGET_FILE"
    set +a
    echo "✅ Environment variables loaded"
fi

# Validate required environment variables
echo "🔍 Validating required environment variables..."

REQUIRED_VARS=(
    "PORT"
    "ENVIRONMENT"
    "DB_HOST"
    "DB_PORT"
    "DB_USER"
    "DB_NAME"
)

MISSING_VARS=()

for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        MISSING_VARS+=("$var")
    fi
done

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    echo "❌ Missing required environment variables:"
    printf " - %s\n" "${MISSING_VARS[@]}"
    exit 1
fi

echo "✅ All required environment variables are set"

# Environment-specific setup
case $ENVIRONMENT in
    development)
        echo "🔧 Setting up development environment..."
        
        # Check if Docker is running
        if ! docker info > /dev/null 2>&1; then
            echo "❌ Docker is not running. Please start Docker first."
            exit 1
        fi
        
        # Start development services
        echo "🐳 Starting development services..."
        docker-compose -f "$PROJECT_ROOT/docker-compose.yml" -f "$PROJECT_ROOT/docker-compose.override.yml" up -d postgres redis pgadmin redis-commander
        
        echo "⏳ Waiting for database to be ready..."
        
        # Wait for PostgreSQL to be ready using a more reliable approach
        MAX_ATTEMPTS=30
        ATTEMPT=0
        
        while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
            ATTEMPT=$((ATTEMPT + 1))
            echo "Attempt $ATTEMPT/$MAX_ATTEMPTS: Checking database connection..."
            
            # Try to connect to PostgreSQL using docker exec
            if docker-compose -f "$PROJECT_ROOT/docker-compose.yml" -f "$PROJECT_ROOT/docker-compose.override.yml" exec -T postgres psql -U shield -d shield_auth -c "SELECT 1;" > /dev/null 2>&1; then
                echo "✅ Database is ready!"
                break
            fi
            
            if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
                echo "⚠️  Database timeout after $MAX_ATTEMPTS attempts"
                echo "🔍 Checking container status..."
                docker-compose -f "$PROJECT_ROOT/docker-compose.yml" -f "$PROJECT_ROOT/docker-compose.override.yml" ps postgres
                echo ""
                echo "📋 Container logs (last 10 lines):"
                docker-compose -f "$PROJECT_ROOT/docker-compose.yml" -f "$PROJECT_ROOT/docker-compose.override.yml" logs --tail=10 postgres
                echo ""
                echo "💡 You can continue anyway - the database might need more time"
                echo "💡 Try running: docker-compose logs postgres"
                break
            fi
            
            sleep 3
        done
        
        echo "✅ Development environment ready!"
        echo "📊 Services available:"
        echo "  - PostgreSQL: localhost:5432"
        echo "  - Redis: localhost:6379"
        echo "  - pgAdmin: http://localhost:5050 (admin@shield.local / admin123)"
        echo "  - Redis Commander: http://localhost:8082"
        ;;
        
    staging|production)
        echo "🔧 Setting up $ENVIRONMENT environment..."
        echo "⚠️  Make sure all external services are configured:"
        echo "  - Database connection"
        echo "  - Redis/ElastiCache"
        echo "  - AWS Cognito"
        echo "  - Monitoring services"
        ;;
        
    test)
        echo "🔧 Setting up test environment..."
        echo "🧪 Test environment configured for automated testing"
        ;;
esac

echo ""
echo "🎉 Environment setup complete!"
echo "💡 To start the application:"
echo "   For development: docker-compose up shield-api"
echo "   For local development with hot reload: air"
echo "   For production: docker-compose -f docker-compose.yml up"

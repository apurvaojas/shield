#!/bin/bash

# Shield Platform Quick Start Script
# This script sets up the development environment and starts the application

set -e

echo "🛡️  Shield Platform Development Setup"
echo "======================================"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

# Check prerequisites
echo "📋 Checking prerequisites..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.22 or later."
    exit 1
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

echo "✅ Prerequisites check passed"

# Install development tools
echo "🔧 Installing development tools..."
make install-tools 2>/dev/null || {
    echo "Installing tools manually..."
    go install github.com/cosmtrek/air@latest
    go install github.com/swaggo/swag/cmd/swag@latest
}

# Setup development environment
echo "🌍 Setting up development environment..."
make setup-env ENV=development

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from development template..."
    cp .env.development .env
fi

# Start infrastructure services
echo "🐳 Starting infrastructure services..."
make docker-dev

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if database is ready
echo "🔍 Checking database connection..."
timeout 60 bash -c 'until docker-compose exec -T postgres pg_isready -U shield; do sleep 2; done' || {
    echo "❌ Database failed to start. Check Docker logs:"
    docker-compose logs postgres
    exit 1
}

# Generate documentation
echo "📚 Generating API documentation..."
make docs 2>/dev/null || echo "⚠️  Swagger docs generation skipped (swag not found)"

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Generate self-signed SSL certificate for local development if not present
CERT_DIR="$PROJECT_ROOT/dev-certs"
CERT_KEY="$CERT_DIR/dev.localhost.key"
CERT_CRT="$CERT_DIR/dev.localhost.crt"

mkdir -p "$CERT_DIR"
if [ ! -f "$CERT_KEY" ] || [ ! -f "$CERT_CRT" ]; then
    echo "🔐 Generating self-signed SSL certificate for https://localhost ..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$CERT_KEY" \
        -out "$CERT_CRT" \
        -subj "/C=US/ST=Dev/L=Localhost/O=Dev/OU=Dev/CN=localhost"
    echo "✅ SSL certificate generated at $CERT_CRT and $CERT_KEY"
else
    echo "🔐 SSL certificate already exists at $CERT_CRT and $CERT_KEY"
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "📊 Services Status:"
echo "  ✅ PostgreSQL: localhost:5432"
echo "  ✅ Redis: localhost:6379"
echo "  ✅ pgAdmin: http://localhost:5050"
echo "  ✅ Redis Commander: http://localhost:8082"
echo ""
echo "🚀 To start the application:"
echo "  Development (hot reload): make dev"
echo "  Normal run:              make run"
echo "  Docker:                  make docker-up"
echo ""
echo "📖 Other useful commands:"
echo "  make test                # Run tests"
echo "  make docs                # Generate API docs"
echo "  make lint                # Run linter"
echo "  make help                # Show all commands"
echo ""
echo "🌐 Once started, the API will be available at:"
echo "  Main API: http://localhost:8081"
echo "  Swagger UI: http://localhost:8081/swagger/index.html"
echo ""
echo "🔑 SSL Certificate for local development:"
echo "  Certificate: $CERT_CRT"
echo "  Key: $CERT_KEY"
echo ""
echo "📦 Note: Ensure your Docker container is configured to use the host's Docker daemon."
echo "  This is typically done by mounting the Docker socket:"
echo "    -v /var/run/docker.sock:/var/run/docker.sock"
echo "  in your Docker run or compose configuration."

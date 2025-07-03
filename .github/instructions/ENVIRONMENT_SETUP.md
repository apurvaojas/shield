# Environment Configuration Guide

This guide explains how to set up and manage different environments for the Shield Platform.

## 🌍 Available Environments

- **Development** (`.env.development`) - Local development with Docker services
- **Staging** (`.env.staging`) - Pre-production environment 
- **Production** (`.env.production`) - Production environment
- **Test** (`.env.test`) - Automated testing environment

## 🚀 Quick Start

### 1. Clone and Setup
```bash
git clone <repository-url>
cd shield
./dev-setup.sh
```

### 2. Start Development
```bash
# Start with hot reload
make dev

# Or start normally
make run
```

## 📁 Environment Files

### `.env.example`
Template file with all available configuration options. Copy this to create your environment files.

### `.env.development` 
Local development configuration:
- Uses local Docker services (PostgreSQL, Redis)
- Debug logging enabled
- Relaxed CORS and rate limiting
- Development Cognito pool

### `.env.staging`
Staging environment configuration:
- Points to staging AWS resources
- Info-level logging
- Moderate rate limiting
- All features enabled for testing

### `.env.production`
Production configuration:
- Uses environment variables for sensitive data
- Warn-level logging only
- Strict rate limiting and security
- All production features enabled

### `.env.test`
Testing configuration:
- In-memory or test databases
- Minimal logging
- Disabled rate limiting
- Mock services where possible

## 🔧 Environment Setup

### Manual Setup
```bash
# Setup specific environment
./scripts/env.sh [development|staging|production|test]

# This will:
# 1. Copy the appropriate .env file to .env
# 2. Validate required variables
# 3. Start necessary services (for development)
```

### Using Makefile
```bash
# Setup development environment
make setup-env ENV=development

# Setup production environment  
make setup-env ENV=production
```

## 🐳 Docker Usage

### Development with Hot Reload
```bash
# Start all development services
make docker-dev

# View logs
make docker-logs

# Stop services
make docker-down
```

### Production Deployment
```bash
# Build production image
make docker-build

# Start production services
make docker-up
```

## 🔐 Environment Variables

### Required Variables
- `PORT` - Server port (default: 8081)
- `ENVIRONMENT` - Environment name
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database connection
- `COGNITO_USER_POOL_ID`, `COGNITO_APP_CLIENT_ID` - AWS Cognito configuration

### Optional Variables
- `REDIS_HOST`, `REDIS_PORT` - Redis configuration
- `JWT_SECRET` - JWT signing secret
- `RATE_LIMIT_*` - Rate limiting configuration
- `FEATURE_*` - Feature flags
- `CORS_*` - CORS configuration

### Sensitive Variables (Production)
For production and staging, use environment variables instead of files:
```bash
export PROD_DB_PASSWORD="your-secure-password"
export PROD_JWT_SECRET="your-jwt-secret"
export PROD_COGNITO_APP_CLIENT_SECRET="your-cognito-secret"
```

## 🏗️ Docker Architecture

### Development Stack
- **shield-api**: Main application with Air hot reload
- **postgres**: PostgreSQL database
- **redis**: Redis for sessions
- **pgadmin**: Database management UI
- **redis-commander**: Redis management UI
- **opa**: Open Policy Agent for authorization

### Production Stack
- **shield-api**: Optimized production build
- **postgres**: PostgreSQL database
- **redis**: Redis with authentication
- **prometheus**: Metrics collection
- **jaeger**: Distributed tracing

## 📋 Configuration Examples

### Development
```env
ENVIRONMENT=development
LOG_LEVEL=debug
DB_HOST=localhost
RATE_LIMIT_REQUESTS_PER_MINUTE=1000
FEATURE_MULTI_FACTOR_AUTH=false
```

### Production
```env
ENVIRONMENT=production
LOG_LEVEL=warn
DB_HOST=${PROD_DB_HOST}
RATE_LIMIT_REQUESTS_PER_MINUTE=100
FEATURE_MULTI_FACTOR_AUTH=true
```

## 🛠️ Development Tools

### Hot Reload with Air
```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Start with hot reload
air -c .air.toml
# or
make dev
```

### Database Management
- **pgAdmin**: http://localhost:5050 (admin@shield.local / admin123)
- **Redis Commander**: http://localhost:8082

### API Documentation
- **Swagger UI**: http://localhost:8081/swagger/index.html
- Generate docs: `make docs`

## 🚨 Troubleshooting

### Common Issues

1. **Docker services not starting**
   ```bash
   # Check Docker is running
   docker info
   
   # Check logs
   docker-compose logs
   ```

2. **Environment variables not loading**
   ```bash
   # Verify .env file exists
   ls -la .env*
   
   # Check file content
   cat .env
   ```

3. **Database connection errors**
   ```bash
   # Check PostgreSQL is ready
   docker-compose exec postgres pg_isready -U shield
   
   # Check connection details
   echo $DB_HOST $DB_PORT $DB_USER
   ```

### Reset Development Environment
```bash
# Stop all services
make docker-down

# Clean up
make clean

# Remove volumes (⚠️ This will delete data)
docker-compose down -v

# Start fresh
./dev-setup.sh
```

## 📚 Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Air Hot Reload Tool](https://github.com/cosmtrek/air)
- [Viper Configuration Management](https://github.com/spf13/viper)
- [AWS Cognito Documentation](https://docs.aws.amazon.com/cognito/)

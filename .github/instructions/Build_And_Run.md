# 🚀 Building and Running the Project

## 📋 Prerequisites

| Requirement | Version  | Purpose                    |
|------------|----------|----------------------------|
| Go         | ≥ 1.22.5 | Backend development       |
| Docker     | ≥ 24.0.0 | Containerization          |
| Make       | ≥ 4.0    | Build automation          |
| Air        | latest   | Hot reload for development|

## 🔧 Environment Setup

1. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Required environment variables:**
   ```env
   # Database
   MASTER_DB_USER=postgres
   MASTER_DB_PASSWORD=your_password
   MASTER_DB_NAME=identity_manager

   # Server
   SERVER_PORT=8001
   DEBUGGER_PORT=2345

   # AWS Cognito
   COGNITO_REGION=us-east-1
   COGNITO_USER_POOL_ID=your_pool_id
   COGNITO_CLIENT_ID=your_client_id
   COGNITO_CLIENT_SECRET=your_client_secret
   
   # OPA
   OPA_SERVER_URL=http://localhost:8181

   # Auth
   JWT_SECRET=your_jwt_secret
   PASSWORD_SALT=your_password_salt
   ```

## 🏗️ Development Mode

```bash
# Start development environment with hot reload
make dev

# Development endpoints:
- API Server: http://localhost:8001
- PgAdmin: http://localhost:5050
- Debugger Port: 2345

# Auth endpoints:
- POST /api/v1/auth/signup - User registration
- POST /api/v1/auth/verify - Email verification
```# 🚀 Building and Running the Project

## 📋 Prerequisites

| Requirement | Version  | Purpose                    |
|------------|----------|----------------------------|
| Go         | ≥ 1.22.5 | Backend development       |
| Docker     | ≥ 24.0.0 | Containerization          |
| Make       | ≥ 4.0    | Build automation          |
| Air        | latest   | Hot reload for development|

## 🔧 Environment Setup

1. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Required environment variables:**
   ```env
   # Database
   MASTER_DB_USER=postgres
   MASTER_DB_PASSWORD=your_password
   MASTER_DB_NAME=identity_manager

   # Server
   SERVER_PORT=8001
   DEBUGGER_PORT=2345

   # AWS Cognito
   COGNITO_REGION=us-east-1
   COGNITO_USER_POOL_ID=your_pool_id
   COGNITO_CLIENT_ID=your_client_id
   COGNITO_CLIENT_SECRET=your_client_secret # Important for some Cognito operations
   
   # OPA
   OPA_SERVER_URL=http://localhost:8181
   ```

## 🏗️ Development Mode

**Note:** Modules like `authn` and `authz` are integrated into the main Shield Platform API. They do not have separate build or run commands; they are built and run as part of the main application using the commands below.

```bash
# Start development environment with hot reload
make dev

# Development endpoints:
- API Server: http://localhost:8001
- PgAdmin: http://localhost:5050
- Debugger Port: 2345
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/auth/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📦 Production Build

```bash
# Build production Docker image
make production

# Or build without starting
make build
```

## 🔍 Debugging

1. **VS Code Configuration:**
   ```json
   {
     "version": "0.2.0",
     "configurations": [
       {
         "name": "Remote Debug",
         "type": "go",
         "request": "attach",
         "mode": "remote",
         "remotePath": "/app",
         "port": 2345,
         "host": "127.0.0.1"
       }
     ]
   }
   ```

2. **Delve is already configured in development container**

## 📊 Monitoring

1. **Metrics endpoint:**
   ```
   GET /metrics
   ```

2. **Health check:**
   ```
   GET /health
   ```

## 🗄️ Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir db/migrations -seq add_new_table

# Run migrations
migrate -path db/migrations -database "postgres://user:pass@localhost:5432/db?sslmode=disable" up

# Rollback
migrate -path db/migrations -database "postgres://user:pass@localhost:5432/db?sslmode=disable" down
```

## 🔐 Setting up Auth Providers

1. **AWS Cognito:**
   ```bash
   # Configure AWS credentials
   aws configure

   # Create User Pool (if needed)
   aws cognito-idp create-user-pool --pool-name YourPoolName
   ```

2. **Other Providers:**
   - Follow instructions in `docs/auth-providers/`
   - Update provider configuration in `.env`

## 📝 Logs

- Development logs are streamed to console
- Production logs use structured JSON format
- Log levels: DEBUG, INFO, WARN, ERROR

## 🧹 Cleanup

```bash
# Remove all containers and volumes
make clean

# Remove specific resources
docker-compose -f docker-compose-dev.yml down -v
```

## 🚨 Common Issues

1. **Hot reload not working:**
   - Check `air.toml` configuration
   - Ensure volume mounts are correct

2. **Database connection issues:**
   - Verify PostgreSQL is running: `docker ps`
   - Check connection string in `.env`
   - Try connecting via pgAdmin

3. **AWS Cognito errors:**
   - Verify AWS credentials
   - Check User Pool configuration
   - Ensure Cognito endpoints are accessible

## 🔄 CI/CD Pipeline

- GitHub Actions workflow included
- Automated tests on PR
- Production deployment on main branch merge

For more detailed documentation, see:
- `/docs/architecture/`
- `/docs/api/`
- `/docs/deployment/`

# Load environment variables from .env file
ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: help build build-container build-container-pretoken start build-start start-watch clean docs test-coverage lint tidy

help:
	@echo ''
	@echo 'Usage: make [TARGET]'
	@echo 'Targets:'
	@echo '  build                    - Build Go app binary and generate swagger docs'
	@echo '  build-container          - Build Docker image for main app'
	@echo '  build-container-pretoken - Build Docker image for pretoken'
	@echo '  start                    - Run docker compose'
	@echo '  build-start              - Build and run docker compose'
	@echo '  start-watch              - Run docker compose in watch mode'
	@echo '  docs                     - Generate swagger documentation'
	@echo '  test-coverage            - Run tests with coverage report'
	@echo '  lint                     - Run golangci-lint on codebase'
	@echo '  tidy                     - Tidy and verify Go dependencies'
	@echo '  clean                    - Clean up docker containers and volumes'
	@echo ''

# Build Go binary and generate swagger documentation
build: docs
	@echo "Building Shield API binary..."
	go build -o shield-app ./cmd/app
	@echo "✓ Build complete: shield-app"

# Generate swagger documentation
docs:
	@echo "Generating swagger docs from cmd/app..."
	cd cmd/app && swag init -g main.go -o ../../docs --parseDependency
	@echo "✓ Swagger docs generated in ./docs/"

# Build Docker image for main app
build-container:
	@echo "Building Docker image for Shield API..."
	docker build -f Dockerfile -t shield-api:latest .
	@echo "✓ Docker image built: shield-api:latest"

# Build Docker image for pretoken
build-container-pretoken:
	@echo "Building Docker image for Pretoken service..."
	docker build -f cmd/pretoken/Dockerfile -t shield-pretoken:latest ./cmd/pretoken
	@echo "✓ Docker image built: shield-pretoken:latest"

# Run docker compose
start:
	@echo "Starting services with docker compose..."
	if [ ! -f .env ]; then cp .env.example .env; fi
	docker-compose -f docker-compose-dev.yml up

# Build and run docker compose
build-start:
	@echo "Building and starting services with docker compose..."
	if [ ! -f .env ]; then cp .env.example .env; fi
	docker-compose -f docker-compose-dev.yml up --build

# Run docker compose in watch mode (detached)
start-watch:
	@echo "Starting services in watch mode..."
	if [ ! -f .env ]; then cp .env.example .env; fi
	docker-compose -f docker-compose-dev.yml up -d
	@echo "✓ Services running in background. Use 'docker-compose logs -f' to view logs"

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	@echo "Generating coverage HTML report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Run golangci-lint on codebase
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...
	@echo "✓ Linting complete"

# Tidy and verify Go dependencies
tidy:
	@echo "Tidying Go dependencies..."
	go mod tidy
	@echo "Verifying Go dependencies..."
	go mod verify
	@echo "✓ Dependencies tidy and verified"

# Clean up all containers and volumes
clean:
	@echo "Cleaning up Docker containers and volumes..."
	docker-compose -f docker-compose-prod.yml down -v
	docker-compose -f docker-compose-dev.yml down -v
	@echo "✓ Cleanup complete"

.PHONY: help build run-collector run-processor run-api clean migrate migrate-status docker-up docker-down docker-logs

# Default target
help:
	@echo "Log Analytics System - Available commands:"
	@echo "  make docker-up       - Start Docker infrastructure (Kafka, MySQL, Zookeeper)"
	@echo "  make docker-down     - Stop Docker infrastructure"
	@echo "  make docker-logs     - Show Docker container logs"
	@echo "  make migrate         - Run all database migrations"
	@echo "  make migrate-status  - Show migration status"
	@echo "  make run-collector   - Run the log collector (generates sample logs)"
	@echo "  make run-processor   - Run the log processor (consumes from Kafka)"
	@echo "  make run-api         - Run the API server and dashboard"
	@echo "  make build           - Build all Go binaries"
	@echo "  make clean           - Clean build artifacts"
	@echo ""
	@echo "Quick Start:"
	@echo "  1. make docker-up"
	@echo "  2. make migrate"
	@echo "  3. make run-processor (in one terminal)"
	@echo "  4. make run-collector (in another terminal)"
	@echo "  5. make run-api (in another terminal)"
	@echo ""
	@echo "Access Dashboard: http://localhost:8080"
	@echo "Access Kafka UI: http://localhost:8081 (Kafka management)"

# Build all binaries
build:
	@echo "Building Go binaries..."
	go build -o bin/log-collector cmd/log-collector/main.go
	go build -o bin/log-processor cmd/log-processor/main.go
	go build -o bin/api-server cmd/api-server/main.go
	go build -o bin/migration cmd/migration/main.go
	@echo "Build complete!"

# Run all database migrations
migrate: build
	@echo "Running database migrations..."
	./bin/migration run
	@echo "Migrations completed!"

# Show migration status
migrate-status: build
	@echo "Showing migration status..."
	./bin/migration status

# Run log collector
run-collector: build
	@echo "Starting log collector..."
	./bin/log-collector

# Run log processor
run-processor: build
	@echo "Starting log processor..."
	./bin/log-processor

# Run API server
run-api: build
	@echo "Starting API server..."
	./bin/api-server

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	@echo "Clean complete!"

# Docker commands
docker-up:
	@echo "Starting Docker infrastructure..."
	docker-compose up -d
	@echo "Docker infrastructure started!"
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services should be ready now."

docker-down:
	@echo "Stopping Docker infrastructure..."
	docker-compose down
	@echo "Docker infrastructure stopped!"

docker-logs:
	@echo "Showing Docker container logs..."
	docker-compose logs -f

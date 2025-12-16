.PHONY: help install dev migrate build docker-up docker-down test clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install Go dependencies
	go mod download
	go mod tidy

dev: ## Run the API server locally (requires PostgreSQL and Redis)
	go run cmd/api/main.go

migrate: ## Run database migrations
	go run cmd/migrate/main.go

build: ## Build the API binary
	go build -o bin/api cmd/api/main.go
	go build -o bin/migrate cmd/migrate/main.go

docker-up: ## Start all services with Docker Compose
	cd docker && docker-compose up -d

docker-down: ## Stop all Docker services
	cd docker && docker-compose down

docker-logs: ## View Docker logs
	cd docker && docker-compose logs -f

test: ## Run tests
	go test ./... -v -cover

test-coverage: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...

.env: ## Create .env file from example
	cp .env.example .env
	@echo "Created .env file. Please update with your configuration."

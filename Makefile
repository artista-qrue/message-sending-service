
.PHONY: build run test clean docker-build docker-up docker-down swagger-gen help

APP_NAME=message-sending-service
DOCKER_COMPOSE=docker-compose
GO=go

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	$(GO) build -o bin/$(APP_NAME) ./cmd/server

run: ## Run the application locally
	$(GO) run ./cmd/server

test: ## Run tests
	$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):latest .

docker-up: ## Start all services with Docker Compose
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all services
	$(DOCKER_COMPOSE) down

docker-logs: ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-restart: ## Restart the main service
	$(DOCKER_COMPOSE) restart message-service

db-migrate: ## Run database migrations
	$(DOCKER_COMPOSE) exec postgres psql -U insider -d insider_messaging -f /docker-entrypoint-initdb.d/init.sql

db-reset: ## Reset database
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d postgres
	sleep 5
	$(DOCKER_COMPOSE) up -d

setup: ## Setup development environment
	$(GO) mod download
	$(DOCKER_COMPOSE) up -d postgres redis

test-api: ## Test API endpoints
	@echo "Testing health endpoint..."
	curl -f http://localhost:8080/health || echo "❌ Health check failed"
	@echo "\nTesting scheduler status..."
	curl -f http://localhost:8080/api/v1/scheduler/status || echo "❌ Scheduler status failed"

swagger-gen: ## Generate Swagger documentation
	swag init -g cmd/server/main.go -o docs/

lint: ## Run linter
	golangci-lint run

fmt: ## Format Go code
	$(GO) fmt ./...

deps: ## Install dependencies
	$(GO) mod download
	$(GO) mod tidy

deploy: docker-build docker-up ## Build and deploy with Docker

dev: setup build run ## Setup, build and run for development

ci: deps fmt lint test ## Run CI pipeline

logs: ## Show application logs
	$(DOCKER_COMPOSE) logs -f message-service

stats: ## Show container stats
	docker stats

clean-docker: ## Clean Docker resources
	docker system prune -f
	docker volume prune -f

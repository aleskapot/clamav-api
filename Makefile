.PHONY: help build run test test-cover test-race clean fmt lint tidy docker-build docker-run docker-stop docker-compose-up docker-compose-down

# Variables
BINARY_NAME=clamav-api
BUILD_DIR=./bin
CONFIG_PATH?=$(shell pwd)/configs/config.yaml
DOCKER_IMAGE=clamav-api
DOCKER_PORT=8080

help: ## Show this help message
	@echo "ClamAV API - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

build-debug: ## Build with debug symbols
	@echo "Building $(BINARY_NAME) (debug)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: ## Run the server
	@echo "Starting server..."
	CONFIG_PATH=$(CONFIG_PATH) go run ./cmd/server

run-config: ## Run with specific config file
	@CONFIG_PATH=$(CONFIG_PATH) go run ./cmd/server

# Testing
test: ## Run all tests
	go test ./tests/... -v

test-cover: ## Run tests with coverage report
	go test ./tests/... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-race: ## Run tests with race detector
	go test ./tests/... -v -race

test-unit: ## Run only unit tests (no integration)
	go test ./tests/... -v -short

# Code Quality
fmt: ## Format code
	go fmt ./...

tidy: ## Tidy dependencies
	go mod tidy

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

check: fmt tidy vet ## Run all code checks

# Cleanup
clean: ## Remove build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Docker
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run -d --name $(BINARY_NAME) \
		-p $(DOCKER_PORT):8080 \
		-v $(CONFIG_PATH):/etc/clamav-api/config.yaml:ro \
		$(DOCKER_IMAGE)

docker-stop: ## Stop Docker container
	docker stop $(BINARY_NAME) || true
	docker rm $(BINARY_NAME) || true

docker-clean: docker-stop ## Stop and remove container

# Development helpers
dev: tidy run ## Tidy deps and run server

restart: clean build run ## Clean, build and run

# OpenAPI
openapi-validate: ## Validate OpenAPI spec (requires redocly or swagger CLI)
	@echo "Validating openapi.yaml..."
	@if command -v redocly &> /dev/null; then \
		redocly lint openapi.yaml; \
	elif command -v swagger &> /dev/null; then \
		swagger validate openapi.yaml; \
	else \
		echo "Install redocly-cli or swagger-cli to validate"; \
	fi

# ClamAV (for local development)
run-clamav: ## Run ClamAV in Docker for local testing
	docker run -d --name clamav \
		-p 3310:3310 \
		--health-cmd="clamdcheck" \
		--health-interval=30s \
		mwader/clamav

stop-clamav: ## Stop ClamAV container
	docker stop clamav || true
	docker rm clamav || true

# Server info
version: ## Show Go version
	@go version

deps: ## Show dependencies
	go mod graph | head -20

# Default target
all: check build

# Docker Compose
docker-compose-up: ## Start all services with docker-compose
	docker-compose up -d
	@echo "Services started:"
	@docker-compose ps

docker-compose-down: ## Stop all services
	docker-compose down

docker-compose-logs: ## Show logs from all services
	docker-compose logs -f

docker-compose-restart: ## Restart all services
	docker-compose restart

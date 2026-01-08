.PHONY: all build test test-unit test-integration test-standalone test-cluster test-sentinel lint vet fmt clean
.PHONY: docker-standalone docker-cluster docker-sentinel docker-stop-all
.PHONY: help coverage

# Load .env file if it exists
-include .env
export

# Default port if not set
FALKORDB_PORT ?= 6379
FALKORDB_HOST ?= localhost

# Default target
all: build test-unit

# Build the library
build:
	go build ./...

# Run unit tests only (no FalkorDB required)
test-unit:
	go test -v -race ./internal/...

# Alias for unit tests
test: test-unit

# Run all tests including integration (requires FalkorDB)
test-all: test-unit test-integration

# Run tests with coverage
coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint the code
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Run go vet
vet:
	go vet ./...

# Format the code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	go clean
	rm -f coverage.out coverage.html

# === Docker Commands ===

# Start standalone FalkorDB
docker-standalone:
	docker compose -f docker/standalone-compose.yml up -d
	@echo "Waiting for FalkorDB to be ready..."
	@sleep 3
	@docker exec falkordb-standalone redis-cli ping || (echo "FalkorDB not ready, waiting..." && sleep 5)
	@echo "FalkorDB standalone is ready on localhost:$(FALKORDB_PORT)"

# Stop standalone FalkorDB
docker-standalone-stop:
	docker compose -f docker/standalone-compose.yml down

# Start FalkorDB cluster
docker-cluster:
	docker compose -f docker/cluster-compose.yml up -d
	@echo "Waiting for cluster to be ready..."
	@sleep 20
	@echo "FalkorDB cluster is ready on localhost:17000-17005"

# Stop FalkorDB cluster
docker-cluster-stop:
	docker compose -f docker/cluster-compose.yml down

# Start FalkorDB sentinel
docker-sentinel:
	docker compose -f docker/sentinel-compose.yml up -d
	@echo "Waiting for sentinel to be ready..."
	@sleep 10
	@echo "FalkorDB sentinel is ready on localhost:26379"

# Stop FalkorDB sentinel
docker-sentinel-stop:
	docker compose -f docker/sentinel-compose.yml down

# Stop all docker containers
docker-stop-all: docker-standalone-stop docker-cluster-stop docker-sentinel-stop

# === Test Commands with Docker ===

# Run integration tests (standalone mode)
test-integration: docker-standalone
	FALKORDB_HOST=$(FALKORDB_HOST) FALKORDB_PORT=$(FALKORDB_PORT) go test -v -race ./tests/integration/...
	$(MAKE) docker-standalone-stop

# Run standalone tests (alias)
test-standalone: test-integration

# Run cluster tests
test-cluster: docker-cluster
	FALKORDB_HOST=localhost FALKORDB_PORT=17003 go test -v -race -run TestCluster ./tests/integration/...
	$(MAKE) docker-cluster-stop

# Run sentinel tests
test-sentinel: docker-sentinel
	SENTINEL_HOST=localhost SENTINEL_PORT=26379 go test -v -race -run TestSentinel ./tests/integration/...
	$(MAKE) docker-sentinel-stop

# === Example Commands ===

# Run the basic example
example: docker-standalone
	@echo "Running basic example..."
	cd examples/basic && FALKORDB_HOST=$(FALKORDB_HOST) FALKORDB_PORT=$(FALKORDB_PORT) go run main.go
	$(MAKE) docker-standalone-stop

# === Help ===

help:
	@echo "FalkorDB Go Client - Available targets:"
	@echo ""
	@echo "Environment Variables (or set in .env file):"
	@echo "  FALKORDB_HOST - Host address (default: localhost)"
	@echo "  FALKORDB_PORT - Port number (default: 6379)"
	@echo ""
	@echo "Build & Test:"
	@echo "  make build           - Build the library"
	@echo "  make test            - Run unit tests (no FalkorDB required)"
	@echo "  make test-unit       - Run unit tests (no FalkorDB required)"
	@echo "  make test-integration- Run integration tests (starts Docker)"
	@echo "  make test-all        - Run all tests"
	@echo "  make coverage        - Run tests with coverage report"
	@echo "  make lint            - Run golangci-lint"
	@echo "  make vet             - Run go vet"
	@echo "  make fmt             - Format the code"
	@echo "  make clean           - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-standalone      - Start standalone FalkorDB"
	@echo "  make docker-standalone-stop - Stop standalone FalkorDB"
	@echo "  make docker-cluster         - Start FalkorDB cluster (ports 17000-17005)"
	@echo "  make docker-cluster-stop    - Stop FalkorDB cluster"
	@echo "  make docker-sentinel        - Start FalkorDB sentinel (port 26379)"
	@echo "  make docker-sentinel-stop   - Stop FalkorDB sentinel"
	@echo "  make docker-stop-all        - Stop all docker containers"
	@echo ""
	@echo "Tests with Docker:"
	@echo "  make test-standalone  - Run standalone tests (starts/stops Docker)"
	@echo "  make test-cluster     - Run cluster tests (starts/stops Docker)"
	@echo "  make test-sentinel    - Run sentinel tests (starts/stops Docker)"
	@echo "  make test-integration - Run all integration tests"
	@echo ""
	@echo "Examples:"
	@echo "  make example          - Run the basic example"
	@echo ""
	@echo "Usage:"
	@echo "  FALKORDB_PORT=16379 make test-integration  # Use custom port"
	@echo "  cp .env.example .env && make test-integration  # Use .env file"

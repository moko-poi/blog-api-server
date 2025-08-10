# Blog API Server Makefile
# Based on Go development best practices

# Variables
BINARY_NAME = blog-api-server
CMD_DIR = ./cmd/server
BUILD_DIR = ./build
COVERAGE_OUT = ./coverage.out

# Go related variables
GOBASE = $(shell pwd)
GOPATH = $(GOBASE)/vendor:$(GOBASE)
GOBIN = $(GOBASE)/bin
GOFILES = $(wildcard *.go)

# Default target
.DEFAULT_GOAL := help

# Help target - lists all available targets
help: ## Show this help message
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
.PHONY: dev
dev: ## Run the application in development mode with hot reload
	@echo "Starting development server with hot reload..."
	go run $(CMD_DIR)/main.go

.PHONY: build
build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

.PHONY: build-all
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/main.go
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/main.go

# Testing targets
.PHONY: test
test: ## Run all tests with race detection
	@echo "Running tests..."
	go test -v -race -buildvcs ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -race -buildvcs -coverprofile=$(COVERAGE_OUT) ./...
	go tool cover -html=$(COVERAGE_OUT) -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -race -buildvcs -tags=integration ./...

# Code quality targets
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	gofmt -l -w .

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run static analysis with golangci-lint

.PHONY: golangci-lint
golangci-lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix; \
	else \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run --fix; \
	fi

.PHONY: staticcheck
staticcheck: ## Run static analysis with staticcheck (legacy)
	@echo "Running static analysis with staticcheck..."
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...

.PHONY: audit
audit: fmt vet lint vuln test ## Run comprehensive code quality audit
	@echo "Running comprehensive audit..."
	go mod tidy -diff
	go mod verify
	@test -z "$$(gofmt -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

.PHONY: vuln
vuln: ## Check for vulnerabilities
	@echo "Checking for vulnerabilities..."
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Dependency management
.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Managing dependencies..."
	go mod download
	go mod tidy

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Clean targets
.PHONY: clean
clean: ## Clean build artifacts and test cache
	@echo "Cleaning..."
	go clean ./...
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_OUT) coverage.html

.PHONY: clean-deps
clean-deps: ## Clean dependency cache
	@echo "Cleaning dependency cache..."
	go clean -modcache

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(BINARY_NAME):latest

# Development environment
.PHONY: setup
setup: deps install-tools setup-hooks ## Initial project setup
	@echo "Setting up development environment..."
	@echo "✅ Dependencies installed"
	@echo "✅ Development tools installed"
	@echo "✅ Git hooks configured"
	@echo "✅ Run 'make dev' to start development server"
	@echo "✅ Run 'make test' to run tests"
	@echo "✅ Run 'make audit' for code quality checks"

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📦 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2; \
	fi
	@if ! command -v lefthook >/dev/null 2>&1; then \
		echo "📦 Installing lefthook..."; \
		go install github.com/evilmartians/lefthook@v1.10.0; \
	fi
	@echo "✅ Development tools installed"

.PHONY: setup-hooks
setup-hooks: ## Setup Git hooks with lefthook
	@echo "Setting up Git hooks..."
	@if command -v lefthook >/dev/null 2>&1; then \
		lefthook install; \
		echo "✅ Git hooks configured with lefthook"; \
	else \
		echo "⚠️  lefthook not found. Run 'make install-tools' first"; \
	fi

.PHONY: hooks-run
hooks-run: ## Run all pre-commit hooks manually
	@echo "Running pre-commit hooks..."
	@if command -v lefthook >/dev/null 2>&1; then \
		lefthook run pre-commit; \
	else \
		echo "❌ lefthook not installed. Run 'make install-tools' first"; \
		exit 1; \
	fi

# Check if required tools are installed
.PHONY: check-tools
check-tools: ## Check if required development tools are installed
	@echo "Checking development tools..."
	@command -v go >/dev/null 2>&1 || (echo "Go is not installed" && exit 1)
	@command -v docker >/dev/null 2>&1 || echo "⚠️  Docker is not installed (optional)"
	@echo "✅ Required tools are available"
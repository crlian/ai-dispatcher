.PHONY: build test clean install build-all help

# Binary name
BINARY_NAME=ai-dispatcher

# Version from git
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Build directory
BUILD_DIR=bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)

	# macOS (Intel)
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .

	# macOS (Apple Silicon)
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

	# Linux (amd64)
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

	# Linux (arm64)
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .

	# Windows (amd64)
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

	@echo "All builds complete!"
	@ls -lh $(BUILD_DIR)

test: ## Run all tests
	@echo "Running tests..."
	go test ./... -v -cover

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

install: build ## Install binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation complete!"

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	go mod tidy

run: build ## Build and run the binary
	@$(BUILD_DIR)/$(BINARY_NAME)

# NoPlaceLike - Professional Distributed Network Resource Sharing Platform
# Development Makefile

# Variables
BINARY_NAME=noplacelike
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Directories
BUILD_DIR=build
DIST_DIR=dist
DOCS_DIR=docs

# Default target
.PHONY: all
all: test build

# Help target
.PHONY: help
help:
	@echo "NoPlaceLike Development Commands"
	@echo "================================"
	@echo ""
	@echo "Building:"
	@echo "  build         Build the binary"
	@echo "  build-all     Build for all platforms"
	@echo "  build-linux   Build for Linux"
	@echo "  build-windows Build for Windows"
	@echo "  build-darwin  Build for macOS"
	@echo ""
	@echo "Development:"
	@echo "  dev           Start development server with live reload"
	@echo "  dev-setup     Install development dependencies"
	@echo "  run           Run the application locally"
	@echo "  clean         Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test          Run all tests"
	@echo "  test-unit     Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  test-race     Run tests with race detection"
	@echo "  benchmark     Run benchmarks"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt           Format Go code"
	@echo "  lint          Run linter"
	@echo "  vet           Run go vet"
	@echo "  security      Run security scanner"
	@echo "  check         Run all code quality checks"
	@echo ""
	@echo "Documentation:"
	@echo "  docs          Generate documentation"
	@echo "  docs-serve    Serve documentation locally"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-run    Run Docker container"
	@echo "  docker-push   Push Docker image"
	@echo ""
	@echo "Release:"
	@echo "  release       Create a new release"
	@echo "  package       Package for distribution"

# Build targets
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

.PHONY: build-all
build-all: build-linux build-windows build-darwin

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .

# Development targets
.PHONY: dev
dev:
	@echo "Starting development server..."
	@which air > /dev/null || (echo "Installing air for live reload..." && go install github.com/cosmtrek/air@latest)
	air

.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing development dependencies..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development environment ready!"

.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf coverage.out
	rm -rf *.log

# Testing targets
.PHONY: test
test:
	@echo "Running all tests..."
	$(GOTEST) -v ./...

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...

.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Code quality targets
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) -s -w .
	goimports -w .

.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

.PHONY: security
security:
	@echo "Running security scanner..."
	gosec ./...

.PHONY: check
check: fmt vet lint security test
	@echo "All code quality checks passed!"

# Documentation targets
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@mkdir -p $(DOCS_DIR)
	swag init -g main.go -o $(DOCS_DIR)/swagger
	@echo "Documentation generated in $(DOCS_DIR)/"

.PHONY: docs-serve
docs-serve: docs
	@echo "Serving documentation on http://localhost:6060"
	godoc -http=:6060

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t nathfavour/noplacelike:$(VERSION) .
	docker tag nathfavour/noplacelike:$(VERSION) nathfavour/noplacelike:latest

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 9090:9090 nathfavour/noplacelike:latest

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	docker push nathfavour/noplacelike:$(VERSION)
	docker push nathfavour/noplacelike:latest

# Release targets
.PHONY: release
release: check build-all
	@echo "Creating release $(VERSION)..."
	@mkdir -p $(DIST_DIR)
	# Create checksums
	cd $(DIST_DIR) && sha256sum * > checksums.txt
	@echo "Release $(VERSION) created in $(DIST_DIR)/"

.PHONY: package
package: release
	@echo "Packaging for distribution..."
	@mkdir -p $(DIST_DIR)/packages
	# Create tar.gz packages
	cd $(DIST_DIR) && \
	tar -czf packages/$(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf packages/$(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf packages/$(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar -czf packages/$(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-arm64
	# Create zip for Windows
	cd $(DIST_DIR) && \
	zip packages/$(BINARY_NAME)-windows-amd64-$(VERSION).zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Packages created in $(DIST_DIR)/packages/"

# Dependency management
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

.PHONY: deps-verify
deps-verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Installation targets
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

# Utility targets
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"

.PHONY: info
info:
	@echo "Project Information"
	@echo "=================="
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"
	@echo "Git Commit: $(shell git rev-parse HEAD)"
	@echo "Git Branch: $(shell git rev-parse --abbrev-ref HEAD)"
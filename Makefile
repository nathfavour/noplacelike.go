# NoPlaceLike 2.0 Build System
# Professional Go build automation with comprehensive tooling

# Build variables
BINARY_NAME=noplacelike
PACKAGE=github.com/nathfavour/noplacelike.go
BUILD_DIR=build
DIST_DIR=dist
DOCKER_IMAGE=nathfavour/noplacelike
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT) -s -w"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
    OS = linux
endif
ifeq ($(UNAME_S),Darwin)
    OS = darwin
endif
ifeq ($(UNAME_S),Windows)
    OS = windows
endif

ifeq ($(UNAME_M),x86_64)
    ARCH = amd64
endif
ifeq ($(UNAME_M),aarch64)
    ARCH = arm64
endif
ifeq ($(UNAME_M),arm64)
    ARCH = arm64
endif

# Default target
.PHONY: all
all: clean deps lint test build

# Help target
.PHONY: help
help:
	@echo "NoPlaceLike 2.0 Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary for current platform"
	@echo "  build-all      - Build binaries for all platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  test-race      - Run tests with race detection"
	@echo "  lint           - Run linters"
	@echo "  fmt            - Format code"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-push    - Push Docker image"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  dev            - Start development server"
	@echo "  docs           - Generate documentation"
	@echo "  release        - Create release package"

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf coverage.out
	rm -rf coverage.html

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) tidy
	$(GOMOD) download

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) -s -w .

# Run linters
.PHONY: lint
lint:
	$(GOLINT) run ./...

# Install development tools
.PHONY: install-tools
install-tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/goreleaser/goreleaser@latest
	$(GOGET) github.com/swaggo/swag/cmd/swag@latest

# Build for current platform
.PHONY: build
build: deps
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/noplacelike

# Build for all platforms
.PHONY: build-all
build-all: deps
	mkdir -p $(BUILD_DIR)
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/noplacelike
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/noplacelike
	# macOS
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/noplacelike
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/noplacelike
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/noplacelike

# Install binary
.PHONY: install
install: build
	$(GOCMD) install $(LDFLAGS) ./cmd/noplacelike

# Development server
.PHONY: dev
dev: build
	./$(BUILD_DIR)/$(BINARY_NAME) --log-level debug --enable-profiling

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
.PHONY: test-race
test-race:
	$(GOTEST) -v -race ./...

# Run integration tests
.PHONY: test-integration
test-integration:
	$(GOTEST) -v -tags=integration ./tests/integration/...

# Run load tests
.PHONY: test-load
test-load:
	$(GOTEST) -v -tags=load ./tests/load/...

# Run security tests
.PHONY: test-security
test-security:
	$(GOTEST) -v -tags=security ./tests/security/...

# Benchmark tests
.PHONY: bench
bench:
	$(GOTEST) -bench=. -benchmem ./...

# Docker build
.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Docker push
.PHONY: docker-push
docker-push: docker-build
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

# Generate documentation
.PHONY: docs
docs:
	swag init -g cmd/noplacelike/main.go -o docs/

# Create release package
.PHONY: release
release: clean build-all
	mkdir -p $(DIST_DIR)
	# Create archives for each platform
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-arm64
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-amd64
	tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64
	zip -j $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	# Generate checksums
	cd $(DIST_DIR) && sha256sum * > checksums.txt
	@echo "Release packages created in $(DIST_DIR)/"

# Create Docker Compose environment
.PHONY: docker-compose-up
docker-compose-up:
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down:
	docker-compose down

# Kubernetes deployment
.PHONY: k8s-deploy
k8s-deploy:
	kubectl apply -f deployments/kubernetes/

.PHONY: k8s-delete
k8s-delete:
	kubectl delete -f deployments/kubernetes/

# Performance profiling
.PHONY: profile-cpu
profile-cpu:
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...
	$(GOCMD) tool pprof cpu.prof

.PHONY: profile-mem
profile-mem:
	$(GOTEST) -memprofile=mem.prof -bench=. ./...
	$(GOCMD) tool pprof mem.prof

# Code quality checks
.PHONY: quality
quality: fmt lint test-coverage
	@echo "Code quality checks completed"

# Security scan
.PHONY: security-scan
security-scan:
	gosec ./...

# Dependency vulnerability check
.PHONY: vuln-check
vuln-check:
	$(GOCMD) list -json -deps ./... | nancy sleuth

# Generate mocks
.PHONY: mocks
mocks:
	mockgen -source=internal/core/interfaces.go -destination=internal/mocks/core_mocks.go

# Update dependencies
.PHONY: deps-update
deps-update:
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Cleanup development environment
.PHONY: dev-clean
dev-clean: clean
	docker system prune -f
	$(GOCMD) clean -modcache

# CI/CD pipeline simulation
.PHONY: ci
ci: deps lint test-race test-coverage build-all

# Show build information
.PHONY: info
info:
	@echo "Build Information:"
	@echo "  Binary:     $(BINARY_NAME)"
	@echo "  Version:    $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  OS/Arch:    $(OS)/$(ARCH)"
	@echo "  Go Version: $(shell $(GOCMD) version)"

# Watch for changes and rebuild
.PHONY: watch
watch:
	@which fswatch > /dev/null || (echo "fswatch not installed. Install with: brew install fswatch" && exit 1)
	fswatch -o . -e ".*" -i "\\.go$$" | xargs -n1 -I{} make build

# Performance monitoring
.PHONY: monitor
monitor:
	@echo "Starting performance monitoring..."
	@echo "Metrics available at: http://localhost:9090/metrics"
	@echo "Pprof available at: http://localhost:6060/debug/pprof/"

# Version information
.PHONY: version
version:
	@echo $(VERSION)
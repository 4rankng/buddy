# Oncall CLI Makefile
# Build and development automation for the payment team dashboard

# Variables
APP_NAME := oncall
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"
GO_FILES := $(shell find . -name "*.go" -type f)

# Default target
.PHONY: default
default: build

# Build the application
.PHONY: build
build:
	@echo "🔨 Building $(APP_NAME)..."
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/oncall
	@echo "✅ Build complete: bin/$(APP_NAME)"

# Build with debug information
.PHONY: build-debug
build-debug:
	@echo "🔨 Building $(APP_NAME) in debug mode..."
	go build -gcflags="all=-N -l" -o bin/$(APP_NAME)-debug ./cmd/oncall
	@echo "✅ Debug build complete: bin/$(APP_NAME)-debug"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	@echo "✅ Clean complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run the payment dashboard
.PHONY: run
run: build
	@echo "🚀 Starting payment team dashboard..."
	./bin/$(APP_NAME)

# Alias for run command
.PHONY: run-payment
run-payment: run

# Format code
.PHONY: fmt
fmt:
	@echo "💅 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Lint code (requires golangci-lint)
.PHONY: lint
lint:
	@echo "🔍 Linting code..."
	golangci-lint run
	@echo "✅ Linting complete"

# Security scan (requires gosec)
.PHONY: security
security:
	@echo "🔒 Running security scan..."
	gosec ./...
	@echo "✅ Security scan complete"

# Run all checks (fmt, test, lint)
.PHONY: check
check: fmt test lint
	@echo "✅ All checks passed"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p bin

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 ./cmd/oncall

	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-arm64 ./cmd/oncall

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 ./cmd/oncall

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-arm64 ./cmd/oncall

	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/oncall

	@echo "✅ Multi-platform builds complete:"
	@ls -la bin/

# Development mode - watch for changes and rebuild
.PHONY: dev
dev:
	@echo "👀 Starting development mode (watching for changes)..."
	@echo "Install 'air' for live reload: go install github.com/cosmtrek/air@latest"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "❌ 'air' not found. Install with: go install github.com/cosmtrek/air@latest"; \
	fi

# Generate documentation
.PHONY: docs
docs:
	@echo "📚 Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "📖 Starting documentation server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "❌ 'godoc' not found. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Create release package
.PHONY: release
release: clean build-all
	@echo "📦 Creating release package..."
	@mkdir -p release
	@cd bin && \
	for file in $(APP_NAME)-*; do \
		if [[ $$file == *.exe ]]; then \
			zip -r ../release/$${file%.exe}.zip $$file; \
		else \
			tar -czf ../release/$$file.tar.gz $$file; \
		fi; \
	done
	@echo "✅ Release packages created in release/ directory"
	@ls -la release/

# Docker build
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	@echo "✅ Docker image built: $(APP_NAME):$(VERSION)"

# Install to system PATH
.PHONY: install
install: build
	@echo "📦 Installing $(APP_NAME) to /usr/local/bin..."
	sudo cp bin/$(APP_NAME) /usr/local/bin/
	@echo "✅ Installation complete. Run '$(APP_NAME) payment' to start."

# Uninstall from system PATH
.PHONY: uninstall
uninstall:
	@echo "🗑️  Uninstalling $(APP_NAME) from /usr/local/bin..."
	sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "✅ Uninstallation complete."

# Show version information
.PHONY: version
version:
	@echo "📋 $(APP_NAME) Information:"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(shell go version)"
	@echo "  Git Commit: $(shell git rev-parse HEAD 2>/dev/null || echo 'unknown')"

# Help target
.PHONY: help
help:
	@echo "🚀 Oncall CLI Makefile"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          Build the application"
	@echo "  build-debug    Build with debug symbols"
	@echo "  build-all      Build for all platforms"
	@echo "  clean          Remove build artifacts"
	@echo ""
	@echo "Development Commands:"
	@echo "  run            Build and run payment dashboard"
	@echo "  run-payment    Build and run payment dashboard (alias)"
	@echo "  dev            Live reload development mode"
	@echo "  fmt            Format code"
	@echo "  lint           Lint code (requires golangci-lint)"
	@echo ""
	@echo "Testing Commands:"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  check          Run fmt, test, and lint"
	@echo ""
	@echo "System Commands:"
	@echo "  deps           Install dependencies"
	@echo "  install        Install to system PATH"
	@echo "  uninstall      Remove from system PATH"
	@echo "  version        Show version information"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build && ./bin/oncall           # Shows payment dashboard"
	@echo "  make run-payment                     # Build and run payment dashboard"
	@echo "  make dev"
.PHONY: build lint deps help

# Default target - builds for current platform
build:
	@echo "Building mybuddy and sgbuddy for current platform..."
	@mkdir -p bin
	@go build -o bin/mybuddy ./cmd/mybuddy || exit 1
	@go build -o bin/sgbuddy ./cmd/sgbuddy || exit 1
	@echo "Build complete. Binaries available in bin/ directory:"
	@ls -la bin/

# Run linters
lint:
	@echo "Running linters..."
	@echo "Running gofmt..."
	@if [ "$(shell gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "gofmt found issues:"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "Running go vet..."
	go vet ./...
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi
	@echo "Linting complete!"

# Install dependencies
deps:
	go mod download
	go mod tidy

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build both binaries for current platform"
	@echo "  lint       - Run Go linters (gofmt, go vet, golangci-lint)"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  help       - Show this help message"
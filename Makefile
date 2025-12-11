.PHONY: build build-all build-mac build-linux clean test

# Default target
build:
	@echo "Building mybuddy and sgbuddy..."
	@mkdir -p bin
	go build -o bin/mybuddy ./cmd/mybuddy
	go build -o bin/sgbuddy ./cmd/sgbuddy
	@echo "Build complete. Binaries available in bin/ directory."

# Build for multiple platforms
build-all: build-mac build-linux

# Build for macOS
build-mac:
	@echo "Building for macOS..."
	@mkdir -p bin/darwin-amd64 bin/darwin-arm64
	GOOS=darwin GOARCH=amd64 go build -o bin/darwin-amd64/mybuddy ./cmd/mybuddy
	GOOS=darwin GOARCH=amd64 go build -o bin/darwin-amd64/sgbuddy ./cmd/sgbuddy
	GOOS=darwin GOARCH=arm64 go build -o bin/darwin-arm64/mybuddy ./cmd/mybuddy
	GOOS=darwin GOARCH=arm64 go build -o bin/darwin-arm64/sgbuddy ./cmd/sgbuddy
	@echo "macOS builds complete."

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p bin/linux-amd64
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/mybuddy ./cmd/mybuddy
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/sgbuddy ./cmd/sgbuddy
	@echo "Linux builds complete."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	@echo "Clean complete."

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Quick dev build (current platform only)
dev:
	go build -o mybuddy ./cmd/mybuddy
	go build -o sgbuddy ./cmd/sgbuddy

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build both binaries for current platform"
	@echo "  build-all  - Build for all supported platforms"
	@echo "  build-mac  - Build for macOS (amd64 and arm64)"
	@echo "  build-linux- Build for Linux (amd64)"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  dev        - Quick build for development"
	@echo "  help       - Show this help message"
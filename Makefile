.PHONY: build build-my build-sg deploy lint deps test help

BIN_DIR := $(HOME)/bin

# Default target - builds both applications
build: build-my build-sg

# Build mybuddy with MY environment
build-my:
	@echo "Building mybuddy with Malaysia environment..."
	@mkdir -p bin
	@eval $$(grep -v '^#' .env.my | grep '=' | xargs) && \
	go build -ldflags "-X buddy/internal/buildinfo.JiraDomain=$$JIRA_DOMAIN \
	                   -X buddy/internal/buildinfo.JiraUsername=$$JIRA_USERNAME \
	                   -X buddy/internal/buildinfo.JiraApiKey=$$JIRA_API_KEY \
	                   -X buddy/internal/buildinfo.DoormanUsername=$$DOORMAN_USERNAME \
	                   -X buddy/internal/buildinfo.DoormanPassword=$$DOORMAN_PASSWORD \
	                   -X buddy/internal/buildinfo.BuildEnvironment=my" \
		-o bin/mybuddy ./cmd/mybuddy || exit 1
	@echo "mybuddy built successfully"

# Build sgbuddy with SG environment
build-sg:
	@echo "Building sgbuddy with Singapore environment..."
	@mkdir -p bin
	@eval $$(grep -v '^#' .env.sg | grep '=' | xargs) && \
	go build -ldflags "-X buddy/internal/buildinfo.JiraDomain=$$JIRA_DOMAIN \
	                   -X buddy/internal/buildinfo.JiraUsername=$$JIRA_USERNAME \
	                   -X buddy/internal/buildinfo.JiraApiKey=$$JIRA_API_KEY \
	                   -X buddy/internal/buildinfo.DoormanUsername=$$DOORMAN_USERNAME \
	                   -X buddy/internal/buildinfo.DoormanPassword=$$DOORMAN_PASSWORD \
	                   -X buddy/internal/buildinfo.BuildEnvironment=sg" \
		-o bin/sgbuddy ./cmd/sgbuddy || exit 1
	@echo "sgbuddy built successfully"

# Deploy binaries to user's bin directory
deploy: build
	@echo "Building and deploying binaries..."
	@$(MAKE) build
	@echo "Deployed to $(BIN_DIR)"
	@mkdir -p "$(BIN_DIR)"
	@mv -f bin/mybuddy "$(BIN_DIR)/mybuddy"
	@mv -f bin/sgbuddy "$(BIN_DIR)/sgbuddy"		
	@echo "You can now use 'mybuddy' and 'sgbuddy' commands from anywhere."

# Run linters
lint:
	@echo "Running linters..."
	@echo "Running gofmt..."
	@if [ "$$(gofmt -s -l . | grep -v '^vendor/' | wc -l)" -gt 0 ]; then \
		echo "gofmt found issues:"; \
		gofmt -s -l . | grep -v '^vendor/'; \
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

# Run all unit tests
test:
	@echo "Running unit tests..."
	go test -v ./...

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build both binaries with their respective environments"
	@echo "  build-my   - Build mybuddy with Malaysia environment (.env.my)"
	@echo "  build-sg   - Build sgbuddy with Singapore environment (.env.sg)"
	@echo "  deploy     - Build and install binaries to ~/.local/bin for system-wide use"
	@echo "  lint       - Run Go linters (gofmt, go vet, golangci-lint)"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  test       - Run all unit tests"
	@echo "  help       - Show this help message"
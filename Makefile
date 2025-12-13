.PHONY: build build-my build-sg lint deps help

# Default target - builds both applications
build: build-my build-sg

# Build mybuddy with MY environment
build-my:
	@echo "Building mybuddy with Malaysia environment..."
	@mkdir -p bin
	@eval $$(grep -v '^#' .env.my | grep '=' | xargs) && \
	go build -ldflags "-X buddy/internal/compiletime.JiraDomain=$$JIRA_DOMAIN \
	                   -X buddy/internal/compiletime.JiraUsername=$$JIRA_USERNAME \
	                   -X buddy/internal/compiletime.JiraApiKey=$$JIRA_API_KEY \
	                   -X buddy/internal/compiletime.DoormanUsername=$$DOORMAN_USERNAME \
	                   -X buddy/internal/compiletime.DoormanPassword=$$DOORMAN_PASSWORD \
	                   -X buddy/internal/compiletime.BuildEnvironment=my" \
		-o bin/mybuddy ./cmd/mybuddy || exit 1
	@echo "mybuddy built successfully"

# Build sgbuddy with SG environment
build-sg:
	@echo "Building sgbuddy with Singapore environment..."
	@mkdir -p bin
	@eval $$(grep -v '^#' .env.sg | grep '=' | xargs) && \
	go build -ldflags "-X buddy/internal/compiletime.JiraDomain=$$JIRA_DOMAIN \
	                   -X buddy/internal/compiletime.JiraUsername=$$JIRA_USERNAME \
	                   -X buddy/internal/compiletime.JiraApiKey=$$JIRA_API_KEY \
	                   -X buddy/internal/compiletime.DoormanUsername=$$DOORMAN_USERNAME \
	                   -X buddy/internal/compiletime.DoormanPassword=$$DOORMAN_PASSWORD \
	                   -X buddy/internal/compiletime.BuildEnvironment=sg" \
		-o bin/sgbuddy ./cmd/sgbuddy || exit 1
	@echo "sgbuddy built successfully"

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
	@echo "  build      - Build both binaries with their respective environments"
	@echo "  build-my   - Build mybuddy with Malaysia environment (.env.my)"
	@echo "  build-sg   - Build sgbuddy with Singapore environment (.env.sg)"
	@echo "  lint       - Run Go linters (gofmt, go vet, golangci-lint)"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  help       - Show this help message"
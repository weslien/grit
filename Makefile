.PHONY: build clean test lint fmt install deps help ci coverage bench integration docker

BINARY_NAME=grit
BINARY_PATH=./bin/$(BINARY_NAME)
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT = $(shell git rev-parse HEAD)
DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS = -ldflags="-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

# Default target
all: clean deps lint test build

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod verify

# Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	mkdir -p ./bin
	go build $(LDFLAGS) -o $(BINARY_PATH) main.go
	@echo "✅ Built $(BINARY_PATH)"

# Build for multiple platforms
build-all:
	@echo "🌍 Building for multiple platforms..."
	mkdir -p ./dist
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o ./dist/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o ./dist/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o ./dist/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o ./dist/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o ./dist/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "✅ Built binaries in ./dist/"

# Install locally
install: build
	@echo "📥 Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Installed $(BINARY_NAME)"

# Legacy install (go install)
install-go:
	@echo "Installing grit..."
	@go install
	@echo "✅"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	gofmt -s -w .
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .
	@echo "✅ Code formatted"

# Run linter
lint:
	@echo "🔍 Running linter..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run
	@echo "✅ Linting passed"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v -race ./...

# Legacy test with coverage
test-legacy:
	go test -v -cover ./...

# Run tests with coverage
coverage:
	@echo "📊 Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out
	@echo "📊 Coverage report generated: coverage.html"

# Run benchmarks
bench:
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem ./...

# Run integration tests
integration: build
	@echo "🔗 Running integration tests..."
	./$(BINARY_PATH) --help
	./$(BINARY_PATH) --version
	./$(BINARY_PATH) completion bash > /dev/null
	@echo "✅ Integration tests passed"

# Complete CI checks (run locally)
ci: deps fmt lint test coverage integration
	@echo "🎉 All CI checks passed!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning up..."
	rm -rf ./bin ./dist ./coverage.out ./coverage.html
	go clean -cache -testcache
	@echo "✅ Cleaned up"

# Legacy clean
clean-legacy:
	rm -rf bin

# Generate completion scripts
completions: build
	@echo "🔧 Generating completion scripts..."
	mkdir -p ./completions
	./$(BINARY_PATH) completion bash > ./completions/$(BINARY_NAME).bash
	./$(BINARY_PATH) completion zsh > ./completions/_$(BINARY_NAME)
	./$(BINARY_PATH) completion fish > ./completions/$(BINARY_NAME).fish
	./$(BINARY_PATH) completion powershell > ./completions/$(BINARY_NAME).ps1
	@echo "✅ Completion scripts generated in ./completions/"

# Build Docker image
docker:
	@echo "🐳 Building Docker image..."
	docker build -t grit:$(VERSION) .
	docker tag grit:$(VERSION) grit:latest
	@echo "✅ Docker image built: grit:$(VERSION)"

# Run security checks
security:
	@echo "🔒 Running security checks..."
	go install github.com/securecodewarrior/github-action-gosec/cmd/gosec@latest
	gosec ./...
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...
	@echo "✅ Security checks passed"

# Development setup
dev-setup:
	@echo "🛠️  Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securecodewarrior/github-action-gosec/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "✅ Development tools installed"

# Uninstall
uninstall:
	rm -f $$(shell go env GOPATH)/bin/grit

# Generate schema (legacy)
schema:
	@echo "Generating grit schema..."
	go run cmd/schema/main.go

# Show version
version:
	@go version

# Display help
help:
	@echo "🚀 Grit Development Commands"
	@echo ""
	@echo "Build Commands:"
	@echo "  build      - Build the binary"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  install    - Install binary to /usr/local/bin"
	@echo "  install-go - Install using go install"
	@echo ""
	@echo "Quality Commands:"
	@echo "  fmt        - Format code (gofmt + goimports)"
	@echo "  lint       - Run golangci-lint"
	@echo "  test       - Run unit tests"
	@echo "  coverage   - Run tests with coverage report"
	@echo "  bench      - Run benchmarks"
	@echo "  security   - Run security checks"
	@echo ""
	@echo "CI/CD Commands:"
	@echo "  ci         - Run all CI checks locally"
	@echo "  integration- Run integration tests"
	@echo ""
	@echo "Utility Commands:"
	@echo "  deps       - Install dependencies"
	@echo "  clean      - Clean build artifacts"
	@echo "  completions- Generate shell completions"
	@echo "  docker     - Build Docker image"
	@echo "  dev-setup  - Install development tools"
	@echo ""
	@echo "Examples:"
	@echo "  make all          # Clean, deps, lint, test, build"
	@echo "  make ci           # Run full CI pipeline locally"
	@echo "  make build-all    # Build for all platforms"

# Prevent make from treating file names as targets
$(BINARY_PATH): build
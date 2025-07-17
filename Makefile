# Cheesebox Makefile

# Variables
BINARY_NAME=cx
BINARY_PATH=./$(BINARY_NAME)
INSTALL_PATH=/usr/local/bin/$(BINARY_NAME)
GO_FILES=$(shell find . -name "*.go")

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build: $(BINARY_PATH)

$(BINARY_PATH): $(GO_FILES) go.mod go.sum
	@echo "🔨 Building Cheesebox..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) .
	@echo "✅ Build complete: $(BINARY_PATH)"

# Install dependencies
.PHONY: deps
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

# Install the binary system-wide
.PHONY: install
install: build
	@echo "📦 Installing Cheesebox to $(INSTALL_PATH)..."
	sudo cp $(BINARY_PATH) $(INSTALL_PATH)
	@echo "✅ Cheesebox installed! Run 'cx' to get started."

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "🗑️  Uninstalling Cheesebox..."
	sudo rm -f $(INSTALL_PATH)
	@echo "✅ Cheesebox uninstalled"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test ./...
	@echo "✅ Tests complete"

# Run with race detection
.PHONY: test-race
test-race:
	@echo "🧪 Running tests with race detection..."
	go test -race ./...
	@echo "✅ Race tests complete"

# Format code
.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Run linter
.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	golangci-lint run
	@echo "✅ Linting complete"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "✅ Clean complete"

# Development build with debugging
.PHONY: dev
dev:
	@echo "🔨 Building for development..."
	go build -o $(BINARY_NAME) .
	@echo "✅ Development build complete"

# Run the application
.PHONY: run
run: build
	./$(BINARY_NAME)

# Create a release build
.PHONY: release
release:
	@echo "📦 Creating release build..."
	CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always)" -o $(BINARY_NAME) .
	@echo "✅ Release build complete"

# Cross-compile for different platforms
.PHONY: build-all
build-all:
	@echo "🔨 Cross-compiling for multiple platforms..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o dist/cx-linux-amd64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o dist/cx-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o dist/cx-darwin-arm64 .
	@echo "✅ Cross-compilation complete"

# Check for Ollama
.PHONY: check-ollama
check-ollama:
	@echo "🔍 Checking Ollama setup..."
	@if command -v ollama >/dev/null 2>&1; then \
		echo "✅ Ollama is installed"; \
		if ollama list | grep -q nomic-embed-text; then \
			echo "✅ nomic-embed-text model is available"; \
		else \
			echo "❌ nomic-embed-text model not found. Run: ollama pull nomic-embed-text"; \
		fi \
	else \
		echo "❌ Ollama not found. Install from: https://ollama.ai"; \
	fi

# Setup development environment
.PHONY: setup
setup: deps
	@echo "🛠️  Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📦 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "✅ Development environment ready"

# Show help
.PHONY: help
help:
	@echo "🧀 Cheesebox Development Commands"
	@echo ""
	@echo "Building:"
	@echo "  build      Build the binary"
	@echo "  dev        Development build"
	@echo "  release    Release build with optimizations"
	@echo "  build-all  Cross-compile for multiple platforms"
	@echo ""
	@echo "Installation:"
	@echo "  install    Install binary system-wide"
	@echo "  uninstall  Remove installed binary"
	@echo ""
	@echo "Development:"
	@echo "  deps       Install dependencies"
	@echo "  setup      Setup development environment"
	@echo "  test       Run tests"
	@echo "  test-race  Run tests with race detection"
	@echo "  fmt        Format code"
	@echo "  lint       Run linter"
	@echo ""
	@echo "Utilities:"
	@echo "  run        Build and run the application"
	@echo "  clean      Clean build artifacts"
	@echo "  check-ollama Check Ollama setup"
	@echo "  help       Show this help message"

# Default help target
.DEFAULT_GOAL := help
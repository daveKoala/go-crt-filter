.PHONY: help setup install-tools dev run build test clean

# Default target - show available commands
help:
	@echo "Available commands:"
	@echo "  make setup        - First-time setup: install all tools and dependencies"
	@echo "  make install-tools - Install development tools (air, etc.)"
	@echo "  make dev          - Start development server with hot-reload (air)"
	@echo "  make run          - Run server once (no hot-reload)"
	@echo "  make build        - Build the binary"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"

# First-time setup for new developers
setup: install-tools
	@echo "Installing project dependencies..."
	go mod download
	go mod tidy
	@echo ""
	@echo "Setup complete! Run 'make dev' to start the development server."

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@echo "Installing air (hot-reload)..."
	@go install github.com/air-verse/air@latest
	@echo "Tools installed successfully!"

# Start development server with hot-reload
dev:
	@echo "Starting development server with hot-reload..."
	@GOBIN=$$(go env GOPATH)/bin; \
	if [ ! -f "$$GOBIN/air" ]; then \
		echo "Error: air not found. Run 'make install-tools' first."; \
		exit 1; \
	fi; \
	$$GOBIN/air

# Run server once (no hot-reload)
run:
	go run ./cmd/server/main.go

# Build the binary
build:
	@echo "Building binary..."
	go build -o bin/server ./cmd/server

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf tmp/
	rm -rf bin/
	@echo "Clean complete!"

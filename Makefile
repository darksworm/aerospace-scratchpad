# Makefile for aerospace-scratchpad

.PHONY: all build clean swift go embed test help dev install

# Default target
all: build

# Build the complete project
build: embed go

# Build Swift window manager
swift:
	@echo "Building Swift window manager..."
	cd swift-window-manager && swift build -c release

# Embed the Swift binary for Go compilation
embed: swift
	@echo "Copying window manager binary for embedding..."
	cp swift-window-manager/.build/release/window-manager internal/aerospace/window-manager

# Build Go binary with embedded Swift window manager
go: embed
	@echo "Building Go binary with embedded Swift window manager..."
	go build -o aerospace-scratchpad
	@echo "Build complete! Size: $$(ls -lh aerospace-scratchpad | awk '{print $$5}')"

# Run tests
test:
	@echo "Running Go tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f aerospace-scratchpad
	rm -f internal/aerospace/window-manager
	rm -rf swift-window-manager/.build

# Development build (faster, debug mode)
dev:
	@echo "Building Swift window manager (debug)..."
	cd swift-window-manager && swift build
	@echo "Copying debug binary for embedding..."
	cp swift-window-manager/.build/debug/window-manager internal/aerospace/window-manager
	@echo "Building Go binary..."
	go build -o aerospace-scratchpad
	@echo "Development build complete!"

# Install to a local bin directory
install: build
	@echo "Installing to ~/bin..."
	mkdir -p ~/bin
	cp aerospace-scratchpad ~/bin/
	@echo "Installed to ~/bin/aerospace-scratchpad"

# Build for release (optimized)
release: embed
	@echo "Building optimized release binary..."
	go build -ldflags="-s -w" -o aerospace-scratchpad
	@echo "Release build complete! Size: $$(ls -lh aerospace-scratchpad | awk '{print $$5}')"

# Run linting
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  all      - Build everything (default)"
	@echo "  build    - Build the complete project"
	@echo "  swift    - Build only the Swift window manager"
	@echo "  go       - Build only the Go binary (requires embed)"
	@echo "  embed    - Copy Swift binary for Go embedding"
	@echo "  dev      - Development build (debug mode)"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  install  - Install to ~/bin"
	@echo "  release  - Build optimized release binary"
	@echo "  lint     - Run linting"
	@echo "  help     - Show this help message"
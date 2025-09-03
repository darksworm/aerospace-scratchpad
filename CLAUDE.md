# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is aerospace-scratchpad (fork: aerospace-sticky), a Go CLI tool that implements I3/Sway-like scratchpad functionality for AeroSpace window manager on macOS. The project embeds a Swift window manager binary for advanced window positioning and sizing capabilities.

## Development Commands

### Building
- `make build` - Complete build (Swift + Go with embedding)  
- `make dev` - Development build (debug mode, faster)
- `make release` - Optimized release build with stripped symbols

### Testing and Quality
- `go test ./...` - Run all Go tests
- `make test` - Run tests via Makefile
- `make lint` - Run golangci-lint (requires golangci-lint installation)
- `go fmt ./...` - Format Go code

### Swift Component
- `make swift` - Build only the Swift window manager
- `cd swift-window-manager && swift build` - Direct Swift build

### Utilities
- `make clean` - Remove all build artifacts
- `make install` - Install to ~/bin/aerospace-scratchpad

## Architecture

### Core Components

**CLI Structure (cmd/)**
- `root.go` - Cobra CLI setup with commands: move, show, summon, next, info
- `show.go`/`summon.go` - Main scratchpad operations for showing/hiding windows
- `move.go` - Move windows to scratchpad workspace
- `next.go` - Cycle through scratchpad windows

**Internal Packages**
- `aerospace/client.go` - Extended AeroSpace client with window geometry, fullscreen, and Swift window manager integration
- `aerospace/workspace.go` - Workspace management utilities  
- `logger/logger.go` - Configurable logging system with environment variable control
- `registry/registry.go` - Window tracking and state management
- `tracker/tracker.go` - Window position/state persistence

**Swift Integration**
- `swift-window-manager/` - Swift package for precise window positioning on macOS
- Binary is embedded into Go executable via `//go:embed` directive
- Extracted to temp file at runtime for execution

### Key Architecture Patterns

**Hybrid Go/Swift Architecture**: Go handles CLI logic and AeroSpace IPC, Swift handles native macOS window management for precise positioning and notch-aware layouts.

**Embedded Binary Strategy**: Swift window manager is compiled and embedded into the Go binary, eliminating external dependencies at runtime.

**IPC Communication**: Uses github.com/cristianoliveira/aerospace-ipc for efficient Unix socket communication with AeroSpace, avoiding subprocess spawning overhead.

**Workspace-based Scratchpad**: Windows are moved to a dedicated ".scratchpad" workspace rather than being hidden/minimized.

## Environment Variables

**Logging Control**
- `AEROSPACE_SCRATCHPAD_LOGS_PATH` - Log file path (default: `/tmp/aerospace-scratchpad.log`)
- `AEROSPACE_SCRATCHPAD_LOGS_LEVEL` - Log level (default: `DISABLED`, options: `DEBUG`, `INFO`, etc.)

## Dependencies

- **Go 1.24.2+** - Core language
- **AeroSpace 0.15.x+** - Window manager dependency
- **Swift 5.5+** - For window-manager component
- **github.com/spf13/cobra** - CLI framework
- **github.com/cristianoliveira/aerospace-ipc** - AeroSpace communication

## Testing

Tests use:
- `github.com/gkampitakis/go-snaps` - Snapshot testing
- `go.uber.org/mock` - Mocking framework
- Standard Go testing with `*_test.go` files in cmd/ directory

Run individual test files: `go test ./cmd/show_test.go`

## Build Pipeline

CI runs on GitHub Actions:
1. Go 1.24+ setup
2. `go mod download`  
3. `make test`
4. `make lint` (requires golangci-lint)
5. Code formatting validation

The build process requires both Go and Swift toolchains due to the hybrid architecture.
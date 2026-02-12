# AGENTS.md - Coding Agent Instructions

## Project Overview

This is `gowebdavd` - a simple WebDAV server written in Go. It serves a local directory over WebDAV protocol with daemon mode support (background service management). The project follows standard Go project layout conventions.

## Project Structure

```
gowebdavd/
├── cmd/
│   └── gowebdavd/
│       └── main.go              # Application entry point
├── internal/
│   ├── daemon/
│   │   ├── daemon_unix.go       # Unix-specific daemon implementation
│   │   ├── daemon_windows.go    # Windows-specific daemon implementation
│   │   └── daemon_test.go       # Daemon tests
│   ├── pidfile/
│   │   ├── pidfile.go           # PID file interface and implementation
│   │   └── pidfile_test.go      # PID file tests
│   ├── process/
│   │   ├── process.go           # Process management interfaces
│   │   ├── process_unix.go      # Unix-specific implementation
│   │   ├── process_windows.go   # Windows-specific implementation
│   │   ├── mock.go              # Mock implementations for testing
│   │   └── process_test.go      # Process tests
│   └── server/
│       ├── server.go            # WebDAV server implementation
│       └── server_test.go       # Server tests
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── build.cmd                    # Windows build script
└── AGENTS.md                    # This file
```

## Build Commands

```bash
# Build for current platform
go build -o gowebdavd ./cmd/gowebdavd

# Build with optimizations (production)
go build -ldflags="-s -w" -o gowebdavd ./cmd/gowebdavd

# Build for Windows
go build -ldflags="-s -w" -o gowebdavd.exe ./cmd/gowebdavd

# Run the server directly (foreground)
go run ./cmd/gowebdavd run -dir /path/to/folder -port 8080 -bind 127.0.0.1

# Cross-compile examples
GOOS=linux GOARCH=amd64 go build -o gowebdavd-linux ./cmd/gowebdavd
GOOS=darwin GOARCH=amd64 go build -o gowebdavd-darwin ./cmd/gowebdavd
GOOS=windows GOARCH=amd64 go build -o gowebdavd.exe ./cmd/gowebdavd
```

## Usage Commands

```bash
# Start WebDAV server in background
./gowebdavd start -dir /path/to/folder -port 8080 -bind 127.0.0.1

# Check service status
./gowebdavd status

# Stop background WebDAV server
./gowebdavd stop

# Run server in foreground (for testing/debugging)
./gowebdavd run -dir /path/to/folder -port 8080 -bind 127.0.0.1
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/daemon/...

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Lint Commands

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

## Package Overview

### cmd/gowebdavd
Main application entry point. Contains CLI argument parsing and command dispatch.

### internal/daemon
Daemon management functionality for starting, stopping, and checking service status. Platform-specific implementations for Unix and Windows.

### internal/pidfile
PID file management interface and implementation. Handles reading, writing, and removing PID files.

### internal/process
Process management interfaces and platform-specific implementations. Includes mock implementations for testing.

### internal/server
WebDAV HTTP server implementation. Wraps the golang.org/x/net/webdav handler.

## Key Design Decisions

1. **Standard Go Layout**: Follows the Standard Go Project Layout with `cmd/` for main applications and `internal/` for private packages.

2. **Interface-Based Design**: Uses interfaces (PIDFile, Manager, Process) to enable testing with mocks.

3. **Platform Abstraction**: Platform-specific code separated using build tags (`//go:build !windows` and `//go:build windows`).

4. **Test Coverage**: Mock implementations provided in `mock.go` files for cross-package testing.

## Adding New Features

1. **New Commands**: Add command handlers in `cmd/gowebdavd/main.go`
2. **New Internal Packages**: Create under `internal/` with appropriate interface definitions
3. **Tests**: Add `*_test.go` files alongside source files
4. **Mocks**: If needed by other packages, add to `mock.go` files

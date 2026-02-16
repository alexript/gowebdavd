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
│   ├── logger/
│   │   ├── logger.go            # HTTP request logging
│   │   └── logger_test.go       # Logger tests
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

### Linux/macOS/WSL (via Makefile)

```bash
# Recommended
make build           # builds bin/gowebdavd
make build-release   # builds bin/gowebdavd with -s -w

# Alternative (plain go)
go build -o bin/gowebdavd ./cmd/gowebdavd
go build -ldflags="-s -w" -o bin/gowebdavd ./cmd/gowebdavd

# Run the server directly (foreground)
go run ./cmd/gowebdavd run -dir /path/to/folder -port 8080 -bind 127.0.0.1
```

### Windows Native (via build.cmd)

```cmd
build.cmd build         # builds bin\gowebdavd.exe
build.cmd build-release # builds bin\gowebdavd.exe with -s -w
build.cmd run           # build and run in foreground
```

### Cross-compile Examples

```bash
GOOS=linux GOARCH=amd64 go build -o bin/gowebdavd-linux ./cmd/gowebdavd
GOOS=darwin GOARCH=amd64 go build -o bin/gowebdavd-darwin ./cmd/gowebdavd
GOOS=windows GOARCH=amd64 go build -o bin/gowebdavd.exe ./cmd/gowebdavd
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

# Start with HTTP request logging enabled
./gowebdavd start -dir /path/to/folder -log

# Start with custom log directory (directory must exist)
./gowebdavd start -dir /path/to/folder -log -log-dir /var/log/gowebdavd
```

## Test Commands

### Linux/macOS/WSL (via Makefile)

```bash
make test            # go test ./...
make cover           # coverage.out + summary
```

### Windows Native (via build.cmd)

```cmd
build.cmd test       # go test ./...
build.cmd cover      # coverage.out + summary
```

### Direct go Commands (all platforms)

```bash
go test ./...
go test -cover ./...
go test -v ./...
go test ./internal/daemon/...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Lint Commands

### Direct go Commands (all platforms)

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

### Via Makefile (Linux/macOS/WSL)

```bash
make fmt
make vet
make tidy
```

### Via build.cmd (Windows native)

```cmd
build.cmd fmt
build.cmd vet
build.cmd tidy
```

## Clean Commands

### Linux/macOS/WSL (via Makefile)

```bash
make clean           # go clean; purge test cache; remove bin/ and coverage files
```

### Windows Native (via build.cmd)

```cmd
build.cmd clean      # go clean; purge test cache; remove bin\ and coverage files
```

### Direct Commands (all platforms)

```bash
# Clean Go build/test artifacts
go clean             # removes build cache and object files
go clean -testcache  # expires cached test results

# Remove build outputs (Linux/macOS)
rm -rf bin coverage.out coverage.html coverage

# Remove build outputs (Windows CMD)
rmdir /s /q bin
del /f coverage.out coverage.html
```

## Package Overview

### cmd/gowebdavd
Main application entry point. Contains CLI argument parsing and command dispatch.

### internal/daemon
Daemon management functionality for starting, stopping, and checking service status. Platform-specific implementations for Unix and Windows.

### internal/logger
HTTP request logging with automatic log rotation. Log files are stored in platform-specific directories and automatically cleaned up after 1 month.

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

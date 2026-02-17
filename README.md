# gowebdavd

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple WebDAV server written in Go with daemon mode support for background service management.

## Features

- 🚀 **Easy to use** - Simple CLI with intuitive commands
- 🔧 **Daemon mode** - Run as a background service
- 🖥️ **Cross-platform** - Works on Linux, macOS, and Windows
- 📝 **HTTP request logging** - Optional logging to files with automatic cleanup
- 📁 **Serve any directory** - Point to any folder on your system
- 🔒 **Secure by default** - Binds to localhost only
- 🧪 **Well tested** - Comprehensive test coverage

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/alexript/gowebdavd.git
cd gowebdavd

# Build (Linux/macOS/WSL - via Makefile)
make build   # produces bin/gowebdavd

# Build (Windows native - via build.cmd)
build.cmd build   # produces bin\gowebdavd.exe

# Or build for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/gowebdavd-linux ./cmd/gowebdavd
GOOS=darwin GOARCH=amd64 go build -o bin/gowebdavd-darwin ./cmd/gowebdavd
GOOS=windows GOARCH=amd64 go build -o bin/gowebdavd.exe ./cmd/gowebdavd
```

## Usage

### Quick Start

```bash
# Start WebDAV server in background (from bin/)
./bin/gowebdavd start -dir /path/to/folder -port 8080

# Check if service is running
./bin/gowebdavd status

# Stop the service
./bin/gowebdavd stop
```

### Available Commands

| Command | Description |
|---------|-------------|
| `start` | Start WebDAV server in background |
| `stop`  | Stop the background WebDAV server |
| `status`| Show current service status |
| `run`   | Run WebDAV server in foreground |

### Command Options

All `start` and `run` commands support the following flags:

- `-dir` - Directory to serve (default: current directory)
- `-port` - Port to listen on (default: 8080)
- `-bind` - IP address to bind to (default: 127.0.0.1)
- `-log` - Enable HTTP request logging (default: false)
- `-log-dir` - Custom log directory (requires `-log`, must exist)
- `-no-lock` - Disable WebDAV locking (for davfs2 compatibility)

### Examples

#### Serve current directory

```bash
./bin/gowebdavd start
```

#### Serve specific directory on custom port

```bash
./bin/gowebdavd start -dir /home/user/documents -port 9090
```

#### Bind to all interfaces (use with caution)

```bash
./bin/gowebdavd start -dir /srv/webdav -bind 0.0.0.0 -port 8080
```

#### Run in foreground for debugging

```bash
./bin/gowebdavd run -dir /path/to/folder -port 8080
```

## Use in Scripts

### Bash Example

```bash
#!/bin/bash

# Start WebDAV server
./gowebdavd start -dir /data -port 8080
sleep 2  # Wait for server to start

# Use WebDAV (example with curl)
curl -X PROPFIND http://127.0.0.1:8080/

# Do some work...

# Stop WebDAV server
./gowebdavd stop
```

### PowerShell Example (Windows)

```powershell
# Start WebDAV server
.\bin\gowebdavd.exe start -dir C:\Data -port 8080
Start-Sleep -Seconds 2

# Use WebDAV...

# Stop WebDAV server
.\bin\gowebdavd.exe stop
```

## Logging

The WebDAV server supports optional HTTP request logging. When enabled, all HTTP requests are logged to timestamped log files.

### Default Behavior

- Logs are stored in:
  - **Linux/macOS**: `~/.local/share/gowebdavd/logs/`
  - **Windows**: `%LOCALAPPDATA%\gowebdavd\logs\`
- Log files are named: `gowebdavd_YYYY-MM-DD_HH-MM-SS.log`
- Log files older than 1 month are automatically cleaned up
- Each log entry includes: client IP, HTTP method, URL path, status code, duration, and user agent

### Enable Logging

```bash
# Start with logging enabled
./bin/gowebdavd start -dir /path/to/folder -log
```

### Custom Log Directory

You can specify a custom log directory (the directory must already exist):

```bash
# Use custom log directory
./bin/gowebdavd start -dir /data -log -log-dir /var/log/gowebdavd
```

### davfs2 Compatibility

Some WebDAV clients like davfs2 may have issues with the standard WebDAV locking mechanism. The `-no-lock` flag disables WebDAV locking, which can resolve issues with git operations on davfs2 mounts:

```bash
# Start with locking disabled for davfs2 compatibility
./bin/gowebdavd start -dir /data -no-lock
```

**Use case**: When using gowebdavd with davfs2 mounts and git, you may encounter "Input/output error" during `git init`. This is caused by the WebDAV MOVE operation requiring lock tokens for both source and destination paths. The `-no-lock` flag resolves this by accepting all lock tokens without validation.

**Note**: Disabling locks reduces WebDAV protocol compliance but improves compatibility with certain clients.

### Log Format

Each log entry follows this format:

```
2026/02/16 10:30:45 127.0.0.1:54321 PROPFIND /documents 207 2.345ms curl/7.68.0
```

Format: `timestamp client_ip method path status_code duration user_agent`

## Project Structure

```
gowebdavd/
├── cmd/gowebdavd/        # Application entry point
├── internal/
│   ├── daemon/           # Daemon management
│   ├── logger/           # HTTP request logging
│   ├── pidfile/          # PID file operations
│   ├── process/          # Process management
│   └── server/           # WebDAV server
├── Makefile              # Build automation (Linux/macOS/WSL)
├── build.cmd             # Build automation (Windows native)
├── go.mod
├── go.sum
└── README.md
```

## Development

### Prerequisites

- Go 1.25 or later

### Build

**Linux/macOS/WSL:**
```bash
make build   # produces bin/gowebdavd
```

**Windows (native CMD/PowerShell):**
```cmd
build.cmd build   # produces bin\gowebdavd.exe
```

### Test

**Linux/macOS/WSL:**
```bash
# Run all tests
make test

# Run tests with coverage
make cover
```

**Windows (native CMD/PowerShell):**
```cmd
build.cmd test
build.cmd cover
```

**Direct go command (all platforms):**
```bash
# Run tests with verbose output
go test -v ./...
```

### Clean Up

To remove build and test artifacts:

**Linux/macOS/WSL:**
```bash
make clean
```

**Windows (native CMD/PowerShell):**
```cmd
build.cmd clean
```

This runs `go clean`, clears test cache, and removes `bin/` and coverage files.


### Project Layout

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

- `cmd/gowebdavd/` - Main application entry point
- `internal/` - Private application code
  - `daemon/` - Background service management
  - `logger/` - HTTP request logging with file rotation
  - `pidfile/` - PID file handling
  - `process/` - Process management with mocks for testing
  - `server/` - WebDAV HTTP server

## Architecture

### Interface-Based Design

The project uses interfaces for testability:

- **PIDFile** - PID file operations (read, write, remove)
- **ProcessManager** - Process management (IsRunning, FindProcess, Terminate, Kill)
- **Process** - Process operations (Signal, Kill, Pid)

### Cross-Platform Support

Platform-specific implementations are separated using build tags:
- `*_unix.go` - Linux and macOS implementation
- `*_windows.go` - Windows implementation

## Security Considerations

- **Default bind address**: 127.0.0.1 (localhost) - only accessible from the local machine
- **PID file location**: Stored in the user's temp directory
- **No authentication**: This is a simple file server; do not expose to untrusted networks without additional security measures

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## Acknowledgments

- Built with [golang.org/x/net/webdav](https://pkg.go.dev/golang.org/x/net/webdav) WebDAV implementation

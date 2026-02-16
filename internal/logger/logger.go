// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

// Package logger provides HTTP request logging functionality.
package logger

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Logger handles HTTP request logging
type Logger struct {
	enabled bool
	file    *os.File
	logger  *log.Logger
}

// New creates a new Logger instance
// logDir: custom log directory path. If empty, uses default directory.
// When custom directory is specified, it must exist (won't be created automatically).
func New(enabled bool, logDir string) (*Logger, error) {
	if !enabled {
		return &Logger{enabled: false}, nil
	}

	var err error
	useDefaultDir := logDir == ""

	if useDefaultDir {
		logDir, err = getLogDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get log directory: %w", err)
		}

		// Create default directory if it doesn't exist
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	} else {
		// For custom directory, check if it exists (don't create)
		info, err := os.Stat(logDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("log directory does not exist: %s", logDir)
			}
			return nil, fmt.Errorf("failed to access log directory: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("log path is not a directory: %s", logDir)
		}
	}

	if err := cleanupOldLogs(logDir); err != nil {
		// Log cleanup errors but don't fail
		log.Printf("Warning: failed to cleanup old logs: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("gowebdavd_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &Logger{
		enabled: enabled,
		file:    file,
		logger:  log.New(file, "", log.LstdFlags),
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Middleware returns HTTP middleware that logs requests
func (l *Logger) Middleware(next http.Handler) http.Handler {
	if !l.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		l.logger.Printf("%s %s %s %d %s %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			r.UserAgent(),
		)
	})
}

// Enabled returns whether logging is enabled
func (l *Logger) Enabled() bool {
	return l.enabled
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// getLogDir returns the log directory path based on OS
func getLogDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		// On Windows, use %LOCALAPPDATA%\gowebdavd\logs
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(homeDir, "AppData", "Local")
		}
		return filepath.Join(localAppData, "gowebdavd", "logs"), nil
	}

	// On Unix-like systems, use ~/.local/share/gowebdavd/logs
	return filepath.Join(homeDir, ".local", "share", "gowebdavd", "logs"), nil
}

// cleanupOldLogs removes log files older than 1 month
func cleanupOldLogs(logDir string) error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cutoff := time.Now().AddDate(0, -1, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasPrefix(entry.Name(), "gowebdavd_") || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			logPath := filepath.Join(logDir, entry.Name())
			if err := os.Remove(logPath); err != nil {
				log.Printf("Warning: failed to remove old log file %s: %v", logPath, err)
			}
		}
	}

	return nil
}

// NewNopLogger creates a no-op logger for testing
func NewNopLogger() *Logger {
	return &Logger{enabled: false}
}

// NewWithWriter creates a logger with a custom writer (for testing)
func NewWithWriter(w io.Writer, enabled bool) *Logger {
	if !enabled {
		return &Logger{enabled: false}
	}
	return &Logger{
		enabled: enabled,
		logger:  log.New(w, "", log.LstdFlags),
	}
}

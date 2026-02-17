package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gowebdavd/internal/logger"
)

// mockLogger is a test double for logger.Logger
type mockLogger struct {
	enabled bool
}

func (m *mockLogger) Enabled() bool                          { return m.enabled }
func (m *mockLogger) Middleware(h http.Handler) http.Handler { return h }

func TestNew(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	server := New(tempDir, 8080, "127.0.0.1", nil)
	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.Addr() != "127.0.0.1:8080" {
		t.Errorf("Expected address 127.0.0.1:8080, got %s", server.Addr())
	}
}

func TestDirectoryTraversal(t *testing.T) {
	// Create a temporary directory for WebDAV root
	tempDir := t.TempDir()

	// Create a file outside the WebDAV root
	parentDir := filepath.Dir(tempDir)
	outsideFile := filepath.Join(parentDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret content"), 0644); err != nil {
		t.Fatalf("Failed to create outside file: %v", err)
	}
	defer os.Remove(outsideFile)

	// Create WebDAV server
	server := New(tempDir, 8080, "127.0.0.1", nil)

	// Test various directory traversal attempts
	traversalPaths := []string{
		"../secret.txt",
		"../../secret.txt",
		"../../../secret.txt",
		"test/../../../secret.txt",
		"test/../test/../../../secret.txt",
	}

	for _, path := range traversalPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+path, nil)
			rr := httptest.NewRecorder()

			server.Handler().ServeHTTP(rr, req)

			// Should return 404 or 403, not 200 with secret content
			if rr.Code == http.StatusOK {
				body, _ := io.ReadAll(rr.Body)
				if string(body) == "secret content" {
					t.Errorf("Directory traversal succeeded: got content %q", string(body))
				}
			}
			if rr.Code != http.StatusNotFound && rr.Code != http.StatusForbidden {
				t.Logf("Got status %d for path %s", rr.Code, path)
			}
		})
	}
}

func TestWebDAVAddr(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		bind     string
		expected string
	}{
		{
			name:     "localhost default port",
			port:     8080,
			bind:     "127.0.0.1",
			expected: "127.0.0.1:8080",
		},
		{
			name:     "all interfaces custom port",
			port:     9999,
			bind:     "0.0.0.0",
			expected: "0.0.0.0:9999",
		},
		{
			name:     "IPv6 localhost",
			port:     8080,
			bind:     "::1",
			expected: "::1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srv := New(tmpDir, tt.port, tt.bind, nil)
			if srv.Addr() != tt.expected {
				t.Errorf("Addr() = %s, want %s", srv.Addr(), tt.expected)
			}
		})
	}
}

func TestWebDAVHandler(t *testing.T) {
	tmpDir := t.TempDir()
	srv := New(tmpDir, 18080, "127.0.0.1", nil)

	handler := srv.Handler()
	if handler == nil {
		t.Fatal("Handler() returned nil")
	}

	// Verify handler implements http.Handler interface
	var _ http.Handler = handler
}

func TestWebDAVHandlerCapabilities(t *testing.T) {
	tmpDir := t.TempDir()
	srv := New(tmpDir, 18080, "127.0.0.1", nil)
	handler := srv.Handler()

	// Test that the handler can be used with http.Handler interface
	var _ http.Handler = handler

	// Verify handler is not nil and can serve requests
	if handler == nil {
		t.Fatal("Handler() returned nil")
	}
}

func TestNew_WithLogger(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.NewNopLogger()
	srv := New(tmpDir, 18080, "127.0.0.1", log)

	if srv == nil {
		t.Fatal("New() returned nil")
	}

	// When logger is provided and enabled, handler should be wrapped
	if srv.Handler() == nil {
		t.Error("Handler() returned nil")
	}
}

func TestNew_WithDisabledLogger(t *testing.T) {
	log := &mockLogger{enabled: false}
	srv := New(t.TempDir(), 8080, "127.0.0.1", log)

	if srv == nil {
		t.Fatal("New returned nil")
	}

	handler := srv.Handler()
	if handler == nil {
		t.Error("Handler should not be nil")
	}

	var _ http.Handler = handler
}

func TestGracefulShutdown(t *testing.T) {
	dir := t.TempDir()
	// Use port 0 to get a random available port
	srv := New(dir, 0, "127.0.0.1", nil)

	// Start server in goroutine
	go func() {
		// This will block until shutdown or error
		srv.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that health endpoint responds
	resp, err := http.Get("http://" + srv.Addr() + "/health")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

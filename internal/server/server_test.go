package server

import (
	"net/http"
	"testing"

	"golang.org/x/net/webdav"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	srv := New(tmpDir, 18080, "127.0.0.1")

	if srv == nil {
		t.Fatal("New() returned nil")
	}

	expectedAddr := "127.0.0.1:18080"
	if srv.Addr() != expectedAddr {
		t.Errorf("Addr() = %s, want %s", srv.Addr(), expectedAddr)
	}

	if srv.Handler() == nil {
		t.Error("Handler() returned nil")
	}

	// Verify handler is webdav.Handler
	handler, ok := srv.Handler().(*webdav.Handler)
	if !ok {
		t.Error("Handler() should return *webdav.Handler")
	}

	if handler.FileSystem == nil {
		t.Error("Handler.FileSystem is nil")
	}

	if handler.LockSystem == nil {
		t.Error("Handler.LockSystem is nil")
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
			srv := New(tmpDir, tt.port, tt.bind)
			if srv.Addr() != tt.expected {
				t.Errorf("Addr() = %s, want %s", srv.Addr(), tt.expected)
			}
		})
	}
}

func TestWebDAVHandler(t *testing.T) {
	tmpDir := t.TempDir()
	srv := New(tmpDir, 18080, "127.0.0.1")

	handler := srv.Handler()
	if handler == nil {
		t.Fatal("Handler() returned nil")
	}

	// Verify handler responds to WebDAV methods
	// We can't test actual HTTP requests without starting the server,
	// but we can verify the handler type
	_, ok := handler.(*webdav.Handler)
	if !ok {
		t.Error("Handler() should be *webdav.Handler")
	}
}

func TestWebDAVHandlerCapabilities(t *testing.T) {
	tmpDir := t.TempDir()
	srv := New(tmpDir, 18080, "127.0.0.1")
	handler := srv.Handler().(*webdav.Handler)

	// Test that the handler can be used with http.Handler interface
	var _ http.Handler = handler

	// Verify FileSystem is set correctly
	fs := handler.FileSystem
	if fs == nil {
		t.Fatal("FileSystem is nil")
	}

	// Verify it's a webdav.Dir
	_, ok := fs.(webdav.Dir)
	if !ok {
		t.Error("FileSystem should be webdav.Dir")
	}
}

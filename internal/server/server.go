// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

// Package server provides WebDAV server functionality.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/webdav"
)

// Logger interface for HTTP request logging
type Logger interface {
	Middleware(http.Handler) http.Handler
	Enabled() bool
}

// WebDAV wraps the WebDAV HTTP server
type WebDAV struct {
	server   *http.Server
	addr     string
	listener net.Listener
	logger   Logger
	root     string
}

// traversalProtection is a middleware that prevents directory traversal attacks.
// It wraps an http.Handler and blocks any requests containing ".." in the path.
type traversalProtection struct {
	handler http.Handler
	root    string
}

func (t *traversalProtection) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean and normalize the request path
	cleanPath := path.Clean(r.URL.Path)

	// Check for traversal attempts (path containing "..")
	if strings.Contains(cleanPath, "..") || strings.HasPrefix(cleanPath, "/..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	t.handler.ServeHTTP(w, r)
}

// New creates a new WebDAV server instance
func New(folder string, port int, bind string, log Logger) *WebDAV {
	davHandler := &webdav.Handler{
		FileSystem: webdav.Dir(folder),
		LockSystem: webdav.NewMemLS(),
	}

	var webdavHandler http.Handler = davHandler

	// Add directory traversal protection
	webdavHandler = &traversalProtection{
		handler: webdavHandler,
		root:    folder,
	}

	if log != nil && log.Enabled() {
		webdavHandler = log.Middleware(webdavHandler)
	}

	// Create mux with health endpoint
	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(healthHandler))
	mux.Handle("/", webdavHandler)

	addr := bind + ":" + strconv.Itoa(port)

	return &WebDAV{
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		addr:   addr,
		logger: log,
		root:   folder,
	}
}

// healthHandler responds with 200 OK when server is ready
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Start starts the WebDAV server with graceful shutdown support
func (s *WebDAV) Start() error {
	// Create listener if not already created (for dynamic port allocation)
	if s.listener == nil {
		listener, err := net.Listen("tcp", s.addr)
		if err != nil {
			return fmt.Errorf("failed to create listener: %w", err)
		}
		s.listener = listener
	}

	fmt.Printf("WebDAV server: http://%s\n", s.Addr())

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to capture server errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := s.server.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for either an interrupt signal or a server error
	select {
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal: %v\n", sig)
		return s.shutdown()
	case err := <-errChan:
		return err
	}
}

// shutdown gracefully shuts down the server with a 30-second timeout
func (s *WebDAV) shutdown() error {
	fmt.Println("Shutting down server...")

	// Create a context with 30-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	fmt.Println("Server stopped gracefully")
	return nil
}

// Addr returns the server address.
// If the server is listening on a dynamically allocated port (port 0),
// it returns the actual address including the assigned port.
func (s *WebDAV) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// Handler returns the HTTP handler.
// The handler includes all configured middleware (logging, traversal protection)
// and serves both WebDAV requests and the health endpoint.
func (s *WebDAV) Handler() http.Handler {
	return s.server.Handler
}

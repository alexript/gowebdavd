// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

// Package server provides WebDAV server functionality.
package server

import (
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/net/webdav"
)

// WebDAV wraps the WebDAV HTTP server
type WebDAV struct {
	handler http.Handler
	addr    string
}

// New creates a new WebDAV server instance
func New(folder string, port int, bind string) *WebDAV {
	handler := &webdav.Handler{
		FileSystem: webdav.Dir(folder),
		LockSystem: webdav.NewMemLS(),
	}

	return &WebDAV{
		handler: handler,
		addr:    bind + ":" + strconv.Itoa(port),
	}
}

// Start starts the WebDAV server (blocking)
func (s *WebDAV) Start() error {
	fmt.Printf("WebDAV server: http://%s\n", s.addr)
	if err := http.ListenAndServe(s.addr, s.handler); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Addr returns the server address
func (s *WebDAV) Addr() string {
	return s.addr
}

// Handler returns the HTTP handler
func (s *WebDAV) Handler() http.Handler {
	return s.handler
}

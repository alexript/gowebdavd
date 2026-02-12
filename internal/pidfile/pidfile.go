// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

// Package pidfile provides functionality for managing PID files.
package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// File defines the interface for PID file operations
type File interface {
	Read() (int, error)
	Write(pid int) error
	Remove() error
	Path() string
}

// file implements File interface
type file struct {
	path string
}

// New creates a new File instance with default path
func New() File {
	return &file{
		path: filepath.Join(os.TempDir(), "gowebdavd.pid"),
	}
}

// NewWithPath creates a new File instance with custom path
func NewWithPath(path string) File {
	return &file{path: path}
}

// Read reads the PID from the file
func (p *file) Read() (int, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %w", err)
	}
	return pid, nil
}

// Write writes the PID to the file
func (p *file) Write(pid int) error {
	return os.WriteFile(p.path, []byte(strconv.Itoa(pid)), 0644)
}

// Remove deletes the PID file
func (p *file) Remove() error {
	return os.Remove(p.path)
}

// Path returns the path to the PID file
func (p *file) Path() string {
	return p.path
}

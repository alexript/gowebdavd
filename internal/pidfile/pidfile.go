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

// File defines the interface for PID file operations.
// It supports atomic locking to prevent race conditions during
// concurrent access.
type File interface {
	Read() (int, error)
	Write(pid int) error
	Remove() error
	Path() string
	Lock() error
	Unlock() error
}

// file implements File interface
type file struct {
	path string
	fd   int // file descriptor for locking (-1 if not locked)
}

// New creates a new File instance with the default path.
// The default path is $TMPDIR/gowebdavd.pid or /tmp/gowebdavd.pid if TMPDIR is not set.
// The returned File is not locked; callers must call Lock() before operations.
func New() File {
	return &file{
		path: filepath.Join(os.TempDir(), "gowebdavd.pid"),
	}
}

// NewWithPath creates a new File instance with a custom path.
// This is useful for specifying a system-wide PID location like /var/run.
// The returned File is not locked; callers must call Lock() before operations.
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

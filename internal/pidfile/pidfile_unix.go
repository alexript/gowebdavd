//go:build !windows

// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package pidfile

import (
	"fmt"
	"syscall"
)

// Lock acquires an advisory lock on the PID file using flock.
// This is a blocking call that waits until the lock is available.
func (p *file) Lock() error {
	// Open or create the lock file
	fd, err := syscall.Open(p.path, syscall.O_RDWR|syscall.O_CREAT, 0644)
	if err != nil {
		return fmt.Errorf("failed to open PID file for locking: %w", err)
	}

	// Acquire exclusive lock (blocking)
	err = syscall.Flock(fd, syscall.LOCK_EX)
	if err != nil {
		syscall.Close(fd)
		return fmt.Errorf("failed to lock PID file: %w", err)
	}

	p.fd = fd
	return nil
}

// Unlock releases the advisory lock on the PID file.
func (p *file) Unlock() error {
	if p.fd < 0 {
		return nil // Already unlocked
	}

	// Release the lock
	err := syscall.Flock(p.fd, syscall.LOCK_UN)
	if err != nil {
		return fmt.Errorf("failed to unlock PID file: %w", err)
	}

	// Close the file descriptor
	syscall.Close(p.fd)
	p.fd = -1
	return nil
}

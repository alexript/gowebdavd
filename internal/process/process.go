// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

// Package process provides interfaces and implementations for process management.
package process

// Process represents an OS process
type Process interface {
	Signal(sig int) error
	Kill() error
	Pid() int
}

// Manager defines the interface for process operations
type Manager interface {
	IsRunning(pid int) bool
	FindProcess(pid int) (Process, error)
	Terminate(pid int) error
	Kill(pid int) error
}

//go:build windows

// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package pidfile

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	lockFileEx   = kernel32.NewProc("LockFileEx")
	unlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	LOCKFILE_EXCLUSIVE_LOCK   = 0x00000002
	LOCKFILE_FAIL_IMMEDIATELY = 0x00000001
)

// Lock acquires an exclusive lock on the PID file.
// This is a blocking call that waits until the lock is available.
func (p *file) Lock() error {
	// Open or create the file
	pathPtr, err := syscall.UTF16PtrFromString(p.path)
	if err != nil {
		return fmt.Errorf("failed to convert path: %w", err)
	}

	handle, _, err := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateFileW").Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(syscall.GENERIC_READ|syscall.GENERIC_WRITE),
		0, // No sharing (exclusive access)
		0, // Default security
		uintptr(syscall.OPEN_ALWAYS),
		uintptr(syscall.FILE_ATTRIBUTE_NORMAL),
		0,
	)

	if syscall.Handle(handle) == syscall.InvalidHandle {
		return fmt.Errorf("failed to open PID file for locking: %w", err)
	}

	p.fd = int(handle)
	return nil
}

// Unlock releases the lock on the PID file.
func (p *file) Unlock() error {
	if p.fd < 0 {
		return nil // Already unlocked
	}

	handle := syscall.Handle(p.fd)
	syscall.CloseHandle(handle)
	p.fd = -1
	return nil
}

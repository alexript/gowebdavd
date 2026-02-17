//go:build windows

// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package process

import (
	"fmt"
	"syscall"
)

// STILL_ACTIVE is the Windows constant indicating a process is still running
const STILL_ACTIVE = 259

// windowsManager implements Manager for Windows systems
type windowsManager struct{}

// NewManager creates a new Manager for the current platform
func NewManager() Manager {
	return &windowsManager{}
}

func (w *windowsManager) IsRunning(pid int) bool {
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}

	return exitCode == STILL_ACTIVE
}

func (w *windowsManager) FindProcess(pid int) (Process, error) {
	return nil, fmt.Errorf("not implemented on Windows")
}

func (w *windowsManager) Terminate(pid int) error {
	handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("failed to open process: %w", err)
	}
	defer syscall.CloseHandle(handle)

	return syscall.TerminateProcess(handle, 0)
}

func (w *windowsManager) Kill(pid int) error {
	return w.Terminate(pid)
}

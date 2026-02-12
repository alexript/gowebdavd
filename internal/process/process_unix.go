//go:build !windows

// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package process

import (
	"fmt"
	"os"
	"syscall"
)

// unixProcess wraps os.Process for Unix systems
type unixProcess struct {
	process *os.Process
}

func (p *unixProcess) Signal(sig int) error {
	return p.process.Signal(syscall.Signal(sig))
}

func (p *unixProcess) Kill() error {
	return p.process.Kill()
}

func (p *unixProcess) Pid() int {
	return p.process.Pid
}

// unixManager implements Manager for Unix systems
type unixManager struct{}

// NewManager creates a new Manager for the current platform
func NewManager() Manager {
	return &unixManager{}
}

func (u *unixManager) IsRunning(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil
}

func (u *unixManager) FindProcess(pid int) (Process, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to find process: %w", err)
	}
	return &unixProcess{process: proc}, nil
}

func (u *unixManager) Terminate(pid int) error {
	proc, err := u.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(int(syscall.SIGTERM))
}

func (u *unixManager) Kill(pid int) error {
	proc, err := u.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

//go:build windows

// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

// Package daemon provides daemon management functionality.
package daemon

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"

	"gowebdavd/internal/pidfile"
	"gowebdavd/internal/process"
)

// Daemon manages the WebDAV background service
type Daemon struct {
	pidFile  pidfile.File
	procMgr  process.Manager
	execPath string
}

// New creates a new Daemon instance
func New(pf pidfile.File, pm process.Manager, execPath string) *Daemon {
	return &Daemon{
		pidFile:  pf,
		procMgr:  pm,
		execPath: execPath,
	}
}

// Start starts the WebDAV service in background
func (d *Daemon) Start(folder string, port int, bind string, enableLog bool, logDir string) error {
	pid, err := d.pidFile.Read()
	if err == nil && d.procMgr.IsRunning(pid) {
		fmt.Printf("Service is already running (PID: %d)\n", pid)
		return nil
	}

	if err == nil {
		d.pidFile.Remove()
	}

	args := []string{"run", "-dir", folder, "-port", strconv.Itoa(port), "-bind", bind}
	if enableLog {
		args = append(args, "-log")
		if logDir != "" {
			args = append(args, "-log-dir", logDir)
		}
	}

	cmd := exec.Command(d.execPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	if err := d.pidFile.Write(cmd.Process.Pid); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to write PID: %w", err)
	}

	fmt.Printf("Service started (PID: %d)\n", cmd.Process.Pid)
	return nil
}

// Stop stops the WebDAV service
func (d *Daemon) Stop() error {
	pid, err := d.pidFile.Read()
	if err != nil {
		fmt.Println("Service is not running")
		return nil
	}

	if !d.procMgr.IsRunning(pid) {
		d.pidFile.Remove()
		fmt.Println("Service is not running")
		return nil
	}

	if err := d.procMgr.Terminate(pid); err != nil {
		if err := d.procMgr.Kill(pid); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}
	}

	d.pidFile.Remove()
	fmt.Println("Service stopped")
	return nil
}

// Status checks the service status
func (d *Daemon) Status() error {
	pid, err := d.pidFile.Read()
	if err != nil {
		fmt.Println("Service is not running")
		return nil
	}

	if d.procMgr.IsRunning(pid) {
		fmt.Printf("Service is running (PID: %d)\n", pid)
	} else {
		fmt.Printf("PID file exists but process %d not found\n", pid)
		d.pidFile.Remove()
	}
	return nil
}

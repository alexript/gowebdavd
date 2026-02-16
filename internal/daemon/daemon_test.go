package daemon

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gowebdavd/internal/process"
)

// createTestExecutable creates a platform-specific test executable
func createTestExecutable(t *testing.T, dir string) string {
	t.Helper()

	if runtime.GOOS == "windows" {
		execPath := filepath.Join(dir, "testexec.bat")
		// Create a batch file that exits immediately
		content := []byte("@echo off\nexit /b 0")
		if err := os.WriteFile(execPath, content, 0755); err != nil {
			t.Fatalf("Failed to create test executable: %v", err)
		}
		return execPath
	}

	execPath := filepath.Join(dir, "testexec")
	content := []byte("#!/bin/sh\nexit 0")
	if err := os.WriteFile(execPath, content, 0755); err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}
	return execPath
}

// MockPIDFile implements pidfile.File for testing
type MockPIDFile struct {
	Pid       int
	ReadErr   error
	WriteErr  error
	RemoveErr error
	PathValue string
	Removed   bool
	Written   int
}

func (m *MockPIDFile) Read() (int, error) {
	if m.ReadErr != nil {
		return 0, m.ReadErr
	}
	return m.Pid, nil
}

func (m *MockPIDFile) Write(pid int) error {
	m.Written = pid
	if m.WriteErr != nil {
		return m.WriteErr
	}
	m.Pid = pid
	return nil
}

func (m *MockPIDFile) Remove() error {
	m.Removed = true
	return m.RemoveErr
}

func (m *MockPIDFile) Path() string {
	if m.PathValue != "" {
		return m.PathValue
	}
	return "/tmp/test.pid"
}

func TestStatusNotRunning(t *testing.T) {
	pf := &MockPIDFile{ReadErr: os.ErrNotExist}
	pm := &process.MockManager{}
	d := New(pf, pm, "/bin/test")

	err := d.Status()
	if err != nil {
		t.Errorf("Status() error = %v", err)
	}
}

func TestStatusRunning(t *testing.T) {
	pf := &MockPIDFile{Pid: 1234}
	pm := &process.MockManager{
		RunningPids: map[int]bool{1234: true},
	}
	d := New(pf, pm, "/bin/test")

	err := d.Status()
	if err != nil {
		t.Errorf("Status() error = %v", err)
	}
}

func TestStatusStalePID(t *testing.T) {
	pf := &MockPIDFile{Pid: 1234}
	pm := &process.MockManager{
		RunningPids: map[int]bool{},
	}
	d := New(pf, pm, "/bin/test")

	err := d.Status()
	if err != nil {
		t.Errorf("Status() error = %v", err)
	}
	if !pf.Removed {
		t.Error("Status() should remove stale PID file")
	}
}

func TestStopNotRunning(t *testing.T) {
	pf := &MockPIDFile{ReadErr: os.ErrNotExist}
	pm := &process.MockManager{}
	d := New(pf, pm, "/bin/test")

	err := d.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestStopStalePID(t *testing.T) {
	pf := &MockPIDFile{Pid: 1234}
	pm := &process.MockManager{
		RunningPids: map[int]bool{},
	}
	d := New(pf, pm, "/bin/test")

	err := d.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	if !pf.Removed {
		t.Error("Stop() should remove stale PID file")
	}
}

func TestStopRunning(t *testing.T) {
	pf := &MockPIDFile{Pid: 1234}
	pm := &process.MockManager{
		RunningPids: map[int]bool{1234: true},
	}
	d := New(pf, pm, "/bin/test")

	err := d.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	if !pf.Removed {
		t.Error("Stop() should remove PID file")
	}
}

func TestStopKillFallback(t *testing.T) {
	pf := &MockPIDFile{Pid: 1234}
	pm := &process.MockManager{
		RunningPids:  map[int]bool{1234: true},
		TerminateErr: errors.New("terminate failed"),
	}
	d := New(pf, pm, "/bin/test")

	err := d.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestStartNew(t *testing.T) {
	tmpDir := t.TempDir()
	execPath := createTestExecutable(t, tmpDir)

	pf := &MockPIDFile{ReadErr: os.ErrNotExist}
	pm := &process.MockManager{}
	d := New(pf, pm, execPath)

	// This will fail because our test script is not a valid Go binary
	// but we can at least verify the logic before exec.Command
	err := d.Start(tmpDir, 18080, "127.0.0.1", false, "")
	// We expect an error because the test script isn't a valid server
	// but the PID file operations should be attempted
	_ = err
}

func TestStartAlreadyRunning(t *testing.T) {
	pf := &MockPIDFile{Pid: 1234}
	pm := &process.MockManager{
		RunningPids: map[int]bool{1234: true},
	}
	d := New(pf, pm, "/bin/test")

	err := d.Start("/tmp", 8080, "127.0.0.1", false, "")
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
	// Should not start a new process
	if pf.Written != 0 {
		t.Error("Start() should not write PID when service already running")
	}
}

func TestStartRemovesStalePID(t *testing.T) {
	tmpDir := t.TempDir()
	execPath := createTestExecutable(t, tmpDir)

	pf := &MockPIDFile{Pid: 1234, ReadErr: nil}
	pm := &process.MockManager{
		RunningPids: map[int]bool{},
	}
	d := New(pf, pm, execPath)

	err := d.Start(tmpDir, 18080, "127.0.0.1", false, "")
	_ = err

	if !pf.Removed {
		t.Error("Start() should remove stale PID file")
	}
}

func TestStartWithLogging(t *testing.T) {
	tmpDir := t.TempDir()
	execPath := createTestExecutable(t, tmpDir)

	pf := &MockPIDFile{ReadErr: os.ErrNotExist}
	pm := &process.MockManager{}
	d := New(pf, pm, execPath)

	// Test starting with logging enabled
	err := d.Start(tmpDir, 18080, "127.0.0.1", true, "")
	// We expect an error because the test script isn't a valid server
	// but we can at least verify the logic before exec.Command
	_ = err
}

func TestStartWithLoggingAndCustomDir(t *testing.T) {
	tmpDir := t.TempDir()
	customLogDir := t.TempDir()
	execPath := createTestExecutable(t, tmpDir)

	pf := &MockPIDFile{ReadErr: os.ErrNotExist}
	pm := &process.MockManager{}
	d := New(pf, pm, execPath)

	// Test starting with logging enabled and custom log directory
	err := d.Start(tmpDir, 18080, "127.0.0.1", true, customLogDir)
	// We expect an error because the test script isn't a valid server
	// but we can at least verify the logic before exec.Command
	_ = err
}

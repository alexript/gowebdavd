// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package process

// MockProcess implements Process interface for testing
type MockProcess struct {
	PidValue  int
	SignalErr error
	KillErr   error
	Signaled  bool
	Killed    bool
}

// Signal sends a signal to the process
func (m *MockProcess) Signal(sig int) error {
	m.Signaled = true
	return m.SignalErr
}

// Kill kills the process
func (m *MockProcess) Kill() error {
	m.Killed = true
	return m.KillErr
}

// Pid returns the process ID
func (m *MockProcess) Pid() int {
	return m.PidValue
}

// MockManager implements Manager for testing
type MockManager struct {
	RunningPids  map[int]bool
	FindErr      error
	TerminateErr error
	KillErr      error
	FoundProcess Process
}

// IsRunning checks if a process is running
func (m *MockManager) IsRunning(pid int) bool {
	return m.RunningPids[pid]
}

// FindProcess finds a process by PID
func (m *MockManager) FindProcess(pid int) (Process, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	if m.FoundProcess != nil {
		return m.FoundProcess, nil
	}
	return &MockProcess{PidValue: pid}, nil
}

// Terminate terminates a process
func (m *MockManager) Terminate(pid int) error {
	return m.TerminateErr
}

// Kill kills a process
func (m *MockManager) Kill(pid int) error {
	return m.KillErr
}

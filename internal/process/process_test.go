package process

import (
	"testing"
)

// testError for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestMockManagerIsRunning(t *testing.T) {
	mgr := &MockManager{
		RunningPids: map[int]bool{
			1234: true,
			5678: false,
		},
	}

	if !mgr.IsRunning(1234) {
		t.Error("Expected process 1234 to be running")
	}

	if mgr.IsRunning(5678) {
		t.Error("Expected process 5678 to not be running")
	}

	if mgr.IsRunning(9999) {
		t.Error("Expected process 9999 to not be running")
	}
}

func TestMockManagerFindProcess(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mgr := &MockManager{
			FoundProcess: &MockProcess{PidValue: 1234},
		}

		proc, err := mgr.FindProcess(1234)
		if err != nil {
			t.Errorf("FindProcess() error = %v", err)
		}
		if proc.Pid() != 1234 {
			t.Errorf("FindProcess() pid = %d, want 1234", proc.Pid())
		}
	})

	t.Run("error", func(t *testing.T) {
		mgr := &MockManager{
			FindErr: &testError{msg: "test error"},
		}

		_, err := mgr.FindProcess(1234)
		if err == nil {
			t.Error("FindProcess() should return error")
		}
	})
}

func TestMockProcess(t *testing.T) {
	proc := &MockProcess{PidValue: 1234}

	if proc.Pid() != 1234 {
		t.Errorf("Pid() = %d, want 1234", proc.Pid())
	}

	err := proc.Signal(15)
	if err != nil {
		t.Errorf("Signal() error = %v", err)
	}
	if !proc.Signaled {
		t.Error("Signal() should set Signaled flag")
	}

	err = proc.Kill()
	if err != nil {
		t.Errorf("Kill() error = %v", err)
	}
	if !proc.Killed {
		t.Error("Kill() should set Killed flag")
	}
}

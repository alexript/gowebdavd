package pidfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	pf := New()
	if pf == nil {
		t.Fatal("New() returned nil")
	}

	path := pf.Path()
	if path == "" {
		t.Error("Path() returned empty string")
	}

	expectedPath := filepath.Join(os.TempDir(), "gowebdavd.pid")
	if path != expectedPath {
		t.Errorf("Path() = %s, want %s", path, expectedPath)
	}
}

func TestNewWithPath(t *testing.T) {
	customPath := "/custom/path/test.pid"
	pf := NewWithPath(customPath)
	if pf == nil {
		t.Fatal("NewWithPath() returned nil")
	}

	if pf.Path() != customPath {
		t.Errorf("Path() = %s, want %s", pf.Path(), customPath)
	}
}

func TestFileWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	pf := &file{path: filepath.Join(tmpDir, "test.pid")}

	// Test Write
	testPID := 12345
	err := pf.Write(testPID)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Test Read
	pid, err := pf.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if pid != testPID {
		t.Errorf("Read() = %d, want %d", pid, testPID)
	}
}

func TestFileReadNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	pf := &file{path: filepath.Join(tmpDir, "nonexistent.pid")}

	_, err := pf.Read()
	if err == nil {
		t.Error("Read() should return error for non-existent file")
	}
}

func TestFileReadInvalidPID(t *testing.T) {
	tmpDir := t.TempDir()
	pf := &file{path: filepath.Join(tmpDir, "test.pid")}

	// Write invalid content
	err := os.WriteFile(pf.path, []byte("invalid"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = pf.Read()
	if err == nil {
		t.Error("Read() should return error for invalid PID")
	}
}

func TestFileRemove(t *testing.T) {
	tmpDir := t.TempDir()
	pf := &file{path: filepath.Join(tmpDir, "test.pid")}

	// Create file
	err := pf.Write(12345)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(pf.path); os.IsNotExist(err) {
		t.Fatal("File should exist after Write")
	}

	// Remove file
	err = pf.Remove()
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// Verify file removed
	if _, err := os.Stat(pf.path); !os.IsNotExist(err) {
		t.Error("File should not exist after Remove")
	}
}

func TestFileRemoveNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	pf := &file{path: filepath.Join(tmpDir, "nonexistent.pid")}

	err := pf.Remove()
	if err == nil {
		t.Error("Remove() should return error for non-existent file")
	}
}

func TestFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	expectedPath := filepath.Join(tmpDir, "test.pid")
	pf := &file{path: expectedPath}

	path := pf.Path()
	if path != expectedPath {
		t.Errorf("Path() = %s, want %s", path, expectedPath)
	}
}

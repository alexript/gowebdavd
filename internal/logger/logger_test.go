// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNew_Disabled(t *testing.T) {
	logger, err := New(false, "")
	if err != nil {
		t.Fatalf("New(false, \"\") error = %v", err)
	}
	if logger.Enabled() {
		t.Error("Expected logger to be disabled")
	}
	if err := logger.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestNew_EnabledDefaultDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override getLogDir to use temp directory
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", tempDir)
		os.Setenv("LOCALAPPDATA", filepath.Join(tempDir, "AppData", "Local"))
	} else {
		os.Setenv("HOME", tempDir)
	}

	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", originalUserProfile)
			os.Unsetenv("LOCALAPPDATA")
		} else {
			os.Setenv("HOME", originalHome)
		}
	}()

	logger, err := New(true, "")
	if err != nil {
		t.Fatalf("New(true, \"\") error = %v", err)
	}
	defer logger.Close()

	if !logger.Enabled() {
		t.Error("Expected logger to be enabled")
	}
}

func TestNew_CustomDir(t *testing.T) {
	// Create a custom log directory
	customDir := t.TempDir()

	logger, err := New(true, customDir)
	if err != nil {
		t.Fatalf("New(true, customDir) error = %v", err)
	}
	defer logger.Close()

	if !logger.Enabled() {
		t.Error("Expected logger to be enabled")
	}
}

func TestNew_CustomDirNotExists(t *testing.T) {
	// Use a non-existent directory
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")

	logger, err := New(true, nonExistentDir)
	if err == nil {
		logger.Close()
		t.Fatal("Expected error for non-existent directory")
	}

	expectedErr := "log directory does not exist"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr, err)
	}
}

func TestNew_CustomDirIsFile(t *testing.T) {
	// Create a file instead of directory
	tempDir := t.TempDir()
	notADir := filepath.Join(tempDir, "notadir")
	if err := os.WriteFile(notADir, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	logger, err := New(true, notADir)
	if err == nil {
		logger.Close()
		t.Fatal("Expected error when path is not a directory")
	}

	expectedErr := "log path is not a directory"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr, err)
	}
}

func TestMiddleware_Disabled(t *testing.T) {
	logger := NewNopLogger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := logger.Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestMiddleware_Enabled(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, true)
	defer logger.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := logger.Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected log output, got empty string")
	}

	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected log to contain 'GET'")
	}

	if !strings.Contains(logOutput, "/test") {
		t.Error("Expected log to contain '/test'")
	}
}

func TestResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusNotFound)

	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.statusCode)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected recorder code %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestCleanupOldLogs(t *testing.T) {
	tempDir := t.TempDir()

	// Create a recent log file
	recentFile := filepath.Join(tempDir, "gowebdavd_recent.log")
	if err := os.WriteFile(recentFile, []byte("recent"), 0644); err != nil {
		t.Fatalf("Failed to create recent log file: %v", err)
	}

	// Create an old log file (2 months ago)
	oldFile := filepath.Join(tempDir, "gowebdavd_old.log")
	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}

	// Set modification time to 2 months ago
	oldTime := time.Now().AddDate(0, -2, 0)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old file time: %v", err)
	}

	// Run cleanup
	if err := cleanupOldLogs(tempDir); err != nil {
		t.Fatalf("cleanupOldLogs error = %v", err)
	}

	// Check that old file was removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Expected old log file to be removed")
	}

	// Check that recent file still exists
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Error("Expected recent log file to exist")
	}
}

func TestCleanupOldLogs_SkipNonLogFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create a non-log file
	nonLogFile := filepath.Join(tempDir, "other.txt")
	if err := os.WriteFile(nonLogFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create non-log file: %v", err)
	}

	// Set modification time to 2 months ago
	oldTime := time.Now().AddDate(0, -2, 0)
	if err := os.Chtimes(nonLogFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set old file time: %v", err)
	}

	// Run cleanup
	if err := cleanupOldLogs(tempDir); err != nil {
		t.Fatalf("cleanupOldLogs error = %v", err)
	}

	// Check that non-log file still exists
	if _, err := os.Stat(nonLogFile); os.IsNotExist(err) {
		t.Error("Expected non-log file to exist")
	}
}

func TestCleanupOldLogs_NonExistentDir(t *testing.T) {
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")

	// Should not error for non-existent directory
	if err := cleanupOldLogs(nonExistentDir); err != nil {
		t.Errorf("cleanupOldLogs error = %v", err)
	}
}

func TestGetLogDir(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	originalLocalAppData := os.Getenv("LOCALAPPDATA")

	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
		if originalLocalAppData != "" {
			os.Setenv("LOCALAPPDATA", originalLocalAppData)
		} else {
			os.Unsetenv("LOCALAPPDATA")
		}
	}()

	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", tempDir)
		os.Setenv("LOCALAPPDATA", filepath.Join(tempDir, "AppData", "Local"))

		logDir, err := getLogDir()
		if err != nil {
			t.Fatalf("getLogDir error = %v", err)
		}

		expected := filepath.Join(tempDir, "AppData", "Local", "gowebdavd", "logs")
		if logDir != expected {
			t.Errorf("Expected %s, got %s", expected, logDir)
		}
	} else {
		os.Setenv("HOME", tempDir)

		logDir, err := getLogDir()
		if err != nil {
			t.Fatalf("getLogDir error = %v", err)
		}

		expected := filepath.Join(tempDir, ".local", "share", "gowebdavd", "logs")
		if logDir != expected {
			t.Errorf("Expected %s, got %s", expected, logDir)
		}
	}
}

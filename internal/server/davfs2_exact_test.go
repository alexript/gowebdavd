// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMoveWithoutLockToken проверяет MOVE без токена (как в реальном davfs2)
func TestMoveWithoutLockToken(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем .git директорию
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Создаем файл для тестирования
	lockFile := filepath.Join(gitDir, "config.lock")
	if err := os.WriteFile(lockFile, []byte("[core]\n"), 0644); err != nil {
		t.Fatalf("Failed to create lock file: %v", err)
	}

	// MOVE без токена и без If заголовка
	req := httptest.NewRequest("MOVE", "/.git/config.lock", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Destination", "http://127.0.0.1:8080/.git/config")
	req.Header.Set("Overwrite", "T")
	// НЕ устанавливаем If заголовок!

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	t.Logf("MOVE without lock token status: %d", rec.Code)
	t.Logf("MOVE response: %s", rec.Body.String())

	// Проверяем результат
	configPath := filepath.Join(tempDir, ".git", "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Log("config file was not created (MOVE failed)")
	} else {
		t.Log("config file was created successfully")
		content, _ := os.ReadFile(configPath)
		t.Logf("Content: %s", string(content))
	}
}

// TestMoveWithUnlockedDestination проверяет MOVE когда назначение не заблокировано
func TestMoveWithUnlockedDestination(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем файл и берем lock
	sourceFile := filepath.Join(tempDir, "locked.txt")
	if err := os.WriteFile(sourceFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// LOCK
	lockReq := httptest.NewRequest("LOCK", "/locked.txt", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
	lockReq.Header.Set("Content-Type", "text/xml")

	lockRec := httptest.NewRecorder()
	handler.ServeHTTP(lockRec, lockReq)

	lockToken := lockRec.Header().Get("Lock-Token")
	t.Logf("Lock token: %s", lockToken)

	// MOVE без If заголовка (проверяем работает ли)
	moveReq := httptest.NewRequest("MOVE", "/locked.txt", nil)
	moveReq.Host = "127.0.0.1:8080"
	moveReq.Header.Set("Destination", "http://127.0.0.1:8080/unlocked.txt")
	// НЕ устанавливаем If заголовок

	moveRec := httptest.NewRecorder()
	handler.ServeHTTP(moveRec, moveReq)

	t.Logf("MOVE without If header status: %d", moveRec.Code)

	if moveRec.Code == http.StatusLocked {
		t.Log("MOVE correctly returned 423 because source is locked")
	} else if moveRec.Code == http.StatusCreated || moveRec.Code == http.StatusNoContent {
		t.Log("MOVE succeeded without lock token - this is unexpected")
	}
}

// TestDavfs2ExactLogScenario воспроизводит точно логи davfs2
func TestDavfs2ExactLogScenario(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем .git директорию
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	var lockToken string

	// LOCK /t/tt/.git/config.lock
	t.Run("lock", func(t *testing.T) {
		req := httptest.NewRequest("LOCK", "/.git/config.lock", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
  <owner>git</owner>
</lockinfo>`))
		req.Header.Set("Content-Type", "text/xml; charset=utf-8")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		lockToken = rec.Header().Get("Lock-Token")
		t.Logf("LOCK /t/tt/.git/config.lock %d", rec.Code)
	})

	// HEAD /t/tt/.git/config.lock
	t.Run("head", func(t *testing.T) {
		req := httptest.NewRequest("HEAD", "/.git/config.lock", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		t.Logf("HEAD /t/tt/.git/config.lock %d", rec.Code)
	})

	// PUT /t/tt/.git/config.lock
	t.Run("put", func(t *testing.T) {
		config := `[core]
repositoryformatversion = 0
`
		req := httptest.NewRequest("PUT", "/.git/config.lock", strings.NewReader(config))
		req.Header.Set("Content-Type", "text/plain")
		if lockToken != "" {
			req.Header.Set("If", fmt.Sprintf("(%s)", lockToken))
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		t.Logf("PUT /t/tt/.git/config.lock %d", rec.Code)
	})

	// MOVE /t/tt/.git/config.lock (как в логах: 412)
	t.Run("move", func(t *testing.T) {
		req := httptest.NewRequest("MOVE", "/.git/config.lock", nil)
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Destination", "http://127.0.0.1:8080/.git/config")
		// davfs2 отправляет If заголовок с токеном
		if lockToken != "" {
			req.Header.Set("If", fmt.Sprintf("(%s)", lockToken))
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		t.Logf("MOVE /t/tt/.git/config.lock %d (expected: 412)", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	})

	// DELETE /t/tt/.git/config.lock
	t.Run("delete", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/.git/config.lock", nil)
		if lockToken != "" {
			req.Header.Set("If", fmt.Sprintf("(%s)", lockToken))
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		t.Logf("DELETE /t/tt/.git/config.lock %d", rec.Code)
	})
}

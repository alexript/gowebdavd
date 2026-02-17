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

// TestGitInitWithNoLockSystem полный сценарий git init с NoOpLockSystem
func TestGitInitWithNoLockSystem(t *testing.T) {
	tempDir := t.TempDir()
	// Создаем сервер с NoOpLockSystem
	srv := NewWithLockSystem(tempDir, 8080, "127.0.0.1", nil, true)
	handler := srv.Handler()

	// Создаем .git директорию
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	var lockToken string

	// Шаг 1: LOCK config.lock (теперь всегда успешно)
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
		t.Logf("Lock token: %s", lockToken)
		t.Logf("LOCK status: %d", rec.Code)

		if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
			t.Fatalf("LOCK failed with status %d", rec.Code)
		}
	})

	// Шаг 2: HEAD config.lock
	t.Run("head", func(t *testing.T) {
		req := httptest.NewRequest("HEAD", "/.git/config.lock", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		t.Logf("HEAD status: %d", rec.Code)
		if rec.Code != http.StatusOK {
			t.Errorf("HEAD failed with status %d", rec.Code)
		}
	})

	// Шаг 3: PUT config.lock
	t.Run("put", func(t *testing.T) {
		config := `[core]
repositoryformatversion = 0
filemode = true
bare = false
logallrefupdates = true
`
		req := httptest.NewRequest("PUT", "/.git/config.lock", strings.NewReader(config))
		req.Header.Set("Content-Type", "text/plain")
		if lockToken != "" {
			req.Header.Set("If", fmt.Sprintf("(%s)", lockToken))
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		t.Logf("PUT status: %d", rec.Code)
		if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
			t.Errorf("PUT failed with status %d", rec.Code)
		}
	})

	// Шаг 4: MOVE config.lock -> config (теперь должно работать!)
	t.Run("move", func(t *testing.T) {
		req := httptest.NewRequest("MOVE", "/.git/config.lock", nil)
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Destination", "http://127.0.0.1:8080/.git/config")
		req.Header.Set("Overwrite", "T")
		if lockToken != "" {
			req.Header.Set("If", fmt.Sprintf("(%s)", lockToken))
			t.Logf("Using If header: (%s)", lockToken)
		}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		t.Logf("MOVE status: %d", rec.Code)
		t.Logf("MOVE response: %s", rec.Body.String())

		// С NoOpLockSystem MOVE должен работать
		if rec.Code != http.StatusCreated && rec.Code != http.StatusNoContent {
			t.Errorf("MOVE failed with status %d", rec.Code)
		}
	})

	// Шаг 5: Проверяем результат
	t.Run("verify", func(t *testing.T) {
		configPath := filepath.Join(tempDir, ".git", "config")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		} else {
			content, _ := os.ReadFile(configPath)
			t.Logf("config file created:\n%s", string(content))
		}

		lockPath := filepath.Join(tempDir, ".git", "config.lock")
		if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
			t.Error("config.lock file still exists (should be moved)")
		}
	})
}

// TestMultipleLocksWithNoLockSystem проверяет множественные операции с блокировками
func TestMultipleLocksWithNoLockSystem(t *testing.T) {
	tempDir := t.TempDir()
	srv := NewWithLockSystem(tempDir, 8080, "127.0.0.1", nil, true)
	handler := srv.Handler()

	// Создаем несколько файлов с lock'ами
	files := []string{"file1.lock", "file2.lock", "file3.lock"}

	for _, filename := range files {
		t.Run(filename, func(t *testing.T) {
			// LOCK
			lockReq := httptest.NewRequest("LOCK", "/"+filename, strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
			lockReq.Header.Set("Content-Type", "text/xml")

			lockRec := httptest.NewRecorder()
			handler.ServeHTTP(lockRec, lockReq)

			token := lockRec.Header().Get("Lock-Token")
			t.Logf("LOCK %s: %d, token: %s", filename, lockRec.Code, token)

			// PUT
			putReq := httptest.NewRequest("PUT", "/"+filename, strings.NewReader("content"))
			putReq.Header.Set("Content-Type", "text/plain")
			putReq.Header.Set("If", fmt.Sprintf("(%s)", token))

			putRec := httptest.NewRecorder()
			handler.ServeHTTP(putRec, putReq)

			t.Logf("PUT %s: %d", filename, putRec.Code)

			// MOVE
			destName := strings.Replace(filename, ".lock", "", 1)
			moveReq := httptest.NewRequest("MOVE", "/"+filename, nil)
			moveReq.Header.Set("Destination", "/"+destName)
			moveReq.Header.Set("If", fmt.Sprintf("(%s)", token))

			moveRec := httptest.NewRecorder()
			handler.ServeHTTP(moveRec, moveReq)

			t.Logf("MOVE %s -> %s: %d", filename, destName, moveRec.Code)
		})
	}
}

// TestNoLockComparison сравнивает поведение с и без NoOpLockSystem
func TestNoLockComparison(t *testing.T) {
	tempDir1 := t.TempDir()
	tempDir2 := t.TempDir()

	// Обычный сервер
	normalSrv := New(tempDir1, 8080, "127.0.0.1", nil)
	// Сервер без блокировок
	noLockSrv := NewWithLockSystem(tempDir2, 8080, "127.0.0.1", nil, true)

	// Создаем тестовые файлы
	normalFile := filepath.Join(tempDir1, "test.lock")
	noLockFile := filepath.Join(tempDir2, "test.lock")
	os.WriteFile(normalFile, []byte("content"), 0644)
	os.WriteFile(noLockFile, []byte("content"), 0644)

	// Получаем lock'и
	normalLockReq := httptest.NewRequest("LOCK", "/test.lock", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
	normalLockReq.Header.Set("Content-Type", "text/xml")
	normalLockRec := httptest.NewRecorder()
	normalSrv.Handler().ServeHTTP(normalLockRec, normalLockReq)

	noLockLockReq := httptest.NewRequest("LOCK", "/test.lock", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
	noLockLockReq.Header.Set("Content-Type", "text/xml")
	noLockLockRec := httptest.NewRecorder()
	noLockSrv.Handler().ServeHTTP(noLockLockRec, noLockLockReq)

	normalToken := normalLockRec.Header().Get("Lock-Token")
	noLockToken := noLockLockRec.Header().Get("Lock-Token")

	t.Logf("Normal lock token: %s", normalToken)
	t.Logf("NoLock token: %s", noLockToken)

	// Пробуем MOVE
	t.Run("normal_server_move", func(t *testing.T) {
		req := httptest.NewRequest("MOVE", "/test.lock", nil)
		req.Header.Set("Destination", "/dest.txt")
		req.Header.Set("If", fmt.Sprintf("(%s)", normalToken))

		rec := httptest.NewRecorder()
		normalSrv.Handler().ServeHTTP(rec, req)

		t.Logf("Normal server MOVE status: %d", rec.Code)
	})

	t.Run("nolock_server_move", func(t *testing.T) {
		req := httptest.NewRequest("MOVE", "/test.lock", nil)
		req.Header.Set("Destination", "/dest.txt")
		req.Header.Set("If", fmt.Sprintf("(%s)", noLockToken))

		rec := httptest.NewRecorder()
		noLockSrv.Handler().ServeHTTP(rec, req)

		t.Logf("NoLock server MOVE status: %d", rec.Code)
	})
}

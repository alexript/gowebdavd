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

// TestMoveWhenDestinationExists проверяет MOVE когда целевой файл существует
func TestMoveWhenDestinationExists(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем оба файла
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")
	os.WriteFile(srcFile, []byte("source content"), 0644)
	os.WriteFile(dstFile, []byte("dest content"), 0644)

	// MOVE без токенов
	req := httptest.NewRequest("MOVE", "/source.txt", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Destination", "http://127.0.0.1:8080/dest.txt")
	req.Header.Set("Overwrite", "T")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	t.Logf("MOVE when dst exists status: %d", rec.Code)
}

// TestMoveWhenDestinationDoesNotExist проверяет MOVE когда целевой файл НЕ существует
func TestMoveWhenDestinationDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем только src
	srcFile := filepath.Join(tempDir, "source.txt")
	os.WriteFile(srcFile, []byte("source content"), 0644)

	// MOVE без токенов
	req := httptest.NewRequest("MOVE", "/source.txt", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Destination", "http://127.0.0.1:8080/dest.txt")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	t.Logf("MOVE when dst NOT exists status: %d", rec.Code)
}

// TestMoveWithBothLocked проверяет MOVE когда оба файла заблокированы
func TestMoveWithBothLocked(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем оба файла
	srcFile := filepath.Join(tempDir, "src.txt")
	dstFile := filepath.Join(tempDir, "dst.txt")
	os.WriteFile(srcFile, []byte("src"), 0644)
	os.WriteFile(dstFile, []byte("dst"), 0644)

	// Блокируем src
	srcLockReq := httptest.NewRequest("LOCK", "/src.txt", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
	srcLockReq.Header.Set("Content-Type", "text/xml")
	srcLockRec := httptest.NewRecorder()
	handler.ServeHTTP(srcLockRec, srcLockReq)
	srcToken := srcLockRec.Header().Get("Lock-Token")

	// Блокируем dst
	dstLockReq := httptest.NewRequest("LOCK", "/dst.txt", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
	dstLockReq.Header.Set("Content-Type", "text/xml")
	dstLockRec := httptest.NewRecorder()
	handler.ServeHTTP(dstLockRec, dstLockReq)
	dstToken := dstLockRec.Header().Get("Lock-Token")

	t.Logf("Src token: %s", srcToken)
	t.Logf("Dst token: %s", dstToken)

	// MOVE с обоими токенами
	moveReq := httptest.NewRequest("MOVE", "/src.txt", nil)
	moveReq.Host = "127.0.0.1:8080"
	moveReq.Header.Set("Destination", "http://127.0.0.1:8080/dst.txt")
	moveReq.Header.Set("Overwrite", "T")
	// Если заголовок If содержит оба токена?
	if srcToken != "" && dstToken != "" {
		moveReq.Header.Set("If", fmt.Sprintf("(%s) (%s)", srcToken, dstToken))
		t.Logf("If header: (%s) (%s)", srcToken, dstToken)
	}

	moveRec := httptest.NewRecorder()
	handler.ServeHTTP(moveRec, moveReq)

	t.Logf("MOVE with both locks status: %d", moveRec.Code)
	t.Logf("Response: %s", moveRec.Body.String())
}

// TestMoveWithOnlySourceLocked проверяет MOVE когда только src заблокирован
func TestMoveWithOnlySourceLocked(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем оба файла
	srcFile := filepath.Join(tempDir, "src.txt")
	dstFile := filepath.Join(tempDir, "dst.txt")
	os.WriteFile(srcFile, []byte("src"), 0644)
	os.WriteFile(dstFile, []byte("dst"), 0644)

	// Блокируем только src
	lockReq := httptest.NewRequest("LOCK", "/src.txt", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
	lockReq.Header.Set("Content-Type", "text/xml")
	lockRec := httptest.NewRecorder()
	handler.ServeHTTP(lockRec, lockReq)
	token := lockRec.Header().Get("Lock-Token")

	t.Logf("Lock token: %s", token)

	// MOVE с токеном только для src
	moveReq := httptest.NewRequest("MOVE", "/src.txt", nil)
	moveReq.Host = "127.0.0.1:8080"
	moveReq.Header.Set("Destination", "http://127.0.0.1:8080/dst.txt")
	moveReq.Header.Set("Overwrite", "T")
	if token != "" {
		moveReq.Header.Set("If", fmt.Sprintf("(%s)", token))
		t.Logf("If header: (%s)", token)
	}

	moveRec := httptest.NewRecorder()
	handler.ServeHTTP(moveRec, moveReq)

	t.Logf("MOVE with only src locked status: %d", moveRec.Code)
	t.Logf("Response: %s", moveRec.Body.String())

	if moveRec.Code == http.StatusPreconditionFailed {
		t.Log("MOVE failed with 412 because dst exists and may have different lock")
	}
}

// TestMoveGitScenarioWithOverwrite проверяет сценарий git с Overwrite: T
func TestMoveGitScenarioWithOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем .git директорию
	gitDir := filepath.Join(tempDir, ".git")
	os.MkdirAll(gitDir, 0755)

	var lockToken string

	// LOCK config.lock
	lockReq := httptest.NewRequest("LOCK", "/.git/config.lock", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
	lockReq.Header.Set("Content-Type", "text/xml")
	lockRec := httptest.NewRecorder()
	handler.ServeHTTP(lockRec, lockReq)
	lockToken = lockRec.Header().Get("Lock-Token")

	// PUT config.lock
	putReq := httptest.NewRequest("PUT", "/.git/config.lock", strings.NewReader("[core]\n"))
	putReq.Header.Set("Content-Type", "text/plain")
	if lockToken != "" {
		putReq.Header.Set("If", fmt.Sprintf("(%s)", lockToken))
	}
	putRec := httptest.NewRecorder()
	handler.ServeHTTP(putRec, putReq)

	t.Logf("PUT status: %d", putRec.Code)

	// MOVE с Overwrite: T и без If заголовка
	moveReq := httptest.NewRequest("MOVE", "/.git/config.lock", nil)
	moveReq.Host = "127.0.0.1:8080"
	moveReq.Header.Set("Destination", "http://127.0.0.1:8080/.git/config")
	moveReq.Header.Set("Overwrite", "T")
	// НЕ добавляем If заголовок!

	moveRec := httptest.NewRecorder()
	handler.ServeHTTP(moveRec, moveReq)

	t.Logf("MOVE without If header status: %d", moveRec.Code)

	if moveRec.Code == http.StatusLocked {
		t.Log("Server requires lock token for MOVE (returns 423)")
	} else if moveRec.Code == http.StatusCreated || moveRec.Code == http.StatusNoContent {
		t.Log("MOVE succeeded without lock token")
	}
}

// TestDavfs2Workaround проверяет обходное решение для davfs2
func TestDavfs2Workaround(t *testing.T) {
	tempDir := t.TempDir()
	server := New(tempDir, 8080, "127.0.0.1", nil)
	handler := server.Handler()

	// Создаем .git директорию
	gitDir := filepath.Join(tempDir, ".git")
	os.MkdirAll(gitDir, 0755)

	var lockToken string

	// LOCK config.lock
	lockReq := httptest.NewRequest("LOCK", "/.git/config.lock", strings.NewReader(`
<lockinfo xmlns="DAV:">
  <lockscope><exclusive/></lockscope>
  <locktype><write/></locktype>
</lockinfo>`))
	lockReq.Header.Set("Content-Type", "text/xml")
	lockRec := httptest.NewRecorder()
	handler.ServeHTTP(lockRec, lockReq)
	lockToken = lockRec.Header().Get("Lock-Token")

	// PUT config.lock
	putReq := httptest.NewRequest("PUT", "/.git/config.lock", strings.NewReader("[core]\n"))
	putReq.Header.Set("Content-Type", "text/plain")
	if lockToken != "" {
		putReq.Header.Set("If", fmt.Sprintf("(%s)", lockToken))
	}
	putRec := httptest.NewRecorder()
	handler.ServeHTTP(putRec, putReq)

	// Попытка 1: MOVE с If заголовком (как делает davfs2) → 412
	move1Req := httptest.NewRequest("MOVE", "/.git/config.lock", nil)
	move1Req.Host = "127.0.0.1:8080"
	move1Req.Header.Set("Destination", "http://127.0.0.1:8080/.git/config")
	move1Req.Header.Set("Overwrite", "T")
	if lockToken != "" {
		move1Req.Header.Set("If", fmt.Sprintf("(%s)", lockToken))
	}
	move1Rec := httptest.NewRecorder()
	handler.ServeHTTP(move1Rec, move1Req)
	t.Logf("MOVE with If header: %d", move1Rec.Code)

	// Попытка 2: UNLOCK + MOVE
	unlockReq := httptest.NewRequest("UNLOCK", "/.git/config.lock", nil)
	unlockReq.Header.Set("Lock-Token", lockToken)
	unlockRec := httptest.NewRecorder()
	handler.ServeHTTP(unlockRec, unlockReq)
	t.Logf("UNLOCK: %d", unlockRec.Code)

	move2Req := httptest.NewRequest("MOVE", "/.git/config.lock", nil)
	move2Req.Host = "127.0.0.1:8080"
	move2Req.Header.Set("Destination", "http://127.0.0.1:8080/.git/config")
	move2Req.Header.Set("Overwrite", "T")
	move2Rec := httptest.NewRecorder()
	handler.ServeHTTP(move2Rec, move2Req)
	t.Logf("MOVE after UNLOCK: %d", move2Rec.Code)

	// Проверяем результат
	configPath := filepath.Join(tempDir, ".git", "config")
	if _, err := os.Stat(configPath); err == nil {
		t.Log("Success! config file created")
	} else {
		t.Log("config file not created")
	}
}

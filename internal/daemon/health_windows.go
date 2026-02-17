//go:build windows

// Copyright (c) 2026 gowebdavd contributors
// SPDX-License-Identifier: MIT

package daemon

import (
	"fmt"
	"net/http"
	"time"
)

// waitForService waits for the service to respond with 200 OK
func waitForService(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for service to become ready")
}

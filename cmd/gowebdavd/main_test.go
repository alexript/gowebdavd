package main

import (
	"testing"
)

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid port 1", 1, false},
		{"valid port 80", 80, false},
		{"valid port 8080", 8080, false},
		{"valid port 65535", 65535, false},
		{"invalid port 0", 0, true},
		{"invalid port -1", -1, true},
		{"invalid port 65536", 65536, true},
		{"invalid port 99999", 99999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePort(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

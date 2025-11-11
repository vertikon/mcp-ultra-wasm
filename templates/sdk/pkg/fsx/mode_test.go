// pkg/fsx/mode_test.go
package fsx

import (
	"os"
	"testing"
)

func TestFileModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant os.FileMode
		expected os.FileMode
	}{
		{"FileModeUserRW", FileModeUserRW, 0o600},
		{"FileModeUserRWGroupR", FileModeUserRWGroupR, 0o640},
		{"FileModeUserRWXGroupRX", FileModeUserRWXGroupRX, 0o750},
		{"FileModePublicRead", FileModePublicRead, 0o644},
		{"FileModeDirUserRWX", FileModeDirUserRWX, 0o700},
		{"FileModeDirPublic", FileModeDirPublic, 0o755},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %#o, want %#o", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestFileModeOctalRepresentation(t *testing.T) {
	// Test that constants are in octal (not decimal)
	if FileModeUserRW != 0o600 {
		t.Errorf("FileModeUserRW should be 0o600 (octal), got %#o", FileModeUserRW)
	}

	if FileModeDirPublic != 0o755 {
		t.Errorf("FileModeDirPublic should be 0o755 (octal), got %#o", FileModeDirPublic)
	}
}

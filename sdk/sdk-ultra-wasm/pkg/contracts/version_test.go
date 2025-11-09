// pkg/contracts/version_test.go
package contracts_test

import (
	"testing"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/contracts"
)

func TestCompatibleWith(t *testing.T) {
	tests := []struct {
		name           string
		pluginVersion  string
		wantCompatible bool
	}{
		{"same version", "1.0.0", true},
		{"different patch", "1.0.1", true},
		{"different minor", "1.1.0", true},
		{"different major", "2.0.0", false},
		{"older major", "0.9.0", false},
		{"invalid version", "", false},
		{"partial version", "1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contracts.CompatibleWith(tt.pluginVersion)
			if got != tt.wantCompatible {
				t.Errorf("CompatibleWith(%q) = %v, want %v",
					tt.pluginVersion, got, tt.wantCompatible)
			}
		})
	}
}

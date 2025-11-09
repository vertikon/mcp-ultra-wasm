// pkg/httpx/status_test.go
package httpx

import "testing"

func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"StatusOK", StatusOK, 200},
		{"StatusNoContent", StatusNoContent, 204},
		{"StatusBadRequest", StatusBadRequest, 400},
		{"StatusUnauthorized", StatusUnauthorized, 401},
		{"StatusForbidden", StatusForbidden, 403},
		{"StatusInternalServerError", StatusInternalServerError, 500},
		{"StatusBadGateway", StatusBadGateway, 502},
		{"StatusServiceUnavailable", StatusServiceUnavailable, 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

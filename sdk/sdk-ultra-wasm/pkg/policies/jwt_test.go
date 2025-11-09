// pkg/policies/jwt_test.go
package policies

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

var errInvalidToken = errors.New("invalid token")

type mockTokenValidator struct {
	validateFunc func(raw string) (subject string, roles []string, err error)
}

func (m *mockTokenValidator) Validate(raw string) (subject string, roles []string, err error) {
	return m.validateFunc(raw)
}

func TestAuthValidToken(t *testing.T) {
	validator := &mockTokenValidator{
		validateFunc: func(raw string) (string, []string, error) {
			if raw == "valid-token" {
				return "user123", []string{"admin", "editor"}, nil
			}
			return "", nil, errInvalidToken
		},
	}

	middleware := Auth(validator)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := FromIdentity(r.Context())
		if identity == nil {
			t.Error("Identity not set in context")
			return
		}

		if identity.Subject != "user123" {
			t.Errorf("Expected subject 'user123', got '%s'", identity.Subject)
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthInvalidToken(t *testing.T) {
	validator := &mockTokenValidator{
		validateFunc: func(_ string) (string, []string, error) {
			return "", nil, errInvalidToken
		},
	}

	middleware := Auth(validator)

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("Handler should not be called for invalid token")
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMissingBearer(t *testing.T) {
	validator := &mockTokenValidator{
		validateFunc: func(_ string) (string, []string, error) {
			return "user123", []string{"admin"}, nil
		},
	}

	middleware := Auth(validator)

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("Handler should not be called without Bearer prefix")
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "just-a-token")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMissingHeader(t *testing.T) {
	validator := &mockTokenValidator{
		validateFunc: func(_ string) (string, []string, error) {
			return "user123", []string{"admin"}, nil
		},
	}

	middleware := Auth(validator)

	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("Handler should not be called without Authorization header")
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthRolesPassedToContext(t *testing.T) {
	expectedRoles := []string{"admin", "editor", "viewer"}

	validator := &mockTokenValidator{
		validateFunc: func(_ string) (string, []string, error) {
			return "user456", expectedRoles, nil
		},
	}

	middleware := Auth(validator)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := FromIdentity(r.Context())
		if identity == nil {
			t.Fatal("Identity not set in context")
		}

		if len(identity.Roles) != len(expectedRoles) {
			t.Errorf("Expected %d roles, got %d", len(expectedRoles), len(identity.Roles))
		}

		for i, role := range expectedRoles {
			if identity.Roles[i] != role {
				t.Errorf("Expected role %s at index %d, got %s", role, i, identity.Roles[i])
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer my-token")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

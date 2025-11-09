// pkg/policies/rbac_test.go
package policies

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRoleWithPermission(t *testing.T) {
	middleware := RequireRole("admin")

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	wrapped := middleware(handler)

	ctx := WithIdentity(context.Background(), "user123", []string{"admin", "editor"})
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "ok" {
		t.Errorf("Expected body 'ok', got '%s'", w.Body.String())
	}
}

func TestRequireRoleWithoutPermission(t *testing.T) {
	middleware := RequireRole("superadmin")

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	ctx := WithIdentity(context.Background(), "user123", []string{"admin", "editor"})
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRequireRoleNoIdentity(t *testing.T) {
	middleware := RequireRole("admin")

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestRequireAnyRoleWithPermission(t *testing.T) {
	middleware := RequireAnyRole("admin", "superadmin", "moderator")

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	wrapped := middleware(handler)

	ctx := WithIdentity(context.Background(), "user123", []string{"editor", "moderator"})
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequireAnyRoleWithoutPermission(t *testing.T) {
	middleware := RequireAnyRole("admin", "superadmin")

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	ctx := WithIdentity(context.Background(), "user123", []string{"editor", "viewer"})
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRequireAnyRoleNoIdentity(t *testing.T) {
	middleware := RequireAnyRole("admin", "editor")

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

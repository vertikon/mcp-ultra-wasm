// pkg/policies/jwt.go
package policies

import (
	"net/http"
	"strings"
)

// TokenValidator valida tokens JWT
type TokenValidator interface {
	// Validate valida token e retorna subject e roles
	Validate(raw string) (subject string, roles []string, err error)
}

// Auth middleware de autenticação JWT
func Auth(tv TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "missing bearer token", StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			sub, roles, err := tv.Validate(token)
			if err != nil {
				http.Error(w, "invalid token", StatusUnauthorized)
				return
			}

			ctx := WithIdentity(r.Context(), sub, roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

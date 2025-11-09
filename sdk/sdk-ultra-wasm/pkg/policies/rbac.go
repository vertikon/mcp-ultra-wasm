// pkg/policies/rbac.go
package policies

import "net/http"

// RequireRole middleware que exige papel específico
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := FromIdentity(r.Context())
			if id == nil {
				http.Error(w, "unauthorized", StatusUnauthorized)
				return
			}

			for _, rr := range id.Roles {
				if rr == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "forbidden", StatusForbidden)
		})
	}
}

// RequireAnyRole middleware que exige qualquer um dos papéis
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := FromIdentity(r.Context())
			if id == nil {
				http.Error(w, "unauthorized", StatusUnauthorized)
				return
			}

			for _, userRole := range id.Roles {
				for _, allowedRole := range roles {
					if userRole == allowedRole {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			http.Error(w, "forbidden", StatusForbidden)
		})
	}
}

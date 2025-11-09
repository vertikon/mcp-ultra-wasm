// pkg/contracts/middleware.go
package contracts

import "net/http"

// MiddlewareInjector permite plugins registrarem middlewares
// SemVer: v1.0.0 - Interface estável
type MiddlewareInjector interface {
	// Name retorna identificador do middleware
	Name() string

	// Priority define ordem de execução (menor = primeiro)
	Priority() int

	// Middleware retorna a função middleware
	Middleware() func(http.Handler) http.Handler
}

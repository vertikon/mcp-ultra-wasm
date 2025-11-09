// pkg/contracts/route.go
package contracts

import "net/http"

// RouteInjector permite plugins registrarem rotas HTTP
// SemVer: v1.0.0 - Interface estável
type RouteInjector interface {
	// Name retorna identificador único do plugin
	Name() string

	// Version retorna versão SemVer do plugin
	Version() string

	// Routes retorna rotas a serem registradas
	// Formato: método, path, handler
	Routes() []Route
}

// Route define uma rota HTTP
type Route struct {
	Handler http.HandlerFunc // Handler da rota
	Method  string           // GET, POST, PUT, DELETE, etc
	Path    string           // /api/v1/resource
}

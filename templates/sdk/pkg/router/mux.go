// pkg/router/mux.go
package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Mux é fachada para roteador HTTP
// Permite trocar implementação sem quebrar plugins
type Mux struct {
	R *mux.Router // Exported para compatibilidade
}

// NewMux cria novo roteador
func NewMux() *Mux {
	return &Mux{
		R: mux.NewRouter(),
	}
}

// Handle registra rota
func (m *Mux) Handle(method, path string, handler http.HandlerFunc) {
	m.R.HandleFunc(path, handler).Methods(method)
}

// Use adiciona middleware global
func (m *Mux) Use(middleware func(http.Handler) http.Handler) {
	m.R.Use(mux.MiddlewareFunc(middleware))
}

// ServeHTTP implementa http.Handler
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.R.ServeHTTP(w, r)
}

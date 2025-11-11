package metrics

import (
	"net"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry encapsula o registro de métricas e exposição HTTP.
type Registry struct {
	server *http.Server
	once   sync.Once
}

// NewRegistry cria um novo Registry com namespace customizado.
func NewRegistry(namespace string) *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	return reg
}

// Serve expõe o endpoint /metrics utilizando o registry informado.
func (r *Registry) Serve(addr string, reg *prometheus.Registry) {
	r.once.Do(func() {
		r.server = &http.Server{
			Addr:    addr,
			Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		}

		go func() {
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return
			}
			_ = r.server.Serve(ln)
		}()
	})
}

// Shutdown encerra o servidor de métricas.
func (r *Registry) Shutdown() error {
	if r.server == nil {
		return nil
	}
	return r.server.Close()
}


// Package httpx provides a facade for HTTP routing using chi.
// This package encapsulates the chi router to prevent direct dependencies
// throughout the codebase.
package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Router defines the interface for HTTP routing operations.
// It encapsulates chi.Mux functionality.
type Router interface {
	// Use appends one or more middlewares onto the Router stack.
	Use(middlewares ...func(http.Handler) http.Handler)

	// Method adds a route for the given HTTP method, pattern, and handler.
	Method(method, pattern string, h http.Handler)

	// Get registers a GET route with the given pattern and handler.
	Get(pattern string, h http.HandlerFunc)

	// Post registers a POST route with the given pattern and handler.
	Post(pattern string, h http.HandlerFunc)

	// Put registers a PUT route with the given pattern and handler.
	Put(pattern string, h http.HandlerFunc)

	// Delete registers a DELETE route with the given pattern and handler.
	Delete(pattern string, h http.HandlerFunc)

	// Patch registers a PATCH route with the given pattern and handler.
	Patch(pattern string, h http.HandlerFunc)

	// Mount attaches another http.Handler along a routing path.
	Mount(pattern string, h http.Handler)

	// ServeHTTP makes the router implement http.Handler.
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	// Group creates a new inline-Router with the same middleware stack.
	Group(fn func(r Router)) Router

	// Route creates a new Mux with the same middleware stack and mounts it.
	Route(pattern string, fn func(r Router)) Router

	// With creates a new inline-Router with the same middleware stack.
	With(middlewares ...func(http.Handler) http.Handler) Router
}

// router is the internal implementation that wraps chi.Mux.
type router struct {
	*chi.Mux
}

// NewRouter creates and returns a new Router instance.
func NewRouter() Router {
	return &router{chi.NewRouter()}
}

// Group creates a new inline-Router with the same middleware stack.
func (r *router) Group(fn func(r Router)) Router {
	newMux := chi.NewRouter()
	// Copy middleware stack
	newMux.Use(r.Mux.Middlewares()...)
	fn(&router{newMux})
	return &router{newMux}
}

// Route creates a new Mux with the same middleware stack and mounts it.
func (r *router) Route(pattern string, fn func(r Router)) Router {
	subRouter := chi.NewRouter()
	subRouter.Use(r.Mux.Middlewares()...)
	fn(&router{subRouter})
	r.Mux.Mount(pattern, subRouter)
	return &router{subRouter}
}

// With creates a new inline-Router with the same middleware stack.
func (r *router) With(middlewares ...func(http.Handler) http.Handler) Router {
	newMux := chi.NewRouter()
	// Copy existing middlewares
	newMux.Use(r.Mux.Middlewares()...)
	// Add new middlewares
	newMux.Use(middlewares...)
	return &router{newMux}
}

// CORS returns a CORS middleware with the given options.
func CORS(opts cors.Options) func(http.Handler) http.Handler {
	return cors.New(opts).Handler
}

// DefaultCORS returns a CORS middleware with sensible defaults.
func DefaultCORS() func(http.Handler) http.Handler {
	return CORS(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// Common chi middlewares re-exported for convenience

// RequestID is a middleware that injects a request ID into the context.
func RequestID(next http.Handler) http.Handler {
	return chimw.RequestID(next)
}

// RealIP is a middleware that sets the RemoteAddr to the results of parsing
// the X-Real-IP or X-Forwarded-For headers.
func RealIP(next http.Handler) http.Handler {
	return chimw.RealIP(next)
}

// Recoverer is a middleware that recovers from panics, logs the panic,
// and returns a HTTP 500 status if possible.
func Recoverer(next http.Handler) http.Handler {
	return chimw.Recoverer(next)
}

// Logger is a middleware that logs the start and end of each request.
func Logger(next http.Handler) http.Handler {
	return chimw.Logger(next)
}

// Compress is a middleware that compresses response body.
func Compress(level int, types ...string) func(http.Handler) http.Handler {
	return chimw.Compress(level, types...)
}

// Timeout is a middleware that cancels the context after the given timeout.
// Note: This is a placeholder. Implement custom timeout logic if needed.
func Timeout(timeout int) func(http.Handler) http.Handler {
	// For now, we don't provide a default timeout middleware
	// Implement based on your specific needs
	return func(next http.Handler) http.Handler {
		return next
	}
}

// URLParam returns the url parameter from a http.Request object.
func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// URLParamFromCtx returns the url parameter from the context.
func URLParamFromCtx(ctx interface{}, key string) string {
	// Type assertion to ensure we're working with the right context type
	if chiCtx, ok := ctx.(chi.Context); ok {
		return chiCtx.URLParam(key)
	}
	return ""
}

// ResponseWriter is an interface that wraps http.ResponseWriter and adds
// methods to capture the status code and bytes written.
type ResponseWriter interface {
	http.ResponseWriter
	Status() int
	BytesWritten() int
}

// wrapResponseWriter is an implementation of ResponseWriter.
type wrapResponseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
	wroteHeader  bool
}

// NewWrapResponseWriter creates a new wrapped response writer.
func NewWrapResponseWriter(w http.ResponseWriter, protoMajor int) ResponseWriter {
	return &wrapResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// WriteHeader captures the status code and calls the underlying WriteHeader.
func (w *wrapResponseWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.status = status
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(status)
	}
}

// Write captures bytes written and calls the underlying Write.
func (w *wrapResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

// Status returns the HTTP status code.
func (w *wrapResponseWriter) Status() int {
	return w.status
}

// BytesWritten returns the number of bytes written.
func (w *wrapResponseWriter) BytesWritten() int {
	return w.bytesWritten
}

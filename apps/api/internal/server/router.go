// Package server builds the HTTP router and runs the server process.
package server

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"

	"github.com/smartreversal/smartreversal/apps/api/internal/observability"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/httpx"
)

// Options configures NewRouter.
type Options struct {
	Logger  *slog.Logger
	Version string
	// Service names the process in /v1/health. Default "api".
	Service string
	// Checkers are run by GET /v1/ready.
	Checkers []Checker
}

// NewRouter returns the API router. Routes live under /v1.
func NewRouter(o Options) http.Handler {
	r := chi.NewRouter()
	r.NotFound(httpx.NotFound)
	r.MethodNotAllowed(httpx.MethodNotAllowed)

	r.Use(observability.TraceMiddleware)
	r.Use(observability.RequestLogger(o.Logger))
	r.Use(recoverer(o.Logger))

	service := o.Service
	if service == "" {
		service = "api"
	}
	h := &health{service: service, version: o.Version, checkers: o.Checkers}
	r.Route("/v1", func(v1 chi.Router) {
		v1.Get("/health", h.live)
		v1.Get("/ready", h.ready)
	})
	return r
}

// recoverer turns a panic into a 500 in the standard envelope. The panic value
// and stack go to the log only, never to the client.
func recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				logger.ErrorContext(r.Context(), "panic_recovered",
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
					slog.String("trace_id", observability.TraceID(r.Context())),
				)
				httpx.WriteError(w, r, http.StatusInternalServerError, httpx.CodeInternal, "internal error")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

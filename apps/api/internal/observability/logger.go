package observability

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type correlationIDKey struct{}

// NewLogger initializes a structured JSON logger adhering to enterprise observability standards.
func NewLogger(env string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if env == "development" {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

// CorrelationMiddleware injects or propagates an X-Correlation-ID header and context value.
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := r.Header.Get("X-Correlation-ID")
		if corrID == "" {
			corrID = uuid.NewString()
		}
		w.Header().Set("X-Correlation-ID", corrID)
		ctx := context.WithValue(r.Context(), correlationIDKey{}, corrID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCorrelationID extracts the correlation ID from context.
func GetCorrelationID(ctx context.Context) string {
	if val, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return val
	}
	return ""
}

// RequestLoggerMiddleware logs structured request metadata with elapsed duration.
func RequestLoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			corrID := GetCorrelationID(r.Context())

			next.ServeHTTP(w, r)

			logger.Info("http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("correlation_id", corrID),
				slog.Duration("duration_ms", time.Since(start)),
			)
		})
	}
}

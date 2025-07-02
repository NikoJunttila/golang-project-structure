package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/nikojunttila/community/internal/logger"
	"github.com/rs/zerolog"
)

const slowRequestThreshold = 800 * time.Millisecond

// RequestLogger returns a middleware that:
// 1. Enriches each request with a contextual logger (e.g. request_id, method, path).
// 2. Logs the request after completion, with appropriate severity based on status code.
func RequestLogger(baseLogger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Extract request ID (set by chi's middleware.RequestID if enabled)
			requestID := middleware.GetReqID(r.Context())
			// realIP := middleware.GetR
			// Create a contextual logger with request metadata
			reqLogger := baseLogger.With().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote", r.RemoteAddr).
				Logger()

			// Inject the contextual logger into the request context
			ctx := logger.NewContext(r.Context(), &reqLogger)
			r = r.WithContext(ctx)

			// Serve the request
			next.ServeHTTP(ww, r)

			// Log request summary after handling
			log := logger.FromContext(ctx)
			duration := time.Since(start)

			entry := log.With().
				Str("request_id", requestID).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Dur("latency", duration).
				Logger()

			// Check if the request duration exceeds the threshold
			if duration > slowRequestThreshold {
				entry.Warn().Msg("Slow request detected")
			}

			switch {
			case ww.Status() >= 500:
				entry.Error().Msg("Server error")
			case ww.Status() >= 400:
				entry.Warn().Msg("Client error")
			default:
				entry.Info().Msg("Request handled successfully")
			}
		})
	}
}

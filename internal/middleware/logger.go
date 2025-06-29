package customMiddleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// RequestLogger is a middleware factory that returns a new middleware handler
// for logging HTTP requests using the provided zerolog.Logger.
func RequestLogger(l zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Prepare fields for logging.
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Defer the logging of the request completion.
			defer func() {
				// Log request details using the logger instance passed to the middleware.
				l.Info().
					Str("ip", r.RemoteAddr).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Int("status", ww.Status()).
					Dur("latency", time.Since(start)).
					Int("bytes", ww.BytesWritten()).
					Msg("Request handled")
			}()

			// Call the next handler in the chain.
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}


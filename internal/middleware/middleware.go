package middleware

import (
	"time"

	"github.com/didip/tollbooth/v7"         // Use v7
	"github.com/didip/tollbooth/v7/limiter" // Use v7
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)
//InitializeMiddleware contains middleware that all routes use
func InitializeMiddleware(r *chi.Mux, logger zerolog.Logger) {
	//rate limiting - 10 requests per second per IP
	//First param is requests per second and clears inactive IPS after hour
	lmt := tollbooth.NewLimiter(10.0, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	// Allow burst of up to 50 requests
	lmt.SetBurst(50)

	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(RequestLogger(logger))
	r.Use(tollbooth_chi.LimitHandler(lmt))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
}

// func middlewareTest(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Info().Msgf("middleware %v", r.Cookies())
// 		next.ServeHTTP(w, r)
// 	})
// }

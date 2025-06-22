package customMiddleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
  "github.com/rs/zerolog/log"
)

func InitializeMiddleware(r *chi.Mux) {
	//middleware is called in reverse order.
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(middlewareTest) //first called
}

func middlewareTest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msgf("middleware %v", r.Cookies())
		next.ServeHTTP(w, r)
	})
}

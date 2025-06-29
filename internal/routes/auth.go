package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/internal/handlers"
)

func registerAuthRoutes(r chi.Router) {
	r.Get("/foo", handlers.GetFooHandler)
	r.Get("/profile", handlers.GetProfileHandler)

	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		w.Write(fmt.Appendf(nil, "Welcome to admin dashboard, %v", claims["username"]))
	})
}

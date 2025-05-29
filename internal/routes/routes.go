package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/handlers"
)

func InitializeRoutes(r *chi.Mux) {
	// Group for authenticated (non-public) routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome anonymous"))
	})

	r.Route("/auth", func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtauth.Verifier(auth.GetTokenAuth()))

		// Handle valid / invalid tokens. In this example, we use
		// the provided authenticator middleware, but you can write your
		// own very easily, look at the Authenticator method in jwtauth.go
		// and tweak it, its not scary.
		r.Use(jwtauth.Authenticator(auth.GetTokenAuth()))
		registerAuthRoutes(r)
	})

	// Group for public routes
	r.Route("/public", func(r chi.Router) {
		registerPublicRoutes(r)
	})
}

func registerAuthRoutes(r chi.Router) {
	r.Get("/foo", handlers.GetFooHandler)
	r.Get("/profile", handlers.GetProfileHandler)

	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		w.Write(fmt.Appendf(nil, "Welcome to admin dashboard, %v", claims["username"]))
	})
}

func registerPublicRoutes(r chi.Router) {
	r.Get("/{provider}/begin", handlers.GetBeginAuth)
	r.Get("/{provider}/callback", handlers.GetAuthCallBack)

	r.Post("/email_create", handlers.PostCreateUserHandlerEmail)
	r.Post("/email_login", handlers.PostLoginHandler)

}

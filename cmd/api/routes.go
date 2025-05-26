package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/auth"
	"github.com/nikojunttila/community/handlers"
)

func initializeRoutes(r *chi.Mux) {
	// Protected routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtauth.Verifier(auth.GetTokenAuth()))

		// Handle valid / invalid tokens. In this example, we use
		// the provided authenticator middleware, but you can write your
		// own very easily, look at the Authenticator method in jwtauth.go
		// and tweak it, its not scary.
		r.Use(jwtauth.Authenticator(auth.GetTokenAuth()))

		r.Get("/foo", handlers.GetFooHandler)
		r.Get("/profile", handlers.GetProfileHandler)
		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write(fmt.Appendf(nil, "protected area. hi %v", claims["username"]))
		})
	})

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/google", handlers.GetGoogleLogin)
		r.Get("/callback", handlers.GetGoogleCallBack)

		r.Post("/create", handlers.PostCreateUserHandler)
		r.Post("/login", handlers.GetLoginHandler)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome anonymous"))
		})
	})

}

// Package routes is routing logic
package routes

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/handlers"
	"github.com/nikojunttila/community/internal/middleware"
)

// InitializeRoutes is used to setup all routes belonging to this web service
func InitializeRoutes(r *chi.Mux) {
	r.Get("/health", handlers.HealthCheck)
	r.Get("/health/db", handlers.HealthCheckDB)
	r.Get("/healthz", handlers.HealthCheck)

	r.Get("/upload", handlers.GetUploadPageHandler)
	r.Post("/upload", handlers.PostFileUploadHandler)

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/files", filesDir) //files/index.html servers file

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	r.Route("/user", func(r chi.Router) {
		r.Use(middleware.RequireRoles(auth.User, auth.Admin))
		r.Get("/foo", handlers.GetFooHandler)
	})

	r.Route("/admin", func(r chi.Router) {
		//log access to database for later identification on important endpoints
		r.Use(jwtauth.Verifier(auth.GetTokenAuth()))
		r.Use(jwtauth.Authenticator(auth.GetTokenAuth()))
		r.Use(middleware.RequireRoles(auth.Admin))

		r.Use(middleware.AdminAuditMiddleware())
		registerAdminRoutes(r)
	})

	r.Route("/two", func(r chi.Router) {
		twoFactorRoutes(r)
	})

	r.Route("/twoauth", func(r chi.Router) {
		//without these we cant find jwt from context
		r.Use(jwtauth.Verifier(auth.GetTokenAuth()))
		r.Use(jwtauth.Authenticator(auth.GetTokenAuth()))
		twoFactorRoutesAuth(r)
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

func registerPublicRoutes(r chi.Router) {
	r.Get("/{provider}/begin", handlers.GetBeginAuth)
	r.Get("/{provider}/callback", handlers.GetAuthCallBack)

	r.Get("/email_create", handlers.GetCreatePage)
	r.Post("/email_create", handlers.PostCreateUserHandlerEmail)
	r.Post("/email_login", handlers.PostLoginHandler)
}

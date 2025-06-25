package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/handlers"
)

func InitializeRoutes(r *chi.Mux) {
	// Group for authenticated (non-public) routes
	workDir, _ := os.Getwd()

	filesDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/files", filesDir) //files/index.html servers file

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	r.Route("/two", func(r chi.Router) {
		registerTwoFactorRoutes(r)
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
func registerTwoFactorRoutes(r chi.Router) {
	r.Get("/", handlers.GetHomeHandler)
	r.Get("/login", handlers.LoginHandler)
	r.Post("/login", handlers.LoginHandler)
	r.Get("/dashboard", handlers.GetDashboardHandler)
	r.Get("/generate-otp", handlers.GetGenerateOTPHandler)
	r.Get("/validate-otp", handlers.ValidateOTPHandler)
	r.Post("/validate-otp", handlers.ValidateOTPHandler)
	r.Get("/debug", handlers.DebugHandler)

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

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

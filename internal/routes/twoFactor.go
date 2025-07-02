package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/nikojunttila/community/internal/handlers"
)

func twoFactorRoutes(r chi.Router) {
	r.Get("/", handlers.GetHomeHandler)
	r.Get("/login", handlers.LoginHandler)
	r.Post("/login", handlers.LoginHandler)
}

func twoFactorRoutesAuth(r chi.Router) {
	r.Get("/dashboard", handlers.GetDashboardHandler)
	r.Get("/generate-otp", handlers.GetGenerateOTPHandler)
	r.Get("/validate-otp", handlers.ValidateOTPHandler)
	r.Post("/validate-otp", handlers.ValidateOTPHandler)
}

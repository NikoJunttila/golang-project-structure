package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/nikojunttila/community/internal/handlers"
)

func registerAdminRoutes(r chi.Router) {
	r.Get("/profile", handlers.GetProfileAdmin)
}

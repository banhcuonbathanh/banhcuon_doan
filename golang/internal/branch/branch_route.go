package branch

import (
	"english-ai-full/pkg/middleware/auth"

	"github.com/go-chi/chi"
)

// RegisterRoutes registers all branch-related routes
func RegisterRoutes(r *chi.Mux, handler *Handler) {
	r.Route("/branches", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Post("/", handler.CreateBranch)
			r.Get("/{id}", handler.GetBranchByID)
		})
	})
}

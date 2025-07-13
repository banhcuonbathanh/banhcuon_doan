package account

import (
	"net/http"

	"english-ai-full/pkg/middleware/auth"
	"github.com/go-chi/chi"
)

func RegisterRoutes(r *chi.Mux, handler Handler) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is running"))
	})
	r.Route("/accounts", func(r chi.Router) {
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/logout", handler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Post("/", handler.CreateAccount)
			r.Post("/{id}", handler.UpdateUserByID)
			r.Get("/{id}", handler.FindAccountByID)
			r.Delete("/{id}", handler.DeleteUser)
			r.Get("/email/{email}", handler.FindByEmail)
		})
	})
}

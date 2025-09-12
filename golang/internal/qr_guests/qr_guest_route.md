package qr_guests

import (
	"english-ai-full/pkg/middleware/auth"
	"net/http"

	"github.com/go-chi/chi"
)

func RegisterGuestRoutes(r *chi.Mux, handler *GuestHandlerController) *chi.Mux {
	r.Get("/qr/guest/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("guest test is running"))
	})

	//----------------------------

	//----------------------------
	r.Route("/qr/guest", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Post("/login", handler.GuestLogin)
		r.Post("/refresh-token", handler.RefreshToken)

		// Protected routes (authentication required)
		r.Group(func(r chi.Router) {
			// r.Use(middleware.GetAuthMiddlewareFunc(tokenMaker))
			r.Post("/logout", handler.GuestLogout)
			// r.Post("/orders", handler.CreateOrders)
			// r.Get("/orders/{guestId}", handler.GetOrders)
		})
	})

	return r
}

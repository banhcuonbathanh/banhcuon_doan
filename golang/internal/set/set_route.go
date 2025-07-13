package set_qr

import (
	"net/http"

	"github.com/go-chi/chi"
)

func RegisterSetRoutes(r *chi.Mux, handler *SetHandlerController) *chi.Mux {
	r.Get("/sets-test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server set is running"))
	})

	r.Route("/sets", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// r.Use(middleware.GetAuthMiddlewareFunc(handler.TokenMaker))

			r.Get("/", handler.GetSetProtoListDetail)
			r.Post("/", handler.CreateSetProto)
			r.Get("/", handler.GetSetProtoListDetail)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetSetProtoDetail)
				r.Put("/", handler.UpdateSetProto)
				r.Delete("/", handler.DeleteSetProto)
			})
		})
	})

	return r
}
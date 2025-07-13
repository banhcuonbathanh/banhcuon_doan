package dish_grpc

import (
	"net/http"

	"github.com/go-chi/chi"
)

func RegisterDishRoutes(r *chi.Mux, handler *DishHandlerController) *chi.Mux {
	// tokenMaker := handler.TokenMaker
	r.Get("/dishes-test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server dish is running"))
	})

	r.Route("/dishes", func(r chi.Router) {


		r.Group(func(r chi.Router) {
			// r.Use(middleware.GetAuthMiddlewareFunc(tokenMaker))

			r.Get("/", handler.GetDishList)
			r.Post("/", handler.CreateDish)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetDishDetail)
				r.Put("/", handler.UpdateDish)
				r.Delete("/", handler.DeleteDish)
			})
		})
	})

	return r
}
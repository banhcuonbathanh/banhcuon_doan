package delivery_grpc

import (
	"net/http"

	"github.com/go-chi/chi"
)

func RegisterDeliveryRoutes(r *chi.Mux, handler *DeliveryHandlerController) *chi.Mux {
	// Health check endpoint
	r.Get("/delivery-test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server delivery is running"))
	})

	r.Route("/delivery", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// If you need authentication middleware, uncomment and adjust the following line:
			// r.Use(middleware.GetAuthMiddlewareFunc(handler.TokenMaker))

			// Basic CRUD operations
			r.Post("/", handler.CreateDelivery)
			r.Get("/", handler.GetDeliveriesListDetail)

			// ID-based operations
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetDeliveryDetail)
				r.Put("/", handler.UpdateDelivery)
				r.Delete("/", handler.DeleteDelivery)
			})

			// Client name based operation
			r.Get("/client/{name}", handler.GetDeliveryByClientName)
		})
	})

	return r
}
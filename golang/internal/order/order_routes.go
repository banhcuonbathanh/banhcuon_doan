package order_grpc

import (
	"net/http"

	"github.com/go-chi/chi"
	// middleware "english-ai-full/ecomm-api"
)

func RegisterOrderRoutes(r *chi.Mux, handler *OrderHandlerController) *chi.Mux {
	r.Get("/orders-test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server order is running"))
	})

	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// If you need authentication middleware, uncomment and adjust the following line:
			// r.Use(middleware.GetAuthMiddlewareFunc(handler.TokenMaker))

			r.Post("/", handler.CreateOrder)
			r.Get("/", handler.GetOrderProtoListDetail)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetOrderDetail)
				r.Put("/", handler.UpdateOrder)
			})

			r.Post("/pay/{guest_id}", handler.PayOrders)
		})
	})

	return r
}
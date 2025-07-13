// routes/websocket_routes.go
package ws

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	order "english-ai-full/internal/order"
)

func RegisterWebSocketRoutes(r *chi.Mux, wsHandler *WebSocketHandler, orderHandler *order.OrderHandlerController) *chi.Mux {
	r.Route("/ws", func(r chi.Router) {
		// User WebSocket connections
		r.Get("/user/{user_id}", func(w http.ResponseWriter, r *http.Request) {
			userIDStr := chi.URLParam(r, "user_id")
			userName := r.URL.Query().Get("username")

			// Convert userIDStr to int64
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				// Handle the error, e.g., return an HTTP error response
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}

			wsHandler.HandleUserConnection(w, r, userID, userName)
		})

		// Guest WebSocket connections
		r.Get("/guest/{guest_id}", func(w http.ResponseWriter, r *http.Request) {
			guestIDStr := chi.URLParam(r, "guest_id")
			guestName := r.URL.Query().Get("guestname")

			guestID, err := strconv.ParseInt(guestIDStr, 10, 64)
			if err != nil {
				// Handle the error, e.g., return an HTTP error response
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}

			wsHandler.HandleGuestConnection(w, r, guestID, guestName)
		})
	})

	// Order-related WebSocket notifications
	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				orderHandler.CreateOrder(w, r)
				// Broadcast order creation to relevant clients
				wsHandler.BroadcastOrderUpdate("order_created", r.Context())
			})

			r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
				orderHandler.UpdateOrder(w, r)
				// Broadcast order update to relevant clients
				wsHandler.BroadcastOrderUpdate("order_updated", r.Context())
			})
		})
	})

	// Delivery-related WebSocket notifications
	r.Route("/deliveries", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				// Handle delivery creation
				wsHandler.BroadcastDeliveryUpdate("delivery_created", r.Context())
			})

			r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
				// Handle delivery update
				wsHandler.BroadcastDeliveryUpdate("delivery_updated", r.Context())
			})
		})
	})

	return r
}

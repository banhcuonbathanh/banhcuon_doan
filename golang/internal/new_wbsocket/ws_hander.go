package ws

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
    upgrader         websocket.Upgrader
    websocketService WebSocketService
}

func (h *WebSocketHandler) HandleUserConnection(w http.ResponseWriter, r *http.Request, userID int64, userName string) {

    
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
        return
    }

    client := NewClient(conn, h.websocketService, userID, userName, false)
    h.websocketService.RegisterClient(client)

    // Start client routines
    go client.readPump()
    go client.writePump()
}

func (h *WebSocketHandler) HandleGuestConnection(w http.ResponseWriter, r *http.Request, guestID int64, guestName string) {
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
        return
    }

    client := NewClient(conn, h.websocketService, guestID, guestName, true)
    h.websocketService.RegisterClient(client)

    // Start client routines
    go client.readPump()
    go client.writePump()
}

func (h *WebSocketHandler) BroadcastOrderUpdate(messageType string, ctx context.Context) {
    message := &Message{
        Type:      messageType,
        Timestamp: time.Now(),
        // Add relevant order data from context
    }
    h.websocketService.BroadcastMessage(message)
}

func (h *WebSocketHandler) BroadcastDeliveryUpdate(messageType string, ctx context.Context) {
    message := &Message{
        Type:      messageType,
        Timestamp: time.Now(),
        // Add relevant delivery data from context
    }
    h.websocketService.BroadcastMessage(message)
}
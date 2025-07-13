package ws2

import (
	"encoding/json"

	"log"
	"net/http"
	"time"

	"english-ai-full/token"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

type WebSocketRouter struct {
    hub           *Hub
    deliveryQueue chan DeliveryUpdate

    // new 
 
    TokenMaker *token.JWTMaker
}

type DeliveryUpdate struct {
    Action     string      `json:"action"`
    DeliveryID string      `json:"deliveryId"`
    Payload    interface{} `json:"payload"`
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true 
    },
}

// new
func NewWebSocketRouter(h *Hub, tokenMaker *token.JWTMaker) *WebSocketRouter {
	router := &WebSocketRouter{
		hub:           h,
		deliveryQueue: make(chan DeliveryUpdate, 100),
        TokenMaker: tokenMaker,  // Initialize the TokenMaker
	 
	}
	
	go router.processDeliveryUpdates()
	
	return router
}
// func NewWebSocketRouter(h *Hub) *WebSocketRouter {
//     router := &WebSocketRouter{
//         hub:           h,
//         deliveryQueue: make(chan DeliveryUpdate, 100),
//     }
    
  
//     go router.processDeliveryUpdates()
    
//     return router
// }

func (wr *WebSocketRouter) RegisterRoutes(r chi.Router) {
    r.Route("/ws", func(r chi.Router) {
        // First register the API routes
        r.Route("/api", func(r chi.Router) {
      
            r.Post("/ws-auth", wr.handleTokenGeneration)
        })

        // Then register WebSocket endpoint routes
        r.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
            log.Println("golang/quanqr/ws2/ws2_route.go user RegisterRoutes")
            wr.handleWebSocket(w, r, RoleUser)
        })

        r.Get("/guest/{id}", func(w http.ResponseWriter, r *http.Request) {
            log.Println("golang/quanqr/ws2/ws2_route.go guest RegisterRoutes")
            wr.handleWebSocket(w, r, RoleGuest)
        })

        r.Get("/kitchen/{id}", func(w http.ResponseWriter, r *http.Request) {
            log.Println("golang/quanqr/ws2/ws2_route.go kitchen RegisterRoutes")
            wr.handleWebSocket(w, r, RoleKitchen)
        })

        r.Get("/employee/{id}", func(w http.ResponseWriter, r *http.Request) {
            log.Println("golang/quanqr/ws2/ws2_route.go employee RegisterRoutes")
            wr.handleWebSocket(w, r, RoleEmployee)
        })

        r.Get("/admin/{id}", func(w http.ResponseWriter, r *http.Request) {
            wr.handleWebSocket(w, r, RoleAdmin)
        })
    })
}

func (wr *WebSocketRouter) handleWebSocket(w http.ResponseWriter, r *http.Request, role Role) {
    log.Println("golang/quanqr/ws2/ws2_route.go handleWebSocket")

    // Extract all required parameters
    userToken := r.URL.Query().Get("token")
    tableToken := r.URL.Query().Get("tableToken")
    email := r.URL.Query().Get("email")

    // Validate required parameters
    if email == "" {
        log.Printf("Error: email parameter is required")
        http.Error(w, "Email parameter is required", http.StatusBadRequest)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Error upgrading connection: %v", err)
        return
    }

    client := &Client{
        Hub:    wr.hub,
        Conn:   conn,
        Send:   make(chan []byte, 256),
        Role:   role,
        ID:     chi.URLParam(r, "id"),
        RoomID: r.URL.Query().Get("roomId"),
        UserData: map[string]interface{}{
            "token":      userToken,
            "tableToken": tableToken,
            "email":      email,
        },
    }

    client.Hub.Register <- client
  
    go client.ReadPump()
    go client.WritePump()
}


func (wr *WebSocketRouter) BroadcastDeliveryUpdate(action string, deliveryID string, payload interface{}) {
    wr.deliveryQueue <- DeliveryUpdate{
        Action:     action,
        DeliveryID: deliveryID,
        Payload:    payload,
    }
}

// processDeliveryUpdates handles the delivery update queue
func (wr *WebSocketRouter) processDeliveryUpdates() {
    for update := range wr.deliveryQueue {
        message := Message{
            Type:    "delivery",
            Action:  update.Action,
            Payload: update.Payload,
            Role:    RoleEmployee, // Default to employee, can be modified based on needs
        }

        // Marshal the message
        data, err := json.Marshal(message)
        if err != nil {
            log.Printf("Error marshaling delivery update: %v", err)
            continue
        }


        wr.hub.Broadcast <- data
    }
}


func (wr *WebSocketRouter) GetConnectedClientsCount() map[Role]int {
    counts := make(map[Role]int)
    wr.hub.mu.Lock()
    defer wr.hub.mu.Unlock()

    for client := range wr.hub.Clients {
        counts[client.Role]++
    }
    return counts
}

// new ---------------
func (wr *WebSocketRouter) handleTokenGeneration(w http.ResponseWriter, r *http.Request) {
    log.Println("golang/quanqr/ws2/ws2_route.go user RegisterRoutes ws token create")
	var req TokenRequestWS
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	wsRole := Role(req.Role)
	if !isValidRole(wsRole) {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Create token with 24 hour expiration
	token, claims, err := wr.TokenMaker.CreateToken(
		req.UserID,
		req.Email,
		string(wsRole),
		24*time.Hour,
	)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	response := TokenResponseWS{
		Token:     token,
		ExpiresAt: claims.ExpiresAt.Time,
		Role:      claims.Role,
		UserID:    claims.ID,
		Email:     claims.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


// func (wr *WebSocketRouter) handleWebSocket(w http.ResponseWriter, r *http.Request, role Role) {
// 	// Validate token
// 	token := r.URL.Query().Get("token")
// 	if token != "" {
// 		claims, err := wr.tokenMaker.VerifyToken(token)
// 		if err != nil || Role(claims.Role) != role {
// 			http.Error(w, "Invalid token or role mismatch", http.StatusUnauthorized)
// 			return
// 		}
// 	}

// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Printf("Error upgrading connection: %v", err)
// 		return
// 	}

// 	tableToken := r.URL.Query().Get("tableToken")
// 	client := &Client{
// 		Hub:    wr.hub,
// 		Conn:   conn,
// 		Send:   make(chan []byte, 256),
// 		Role:   role,
// 		ID:     chi.URLParam(r, "id"),
// 		RoomID: r.URL.Query().Get("roomId"),
// 		UserData: map[string]interface{}{
// 			"token":      token,
// 			"tableToken": tableToken,
// 		},
// 	}

// 	client.Hub.Register <- client
// 	go client.ReadPump()
// 	go client.WritePump()
// }

// func (wr *WebSocketRouter) BroadcastDeliveryUpdate(action string, deliveryID string, payload interface{}) {
// 	wr.deliveryQueue <- DeliveryUpdate{
// 		Action:     action,
// 		DeliveryID: deliveryID,
// 		Payload:    payload,
// 	}
// }

// func (wr *WebSocketRouter) processDeliveryUpdates() {
// 	for update := range wr.deliveryQueue {
// 		message := Message{
// 			Type:    TypeDelivery,
// 			Action:  update.Action,
// 			Payload: update.Payload,
// 			Role:    RoleEmployee,
// 		}

// 		data, err := json.Marshal(message)
// 		if err != nil {
// 			log.Printf("Error marshaling delivery update: %v", err)
// 			continue
// 		}

// 		wr.hub.Broadcast <- data
// 	}
// }

// func (wr *WebSocketRouter) GetConnectedClientsCount() map[Role]int {
// 	counts := make(map[Role]int)
// 	wr.hub.mu.Lock()
// 	defer wr.hub.mu.Unlock()

// 	for client := range wr.hub.Clients {
// 		counts[client.Role]++
// 	}
// 	return counts
// }

// Helper function to validate roles
func isValidRole(role Role) bool {
	validRoles := map[Role]bool{
		RoleAdmin:    true,
		RoleGuest:    true,
		RoleUser:     true,
		RoleKitchen:  true,
		RoleEmployee: true,
	}
	return validRoles[role]
}

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }
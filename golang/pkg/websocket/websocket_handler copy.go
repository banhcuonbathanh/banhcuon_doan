package websockethandler

// import (
// 	"log"
// 	"net/http"
// 	"time"

// 	"github.com/gorilla/websocket"
// )

// type Handler struct {
// 	upgrader websocket.Upgrader
// }

// type Message struct {
// 	Type string `json:"type"`
// 	Content string `json:"content"`
// }

// func NewHandler() *Handler {
// 	return &Handler{
// 		upgrader: websocket.Upgrader{
// 			ReadBufferSize:  1024,
// 			WriteBufferSize: 1024,
// 			CheckOrigin: func(r *http.Request) bool {
// 				return true
// 			},
// 		},
// 	}
// }

// func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
// 	log.Println("WebSocket handler called")

// 	conn, err := h.upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Printf("Failed to upgrade connection: %v", err)
// 		return
// 	}
// 	defer conn.Close()

// 	log.Println("WebSocket connection established")

// 	// Set read deadline
// 	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

// 	// Set pong handler
// 	conn.SetPongHandler(func(string) error {
// 		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
// 		return nil
// 	})

// 	// Start a goroutine for ping and periodic messages
// 	go func() {
// 		pingTicker := time.NewTicker(54 * time.Second)
// 		messageTicker := time.NewTicker(10 * time.Second)
// 		defer pingTicker.Stop()
// 		defer messageTicker.Stop()
// 		for {
// 			select {
// 			case <-pingTicker.C:
// 				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
// 					log.Printf("ping error: %v", err)
// 					return
// 				}
// 			case <-messageTicker.C:
// 				message := Message{
// 					Type: "server_message",
// 					Content: "Periodic message from server",
// 				}
// 				if err := conn.WriteJSON(message); err != nil {
// 					log.Printf("error sending periodic message: %v", err)
// 					return
// 				}
// 			}
// 		}
// 	}()

// 	for {
// 		var receivedMessage Message
// 		err := conn.ReadJSON(&receivedMessage)
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				log.Printf("error reading message: %v", err)
// 			}
// 			break
// 		}
// 		log.Printf("Received message: %+v", receivedMessage)

// 		// Send a custom response
// 		response := Message{
// 			Type: "server_response",
// 			Content: "Server received: " + receivedMessage.Content,
// 		}
// 		err = conn.WriteJSON(response)
// 		if err != nil {
// 			log.Printf("error writing message: %v", err)
// 			break
// 		}
// 	}

// 	log.Println("WebSocket connection closed")
// }
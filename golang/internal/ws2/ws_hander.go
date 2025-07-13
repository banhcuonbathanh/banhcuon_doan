package ws2

import (

	"log"
)



type WebSocketHandler struct {
    hub *Hub

    // new 


	
}

// func NewWebSocketHandler(   CombinedMessageHandler *CombinedMessageHandler ) *WebSocketHandler {
// 	log.Println("golang/quanqr/ws2/ws_hander.go")
//     return &WebSocketHandler{
//         hub: NewHub(CombinedMessageHandler),
//     }
// }
// new

func NewWebSocketHandler(combinedMessageHandler *CombinedMessageHandler) *WebSocketHandler {
    log.Println("golang/quanqr/ws2/ws_hander.go")
	hub := NewHub(combinedMessageHandler)
	// router := NewWebSocketRouter(hub, " secreat key for ws 12412kj4h213k")
	
	return &WebSocketHandler{
		hub:        hub,
		// router:     router,

	}
}
package websocket_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	websocket_model "english-ai-full/pkg/websocket/websocker_model"
	"english-ai-full/pkg/websocket/websocket_repository"

	"github.com/google/uuid"

	order_grpc "english-ai-full/internal/order"
	"english-ai-full/internal/proto_qr/order"
)

// First, let's update the OrderContent struct to match the expected format
type OrderContent struct {
	Order struct {
		GuestID        int64                  `json:"guest_id"`
		UserID         int64                  `json:"user_id"`
		IsGuest        bool                   `json:"is_guest"`
		TableNumber    int64                  `json:"table_number"`
		OrderHandlerID int64                  `json:"order_handler_id"`
		Status         string                 `json:"status"`
		CreatedAt      string                 `json:"created_at,omitempty"`
		UpdatedAt      string                 `json:"updated_at,omitempty"`
		TotalPrice     int32                  `json:"total_price"`
		OrderName      string                 `json:"order_name"`
		DishItems      []order_grpc.OrderDish `json:"dish_items"`
		SetItems       []order_grpc.OrderSet  `json:"set_items"`
		Topping        string                 `json:"topping"`
		TrackingOrder  string                 `json:"tracking_order"`
		TakeAway       bool                   `json:"take_away"`
		ChiliNumber    int64                  `json:"chili_number"`
		TableToken     string                 `json:"Table_token"`
	} `json:"order"`
}

// Update the CreateOrderRequestType to match the incoming data

type ClientType string

const (
	UserClient  ClientType = "user"
	GuestClient ClientType = "guest"
)

type ClientIdentifier struct {
	ID       string
	Type     ClientType
	UserName string
}

type WebSocketService interface {
	RegisterClient(client *Client)
	UnregisterClient(client *Client)
	BroadcastMessage(message *websocket_model.Message)
	SendMessageToUser(fromUser, toUser string, messageType string, content interface{}, tableID, orderID string) error
	SendMessageToGuest(fromUser string, guestID string, messageType string, content interface{}, tableID, orderID string) error
	Run()
}

type webSocketService struct {
	clients      map[ClientIdentifier]map[*Client]bool
	broadcast    chan *websocket_model.Message
	register     chan *Client
	unregister   chan *Client
	mutex        sync.RWMutex
	repo         websocket_repository.MessageRepository
	orderHandler *order_grpc.OrderHandlerController
}

func NewWebSocketService(repo websocket_repository.MessageRepository, orderHandler *order_grpc.OrderHandlerController) WebSocketService {
	return &webSocketService{
		clients:      make(map[ClientIdentifier]map[*Client]bool),
		broadcast:    make(chan *websocket_model.Message),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		repo:         repo,
		orderHandler: orderHandler,
	}
}

func (s *webSocketService) UnregisterClient(client *Client) {
	s.unregister <- client
}

func (s *webSocketService) BroadcastMessage(message *websocket_model.Message) {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go BroadcastMessage")
	messageBytes, _ := json.Marshal(message)
	log.Printf("Broadcasting message: %s", string(messageBytes))
	s.broadcast <- message
}

func (s *webSocketService) SendMessageToUser(fromUser, toUser string, messageType string, content interface{}, tableID, orderID string) error {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go SendMessageToUser")
	message := s.createMessage(fromUser, toUser, messageType, content, tableID, orderID)

	return s.SendToClient(ClientIdentifier{ID: toUser, Type: UserClient}, message)
}

func (s *webSocketService) createMessage(fromUser, toUser string, messageType string, content interface{}, tableID, orderID string) *websocket_model.Message {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go createMessage")
	message := &websocket_model.Message{
		ID:        uuid.New().String(),
		Type:      messageType,
		Content:   content,
		Sender:    fromUser,
		FromUser:  fromUser,
		ToUser:    toUser,
		Timestamp: time.Now(),
		TableID:   tableID,
		OrderID:   orderID,
	}
	return message
}

func (s *webSocketService) Run() {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go Run")
	for {
		select {
		case client := <-s.register:
			log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go register")
			s.mutex.Lock()
			identifier := s.getClientIdentifier(client)
			if _, exists := s.clients[identifier]; !exists {
				s.clients[identifier] = make(map[*Client]bool)
			}
			s.clients[identifier][client] = true
			log.Printf("Client registered - Type: %v, ID: %s, Name: %s, Total clients for this identifier: %d",
				identifier.Type, identifier.ID, identifier.UserName, len(s.clients[identifier]))
			log.Printf("Total unique clients: %d", len(s.clients))
			s.mutex.Unlock()

		case client := <-s.unregister:
			log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go unregister")
			s.mutex.Lock()
			identifier := s.getClientIdentifier(client)
			if clients, exists := s.clients[identifier]; exists {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(s.clients, identifier)
					}
					log.Printf("Client unregistered - Type: %v, ID: %s, Name: %s, Remaining clients for this identifier: %d",
						identifier.Type, identifier.ID, identifier.UserName, len(clients))
				}
			}
			s.mutex.Unlock()

		case message := <-s.broadcast:
			log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go broadcast")
			s.mutex.RLock()
			for identifier, clients := range s.clients {
				for client := range clients {
					select {
					case client.send <- message:
						log.Printf("Broadcast message sent to %v %s (%s)",
							identifier.Type, identifier.ID, client.userName)
					default:
						log.Printf("Failed to broadcast to %v %s (%s)",
							identifier.Type, identifier.ID, client.userName)
					}
				}
			}
			s.mutex.RUnlock()

			if err := s.repo.SaveMessage(message); err != nil {
				log.Printf("Error saving message: %v", err)
			}
		}
	}
}

func (s *webSocketService) SendToClient(identifier ClientIdentifier, message *websocket_model.Message) error {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go SendToClient")
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var lastError error
	messagesSent := 0

	// First try exact match
	if clients, exists := s.clients[identifier]; exists {
		for client := range clients {
			select {
			case client.send <- message:
				messagesSent++
				log.Printf("Message sent to %v %s (%s)", identifier.Type, identifier.ID, client.userName)
			default:
				lastError = fmt.Errorf("failed to send message to %v %s (%s)",
					identifier.Type, identifier.ID, client.userName)
			}
		}
	} else {
		// If no exact match, try finding by ID and Type only
		found := false
		for clientId, clients := range s.clients {
			if clientId.ID == identifier.ID && clientId.Type == identifier.Type {
				found = true
				for client := range clients {
					select {
					case client.send <- message:
						messagesSent++
						log.Printf("Message sent to %v %s (%s)", clientId.Type, clientId.ID, client.userName)
					default:
						lastError = fmt.Errorf("failed to send message to %v %s (%s)",
							clientId.Type, clientId.ID, client.userName)
					}
				}
			}
		}
		if !found {
			return fmt.Errorf("no connected clients found for %v type: %v", identifier.Type, identifier.ID)
		}
	}

	if messagesSent > 0 {
		return nil
	}
	return lastError
}

func (s *webSocketService) SendMessageToGuest(fromUser string, guestID string, messageType string, content interface{}, tableID, orderID string) error {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go SendMessageToGuest")
	message := s.createMessage(fromUser, guestID, messageType, content, tableID, orderID)

	// Log the current state before sending
	s.mutex.RLock()
	log.Printf("Debug - Before sending - All registered clients:")
	for id, clients := range s.clients {
		log.Printf("ClientIdentifier: %+v, Number of connections: %d", id, len(clients))
	}
	s.mutex.RUnlock()

	identifier := ClientIdentifier{
		ID:   guestID,
		Type: GuestClient,
	}

	log.Printf("Attempting to send message from user %s to guest %s", fromUser, guestID)
	err := s.SendToClient(identifier, message)
	if err != nil {
		log.Printf("Error in SendMessageToGuest - Failed to send message: %v", err)
		return err
	}
	return nil
}

func (s *webSocketService) RegisterClient(client *Client) {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go RegisterClient")
	identifier := s.getClientIdentifier(client)

	// Log registration attempt
	log.Printf("Attempting to register client - UserID: %s, UserName: %s, IsGuest: %v",
		client.userID, client.userName, client.isGuest)
	log.Printf("Client identifier details - ID: %s, Type: %v, UserName: %s",
		identifier.ID, identifier.Type, identifier.UserName)

	s.register <- client

	// Wait a short time to ensure registration is complete
	time.Sleep(100 * time.Millisecond)

	// Verify registration
	s.mutex.RLock()
	if clients, exists := s.clients[identifier]; exists {
		log.Printf("Client successfully registered with %d active connections", len(clients))
	} else {
		log.Printf("Warning: Client registration may not be complete")
	}
	s.mutex.RUnlock()
}

func (s *webSocketService) getClientIdentifier(client *Client) ClientIdentifier {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go getClientIdentifier")
	clientType := UserClient
	if client.isGuest {
		clientType = GuestClient
	}

	identifier := ClientIdentifier{
		ID:       client.userID,
		Type:     clientType,
		UserName: client.userName,
	}

	log.Printf("Debug - Created identifier: %+v for client: %+v", identifier, client)
	return identifier
}

func ToPBDishOrderItems(items []websocket_model.DishOrderItem) []*order.DishOrderItem {
	pbItems := make([]*order.DishOrderItem, len(items))
	for i, item := range items {
		pbItems[i] = &order.DishOrderItem{
			DishId:   item.DishID,
			Quantity: item.Quantity,
		}
	}
	return pbItems
}

func ToPBSetOrderItems(items []websocket_model.SetOrderItem) []*order.SetOrderItem {
	pbItems := make([]*order.SetOrderItem, len(items))
	for i, item := range items {
		pbItems[i] = &order.SetOrderItem{
			SetId:    item.SetID,
			Quantity: item.Quantity,
		}
	}
	return pbItems
}

// MockResponseWriter implements http.ResponseWriter for testing
type MockResponseWriter struct {
	Headers http.Header
	Body    bytes.Buffer
	Status  int
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		Headers: make(http.Header),
	}
}

func (m *MockResponseWriter) Header() http.Header {
	return m.Headers
}

func (m *MockResponseWriter) Write(body []byte) (int, error) {
	return m.Body.Write(body)
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.Status = statusCode
}

// Update the type

// func ToPBCreateOrderRequest(req CreateOrderRequestType) *order.CreateOrderRequest {
//     var guestID, userID int64
//     if req.GuestID != nil {
//         guestID = int64(*req.GuestID)
//     }
//     if req.UserID != nil {
//         userID = int64(*req.UserID)
//     }

//     return &order.CreateOrderRequest{
//         GuestId:        guestID,
//         UserId:         userID,
//         IsGuest:        req.IsGuest,
//         TableNumber:    int64(req.TableNumber),
//         OrderHandlerId: int64(req.OrderHandlerID),
//         Status:         req.Status,
//         CreatedAt:      timestamppb.Now(),
//         UpdatedAt:      timestamppb.Now(),
//         TotalPrice:     int32(req.TotalPrice),
//         DishItems:      ToPBDishOrderItems(req.DishItems),
//         SetItems:       ToPBSetOrderItems(req.SetItems),
//         BowChili:       int64(req.BowChili),
//         BowNoChili:     int64(req.BowNoChili),
//         TakeAway:       req.TakeAway,
//         ChiliNumber:    int64(req.ChiliNumber),
//         TableToken:     req.TableToken,
//         OrderName:      req.OrderName,
//     }
// }

// MockResponseWriter implements http.ResponseWriter

//

func (s *webSocketService) handleOrderMessage(message *websocket_model.Message) error {
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go")
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go %v, message: ", message.Content)
	// Convert message content to JSON bytes
	contentBytes, err := json.Marshal(message.Content)
	if err != nil {
		return fmt.Errorf("error marshaling content: %v", err)
	}

	var orderContent OrderContent
	if err := json.Unmarshal(contentBytes, &orderContent); err != nil {
		log.Printf("Error unmarshaling order content: %v, content: %s", err, string(contentBytes))
		return fmt.Errorf("error unmarshaling order content: %v", err)
	}
	log.Printf("golang/ecomm-api/websocket/websocket_service/websocket_service.go %v, orderContent: ", orderContent)

	// Create the CreateOrderRequestType from the Order data
	requestData := order_grpc.CreateOrderRequestType{
		GuestID:        orderContent.Order.GuestID,
		UserID:         orderContent.Order.UserID,
		IsGuest:        orderContent.Order.IsGuest,
		TableNumber:    orderContent.Order.TableNumber,
		OrderHandlerID: orderContent.Order.OrderHandlerID,
		Status:         orderContent.Order.Status,
		TotalPrice:     orderContent.Order.TotalPrice,
		OrderName:      orderContent.Order.OrderName,
		DishItems:      orderContent.Order.DishItems,
		SetItems:       orderContent.Order.SetItems,
		Topping:        orderContent.Order.Topping,
		TrackingOrder:  orderContent.Order.TrackingOrder,
		TakeAway:       orderContent.Order.TakeAway,
		ChiliNumber:    orderContent.Order.ChiliNumber,
		TableToken:     orderContent.Order.TableToken,
	}

	// Create a request body
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestData); err != nil {
		return fmt.Errorf("error encoding request data: %v", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, "/orders", &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a mock response writer
	rw := NewMockResponseWriter()

	// Call the CreateOrder handler
	s.orderHandler.CreateOrder(rw, req)

	// Check the response status
	if rw.Status >= 400 {
		return fmt.Errorf("error creating order: %s", rw.Body.String())
	}

	// Parse the response
	var orderResponse map[string]interface{}
	if err := json.NewDecoder(&rw.Body).Decode(&orderResponse); err != nil {
		return fmt.Errorf("error decoding order response: %v", err)
	}

	// Create response message
	responseMessage := &websocket_model.Message{
		Type:      "ORDER_CREATED",
		Content:   orderResponse,
		Sender:    message.Sender,
		ToUser:    message.ToUser,
		Timestamp: time.Now(),
		TableID:   message.TableID,
		OrderID:   fmt.Sprintf("%v", orderResponse["order_id"]),
	}

	// Send response back to the client
	return s.SendToClient(ClientIdentifier{
		ID:   message.ToUser,
		Type: GuestClient,
	}, responseMessage)
}

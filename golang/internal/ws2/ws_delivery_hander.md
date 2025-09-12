package ws2

import (
	"bytes"
	"context"
	"encoding/json"
	delivery "english-ai-full/internal/delivery"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type DeliveryMessageHandler struct {
	// DefaultMessageHandler
	deliveryHandler *delivery.DeliveryHandlerController
	broadcaster     *Broadcaster
}

func NewDeliveryMessageHandler(deliveryHandler *delivery.DeliveryHandlerController) *DeliveryMessageHandler {
	return &DeliveryMessageHandler{
		deliveryHandler: deliveryHandler,
	}
}

func (h *DeliveryMessageHandler) SetBroadcaster(b *Broadcaster) {
	h.broadcaster = b
}

func (h *DeliveryMessageHandler) Handle(c *Client, msg Message) {
	log.Printf("BEGIN DeliveryMessageHandler.Handle - Type: %s, Action: %s", msg.Type, msg.Action)
	defer log.Printf("END DeliveryMessageHandler.Handle")

	switch msg.Action {
	case "create":
		h.handleDeliveryMessageToStaff(c, msg)
		h.handleDirectMessageDelivery(c, msg)
	case "update_status":
		h.handleUpdateDeliveryStatus(msg)
	case "assign":
		h.handleAssignDelivery(msg)
	}
	// default:
	//     h.DefaultMessageHandler.HandleMessage(c, msg)

}

func (h *DeliveryMessageHandler) handleDeliveryMessageToStaff(c *Client, msg Message) {
	log.Printf("golang/quanqr/ws2/ws_delivery_hander.go BEGIN handleDeliveryMessageToStaff")
	defer log.Printf("END handleDeliveryMessageToStaff")

	// Extract the direct message
	data, _ := json.Marshal(msg.Payload)
	var directMsg DirectMessage
	if err := json.Unmarshal(data, &directMsg); err != nil {
		log.Printf("error unmarshaling direct message: %v", err)
		return
	}

	// First, create the delivery in the database
	if err := h.createDelivery(directMsg.Payload); err != nil {
		log.Printf("Error creating delivery: %v", err)
		return
	}

	// After successful delivery creation, broadcast to staff
	staffMsg := Message{
		Type:    "delivery",
		Action:  "new_delivery",
		Payload: directMsg.Payload,
	}

	// Broadcast to all staff members
	if err := c.Hub.BroadcastToStaff(directMsg.FromUserID, staffMsg); err != nil {
		log.Printf("Error broadcasting delivery to staff: %v", err)
		return
	}

	log.Printf("Successfully created delivery and broadcasted to staff")
}

func (h *DeliveryMessageHandler) handleUpdateDeliveryStatus(msg Message) {
	var updateReq struct {
		DeliveryID string `json:"delivery_id"`
		Status     string `json:"status"`
	}

	data, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(data, &updateReq); err != nil {
		log.Printf("error unmarshaling status update: %v", err)
		return
	}

	staffMsg := Message{
		Type:   "delivery",
		Action: "status_updated",
		Payload: map[string]interface{}{
			"delivery_id": updateReq.DeliveryID,
			"status":      updateReq.Status,
		},
	}

	h.broadcaster.BroadcastToStaff(staffMsg)
}

func (h *DeliveryMessageHandler) handleAssignDelivery(msg Message) {
	var assignReq struct {
		DeliveryID string `json:"delivery_id"`
		DriverID   string `json:"driver_id"`
	}

	data, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(data, &assignReq); err != nil {
		log.Printf("error unmarshaling assignment request: %v", err)
		return
	}

	// Assign the delivery using the controller
	// err := h.deliveryHandler.AssignDriver(assignReq.DeliveryID, assignReq.DriverID)
	// if err != nil {
	//     log.Printf("Error assigning delivery: %v", err)
	//     return
	// }

	// Broadcast assignment to staff
	staffMsg := Message{
		Type:   "delivery",
		Action: "delivery_assigned",
		Payload: map[string]interface{}{
			"delivery_id": assignReq.DeliveryID,
			"driver_id":   assignReq.DriverID,
		},
	}

	h.broadcaster.BroadcastToStaff(staffMsg)
}

// heler
func (h *DeliveryMessageHandler) createDelivery(payload interface{}) error {
	log.Printf("BEGIN createDelivery")
	defer log.Printf("END createDelivery")

	// Convert payload to map
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("error: payload is not a map[string]interface{}")
	}

	// [Previous helper functions remain the same...]
	safeFloat64 := func(v interface{}) float64 {
		if v == nil {
			return 0
		}
		switch i := v.(type) {
		case float64:
			return i
		case int:
			return float64(i)
		case int64:
			return float64(i)
		default:
			return 0
		}
	}

	safeBool := func(v interface{}) bool {
		if v == nil {
			return false
		}
		b, ok := v.(bool)
		if !ok {
			return false
		}
		return b
	}

	safeString := func(v interface{}) string {
		if v == nil {
			return ""
		}
		s, ok := v.(string)
		if !ok {
			return ""
		}
		return s
	}

	safeInt64 := func(v interface{}) int64 {
		if v == nil {
			return 0
		}
		switch i := v.(type) {
		case float64:
			return int64(i)
		case int:
			return int64(i)
		case int64:
			return i
		default:
			return 0
		}
	}

	getMapValue := func(m map[string]interface{}, keys ...string) interface{} {
		for _, key := range keys {
			if val, exists := m[key]; exists && val != nil {
				return val
			}
		}
		return nil
	}

	// Parse dish items
	var dishItems []delivery.DishDeliveryItem
	if rawDishItems, ok := getMapValue(payloadMap, "dish_items", "dishItems").([]interface{}); ok {
		for _, item := range rawDishItems {
			if dishMap, ok := item.(map[string]interface{}); ok {
				dishItems = append(dishItems, delivery.DishDeliveryItem{
					DishID:   int64(safeFloat64(getMapValue(dishMap, "dish_id", "dishId"))),
					Quantity: int64(safeFloat64(getMapValue(dishMap, "quantity"))),
				})
			}
		}
	}

	isGuest := safeBool(getMapValue(payloadMap, "is_guest", "isGuest"))

	// Parse scheduled time
	var scheduledTime time.Time
	if timeStr := safeString(getMapValue(payloadMap, "scheduled_time", "scheduledTime")); timeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, timeStr); err == nil {
			scheduledTime = parsed
		} else {
			scheduledTime = time.Now()
		}
	} else {
		scheduledTime = time.Now()
	}

	deliveryReq := delivery.CreateDeliveryRequest{
		GuestID:         safeInt64(getMapValue(payloadMap, "guest_id", "guestId")), // Updated to use guest_id
		UserID:          safeInt64(getMapValue(payloadMap, "user_id", "userId")),   // Updated to use user_id
		IsGuest:         isGuest,                                                   // Determined based on payload
		TableNumber:     int64(safeFloat64(getMapValue(payloadMap, "table_number", "tableNumber"))),
		OrderHandlerID:  safeInt64(getMapValue(payloadMap, "order_handler_id", "orderHandlerId")),
		Status:          safeString(getMapValue(payloadMap, "status")), // Status field
		TotalPrice:      int32(safeFloat64(getMapValue(payloadMap, "total_price", "totalPrice"))),
		DishItems:       dishItems, // Dish items array
		BowChili:        int64(safeFloat64(getMapValue(payloadMap, "bow_chili", "bowChili"))),
		BowNoChili:      int64(safeFloat64(getMapValue(payloadMap, "bow_no_chili", "bowNoChili"))),
		TakeAway:        safeBool(getMapValue(payloadMap, "take_away", "takeAway")),
		ChiliNumber:     int64(safeFloat64(getMapValue(payloadMap, "chili_number", "chiliNumber"))),
		TableToken:      safeString(getMapValue(payloadMap, "table_token", "tableToken")),
		ClientName:      safeString(getMapValue(payloadMap, "client_name", "clientName")),
		DeliveryAddress: safeString(getMapValue(payloadMap, "delivery_address", "deliveryAddress")),
		DeliveryContact: safeString(getMapValue(payloadMap, "delivery_contact", "deliveryContact")),
		DeliveryNotes:   safeString(getMapValue(payloadMap, "delivery_notes", "deliveryNotes")),
		ScheduledTime:   scheduledTime,                                             // Scheduled time field
		OrderID:         safeInt64(getMapValue(payloadMap, "order_id", "orderId")), // Order ID field
		DeliveryFee:     int32(safeFloat64(getMapValue(payloadMap, "delivery_fee", "deliveryFee"))),
		DeliveryStatus:  "Pending", // Defaulted to "pending"
	}
	// Modified ID assignment logic with snake_case support
	if isGuest {
		guestId := safeInt64(getMapValue(payloadMap, "guest_id", "guestId"))
		if guestId != 0 {
			deliveryReq.GuestID = guestId
		}
		deliveryReq.UserID = 0
	} else {
		userId := safeInt64(getMapValue(payloadMap, "user_id", "userId"))
		if userId != 0 {
			deliveryReq.UserID = userId
		}
		deliveryReq.GuestID = 0
	}

	// Validate required fields
	if deliveryReq.IsGuest && deliveryReq.GuestID == 0 {
		return fmt.Errorf("error: guest delivery requires guest_id")
	}

	if !deliveryReq.IsGuest && deliveryReq.UserID == 0 {
		return fmt.Errorf("error: user delivery requires user_id")
	}

	if deliveryReq.DeliveryAddress == "" {
		return fmt.Errorf("error: delivery address is required")
	}

	if deliveryReq.DeliveryContact == "" {
		return fmt.Errorf("error: delivery contact is required")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Serialize the delivery request to JSON
	deliveryReqJSON, err := json.Marshal(deliveryReq)
	if err != nil {
		return fmt.Errorf("error marshaling delivery request: %v", err)
	}

	bodyReader := bytes.NewReader(deliveryReqJSON)

	// bodyReader := bytes.NewBuffer(deliveryReqJSON)
	log.Printf("golang/quanqr/ws2/ws_delivery_hander.go bodyReader deliveryReqJSON  %+v", deliveryReq)
	// Create the request with the serialized delivery data
	r := &http.Request{
		Method: "POST",
		URL:    &url.URL{},
		Header: make(http.Header),
		Body:   io.NopCloser(bodyReader),
		GetBody: func() (io.ReadCloser, error) {
			r := bytes.NewReader(deliveryReqJSON)
			return io.NopCloser(r), nil
		},
		ContentLength: int64(len(deliveryReqJSON)),
	}
	r.Header.Set("Content-Type", "application/json")

	// Create the delivery
	w := NewResponseWriter()
	h.deliveryHandler.CreateDelivery3(w, r.WithContext(ctx))

	if w.StatusCode != http.StatusCreated {
		return fmt.Errorf("error creating delivery: received status code %d", w.StatusCode)
	}

	return nil
}

func (h *DeliveryMessageHandler) handleDirectMessageDelivery(c *Client, msg Message) {
	log.Printf("golang/quanqr/ws2/ws_delivery_hander.go BEGIN handleDirectMessage")
	defer log.Printf("golang/quanqr/ws2/ws_delivery_hander.go END handleDirectMessage")

	var directMsg DirectMessage
	data, _ := json.Marshal(msg.Payload)
	if err := json.Unmarshal(data, &directMsg); err != nil {
		log.Printf("error unmarshaling direct message: %v", err)
		return
	}

	log.Printf("golang/quanqr/ws2/ws_delivery_hander.go Sending direct message from %s to %s", directMsg.FromUserID, directMsg.ToUserID)
	log.Printf("golang/quanqr/ws2/ws_delivery_hander.go directMsg.Payload %v", directMsg.Payload)

	if err := c.Hub.SendDirectMessage(directMsg.FromUserID, directMsg.ToUserID, directMsg.Type, directMsg.Action, directMsg.Payload); err != nil {
		log.Printf("error sending direct message: %v", err)
	}
}

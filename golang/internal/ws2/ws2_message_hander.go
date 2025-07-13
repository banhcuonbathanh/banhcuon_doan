package ws2

import (
	"log"
)

type CombinedMessageHandler struct {
    orderHandler    *OrderMessageHandler
    deliveryHandler *DeliveryMessageHandler
}

func NewCombinedMessageHandler(orderHandler *OrderMessageHandler, deliveryHandler *DeliveryMessageHandler) *CombinedMessageHandler {
    return &CombinedMessageHandler{
        orderHandler:    orderHandler,
        deliveryHandler: deliveryHandler,
    }
}

func (h *CombinedMessageHandler) DispatcherMessage(c *Client, msg Message) {
    switch msg.Type {
    case "order":
        h.orderHandler.Handle(c, msg)
    case "delivery":
        h.deliveryHandler.Handle(c, msg)
    default:
        // Handle other message types or use default handler
        log.Printf(" golang/quanqr/ws2/ws2_message_hander.go Unhandled message type: %s", msg.Type)
    }
}
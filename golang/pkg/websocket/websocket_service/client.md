package websocket_service

import (
	"log"
	"time"

	websocket_model "english-ai-full/pkg/websocket/websocker_model"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 30 * time.Second // Increased from 10s
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 65536
)

type Client struct {
	conn     *websocket.Conn
	send     chan *websocket_model.Message
	service  WebSocketService
	userID   string
	userName string
	isGuest  bool
	closed   bool
}

func NewClient(conn *websocket.Conn, service WebSocketService, userID string, userName string, isGuest bool) *Client {
	return &Client{
		conn:     conn,
		send:     make(chan *websocket_model.Message, 256),
		service:  service,
		userID:   userID,
		userName: userName,
		isGuest:  isGuest,
		closed:   false,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		if !c.closed {
			c.closed = true
			c.service.UnregisterClient(c)
			c.conn.Close()
			log.Printf("ReadPump closed for %s user %s", c.getClientType(), c.userID)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var message websocket_model.Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ReadPump error for %s user %s: %v", c.getClientType(), c.userID, err)
			}
			break
		}

		// Set sender information
		message.Sender = c.userID
		message.FromUser = c.userID
		message.Timestamp = time.Now()

		log.Printf("Received message from %s user %s to %s: %v",
			c.getClientType(), c.userID, message.ToUser, message.Type)

		// Handle order messages
		if message.Type == "CHAT_MESSAGE" {
			if orderService, ok := c.service.(*webSocketService); ok {
				err = orderService.handleOrderMessage(&message)
				if err != nil {
					log.Printf("Error handling order message: %v", err)
					continue
				}
			}
		}

		if message.ToUser != "" {
			// Handle direct message
			if c.isGuest {
				err = c.service.SendMessageToGuest(
					message.FromUser,
					message.ToUser,
					message.Type,
					message.Content,
					message.TableID,
					message.OrderID,
				)
			} else {
				err = c.service.SendMessageToUser(
					message.FromUser,
					message.ToUser,
					message.Type,
					message.Content,
					message.TableID,
					message.OrderID,
				)
			}

			if err != nil {
				log.Printf("Error sending direct message from %s user %s: %v",
					c.getClientType(), c.userID, err)
			}
		} else {
			// Broadcast message
			c.service.BroadcastMessage(&message)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if !c.closed {
			c.closed = true
			c.conn.Close()
			log.Printf("WritePump closed for %s user %s", c.getClientType(), c.userID)
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(message)
			if err != nil {
				log.Printf("Error writing message to %s user %s: %v",
					c.getClientType(), c.userID, err)
				return
			}
			log.Printf("Successfully sent message to %s user %s",
				c.getClientType(), c.userID)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping to %s user %s: %v",
					c.getClientType(), c.userID, err)
				return
			}
			log.Printf("Sent ping to %s user %s", c.getClientType(), c.userID)
		}
	}
}

func (c *Client) getClientType() string {
	if c.isGuest {
		return "guest"
	}
	return "user"
}

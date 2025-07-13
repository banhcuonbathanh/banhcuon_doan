package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)


const (
    writeWait      = 30 * time.Second  // Increased from 10s
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 65536
)


type Client struct {
    conn     *websocket.Conn
    send     chan *Message
    service  WebSocketService
    userID   int64
    userName string
    isGuest  bool
    closed   bool
}

func NewClient(conn *websocket.Conn, service WebSocketService, userID int64, userName string, isGuest bool) *Client {
    return &Client{
        conn:     conn,
        send:     make(chan *Message, 256),
        service:  service,
        userID:   userID,
        userName: userName,
        isGuest:  isGuest,
        closed:   false,
    }
}

func (c *Client) readPump() {
    defer func() {
        log.Printf("Client readPump ended for userID: %d, userName: %s, isGuest: %t", c.userID, c.userName, c.isGuest)
        c.service.UnregisterClient(c)
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("readPump error for userID: %d, userName: %s, isGuest: %t: %v", c.userID, c.userName, c.isGuest, err)
            }
            break
        }

        // Process incoming message
        msg := &Message{}
        if err := json.Unmarshal(message, msg); err != nil {
            log.Printf("Error unmarshaling message for userID: %d, userName: %s, isGuest: %t: %v", c.userID, c.userName, c.isGuest, err)
            continue
        }

        // Handle different message types
        switch msg.Type {
        case "order_update":
            c.handleOrderUpdate(msg)
        case "delivery_update":
            c.handleDeliveryUpdate(msg)
        default:
            log.Printf("Received unknown message type '%s' for userID: %d, userName: %s, isGuest: %t", msg.Type, c.userID, c.userName, c.isGuest)
        }
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        log.Printf("Client writePump ended for userID: %d, userName: %s, isGuest: %t", c.userID, c.userName, c.isGuest)
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                log.Printf("Client writePump channel closed for userID: %d, userName: %s, isGuest: %t", c.userID, c.userName, c.isGuest)
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                log.Printf("Error getting writer for userID: %d, userName: %s, isGuest: %t: %v", c.userID, c.userName, c.isGuest, err)
                return
            }

            if err := json.NewEncoder(w).Encode(message); err != nil {
                log.Printf("Error encoding message for userID: %d, userName: %s, isGuest: %t: %v", c.userID, c.userName, c.isGuest, err)
            }

            if err := w.Close(); err != nil {
                log.Printf("Error closing writer for userID: %d, userName: %s, isGuest: %t: %v", c.userID, c.userName, c.isGuest, err)
                return
            }
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                log.Printf("Error writing ping message for userID: %d, userName: %s, isGuest: %t: %v", c.userID, c.userName, c.isGuest, err)
                return
            }
        }
    }
}
func (c *Client) handleOrderUpdate(msg *Message) {
    // Implement the logic to handle order update messages
    log.Printf("Received order update message: %v", msg)
}

func (c *Client) handleDeliveryUpdate(msg *Message) {
    // Implement the logic to handle delivery update messages
    log.Printf("Received delivery update message: %v", msg)
}
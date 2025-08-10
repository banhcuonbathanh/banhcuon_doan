# Middleware & WebSocket Package Documentation

This documentation covers the authentication middleware and real-time WebSocket communication system for a Go application, specifically designed for restaurant/table management systems.

## Table of Contents

- [Package Overview](#package-overview)
- [Authentication Middleware](#authentication-middleware)
- [WebSocket System](#websocket-system)
- [Package Structure](#package-structure)
- [Usage Guide](#usage-guide)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Package Overview

The system consists of two main components:

### 1. Authentication Middleware (`pkg/middleware/auth/`)
- JWT token validation
- Role-based access control
- HTTP request authentication
- Error handling for unauthorized access

### 2. WebSocket System (`pkg/websocket/`)
- Real-time bidirectional communication
- User and guest client management
- Order processing integration
- Message broadcasting and direct messaging

## Package Structure

```
pkg/
├── middleware/
│   └── auth/
│       ├── errors.go           # Authentication error definitions
│       └── middleware.go       # JWT authentication middleware
├── pkg/
│   └── http/
│       └── httpserve.go       # HTTP error response handler
└── websocket/
    ├── websocker_model/       # Data models for WebSocket messages
    │   └── websocker_message_model.go
    ├── websocket_handler/     # WebSocket connection handlers
    │   └── websocket_handler.go
    ├── websocket_repository/  # Message storage interface
    │   └── message_repository.go
    └── websocket_service/     # Core WebSocket business logic
        ├── client.go          # Client connection management
        └── websocket_service.go # Service orchestration
```

## Authentication Middleware

### Core Components

#### 1. Error Definitions (`errors.go`)
```go
var (
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidToken = errors.New("invalid token")
)
```

#### 2. Role System
```go
type Role string

const (
    RoleGuest    Role = "guest"
    RoleAdmin    Role = "admin"
    RoleEmployee Role = "employee"
    RoleOwner    Role = "owner"
)
```

#### 3. Authentication Functions
- **Token Verification**: Validates JWT tokens from Authorization headers
- **Role Checking**: Ensures users have appropriate permissions
- **Context Enhancement**: Adds user information to request context

### Key Features

- **Bearer Token Support**: Standard `Authorization: Bearer <token>` format
- **Context Injection**: User claims available in request context
- **Error Handling**: Standardized HTTP error responses
- **Token Parsing**: Integration with utils package for token validation

## WebSocket System

### Architecture Overview

The WebSocket system uses a hub-and-spoke model with the following components:

1. **WebSocketHandler**: HTTP upgrade and connection management
2. **WebSocketService**: Central message routing and client management
3. **Client**: Individual connection wrapper with read/write pumps
4. **Repository**: Message persistence layer

### Client Types

```go
type ClientType string

const (
    UserClient  ClientType = "user"
    GuestClient ClientType = "guest"
)
```

### Message Flow

1. **Connection**: Client connects via HTTP upgrade
2. **Registration**: Client registered with service
3. **Authentication**: User/guest identification
4. **Messaging**: Bidirectional message exchange
5. **Cleanup**: Graceful connection termination

## Usage Guide

### Authentication Middleware Setup

#### Basic Authentication

```go
import (
    "english-ai-full/pkg/middleware/auth"
    "english-ai-full/token"
)

func setupRoutes() {
    // Protected routes
    protectedRouter := mux.NewRouter()
    protectedRouter.Use(auth.AuthMiddleware)
    
    // Your protected endpoints
    protectedRouter.HandleFunc("/api/users", getUsersHandler)
    protectedRouter.HandleFunc("/api/orders", getOrdersHandler)
}
```

#### Role-Based Protection

```go
func roleProtectedHandler(allowedRoles []auth.Role) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get user from context (set by AuthMiddleware)
        userClaims := r.Context().Value("user").(jwt.MapClaims)
        userRole := auth.Role(userClaims["role"].(string))
        
        if !isRoleAllowed(userRole, allowedRoles) {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        // Handle the request
    }
}

// Usage
adminOnly := roleProtectedHandler([]auth.Role{auth.RoleAdmin, auth.RoleOwner})
http.HandleFunc("/admin/dashboard", adminOnly)
```

### WebSocket System Setup

#### 1. Initialize Components

```go
import (
    "english-ai-full/pkg/websocket/websocket_handler"
    "english-ai-full/pkg/websocket/websocket_service"
    "english-ai-full/pkg/websocket/websocket_repository"
)

func setupWebSocket() {
    // Initialize repository
    messageRepo := websocket_repository.NewInMemoryMessageRepository()
    
    // Initialize service
    wsService := websocket_service.NewWebSocketService(messageRepo, orderHandler)
    
    // Initialize handler
    wsHandler := websocket_handler.NewWebSocketHandler(wsService)
    
    // Start the service
    go wsService.Run()
    
    // Setup HTTP endpoints
    http.HandleFunc("/ws", wsHandler.HandleWebSocket)
    http.HandleFunc("/api/send-message", wsHandler.HandleSendMessage)
}
```

#### 2. Client Connection

```go
// Client-side JavaScript example
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function(event) {
    console.log('Connected to WebSocket');
    
    // Send initial message
    ws.send(JSON.stringify({
        type: 'user_join',
        content: { userId: '123', userName: 'John Doe' },
        timestamp: new Date().toISOString()
    }));
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
};
```

## API Reference

### Authentication Middleware

#### `AuthMiddleware(next http.Handler) http.Handler`
Main authentication middleware function.

**Features:**
- Validates Bearer tokens
- Adds user claims to context
- Returns 401 for invalid/missing tokens

**Context Keys:**
- `"user"`: Contains JWT claims map

#### `verifyClaimsFromAuthHeader(r *http.Request, tokenMaker *token.JWTMaker) (*token.UserClaims, error)`
Extracts and verifies JWT token from Authorization header.

**Parameters:**
- `r`: HTTP request
- `tokenMaker`: JWT token maker instance

**Returns:**
- User claims if valid
- Error if invalid/missing

### WebSocket Service Interface

```go
type WebSocketService interface {
    RegisterClient(client *Client)
    UnregisterClient(client *Client)
    BroadcastMessage(message *websocket_model.Message)
    SendMessageToUser(fromUser, toUser string, messageType string, content interface{}, tableID, orderID string) error
    SendMessageToGuest(fromUser string, guestID string, messageType string, content interface{}, tableID, orderID string) error
    Run()
}
```

### Message Models

#### Core Message Structure
```go
type Message struct {
    Type      string      `json:"type"`
    Content   interface{} `json:"content"`
    Sender    string      `json:"sender"`
    Recipient string      `json:"recipient,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
    TableID   string      `json:"table_id,omitempty"`
    OrderID   string      `json:"order_id,omitempty"`
    ID        string      `json:"id,omitempty"`
    FromUser  string      `json:"fromUser"`
    ToUser    string      `json:"toUser"`
}
```

#### Order-Specific Models
```go
type CreateOrderRequest struct {
    GuestID        *int            `json:"guest_id"`
    UserID         *int            `json:"user_id"`
    IsGuest        bool            `json:"is_guest"`
    TableNumber    int             `json:"table_number"`
    OrderHandlerID int             `json:"order_handler_id"`
    Status         string          `json:"status"`
    TotalPrice     float64         `json:"total_price"`
    DishItems      []DishOrderItem `json:"dish_items"`
    SetItems       []SetOrderItem  `json:"set_items"`
    // ... other fields
}
```

## Examples

### Complete Restaurant System Implementation

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "english-ai-full/pkg/middleware/auth"
    "english-ai-full/pkg/websocket/websocket_handler"
    "english-ai-full/pkg/websocket/websocket_service"
    "english-ai-full/pkg/websocket/websocket_repository"
    "english-ai-full/token"
)

func main() {
    // Initialize token maker
    tokenMaker := token.NewJWTMaker("your-secret-key")
    
    // Setup WebSocket system
    messageRepo := websocket_repository.NewInMemoryMessageRepository()
    wsService := websocket_service.NewWebSocketService(messageRepo, nil)
    wsHandler := websocket_handler.NewWebSocketHandler(wsService)
    
    // Start WebSocket service
    go wsService.Run()
    
    // Setup routes
    mux := http.NewServeMux()
    
    // WebSocket endpoint (requires authentication)
    mux.HandleFunc("/ws", auth.AuthMiddleware(
        http.HandlerFunc(wsHandler.HandleWebSocket),
    ).ServeHTTP)
    
    // API endpoints
    mux.HandleFunc("/api/send-message", auth.AuthMiddleware(
        http.HandlerFunc(wsHandler.HandleSendMessage),
    ).ServeHTTP)
    
    // Role-protected endpoints
    mux.HandleFunc("/admin/dashboard", auth.AuthMiddleware(
        roleProtected([]auth.Role{auth.RoleAdmin, auth.RoleOwner}, 
            http.HandlerFunc(adminDashboard)),
    ).ServeHTTP)
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}

func roleProtected(allowedRoles []auth.Role, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userClaims := r.Context().Value("user").(map[string]interface{})
        userRole := auth.Role(userClaims["role"].(string))
        
        if !isRoleAllowed(userRole, allowedRoles) {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func adminDashboard(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Admin Dashboard"))
}
```

### Table Management WebSocket Flow

```go
// Client connects with table token
func handleTableConnection(wsService websocket_service.WebSocketService) {
    // When guest scans QR code at table
    message := &websocket_model.Message{
        Type: "table_join",
        Content: map[string]interface{}{
            "table_number": 5,
            "guest_name": "John Doe",
            "table_token": "short-table-token-here",
        },
        Sender: "guest_123",
    }
    
    // Broadcast table occupancy
    wsService.BroadcastMessage(message)
    
    // Notify staff
    wsService.SendMessageToUser(
        "guest_123", 
        "staff_456", 
        "guest_seated",
        map[string]interface{}{
            "table": 5,
            "guest": "John Doe",
            "time": time.Now(),
        },
        "table_5",
        "",
    )
}
```

### Order Processing via WebSocket

```go
// Handle order placement through WebSocket
func processOrder(wsService websocket_service.WebSocketService) {
    orderMessage := &websocket_model.Message{
        Type: "CHAT_MESSAGE",
        Content: map[string]interface{}{
            "order": map[string]interface{}{
                "guest_id": 123,
                "table_number": 5,
                "dish_items": []map[string]interface{}{
                    {"dish_id": 1, "quantity": 2},
                    {"dish_id": 3, "quantity": 1},
                },
                "total_price": 25.99,
                "table_token": "table-token-here",
            },
        },
        FromUser: "guest_123",
        ToUser: "kitchen_staff",
        TableID: "table_5",
    }
    
    // This will trigger order processing in the service
    wsService.BroadcastMessage(orderMessage)
}
```

### Frontend Integration

```javascript
class RestaurantWebSocket {
    constructor(wsUrl, authToken) {
        this.wsUrl = wsUrl;
        this.authToken = authToken;
        this.ws = null;
        this.messageHandlers = new Map();
    }
    
    connect() {
        // Note: WebSocket doesn't support custom headers
        // Token should be passed as query parameter or during handshake
        this.ws = new WebSocket(`${this.wsUrl}?token=${this.authToken}`);
        
        this.ws.onopen = (event) => {
            console.log('Connected to restaurant WebSocket');
            this.sendJoinMessage();
        };
        
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
        
        this.ws.onclose = (event) => {
            console.log('WebSocket connection closed');
            // Implement reconnection logic
            setTimeout(() => this.connect(), 5000);
        };
    }
    
    sendJoinMessage() {
        this.send({
            type: 'user_join',
            content: {
                userId: this.userId,
                userName: this.userName
            }
        });
    }
    
    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                ...message,
                timestamp: new Date().toISOString(),
                id: this.generateId()
            }));
        }
    }
    
    // Order placement
    placeOrder(orderData) {
        this.send({
            type: 'CHAT_MESSAGE',
            content: { order: orderData },
            toUser: 'kitchen_staff'
        });
    }
    
    // Direct messaging
    sendMessage(toUser, content, isGuest = false) {
        this.send({
            type: 'direct_message',
            content: content,
            toUser: toUser,
            isGuest: isGuest
        });
    }
    
    handleMessage(message) {
        const handler = this.messageHandlers.get(message.type);
        if (handler) {
            handler(message);
        } else {
            console.log('Unhandled message type:', message.type, message);
        }
    }
    
    onMessageType(type, handler) {
        this.messageHandlers.set(type, handler);
    }
    
    generateId() {
        return Date.now().toString(36) + Math.random().toString(36).substr(2);
    }
}

// Usage
const restaurantWS = new RestaurantWebSocket('ws://localhost:8080/ws', authToken);

// Set up message handlers
restaurantWS.onMessageType('ORDER_CREATED', (message) => {
    console.log('Order created:', message.content);
    updateOrderStatus(message.content);
});

restaurantWS.onMessageType('direct_message', (message) => {
    displayChatMessage(message);
});

restaurantWS.onMessageType('ORDER_STATUS_UPDATE', (message) => {
    updateOrderStatus(message.content);
});

// Connect
restaurantWS.connect();
```

## Best Practices

### Security

1. **Token Validation**
   ```go
   // ✅ Always validate tokens
   claims, err := tokenMaker.VerifyToken(token)
   if err != nil {
       return handleAuthError(w, err)
   }
   
   // ✅ Check token expiration
   if time.Now().After(claims.ExpiresAt.Time) {
       return handleAuthError(w, ErrTokenExpired)
   }
   ```

2. **Role-Based Access**
   ```go
   // ✅ Implement granular permissions
   func requireRoles(roles ...Role) middleware {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               userRole := getUserRole(r.Context())
               if !isRoleAllowed(userRole, roles) {
                   http.Error(w, "Forbidden", http.StatusForbidden)
                   return
               }
               next.ServeHTTP(w, r)
           })
       }
   }
   ```

3. **WebSocket Authentication**
   ```go
   // ✅ Validate WebSocket connections
   func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
       // Authentication should happen before upgrade
       userID := r.Context().Value("userID").(string)
       if userID == "" {
           http.Error(w, "Unauthorized", http.StatusUnauthorized)
           return
       }
       
       // Proceed with upgrade
       conn, err := h.upgrader.Upgrade(w, r, nil)
       // ...
   }
   ```

### Performance

1. **Connection Management**
   ```go
   // ✅ Set appropriate timeouts
   const (
       writeWait      = 30 * time.Second
       pongWait       = 60 * time.Second
       pingPeriod     = (pongWait * 9) / 10
       maxMessageSize = 65536
   )
   ```

2. **Memory Management**
   ```go
   // ✅ Clean up connections
   defer func() {
       if !c.closed {
           c.closed = true
           c.service.UnregisterClient(c)
           c.conn.Close()
       }
   }()
   ```

3. **Message Buffering**
   ```go
   // ✅ Use buffered channels
   send: make(chan *websocket_model.Message, 256)
   ```

### Error Handling

1. **Graceful Degradation**
   ```go
   // ✅ Handle WebSocket failures gracefully
   select {
   case client.send <- message:
       log.Printf("Message sent successfully")
   default:
       log.Printf("Failed to send message, client may be disconnected")
       // Don't block, continue with other clients
   }
   ```

2. **Logging**
   ```go
   // ✅ Comprehensive logging
   log.Printf("Client %s (%s) connected from %s", 
       client.userID, client.getClientType(), r.RemoteAddr)
   ```

### Testing

```go
func TestAuthMiddleware(t *testing.T) {
    tokenMaker := token.NewJWTMaker("test-secret")
    
    // Create test token
    token, _, err := tokenMaker.CreateToken(123, "test@example.com", "admin", time.Hour)
    require.NoError(t, err)
    
    // Create test request
    req := httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    // Test middleware
    handler := auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := r.Context().Value("user")
        assert.NotNil(t, user)
        w.WriteHeader(http.StatusOK)
    }))
    
    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWebSocketService(t *testing.T) {
    repo := websocket_repository.NewInMemoryMessageRepository()
    service := websocket_service.NewWebSocketService(repo, nil)
    
    // Test client registration
    // Test message sending
    // Test cleanup
}
```

This comprehensive system provides robust authentication and real-time communication capabilities for restaurant management applications, with support for both staff users and guest customers, order processing, and table management.
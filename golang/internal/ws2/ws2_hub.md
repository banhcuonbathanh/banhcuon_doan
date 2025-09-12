package ws2

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type ClientInfo struct {
    ID       string                 `json:"id"`
    Role     Role                   `json:"role"`
    RoomID   string                `json:"roomId"`
    JoinedAt time.Time             `json:"joinedAt"`
    UserData map[string]interface{} `json:"userData"`
}

type Hub struct {
    Clients               map[*Client]bool
    Broadcast            chan []byte
    Register             chan *Client
    Unregister           chan *Client
    RoomMap              map[string]map[*Client]bool
    CombinedMessageHandler *CombinedMessageHandler
    mu                   sync.Mutex
    RegisteredClients    map[string]ClientInfo
}

func NewHub(combinedMessageHandler *CombinedMessageHandler) *Hub {
    log.Printf("Creating new Hub with message handler type: %T", combinedMessageHandler)
    return &Hub{
        Broadcast:            make(chan []byte),
        Register:            make(chan *Client),
        Unregister:          make(chan *Client),
        Clients:             make(map[*Client]bool),
        RoomMap:             make(map[string]map[*Client]bool),
        CombinedMessageHandler: combinedMessageHandler,
        RegisteredClients:   make(map[string]ClientInfo),
    }
}



func (h *Hub) unregisterClient(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if _, ok := h.Clients[client]; ok {
        delete(h.Clients, client)
        delete(h.RegisteredClients, client.ID)
        
        if client.RoomID != "" && h.RoomMap[client.RoomID] != nil {
            delete(h.RoomMap[client.RoomID], client)
            if len(h.RoomMap[client.RoomID]) == 0 {
                delete(h.RoomMap, client.RoomID)
            }
        }
        close(client.Send)
        log.Printf("Client unregistered - ID: %s, Role: %s", client.ID, client.Role)
    }
}

func (h *Hub) broadcastMessage(message []byte) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    for client := range h.Clients {
        select {
        case client.Send <- message:
        default:
            close(client.Send)
            delete(h.Clients, client)
        }
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.Register:
            h.registerClient(client)
        case client := <-h.Unregister:
            h.unregisterClient(client)
        case message := <-h.Broadcast:
            h.broadcastMessage(message)
        }
    }
}

func (h *Hub) SendDirectMessage(fromUserID, toUserID string, msgType, action string, payload interface{}) error {
    h.mu.Lock()
    defer h.mu.Unlock()

    var targetClient *Client
    for client := range h.Clients {
        if client.Role == RoleUser && client.ID == toUserID {
            targetClient = client
            break
        }
    }

    if targetClient == nil {
        return fmt.Errorf("target user %s not found", toUserID)
    }

    directMsg := DirectMessage{
        FromUserID: fromUserID,
        ToUserID:   toUserID,
        Type:       msgType,
        Action:     action,
        Payload:    payload,
    }

    msg := Message{
        Type:    TypeMessage(msgType),
        Action:  action,
        Payload: directMsg,
        Role:    RoleUser,
    }

    data, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("error marshaling message: %v", err)
    }

    select {
    case targetClient.Send <- data:
        return nil
    default:
        close(targetClient.Send)
        delete(h.Clients, targetClient)
        return fmt.Errorf("failed to send message to user %s", toUserID)
    }
}

func (h *Hub) BroadcastToStaff(fromUserID string, msg Message) error {
    h.mu.Lock()
    defer h.mu.Unlock()

    msg.Role = RoleEmployee

    data, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("error marshaling message: %v", err)
    }

    // Log header for successful message sends
    log.Printf("+%s+%s+%s+%s+", 
        strings.Repeat("-", 36),   // FromUserID
        strings.Repeat("-", 15),   // Recipient Role 
        strings.Repeat("-", 36),   // Recipient Client ID
        strings.Repeat("-", 20),    // Status
    )
    log.Printf("| %-36s | %-15s | %-36s | %-20s |", 
        "From User ID", "Recipient Role", "Recipient Client ID", "Status")
    log.Printf("+%s+%s+%s+%s+", 
        strings.Repeat("-", 36), 
        strings.Repeat("-", 15), 
        strings.Repeat("-", 36),
        strings.Repeat("-", 20),
    )

    successfulSends := 0
    staffRoles := map[Role]bool{
        RoleAdmin:    true,
        RoleEmployee: true,
        RoleKitchen:  true,
    }

    for client := range h.Clients {
        if staffRoles[client.Role] {
            select {
            case client.Send <- data:
                successfulSends++
                log.Printf("| %-36s | %-15s | %-36s | %-20s |", 
                    fromUserID, 
                    client.Role, 
                    client.ID, 
                    "Message Sent ✓",
                )
            default:
                close(client.Send)
                delete(h.Clients, client)
                log.Printf("| %-36s | %-15s | %-36s | %-20s |", 
                    fromUserID, 
                    client.Role, 
                    client.ID, 
                    "Send Failed ✗",
                )
            }
        }
    }

    // Table footer
    log.Printf("+%s+%s+%s+%s+", 
        strings.Repeat("-", 36), 
        strings.Repeat("-", 15), 
        strings.Repeat("-", 36),
        strings.Repeat("-", 20),
    )
    log.Printf("Total Successful Sends: %d", successfulSends)

    if successfulSends == 0 {
        return fmt.Errorf("no staff members available to receive the message")
    }

    return nil
}

func (h *Hub) IsClientRegistered(clientID string) bool {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    _, exists := h.RegisteredClients[clientID]
    return exists
}

func (h *Hub) GetClientsByRole(role Role) []ClientInfo {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    var clients []ClientInfo
    for _, info := range h.RegisteredClients {
        if info.Role == role {
            clients = append(clients, info)
        }
    }
    return clients
}



// 





func (h *Hub) registerClient(client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    // First check if client already exists and has a different role
    if existingInfo, exists := h.RegisteredClients[client.ID]; exists {
        // If client exists with different role, log it
        if existingInfo.Role != client.Role {
            log.Printf("Client %s role changed from %s to %s", client.ID, existingInfo.Role, client.Role)
        }
    }
    
    h.Clients[client] = true
    if client.RoomID != "" {
        if h.RoomMap[client.RoomID] == nil {
            h.RoomMap[client.RoomID] = make(map[*Client]bool)
        }
        h.RoomMap[client.RoomID][client] = true
    }

    // Perform type assertion for UserData
    userData, ok := client.UserData.(map[string]interface{})
    if !ok {
        userData = make(map[string]interface{})
    }

    clientInfo := ClientInfo{
        ID:       client.ID,
        Role:     client.Role,
        RoomID:   client.RoomID,
        JoinedAt: time.Now(),
        UserData: userData,
    }
    
    // Only update if client doesn't exist or role has changed
    if existing, exists := h.RegisteredClients[client.ID]; !exists || existing.Role != client.Role {
        h.RegisteredClients[client.ID] = clientInfo
    }

    log.Printf("Client registered - ID: %s, Role: %s, Room: %s", client.ID, client.Role, client.RoomID)
    h.logRegisteredClients()
}





func (h *Hub) logRegisteredClients() {
    total := len(h.RegisteredClients)
    
    // Log total count
    log.Printf("Total Clients: %d\n+%s+%s+%s+", total, 
        strings.Repeat("-", 36), strings.Repeat("-", 15), strings.Repeat("-", 30))
    
    // Table header
    log.Printf("| %-36s | %-13s | %-28s |", "Client ID", "Role", "Email")
    log.Printf("+%s+%s+%s+", 
        strings.Repeat("-", 36), strings.Repeat("-", 15), strings.Repeat("-", 30))
    
    // Table content
    for _, info := range h.RegisteredClients {
        // Extract email from UserData, defaulting to "N/A" if not found
        email, _ := info.UserData["email"].(string)
        if email == "" {
            email = "N/A"
        }
        
        log.Printf("| %-36s | %-13s | %-28s |", info.ID, info.Role, email)
    }
    
    // Table footer
    log.Printf("+%s+%s+%s+", 
        strings.Repeat("-", 36), strings.Repeat("-", 15), strings.Repeat("-", 30))
}



func (h *Hub) ListClients() []map[string]interface{} {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    clientList := make([]map[string]interface{}, 0)
    
    for id, info := range h.RegisteredClients {
        clientInfo := map[string]interface{}{
            "id":       id,
            "role":     info.Role,
            "roomID":   info.RoomID,
            "joinedAt": info.JoinedAt,
            "active":   true,
            "userData": info.UserData,
        }
        clientList = append(clientList, clientInfo)
    }
    
    // Updated logging to use h.RegisteredClients directly
    log.Printf("golang/quanqr/ws2/ws2_hub.go Total connected clients: %d", len(h.RegisteredClients))
    
    // Accurate role-based counting using RegisteredClients
    for role := range map[Role]bool{
        RoleGuest: true,
        RoleUser: true,
        RoleEmployee: true,
        RoleAdmin: true,
        RoleKitchen: true,
    } {
        count := 0
        for _, info := range h.RegisteredClients {
            if info.Role == role {
                count++
            }
        }
        if count > 0 {
            log.Printf("Clients with role %s: %d", role, count)
        }
    }
    
    return clientList
}
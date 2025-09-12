package ws2

import (
	"encoding/json"
	"log"
)

type Broadcaster struct {
    hub *Hub
}

func NewBroadcaster(h *Hub) *Broadcaster {
    return &Broadcaster{hub: h}
}

func (b *Broadcaster) BroadcastToRole(msg Message, role Role) {
    data, _ := json.Marshal(msg)
    b.hub.mu.Lock()
    for client := range b.hub.Clients {
        if client.Role == role {
            select {
            case client.Send <- data:
            default:
                close(client.Send)
                delete(b.hub.Clients, client)
            }
        }
    }
    b.hub.mu.Unlock()
}

func (b *Broadcaster) BroadcastToRoom(roomID string, msg Message) {
    data, _ := json.Marshal(msg)
    b.hub.mu.Lock()
    if clients, ok := b.hub.RoomMap[roomID]; ok {
        for client := range clients {
            select {
            case client.Send <- data:
            default:
                close(client.Send)
                delete(b.hub.RoomMap[roomID], client)
                delete(b.hub.Clients, client)
            }
        }
    }
    b.hub.mu.Unlock()
}
// new 

func (b *Broadcaster) BroadcastToStaff(msg Message) {
    // Define staff roles
    staffRoles := map[Role]bool{
        RoleAdmin:    true,
        RoleEmployee: true,
        RoleKitchen:  true,
    }

    // Marshal the message once
    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Error marshaling message: %v", err)
        return
    }

    // Lock the hub to safely access clients
    b.hub.mu.Lock()
    defer b.hub.mu.Unlock()

    // Iterate through all clients
    for client := range b.hub.Clients {
        // Check if client's role is in staffRoles
        if staffRoles[client.Role] {
            select {
            case client.Send <- data:
                log.Printf("Message sent to %s role: %s", client.Role, string(data))
            default:
                log.Printf("Failed to send to client %s, removing from hub", client.ID)
                close(client.Send)
                delete(b.hub.Clients, client)
            }
        }
    }
}
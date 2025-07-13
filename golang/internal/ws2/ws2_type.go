package ws2

import "time"
type Role string
type TypeMessage string
const (
    RoleGuest    Role = "Guest"
    RoleUser     Role = "User"
    RoleEmployee Role = "Employee"
    RoleAdmin    Role = "Admin"
    RoleKitchen  Role = "Kitchen"
)

const (
    TypeDelivery    TypeMessage = "delivery"
 
)


type Message struct {
    Type    TypeMessage      `json:"type"`
    Action  string      `json:"action"`
    Payload interface{} `json:"payload"`
    Role    Role        `json:"role"`
    RoomID  string      `json:"roomId,omitempty"`
}



type DirectMessage struct {
    FromUserID string      `json:"fromUserId"`
    ToUserID   string      `json:"toUserId"`
    Type       string      `json:"type"`
    Action     string      `json:"action"`
    Payload    interface{} `json:"payload"`
}

type TokenRequestWS struct {
	UserID int64  `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
type TokenResponseWS struct {
	Token          string    `json:"token"`
	ExpiresAt      time.Time `json:"expiresAt"`
	Role           string    `json:"role"`
	UserID         int64     `json:"userId"`
	Email          string    `json:"email"`
}
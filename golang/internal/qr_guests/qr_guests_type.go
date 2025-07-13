package qr_guests

import (
	"time"
)
// type GuestInfo struct {
//     ID          int64     `json:"id"`
//     Name        string    `json:"name"`
//     Role        string    `json:"role"`
//     TableNumber int32     `json:"table_number"`
//     CreatedAt   time.Time `json:"created_at"`
//     UpdatedAt   time.Time `json:"updated_at"`
// }
type GuestInfo struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	TableNumber int32     `json:"table_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type Guest struct {
	ID                   int64     `json:"id"`
	Name                 string    `json:"name"`
	TableNumber          int32     `json:"table_number"`
	RefreshToken         string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type GuestLoginRequest struct {
	Name        string `json:"name"`
	TableNumber int32  `json:"table_number"`
	Token       string `json:"token"`
}
// new 
type GuestLoginResponse struct {

    AccessToken          string    `json:"access_token"`
    RefreshToken         string    `json:"refresh_token"`
    Guest                GuestInfo `json:"guest"`

    AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
    RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`


		    SessionID             string    `json:"session_id"`
}




type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Message      string `json:"message"`
}
type GuestGetOrdersGRPCRequest struct {
	GuestID int64 `json:"guestId"`
}


// type CreateOrderItem struct {
// 	DishID   int64 `json:"dish_id"`
// 	Quantity int32 `json:"quantity"`
// 	GuestID  int64 `json:"guest_id"`
// }

// type CreateOrdersRequest struct {
// 	Items []CreateOrderItem `json:"items"`
// }

// type Order struct {
// 	ID          int64     `json:"id"`
// 	GuestID     int64     `json:"guest_id"`
// 	TableNumber int32     `json:"table_number"`
// 	DishID      int64     `json:"dish_id"`
// 	Quantity    int32     `json:"quantity"`
// 	Status      string    `json:"status"`
// 	CreatedAt   time.Time `json:"created_at"`
// 	UpdatedAt   time.Time `json:"updated_at"`
// }

// type OrdersResponse struct {
// 	Data    []Order `json:"data"`
// 	Message string  `json:"message"`
// }



// type ListOrdersResponse struct {
// 	Orders  []Order `json:"orders"`
// 	Message string  `json:"message"`
// }
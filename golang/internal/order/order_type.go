package order_grpc

import (
	"time"
)
type OrderType struct {
    ID             int64           `json:"id"`
    GuestID        int64           `json:"guest_id"`
    UserID         int64           `json:"user_id"`
    IsGuest        bool            `json:"is_guest"`
    TableNumber    int64           `json:"table_number"`
    OrderHandlerID int64           `json:"order_handler_id"`
    Status         string          `json:"status"`
    CreatedAt      time.Time       `json:"created_at"`
    UpdatedAt      time.Time       `json:"updated_at"`
    TotalPrice     int32           `json:"total_price"`
    DishItems      []OrderDish     `json:"dish_items"`
    SetItems       []OrderSet      `json:"set_items"`
    Topping       string           `json:"topping"`
    TrackingOrder     string           `json:"tracking_order"`
    TakeAway       bool            `json:"take_away"`
    ChiliNumber    int64           `json:"chili_number"`
    TableToken     string          `json:"table_token"`
    OrderName      string          `json:"order_name"`  // New field
}

type CreateOrderRequestType struct {
    GuestID        int64           `json:"guest_id"`
    UserID         int64           `json:"user_id"`
    IsGuest        bool            `json:"is_guest"`
    TableNumber    int64           `json:"table_number"`
    OrderHandlerID int64           `json:"order_handler_id"`
    Status         string          `json:"status"`
    CreatedAt      time.Time       `json:"created_at"`
    UpdatedAt      time.Time       `json:"updated_at"`
    TotalPrice     int32           `json:"total_price"`
    DishItems      []OrderDish     `json:"dish_items"`
    SetItems       []OrderSet      `json:"set_items"`
    Topping       string           `json:"topping"`
    TrackingOrder     string           `json:"tracking_order"`
    TakeAway       bool            `json:"take_away"`
    ChiliNumber    int64           `json:"chili_number"`
    TableToken     string          `json:"table_token"`
    OrderName      string          `json:"order_name"`  // New field
}

type UpdateOrderRequestType struct {
    ID             int64           `json:"id"`
    GuestID        int64           `json:"guest_id"`
    UserID         int64           `json:"user_id"`
    TableNumber    int64           `json:"table_number"`
    OrderHandlerID int64           `json:"order_handler_id"`
    Status         string          `json:"status"`
    TotalPrice     int32           `json:"total_price"`
    DishItems      []OrderDish     `json:"dish_items"`
    SetItems       []OrderSet      `json:"set_items"`
    IsGuest        bool            `json:"is_guest"`
    Topping       string           `json:"topping"`
    TrackingOrder     string           `json:"tracking_order"`
    TakeAway       bool            `json:"take_away"`
    ChiliNumber    int64           `json:"chili_number"`
    TableToken     string          `json:"table_token"`
    OrderName      string          `json:"order_name"`  // New field
}

type OrderDetailedResponse struct {
    DataSet         []OrderSetDetailed    `json:"data_set"`
    DataDish        []OrderDetailedDish   `json:"data_dish"`
    ID              int64                 `json:"id"`
    GuestID         int64                 `json:"guest_id"`
    UserID          int64                 `json:"user_id"`
    TableNumber     int64                 `json:"table_number"`
    OrderHandlerID  int64                 `json:"order_handler_id"`
    Status          string                `json:"status"`
    TotalPrice      int32                 `json:"total_price"`
    IsGuest         bool                  `json:"is_guest"`
    Topping       string           `json:"topping"`
    TrackingOrder     string           `json:"tracking_order"`
    TakeAway        bool                  `json:"take_away"`
    ChiliNumber     int64                 `json:"chili_number"`
    TableToken      string                `json:"table_token"`
    OrderName       string                `json:"order_name"`  // New field
}
// new
// GetOrdersRequest struct
type GetOrdersRequestType struct {
    Page     int32 `json:"page"`
    PageSize int32 `json:"page_size"`
}

// PaginationInfo struct
type PaginationInfo struct {
    CurrentPage int32 `json:"current_page"`
    TotalPages  int32 `json:"total_pages"`
    TotalItems  int64 `json:"total_items"`
    PageSize    int32 `json:"page_size"`
}

// OrderListResponse struct
type OrderListResponse struct {
    Data       []OrderType    `json:"data"`
    Pagination PaginationInfo `json:"pagination"`
}

////-------

type OrderDish struct {
    DishID   int64 `json:"dish_id"`
    Quantity int64 `json:"quantity"`
}


type OrderSet struct {
    SetID   int64 `json:"set_id"`
    Quantity int64 `json:"quantity"`
}

// CreateOrderRequest struct
// GetOrdersRequest struct

// PayOrdersRequest struct
type PayOrdersRequestType struct {
	GuestID *int64 `json:"guest_id,omitempty"`
	UserID  *int64 `json:"user_id,omitempty"`
}

// OrderResponse struct
type OrderResponse struct {
	Data OrderType `json:"data"`
}



// OrderIDParam struct
type OrderIDParam struct {
	ID int64 `json:"id"`
}

// OrderDetailIDParam struct
type OrderDetailIDParam struct {
	ID int64 `json:"id"`
}

// Guest struct
type Guest struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	TableNumber int32     `json:"table_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}


type OrderDetailedDish struct {
    DishID      int64  `json:"dish_id"`
    Quantity    int64  `json:"quantity"`
    Name        string `json:"name"`
    Price       int32  `json:"price"`
    Description string `json:"description"`
    Image       string `json:"image"`
    Status      string `json:"status"`
}

type OrderSetDetailed struct {
    ID          int64            `json:"id"`
    Name        string           `json:"name"`
    Description string           `json:"description"`
    Dishes      []OrderDetailedDish `json:"dishes"`
    UserID      int32            `json:"userId"`
    CreatedAt   time.Time        `json:"created_at"`
    UpdatedAt   time.Time        `json:"updated_at"`
    IsFavourite bool             `json:"is_favourite"`
    LikeBy      []int64          `json:"like_by"`
    IsPublic    bool             `json:"is_public"`
    Image       string           `json:"image"`
    Price       int32            `json:"price"`
      Quantity       int64            `json:"quantity"`
}



type OrderDetailedListResponse struct {
     Data      []OrderDetailedResponse `json:"data"`
      Pagination PaginationInfo `json:"pagination"`
}


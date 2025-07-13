package delivery_grpc

import (
	"time"
)

type Delivery struct {
    ID                    int64     `json:"id" db:"id"`
    GuestID               int64     `json:"guest_id" db:"guest_id"`
    UserID                int64     `json:"user_id" db:"user_id"`
    IsGuest               bool      `json:"is_guest" db:"is_guest"`
    TableNumber           int64     `json:"table_number" db:"table_number"`
    OrderHandlerID        int64     `json:"order_handler_id" db:"order_handler_id"`
    Status                string    `json:"status" db:"status"`
    CreatedAt             time.Time `json:"created_at" db:"created_at"`
    UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
    TotalPrice            int32     `json:"total_price" db:"total_price"`
    DishItems             []DishDeliveryItem `json:"dish_items"`


	OrderID				int64     `json:"order_id" db:"order_id"`
    BowChili              int64     `json:"bow_chili" db:"bow_chili"`
    BowNoChili            int64     `json:"bow_no_chili" db:"bow_no_chili"`
    TakeAway              bool      `json:"take_away" db:"take_away"`
    ChiliNumber           int64     `json:"chili_number" db:"chili_number"`
    TableToken            string    `json:"table_token" db:"table_token"`
    ClientName            string    `json:"client_name" db:"client_name"`
    // Additional suggested fields
    DeliveryAddress       string    `json:"delivery_address" db:"delivery_address"`
    DeliveryContact      string    `json:"delivery_contact" db:"delivery_contact"`
    DeliveryNotes        string    `json:"delivery_notes" db:"delivery_notes"`
    ScheduledTime        time.Time `json:"scheduled_time" db:"scheduled_time"`
    DeliveryFee          int32     `json:"delivery_fee" db:"delivery_fee"`

    DeliveryStatus       string    `json:"delivery_status" db:"delivery_status"`
    EstimatedDeliveryTime time.Time `json:"estimated_delivery_time" db:"estimated_delivery_time"`
    ActualDeliveryTime   time.Time `json:"actual_delivery_time" db:"actual_delivery_time"`
}

type DeliveryDetailedDish struct {
    DishID      int64  `json:"dish_id" db:"dish_id"`
    Quantity    int64  `json:"quantity" db:"quantity"`
    Name        string `json:"name" db:"name"`
    Price       int32  `json:"price" db:"price"`
    Description string `json:"description" db:"description"`
    Image       string `json:"image" db:"image"`
    Status      string `json:"status" db:"status"`
}

type DishDeliveryItem struct {

    DishID     int64 `json:"dish_id" db:"dish_id"`
    Quantity   int64 `json:"quantity" db:"quantity"`
}

type CreateDeliveryRequest struct {
    GuestID               int64     `json:"guest_id"`
    UserID                int64     `json:"user_id"`
    IsGuest               bool      `json:"is_guest"`
    TableNumber           int64     `json:"table_number"`
    OrderHandlerID        int64     `json:"order_handler_id"`
    Status                string    `json:"status"`
    TotalPrice            int32     `json:"total_price"`
    DishItems             []DishDeliveryItem `json:"dish_items"`
    BowChili              int64     `json:"bow_chili"`
    BowNoChili            int64     `json:"bow_no_chili"`
    TakeAway              bool      `json:"take_away"`
    ChiliNumber           int64     `json:"chili_number"`
    TableToken            string    `json:"table_token"`
    ClientName            string    `json:"client_name"`
    DeliveryAddress       string    `json:"delivery_address"`
    DeliveryContact       string    `json:"delivery_contact"`
    DeliveryNotes         string    `json:"delivery_notes"`
    ScheduledTime         time.Time `json:"scheduled_time"`
    OrderID 				int64     `json:"order_id" db:"order_id"`

    DeliveryFee          int32     `json:"delivery_fee" db:"delivery_fee"`

    DeliveryStatus       string    `json:"delivery_status" db:"delivery_status"`
}

type UpdateDeliveryRequest struct {
    ID                    int64     `json:"id"`
    Status                string    `json:"status"`
    DeliveryStatus       string    `json:"delivery_status"`
    DriverID             int64     `json:"driver_id"`
    EstimatedDeliveryTime time.Time `json:"estimated_delivery_time"`
    ActualDeliveryTime   time.Time `json:"actual_delivery_time"`
    DeliveryNotes        string    `json:"delivery_notes"`
}

type DeliveryDetailedResponse struct {
    ID                    int64     `json:"id"`
    GuestID               int64     `json:"guest_id"`
    UserID                int64     `json:"user_id"`
    TableNumber           int64     `json:"table_number"`
    OrderHandlerID        int64     `json:"order_handler_id"`
    Status                string    `json:"status"`
    TotalPrice            int32     `json:"total_price"`
    DataDish              []DeliveryDetailedDish `json:"data_dish"`
    IsGuest               bool      `json:"is_guest"`
    BowChili              int64     `json:"bow_chili"`
    BowNoChili            int64     `json:"bow_no_chili"`
    TakeAway              bool      `json:"take_away"`
    ChiliNumber           int64     `json:"chili_number"`
    TableToken            string    `json:"table_token"`
    ClientName            string    `json:"client_name"`
    DeliveryStatus       string    `json:"delivery_status"`
    DriverID             int64     `json:"driver_id"`
    DeliveryAddress      string    `json:"delivery_address"`
    EstimatedDeliveryTime time.Time `json:"estimated_delivery_time"`
        DeliveryContact      string    `json:"delivery_contact" db:"delivery_contact"`
    DeliveryNotes        string    `json:"delivery_notes" db:"delivery_notes"`
}

type DeliveryDetailedListResponse struct {
    Data       []DeliveryDetailedResponse `json:"data"`
    Pagination PaginationInfo            `json:"pagination"`
}

type PaginationInfo struct {
    CurrentPage int32 `json:"current_page"`
    TotalPages  int32 `json:"total_pages"`
    TotalItems  int64 `json:"total_items"`
    PageSize    int32 `json:"page_size"`
}

type GetDeliveriesRequest struct {
    Page     int32 `json:"page"`
    PageSize int32 `json:"page_size"`
}

type DeliveryIDParam struct {
    ID int64 `json:"id"`
}

type DeliveryClientNameParam struct {
    Name string `json:"name"`
}

type DeliveryResponse struct {
    ID                    int64     `json:"id" db:"id"`
    GuestID               int64     `json:"guest_id" db:"guest_id"`
    UserID                int64     `json:"user_id" db:"user_id"`
    IsGuest               bool      `json:"is_guest" db:"is_guest"`
    TableNumber           int64     `json:"table_number" db:"table_number"`
    OrderHandlerID        int64     `json:"order_handler_id" db:"order_handler_id"`
    Status                string    `json:"status" db:"status"`
    CreatedAt             time.Time `json:"created_at" db:"created_at"`
    UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
    TotalPrice            int32     `json:"total_price" db:"total_price"`
    DishItems             []DishDeliveryItem `json:"dish_items"`


	OrderID				int64     `json:"order_id" db:"order_id"`
    BowChili              int64     `json:"bow_chili" db:"bow_chili"`
    BowNoChili            int64     `json:"bow_no_chili" db:"bow_no_chili"`
    TakeAway              bool      `json:"take_away" db:"take_away"`
    ChiliNumber           int64     `json:"chili_number" db:"chili_number"`
    TableToken            string    `json:"table_token" db:"table_token"`
    ClientName            string    `json:"client_name" db:"client_name"`
    // Additional suggested fields
    DeliveryAddress       string    `json:"delivery_address" db:"delivery_address"`
    DeliveryContact      string    `json:"delivery_contact" db:"delivery_contact"`
    DeliveryNotes        string    `json:"delivery_notes" db:"delivery_notes"`
    ScheduledTime        time.Time `json:"scheduled_time" db:"scheduled_time"`
    DeliveryFee          int32     `json:"delivery_fee" db:"delivery_fee"`

    DeliveryStatus       string    `json:"delivery_status" db:"delivery_status"`
    EstimatedDeliveryTime time.Time `json:"estimated_delivery_time" db:"estimated_delivery_time"`
    ActualDeliveryTime   time.Time `json:"actual_delivery_time" db:"actual_delivery_time"`
}

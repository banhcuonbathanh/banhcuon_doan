package tables_test

import (
	"time"
)

type TableStatus string

const (
	TableStatusAvailable    TableStatus = "AVAILABLE"
	TableStatusOccupied     TableStatus = "OCCUPIED"
	TableStatusReserved     TableStatus = "RESERVED"
	TableStatusOutOfService TableStatus = "OUT_OF_SERVICE"
	TableStatusTakeAway     TableStatus = "TAKE_AWAY"    // Added new status
)

type Table struct {
	Number    int32       `json:"number"`
	Capacity  int32       `json:"capacity"`
	Status    TableStatus `json:"status"`
	Token     string      `json:"token"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type CreateTableRequest struct {
	Number   int32       `json:"number"`
	Capacity int32       `json:"capacity"`
	Status   TableStatus `json:"status"`
}

type UpdateTableRequest struct {
	Number      int32       `json:"number"`
	ChangeToken bool        `json:"change_token"`
	Capacity    int32       `json:"capacity"`
	Status      TableStatus `json:"status"`
}

type TableResponse struct {
	Data    Table  `json:"data"`
	Message string `json:"message"`
}

type TableListResponse struct {
	Data    []Table `json:"data"`
	Message string  `json:"message"`
}

type TableNumberRequest struct {
	Number int32 `json:"number"`
}

// Existing Guest and Order types remain unchanged
type Guest struct {
	ID                    int64     `json:"id"`
	Name                  string    `json:"name"`
	TableNumber           int32     `json:"table_number"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type Order struct {
	ID              int64     `json:"id"`
	GuestID         int64     `json:"guest_id"`
	TableNumber     int32     `json:"table_number"`
	DishSnapshotID  int64     `json:"dish_snapshot_id"`
	Quantity        int32     `json:"quantity"`
	OrderHandlerID  int64     `json:"order_handler_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
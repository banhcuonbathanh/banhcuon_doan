package set_qr

import (
	"time"
)

// Go structs
type SetSnapshot struct {
    ID          int32       `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Dishes      []SetDish   `json:"dishes"`
    UserID      int32       `json:"userId"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
    SetID       int32       `json:"set_id"`
    IsPublic    bool        `json:"is_public"`
    Image       string      `json:"image"`
    Price       int32     `json:"price"`
}

type Set struct {
    ID          int64       `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Dishes      []SetDish   `json:"dishes"`
    UserID      int32       `json:"userId"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
    IsFavourite bool        `json:"is_favourite"`
    LikeBy      []int64     `json:"like_by"`
    IsPublic    bool        `json:"is_public"`
    Image       string      `json:"image"`
    Price       int32     `json:"price"`
}

type SetDish struct {
    DishID   int64 `json:"dish_id"`
    Quantity int64 `json:"quantity"`
}

type CreateSetRequest struct {
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Dishes      []SetDish `json:"dishes"`
    UserID      int32     `json:"userId"`
    IsPublic    bool      `json:"is_public"`
    Image       string    `json:"image"`
    Price       int32   `json:"price"`
}

type UpdateSetRequest struct {
    ID          int32     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Dishes      []SetDish `json:"dishes"`
    IsPublic    bool      `json:"is_public"`
    Image       string    `json:"image"`
    Price       int32   `json:"price"`
}

type SetResponse struct {
    Data Set `json:"data"`
}

type SetListResponse struct {
    Data []Set `json:"data"`
}

type SetIDParam struct {
    ID int32 `json:"id"`
}

// ----------

type SetDetailedDish struct {
    DishID      int64  `json:"dish_id"`
    Quantity    int64  `json:"quantity"`
    Name        string `json:"name"`
    Price       int32  `json:"price"`
    Description string `json:"description"`
    Image       string `json:"image"`
    Status      string `json:"status"`
}

type SetDetailed struct {
    ID          int64            `json:"id"`
    Name        string           `json:"name"`
    Description string           `json:"description"`
    Dishes      []SetDetailedDish `json:"dishes"`
    UserID      int32            `json:"userId"`
    CreatedAt   time.Time        `json:"created_at"`
    UpdatedAt   time.Time        `json:"updated_at"`
    IsFavourite bool             `json:"is_favourite"`
    LikeBy      []int64          `json:"like_by"`
    IsPublic    bool             `json:"is_public"`
    Image       string           `json:"image"`
    Price       int32            `json:"price"`
}

type SetDetailedListResponse struct {
    Data []SetDetailed `json:"data"`
}

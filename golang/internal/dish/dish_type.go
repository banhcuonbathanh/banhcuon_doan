package dish_grpc

import (
	"time"
)

type Dish struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Price       int32     `json:"price"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateDishRequest struct {
	Name        string `json:"name"`
	Price       int32  `json:"price"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Status      string `json:"status,omitempty"`
}

type UpdateDishRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Price       int32  `json:"price"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Status      string `json:"status,omitempty"`
}

type DishResponse struct {
	Data    Dish   `json:"data"`
	Message string `json:"message"`
}

type DishListResponse struct {
	Data    []Dish `json:"data"`
	Message string `json:"message"`
}
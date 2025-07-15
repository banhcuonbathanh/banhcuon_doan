package model

import (
	"time"
)

type ProductReq struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Image        string  `json:"image"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	Rating       int64   `json:"rating"`
	NumReviews   int64   `json:"num_reviews"`
	Price        float32 `json:"price"`
	CountInStock int64   `json:"count_in_stock"`
}

type ProductRes struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Image        string     `json:"image"`
	Category     string     `json:"category"`
	Description  string     `json:"description"`
	Rating       int64      `json:"rating"`
	NumReviews   int64      `json:"num_reviews"`
	Price        float32    `json:"price"`
	CountInStock int64      `json:"count_in_stock"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type OrderReq struct {
	Items         []*OrderItem `json:"items"`
	PaymentMethod string       `json:"payment_method"`
	TaxPrice      float32      `json:"tax_price"`
	ShippingPrice float32      `json:"shipping_price"`
	TotalPrice    float32      `json:"total_price"`
}

type OrderItem struct {
	Name      string  `json:"name"`
	Quantity  int64   `json:"quantity"`
	Image     string  `json:"image"`
	Price     float32 `json:"price"`
	ProductID int64   `json:"product_id"`
}

type OrderRes struct {
	ID            int64        `json:"id"`
	Items         []*OrderItem `json:"items"`
	PaymentMethod string       `json:"payment_method"`
	TaxPrice      float32      `json:"tax_price"`
	ShippingPrice float32      `json:"shipping_price"`
	TotalPrice    float32      `json:"total_price"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     *time.Time   `json:"updated_at"`
}

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Password     string    `json:"password"`
	Role         string    `json:"role"` // Changed from IsAdmin bool to Role string
	Phone        string    `json:"phone"`
	Image        string    `json:"image"`
	Address      string    `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	FavoriteFood []string  `json:"favorite_food"` // Added field for favorite food
}

type AccountInput struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Role      string    `json:"role"` // Changed from IsAdmin bool to Role string
	Phone     string    `json:"phone"`
	Image     string    `json:"image"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserResModel struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"` // Changed from IsAdmin bool to Role string
	Phone        string    `json:"phone"`
	Image        string    `json:"image"`
	Address      string    `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	FavoriteFood []string  `json:"favorite_food"` // Added field for favorite food
}

type RenewAccessTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RenewAccessTokenRes struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

type Session struct {
	ID           string
	UserEmail    string
	RefreshToken string
	IsRevoked    bool
	ExpiresAt    time.Time
}

type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleOwner    Role = "OWNER"
	RoleEmployee Role = "EMPLOYEE"
)

// NEW
type Account struct {
	ID        int64
	BranchID  int64
	Name      string
	Email     string
	Password  string
	Avatar    string
	Title     string
	Role      Role
	OwnerID   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RegisterUserReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRes struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	User         AccountLoginResponse `json:"user"`
}

type AccountLoginResponse struct {
	ID       int64  `json:"id"`
	BranchID int64  `json:"branch_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	OwnerID  int64  `json:"owner_id"`
}

// type CreateUserRequest struct {
// 	BranchID int64
// 	Name     string
// 	Email    string
// 	Password string
// 	Avatar   string
// 	Title    string
// 	Role     string
// 	OwnerID  int64
// }
type CreateUserRequest struct {
	BranchID int64  `json:"branch_id" validate:"required,gt=0"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Avatar   string `json:"avatar" validate:"omitempty,url"`
	Title    string `json:"title" validate:"required,min=2,max=100"`
	Role     string `json:"role" validate:"required,oneof=admin user manager"`
	OwnerID  int64  `json:"owner_id" validate:"required,gt=0"`
}
type UpdateUserRequest struct {
	ID       int64  `json:"id"`
	BranchID int64  `json:"branch_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	OwnerID  int64  `json:"owner_id"`
}

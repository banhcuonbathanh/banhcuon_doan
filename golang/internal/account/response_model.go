package account

import "time"

type FindByEmailResponse struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Title     string    `json:"title"`
	Role      string    `json:"role"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status bool   `json:"status"`
}

type CreateUserResponse struct {
	BranchID int64  `json:"branch_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	OwnerID  int64  `json:"owner_id"`
}

type FindAccountByIDResponse struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Title     string    `json:"title"`
	Role      string    `json:"role"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateUserResponse struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Title     string    `json:"title"`
	Role      string    `json:"role"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeleteUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

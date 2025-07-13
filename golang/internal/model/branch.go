package model

import (
	"time"
)

// Branch represents a restaurant branch
type Branch struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone,omitempty"`
	ManagerID int64     `json:"manager_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateBranchResponse represents the response when creating a branch
type CreateBranchResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone,omitempty"`
	ManagerID int64     `json:"manager_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetBranchResponse represents the response when getting a branch
type GetBranchResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone,omitempty"`
	ManagerID int64     `json:"manager_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListBranchesResponse represents the response when listing branches
type ListBranchesResponse struct {
	Branches []GetBranchResponse `json:"branches"`
	Total    int64               `json:"total"`
}

// DeleteBranchResponse represents the response when deleting a branch
type DeleteBranchResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

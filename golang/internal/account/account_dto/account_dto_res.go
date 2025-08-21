
// internal/account/account_dto/account_dto_res.go
package account_dto



import "time"



type UserSummary struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	BranchID int64  `json:"branch_id"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type UserTokenInfo struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	BranchID  int64     `json:"branch_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ===== AUTHENTICATION RESPONSES =====
type RegisterUserResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    time.Time   `json:"expires_at"`
	User         UserProfile `json:"user"`
	Success      bool        `json:"success"`
}

type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ===== USER MANAGEMENT RESPONSES =====
type UserResponse struct {
	User UserProfile `json:"user"`
}

// type UserProfileResponse struct {
// 	User UserProfile `json:"user"`
// }

// type UpdateUserResponse struct {
// 	User    UserProfile `json:"user"`
// 	Success bool        `json:"success"`
// 	Message string      `json:"message,omitempty"`
// }

// type DeleteUserResponse struct {
// 	Success bool   `json:"success"`
// 	Message string `json:"message"`
// }

// type UsersListResponse struct {
// 	Users []UserProfile `json:"users"`
// 	Total int64         `json:"total"`
// }

// ===== PASSWORD MANAGEMENT RESPONSES =====
type ChangePasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ForgotPasswordResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ResetToken string `json:"reset_token,omitempty"` // Optional: for testing purposes
}

type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ===== ACCOUNT VERIFICATION RESPONSES =====
type VerifyEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ResendVerificationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateAccountStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ===== SEARCH AND FILTERING RESPONSES =====
// type SearchUsersResponse struct {
// 	Users      []UserSummary    `json:"users"`
// 	TotalCount int64            `json:"total_count"`
// 	Page       int32            `json:"page"`
// 	PageSize   int32            `json:"page_size"`
// 	TotalPages int32            `json:"total_pages"`
// 	Pagination PaginationInfo   `json:"pagination"`
// }

// type PaginationInfo struct {
// 	CurrentPage  int32 `json:"current_page"`
// 	TotalPages   int32 `json:"total_pages"`
// 	TotalItems   int64 `json:"total_items"`
// 	ItemsPerPage int32 `json:"items_per_page"`
// 	HasNext      bool  `json:"has_next"`
// 	HasPrevious  bool  `json:"has_previous"`
// }

type FindByRoleResponse struct {
	Users []UserProfile `json:"users"`
	Role  string        `json:"role"`
	Total int64         `json:"total"`
}

type FindByBranchResponse struct {
	Users    []UserProfile `json:"users"`
	BranchID int64         `json:"branch_id"`
	Total    int64         `json:"total"`
}

// ===== TOKEN MANAGEMENT RESPONSES =====
type RefreshTokenResponse struct {
	Success      bool      `json:"success"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type ValidateTokenResponse struct {
	Valid     bool          `json:"valid"`
	UserID    int64         `json:"user_id"`
	Message   string        `json:"message"`
	ExpiresAt time.Time     `json:"expires_at"`
	TokenInfo UserTokenInfo `json:"token_info,omitempty"`
}

// ===== BULK OPERATIONS RESPONSES =====
type BulkDeleteUsersResponse struct {
	Success       bool    `json:"success"`
	DeletedCount  int64   `json:"deleted_count"`
	FailedUserIDs []int64 `json:"failed_user_ids,omitempty"`
	Message       string  `json:"message"`
}

type BulkUpdateStatusResponse struct {
	Success       bool    `json:"success"`
	UpdatedCount  int64   `json:"updated_count"`
	FailedUserIDs []int64 `json:"failed_user_ids,omitempty"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
}

// ===== STATISTICS RESPONSES =====
type UserStatisticsResponse struct {
	TotalUsers     int64                    `json:"total_users"`
	ActiveUsers    int64                    `json:"active_users"`
	InactiveUsers  int64                    `json:"inactive_users"`
	UsersByRole    map[string]int64         `json:"users_by_role"`
	UsersByBranch  map[string]int64         `json:"users_by_branch"`
	UsersByStatus  map[string]int64         `json:"users_by_status"`
	RecentSignups  int64                    `json:"recent_signups"`
	LastUpdated    time.Time                `json:"last_updated"`
}

// ===== ERROR RESPONSES =====
// type ErrorResponse struct {
// 	Success   bool   `json:"success"`
// 	Error     string `json:"error"`
// 	Message   string `json:"message"`
// 	Code      string `json:"code,omitempty"`
// 	Timestamp time.Time `json:"timestamp"`
// }

type ValidationErrorResponse struct {
	Success bool                    `json:"success"`
	Error   string                  `json:"error"`
	Details map[string][]string     `json:"details"` // Field -> list of errors
}

// ===== GENERIC API RESPONSES =====
// type APIResponse struct {
// 	Success   bool        `json:"success"`
// 	Message   string      `json:"message,omitempty"`
// 	Data      interface{} `json:"data,omitempty"`
// 	Error     string      `json:"error,omitempty"`
// 	Timestamp time.Time   `json:"timestamp"`
// }

type ListResponse struct {
	Items      interface{}    `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
	Total      int64          `json:"total"`
}

// ===== HEALTH CHECK RESPONSES =====
type HealthCheckResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

// type RegisterResponse struct {
// 	ID     int64  `json:"id"`
// 	Name   string `json:"name"`
// 	Email  string `json:"email"`
// 	Status bool   `json:"status"`
// }

// type CreateUserResponse struct {
// 	ID       int64  `json:"id"`
// 	Name     string `json:"name"`
// 	Email    string `json:"email"`
// 	Avatar   string `json:"avatar"`
// 	Title    string `json:"title"`
// 	Role     string `json:"role"`
// 	BranchID int64  `json:"branch_id"`
// 	Status   string `json:"status"`
// 	Created  bool   `json:"created"`
// 	OwnerID int64 `json:"owner_id"`
// }


// type FindAccountByIDResponse struct {
// 	ID        int64     `json:"id"`
// 	Name      string    `json:"name"`
// 	Email     string    `json:"email"`
// 	Avatar    string    `json:"avatar"`
// 	Title     string    `json:"title"`
// 	Role      string    `json:"role"`
// 	BranchID  int64     `json:"branch_id"`
// 	Status    string    `json:"status"`
// 	Created   bool      `json:"created"`
// 	OwnerID   int64     `json:"owner_id"`
// 	CreatedAt time.Time `json:"created_at"`
// 	UpdatedAt time.Time `json:"updated_at"`
// }

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

// RegisterResponse represents the registration response
type RegisterResponse struct {
	ID     int64  `json:"id" example:"123"`
	Name   string `json:"name" example:"John Doe"`
	Email  string `json:"email" example:"john.doe@example.com"`
	Status bool   `json:"status" example:"true"`
}

// CreateUserResponse represents the user creation response
type CreateUserResponse struct {
	BranchID int64  `json:"branch_id" example:"1"`
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"john.doe@example.com"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Title    string `json:"title" example:"Manager"`
	Role     string `json:"role" example:"admin"`
	OwnerID  int64  `json:"owner_id" example:"1"`
}

// FindAccountByIDResponse represents the find account by ID response
type FindAccountByIDResponse struct {
	ID        int64     `json:"id" example:"123"`
	BranchID  int64     `json:"branch_id" example:"1"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Avatar    string    `json:"avatar" example:"https://example.com/avatar.jpg"`
	Title     string    `json:"title" example:"Manager"`
	Role      string    `json:"role" example:"admin"`
	OwnerID   int64     `json:"owner_id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// UserProfile represents a user profile
type UserProfile struct {
	ID        int64     `json:"id" example:"123"`
	BranchID  int64     `json:"branch_id" example:"1"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Avatar    string    `json:"avatar" example:"https://example.com/avatar.jpg"`
	Title     string    `json:"title" example:"Manager"`
	Role      string    `json:"role" example:"admin"`
	OwnerID   int64     `json:"owner_id" example:"1"`
	CreatedAt time.Time `json:"created_at,omitempty" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at,omitempty" example:"2023-01-01T00:00:00Z"`
}

// UserProfileResponse represents the user profile response
type UserProfileResponse struct {
	User UserProfile `json:"user"`
}

// UpdateUserResponse represents the user update response
type UpdateUserResponse struct {
	User    UserProfile `json:"user"`
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"User updated successfully"`
}

// DeleteUserResponse represents the user deletion response
type DeleteUserResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"User deleted successfully"`
}

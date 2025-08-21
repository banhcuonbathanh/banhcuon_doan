// internal/account/account_dto/account_dto_auth.go
package account_dto

import "time"

// // LoginRequest represents the login request payload
// // swagger:model LoginRequest
// type LoginUserRes struct {
// 	AccessToken  string                `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
// 	RefreshToken string                `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
// 	User         AccountLoginResponse  `json:"user"`
// }

// // AccountLoginResponse represents the user data in login response
// type AccountLoginResponse struct {
// 	ID       int64  `json:"id" example:"123"`
// 	BranchID int64  `json:"branch_id" example:"1"`
// 	Name     string `json:"name" example:"John Doe"`
// 	Email    string `json:"email" example:"john.doe@example.com"`
// 	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
// 	Title    string `json:"title" example:"Manager"`
// 	Role     string `json:"role" example:"admin"`
// 	OwnerID  int64  `json:"owner_id" example:"1"`
// }

// Account represents a user account
type Account struct {
	ID        int64     `json:"id" example:"123"`
	BranchID  int64     `json:"branch_id" example:"1"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Avatar    string    `json:"avatar" example:"https://example.com/avatar.jpg"`
	Title     string    `json:"title" example:"Manager"`
	Role      Role      `json:"role" example:"admin"`
	OwnerID   int64     `json:"owner_id" example:"1"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// Role represents user roles
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleUser    Role = "user"
)

// AccountStatus represents account status
type AccountStatus string

const (
	StatusActive    AccountStatus = "active"
	StatusInactive  AccountStatus = "inactive"
	StatusSuspended AccountStatus = "suspended"
	StatusPending   AccountStatus = "pending"
)

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int32 `json:"page" example:"1"`
	PageSize   int32 `json:"page_size" example:"10"`
	TotalCount int64 `json:"total_count" example:"100"`
	TotalPages int32 `json:"total_pages" example:"10"`
	HasNext    bool  `json:"has_next" example:"true"`
	HasPrev    bool  `json:"has_prev" example:"false"`
}

// SortInfo represents sorting information
type SortInfo struct {
	SortBy    string `json:"sort_by" example:"created_at"`
	SortOrder string `json:"sort_order" example:"desc"`
}

// SearchFilters represents search filter options
type SearchFilters struct {
	Query        string   `json:"query,omitempty" example:"john"`
	Role         string   `json:"role,omitempty" example:"admin"`
	BranchID     int64    `json:"branch_id,omitempty" example:"1"`
	StatusFilter []string `json:"status_filter,omitempty" example:"active,pending"`
}

// APIResponse represents a generic API response wrapper
type APIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty" example:""`
}

// ValidationError represents field validation errors
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Email is required"`
	Tag     string `json:"tag" example:"required"`
	Value   string `json:"value" example:""`
}

// ErrorResponse represents detailed error response
type ErrorResponse struct {
	Error       string            `json:"error" example:"validation_error"`
	Message     string            `json:"message" example:"Validation failed"`
	Code        int               `json:"code" example:"400"`
	Details     map[string]string `json:"details,omitempty"`
	Validations []ValidationError `json:"validations,omitempty"`
	Timestamp   time.Time         `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   int64  `json:"user_id" example:"123"`
	Email    string `json:"email" example:"user@example.com"`
	Role     string `json:"role" example:"admin"`
	BranchID int64  `json:"branch_id" example:"1"`
}

// RefreshTokenClaims represents refresh token claims
type RefreshTokenClaims struct {
	UserID int64  `json:"user_id" example:"123"`
	Email  string `json:"email" example:"user@example.com"`
}

// PasswordRequirements represents password validation requirements
type PasswordRequirements struct {
	MinLength        int  `json:"min_length" example:"8"`
	RequireUppercase bool `json:"require_uppercase" example:"true"`
	RequireLowercase bool `json:"require_lowercase" example:"true"`
	RequireNumbers   bool `json:"require_numbers" example:"true"`
	RequireSpecial   bool `json:"require_special" example:"true"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers    int64 `json:"total_users" example:"1000"`
	ActiveUsers   int64 `json:"active_users" example:"800"`
	InactiveUsers int64 `json:"inactive_users" example:"150"`
	PendingUsers  int64 `json:"pending_users" example:"50"`
}

// BranchStats represents branch-specific user statistics
type BranchStats struct {
	BranchID    int64     `json:"branch_id" example:"1"`
	BranchName  string    `json:"branch_name" example:"Main Branch"`
	UserCount   int64     `json:"user_count" example:"25"`
	UserStats   UserStats `json:"user_stats"`
	LastUpdated time.Time `json:"last_updated" example:"2023-01-01T00:00:00Z"`
}

// ActivityLog represents user activity logging
type ActivityLog struct {
	ID        int64     `json:"id" example:"1"`
	UserID    int64     `json:"user_id" example:"123"`
	Action    string    `json:"action" example:"login"`
	Resource  string    `json:"resource" example:"account"`
	IPAddress string    `json:"ip_address" example:"192.168.1.100"`
	UserAgent string    `json:"user_agent" example:"Mozilla/5.0..."`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

// NotificationSettings represents user notification preferences
type NotificationSettings struct {
	EmailNotifications bool `json:"email_notifications" example:"true"`
	SMSNotifications   bool `json:"sms_notifications" example:"false"`
	PushNotifications  bool `json:"push_notifications" example:"true"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Language      string               `json:"language" example:"en"`
	Timezone      string               `json:"timezone" example:"UTC"`
	DateFormat    string               `json:"date_format" example:"YYYY-MM-DD"`
	Theme         string               `json:"theme" example:"light"`
	Notifications NotificationSettings `json:"notifications"`
}

// ExtendedUserProfile represents extended user profile with preferences
type ExtendedUserProfile struct {
	Account     Account         `json:"account"`
	Preferences UserPreferences `json:"preferences"`
	LastLogin   *time.Time      `json:"last_login,omitempty" example:"2023-01-01T00:00:00Z"`
	LoginCount  int64           `json:"login_count" example:"42"`
}

// BulkOperation represents bulk operation request
type BulkOperation struct {
	Operation string  `json:"operation" example:"update_status"`
	UserIDs   []int64 `json:"user_ids" example:"[1,2,3,4,5]"`
	Data      map[string]interface{} `json:"data"`
}

// BulkOperationResult represents bulk operation response
type BulkOperationResult struct {
	Success     []int64 `json:"success" example:"[1,2,3]"`
	Failed      []int64 `json:"failed" example:"[4,5]"`
	Errors      map[string]string `json:"errors,omitempty"`
	TotalCount  int     `json:"total_count" example:"5"`
	SuccessCount int    `json:"success_count" example:"3"`
	FailureCount int    `json:"failure_count" example:"2"`
}

// ExportRequest represents data export request
type ExportRequest struct {
	Format    string            `json:"format" example:"csv"`
	Filters   SearchFilters     `json:"filters"`
	Fields    []string          `json:"fields" example:"[\"name\",\"email\",\"role\"]"`
	SortBy    string            `json:"sort_by" example:"created_at"`
	SortOrder string            `json:"sort_order" example:"desc"`
}

// ImportRequest represents data import request
type ImportRequest struct {
	Format    string                   `json:"format" example:"csv"`
	Data      []map[string]interface{} `json:"data"`
	Options   map[string]interface{}   `json:"options"`
	Overwrite bool                     `json:"overwrite" example:"false"`
}

// ImportResult represents data import response
type ImportResult struct {
	TotalRecords    int                    `json:"total_records" example:"100"`
	SuccessCount    int                    `json:"success_count" example:"95"`
	FailureCount    int                    `json:"failure_count" example:"5"`
	Errors          []map[string]string    `json:"errors,omitempty"`
	CreatedRecords  []int64                `json:"created_records,omitempty"`
	UpdatedRecords  []int64                `json:"updated_records,omitempty"`
	SkippedRecords  []int64                `json:"skipped_records,omitempty"`
}

// HealthCheck represents system health status
type HealthCheck struct {
	Status    string            `json:"status" example:"healthy"`
	Timestamp time.Time         `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version" example:"1.0.0"`
	Uptime    string            `json:"uptime" example:"72h30m"`
}

// // RegisterUserRequest represents the user registration request payload
// // swagger:model RegisterUserRequest
// type RegisterUserRequest struct {
// 	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
// 	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
// 	Password string `json:"password" validate:"required" example:"SecurePass123!"`
// }

// LoginUserRes represents the login response
// swagger:model LoginUserRes
type LoginUserRes struct {
	AccessToken  string                `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string                `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         AccountLoginResponse  `json:"user"`
}

// AccountLoginResponse represents the user data in login response
// swagger:model AccountLoginResponse
type AccountLoginResponse struct {
	ID       int64  `json:"id" example:"123"`
	BranchID int64  `json:"branch_id" example:"1"`
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"john.doe@example.com"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Title    string `json:"title" example:"Manager"`
	Role     string `json:"role" example:"admin"`
	OwnerID  int64  `json:"owner_id" example:"1"`
}

// internal/account/account_dto/account_dto_req.go
package account_dto

// LoginRequest represents the login request payload
// swagger:model account_dto.LoginRequest
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// RegisterUserRequest represents the user registration request payload
// swagger:model account_dto.RegisterUserRequest
type RegisterUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	Email    string `json:"email" validate:"required,email,uniqueemail" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,password" example:"SecurePass123!"`
}
type CreateUserRequest struct {
	BranchID int64  `json:"branch_id"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,strongpassword"`
	Avatar   string `json:"avatar,omitempty" validate:"omitempty,url"`
	Title    string `json:"title,omitempty" validate:"omitempty,max=200"`
	Role     string `json:"role" validate:"required,userrole"`
	OwnerID  int64  `json:"owner_id,omitempty"`
}
// CreateUserRequest represents the user creation request payload


// UpdateUserRequest represents the user update request payload
type UpdateUserRequest struct {
	BranchID int64  `json:"branch_id,omitempty" example:"1"`
	Name     string `json:"name,omitempty" validate:"omitempty,min=2,max=100" example:"John Doe"`
	Email    string `json:"email,omitempty" validate:"omitempty,email" example:"john.doe@example.com"`
	Avatar   string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	Title    string `json:"title,omitempty" example:"Senior Manager"`
	Role     string `json:"role,omitempty" validate:"omitempty,role" example:"manager"`
	OwnerID  int64  `json:"owner_id,omitempty" example:"1"`
}


// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	  UserID          int64  `json:"user_id"`   
	CurrentPassword string `json:"current_password" validate:"required" example:"oldPassword123"`
	NewPassword     string `json:"new_password" validate:"required,password" example:"newPassword123!"`
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

// ResetPasswordRequest represents the reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required" example:"reset_token_here"`
	NewPassword string `json:"new_password" validate:"required,password" example:"newPassword123!"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"refresh_token_here"`
}

// ResendVerificationRequest represents the resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

// UpdateAccountStatusRequest represents the update account status request
type UpdateAccountStatusRequest struct {
	UserID     int    `json:"id" validate:"required,min=1"`
	Status string `json:"status" validate:"required,oneof=active inactive suspended pending" example:"active"`
}

// SearchUsersRequest represents the search users query parameters
type SearchUsersRequest struct {
	Query     string `json:"q,omitempty" example:"john"`
	Role      string `json:"role,omitempty" example:"admin"`
	BranchID  int64  `json:"branch_id,omitempty" example:"1"`
	Status    string `json:"status,omitempty" example:"active,pending"`
	Page      int32  `json:"page,omitempty" example:"1"`
	PageSize  int32  `json:"page_size,omitempty" example:"10"`
	SortBy    string `json:"sort_by,omitempty" example:"created_at"`
	SortOrder string `json:"sort_order,omitempty" example:"desc"`
}

// PaginationResponse represents pagination information
type PaginationResponse struct {
	Page       int32 `json:"page" example:"1"`
	PageSize   int32 `json:"page_size" example:"10"`
	TotalCount int64 `json:"total_count" example:"100"`
	TotalPages int32 `json:"total_pages" example:"10"`
	HasNext    bool  `json:"has_next" example:"true"`
	HasPrev    bool  `json:"has_prev" example:"false"`
}

// UsersListResponse represents a paginated list of users
type UsersListResponse struct {
	Users      []UserProfile      `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}

// SearchUsersResponse represents search results with pagination
type SearchUsersResponse struct {
	Users      []UserProfile      `json:"users"`
	TotalCount int64              `json:"total_count" example:"50"`
	Page       int32              `json:"page" example:"1"`
	PageSize   int32              `json:"page_size" example:"10"`
	TotalPages int32              `json:"total_pages" example:"5"`
	Pagination PaginationResponse `json:"pagination"`
}

// APIErrorResponse represents an API error response
type APIErrorResponse struct {
	Error   string                 `json:"error" example:"validation_error"`
	Message string                 `json:"message" example:"Validation failed"`
	Code    int                    `json:"code" example:"400"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// TokenResponse represents token-related responses
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt    int64  `json:"expires_at,omitempty" example:"1640995200"`
}

// TokenValidationResponse represents token validation response
type TokenValidationResponse struct {
	Valid     bool   `json:"valid" example:"true"`
	ExpiresAt int64  `json:"expires_at" example:"1640995200"`
	Message   string `json:"message" example:"Token is valid"`
	UserID    int64  `json:"id" example:"123"`
}

// new

type LogoutRequest struct {
    UserID int    `json:"user_id" validate:"required"`
    Token  string `json:"token" validate:"required"`
}

type ValidateTokenRequest struct {
    Token string `json:"token" validate:"required"`
}

type VerifyEmailRequest struct {
	VerificationToken string `json:"verification_token" validate:"required"`
}


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

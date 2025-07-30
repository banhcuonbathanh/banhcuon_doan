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

// CreateUserRequest represents the user creation request payload
type CreateUserRequest struct {
	 ID       int64  `json:"id"`  
	BranchID int64  `json:"branch_id" validate:"required" example:"1"`
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	Email    string `json:"email" validate:"required,email,uniqueemail" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,password" example:"SecurePass123!"`
	Avatar   string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	Title    string `json:"title,omitempty" example:"Manager"`
	Role     string `json:"role" validate:"required,role" example:"admin"`
	OwnerID  int64  `json:"owner_id,omitempty" example:"1"`
}

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
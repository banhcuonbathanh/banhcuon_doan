package account_dto

// ===== AUTHENTICATION REQUESTS =====

type CreateUserRequest struct {
	ID       int64  `json:"id" validate:"required,gt=0"`
	Name     string `json:"name" validate:"omitempty,min=2,max=100"`
	Email    string `json:"email" validate:"omitempty,email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	BranchID int64  `json:"branch_id" validate:"omitempty,gt=0"`
	Password string `json:"password" validate:"required"`
	OwnerID int64 `json:"owner_id" validate:"required,gt=0"`
}

type RegisterUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	BranchID int64  `json:"branch_id" validate:"required,gt=0"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ===== AUTHENTICATION RESPONSES =====

type RegisterResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status bool   `json:"status"`
}

type LoginResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	BranchID     int64  `json:"branch_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type CreateUserResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	BranchID int64  `json:"branch_id"`
	Status   string `json:"status"`
	Created  bool   `json:"created"`
    OwnerID int64 `json:"owner_id"`
}

// ===== USER MANAGEMENT REQUESTS =====
type UpdateUserRequest struct {
	ID       int64  `json:"id" validate:"required,gt=0"`
	Name     string `json:"name" validate:"omitempty,min=2,max=100"`
	Email    string `json:"email" validate:"omitempty,email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	BranchID int64  `json:"branch_id" validate:"omitempty,gt=0"`
	OwnerID int64 `json:"owner_id" validate:"required,gt=0"`
}

type DeleteUserRequest struct {
	UserID int64 `json:"user_id" validate:"required,gt=0"`
}

// ===== USER MANAGEMENT RESPONSES =====

type UpdateUserResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	BranchID int64  `json:"branch_id"`
	Updated  bool   `json:"updated"`
}

type DeleteUserResponse struct {
	UserID  int64 `json:"user_id"`
	Deleted bool  `json:"deleted"`
	Message string `json:"message"`
}

// ===== PASSWORD MANAGEMENT REQUESTS =====
type ChangePasswordRequest struct {
	UserID          int64  `json:"user_id" validate:"required,gt=0"`
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ===== PASSWORD MANAGEMENT RESPONSES =====

type ChangePasswordResponse struct {
	UserID  int64  `json:"user_id"`
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

type ForgotPasswordResponse struct {
	Email   string `json:"email"`
	Sent    bool   `json:"sent"`
	Message string `json:"message"`
}

type ResetPasswordResponse struct {
	Reset   bool   `json:"reset"`
	Message string `json:"message"`
}

// ===== ACCOUNT VERIFICATION REQUESTS =====
type VerifyEmailRequest struct {
	VerificationToken string `json:"verification_token" validate:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdateAccountStatusRequest struct {
	UserID int64  `json:"user_id" validate:"required,gt=0"`
	Status string `json:"status" validate:"required,oneof=active inactive suspended pending"`
}

// ===== ACCOUNT VERIFICATION RESPONSES =====

type VerifyEmailResponse struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

type ResendVerificationResponse struct {
	Email   string `json:"email"`
	Sent    bool   `json:"sent"`
	Message string `json:"message"`
}

type UpdateAccountStatusResponse struct {
	UserID  int64  `json:"user_id"`
	Status  string `json:"status"`
	Updated bool   `json:"updated"`
}

// ===== SEARCH AND FILTERING REQUESTS =====
type FindByRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

type FindByBranchRequest struct {
	BranchID int64 `json:"branch_id" validate:"required,gt=0"`
}

type FindByEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type FindByIDRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

type SearchUsersRequest struct {
	Query    string `json:"query" validate:"omitempty,min=1"`
	Role     string `json:"role,omitempty"`
	BranchID int64  `json:"branch_id,omitempty"`
	Page     int32  `json:"page" validate:"min=1"`
	PageSize int32  `json:"page_size" validate:"min=1,max=100"`
}

// ===== SEARCH AND FILTERING RESPONSES =====

type UserProfile struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Title     string `json:"title"`
	Role      string `json:"role"`
	BranchID  int64  `json:"branch_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type SearchUsersResponse struct {
	Users      []UserProfile `json:"users"`
	Total      int64         `json:"total"`
	Page       int32         `json:"page"`
	PageSize   int32         `json:"page_size"`
	TotalPages int32         `json:"total_pages"`
}

type FindUsersResponse struct {
	Users []UserProfile `json:"users"`
	Count int64         `json:"count"`
}

// ===== TOKEN MANAGEMENT REQUESTS =====
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type LogoutRequest struct {
	UserID int64  `json:"user_id" validate:"required,gt=0"`
	Token  string `json:"token,omitempty"` // Optional: if using token-based authentication
}

// ===== TOKEN MANAGEMENT RESPONSES =====

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID int64  `json:"user_id,omitempty"`
	Role   string `json:"role,omitempty"`
}

type LogoutResponse struct {
	UserID  int64  `json:"user_id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ===== PROFILE REQUESTS =====
type GetUserProfileRequest struct {
	UserID int64 `json:"user_id" validate:"required,gt=0"`
}

type GetUsersByBranchRequest struct {
	BranchID int64 `json:"branch_id" validate:"required,gt=0"`
}

type GetUsersByOwnerRequest struct {
	OwnerID int64 `json:"owner_id" validate:"required,gt=0"`
}

// ===== PROFILE RESPONSES =====

type GetUserProfileResponse struct {
	User UserProfile `json:"user"`
}

type GetUsersByBranchResponse struct {
	Users    []UserProfile `json:"users"`
	BranchID int64         `json:"branch_id"`
	Count    int64         `json:"count"`
}

type GetUsersByOwnerResponse struct {
	Users   []UserProfile `json:"users"`
	OwnerID int64         `json:"owner_id"`
	Count   int64         `json:"count"`
}

// ===== BULK OPERATIONS REQUESTS =====
type BulkDeleteUsersRequest struct {
	UserIDs []int64 `json:"user_ids" validate:"required,min=1,dive,gt=0"`
}

type BulkUpdateStatusRequest struct {
	UserIDs []int64 `json:"user_ids" validate:"required,min=1,dive,gt=0"`
	Status  string  `json:"status" validate:"required,oneof=active inactive suspended pending"`
}

// ===== BULK OPERATIONS RESPONSES =====

type BulkDeleteUsersResponse struct {
	DeletedIDs []int64 `json:"deleted_ids"`
	FailedIDs  []int64 `json:"failed_ids"`
	Success    bool    `json:"success"`
	Message    string  `json:"message"`
}

type BulkUpdateStatusResponse struct {
	UpdatedIDs []int64 `json:"updated_ids"`
	FailedIDs  []int64 `json:"failed_ids"`
	Status     string  `json:"status"`
	Success    bool    `json:"success"`
	Message    string  `json:"message"`
}

// ===== PAGINATION REQUEST =====
type PaginationRequest struct {
	Page     int32 `json:"page" validate:"min=1"`
	PageSize int32 `json:"page_size" validate:"min=1,max=100"`
}

// ===== FILTERS REQUEST =====
type UserFiltersRequest struct {
	Role      string `json:"role,omitempty"`
	BranchID  int64  `json:"branch_id,omitempty"`
	Status    string `json:"status,omitempty"`
	OwnerID   int64  `json:"owner_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"` // Format: "2006-01-02" or date range
}

// ===== COMBINED SEARCH REQUEST =====
type AdvancedSearchRequest struct {
	Query      string                `json:"query,omitempty"`
	Filters    UserFiltersRequest    `json:"filters"`
	Pagination PaginationRequest     `json:"pagination"`
	SortBy     string               `json:"sort_by,omitempty" validate:"omitempty,oneof=id name email created_at updated_at"`
	SortOrder  string               `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// ===== COMBINED SEARCH RESPONSE =====
type AdvancedSearchResponse struct {
	Users      []UserProfile `json:"users"`
	Total      int64         `json:"total"`
	Page       int32         `json:"page"`
	PageSize   int32         `json:"page_size"`
	TotalPages int32         `json:"total_pages"`
	Filters    UserFiltersRequest `json:"applied_filters"`
}

// ===== GENERIC RESPONSES =====

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
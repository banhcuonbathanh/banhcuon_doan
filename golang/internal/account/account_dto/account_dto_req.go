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

// ===== BULK OPERATIONS REQUESTS =====
type BulkDeleteUsersRequest struct {
	UserIDs []int64 `json:"user_ids" validate:"required,min=1,dive,gt=0"`
}

type BulkUpdateStatusRequest struct {
	UserIDs []int64 `json:"user_ids" validate:"required,min=1,dive,gt=0"`
	Status  string  `json:"status" validate:"required,oneof=active inactive suspended pending"`
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
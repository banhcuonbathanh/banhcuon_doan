// model/auth_models.go
package account_dto

import "time"

// LoginUserRes represents the login response
type LoginUserRes struct {
	AccessToken  string                `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string                `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         AccountLoginResponse  `json:"user"`
}

// AccountLoginResponse represents the user data in login response
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
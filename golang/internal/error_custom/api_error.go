
package errorcustom

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents a structured API error with detailed information
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
}

// NewAPIError creates a new APIError instance
func NewAPIError(code, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// WithDetail adds a key-value pair to error details
func (e *APIError) WithDetail(key string, value interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// ToJSON converts the error to JSON bytes
func (e *APIError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Domain-specific error types
// =============================

// UserNotFoundError represents when a user cannot be found
type UserNotFoundError struct {
	ID    int64  `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
}

func (e *UserNotFoundError) Error() string {
	if e.ID != 0 {
		return fmt.Sprintf("user with ID %d not found", e.ID)
	}
	if e.Email != "" {
		return fmt.Sprintf("user with email %s not found", e.Email)
	}
	return "user not found"
}

func (e *UserNotFoundError) ToAPIError() *APIError {
	apiErr := NewAPIError("USER_NOT_FOUND", e.Error(), http.StatusNotFound)
	if e.ID != 0 {
		apiErr.WithDetail("user_id", e.ID)
	}
	if e.Email != "" {
		apiErr.WithDetail("email", e.Email)
	}
	return apiErr
}

// ValidationError represents validation failures
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) ToAPIError() *APIError {
	return NewAPIError("VALIDATION_ERROR", e.Error(), http.StatusBadRequest).
		WithDetail("field", e.Field).
		WithDetail("value", e.Value)
}

// AuthenticationError represents authentication failures
type AuthenticationError struct {
	Reason string `json:"reason"`
}

func (e *AuthenticationError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("authentication failed: %s", e.Reason)
	}
	return "authentication failed"
}

func (e *AuthenticationError) ToAPIError() *APIError {
	return NewAPIError("AUTHENTICATION_ERROR", "Invalid credentials", http.StatusUnauthorized).
		WithDetail("reason", e.Reason)
}

// AuthorizationError represents authorization failures
type AuthorizationError struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

func (e *AuthorizationError) Error() string {
	return fmt.Sprintf("not authorized to %s %s", e.Action, e.Resource)
}

func (e *AuthorizationError) ToAPIError() *APIError {
	return NewAPIError("AUTHORIZATION_ERROR", "Access denied", http.StatusForbidden).
		WithDetail("action", e.Action).
		WithDetail("resource", e.Resource)
}

// DuplicateEmailError represents email already exists error
type DuplicateEmailError struct {
	Email string `json:"email"`
}

func (e *DuplicateEmailError) Error() string {
	return fmt.Sprintf("email %s already exists", e.Email)
}

func (e *DuplicateEmailError) ToAPIError() *APIError {
	return NewAPIError("DUPLICATE_EMAIL", "Email already registered", http.StatusConflict).
		WithDetail("email", e.Email)
}

// InvalidTokenError represents token-related errors
type InvalidTokenError struct {
	TokenType string `json:"token_type"`
	Reason    string `json:"reason"`
}

func (e *InvalidTokenError) Error() string {
	return fmt.Sprintf("invalid %s token: %s", e.TokenType, e.Reason)
}

func (e *InvalidTokenError) ToAPIError() *APIError {
	return NewAPIError("INVALID_TOKEN", "Token is invalid or expired", http.StatusUnauthorized).
		WithDetail("token_type", e.TokenType).
		WithDetail("reason", e.Reason)
}

// BranchNotFoundError represents when a branch cannot be found
type BranchNotFoundError struct {
	ID int64 `json:"id"`
}

func (e *BranchNotFoundError) Error() string {
	return fmt.Sprintf("branch with ID %d not found", e.ID)
}

func (e *BranchNotFoundError) ToAPIError() *APIError {
	return NewAPIError("BRANCH_NOT_FOUND", e.Error(), http.StatusNotFound).
		WithDetail("branch_id", e.ID)
}

// PasswordValidationError represents password validation failures
type PasswordValidationError struct {
	Requirements []string `json:"requirements"`
}

func (e *PasswordValidationError) Error() string {
	return "password does not meet requirements"
}

func (e *PasswordValidationError) ToAPIError() *APIError {
	return NewAPIError("WEAK_PASSWORD", "Password does not meet security requirements", http.StatusBadRequest).
		WithDetail("requirements", e.Requirements)
}

// Helper functions for common error patterns
// ==========================================

// NewUserNotFoundByID creates a UserNotFoundError with ID
func NewUserNotFoundByID(id int64) *UserNotFoundError {
	return &UserNotFoundError{ID: id}
}

// NewUserNotFoundByEmail creates a UserNotFoundError with email
func NewUserNotFoundByEmail(email string) *UserNotFoundError {
	return &UserNotFoundError{Email: email}
}

// NewValidationError creates a ValidationError
func NewValidationError(field, message, value string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// NewAuthenticationError creates an AuthenticationError
func NewAuthenticationError(reason string) *AuthenticationError {
	return &AuthenticationError{Reason: reason}
}

// NewDuplicateEmailError creates a DuplicateEmailError
func NewDuplicateEmailError(email string) *DuplicateEmailError {
	return &DuplicateEmailError{Email: email}
}

// NewInvalidTokenError creates an InvalidTokenError
func NewInvalidTokenError(tokenType, reason string) *InvalidTokenError {
	return &InvalidTokenError{
		TokenType: tokenType,
		Reason:    reason,
	}
}

// Error code constants
const (
	// User-related errors
	ErrCodeUserNotFound    = "USER_NOT_FOUND"
	ErrCodeDuplicateEmail  = "DUPLICATE_EMAIL"
	ErrCodeWeakPassword    = "WEAK_PASSWORD"
	
	// Auth-related errors
	ErrCodeAuthFailed      = "AUTHENTICATION_ERROR"
	ErrCodeAccessDenied    = "AUTHORIZATION_ERROR"
	ErrCodeInvalidToken    = "INVALID_TOKEN"
		ErrCodeNotFound        = "NOT_FOUND"    
	// Validation errors
	ErrCodeValidationError = "VALIDATION_ERROR"
	ErrCodeInvalidInput    = "INVALID_INPUT"
	
	// System errors
	ErrCodeInternalError   = "INTERNAL_ERROR"
	ErrCodeServiceError    = "SERVICE_ERROR"
)
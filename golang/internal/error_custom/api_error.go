// internal/error_custom/errors.go

package errorcustom

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError represents a structured API error with detailed information
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Layer      string                 `json:"layer,omitempty"`      // handler, service, repository
	Operation  string                 `json:"operation,omitempty"`  // login, register, etc.
	Cause      error                  `json:"-"`                    // Original error for internal use
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




// ErrorResponse represents the standard error format for API responses
// swagger:model ErrorResponse
type ErrorResponse struct {
	Code    string                 `json:"code" example:"validation_error"`
	Message string                 `json:"message" example:"Validation failed"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToErrorResponse converts APIError to Swagger-compatible format
func (e *APIError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
}

// NewErrorResponse creates a new ErrorResponse instance
func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// WithDetail adds a detail to ErrorResponse
func (er ErrorResponse) WithDetail(key string, value interface{}) ErrorResponse {
	if er.Details == nil {
		er.Details = make(map[string]interface{})
	}
	er.Details[key] = value
	return er
}




func (e *ServiceError) Error() string {
	return e.Message
}

func (e *ServiceError) ToAPIError() *APIError {
	return NewAPIError(
		ErrCodeServiceError, 
		e.Message, 
		http.StatusInternalServerError,
	)
}

// RepositoryError represents an error from the repository/data layer
type RepositoryError struct {
	Message string
	Details map[string]interface{}
}

func (e *RepositoryError) Error() string {
	return e.Message
}

func (e *RepositoryError) ToAPIError() *APIError {
	return NewAPIError(
		ErrCodeInternalError, 
		e.Message, 
		http.StatusInternalServerError,
	)
}

// Add to APIError methods in internal/error_custom/errors.go
func (e *APIError) GetLogContext() map[string]interface{} {
    context := map[string]interface{}{
        "code":        e.Code,
        "message":     e.Message,
        "http_status": e.HTTPStatus,
    }
    if e.Details != nil {
        context["details"] = e.Details
    }
    return context
}



// Add this new function to parse gRPC errors


// Add this new constructor for context-rich API errors







// Add these functions to your errorcustom package
// Make sure to import "strings" at the top of the file

// Helper function to determine if error is related to user not found vs password mismatch
func IsUserNotFoundError(err error) bool {

	return strings.Contains(err.Error(), "user not found") || strings.Contains(err.Error(), "email not found")
}

// Helper function to determine if error is password related
func IsPasswordError(err error) bool {

	return strings.Contains(err.Error(), "password")
}

// ParseGRPCError parses gRPC error messages and creates appropriate errors
func ParseGRPCError(err error, operation string, email string) error {
	if err == nil {
		return nil
	}
	
	errMsg := err.Error()
	
	// Check for specific error patterns
	switch {
	case strings.Contains(errMsg, "user not found"):
		return NewEmailNotFoundError(email).ToAPIError()
	case strings.Contains(errMsg, "invalid email or password"):
		// This is ambiguous - we need better error handling from service layer
		return NewPasswordMismatchError(email).ToAPIError()
	case strings.Contains(errMsg, "account disabled"):
		return NewAccountDisabledError(email).ToAPIError()  
	case strings.Contains(errMsg, "account locked"):
		return NewAccountLockedError(email, "security policy").ToAPIError()
	case strings.Contains(errMsg, "already exists"):
		return NewAPIErrorWithContext(
			ErrCodeDuplicateEmail,
			"Email already registered", 
			http.StatusConflict,
			"service",
			operation,
			err,
		)
	case strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "unavailable"):
		return NewServiceError("AccountService", operation, "Service unavailable", err, true).ToAPIError()
	default:
		return NewServiceError("AccountService", operation, "Unknown service error", err, false).ToAPIError()
	}
}

// Enhanced error creation functions with better context
// ====================================================

// NewEmailNotFoundError creates an authentication error for email not found
func NewEmailNotFoundError(email string) *AuthenticationError {
	return &AuthenticationError{
		Email:     email,
		Reason:    "email not found",
		Step:      "email_check",
		UserFound: false,
	}
}

// NewPasswordMismatchError creates an authentication error for password mismatch
func NewPasswordMismatchError(email string) *AuthenticationError {
	return &AuthenticationError{
		Email:     email,
		Reason:    "password mismatch", 
		Step:      "password_check",
		UserFound: true,
	}
}

// NewAccountDisabledError creates an authentication error for disabled account
func NewAccountDisabledError(email string) *AuthenticationError {
	return &AuthenticationError{
		Email:     email,
		Reason:    "account disabled",
		Step:      "status_check", 
		UserFound: true,
	}
}

// NewAccountLockedError creates an authentication error for locked account
func NewAccountLockedError(email string, lockReason string) *AuthenticationError {
	return &AuthenticationError{
		Email:     email,
		Reason:    fmt.Sprintf("account locked: %s", lockReason),
		Step:      "status_check",
		UserFound: true,
	}
}


// Add these type definitions to your errorcustom package

// AuthenticationError represents authentication failures with specific reasons
type AuthenticationError struct {
	Email     string `json:"email,omitempty"`
	Reason    string `json:"reason"`
	Step      string `json:"step,omitempty"`      // email_check, password_check, token_validation
	UserFound bool   `json:"user_found,omitempty"` // Whether user exists in system
}




// ServiceError represents errors from service layer operations
type ServiceError struct {
	Service   string `json:"service"`
	Method    string `json:"method"`
	Message   string `json:"message"`
	Cause     error  `json:"-"`
	Retryable bool   `json:"retryable"`
}




// NewServiceError creates a new service error
func NewServiceError(service, method, message string, cause error, retryable bool) *ServiceError {
	return &ServiceError{
		Service:   service,
		Method:    method,
		Message:   message,
		Cause:     cause,
		Retryable: retryable,
	}
}

// APIError with enhanced context tracking


// WithDetail adds a key-value pair to error details

// NewAPIErrorWithContext creates a new APIError instance with context
func NewAPIErrorWithContext(code, message string, httpStatus int, layer, operation string, cause error) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Layer:      layer,
		Operation:  operation,
		Cause:      cause,
	}
}


// Error code constants
const (

	ErrCodeRepositoryError = "REPOSITORY_ERROR"
)






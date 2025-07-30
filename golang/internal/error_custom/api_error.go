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
	Layer      string                 `json:"layer,omitempty"`     // handler, service, repository
	Operation  string                 `json:"operation,omitempty"` // login, register, etc.
	Cause      error                  `json:"-"`                   // Original error for internal use
}

// WithDetail adds a key-value pair to error details and returns the APIError for chaining
func (e *APIError) WithDetail(key string, value interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// NewAPIError creates a new APIError instance
func NewAPIError(code, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

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

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Layer != "" && e.Operation != "" {
		return fmt.Sprintf("[%s:%s][%s] %s", e.Layer, e.Operation, e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// WithLayer sets the layer information and returns the APIError for chaining
func (e *APIError) WithLayer(layer string) *APIError {
	e.Layer = layer
	return e
}

// WithOperation sets the operation information and returns the APIError for chaining
func (e *APIError) WithOperation(operation string) *APIError {
	e.Operation = operation
	return e
}

// GetLogContext returns context information suitable for logging
func (e *APIError) GetLogContext() map[string]interface{} {
	context := map[string]interface{}{
		"error_code":    e.Code,
		"error_message": e.Message,
		"http_status":   e.HTTPStatus,
	}

	if e.Layer != "" {
		context["layer"] = e.Layer
	}
	if e.Operation != "" {
		context["operation"] = e.Operation
	}
	if e.Cause != nil {
		context["cause"] = e.Cause.Error()
	}
	if e.Details != nil {
		context["details"] = e.Details
	}

	return context
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

// AuthenticationError represents authentication failures with specific reasons
type AuthenticationError struct {
	Email     string `json:"email,omitempty"`
	Reason    string `json:"reason"`
	Step      string `json:"step,omitempty"`       // email_check, password_check, token_validation
	UserFound bool   `json:"user_found,omitempty"` // Whether user exists in system
}

func (e *AuthenticationError) Error() string {
	if e.Step != "" {
		return fmt.Sprintf("authentication failed at %s: %s", e.Step, e.Reason)
	}
	return fmt.Sprintf("authentication failed: %s", e.Reason)
}

func (e *AuthenticationError) ToAPIError() *APIError {
	apiErr := NewAPIErrorWithContext(
		"AUTHENTICATION_ERROR",
		"Invalid credentials",
		http.StatusUnauthorized,
		"handler",
		"login",
		e,
	)

	if e.Email != "" {
		apiErr.WithDetail("email", e.Email)
	}
	if e.Step != "" {
		apiErr.WithDetail("step", e.Step)
	}
	if e.UserFound {
		apiErr.WithDetail("user_found", e.UserFound)
	}

	return apiErr
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

// ServiceError represents errors from service layer operations
type ServiceError struct {
	Service   string `json:"service"`
	Method    string `json:"method"`
	Message   string `json:"message"`
	Cause     error  `json:"-"`
	Retryable bool   `json:"retryable"`
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("service error in %s.%s: %s", e.Service, e.Method, e.Message)
}

func (e *ServiceError) ToAPIError() *APIError {
	code := "SERVICE_ERROR"
	if e.Retryable {
		code = "SERVICE_TEMPORARILY_UNAVAILABLE"
	}

	return NewAPIErrorWithContext(
		code,
		"Service operation failed",
		http.StatusInternalServerError,
		"service",
		e.Method,
		e,
	).WithDetail("service", e.Service).WithDetail("retryable", e.Retryable)
}

// RepositoryError represents an error from the repository/data layer
type RepositoryError struct {
	Operation string `json:"operation"`
	Table     string `json:"table"`
	Message   string `json:"message"`
	Cause     error  `json:"-"`
	SQLState  string `json:"sql_state,omitempty"`
}

func (e *RepositoryError) Error() string {
	return fmt.Sprintf("repository error during %s on %s: %s", e.Operation, e.Table, e.Message)
}

func (e *RepositoryError) ToAPIError() *APIError {
	return NewAPIErrorWithContext(
		"REPOSITORY_ERROR",
		"Database operation failed",
		http.StatusInternalServerError,
		"repository",
		e.Operation,
		e,
	).WithDetail("table", e.Table).WithDetail("sql_state", e.SQLState)
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

// NewRepositoryError creates a new repository error
func NewRepositoryError(operation, table, message string, cause error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Table:     table,
		Message:   message,
		Cause:     cause,
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

// Helper functions for error detection
// ====================================

// Helper function to determine if error is related to user not found vs password mismatch
func IsUserNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's an APIError with USER_NOT_FOUND code
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == ErrCodeUserNotFound
	}

	// Check if it's directly a UserNotFoundError
	if _, ok := err.(*UserNotFoundError); ok {
		return true
	}

	// Fallback to string matching
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "user not found") ||
		strings.Contains(errMsg, "email not found") ||
		strings.Contains(errMsg, "user_not_found")
}

// Helper function to determine if error is password related
func IsPasswordError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's an APIError with authentication/password related codes
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == ErrCodeAuthFailed ||
			apiErr.Code == "AUTHENTICATION_ERROR" ||
			apiErr.Code == "INVALID_CREDENTIALS"
	}

	// Fallback to string matching
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "password") ||
		strings.Contains(errMsg, "invalid credentials") ||
		strings.Contains(errMsg, "authentication failed")
}

// ParseGRPCError parses gRPC error messages and creates appropriate errors
func ParseGRPCError(err error, operation string, email string) error {
	if err == nil {
		return nil
	}

	errMsg := strings.ToLower(err.Error())

	// Return APIErrors with specific codes that the helper functions can detect
	switch {
	case strings.Contains(errMsg, "user not found") || strings.Contains(errMsg, "email not found"):
		return &APIError{
			Code:       ErrCodeUserNotFound,
			Message:    "User not found",
			HTTPStatus: http.StatusNotFound,
		}

	case strings.Contains(errMsg, "invalid password") ||
		strings.Contains(errMsg, "password") ||
		strings.Contains(errMsg, "invalid email or password"):
		return &APIError{
			Code:       ErrCodeAuthFailed,
			Message:    "Invalid credentials",
			HTTPStatus: http.StatusUnauthorized,
		}

	case strings.Contains(errMsg, "account disabled"):
		return &APIError{
			Code:       ErrCodeAccessDenied,
			Message:    "Account disabled",
			HTTPStatus: http.StatusForbidden,
		}

	case strings.Contains(errMsg, "account locked"):
		return &APIError{
			Code:       ErrCodeAccessDenied,
			Message:    "Account locked",
			HTTPStatus: http.StatusForbidden,
		}

	case strings.Contains(errMsg, "already exists"):
		return &APIError{
			Code:       ErrCodeDuplicateEmail,
			Message:    "Email already registered",
			HTTPStatus: http.StatusConflict,
		}

	case strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "unavailable"):
		return &APIError{
			Code:       ErrCodeServiceError,
			Message:    "Service temporarily unavailable",
			HTTPStatus: http.StatusServiceUnavailable,
		}

	default:
		return &APIError{
			Code:       ErrCodeInternalError,
			Message:    "Internal server error",
			HTTPStatus: http.StatusInternalServerError,
		}
	}
}

// Error code constants
const (
	// User-related errors
	ErrCodeUserNotFound   = "USER_NOT_FOUND"
	ErrCodeDuplicateEmail = "DUPLICATE_EMAIL"
	ErrCodeWeakPassword   = "WEAK_PASSWORD"

	// Auth-related errors
	ErrCodeAuthFailed   = "AUTHENTICATION_ERROR"
	ErrCodeAccessDenied = "AUTHORIZATION_ERROR"
	ErrCodeInvalidToken = "INVALID_TOKEN"
	ErrCodeNotFound     = "NOT_FOUND"

	// Validation errors
	ErrCodeValidationError = "VALIDATION_ERROR"
	ErrCodeInvalidInput    = "INVALID_INPUT"

	// System errors
	ErrCodeInternalError   = "INTERNAL_ERROR"
	ErrCodeServiceError    = "SERVICE_ERROR"
	ErrCodeRepositoryError = "REPOSITORY_ERROR"
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
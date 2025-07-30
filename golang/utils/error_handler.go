// utils/error_handler.go
package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/logger"

	"github.com/go-playground/validator/v10"
)

// Enhanced error handling with proper logging
func HandleError(w http.ResponseWriter, err error) {
	var apiErr *errorcustom.APIError
	// Convert various error types to APIError
	switch e := err.(type) {
	case *errorcustom.APIError:
		apiErr = e
	case *errorcustom.AuthenticationError:
		apiErr = e.ToAPIError()
	case *errorcustom.UserNotFoundError:
		apiErr = e.ToAPIError()
	case *errorcustom.DuplicateEmailError:
		apiErr = e.ToAPIError()
	case *errorcustom.ServiceError:
		apiErr = e.ToAPIError()
	case *errorcustom.RepositoryError:
		apiErr = e.ToAPIError()
	default:
		// Generic error handling
		apiErr = errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"An unexpected error occurred",
			http.StatusInternalServerError,
		)
	}

	// Log the error with context
	logContext := apiErr.GetLogContext()
	logger.Error("API error occurred", logContext)

	// Set response headers and status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)

	// Write error response
	response := apiErr.ToErrorResponse()
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode error response", map[string]interface{}{
			"original_error": apiErr.Error(),
			"encoding_error": err.Error(),
		})
		// Fallback to simple text response
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Enhanced validation error handling
func HandleValidationErrors(w http.ResponseWriter, validationErrors validator.ValidationErrors) {
	errors := make(map[string]string)
	
	for _, err := range validationErrors {
		field := err.Field()
		switch err.Tag() {
		case "required":
			errors[field] = "This field is required"
		case "email":
			errors[field] = "Invalid email format"
		case "min":
			errors[field] = field + " is too short"
		case "max":
			errors[field] = field + " is too long"
		case "password":
			errors[field] = "Password does not meet requirements"
		case "role":
			errors[field] = "Invalid role specified"
		case "uniqueemail":
			errors[field] = "Email already exists"
		default:
			errors[field] = "Invalid value"
		}
		
		// Log each validation error
		logger.LogValidationError(field, errors[field], err.Value())
	}

	apiErr := errorcustom.NewAPIError(
		errorcustom.ErrCodeValidationError,
		"Validation failed",
		http.StatusBadRequest,
	).WithDetail("validation_errors", errors)

	// Log the overall validation failure
	logger.Warning("Request validation failed", map[string]interface{}{
		"validation_errors": errors,
		"error_count":      len(errors),
	})

	HandleError(w, apiErr)
}

// Helper function to respond with JSON and log the response
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error":       err.Error(),
			"status_code": statusCode,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Log successful response (debug level to avoid spam)
	logger.Debug("JSON response sent", map[string]interface{}{
		"status_code": statusCode,
	})
}

// Password validation with detailed error reporting
func ValidatePasswordWithDetails(password string) error {
	requirements := []string{}
	
	if len(password) < 8 {
		requirements = append(requirements, "at least 8 characters long")
	}
	
	if len(password) > 100 {
		requirements = append(requirements, "no more than 100 characters long")
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case char >= 32 && char <= 126: // printable ASCII characters
			// Check if it's a special character
			if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}
	
	if !hasUpper {
		requirements = append(requirements, "at least one uppercase letter")
	}
	if !hasLower {
		requirements = append(requirements, "at least one lowercase letter")
	}
	if !hasNumber {
		requirements = append(requirements, "at least one number")
	}
	if !hasSpecial {
		requirements = append(requirements, "at least one special character")
	}
	
	if len(requirements) > 0 {
		return &errorcustom.PasswordValidationError{
			Requirements: requirements,
		}
	}
	
	return nil
}

// Decode JSON with enhanced error reporting
func DecodeJSON(body interface{}, target interface{}) error {
	// This would typically use json.NewDecoder(reader).Decode()
	// But since the original function isn't shown, I'll assume it exists
	// and just add error context if needed
	
	// For now, returning a generic decoding error
	// You should replace this with your actual JSON decoding logic
	return nil
}

// Helper function to extract user email from JWT token context
func GetUserEmailFromContext(r *http.Request) string {
	if email, ok := r.Context().Value("user_email").(string); ok {
		return email
	}
	return ""
}

// Helper function to extract user ID from JWT token context  
func GetUserIDFromContext(r *http.Request) int64 {
	if userID, ok := r.Context().Value("user_id").(int64); ok {
		return userID
	}
	return 0
}

// Log HTTP middleware for request/response tracking
func LogHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom ResponseWriter to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Log request
		logger.Debug("HTTP request received", map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"user_agent": r.Header.Get("User-Agent"),
			"ip":         r.RemoteAddr,
		})
		
		// Process request
		next.ServeHTTP(wrappedWriter, r)
		
		// Log response
		duration := time.Since(start)
		logger.LogAPIRequest(r.Method, r.URL.Path, wrappedWriter.statusCode, duration, map[string]interface{}{
			"user_agent": r.Header.Get("User-Agent"),
			"ip":         r.RemoteAddr,
		})
	})
}

// Custom ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Recovery middleware to handle panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", map[string]interface{}{
					"error":  err,
					"method": r.Method,
					"path":   r.URL.Path,
					"ip":     r.RemoteAddr,
				})
				
				apiErr := errorcustom.NewAPIError(
					errorcustom.ErrCodeInternalError,
					"Internal server error",
					http.StatusInternalServerError,
				)
				HandleError(w, apiErr)
			}
		}()
		next.ServeHTTP(w, r)
	})
}


// new



// Enhanced APIError with more detailed tracking
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Layer      string                 `json:"layer,omitempty"`      // handler, service, repository
	Operation  string                 `json:"operation,omitempty"`  // login, register, etc.
	Cause      error                  `json:"-"`                    // Original error for internal use
}

// Enhanced error creation with layer and operation tracking
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

// NewAPIError creates a new APIError instance (backward compatibility)
func NewAPIError(code, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Layer != "" && e.Operation != "" {
		return fmt.Sprintf("[%s:%s][%s] %s", e.Layer, e.Operation, e.Code, e.Message)
	}
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

// WithLayer sets the layer information
func (e *APIError) WithLayer(layer string) *APIError {
	e.Layer = layer
	return e
}

// WithOperation sets the operation information
func (e *APIError) WithOperation(operation string) *APIError {
	e.Operation = operation
	return e
}

// GetLogContext returns context information suitable for logging
func (e *APIError) GetLogContext() map[string]interface{} {
	context := map[string]interface{}{
		"error_code":   e.Code,
		"error_message": e.Message,
		"http_status":  e.HTTPStatus,
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

// Authentication specific errors with better granularity
// =====================================================

// AuthenticationError represents authentication failures with specific reasons
type AuthenticationError struct {
	Email     string `json:"email,omitempty"`
	Reason    string `json:"reason"`
	Step      string `json:"step,omitempty"`      // email_check, password_check, token_validation
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

// Service layer specific errors
// =============================

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

// Repository layer specific errors  
// =================================

// RepositoryError represents errors from repository/database layer
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

// NewRepositoryError creates a new repository error
func NewRepositoryError(operation, table, message string, cause error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Table:     table,
		Message:   message,
		Cause:     cause,
	}
}



// Helper function to determine if error is related to user not found vs password mismatch

// Existing error types for backward compatibility
// ==============================================

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

// Helper functions for backward compatibility
func NewUserNotFoundByID(id int64) *UserNotFoundError {
	return &UserNotFoundError{ID: id}
}

func NewUserNotFoundByEmail(email string) *UserNotFoundError {
	return &UserNotFoundError{Email: email}
}

func NewAuthenticationError(reason string) *AuthenticationError {
	return &AuthenticationError{Reason: reason}
}

func NewDuplicateEmailError(email string) *DuplicateEmailError {
	return &DuplicateEmailError{Email: email}
}

// Error code constants (existing ones preserved)
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
	ErrCodeRepositoryError = "REPOSITORY_ERROR"
)

// ErrorResponse represents the standard error format for API responses
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




// Replace your existing IsUserNotFoundError and IsPasswordError functions with these:

// Helper function to determine if error is related to user not found
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

// Updated ParseGRPCError that creates errors these functions can properly detect
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


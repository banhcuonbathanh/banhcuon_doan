// utils/error_handler.go
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/logger"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)








func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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

// utils/error_handler.go


// Optimized validation error handling - single WARNING log
func HandleValidationErrors(w http.ResponseWriter, validationErrors validator.ValidationErrors, requestID string) {
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
	}

	// Single aggregated WARNING log for all validation errors


	apiErr := errorcustom.NewAPIError(
		errorcustom.ErrCodeValidationError,
		"Validation failed",
		http.StatusBadRequest,
	).WithDetail("validation_errors", errors)

	HandleError(w, apiErr, requestID)
}

// Helper function to respond with JSON and log the response
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error":       err.Error(),
			"status_code": statusCode,
			"request_id":  requestID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Successful responses are logged at INFO level in LogAPIRequest
}

// Enhanced password validation with single error log
func ValidatePasswordWithDetails(password string, requestID string) error {
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
		case char >= 32 && char <= 126:
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
		// Log password validation failure at WARNING level
		logger.Warning("Password validation failed", map[string]interface{}{
			"requirements": requirements,
			"request_id":   requestID,
			"type":         "password_validation",
		})
		
		return &errorcustom.PasswordValidationError{
			Requirements: requirements,
		}
	}
	
	return nil
}







func RespondWithAPIError(w http.ResponseWriter, apiErr *errorcustom.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)
	
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		log.Printf("JSON encoding error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getValidationMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	case "uuid", "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "numeric", "number":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid date/time (format: %s)", field, param)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Secure ID parameter parsing
func ParseIDParam(r *http.Request, paramName string) (int64, *errorcustom.APIError) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			http.StatusBadRequest,
		)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Invalid %s: must be a positive integer", paramName),
			http.StatusBadRequest,
		).WithDetail("value", idStr)
	}

	return id, nil
}

// Secure string parameter handling
func GetStringParam(r *http.Request, paramName string, minLen int) (string, *errorcustom.APIError) {
	value := chi.URLParam(r, paramName)
	if value == "" {
		return "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			http.StatusBadRequest,
		)
	}

	if minLen > 0 && len(value) < minLen {
		return "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("%s must be at least %d characters", paramName, minLen),
			http.StatusBadRequest,
		)
	}

	for _, r := range value {
		if r < 32 || r == 127 { // Block control characters
			return "", errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				fmt.Sprintf("Invalid characters in %s", paramName),
				http.StatusBadRequest,
			)
		}
	}

	return value, nil
}

// Robust password validation
func ValidatePassword(password string) *errorcustom.APIError {
	const (
		minLength = 8
		maxLength = 128
	)

	if len(password) < minLength {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
			fmt.Sprintf("Password must be at least %d characters", minLength),
			http.StatusBadRequest,
		)
	}

	if len(password) > maxLength {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
			fmt.Sprintf("Password cannot exceed %d characters", maxLength),
			http.StatusBadRequest,
		)
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?/", c):
			hasSpecial = true
		}
	}

	var requirements []string
	if !hasUpper {
		requirements = append(requirements, "uppercase letter")
	}
	if !hasLower {
		requirements = append(requirements, "lowercase letter")
	}
	if !hasDigit {
		requirements = append(requirements, "digit")
	}
	if !hasSpecial {
		requirements = append(requirements, "special character")
	}

	if len(requirements) > 0 {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
			"Password does not meet security requirements",
			http.StatusBadRequest,
		).WithDetail("requirements", requirements)
	}

	return nil
}

// Safe pagination parameters
func GetPaginationParams(r *http.Request) (limit, offset int64, apiErr *errorcustom.APIError) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit = 10
	offset = 0

	if limitStr != "" {
		l, err := strconv.ParseInt(limitStr, 10, 64)
		switch {
		case err != nil:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid limit parameter",
				http.StatusBadRequest,
			)
		case l < 1:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Limit must be at least 1",
				http.StatusBadRequest,
			)
		case l > 100:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Limit cannot exceed 100",
				http.StatusBadRequest,
			)
		default:
			limit = l
		}
	}

	if offsetStr != "" {
		o, err := strconv.ParseInt(offsetStr, 10, 64)
		switch {
		case err != nil:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid offset parameter",
				http.StatusBadRequest,
			)
		case o < 0:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Offset cannot be negative",
				http.StatusBadRequest,
			)
		default:
			offset = o
		}
	}

	return limit, offset, nil
}

// Safe sorting parameters
func GetSortParams(r *http.Request, allowedFields []string) (sortBy, sortOrder string, apiErr *errorcustom.APIError) {
	sortBy = strings.ToLower(r.URL.Query().Get("sort_by"))
	sortOrder = strings.ToLower(r.URL.Query().Get("sort_order"))

	// Set safe defaults
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		return "", "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Invalid sort order. Use 'asc' or 'desc'",
			http.StatusBadRequest,
		)
	}

	// Validate field against allowlist
	if len(allowedFields) > 0 {
		valid := false
		for _, field := range allowedFields {
			if sortBy == field {
				valid = true
				break
			}
		}

		if !valid {
			return "", "", errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				fmt.Sprintf("Invalid sort field. Allowed: %v", allowedFields),
				http.StatusBadRequest,
			)
		}
	}

	return sortBy, sortOrder, nil
}

// Utility function for pagination metadata
func CalculatePagination(total, limit, offset int) (currentPage, totalPages int) {
	currentPage = (offset / limit) + 1
	totalPages = total / limit
	if total%limit > 0 {
		totalPages++
	}
	return currentPage, totalPages
}



func CalculatePaginationBounds(start, end, total int) (int, int) {
	if start >= total {
		start = total
	}
	if end >= total {
		end = total
	}
	return start, end
}


func RespondWithError(w http.ResponseWriter, status int, message string, requestID string) {
	RespondWithJSON(w, status, map[string]string{"error": message}, requestID)
}




// ytrhfhghjgvjhvhjv



// utils/error_handler.go

// Enhanced error handling with proper logging optimization
func HandleError(w http.ResponseWriter, err error, requestID string) {
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

	// Only log ERROR level for critical system failures (5xx errors)
	if apiErr.HTTPStatus >= 500 {
		logContext := apiErr.GetLogContext()
		logContext["request_id"] = requestID
		logger.Error("Critical system error occurred", logContext)
	}
	// For client errors (4xx), we'll log them in the API request log instead

	// Set response headers and status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)

	// Write error response
	response := apiErr.ToErrorResponse()
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode error response", map[string]interface{}{
			"original_error": apiErr.Error(),
			"encoding_error": err.Error(),
			"request_id":     requestID,
		})
		// Fallback to simple text response
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}




// Enhanced JSON decoding with optional raw request logging
func DecodeJSON(body io.Reader, target interface{}, requestID string, logRawBody bool) error {
	// Read the body first
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Failed to read request body",
			http.StatusBadRequest,
		).WithDetail("error", err.Error())
	}


	// Perform JSON decoding
	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Invalid JSON format",
			http.StatusBadRequest,
		).WithDetail("error", err.Error())
	}
	
	return nil
}

// Request ID middleware
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		
		// Add request ID to response headers for client debugging
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := r.Context()
		ctx = withRequestID(ctx, requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// Log HTTP middleware with optimized logging levels
func LogHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	
		requestID := GetRequestIDFromContext(r.Context())
		clientIP := GetClientIP(r)
		
		// Only DEBUG level for request initiation
		logger.Debug("HTTP request received", map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"user_agent": r.Header.Get("User-Agent"),
			"ip":         clientIP,
			"request_id": requestID,
		})
		
		// Create a custom ResponseWriter to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process request
		next.ServeHTTP(wrappedWriter, r)
		


	})
}

// Custom ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}


// Recovery middleware with ERROR level logging for panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestIDFromContext(r.Context())
				
				logger.Error("Panic recovered", map[string]interface{}{
					"error":      err,
					"method":     r.Method,
					"path":       r.URL.Path,
					"ip":         GetClientIP(r),
					"request_id": requestID,
				})
				
				apiErr := errorcustom.NewAPIError(
					errorcustom.ErrCodeInternalError,
					"Internal server error",
					http.StatusInternalServerError,
				)
				HandleError(w, apiErr, requestID)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Helper functions for request ID management
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), rand.Int63n(1000))
}

func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
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

// GetClientIP extracts the real client IP from request
func GetClientIP(r *http.Request) string {
	// Get IP from X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP (client IP)
		return strings.Split(forwarded, ",")[0]
	}
	
	// Get IP from X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// Get IP from request RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}


// error_system/core.go
package error_system

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ErrorCode represents specific error types that clients can handle
type ErrorCode string

const (
	// User/Account errors
	ErrAccountDuplicate    ErrorCode = "ACCOUNT_DUPLICATE"
	ErrAccountNotFound     ErrorCode = "ACCOUNT_NOT_FOUND"
	ErrInvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
	ErrAccountSuspended    ErrorCode = "ACCOUNT_SUSPENDED"
	
	// Validation errors
	ErrInvalidInput        ErrorCode = "INVALID_INPUT"
	ErrValidationFailed    ErrorCode = "VALIDATION_FAILED"
	ErrMissingField        ErrorCode = "MISSING_FIELD"
	
	// System errors
	ErrDatabaseError       ErrorCode = "DATABASE_ERROR"
	ErrServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrTimeout             ErrorCode = "TIMEOUT"
	ErrContextCancelled    ErrorCode = "REQUEST_CANCELLED"
	
	// Permission errors
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden          ErrorCode = "FORBIDDEN"
	
	// Generic fallback
	ErrSystemError         ErrorCode = "SYSTEM_ERROR"
)

// AppError represents our application error with all necessary context
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Internal   error                  `json:"-"` // Original error for logging

		
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// IsAppError checks if error is our AppError type
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// Core error creation functions
func NewError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatusForCode(code),
		Details:    make(map[string]interface{}),
	}
}

func NewErrorWithInternal(code ErrorCode, message string, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatusForCode(code),
		Details:    make(map[string]interface{}),
		Internal:   internal,
	}
}

// Specific error constructors for common cases
func AccountDuplicate(email string) *AppError {
	return &AppError{
		Code:       ErrAccountDuplicate,
		Message:    "Account with this email already exists",
		HTTPStatus: http.StatusConflict,
		Details: map[string]interface{}{
			"email": maskEmail(email),
		},
	}
}

func ValidationError(field, reason string) *AppError {
	return &AppError{
		Code:       ErrValidationFailed,
		Message:    fmt.Sprintf("Validation failed: %s", reason),
		HTTPStatus: http.StatusBadRequest,
		Details: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
	}
}

func DatabaseError(operation string, internal error) *AppError {
	// Check if it's a known database error pattern
	if strings.Contains(strings.ToLower(internal.Error()), "duplicate") {
		if strings.Contains(strings.ToLower(internal.Error()), "email") {
			return &AppError{
				Code:       ErrAccountDuplicate,
				Message:    "Account with this email already exists",
				HTTPStatus: http.StatusConflict,
				Internal:   internal,
				Details: map[string]interface{}{
					"operation": operation,
				},
			}
		}
		return &AppError{
			Code:       ErrValidationFailed,
			Message:    "Duplicate entry detected",
			HTTPStatus: http.StatusConflict,
			Internal:   internal,
		}
	}
	
	return &AppError{
		Code:       ErrDatabaseError,
		Message:    "Database operation failed",
		HTTPStatus: http.StatusInternalServerError,
		Internal:   internal,
		Details: map[string]interface{}{
			"operation": operation,
		},
	}
}

func ContextError(ctx context.Context) *AppError {
	if ctx.Err() == context.Canceled {
		return &AppError{
			Code:       ErrContextCancelled,
			Message:    "Request was cancelled",
			HTTPStatus: http.StatusRequestTimeout,
			Internal:   ctx.Err(),
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		return &AppError{
			Code:       ErrTimeout,
			Message:    "Request timeout",
			HTTPStatus: http.StatusRequestTimeout,
			Internal:   ctx.Err(),
		}
	}
	return &AppError{
		Code:       ErrSystemError,
		Message:    "Context error",
		HTTPStatus: http.StatusInternalServerError,
		Internal:   ctx.Err(),
	}
}

// getHTTPStatusForCode maps error codes to HTTP status codes
func getHTTPStatusForCode(code ErrorCode) int {
	switch code {
	case ErrAccountDuplicate:
		return http.StatusConflict
	case ErrAccountNotFound:
		return http.StatusNotFound
	case ErrInvalidCredentials:
		return http.StatusUnauthorized
	case ErrInvalidInput, ErrValidationFailed, ErrMissingField:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrTimeout, ErrContextCancelled:
		return http.StatusRequestTimeout
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// HTTP Response helpers
func (e *AppError) WriteHTTPResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus)
	
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    e.Code,
			"message": e.Message,
		},
	}
	
	if len(e.Details) > 0 {
		response["error"].(map[string]interface{})["details"] = e.Details
	}
	
	json.NewEncoder(w).Encode(response)
}

// Helper function to mask email for logging/responses
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 {
		if len(email) <= 2 {
			return "***"
		}
		return email[:1] + "***"
	}
	
	username := email[:atIndex]
	domain := email[atIndex:]
	
	if len(username) <= 2 {
		return "**" + domain
	} else if len(username) <= 4 {
		return username[:1] + "**" + username[len(username)-1:] + domain
	} else {
		return username[:2] + "***" + username[len(username)-1:] + domain
	}
}
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
	MessageVN  string                 `json:"message_vn"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Internal   error                  `json:"-"`
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
		MessageVN:  getVietnameseMessage(code, message),
		HTTPStatus: getHTTPStatusForCode(code),
		Details:    make(map[string]interface{}),
	}
}

func NewErrorWithMessages(code ErrorCode, message, messageVN string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		MessageVN:  messageVN,
		HTTPStatus: getHTTPStatusForCode(code),
		Details:    make(map[string]interface{}),
	}
}

func NewErrorWithInternal(code ErrorCode, message string, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		MessageVN:  getVietnameseMessage(code, message),
		HTTPStatus: getHTTPStatusForCode(code),
		Details:    make(map[string]interface{}),
		Internal:   internal,
	}
}

func NewErrorWithInternalAndMessages(code ErrorCode, message, messageVN string, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		MessageVN:  messageVN,
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
		MessageVN:  "Tài khoản với email này đã tồn tại",
		HTTPStatus: http.StatusConflict,
		Details: map[string]interface{}{
			"email": maskEmail(email),
		},
	}
}

func AccountNotFound() *AppError {
	return &AppError{
		Code:       ErrAccountNotFound,
		Message:    "Account not found",
		MessageVN:  "Không tìm thấy tài khoản",
		HTTPStatus: http.StatusNotFound,
		Details:    make(map[string]interface{}),
	}
}

func InvalidCredentials() *AppError {
	return &AppError{
		Code:       ErrInvalidCredentials,
		Message:    "Invalid email or password",
		MessageVN:  "Email hoặc mật khẩu không hợp lệ",
		HTTPStatus: http.StatusUnauthorized,
		Details:    make(map[string]interface{}),
	}
}

func AccountSuspended() *AppError {
	return &AppError{
		Code:       ErrAccountSuspended,
		Message:    "Account has been suspended",
		MessageVN:  "Tài khoản đã bị tạm ngưng",
		HTTPStatus: http.StatusForbidden,
		Details:    make(map[string]interface{}),
	}
}

func ValidationError(field, reason string) *AppError {
	return &AppError{
		Code:       ErrValidationFailed,
		Message:    fmt.Sprintf("Validation failed: %s", reason),
		MessageVN:  fmt.Sprintf("Xác thực thất bại: %s", reason),
		HTTPStatus: http.StatusBadRequest,
		Details: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
	}
}

func MissingField(field string) *AppError {
	return &AppError{
		Code:       ErrMissingField,
		Message:    fmt.Sprintf("Missing required field: %s", field),
		MessageVN:  fmt.Sprintf("Thiếu trường bắt buộc: %s", field),
		HTTPStatus: http.StatusBadRequest,
		Details: map[string]interface{}{
			"field": field,
		},
	}
}

func Unauthorized() *AppError {
	return &AppError{
		Code:       ErrUnauthorized,
		Message:    "Authentication required",
		MessageVN:  "Yêu cầu xác thực",
		HTTPStatus: http.StatusUnauthorized,
		Details:    make(map[string]interface{}),
	}
}

func Forbidden() *AppError {
	return &AppError{
		Code:       ErrForbidden,
		Message:    "Access denied",
		MessageVN:  "Truy cập bị từ chối",
		HTTPStatus: http.StatusForbidden,
		Details:    make(map[string]interface{}),
	}
}

func DatabaseError(operation string, internal error) *AppError {
	// Check if it's a known database error pattern
	if strings.Contains(strings.ToLower(internal.Error()), "duplicate") {
		if strings.Contains(strings.ToLower(internal.Error()), "email") {
			return &AppError{
				Code:       ErrAccountDuplicate,
				Message:    "Account with this email already exists",
				MessageVN:  "Tài khoản với email này đã tồn tại",
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
			MessageVN:  "Phát hiện bản ghi trùng lặp",
			HTTPStatus: http.StatusConflict,
			Internal:   internal,
		}
	}
	
	return &AppError{
		Code:       ErrDatabaseError,
		Message:    "Database operation failed",
		MessageVN:  "Thao tác cơ sở dữ liệu thất bại",
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
			MessageVN:  "Yêu cầu đã bị hủy",
			HTTPStatus: http.StatusRequestTimeout,
			Internal:   ctx.Err(),
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		return &AppError{
			Code:       ErrTimeout,
			Message:    "Request timeout",
			MessageVN:  "Yêu cầu hết thời gian chờ",
			HTTPStatus: http.StatusRequestTimeout,
			Internal:   ctx.Err(),
		}
	}
	return &AppError{
		Code:       ErrSystemError,
		Message:    "Context error",
		MessageVN:  "Lỗi ngữ cảnh",
		HTTPStatus: http.StatusInternalServerError,
		Internal:   ctx.Err(),
	}
}

func ServiceUnavailable() *AppError {
	return &AppError{
		Code:       ErrServiceUnavailable,
		Message:    "Service temporarily unavailable",
		MessageVN:  "Dịch vụ tạm thời không khả dụng",
		HTTPStatus: http.StatusServiceUnavailable,
		Details:    make(map[string]interface{}),
	}
}

// getVietnameseMessage provides default Vietnamese translations
func getVietnameseMessage(code ErrorCode, englishMessage string) string {
	defaultMessages := map[ErrorCode]string{
		ErrAccountDuplicate:   "Tài khoản với email này đã tồn tại",
		ErrAccountNotFound:    "Không tìm thấy tài khoản",
		ErrInvalidCredentials: "Email hoặc mật khẩu không hợp lệ",
		ErrAccountSuspended:   "Tài khoản đã bị tạm ngưng",
		ErrInvalidInput:       "Dữ liệu đầu vào không hợp lệ",
		ErrValidationFailed:   "Xác thực thất bại",
		ErrMissingField:       "Thiếu trường bắt buộc",
		ErrDatabaseError:      "Lỗi cơ sở dữ liệu",
		ErrServiceUnavailable: "Dịch vụ tạm thời không khả dụng",
		ErrTimeout:            "Hết thời gian chờ",
		ErrContextCancelled:   "Yêu cầu đã bị hủy",
		ErrUnauthorized:       "Yêu cầu xác thực",
		ErrForbidden:         "Truy cập bị từ chối",
		ErrSystemError:       "Lỗi hệ thống",
	}
	
	if vnMsg, exists := defaultMessages[code]; exists {
		return vnMsg
	}
	
	// Fallback to English message if no Vietnamese translation available
	return englishMessage
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
	case ErrAccountSuspended:
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
			"code":       e.Code,
			"message":    e.Message,
			"message_vn": e.MessageVN,
		},
	}
	
	if len(e.Details) > 0 {
		response["error"].(map[string]interface{})["details"] = e.Details
	}
	
	json.NewEncoder(w).Encode(response)
}

// WriteHTTPResponseWithLang allows language-specific response
func (e *AppError) WriteHTTPResponseWithLang(w http.ResponseWriter, lang string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus)
	
	message := e.Message
	if lang == "vi" || lang == "vn" {
		message = e.MessageVN
	}
	
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    e.Code,
			"message": message,
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

// 1. First, update error_system/core.go to add the enhanced validation function
func EnhancedValidationErrorWithDetails(field string, value interface{}, constraint string, allowedValues []string) *AppError {
	var message, messageVN string
	
	switch constraint {
	case "oneof":
		message = fmt.Sprintf("Invalid value '%v' for field '%s'. Allowed values are: %s", 
			value, field, strings.Join(allowedValues, ", "))
		messageVN = fmt.Sprintf("Giá trị '%v' không hợp lệ cho trường '%s'. Các giá trị được phép: %s", 
			value, field, strings.Join(allowedValues, ", "))
	case "required":
		message = fmt.Sprintf("Field '%s' is required", field)
		messageVN = fmt.Sprintf("Trường '%s' là bắt buộc", field)
	case "email":
		message = fmt.Sprintf("Field '%s' must be a valid email address", field)
		messageVN = fmt.Sprintf("Trường '%s' phải là địa chỉ email hợp lệ", field)
	case "min":
		message = fmt.Sprintf("Field '%s' is too short", field)
		messageVN = fmt.Sprintf("Trường '%s' quá ngắn", field)
	case "max":
		message = fmt.Sprintf("Field '%s' is too long", field)
		messageVN = fmt.Sprintf("Trường '%s' quá dài", field)
	default:
		message = fmt.Sprintf("Field '%s' failed validation", field)
		messageVN = fmt.Sprintf("Trường '%s' không hợp lệ", field)
	}
	
	details := map[string]interface{}{
		"field":      field,
		"value":      value,
		"constraint": constraint,
	}
	
	if len(allowedValues) > 0 {
		details["allowed_values"] = allowedValues
	}
	
	return &AppError{
		Code:       ErrValidationFailed,
		Message:    message,
		MessageVN:  messageVN,
		HTTPStatus: http.StatusBadRequest,
		Details:    details,
	}
}

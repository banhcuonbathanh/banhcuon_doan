// error_system/repository_handler.go
package error_system

import (
	"strings"

	"english-ai-full/logger/core"
)

// RepositoryErrorHandler handles database/repository layer errors
type RepositoryErrorHandler struct {
	logger *core.CoreLogger
	domain string
}

func NewRepositoryErrorHandler(logger *core.CoreLogger, domain string) *RepositoryErrorHandler {
	return &RepositoryErrorHandler{
		logger: logger,
		domain: domain,
	}
}

// Handle converts repository errors to AppError with proper logging
func (h *RepositoryErrorHandler) Handle(err error, operation, table string, context map[string]interface{}) *AppError {
	if err == nil {
		return nil
	}
	
	// Check if it's already our AppError - pass through unchanged
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, table, context)
		return appErr
	}
	
	// Handle database-specific errors
	appErr := h.handleDatabaseError(err, )
	h.logError(appErr, operation, table, context)
	
	return appErr
}

func (h *RepositoryErrorHandler) handleDatabaseError(err error) *AppError {
	errStr := strings.ToLower(err.Error())
	
	// Check for duplicate key violations - MOST SPECIFIC FIRST
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique constraint") {
		if strings.Contains(errStr, "email") {
			return NewErrorWithInternalAndMessages(
				ErrAccountDuplicate,
				"Account with this email already exists",
				"Tài khoản với email này đã tồn tại",
				err,
			)
		}
		return NewErrorWithInternalAndMessages(
			ErrValidationFailed,
			"Duplicate entry detected",
			"Phát hiện bản ghi trùng lặp",
			err,
		)
	}
	
	// Check for foreign key constraint violations
	if strings.Contains(errStr, "foreign key") || strings.Contains(errStr, "fk_") {
		return NewErrorWithInternalAndMessages(
			ErrValidationFailed,
			"Referenced record does not exist",
			"Bản ghi được tham chiếu không tồn tại",
			err,
		)
	}
	
	// Check for not null constraint violations
	if strings.Contains(errStr, "not null") || strings.Contains(errStr, "null value") {
		return NewErrorWithInternalAndMessages(
			ErrValidationFailed,
			"Required field cannot be empty",
			"Trường bắt buộc không được để trống",
			err,
		)
	}
	
	// Check for check constraint violations
	if strings.Contains(errStr, "check constraint") {
		return NewErrorWithInternalAndMessages(
			ErrValidationFailed,
			"Data constraint violation",
			"Vi phạm ràng buộc dữ liệu",
			err,
		)
	}
	
	// Check for connection/network errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") || 
	   strings.Contains(errStr, "connect") || strings.Contains(errStr, "refused") {
		return NewErrorWithInternalAndMessages(
			ErrServiceUnavailable,
			"Database connection error",
			"Lỗi kết nối cơ sở dữ liệu",
			err,
		)
	}
	
	// Check for timeout errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return NewErrorWithInternalAndMessages(
			ErrTimeout,
			"Database operation timeout",
			"Thao tác cơ sở dữ liệu hết thời gian chờ",
			err,
		)
	}
	
	// Check for transaction errors
	if strings.Contains(errStr, "transaction") || strings.Contains(errStr, "rollback") {
		return NewErrorWithInternalAndMessages(
			ErrDatabaseError,
			"Transaction failed",
			"Giao dịch thất bại",
			err,
		)
	}
	
	// Check for permission/access errors
	if strings.Contains(errStr, "permission") || strings.Contains(errStr, "access denied") ||
	   strings.Contains(errStr, "insufficient") {
		return NewErrorWithInternalAndMessages(
			ErrForbidden,
			"Database access denied",
			"Truy cập cơ sở dữ liệu bị từ chối",
			err,
		)
	}
	
	// Check for table/column not found errors
	if strings.Contains(errStr, "table") && (strings.Contains(errStr, "not found") || strings.Contains(errStr, "doesn't exist")) ||
	   strings.Contains(errStr, "column") && (strings.Contains(errStr, "not found") || strings.Contains(errStr, "doesn't exist")) {
		return NewErrorWithInternalAndMessages(
			ErrSystemError,
			"Database schema error",
			"Lỗi cấu trúc cơ sở dữ liệu",
			err,
		)
	}
	
	// Check for data too long errors
	if strings.Contains(errStr, "data too long") || strings.Contains(errStr, "value too long") {
		return NewErrorWithInternalAndMessages(
			ErrValidationFailed,
			"Data exceeds maximum length",
			"Dữ liệu vượt quá độ dài tối đa",
			err,
		)
	}
	
	// Check for disk full/storage errors
	if strings.Contains(errStr, "disk full") || strings.Contains(errStr, "no space") ||
	   strings.Contains(errStr, "storage") {
		return NewErrorWithInternalAndMessages(
			ErrServiceUnavailable,
			"Database storage error",
			"Lỗi lưu trữ cơ sở dữ liệu",
			err,
		)
	}
	
	// Generic database error
	return NewErrorWithInternalAndMessages(
		ErrDatabaseError,
		"Database operation failed",
		"Thao tác cơ sở dữ liệu thất bại",
		err,
	)
}

// HandleWithCustomMessage allows custom error messages while maintaining proper error codes
func (h *RepositoryErrorHandler) HandleWithCustomMessage(err error, operation, table, customMessage, customMessageVN string, context map[string]interface{}) *AppError {
	if err == nil {
		return nil
	}
	
	// Check if it's already our AppError - pass through unchanged
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, table, context)
		return appErr
	}
	
	// Determine appropriate error code from database error
	appErr := h.handleDatabaseError(err, )
	
	// Override messages with custom ones
	appErr.Message = customMessage
	appErr.MessageVN = customMessageVN
	
	h.logError(appErr, operation, table, context)
	
	return appErr
}

// HandleNotFound creates a not found error for repository operations
func (h *RepositoryErrorHandler) HandleNotFound(entity, identifier string) *AppError {
	return NewErrorWithMessages(
		ErrAccountNotFound,
		"Record not found",
		"Không tìm thấy bản ghi",
	)
}

// HandleDuplicateKey creates a duplicate key error with entity-specific message
func (h *RepositoryErrorHandler) HandleDuplicateKey(entity, field, value string) *AppError {
	message := "Duplicate entry detected"
	messageVN := "Phát hiện bản ghi trùng lặp"
	
	if field == "email" {
		message = "Account with this email already exists"
		messageVN = "Tài khoản với email này đã tồn tại"
	}
	
	appErr := NewErrorWithMessages(ErrAccountDuplicate, message, messageVN)
	appErr.Details = map[string]interface{}{
		"entity": entity,
		"field":  field,
		"value":  maskSensitiveValue(field, value),
	}
	
	return appErr
}

func (h *RepositoryErrorHandler) logError(appErr *AppError, operation, table string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"error_code": appErr.Code,
		"operation":  operation,
		"table":      table,
		"domain":     h.domain,
		"layer":      "repository",
		"message":    appErr.Message,
		"message_vn": appErr.MessageVN,
	}
	
	// Merge additional context
	for k, v := range context {
		logContext[k] = v
	}
	
	// Add internal error details for logging (not for client)
	if appErr.Internal != nil {
		logContext["internal_error"] = appErr.Internal.Error()
	}
	
	// Add error details if present
	if len(appErr.Details) > 0 {
		logContext["error_details"] = appErr.Details
	}
	
	h.logger.Error("Repository error occurred", logContext)
}

// Helper function to mask sensitive values in logs
func maskSensitiveValue(field, value string) string {
	switch strings.ToLower(field) {
	case "email":
		return maskEmail(value)
	case "phone", "mobile":
		return maskPhone(value)
	case "password", "pwd", "pass":
		return "***"
	default:
		// For other fields, mask if longer than 10 characters
		if len(value) > 10 {
			return value[:3] + "***" + value[len(value)-3:]
		}
		return value
	}
}

// Helper function to mask phone numbers
func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "***"
	}
	if len(phone) <= 8 {
		return phone[:2] + "***" + phone[len(phone)-2:]
	}
	return phone[:3] + "***" + phone[len(phone)-3:]
}
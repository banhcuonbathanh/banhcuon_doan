// error_system/repository_handler.go
package error_system

import (
	"database/sql"
	"errors"
	"strings"

	"english-ai-full/logger/core"
	"github.com/lib/pq" // Added for PostgreSQL error handling
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
	
	// NEW: Check for sql.ErrNoRows first (most common case)
	if errors.Is(err, sql.ErrNoRows) {
		appErr := NewErrorWithMessages(
			ErrAccountNotFound,
			"Record not found",
			"Không tìm thấy bản ghi",
		)
		h.logError(appErr, operation, table, context)
		return appErr
	}
	
	// NEW: Handle PostgreSQL-specific errors using lib/pq
	if pqErr, ok := err.(*pq.Error); ok {
		appErr := h.handlePostgresError(pqErr, context)
		h.logError(appErr, operation, table, context)
		return appErr
	}
	
	// Handle database-specific errors (existing logic as fallback)
	appErr := h.handleDatabaseError(err)
	h.logError(appErr, operation, table, context)
	
	return appErr
}

// NEW: handlePostgresError processes PostgreSQL-specific errors with detailed context
func (h *RepositoryErrorHandler) handlePostgresError(pqErr *pq.Error, context map[string]interface{}) *AppError {
	// Log the PostgreSQL error details for debugging
	h.logPostgresError(pqErr)
	
	switch pqErr.Code {
	case "23505": // unique_violation
		return h.handleUniqueViolation(pqErr, context)
		
	case "23503": // foreign_key_violation
		return h.handleForeignKeyViolation(pqErr, context)
		
	case "23502": // not_null_violation
		return h.handleNotNullViolation(pqErr)
		
	case "23514": // check_violation
		return h.handleCheckViolation(pqErr)
		
	case "42P01": // undefined_table
		return NewErrorWithInternalAndMessages(
			ErrSystemError,
			"Database configuration error - table not found",
			"Lỗi cấu hình cơ sở dữ liệu - không tìm thấy bảng",
			pqErr,
		)
		
	case "42703": // undefined_column
		return NewErrorWithInternalAndMessages(
			ErrSystemError,
			"Database configuration error - column not found",
			"Lỗi cấu hình cơ sở dữ liệu - không tìm thấy cột",
			pqErr,
		)
		
	case "08000", "08003", "08006": // connection errors
		return NewErrorWithInternalAndMessages(
			ErrServiceUnavailable,
			"Database connection error",
			"Lỗi kết nối cơ sở dữ liệu",
			pqErr,
		)
		
	case "57014": // query_canceled
		return NewErrorWithInternalAndMessages(
			ErrTimeout,
			"Database query was canceled",
			"Truy vấn cơ sở dữ liệu đã bị hủy",
			pqErr,
		)
		
	case "53300": // too_many_connections
		return NewErrorWithInternalAndMessages(
			ErrServiceUnavailable,
			"Too many database connections",
			"Quá nhiều kết nối cơ sở dữ liệu",
			pqErr,
		)
		
	default:
		// Generic PostgreSQL error
		return NewErrorWithInternalAndMessages(
			ErrDatabaseError,
			"Database operation failed: " + pqErr.Message,
			"Thao tác cơ sở dữ liệu thất bại",
			pqErr,
		)
	}
}

// NEW: handleUniqueViolation processes unique constraint violations with field detection
func (h *RepositoryErrorHandler) handleUniqueViolation(pqErr *pq.Error, context map[string]interface{}) *AppError {
	constraint := strings.ToLower(pqErr.Constraint)
	detail := strings.ToLower(pqErr.Detail)
	
	// Determine which field caused the violation
	var field, message, messageVN string
	var value interface{}
	
	if strings.Contains(constraint, "email") || strings.Contains(detail, "email") {
		field = "email"
		message = "Account with this email already exists"
		messageVN = "Tài khoản với email này đã tồn tại"
		if context != nil {
			if email, ok := context["email"]; ok {
				value = email
			}
		}
	} else if strings.Contains(constraint, "phone") || strings.Contains(detail, "phone") {
		field = "phone"
		message = "Phone number already exists"
		messageVN = "Số điện thoại đã tồn tại"
		if context != nil {
			if phone, ok := context["phone"]; ok {
				value = phone
			}
		}
	} else if strings.Contains(constraint, "username") || strings.Contains(detail, "username") {
		field = "username"
		message = "Username already exists"
		messageVN = "Tên người dùng đã tồn tại"
		if context != nil {
			if username, ok := context["username"]; ok {
				value = username
			}
		}
	} else {
		// Generic duplicate error
		field = "unknown"
		message = "Duplicate entry detected"
		messageVN = "Phát hiện bản ghi trùng lặp"
	}
	
	appErr := NewErrorWithInternalAndMessages(
		ErrAccountDuplicate,
		message,
		messageVN,
		pqErr,
	)
	
	// Add detailed information
	appErr.Details = map[string]interface{}{
		"field":      field,
		"constraint": pqErr.Constraint,
		"reason":     "duplicate",
	}
	
	if value != nil {
		appErr.Details["value"] = maskSensitiveValue(field, toString(value))
	}
	
	return appErr
}

// NEW: handleForeignKeyViolation processes foreign key constraint violations with intelligent field detection
func (h *RepositoryErrorHandler) handleForeignKeyViolation(pqErr *pq.Error, context map[string]interface{}) *AppError {
	constraint := strings.ToLower(pqErr.Constraint)

	
	// Determine which foreign key is violated and provide specific error
	var field, entity, message, messageVN string
	var value interface{}
	
	// Check constraint name for foreign key references
	if strings.Contains(constraint, "owner_id") {
		field = "owner_id"
		entity = "owner"
		message = "The specified owner does not exist"
		messageVN = "Chủ sở hữu được chỉ định không tồn tại"
		if context != nil {
			value = context["owner_id"]
		}
	} else if strings.Contains(constraint, "branch_id") {
		field = "branch_id"
		entity = "branch"
		message = "The specified branch does not exist"
		messageVN = "Chi nhánh được chỉ định không tồn tại"
		if context != nil {
			value = context["branch_id"]
		}
	} else if strings.Contains(constraint, "parent_id") {
		field = "parent_id"
		entity = "parent"
		message = "The specified parent record does not exist"
		messageVN = "Bản ghi cha được chỉ định không tồn tại"
		if context != nil {
			value = context["parent_id"]
		}
	} else if strings.Contains(constraint, "user_id") {
		field = "user_id"
		entity = "user"
		message = "The specified user does not exist"
		messageVN = "Người dùng được chỉ định không tồn tại"
		if context != nil {
			value = context["user_id"]
		}
	} else if strings.Contains(constraint, "role_id") {
		field = "role_id"
		entity = "role"
		message = "The specified role does not exist"
		messageVN = "Vai trò được chỉ định không tồn tại"
		if context != nil {
			value = context["role_id"]
		}
	} else if strings.Contains(constraint, "department_id") {
		field = "department_id"
		entity = "department"
		message = "The specified department does not exist"
		messageVN = "Phòng ban được chỉ định không tồn tại"
		if context != nil {
			value = context["department_id"]
		}
	} else if strings.Contains(constraint, "category_id") {
		field = "category_id"
		entity = "category"
		message = "The specified category does not exist"
		messageVN = "Danh mục được chỉ định không tồn tại"
		if context != nil {
			value = context["category_id"]
		}
	} else {
		// Generic foreign key error
		field = "reference"
		entity = "record"
		message = "Referenced record does not exist"
		messageVN = "Bản ghi được tham chiếu không tồn tại"
	}
	
	appErr := NewErrorWithInternalAndMessages(
		ErrValidationFailed,
		message,
		messageVN,
		pqErr,
	)
	
	// Add detailed information for debugging
	appErr.Details = map[string]interface{}{
		"field":      field,
		"entity":     entity,
		"constraint": pqErr.Constraint,
		"reason":     "invalid_reference",
	}
	
	if value != nil {
		appErr.Details["value"] = value
	}
	
	return appErr
}

// NEW: handleNotNullViolation processes NOT NULL constraint violations
func (h *RepositoryErrorHandler) handleNotNullViolation(pqErr *pq.Error) *AppError {
	column := pqErr.Column
	if column == "" {
		column = "unknown"
	}
	
	message := "Required field is missing: " + column
	messageVN := "Trường bắt buộc bị thiếu: " + column
	
	appErr := NewErrorWithInternalAndMessages(
		ErrValidationFailed,
		message,
		messageVN,
		pqErr,
	)
	
	appErr.Details = map[string]interface{}{
		"field":  column,
		"reason": "required",
	}
	
	return appErr
}

// NEW: handleCheckViolation processes CHECK constraint violations
func (h *RepositoryErrorHandler) handleCheckViolation(pqErr *pq.Error) *AppError {
	constraint := pqErr.Constraint
	
	message := "Data validation failed"
	messageVN := "Xác thực dữ liệu thất bại"
	
	// Provide more specific messages based on common check constraints
	constraintLower := strings.ToLower(constraint)
	if strings.Contains(constraintLower, "positive") {
		message = "Value must be positive"
		messageVN = "Giá trị phải là số dương"
	} else if strings.Contains(constraintLower, "range") {
		message = "Value is out of allowed range"
		messageVN = "Giá trị nằm ngoài phạm vi cho phép"
	} else if strings.Contains(constraintLower, "status") {
		message = "Invalid status value"
		messageVN = "Giá trị trạng thái không hợp lệ"
	}
	
	appErr := NewErrorWithInternalAndMessages(
		ErrValidationFailed,
		message,
		messageVN,
		pqErr,
	)
	
	appErr.Details = map[string]interface{}{
		"constraint": constraint,
		"reason":     "check_violation",
	}
	
	return appErr
}

// NEW: logPostgresError logs detailed PostgreSQL error information for debugging
func (h *RepositoryErrorHandler) logPostgresError(pqErr *pq.Error) {
	if h.logger != nil {
		h.logger.Debug("PostgreSQL error details", map[string]interface{}{
			"code":       string(pqErr.Code),
			"message":    pqErr.Message,
			"detail":     pqErr.Detail,
			"hint":       pqErr.Hint,
			"position":   pqErr.Position,
			"constraint": pqErr.Constraint,
			"table":      pqErr.Table,
			"column":     pqErr.Column,
			"schema":     pqErr.Schema,
			"severity":   pqErr.Severity,
		})
	}
}

// EXISTING: handleDatabaseError - kept as fallback for non-PostgreSQL errors
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

// EXISTING: HandleWithCustomMessage allows custom error messages while maintaining proper error codes
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
	appErr := h.handleDatabaseError(err)
	
	// Override messages with custom ones
	appErr.Message = customMessage
	appErr.MessageVN = customMessageVN
	
	h.logError(appErr, operation, table, context)
	
	return appErr
}

// EXISTING: HandleNotFound creates a not found error for repository operations
func (h *RepositoryErrorHandler) HandleNotFound(entity, identifier string) *AppError {
	return NewErrorWithMessages(
		ErrAccountNotFound,
		"Record not found",
		"Không tìm thấy bản ghi",
	)
}

// EXISTING: HandleDuplicateKey creates a duplicate key error with entity-specific message
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

// EXISTING: logError logs error with context
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

// EXISTING: Helper function to mask sensitive values in logs
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

// EXISTING: Helper function to mask phone numbers
func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "***"
	}
	if len(phone) <= 8 {
		return phone[:2] + "***" + phone[len(phone)-2:]
	}
	return phone[:3] + "***" + phone[len(phone)-3:]
}

// EXISTING: logErrorWithCause logs error with cause information using the repository logger
func (h *RepositoryErrorHandler) logErrorWithCause(cause, operation, table string, context map[string]interface{}) {
	if h.logger != nil {
		enhancedContext := make(map[string]interface{})
		for k, v := range context {
			enhancedContext[k] = v
		}
		enhancedContext[core.FieldCause] = cause
		enhancedContext[core.FieldOperation] = operation
		enhancedContext[core.FieldTable] = table
		enhancedContext["error_handler"] = "RepositoryErrorHandler"
		
		h.logger.ErrorWithCause("Database error categorized", cause, core.LayerRepository, operation, enhancedContext)
	}
}

// NEW: Helper function to convert interface{} to string
func toString(val interface{}) string {
	if val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}
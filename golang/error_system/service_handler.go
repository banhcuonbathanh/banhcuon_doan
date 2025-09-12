// error_system/service_handler.go
package error_system

import (
	"context"
	"strings"

	"english-ai-full/logger/core"
)

// ServiceErrorHandler handles service layer errors
type ServiceErrorHandler struct {
	logger *core.CoreLogger
	domain string
}

func NewServiceErrorHandler(logger *core.CoreLogger, domain string) *ServiceErrorHandler {
	return &ServiceErrorHandler{
		logger: logger,
		domain: domain,
	}
}

func (h *ServiceErrorHandler) Handle(err error, operation string, context map[string]interface{}) *AppError {
	if err == nil {
		return nil
	}
	
	// CRITICAL: Check if it's already our AppError - PASS THROUGH UNCHANGED
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, context)
		return appErr // DON'T CREATE NEW ERROR
	}
	
	// Handle service-specific errors
	appErr := h.handleServiceError(err, operation)
	h.logError(appErr, operation, context)
	return appErr
}

func (h *ServiceErrorHandler) handleServiceError(err error, operation string) *AppError {
	errStr := strings.ToLower(err.Error())
	
	// Handle validation errors (from validator package or custom validation)
	if strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid") {
		return NewErrorWithInternalAndMessages(
			ErrValidationFailed,
			"Input validation failed",
			"Xác thực đầu vào thất bại",
			err,
		)
	}
	
	// Handle authentication/authorization errors
	if strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "authentication") {
		return NewErrorWithInternalAndMessages(
			ErrUnauthorized,
			"Authentication required",
			"Yêu cầu xác thực",
			err,
		)
	}
	
	if strings.Contains(errStr, "forbidden") || strings.Contains(errStr, "access denied") {
		return NewErrorWithInternalAndMessages(
			ErrForbidden,
			"Access denied",
			"Truy cập bị từ chối",
			err,
		)
	}
	
	// Handle business logic errors
	if strings.Contains(errStr, "business rule") || strings.Contains(errStr, "business logic") {
		return NewErrorWithInternalAndMessages(
			ErrValidationFailed,
			"Business rule violation",
			"Vi phạm quy tắc kinh doanh",
			err,
		)
	}
	
	// Handle external service errors
	if strings.Contains(errStr, "external") || strings.Contains(errStr, "third party") ||
	   strings.Contains(errStr, "api") {
		return NewErrorWithInternalAndMessages(
			ErrServiceUnavailable,
			"External service error",
			"Lỗi dịch vụ bên ngoài",
			err,
		)
	}
	
	// Handle context errors
	if strings.Contains(errStr, "context") {
		if strings.Contains(errStr, "canceled") || strings.Contains(errStr, "cancelled") {
			return NewErrorWithInternalAndMessages(
				ErrContextCancelled,
				"Operation was cancelled",
				"Thao tác đã bị hủy",
				err,
			)
		}
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") {
			return NewErrorWithInternalAndMessages(
				ErrTimeout,
				"Operation timeout",
				"Thao tác hết thời gian chờ",
				err,
			)
		}
	}
	
	// Handle network/connection errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") ||
	   strings.Contains(errStr, "dial") || strings.Contains(errStr, "refused") {
		return NewErrorWithInternalAndMessages(
			ErrServiceUnavailable,
			"Network connection error",
			"Lỗi kết nối mạng",
			err,
		)
	}
	
	// Handle file/storage errors
	if strings.Contains(errStr, "file not found") || strings.Contains(errStr, "no such file") {
		return NewErrorWithInternalAndMessages(
			ErrAccountNotFound,
			"Resource not found",
			"Không tìm thấy tài nguyên",
			err,
		)
	}
	
	if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access is denied") {
		return NewErrorWithInternalAndMessages(
			ErrForbidden,
			"File access denied",
			"Truy cập tệp bị từ chối",
			err,
		)
	}
	
	// Handle parsing/encoding errors
	if strings.Contains(errStr, "json") || strings.Contains(errStr, "xml") ||
	   strings.Contains(errStr, "parsing") || strings.Contains(errStr, "decode") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"Data format error",
			"Lỗi định dạng dữ liệu",
			err,
		)
	}
	
	// Generic service error - only for non-AppError types
	return NewErrorWithInternalAndMessages(
		ErrSystemError,
		"Service operation failed",
		"Thao tác dịch vụ thất bại",
		err,
	)
}

// HandleWithCustomMessage allows custom error messages while maintaining proper error codes
func (h *ServiceErrorHandler) HandleWithCustomMessage(err error, operation, customMessage, customMessageVN string, context map[string]interface{}) *AppError {
	if err == nil {
		return nil
	}
	
	// Check if it's already our AppError - pass through unchanged
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, context)
		return appErr
	}
	
	// Determine appropriate error code from service error
	appErr := h.handleServiceError(err, operation)
	
	// Override messages with custom ones
	appErr.Message = customMessage
	appErr.MessageVN = customMessageVN
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleContext creates appropriate error from context
func (h *ServiceErrorHandler) HandleContext(ctx context.Context, operation string) *AppError {
	if ctx.Err() == nil {
		return nil
	}
	
	appErr := ContextError(ctx)
	h.logError(appErr, operation, map[string]interface{}{})
	
	return appErr
}

// HandleValidation creates validation error with field details
func (h *ServiceErrorHandler) HandleValidation(field, reason, reasonVN string, operation string) *AppError {
	appErr := NewErrorWithMessages(
		ErrValidationFailed,
		"Validation failed: "+reason,
		"Xác thực thất bại: "+reasonVN,
	)
	
	appErr.Details = map[string]interface{}{
		"field":     field,
		"reason":    reason,
		"reason_vn": reasonVN,
	}
	
	context := map[string]interface{}{
		"field":  field,
		"reason": reason,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleBusinessRule creates business rule violation error
func (h *ServiceErrorHandler) HandleBusinessRule(rule, ruleVN, operation string, details map[string]interface{}) *AppError {
	appErr := NewErrorWithMessages(
		ErrValidationFailed,
		"Business rule violation: "+rule,
		"Vi phạm quy tắc kinh doanh: "+ruleVN,
	)
	
	if details != nil {
		appErr.Details = details
	}
	
	context := map[string]interface{}{
		"business_rule":    rule,
		"business_rule_vn": ruleVN,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleExternalService creates external service error
func (h *ServiceErrorHandler) HandleExternalService(serviceName string, internal error, operation string) *AppError {
	appErr := NewErrorWithInternalAndMessages(
		ErrServiceUnavailable,
		"External service unavailable: "+serviceName,
		"Dịch vụ bên ngoài không khả dụng: "+serviceName,
		internal,
	)
	
	appErr.Details = map[string]interface{}{
		"service": serviceName,
	}
	
	context := map[string]interface{}{
		"external_service": serviceName,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleNotFound creates service-level not found error
func (h *ServiceErrorHandler) HandleNotFound(entity, identifier, operation string) *AppError {
	message := "Resource not found"
	messageVN := "Không tìm thấy tài nguyên"
	
	if entity == "user" || entity == "account" {
		message = "Account not found"
		messageVN = "Không tìm thấy tài khoản"
	}
	
	appErr := NewErrorWithMessages(ErrAccountNotFound, message, messageVN)
	appErr.Details = map[string]interface{}{
		"entity":     entity,
		"identifier": identifier,
	}
	
	context := map[string]interface{}{
		"entity":     entity,
		"identifier": identifier,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

func (h *ServiceErrorHandler) logError(appErr *AppError, operation string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"error_code": appErr.Code,
		"operation":  operation,
		"domain":     h.domain,
		"layer":      "service",
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
	
	h.logger.Error("Service error occurred", logContext)
}
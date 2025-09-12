// error_system/handler_handler.go
package error_system

import (
	"net/http"

	"strings"

	"english-ai-full/logger/core"
)

// HandlerErrorHandler handles HTTP handler layer errors
type HandlerErrorHandler struct {
	logger *core.CoreLogger
	domain string
}

func NewHandlerErrorHandler(logger *core.CoreLogger, domain string) *HandlerErrorHandler {
	return &HandlerErrorHandler{
		logger: logger,
		domain: domain,
	}
}

func (h *HandlerErrorHandler) Handle(err error, operation string, context map[string]interface{}) *AppError {
	if err == nil {
		return nil
	}
	
	// CRITICAL: Check if it's already our AppError - PASS THROUGH UNCHANGED
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, context)
		return appErr // DON'T CREATE NEW ERROR
	}
	
	// Handle HTTP handler-specific errors
	appErr := h.handleHTTPError(err, operation)
	h.logError(appErr, operation, context)
	return appErr
}

func (h *HandlerErrorHandler) handleHTTPError(err error, operation string) *AppError {
	errStr := strings.ToLower(err.Error())
	
	// Handle JSON decode/encode errors
	if strings.Contains(errStr, "json") {
		if strings.Contains(errStr, "decode") || strings.Contains(errStr, "unmarshal") {
			return NewErrorWithInternalAndMessages(
				ErrInvalidInput,
				"Invalid JSON format in request",
				"Định dạng JSON không hợp lệ trong yêu cầu",
				err,
			)
		}
		if strings.Contains(errStr, "encode") || strings.Contains(errStr, "marshal") {
			return NewErrorWithInternalAndMessages(
				ErrSystemError,
				"Response encoding error",
				"Lỗi mã hóa phản hồi",
				err,
			)
		}
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"JSON processing error",
			"Lỗi xử lý JSON",
			err,
		)
	}
	
	// Handle request parsing errors
	if strings.Contains(errStr, "parse") || strings.Contains(errStr, "malformed") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"Request parsing error",
			"Lỗi phân tích yêu cầu",
			err,
		)
	}
	
	// Handle content type errors
	if strings.Contains(errStr, "content-type") || strings.Contains(errStr, "media type") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"Unsupported content type",
			"Loại nội dung không được hỗ trợ",
			err,
		)
	}
	
	// Handle request size errors
	if strings.Contains(errStr, "request too large") || strings.Contains(errStr, "body too large") ||
	   strings.Contains(errStr, "payload too large") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"Request body too large",
			"Nội dung yêu cầu quá lớn",
			err,
		)
	}
	
	// Handle form parsing errors
	if strings.Contains(errStr, "form") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"Form data parsing error",
			"Lỗi phân tích dữ liệu biểu mẫu",
			err,
		)
	}
	
	// Handle multipart/file upload errors
	if strings.Contains(errStr, "multipart") || strings.Contains(errStr, "file upload") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"File upload error",
			"Lỗi tải lên tệp",
			err,
		)
	}
	
	// Handle URL/path errors
	if strings.Contains(errStr, "url") || strings.Contains(errStr, "path") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"Invalid URL or path",
			"URL hoặc đường dẫn không hợp lệ",
			err,
		)
	}
	
	// Handle query parameter errors
	if strings.Contains(errStr, "query") || strings.Contains(errStr, "parameter") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"Invalid query parameters",
			"Tham số truy vấn không hợp lệ",
			err,
		)
	}
	
	// Handle header errors
	if strings.Contains(errStr, "header") {
		return NewErrorWithInternalAndMessages(
			ErrInvalidInput,
			"Invalid request headers",
			"Tiêu đề yêu cầu không hợp lệ",
			err,
		)
	}
	
	// Handle authentication header errors
	if strings.Contains(errStr, "authorization") || strings.Contains(errStr, "bearer") ||
	   strings.Contains(errStr, "token") {
		return NewErrorWithInternalAndMessages(
			ErrUnauthorized,
			"Invalid authentication token",
			"Token xác thực không hợp lệ",
			err,
		)
	}
	
	// Handle request timeout errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") {
		return NewErrorWithInternalAndMessages(
			ErrTimeout,
			"Request processing timeout",
			"Hết thời gian xử lý yêu cầu",
			err,
		)
	}
	
	// Handle connection errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "broken pipe") ||
	   strings.Contains(errStr, "reset by peer") {
		return NewErrorWithInternalAndMessages(
			ErrServiceUnavailable,
			"Connection error",
			"Lỗi kết nối",
			err,
		)
	}
	
	// Generic handler error - only for non-AppError types
	return NewErrorWithInternalAndMessages(
		ErrSystemError,
		"Request processing failed",
		"Xử lý yêu cầu thất bại",
		err,
	)
}

// HandleWithCustomMessage allows custom error messages while maintaining proper error codes
func (h *HandlerErrorHandler) HandleWithCustomMessage(err error, operation, customMessage, customMessageVN string, context map[string]interface{}) *AppError {
	if err == nil {
		return nil
	}
	
	// Check if it's already our AppError - pass through unchanged
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, context)
		return appErr
	}
	
	// Determine appropriate error code from handler error
	appErr := h.handleHTTPError(err, operation)
	
	// Override messages with custom ones
	appErr.Message = customMessage
	appErr.MessageVN = customMessageVN
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleJSONError creates specific JSON parsing error
func (h *HandlerErrorHandler) HandleJSONError(field string, operation string, internal error) *AppError {
	appErr := NewErrorWithInternalAndMessages(
		ErrInvalidInput,
		"Invalid JSON format",
		"Định dạng JSON không hợp lệ",
		internal,
	)
	
	if field != "" {
		appErr.Details = map[string]interface{}{
			"field": field,
		}
	}
	
	context := map[string]interface{}{
		"field": field,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleValidationError creates validation error for request data
func (h *HandlerErrorHandler) HandleValidationError(field, reason, reasonVN string, operation string) *AppError {
	appErr := NewErrorWithMessages(
		ErrValidationFailed,
		"Request validation failed: "+reason,
		"Xác thực yêu cầu thất bại: "+reasonVN,
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

// HandleMissingHeader creates missing header error
func (h *HandlerErrorHandler) HandleMissingHeader(headerName, operation string) *AppError {
	appErr := NewErrorWithMessages(
		ErrMissingField,
		"Missing required header: "+headerName,
		"Thiếu tiêu đề bắt buộc: "+headerName,
	)
	
	appErr.Details = map[string]interface{}{
		"header": headerName,
	}
	
	context := map[string]interface{}{
		"missing_header": headerName,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleInvalidToken creates invalid authentication token error
func (h *HandlerErrorHandler) HandleInvalidToken(operation string, internal error) *AppError {
	appErr := NewErrorWithInternalAndMessages(
		ErrUnauthorized,
		"Invalid or expired authentication token",
		"Token xác thực không hợp lệ hoặc đã hết hạn",
		internal,
	)
	
	h.logError(appErr, operation, map[string]interface{}{})
	
	return appErr
}

// HandleRequestTooLarge creates request too large error
func (h *HandlerErrorHandler) HandleRequestTooLarge(maxSize int64, operation string) *AppError {
	appErr := NewErrorWithMessages(
		ErrInvalidInput,
		"Request body exceeds maximum size limit",
		"Nội dung yêu cầu vượt quá giới hạn kích thước tối đa",
	)
	
	appErr.Details = map[string]interface{}{
		"max_size_bytes": maxSize,
		"max_size_mb":    maxSize / (1024 * 1024),
	}
	
	context := map[string]interface{}{
		"max_size": maxSize,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleMethodNotAllowed creates method not allowed error
func (h *HandlerErrorHandler) HandleMethodNotAllowed(method, operation string, allowedMethods []string) *AppError {
	appErr := NewErrorWithMessages(
		ErrInvalidInput,
		"HTTP method not allowed: "+method,
		"Phương thức HTTP không được phép: "+method,
	)
	
	appErr.HTTPStatus = http.StatusMethodNotAllowed
	appErr.Details = map[string]interface{}{
		"method":          method,
		"allowed_methods": allowedMethods,
	}
	
	context := map[string]interface{}{
		"method":          method,
		"allowed_methods": allowedMethods,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

// HandleRateLimit creates rate limit exceeded error
func (h *HandlerErrorHandler) HandleRateLimit(operation string, limit int, window string) *AppError {
	appErr := NewErrorWithMessages(
		ErrServiceUnavailable,
		"Rate limit exceeded",
		"Vượt quá giới hạn tốc độ",
	)
	
	appErr.HTTPStatus = http.StatusTooManyRequests
	appErr.Details = map[string]interface{}{
		"limit":  limit,
		"window": window,
	}
	
	context := map[string]interface{}{
		"rate_limit": limit,
		"window":     window,
	}
	
	h.logError(appErr, operation, context)
	
	return appErr
}

func (h *HandlerErrorHandler) logError(appErr *AppError, operation string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"error_code": appErr.Code,
		"operation":  operation,
		"domain":     h.domain,
		"layer":      "handler",
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
	
	h.logger.Error("Handler error occurred", logContext)
}
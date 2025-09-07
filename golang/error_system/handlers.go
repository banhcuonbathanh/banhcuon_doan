// error_system/handlers.go
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
	
	// Check if it's already our AppError
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, table, context)
		return appErr
	}
	
	// Check if it's a context error
	// if err == context.Canceled || err == context.DeadlineExceeded {
	// 	appErr := ContextError(context.Background())
	// 	h.logError(appErr, operation, table, context)
	// 	return appErr
	// }
	
	// Handle database-specific errors
	appErr := h.handleDatabaseError(err, operation, table)
	h.logError(appErr, operation, table, context)
	
	return appErr
}

func (h *RepositoryErrorHandler) handleDatabaseError(err error, operation, table string) *AppError {
	errStr := strings.ToLower(err.Error())
	
	// Check for duplicate key violations
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique constraint") {
		if strings.Contains(errStr, "email") {
			return NewErrorWithInternal(ErrAccountDuplicate, "Account with this email already exists", err)
		}
		return NewErrorWithInternal(ErrValidationFailed, "Duplicate entry detected", err)
	}
	
	// Check for constraint violations
	if strings.Contains(errStr, "constraint") {
		return NewErrorWithInternal(ErrValidationFailed, "Data constraint violation", err)
	}
	
	// Check for connection/network errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") {
		return NewErrorWithInternal(ErrServiceUnavailable, "Database connection error", err)
	}
	
	// Check for timeout errors
	if strings.Contains(errStr, "timeout") {
		return NewErrorWithInternal(ErrTimeout, "Database operation timeout", err)
	}
	
	// Generic database error
	return NewErrorWithInternal(ErrDatabaseError, "Database operation failed", err)
}

func (h *RepositoryErrorHandler) logError(appErr *AppError, operation, table string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"error_code": appErr.Code,
		"operation":  operation,
		"table":      table,
		"domain":     h.domain,
		"layer":      "repository",
	}
	
	// Merge additional context
	for k, v := range context {
		logContext[k] = v
	}
	
	// Add internal error details for logging (not for client)
	if appErr.Internal != nil {
		logContext["internal_error"] = appErr.Internal.Error()
	}
	
	h.logger.Error("Repository error occurred", logContext)
}

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
	
	// Check if it's already our AppError - just pass it through
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, context)
		return appErr
	}
	
	// Check for context errors
	// if err == context.Canceled || err == context.DeadlineExceeded {
	// 	appErr := ContextError(context.Background())
	// 	h.logError(appErr, operation, context)
	// 	return appErr
	// }
	
	// Handle validation errors (from validator package)
	if strings.Contains(err.Error(), "validation") {
		appErr := NewErrorWithInternal(ErrValidationFailed, "Input validation failed", err)
		h.logError(appErr, operation, context)
		return appErr
	}
	
	// Generic service error
	appErr := NewErrorWithInternal(ErrSystemError, "Service operation failed", err)
	h.logError(appErr, operation, context)
	return appErr
}

func (h *ServiceErrorHandler) logError(appErr *AppError, operation string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"error_code": appErr.Code,
		"operation":  operation,
		"domain":     h.domain,
		"layer":      "service",
	}
	
	for k, v := range context {
		logContext[k] = v
	}
	
	if appErr.Internal != nil {
		logContext["internal_error"] = appErr.Internal.Error()
	}
	
	h.logger.Error("Service error occurred", logContext)
}

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
	
	// Check if it's already our AppError - just pass it through
	if appErr, ok := IsAppError(err); ok {
		h.logError(appErr, operation, context)
		return appErr
	}
	
	// Handle JSON decode errors
	if strings.Contains(err.Error(), "json") || strings.Contains(err.Error(), "decode") {
		appErr := NewErrorWithInternal(ErrInvalidInput, "Invalid JSON format", err)
		h.logError(appErr, operation, context)
		return appErr
	}
	
	// Generic handler error
	appErr := NewErrorWithInternal(ErrSystemError, "Request processing failed", err)
	h.logError(appErr, operation, context)
	return appErr
}

func (h *HandlerErrorHandler) logError(appErr *AppError, operation string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"error_code": appErr.Code,
		"operation":  operation,
		"domain":     h.domain,
		"layer":      "handler",
	}
	
	for k, v := range context {
		logContext[k] = v
	}
	
	if appErr.Internal != nil {
		logContext["internal_error"] = appErr.Internal.Error()
	}
	
	h.logger.Error("Handler error occurred", logContext)
}


	// Check for context errors
	// if err == context.Canceled || err == context.DeadlineExceeded {
	// 	appErr := ContextError(context.Background())
	// 	h.logError(appErr, operation, context)
	// 	return appErr
	// }
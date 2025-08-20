
// ============================================================================
// FILE: golang/internal/error_custom/layer/service_errors.go
// ============================================================================
package errorcustom

import (
	"context"



	"english-ai-full/logger"
)


type ServiceErrorManager struct{}

// NewServiceErrorManager creates a new service error manager
func NewServiceErrorManager() *ServiceErrorManager {
	return &ServiceErrorManager{}
}

// ============================================================================
// BUSINESS LOGIC ERROR HANDLING
// ============================================================================




// ============================================================================
// TRANSACTION ERROR HANDLING
// ============================================================================






// ============================================================================
// HELPER METHODS
// ============================================================================

// logServiceError logs service errors with appropriate severity
func (s *ServiceErrorManager) logServiceError(apiErr * APIError, operation string) {
	logContext := apiErr.GetLogContext()
	logContext["operation"] = operation
	logContext["layer"] = "service"

	if apiErr.HTTPStatus >= 500 {
		logger.Error("Service layer error", logContext)
	} else if apiErr.HTTPStatus >= 400 {
		logger.Warning("Service layer warning", logContext)
	} else {
		logger.Info("Service layer info", logContext)
	}
}








// new 12121212

// WrapRepositoryError wraps repository errors with service layer context
func (s *ServiceErrorManager) WrapRepositoryError(err error, domain, operation string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// Convert to API error if not already
	apiErr := ConvertToAPIError(err)
	if apiErr == nil {
		apiErr = NewAPIError(
			GetSystemErrorCode(domain),
			"Service operation failed",
			500,
		)
	}

	// Add service layer context
	apiErr.WithDomain(domain).
		WithLayer("service").
		WithOperation(operation).
		WithCause(err)

	// Add additional context
	for k, v := range context {
		apiErr.WithDetail(k, v)
	}

	// Log the error
	s.logServiceError(apiErr, operation)

	return apiErr // Return as error interface
}

// HandleBusinessRuleViolation creates a business logic error
func (s *ServiceErrorManager) HandleBusinessRuleViolation(domain, rule, description string, context map[string]interface{}) error {
	businessErr := NewBusinessLogicErrorWithContext(domain, rule, description, context)
	apiErr := businessErr.ToAPIError().
		WithLayer("service")

	logger.Warning("Business rule violation", map[string]interface{}{
		"domain":      domain,
		"rule":        rule,
		"description": description,
		"context":     context,
		"layer":       "service",
	})

	return apiErr
}

// HandleExternalServiceError handles errors from external services
func (s *ServiceErrorManager) HandleExternalServiceError(err error, domain, service, operation string, retryable bool) error {
	extServiceErr := NewExternalServiceError(domain, service, operation, "External service error", err, retryable)
	apiErr := extServiceErr.ToAPIError().
		WithLayer("service")

	// Log external service errors
	logger.Error("External service error", map[string]interface{}{
		"domain":    domain,
		"service":   service,
		"operation": operation,
		"retryable": retryable,
		"error":     err.Error(),
		"layer":     "service",
	})

	return apiErr
}

// HandleTransactionError handles database transaction errors
func (s *ServiceErrorManager) HandleTransactionError(err error, domain, operation string) error {
	if err == nil {
		return nil
	}

	systemErr := NewSystemError(domain, "database_transaction", operation, "Transaction failed", err)
	apiErr := systemErr.ToAPIError().
		WithLayer("service")

	logger.Error("Database transaction error", map[string]interface{}{
		"domain":    domain,
		"operation": operation,
		"error":     err.Error(),
		"layer":     "service",
	})

	return apiErr
}

// HandleContextError handles context-related errors (timeouts, cancellations)
func (s *ServiceErrorManager) HandleContextError(ctx context.Context, domain, operation string) error {
	err := ctx.Err()
	if err == nil {
		return nil
	}

	switch err {
	case context.DeadlineExceeded:
		timeoutErr := NewAPIError(
			GetTimeoutCode(domain),
			"Operation timed out",
			408,
		).WithDomain(domain).
			WithLayer("service").
			WithOperation(operation).
			WithRetryable(true)

		logger.Warning("Service operation timeout", map[string]interface{}{
			"domain":    domain,
			"operation": operation,
			"layer":     "service",
		})

		return timeoutErr

	case context.Canceled:
		cancelErr := NewAPIError(
			GetSystemErrorCode(domain),
			"Operation was cancelled",
			499,
		).WithDomain(domain).
			WithLayer("service").
			WithOperation(operation)

		logger.Info("Service operation cancelled", map[string]interface{}{
			"domain":    domain,
			"operation": operation,
			"layer":     "service",
		})

		return cancelErr

	default:
		systemErr := NewSystemError(domain, "context", operation, "Context error", err)
		return systemErr.ToAPIError().WithLayer("service")
	}
}

// ValidateBusinessRules validates business rules and returns collected errors
func (s *ServiceErrorManager) ValidateBusinessRules(domain string, validations map[string]func() error) error {
	errorCollection := NewErrorCollection(domain)

	for ruleName, validationFunc := range validations {
		if err := validationFunc(); err != nil {
			if businessErr, ok := err.(*BusinessLogicError); ok {
				errorCollection.Add(businessErr)
			} else {
				// Convert generic error to business logic error
				businessErr := NewBusinessLogicError(domain, ruleName, err.Error())
				errorCollection.Add(businessErr)
			}
		}
	}

	if errorCollection.HasErrors() {
		apiErr := errorCollection.ToAPIError()
		if apiErr != nil {
			apiErr.WithLayer("service")
		}
		return apiErr
	}

	return nil
}

// new 1212121212
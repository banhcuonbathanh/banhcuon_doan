
// ============================================================================
// FILE: golang/internal/error_custom/layer/repository_errors.go
// ============================================================================
package errorcustom

import (
	"database/sql"
	"fmt"
	"strings"


	"english-ai-full/logger"
)

// RepositoryErrorManager manages data access layer errors
type RepositoryErrorManager struct{}

// NewRepositoryErrorManager creates a new repository error manager
func NewRepositoryErrorManager() *RepositoryErrorManager {
	return &RepositoryErrorManager{}
}

// ============================================================================
// DATABASE ERROR HANDLING
// ============================================================================


// ============================================================================
// ERROR TYPE DETECTION HELPERS
// ============================================================================

// isConstraintViolation checks if error is a constraint violation
func (r *RepositoryErrorManager) isConstraintViolation(err error) bool {
	errMsg := strings.ToLower(err.Error())
	constraintKeywords := []string{
		"unique constraint",
		"duplicate key",
		"foreign key constraint",
		"check constraint",
		"unique_violation",
		"foreign_key_violation",
		"check_violation",
	}

	for _, keyword := range constraintKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// isConnectionError checks if error is a connection error
func (r *RepositoryErrorManager) isConnectionError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	connectionKeywords := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"no connection",
		"database is locked",
		"server has gone away",
	}

	for _, keyword := range connectionKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// isDeadlockError checks if error is a deadlock
func (r *RepositoryErrorManager) isDeadlockError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "deadlock")
}

// isTimeoutError checks if error is a timeout
func (r *RepositoryErrorManager) isTimeoutError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	timeoutKeywords := []string{
		"timeout",
		"timed out",
		"context deadline exceeded",
	}

	for _, keyword := range timeoutKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// extractFieldFromConstraintError attempts to extract field name from constraint error
func (r *RepositoryErrorManager) extractFieldFromConstraintError(err error) string {
	errMsg := err.Error()
	
	// Try to extract field name from common error patterns
	// This is database-specific and may need adjustment
	if strings.Contains(errMsg, "UNIQUE constraint failed:") {
		parts := strings.Split(errMsg, ":")
		if len(parts) > 1 {
			fieldPart := strings.TrimSpace(parts[1])
			if strings.Contains(fieldPart, ".") {
				fieldParts := strings.Split(fieldPart, ".")
				return fieldParts[len(fieldParts)-1]
			}
			return fieldPart
		}
	}
	
	return "unknown_field"
}

// extractValueFromContext extracts field value from context
func (r *RepositoryErrorManager) extractValueFromContext(context map[string]interface{}, field string) interface{} {
	if context == nil {
		return nil
	}
	
	if value, exists := context[field]; exists {
		return value
	}
	
	// Try common variations
	variations := []string{
		field,
		strings.ToLower(field),
		strings.ToUpper(field),
		strings.Title(field),
	}
	
	for _, variation := range variations {
		if value, exists := context[variation]; exists {
			return value
		}
	}
	
	return nil
}


// new 1212121212

// ============================================================================
// REPOSITORY ERROR MANAGER - Updated to return error
// ============================================================================

// HandleDatabaseError handles various database errors
func (r *RepositoryErrorManager) HandleDatabaseError(err error, domain, table, operation string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// Handle specific database errors
	switch {
	case err == sql.ErrNoRows:
		return r.handleNoRowsError(domain, table, operation, context)
	case r.isConstraintViolation(err):
		return r.handleConstraintViolation(err, domain, table, operation, context)
	case r.isConnectionError(err):
		return r.handleConnectionError(err, domain, table, operation)
	case r.isDeadlockError(err):
		return r.handleDeadlockError(err, domain, table, operation)
	case r.isTimeoutError(err):
		return r.handleTimeoutError(err, domain, table, operation)
	default:
		return r.handleGenericDatabaseError(err, domain, table, operation, context)
	}
}

// Updated helper methods to return error instead of *APIError
func (r *RepositoryErrorManager) handleNoRowsError(domain, table, operation string, context map[string]interface{}) error {
	resourceType := strings.TrimSuffix(table, "s") // Remove plural 's'
	
	notFoundErr := NewNotFoundError(domain, resourceType, nil)
	if len(context) > 0 {
		notFoundErr.Identifiers = context
	}

	apiErr := notFoundErr.ToAPIError().
		WithLayer("repository").
		WithOperation(operation).
		WithDetail("table", table)

	logger.Info("No rows found in database", map[string]interface{}{
		"domain":    domain,
		"table":     table,
		"operation": operation,
		"context":   context,
		"layer":     "repository",
	})

	return apiErr
}

func (r *RepositoryErrorManager) handleConstraintViolation(err error, domain, table, operation string, context map[string]interface{}) error {
	errMsg := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errMsg, "unique") || strings.Contains(errMsg, "duplicate"):
		// Extract field name from error message if possible
		field := r.extractFieldFromConstraintError(err)
		value := r.extractValueFromContext(context, field)
		
		duplicateErr := NewDuplicateError(domain, table, field, value)
		apiErr := duplicateErr.ToAPIError().
			WithLayer("repository").
			WithOperation(operation).
			WithDetail("table", table).
			WithCause(err)

		logger.Warn("Database constraint violation - duplicate", map[string]interface{}{
			"domain":    domain,
			"table":     table,
			"operation": operation,
			"field":     field,
			"error":     err.Error(),
			"layer":     "repository",
		})

		return apiErr

	case strings.Contains(errMsg, "foreign key"):
		businessErr := NewBusinessLogicError(
			domain,
			"foreign_key_constraint",
			"Referenced record does not exist or cannot be deleted due to dependencies",
		)
		
		apiErr := businessErr.ToAPIError().
			WithLayer("repository").
			WithOperation(operation).
			WithDetail("table", table).
			WithDetail("constraint_type", "foreign_key").
			WithCause(err)

		logger.Warn("Database constraint violation - foreign key", map[string]interface{}{
			"domain":    domain,
			"table":     table,
			"operation": operation,
			"error":     err.Error(),
			"layer":     "repository",
		})

		return apiErr

	case strings.Contains(errMsg, "check"):
		validationErr := NewValidationError(
			domain,
			"check_constraint",
			"Data violates database check constraint",
			nil,
		)
		
		apiErr := validationErr.ToAPIError().
			WithLayer("repository").
			WithOperation(operation).
			WithDetail("table", table).
			WithDetail("constraint_type", "check").
			WithCause(err)

		return apiErr

	default:
		systemErr := NewSystemError(
			domain,
			"database",
			operation,
			"Database constraint violation",
			err,
		)
		
		return systemErr.ToAPIError().
			WithLayer("repository").
			WithDetail("table", table)
	}
}

func (r *RepositoryErrorManager) handleConnectionError(err error, domain, table, operation string) error {
	serviceErr := NewExternalServiceError(
		domain,
		"database",
		operation,
		"Database connection failed",
		err,
		true, // retryable
	)

	apiErr := serviceErr.ToAPIError().
		WithLayer("repository").
		WithDetail("table", table)

	logger.Error("Database connection error", map[string]interface{}{
		"domain":    domain,
		"table":     table,
		"operation": operation,
		"error":     err.Error(),
		"retryable": true,
		"layer":     "repository",
	})

	return apiErr
}

func (r *RepositoryErrorManager) handleDeadlockError(err error, domain, table, operation string) error {
	systemErr := NewSystemError(
		domain,
		"database",
		operation,
		"Database deadlock detected",
		err,
	)

	apiErr := systemErr.ToAPIError().
		WithLayer("repository").
		WithDetail("table", table).
		WithDetail("deadlock", true).
		WithRetryable(true)

	logger.Warn("Database deadlock detected", map[string]interface{}{
		"domain":    domain,
		"table":     table,
		"operation": operation,
		"error":     err.Error(),
		"retryable": true,
		"layer":     "repository",
	})

	return apiErr
}

func (r *RepositoryErrorManager) handleTimeoutError(err error, domain, table, operation string) error {
	timeoutErr := NewAPIError(
		GetTimeoutCode(domain),
		"Database operation timed out",
		408,
	).WithDomain(domain).
		WithLayer("repository").
		WithOperation(operation).
		WithDetail("table", table).
		WithRetryable(true).
		WithCause(err)

	logger.Warn("Database operation timeout", map[string]interface{}{
		"domain":    domain,
		"table":     table,
		"operation": operation,
		"error":     err.Error(),
		"retryable": true,
		"layer":     "repository",
	})

	return timeoutErr
}

func (r *RepositoryErrorManager) handleGenericDatabaseError(err error, domain, table, operation string, context map[string]interface{}) error {
	systemErr := NewSystemError(
		domain,
		"database",
		operation,
		fmt.Sprintf("Database operation failed on table '%s'", table),
		err,
	)

	apiErr := systemErr.ToAPIError().
		WithLayer("repository").
		WithDetail("table", table)

	// Add context details
	for k, v := range context {
		apiErr.WithDetail(k, v)
	}

	logger.Error("Generic database error", map[string]interface{}{
		"domain":    domain,
		"table":     table,
		"operation": operation,
		"context":   context,
		"error":     err.Error(),
		"layer":     "repository",
	})

	return apiErr
}

// new 1212121212
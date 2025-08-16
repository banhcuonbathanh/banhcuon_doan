// internal/error_custom/constructors.go
// Domain-aware error constructors
package errorcustom

import "fmt"

// ============================================================================
// GENERIC ERROR CONSTRUCTORS
// ============================================================================

// NewNotFoundError creates a not found error for any domain
func NewNotFoundError(domain, resourceType string, resourceID interface{}) *NotFoundError {
	return &NotFoundError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeNotFound,
		},
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// NewNotFoundErrorWithIdentifiers creates a not found error with multiple identifiers
func NewNotFoundErrorWithIdentifiers(domain, resourceType string, identifiers map[string]interface{}) *NotFoundError {
	return &NotFoundError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeNotFound,
		},
		ResourceType: resourceType,
		Identifiers:  identifiers,
	}
}

// NewValidationError creates a validation error for any domain
func NewValidationError(domain, field, message string, value interface{}) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeValidation,
		},
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// NewValidationErrorWithRules creates a validation error with rule details
func NewValidationErrorWithRules(domain, field, message string, value interface{}, rules map[string]interface{}) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeValidation,
		},
		Field:   field,
		Message: message,
		Value:   value,
		Rules:   rules,
	}
}

// NewDuplicateError creates a duplicate resource error for any domain
func NewDuplicateError(domain, resourceType, field string, value interface{}) *DuplicateError {
	return &DuplicateError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeDuplicate,
		},
		ResourceType: resourceType,
		Field:        field,
		Value:        value,
	}
}

// NewAuthenticationError creates an authentication error for any domain
func NewAuthenticationError(domain, reason string) *AuthenticationError {
	return &AuthenticationError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeAuthentication,
		},
		Reason: reason,
	}
}

// NewAuthenticationErrorWithStep creates an authentication error with step info
func NewAuthenticationErrorWithStep(domain, reason, step string, context map[string]interface{}) *AuthenticationError {
	return &AuthenticationError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeAuthentication,
		},
		Reason:  reason,
		Step:    step,
		Context: context,
	}
}

// NewAuthorizationError creates an authorization error for any domain
func NewAuthorizationError(domain, action, resource string) *AuthorizationError {
	return &AuthorizationError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeAuthorization,
		},
		Action:   action,
		Resource: resource,
	}
}

// NewAuthorizationErrorWithContext creates an authorization error with context
func NewAuthorizationErrorWithContext(domain, action, resource string, context map[string]interface{}) *AuthorizationError {
	return &AuthorizationError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeAuthorization,
		},
		Action:   action,
		Resource: resource,
		Context:  context,
	}
}

// NewBusinessLogicError creates a business logic error for any domain
func NewBusinessLogicError(domain, rule, description string) *BusinessLogicError {
	return &BusinessLogicError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeBusinessLogic,
		},
		Rule:        rule,
		Description: description,
	}
}

// NewBusinessLogicErrorWithContext creates a business logic error with context
func NewBusinessLogicErrorWithContext(domain, rule, description string, context map[string]interface{}) *BusinessLogicError {
	return &BusinessLogicError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeBusinessLogic,
		},
		Rule:        rule,
		Description: description,
		Context:     context,
	}
}

// NewExternalServiceError creates an external service error for any domain
func NewExternalServiceError(domain, service, operation, message string, cause error, retryable bool) *ExternalServiceError {
	return &ExternalServiceError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeExternalService,
		},
		Service:   service,
		Operation: operation,
		Message:   message,
		Cause:     cause,
		Retryable: retryable,
	}
}

// NewSystemError creates a system error for any domain
func NewSystemError(domain, component, operation, message string, cause error) *SystemError {
	return &SystemError{
		BaseError: BaseError{
			Domain:    domain,
			ErrorType: ErrorTypeSystem,
		},
		Component: component,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// ============================================================================
// USER DOMAIN ERROR CONSTRUCTORS (for backward compatibility)
// ============================================================================

// NewUserNotFoundByID creates a user not found error with ID
func NewUserNotFoundByID(id int64) *NotFoundError {
	return NewNotFoundError(DomainAccount, "user", id)
}

// NewUserNotFoundByEmail creates a user not found error with email
func NewUserNotFoundByEmail(email string) *NotFoundError {
	return NewNotFoundErrorWithIdentifiers(DomainAccount, "user", map[string]interface{}{
		"email": email,
	})
}

// NewEmailNotFoundError creates an authentication error for email not found
func NewEmailNotFoundError(email string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAccount,
		"email not found",
		"email_check",
		map[string]interface{}{
			"email":      email,
			"user_found": false,
		},
	)
}

// NewPasswordMismatchError creates an authentication error for password mismatch
func NewPasswordMismatchError(email string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAccount,
		"password mismatch",
		"password_check",
		map[string]interface{}{
			"email":      email,
			"user_found": true,
		},
	)
}

// NewAccountDisabledError creates an authentication error for disabled account
func NewAccountDisabledError(email string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAccount,
		"account disabled",
		"status_check",
		map[string]interface{}{
			"email":      email,
			"user_found": true,
		},
	)
}

// NewAccountLockedError creates an authentication error for locked account
func NewAccountLockedError(email string, lockReason string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAccount,
		fmt.Sprintf("account locked: %s", lockReason),
		"status_check",
		map[string]interface{}{
			"email":       email,
			"user_found":  true,
			"lock_reason": lockReason,
		},
	)
}

// NewDuplicateEmailError creates a duplicate email error
func NewDuplicateEmailError(email string) *DuplicateError {
	return NewDuplicateError(DomainAccount, "user", "email", email)
}

// NewWeakPasswordError creates a password validation error
func NewWeakPasswordError(requirements []string) *ValidationError {
	return NewValidationErrorWithRules(
		DomainAccount,
		"password",
		"Password does not meet security requirements",
		"[REDACTED]",
		map[string]interface{}{
			"requirements": requirements,
		},
	)
}












// ============================================================================
// ADMIN DOMAIN ERROR CONSTRUCTORS (example for new domain)
// ============================================================================

// NewInsufficientPrivilegesError creates an authorization error for admin operations
func NewInsufficientPrivilegesError(userID int64, requiredRole, currentRole string) *AuthorizationError {
	return NewAuthorizationErrorWithContext(
		DomainAdmin,
		"admin_operation",
		"system",
		map[string]interface{}{
			"user_id":       userID,
			"required_role": requiredRole,
			"current_role":  currentRole,
		},
	)
}

// NewBulkOperationLimitError creates a business logic error for bulk operation limits
func NewBulkOperationLimitError(operation string, requested, maxAllowed int) *BusinessLogicError {
	return NewBusinessLogicErrorWithContext(
		DomainAdmin,
		"bulk_operation_limit",
		fmt.Sprintf("Bulk %s operation exceeds maximum limit", operation),
		map[string]interface{}{
			"operation":   operation,
			"requested":   requested,
			"max_allowed": maxAllowed,
		},
	)
}

// ============================================================================
// SYSTEM DOMAIN ERROR CONSTRUCTORS
// ============================================================================

// NewDatabaseError creates a system error for database operations
func NewDatabaseError(operation, table string, cause error) *SystemError {
	return NewSystemError(
		DomainSystem,
		"database",
		operation,
		fmt.Sprintf("Database operation failed on table '%s'", table),
		cause,
	)
}

// NewCacheError creates a system error for cache operations
func NewCacheError(operation, key string, cause error) *SystemError {
	return NewSystemError(
		DomainSystem,
		"cache",
		operation,
		fmt.Sprintf("Cache operation failed for key '%s'", key),
		cause,
	)
}

// NewFileSystemError creates a system error for file system operations
func NewFileSystemError(operation, path string, cause error) *SystemError {
	return NewSystemError(
		DomainSystem,
		"filesystem",
		operation,
		fmt.Sprintf("File system operation failed for path '%s'", path),
		cause,
	)
}

// ============================================================================
// LEGACY CONSTRUCTORS (for backward compatibility)
// ============================================================================

// Legacy constructors that map to the new system
type UserNotFoundError = NotFoundError
type DuplicateEmailError = DuplicateError
type PasswordValidationError = ValidationError
type ServiceError = ExternalServiceError
type RepositoryError = SystemError

// NewServiceError creates a legacy service error (maps to ExternalServiceError)
func NewServiceError(service, method, message string, cause error, retryable bool) *ExternalServiceError {
	return NewExternalServiceError("", service, method, message, cause, retryable)
}

// NewRepositoryError creates a legacy repository error (maps to SystemError)
func NewRepositoryError(operation, table, message string, cause error) *SystemError {
	return NewSystemError(DomainSystem, "repository", operation, fmt.Sprintf("%s on table %s", message, table), cause)
}


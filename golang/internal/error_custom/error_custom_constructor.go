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
	return NewNotFoundError(DomainUser, "user", id)
}

// NewUserNotFoundByEmail creates a user not found error with email
func NewUserNotFoundByEmail(email string) *NotFoundError {
	return NewNotFoundErrorWithIdentifiers(DomainUser, "user", map[string]interface{}{
		"email": email,
	})
}

// NewEmailNotFoundError creates an authentication error for email not found
func NewEmailNotFoundError(email string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainUser,
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
		DomainUser,
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
		DomainUser,
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
		DomainUser,
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
	return NewDuplicateError(DomainUser, "user", "email", email)
}

// NewWeakPasswordError creates a password validation error
func NewWeakPasswordError(requirements []string) *ValidationError {
	return NewValidationErrorWithRules(
		DomainUser,
		"password",
		"Password does not meet security requirements",
		"[REDACTED]",
		map[string]interface{}{
			"requirements": requirements,
		},
	)
}

// ============================================================================
// COURSE DOMAIN ERROR CONSTRUCTORS (example for new domain)
// ============================================================================

// NewCourseNotFoundError creates a course not found error
func NewCourseNotFoundError(courseID int64) *NotFoundError {
	return NewNotFoundError(DomainCourse, "course", courseID)
}

// NewCourseNotFoundBySlugError creates a course not found error by slug
func NewCourseNotFoundBySlugError(slug string) *NotFoundError {
	return NewNotFoundErrorWithIdentifiers(DomainCourse, "course", map[string]interface{}{
		"slug": slug,
	})
}

// NewCourseAccessDeniedError creates a course access denied error
func NewCourseAccessDeniedError(userID, courseID int64) *AuthorizationError {
	return NewAuthorizationErrorWithContext(
		DomainCourse,
		"access",
		"course",
		map[string]interface{}{
			"user_id":   userID,
			"course_id": courseID,
		},
	)
}

// NewCourseEnrollmentClosedError creates a business logic error for closed enrollment
func NewCourseEnrollmentClosedError(courseID int64) *BusinessLogicError {
	return NewBusinessLogicErrorWithContext(
		DomainCourse,
		"enrollment_period",
		"Course enrollment is closed",
		map[string]interface{}{
			"course_id": courseID,
		},
	)
}

// NewCourseCapacityExceededError creates a business logic error for capacity exceeded
func NewCourseCapacityExceededError(courseID int64, maxCapacity, currentEnrollment int) *BusinessLogicError {
	return NewBusinessLogicErrorWithContext(
		DomainCourse,
		"enrollment_capacity",
		"Course has reached maximum capacity",
		map[string]interface{}{
			"course_id":          courseID,
			"max_capacity":       maxCapacity,
			"current_enrollment": currentEnrollment,
		},
	)
}

// ============================================================================
// PAYMENT DOMAIN ERROR CONSTRUCTORS (example for new domain)
// ============================================================================

// NewPaymentNotFoundError creates a payment not found error
func NewPaymentNotFoundError(paymentID string) *NotFoundError {
	return NewNotFoundError(DomainPayment, "payment", paymentID)
}

// NewInsufficientFundsError creates a business logic error for insufficient funds
func NewInsufficientFundsError(userID int64, required, available float64) *BusinessLogicError {
	return NewBusinessLogicErrorWithContext(
		DomainPayment,
		"sufficient_funds",
		"Insufficient funds for this transaction",
		map[string]interface{}{
			"user_id":   userID,
			"required":  required,
			"available": available,
		},
	)
}

// NewPaymentProviderError creates an external service error for payment providers
func NewPaymentProviderError(provider, operation string, cause error, retryable bool) *ExternalServiceError {
	return NewExternalServiceError(
		DomainPayment,
		provider,
		operation,
		"Payment provider service error",
		cause,
		retryable,
	)
}

// NewPaymentExpiredError creates a business logic error for expired payments
func NewPaymentExpiredError(paymentID string) *BusinessLogicError {
	return NewBusinessLogicErrorWithContext(
		DomainPayment,
		"payment_expiry",
		"Payment session has expired",
		map[string]interface{}{
			"payment_id": paymentID,
		},
	)
}

// ============================================================================
// CONTENT DOMAIN ERROR CONSTRUCTORS (example for new domain)
// ============================================================================

// NewContentNotFoundError creates a content not found error
func NewContentNotFoundError(contentID int64) *NotFoundError {
	return NewNotFoundError(DomainContent, "content", contentID)
}

// NewContentAccessDeniedError creates a content access denied error
func NewContentAccessDeniedError(userID, contentID int64, reason string) *AuthorizationError {
	return NewAuthorizationErrorWithContext(
		DomainContent,
		"access",
		"content",
		map[string]interface{}{
			"user_id":    userID,
			"content_id": contentID,
			"reason":     reason,
		},
	)
}

// NewContentTypeNotSupportedError creates a validation error for unsupported content types
func NewContentTypeNotSupportedError(contentType string, supportedTypes []string) *ValidationError {
	return NewValidationErrorWithRules(
		DomainContent,
		"content_type",
		fmt.Sprintf("Content type '%s' is not supported", contentType),
		contentType,
		map[string]interface{}{
			"supported_types": supportedTypes,
		},
	)
}

// NewContentSizeLimitError creates a validation error for content size limits
func NewContentSizeLimitError(actualSize, maxSize int64) *ValidationError {
	return NewValidationErrorWithRules(
		DomainContent,
		"content_size",
		"Content size exceeds maximum allowed limit",
		actualSize,
		map[string]interface{}{
			"max_size_bytes": maxSize,
			"max_size_mb":    maxSize / (1024 * 1024),
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

// new 1212121212


// new 121212121212
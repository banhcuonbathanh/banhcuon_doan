// internal/error_custom/constructors.go
// Domain-aware error constructors
package errorcustom

import (

	"fmt"
	
)


// NewNotFoundError creates a not found error with resource ID
func NewNotFoundError(domain, resourceType string, resourceID interface{}) *NotFoundError {
    return &NotFoundError{
        BaseError:    BaseError{Domain: domain, ErrorType: ErrorTypeNotFound},
        ResourceType: resourceType,
        ResourceID:   resourceID,
    }
}

// NewNotFoundErrorWithIdentifiers creates a not found error with multiple identifiers
func NewNotFoundErrorWithIdentifiers(domain, resourceType string, identifiers map[string]interface{}) *NotFoundError {
    return &NotFoundError{
        BaseError:    BaseError{Domain: domain, ErrorType: ErrorTypeNotFound},
        ResourceType: resourceType,
        Identifiers:  identifiers,
    }
}

// NewNotFoundErrorWithContext creates a not found error with context only
func NewNotFoundErrorWithContext(domain, resourceType string, context map[string]interface{}) *NotFoundError {
    return &NotFoundError{
        BaseError:    BaseError{Domain: domain, ErrorType: ErrorTypeNotFound},
        ResourceType: resourceType,
        Context:      context,
    }
}

// ----------------------------
// VALIDATION ERRORS
// ----------------------------

// NewValidationError creates a basic validation error
func NewValidationError(domain, field, message string, value interface{}) *ValidationError {
    return &ValidationError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeValidation},
        Field:     field,
        Message:   message,
        Value:     value,
    }
}

// NewValidationErrorWithRules creates a validation error with rule metadata
func NewValidationErrorWithRules(domain, field, message string, value interface{}, rules map[string]interface{}) *ValidationError {
    return &ValidationError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeValidation},
        Field:     field,
        Message:   message,
        Value:     value,
        Rules:     rules,
    }
}

// NewValidationErrorWithContext creates a validation error with context
func NewValidationErrorWithContext(domain, field, message string, value interface{}, context map[string]interface{}) *ValidationError {
    return &ValidationError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeValidation},
        Field:     field,
        Message:   message,
        Value:     value,
        Context:   context,
    }
}

// ----------------------------
// AUTHENTICATION ERRORS
// ----------------------------

// NewAuthenticationError creates a basic authentication error
func NewAuthenticationError(domain, reason string) *AuthenticationError {
    return &AuthenticationError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeAuthentication},
        Reason:    reason,
    }
}

// NewAuthenticationErrorWithStep creates an auth error with step and context
func NewAuthenticationErrorWithStep(domain, reason, step string, context map[string]interface{}) *AuthenticationError {
    return &AuthenticationError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeAuthentication},
        Reason:    reason,
        Step:      step,
        Context:   context,
    }
}

// NewAuthenticationErrorWithContext creates an auth error with context (no step)
func NewAuthenticationErrorWithContext(domain, reason string, context map[string]interface{}) *AuthenticationError {
    return &AuthenticationError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeAuthentication},
        Reason:    reason,
        Context:   context,
    }
}

// ----------------------------
// AUTHORIZATION ERRORS
// ----------------------------

// NewAuthorizationError creates a basic authorization error
func NewAuthorizationError(domain, action, resource string) *AuthorizationError {
    return &AuthorizationError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeAuthorization},
        Action:    action,
        Resource:  resource,
    }
}

// NewAuthorizationErrorWithContext creates an authorization error with context
func NewAuthorizationErrorWithContext(domain, action, resource string, context map[string]interface{}) *AuthorizationError {
    return &AuthorizationError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeAuthorization},
        Action:    action,
        Resource:  resource,
        Context:   context,
    }
}

// ----------------------------
// BUSINESS LOGIC ERRORS
// ----------------------------

// NewBusinessLogicError creates a basic business logic error
func NewBusinessLogicError(domain, rule, description string) *BusinessLogicError {
    return &BusinessLogicError{
        BaseError:   BaseError{Domain: domain, ErrorType: ErrorTypeBusinessLogic},
        Rule:        rule,
        Description: description,
    }
}

// NewBusinessLogicErrorWithContext creates a business logic error with context
func NewBusinessLogicErrorWithContext(domain, rule, description string, context map[string]interface{}) *BusinessLogicError {
    return &BusinessLogicError{
        BaseError:   BaseError{Domain: domain, ErrorType: ErrorTypeBusinessLogic},
        Rule:        rule,
        Description: description,
        Context:     context,
    }
}

// ----------------------------
// DUPLICATE / CONFLICT ERRORS
// ----------------------------

// NewDuplicateError creates a duplicate field error (e.g., email)
func NewDuplicateError(domain, resourceType, field string, value interface{}) *DuplicateError {
    return &DuplicateError{
        BaseError:    BaseError{Domain: domain, ErrorType: ErrorTypeDuplicate},
        ResourceType: resourceType,
        Field:        field,
        Value:        value,
    }
}

// NewConflictErrorWithContext creates a generic conflict error with message and context
func NewConflictErrorWithContext(domain, resourceType, message string, context map[string]interface{}) *ConflictError {
    return &ConflictError{
        BaseError:    BaseError{Domain: domain, ErrorType: ErrorTypeConflict},
        ResourceType: resourceType,
        Message:      message,
        Context:      context,
    }
}




func NewRateLimitErrorWithContext(domain, operation, message string, context map[string]interface{}) *RateLimitError {
    return &RateLimitError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeRateLimit},
        Operation: operation,
        Message:   message,
        Context:   context,
    }
}

// ----------------------------
// EXTERNAL SERVICE & SYSTEM ERRORS
// ----------------------------

// NewExternalServiceError creates an external service error
func NewExternalServiceError(domain, service, operation, message string, cause error, retryable bool) *ExternalServiceError {
    return &ExternalServiceError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeExternalService},
        Service:   service,
        Operation: operation,
        Message:   message,
        Cause:     cause,
        Retryable: retryable,
    }
}

// NewSystemError creates a system error
func NewSystemError(domain, component, operation, message string, cause error) *SystemError {
    return &SystemError{
        BaseError: BaseError{Domain: domain, ErrorType: ErrorTypeSystem},
        Component: component,
        Operation: operation,
        Message:   message,
        Cause:     cause,
    }
}



func NewUserNotFoundByID(id int64) *NotFoundError {
    return NewNotFoundError(DomainAccount, "user", id)
}

func NewUserNotFoundByEmail(email string) *NotFoundError {
    return NewNotFoundErrorWithIdentifiers(DomainAccount, "user", map[string]interface{}{"email": email})
}

func NewEmailNotFoundError(email string) *AuthenticationError {
    return NewAuthenticationErrorWithStep(DomainAccount, "email not found", "email_check", map[string]interface{}{"email": email, "user_found": false})
}

func NewPasswordMismatchError(email string) *AuthenticationError {
    return NewAuthenticationErrorWithStep(DomainAccount, "password mismatch", "password_check", map[string]interface{}{"email": email, "user_found": true})
}

func NewAccountDisabledError(email string) *AuthenticationError {
    return NewAuthenticationErrorWithStep(DomainAccount, "account disabled", "status_check", map[string]interface{}{"email": email, "user_found": true})
}

func NewAccountLockedError(email string, lockReason string) *AuthenticationError {
    return NewAuthenticationErrorWithStep(DomainAccount, fmt.Sprintf("account locked: %s", lockReason), "status_check", map[string]interface{}{"email": email, "user_found": true, "lock_reason": lockReason})
}

func NewDuplicateEmailError(email string) *DuplicateError {
    return NewDuplicateError(DomainAccount, "user", "email", email)
}

func NewWeakPasswordError(requirements []string) *ValidationError {
    return NewValidationErrorWithRules(DomainAccount, "password", "Password does not meet security requirements", "[REDACTED]", map[string]interface{}{"requirements": requirements})
}

// ============================================================================
// ADMIN DOMAIN ERROR CONSTRUCTORS
// ============================================================================

func NewInsufficientPrivilegesError(userID int64, requiredRole, currentRole string) *AuthorizationError {
    return NewAuthorizationErrorWithContext(DomainAdmin, "admin_operation", "system", map[string]interface{}{
        "user_id":       userID,
        "required_role": requiredRole,
        "current_role":  currentRole,
    })
}

func NewBulkOperationLimitError(operation string, requested, maxAllowed int) *BusinessLogicError {
    return NewBusinessLogicErrorWithContext(DomainAdmin, "bulk_operation_limit", fmt.Sprintf("Bulk %s operation exceeds maximum limit", operation), map[string]interface{}{
        "operation":   operation,
        "requested":   requested,
        "max_allowed": maxAllowed,
    })
}



func NewDatabaseError(operation, table string, cause error) *SystemError {
    return NewSystemError(DomainSystem, "database", operation, fmt.Sprintf("Database operation failed on table '%s'", table), cause)
}

func NewCacheError(operation, key string, cause error) *SystemError {
    return NewSystemError(DomainSystem, "cache", operation, fmt.Sprintf("Cache operation failed for key '%s'", key), cause)
}

func NewFileSystemError(operation, path string, cause error) *SystemError {
    return NewSystemError(DomainSystem, "filesystem", operation, fmt.Sprintf("File system operation failed for path '%s'", path), cause)
}

// ============================================================================
// LEGACY CONSTRUCTORS (backward compatibility)
// ============================================================================

type UserNotFoundError = NotFoundError
type DuplicateEmailError = DuplicateError
type PasswordValidationError = ValidationError
type ServiceError = ExternalServiceError
type RepositoryError = SystemError

func NewServiceError(service, method, message string, cause error, retryable bool) *ExternalServiceError {
    return NewExternalServiceError("", service, method, message, cause, retryable)
}

func NewRepositoryError(operation, table, message string, cause error) *SystemError {
    return NewSystemError(DomainSystem, "repository", operation, fmt.Sprintf("%s on table %s", message, table), cause)
}




// NewSecurityError creates a security-related error (unauthorized origin, suspicious activity, etc.)
func NewSecurityError(domain, securityCode, message string) *BusinessLogicError {
    return &BusinessLogicError{
        BaseError:   BaseError{Domain: domain, ErrorType: ErrorTypeBusinessLogic},
        Rule:        securityCode,
        Description: message,
    }
}

// NewSecurityErrorWithContext creates a security error with additional context
func NewSecurityErrorWithContext(domain, securityCode, message string, context map[string]interface{}) *BusinessLogicError {
    return &BusinessLogicError{
        BaseError:   BaseError{Domain: domain, ErrorType: ErrorTypeBusinessLogic},
        Rule:        securityCode,
        Description: message,
        Context:     context,
    }
}

// NewOriginNotAllowedError creates a specific error for unauthorized origins
func NewOriginNotAllowedError(domain, origin string, allowedOrigins []string) *BusinessLogicError {
    return NewSecurityErrorWithContext(
        domain,
        "unauthorized_origin",
        "Request origin not allowed",
        map[string]interface{}{
            "origin":          origin,
            "allowed_origins": allowedOrigins,
        },
    )
}

// NewSuspiciousActivityError creates an error for suspicious security activities
func NewSuspiciousActivityError(domain, activity, reason string, context map[string]interface{}) *BusinessLogicError {
    return NewSecurityErrorWithContext(
        domain,
        "suspicious_activity",
        fmt.Sprintf("Suspicious %s detected: %s", activity, reason),
        context,
    )
}

// NewRateLimitExceededError creates a rate limit error
func NewRateLimitExceededError(domain, operation string, limit int, timeWindow string) *RateLimitError {
    return NewRateLimitErrorWithContext(
        domain,
        operation,
        fmt.Sprintf("Rate limit exceeded: %d requests per %s", limit, timeWindow),
        map[string]interface{}{
            "limit":       limit,
            "time_window": timeWindow,
        },
    )
}
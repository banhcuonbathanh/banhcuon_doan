
// ============================================================================
// FILE: golang/internal/error_custom/domain/auth_errors.go
// ============================================================================
package errorcustom

import (

	"fmt"
	"time"
)

// Auth domain error constructors
type AuthDomainErrors struct{}

func NewAuthDomainErrors() *AuthDomainErrors {
	return &AuthDomainErrors{}
}

// Token Errors
func (a *AuthDomainErrors) NewInvalidTokenError(tokenType string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAuth,
		fmt.Sprintf("invalid %s token", tokenType),
		"token_validation",
		map[string]interface{}{
			"token_type": tokenType,
		},
	)
}

func (a *AuthDomainErrors) NewExpiredTokenError(tokenType string, expiredAt time.Time) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAuth,
		fmt.Sprintf("%s token has expired", tokenType),
		"token_expiry_check",
		map[string]interface{}{
			"token_type": tokenType,
			"expired_at": expiredAt,
		},
	)
}

func (a *AuthDomainErrors) NewMissingTokenError(tokenType string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		 DomainAuth,
		fmt.Sprintf("missing %s token", tokenType),
		"token_presence_check",
		map[string]interface{}{
			"token_type": tokenType,
		},
	)
}

// Session Errors
func (a *AuthDomainErrors) NewSessionExpiredError(sessionID string) * AuthenticationError {
	return  NewAuthenticationErrorWithStep(
		 DomainAuth,
		"session has expired",
		"session_validation",
		map[string]interface{}{
			"session_id": sessionID,
		},
	)
}

func (a *AuthDomainErrors) NewInvalidSessionError(sessionID string) * AuthenticationError {
	return  NewAuthenticationErrorWithStep(
		 DomainAuth,
		"invalid session",
		"session_validation",
		map[string]interface{}{
			"session_id": sessionID,
		},
	)
}

// Permission Errors
func (a *AuthDomainErrors) NewInsufficientPermissionsError(userID int64, requiredPermission string, userPermissions []string) * AuthorizationError {
	return  NewAuthorizationErrorWithContext(
		 DomainAuth,
		"access",
		"resource",
		map[string]interface{}{
			"user_id":             userID,
			"required_permission": requiredPermission,
			"user_permissions":    userPermissions,
		},
	)
}

func (a *AuthDomainErrors) NewRoleNotAuthorizedError(userID int64, userRole, requiredRole string) * AuthorizationError {
	return  NewAuthorizationErrorWithContext(
		 DomainAuth,
		"role_access",
		"resource",
		map[string]interface{}{
			"user_id":       userID,
			"user_role":     userRole,
			"required_role": requiredRole,
		},
	)
}

// Rate Limiting Errors
func (a *AuthDomainErrors) NewTooManyLoginAttemptsError(email string, remainingTime time.Duration) * BusinessLogicError {
	return  NewBusinessLogicErrorWithContext(
		 DomainAuth,
		"login_rate_limit",
		"Too many login attempts. Please try again later",
		map[string]interface{}{
			"email":          email,
			"remaining_time": remainingTime.String(),
		},
	)
}

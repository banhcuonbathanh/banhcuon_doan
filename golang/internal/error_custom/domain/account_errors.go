// golang/internal/error_custom/domain/account_errors.go
package error_custom_domain
import (

	"fmt"

	errorcustom "english-ai-full/internal/error_custom"

)

// AccountDomainErrors provides account domain error constructors
type AccountDomainErrors struct{}

// NewAccountDomainErrors creates a new account domain error helper
func NewAccountDomainErrors() *AccountDomainErrors {
	return &AccountDomainErrors{}
}

// User Authentication Errors
func (u *AccountDomainErrors) NewEmailNotFoundError(email string) *errorcustom.AuthenticationError {
	return errorcustom.NewAuthenticationErrorWithStep(
		errorcustom.DomainAccount,
		"email not found",
		"email_verification",
		map[string]interface{}{
			"email":      email,
			"user_found": false,
		},
	)
}

func (u *AccountDomainErrors) NewPasswordMismatchError(email string) *errorcustom.AuthenticationError {
	return errorcustom.NewAuthenticationErrorWithStep(
		errorcustom.DomainAccount,
		"password mismatch",
		"password_verification",
		map[string]interface{}{
			"email":      email,
			"user_found": true,
		},
	)
}

func (u *AccountDomainErrors) NewAccountDisabledError(email string, reason string) *errorcustom.AuthenticationError {
	return errorcustom.NewAuthenticationErrorWithStep(
		errorcustom.DomainAccount,
		fmt.Sprintf("account disabled: %s", reason),
		"account_status_check",
		map[string]interface{}{
			"email":           email,
			"user_found":      true,
			"disabled_reason": reason,
		},
	)
}

func (u *AccountDomainErrors) NewAccountLockedError(email string, lockReason string, unlockTime interface{}) *errorcustom.AuthenticationError {
	return errorcustom.NewAuthenticationErrorWithStep(
		errorcustom.DomainAccount,
		fmt.Sprintf("account locked: %s", lockReason),
		"account_lock_check",
		map[string]interface{}{
			"email":       email,
			"user_found":  true,
			"lock_reason": lockReason,
			"unlock_time": unlockTime,
		},
	)
}

// User Resource Errors
func (u *AccountDomainErrors) NewUserNotFoundByID(userID int64) *errorcustom.NotFoundError {
	return errorcustom.NewNotFoundError(errorcustom.DomainAccount, "user", userID)
}

func (u *AccountDomainErrors) NewUserNotFoundByEmail(email string) *errorcustom.NotFoundError {
	return errorcustom.NewNotFoundErrorWithIdentifiers(errorcustom.DomainAccount, "user", map[string]interface{}{
		"email": email,
	})
}

func (u *AccountDomainErrors) NewUserNotFoundByUsername(username string) *errorcustom.NotFoundError {
	return errorcustom.NewNotFoundErrorWithIdentifiers(errorcustom.DomainAccount, "user", map[string]interface{}{
		"username": username,
	})
}

// User Validation Errors
func (u *AccountDomainErrors) NewDuplicateEmailError(email string) *errorcustom.DuplicateError {
	return errorcustom.NewDuplicateError(errorcustom.DomainAccount, "user", "email", email)
}

func (u *AccountDomainErrors) NewDuplicateUsernameError(username string) *errorcustom.DuplicateError {
	return errorcustom.NewDuplicateError(errorcustom.DomainAccount, "user", "username", username)
}

func (u *AccountDomainErrors) NewWeakPasswordError(requirements []string) *errorcustom.ValidationError {
	return errorcustom.NewValidationErrorWithRules(
		errorcustom.DomainAccount,
		"password",
		"Password does not meet security requirements",
		"[REDACTED]",
		map[string]interface{}{
			"requirements": requirements,
		},
	)
}

func (u *AccountDomainErrors) NewInvalidEmailFormatError(email string) *errorcustom.ValidationError {
	return errorcustom.NewValidationError(
		errorcustom.DomainAccount,
		"email",
		"Invalid email format",
		email,
	)
}

// Account Resource Errors
func (u *AccountDomainErrors) NewAccountNotFoundError(accountID int64) *errorcustom.NotFoundError {
	return errorcustom.NewNotFoundError(errorcustom.DomainAccount, "account", accountID)
}

func (u *AccountDomainErrors) NewAccountClosedError(accountID int64) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAccount,
		"account_closed",
		"Account has been closed and cannot be accessed",
		map[string]interface{}{
			"account_id": accountID,
			"status":     "closed",
		},
	)
}

// User Business Logic Errors
func (u *AccountDomainErrors) NewEmailVerificationRequiredError(userID int64, email string) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAccount,
		"email_verification_required",
		"Email verification is required to proceed",
		map[string]interface{}{
			"user_id": userID,
			"email":   email,
		},
	)
}

func (u *AccountDomainErrors) NewPasswordResetRequiredError(userID int64) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAccount,
		"password_reset_required",
		"Password reset is required for security reasons",
		map[string]interface{}{
			"user_id": userID,
		},
	)
}

func (u *AccountDomainErrors) NewUserProfileIncompleteError(userID int64, missingFields []string) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAccount,
		"profile_incomplete",
		"User profile must be completed before proceeding",
		map[string]interface{}{
			"user_id":        userID,
			"missing_fields": missingFields,
		},
	)
}

// Additional Account Status Errors
func (u *AccountDomainErrors) NewAccountSuspendedError(accountID int64, reason string) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAccount,
		"account_suspended",
		"Account has been suspended",
		map[string]interface{}{
			"account_id": accountID,
			"reason":     reason,
			"status":     "suspended",
		},
	)
}

func (u *AccountDomainErrors) NewAccountPendingVerificationError(accountID int64) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAccount,
		"account_pending_verification",
		"Account is pending verification and cannot be used yet",
		map[string]interface{}{
			"account_id": accountID,
			"status":     "pending_verification",
		},
	)
}

func (u *AccountDomainErrors) NewInsufficientPermissionsError(userID int64, requiredPermission string) *errorcustom.AuthorizationError {
	return errorcustom.NewAuthorizationErrorWithContext(
		errorcustom.DomainAccount,
		"access",
		requiredPermission,
		map[string]interface{}{
			"user_id":             userID,
			"required_permission": requiredPermission,
		},
	)
}
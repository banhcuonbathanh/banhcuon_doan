// golang/internal/error_custom/account_domain_errors.go
package errorcustom

// AccountDomainErrors provides account domain error constructors
type AccountDomainErrors struct{}

// NewAccountDomainErrors creates a new account domain error helper
func NewAccountDomainErrors() *AccountDomainErrors {
	return &AccountDomainErrors{}
}

// User Authentication Errors
func (u *AccountDomainErrors) NewEmailNotFoundError(email string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAccount, // Remove 'errorcustom.' prefix - same package
		"email not found",
		"email_verification",
		map[string]interface{}{
			"email":      email,
			"user_found": false,
		},
	)
}

func (u *AccountDomainErrors) NewPasswordMismatchError(email string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAccount,
		"password mismatch",
		"password_verification",
		map[string]interface{}{
			"email":      email,
			"user_found": true,
		},
	)
}

func (u *AccountDomainErrors) NewAccountDisabledError(email string, reason string) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAccount,
		fmt.Sprintf("account disabled: %s", reason),
		"account_status_check",
		map[string]interface{}{
			"email":           email,
			"user_found":      true,
			"disabled_reason": reason,
		},
	)
}

func (u *AccountDomainErrors) NewAccountLockedError(email string, lockReason string, unlockTime interface{}) *AuthenticationError {
	return NewAuthenticationErrorWithStep(
		DomainAccount,
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
func (u *AccountDomainErrors) NewUserNotFoundByID(userID int64) *NotFoundError {
	return NewNotFoundError(DomainAccount, "user", userID)
}

func (u *AccountDomainErrors) NewUserNotFoundByEmail(email string) *NotFoundError {
	return NewNotFoundErrorWithIdentifiers(DomainAccount, "user", map[string]interface{}{
		"email": email,
	})
}

func (u *AccountDomainErrors) NewUserNotFoundByUsername(username string) *NotFoundError {
	return NewNotFoundErrorWithIdentifiers(DomainAccount, "user", map[string]interface{}{
		"username": username,
	})
}

// User Validation Errors
func (u *AccountDomainErrors) NewDuplicateEmailError(email string) *DuplicateError {
	return NewDuplicateError(DomainAccount, "user", "email", email)
}

func (u *AccountDomainErrors) NewDuplicateUsernameError(username string) *DuplicateError {
	return NewDuplicateError(DomainAccount, "user", "username", username)
}

func (u *AccountDomainErrors) NewWeakPasswordError(requirements []string) *ValidationError {
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

func (u *AccountDomainErrors) NewInvalidEmailFormatError(email string) *ValidationError {
	return NewValidationError(
		DomainAccount,
		"email",
		"Invalid email format",
		email,
	)
}

// User Business Logic Errors
func (u *AccountDomainErrors) NewEmailVerificationRequiredError(userID int64, email string) *BusinessLogicError {
	return NewBusinessLogicErrorWithContext(
		DomainAccount,
		"email_verification_required",
		"Email verification is required to proceed",
		map[string]interface{}{
			"user_id": userID,
			"email":   email,
		},
	)
}

func (u *AccountDomainErrors) NewPasswordResetRequiredError(userID int64) *BusinessLogicError {
	return NewBusinessLogicErrorWithContext(
		DomainAccount,
		"password_reset_required",
		"Password reset is required for security reasons",
		map[string]interface{}{
			"user_id": userID,
		},
	)
}

func (u *AccountDomainErrors) NewUserProfileIncompleteError(userID int64, missingFields []string) *BusinessLogicError {
	return NewBusinessLogicErrorWithContext(
		DomainAccount,
		"profile_incomplete",
		"User profile must be completed before proceeding",
		map[string]interface{}{
			"user_id":        userID,
			"missing_fields": missingFields,
		},
	)
}
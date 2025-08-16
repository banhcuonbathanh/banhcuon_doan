
// ============================================================================
// FILE: golang/internal/error_custom/unified_handler.go
// ============================================================================
package errorcustom

import (
	"context"
	"net/http"

	"english-ai-full/internal/error_custom/domain"
	"english-ai-full/internal/error_custom/layer"
)

// UnifiedErrorHandler provides a single interface for all error handling needs
type UnifiedErrorHandler struct {
	errorFactory *ErrorFactory
}

// NewUnifiedErrorHandler creates a new unified error handler
func NewUnifiedErrorHandler() *UnifiedErrorHandler {
	return &UnifiedErrorHandler{
		errorFactory: NewErrorFactory(),
	}
}

// ============================================================================
// HANDLER LAYER METHODS
// ============================================================================

// HandleHTTPError handles errors at the HTTP handler layer
func (ueh *UnifiedErrorHandler) HandleHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := GetRequestIDFromContext(r.Context())
	domain := GetDomainFromContext(r.Context())
	
	ueh.errorFactory.HandlerErrorMgr.RespondWithError(w, err, domain, requestID)
}

// ParseIDParam safely parses ID parameters from URL
func (ueh *UnifiedErrorHandler) ParseIDParam(r *http.Request, paramName string) (int64, error) {
	requestID := GetRequestIDFromContext(r.Context())
	domain := GetDomainFromContext(r.Context())
	
	return ueh.errorFactory.HandlerErrorMgr.ParseIDParameter(r, paramName, domain, requestID)
}

// ParsePaginationParams safely parses pagination parameters
func (ueh *UnifiedErrorHandler) ParsePaginationParams(r *http.Request) (limit, offset int64, err error) {
	requestID := GetRequestIDFromContext(r.Context())
	domain := GetDomainFromContext(r.Context())
	
	return ueh.errorFactory.HandlerErrorMgr.ParsePaginationParameters(r, domain, requestID)
}

// ============================================================================
// SERVICE LAYER METHODS
// ============================================================================

// WrapRepositoryError wraps repository errors with service context
func (ueh *UnifiedErrorHandler) WrapRepositoryError(err error, domain, operation string, context map[string]interface{}) error {
	return ueh.errorFactory.ServiceErrorMgr.WrapRepositoryError(err, domain, operation, context)
}

// HandleBusinessRuleViolation creates business logic errors
func (ueh *UnifiedErrorHandler) HandleBusinessRuleViolation(domain, rule, description string, context map[string]interface{}) error {
	return ueh.errorFactory.ServiceErrorMgr.HandleBusinessRuleViolation(domain, rule, description, context)
}

// HandleExternalServiceError handles external service failures
func (ueh *UnifiedErrorHandler) HandleExternalServiceError(err error, domain, service, operation string, retryable bool) error {
	return ueh.errorFactory.ServiceErrorMgr.HandleExternalServiceError(err, domain, service, operation, retryable)
}

// HandleContextError handles context-related errors
func (ueh *UnifiedErrorHandler) HandleContextError(ctx context.Context, domain, operation string) error {
	return ueh.errorFactory.ServiceErrorMgr.HandleContextError(ctx, domain, operation)
}

// ValidateBusinessRules validates multiple business rules
func (ueh *UnifiedErrorHandler) ValidateBusinessRules(domain string, validations map[string]func() error) error {
	return ueh.errorFactory.ServiceErrorMgr.ValidateBusinessRules(domain, validations)
}

// ============================================================================
// REPOSITORY LAYER METHODS
// ============================================================================

// HandleDatabaseError handles database-specific errors
func (ueh *UnifiedErrorHandler) HandleDatabaseError(err error, domain, table, operation string, context map[string]interface{}) error {
	return ueh.errorFactory.RepositoryErrorMgr.HandleDatabaseError(err, domain, table, operation, context)
}

// ============================================================================
// DOMAIN-SPECIFIC ERROR METHODS
// ============================================================================

// User Domain Errors
func (ueh *UnifiedErrorHandler) NewUserNotFoundByID(userID int64) error {
	return ueh.errorFactory.UserErrors.NewUserNotFoundByID(userID)
}

func (ueh *UnifiedErrorHandler) NewUserNotFoundByEmail(email string) error {
	return ueh.errorFactory.UserErrors.NewUserNotFoundByEmail(email)
}

func (ueh *UnifiedErrorHandler) NewDuplicateEmailError(email string) error {
	return ueh.errorFactory.UserErrors.NewDuplicateEmailError(email)
}

func (ueh *UnifiedErrorHandler) NewWeakPasswordError(requirements []string) error {
	return ueh.errorFactory.UserErrors.NewWeakPasswordError(requirements)
}

func (ueh *UnifiedErrorHandler) NewEmailNotFoundError(email string) error {
	return ueh.errorFactory.UserErrors.NewEmailNotFoundError(email)
}

func (ueh *UnifiedErrorHandler) NewPasswordMismatchError(email string) error {
	return ueh.errorFactory.UserErrors.NewPasswordMismatchError(email)
}

func (ueh *UnifiedErrorHandler) NewAccountDisabledError(email, reason string) error {
	return ueh.errorFactory.UserErrors.NewAccountDisabledError(email, reason)
}

// Auth Domain Errors
func (ueh *UnifiedErrorHandler) NewInvalidTokenError(tokenType string) error {
	return ueh.errorFactory.AuthErrors.NewInvalidTokenError(tokenType)
}

func (ueh *UnifiedErrorHandler) NewSessionExpiredError(sessionID string) error {
	return ueh.errorFactory.AuthErrors.NewSessionExpiredError(sessionID)
}

func (ueh *UnifiedErrorHandler) NewInsufficientPermissionsError(userID int64, requiredPermission string, userPermissions []string) error {
	return ueh.errorFactory.AuthErrors.NewInsufficientPermissionsError(userID, requiredPermission, userPermissions)
}

// Branch Domain Errors
func (ueh *UnifiedErrorHandler) NewBranchNotFoundError(branchID int64) error {
	return ueh.errorFactory.BranchErrors.NewBranchNotFoundError(branchID)
}

func (ueh *UnifiedErrorHandler) NewBranchNotFoundByCodeError(branchCode string) error {
	return ueh.errorFactory.BranchErrors.NewBranchNotFoundByCodeError(branchCode)
}

func (ueh *UnifiedErrorHandler) NewBranchInactiveError(branchID int64) error {
	return ueh.errorFactory.BranchErrors.NewBranchInactiveError(branchID)
}

// Admin Domain Errors
func (ueh *UnifiedErrorHandler) NewInsufficientAdminPrivilegesError(userID int64, requiredRole, currentRole string) error {
	return ueh.errorFactory.AdminErrors.NewInsufficientAdminPrivilegesError(userID, requiredRole, currentRole)
}

func (ueh *UnifiedErrorHandler) NewBulkOperationLimitError(operation string, requested, maxAllowed int) error {
	return ueh.errorFactory.AdminErrors.NewBulkOperationLimitError(operation, requested, maxAllowed)
}

// Account Domain Errors
func (ueh *UnifiedErrorHandler) NewAccountNotFoundError(accountID int64) error {
	return ueh.errorFactory.AccountErrors.NewAccountNotFoundError(accountID)
}

func (ueh *UnifiedErrorHandler) NewAccountClosedError(accountID int64) error {
	return ueh.errorFactory.AccountErrors.NewAccountClosedError(accountID)
}
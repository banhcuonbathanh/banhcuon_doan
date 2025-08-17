
// ============================================================================
// FILE: golang/internal/error_custom/unified_handler.go
// ============================================================================
package errorcustom

import (
	"context"
	"net/http"


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

// new 1212121212


// ============================================================================
// UNIFIED ERROR HANDLER - Already correct, but shown for completeness
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

// HandleDatabaseError handles database-specific errors
func (ueh *UnifiedErrorHandler) HandleDatabaseError(err error, domain, table, operation string, context map[string]interface{}) error {
	return ueh.errorFactory.RepositoryErrorMgr.HandleDatabaseError(err, domain, table, operation, context)
}

// new 12121212





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


// GetSortParamsWithDomain safely parses sorting parameters with domain context
func (ueh *UnifiedErrorHandler) GetSortParamsWithDomain(r *http.Request, allowedFields []string, domain string) (sortBy, sortOrder string, err error) {
	requestID := GetRequestIDFromContext(r.Context())
	
	return ueh.errorFactory.HandlerErrorMgr.ParseSortingParameters(r, allowedFields, domain, requestID)
}
// ParsePaginationParams safely parses pagination parameters
func (ueh *UnifiedErrorHandler) ParsePaginationParams(r *http.Request) (limit, offset int64, err error) {
	requestID := GetRequestIDFromContext(r.Context())
	domain := GetDomainFromContext(r.Context())
	
	return ueh.errorFactory.HandlerErrorMgr.ParsePaginationParameters(r, domain, requestID)
}

// DecodeJSONRequest decodes JSON request body with error handling
func (ueh *UnifiedErrorHandler) DecodeJSONRequest(r *http.Request, target interface{}) error {
	requestID := GetRequestIDFromContext(r.Context())
	domain := GetDomainFromContext(r.Context())
	
	return ueh.errorFactory.HandlerErrorMgr.DecodeJSONRequest(r, target, domain, requestID)
}

// RespondWithSuccess sends successful response
func (ueh *UnifiedErrorHandler) RespondWithSuccess(w http.ResponseWriter, data interface{}) {
	requestID := GetRequestIDFromContext(context.Background()) // You'll need to pass context here
	domain := GetDomainFromContext(context.Background())        // You'll need to pass context here
	
	ueh.errorFactory.HandlerErrorMgr.RespondWithSuccess(w, data, domain, requestID)
}







// HandleError processes errors through the unified error handling system
func (ueh *UnifiedErrorHandler) HandleError(domain string, err error) error {
	if err == nil {
		return nil
	}
	
	// Convert the error to APIError if it's not already
	apiErr := ConvertToAPIError(err)
	if apiErr == nil {
		// Create a generic system error if conversion fails
		apiErr = NewAPIError(
			ErrorTypeSystem,
			"An unexpected error occurred",
			http.StatusInternalServerError,
		).WithDomain(domain)
	}
	
	// Ensure the domain is set
	if apiErr.Domain == "" {
		apiErr = apiErr.WithDomain(domain)
	}
	
	return apiErr
}

// ParseGRPCError converts gRPC errors to domain-aware API errors
func (ueh *UnifiedErrorHandler) ParseGRPCError(err error, domain, operation string, context map[string]interface{}) error {
	return ParseGRPCError(err, domain, operation, context)
}



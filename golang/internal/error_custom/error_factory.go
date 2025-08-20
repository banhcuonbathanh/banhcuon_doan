
// ============================================================================
// FILE: golang/internal/error_custom/error_factory.go
// ============================================================================
package errorcustom

import (

	

	error_custom_domain "english-ai-full/internal/error_custom/domain"

)

// ErrorFactory provides centralized access to all error managers
type ErrorFactory struct {
	// Domain error managers
	AuthErrors    *AuthDomainErrors
	BranchErrors  *BranchDomainErrors
	AdminErrors   *AdminDomainErrors
	AccountErrors *error_custom_domain.AccountDomainErrors

	// Layer error managers
	HandlerErrorMgr    *HandlerErrorManager  // ← Added this missing field
	ServiceErrorMgr    *ServiceErrorManager
	RepositoryErrorMgr *RepositoryErrorManager
}

// NewErrorFactory creates a new error factory with all managers
func NewErrorFactory() *ErrorFactory {
	return &ErrorFactory{
		// Initialize domain error managers
		AuthErrors:    NewAuthDomainErrors(),
		BranchErrors:  NewBranchDomainErrors(),
		AdminErrors:   NewAdminDomainErrors(),
		AccountErrors: NewAccountDomainErrors(),

		// Initialize layer error managers
		HandlerErrorMgr:    NewHandlerErrorManager(),    // ← Added this initialization
		ServiceErrorMgr:    NewServiceErrorManager(),
		RepositoryErrorMgr: NewRepositoryErrorManager(),
	}
}
// GetDomainErrors returns domain-specific error manager
func (ef *ErrorFactory) GetDomainErrors(domain string) interface{} {
	switch domain {
	
	case DomainAuth:
		return ef.AuthErrors
	case "branch":
		return ef.BranchErrors
	case DomainAdmin:
		return ef.AdminErrors
	case "account":
		return ef.AccountErrors
	default:
		return nil
	}
}

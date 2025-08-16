
// ============================================================================
// FILE: golang/internal/error_custom/error_factory.go
// ============================================================================
package errorcustom

import (
	"english-ai-full/internal/error_custom/domain"
	"english-ai-full/internal/error_custom/layer"
)

// ErrorFactory provides centralized access to all error managers
type ErrorFactory struct {
	// Domain error managers
	UserErrors    *domain.UserDomainErrors
	AuthErrors    *domain.AuthDomainErrors
	BranchErrors  *domain.BranchDomainErrors
	AdminErrors   *domain.AdminDomainErrors
	AccountErrors *domain.AccountDomainErrors

	// Layer error managers
	HandlerErrorMgr    *layer.HandlerErrorManager
	ServiceErrorMgr    *layer.ServiceErrorManager
	RepositoryErrorMgr *layer.RepositoryErrorManager
}

// NewErrorFactory creates a new error factory with all managers
func NewErrorFactory() *ErrorFactory {
	return &ErrorFactory{
		// Initialize domain error managers
		UserErrors:    domain.NewUserDomainErrors(),
		AuthErrors:    domain.NewAuthDomainErrors(),
		BranchErrors:  domain.NewBranchDomainErrors(),
		AdminErrors:   domain.NewAdminDomainErrors(),
		AccountErrors: domain.NewAccountDomainErrors(),

		// Initialize layer error managers
		HandlerErrorMgr:    layer.NewHandlerErrorManager(),
		ServiceErrorMgr:    layer.NewServiceErrorManager(),
		RepositoryErrorMgr: layer.NewRepositoryErrorManager(),
	}
}

// GetDomainErrors returns domain-specific error manager
func (ef *ErrorFactory) GetDomainErrors(domain string) interface{} {
	switch domain {
	case DomainUser:
		return ef.UserErrors
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


// ============================================================================
// FILE: golang/internal/error_custom/error_factory.go
// ============================================================================
package errorcustom


// ErrorFactory provides centralized access to all error managers
type ErrorFactory struct {
	// Domain error managers

	AuthErrors    * AuthDomainErrors
	BranchErrors  * BranchDomainErrors
	AdminErrors   * AdminDomainErrors
	AccountErrors * AccountDomainErrors

	// Layer error managers

	ServiceErrorMgr    * ServiceErrorManager
	RepositoryErrorMgr * RepositoryErrorManager
}

// NewErrorFactory creates a new error factory with all managers
func NewErrorFactory() *ErrorFactory {
	return &ErrorFactory{
		// Initialize domain error managers
	
		AuthErrors:     NewAuthDomainErrors(),
		BranchErrors:   NewBranchDomainErrors(),
		AdminErrors:    NewAdminDomainErrors(),
		AccountErrors:  NewAccountDomainErrors(),

		// Initialize layer error managers

		ServiceErrorMgr:     NewServiceErrorManager(),
		RepositoryErrorMgr:  NewRepositoryErrorManager(),
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

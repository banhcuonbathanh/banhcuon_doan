
// ============================================================================
// FILE: golang/internal/error_custom/error_factory.go
// ============================================================================
package errorcustom



// ErrorFactory provides centralized access to all error managers
type ErrorFactory struct {


	// Layer error managers
	HandlerErrorMgr    *HandlerErrorManager  // ← Added this missing field
	ServiceErrorMgr    *ServiceErrorManager
	RepositoryErrorMgr *RepositoryErrorManager
}

// NewErrorFactory creates a new error factory with all managers
func NewErrorFactory() *ErrorFactory {
	return &ErrorFactory{
		// Initialize domain error managers


		// Initialize layer error managers
		HandlerErrorMgr:    NewHandlerErrorManager(),    // ← Added this initialization
		ServiceErrorMgr:    NewServiceErrorManager(),
		RepositoryErrorMgr: NewRepositoryErrorManager(),
	}
}


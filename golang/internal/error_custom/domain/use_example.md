
// ============================================================================
// USAGE EXAMPLE:
// ============================================================================

/*
// In your handlers, services, or repositories:

import (
    "english-ai-full/internal/error_custom/domain"
)

func main() {
    // Initialize domain error helpers
    userErrors := domain.NewUserDomainErrors()
    authErrors := domain.NewAuthDomainErrors()
    branchErrors := domain.NewBranchDomainErrors()
    adminErrors := domain.NewAdminDomainErrors()
    accountErrors := domain.NewAccountDomainErrors()

    // Usage examples:
    
    // User domain errors
    err1 := userErrors.NewEmailNotFoundError("user@example.com")
    err2 := userErrors.NewWeakPasswordError([]string{"at least 8 characters", "one uppercase letter"})
    
    // Auth domain errors
    err3 := authErrors.NewExpiredTokenError("JWT", time.Now().Add(-time.Hour))
    err4 := authErrors.NewInsufficientPermissionsError(123, "admin", []string{"user", "viewer"})
    
    // Branch domain errors
    err5 := branchErrors.NewBranchNotFoundByCodeError("NYC001")
    err6 := branchErrors.NewBranchCapacityExceededError(1, 100, 100)
    
    // Admin domain errors
    err7 := adminErrors.NewBulkOperationLimitError("user_creation", 1000, 500)
    
    // Account domain errors
    err8 := accountErrors.NewAccountClosedError(12345)
}
*/
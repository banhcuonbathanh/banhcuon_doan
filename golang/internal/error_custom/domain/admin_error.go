// ============================================================================
// FILE: golang/internal/error_custom/domain/admin_errors.go
// ============================================================================
package domain

import (
	errorcustom "english-ai-full/internal/error_custom"
	"fmt"
)

// Admin domain error constructors
type AdminDomainErrors struct{}

func NewAdminDomainErrors() *AdminDomainErrors {
	return &AdminDomainErrors{}
}

// Admin Authorization Errors
func (a *AdminDomainErrors) NewInsufficientAdminPrivilegesError(userID int64, requiredRole, currentRole string) *errorcustom.AuthorizationError {
	return errorcustom.NewAuthorizationErrorWithContext(
		errorcustom.DomainAdmin,
		"admin_operation",
		"system",
		map[string]interface{}{
			"user_id":       userID,
			"required_role": requiredRole,
			"current_role":  currentRole,
		},
	)
}

func (a *AdminDomainErrors) NewSystemMaintenanceModeError() *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicError(
		errorcustom.DomainAdmin,
		"system_maintenance",
		"System is currently in maintenance mode",
	)
}

// Bulk Operations Errors
func (a *AdminDomainErrors) NewBulkOperationLimitError(operation string, requested, maxAllowed int) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAdmin,
		"bulk_operation_limit",
		fmt.Sprintf("Bulk %s operation exceeds maximum limit", operation),
		map[string]interface{}{
			"operation":   operation,
			"requested":   requested,
			"max_allowed": maxAllowed,
		},
	)
}

func (a *AdminDomainErrors) NewBulkOperationPartialFailureError(operation string, totalRequested, successful, failed int, failures []string) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAdmin,
		"bulk_operation_partial_failure",
		fmt.Sprintf("Bulk %s operation completed with %d failures out of %d requests", operation, failed, totalRequested),
		map[string]interface{}{
			"operation":        operation,
			"total_requested":  totalRequested,
			"successful_count": successful,
			"failed_count":     failed,
			"failure_details":  failures,
		},
	)
}

// Resource Management Errors
func (a *AdminDomainErrors) NewResourceQuotaExceededError(resourceType string, currentUsage, maxQuota int64) *errorcustom.BusinessLogicError {
	return errorcustom.NewBusinessLogicErrorWithContext(
		errorcustom.DomainAdmin,
		"resource_quota",
		fmt.Sprintf("%s quota exceeded", resourceType),
		map[string]interface{}{
			"resource_type":   resourceType,
			"current_usage":   currentUsage,
			"maximum_quota":   maxQuota,
			"usage_percent":   float64(currentUsage) / float64(maxQuota) * 100,
		},
	)
}
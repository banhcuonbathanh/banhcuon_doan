// ============================================================================
// FILE: golang/internal/error_custom/domain/branch_errors.go
// ============================================================================
package errorcustom

const DomainBranch = "branch"

// Branch domain error constructors
type BranchDomainErrors struct{}

func NewBranchDomainErrors() *BranchDomainErrors {
	return &BranchDomainErrors{}
}

// Branch Resource Errors
func (b *BranchDomainErrors) NewBranchNotFoundError(branchID int64) * NotFoundError {
	return  NewNotFoundError(DomainBranch, "branch", branchID)
}

func (b *BranchDomainErrors) NewBranchNotFoundByCodeError(branchCode string) * NotFoundError {
	return  NewNotFoundErrorWithIdentifiers(DomainBranch, "branch", map[string]interface{}{
		"branch_code": branchCode,
	})
}

func (b *BranchDomainErrors) NewBranchNotFoundByLocationError(city, state string) * NotFoundError {
	return  NewNotFoundErrorWithIdentifiers(DomainBranch, "branch", map[string]interface{}{
		"city":  city,
		"state": state,
	})
}

// Branch Validation Errors
func (b *BranchDomainErrors) NewDuplicateBranchCodeError(branchCode string) * DuplicateError {
	return  NewDuplicateError(DomainBranch, "branch", "branch_code", branchCode)
}

func (b *BranchDomainErrors) NewInvalidBranchCodeFormatError(branchCode string) * ValidationError {
	return  NewValidationErrorWithRules(
		DomainBranch,
		"branch_code",
		"Branch code must be 3-10 alphanumeric characters",
		branchCode,
		map[string]interface{}{
			"min_length": 3,
			"max_length": 10,
			"pattern":    "alphanumeric",
		},
	)
}

// Branch Business Logic Errors
func (b *BranchDomainErrors) NewBranchInactiveError(branchID int64) * BusinessLogicError {
	return  NewBusinessLogicErrorWithContext(
		DomainBranch,
		"branch_status",
		"Branch is currently inactive",
		map[string]interface{}{
			"branch_id": branchID,
			"status":    "inactive",
		},
	)
}

func (b *BranchDomainErrors) NewBranchCapacityExceededError(branchID int64, currentCapacity, maxCapacity int) * BusinessLogicError {
	return  NewBusinessLogicErrorWithContext(
		DomainBranch,
		"branch_capacity",
		"Branch has reached maximum capacity",
		map[string]interface{}{
			"branch_id":         branchID,
			"current_capacity":  currentCapacity,
			"maximum_capacity":  maxCapacity,
		},
	)
}


// internal/account/account_service/account_service_base.go
package account_service

import (
	"context"
	"fmt"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/model"
	"english-ai-full/internal/proto_qr/account"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// BaseServiceStruct provides common utilities and methods for account service operations
type BaseServiceStruct struct {
	*ServiceStruct
}

// Common utility methods

// validatePaginationParams validates and sets default pagination parameters
func (s *ServiceStruct) validatePaginationParams(page, pageSize int32) (int32, int32, error) {
	// Set defaults if not provided
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	
	// Set maximum limits
	const maxPageSize = 100
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	
	return page, pageSize, nil
}

// validateSortParams validates and sets default sort parameters
func (s *ServiceStruct) validateSortParams(sortBy, sortOrder string) (string, string) {
	// Valid sort fields
	validSortFields := map[string]bool{
		"id":         true,
		"name":       true,
		"email":      true,
		"role":       true,
		"created_at": true,
		"updated_at": true,
		"branch_id":  true,
	}
	
	// Default sort field
	if sortBy == "" || !validSortFields[sortBy] {
		sortBy = "created_at"
	}
	
	// Default sort order
	if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
		sortOrder = "desc"
	}
	
	return sortBy, sortOrder
}

// Data conversion utilities

// modelsToProtoAccounts converts slice of model.Account to slice of protobuf accounts
func (s *ServiceStruct) modelsToProtoAccounts(users []model.Account) []*account.Account {
	var accounts []*account.Account
	for _, user := range users {
		accounts = append(accounts, s.modelToProto(user))
	}
	return accounts
}

// searchResultsToProtoAccounts converts search results to protobuf accounts
func (s *ServiceStruct) searchResultsToProtoAccounts(users []model.Account) []*account.Account {
	var accounts []*account.Account
	for i := range users {
		user := &users[i] // Get pointer to avoid copying
		accounts = append(accounts, &account.Account{
			Id:        user.ID,        // Note: changed from user.Id to user.ID
			BranchId:  user.BranchID,  // Note: changed from user.BranchId to user.BranchID
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      string(user.Role),
			OwnerId:   user.OwnerID,   // Note: changed from user.OwnerId to user.OwnerID
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		})
	}
	return accounts
}

// Error handling utilities

// handleRepositoryError provides consistent error handling for repository errors
func (s *ServiceStruct) handleRepositoryError(err error, operation, table string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// Check for specific error types using string matching
	errorStr := strings.ToLower(err.Error())
	
	// User not found errors
	if strings.Contains(errorStr, "not found") {
		if userID, ok := context["user_id"].(int64); ok {
			return errorcustom.NewUserNotFoundByID(userID)
		}
		if email, ok := context["email"].(string); ok {
			return errorcustom.NewUserNotFoundByEmail(email)
		}
		return errorcustom.NewRepositoryError(operation, table, "Resource not found", err)
	}
	
	// Duplicate/constraint errors
	if strings.Contains(errorStr, "duplicate") || 
	   strings.Contains(errorStr, "already exists") ||
	   strings.Contains(errorStr, "constraint") {
		if email, ok := context["email"].(string); ok {
			return errorcustom.NewDuplicateEmailError(email)
		}
		return errorcustom.NewRepositoryError(operation, table, "Resource already exists", err)
	}
	
	// Connection/timeout errors (potentially retryable)
	if strings.Contains(errorStr, "connection") ||
	   strings.Contains(errorStr, "timeout") ||
	   strings.Contains(errorStr, "unavailable") {
		return errorcustom.NewServiceError(
			"AccountService",
			operation,
			"Service temporarily unavailable",
			err,
			true, // retryable
		)
	}
	
	// Generic repository error
	return errorcustom.NewRepositoryError(operation, table, err.Error(), err)
}

// handleServiceError provides consistent error handling for service-level errors
func (s *ServiceStruct) handleServiceError(operation, message string, err error, retryable bool) error {
	return errorcustom.NewServiceError("AccountService", operation, message, err, retryable)
}

// Validation utilities

// validateUserRole validates if the provided role is valid
func (s *ServiceStruct) validateUserRole(role string) error {
	validRoles := map[string]bool{
		"admin":   true,
		"user":    true,
		"manager": true,
		"guest":   true,
	}
	
	if !validRoles[strings.ToLower(role)] {
		return errorcustom.NewValidationError(
			"role",
			fmt.Sprintf("Invalid role: %s. Valid roles are: admin, user, manager, guest", role),
			role,
		)
	}
	
	return nil
}

// validateAccountStatus validates if the provided status is valid
func (s *ServiceStruct) validateAccountStatus(status string) error {
	validStatuses := map[string]bool{
		"active":    true,
		"inactive":  true,
		"suspended": true,
		"pending":   true,
	}
	
	if !validStatuses[strings.ToLower(status)] {
		return errorcustom.NewValidationError(
			"status",
			fmt.Sprintf("Invalid status: %s. Valid statuses are: active, inactive, suspended, pending", status),
			status,
		)
	}
	
	return nil
}

// validateEmail performs basic email validation
func (s *ServiceStruct) validateEmail(email string) error {
	if email == "" {
		return errorcustom.NewValidationError("email", "Email is required", "")
	}
	
	if !strings.Contains(email, "@") {
		return errorcustom.NewValidationError("email", "Invalid email format", email)
	}
	
	return nil
}

// Business logic helpers

// checkUserPermissions validates if the requesting user has permission to perform an operation
func (s *ServiceStruct) checkUserPermissions(ctx context.Context, requestingUserID, targetUserID int64, operation string) error {
	// If user is operating on themselves, allow most operations
	if requestingUserID == targetUserID {
		return nil
	}
	
	// For operations on other users, need to check role/permissions
	// This would typically involve checking the requesting user's role
	requestingUser, err := s.userRepo.FindByID(ctx, requestingUserID)
	if err != nil {
		return s.handleRepositoryError(err, "check_permissions", "users", map[string]interface{}{
			"user_id": requestingUserID,
		})
	}
	
	// Admin users can perform any operation
	if strings.ToLower(string(requestingUser.Role)) == "admin" {
		return nil
	}
	
	// Manager users can perform operations on users in their branch
	if strings.ToLower(string(requestingUser.Role)) == "manager" {
		targetUser, err := s.userRepo.FindByID(ctx, targetUserID)
		if err != nil {
			return s.handleRepositoryError(err, "check_permissions", "users", map[string]interface{}{
				"user_id": targetUserID,
			})
		}
		
		if requestingUser.BranchID == targetUser.BranchID {
			return nil
		}
	}
	
	// Regular users can only operate on themselves
	return &errorcustom.AuthorizationError{
		Action:   operation,
		Resource: fmt.Sprintf("user_%d", targetUserID),
	}
}

// Logging utilities

// logServiceCall logs service method calls with context
func (s *ServiceStruct) logServiceCall(method string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"service": "AccountService",
		"method":  method,
	}
	
	// Merge provided context
	for k, v := range context {
		logContext[k] = v
	}
	
	s.logger.Info("Service method called", logContext)
}

// logServiceError logs service errors with full context
func (s *ServiceStruct) logServiceError(method string, err error, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"service": "AccountService",
		"method":  method,
		"error":   err.Error(),
	}
	
	// Merge provided context
	for k, v := range context {
		logContext[k] = v
	}
	
	s.logger.Error("Service method error", logContext)
}

// Pagination utilities

// createPaginationInfo creates pagination metadata for responses
func (s *ServiceStruct) createPaginationInfo(page, pageSize int32, totalCount int64) *account.PaginationInfo {
	totalPages := int32((totalCount + int64(pageSize) - 1) / int64(pageSize)) // Ceiling division
	hasNext := page < totalPages
	hasPrev := page > 1
	
	return &account.PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// Email notification helpers

// sendEmailAsync sends email notifications asynchronously with error logging
func (s *ServiceStruct) sendEmailAsync(emailFunc func() error, operation string, userEmail string) {
	if s.emailService == nil {
		return
	}
	
	go func() {
		if err := emailFunc(); err != nil {
			s.logger.Error("Failed to send email", map[string]interface{}{
				"operation":  operation,
				"user_email": userEmail,
				"error":      err.Error(),
			})
		} else {
			s.logger.Info("Email sent successfully", map[string]interface{}{
				"operation":  operation,
				"user_email": userEmail,
			})
		}
	}()
}

// Service dependency checks

// requireTokenMaker checks if token maker is available and returns error if not
func (s *ServiceStruct) requireTokenMaker(operation string) error {
	if s.tokenMaker == nil {
		return s.handleServiceError(
			operation,
			"Token functionality not available",
			fmt.Errorf("missing tokenMaker dependency"),
			false,
		)
	}
	return nil
}

// requireEmailService checks if email service is available and returns error if not
func (s *ServiceStruct) requireEmailService(operation string) error {
	if s.emailService == nil {
		return s.handleServiceError(
			operation,
			"Email functionality not available",
			fmt.Errorf("missing emailService dependency"),
			false,
		)
	}
	return nil
}

// requirePasswordHasher checks if password hasher is available and returns error if not
func (s *ServiceStruct) requirePasswordHasher(operation string) error {
	if s.passwordHash == nil {
		return s.handleServiceError(
			operation,
			"Password hashing functionality not available",
			fmt.Errorf("missing passwordHash dependency"),
			false,
		)
	}
	return nil
}
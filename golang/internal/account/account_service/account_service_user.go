// internal/account/account_service_user.go
package account_service

import (
	"context"
	"net/http"

	"strings"
	"time"

	"english-ai-full/internal/model"
	"english-ai-full/internal/proto_qr/account"

	errorcustom "english-ai-full/internal/error_custom"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateUser handles user creation
func (s *ServiceStruct) CreateUser(ctx context.Context, req *account.AccountReq) (*account.Account, error) {
	// Hash password if provided
	hashedPassword := req.Password
	if s.passwordHash != nil && req.Password != "" {
		hashed, err := s.passwordHash.HashPassword(req.Password)
		if err != nil {
			return nil, pkgerrors.WithStack(err)
		}
		hashedPassword = hashed
	}

	user, err := s.userRepo.CreateUser(ctx, model.Account{
		BranchID:  req.BranchId,
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		Avatar:    req.Avatar,
		Title:     req.Title,
		Role:      model.Role(req.Role),
		OwnerID:   req.OwnerId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Send welcome email if email service is available
	if s.emailService != nil {
		go func() {
			if err := s.emailService.SendWelcomeEmail(context.Background(), user.Email, user.Name); err != nil {
				s.logger.Error("Failed to send welcome email")
			}
		}()
	}

	return &account.Account{
		Id:        user.ID,
		BranchId:  user.BranchID,
		Name:      user.Name,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Title:     user.Title,
		Role:      string(user.Role),
		OwnerId:   user.OwnerID,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

// UpdateUser handles user updates
func (s *ServiceStruct) UpdateUser(ctx context.Context, req *account.UpdateUserReq) (*account.AccountRes, error) {
	// Update user in repository
	user, err := s.userRepo.UpdateUser(ctx, model.Account{
		ID:       req.Id,
		BranchID: req.BranchId,
		Name:     req.Name,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     model.Role(req.Role),
		OwnerID:  req.OwnerId,
	})
	if err != nil {
		// Check if it's a user not found error using the utility function
		if errorcustom.IsUserNotFoundError(err) {
			return nil, errorcustom.NewUserNotFoundByID(req.Id)
		}
		
		// Check if it's a duplicate email error
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || 
		   strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return nil, errorcustom.NewDuplicateEmailError(req.Email)
		}
		
		// For database/repository errors, wrap them appropriately
		if strings.Contains(strings.ToLower(err.Error()), "database") ||
		   strings.Contains(strings.ToLower(err.Error()), "sql") {
			return nil, errorcustom.NewRepositoryError("update", "users", err.Error(), err)
		}
		
		// Generic service error for other cases
		return nil, errorcustom.NewServiceError("UserService", "UpdateUser", err.Error(), err, false)
	}

	// Return successful response
	return &account.AccountRes{
		Account: &account.Account{
			Id:        user.ID,
			BranchId:  user.BranchID,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      string(user.Role),
			OwnerId:   user.OwnerID,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}
// DeleteUser handles user deletion// DeleteUser handles user deletion
func (s *ServiceStruct) DeleteUser(ctx context.Context, req *account.DeleteAccountReq) (*account.DeleteAccountRes, error) {
	// Get user info before deletion for email notification
	var user model.Account
	if s.emailService != nil {
		var err error
		user, err = s.userRepo.FindByID(ctx, req.UserID)
		if err != nil && !errorcustom.IsUserNotFoundError(err) {
			// Log repository error but don't fail the deletion process
			s.logger.Error("Failed to get user info before deletion", map[string]interface{}{
				"user_id": req.UserID,
				"error":   err.Error(),
			})
		}
	}

	// Attempt to delete the user
	err := s.userRepo.DeleteUser(ctx, req.UserID)
	if err != nil {
		// Check if it's a user not found error using the utility function
		if errorcustom.IsUserNotFoundError(err) {
			return nil, errorcustom.NewUserNotFoundByID(req.UserID)
		}
		
		// For database/repository errors, wrap them appropriately
		if strings.Contains(strings.ToLower(err.Error()), "database") ||
		   strings.Contains(strings.ToLower(err.Error()), "sql") ||
		   strings.Contains(strings.ToLower(err.Error()), "constraint") {
			return nil, errorcustom.NewRepositoryError("delete", "users", err.Error(), err)
		}
		
		// Generic service error for other cases
		return nil, errorcustom.NewServiceError("UserService", "DeleteUser", err.Error(), err, false)
	}

	// Send account deactivation email if email service is available
	if s.emailService != nil && user.Email != "" {
		go func() {
			if err := s.emailService.SendAccountDeactivationEmail(context.Background(), user.Email, user.Name); err != nil {
				s.logger.Error("Failed to send account deactivation email", map[string]interface{}{
					"user_id":    req.UserID,
					"user_email": user.Email,
					"error":      err.Error(),
				})
			}
		}()
	}

	return &account.DeleteAccountRes{
		Success: true,
	}, nil
}

// UpdateAccountStatus handles account status updates
func (s *ServiceStruct) UpdateAccountStatus(ctx context.Context, req *account.UpdateAccountStatusReq) (*account.UpdateAccountStatusRes, error) {
	// Update account status in repository
	err := s.userRepo.UpdateAccountStatus(ctx, req.UserId, req.Status)
	if err != nil {
		// Check if it's a user not found error using the utility function
		if errorcustom.IsUserNotFoundError(err) {
			return nil, errorcustom.NewUserNotFoundByID(req.UserId)
		}
		
		// Check for invalid status errors
		if strings.Contains(strings.ToLower(err.Error()), "invalid status") ||
		   strings.Contains(strings.ToLower(err.Error()), "invalid value") {
			return nil, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid account status provided",
				http.StatusBadRequest,
			).WithDetail("status", req.Status)
		}
		
		// For database/repository errors, wrap them appropriately
		if strings.Contains(strings.ToLower(err.Error()), "database") ||
		   strings.Contains(strings.ToLower(err.Error()), "sql") {
			return nil, errorcustom.NewRepositoryError("update_status", "users", err.Error(), err)
		}
		
		// Generic service error for other cases
		return nil, errorcustom.NewServiceError("UserService", "UpdateAccountStatus", err.Error(), err, false)
	}

	return &account.UpdateAccountStatusRes{
		Success: true,
		Message: "Account status updated successfully",
	}, nil
}
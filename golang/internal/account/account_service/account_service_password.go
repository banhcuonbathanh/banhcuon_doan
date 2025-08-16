// internal/account/account_service_password.go
package account_service

import (
	"context"
	"fmt"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"
)

// ChangePassword handles password change requests
func (s *AccountService) ChangePassword(ctx context.Context, req *account.ChangePasswordReq) (*account.ChangePasswordRes, error) {
	// Verify current password
	user, err := s.userRepo.FindByID(ctx, req.UserId)
	if err != nil {
		// Check if it's a user not found error using string matching
		if strings.Contains(err.Error(), "not found") {
			return nil, errorcustom.NewUserNotFoundByID(req.UserId)
		}
		// For other repository errors
		repoErr := errorcustom.NewRepositoryError(
			"find_user",
			"users",
			"Failed to find user for password change",
			err,
		)
		return nil, repoErr
	}

	var isValidPassword bool
	if s.passwordHash != nil {
		isValidPassword = s.passwordHash.ComparePassword(user.Password, req.CurrentPassword)
	} else {
		isValidPassword = utils.Compare(user.Password, req.CurrentPassword)
	}

	if !isValidPassword {
		// Return authentication error for wrong current password
		return nil, errorcustom.NewPasswordMismatchError(user.Email)
	}

	// Hash new password
	hashedPassword := req.NewPassword
	if s.passwordHash != nil {
		hashed, err := s.passwordHash.HashPassword(req.NewPassword)
		if err != nil {
			serviceErr := errorcustom.NewServiceError(
				"AccountService",
				"ChangePassword",
				"Failed to hash new password",
				err,
				false,
			)
			return nil, serviceErr
		}
		hashedPassword = hashed
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, req.UserId, hashedPassword)
	if err != nil {
		repoErr := errorcustom.NewRepositoryError(
			"update_password",
			"users",
			"Failed to update user password",
			err,
		)
		return nil, repoErr
	}

	// Send password changed notification
	if s.emailService != nil {
		go func() {
			if err := s.emailService.SendPasswordChangedNotification(context.Background(), user.Email, user.Name); err != nil {
				s.logger.Error("Failed to send password changed notification")
			}
		}()
	}

	return &account.ChangePasswordRes{
		Success: true,
		Message: "Password changed successfully",
	}, nil
}

// ForgotPassword handles forgot password requests
func (s *AccountService) ForgotPassword(ctx context.Context, req *account.ForgotPasswordReq) (*account.ForgotPasswordRes, error) {
	// Check if user exists
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Check if it's a user not found error using string matching
		if strings.Contains(err.Error(), "not found") {
			// Don't reveal that email doesn't exist for security
			// Return success response regardless
			return &account.ForgotPasswordRes{
				Success: true,
				Message: "If the email exists, a password reset link has been sent",
			}, nil
		}
		// For other repository errors
		repoErr := errorcustom.NewRepositoryError(
			"find_user_by_email",
			"users",
			"Failed to find user for password reset",
			err,
		)
		return nil, repoErr
	}

	if s.tokenMaker == nil || s.emailService == nil {
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"ForgotPassword",
			"Password reset functionality not available",
			fmt.Errorf("missing tokenMaker or emailService dependency"),
			false,
		)
		return nil, serviceErr
	}

	// Create reset token
	resetToken, err := s.tokenMaker.CreateResetToken(user.Email)
	if err != nil {
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"ForgotPassword",
			"Failed to create reset token",
			err,
			false,
		)
		return nil, serviceErr
	}

	// Store reset token
	err = s.userRepo.StoreResetToken(ctx, user.Email, resetToken)
	if err != nil {
		repoErr := errorcustom.NewRepositoryError(
			"store_reset_token",
			"password_reset_tokens",
			"Failed to store reset token",
			err,
		)
		return nil, repoErr
	}

	// Send reset email
	go func() {
		if err := s.emailService.SendPasswordResetEmail(context.Background(), user.Email, resetToken); err != nil {
			s.logger.Error("Failed to send password reset email")
		}
	}()

	return &account.ForgotPasswordRes{
		Success: true,
		Message: "Password reset link has been sent to your email",
	}, nil
}

// ResetPassword handles password reset requests
func (s *AccountService) ResetPassword(ctx context.Context, req *account.ResetPasswordReq) (*account.ResetPasswordRes, error) {
	if s.tokenMaker == nil {
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"ResetPassword",
			"Password reset functionality not available",
			fmt.Errorf("missing tokenMaker dependency"),
			false,
		)
		return nil, serviceErr
	}

	// Validate reset token
	email, err := s.userRepo.ValidateResetToken(ctx, req.Token)
	if err != nil {
		// Check if it's token validation specific error
		if strings.Contains(err.Error(), "expired") {
			return nil, errorcustom.NewAuthenticationErrorWithStep(
    errorcustom.DomainAuth, 
    "token has expired", 
    "token_validation",
    map[string]interface{}{
        "token_type": "reset_token",
        "reason":     "expired",
    },
)
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "not found") {
		return nil, errorcustom.NewAuthenticationErrorWithStep(
    errorcustom.DomainAuth, 
    "token has expired", 
    "token_validation",
    map[string]interface{}{
        "token_type": "reset_token",
        "reason":     "expired",
    },
)
		}
		// For other repository errors
		repoErr := errorcustom.NewRepositoryError(
			"validate_reset_token",
			"password_reset_tokens",
			"Failed to validate reset token",
			err,
		)
		return nil, repoErr
	}

	// Hash new password
	hashedPassword := req.NewPassword
	if s.passwordHash != nil {
		hashed, err := s.passwordHash.HashPassword(req.NewPassword)
		if err != nil {
			serviceErr := errorcustom.NewServiceError(
				"AccountService",
				"ResetPassword",
				"Failed to hash new password",
				err,
				false,
			)
			return nil, serviceErr
		}
		hashedPassword = hashed
	}

	// Get user to update password
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Check if it's a user not found error using string matching
		if strings.Contains(err.Error(), "not found") {
			return nil, errorcustom.NewUserNotFoundByEmail(email)
		}
		// For other repository errors
		repoErr := errorcustom.NewRepositoryError(
			"find_user_by_email",
			"users",
			"Failed to find user for password reset",
			err,
		)
		return nil, repoErr
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, user.ID, hashedPassword)
	if err != nil {
		repoErr := errorcustom.NewRepositoryError(
			"update_password",
			"users",
			"Failed to reset user password",
			err,
		)
		return nil, repoErr
	}

	// Send password reset confirmation email
	// if s.emailService != nil {
	// 	go func() {
	// 		if err := s.emailService.SendPasswordResetConfirmation(context.Background(), user.Email, user.Name); err != nil {
	// 			s.logger.Error("Failed to send password reset confirmation email")
	// 		}
	// 	}()
	// }

	return &account.ResetPasswordRes{
		Success: true,
		Message: "Password has been reset successfully",
	}, nil
}
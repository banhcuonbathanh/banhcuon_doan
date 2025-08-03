// internal/account/account_service_email.go
package account_service

import (
	"context"
	"fmt"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/proto_qr/account"
)



// VerifyEmail handles email verification requests
func (s *ServiceStruct) VerifyEmail(ctx context.Context, req *account.VerifyEmailReq) (*account.VerifyEmailRes, error) {
	if s.tokenMaker == nil {
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"VerifyEmail",
			"Email verification functionality not available",
			fmt.Errorf("missing tokenMaker dependency"),
			false,
		)
		return nil, serviceErr
	}

	// Validate verification token
	email, err := s.userRepo.ValidateVerificationToken(ctx, req.VerificationToken)
	if err != nil {
		// Check if it's token validation specific error
		if strings.Contains(err.Error(), "expired") {
			return nil, errorcustom.NewInvalidTokenError("verification_token", "verification token has expired")
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "not found") {
			return nil, errorcustom.NewInvalidTokenError("verification_token", "verification token is invalid")
		}
		// For other repository errors
		repoErr := errorcustom.NewRepositoryError(
			"validate_verification_token",
			"email_verification_tokens",
			"Failed to validate verification token",
			err,
		)
		return nil, repoErr
	}

	// Mark email as verified
	err = s.userRepo.MarkEmailAsVerified(ctx, email)
	if err != nil {
		repoErr := errorcustom.NewRepositoryError(
			"mark_email_verified",
			"users",
			"Failed to mark email as verified",
			err,
		)
		return nil, repoErr
	}

	return &account.VerifyEmailRes{
		Success: true,
		Message: "Email verified successfully",
	}, nil
}

// ResendVerification handles resend verification email requests
func (s *ServiceStruct) ResendVerification(ctx context.Context, req *account.ResendVerificationReq) (*account.ResendVerificationRes, error) {
	if s.tokenMaker == nil || s.emailService == nil {
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"ResendVerification",
			"Email verification functionality not available",
			fmt.Errorf("missing tokenMaker or emailService dependency"),
			false,
		)
		return nil, serviceErr
	}

	// Check if user exists
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Check if it's a user not found error using string matching
		if strings.Contains(err.Error(), "not found") {
			// Don't reveal that email doesn't exist for security
			// Return success response regardless
			return &account.ResendVerificationRes{
				Success: true,
				Message: "If the email exists and is unverified, a verification email has been sent",
			}, nil
		}
		// For other repository errors
		repoErr := errorcustom.NewRepositoryError(
			"find_user_by_email",
			"users",
			"Failed to find user for resend verification",
			err,
		)
		return nil, repoErr
	}

	// Check if email verification is needed
	// Note: Since model.Account doesn't have EmailVerified field,
	// we'll let the repository method handle duplicate verification attempts
	// The MarkEmailAsVerified method should be idempotent

	// Create new verification token
	verificationToken, err := s.tokenMaker.CreateVerificationToken(user.Email)
	if err != nil {
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"ResendVerification",
			"Failed to create verification token",
			err,
			false,
		)
		return nil, serviceErr
	}

	// Store verification token
	err = s.userRepo.StoreVerificationToken(ctx, user.Email, verificationToken)
	if err != nil {
		repoErr := errorcustom.NewRepositoryError(
			"store_verification_token",
			"email_verification_tokens",
			"Failed to store verification token",
			err,
		)
		return nil, repoErr
	}

	// Send verification email
	go func() {
		if err := s.emailService.SendVerificationEmail(context.Background(), user.Email, verificationToken); err != nil {
			s.logger.Error("Failed to send verification email")
		}
	}()

	return &account.ResendVerificationRes{
		Success: true,
		Message: "Verification email has been sent",
	}, nil
}
// ResendVerification handles resend verification email requests

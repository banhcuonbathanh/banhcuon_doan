// internal/account/account_service_auth.go
package account_service

import (
	"context"
	"errors"
	"time"

	"english-ai-full/internal/model"
	"english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Register handles user registration
func (s *ServiceStruct) Register(ctx context.Context, req *account.RegisterReq) (*account.RegisterRes, error) {
	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword := req.Password
	if s.passwordHash != nil {
		hashed, err := s.passwordHash.HashPassword(req.Password)
		if err != nil {
			return nil, pkgerrors.WithStack(err)
		}
		hashedPassword = hashed
	}

	user, err := s.userRepo.Register(ctx, model.Account{
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Send verification email if email service is available
	if s.emailService != nil && s.tokenMaker != nil {
		verificationToken, err := s.tokenMaker.CreateVerificationToken(user.Email)
		if err != nil {
			s.logger.Error("Failed to create verification token")
		} else {
			// Store verification token
			if err := s.userRepo.StoreVerificationToken(ctx, user.Email, verificationToken); err != nil {
				s.logger.Error("Failed to store verification token")
			} else {
				go func() {
					if err := s.emailService.SendVerificationEmail(context.Background(), user.Email, verificationToken); err != nil {
						s.logger.Error("Failed to send verification email")
					}
				}()
			}
		}
	}

	return &account.RegisterRes{
		Success: true,
		Id:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
	}, nil
}

// Login handles user authentication
func (s *ServiceStruct) Login(ctx context.Context, req *account.LoginReq) (*account.AccountRes, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Verify password
	var isValidPassword bool
	if s.passwordHash != nil {
		isValidPassword = s.passwordHash.ComparePassword(user.Password, req.Password)
	} else {
		// Fallback to utils.Compare for backward compatibility
		isValidPassword = utils.Compare(user.Password, req.Password)
	}

	if !isValidPassword {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token if token maker is available
	var token string
	if s.tokenMaker != nil {
		token, err = s.tokenMaker.CreateToken(user)
		if err != nil {
			s.logger.Error("Failed to create token")
		}
	}

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
		Token: token,
	}, nil
}

// Logout handles user logout
func (s *ServiceStruct) Logout(ctx context.Context, req *account.LogoutReq) (*account.LogoutRes, error) {
	// For JWT-based authentication, logout is typically handled client-side
	// by removing the token. Server-side logout would require token blacklisting.
	// This is a placeholder implementation.
	
	s.logger.Info("User logout")
	
	return &account.LogoutRes{
		Success: true,
		Message: "Successfully logged out",
	}, nil
}

// RefreshToken handles token refresh requests
func (s *ServiceStruct) RefreshToken(ctx context.Context, req *account.RefreshTokenReq) (*account.RefreshTokenRes, error) {
	if s.tokenMaker == nil {
		return nil, errors.New("token functionality not available")
	}

	// Validate refresh token
	user, err := s.tokenMaker.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Create new access token
	accessToken, err := s.tokenMaker.CreateToken(*user)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Create new refresh token
	refreshToken, err := s.tokenMaker.CreateRefreshToken(*user)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	return &account.RefreshTokenRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken handles token validation requests
func (s *ServiceStruct) ValidateToken(ctx context.Context, req *account.ValidateTokenReq) (*account.ValidateTokenRes, error) {
	if s.tokenMaker == nil {
		return nil, errors.New("token functionality not available")
	}

	user, err := s.tokenMaker.VerifyToken(req.Token)
	if err != nil {
		return &account.ValidateTokenRes{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	return &account.ValidateTokenRes{
		Valid:  true,
		UserId: user.ID,
		// Note: No Email or Role fields available in the protobuf definition
		// You'll need to add these to your proto file if needed
	}, nil
}
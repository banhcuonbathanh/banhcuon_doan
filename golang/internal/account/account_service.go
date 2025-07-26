package account

import (
	"context"
	"errors"
	"time"

	"english-ai-full/internal/model"
	"english-ai-full/internal/proto_qr/account"
	logg "english-ai-full/logger"
	"english-ai-full/utils"

	pkgerrors "github.com/pkg/errors"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServiceStruct struct {
	userRepo      AccountRepositoryInterface
	logger        *logg.Logger
	tokenMaker    TokenMakerInterface
	passwordHash  PasswordHasherInterface
	emailService  EmailServiceInterface
	account.UnimplementedAccountServiceServer
}

// Updated constructor to accept interfaces for better testability
func NewAccountService(userRepo AccountRepositoryInterface, tokenMaker TokenMakerInterface, passwordHash PasswordHasherInterface, emailService EmailServiceInterface) *ServiceStruct {
	return &ServiceStruct{
		userRepo:     userRepo,
		tokenMaker:   tokenMaker,
		passwordHash: passwordHash,
		emailService: emailService,
		logger:       logg.NewLogger(),
	}
}

// Legacy constructor for backward compatibility
func NewAccountServiceLegacy(userRepo *Repository) *ServiceStruct {
	return &ServiceStruct{
		userRepo: userRepo,
		logger:   logg.NewLogger(),
	}
}

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
		OwnerID:   req.BranchId,
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
			s.logger.Error("Failed to create verification token", )
		} else {
			// Store verification token
			if err := s.userRepo.StoreVerificationToken(ctx, user.Email, verificationToken); err != nil {
				s.logger.Error("Failed to store verification token", )
			} else {
				go func() {
					if err := s.emailService.SendVerificationEmail(context.Background(), user.Email, verificationToken); err != nil {
						s.logger.Error("Failed to send verification email", )
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
			s.logger.Error("Failed to create token", )
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

func (s *ServiceStruct) Logout(ctx context.Context, req *account.LogoutReq) (*account.LogoutRes, error) {
	// For JWT-based authentication, logout is typically handled client-side
	// by removing the token. Server-side logout would require token blacklisting.
	// This is a placeholder implementation.
	
	s.logger.Info("User logout", )
	
	return &account.LogoutRes{
		Success: true,
		Message: "Successfully logged out",
	}, nil
}

func (s *ServiceStruct) FindByEmail(ctx context.Context, req *account.FindByEmailReq) (*account.AccountRes, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, ErrorUserNotFound
		}
		return nil, pkgerrors.WithStack(err)
	}

	return &account.AccountRes{Account: &account.Account{
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
	}}, nil
}

func (s *ServiceStruct) FindByID(ctx context.Context, req *account.FindByIDReq) (*account.FindByIDRes, error) {
	user, err := s.userRepo.FindByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, pkgerrors.WithStack(ErrorUserNotFound)
		}
		return nil, pkgerrors.WithStack(err)
	}

	return &account.FindByIDRes{Account: &account.Account{
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
	}}, nil
}

func (s *ServiceStruct) FindAllUsers(ctx context.Context, req *emptypb.Empty) (*account.AccountList, error) {
	users, err := s.userRepo.FindAllUsers(ctx)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var accountList []*account.Account
	for _, user := range users {
		accountList = append(accountList, &account.Account{
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
		})
	}

	return &account.AccountList{
		Accounts: accountList,
		Total:    int32(len(accountList)),
	}, nil
}

func (s *ServiceStruct) UpdateUser(ctx context.Context, req *account.UpdateUserReq) (*account.AccountRes, error) {
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
		if errors.Is(err, ErrorUserNotFound) {
			return nil, pkgerrors.WithStack(ErrorUserNotFound)
		}
		return nil, pkgerrors.WithStack(err)
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
	}, nil
}

func (s *ServiceStruct) DeleteUser(ctx context.Context, req *account.DeleteAccountReq) (*account.DeleteAccountRes, error) {
	// Get user info before deletion for email notification
	var user model.Account
	if s.emailService != nil {
		var err error
		user, err = s.userRepo.FindByID(ctx, req.UserID)
		if err != nil && !errors.Is(err, ErrorUserNotFound) {
			s.logger.Error("Failed to get user info before deletion")
		}
	}

	err := s.userRepo.DeleteUser(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, pkgerrors.WithStack(ErrorUserNotFound)
		}
		return nil, pkgerrors.WithStack(err)
	}

	// Send account deactivation email if email service is available
	if s.emailService != nil && user.Email != "" {
		go func() {
			if err := s.emailService.SendAccountDeactivationEmail(context.Background(), user.Email, user.Name); err != nil {
				s.logger.Error("Failed to send account deactivation email", )
			}
		}()
	}

	return &account.DeleteAccountRes{
		Success: true,
	}, nil
}

// Password management methods
func (s *ServiceStruct) ChangePassword(ctx context.Context, req *account.ChangePasswordReq) (*account.ChangePasswordRes, error) {
	// Verify current password
	user, err := s.userRepo.FindByID(ctx, req.UserId)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var isValidPassword bool
	if s.passwordHash != nil {
		isValidPassword = s.passwordHash.ComparePassword(user.Password, req.CurrentPassword)
	} else {
		isValidPassword = utils.Compare(user.Password, req.CurrentPassword)
	}

	if !isValidPassword {
		return nil, errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword := req.NewPassword
	if s.passwordHash != nil {
		hashed, err := s.passwordHash.HashPassword(req.NewPassword)
		if err != nil {
			return nil, pkgerrors.WithStack(err)
		}
		hashedPassword = hashed
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, req.UserId, hashedPassword)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Send password changed notification
	if s.emailService != nil {
		go func() {
			if err := s.emailService.SendPasswordChangedNotification(context.Background(), user.Email, user.Name); err != nil {
				s.logger.Error("Failed to send password changed notification", )
			}
		}()
	}

	return &account.ChangePasswordRes{
		Success: true,
		Message: "Password changed successfully",
	}, nil
}

func (s *ServiceStruct) ForgotPassword(ctx context.Context, req *account.ForgotPasswordReq) (*account.ForgotPasswordRes, error) {
	// Check if user exists
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			// Don't reveal that email doesn't exist for security
			return &account.ForgotPasswordRes{
				Success: true,
				Message: "If the email exists, a password reset link has been sent",
			}, nil
		}
		return nil, pkgerrors.WithStack(err)
	}

	if s.tokenMaker == nil || s.emailService == nil {
		return nil, errors.New("password reset functionality not available")
	}

	// Create reset token
	resetToken, err := s.tokenMaker.CreateResetToken(user.Email)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Store reset token
	err = s.userRepo.StoreResetToken(ctx, user.Email, resetToken)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Send reset email
	go func() {
		if err := s.emailService.SendPasswordResetEmail(context.Background(), user.Email, resetToken); err != nil {
			s.logger.Error("Failed to send password reset email", )
		}
	}()

	return &account.ForgotPasswordRes{
		Success: true,
		Message: "Password reset link has been sent to your email",
	}, nil
}

func (s *ServiceStruct) ResetPassword(ctx context.Context, req *account.ResetPasswordReq) (*account.ResetPasswordRes, error) {
	if s.tokenMaker == nil {
		return nil, errors.New("password reset functionality not available")
	}

	// Validate reset token
	email, err := s.userRepo.ValidateResetToken(ctx, req.Token)
	if err != nil {
		return nil, errors.New("invalid or expired reset token")
	}

	// Hash new password
	hashedPassword := req.NewPassword
	if s.passwordHash != nil {
		hashed, err := s.passwordHash.HashPassword(req.NewPassword)
		if err != nil {
			return nil, pkgerrors.WithStack(err)
		}
		hashedPassword = hashed
	}

	// Get user to update password
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, user.ID, hashedPassword)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	return &account.ResetPasswordRes{
		Success: true,
		Message: "Password has been reset successfully",
	}, nil
}

// Account verification methods
func (s *ServiceStruct) VerifyEmail(ctx context.Context, req *account.VerifyEmailReq) (*account.VerifyEmailRes, error) {
	if s.tokenMaker == nil {
		return nil, errors.New("email verification functionality not available")
	}

	// Validate verification token - try one of these field names:
	email, err := s.userRepo.ValidateVerificationToken(ctx, req.VerificationToken) // or
	// email, err := s.userRepo.ValidateVerificationToken(ctx, req.Code)           // or
	// email, err := s.userRepo.ValidateVerificationToken(ctx, req.VerifyToken)    // or
	if err != nil {
		return nil, errors.New("invalid or expired verification token")
	}

	// Mark email as verified
	err = s.userRepo.MarkEmailAsVerified(ctx, email)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	return &account.VerifyEmailRes{
		Success: true,
		Message: "Email verified successfully",
	}, nil
}

func (s *ServiceStruct) ResendVerification(ctx context.Context, req *account.ResendVerificationReq) (*account.ResendVerificationRes, error) {
	if s.tokenMaker == nil || s.emailService == nil {
		return nil, errors.New("email verification functionality not available")
	}

	// Check if user exists
	_, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, errors.New("email not found")
		}
		return nil, pkgerrors.WithStack(err)
	}

	// Create new verification token
	verificationToken, err := s.tokenMaker.CreateVerificationToken(req.Email)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Store verification token
	err = s.userRepo.StoreVerificationToken(ctx, req.Email, verificationToken)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	// Send verification email
	go func() {
		if err := s.emailService.SendVerificationEmail(context.Background(), req.Email, verificationToken); err != nil {
			s.logger.Error("Failed to send verification email", )
		}
	}()

	return &account.ResendVerificationRes{
		Success: true,
		Message: "Verification email has been sent",
	}, nil
}

func (s *ServiceStruct) UpdateAccountStatus(ctx context.Context, req *account.UpdateAccountStatusReq) (*account.UpdateAccountStatusRes, error) {
	err := s.userRepo.UpdateAccountStatus(ctx, req.UserId, req.Status)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, pkgerrors.WithStack(ErrorUserNotFound)
		}
		return nil, pkgerrors.WithStack(err)
	}

	return &account.UpdateAccountStatusRes{
		Success: true,
		Message: "Account status updated successfully",
	}, nil
}

// Enhanced search and filtering methods
func (s *ServiceStruct) FindByRole(ctx context.Context, req *account.FindByRoleReq) (*account.AccountList, error) {
	users, err := s.userRepo.FindByRole(ctx, req.Role)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var accountList []*account.Account
	for _, user := range users {
		accountList = append(accountList, &account.Account{
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
		})
	}

	return &account.AccountList{
		Accounts: accountList,
		Total:    int32(len(accountList)),
	}, nil
}

func (s *ServiceStruct) FindByBranch(ctx context.Context, req *account.FindByBranchReq) (*account.AccountList, error) {
	users, err := s.userRepo.FindByBranchID(ctx, req.BranchId)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var accountList []*account.Account
	for _, user := range users {
		accountList = append(accountList, &account.Account{
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
		})
	}

	return &account.AccountList{
		Accounts: accountList,
		Total:    int32(len(accountList)),
	}, nil
}

// func (s *ServiceStruct) SearchUsers(ctx context.Context, req *account.SearchUsersReq) (*account.AccountList, error) {
// 	// Extract pagination info
// 	var page, pageSize int32
// 	if req.Pagination != nil {
// 		page = req.Pagination.Page
// 		pageSize = req.Pagination.PageSize
// 	} else {
// 		// Default values if pagination is not provided
// 		page = 1
// 		pageSize = 10
// 	}

// 	// Extract sort info
// 	var sortBy, sortOrder string
// 	if req.Sort != nil {
// 		sortBy = req.Sort.SortBy
// 		sortOrder = req.Sort.SortOrder
// 	}

// 	users, totalCount, err := s.userRepo.SearchUsers(ctx, req.Query, req.Role, req.BranchId, req.StatusFilter, page, pageSize, sortBy, sortOrder)
// 	if err != nil {
// 		return nil, pkgerrors.WithStack(err)
// 	}

// 	var accountList []*account.Account
// 	for _, user := range users {
// 		accountList = append(accountList, &account.Account{
// 			Id:        user.Id,
// 			BranchId:  user.BranchId,
// 			Name:      user.Name,
// 			Email:     user.Email,
// 			Avatar:    user.Avatar,
// 			Title:     user.Title,
// 			Role:      user.Role, // Remove string() conversion since user.Role is already a string
// 			OwnerId:   user.OwnerId,
// 			CreatedAt: user.CreatedAt, // Direct assignment since both are *timestamppb.Timestamp
// 			UpdatedAt: user.UpdatedAt, // Direct assignment since both are *timestamppb.Timestamp
// 		})
// 	}

// 	// Calculate pagination info
// 	totalPages := int32((totalCount + int64(pageSize) - 1) / int64(pageSize)) // Ceiling division
// 	hasNext := page < totalPages
// 	hasPrev := page > 1

// 	// Create pagination info
// 	paginationInfo := &account.PaginationInfo{
// 		Page:       page,
// 		PageSize:   pageSize,
// 		TotalPages: totalPages,
// 		HasNext:    hasNext,
// 		HasPrev:    hasPrev,
// 	}

// 	return &account.AccountList{
// 		Accounts:   accountList,
// 		Total:      int32(totalCount),
// 		Pagination: paginationInfo, // Include pagination info in response
// 	}, nil
// }
// Token management methods
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

// Business logic helper methods
func (s *ServiceStruct) ValidateUserCredentials(ctx context.Context, email, password string) (model.Account, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return model.Account{}, pkgerrors.WithStack(err)
	}

	var isValidPassword bool
	if s.passwordHash != nil {
		isValidPassword = s.passwordHash.ComparePassword(user.Password, password)
	} else {
		isValidPassword = utils.Compare(user.Password, password)
	}

	if !isValidPassword {
		return model.Account{}, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *ServiceStruct) DeactivateUser(ctx context.Context, userID int64) error {
	return s.userRepo.UpdateAccountStatus(ctx, userID, "inactive")
}

func (s *ServiceStruct) GetUsersByBranch(ctx context.Context, branchID int64) ([]model.Account, error) {
	return s.userRepo.FindByBranchID(ctx, branchID)
}

// Helper method to convert model.Account to protobuf Account
func (s *ServiceStruct) modelToProto(user model.Account) *account.Account {
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
	}
}

// Compile-time check to ensure ServiceStruct implements AccountServiceInterface



// new 12121212
func (s *ServiceStruct) SearchUsers(ctx context.Context, req *account.SearchUsersReq) (*account.SearchUsersRes, error) {
	// Extract pagination info
	var page, pageSize int32
	if req.Pagination != nil {
		page = req.Pagination.Page
		pageSize = req.Pagination.PageSize
	} else {
		// Default values if pagination is not provided
		page = 1
		pageSize = 10
	}

	// Extract sort info
	var sortBy, sortOrder string
	if req.Sort != nil {
		sortBy = req.Sort.SortBy
		sortOrder = req.Sort.SortOrder
	}

	users, totalCount, err := s.userRepo.SearchUsers(ctx, req.Query, req.Role, req.BranchId, req.StatusFilter, page, pageSize, sortBy, sortOrder)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var accounts []*account.Account
	// Fix: Use index-based loop or range over pointers to avoid copying the struct
	for i := range users {
		user := &users[i] // Get pointer to avoid copying
		accounts = append(accounts, &account.Account{
			Id:        user.Id,
			BranchId:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerId:   user.OwnerId,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	// Calculate pagination info
	totalPages := int32((totalCount + int64(pageSize) - 1) / int64(pageSize)) // Ceiling division
	hasNext := page < totalPages
	hasPrev := page > 1

	// Create pagination info
	paginationInfo := &account.PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}

	// Return SearchUsersRes with the accounts slice directly
	return &account.SearchUsersRes{
		Accounts:   accounts,
		Total:      int32(totalCount),
		Pagination: paginationInfo,
	}, nil
}
// new 12121212

	

var _ AccountServiceInterface = (*ServiceStruct)(nil)
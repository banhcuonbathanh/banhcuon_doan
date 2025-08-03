// internal/account/account_service_main.go
package account_service

import (
	"context"
	account_interface "english-ai-full/internal/account"
	"english-ai-full/internal/model"
	"english-ai-full/internal/proto_qr/account"
	logg "english-ai-full/logger"
	"english-ai-full/utils"
	"errors"
pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ServiceStruct is the main service structure that implements all account-related functionality
type ServiceStruct struct {
	userRepo      account_interface.AccountRepositoryInterface
	logger        *logg.Logger
	tokenMaker    account_interface.TokenMakerInterface
	passwordHash  account_interface.PasswordHasherInterface
	emailService  account_interface.EmailServiceInterface
	account.UnimplementedAccountServiceServer
}

// NewAccountService creates a new account service with all dependencies
func NewAccountService(
	userRepo account_interface.AccountRepositoryInterface,
	tokenMaker account_interface.TokenMakerInterface,
	passwordHash account_interface.PasswordHasherInterface,
	emailService account_interface.EmailServiceInterface,
) *ServiceStruct {
	return &ServiceStruct{
		userRepo:     userRepo,
		tokenMaker:   tokenMaker,
		passwordHash: passwordHash,
		emailService: emailService,
		logger:       logg.NewLogger(),
	}
}

// NewAccountServiceLegacy creates a service with minimal dependencies for backward compatibility
func NewAccountServiceLegacy(userRepo account_interface.AccountRepositoryInterface) *ServiceStruct {
	return &ServiceStruct{
		userRepo: userRepo,
		logger:   logg.NewLogger(),
	}
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

// Compile-time check to ensure ServiceStruct implements AccountServiceInterface
var _ account_interface.AccountServiceInterface = (*ServiceStruct)(nil)
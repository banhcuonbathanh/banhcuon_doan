// internal/account/account_service/account_service.go
package account_service

import (
	"context"

	"time"

	error_custom "english-ai-full/error_custom"
	account_interface "english-ai-full/internal/account"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/proto_qr/account"
	logg "english-ai-full/logger"

	utils_config "english-ai-full/utils/config"

	"github.com/go-playground/validator"
	"github.com/pkg/errors"
)

// AccountService implements the main service structure with all account-related functionality
type AccountService struct {
	userRepo      account_interface.AccountRepositoryInterface
	logger        *logg.SpecializedLogger
	tokenMaker    account_interface.TokenMakerInterface
	passwordHash  account_interface.PasswordHasherInterface
	emailService  account_interface.EmailServiceInterface
	errorHandler  *error_custom.ServiceErrorManager
	config        *utils_config.Config
	domain        string
	account.UnimplementedAccountServiceServer
		validator     *validator.Validate
}

// NewAccountService creates a new account service with all dependencies
func NewAccountService(
	userRepo account_interface.AccountRepositoryInterface,
	tokenMaker account_interface.TokenMakerInterface,
	passwordHash account_interface.PasswordHasherInterface,
	emailService account_interface.EmailServiceInterface,
) *AccountService {
	return &AccountService{
		userRepo:     userRepo,
		tokenMaker:   tokenMaker,
		passwordHash: passwordHash,
		emailService: emailService,
		logger:       logg.NewSpecializedServiceLogger(),
		errorHandler: error_custom.NewServiceErrorManager(),
		config:       utils_config.GetConfig(),
		domain:       "account",
				validator:    validator.New(),
	}
}

// NewAccountServiceLegacy creates a service with minimal dependencies for backward compatibility


// ===== CORE SERVICE METHODS =====

// CreateUser creates a new user account with comprehensive validation and error handling
func (s *AccountService) CreateUser(ctx context.Context, req *account.AccountReq) (*account.Account, error) {
	const operation = "create_user"
	startTime := time.Now()
	
	operationCtx := s.buildOperationContext(operation, map[string]interface{}{
		"email": req.Email,
		"role":  req.Role,
	})

	// Context cancellation check
	if err := ctx.Err(); err != nil {
		return nil, s.handleServiceError(err, operation, operationCtx, &startTime)
	}

	// Input validation
	if err := s.validator.Struct(req); err != nil {
		validationErr := s.formatValidationError(err, req.Email)
		operationCtx["validation_error"] = validationErr.Error()
		return nil, s.handleServiceError(validationErr, operation, operationCtx, &startTime)
	}

	// Convert protobuf request to DTO
	userDTO := account_dto.Account{
		BranchID: req.BranchId,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     account_dto.Role(req.Role),
		OwnerID:  req.OwnerId,
		Status:   "active", // Set default status
	}

	// Hash password if password hasher is available
	if s.passwordHash != nil {
		hashedPassword, err := s.passwordHash.HashPassword(userDTO.Password)
		if err != nil {
			operationCtx["hash_error"] = "failed to hash password"
			return nil, s.handleServiceError(
				errors.Wrap(err, "failed to hash password"), 
				operation, operationCtx, &startTime,
			)
		}
		userDTO.Password = hashedPassword
	}

	// Call repository to create user with DTO
	createdUser, err := s.userRepo.CreateUser(ctx, userDTO)
	if err != nil {
		operationCtx["repository_error"] = "failed to create user in repository"
		return nil, s.handleServiceError(err, operation, operationCtx, &startTime)
	}

	// Log success
	operationCtx["user_id"] = createdUser.ID
	s.handleServiceSuccess(operation, operationCtx, startTime)

	// Send welcome email if email service is available
	if s.emailService != nil {
		go func() {
			emailCtx := context.Background()
			if err := s.emailService.SendWelcomeEmail(emailCtx, createdUser.Email, createdUser.Name); err != nil {
				s.logger.LogServiceCall("account", "send_welcome_email", false, err, map[string]interface{}{
					"user_id":    createdUser.ID,
					"user_email": createdUser.Email,
				})
			}
		}()
	}

	// Convert DTO to protobuf message before returning
	return s.convertDTOToProto(&createdUser), nil
}

// AuthenticateUser authenticates a user with email and password
// func (s *AccountService) AuthenticateUser(ctx context.Context, email, password string) (account_dto.Account, string, error) {
// 	const operation = "authenticate_user"
// 	startTime := time.Now()
	
// 	operationCtx := s.buildOperationContext(operation, map[string]interface{}{
// 		"email": email,
// 	})

// 	// Context cancellation check
// 	if err := ctx.Err(); err != nil {
// 		return account_dto.Account{}, "", s.handleServiceError(err, operation, operationCtx, &startTime)
// 	}

// 	// Input validation
// 	if email == "" || password == "" {
// 		err := error_custom.NewValidationError("account", "credentials", "email and password are required", email)
// 		return account_dto.Account{}, "", s.handleServiceError(err, operation, operationCtx, &startTime)
// 	}

// 	// Get user by email (assuming this method exists in repository)
// 	user, err := s.userRepo.GetUserByEmail(ctx, email)
// 	if err != nil {
// 		operationCtx["repository_error"] = "user not found or database error"
// 		return account_dto.Account{}, "", s.handleServiceError(err, operation, operationCtx, &startTime)
// 	}

// 	// Verify password if password hasher is available
// 	if s.passwordHash != nil {
// 		if err := s.passwordHash.CheckPassword(password, user.Password); err != nil {
// 			operationCtx["auth_error"] = "invalid password"
// 			authErr := error_custom.NewAuthenticationError("invalid credentials", email)
// 			return account_dto.Account{}, "", s.handleServiceError(authErr, operation, operationCtx, &startTime)
// 		}
// 	}

// 	// Check if user is active
// 	if user.Status != "active" {
// 		operationCtx["status_error"] = fmt.Sprintf("user status is %s", user.Status)
// 		statusErr := error_custom.NewAuthenticationError("account is not active", email)
// 		return account_dto.Account{}, "", s.handleServiceError(statusErr, operation, operationCtx, &startTime)
// 	}

// 	// Generate token if token maker is available
// 	var token string
// 	if s.tokenMaker != nil {
// 		token, err = s.tokenMaker.CreateToken(user.Email, user.Role, time.Hour*24) // 24 hours token
// 		if err != nil {
// 			operationCtx["token_error"] = "failed to generate token"
// 			return account_dto.Account{}, "", s.handleServiceError(
// 				errors.Wrap(err, "failed to generate token"), 
// 				operation, operationCtx, &startTime,
// 			)
// 		}
// 	}

// 	// Log success
// 	operationCtx["user_id"] = user.ID
// 	s.handleServiceSuccess(operation, operationCtx, startTime)

// 	return user, token, nil
// }

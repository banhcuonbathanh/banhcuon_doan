// internal/account/account_service/account_service.go
package account_service

import (
	"context"

	"strings"
	"time"

	account_interface "english-ai-full/internal/account"
	"english-ai-full/internal/account/account_dto"
	error_custom "english-ai-full/error_custom"
	"english-ai-full/internal/proto_qr/account"
	logg "english-ai-full/logger"
	"english-ai-full/utils"
	utils_config "english-ai-full/utils/config"

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
	}
}

// NewAccountServiceLegacy creates a service with minimal dependencies for backward compatibility
func NewAccountServiceLegacy(userRepo account_interface.AccountRepositoryInterface) *AccountService {
	return &AccountService{
		userRepo:     userRepo,
		logger:       logg.NewSpecializedServiceLogger(),
		errorHandler: error_custom.NewServiceErrorManager(),
		config:       utils_config.GetConfig(),
		domain:       "account",
	}
}

// ===== HELPER METHODS =====

// buildOperationContext creates standardized context for logging and error handling
func (s *AccountService) buildOperationContext(operation string, userInfo map[string]interface{}) map[string]interface{} {
	ctx := map[string]interface{}{
		"service":   "account",
		"operation": operation,
		"domain":    s.domain,
	}
	
	// Merge user-specific context
	for key, value := range userInfo {
		ctx[key] = value
	}
	
	return ctx
}

// validateUserInput performs comprehensive input validation
func (s *AccountService) validateUserInput(user account_dto.Account) error {
	var validationErrors []string

	// Email validation
	if user.Email == "" {
		validationErrors = append(validationErrors, "email is required")
	} else if !utils.IsValidEmail(user.Email) {
		validationErrors = append(validationErrors, "invalid email format")
	}

	// Name validation
	if user.Name == "" {
		validationErrors = append(validationErrors, "name is required")
	} else if len(strings.TrimSpace(user.Name)) < 2 {
		validationErrors = append(validationErrors, "name must be at least 2 characters")
	}

	// Password validation
	if user.Password == "" {
		validationErrors = append(validationErrors, "password is required")
	} else if len(user.Password) < 8 {
		validationErrors = append(validationErrors, "password must be at least 8 characters")
	}

	// Role validation
	if user.Role == "" {
		validationErrors = append(validationErrors, "role is required")
	}

	if len(validationErrors) > 0 {
		return error_custom.NewValidationError("account", "input", strings.Join(validationErrors, "; "), user.Email)
	}

	return nil
}


// handleServiceError wraps errors with service context and logs them
func (s *AccountService) handleServiceError(err error, operation string, context map[string]interface{}, startTime *time.Time) error {
	if err == nil {
		return nil
	}

	// Add timing information if provided
	if startTime != nil {
		context["duration_ms"] = time.Since(*startTime).Milliseconds()
	}

	// Add service context
	context["service"] = "account"
	context["operation"] = operation
	context["layer"] = "service"

	// Log the error
	s.logger.LogServiceCall("account", operation, false, err, context)

	// Convert to APIError and add service layer context
	apiErr := error_custom.ConvertToAPIError(err)
	if apiErr == nil {
		apiErr = error_custom.NewAPIError(
			"ACCOUNT_SERVICE_ERROR",
			"Account service operation failed",
			500,
		)
	}

	// Add service layer context
	apiErr.WithDomain(s.domain).
		WithLayer("service").
		WithOperation(operation).
		WithCause(err)

	// Add additional context
	for k, v := range context {
		apiErr.WithDetail(k, v)
	}

	return apiErr
}
// handleServiceSuccess logs successful operations
func (s *AccountService) handleServiceSuccess(operation string, context map[string]interface{}, startTime time.Time) {
	context["duration_ms"] = time.Since(startTime).Milliseconds()
	context["success"] = true
	s.logger.LogServiceCall("account", operation, true, nil, context)
}
// ===== CORE SERVICE METHODS =====

// CreateUser creates a new user account with comprehensive validation and error handling
func (s *AccountService) CreateUser(ctx context.Context, user account_dto.Account) (account_dto.Account, error) {
	const operation = "create_user"
	startTime := time.Now()
	
	operationCtx := s.buildOperationContext(operation, map[string]interface{}{
		"email": user.Email,
		"role":  string(user.Role),
	})

	// Context cancellation check
	if err := ctx.Err(); err != nil {
		return account_dto.Account{}, s.handleServiceError(err, operation, operationCtx, &startTime)
	}

	// Input validation
	if err := s.validateUserInput(user); err != nil {
		operationCtx["validation_error"] = err.Error()
		return account_dto.Account{}, s.handleServiceError(err, operation, operationCtx, &startTime)
	}

	// Hash password if password hasher is available
	if s.passwordHash != nil {
		hashedPassword, err := s.passwordHash.HashPassword(user.Password)
		if err != nil {
			operationCtx["hash_error"] = "failed to hash password"
			return account_dto.Account{}, s.handleServiceError(
				errors.Wrap(err, "failed to hash password"), 
				operation, operationCtx, &startTime,
			)
		}
		user.Password = hashedPassword
	}

	// Set default status if not provided
	if user.Status == "" {
		user.Status = "active"
	}

	// Call repository to create user
	createdUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		operationCtx["repository_error"] = "failed to create user in repository"
		return account_dto.Account{}, s.handleServiceError(err, operation, operationCtx, &startTime)
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

	return createdUser, nil
}

// CreateUserProto creates a new user account using Proto request/response
func (s *AccountService) CreateUserProto(ctx context.Context, req *account_dto.CreateUserRequest) (*account_dto.CreateUserResponse, error) {
   const operation = "create_user_proto"
   startTime := time.Now()
   
   operationCtx := s.buildOperationContext(operation, map[string]interface{}{
   	"email": req.Email,
   	"role":  req.Role,
   })

   // Convert proto to DTO
   userDTO := account_dto.Account{
   	BranchID: req.BranchID,
   	Name:     req.Name,
   	Email:    req.Email,
   	Password: req.Password,
   	Avatar:   req.Avatar,
   	Title:    req.Title,
   	Role:     account_dto.Role(req.Role),
   	OwnerID:  req.OwnerID,
   }

   // Create user using DTO method
   createdUser, err := s.CreateUser(ctx, userDTO)
   if err != nil {
   	return nil, s.handleServiceError(err, operation, operationCtx, &startTime)
   }

   // Log success and convert to proto response
   s.handleServiceSuccess(operation, operationCtx, startTime)
   
   return &account_dto.CreateUserResponse{
   	BranchID: createdUser.BranchID,
   	Name:     createdUser.Name,
   	Email:    createdUser.Email,
   	Avatar:   createdUser.Avatar,
   	Title:    createdUser.Title,
   	Role:     string(createdUser.Role),
   	OwnerID:  createdUser.OwnerID,
   }, nil
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

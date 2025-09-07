package account_service

import (
	"context"
	"strings"
	"time"

	"english-ai-full/error_system"
	"english-ai-full/internal"

	account_interface "english-ai-full/internal/account"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger/core"
	utils_config "english-ai-full/utils/config"

	"github.com/go-playground/validator"

)

// AccountService implements the main service structure with all account-related functionality
type AccountService struct {
	userRepo      account_interface.AccountRepositoryInterface
	logger        *core.CoreLogger
	tokenMaker    account_interface.TokenMakerInterface
	passwordHash  account_interface.PasswordHasherInterface
	emailService  account_interface.EmailServiceInterface
	errorHandler *error_system.ServiceErrorHandler
	config        *utils_config.Config
	domain        string
	layerContext  *common.ServiceLayerContext
	validator     *validator.Validate
	account.UnimplementedAccountServiceServer
}

// NewAccountService creates a new account service with all dependencies
func NewAccountService(
	userRepo account_interface.AccountRepositoryInterface,
	tokenMaker account_interface.TokenMakerInterface,
	passwordHash account_interface.PasswordHasherInterface,
	emailService account_interface.EmailServiceInterface,
) *AccountService {
	// Create layer context first
	layerContext := common.NewServiceLayerContext("account", "account-service")
	
	// Create logger using layer context - it will be pre-configured
	logger := layerContext.NewServiceLogger()
	errorHandler := error_system.NewServiceErrorHandler(logger, "account")
	return &AccountService{
		userRepo:     userRepo,
		tokenMaker:   tokenMaker,
		passwordHash: passwordHash,
		emailService: emailService,
		logger:       logger,
	errorHandler: errorHandler,
		config:       utils_config.GetConfig(),
		domain:       "account",
		layerContext: layerContext,
		validator:    validator.New(),
	}
}

// CreateUser creates a new user account with comprehensive validation and error handling
func (s *AccountService) CreateUser(ctx context.Context, req *account.AccountReq) (*account.Account, error) {
	const operation = "create_user"
	startTime := time.Now()
	
	// Build operation context using service layer context
	operationCtx := s.layerContext.BuildOperationContext(operation, map[string]interface{}{
		"email": maskEmail(req.Email), // Use masked email for security
		"role":  req.Role,
		"branch_id": req.BranchId,
		"owner_id": req.OwnerId,
	})

	// Set operation in logger
	s.logger.SetOperation(operation)

	// Log operation start with enhanced context
	s.logger.Info(core.MsgOperationStarted, s.layerContext.MergeWithContext(map[string]interface{}{
		"operation":    operation,
		"target_user":  maskEmail(req.Email),
		"requested_role": req.Role,
		"branch_id":    req.BranchId,
	}))

	// Context cancellation check
	if err := ctx.Err(); err != nil {
		s.logger.ErrorWithCause(core.MsgContextError, core.CauseContextCancelled, 
			core.LayerService, operation, s.layerContext.MergeWithContext(operationCtx))

				appErr := s.errorHandler.Handle(err, operation, operationCtx)
		return nil, appErr
	}

	// Input validation
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Request validation failed", s.layerContext.MergeWithContext(map[string]interface{}{
			"error": err.Error(),
			"email": maskEmail(req.Email),
		}))
		appErr := s.errorHandler.Handle(err, operation, operationCtx)
		return nil, appErr
	}

	// Log validation success
	s.logger.Info("Request validation completed successfully", s.layerContext.MergeWithContext(map[string]interface{}{
		"operation": operation,
		"email": maskEmail(req.Email),
		"validation_duration_ms": time.Since(startTime).Milliseconds(),
	}))

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
		hashStartTime := time.Now()
		
		s.logger.Debug("Starting password hashing", s.layerContext.MergeWithContext(map[string]interface{}{
			"operation": operation,
			"email": maskEmail(req.Email),
		}))

		hashedPassword, err := s.passwordHash.HashPassword(userDTO.Password)
		if err != nil {
			operationCtx["hash_error"] = "failed to hash password"
			
			s.logger.ErrorWithCause("Password hashing failed", "password_hash_error",
				core.LayerService, operation, s.layerContext.MergeWithContext(map[string]interface{}{
					"email": maskEmail(req.Email),
					"hash_duration_ms": time.Since(hashStartTime).Milliseconds(),
					"error": err.Error(),
				}))
			
			// Use ServiceErrorManager to handle this as a system error
		appErr := s.errorHandler.Handle(err, operation, operationCtx)
			return nil, appErr
		}
		
		userDTO.Password = hashedPassword
		
		s.logger.Debug("Password hashing completed", s.layerContext.MergeWithContext(map[string]interface{}{
			"operation": operation,
			"email": maskEmail(req.Email),
			"hash_duration_ms": time.Since(hashStartTime).Milliseconds(),
		}))
	}

	// Log repository call attempt
	repoStartTime := time.Now()
	s.logger.Info("Calling repository to create user", s.layerContext.MergeWithContext(map[string]interface{}{
		"operation": operation,
		"target_layer": core.LayerRepository,
		"target_function": core.FuncCreateUser,
		"email": maskEmail(req.Email),
		"role": req.Role,
	}))

	// Call repository to create user with DTO
	createdUser, err := s.userRepo.CreateUser(ctx, userDTO)
	if err != nil {
		operationCtx["repository_error"] = "failed to create user in repository"
		operationCtx["repository_duration_ms"] = time.Since(repoStartTime).Milliseconds()
		
		// Enhanced error logging with repository context
		s.logger.ErrorWithDomainAndCause("Repository call failed", s.layerContext.Domain, 
			core.CauseDatabaseError, core.LayerService, operation, 
			s.layerContext.MergeWithContext(map[string]interface{}{
				"target_layer": core.LayerRepository,
				"target_function": core.FuncCreateUser,
				"email": maskEmail(req.Email),
				"repository_duration_ms": time.Since(repoStartTime).Milliseconds(),
				"total_duration_ms": time.Since(startTime).Milliseconds(),
				"error": err.Error(),
			}))
		
appErr := s.errorHandler.Handle(err, operation, operationCtx)
		return nil, appErr
	}

	// Log successful repository call
	repositoryDuration := time.Since(repoStartTime)
	s.logger.Info("Repository call completed successfully", s.layerContext.MergeWithContext(map[string]interface{}{
		"operation": operation,
		"target_layer": core.LayerRepository,
		"target_function": core.FuncCreateUser,
		"user_id": createdUser.ID,
		"email": maskEmail(createdUser.Email),
		"repository_duration_ms": repositoryDuration.Milliseconds(),
	}))

	// Update operation context with success details
	operationCtx["user_id"] = createdUser.ID
	operationCtx["created_at"] = createdUser.CreatedAt
	operationCtx["repository_duration_ms"] = repositoryDuration.Milliseconds()

	// Log main operation success
	totalDuration := time.Since(startTime)
	s.logger.Info(core.MsgOperationCompleted, s.layerContext.MergeWithContext(map[string]interface{}{
		"operation": operation,
		"user_id": createdUser.ID,
		"email": maskEmail(createdUser.Email),
		"role": createdUser.Role,
		"branch_id": createdUser.BranchID,
		"success": true,
		"duration_ms": totalDuration.Milliseconds(),
		"repository_duration_ms": repositoryDuration.Milliseconds(),
	}))

	// Send welcome email if email service is available (async)
	if s.emailService != nil {
		s.logger.Debug("Initiating welcome email send", s.layerContext.MergeWithContext(map[string]interface{}{
			"operation": "send_welcome_email",
			"user_id": createdUser.ID,
			"email": maskEmail(createdUser.Email),
			"async": true,
		}))

		go func() {
			emailStartTime := time.Now()
			emailCtx := context.Background()
			
			if err := s.emailService.SendWelcomeEmail(emailCtx, createdUser.Email, createdUser.Name); err != nil {
				s.logger.ErrorWithCause("Welcome email send failed", core.CauseEmailSendFailed,
					core.LayerService, "send_welcome_email", s.layerContext.MergeWithContext(map[string]interface{}{
						"user_id": createdUser.ID,
						"email": maskEmail(createdUser.Email),
						"email_duration_ms": time.Since(emailStartTime).Milliseconds(),
						"error": err.Error(),
					}))
			} else {
				s.logger.Info("Welcome email sent successfully", s.layerContext.MergeWithContext(map[string]interface{}{
					"operation": "send_welcome_email",
					"user_id": createdUser.ID,
					"email": maskEmail(createdUser.Email),
					"email_duration_ms": time.Since(emailStartTime).Milliseconds(),
					"success": true,
				}))
			}
		}()
	}

	// Final success handling
	s.handleServiceSuccess(operation, operationCtx, startTime)

	// Convert DTO to protobuf message before returning
	return s.convertDTOToProto(&createdUser), nil
}

// Helper function to mask email for logging
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "invalid_email"
	}
	
	username := parts[0]
	domain := parts[1]
	
	if len(username) <= 2 {
		return "**@" + domain
	}
	
	maskedUsername := username[:2] + strings.Repeat("*", len(username)-2)
	return maskedUsername + "@" + domain
}

// Updated handleServiceSuccess method to use enhanced logging
func (s *AccountService) handleServiceSuccess(operation string, operationCtx map[string]interface{}, startTime time.Time) {
	duration := time.Since(startTime)
	
	// Update context with success metrics
	successCtx := s.layerContext.MergeWithContext(operationCtx)
	successCtx["success"] = true
	successCtx["duration_ms"] = duration.Milliseconds()
	successCtx["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	
	s.logger.Info("Service operation completed successfully", successCtx)
}

// Helper method to determine error cause
func (s *AccountService) determineErrorCause(err error) string {
	errStr := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return core.CauseNetworkError
	case strings.Contains(errStr, "timeout"):
		return core.CauseTimeout
	default:
		return core.CauseServiceError
	}
}

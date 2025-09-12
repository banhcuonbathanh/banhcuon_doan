package account_service

import (

	"strings"
	"time"

	"english-ai-full/error_system"
	"english-ai-full/internal"


	account_interface "english-ai-full/internal/account"

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
	errorHandler  *error_system.ServiceErrorHandler
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


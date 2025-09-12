// internal/account/account_handler/account_handler.go
package account_handler

import (


	"github.com/go-playground/validator/v10"


	"english-ai-full/error_system"
	"english-ai-full/internal"

	"english-ai-full/logger/core"


	pb "english-ai-full/internal/proto_qr/account"
)

type AccountHandler struct {
	userClient      pb.AccountServiceClient
	errorHandler *error_system.HandlerErrorHandler
	logger          *core.CoreLogger
	layerContext    *common.HandlerLayerContext
	validator       *validator.Validate
}

// NewAccountHandler creates a new account handler with dependencies
func NewAccountHandler(userClient pb.AccountServiceClient) *AccountHandler {
	// Create layer context first
	layerContext := common.NewHandlerLayerContext("account", "account-service")
	
	// Create logger using layer context - it will be pre-configured
	logger := layerContext.NewHandlerLogger()
	errorHandler := error_system.NewHandlerErrorHandler(logger, "account")
	return &AccountHandler{
		userClient:      userClient,
		errorHandler: errorHandler,
		logger:          logger,
		layerContext:    layerContext,
		validator:       validator.New(),
	}
}

// Utility function for email masking
func maskEmail(email string) string {
	if len(email) < 3 {
		return "***"
	}
	at := -1
	for i, char := range email {
		if char == '@' {
			at = i
			break
		}
	}
	if at == -1 {
		return "***"
	}
	if at < 2 {
		return "***@" + email[at+1:]
	}
	return email[:1] + "***@" + email[at+1:]
}



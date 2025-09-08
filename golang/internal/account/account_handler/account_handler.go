// internal/account/account_handler/account_handler.go
package account_handler

import (
	"context"
	

	"net/http"
	"time"

	"github.com/go-playground/validator/v10"


	"english-ai-full/error_system"
	"english-ai-full/internal"
	"english-ai-full/internal/account/account_dto"
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

func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	const operation = "register"
	startTime := time.Now()
	
	// Extract request context information
	requestID := h.layerContext.GetRequestID(r)
	clientIP := h.layerContext.GetClientIP(r)
	
	// Build operation context using layer context
	operationCtx := h.layerContext.BuildOperationContext(operation, r, requestID)
	operationCtx["client_ip"] = clientIP
	
	// Set operation in logger
	h.logger.SetOperation(operation)
	
	// Log with layer context
	h.logger.Info("=== REGISTER FUNCTION STARTED ===", h.layerContext.MergeWithContext(map[string]interface{}{
		"test":       "logging_verification",
		"endpoint":   r.URL.Path,
		"method":     r.Method,
		"request_id": requestID,
		"client_ip":  clientIP,
	}))

	// Log request start with layer context
	h.logRequestStart(requestID, r.Method, r.URL.Path, "")

	// Context timeout check
	if err := r.Context().Err(); err != nil {
		h.handleError(w, err, operation, operationCtx, startTime)
		return
	}

	// Parse and validate request body
	var registerRequest account_dto.CreateUserRequest
	if err := h.decodeJSONRequest(r, &registerRequest); err != nil {
		h.handleError(w, err, operation, operationCtx, startTime)
		return
	}

	// Add parsed data to operation context (without sensitive info)
	operationCtx["email"] = maskEmail(registerRequest.Email)
	operationCtx["role"] = registerRequest.Role
	operationCtx["branch_id"] = registerRequest.BranchID

	// Additional request validation
	if err := h.validator.Struct(registerRequest); err != nil {
		h.logger.Error("Request validation failed", h.layerContext.MergeWithContext(map[string]interface{}{
			"error": err.Error(),
			"email": maskEmail(registerRequest.Email),
		}))
		
		// Convert validator errors to our format
		appErr := error_system.ValidationError("request", "Validation failed")
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			field := validationErrors[0].Field()
			reason := validationErrors[0].Tag()
			appErr = error_system.ValidationError(field, reason)
		}
		
		h.writeErrorResponse(w, appErr, requestID, startTime)
		return
	}
	
	// Create service context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Create protobuf request using AccountReq message
	pbRequest := &pb.AccountReq{
		BranchId: registerRequest.BranchID,
		Name:     registerRequest.Name,
		Email:    registerRequest.Email,
		Password: registerRequest.Password,
		Avatar:   registerRequest.Avatar,
		Title:    registerRequest.Title,
		Role:     registerRequest.Role,
		OwnerId:  registerRequest.OwnerID,
	}

	// Log service call attempt with layer context
	h.logger.Info("Calling CreateUser service", h.layerContext.MergeWithContext(map[string]interface{}{
		"service_method": "CreateUser",
		"target_service": "UserService",
		"email":          maskEmail(registerRequest.Email),
		"request_id":     requestID,
	}))

	// Call service layer using CreateUser RPC
	createdUser, err := h.userClient.CreateUser(ctx, pbRequest)
	if err != nil {
		duration := time.Since(startTime)
		operationCtx["duration_ms"] = duration.Milliseconds()
		
		h.logger.Error("Service call failed", h.layerContext.MergeWithContext(map[string]interface{}{
			"error":       err.Error(),
			"email":       maskEmail(registerRequest.Email),
			"request_id":  requestID,
			"duration_ms": duration.Milliseconds(),
		}))
		
		h.handleError(w, err, operation, operationCtx, startTime)
		return
	}

	// Log successful service call with layer context
	operationCtx["user_id"] = createdUser.Id
	operationCtx["created_at"] = createdUser.CreatedAt

	// Prepare response (exclude sensitive data)
	responseData := h.prepareUserResponse(createdUser)

	// Log successful request completion
	h.logRequestEnd(requestID, http.StatusCreated, startTime)

	// Send successful response
		h.writeSuccessResponse(w, responseData, http.StatusCreated, requestID)
	

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


// new asdfasdfasd

// internal/account/account_handler/account_handler.go - Updated handleError method


// ... (keep all existing code, only replacing handleError method)

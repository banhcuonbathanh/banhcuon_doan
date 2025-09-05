// internal/account/account_handler/account_handler.go
package account_handler

import (
	"context"
	
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"english-ai-full/internal"
	"english-ai-full/internal/account/account_dto"

	error_custom "english-ai-full/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	logg "english-ai-full/logger"
)


type AccountHandler struct {
	userClient      pb.AccountServiceClient
	handlerErrorMgr *error_custom.HandlerErrorManager
	logger          *logg.SpecializedLogger
	layerContext    *internal.LayerContext 
	validator       *validator.Validate
}

// NewAccountHandler creates a new account handler with dependencies
func NewAccountHandler(userClient pb.AccountServiceClient) *AccountHandler {
	return &AccountHandler{
		userClient:      userClient,
		handlerErrorMgr: error_custom.NewHandlerErrorManager(),
		logger:          logg.NewSpecializedHandlerLogger(),
			layerContext: &internal.LayerContext{
			Domain:      "account",
			Layer:       "handler",
			Service:     "account-service",
			Version:     "1.0.0",        // Optional: add version info
			Environment: "production",    // Optional: add environment info
		},
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
	
	// Log with layer context
	h.logger.Info("=== REGISTER FUNCTION STARTED ===", h.layerContext.MergeWithContext(map[string]interface{}{
		"test":     "logging_verification",
		"endpoint": r.URL.Path,
		"method":   r.Method,
	}))

	// Log request start with layer context
	h.logger.LogRequestStart(requestID, r.Method, r.URL.Path, "")

	// Context timeout check
	if err := r.Context().Err(); err != nil {
		h.logger.Error("Request context cancelled", h.layerContext.MergeWithContext(map[string]interface{}{
			"error":      err.Error(),
			"request_id": requestID,
		}))
		h.handlerErrorMgr.RespondWithError(w, err, h.layerContext.Domain, requestID)
		return
	}

	// Parse and validate request body
	var registerRequest account_dto.CreateUserRequest
	if err := h.handlerErrorMgr.DecodeJSONRequest(r, &registerRequest, h.layerContext.Domain, requestID); err != nil {
		operationCtx["decode_error"] = "failed to parse request body"
		h.logRequestEnd(requestID, http.StatusBadRequest, startTime)
		h.handlerErrorMgr.RespondWithError(w, err, h.layerContext.Domain, requestID)
		return
	}

	// Add parsed data to operation context (without sensitive info)
	operationCtx["email"] = registerRequest.Email
	operationCtx["role"] = registerRequest.Role
	operationCtx["branch_id"] = registerRequest.BranchID

	// Additional request validation
	if err := h.validator.Struct(registerRequest); err != nil {
		h.logger.LogStructValidationError("CreateUserRequest", registerRequest, "Request validation failed")
		
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			h.handlerErrorMgr.HandleValidationErrors(w, validationErrors, h.layerContext.Domain, requestID)
		} else {
			h.handlerErrorMgr.RespondWithError(w, err, h.layerContext.Domain, requestID)
		}
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
		"email":          registerRequest.Email, // This will be masked in logger
		"request_id":     requestID,
	}))

	// Call service layer using CreateUser RPC
	createdUser, err := h.userClient.CreateUser(ctx, pbRequest)
	if err != nil {
		// Determine appropriate HTTP status code based on error type
		statusCode := h.getHTTPStatusFromError(err)
		h.logRequestEnd(requestID, statusCode, startTime)
		h.handlerErrorMgr.RespondWithError(w, err, h.layerContext.Domain, requestID)
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
	h.handlerErrorMgr.RespondWithCreated(w, responseData, h.layerContext.Domain, requestID)
	
	// Final success log with layer context
	h.logger.Info("User registration completed successfully", h.layerContext.MergeWithContext(map[string]interface{}{
		"user_id":     createdUser.Id,
		"email":       createdUser.Email, // Will be masked
		"request_id":  requestID,
		"duration_ms": time.Since(startTime).Milliseconds(),
	}))
}




func (h *AccountHandler) getHTTPStatusFromError(err error) int {
	// Your error to HTTP status mapping logic
	// This is just an example
	errMsg := err.Error()
	switch {
	case contains(errMsg, "already exists"):
		return http.StatusConflict
	case contains(errMsg, "invalid"):
		return http.StatusBadRequest
	case contains(errMsg, "unauthorized"):
		return http.StatusUnauthorized
	case contains(errMsg, "forbidden"):
		return http.StatusForbidden
	case contains(errMsg, "not found"):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (h *AccountHandler) logRequestEnd(requestID string, statusCode int, startTime time.Time) {
	duration := time.Since(startTime)
	h.logger.LogRequestEnd(requestID, statusCode, duration, 0) // 0 for response size if not tracked
}



// Helper function
func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}


func (h *AccountHandler) prepareUserResponse(user *pb.Account) interface{} {
	// Your response preparation logic
	return map[string]interface{}{
		"id":       user.Id,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
		"created_at": user.CreatedAt,
		// Don't include sensitive fields like password
	}
}

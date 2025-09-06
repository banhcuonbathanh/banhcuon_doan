// internal/account/account_handler/account_handler.go
package account_handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"english-ai-full/internal"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"

	error_custom "english-ai-full/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
)

type AccountHandler struct {
	userClient      pb.AccountServiceClient
	handlerErrorMgr *error_custom.HandlerErrorManager
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

	return &AccountHandler{
		userClient:      userClient,
		handlerErrorMgr: error_custom.NewHandlerErrorManager(),
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
		h.logger.ErrorWithCause("Request context cancelled", err.Error(), core.LayerHandler, operation, 
			h.layerContext.MergeWithContext(map[string]interface{}{
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
	operationCtx["email"] = maskEmail(registerRequest.Email)
	operationCtx["role"] = registerRequest.Role
	operationCtx["branch_id"] = registerRequest.BranchID

	// Additional request validation
	if err := h.validator.Struct(registerRequest); err != nil {
		h.logStructValidationError("CreateUserRequest", registerRequest, "Request validation failed")
		
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
		"email":          maskEmail(registerRequest.Email),
		"request_id":     requestID,
	}))

	// Call service layer using CreateUser RPC
	createdUser, err := h.userClient.CreateUser(ctx, pbRequest)
	if err != nil {
		// Determine appropriate HTTP status code based on error type
		statusCode := h.getHTTPStatusFromError(err)
		h.logRequestEnd(requestID, statusCode, startTime)
		
		// Enhanced error logging
		h.logger.ErrorWithDomainAndCause("Service call failed", h.layerContext.Domain, err.Error(), core.LayerHandler, operation,
			h.layerContext.MergeWithContext(map[string]interface{}{
				"service_method": "CreateUser",
				"email":          maskEmail(registerRequest.Email),
				"request_id":     requestID,
				"status_code":    statusCode,
			}))
		
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
	h.logger.Info(core.MsgRegisterCompleted, h.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldUserID:    createdUser.Id,
		core.FieldEmail:     maskEmail(createdUser.Email),
		core.FieldRequestID: requestID,
		core.FieldDurationMS: time.Since(startTime).Milliseconds(),
	}))
}

// Helper methods
func (h *AccountHandler) logRequestStart(requestID, method, path, userAgent string) {
	h.logger.Info(core.MsgRequestStarted, h.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldRequestID:  requestID,
		core.FieldMethod:     method,
		core.FieldPath:       path,
		core.FieldUserAgent:  userAgent,
	}))
}

func (h *AccountHandler) logRequestEnd(requestID string, statusCode int, startTime time.Time) {
	duration := time.Since(startTime)
	
	logFunc := h.logger.Info
	if statusCode >= 400 {
		logFunc = h.logger.Error
	} else if statusCode >= 300 {
		logFunc = h.logger.Warn
	}
	
	logFunc(core.MsgRequestCompleted, h.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldRequestID:   requestID,
		core.FieldStatusCode:  statusCode,
		core.FieldDurationMS:  duration.Milliseconds(),
	}))
}

func (h *AccountHandler) logStructValidationError(structName string, data interface{}, message string) {
	h.logger.ErrorWithCause(core.MsgValidationError, core.CauseStructValidationFailed, core.LayerValidation, core.OperationValidateStruct,
		h.layerContext.MergeWithContext(map[string]interface{}{
			core.FieldStructName: structName,
			core.FieldMessage:    message,
		}))
}

func (h *AccountHandler) getHTTPStatusFromError(err error) int {
	// Simple implementation - you can enhance this based on your error types
	if err == nil {
		return http.StatusOK
	}
	
	// You can add more sophisticated error type checking here
	return http.StatusInternalServerError
}

func (h *AccountHandler) prepareUserResponse(user *pb.Account) map[string]interface{} {
	return map[string]interface{}{
		core.ResponseFieldID:        user.Id,
		core.ResponseFieldEmail:     user.Email,
		core.ResponseFieldName:      user.Name,
		core.ResponseFieldRole:      user.Role,
		core.ResponseFieldCreatedAt: user.CreatedAt,
		// Don't include password or other sensitive fields
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
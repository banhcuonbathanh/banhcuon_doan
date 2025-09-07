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

package account_handler

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"english-ai-full/error_system"
	"english-ai-full/internal"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger/core"

	pb "english-ai-full/internal/proto_qr/account"
)

// ... (keep all existing code, only replacing handleError method)

// FIXED: handleError processes errors and sends appropriate HTTP responses
func (h *AccountHandler) handleError(w http.ResponseWriter, err error, operation string, context map[string]interface{}, startTime time.Time) {
	duration := time.Since(startTime)
	context["duration_ms"] = duration.Milliseconds()

	// CRITICAL: Check if it's already our AppError - DON'T WRAP IT
	if appErr, ok := error_system.IsAppError(err); ok {
		h.logger.Info("Received AppError, passing through unchanged", h.layerContext.MergeWithContext(map[string]interface{}{
			"error_code":    appErr.Code,
			"error_message": appErr.Message,
			"operation":     operation,
			"layer":         "handler",
		}))
		h.writeErrorResponse(w, appErr, context["request_id"].(string), startTime)
		return
	}

	// NEW: Check if it's a gRPC error that contains our AppError
	if grpcAppErr := h.extractAppErrorFromGRPC(err); grpcAppErr != nil {
		h.logger.Info("Extracted AppError from gRPC error", h.layerContext.MergeWithContext(map[string]interface{}{
			"error_code":        grpcAppErr.Code,
			"error_message":     grpcAppErr.Message,
			"original_grpc_error": err.Error(),
			"operation":         operation,
			"layer":             "handler",
		}))
		h.writeErrorResponse(w, grpcAppErr, context["request_id"].(string), startTime)
		return
	}

	// Only convert to AppError if it's not already one and not a recoverable gRPC error
	h.logger.Info("Converting raw error to AppError", h.layerContext.MergeWithContext(map[string]interface{}{
		"raw_error": err.Error(),
		"operation": operation,
		"layer":     "handler",
	}))
	
	appErr := h.errorHandler.Handle(err, operation, context)
	h.writeErrorResponse(w, appErr, context["request_id"].(string), startTime)
}

// NEW: Extract AppError information from gRPC errors
func (h *AccountHandler) extractAppErrorFromGRPC(err error) *error_system.AppError {
	// Check if it's a gRPC status error
	if grpcStatus, ok := status.FromError(err); ok {
		return h.parseGRPCStatusToAppError(grpcStatus)
	}
	
	// Check if it's a gRPC error string format
	if strings.Contains(err.Error(), "rpc error:") {
		return h.parseGRPCStringToAppError(err.Error())
	}
	
	return nil
}

// Parse gRPC status to AppError
func (h *AccountHandler) parseGRPCStatusToAppError(grpcStatus *status.Status) *error_system.AppError {
	message := grpcStatus.Message()
	
	// Look for our AppError pattern in the message: [ERROR_CODE] Message
	return h.extractAppErrorFromMessage(message)
}

// Parse gRPC error string to AppError  
func (h *AccountHandler) parseGRPCStringToAppError(errStr string) *error_system.AppError {
	// Pattern: "rpc error: code = Unknown desc = [ERROR_CODE] Message"
	// Extract the description part after "desc = "
	descIndex := strings.Index(errStr, "desc = ")
	if descIndex == -1 {
		return nil
	}
	
	description := errStr[descIndex+7:] // Skip "desc = "
	return h.extractAppErrorFromMessage(description)
}

// Extract AppError from message with pattern [ERROR_CODE] Message
func (h *AccountHandler) extractAppErrorFromMessage(message string) *error_system.AppError {
	// Regex to match [ERROR_CODE] Message pattern
	re := regexp.MustCompile(`^\[([A-Z_]+)\]\s*(.+)$`)
	matches := re.FindStringSubmatch(message)
	
	if len(matches) != 3 {
		h.logger.Debug("Could not extract AppError from message", h.layerContext.MergeWithContext(map[string]interface{}{
			"message": message,
			"matches_count": len(matches),
		}))
		return nil
	}
	
	errorCode := error_system.ErrorCode(matches[1])
	errorMessage := strings.TrimSpace(matches[2])
	
	h.logger.Debug("Successfully extracted AppError from message", h.layerContext.MergeWithContext(map[string]interface{}{
		"extracted_code": errorCode,
		"extracted_message": errorMessage,
		"original_message": message,
	}))
	
	// Create AppError with extracted information
	appErr := error_system.NewError(errorCode, errorMessage)
	
	// Add any additional context based on error code
	switch errorCode {
	case error_system.ErrAccountDuplicate:
		// Extract email from context if available
		if email, exists := h.getEmailFromContext(); exists {
			appErr.Details["email"] = error_system.MaskEmail(email)
		}
	}
	
	return appErr
}

// Helper to get email from current request context (if needed)
func (h *AccountHandler) getEmailFromContext() (string, bool) {
	// This would depend on how you store request context
	// For now, return empty - you can enhance this if needed
	return "", false
}

// Updated writeErrorResponse with better logging
func (h *AccountHandler) writeErrorResponse(w http.ResponseWriter, appErr *error_system.AppError, requestID string, startTime time.Time) {
	duration := time.Since(startTime)
	
	// Log the error with the actual error code being returned
	h.logger.Error("Request completed with error", h.layerContext.MergeWithContext(map[string]interface{}{
		"error_code":          appErr.Code,
		"error_message":       appErr.Message,
		"status_code":         appErr.HTTPStatus,
		"request_id":          requestID,
		"duration_ms":         duration.Milliseconds(),
		"final_response_code": string(appErr.Code), // Explicitly log what's being returned
		"response_will_contain": map[string]interface{}{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	}))

	// Write the HTTP response
	appErr.WriteHTTPResponse(w)
}
// new asdfasdfsd
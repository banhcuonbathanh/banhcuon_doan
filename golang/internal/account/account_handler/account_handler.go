// internal/account/account_handler/account_handler.go
package account_handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"english-ai-full/internal/account/account_dto"

	error_custom "english-ai-full/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	logg "english-ai-full/logger"

)

// AccountHandlerInterface defines the contract for account handlers

// AccountHandler implements the handler layer for account operations
type AccountHandler struct {
	userClient      pb.AccountServiceClient
	handlerErrorMgr *error_custom.HandlerErrorManager
	logger          *logg.SpecializedLogger
	domain          string
	validator       *validator.Validate
}

// NewAccountHandler creates a new account handler with dependencies
func NewAccountHandler(userClient pb.AccountServiceClient) *AccountHandler {
	return &AccountHandler{
		userClient:      userClient,
		handlerErrorMgr: error_custom.NewHandlerErrorManager(),
		// logger:          logg.NewSpecializedHandlerLogger(),
		domain:          "account",
		validator:       validator.New(),
			logger:          logg.NewSmartLogger("account_handler"), 
	}
}

// Enhanced Register function with auto-configured logging
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Create request-scoped logger with automatic context extraction
	requestLogger := h.logger.NewRequestLogger(r)
	
	// Log request start - automatically includes method, endpoint, IP, etc.
	requestLogger.LogRequestStart()
	
	// Context timeout check with automatic error logging
	if err := r.Context().Err(); err != nil {
		requestLogger.Error("Request context cancelled", map[string]interface{}{
			"error": err.Error(),
		})
		h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestLogger.GetRequestID(r))
		return
	}

	// Parse and validate request body
	var registerRequest account_dto.CreateUserRequest
	if err := h.handlerErrorMgr.DecodeJSONRequest(r, &registerRequest, h.domain, requestLogger.GetRequestID(r)); err != nil {
		// Auto-configured validation error logging
		requestLogger.LogValidationError(err)
		requestLogger.LogRequestEnd(http.StatusBadRequest)
		h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestLogger.GetRequestID(r))
		return
	}

	// Add parsed data to context for automatic inclusion in logs
	requestLogger.context.WithFields(map[string]interface{}{
		"email":     registerRequest.Email,
		"role":      registerRequest.Role,
		"branch_id": registerRequest.BranchID,
	})

	// Validation with auto-configured error logging
	if err := h.validator.Struct(registerRequest); err != nil {
		requestLogger.LogValidationError(err)
		
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			h.handlerErrorMgr.HandleValidationErrors(w, validationErrors, h.domain, requestLogger.GetRequestID(r))
		} else {
			h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestLogger.GetRequestID(r))
		}
		requestLogger.LogRequestEnd(http.StatusBadRequest)
		return
	}
	
	// Create service context
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Create protobuf request
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

	// Log service call with automatic timing and error handling
	requestLogger.LogServiceCallStart("UserService", "CreateUser")
	serviceCallStart := time.Now()
	
	createdUser, err := h.userClient.CreateUser(ctx, pbRequest)
	serviceCallDuration := time.Since(serviceCallStart)
	
	// Automatic service call logging with error details
	requestLogger.LogServiceCallEnd("UserService", "CreateUser", err, serviceCallDuration)
	
	if err != nil {
		statusCode := h.getHTTPStatusFromError(err)
		requestLogger.LogRequestEnd(statusCode)
		h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestLogger.GetRequestID(r))
		return
	}

	// Log business event - automatically includes user context
	requestLogger.LogBusinessEvent("user_created", createdUser.Id, "user", "create", map[string]interface{}{
		"email": createdUser.Email,
		"role":  createdUser.Role,
	})

	// Prepare response
	responseData := h.prepareUserResponse(createdUser)

	// Send successful response
	h.handlerErrorMgr.RespondWithCreated(w, responseData, h.domain, requestLogger.GetRequestID(r))
	
	// Log successful completion - automatically includes duration and status
	requestLogger.LogRequestEnd(http.StatusCreated)
	
	// Additional success logging with automatic masking and context
	requestLogger.Info("User registration completed successfully", map[string]interface{}{
		"user_id": createdUser.Id,
		"email":   createdUser.Email, // Auto-masked by logger
	})
}

// Helper methods (you might already have these)
func (h *AccountHandler) getRequestID(r *http.Request) string {
	// Implementation to extract request ID from headers or generate one
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return requestID
}

func (h *AccountHandler) getClientIP(r *http.Request) string {
	// Get client IP from various headers
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
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

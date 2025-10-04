package account_handler

import (
	"context"
	"english-ai-full/error_system"
	"english-ai-full/internal/account/account_dto"
	pb "english-ai-full/internal/proto_qr/account"
	"net/http"
	"time"

	"github.com/go-playground/validator"
)


func (h *AccountHandler) Login(w http.ResponseWriter, r *http.Request) {
	const operation = "login"
	startTime := time.Now()
	
	// Extract request context information
	requestID := h.getRequestIDFromContext(r.Context())
	clientIP := h.layerContext.GetClientIP(r)
	
	// Build operation context using layer context
	operationCtx := h.layerContext.BuildOperationContext(operation, r, requestID)
	operationCtx["client_ip"] = clientIP
	
	// Set operation in logger
	h.logger.SetOperation(operation)
	
	// Log request start with layer context
	h.logRequestStart(requestID, r.Method, r.URL.Path, "")

	// Context timeout check
	if err := r.Context().Err(); err != nil {
		h.handleError(w, err, operation, operationCtx, startTime)
		return
	}

	// Parse and validate request body
	var loginRequest account_dto.LoginRequest
	if err := h.decodeJSONRequest(r, &loginRequest); err != nil {
		h.handleError(w, err, operation, operationCtx, startTime)
		return
	}

	// Add parsed data to operation context (without sensitive info)
	operationCtx["email"] = maskEmail(loginRequest.Email)

	// Validate request using validator
	if err := h.validator.Struct(loginRequest); err != nil {
		h.logger.Error("Request validation failed", h.layerContext.MergeWithContext(map[string]interface{}{
			"error": err.Error(),
			"email": maskEmail(loginRequest.Email),
		}))
		
		// Convert validator errors to our format
		appErr := h.handleValidationError(err)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			field := validationErrors[0].Field()
			reason := validationErrors[0].Tag()
			appErr = error_system.ValidationError(field, reason)
		}
		
		h.writeErrorResponse(w, appErr, requestID, startTime)
		return
	}

	// Additional business validation (optional)
	if loginRequest.Email == "" {
		h.logger.Error("Empty email provided", h.layerContext.MergeWithContext(map[string]interface{}{
			"request_id": requestID,
		}))
		
		appErr := error_system.ValidationError("email", "required")
		appErr.Message = "Email is required"
		appErr.MessageVN = "Email là bắt buộc"
		
		h.writeErrorResponse(w, appErr, requestID, startTime)
		return
	}

	if loginRequest.Password == "" {
		h.logger.Error("Empty password provided", h.layerContext.MergeWithContext(map[string]interface{}{
			"email":      maskEmail(loginRequest.Email),
			"request_id": requestID,
		}))
		
		appErr := error_system.ValidationError("password", "required")
		appErr.Message = "Password is required"
		appErr.MessageVN = "Mật khẩu là bắt buộc"
		
		h.writeErrorResponse(w, appErr, requestID, startTime)
		return
	}

	// Create service context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Create protobuf request using LoginReq message
	pbRequest := &pb.LoginReq{
		Email:    loginRequest.Email,
		Password: loginRequest.Password,
	}

	// Call service layer using Login RPC
	loginResponse, err := h.userClient.Login(ctx, pbRequest)
	if err != nil {
		duration := time.Since(startTime)
		operationCtx["duration_ms"] = duration.Milliseconds()
		
		h.logger.Error("Service call failed", h.layerContext.MergeWithContext(map[string]interface{}{
			"error":       err.Error(),
			"email":       maskEmail(loginRequest.Email),
			"request_id":  requestID,
			"duration_ms": duration.Milliseconds(),
		}))
		
		h.handleError(w, err, operation, operationCtx, startTime)
		return
	}

	// Validate response from service
	if loginResponse == nil {
		h.logger.Error("Received nil response from service", h.layerContext.MergeWithContext(map[string]interface{}{
			"email":      maskEmail(loginRequest.Email),
			"request_id": requestID,
		}))
		
		appErr := error_system.NewError(error_system.ErrSystemError, "Login service returned invalid response")
		h.writeErrorResponse(w, appErr, requestID, startTime)
		return
	}

	if loginResponse.Account == nil {
		h.logger.Error("Received response with nil account", h.layerContext.MergeWithContext(map[string]interface{}{
			"email":      maskEmail(loginRequest.Email),
			"request_id": requestID,
		}))
		
		appErr := error_system.NewError(error_system.ErrSystemError, "Login service returned invalid account data")
		h.writeErrorResponse(w, appErr, requestID, startTime)
		return
	}

	// Log successful service call with layer context
	operationCtx["user_id"] = loginResponse.Account.Id
	operationCtx["login_time"] = time.Now().Unix()
	operationCtx["user_role"] = loginResponse.Account.Role

	h.logger.Info("Login successful", h.layerContext.MergeWithContext(map[string]interface{}{
		"user_id":     loginResponse.Account.Id,
		"email":       maskEmail(loginResponse.Account.Email),
		"role":        loginResponse.Account.Role,
		"branch_id":   loginResponse.Account.BranchId,
		"request_id":  requestID,
		"duration_ms": time.Since(startTime).Milliseconds(),
	}))

	// Prepare response data
	responseData := account_dto.LoginUserRes{
		AccessToken:  loginResponse.AccessToken,
		RefreshToken: loginResponse.RefreshToken,
		User: account_dto.AccountLoginResponse{
			ID:       loginResponse.Account.Id,
			BranchID: loginResponse.Account.BranchId,
			Name:     loginResponse.Account.Name,
			Email:    loginResponse.Account.Email,
			Avatar:   loginResponse.Account.Avatar,
			Title:    loginResponse.Account.Title,
			Role:     loginResponse.Account.Role,
			OwnerID:  loginResponse.Account.OwnerId,
		},
	}

	// Log successful request completion
	h.logRequestEnd(requestID, http.StatusOK, startTime)

	// Send successful response
	h.writeSuccessResponse(w, responseData, http.StatusOK, requestID)
}
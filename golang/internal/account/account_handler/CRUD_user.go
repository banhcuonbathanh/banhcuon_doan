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


func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	const operation = "register"
	startTime := time.Now()
	
	// Extract request context information

	    requestID := h.getRequestIDFromContext(r.Context())
	// requestID := h.layerContext.GetRequestID(r)
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
		appErr := h.handleValidationError(err)
		// appErr := error_system.ValidationError("request", "Validation failed")
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			field := validationErrors[0].Field()
			reason := validationErrors[0].Tag()
			appErr = error_system.ValidationError(field, reason)
		}
		
		h.writeErrorResponse(w, appErr, requestID, startTime)
		return
	}
if err := registerRequest.ValidatePasswordMatch(); err != nil {
	h.logger.Error("Password confirmation validation failed", h.layerContext.MergeWithContext(map[string]interface{}{
		"error": err.Error(),
		"email": maskEmail(registerRequest.Email),
	}))
	
	appErr := error_system.EnhancedValidationError(
		"confirm_password", 
		"***", // masked password
		"match", 
		"password",
	)
	appErr.Message = "Password and confirm password do not match"
	appErr.MessageVN = "Mật khẩu và xác nhận mật khẩu không khớp"
	
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


// login start
// Login handles user authentication
// Login handles user authentication
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
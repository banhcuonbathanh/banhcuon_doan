package account_handler

import (
	"fmt"
	"net/http"
	"strings"

	"english-ai-full/internal/account/account_dto"
	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"

	"github.com/go-playground/validator/v10"
)

// Register handles user registration HTTP requests
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Setup operation context
	opCtx := h.setupOperationContext(r, "register_user")
	h.logOperationStart(opCtx)
	h.SetLoggerOperation("register_user")
	
	var statusCode int
	var err error
	
	defer func() {
		h.logOperationEnd(opCtx, err, statusCode)
	}()
	
	// Decode JSON request body
	var req account_dto.RegisterUserRequest
	if err = h.errorHandler.DecodeJSONRequest(r, &req); err != nil {
		statusCode = http.StatusBadRequest
		h.logger.ErrorWithCause(
			"Failed to decode registration request",
			"invalid_json",
			logger.LayerHandler,
			"register_user",
			map[string]interface{}{
				"request_id": opCtx.RequestID,
				"error":      err.Error(),
			},
		)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Validate the request using the validator
	if err = h.validator.Struct(&req); err != nil {
		statusCode = http.StatusBadRequest
		
		// Handle validation errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			h.logger.LogValidationError("registration_request", "Validation failed", req)
			
			// Create detailed validation error response
			validationErr := h.buildValidationError(validationErrors)
			h.errorHandler.HandleHTTPError(w, r, validationErr)
			return
		}
		
		// Generic validation error
		validationErr := errorcustom.NewValidationError(h.domain, "request", "Invalid registration request", req)
		h.errorHandler.HandleHTTPError(w, r, validationErr)
		return
	}
	
	// Log registration attempt (without sensitive data)
	h.logger.LogUserActivity("0", h.maskEmail(req.Email), "register_attempt", "user_account", map[string]interface{}{
		"request_id": opCtx.RequestID,
		"name":       req.Name,
		"operation":  "register_user",
	})
	
	// Enrich context for gRPC call
	ctx := h.enrichContext(r.Context(), "register_user", map[string]interface{}{
		"email": req.Email,
		"name":  req.Name,
	})
	
	// Convert DTO to gRPC request
	grpcReq := &pb.RegisterReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	
	// Call the gRPC service
	registerRes, err := h.userClient.Register(ctx, grpcReq)
	if err != nil {
		statusCode = h.determineErrorStatusCode(err)
		
		// Log the error with appropriate level based on error type
		h.logRegistrationError(err, req.Email, opCtx.RequestID)
		
		// Convert gRPC error to domain error
		domainErr := h.errorHandler.ParseGRPCError(err, h.domain, "register_user", map[string]interface{}{
			"email":      h.maskEmail(req.Email),
			"request_id": opCtx.RequestID,
		})
		
		h.errorHandler.HandleHTTPError(w, r, domainErr)
		return
	}
	
	// Transform gRPC response to DTO response
	response := &account_dto.RegisterUserResponse{
		ID:      registerRes.GetId(),
		Name:    registerRes.GetName(),
		Email:   registerRes.GetEmail(),
		Success: true,
		Message: "User registered successfully",
	}
	
	// Log successful registration
	h.logger.LogUserActivity(
		fmt.Sprintf("%d", response.ID),
		response.Email,
		"register_success",
		"user_account",
		map[string]interface{}{
			"request_id": opCtx.RequestID,
			"user_id":    response.ID,
			"name":       response.Name,
			"operation":  "register_user",
		},
	)
	
	// Set success status code
	statusCode = http.StatusCreated
	
	// Send success response
	h.errorHandler.RespondWithSuccess(w, r, response)
}

// buildValidationError creates a structured validation error from validator errors
func (h *AccountHandler) buildValidationError(validationErrors validator.ValidationErrors) error {
	errorDetails := make(map[string]interface{})
	
	for _, err := range validationErrors {
		field := err.Field()
		tag := err.Tag()
		value := err.Value()
		
		var message string
		switch tag {
		case "required":
			message = fmt.Sprintf("%s is required", field)
		case "email":
			message = "Invalid email format"
		case "min":
			message = fmt.Sprintf("%s must be at least %s characters", field, err.Param())
		case "max":
			message = fmt.Sprintf("%s must be at most %s characters", field, err.Param())
		case "strongpassword":
			message = "Password must contain at least 8 characters with uppercase, lowercase, number and special character"
		case "userrole":
			message = "Invalid user role. Must be one of: admin, teacher, student"
		default:
			message = fmt.Sprintf("%s is invalid", field)
		}
		
		errorDetails[field] = message
		
		// Log individual validation error
		h.logger.LogValidationError(field, message, value)
	}
	
	return errorcustom.NewValidationErrorWithContext(
		h.domain,
		"request",
		"Registration validation failed",
		"",
		errorDetails,
	)
}

// logRegistrationError logs registration errors with appropriate context
func (h *AccountHandler) logRegistrationError(err error, email, requestID string) {
	maskedEmail := h.maskEmail(email)
	
	switch {
	case errorcustom.IsValidationError(err):
		h.logger.LogValidationError("registration", err.Error(), maskedEmail)
	case errorcustom.IsDuplicateError(err):
		h.logger.Warning("Duplicate registration attempt", map[string]interface{}{
			"email":      maskedEmail,
			"request_id": requestID,
			"error":      err.Error(),
			"operation":  "register_user",
		})
		
		// Log security event for potential abuse
		logger.LogSecurityEvent(
			"duplicate_registration",
			"Attempt to register with existing email",
			"low",
			map[string]interface{}{
				"email":      maskedEmail,
				"request_id": requestID,
			},
		)
	case errorcustom.IsBusinessLogicError(err):
		h.logger.ErrorWithCause(
			"Business rule violation during registration",
			"business_logic_error",
			logger.LayerHandler,
			"register_user",
			map[string]interface{}{
				"email":      maskedEmail,
				"request_id": requestID,
				"error":      err.Error(),
			},
		)
	case errorcustom.IsExternalServiceError(err):
		h.logger.ErrorWithCause(
			"External service error during registration",
			"external_service_error",
			logger.LayerHandler,
			"register_user",
			map[string]interface{}{
				"email":      maskedEmail,
				"request_id": requestID,
				"error":      err.Error(),
				"retryable":  errorcustom.IsRetryableError(err),
			},
		)
	default:
		h.logger.ErrorWithCause(
			"Registration failed",
			"service_error",
			logger.LayerHandler,
			"register_user",
			map[string]interface{}{
				"email":      maskedEmail,
				"request_id": requestID,
				"error":      err.Error(),
			},
		)
	}
}

// determineErrorStatusCode determines the appropriate HTTP status code for errors
func (h *AccountHandler) determineErrorStatusCode(err error) int {
	switch {
	case errorcustom.IsValidationError(err):
		return http.StatusBadRequest
	case errorcustom.IsDuplicateError(err):
		return http.StatusConflict
	case errorcustom.IsAuthenticationError(err):
		return http.StatusUnauthorized
	case errorcustom.IsAuthorizationError(err):
		return http.StatusForbidden
	case errorcustom.IsNotFoundError(err):
		return http.StatusNotFound
	case errorcustom.IsBusinessLogicError(err):
		return http.StatusUnprocessableEntity
	case errorcustom.IsExternalServiceError(err):
		if errorcustom.IsRetryableError(err) {
			return http.StatusServiceUnavailable
		}
		return http.StatusBadGateway
	case errorcustom.IsRateLimitError(err):
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// maskEmail masks email address for secure logging
func (h *AccountHandler) maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	
	username := parts[0]
	domain := parts[1]
	
	var maskedUsername string
	if len(username) <= 2 {
		maskedUsername = strings.Repeat("*", len(username))
	} else {
		maskedUsername = string(username[0]) + strings.Repeat("*", len(username)-2) + string(username[len(username)-1])
	}
	
	return maskedUsername + "@" + domain
}
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
	"english-ai-full/utils"
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
		logger:          logg.NewSpecializedHandlerLogger(),
		domain:          "account",
		validator:       validator.New(),
	}
}

// Register handles user registration requests new 1212121212121212
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	const operation = "register"
	startTime := time.Now()
	
	// Extract request context information
	requestID := h.getRequestID(r)
	clientIP := h.getClientIP(r)
	
	// Build operation context for logging
	operationCtx := map[string]interface{}{
		"operation":  operation,
		"layer":      "handler",
		"domain":     h.domain,
		"request_id": requestID,
		"client_ip":  clientIP,
		"method":     r.Method,
		"path":       r.URL.Path,
	}

	// Log request start
	h.logger.LogRequestStart(requestID, r.Method, r.URL.Path, "")

	// Context timeout check
	if err := r.Context().Err(); err != nil {
		h.logger.Error("Request context cancelled", map[string]interface{}{
			"error":      err.Error(),
			"request_id": requestID,
		})
		h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestID)
		return
	}

	// Parse and validate request body
	var registerRequest account_dto.CreateUserRequest
	if err := h.handlerErrorMgr.DecodeJSONRequest(r, &registerRequest, h.domain, requestID); err != nil {
		operationCtx["decode_error"] = "failed to parse request body"
		h.logRequestEnd(requestID, http.StatusBadRequest, startTime)
		h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestID)
		return
	}

	// Add parsed data to context (without sensitive info)
	operationCtx["email"] = registerRequest.Email
	operationCtx["role"] = registerRequest.Role
	operationCtx["branch_id"] = registerRequest.BranchID

	// Additional request validation
if err := h.validator.Struct(registerRequest); err != nil {
    h.logger.LogStructValidationError("CreateUserRequest", registerRequest, "Request validation failed")
    
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        // Use the HandleValidationErrors function from HandlerErrorManager
        h.handlerErrorMgr.HandleValidationErrors(w, validationErrors, h.domain, requestID)
    } else {
        // For non-validation errors, use RespondWithError with the error directly
        // Let RespondWithError handle the conversion to APIError
        h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestID)
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

	// Call service layer using CreateUser RPC
	createdUser, err := h.userClient.CreateUser(ctx, pbRequest)
	if err != nil {
		// Log service call failure
		h.logger.LogServiceCall(h.domain, "CreateUser", false, err, operationCtx)
		
		// Determine appropriate HTTP status code based on error type
		statusCode := h.getHTTPStatusFromError(err)
		h.logRequestEnd(requestID, statusCode, startTime)
		h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestID)
		return
	}

	// Log successful service call
	operationCtx["user_id"] = createdUser.Id
	operationCtx["created_at"] = createdUser.CreatedAt
	h.logger.LogServiceCall(h.domain, "CreateUser", true, nil, operationCtx)

	// Prepare response (exclude sensitive data)
	responseData := h.prepareUserResponse(createdUser)

	// Log successful request completion
	h.logRequestEnd(requestID, http.StatusCreated, startTime)

	// Log user activity
	h.logger.LogUserActivity(
		fmt.Sprintf("%d", createdUser.Id), 
		"register",
		"user_account",
		map[string]interface{}{
			"email":      createdUser.Email,
			"request_id": requestID,
			"client_ip":  clientIP,
		},
	)

	// Send successful response
	h.handlerErrorMgr.RespondWithCreated(w, responseData, h.domain, requestID)
}


// prepareUserResponse prepares the user response excluding sensitive data
func (h *AccountHandler) prepareUserResponse(user *pb.Account) map[string]interface{} {
	return map[string]interface{}{
		"id":         user.Id,
		"branch_id":  user.BranchId,
		"name":       user.Name,
		"email":      user.Email,
		"avatar":     user.Avatar,
		"title":      user.Title,
		"role":       user.Role,
		"owner_id":   user.OwnerId,
		"status":     user.Status.String(),
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
		// Note: password is intentionally excluded from response
	}
}

// ===== VALIDATION METHODS ===== new 12121212121212


// validateRequiredFields checks if required fields are present and not empty
func (h *AccountHandler) validateRequiredFields(req account_dto.Account, fields []string) error {
	errorCollection := error_custom.NewErrorCollection(h.domain)

	for _, field := range fields {
		switch field {
		case "email":
			if req.Email == "" {
				errorCollection.Add(error_custom.NewValidationError(h.domain, "email", "Email is required", nil))
			}
		case "name":
			if req.Name == "" {
				errorCollection.Add(error_custom.NewValidationError(h.domain, "name", "Name is required", nil))
			}
		case "password":
			if req.Password == "" {
				errorCollection.Add(error_custom.NewValidationError(h.domain, "password", "Password is required", nil))
			}
		case "role":
			if req.Role == "" {
				errorCollection.Add(error_custom.NewValidationError(h.domain, "role", "Role is required", nil))
			}
		}
	}

	if errorCollection.HasErrors() {
		return errorCollection.ToAPIError()
	}

	return nil
}

// validateEmail validates email format and constraints
func (h *AccountHandler) validateEmail(email string) error {
	if !utils.IsValidEmail(email) {
		return error_custom.NewValidationError(h.domain, "email", "Invalid email format", email)
	}

	if len(email) > 254 {
		return error_custom.NewValidationError(h.domain, "email", "Email address is too long (maximum 254 characters)", email)
	}

	return nil
}

// validatePassword validates password strength and format
func (h *AccountHandler) validatePassword(password string) error {
	if len(password) < 8 {
		return error_custom.NewValidationError(h.domain, "password", "Password must be at least 8 characters long", nil)
	}

	if len(password) > 100 {
		return error_custom.NewValidationError(h.domain, "password", "Password must be no more than 100 characters long", nil)
	}

	// Check password complexity
	if !h.isPasswordStrong(password) {
		return error_custom.NewValidationError(
			h.domain, 
			"password", 
			"Password must contain uppercase, lowercase, number, and special character", 
			nil,
		)
	}

	return nil
}

// validateName validates user name format and constraints
func (h *AccountHandler) validateName(name string) error {
	if len(name) < 2 {
		return error_custom.NewValidationError(h.domain, "name", "Name must be at least 2 characters long", name)
	}

	if len(name) > 100 {
		return error_custom.NewValidationError(h.domain, "name", "Name must be no more than 100 characters long", name)
	}

	return nil
}

// validateRole validates user role
func (h *AccountHandler) validateRole(role string) error {
	validRoles := []string{"user", "admin", "moderator"}
	
	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}

	return error_custom.NewValidationError(
		h.domain, 
		"role", 
		"Invalid role. Must be one of: user, admin, moderator", 
		role,
	)
}

// ===== UTILITY METHODS =====



// isPasswordStrong checks password complexity requirements
func (h *AccountHandler) isPasswordStrong(password string) bool {
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case char >= 32 && char <= 126: // printable ASCII range
			// Check if it's a special character (not letter or number)
			if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// getHTTPStatusFromError determines appropriate HTTP status code from error
func (h *AccountHandler) getHTTPStatusFromError(err error) int {
	if apiErr, ok := err.(*error_custom.APIError); ok {
		return apiErr.HTTPStatus
	}

	// Fallback based on error type
	switch err.(type) {
	case *error_custom.ValidationError:
		return http.StatusBadRequest
	case *error_custom.DuplicateError:
		return http.StatusConflict
	case *error_custom.NotFoundError:
		return http.StatusNotFound
	case *error_custom.AuthenticationError:
		return http.StatusUnauthorized
	case *error_custom.AuthorizationError:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// getRequestID extracts request ID from context or headers
func (h *AccountHandler) getRequestID(r *http.Request) string {
	// Try context first
	if requestID, ok := r.Context().Value("request_id").(string); ok {
		return requestID
	}

	// Try header
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate new one if not found
	return utils.GenerateUUID()
}

// getClientIP extracts client IP address from request
func (h *AccountHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	return r.RemoteAddr
}

// logRequestEnd logs the completion of a request
func (h *AccountHandler) logRequestEnd(requestID string, statusCode int, startTime time.Time) {
	duration := time.Since(startTime)
	h.logger.LogRequestEnd(requestID, statusCode, duration, 0) // response size can be added if needed
}

// ===== ROUTER SETUP HELPER =====


func getRequestID(r *http.Request) string {
    // Option 1: From header (if set by middleware)
    if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
        return requestID
    }
    
    // Option 2: From context (if set by middleware)
    if requestID := r.Context().Value("request_id"); requestID != nil {
        if id, ok := requestID.(string); ok {
            return id
        }
    }
    
    // Option 3: Generate new UUID (you'll need to import a UUID package)
    // return uuid.New().String()
    
    // Fallback: generate simple ID
    return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
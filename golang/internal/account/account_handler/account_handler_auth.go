// internal/account/account_handler/account_handler_auth.go
package account_handler

import (
	"net/http"
	"time"
utils_config "english-ai-full/utils/config"
	"english-ai-full/internal/account/account_dto"
	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/mapping"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)

// Login handles user authentication with comprehensive error tracking
func (h *AccountHandler) Login(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Extract IP and User-Agent for security logging
	clientIP := utils.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	logger.Debug("Login attempt initiated", map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	var req account_dto.LoginRequest
	if err := utils.DecodeJSON(r.Body, &req, "login", false); err != nil {
		logger.Error("Failed to decode login request", map[string]interface{}{
			"error":      err.Error(),
			"ip":         clientIP,
			"user_agent": userAgent,
		})
		utils.HandleError(w, err, "login")
		return
	}

	// Log the email being attempted (for security monitoring)
	logger.Info("Login attempt for email", map[string]interface{}{
		"email":      req.Email,
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	// Validate request structure
	if err := h.validator.Struct(&req); err != nil {
		logger.LogValidationError("login_request", "Request validation failed", req)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors, "login")
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "login")
		}
		return
	}

	// Call authentication service
	logger.Debug("Calling authentication service", map[string]interface{}{
		"email":   req.Email,
		"service": "AccountService",
		"method":  "Login",
	})

	userRes, err := h.userClient.Login(r.Context(), &pb.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		// Parse the gRPC error with enhanced context
		parsedErr := errorcustom.ParseGRPCError(err, "login", req.Email)
		
		// Determine specific failure reason for logging and client response
		var failureReason string
		var userExists bool
		var clientError *errorcustom.APIError
		
		if errorcustom.IsUserNotFoundError(parsedErr) {
			failureReason = "email_not_found"
			userExists = false
			
			// Create detailed error for client
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeUserNotFound,
				"User with this email address was not found",
				http.StatusNotFound,
				"handler",
				"login",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "email_verification").
			  WithDetail("user_found", false)
			
			logger.Warning("Login failed - email not found", map[string]interface{}{
				"email":         req.Email,
				"ip":            clientIP,
				"user_agent":    userAgent,
				"failure_reason": failureReason,
				"layer":         "handler",
				"operation":     "login",
			})
			
		} else if errorcustom.IsPasswordError(parsedErr) {
			failureReason = "password_mismatch"
			userExists = true
			
			// Create detailed error for client
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeAuthFailed,
				"The password you entered is incorrect",
				http.StatusUnauthorized,
				"handler",
				"login",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "password_verification").
			  WithDetail("user_found", true)
			
			logger.Warning("Login failed - password mismatch", map[string]interface{}{
				"email":         req.Email,
				"ip":            clientIP,
				"user_agent":    userAgent,
				"failure_reason": failureReason,
				"layer":         "handler",
				"operation":     "login",
			})
			
		} else if apiErr, ok := parsedErr.(*errorcustom.APIError); ok {
			if apiErr.Code == errorcustom.ErrCodeAccessDenied {
				failureReason = "account_disabled_or_locked"
				userExists = true
				
				// Create detailed error for client
				clientError = errorcustom.NewAPIErrorWithContext(
					errorcustom.ErrCodeAccessDenied,
					"Your account has been disabled or locked. Please contact support.",
					http.StatusForbidden,
					"handler",
					"login",
					err,
				).WithDetail("email", req.Email).
				  WithDetail("step", "account_status_check").
				  WithDetail("user_found", true)
			} else {
				failureReason = "service_error"
				userExists = false
				
				// Create detailed error for client
				clientError = errorcustom.NewAPIErrorWithContext(
					errorcustom.ErrCodeServiceError,
					"Authentication service is temporarily unavailable. Please try again later.",
					http.StatusServiceUnavailable,
					"handler",
					"login",
					err,
				).WithDetail("email", req.Email).
				  WithDetail("step", "service_call").
				  WithDetail("retryable", true)
			}
			
			logger.Error("Login failed - service error", map[string]interface{}{
				"email":         req.Email,
				"ip":            clientIP,
				"user_agent":    userAgent,
				"failure_reason": failureReason,
				"error_code":    apiErr.Code,
				"grpc_error":    err.Error(),
				"layer":         "handler",
				"operation":     "login",
			})
		} else {
			failureReason = "unknown_error"
			userExists = false
			
			// Create detailed error for client
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeInternalError,
				"An unexpected error occurred during authentication. Please try again.",
				http.StatusInternalServerError,
				"handler",
				"login",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "authentication_process").
			  WithDetail("retryable", true)
			
			logger.Error("Login failed - unknown error", map[string]interface{}{
				"email":         req.Email,
				"ip":            clientIP,
				"user_agent":    userAgent,
				"failure_reason": failureReason,
				"grpc_error":    err.Error(),
				"layer":         "handler",
				"operation":     "login",
			})
		}
		
		// Log authentication attempt for security monitoring
		logger.LogAuthAttempt(req.Email, false, failureReason)
		
		// Log service call failure with detailed context
		logger.LogServiceCall("AccountService", "Login", false, err, map[string]interface{}{
			"email":       req.Email,
			"user_exists": userExists,
			"ip":          clientIP,
			"layer":       "handler",
			"operation":   "login",
		})
		
		// Log API request completion with failure
		logger.LogAPIRequest(r.Method, r.URL.Path, clientError.HTTPStatus, time.Since(startTime), map[string]interface{}{
			"email":          req.Email,
			"failure_reason": failureReason,
			"ip":             clientIP,
			"layer":          clientError.Layer,
			"operation":      clientError.Operation,
		})
		
		// Log the detailed APIError structure to terminal
		logger.Error("APIError details", clientError.GetLogContext())
		
		// Return the detailed error to client
		utils.HandleError(w, clientError, "login")
		return
	}

	// Authentication successful - log detailed success information
	logger.LogServiceCall("AccountService", "Login", true, nil, map[string]interface{}{
		"email":     req.Email,
		"ip":        clientIP,
		"layer":     "handler", 
		"operation": "login",
	})

	logger.LogAuthAttempt(req.Email, true, "credentials_validated")

	// Convert protobuf response to internal model
	user := mapping.ToPBUserRes(userRes)
	
	logger.Debug("Generating authentication tokens", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	})
		config := utils_config.GetConfig()
tokenMaker := utils.NewJWTTokenMaker(config.JWT.SecretKey)
	// Generate access token
accessToken, err := tokenMaker.CreateToken(user)
	if err != nil {
		logger.Error("Failed to generate access token", map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
			"error":   err.Error(),
			"ip":      clientIP,
		})
		
		tokenErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInternalError,
			"Authentication processing failed during token generation",
			http.StatusInternalServerError,
			"handler",
			"access_token_generation",
			err,
		).WithDetail("user_id", user.ID).
		  WithDetail("email", user.Email).
		  WithDetail("step", "access_token_generation")
		
		logger.Error("APIError details", tokenErr.GetLogContext())
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(startTime), map[string]interface{}{
			"email": req.Email,
			"error": "access_token_generation_failed",
			"ip":    clientIP,
		})
		
		utils.HandleError(w, tokenErr, "login")
		return
	}

	// Generate refresh token
		refreshToken, err := tokenMaker.CreateRefreshToken(user)
	if err != nil {
		logger.Error("Failed to generate refresh token", map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
			"error":   err.Error(),
			"ip":      clientIP,
		})
		
		tokenErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInternalError,
			"Authentication processing failed during refresh token generation",
			http.StatusInternalServerError,
			"handler",
			"refresh_token_generation",
			err,
		).WithDetail("user_id", user.ID).
		  WithDetail("email", user.Email).
		  WithDetail("step", "refresh_token_generation")
		
		logger.Error("APIError details", tokenErr.GetLogContext())
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(startTime), map[string]interface{}{
			"email": req.Email,
			"error": "refresh_token_generation_failed",
			"ip":    clientIP,
		})
		
		utils.HandleError(w, tokenErr, "login")
		return
	}

	// Prepare successful response
	response := model.LoginUserRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: model.AccountLoginResponse{
			ID:       user.ID,
			BranchID: user.BranchID,
			Name:     user.Name,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Title:    user.Title,
			Role:     string(user.Role),
			OwnerID:  user.OwnerID,
		},
	}

	// Log successful login completion with full context
	logger.Info("Login completed successfully", map[string]interface{}{
		"user_id":    user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"branch_id":  user.BranchID,
		"ip":         clientIP,
		"user_agent": userAgent,
		"duration":   time.Since(startTime).Milliseconds(),
		"layer":      "handler",
		"operation":  "login",
	})

	// Log successful API request
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(startTime), map[string]interface{}{
		"email":     req.Email,
		"user_id":   user.ID,
		"ip":        clientIP,
		"layer":     "handler",
		"operation": "login",
	})

	utils.RespondWithJSON(w, http.StatusOK, response, "login")
}

// Register with enhanced error handling
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientIP := utils.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	logger.Debug("Registration attempt initiated", map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	var req account_dto.RegisterUserRequest
	if err := utils.DecodeJSON(r.Body, &req, "register", false); err != nil {
		logger.Error("Failed to decode registration request", map[string]interface{}{
			"error":      err.Error(),
			"ip":         clientIP,
			"user_agent": userAgent,
		})
		utils.HandleError(w, err, "register")
		return
	}

	logger.Info("Registration attempt for email", map[string]interface{}{
		"email":      req.Email,
		"name":       req.Name,
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.LogValidationError("register_request", "Request validation failed", req)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors, "register")
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "register")
		}
		return
	}

	// Validate password
	if err := utils.ValidatePasswordWithDetails(req.Password, "register"); err != nil {
		logger.LogValidationError("password", "Password validation failed", "***hidden***")
		
		// Create detailed password validation error
		passwordErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeWeakPassword,
			"Password does not meet security requirements",
			http.StatusBadRequest,
			"handler",
			"register",
			err,
		).WithDetail("email", req.Email).
		  WithDetail("step", "password_validation")
		
		logger.Error("APIError details", passwordErr.GetLogContext())
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(startTime), map[string]interface{}{
			"email":     req.Email,
			"error":     "weak_password",
			"ip":        clientIP,
			"layer":     "handler",
			"operation": "register",
		})
		
		utils.HandleError(w, passwordErr, "register")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Error("Password hashing failed", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
			"ip":    clientIP,
		})
		
		hashErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInternalError,
			"Password processing failed during hashing",
			http.StatusInternalServerError,
			"handler",
			"password_hashing",
			err,
		).WithDetail("email", req.Email).
		  WithDetail("step", "password_hashing")
		
		logger.Error("APIError details", hashErr.GetLogContext())
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(startTime), map[string]interface{}{
			"email":     req.Email,
			"error":     "password_hashing_failed",
			"ip":        clientIP,
			"layer":     "handler",
			"operation": "register",
		})
		
		utils.HandleError(w, hashErr, "register")
		return
	}

	// Call registration service
	logger.Debug("Calling registration service", map[string]interface{}{
		"email":   req.Email,
		"name":    req.Name,
		"service": "AccountService",
		"method":  "Register",
	})

	userRes, err := h.userClient.Register(r.Context(), &pb.RegisterReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	})

	if err != nil {
		// Parse the gRPC error with enhanced context
		parsedErr := errorcustom.ParseGRPCError(err, "register", req.Email)
		var clientError *errorcustom.APIError
		
		if apiErr, ok := parsedErr.(*errorcustom.APIError); ok && apiErr.Code == errorcustom.ErrCodeDuplicateEmail {
			// Create detailed duplicate email error
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeDuplicateEmail,
				"An account with this email address already exists",
				http.StatusConflict,
				"handler",
				"register",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "email_uniqueness_check").
			  WithDetail("suggestion", "Try logging in instead or use a different email address")
			
			logger.Warning("Registration failed - email already exists", map[string]interface{}{
				"email":      req.Email,
				"name":       req.Name,
				"ip":         clientIP,
				"user_agent": userAgent,
				"layer":      "handler",
				"operation":  "register",
			})
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusConflict, time.Since(startTime), map[string]interface{}{
				"email":     req.Email,
				"error":     "duplicate_email",
				"ip":        clientIP,
				"layer":     "handler",
				"operation": "register",
			})
		} else {
			// Create detailed service error
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"Registration service is temporarily unavailable. Please try again later.",
				http.StatusServiceUnavailable,
				"handler",
				"register",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "service_call").
			  WithDetail("retryable", true)
			
			logger.Error("Registration failed - service error", map[string]interface{}{
				"email":      req.Email,
				"name":       req.Name,
				"grpc_error": err.Error(),
				"ip":         clientIP,
				"user_agent": userAgent,
				"layer":      "handler",
				"operation":  "register",
			})

			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(startTime), map[string]interface{}{
				"email":     req.Email,
				"error":     "service_error",
				"ip":        clientIP,
				"layer":     "handler",
				"operation": "register",
			})
		}

		logger.LogServiceCall("AccountService", "Register", false, err, map[string]interface{}{
			"email":     req.Email,
			"name":      req.Name,
			"ip":        clientIP,
			"layer":     "handler",
			"operation": "register",
		})

		// Log the detailed APIError structure to terminal
		logger.Error("APIError details", clientError.GetLogContext())

		utils.HandleError(w, clientError, "register")
		return
	}

	// Log successful registration
	logger.LogServiceCall("AccountService", "Register", true, nil, map[string]interface{}{
		"email":     req.Email,
		"name":      req.Name,
		"user_id":   userRes.Id,
		"ip":        clientIP,
		"layer":     "handler",
		"operation": "register",
	})

	logger.Info("Registration completed successfully", map[string]interface{}{
		"user_id":    userRes.Id,
		"email":      userRes.Email,
		"name":       userRes.Name,
		"ip":         clientIP,
		"user_agent": userAgent,
		"duration":   time.Since(startTime).Milliseconds(),
		"layer":      "handler",
		"operation":  "register",
	})

	response := account_dto.RegisterResponse{
		ID:     userRes.Id,
		Name:   userRes.Name,
		Email:  userRes.Email,
		Status: userRes.Success,
	}

	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusCreated, time.Since(startTime), map[string]interface{}{
		"email":     req.Email,
		"user_id":   userRes.Id,
		"ip":        clientIP,
		"layer":     "handler",
		"operation": "register",
	})

	utils.RespondWithJSON(w, http.StatusCreated, response, "register")
}

// Logout with detailed logging
func (h *AccountHandler) Logout(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientIP := utils.GetClientIP(r)
	
	// Extract user context if available (from JWT middleware)
	userID, _ := h.getUserIDFromContext(r.Context())
	userEmail := utils.GetUserEmailFromContext(r)
	
	logger.Info("User logout initiated", map[string]interface{}{
		"user_id":    userID,
		"user_email": userEmail,
		"ip":         clientIP,
		"layer":      "handler",
		"operation":  "logout",
	})

	// Here you could invalidate tokens if you maintain a token blacklist
	// h.tokenService.InvalidateTokens(userID)

	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(startTime), map[string]interface{}{
		"user_id":   userID,
		"ip":        clientIP,
		"layer":     "handler",
		"operation": "logout",
	})

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	}, "logout")
}
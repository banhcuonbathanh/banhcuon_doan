// internal/account/account_handler/account_handler_auth.go
package account_handler

import (
	"net/http"

	"time"

	"english-ai-full/internal/account/account_dto"
	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/mapping"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)






// new 


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
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		logger.Error("Failed to decode login request", map[string]interface{}{
			"error":      err.Error(),
			"ip":         clientIP,
			"user_agent": userAgent,
		})
		utils.HandleError(w, err)
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
			utils.HandleValidationErrors(w, validationErrors)
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			))
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
		
		// Determine specific failure reason for logging
		var failureReason string
		var userExists bool
		
		if errorcustom.IsUserNotFoundError(parsedErr) {
			failureReason = "email_not_found"
			userExists = false
			logger.Warning("Login failed - email not found", map[string]interface{}{
				"email":         req.Email,
				"ip":            clientIP,
				"user_agent":    userAgent,
				"failure_reason": failureReason,
			})
		} else if errorcustom.IsPasswordError(parsedErr) {
			failureReason = "password_mismatch"
			userExists = true
			logger.Warning("Login failed - password mismatch", map[string]interface{}{
				"email":         req.Email,
				"ip":            clientIP,
				"user_agent":    userAgent,
				"failure_reason": failureReason,
			})
		} else if apiErr, ok := parsedErr.(*errorcustom.APIError); ok {
			if apiErr.Code == errorcustom.ErrCodeAccessDenied {
				failureReason = "account_disabled_or_locked"
				userExists = true
			} else {
				failureReason = "service_error"
				userExists = false
			}
			logger.Error("Login failed - service error", map[string]interface{}{
				"email":         req.Email,
				"ip":            clientIP,
				"user_agent":    userAgent,
				"failure_reason": failureReason,
				"error_code":    apiErr.Code,
				"grpc_error":    err.Error(),
			})
		} else {
			failureReason = "unknown_error"
			userExists = false
			logger.Error("Login failed - unknown error", map[string]interface{}{
				"email":         req.Email,
				"ip":            clientIP,
				"user_agent":    userAgent,
				"failure_reason": failureReason,
				"grpc_error":    err.Error(),
			})
		}
		
		// Log authentication attempt for security monitoring
		logger.LogAuthAttempt(req.Email, false, failureReason)
		
		// Log service call failure with detailed context
		logger.LogServiceCall("AccountService", "Login", false, err, map[string]interface{}{
			"email":       req.Email,
			"user_exists": userExists,
			"ip":          clientIP,
		})
		
		// Record login attempt for rate limiting (if implemented)
		// h.loginAttemptTracker.RecordAttempt(r.Context(), req.Email, clientIP, userAgent, false, failureReason)
		
		// Log API request completion with failure
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusUnauthorized, time.Since(startTime), map[string]interface{}{
			"email":          req.Email,
			"failure_reason": failureReason,
			"ip":             clientIP,
		})
		
		// Return generic authentication error to client (security best practice)
		// We don't want to leak information about whether the email exists or not
		authErr := errorcustom.NewAuthenticationError("invalid credentials")
		utils.HandleError(w, authErr)
		return
	}

	// Authentication successful - log detailed success information
	logger.LogServiceCall("AccountService", "Login", true, nil, map[string]interface{}{
		"email":     req.Email,
	
	
		"ip":        clientIP,
	})

	logger.LogAuthAttempt(req.Email, true, "credentials_validated")

	// Convert protobuf response to internal model
	user := mapping.ToPBUserRes(userRes)
	
	logger.Debug("Generating authentication tokens", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	})

	// Generate access token
	accessToken, err := utils.GenerateJWTToken(user)
	if err != nil {
		logger.Error("Failed to generate access token", map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
			"error":   err.Error(),
			"ip":      clientIP,
		})
		
		tokenErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInternalError,
			"Authentication processing failed",
			http.StatusInternalServerError,
			"handler",
			"access_token_generation",
			err,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(startTime), map[string]interface{}{
			"email": req.Email,
			"error": "access_token_generation_failed",
			"ip":    clientIP,
		})
		
		utils.HandleError(w, tokenErr)
		return
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		logger.Error("Failed to generate refresh token", map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
			"error":   err.Error(),
			"ip":      clientIP,
		})
		
		tokenErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInternalError,
			"Authentication processing failed",
			http.StatusInternalServerError,
			"handler",
			"refresh_token_generation",
			err,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(startTime), map[string]interface{}{
			"email": req.Email,
			"error": "refresh_token_generation_failed",
			"ip":    clientIP,
		})
		
		utils.HandleError(w, tokenErr)
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

	// Record successful login attempt for rate limiting (if implemented)
	// h.loginAttemptTracker.RecordAttempt(r.Context(), req.Email, clientIP, userAgent, true, "")

	// Log successful login completion with full context
	logger.Info("Login completed successfully", map[string]interface{}{
		"user_id":    user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"branch_id":  user.BranchID,
		"ip":         clientIP,
		"user_agent": userAgent,
		"duration":   time.Since(startTime).Milliseconds(),
	})

	// Log successful API request
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(startTime), map[string]interface{}{
		"email":   req.Email,
		"user_id": user.ID,
		"ip":      clientIP,
	})

	utils.RespondWithJSON(w, http.StatusOK, response)
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
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		logger.Error("Failed to decode registration request", map[string]interface{}{
			"error":      err.Error(),
			"ip":         clientIP,
			"user_agent": userAgent,
		})
		utils.HandleError(w, err)
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
			utils.HandleValidationErrors(w, validationErrors)
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			))
		}
		return
	}

	// Validate password
	if err := utils.ValidatePasswordWithDetails(req.Password); err != nil {
		logger.LogValidationError("password", "Password validation failed", "***hidden***")
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(startTime), map[string]interface{}{
			"email": req.Email,
			"error": "weak_password",
			"ip":    clientIP,
		})
		utils.HandleError(w, err)
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
			"Password processing failed",
			http.StatusInternalServerError,
			"handler",
			"password_hashing",
			err,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(startTime), map[string]interface{}{
			"email": req.Email,
			"error": "password_hashing_failed",
			"ip":    clientIP,
		})
		
		utils.HandleError(w, hashErr)
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
		
		if apiErr, ok := parsedErr.(*errorcustom.APIError); ok && apiErr.Code == errorcustom.ErrCodeDuplicateEmail {
			logger.Warning("Registration failed - email already exists", map[string]interface{}{
				"email":      req.Email,
				"name":       req.Name,
				"ip":         clientIP,
				"user_agent": userAgent,
			})
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusConflict, time.Since(startTime), map[string]interface{}{
				"email": req.Email,
				"error": "duplicate_email",
				"ip":    clientIP,
			})
			
			utils.HandleError(w, errorcustom.NewDuplicateEmailError(req.Email))
			return
		} else {
			logger.Error("Registration failed - service error", map[string]interface{}{
				"email":      req.Email,
				"name":       req.Name,
				"grpc_error": err.Error(),
				"ip":         clientIP,
				"user_agent": userAgent,
			})
		}

		logger.LogServiceCall("AccountService", "Register", false, err, map[string]interface{}{
			"email": req.Email,
			"name":  req.Name,
			"ip":    clientIP,
		})

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(startTime), map[string]interface{}{
			"email": req.Email,
			"error": "service_error",
			"ip":    clientIP,
		})

		utils.HandleError(w, parsedErr)
		return
	}

	// Log successful registration
	logger.LogServiceCall("AccountService", "Register", true, nil, map[string]interface{}{
		"email":   req.Email,
		"name":    req.Name,
		"user_id": userRes.Id,
		"ip":      clientIP,
	})

	logger.Info("Registration completed successfully", map[string]interface{}{
		"user_id":    userRes.Id,
		"email":      userRes.Email,
		"name":       userRes.Name,
		"ip":         clientIP,
		"user_agent": userAgent,
		"duration":   time.Since(startTime).Milliseconds(),
	})

	response := account_dto.RegisterResponse{
		ID:     userRes.Id,
		Name:   userRes.Name,
		Email:  userRes.Email,
		Status: userRes.Success,
	}

	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusCreated, time.Since(startTime), map[string]interface{}{
		"email":   req.Email,
		"user_id": userRes.Id,
		"ip":      clientIP,
	})

	utils.RespondWithJSON(w, http.StatusCreated, response)
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
	})

	// Here you could invalidate tokens if you maintain a token blacklist
	// h.tokenService.InvalidateTokens(userID)

	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(startTime), map[string]interface{}{
		"user_id": userID,
		"ip":      clientIP,
	})

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}
// new done 
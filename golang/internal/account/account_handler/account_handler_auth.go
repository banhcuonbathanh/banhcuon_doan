// internal/account/account_handler/account_handler_auth.go
package account_handler

import (
	"fmt"
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
	start := time.Now()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("login")
	
	// Extract client information for security logging
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	// Create base context with client information
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})
	
	handlerLog.Info("Login attempt initiated", baseContext)

	// Decode request body
	var req account_dto.LoginRequest
	if err := errorcustom.DecodeJSON(r.Body, &req, "login", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to decode login request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)
		
		// Log API request with decode error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.HandleError(w, err, "login")
		return
	}

	// Add email to context for subsequent logging
	baseContext["email"] = req.Email
	
	handlerLog.Info("Login attempt for user", baseContext)

	// Validate request structure
	if err := h.validator.Struct(&req); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"validation_error": err.Error(),
		})
		
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Log each validation error individually
			for _, validationError := range validationErrors {
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					validationError.Value(),
				)
			}
			
			logger.WarningWithCause(
				"Login request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			errorcustom.HandleValidationErrors(w, validationErrors, "login", req.Email)
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during login",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "login")
		}
		
		// Log API request with validation error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Request validation passed", baseContext)

	// Call authentication service
	serviceStart := time.Now()
	handlerLog.Debug("Calling authentication service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "Login",
	}))

	userRes, err := h.userClient.Login(r.Context(), &pb.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	serviceDuration := time.Since(serviceStart)

	if err != nil {
		// Parse the gRPC error with enhanced context
	parsedErr := errorcustom.ParseGRPCError(err, "login", req.Email, map[string]interface{}{
	"service":     "AccountService",
	"method":      "Login",
	"duration_ms": serviceDuration.Milliseconds(),
	"client_ip":   clientIP,
	"user_agent":  userAgent,
})
		
		// Determine specific failure reason for logging and client response
		var failureReason string
		var userExists bool
		var clientError *errorcustom.APIError
		var httpStatus int
		
		if errorcustom.IsUserNotFoundError(parsedErr) {
			failureReason = "email_not_found"
			userExists = false
			httpStatus = http.StatusNotFound
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeUserNotFound,
				"User with this email address was not found",
				httpStatus,
				"handler",
				"login",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "email_verification").
			  WithDetail("user_found", false)
			
			logger.WarningWithCause(
				"Login failed - user not found",
				failureReason,
				logger.LayerHandler,
				"authenticate_user",
				utils.MergeContext(baseContext, map[string]interface{}{
					"user_exists": userExists,
				}),
			)
			
		} else if errorcustom.IsPasswordError(parsedErr) {
			failureReason = "password_mismatch"
			userExists = true
			httpStatus = http.StatusUnauthorized
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeAuthFailed,
				"The password you entered is incorrect",
				httpStatus,
				"handler",
				"login",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "password_verification").
			  WithDetail("user_found", true)
			
			logger.WarningWithCause(
				"Login failed - invalid credentials",
				failureReason,
				logger.LayerHandler,
				"authenticate_user",
				utils.MergeContext(baseContext, map[string]interface{}{
					"user_exists": userExists,
				}),
			)
			
			// Log security event for failed password attempts
			logger.LogSecurityEvent(
				"failed_login_attempt",
				"Invalid password provided for existing user",
				"medium",
				utils.MergeContext(baseContext, map[string]interface{}{
					"attempt_type": "password_mismatch",
				}),
			)
			
		} else if apiErr, ok := parsedErr.(*errorcustom.APIError); ok {
			if apiErr.Code == errorcustom.ErrCodeAccessDenied {
				failureReason = "account_disabled_or_locked"
				userExists = true
				httpStatus = http.StatusForbidden
				
	if errorcustom.IsUserNotFoundError(parsedErr) {
	failureReason = "email_not_found"
	userExists = false
	httpStatus = http.StatusNotFound
	
	clientError = errorcustom.NewAPIErrorWithContext(
		errorcustom.ErrCodeUserNotFound,
		"User with this email address was not found",
		httpStatus,
		errorcustom.DomainUser,  // Add the domain parameter
		"handler",
		"login",
		err,
	).WithDetail("email", req.Email).
	  WithDetail("step", "email_verification").
	  WithDetail("user_found", false)
				
				// Log security event for account access attempts
				logger.LogSecurityEvent(
					"disabled_account_access_attempt",
					"Login attempt on disabled/locked account",
					"high",
					utils.MergeContext(baseContext, map[string]interface{}{
						"account_status": "disabled_or_locked",
					}),
				)
			} else {
				failureReason = "service_error"
				userExists = false
				httpStatus = http.StatusServiceUnavailable
				
				clientError = errorcustom.NewAPIErrorWithContext(
					errorcustom.ErrCodeServiceError,
					"Authentication service is temporarily unavailable. Please try again later.",
					httpStatus,
					"handler",
					"login",
					err,
				).WithDetail("email", req.Email).
				  WithDetail("step", "service_call").
				  WithDetail("retryable", true)
			}
			
			logger.ErrorWithCause(
				"Login failed - service error",
				failureReason,
				logger.LayerExternal,
				"authenticate_user",
				utils.MergeContext(baseContext, map[string]interface{}{
					"error_code":  apiErr.Code,
					"grpc_error":  err.Error(),
					"user_exists": userExists,
				}),
			)
		} else {
			failureReason = "unknown_error"
			userExists = false
			httpStatus = http.StatusInternalServerError
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeInternalError,
				"An unexpected error occurred during authentication. Please try again.",
				httpStatus,
				"handler",
				"login",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "authentication_process").
			  WithDetail("retryable", true)
			
			logger.ErrorWithCause(
				"Login failed - unknown error",
				failureReason,
				logger.LayerHandler,
				"authenticate_user",
				utils.MergeContext(baseContext, map[string]interface{}{
					"grpc_error":  err.Error(),
					"user_exists": userExists,
				}),
			)
		}
		
		// Log authentication attempt for security monitoring
		logger.LogAuthAttempt(req.Email, false, failureReason, utils.MergeContext(baseContext, map[string]interface{}{
			"user_exists": userExists,
		}))
		
		// Log service call failure
		logger.LogServiceCall(
			"AccountService",
			"Login",
			false,
			err,
			utils.MergeContext(baseContext, map[string]interface{}{
				"user_exists":  userExists,
				"duration_ms":  serviceDuration.Milliseconds(),
			}),
		)
		
		// Log API request completion with failure
		logger.LogAPIRequest(
			r.Method,
			r.URL.Path,
			httpStatus,
			time.Since(start),
			utils.MergeContext(baseContext, map[string]interface{}{
				"failure_reason": failureReason,
				"user_exists":    userExists,
			}),
		)
		
		// Log the detailed APIError structure
		handlerLog.Error("Login failed with detailed error", clientError.GetLogContext())
		
		errorcustom.HandleError(w, clientError, "login")
		return
	}

	// Authentication successful
	logger.LogServiceCall(
		"AccountService",
		"Login",
		true,
		nil,
		utils.MergeContext(baseContext, map[string]interface{}{
			"duration_ms": serviceDuration.Milliseconds(),
		}),
	)

	logger.LogAuthAttempt(req.Email, true, "credentials_validated", baseContext)

	// Convert protobuf response to internal model
	user := mapping.ToPBUserRes(userRes)
	
	// Add user information to context
	userContext := utils.MergeContext(baseContext, map[string]interface{}{
		"user_id": user.ID,
		"role":    user.Role,
	})
	
	handlerLog.Debug("Generating authentication tokens", userContext)

	// Generate tokens
	config := utils_config.GetConfig()
	tokenMaker := utils.NewJWTTokenMaker(config.JWT.SecretKey)
	
	// Generate access token
	accessToken, err := tokenMaker.CreateToken(user)
	if err != nil {
		context := utils.MergeContext(userContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to generate access token",
			"token_generation_error",
			logger.LayerHandler,
			"generate_access_token",
			context,
		)
		
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
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), context)
		
		errorcustom.HandleError(w, tokenErr, "login")
		return
	}

	// Generate refresh token
	refreshToken, err := tokenMaker.CreateRefreshToken(user)
	if err != nil {
		context := utils.MergeContext(userContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to generate refresh token",
			"token_generation_error",
			logger.LayerHandler,
			"generate_refresh_token",
			context,
		)
		
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
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), context)
		
		errorcustom.HandleError(w, tokenErr, "login")
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

	// Final context with complete information
	finalContext := utils.MergeContext(userContext, map[string]interface{}{
		"branch_id":   user.BranchID,
		"duration_ms": time.Since(start).Milliseconds(),
	})

	// Log successful user activity
	logger.LogUserActivity(
		string(rune(user.ID)) ,
		user.Email,
		"login",
		"authentication",
		finalContext,
	)

	// Log performance
	logger.LogPerformance("user_login", time.Since(start), finalContext)

	// Log successful login completion
	handlerLog.Info("Login completed successfully", finalContext)

	// Log successful API request
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), finalContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, response, "login")
}

// Register with enhanced error handling
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("register")
	
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})
	
	handlerLog.Info("Registration attempt initiated", baseContext)

	// Decode request body
	var req account_dto.RegisterUserRequest
	if err := errorcustom.DecodeJSON(r.Body, &req, "register", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to decode registration request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.HandleError(w, err, "register")
		return
	}

	// Add user info to context
	baseContext["email"] = req.Email
	baseContext["name"] = req.Name
	
	handlerLog.Info("Registration attempt for user", baseContext)

	// Validate request structure
	if err := h.validator.Struct(req); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"validation_error": err.Error(),
		})
		
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Log each validation error
			for _, validationError := range validationErrors {
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					validationError.Value(),
				)
			}
			
			logger.WarningWithCause(
				"Registration request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			errorcustom.HandleValidationErrors(w, validationErrors, "register")
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during registration",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "register")
		}
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Request validation passed", baseContext)

	// Validate password with detailed logging
	if err := errorcustom.ValidatePasswordWithDetails(req.Password, "register"); err != nil {
		logger.LogValidationError("password", "Password validation failed", "***masked***")
		
		logger.WarningWithCause(
			"Password validation failed",
			"weak_password",
			logger.LayerHandler,
			"validate_password",
			baseContext,
		)
		
		passwordErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeWeakPassword,
			"Password does not meet security requirements",
			http.StatusBadRequest,
			"handler",
			"register",
			err,
		).WithDetail("email", req.Email).
		  WithDetail("step", "password_validation")
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), baseContext)
		
		errorcustom.HandleError(w, passwordErr, "register")
		return
	}

	handlerLog.Debug("Password validation passed", baseContext)

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Password hashing failed",
			"password_hashing_error",
			logger.LayerHandler,
			"hash_password",
			context,
		)
		
		hashErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInternalError,
			"Password processing failed during hashing",
			http.StatusInternalServerError,
			"handler",
			"password_hashing",
			err,
		).WithDetail("email", req.Email).
		  WithDetail("step", "password_hashing")
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), context)
		
		errorcustom.HandleError(w, hashErr, "register")
		return
	}

	// Call registration service
	serviceStart := time.Now()
	handlerLog.Debug("Calling registration service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "Register",
	}))

	userRes, err := h.userClient.Register(r.Context(), &pb.RegisterReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	})
	serviceDuration := time.Since(serviceStart)

	if err != nil {
		// Parse the gRPC error
		parsedErr := errorcustom.ParseGRPCError(err, "register", req.Email)
		var clientError *errorcustom.APIError
		var httpStatus int
		var failureReason string
		
		if apiErr, ok := parsedErr.(*errorcustom.APIError); ok && apiErr.Code == errorcustom.ErrCodeDuplicateEmail {
			failureReason = "duplicate_email"
			httpStatus = http.StatusConflict
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeDuplicateEmail,
				"An account with this email address already exists",
				httpStatus,
				"handler",
				"register",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "email_uniqueness_check").
			  WithDetail("suggestion", "Try logging in instead or use a different email address")
			
			logger.WarningWithCause(
				"Registration failed - duplicate email",
				failureReason,
				logger.LayerHandler,
				"check_email_uniqueness",
				baseContext,
			)
			
		} else {
			failureReason = "service_error"
			httpStatus = http.StatusServiceUnavailable
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"Registration service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"register",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "service_call").
			  WithDetail("retryable", true)
			
			logger.ErrorWithCause(
				"Registration failed - service error",
				failureReason,
				logger.LayerExternal,
				"create_user",
				utils.MergeContext(baseContext, map[string]interface{}{
					"grpc_error": err.Error(),
				}),
			)
		}

		// Log service call failure
		logger.LogServiceCall(
			"AccountService",
			"Register",
			false,
			err,
			utils.MergeContext(baseContext, map[string]interface{}{
				"duration_ms": serviceDuration.Milliseconds(),
			}),
		)

		// Log API request completion
		logger.LogAPIRequest(
			r.Method,
			r.URL.Path,
			httpStatus,
			time.Since(start),
			utils.MergeContext(baseContext, map[string]interface{}{
				"failure_reason": failureReason,
			}),
		)

		handlerLog.Error("Registration failed with detailed error", clientError.GetLogContext())
		
		errorcustom.HandleError(w, clientError, "register")
		return
	}

	// Registration successful
	userContext := utils.MergeContext(baseContext, map[string]interface{}{
		"user_id": userRes.Id,
	})

	// Log successful service call
	logger.LogServiceCall(
		"AccountService",
		"Register",
		true,
		nil,
		utils.MergeContext(userContext, map[string]interface{}{
			"duration_ms": serviceDuration.Milliseconds(),
		}),
	)

	// Log user activity
	logger.LogUserActivity(
		fmt.Sprint(userRes.Id) ,
		userRes.Email,
		"create",
		"user_account",
		userContext,
	)

	// Log performance
	logger.LogPerformance("user_registration", time.Since(start), userContext)

	handlerLog.Info("Registration completed successfully", utils.MergeContext(userContext, map[string]interface{}{
		"duration_ms": time.Since(start).Milliseconds(),
	}))

	response := account_dto.RegisterResponse{
		ID:     userRes.Id,
		Name:   userRes.Name,
		Email:  userRes.Email,
		Status: userRes.Success,
	}

	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusCreated, time.Since(start), userContext)

	errorcustom.RespondWithJSON(w, http.StatusCreated, response, "register")
}

// Logout with detailed logging
func (h *AccountHandler) Logout(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("logout")
	
	clientIP := errorcustom.GetClientIP(r)
	
	// Extract user context if available (from JWT middleware)
	userID, _ := h.getUserIDFromContext(r.Context())
	userEmail := errorcustom.GetUserEmailFromContext(r)
	
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"user_id":    userID,
		"user_email": userEmail,
		"ip":         clientIP,
	})
	
	handlerLog.Info("User logout initiated", baseContext)

	// Here you could invalidate tokens if you maintain a token blacklist
	// Example: h.tokenService.InvalidateTokens(userID)
	


	// Log performance
	logger.LogPerformance("user_logout", time.Since(start), baseContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), baseContext)

	handlerLog.Info("Logout completed successfully", baseContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	}, "logout")
}
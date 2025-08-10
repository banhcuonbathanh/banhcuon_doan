package account_handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)

// ChangePassword handles password change requests with comprehensive logging and error handling
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("change_password")

	// Extract client information for security logging
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create base context with client information
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	handlerLog.Info("Password change request initiated", baseContext)

	// Extract user ID from context
	userID, err := h.getUserIDFromContext(ctx)
	if err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})

		logger.ErrorWithCause(
			"Failed to extract user ID from context",
			"missing_user_context",
			logger.LayerHandler,
			"extract_user_id",
			context,
		)

		// Log security event for unauthorized access attempt
		logger.LogSecurityEvent(
			"unauthorized_password_change_attempt",
			"Password change request without valid user context",
			"high",
			context,
		)

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusUnauthorized, time.Since(start), context)

		errorcustom.HandleError(w, err, "change_password")
		return
	}

	// Add user ID to context for subsequent logs
	baseContext["user_id"] = userID

	handlerLog.Info("Password change request for authenticated user", baseContext)

	// Decode request body
	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,password"`
	}

	if err := errorcustom.DecodeJSON(r.Body, &req, "change_password", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})

		logger.ErrorWithCause(
			"Failed to decode password change request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)

		errorcustom.HandleError(w, err, "change_password")
		return
	}

	handlerLog.Debug("Request body decoded successfully", baseContext)

	// Validate request structure
	if err := h.validator.Struct(&req); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"validation_error": err.Error(),
		})

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Log each validation error
			for _, validationError := range validationErrors {
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					"***masked***", // Mask password values
				)
			}

			logger.WarningWithCause(
				"Password change request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)

			errorcustom.HandleValidationErrors(w, validationErrors, "change_password")
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during password change",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)

			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "change_password")
		}

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Request validation passed", baseContext)

	// Validate new password strength
	if err := errorcustom.ValidatePasswordWithDetails(req.NewPassword, "change_password"); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"password_validation_error": err.Error(),
		})

		logger.LogValidationError("new_password", "Password strength validation failed", "***masked***")

		logger.WarningWithCause(
			"New password validation failed",
			"weak_password",
			logger.LayerHandler,
			"validate_password_strength",
			context,
		)

		passwordErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeWeakPassword,
			"New password does not meet security requirements",
			http.StatusBadRequest,
			"handler",
			"change_password",
			err,
		).WithDetail("user_id", fmt.Sprint(userID)).
		  WithDetail("step", "password_strength_validation")

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)

		errorcustom.HandleError(w, passwordErr, "change_password")
		return
	}

	handlerLog.Debug("Password strength validation passed", baseContext)

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
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
		).WithDetail("user_id", fmt.Sprint(userID)).
		  WithDetail("step", "password_hashing")

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), context)

		errorcustom.HandleError(w, hashErr, "change_password")
		return
	}

	// Call password change service
	serviceStart := time.Now()
	handlerLog.Debug("Calling password change service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "ChangePassword",
	}))

	res, err := h.userClient.ChangePassword(ctx, &pb.ChangePasswordReq{
		UserId:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     hashedPassword,
	})
	serviceDuration := time.Since(serviceStart)

	if err != nil {
	

		var failureReason string
		var clientError *errorcustom.APIError
		var httpStatus int

		if strings.Contains(err.Error(), "invalid password") {
			failureReason = "invalid_current_password"
			httpStatus = http.StatusBadRequest

			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeInvalidInput,
				"The current password you entered is incorrect",
				httpStatus,
				"handler",
				"change_password",
				err,
			).WithDetail("user_id", fmt.Sprint(userID)).
			  WithDetail("step", "current_password_verification")

			logger.WarningWithCause(
				"Password change failed - invalid current password",
				failureReason,
				logger.LayerHandler,
				"change_password",
				baseContext,
			)

			// Log security event for invalid password attempts
			logger.LogSecurityEvent(
				"invalid_password_change_attempt",
				"User attempted to change password with incorrect current password",
				"medium",
				utils.MergeContext(baseContext, map[string]interface{}{
					"attempt_type": "invalid_current_password",
				}),
			)

		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError

			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"Password change service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"change_password",
				err,
			).WithDetail("user_id", fmt.Sprint(userID)).
			  WithDetail("step", "service_call").
			  WithDetail("retryable", true)

			logger.ErrorWithCause(
				"Password change failed - service error",
				failureReason,
				logger.LayerExternal,
				"change_password",
				utils.MergeContext(baseContext, map[string]interface{}{
					"grpc_error": err.Error(),
				}),
			)
		}

		// Log service call failure
		logger.LogServiceCall(
			"AccountService",
			"ChangePassword",
			false,
			err,
			utils.MergeContext(baseContext, map[string]interface{}{
				"duration_ms": serviceDuration.Milliseconds(),
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
			}),
		)

		handlerLog.Error("Password change failed with detailed error", clientError.GetLogContext())

		errorcustom.HandleError(w, clientError, "change_password")
		return
	}

	// Password change successful
	logger.LogServiceCall(
		"AccountService",
		"ChangePassword",
		true,
		nil,
		utils.MergeContext(baseContext, map[string]interface{}{
			"duration_ms": serviceDuration.Milliseconds(),
		}),
	)

	// Log user activity for successful password change
	logger.LogUserActivity(
		fmt.Sprint(userID),
		"unknown", // email not available in this context
		"update",
		"password_change",
		baseContext,
	)

	// Log security event for successful password change
	logger.LogSecurityEvent(
		"password_changed",
		"User successfully changed password",
		"low",
		utils.MergeContext(baseContext, map[string]interface{}{
			"action": "password_change_successful",
		}),
	)

	// Log performance
	totalDuration := time.Since(start)
	logger.LogPerformance("change_password", totalDuration, baseContext)

	// Log API request completion
	logger.LogAPIRequest(
		r.Method,
		r.URL.Path,
		http.StatusOK,
		totalDuration,
		utils.MergeContext(baseContext, map[string]interface{}{
			"password_change_success": res.Success,
			"service_message":         res.Message,
		}),
	)

	handlerLog.Info("Password change completed successfully", utils.MergeContext(baseContext, map[string]interface{}{
		"service_message": res.Message,
	}))

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "change_password")
}

// ForgotPassword handles forgot password requests with comprehensive logging and security measures
func (h *AccountHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("forgot_password")

	// Extract client information for security logging
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create base context
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	handlerLog.Info("Forgot password request initiated", baseContext)

	// Decode request body
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := errorcustom.DecodeJSON(r.Body, &req, "forgot_password", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})

		logger.ErrorWithCause(
			"Failed to decode forgot password request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)

		errorcustom.HandleError(w, err, "forgot_password")
		return
	}

	// Add email to context (masked for privacy)
	baseContext["email"] = utils.MaskSensitiveValue("email", req.Email)

	handlerLog.Info("Forgot password request for email", baseContext)

	// Validate request structure
	if err := h.validator.Struct(&req); err != nil {
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
				"Forgot password request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)

			errorcustom.HandleValidationErrors(w, validationErrors, "forgot_password")
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during forgot password",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)

			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "forgot_password")
		}

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Request validation passed", baseContext)

	// Call forgot password service
	serviceStart := time.Now()
	handlerLog.Debug("Calling forgot password service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "ForgotPassword",
	}))

	res, err := h.userClient.ForgotPassword(ctx, &pb.ForgotPasswordReq{
		Email: req.Email,
	})
	serviceDuration := time.Since(serviceStart)

	if err != nil {
		var failureReason string

		// Log service call failure but don't reveal details for security
		logger.LogServiceCall(
			"AccountService",
			"ForgotPassword",
			false,
			err,
			utils.MergeContext(baseContext, map[string]interface{}{
				"duration_ms": serviceDuration.Milliseconds(),
			}),
		)

		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "user not found") {
			failureReason = "user_not_found"

			logger.WarningWithCause(
				"Forgot password - user not found",
				failureReason,
				logger.LayerHandler,
				"forgot_password",
				baseContext,
			)

			// Log security event for email enumeration attempt
			logger.LogSecurityEvent(
				"password_reset_unknown_email",
				"Password reset requested for non-existent email",
				"low",
				baseContext,
			)
		} else {
			failureReason = "service_error"

			logger.ErrorWithCause(
				"Forgot password failed - service error",
				failureReason,
				logger.LayerExternal,
				"forgot_password",
				utils.MergeContext(baseContext, map[string]interface{}{
					"grpc_error": err.Error(),
				}),
			)
		}

		// Always return success for security reasons (prevent email enumeration)
		logger.LogAPIRequest(
			r.Method,
			r.URL.Path,
			http.StatusOK,
			time.Since(start),
			utils.MergeContext(baseContext, map[string]interface{}{
				"failure_reason":    failureReason,
				"security_response": true,
			}),
		)

		handlerLog.Info("Forgot password request completed with security response", utils.MergeContext(baseContext, map[string]interface{}{
			"security_message": "Generic success message returned for security",
		}))

		errorcustom.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If the email exists, a password reset link has been sent",
		}, "forgot_password")
		return
	}

	// Forgot password successful
	logger.LogServiceCall(
		"AccountService",
		"ForgotPassword",
		true,
		nil,
		utils.MergeContext(baseContext, map[string]interface{}{
			"duration_ms": serviceDuration.Milliseconds(),
		}),
	)

	// Log user activity
	logger.LogUserActivity(
		"unknown", // user_id not available
		req.Email,
		"request",
		"password_reset",
		baseContext,
	)

	// Log security event for legitimate password reset request
	logger.LogSecurityEvent(
		"password_reset_requested",
		"User requested password reset",
		"low",
		utils.MergeContext(baseContext, map[string]interface{}{
			"action": "password_reset_request",
		}),
	)

	// Log performance
	totalDuration := time.Since(start)
	logger.LogPerformance("forgot_password", totalDuration, baseContext)

	// Log API request completion
	logger.LogAPIRequest(
		r.Method,
		r.URL.Path,
		http.StatusOK,
		totalDuration,
		utils.MergeContext(baseContext, map[string]interface{}{
			"reset_request_success": res.Success,
			"service_message":       res.Message,
		}),
	)

	handlerLog.Info("Forgot password request completed successfully", utils.MergeContext(baseContext, map[string]interface{}{
		"service_message": res.Message,
	}))

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "forgot_password")
}

// ResetPassword handles password reset requests with comprehensive logging and security measures
func (h *AccountHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("reset_password")

	// Extract client information for security logging
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create base context
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	handlerLog.Info("Password reset request initiated", baseContext)

	// Decode request body
	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,password"`
	}

	if err := errorcustom.DecodeJSON(r.Body, &req, "reset_password", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})

		logger.ErrorWithCause(
			"Failed to decode password reset request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)

		errorcustom.HandleError(w, err, "reset_password")
		return
	}

	// Add token info to context (masked for security)
	baseContext["token"] = utils.MaskSensitiveValue("token", req.Token)

	handlerLog.Info("Password reset request with token", baseContext)

	// Validate request structure
	if err := h.validator.Struct(&req); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"validation_error": err.Error(),
		})

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Log each validation error (mask password fields)
			for _, validationError := range validationErrors {
				value := validationError.Value()
				if validationError.Field() == "NewPassword" {
					value = "***masked***"
				}
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					value,
				)
			}

			logger.WarningWithCause(
				"Password reset request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)

			errorcustom.HandleValidationErrors(w, validationErrors, "reset_password")
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during password reset",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)

			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "reset_password")
		}

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Request validation passed", baseContext)

	// Validate new password strength
	if err := errorcustom.ValidatePasswordWithDetails(req.NewPassword, "reset_password"); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"password_validation_error": err.Error(),
		})

		logger.LogValidationError("new_password", "Password strength validation failed", "***masked***")

		logger.WarningWithCause(
			"New password validation failed during reset",
			"weak_password",
			logger.LayerHandler,
			"validate_password_strength",
			context,
		)

		passwordErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeWeakPassword,
			"New password does not meet security requirements",
			http.StatusBadRequest,
			"handler",
			"reset_password",
			err,
		).WithDetail("step", "password_strength_validation")

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)

		errorcustom.HandleError(w, passwordErr, "reset_password")
		return
	}

	handlerLog.Debug("Password strength validation passed", baseContext)

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})

		logger.ErrorWithCause(
			"Password hashing failed during reset",
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
		).WithDetail("step", "password_hashing")

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), context)

		errorcustom.HandleError(w, hashErr, "reset_password")
		return
	}

	// Call password reset service
	serviceStart := time.Now()
	handlerLog.Debug("Calling password reset service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "ResetPassword",
	}))

	res, err := h.userClient.ResetPassword(ctx, &pb.ResetPasswordReq{
		Token:       req.Token,
		NewPassword: hashedPassword,
	})
	serviceDuration := time.Since(serviceStart)

	if err != nil {
		var failureReason string
		var clientError *errorcustom.APIError
		var httpStatus int

		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			failureReason = "invalid_or_expired_token"
			httpStatus = http.StatusBadRequest

			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeInvalidInput,
				"The password reset token is invalid or has expired",
				httpStatus,
				"handler",
				"reset_password",
				err,
			).WithDetail("token_status", "invalid_or_expired").
			  WithDetail("step", "token_verification")

			logger.WarningWithCause(
				"Password reset failed - invalid or expired token",
				failureReason,
				logger.LayerHandler,
				"reset_password",
				baseContext,
			)

			// Log security event for invalid token usage
			logger.LogSecurityEvent(
				"invalid_reset_token",
				"Attempt to use invalid or expired password reset token",
				"medium",
				utils.MergeContext(baseContext, map[string]interface{}{
					"token_status": "invalid_or_expired",
				}),
			)

		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError

			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"Password reset service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"reset_password",
				err,
			).WithDetail("step", "service_call").
			  WithDetail("retryable", true)

			logger.ErrorWithCause(
				"Password reset failed - service error",
				failureReason,
				logger.LayerExternal,
				"reset_password",
				utils.MergeContext(baseContext, map[string]interface{}{
					"grpc_error": err.Error(),
				}),
			)
		}

		// Log service call failure
		logger.LogServiceCall(
			"AccountService",
			"ResetPassword",
			false,
			err,
			utils.MergeContext(baseContext, map[string]interface{}{
				"duration_ms": serviceDuration.Milliseconds(),
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
			}),
		)

		handlerLog.Error("Password reset failed with detailed error", clientError.GetLogContext())

		errorcustom.HandleError(w, clientError, "reset_password")
		return
	}

	// Password reset successful
	logger.LogServiceCall(
		"AccountService",
		"ResetPassword",
		true,
		nil,
		utils.MergeContext(baseContext, map[string]interface{}{
			"duration_ms": serviceDuration.Milliseconds(),
		}),
	)

	// Log user activity (user_id not available from token context)
	logger.LogUserActivity(
		"unknown", // user_id not available from token
		"unknown", // email not available from response
		"reset",
		"password_reset",
		baseContext,
	)

	// Log security event for successful password reset
	logger.LogSecurityEvent(
		"password_reset_completed",
		"User successfully completed password reset",
		"low",
		utils.MergeContext(baseContext, map[string]interface{}{
			"action": "password_reset_successful",
		}),
	)

	// Log performance
	totalDuration := time.Since(start)
	logger.LogPerformance("reset_password", totalDuration, baseContext)

	// Log API request completion
	logger.LogAPIRequest(
		r.Method,
		r.URL.Path,
		http.StatusOK,
		totalDuration,
		utils.MergeContext(baseContext, map[string]interface{}{
			"password_reset_success": res.Success,
			"service_message":        res.Message,
		}),
	)

	handlerLog.Info("Password reset completed successfully", utils.MergeContext(baseContext, map[string]interface{}{
		"service_message": res.Message,
	}))

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "reset_password")
}
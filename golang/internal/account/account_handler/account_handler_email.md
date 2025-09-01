package account_handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	dto "english-ai-full/internal/account/account_dto"
	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)

// VerifyEmail handles email verification requests with comprehensive logging and error handling
func (h *AccountHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	  domain := errorcustom.GetDomainFromContext(r.Context())
    if domain == "" {
        domain = h.domain // fallback to struct field
    }
    h.WithRequestID(r) 
	start := time.Now()
	ctx := r.Context()

	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("verify_email")

	// Extract client information for security logging
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create base context with client information
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	handlerLog.Info("Email verification request initiated", baseContext)

	// Extract and validate token parameter
token, apiErr := errorcustom.GetStringParamWithDomain(r, "token", domain, 32)
	if apiErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": apiErr.Error(),
		})

		logger.ErrorWithCause(
			"Invalid token parameter",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_token",
			context,
		)

		// Log API request with error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)

			errorcustom.RespondWithError(w, http.StatusBadRequest, "Invalid token parameter", domain, h.requestID)
		return
	}

	// Add token info to context (masked for security)
	baseContext["token"] = utils.MaskSensitiveValue("token", token)
	
	handlerLog.Debug("Token parameter extracted successfully", baseContext)

	// Call verification service
	serviceStart := time.Now()
	handlerLog.Debug("Calling email verification service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "VerifyEmail",
	}))

	res, err := h.userClient.VerifyEmail(ctx, &pb.VerifyEmailReq{
		VerificationToken: token,
	})
	serviceDuration := time.Since(serviceStart)

	if err != nil {
		// Parse the gRPC error with enhanced context
	
		
		var failureReason string
		var clientError *errorcustom.APIError
		var httpStatus int

		// Determine specific failure reason
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			failureReason = "invalid_or_expired_token"
			httpStatus = http.StatusBadRequest

	clientError = errorcustom.NewAPIErrorWithContext(
	errorcustom.ErrCodeInvalidInput,
	"The verification token is invalid or has expired",
	httpStatus,
	domain, // Add domain parameter
	"handler",
	"verify_email",
	err,
).WithDetail("token_status", "invalid_or_expired").
  WithDetail("step", "token_verification")

			logger.WarningWithCause(
				"Email verification failed - invalid or expired token",
				failureReason,
				logger.LayerHandler,
				"verify_email",
				baseContext,
			)

			// Log security event for invalid token usage
			logger.LogSecurityEvent(
				"invalid_verification_token",
				"Attempt to use invalid or expired verification token",
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
				"Email verification service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"verify_email",
					domain, // Add domain parameter
				err,
			).WithDetail("step", "service_call").
			  WithDetail("retryable", true)

			logger.ErrorWithCause(
				"Email verification failed - service error",
				failureReason,
				logger.LayerExternal,
				"verify_email",
				utils.MergeContext(baseContext, map[string]interface{}{
					"grpc_error": err.Error(),
				}),
			)
		}

		// Log service call failure
		logger.LogServiceCall(
			"AccountService",
			"VerifyEmail",
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

		handlerLog.Error("Email verification failed with detailed error", clientError.GetLogContext())

		errorcustom.HandleError(w, clientError, "verify_email")
		return
	}

	// Verification successful
	logger.LogServiceCall(
		"AccountService",
		"VerifyEmail",
		true,
		nil,
		utils.MergeContext(baseContext, map[string]interface{}{
			"duration_ms": serviceDuration.Milliseconds(),
		}),
	)

	// Log user activity for successful verification
	logger.LogUserActivity(
		"unknown", // user_id not available in this context
		"unknown", // email not available in response
		"verify",
		"email_verification",
		baseContext,
	)

	// Log performance
	totalDuration := time.Since(start)
	logger.LogPerformance("verify_email", totalDuration, baseContext)

	// Log API request completion
	logger.LogAPIRequest(
		r.Method,
		r.URL.Path,
		http.StatusOK,
		totalDuration,
		utils.MergeContext(baseContext, map[string]interface{}{
			"verification_success": res.Success,
			"service_message":      res.Message,
		}),
	)

	handlerLog.Info("Email verification completed successfully", utils.MergeContext(baseContext, map[string]interface{}{
		"service_message": res.Message,
	}))

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "verify_email")
}

// ResendVerification handles resend verification email requests with comprehensive logging
func (h *AccountHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("resend_verification")

	// Extract client information
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create base context
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	handlerLog.Info("Resend verification request initiated", baseContext)

	// Decode request body
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := errorcustom.DecodeJSON(r.Body, &req, "resend_verification", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})

		logger.ErrorWithCause(
			"Failed to decode resend verification request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)

		errorcustom.HandleError(w, err, "resend_verification")
		return
	}

	// Add email to context (masked for privacy)
	baseContext["email"] = utils.MaskSensitiveValue("email", req.Email)

	handlerLog.Info("Resend verification attempt for user", baseContext)

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
				"Resend verification request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
requestID := errorcustom.GetRequestIDFromContext(ctx) 
			errorcustom.HandleValidationErrors(w, validationErrors, h.domain, requestID)
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during resend verification",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)

			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "resend_verification")
		}

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Request validation passed", baseContext)

	// Call resend verification service
	serviceStart := time.Now()
	handlerLog.Debug("Calling resend verification service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "ResendVerification",
	}))

	res, err := h.userClient.ResendVerification(ctx, &pb.ResendVerificationReq{
		Email: req.Email,
	})
	serviceDuration := time.Since(serviceStart)

	if err != nil {
		var failureReason string
		var clientError *errorcustom.APIError
		var httpStatus int

		if strings.Contains(err.Error(), "user not found") {
			// For security, we don't reveal if user exists or not
			failureReason = "user_not_found"
			httpStatus = http.StatusOK // Return 200 for security

			logger.WarningWithCause(
				"Resend verification - user not found",
				failureReason,
				logger.LayerHandler,
				"resend_verification",
				baseContext,
			)

			// Log security event for email enumeration attempt
			logger.LogSecurityEvent(
				"email_enumeration_attempt",
				"Attempt to resend verification for non-existent email",
				"low",
				baseContext,
			)

			// Log service call
			logger.LogServiceCall(
				"AccountService",
				"ResendVerification",
				false,
				err,
				utils.MergeContext(baseContext, map[string]interface{}{
					"duration_ms": serviceDuration.Milliseconds(),
				}),
			)

			// Log API request - return generic success message for security
			logger.LogAPIRequest(r.Method, r.URL.Path, httpStatus, time.Since(start), utils.MergeContext(baseContext, map[string]interface{}{
				"failure_reason": failureReason,
				"security_response": true,
			}))

			errorcustom.RespondWithJSON(w, http.StatusOK, map[string]string{
				"message": "If the email exists and is unverified, a verification email has been sent",
			}, "resend_verification")
			return
		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError
	  // new asdfasdfasdf


	clientError = errorcustom.NewAPIErrorWithContext(
    errorcustom.ErrCodeServiceError,
    "Email verification service is temporarily unavailable. Please try again later.",
    httpStatus,
    "user",                    // domain parameter (was missing)
    "handler",                 // layer parameter  
    "resend_verification",     // operation parameter
    err,                       // cause parameter
).WithDetail("email", utils.MaskSensitiveValue("email", req.Email)).
  WithDetail("step", "service_call").
  WithDetail("retryable", true)


			  // new asdfasdfasdf
			logger.ErrorWithCause(
				"Resend verification failed - service error",
				failureReason,
				logger.LayerExternal,
				"resend_verification",
				utils.MergeContext(baseContext, map[string]interface{}{
					"grpc_error": err.Error(),
				}),
			)
		}

		// Log service call failure
		logger.LogServiceCall(
			"AccountService",
			"ResendVerification",
			false,
			err,
			utils.MergeContext(baseContext, map[string]interface{}{
				"duration_ms": serviceDuration.Milliseconds(),
			}),
		)

		if clientError != nil {
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

			handlerLog.Error("Resend verification failed with detailed error", clientError.GetLogContext())

			errorcustom.HandleError(w, clientError, "resend_verification")
		}
		return
	}

	// Resend verification successful
	logger.LogServiceCall(
		"AccountService",
		"ResendVerification",
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
		"resend",
		"email_verification",
		baseContext,
	)

	// Log performance
	totalDuration := time.Since(start)
	logger.LogPerformance("resend_verification", totalDuration, baseContext)

	// Log API request completion
	logger.LogAPIRequest(
		r.Method,
		r.URL.Path,
		http.StatusOK,
		totalDuration,
		utils.MergeContext(baseContext, map[string]interface{}{
			"resend_success": res.Success,
			"service_message": res.Message,
		}),
	)

	handlerLog.Info("Resend verification completed successfully", utils.MergeContext(baseContext, map[string]interface{}{
		"service_message": res.Message,
	}))

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "resend_verification")
}

// FindByEmail handles find user by email requests with comprehensive logging
func (h *AccountHandler) FindByEmail(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("find_by_email")

	// Extract client information
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create base context
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})

	handlerLog.Info("Find user by email request initiated", baseContext)

	// Extract and validate email parameter
	email, apiErr := errorcustom.GetStringParamWithDomain(r, "email", h.domain, 320)
	if apiErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": apiErr.Error(),
		})

		logger.ErrorWithCause(
			"Invalid email parameter",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_email",
			context,
		)

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)

		errorcustom.HandleDomainError(w, apiErr, h.domain, h.requestID)
		return
	}

	// Add email to context (masked for privacy)
	baseContext["email"] = utils.MaskSensitiveValue("email", email)

	handlerLog.Info("Find user by email attempt", baseContext)

	// Validate email format
	var req struct {
		Email string `validate:"required,email"`
	}
	req.Email = email

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
				"Email format validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_email_format",
				context,
			)

			errorcustom.HandleValidationErrors(w, validationErrors, "find_by_email", h.domain)
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during email validation",
				"validation_system_error",
				logger.LayerHandler,
				"validate_email_format",
				context,
			)

			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Invalid email format",
				http.StatusBadRequest,
			), "find_by_email")
		}

		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Email format validation passed", baseContext)

	// Call find by email service
	serviceStart := time.Now()
	handlerLog.Debug("Calling find by email service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "FindByEmail",
	}))

	res, err := h.userClient.FindByEmail(ctx, &pb.FindByEmailReq{
		Email: email,
	})
	serviceDuration := time.Since(serviceStart)

	if err != nil {
		var failureReason string
		var clientError *errorcustom.APIError
		var httpStatus int

		if strings.Contains(err.Error(), "not found") {
			failureReason = "user_not_found"
			httpStatus = http.StatusNotFound

			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeNotFound,
				"No user found with the provided email address",
				httpStatus,
				"handler",
				"find_by_email",
				h.domain,
				err,
			
			).WithDetail("email", utils.MaskSensitiveValue("email", email)).
			  WithDetail("step", "user_lookup")

			logger.WarningWithCause(
				"User not found with provided email",
				failureReason,
				logger.LayerHandler,
				"find_by_email",
				baseContext,
			)

		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError

			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"User lookup service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"find_by_email",
				h.domain,
				err,
			).WithDetail("email", utils.MaskSensitiveValue("email", email)).
			  WithDetail("step", "service_call").
			  WithDetail("retryable", true)

			logger.ErrorWithCause(
				"Find by email failed - service error",
				failureReason,
				logger.LayerExternal,
				"find_by_email",
				utils.MergeContext(baseContext, map[string]interface{}{
					"grpc_error": err.Error(),
				}),
			)
		}

		// Log service call failure
		logger.LogServiceCall(
			"AccountService",
			"FindByEmail",
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

		handlerLog.Error("Find by email failed with detailed error", clientError.GetLogContext())

		errorcustom.HandleError(w, clientError, "find_by_email")
		return
	}

	// Find by email successful
	userContext := utils.MergeContext(baseContext, map[string]interface{}{
		"user_id":   res.Account.Id,
		"branch_id": res.Account.BranchId,
		"role":      res.Account.Role,
	})

	logger.LogServiceCall(
		"AccountService",
		"FindByEmail",
		true,
		nil,
		utils.MergeContext(userContext, map[string]interface{}{
			"duration_ms": serviceDuration.Milliseconds(),
		}),
	)

	// Log user activity
	logger.LogUserActivity(
		fmt.Sprint(res.Account.Id),
		res.Account.Email,
		"lookup",
		"user_profile",
		userContext,
	)

	// Prepare response
	response := dto.FindAccountByIDResponse{
		ID:        res.Account.Id,
		BranchID:  res.Account.BranchId,
		Name:      res.Account.Name,
		Email:     res.Account.Email,
		Avatar:    res.Account.Avatar,
		Title:     res.Account.Title,
		Role:      res.Account.Role,
		OwnerID:   res.Account.OwnerId,
		CreatedAt: res.Account.CreatedAt.AsTime(),
		UpdatedAt: res.Account.UpdatedAt.AsTime(),
	}

	// Log performance
	totalDuration := time.Since(start)
	logger.LogPerformance("find_by_email", totalDuration, userContext)

	// Log API request completion
	logger.LogAPIRequest(
		r.Method,
		r.URL.Path,
		http.StatusOK,
		totalDuration,
		utils.MergeContext(userContext, map[string]interface{}{
			"user_found": true,
		}),
	)

	handlerLog.Info("Find by email completed successfully", utils.MergeContext(userContext, map[string]interface{}{
		"user_name": res.Account.Name,
	}))

	errorcustom.RespondWithJSON(w, http.StatusOK, response, "find_by_email")
}
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

// RefreshToken handles token refresh requests with comprehensive logging
func (h *AccountHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("refresh_token")
	
	// Extract client information for security logging
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})
	
	handlerLog.Info("Token refresh request initiated", baseContext)

	// Decode request body
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := errorcustom.DecodeJSON(r.Body, &req, "refresh_token", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to decode refresh token request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)
		
		// Log API request with decode error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.HandleError(w, err, "refresh_token")
		return
	}

	// Basic token info for logging (masked for security)
	tokenContext := utils.MergeContext(baseContext, map[string]interface{}{
		"token_length": len(req.RefreshToken),
		"token_prefix": maskToken(req.RefreshToken, 6),
	})
	
	handlerLog.Debug("Refresh token request decoded", tokenContext)

	// Validate request structure
	if err := h.validator.Struct(&req); err != nil {
		context := utils.MergeContext(tokenContext, map[string]interface{}{
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
				"Refresh token request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			// Log API request with validation error
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleValidationErrors(w, validationErrors, "refresh_token")
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during token refresh",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			clientError := errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
				"handler",
				"refresh_token",
				err,
			)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleError(w, clientError, "refresh_token")
		}
		return
	}

	handlerLog.Debug("Request validation passed", tokenContext)

	// Call token refresh service
	serviceStart := time.Now()
	handlerLog.Debug("Calling token refresh service", utils.MergeContext(tokenContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "RefreshToken",
	}))

	res, err := h.userClient.RefreshToken(ctx, &pb.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	serviceDuration := time.Since(serviceStart)

	// Log service call
	serviceContext := utils.MergeContext(tokenContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "RefreshToken",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		var httpStatus int
		var clientError *errorcustom.APIError
		var failureReason string
		
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			failureReason = "invalid_or_expired_token"
			httpStatus = http.StatusUnauthorized
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeInvalidToken,
				"The refresh token is invalid or has expired. Please log in again.",
				httpStatus,
				"handler",
				"refresh_token",
				err,
			).WithDetail("token_type", "refresh").
			  WithDetail("step", "token_validation").
			  WithDetail("suggestion", "Please log in again to obtain new tokens")
			
			// Log security event for invalid token usage
			logger.LogSecurityEvent(
				"invalid_refresh_token_usage",
				"Attempt to use invalid or expired refresh token",
				"medium",
				tokenContext,
			)
			
			logger.WarningWithCause(
				"Token refresh failed - invalid or expired token",
				failureReason,
				logger.LayerHandler,
				"refresh_token",
				tokenContext,
			)
		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"Token refresh service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"refresh_token",
				err,
			).WithDetail("step", "service_call").
			  WithDetail("retryable", true)
			
			logger.ErrorWithCause(
				"Token refresh service call failed",
				failureReason,
				logger.LayerExternal,
				"refresh_token",
				utils.MergeContext(tokenContext, map[string]interface{}{
					"error": err.Error(),
				}),
			)
		}
		
		// Log service call failure
		logger.LogServiceCall("user-service", "RefreshToken", false, err, utils.MergeContext(serviceContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		// Log API request completion
		logger.LogAPIRequest(r.Method, r.URL.Path, httpStatus, time.Since(start), utils.MergeContext(tokenContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		errorcustom.HandleError(w, clientError, "refresh_token")
		return
	}

	// Token refresh successful
	logger.LogServiceCall("user-service", "RefreshToken", true, nil, serviceContext)

	// Add response info to context (masked tokens)
	responseContext := utils.MergeContext(tokenContext, map[string]interface{}{
		"new_access_token_length":  len(res.AccessToken),
		"new_refresh_token_length": len(res.RefreshToken),
		"expires_at":               res.ExpiresAt,
	})

	// Log performance
	logger.LogPerformance("refresh_token", time.Since(start), responseContext)

	handlerLog.Info("Token refresh completed successfully", responseContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), responseContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_at":    res.ExpiresAt,
	}, "refresh_token")
}

// ValidateToken handles token validation requests with comprehensive logging
func (h *AccountHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("validate_token")
	
	// Extract client information for security logging
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})
	
	handlerLog.Info("Token validation request initiated", baseContext)

	// Extract authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"missing_header": "Authorization",
		})
		
		logger.WarningWithCause(
			"Token validation failed - missing authorization header",
			"missing_auth_header",
			logger.LayerHandler,
			"extract_auth_header",
			context,
		)
		
		// Log security event for missing auth header
		logger.LogSecurityEvent(
			"missing_authorization_header",
			"Request to validate token without Authorization header",
			"low",
			context,
		)
		
		clientError := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInvalidInput,
			"Missing authorization header",
			http.StatusUnauthorized,
			"handler",
			"validate_token",
			nil,
		).WithDetail("expected_header", "Authorization").
		  WithDetail("header_format", "Bearer <token>")
		
		// Log API request with missing header error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusUnauthorized, time.Since(start), context)
		
		errorcustom.HandleError(w, clientError, "validate_token")
		return
	}

	// Extract token from Bearer header
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"invalid_header_format": authHeader,
		})
		
		logger.WarningWithCause(
			"Token validation failed - invalid authorization header format",
			"invalid_auth_header_format",
			logger.LayerHandler,
			"parse_auth_header",
			context,
		)
		
		// Log security event for malformed auth header
		logger.LogSecurityEvent(
			"malformed_authorization_header",
			"Request with malformed Authorization header",
			"low",
			context,
		)
		
		clientError := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInvalidInput,
			"Invalid authorization header format. Expected: Bearer <token>",
			http.StatusUnauthorized,
			"handler",
			"validate_token",
			nil,
		).WithDetail("provided_format", authHeader).
		  WithDetail("expected_format", "Bearer <token>")
		
		// Log API request with invalid format error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusUnauthorized, time.Since(start), context)
		
		errorcustom.HandleError(w, clientError, "validate_token")
		return
	}

	// Create token context (masked for security)
	tokenContext := utils.MergeContext(baseContext, map[string]interface{}{
		"token_length": len(token),
		"token_prefix": maskToken(token, 6),
	})
	
	handlerLog.Debug("Token extracted from Authorization header", tokenContext)

	// Call token validation service
	serviceStart := time.Now()
	handlerLog.Debug("Calling token validation service", utils.MergeContext(tokenContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "ValidateToken",
	}))

	res, err := h.userClient.ValidateToken(ctx, &pb.ValidateTokenReq{
		Token: token,
	})
	serviceDuration := time.Since(serviceStart)

	// Log service call
	serviceContext := utils.MergeContext(tokenContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "ValidateToken",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		var httpStatus int
		var clientError *errorcustom.APIError
		var failureReason string
		
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			failureReason = "invalid_or_expired_token"
			httpStatus = http.StatusUnauthorized
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeInvalidToken,
				"The access token is invalid or has expired",
				httpStatus,
				"handler",
				"validate_token",
				err,
			).WithDetail("token_type", "access").
			  WithDetail("step", "token_validation").
			  WithDetail("suggestion", "Please obtain a new access token using your refresh token or log in again")
			
			// Log security event for invalid token usage
			logger.LogSecurityEvent(
				"invalid_access_token_usage",
				"Attempt to use invalid or expired access token",
				"medium",
				tokenContext,
			)
			
			logger.WarningWithCause(
				"Token validation failed - invalid or expired token",
				failureReason,
				logger.LayerHandler,
				"validate_token",
				tokenContext,
			)
		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"Token validation service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"validate_token",
				err,
			).WithDetail("step", "service_call").
			  WithDetail("retryable", true)
			
			logger.ErrorWithCause(
				"Token validation service call failed",
				failureReason,
				logger.LayerExternal,
				"validate_token",
				utils.MergeContext(tokenContext, map[string]interface{}{
					"error": err.Error(),
				}),
			)
		}
		
		// Log service call failure
		logger.LogServiceCall("user-service", "ValidateToken", false, err, utils.MergeContext(serviceContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		// Log API request completion
		logger.LogAPIRequest(r.Method, r.URL.Path, httpStatus, time.Since(start), utils.MergeContext(tokenContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		errorcustom.HandleError(w, clientError, "validate_token")
		return
	}

	// Token validation successful
	logger.LogServiceCall("user-service", "ValidateToken", true, nil, serviceContext)

	// Add validation result to context
	validationContext := utils.MergeContext(tokenContext, map[string]interface{}{
		"token_valid": res.Valid,
		"user_id":     res.UserId,
		"expires_at":  res.ExpiresAt,
	})

	if res.Valid {
		handlerLog.Info("Token validation successful", validationContext)
		
		// Log user activity for successful token validation
		logger.LogUserActivity(
			fmt.Sprint(res.UserId),
			"", // email not available in this context
			"validate",
			"access_token",
			validationContext,
		)
	} else {
		handlerLog.Warning("Token validation returned invalid result", validationContext)
		
		// Log security event for invalid token
		logger.LogSecurityEvent(
			"token_validation_failed",
			"Token validation service returned invalid token result",
			"medium",
			validationContext,
		)
	}

	// Log performance
	logger.LogPerformance("validate_token", time.Since(start), validationContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), validationContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"valid":      res.Valid,
		"expires_at": res.ExpiresAt,
		"message":    res.Message,
		"id":         res.UserId,
	}, "validate_token")
}

// Helper function to mask tokens for secure logging
func maskToken(token string, visibleChars int) string {
	if len(token) <= visibleChars {
		return strings.Repeat("*", len(token))
	}
	
	visible := token[:visibleChars]
	masked := strings.Repeat("*", len(token)-visibleChars)
	return visible + masked
}
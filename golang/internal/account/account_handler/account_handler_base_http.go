package account_handler

import (
	"fmt"
	"net/http"

	"time"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/logger"
	"english-ai-full/utils"
)

// ============================================================================
// HTTP REQUEST/RESPONSE HANDLING
// ============================================================================

// HandleHTTPError handles HTTP errors using the unified error handler
func (h *BaseAccountHandler) HandleHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	h.errorHandler.HandleHTTPError(w, r, err)
}

// RespondWithSuccess sends successful response with domain context
func (h *BaseAccountHandler) RespondWithSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	errorcustom.RespondWithDomainSuccess(w, data, h.domain, requestID)
}

// RespondWithError sends error response with domain context
func (h *BaseAccountHandler) RespondWithError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	// Process error through domain handler
	processedErr := h.errorHandler.HandleError(h.domain, err)
	
	// Handle with domain context
	errorcustom.HandleDomainError(w, processedErr, h.domain, requestID)
}

// DecodeJSONRequest decodes JSON request using unified error handler
func (h *BaseAccountHandler) DecodeJSONRequest(r *http.Request, target interface{}) error {
	return h.errorHandler.DecodeJSONRequest(r, target)
}

// ============================================================================
// PARAMETER PARSING
// ============================================================================

// ParseIDParam safely parses ID parameters with domain validation
func (h *BaseAccountHandler) ParseIDParam(r *http.Request, paramName string) (int64, error) {
	return h.errorHandler.ParseIDParam(r, paramName)
}

// parseStringParam safely parses string parameters with domain validation
func (h *BaseAccountHandler) parseStringParam(r *http.Request, paramName string, minLen int) (string, error) {
	return errorcustom.GetStringParamWithDomain(r, paramName, h.domain, minLen)
}

// getPaginationParams extracts pagination parameters
func (h *BaseAccountHandler) getPaginationParams(r *http.Request) (page, pageSize int32, err error) {
	limit, offset, err := h.errorHandler.ParsePaginationParams(r)
	if err != nil {
		return 0, 0, err
	}
	
	// Handle edge case: avoid division by zero
	if limit == 0 {
		limit = 10 // default page size
	}
	
	// Convert offset-based to page-based pagination
	page = int32((offset / limit) + 1)
	pageSize = int32(limit)
	
	// Log pagination parameters
	if h.logger != nil {
		requestID := errorcustom.GetRequestIDFromContext(r.Context())
		h.logger.Debug("Pagination parameters processed", map[string]interface{}{
			"page":       page,
			"page_size":  pageSize,
			"limit":      limit,
			"offset":     offset,
			"domain":     h.domain,
			"request_id": requestID,
		})
	}
	
	return page, pageSize, nil
}

// getSortingParams extracts sorting parameters with domain-aware validation
func (h *BaseAccountHandler) getSortingParams(r *http.Request, allowedFields []string) (sortBy, sortOrder string, err error) {
	return h.errorHandler.GetSortParamsWithDomain(r, allowedFields, h.domain)
}

// ============================================================================
// REQUEST LOGGING AND MIDDLEWARE
// ============================================================================

// withRequestLogging wraps handler functions with comprehensive request logging
func (h *BaseAccountHandler) withRequestLogging(operation string, handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opCtx := h.setupOperationContext(r, operation)
		opCtx.Context["method"] = r.Method
		opCtx.Context["path"] = r.URL.Path
		
		h.logOperationStart(opCtx)
		
		// Create response writer wrapper to capture status code
		responseWriter := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Execute handler
		handler(responseWriter, r)
		
		// Log operation end
		h.logOperationEnd(opCtx, nil, responseWriter.statusCode)
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture status codes
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// logHandlerStart logs the beginning of a handler operation with domain context
func (h *BaseAccountHandler) logHandlerStart(r *http.Request, operation string, additionalContext map[string]interface{}) map[string]interface{} {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	clientIP := errorcustom.GetClientIP(r)
	
	baseContext := utils.CreateBaseContext(r, utils.MergeContext(additionalContext, map[string]interface{}{
		"operation":          operation,
		"component":          "account_handler",
		"domain":             h.domain,
		"handler_start_time": time.Now(),
		"request_id":         requestID,
		"client_ip":          clientIP,
	}))
	
	logger.InfoWithOperation(
		"Domain handler operation started",
		logger.LayerHandler,
		operation,
		baseContext,
	)
	
	return baseContext
}

// logHandlerEnd logs the completion with enhanced domain context
func (h *BaseAccountHandler) logHandlerEnd(r *http.Request, operation string, statusCode int, startTime time.Time, additionalContext map[string]interface{}) {
	duration := time.Since(startTime)
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	context := utils.CreateBaseContext(r, utils.MergeContext(additionalContext, map[string]interface{}{
		"operation":   operation,
		"component":   "account_handler",
		"domain":      h.domain,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"request_id":  requestID,
	}))
	
	// Log API request completion with domain context
	logger.LogAPIRequest(r.Method, r.URL.Path, statusCode, duration, context)
	
	// Log performance metrics with domain tagging
	logger.LogPerformance(fmt.Sprintf("%s_%s", h.domain, operation), duration, context)
	
	// Enhanced status-based logging
	if statusCode >= 200 && statusCode < 300 {
		logger.InfoWithOperation(
			"Domain handler operation completed successfully",
			logger.LayerHandler,
			operation,
			context,
		)
	} else if statusCode >= 400 && statusCode < 500 {
		logger.WarningWithCause(
			"Domain handler operation completed with client error",
			"client_error",
			logger.LayerHandler,
			operation,
			context,
		)
	} else if statusCode >= 500 {
		logger.ErrorWithCause(
			"Domain handler operation completed with server error",
			"server_error",
			logger.LayerHandler,
			operation,
			context,
		)
		
		// Alert on server errors in production
		if h.config.IsProduction() {
			h.alertOpsTeam(operation, statusCode, context)
		}
	}
}

// ============================================================================
// SECURITY HELPERS
// ============================================================================

// logSecurityEvent logs security-related events with domain context
func (h *BaseAccountHandler) logSecurityEvent(eventType string, description string, severity string, context map[string]interface{}) {
	securityContext := utils.MergeContext(context, map[string]interface{}{
		"domain":    h.domain,
		"component": "account_handler",
		"timestamp": time.Now().UTC(),
	})
	
	logger.LogSecurityEvent(eventType, description, severity, securityContext)
}

// validateRequestOrigin validates the origin of the request for security
func (h *BaseAccountHandler) validateRequestOrigin(r *http.Request) error {
	origin := r.Header.Get("Origin")
	referer := r.Header.Get("Referer")
	
	// Skip validation for same-origin requests
	if origin == "" && referer == "" {
		return nil
	}
	
	allowedOrigins := h.config.GetAllowedOrigins()
	
	// Check origin against allowed list
	if origin != "" {
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return nil
			}
		}
		
		// Log suspicious origin
		h.logSecurityEvent(
			"suspicious_origin",
			"Request from unauthorized origin",
			"medium",
			map[string]interface{}{
				"origin":          origin,
				"referer":         referer,
				"allowed_origins": allowedOrigins,
			},
		)
		
		return errorcustom.NewSecurityError(
			h.domain,
			"unauthorized_origin",
			"Request origin not allowed",
		)
	}
	
	return nil
}

// ============================================================================
// ERROR RECOVERY AND RESILIENCE
// ============================================================================

// measureOperation measures operation duration and logs performance metrics
func (h *BaseAccountHandler) measureOperation(operation string, fn func() error) error {
	start := time.Now()
	
	h.logger.Debug("Starting measured operation", map[string]interface{}{
		"operation": operation,
		"domain":    h.domain,
	})
	
	err := fn()
	duration := time.Since(start)
	
	// Log performance metrics
	logger.LogPerformance(fmt.Sprintf("%s_%s", h.domain, operation), duration, map[string]interface{}{
		"operation":   operation,
		"domain":      h.domain,
		"success":     err == nil,
		"duration_ms": duration.Milliseconds(),
	})
	
	if err != nil {
		h.logger.Warning("Measured operation completed with error", map[string]interface{}{
			"operation":   operation,
			"domain":      h.domain,
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		})
	} else {
		h.logger.Debug("Measured operation completed successfully", map[string]interface{}{
			"operation":   operation,
			"domain":      h.domain,
			"duration_ms": duration.Milliseconds(),
		})
	}
	
	return err
}

// handleDomainError handles errors with full domain context and configuration
func (h *BaseAccountHandler) handleDomainError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := errorcustom.GetRequestIDFromContext(r.Context())
	
	// Process error through domain-aware error handler
	processedErr := h.errorHandler.HandleError(h.domain, err)
	
	// Handle the processed error with domain context
	errorcustom.HandleDomainError(w, processedErr, h.domain, requestID)
}

// alertOpsTeam sends alerts for critical errors in production
func (h *BaseAccountHandler) alertOpsTeam(operation string, statusCode int, context map[string]interface{}) {
	if !h.config.IsProduction() {
		return
	}
	
	alertContext := utils.MergeContext(context, map[string]interface{}{
		"alert_type":   "critical_error",
		"domain":       h.domain,
		"operation":    operation,
		"status_code":  statusCode,
		"timestamp":    time.Now().UTC(),
		"environment":  h.config.Environment,
	})
	
	logger.LogSecurityEvent(
		"critical_system_error",
		"Critical error in account handler",
		"high",
		alertContext,
	)
	
	// Here you would integrate with your alerting system
	// e.g., PagerDuty, Slack, email notifications
}

// ============================================================================
// COMPLETE HTTP HANDLER EXAMPLES
// ============================================================================


// HandleGetUser demonstrates a complete HTTP handler using the unified approach
func (h *BaseAccountHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	opCtx := h.setupOperationContext(r, "get_user")
	h.logOperationStart(opCtx)
	
	// Parse ID parameter using UnifiedErrorHandler
	userID, err := h.errorHandler.ParseIDParam(r, "id")
	if err != nil {
		h.logOperationEnd(opCtx, err, http.StatusBadRequest)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Get requesting user ID from context
	requestingUserID, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		h.logOperationEnd(opCtx, err, http.StatusUnauthorized)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Check permissions
	if err := h.checkUserPermissions(requestingUserID, "admin", "user_read"); err != nil {
		h.logOperationEnd(opCtx, err, http.StatusForbidden)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Get user data
	user, err := h.getUserByID(userID)
	if err != nil {
		h.logOperationEnd(opCtx, err, http.StatusInternalServerError)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Success response
	h.logOperationEnd(opCtx, nil, http.StatusOK)
	h.errorHandler.RespondWithSuccess(w, user)
}

// HandleCreateUser demonstrates a complete user creation handler
func (h *BaseAccountHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	opCtx := h.setupOperationContext(r, "create_user")
	h.logOperationStart(opCtx)
	
	var req CreateUserRequest
	
	// Decode and validate JSON request
	if err := h.validateAndDecodeRequest(r, &req, "create_user"); err != nil {
		h.logOperationEnd(opCtx, err, http.StatusBadRequest)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Validate business rules
	businessContext := map[string]interface{}{
		"email":    req.Email,
		"password": req.Password,
		"role":     req.Role,
	}
	
	if err := h.validateBusinessRules("user_registration", businessContext); err != nil {
		h.logOperationEnd(opCtx, err, http.StatusBadRequest)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Check email uniqueness
	if err := h.checkEmailUniqueness(r.Context(), req.Email); err != nil {
		h.logOperationEnd(opCtx, err, http.StatusConflict)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Create user via gRPC (this would be handled in the gRPC file)
	user, err := h.createUserViaGRPC(r.Context(), req)
	if err != nil {
		h.logOperationEnd(opCtx, err, http.StatusInternalServerError)
		h.errorHandler.HandleHTTPError(w, r, err)
		return
	}
	
	// Success response
	h.logOperationEnd(opCtx, nil, http.StatusCreated)
	h.errorHandler.RespondWithSuccess(w, user)
}

// ============================================================================
// RATE LIMITING AND THROTTLING
// ============================================================================

// applyRateLimit applies rate limiting based on configuration
func (h *BaseAccountHandler) applyRateLimit(w http.ResponseWriter, r *http.Request, operation string) error {
	if !h.isRateLimitEnabled() {
		return nil
	}
	
	clientIP := errorcustom.GetClientIP(r)
	
	// Check rate limit (this would integrate with your rate limiting system)
	if h.isRateLimitExceeded(clientIP, operation) {
		h.logSecurityEvent(
			"rate_limit_exceeded",
			"Client exceeded rate limit",
			"medium",
			map[string]interface{}{
				"client_ip": clientIP,
				"operation": operation,
			},
		)
		
		return errorcustom.NewRateLimitError(h.domain, operation)
	}
	
	return nil
}

// isRateLimitEnabled checks if rate limiting is enabled
func (h *BaseAccountHandler) isRateLimitEnabled() bool {
	if h.config != nil {
		return h.config.IsRateLimitEnabled()
	}
	return false
}

// isRateLimitExceeded checks if rate limit is exceeded (stub implementation)
func (h *BaseAccountHandler) isRateLimitExceeded(clientIP, operation string) bool {
	// This would integrate with your actual rate limiting system
	// e.g., Redis-based rate limiter, in-memory cache, etc.
	return false
}
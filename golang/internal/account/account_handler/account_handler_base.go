package account_handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/logger"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)

type BaseAccountHandler struct {
	userClient pb.AccountServiceClient
	validator  *validator.Validate
	logger     *logger.Logger // Base logger for this handler
}

// NewBaseHandler creates a new base account handler with comprehensive logging setup
func NewBaseHandler(userClient pb.AccountServiceClient) *BaseAccountHandler {
	// Create handler-specific logger
	handlerLogger := logger.NewHandlerLogger()
	handlerLogger.AddGlobalField("component", "account_handler")
	handlerLogger.AddGlobalField("layer", "handler")
	
	// Initialize validator with custom validations
	v := validator.New()
	
	// Register custom validation functions with logging
	logger.Debug("Registering custom validation functions", map[string]interface{}{
		"validations": []string{"password", "role", "uniqueemail"},
		"component":   "account_handler",
	})
	
	v.RegisterValidation("password", ValidatePassword)
	v.RegisterValidation("role", ValidateRole)
	v.RegisterValidation("uniqueemail", ValidateEmailUnique(userClient))

	handler := &BaseAccountHandler{
		userClient: userClient,
		validator:  v,
		logger:     handlerLogger,
	}

	logger.Info("Base account handler initialized successfully", map[string]interface{}{
		"component":          "account_handler",
		"validator_setup":    true,
		"custom_validations": 3,
		"grpc_client_ready":  userClient != nil,
	})

	return handler
}

// getUserIDFromContext extracts user ID from request context with detailed logging
func (h *BaseAccountHandler) getUserIDFromContext(ctx context.Context) (int64, error) {
	start := time.Now()
	
	// Create context for logging
	logContext := map[string]interface{}{
		"operation": "get_user_id_from_context",
		"component": "account_handler",
	}
	
	h.logger.Debug("Extracting user ID from context", logContext)
	
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		// Check for alternative context keys that might be used
		if userIDStr, exists := ctx.Value("user_id").(string); exists {
			if parsedID, parseErr := strconv.ParseInt(userIDStr, 10, 64); parseErr == nil {
				logger.Warning("User ID found as string in context, converting to int64", utils.MergeContext(logContext, map[string]interface{}{
					"user_id_string": userIDStr,
					"user_id_int64":  parsedID,
					"duration_ms":    time.Since(start).Milliseconds(),
				}))
				
				return parsedID, nil
			}
		}
		
		// Log the failed context extraction with available context keys
		contextKeys := h.getContextKeys(ctx)
		
		logger.ErrorWithCause(
			"User ID not found in request context",
			"missing_user_context",
			logger.LayerHandler,
			"extract_user_id",
			utils.MergeContext(logContext, map[string]interface{}{
				"available_context_keys": contextKeys,
				"expected_key":           "user_id",
				"expected_type":          "int64",
				"duration_ms":           time.Since(start).Milliseconds(),
			}),
		)
		
		// Log security event for unauthorized access attempt
		logger.LogSecurityEvent(
			"unauthorized_context_access",
			"Request made without valid user context",
			"medium",
			utils.MergeContext(logContext, map[string]interface{}{
				"missing_context": "user_id",
				"auth_required":   true,
			}),
		)
		
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"User ID not found in context",
			http.StatusUnauthorized,
		)
	}
	
	// Log successful context extraction
	h.logger.Debug("User ID successfully extracted from context", utils.MergeContext(logContext, map[string]interface{}{
		"user_id":     userID,
		"duration_ms": time.Since(start).Milliseconds(),
	}))
	
	return userID, nil
}

// getPaginationParams extracts and validates pagination parameters with comprehensive logging
func (h *BaseAccountHandler) getPaginationParams(r *http.Request) (page, pageSize int32, apiErr *errorcustom.APIError) {
	start := time.Now()
	
	// Create base context for logging
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"operation": "get_pagination_params",
		"component": "account_handler",
	})
	
	// Extract query parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	
	// Log pagination request
	h.logger.Debug("Processing pagination parameters", utils.MergeContext(baseContext, map[string]interface{}{
		"page_param":      pageStr,
		"page_size_param": pageSizeStr,
	}))
	
	// Set defaults
	page, pageSize = 1, 10
	defaultsApplied := map[string]interface{}{
		"default_page":      page,
		"default_page_size": pageSize,
	}
	
	// Validate and parse page parameter
	if pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 32); err != nil {
			context := utils.MergeContext(baseContext, map[string]interface{}{
				"page_param":    pageStr,
				"parse_error":   err.Error(),
				"expected_type": "positive integer",
			})
			
			logger.LogValidationError("page", "invalid_format", pageStr)
			
			logger.ErrorWithCause(
				"Invalid page parameter format",
				"parameter_parsing_error",
				logger.LayerHandler,
				"validate_pagination",
				context,
			)
			
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid page parameter: must be a positive integer",
				http.StatusBadRequest,
			)
		} else if p < 1 {
			context := utils.MergeContext(baseContext, map[string]interface{}{
				"page_param":   pageStr,
				"parsed_value": p,
				"min_allowed":  1,
			})
			
			logger.LogValidationError("page", "min_value", p)
			
			logger.WarningWithCause(
				"Page parameter below minimum value",
				"parameter_range_error",
				logger.LayerHandler,
				"validate_pagination",
				context,
			)
			
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid page parameter: must be a positive integer",
				http.StatusBadRequest,
			)
		} else {
			page = int32(p)
			h.logger.Debug("Page parameter parsed successfully", utils.MergeContext(baseContext, map[string]interface{}{
				"page_param": pageStr,
				"page_value": page,
			}))
		}
	} else {
		h.logger.Debug("Using default page value", utils.MergeContext(baseContext, defaultsApplied))
	}
	
	// Validate and parse page_size parameter
	if pageSizeStr != "" {
		if ps, err := strconv.ParseInt(pageSizeStr, 10, 32); err != nil {
			context := utils.MergeContext(baseContext, map[string]interface{}{
				"page_size_param": pageSizeStr,
				"parse_error":     err.Error(),
				"expected_type":   "integer between 1 and 100",
			})
			
			logger.LogValidationError("page_size", "invalid_format", pageSizeStr)
			
			logger.ErrorWithCause(
				"Invalid page_size parameter format",
				"parameter_parsing_error",
				logger.LayerHandler,
				"validate_pagination",
				context,
			)
			
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid page_size parameter: must be between 1 and 100",
				http.StatusBadRequest,
			)
		} else if ps < 1 || ps > 100 {
			context := utils.MergeContext(baseContext, map[string]interface{}{
				"page_size_param": pageSizeStr,
				"parsed_value":    ps,
				"min_allowed":     1,
				"max_allowed":     100,
			})
			
			logger.LogValidationError("page_size", "range_violation", ps)
			
			logger.WarningWithCause(
				"Page size parameter outside allowed range",
				"parameter_range_error",
				logger.LayerHandler,
				"validate_pagination",
				context,
			)
			
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid page_size parameter: must be between 1 and 100",
				http.StatusBadRequest,
			)
		} else {
			pageSize = int32(ps)
			h.logger.Debug("Page size parameter parsed successfully", utils.MergeContext(baseContext, map[string]interface{}{
				"page_size_param": pageSizeStr,
				"page_size_value": pageSize,
			}))
		}
	} else {
		h.logger.Debug("Using default page size value", utils.MergeContext(baseContext, defaultsApplied))
	}
	
	// Log final pagination parameters
	finalContext := utils.MergeContext(baseContext, map[string]interface{}{
		"final_page":      page,
		"final_page_size": pageSize,
		"duration_ms":     time.Since(start).Milliseconds(),
	})
	
	h.logger.Info("Pagination parameters processed successfully", finalContext)
	
	// Log performance metrics for pagination parsing
	logger.LogPerformance("pagination_parsing", time.Since(start), finalContext)
	
	return page, pageSize, nil
}

// validateRequest performs comprehensive request validation with detailed logging
func (h *BaseAccountHandler) validateRequest(request interface{}, operation string, baseContext map[string]interface{}) error {
	start := time.Now()
	
	context := utils.MergeContext(baseContext, map[string]interface{}{
		"operation":     "validate_request",
		"target_operation": operation,
		"request_type":  utils.GetTypeName(request),
	})
	
	h.logger.Debug("Starting request validation", context)
	
	if err := h.validator.Struct(request); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Log each validation error individually
			errorDetails := make([]map[string]interface{}, 0, len(validationErrors))
			
			for _, validationError := range validationErrors {
				errorDetail := map[string]interface{}{
					"field":      validationError.Field(),
					"tag":        validationError.Tag(),
					"value":      utils.MaskSensitiveValue(validationError.Field(), validationError.Value()),
					"param":      validationError.Param(),
					"namespace":  validationError.Namespace(),
				}
				errorDetails = append(errorDetails, errorDetail)
				
				// Log individual validation error
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					validationError.Value(),
				)
			}
			
			// Log comprehensive validation failure
			logger.WarningWithCause(
				"Request validation failed with multiple errors",
				"validation_failed",
				logger.LayerHandler,
				"validate_request",
				utils.MergeContext(context, map[string]interface{}{
					"validation_errors": errorDetails,
					"error_count":       len(validationErrors),
					"duration_ms":       time.Since(start).Milliseconds(),
				}),
			)
			
			return err
		} else {
			// Unexpected validation error
			logger.ErrorWithCause(
				"Unexpected validation system error",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				utils.MergeContext(context, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(start).Milliseconds(),
				}),
			)
			
			return err
		}
	}
	
	// Log successful validation
	h.logger.Debug("Request validation completed successfully", utils.MergeContext(context, map[string]interface{}{
		"duration_ms": time.Since(start).Milliseconds(),
	}))
	
	// Log performance metrics
	logger.LogPerformance("request_validation", time.Since(start), context)
	
	return nil
}

// logHandlerStart logs the beginning of a handler operation
func (h *BaseAccountHandler) logHandlerStart(r *http.Request, operation string, additionalContext map[string]interface{}) map[string]interface{} {
	baseContext := utils.CreateBaseContext(r, utils.MergeContext(additionalContext, map[string]interface{}{
		"operation": operation,
		"component": "account_handler",
		"handler_start_time": time.Now(),
	}))
	
	logger.InfoWithOperation(
		"Handler operation started",
		logger.LayerHandler,
		operation,
		baseContext,
	)
	
	return baseContext
}

// logHandlerEnd logs the completion of a handler operation
func (h *BaseAccountHandler) logHandlerEnd(r *http.Request, operation string, statusCode int, startTime time.Time, additionalContext map[string]interface{}) {
	duration := time.Since(startTime)
	
	context := utils.CreateBaseContext(r, utils.MergeContext(additionalContext, map[string]interface{}{
		"operation":   operation,
		"component":   "account_handler",
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
	}))
	
	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, statusCode, duration, context)
	
	// Log performance metrics
	logger.LogPerformance(operation, duration, context)
	
	// Log handler completion
	if statusCode >= 200 && statusCode < 300 {
		logger.InfoWithOperation(
			"Handler operation completed successfully",
			logger.LayerHandler,
			operation,
			context,
		)
	} else if statusCode >= 400 && statusCode < 500 {
		logger.WarningWithCause(
			"Handler operation completed with client error",
			"client_error",
			logger.LayerHandler,
			operation,
			context,
		)
	} else if statusCode >= 500 {
		logger.ErrorWithCause(
			"Handler operation completed with server error",
			"server_error",
			logger.LayerHandler,
			operation,
			context,
		)
	}
}

// getContextKeys extracts available context keys for debugging purposes
func (h *BaseAccountHandler) getContextKeys(ctx context.Context) []string {
	// This is a helper function to extract context keys for debugging
	// In a real implementation, you might have a context wrapper that tracks keys
	keys := []string{}
	
	// Check for common context keys
	commonKeys := []string{"user_id", "user_email", "request_id", "trace_id", "span_id", "tenant_id"}
	for _, key := range commonKeys {
		if value := ctx.Value(key); value != nil {
			keys = append(keys, key)
		}
	}
	
	return keys
}

// Helper method to get the logger instance
func (h *BaseAccountHandler) GetLogger() *logger.Logger {
	return h.logger
}

// Helper method to set operation context for the logger
func (h *BaseAccountHandler) SetLoggerOperation(operation string) {
	h.logger.SetOperation(operation)
}
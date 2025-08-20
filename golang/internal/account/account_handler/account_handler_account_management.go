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


func (h *AccountHandler) UpdateAccountStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this layer
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("update_account_status")
	
	// Base context for all logs
	baseContext := utils.CreateBaseContext(r, nil)
	
	// Get request ID for domain error handling
	requestID := errorcustom.GetRequestIDFromContext(ctx)
	domain := "account"
	
	handlerLog.Info("Account status update request started", baseContext)

	// Parse ID parameter
	id, err := errorcustom.ParseIDParam(r, "id", domain)
if err != nil {
	// Convert the error to APIError if it's not already one
	var apiErr *errorcustom.APIError
	if validationErr, ok := err.(*errorcustom.ValidationError); ok {
		apiErr = validationErr.ToAPIError()
	} else {
		apiErr = errorcustom.ConvertToAPIError(err)
	}
	
	context := utils.MergeContext(baseContext, map[string]interface{}{
		"error": err.Error(),
	})
	
	logger.ErrorWithCause(
		"Invalid ID parameter",
		"invalid_parameter",
		logger.LayerHandler,
		"parse_id",
		context,
	)
	
	// Log API request with error
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
	
	errorcustom.HandleError(w, apiErr, domain)
	return
}

	// Add user ID to context for subsequent logs
	baseContext["user_id"] = id

	var req struct {
		Status string `json:"status" validate:"required,oneof=active inactive suspended pending"`
	}

	// Decode request body
	if err := errorcustom.DecodeJSON(r.Body, &req, "update_account_status",h.domain); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to decode request body",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)
		
		// Log API request with error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.HandleError(w, err, "update_account_status")
		return
	}

	// Add requested status to context
	baseContext["requested_status"] = req.Status
	
	handlerLog.Debug("Request body decoded successfully", baseContext)

	// Validate request
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
				"Request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
		   errorcustom.HandleValidationErrors(w, validationErrors, domain, requestID)
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			errorcustom.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			), "update_account_status")
		}
		
		// Log API request with validation error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Request validation passed", baseContext)

	// Call user service
	serviceStart := time.Now()
	res, err := h.userClient.UpdateAccountStatus(ctx, &pb.UpdateAccountStatusReq{
		UserId: id,
		Status: req.Status,
	})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "UpdateAccountStatus",
		"duration_ms": serviceDuration.Milliseconds(),
	})
	
	if err != nil {
		// Determine error type and appropriate response
		if strings.Contains(err.Error(), "not found") {
			logger.LogServiceCall("user-service", "UpdateAccountStatus", false, err, serviceContext)
			
			logger.WarningWithCause(
				"User not found for status update",
				"user_not_found",
				logger.LayerHandler,
				"update_account_status",
				baseContext,
			)
			
			// Log API request with not found error
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusNotFound, time.Since(start), baseContext)
			
			errorcustom.HandleError(w, errorcustom.NewUserNotFoundByID(id), "update_account_status")
			return
		}
		
		// Log service call failure
		logger.LogServiceCall("user-service", "UpdateAccountStatus", false, err, serviceContext)
		
		logger.ErrorWithCause(
			"User service call failed",
			"service_error",
			logger.LayerExternal,
			"update_account_status",
			utils.MergeContext(baseContext, map[string]interface{}{
				"error": err.Error(),
			}),
		)
		
		// Log API request with service error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), baseContext)
		
		errorcustom.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Failed to update account status",
			http.StatusInternalServerError,
		), "update_account_status")
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "UpdateAccountStatus", true, nil, serviceContext)
	
	// Determine response status
	status := http.StatusOK
	if !res.Success {
		status = http.StatusBadRequest
		
		logger.WarningWithCause(
			"Account status update failed at service level",
			"service_business_logic_error",
			logger.LayerService,
			"update_account_status",
			utils.MergeContext(baseContext, map[string]interface{}{
				"service_message": res.Message,
				"service_success": res.Success,
			}),
		)
	} else {
		// Log successful user activity
		logger.LogUserActivity(
			fmt.Sprint(id),
			"", // email not available in this context
			"update",
			"account_status",
			utils.MergeContext(baseContext, map[string]interface{}{
				"old_status": "unknown", // Would need to be passed from service
				"new_status": req.Status,
			}),
		)
		
		handlerLog.Info("Account status updated successfully", utils.MergeContext(baseContext, map[string]interface{}{
			"service_message": res.Message,
		}))
	}

	// Log performance
	totalDuration := time.Since(start)
	logger.LogPerformance("update_account_status", totalDuration, baseContext)
	
	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, status, totalDuration, utils.MergeContext(baseContext, map[string]interface{}{
		"response_success": res.Success,
		"response_message": res.Message,
	}))

	// Send response
	errorcustom.RespondWithJSON(w, status, map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}, "update_account_status")
}
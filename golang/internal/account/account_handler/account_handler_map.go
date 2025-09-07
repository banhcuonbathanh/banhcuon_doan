package account_handler

import (
	"encoding/json"
	"english-ai-full/error_system"
	"english-ai-full/logger/core"
	"net/http"
	"time"

		pb "english-ai-full/internal/proto_qr/account"
)

// FIXED: handleError processes errors and sends appropriate HTTP responses
func (h *AccountHandler) handleError(w http.ResponseWriter, err error, operation string, context map[string]interface{}, startTime time.Time) {
	duration := time.Since(startTime)
	context["duration_ms"] = duration.Milliseconds()

	// CRITICAL: Check if it's already our AppError - DON'T WRAP IT
	if appErr, ok := error_system.IsAppError(err); ok {
		h.logger.Info("Received AppError, passing through unchanged", h.layerContext.MergeWithContext(map[string]interface{}{
			"error_code":    appErr.Code,
			"error_message": appErr.Message,
			"operation":     operation,
			"layer":         "handler",
		}))
		h.writeErrorResponse(w, appErr, context["request_id"].(string), startTime)
		return
	}

	// Only convert to AppError if it's not already one
	h.logger.Info("Converting raw error to AppError", h.layerContext.MergeWithContext(map[string]interface{}{
		"raw_error": err.Error(),
		"operation": operation,
		"layer":     "handler",
	}))
	
	appErr := h.errorHandler.Handle(err, operation, context)
	h.writeErrorResponse(w, appErr, context["request_id"].(string), startTime)
}

// writeErrorResponse sends the error response to the client
func (h *AccountHandler) writeErrorResponse(w http.ResponseWriter, appErr *error_system.AppError, requestID string, startTime time.Time) {
	duration := time.Since(startTime)
	
	// Log the error with the actual error code being returned
	h.logger.Error("Request completed with error", h.layerContext.MergeWithContext(map[string]interface{}{
		"error_code":   appErr.Code,
		"error_message": appErr.Message,
		"status_code":  appErr.HTTPStatus,
		"request_id":   requestID,
		"duration_ms":  duration.Milliseconds(),
		"final_response_code": string(appErr.Code), // Explicitly log what's being returned
	}))

	// Write the HTTP response
	appErr.WriteHTTPResponse(w)
}

// decodeJSONRequest parses JSON request body into the destination struct
func (h *AccountHandler) decodeJSONRequest(r *http.Request, dest interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return error_system.NewErrorWithInternal(error_system.ErrInvalidInput, "Invalid JSON format", err)
	}
	return nil
}

// writeSuccessResponse sends a successful HTTP response to the client
func (h *AccountHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}, statusCode int, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}
	
	json.NewEncoder(w).Encode(response)
}



// Helper methods
func (h *AccountHandler) logRequestStart(requestID, method, path, userAgent string) {
	h.logger.Info(core.MsgRequestStarted, h.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldRequestID:  requestID,
		core.FieldMethod:     method,
		core.FieldPath:       path,
		core.FieldUserAgent:  userAgent,
	}))
}

func (h *AccountHandler) logRequestEnd(requestID string, statusCode int, startTime time.Time) {
	duration := time.Since(startTime)
	
	logFunc := h.logger.Info
	if statusCode >= 400 {
		logFunc = h.logger.Error
	} else if statusCode >= 300 {
		logFunc = h.logger.Warn
	}
	
	logFunc(core.MsgRequestCompleted, h.layerContext.MergeWithContext(map[string]interface{}{
		core.FieldRequestID:   requestID,
		core.FieldStatusCode:  statusCode,
		core.FieldDurationMS:  duration.Milliseconds(),
	}))
}

func (h *AccountHandler) logStructValidationError(structName string, data interface{}, message string) {
	h.logger.ErrorWithCause(core.MsgValidationError, core.CauseStructValidationFailed, core.LayerValidation, core.OperationValidateStruct,
		h.layerContext.MergeWithContext(map[string]interface{}{
			core.FieldStructName: structName,
			core.FieldMessage:    message,
		}))
}

func (h *AccountHandler) getHTTPStatusFromError(err error) int {
	// Simple implementation - you can enhance this based on your error types
	if err == nil {
		return http.StatusOK
	}
	
	// You can add more sophisticated error type checking here
	return http.StatusInternalServerError
}

func (h *AccountHandler) prepareUserResponse(user *pb.Account) map[string]interface{} {
	return map[string]interface{}{
		core.ResponseFieldID:        user.Id,
		core.ResponseFieldEmail:     user.Email,
		core.ResponseFieldName:      user.Name,
		core.ResponseFieldRole:      user.Role,
		core.ResponseFieldCreatedAt: user.CreatedAt,
		// Don't include password or other sensitive fields
	}
}


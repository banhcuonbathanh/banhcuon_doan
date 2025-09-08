package account_handler

import (
	"encoding/json"
	"english-ai-full/error_system"
	"english-ai-full/logger/core"
	"english-ai-full/utils"
	"net/http"
	"regexp"
	"strings"
	"time"

	pb "english-ai-full/internal/proto_qr/account"

	"google.golang.org/grpc/status"
)


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



// new asdfgasdfads

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

	// NEW: Check if it's a gRPC error that contains our AppError
	if grpcAppErr := h.extractAppErrorFromGRPC(err); grpcAppErr != nil {
		h.logger.Info("Extracted AppError from gRPC error", h.layerContext.MergeWithContext(map[string]interface{}{
			"error_code":        grpcAppErr.Code,
			"error_message":     grpcAppErr.Message,
			"original_grpc_error": err.Error(),
			"operation":         operation,
			"layer":             "handler",
		}))
		h.writeErrorResponse(w, grpcAppErr, context["request_id"].(string), startTime)
		return
	}

	// Only convert to AppError if it's not already one and not a recoverable gRPC error
	h.logger.Info("Converting raw error to AppError", h.layerContext.MergeWithContext(map[string]interface{}{
		"raw_error": err.Error(),
		"operation": operation,
		"layer":     "handler",
	}))
	
	appErr := h.errorHandler.Handle(err, operation, context)
	h.writeErrorResponse(w, appErr, context["request_id"].(string), startTime)
}

// NEW: Extract AppError information from gRPC errors
func (h *AccountHandler) extractAppErrorFromGRPC(err error) *error_system.AppError {
	// Check if it's a gRPC status error
	if grpcStatus, ok := status.FromError(err); ok {
		return h.parseGRPCStatusToAppError(grpcStatus)
	}
	
	// Check if it's a gRPC error string format
	if strings.Contains(err.Error(), "rpc error:") {
		return h.parseGRPCStringToAppError(err.Error())
	}
	
	return nil
}

// Parse gRPC status to AppError
func (h *AccountHandler) parseGRPCStatusToAppError(grpcStatus *status.Status) *error_system.AppError {
	message := grpcStatus.Message()
	
	// Look for our AppError pattern in the message: [ERROR_CODE] Message
	return h.extractAppErrorFromMessage(message)
}

// Parse gRPC error string to AppError  
func (h *AccountHandler) parseGRPCStringToAppError(errStr string) *error_system.AppError {
	// Pattern: "rpc error: code = Unknown desc = [ERROR_CODE] Message"
	// Extract the description part after "desc = "
	descIndex := strings.Index(errStr, "desc = ")
	if descIndex == -1 {
		return nil
	}
	
	description := errStr[descIndex+7:] // Skip "desc = "
	return h.extractAppErrorFromMessage(description)
}

// Extract AppError from message with pattern [ERROR_CODE] Message
func (h *AccountHandler) extractAppErrorFromMessage(message string) *error_system.AppError {
	// Regex to match [ERROR_CODE] Message pattern
	re := regexp.MustCompile(`^\[([A-Z_]+)\]\s*(.+)$`)
	matches := re.FindStringSubmatch(message)
	
	if len(matches) != 3 {
		h.logger.Debug("Could not extract AppError from message", h.layerContext.MergeWithContext(map[string]interface{}{
			"message": message,
			"matches_count": len(matches),
		}))
		return nil
	}
	
	errorCode := error_system.ErrorCode(matches[1])
	errorMessage := strings.TrimSpace(matches[2])
	
	h.logger.Debug("Successfully extracted AppError from message", h.layerContext.MergeWithContext(map[string]interface{}{
		"extracted_code": errorCode,
		"extracted_message": errorMessage,
		"original_message": message,
	}))
	
	// Create AppError with extracted information
	appErr := error_system.NewError(errorCode, errorMessage)
	
	// Add any additional context based on error code
	switch errorCode {
	case error_system.ErrAccountDuplicate:
		// Extract email from context if available
		if email, exists := h.getEmailFromContext(); exists {
			appErr.Details["email"] = utils.MaskEmail(email)
		}
	}
	
	return appErr
}

// Helper to get email from current request context (if needed)
func (h *AccountHandler) getEmailFromContext() (string, bool) {
	// This would depend on how you store request context
	// For now, return empty - you can enhance this if needed
	return "", false
}

// Updated writeErrorResponse with better logging
func (h *AccountHandler) writeErrorResponse(w http.ResponseWriter, appErr *error_system.AppError, requestID string, startTime time.Time) {
	duration := time.Since(startTime)
	
	// Log the error with the actual error code being returned
	h.logger.Error("Request completed with error", h.layerContext.MergeWithContext(map[string]interface{}{
		"error_code":          appErr.Code,
		"error_message":       appErr.Message,
		"status_code":         appErr.HTTPStatus,
		"request_id":          requestID,
		"duration_ms":         duration.Milliseconds(),
		"final_response_code": string(appErr.Code), // Explicitly log what's being returned
		"response_will_contain": map[string]interface{}{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	}))

	// Write the HTTP response
	appErr.WriteHTTPResponse(w)
}
// new asdfasdfsd
// new done asdfadsf
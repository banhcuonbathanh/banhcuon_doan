package account_handler

import (
	"context"
	"encoding/json"
	"english-ai-full/error_system"
	"english-ai-full/internal/account/account_dto"
	pb "english-ai-full/internal/proto_qr/account"
	"net/http"
	"time"
)

func (h *AccountHandler) DailyCleanup(w http.ResponseWriter, r *http.Request) {
	const operation = "daily_cleanup"
	startTime := time.Now()
	
	// Extract request context information
	requestID := h.getRequestIDFromContext(r.Context())
	clientIP := h.layerContext.GetClientIP(r)
	
	// Build operation context using layer context
	operationCtx := h.layerContext.BuildOperationContext(operation, r, requestID)
	operationCtx["client_ip"] = clientIP
	operationCtx["job"] = "token_cleanup"
	
	// Set operation in logger
	h.logger.SetOperation(operation)
	
	// Log request start with layer context
	h.logRequestStart(requestID, r.Method, r.URL.Path, "")

	// Context timeout check
	if err := r.Context().Err(); err != nil {
		h.handleError(w, err, operation, operationCtx, startTime)
		return
	}

	// Create service context with timeout (longer timeout for cleanup operations)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Create protobuf request for cleanup (empty request)
	pbRequest := &pb.RunDailyCleanupReq{}

	// Log cleanup initiation
	h.logger.Info("Initiating daily token cleanup", h.layerContext.MergeWithContext(map[string]interface{}{
		"request_id": requestID,
		"job":        "token_cleanup",
	}))

	// Call service layer using RunDailyCleanup RPC
	cleanupResponse, err := h.userClient.RunDailyCleanup(ctx, pbRequest)
	if err != nil {
		duration := time.Since(startTime)
		operationCtx["duration_ms"] = duration.Milliseconds()
		
		h.logger.Error("Daily token cleanup failed", h.layerContext.MergeWithContext(map[string]interface{}{
			"error":       err.Error(),
			"request_id":  requestID,
			"duration_ms": duration.Milliseconds(),
			"job":         "token_cleanup",
		}))
		
		h.handleError(w, err, operation, operationCtx, startTime)
		return
	}

	// Validate response from service
	if cleanupResponse == nil {
		h.logger.Error("Received nil response from cleanup service", h.layerContext.MergeWithContext(map[string]interface{}{
			"request_id": requestID,
			"job":        "token_cleanup",
		}))
		
		appErr := error_system.NewError(error_system.ErrSystemError, "Cleanup service returned invalid response")
		h.writeErrorResponse(w, appErr, requestID, startTime)
		return
	}

	// Log successful cleanup with statistics
	operationCtx["cleanup_time"] = time.Now().Unix()
	
	h.logger.Info("Daily token cleanup completed successfully", h.layerContext.MergeWithContext(map[string]interface{}{
		"request_id":  requestID,
		"duration_ms": time.Since(startTime).Milliseconds(),
		"job":         "token_cleanup",
	}))

	// Prepare response data
	responseData := map[string]interface{}{
		"success":     cleanupResponse.Success,
		"message":     cleanupResponse.Message,
		"job":         "token_cleanup",
		"duration_ms": cleanupResponse.DurationMs,
	}

	// Log successful request completion
	h.logRequestEnd(requestID, http.StatusOK, startTime)

	// Send successful response
	h.writeSuccessResponse(w, responseData, http.StatusOK, requestID)
}


func (h *AccountHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
    var req account_dto.GetNewAccessTokenRequest
    
    // Parse request body
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if req.RefreshToken == "" {
        http.Error(w, "Refresh token is required", http.StatusBadRequest)
        return
    }
    
    // Call service layer
    accessToken, newRefreshToken, expiresIn, err := h.accountService.RefreshToken(r.Context(), req.RefreshToken)
    if err != nil {
        switch err {
        case services.ErrInvalidToken:
            http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
        case services.ErrUserNotFound:
            http.Error(w, "User not found", http.StatusNotFound)
        default:
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }
    
    // Prepare response
    response := RefreshTokenResponse{
        AccessToken:  accessToken,
        RefreshToken: newRefreshToken,
        ExpiresIn:    expiresIn,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
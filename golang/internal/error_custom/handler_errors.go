package errorcustom

import (
	
	"encoding/json"
	"english-ai-full/logger"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)
func NewHandlerErrorManager() *HandlerErrorManager {
    return &HandlerErrorManager{}
}


type HandlerErrorManager struct{}



// RespondWithError converts various error types to APIError and responds appropriately
func (h *HandlerErrorManager) RespondWithError(w http.ResponseWriter, err error, domain, requestID string) {
	if err == nil {
		return
	}

	// Convert to APIError if not already
	apiErr := ConvertToAPIError(err)
	if apiErr == nil {
		apiErr = NewAPIError(
			GetSystemErrorCode(domain),
			"An unexpected error occurred",
			http.StatusInternalServerError,
		).WithDomain(domain)
	}

	// Ensure domain and request ID are set
	if apiErr.Domain == "" {
		apiErr.WithDomain(domain)
	}
	if apiErr.RequestID == "" {
		apiErr.RequestID = requestID
	}

	// Add handler layer context
	apiErr.WithLayer("handler")

	// Log error if necessary
	if ShouldLogError(apiErr) {
		h.logError(apiErr, requestID)
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(apiErr.HTTPStatus)

	// Create response
	response := apiErr.ToErrorResponse()
	
	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode error response", map[string]interface{}{
			"original_error": apiErr.Error(),
			"encoding_error": err.Error(),
			"request_id":     requestID,
			"domain":         domain,
		})
		
		// Fallback to plain text error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ============================================================================
// REQUEST PARAMETER PARSING METHODS
// ============================================================================

// ParseIDParameter securely parses ID parameter from URL
func (h *HandlerErrorManager) ParseIDParameter(r *http.Request, paramName, domain, requestID string) (int64, error) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		return 0, NewValidationError(
			domain,
			paramName,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			nil,
		).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
	}

	// Validate format
	if !h.isValidIDFormat(idStr) {
		return 0, NewValidationError(
			domain,
			paramName,
			fmt.Sprintf("Invalid %s format: must be a positive integer", paramName),
			idStr,
		).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, NewValidationError(
			domain,
			paramName,
			fmt.Sprintf("Invalid %s: must be a positive integer", paramName),
			idStr,
		).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
	}

	return id, nil
}

// ParsePaginationParameters safely parses pagination parameters
func (h *HandlerErrorManager) ParsePaginationParameters(r *http.Request, domain, requestID string) (limit, offset int64, err error) {
	// Set defaults
	limit = 10
	offset = 0

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, parseErr := strconv.ParseInt(limitStr, 10, 64); parseErr != nil {
			return 0, 0, NewValidationError(
				domain,
				"limit",
				"Invalid limit parameter: must be an integer",
				limitStr,
			).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
		} else if l < 1 || l > 100 {
			return 0, 0, NewValidationError(
				domain,
				"limit",
				"Limit must be between 1 and 100",
				l,
			).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
		} else {
			limit = l
		}
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, parseErr := strconv.ParseInt(offsetStr, 10, 64); parseErr != nil {
			return 0, 0, NewValidationError(
				domain,
				"offset",
				"Invalid offset parameter: must be an integer",
				offsetStr,
			).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
		} else if o < 0 {
			return 0, 0, NewValidationError(
				domain,
				"offset",
				"Offset cannot be negative",
				o,
			).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
		} else {
			offset = o
		}
	}

	return limit, offset, nil
}



// DecodeJSONRequest decodes JSON request body with error handling
func (h *HandlerErrorManager) DecodeJSONRequest(r *http.Request, target interface{}, domain, requestID string) error {
	if r.Body == nil {
		return NewValidationError(
			domain,
			"request_body",
			"Request body is required",
			nil,
		).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return NewAPIError(
			GetInvalidInputCode(domain),
			"Failed to read request body",
			http.StatusBadRequest,
		).WithDomain(domain).WithLayer("handler").WithDetail("request_id", requestID).WithCause(err)
	}

	if len(bodyBytes) == 0 {
		return NewValidationError(
			domain,
			"request_body",
			"Request body cannot be empty",
			nil,
		).ToAPIError().WithLayer("handler").WithDetail("request_id", requestID)
	}

	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return h.handleJSONDecodeError(err, domain, requestID)
	}

	return nil
}

// ============================================================================
// VALIDATION HELPER METHODS
// ============================================================================

// ValidateRequiredFields validates that required fields are present and not empty
func (h *HandlerErrorManager) ValidateRequiredFields(data map[string]interface{}, requiredFields []string, domain, requestID string) error {
	errorCollection := NewErrorCollection(domain)

	for _, field := range requiredFields {
		value, exists := data[field]
		if !exists {
			errorCollection.Add(NewValidationError(
				domain,
				field,
				fmt.Sprintf("Field '%s' is required", field),
				nil,
			))
			continue
		}

		// Check for empty values
		if h.isEmpty(value) {
			errorCollection.Add(NewValidationError(
				domain,
				field,
				fmt.Sprintf("Field '%s' cannot be empty", field),
				value,
			))
		}
	}

	if errorCollection.HasErrors() {
		apiErr := errorCollection.ToAPIError()
		if apiErr != nil {
			apiErr.WithLayer("handler").WithDetail("request_id", requestID)
		}
		return apiErr
	}

	return nil
}

// ============================================================================
// SUCCESS RESPONSE METHODS
// ============================================================================

// RespondWithSuccess sends successful JSON response
func (h *HandlerErrorManager) RespondWithSuccess(w http.ResponseWriter, data interface{}, domain, requestID string) {
	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	if domain != "" {
		response["domain"] = domain
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode success response", map[string]interface{}{
			"error":      err.Error(),
			"request_id": requestID,
			"domain":     domain,
		})
		
		// Send fallback response
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// RespondWithCreated sends successful creation response
func (h *HandlerErrorManager) RespondWithCreated(w http.ResponseWriter, data interface{}, domain, requestID string) {
	response := map[string]interface{}{
		"success": true,
		"data":    data,
		"message": "Resource created successfully",
	}

	if domain != "" {
		response["domain"] = domain
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode created response", map[string]interface{}{
			"error":      err.Error(),
			"request_id": requestID,
			"domain":     domain,
		})
		
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ============================================================================
// PRIVATE HELPER METHODS
// ============================================================================

// logError logs error with appropriate severity
func (h *HandlerErrorManager) logError(apiErr *APIError, requestID string) {
	severity := GetErrorSeverity(apiErr)
	logContext := apiErr.GetLogContext()
	logContext["request_id"] = requestID
	logContext["layer"] = "handler"

	switch severity {
	case "ERROR":
		logger.Error("Handler layer error occurred", logContext)
	case "WARNING":
		logger.Warning("Handler layer warning occurred", logContext)
	default:
		logger.Info("Handler layer info occurred", logContext)
	}
}

// handleJSONDecodeError handles JSON decoding errors
func (h *HandlerErrorManager) handleJSONDecodeError(err error, domain, requestID string) error {
	errMsg := err.Error()
	
	switch {
	case strings.Contains(errMsg, "unexpected end of JSON input"):
		return NewAPIError(
			GetInvalidInputCode(domain),
			"Request body cannot be empty",
			http.StatusBadRequest,
		).WithDomain(domain).WithLayer("handler").WithDetail("request_id", requestID)

	case strings.Contains(errMsg, "invalid character"):
		return NewAPIError(
			GetInvalidInputCode(domain),
			"Invalid JSON syntax",
			http.StatusBadRequest,
		).WithDomain(domain).WithLayer("handler").WithDetail("request_id", requestID)

	case strings.Contains(errMsg, "cannot unmarshal"):
		return NewAPIError(
			GetValidationCode(domain),
			"JSON structure does not match expected format",
			http.StatusBadRequest,
		).WithDomain(domain).WithLayer("handler").WithDetail("request_id", requestID)

	default:
		return NewAPIError(
			GetInvalidInputCode(domain),
			"Invalid JSON format",
			http.StatusBadRequest,
		).WithDomain(domain).WithLayer("handler").WithDetail("request_id", requestID).WithCause(err)
	}
}

// isValidIDFormat checks if ID string format is valid
func (h *HandlerErrorManager) isValidIDFormat(idStr string) bool {
	if idStr == "" {
		return false
	}
	
	// Check if it contains only digits
	for _, char := range idStr {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}

// isEmpty checks if a value is empty
func (h *HandlerErrorManager) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}


func (hem *HandlerErrorManager) ParseSortingParameters(r *http.Request, allowedFields []string, domain, requestID string) (sortBy, sortOrder string, err error) {
	// Get sort_by parameter
	sortBy = r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = r.URL.Query().Get("sortBy")
	}
	if sortBy == "" {
		sortBy = r.URL.Query().Get("order_by")
	}
	
	// Get sort_order parameter
	sortOrder = r.URL.Query().Get("sort_order")
	if sortOrder == "" {
		sortOrder = r.URL.Query().Get("sortOrder")
	}
	if sortOrder == "" {
		sortOrder = r.URL.Query().Get("order")
	}
	
	// Set defaults if empty
	if sortBy == "" {
		sortBy = "id" // Default sort field
	}
	if sortOrder == "" {
		sortOrder = "desc" // Default sort order
	}
	
	// Validate sort order
	sortOrder = strings.ToLower(sortOrder)
	if sortOrder != "asc" && sortOrder != "desc" {
		return "", "", NewValidationErrorWithRules(
			domain,
			"sort_order",
			"Sort order must be 'asc' or 'desc'",
			sortOrder,
			map[string]interface{}{
				"allowed_values": []string{"asc", "desc"},
				"request_id":     requestID,
			},
		)
	}
	
	// Validate sort field against allowed fields
	if len(allowedFields) > 0 {
		isAllowed := false
		for _, field := range allowedFields {
			if sortBy == field {
				isAllowed = true
				break
			}
		}
		
		if !isAllowed {
			return "", "", NewValidationErrorWithRules(
				domain,
				"sort_by",
				fmt.Sprintf("Invalid sort field. Allowed fields: %s", strings.Join(allowedFields, ", ")),
				sortBy,
				map[string]interface{}{
					"allowed_fields": allowedFields,
					"request_id":     requestID,
				},
			)
		}
	}
	
	return sortBy, sortOrder, nil
}
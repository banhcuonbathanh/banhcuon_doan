// ============================================================================
// FILE: golang/internal/error_custom/layer/handler_errors.go
// ============================================================================
package layer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/logger"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

// HandlerErrorManager manages HTTP layer errors
type HandlerErrorManager struct {
	validator *validator.Validate
}

// NewHandlerErrorManager creates a new handler error manager
func NewHandlerErrorManager() *HandlerErrorManager {
	return &HandlerErrorManager{
		validator: validator.New(),
	}
}

// ============================================================================
// HTTP REQUEST PARSING ERRORS
// ============================================================================

// HandleJSONDecodeError handles JSON decoding errors with detailed context
// HandleJSONDecodeError handles JSON decoding errors with detailed context
func (h *HandlerErrorManager) HandleJSONDecodeError(err error, domain, requestID string) *errorcustom.APIError {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	
	switch {
	case err.Error() == "unexpected end of JSON input":
		return errorcustom.NewAPIError(
			errorcustom.GetInvalidInputCode(domain),
			"Request body cannot be empty",
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("error_type", "empty_body").
			WithDetail("request_id", requestID)

	case strings.HasPrefix(err.Error(), "invalid character"):
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			syntaxError = syntaxErr
		}
		return errorcustom.NewAPIError(
			errorcustom.GetInvalidInputCode(domain),
			"Invalid JSON syntax",
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("error_type", "json_syntax").
			WithDetail("byte_offset", syntaxError.Offset).
			WithDetail("request_id", requestID)

	case func() bool {
		var ok bool
		unmarshalTypeError, ok = err.(*json.UnmarshalTypeError)
		return ok
	}():
		return errorcustom.NewAPIError(
			errorcustom.GetInvalidInputCode(domain),
			fmt.Sprintf("Invalid value type for field '%s'", unmarshalTypeError.Field),
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("error_type", "type_mismatch").
			WithDetail("field", unmarshalTypeError.Field).
			WithDetail("expected_type", unmarshalTypeError.Type.String()).
			WithDetail("request_id", requestID)

	default:
		return errorcustom.NewAPIError(
			errorcustom.GetInvalidInputCode(domain),
			"Invalid JSON format",
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("error_type", "json_decode").
			WithDetail("original_error", err.Error()).
			WithDetail("request_id", requestID)
	}
}

// HandleValidationError handles struct validation errors
func (h *HandlerErrorManager) HandleValidationError(validationErrs validator.ValidationErrors, domain, requestID string) *errorcustom.APIError {
	errorCollection := errorcustom.NewErrorCollection(domain)
	
	for _, err := range validationErrs {
		field := strings.ToLower(err.Field())
		message := h.getValidationMessage(err)
		
		validationErr := errorcustom.NewValidationErrorWithRules(
			domain,
			field,
			message,
			err.Value(),
			map[string]interface{}{
				"tag":   err.Tag(),
				"param": err.Param(),
			},
		)
		errorCollection.Add(validationErr)
	}

	apiErr := errorCollection.ToAPIError()
	if apiErr != nil {
		apiErr.WithLayer("handler").WithDetail("request_id", requestID)
	}
	
	return apiErr
}

// ============================================================================
// URL PARAMETER PARSING ERRORS
// ============================================================================

// ParseIDParameter safely parses ID parameters with enhanced error handling
func (h *HandlerErrorManager) ParseIDParameter(r *http.Request, paramName, domain, requestID string) (int64, *errorcustom.APIError) {
	idStr := chi.URLParam(r, paramName)
	
	if idStr == "" {
		return 0, errorcustom.NewAPIError(
			errorcustom.GetValidationCode(domain),
			fmt.Sprintf("Missing required URL parameter: %s", paramName),
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("parameter_name", paramName).
			WithDetail("parameter_type", "id").
			WithDetail("request_id", requestID)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errorcustom.NewAPIError(
			errorcustom.GetValidationCode(domain),
			fmt.Sprintf("Invalid %s format: must be a valid integer", paramName),
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("parameter_name", paramName).
			WithDetail("parameter_value", idStr).
			WithDetail("expected_type", "int64").
			WithDetail("request_id", requestID)
	}

	if id <= 0 {
		return 0, errorcustom.NewAPIError(
			errorcustom.GetValidationCode(domain),
			fmt.Sprintf("Invalid %s: must be a positive integer", paramName),
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("parameter_name", paramName).
			WithDetail("parameter_value", id).
			WithDetail("minimum_value", 1).
			WithDetail("request_id", requestID)
	}

	return id, nil
}

// ParseStringParameter safely parses string parameters with validation
func (h *HandlerErrorManager) ParseStringParameter(r *http.Request, paramName, domain, requestID string, minLen, maxLen int) (string, *errorcustom.APIError) {
	value := chi.URLParam(r, paramName)
	
	if value == "" {
		return "", errorcustom.NewAPIError(
			errorcustom.GetValidationCode(domain),
			fmt.Sprintf("Missing required URL parameter: %s", paramName),
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("parameter_name", paramName).
			WithDetail("parameter_type", "string").
			WithDetail("request_id", requestID)
	}

	if minLen > 0 && len(value) < minLen {
		return "", errorcustom.NewAPIError(
			errorcustom.GetValidationCode(domain),
			fmt.Sprintf("Parameter %s must be at least %d characters long", paramName, minLen),
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("parameter_name", paramName).
			WithDetail("parameter_value", value).
			WithDetail("minimum_length", minLen).
			WithDetail("current_length", len(value)).
			WithDetail("request_id", requestID)
	}

	if maxLen > 0 && len(value) > maxLen {
		return "", errorcustom.NewAPIError(
			errorcustom.GetValidationCode(domain),
			fmt.Sprintf("Parameter %s cannot exceed %d characters", paramName, maxLen),
			http.StatusBadRequest,
		).WithDomain(domain).
			WithLayer("handler").
			WithDetail("parameter_name", paramName).
			WithDetail("parameter_value", value).
			WithDetail("maximum_length", maxLen).
			WithDetail("current_length", len(value)).
			WithDetail("request_id", requestID)
	}

	return value, nil
}

// ============================================================================
// QUERY PARAMETER PARSING ERRORS
// ============================================================================

// ParsePaginationParameters safely parses pagination with comprehensive validation
func (h *HandlerErrorManager) ParsePaginationParameters(r *http.Request, domain, requestID string) (limit, offset int64, apiErr *errorcustom.APIError) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	// Default values
	limit = 10
	offset = 0

	// Parse limit
	if limitStr != "" {
		l, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.GetValidationCode(domain),
				"Invalid limit parameter: must be a valid integer",
				http.StatusBadRequest,
			).WithDomain(domain).
				WithLayer("handler").
				WithDetail("parameter_name", "limit").
				WithDetail("parameter_value", limitStr).
				WithDetail("request_id", requestID)
		}

		if l < 1 {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.GetValidationCode(domain),
				"Invalid limit parameter: must be at least 1",
				http.StatusBadRequest,
			).WithDomain(domain).
				WithLayer("handler").
				WithDetail("parameter_name", "limit").
				WithDetail("parameter_value", l).
				WithDetail("minimum_value", 1).
				WithDetail("request_id", requestID)
		}

		if l > 100 {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.GetValidationCode(domain),
				"Invalid limit parameter: cannot exceed 100",
				http.StatusBadRequest,
			).WithDomain(domain).
				WithLayer("handler").
				WithDetail("parameter_name", "limit").
				WithDetail("parameter_value", l).
				WithDetail("maximum_value", 100).
				WithDetail("request_id", requestID)
		}

		limit = l
	}

	// Parse offset
	if offsetStr != "" {
		o, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.GetValidationCode(domain),
				"Invalid offset parameter: must be a valid integer",
				http.StatusBadRequest,
			).WithDomain(domain).
				WithLayer("handler").
				WithDetail("parameter_name", "offset").
				WithDetail("parameter_value", offsetStr).
				WithDetail("request_id", requestID)
		}

		if o < 0 {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.GetValidationCode(domain),
				"Invalid offset parameter: cannot be negative",
				http.StatusBadRequest,
			).WithDomain(domain).
				WithLayer("handler").
				WithDetail("parameter_name", "offset").
				WithDetail("parameter_value", o).
				WithDetail("minimum_value", 0).
				WithDetail("request_id", requestID)
		}

		offset = o
	}

	return limit, offset, nil
}

// ============================================================================
// HTTP RESPONSE HELPERS
// ============================================================================

// RespondWithError sends a standardized error response
func (h *HandlerErrorManager) RespondWithError(w http.ResponseWriter, err error, domain, requestID string) {
	apiErr := errorcustom.ConvertToAPIError(err)
	if apiErr == nil {
		apiErr = errorcustom.NewAPIError(
			errorcustom.GetSystemErrorCode(domain),
			"An unexpected error occurred",
			http.StatusInternalServerError,
		)
	}

	// Ensure domain and layer are set
	if apiErr.Domain == "" {
		apiErr.WithDomain(domain)
	}
	apiErr.WithLayer("handler")

	// Log error if needed
	if h.shouldLogError(apiErr) {
		severity := h.getErrorSeverity(apiErr)
		logContext := apiErr.GetLogContext()
		logContext["request_id"] = requestID

		switch severity {
		case "ERROR":
			logger.Error("Handler layer error", logContext)
		case "WARNING":
			logger.Warning("Handler layer warning", logContext)
		default:
			logger.Info("Handler layer info", logContext)
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(apiErr.HTTPStatus)

	// Send response
	response := apiErr.ToErrorResponse()
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode error response", map[string]interface{}{
			"original_error": apiErr.Error(),
			"encoding_error": err.Error(),
			"request_id":     requestID,
			"layer":          "handler",
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// getValidationMessage returns user-friendly validation messages
func (h *HandlerErrorManager) getValidationMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "uuid", "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid date/time", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// shouldLogError determines if error should be logged
func (h *HandlerErrorManager) shouldLogError(err *errorcustom.APIError) bool {
	// Always log server errors
	if err.HTTPStatus >= 500 {
		return true
	}
	
	// Don't log common client errors
	if err.HTTPStatus == http.StatusBadRequest || 
	   err.HTTPStatus == http.StatusUnauthorized ||
	   err.HTTPStatus == http.StatusNotFound {
		return false
	}
	
	return true
}

// getErrorSeverity returns error severity for logging
func (h *HandlerErrorManager) getErrorSeverity(err *errorcustom.APIError) string {
	if err.HTTPStatus >= 500 {
		return "ERROR"
	}
	if err.HTTPStatus >= 400 {
		return "INFO"
	}
	return "WARNING"
}

package account

import (
	"encoding/json"
	error "english-ai-full/internal/error_custom"
	"fmt"
	"log"
	"net/http"

	"errors"
)
var (
	ErrorUserNotFound   = errors.New("user not found")
	ErrUpdateUserFailed = errors.New("update user failed")
	ErrMissingParameter = errors.New("missing parameter")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrDecodeFailed     = errors.New("decode failed")
)



// Predefined error constructors with detailed information
func NewUserNotFoundError(userID string) *error.APIError {
	return &error.APIError{
		Code:       "USER_NOT_FOUND",
		Message:    "The requested user does not exist",
		HTTPStatus: http.StatusNotFound,
		Details: map[string]interface{}{
			"user_id": userID,
			"resource": "user",
			"suggestion": "Verify the user ID and try again",
		},
	}
}

func NewUpdateUserFailedError(userID string, reason string) *error.APIError {
	return &error.APIError{
		Code:       "UPDATE_USER_FAILED",
		Message:    "Failed to update user information",
		HTTPStatus: http.StatusInternalServerError,
		Details: map[string]interface{}{
			"user_id": userID,
			"reason": reason,
			"resource": "user",
			"suggestion": "Check user permissions and data validity",
		},
	}
}

func NewMissingParameterError(paramName string, paramType string) *error.APIError {
	return &error.APIError{
		Code:       "MISSING_PARAMETER",
		Message:    fmt.Sprintf("Required parameter '%s' is missing", paramName),
		HTTPStatus: http.StatusBadRequest,
		Details: map[string]interface{}{
			"parameter_name": paramName,
			"parameter_type": paramType,
			"location": "request body or query parameters",
			"suggestion": fmt.Sprintf("Include the '%s' parameter in your request", paramName),
		},
	}
}

func NewInvalidParameterError(paramName string, paramValue interface{}, expectedFormat string) *error.APIError {
	return &error.APIError{
		Code:       "INVALID_PARAMETER",
		Message:    fmt.Sprintf("Parameter '%s' has invalid format or value", paramName),
		HTTPStatus: http.StatusBadRequest,
		Details: map[string]interface{}{
			"parameter_name": paramName,
			"provided_value": paramValue,
			"expected_format": expectedFormat,
			"suggestion": fmt.Sprintf("Ensure '%s' matches the expected format: %s", paramName, expectedFormat),
		},
	}
}

func NewDecodeFailedError(contentType string, reason string) *error.APIError {
	return &error.APIError{
		Code:       "DECODE_FAILED",
		Message:    "Failed to decode request body",
		HTTPStatus: http.StatusBadRequest,
		Details: map[string]interface{}{
			"content_type": contentType,
			"reason": reason,
			"suggestion": "Verify request body format and content-type header",
			"expected_format": "application/json",
		},
	}
}

// Validation error for multiple field validation failures
func NewValidationError(fieldErrors map[string]string) *error.APIError {
	details := make(map[string]interface{})
	details["fields"] = fieldErrors
	details["suggestion"] = "Fix the validation errors and try again"
	
	return &error.APIError{
		Code:       "VALIDATION_ERROR",
		Message:    "One or more fields failed validation",
		HTTPStatus: http.StatusBadRequest,
		Details:    details,
	}
}

// Database error for internal database issues
func NewDatabaseError(operation string, table string) *error.APIError {
	return &error.APIError{
		Code:       "DATABASE_ERROR",
		Message:    "Internal database error occurred",
		HTTPStatus: http.StatusInternalServerError,
		Details: map[string]interface{}{
			"operation": operation,
			"table": table,
			"suggestion": "Please try again later or contact support",
		},
	}
}

// Authorization error
func NewAuthorizationError(action string, resource string) *error.APIError {
	return &error.APIError{
		Code:       "INSUFFICIENT_PERMISSIONS",
		Message:    "You don't have permission to perform this action",
		HTTPStatus: http.StatusForbidden,
		Details: map[string]interface{}{
			"action": action,
			"resource": resource,
			"suggestion": "Contact your administrator for access",
		},
	}
}

// Helper function to send error response
func SendErrorResponse(w http.ResponseWriter, err *error.APIError) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(err.HTTPStatus)
    
    // Get JSON bytes and handle potential marshaling error
    jsonBytes, jsonErr := err.ToJSON()
    if jsonErr != nil {
        // Handle JSON serialization failure
        log.Printf("Failed to marshal error: %v", jsonErr)
        
        // Create fallback error response
        fallback := map[string]string{
            "code":    "internal_error",
            "message": "Failed to generate error response",
        }
        fallbackJson, _ := json.Marshal(fallback)
        w.Write(fallbackJson)
        return
    }
    
    // Write original JSON response
    w.Write(jsonBytes)
}


// Mock function for example

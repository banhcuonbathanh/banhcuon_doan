package errorcustom

import (
	"encoding/json"
	"fmt"
)

// APIError represents a structured API error with detailed information
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
}

// NewAPIError creates a new APIError instance
func NewAPIError(code, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// WithDetail adds a key-value pair to error details
func (e *APIError) WithDetail(key string, value interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// ToJSON converts the error to JSON bytes
func (e *APIError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
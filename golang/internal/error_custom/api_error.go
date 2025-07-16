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
	HTTPStatus int                    `json:"-"` // Don't expose HTTP status in JSON
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ToJSON converts the error to JSON bytes
func (e *APIError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

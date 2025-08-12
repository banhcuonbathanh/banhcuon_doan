// internal/error_custom/core.go
// Package errorcustom provides structured error handling for multi-domain API
package errorcustom

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ============================================================================
// CORE ERROR TYPES
// ============================================================================

// APIError represents a structured API error with detailed information
// type APIError struct {
// 	Code       string                 `json:"code"`
// 	Message    string                 `json:"message"`
// 	Details    map[string]interface{} `json:"details,omitempty"`
// 	HTTPStatus int                    `json:"-"`
// 	Domain     string                 `json:"domain,omitempty"`     // user, course, payment, etc.
// 	Layer      string                 `json:"layer,omitempty"`      // handler, service, repository
// 	Operation  string                 `json:"operation,omitempty"`  // login, register, create_course, etc.
// 	Cause      error                  `json:"-"`                    // Original error for internal use
// 	Retryable  bool                   `json:"retryable,omitempty"`  // Whether the operation can be retried
// }

// ErrorResponse represents the standard error format for API responses
// swagger:model ErrorResponse
type ErrorResponse struct {
	Code    string                 `json:"code" example:"validation_error"`
	Message string                 `json:"message" example:"Validation failed"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// DomainError interface that all domain-specific errors must implement
type DomainError interface {
	error
	ToAPIError() *APIError
	GetDomain() string
	GetErrorType() string
}

// BaseError provides common functionality for domain errors
type BaseError struct {
	Domain    string `json:"domain"`
	ErrorType string `json:"error_type"`
}

func (b *BaseError) GetDomain() string {
	return b.Domain
}

func (b *BaseError) GetErrorType() string {
	return b.ErrorType
}

// ============================================================================
// GENERIC ERROR TYPES
// ============================================================================

// NotFoundError represents when a resource cannot be found
type NotFoundError struct {
	BaseError
	ResourceType string                 `json:"resource_type"`
	ResourceID   interface{}            `json:"resource_id,omitempty"`
	Identifiers  map[string]interface{} `json:"identifiers,omitempty"`
}

// ValidationError represents validation failures
type ValidationError struct {
	BaseError
	Field   string                 `json:"field"`
	Message string                 `json:"message"`
	Value   interface{}            `json:"value,omitempty"`
	Rules   map[string]interface{} `json:"rules,omitempty"`
}

// DuplicateError represents duplicate resource errors
type DuplicateError struct {
	BaseError
	ResourceType string                 `json:"resource_type"`
	Field        string                 `json:"field"`
	Value        interface{}            `json:"value"`
	Constraints  map[string]interface{} `json:"constraints,omitempty"`
}

// AuthenticationError represents authentication failures
type AuthenticationError struct {
	BaseError
	Reason    string                 `json:"reason"`
	Step      string                 `json:"step,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// AuthorizationError represents authorization failures
type AuthorizationError struct {
	BaseError
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// BusinessLogicError represents business rule violations
type BusinessLogicError struct {
	BaseError
	Rule        string                 `json:"rule"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ExternalServiceError represents errors from external services
type ExternalServiceError struct {
	BaseError
	Service   string                 `json:"service"`
	Operation string                 `json:"operation"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"-"`
	Retryable bool                   `json:"retryable"`
}

// SystemError represents internal system errors
type SystemError struct {
	BaseError
	Component string                 `json:"component"`
	Operation string                 `json:"operation"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"-"`
}

// ============================================================================
// CORE ERROR METHODS
// ============================================================================

// Error implements the error interface
func (e *APIError) Error() string {
	parts := []string{}
	
	if e.Domain != "" {
		parts = append(parts, e.Domain)
	}
	if e.Layer != "" {
		parts = append(parts, e.Layer)
	}
	if e.Operation != "" {
		parts = append(parts, e.Operation)
	}
	
	prefix := ""
	if len(parts) > 0 {
		prefix = fmt.Sprintf("[%s]", strings.Join(parts, ":"))
	}
	
	return fmt.Sprintf("%s[%s] %s", prefix, e.Code, e.Message)
}

// WithDomain sets the domain information and returns the APIError for chaining
func (e *APIError) WithDomain(domain string) *APIError {
	e.Domain = domain
	return e
}

// WithLayer sets the layer information and returns the APIError for chaining
func (e *APIError) WithLayer(layer string) *APIError {
	e.Layer = layer
	return e
}

// WithOperation sets the operation information and returns the APIError for chaining
func (e *APIError) WithOperation(operation string) *APIError {
	e.Operation = operation
	return e
}

// WithDetail adds a detail to the error and returns the APIError for chaining
func (e *APIError) WithDetail(key string, value interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithCause sets the underlying cause and returns the APIError for chaining
func (e *APIError) WithCause(cause error) *APIError {
	e.Cause = cause
	return e
}

// WithRetryable sets whether the error is retryable
func (e *APIError) WithRetryable(retryable bool) *APIError {
	e.Retryable = retryable
	return e
}

// GetLogContext returns context information suitable for logging
func (e *APIError) GetLogContext() map[string]interface{} {
	context := map[string]interface{}{
		"error_code":    e.Code,
		"error_message": e.Message,
		"http_status":   e.HTTPStatus,
	}

	if e.Domain != "" {
		context["domain"] = e.Domain
	}
	if e.Layer != "" {
		context["layer"] = e.Layer
	}
	if e.Operation != "" {
		context["operation"] = e.Operation
	}
	if e.Cause != nil {
		context["cause"] = e.Cause.Error()
	}
	if e.Retryable {
		context["retryable"] = e.Retryable
	}

	return context
}

// ToJSON converts the error to JSON bytes
func (e *APIError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// ToErrorResponse converts APIError to Swagger-compatible format
func (e *APIError) ToErrorResponse() ErrorResponse {
	response := ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
	
	// Add retryable info if applicable
	if e.Retryable {
		if response.Details == nil {
			response.Details = make(map[string]interface{})
		}
		response.Details["retryable"] = e.Retryable
	}
	
	return response
}

// ============================================================================
// GENERIC ERROR IMPLEMENTATIONS
// ============================================================================

func (e *NotFoundError) Error() string {
	if e.ResourceID != nil {
		return fmt.Sprintf("%s with ID %v not found", e.ResourceType, e.ResourceID)
	}
	if len(e.Identifiers) > 0 {
		var pairs []string
		for k, v := range e.Identifiers {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
		}
		return fmt.Sprintf("%s with %s not found", e.ResourceType, strings.Join(pairs, ", "))
	}
	return fmt.Sprintf("%s not found", e.ResourceType)
}

func (e *NotFoundError) ToAPIError() *APIError {
	apiErr := NewAPIError(
		GetNotFoundCode(e.Domain),
		e.Error(),
		http.StatusNotFound,
	).WithDomain(e.Domain)
	
	if e.ResourceID != nil {
		apiErr.WithDetail("resource_id", e.ResourceID)
	}
	if e.ResourceType != "" {
		apiErr.WithDetail("resource_type", e.ResourceType)
	}
	if len(e.Identifiers) > 0 {
		apiErr.WithDetail("identifiers", e.Identifiers)
	}
	
	return apiErr
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) ToAPIError() *APIError {
	apiErr := NewAPIError(
		GetValidationCode(e.Domain),
		e.Error(),
		http.StatusBadRequest,
	).WithDomain(e.Domain).
		WithDetail("field", e.Field)
	
	if e.Value != nil {
		apiErr.WithDetail("value", e.Value)
	}
	if len(e.Rules) > 0 {
		apiErr.WithDetail("rules", e.Rules)
	}
	
	return apiErr
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("%s with %s=%v already exists", e.ResourceType, e.Field, e.Value)
}

func (e *DuplicateError) ToAPIError() *APIError {
	return NewAPIError(
		GetDuplicateCode(e.Domain),
		e.Error(),
		http.StatusConflict,
	).WithDomain(e.Domain).
		WithDetail("resource_type", e.ResourceType).
		WithDetail("field", e.Field).
		WithDetail("value", e.Value)
}

func (e *AuthenticationError) Error() string {
	if e.Step != "" {
		return fmt.Sprintf("authentication failed at %s: %s", e.Step, e.Reason)
	}
	return fmt.Sprintf("authentication failed: %s", e.Reason)
}

func (e *AuthenticationError) ToAPIError() *APIError {
	apiErr := NewAPIError(
		GetAuthenticationCode(e.Domain),
		"Authentication failed",
		http.StatusUnauthorized,
	).WithDomain(e.Domain).
		WithDetail("reason", e.Reason)
	
	if e.Step != "" {
		apiErr.WithDetail("step", e.Step)
	}
	if len(e.Context) > 0 {
		for k, v := range e.Context {
			apiErr.WithDetail(k, v)
		}
	}
	
	return apiErr
}

func (e *AuthorizationError) Error() string {
	return fmt.Sprintf("not authorized to %s %s", e.Action, e.Resource)
}

func (e *AuthorizationError) ToAPIError() *APIError {
	apiErr := NewAPIError(
		GetAuthorizationCode(e.Domain),
		"Access denied",
		http.StatusForbidden,
	).WithDomain(e.Domain).
		WithDetail("action", e.Action).
		WithDetail("resource", e.Resource)
	
	if len(e.Context) > 0 {
		for k, v := range e.Context {
			apiErr.WithDetail(k, v)
		}
	}
	
	return apiErr
}

func (e *BusinessLogicError) Error() string {
	return fmt.Sprintf("business rule violation: %s - %s", e.Rule, e.Description)
}

func (e *BusinessLogicError) ToAPIError() *APIError {
	apiErr := NewAPIError(
		GetBusinessLogicCode(e.Domain),
		e.Description,
		http.StatusUnprocessableEntity,
	).WithDomain(e.Domain).
		WithDetail("rule", e.Rule)
	
	if len(e.Context) > 0 {
		for k, v := range e.Context {
			apiErr.WithDetail(k, v)
		}
	}
	
	return apiErr
}

func (e *ExternalServiceError) Error() string {
	return fmt.Sprintf("external service error in %s.%s: %s", e.Service, e.Operation, e.Message)
}

func (e *ExternalServiceError) ToAPIError() *APIError {
	code := GetExternalServiceCode(e.Domain)
	if e.Retryable {
		code = GetServiceUnavailableCode(e.Domain)
	}
	
	status := http.StatusInternalServerError
	if e.Retryable {
		status = http.StatusServiceUnavailable
	}
	
	return NewAPIError(code, "External service error", status).
		WithDomain(e.Domain).
		WithDetail("service", e.Service).
		WithDetail("operation", e.Operation).
		WithRetryable(e.Retryable).
		WithCause(e.Cause)
}

func (e *SystemError) Error() string {
	return fmt.Sprintf("system error in %s.%s: %s", e.Component, e.Operation, e.Message)
}

func (e *SystemError) ToAPIError() *APIError {
	return NewAPIError(
		GetSystemErrorCode(e.Domain),
		"Internal system error",
		http.StatusInternalServerError,
	).WithDomain(e.Domain).
		WithDetail("component", e.Component).
		WithDetail("operation", e.Operation).
		WithCause(e.Cause)
}

// ============================================================================
// CORE CONSTRUCTORS
// ============================================================================

// NewAPIError creates a new APIError instance
// func NewAPIError(code, message string, httpStatus int) *APIError {
// 	return &APIError{
// 		Code:       code,
// 		Message:    message,
// 		HTTPStatus: httpStatus,
// 		Details:    make(map[string]interface{}),
// 	}
// }

// NewAPIErrorWithContext creates a new APIError instance with full context
func NewAPIErrorWithContext(code, message string, httpStatus int, domain, layer, operation string, cause error) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Domain:     domain,
		Layer:      layer,
		Operation:  operation,
		Cause:      cause,
		Details:    make(map[string]interface{}),
	}
}

// NewErrorResponse creates a new ErrorResponse instance
func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail to ErrorResponse
func (er ErrorResponse) WithDetail(key string, value interface{}) ErrorResponse {
	if er.Details == nil {
		er.Details = make(map[string]interface{})
	}
	er.Details[key] = value
	return er
}


// new start12121


// // Supported domains
// const (
// 	DomainUser    = "user"
// 	DomainCourse  = "course"
// 	DomainPayment = "payment"
// 	DomainAuth    = "auth"
// 	DomainAdmin   = "admin"
// 	DomainContent = "content"
// 	DomainSystem  = "system"
// )

// Core error types
// const (
// 	ErrorTypeNotFound         = "NOT_FOUND"
// 	ErrorTypeValidation       = "VALIDATION_ERROR"
// 	ErrorTypeDuplicate        = "DUPLICATE"
// 	ErrorTypeAuthentication   = "AUTHENTICATION_ERROR"
// 	ErrorTypeAuthorization    = "AUTHORIZATION_ERROR"
// 	ErrorTypeBusinessLogic    = "BUSINESS_LOGIC_ERROR"
// 	ErrorTypeExternalService  = "EXTERNAL_SERVICE_ERROR"
// 	ErrorTypeSystem           = "SYSTEM_ERROR"
// 	ErrorTypeDatabase         = "DATABASE_ERROR"
// 	ErrorTypeRateLimit        = "RATE_LIMIT_ERROR"
// 	ErrorTypeTimeout          = "TIMEOUT_ERROR"
// )

// Context keys for request metadata
type contextKey string

const (
	ContextKeyRequestID contextKey = "request_id"
	ContextKeyDomain    contextKey = "domain"
	ContextKeyUserID    contextKey = "user_id"
)

// APIError represents the core error structure with domain awareness
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	HTTPStatus int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Domain     string                 `json:"domain,omitempty"`
	Layer      string                 `json:"-"`
	Operation  string                 `json:"-"`
	Retryable  bool                   `json:"-"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	Cause      error                  `json:"-"`
}

// Error implements the error interface
// func (e *APIError) Error() string {
// 	if e.Domain != "" {
// 		return fmt.Sprintf("[%s] %s: %s", e.Domain, e.Code, e.Message)
// 	}
// 	return fmt.Sprintf("%s: %s", e.Code, e.Message)
// }

// WithDomain adds domain context to the error
// func (e *APIError) WithDomain(domain string) *APIError {
// 	e.Domain = domain
// 	return e
// }

// WithDetail adds additional context to the error
// func (e *APIError) WithDetail(key string, value interface{}) *APIError {
// 	if e.Details == nil {
// 		e.Details = make(map[string]interface{})
// 	}
// 	e.Details[key] = value
// 	return e
// }

// WithLayer adds layer information for debugging
// func (e *APIError) WithLayer(layer string) *APIError {
// 	e.Layer = layer
// 	return e
// }

// WithOperation adds operation context
// func (e *APIError) WithOperation(operation string) *APIError {
// 	e.Operation = operation
// 	return e
// }

// // WithRetryable marks error as retryable or not
// func (e *APIError) WithRetryable(retryable bool) *APIError {
// 	e.Retryable = retryable
// 	return e
// }

// WithCause adds the underlying cause
// func (e *APIError) WithCause(cause error) *APIError {
// 	e.Cause = cause
// 	return e
// }

// ToErrorResponse converts to HTTP response format
// func (e *APIError) ToErrorResponse() map[string]interface{} {
// 	response := map[string]interface{}{
// 		"code":      e.Code,
// 		"message":   e.Message,
// 		"timestamp": e.Timestamp.UTC().Format(time.RFC3339),
// 	}
	
// 	if len(e.Details) > 0 {
// 		response["details"] = e.Details
// 	}
	
// 	if e.RequestID != "" {
// 		response["request_id"] = e.RequestID
// 	}
	
// 	if e.Domain != "" {
// 		response["domain"] = e.Domain
// 	}
	
// 	return response
// }

// GetLogContext returns structured logging context
// func (e *APIError) GetLogContext() map[string]interface{} {
// 	context := map[string]interface{}{
// 		"error_code":   e.Code,
// 		"error_msg":    e.Message,
// 		"http_status":  e.HTTPStatus,
// 		"retryable":    e.Retryable,
// 		"timestamp":    e.Timestamp,
// 	}
	
// 	if e.Domain != "" {
// 		context["domain"] = e.Domain
// 	}
	
// 	if e.Layer != "" {
// 		context["layer"] = e.Layer
// 	}
	
// 	if e.Operation != "" {
// 		context["operation"] = e.Operation
// 	}
	
// 	if e.RequestID != "" {
// 		context["request_id"] = e.RequestID
// 	}
	
// 	if e.Details != nil {
// 		context["details"] = e.Details
// 	}
	
// 	return context
// }

// NewAPIError creates a new API error
func NewAPIError(code, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Details:    make(map[string]interface{}),
		Timestamp:  time.Now().UTC(),
		Retryable:  false,
	}
}

// ErrorCollection manages multiple related errors
// type ErrorCollection struct {
// 	Domain string      `json:"domain"`
// 	Errors []*APIError `json:"errors"`
// }

// NewErrorCollection creates a new error collection for a domain
// func NewErrorCollection(domain string) *ErrorCollection {
// 	return &ErrorCollection{
// 		Domain: domain,
// 		Errors: make([]*APIError, 0),
// 	}
// }

// Add adds an error to the collection
func (ec *ErrorCollection) Add(err error) {
	apiErr := ConvertToAPIError(err)
	if apiErr != nil {
		if apiErr.Domain == "" {
			apiErr.Domain = ec.Domain
		}
		ec.Errors = append(ec.Errors, apiErr)
	}
}

// HasErrors returns true if collection has errors
func (ec *ErrorCollection) HasErrors() bool {
	return len(ec.Errors) > 0
}

// Count returns the number of errors
func (ec *ErrorCollection) Count() int {
	return len(ec.Errors)
}

// ToAPIError converts collection to single API error
// func (ec *ErrorCollection) ToAPIError() *APIError {
// 	if len(ec.Errors) == 0 {
// 		return nil
// 	}
	
// 	if len(ec.Errors) == 1 {
// 		return ec.Errors[0]
// 	}
	
// 	// Multiple errors - create collection error
// 	code := GetValidationCode(ec.Domain)
// 	apiErr := NewAPIError(
// 		code,
// 		"Multiple validation errors occurred",
// 		http.StatusBadRequest,
// 	).WithDomain(ec.Domain)
	
// 	// Convert errors to map format
// 	errorDetails := make([]map[string]interface{}, len(ec.Errors))
// 	for i, err := range ec.Errors {
// 		errorDetails[i] = err.ToErrorResponse()
// 	}
	
// 	apiErr.WithDetail("errors", errorDetails)
// 	return apiErr
// }

// new end121
// internal/error_custom/utilities.go
// Error detection and utility functions
package errorcustom

import (
	"fmt"
	"net/http"
	"strings"
)

// ============================================================================
// ERROR TYPE DETECTION
// ============================================================================

// IsDomainError checks if an error is a domain-specific error
func IsDomainError(err error, domain string) bool {
	if err == nil {
		return false
	}

	if domainErr, ok := err.(DomainError); ok {
		return domainErr.GetDomain() == domain
	}

	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Domain == domain
	}

	return false
}

// IsErrorType checks if an error matches a specific type
func IsErrorType(err error, errorType string) bool {
	if err == nil {
		return false
	}

	if domainErr, ok := err.(DomainError); ok {
		return domainErr.GetErrorType() == errorType
	}

	if apiErr, ok := err.(*APIError); ok {
		baseType := GetBaseErrorType(apiErr.Code)
		return baseType == errorType
	}

	return false
}

// IsNotFoundError determines if error is a not found error
func IsNotFoundError(err error) bool {
	return IsErrorType(err, ErrorTypeNotFound) || 
		   isErrorByStringMatch(err, "not found")
}

// IsValidationError determines if error is a validation error
func IsValidationError(err error) bool {
	return IsErrorType(err, ErrorTypeValidation) || 
		   isErrorByStringMatch(err, "validation", "invalid")
}

// IsAuthenticationError determines if error is an authentication error
func IsAuthenticationError(err error) bool {
	return IsErrorType(err, ErrorTypeAuthentication) || 
		   isErrorByStringMatch(err, "authentication", "credentials", "login")
}

// IsAuthorizationError determines if error is an authorization error
func IsAuthorizationError(err error) bool {
	return IsErrorType(err, ErrorTypeAuthorization) || 
		   isErrorByStringMatch(err, "authorization", "access denied", "forbidden")
}

// IsBusinessLogicError determines if error is a business logic error
func IsBusinessLogicError(err error) bool {
	return IsErrorType(err, ErrorTypeBusinessLogic)
}

// IsExternalServiceError determines if error is an external service error
func IsExternalServiceError(err error) bool {
	return IsErrorType(err, ErrorTypeExternalService) || 
		   IsErrorType(err, ErrorTypeServiceUnavailable)
}

// IsSystemError determines if error is a system error
func IsSystemError(err error) bool {
	return IsErrorType(err, ErrorTypeSystem)
}

// IsRetryableError determines if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check ExternalServiceError
	if extErr, ok := err.(*ExternalServiceError); ok {
		return extErr.Retryable
	}

	// Check APIError with retryable flag or status codes
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.Retryable {
			return true
		}
		
		return apiErr.HTTPStatus == http.StatusServiceUnavailable ||
			apiErr.HTTPStatus == http.StatusRequestTimeout ||
			apiErr.HTTPStatus == http.StatusTooManyRequests ||
			(apiErr.HTTPStatus >= 500 && apiErr.HTTPStatus < 600)
	}

	return false
}

// IsClientError determines if an error is a client error (4xx)
func IsClientError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500
	}
	return false
}

// IsServerError determines if an error is a server error (5xx)
func IsServerError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.HTTPStatus >= 500 && apiErr.HTTPStatus < 600
	}
	return false
}



// IsUserNotFoundError determines if error is related to user not found
func IsUserNotFoundError(err error) bool {
	return IsDomainError(err, DomainAccount) && IsNotFoundError(err)
}

// IsPasswordError determines if error is password related
func IsPasswordError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a domain-specific authentication error
	if IsDomainError(err, DomainAccount) && IsAuthenticationError(err) {
		return true
	}

	// Check APIError codes
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == ErrCodeAuthFailed ||
			apiErr.Code == "AUTHENTICATION_ERROR" ||
			apiErr.Code == "INVALID_CREDENTIALS"
	}

	// Fallback to string matching
	return isErrorByStringMatch(err, "password", "invalid credentials", "authentication failed")
}



// ParseGRPCError parses gRPC error messages and creates appropriate domain errors
func ParseGRPCError(err error, domain, operation string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "not found"):
		resourceType := "resource"
		if domain != "" {
			resourceType = domain
		}
		
		notFoundErr := NewNotFoundError(domain, resourceType, nil)
		apiErr := notFoundErr.ToAPIError().
			WithLayer("service").
			WithOperation(operation).
			WithCause(err)
		
		// Add context if provided
		for k, v := range context {
			apiErr.WithDetail(k, v)
		}
		
		return apiErr

	case strings.Contains(errMsg, "invalid password") ||
		 strings.Contains(errMsg, "password") ||
		 strings.Contains(errMsg, "invalid email or password"):
		
		authErr := NewAuthenticationError(domain, "Invalid credentials")
		apiErr := authErr.ToAPIError().
			WithLayer("service").
			WithOperation(operation).
			WithCause(err)
		
		for k, v := range context {
			apiErr.WithDetail(k, v)
		}
		
		return apiErr

	case strings.Contains(errMsg, "account disabled"):
		authErr := NewAuthenticationError(domain, "Account disabled")
		return authErr.ToAPIError().
			WithLayer("service").
			WithOperation(operation).
			WithCause(err)

	case strings.Contains(errMsg, "account locked"):
		authErr := NewAuthenticationError(domain, "Account locked")
		return authErr.ToAPIError().
			WithLayer("service").
			WithOperation(operation).
			WithCause(err)

	case strings.Contains(errMsg, "already exists"):
		duplicateErr := NewDuplicateError(domain, domain, "identifier", context["email"])
		return duplicateErr.ToAPIError().
			WithLayer("service").
			WithOperation(operation).
			WithCause(err)

	case strings.Contains(errMsg, "connection refused") || 
		 strings.Contains(errMsg, "unavailable"):
		
		serviceErr := NewExternalServiceError(domain, "grpc", operation, "Service unavailable", err, true)
		return serviceErr.ToAPIError().
			WithLayer("service").
			WithOperation(operation)

	default:
		systemErr := NewSystemError(domain, "grpc", operation, "Internal server error", err)
		return systemErr.ToAPIError().
			WithLayer("service").
			WithOperation(operation)
	}
}


func ConvertToAPIError(err error) *APIError {
	if err == nil {
		return nil
	}

	// Already an APIError
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}

	// Domain error
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.ToAPIError()
	}

	// Legacy error types
	switch e := err.(type) {
	case *NotFoundError:
		return e.ToAPIError()
	case *ValidationError:
		return e.ToAPIError()
	case *DuplicateError:
		return e.ToAPIError()
	case *AuthenticationError:
		return e.ToAPIError()
	case *AuthorizationError:
		return e.ToAPIError()
	case *BusinessLogicError:
		return e.ToAPIError()
	case *ExternalServiceError:
		return e.ToAPIError()
	case *SystemError:
		return e.ToAPIError()
	default:
		// Generic error
		return NewAPIError(
			ErrorTypeSystem,
			"An unexpected error occurred",
			http.StatusInternalServerError,
		).WithCause(err)
	}
}


type ErrorCollection struct {
	Errors []error `json:"errors"`
	Domain string  `json:"domain"`
}

// NewErrorCollection creates a new error collection
func NewErrorCollection(domain string) *ErrorCollection {
	return &ErrorCollection{
		Errors: make([]error, 0),
		Domain: domain,
	}
}

// Add adds an error to the collection
// func (ec *ErrorCollection) Add(err error) {
// 	if err != nil {
// 		ec.Errors = append(ec.Errors, err)
// 	}
// }

// HasErrors returns true if the collection has errors
// func (ec *ErrorCollection) HasErrors() bool {
// 	return len(ec.Errors) > 0
// }

// ToAPIError converts the collection to a single APIError
func (ec *ErrorCollection) ToAPIError() *APIError {
	if !ec.HasErrors() {
		return nil
	}

	if len(ec.Errors) == 1 {
		return ConvertToAPIError(ec.Errors[0])
	}

	// Multiple errors - create a composite error
	apiErr := NewAPIError(
		GetValidationCode(ec.Domain),
		"Multiple validation errors occurred",
		http.StatusBadRequest,
	).WithDomain(ec.Domain)

	errorDetails := make([]map[string]interface{}, 0, len(ec.Errors))
	for _, err := range ec.Errors {
		if convertedErr := ConvertToAPIError(err); convertedErr != nil {
			errorDetails = append(errorDetails, map[string]interface{}{
				"code":    convertedErr.Code,
				"message": convertedErr.Message,
				"details": convertedErr.Details,
			})
		}
	}

	apiErr.WithDetail("errors", errorDetails)
	return apiErr
}

// Error implements the error interface
func (ec *ErrorCollection) Error() string {
	if !ec.HasErrors() {
		return ""
	}

	if len(ec.Errors) == 1 {
		return ec.Errors[0].Error()
	}

	var messages []string
	for _, err := range ec.Errors {
		messages = append(messages, err.Error())
	}

	return strings.Join(messages, "; ")
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// isErrorByStringMatch performs case-insensitive string matching on error messages
func isErrorByStringMatch(err error, keywords ...string) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	for _, keyword := range keywords {
		if strings.Contains(errMsg, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// GetErrorSeverity returns the severity level of an error for logging purposes
func GetErrorSeverity(err error) string {
	if err == nil {
		return "INFO"
	}

	if IsServerError(err) {
		return "ERROR"
	}

	if IsExternalServiceError(err) {
		return "WARNING"
	}

	if IsClientError(err) {
		return "INFO"
	}

	return "WARNING"
}

// ShouldLogError determines if an error should be logged based on its type and severity
func ShouldLogError(err error) bool {
	if err == nil {
		return false
	}

	// Always log server errors
	if IsServerError(err) {
		return true
	}

	// Log external service errors
	if IsExternalServiceError(err) {
		return true
	}

	// Don't log client validation errors
	if IsValidationError(err) || IsAuthenticationError(err) {
		return false
	}

	return true
}


// utilities.go

// ParseIDParamWithDomain parses an ID parameter with domain context


// // GetRequestIDFromContext retrieves request ID from context
// func GetRequestIDFromContext(ctx context.Context) string {
// 	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
// 		return reqID
// 	}
// 	return ""
// }


// HELPER FUNCTIONS FOR TYPE CONVERSION
// ============================================================================

// AsAPIError safely converts error to *APIError if possible
func AsAPIError(err error) (*APIError, bool) {
	if err == nil {
		return nil, false
	}
	
	if apiErr, ok := err.(*APIError); ok {
		return apiErr, true
	}
	
	// Try to convert using ConvertToAPIError
	if apiErr := ConvertToAPIError(err); apiErr != nil {
		return apiErr, true
	}
	
	return nil, false
}


func MustAPIError(err error) *APIError {
	if apiErr, ok := AsAPIError(err); ok {
		return apiErr
	}
	panic(fmt.Sprintf("error is not an APIError: %T", err))
}
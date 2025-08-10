// Package errorcustom provides HTTP error handling utilities
package errorcustom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"english-ai-full/logger"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

// ============================================================================
// HTTP MIDDLEWARE
// ============================================================================

// RequestIDMiddleware adds unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		
		// Add request ID to response headers for client debugging
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := withRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// LogHTTPMiddleware logs HTTP requests with optimized logging levels
func LogHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestIDFromContext(r.Context())
		clientIP := GetClientIP(r)
		
		// Only DEBUG level for request initiation
		logger.Debug("HTTP request received", map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"user_agent": r.Header.Get("User-Agent"),
			"ip":         clientIP,
			"request_id": requestID,
		})
		
		// Create a custom ResponseWriter to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process request
		next.ServeHTTP(wrappedWriter, r)
	})
}

// RecoveryMiddleware recovers from panics with ERROR level logging
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestIDFromContext(r.Context())
				
				logger.Error("Panic recovered", map[string]interface{}{
					"error":      err,
					"method":     r.Method,
					"path":       r.URL.Path,
					"ip":         GetClientIP(r),
					"request_id": requestID,
				})
				
				apiErr := NewAPIError(
					ErrCodeInternalError,
					"Internal server error",
					http.StatusInternalServerError,
				)
				HandleError(w, apiErr, requestID)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// CUSTOM RESPONSE WRITER
// ============================================================================

// Custom ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ============================================================================
// ERROR HANDLING
// ============================================================================

// HandleError converts various error types to APIError and responds appropriately
func HandleError(w http.ResponseWriter, err error, requestID string) {
	var apiErr *APIError
	
	// Convert various error types to APIError
	switch e := err.(type) {
	case *APIError:
		apiErr = e
	case *AuthenticationError:
		apiErr = e.ToAPIError()
	case *UserNotFoundError:
		apiErr = e.ToAPIError()
	case *DuplicateEmailError:
		apiErr = e.ToAPIError()
	case *ServiceError:
		apiErr = e.ToAPIError()
	case *RepositoryError:
		apiErr = e.ToAPIError()
	default:
		// Generic error handling
		apiErr = NewAPIError(
			ErrCodeInternalError,
			"An unexpected error occurred",
			http.StatusInternalServerError,
		)
	}

	// Only log ERROR level for critical system failures (5xx errors)
	if apiErr.HTTPStatus >= 500 {
		logContext := apiErr.GetLogContext()
		logContext["request_id"] = requestID
		logger.Error("Critical system error occurred", logContext)
	}

	// Set response headers and status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)

	// Write error response
	response := apiErr.ToErrorResponse()
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode error response", map[string]interface{}{
			"original_error": apiErr.Error(),
			"encoding_error": err.Error(),
			"request_id":     requestID,
		})
		// Fallback to simple text response
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleValidationErrors processes validator validation errors
func HandleValidationErrors(w http.ResponseWriter, validationErrors validator.ValidationErrors, requestID string) {
	errors := make(map[string]string)
	
	for _, err := range validationErrors {
		field := err.Field()
		errors[field] = getValidationMessage(err)
	}

	apiErr := NewAPIError(
		ErrCodeValidationError,
		"Validation failed",
		http.StatusBadRequest,
	).WithDetail("validation_errors", errors)

	HandleError(w, apiErr, requestID)
}

// ============================================================================
// HTTP RESPONSE UTILITIES
// ============================================================================

// RespondWithJSON sends JSON response and logs appropriately
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error":       err.Error(),
			"status_code": statusCode,
			"request_id":  requestID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// RespondWithError sends error response
func RespondWithError(w http.ResponseWriter, status int, message string, requestID string) {
	RespondWithJSON(w, status, map[string]string{"error": message}, requestID)
}

// RespondWithAPIError sends APIError response
func RespondWithAPIError(w http.ResponseWriter, apiErr *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)
	
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		logger.Error("JSON encoding error", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ============================================================================
// REQUEST PARSING UTILITIES
// ============================================================================

// DecodeJSON decodes JSON request body with optional raw request logging
func DecodeJSON(body io.Reader, target interface{}, requestID string, logRawBody bool) error {
	// Read the body first
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return NewAPIError(
			ErrCodeInvalidInput,
			"Failed to read request body",
			http.StatusBadRequest,
		).WithDetail("error", err.Error())
	}

	// Perform JSON decoding
	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return NewAPIError(
			ErrCodeInvalidInput,
			"Invalid JSON format",
			http.StatusBadRequest,
		).WithDetail("error", err.Error())
	}
	
	return nil
}

// ParseIDParam securely parses ID parameter from URL
func ParseIDParam(r *http.Request, paramName string) (int64, *APIError) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		return 0, NewAPIError(
			ErrCodeInvalidInput,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			http.StatusBadRequest,
		)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, NewAPIError(
			ErrCodeInvalidInput,
			fmt.Sprintf("Invalid %s: must be a positive integer", paramName),
			http.StatusBadRequest,
		).WithDetail("value", idStr)
	}

	return id, nil
}

// GetStringParam securely handles string parameters
func GetStringParam(r *http.Request, paramName string, minLen int) (string, *APIError) {
	value := chi.URLParam(r, paramName)
	if value == "" {
		return "", NewAPIError(
			ErrCodeInvalidInput,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			http.StatusBadRequest,
		)
	}

	if minLen > 0 && len(value) < minLen {
		return "", NewAPIError(
			ErrCodeInvalidInput,
			fmt.Sprintf("%s must be at least %d characters", paramName, minLen),
			http.StatusBadRequest,
		)
	}

	for _, r := range value {
		if r < 32 || r == 127 { // Block control characters
			return "", NewAPIError(
				ErrCodeInvalidInput,
				fmt.Sprintf("Invalid characters in %s", paramName),
				http.StatusBadRequest,
			)
		}
	}

	return value, nil
}

// GetPaginationParams safely parses pagination parameters
func GetPaginationParams(r *http.Request) (limit, offset int64, apiErr *APIError) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit = 10
	offset = 0

	if limitStr != "" {
		l, err := strconv.ParseInt(limitStr, 10, 64)
		switch {
		case err != nil:
			return 0, 0, NewAPIError(
				ErrCodeInvalidInput,
				"Invalid limit parameter",
				http.StatusBadRequest,
			)
		case l < 1:
			return 0, 0, NewAPIError(
				ErrCodeInvalidInput,
				"Limit must be at least 1",
				http.StatusBadRequest,
			)
		case l > 100:
			return 0, 0, NewAPIError(
				ErrCodeInvalidInput,
				"Limit cannot exceed 100",
				http.StatusBadRequest,
			)
		default:
			limit = l
		}
	}

	if offsetStr != "" {
		o, err := strconv.ParseInt(offsetStr, 10, 64)
		switch {
		case err != nil:
			return 0, 0, NewAPIError(
				ErrCodeInvalidInput,
				"Invalid offset parameter",
				http.StatusBadRequest,
			)
		case o < 0:
			return 0, 0, NewAPIError(
				ErrCodeInvalidInput,
				"Offset cannot be negative",
				http.StatusBadRequest,
			)
		default:
			offset = o
		}
	}

	return limit, offset, nil
}

// GetSortParams safely parses sorting parameters
func GetSortParams(r *http.Request, allowedFields []string) (sortBy, sortOrder string, apiErr *APIError) {
	sortBy = strings.ToLower(r.URL.Query().Get("sort_by"))
	sortOrder = strings.ToLower(r.URL.Query().Get("sort_order"))

	// Set safe defaults
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		return "", "", NewAPIError(
			ErrCodeInvalidInput,
			"Invalid sort order. Use 'asc' or 'desc'",
			http.StatusBadRequest,
		)
	}

	// Validate field against allowlist
	if len(allowedFields) > 0 {
		valid := false
		for _, field := range allowedFields {
			if sortBy == field {
				valid = true
				break
			}
		}

		if !valid {
			return "", "", NewAPIError(
				ErrCodeInvalidInput,
				fmt.Sprintf("Invalid sort field. Allowed: %v", allowedFields),
				http.StatusBadRequest,
			)
		}
	}

	return sortBy, sortOrder, nil
}

// ============================================================================
// VALIDATION UTILITIES
// ============================================================================

// ValidatePassword performs robust password validation
func ValidatePassword(password string) *APIError {
	const (
		minLength = 8
		maxLength = 128
	)

	if len(password) < minLength {
		return NewAPIError(
			ErrCodeWeakPassword,
			fmt.Sprintf("Password must be at least %d characters", minLength),
			http.StatusBadRequest,
		)
	}

	if len(password) > maxLength {
		return NewAPIError(
			ErrCodeWeakPassword,
			fmt.Sprintf("Password cannot exceed %d characters", maxLength),
			http.StatusBadRequest,
		)
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?/", c):
			hasSpecial = true
		}
	}

	var requirements []string
	if !hasUpper {
		requirements = append(requirements, "uppercase letter")
	}
	if !hasLower {
		requirements = append(requirements, "lowercase letter")
	}
	if !hasDigit {
		requirements = append(requirements, "digit")
	}
	if !hasSpecial {
		requirements = append(requirements, "special character")
	}

	if len(requirements) > 0 {
		return NewAPIError(
			ErrCodeWeakPassword,
			"Password does not meet security requirements",
			http.StatusBadRequest,
		).WithDetail("requirements", requirements)
	}

	return nil
}

// ValidatePasswordWithDetails performs enhanced password validation with logging
func ValidatePasswordWithDetails(password string, requestID string) error {
	requirements := []string{}
	
	if len(password) < 8 {
		requirements = append(requirements, "at least 8 characters long")
	}
	
	if len(password) > 100 {
		requirements = append(requirements, "no more than 100 characters long")
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case char >= 32 && char <= 126:
			if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}
	
	if !hasUpper {
		requirements = append(requirements, "at least one uppercase letter")
	}
	if !hasLower {
		requirements = append(requirements, "at least one lowercase letter")
	}
	if !hasNumber {
		requirements = append(requirements, "at least one number")
	}
	if !hasSpecial {
		requirements = append(requirements, "at least one special character")
	}
	
	if len(requirements) > 0 {
		// Log password validation failure at WARNING level
		logger.Warning("Password validation failed", map[string]interface{}{
			"requirements": requirements,
			"request_id":   requestID,
			"type":         "password_validation",
		})
		
		return &PasswordValidationError{
			Requirements: requirements,
		}
	}
	
	return nil
}

// getValidationMessage returns user-friendly validation error messages
func getValidationMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	case "uuid", "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "numeric", "number":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid date/time (format: %s)", field, param)
	case "password":
		return "Password does not meet requirements"
	case "role":
		return "Invalid role specified"
	case "uniqueemail":
		return "Email already exists"
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// CalculatePagination calculates pagination metadata
func CalculatePagination(total, limit, offset int) (currentPage, totalPages int) {
	currentPage = (offset / limit) + 1
	totalPages = total / limit
	if total%limit > 0 {
		totalPages++
	}
	return currentPage, totalPages
}

// CalculatePaginationBounds calculates safe pagination bounds
func CalculatePaginationBounds(start, end, total int) (int, int) {
	if start >= total {
		start = total
	}
	if end >= total {
		end = total
	}
	return start, end
}

// ============================================================================
// CONTEXT UTILITIES
// ============================================================================

// generateRequestID creates a unique request identifier
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), rand.Int63n(1000))
}

// withRequestID adds request ID to context
func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
}

// GetUserEmailFromContext extracts user email from JWT token context
func GetUserEmailFromContext(r *http.Request) string {
	if email, ok := r.Context().Value("user_email").(string); ok {
		return email
	}
	return ""
}

// GetUserIDFromContext extracts user ID from JWT token context  
func GetUserIDFromContext(r *http.Request) int64 {
	if userID, ok := r.Context().Value("user_id").(int64); ok {
		return userID
	}
	return 0
}

// GetClientIP extracts the real client IP from request
func GetClientIP(r *http.Request) string {
	// Get IP from X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP (client IP)
		return strings.Split(forwarded, ",")[0]
	}
	
	// Get IP from X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// Get IP from request RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}


// new start

// Email validation regex pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// IsValidEmail validates email format using regex
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	
	// Basic length check
	if len(email) > 254 {
		return false
	}
	
	// Check for basic format
	if !strings.Contains(email, "@") {
		return false
	}
	
	// Split into local and domain parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	
	local, domain := parts[0], parts[1]
	
	// Validate local part
	if len(local) == 0 || len(local) > 64 {
		return false
	}
	
	// Validate domain part
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	
	// Use regex for final validation
	return emailRegex.MatchString(email)
}

// ValidateEmail performs comprehensive email validation with detailed error reporting
func ValidateEmail(email string, requestID string) *APIError {
	if email == "" {
		return NewAPIError(
			ErrCodeValidationError,
			"Email is required",
			http.StatusBadRequest,
		).WithDetail("field", "email")
	}
	
	// Trim whitespace
	email = strings.TrimSpace(email)
	
	// Length validation
	if len(email) > 254 {
		return NewAPIError(
			ErrCodeValidationError,
			"Email address is too long (maximum 254 characters)",
			http.StatusBadRequest,
		).WithDetail("field", "email").WithDetail("max_length", 254)
	}
	
	// Basic @ symbol check
	if !strings.Contains(email, "@") {
		return NewAPIError(
			ErrCodeValidationError,
			"Email must contain @ symbol",
			http.StatusBadRequest,
		).WithDetail("field", "email")
	}
	
	// Split validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return NewAPIError(
			ErrCodeValidationError,
			"Email format is invalid",
			http.StatusBadRequest,
		).WithDetail("field", "email")
	}
	
	local, domain := parts[0], parts[1]
	
	// Local part validation
	if len(local) == 0 {
		return NewAPIError(
			ErrCodeValidationError,
			"Email local part cannot be empty",
			http.StatusBadRequest,
		).WithDetail("field", "email")
	}
	
	if len(local) > 64 {
		return NewAPIError(
			ErrCodeValidationError,
			"Email local part is too long (maximum 64 characters)",
			http.StatusBadRequest,
		).WithDetail("field", "email").WithDetail("max_local_length", 64)
	}
	
	// Domain part validation
	if len(domain) == 0 {
		return NewAPIError(
			ErrCodeValidationError,
			"Email domain cannot be empty",
			http.StatusBadRequest,
		).WithDetail("field", "email")
	}
	
	if len(domain) > 253 {
		return NewAPIError(
			ErrCodeValidationError,
			"Email domain is too long (maximum 253 characters)",
			http.StatusBadRequest,
		).WithDetail("field", "email").WithDetail("max_domain_length", 253)
	}
	
	// Final regex validation
	if !emailRegex.MatchString(email) {
		return NewAPIError(
			ErrCodeValidationError,
			"Email format is invalid",
			http.StatusBadRequest,
		).WithDetail("field", "email").WithDetail("expected_format", "user@domain.com")
	}
	
	return nil
}

// Fixed ValidateEmailFormat function for validator with logging
func ValidateEmailFormat(fl validator.FieldLevel) bool {
	fieldName := fl.FieldName()
	email := fl.Field().String()
	
	

	// Use the IsValidEmail function we just created
	isValid := IsValidEmail(email)

	if !isValid {
		// Log validation failure using your existing logger
		logger.Warning("Email format validation failed", map[string]interface{}{
			"field":           fieldName,
			"email":           email,
			"is_valid":        false,
			"expected_format": "user@domain.com",
			"layer":           "validation",
			"operation":       "validate_email_format",
		})
	}

	return isValid
}
// new end
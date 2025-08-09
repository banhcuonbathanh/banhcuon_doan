// internal/error_custom/errors.go
package errorcustom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"


	"english-ai-full/logger"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)








func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}



// new












// utils/error_handler.go


// Optimized validation error handling - single WARNING log
func HandleValidationErrors(w http.ResponseWriter, validationErrors validator.ValidationErrors, requestID string) {
	errors := make(map[string]string)
	
	for _, err := range validationErrors {
		field := err.Field()
		switch err.Tag() {
		case "required":
			errors[field] = "This field is required"
		case "email":
			errors[field] = "Invalid email format"
		case "min":
			errors[field] = field + " is too short"
		case "max":
			errors[field] = field + " is too long"
		case "password":
			errors[field] = "Password does not meet requirements"
		case "role":
			errors[field] = "Invalid role specified"
		case "uniqueemail":
			errors[field] = "Email already exists"
		default:
			errors[field] = "Invalid value"
		}
	}

	// Single aggregated WARNING log for all validation errors


	apiErr := errorcustom.NewAPIError(
		errorcustom.ErrCodeValidationError,
		"Validation failed",
		http.StatusBadRequest,
	).WithDetail("validation_errors", errors)

	HandleError(w, apiErr, requestID)
}

// Helper function to respond with JSON and log the response
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
	
	// Successful responses are logged at INFO level in LogAPIRequest
}

// Enhanced password validation with single error log
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
		
		return &errorcustom.PasswordValidationError{
			Requirements: requirements,
		}
	}
	
	return nil
}







func RespondWithAPIError(w http.ResponseWriter, apiErr *errorcustom.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)
	
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		log.Printf("JSON encoding error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

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
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Secure ID parameter parsing
func ParseIDParam(r *http.Request, paramName string) (int64, *errorcustom.APIError) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			http.StatusBadRequest,
		)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Invalid %s: must be a positive integer", paramName),
			http.StatusBadRequest,
		).WithDetail("value", idStr)
	}

	return id, nil
}

// Secure string parameter handling
func GetStringParam(r *http.Request, paramName string, minLen int) (string, *errorcustom.APIError) {
	value := chi.URLParam(r, paramName)
	if value == "" {
		return "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			http.StatusBadRequest,
		)
	}

	if minLen > 0 && len(value) < minLen {
		return "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			fmt.Sprintf("%s must be at least %d characters", paramName, minLen),
			http.StatusBadRequest,
		)
	}

	for _, r := range value {
		if r < 32 || r == 127 { // Block control characters
			return "", errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				fmt.Sprintf("Invalid characters in %s", paramName),
				http.StatusBadRequest,
			)
		}
	}

	return value, nil
}

// Robust password validation
func ValidatePassword(password string) *errorcustom.APIError {
	const (
		minLength = 8
		maxLength = 128
	)

	if len(password) < minLength {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
			fmt.Sprintf("Password must be at least %d characters", minLength),
			http.StatusBadRequest,
		)
	}

	if len(password) > maxLength {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
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
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeWeakPassword,
			"Password does not meet security requirements",
			http.StatusBadRequest,
		).WithDetail("requirements", requirements)
	}

	return nil
}

// Safe pagination parameters
func GetPaginationParams(r *http.Request) (limit, offset int64, apiErr *errorcustom.APIError) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit = 10
	offset = 0

	if limitStr != "" {
		l, err := strconv.ParseInt(limitStr, 10, 64)
		switch {
		case err != nil:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid limit parameter",
				http.StatusBadRequest,
			)
		case l < 1:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Limit must be at least 1",
				http.StatusBadRequest,
			)
		case l > 100:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
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
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid offset parameter",
				http.StatusBadRequest,
			)
		case o < 0:
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Offset cannot be negative",
				http.StatusBadRequest,
			)
		default:
			offset = o
		}
	}

	return limit, offset, nil
}

// Safe sorting parameters
func GetSortParams(r *http.Request, allowedFields []string) (sortBy, sortOrder string, apiErr *errorcustom.APIError) {
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
		return "", "", errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
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
			return "", "", errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				fmt.Sprintf("Invalid sort field. Allowed: %v", allowedFields),
				http.StatusBadRequest,
			)
		}
	}

	return sortBy, sortOrder, nil
}

// Utility function for pagination metadata
func CalculatePagination(total, limit, offset int) (currentPage, totalPages int) {
	currentPage = (offset / limit) + 1
	totalPages = total / limit
	if total%limit > 0 {
		totalPages++
	}
	return currentPage, totalPages
}



func CalculatePaginationBounds(start, end, total int) (int, int) {
	if start >= total {
		start = total
	}
	if end >= total {
		end = total
	}
	return start, end
}


func RespondWithError(w http.ResponseWriter, status int, message string, requestID string) {
	RespondWithJSON(w, status, map[string]string{"error": message}, requestID)
}




// ytrhfhghjgvjhvhjv



// utils/error_handler.go

// Enhanced error handling with proper logging optimization
func HandleError(w http.ResponseWriter, err error, requestID string) {
	var apiErr *errorcustom.APIError
	
	// Convert various error types to APIError
	switch e := err.(type) {
	case *errorcustom.APIError:
		apiErr = e
	case *errorcustom.AuthenticationError:
		apiErr = e.ToAPIError()
	case *errorcustom.UserNotFoundError:
		apiErr = e.ToAPIError()
	case *errorcustom.DuplicateEmailError:
		apiErr = e.ToAPIError()
	case *errorcustom.ServiceError:
		apiErr = e.ToAPIError()
	case *errorcustom.RepositoryError:
		apiErr = e.ToAPIError()
	default:
		// Generic error handling
		apiErr = errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
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
	// For client errors (4xx), we'll log them in the API request log instead

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




// Enhanced JSON decoding with optional raw request logging
func DecodeJSON(body io.Reader, target interface{}, requestID string, logRawBody bool) error {
	// Read the body first
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Failed to read request body",
			http.StatusBadRequest,
		).WithDetail("error", err.Error())
	}


	// Perform JSON decoding
	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"Invalid JSON format",
			http.StatusBadRequest,
		).WithDetail("error", err.Error())
	}
	
	return nil
}

// Request ID middleware
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		
		// Add request ID to response headers for client debugging
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := r.Context()
		ctx = withRequestID(ctx, requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// Log HTTP middleware with optimized logging levels
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

// Custom ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}


// Recovery middleware with ERROR level logging for panics
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
				
				apiErr := errorcustom.NewAPIError(
					errorcustom.ErrCodeInternalError,
					"Internal server error",
					http.StatusInternalServerError,
				)
				HandleError(w, apiErr, requestID)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Helper functions for request ID management
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), rand.Int63n(1000))
}

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

// Helper function to extract user email from JWT token context
func GetUserEmailFromContext(r *http.Request) string {
	if email, ok := r.Context().Value("user_email").(string); ok {
		return email
	}
	return ""
}

// Helper function to extract user ID from JWT token context  
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


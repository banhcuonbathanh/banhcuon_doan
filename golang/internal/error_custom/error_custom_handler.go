// internal/error_custom/handler.go
// Updated HTTP error handling with domain support
package errorcustom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

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

// DomainMiddleware adds domain context to requests based on route patterns
func DomainMiddleware(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := withDomain(r.Context(), domain)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// LogHTTPMiddleware logs HTTP requests with domain-aware logging
func LogHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestIDFromContext(r.Context())
		domain := GetDomainFromContext(r.Context())
		clientIP := GetClientIP(r)
		
		// Only DEBUG level for request initiation
		logger.Debug("HTTP request received", map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"user_agent": r.Header.Get("User-Agent"),
			"ip":         clientIP,
			"request_id": requestID,
			"domain":     domain,
		})
		
		// Create a custom ResponseWriter to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process request
		next.ServeHTTP(wrappedWriter, r)
	})
}

// RecoveryMiddleware recovers from panics with domain-aware ERROR level logging
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestIDFromContext(r.Context())
				domain := GetDomainFromContext(r.Context())
				
				logger.Error("Panic recovered", map[string]interface{}{
					"error":      err,
					"method":     r.Method,
					"path":       r.URL.Path,
					"ip":         GetClientIP(r),
					"request_id": requestID,
					"domain":     domain,
				})
				
				apiErr := NewAPIError(
					GetSystemErrorCode(domain),
					"Internal server error",
					http.StatusInternalServerError,
				).WithDomain(domain)
				
				HandleError(w, apiErr, requestID)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// CUSTOM RESPONSE WRITER
// ============================================================================


func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ============================================================================
// ERROR HANDLING
// ============================================================================

// HandleError converts various error types to APIError and responds appropriately
func HandleError(w http.ResponseWriter, err error, requestID string) {
	apiErr := ConvertToAPIError(err)
	if apiErr == nil {
		// Fallback for nil errors
		apiErr = NewAPIError(
			ErrorTypeSystem,
			"An unexpected error occurred",
			http.StatusInternalServerError,
		)
	}

	// Determine if we should log this error
	if ShouldLogError(apiErr) {
		severity := GetErrorSeverity(apiErr)
		logContext := apiErr.GetLogContext()
		logContext["request_id"] = requestID

		switch severity {
		case "ERROR":
			logger.Error("Critical system error occurred", logContext)
		case "WARNING":
			logger.Warning("Service error occurred", logContext)
		default:
			logger.Info("Request error occurred", logContext)
		}
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

// HandleDomainError handles domain-specific errors with context
func HandleDomainError(w http.ResponseWriter, err error, domain, requestID string) {
	apiErr := ConvertToAPIError(err)
	if apiErr == nil {
		apiErr = NewAPIError(
			GetSystemErrorCode(domain),
			"An unexpected error occurred",
			http.StatusInternalServerError,
		)
	}

	// Ensure domain is set
	if apiErr.Domain == "" {
		apiErr.WithDomain(domain)
	}

	HandleError(w, apiErr, requestID)
}

// HandleValidationErrors processes validator validation errors with domain context
func HandleValidationErrors(w http.ResponseWriter, validationErrors validator.ValidationErrors, domain, requestID string) {
	errorCollection := NewErrorCollection(domain)
	
	for _, err := range validationErrors {
		field := err.Field()
		message := getValidationMessage(err)
		
		validationErr := NewValidationError(domain, field, message, err.Value())
		errorCollection.Add(validationErr)
	}

	HandleError(w, errorCollection.ToAPIError(), requestID)
}

// HandleMultipleErrors handles multiple errors as a collection
func HandleMultipleErrors(w http.ResponseWriter, errors []error, domain, requestID string) {
	if len(errors) == 0 {
		return
	}

	errorCollection := NewErrorCollection(domain)
	for _, err := range errors {
		errorCollection.Add(err)
	}

	HandleError(w, errorCollection.ToAPIError(), requestID)
}

// ============================================================================
// HTTP RESPONSE UTILITIES
// ============================================================================

// RespondWithJSON sends JSON response with domain context
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

// RespondWithError sends error response with domain context
func RespondWithError(w http.ResponseWriter, status int, message, domain, requestID string) {
	apiErr := NewAPIError(
		GetSystemErrorCode(domain),
		message,
		status,
	).WithDomain(domain)
	
	HandleError(w, apiErr, requestID)
}

// RespondWithDomainSuccess sends successful response with domain metadata
func RespondWithDomainSuccess(w http.ResponseWriter, data interface{}, domain, requestID string) {
	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}
	
	if domain != "" {
		response["domain"] = domain
	}
	
	RespondWithJSON(w, http.StatusOK, response, requestID)
}

// ============================================================================
// REQUEST PARSING UTILITIES WITH DOMAIN SUPPORT
// ============================================================================

// DecodeJSONWithDomain decodes JSON request body with domain context
func DecodeJSONWithDomain(body io.Reader, target interface{}, domain, requestID string) error {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return NewAPIError(
			GetInvalidInputCode(domain),
			"Failed to read request body",
			http.StatusBadRequest,
		).WithDomain(domain).WithDetail("error", err.Error())
	}

	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return NewAPIError(
			GetInvalidInputCode(domain),
			"Invalid JSON format",
			http.StatusBadRequest,
		).WithDomain(domain).WithDetail("error", err.Error())
	}
	
	return nil
}

// ParseIDParamWithDomain securely parses ID parameter with domain context
func ParseIDParamWithDomain(r *http.Request, paramName, domain string) (int64, error) {
	idStr := chi.URLParam(r, paramName)
	if idStr == "" {
		return 0, NewValidationError(
			domain,
			paramName,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			nil,
		)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, NewValidationError(
			domain,
			paramName,
			fmt.Sprintf("Invalid %s: must be a positive integer", paramName),
			idStr,
		)
	}

	return id, nil
}

// GetStringParamWithDomain securely handles string parameters with domain context
func GetStringParamWithDomain(r *http.Request, paramName, domain string, minLen int) (string, error) {
	value := chi.URLParam(r, paramName)
	if value == "" {
		return "", NewValidationError(
			domain,
			paramName,
			fmt.Sprintf("Missing required parameter: %s", paramName),
			nil,
		)
	}

	if minLen > 0 && len(value) < minLen {
		return "", NewValidationError(
			domain,
			paramName,
			fmt.Sprintf("%s must be at least %d characters", paramName, minLen),
			value,
		)
	}

	for _, r := range value {
		if r < 32 || r == 127 {
			return "", NewValidationError(
				domain,
				paramName,
				fmt.Sprintf("Invalid characters in %s", paramName),
				value,
			)
		}
	}

	return value, nil
}

// GetPaginationParamsWithDomain safely parses pagination parameters with domain context
func GetPaginationParamsWithDomain(r *http.Request, domain string) (limit, offset int64, err error) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit = 10
	offset = 0

	if limitStr != "" {
		l, parseErr := strconv.ParseInt(limitStr, 10, 64)
		if parseErr != nil {
			return 0, 0, NewValidationError(domain, "limit", "Invalid limit parameter", limitStr)
		}
		if l < 1 {
			return 0, 0, NewValidationError(domain, "limit", "Limit must be at least 1", l)
		}
		if l > 100 {
			return 0, 0, NewValidationError(domain, "limit", "Limit cannot exceed 100", l)
		}
		limit = l
	}

	if offsetStr != "" {
		o, parseErr := strconv.ParseInt(offsetStr, 10, 64)
		if parseErr != nil {
			return 0, 0, NewValidationError(domain, "offset", "Invalid offset parameter", offsetStr)
		}
		if o < 0 {
			return 0, 0, NewValidationError(domain, "offset", "Offset cannot be negative", o)
		}
		offset = o
	}

	return limit, offset, nil
}

// GetSortParamsWithDomain safely parses sorting parameters with domain context
func GetSortParamsWithDomain(r *http.Request, allowedFields []string, domain string) (sortBy, sortOrder string, err error) {
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
		return "", "", NewValidationError(
			domain,
			"sort_order",
			"Invalid sort order. Use 'asc' or 'desc'",
			sortOrder,
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
			return "", "", NewValidationErrorWithRules(
				domain,
				"sort_by",
				fmt.Sprintf("Invalid sort field. Allowed: %v", allowedFields),
				sortBy,
				map[string]interface{}{
					"allowed_fields": allowedFields,
				},
			)
		}
	}

	return sortBy, sortOrder, nil
}

// ============================================================================
// DOMAIN-AWARE VALIDATION UTILITIES
// ============================================================================

// ValidatePasswordWithDomain performs enhanced password validation with domain context
func ValidatePasswordWithDomain(password, domain, requestID string) error {
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
		logger.Warning("Password validation failed", map[string]interface{}{
			"requirements": requirements,
			"request_id":   requestID,
			"domain":       domain,
			"type":         "password_validation",
		})
		
		return NewWeakPasswordError(requirements)
	}
	
	return nil
}

// ValidateEmailWithDomain performs comprehensive email validation with domain context
func ValidateEmailWithDomain(email, domain, requestID string) error {
	if email == "" {
		return NewValidationError(domain, "email", "Email is required", nil)
	}
	
	// Trim whitespace
	email = strings.TrimSpace(email)
	
	// Length validation
	if len(email) > 254 {
		return NewValidationErrorWithRules(
			domain,
			"email",
			"Email address is too long (maximum 254 characters)",
			email,
			map[string]interface{}{"max_length": 254},
		)
	}
	
	// Basic @ symbol check
	if !strings.Contains(email, "@") {
		return NewValidationError(domain, "email", "Email must contain @ symbol", email)
	}
	
	// Split validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return NewValidationError(domain, "email", "Email format is invalid", email)
	}
	
	local, domainPart := parts[0], parts[1]
	
	// Local part validation
	if len(local) == 0 {
		return NewValidationError(domain, "email", "Email local part cannot be empty", email)
	}
	
	if len(local) > 64 {
		return NewValidationErrorWithRules(
			domain,
			"email",
			"Email local part is too long (maximum 64 characters)",
			email,
			map[string]interface{}{"max_local_length": 64},
		)
	}
	
	// Domain part validation
	if len(domainPart) == 0 {
		return NewValidationError(domain, "email", "Email domain cannot be empty", email)
	}
	
	if len(domainPart) > 253 {
		return NewValidationErrorWithRules(
			domain,
			"email",
			"Email domain is too long (maximum 253 characters)",
			email,
			map[string]interface{}{"max_domain_length": 253},
		)
	}
	
	// Final regex validation
	if !emailRegex.MatchString(email) {
		return NewValidationErrorWithRules(
			domain,
			"email",
			"Email format is invalid",
			email,
			map[string]interface{}{"expected_format": "user@domain.com"},
		)
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
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// ============================================================================
// CONTEXT UTILITIES
// ============================================================================


// withRequestID adds request ID to context
func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

// withDomain adds domain to context
func withDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, "domain", domain)
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
// func GetClientIP(r *http.Request) string {
// 	// Get IP from X-Forwarded-For header
// 	forwarded := r.Header.Get("X-Forwarded-For")
// 	if forwarded != "" {
// 		// Take the first IP (client IP)
// 		return strings.Split(forwarded, ",")[0]
// 	}
	
// 	// Get IP from X-Real-IP header
// 	realIP := r.Header.Get("X-Real-IP")
// 	if realIP != "" {
// 		return realIP
// 	}
	
// 	// Get IP from request RemoteAddr
// 	ip, _, err := net.SplitHostPort(r.RemoteAddr)
// 	if err != nil {
// 		return r.RemoteAddr
// 	}
// 	return ip
// }

// ============================================================================
// EMAIL VALIDATION (maintained for compatibility)
// ============================================================================

// Email validation regex pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
// internal/error_custom/handler.go
// Updated HTTP error handling with domain support


// ============================================================================
// HTTP MIDDLEWARE
// ============================================================================









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

// ValidateEmailFormat function for validator with domain-aware logging
func ValidateEmailFormat(fl validator.FieldLevel) bool {
	fieldName := fl.FieldName()
	email := fl.Field().String()
	
	// Use the IsValidEmail function
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

// Add this function to your golang/internal/error_custom/error_custom_handler.go file

// DebugMiddleware provides detailed request/response logging for development
func DebugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestIDFromContext(r.Context())
		domain := GetDomainFromContext(r.Context())
		clientIP := GetClientIP(r)
		
		// Log incoming request details
		logger.Debug("Debug: Incoming request details", map[string]interface{}{
			"method":       r.Method,
			"path":         r.URL.Path,
			"query":        r.URL.RawQuery,
			"headers":      r.Header,
			"user_agent":   r.Header.Get("User-Agent"),
			"content_type": r.Header.Get("Content-Type"),
			"ip":           clientIP,
			"request_id":   requestID,
			"domain":       domain,
		})
		
		// Create a custom ResponseWriter to capture response details
		debugWriter := &debugResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			responseSize:   0,
		}
		
		// Record start time
		startTime := time.Now()
		
		// Process the request
		next.ServeHTTP(debugWriter, r)
		
		// Calculate processing time
		duration := time.Since(startTime)
		
		// Log response details
		logger.Debug("Debug: Response details", map[string]interface{}{
			"status_code":    debugWriter.statusCode,
			"response_size":  debugWriter.responseSize,
			"duration_ms":    duration.Milliseconds(),
			"duration_ns":    duration.Nanoseconds(),
			"request_id":     requestID,
			"domain":         domain,
		})
	})
}

// debugResponseWriter wraps http.ResponseWriter to capture response details
type debugResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func (drw *debugResponseWriter) WriteHeader(code int) {
	drw.statusCode = code
	drw.ResponseWriter.WriteHeader(code)
}

func (drw *debugResponseWriter) Write(data []byte) (int, error) {
	size, err := drw.ResponseWriter.Write(data)
	drw.responseSize += size
	return size, err
}


// Add this function to your golang/internal/error_custom/error_custom_handler.go file

// DomainContextMiddleware automatically detects and sets domain context based on URL patterns
func DomainContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := detectDomainFromPath(r.URL.Path)
		
		// Add domain to context
		ctx := withDomain(r.Context(), domain)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// detectDomainFromPath automatically detects domain from URL path
func detectDomainFromPath(path string) string {
	path = strings.ToLower(path)
	
	// API route patterns
	switch {
	case strings.Contains(path, "/api/accounts") || strings.Contains(path, "/api/users") || strings.Contains(path, "/api/auth"):
		return DomainAccount
	case strings.Contains(path, "/api/branches"):
		return "branch"
	case strings.Contains(path, "/api/courses"):
		return DomainCourse
	case strings.Contains(path, "/api/payments"):
		return DomainPayment
	case strings.Contains(path, "/api/content"):
		return DomainContent
	case strings.Contains(path, "/api/admin"):
		return DomainAdmin
	case strings.Contains(path, "/swagger") || strings.Contains(path, "/health") || strings.Contains(path, "/metrics"):
		return DomainSystem
	default:
		return DomainSystem // Default fallback
	}
}

// Add this function to your golang/internal/error_custom/error_custom_handler.go file



// JWTValidationMiddleware validates JWT tokens and adds user context
func JWTValidationMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip JWT validation for certain routes
			if shouldSkipJWTValidation(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				domain := GetDomainFromContext(r.Context())
				requestID := GetRequestIDFromContext(r.Context())
				
				authErr := NewAuthenticationError(domain, "Missing authorization token")
				HandleError(w, authErr.ToAPIError(), requestID)
				return
			}

			// Check Bearer token format
			tokenParts := strings.SplitN(authHeader, " ", 2)
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				domain := GetDomainFromContext(r.Context())
				requestID := GetRequestIDFromContext(r.Context())
				
				authErr := NewAuthenticationError(domain, "Invalid authorization header format")
				HandleError(w, authErr.ToAPIError(), requestID)
				return
			}

			tokenString := tokenParts[1]

			// Parse and validate the JWT token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			})

			if err != nil {
				domain := GetDomainFromContext(r.Context())
				requestID := GetRequestIDFromContext(r.Context())
				
				authErr := NewAuthenticationError(domain, "Invalid or expired token")
				HandleError(w, authErr.ToAPIError().WithCause(err), requestID)
				return
			}

			// Check if token is valid and get claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Add user information to context
				ctx := r.Context()
				
				if userID, ok := claims["user_id"]; ok {
					ctx = context.WithValue(ctx, "user_id", int64(userID.(float64)))
				}
				
				if email, ok := claims["email"].(string); ok {
					ctx = context.WithValue(ctx, "user_email", email)
				}
				
				if role, ok := claims["role"].(string); ok {
					ctx = context.WithValue(ctx, "user_role", role)
				}

				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			} else {
				domain := GetDomainFromContext(r.Context())
				requestID := GetRequestIDFromContext(r.Context())
				
				authErr := NewAuthenticationError(domain, "Invalid token claims")
				HandleError(w, authErr.ToAPIError(), requestID)
				return
			}
		})
	}
}

// shouldSkipJWTValidation determines if JWT validation should be skipped for certain routes
func shouldSkipJWTValidation(path string) bool {
	skipRoutes := []string{
		"/api/accounts/login",
		"/api/accounts/register", 
		"/api/auth/login",
		"/api/auth/register",
		"/swagger",
		"/health",
		"/metrics",
		"/ping",
	}
	
	path = strings.ToLower(path)
	
	for _, skipRoute := range skipRoutes {
		if strings.HasPrefix(path, skipRoute) {
			return true
		}
	}
	
	return false
}

// LogCriticalError logs critical system errors with enhanced context
func LogCriticalError(errorType string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"error_type": errorType,
		"severity":   "CRITICAL",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
	
	// Merge provided context
	for key, value := range context {
		logContext[key] = value
	}
	
	logger.Error("Critical system error", logContext)
}


// Add this function to your golang/internal/error_custom/error_custom_handler.go file


// RateLimiter holds rate limiting data for a domain
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

var (
	rateLimiters = make(map[string]*RateLimiter)
	rateLimiterMutex sync.RWMutex
)

// RateLimitMiddleware provides rate limiting per domain and IP
func RateLimitMiddleware(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := GetClientIP(r)
			key := fmt.Sprintf("%s:%s", domain, clientIP)
			
			// Get or create rate limiter for this domain
			rateLimiterMutex.RLock()
			limiter, exists := rateLimiters[domain]
			rateLimiterMutex.RUnlock()
			
			if !exists {
				rateLimiterMutex.Lock()
				if limiter, exists = rateLimiters[domain]; !exists {
					limiter = &RateLimiter{
						requests: make(map[string][]time.Time),
						limit:    100, // 100 requests per minute by default
						window:   time.Minute,
					}
					rateLimiters[domain] = limiter
				}
				rateLimiterMutex.Unlock()
			}
			
			// Check rate limit
			now := time.Now()
			limiter.mutex.Lock()
			
			// Clean old requests outside the time window
			if requests, exists := limiter.requests[key]; exists {
				validRequests := make([]time.Time, 0, len(requests))
				for _, reqTime := range requests {
					if now.Sub(reqTime) < limiter.window {
						validRequests = append(validRequests, reqTime)
					}
				}
				limiter.requests[key] = validRequests
			}
			
			// Check if limit exceeded
			if len(limiter.requests[key]) >= limiter.limit {
				limiter.mutex.Unlock()
				
				requestID := GetRequestIDFromContext(r.Context())
				rateLimitErr := NewAPIError(
					GetRateLimitCode(domain),
					"Rate limit exceeded. Please try again later.",
					http.StatusTooManyRequests,
				).WithDomain(domain).
					WithDetail("limit", limiter.limit).
					WithDetail("window", limiter.window.String()).
					WithDetail("client_ip", clientIP)
				
				HandleError(w, rateLimitErr, requestID)
				return
			}
			
			// Add current request
			limiter.requests[key] = append(limiter.requests[key], now)
			limiter.mutex.Unlock()
			
			next.ServeHTTP(w, r)
		})
	}
}

// In your errorcustom or utils package, add:
func ValidatePassword(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("password must be at least 8 characters long")
    }
    // Add more validation rules as needed
    return nil
}
// internal/error_custom/error_custom_handler.go
// Cleaned HTTP error handling with domain support
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

	"github.com/go-playground/validator/v10"
)

// ============================================================================
// MIDDLEWARE
// ============================================================================

// RequestIDMiddleware adds unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		w.Header().Set("X-Request-ID", requestID)
		ctx := withRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// DomainContextMiddleware automatically detects and sets domain context
func DomainContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := detectDomainFromPath(r.URL.Path)
		ctx := withDomain(r.Context(), domain)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LogHTTPMiddleware logs HTTP requests with domain-aware logging
func LogHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestIDFromContext(r.Context())
		domain := GetDomainFromContext(r.Context())
		
		logger.Debug("HTTP request received", map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"ip":         GetClientIP(r),
			"request_id": requestID,
			"domain":     domain,
		})
		
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics with domain-aware logging
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

// JWTValidationMiddleware validates JWT tokens and adds user context
func JWTValidationMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldSkipJWTValidation(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handleAuthError(w, r, "Missing authorization token")
				return
			}

			tokenParts := strings.SplitN(authHeader, " ", 2)
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				handleAuthError(w, r, "Invalid authorization header format")
				return
			}

			token, err := jwt.Parse(tokenParts[1], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			})

			if err != nil {
				handleAuthError(w, r, "Invalid or expired token")
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				ctx := addUserToContext(r.Context(), claims)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				handleAuthError(w, r, "Invalid token claims")
			}
		})
	}
}

// RateLimitMiddleware provides rate limiting per domain and IP
func RateLimitMiddleware(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := GetClientIP(r)
			if isRateLimited(domain, clientIP) {
				requestID := GetRequestIDFromContext(r.Context())
				rateLimitErr := NewAPIError(
					GetRateLimitCode(domain),
					"Rate limit exceeded. Please try again later.",
					http.StatusTooManyRequests,
				).WithDomain(domain)
				
				HandleError(w, rateLimitErr, requestID)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// ERROR HANDLING
// ============================================================================

// HandleError converts various error types to APIError and responds appropriately
func HandleError(w http.ResponseWriter, err error, requestID string) {
	apiErr := ConvertToAPIError(err)
	if apiErr == nil {
		apiErr = NewAPIError(
			ErrorTypeSystem,
			"An unexpected error occurred",
			http.StatusInternalServerError,
		)
	}

	if ShouldLogError(apiErr) {
		logError(apiErr, requestID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus)

	response := apiErr.ToErrorResponse()
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode error response", map[string]interface{}{
			"original_error": apiErr.Error(),
			"encoding_error": err.Error(),
			"request_id":     requestID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleValidationErrors processes validator validation errors
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

// ============================================================================
// REQUEST PARSING UTILITIES
// ============================================================================

// DecodeJSON decodes JSON request body with error handling
func DecodeJSON(body io.Reader, target interface{}, domain, requestID string) error {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return NewAPIError(
			GetInvalidInputCode(domain),
			"Failed to read request body",
			http.StatusBadRequest,
		).WithDomain(domain)
	}

	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return handleJSONError(err, domain)
	}
	
	return nil
}
// new 1212121


func GetPaginationParams(r *http.Request, domain string) (limit, offset int64, err error) {
	limit = 10  // default
	offset = 0  // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, parseErr := strconv.ParseInt(limitStr, 10, 64); parseErr != nil {
			return 0, 0, NewValidationError(domain, "limit", "Invalid limit parameter", limitStr)
		} else if l < 1 || l > 100 {
			return 0, 0, NewValidationError(domain, "limit", "Limit must be between 1 and 100", l)
		} else {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, parseErr := strconv.ParseInt(offsetStr, 10, 64); parseErr != nil {
			return 0, 0, NewValidationError(domain, "offset", "Invalid offset parameter", offsetStr)
		} else if o < 0 {
			return 0, 0, NewValidationError(domain, "offset", "Offset cannot be negative", o)
		} else {
			offset = o
		}
	}

	return limit, offset, nil
}



var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail performs comprehensive email validation
func ValidateEmail(email, domain string) error {
	if email == "" {
		return NewValidationError(domain, "email", "Email is required", nil)
	}
	
	email = strings.TrimSpace(email)
	
	if len(email) > 254 {
		return NewValidationError(domain, "email", "Email address is too long (maximum 254 characters)", email)
	}
	
	if !emailRegex.MatchString(email) {
		return NewValidationError(domain, "email", "Email format is invalid", email)
	}
	
	return nil
}

// ValidatePassword performs password validation
func ValidatePassword(password, domain string) error {
	if len(password) < 8 {
		return NewValidationError(domain, "password", "Password must be at least 8 characters long", nil)
	}
	
	if len(password) > 100 {
		return NewValidationError(domain, "password", "Password must be no more than 100 characters long", nil)
	}
	
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasNumber := strings.ContainsAny(password, "0123456789")
	hasSpecial := strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?")
	
	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return NewValidationError(domain, "password", 
			"Password must contain uppercase, lowercase, number, and special character", nil)
	}
	
	return nil
}



// RespondWithJSON sends JSON response
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error":       err.Error(),
			"request_id":  requestID,
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// RespondWithSuccess sends successful response
func RespondWithSuccess(w http.ResponseWriter, data interface{}, domain, requestID string) {
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
// CONTEXT UTILITIES
// ============================================================================

func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, "request_id", requestID)
}

func withDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, "domain", domain)
}





func GetUserEmailFromContext(r *http.Request) string {
	if email, ok := r.Context().Value("user_email").(string); ok {
		return email
	}
	return ""
}

func GetUserIDFromContext(r *http.Request) int64 {
	if userID, ok := r.Context().Value("user_id").(int64); ok {
		return userID
	}
	return 0
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func detectDomainFromPath(path string) string {
	path = strings.ToLower(path)
	
	switch {
	case strings.Contains(path, "/api/accounts") || strings.Contains(path, "/api/auth"):
		return DomainAccount
	case strings.Contains(path, "/api/branches"):
		return "branch"
	case strings.Contains(path, "/api/admin"):
		return DomainAdmin
	default:
		return DomainSystem
	}
}

func shouldSkipJWTValidation(path string) bool {
	skipRoutes := []string{
		"/api/accounts/login",
		"/api/accounts/register", 
		"/api/auth/login",
		"/api/auth/register",
		"/swagger",
		"/health",
		"/metrics",
	}
	
	path = strings.ToLower(path)
	for _, skipRoute := range skipRoutes {
		if strings.HasPrefix(path, skipRoute) {
			return true
		}
	}
	return false
}

func handleAuthError(w http.ResponseWriter, r *http.Request, message string) {
	domain := GetDomainFromContext(r.Context())
	requestID := GetRequestIDFromContext(r.Context())
	authErr := NewAuthenticationError(domain, message)
	HandleError(w, authErr.ToAPIError(), requestID)
}

func addUserToContext(ctx context.Context, claims jwt.MapClaims) context.Context {
	if userID, ok := claims["user_id"]; ok {
		ctx = context.WithValue(ctx, "user_id", int64(userID.(float64)))
	}
	if email, ok := claims["email"].(string); ok {
		ctx = context.WithValue(ctx, "user_email", email)
	}
	if role, ok := claims["role"].(string); ok {
		ctx = context.WithValue(ctx, "user_role", role)
	}
	return ctx
}

func handleJSONError(err error, domain string) error {
	switch {
	case err.Error() == "unexpected end of JSON input":
		return NewAPIError(GetInvalidInputCode(domain), "Request body cannot be empty", http.StatusBadRequest).WithDomain(domain)
	case strings.HasPrefix(err.Error(), "invalid character"):
		return NewAPIError(GetInvalidInputCode(domain), "Invalid JSON syntax", http.StatusBadRequest).WithDomain(domain)
	default:
		return NewAPIError(GetInvalidInputCode(domain), "Invalid JSON format", http.StatusBadRequest).WithDomain(domain)
	}
}

func logError(apiErr *APIError, requestID string) {
	severity := GetErrorSeverity(apiErr)
	logContext := apiErr.GetLogContext()
	logContext["request_id"] = requestID

	switch severity {
	case "ERROR":
		logger.Error("Critical system error occurred", logContext)
	case "WARNING":
		logger.Warn("Service error occurred", logContext)
	default:
		logger.Info("Request error occurred", logContext)
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
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Rate limiting implementation (simplified)
var (
	rateLimiters = make(map[string]*RateLimiter)
	rateLimiterMutex sync.RWMutex
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func isRateLimited(domain, clientIP string) bool {
	// Simplified rate limiting logic
	// Implementation details depend on your specific requirements
	return false // Placeholder
}
// 


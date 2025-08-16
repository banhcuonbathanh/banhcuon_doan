// ============================================================================
// FILE: golang/internal/error_custom/middleware.go
// ============================================================================
package errorcustom

import (
	"context"
	"english-ai-full/logger"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// ErrorMiddleware provides enhanced error handling middleware
type ErrorMiddleware struct {
	errorFactory *ErrorFactory
}

// NewErrorMiddleware creates a new error middleware
func NewErrorMiddleware() *ErrorMiddleware {
	return &ErrorMiddleware{
		errorFactory: NewErrorFactory(),
	}
}

// RequestIDMiddleware adds unique request ID to each request
func (em *ErrorMiddleware) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		
		// Add request ID to response headers for client debugging
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// DomainMiddleware adds domain context to requests based on route patterns
func (em *ErrorMiddleware) DomainMiddleware(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "domain", domain)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// AutoDomainMiddleware automatically detects domain from URL path
func (em *ErrorMiddleware) AutoDomainMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := em.detectDomainFromPath(r.URL.Path)
		ctx := context.WithValue(r.Context(), "domain", domain)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics with domain-aware error handling
func (em *ErrorMiddleware) RecoveryMiddleware(next http.Handler) http.Handler {
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
					"layer":      "middleware",
				})
				
				apiErr := NewAPIError(
					GetSystemErrorCode(domain),
					"Internal server error",
					http.StatusInternalServerError,
				).WithDomain(domain).
					WithLayer("middleware")
				
				em.errorFactory.HandlerErrorMgr.RespondWithError(w, apiErr, domain, requestID)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware provides enhanced request/response logging
func (em *ErrorMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestIDFromContext(r.Context())
		domain := GetDomainFromContext(r.Context())
		clientIP := GetClientIP(r)
		startTime := time.Now()
		
		// Log incoming request
		logger.Debug("HTTP request received", map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"user_agent": r.Header.Get("User-Agent"),
			"ip":         clientIP,
			"request_id": requestID,
			"domain":     domain,
		})
		
		// Create response writer wrapper to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process request
		next.ServeHTTP(wrappedWriter, r)
		
		// Log response
		duration := time.Since(startTime)
		logger.Debug("HTTP request completed", map[string]interface{}{
			"status_code": wrappedWriter.statusCode,
			"duration_ms": duration.Milliseconds(),
			"request_id":  requestID,
			"domain":      domain,
		})
	})
}

// detectDomainFromPath automatically detects domain from URL path
func (em *ErrorMiddleware) detectDomainFromPath(path string) string {
	path = strings.ToLower(path)
	
	switch {
	case strings.Contains(path, "/api/accounts") || strings.Contains(path, "/api/users"):
		return DomainAccount
	case strings.Contains(path, "/api/auth"):
		return DomainAuth
	case strings.Contains(path, "/api/branches"):
		return "branch"



	case strings.Contains(path, "/api/admin"):
		return DomainAdmin
	default:
		return DomainSystem
	}
}

// generateRequestID creates a unique request identifier
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), rand.Int63n(1000))
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
}

// GetDomainFromContext extracts domain from context
func GetDomainFromContext(ctx context.Context) string {
	if domain, ok := ctx.Value("domain").(string); ok {
		return domain
	}
	return DomainSystem
}

// GetClientIP extracts the real client IP from request
func GetClientIP(r *http.Request) string {
	// Get IP from X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	
	// Get IP from X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// Get IP from request RemoteAddr
	return r.RemoteAddr
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}


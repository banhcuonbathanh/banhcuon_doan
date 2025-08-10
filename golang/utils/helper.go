package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// MergeContext combines two context maps, with additional context overriding base context
// for any duplicate keys. This is useful for creating rich logging contexts.
//
// Parameters:
//   - base: The base context map (typically contains common fields like request_id, user_id, etc.)
//   - additional: Additional context to merge (specific to the current operation)
//
// Returns:
//   - A new map containing all key-value pairs from both maps
//
// Example:
//   base := map[string]interface{}{
//       "user_id": "123",
//       "request_id": "req_456",
//   }
//   additional := map[string]interface{}{
//       "error": "validation failed",
//       "field": "email",
//   }
//   merged := MergeContext(base, additional)
//   // Result: {"user_id": "123", "request_id": "req_456", "error": "validation failed", "field": "email"}
func MergeContext(base map[string]interface{}, additional map[string]interface{}) map[string]interface{} {
	// Handle nil cases
	if base == nil && additional == nil {
		return make(map[string]interface{})
	}
	if base == nil {
		// Create a copy of additional to avoid modifying the original
		result := make(map[string]interface{}, len(additional))
		for k, v := range additional {
			result[k] = v
		}
		return result
	}
	if additional == nil {
		// Create a copy of base to avoid modifying the original
		result := make(map[string]interface{}, len(base))
		for k, v := range base {
			result[k] = v
		}
		return result
	}

	// Create new map with capacity for both maps
	merged := make(map[string]interface{}, len(base)+len(additional))
	
	// Copy base context first
	for k, v := range base {
		merged[k] = v
	}
	
	// Add additional context (will override base values for duplicate keys)
	for k, v := range additional {
		merged[k] = v
	}
	
	return merged
}

// MergeMultipleContexts merges multiple context maps into one.
// Later contexts override earlier ones for duplicate keys.
//
// Parameters:
//   - contexts: Variable number of context maps to merge
//
// Returns:
//   - A new map containing all key-value pairs from all input maps
//
// Example:
//   ctx1 := map[string]interface{}{"user_id": "123"}
//   ctx2 := map[string]interface{}{"request_id": "req_456"}
//   ctx3 := map[string]interface{}{"operation": "update"}
//   merged := MergeMultipleContexts(ctx1, ctx2, ctx3)
func MergeMultipleContexts(contexts ...map[string]interface{}) map[string]interface{} {
	if len(contexts) == 0 {
		return make(map[string]interface{})
	}
	
	// Calculate total capacity needed
	totalSize := 0
	for _, ctx := range contexts {
		if ctx != nil {
			totalSize += len(ctx)
		}
	}
	
	merged := make(map[string]interface{}, totalSize)
	
	// Merge all contexts in order (later contexts override earlier ones)

	return merged
}

// AddToContext adds a single key-value pair to an existing context map.
// If the context is nil, creates a new map.
//
// Parameters:
//   - context: The existing context map (can be nil)
//   - key: The key to add
//   - value: The value to add
//
// Returns:
//   - A new map with the added key-value pair
//
// Example:
//   ctx := map[string]interface{}{"user_id": "123"}
//   newCtx := AddToContext(ctx, "error", "validation failed")
//   // Result: {"user_id": "123", "error": "validation failed"}
func AddToContext(context map[string]interface{}, key string, value interface{}) map[string]interface{} {
	if context == nil {
		return map[string]interface{}{key: value}
	}
	
	// Create a copy to avoid modifying the original
	result := make(map[string]interface{}, len(context)+1)
	for k, v := range context {
		result[k] = v
	}
	result[key] = value
	
	return result
}

// CreateBaseContext creates a base context map with common fields for HTTP requests.
// This is typically used at the start of request handlers.
//
// Parameters:
//   - r: HTTP request object
//   - additionalFields: Optional additional fields to include
//
// Returns:
//   - A context map with common request fields
//
// Example:
//   baseCtx := CreateBaseContext(r, map[string]interface{}{
//       "user_id": "123",
//       "tenant_id": "org_456",
//   })
func CreateBaseContext(r *http.Request, additionalFields map[string]interface{}) map[string]interface{} {
	context := map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"ip":         r.RemoteAddr,
		"user_agent": r.UserAgent(),
	}
	
	// Add query parameters if present
	if len(r.URL.RawQuery) > 0 {
		context["query"] = r.URL.RawQuery
	}
	

	
	return context
}




// GetTypeName returns the type name of an interface{} value
// This is useful for logging the type of requests being processed
//
// Parameters:
//   - obj: The object to get the type name for
//
// Returns:
//   - A string representation of the type name
//
// Example:
//   type User struct { Name string }
//   user := User{Name: "John"}
//   typeName := GetTypeName(user) // Returns "User"
func GetTypeName(obj interface{}) string {
	if obj == nil {
		return "nil"
	}
	
	t := reflect.TypeOf(obj)
	
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	// Return the type name
	if t.PkgPath() != "" {
		// Return package.TypeName for types from other packages
		return t.PkgPath() + "." + t.Name()
	}
	
	return t.Name()
}

// MaskSensitiveValue masks sensitive field values for logging
// This prevents sensitive data from being logged in plain text
//
// Parameters:
//   - fieldName: The name of the field being logged
//   - value: The value to potentially mask
//
// Returns:
//   - The original value if not sensitive, or a masked version if sensitive
//
// Example:
//   masked := MaskSensitiveValue("password", "secret123") // Returns "***masked***"
//   normal := MaskSensitiveValue("name", "John")          // Returns "John"
func MaskSensitiveValue(fieldName string, value interface{}) interface{} {
	if value == nil {
		return nil
	}
	
	fieldLower := strings.ToLower(fieldName)
	
	// List of sensitive field patterns
	sensitivePatterns := []string{
		"password",
		"token",
		"secret",
		"key",
		"credential",
		"auth",
		"ssn",
		"social",
		"credit",
		"card",
		"cvv",
		"pin",
		"otp",
		"code", // For verification codes
	}
	
	// Check if field name contains sensitive patterns
	for _, pattern := range sensitivePatterns {
		if strings.Contains(fieldLower, pattern) {
			// For strings, show partial information
			if str, ok := value.(string); ok {
				if len(str) == 0 {
					return "***empty***"
				}
				if len(str) <= 3 {
					return "***masked***"
				}
				// Show first and last character for longer strings
				return string(str[0]) + "***" + string(str[len(str)-1])
			}
			
			// For other types, just mask completely
			return "***masked***"
		}
	}
	
	// Special handling for email addresses
	if fieldLower == "email" || strings.Contains(fieldLower, "email") {
		return maskEmail(value)
	}
	
	// Return original value if not sensitive
	return value
}

// maskEmail masks email addresses for logging while preserving some structure
// This is a helper function used by MaskSensitiveValue
//
// Parameters:
//   - value: The email value to mask
//
// Returns:
//   - A masked version of the email address
//
// Example:
//   masked := maskEmail("user@example.com") // Returns "u***r@example.com"
func maskEmail(value interface{}) interface{} {
	str, ok := value.(string)
	if !ok {
		return value
	}
	
	if len(str) == 0 {
		return str
	}
	
	// Check if it looks like an email
	if !strings.Contains(str, "@") {
		return value
	}
	
	parts := strings.Split(str, "@")
	if len(parts) != 2 {
		return value
	}
	
	username := parts[0]
	domain := parts[1]
	
	// Mask username part
	if len(username) <= 2 {
		return "***@" + domain
	}
	
	maskedUsername := string(username[0]) + "***" + string(username[len(username)-1])
	return maskedUsername + "@" + domain
}

// GetRequestID extracts or generates a request ID for logging correlation
// This helps in tracing requests across different components
//
// Parameters:
//   - r: HTTP request object
//
// Returns:
//   - A request ID string
//
// Example:
//   requestID := GetRequestID(r) // Returns existing ID or generates new one
func GetRequestID(r *http.Request) string {
	// Check common request ID headers
	requestIDHeaders := []string{
		"X-Request-ID",
		"X-Request-Id",
		"X-Correlation-ID",
		"X-Correlation-Id",
		"Request-ID",
		"Request-Id",
	}
	
	for _, header := range requestIDHeaders {
		if id := r.Header.Get(header); id != "" {
			return id
		}
	}
	
	// Check context
	if id := r.Context().Value("request_id"); id != nil {
		if idStr, ok := id.(string); ok {
			return idStr
		}
	}
	
	// Generate new UUID-like ID if none found
	return generateSimpleID()
}

// generateSimpleID generates a simple unique identifier
// This is a basic implementation - in production, you might want to use a proper UUID library
func generateSimpleID() string {
	// This is a simplified implementation
	// In production, use a proper UUID library like github.com/google/uuid
	return "req_" + randomString(12)
}

// randomString generates a random string of specified length
// This is a helper function for generateSimpleID
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// ValidateAndSanitizeContext validates and sanitizes context data for logging
// This ensures context data is safe and appropriate for logging
//
// Parameters:
//   - context: The context map to validate and sanitize
//
// Returns:
//   - A sanitized version of the context map
//
// Example:
//   sanitized := ValidateAndSanitizeContext(map[string]interface{}{
//       "user_id": 123,
//       "password": "secret",
//   })
//   // Returns: {"user_id": 123, "password": "***masked***"}
func ValidateAndSanitizeContext(context map[string]interface{}) map[string]interface{} {
	if context == nil {
		return make(map[string]interface{})
	}
	
	sanitized := make(map[string]interface{}, len(context))
	
	for key, value := range context {
		// Sanitize the value based on the key
		sanitized[key] = MaskSensitiveValue(key, value)
	}
	
	return sanitized
}

// FormatDuration formats a time.Duration for human-readable logging
// This provides consistent duration formatting across the application
//
// Parameters:
//   - duration: The duration to format
//
// Returns:
//   - A human-readable string representation of the duration
//
// Example:
//   formatted := FormatDuration(1500 * time.Millisecond) // Returns "1.50s"
func FormatDuration(duration time.Duration) string {
	if duration < time.Microsecond {
		return duration.String()
	}
	
	if duration < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(duration.Nanoseconds())/1000.0)
	}
	
	if duration < time.Second {
		return fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1000000.0)
	}
	
	if duration < time.Minute {
		return fmt.Sprintf("%.2fs", duration.Seconds())
	}
	
	return duration.String()
}
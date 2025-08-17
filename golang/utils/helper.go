package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"
	"unicode"
)

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



// ContainsUppercase checks if the string contains at least one uppercase letter
func ContainsUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func ContainsLowercase(s string) bool {
    for _, char := range s {
        if unicode.IsLower(char) {
            return true
        }
    }
    return false
}

func ContainsNumbers(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func ContainsSpecialChars(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	for _, char := range s {
		if strings.ContainsRune(specialChars, char) {
			return true
		}
	}
	return false
}
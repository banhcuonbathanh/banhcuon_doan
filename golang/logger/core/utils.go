// logger/core/utils.go - Utility functions for logging
package core

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GetTimestamp returns current timestamp in milliseconds
func GetTimestamp() int64 {
	return time.Now().UnixMilli()
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetRequestIDFromHeader extracts request ID from HTTP header or generates one
func GetRequestIDFromHeader(headers map[string][]string) string {
	if reqID := headers["X-Request-Id"]; len(reqID) > 0 && reqID[0] != "" {
		return reqID[0]
	}
	if reqID := headers["x-request-id"]; len(reqID) > 0 && reqID[0] != "" {
		return reqID[0]
	}
	// Generate a simple request ID if none exists
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// FormatDuration formats duration in milliseconds to human readable format
func FormatDuration(durationMS int64) string {
	if durationMS < 1000 {
		return fmt.Sprintf("%dms", durationMS)
	} else if durationMS < 60000 {
		return fmt.Sprintf("%.2fs", float64(durationMS)/1000)
	} else {
		return fmt.Sprintf("%.2fm", float64(durationMS)/60000)
	}
}

// SanitizeForLogging removes sensitive information from values for logging
func SanitizeForLogging(key string, value interface{}) interface{} {
	sensitiveKeys := map[string]bool{
		"password":      true,
		"secret":        true,
		"token":         true,
		"authorization": true,
		"api_key":       true,
		"private_key":   true,
	}
	
	if sensitiveKeys[strings.ToLower(key)] {
		return "[REDACTED]"
	}
	
	return value
}

// BuildLogContext creates a standardized log context
func BuildLogContext(component, operation string, fields map[string]interface{}) map[string]interface{} {
	logContext := make(map[string]interface{})
	
	// Add standard fields
	logContext["component"] = component
	logContext["operation"] = operation
	logContext["timestamp"] = time.Now().Unix()
	
	// Add custom fields
	for k, v := range fields {
		logContext[k] = SanitizeForLogging(k, v)
	}
	
	return logContext
}
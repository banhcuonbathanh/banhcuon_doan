// internal/logger/specialized.go - Specialized logging methods for different domains
package logger

import (
	"fmt"
	"strings"
	"time"
	"english-ai-full/logger/core"
)

// SpecializedLogger wraps CoreLogger to add domain-specific methods
type SpecializedLogger struct {
	*core.CoreLogger
}

// NewSpecializedLogger creates a new specialized logger
func NewSpecializedLogger(coreLogger *core.CoreLogger) *SpecializedLogger {
	return &SpecializedLogger{
		CoreLogger: coreLogger,
	}
}

// maskEmail masks the email for security logging
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "invalid_email"
	}
	
	username := parts[0]
	domain := parts[1]
	
	if len(username) <= 2 {
		return fmt.Sprintf("%s***@%s", username[:1], domain)
	}
	
	return fmt.Sprintf("%s***%s@%s", username[:2], username[len(username)-1:], domain)
}

// Enhanced authentication logging
func (l *SpecializedLogger) LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":      "authentication",
		"layer":          core.LayerAuth,
		"email":          maskEmail(email),
		"success":        success,
		"reason":         reason,
		"type":           "auth_attempt",
		"security_event": !success,
	}
	
	if !success {
		fields["cause"] = reason
	}
	
	if len(additionalContext) > 0 {
		for k, v := range additionalContext[0] {
			fields[k] = v
		}
	}
	
	message := fmt.Sprintf("Authentication %s for %s", 
		map[bool]string{true: "successful", false: "failed"}[success], 
		maskEmail(email))
	
	if success {
		l.Info(message, fields)
	} else {
		l.Warn(message, fields)
	}
}

// LogPasswordReset logs password reset attempts
func (l *SpecializedLogger) LogPasswordReset(email string, success bool, reason string) {
	fields := map[string]interface{}{
		"operation":      "password_reset",
		"layer":          core.LayerAuth,
		"email":          maskEmail(email),
		"success":        success,
		"type":           "password_reset",
		"security_event": true,
	}
	
	message := fmt.Sprintf("Password reset %s for %s: %s", 
		map[bool]string{true: "successful", false: "failed"}[success], 
		maskEmail(email), reason)
	
	if success {
		l.Info(message, fields)
	} else {
		l.Warn(message, fields)
	}
}

// LogSessionAction logs session-related actions
func (l *SpecializedLogger) LogSessionAction(action string, sessionID string, userID string, success bool) {
	fields := map[string]interface{}{
		"operation":  "session_management",
		"layer":      core.LayerAuth,
		"action":     action,
		"session_id": sessionID,
		"user_id":    userID,
		"success":    success,
		"type":       "session_action",
	}
	
	message := fmt.Sprintf("Session %s %s", action, 
		map[bool]string{true: "successful", false: "failed"}[success])
	
	if success {
		l.Info(message, fields)
	} else {
		l.Warn(message, fields)
	}
}

// LogDatabaseOperation logs database operations with performance metrics
func (l *SpecializedLogger) LogDatabaseOperation(operation string, table string, duration time.Duration, success bool, rowsAffected int64) {
	fields := map[string]interface{}{
		"operation":     operation,
		"layer":         core.LayerDatabase,
		"table":         table,
		"duration_ms":   duration.Milliseconds(),
		"success":       success,
		"rows_affected": rowsAffected,
		"type":          "db_operation",
	}
	
	message := fmt.Sprintf("Database %s on %s completed in %v", operation, table, duration)
	
	if success {
		l.Info(message, fields)
	} else {
		l.Error(message, fields)
	}
}

// LogAPICall logs external API calls
func (l *SpecializedLogger) LogAPICall(endpoint string, method string, statusCode int, duration time.Duration) {
	success := statusCode >= 200 && statusCode < 300
	
	fields := map[string]interface{}{
		"operation":    "api_call",
		"layer":        core.LayerExternal,
		"endpoint":     endpoint,
		"method":       method,
		"status_code":  statusCode,
		"duration_ms":  duration.Milliseconds(),
		"success":      success,
		"type":         "api_call",
	}
	
	message := fmt.Sprintf("API %s %s returned %d in %v", method, endpoint, statusCode, duration)
	
	if success {
		l.Info(message, fields)
	} else {
		l.Warn(message, fields)
	}
}

// LogCacheOperation logs cache operations
func (l *SpecializedLogger) LogCacheOperation(operation string, key string, hit bool, duration time.Duration) {
	fields := map[string]interface{}{
		"operation":   operation,
		"layer":       core.LayerCache,
		"cache_key":   key,
		"cache_hit":   hit,
		"duration_ms": duration.Milliseconds(),
		"type":        "cache_operation",
	}
	
	message := fmt.Sprintf("Cache %s for key %s", operation, key)
	if operation == "get" {
		message += fmt.Sprintf(" - %s", map[bool]string{true: "HIT", false: "MISS"}[hit])
	}
	
	l.Debug(message, fields)
}

// LogValidationError logs validation errors
func (l *SpecializedLogger) LogValidationError(field string, value interface{}, rule string, message string) {
	fields := map[string]interface{}{
		"operation": "validation",
		"layer":     core.LayerValidation,
		"field":     field,
		"value":     value,
		"rule":      rule,
		"type":      "validation_error",
	}
	
	l.Warn(fmt.Sprintf("Validation failed for field %s: %s", field, message), fields)
}

// LogSecurityEvent logs security-related events
func (l *SpecializedLogger) LogSecurityEvent(eventType string, severity string, description string, additionalFields map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":      "security_event",
		"layer":          core.LayerSecurity,
		"event_type":     eventType,
		"severity":       severity,
		"description":    description,
		"security_event": true,
		"type":           "security",
	}
	
	// Merge additional fields
	if additionalFields != nil {
		for k, v := range additionalFields {
			fields[k] = v
		}
	}
	
	message := fmt.Sprintf("Security event: %s - %s", eventType, description)
	
	switch severity {
	case "low":
		l.Info(message, fields)
	case "medium":
		l.Warn(message, fields)
	case "high", "critical":
		l.Error(message, fields)
	default:
		l.Warn(message, fields)
	}
}
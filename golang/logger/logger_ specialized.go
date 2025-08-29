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

// LogUserActivity logs user activities
func (l *SpecializedLogger) LogUserActivity(userID string, action string, resource string, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"operation": "user_activity",
		"layer":     core.LayerHandler,
		"user_id":   userID,
		"action":    action,
		"resource":  resource,
		"type":      "user_activity",
	}
	
	// Merge metadata
	if metadata != nil {
		for k, v := range metadata {
			fields[k] = v
		}
	}
	
	message := fmt.Sprintf("User %s performed %s on %s", userID, action, resource)
	l.Info(message, fields)
}

// LogMetric logs performance metrics
func (l *SpecializedLogger) LogMetric(name string, value float64, unit string, tags map[string]string) {
	fields := map[string]interface{}{
		"operation":    "metric_collection",
		"layer":        "metrics",
		"metric_name":  name,
		"metric_value": value,
		"metric_unit":  unit,
		"type":         "metric",
	}
	
	// Add tags as fields
	if tags != nil {
		for k, v := range tags {
			fields["tag_"+k] = v
		}
	}
	
	message := fmt.Sprintf("Metric: %s = %f %s", name, value, unit)
	l.Debug(message, fields)
}

// LogPerformance logs performance data
func (l *SpecializedLogger) LogPerformance(operation string, duration time.Duration, success bool, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":   "performance_tracking",
		"layer":       "performance",
		"perf_operation": operation,
		"duration_ms": duration.Milliseconds(),
		"success":     success,
		"type":        "performance",
	}
	
	// Merge metadata
	if metadata != nil {
		for k, v := range metadata {
			fields[k] = v
		}
	}
	
	message := fmt.Sprintf("Performance: %s completed in %v", operation, duration)
	if success {
		l.Info(message, fields)
	} else {
		l.Warn(message, fields)
	}
}

// LogRequestStart logs the start of a request
func (l *SpecializedLogger) LogRequestStart(requestID string, method string, endpoint string, userID string) {
	fields := map[string]interface{}{
		"operation":  "request_tracking",
		"layer":      core.LayerHandler,
		"request_id": requestID,
		"method":     method,
		"endpoint":   endpoint,
		"user_id":    userID,
		"phase":      "start",
		"type":       "request",
	}
	
	message := fmt.Sprintf("Request started: %s %s", method, endpoint)
	l.Debug(message, fields)
}

// LogRequestEnd logs the end of a request
func (l *SpecializedLogger) LogRequestEnd(requestID string, statusCode int, duration time.Duration, responseSize int64) {
	success := statusCode >= 200 && statusCode < 300
	
	fields := map[string]interface{}{
		"operation":     "request_tracking",
		"layer":         core.LayerHandler,
		"request_id":    requestID,
		"status_code":   statusCode,
		"duration_ms":   duration.Milliseconds(),
		"response_size": responseSize,
		"success":       success,
		"phase":         "end",
		"type":          "request",
	}
	
	message := fmt.Sprintf("Request completed: %d in %v", statusCode, duration)
	
	if success {
		l.Info(message, fields)
	} else if statusCode >= 400 {
		l.Error(message, fields)
	} else {
		l.Warn(message, fields)
	}
}

// LogBusinessEvent logs business logic events
func (l *SpecializedLogger) LogBusinessEvent(eventType string, entityID string, entityType string, action string, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":   "business_event",
		"layer":       core.LayerService,
		"event_type":  eventType,
		"entity_id":   entityID,
		"entity_type": entityType,
		"action":      action,
		"type":        "business",
	}
	
	// Merge metadata
	if metadata != nil {
		for k, v := range metadata {
			fields[k] = v
		}
	}
	
	message := fmt.Sprintf("Business event: %s %s on %s %s", action, eventType, entityType, entityID)
	l.Info(message, fields)
}

// LogSecurityEvent logs security-related events
func (l *SpecializedLogger) LogSecurityEvent(eventType string, severity string, userID string, ip string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":      "security_event",
		"layer":          core.LayerSecurity,
		"event_type":     eventType,
		"severity":       severity,
		"user_id":        userID,
		"ip_address":     ip,
		"security_event": true,
		"type":           "security",
	}
	
	// Merge details
	if details != nil {
		for k, v := range details {
			fields[k] = v
		}
	}
	
	message := fmt.Sprintf("Security event: %s (severity: %s)", eventType, severity)
	
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

// LogHealthCheck logs system health checks
func (l *SpecializedLogger) LogHealthCheck(service string, status string, duration time.Duration, details map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":   "health_check",
		"layer":       "health",
		"service":     service,
		"status":      status,
		"duration_ms": duration.Milliseconds(),
		"type":        "health",
	}
	
	// Merge details
	if details != nil {
		for k, v := range details {
			fields[k] = v
		}
	}
	
	message := fmt.Sprintf("Health check: %s - %s", service, status)
	
	if status == "healthy" || status == "ok" {
		l.Debug(message, fields)
	} else if status == "degraded" || status == "warning" {
		l.Warn(message, fields)
	} else {
		l.Error(message, fields)
	}
}
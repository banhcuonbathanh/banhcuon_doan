// logger/specialized.go - Specialized logging methods for different domains
package logger

import (
	"fmt"
	"strings"
	"time"
)

// Enhanced authentication logging with readable format
func (l *Logger) LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {
	context := map[string]interface{}{
		"operation":      "authentication",
		"layer":          LayerAuth,
		"email":          maskEmail(email),
		"success":        success,
		"reason":         reason,
		"type":           "auth_attempt",
		"security_event": !success,
	}
	
	if !success {
		context["cause"] = reason
	}
	
	if len(additionalContext) > 0 {
		for k, v := range additionalContext[0] {
			context[k] = v
		}
	}
	
	if success {
		l.Info("Authentication successful", context)
	} else {
		l.Warning("Authentication failed", context)
	}
}

// Enhanced API request logging - now much more readable
func (l *Logger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"type":        "api_request",
		"layer":       LayerHandler,
		"operation":   fmt.Sprintf("%s_%s", method, strings.ReplaceAll(path, "/", "_")),
	}
	
	// Add performance categories
	switch {
	case duration.Milliseconds() > 5000:
		logContext["performance"] = "very_slow"
		logContext["cause"] = "performance_issue"
	case duration.Milliseconds() > 2000:
		logContext["performance"] = "slow"
		logContext["cause"] = "performance_degradation"
	case duration.Milliseconds() > 1000:
		logContext["performance"] = "moderate"
	default:
		logContext["performance"] = "fast"
	}
	
	// Add error causes based on status codes
	if statusCode >= 500 {
		logContext["cause"] = "server_error"
	} else if statusCode >= 400 {
		logContext["cause"] = "client_error"
	}
	

	
	// Create readable message
	message := fmt.Sprintf("%s %s → %d", method, path, statusCode)
	
	switch {
	case statusCode >= 500:
		l.Error(message, logContext)
	case statusCode >= 400:
		l.Warning(message, logContext)
	case duration.Milliseconds() > 2000:
		l.Warning(message+" (slow)", logContext)
	default:
		l.Info(message, logContext)
	}
}

// Enhanced service call logging
func (l *Logger) LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"service":   service,
		"method":    method,
		"success":   success,
		"type":      "service_call",
		"layer":     LayerService,
		"operation": fmt.Sprintf("%s_%s", service, method),
	}
	

	
	message := fmt.Sprintf("%s.%s", service, method)
	
	if err != nil {
		logContext["error"] = err.Error()
		logContext["error_type"] = fmt.Sprintf("%T", err)
		logContext["cause"] = categorizeError(err)
		
		if isRetryableError(err) {
			logContext["retryable"] = true
		}
		
		l.Error(message+" failed", logContext)
	} else {
		l.Debug(message+" succeeded", logContext)
	}
}

// Enhanced DB operation logging
func (l *Logger) LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"operation": operation,
		"table":     table,
		"success":   success,
		"type":      "db_operation",
		"layer":     LayerRepository,
	}
	

	
	message := fmt.Sprintf("DB %s on %s", operation, table)
	
	if err != nil {
		logContext["error"] = err.Error()
		logContext["error_type"] = fmt.Sprintf("%T", err)
		logContext["cause"] = categorizeDBError(err)
		l.Error(message+" failed", logContext)
	} else if success {
		l.Debug(message+" succeeded", logContext)
	}
}

func (l *Logger) LogValidationError(field, message string, value interface{}) {
	context := map[string]interface{}{
		"field":     field,
		"message":   message,
		"type":      "validation_error",
		"layer":     LayerValidation,
		"operation": "validate_" + field,
		"cause":     "validation_failed",
	}
	
	fieldLower := strings.ToLower(field)
	if fieldLower == "password" || fieldLower == "token" || fieldLower == "secret" {
		context["value"] = "***hidden***"
		context["value_length"] = getValueLength(value)
	} else {
		context["value"] = value
		context["value_type"] = fmt.Sprintf("%T", value)
	}
	
	l.Warning(fmt.Sprintf("Validation failed for %s", field), context)
}

func (l *Logger) LogUserActivity(userID, email, action string, resource string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"user_id":   userID,
		"email":     maskEmail(email),
		"action":    action,
		"resource":  resource,
		"type":      "user_activity",
		"layer":     LayerHandler,
		"operation": fmt.Sprintf("%s_%s", action, resource),
	}
	

	l.Info(fmt.Sprintf("User %s performed %s on %s", maskEmail(email), action, resource), logContext)
}

func (l *Logger) LogSecurityEvent(eventType, description string, severity string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"event_type":     eventType,
		"description":    description,
		"severity":       severity,
		"type":           "security_event",
		"layer":          LayerSecurity,
		"operation":      "security_check",
		"security_event": true,
		"cause":          eventType,
	}
	
	
	
	message := fmt.Sprintf("Security: %s", description)
	
	switch strings.ToLower(severity) {
	case "critical", "high":
		l.Error(message, logContext)
	case "medium":
		l.Warning(message, logContext)
	default:
		l.Info(message, logContext)
	}
}

func (l *Logger) LogMetric(metricName string, value interface{}, unit string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"metric_name": metricName,
		"value":       value,
		"unit":        unit,
		"type":        "metric",
		"layer":       "monitoring",
		"operation":   "metric_collection",
	}
	

	l.Info(fmt.Sprintf("Metric: %s = %v %s", metricName, value, unit), logContext)
}

func (l *Logger) LogPerformance(operation string, duration time.Duration, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
		"duration_ns": duration.Nanoseconds(),
		"type":        "performance",
		"layer":       "monitoring",
	}
	
	// Add performance categories and causes
	switch {
	case duration.Milliseconds() > 10000:
		logContext["category"] = "critical_slow"
		logContext["cause"] = "critical_performance_issue"
	case duration.Milliseconds() > 5000:
		logContext["category"] = "very_slow"
		logContext["cause"] = "severe_performance_issue"
	case duration.Milliseconds() > 2000:
		logContext["category"] = "slow"
		logContext["cause"] = "performance_degradation"
	case duration.Milliseconds() > 1000:
		logContext["category"] = "moderate"
		logContext["cause"] = "minor_performance_issue"
	case duration.Milliseconds() > 500:
		logContext["category"] = "acceptable"
	default:
		logContext["category"] = "fast"
	}

	
	message := fmt.Sprintf("Performance: %s took %v", operation, duration)
	
	if duration.Milliseconds() > 5000 {
		l.Warning(message, logContext)
	} else {
		l.Debug(message, logContext)
	}
}
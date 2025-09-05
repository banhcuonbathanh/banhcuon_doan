// internal/logger/specialized.go - Enhanced specialized logging with auto-configuration
package logger

import (

	"fmt"
	"net/http"
	"strings"
	"time"
	"english-ai-full/logger/core"
)

// SpecializedLogger wraps CoreLogger with domain-specific methods and auto-configuration
type SpecializedLogger struct {
	*core.CoreLogger
}

// NewSpecializedLogger creates a new specialized logger
func NewSpecializedLogger(coreLogger *core.CoreLogger) *SpecializedLogger {
	return &SpecializedLogger{
		CoreLogger: coreLogger,
	}
}

// RequestLogger provides request-scoped logging with automatic context management
type RequestLogger struct {
	*SpecializedLogger
	context    *core.LogContext
	startTime  time.Time
}

// NewRequestLogger creates a request-scoped logger with automatic context extraction
func (l *SpecializedLogger) NewRequestLogger(r *http.Request) *RequestLogger {
	ctx := core.NewLogContext().
		WithRequestID(l.GetRequestID(r)).
		WithLayer(core.LayerHandler).
		WithField("method", r.Method).
		WithField("endpoint", r.URL.Path).
		WithField("client_ip", l.getClientIP(r))

	return &RequestLogger{
		SpecializedLogger: l,
		context:          ctx,
		startTime:        time.Now(),
	}
}

// ServiceLogger provides service-scoped logging
type ServiceLogger struct {
	*SpecializedLogger
	context *core.LogContext
}

// NewServiceLogger creates a service-scoped logger
func (l *SpecializedLogger) NewServiceLogger(serviceName, operation string) *ServiceLogger {
	ctx := core.NewLogContext().
		WithLayer(core.LayerService).
		WithOperation(operation).
		WithField("service", serviceName)

	return &ServiceLogger{
		SpecializedLogger: l,
		context:          ctx,
	}
}

// DatabaseLogger provides database-scoped logging
type DatabaseLogger struct {
	*SpecializedLogger
	context *core.LogContext
}

// NewDatabaseLogger creates a database-scoped logger
func (l *SpecializedLogger) NewDatabaseLogger(table, operation string) *DatabaseLogger {
	ctx := core.NewLogContext().
		WithLayer(core.LayerDatabase).
		WithOperation(operation).
		WithField("table", table)

	return &DatabaseLogger{
		SpecializedLogger: l,
		context:          ctx,
	}
}

// Request Logger Methods - Auto-configured for handler layer
func (rl *RequestLogger) LogRequestStart() {
	rl.CoreLogger.WithContext(rl.context).Info("Request started", map[string]interface{}{
		"phase": "start",
		"type":  "request",
	})
}

func (rl *RequestLogger) LogRequestEnd(statusCode int) {
	duration := time.Since(rl.startTime)
	success := statusCode >= 200 && statusCode < 300
	
	fields := map[string]interface{}{
		"phase":       "end",
		"type":        "request",
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"success":     success,
	}
	
	message := fmt.Sprintf("Request completed: %d in %v", statusCode, duration)
	
	contextLogger := rl.CoreLogger.WithContext(rl.context)
	if success {
		contextLogger.Info(message, fields)
	} else {
		fields["cause"] = rl.getErrorCauseFromStatus(statusCode)
		contextLogger.Error(message, fields)
	}
}

func (rl *RequestLogger) LogValidationError(err error) {
	fields := map[string]interface{}{
		"type":  "validation_error",
		"cause": "request_validation_failed",
		"error": err.Error(),
	}
	
	rl.CoreLogger.WithContext(rl.context).Error("Request validation failed", fields)
}

func (rl *RequestLogger) LogServiceCallStart(serviceName, method string) {
	fields := map[string]interface{}{
		"type":    "service_call",
		"phase":   "start",
		"service": serviceName,
		"method":  method,
	}
	
	rl.CoreLogger.WithContext(rl.context).Debug(fmt.Sprintf("Calling %s.%s", serviceName, method), fields)
}

func (rl *RequestLogger) LogServiceCallEnd(serviceName, method string, err error, duration time.Duration) {
	success := err == nil
	fields := map[string]interface{}{
		"type":        "service_call",
		"phase":       "end",
		"service":     serviceName,
		"method":      method,
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	}
	
	message := fmt.Sprintf("%s.%s completed in %v", serviceName, method, duration)
	
	contextLogger := rl.CoreLogger.WithContext(rl.context)
	if success {
		contextLogger.Info(message, fields)
	} else {
		fields["error"] = err.Error()
		fields["cause"] = "service_call_failed"
		contextLogger.Error(message, fields)
	}
}

func (rl *RequestLogger) LogBusinessEvent(eventType, entityID, entityType, action string, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"type":        "business_event",
		"event_type":  eventType,
		"entity_id":   entityID,
		"entity_type": entityType,
		"action":      action,
	}
	
	// Merge metadata
	if metadata != nil {
		for k, v := range metadata {
			fields[k] = v
		}
	}
	
	message := fmt.Sprintf("Business event: %s %s on %s %s", action, eventType, entityType, entityID)
	rl.CoreLogger.WithContext(rl.context).Info(message, fields)
}

func (rl *RequestLogger) Error(message string, fields ...map[string]interface{}) {
	rl.CoreLogger.WithContext(rl.context).Error(message, fields...)
}

func (rl *RequestLogger) Info(message string, fields ...map[string]interface{}) {
	rl.CoreLogger.WithContext(rl.context).Info(message, fields...)
}

func (rl *RequestLogger) Warn(message string, fields ...map[string]interface{}) {
	rl.CoreLogger.WithContext(rl.context).Warn(message, fields...)
}

func (rl *RequestLogger) Debug(message string, fields ...map[string]interface{}) {
	rl.CoreLogger.WithContext(rl.context).Debug(message, fields...)
}

// Service Logger Methods - Auto-configured for service layer
func (sl *ServiceLogger) LogOperationStart() {
	sl.CoreLogger.WithContext(sl.context).Debug("Service operation started", map[string]interface{}{
		"phase": "start",
		"type":  "service_operation",
	})
}

func (sl *ServiceLogger) LogOperationEnd(success bool, err error) {
	duration := sl.context.Duration()
	fields := map[string]interface{}{
		"phase":       "end",
		"type":        "service_operation", 
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	}
	
	message := fmt.Sprintf("Service operation completed in %v", duration)
	
	contextLogger := sl.CoreLogger.WithContext(sl.context)
	if success {
		contextLogger.Info(message, fields)
	} else {
		if err != nil {
			fields["error"] = err.Error()
			fields["cause"] = "service_operation_failed"
		}
		contextLogger.Error(message, fields)
	}
}

func (sl *ServiceLogger) LogDatabaseCall(operation, table string, success bool, rowsAffected int64, duration time.Duration) {
	fields := map[string]interface{}{
		"type":          "database_call",
		"db_operation":  operation,
		"table":         table,
		"success":       success,
		"rows_affected": rowsAffected,
		"duration_ms":   duration.Milliseconds(),
	}
	
	message := fmt.Sprintf("Database %s on %s completed in %v", operation, table, duration)
	
	contextLogger := sl.CoreLogger.WithContext(sl.context)
	if success {
		contextLogger.Info(message, fields)
	} else {
		fields["cause"] = fmt.Sprintf("database_%s_failed", strings.ToLower(operation))
		contextLogger.Error(message, fields)
	}
}

func (sl *ServiceLogger) LogExternalAPICall(endpoint, method string, statusCode int, duration time.Duration) {
	success := statusCode >= 200 && statusCode < 300
	fields := map[string]interface{}{
		"type":        "external_api_call",
		"endpoint":    endpoint,
		"method":      method,
		"status_code": statusCode,
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	}
	
	message := fmt.Sprintf("API %s %s returned %d in %v", method, endpoint, statusCode, duration)
	
	contextLogger := sl.CoreLogger.WithContext(sl.context)
	if success {
		contextLogger.Info(message, fields)
	} else {
		if statusCode >= 500 {
			fields["cause"] = "external_service_error"
		} else if statusCode >= 400 {
			fields["cause"] = "client_request_error"
		} else {
			fields["cause"] = "api_call_failed"
		}
		contextLogger.Error(message, fields)
	}
}

func (sl *ServiceLogger) Error(message string, fields ...map[string]interface{}) {
	sl.CoreLogger.WithContext(sl.context).Error(message, fields...)
}

func (sl *ServiceLogger) Info(message string, fields ...map[string]interface{}) {
	sl.CoreLogger.WithContext(sl.context).Info(message, fields...)
}

func (sl *ServiceLogger) Warn(message string, fields ...map[string]interface{}) {
	sl.CoreLogger.WithContext(sl.context).Warn(message, fields...)
}

func (sl *ServiceLogger) Debug(message string, fields ...map[string]interface{}) {
	sl.CoreLogger.WithContext(sl.context).Debug(message, fields...)
}

// Database Logger Methods - Auto-configured for database layer
func (dl *DatabaseLogger) LogQueryStart(query string) {
	fields := map[string]interface{}{
		"type":  "database_query",
		"phase": "start",
		"query": dl.maskQuery(query), // Mask sensitive data
	}
	
	dl.CoreLogger.WithContext(dl.context).Debug("Database query started", fields)
}

func (dl *DatabaseLogger) LogQueryEnd(success bool, rowsAffected int64, err error) {
	duration := dl.context.Duration()
	fields := map[string]interface{}{
		"type":          "database_query",
		"phase":         "end",
		"success":       success,
		"rows_affected": rowsAffected,
		"duration_ms":   duration.Milliseconds(),
	}
	
	message := fmt.Sprintf("Database query completed in %v, affected %d rows", duration, rowsAffected)
	
	contextLogger := dl.CoreLogger.WithContext(dl.context)
	if success {
		contextLogger.Info(message, fields)
	} else {
		if err != nil {
			fields["error"] = err.Error()
		}
		fields["cause"] = "database_query_failed"
		contextLogger.Error(message, fields)
	}
}

func (dl *DatabaseLogger) LogTransaction(action string, success bool) {
	fields := map[string]interface{}{
		"type":    "database_transaction",
		"action":  action, // begin, commit, rollback
		"success": success,
	}
	
	message := fmt.Sprintf("Database transaction %s", action)
	
	contextLogger := dl.CoreLogger.WithContext(dl.context)
	if success {
		contextLogger.Info(message, fields)
	} else {
		fields["cause"] = fmt.Sprintf("transaction_%s_failed", action)
		contextLogger.Error(message, fields)
	}
}

func (dl *DatabaseLogger) Error(message string, fields ...map[string]interface{}) {
	dl.CoreLogger.WithContext(dl.context).Error(message, fields...)
}

func (dl *DatabaseLogger) Info(message string, fields ...map[string]interface{}) {
	dl.CoreLogger.WithContext(dl.context).Info(message, fields...)
}

func (dl *DatabaseLogger) Warn(message string, fields ...map[string]interface{}) {
	dl.CoreLogger.WithContext(dl.context).Warn(message, fields...)
}

func (dl *DatabaseLogger) Debug(message string, fields ...map[string]interface{}) {
	dl.CoreLogger.WithContext(dl.context).Debug(message, fields...)
}

// Traditional specialized methods with auto-configuration
func (l *SpecializedLogger) LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {
	ctx := core.NewLogContext().
		WithLayer(core.LayerAuth).
		WithOperation("authentication").
		WithField("email", l.maskEmail(email)).
		WithField("success", success).
		WithField("reason", reason).
		WithField("type", "auth_attempt").
		WithField("security_event", !success)
	
	if len(additionalContext) > 0 {
		ctx.WithFields(additionalContext[0])
	}
	
	if !success {
		ctx.WithField("cause", reason)
	}
	
	message := fmt.Sprintf("Authentication %s for %s", 
		map[bool]string{true: "successful", false: "failed"}[success], 
		l.maskEmail(email))
	
	contextLogger := l.CoreLogger.WithContext(ctx)
	if success {
		contextLogger.Info(message)
	} else {
		contextLogger.Error(message)
	}
}

func (l *SpecializedLogger) LogPasswordReset(email string, success bool, reason string) {
	ctx := core.NewLogContext().
		WithLayer(core.LayerAuth).
		WithOperation("password_reset").
		WithField("email", l.maskEmail(email)).
		WithField("success", success).
		WithField("type", "password_reset").
		WithField("security_event", true)
	
	message := fmt.Sprintf("Password reset %s for %s: %s", 
		map[bool]string{true: "successful", false: "failed"}[success], 
		l.maskEmail(email), reason)
	
	contextLogger := l.CoreLogger.WithContext(ctx)
	if success {
		contextLogger.Info(message)
	} else {
		ctx.WithField("cause", reason)
		contextLogger.Warn(message)
	}
}

func (l *SpecializedLogger) LogSessionAction(action string, sessionID string, userID string, success bool) {
	ctx := core.NewLogContext().
		WithLayer(core.LayerAuth).
		WithOperation("session_management").
		WithField("action", action).
		WithField("session_id", sessionID).
		WithField("user_id", userID).
		WithField("success", success).
		WithField("type", "session_action")
	
	if !success {
		ctx.WithField("cause", fmt.Sprintf("session_%s_failed", action))
	}
	
	message := fmt.Sprintf("Session %s %s", action, 
		map[bool]string{true: "successful", false: "failed"}[success])
	
	contextLogger := l.CoreLogger.WithContext(ctx)
	if success {
		contextLogger.Info(message)
	} else {
		contextLogger.Warn(message)
	}
}

func (l *SpecializedLogger) LogCacheOperation(operation string, key string, hit bool, duration time.Duration) {
	ctx := core.NewLogContext().
		WithLayer(core.LayerCache).
		WithOperation(operation).
		WithField("cache_key", key).
		WithField("cache_hit", hit).
		WithField("duration_ms", duration.Milliseconds()).
		WithField("type", "cache_operation")
	
	message := fmt.Sprintf("Cache %s for key %s", operation, key)
	if operation == "get" {
		message += fmt.Sprintf(" - %s", map[bool]string{true: "HIT", false: "MISS"}[hit])
	}
	
	l.CoreLogger.WithContext(ctx).Debug(message)
}

func (l *SpecializedLogger) LogSecurityEvent(eventType string, severity string, userID string, ip string, details map[string]interface{}) {
	ctx := core.NewLogContext().
		WithLayer(core.LayerSecurity).
		WithOperation("security_event").
		WithField("event_type", eventType).
		WithField("severity", severity).
		WithField("user_id", userID).
		WithField("ip_address", ip).
		WithField("security_event", true).
		WithField("type", "security").
		WithField("cause", fmt.Sprintf("security_event_%s", strings.ToLower(eventType)))
	
	if details != nil {
		ctx.WithFields(details)
	}
	
	message := fmt.Sprintf("Security event: %s (severity: %s)", eventType, severity)
	
	contextLogger := l.CoreLogger.WithContext(ctx)
	switch severity {
	case "low":
		contextLogger.Info(message)
	case "medium":
		contextLogger.Warn(message)
	case "high", "critical":
		contextLogger.Error(message)
	default:
		contextLogger.Warn(message)
	}
}

func (l *SpecializedLogger) LogHealthCheck(service string, status string, duration time.Duration, details map[string]interface{}) {
	ctx := core.NewLogContext().
		WithLayer("health").
		WithOperation("health_check").
		WithField("service", service).
		WithField("status", status).
		WithField("duration_ms", duration.Milliseconds()).
		WithField("type", "health")
	
	if details != nil {
		ctx.WithFields(details)
	}
	
	if status != "healthy" && status != "ok" {
		ctx.WithField("cause", fmt.Sprintf("health_check_%s", status))
	}
	
	message := fmt.Sprintf("Health check: %s - %s", service, status)
	
	contextLogger := l.CoreLogger.WithContext(ctx)
	if status == "healthy" || status == "ok" {
		contextLogger.Debug(message)
	} else if status == "degraded" || status == "warning" {
		contextLogger.Warn(message)
	} else {
		contextLogger.Error(message)
	}
}

func (l *SpecializedLogger) LogUserActivity(userID string, action string, resource string, metadata map[string]interface{}) {
	ctx := core.NewLogContext().
		WithLayer(core.LayerHandler).
		WithOperation("user_activity").
		WithField("user_id", userID).
		WithField("action", action).
		WithField("resource", resource).
		WithField("type", "user_activity")
	
	if metadata != nil {
		ctx.WithFields(metadata)
	}
	
	message := fmt.Sprintf("User %s performed %s on %s", userID, action, resource)
	l.CoreLogger.WithContext(ctx).Info(message)
}

func (l *SpecializedLogger) LogStructValidationError(structName string, structValue interface{}, message string) {
	ctx := core.NewLogContext().
		WithLayer(core.LayerValidation).
		WithOperation("validation").
		WithField("struct_name", structName).
		WithField("struct_value", structValue).
		WithField("type", "struct_validation_error").
		WithField("cause", fmt.Sprintf("validation_failed_%s", structName))
	
	l.CoreLogger.WithContext(ctx).Warn(fmt.Sprintf("Struct validation failed for %s: %s", structName, message))
}

func (l *SpecializedLogger) LogMetric(name string, value float64, unit string, tags map[string]string) {
	ctx := core.NewLogContext().
		WithLayer("metrics").
		WithOperation("metric_collection").
		WithField("metric_name", name).
		WithField("metric_value", value).
		WithField("metric_unit", unit).
		WithField("type", "metric")
	
	if tags != nil {
		for k, v := range tags {
			ctx.WithField(fmt.Sprintf("tag_%s", k), v)
		}
	}
	
	message := fmt.Sprintf("Metric: %s = %f %s", name, value, unit)
	l.CoreLogger.WithContext(ctx).Debug(message)
}

func (l *SpecializedLogger) LogPerformance(operation string, duration time.Duration, success bool, metadata map[string]interface{}) {
	ctx := core.NewLogContext().
		WithLayer("performance").
		WithOperation("performance_tracking").
		WithField("perf_operation", operation).
		WithField("duration_ms", duration.Milliseconds()).
		WithField("success", success).
		WithField("type", "performance")
	
	if metadata != nil {
		ctx.WithFields(metadata)
	}
	
	if !success {
		ctx.WithField("cause", fmt.Sprintf("performance_%s_degraded", operation))
	}
	
	message := fmt.Sprintf("Performance: %s completed in %v", operation, duration)
	
	contextLogger := l.CoreLogger.WithContext(ctx)
	if success {
		contextLogger.Info(message)
	} else {
		contextLogger.Warn(message)
	}
}

// Helper methods
func (l *SpecializedLogger) maskEmail(email string) string {
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

func (l *SpecializedLogger) maskQuery(query string) string {
	// Simple query masking - enhance as needed
	sensitivePatterns := []string{"password", "token", "secret", "key"}
	maskedQuery := query
	
	for _, pattern := range sensitivePatterns {
		maskedQuery = strings.ReplaceAll(
			strings.ToLower(maskedQuery), 
			pattern, 
			"***",
		)
	}
	
	return maskedQuery
}

func (l *SpecializedLogger) GetRequestID(r *http.Request) string {
	// Try multiple common request ID headers
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = r.Header.Get("X-Request-Id")
	}
	if requestID == "" {
		requestID = r.Header.Get("Request-ID")
	}
	if requestID == "" {
		// Generate a simple request ID if none found
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return requestID
}
func (l *SpecializedLogger) getClientIP(r *http.Request) string {
	// Try various headers for client IP
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if strings.Contains(ip, ",") {
			ip = strings.TrimSpace(strings.Split(ip, ",")[0])
		}
		return ip
	}
	
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	
	return r.RemoteAddr
}

func (l *SpecializedLogger) getErrorCauseFromStatus(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "server_error"
	case statusCode >= 400:
		return "client_error"
	default:
		return "request_failed"
	}
}
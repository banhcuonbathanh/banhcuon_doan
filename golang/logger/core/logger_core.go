// internal/logger/core/logger_core.go - Enhanced core types and structures
package core

import (
	"sync"
	"time"
)

// Level represents log levels with proper ordering
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var levelNames = map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO", 
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
}

func (l Level) String() string {
	if name, exists := levelNames[l]; exists {
		return name
	}
	return "UNKNOWN"
}

// Output formats
const (
	FormatJSON   = "json"
	FormatText   = "text"
	FormatPretty = "pretty"
)

// Layer constants for better organization
const (
	LayerHandler    = "handler"
	LayerService    = "service" 
	LayerRepository = "repository"
	LayerMiddleware = "middleware"
	LayerAuth       = "auth"
	LayerValidation = "validation"
	LayerCache      = "cache"
	LayerDatabase   = "database"
	LayerExternal   = "external"
	LayerSecurity   = "security"
)

// LogEntry represents a structured log entry with enhanced metadata
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       Level                  `json:"level"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Caller      string                 `json:"caller,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Component   string                 `json:"component,omitempty"`
	Operation   string                 `json:"operation,omitempty"`
	Duration    time.Duration          `json:"duration_ns,omitempty"`
	ErrorCode   string                 `json:"error_code,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Cause       string                 `json:"cause,omitempty"`
	Layer       string                 `json:"layer,omitempty"`
}

// Logger represents the main logger with enhanced capabilities
type CoreLogger struct {
	level         Level
	outputManager OutputManager
	asyncEnabled  bool
	buffer        LogBuffer
	contextFields map[string]interface{}
	component     string
	layer         string
	operation     string
	environment   string
	mu            sync.RWMutex
}

// OutputManager interface for managing multiple outputs
type OutputManager interface {
	WriteToAll(entry *LogEntry) error
	WriteToOutput(name string, entry *LogEntry) error
	AddOutput(name string, output Output) error
	RemoveOutput(name string) error
	Close() error
}

// Output interface for different output destinations
type Output interface {
	Write(entry *LogEntry) error
	Close() error
}

// LogBuffer interface for async processing
type LogBuffer interface {
	Add(entry *LogEntry) error
	Flush() error
	Close() error
}

// NewLogger creates a new enhanced logger instance
func NewLogger() *CoreLogger {
	return &CoreLogger{
		level:         InfoLevel,
		contextFields: make(map[string]interface{}),
		environment:   "development",
	}
}

// Configuration methods
func (l *CoreLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *CoreLogger) SetComponent(component string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.component = component
}

func (l *CoreLogger) SetLayer(layer string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.layer = layer
}

func (l *CoreLogger) SetOperation(operation string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.operation = operation
}

func (l *CoreLogger) SetEnvironment(env string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.environment = env
}

func (l *CoreLogger) AddContextField(key string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.contextFields[key] = value
}

func (l *CoreLogger) RemoveContextField(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.contextFields, key)
}

// Core logging methods
func (l *CoreLogger) Debug(message string, fields ...map[string]interface{}) {
	l.log(DebugLevel, message, fields...)
}

func (l *CoreLogger) Info(message string, fields ...map[string]interface{}) {
	l.log(InfoLevel, message, fields...)
}

func (l *CoreLogger) Warn(message string, fields ...map[string]interface{}) {
	l.log(WarnLevel, message, fields...)
}

func (l *CoreLogger) Error(message string, fields ...map[string]interface{}) {
	l.log(ErrorLevel, message, fields...)
}

func (l *CoreLogger) Fatal(message string, fields ...map[string]interface{}) {
	l.log(FatalLevel, message, fields...)
	// Note: In production, this might call os.Exit(1)
}

// Enhanced logging methods
func (l *CoreLogger) ErrorWithCause(message, cause, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["cause"] = cause
	mergedFields["layer"] = layer  
	mergedFields["operation"] = operation
	l.log(ErrorLevel, message, mergedFields)
}

func (l *CoreLogger) WarnWithCause(message, cause, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["cause"] = cause
	mergedFields["layer"] = layer
	mergedFields["operation"] = operation
	l.log(WarnLevel, message, mergedFields)
}

func (l *CoreLogger) InfoWithOperation(message, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["layer"] = layer
	mergedFields["operation"] = operation
	l.log(InfoLevel, message, mergedFields)
}

// Specialized logging methods - Add these missing methods
func (l *CoreLogger) LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":      "authentication",
		"layer":          LayerAuth,
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
	
	message := "Authentication " + map[bool]string{true: "successful", false: "failed"}[success] + " for " + maskEmail(email)
	
	if success {
		l.Info(message, fields)
	} else {
		l.Warn(message, fields)
	}
}

func (l *CoreLogger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {
	success := statusCode >= 200 && statusCode < 300
	
	fields := map[string]interface{}{
		"operation":    "api_request",
		"layer":        LayerHandler,
		"method":       method,
		"path":         path,
		"status_code":  statusCode,
		"duration_ms":  duration.Milliseconds(),
		"success":      success,
		"type":         "api_request",
	}
	
	// Merge context fields
	if context != nil {
		for k, v := range context {
			fields[k] = v
		}
	}
	
	message := "API " + method + " " + path + " returned " + string(rune(statusCode))
	
	if success {
		l.Info(message, fields)
	} else {
		l.Warn(message, fields)
	}
}

func (l *CoreLogger) LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {
	fields := map[string]interface{}{
		"operation": "service_call",
		"layer":     LayerService,
		"service":   service,
		"method":    method,
		"success":   success,
		"type":      "service_call",
	}
	
	if err != nil {
		fields["error"] = err.Error()
	}
	
	// Merge context fields
	if context != nil {
		for k, v := range context {
			fields[k] = v
		}
	}
	
	message := "Service call " + service + "." + method
	
	if success {
		l.Info(message, fields)
	} else {
		l.Error(message, fields)
	}
}

func (l *CoreLogger) LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {
	fields := map[string]interface{}{
		"operation": "db_operation",
		"layer":     LayerDatabase,
		"db_operation": operation,
		"table":     table,
		"success":   success,
		"type":      "db_operation",
	}
	
	if err != nil {
		fields["error"] = err.Error()
	}
	
	// Merge context fields
	if context != nil {
		for k, v := range context {
			fields[k] = v
		}
	}
	
	message := "Database " + operation + " on " + table
	
	if success {
		l.Info(message, fields)
	} else {
		l.Error(message, fields)
	}
}

func (l *CoreLogger) LogValidationError(field, message string, value interface{}) {
	fields := map[string]interface{}{
		"operation": "validation",
		"layer":     LayerValidation,
		"field":     field,
		"value":     value,
		"type":      "validation_error",
	}
	
	l.Warn("Validation failed for field "+field+": "+message, fields)
}

func (l *CoreLogger) LogUserActivity(userID, email, action, resource string, context map[string]interface{}) {
	fields := map[string]interface{}{
		"operation": "user_activity",
		"layer":     LayerHandler,
		"user_id":   userID,
		"email":     maskEmail(email),
		"action":    action,
		"resource":  resource,
		"type":      "user_activity",
	}
	

	
	message := "User " + userID + " performed " + action + " on " + resource
	l.Info(message, fields)
}

func (l *CoreLogger) LogSecurityEvent(eventType, description, severity string, context map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":      "security_event",
		"layer":          LayerSecurity,
		"event_type":     eventType,
		"description":    description,
		"severity":       severity,
		"security_event": true,
		"type":           "security",
	}
	
	// Merge context fields
	if context != nil {
		for k, v := range context {
			fields[k] = v
		}
	}
	
	message := "Security event: " + eventType + " - " + description
	
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

func (l *CoreLogger) LogMetric(metricName string, value interface{}, unit string, context map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":    "metric_collection",
		"layer":        "metrics",
		"metric_name":  metricName,
		"metric_value": value,
		"metric_unit":  unit,
		"type":         "metric",
	}
	
	// Merge context fields
	if context != nil {
		for k, v := range context {
			fields[k] = v
		}
	}
	
	message := "Metric: " + metricName
	l.Debug(message, fields)
}

func (l *CoreLogger) LogPerformance(operation string, duration time.Duration, context map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":       "performance_tracking",
		"layer":           "performance",
		"perf_operation":  operation,
		"duration_ms":     duration.Milliseconds(),
		"type":            "performance",
	}
	
	// Merge context fields
	if context != nil {
		for k, v := range context {
			fields[k] = v
		}
	}
	
	message := "Performance: " + operation + " completed"
	l.Info(message, fields)
}

// WriteToOutput writes directly to a specific output
func (l *CoreLogger) WriteToOutput(outputName string, entry *LogEntry) error {
	if l.outputManager != nil {
		return l.outputManager.WriteToOutput(outputName, entry)
	}
	return nil
}

// Core logging implementation
func (l *CoreLogger) log(level Level, message string, fields ...map[string]interface{}) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	
	// Create log entry
	entry := &LogEntry{
		Timestamp:   time.Now(),
		Level:       level,
		Message:     message,
		Fields:      l.mergeFields(fields...),
		Component:   l.component,
		Layer:       l.layer,
		Operation:   l.operation,
		Environment: l.environment,
	}
	
	// Add caller information
	if caller := getCaller(3); caller != "" {
		entry.Caller = caller
	}
	l.mu.RUnlock()
	
	// Write to outputs
	if l.outputManager != nil {
		l.outputManager.WriteToAll(entry)
	}
	
	// Add to async buffer if enabled
	if l.asyncEnabled && l.buffer != nil {
		l.buffer.Add(entry)
	}
}

// Helper methods
func (l *CoreLogger) mergeFields(fields ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	// Add context fields first
	for k, v := range l.contextFields {
		merged[k] = v
	}
	
	// Add provided fields (will override context fields if same key)
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			merged[k] = v
		}
	}
	
	return merged
}

// Helper function to mask email for security
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	// Simple masking implementation
	if len(email) > 3 {
		return email[:2] + "***" + email[len(email)-1:]
	}
	return "***"
}

func getCaller(skip int) string {
	// Implementation would use runtime.Caller to get file:line info
	// Simplified for this example
	return ""
}
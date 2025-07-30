// logger/logger.go - Enhanced version
package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Logger levels with numeric values for comparison
const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

var levelNames = map[int]string{
	DebugLevel:   "DEBUG",
	InfoLevel:    "INFO",
	WarningLevel: "WARNING",
	ErrorLevel:   "ERROR",
	FatalLevel:   "FATAL",
}

// LogEntry represents a structured log entry with enhanced metadata
type LogEntry struct {
	Timestamp    string                 `json:"timestamp"`
	Level        string                 `json:"level"`
	Message      string                 `json:"message"`
	Context      map[string]interface{} `json:"context,omitempty"`
	File         string                 `json:"file,omitempty"`
	Function     string                 `json:"function,omitempty"`
	Line         int                    `json:"line,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	TraceID      string                 `json:"trace_id,omitempty"`
	Component    string                 `json:"component,omitempty"`    // handler, service, repository
	Operation    string                 `json:"operation,omitempty"`    // login, register, etc.
	Duration     int64                  `json:"duration_ms,omitempty"`  // Operation duration in milliseconds
	ErrorCode    string                 `json:"error_code,omitempty"`
	Environment  string                 `json:"environment,omitempty"`
}

// Logger structure with enhanced capabilities and thread safety
type Logger struct {
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
	enableJSON    bool
	enableDebug   bool
	minLevel      int
	environment   string
	component     string
	mutex         sync.RWMutex
	contextFields map[string]interface{} // Global context fields
}

// NewLogger creates a new enhanced Logger instance with configuration
func NewLogger() *Logger {
	environment := getEnvironment()
	
	logger := &Logger{
		debugLogger:   log.New(os.Stdout, "", 0),
		infoLogger:    log.New(os.Stdout, "", 0),
		warningLogger: log.New(os.Stdout, "", 0),
		errorLogger:   log.New(os.Stderr, "", 0),
		fatalLogger:   log.New(os.Stderr, "", 0),
		enableJSON:    true,
		enableDebug:   environment == "development",
		minLevel:      getMinLogLevel(environment),
		environment:   environment,
		contextFields: make(map[string]interface{}),
	}
	
	// Set initial global context
	logger.contextFields["environment"] = environment
	logger.contextFields["service"] = "restaurant-api" // Configure as needed
	
	return logger
}

// Configuration methods with thread safety
func (l *Logger) SetJSONLogging(enable bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.enableJSON = enable
}

func (l *Logger) SetDebugLogging(enable bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.enableDebug = enable
}

func (l *Logger) SetMinLevel(level int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.minLevel = level
}

func (l *Logger) SetComponent(component string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.component = component
}

// Add global context field
func (l *Logger) AddGlobalField(key string, value interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.contextFields[key] = value
}

// Remove global context field
func (l *Logger) RemoveGlobalField(key string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.contextFields, key)
}

// Enhanced caller info retrieval with configurable skip levels
func (l *Logger) getCallerInfo(skip int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", "", 0
	}
	
	function := runtime.FuncForPC(pc).Name()
	
	// Shorten file path for readability
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}
	
	// Shorten function name
	if lastDot := strings.LastIndex(function, "."); lastDot >= 0 {
		function = function[lastDot+1:]
	}
	
	return file, function, line
}

// Enhanced context merging with conflict resolution
func (l *Logger) mergeContext(baseContext, additionalContext map[string]interface{}) map[string]interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	merged := make(map[string]interface{})
	
	// Add global context fields first
	for k, v := range l.contextFields {
		merged[k] = v
	}
	
	// Add base context (overwrites global if conflict)
	if baseContext != nil {
		for k, v := range baseContext {
			merged[k] = v
		}
	}
	
	// Add additional context (overwrites all if conflict)
	if additionalContext != nil {
		for k, v := range additionalContext {
			merged[k] = v
		}
	}
	
	return merged
}

// Core logging method with enhanced features
func (l *Logger) logWithContext(level int, message string, context map[string]interface{}, skip int) {
	l.mutex.RLock()
	enableJSON := l.enableJSON
	enableDebug := l.enableDebug
	minLevel := l.minLevel
	component := l.component
	environment := l.environment
	l.mutex.RUnlock()
	
	// Check if logging is enabled for this level
	if level < minLevel {
		return
	}
	
	// Skip debug logs if not enabled
	if level == DebugLevel && !enableDebug {
		return
	}
	
	levelStr := levelNames[level]
	var logger *log.Logger
	
	switch level {
	case DebugLevel:
		logger = l.debugLogger
	case InfoLevel:
		logger = l.infoLogger
	case WarningLevel:
		logger = l.warningLogger
	case ErrorLevel:
		logger = l.errorLogger
	case FatalLevel:
		logger = l.fatalLogger
	default:
		logger = l.infoLogger
	}
	
	if enableJSON {
		file, function, line := l.getCallerInfo(skip + 1) // +1 to account for this method
		
		// Merge all context
		mergedContext := l.mergeContext(context, nil)
		
		entry := LogEntry{
			Timestamp:   time.Now().Format("2006-01-02 15:04:05.000"),
			Level:       levelStr,
			Message:     message,
			Context:     mergedContext,
			File:        file,
			Function:    function,
			Line:        line,
			Component:   component,
			Environment: environment,
		}
		
		// Extract special fields from context if present
		if mergedContext != nil {
			if requestID, ok := mergedContext["request_id"].(string); ok {
				entry.RequestID = requestID
			}
			if userID, ok := mergedContext["user_id"].(string); ok {
				entry.UserID = userID
			}
			if sessionID, ok := mergedContext["session_id"].(string); ok {
				entry.SessionID = sessionID
			}
			if traceID, ok := mergedContext["trace_id"].(string); ok {
				entry.TraceID = traceID
			}
			if operation, ok := mergedContext["operation"].(string); ok {
				entry.Operation = operation
			}
			if duration, ok := mergedContext["duration_ms"].(int64); ok {
				entry.Duration = duration
			}
			if errorCode, ok := mergedContext["error_code"].(string); ok {
				entry.ErrorCode = errorCode
			}
		}
		
		jsonData, err := json.Marshal(entry)
		if err != nil {
			// Fallback to simple logging if JSON fails
			logger.Printf("[%s] %s | Context: %+v | JSON Error: %v", levelStr, message, mergedContext, err)
		} else {
			logger.Println(string(jsonData))
		}
	} else {
		// Simple text format with enhanced readability
		if context != nil && len(context) > 0 {
			logger.Printf("[%s] %s | %+v", levelStr, message, context)
		} else {
			logger.Printf("[%s] %s", levelStr, message)
		}
	}
	
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Public logging methods with consistent interface
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(DebugLevel, message, ctx, 3)
}

func (l *Logger) Info(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(InfoLevel, message, ctx, 3)
}

func (l *Logger) Warning(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(WarningLevel, message, ctx, 3)
}

func (l *Logger) Error(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(ErrorLevel, message, ctx, 3)
}

func (l *Logger) Fatal(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(FatalLevel, message, ctx, 3)
}

// Specialized logging methods with enhanced context

// Enhanced authentication logging with security considerations
func (l *Logger) LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {
	context := map[string]interface{}{
		"operation": "authentication",
		"email":     maskEmail(email), // Mask email for security
		"success":   success,
		"reason":    reason,
		"type":      "auth_attempt",
	}
	
	// Merge additional context if provided
	if len(additionalContext) > 0 {
		for k, v := range additionalContext[0] {
			context[k] = v
		}
	}
	
	if success {
		l.Info("Authentication attempt successful", context)
	} else {
		// Add security monitoring fields for failed attempts
		context["security_event"] = true
		l.Warning("Authentication attempt failed", context)
	}
}

// Enhanced database operation logging with performance metrics
func (l *Logger) LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"operation": operation,
		"table":     table,
		"success":   success,
		"type":      "db_operation",
		"component": "repository",
	}
	
	// Merge additional context
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
	}
	
	if err != nil {
		logContext["error"] = err.Error()
		logContext["error_type"] = fmt.Sprintf("%T", err)
		l.Error(fmt.Sprintf("Database %s operation failed on %s", operation, table), logContext)
	} else if success {
		l.Debug(fmt.Sprintf("Database %s operation successful on %s", operation, table), logContext)
	}
}

// Enhanced service call logging with timeout and retry context
func (l *Logger) LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"service":   service,
		"method":    method,
		"success":   success,
		"type":      "service_call",
		"component": "service",
	}
	
	// Merge additional context
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
	}
	
	if err != nil {
		logContext["error"] = err.Error()
		logContext["error_type"] = fmt.Sprintf("%T", err)
		
		// Check if it's a retryable error
		if isRetryableError(err) {
			logContext["retryable"] = true
		}
		
		l.Error(fmt.Sprintf("Service call failed: %s.%s", service, method), logContext)
	} else {
		l.Debug(fmt.Sprintf("Service call successful: %s.%s", service, method), logContext)
	}
}

// Enhanced API request logging with comprehensive metrics
func (l *Logger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"method":       method,
		"path":         path,
		"status_code":  statusCode,
		"duration_ms":  duration.Milliseconds(),
		"type":         "api_request",
		"component":    "handler",
	}
	
	// Add performance categories
	switch {
	case duration.Milliseconds() > 5000:
		logContext["performance"] = "very_slow"
	case duration.Milliseconds() > 2000:
		logContext["performance"] = "slow"
	case duration.Milliseconds() > 1000:
		logContext["performance"] = "moderate"
	default:
		logContext["performance"] = "fast"
	}
	
	// Merge additional context
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
	}
	
	// Log at appropriate level based on status code and performance
	switch {
	case statusCode >= 500:
		l.Error("API request completed with server error", logContext)
	case statusCode >= 400:
		l.Warning("API request completed with client error", logContext)
	case duration.Milliseconds() > 2000:
		l.Warning("API request completed successfully but slowly", logContext)
	default:
		l.Info("API request completed successfully", logContext)
	}
}

// Enhanced validation error logging with field analysis
func (l *Logger) LogValidationError(field, message string, value interface{}) {
	context := map[string]interface{}{
		"field":     field,
		"message":   message,
		"type":      "validation_error",
		"component": "validator",
	}
	
	// Safely handle the value (mask sensitive fields)
	fieldLower := strings.ToLower(field)
	if fieldLower == "password" || fieldLower == "token" || fieldLower == "secret" {
		context["value"] = "***hidden***"
		context["value_length"] = getValueLength(value)
	} else {
		context["value"] = value
		context["value_type"] = fmt.Sprintf("%T", value)
	}
	
	l.Warning("Validation error occurred", context)
}

// Business logic logging methods

// Log user activity for audit trails
func (l *Logger) LogUserActivity(userID, email, action string, resource string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"user_id":   userID,
		"email":     maskEmail(email),
		"action":    action,
		"resource":  resource,
		"type":      "user_activity",
		"component": "audit",
	}
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
	}
	
	l.Info("User activity logged", logContext)
}

// Log security events for monitoring
func (l *Logger) LogSecurityEvent(eventType, description string, severity string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"event_type":     eventType,
		"description":    description,
		"severity":       severity,
		"type":           "security_event",
		"component":      "security",
		"security_event": true, // Flag for security monitoring systems
	}
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
	}
	
	// Log at appropriate level based on severity
	switch strings.ToLower(severity) {
	case "critical", "high":
		l.Error(fmt.Sprintf("Security event: %s", description), logContext)
	case "medium":
		l.Warning(fmt.Sprintf("Security event: %s", description), logContext)
	default:
		l.Info(fmt.Sprintf("Security event: %s", description), logContext)
	}
}

// Log business metrics and KPIs
func (l *Logger) LogMetric(metricName string, value interface{}, unit string, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"metric_name": metricName,
		"value":       value,
		"unit":        unit,
		"type":        "metric",
		"component":   "metrics",
	}
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
	}
	
	l.Info(fmt.Sprintf("Metric recorded: %s = %v %s", metricName, value, unit), logContext)
}

// Log performance benchmarks
func (l *Logger) LogPerformance(operation string, duration time.Duration, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"operation":    operation,
		"duration_ms":  duration.Milliseconds(),
		"duration_ns":  duration.Nanoseconds(),
		"type":         "performance",
		"component":    "benchmark",
	}
	
	// Add performance categories
	switch {
	case duration.Milliseconds() > 10000:
		logContext["category"] = "critical_slow"
	case duration.Milliseconds() > 5000:
		logContext["category"] = "very_slow"
	case duration.Milliseconds() > 2000:
		logContext["category"] = "slow"
	case duration.Milliseconds() > 1000:
		logContext["category"] = "moderate"
	case duration.Milliseconds() > 500:
		logContext["category"] = "acceptable"
	default:
		logContext["category"] = "fast"
	}
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
	}
	
	// Log at appropriate level based on performance
	if duration.Milliseconds() > 5000 {
		l.Warning(fmt.Sprintf("Performance: %s took %v", operation, duration), logContext)
	} else {
		l.Debug(fmt.Sprintf("Performance: %s took %v", operation, duration), logContext)
	}
}

// Helper functions for enhanced logging

// Mask email for privacy while keeping it useful for debugging
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***invalid_email***"
	}
	
	username := parts[0]
	domain := parts[1]
	
	// Mask username but keep first and last character if long enough
	var maskedUsername string
	if len(username) <= 2 {
		maskedUsername = "**"
	} else if len(username) <= 4 {
		maskedUsername = string(username[0]) + "**" + string(username[len(username)-1])
	} else {
		maskedUsername = string(username[0]) + "***" + string(username[len(username)-1])
	}
	
	return maskedUsername + "@" + domain
}

// Get value length safely
func getValueLength(value interface{}) int {
	if value == nil {
		return 0
	}
	
	switch v := value.(type) {
	case string:
		return len(v)
	case []byte:
		return len(v)
	default:
		return len(fmt.Sprintf("%v", v))
	}
}

// Check if error is retryable (enhance based on your error types)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errMsg := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"service unavailable",
		"deadline exceeded",
		"context deadline exceeded",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}
	
	return false
}

// Environment detection
func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development" // Default
	}
	return env
}

// Get minimum log level based on environment
func getMinLogLevel(environment string) int {
	switch strings.ToLower(environment) {
	case "production", "prod":
		return InfoLevel
	case "staging", "stage":
		return InfoLevel
	case "testing", "test":
		return DebugLevel
	default: // development
		return DebugLevel
	}
}

// Global logger instance with enhanced initialization
var GlobalLogger = NewLogger()

// Convenience functions for global logger usage
func Debug(message string, context ...map[string]interface{}) {
	GlobalLogger.Debug(message, context...)
}

func Info(message string, context ...map[string]interface{}) {
	GlobalLogger.Info(message, context...)
}

func Warning(message string, context ...map[string]interface{}) {
	GlobalLogger.Warning(message, context...)
}

func Error(message string, context ...map[string]interface{}) {
	GlobalLogger.Error(message, context...)
}

func Fatal(message string, context ...map[string]interface{}) {
	GlobalLogger.Fatal(message, context...)
}

// Global convenience functions for specialized logging
func LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {
	GlobalLogger.LogAuthAttempt(email, success, reason, additionalContext...)
}

func LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {
	GlobalLogger.LogDBOperation(operation, table, success, err, context)
}

func LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {
	GlobalLogger.LogServiceCall(service, method, success, err, context)
}

func LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {
	GlobalLogger.LogAPIRequest(method, path, statusCode, duration, context)
}

func LogValidationError(field, message string, value interface{}) {
	GlobalLogger.LogValidationError(field, message, value)
}

func LogUserActivity(userID, email, action, resource string, context map[string]interface{}) {
	GlobalLogger.LogUserActivity(userID, email, action, resource, context)
}

func LogSecurityEvent(eventType, description, severity string, context map[string]interface{}) {
	GlobalLogger.LogSecurityEvent(eventType, description, severity, context)
}

func LogMetric(metricName string, value interface{}, unit string, context map[string]interface{}) {
	GlobalLogger.LogMetric(metricName, value, unit, context)
}

func LogPerformance(operation string, duration time.Duration, context map[string]interface{}) {
	GlobalLogger.LogPerformance(operation, duration, context)
}

// Configuration functions for global logger
func SetJSONLogging(enable bool) {
	GlobalLogger.SetJSONLogging(enable)
}

func SetDebugLogging(enable bool) {
	GlobalLogger.SetDebugLogging(enable)
}

func SetMinLevel(level int) {
	GlobalLogger.SetMinLevel(level)
}

func SetComponent(component string) {
	GlobalLogger.SetComponent(component)
}

func AddGlobalField(key string, value interface{}) {
	GlobalLogger.AddGlobalField(key, value)
}

func RemoveGlobalField(key string) {
	GlobalLogger.RemoveGlobalField(key)
}

// Logger initialization for different components
func NewComponentLogger(component string) *Logger {
	logger := NewLogger()
	logger.SetComponent(component)
	return logger
}

// Structured logging for different layers
func NewHandlerLogger() *Logger {
	return NewComponentLogger("handler")
}

func NewServiceLogger() *Logger {
	return NewComponentLogger("service")
}

func NewRepositoryLogger() *Logger {
	return NewComponentLogger("repository")
}

func NewMiddlewareLogger() *Logger {
	return NewComponentLogger("middleware")
}
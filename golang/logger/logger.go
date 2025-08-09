// logger/logger.go - Enhanced version with improved readability and error tracking
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
	WarningLevel: "WARN",
	ErrorLevel:   "ERROR",
	FatalLevel:   "FATAL",
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
	Component    string                 `json:"component,omitempty"`
	Operation    string                 `json:"operation,omitempty"`
	Duration     int64                  `json:"duration_ms,omitempty"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	Environment  string                 `json:"environment,omitempty"`
	// New fields for enhanced error tracking
	Cause        string                 `json:"cause,omitempty"`
	Layer        string                 `json:"layer,omitempty"`
}

// Logger structure with enhanced capabilities and thread safety
type Logger struct {
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
	outputFormat  string
	enableDebug   bool
	minLevel      int
	environment   string
	component     string
	layer         string
	operation     string
	mutex         sync.RWMutex
	contextFields map[string]interface{}
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
		outputFormat:  getOutputFormat(environment),
		enableDebug:   environment == "development",
		minLevel:      getMinLogLevel(environment),
		environment:   environment,
		contextFields: make(map[string]interface{}),
	}
	
	// Set initial global context
	logger.contextFields["environment"] = environment
	logger.contextFields["service"] = "restaurant-api"
	
	return logger
}

// Configuration methods with thread safety
func (l *Logger) SetOutputFormat(format string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.outputFormat = format
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

func (l *Logger) SetLayer(layer string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.layer = layer
}

func (l *Logger) SetOperation(operation string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.operation = operation
}

func (l *Logger) AddGlobalField(key string, value interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.contextFields[key] = value
}

func (l *Logger) RemoveGlobalField(key string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.contextFields, key)
}

// Enhanced caller info retrieval
func (l *Logger) getCallerInfo(skip int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", "", 0
	}
	
	function := runtime.FuncForPC(pc).Name()
	
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}
	
	if lastDot := strings.LastIndex(function, "."); lastDot >= 0 {
		function = function[lastDot+1:]
	}
	
	return file, function, line
}

// Enhanced context merging
func (l *Logger) mergeContext(baseContext, additionalContext map[string]interface{}) map[string]interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	merged := make(map[string]interface{})
	
	for k, v := range l.contextFields {
		merged[k] = v
	}
	
	if baseContext != nil {
		for k, v := range baseContext {
			merged[k] = v
		}
	}
	
	if additionalContext != nil {
		for k, v := range additionalContext {
			merged[k] = v
		}
	}
	
	return merged
}

// Format log entry as pretty text with enhanced error info
func (l *Logger) formatPretty(entry LogEntry) string {
	timestamp := time.Now().Format("15:04:05.000")
	
	// Get level emoji and color
	levelDisplay := formatLevel(entry.Level)
	
	// Build main message
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	parts = append(parts, levelDisplay)
	
	// Add layer if present
	if entry.Layer != "" {
		parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(entry.Layer)))
	}
	
	// Add component if present
	if entry.Component != "" {
		parts = append(parts, fmt.Sprintf("<%s>", entry.Component))
	}
	
	// Add operation if present
	if entry.Operation != "" {
		parts = append(parts, fmt.Sprintf("{%s}", entry.Operation))
	}
	
	parts = append(parts, entry.Message)
	
	mainLine := strings.Join(parts, " ")
	
	// Add important context on the same line
	var contextParts []string
	
	// Add key identifiers
	if email, ok := entry.Context["email"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("user=%v", email))
	}
	if ip, ok := entry.Context["ip"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("ip=%v", ip))
	}
	if entry.Duration > 0 {
		contextParts = append(contextParts, fmt.Sprintf("took=%dms", entry.Duration))
	}
	if statusCode, ok := entry.Context["status_code"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("status=%v", statusCode))
	}
	if reason, ok := entry.Context["failure_reason"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("reason=%v", reason))
	}
	if errorMsg, ok := entry.Context["error"]; ok {
		contextParts = append(contextParts, fmt.Sprintf("error=%v", errorMsg))
	}
	
	// Add cause if present (for errors)
	if entry.Cause != "" {
		contextParts = append(contextParts, fmt.Sprintf("cause=%s", entry.Cause))
	}
	
	if len(contextParts) > 0 {
		mainLine += " | " + strings.Join(contextParts, " ")
	}
	
	// Add file/line info for debug/error levels
	if entry.Level == "DEBUG" || entry.Level == "ERROR" {
		if entry.File != "" {
			mainLine += fmt.Sprintf(" (%s:%d)", entry.File, entry.Line)
		}
	}
	
	return mainLine
}

// Format log entry as simple text with enhanced error info
func (l *Logger) formatText(entry LogEntry) string {
	timestamp := time.Now().Format("15:04:05")
	
	// Build message with layer and operation info
	var msgParts []string
	if entry.Layer != "" {
		msgParts = append(msgParts, fmt.Sprintf("[%s]", entry.Layer))
	}
	if entry.Operation != "" {
		msgParts = append(msgParts, fmt.Sprintf("{%s}", entry.Operation))
	}
	msgParts = append(msgParts, entry.Message)
	
	msg := fmt.Sprintf("[%s] %s: %s", timestamp, entry.Level, strings.Join(msgParts, " "))
	
	// Add minimal essential context
	if entry.Context != nil {
		var essentials []string
		
		// Only show really important stuff
		if email, ok := entry.Context["email"]; ok {
			essentials = append(essentials, fmt.Sprintf("user=%v", email))
		}
		if entry.Duration > 0 {
			essentials = append(essentials, fmt.Sprintf("%dms", entry.Duration))
		}
		if errorMsg, ok := entry.Context["error"]; ok {
			essentials = append(essentials, fmt.Sprintf("error=%v", errorMsg))
		}
		if entry.Cause != "" {
			essentials = append(essentials, fmt.Sprintf("cause=%s", entry.Cause))
		}
		
		if len(essentials) > 0 {
			msg += " (" + strings.Join(essentials, " ") + ")"
		}
	}
	
	return msg
}

// Core logging method with multiple output formats and enhanced error tracking
func (l *Logger) logWithContext(level int, message string, context map[string]interface{}, skip int) {
	l.mutex.RLock()
	outputFormat := l.outputFormat
	enableDebug := l.enableDebug
	minLevel := l.minLevel
	component := l.component
	layer := l.layer
	operation := l.operation
	environment := l.environment
	l.mutex.RUnlock()
	
	if level < minLevel {
		return
	}
	
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
	
	// Create log entry
	file, function, line := l.getCallerInfo(skip + 1)
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
		Layer:       layer,
		Operation:   operation,
		Environment: environment,
	}
	
	// Extract special fields from context
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
		if contextOperation, ok := mergedContext["operation"].(string); ok {
			entry.Operation = contextOperation
		}
		if contextLayer, ok := mergedContext["layer"].(string); ok {
			entry.Layer = contextLayer
		}
		if cause, ok := mergedContext["cause"].(string); ok {
			entry.Cause = cause
		}
		if duration, ok := mergedContext["duration_ms"].(int64); ok {
			entry.Duration = duration
		}
		if errorCode, ok := mergedContext["error_code"].(string); ok {
			entry.ErrorCode = errorCode
		}
	}
	
	// Output based on format
	switch outputFormat {
	case FormatJSON:
		jsonData, err := json.Marshal(entry)
		if err != nil {
			logger.Printf("[%s] %s | JSON Error: %v", levelStr, message, err)
		} else {
			logger.Println(string(jsonData))
		}
	case FormatPretty:
		logger.Println(l.formatPretty(entry))
	case FormatText:
		logger.Println(l.formatText(entry))
	default:
		logger.Println(l.formatPretty(entry)) // Default to pretty
	}
	
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Format level with emoji and colors
func formatLevel(level string) string {
	switch level {
	case "DEBUG":
		return "🔍 DEBUG"
	case "INFO":
		return "ℹ️  INFO"
	case "WARN":
		return "⚠️  WARN"
	case "ERROR":
		return "❌ ERROR"
	case "FATAL":
		return "💀 FATAL"
	default:
		return level
	}
}

// Enhanced logging methods with layer, operation, and cause support
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

// New enhanced error logging methods
func (l *Logger) ErrorWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {
	ctx := map[string]interface{}{
		"cause":     cause,
		"layer":     layer,
		"operation": operation,
	}
	
	if len(context) > 0 {
		for k, v := range context[0] {
			ctx[k] = v
		}
	}
	
	l.logWithContext(ErrorLevel, message, ctx, 3)
}

func (l *Logger) WarningWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {
	ctx := map[string]interface{}{
		"cause":     cause,
		"layer":     layer,
		"operation": operation,
	}
	
	if len(context) > 0 {
		for k, v := range context[0] {
			ctx[k] = v
		}
	}
	
	l.logWithContext(WarningLevel, message, ctx, 3)
}

func (l *Logger) InfoWithOperation(message string, layer string, operation string, context ...map[string]interface{}) {
	ctx := map[string]interface{}{
		"layer":     layer,
		"operation": operation,
	}
	
	if len(context) > 0 {
		for k, v := range context[0] {
			ctx[k] = v
		}
	}
	
	l.logWithContext(InfoLevel, message, ctx, 3)
}

// Specialized logging methods (enhanced with new fields)

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
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
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
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
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
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
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
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
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
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
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
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
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
	
	if context != nil {
		for k, v := range context {
			logContext[k] = v
		}
	}
	
	message := fmt.Sprintf("Performance: %s took %v", operation, duration)
	
	if duration.Milliseconds() > 5000 {
		l.Warning(message, logContext)
	} else {
		l.Debug(message, logContext)
	}
}

// Helper functions (enhanced with error categorization)
func categorizeError(err error) string {
	if err == nil {
		return ""
	}
	
	errMsg := strings.ToLower(err.Error())
	
	// Network related errors
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "network") {
		return "network_error"
	}
	
	// Timeout errors
	if strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline exceeded") {
		return "timeout_error"
	}
	
	// Authentication errors
	if strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "authentication") ||
		strings.Contains(errMsg, "invalid token") {
		return "auth_error"
	}
	
	// Validation errors
	if strings.Contains(errMsg, "validation") ||
		strings.Contains(errMsg, "invalid input") ||
		strings.Contains(errMsg, "bad request") {
		return "validation_error"
	}
	
	// Permission errors
	if strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "access denied") {
		return "permission_error"
	}
	
	// Resource errors
	if strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") {
		return "resource_not_found"
	}
	
	// Conflict errors
	if strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "conflict") {
		return "resource_conflict"
	}
	
	// External service errors
	if strings.Contains(errMsg, "service unavailable") ||
		strings.Contains(errMsg, "bad gateway") {
		return "external_service_error"
	}
	
	return "unknown_error"
}

func categorizeDBError(err error) string {
	if err == nil {
		return ""
	}
	
	errMsg := strings.ToLower(err.Error())
	
	// Connection errors
	if strings.Contains(errMsg, "connection") {
		return "db_connection_error"
	}
	
	// Constraint violations
	if strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "primary key") {
		return "db_constraint_violation"
	}
	
	// Syntax errors
	if strings.Contains(errMsg, "syntax error") ||
		strings.Contains(errMsg, "invalid sql") {
		return "db_syntax_error"
	}
	
	// Data errors
	if strings.Contains(errMsg, "data too long") ||
		strings.Contains(errMsg, "out of range") {
		return "db_data_error"
	}
	
	// Transaction errors
	if strings.Contains(errMsg, "deadlock") ||
		strings.Contains(errMsg, "lock timeout") {
		return "db_transaction_error"
	}
	
	// Permission errors
	if strings.Contains(errMsg, "access denied") ||
		strings.Contains(errMsg, "permission denied") {
		return "db_permission_error"
	}
	
	return "db_unknown_error"
}

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

func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func getOutputFormat(environment string) string {
	format := os.Getenv("LOG_FORMAT")
	if format != "" {
		return format
	}
	
	// Default formats based on environment
	switch strings.ToLower(environment) {
	case "production", "prod":
		return FormatJSON
	case "staging", "stage":
		return FormatJSON
	default: // development, testing
		return FormatPretty
	}
}

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

// Global logger instance
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

// Enhanced global convenience functions
func ErrorWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {
	GlobalLogger.ErrorWithCause(message, cause, layer, operation, context...)
}

func WarningWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {
	GlobalLogger.WarningWithCause(message, cause, layer, operation, context...)
}

func InfoWithOperation(message string, layer string, operation string, context ...map[string]interface{}) {
	GlobalLogger.InfoWithOperation(message, layer, operation, context...)
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

// Configuration functions
func SetOutputFormat(format string) {
	GlobalLogger.SetOutputFormat(format)
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

func SetLayer(layer string) {
	GlobalLogger.SetLayer(layer)
}

func SetOperation(operation string) {
	GlobalLogger.SetOperation(operation)
}

func AddGlobalField(key string, value interface{}) {
	GlobalLogger.AddGlobalField(key, value)
}

func RemoveGlobalField(key string) {
	GlobalLogger.RemoveGlobalField(key)
}

// Enhanced component loggers with layer support
func NewComponentLogger(component string) *Logger {
	logger := NewLogger()
	logger.SetComponent(component)
	return logger
}

func NewLayerLogger(layer string) *Logger {
	logger := NewLogger()
	logger.SetLayer(layer)
	return logger
}

func NewHandlerLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("handler")
	logger.SetLayer(LayerHandler)
	return logger
}

func NewServiceLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("service")
	logger.SetLayer(LayerService)
	return logger
}

func NewRepositoryLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("repository")
	logger.SetLayer(LayerRepository)
	return logger
}

func NewMiddlewareLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("middleware")
	logger.SetLayer(LayerMiddleware)
	return logger
}

func NewAuthLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("auth")
	logger.SetLayer(LayerAuth)
	return logger
}

func NewValidationLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("validation")
	logger.SetLayer(LayerValidation)
	return logger
}

func NewCacheLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("cache")
	logger.SetLayer(LayerCache)
	return logger
}

func NewDatabaseLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("database")
	logger.SetLayer(LayerDatabase)
	return logger
}

func NewExternalLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("external")
	logger.SetLayer(LayerExternal)
	return logger
}

func NewSecurityLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("security")
	logger.SetLayer(LayerSecurity)
	return logger
}
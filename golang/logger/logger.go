package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// Logger levels
const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	File      string                 `json:"file,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Line      int                    `json:"line,omitempty"`
}

// Logger structure with enhanced capabilities
type Logger struct {
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
	enableJSON    bool
	enableDebug   bool
}

// NewLogger creates a new enhanced Logger instance
func NewLogger() *Logger {
	return &Logger{
		debugLogger:   log.New(os.Stdout, "", 0),
		infoLogger:    log.New(os.Stdout, "", 0),
		warningLogger: log.New(os.Stdout, "", 0),
		errorLogger:   log.New(os.Stderr, "", 0),
		fatalLogger:   log.New(os.Stderr, "", 0),
		enableJSON:    true,  // Enable JSON logging for better parsing
		enableDebug:   true,  // Enable debug logs in development
	}
}

// SetJSONLogging enables or disables JSON formatted logging
func (l *Logger) SetJSONLogging(enable bool) {
	l.enableJSON = enable
}

// SetDebugLogging enables or disables debug logging
func (l *Logger) SetDebugLogging(enable bool) {
	l.enableDebug = enable
}

// getCallerInfo retrieves information about the calling function
func (l *Logger) getCallerInfo() (string, string, int) {
	pc, file, line, ok := runtime.Caller(3) // Skip 3 frames: getCallerInfo, logWithContext, and the public method
	if !ok {
		return "", "", 0
	}
	
	function := runtime.FuncForPC(pc).Name()
	return file, function, line
}

// logWithContext logs a message with additional context
func (l *Logger) logWithContext(level int, message string, context map[string]interface{}) {
	var levelStr string
	var logger *log.Logger
	
	switch level {
	case DebugLevel:
		if !l.enableDebug {
			return
		}
		levelStr = "DEBUG"
		logger = l.debugLogger
	case InfoLevel:
		levelStr = "INFO"
		logger = l.infoLogger
	case WarningLevel:
		levelStr = "WARNING"
		logger = l.warningLogger
	case ErrorLevel:
		levelStr = "ERROR"
		logger = l.errorLogger
	case FatalLevel:
		levelStr = "FATAL"
		logger = l.fatalLogger
	default:
		levelStr = "INFO"
		logger = l.infoLogger
	}
	
	if l.enableJSON {
		file, function, line := l.getCallerInfo()
		
		entry := LogEntry{
			Timestamp: time.Now().Format("2006-01-02 15:04:05.000"),
			Level:     levelStr,
			Message:   message,
			Context:   context,
			File:      file,
			Function:  function,
			Line:      line,
		}
		
		jsonData, err := json.Marshal(entry)
		if err != nil {
			// Fallback to simple logging if JSON fails
			logger.Printf("[%s] %s %v", levelStr, message, context)
		} else {
			logger.Println(string(jsonData))
		}
	} else {
		if context != nil && len(context) > 0 {
			logger.Printf("[%s] %s | Context: %+v", levelStr, message, context)
		} else {
			logger.Printf("[%s] %s", levelStr, message)
		}
	}
	
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug logs a debug message with optional context
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(DebugLevel, message, ctx)
}

// Info logs an info message with optional context
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(InfoLevel, message, ctx)
}

// Warning logs a warning message with optional context
func (l *Logger) Warning(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(WarningLevel, message, ctx)
}

// Error logs an error message with optional context
func (l *Logger) Error(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(ErrorLevel, message, ctx)
}

// Fatal logs a fatal message with optional context and exits
func (l *Logger) Fatal(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}
	l.logWithContext(FatalLevel, message, ctx)
}

// Authentication specific logging methods
func (l *Logger) LogAuthAttempt(email string, success bool, reason string) {
	context := map[string]interface{}{
		"email":   email,
		"success": success,
		"reason":  reason,
		"type":    "auth_attempt",
	}
	
	if success {
		l.Info("Authentication attempt successful", context)
	} else {
		l.Warning("Authentication attempt failed", context)
	}
}

// Database operation logging
func (l *Logger) LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"operation": operation,
		"table":     table,
		"success":   success,
		"type":      "db_operation",
	}
	
	// Merge additional context
	for k, v := range context {
		logContext[k] = v
	}
	
	if err != nil {
		logContext["error"] = err.Error()
		l.Error(fmt.Sprintf("Database %s operation failed on %s", operation, table), logContext)
	} else if success {
		l.Debug(fmt.Sprintf("Database %s operation successful on %s", operation, table), logContext)
	}
}

// Service call logging
func (l *Logger) LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"service": service,
		"method":  method,
		"success": success,
		"type":    "service_call",
	}
	
	// Merge additional context
	for k, v := range context {
		logContext[k] = v
	}
	
	if err != nil {
		logContext["error"] = err.Error()
		l.Error(fmt.Sprintf("Service call failed: %s.%s", service, method), logContext)
	} else {
		l.Debug(fmt.Sprintf("Service call successful: %s.%s", service, method), logContext)
	}
}

// API request logging
func (l *Logger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {
	logContext := map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"type":        "api_request",
	}
	
	// Merge additional context
	for k, v := range context {
		logContext[k] = v
	}
	
	if statusCode >= 400 {
		l.Warning("API request completed with error", logContext)
	} else {
		l.Info("API request completed successfully", logContext)
	}
}

// Validation error logging
func (l *Logger) LogValidationError(field, message string, value interface{}) {
	context := map[string]interface{}{
		"field":   field,
		"message": message,
		"value":   value,
		"type":    "validation_error",
	}
	l.Warning("Validation error occurred", context)
}

// Global logger instance
var GlobalLogger = NewLogger()

// Convenience functions for global logger
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

func LogAuthAttempt(email string, success bool, reason string) {
	GlobalLogger.LogAuthAttempt(email, success, reason)
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
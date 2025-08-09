// logger/core.go - Core logger implementation and main functionality
package logger

import (
	"encoding/json"
	
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

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

// Basic logging methods
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

// Enhanced error logging methods
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
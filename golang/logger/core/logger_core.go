// internal/logger/core/types.go - Enhanced core types and structures
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

func getCaller(skip int) string {
	// Implementation would use runtime.Caller to get file:line info
	// Simplified for this example
	return ""
}
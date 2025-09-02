// internal/logger/core/logger_core.go - Simplified core types and structures
package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

// Output interface for different output destinations
type Output interface {
	Write(entry *LogEntry) error
	Close() error
}

// OutputManager interface for managing multiple outputs
type OutputManager interface {
	WriteToAll(entry *LogEntry) error
	WriteToOutput(name string, entry *LogEntry) error
	AddOutput(name string, output Output) error
	RemoveOutput(name string) error
	Close() error
}

// Logger represents the main logger with enhanced capabilities
type CoreLogger struct {
	level         Level
	outputManager OutputManager
	contextFields map[string]interface{}
	component     string
	layer         string
	operation     string
	environment   string
	mu            sync.RWMutex
}

// NewLogger creates a new enhanced logger instance
func NewLogger() *CoreLogger {
	logger := &CoreLogger{
		level:         InfoLevel,
		contextFields: make(map[string]interface{}),
		environment:   "development",
	}
	
	// Create a default console output manager with rich formatting
	outputManager := NewDefaultOutputManager()
	logger.SetOutputManager(outputManager)
	
	return logger
}

// SetOutputManager sets the output manager for this logger
func (l *CoreLogger) SetOutputManager(om OutputManager) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.outputManager = om
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
	
	om := l.outputManager
	l.mu.RUnlock()
	
	// Write to outputs
	if om != nil {
		om.WriteToAll(entry)
	} else {
		// Fallback to simple console if no output manager
		fmt.Fprintf(os.Stdout, "[%s] %s %s\n", 
			entry.Timestamp.Format("2006-01-02 15:04:05"), 
			entry.Level.String(), 
			entry.Message)
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
	// Simple implementation - you can enhance this with runtime.Caller
	return ""
}

// RichConsoleOutput - Simplified but rich console output
type RichConsoleOutput struct {
	useColors bool
	mu        sync.Mutex
}

func NewRichConsoleOutput(useColors bool) *RichConsoleOutput {
	return &RichConsoleOutput{
		useColors: useColors,
	}
}

func (rco *RichConsoleOutput) Write(entry *LogEntry) error {
	rco.mu.Lock()
	defer rco.mu.Unlock()
	
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	
	// Build the log line with all information
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	
	// Level with color
	levelStr := entry.Level.String()
	if rco.useColors {
		levelStr = rco.colorizeLevel(entry.Level, levelStr)
	}
	parts = append(parts, levelStr)
	
	// Layer and Component
	if entry.Layer != "" {
		layerStr := fmt.Sprintf("[%s]", strings.ToUpper(entry.Layer))
		if rco.useColors {
			layerStr = rco.colorize("\033[94m", layerStr) // Light blue
		}
		parts = append(parts, layerStr)
	}
	
	if entry.Component != "" {
		componentStr := fmt.Sprintf("<%s>", entry.Component)
		if rco.useColors {
			componentStr = rco.colorize("\033[95m", componentStr) // Magenta
		}
		parts = append(parts, componentStr)
	}
	
	if entry.Operation != "" {
		operationStr := fmt.Sprintf("{%s}", entry.Operation)
		if rco.useColors {
			operationStr = rco.colorize("\033[96m", operationStr) // Cyan
		}
		parts = append(parts, operationStr)
	}
	
	// Message
	parts = append(parts, entry.Message)
	
	// Join main parts
	logLine := strings.Join(parts, " ")
	
	// Add fields as JSON on the same line if they exist
	if len(entry.Fields) > 0 {
		fieldsJSON, err := json.Marshal(entry.Fields)
		if err == nil {
			fieldStr := fmt.Sprintf(" | %s", string(fieldsJSON))
			if rco.useColors {
				fieldStr = rco.colorize("\033[90m", fieldStr) // Dark gray
			}
			logLine += fieldStr
		}
	}
	
	// Write to appropriate stream
	if entry.Level >= ErrorLevel {
		fmt.Fprintf(os.Stderr, "%s\n", logLine)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", logLine)
	}
	
	return nil
}

func (rco *RichConsoleOutput) colorizeLevel(level Level, text string) string {
	if !rco.useColors {
		return text
	}
	
	var color string
	switch level {
	case DebugLevel:
		color = "\033[36m" // Cyan
	case InfoLevel:
		color = "\033[32m" // Green
	case WarnLevel:
		color = "\033[33m" // Yellow
	case ErrorLevel:
		color = "\033[31m" // Red
	case FatalLevel:
		color = "\033[35m\033[1m" // Bold Magenta
	default:
		return text
	}
	
	return color + text + "\033[0m"
}

func (rco *RichConsoleOutput) colorize(color, text string) string {
	if !rco.useColors {
		return text
	}
	return color + text + "\033[0m"
}

func (rco *RichConsoleOutput) Close() error {
	return nil
}

// Default output manager implementation
type defaultOutputManager struct {
	outputs map[string]Output
	mu      sync.RWMutex
}

func NewDefaultOutputManager() OutputManager {
	manager := &defaultOutputManager{
		outputs: make(map[string]Output),
	}
	
	// Add a rich console output by default
	consoleOutput := NewRichConsoleOutput(true) // Enable colors
	manager.AddOutput("console", consoleOutput)
	
	return manager
}

func (dom *defaultOutputManager) AddOutput(name string, output Output) error {
	dom.mu.Lock()
	defer dom.mu.Unlock()
	dom.outputs[name] = output
	return nil
}

func (dom *defaultOutputManager) RemoveOutput(name string) error {
	dom.mu.Lock()
	defer dom.mu.Unlock()
	
	if output, exists := dom.outputs[name]; exists {
		output.Close()
		delete(dom.outputs, name)
	}
	return nil
}

func (dom *defaultOutputManager) WriteToOutput(name string, entry *LogEntry) error {
	dom.mu.RLock()
	output, exists := dom.outputs[name]
	dom.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("output %s not found", name)
	}
	
	return output.Write(entry)
}

func (dom *defaultOutputManager) WriteToAll(entry *LogEntry) error {
	dom.mu.RLock()
	outputs := make([]Output, 0, len(dom.outputs))
	for _, output := range dom.outputs {
		outputs = append(outputs, output)
	}
	dom.mu.RUnlock()
	
	var errors []error
	for _, output := range outputs {
		if err := output.Write(entry); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to write to %d outputs: %v", len(errors), errors)
	}
	
	return nil
}

func (dom *defaultOutputManager) Close() error {
	dom.mu.Lock()
	defer dom.mu.Unlock()
	
	var errors []error
	for name, output := range dom.outputs {
		if err := output.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close output %s: %w", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to close outputs: %v", errors)
	}
	
	return nil
}
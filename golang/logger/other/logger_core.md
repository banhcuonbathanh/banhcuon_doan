// logger/core/logger_core.go - Enhanced with better caller tracking and error context
package core

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
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

// Enhanced CallerInfo for better error tracking
type CallerInfo struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Package  string `json:"package"`
}

func (c CallerInfo) String() string {
	return fmt.Sprintf("%s:%d", c.File, c.Line)
}

// LogEntry represents a structured log entry with enhanced metadata
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       Level                  `json:"level"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Caller      *CallerInfo            `json:"caller,omitempty"`
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
	StackTrace  []CallerInfo           `json:"stack_trace,omitempty"`
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
	captureStack  bool
	stackDepth    int
	mu            sync.RWMutex
}

// NewLogger creates a new enhanced logger instance
func NewLogger() *CoreLogger {
	logger := &CoreLogger{
		level:         InfoLevel,
		contextFields: make(map[string]interface{}),
		environment:   "development",
		captureStack:  true,  // Enable stack capture for better debugging
		stackDepth:    5,     // Capture up to 5 stack frames for errors
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

func (l *CoreLogger) SetStackCapture(enabled bool, depth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.captureStack = enabled
	if depth > 0 {
		l.stackDepth = depth
	}
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
	l.log(DebugLevel, message, 3, fields...)
}

func (l *CoreLogger) Info(message string, fields ...map[string]interface{}) {
	l.log(InfoLevel, message, 3, fields...)
}

func (l *CoreLogger) Warn(message string, fields ...map[string]interface{}) {
	l.log(WarnLevel, message, 3, fields...)
}

func (l *CoreLogger) Error(message string, fields ...map[string]interface{}) {
	l.log(ErrorLevel, message, 3, fields...)
}

func (l *CoreLogger) Fatal(message string, fields ...map[string]interface{}) {
	l.log(FatalLevel, message, 3, fields...)
	// Note: In production, this might call os.Exit(1)
}

func (l *CoreLogger) InfoWithOperation(message, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["layer"] = layer
	mergedFields["operation"] = operation
	l.log(InfoLevel, message, 3, mergedFields)
}

// Enhanced ErrorWithCause method with better caller tracking
func (l *CoreLogger) ErrorWithCause(message, cause, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["cause"] = cause
	mergedFields["layer"] = layer  
	mergedFields["operation"] = operation
	
	// Add domain if not already present
	if _, exists := mergedFields["domain"]; !exists {
		// Try to infer domain from layer
		switch layer {
		case LayerAuth:
			mergedFields["domain"] = "auth"
		case LayerDatabase:
			mergedFields["domain"] = "database"
		case LayerExternal:
			mergedFields["domain"] = "external"
		case LayerValidation:
			mergedFields["domain"] = "validation"
		case LayerSecurity:
			mergedFields["domain"] = "security"
		case LayerHandler:
			mergedFields["domain"] = "handler"
		case LayerService:
			mergedFields["domain"] = "service"
		case LayerRepository:
			mergedFields["domain"] = "repository"
		default:
			mergedFields["domain"] = "system"
		}
	}
	
	l.log(ErrorLevel, message, 3, mergedFields)
}

// Enhanced WarnWithCause method
func (l *CoreLogger) WarnWithCause(message, cause, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["cause"] = cause
	mergedFields["layer"] = layer
	mergedFields["operation"] = operation
	
	// Add domain if not already present
	if _, exists := mergedFields["domain"]; !exists {
		// Try to infer domain from layer
		switch layer {
		case LayerAuth:
			mergedFields["domain"] = "auth"
		case LayerDatabase:
			mergedFields["domain"] = "database"
		case LayerExternal:
			mergedFields["domain"] = "external"
		case LayerValidation:
			mergedFields["domain"] = "validation"
		case LayerSecurity:
			mergedFields["domain"] = "security"
		case LayerHandler:
			mergedFields["domain"] = "handler"
		case LayerService:
			mergedFields["domain"] = "service"
		case LayerRepository:
			mergedFields["domain"] = "repository"
		default:
			mergedFields["domain"] = "system"
		}
	}
	
	l.log(WarnLevel, message, 3, mergedFields)
}

// New method for enhanced error logging with domain
func (l *CoreLogger) ErrorWithDomainAndCause(message, domain, cause, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["domain"] = domain
	mergedFields["cause"] = cause
	mergedFields["layer"] = layer  
	mergedFields["operation"] = operation
	
	l.log(ErrorLevel, message, 3, mergedFields)
}

// Core logging implementation with enhanced caller tracking
func (l *CoreLogger) log(level Level, message string, callerSkip int, fields ...map[string]interface{}) {
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
	
	// Enhanced caller information
	if caller := l.getEnhancedCaller(callerSkip + 1); caller != nil {
		entry.Caller = caller
	}
	
	// Capture stack trace for errors and above
	if level >= ErrorLevel && l.captureStack {
		entry.StackTrace = l.captureStackTrace(callerSkip+1, l.stackDepth)
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

// Enhanced caller information gathering
func (l *CoreLogger) getEnhancedCaller(skip int) *CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}
	
	// Get function information
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return nil
	}
	
	fullFuncName := fn.Name()
	
	// Extract package and function name
	var pkg, funcName string
	if lastSlash := strings.LastIndex(fullFuncName, "/"); lastSlash >= 0 {
		pkg = fullFuncName[:lastSlash]
		funcName = fullFuncName[lastSlash+1:]
	} else {
		funcName = fullFuncName
	}
	
	// Clean up the function name (remove package prefix if present)
	if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
		if len(funcName) > lastDot+1 {
			// Keep the package part for context
			pkg = funcName[:lastDot]
			funcName = funcName[lastDot+1:]
		}
	}
	
	// Clean up file path to show only relevant part
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}
	
	return &CallerInfo{
		Function: funcName,
		File:     file,
		Line:     line,
		Package:  pkg,
	}
}

// Capture stack trace for better error debugging
func (l *CoreLogger) captureStackTrace(skip, depth int) []CallerInfo {
	var stack []CallerInfo
	
	for i := skip; i < skip+depth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		fullFuncName := fn.Name()
		
		// Extract package and function name
		var pkg, funcName string
		if lastSlash := strings.LastIndex(fullFuncName, "/"); lastSlash >= 0 {
			pkg = fullFuncName[:lastSlash]
			funcName = fullFuncName[lastSlash+1:]
		} else {
			funcName = fullFuncName
		}
		
		// Clean up the function name
		if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
			if len(funcName) > lastDot+1 {
				pkg = funcName[:lastDot]
				funcName = funcName[lastDot+1:]
			}
		}
		
		// Clean up file path
		if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
			file = file[lastSlash+1:]
		}
		
		stack = append(stack, CallerInfo{
			Function: funcName,
			File:     file,
			Line:     line,
			Package:  pkg,
		})
	}
	
	return stack
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

// Enhanced RichConsoleOutput with better error formatting
type RichConsoleOutput struct {
	useColors   bool
	showCaller  bool
	showStack   bool
	mu          sync.Mutex
}

func NewRichConsoleOutput(useColors bool) *RichConsoleOutput {
	return &RichConsoleOutput{
		useColors:  useColors,
		showCaller: true,
		showStack:  true, // Show stack trace for errors
	}
}

func (rco *RichConsoleOutput) Write(entry *LogEntry) error {
	rco.mu.Lock()
	defer rco.mu.Unlock()
	
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	
	// Build the log line with all information
	var parts []string
	
	// Timestamp with grey color for all levels
	timestampStr := fmt.Sprintf("[%s]", timestamp)
	if rco.useColors {
		timestampStr = rco.colorize("\033[90m", timestampStr) // Dark grey for all timestamps
	}
	parts = append(parts, timestampStr)
	
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
	
	// Enhanced caller information for errors
	if rco.showCaller && entry.Caller != nil && entry.Level >= ErrorLevel {
		callerStr := fmt.Sprintf("@%s.%s:%d", entry.Caller.Package, entry.Caller.Function, entry.Caller.Line)
		if rco.useColors {
			callerStr = rco.colorize("\033[93m", callerStr) // Bright yellow
		}
		parts = append(parts, callerStr)
	}
	
	// Enhanced message with domain and cause for errors/warnings
	message := entry.Message
	if entry.Level >= ErrorLevel {
		// Extract domain and cause from fields for enhanced error display
		domain := ""
		cause := ""
		
		if entry.Fields != nil {
			if d, ok := entry.Fields["domain"].(string); ok {
				domain = d
			}
			if c, ok := entry.Fields["cause"].(string); ok {
				cause = c
			}
		}
		
		// If we have cause from the entry itself, use that
		if entry.Cause != "" {
			cause = entry.Cause
		}
		
		// Build enhanced error message
		if domain != "" && cause != "" {
			message = fmt.Sprintf("%s [domain=%s] [cause=%s]", message, domain, cause)
		} else if domain != "" {
			message = fmt.Sprintf("%s [domain=%s]", message, domain)
		} else if cause != "" {
			message = fmt.Sprintf("%s [cause=%s]", message, cause)
		}
	}
	
	parts = append(parts, message)
	
	// Join main parts
	logLine := strings.Join(parts, " ")
	
	// Add essential fields as JSON on the same line if they exist
	if len(entry.Fields) > 0 {
		// Filter out domain and cause since we already included them in the message for errors
		filteredFields := make(map[string]interface{})
		for k, v := range entry.Fields {
			// Skip domain and cause for error levels since they're in the message
			if entry.Level >= ErrorLevel && (k == "domain" || k == "cause") {
				continue
			}
			// Include other important fields
			if k == "request_id" || k == "user_id" || k == "email" || k == "method" || 
			   k == "endpoint" || k == "status_code" || k == "duration_ms" || 
			   k == "error_code" || k == "service" || k == "layer" || k == "operation" ||
			   k == "table" || k == "function" || k == "attempted_email" || k == "attempted_role" ||
			   k == "error" {
				filteredFields[k] = v
			}
		}
		
		if len(filteredFields) > 0 {
			fieldsJSON, err := json.Marshal(filteredFields)
			if err == nil {
				fieldStr := fmt.Sprintf(" | %s", string(fieldsJSON))
				if rco.useColors {
					fieldStr = rco.colorize("\033[90m", fieldStr) // Dark gray
				}
				logLine += fieldStr
			}
		}
	}
	
	// Write main log line to appropriate stream
	if entry.Level >= ErrorLevel {
		fmt.Fprintf(os.Stderr, "%s\n", logLine)
		
		// Add stack trace for errors if available and enabled
		if rco.showStack && len(entry.StackTrace) > 0 {
			stackStr := rco.formatStackTrace(entry.StackTrace)
			if rco.useColors {
				stackStr = rco.colorize("\033[90m", stackStr) // Dark gray
			}
			fmt.Fprintf(os.Stderr, "%s\n", stackStr)
		}
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", logLine)
	}
	
	return nil
}

func (rco *RichConsoleOutput) formatStackTrace(stack []CallerInfo) string {
	if len(stack) == 0 {
		return ""
	}
	
	var lines []string
	lines = append(lines, "Stack trace:")
	
	for i, frame := range stack {
		line := fmt.Sprintf("  %d. %s.%s() at %s:%d", 
			i+1, frame.Package, frame.Function, frame.File, frame.Line)
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n")
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
// internal/logger/core/logger_core.go - Enhanced core with better context management
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

// Enhanced Layer constants with auto-configuration
const (
	LayerHandler     = "handler"
	LayerService     = "service" 
	LayerRepository  = "repository"
	LayerMiddleware  = "middleware"
	LayerAuth        = "auth"
	LayerValidation  = "validation"
	LayerCache       = "cache"
	LayerDatabase    = "database"
	LayerExternal    = "external"
	LayerSecurity    = "security"
	LayerGateway     = "gateway"
	LayerQueue       = "queue"
	LayerScheduler   = "scheduler"
	LayerWebsocket   = "websocket"
	LayerEmail       = "email"
)

// Domain mapping for automatic domain inference
var LayerToDomain = map[string]string{
	LayerHandler:     "handler",
	LayerService:     "service",
	LayerRepository:  "repository",
	LayerMiddleware:  "middleware", 
	LayerAuth:        "auth",
	LayerValidation:  "validation",
	LayerCache:       "cache",
	LayerDatabase:    "database",
	LayerExternal:    "external",
	LayerSecurity:    "security",
	LayerGateway:     "gateway",
	LayerQueue:       "queue",
	LayerScheduler:   "scheduler",
	LayerWebsocket:   "websocket",
	LayerEmail:       "email",
}

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
	Domain      string                 `json:"domain,omitempty"`
}

// LogContext holds contextual information for a logging session
type LogContext struct {
	RequestID   string
	UserID      string
	SessionID   string
	TraceID     string
	Component   string
	Operation   string
	Layer       string
	Domain      string
	StartTime   time.Time
	Fields      map[string]interface{}
}

// NewLogContext creates a new logging context
func NewLogContext() *LogContext {
	return &LogContext{
		StartTime: time.Now(),
		Fields:    make(map[string]interface{}),
	}
}

// WithRequestID adds request ID to context
func (lc *LogContext) WithRequestID(requestID string) *LogContext {
	lc.RequestID = requestID
	return lc
}

// WithUserID adds user ID to context
func (lc *LogContext) WithUserID(userID string) *LogContext {
	lc.UserID = userID
	return lc
}

// WithLayer adds layer and auto-infers domain
func (lc *LogContext) WithLayer(layer string) *LogContext {
	lc.Layer = layer
	if domain, exists := LayerToDomain[layer]; exists {
		lc.Domain = domain
	}
	return lc
}

// WithOperation adds operation
func (lc *LogContext) WithOperation(operation string) *LogContext {
	lc.Operation = operation
	return lc
}

// WithField adds a field to context
func (lc *LogContext) WithField(key string, value interface{}) *LogContext {
	lc.Fields[key] = value
	return lc
}

// WithFields adds multiple fields to context
func (lc *LogContext) WithFields(fields map[string]interface{}) *LogContext {
	for k, v := range fields {
		lc.Fields[k] = v
	}
	return lc
}

// Duration returns elapsed time since context creation
func (lc *LogContext) Duration() time.Duration {
	return time.Since(lc.StartTime)
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

// Enhanced CoreLogger with context support
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

// Context-aware logging methods
func (l *CoreLogger) WithContext(ctx *LogContext) *ContextLogger {
	return &ContextLogger{
		coreLogger: l,
		context:    ctx,
	}
}

// Traditional logging methods
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
}

// Enhanced logging with automatic context
func (l *CoreLogger) InfoWithOperation(message, layer, operation string, fields ...map[string]interface{}) {
	ctx := NewLogContext().WithLayer(layer).WithOperation(operation)
	l.WithContext(ctx).Info(message, fields...)
}

func (l *CoreLogger) ErrorWithCause(message, cause, layer, operation string, fields ...map[string]interface{}) {
	ctx := NewLogContext().WithLayer(layer).WithOperation(operation).WithField("cause", cause)
	l.WithContext(ctx).Error(message, fields...)
}

func (l *CoreLogger) WarnWithCause(message, cause, layer, operation string, fields ...map[string]interface{}) {
	ctx := NewLogContext().WithLayer(layer).WithOperation(operation).WithField("cause", cause)
	l.WithContext(ctx).Warn(message, fields...)
}

// Core logging implementation
func (l *CoreLogger) log(level Level, message string, fields ...map[string]interface{}) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	
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
	
	// Auto-infer domain from layer
	if l.layer != "" {
		if domain, exists := LayerToDomain[l.layer]; exists {
			entry.Domain = domain
		}
	}
	
	if caller := getCaller(3); caller != "" {
		entry.Caller = caller
	}
	
	om := l.outputManager
	l.mu.RUnlock()
	
	if om != nil {
		om.WriteToAll(entry)
	} else {
		fmt.Fprintf(os.Stdout, "[%s] %s %s\n", 
			entry.Timestamp.Format("2006-01-02 15:04:05"), 
			entry.Level.String(), 
			entry.Message)
	}
}

// ContextLogger wraps CoreLogger with context
type ContextLogger struct {
	coreLogger *CoreLogger
	context    *LogContext
}

func (cl *ContextLogger) Debug(message string, fields ...map[string]interface{}) {
	cl.logWithContext(DebugLevel, message, fields...)
}

func (cl *ContextLogger) Info(message string, fields ...map[string]interface{}) {
	cl.logWithContext(InfoLevel, message, fields...)
}

func (cl *ContextLogger) Warn(message string, fields ...map[string]interface{}) {
	cl.logWithContext(WarnLevel, message, fields...)
}

func (cl *ContextLogger) Error(message string, fields ...map[string]interface{}) {
	cl.logWithContext(ErrorLevel, message, fields...)
}

func (cl *ContextLogger) Fatal(message string, fields ...map[string]interface{}) {
	cl.logWithContext(FatalLevel, message, fields...)
}

func (cl *ContextLogger) logWithContext(level Level, message string, fields ...map[string]interface{}) {
	cl.coreLogger.mu.RLock()
	if level < cl.coreLogger.level {
		cl.coreLogger.mu.RUnlock()
		return
	}
	
	// Merge all fields: core context + log context + provided fields
	mergedFields := make(map[string]interface{})
	
	// Add core logger context fields
	for k, v := range cl.coreLogger.contextFields {
		mergedFields[k] = v
	}
	
	// Add log context fields
	for k, v := range cl.context.Fields {
		mergedFields[k] = v
	}
	
	// Add provided fields
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			mergedFields[k] = v
		}
	}
	
	entry := &LogEntry{
		Timestamp:   time.Now(),
		Level:       level,
		Message:     message,
		Fields:      mergedFields,
		Component:   cl.getComponent(),
		Layer:       cl.getLayer(),
		Operation:   cl.getOperation(),
		Environment: cl.coreLogger.environment,
		RequestID:   cl.context.RequestID,
		UserID:      cl.context.UserID,
		SessionID:   cl.context.SessionID,
		TraceID:     cl.context.TraceID,
		Domain:      cl.getDomain(),
		Duration:    cl.context.Duration(),
	}
	
	// Extract cause from context or fields
	if cause, exists := mergedFields["cause"]; exists {
		if causeStr, ok := cause.(string); ok {
			entry.Cause = causeStr
		}
	}
	
	if caller := getCaller(3); caller != "" {
		entry.Caller = caller
	}
	
	om := cl.coreLogger.outputManager
	cl.coreLogger.mu.RUnlock()
	
	if om != nil {
		om.WriteToAll(entry)
	}
}

// Helper methods for ContextLogger
func (cl *ContextLogger) getComponent() string {
	if cl.context.Component != "" {
		return cl.context.Component
	}
	return cl.coreLogger.component
}

func (cl *ContextLogger) getLayer() string {
	if cl.context.Layer != "" {
		return cl.context.Layer
	}
	return cl.coreLogger.layer
}

func (cl *ContextLogger) getOperation() string {
	if cl.context.Operation != "" {
		return cl.context.Operation
	}
	return cl.coreLogger.operation
}

func (cl *ContextLogger) getDomain() string {
	if cl.context.Domain != "" {
		return cl.context.Domain
	}
	layer := cl.getLayer()
	if domain, exists := LayerToDomain[layer]; exists {
		return domain
	}
	return "system"
}

// Helper methods
func (l *CoreLogger) mergeFields(fields ...map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	for k, v := range l.contextFields {
		merged[k] = v
	}
	
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			merged[k] = v
		}
	}
	
	return merged
}

func getCaller(skip int) string {
	// Simple implementation - enhance with runtime.Caller if needed
	return ""
}

// RichConsoleOutput - Enhanced console output
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
	
	var parts []string
	
	// Timestamp
	timestampStr := fmt.Sprintf("[%s]", timestamp)
	if rco.useColors {
		timestampStr = rco.colorize("\033[90m", timestampStr)
	}
	parts = append(parts, timestampStr)
	
	// Level with color
	levelStr := entry.Level.String()
	if rco.useColors {
		levelStr = rco.colorizeLevel(entry.Level, levelStr)
	}
	parts = append(parts, levelStr)
	
	// Layer, Component, Operation
	if entry.Layer != "" {
		layerStr := fmt.Sprintf("[%s]", strings.ToUpper(entry.Layer))
		if rco.useColors {
			layerStr = rco.colorize("\033[94m", layerStr)
		}
		parts = append(parts, layerStr)
	}
	
	if entry.Component != "" {
		componentStr := fmt.Sprintf("<%s>", entry.Component)
		if rco.useColors {
			componentStr = rco.colorize("\033[95m", componentStr)
		}
		parts = append(parts, componentStr)
	}
	
	if entry.Operation != "" {
		operationStr := fmt.Sprintf("{%s}", entry.Operation)
		if rco.useColors {
			operationStr = rco.colorize("\033[96m", operationStr)
		}
		parts = append(parts, operationStr)
	}
	
	// Enhanced message with domain and cause for errors/warnings
	message := entry.Message
	if entry.Level >= ErrorLevel && (entry.Domain != "" || entry.Cause != "") {
		var extras []string
		if entry.Domain != "" {
			extras = append(extras, fmt.Sprintf("domain=%s", entry.Domain))
		}
		if entry.Cause != "" {
			extras = append(extras, fmt.Sprintf("cause=%s", entry.Cause))
		}
		if len(extras) > 0 {
			message = fmt.Sprintf("%s [%s]", message, strings.Join(extras, "] ["))
		}
	}
	
	parts = append(parts, message)
	logLine := strings.Join(parts, " ")
	
	// Add essential fields as JSON
	if len(entry.Fields) > 0 {
		filteredFields := make(map[string]interface{})
		for k, v := range entry.Fields {
			if entry.Level >= ErrorLevel && (k == "domain" || k == "cause") {
				continue
			}
			if k == "request_id" || k == "user_id" || k == "email" || k == "method" || 
			   k == "endpoint" || k == "status_code" || k == "duration_ms" || 
			   k == "error_code" || k == "service" {
				filteredFields[k] = v
			}
		}
		
		if len(filteredFields) > 0 {
			fieldsJSON, err := json.Marshal(filteredFields)
			if err == nil {
				fieldStr := fmt.Sprintf(" | %s", string(fieldsJSON))
				if rco.useColors {
					fieldStr = rco.colorize("\033[90m", fieldStr)
				}
				logLine += fieldStr
			}
		}
	}
	
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
		color = "\033[36m"
	case InfoLevel:
		color = "\033[32m"
	case WarnLevel:
		color = "\033[33m"
	case ErrorLevel:
		color = "\033[31m"
	case FatalLevel:
		color = "\033[35m\033[1m"
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
	
	consoleOutput := NewRichConsoleOutput(true)
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
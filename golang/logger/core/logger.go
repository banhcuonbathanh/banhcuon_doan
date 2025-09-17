// logger/core/logger_with_filter.go - Updated logger with layer filtering
package core

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Enhanced CoreLogger with layer filtering capabilities
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
	layerFilter   *LayerFilter  // New: Layer filtering system
	mu            sync.RWMutex
}

// NewLogger creates a new enhanced logger instance with layer filtering
func NewLogger() *CoreLogger {
	logger := &CoreLogger{
		level:         InfoLevel,
		contextFields: make(map[string]interface{}),
		environment:   "development",
		captureStack:  true,  // Enable stack capture for better debugging
		stackDepth:    5,     // Capture up to 5 stack frames for errors
		layerFilter:   NewLayerFilter(), // Initialize layer filter
	}
	
	// Create a default console output manager with rich formatting
	outputManager := NewDefaultOutputManager()
	logger.SetOutputManager(outputManager)
	
	return logger
}

// Layer filtering configuration methods
func (l *CoreLogger) EnableOnlyLayers(layers ...string) {
	l.layerFilter.EnableOnlyLayers(layers...)
}

func (l *CoreLogger) DisableLayers(layers ...string) {
	l.layerFilter.DisableLayers(layers...)
}

func (l *CoreLogger) EnableAllLayers() {
	l.layerFilter.EnableAllLayers()
}

func (l *CoreLogger) SetLayerLevel(layer string, level Level) {
	l.layerFilter.SetLayerLevel(layer, level)
}

func (l *CoreLogger) RemoveLayerLevel(layer string) {
	l.layerFilter.RemoveLayerLevel(layer)
}

func (l *CoreLogger) ClearLayerLevels() {
	l.layerFilter.ClearLayerLevels()
}

func (l *CoreLogger) GetLayerFilterStatus() string {
	return l.layerFilter.GetFilterStatus()
}

func (l *CoreLogger) SetLayerFilterConfig(config LayerFilterConfig) {
	l.layerFilter.SetConfig(config)
}

func (l *CoreLogger) GetLayerFilterConfig() LayerFilterConfig {
	return l.layerFilter.GetConfig()
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
	l.logWithLayer(InfoLevel, message, layer, 3, mergedFields)
}

// Enhanced ErrorWithCause method with better caller tracking
func (l *CoreLogger) ErrorWithCause(message, cause, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["cause"] = cause
	mergedFields["layer"] = layer  
	mergedFields["operation"] = operation
	
	// Add domain if not already present
	if _, exists := mergedFields["domain"]; !exists {
		mergedFields["domain"] = l.inferDomainFromLayer(layer)
	}
	
	l.logWithLayer(ErrorLevel, message, layer, 3, mergedFields)
}

// Enhanced WarnWithCause method
func (l *CoreLogger) WarnWithCause(message, cause, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["cause"] = cause
	mergedFields["layer"] = layer
	mergedFields["operation"] = operation
	
	// Add domain if not already present
	if _, exists := mergedFields["domain"]; !exists {
		mergedFields["domain"] = l.inferDomainFromLayer(layer)
	}
	
	l.logWithLayer(WarnLevel, message, layer, 3, mergedFields)
}

// New method for enhanced error logging with domain
func (l *CoreLogger) ErrorWithDomainAndCause(message, domain, cause, layer, operation string, fields ...map[string]interface{}) {
	mergedFields := l.mergeFields(fields...)
	mergedFields["domain"] = domain
	mergedFields["cause"] = cause
	mergedFields["layer"] = layer  
	mergedFields["operation"] = operation
	
	l.logWithLayer(ErrorLevel, message, layer, 3, mergedFields)
}

// New layer-specific logging methods
// New layer-specific logging methods
func (l *CoreLogger) DebugInLayer(layer, message string, fields ...map[string]interface{}) {
	l.logWithLayer(DebugLevel, message, layer, 3, l.mergeFields(fields...))
}

func (l *CoreLogger) InfoInLayer(layer, message string, fields ...map[string]interface{}) {
	l.logWithLayer(InfoLevel, message, layer, 3, l.mergeFields(fields...))
}

func (l *CoreLogger) WarnInLayer(layer, message string, fields ...map[string]interface{}) {
	l.logWithLayer(WarnLevel, message, layer, 3, l.mergeFields(fields...))
}

func (l *CoreLogger) ErrorInLayer(layer, message string, fields ...map[string]interface{}) {
	l.logWithLayer(ErrorLevel, message, layer, 3, l.mergeFields(fields...))
}

// inferDomainFromLayer infers domain from layer
func (l *CoreLogger) inferDomainFromLayer(layer string) string {
	switch layer {
	case LayerAuth:
		return "auth"
	case LayerDatabase:
		return "database"
	case LayerExternal:
		return "external"
	case LayerValidation:
		return "validation"
	case LayerSecurity:
		return "security"
	case LayerHandler:
		return "handler"
	case LayerService:
		return "service"
	case LayerRepository:
		return "repository"
	default:
		return "system"
	}
}

// Core logging implementation with enhanced caller tracking and layer filtering
func (l *CoreLogger) log(level Level, message string, callerSkip int, fields ...map[string]interface{}) {
	l.mu.RLock()
	currentLayer := l.layer
	l.mu.RUnlock()
	
	l.logWithLayer(level, message, currentLayer, callerSkip+1, fields...)
}

// logWithLayer handles logging with specific layer and applies filtering
func (l *CoreLogger) logWithLayer(level Level, message string, layer string, callerSkip int, fields ...map[string]interface{}) {
	l.mu.RLock()
	
	// Check global level first
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	
	// Check layer filtering - this is the new filtering logic
	if !l.layerFilter.ShouldLog(layer, level) {
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
		Layer:       layer,  // Use the specific layer passed in
		Operation:   l.operation,
		Environment: l.environment,
	}
	
	// Enhanced caller information
	if caller := getEnhancedCaller(callerSkip + 1); caller != nil {
		entry.Caller = caller
	}
	
	// Capture stack trace for errors and above
	if level >= ErrorLevel && l.captureStack {
		entry.StackTrace = captureStackTrace(callerSkip+1, l.stackDepth)
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
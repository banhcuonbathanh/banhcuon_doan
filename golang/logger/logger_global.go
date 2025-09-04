// internal/logger/logger_global.go - Enhanced with proper output configuration
package logger

import (
	"english-ai-full/logger/core"

	"strings"

)

// Global logger instance


func init() {
	GlobalLogger = NewDefaultLogger()
	
	// Configure outputs for the global logger
	setupDefaultOutputs(GlobalLogger)
}

// setupDefaultOutputs configures console output for the logger
func setupDefaultOutputs(logger *core.CoreLogger) {
	// Create output manager
	outputManager := NewOutputManager()
	
	// Create console formatter based on environment
	environment := getEnvironment()
	var formatter Formatter
	
	switch strings.ToLower(environment) {
	case "production", "prod":
		formatter = NewJSONFormatter()
	case "development", "dev":
		formatter = NewPrettyFormatter(true) // With colors
	default:
		formatter = NewTextFormatter()
	}
	
	// Create console output
	consoleOutput := NewConsoleOutput(formatter, true)
	
	// Add console output to manager
	outputManager.AddOutput("console", consoleOutput)
	
	// Set the output manager on the logger
	// Note: You'll need to add this method to CoreLogger
	setOutputManager(logger, outputManager)
}

// Helper function to set output manager (you'll need to add this to CoreLogger)
func setOutputManager(logger *core.CoreLogger, outputManager *OutputManager) {
	// This is a workaround - you should add a SetOutputManager method to CoreLogger
	// For now, we'll use reflection or modify the CoreLogger struct
	// Add this method to your CoreLogger struct:
	// func (l *CoreLogger) SetOutputManager(om OutputManager) {
	//     l.outputManager = om
	// }
}




// Global convenience functions for basic logging
func Debug(message string, fields ...map[string]interface{}) {
	GlobalLogger.Debug(message, fields...)
}

func Info(message string, fields ...map[string]interface{}) {
	GlobalLogger.Info(message, fields...)
}

func Warn(message string, fields ...map[string]interface{}) {
	GlobalLogger.Warn(message, fields...)
}

func Error(message string, fields ...map[string]interface{}) {
	GlobalLogger.Error(message, fields...)
}

func Fatal(message string, fields ...map[string]interface{}) {
	GlobalLogger.Fatal(message, fields...)
}

// Enhanced global convenience functions
func ErrorWithCause(message string, cause string, layer string, operation string, fields ...map[string]interface{}) {
	GlobalLogger.ErrorWithCause(message, cause, layer, operation, fields...)
}

func WarnWithCause(message string, cause string, layer string, operation string, fields ...map[string]interface{}) {
	GlobalLogger.WarnWithCause(message, cause, layer, operation, fields...)
}

func InfoWithOperation(message string, layer string, operation string, fields ...map[string]interface{}) {
	GlobalLogger.InfoWithOperation(message, layer, operation, fields...)
}


func SetLevel(level core.Level) {
	GlobalLogger.SetLevel(level)
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
	GlobalLogger.AddContextField(key, value)
}

func RemoveGlobalField(key string) {
	GlobalLogger.RemoveContextField(key)
}




// Legacy compatibility wrapper
type Logger struct {
	*core.CoreLogger
}

// NewLogger creates a new logger instance (maintains compatibility)
func NewLogger() *Logger {
	logger := &Logger{
		CoreLogger: NewDefaultLogger(),
	}
	setupDefaultOutputs(logger.CoreLogger)
	return logger
}

// Legacy compatibility methods - these delegate to the new enhanced methods
func (l *Logger) Warning(message string, context ...map[string]interface{}) {
	l.Warn(message, context...)
}

func (l *Logger) SetOutputFormat(format string) {
	// This would be handled by output configuration in the new system
	// For now, we'll maintain compatibility but log a deprecation notice
	l.Debug("SetOutputFormat is deprecated, use output configuration instead", map[string]interface{}{
		"format": format,
		"notice": "deprecated_method",
	})
}

func (l *Logger) SetDebugLogging(enable bool) {
	if enable {
		l.SetLevel(core.DebugLevel)
	} else {
		l.SetLevel(core.InfoLevel)
	}
}

func (l *Logger) SetMinLevel(level int) {
	// Convert old integer levels to new Level type
	switch level {
	case 0: // DebugLevel
		l.SetLevel(core.DebugLevel)
	case 1: // InfoLevel
		l.SetLevel(core.InfoLevel)
	case 2: // WarningLevel
		l.SetLevel(core.WarnLevel)
	case 3: // ErrorLevel
		l.SetLevel(core.ErrorLevel)
	case 4: // FatalLevel
		l.SetLevel(core.FatalLevel)
	default:
		l.SetLevel(core.InfoLevel)
	}
}
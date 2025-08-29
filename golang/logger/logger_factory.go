// internal/logger/factory.go - Factory constructors and global convenience functions
package logger

import (
	"english-ai-full/logger/core"

)

// Global logger instances

var GlobalSpecializedLogger *SpecializedLogger

func init() {
	GlobalLogger = NewDefaultLogger()
	GlobalSpecializedLogger = NewSpecializedLogger(GlobalLogger)
}


// Factory functions for specialized loggers
func NewDefaultSpecializedLogger() *SpecializedLogger {
	coreLogger := NewDefaultLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedComponentLogger(component string) *SpecializedLogger {
	coreLogger := NewComponentLogger(component)
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedLayerLogger(layer string) *SpecializedLogger {
	coreLogger := NewLayerLogger(layer)
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedHandlerLogger() *SpecializedLogger {
	coreLogger := NewHandlerLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedServiceLogger() *SpecializedLogger {
	coreLogger := NewServiceLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedRepositoryLogger() *SpecializedLogger {
	coreLogger := NewRepositoryLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedMiddlewareLogger() *SpecializedLogger {
	coreLogger := NewMiddlewareLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedAuthLogger() *SpecializedLogger {
	coreLogger := NewAuthLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedValidationLogger() *SpecializedLogger {
	coreLogger := NewValidationLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedCacheLogger() *SpecializedLogger {
	coreLogger := NewCacheLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedDatabaseLogger() *SpecializedLogger {
	coreLogger := NewDatabaseLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedExternalLogger() *SpecializedLogger {
	coreLogger := NewExternalLogger()
	return NewSpecializedLogger(coreLogger)
}

func NewSpecializedSecurityLogger() *SpecializedLogger {
	coreLogger := NewSecurityLogger()
	return NewSpecializedLogger(coreLogger)
}



// Legacy compatibility - Logger type for backward compatibility
type Logger struct {
	*core.CoreLogger
}

// NewCompatibilityLogger creates a new logger instance (maintains compatibility with old API)
func NewCompatibilityLogger() *Logger {
	return &Logger{
		CoreLogger: NewDefaultLogger(),
	}
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
		l.CoreLogger.SetLevel(core.DebugLevel)
	} else {
		l.CoreLogger.SetLevel(core.InfoLevel)
	}
}

func (l *Logger) SetMinLevel(level int) {
	// Convert old integer levels to new Level type
	switch level {
	case 0: // DebugLevel
		l.CoreLogger.SetLevel(core.DebugLevel)
	case 1: // InfoLevel
		l.CoreLogger.SetLevel(core.InfoLevel)
	case 2: // WarningLevel
		l.CoreLogger.SetLevel(core.WarnLevel)
	case 3: // ErrorLevel
		l.CoreLogger.SetLevel(core.ErrorLevel)
	case 4: // FatalLevel
		l.CoreLogger.SetLevel(core.FatalLevel)
	default:
		l.CoreLogger.SetLevel(core.InfoLevel)
	}
}

// Additional utility functions
func WithContext(fields map[string]interface{}) *SpecializedLogger {
	logger := NewDefaultSpecializedLogger()
	if fields != nil {
		for k, v := range fields {
			logger.AddContextField(k, v)
		}
	}
	return logger
}

func WithConfig(component, layer, operation, environment string) *SpecializedLogger {
	logger := NewDefaultSpecializedLogger()
	if component != "" {
		logger.SetComponent(component)
	}
	if layer != "" {
		logger.SetLayer(layer)
	}
	if operation != "" {
		logger.SetOperation(operation)
	}
	if environment != "" {
		logger.SetEnvironment(environment)
	}
	return logger
}
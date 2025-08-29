// internal/logger/logger_factory.go - Factory constructors and global convenience functions
package logger



// Global logger instances
var GlobalSpecializedLogger *SpecializedLogger

func init() {
	GlobalLogger := NewDefaultLogger()
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

// NewCompatibilityLogger creates a new logger instance (maintains compatibility with old API)
func NewCompatibilityLogger() *Logger {
	return &Logger{
		CoreLogger: NewDefaultLogger(),
	}
}

// Additional utility functions
func WithContext(fields map[string]interface{}) *SpecializedLogger {
	logger := NewDefaultSpecializedLogger()

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
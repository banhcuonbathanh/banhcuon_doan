// internal/logger/logger_global.go - Factory constructors and global convenience functions
package logger

import (
	"english-ai-full/logger/core"
	"os"
	"strings"
	"time"
)

// Global logger instance
var GlobalLogger *core.CoreLogger

func init() {
	GlobalLogger = NewDefaultLogger()
}

// Factory functions for creating specialized loggers
func NewDefaultLogger() *core.CoreLogger {
	logger := core.NewLogger()
	
	// Set environment-based defaults
	environment := getEnvironment()
	logger.SetEnvironment(environment)
	logger.SetLevel(getMinLogLevel(environment))
	
	// Add global context fields
	logger.AddContextField("environment", environment)
	logger.AddContextField("service", "shopeasy-api")
	
	return logger
}

func NewComponentLogger(component string) *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent(component)
	return logger
}

func NewLayerLogger(layer string) *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetLayer(layer)
	return logger
}

func NewHandlerLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("handler")
	logger.SetLayer(core.LayerHandler)
	return logger
}

func NewServiceLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("service")
	logger.SetLayer(core.LayerService)
	return logger
}

func NewRepositoryLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("repository")
	logger.SetLayer(core.LayerRepository)
	return logger
}

func NewMiddlewareLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("middleware")
	logger.SetLayer(core.LayerMiddleware)
	return logger
}

func NewAuthLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("auth")
	logger.SetLayer(core.LayerAuth)
	return logger
}

func NewValidationLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("validation")
	logger.SetLayer(core.LayerValidation)
	return logger
}

func NewCacheLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("cache")
	logger.SetLayer(core.LayerCache)
	return logger
}

func NewDatabaseLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("database")
	logger.SetLayer(core.LayerDatabase)
	return logger
}

func NewExternalLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("external")
	logger.SetLayer(core.LayerExternal)
	return logger
}

func NewSecurityLogger() *core.CoreLogger {
	logger := NewDefaultLogger()
	logger.SetComponent("security")
	logger.SetLayer(core.LayerSecurity)
	return logger
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

// Global convenience functions for specialized logging
func LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {
	GlobalLogger.LogAuthAttempt(email, success, reason, additionalContext...)
}

func LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {
	GlobalLogger.LogAPIRequest(method, path, statusCode, duration, context)
}

func LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {
	GlobalLogger.LogServiceCall(service, method, success, err, context)
}

func LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {
	GlobalLogger.LogDBOperation(operation, table, success, err, context)
}

func LogValidationError(field, message string, value interface{}) {
	GlobalLogger.LogValidationError(field, message, value)
}

func LogUserActivity(userID, email, action, resource string, context map[string]interface{}) {
	GlobalLogger.LogUserActivity(userID, email, action, resource, context)
}

func LogSecurityEvent(eventType, description, severity string, context map[string]interface{}) {
	GlobalLogger.LogSecurityEvent(eventType, description, severity, context)
}

func LogMetric(metricName string, value interface{}, unit string, context map[string]interface{}) {
	GlobalLogger.LogMetric(metricName, value, unit, context)
}

func LogPerformance(operation string, duration time.Duration, context map[string]interface{}) {
	GlobalLogger.LogPerformance(operation, duration, context)
}

// Configuration functions for global logger
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

// Helper functions
func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func getMinLogLevel(environment string) core.Level {
	switch strings.ToLower(environment) {
	case "production", "prod":
		return core.InfoLevel
	case "staging", "stage":
		return core.InfoLevel
	case "testing", "test":
		return core.DebugLevel
	default: // development
		return core.DebugLevel
	}
}

// Legacy compatibility wrapper
type Logger struct {
	*core.CoreLogger
}

// NewLogger creates a new logger instance (maintains compatibility)
func NewLogger() *Logger {
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
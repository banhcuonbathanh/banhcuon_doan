// internal/logger/factory.go - Factory constructors and global convenience functions
package logger

import (

	"time"
)

// Global logger instance - using the specialized logger wrapper



// Factory functions for creating specialized loggers


// NewDefaultSpecializedLogger creates a default specialized logger with all features
func NewDefaultSpecializedLogger() *SpecializedLogger {
	coreLogger := NewDefaultLogger()
	return NewSpecializedLogger(coreLogger)
}



// Specialized factory functions that return SpecializedLogger
func NewSpecializedComponentLogger(component string) *SpecializedLogger {
	logger := NewComponentLogger(component)
	return NewSpecializedLogger(logger)
}

func NewSpecializedHandlerLogger() *SpecializedLogger {
	logger := NewHandlerLogger()
	return NewSpecializedLogger(logger)
}

func NewSpecializedServiceLogger() *SpecializedLogger {
	logger := NewServiceLogger()
	return NewSpecializedLogger(logger)
}

func NewSpecializedRepositoryLogger() *SpecializedLogger {
	logger := NewRepositoryLogger()
	return NewSpecializedLogger(logger)
}

func NewSpecializedAuthLogger() *SpecializedLogger {
	logger := NewAuthLogger()
	return NewSpecializedLogger(logger)
}

func NewSpecializedDatabaseLogger() *SpecializedLogger {
	logger := NewDatabaseLogger()
	return NewSpecializedLogger(logger)
}









func LogPasswordReset(email string, success bool, reason string) {
	GlobalLogger.LogPasswordReset(email, success, reason)
}

func LogSessionAction(action string, sessionID string, userID string, success bool) {
	GlobalLogger.LogSessionAction(action, sessionID, userID, success)
}

func LogDatabaseOperation(operation string, table string, duration time.Duration, success bool, rowsAffected int64) {
	GlobalLogger.LogDatabaseOperation(operation, table, duration, success, rowsAffected)
}

func LogAPICall(endpoint string, method string, statusCode int, duration time.Duration) {
	GlobalLogger.LogAPICall(endpoint, method, statusCode, duration)
}

func LogCacheOperation(operation string, key string, hit bool, duration time.Duration) {
	GlobalLogger.LogCacheOperation(operation, key, hit, duration)
}






// logger/global.go - Global logger instance and convenience functions
package logger

import "time"

// Global logger instance
var GlobalLogger = NewLogger()

// Convenience functions for global logger usage
func Debug(message string, context ...map[string]interface{}) {
	GlobalLogger.Debug(message, context...)
}

func Info(message string, context ...map[string]interface{}) {
	GlobalLogger.Info(message, context...)
}

func Warning(message string, context ...map[string]interface{}) {
	GlobalLogger.Warning(message, context...)
}

func Error(message string, context ...map[string]interface{}) {
	GlobalLogger.Error(message, context...)
}

func Fatal(message string, context ...map[string]interface{}) {
	GlobalLogger.Fatal(message, context...)
}

// Enhanced global convenience functions
func ErrorWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {
	GlobalLogger.ErrorWithCause(message, cause, layer, operation, context...)
}

func WarningWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {
	GlobalLogger.WarningWithCause(message, cause, layer, operation, context...)
}

func InfoWithOperation(message string, layer string, operation string, context ...map[string]interface{}) {
	GlobalLogger.InfoWithOperation(message, layer, operation, context...)
}

// Global convenience functions for specialized logging
func LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {
	GlobalLogger.LogAuthAttempt(email, success, reason, additionalContext...)
}

func LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {
	GlobalLogger.LogDBOperation(operation, table, success, err, context)
}

func LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {
	GlobalLogger.LogServiceCall(service, method, success, err, context)
}

func LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {
	GlobalLogger.LogAPIRequest(method, path, statusCode, duration, context)
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

// Configuration functions
func SetOutputFormat(format string) {
	GlobalLogger.SetOutputFormat(format)
}

func SetDebugLogging(enable bool) {
	GlobalLogger.SetDebugLogging(enable)
}

func SetMinLevel(level int) {
	GlobalLogger.SetMinLevel(level)
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
	GlobalLogger.AddGlobalField(key, value)
}

func RemoveGlobalField(key string) {
	GlobalLogger.RemoveGlobalField(key)
}
// logger/utils.go - Helper utilities and error categorization
package logger

import (
	"fmt"
	"os"
	"strings"
)

// Helper functions (enhanced with error categorization)
func categorizeError(err error) string {
	if err == nil {
		return ""
	}
	
	errMsg := strings.ToLower(err.Error())
	
	// Network related errors
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "network") {
		return "network_error"
	}
	
	// Timeout errors
	if strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "deadline exceeded") {
		return "timeout_error"
	}
	
	// Authentication errors
	if strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "authentication") ||
		strings.Contains(errMsg, "invalid token") {
		return "auth_error"
	}
	
	// Validation errors
	if strings.Contains(errMsg, "validation") ||
		strings.Contains(errMsg, "invalid input") ||
		strings.Contains(errMsg, "bad request") {
		return "validation_error"
	}
	
	// Permission errors
	if strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "access denied") {
		return "permission_error"
	}
	
	// Resource errors
	if strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") {
		return "resource_not_found"
	}
	
	// Conflict errors
	if strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "conflict") {
		return "resource_conflict"
	}
	
	// External service errors
	if strings.Contains(errMsg, "service unavailable") ||
		strings.Contains(errMsg, "bad gateway") {
		return "external_service_error"
	}
	
	return "unknown_error"
}

func categorizeDBError(err error) string {
	if err == nil {
		return ""
	}
	
	errMsg := strings.ToLower(err.Error())
	
	// Connection errors
	if strings.Contains(errMsg, "connection") {
		return "db_connection_error"
	}
	
	// Constraint violations
	if strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "primary key") {
		return "db_constraint_violation"
	}
	
	// Syntax errors
	if strings.Contains(errMsg, "syntax error") ||
		strings.Contains(errMsg, "invalid sql") {
		return "db_syntax_error"
	}
	
	// Data errors
	if strings.Contains(errMsg, "data too long") ||
		strings.Contains(errMsg, "out of range") {
		return "db_data_error"
	}
	
	// Transaction errors
	if strings.Contains(errMsg, "deadlock") ||
		strings.Contains(errMsg, "lock timeout") {
		return "db_transaction_error"
	}
	
	// Permission errors
	if strings.Contains(errMsg, "access denied") ||
		strings.Contains(errMsg, "permission denied") {
		return "db_permission_error"
	}
	
	return "db_unknown_error"
}

func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***invalid_email***"
	}
	
	username := parts[0]
	domain := parts[1]
	
	var maskedUsername string
	if len(username) <= 2 {
		maskedUsername = "**"
	} else if len(username) <= 4 {
		maskedUsername = string(username[0]) + "**" + string(username[len(username)-1])
	} else {
		maskedUsername = string(username[0]) + "***" + string(username[len(username)-1])
	}
	
	return maskedUsername + "@" + domain
}

func getValueLength(value interface{}) int {
	if value == nil {
		return 0
	}
	
	switch v := value.(type) {
	case string:
		return len(v)
	case []byte:
		return len(v)
	default:
		return len(fmt.Sprintf("%v", v))
	}
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errMsg := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"service unavailable",
		"deadline exceeded",
		"context deadline exceeded",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}
	
	return false
}

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

func getOutputFormat(environment string) string {
	format := os.Getenv("LOG_FORMAT")
	if format != "" {
		return format
	}
	
	// Default formats based on environment
	switch strings.ToLower(environment) {
	case "production", "prod":
		return FormatJSON
	case "staging", "stage":
		return FormatJSON
	default: // development, testing
		return FormatPretty
	}
}

func getMinLogLevel(environment string) int {
	switch strings.ToLower(environment) {
	case "production", "prod":
		return InfoLevel
	case "staging", "stage":
		return InfoLevel
	case "testing", "test":
		return DebugLevel
	default: // development
		return DebugLevel
	}
}
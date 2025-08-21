# Go Logger Package Usage Guide

## Overview
This is a comprehensive logging package for Go applications with structured logging, multiple output formats, and specialized logging methods for different domains (auth, API, database, etc.).

## Package Structure
```
logger/
├── logger_types.go      - Core types and constants
├── logger_core.go       - Main logger implementation
├── logger_factory.go    - Component-specific logger constructors
├── logger_specialized.go - Domain-specific logging methods
├── logger_formatters.go - Output formatting implementations
├── logger_global.go     - Global logger instance and convenience functions
└── logger_utils.go      - Helper utilities and error categorization
```

## Constants and Types

### Log Levels
```go
const (
    DebugLevel = iota
    InfoLevel
    WarningLevel
    ErrorLevel
    FatalLevel
)
```

### Output Formats
```go
const (
    FormatJSON   = "json"
    FormatText   = "text"
    FormatPretty = "pretty"
)
```

### Layer Constants
```go
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
```

## Core Logger Functions

### Constructor
```go
func NewLogger() *Logger {}
```

### Configuration Methods
```go
func (l *Logger) SetOutputFormat(format string) {}
func (l *Logger) SetDebugLogging(enable bool) {}
func (l *Logger) SetMinLevel(level int) {}
func (l *Logger) SetComponent(component string) {}
func (l *Logger) SetLayer(layer string) {}
func (l *Logger) SetOperation(operation string) {}
func (l *Logger) AddGlobalField(key string, value interface{}) {}
func (l *Logger) RemoveGlobalField(key string) {}
```

### Basic Logging Methods
```go
func (l *Logger) Debug(message string, context ...map[string]interface{}) {}
func (l *Logger) Info(message string, context ...map[string]interface{}) {}
func (l *Logger) Warning(message string, context ...map[string]interface{}) {}
func (l *Logger) Error(message string, context ...map[string]interface{}) {}
func (l *Logger) Fatal(message string, context ...map[string]interface{}) {}
```

### Enhanced Logging Methods
```go
func (l *Logger) ErrorWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {}
func (l *Logger) WarningWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {}
func (l *Logger) InfoWithOperation(message string, layer string, operation string, context ...map[string]interface{}) {}
```

## Factory Functions

### Component-Specific Loggers
```go
func NewComponentLogger(component string) *Logger {}
func NewLayerLogger(layer string) *Logger {}
func NewHandlerLogger() *Logger {}
func NewServiceLogger() *Logger {}
func NewRepositoryLogger() *Logger {}
func NewMiddlewareLogger() *Logger {}
func NewAuthLogger() *Logger {}
func NewValidationLogger() *Logger {}
func NewCacheLogger() *Logger {}
func NewDatabaseLogger() *Logger {}
func NewExternalLogger() *Logger {}
func NewSecurityLogger() *Logger {}
```

## Specialized Logging Methods

### Authentication Logging
```go
func (l *Logger) LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {}
```

### API Request Logging
```go
func (l *Logger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {}
```

### Service Call Logging
```go
func (l *Logger) LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {}
```

### Database Operation Logging
```go
func (l *Logger) LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {}
```

### Validation Logging
```go
func (l *Logger) LogValidationError(field, message string, value interface{}) {}
```

### User Activity Logging
```go
func (l *Logger) LogUserActivity(userID, email, action string, resource string, context map[string]interface{}) {}
```

### Security Event Logging
```go
func (l *Logger) LogSecurityEvent(eventType, description string, severity string, context map[string]interface{}) {}
```

### Metrics and Performance Logging
```go
func (l *Logger) LogMetric(metricName string, value interface{}, unit string, context map[string]interface{}) {}
func (l *Logger) LogPerformance(operation string, duration time.Duration, context map[string]interface{}) {}
```

## Global Convenience Functions

### Basic Global Logging
```go
func Debug(message string, context ...map[string]interface{}) {}
func Info(message string, context ...map[string]interface{}) {}
func Warning(message string, context ...map[string]interface{}) {}
func Error(message string, context ...map[string]interface{}) {}
func Fatal(message string, context ...map[string]interface{}) {}
```

### Enhanced Global Logging
```go
func ErrorWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {}
func WarningWithCause(message string, cause string, layer string, operation string, context ...map[string]interface{}) {}
func InfoWithOperation(message string, layer string, operation string, context ...map[string]interface{}) {}
```

### Global Specialized Logging
```go
func LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{}) {}
func LogDBOperation(operation, table string, success bool, err error, context map[string]interface{}) {}
func LogServiceCall(service, method string, success bool, err error, context map[string]interface{}) {}
func LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{}) {}
func LogValidationError(field, message string, value interface{}) {}
func LogUserActivity(userID, email, action, resource string, context map[string]interface{}) {}
func LogSecurityEvent(eventType, description, severity string, context map[string]interface{}) {}
func LogMetric(metricName string, value interface{}, unit string, context map[string]interface{}) {}
func LogPerformance(operation string, duration time.Duration, context map[string]interface{}) {}
```

### Global Configuration Functions
```go
func SetOutputFormat(format string) {}
func SetDebugLogging(enable bool) {}
func SetMinLevel(level int) {}
func SetComponent(component string) {}
func SetLayer(layer string) {}
func SetOperation(operation string) {}
func AddGlobalField(key string, value interface{}) {}
func RemoveGlobalField(key string) {}
```

## Utility Functions

### Error Categorization
```go
func categorizeError(err error) string {}
func categorizeDBError(err error) string {}
```

### Helper Functions
```go
func maskEmail(email string) string {}
func getValueLength(value interface{}) int {}
func isRetryableError(err error) bool {}
func getEnvironment() string {}
func getOutputFormat(environment string) string {}
func getMinLogLevel(environment string) int {}
```

### Formatting Functions
```go
func (l *Logger) formatPretty(entry LogEntry) string {}
func (l *Logger) formatText(entry LogEntry) string {}
func formatLevel(level string) string {}
```

## Usage Examples

### Basic Usage
```go
import "your-project/logger"

// Using global logger
logger.Info("Server started", map[string]interface{}{
    "port": 8080,
    "environment": "development",
})

// Using component-specific logger
authLogger := logger.NewAuthLogger()
authLogger.LogAuthAttempt("user@example.com", false, "invalid_password")
```

### Advanced Usage
```go
// Custom logger with specific configuration
customLogger := logger.NewLogger()
customLogger.SetOutputFormat(logger.FormatJSON)
customLogger.SetComponent("payment-service")
customLogger.SetLayer(logger.LayerService)

// Enhanced error logging with cause
logger.ErrorWithCause(
    "Payment processing failed",
    "payment_gateway_timeout",
    logger.LayerService,
    "process_payment",
    map[string]interface{}{
        "payment_id": "pay_123",
        "amount": 99.99,
    },
)
```

## Environment Variables

- `APP_ENV` or `ENVIRONMENT`: Sets the application environment (development, staging, production)
- `LOG_FORMAT`: Override the default log format (json, text, pretty)

## Features

1. **Structured Logging**: All logs include contextual information in a structured format
2. **Multiple Output Formats**: JSON for production, pretty format for development
3. **Layer-based Organization**: Organize logs by application layers (handler, service, repository, etc.)
4. **Security-aware**: Automatic email masking and sensitive field handling
5. **Error Categorization**: Automatic categorization of errors for better monitoring
6. **Performance Tracking**: Built-in performance categorization and cause analysis
7. **Thread-safe**: All operations are thread-safe with proper mutex usage
8. **Environment-aware**: Different default configurations based on environment
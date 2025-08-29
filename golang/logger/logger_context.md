# Go Logger Package Context

## Overview
This is a comprehensive, production-ready logging package for Go applications with structured logging, multiple output formats, and specialized domain logging capabilities.

## Architecture

### Core Components

1. **Core Logger (`logger_core.go`)**
   - Main logging engine with thread-safe operations
   - Supports hierarchical log levels (Debug, Info, Warn, Error, Fatal)
   - Structured logging with rich metadata fields
   - Context-aware logging with request tracing support

2. **Specialized Logger (`logger_specialized.go`)**
   - Domain-specific logging methods for common use cases
   - Authentication, API calls, database operations, security events
   - Performance monitoring and metrics collection
   - Business event tracking

3. **Output System (`logget_output.go`)**
   - Multiple output destinations (console, file, multi-output)
   - Buffered and filtered outputs
   - Thread-safe output management

4. **Formatters (`logger_formatters.go`)**
   - JSON formatting for structured logs
   - Text formatting for human readability
   - Pretty formatting with colors and emojis
   - Runtime caller information

5. **Factory & Global Functions (`logger_global.go`, `logger_factory.go`)**
   - Pre-configured logger instances for different layers
   - Global convenience functions
   - Environment-based configuration

## Key Features

### Structured Logging
```go
logger.Info("User login successful", map[string]interface{}{
    "user_id": "12345",
    "email": "user@example.com",
    "ip": "192.168.1.1",
})
```

### Layer-based Architecture
- Handler, Service, Repository, Middleware layers
- Auth, Validation, Cache, Database, External layers
- Security and Performance monitoring layers

### Specialized Methods
- `LogAuthAttempt()` - Authentication events
- `LogAPIRequest()` - HTTP request/response logging  
- `LogDBOperation()` - Database operation tracking
- `LogSecurityEvent()` - Security incident logging
- `LogPerformance()` - Performance metrics

### Context Fields
- Request ID, User ID, Session ID for tracing
- Component, Layer, Operation for categorization
- Environment, Caller information
- Duration tracking for performance

### Security Features
- Email masking for PII protection
- Security event classification
- Audit trail capabilities

### Output Options
- Console output with optional colors
- File output with rotation support
- JSON format for log aggregation
- Pretty format for development

## Usage Patterns

### Basic Logging
```go
// Global functions
logger.Info("Application started")
logger.Error("Database connection failed", map[string]interface{}{
    "error": err.Error(),
    "host": "localhost:5432",
})
```

### Component-specific Logging
```go
authLogger := logger.NewAuthLogger()
authLogger.LogAuthAttempt("user@example.com", false, "invalid_password")
```

### Specialized Logging
```go
specialized := logger.NewSpecializedLogger(coreLogger)
specialized.LogAPICall("/api/users", "GET", 200, 150*time.Millisecond)
specialized.LogPerformance("user_search", 25*time.Millisecond, true, nil)
```

## Configuration

### Environment-based Setup
- Production: JSON format, Info level minimum
- Development: Pretty format, Debug level
- Staging: JSON format, Info level

### Layer Constants
Pre-defined constants for consistent layer identification:
- `LayerHandler`, `LayerService`, `LayerRepository`
- `LayerAuth`, `LayerValidation`, `LayerCache`
- `LayerDatabase`, `LayerExternal`, `LayerSecurity`

## Thread Safety
All logger components are thread-safe with appropriate mutex usage for concurrent access in web applications.

## Legacy Compatibility
Maintains backward compatibility with older logging interfaces while providing enhanced functionality through the new structured approach.

This logging package provides enterprise-grade logging capabilities with excellent observability, security, and maintainability features suitable for production Go applications.

# Complete Function List for Go Logger Package

## Core Logger Functions (`logger_core.go`)

### Constructor
- `NewLogger() *CoreLogger`

### Configuration Methods
- `SetLevel(level Level)`
- `SetComponent(component string)`
- `SetLayer(layer string)`
- `SetOperation(operation string)`
- `SetEnvironment(env string)`
- `AddContextField(key string, value interface{})`
- `RemoveContextField(key string)`

### Basic Logging Methods
- `Debug(message string, fields ...map[string]interface{})`
- `Info(message string, fields ...map[string]interface{})`
- `Warn(message string, fields ...map[string]interface{})`
- `Error(message string, fields ...map[string]interface{})`
- `Fatal(message string, fields ...map[string]interface{})`

### Enhanced Logging Methods
- `ErrorWithCause(message, cause, layer, operation string, fields ...map[string]interface{})`
- `WarnWithCause(message, cause, layer, operation string, fields ...map[string]interface{})`
- `InfoWithOperation(message, layer, operation string, fields ...map[string]interface{})`

### Specialized Core Methods
- `LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{})`
- `LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{})`
- `LogServiceCall(service, method string, success bool, err error, context map[string]interface{})`
- `LogDBOperation(operation, table string, success bool, err error, context map[string]interface{})`
- `LogValidationError(field, message string, value interface{})`
- `LogUserActivity(userID, email, action, resource string, context map[string]interface{})`
- `LogSecurityEvent(eventType, description, severity string, context map[string]interface{})`
- `LogMetric(metricName string, value interface{}, unit string, context map[string]interface{})`
- `LogPerformance(operation string, duration time.Duration, context map[string]interface{})`

### Internal Methods
- `WriteToOutput(outputName string, entry *LogEntry) error`
- `log(level Level, message string, fields ...map[string]interface{})`
- `mergeFields(fields ...map[string]interface{}) map[string]interface{}`

### Helper Functions
- `maskEmail(email string) string`
- `getCaller(skip int) string`

## Specialized Logger Functions (`logger_specialized.go`)

### Constructor
- `NewSpecializedLogger(coreLogger *core.CoreLogger) *SpecializedLogger`

### Authentication Logging
- `LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{})`
- `LogPasswordReset(email string, success bool, reason string)`
- `LogSessionAction(action string, sessionID string, userID string, success bool)`

### Database & API Logging
- `LogDatabaseOperation(operation string, table string, duration time.Duration, success bool, rowsAffected int64)`
- `LogAPICall(endpoint string, method string, statusCode int, duration time.Duration)`
- `LogCacheOperation(operation string, key string, hit bool, duration time.Duration)`

### Validation & User Activity
- `LogValidationError(field string, value interface{}, rule string, message string)`
- `LogUserActivity(userID string, action string, resource string, metadata map[string]interface{})`

### Metrics & Performance
- `LogMetric(name string, value float64, unit string, tags map[string]string)`
- `LogPerformance(operation string, duration time.Duration, success bool, metadata map[string]interface{})`

### Request Tracking
- `LogRequestStart(requestID string, method string, endpoint string, userID string)`
- `LogRequestEnd(requestID string, statusCode int, duration time.Duration, responseSize int64)`

### Business & Security Events
- `LogBusinessEvent(eventType string, entityID string, entityType string, action string, metadata map[string]interface{})`
- `LogSecurityEvent(eventType string, severity string, userID string, ip string, details map[string]interface{})`
- `LogHealthCheck(service string, status string, duration time.Duration, details map[string]interface{})`

### Helper Functions
- `maskEmail(email string) string`

## Factory Functions (`logger_factory.go`, `logger_global.go`)

### Default Constructors
- `NewDefaultLogger() *core.CoreLogger`
- `NewDefaultSpecializedLogger() *SpecializedLogger`

### Component-Specific Constructors
- `NewComponentLogger(component string) *core.CoreLogger`
- `NewLayerLogger(layer string) *core.CoreLogger`
- `NewHandlerLogger() *core.CoreLogger`
- `NewServiceLogger() *core.CoreLogger`
- `NewRepositoryLogger() *core.CoreLogger`
- `NewMiddlewareLogger() *core.CoreLogger`
- `NewAuthLogger() *core.CoreLogger`
- `NewValidationLogger() *core.CoreLogger`
- `NewCacheLogger() *core.CoreLogger`
- `NewDatabaseLogger() *core.CoreLogger`
- `NewExternalLogger() *core.CoreLogger`
- `NewSecurityLogger() *core.CoreLogger`

### Specialized Factory Functions
- `NewSpecializedComponentLogger(component string) *SpecializedLogger`
- `NewSpecializedLayerLogger(layer string) *SpecializedLogger`
- `NewSpecializedHandlerLogger() *SpecializedLogger`
- `NewSpecializedServiceLogger() *SpecializedLogger`
- `NewSpecializedRepositoryLogger() *SpecializedLogger`
- `NewSpecializedMiddlewareLogger() *SpecializedLogger`
- `NewSpecializedAuthLogger() *SpecializedLogger`
- `NewSpecializedValidationLogger() *SpecializedLogger`
- `NewSpecializedCacheLogger() *SpecializedLogger`
- `NewSpecializedDatabaseLogger() *SpecializedLogger`
- `NewSpecializedExternalLogger() *SpecializedLogger`
- `NewSpecializedSecurityLogger() *SpecializedLogger`

### Utility Constructors
- `NewCompatibilityLogger() *Logger`
- `WithContext(fields map[string]interface{}) *SpecializedLogger`
- `WithConfig(component, layer, operation, environment string) *SpecializedLogger`

### Global Convenience Functions
- `Debug(message string, fields ...map[string]interface{})`
- `Info(message string, fields ...map[string]interface{})`
- `Warn(message string, fields ...map[string]interface{})`
- `Error(message string, fields ...map[string]interface{})`
- `Fatal(message string, fields ...map[string]interface{})`
- `ErrorWithCause(message string, cause string, layer string, operation string, fields ...map[string]interface{})`
- `WarnWithCause(message string, cause string, layer string, operation string, fields ...map[string]interface{})`
- `InfoWithOperation(message string, layer string, operation string, fields ...map[string]interface{})`

### Global Specialized Functions
- `LogAuthAttempt(email string, success bool, reason string, additionalContext ...map[string]interface{})`
- `LogAPIRequest(method, path string, statusCode int, duration time.Duration, context map[string]interface{})`
- `LogServiceCall(service, method string, success bool, err error, context map[string]interface{})`
- `LogDBOperation(operation, table string, success bool, err error, context map[string]interface{})`
- `LogValidationError(field, message string, value interface{})`
- `LogUserActivity(userID, email, action, resource string, context map[string]interface{})`
- `LogSecurityEvent(eventType, description, severity string, context map[string]interface{})`
- `LogMetric(metricName string, value interface{}, unit string, context map[string]interface{})`
- `LogPerformance(operation string, duration time.Duration, context map[string]interface{})`

### Global Configuration Functions
- `SetLevel(level core.Level)`
- `SetComponent(component string)`
- `SetLayer(layer string)`
- `SetOperation(operation string)`
- `AddGlobalField(key string, value interface{})`
- `RemoveGlobalField(key string)`

### Helper Functions
- `getEnvironment() string`
- `getMinLogLevel(environment string) core.Level`

### Legacy Compatibility Methods
- `NewLogger() *Logger`
- `Warning(message string, context ...map[string]interface{})`
- `SetOutputFormat(format string)`
- `SetDebugLogging(enable bool)`
- `SetMinLevel(level int)`

## Formatter Functions (`logger_formatters.go`)

### JSON Formatter
- `NewJSONFormatter() *JSONFormatter`
- `NewPrettyJSONFormatter() *JSONFormatter`
- `Format(entry *core.LogEntry) (string, error)` (JSONFormatter method)

### Text Formatter
- `NewTextFormatter() *TextFormatter`
- `Format(entry *core.LogEntry) (string, error)` (TextFormatter method)

### Pretty Formatter
- `NewPrettyFormatter(colors bool) *PrettyFormatter`
- `Format(entry *core.LogEntry) (string, error)` (PrettyFormatter method)
- `formatLevel(level core.Level) string`
- `formatImportantContext(entry *core.LogEntry) []string`
- `colorize(color, text string) string`

### Helper Functions
- `getCaller(skip int) string`

## Output Functions (`logget_output.go`)

### Output Manager
- `NewOutputManager() *OutputManager`
- `AddOutput(name string, output Output) error`
- `RemoveOutput(name string) error`
- `WriteToOutput(name string, entry *core.LogEntry) error`
- `WriteToAll(entry *core.LogEntry) error`
- `Close() error`

### Console Output
- `NewConsoleOutput(formatter Formatter, colors bool) *ConsoleOutput`
- `Write(entry *core.LogEntry) error` (ConsoleOutput method)
- `addColor(level core.Level, message string) string`
- `Close() error` (ConsoleOutput method)

### File Output
- `NewSimpleFileOutput(formatter Formatter, filename string) (*SimpleFileOutput, error)`
- `Write(entry *core.LogEntry) error` (SimpleFileOutput method)
- `Close() error` (SimpleFileOutput method)

### Multi Output
- `NewMultiOutput(outputs ...Output) *MultiOutput`
- `AddOutput(output Output)`
- `Write(entry *core.LogEntry) error` (MultiOutput method)
- `Close() error` (MultiOutput method)

### Filtered Output
- `NewFilteredOutput(output Output, minLevel core.Level) *FilteredOutput`
- `Write(entry *core.LogEntry) error` (FilteredOutput method)
- `Close() error` (FilteredOutput method)

### Buffered Output
- `NewBufferedOutput(output Output, maxSize int) *BufferedOutput`
- `Write(entry *core.LogEntry) error` (BufferedOutput method)
- `Flush() error`
- `flushLocked() error`
- `Close() error` (BufferedOutput method)

## Type Functions (`logger_type.go`)

### Buffer Implementations
- `NewSimpleLogBuffer(maxSize int) *SimpleLogBuffer`
- `Add(entry *core.LogEntry) error` (SimpleLogBuffer method)
- `Flush() error` (SimpleLogBuffer method)
- `Close() error` (SimpleLogBuffer method)

### Channel Buffer
- `NewChannelLogBuffer(bufferSize int, processor func(*core.LogEntry) error) *ChannelLogBuffer`
- `Add(entry *core.LogEntry) error` (ChannelLogBuffer method)
- `process()` (internal goroutine method)
- `Flush() error` (ChannelLogBuffer method)
- `Close() error` (ChannelLogBuffer method)

### Filter Implementations
- `NewLevelFilter(minLevel core.Level) *LevelFilter`
- `ShouldLog(entry *core.LogEntry) bool` (LevelFilter method)
- `NewComponentFilter(components ...string) *ComponentFilter`
- `ShouldLog(entry *core.LogEntry) bool` (ComponentFilter method)
- `NewLayerFilter(layers ...string) *LayerFilter`
- `ShouldLog(entry *core.LogEntry) bool` (LayerFilter method)
- `NewCompositeFilter(mode FilterMode, filters ...LogFilter) *CompositeFilter`
- `ShouldLog(entry *core.LogEntry) bool` (CompositeFilter method)

## Core Type Methods

### Level Type
- `String() string` (Level method)

**Total Function Count: ~150+ functions**

This comprehensive logging package provides extensive functionality for enterprise-grade logging with specialized methods for different domains, multiple output formats, filtering capabilities, and both synchronous and asynchronous processing options.
# Go Logging System - Complete Guide

## Overview

This is a comprehensive, production-ready logging system for Go applications that provides structured logging, multiple output formats, thread safety, and specialized logging methods for different types of operations. The system automatically adapts to different environments and provides both simple and advanced logging capabilities.

## Key Features

- **Multi-format Output**: JSON, Pretty, and Text formats
- **Environment-aware**: Automatically adjusts behavior based on environment
- **Thread-safe**: Concurrent logging operations are safe
- **Structured Logging**: Rich context and metadata support
- **Specialized Methods**: Purpose-built methods for authentication, API requests, database operations, etc.
- **Performance Monitoring**: Built-in timing and performance categorization
- **Security-focused**: Automatic masking of sensitive data
- **Component-based**: Organize logs by application components

## Architecture

### Core Components

1. **Logger Structure**: Thread-safe logger with configurable outputs
2. **LogEntry**: Structured log entry with rich metadata
3. **Global Logger**: Singleton instance for application-wide use
4. **Component Loggers**: Specialized loggers for different parts of your application
5. **Formatting System**: Multiple output formats for different environments

### Log Levels

```go
const (
    DebugLevel   = 0  // Detailed debugging information
    InfoLevel    = 1  // General information
    WarningLevel = 2  // Warning conditions
    ErrorLevel   = 3  // Error conditions
    FatalLevel   = 4  // Fatal errors (causes program exit)
)
```

### Output Formats

- **JSON**: Structured JSON output for production/staging
- **Pretty**: Human-readable format with emojis and colors for development
- **Text**: Simple text format for basic logging needs

## Quick Start Guide

### 1. Basic Logging

```go
package main

import (
    "your-app/logger"
)

func main() {
    // Simple logging
    logger.Info("Application started")
    logger.Debug("Debug information")
    logger.Warning("This is a warning")
    logger.Error("An error occurred")
    
    // Logging with context
    logger.Info("User action", map[string]interface{}{
        "user_id": "123",
        "action": "login",
        "ip": "192.168.1.1",
    })
}
```

### 2. Component-based Logging

```go
// Create component-specific loggers
handlerLogger := logger.NewHandlerLogger()
serviceLogger := logger.NewServiceLogger()
repoLogger := logger.NewRepositoryLogger()

// Use in your handlers
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    handlerLogger.Info("Getting user", map[string]interface{}{
        "request_id": getRequestID(r),
        "path": r.URL.Path,
    })
}

// Use in your services
func (s *UserService) CreateUser(user User) error {
    serviceLogger.Info("Creating new user", map[string]interface{}{
        "email": user.Email,
        "operation": "create_user",
    })
}
```

### 3. Custom Logger Configuration

```go
// Create custom logger
customLogger := logger.NewLogger()
customLogger.SetComponent("payment")
customLogger.SetOutputFormat("json")
customLogger.AddGlobalField("service", "payment-service")

// Use custom logger
customLogger.Info("Payment processed", map[string]interface{}{
    "amount": 100.00,
    "currency": "USD",
    "transaction_id": "txn_123",
})
```

## Environment Configuration

### Environment Variables

The logger automatically configures itself based on environment variables:

```bash
# Set environment (affects default log level and format)
export APP_ENV=production          # Options: development, production, staging, testing
export ENVIRONMENT=production      # Alternative environment variable

# Override log format
export LOG_FORMAT=json            # Options: json, pretty, text

# These are automatically detected:
# - production/staging: JSON format, INFO level minimum
# - development/testing: Pretty format, DEBUG level minimum
```

### Environment Behavior

| Environment | Default Format | Default Level | Debug Enabled |
|-------------|---------------|---------------|---------------|
| development | pretty        | DEBUG         | Yes           |
| testing     | pretty        | DEBUG         | Yes           |
| staging     | json          | INFO          | No            |
| production  | json          | INFO          | No            |

## Specialized Logging Methods

### 1. Authentication Logging

```go
// Log authentication attempts
logger.LogAuthAttempt("user@example.com", true, "password_correct")
logger.LogAuthAttempt("user@example.com", false, "invalid_password", map[string]interface{}{
    "ip": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "attempt_count": 3,
})

// Output (Pretty format):
// [14:30:25.123] ℹ️  INFO [HANDLER] <authentication> Authentication successful | user=u***@example.com
// [14:30:26.456] ⚠️  WARN [HANDLER] <authentication> Authentication failed | user=u***@example.com ip=192.168.1.100
```

### 2. API Request Logging

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer to capture status code
        recorder := &statusRecorder{ResponseWriter: w}
        next.ServeHTTP(recorder, r)
        
        duration := time.Since(start)
        
        logger.LogAPIRequest(
            r.Method,
            r.URL.Path,
            recorder.statusCode,
            duration,
            map[string]interface{}{
                "user_id": getUserID(r),
                "ip": r.RemoteAddr,
                "user_agent": r.UserAgent(),
            },
        )
    })
}

// Output examples:
// [14:30:25.123] ℹ️  INFO [HANDLER] GET /api/users/123 → 200 | took=45ms status=200
// [14:30:26.456] ⚠️  WARN [HANDLER] POST /api/login → 401 | took=120ms status=401 reason=invalid_credentials
// [14:30:27.789] ❌ ERROR [HANDLER] GET /api/orders → 500 | took=2340ms status=500 error=database_connection_failed
```

### 3. Database Operation Logging

```go
func (r *UserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    start := time.Now()
    
    user, err := r.db.GetUser(ctx, id)
    
    logger.LogDBOperation("SELECT", "users", err == nil, err, map[string]interface{}{
        "user_id": id,
        "duration_ms": time.Since(start).Milliseconds(),
        "query_type": "get_by_id",
    })
    
    return user, err
}

// Output:
// [14:30:25.123] 🔍 DEBUG [REPOSITORY] DB SELECT on users succeeded | took=23ms
// [14:30:26.456] ❌ ERROR [REPOSITORY] DB INSERT on users failed | error=duplicate_key_violation
```

### 4. Service Call Logging

```go
func (s *PaymentService) ProcessPayment(payment Payment) error {
    err := s.externalPaymentAPI.Charge(payment)
    
    logger.LogServiceCall("payment-gateway", "charge", err == nil, err, map[string]interface{}{
        "amount": payment.Amount,
        "currency": payment.Currency,
        "payment_id": payment.ID,
        "gateway": "stripe",
    })
    
    return err
}

// Output:
// [14:30:25.123] 🔍 DEBUG [SERVICE] payment-gateway.charge succeeded | amount=100.00 currency=USD
// [14:30:26.456] ❌ ERROR [SERVICE] payment-gateway.charge failed | error=insufficient_funds retryable=false
```

### 5. User Activity Logging

```go
func (h *DocumentHandler) DownloadDocument(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r)
    userEmail := getUserEmail(r)
    documentID := r.URL.Query().Get("id")
    
    // ... download logic ...
    
    logger.LogUserActivity(userID, userEmail, "download", "document", map[string]interface{}{
        "document_id": documentID,
        "file_size": fileSize,
        "ip": r.RemoteAddr,
    })
}

// Output:
// [14:30:25.123] ℹ️  INFO [AUDIT] User j***@company.com performed download on document | user=j***@company.com document_id=doc_123
```

### 6. Security Event Logging

```go
func (m *SecurityMiddleware) DetectSuspiciousActivity(r *http.Request) {
    if isSuspicious(r) {
        logger.LogSecurityEvent(
            "suspicious_activity",
            "Multiple failed login attempts from same IP",
            "high",
            map[string]interface{}{
                "ip": r.RemoteAddr,
                "user_agent": r.UserAgent(),
                "attempt_count": getAttemptCount(r.RemoteAddr),
                "time_window": "5_minutes",
            },
        )
    }
}

// Output:
// [14:30:25.123] ❌ ERROR [SECURITY] Security: Multiple failed login attempts from same IP | ip=192.168.1.100 attempt_count=5
```

### 7. Performance Monitoring

```go
func (s *DataService) ProcessLargeDataset(data []DataItem) error {
    start := time.Now()
    defer func() {
        logger.LogPerformance("process_large_dataset", time.Since(start), map[string]interface{}{
            "item_count": len(data),
            "batch_size": 1000,
        })
    }()
    
    // ... processing logic ...
    return nil
}

// Output:
// [14:30:25.123] 🔍 DEBUG [BENCHMARK] Performance: process_large_dataset took 2.3s | item_count=5000 category=slow
// [14:30:26.456] ⚠️  WARN [BENCHMARK] Performance: process_large_dataset took 8.7s | item_count=10000 category=very_slow
```

### 8. Validation Error Logging

```go
func (v *UserValidator) ValidateUser(user User) error {
    if user.Email == "" {
        logger.LogValidationError("email", "email is required", user.Email)
        return errors.New("email is required")
    }
    
    if len(user.Password) < 8 {
        logger.LogValidationError("password", "password too short", user.Password)
        return errors.New("password too short")
    }
    
    return nil
}

// Output:
// [14:30:25.123] ⚠️  WARN [VALIDATOR] Validation failed for email | field=email value=""
// [14:30:26.456] ⚠️  WARN [VALIDATOR] Validation failed for password | field=password value=***hidden*** value_length=5
```

### 9. Metrics Logging

```go
func (m *MetricsCollector) RecordUserRegistration() {
    logger.LogMetric("user_registrations", 1, "count", map[string]interface{}{
        "source": "web_app",
        "timestamp": time.Now().Unix(),
    })
}

func (m *MetricsCollector) RecordResponseTime(endpoint string, duration time.Duration) {
    logger.LogMetric("response_time", duration.Milliseconds(), "milliseconds", map[string]interface{}{
        "endpoint": endpoint,
        "method": "GET",
    })
}

// Output:
// [14:30:25.123] ℹ️  INFO [METRICS] Metric: user_registrations = 1 count | source=web_app
// [14:30:26.456] ℹ️  INFO [METRICS] Metric: response_time = 150 milliseconds | endpoint=/api/users method=GET
```

## Output Format Examples

### Pretty Format (Development)

```
[14:30:25.123] ℹ️  INFO [HANDLER] <login> User authentication successful | user=j***@example.com ip=192.168.1.1 took=45ms
[14:30:26.456] ⚠️  WARN [SERVICE] <payment> Payment processing slow | amount=150.00 took=2300ms status=pending
[14:30:27.789] ❌ ERROR [REPOSITORY] DB INSERT on orders failed | error=connection_timeout retryable=true (db.go:245)
```

### JSON Format (Production)

```json
{
  "timestamp": "2024-01-15 14:30:25.123",
  "level": "INFO",
  "message": "User authentication successful",
  "context": {
    "user_id": "123",
    "email": "j***@example.com",
    "ip": "192.168.1.1",
    "duration_ms": 45,
    "operation": "login",
    "type": "auth_attempt"
  },
  "component": "handler",
  "environment": "production",
  "file": "auth.go",
  "line": 156
}
```

### Text Format (Simple)

```
[14:30:25] INFO: User authentication successful (user=j***@example.com 45ms)
[14:30:26] WARN: Payment processing slow (150.00 2300ms)
[14:30:27] ERROR: DB INSERT on orders failed (error=connection_timeout)
```

## Advanced Usage Patterns

### 1. Request Context Logging

```go
func WithRequestContext(r *http.Request) *logger.Logger {
    requestLogger := logger.NewLogger()
    requestLogger.AddGlobalField("request_id", getRequestID(r))
    requestLogger.AddGlobalField("user_id", getUserID(r))
    requestLogger.AddGlobalField("ip", r.RemoteAddr)
    return requestLogger
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    log := WithRequestContext(r)
    
    log.Info("Processing get user request")
    // All subsequent logs will include request context
}
```

### 2. Structured Error Handling

```go
func (s *UserService) CreateUser(user User) error {
    if err := s.validator.Validate(user); err != nil {
        logger.Error("User validation failed", map[string]interface{}{
            "error": err.Error(),
            "error_type": "validation_error",
            "user_email": logger.maskEmail(user.Email),
            "operation": "create_user",
        })
        return err
    }
    
    if err := s.repository.Save(user); err != nil {
        logger.Error("Failed to save user", map[string]interface{}{
            "error": err.Error(),
            "error_type": "database_error",
            "operation": "create_user",
            "retryable": logger.isRetryableError(err),
        })
        return err
    }
    
    logger.Info("User created successfully", map[string]interface{}{
        "user_id": user.ID,
        "user_email": logger.maskEmail(user.Email),
        "operation": "create_user",
    })
    
    return nil
}
```

### 3. Performance Monitoring Middleware

```go
func PerformanceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        next.ServeHTTP(w, r)
        
        duration := time.Since(start)
        
        // Log slow requests
        if duration > 2*time.Second {
            logger.LogPerformance("slow_request", duration, map[string]interface{}{
                "method": r.Method,
                "path": r.URL.Path,
                "user_id": getUserID(r),
            })
        }
    })
}
```

### 4. Contextual Business Logic Logging

```go
func (s *OrderService) ProcessOrder(order Order) error {
    orderLogger := logger.NewComponentLogger("order-processor")
    orderLogger.AddGlobalField("order_id", order.ID)
    orderLogger.AddGlobalField("customer_id", order.CustomerID)
    
    orderLogger.Info("Starting order processing")
    
    // Validate inventory
    if !s.inventory.HasStock(order.Items) {
        orderLogger.Warning("Insufficient inventory", map[string]interface{}{
            "requested_items": len(order.Items),
            "operation": "inventory_check",
        })
        return errors.New("insufficient inventory")
    }
    
    // Process payment
    if err := s.payment.Charge(order.Total); err != nil {
        orderLogger.Error("Payment failed", map[string]interface{}{
            "amount": order.Total,
            "error": err.Error(),
            "operation": "payment_processing",
        })
        return err
    }
    
    orderLogger.Info("Order processed successfully", map[string]interface{}{
        "total_amount": order.Total,
        "item_count": len(order.Items),
    })
    
    return nil
}
```

## Configuration and Customization

### 1. Environment-based Configuration

```go
func configureLogger() {
    switch os.Getenv("APP_ENV") {
    case "production":
        logger.SetOutputFormat("json")
        logger.SetMinLevel(logger.InfoLevel)
        logger.SetDebugLogging(false)
    case "development":
        logger.SetOutputFormat("pretty")
        logger.SetMinLevel(logger.DebugLevel)
        logger.SetDebugLogging(true)
    }
}
```

### 2. Custom Global Fields

```go
func init() {
    logger.AddGlobalField("service", "user-api")
    logger.AddGlobalField("version", "1.2.3")
    logger.AddGlobalField("build", os.Getenv("BUILD_NUMBER"))
    logger.AddGlobalField("instance_id", getInstanceID())
}
```

### 3. Component-specific Configuration

```go
// Create specialized loggers for different components
func NewDatabaseLogger() *logger.Logger {
    dbLogger := logger.NewComponentLogger("database")
    dbLogger.AddGlobalField("connection_pool", "primary")
    return dbLogger
}

func NewCacheLogger() *logger.Logger {
    cacheLogger := logger.NewComponentLogger("cache")
    cacheLogger.AddGlobalField("cache_type", "redis")
    return cacheLogger
}
```

## Security and Privacy Features

### 1. Automatic Data Masking

The logger automatically masks sensitive information:

- **Email addresses**: `john.doe@example.com` → `j***@example.com`
- **Passwords**: Always shown as `***hidden***`
- **Tokens**: Always shown as `***hidden***`
- **Secrets**: Always shown as `***hidden***`

### 2. Safe Context Logging

```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    credentials := parseCredentials(r)
    
    // Safe logging - sensitive fields are automatically masked
    logger.LogAuthAttempt(credentials.Email, false, "invalid_password", map[string]interface{}{
        "password_length": len(credentials.Password), // Length is safe to log
        "ip": r.RemoteAddr,
        "user_agent": r.UserAgent(),
    })
}
```

## Integration Patterns

### 1. Gin Web Framework Integration

```go
func GinLoggingMiddleware() gin.HandlerFunc {
    return gin.LoggerWithConfig(gin.LoggerConfig{
        Formatter: func(param gin.LogFormatterParams) string {
            logger.LogAPIRequest(
                param.Method,
                param.Path,
                param.StatusCode,
                param.Latency,
                map[string]interface{}{
                    "client_ip": param.ClientIP,
                    "user_agent": param.Request.UserAgent(),
                },
            )
            return "" // Return empty string since we're using our custom logger
        },
    })
}
```

### 2. GORM Integration

```go
func NewGormLogger() gormlogger.Interface {
    return &GormLogger{logger: logger.NewComponentLogger("gorm")}
}

type GormLogger struct {
    logger *logger.Logger
}

func (g *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
    return g
}

func (g *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    g.logger.Info(msg, map[string]interface{}{"data": data})
}

func (g *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    g.logger.Warning(msg, map[string]interface{}{"data": data})
}

func (g *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    g.logger.Error(msg, map[string]interface{}{"data": data})
}

func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    sql, rows := fc()
    duration := time.Since(begin)
    
    g.logger.LogDBOperation("QUERY", "multiple", err == nil, err, map[string]interface{}{
        "sql": sql,
        "rows_affected": rows,
        "duration_ms": duration.Milliseconds(),
    })
}
```

### 3. HTTP Client Integration

```go
type LoggingHTTPClient struct {
    client *http.Client
    logger *logger.Logger
}

func (c *LoggingHTTPClient) Do(req *http.Request) (*http.Response, error) {
    start := time.Now()
    
    resp, err := c.client.Do(req)
    duration := time.Since(start)
    
    context := map[string]interface{}{
        "method": req.Method,
        "url": req.URL.String(),
        "duration_ms": duration.Milliseconds(),
    }
    
    if resp != nil {
        context["status_code"] = resp.StatusCode
    }
    
    c.logger.LogServiceCall("http_client", req.Method, err == nil, err, context)
    
    return resp, err
}
```

## Best Practices

### 1. Consistent Context Keys

Use consistent keys across your application:

```go
const (
    ContextKeyUserID    = "user_id"
    ContextKeyRequestID = "request_id"
    ContextKeyOperation = "operation"
    ContextKeyComponent = "component"
    ContextKeyDuration  = "duration_ms"
)
```

### 2. Log Level Guidelines

- **DEBUG**: Detailed debugging information, disabled in production
- **INFO**: General application flow, successful operations
- **WARNING**: Potentially harmful situations, slow operations
- **ERROR**: Error events that don't stop application execution
- **FATAL**: Very severe errors that cause application termination

### 3. Performance Considerations

```go
// Good: Use appropriate log levels
if logger.IsDebugEnabled() {
    logger.Debug("Expensive debug info", expensiveDebugContext())
}

// Good: Avoid expensive operations in production
func expensiveDebugContext() map[string]interface{} {
    if !logger.IsDebugEnabled() {
        return nil
    }
    // Expensive context building
    return buildExpensiveContext()
}
```

### 4. Error Context

Always provide rich context for errors:

```go
func (s *PaymentService) ProcessPayment(payment Payment) error {
    err := s.gateway.Charge(payment.Amount)
    if err != nil {
        // Rich error context
        logger.Error("Payment processing failed", map[string]interface{}{
            "error": err.Error(),
            "payment_id": payment.ID,
            "amount": payment.Amount,
            "currency": payment.Currency,
            "customer_id": payment.CustomerID,
            "gateway": "stripe",
            "retry_count": payment.RetryCount,
            "error_type": fmt.Sprintf("%T", err),
        })
        return err
    }
    return nil
}
```

## Troubleshooting

### Common Issues

1. **Logs not appearing**
   - Check `APP_ENV` or `ENVIRONMENT` variable
   - Verify log level configuration
   - Ensure DEBUG logging is enabled for debug messages

2. **Wrong format output**
   - Check `LOG_FORMAT` environment variable
   - Verify environment-based format defaults

3. **Missing context**
   - Ensure global fields are set correctly
   - Check component logger configuration

### Debug Configuration

```go
// Enable debug logging
logger.SetDebugLogging(true)
logger.SetMinLevel(logger.DebugLevel)
logger.SetOutputFormat("pretty")

// Add debug context
logger.AddGlobalField("debug", true)
logger.AddGlobalField("pid", os.Getpid())
```

This comprehensive logging system provides everything you need for production-ready Go applications with excellent observability, security, and developer experience.
# Complete Multi-Domain Error Handling System Guide

A comprehensive, production-ready error handling system for Go APIs supporting multiple business domains with advanced configuration management and robust error processing.

## 🚀 System Overview

This enhanced error handling system provides enterprise-grade error management with:

- **Multi-Domain Architecture**: Support for user, course, payment, auth, admin, content, and system domains
- **Configuration-Driven**: Environment-specific error handling behavior through Viper configuration
- **Domain-Aware Error Codes**: Dynamic error code generation with domain prefixes
- **Rich Error Context**: Detailed error information with request tracking and business context
- **Production-Ready**: Comprehensive logging, middleware, and HTTP handling
- **Backward Compatible**: Maintains compatibility with existing error types

## 📁 Complete File Structure

```
internal/error_custom/
├── core.go              # Core error types, interfaces, and APIError
├── codes.go             # Domain-aware error code system
├── constructors.go      # Error creation functions for all domains
├── utilities.go         # Error detection, parsing, and utility functions
└── handler.go           # HTTP middleware, error handling, and request parsing

utils_config/
└── config.go            # Configuration management with domain-specific settings
```

## 🏗️ Enhanced Architecture

### Core Components

1. **APIError with Full Context**: Central error type with domain, layer, operation, and request tracking
2. **Domain-Aware Configuration**: Configurable error handling behavior per domain and environment
3. **Generic Error Types**: Reusable error types (NotFoundError, ValidationError, BusinessLogicError, etc.)
4. **Error Collections**: Aggregate multiple related errors with domain context
5. **Advanced Middleware**: Request ID tracking, domain context, recovery, and logging
6. **Configuration Management**: Environment-specific error handling settings

### Supported Domains

- `user` - User management, authentication, and profiles
- `course` - Learning content, enrollment, and course management
- `payment` - Payment processing, billing, and financial operations
- `auth` - Authentication, authorization, and security
- `admin` - Administrative operations and management
- `content` - Content management and media handling
- `system` - System-level operations and infrastructure

## 🔧 Advanced Usage Examples

### Configuration-Driven Error Handling

```go
// Initialize configuration with domain-specific settings
func main() {
    err := utils_config.InitializeConfig("./config.yaml")
    if err != nil {
        panic(err)
    }
    
    config := utils_config.GetConfig()
    
    // Create domain-aware error handler
    errorHandler := utils_config.NewDomainAwareErrorHandler(config)
    
    // Use in your application
    if err := someOperation(); err != nil {
        handledErr := errorHandler.HandleError("user", err)
        // Process handled error...
    }
}
```

### Advanced Error Creation with Rich Context

```go
// User domain errors with full context
func HandleUserLogin(email, password string) error {
    requestID := generateRequestID()
    
    // Email validation with domain context
    if err := errorcustom.ValidateEmailWithDomain(email, "user", requestID); err != nil {
        return err
    }
    
    // Business logic error with rich context
    if user.IsLocked() {
        return errorcustom.NewAccountLockedError(email, "too_many_failed_attempts")
    }
    
    // Authentication error with step tracking
    if !verifyPassword(password, user.PasswordHash) {
        return errorcustom.NewPasswordMismatchError(email)
    }
    
    return nil
}

// Course domain errors with enrollment logic
func HandleCourseEnrollment(userID, courseID int64) error {
    // Check course capacity
    if course.CurrentEnrollment >= course.MaxCapacity {
        return errorcustom.NewCourseCapacityExceededError(
            courseID, 
            course.MaxCapacity, 
            course.CurrentEnrollment,
        )
    }
    
    // Check enrollment period
    if time.Now().After(course.EnrollmentDeadline) {
        return errorcustom.NewCourseEnrollmentClosedError(courseID)
    }
    
    return nil
}

// Payment domain errors with retry logic
func ProcessPayment(userID int64, amount float64) error {
    // Check sufficient funds
    if userBalance < amount {
        return errorcustom.NewInsufficientFundsError(userID, amount, userBalance)
    }
    
    // External service error with retry capability
    if err := paymentProvider.Charge(amount); err != nil {
        return errorcustom.NewPaymentProviderError(
            "stripe", 
            "charge", 
            err, 
            true, // retryable
        )
    }
    
    return nil
}
```

### Complete HTTP Handler with Domain Context

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    requestID := errorcustom.GetRequestIDFromContext(r.Context())
    domain := errorcustom.GetDomainFromContext(r.Context()) // "user" from middleware
    
    // Parse and validate request
    var req CreateUserRequest
    if err := errorcustom.DecodeJSONWithDomain(r.Body, &req, domain, requestID); err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    // Validate request with error collection
    errorCollection := errorcustom.NewErrorCollection(domain)
    
    // Email validation
    if err := errorcustom.ValidateEmailWithDomain(req.Email, domain, requestID); err != nil {
        errorCollection.Add(err)
    }
    
    // Password validation
    if err := errorcustom.ValidatePasswordWithDomain(req.Password, domain, requestID); err != nil {
        errorCollection.Add(err)
    }
    
    // Handle validation errors
    if errorCollection.HasErrors() {
        errorcustom.HandleError(w, errorCollection.ToAPIError(), requestID)
        return
    }
    
    // Business logic
    user, err := userService.CreateUser(req)
    if err != nil {
        // Convert service errors to domain-aware API errors
        apiErr := errorcustom.ParseGRPCError(err, domain, "create_user", map[string]interface{}{
            "email": req.Email,
        })
        errorcustom.HandleDomainError(w, apiErr, domain, requestID)
        return
    }
    
    // Success response with domain context
    errorcustom.RespondWithDomainSuccess(w, user, domain, requestID)
}
```

### Advanced Middleware Setup

```go
func SetupRoutes() chi.Router {
    r := chi.NewRouter()
    
    // Global middleware with comprehensive error handling
    r.Use(errorcustom.RequestIDMiddleware)          // Generate unique request IDs
    r.Use(errorcustom.LogHTTPMiddleware)            // Domain-aware logging
    r.Use(errorcustom.RecoveryMiddleware)           // Panic recovery with domain context
    
    // Domain-specific route groups
    r.Route("/api/users", func(r chi.Router) {
        r.Use(errorcustom.DomainMiddleware("user"))
        r.Post("/", CreateUserHandler)
        r.Get("/{id}", GetUserHandler)
        r.Put("/{id}", UpdateUserHandler)
        r.Delete("/{id}", DeleteUserHandler)
    })
    
    r.Route("/api/courses", func(r chi.Router) {
        r.Use(errorcustom.DomainMiddleware("course"))
        r.Get("/", ListCoursesHandler)
        r.Post("/", CreateCourseHandler)
        r.Post("/{course_id}/enroll", EnrollInCourseHandler)
    })
    
    r.Route("/api/payments", func(r chi.Router) {
        r.Use(errorcustom.DomainMiddleware("payment"))
        r.Post("/", ProcessPaymentHandler)
        r.Get("/{payment_id}", GetPaymentStatusHandler)
    })
    
    return r
}
```

### Environment-Specific Configuration

```yaml
# config.yaml
environment: "production"
app_name: "English AI"

# Domain configuration
domains:
  enabled:
    - "user"
    - "course"
    - "payment"
    - "auth"
    - "admin"
    - "content"
    - "system"
  default: "system"
  error_tracking:
    enabled: true
    log_level: "info"
  user:
    max_login_attempts: 3        # Stricter in production
    password_complexity: true
    email_verification: true
  course:
    enrollment_validation: true
    prerequisite_check: true
  payment:
    provider_timeout: "30s"
    retry_attempts: 3
    webhook_validation: true

# Error handling configuration
error_handling:
  include_stack_trace: false     # Security consideration for production
  sanitize_sensitive_data: true
  request_id_required: true

# Development overrides
---
environment: "development"
domains:
  error_tracking:
    log_level: "debug"
  user:
    max_login_attempts: 10       # More lenient in development
error_handling:
  include_stack_trace: true      # Debugging information
```

## 🔍 Advanced Error Detection

### Comprehensive Error Type Checking

```go
// Domain-specific error detection
if errorcustom.IsDomainError(err, "user") {
    // Handle user domain errors
}

if errorcustom.IsUserNotFoundError(err) {
    // Specific user not found handling
}

// Error type classification
if errorcustom.IsRetryableError(err) {
    // Implement retry logic
    go retryOperation(operation, maxRetries)
}

if errorcustom.IsClientError(err) {
    // Client-side error - don't retry
    logClientError(err)
} else if errorcustom.IsServerError(err) {
    // Server-side error - investigate
    alertOpsTeam(err)
}

// Business logic error handling
if errorcustom.IsBusinessLogicError(err) {
    // Handle business rule violations
    notifyBusinessOwner(err)
}
```

### gRPC Error Integration

```go
func HandleGRPCServiceCall(req *UserRequest) (*UserResponse, error) {
    resp, err := grpcClient.GetUser(context.Background(), req)
    if err != nil {
        // Convert gRPC error to domain-aware API error
        domainErr := errorcustom.ParseGRPCError(err, "user", "get_user", map[string]interface{}{
            "user_id": req.UserId,
            "email":   req.Email,
        })
        return nil, domainErr
    }
    return resp, nil
}
```

## 🌐 Dynamic Error Code System

### Error Code Patterns

The system generates consistent, domain-aware error codes:

```go
// Generic codes (no domain prefix)
"NOT_FOUND"
"VALIDATION_ERROR"
"AUTHENTICATION_ERROR"

// Domain-specific codes (with domain prefix)
"user_NOT_FOUND"           // User not found
"course_VALIDATION_ERROR"   // Course validation failure
"payment_EXTERNAL_SERVICE_ERROR"  // Payment provider error
"auth_AUTHORIZATION_ERROR"  // Authorization failure
"content_BUSINESS_LOGIC_ERROR"    // Content business rule violation
```

### Code Generation and Utilities

```go
// Generate domain-specific codes dynamically
userNotFoundCode := errorcustom.GetNotFoundCode("user")        // "user_NOT_FOUND"
paymentErrorCode := errorcustom.GetExternalServiceCode("payment") // "payment_EXTERNAL_SERVICE_ERROR"

// Extract domain from error codes
domain := errorcustom.ExtractDomainFromCode("user_VALIDATION_ERROR") // "user"
baseType := errorcustom.GetBaseErrorType("user_VALIDATION_ERROR")    // "VALIDATION_ERROR"

// Check if code belongs to domain
isUserError := errorcustom.IsErrorCodeForDomain("user_NOT_FOUND", "user") // true
```

## 📊 Enhanced Response Formats

### Successful Response with Domain Context

```json
{
  "success": true,
  "data": {
    "id": 123,
    "email": "user@example.com",
    "name": "John Doe"
  },
  "domain": "user"
}
```

### Single Error Response

```json
{
  "code": "user_NOT_FOUND",
  "message": "User with ID 123 not found",
  "details": {
    "resource_id": 123,
    "resource_type": "user",
    "retryable": false
  }
}
```

### Multiple Validation Errors

```json
{
  "code": "user_VALIDATION_ERROR",
  "message": "Multiple validation errors occurred",
  "details": {
    "errors": [
      {
        "code": "user_VALIDATION_ERROR",
        "message": "Email format is invalid",
        "details": {
          "field": "email",
          "value": "invalid-email",
          "expected_format": "user@domain.com"
        }
      },
      {
        "code": "user_VALIDATION_ERROR", 
        "message": "Password does not meet security requirements",
        "details": {
          "field": "password",
          "requirements": ["at least one uppercase letter", "at least one number"]
        }
      }
    ]
  }
}
```

### Business Logic Error with Context

```json
{
  "code": "course_BUSINESS_LOGIC_ERROR",
  "message": "Course has reached maximum capacity",
  "details": {
    "rule": "enrollment_capacity",
    "course_id": 456,
    "max_capacity": 100,
    "current_enrollment": 100
  }
}
```

## 🛠️ Configuration Integration

### Domain-Specific Error Handling

```go
// Configuration-driven error handler initialization
config := utils_config.GetConfig()
errorHandler := utils_config.NewDomainAwareErrorHandler(config)

// Example usage in service layer
func (s *UserService) ValidateLogin(email, password string) error {
    // Check max login attempts from configuration
    maxAttempts := config.GetMaxLoginAttempts()
    if user.FailedAttempts >= maxAttempts {
        return errorcustom.NewAccountLockedError(email, "max_attempts_exceeded")
    }
    
    // Use configuration for password complexity
    if config.IsPasswordComplexityRequired() {
        if err := errorcustom.ValidatePasswordWithDomain(password, "user", requestID); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Environment-Specific Behavior

```go
// Development environment - more lenient, detailed errors
if config.Environment == "development" {
    // Include stack traces
    apiErr.WithDetail("stack_trace", string(debug.Stack()))
    
    // More detailed error messages
    apiErr.WithDetail("internal_error", originalError.Error())
}

// Production environment - security-focused
if config.Environment == "production" {
    // Sanitize sensitive data
    if config.ShouldSanitizeSensitiveData() {
        apiErr = sanitizeErrorDetails(apiErr)
    }
    
    // Generic error messages for security
    if errorcustom.IsSystemError(err) {
        apiErr.Message = "Internal server error"
    }
}
```

## 🔒 Security Features

### Sensitive Data Sanitization

```go
// Automatic password redaction
passwordErr := errorcustom.NewValidationError(
    "user",
    "password", 
    "Password too weak",
    "[REDACTED]", // Value automatically redacted
)

// Context sanitization in production
func sanitizeErrorDetails(apiErr *APIError) *APIError {
    sensitiveFields := []string{"password", "token", "secret", "key"}
    
    for _, field := range sensitiveFields {
        if _, exists := apiErr.Details[field]; exists {
            apiErr.Details[field] = "[REDACTED]"
        }
    }
    
    return apiErr
}
```

### Request Tracking and Security

```go
// Every request gets unique tracking ID
requestID := errorcustom.GetRequestIDFromContext(r.Context())

// Client IP extraction with proxy support
clientIP := errorcustom.GetClientIP(r) // Handles X-Forwarded-For, X-Real-IP

// Security audit logging
logger.Info("User authentication failed", map[string]interface{}{
    "email":      email,
    "ip":         clientIP,
    "request_id": requestID,
    "domain":     "user",
    "reason":     "invalid_credentials",
})
```

## 📈 Advanced Request Processing

### Safe Parameter Parsing

```go
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    requestID := errorcustom.GetRequestIDFromContext(r.Context())
    domain := errorcustom.GetDomainFromContext(r.Context())
    
    // Safe ID parameter parsing with domain context
    userID, err := errorcustom.ParseIDParamWithDomain(r, "id", domain)
    if err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    // Safe pagination with domain validation
    limit, offset, err := errorcustom.GetPaginationParamsWithDomain(r, domain)
    if err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    // Safe sorting with allowlist validation
    allowedSortFields := []string{"id", "email", "created_at", "updated_at"}
    sortBy, sortOrder, err := errorcustom.GetSortParamsWithDomain(r, allowedSortFields, domain)
    if err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    // Business logic...
}
```

### Advanced Validation with Error Collections

```go
func ValidateComplexRequest(req *ComplexRequest, domain string) error {
    errorCollection := errorcustom.NewErrorCollection(domain)
    
    // Multiple field validations
    if err := errorcustom.ValidateEmailWithDomain(req.Email, domain, ""); err != nil {
        errorCollection.Add(err)
    }
    
    if err := errorcustom.ValidatePasswordWithDomain(req.Password, domain, ""); err != nil {
        errorCollection.Add(err)
    }
    
    // Custom business rule validation
    if req.Age < 18 {
        ageErr := errorcustom.NewBusinessLogicErrorWithContext(
            domain,
            "minimum_age",
            "User must be at least 18 years old",
            map[string]interface{}{
                "provided_age": req.Age,
                "required_age": 18,
            },
        )
        errorCollection.Add(ageErr)
    }
    
    // Return all errors at once for better UX
    if errorCollection.HasErrors() {
        return errorCollection.ToAPIError()
    }
    
    return nil
}
```

## 🎯 Domain-Specific Examples

### User Domain - Complete Authentication Flow

```go
func AuthenticateUser(email, password string) error {
    requestID := generateRequestID()
    
    // Step 1: Email validation
    if err := errorcustom.ValidateEmailWithDomain(email, "user", requestID); err != nil {
        return err
    }
    
    // Step 2: Find user
    user, err := userRepo.FindByEmail(email)
    if err != nil {
        if errorcustom.IsNotFoundError(err) {
            return errorcustom.NewEmailNotFoundError(email)
        }
        return errorcustom.NewDatabaseError("select", "users", err)
    }
    
    // Step 3: Check account status
    if !user.IsActive {
        return errorcustom.NewAccountDisabledError(email)
    }
    
    if user.IsLocked() {
        return errorcustom.NewAccountLockedError(email, "security_policy")
    }
    
    // Step 4: Verify password
    if !verifyPassword(password, user.PasswordHash) {
        return errorcustom.NewPasswordMismatchError(email)
    }
    
    return nil
}
```

### Course Domain - Enrollment Business Logic

```go
func EnrollUserInCourse(userID, courseID int64) error {
    // Comprehensive enrollment validation
    errorCollection := errorcustom.NewErrorCollection("course")
    
    // Check course exists
    course, err := courseRepo.FindByID(courseID)
    if err != nil {
        if errorcustom.IsNotFoundError(err) {
            return errorcustom.NewCourseNotFoundError(courseID)
        }
        return err
    }
    
    // Check enrollment capacity
    if course.CurrentEnrollment >= course.MaxCapacity {
        errorCollection.Add(errorcustom.NewCourseCapacityExceededError(
            courseID, 
            course.MaxCapacity, 
            course.CurrentEnrollment,
        ))
    }
    
    // Check enrollment period
    if time.Now().After(course.EnrollmentDeadline) {
        errorCollection.Add(errorcustom.NewCourseEnrollmentClosedError(courseID))
    }
    
    // Check user authorization
    if !canUserEnroll(userID, courseID) {
        errorCollection.Add(errorcustom.NewCourseAccessDeniedError(userID, courseID))
    }
    
    if errorCollection.HasErrors() {
        return errorCollection.ToAPIError()
    }
    
    return nil
}
```

### Payment Domain - Transaction Processing

```go
func ProcessTransaction(userID int64, amount float64, paymentMethod string) error {
    requestID := generateRequestID()
    
    // Validate payment amount
    if amount <= 0 {
        return errorcustom.NewValidationError(
            "payment",
            "amount",
            "Payment amount must be greater than zero",
            amount,
        )
    }
    
    // Check user balance (if applicable)
    balance, err := getUserBalance(userID)
    if err != nil {
        return errorcustom.NewExternalServiceError(
            "payment",
            "balance_service",
            "get_balance",
            "Failed to retrieve user balance",
            err,
            true, // retryable
        )
    }
    
    // Business logic validation
    if balance < amount {
        return errorcustom.NewInsufficientFundsError(userID, amount, balance)
    }
    
    // Process with external provider
    if err := paymentProvider.ProcessPayment(amount, paymentMethod); err != nil {
        // Determine if error is retryable based on provider response
        retryable := isPaymentProviderErrorRetryable(err)
        
        return errorcustom.NewPaymentProviderError(
            "stripe",
            "process_payment",
            err,
            retryable,
        )
    }
    
    return nil
}
```

## 🔄 Migration and Backward Compatibility

### Gradual Migration Strategy

```go
// Phase 1: Use legacy errors alongside new system
func LegacyHandler(w http.ResponseWriter, r *http.Request) {
    // Old error handling (still works)
    if err := someOldFunction(); err != nil {
        // Convert legacy error to new system
        apiErr := errorcustom.ConvertToAPIError(err)
        errorcustom.HandleError(w, apiErr, requestID)
        return
    }
}

// Phase 2: Mixed approach with domain context
func TransitionHandler(w http.ResponseWriter, r *http.Request) {
    domain := "user"
    requestID := errorcustom.GetRequestIDFromContext(r.Context())
    
    // Mix of old and new constructors
    if err := validateLegacyWay(); err != nil {
        // Wrap in domain context
        apiErr := errorcustom.ConvertToAPIError(err).WithDomain(domain)
        errorcustom.HandleError(w, apiErr, requestID)
        return
    }
}

// Phase 3: Full new system implementation
func ModernHandler(w http.ResponseWriter, r *http.Request) {
    domain := errorcustom.GetDomainFromContext(r.Context())
    requestID := errorcustom.GetRequestIDFromContext(r.Context())
    
    // Pure new system approach
    if err := errorcustom.NewUserNotFoundByID(123); err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
}
```

## 📊 Monitoring and Observability

### Structured Logging Integration

```go
// Different log levels based on error severity
func LogErrorWithContext(err error, requestID string) {
    apiErr := errorcustom.ConvertToAPIError(err)
    severity := errorcustom.GetErrorSeverity(apiErr)
    logContext := apiErr.GetLogContext()
    logContext["request_id"] = requestID
    
    switch severity {
    case "ERROR":
        // Critical system failures (5xx errors)
        logger.Error("Critical system error", logContext)
        
        // Alert operations team
        alertOpsTeam(apiErr)
        
    case "WARNING":
        // External service issues, retryable errors
        logger.Warning("Service error occurred", logContext)
        
        // Monitor for patterns
        incrementErrorMetric(apiErr.Domain, apiErr.Code)
        
    case "INFO":
        // Client errors (4xx), expected business logic errors
        logger.Info("Request error occurred", logContext)
        
        // Track for analytics
        trackClientError(apiErr)
        
    default:
        logger.Debug("Request processed with error", logContext)
    }
}
```

### Error Metrics and Analytics

```go
// Track error patterns by domain
func TrackErrorMetrics(apiErr *APIError) {
    metrics := map[string]interface{}{
        "domain":      apiErr.Domain,
        "error_code":  apiErr.Code,
        "error_type":  errorcustom.GetBaseErrorType(apiErr.Code),
        "http_status": apiErr.HTTPStatus,
        "retryable":   apiErr.Retryable,
        "timestamp":   time.Now(),
    }
    
    // Send to monitoring system
    metricsCollector.Increment("api_errors_total", metrics)
    
    // Track domain-specific patterns
    if apiErr.Domain != "" {
        metricsCollector.Increment(
            fmt.Sprintf("domain_%s_errors", apiErr.Domain),
            metrics,
        )
    }
}
```

## 🧪 Comprehensive Testing Patterns

### Domain Error Testing

```go
func TestUserDomainErrors(t *testing.T) {
    tests := []struct {
        name           string
        errorFunc      func() error
        expectedCode   string
        expectedStatus int
        expectedDomain string
    }{
        {
            name:           "user not found by ID",
            errorFunc:      func() error { return errorcustom.NewUserNotFoundByID(123) },
            expectedCode:   "user_NOT_FOUND",
            expectedStatus: 404,
            expectedDomain: "user",
        },
        {
            name:           "duplicate email",
            errorFunc:      func() error { return errorcustom.NewDuplicateEmailError("test@example.com") },
            expectedCode:   "user_DUPLICATE",
            expectedStatus: 409,
            expectedDomain: "user",
        },
        {
            name:           "weak password",
            errorFunc:      func() error { return errorcustom.NewWeakPasswordError([]string{"uppercase"}) },
            expectedCode:   "user_VALIDATION_ERROR",
            expectedStatus: 400,
            expectedDomain: "user",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.errorFunc()
            apiErr := errorcustom.ConvertToAPIError(err)
            
            assert.Equal(t, tt.expectedCode, apiErr.Code)
            assert.Equal(t, tt.expectedStatus, apiErr.HTTPStatus)
            assert.Equal(t, tt.expectedDomain, apiErr.Domain)
            assert.NotEmpty(t, apiErr.Details)
        })
    }
}
```

### Error Collection Testing

```go
func TestErrorCollection(t *testing.T) {
    collection := errorcustom.NewErrorCollection("user")
    
    // Add multiple errors
    collection.Add(errorcustom.NewValidationError("user", "email", "Invalid format", "bad-email"))
    collection.Add(errorcustom.NewValidationError("user", "password", "Too weak", "[REDACTED]"))
    
    assert.True(t, collection.HasErrors())
    
    apiErr := collection.ToAPIError()
    assert.Equal(t, "user_VALIDATION_ERROR", apiErr.Code)
    assert.Equal(t, "user", apiErr.Domain)
    
    errors, ok := apiErr.Details["errors"].([]map[string]interface{})
    assert.True(t, ok)
    assert.Len(t, errors, 2)
}
```

## 🚀 Performance Optimizations

### Error Reuse and Caching

```go
// Pre-create common errors for performance
var (
    commonUserErrors = map[string]*APIError{
        "invalid_email": errorcustom.NewValidationError("user", "email", "Invalid email format", nil).ToAPIError(),
        "weak_password": errorcustom.NewWeakPasswordError([]string{"complexity"}).ToAPIError(),
    }
)

// Reuse common errors
func GetCommonUserError(errorType string) *APIError {
    if err, exists := commonUserErrors[errorType]; exists {
        // Clone to avoid mutation
        return cloneAPIError(err)
    }
    return nil
}
```

### Efficient Error Processing

```go
// Smart logging to avoid noise
func (deh *DomainErrorHandler) HandleError(domain string, err error) error {
    // Only log if it should be logged based on configuration
    if errorcustom.ShouldLogError(err) {
        severity := errorcustom.GetErrorSeverity(err)
        
        // Use appropriate log level based on configuration
        logLevel := deh.config.GetDomainErrorLogLevel()
        if shouldLog(severity, logLevel) {
            logErrorWithSeverity(err, severity)
        }
    }
    
    return err
}
```

## 🔮 Advanced Features

### Custom Error Templates

```go
// Domain-specific error message templates
var errorTemplates = map[string]map[string]string{
    "user": {
        "NOT_FOUND":        "User account not found",
        "DUPLICATE":        "An account with this email already exists",
        "AUTHENTICATION":   "Invalid login credentials",
    },
    "course": {
        "NOT_FOUND":        "Course not available",
        "BUSINESS_LOGIC":   "Course enrollment requirements not met",
        "AUTHORIZATION":    "Course access not permitted",
    },
    "payment": {
        "EXTERNAL_SERVICE": "Payment processing temporarily unavailable",
        "BUSINESS_LOGIC":   "Payment requirements not satisfied",
    },
}

// GetErrorTemplate returns domain-specific error message template
func GetErrorTemplate(domain, errorType string) string {
    if domainTemplates, exists := errorTemplates[domain]; exists {
        if template, exists := domainTemplates[errorType]; exists {
            return template
        }
    }
    
    // Fallback to generic template
    return fmt.Sprintf("An error occurred in %s domain", domain)
}
```

### Webhook Error Handling

```go
// Payment webhook error handling with domain context
func PaymentWebhookHandler(w http.ResponseWriter, r *http.Request) {
    requestID := errorcustom.GetRequestIDFromContext(r.Context())
    domain := "payment"
    
    // Validate webhook signature
    if !validateWebhookSignature(r) {
        authErr := errorcustom.NewAuthenticationError(domain, "invalid webhook signature")
        errorcustom.HandleDomainError(w, authErr, domain, requestID)
        return
    }
    
    // Parse webhook payload
    var payload WebhookPayload
    if err := errorcustom.DecodeJSONWithDomain(r.Body, &payload, domain, requestID); err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    // Process webhook with error handling
    if err := processPaymentWebhook(payload); err != nil {
        // Convert to domain-aware error
        domainErr := errorcustom.ParseGRPCError(err, domain, "process_webhook", map[string]interface{}{
            "payment_id": payload.PaymentID,
            "event_type": payload.EventType,
        })
        errorcustom.HandleDomainError(w, domainErr, domain, requestID)
        return
    }
    
    errorcustom.RespondWithDomainSuccess(w, map[string]string{"status": "processed"}, domain, requestID)
}
```

### Batch Operation Error Handling

```go
// Handle bulk operations with partial success tracking
func BulkUpdateUsers(updates []UserUpdate) (*BulkResult, error) {
    result := &BulkResult{
        TotalProcessed: 0,
        Successful:     0,
        Failed:         0,
        Errors:         make([]BulkError, 0),
    }
    
    errorCollection := errorcustom.NewErrorCollection("user")
    
    for i, update := range updates {
        result.TotalProcessed++
        
        // Validate individual update
        if err := validateUserUpdate(update); err != nil {
            result.Failed++
            
            // Create bulk-specific error with context
            bulkErr := errorcustom.NewValidationErrorWithRules(
                "user",
                fmt.Sprintf("updates[%d]", i),
                "Bulk update validation failed",
                update,
                map[string]interface{}{
                    "index":     i,
                    "operation": "bulk_update",
                },
            )
            
            errorCollection.Add(bulkErr)
            result.Errors = append(result.Errors, BulkError{
                Index:   i,
                Message: err.Error(),
                Code:    bulkErr.ToAPIError().Code,
            })
            continue
        }
        
        // Process individual update
        if err := processUserUpdate(update); err != nil {
            result.Failed++
            
            domainErr := errorcustom.ParseGRPCError(err, "user", "bulk_update", map[string]interface{}{
                "index":   i,
                "user_id": update.UserID,
            })
            
            errorCollection.Add(domainErr)
            result.Errors = append(result.Errors, BulkError{
                Index:   i,
                Message: domainErr.Error(),
                Code:    errorcustom.ConvertToAPIError(domainErr).Code,
            })
            continue
        }
        
        result.Successful++
    }
    
    // Return partial success with error details
    if errorCollection.HasErrors() {
        result.PartialSuccess = result.Successful > 0
        return result, errorCollection.ToAPIError()
    }
    
    return result, nil
}
```

## 🎛️ Advanced Configuration Usage

### Runtime Configuration Updates

```go
// Dynamic configuration updates for error handling
func UpdateErrorHandlingConfig(domain string, config DomainConfig) error {
    currentConfig := utils_config.GetConfig()
    
    // Validate configuration changes
    if !currentConfig.IsDomainEnabled(domain) {
        return errorcustom.NewValidationError(
            "admin",
            "domain",
            "Cannot update configuration for disabled domain",
            domain,
        )
    }
    
    // Update domain-specific settings
    switch domain {
    case "user":
        if config.User.MaxLoginAttempts < 1 || config.User.MaxLoginAttempts > 20 {
            return errorcustom.NewValidationErrorWithRules(
                "admin",
                "max_login_attempts",
                "Max login attempts must be between 1 and 20",
                config.User.MaxLoginAttempts,
                map[string]interface{}{
                    "min": 1,
                    "max": 20,
                },
            )
        }
        
    case "payment":
        if config.Payment.RetryAttempts > 10 {
            return errorcustom.NewValidationError(
                "admin",
                "retry_attempts",
                "Payment retry attempts cannot exceed 10",
                config.Payment.RetryAttempts,
            )
        }
    }
    
    // Apply configuration updates
    return applyDomainConfig(domain, config)
}
```

### Feature Flag Integration

```go
// Feature flag driven error handling
func HandleFeatureFlaggedOperation(operation string, domain string) error {
    config := utils_config.GetConfig()
    
    // Check if feature is enabled for domain
    if !isFeatureEnabledForDomain(operation, domain) {
        return errorcustom.NewBusinessLogicErrorWithContext(
            domain,
            "feature_availability",
            fmt.Sprintf("Feature '%s' is not available in %s domain", operation, domain),
            map[string]interface{}{
                "feature":   operation,
                "domain":    domain,
                "available": false,
            },
        )
    }
    
    return nil
}
```

## 🛡️ Advanced Security Patterns

### Rate Limiting Integration

```go
// Rate limiting with domain-aware error responses
func RateLimitMiddleware(domain string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := errorcustom.GetClientIP(r)
            requestID := errorcustom.GetRequestIDFromContext(r.Context())
            
            // Check rate limit for domain
            if rateLimiter.IsExceeded(clientIP, domain) {
                rateLimitErr := errorcustom.NewAPIError(
                    errorcustom.GetRateLimitCode(domain),
                    "Rate limit exceeded for this domain",
                    http.StatusTooManyRequests,
                ).WithDomain(domain).
                  WithDetail("client_ip", clientIP).
                  WithDetail("retry_after", rateLimiter.GetRetryAfter(clientIP, domain)).
                  WithRetryable(true)
                
                errorcustom.HandleError(w, rateLimitErr, requestID)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Input Sanitization and Validation

```go
// Advanced input sanitization with domain-specific rules
func SanitizeAndValidateInput(input map[string]interface{}, domain string) error {
    errorCollection := errorcustom.NewErrorCollection(domain)
    
    for field, value := range input {
        // Domain-specific sanitization rules
        sanitized, err := sanitizeFieldForDomain(field, value, domain)
        if err != nil {
            errorCollection.Add(errorcustom.NewValidationError(
                domain,
                field,
                "Input sanitization failed",
                value,
            ))
            continue
        }
        
        // Update with sanitized value
        input[field] = sanitized
        
        // Domain-specific validation rules
        if err := validateFieldForDomain(field, sanitized, domain); err != nil {
            errorCollection.Add(err)
        }
    }
    
    if errorCollection.HasErrors() {
        return errorCollection.ToAPIError()
    }
    
    return nil
}
```

## 📈 Production Deployment Patterns

### Health Check Integration

```go
// Health check with domain-aware error reporting
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    requestID := errorcustom.GetRequestIDFromContext(r.Context())
    
    healthStatus := make(map[string]interface{})
    hasErrors := false
    
    // Check each domain's health
    domains := []string{"user", "course", "payment", "content"}
    
    for _, domain := range domains {
        status, err := checkDomainHealth(domain)
        if err != nil {
            hasErrors = true
            
            // Create domain-specific health error
            healthErr := errorcustom.NewExternalServiceError(
                domain,
                "health_check",
                "status_check",
                "Domain health check failed",
                err,
                true,
            )
            
            healthStatus[domain] = map[string]interface{}{
                "status": "unhealthy",
                "error":  healthErr.ToAPIError().ToErrorResponse(),
            }
        } else {
            healthStatus[domain] = map[string]interface{}{
                "status": "healthy",
                "details": status,
            }
        }
    }
    
    // Return appropriate status
    if hasErrors {
        errorcustom.RespondWithJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
            "status":  "degraded",
            "domains": healthStatus,
        }, requestID)
        return
    }
    
    errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
        "status":  "healthy",
        "domains": healthStatus,
    }, requestID)
}
```

### Circuit Breaker Integration

```go
// Circuit breaker with domain-aware error handling
func CallExternalServiceWithCircuitBreaker(domain, service, operation string, fn func() error) error {
    circuitBreaker := getCircuitBreakerForDomain(domain, service)
    
    err := circuitBreaker.Call(fn)
    if err != nil {
        // Check if circuit breaker is open
        if circuitBreaker.IsOpen() {
            return errorcustom.NewExternalServiceError(
                domain,
                service,
                operation,
                "Service temporarily unavailable due to circuit breaker",
                err,
                true, // retryable when circuit closes
            )
        }
        
        // Regular external service error
        return errorcustom.NewExternalServiceError(
            domain,
            service,
            operation,
            "External service call failed",
            err,
            false,
        )
    }
    
    return nil
}
```

## 🔧 Development Tools and Utilities

### Error Code Generator

```go
// Generate all possible error codes for a domain (useful for documentation)
func GenerateErrorCodesForDomain(domain string) map[string]string {
    return errorcustom.GetDomainSpecificCodes(domain)
}

// Example output for "user" domain:
// {
//   "NOT_FOUND": "user_NOT_FOUND",
//   "VALIDATION_ERROR": "user_VALIDATION_ERROR", 
//   "DUPLICATE": "user_DUPLICATE",
//   "AUTHENTICATION_ERROR": "user_AUTHENTICATION_ERROR",
//   "AUTHORIZATION_ERROR": "user_AUTHORIZATION_ERROR",
//   "BUSINESS_LOGIC_ERROR": "user_BUSINESS_LOGIC_ERROR",
//   "EXTERNAL_SERVICE_ERROR": "user_EXTERNAL_SERVICE_ERROR",
//   "SYSTEM_ERROR": "user_SYSTEM_ERROR"
// }
```

### Error Documentation Generator

```go
// Generate error documentation for API docs
func GenerateErrorDocumentation() map[string]DomainErrorDocs {
    docs := make(map[string]DomainErrorDocs)
    
    domains := []string{"user", "course", "payment", "auth", "admin", "content", "system"}
    
    for _, domain := range domains {
        docs[domain] = DomainErrorDocs{
            Domain:      domain,
            ErrorCodes:  errorcustom.GetDomainSpecificCodes(domain),
            Examples:    generateErrorExamplesForDomain(domain),
            Description: getDomainDescription(domain),
        }
    }
    
    return docs
}
```

## 📚 Best Practices and Guidelines

### Error Handling Hierarchy

1. **Configuration First**: Use configuration to drive error handling behavior
2. **Domain Context**: Always provide domain context for better error tracking
3. **Rich Details**: Include relevant business context in error details
4. **Security Conscious**: Sanitize sensitive data in error responses
5. **Performance Aware**: Use error collections for bulk operations
6. **Monitoring Ready**: Include structured logging and metrics integration

### Code Organization Patterns

```go
// Service layer - business logic errors
func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    // Input validation
    if err := s.validateCreateUserRequest(req); err != nil {
        return nil, err // Already domain-aware
    }
    
    // Business logic
    if exists, err := s.userRepo.ExistsByEmail(req.Email); err != nil {
        return nil, errorcustom.NewDatabaseError("select", "users", err)
    } else if exists {
        return nil, errorcustom.NewDuplicateEmailError(req.Email)
    }
    
    // Create user
    user, err := s.userRepo.Create(req)
    if err != nil {
        return nil, errorcustom.NewDatabaseError("insert", "users", err)
    }
    
    return user, nil
}

// Repository layer - data access errors
func (r *UserRepository) FindByID(id int64) (*User, error) {
    var user User
    
    err := r.db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&user)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errorcustom.NewUserNotFoundByID(id)
        }
        return nil, errorcustom.NewDatabaseError("select", "users", err)
    }
    
    return &user, nil
}

// Handler layer - HTTP context and response handling
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    requestID := errorcustom.GetRequestIDFromContext(r.Context())
    domain := errorcustom.GetDomainFromContext(r.Context())
    
    // Parse parameters safely
    userID, err := errorcustom.ParseIDParamWithDomain(r, "id", domain)
    if err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    // Service call
    user, err := userService.GetByID(userID)
    if err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    errorcustom.RespondWithDomainSuccess(w, user, domain, requestID)
}
```

## 🔍 Advanced Debugging and Troubleshooting

### Error Tracing

```go
// Add tracing information to errors
func TraceError(err error, component, operation string) error {
    if err == nil {
        return nil
    }
    
    apiErr := errorcustom.ConvertToAPIError(err)
    if apiErr == nil {
        return err
    }
    
    // Add trace information
    apiErr.WithLayer(component).WithOperation(operation)
    
    // Add timing information
    apiErr.WithDetail("traced_at", time.Now().UTC())
    apiErr.WithDetail("trace_id", generateTraceID())
    
    return apiErr
}

// Error correlation across services
func CorrelateErrors(errors []error, correlationID string) error {
    if len(errors) == 0 {
        return nil
    }
    
    if len(errors) == 1 {
        apiErr := errorcustom.ConvertToAPIError(errors[0])
        return apiErr.WithDetail("correlation_id", correlationID)
    }
    
    // Multiple service errors
    collection := errorcustom.NewErrorCollection("system")
    for _, err := range errors {
        collection.Add(err)
    }
    
    apiErr := collection.ToAPIError()
    return apiErr.WithDetail("correlation_id", correlationID)
}
```

### Error Analytics

```go
// Error pattern analysis
type ErrorAnalytics struct {
    Domain       string    `json:"domain"`
    ErrorType    string    `json:"error_type"`
    Count        int       `json:"count"`
    LastOccurred time.Time `json:"last_occurred"`
    Trend        string    `json:"trend"` // increasing, decreasing, stable
}

func AnalyzeErrorPatterns(timeWindow time.Duration) map[string]ErrorAnalytics {
    // Analyze error patterns by domain
    patterns := make(map[string]ErrorAnalytics)
    
    // Query error logs and generate analytics
    errorLogs := queryErrorLogs(timeWindow)
    
    for _, log := range errorLogs {
        key := fmt.Sprintf("%s_%s", log.Domain, log.ErrorType)
        
        if analytics, exists := patterns[key]; exists {
            analytics.Count++
            if log.Timestamp.After(analytics.LastOccurred) {
                analytics.LastOccurred = log.Timestamp
            }
        } else {
            patterns[key] = ErrorAnalytics{
                Domain:       log.Domain,
                ErrorType:    log.ErrorType,
                Count:        1,
                LastOccurred: log.Timestamp,
                Trend:        calculateTrend(log),
            }
        }
    }
    
    return patterns
}
```

## 📋 Implementation Checklist

### Phase 1: Core Setup
- [ ] Implement core error types and interfaces
- [ ] Set up domain-aware error code system
- [ ] Create basic error constructors
- [ ] Implement error detection utilities
- [ ] Set up basic HTTP error handling

### Phase 2: Configuration Integration
- [ ] Implement Viper-based configuration management
- [ ] Add domain-specific configuration structures
- [ ] Create environment-specific defaults
- [ ] Implement configuration-driven error handler
- [ ] Add runtime configuration validation

### Phase 3: Advanced Features
- [ ] Implement comprehensive middleware stack
- [ ] Add error collections and bulk error handling
- [ ] Create advanced request parsing utilities
- [ ] Implement gRPC error integration
- [ ] Add security features (sanitization, rate limiting)

### Phase 4: Production Features
- [ ] Implement structured logging integration
- [ ] Add error metrics and monitoring
- [ ] Create health check integration
- [ ] Implement circuit breaker support
- [ ] Add error analytics and reporting

### Phase 5: Testing and Documentation
- [ ] Write comprehensive unit tests
- [ ] Create integration tests
- [ ] Generate API documentation
- [ ] Create migration guides
- [ ] Performance testing and optimization

## 🚦 Error Severity and Response Guidelines

### Error Severity Levels

```go
// ERROR (5xx) - Critical system issues
- Database connection failures
- System panics and crashes
- Critical external service failures
- Memory/resource exhaustion

// WARNING (5xx retryable) - Service degradation
- External service timeouts
- Circuit breaker trips
- Temporary resource unavailability
- Retryable external service errors

// INFO (4xx) - Client errors
- Validation failures
- Authentication errors
- Authorization denials
- Business logic violations

// DEBUG - Request tracing
- Successful request processing
- Performance metrics
- Request/response logging
```

### Response Time Guidelines

```go
// Configure timeouts per domain
domains:
  user:
    request_timeout: "5s"      # User operations should be fast
    max_login_attempts: 3
  course:
    request_timeout: "10s"     # Course operations can be slower
    enrollment_validation: true
  payment:
    request_timeout: "30s"     # Payment operations need more time
    provider_timeout: "25s"
    retry_attempts: 3
```

## 🔮 Future Enhancements and Roadmap

### Planned Features

1. **Error Correlation**: Cross-service error correlation and distributed tracing
2. **Machine Learning**: Error pattern prediction and anomaly detection
3. **Auto-Recovery**: Intelligent retry mechanisms with backoff strategies
4. **Error Aggregation**: Real-time error aggregation and alerting
5. **Custom Error Templates**: User-customizable error message templates
6. **Multi-Language Support**: Internationalized error messages
7. **Error Workflow**: Automated error escalation and resolution workflows

### Integration Opportunities

```go
// Prometheus metrics integration
func RecordErrorMetrics(apiErr *APIError) {
    errorCounter.WithLabelValues(
        apiErr.Domain,
        apiErr.Code,
        strconv.Itoa(apiErr.HTTPStatus),
    ).Inc()
    
    if apiErr.Retryable {
        retryableErrorCounter.WithLabelValues(apiErr.Domain).Inc()
    }
}

// OpenTelemetry tracing integration
func TraceError(ctx context.Context, err error) {
    span := trace.SpanFromContext(ctx)
    
    if apiErr, ok := err.(*errorcustom.APIError); ok {
        span.SetAttributes(
            attribute.String("error.domain", apiErr.Domain),
            attribute.String("error.code", apiErr.Code),
            attribute.Int("error.http_status", apiErr.HTTPStatus),
            attribute.Bool("error.retryable", apiErr.Retryable),
        )
        span.SetStatus(codes.Error, apiErr.Message)
    }
}
```

## 🎯 Domain-Specific Implementation Examples

### Complete User Domain Implementation

```go
// User service with comprehensive error handling
type UserService struct {
    config   *utils_config.Config
    repo     UserRepository
    cache    CacheService
    emailSvc EmailService
}

func (s *UserService) RegisterUser(req RegisterUserRequest) (*User, error) {
    requestID := req.RequestID
    domain := "user"
    
    // Configuration-driven validation
    errorCollection := errorcustom.NewErrorCollection(domain)
    
    // Email validation with domain context
    if err := errorcustom.ValidateEmailWithDomain(req.Email, domain, requestID); err != nil {
        errorCollection.Add(err)
    }
    
    // Password validation based on configuration
    if s.config.IsPasswordComplexityRequired() {
        if err := errorcustom.ValidatePasswordWithDomain(req.Password, domain, requestID); err != nil {
            errorCollection.Add(err)
        }
    }
    
    // Check for validation errors
    if errorCollection.HasErrors() {
        return nil, errorCollection.ToAPIError()
    }
    
    // Check for duplicate email
    exists, err := s.repo.ExistsByEmail(req.Email)
    if err != nil {
        return nil, errorcustom.NewDatabaseError("select", "users", err)
    }
    if exists {
        return nil, errorcustom.NewDuplicateEmailError(req.Email)
    }
    
    // Create user
    user, err := s.repo.Create(req)
    if err != nil {
        return nil, errorcustom.NewDatabaseError("insert", "users", err)
    }
    
    // Send verification email if required
    if s.config.IsEmailVerificationRequired() {
        if err := s.emailSvc.SendVerificationEmail(user.Email); err != nil {
            // Log but don't fail registration
            logger.Warning("Failed to send verification email", map[string]interface{}{
                "user_id":    user.ID,
                "email":      user.Email,
                "error":      err.Error(),
                "request_id": requestID,
                "domain":     domain,
            })
        }
    }
    
    return user, nil
}
```

## 🎪 Complete Example Application

```go
// main.go - Complete application setup
func main() {
    // Initialize configuration
    if err := utils_config.InitializeConfig("./config.yaml"); err != nil {
        log.Fatal("Failed to initialize configuration:", err)
    }
    
    config := utils_config.GetConfig()
    
    // Initialize services
    userService := NewUserService(config)
    courseService := NewCourseService(config)
    paymentService := NewPaymentService(config)
    
    // Setup router with complete middleware stack
    r := chi.NewRouter()
    
    // Global middleware
    r.Use(errorcustom.RequestIDMiddleware)
    r.Use(errorcustom.LogHTTPMiddleware)
    r.Use(errorcustom.RecoveryMiddleware)
    
    // Domain-specific routes with complete error handling
    setupUserRoutes(r, userService)
    setupCourseRoutes(r, courseService)
    setupPaymentRoutes(r, paymentService)
    
    // Health check endpoint
    r.Get("/health", HealthCheckHandler)
    
    // Start server with proper error handling
    server := &http.Server{
        Addr:         fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port),
        Handler:      r,
        ReadTimeout:  config.Server.ReadTimeout,
        WriteTimeout: config.Server.WriteTimeout,
        IdleTimeout:  config.Server.IdleTimeout,
    }
    
    log.Printf("Server starting on %s", server.Addr)
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal("Server failed to start:", err)
    }
}

func setupUserRoutes(r chi.Router, service *UserService) {
    r.Route("/api/users", func(r chi.Router) {
        r.Use(errorcustom.DomainMiddleware("user"))
        r.Use(RateLimitMiddleware("user"))
        
        r.Post("/", CreateUserHandler(service))
        r.Post("/login", LoginHandler(service))
        r.Get("/{id}", GetUserHandler(service))
        r.Put("/{id}", UpdateUserHandler(service))
        r.Delete("/{id}", DeleteUserHandler(service))
        
        // Bulk operations
        r.Post("/bulk", BulkCreateUsersHandler(service))
        r.Put("/bulk", BulkUpdateUsersHandler(service))
    })
}
```

This comprehensive system provides enterprise-grade error handling with extensive configurability, robust security features, and production-ready monitoring capabilities. The domain-aware architecture ensures scalable error management as your application grows across multiple business domains.
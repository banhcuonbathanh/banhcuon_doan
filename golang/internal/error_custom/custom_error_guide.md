# Go Error Handling System - Complete Guide

## Overview

This is a comprehensive, production-ready error handling system for Go web APIs that provides structured error types, consistent HTTP responses, intelligent logging, and robust error management across all application layers. The system integrates seamlessly with the logging system and provides excellent debugging and monitoring capabilities.

## Key Features

- **Structured Error Types**: Domain-specific errors with rich context
- **Layered Error Handling**: Track errors across handler → service → repository layers
- **Intelligent Logging**: Optimized logging levels to reduce noise
- **Consistent API Responses**: Standardized JSON error responses
- **Security-First**: Automatic sensitive data masking
- **Middleware Integration**: Request tracking, recovery, and logging
- **Validation Integration**: Comprehensive input validation with detailed messages
- **Authentication Errors**: Granular authentication failure tracking

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Error Handling System                     │
├─────────────────────────────────────────────────────────────┤
│  Layer 1: HTTP Middleware (Recovery, Logging, Request ID)   │
│  Layer 2: Handler Error Processing (APIError conversion)     │
│  Layer 3: Domain-Specific Errors (Auth, Validation, etc.)   │
│  Layer 4: Service & Repository Errors (Business logic)      │
│  Layer 5: Structured Responses (JSON, logging integration)  │
└─────────────────────────────────────────────────────────────┘
```

### Error Flow

```
Request → Middleware → Handler → Service → Repository
   ↓         ↓          ↓         ↓         ↓
Logging → Recovery → APIError → ServiceError → RepositoryError
   ↓         ↓          ↓         ↓         ↓
Response ← JSON ← StandardFormat ← Conversion ← ErrorCapture
```

## Quick Start Guide

### 1. Basic Error Handling in Handlers

```go
package handlers

import (
    "net/http"
    "your-app/utils"
    "your-app/internal/error_custom"
)

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    requestID := utils.GetRequestIDFromContext(r.Context())
    
    // Parse URL parameter
    userID, apiErr := utils.ParseIDParam(r, "id")
    if apiErr != nil {
        utils.HandleError(w, apiErr, requestID)
        return
    }
    
    // Call service
    user, err := h.userService.GetUser(userID)
    if err != nil {
        utils.HandleError(w, err, requestID)
        return
    }
    
    // Success response
    utils.RespondWithJSON(w, http.StatusOK, user, requestID)
}
```

### 2. Service Layer Error Creation

```go
package services

import (
    errorcustom "your-app/internal/error_custom"
)

func (s *UserService) GetUser(id int64) (*User, error) {
    user, err := s.repository.GetUser(id)
    if err != nil {
        // Convert repository error to service error
        return nil, errorcustom.NewServiceError(
            "UserService",
            "GetUser", 
            "Failed to retrieve user",
            err,
            false, // not retryable
        )
    }
    
    if user == nil {
        return nil, errorcustom.NewUserNotFoundByID(id)
    }
    
    return user, nil
}
```

### 3. Repository Layer Error Handling

```go
package repositories

import (
    errorcustom "your-app/internal/error_custom"
)

func (r *UserRepository) GetUser(id int64) (*User, error) {
    var user User
    err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // Let service layer handle not found
        }
        
        return nil, errorcustom.NewRepositoryError(
            "SELECT",
            "users",
            "Database query failed", 
            err,
        )
    }
    
    return &user, nil
}
```

## Domain-Specific Error Types

### 1. Authentication Errors

Provides detailed authentication failure tracking with specific steps and reasons:

```go
// Email not found during login
func (s *AuthService) Login(email, password string) (*Token, error) {
    user, err := s.userRepo.GetByEmail(email)
    if err != nil {
        return nil, err
    }
    
    if user == nil {
        // Specific error for email not found
        return nil, errorcustom.NewEmailNotFoundError(email)
    }
    
    if !s.passwordService.Verify(password, user.PasswordHash) {
        // Specific error for password mismatch
        return nil, errorcustom.NewPasswordMismatchError(email)
    }
    
    if !user.Active {
        return nil, errorcustom.NewAccountDisabledError(email)
    }
    
    // Generate token...
    return token, nil
}

// Usage in handler
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    requestID := utils.GetRequestIDFromContext(r.Context())
    
    var req LoginRequest
    if err := utils.DecodeJSON(r.Body, &req, requestID, false); err != nil {
        utils.HandleError(w, err, requestID)
        return
    }
    
    token, err := h.authService.Login(req.Email, req.Password)
    if err != nil {
        // Authentication errors are automatically logged with appropriate context
        utils.HandleError(w, err, requestID)
        return
    }
    
    utils.RespondWithJSON(w, http.StatusOK, token, requestID)
}
```

**Output Examples:**
```json
// Email not found
{
  "code": "AUTHENTICATION_ERROR",
  "message": "Invalid credentials",
  "details": {
    "email": "user@example.com",
    "step": "email_check",
    "user_found": false
  }
}

// Password mismatch  
{
  "code": "AUTHENTICATION_ERROR", 
  "message": "Invalid credentials",
  "details": {
    "email": "user@example.com",
    "step": "password_check", 
    "user_found": true
  }
}
```

### 2. Validation Errors

Comprehensive input validation with field-specific error messages:

```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    requestID := utils.GetRequestIDFromContext(r.Context())
    
    var req CreateUserRequest
    if err := utils.DecodeJSON(r.Body, &req, requestID, false); err != nil {
        utils.HandleError(w, err, requestID)
        return
    }
    
    // Validate request
    if err := h.validator.Struct(req); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            utils.HandleValidationErrors(w, validationErrors, requestID)
            return
        }
        utils.HandleError(w, err, requestID)
        return
    }
    
    // Additional business logic validation
    if err := utils.ValidatePasswordWithDetails(req.Password, requestID); err != nil {
        utils.HandleError(w, err, requestID)
        return
    }
    
    // Process request...
}

// Custom validation tags
type CreateUserRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8,max=100"`
    FirstName string `json:"first_name" validate:"required,min=2,max=50"`
    LastName  string `json:"last_name" validate:"required,min=2,max=50"`
    Role      string `json:"role" validate:"required,oneof=admin user manager"`
}
```

**Output Example:**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "details": {
    "validation_errors": {
      "email": "Invalid email format",
      "password": "Password is too short",
      "first_name": "This field is required"
    }
  }
}
```

### 3. Service Layer Errors

Track errors with service context and retry information:

```go
func (s *PaymentService) ProcessPayment(payment Payment) error {
    // Call external payment API
    err := s.externalAPI.Charge(payment.Amount)
    if err != nil {
        // Determine if error is retryable
        retryable := s.isRetryableError(err)
        
        return errorcustom.NewServiceError(
            "PaymentService",
            "ProcessPayment",
            "Payment processing failed",
            err,
            retryable,
        )
    }
    
    return nil
}

func (s *PaymentService) isRetryableError(err error) bool {
    errMsg := strings.ToLower(err.Error())
    return strings.Contains(errMsg, "timeout") ||
           strings.Contains(errMsg, "connection") ||
           strings.Contains(errMsg, "temporary")
}
```

### 4. Repository/Database Errors

Track database operations with table and operation context:

```go
func (r *UserRepository) CreateUser(user *User) error {
    _, err := r.db.Exec(`
        INSERT INTO users (email, password_hash, first_name, last_name, role)
        VALUES ($1, $2, $3, $4, $5)
    `, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role)
    
    if err != nil {
        // Check for specific database errors
        if isDuplicateKeyError(err) {
            return errorcustom.NewDuplicateEmailError(user.Email)
        }
        
        return errorcustom.NewRepositoryError(
            "INSERT",
            "users", 
            "Failed to create user",
            err,
        )
    }
    
    return nil
}

func isDuplicateKeyError(err error) bool {
    return strings.Contains(err.Error(), "duplicate key") ||
           strings.Contains(err.Error(), "unique constraint")
}
```

## Advanced Usage Patterns

### 1. Layered Error Context

Track errors as they flow through application layers:

```go
// Handler Layer
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    requestID := utils.GetRequestIDFromContext(r.Context())
    userID := utils.GetUserIDFromContext(r)
    
    var req CreateOrderRequest
    if err := utils.DecodeJSON(r.Body, &req, requestID, false); err != nil {
        utils.HandleError(w, err, requestID)
        return
    }
    
    order, err := h.orderService.CreateOrder(userID, req)
    if err != nil {
        // Add handler context to existing error
        if apiErr, ok := err.(*errorcustom.APIError); ok {
            apiErr.WithLayer("handler").WithOperation("create_order")
        }
        utils.HandleError(w, err, requestID)
        return
    }
    
    utils.RespondWithJSON(w, http.StatusCreated, order, requestID)
}

// Service Layer  
func (s *OrderService) CreateOrder(userID int64, req CreateOrderRequest) (*Order, error) {
    // Validate inventory
    if !s.inventoryService.HasStock(req.Items) {
        return nil, errorcustom.NewServiceError(
            "OrderService",
            "CreateOrder", 
            "Insufficient inventory",
            nil,
            false,
        ).WithLayer("service").WithDetail("user_id", userID)
    }
    
    // Create order in database
    order, err := s.orderRepo.Create(userID, req.Items)
    if err != nil {
        // Pass through repository error with service context
        if apiErr, ok := err.(*errorcustom.APIError); ok {
            apiErr.WithLayer("service").WithOperation("create_order")
        }
        return nil, err
    }
    
    return order, nil
}

// Repository Layer
func (r *OrderRepository) Create(userID int64, items []OrderItem) (*Order, error) {
    tx, err := r.db.Begin()
    if err != nil {
        return nil, errorcustom.NewRepositoryError(
            "BEGIN_TRANSACTION",
            "orders",
            "Failed to start transaction",
            err,
        ).WithLayer("repository")
    }
    defer tx.Rollback()
    
    // Insert order and items...
    
    return order, nil
}
```

### 2. Middleware Integration

Complete request lifecycle management with error handling:

```go
func SetupMiddleware(r chi.Router) {
    // Request ID middleware (first)
    r.Use(utils.RequestIDMiddleware)
    
    // Recovery middleware (catch panics) 
    r.Use(utils.RecoveryMiddleware)
    
    // HTTP logging middleware
    r.Use(utils.LogHTTPMiddleware)
    
    // Authentication middleware
    r.Use(auth.JWTMiddleware)
}

// Custom authentication middleware with error handling
func JWTMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := utils.GetRequestIDFromContext(r.Context())
        
        token := extractToken(r)
        if token == "" {
            err := errorcustom.NewInvalidTokenError("JWT", "missing token")
            utils.HandleError(w, err, requestID)
            return
        }
        
        claims, err := validateToken(token)
        if err != nil {
            authErr := errorcustom.NewInvalidTokenError("JWT", "invalid or expired")
            utils.HandleError(w, authErr, requestID)
            return
        }
        
        // Add user context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "user_email", claims.Email)
        r = r.WithContext(ctx)
        
        next.ServeHTTP(w, r)
    })
}
```

### 3. Error Recovery and Retry Logic

Handle transient failures with intelligent retry:

```go
func (s *EmailService) SendEmail(to, subject, body string) error {
    const maxRetries = 3
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err := s.sendEmailAttempt(to, subject, body)
        if err == nil {
            return nil
        }
        
        // Check if error is retryable
        if !errorcustom.IsRetryableError(err) {
            return err
        }
        
        if attempt < maxRetries {
            // Exponential backoff
            backoff := time.Duration(attempt*attempt) * time.Second
            time.Sleep(backoff)
            continue
        }
        
        // Final attempt failed
        return errorcustom.NewServiceError(
            "EmailService",
            "SendEmail",
            "Failed to send email after retries",
            err,
            false, // no more retries
        ).WithDetail("attempts", maxRetries).WithDetail("recipient", to)
    }
    
    return nil
}

func (s *EmailService) sendEmailAttempt(to, subject, body string) error {
    err := s.smtpClient.Send(to, subject, body)
    if err != nil {
        return errorcustom.NewServiceError(
            "SMTPClient",
            "Send",
            "SMTP send failed",
            err,
            s.isTemporaryError(err),
        )
    }
    return nil
}
```

## HTTP Middleware Components

### 1. Request ID Middleware

Tracks requests across the entire lifecycle:

```go
// Automatically generates unique request IDs
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := generateRequestID() // Format: req_1642685443123_456
        
        // Add to response headers for client debugging
        w.Header().Set("X-Request-ID", requestID)
        
        // Add to request context
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        r = r.WithContext(ctx)
        
        next.ServeHTTP(w, r)
    })
}

// Usage in handlers
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    requestID := utils.GetRequestIDFromContext(r.Context())
    // requestID is now available for all logging and error handling
}
```

### 2. Recovery Middleware

Handles panics gracefully with proper error responses:

```go
// Catches panics and converts them to proper error responses
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                requestID := utils.GetRequestIDFromContext(r.Context())
                
                // Log panic with full context
                logger.Error("Panic recovered", map[string]interface{}{
                    "error":      err,
                    "method":     r.Method,
                    "path":       r.URL.Path,
                    "ip":         utils.GetClientIP(r),
                    "request_id": requestID,
                    "stack_trace": string(debug.Stack()), // Include in debug mode
                })
                
                // Return proper API error response
                apiErr := errorcustom.NewAPIError(
                    errorcustom.ErrCodeInternalError,
                    "Internal server error",
                    http.StatusInternalServerError,
                )
                utils.HandleError(w, apiErr, requestID)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### 3. HTTP Logging Middleware

Optimized request/response logging:

```go
// Logs all HTTP requests and responses with performance data
func LogHTTPMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := utils.GetRequestIDFromContext(r.Context())
        clientIP := utils.GetClientIP(r)
        
        // Wrap response writer to capture status code
        recorder := &utils.ResponseRecorder{ResponseWriter: w}
        
        // Process request
        next.ServeHTTP(recorder, r)
        
        duration := time.Since(start)
        
        // Use the logger's LogAPIRequest method for consistent formatting
        logger.LogAPIRequest(
            r.Method,
            r.URL.Path,
            recorder.StatusCode,
            duration,
            map[string]interface{}{
                "request_id": requestID,
                "ip":         clientIP,
                "user_agent": r.Header.Get("User-Agent"),
                "user_id":    utils.GetUserIDFromContext(r),
            },
        )
    })
}
```

## Input Validation & Security

### 1. Comprehensive Parameter Validation

```go
// Safe ID parameter parsing with validation
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    requestID := utils.GetRequestIDFromContext(r.Context())
    
    // Parse and validate ID parameter
    userID, apiErr := utils.ParseIDParam(r, "id")
    if apiErr != nil {
        utils.HandleError(w, apiErr, requestID)
        return
    }
    
    // Parse and validate string parameters
    section, apiErr := utils.GetStringParam(r, "section", 2)
    if apiErr != nil {
        utils.HandleError(w, apiErr, requestID)
        return
    }
    
    // Parse and validate pagination
    limit, offset, apiErr := utils.GetPaginationParams(r)
    if apiErr != nil {
        utils.HandleError(w, apiErr, requestID)
        return
    }
    
    // Parse and validate sorting
    sortBy, sortOrder, apiErr := utils.GetSortParams(r, []string{"name", "email", "created_at"})
    if apiErr != nil {
        utils.HandleError(w, apiErr, requestID)
        return
    }
    
    // All parameters are now validated and safe to use
}
```

### 2. Password Validation with Security Requirements

```go
func ValidatePassword(password string) error {
    // Length requirements
    if len(password) < 8 {
        return errorcustom.NewAPIError(
            errorcustom.ErrCodeWeakPassword,
            "Password must be at least 8 characters",
            http.StatusBadRequest,
        )
    }
    
    if len(password) > 128 {
        return errorcustom.NewAPIError(
            errorcustom.ErrCodeWeakPassword,
            "Password cannot exceed 128 characters", 
            http.StatusBadRequest,
        )
    }
    
    // Character requirements
    var hasUpper, hasLower, hasDigit, hasSpecial bool
    
    for _, c := range password {
        switch {
        case c >= 'A' && c <= 'Z':
            hasUpper = true
        case c >= 'a' && c <= 'z':
            hasLower = true
        case c >= '0' && c <= '9':
            hasDigit = true
        case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?/", c):
            hasSpecial = true
        }
    }
    
    var missing []string
    if !hasUpper { missing = append(missing, "uppercase letter") }
    if !hasLower { missing = append(missing, "lowercase letter") }
    if !hasDigit { missing = append(missing, "digit") }
    if !hasSpecial { missing = append(missing, "special character") }
    
    if len(missing) > 0 {
        return errorcustom.NewAPIError(
            errorcustom.ErrCodeWeakPassword,
            "Password does not meet security requirements",
            http.StatusBadRequest,
        ).WithDetail("missing_requirements", missing)
    }
    
    return nil
}
```

### 3. JSON Decoding with Security

```go
// Safe JSON decoding with size limits and validation
func DecodeJSON(body io.Reader, target interface{}, requestID string, logRawBody bool) error {
    // Limit request body size to prevent DoS attacks
    limitedReader := io.LimitReader(body, 1<<20) // 1MB limit
    
    bodyBytes, err := io.ReadAll(limitedReader)
    if err != nil {
        return errorcustom.NewAPIError(
            errorcustom.ErrCodeInvalidInput,
            "Failed to read request body",
            http.StatusBadRequest,
        ).WithDetail("error", err.Error())
    }
    
    // Optional raw body logging for debugging (disabled in production)
    if logRawBody && len(bodyBytes) > 0 {
        logger.Debug("Raw request body", map[string]interface{}{
            "request_id": requestID,
            "body_size":  len(bodyBytes),
            "body":       string(bodyBytes),
        })
    }
    
    // Decode JSON with security checks
    if err := json.Unmarshal(bodyBytes, target); err != nil {
        return errorcustom.NewAPIError(
            errorcustom.ErrCodeInvalidInput,
            "Invalid JSON format",
            http.StatusBadRequest,
        ).WithDetail("error", err.Error())
    }
    
    return nil
}
```

## Logging Integration & Optimization

The error handling system is designed to work seamlessly with the logging system while avoiding log noise:

### 1. Intelligent Log Level Selection

```go
func HandleError(w http.ResponseWriter, err error, requestID string) {
    var apiErr *errorcustom.APIError
    
    // Convert various error types to APIError
    switch e := err.(type) {
    case *errorcustom.APIError:
        apiErr = e
    case *errorcustom.AuthenticationError:
        apiErr = e.ToAPIError()
    // ... other conversions
    }
    
    // Intelligent logging based on error severity
    switch {
    case apiErr.HTTPStatus >= 500:
        // Server errors: Always log as ERROR
        logContext := apiErr.GetLogContext()
        logContext["request_id"] = requestID
        logger.Error("Server error occurred", logContext)
        
    case apiErr.HTTPStatus == 401:
        // Authentication failures: Log as WARNING for security monitoring
        logger.LogAuthAttempt(
            getEmailFromError(apiErr),
            false,
            apiErr.Message,
            map[string]interface{}{"request_id": requestID},
        )
        
    case apiErr.HTTPStatus == 400:
        // Validation errors: Logged in LogAPIRequest middleware as INFO/WARNING
        // No additional logging here to avoid noise
        
    default:
        // Other client errors: Minimal logging
        logger.Debug("Client error", map[string]interface{}{
            "error_code":  apiErr.Code,
            "http_status": apiErr.HTTPStatus,
            "request_id":  requestID,
        })
    }
    
    // Send response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(apiErr.HTTPStatus)
    json.NewEncoder(w).Encode(apiErr.ToErrorResponse())
}
```

### 2. Context-Rich Error Logging

```go
// APIError provides rich context for logging
func (e *APIError) GetLogContext() map[string]interface{} {
    context := map[string]interface{}{
        "error_code":    e.Code,
        "error_message": e.Message,
        "http_status":   e.HTTPStatus,
    }
    
    if e.Layer != "" {
        context["layer"] = e.Layer
    }
    if e.Operation != "" {
        context["operation"] = e.Operation
    }
    if e.Cause != nil {
        context["cause"] = e.Cause.Error()
        context["cause_type"] = fmt.Sprintf("%T", e.Cause)
    }
    if len(e.Details) > 0 {
        context["details"] = e.Details
    }
    
    return context
}

// Usage example in service layer
func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    user, err := s.repository.Create(req)
    if err != nil {
        // Create service error with rich context
        serviceErr := errorcustom.NewServiceError(
            "UserService",
            "CreateUser",
            "Failed to create user",
            err,
            false,
        ).WithDetail("email", req.Email).WithDetail("role", req.Role)
        
        // Error will be logged with all this context when handled
        return nil, serviceErr
    }
    
    return user, nil
}
```

## Production Best Practices

### 1. Security Considerations

```go
// Never log sensitive information
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := utils.DecodeJSON(r.Body, &req, requestID, false); err != nil {
        utils.HandleError(w, err, requestID)
        return
    }
    
    // Validate credentials
    token, err := h.authService.Login(req.Email, req.Password)
    if err != nil {
        // Log authentication attempt without password
        logger.LogAuthAttempt(
            req.Email,
            false,
            "invalid_credentials", 
            map[string]interface{}{
                "ip": utils.GetClientIP(r),
                "user_agent": r.Header.Get("User-Agent"),
                "request_id": requestID,
                // Never log: password, tokens, secrets
            },
        )
        
        utils.HandleError(w, err, requestID)
        return
    }
    
    // Success logging
    logger.LogAuthAttempt(req.Email, true, "login_successful")
    utils.RespondWithJSON(w, http.StatusOK, token, requestID)
}
```

### 2. Error Rate Monitoring

```go
// Track error rates for monitoring and alerting
func (h *BaseHandler) TrackErrorRate(errorCode string, httpStatus int) {
    logger.LogMetric("api_errors", 1, "count", map[string]interface{}{
        "error_code":  errorCode,
        "http_status": httpStatus,
        "component":   "handler",
    })
    
    // Alert on high error rates
    if httpStatus >= 500 {
        logger.LogSecurityEvent(
            "high_error_rate",
            "Server error rate threshold exceeded",
            "medium",
            map[string]interface{}{
                "error_code":  errorCode,
                "http_status": httpStatus,
            },
        )
    }
}
```

### 3. Performance Impact Management

```go
// Minimize performance impact of error handling
func HandleError(w http.ResponseWriter, err error, requestID string) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 10*time.Millisecond {
            logger.Warning("Slow error handling", map[string]interface{}{
                "duration_ms": duration.Milliseconds(),
                "request_id":  requestID,
            })
        }
    }()
    
    // Fast path for common errors
    if apiErr, ok := err.(*errorcustom.APIError); ok {
        respondWithAPIError(w, apiErr, requestID)
        return
    }
    
    // Slower path for error conversion
    convertedErr := convertToAPIError(err)
    respondWithAPIError(w, convertedErr, requestID)
}
```

## Integration Examples

### 1. Database Integration (GORM)

```go
func (r *UserRepository) Create(user *User) error {
    if err := r.db.Create(user).Error; err != nil {
        // Handle GORM-specific errors
        switch {
        case errors.Is(err, gorm.ErrDuplicatedKey):
            return errorcustom.NewDuplicateEmailError(user.Email)
        case errors.Is(err, gorm.ErrInvalidTransaction):
            return errorcustom.NewRepositoryError(
                "CREATE",
                "users",
                "Transaction error",
                err,
            )
        default:
            return errorcustom.NewRepositoryError(
                "CREATE", 
                "users",
                "Database operation failed",
                err,
            )
        }
    }
    return nil
}
```

### 2. HTTP Client Integration

```go
func (c *ExternalAPIClient) CallService(endpoint string, data interface{}) error {
    resp, err := c.httpClient.Post(endpoint, "application/json", data)
    if err != nil {
        return errorcustom.NewServiceError(
            "ExternalAPI",
            "CallService",
            "HTTP request failed",
            err,
            true, // retryable
        )
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return errorcustom.NewServiceError(
            "ExternalAPI",
            "CallService",
            fmt.Sprintf("API returned %d", resp.StatusCode),
            fmt.Errorf("response: %s", string(body)),
            resp.StatusCode >= 500, // 5xx errors are retryable
        ).WithDetail("status_code", resp.StatusCode).
          WithDetail("endpoint", endpoint)
    }
    
    return nil
}
```

### 3. gRPC Integration

```go
func (c *UserGRPCClient) GetUser(ctx context.Context, userID int64) (*User, error) {
    resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{
        UserId: userID,
    })
    
    if err != nil {
        // Parse gRPC status codes
        if st, ok := status.FromError(err); ok {
            switch st.Code() {
            case codes.NotFound:
                return nil, errorcustom.NewUserNotFoundByID(userID)
            case codes.Unauthenticated:
                return nil, errorcustom.NewInvalidTokenError("gRPC", "unauthenticated")
            case codes.Unavailable:
                return nil, errorcustom.NewServiceError(
                    "UserGRPC",
                    "GetUser",
                    "Service unavailable",
                    err,
                    true, // retryable
                )
            default:
                return nil, errorcustom.NewServiceError(
                    "UserGRPC",
                    "GetUser", 
                    "gRPC call failed",
                    err,
                    false,
                )
            }
        }
        
        return nil, errorcustom.NewServiceError(
            "UserGRPC",
            "GetUser",
            "Unknown gRPC error",
            err,
            false,
        )
    }
    
    return convertFromProto(resp.User), nil
}
```

## Testing Error Handling

### 1. Unit Testing Error Scenarios

```go
func TestUserHandler_GetUser_NotFound(t *testing.T) {
    // Setup
    mockService := &MockUserService{}
    handler := NewUserHandler(mockService)
    
    // Mock service to return user not found error
    mockService.On("GetUser", int64(123)).Return(
        nil, 
        errorcustom.NewUserNotFoundByID(123),
    )
    
    // Create test request
    req := httptest.NewRequest("GET", "/users/123", nil)
    req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-123"))
    rr := httptest.NewRecorder()
    
    // Execute
    handler.GetUser(rr, req)
    
    // Assert
    assert.Equal(t, http.StatusNotFound, rr.Code)
    
    var response errorcustom.ErrorResponse
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, errorcustom.ErrCodeUserNotFound, response.Code)
    assert.Equal(t, "user with ID 123 not found", response.Message)
}
```

### 2. Integration Testing Error Flows

```go
func TestUserAPI_CreateUser_ValidationError(t *testing.T) {
    // Setup test server
    server := setupTestServer()
    defer server.Close()
    
    // Invalid request payload
    payload := map[string]interface{}{
        "email":      "invalid-email",
        "password":   "weak",
        "first_name": "",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    server.Handler.ServeHTTP(rr, req)
    
    // Assert validation errors
    assert.Equal(t, http.StatusBadRequest, rr.Code)
    
    var response errorcustom.ErrorResponse
    json.Unmarshal(rr.Body.Bytes(), &response)
    assert.Equal(t, errorcustom.ErrCodeValidationError, response.Code)
    
    // Check validation details
    errors := response.Details["validation_errors"].(map[string]interface{})
    assert.Contains(t, errors, "email")
    assert.Contains(t, errors, "password") 
    assert.Contains(t, errors, "first_name")
}
```

### 3. Error Handler Testing

```go
func TestHandleError_APIError(t *testing.T) {
    rr := httptest.NewRecorder()
    
    apiErr := errorcustom.NewAPIError(
        errorcustom.ErrCodeUserNotFound,
        "User not found",
        http.StatusNotFound,
    ).WithDetail("user_id", 123)
    
    utils.HandleError(rr, apiErr, "test-request-id")
    
    assert.Equal(t, http.StatusNotFound, rr.Code)
    assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
    
    var response errorcustom.ErrorResponse
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, errorcustom.ErrCodeUserNotFound, response.Code)
    assert.Equal(t, "User not found", response.Message)
    assert.Equal(t, float64(123), response.Details["user_id"])
}
```

## Monitoring and Observability

### 1. Error Metrics Collection

```go
// Collect error metrics for monitoring dashboards
func (h *BaseHandler) collectErrorMetrics(apiErr *errorcustom.APIError, requestID string) {
    // Basic error metrics
    logger.LogMetric("http_errors_total", 1, "count", map[string]interface{}{
        "error_code":  apiErr.Code,
        "http_status": apiErr.HTTPStatus,
        "layer":       apiErr.Layer,
        "operation":   apiErr.Operation,
    })
    
    // Response time metrics for errors
    if duration := getRequestDuration(requestID); duration > 0 {
        logger.LogMetric("error_response_time", duration.Milliseconds(), "ms", map[string]interface{}{
            "error_code":  apiErr.Code,
            "http_status": apiErr.HTTPStatus,
        })
    }
    
    // Error rate by endpoint
    if endpoint := getEndpointFromContext(requestID); endpoint != "" {
        logger.LogMetric("endpoint_errors", 1, "count", map[string]interface{}{
            "endpoint":   endpoint,
            "error_code": apiErr.Code,
        })
    }
}
```

### 2. Alerting Integration

```go
// Set up alerts for critical error patterns
func (a *AlertManager) checkErrorPatterns(apiErr *errorcustom.APIError, requestID string) {
    switch {
    case apiErr.HTTPStatus >= 500:
        a.triggerServerErrorAlert(apiErr, requestID)
    case apiErr.Code == errorcustom.ErrCodeAuthFailed:
        a.trackAuthFailures(apiErr, requestID)
    case isRateLimitError(apiErr):
        a.checkRateLimitThreshold(apiErr, requestID)
    }
}

func (a *AlertManager) triggerServerErrorAlert(apiErr *errorcustom.APIError, requestID string) {
    logger.LogSecurityEvent(
        "server_error",
        "Server error occurred",
        "high",
        map[string]interface{}{
            "error_code":  apiErr.Code,
            "request_id":  requestID,
            "layer":       apiErr.Layer,
            "operation":   apiErr.Operation,
        },
    )
    
    // Trigger external alerting system
    a.sendAlert("SERVER_ERROR", map[string]interface{}{
        "severity":   "high",
        "error_code": apiErr.Code,
        "message":    apiErr.Message,
        "request_id": requestID,
    })
}
```

### 3. Error Trend Analysis

```go
// Analyze error trends for operational insights
func (a *Analytics) analyzeErrorTrends(timeWindow time.Duration) {
    errors := a.getErrorsInWindow(timeWindow)
    
    // Group by error code
    errorCounts := make(map[string]int)
    for _, err := range errors {
        errorCounts[err.Code]++
    }
    
    // Identify top errors
    for code, count := range errorCounts {
        if count > a.getThreshold(code) {
            logger.LogMetric("error_trend_alert", count, "count", map[string]interface{}{
                "error_code":  code,
                "time_window": timeWindow.String(),
                "threshold":   a.getThreshold(code),
            })
        }
    }
    
    // Analyze error distribution by layer
    layerErrors := a.groupErrorsByLayer(errors)
    for layer, count := range layerErrors {
        logger.LogMetric("layer_errors", count, "count", map[string]interface{}{
            "layer":       layer,
            "time_window": timeWindow.String(),
        })
    }
}
```

## Configuration and Environment Setup

### 1. Environment-based Error Handling

```go
// Configure error handling based on environment
func setupErrorHandling() {
    env := os.Getenv("APP_ENV")
    
    switch env {
    case "development":
        // Verbose error messages for debugging
        errorcustom.SetVerboseMode(true)
        errorcustom.SetIncludeStackTrace(true)
        errorcustom.SetLogRawRequests(true)
        
    case "staging":
        // Moderate verbosity for testing
        errorcustom.SetVerboseMode(true)
        errorcustom.SetIncludeStackTrace(false)
        errorcustom.SetLogRawRequests(false)
        
    case "production":
        // Minimal error exposure for security
        errorcustom.SetVerboseMode(false)
        errorcustom.SetIncludeStackTrace(false)
        errorcustom.SetLogRawRequests(false)
        errorcustom.SetSanitizeErrors(true)
    }
}
```

### 2. Error Handler Configuration

```go
// Configuration structure for error handling
type ErrorConfig struct {
    VerboseMode       bool   `json:"verbose_mode"`
    IncludeStackTrace bool   `json:"include_stack_trace"`
    LogRawRequests    bool   `json:"log_raw_requests"`
    SanitizeErrors    bool   `json:"sanitize_errors"`
    MaxErrorDetails   int    `json:"max_error_details"`
    AlertingEnabled   bool   `json:"alerting_enabled"`
}

// Load configuration from file or environment
func LoadErrorConfig() *ErrorConfig {
    config := &ErrorConfig{
        VerboseMode:       getEnvBool("ERROR_VERBOSE_MODE", false),
        IncludeStackTrace: getEnvBool("ERROR_INCLUDE_STACK", false),
        LogRawRequests:    getEnvBool("ERROR_LOG_RAW", false),
        SanitizeErrors:    getEnvBool("ERROR_SANITIZE", true),
        MaxErrorDetails:   getEnvInt("ERROR_MAX_DETAILS", 10),
        AlertingEnabled:   getEnvBool("ERROR_ALERTING", true),
    }
    
    return config
}
```

## Common Error Patterns and Solutions

### 1. Database Connection Handling

```go
func (r *BaseRepository) handleDatabaseError(err error, operation, table string) error {
    if err == nil {
        return nil
    }
    
    switch {
    case isDatabaseConnectionError(err):
        return errorcustom.NewRepositoryError(
            operation,
            table,
            "Database connection failed",
            err,
        ).WithDetail("retryable", true)
        
    case isDatabaseTimeoutError(err):
        return errorcustom.NewRepositoryError(
            operation,
            table,
            "Database operation timed out",
            err,
        ).WithDetail("retryable", true)
        
    case isDuplicateKeyError(err):
        return errorcustom.NewRepositoryError(
            operation,
            table,
            "Duplicate key constraint violation",
            err,
        ).WithDetail("retryable", false)
        
    case isForeignKeyError(err):
        return errorcustom.NewRepositoryError(
            operation,
            table,
            "Foreign key constraint violation",
            err,
        ).WithDetail("retryable", false)
        
    default:
        return errorcustom.NewRepositoryError(
            operation,
            table,
            "Database operation failed",
            err,
        )
    }
}
```

### 2. External API Error Handling

```go
func (c *ExternalClient) handleAPIError(resp *http.Response, err error) error {
    if err != nil {
        // Network/connection errors
        return errorcustom.NewServiceError(
            "ExternalAPI",
            "Request",
            "Network error",
            err,
            true, // retryable
        )
    }
    
    switch resp.StatusCode {
    case 400:
        return errorcustom.NewServiceError(
            "ExternalAPI",
            "Request",
            "Bad request to external service",
            nil,
            false,
        ).WithDetail("status_code", resp.StatusCode)
        
    case 401:
        return errorcustom.NewServiceError(
            "ExternalAPI", 
            "Request",
            "Authentication failed with external service",
            nil,
            false,
        ).WithDetail("status_code", resp.StatusCode)
        
    case 429:
        return errorcustom.NewServiceError(
            "ExternalAPI",
            "Request", 
            "Rate limit exceeded",
            nil,
            true, // retryable after delay
        ).WithDetail("status_code", resp.StatusCode)
        
    case 500, 502, 503, 504:
        return errorcustom.NewServiceError(
            "ExternalAPI",
            "Request",
            "External service error", 
            nil,
            true, // retryable
        ).WithDetail("status_code", resp.StatusCode)
        
    default:
        if resp.StatusCode >= 400 {
            return errorcustom.NewServiceError(
                "ExternalAPI",
                "Request",
                fmt.Sprintf("HTTP %d error", resp.StatusCode),
                nil,
                resp.StatusCode >= 500,
            ).WithDetail("status_code", resp.StatusCode)
        }
    }
    
    return nil
}
```

## Performance Optimization

### 1. Error Object Pooling

```go
// Pool APIError objects to reduce allocations
var apiErrorPool = sync.Pool{
    New: func() interface{} {
        return &errorcustom.APIError{
            Details: make(map[string]interface{}),
        }
    },
}

func NewPooledAPIError(code, message string, httpStatus int) *errorcustom.APIError {
    err := apiErrorPool.Get().(*errorcustom.APIError)
    err.Code = code
    err.Message = message
    err.HTTPStatus = httpStatus
    err.Layer = ""
    err.Operation = ""
    err.Cause = nil
    
    // Clear details map
    for k := range err.Details {
        delete(err.Details, k)
    }
    
    return err
}

func ReleaseAPIError(err *errorcustom.APIError) {
    apiErrorPool.Put(err)
}
```

### 2. Response Caching for Common Errors

```go
var commonErrorResponses = map[string][]byte{
    "USER_NOT_FOUND": []byte(`{"code":"USER_NOT_FOUND","message":"User not found"}`),
    "INVALID_INPUT":  []byte(`{"code":"INVALID_INPUT","message":"Invalid input provided"}`),
    "ACCESS_DENIED":  []byte(`{"code":"ACCESS_DENIED","message":"Access denied"}`),
}

func respondWithCachedError(w http.ResponseWriter, errorCode string, httpStatus int) bool {
    if response, exists := commonErrorResponses[errorCode]; exists {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(httpStatus)
        w.Write(response)
        return true
    }
    return false
}
```

## Troubleshooting Guide

### Common Issues and Solutions

1. **Too Many Error Logs**
   ```go
   // Problem: Every validation error creates an ERROR log
   // Solution: Use WARNING for client errors, ERROR only for server errors
   
   if apiErr.HTTPStatus >= 500 {
       logger.Error("Server error", context)
   } else if apiErr.HTTPStatus == 401 {
       logger.Warning("Authentication failed", context) 
   }
   // No logging for 400-level validation errors (logged in API middleware)
   ```

2. **Missing Error Context**
   ```go
   // Problem: Errors lack debugging information
   // Solution: Always add context at each layer
   
   if err != nil {
       return errorcustom.NewServiceError(
           "UserService",
           "CreateUser",
           "Failed to create user",
           err,
           false,
       ).WithDetail("user_email", user.Email).
         WithDetail("request_id", requestID).
         WithDetail("timestamp", time.Now().Unix())
   }
   ```

3. **Inconsistent Error Responses**
   ```go
   // Problem: Different error formats across handlers
   // Solution: Always use utils.HandleError()
   
   // Wrong:
   http.Error(w, "Something went wrong", 500)
   
   // Correct:
   apiErr := errorcustom.NewAPIError("INTERNAL_ERROR", "Something went wrong", 500)
   utils.HandleError(w, apiErr, requestID)
   ```

This comprehensive error handling system provides robust, secure, and maintainable error management for production Go applications. The system automatically handles logging optimization, security considerations, and provides excellent observability for debugging and monitoring.
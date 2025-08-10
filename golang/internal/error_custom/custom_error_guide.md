# Multi-Domain Error Handling System

A comprehensive, scalable error handling system designed to support multiple domains in a Go API application.

## 🚀 Overview

This reorganized error system addresses the limitations of the original account-only focused error handling by providing:

- **Domain-Aware Architecture**: Support for 10+ domains (user, course, payment, content, admin, etc.)
- **Consistent Error Structure**: Standardized error types across all domains
- **Flexible Error Codes**: Dynamic error code generation based on domain and error type
- **Rich Context**: Detailed error information for debugging and client responses
- **Backward Compatibility**: Maintains compatibility with existing error types

## 📁 File Structure

```
internal/error_custom/
├── core.go           # Core error types and interfaces
├── codes.go          # Domain-aware error code system
├── constructors.go   # Error creation functions
├── utilities.go      # Error detection and parsing utilities
└── handler.go        # HTTP error handling and middleware
```

## 🏗️ Architecture

### Core Components

1. **APIError**: Central error type with domain, layer, and operation context
2. **DomainError Interface**: Contract for all domain-specific errors
3. **Generic Error Types**: Reusable error types (NotFoundError, ValidationError, etc.)
4. **Error Collection**: Handles multiple related errors
5. **Domain-Aware Codes**: Dynamic error code generation

### Domain Support

Currently supports these domains (easily extensible):
- `user` - User management and authentication
- `course` - Course and learning content
- `payment` - Payment processing
- `auth` - Authentication and authorization
- `admin` - Administrative operations
- `content` - Content management
- `system` - System-level operations

## 🔧 Usage Examples

### Basic Error Creation

```go
// User domain errors
userNotFound := errorcustom.NewUserNotFoundByID(123)
duplicateEmail := errorcustom.NewDuplicateEmailError("user@example.com")
weakPassword := errorcustom.NewWeakPasswordError([]string{"uppercase letter"})

// Course domain errors
courseNotFound := errorcustom.NewCourseNotFoundError(456)
accessDenied := errorcustom.NewCourseAccessDeniedError(userID, courseID)
enrollmentClosed := errorcustom.NewCourseEnrollmentClosedError(courseID)

// Payment domain errors
insufficientFunds := errorcustom.NewInsufficientFundsError(userID, 100.00, 50.00)
providerError := errorcustom.NewPaymentProviderError("stripe", "charge", err, true)
```

### HTTP Handler Integration

```go
func UserHandler(w http.ResponseWriter, r *http.Request) {
    requestID := errorcustom.GetRequestIDFromContext(r.Context())
    domain := "user"
    
    // Parse parameters with domain context
    userID, err := errorcustom.ParseIDParamWithDomain(r, "id", domain)
    if err != nil {
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    // Business logic...
    if userNotExists {
        err := errorcustom.NewUserNotFoundByID(userID)
        errorcustom.HandleDomainError(w, err, domain, requestID)
        return
    }
    
    // Success response
    errorcustom.RespondWithDomainSuccess(w, userData, domain, requestID)
}
```

### Middleware Setup

```go
r := chi.NewRouter()

// Global middleware
r.Use(errorcustom.RequestIDMiddleware)
r.Use(errorcustom.LogHTTPMiddleware)
r.Use(errorcustom.RecoveryMiddleware)

// Domain-specific routes
r.Route("/api/users", func(r chi.Router) {
    r.Use(errorcustom.DomainMiddleware("user"))
    r.Get("/{id}", UserHandler)
})

r.Route("/api/courses", func(r chi.Router) {
    r.Use(errorcustom.DomainMiddleware("course"))
    r.Get("/{course_id}", CourseHandler)
})
```

### Validation with Multiple Errors

```go
func ValidateRegistration(req *RegistrationRequest, domain string) error {
    errorCollection := errorcustom.NewErrorCollection(domain)
    
    // Email validation
    if err := errorcustom.ValidateEmailWithDomain(req.Email, domain, requestID); err != nil {
        errorCollection.Add(err)
    }
    
    // Password validation
    if err := errorcustom.ValidatePasswordWithDomain(req.Password, domain, requestID); err != nil {
        errorCollection.Add(err)
    }
    
    // Return combined errors
    if errorCollection.HasErrors() {
        return errorCollection.ToAPIError()
    }
    
    return nil
}
```

## 🔍 Error Detection

The system provides robust error detection utilities:

```go
// Check error types
if errorcustom.IsNotFoundError(err) { /* handle not found */ }
if errorcustom.IsValidationError(err) { /* handle validation */ }
if errorcustom.IsRetryableError(err) { /* retry logic */ }

// Check domain-specific errors
if errorcustom.IsDomainError(err, "user") { /* user domain error */ }
if errorcustom.IsUserNotFoundError(err) { /* specific user not found */ }

// Error classification
if errorcustom.IsClientError(err) { /* 4xx errors */ }
if errorcustom.IsServerError(err) { /* 5xx errors */ }
```

## 🌐 Error Codes

Error codes follow a consistent pattern:

- **Generic**: `NOT_FOUND`, `VALIDATION_ERROR`, `AUTHENTICATION_ERROR`
- **Domain-Specific**: `user_NOT_FOUND`, `course_VALIDATION_ERROR`, `payment_EXTERNAL_SERVICE_ERROR`

Examples:
```go
// These all generate appropriate domain-specific codes
errorcustom.NewUserNotFoundByID(123)           // → user_NOT_FOUND
errorcustom.NewCourseNotFoundError(456)        // → course_NOT_FOUND  
errorcustom.NewPaymentProviderError(...)       // → payment_EXTERNAL_SERVICE_ERROR
```

## 📊 Response Format

All errors return a consistent JSON structure:

```json
{
  "code": "user_NOT_FOUND",
  "message": "User with ID 123 not found",
  "details": {
    "user_id": 123,
    "retryable": false
  }
}
```

## 🔄 Migration Guide

### From Existing System

The new system is backward compatible. Existing error types still work:

```go
// Old way (still works)
oldErr := &errorcustom.UserNotFoundError{ID: 123}
apiErr := oldErr.ToAPIError()

// New way (recommended)
newErr := errorcustom.NewUserNotFoundByID(123)
apiErr := newErr.ToAPIError()
```

### Adding New Domains

1. **Add domain constant** in `codes.go`:
```go
const DomainInventory = "inventory"
```

2. **Create domain-specific constructors** in `constructors.go`:
```go
func NewInventoryItemNotFoundError(itemID int64) *NotFoundError {
    return NewNotFoundError(DomainInventory, "inventory_item", itemID)
}
```

3. **Add business logic errors** as needed:
```go
func NewInventoryOutOfStockError(itemID int64, requested, available int) *BusinessLogicError {
    return NewBusinessLogicErrorWithContext(
        DomainInventory,
        "stock_availability",
        "Insufficient inventory",
        map[string]interface{}{
            "item_id": itemID,
            "requested": requested,
            "available": available,
        },
    )
}
```

## 🛡️ Security Considerations

- **Sensitive Data**: Passwords and tokens are automatically redacted in error details
- **Information Disclosure**: Error messages are crafted to avoid exposing internal system details
- **Request IDs**: All errors include request IDs for security audit trails

## 📈 Performance

- **Error Reuse**: Common errors can be pre-created and reused
- **Lazy Evaluation**: Error details are only computed when needed
- **Memory Efficient**: Error collections use slices, not maps
- **Logging Optimization**: Smart logging levels prevent noise

## 🧪 Testing

The system includes comprehensive testing patterns:

```go
func TestUserNotFoundError(t *testing.T) {
    err := errorcustom.NewUserNotFoundByID(123)
    apiErr := err.ToAPIError()
    
    assert.Equal(t, "user_NOT_FOUND", apiErr.Code)
    assert.Equal(t, 404, apiErr.HTTPStatus)
    assert.Equal(t, "user", apiErr.Domain)
    assert.Contains(t, apiErr.Details, "user_id")
}
```

## 🚦 Logging Strategy

The system uses intelligent logging levels:

- **ERROR**: System failures (5xx errors)
- **WARNING**: External service issues, retryable errors
- **INFO**: Client errors (4xx errors) 
- **DEBUG**: Request tracing

## 🔮 Future Enhancements

- **Metrics Integration**: Error rate monitoring per domain
- **Internationalization**: Multi-language error messages
- **Error Templates**: Configurable error message templates
- **Webhook Integration**: Error notification systems
- **Analytics**: Error pattern analysis and reporting

## 📚 Best Practices

1. **Use Domain Context**: Always specify domain for better error tracking
2. **Leverage Collections**: Use ErrorCollection for multiple related errors
3. **Consistent Codes**: Follow the domain_TYPE pattern for error codes
4. **Rich Details**: Include relevant context in error details
5. **Proper Logging**: Use appropriate logging levels for different error types
6. **Test Coverage**: Write tests for all error scenarios
7. **Documentation**: Document domain-specific business rules and error conditions

## 🤝 Contributing

When adding new domains or error types:

1. Follow the established patterns in existing domains
2. Add comprehensive examples in the usage examples file
3. Update this README with new domain information
4. Include tests for new error types
5. Consider backward compatibility implications
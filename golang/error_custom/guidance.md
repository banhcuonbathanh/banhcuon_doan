# Go Custom Error System - Complete Usage Guide

## Table of Contents
1. [Overview](#overview)
2. [Core Components](#core-components)
3. [Error Codes & Constants](#error-codes--constants)
4. [Error Constructors](#error-constructors)
5. [Error Handling Middleware](#error-handling-middleware)
6. [Handler Layer Functions](#handler-layer-functions)
7. [Service Layer Functions](#service-layer-functions)
8. [Repository Layer Functions](#repository-layer-functions)
9. [Unified Error Handler](#unified-error-handler)
10. [Utility Functions](#utility-functions)
11. [Usage Examples](#usage-examples)

## Overview

This system provides comprehensive, domain-aware error handling for Go web applications with structured error responses, middleware integration, and layered error management.

### Key Features
- Domain-aware error categorization (account, auth, admin, system)
- HTTP status code mapping
- Structured JSON error responses
- Request tracing with unique IDs
- Multi-layer error handling (handler, service, repository)
- Middleware integration for automatic error processing

## Core Components

### Base Error Types
```go
type APIError struct {
    Code       string
    Message    string
    Details    map[string]interface{}
    HTTPStatus int
    Domain     string
    Layer      string
    Operation  string
    Cause      error
    Retryable  bool
    Timestamp  time.Time
    RequestID  string
}

type BaseError struct {
    Domain    string
    ErrorType string
}

type DomainError interface {
    error
    ToAPIError() *APIError
    GetDomain() string
    GetErrorType() string
}
```

## Error Codes & Constants

### Domain Constants
```go
const (
    DomainAccount = "account"
    DomainAuth    = "auth" 
    DomainAdmin   = "admin"
    DomainSystem  = "system"
)
```

### Error Type Constants
```go
const (
    ErrorTypeNotFound           = "NOT_FOUND"
    ErrorTypeValidation         = "VALIDATION_ERROR"
    ErrorTypeDuplicate          = "DUPLICATE"
    ErrorTypeAuthentication     = "AUTHENTICATION_ERROR"
    ErrorTypeAuthorization      = "AUTHORIZATION_ERROR"
    ErrorTypeBusinessLogic      = "BUSINESS_LOGIC_ERROR"
    ErrorTypeExternalService    = "EXTERNAL_SERVICE_ERROR"
    ErrorTypeServiceUnavailable = "SERVICE_UNAVAILABLE"
    ErrorTypeSystem             = "SYSTEM_ERROR"
    ErrorTypeInvalidInput       = "INVALID_INPUT"
    ErrorTypeRateLimit          = "RATE_LIMIT"
    ErrorTypeTimeout            = "TIMEOUT"
    ErrorTypeConflict           = "conflict_error"
    ErrorTypeDatabase           = "DATABASE"
)
```

### Code Generation Functions
```go
func GetServiceUnavailableCode(domain string) string {}
func GetSystemErrorCode(domain string) string {}
func GetInvalidInputCode(domain string) string {}
func GetDomainCode(baseCode, domain string) string {}
func GetNotFoundCode(domain string) string {}
func GetValidationCode(domain string) string {}
func GetDuplicateCode(domain string) string {}
func GetAuthenticationCode(domain string) string {}
func GetAuthorizationCode(domain string) string {}
func GetBusinessLogicCode(domain string) string {}
func GetExternalServiceCode(domain string) string {}
func GetSystemCode(domain string) string {}
func GetDatabaseCode(domain string) string {}
func GetRateLimitCode(domain string) string {}
func GetTimeoutCode(domain string) string {}
```

### Code Analysis Functions
```go
func ExtractDomainFromCode(code string) string {}
func GetBaseErrorType(code string) string {}
func IsErrorCodeForDomain(code, domain string) bool {}
func GetDomainSpecificCodes(domain string) map[string]string {}
```

## Error Constructors

### NotFound Errors
```go
func NewNotFoundError(domain, resourceType string, resourceID interface{}) *NotFoundError {}
func NewNotFoundErrorWithIdentifiers(domain, resourceType string, identifiers map[string]interface{}) *NotFoundError {}
func NewNotFoundErrorWithContext(domain, resourceType string, context map[string]interface{}) *NotFoundError {}
```

### Validation Errors
```go
func NewValidationError(domain, field, message string, value interface{}) *ValidationError {}
func NewValidationErrorWithRules(domain, field, message string, value interface{}, rules map[string]interface{}) *ValidationError {}
func NewValidationErrorWithContext(domain, field, message string, value interface{}, context map[string]interface{}) *ValidationError {}
```

### Authentication Errors
```go
func NewAuthenticationError(domain, reason string) *AuthenticationError {}
func NewAuthenticationErrorWithStep(domain, reason, step string, context map[string]interface{}) *AuthenticationError {}
func NewAuthenticationErrorWithContext(domain, reason string, context map[string]interface{}) *AuthenticationError {}
```

### Authorization Errors
```go
func NewAuthorizationError(domain, action, resource string) *AuthorizationError {}
func NewAuthorizationErrorWithContext(domain, action, resource string, context map[string]interface{}) *AuthorizationError {}
```

### Business Logic Errors
```go
func NewBusinessLogicError(domain, rule, description string) *BusinessLogicError {}
func NewBusinessLogicErrorWithContext(domain, rule, description string, context map[string]interface{}) *BusinessLogicError {}
```

### Duplicate/Conflict Errors
```go
func NewDuplicateError(domain, resourceType, field string, value interface{}) *DuplicateError {}
func NewConflictErrorWithContext(domain, resourceType, message string, context map[string]interface{}) *ConflictError {}
```

### Rate Limit Errors
```go
func NewRateLimitErrorWithContext(domain, operation, message string, context map[string]interface{}) *RateLimitError {}
```

### External Service & System Errors
```go
func NewExternalServiceError(domain, service, operation, message string, cause error, retryable bool) *ExternalServiceError {}
func NewSystemError(domain, component, operation, message string, cause error) *SystemError {}
```

### Domain-Specific Constructors

#### Account Domain
```go
func NewUserNotFoundByID(id int64) *NotFoundError {}
func NewUserNotFoundByEmail(email string) *NotFoundError {}
func NewEmailNotFoundError(email string) *AuthenticationError {}
func NewPasswordMismatchError(email string) *AuthenticationError {}
func NewAccountDisabledError(email string) *AuthenticationError {}
func NewAccountLockedError(email string, lockReason string) *AuthenticationError {}
func NewDuplicateEmailError(email string) *DuplicateError {}
func NewWeakPasswordError(requirements []string) *ValidationError {}
```

#### Admin Domain
```go
func NewInsufficientPrivilegesError(userID int64, requiredRole, currentRole string) *AuthorizationError {}
func NewBulkOperationLimitError(operation string, requested, maxAllowed int) *BusinessLogicError {}
```

#### System Domain
```go
func NewDatabaseError(operation, table string, cause error) *SystemError {}
func NewCacheError(operation, key string, cause error) *SystemError {}
func NewFileSystemError(operation, path string, cause error) *SystemError {}
```

### Security Errors
```go
func NewSecurityError(domain, securityCode, message string) *BusinessLogicError {}
func NewSecurityErrorWithContext(domain, securityCode, message string, context map[string]interface{}) *BusinessLogicError {}
func NewOriginNotAllowedError(domain, origin string, allowedOrigins []string) *BusinessLogicError {}
func NewSuspiciousActivityError(domain, activity, reason string, context map[string]interface{}) *BusinessLogicError {}
func NewRateLimitExceededError(domain, operation string, limit int, timeWindow string) *RateLimitError {}
```

### Core API Error Constructors
```go
func NewAPIError(code, message string, httpStatus int) *APIError {}
func NewAPIErrorWithContext(code, message string, httpStatus int, domain, layer, operation string, cause error) *APIError {}
func NewErrorResponse(code, message string) ErrorResponse {}
```

## Error Handling Middleware

### Error Middleware Manager
```go
func NewErrorMiddleware() *ErrorMiddleware {}
```

### Middleware Functions
```go
func (em *ErrorMiddleware) RequestIDMiddleware(next http.Handler) http.Handler {}
func (em *ErrorMiddleware) DomainMiddleware(domain string) func(http.Handler) http.Handler {}
func (em *ErrorMiddleware) AutoDomainMiddleware(next http.Handler) http.Handler {}
func (em *ErrorMiddleware) RecoveryMiddleware(next http.Handler) http.Handler {}
func (em *ErrorMiddleware) LoggingMiddleware(next http.Handler) http.Handler {}
```

### Legacy Middleware Functions
```go
func RequestIDMiddleware(next http.Handler) http.Handler {}
func DomainContextMiddleware(next http.Handler) http.Handler {}
func LogHTTPMiddleware(next http.Handler) http.Handler {}
func RecoveryMiddleware(next http.Handler) http.Handler {}
func JWTValidationMiddleware(secretKey string) func(http.Handler) http.Handler {}
func RateLimitMiddleware(domain string) func(http.Handler) http.Handler {}
```

## Handler Layer Functions

### Handler Error Manager
```go
func NewHandlerErrorManager() *HandlerErrorManager {}
```

### Response Functions
```go
func (h *HandlerErrorManager) RespondWithError(w http.ResponseWriter, err error, domain, requestID string) {}
func (h *HandlerErrorManager) RespondWithSuccess(w http.ResponseWriter, data interface{}, domain, requestID string) {}
func (h *HandlerErrorManager) RespondWithCreated(w http.ResponseWriter, data interface{}, domain, requestID string) {}
```

### Parameter Parsing
```go
func (h *HandlerErrorManager) ParseIDParameter(r *http.Request, paramName, domain, requestID string) (int64, error) {}
func (h *HandlerErrorManager) ParsePaginationParameters(r *http.Request, domain, requestID string) (limit, offset int64, err error) {}
func (h *HandlerErrorManager) ParseSortingParameters(r *http.Request, allowedFields []string, domain, requestID string) (sortBy, sortOrder string, err error) {}
```

### Request Processing
```go
func (h *HandlerErrorManager) DecodeJSONRequest(r *http.Request, target interface{}, domain, requestID string) error {}
func (h *HandlerErrorManager) ValidateRequiredFields(data map[string]interface{}, requiredFields []string, domain, requestID string) error {}
```

### Legacy Handler Functions
```go
func HandleError(w http.ResponseWriter, err error, requestID string) {}
func HandleValidationErrors(w http.ResponseWriter, validationErrors validator.ValidationErrors, domain, requestID string) {}
func DecodeJSON(body io.Reader, target interface{}, domain, requestID string) error {}
func GetPaginationParams(r *http.Request, domain string) (limit, offset int64, err error) {}
func ValidateEmail(email, domain string) error {}
func ValidatePassword(password, domain string) error {}
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}, requestID string) {}
func RespondWithSuccess(w http.ResponseWriter, data interface{}, domain, requestID string) {}
```

## Service Layer Functions

### Service Error Manager
```go
func NewServiceErrorManager() *ServiceErrorManager {}
```

### Error Handling Functions
```go
func (s *ServiceErrorManager) WrapRepositoryError(err error, domain, operation string, context map[string]interface{}) error {}
func (s *ServiceErrorManager) HandleBusinessRuleViolation(domain, rule, description string, context map[string]interface{}) error {}
func (s *ServiceErrorManager) HandleExternalServiceError(err error, domain, service, operation string, retryable bool) error {}
func (s *ServiceErrorManager) HandleContextError(ctx context.Context, domain, operation string) error {}
func (s *ServiceErrorManager) ValidateBusinessRules(domain string, validations map[string]func() error) error {}
```

## Repository Layer Functions

### Repository Error Manager
```go
func NewRepositoryErrorManager() *RepositoryErrorManager {}
```

### Database Error Handling
```go
func (r *RepositoryErrorManager) HandleDatabaseError(err error, domain, table, operation string, context map[string]interface{}) error {}
```

## Unified Error Handler

### Main Handler
```go
func NewUnifiedErrorHandler() *UnifiedErrorHandler {}
```

### HTTP Handler Methods
```go
func (ueh *UnifiedErrorHandler) HandleHTTPError(w http.ResponseWriter, r *http.Request, err error) {}
func (ueh *UnifiedErrorHandler) ParseIDParam(r *http.Request, paramName string, domain string) (int64, error) {}
func (ueh *UnifiedErrorHandler) ParseStringParam(r *http.Request, paramName string, minLen int) (string, error) {}
func (ueh *UnifiedErrorHandler) GetSortParamsWithDomain(r *http.Request, allowedFields []string, domain string) (sortBy, sortOrder string, err error) {}
func (ueh *UnifiedErrorHandler) ParsePaginationParams(r *http.Request) (limit, offset int64, err error) {}
func (ueh *UnifiedErrorHandler) DecodeJSONRequest(r *http.Request, target interface{}) error {}
func (ueh *UnifiedErrorHandler) RespondWithSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {}
```

### Service Layer Methods
```go
func (ueh *UnifiedErrorHandler) WrapRepositoryError(err error, domain, operation string, context map[string]interface{}) error {}
func (ueh *UnifiedErrorHandler) HandleBusinessRuleViolation(domain, rule, description string, context map[string]interface{}) error {}
func (ueh *UnifiedErrorHandler) HandleExternalServiceError(err error, domain, service, operation string, retryable bool) error {}
func (ueh *UnifiedErrorHandler) HandleContextError(ctx context.Context, domain, operation string) error {}
func (ueh *UnifiedErrorHandler) ValidateBusinessRules(domain string, validations map[string]func() error) error {}
func (ueh *UnifiedErrorHandler) ParseGRPCError(err error, domain, operation string, context map[string]interface{}) error {}
```

### Repository Layer Methods
```go
func (ueh *UnifiedErrorHandler) HandleDatabaseError(err error, domain, table, operation string, context map[string]interface{}) error {}
```

### General Methods
```go
func (ueh *UnifiedErrorHandler) HandleError(domain string, err error) error {}
```

## Utility Functions

### Error Detection
```go
func IsDomainError(err error, domain string) bool {}
func IsErrorType(err error, errorType string) bool {}
func IsNotFoundError(err error) bool {}
func IsValidationError(err error) bool {}
func IsAuthenticationError(err error) bool {}
func IsAuthorizationError(err error) bool {}
func IsBusinessLogicError(err error) bool {}
func IsExternalServiceError(err error) bool {}
func IsSystemError(err error) bool {}
func IsRetryableError(err error) bool {}
func IsClientError(err error) bool {}
func IsServerError(err error) bool {}
func IsUserNotFoundError(err error) bool {}
func IsPasswordError(err error) bool {}
```

### Error Conversion
```go
func ConvertToAPIError(err error) *APIError {}
func ParseGRPCError(err error, domain, operation string, context map[string]interface{}) error {}
func AsAPIError(err error) (*APIError, bool) {}
func MustAPIError(err error) *APIError {}
```

### Error Collections
```go
func NewErrorCollection(domain string) *ErrorCollection {}
func (ec *ErrorCollection) Add(err error) {}
func (ec *ErrorCollection) HasErrors() bool {}
func (ec *ErrorCollection) Count() int {}
func (ec *ErrorCollection) ToAPIError() *APIError {}
func (ec *ErrorCollection) Error() string {}
```

### Logging & Analysis
```go
func GetErrorSeverity(err error) string {}
func ShouldLogError(err error) bool {}
```

### Context Functions
```go
func GetRequestIDFromContext(ctx context.Context) string {}
func GetDomainFromContext(ctx context.Context) string {}
func GetUserEmailFromContext(r *http.Request) string {}
func GetUserIDFromContext(r *http.Request) int64 {}
func GetClientIP(r *http.Request) string {}
```

### Factory Functions
```go
func NewErrorFactory() *ErrorFactory {}
```

## Usage Examples

### 1. Basic Error Creation
```go
// Not found error
err := NewUserNotFoundByID(123)

// Validation error
err := NewValidationError("account", "email", "Invalid email format", "invalid-email")

// Authentication error
err := NewPasswordMismatchError("user@example.com")
```

### 2. Middleware Setup
```go
r := chi.NewRouter()
errorMW := NewErrorMiddleware()

r.Use(errorMW.RequestIDMiddleware)
r.Use(errorMW.AutoDomainMiddleware) 
r.Use(errorMW.RecoveryMiddleware)
r.Use(errorMW.LoggingMiddleware)
```

### 3. Handler Usage
```go
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    ueh := NewUnifiedErrorHandler()
    
    userID, err := ueh.ParseIDParam(r, "id", DomainAccount)
    if err != nil {
        ueh.HandleHTTPError(w, r, err)
        return
    }
    
    user, err := h.userService.GetUser(userID)
    if err != nil {
        ueh.HandleHTTPError(w, r, err)
        return
    }
    
    ueh.RespondWithSuccess(w, r, user)
}
```

### 4. Service Layer Usage
```go
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    ueh := NewUnifiedErrorHandler()
    
    // Check for duplicate email
    if s.userExists(req.Email) {
        return nil, NewDuplicateEmailError(req.Email)
    }
    
    // Validate business rules
    err := ueh.ValidateBusinessRules(DomainAccount, map[string]func() error{
        "email_format": func() error {
            return ValidateEmail(req.Email, DomainAccount)
        },
        "password_strength": func() error {
            return ValidatePassword(req.Password, DomainAccount)
        },
    })
    if err != nil {
        return nil, err
    }
    
    // Create user in database
    user, err := s.userRepo.Create(req)
    if err != nil {
        return nil, ueh.WrapRepositoryError(err, DomainAccount, "create_user", nil)
    }
    
    return user, nil
}
```

### 5. Error Detection in Business Logic
```go
func handleUserError(err error) {
    switch {
    case IsUserNotFoundError(err):
        log.Info("User not found")
    case IsPasswordError(err):
        log.Warning("Authentication failed")
    case IsValidationError(err):
        log.Info("Validation failed")
    case IsRetryableError(err):
        log.Warning("Temporary failure, can retry")
    default:
        log.Error("Unexpected error")
    }
}
```

This system provides comprehensive error handling with domain awareness, structured responses, and seamless integration across all application layers.
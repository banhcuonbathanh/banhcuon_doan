# Go Error Handling System - Complete Context & API Reference

## System Overview

This is a comprehensive, domain-aware error handling system for a Go web application. It provides structured error handling across multiple layers (Handler, Service, Repository) with domain-specific error types and codes.

## Architecture

The system follows a layered architecture:
- **Handler Layer**: HTTP request/response handling
- **Service Layer**: Business logic and rules
- **Repository Layer**: Database operations
- **Unified Handler**: Single interface for all layers

## Domain Structure

The system supports multiple domains:
- `account` - User accounts and authentication
- `auth` - Authentication services
- `admin` - Administrative operations
- `system` - System-level operations

---

## Constants and Enums

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

### Legacy Error Codes
```go
const (
    ErrCodeUserNotFound     = "user_NOT_FOUND"
    ErrCodeDuplicateEmail   = "user_DUPLICATE"
    ErrCodeWeakPassword     = "user_WEAK_PASSWORD"
    ErrCodeAuthFailed       = "auth_AUTHENTICATION_ERROR"
    ErrCodeAccessDenied     = "auth_AUTHORIZATION_ERROR"
    ErrCodeInvalidToken     = "auth_INVALID_TOKEN"
    ErrCodeNotFound         = "NOT_FOUND"
    ErrCodeValidationError  = "VALIDATION_ERROR"
    ErrCodeInvalidInput     = "INVALID_INPUT"
    ErrCodeInternalError    = "SYSTEM_ERROR"
    ErrCodeServiceError     = "EXTERNAL_SERVICE_ERROR"
    ErrCodeRepositoryError  = "system_REPOSITORY_ERROR"
)
```

### Context Keys
```go
const (
    ContextKeyRequestID contextKey = "request_id"
    ContextKeyDomain    contextKey = "domain"
    ContextKeyUserID    contextKey = "user_id"
)
```

---

## Core Types

### Base Types
```go
type contextKey string

type DomainError interface {
    error
    ToAPIError() *APIError
    GetDomain() string
    GetErrorType() string
}

type BaseError struct {
    Domain    string `json:"domain"`
    ErrorType string `json:"error_type"`
}
```

### Primary Error Types
```go
type APIError struct {
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    HTTPStatus int                    `json:"-"`
    Domain     string                 `json:"domain,omitempty"`
    Layer      string                 `json:"layer,omitempty"`
    Operation  string                 `json:"operation,omitempty"`
    Cause      error                  `json:"-"`
    Retryable  bool                   `json:"retryable,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
    RequestID  string                 `json:"request_id,omitempty"`
}

type ErrorResponse struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### Specific Error Types
```go
type NotFoundError struct {
    BaseError
    ResourceType string                 `json:"resource_type"`
    ResourceID   interface{}            `json:"resource_id,omitempty"`
    Identifiers  map[string]interface{} `json:"identifiers,omitempty"`
    Context      map[string]interface{} `json:"context,omitempty"`
}

type ValidationError struct {
    BaseError
    Field   string                 `json:"field"`
    Message string                 `json:"message"`
    Value   interface{}            `json:"value,omitempty"`
    Rules   map[string]interface{} `json:"rules,omitempty"`
    Context map[string]interface{} `json:"context,omitempty"`
}

type DuplicateError struct {
    BaseError
    ResourceType string                 `json:"resource_type"`
    Field        string                 `json:"field"`
    Value        interface{}            `json:"value"`
    Constraints  map[string]interface{} `json:"constraints,omitempty"`
}

type AuthenticationError struct {
    BaseError
    Reason  string                 `json:"reason"`
    Step    string                 `json:"step,omitempty"`
    Context map[string]interface{} `json:"context,omitempty"`
}

type AuthorizationError struct {
    BaseError
    Action   string                 `json:"action"`
    Resource string                 `json:"resource"`
    Context  map[string]interface{} `json:"context,omitempty"`
}

type BusinessLogicError struct {
    BaseError
    Rule        string                 `json:"rule"`
    Description string                 `json:"description"`
    Context     map[string]interface{} `json:"context,omitempty"`
}

type ConflictError struct {
    BaseError
    ResourceType string                 `json:"resource_type"`
    Field        string                 `json:"field,omitempty"`
    Value        interface{}            `json:"value,omitempty"`
    Message      string                 `json:"message"`
    Context      map[string]interface{} `json:"context,omitempty"`
}

type RateLimitError struct {
    BaseError
    Operation string                 `json:"operation"`
    Message   string                 `json:"message"`
    Context   map[string]interface{} `json:"context,omitempty"`
}

type ExternalServiceError struct {
    BaseError
    Service   string `json:"service"`
    Operation string `json:"operation"`
    Message   string `json:"message"`
    Cause     error  `json:"-"`
    Retryable bool   `json:"retryable"`
}

type SystemError struct {
    BaseError
    Component string `json:"component"`
    Operation string `json:"operation"`
    Message   string `json:"message"`
    Cause     error  `json:"-"`
}
```

### Manager Types
```go
type ErrorFactory struct {
    HandlerErrorMgr    *HandlerErrorManager
    ServiceErrorMgr    *ServiceErrorManager
    RepositoryErrorMgr *RepositoryErrorManager
}

type HandlerErrorManager struct{}
type ServiceErrorManager struct{}
type RepositoryErrorManager struct{}
type UnifiedErrorHandler struct {
    errorFactory *ErrorFactory
}

type ErrorMiddleware struct {
    errorFactory *ErrorFactory
}

type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.RWMutex
    limit    int
    window   time.Duration
}
```

---

## Function Signatures

### Error Code Generation Functions
```go
func GetServiceUnavailableCode(domain string) string
func GetSystemErrorCode(domain string) string
func GetInvalidInputCode(domain string) string
func GetDomainCode(baseCode, domain string) string
func GetNotFoundCode(domain string) string
func GetValidationCode(domain string) string
func GetDuplicateCode(domain string) string
func GetAuthenticationCode(domain string) string
func GetAuthorizationCode(domain string) string
func GetBusinessLogicCode(domain string) string
func GetExternalServiceCode(domain string) string
func GetSystemCode(domain string) string
func GetDatabaseCode(domain string) string
func GetRateLimitCode(domain string) string
func GetTimeoutCode(domain string) string
func ExtractDomainFromCode(code string) string
func GetBaseErrorType(code string) string
func IsErrorCodeForDomain(code, domain string) bool
func GetDomainSpecificCodes(domain string) map[string]string
func splitErrorCode(code string) []string
```

### Core Constructor Functions
```go
func NewAPIError(code, message string, httpStatus int) *APIError
func NewAPIErrorWithContext(code, message string, httpStatus int, domain, layer, operation string, cause error) *APIError
func NewErrorResponse(code, message string) ErrorResponse
func NewErrorFactory() *ErrorFactory
```

### Specific Error Constructors - NotFound Errors
```go
func NewNotFoundError(domain, resourceType string, resourceID interface{}) *NotFoundError
func NewNotFoundErrorWithIdentifiers(domain, resourceType string, identifiers map[string]interface{}) *NotFoundError
func NewNotFoundErrorWithContext(domain, resourceType string, context map[string]interface{}) *NotFoundError
```

### Specific Error Constructors - Validation Errors
```go
func NewValidationError(domain, field, message string, value interface{}) *ValidationError
func NewValidationErrorWithRules(domain, field, message string, value interface{}, rules map[string]interface{}) *ValidationError
func NewValidationErrorWithContext(domain, field, message string, value interface{}, context map[string]interface{}) *ValidationError
```

### Specific Error Constructors - Authentication Errors
```go
func NewAuthenticationError(domain, reason string) *AuthenticationError
func NewAuthenticationErrorWithStep(domain, reason, step string, context map[string]interface{}) *AuthenticationError
func NewAuthenticationErrorWithContext(domain, reason string, context map[string]interface{}) *AuthenticationError
```

### Specific Error Constructors - Authorization Errors
```go
func NewAuthorizationError(domain, action, resource string) *AuthorizationError
func NewAuthorizationErrorWithContext(domain, action, resource string, context map[string]interface{}) *AuthorizationError
```

### Specific Error Constructors - Business Logic Errors
```go
func NewBusinessLogicError(domain, rule, description string) *BusinessLogicError
func NewBusinessLogicErrorWithContext(domain, rule, description string, context map[string]interface{}) *BusinessLogicError
```

### Specific Error Constructors - Duplicate/Conflict Errors
```go
func NewDuplicateError(domain, resourceType, field string, value interface{}) *DuplicateError
func NewConflictErrorWithContext(domain, resourceType, message string, context map[string]interface{}) *ConflictError
func NewRateLimitErrorWithContext(domain, operation, message string, context map[string]interface{}) *RateLimitError
```

### Specific Error Constructors - External/System Errors
```go
func NewExternalServiceError(domain, service, operation, message string, cause error, retryable bool) *ExternalServiceError
func NewSystemError(domain, component, operation, message string, cause error) *SystemError
```

### Domain-Specific Error Constructors
```go
func NewUserNotFoundByID(id int64) *NotFoundError
func NewUserNotFoundByEmail(email string) *NotFoundError
func NewEmailNotFoundError(email string) *AuthenticationError
func NewPasswordMismatchError(email string) *AuthenticationError
func NewAccountDisabledError(email string) *AuthenticationError
func NewAccountLockedError(email string, lockReason string) *AuthenticationError
func NewDuplicateEmailError(email string) *DuplicateError
func NewWeakPasswordError(requirements []string) *ValidationError
func NewInsufficientPrivilegesError(userID int64, requiredRole, currentRole string) *AuthorizationError
func NewBulkOperationLimitError(operation string, requested, maxAllowed int) *BusinessLogicError
func NewDatabaseError(operation, table string, cause error) *SystemError
func NewCacheError(operation, key string, cause error) *SystemError
func NewFileSystemError(operation, path string, cause error) *SystemError
```

### Security Error Constructors
```go
func NewSecurityError(domain, securityCode, message string) *BusinessLogicError
func NewSecurityErrorWithContext(domain, securityCode, message string, context map[string]interface{}) *BusinessLogicError
func NewOriginNotAllowedError(domain, origin string, allowedOrigins []string) *BusinessLogicError
func NewSuspiciousActivityError(domain, activity, reason string, context map[string]interface{}) *BusinessLogicError
func NewRateLimitExceededError(domain, operation string, limit int, timeWindow string) *RateLimitError
```

### Legacy Constructor Functions
```go
func NewServiceError(service, method, message string, cause error, retryable bool) *ExternalServiceError
func NewRepositoryError(operation, table, message string, cause error) *SystemError
```

### APIError Methods
```go
func (e *APIError) Error() string
func (e *APIError) WithDomain(domain string) *APIError
func (e *APIError) WithLayer(layer string) *APIError
func (e *APIError) WithOperation(operation string) *APIError
func (e *APIError) WithDetail(key string, value interface{}) *APIError
func (e *APIError) WithCause(cause error) *APIError
func (e *APIError) WithRetryable(retryable bool) *APIError
func (e *APIError) GetLogContext() map[string]interface{}
func (e *APIError) ToJSON() ([]byte, error)
func (e *APIError) ToErrorResponse() ErrorResponse
```

### BaseError Methods
```go
func (b *BaseError) GetDomain() string
func (b *BaseError) GetErrorType() string
```

### Specific Error Type Methods
```go
func (e *NotFoundError) Error() string
func (e *NotFoundError) ToAPIError() *APIError
func (e *ValidationError) Error() string
func (e *ValidationError) ToAPIError() *APIError
func (e *DuplicateError) Error() string
func (e *DuplicateError) ToAPIError() *APIError
func (e *AuthenticationError) Error() string
func (e *AuthenticationError) ToAPIError() *APIError
func (e *AuthorizationError) Error() string
func (e *AuthorizationError) ToAPIError() *APIError
func (e *BusinessLogicError) Error() string
func (e *BusinessLogicError) ToAPIError() *APIError
func (e *ExternalServiceError) Error() string
func (e *ExternalServiceError) ToAPIError() *APIError
func (e *SystemError) Error() string
func (e *SystemError) ToAPIError() *APIError
func (e *RateLimitError) Error() string
func (e *ConflictError) Error() string
```

### ErrorResponse Methods
```go
func (er ErrorResponse) WithDetail(key string, value interface{}) ErrorResponse
```

### Manager Creation Functions
```go
func NewHandlerErrorManager() *HandlerErrorManager
func NewServiceErrorManager() *ServiceErrorManager
func NewRepositoryErrorManager() *RepositoryErrorManager
func NewUnifiedErrorHandler() *UnifiedErrorHandler
func NewErrorMiddleware() *ErrorMiddleware
```

### Handler Error Manager Methods
```go
func (h *HandlerErrorManager) RespondWithError(w http.ResponseWriter, err error, domain, requestID string)
func (h *HandlerErrorManager) ParseIDParameter(r *http.Request, paramName, domain, requestID string) (int64, error)
func (h *HandlerErrorManager) ParsePaginationParameters(r *http.Request, domain, requestID string) (limit, offset int64, err error)
func (h *HandlerErrorManager) DecodeJSONRequest(r *http.Request, target interface{}, domain, requestID string) error
func (h *HandlerErrorManager) ValidateRequiredFields(data map[string]interface{}, requiredFields []string, domain, requestID string) error
func (h *HandlerErrorManager) RespondWithSuccess(w http.ResponseWriter, data interface{}, domain, requestID string)
func (h *HandlerErrorManager) RespondWithCreated(w http.ResponseWriter, data interface{}, domain, requestID string)
func (h *HandlerErrorManager) ParseSortingParameters(r *http.Request, allowedFields []string, domain, requestID string) (sortBy, sortOrder string, err error)
```

### Service Error Manager Methods
```go
func (s *ServiceErrorManager) WrapRepositoryError(err error, domain, operation string, context map[string]interface{}) error
func (s *ServiceErrorManager) HandleBusinessRuleViolation(domain, rule, description string, context map[string]interface{}) error
func (s *ServiceErrorManager) HandleExternalServiceError(err error, domain, service, operation string, retryable bool) error
func (s *ServiceErrorManager) HandleTransactionError(err error, domain, operation string) error
func (s *ServiceErrorManager) HandleContextError(ctx context.Context, domain, operation string) error
func (s *ServiceErrorManager) ValidateBusinessRules(domain string, validations map[string]func() error) error
```

### Repository Error Manager Methods
```go
func (r *RepositoryErrorManager) HandleDatabaseError(err error, domain, table, operation string, context map[string]interface{}) error
```

### Unified Error Handler Methods
```go
func (ueh *UnifiedErrorHandler) HandleHTTPError(w http.ResponseWriter, r *http.Request, err error)
func (ueh *UnifiedErrorHandler) ParseIDParam(r *http.Request, paramName string, domain string) (int64, error)
func (ueh *UnifiedErrorHandler) ParseStringParam(r *http.Request, paramName string, minLen int) (string, error)
func (ueh *UnifiedErrorHandler) GetSortParamsWithDomain(r *http.Request, allowedFields []string, domain string) (sortBy, sortOrder string, err error)
func (ueh *UnifiedErrorHandler) ParsePaginationParams(r *http.Request) (limit, offset int64, err error)
func (ueh *UnifiedErrorHandler) DecodeJSONRequest(r *http.Request, target interface{}) error
func (ueh *UnifiedErrorHandler) RespondWithSuccess(w http.ResponseWriter, r *http.Request, data interface{})
func (ueh *UnifiedErrorHandler) WrapRepositoryError(err error, domain, operation string, context map[string]interface{}) error
func (ueh *UnifiedErrorHandler) HandleBusinessRuleViolation(domain, rule, description string, context map[string]interface{}) error
func (ueh *UnifiedErrorHandler) HandleExternalServiceError(err error, domain, service, operation string, retryable bool) error
func (ueh *UnifiedErrorHandler) HandleContextError(ctx context.Context, domain, operation string) error
func (ueh *UnifiedErrorHandler) ValidateBusinessRules(domain string, validations map[string]func() error) error
func (ueh *UnifiedErrorHandler) ParseGRPCError(err error, domain, operation string, context map[string]interface{}) error
func (ueh *UnifiedErrorHandler) HandleDatabaseError(err error, domain, table, operation string, context map[string]interface{}) error
func (ueh *UnifiedErrorHandler) HandleError(domain string, err error) error
```

### Middleware Functions
```go
func RequestIDMiddleware(next http.Handler) http.Handler
func DomainContextMiddleware(next http.Handler) http.Handler
func LogHTTPMiddleware(next http.Handler) http.Handler
func RecoveryMiddleware(next http.Handler) http.Handler
func JWTValidationMiddleware(secretKey string) func(http.Handler) http.Handler
func RateLimitMiddleware(domain string) func(http.Handler) http.Handler
```

### Error Middleware Methods
```go
func (em *ErrorMiddleware) RequestIDMiddleware(next http.Handler) http.Handler
func (em *ErrorMiddleware) DomainMiddleware(domain string) func(http.Handler) http.Handler
func (em *ErrorMiddleware) AutoDomainMiddleware(next http.Handler) http.Handler
func (em *ErrorMiddleware) RecoveryMiddleware(next http.Handler) http.Handler
func (em *ErrorMiddleware) LoggingMiddleware(next http.Handler) http.Handler
```

### HTTP Handling Functions
```go
func HandleError(w http.ResponseWriter, err error, requestID string)
func HandleValidationErrors(w http.ResponseWriter, validationErrors validator.ValidationErrors, domain, requestID string)
func DecodeJSON(body io.Reader, target interface{}, domain, requestID string) error
func GetPaginationParams(r *http.Request, domain string) (limit, offset int64, err error)
func ValidateEmail(email, domain string) error
func ValidatePassword(password, domain string) error
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}, requestID string)
func RespondWithSuccess(w http.ResponseWriter, data interface{}, domain, requestID string)
```

### Context Utility Functions
```go
func withRequestID(ctx context.Context, requestID string) context.Context
func withDomain(ctx context.Context, domain string) context.Context
func GetUserEmailFromContext(r *http.Request) string
func GetUserIDFromContext(r *http.Request) int64
func generateRequestID() string
func GetRequestIDFromContext(ctx context.Context) string
func GetDomainFromContext(ctx context.Context) string
func GetClientIP(r *http.Request) string
```

### Helper Functions
```go
func detectDomainFromPath(path string) string
func shouldSkipJWTValidation(path string) bool
func handleAuthError(w http.ResponseWriter, r *http.Request, message string)
func addUserToContext(ctx context.Context, claims jwt.MapClaims) context.Context
func handleJSONError(err error, domain string) error
func logError(apiErr *APIError, requestID string)
func getValidationMessage(fe validator.FieldError) string
func isRateLimited(domain, clientIP string) bool
```

### Error Collection Functions (Referenced but not fully defined in provided code)
```go
func NewErrorCollection(domain string) *ErrorCollection
func (ec *ErrorCollection) Add(err error)
func (ec *ErrorCollection) HasErrors() bool
func (ec *ErrorCollection) Count() int
func (ec *ErrorCollection) ToAPIError() *APIError
```

### Utility Functions (Referenced but not fully defined)
```go
func ConvertToAPIError(err error) *APIError
func ShouldLogError(apiErr *APIError) bool
func GetErrorSeverity(apiErr *APIError) string
func ParseGRPCError(err error, domain, operation string, context map[string]interface{}) error
```

---

## Usage Patterns

### Basic Error Creation
```go
// Create domain-specific validation error
err := NewValidationError("account", "email", "Invalid email format", "invalid-email")

// Create not found error with ID
err := NewUserNotFoundByID(123)

// Create business logic error
err := NewBusinessLogicError("account", "insufficient_balance", "Account balance too low")
```

### Handler Layer Usage
```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID, err := h.errorHandler.ParseIDParam(r, "id", "account")
    if err != nil {
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    // ... business logic
    
    h.errorHandler.RespondWithSuccess(w, r, userData)
}
```

### Service Layer Usage
```go
func (s *Service) CreateUser(userData UserData) error {
    // Validate business rules
    err := s.errorHandler.ValidateBusinessRules("account", map[string]func() error{
        "unique_email": func() error { return s.validateUniqueEmail(userData.Email) },
        "password_strength": func() error { return s.validatePassword(userData.Password) },
    })
    if err != nil {
        return err
    }
    
    // Handle repository errors
    if err := s.repo.CreateUser(userData); err != nil {
        return s.errorHandler.WrapRepositoryError(err, "account", "create_user", map[string]interface{}{
            "email": userData.Email,
        })
    }
    
    return nil
}
```

### Repository Layer Usage
```go
func (r *Repository) GetUserByID(id int64) (*User, error) {
    user := &User{}
    err := r.db.Get(user, "SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, r.errorHandler.HandleDatabaseError(err, "account", "users", "select", map[string]interface{}{
            "user_id": id,
        })
    }
    return user, nil
}
```

This system provides comprehensive error handling with domain awareness, structured logging, and consistent HTTP responses across all application layers.
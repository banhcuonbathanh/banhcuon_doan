# Combining Configuration and Error Systems - Complete Integration Guide

## Overview

This guide shows how to integrate the Go Configuration Package with the Custom Error System to create a unified, domain-aware application architecture that provides:

- **Domain-aware error handling** based on configuration settings
- **Environment-specific error behavior** (dev vs production)
- **Configuration-driven error policies** per domain
- **Unified initialization** and lifecycle management
- **Hot-reloading** of both config and error handling policies

## Integration Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Config        │    │   Error         │
│   System        │◄──►│   System        │
└─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────────────────────────────┐
│        Unified Manager                  │
│  - Domain Configuration                 │
│  - Error Policies                       │
│  - Environment Awareness                │
│  - Hot Reload Support                   │
└─────────────────────────────────────────┘
```

## 1. Enhanced Configuration Structure

### Extended Config with Error Handling Settings

```go
// Add to your existing Config struct
type Config struct {
    // ... existing fields ...
    
    // Enhanced error handling configuration
    ErrorHandling ErrorHandlingConfig `mapstructure:"error_handling" yaml:"error_handling"`
    
    // Domain-specific error policies
    DomainErrorPolicies map[string]DomainErrorPolicy `mapstructure:"domain_error_policies" yaml:"domain_error_policies"`
}

type ErrorHandlingConfig struct {
    IncludeStackTrace      bool   `mapstructure:"include_stack_trace" yaml:"include_stack_trace"`
    SanitizeSensitiveData  bool   `mapstructure:"sanitize_sensitive_data" yaml:"sanitize_sensitive_data"`
    LogLevel               string `mapstructure:"log_level" yaml:"log_level"`
    EnableDetailedLogging  bool   `mapstructure:"enable_detailed_logging" yaml:"enable_detailed_logging"`
    EnableMetrics          bool   `mapstructure:"enable_metrics" yaml:"enable_metrics"`
    MaxErrorsPerMinute     int    `mapstructure:"max_errors_per_minute" yaml:"max_errors_per_minute"`
    AlertOnCriticalErrors  bool   `mapstructure:"alert_on_critical_errors" yaml:"alert_on_critical_errors"`
}

type DomainErrorPolicy struct {
    EnableStackTrace     bool              `mapstructure:"enable_stack_trace" yaml:"enable_stack_trace"`
    LogLevel            string            `mapstructure:"log_level" yaml:"log_level"`
    RetryAttempts       int               `mapstructure:"retry_attempts" yaml:"retry_attempts"`
    AlertThreshold      int               `mapstructure:"alert_threshold" yaml:"alert_threshold"`
    CustomMessages      map[string]string `mapstructure:"custom_messages" yaml:"custom_messages"`
    SanitizeFields      []string          `mapstructure:"sanitize_fields" yaml:"sanitize_fields"`
}
```

### Enhanced YAML Configuration

```yaml
# Add to your existing config.yaml
error_handling:
  include_stack_trace: true    # false in production
  sanitize_sensitive_data: false # true in production
  log_level: "debug"          # "error" in production
  enable_detailed_logging: true
  enable_metrics: true
  max_errors_per_minute: 100
  alert_on_critical_errors: false # true in production

domain_error_policies:
  account:
    enable_stack_trace: true
    log_level: "info"
    retry_attempts: 3
    alert_threshold: 10
    custom_messages:
      authentication_failed: "Please check your credentials and try again"
      account_locked: "Your account has been temporarily locked for security"
    sanitize_fields: ["password", "token", "secret"]
    
  auth:
    enable_stack_trace: false
    log_level: "warn"
    retry_attempts: 1
    alert_threshold: 5
    custom_messages:
      invalid_token: "Session expired, please login again"
    sanitize_fields: ["jwt_token", "refresh_token"]
    
  admin:
    enable_stack_trace: true
    log_level: "debug"
    retry_attempts: 2
    alert_threshold: 3
    custom_messages:
      insufficient_privileges: "Administrator access required"
    sanitize_fields: ["admin_key", "system_token"]
    
  system:
    enable_stack_trace: true
    log_level: "error"
    retry_attempts: 5
    alert_threshold: 1
    custom_messages:
      database_error: "System temporarily unavailable, please try again"
    sanitize_fields: ["db_password", "api_key", "connection_string"]

---
# Production overrides
environment: "production"

error_handling:
  include_stack_trace: false
  sanitize_sensitive_data: true
  log_level: "error"
  enable_detailed_logging: false
  alert_on_critical_errors: true
  max_errors_per_minute: 50

domain_error_policies:
  account:
    enable_stack_trace: false
    log_level: "warn"
    alert_threshold: 5
```

## 2. Unified Manager Implementation

### Main Integration Manager

```go
package integration

import (
    "context"
    "fmt"
    "sync"
    
    "your-app/internal/error_custom"
    "your-app/utils/config"
)

type UnifiedManager struct {
    configManager    *config.ConfigManager
    errorHandler     *error_custom.UnifiedErrorHandler
    domainPolicies   map[string]*DomainErrorPolicy
    mu              sync.RWMutex
    initialized     bool
}

func NewUnifiedManager() *UnifiedManager {
    return &UnifiedManager{
        domainPolicies: make(map[string]*DomainErrorPolicy),
    }
}

func (um *UnifiedManager) Initialize(ctx context.Context, configPath string) error {
    um.mu.Lock()
    defer um.mu.Unlock()
    
    // Initialize configuration system
    um.configManager = config.NewConfigManager()
    if err := um.configManager.Load(ctx, configPath); err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Initialize error handling system
    um.errorHandler = error_custom.NewUnifiedErrorHandler()
    
    // Configure error handling based on config
    if err := um.configureErrorHandling(); err != nil {
        return fmt.Errorf("failed to configure error handling: %w", err)
    }
    
    // Set up config change callbacks
    um.configManager.RegisterCallback(um.onConfigChange)
    
    // Start watching for config changes
    go func() {
        if err := um.configManager.Watch(ctx); err != nil {
            // Log error but don't stop the application
            fmt.Printf("Config watcher stopped: %v\n", err)
        }
    }()
    
    um.initialized = true
    return nil
}

func (um *UnifiedManager) configureErrorHandling() error {
    cfg := um.configManager.GetConfig()
    
    // Configure global error handling
    um.configureGlobalErrorSettings(cfg)
    
    // Configure domain-specific policies
    for domain, policy := range cfg.DomainErrorPolicies {
        um.configureDomainPolicy(domain, policy)
    }
    
    return nil
}

func (um *UnifiedManager) onConfigChange(oldConfig, newConfig *config.Config) error {
    um.mu.Lock()
    defer um.mu.Unlock()
    
    fmt.Println("Configuration changed, updating error handling...")
    return um.configureErrorHandling()
}
```

## 3. Configuration-Aware Error Handler

### Enhanced Error Handler with Config Integration

```go
type ConfigAwareErrorHandler struct {
    *error_custom.UnifiedErrorHandler
    config         *config.Config
    domainPolicies map[string]config.DomainErrorPolicy
    mu            sync.RWMutex
}

func NewConfigAwareErrorHandler(cfg *config.Config) *ConfigAwareErrorHandler {
    return &ConfigAwareErrorHandler{
        UnifiedErrorHandler: error_custom.NewUnifiedErrorHandler(),
        config:             cfg,
        domainPolicies:     cfg.DomainErrorPolicies,
    }
}

func (ceh *ConfigAwareErrorHandler) HandleHTTPError(w http.ResponseWriter, r *http.Request, err error) {
    ceh.mu.RLock()
    defer ceh.mu.RUnlock()
    
    domain := ceh.getDomainFromRequest(r)
    
    // Get domain-specific policy
    policy, exists := ceh.domainPolicies[domain]
    if !exists {
        // Use default error handling
        ceh.UnifiedErrorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    // Apply configuration-based error handling
    apiErr := ceh.processErrorWithPolicy(err, domain, policy)
    
    // Log based on policy
    ceh.logErrorWithPolicy(apiErr, policy)
    
    // Respond with configured behavior
    ceh.respondWithPolicy(w, r, apiErr, policy)
}

func (ceh *ConfigAwareErrorHandler) processErrorWithPolicy(err error, domain string, policy config.DomainErrorPolicy) *error_custom.APIError {
    apiErr := error_custom.ConvertToAPIError(err)
    
    // Apply custom messages
    if customMsg, exists := policy.CustomMessages[apiErr.Code]; exists {
        apiErr.Message = customMsg
    }
    
    // Sanitize sensitive data based on policy
    if ceh.config.ErrorHandling.SanitizeSensitiveData {
        apiErr = ceh.sanitizeAPIError(apiErr, policy.SanitizeFields)
    }
    
    // Add/remove stack trace based on policy
    if !policy.EnableStackTrace {
        apiErr.Details = ceh.removeStackTrace(apiErr.Details)
    }
    
    return apiErr
}
```

## 4. Domain-Aware Error Factory

### Configuration-Driven Error Creation

```go
type ConfigAwareErrorFactory struct {
    config *config.Config
    mu     sync.RWMutex
}

func NewConfigAwareErrorFactory(cfg *config.Config) *ConfigAwareErrorFactory {
    return &ConfigAwareErrorFactory{config: cfg}
}

func (cef *ConfigAwareErrorFactory) NewUserNotFoundError(userID int64) error {
    cef.mu.RLock()
    defer cef.mu.RUnlock()
    
    domain := error_custom.DomainAccount
    
    // Check if domain is enabled
    if !cef.config.IsDomainEnabled(domain) {
        // Return generic system error if domain disabled
        return error_custom.NewSystemError(error_custom.DomainSystem, 
            "user_service", "get_user", "Resource not available", nil)
    }
    
    // Get domain-specific policy
    policy, exists := cef.config.DomainErrorPolicies[domain]
    if exists {
        if customMsg, hasCustom := policy.CustomMessages["user_not_found"]; hasCustom {
            return cef.createCustomNotFoundError(domain, customMsg, userID)
        }
    }
    
    // Use standard error
    return error_custom.NewUserNotFoundByID(userID)
}

func (cef *ConfigAwareErrorFactory) NewAuthenticationError(reason string) error {
    domain := error_custom.DomainAuth
    
    // Check configuration for auth domain
    if policy, exists := cef.config.DomainErrorPolicies[domain]; exists {
        if customMsg, hasCustom := policy.CustomMessages["authentication_failed"]; hasCustom {
            return error_custom.NewAuthenticationErrorWithContext(domain, customMsg, 
                map[string]interface{}{
                    "original_reason": reason,
                    "retry_attempts": policy.RetryAttempts,
                })
        }
    }
    
    return error_custom.NewAuthenticationError(domain, reason)
}
```

## 5. Middleware Integration

### Combined Middleware Setup

```go
func SetupMiddleware(r *chi.Mux, unifiedManager *UnifiedManager) {
    config := unifiedManager.GetConfig()
    errorHandler := unifiedManager.GetErrorHandler()
    
    // Request ID middleware (always first)
    r.Use(middleware.RequestID)
    
    // Domain detection middleware
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            domain := detectDomainFromPath(req.URL.Path)
            
            // Check if domain is enabled in config
            if !config.IsDomainEnabled(domain) {
                errorHandler.HandleHTTPError(w, req, 
                    error_custom.NewBusinessLogicError(error_custom.DomainSystem, 
                        "domain_disabled", "This service is currently unavailable"))
                return
            }
            
            // Add domain to context
            ctx := context.WithValue(req.Context(), "domain", domain)
            next.ServeHTTP(w, req.WithContext(ctx))
        })
    })
    
    // Configuration-aware recovery middleware
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    domain := getDomainFromContext(req.Context())
                    
                    // Create panic error based on config
                    var panicErr error
                    if config.ErrorHandling.IncludeStackTrace {
                        panicErr = error_custom.NewSystemError(domain, "panic_handler", 
                            "request_processing", fmt.Sprintf("Panic: %v", err), nil)
                    } else {
                        panicErr = error_custom.NewSystemError(domain, "panic_handler", 
                            "request_processing", "Internal server error", nil)
                    }
                    
                    errorHandler.HandleHTTPError(w, req, panicErr)
                }
            }()
            next.ServeHTTP(w, req)
        })
    })
    
    // Rate limiting based on config
    if config.IsRateLimitEnabled() {
        r.Use(rateLimitMiddleware(config))
    }
}
```

## 6. Service Layer Integration

### Configuration-Aware Service Base

```go
type BaseService struct {
    config       *config.Config
    errorHandler *ConfigAwareErrorHandler
    errorFactory *ConfigAwareErrorFactory
}

func NewBaseService(cfg *config.Config) *BaseService {
    return &BaseService{
        config:       cfg,
        errorHandler: NewConfigAwareErrorHandler(cfg),
        errorFactory: NewConfigAwareErrorFactory(cfg),
    }
}

func (bs *BaseService) HandleError(domain string, err error) error {
    // Get retry configuration from domain policy
    if policy, exists := bs.config.DomainErrorPolicies[domain]; exists {
        if policy.RetryAttempts > 0 && error_custom.IsRetryableError(err) {
            return bs.wrapRetryableError(err, domain, policy.RetryAttempts)
        }
    }
    
    return bs.errorHandler.HandleError(domain, err)
}

func (bs *BaseService) ValidateBusinessRules(domain string, validations map[string]func() error) error {
    // Use configuration to determine which validations to run
    enabledDomains := bs.config.GetEnabledDomains()
    
    // Skip validation if domain not enabled
    if !contains(enabledDomains, domain) {
        return nil
    }
    
    return bs.errorHandler.ValidateBusinessRules(domain, validations)
}
```

## 7. Complete Usage Example

### Application Initialization

```go
func main() {
    ctx := context.Background()
    
    // Initialize unified manager
    unifiedManager := NewUnifiedManager()
    if err := unifiedManager.Initialize(ctx, "./config/config.yaml"); err != nil {
        log.Fatal("Failed to initialize unified manager:", err)
    }
    
    // Setup router with integrated middleware
    r := chi.NewRouter()
    SetupMiddleware(r, unifiedManager)
    
    // Setup handlers with config-aware error handling
    userHandler := NewUserHandler(unifiedManager)
    r.Route("/api/v1/users", func(r chi.Router) {
        r.Get("/{id}", userHandler.GetUser)
        r.Post("/", userHandler.CreateUser)
    })
    
    // Start server
    config := unifiedManager.GetConfig()
    serverAddr := config.GetServerAddress()
    
    log.Printf("Starting server on %s", serverAddr)
    log.Fatal(http.ListenAndServe(serverAddr, r))
}
```

### Handler Implementation

```go
type UserHandler struct {
    userService  *UserService
    errorHandler *ConfigAwareErrorHandler
    config       *config.Config
}

func NewUserHandler(um *UnifiedManager) *UserHandler {
    return &UserHandler{
        userService:  NewUserService(um.GetConfig()),
        errorHandler: NewConfigAwareErrorHandler(um.GetConfig()),
        config:       um.GetConfig(),
    }
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // Parse ID with domain-aware validation
    userID, err := h.errorHandler.ParseIDParam(r, "id", error_custom.DomainAccount)
    if err != nil {
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    // Call service with configuration context
    user, err := h.userService.GetUserByID(r.Context(), userID)
    if err != nil {
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    h.errorHandler.RespondWithSuccess(w, r, user)
}
```

## 8. Hot Reload Support

### Configuration Change Handling

```go
func (um *UnifiedManager) onConfigChange(oldConfig, newConfig *config.Config) error {
    // Compare error handling configuration
    if um.errorHandlingChanged(oldConfig, newConfig) {
        log.Println("Error handling configuration changed, updating...")
        
        // Update error handler with new configuration
        um.errorHandler = NewConfigAwareErrorHandler(newConfig)
        
        // Update domain policies
        um.updateDomainPolicies(newConfig.DomainErrorPolicies)
        
        // Notify all components of the change
        um.notifyConfigChange(newConfig)
    }
    
    return nil
}

func (um *UnifiedManager) errorHandlingChanged(old, new *config.Config) bool {
    return old.ErrorHandling.IncludeStackTrace != new.ErrorHandling.IncludeStackTrace ||
           old.ErrorHandling.SanitizeSensitiveData != new.ErrorHandling.SanitizeSensitiveData ||
           old.ErrorHandling.LogLevel != new.ErrorHandling.LogLevel ||
           !reflect.DeepEqual(old.DomainErrorPolicies, new.DomainErrorPolicies)
}
```

## 9. Testing Integration

### Test Helper Functions

```go
func SetupTestEnvironment() *UnifiedManager {
    // Set test environment
    os.Setenv("APP_ENV", "testing")
    
    um := NewUnifiedManager()
    ctx := context.Background()
    
    // Initialize with test config
    if err := um.Initialize(ctx, "./config/test_config.yaml"); err != nil {
        panic(fmt.Sprintf("Failed to initialize test environment: %v", err))
    }
    
    return um
}

func TestErrorHandlingIntegration(t *testing.T) {
    um := SetupTestEnvironment()
    defer um.Stop()
    
    config := um.GetConfig()
    errorHandler := NewConfigAwareErrorHandler(config)
    
    // Test domain-specific error handling
    err := error_custom.NewUserNotFoundByID(123)
    
    // Verify configuration affects error processing
    apiErr := errorHandler.processErrorWithPolicy(err, "account", 
        config.DomainErrorPolicies["account"])
    
    assert.NotNil(t, apiErr)
    assert.Equal(t, "account", apiErr.Domain)
}
```

## 10. Benefits of Integration

### Key Advantages

1. **Unified Configuration**: Single source of truth for both app config and error policies
2. **Environment Awareness**: Different error behaviors for dev/staging/production
3. **Hot Reload**: Dynamic updates to error handling without restart
4. **Domain Consistency**: Consistent error handling across all domains
5. **Simplified Management**: One initialization, one configuration file
6. **Type Safety**: Compile-time validation of configuration structure
7. **Testing Support**: Easy to mock and test different configurations

### Performance Benefits

- **Reduced Memory Usage**: Shared configuration across components
- **Faster Error Processing**: Pre-configured policies reduce runtime decisions
- **Efficient Hot Reload**: Only affected components are updated
- **Optimized Logging**: Configuration-driven log levels prevent unnecessary processing

## Conclusion

This integration creates a powerful, unified system that:
- Centralizes configuration management
- Provides consistent, domain-aware error handling
- Supports hot reloading of both config and error policies
- Scales from development to production environments
- Maintains type safety and performance

The combined system eliminates configuration drift, reduces boilerplate code, and provides a solid foundation for enterprise-grade Go applications.
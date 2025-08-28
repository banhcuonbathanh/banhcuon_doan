# ShopEasy Logger Integration Guide

## 📋 Integration Overview

This guide shows how to integrate the comprehensive logger system into your existing ShopEasy project, replacing basic console output with structured, environment-aware logging.

## 📁 Step 1: Add Logger Package to Project

### 1.1 Update Project Structure
```bash
# Your updated structure will be:
shopeasy-app/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── errors/           # Custom error system
│   ├── handlers/         # HTTP handlers
│   ├── services/         # Business logic
│   ├── integration/      # Combined config+error system
│   └── logger/           # ⭐ NEW: Logger system
├── configs/              # Configuration files
├── pkg/                  # Shared packages
└── tests/                # Test files
```

### 1.2 Create Logger Package Files
Create these files in `internal/logger/`:

**internal/logger/logger_types.go**
```go
// logger/types.go - Core types, constants, and structures
package logger

import (
    "log"
    "sync"
)

// Logger levels with numeric values for comparison
const (
    DebugLevel = iota
    InfoLevel
    WarningLevel
    ErrorLevel
    FatalLevel
)

var levelNames = map[int]string{
    DebugLevel:   "DEBUG",
    InfoLevel:    "INFO",
    WarningLevel: "WARN",
    ErrorLevel:   "ERROR",
    FatalLevel:   "FATAL",
}

// Output formats
const (
    FormatJSON   = "json"
    FormatText   = "text"
    FormatPretty = "pretty"
)

// Layer constants for better organization
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

// LogEntry represents a structured log entry with enhanced metadata
type LogEntry struct {
    Timestamp    string                 `json:"timestamp"`
    Level        string                 `json:"level"`
    Message      string                 `json:"message"`
    Context      map[string]interface{} `json:"context,omitempty"`
    File         string                 `json:"file,omitempty"`
    Function     string                 `json:"function,omitempty"`
    Line         int                    `json:"line,omitempty"`
    RequestID    string                 `json:"request_id,omitempty"`
    UserID       string                 `json:"user_id,omitempty"`
    SessionID    string                 `json:"session_id,omitempty"`
    TraceID      string                 `json:"trace_id,omitempty"`
    Component    string                 `json:"component,omitempty"`
    Operation    string                 `json:"operation,omitempty"`
    Duration     int64                  `json:"duration_ms,omitempty"`
    ErrorCode    string                 `json:"error_code,omitempty"`
    Environment  string                 `json:"environment,omitempty"`
    Cause        string                 `json:"cause,omitempty"`
    Layer        string                 `json:"layer,omitempty"`
}

// Logger structure with enhanced capabilities and thread safety
type Logger struct {
    debugLogger   *log.Logger
    infoLogger    *log.Logger
    warningLogger *log.Logger
    errorLogger   *log.Logger
    fatalLogger   *log.Logger
    outputFormat  string
    enableDebug   bool
    minLevel      int
    environment   string
    component     string
    layer         string
    operation     string
    mutex         sync.RWMutex
    contextFields map[string]interface{}
}
```

## 📊 Step 2: Update Configuration Files

### 2.1 Update Development Configuration
Update `configs/development.yaml` to include logger settings:

```yaml
app:
  name: "ShopEasy"
  environment: "development"
  version: "1.0.0"
  debug: true

server:
  host: "localhost"
  port: 8080
  read_timeout: 30
  write_timeout: 30

database:
  host: "localhost"
  port: 5432
  name: "shopeasy_dev"
  user: "developer"
  password: "devpass123"

# ⭐ NEW: Logger configuration
logging:
  format: "pretty"              # pretty, json, text
  level: "debug"                # debug, info, warn, error, fatal
  enable_debug: true
  enable_caller_info: true
  component: "shopeasy-api"
  global_fields:
    service: "shopeasy"
    version: "1.0.0"

error_handling:
  include_stack_trace: true
  sanitize_sensitive_data: false
  log_level: "debug"
  enable_detailed_logging: true
  enable_metrics: true
  max_errors_per_minute: 100
  alert_on_critical_errors: false

domain_error_policies:
  account:
    enable_stack_trace: true
    log_level: "debug"
    retry_attempts: 2
    alert_threshold: 10
    custom_messages:
      USER_NOT_FOUND: "We couldn't find an account with that information"
      EMAIL_ALREADY_EXISTS: "An account with this email already exists"
      VALIDATION_FAILED: "Please check your input and try again"
    sanitize_fields: ["password", "credit_card"]
    
  auth:
    enable_stack_trace: true
    log_level: "info"
    retry_attempts: 1
    alert_threshold: 5
    custom_messages:
      AUTHENTICATION_FAILED: "Invalid email or password"
    sanitize_fields: ["password", "token"]
    
  payment:
    enable_stack_trace: false
    log_level: "warn"
    retry_attempts: 3
    alert_threshold: 5
    custom_messages:
      PAYMENT_DECLINED: "Your payment was declined. Please try a different payment method"
    sanitize_fields: ["card_number", "cvv", "bank_account"]
    
  cart:
    enable_stack_trace: true
    log_level: "info"
    retry_attempts: 1
    alert_threshold: 20
    custom_messages:
      ITEM_OUT_OF_STOCK: "Sorry, this item is currently out of stock"
    sanitize_fields: []
    
  admin:
    enable_stack_trace: true
    log_level: "debug"
    retry_attempts: 1
    alert_threshold: 3
    custom_messages:
      INSUFFICIENT_PRIVILEGES: "Administrator access required"
    sanitize_fields: ["admin_token", "api_key"]
```

### 2.2 Update Production Configuration
Update `configs/production.yaml`:

```yaml
app:
  name: "ShopEasy"
  environment: "production"
  version: "1.0.0"
  debug: false

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30
  write_timeout: 30

database:
  host: "prod-db.company.com"
  port: 5432
  name: "shopeasy_prod"
  user: "produser"
  password: "${DB_PASSWORD}"

# ⭐ Production Logger Settings
logging:
  format: "json"                # JSON format for production logging
  level: "info"                 # Only info and above in production
  enable_debug: false
  enable_caller_info: false     # No caller info in production
  component: "shopeasy-api"
  global_fields:
    service: "shopeasy"
    version: "1.0.0"
    environment: "production"

error_handling:
  include_stack_trace: false
  sanitize_sensitive_data: true
  log_level: "error"
  enable_detailed_logging: false
  enable_metrics: true
  max_errors_per_minute: 50
  alert_on_critical_errors: true

# ... rest of domain_error_policies with production-safe settings
```

## 🔧 Step 3: Update Config Structure

### 3.1 Update Config Types
Update `internal/config/config.go` to include logging configuration:

```go
package config

import (
    "context"
    "fmt"
    "path/filepath"
    "sync"
    
    "github.com/fsnotify/fsnotify"
    "github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
    App      AppConfig      `mapstructure:"app" yaml:"app"`
    Server   ServerConfig   `mapstructure:"server" yaml:"server"`
    Database DatabaseConfig `mapstructure:"database" yaml:"database"`
    Logging  LoggingConfig  `mapstructure:"logging" yaml:"logging"`  // ⭐ NEW
    
    // Error handling configuration
    ErrorHandling       ErrorHandlingConfig            `mapstructure:"error_handling" yaml:"error_handling"`
    DomainErrorPolicies map[string]DomainErrorPolicy   `mapstructure:"domain_error_policies" yaml:"domain_error_policies"`
}

// ⭐ NEW: Logging configuration
type LoggingConfig struct {
    Format           string                 `mapstructure:"format" yaml:"format"`
    Level            string                 `mapstructure:"level" yaml:"level"`
    EnableDebug      bool                   `mapstructure:"enable_debug" yaml:"enable_debug"`
    EnableCallerInfo bool                   `mapstructure:"enable_caller_info" yaml:"enable_caller_info"`
    Component        string                 `mapstructure:"component" yaml:"component"`
    GlobalFields     map[string]interface{} `mapstructure:"global_fields" yaml:"global_fields"`
}

// ... rest of existing types remain the same
```

### 3.2 Update Config Defaults
Update the `setDefaults()` method in `internal/config/config.go`:

```go
func (cm *ConfigManager) setDefaults() {
    // App defaults
    cm.viper.SetDefault("app.name", "ShopEasy")
    cm.viper.SetDefault("app.environment", "development")
    cm.viper.SetDefault("app.version", "1.0.0")
    cm.viper.SetDefault("app.debug", true)
    
    // Server defaults
    cm.viper.SetDefault("server.host", "localhost")
    cm.viper.SetDefault("server.port", 8080)
    cm.viper.SetDefault("server.read_timeout", 30)
    cm.viper.SetDefault("server.write_timeout", 30)
    
    // ⭐ NEW: Logging defaults
    cm.viper.SetDefault("logging.format", "pretty")
    cm.viper.SetDefault("logging.level", "debug")
    cm.viper.SetDefault("logging.enable_debug", true)
    cm.viper.SetDefault("logging.enable_caller_info", true)
    cm.viper.SetDefault("logging.component", "shopeasy-api")
    
    // Error handling defaults
    cm.viper.SetDefault("error_handling.include_stack_trace", true)
    cm.viper.SetDefault("error_handling.sanitize_sensitive_data", false)
    cm.viper.SetDefault("error_handling.log_level", "debug")
    cm.viper.SetDefault("error_handling.enable_detailed_logging", true)
    cm.viper.SetDefault("error_handling.max_errors_per_minute", 100)
}
```

## 🔗 Step 4: Update Integration Layer

### 4.1 Update Unified Manager
Update `internal/integration/manager.go`:

```go
package integration

import (
    "context"
    "fmt"
    "sync"
    
    "github.com/yourusername/shopeasy-app/internal/config"
    "github.com/yourusername/shopeasy-app/internal/errors"
    "github.com/yourusername/shopeasy-app/internal/logger"  // ⭐ NEW
)

type UnifiedManager struct {
    configManager *config.ConfigManager
    errorHandler  *ConfigAwareErrorHandler
    logger        *logger.Logger  // ⭐ NEW
    mu            sync.RWMutex
    initialized   bool
}

func NewUnifiedManager() *UnifiedManager {
    return &UnifiedManager{}
}

func (um *UnifiedManager) Initialize(ctx context.Context, configPath string) error {
    um.mu.Lock()
    defer um.mu.Unlock()
    
    // Step 1: Initialize basic logger for startup
    um.logger = logger.NewLogger()
    um.logger.Info("🚀 Initializing Unified Manager...")
    
    // Step 2: Initialize configuration
    um.logger.Info("📊 Loading configuration...")
    um.configManager = config.NewConfigManager()
    if err := um.configManager.Load(ctx, configPath); err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Step 3: Configure logger with loaded config
    um.logger.Info("🔧 Configuring logger...")
    cfg := um.configManager.GetConfig()
    um.configureLogger(cfg)
    
    // Step 4: Initialize error handler with config
    um.logger.Info("🚨 Setting up error handling...")
    um.errorHandler = NewConfigAwareErrorHandler(cfg, um.logger)
    
    // Step 5: Register for config changes
    um.logger.Info("👂 Setting up config change listener...")
    um.configManager.RegisterCallback(um.onConfigChange)
    
    // Step 6: Start config watcher (non-blocking)
    go func() {
        um.logger.Info("👀 Starting config watcher...")
        if err := um.configManager.Watch(ctx); err != nil {
            um.logger.Warning("⚠️ Config watcher stopped", map[string]interface{}{
                "error": err.Error(),
            })
        }
    }()
    
    um.initialized = true
    um.logger.Info("✅ Unified Manager initialized successfully!")
    
    return nil
}

// ⭐ NEW: Configure logger based on config
func (um *UnifiedManager) configureLogger(cfg *config.Config) {
    // Set format
    um.logger.SetOutputFormat(cfg.Logging.Format)
    
    // Set log level
    level := logger.InfoLevel
    switch cfg.Logging.Level {
    case "debug":
        level = logger.DebugLevel
    case "info":
        level = logger.InfoLevel
    case "warn":
        level = logger.WarningLevel
    case "error":
        level = logger.ErrorLevel
    case "fatal":
        level = logger.FatalLevel
    }
    um.logger.SetMinLevel(level)
    
    // Set debug flag
    um.logger.SetDebugLogging(cfg.Logging.EnableDebug)
    
    // Set component
    um.logger.SetComponent(cfg.Logging.Component)
    
    // Add global fields
    for key, value := range cfg.Logging.GlobalFields {
        um.logger.AddGlobalField(key, value)
    }
    
    // Add environment
    um.logger.AddGlobalField("environment", cfg.App.Environment)
}

func (um *UnifiedManager) onConfigChange(oldConfig, newConfig *config.Config) error {
    um.logger.Info("🔄 Configuration changed, updating systems...")
    
    um.mu.Lock()
    defer um.mu.Unlock()
    
    // Update logger with new config
    um.configureLogger(newConfig)
    
    // Update error handler with new config
    um.errorHandler = NewConfigAwareErrorHandler(newConfig, um.logger)
    
    um.logger.Info("✅ All systems updated with new configuration")
    return nil
}

func (um *UnifiedManager) GetConfig() *config.Config {
    um.mu.RLock()
    defer um.mu.RUnlock()
    return um.configManager.GetConfig()
}

func (um *UnifiedManager) GetErrorHandler() *ConfigAwareErrorHandler {
    um.mu.RLock()
    defer um.mu.RUnlock()
    return um.errorHandler
}

// ⭐ NEW: Get logger instance
func (um *UnifiedManager) GetLogger() *logger.Logger {
    um.mu.RLock()
    defer um.mu.RUnlock()
    return um.logger
}

func (um *UnifiedManager) IsInitialized() bool {
    um.mu.RLock()
    defer um.mu.RUnlock()
    return um.initialized
}
```

### 4.2 Update Error Handler
Update `internal/integration/error_handler.go`:

```go
package integration

import (
    "encoding/json"
    "fmt"
    "net/http"
    "runtime"
    "strings"
    "sync"
    
    "github.com/yourusername/shopeasy-app/internal/config"
    "github.com/yourusername/shopeasy-app/internal/errors"
    "github.com/yourusername/shopeasy-app/internal/logger"  // ⭐ NEW
)

type ConfigAwareErrorHandler struct {
    config *config.Config
    logger *logger.Logger  // ⭐ NEW
    mu     sync.RWMutex
}

func NewConfigAwareErrorHandler(cfg *config.Config, log *logger.Logger) *ConfigAwareErrorHandler {
    return &ConfigAwareErrorHandler{
        config: cfg,
        logger: log,  // ⭐ NEW
    }
}

func (ceh *ConfigAwareErrorHandler) HandleHTTPError(w http.ResponseWriter, r *http.Request, err error) {
    ceh.mu.RLock()
    defer ceh.mu.RUnlock()
    
    // Convert to API error
    apiErr := errors.ConvertToAPIError(err)
    
    // Get domain from error or request
    domain := ceh.getDomainFromError(apiErr, r)
    
    // Apply domain-specific processing
    processedErr := ceh.processErrorWithPolicy(apiErr, domain)
    
    // ⭐ NEW: Use structured logging instead of fmt.Printf
    ceh.logError(processedErr, r, domain)
    
    // Send response
    ceh.sendErrorResponse(w, r, processedErr)
}

// ⭐ UPDATED: Enhanced error logging with structured logger
func (ceh *ConfigAwareErrorHandler) logError(apiErr *errors.APIError, r *http.Request, domain string) {
    // Create structured context
    context := map[string]interface{}{
        "error_code":   apiErr.Code,
        "domain":       domain,
        "http_method":  r.Method,
        "path":         r.URL.Path,
        "user_agent":   r.UserAgent(),
        "remote_addr":  r.RemoteAddr,
    }
    
    // Add error details if available
    if len(apiErr.Details) > 0 {
        context["error_details"] = apiErr.Details
    }
    
    // Add request ID if available
    if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
        context["request_id"] = requestID
    }
    
    // Get domain policy for logging level
    logLevel := ceh.config.ErrorHandling.LogLevel
    if policy, exists := ceh.config.DomainErrorPolicies[domain]; exists {
        logLevel = policy.LogLevel
    }
    
    // Create logger with appropriate layer
    errorLogger := ceh.logger
    errorLogger.SetLayer(logger.LayerHandler)
    errorLogger.SetOperation(fmt.Sprintf("handle_%s_error", domain))
    
    message := fmt.Sprintf("API Error: %s", apiErr.Message)
    
    // Log based on configured level
    switch strings.ToLower(logLevel) {
    case "debug":
        errorLogger.Debug(message, context)
    case "info":
        errorLogger.Info(message, context)
    case "warn":
        errorLogger.Warning(message, context)
    case "error":
        errorLogger.Error(message, context)
    }
}

// ... rest of the methods remain mostly the same, but replace fmt.Printf with logger calls

func (ceh *ConfigAwareErrorHandler) RespondWithSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    response := map[string]interface{}{
        "success": true,
        "data":    data,
    }
    
    // ⭐ NEW: Log successful responses
    ceh.logger.InfoWithOperation(
        "API Success Response",
        logger.LayerHandler,
        "respond_success",
        map[string]interface{}{
            "http_method": r.Method,
            "path":        r.URL.Path,
            "status_code": http.StatusOK,
        },
    )
    
    json.NewEncoder(w).Encode(response)
}
```

## 🛠️ Step 5: Update Services

### 5.1 Update User Service
Update `internal/services/user_service.go`:

```go
package services

import (
    "context"
    
    "github.com/yourusername/shopeasy-app/internal/config"
    "github.com/yourusername/shopeasy-app/internal/errors"
    "github.com/yourusername/shopeasy-app/internal/logger"  // ⭐ NEW
)

type User struct {
    ID    int64  `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}

type CreateUserRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
}

type UserService struct {
    config *config.Config
    logger *logger.Logger  // ⭐ NEW
    users  map[string]*User
}

func NewUserService(cfg *config.Config, log *logger.Logger) *UserService {
    serviceLogger := logger.NewServiceLogger()  // ⭐ Create service-specific logger
    serviceLogger.SetComponent("user-service")
    
    return &UserService{
        config: cfg,
        logger: serviceLogger,  // ⭐ NEW
        users:  make(map[string]*User),
    }
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // ⭐ NEW: Structured logging for service operations
    s.logger.InfoWithOperation(
        "Creating new user",
        logger.LayerService,
        "create_user",
        map[string]interface{}{
            "email": logger.maskEmail(req.Email),  // Use logger's email masking
        },
    )
    
    // Validate input
    if req.Email == "" {
        s.logger.LogValidationError("email", "Email is required", req.Email)
        return nil, errors.NewValidationError(errors.DomainAccount, "Email is required")
    }
    if req.Password == "" {
        s.logger.LogValidationError("password", "Password is required", nil)
        return nil, errors.NewValidationError(errors.DomainAccount, "Password is required")
    }
    if len(req.Password) < 8 {
        s.logger.LogValidationError("password", "Password must be at least 8 characters", len(req.Password))
        return nil, errors.NewValidationError(errors.DomainAccount, "Password must be at least 8 characters")
    }
    if req.Name == "" {
        s.logger.LogValidationError("name", "Name is required", req.Name)
        return nil, errors.NewValidationError(errors.DomainAccount, "Name is required")
    }
    
    // Check if email already exists
    if _, exists := s.users[req.Email]; exists {
        s.logger.Warning("User creation failed - email already exists", map[string]interface{}{
            "email":     logger.maskEmail(req.Email),
            "operation": "create_user",
            "cause":     "duplicate_email",
        })
        return nil, errors.NewEmailExistsError(errors.DomainAccount, req.Email)
    }
    
    // Create user (simulate database operation)
    user := &User{
        ID:    int64(len(s.users) + 1),
        Email: req.Email,
        Name:  req.Name,
    }
    
    s.users[req.Email] = user
    
    // ⭐ NEW: Log successful user creation
    s.logger.LogUserActivity(
        fmt.Sprintf("%d", user.ID),
        user.Email,
        "user_created",
        "user_account",
        map[string]interface{}{
            "user_id": user.ID,
            "name":    user.Name,
        },
    )
    
    return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*User, error) {
    s.logger.InfoWithOperation(
        "Retrieving user by ID",
        logger.LayerService,
        "get_user_by_id",
        map[string]interface{}{
            "user_id": userID,
        },
    )
    
    for _, user := range s.users {
        if user.ID == userID {
            s.logger.Debug("User found", map[string]interface{}{
                "user_id": userID,
                "email":   logger.maskEmail(user.Email),
            })
            return user, nil
        }
    }
    
    s.logger.Warning("User not found", map[string]interface{}{
        "user_id":   userID,
        "operation": "get_user_by_id",
        "cause":     "user_not_found",
    })
    
    return nil, errors.NewUserNotFoundError(errors.DomainAccount, userID)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
    s.logger.InfoWithOperation(
        "Retrieving user by email",
        logger.LayerService,
        "get_user_by_email",
        map[string]interface{}{
            "email": logger.maskEmail(email),
        },
    )
    
    user, exists := s.users[email]
    if !exists {
        s.logger.Warning("User not found by email", map[string]interface{}{
            "email":     logger.maskEmail(email),
            "operation": "get_user_by_email",
            "cause":     "user_not_found",
        })
        return nil, errors.NewUserNotFoundError(errors.DomainAccount, email)
    }
    
    return user, nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
    // ⭐ NEW: Log authentication attempt
    s.logger.LogAuthAttempt(email, false, "starting_authentication")
    
    // In real app: hash comparison, etc.
    user, exists := s.users[email]
    if !exists {
        s.logger.LogAuthAttempt(email, false, "user_not_found")
        return nil, errors.NewAuthenticationError(errors.DomainAuth, "invalid_credentials")
    }
    
    // Simulate password validation (always fails for demo)
    if password != "correct_password" {
        s.logger.LogAuthAttempt(email, false, "invalid_password")
        return nil, errors.NewAuthenticationError(errors.DomainAuth, "invalid_password")
    }
    
    // ⭐ NEW: Log successful authentication
    s.logger.LogAuthAttempt(email, true, "authentication_successful")
    
    return user, nil
}
```

### 5.2 Update Payment Service
Update `internal/services/payment_service.go`:

```go
package services

import (
    "context"
    "fmt"
    "math/rand"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/config"
    "github.com/yourusername/shopeasy-app/internal/errors"
    "github.com/yourusername/shopeasy-app/internal/logger"  // ⭐ NEW
)

type PaymentService struct {
    config *config.Config
    logger *logger.Logger  // ⭐ NEW
}

func NewPaymentService(cfg *config.Config, log *logger.Logger) *PaymentService {
    serviceLogger := logger.NewServiceLogger()
    serviceLogger.SetComponent("payment-service")
    
    return &PaymentService{
        config: cfg,
        logger: serviceLogger,
    }
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error) {
    start := time.Now()
    
    // ⭐ NEW: Log payment processing start
    s.logger.InfoWithOperation(
        "Processing payment",
        logger.LayerService,
        "process_payment",
        map[string]interface{}{
            "amount":   req.Amount,
            "order_id": req.OrderID,
            "card_last_four": s.maskCardNumber(req.CardNumber),
        },
    )
    
    // Validate payment request
    if req.Amount <= 0 {
        s.logger.LogValidationError("amount", "Amount must be greater than 0", req.Amount)
        return nil, errors.NewValidationError(errors.DomainPayment, "Amount must be greater than 0")
    }
    
    if req.CardNumber == "" {
        s.logger.LogValidationError("card_number", "Card number is required", nil)
        return nil, errors.NewValidationError(errors.DomainPayment, "Card number is required")
    }
    
    if req.CVV == "" {
        s.logger.LogValidationError("cvv", "CVV is required", nil)
        return nil, errors.NewValidationError(errors.DomainPayment, "CVV is required")
    }
    
    // Simulate payment processing
    rand.Seed(time.Now().UnixNano())
    
    if rand.Float32() < 0.3 {
        reasons := []string{"insufficient_funds", "card_declined", "expired_card", "invalid_cvv"}
        reason := reasons[rand.Intn(len(reasons))]
        
        // ⭐ NEW: Log payment failure
        s.logger.ErrorWithCause(
            "Payment processing failed",
            reason,
            logger.LayerService,
            "process_payment",
            map[string]interface{}{
                "amount":           req.Amount,
                "order_id":         req.OrderID,
                "failure_reason":   reason,
                "processing_time_ms": time.Since(start).Milliseconds(),
                "card_last_four":   s.maskCardNumber(req.CardNumber),
            },
        )
        
        return nil, errors.NewPaymentError(errors.DomainPayment, reason, req.Amount)
    }
    
    // Successful payment
    result := &PaymentResult{
        TransactionID: generateTransactionID(),
        Status:        "completed",
        Amount:        req.Amount,
        ProcessedAt:   time.Now(),
    }
    
    duration := time.Since(start)
    
    // ⭐ NEW: Log successful payment with performance metrics
    s.logger.Info("Payment processed successfully", map[string]interface{}{
        "transaction_id":     result.TransactionID,
        "amount":            result.Amount,
        "order_id":          req.OrderID,
        "processing_time_ms": duration.Milliseconds(),
        "status":            result.Status,
        "card_last_four":    s.maskCardNumber(req.CardNumber),
    })
    
    // Log performance if slow
    if duration.Milliseconds() > 1000 {
        s.logger.LogPerformance("payment_processing", duration, map[string]interface{}{
            "transaction_id": result.TransactionID,
            "amount":        result.Amount,
        })
    }
    
    return result, nil
}

// ⭐ NEW: Helper method to mask card numbers
func (s *PaymentService) maskCardNumber(cardNumber string) string {
    if len(cardNumber) < 4 {
        return "****"
    }
    return "****-****-****-" + cardNumber[len(cardNumber)-4:]
}


## 🎯 Step 6: Update Handlers

### 6.1 Update User Handler
Update `internal/handlers/user_handler.go`:

```go
package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/yourusername/shopeasy-app/internal/config"
    "github.com/yourusername/shopeasy-app/internal/errors"
    "github.com/yourusername/shopeasy-app/internal/integration"
    "github.com/yourusername/shopeasy-app/internal/logger"  // ⭐ NEW
    "github.com/yourusername/shopeasy-app/internal/services"
)

type UserHandler struct {
    userService  *services.UserService
    errorHandler *integration.ConfigAwareErrorHandler
    logger       *logger.Logger  // ⭐ NEW
    config       *config.Config
}

func NewUserHandler(um *integration.UnifiedManager) *UserHandler {
    handlerLogger := logger.NewHandlerLogger()
    handlerLogger.SetComponent("user-handler")
    
    return &UserHandler{
        userService:  services.NewUserService(um.GetConfig(), um.GetLogger()),
        errorHandler: um.GetErrorHandler(),
        logger:       handlerLogger,  // ⭐ NEW
        config:       um.GetConfig(),
    }
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // ⭐ NEW: Log incoming request
    h.logger.InfoWithOperation(
        "Handling create user request",
        logger.LayerHandler,
        "create_user",
        map[string]interface{}{
            "method":     r.Method,
            "path":       r.URL.Path,
            "user_agent": r.UserAgent(),
            "remote_ip":  r.RemoteAddr,
        },
    )
    
    var req services.CreateUserRequest
    
    // Parse request body
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Warning("Failed to parse request body", map[string]interface{}{
            "error":     err.Error(),
            "operation": "create_user",
            "cause":     "invalid_json",
        })
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    // Create user
    user, err := h.userService.CreateUser(r.Context(), req)
    if err != nil {
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    duration := time.Since(start)
    
    // ⭐ NEW: Log successful response with API request logging
    h.logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, map[string]interface{}{
        "user_id":    user.ID,
        "user_email": req.Email,
        "operation":  "create_user",
    })
    
    // Success response
    h.errorHandler.RespondWithSuccess(w, r, map[string]interface{}{
        "message": "User created successfully",
        "user":    user,
    })
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // Get user ID from URL
    userIDStr := chi.URLParam(r, "id")
    if userIDStr == "" {
        h.logger.Warning("Missing user ID in request", map[string]interface{}{
            "path":      r.URL.Path,
            "operation": "get_user",
            "cause":     "missing_parameter",
        })
        h.errorHandler.HandleHTTPError(w, r, 
            errors.NewValidationError(errors.DomainAccount, "User ID is required"))
        return
    }
    
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        h.logger.Warning("Invalid user ID format", map[string]interface{}{
            "user_id_str": userIDStr,
            "error":       err.Error(),
            "operation":   "get_user",
            "cause":       "invalid_format",
        })
        h.errorHandler.HandleHTTPError(w, r, 
            errors.NewValidationError(errors.DomainAccount, "Invalid user ID format"))
        return
    }
    
    // Get user
    user, err := h.userService.GetUserByID(r.Context(), userID)
    if err != nil {
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    duration := time.Since(start)
    
    // ⭐ NEW: Log successful API request
    h.logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, map[string]interface{}{
        "user_id":   userID,
        "operation": "get_user",
    })
    
    // Success response
    h.errorHandler.RespondWithSuccess(w, r, user)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    // Parse request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Warning("Failed to parse login request", map[string]interface{}{
            "error":     err.Error(),
            "operation": "login_user",
            "cause":     "invalid_json",
        })
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    // ⭐ NEW: Log login attempt (email will be masked in logger)
    h.logger.Info("User login attempt", map[string]interface{}{
        "email":      req.Email,
        "operation":  "login_user",
        "remote_ip":  r.RemoteAddr,
        "user_agent": r.UserAgent(),
    })
    
    // Authenticate user
    user, err := h.userService.AuthenticateUser(r.Context(), req.Email, req.Password)
    if err != nil {
        // This will automatically log the auth failure in the service
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    duration := time.Since(start)
    
    // ⭐ NEW: Log successful login
    h.logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, map[string]interface{}{
        "user_id":   user.ID,
        "email":     req.Email,
        "operation": "login_user",
        "success":   true,
    })
    
    // Success response
    h.errorHandler.RespondWithSuccess(w, r, map[string]interface{}{
        "message": "Login successful",
        "user":    user,
        "token":   "jwt_token_here", // In real app: generate JWT
    })
}
```

### 6.2 Update Payment Handler
Update `internal/handlers/payment_handler.go`:

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/config"
    "github.com/yourusername/shopeasy-app/internal/integration"
    "github.com/yourusername/shopeasy-app/internal/logger"  // ⭐ NEW
    "github.com/yourusername/shopeasy-app/internal/services"
)

type PaymentHandler struct {
    paymentService *services.PaymentService
    errorHandler   *integration.ConfigAwareErrorHandler
    logger         *logger.Logger  // ⭐ NEW
    config         *config.Config
}

func NewPaymentHandler(um *integration.UnifiedManager) *PaymentHandler {
    handlerLogger := logger.NewHandlerLogger()
    handlerLogger.SetComponent("payment-handler")
    
    return &PaymentHandler{
        paymentService: services.NewPaymentService(um.GetConfig(), um.GetLogger()),
        errorHandler:   um.GetErrorHandler(),
        logger:         handlerLogger,  // ⭐ NEW
        config:         um.GetConfig(),
    }
}

func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    var req services.PaymentRequest
    
    // Parse request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Warning("Failed to parse payment request", map[string]interface{}{
            "error":     err.Error(),
            "operation": "process_payment",
            "cause":     "invalid_json",
        })
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    // ⭐ NEW: Log payment processing attempt (with sensitive data masking)
    h.logger.Info("Processing payment request", map[string]interface{}{
        "amount":         req.Amount,
        "order_id":       req.OrderID,
        "card_last_four": maskCardNumber(req.CardNumber),
        "operation":      "process_payment",
        "remote_ip":      r.RemoteAddr,
    })
    
    // Process payment
    result, err := h.paymentService.ProcessPayment(r.Context(), req)
    if err != nil {
        // Payment errors will be handled according to payment domain policy:
        // - No stack traces (even in dev)
        // - Custom user-friendly messages
        // - Sensitive fields (card_number, cvv) sanitized
        // - Logged at error level
        h.errorHandler.HandleHTTPError(w, r, err)
        return
    }
    
    duration := time.Since(start)
    
    // ⭐ NEW: Log successful payment processing
    h.logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, map[string]interface{}{
        "transaction_id": result.TransactionID,
        "amount":         result.Amount,
        "order_id":       req.OrderID,
        "operation":      "process_payment",
        "status":         result.Status,
    })
    
    // Success response
    h.errorHandler.RespondWithSuccess(w, r, map[string]interface{}{
        "message": "Payment processed successfully",
        "result":  result,
    })
}

// ⭐ NEW: Helper function to mask card numbers
func maskCardNumber(cardNumber string) string {
    if len(cardNumber) < 4 {
        return "****"
    }
    return "****-****-****-" + cardNumber[len(cardNumber)-4:]
}
```

## 🛣️ Step 7: Update Main Application

### 7.1 Update Main Application
Update `cmd/server/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/yourusername/shopeasy-app/internal/integration"
    "github.com/yourusername/shopeasy-app/internal/logger"  // ⭐ NEW
)

func main() {
    // ⭐ NEW: Initialize basic logger for startup
    startupLogger := logger.NewLogger()
    startupLogger.SetComponent("startup")
    startupLogger.Info("🛒 Starting ShopEasy Application...")
    
    // Create context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Determine config file based on environment
    configFile := getConfigFile()
    startupLogger.Info("📁 Using config file", map[string]interface{}{
        "config_file": configFile,
    })
    
    // Initialize unified manager
    unifiedManager := integration.NewUnifiedManager()
    if err := unifiedManager.Initialize(ctx, configFile); err != nil {
        startupLogger.Fatal("❌ Failed to initialize application", map[string]interface{}{
            "error": err.Error(),
        })
    }
    
    // Get the configured logger from unified manager
    appLogger := unifiedManager.GetLogger()
    appLogger.SetComponent("main")
    
    // Setup routes
    router := setupRoutes(unifiedManager)
    
    // Get server configuration
    config := unifiedManager.GetConfig()
    serverAddr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
    
    // Create HTTP server
    server := &http.Server{
        Addr:         serverAddr,
        Handler:      router,
        ReadTimeout:  time.Duration(config.Server.ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(config.Server.WriteTimeout) * time.Second,
    }
    
    // Start server in goroutine
    go func() {
        appLogger.Info("🚀 Server starting", map[string]interface{}{
            "address":     serverAddr,
            "environment": config.App.Environment,
            "debug_mode":  config.App.Debug,
            "log_format":  config.Logging.Format,
            "log_level":   config.Logging.Level,
        })
        
        appLogger.Info("🔗 Available endpoints", map[string]interface{}{
            "endpoints": []string{
                "POST /api/v1/users - Create user",
                "GET  /api/v1/users/{id} - Get user",
                "POST /api/v1/auth/login - User login",
                "POST /api/v1/payments/process - Process payment",
                "GET  /health - Health check",
            },
        })
        
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            appLogger.Fatal("❌ Server failed to start", map[string]interface{}{
                "error": err.Error(),
            })
        }
    }()
    
    // Wait for interrupt signal for graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    appLogger.Info("🛑 Shutting down server...")
    
    // Create shutdown context with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()
    
    // Shutdown server gracefully
    if err := server.Shutdown(shutdownCtx); err != nil {
        appLogger.Error("⚠️ Server forced to shutdown", map[string]interface{}{
            "error": err.Error(),
        })
    } else {
        appLogger.Info("✅ Server shut down gracefully")
    }
}

func getConfigFile() string {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    configFile := fmt.Sprintf("./configs/%s.yaml", env)
    
    // Check if file exists, fallback to development
    if _, err := os.Stat(configFile); os.IsNotExist(err) {
        log.Printf("⚠️ Config file %s not found, using development config\n", configFile)
        configFile = "./configs/development.yaml"
    }
    
    return configFile
}
```

### 7.2 Update Routes Setup
Update `cmd/server/routes.go`:

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    
    "github.com/yourusername/shopeasy-app/internal/handlers"
    "github.com/yourusername/shopeasy-app/internal/integration"
    "github.com/yourusername/shopeasy-app/internal/logger"  // ⭐ NEW
)

func setupRoutes(um *integration.UnifiedManager) *chi.Mux {
    r := chi.NewRouter()
    
    // ⭐ NEW: Create middleware logger
    middlewareLogger := logger.NewMiddlewareLogger()
    middlewareLogger.SetComponent("chi-router")
    
    // Basic middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    
    // ⭐ NEW: Custom logging middleware using our logger
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Create a custom response writer to capture status code
            ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
            
            // Call next handler
            next.ServeHTTP(ww, r)
            
            // Log the request
            middlewareLogger.LogAPIRequest(
                r.Method,
                r.URL.Path,
                ww.Status(),
                time.Since(start),
                map[string]interface{}{
                    "remote_ip":  r.RemoteAddr,
                    "user_agent": r.UserAgent(),
                    "request_id": middleware.GetReqID(r.Context()),
                },
            )
        })
    })
    
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))
    
    // Create handlers
    userHandler := handlers.NewUserHandler(um)
    paymentHandler := handlers.NewPaymentHandler(um)
    
    // API routes
    r.Route("/api/v1", func(r chi.Router) {
        // User routes (account domain)
        r.Route("/users", func(r chi.Router) {
            r.Post("/", userHandler.CreateUser)           // POST /api/v1/users
            r.Get("/{id}", userHandler.GetUser)           // GET /api/v1/users/123
        })
        
        // Auth routes (auth domain)
        r.Route("/auth", func(r chi.Router) {
            r.Post("/login", userHandler.LoginUser)       // POST /api/v1/auth/login
        })
        
        // Payment routes (payment domain)
        r.Route("/payments", func(r chi.Router) {
            r.Post("/process", paymentHandler.ProcessPayment) // POST /api/v1/payments/process
        })
    })
    
    // Health check with logging
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        healthLogger := um.GetLogger()
        healthLogger.SetComponent("health-check")
        
        healthLogger.Debug("Health check requested", map[string]interface{}{
            "remote_ip": r.RemoteAddr,
        })
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    return r
}
```

## 🧪 Step 8: Add Remaining Logger Files

You'll need to copy the remaining logger files from the provided logger system. Create these files in `internal/logger/`:

1. **logger_core.go** - Main logger implementation
2. **logger_factory.go** - Factory constructors  
3. **logger_specialized.go** - Domain-specific logging methods
4. **logger_formatters.go** - Output formatting
5. **logger_global.go** - Global convenience functions
6. **logger_utils.go** - Helper utilities

Copy the content from the provided files exactly, just make sure to update the package declaration to:
```go
package logger
```

## 🚀 Step 9: Test the Integration

### 9.1 Run the Application
```bash
# Development mode
go run cmd/server/main.go

# You should see structured, beautiful logs like:
# [10:30:15.123] ℹ️  INFO [HANDLER] <startup> 🛒 Starting ShopEasy Application...
# [10:30:15.124] ℹ️  INFO [HANDLER] <startup> 📁 Using config file | config_file=./configs/development.yaml
# [10:30:15.125] ℹ️  INFO 🚀 Initializing Unified Manager...
# [10:30:15.126] ℹ️  INFO 📊 Loading configuration...
# [10:30:15.127] ✅ Configuration loaded from: ./configs/development.yaml
# [10:30:15.128] ℹ️  INFO 🔧 Configuring logger...
# [10:30:15.129] ℹ️  INFO [SERVICE] 🚀 Server starting | address=localhost:8080 environment=development debug_mode=true
```

### 9.2 Test API Endpoints

#### Test User Creation (Success)
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123",
    "name": "John Doe"
  }'

# Console logs will show:
# [10:30:45.001] ℹ️  INFO [HANDLER] <user-handler> {create_user} Handling create user request | method=POST path=/api/v1/users user_agent=curl/7.68.0 remote_ip=127.0.0.1
# [10:30:45.002] ℹ️  INFO [SERVICE] <user-service> {create_user} Creating new user | email=j***n@example.com
# [10:30:45.003] ℹ️  INFO [SERVICE] <user-service> User j***n@example.com performed user_created on user_account | user_id=1 name=John Doe
# [10:30:45.004] ℹ️  INFO [HANDLER] <user-handler> POST /api/v1/users → 200 | user_id=1 user_email=john@example.com operation=create_user took=3ms status=200
```

#### Test Validation Error
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "",
    "password": "123",
    "name": "John Doe"
  }'

# Console logs will show:
# [10:31:15.001] ℹ️  INFO [HANDLER] <user-handler> {create_user} Handling create user request
# [10:31:15.002] ⚠️  WARN [VALIDATION] <user-service> {validate_email} Validation failed for email | field=email message=Email is required value= value_type=string cause=validation_failed
# [10:31:15.003] ❌ ERROR [HANDLER] <user-handler> {handle_account_error} API Error: Please check your input and try again | error_code=VALIDATION_FAILED domain=account http_method=POST path=/api/v1/users
```

#### Test Payment Processing
```bash
curl -X POST http://localhost:8080/api/v1/payments/process \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 99.99,
    "card_number": "4111111111111111",
    "cvv": "123",
    "expiry_date": "12/25",
    "order_id": "order123"
  }'

# Success logs:
# [10:32:01.001] ℹ️  INFO [HANDLER] <payment-handler> Processing payment request | amount=99.99 order_id=order123 card_last_four=****-****-****-1111 operation=process_payment remote_ip=127.0.0.1
# [10:32:01.002] ℹ️  INFO [SERVICE] <payment-service> {process_payment} Processing payment | amount=99.99 order_id=order123 card_last_four=****-****-****-1111
# [10:32:01.055] ℹ️  INFO [SERVICE] <payment-service> Payment processed successfully | transaction_id=txn_1692123456789 amount=99.99 order_id=order123 processing_time_ms=53 status=completed card_last_four=****-****-****-1111

# Failure logs (if payment fails):
# [10:32:15.001] ❌ ERROR [SERVICE] <payment-service> {process_payment} Payment processing failed | amount=99.99 order_id=order123 failure_reason=insufficient_funds processing_time_ms=45 card_last_four=****-****-****-1111 cause=insufficient_funds
```

## 🎯 Step 10: Environment-Specific Behavior

### 10.1 Development Environment
```bash
# Pretty formatted logs with emojis and colors
[10:30:15.123] 🔍 DEBUG [SERVICE] <user-service> {create_user} Creating new user | email=j***n@example.com
[10:30:15.124] ℹ️  INFO [HANDLER] <user-handler> {create_user} User created successfully
[10:30:15.125] ⚠️  WARN [AUTH] <auth-service> Authentication failed | email=j***n@example.com success=false reason=invalid_password
[10:30:15.126] ❌ ERROR [PAYMENT] <payment-service> {process_payment} Payment processing failed | cause=insufficient_funds (payment_service.go:45)
```

### 10.2 Production Environment
```bash
# Set production environment
export APP_ENV=production
go run cmd/server/main.go

# JSON formatted logs suitable for log aggregation
{"timestamp":"2024-08-22 10:30:15.123","level":"INFO","message":"Server starting","context":{"address":"0.0.0.0:8080","environment":"production"},"component":"main","environment":"production","service":"shopeasy","version":"1.0.0"}
{"timestamp":"2024-08-22 10:30:16.001","level":"ERROR","message":"Payment processing failed","context":{"amount":99.99,"cause":"payment_declined","transaction_id":"txn_123"},"component":"payment-service","layer":"service","operation":"process_payment","environment":"production"}
```

## ✅ Summary of Integration

You've successfully integrated the comprehensive logger system into ShopEasy with:

1. **🔧 Configuration Integration**: Logger settings in YAML config files
2. **🎯 Structured Logging**: Replace all `fmt.Printf` with structured, contextual logging
3. **🌍 Environment Awareness**: Pretty logs in dev, JSON in production
4. **🔐 Security**: Automatic email masking and sensitive data protection
5. **📊 Performance Tracking**: Automatic request duration and performance categorization
6. **🚨 Error Correlation**: Enhanced error logging with causes and context
7. **🔄 Hot Reloading**: Logger configuration updates without restart
8. **📈 Domain-Specific**: Different logging behaviors for different business domains

The system now provides:
- **Beautiful development logs** with emojis and colors
- **Production-ready JSON logs** for log aggregation
- **Automatic sensitive data masking**
- **Performance monitoring and alerting**
- **Structured error tracking with causes**
- **Request/response correlation**
- **Security event logging**
- **Database operation tracking**

Every log entry now includes rich contextual information making debugging and monitoring much more effective!
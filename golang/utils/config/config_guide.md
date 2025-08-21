# Go Configuration Package Usage Guide

## Overview
This package provides a comprehensive configuration management system for a Go application with hot-reloading, environment-specific overrides, and validation capabilities.

## Quick Start

### 1. Initialize Configuration
```go
// Initialize with config file
err := InitializeConfig("./config/config.yaml")
if err != nil {
    panic(err)
}

// Or use the must version that panics on error
MustInitializeConfig("./config/config.yaml")

// Get the global config
config := GetConfig()
```

### 2. Basic Usage Examples
```go
// Check environment
if config.IsDevelopment() {
    // Development-specific logic
}

// Get database URL
dbURL := config.GetDatabaseURL()

// Check if role is valid
if config.IsValidRole("admin") {
    // Allow access
}

// Get server address
serverAddr := config.GetServerAddress()
```

## All Function Signatures

### ConfigManager Functions

#### Core Manager Functions
```go
func NewConfigManager() *ConfigManager {}

func (cm *ConfigManager) Load(ctx context.Context, configPath string) (*Config, error) {}

func (cm *ConfigManager) Reload(ctx context.Context) error {}

func (cm *ConfigManager) Watch(ctx context.Context) error {}

func (cm *ConfigManager) GetConfig() *Config {}

func (cm *ConfigManager) RegisterCallback(callback ConfigChangeCallback) {}

func (cm *ConfigManager) Validate() error {}

func (cm *ConfigManager) Stop() error {}
```

#### Internal Manager Functions
```go
func (cm *ConfigManager) setDefaults() {}

func (cm *ConfigManager) handleEnvironmentOverrides() error {}

func (cm *ConfigManager) handleCommonEnvironmentVariables() {}

func (cm *ConfigManager) handleDockerOverrides() error {}

func (cm *ConfigManager) handleProductionOverrides() error {}

func (cm *ConfigManager) handleTestingOverrides() error {}

func (cm *ConfigManager) validateConfig(config *Config) error {}

func (cm *ConfigManager) validateCrossFields(config *Config) error {}

func (cm *ConfigManager) setEnvironmentDefaults(env string) {}
```

### Global Configuration Functions

#### Initialization Functions
```go
func InitializeConfig(configPath string) error {}

func MustInitializeConfig(configPath string) {}

func GetConfig() *Config {}

func GetConfigManager() *ConfigManager {}

func ReloadGlobalConfig() error {}

func IsConfigInitialized() bool {}
```

#### Legacy Support Functions
```go
func LoadServerLegacy() (*LegacyServerConfig, error) {}
```

### Config Utility Functions

#### Environment Check Functions
```go
func (c *Config) IsDevelopment() bool {}

func (c *Config) IsProduction() bool {}

func (c *Config) IsStaging() bool {}

func (c *Config) IsTesting() bool {}

func (c *Config) IsDocker() bool {}

func (c *Config) IsDebugEnabled() bool {}
```

#### Server Configuration Functions
```go
func (c *Config) GetServerAddress() string {}

func (c *Config) GetGRPCAddress() string {}

func (c *Config) IsHTTPSRequired() bool {}
```

#### Database Configuration Functions
```go
func (c *Config) GetDatabaseURL() string {}

func (c *Config) GetDatabaseDSN() string {}
```

#### Security Configuration Functions
```go
func (c *Config) GetAllowedOrigins() []string {}

func (c *Config) IsOriginAllowed(origin string) bool {}

func (c *Config) GetAllowedOriginsString() string {}

func (c *Config) GetAccountLockoutMinutes() int {}

func (c *Config) GetAccountLockoutDuration() time.Duration {}

func (c *Config) GetSessionTimeout() time.Duration {}

func (c *Config) IsCSRFEnabled() bool {}

func (c *Config) IsCORSEnabled() bool {}
```

#### Role and Status Validation Functions
```go
func (c *Config) IsValidRole(role string) bool {}

func (c *Config) GetValidRolesString() string {}

func (c *Config) IsValidAccountStatus(status string) bool {}

func (c *Config) GetValidAccountStatusesString() string {}

func (c *Config) GetValidAccountStatusesMap() map[string]bool {}
```

#### Email Configuration Functions
```go
func (c *Config) GetEmailFromAddress() string {}

func (c *Config) GetSMTPAddress() string {}

func (c *Config) IsEmailVerificationRequired() bool {}

func (c *Config) GetAllowedEmailDomains() []string {}

func (c *Config) IsEmailDomainAllowed(domain string) bool {}

func (c *Config) IsEmailAllowed(email string) bool {}

func (c *Config) GetAllowedEmailDomainsString() string {}
```

#### Password Policy Functions
```go
func (c *Config) GetPasswordPolicy() map[string]interface{} {}

func (c *Config) GetMinPasswordLength() int {}

func (c *Config) GetMaxPasswordLength() int {}

func (c *Config) GetPasswordSpecialChars() string {}

func (c *Config) IsPasswordComplexityRequired() bool {}
```

#### JWT Configuration Functions
```go
func (c *Config) GetJWTSecretBytes() []byte {}
```

#### Rate Limiting Functions
```go
func (c *Config) IsRateLimitEnabled() bool {}
```

#### External API Functions
```go
func (c *Config) GetAnthropicAPIURL() string {}

func (c *Config) GetQuanAnAddress() string {}
```

#### Domain Configuration Functions
```go
func (c *Config) IsDomainEnabled(domain string) bool {}

func (c *Config) GetDefaultDomain() string {}

func (c *Config) GetEnabledDomains() []string {}

func (c *Config) GetMaxLoginAttempts() int {}
```

#### Error Handling Functions
```go
func (c *Config) GetDomainErrorLogLevel() string {}

func (c *Config) ShouldIncludeStackTrace() bool {}

func (c *Config) ShouldSanitizeSensitiveData() bool {}
```

### Domain Error Handling Functions
```go
func NewDomainAwareErrorHandler(config *Config) *DomainErrorHandler {}

func (deh *DomainErrorHandler) HandleError(domain string, err error) error {}

func (deh *DomainErrorHandler) handleUserError(err error) error {}

func (deh *DomainErrorHandler) handleSystemError(err error) error {}

func InitializeDomainErrorHandling() {}
```

### Utility Functions
```go
func getEnvOrDefault(key, defaultValue string) string {}
```

### Interface Types
```go
type ConfigChangeCallback func(oldConfig, newConfig *Config) error

type ConfigLoader interface {
    Load(ctx context.Context, configPath string) (*Config, error)
    Reload(ctx context.Context) error
    Watch(ctx context.Context) error
    GetConfig() *Config
    RegisterCallback(callback ConfigChangeCallback)
    Validate() error
    Stop() error
}
```

## Configuration Structure

### Main Configuration Fields
- `Environment` - Current environment (development, production, staging, testing, docker)
- `AppName` - Application name
- `Version` - Application version
- `Debug` - Debug mode flag
- `Server` - Server configuration (ports, timeouts, TLS)
- `Database` - Database connection settings
- `Security` - Security policies and settings
- `Password` - Password validation rules
- `Pagination` - Default pagination settings
- `JWT` - JSON Web Token configuration
- `Email` - Email service settings
- `RateLimit` - Rate limiting configuration
- `ExternalAPIs` - External service configurations
- `Logging` - Logging configuration
- `ValidRoles` - List of valid user roles
- `ValidAccountStatuses` - List of valid account statuses
- `Domains` - Domain-specific configurations
- `ErrorHandling` - Error handling policies

## Environment Variables Support

The package automatically reads from environment variables with the `ENGLISH_AI_` prefix:
- `ENGLISH_AI_DATABASE_HOST` → `database.host`
- `ENGLISH_AI_JWT_SECRET_KEY` → `jwt.secret_key`
- `ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY` → `external_apis.anthropic.api_key`

Common environment variables:
- `APP_ENV` - Override environment
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET`
- `ANTHROPIC_API_KEY`
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`

## Configuration File Structure

The package supports YAML configuration files with environment-specific overrides:

```yaml
# Main configuration
environment: development
app_name: "English AI"
# ... other settings

---
# Production overrides
environment: "production"
# Production-specific overrides
```

## Usage Patterns

### 1. Application Startup
```go
func main() {
    // Initialize configuration
    err := InitializeConfig("./config/config.yaml")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    config := GetConfig()
    
    // Start server
    server := NewServer(config.GetServerAddress())
    server.Start()
}
```

### 2. Configuration Watching
```go
func setupConfigWatching(ctx context.Context) {
    cm := GetConfigManager()
    
    // Register callback for config changes
    cm.RegisterCallback(func(old, new *Config) error {
        log.Println("Configuration changed, reloading...")
        return nil
    })
    
    // Start watching
    if err := cm.Watch(ctx); err != nil {
        log.Fatal("Failed to start config watcher:", err)
    }
}
```

### 3. Domain-Specific Error Handling
```go
func handleUserRequest(domain string) {
    config := GetConfig()
    errorHandler := NewDomainAwareErrorHandler(config)
    
    // Use domain-aware error handling
    if err := someOperation(); err != nil {
        return errorHandler.HandleError("user", err)
    }
}
```

This configuration system provides a robust, flexible way to manage application settings across different environments with comprehensive validation and hot-reloading capabilities.


# Development configuration
environment: development
app_name: "English AI"
version: "1.0.0"
debug: true

server:
  address: "localhost"
  port: 8888
  grpc_address: "localhost" 
  grpc_port: 50051
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  tls_enabled: false

database:
  host: "localhost"
  port: 5432
  name: "english_ai_dev"
  user: "postgres"
  password: "" # Set via environment variable
  ssl_mode: "disable"
  max_connections: 25
  max_idle_conns: 10
  conn_max_lifetime: "1h"
  conn_max_idle_time: "10m"

security:
  max_login_attempts: 5
  account_lockout_minutes: 15
  session_timeout: "24h"
  csrf_enabled: true
  cors_enabled: true
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
    - "http://localhost:8888"
  require_https: false

password:
  min_length: 8
  max_length: 128
  require_uppercase: true
  require_lowercase: true
  require_numbers: true
  require_special: true
  special_chars: "!@#$%^&*()_+-=[]{}|;:,.<>?/"

pagination:
  default_size: 10
  max_size: 100
  limit: 1000

jwt:
  secret_key: "kIOopC3C7wA8DQH6FOF2Jfn+UZP8Q02nGxr/EgFMOmo="
  expiration_hours: 24
  refresh_token_expiration_days: 30
  issuer: "english-ai-dev"
  algorithm: "HS256"
  refresh_threshold: "2h"

email:
  verification_enabled: false
  verification_expiry_hours: 24
  require_verification: false
  smtp_host: "localhost"
  smtp_port: 1025
  from_address: "noreply@english-ai.dev"
  from_name: "English AI Development"
  templates:
    verification_template: "./templates/email/verification.html"
    welcome_template: "./templates/email/welcome.html"
    reset_password_template: "./templates/email/reset_password.html"

rate_limit:
  enabled: true
  per_minute: 60
  per_hour: 3600
  burst_size: 10
  window_size: "1m"

external_apis:
  anthropic:
    api_key: "dummy_key_for_dev"
    api_url: "https://api.anthropic.com"
    timeout: "30s"
    max_retries: 3
  quan_an:
    address: "localhost:8081"
    timeout: "10s"
    max_retries: 3

logging:
  level: "debug"
  format: "text"
  output: "stdout"
  max_size: 100
  max_backups: 3
  max_age: 28
  compress: false

# Global valid roles and statuses (applies to ALL environments)
valid_roles:
  - "admin"
  - "user"
  - "manager"
  - "guest"    # Added guest role from your original code

valid_account_statuses:
  - "active"
  - "inactive"
  - "suspended"
  - "pending"

# Domain-specific configuration
domains:
  enabled:
    - "account"
    - "branch"
    - "order"
    - "delivery"
    - "dish"
    - "table"
    - "guest"
    - "set"
    - "websocket"
    - "system"
    - "image"
  default: "system"
  
  account:
    max_login_attempts: 5
    password_complexity: true
    jwt_required: true
    
  branch:
    validation_required: true
    cache_enabled: true
    
  order:
    status_transitions: true
    websocket_notifications: true
    payment_validation: true
    
  delivery:
    real_time_tracking: true
    status_updates: true
    websocket_required: true
    
  websocket:
    jwt_validation: true
    message_size_limit: "1MB"
    connection_timeout: "30s"

# Error handling configuration
error_handling:
  include_stack_trace: true    # true for development
  sanitize_sensitive_data: false # false for development
  log_level: "debug"

---
# Production overrides (this section overrides above values in production)
environment: "production"

# Production-specific security
security:
  require_https: true
  max_login_attempts: 3  # Stricter in production

# Production error handling
error_handling:
  include_stack_trace: false
  sanitize_sensitive_data: true
  log_level: "info"

# Production logging
logging:
  level: "info"
  format: "json"

# Production can have different valid roles if needed
# valid_roles:
#   - "admin"
#   - "user"
#   - "manager"
#   # Note: No "guest" role in production

domains:
  account:
    max_login_attempts: 3  # Different from development

    package utils_config

// setDefaults sets default configuration values
func (cm *ConfigManager) setDefaults() {
	// Environment settings
	cm.viper.SetDefault("environment", EnvDevelopment)
	cm.viper.SetDefault("app_name", "English AI")
	cm.viper.SetDefault("version", "1.0.0")
	cm.viper.SetDefault("debug", false)

	// Server settings
	cm.viper.SetDefault("server.address", "localhost")
	cm.viper.SetDefault("server.port", 8080)
	cm.viper.SetDefault("server.grpc_address", "localhost")
	cm.viper.SetDefault("server.grpc_port", 50051)
	cm.viper.SetDefault("server.read_timeout", "30s")
	cm.viper.SetDefault("server.write_timeout", "30s")
	cm.viper.SetDefault("server.idle_timeout", "120s")
	cm.viper.SetDefault("server.tls_enabled", false)

	// Database settings
	cm.viper.SetDefault("database.host", "localhost")
	cm.viper.SetDefault("database.port", 5432)
	cm.viper.SetDefault("database.name", "english_ai")
	cm.viper.SetDefault("database.user", "postgres")
	cm.viper.SetDefault("database.password", "")
	cm.viper.SetDefault("database.ssl_mode", "disable")
	cm.viper.SetDefault("database.max_connections", 25)
	cm.viper.SetDefault("database.max_idle_conns", 10)
	cm.viper.SetDefault("database.conn_max_lifetime", "1h")
	cm.viper.SetDefault("database.conn_max_idle_time", "10m")
	// Set a default database URL
	cm.viper.SetDefault("database.url", "postgres://postgres:@localhost:5432/english_ai?sslmode=disable")

	// Security settings
	cm.viper.SetDefault("security.max_login_attempts", 5)
	cm.viper.SetDefault("security.account_lockout_minutes", 15)
	cm.viper.SetDefault("security.session_timeout", "24h")
	cm.viper.SetDefault("security.csrf_enabled", true)
	cm.viper.SetDefault("security.cors_enabled", true)
	cm.viper.SetDefault("security.allowed_origins", []string{"http://localhost:3000"})
	cm.viper.SetDefault("security.require_https", false)

	// Password settings
	cm.viper.SetDefault("password.min_length", 8)
	cm.viper.SetDefault("password.max_length", 128)
	cm.viper.SetDefault("password.require_uppercase", true)
	cm.viper.SetDefault("password.require_lowercase", true)
	cm.viper.SetDefault("password.require_numbers", true)
	cm.viper.SetDefault("password.require_special", true)
	cm.viper.SetDefault("password.special_chars", "!@#$%^&*()_+-=[]{}|;:,.<>?/")

	// Pagination settings
	cm.viper.SetDefault("pagination.default_size", 10)
	cm.viper.SetDefault("pagination.max_size", 100)
	cm.viper.SetDefault("pagination.limit", 1000)

	// JWT settings
	cm.viper.SetDefault("jwt.secret_key", "kIOopC3C7wA8DQH6FOF2Jfn+UZP8Q02nGxr/EgFMOmo=")
	cm.viper.SetDefault("jwt.expiration_hours", 24)
	cm.viper.SetDefault("jwt.refresh_token_expiration_days", 30)
	cm.viper.SetDefault("jwt.issuer", "english-ai")
	cm.viper.SetDefault("jwt.algorithm", "HS256")
	cm.viper.SetDefault("jwt.refresh_threshold", "2h")

	// Email settings
	cm.viper.SetDefault("email.verification_enabled", false) // Changed to false for development
	cm.viper.SetDefault("email.verification_expiry_hours", 24)
	cm.viper.SetDefault("email.require_verification", false)
	cm.viper.SetDefault("email.smtp_host", "localhost")
	cm.viper.SetDefault("email.smtp_port", 587)
	cm.viper.SetDefault("email.smtp_user", "")
	cm.viper.SetDefault("email.smtp_password", "")
	cm.viper.SetDefault("email.from_address", "noreply@english-ai.dev") // Default valid email
	cm.viper.SetDefault("email.from_name", "English AI")

	// Rate limiting settings
	cm.viper.SetDefault("rate_limit.enabled", true)
	cm.viper.SetDefault("rate_limit.per_minute", 60)
	cm.viper.SetDefault("rate_limit.per_hour", 3600)
	cm.viper.SetDefault("rate_limit.burst_size", 10)
	cm.viper.SetDefault("rate_limit.window_size", "1m")

	// External API settings
	cm.viper.SetDefault("external_apis.anthropic.api_key", "dummy_key_for_dev") // Default dummy key
	cm.viper.SetDefault("external_apis.anthropic.api_url", "https://api.anthropic.com")
	cm.viper.SetDefault("external_apis.anthropic.timeout", "30s")
	cm.viper.SetDefault("external_apis.anthropic.max_retries", 3)
	cm.viper.SetDefault("external_apis.quan_an.address", "localhost:8081") // Default address
	cm.viper.SetDefault("external_apis.quan_an.timeout", "10s")
	cm.viper.SetDefault("external_apis.quan_an.max_retries", 3)

	// Logging settings
	cm.viper.SetDefault("logging.level", "info")
	cm.viper.SetDefault("logging.format", "json")
	cm.viper.SetDefault("logging.output", "stdout")
	cm.viper.SetDefault("logging.max_size", 100)
	cm.viper.SetDefault("logging.max_backups", 3)
	cm.viper.SetDefault("logging.max_age", 28)
	cm.viper.SetDefault("logging.compress", true)

	// Valid roles
	cm.viper.SetDefault("valid_roles", []string{"admin", "user", "manager"})

	// Domain configuration for multi-domain error handling
	cm.viper.SetDefault("domains.enabled", []string{
		"account",
	
		"auth",
		"admin",
		
		"system",
	})

		// Domain-specific settings
	cm.viper.SetDefault("domains.default", "system")
	cm.viper.SetDefault("domains.error_tracking.enabled", true)
	cm.viper.SetDefault("domains.error_tracking.log_level", "info")

	// Valid roles (existing)
cm.viper.SetDefault("valid_roles", []string{"admin", "user", "manager"})

// Add valid account statuses
cm.viper.SetDefault("valid_account_statuses", []string{"active", "inactive", "suspended", "pending"})
}
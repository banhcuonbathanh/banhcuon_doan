# Go Configuration Management System Documentation

## Overview

This is a comprehensive configuration management system for a Go-based English AI application. It provides hierarchical configuration loading, environment-specific overrides, hot-reloading capabilities, and domain-aware error handling.

## Package Structure

```
golang/utils/config/
├── config.yaml                    # Main configuration file
├── utils_config_type.go           # Type definitions and structs
├── utils_config_manager.go        # Core configuration manager
├── utils_config_default.go        # Default value setters
├── utils_config_env.go            # Environment-specific overrides
├── utils_config_global.go         # Global configuration management
├── utils_config_interface.go      # Interface definitions
└── utils_config_utility.go        # Utility methods and helpers
```

## Constants

### Environment Constants
```go
const (
    EnvDevelopment = "development"
    EnvStaging     = "staging"
    EnvProduction  = "production"
    EnvTesting     = "testing"
    EnvDocker      = "docker"
)
```

### Default Values
```go
// Server defaults
DefaultServerAddress = "localhost"
DefaultServerPort = 8080
DefaultGRPCPort = 50051

// Database defaults
DefaultDatabaseName = "english_ai"
DefaultDatabaseUser = "postgres"
DefaultDatabasePort = 5432

// Security defaults
DefaultMaxLoginAttempts = 5
DefaultSessionTimeout = "24h"
DefaultPasswordMinLength = 8

// JWT defaults
DefaultJWTAlgorithm = "HS256"
DefaultJWTExpirationHours = 24

// Rate limiting defaults
DefaultRateLimitPerMinute = 60
DefaultRateLimitPerHour = 3600
```

## Core Types and Structs

### Main Configuration Structure
```go
type Config struct {
    Environment    string                `mapstructure:"environment" validate:"required,oneof=development staging production testing docker"`
    AppName        string                `mapstructure:"app_name" validate:"required,min=1,max=50"`
    Version        string                `mapstructure:"version" validate:"required"`
    Debug          bool                  `mapstructure:"debug"`
    
    Server         ServerConfig          `mapstructure:"server" validate:"required"`
    Database       DatabaseConfig        `mapstructure:"database" validate:"required"`
    Security       SecurityConfig        `mapstructure:"security" validate:"required"`
    Password       PasswordConfig        `mapstructure:"password" validate:"required"`
    Pagination     PaginationConfig      `mapstructure:"pagination" validate:"required"`
    JWT            JWTConfig             `mapstructure:"jwt" validate:"required"`
    Email          EmailConfig           `mapstructure:"email" validate:"required"`
    RateLimit      RateLimitConfig       `mapstructure:"rate_limit" validate:"required"`
    ExternalAPIs   ExternalAPIConfig     `mapstructure:"external_apis" validate:"required"`
    Logging        LoggingConfig         `mapstructure:"logging" validate:"required"`
    Domains        DomainConfig          `mapstructure:"domains"`
    ErrorHandling  ErrorHandlingConfig   `mapstructure:"error_handling"`
    
    ValidRoles           []string `mapstructure:"valid_roles" validate:"required,min=1,dive,required"`
    ValidAccountStatuses []string `mapstructure:"valid_account_statuses" validate:"required,min=1,dive,required"`
}
```

### Server Configuration
```go
type ServerConfig struct {
    Address      string        `mapstructure:"address" validate:"required"`
    Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
    GRPCAddress  string        `mapstructure:"grpc_address" validate:"required"`
    GRPCPort     int           `mapstructure:"grpc_port" validate:"required,min=1,max=65535"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required"`
    WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required"`
    IdleTimeout  time.Duration `mapstructure:"idle_timeout" validate:"required"`
    TLSEnabled   bool          `mapstructure:"tls_enabled"`
    CertFile     string        `mapstructure:"cert_file"`
    KeyFile      string        `mapstructure:"key_file"`
}
```

### Database Configuration
```go
type DatabaseConfig struct {
    URL             string        `mapstructure:"url" validate:"required"`
    Host            string        `mapstructure:"host" validate:"required"`
    Port            int           `mapstructure:"port" validate:"required,min=1,max=65535"`
    Name            string        `mapstructure:"name" validate:"required"`
    User            string        `mapstructure:"user" validate:"required"`
    Password        string        `mapstructure:"password"`
    SSLMode         string        `mapstructure:"ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
    MaxConnections  int           `mapstructure:"max_connections" validate:"required,min=1,max=100"`
    MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"required,min=1"`
    ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" validate:"required"`
    ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" validate:"required"`
}
```

### Security Configuration
```go
type SecurityConfig struct {
    MaxLoginAttempts      int           `mapstructure:"max_login_attempts" validate:"required,min=1,max=20"`
    AccountLockoutMinutes int           `mapstructure:"account_lockout_minutes" validate:"required,min=1"`
    SessionTimeout        time.Duration `mapstructure:"session_timeout" validate:"required"`
    CSRFEnabled           bool          `mapstructure:"csrf_enabled"`
    CORSEnabled           bool          `mapstructure:"cors_enabled"`
    AllowedOrigins        []string      `mapstructure:"allowed_origins"`
    AllowedEmailDomains   []string      `mapstructure:"allowed_email_domains"`
    RequireHTTPS          bool          `mapstructure:"require_https"`
}
```

### JWT Configuration
```go
type JWTConfig struct {
    SecretKey                  string        `mapstructure:"secret_key" validate:"required,min=32"`
    ExpirationHours            int           `mapstructure:"expiration_hours" validate:"required,min=1,max=720"`
    RefreshTokenExpirationDays int           `mapstructure:"refresh_token_expiration_days" validate:"required,min=1,max=365"`
    Issuer                     string        `mapstructure:"issuer" validate:"required"`
    Algorithm                  string        `mapstructure:"algorithm" validate:"required,oneof=HS256 HS384 HS512 RS256 RS384 RS512"`
    RefreshThreshold           time.Duration `mapstructure:"refresh_threshold" validate:"required"`
}
```

### Domain Configuration
```go
type DomainConfig struct {
    Enabled       []string                    `mapstructure:"enabled"`
    Default       string                     `mapstructure:"default"`
    ErrorTracking DomainErrorTrackingConfig  `mapstructure:"error_tracking"`
    Account       DomainAccountConfig        `mapstructure:"account"`
}

type DomainAccountConfig struct {
    MaxLoginAttempts   int  `mapstructure:"max_login_attempts" validate:"min=1,max=10"`
    PasswordComplexity bool `mapstructure:"password_complexity"`
    EmailVerification  bool `mapstructure:"email_verification"`
}
```

## Core Interfaces

### Configuration Loader Interface
```go
type ConfigLoader interface {
    Load(ctx context.Context, configPath string) (*Config, error)
    Reload(ctx context.Context) error
    Watch(ctx context.Context) error
    GetConfig() *Config
    RegisterCallback(callback ConfigChangeCallback)
    Validate() error
    Stop() error
}

type ConfigChangeCallback func(oldConfig, newConfig *Config) error
```

## Configuration Manager

### ConfigManager Structure
```go
type ConfigManager struct {
    config    *Config
    mu        sync.RWMutex
    validator *validator.Validate
    viper     *viper.Viper
    callbacks []ConfigChangeCallback
}
```

## Key Function Signatures

### Initialization Functions
```go
// Global configuration initialization
func InitializeConfig(configPath string) error
func MustInitializeConfig(configPath string)
func GetConfig() *Config
func GetConfigManager() *ConfigManager
func IsConfigInitialized() bool

// Create new configuration manager
func NewConfigManager() *ConfigManager
```

### Core Configuration Management
```go
// Load and manage configuration
func (cm *ConfigManager) Load(ctx context.Context, configPath string) (*Config, error)
func (cm *ConfigManager) Reload(ctx context.Context) error
func (cm *ConfigManager) Watch(ctx context.Context) error
func (cm *ConfigManager) GetConfig() *Config
func (cm *ConfigManager) RegisterCallback(callback ConfigChangeCallback)
func (cm *ConfigManager) Validate() error
func (cm *ConfigManager) Stop() error

// Set default values
func (cm *ConfigManager) setDefaults()
```

### Environment Handling
```go
// Environment-specific overrides
func (cm *ConfigManager) handleEnvironmentOverrides() error
func (cm *ConfigManager) handleCommonEnvironmentVariables()
func (cm *ConfigManager) handleDockerOverrides() error
func (cm *ConfigManager) handleProductionOverrides() error
func (cm *ConfigManager) handleTestingOverrides() error

// Utility function for environment variables
func getEnvOrDefault(key, defaultValue string) string
```

### Validation Functions
```go
// Configuration validation
func (cm *ConfigManager) validateConfig(config *Config) error
func (cm *ConfigManager) validateCrossFields(config *Config) error
```

## Utility Methods

### Environment Checks
```go
func (c *Config) IsDevelopment() bool
func (c *Config) IsProduction() bool
func (c *Config) IsStaging() bool
func (c *Config) IsTesting() bool
func (c *Config) IsDocker() bool
func (c *Config) IsDebugEnabled() bool
```

### Server Configuration
```go
func (c *Config) GetServerAddress() string
func (c *Config) GetGRPCAddress() string
func (c *Config) IsHTTPSRequired() bool
```

### Database Utilities
```go
func (c *Config) GetDatabaseURL() string
func (c *Config) GetDatabaseDSN() string
```

### Security and Authentication
```go
func (c *Config) IsValidRole(role string) bool
func (c *Config) GetValidRolesString() string
func (c *Config) IsValidAccountStatus(status string) bool
func (c *Config) GetValidAccountStatusesString() string
func (c *Config) GetJWTSecretBytes() []byte
func (c *Config) GetMaxLoginAttempts() int
func (c *Config) GetAccountLockoutDuration() time.Duration
func (c *Config) GetSessionTimeout() time.Duration
```

### CORS and Origins
```go
func (c *Config) GetAllowedOrigins() []string
func (c *Config) IsOriginAllowed(origin string) bool
func (c *Config) GetAllowedOriginsString() string
func (c *Config) IsCSRFEnabled() bool
func (c *Config) IsCORSEnabled() bool
```

### Email Configuration
```go
func (c *Config) GetAllowedEmailDomains() []string
func (c *Config) IsEmailDomainAllowed(domain string) bool
func (c *Config) IsEmailAllowed(email string) bool
func (c *Config) GetEmailFromAddress() string
func (c *Config) GetSMTPAddress() string
func (c *Config) IsEmailVerificationRequired() bool
```

### Password Policy
```go
func (c *Config) GetPasswordPolicy() map[string]interface{}
func (c *Config) GetMinPasswordLength() int
func (c *Config) GetMaxPasswordLength() int
func (c *Config) GetPasswordSpecialChars() string
func (c *Config) IsPasswordComplexityRequired() bool
```

### Domain Management
```go
func (c *Config) IsDomainEnabled(domain string) bool
func (c *Config) GetDefaultDomain() string
func (c *Config) GetEnabledDomains() []string
func (c *Config) GetDomainErrorLogLevel() string
```

### Rate Limiting
```go
func (c *Config) IsRateLimitEnabled() bool
```

### External APIs
```go
func (c *Config) GetAnthropicAPIURL() string
func (c *Config) GetQuanAnAddress() string
```

### Error Handling
```go
func (c *Config) ShouldIncludeStackTrace() bool
func (c *Config) ShouldSanitizeSensitiveData() bool
```

## Legacy Support

### Legacy Configuration Structure
```go
type LegacyServerConfig struct {
    DatabaseURL string
    Port        int
    Address     string
}

func LoadServerLegacy() (*LegacyServerConfig, error)
```

## Domain Error Handling

### Domain Error Handler
```go
type DomainErrorHandler struct {
    config            *Config
    enabledDomains    []string
    defaultDomain     string
    includeStackTrace bool
    sanitizeData      bool
}

func NewDomainAwareErrorHandler(config *Config) *DomainErrorHandler
func (deh *DomainErrorHandler) HandleError(domain string, err error) error
```

## Environment Variables

The system supports the following environment variables:

### Database
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name

### Authentication
- `JWT_SECRET` - JWT secret key
- `ENGLISH_AI_JWT_SECRET_KEY` - Production JWT secret

### External APIs
- `ANTHROPIC_API_KEY` - Anthropic API key
- `ANTHROPIC_API_URL` - Anthropic API URL
- `QUAN_AN_ADDRESS` - QuanAn service address

### Email
- `SMTP_HOST` - SMTP server host
- `SMTP_PORT` - SMTP server port
- `SMTP_USER` - SMTP username
- `SMTP_PASSWORD` - SMTP password
- `EMAIL_FROM_ADDRESS` - From email address

### Application
- `APP_ENV` - Application environment
- `PORT` - Server port
- `ADDRESS` - Server address

## Configuration File Structure

The system uses YAML configuration files with the following structure:

```yaml
environment: development
app_name: "English AI"
version: "1.0.0"
debug: true

server:
  address: "localhost"
  port: 8888
  # ... other server settings

database:
  host: "localhost"
  port: 5432
  # ... other database settings

security:
  max_login_attempts: 5
  # ... other security settings

# ... other configuration sections
```

## Key Features

1. **Hierarchical Configuration**: YAML files + environment variables + defaults
2. **Environment-Specific Overrides**: Different settings per environment
3. **Hot Reloading**: Configuration changes detected automatically
4. **Validation**: Comprehensive struct tag validation + cross-field validation
5. **Domain-Aware**: Different configurations for different application domains
6. **Thread-Safe**: Safe for concurrent access
7. **Legacy Support**: Backward compatibility with existing systems

## Usage Examples

### Basic Initialization
```go
// Initialize configuration
err := InitializeConfig("./config.yaml")
if err != nil {
    panic(err)
}

// Get global configuration
config := GetConfig()
```

### Environment-Specific Usage
```go
if config.IsProduction() {
    // Production-specific logic
} else if config.IsDevelopment() {
    // Development-specific logic
}
```

### Domain Validation
```go
if config.IsDomainEnabled("account") {
    // Handle account domain
}

if config.IsValidRole("admin") {
    // Handle admin role
}
```

This configuration system provides a robust, flexible foundation for managing application settings across different environments and domains.
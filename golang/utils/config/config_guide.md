# Go Configuration Management System - Complete Guide

## Overview

This is a comprehensive configuration management system for a Go application called "English AI". It provides flexible configuration loading from YAML files and environment variables, with validation, hot-reloading, and support for multiple deployment environments.

## Architecture

The system is built around these core components:

- **ConfigManager**: Main configuration manager with hot-reloading capabilities
- **Config Struct**: Type-safe configuration structure with validation
- **Environment Overrides**: Environment-specific configuration handling
- **Global Configuration**: Singleton pattern for application-wide access
- **Backward Compatibility**: Legacy support for existing code

## File Structure

```
golang/utils/config/
├── config.yaml                    # Default configuration file
├── utils_config_manager.go        # Core configuration manager
├── utils_config_type.go           # Configuration data structures
├── utils_config_default.go       # Default configuration values
├── utils_config_env.go            # Environment-specific overrides
├── utils_config_global.go         # Global configuration access
├── utils_config_interface.go      # Configuration interfaces
├── utils_config_utility.go       # Helper utility methods
└── simple.go                      # Backward compatibility layer
```

## Configuration Structure

### Main Configuration Sections

1. **Environment Settings**
   - `environment`: development, staging, production, testing, docker
   - `app_name`: Application name
   - `version`: Application version
   - `debug`: Debug mode flag

2. **Server Configuration**
   - HTTP server settings (address, port, timeouts)
   - gRPC server settings
   - TLS configuration

3. **Database Configuration**
   - PostgreSQL connection settings
   - Connection pool configuration
   - SSL mode settings

4. **Security Settings**
   - Login attempt limits
   - Session timeouts
   - CORS and CSRF configuration
   - HTTPS requirements

5. **Authentication (JWT)**
   - Secret key management
   - Token expiration settings
   - Refresh token configuration

6. **External APIs**
   - Anthropic API configuration
   - QuanAn service configuration

7. **Additional Features**
   - Email configuration
   - Rate limiting
   - Pagination settings
   - Logging configuration

## Quick Start Guide

### 1. Basic Usage

```go
package main

import (
    "context"
    "log"
    "your-app/utils/config"
)

func main() {
    // Initialize configuration
    err := utils_config.InitializeConfig("./config/config.yaml")
    if err != nil {
        log.Fatal("Failed to initialize config:", err)
    }

    // Get configuration
    config := utils_config.GetConfig()
    
    // Use configuration
    fmt.Printf("Server will run on %s\n", config.GetServerAddress())
    fmt.Printf("Database URL: %s\n", config.GetDatabaseURL())
}
```

### 2. Using Configuration Manager Directly

```go
package main

import (
    "context"
    "your-app/utils/config"
)

func main() {
    // Create config manager
    cm := utils_config.NewConfigManager()
    
    // Load configuration
    ctx := context.Background()
    config, err := cm.Load(ctx, "./config.yaml")
    if err != nil {
        panic(err)
    }
    
    // Use configuration
    fmt.Printf("App Name: %s\n", config.AppName)
    fmt.Printf("Environment: %s\n", config.Environment)
}
```

### 3. Hot Reloading (Advanced)

```go
func setupConfigWithHotReload() {
    cm := utils_config.NewConfigManager()
    ctx := context.Background()
    
    // Load initial configuration
    config, err := cm.Load(ctx, "./config.yaml")
    if err != nil {
        panic(err)
    }
    
    // Register change callback
    cm.RegisterCallback(func(oldConfig, newConfig *utils_config.Config) error {
        fmt.Println("Configuration changed!")
        // Handle configuration changes here
        return nil
    })
    
    // Start watching for changes
    err = cm.Watch(ctx)
    if err != nil {
        panic(err)
    }
}
```

## Environment Configuration

### Supported Environments

- **development**: Local development with debug enabled
- **staging**: Pre-production environment
- **production**: Production environment with security enforced
- **testing**: Unit/integration testing
- **docker**: Containerized deployment

### Environment Variable Override

The system automatically reads environment variables with the prefix `ENGLISH_AI_`. For example:

```bash
# Override database settings
export ENGLISH_AI_DATABASE_HOST=prod-db.example.com
export ENGLISH_AI_DATABASE_PASSWORD=secret123

# Override JWT secret
export ENGLISH_AI_JWT_SECRET_KEY=your-secure-secret

# Override API keys
export ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY=your-api-key
```

### Common Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=english_ai

# Server
SERVER_PORT=8080
GRPC_PORT=50051

# Security
JWT_SECRET=your-jwt-secret

# External Services
ANTHROPIC_API_KEY=your-anthropic-key
QUAN_AN_ADDRESS=localhost:8081

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

## Configuration File Examples

### Development Configuration (config.yaml)

```yaml
environment: development
app_name: "English AI"
version: "1.0.0"
debug: true

server:
  address: "localhost"
  port: 8888
  grpc_port: 50051
  read_timeout: "30s"
  write_timeout: "30s"

database:
  host: "localhost"
  port: 5432
  name: "english_ai_dev"
  user: "postgres"
  password: ""  # Set via environment
  ssl_mode: "disable"
  max_connections: 25

security:
  max_login_attempts: 5
  session_timeout: "24h"
  cors_enabled: true
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8888"

jwt:
  secret_key: "your-secret-key"
  expiration_hours: 24
  issuer: "english-ai-dev"

external_apis:
  anthropic:
    api_key: ""  # Set via environment
    api_url: "https://api.anthropic.com"
    timeout: "30s"
  quan_an:
    address: "localhost:8081"
    timeout: "10s"
```

### Production Configuration

```yaml
environment: production
debug: false

server:
  address: "0.0.0.0"
  port: 8080
  tls_enabled: true
  cert_file: "/etc/ssl/certs/app.crt"
  key_file: "/etc/ssl/private/app.key"

security:
  require_https: true
  csrf_enabled: true
  allowed_origins:
    - "https://yourdomain.com"

database:
  ssl_mode: "require"
  max_connections: 50

logging:
  level: "info"
  format: "json"
```

## Utility Methods

The configuration struct provides many utility methods:

```go
config := utils_config.GetConfig()

// Environment checks
if config.IsProduction() {
    // Production-specific logic
}

if config.IsDevelopment() {
    // Development-specific logic
}

// Address helpers
serverAddr := config.GetServerAddress()      // "localhost:8888"
grpcAddr := config.GetGRPCAddress()         // "localhost:50051"
dbURL := config.GetDatabaseURL()            // Full PostgreSQL URL

// Validation helpers
if config.IsValidRole("admin") {
    // Role is valid
}

validRoles := config.GetValidRolesString()   // "admin, user, manager"

// Feature flags
if config.IsEmailVerificationRequired() {
    // Email verification is enabled
}

if config.IsRateLimitEnabled() {
    // Rate limiting is active
}
```

## Validation

The system includes comprehensive validation:

### Struct Tag Validation

```go
type ServerConfig struct {
    Port int `validate:"required,min=1,max=65535"`
    Address string `validate:"required"`
}
```

### Cross-Field Validation

- Password max length must be greater than min length
- Pagination default size cannot exceed max size
- JWT secret must be at least 32 characters in production
- HTTPS must be enabled in production

### Custom Validation

```go
func validateConfig(config *Config) error {
    if config.Environment == "production" {
        if !config.Security.RequireHTTPS {
            return fmt.Errorf("HTTPS required in production")
        }
    }
    return nil
}
```

## Backward Compatibility

For existing code, the system provides a compatibility layer:

```go
// Old way (still supported)
import "your-app/utils"

func main() {
    config, err := utils.LoadServer()
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Database URL:", config.DatabaseURL)
    fmt.Println("Server Address:", config.ServerAddress)
    fmt.Println("JWT Secret:", config.JwtSecret)
}
```

## Best Practices

### 1. Environment-Specific Configs

Create separate config files for different environments:

```
config/
├── config.yaml              # Base configuration
├── config.development.yaml  # Development overrides
├── config.production.yaml   # Production overrides
└── config.testing.yaml      # Testing overrides
```

### 2. Secret Management

Never commit secrets to version control:

```yaml
# In config file
database:
  password: ""  # Empty, will be set via environment

jwt:
  secret_key: ""  # Empty, will be set via environment
```

```bash
# In environment/CI/CD
export ENGLISH_AI_DATABASE_PASSWORD=actual-password
export ENGLISH_AI_JWT_SECRET_KEY=actual-secret
```

### 3. Configuration Validation

Always validate configuration on startup:

```go
func main() {
    err := utils_config.InitializeConfig("./config.yaml")
    if err != nil {
        log.Fatal("Invalid configuration:", err)
    }
    
    config := utils_config.GetConfig()
    if err := validateBusinessLogic(config); err != nil {
        log.Fatal("Business logic validation failed:", err)
    }
}
```

### 4. Configuration Changes

Use callbacks to handle configuration changes gracefully:

```go
cm.RegisterCallback(func(oldConfig, newConfig *Config) error {
    // Restart services that depend on configuration
    if oldConfig.Database.Host != newConfig.Database.Host {
        return restartDatabaseConnection(newConfig.Database)
    }
    return nil
})
```

## Deployment Considerations

### Docker Deployment

Set environment to "docker":

```bash
export ENGLISH_AI_ENVIRONMENT=docker
```

The system will automatically:
- Set server address to "0.0.0.0" 
- Update service addresses for container networking

### Kubernetes Deployment

Use ConfigMaps and Secrets:

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: english-ai-config
data:
  config.yaml: |
    environment: production
    server:
      port: 8080
    # ... rest of config

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: english-ai-secrets
data:
  jwt-secret: <base64-encoded-secret>
  db-password: <base64-encoded-password>
```

### Health Checks

Implement configuration health checks:

```go
func healthCheck() error {
    config := utils_config.GetConfig()
    
    // Check if configuration is loaded
    if config == nil {
        return fmt.Errorf("configuration not loaded")
    }
    
    // Validate configuration
    cm := utils_config.GetConfigManager()
    if err := cm.Validate(); err != nil {
        return fmt.Errorf("configuration invalid: %w", err)
    }
    
    return nil
}
```

## Troubleshooting

### Common Issues

1. **Configuration Not Found**
   ```
   Error: config file not found
   ```
   - Ensure config.yaml exists in the correct path
   - Check file permissions
   - Verify working directory

2. **Validation Errors**
   ```
   Error: validation failed: Field 'Port' failed on 'min' tag
   ```
   - Check configuration values against validation rules
   - Ensure required fields are set
   - Verify data types match expected formats

3. **Environment Variable Override Not Working**
   ```
   Environment variable not being read
   ```
   - Use correct prefix: `ENGLISH_AI_`
   - Use underscores for nested fields: `ENGLISH_AI_DATABASE_HOST`
   - Check environment variable is exported

### Debug Mode

Enable debug logging to troubleshoot configuration issues:

```yaml
debug: true
logging:
  level: debug
```

This will output detailed information about configuration loading and validation.

## Migration Guide

### From Simple Configuration

If you're using the old simple configuration:

**Before:**
```go
config, err := utils.LoadServer()
dbURL := config.DatabaseURL
```

**After:**
```go
utils_config.InitializeConfig("./config.yaml")
config := utils_config.GetConfig()
dbURL := config.GetDatabaseURL()
```

### Adding New Configuration

To add new configuration options:

1. Update the struct in `utils_config_type.go`:
```go
type Config struct {
    // ... existing fields
    NewFeature NewFeatureConfig `mapstructure:"new_feature" json:"new_feature"`
}

type NewFeatureConfig struct {
    Enabled bool   `mapstructure:"enabled" json:"enabled"`
    APIKey  string `mapstructure:"api_key" json:"-" validate:"required"`
}
```

2. Add defaults in `utils_config_default.go`:
```go
cm.viper.SetDefault("new_feature.enabled", false)
cm.viper.SetDefault("new_feature.api_key", "")
```

3. Update config.yaml:
```yaml
new_feature:
  enabled: true
  api_key: ""  # Set via environment
```

This comprehensive guide should help developers understand and effectively use the configuration management system.
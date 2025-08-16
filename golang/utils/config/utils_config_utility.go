//golang/utils/config/utils_config_utility.go

package utils_config

import (
	"fmt"
	"strings"
)

// Utility methods for the Config struct

// IsValidRole checks if a role is valid according to configuration
func (c *Config) IsValidRole(role string) bool {
	for _, validRole := range c.ValidRoles {
		if validRole == role {
			return true
		}
	}
	return false
}

// GetValidRolesString returns valid roles as a comma-separated string
func (c *Config) GetValidRolesString() string {
	return strings.Join(c.ValidRoles, ", ")
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == EnvDevelopment
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == EnvProduction
}

// IsStaging returns true if running in staging mode
func (c *Config) IsStaging() bool {
	return c.Environment == EnvStaging
}

// IsTesting returns true if running in testing mode
func (c *Config) IsTesting() bool {
	return c.Environment == EnvTesting
}

// IsDocker returns true if running in docker mode
func (c *Config) IsDocker() bool {
	return c.Environment == EnvDocker
}

// GetServerAddress returns the full server address (host:port)
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Address, c.Server.Port)
}

// GetGRPCAddress returns the full gRPC server address (host:port)
func (c *Config) GetGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.GRPCAddress, c.Server.GRPCPort)
}

// GetDatabaseURL returns the database connection URL
func (c *Config) GetDatabaseURL() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}
	
	// Construct URL from individual fields
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetDatabaseDSN returns a database DSN string (alias for GetDatabaseURL)
func (c *Config) GetDatabaseDSN() string {
	return c.GetDatabaseURL()
}

// GetJWTSecretBytes returns JWT secret as byte slice
func (c *Config) GetJWTSecretBytes() []byte {
	return []byte(c.JWT.SecretKey)
}

// GetAllowedOriginsString returns allowed origins as a comma-separated string
func (c *Config) GetAllowedOriginsString() string {
	return strings.Join(c.Security.AllowedOrigins, ", ")
}

// IsHTTPSRequired returns true if HTTPS is required
func (c *Config) IsHTTPSRequired() bool {
	return c.Security.RequireHTTPS
}

// IsDebugEnabled returns true if debug mode is enabled
func (c *Config) IsDebugEnabled() bool {
	return c.Debug
}

// GetEmailFromAddress returns the complete email from address
func (c *Config) GetEmailFromAddress() string {
	if c.Email.FromName != "" {
		return fmt.Sprintf("%s <%s>", c.Email.FromName, c.Email.FromAddress)
	}
	return c.Email.FromAddress
}

// GetSMTPAddress returns the SMTP server address
func (c *Config) GetSMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.Email.SMTPHost, c.Email.SMTPPort)
}

// IsEmailVerificationRequired returns true if email verification is required
func (c *Config) IsEmailVerificationRequired() bool {
	return c.Email.RequireVerification && c.Email.VerificationEnabled
}

// IsRateLimitEnabled returns true if rate limiting is enabled
func (c *Config) IsRateLimitEnabled() bool {
	return c.RateLimit.Enabled
}

// GetAnthropicAPIURL returns the Anthropic API URL
func (c *Config) GetAnthropicAPIURL() string {
	return c.ExternalAPIs.Anthropic.APIURL
}

// GetQuanAnAddress returns the QuanAn service address
func (c *Config) GetQuanAnAddress() string {
	return c.ExternalAPIs.QuanAn.Address
}


// Utility methods for domain configuration
func (c *Config) IsDomainEnabled(domain string) bool {
	for _, enabledDomain := range c.Domains.Enabled {
		if enabledDomain == domain {
			return true
		}
	}
	return false
}

func (c *Config) GetDefaultDomain() string {
	if c.Domains.Default != "" {
		return c.Domains.Default
	}
	return "system"
}

func (c *Config) GetEnabledDomains() []string {
	return c.Domains.Enabled
}

func (c *Config) GetMaxLoginAttempts() int {
	return c.Domains.Account.MaxLoginAttempts
}

func (c *Config) IsPasswordComplexityRequired() bool {
	return c.Domains.Account.PasswordComplexity
}





// Integration with error handling system
func (c *Config) GetDomainErrorLogLevel() string {
	return c.Domains.ErrorTracking.LogLevel
}

func (c *Config) ShouldIncludeStackTrace() bool {
	return c.ErrorHandling.IncludeStackTrace
}

func (c *Config) ShouldSanitizeSensitiveData() bool {
	return c.ErrorHandling.SanitizeSensitiveData
}

// example 


// Example usage in error handling integration
func NewDomainAwareErrorHandler(config *Config) *DomainErrorHandler {
	return &DomainErrorHandler{
		config:           config,
		enabledDomains:   config.GetEnabledDomains(),
		defaultDomain:    config.GetDefaultDomain(),
		includeStackTrace: config.ShouldIncludeStackTrace(),
		sanitizeData:     config.ShouldSanitizeSensitiveData(),
	}
}

type DomainErrorHandler struct {
	config            *Config
	enabledDomains    []string
	defaultDomain     string
	includeStackTrace bool
	sanitizeData      bool
}

func (deh *DomainErrorHandler) HandleError(domain string, err error) error {
	// Validate domain is enabled
	if !deh.config.IsDomainEnabled(domain) {
		// Use default domain if specified domain is not enabled
		domain = deh.defaultDomain
	}
	
	// Apply domain-specific error handling logic based on configuration
	switch domain {
	case "user":
		return deh.handleUserError(err)

	default:
		return deh.handleSystemError(err)
	}
}

func (deh *DomainErrorHandler) handleUserError(err error) error {
	// Apply user-specific configuration
	if deh.config.IsPasswordComplexityRequired() {
		// Enhanced password validation
	}
	return err
}





func (deh *DomainErrorHandler) handleSystemError(err error) error {
	// Apply system-level error handling
	return err
}

// Example configuration file (config.yaml)
/*
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
    max_login_attempts: 5
    password_complexity: true
    email_verification: true
  course:
    enrollment_validation: true
    prerequisite_check: true
  payment:
    provider_timeout: "30s"
    retry_attempts: 3
    webhook_validation: true

error_handling:
  include_stack_trace: false
  sanitize_sensitive_data: true
  request_id_required: true
*/

// Environment-specific overrides
func (cm *ConfigManager) setEnvironmentDefaults(env string) {
	switch env {
	case "development":
		cm.viper.SetDefault("error_handling.include_stack_trace", true)
		cm.viper.SetDefault("domains.error_tracking.log_level", "debug")
		cm.viper.SetDefault("domains.user.max_login_attempts", 10) // More lenient in dev
		
	case "production":
		cm.viper.SetDefault("error_handling.include_stack_trace", false)
		cm.viper.SetDefault("domains.error_tracking.log_level", "warn")
		cm.viper.SetDefault("domains.user.max_login_attempts", 3) // Stricter in prod
		cm.viper.SetDefault("domains.payment.webhook_validation", true)
		
	case "testing":
		cm.viper.SetDefault("error_handling.include_stack_trace", true)
		cm.viper.SetDefault("domains.error_tracking.log_level", "debug")
		cm.viper.SetDefault("domains.user.email_verification", false) // Skip in tests
	}
}

// Initialize domain-aware error handling in your application
func InitializeDomainErrorHandling() {
	// Load configuration
	err := InitializeConfig("./config.yaml")
	if err != nil {
		panic(err)
	}
	
	config := GetConfig()
	
	// Create domain-aware error handler
	errorHandler := NewDomainAwareErrorHandler(config)
	
	// Use in your middleware or handlers
	_ = errorHandler
}


// IsValidAccountStatus checks if the given status is valid
func (c *Config) IsValidAccountStatus(status string) bool {
    for _, validStatus := range c.ValidAccountStatuses {
        if strings.EqualFold(validStatus, status) {
            return true
        }
    }
    return false
}

// GetValidAccountStatusesString returns valid statuses as a comma-separated string
func (c *Config) GetValidAccountStatusesString() string {
    return strings.Join(c.ValidAccountStatuses, ", ")
}

// GetValidAccountStatusesMap returns valid statuses as a map for quick lookup
func (c *Config) GetValidAccountStatusesMap() map[string]bool {
    statusMap := make(map[string]bool, len(c.ValidAccountStatuses))
    for _, status := range c.ValidAccountStatuses {
        statusMap[strings.ToLower(status)] = true
    }
    return statusMap
}

// IsValidAccountStatus checks if the given status is valid



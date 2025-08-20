//golang/utils/config/utils_config_utility.go

package utils_config

import (
	"fmt"
	"strings"
	"time"
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


// Add these methods to your Config struct in the utils_config package

// GetAllowedOrigins returns the list of allowed origins from security config
func (c *Config) GetAllowedOrigins() []string {
	if c == nil || len(c.Security.AllowedOrigins) == 0 {
		// Return default allowed origins if none configured
		return []string{"http://localhost:3000", "http://localhost:8080"}
	}
	return c.Security.AllowedOrigins
}

// IsOriginAllowed checks if the given origin is in the allowed origins list
func (c *Config) IsOriginAllowed(origin string) bool {
	if origin == "" {
		return true // Allow empty origin (same-origin requests)
	}
	
	allowedOrigins := c.GetAllowedOrigins()
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}


// GetAccountLockoutMinutes returns account lockout duration in minutes
func (c *Config) GetAccountLockoutMinutes() int {
	if c == nil {
		return 15 // Default value
	}
	return c.Security.AccountLockoutMinutes
}



// IsCSRFEnabled returns whether CSRF protection is enabled
func (c *Config) IsCSRFEnabled() bool {
	if c == nil {
		return true // Default to enabled for security
	}
	return c.Security.CSRFEnabled
}

// IsCORSEnabled returns whether CORS is enabled
func (c *Config) IsCORSEnabled() bool {
	if c == nil {
		return false
	}
	return c.Security.CORSEnabled
}

// Add these methods to your utils_config_utility.go file

// GetAllowedEmailDomains returns the list of allowed email domains from security config
func (c *Config) GetAllowedEmailDomains() []string {
	if c == nil || len(c.Security.AllowedEmailDomains) == 0 {
		// Return empty slice if no restrictions configured (allows all domains)
		return []string{}
	}
	return c.Security.AllowedEmailDomains
}

// IsEmailDomainAllowed checks if the given email domain is in the allowed domains list
func (c *Config) IsEmailDomainAllowed(domain string) bool {
	if domain == "" {
		return false
	}
	
	allowedDomains := c.GetAllowedEmailDomains()
	
	// If no restrictions are configured, allow all domains
	if len(allowedDomains) == 0 {
		return true
	}
	
	for _, allowed := range allowedDomains {
		if strings.EqualFold(domain, allowed) { // Case-insensitive comparison
			return true
		}
	}
	return false
}

// IsEmailAllowed checks if the complete email address has an allowed domain
func (c *Config) IsEmailAllowed(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	
	domain := strings.Split(email, "@")[1]
	return c.IsEmailDomainAllowed(domain)
}

// GetAllowedEmailDomainsString returns allowed email domains as a comma-separated string
func (c *Config) GetAllowedEmailDomainsString() string {
	domains := c.GetAllowedEmailDomains()
	if len(domains) == 0 {
		return "All domains allowed"
	}
	return strings.Join(domains, ", ")
}

// Add these methods to your utils_config_utility.go file



// GetSessionTimeout returns the session timeout duration


// GetPasswordPolicy returns password policy settings as a map
func (c *Config) GetPasswordPolicy() map[string]interface{} {
	if c == nil {
		// Return default password policy
		return map[string]interface{}{
			"min_length":      8,
			"max_length":      128,
			"require_upper":   true,
			"require_lower":   true,
			"require_numbers": true,
			"require_special": true,
			"special_chars":   "!@#$%^&*()_+-=[]{}|;:,.<>?/",
		}
	}
	
	return map[string]interface{}{
		"min_length":      c.Password.MinLength,
		"max_length":      c.Password.MaxLength,
		"require_upper":   c.Password.RequireUppercase,
		"require_lower":   c.Password.RequireLowercase,
		"require_numbers": c.Password.RequireNumbers,
		"require_special": c.Password.RequireSpecial,
		"special_chars":   c.Password.SpecialChars,
	}
}

// Additional password-related utility methods
func (c *Config) GetMinPasswordLength() int {
	if c == nil {
		return 8
	}
	return c.Password.MinLength
}

func (c *Config) GetMaxPasswordLength() int {
	if c == nil {
		return 128
	}
	return c.Password.MaxLength
}



func (c *Config) GetPasswordSpecialChars() string {
	if c == nil {
		return "!@#$%^&*()_+-=[]{}|;:,.<>?/"
	}
	return c.Password.SpecialChars
}



// Add the missing field to your setDefaults() method:
// In your setDefaults() function, make sure you have:
// cm.viper.SetDefault("security.session_timeout", "24h")
// cm.viper.SetDefault("security.max_login_attempts", 5)


// Add these methods to your utils_config_utility.go file


// GetSessionTimeout returns the session timeout duration
func (c *Config) GetSessionTimeout() time.Duration {
	if c == nil {
		return 24 * time.Hour // Default fallback
	}
	
	// SessionTimeout is already a time.Duration, no parsing needed
	if c.Security.SessionTimeout > 0 {
		return c.Security.SessionTimeout
	}
	
	// Fallback to default if not set
	return 24 * time.Hour
}

// GetPasswordPolicy returns password policy settings as a map







func (c *Config) GetAccountLockoutDuration() time.Duration {
	if c == nil {
		return 15 * time.Minute
	}
	return time.Duration(c.Security.AccountLockoutMinutes) * time.Minute
}

// Add the missing field to your setDefaults() method:
// In your setDefaults() function, make sure you have:
// cm.viper.SetDefault("security.session_timeout", "24h")
// cm.viper.SetDefault("security.max_login_attempts", 5)
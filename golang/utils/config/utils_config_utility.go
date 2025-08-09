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
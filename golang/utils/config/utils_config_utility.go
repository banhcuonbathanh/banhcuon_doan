package utils_config

import (

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
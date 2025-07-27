package utils_config

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config holds all configuration for the account handler
type Config struct {
	// Password validation settings
	PasswordMinLength int      `mapstructure:"password_min_length" json:"password_min_length"`
	PasswordMaxLength int      `mapstructure:"password_max_length" json:"password_max_length"`
	ValidRoles       []string `mapstructure:"valid_roles" json:"valid_roles"`
	
	// Pagination settings
	PaginationLimit    int `mapstructure:"pagination_limit" json:"pagination_limit"`
	DefaultPageSize    int `mapstructure:"default_page_size" json:"default_page_size"`
	MaxPageSize        int `mapstructure:"max_page_size" json:"max_page_size"`
	
	// Token settings
	JWTSecretKey          string `mapstructure:"jwt_secret_key" json:"-"` // Hide in JSON
	JWTExpirationHours    int    `mapstructure:"jwt_expiration_hours" json:"jwt_expiration_hours"`
	RefreshTokenExpirationDays int `mapstructure:"refresh_token_expiration_days" json:"refresh_token_expiration_days"`
	
	// Email settings
	EmailVerificationEnabled bool   `mapstructure:"email_verification_enabled" json:"email_verification_enabled"`
	EmailVerificationExpiry  int    `mapstructure:"email_verification_expiry_hours" json:"email_verification_expiry_hours"`
	
	// Security settings
	MaxLoginAttempts     int  `mapstructure:"max_login_attempts" json:"max_login_attempts"`
	AccountLockoutMinutes int  `mapstructure:"account_lockout_minutes" json:"account_lockout_minutes"`
	RequireEmailVerification bool `mapstructure:"require_email_verification" json:"require_email_verification"`
	
	// API Rate limiting
	RateLimitEnabled     bool `mapstructure:"rate_limit_enabled" json:"rate_limit_enabled"`
	RateLimitPerMinute   int  `mapstructure:"rate_limit_per_minute" json:"rate_limit_per_minute"`
	
	// Database settings (if needed)
	DatabaseURL         string `mapstructure:"database_url" json:"-"`
	DatabaseMaxConnections int `mapstructure:"database_max_connections" json:"database_max_connections"`
}

// Global configuration instance
var AppConfig *Config

// LoadConfig loads configuration from various sources using Viper
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}
	
	// Set default values
	setDefaults()
	
	// Set config file path and name
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Look for config in multiple locations
		viper.SetConfigName("config")
		viper.SetConfigType("yaml") // Can also support json, toml, etc.
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/etc/english-ai/")
		viper.AddConfigPath("$HOME/.english-ai")
	}
	
	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ENGLISH_AI") // Will look for ENGLISH_AI_PASSWORD_MIN_LENGTH, etc.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is OK, we can use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}
	
	// Unmarshal config into struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	// Set global config
	AppConfig = config
	
	return config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Password settings
	viper.SetDefault("password_min_length", 8)
	viper.SetDefault("password_max_length", 128)
	viper.SetDefault("valid_roles", []string{"admin", "user", "manager"})
	
	// Pagination settings
	viper.SetDefault("pagination_limit", 100)
	viper.SetDefault("default_page_size", 10)
	viper.SetDefault("max_page_size", 100)
	
	// Token settings
	viper.SetDefault("jwt_expiration_hours", 24)
	viper.SetDefault("refresh_token_expiration_days", 30)
	
	// Email settings
	viper.SetDefault("email_verification_enabled", true)
	viper.SetDefault("email_verification_expiry_hours", 24)
	viper.SetDefault("require_email_verification", false)
	
	// Security settings
	viper.SetDefault("max_login_attempts", 5)
	viper.SetDefault("account_lockout_minutes", 15)
	
	// Rate limiting
	viper.SetDefault("rate_limit_enabled", true)
	viper.SetDefault("rate_limit_per_minute", 60)
	
	// Database settings
	viper.SetDefault("database_max_connections", 25)
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	if c.PasswordMinLength < 1 {
		return fmt.Errorf("password_min_length must be at least 1")
	}
	
	if c.PasswordMaxLength < c.PasswordMinLength {
		return fmt.Errorf("password_max_length must be greater than password_min_length")
	}
	
	if len(c.ValidRoles) == 0 {
		return fmt.Errorf("valid_roles cannot be empty")
	}
	
	if c.DefaultPageSize < 1 || c.DefaultPageSize > c.MaxPageSize {
		return fmt.Errorf("default_page_size must be between 1 and max_page_size")
	}
	
	if c.MaxPageSize > c.PaginationLimit {
		return fmt.Errorf("max_page_size cannot exceed pagination_limit")
	}
	
	if c.JWTExpirationHours < 1 {
		return fmt.Errorf("jwt_expiration_hours must be at least 1")
	}
	
	if c.JWTSecretKey == "" {
		return fmt.Errorf("jwt_secret_key is required")
	}
	
	if len(c.JWTSecretKey) < 32 {
		return fmt.Errorf("jwt_secret_key must be at least 32 characters")
	}
	
	return nil
}

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

// ReloadConfig reloads the configuration (useful for config updates without restart)
func ReloadConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reloading config: %w", err)
	}
	
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("error unmarshaling reloaded config: %w", err)
	}
	
	if err := AppConfig.Validate(); err != nil {
		return fmt.Errorf("reloaded config validation failed: %w", err)
	}
	
	return nil
}

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	return AppConfig
}

// WatchConfig watches for configuration file changes and reloads automatically
func WatchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
		if err := ReloadConfig(); err != nil {
			fmt.Printf("Error reloading config: %v\n", err)
		} else {
			fmt.Println("Config reloaded successfully")
		}
	})
}
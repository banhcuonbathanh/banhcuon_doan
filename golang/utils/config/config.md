package utils_config


import (
	"context"
	"fmt"
	"os"
	
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Environment constants
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging" 
	EnvProduction  = "production"
	EnvTesting     = "testing"
	EnvDocker      = "docker"
)

// Configuration structure with comprehensive validation tags
type Config struct {
	// Environment settings
	Environment string `mapstructure:"environment" json:"environment" validate:"required,oneof=development staging production testing docker"`
	AppName     string `mapstructure:"app_name" json:"app_name" validate:"required,min=1,max=50"`
	Version     string `mapstructure:"version" json:"version" validate:"required"`
	Debug       bool   `mapstructure:"debug" json:"debug"`

	// Server settings
	Server ServerConfig `mapstructure:"server" json:"server" validate:"required"`
	
	// Database settings
	Database DatabaseConfig `mapstructure:"database" json:"database" validate:"required"`
	
	// Security settings
	Security SecurityConfig `mapstructure:"security" json:"security" validate:"required"`
	
	// Password validation settings
	Password PasswordConfig `mapstructure:"password" json:"password" validate:"required"`
	
	// Pagination settings
	Pagination PaginationConfig `mapstructure:"pagination" json:"pagination" validate:"required"`
	
	// Token/JWT settings
	JWT JWTConfig `mapstructure:"jwt" json:"jwt" validate:"required"`
	
	// Email settings
	Email EmailConfig `mapstructure:"email" json:"email" validate:"required"`
	
	// Rate limiting settings
	RateLimit RateLimitConfig `mapstructure:"rate_limit" json:"rate_limit" validate:"required"`
	
	// External API settings
	ExternalAPIs ExternalAPIConfig `mapstructure:"external_apis" json:"external_apis" validate:"required"`
	
	// Logging settings
	Logging LoggingConfig `mapstructure:"logging" json:"logging" validate:"required"`
	
	// Valid user roles
	ValidRoles []string `mapstructure:"valid_roles" json:"valid_roles" validate:"required,min=1,dive,required"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Address      string        `mapstructure:"address" json:"address" validate:"required"`
	Port         int           `mapstructure:"port" json:"port" validate:"required,min=1,max=65535"`
	GRPCAddress  string        `mapstructure:"grpc_address" json:"grpc_address" validate:"required"`
	GRPCPort     int           `mapstructure:"grpc_port" json:"grpc_port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" json:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout" validate:"required"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" json:"idle_timeout" validate:"required"`
	TLSEnabled   bool          `mapstructure:"tls_enabled" json:"tls_enabled"`
	CertFile     string        `mapstructure:"cert_file" json:"cert_file"`
	KeyFile      string        `mapstructure:"key_file" json:"key_file"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string        `mapstructure:"url" json:"-" validate:"required"` // Hidden from JSON
	Host            string        `mapstructure:"host" json:"host" validate:"required"`
	Port            int           `mapstructure:"port" json:"port" validate:"required,min=1,max=65535"`
	Name            string        `mapstructure:"name" json:"name" validate:"required"`
	User            string        `mapstructure:"user" json:"user" validate:"required"`
	Password        string        `mapstructure:"password" json:"-"` // Hidden from JSON
	SSLMode         string        `mapstructure:"ssl_mode" json:"ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxConnections  int           `mapstructure:"max_connections" json:"max_connections" validate:"required,min=1,max=100"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" json:"max_idle_conns" validate:"required,min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" json:"conn_max_idle_time" validate:"required"`
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	MaxLoginAttempts      int           `mapstructure:"max_login_attempts" json:"max_login_attempts" validate:"required,min=1,max=20"`
	AccountLockoutMinutes int           `mapstructure:"account_lockout_minutes" json:"account_lockout_minutes" validate:"required,min=1"`
	SessionTimeout        time.Duration `mapstructure:"session_timeout" json:"session_timeout" validate:"required"`
	CSRFEnabled           bool          `mapstructure:"csrf_enabled" json:"csrf_enabled"`
	CORSEnabled           bool          `mapstructure:"cors_enabled" json:"cors_enabled"`
	AllowedOrigins        []string      `mapstructure:"allowed_origins" json:"allowed_origins"`
	RequireHTTPS          bool          `mapstructure:"require_https" json:"require_https"`
}

// PasswordConfig holds password validation settings
type PasswordConfig struct {
	MinLength        int  `mapstructure:"min_length" json:"min_length" validate:"required,min=4,max=32"`
	MaxLength        int  `mapstructure:"max_length" json:"max_length" validate:"required,min=8,max=256"`
	RequireUppercase bool `mapstructure:"require_uppercase" json:"require_uppercase"`
	RequireLowercase bool `mapstructure:"require_lowercase" json:"require_lowercase"`
	RequireNumbers   bool `mapstructure:"require_numbers" json:"require_numbers"`
	RequireSpecial   bool `mapstructure:"require_special" json:"require_special"`
	SpecialChars     string `mapstructure:"special_chars" json:"special_chars" validate:"required"`
}

// PaginationConfig holds pagination settings
type PaginationConfig struct {
	DefaultSize int `mapstructure:"default_size" json:"default_size" validate:"required,min=1,max=100"`
	MaxSize     int `mapstructure:"max_size" json:"max_size" validate:"required,min=1,max=1000"`
	Limit       int `mapstructure:"limit" json:"limit" validate:"required,min=1"`
}

// JWTConfig holds JWT/token settings
type JWTConfig struct {
	SecretKey                 string        `mapstructure:"secret_key" json:"-" validate:"required,min=32"` // Hidden from JSON
	ExpirationHours           int           `mapstructure:"expiration_hours" json:"expiration_hours" validate:"required,min=1,max=720"`
	RefreshTokenExpirationDays int          `mapstructure:"refresh_token_expiration_days" json:"refresh_token_expiration_days" validate:"required,min=1,max=365"`
	Issuer                    string        `mapstructure:"issuer" json:"issuer" validate:"required"`
	Algorithm                 string        `mapstructure:"algorithm" json:"algorithm" validate:"required,oneof=HS256 HS384 HS512 RS256 RS384 RS512"`
	RefreshThreshold          time.Duration `mapstructure:"refresh_threshold" json:"refresh_threshold" validate:"required"`
}

// EmailConfig holds email-related settings
type EmailConfig struct {
	VerificationEnabled    bool          `mapstructure:"verification_enabled" json:"verification_enabled"`
	VerificationExpiryHours int          `mapstructure:"verification_expiry_hours" json:"verification_expiry_hours" validate:"required,min=1,max=168"`
	RequireVerification    bool          `mapstructure:"require_verification" json:"require_verification"`
	SMTPHost               string        `mapstructure:"smtp_host" json:"smtp_host"`
	SMTPPort               int           `mapstructure:"smtp_port" json:"smtp_port" validate:"min=1,max=65535"`
	SMTPUser               string        `mapstructure:"smtp_user" json:"smtp_user"`
	SMTPPassword           string        `mapstructure:"smtp_password" json:"-"` // Hidden from JSON
	FromAddress            string        `mapstructure:"from_address" json:"from_address" validate:"email"`
	FromName               string        `mapstructure:"from_name" json:"from_name"`
	Templates              EmailTemplates `mapstructure:"templates" json:"templates"`
}

// EmailTemplates holds email template paths
type EmailTemplates struct {
	VerificationTemplate string `mapstructure:"verification_template" json:"verification_template"`
	WelcomeTemplate      string `mapstructure:"welcome_template" json:"welcome_template"`
	ResetPasswordTemplate string `mapstructure:"reset_password_template" json:"reset_password_template"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	Enabled    bool `mapstructure:"enabled" json:"enabled"`
	PerMinute  int  `mapstructure:"per_minute" json:"per_minute" validate:"required,min=1,max=10000"`
	PerHour    int  `mapstructure:"per_hour" json:"per_hour" validate:"required,min=1"`
	BurstSize  int  `mapstructure:"burst_size" json:"burst_size" validate:"required,min=1"`
	WindowSize time.Duration `mapstructure:"window_size" json:"window_size" validate:"required"`
}

// ExternalAPIConfig holds external API settings
type ExternalAPIConfig struct {
	Anthropic AnthropicConfig `mapstructure:"anthropic" json:"anthropic" validate:"required"`
	QuanAn    QuanAnConfig    `mapstructure:"quan_an" json:"quan_an" validate:"required"`
}

// AnthropicConfig holds Anthropic API settings
type AnthropicConfig struct {
	APIKey     string        `mapstructure:"api_key" json:"-" validate:"required"` // Hidden from JSON
	APIURL     string        `mapstructure:"api_url" json:"api_url" validate:"required,url"`
	Timeout    time.Duration `mapstructure:"timeout" json:"timeout" validate:"required"`
	MaxRetries int           `mapstructure:"max_retries" json:"max_retries" validate:"required,min=0,max=10"`
}

// QuanAnConfig holds QuanAn service settings
type QuanAnConfig struct {
	Address    string        `mapstructure:"address" json:"address" validate:"required"`
	Timeout    time.Duration `mapstructure:"timeout" json:"timeout" validate:"required"`
	MaxRetries int           `mapstructure:"max_retries" json:"max_retries" validate:"required,min=0,max=10"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level" json:"level" validate:"required,oneof=debug info warn error fatal panic"`
	Format     string `mapstructure:"format" json:"format" validate:"required,oneof=json text"`
	Output     string `mapstructure:"output" json:"output" validate:"required,oneof=stdout stderr file"`
	FilePath   string `mapstructure:"file_path" json:"file_path"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size" validate:"min=1"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" validate:"min=0"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age" validate:"min=0"`
	Compress   bool   `mapstructure:"compress" json:"compress"`
}

// ConfigManager manages configuration with hot-reloading and validation
type ConfigManager struct {
	config    *Config
	mu        sync.RWMutex
	validator *validator.Validate
	viper     *viper.Viper
	callbacks []ConfigChangeCallback
}

// ConfigChangeCallback is called when configuration changes
type ConfigChangeCallback func(oldConfig, newConfig *Config) error

// ConfigLoader interface for dependency injection
type ConfigLoader interface {
	Load(ctx context.Context, configPath string) (*Config, error)
	Reload(ctx context.Context) error
	Watch(ctx context.Context) error
	GetConfig() *Config
	RegisterCallback(callback ConfigChangeCallback)
	Validate() error
	Stop() error
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		validator: validator.New(),
		viper:     viper.New(),
		callbacks: make([]ConfigChangeCallback, 0),
	}
}

// Load loads configuration from file and environment variables
func (cm *ConfigManager) Load(ctx context.Context, configPath string) (*Config, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Set defaults first
	cm.setDefaults()

	// Configure viper
	if configPath != "" {
		cm.viper.SetConfigFile(configPath)
	} else {
		cm.viper.SetConfigName("config")
		cm.viper.SetConfigType("yaml")
		cm.viper.AddConfigPath(".")
		cm.viper.AddConfigPath("./config")
		cm.viper.AddConfigPath("./configs")
		cm.viper.AddConfigPath("/etc/english-ai/")
		cm.viper.AddConfigPath("$HOME/.english-ai")
	}

	// Environment variable configuration
	cm.viper.AutomaticEnv()
	cm.viper.SetEnvPrefix("ENGLISH_AI")
	cm.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := cm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment-specific settings
	if err := cm.handleEnvironmentOverrides(); err != nil {
		return nil, fmt.Errorf("error handling environment overrides: %w", err)
	}

	config := &Config{}
	if err := cm.viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := cm.validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Perform cross-field validation
	if err := cm.validateCrossFields(config); err != nil {
		return nil, fmt.Errorf("cross-field validation failed: %w", err)
	}

	cm.config = config
	return config, nil
}

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
	cm.viper.SetDefault("database.ssl_mode", "disable")
	cm.viper.SetDefault("database.max_connections", 25)
	cm.viper.SetDefault("database.max_idle_conns", 10)
	cm.viper.SetDefault("database.conn_max_lifetime", "1h")
	cm.viper.SetDefault("database.conn_max_idle_time", "10m")

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
	cm.viper.SetDefault("jwt.expiration_hours", 24)
	cm.viper.SetDefault("jwt.refresh_token_expiration_days", 30)
	cm.viper.SetDefault("jwt.issuer", "english-ai")
	cm.viper.SetDefault("jwt.algorithm", "HS256")
	cm.viper.SetDefault("jwt.refresh_threshold", "2h")

	// Email settings
	cm.viper.SetDefault("email.verification_enabled", true)
	cm.viper.SetDefault("email.verification_expiry_hours", 24)
	cm.viper.SetDefault("email.require_verification", false)
	cm.viper.SetDefault("email.smtp_port", 587)
	cm.viper.SetDefault("email.from_name", "English AI")

	// Rate limiting settings
	cm.viper.SetDefault("rate_limit.enabled", true)
	cm.viper.SetDefault("rate_limit.per_minute", 60)
	cm.viper.SetDefault("rate_limit.per_hour", 3600)
	cm.viper.SetDefault("rate_limit.burst_size", 10)
	cm.viper.SetDefault("rate_limit.window_size", "1m")

	// External API settings
	cm.viper.SetDefault("external_apis.anthropic.api_url", "https://api.anthropic.com")
	cm.viper.SetDefault("external_apis.anthropic.timeout", "30s")
	cm.viper.SetDefault("external_apis.anthropic.max_retries", 3)
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
}

// handleEnvironmentOverrides handles environment-specific configuration overrides
func (cm *ConfigManager) handleEnvironmentOverrides() error {
	env := cm.viper.GetString("environment")

	switch env {
	case EnvDocker:
		return cm.handleDockerOverrides()
	case EnvProduction:
		return cm.handleProductionOverrides()
	case EnvTesting:
		return cm.handleTestingOverrides()
	}

	return nil
}

// handleDockerOverrides applies Docker-specific configuration
func (cm *ConfigManager) handleDockerOverrides() error {
	// Construct database URL from environment variables
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		dbUser := getEnvOrDefault("DB_USER", "postgres")
		dbPassword := getEnvOrDefault("DB_PASSWORD", "")
		dbName := getEnvOrDefault("DB_NAME", "english_ai")
		dbPort := getEnvOrDefault("DB_PORT", "5432")
		
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort, dbName)
		cm.viper.Set("database.url", dbURL)
	}

	// Override server addresses for Docker
	cm.viper.Set("server.address", "0.0.0.0")
	cm.viper.Set("server.grpc_address", "0.0.0.0")

	return nil
}

// handleProductionOverrides applies production-specific configuration
func (cm *ConfigManager) handleProductionOverrides() error {
	// Force secure settings in production
	cm.viper.Set("debug", false)
	cm.viper.Set("security.require_https", true)
	cm.viper.Set("security.csrf_enabled", true)
	cm.viper.Set("server.tls_enabled", true)
	cm.viper.Set("logging.level", "info")

	return nil
}

// handleTestingOverrides applies testing-specific configuration
func (cm *ConfigManager) handleTestingOverrides() error {
	// Use in-memory database for testing
	cm.viper.Set("database.url", "sqlite:///:memory:")
	cm.viper.Set("security.max_login_attempts", 100) // Don't lockout during tests
	cm.viper.Set("rate_limit.enabled", false)         // Disable rate limiting in tests
	cm.viper.Set("logging.level", "error")            // Reduce log noise in tests

	return nil
}

// validateConfig validates the configuration using struct tags
func (cm *ConfigManager) validateConfig(config *Config) error {
	return cm.validator.Struct(config)
}

// validateCrossFields performs cross-field validation
func (cm *ConfigManager) validateCrossFields(config *Config) error {
	// Password length validation
	if config.Password.MaxLength < config.Password.MinLength {
		return fmt.Errorf("password max_length (%d) must be greater than min_length (%d)",
			config.Password.MaxLength, config.Password.MinLength)
	}

	// Pagination validation
	if config.Pagination.DefaultSize > config.Pagination.MaxSize {
		return fmt.Errorf("pagination default_size (%d) cannot exceed max_size (%d)",
			config.Pagination.DefaultSize, config.Pagination.MaxSize)
	}

	// JWT secret validation in production
	if config.Environment == EnvProduction && len(config.JWT.SecretKey) < 64 {
		return fmt.Errorf("JWT secret key must be at least 64 characters in production")
	}

	// HTTPS validation in production
	if config.Environment == EnvProduction && !config.Security.RequireHTTPS {
		return fmt.Errorf("HTTPS must be enabled in production environment")
	}

	// Email verification validation
	if config.Email.RequireVerification && !config.Email.VerificationEnabled {
		return fmt.Errorf("email verification must be enabled if required")
	}

	return nil
}

// Reload reloads the configuration
func (cm *ConfigManager) Reload(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	oldConfig := cm.config

	if err := cm.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error rereading config: %w", err)
	}

	newConfig := &Config{}
	if err := cm.viper.Unmarshal(newConfig); err != nil {
		return fmt.Errorf("error unmarshaling reloaded config: %w", err)
	}

	if err := cm.validateConfig(newConfig); err != nil {
		return fmt.Errorf("reloaded config validation failed: %w", err)
	}

	if err := cm.validateCrossFields(newConfig); err != nil {
		return fmt.Errorf("reloaded config cross-field validation failed: %w", err)
	}

	// Notify callbacks
	for _, callback := range cm.callbacks {
		if err := callback(oldConfig, newConfig); err != nil {
			return fmt.Errorf("config change callback failed: %w", err)
		}
	}

	cm.config = newConfig
	return nil
}

// Watch starts watching for configuration file changes
func (cm *ConfigManager) Watch(ctx context.Context) error {
	cm.viper.WatchConfig()
	cm.viper.OnConfigChange(func(e fsnotify.Event) {
		select {
		case <-ctx.Done():
			return
		default:
			if err := cm.Reload(ctx); err != nil {
				// Log error (in production, use proper logger)
				fmt.Printf("Error reloading config: %v\n", err)
			}
		}
	})
	
	return nil
}

// GetConfig returns the current configuration (thread-safe)
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// RegisterCallback registers a callback for configuration changes
func (cm *ConfigManager) RegisterCallback(callback ConfigChangeCallback) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.callbacks = append(cm.callbacks, callback)
}

// Validate validates the current configuration
func (cm *ConfigManager) Validate() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}
	
	if err := cm.validateConfig(cm.config); err != nil {
		return err
	}
	
	return cm.validateCrossFields(cm.config)
}

// Stop stops the configuration manager
func (cm *ConfigManager) Stop() error {
	// Clean up resources if needed
	return nil
}

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

// GetDatabaseURL returns the complete database URL
func (c *Config) GetDatabaseURL() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}
	
	// Construct URL from individual components
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the complete server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Address, c.Server.Port)
}

// GetGRPCAddress returns the complete gRPC server address
func (c *Config) GetGRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.GRPCAddress, c.Server.GRPCPort)
}

// Helper functions

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Global configuration manager instance (can be replaced with DI)
var globalConfigManager *ConfigManager

// InitConfig initializes the global configuration manager
func InitConfig(ctx context.Context, configPath string) (*Config, error) {
	globalConfigManager = NewConfigManager()
	return globalConfigManager.Load(ctx, configPath)
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *Config {
	if globalConfigManager == nil {
		return nil
	}
	return globalConfigManager.GetConfig()
}

// ReloadGlobalConfig reloads the global configuration
func ReloadGlobalConfig(ctx context.Context) error {
	if globalConfigManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return globalConfigManager.Reload(ctx)
}

// WatchGlobalConfig starts watching for global configuration changes
func WatchGlobalConfig(ctx context.Context) error {
	if globalConfigManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return globalConfigManager.Watch(ctx)
}

// Example usage and testing helpers

// NewTestConfig creates a configuration for testing
func NewTestConfig() *Config {
	return &Config{
		Environment: EnvTesting,
		AppName:     "English AI Test",
		Version:     "test",
		Debug:       true,
		Server: ServerConfig{
			Address:      "localhost",
			Port:         8080,
			GRPCAddress:  "localhost", 
			GRPCPort:     50051,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Database: DatabaseConfig{
			URL:             "sqlite:///:memory:",
			Host:            "localhost",
			Port:            5432,
			Name:            "test_db",
			User:            "test",
			Password:        "test",
			SSLMode:         "disable",
			MaxConnections:  5,
			MaxIdleConns:    2,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Security: SecurityConfig{
			MaxLoginAttempts:      100,
			AccountLockoutMinutes: 1,
			SessionTimeout:        24 * time.Hour,
			CSRFEnabled:           false,
			CORSEnabled:           true,
			AllowedOrigins:        []string{"*"},
			RequireHTTPS:          false,
		},
		Password: PasswordConfig{
			MinLength:        6,
			MaxLength:        128,
			RequireUppercase: false,
			RequireLowercase: false,
			RequireNumbers:   false,
			RequireSpecial:   false,
			SpecialChars:     "!@#$%^&*",
		},
		Pagination: PaginationConfig{
			DefaultSize: 10,
			MaxSize:     100,
			Limit:       1000,
		},
		JWT: JWTConfig{
			SecretKey:                  "test-secret-key-32-characters-long",
			ExpirationHours:            24,
			RefreshTokenExpirationDays: 30,
			Issuer:                     "english-ai-test",
			Algorithm:                  "HS256",
			RefreshThreshold:           2 * time.Hour,
		},
		Email: EmailConfig{
			VerificationEnabled:     false,
			VerificationExpiryHours: 24,
			RequireVerification:     false,
			SMTPHost:                "localhost",
			SMTPPort:                587,
			FromAddress:             "test@example.com",
			FromName:                "Test",
		},
		RateLimit: RateLimitConfig{
			Enabled:    false,
			PerMinute:  1000,
			PerHour:    10000,
			BurstSize:  100,
			WindowSize: time.Minute,
		},
		ExternalAPIs: ExternalAPIConfig{
			Anthropic: AnthropicConfig{
				APIKey:     "test-key",
				APIURL:     "https://api.anthropic.com",
				Timeout:    30 * time.Second,
				MaxRetries: 0,
			},
			QuanAn: QuanAnConfig{
				Address:    "localhost:8081",
				Timeout:    10 * time.Second,
				MaxRetries: 0,
			},
		},
		Logging: LoggingConfig{
			Level:      "error",
			Format:     "text",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   false,
		},
		ValidRoles: []string{"admin", "user"},
	}
}
package utils_config

import "time"

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
	MinLength        int    `mapstructure:"min_length" json:"min_length" validate:"required,min=4,max=32"`
	MaxLength        int    `mapstructure:"max_length" json:"max_length" validate:"required,min=8,max=256"`
	RequireUppercase bool   `mapstructure:"require_uppercase" json:"require_uppercase"`
	RequireLowercase bool   `mapstructure:"require_lowercase" json:"require_lowercase"`
	RequireNumbers   bool   `mapstructure:"require_numbers" json:"require_numbers"`
	RequireSpecial   bool   `mapstructure:"require_special" json:"require_special"`
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
	SecretKey                  string        `mapstructure:"secret_key" json:"-" validate:"required,min=32"` // Hidden from JSON
	ExpirationHours            int           `mapstructure:"expiration_hours" json:"expiration_hours" validate:"required,min=1,max=720"`
	RefreshTokenExpirationDays int           `mapstructure:"refresh_token_expiration_days" json:"refresh_token_expiration_days" validate:"required,min=1,max=365"`
	Issuer                     string        `mapstructure:"issuer" json:"issuer" validate:"required"`
	Algorithm                  string        `mapstructure:"algorithm" json:"algorithm" validate:"required,oneof=HS256 HS384 HS512 RS256 RS384 RS512"`
	RefreshThreshold           time.Duration `mapstructure:"refresh_threshold" json:"refresh_threshold" validate:"required"`
}

// EmailConfig holds email-related settings
type EmailConfig struct {
	VerificationEnabled     bool           `mapstructure:"verification_enabled" json:"verification_enabled"`
	VerificationExpiryHours int            `mapstructure:"verification_expiry_hours" json:"verification_expiry_hours" validate:"required,min=1,max=168"`
	RequireVerification     bool           `mapstructure:"require_verification" json:"require_verification"`
	SMTPHost                string         `mapstructure:"smtp_host" json:"smtp_host"`
	SMTPPort                int            `mapstructure:"smtp_port" json:"smtp_port" validate:"min=1,max=65535"`
	SMTPUser                string         `mapstructure:"smtp_user" json:"smtp_user"`
	SMTPPassword            string         `mapstructure:"smtp_password" json:"-"` // Hidden from JSON
	FromAddress             string         `mapstructure:"from_address" json:"from_address" validate:"email"`
	FromName                string         `mapstructure:"from_name" json:"from_name"`
	Templates               EmailTemplates `mapstructure:"templates" json:"templates"`
}

// EmailTemplates holds email template paths
type EmailTemplates struct {
	VerificationTemplate  string `mapstructure:"verification_template" json:"verification_template"`
	WelcomeTemplate       string `mapstructure:"welcome_template" json:"welcome_template"`
	ResetPasswordTemplate string `mapstructure:"reset_password_template" json:"reset_password_template"`
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	Enabled    bool          `mapstructure:"enabled" json:"enabled"`
	PerMinute  int           `mapstructure:"per_minute" json:"per_minute" validate:"required,min=1,max=10000"`
	PerHour    int           `mapstructure:"per_hour" json:"per_hour" validate:"required,min=1"`
	BurstSize  int           `mapstructure:"burst_size" json:"burst_size" validate:"required,min=1"`
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
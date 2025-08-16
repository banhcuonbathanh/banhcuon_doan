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
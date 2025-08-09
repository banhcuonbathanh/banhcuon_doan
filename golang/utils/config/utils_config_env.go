package utils_config

import (
	"fmt"
	"os"
	"strconv"
)

// handleEnvironmentOverrides handles environment-specific configuration overrides
func (cm *ConfigManager) handleEnvironmentOverrides() error {
	env := cm.viper.GetString("environment")

	// Check for APP_ENV environment variable to override environment
	if appEnv := os.Getenv("APP_ENV"); appEnv != "" {
		env = appEnv
		cm.viper.Set("environment", env)
	}

	// Always handle common environment variables first
	cm.handleCommonEnvironmentVariables()

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

// handleCommonEnvironmentVariables handles environment variables common to all environments
func (cm *ConfigManager) handleCommonEnvironmentVariables() {
	// Database configuration
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cm.viper.Set("database.host", dbHost)
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		if port, err := strconv.Atoi(dbPort); err == nil {
			cm.viper.Set("database.port", port)
		}
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cm.viper.Set("database.user", dbUser)
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		cm.viper.Set("database.password", dbPassword)
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cm.viper.Set("database.name", dbName)
	}

	// Always construct database URL after setting individual fields
	dbHost := cm.viper.GetString("database.host")
	dbPort := cm.viper.GetInt("database.port")
	dbUser := cm.viper.GetString("database.user")
	dbPassword := cm.viper.GetString("database.password")
	dbName := cm.viper.GetString("database.name")
	sslMode := cm.viper.GetString("database.ssl_mode")

	if dbHost != "" && dbUser != "" && dbName != "" {
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
		cm.viper.Set("database.url", dbURL)
	}

	// JWT configuration
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cm.viper.Set("jwt.secret_key", jwtSecret)
	}

	// Anthropic API configuration
	if anthropicKey := os.Getenv("ANTHROPIC_API_KEY"); anthropicKey != "" {
		cm.viper.Set("external_apis.anthropic.api_key", anthropicKey)
	}
	if anthropicURL := os.Getenv("ANTHROPIC_API_URL"); anthropicURL != "" {
		cm.viper.Set("external_apis.anthropic.api_url", anthropicURL)
	}

	// QuanAn service configuration
	if quanAnAddr := os.Getenv("QUAN_AN_ADDRESS"); quanAnAddr != "" {
		cm.viper.Set("external_apis.quan_an.address", quanAnAddr)
	}

	// Email configuration
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		cm.viper.Set("email.smtp_host", smtpHost)
	}
	if smtpPort := os.Getenv("SMTP_PORT"); smtpPort != "" {
		if port, err := strconv.Atoi(smtpPort); err == nil {
			cm.viper.Set("email.smtp_port", port)
		}
	}
	if smtpUser := os.Getenv("SMTP_USER"); smtpUser != "" {
		cm.viper.Set("email.smtp_user", smtpUser)
	}
	if smtpPassword := os.Getenv("SMTP_PASSWORD"); smtpPassword != "" {
		cm.viper.Set("email.smtp_password", smtpPassword)
	}
	if fromAddress := os.Getenv("EMAIL_FROM_ADDRESS"); fromAddress != "" {
		cm.viper.Set("email.from_address", fromAddress)
	}
}

// handleDockerOverrides applies Docker-specific configuration
func (cm *ConfigManager) handleDockerOverrides() error {
	// Override server addresses for Docker
	cm.viper.Set("server.address", "0.0.0.0")
	cm.viper.Set("server.grpc_address", "0.0.0.0")

	// Set default QuanAn address for Docker if not set
	if cm.viper.GetString("external_apis.quan_an.address") == "localhost:8081" {
		cm.viper.Set("external_apis.quan_an.address", "quan_an:8081")
	}

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

	// Ensure JWT secret is set from environment in production
	if jwtSecret := os.Getenv("ENGLISH_AI_JWT_SECRET_KEY"); jwtSecret != "" {
		cm.viper.Set("jwt.secret_key", jwtSecret)
	}

	// Ensure database password is set from environment
	if dbPassword := os.Getenv("ENGLISH_AI_DATABASE_PASSWORD"); dbPassword != "" {
		cm.viper.Set("database.password", dbPassword)
	}

	// Ensure Anthropic API key is set from environment
	if anthropicKey := os.Getenv("ENGLISH_AI_EXTERNAL_APIS_ANTHROPIC_API_KEY"); anthropicKey != "" {
		cm.viper.Set("external_apis.anthropic.api_key", anthropicKey)
	}

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

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}